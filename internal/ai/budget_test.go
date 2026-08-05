package ai

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAIBudgetProfilesUseApprovedShares(t *testing.T) {
	tests := []struct {
		name     string
		kind     state.AIObjectiveKind
		atWar    bool
		expected map[aiBudgetCategory]int
	}{
		{
			name: "expansion", kind: state.AIObjectiveExpand,
			expected: map[aiBudgetCategory]int{aiBudgetArmy: 55, aiBudgetEconomy: 20, aiBudgetResearch: 15, aiBudgetNaval: 10},
		},
		{
			name: "defense", kind: state.AIObjectiveDefend,
			expected: map[aiBudgetCategory]int{aiBudgetArmy: 70, aiBudgetEconomy: 10, aiBudgetResearch: 10, aiBudgetNaval: 10},
		},
		{
			name: "war overrides plan", kind: state.AIObjectiveExpand, atWar: true,
			expected: map[aiBudgetCategory]int{aiBudgetArmy: 70, aiBudgetEconomy: 10, aiBudgetResearch: 10, aiBudgetNaval: 10},
		},
		{
			name: "consolidation", kind: state.AIObjectiveConsolidate,
			expected: map[aiBudgetCategory]int{aiBudgetArmy: 35, aiBudgetEconomy: 35, aiBudgetResearch: 20, aiBudgetNaval: 10},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allocation, _ := allocateAIBudget(100, aiBudgetWeights(test.kind, test.atWar, true))
			for category, expected := range test.expected {
				if got := allocation[category]; got != expected {
					t.Fatalf("%s payı beklenenden farklı: got=%d want=%d allocation=%+v", category, got, expected, allocation)
				}
			}
		})
	}
}

func TestLandlockedBudgetRedistributesNavalShare(t *testing.T) {
	allocation, _ := allocateAIBudget(90, aiBudgetWeights(state.AIObjectiveExpand, false, false))
	if _, exists := allocation[aiBudgetNaval]; exists {
		t.Fatalf("kara devletinde donanma bütçesi oluşmamalıydı: %+v", allocation)
	}
	if allocation[aiBudgetArmy] != 55 || allocation[aiBudgetEconomy] != 20 || allocation[aiBudgetResearch] != 15 {
		t.Fatalf("donanma payı kalan kategorilere oransal dağılmalıydı: %+v", allocation)
	}
	order := aiBudgetExecutionOrder(false)
	if len(order) != 3 || order[0] != aiBudgetResearch || order[1] != aiBudgetEconomy || order[2] != aiBudgetArmy {
		t.Fatalf("kara devleti harcama sırası deterministik değil: %+v", order)
	}
}

func TestAINavalFocusBudgetPrioritizesFleetWithoutStarvingArmy(t *testing.T) {
	allocation, _ := allocateAIBudget(100, aiNavalFocusBudgetWeights(state.AIObjectiveDefend, false))
	if allocation[aiBudgetNaval] != 35 || allocation[aiBudgetArmy] != 45 {
		t.Fatalf("denizci savunma profili beklenen 45/35 kara-deniz dağılımını vermeli: %+v", allocation)
	}
}

func TestPrepareAIBudgetScalesEmergencyReserve(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "1300_ottoman_rise",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":    {ID: "ai", Gold: 500},
			"enemy": {ID: "enemy"},
		},
		Regions:   make(map[world.RegionID]*world.Region),
		Relations: map[string]*faction.Relation{},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ai": {Kind: state.AIObjectiveDefend},
		},
	}
	for index := 0; index < 10; index++ {
		id := world.RegionID(string(rune('a' + index)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "ai", TaxRate: 100, Satisfaction: 50}
	}
	gs.Relations[faction.RelationKey("ai", "enemy")] = &faction.Relation{
		FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar,
	}
	ctx := &StrategicContext{CriticalThreat: true}

	budget := prepareAIBudget(gs, "ai", ctx)
	if budget == nil {
		t.Fatal("1300 senaryosu runtime bütçe üretmeliydi")
	}
	// 40 taban + 10*8 bölge + 30 savaş + 40 kritik tehdit; gelir sıfır.
	if budget.EmergencyGold != 190 || budget.SpendableGold != 310 {
		t.Fatalf("acil rezerv ölçeklemesi yanlış: %+v", budget)
	}
}

func TestAIBudgetReleasesUnusedShareAndProtectsEmergencyGold(t *testing.T) {
	budget := &aiBudget{
		EmergencyGold: 100,
		SpendableGold: 100,
		Allocation: map[aiBudgetCategory]int{
			aiBudgetArmy: 60, aiBudgetEconomy: 15, aiBudgetResearch: 15, aiBudgetNaval: 10,
		},
		Remaining: map[aiBudgetCategory]int{
			aiBudgetArmy: 60, aiBudgetEconomy: 15, aiBudgetResearch: 15, aiBudgetNaval: 10,
		},
		Spent: make(map[aiBudgetCategory]int),
	}
	self := &faction.Faction{Gold: 200, Grain: 20}
	budget.release(aiBudgetArmy)
	if budget.FlexibleGold != 60 {
		t.Fatalf("kullanılmayan ordu payı esnek havuza aktarılmadı: %+v", budget)
	}
	cost := economy.ResourceCost{Gold: 70, Grain: 5}
	if !aiApplyBudgetedCost(self, cost, budget, aiBudgetEconomy) {
		t.Fatal("ekonomi kendi payı ile serbest kalan payı birlikte kullanabilmeliydi")
	}
	if self.Gold != 130 || budget.Spent[aiBudgetEconomy] != 70 || budget.FlexibleGold != 5 {
		t.Fatalf("yumuşak bütçe tüketimi yanlış: self=%+v budget=%+v", self, budget)
	}
	budget.FlexibleGold = 100
	if aiApplyBudgetedCost(self, economy.ResourceCost{Gold: 31}, budget, aiBudgetResearch) {
		t.Fatal("esnek bütçe yetse bile acil rezervin altına inen harcama engellenmeliydi")
	}
}

func TestPrepareAIBudgetKeepsOtherScenariosOnLegacyPath(t *testing.T) {
	gs := &state.GameState{
		ScenarioID: "other",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", Gold: 200},
		},
	}
	if budget := prepareAIBudget(gs, "ai", nil); budget != nil {
		t.Fatalf("diğer senaryolar runtime bütçe yerine legacy rezervi kullanmalıydı: %+v", budget)
	}
	self := gs.Factions["ai"]
	if !aiCanAffordForBudget(self, economy.ResourceCost{Gold: 120}, nil, aiBudgetArmy) {
		t.Fatal("legacy sabit 80 rezerv davranışı korunmalıydı")
	}
	if aiCanAffordForBudget(self, economy.ResourceCost{Gold: 121}, nil, aiBudgetArmy) {
		t.Fatal("legacy rezerv tabanının altına inilmemeliydi")
	}
}
