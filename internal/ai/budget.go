package ai

import (
	"sort"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	aiBudgetBaseReserve      = 40
	aiBudgetReservePerRegion = 8
	aiBudgetReservePerWar    = 30
	aiBudgetCriticalReserve  = 40
	aiBudgetIncomeDivisor    = 3
	aiBudgetIncomeReserveCap = 120
	aiBudgetEmergencyGoldCap = 420
)

type aiBudgetCategory string

const (
	aiBudgetArmy     aiBudgetCategory = "army"
	aiBudgetEconomy  aiBudgetCategory = "economy"
	aiBudgetResearch aiBudgetCategory = "research"
	aiBudgetNaval    aiBudgetCategory = "naval"
)

var aiBudgetTieOrder = map[aiBudgetCategory]int{
	aiBudgetArmy:     0,
	aiBudgetEconomy:  1,
	aiBudgetResearch: 2,
	aiBudgetNaval:    3,
}

type aiBudget struct {
	EmergencyGold int
	SpendableGold int
	Allocation    map[aiBudgetCategory]int
	Remaining     map[aiBudgetCategory]int
	Spent         map[aiBudgetCategory]int
	FlexibleGold  int
	Order         []aiBudgetCategory
	// ResourceReserve, somut bir sonraki deniz yatırımı için diğer bütçe
	// kategorilerinin tüketemeyeceği gerçek malzeme tabanıdır.
	ResourceReserve economy.ResourceCost
	// OperationalGrainReserve, üretim ve asker harcamasından sonra korunacak
	// minimum tahıl stoğudur.
	OperationalGrainReserve int
}

func prepareAIBudget(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext) *aiBudget {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || fid == "" {
		return nil
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return nil
	}

	ownedRegions := 0
	grossIncome := 0
	hasCoast := false
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		ownedRegions++
		grossIncome += gs.RegionProductionSummary(region).Gold
		if region.IsCoastal(gs.Regions) {
			hasCoast = true
		}
	}

	warCount := 0
	criticalThreat := false
	planKind := state.AIObjectiveConsolidate
	if ctx != nil {
		criticalThreat = ctx.CriticalThreat
	}
	if plan := gs.AIPlans[fid]; plan != nil && plan.Kind != "" {
		planKind = plan.Kind
	}
	for _, relation := range gs.Relations {
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		if relation.FactionA == fid || relation.FactionB == fid {
			warCount++
		}
	}

	incomeReserve := minInt(aiBudgetIncomeReserveCap, grossIncome/aiBudgetIncomeDivisor)
	emergency := aiBudgetBaseReserve + ownedRegions*aiBudgetReservePerRegion + warCount*aiBudgetReservePerWar + incomeReserve
	if criticalThreat {
		emergency += aiBudgetCriticalReserve
	}
	emergency = minInt(aiBudgetEmergencyGoldCap, maxInt(aiMinGoldReserve, emergency))
	spendable := maxInt(0, self.Gold-emergency)

	weights := aiBudgetWeights(planKind, warCount > 0, hasCoast)
	if aiHasNavalFocus(gs, fid) && hasCoast {
		weights = aiNavalFocusBudgetWeights(planKind, warCount > 0)
	}
	allocation, _ := allocateAIBudget(spendable, weights)
	remaining := make(map[aiBudgetCategory]int, len(allocation))
	spent := make(map[aiBudgetCategory]int, len(allocation))
	for category, amount := range allocation {
		remaining[category] = amount
		spent[category] = 0
	}
	budget := &aiBudget{
		EmergencyGold: emergency,
		SpendableGold: spendable,
		Allocation:    allocation,
		Remaining:     remaining,
		Spent:         spent,
		Order:         aiBudgetExecutionOrder(hasCoast),
	}
	budget.ResourceReserve = aiMerchantTradeResourceReserve(gs, fid)
	if demand := aiFactionGrainDemand(gs, fid); demand > 0 {
		budget.OperationalGrainReserve = demand * 2
		if criticalThreat {
			budget.OperationalGrainReserve = demand * 3 / 2
		}
		budget.OperationalGrainReserve = maxInt(demand, budget.OperationalGrainReserve)
	}
	return budget
}

func aiBudgetExecutionOrder(hasCoast bool) []aiBudgetCategory {
	order := []aiBudgetCategory{aiBudgetResearch, aiBudgetEconomy}
	if hasCoast {
		order = append(order, aiBudgetNaval)
	}
	return append(order, aiBudgetArmy)
}

