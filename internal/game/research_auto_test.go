package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestAutoStartResearchIfIdleStartsNextResearchableTech(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 25,
				Research: faction.ResearchState{
					Completed: map[string]bool{"root": true},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {
				ID:       "root",
				Category: tech.CategoryMilitary,
			},
			"smithing": {
				ID:            "smithing",
				NameTR:        "Demircilik",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if started := g.autoStartResearchIfIdle(); !started {
		t.Fatal("uygun teknoloji varken otomatik baslatma calismaliydi")
	}

	player := gs.Factions["player"]
	if player.Research.ActiveID != "smithing" {
		t.Fatalf("beklenen aktif research smithing olmaliydi, got=%q", player.Research.ActiveID)
	}
	if player.Research.TurnsLeft != 4 {
		t.Fatalf("turn sayaci teknoloji turu kadar ayarlanmaliydi, got=%d", player.Research.TurnsLeft)
	}
	if player.Gold != 5 {
		t.Fatalf("gold otomatik baslatmada dusmeliydi, got=%d", player.Gold)
	}
}

func TestAutoStartResearchIfIdleIgnoresPausedTechsWhenAnotherTechIsAvailable(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID:   "player",
				Gold: 25,
				Research: faction.ResearchState{
					Completed:   map[string]bool{"root": true},
					PausedTurns: map[string]int{"paused": 3},
				},
			},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {
				ID:       "root",
				Category: tech.CategoryMilitary,
			},
			"paused": {
				ID:            "paused",
				NameTR:        "Duraklatilmis",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
			"smithing": {
				ID:            "smithing",
				NameTR:        "Demircilik",
				Category:      tech.CategoryMilitary,
				Requires:      []string{"root"},
				GoldCost:      20,
				TurnsRequired: 4,
			},
		},
	}
	g := &Game{gs: gs, renderer: &render.Renderer{}}

	if started := g.autoStartResearchIfIdle(); !started {
		t.Fatal("duraklatilmis tech olsa da uygun baska tech varken otomatik baslatma calismaliydi")
	}

	player := gs.Factions["player"]
	if player.Research.ActiveID != "smithing" {
		t.Fatalf("paused tech otomatik secimi kilitlememeliydi, got=%q", player.Research.ActiveID)
	}
	if player.Gold != 5 {
		t.Fatalf("gold otomatik baslatmada dusmeliydi, got=%d", player.Gold)
	}
	if player.Research.PausedTurns["paused"] != 3 {
		t.Fatalf("duraklatilmis tech kaydi korunmaliydi, got=%d", player.Research.PausedTurns["paused"])
	}
}

func TestCheckRegionUnlocksDoesNotUnlockTerrainAreas(t *testing.T) {
	source := &world.Region{
		ID:        "florence",
		Neighbors: []world.RegionID{"area::apennin", "hidden_region"},
	}
	terrain := &world.Region{
		ID:            "area::apennin",
		IsTerrainArea: true,
		TerrainAreaID: "apennin",
		IsLocked:      true,
	}
	normal := &world.Region{ID: "hidden_region", IsLocked: true}
	armyRef := &army.Army{ID: "army", OwnerID: "florence_rep", RegionID: source.ID}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			source.ID:  source,
			terrain.ID: terrain,
			normal.ID:  normal,
		},
		Armies: map[army.ArmyID]*army.Army{armyRef.ID: armyRef},
	}

	unlocked := checkRegionUnlocks(gs)
	if !terrain.IsLocked {
		t.Fatal("terrain area was unlocked by a neighboring army")
	}
	if len(unlocked) != 1 || unlocked[0] != normal.ID {
		t.Fatalf("unexpected unlocked regions: %#v", unlocked)
	}
}

func TestResolveSiegesForcesSurrenderWhenDurationExpiresWithDefender(t *testing.T) {
	const (
		attackerID = faction.FactionID("venice")
		defenderID = faction.FactionID("arnavut_des")
	)
	target := &world.Region{
		ID:        "albania",
		NameTR:    "Arnavutluk",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{"arnavut_home"},
		Buildings: []string{"walls", "walls", "granary"},
	}
	home := &world.Region{
		ID:        "arnavut_home",
		OwnerID:   string(defenderID),
		Neighbors: []world.RegionID{target.ID},
		WorldX:    -100,
	}
	attacker := &army.Army{
		ID:       "venice_army",
		OwnerID:  string(attackerID),
		RegionID: target.ID,
		Units:    []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	defender := &army.Army{
		ID:        "albania_garrison",
		OwnerID:   string(defenderID),
		RegionID:  target.ID,
		Commander: &army.Commander{ID: "albania_commander", OwnerID: string(defenderID)},
		Units:     []army.Unit{{TypeID: "militia", CurrentHP: army.MaxUnitHP}},
	}
	siege := &state.SiegeState{
		RegionID:          target.ID,
		AttackerArmyID:    attacker.ID,
		DefenderArmyID:    defender.ID,
		AttackerFactionID: string(attackerID),
		TurnsElapsed:      4,
		FortLevel:         2,
		GranaryLevel:      1,
	}
	gs := &state.GameState{
		PlayerFactionID: attackerID,
		Regions:         map[world.RegionID]*world.Region{target.ID: target, home.ID: home},
		Armies:          map[army.ArmyID]*army.Army{attacker.ID: attacker, defender.ID: defender},
		Factions: map[faction.FactionID]*faction.Faction{
			attackerID: {ID: attackerID},
			defenderID: {ID: defenderID},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(attackerID, defenderID): {
				FactionA: attackerID,
				FactionB: defenderID,
				Stance:   faction.StanceWar,
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{target.ID: siege},
	}

	if got := siege.TurnsUntilSurrender(); got != 1 {
		t.Fatalf("son çözümleme turu öncesi 1 tur kalmalıydı, got=%d", got)
	}

	(&Game{gs: gs}).resolveSieges()

	if _, ok := gs.Sieges[target.ID]; ok {
		t.Fatal("zorunlu teslimiyet sonrasında kuşatma kaydı kalmamalıydı")
	}
	if target.OwnerID != string(attackerID) {
		t.Fatalf("süre dolunca bölge kuşatmacıya geçmeliydi, owner=%q", target.OwnerID)
	}
	transferred := gs.Armies[defender.ID]
	if transferred == nil {
		t.Fatal("teslim olan savunma ordusu kazananın kuvvetlerine devredilmeliydi")
	}
	if transferred.OwnerID != string(attackerID) {
		t.Fatalf("savunma ordusunun sahibi kuşatan olmalıydı, owner=%q", transferred.OwnerID)
	}
	if transferred.RegionID != target.ID {
		t.Fatalf("savunma ordusu teslim çözümlemesinde otomatik kaçmamalıydı, region=%q", transferred.RegionID)
	}
	if transferred.Commander == nil || transferred.Commander.OwnerID != string(attackerID) {
		t.Fatal("devredilen ordunun komutan sahipliği de güncellenmeliydi")
	}
}
