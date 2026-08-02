package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestBuildWarSummaryIncludesJoinedCoalitionsAndRefusals(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Sunni},
			"pv":     {ID: "pv", NameTR: "Oyuncu Vassalı", Religion: religion.Sunni, OverlordID: "player"},
			"ally":   {ID: "ally", NameTR: "Müttefik", Religion: religion.Sunni},
			"allyv":  {ID: "allyv", NameTR: "Müttefik Vassalı", Religion: religion.Sunni, OverlordID: "ally"},
			"refuse": {ID: "refuse", NameTR: "Gelmeyen", Religion: religion.Sunni},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Religion: religion.Catholic},
			"ev":     {ID: "ev", NameTR: "Düşman Vassalı", Religion: religion.Catholic, OverlordID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "player"},
			"r2": {ID: "r2", OwnerID: "pv"},
			"r3": {ID: "r3", OwnerID: "ally"},
			"r4": {ID: "r4", OwnerID: "allyv"},
			"r5": {ID: "r5", OwnerID: "refuse"},
			"r6": {ID: "r6", OwnerID: "enemy"},
			"r7": {ID: "r7", OwnerID: "ev"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"a2": {ID: "a2", OwnerID: "pv", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"a3": {ID: "a3", OwnerID: "ally", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"a4": {ID: "a4", OwnerID: "allyv", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"a5": {ID: "a5", OwnerID: "enemy", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"a6": {ID: "a6", OwnerID: "ev", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
	}
	g := &Game{gs: gs}
	result := diplomacy.WarDeclarationResult{
		Result: diplomacy.Result{Applied: true},
		PlayerCalls: []diplomacy.WarCallOutcome{
			{FactionID: "ally", NameTR: "Müttefik", Joined: true},
			{FactionID: "refuse", NameTR: "Gelmeyen", Joined: false},
		},
		EnemyCalls: []diplomacy.WarCallOutcome{},
	}

	report := g.buildWarSummary("enemy", result)

	if report.Attacker.TotalStrength != 40 {
		t.Fatalf("saldıran toplam güç 40 olmalıydı, got=%d", report.Attacker.TotalStrength)
	}
	if len(report.Attacker.Participants) != 4 {
		t.Fatalf("saldıran katılımcı sayısı 4 olmalıydı, got=%d", len(report.Attacker.Participants))
	}
	if len(report.Attacker.Refused) != 1 || report.Attacker.Refused[0] != "Gelmeyen" {
		t.Fatalf("gelmeyen müttefik summary'ye yazılmalıydı, got=%v", report.Attacker.Refused)
	}
	if report.Defender.TotalStrength != 20 {
		t.Fatalf("savunan toplam güç 20 olmalıydı, got=%d", report.Defender.TotalStrength)
	}
	if report.BalanceLabel == "" || report.PowerText == "" {
		t.Fatal("üst bilgi boş olmamalı")
	}
}

func TestWarDeclarationShowsSummaryBeforeQueuedMovementContinues(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Sunni},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"source": {ID: "source", OwnerID: "player", Neighbors: []world.RegionID{"target"}},
			"target": {ID: "target", OwnerID: "enemy", Neighbors: []world.RegionID{"source"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", RegionID: "source", MovePoints: 1, MaxMovePoints: 1},
		},
	}
	g := &Game{gs: gs, renderer: render.New(gs)}

	if !g.declareWar("enemy", nil) {
		t.Fatal("savaş ilanı uygulanmalıydı")
	}
	g.pendingWarFollowUp = &render.InputAction{Kind: render.ActionMoveArmy, ArmyID: "army", TargetRegion: "target"}

	if !g.renderer.WarSummaryVisible() || gs.Armies["army"].RegionID != "source" {
		t.Fatal("savaş özeti görünürken normal hareket başlamamalıydı")
	}

	g.renderer.HideWarSummary()
	g.resumeWarDeclarationFlow()
	if g.renderer.WarSummaryVisible() || g.warDeclarationContinuationPending || gs.Armies["army"].RegionID != "target" {
		t.Fatalf("özet kapandıktan sonra bekleyen akış devam etmeli: summary=%v pending=%v army=%+v", g.renderer.WarSummaryVisible(), g.warDeclarationContinuationPending, gs.Armies["army"])
	}
}

func TestWarDeclarationWithoutDefenderShowsSummaryBeforeSiegeDecision(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu", Religion: religion.Sunni},
			"enemy":  {ID: "enemy", NameTR: "Düşman", Religion: religion.Catholic},
		},
		Regions: map[world.RegionID]*world.Region{
			"source": {ID: "source", OwnerID: "player", Neighbors: []world.RegionID{"target"}},
			"target": {
				ID:          "target",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"source"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army": {
				ID:            "army",
				OwnerID:       "player",
				RegionID:      "source",
				MovePoints:    1,
				MaxMovePoints: 1,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	g := &Game{gs: gs, renderer: render.New(gs)}

	if !g.declareWar("enemy", nil) {
		t.Fatal("savaş ilanı uygulanmalıydı")
	}
	g.pendingWarFollowUp = &render.InputAction{Kind: render.ActionDeclareWar, ArmyID: "army", TargetRegion: "target"}
	if !g.renderer.WarSummaryVisible() || g.renderer.ConfirmDialogVisible() {
		t.Fatal("savunucu ordusu olmayan tahkimli hedefte önce yalnız Savaş Özeti görünmeli")
	}

	g.renderer.HideWarSummary()
	g.resumeWarDeclarationFlow()
	if g.renderer.WarSummaryVisible() || !g.renderer.ConfirmDialogVisible() {
		t.Fatal("özet kapandıktan sonra kuşatma kararı açılmalıydı")
	}
}
