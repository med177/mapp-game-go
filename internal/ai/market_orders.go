package ai

import (
	"sort"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// RefreshMarketOrders AI devletlerinin açık pazar arz ve taleplerini mevcut
// stok, altın ve stratejik planlarından yeniden üretir. Emir defteri tur
// başına hazırlanır; başarılı işlemler sonrasında state helper'ları miktarı
// düşürür.
func RefreshMarketOrders(gs *state.GameState) {
	if gs == nil {
		return
	}
	gs.MarketOrders = state.MarketOrderBook{
		SellOffers: make(map[faction.FactionID]map[economy.GoodType]int),
		BuyOrders:  make(map[faction.FactionID]map[economy.GoodType]int),
	}
	for _, fid := range aiSortedFactionIDs(gs) {
		if fid == gs.PlayerFactionID {
			continue
		}
		f := gs.Factions[fid]
		if f == nil || f.IsEliminated {
			continue
		}
		refreshFactionMarketOrders(gs, fid)
	}
}

func refreshFactionMarketOrders(gs *state.GameState, fid faction.FactionID) {
	if gs == nil || fid == "" {
		return
	}
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated {
		return
	}
	ctx := prepareStrategicContext(gs, fid)
	resourceDemand := aiStrategicResourceDemand(gs, fid, ctx)

	for _, good := range economy.TradeGoods() {
		kind, ok := economy.GoodToResourceKind(good)
		if !ok {
			continue
		}
		reserve := marketReserveForGood(gs, fid, good, resourceDemand.Amount(kind))
		stock := economy.FactionResourceAmount(f, kind)
		if surplus := stock - reserve; surplus > 0 {
			gs.SetMarketSellOffer(fid, good, surplus)
		}

		need := resourceDemand.Amount(kind) - stock
		if good == economy.GoodGrain {
			need = gs.StrategicGrainDemand(fid)
		}
		if need <= 0 {
			continue
		}
		price := marketOrderPrice(gs, good)
		if price <= 0 {
			continue
		}
		gold := f.Gold - aiMinGoldReserve
		if gold > 0 {
			if affordable := gold / price; affordable < need {
				need = affordable
			}
			if need > 0 {
				gs.SetMarketBuyOrder(fid, good, need)
			}
		}
	}
}

func marketReserveForGood(gs *state.GameState, fid faction.FactionID, good economy.GoodType, planned int) int {
	if good == economy.GoodGrain {
		return maxInt(100, aiFactionGrainDemand(gs, fid)*aiGrainReserveMonths)
	}
	// Yirmi birim, üretim kaynaklarında güvenli çalışma stoğudur. Planın
	// maliyeti bunun üstündeyse AI hedefini de satıştan önce korur.
	return maxInt(20, planned)
}

func marketOrderPrice(gs *state.GameState, good economy.GoodType) int {
	if gs != nil && gs.MarketPrices != nil && gs.MarketPrices[good] > 0 {
		return gs.MarketPrices[good]
	}
	return economy.BaseGoldValue[good]
}

// aiStrategicResourceDemand açık pazardaki alım talebinin ortak kaynağıdır.
// Üretim, askerî rezerv, kışla önkoşulu ve deniz hedefleri aynı maliyet
// hesabını kullanır; böylece paneldeki talep AI'nin gerçek planından kopmaz.
func aiStrategicResourceDemand(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) economy.ResourceCost {
	if gs == nil || fid == "" {
		return economy.ResourceCost{}
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return economy.ResourceCost{}
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, fid)
	}

	demand := economy.ResourceCost{}
	if unitID := aiSelectStrategicLandUnitForProcurement(gs, self, ctx); unitID != "" {
		demand = aiMaxResourceCost(demand, aiUnitResourceCost(gs.UnitTypes[unitID]))
	}
	if aiLandReserveShortfall(gs, fid, ctx) > 0 {
		if unitID := aiSelectReserveLandUnitForProcurement(gs, self, ctx); unitID != "" {
			demand = aiMaxResourceCost(demand, aiUnitResourceCost(gs.UnitTypes[unitID]))
		}
	}
	if candidate, ok := aiBestBuildingInvestmentForProcurement(gs, fid, ctx); ok {
		demand = aiMaxResourceCost(demand, candidate.Cost)
	}
	if selfManpower := gs.DeployedLandUnits(fid) + aiPendingLandUnitCount(gs, fid); aiNeedsBarracksForMilitaryProduction(gs, fid, ctx, gs.ManpowerCap(fid)-selfManpower) {
		demand = aiMaxResourceCost(demand, aiBarracksResourceCost(gs))
	}
	if gs.UnitTypes != nil {
		if warshipType := gs.UnitTypes["warship"]; warshipType != nil && warshipType.HasAllRequiredTechs(self.Research.Completed) && (len(ctx.NavalThreats) > 0 || len(ctx.ThreatenedPortIDs) > 0) {
			demand = aiMaxResourceCost(demand, aiUnitResourceCost(warshipType))
		}
		if transportType := gs.UnitTypes["transport"]; transportType != nil && transportType.HasAllRequiredTechs(self.Research.Completed) && ctx.navalMission != nil && ctx.navalMission.MissingCapacity > 0 {
			demand = aiMaxResourceCost(demand, aiUnitResourceCost(transportType))
		}
	}
	if navalReserveCost := aiNavalReserveProcurementCost(gs, fid, ctx); navalReserveCost != (economy.ResourceCost{}) {
		demand = aiMaxResourceCost(demand, navalReserveCost)
	}
	if reserve := aiMerchantTradeResourceReserve(gs, fid); reserve != (economy.ResourceCost{}) {
		demand = aiMaxResourceCost(demand, reserve)
	}
	return demand
}

// aiMarketPurchaseFromOrders, AI'nin talep emrini ve satıcının arz emrini
// birlikte tüketerek açık pazardan alım yapar.
func aiMarketPurchaseFromOrders(gs *state.GameState, buyer faction.FactionID, good economy.GoodType, amount int) int {
	if gs == nil || amount <= 0 {
		return 0
	}
	price := marketOrderPrice(gs, good)
	remainingDemand := gs.MarketBuyOrder(buyer, good, price)
	if remainingDemand < amount {
		amount = remainingDemand
	}
	if amount <= 0 {
		return 0
	}
	type supplier struct {
		id      faction.FactionID
		surplus int
	}
	suppliers := make([]supplier, 0, len(gs.Factions))
	for _, supplierID := range aiSortedFactionIDs(gs) {
		if supplierID == buyer || diplomacy.IsWar(gs, buyer, supplierID) {
			continue
		}
		if supplier := gs.Factions[supplierID]; supplier == nil || supplier.IsEliminated {
			continue
		}
		surplus := gs.MarketSellOffer(supplierID, good)
		if surplus > 0 {
			suppliers = append(suppliers, supplier{id: supplierID, surplus: surplus})
		}
	}
	sort.Slice(suppliers, func(i, j int) bool {
		if suppliers[i].surplus != suppliers[j].surplus {
			return suppliers[i].surplus > suppliers[j].surplus
		}
		return suppliers[i].id < suppliers[j].id
	})

	purchased := 0
	for _, candidate := range suppliers {
		if purchased >= amount {
			break
		}
		buy := minInt(amount-purchased, candidate.surplus)
		if buy <= 0 || !economy.TransferGoods(gs.Factions, candidate.id, buyer, good, buy, gs.MarketPrices) {
			continue
		}
		gs.ConsumeMarketSellOffer(candidate.id, good, buy)
		gs.ConsumeMarketBuyOrder(buyer, good, buy)
		purchased += buy
	}
	return purchased
}