func aiBudgetWeights(planKind state.AIObjectiveKind, atWar, hasCoast bool) map[aiBudgetCategory]int {
	weights := map[aiBudgetCategory]int{
		aiBudgetArmy:     35,
		aiBudgetEconomy:  35,
		aiBudgetResearch: 20,
		aiBudgetNaval:    10,
	}
	if atWar || planKind == state.AIObjectiveDefend {
		weights[aiBudgetArmy] = 70
		weights[aiBudgetEconomy] = 10
		weights[aiBudgetResearch] = 10
		weights[aiBudgetNaval] = 10
	} else if planKind == state.AIObjectiveExpand {
		weights[aiBudgetArmy] = 55
		weights[aiBudgetEconomy] = 20
		weights[aiBudgetResearch] = 15
		weights[aiBudgetNaval] = 10
	}
	if !hasCoast {
		delete(weights, aiBudgetNaval)
	}
	return weights
}

// aiNavalFocusBudgetWeights keeps naval-oriented states from treating a large
// warship reserve as a residual 10% expense. The profile itself remains in
// scenario JSON; this helper only defines the common spending behaviour.
func aiNavalFocusBudgetWeights(planKind state.AIObjectiveKind, atWar bool) map[aiBudgetCategory]int {
	if atWar || planKind == state.AIObjectiveDefend {
		return map[aiBudgetCategory]int{
			aiBudgetArmy: 45, aiBudgetEconomy: 10, aiBudgetResearch: 10, aiBudgetNaval: 35,
		}
	}
	return map[aiBudgetCategory]int{
		aiBudgetArmy: 30, aiBudgetEconomy: 20, aiBudgetResearch: 15, aiBudgetNaval: 35,
	}
}

func allocateAIBudget(spendable int, weights map[aiBudgetCategory]int) (map[aiBudgetCategory]int, []aiBudgetCategory) {
	allocation := make(map[aiBudgetCategory]int, len(weights))
	order := make([]aiBudgetCategory, 0, len(weights))
	totalWeight := 0
	for category, weight := range weights {
		if weight <= 0 {
			continue
		}
		totalWeight += weight
		order = append(order, category)
	}
	sort.Slice(order, func(i, j int) bool {
		wi := weights[order[i]]
		wj := weights[order[j]]
		if wi != wj {
			return wi > wj
		}
		return aiBudgetTieOrder[order[i]] < aiBudgetTieOrder[order[j]]
	})
	if spendable <= 0 || totalWeight <= 0 {
		return allocation, order
	}
	allocated := 0
	for _, category := range order {
		amount := spendable * weights[category] / totalWeight
		allocation[category] = amount
		allocated += amount
	}
	for index := 0; allocated < spendable && len(order) > 0; index++ {
		category := order[index%len(order)]
		allocation[category]++
		allocated++
	}
	return allocation, order
}

func (budget *aiBudget) canAfford(self *faction.Faction, cost economy.ResourceCost, category aiBudgetCategory) bool {
	if budget == nil {
		return aiCanAffordWithReserve(self, cost)
	}
	if self == nil || self.Gold-cost.Gold < budget.EmergencyGold {
		return false
	}
	if self.Grain < cost.Grain || self.Iron < cost.Iron || self.Timber < cost.Timber || self.Stone < cost.Stone {
		return false
	}
	if self.Grain-cost.Grain < budget.OperationalGrainReserve {
		return false
	}
	if category != aiBudgetNaval {
		reserve := budget.ResourceReserve
		if self.Gold-cost.Gold < budget.EmergencyGold+reserve.Gold || self.Grain-cost.Grain < reserve.Grain || self.Iron-cost.Iron < reserve.Iron || self.Timber-cost.Timber < reserve.Timber || self.Stone-cost.Stone < reserve.Stone {
			return false
		}
	}
	return cost.Gold <= budget.Remaining[category]+budget.FlexibleGold
}

func (budget *aiBudget) consume(category aiBudgetCategory, gold int) {
	if budget == nil || gold <= 0 {
		return
	}
	fromCategory := minInt(gold, budget.Remaining[category])
	budget.Remaining[category] -= fromCategory
	rest := gold - fromCategory
	budget.FlexibleGold = maxInt(0, budget.FlexibleGold-rest)
	budget.Spent[category] += gold
}

func (budget *aiBudget) release(category aiBudgetCategory) {
	if budget == nil {
		return
	}
	budget.FlexibleGold += budget.Remaining[category]
	budget.Remaining[category] = 0
}

func aiCanAffordForBudget(self *faction.Faction, cost economy.ResourceCost, budget *aiBudget, category aiBudgetCategory) bool {
	if budget == nil {
		return aiCanAffordWithReserve(self, cost)
	}
	return budget.canAfford(self, cost, category)
}

func aiApplyBudgetedCost(self *faction.Faction, cost economy.ResourceCost, budget *aiBudget, category aiBudgetCategory) bool {
	if !aiCanAffordForBudget(self, cost, budget, category) {
		return false
	}
	cost.Apply(self)
	if budget != nil {
		budget.consume(category, cost.Gold)
	}
	return true
}
