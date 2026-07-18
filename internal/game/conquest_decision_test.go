package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func conquestDecisionTestGame() *Game {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: "sunni"},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Religion: "catholic"},
			"third":  {ID: "third", NameTR: "Üçüncü", Religion: "orthodox"},
		},
		Regions: map[world.RegionID]*world.Region{
			"enemy_cap":  {ID: "enemy_cap", OwnerID: "enemy", NameTR: "Düşman Başkenti"},
			"player_cap": {ID: "player_cap", OwnerID: "player", NameTR: "Oyuncu Başkenti"},
			"third_cap":  {ID: "third_cap", OwnerID: "third", NameTR: "Üçüncü Başkent"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar, Score: -80},
			faction.RelationKey("player", "third"): {FactionA: "player", FactionB: "third", Stance: faction.StanceWar, Score: -80},
			faction.RelationKey("enemy", "third"):  {FactionA: "enemy", FactionB: "third", Stance: faction.StanceTrade, Score: 25},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {ID: "atk", OwnerID: "player", RegionID: "player_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"def": {ID: "def", OwnerID: "enemy", RegionID: "enemy_cap", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
	}
	return &Game{gs: gs, renderer: &render.Renderer{}}
}

func TestQueueConquestDecisionForPlayersLastEnemyRegion(t *testing.T) {
	g := conquestDecisionTestGame()

	if ok := g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], false); !ok {
		t.Fatal("son bölge için savaş sonrası karar kuyruğa alınmalıydı")
	}
	if len(g.pendingConquestDecisions) != 1 {
		t.Fatalf("tek karar bekleniyordu, got=%d", len(g.pendingConquestDecisions))
	}
	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "enemy" {
		t.Fatalf("karar verilmeden bölge sahibi değişmemeliydi, got=%s", got)
	}
}

func TestResolvePendingConquestDecisionAnnexesDefender(t *testing.T) {
	g := conquestDecisionTestGame()
	if ok := g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], false); !ok {
		t.Fatal("karar kuyruğa alınamadı")
	}

	g.resolvePendingConquestDecision(false)

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "player" {
		t.Fatalf("annex seçiminde bölge oyuncuya geçmeliydi, got=%s", got)
	}
	if !g.gs.Factions["enemy"].IsEliminated {
		t.Fatal("son bölge ilhak edilince düşman elenmeliydi")
	}
	if len(g.pendingConquestDecisions) != 0 {
		t.Fatalf("karar kuyruğu boşalmalıydı, got=%d", len(g.pendingConquestDecisions))
	}
}

func TestResolvePendingConquestDecisionVassalizesDefender(t *testing.T) {
	g := conquestDecisionTestGame()
	if ok := g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], false); !ok {
		t.Fatal("karar kuyruğa alınamadı")
	}

	g.resolvePendingConquestDecision(true)

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "enemy" {
		t.Fatalf("vassal seçiminde bölge yerel devlette kalmalıydı, got=%s", got)
	}
	if got := g.gs.Factions["enemy"].OverlordID; got != "player" {
		t.Fatalf("düşman oyuncuya bağlanmalıydı, got=%s", got)
	}
	if rel := g.gs.Relations[faction.RelationKey("player", "enemy")]; rel == nil || rel.Stance != faction.StanceAllied {
		t.Fatalf("overlord-vassal ilişkisi allied olmalıydı, got=%+v", rel)
	}
}

