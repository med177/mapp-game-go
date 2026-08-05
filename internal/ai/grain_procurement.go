package ai

import (
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	aiGrainReserveMonths   = 3
	aiGrainPurchaseWindow  = 2
	aiGrainPurchaseMinimum = 12
	aiMilitaryIronReserve  = 40
)

type aiResourceSupplier struct {
	id      faction.FactionID
	surplus int
}

// aiGrainSupplierSurplus, satıcının kendi üç aylık stratejik rezervini
// korurken pazara sunabileceği tahılı döner. Depo kapasitesi sivil stok için
// altı aya kadar büyüyebildiğinden onu satış eşiği yapmak, kıtlıktaki AI'ler
// için pazarda hiç satıcı kalmamasına yol açıyordu.
func aiGrainSupplierSurplus(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" {
		return 0
	}
	supplier := gs.Factions[fid]
	if supplier == nil || supplier.IsEliminated {
		return 0
	}
	reserve := maxInt(100, aiFactionGrainDemand(gs, fid)*aiGrainReserveMonths)
	return maxInt(0, supplier.Grain-reserve)
}

// aiProcureGrain, açık pazardan stratejik rezervi eksik AI fraksiyonuna tahıl
// alır. Savaşta olmayan tüm devletler satıcı olabilir; kaynakta yine güvenli
// rezerv bırakılır.
func aiProcureGrain(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
		return 0
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return 0
	}
	if len(gs.MarketOrders.SellOffers) == 0 || len(gs.MarketOrders.BuyOrders) == 0 {
		RefreshMarketOrders(gs)
	}

	totalDemand := aiFactionGrainDemand(gs, fid)
	if totalDemand <= 0 {
		return 0
	}
	capacity := gs.GrainStorageCapacityForFaction(fid)
	target := maxInt(capacity, totalDemand*aiGrainReserveMonths)
	shortfall := target - self.Grain
	if shortfall <= 0 {
		return 0
	}

	amount := minInt(shortfall, maxInt(aiGrainPurchaseMinimum, totalDemand*aiGrainPurchaseWindow))
	price := gs.MarketPrices[economy.GoodGrain]
	if price <= 0 {
		price = economy.BaseGoldValue[economy.GoodGrain]
	}
	if price <= 0 {
		return 0
	}
	if gs.MarketPrices == nil {
		gs.MarketPrices = make(map[economy.GoodType]int)
	}
	if gs.MarketPrices[economy.GoodGrain] <= 0 {
		gs.MarketPrices[economy.GoodGrain] = price
	}

	goldBudget := self.Gold - aiMinGoldReserve
	if goldBudget <= 0 {
		return 0
	}
	if oneThird := self.Gold / 3; oneThird > 0 && goldBudget > oneThird {
		goldBudget = oneThird
	}
	amount = minInt(amount, goldBudget/price)
	if amount <= 0 {
		return 0
	}

	return aiMarketPurchaseFromOrders(gs, fid, economy.GoodGrain, amount)
}

// aiProcureStrategicResources, AI'nin aynı turda gerçekten değerlendireceği
// üretim adaylarının maliyetlerinden kaynak talebi çıkarır. Kaynak stoğu
// yeterliyse alım yapmaz; eksik olan her ticari malı açık pazardan, altın
// rezervi korunabildiği sürece tamamlar.
func aiProcureStrategicResources(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) economy.ResourceCost {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
		return economy.ResourceCost{}
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return economy.ResourceCost{}
	}
	if len(gs.MarketOrders.SellOffers) == 0 || len(gs.MarketOrders.BuyOrders) == 0 {
		RefreshMarketOrders(gs)
	}
	demand := aiStrategicResourceDemand(gs, fid, ctx)
	return aiProcureMissingResources(gs, fid, demand)
}

// aiProcureMilitaryIron, eski çağrı sözleşmesini koruyan dar sarmalayıcıdır.
// Yeni AI akışı tüm ResourceCost alanlarını aiProcureStrategicResources ile
// ele alır; bu yardımcı eski testler ve doğrudan çağrılar için tutulur.
func aiProcureMilitaryIron(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || !aiFactionAtWar(gs, string(fid)) {
		return 0
	}
	purchased := aiProcureMissingResources(gs, fid, economy.ResourceCost{Iron: aiMilitaryIronReserve})
	return purchased.Iron
}

func aiMaxResourceCost(current, candidate economy.ResourceCost) economy.ResourceCost {
	return economy.ResourceCost{
		Gold:   maxInt(current.Gold, candidate.Gold),
		Grain:  maxInt(current.Grain, candidate.Grain),
		Iron:   maxInt(current.Iron, candidate.Iron),
		Timber: maxInt(current.Timber, candidate.Timber),
		Stone:  maxInt(current.Stone, candidate.Stone),
		Spice:  maxInt(current.Spice, candidate.Spice),
		Cloth:  maxInt(current.Cloth, candidate.Cloth),
	}
}

