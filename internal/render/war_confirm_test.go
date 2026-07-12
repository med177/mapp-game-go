package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestHandleDiplomacyInputOpensWarConfirmModal(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StancePeace, Score: -40},
		},
	}
	r := &Renderer{
		gs:                     gs,
		showDiplomacy:          true,
		diplomacyTargetFaction: "enemy",
		diplomacyActionFocus:   0,
		prevKeys:               map[ebiten.Key]bool{},
		prevMouse:              map[ebiten.MouseButton]bool{},
	}

	sendBtn := buildDiplomacySendButton()
	act := r.handleDiplomacyInput(gameui.InputState{
		MouseX:          sendBtn.X + 5,
		MouseY:          sendBtn.Y + 5,
		LeftJustPressed: true,
	})

	if act.Kind != ActionNone {
		t.Fatalf("savaş ilanı önce modal açmalı, doğrudan aksiyon üretmemeli: %+v", act)
	}
	if !r.warConfirm.show {
		t.Fatal("war confirm modalı açılmalıydı")
	}
	if r.warConfirm.factionID != "enemy" {
		t.Fatalf("yanlış hedef modalı açıldı: %q", r.warConfirm.factionID)
	}
	if r.showDiplomacy {
		t.Fatal("modal açılınca diplomasi paneli kapanmalıydı")
	}
}

func TestFinalizeWarConfirmCarriesSelectedAllies(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}
	r := &Renderer{gs: gs}
	wc := warConfirmState{
		factionID: "enemy",
		preview: diplomacy.WarDeclarationPreview{
			Attacker: diplomacy.WarSidePreview{
				CallableAllies: []diplomacy.WarParticipantPreview{
					{FactionID: "ally_a"},
					{FactionID: "ally_b"},
				},
			},
		},
		selectedAllies: map[faction.FactionID]bool{
			"ally_b": true,
		},
		battleContext: combat.BattleContextLand,
	}

	act := r.finalizeWarConfirm(wc)

	if act.Kind != ActionDeclareWar {
		t.Fatalf("pending hareket yokken yalnız savaş ilanı dönmeliydi, got=%s", act.Kind)
	}
	if len(act.WarAllies) != 1 || act.WarAllies[0] != "ally_b" {
		t.Fatalf("seçili müttefik listesi aksiyona taşınmalıydı, got=%v", act.WarAllies)
	}
}

func TestFinalizeWarConfirmOpensSiegeDecisionWithoutSiegeUnit(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			"src": {
				ID:        "src",
				OwnerID:   "player",
				Neighbors: []world.RegionID{"dst"},
			},
			"dst": {
				ID:          "dst",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"src"},
				Buildings:   []string{"walls"},
				Settlements: []world.Settlement{{ID: "fort", Type: world.SettlementFortress, NameTR: "Kale"}},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"atk": {
				ID:            "atk",
				OwnerID:       "player",
				RegionID:      "src",
				MovePoints:    2,
				MaxMovePoints: 2,
				Units:         []army.Unit{{TypeID: "inf", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 55},
		},
	}
	r := &Renderer{gs: gs}
	wc := warConfirmState{
		factionID:   "enemy",
		pendingArmy: "atk",
		pendingDest: "dst",
		preview: diplomacy.WarDeclarationPreview{
			Attacker: diplomacy.WarSidePreview{},
		},
		battleContext: combat.BattleContextLand,
	}

	act := r.finalizeWarConfirm(wc)

	if act.Kind != ActionDeclareWar {
		t.Fatalf("savaş ilanı geri dönmeliydi, got=%s", act.Kind)
	}
	if !r.confirmDialog.show {
		t.Fatal("kuşatma kararı modalı açılmalıydı")
	}
	if r.confirmDialog.thirdLabel != "" {
		t.Fatalf("kuşatma birimi yokken genel hücum düğmesi gizlenmeliydi, got=%q", r.confirmDialog.thirdLabel)
	}
	if r.confirmDialog.pendingAction.Kind != ActionStartSiege {
		t.Fatalf("kuşatma kararı start siege üretmeliydi, got=%s", r.confirmDialog.pendingAction.Kind)
	}
}

func TestWarConfirmScrollHelpersClampAndHitVisibleRows(t *testing.T) {
	viewport := gameui.Rect{X: 100, Y: 200, W: 320, H: 110}
	entries := []diplomacy.WarParticipantPreview{
		{FactionID: "a"},
		{FactionID: "b"},
		{FactionID: "c"},
		{FactionID: "d"},
		{FactionID: "e"},
	}
	selected := map[faction.FactionID]bool{"c": true}

	if rows := warConfirmVisibleRows(viewport); rows != 2 {
		t.Fatalf("110px viewport için 2 görünür satır bekleniyordu, got=%d", rows)
	}
	if got := clampWarConfirmScroll(len(entries), viewport, 99); got != 3 {
		t.Fatalf("scroll max'a clamplenmeli, got=%d", got)
	}

	boxes := warConfirmCheckboxes(viewport, entries, selected, 2)
	if len(boxes) != 2 {
		t.Fatalf("scroll sonrası yalnız görünür checkbox'lar dönmeli, got=%d", len(boxes))
	}
	if !boxes[0].Checked {
		t.Fatal("scroll ile görünür hale gelen seçili entry checked olmalıydı")
	}
	if idx := warConfirmEntryIndexAt(viewport, entries, 2, boxes[0].Rect.X+2, boxes[0].Rect.Y+2); idx != 2 {
		t.Fatalf("ilk görünür checkbox üçüncü entry'ye maplenmeliydi, got=%d", idx)
	}
}