func TestAnnexVassalTransfersRealmAssets(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Factions["enemy"].OverlordID = "player"
	g.gs.Factions["enemy"].Gold = 90
	g.gs.Factions["enemy"].Grain = 45
	g.gs.Regions["enemy_second"] = &world.Region{ID: "enemy_second", OwnerID: "enemy", NameTR: "İkinci Bölge"}
	g.gs.ProductionQueue = []state.ProductionOrder{
		{ID: "enemy_build", Kind: "building", FactionID: "enemy", RegionID: "enemy_second", TypeID: "farm", TurnsLeft: 2},
	}
	playerGold := g.gs.Factions["player"].Gold
	playerGrain := g.gs.Factions["player"].Grain

	g.annexVassal("enemy")

	if !g.gs.Factions["enemy"].IsEliminated {
		t.Fatal("ilhak edilen vassal elenmiş olarak işaretlenmeli")
	}
	for _, rid := range []world.RegionID{"enemy_cap", "enemy_second"} {
		if got := g.gs.Regions[rid].OwnerID; got != "player" {
			t.Fatalf("%s oyuncuya geçmeliydi, got=%s", rid, got)
		}
	}
	if got := g.gs.Armies["def"].OwnerID; got != "player" {
		t.Fatalf("vassal ordusu oyuncuya devredilmeliydi, got=%s", got)
	}
	if got := g.gs.Factions["player"].Gold; got != playerGold+90 {
		t.Fatalf("vassal altını devredilmeliydi, got=%d", got)
	}
	if got := g.gs.Factions["player"].Grain; got != playerGrain+45 {
		t.Fatalf("vassal tahılı devredilmeliydi, got=%d", got)
	}
	if got := g.gs.ProductionQueue[0].FactionID; got != "player" {
		t.Fatalf("vassal üretim emri oyuncuya devredilmeliydi, got=%s", got)
	}
}

func TestCaptureBesiegedRegionQueuesDecisionOnFinalProvince(t *testing.T) {
	g := conquestDecisionTestGame()
	attacker := g.gs.Armies["atk"]
	target := g.gs.Regions["enemy_cap"]

	collapse, prompted := g.captureBesiegedRegion(attacker, target, true)

	if !prompted {
		t.Fatal("kuşatma fethi son bölgede savaş sonrası karar üretmeliydi")
	}
	if collapse.FactionID != "" {
		t.Fatalf("karar öncesi ilhak uygulanmamalıydı, got=%+v", collapse)
	}
	if got := target.OwnerID; got != "enemy" {
		t.Fatalf("karar öncesi bölge sahibi değişmemeliydi, got=%s", got)
	}
	if len(g.pendingConquestDecisions) != 1 {
		t.Fatalf("tek karar bekleniyordu, got=%d", len(g.pendingConquestDecisions))
	}
}

func TestCaptureBesiegedRegionAIAppliesHybridVassalPolicy(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.ScenarioID = "1300_ottoman_rise"
	g.gs.Factions["ai"] = &faction.Faction{ID: "ai", NameTR: "AI", AIAggressiveness: 60}
	g.gs.Factions["enemy"].AIAggressiveness = 50
	g.gs.Regions["ai_cap"] = &world.Region{ID: "ai_cap", OwnerID: "ai"}
	g.gs.Armies["atk"].OwnerID = "ai"
	g.gs.Armies["atk"].RegionID = "ai_cap"
	g.gs.Armies["atk"].Units = append(g.gs.Armies["atk"].Units, army.Unit{TypeID: "inf", CurrentHP: 100})
	g.gs.UnitTypes = map[string]*army.UnitType{
		"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
	}
	g.gs.Relations[faction.RelationKey("ai", "enemy")] = &faction.Relation{FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar, Score: -80}
	g.gs.AIPlans = map[faction.FactionID]*state.AIPlanState{
		"ai": {
			ObjectiveID:        "unite_anatolian_beyliks",
			Kind:               state.AIObjectiveExpand,
			TargetFactionID:    "enemy",
			TargetRegionIDs:    []world.RegionID{"enemy_cap"},
			AllowVassalization: true,
		},
	}

	collapse, prompted := g.captureBesiegedRegion(g.gs.Armies["atk"], g.gs.Regions["enemy_cap"], false)

	if prompted || collapse.FactionID != "" {
		t.Fatalf("AI vassallığında oyuncu kararı veya eliminasyon olmamalı: prompted=%v collapse=%+v", prompted, collapse)
	}
	if g.gs.Regions["enemy_cap"].OwnerID != "enemy" || g.gs.Factions["enemy"].OverlordID != "ai" {
		t.Fatalf("AI hibrit vassal kararı uygulanmadı: owner=%s overlord=%s", g.gs.Regions["enemy_cap"].OwnerID, g.gs.Factions["enemy"].OverlordID)
	}
}
