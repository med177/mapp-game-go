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

func TestCaptureUnfortifiedRegionDirectlyAnnexesHeldEnemyRegion(t *testing.T) {
	g := conquestDecisionTestGame()
	g.renderer = nil
	g.gs.Armies["atk"].RegionID = "enemy_cap"
	g.gs.Armies["def"].RegionID = "player_cap"

	g.captureUnfortifiedRegion("atk", "enemy_cap")

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "player" {
		t.Fatalf("tahkimatsız bölge doğrudan oyuncuya geçmeliydi, got=%s", got)
	}
	if len(g.pendingConquestDecisions) != 0 {
		t.Fatalf("doğrudan fetih savaş sonrası karar kuyruğu açmamalıydı, got=%d", len(g.pendingConquestDecisions))
	}
}

func TestCaptureUnfortifiedRegionRejectsHeldEnemyArmy(t *testing.T) {
	g := conquestDecisionTestGame()
	g.renderer = nil
	g.gs.Armies["atk"].RegionID = "enemy_cap"

	g.captureUnfortifiedRegion("atk", "enemy_cap")

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "enemy" {
		t.Fatalf("düşman ordusu varken bölge doğrudan ele geçirilmemeliydi, got=%s", got)
	}
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

func TestSuccessorMetadataQueuesThreeWayDecisionAfterBattle(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Factions["successor"] = &faction.Faction{ID: "successor", NameTR: "Ardıl", IsEliminated: true}
	g.gs.Regions["enemy_cap"].SuccessorFactionID = "successor"

	if ok := g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], true); !ok {
		t.Fatal("ardıl metadata'sı olan fetih için karar kuyruğa alınmalıydı")
	}
	if got := g.pendingConquestDecisions[0].SuccessorFactionID; got != "successor" {
		t.Fatalf("ardıl fraksiyon karara taşınmadı: %q", got)
	}
}

func TestSuccessorDecisionReleaseTransfersRegionAndDefeatsPreviousOwner(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Factions["successor"] = &faction.Faction{ID: "successor", NameTR: "Ardıl", IsEliminated: true}
	g.gs.Regions["enemy_cap"].SuccessorFactionID = "successor"
	if !g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], false) {
		t.Fatal("ardıl kararı kuyruğa alınamadı")
	}

	g.resolvePendingSuccessorDecision(successorDecisionRelease)

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "successor" {
		t.Fatalf("serbest bırakmada bölge ardıla geçmeliydi, got=%s", got)
	}
	if !g.gs.Factions["enemy"].IsEliminated {
		t.Fatal("bölgesi alınan eski sahip elenmeliydi")
	}
	if rel := g.gs.Relations[faction.RelationKey("player", "successor")]; rel == nil || rel.Stance != faction.StanceAllied {
		t.Fatalf("ardıl bağımsız müttefik olmalıydı: %+v", rel)
	}
}

func TestActiveSuccessorMetadataDirectlyAnnexesWithoutDecision(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Factions["successor"] = &faction.Faction{ID: "successor", NameTR: "Ardıl"}
	g.gs.Regions["enemy_cap"].SuccessorFactionID = "successor"

	collapse, prompted := g.captureBesiegedRegion(g.gs.Armies["atk"], g.gs.Regions["enemy_cap"], false)

	if prompted {
		t.Fatal("aktif ardıl devlet için vassal/serbest bırak paneli açılmamalıydı")
	}
	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "player" {
		t.Fatalf("aktif ardıl metadata'sında bölge doğrudan oyuncuya ilhak edilmeliydi, got=%s", got)
	}
	if !g.gs.Factions["enemy"].IsEliminated {
		t.Fatal("doğrudan ilhak eski sahibi elemeliydi")
	}
	if len(g.pendingConquestDecisions) != 0 || collapse.FactionID == "" {
		t.Fatalf("doğrudan ilhak karar kuyruğu/collapse üretmeliydi: pending=%d collapse=%+v", len(g.pendingConquestDecisions), collapse)
	}
}

func TestSuccessorDecisionVassalizesEliminatedSuccessor(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Factions["successor"] = &faction.Faction{ID: "successor", NameTR: "Ardıl", IsEliminated: true}
	g.gs.Regions["enemy_cap"].SuccessorFactionID = "successor"
	if !g.queueConquestDecision("player", g.gs.Regions["enemy_cap"], false) {
		t.Fatal("elenmiş ardıl için karar kuyruğa alınamadı")
	}

	g.resolvePendingSuccessorDecision(successorDecisionVassalize)

	if got := g.gs.Regions["enemy_cap"].OwnerID; got != "successor" {
		t.Fatalf("vassal kararında bölge ardıla geçmeliydi, got=%s", got)
	}
	if g.gs.Factions["successor"].IsEliminated || g.gs.Factions["successor"].OverlordID != "player" {
		t.Fatalf("ardıl yeniden kurulup oyuncuya bağlanmalıydı: %+v", g.gs.Factions["successor"])
	}
}

func TestAnnexVassalTransfersRealmAssets(t *testing.T) {
	g := conquestDecisionTestGame()
	g.gs.Turn = 20
	g.gs.Factions["enemy"].OverlordID = "player"
	g.gs.Factions["enemy"].VassalizedTurn = 1
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
