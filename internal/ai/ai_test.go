package ai

import (
	"math/rand"
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiTestTransportType() *army.UnitType {
	return &army.UnitType{ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10}
}

func TestAIHandlesPeaceWhenWarPressureIsHigh(t *testing.T) {
	gs := aiTestState()
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["ai_2"].Gold = 30

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StancePeace {
		t.Fatalf("AI barış aramalıydı, got=%s", rel.Stance)
	}
}

func TestAIQueuesPeaceOfferToPlayer(t *testing.T) {
	gs := aiTestState()
	gs.PlayerFactionID = "player"
	rel := gs.Relations[faction.RelationKey("ai_1", "player")]
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["player"].Gold = 30

	aiHandleDiplomacy(gs, "ai_1")

	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("oyuncuya barış teklifi kuyruğu bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.FromFactionID != "ai_1" || offer.ToFactionID != "player" || offer.Action != string(diplomacy.ActionProposePeace) {
		t.Fatalf("beklenmeyen teklif kaydı: %+v", offer)
	}
	if offer.PriorityReason == "" {
		t.Fatal("barış teklifinde öncelik sebebi kaydı bekleniyordu")
	}
	if rel.Stance != faction.StanceWar {
		t.Fatalf("oyuncu yanıtlayana kadar savaş stance'i korunmalı, got=%s", rel.Stance)
	}
}

func TestAIDoesNotRepeatRejectedPeaceOfferDuringCooldown(t *testing.T) {
	gs := aiTestState()
	gs.PlayerFactionID = "player"
	rel := gs.Relations[faction.RelationKey("ai_1", "player")]
	rel.Stance = faction.StanceWar
	rel.Score = -100
	gs.Factions["player"].Gold = 30

	aiHandleDiplomacy(gs, "ai_1")
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("ilk barış teklifi bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}
	if result := diplomacy.ResolveOffer(gs, 0, false); result.Accepted || result.Applied {
		t.Fatalf("oyuncu reddi kabul edilmiş görünmemeli: %+v", result)
	}

	gs.Turn++
	aiHandleDiplomacy(gs, "ai_1")
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("ret sonrası cooldown sırasında barış teklifi tekrarlanmamalı, got=%+v", gs.DiplomaticOffers)
	}
}

func TestAIPrioritizesPeaceOffersByThreatAndTechGap(t *testing.T) {
	gs := aiTestState()
	gs.Relations[faction.RelationKey("ai_1", "player")].Stance = faction.StanceWar
	gs.Relations[faction.RelationKey("ai_1", "player")].Score = -100
	gs.Relations[faction.RelationKey("ai_2", "player")].Stance = faction.StanceWar
	gs.Relations[faction.RelationKey("ai_2", "player")].Score = -100
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Stance = faction.StancePeace
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = 0
	gs.Armies["player_army"] = &army.Army{
		ID:       "player_army",
		OwnerID:  "player",
		RegionID: "p1",
		Units: []army.Unit{
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
			{TypeID: "inf", CurrentHP: 100},
		},
	}
	gs.Armies["ai1_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}
	gs.Armies["ai2_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Factions["ai_1"].Research.Completed = map[string]bool{}
	gs.Factions["ai_2"].Research.Completed = map[string]bool{"tech_a": true, "tech_b": true}
	gs.Factions["player"].Research.Completed = map[string]bool{"tech_a": true, "tech_b": true, "tech_c": true, "tech_d": true}

	aiHandleDiplomacy(gs, "ai_1")
	aiHandleDiplomacy(gs, "ai_2")

	if len(gs.DiplomaticOffers) != 2 {
		t.Fatalf("iki barış teklifi bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}
	bestIdx, ok := diplomacy.BestOfferIndex(gs, "player")
	if !ok {
		t.Fatal("öncelikli teklif bulunamadı")
	}
	best := gs.DiplomaticOffers[bestIdx]
	if best.FromFactionID != "ai_1" {
		t.Fatalf("daha baskı altındaki teklif önce gelmeliydi, got=%+v", best)
	}
	if best.Priority <= gs.DiplomaticOffers[1-bestIdx].Priority {
		t.Fatalf("öncelik puanı daha yüksek olmalıydı, got=%+v", gs.DiplomaticOffers)
	}
	if !strings.Contains(best.PriorityReason, "tehdit") && !strings.Contains(best.PriorityReason, "teknoloji") {
		t.Fatalf("barış teklif sebebi anlamlı olmalıydı, got=%q", best.PriorityReason)
	}
}

func TestAIQueuesAllianceAndTradeOffersWithPriority(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":    {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Gold: 600, Grain: 400},
			"ai_allied": {ID: "ai_allied", NameTR: "AI Allied", Religion: religion.Catholic, Gold: 450, Grain: 200, Research: faction.ResearchState{Completed: map[string]bool{}}, AIAggressiveness: 50},
			"ai_trader": {ID: "ai_trader", NameTR: "AI Trader", Religion: religion.Catholic, Gold: 450, Grain: 200, Research: faction.ResearchState{Completed: map[string]bool{"tech_a": true, "tech_b": true, "tech_c": true, "tech_d": true, "tech_e": true, "tech_f": true, "tech_g": true, "tech_h": true, "tech_i": true, "tech_j": true}}, AIAggressiveness: 45},
			"common_en": {ID: "common_en", NameTR: "Enemy", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"player_land":    {ID: "player_land", OwnerID: "player", TradeCapacity: 8, Buildings: []string{"port"}, Neighbors: []world.RegionID{"enemy_land", "ai_allied_land", "sea_trade"}},
			"ai_allied_land": {ID: "ai_allied_land", OwnerID: "ai_allied", TradeCapacity: 4, Neighbors: []world.RegionID{"player_land"}},
			"ai_trader_land": {ID: "ai_trader_land", OwnerID: "ai_trader", TradeCapacity: 4, Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_trade"}},
			"sea_trade":      {ID: "sea_trade", IsSea: true, Neighbors: []world.RegionID{"player_land", "ai_trader_land"}},
			"enemy_land":     {ID: "enemy_land", OwnerID: "common_en", TradeCapacity: 4, Neighbors: []world.RegionID{"player_land"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_allied", "player"):    {FactionA: "ai_allied", FactionB: "player", Score: 30, Stance: faction.StancePeace},
			faction.RelationKey("ai_trader", "player"):    {FactionA: "ai_trader", FactionB: "player", Score: 15, Stance: faction.StancePeace},
			faction.RelationKey("ai_allied", "common_en"): {FactionA: "ai_allied", FactionB: "common_en", Score: -80, Stance: faction.StanceWar},
			faction.RelationKey("player", "common_en"):    {FactionA: "player", FactionB: "common_en", Score: -80, Stance: faction.StanceWar},
			faction.RelationKey("ai_trader", "common_en"): {FactionA: "ai_trader", FactionB: "common_en", Score: 0, Stance: faction.StancePeace},
			faction.RelationKey("ai_allied", "ai_trader"): {FactionA: "ai_allied", FactionB: "ai_trader", Score: 0, Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "player_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"ally_army":   {ID: "ally_army", OwnerID: "ai_allied", RegionID: "ai_allied_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"trader_army": {ID: "trader_army", OwnerID: "ai_trader", RegionID: "ai_trader_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"enemy_army":  {ID: "enemy_army", OwnerID: "common_en", RegionID: "enemy_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}
	gs.Factions["player"].Research.Completed = map[string]bool{
		"tech_a": true,
		"tech_b": true,
		"tech_c": true,
		"tech_d": true,
		"tech_e": true,
		"tech_f": true,
		"tech_g": true,
		"tech_h": true,
		"tech_i": true,
		"tech_j": true,
	}

	aiHandleDiplomacy(gs, "ai_allied")
	aiHandleDiplomacy(gs, "ai_trader")

	if len(gs.DiplomaticOffers) != 2 {
		t.Fatalf("ittifak+ticaret teklifi bekleniyordu, got=%d", len(gs.DiplomaticOffers))
	}
	bestIdx, ok := diplomacy.BestOfferIndex(gs, "player")
	if !ok {
		t.Fatal("öncelikli teklif bulunamadı")
	}
	best := gs.DiplomaticOffers[bestIdx]
	other := gs.DiplomaticOffers[1-bestIdx]
	if best.Action != string(diplomacy.ActionProposeAlliance) && best.Action != string(diplomacy.ActionProposeTrade) {
		t.Fatalf("beklenmeyen teklif tipi öncelikli seçildi, got=%+v", best)
	}
	if (best.Action == string(diplomacy.ActionProposeAlliance) && other.Action != string(diplomacy.ActionProposeTrade)) ||
		(best.Action == string(diplomacy.ActionProposeTrade) && other.Action != string(diplomacy.ActionProposeAlliance)) {
		t.Fatalf("ittifak ve ticaret teklifleri birlikte bekleniyordu, got=%+v", gs.DiplomaticOffers)
	}
	if best.PriorityReason == "" || other.PriorityReason == "" {
		t.Fatalf("öncelik sebebi boş olmamalıydı, got=%+v", gs.DiplomaticOffers)
	}
	if best.Action == string(diplomacy.ActionProposeTrade) && !strings.Contains(best.PriorityReason, "ticaret") {
		t.Fatalf("ticaret sebebi görünür olmalıydı, got=%q", best.PriorityReason)
	}
	if best.Action == string(diplomacy.ActionProposeAlliance) && !strings.Contains(best.PriorityReason, "ortak") {
		t.Fatalf("ittifak sebebi görünür olmalıydı, got=%q", best.PriorityReason)
	}
	if other.Action == string(diplomacy.ActionProposeTrade) && !strings.Contains(other.PriorityReason, "ticaret") {
		t.Fatalf("ticaret sebebi görünür olmalıydı, got=%q", other.PriorityReason)
	}
	if other.Action == string(diplomacy.ActionProposeAlliance) && !strings.Contains(other.PriorityReason, "ortak") {
		t.Fatalf("ittifak sebebi görünür olmalıydı, got=%q", other.PriorityReason)
	}
}

func TestAIRespectsDiplomacyOfferQuotaPerTurn(t *testing.T) {
	gs := &state.GameState{
		Difficulty: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"ai": {ID: "ai", NameTR: "AI", Religion: religion.Catholic, AIAggressiveness: 20},
			"t1": {ID: "t1", NameTR: "T1", Religion: religion.Catholic},
			"t2": {ID: "t2", NameTR: "T2", Religion: religion.Catholic},
			"t3": {ID: "t3", NameTR: "T3", Religion: religion.Catholic},
			"t4": {ID: "t4", NameTR: "T4", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"ai_cap": {ID: "ai_cap", OwnerID: "ai", TradeCapacity: 4, Neighbors: []world.RegionID{"t1_cap", "t2_cap", "t3_cap", "t4_cap"}},
			"t1_cap": {ID: "t1_cap", OwnerID: "t1", TradeCapacity: 4, Neighbors: []world.RegionID{"ai_cap"}},
			"t2_cap": {ID: "t2_cap", OwnerID: "t2", TradeCapacity: 4, Neighbors: []world.RegionID{"ai_cap"}},
			"t3_cap": {ID: "t3_cap", OwnerID: "t3", TradeCapacity: 4, Neighbors: []world.RegionID{"ai_cap"}},
			"t4_cap": {ID: "t4_cap", OwnerID: "t4", TradeCapacity: 4, Neighbors: []world.RegionID{"ai_cap"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "t1"): {FactionA: "ai", FactionB: "t1", Score: -100, Stance: faction.StanceWar},
			faction.RelationKey("ai", "t2"): {FactionA: "ai", FactionB: "t2", Score: -100, Stance: faction.StanceWar},
			faction.RelationKey("ai", "t3"): {FactionA: "ai", FactionB: "t3", Score: -100, Stance: faction.StanceWar},
			faction.RelationKey("ai", "t4"): {FactionA: "ai", FactionB: "t4", Score: -100, Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {ID: "ai_army", OwnerID: "ai", RegionID: "ai_cap", Units: []army.Unit{
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
				{TypeID: "inf", CurrentHP: 100},
			}},
			"t1_army": {ID: "t1_army", OwnerID: "t1", RegionID: "t1_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"t2_army": {ID: "t2_army", OwnerID: "t2", RegionID: "t2_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"t3_army": {ID: "t3_army", OwnerID: "t3", RegionID: "t3_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"t4_army": {ID: "t4_army", OwnerID: "t4", RegionID: "t4_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}

	aiHandleDiplomacy(gs, "ai")

	if got := gs.DiplomacyOfferQuotaUsed("ai"); got != 3 {
		t.Fatalf("AI bir turda en fazla 3 diplomasi teklifi göndermeliydi, got=%d", got)
	}
	if got := gs.DiplomacyOfferQuotaRemaining("ai"); got != 0 {
		t.Fatalf("AI teklif hakkı bitmiş olmalıydı, got=%d", got)
	}
}

func TestAIQueuesAllianceOfferAgainstSharedMajorThreat(t *testing.T) {
	gs := &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Gold: 600, Grain: 300},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic, Gold: 420, Grain: 220, AIAggressiveness: 55},
			"threat": {ID: "threat", NameTR: "Threat", Religion: religion.Catholic, Gold: 900, Grain: 500},
		},
		Regions: map[world.RegionID]*world.Region{
			"player_land": {ID: "player_land", OwnerID: "player", TradeCapacity: 6, Neighbors: []world.RegionID{"threat_east"}},
			"ai_land":     {ID: "ai_land", OwnerID: "ai_1", TradeCapacity: 4, Neighbors: []world.RegionID{"threat_west"}},
			"threat_west": {ID: "threat_west", OwnerID: "threat", TradeCapacity: 4, Neighbors: []world.RegionID{"ai_land"}},
			"threat_east": {ID: "threat_east", OwnerID: "threat", TradeCapacity: 4, Neighbors: []world.RegionID{"player_land"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: 28, Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "player_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ai_army":     {ID: "ai_army", OwnerID: "ai_1", RegionID: "ai_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"threat_army": {
				ID:       "threat_army",
				OwnerID:  "threat",
				RegionID: "threat_west",
				Units: []army.Unit{
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
				},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}

	aiHandleDiplomacy(gs, "ai_1")

	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("ortak büyük tehditte tek ittifak teklifi bekleniyordu, got=%d (%+v)", len(gs.DiplomaticOffers), gs.DiplomaticOffers)
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(diplomacy.ActionProposeAlliance) {
		t.Fatalf("beklenen ittifak teklifi, got=%+v", offer)
	}
	if !strings.Contains(offer.PriorityReason, "büyük") {
		t.Fatalf("öncelik sebebi ortak büyük tehdidi görünür kılmalıydı, got=%q", offer.PriorityReason)
	}
}

func TestAIDoesNotQueueAllianceWithoutRegionalOrStrategicBasis(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Gold: 600, Grain: 300},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic, Gold: 420, Grain: 220, AIAggressiveness: 55},
		},
		Regions: map[world.RegionID]*world.Region{
			"player_land": {ID: "player_land", OwnerID: "player", TradeCapacity: 6},
			"ai_land":     {ID: "ai_land", OwnerID: "ai_1", TradeCapacity: 4},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: 40, Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "player_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ai_army":     {ID: "ai_army", OwnerID: "ai_1", RegionID: "ai_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}

	aiHandleDiplomacy(gs, "ai_1")

	if len(gs.DiplomaticOffers) > 0 && gs.DiplomaticOffers[0].Action == string(diplomacy.ActionProposeAlliance) {
		t.Fatalf("uzak ve alakasiz hedefe alliance spam olmamaliydi, got=%+v", gs.DiplomaticOffers)
	}
}

func TestAIDoesNotOfferAllianceToTinyStateWithoutMeaningfulBenefit(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Kucuk Devlet", Religion: religion.Catholic, Gold: 120, Grain: 80},
			"ai_1":   {ID: "ai_1", NameTR: "Buyuk Guc", Religion: religion.Catholic, Gold: 900, Grain: 500, AIAggressiveness: 55},
		},
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player", TradeCapacity: 4, Neighbors: []world.RegionID{"a1"}},
			"a1": {ID: "a1", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"p1", "a2"}},
			"a2": {ID: "a2", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"a1", "a3"}},
			"a3": {ID: "a3", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"a2", "a4"}},
			"a4": {ID: "a4", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"a3"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: 40, Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "p1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ai_army_1": {
				ID:       "ai_army_1",
				OwnerID:  "ai_1",
				RegionID: "a1",
				Units: []army.Unit{
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
				},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}

	aiHandleDiplomacy(gs, "ai_1")

	if len(gs.DiplomaticOffers) > 0 && gs.DiplomaticOffers[0].Action == string(diplomacy.ActionProposeAlliance) {
		t.Fatalf("buyuk guc kucuk devlete menfaatsiz alliance atmamalıydi, got=%+v", gs.DiplomaticOffers)
	}
}

func TestAICancelsUnsupportedAllianceUnderExpansionPressure(t *testing.T) {
	gs := aiTestState()
	gs.Factions["ai_1"].AIExpansionTargets = []faction.FactionID{"ai_2"}
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StanceAllied
	rel.Score = 38

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StancePeace {
		t.Fatalf("genisleme baskisindaki anlamsiz alliance bozulmaliydi, got=%s", rel.Stance)
	}
}

func TestAICancelsAllianceWithTinyStateWhenBenefitIsTooLow(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"ai_1":   {ID: "ai_1", NameTR: "Buyuk Guc", Religion: religion.Catholic},
			"minor":  {ID: "minor", NameTR: "Kucuk Devlet", Religion: religion.Catholic},
			"other1": {ID: "other1", NameTR: "Other 1", Religion: religion.Catholic},
			"other2": {ID: "other2", NameTR: "Other 2", Religion: religion.Catholic},
			"other3": {ID: "other3", NameTR: "Other 3", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"minor_land": {ID: "minor_land", OwnerID: "minor", TradeCapacity: 4, Neighbors: []world.RegionID{"core_1"}},
			"core_1":     {ID: "core_1", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"minor_land", "core_2"}},
			"core_2":     {ID: "core_2", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"core_1", "core_3"}},
			"core_3":     {ID: "core_3", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"core_2", "core_4"}},
			"core_4":     {ID: "core_4", OwnerID: "ai_1", TradeCapacity: 6, Neighbors: []world.RegionID{"core_3"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "minor"): {FactionA: "ai_1", FactionB: "minor", Score: 40, Stance: faction.StanceAllied},
		},
		Armies: map[army.ArmyID]*army.Army{
			"minor_army": {ID: "minor_army", OwnerID: "minor", RegionID: "minor_land", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"core_army": {
				ID:       "core_army",
				OwnerID:  "ai_1",
				RegionID: "core_1",
				Units: []army.Unit{
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
					{TypeID: "inf", CurrentHP: 100},
				},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
	}

	aiHandleDiplomacy(gs, "ai_1")

	if rel := gs.Relations[faction.RelationKey("ai_1", "minor")]; rel == nil || rel.Stance != faction.StancePeace {
		t.Fatalf("menfaatsiz kucuk alliance bozulmaliydi, got=%+v", rel)
	}
}

func TestAIStartsTradeOnHealthyPeace(t *testing.T) {
	gs := aiTestState()
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StancePeace
	rel.Score = 20

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StanceTrade {
		t.Fatalf("AI ticaret başlatmalıydı, got=%s", rel.Stance)
	}
	if len(gs.TradeRoutes) != 2 {
		t.Fatalf("iki yönlü ticaret rotası bekleniyordu, got=%d", len(gs.TradeRoutes))
	}
}

func TestCoalitionUsesDiplomacyEngine(t *testing.T) {
	gs := aiTestState()
	gs.PlayerFactionID = "player"
	gs.Regions["p2"] = &world.Region{ID: "p2", OwnerID: "player"}
	gs.Regions["p3"] = &world.Region{ID: "p3", OwnerID: "player"}
	gs.Regions["p4"] = &world.Region{ID: "p4", OwnerID: "player"}
	gs.Regions["p5"] = &world.Region{ID: "p5", OwnerID: "player"}
	gs.Regions["p6"] = &world.Region{ID: "p6", OwnerID: "player"}
	gs.Regions["p7"] = &world.Region{ID: "p7", OwnerID: "player"}
	gs.Regions["p8"] = &world.Region{ID: "p8", OwnerID: "player"}
	gs.Regions["a1"].Neighbors = []world.RegionID{"p1"}
	gs.Regions["p1"].Neighbors = []world.RegionID{"a1", "b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"p1"}

	FormCoalitionAgainstPlayer(gs, "ai_1")

	playerRel := gs.Relations[faction.RelationKey("ai_1", "player")]
	if playerRel.Stance != faction.StanceWar {
		t.Fatalf("koalisyon oyuncuya savaş açmalıydı, got=%s", playerRel.Stance)
	}
	allyRel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if allyRel.Stance != faction.StanceAllied {
		t.Fatalf("koalisyon AI ittifakı kurmalıydı, got=%s", allyRel.Stance)
	}
}

func TestAIMoveArmyEmbarksIntoFriendlyTransportFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     1,
		Regions: map[world.RegionID]*world.Region{
			"ai_land":    {ID: "ai_land", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_1"}},
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"ai_land", "enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "ai_land",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -30, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	moveArmy(gs, gs.Armies["ai_army"])

	if _, ok := gs.Armies["ai_army"]; ok {
		t.Fatalf("AI kara ordusu embark sonrası haritadan kalkmalıydı")
	}
	fleet := gs.Armies["ai_fleet"]
	if fleet == nil || len(fleet.EmbarkedUnits) != 1 {
		t.Fatalf("AI filosunda embark birimi bekleniyordu, got=%+v", fleet)
	}
}

func TestAIMoveArmyDisembarksFromFleet(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     10,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":   {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"ai_land"}},
			"ai_land": {ID: "ai_land", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if len(gs.Armies["ai_fleet"].EmbarkedUnits) != 0 {
		t.Fatalf("AI çıkarma sonrası filonun embarked birimleri boş olmalı")
	}
	if _, ok := gs.Armies["army_ai_1_11"]; !ok {
		t.Fatalf("çıkarma sonrası yeni kara ordusu bekleniyordu")
	}
}

func TestAIMoveArmyDisembarksToEnemyCoastWhenAtWar(t *testing.T) {
	rand.Seed(1)
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     40,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}, {TypeID: "elite", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -70, Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"elite":     {ID: "elite", Embarkable: true, Attack: 120, Defense: 90, Morale: 90},
			"transport": aiTestTransportType(),
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if gs.Regions["enemy_land"].OwnerID != "ai_1" {
		t.Fatalf("savaşta düşman kıyı çıkarma sonrası sahiplik değişmeli, got=%s", gs.Regions["enemy_land"].OwnerID)
	}
	if _, ok := gs.Armies["army_ai_1_41"]; !ok {
		t.Fatalf("savaşta düşman kıyı çıkarma sonrası kara ordusu oluşmalı")
	}
}

func TestAIMoveArmyDisembarksToEnemyFortressAndStartsSiege(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     70,
		Regions: map[world.RegionID]*world.Region{
			"sea_1": {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"fort"}},
			"fort": {
				ID: "fort", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID: "ai_fleet", OwnerID: "ai_1", RegionID: "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3, MaxMovePoints: 3, IsNaval: true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Stance: faction.StanceWar},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	step := executeMove(gs, gs.Armies["ai_fleet"], "fort", "ai_1")

	if gs.Regions["fort"].OwnerID != "player" {
		t.Fatalf("AI kale kıyısına çıkarma ile bölgeyi hemen almamalıydı, got=%s", gs.Regions["fort"].OwnerID)
	}
	if gs.SiegeAt("fort") == nil {
		t.Fatal("AI kale kıyısında kuşatma başlatmalıydı")
	}
	if step.step.Kind != TurnStepDisembark {
		t.Fatalf("AI çıkarma sonucu disembark step'i bekleniyordu, got=%s", step.step.Kind)
	}
}

func TestAIMoveArmyDoesNotDisembarkToEnemyCoastAtPeace(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     50,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":      {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"enemy_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: 5, Stance: faction.StancePeace},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if gs.Regions["enemy_land"].OwnerID != "player" {
		t.Fatalf("barışta düşman kıyıya çıkarma olmamalı, got=%s", gs.Regions["enemy_land"].OwnerID)
	}
	if _, ok := gs.Armies["army_ai_1_51"]; ok {
		t.Fatalf("barışta düşman kıyıya yeni kara ordusu oluşmamalı")
	}
	if len(gs.Armies["ai_fleet"].EmbarkedUnits) != 1 {
		t.Fatalf("barışta çıkarma olmamalı, cargo korunmalı")
	}
}

func TestAIMoveArmyDisembarksToNeutralCoastAndClaimsRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     60,
		Regions: map[world.RegionID]*world.Region{
			"sea_1":        {ID: "sea_1", IsSea: true, Neighbors: []world.RegionID{"neutral_land"}},
			"neutral_land": {ID: "neutral_land", Religion: "catholic", Neighbors: []world.RegionID{"sea_1"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_fleet": {
				ID:            "ai_fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_1",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    3,
				MaxMovePoints: 3,
				IsNaval:       true,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	moveArmy(gs, gs.Armies["ai_fleet"])

	if gs.Regions["neutral_land"].OwnerID != "ai_1" {
		t.Fatalf("AI sahipsiz kıyıyı sahiplenmeliydi, got=%s", gs.Regions["neutral_land"].OwnerID)
	}
	if _, ok := gs.Armies["army_ai_1_61"]; !ok {
		t.Fatalf("AI çıkarma sonrası kara ordusu oluşmalı")
	}
}

func TestChooseBestMovePrefersSeaWithHigherHostilePressure(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea_home":   {ID: "sea_home", IsSea: true, Neighbors: []world.RegionID{"sea_hot", "sea_cold"}},
			"sea_hot":    {ID: "sea_hot", IsSea: true, Neighbors: []world.RegionID{"sea_home", "enemy_land"}},
			"sea_cold":   {ID: "sea_cold", IsSea: true, Neighbors: []world.RegionID{"sea_home", "ally_land"}},
			"enemy_land": {ID: "enemy_land", OwnerID: "player", Neighbors: []world.RegionID{"sea_hot"}},
			"ally_land":  {ID: "ally_land", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_cold"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -80, Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:            "fleet",
				OwnerID:       "ai_1",
				RegionID:      "sea_home",
				Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
				EmbarkedUnits: []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
				IsNaval:       true,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":       {ID: "inf", Embarkable: true, Attack: 10, Defense: 10, Morale: 50},
			"transport": aiTestTransportType(),
		},
	}

	target := chooseBestMove(gs, gs.Armies["fleet"])
	if target != "sea_hot" {
		t.Fatalf("yüksek baskılı deniz hedefi sea_hot olmalıydı, got=%s", target)
	}
}

func TestChooseBestMoveRelievesOverloadedFriendlyRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"crowded":  {ID: "crowded", OwnerID: "ai_1", BaseGrainOutput: 2, Neighbors: []world.RegionID{"relief", "frontier"}},
			"relief":   {ID: "relief", OwnerID: "ai_1", BaseGrainOutput: 12, Neighbors: []world.RegionID{"crowded"}},
			"frontier": {ID: "frontier", OwnerID: "player", BaseGrainOutput: 4, Neighbors: []world.RegionID{"crowded"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Grain: 0},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni, Grain: 10},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -20, Stance: faction.StancePeace},
		},
		Armies: map[army.ArmyID]*army.Army{
			"stack_1": {ID: "stack_1", OwnerID: "ai_1", RegionID: "crowded", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}, OverCapacityTurns: 2},
			"stack_2": {ID: "stack_2", OwnerID: "ai_1", RegionID: "crowded", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2, Attack: 10, Defense: 10, Morale: 50},
		},
	}

	target := chooseBestMove(gs, gs.Armies["stack_1"])
	if target != "relief" {
		t.Fatalf("AI overloaded dost bölgeden relief'e dağılmalıydı, got=%s", target)
	}
}

func TestAIConsolidateArmiesSkipsOverloadedLandRegion(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"crowded": {ID: "crowded", OwnerID: "ai_1", BaseGrainOutput: 1},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ai_1": {ID: "ai_1", Grain: 0},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "ai_1", RegionID: "crowded", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			"a2": {ID: "a2", OwnerID: "ai_1", RegionID: "crowded", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 3},
		},
	}

	aiConsolidateArmies(gs, "ai_1")

	if len(gs.Armies) != 2 {
		t.Fatalf("overloaded kara bölgede AI orduları birleştirmemeli, got=%d", len(gs.Armies))
	}
}

func TestAINavalStrategyAllowsThirdFleetDuringWarPressure(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		NextArmySeq:     20,
		Regions: map[world.RegionID]*world.Region{
			"coast_a":     {ID: "coast_a", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_a"}, Buildings: []string{"port"}},
			"coast_b":     {ID: "coast_b", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_b"}, Buildings: []string{"port"}},
			"coast_c":     {ID: "coast_c", OwnerID: "ai_1", Neighbors: []world.RegionID{"sea_c"}, Buildings: []string{"port"}},
			"enemy_coast": {ID: "enemy_coast", OwnerID: "player", Neighbors: []world.RegionID{"sea_c"}},
			"sea_a":       {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"coast_a"}},
			"sea_b":       {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"coast_b"}},
			"sea_c":       {ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"coast_c", "enemy_coast"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni, Gold: 500, Grain: 200, Timber: 200, Iron: 50, Stone: 50},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -80, Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_1": {ID: "fleet_1", OwnerID: "ai_1", RegionID: "sea_a", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}},
			"fleet_2": {ID: "fleet_2", OwnerID: "ai_1", RegionID: "sea_b", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port"},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10, GoldCost: 120, TimberCost: 20},
		},
	}

	aiNavalStrategy(gs, "ai_1")

	navalCount := 0
	for _, a := range gs.Armies {
		if a.OwnerID != "ai_1" || !a.IsNaval {
			continue
		}
		navalCount++
	}
	if navalCount != 2 {
		t.Fatalf("transport artık anında filo spawn etmemeli, got=%d", navalCount)
	}
	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("savaş baskısında transport üretim emri bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	order := gs.ProductionQueue[0]
	if order.Kind != aiProductionKindUnit || order.FactionID != "ai_1" || order.RegionID != "coast_c" || order.TypeID != "transport" {
		t.Fatalf("transport baskısı yüksek coast_c limanında kuyruğa girmeli, got=%+v", order)
	}
}

func TestAINavalStrategyQueuesEscortForPendingTransport(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"coast_a": {ID: "coast_a", OwnerID: "ai_1", Buildings: []string{"port", "port", "port"}, Neighbors: []world.RegionID{"sea_a"}},
			"sea_a":   {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"coast_a", "enemy_coast"}},
			"enemy_coast": {
				ID:        "enemy_coast",
				OwnerID:   "player",
				Neighbors: []world.RegionID{"sea_a"},
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni, Gold: 1000, Grain: 200, Timber: 200, Iron: 80, Stone: 80, Research: faction.ResearchState{Completed: map[string]bool{"navigation": true, "naval_doctrine": true}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -80, Stance: faction.StanceWar},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 3},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10, GoldCost: 200, TimberCost: 20, RequiredTech: []string{"navigation"}, TurnsRequired: 2},
			"warship":   {ID: "warship", Category: army.CategoryNavalWar, GoldCost: 400, TimberCost: 34, RequiredTech: []string{"naval_doctrine"}, RequiredBldg: "port", RequiredBldgLevel: 3, TurnsRequired: 4},
		},
	}

	aiNavalStrategy(gs, "ai_1")

	if len(gs.ProductionQueue) != 2 {
		t.Fatalf("pending transport için escort da bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	if gs.ProductionQueue[0].TypeID != "transport" || gs.ProductionQueue[1].TypeID != "warship" {
		t.Fatalf("transport+escort sırası bekleniyordu, got=%+v", gs.ProductionQueue)
	}
}

func TestAINavalStrategyQueuesMultipleEscortsForMultipleFronts(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"coast_a": {ID: "coast_a", OwnerID: "ai_1", Buildings: []string{"port", "port", "port"}, Neighbors: []world.RegionID{"sea_a"}},
			"coast_b": {ID: "coast_b", OwnerID: "ai_1", Buildings: []string{"port", "port", "port"}, Neighbors: []world.RegionID{"sea_b"}},
			"sea_a":   {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"coast_a", "enemy_a"}},
			"sea_b":   {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"coast_b", "enemy_b"}},
			"enemy_a": {ID: "enemy_a", OwnerID: "player", Neighbors: []world.RegionID{"sea_a"}},
			"enemy_b": {ID: "enemy_b", OwnerID: "player", Neighbors: []world.RegionID{"sea_b"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni, Gold: 2500, Grain: 400, Timber: 400, Iron: 150, Stone: 150, Research: faction.ResearchState{Completed: map[string]bool{"navigation": true, "naval_doctrine": true}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -80, Stance: faction.StanceWar},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 3},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_fleet": {ID: "enemy_fleet", OwnerID: "player", RegionID: "sea_a", IsNaval: true, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10, GoldCost: 200, TimberCost: 20, RequiredTech: []string{"navigation"}, TurnsRequired: 2},
			"warship":   {ID: "warship", Category: army.CategoryNavalWar, GoldCost: 400, TimberCost: 34, RequiredTech: []string{"naval_doctrine"}, RequiredBldg: "port", RequiredBldgLevel: 3, TurnsRequired: 4},
		},
	}

	aiNavalStrategy(gs, "ai_1")

	if len(gs.ProductionQueue) != 3 {
		t.Fatalf("iki cephede bir transport ve iki escort bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	if gs.ProductionQueue[0].TypeID != "transport" {
		t.Fatalf("ilk emir transport olmalıydı, got=%+v", gs.ProductionQueue[0])
	}
	if gs.ProductionQueue[1].TypeID != "warship" || gs.ProductionQueue[2].TypeID != "warship" {
		t.Fatalf("escort emirleri warship olmalıydı, got=%+v", gs.ProductionQueue)
	}
	regionCounts := map[world.RegionID]int{}
	for _, order := range gs.ProductionQueue[1:] {
		regionCounts[order.RegionID]++
	}
	if regionCounts["coast_a"] != 1 || regionCounts["coast_b"] != 1 {
		t.Fatalf("escortlar iki farklı cepheye yayılmalıydı, got=%+v", regionCounts)
	}
}

func TestAIBuildsBarracksThroughProductionQueue(t *testing.T) {
	gs := aiTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"barracks": {ID: "barracks", GoldCost: 150, MaxPerRegion: 1, TurnsRequired: 2},
	}
	gs.Factions["ai_1"].Gold = 150

	aiBuildBarracks(gs, "ai_1", economy.ResourceCost{Gold: 150})

	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("kışla anında eklenmek yerine kuyruğa girmeli, got=%d", len(gs.ProductionQueue))
	}
	if got := aiBuildingLevel(gs.Regions["a1"], "barracks"); got != 0 {
		t.Fatalf("kışla tamamlanmadan region.Buildings'e eklenmemeli, got=%d", got)
	}
	order := gs.ProductionQueue[0]
	if order.Kind != aiProductionKindBuilding || order.FactionID != "ai_1" || order.RegionID != "a1" || order.TypeID != "barracks" || order.TurnsLeft != 2 {
		t.Fatalf("beklenmeyen kışla üretim emri: %+v", order)
	}
}

func TestAIRecruitQueuesUnitInsteadOfAddingImmediately(t *testing.T) {
	gs := aiTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"barracks": {ID: "barracks", GoldCost: 150, MaxPerRegion: 1, TurnsRequired: 2},
	}
	gs.Regions["a1"].Buildings = []string{"barracks"}
	gs.Factions["ai_1"].Gold = 500
	gs.Factions["ai_1"].Grain = 500
	gs.UnitTypes["militia"] = &army.UnitType{ID: "militia", GoldCost: 60, TurnsRequired: 1, RequiredBldg: "barracks", Attack: 6, Defense: 5, Morale: 30}
	beforeUnits := len(gs.Armies["ai1_army"].Units)

	if !aiRecruitOne(gs, "ai_1") {
		t.Fatalf("AI uygun bölgede birim üretim emri açabilmeli")
	}

	if len(gs.Armies["ai1_army"].Units) != beforeUnits {
		t.Fatalf("birim tamamlanmadan orduya anında eklenmemeli")
	}
	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("birim üretim emri bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	order := gs.ProductionQueue[0]
	if order.Kind != aiProductionKindUnit || order.FactionID != "ai_1" || order.RegionID != "a1" || order.TypeID != "militia" {
		t.Fatalf("beklenmeyen birim üretim emri: %+v", order)
	}
}

func TestAIRecruitPrefersRegionWithFreeBarracksThroughput(t *testing.T) {
	gs := aiTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"barracks": {ID: "barracks", GoldCost: 150, MaxPerRegion: 2, TurnsRequired: 2},
	}
	gs.Regions["a1"].Buildings = []string{"barracks"}
	gs.Regions["a2"] = &world.Region{ID: "a2", OwnerID: "ai_1", TradeCapacity: 4, Buildings: []string{"barracks"}}
	gs.Armies["ai1_army_2"] = &army.Army{ID: "ai1_army_2", OwnerID: "ai_1", RegionID: "a2", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}}
	gs.Factions["ai_1"].Gold = 500
	gs.Factions["ai_1"].Grain = 500
	gs.UnitTypes["militia"] = &army.UnitType{ID: "militia", GoldCost: 60, TurnsRequired: 1, RequiredBldg: "barracks", Attack: 6, Defense: 5, Morale: 30}
	gs.ProductionQueue = []state.ProductionOrder{
		{ID: "prod_1", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "a1", TypeID: "militia", TurnsLeft: 1},
	}

	if !aiRecruitOne(gs, "ai_1") {
		t.Fatalf("AI serbest throughput'u olan ikinci bölgeye emir yazabilmeliydi")
	}

	if len(gs.ProductionQueue) != 2 {
		t.Fatalf("iki emir bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	last := gs.ProductionQueue[1]
	if last.RegionID != "a2" || last.TypeID != "militia" {
		t.Fatalf("AI doygun olmayan bölgeyi seçmeliydi, got=%+v", last)
	}
}

func TestAIRecruitDoesNotQueueWhenAllBarracksThroughputIsFull(t *testing.T) {
	gs := aiTestState()
	gs.BuildingTypes = map[string]*city.Building{
		"barracks": {ID: "barracks", GoldCost: 150, MaxPerRegion: 1, TurnsRequired: 2},
	}
	gs.Regions["a1"].Buildings = []string{"barracks"}
	gs.Factions["ai_1"].Gold = 500
	gs.Factions["ai_1"].Grain = 500
	gs.UnitTypes["militia"] = &army.UnitType{ID: "militia", GoldCost: 60, TurnsRequired: 1, RequiredBldg: "barracks", Attack: 6, Defense: 5, Morale: 30}
	gs.ProductionQueue = []state.ProductionOrder{
		{ID: "prod_1", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "a1", TypeID: "militia", TurnsLeft: 1},
	}

	if aiRecruitOne(gs, "ai_1") {
		t.Fatalf("AI dolu kışla throughput'una aynı tur ikinci kara emri yazmamalıydı")
	}
	if len(gs.ProductionQueue) != 1 {
		t.Fatalf("ek üretim emri açılmamalıydı, got=%d", len(gs.ProductionQueue))
	}
}

func TestAINavalStrategySkipsSaturatedPortLane(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"coast_a": {ID: "coast_a", OwnerID: "ai_1", Buildings: []string{"port", "port"}, Neighbors: []world.RegionID{"sea_a"}},
			"coast_b": {ID: "coast_b", OwnerID: "ai_1", Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_b"}},
			"coast_c": {ID: "coast_c", OwnerID: "ai_1", Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_c"}},
			"sea_a":   {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"coast_a", "enemy_a"}},
			"sea_b":   {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"coast_b"}},
			"sea_c":   {ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"coast_c"}},
			"enemy_a": {ID: "enemy_a", OwnerID: "player", Neighbors: []world.RegionID{"sea_a"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni, Gold: 1500, Grain: 300, Timber: 300, Iron: 120, Stone: 120},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -80, Stance: faction.StanceWar},
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 1},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10, GoldCost: 200, TimberCost: 20, TurnsRequired: 2},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "coast_a", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_2", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "coast_a", TypeID: "transport", TurnsLeft: 2},
		},
	}

	aiNavalStrategy(gs, "ai_1")

	if len(gs.ProductionQueue) != 3 {
		t.Fatalf("bir yeni deniz emri bekleniyordu, got=%d", len(gs.ProductionQueue))
	}
	last := gs.ProductionQueue[2]
	if last.RegionID != "coast_b" || last.TypeID != "transport" {
		t.Fatalf("dolu liman hattı yerine boş liman hattı seçilmeliydi, got=%+v", last)
	}
}

