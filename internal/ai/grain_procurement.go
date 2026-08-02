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

// aiProcureGrain, aktif ticaret ağı üzerinden stratejik rezervi eksik AI
// fraksiyonuna tahıl alır. Kaynakta da güvenli rezerv bırakılır; böylece büyük
// devletler küçük devletleri satın alırken kendi ordularını kıtlığa sokmaz.
func aiProcureGrain(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
		return 0
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return 0
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

	connected := aiTradeNetworkMembers(gs, fid)
	suppliers := make([]aiResourceSupplier, 0, len(connected))
	for supplierID := range connected {
		if supplierID == fid {
			continue
		}
		supplier := gs.Factions[supplierID]
		if supplier == nil || supplier.IsEliminated {
			continue
		}
		surplus := gs.StrategicGrainSurplus(supplierID)
		if surplus <= 0 {
			continue
		}
		suppliers = append(suppliers, aiResourceSupplier{id: supplierID, surplus: surplus})
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
		if buy <= 0 || !economy.TransferGoods(gs.Factions, candidate.id, fid, economy.GoodGrain, buy, gs.MarketPrices) {
			continue
		}
		purchased += buy
	}
	return purchased
}

// aiProcureStrategicResources, AI'nin aynı turda gerçekten değerlendireceği
// üretim adaylarının maliyetlerinden kaynak talebi çıkarır. Kaynak stoğu
// yeterliyse alım yapmaz; eksik olan her ticari malı aktif ticaret ağı içinden,
// altın rezervi korunabildiği sürece tamamlar.
func aiProcureStrategicResources(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) economy.ResourceCost {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
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
	if candidate, ok := aiBestBuildingInvestmentForProcurement(gs, fid, ctx); ok {
		demand = aiMaxResourceCost(demand, candidate.Cost)
	}

	// Deniz tehdidi veya aktif çıkarma görevi varsa, ilgili üretim maliyeti de
	// aynı tedarik kararının parçasıdır. Teknoloji yoksa henüz savaş gemisi/
	// nakliye üretimi mümkün değildir; araştırma tamamlanana kadar satın alma
	// yapılmaz.
	if gs.UnitTypes != nil {
		if warshipType := gs.UnitTypes["warship"]; warshipType != nil && warshipType.HasAllRequiredTechs(self.Research.Completed) && (len(ctx.NavalThreats) > 0 || len(ctx.ThreatenedPortIDs) > 0) {
			demand = aiMaxResourceCost(demand, aiUnitResourceCost(warshipType))
		}
		if transportType := gs.UnitTypes["transport"]; transportType != nil && transportType.HasAllRequiredTechs(self.Research.Completed) && ctx.navalMission != nil && ctx.navalMission.MissingCapacity > 0 {
			demand = aiMaxResourceCost(demand, aiUnitResourceCost(transportType))
		}
	}
	if reserve := aiMerchantTradeResourceReserve(gs, fid); reserve != (economy.ResourceCost{}) {
		demand = aiMaxResourceCost(demand, reserve)
	}

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
		bought := aiPurchaseResourceFromTradeNetwork(gs, fid, need.good, amount)
		if bought <= 0 {
			continue
		}
		setResourceCostAmount(&purchased, need.kind, bought)
		goldBudget -= bought * need.price
	}
	return purchased
}

func aiPurchaseResourceFromTradeNetwork(gs *state.GameState, fid faction.FactionID, good economy.GoodType, amount int) int {
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
	type supplier struct {
		id      faction.FactionID
		surplus int
	}
	suppliers := make([]supplier, 0)
	kind, _ := economy.GoodToResourceKind(good)
	for supplierID := range aiTradeNetworkMembers(gs, fid) {
		if supplierID == fid {
			continue
		}
		candidate := gs.Factions[supplierID]
		if candidate == nil || candidate.IsEliminated {
			continue
		}
		surplus := maxInt(0, economy.FactionResourceAmount(candidate, kind)-20)
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
		if buy > 0 && economy.TransferGoods(gs.Factions, candidate.id, fid, good, buy, gs.MarketPrices) {
			purchased += buy
		}
	}
	return purchased
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

// aiTradeNetworkMembers returns factions reachable through active trade
// routes. AI procurement follows the same connected-network rule as the
// player's one-time market trade, without requiring a direct bilateral route.
func aiTradeNetworkMembers(gs *state.GameState, fid faction.FactionID) map[faction.FactionID]struct{} {
	connected := map[faction.FactionID]struct{}{fid: {}}
	queue := []faction.FactionID{fid}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, route := range gs.TradeRoutes {
			if route == nil || route.SuspendedTurns > 0 {
				continue
			}
			var next faction.FactionID
			switch faction.FactionID(route.FromFactionID) {
			case current:
				next = faction.FactionID(route.ToFactionID)
			default:
				if faction.FactionID(route.ToFactionID) == current {
					next = faction.FactionID(route.FromFactionID)
				}
			}
			if next == "" {
				continue
			}
			if _, seen := connected[next]; seen {
				continue
			}
			connected[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return connected
}
