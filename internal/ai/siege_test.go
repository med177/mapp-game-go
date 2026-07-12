package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func aiSiegeTestState(withSiegeUnit bool) *state.GameState {
	units := []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}
	if withSiegeUnit {
		units = append(units, army.Unit{TypeID: "siege", CurrentHP: 100})
	}
	return &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"src": {ID: "src", OwnerID: "ai_1", Neighbors: []world.RegionID{"fort"}},
			"fort": {
				ID:          "fort",
				OwnerID:     "player",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Hisar"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "src",
				Units:         units,
				MovePoints:    2,
				MaxMovePoints: 2,
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
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 14, Defense: 12, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}
}

func TestChooseBestMoveCanSelectFortifiedTargetWithoutSiegeUnit(t *testing.T) {
	gs := aiSiegeTestState(false)

	target := chooseBestMove(gs, gs.Armies["ai_army"])

	if target != "fort" {
		t.Fatalf("siege birimi olmayan AI tahkimli hedefi kuşatma adayı olarak seçebilmeliydi, got=%s", target)
	}
}

func TestExecuteMoveStartsSiegeOnFortifiedTarget(t *testing.T) {
	gs := aiSiegeTestState(false)
	a := gs.Armies["ai_army"]

	outcome := executeMove(gs, a, "fort", "ai_1")

	if !outcome.survived {
		t.Fatal("kuşatma başlatan AI ordusu hayatta kalmalıydı")
	}
	if gs.SiegeAt("fort") == nil {
		t.Fatal("AI tahkimli hedefte kuşatma başlatmalıydı")
	}
	if a.RegionID != "src" {
		t.Fatalf("kuşatma başlatan AI ordusu hedefe girmemeli, got=%s", a.RegionID)
	}
	if a.MovePoints != 0 {
		t.Fatalf("kuşatma sonrası hareket puanı bitmeliydi, got=%d", a.MovePoints)
	}
}

func TestChooseBestMoveAndExecuteMoveCanBreakSiegeWithoutConquestInAlliedRegion(t *testing.T) {
	gs := &state.GameState{
		Turn:            3,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "ai_1",
				Neighbors: []world.RegionID{"fort"},
			},
			"fort": {
				ID:          "fort",
				OwnerID:     "ally",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Terrain:     world.TerrainPlain,
				Settlements: []world.Settlement{{ID: "fortress", Type: world.SettlementFortress, NameTR: "Hisar"}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {
				ID:            "ai_army",
				OwnerID:       "ai_1",
				RegionID:      "src",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
			"besieger": {
				ID:            "besieger",
				OwnerID:       "player",
				RegionID:      "fort",
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
				MovePoints:    2,
				MaxMovePoints: 2,
			},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Catholic},
			"ai_1":   {ID: "ai_1", NameTR: "AI 1", Religion: religion.Sunni},
			"ally":   {ID: "ally", NameTR: "Müttefik", Religion: religion.Sunni},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai_1", "player"): {FactionA: "ai_1", FactionB: "player", Score: -70, Stance: faction.StanceWar},
			faction.RelationKey("ai_1", "ally"):   {FactionA: "ai_1", FactionB: "ally", Score: 60, Stance: faction.StanceAllied},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:          "fort",
				AttackerArmyID:    "besieger",
				AttackerFactionID: "player",
				StartedTurn:       3,
				FortLevel:         2,
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf":   {ID: "inf", Category: army.CategoryInfantry, Attack: 14, Defense: 12, Morale: 55},
			"siege": {ID: "siege", Category: army.CategorySiege, Tier: 1, Attack: 8, Defense: 4, Morale: 30},
		},
	}

	target := chooseBestMove(gs, gs.Armies["ai_army"])
	if target != "fort" {
		t.Fatalf("AI kuşatma altındaki müttefik bölgede düşman besiegerı hedefleyebilmeliydi, got=%s", target)
	}

	outcome := executeMove(gs, gs.Armies["ai_army"], target, "ai_1")

	if !outcome.survived {
		t.Fatal("kuşatma savunmasını kıran AI ordusu hayatta kalmalıydı")
	}
	if gs.Armies["ai_army"].RegionID != "fort" {
		t.Fatalf("AI orduyu hedefe sokmalıydı, got=%s", gs.Armies["ai_army"].RegionID)
	}
	if gs.Regions["fort"].OwnerID != "ally" {
		t.Fatalf("kuşatma kaldırılırken bölge sahipliği değişmemeliydi, got=%s", gs.Regions["fort"].OwnerID)
	}
	if gs.SiegeAt("fort") != nil {
		t.Fatal("besieging army yenildiğinde kuşatma kaydı temizlenmeliydi")
	}
}