func TestAIPendingNavalFleetCountTreatsSameSeaQueueAsSingleFleet(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"coast_a": {ID: "coast_a", OwnerID: "ai_1", Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_a"}},
			"coast_b": {ID: "coast_b", OwnerID: "ai_1", Buildings: []string{"port"}, Neighbors: []world.RegionID{"sea_a"}},
			"sea_a":   {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"coast_a", "coast_b"}},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"ai_1": {ID: "ai_1", NameTR: "AI 1"},
		},
		UnitTypes: map[string]*army.UnitType{
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		ProductionQueue: []state.ProductionOrder{
			{ID: "prod_1", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "coast_a", TypeID: "transport", TurnsLeft: 1},
			{ID: "prod_2", Kind: aiProductionKindUnit, FactionID: "ai_1", RegionID: "coast_b", TypeID: "transport", TurnsLeft: 2},
		},
	}

	if got := aiPendingNavalFleetCount(gs, "ai_1"); got != 1 {
		t.Fatalf("aynı denize bakan pending emirler tek filo sayılmalıydı, got=%d", got)
	}
}

func TestAIDeclaresWarOnWeakBorderTarget(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 2
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 60
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -35
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["ai2_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}

	aiHandleDiplomacy(gs, "ai_1")

	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if rel.Stance != faction.StanceWar {
		t.Fatalf("AI zayıf sınır hedefe savaş açmalıydı, got=%s score=%d", rel.Stance, rel.Score)
	}
}

func TestAIDeclaresWarOnExpansionTargetWithMildHostility(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 2
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 60
	gs.Factions["ai_1"].AIExpansionTargets = []faction.FactionID{"ai_2"}
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -5
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["ai2_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}

	aiHandleDiplomacy(gs, "ai_1")

	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if rel.Stance != faction.StanceWar {
		t.Fatalf("AI genişleme hedefindeki sınır komşusuna savaş açmalıydı, got=%s score=%d", rel.Stance, rel.Score)
	}
}

func TestAIDoesNotDeclareWarOnWarmPeaceWithoutExpansionTarget(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 3
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 75
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = 10
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["ai2_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}

	aiHandleDiplomacy(gs, "ai_1")

	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if rel.Stance == faction.StanceWar {
		t.Fatalf("AI hedef listesinde olmayan sıcak barışı savaş için bozmamalıydı")
	}
}

func TestAIDoesNotDeclareWarWhenOutmatched(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 2
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 70
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -60
	gs.Armies["ai1_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}
	gs.Armies["ai2_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}

	aiHandleDiplomacy(gs, "ai_1")

	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	if rel.Stance == faction.StanceWar {
		t.Fatalf("güçsüz AI üstün hedefe savaş açmamalıydı")
	}
}

func TestAIDoesNotBreakTradeForWar(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 3
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 70
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	rel := gs.Relations[faction.RelationKey("ai_1", "ai_2")]
	rel.Stance = faction.StanceTrade
	rel.Score = 40
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["ai2_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}

	aiHandleDiplomacy(gs, "ai_1")

	if rel.Stance != faction.StanceTrade {
		t.Fatalf("AI aktif ticaret ilişkisini savaş için bozmamalıydı, got=%s", rel.Stance)
	}
}

func TestAIDeclaresOnlyOneOpportunityWarPerTurn(t *testing.T) {
	gs := aiTestState()
	gs.Difficulty = 3
	gs.Month = 1
	gs.Factions["ai_1"].AIAggressiveness = 70
	gs.Factions["ai_3"] = &faction.Faction{ID: "ai_3", NameTR: "AI 3", Religion: religion.Catholic, Grain: 100, Gold: 100}
	gs.Regions["a1"].Neighbors = []world.RegionID{"b1", "c1"}
	gs.Regions["b1"].Neighbors = []world.RegionID{"a1"}
	gs.Regions["c1"] = &world.Region{ID: "c1", OwnerID: "ai_3", Neighbors: []world.RegionID{"a1"}, TradeCapacity: 4}
	gs.Relations[faction.RelationKey("ai_1", "ai_2")].Score = -35
	gs.Relations[faction.RelationKey("ai_1", "ai_3")] = &faction.Relation{
		FactionA: "ai_1",
		FactionB: "ai_3",
		Score:    -35,
		Stance:   faction.StancePeace,
	}
	gs.Armies["ai1_army"].Units = []army.Unit{
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
		{TypeID: "inf", CurrentHP: 100},
	}
	gs.Armies["ai2_army"].Units = []army.Unit{{TypeID: "inf", CurrentHP: 100}}
	gs.Armies["ai3_army"] = &army.Army{ID: "ai3_army", OwnerID: "ai_3", RegionID: "c1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}}

	aiEvaluateWarOpportunities(gs, "ai_1")

	wars := 0
	for _, rel := range gs.Relations {
		if rel != nil && rel.Stance == faction.StanceWar && (rel.FactionA == "ai_1" || rel.FactionB == "ai_1") {
			wars++
		}
	}
	if wars != 1 {
		t.Fatalf("AI bir turda tek fırsat savaşı açmalıydı, got=%d", wars)
	}
}

func aiTestState() *state.GameState {
	return &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic, Grain: 100, Gold: 100},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Catholic, Grain: 100, Gold: 100},
			"ai_2":   {ID: "ai_2", NameTR: "AI 2", Religion: religion.Catholic, Grain: 100, Gold: 100},
		},
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"p1": {ID: "p1", OwnerID: "player", TradeCapacity: 4},
			"a1": {ID: "a1", OwnerID: "ai_1", TradeCapacity: 4},
			"b1": {ID: "b1", OwnerID: "ai_2", TradeCapacity: 4},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "ai_2"):   {FactionA: "ai_1", FactionB: "ai_2", Score: 25, Stance: faction.StancePeace},
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -10, Stance: faction.StancePeace},
			faction.RelationKey("ai_2", "player"): {FactionA: "ai_2", FactionB: "player", Score: -10, Stance: faction.StancePeace},
		},
		TradeRoutes: []*economy.TradeRoute{},
		Armies: map[army.ArmyID]*army.Army{
			"ai1_army": {ID: "ai1_army", OwnerID: "ai_1", RegionID: "a1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"ai2_army": {ID: "ai2_army", OwnerID: "ai_2", RegionID: "b1", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 12, Defense: 10, Morale: 60},
		},
	}
}