func aiProcureMissingResources(gs *state.GameState, fid faction.FactionID, demand economy.ResourceCost) economy.ResourceCost {
	if gs == nil || fid == "" {
		return economy.ResourceCost{}
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return economy.ResourceCost{}
	}
	if len(gs.MarketOrders.SellOffers) == 0 || len(gs.MarketOrders.BuyOrders) == 0 {
		RefreshMarketOrders(gs)
	}

	// Tahılın uzun dönemli rezerv kararı ayrı çalışır; burada yalnız üretim
	// maliyetindeki açığı kapatıyoruz.
	type resourceNeed struct {
		kind     economy.ResourceKind
		good     economy.GoodType
		deficit  int
		price    int
		priority int
	}
	needs := make([]resourceNeed, 0, len(economy.TradeGoods()))
	for _, kind := range economy.CostResourceKinds() {
		if kind == economy.ResourceGold {
			continue
		}
		required := demand.Amount(kind)
		deficit := required - economy.FactionResourceAmount(self, kind)
		if deficit <= 0 {
			continue
		}
		good := economy.ResourceDefByKind(kind).TradeGood
		price := aiResourcePrice(gs, good)
		if good == "" || price <= 0 {
			continue
		}
		needs = append(needs, resourceNeed{kind: kind, good: good, deficit: deficit, price: price, priority: required})
	}
	if len(needs) == 0 {
		return economy.ResourceCost{}
	}

	goldBudget := self.Gold - aiMinGoldReserve
	if goldBudget <= 0 {
		return economy.ResourceCost{}
	}
	// Önce daha büyük üretim gereksinimlerini tamamla; altın tüm açığı
	// karşılamıyorsa ucuz bir malın tüm bütçeyi tüketmesi engellenir.
	sort.Slice(needs, func(i, j int) bool {
		left := needs[i].priority * needs[j].price
		right := needs[j].priority * needs[i].price
		if left != right {
			return left > right
		}
		return needs[i].kind < needs[j].kind
	})

	purchased := economy.ResourceCost{}
	for _, need := range needs {
		if goldBudget < need.price {
			continue
		}
		amount := minInt(need.deficit, goldBudget/need.price)
		if amount <= 0 {
			continue
		}
		// Eski/doğrudan çağrıların (ör. askerî rezerv alımı) talebi
		// stratejik plan snapshot'ında bulunmayabilir. Gerçek alım yine aynı
		// emir defterine yazılarak gerçekleşsin.
		if order := gs.MarketBuyOrder(fid, need.good, need.price); order < amount {
			gs.SetMarketBuyOrder(fid, need.good, minInt(need.deficit, goldBudget/need.price))
		}
		bought := aiPurchaseResourceFromOpenMarket(gs, fid, need.good, amount)
		if bought <= 0 {
			continue
		}
		setResourceCostAmount(&purchased, need.kind, bought)
		goldBudget -= bought * need.price
	}
	return purchased
}

func aiPurchaseResourceFromOpenMarket(gs *state.GameState, fid faction.FactionID, good economy.GoodType, amount int) int {
	if amount <= 0 || good == "" {
		return 0
	}
	if gs.MarketPrices == nil {
		gs.MarketPrices = make(economy.CurrentMarketPrice)
	}
	price := aiResourcePrice(gs, good)
	if gs.MarketPrices[good] <= 0 {
		gs.MarketPrices[good] = price
	}
	return aiMarketPurchaseFromOrders(gs, fid, good, amount)
}

func setResourceCostAmount(cost *economy.ResourceCost, kind economy.ResourceKind, amount int) {
	if cost == nil {
		return
	}
	switch kind {
	case economy.ResourceGrain:
		cost.Grain += amount
	case economy.ResourceIron:
		cost.Iron += amount
	case economy.ResourceTimber:
		cost.Timber += amount
	case economy.ResourceStone:
		cost.Stone += amount
	case economy.ResourceSpice:
		cost.Spice += amount
	case economy.ResourceCloth:
		cost.Cloth += amount
	}
}

func aiFactionGrainDemand(gs *state.GameState, fid faction.FactionID) int {
	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > 0 {
		return status.TotalDemand
	}
	demand := 0
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		demand += gs.CivilianGrainDemandForRegion(region)
	}
	for _, armyRef := range gs.Armies {
		if armyRef != nil && armyRef.OwnerID == string(fid) {
			demand += gs.EffectiveArmyGrainUpkeep(armyRef)
		}
	}
	return demand
}
