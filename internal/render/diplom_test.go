package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
)

func TestSortedFactionsSkipsPlayerAndEliminated(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"a":      {ID: "a"},
			"b":      {ID: "b", IsEliminated: true},
			"c":      {ID: "c"},
		},
	}

	got := sortedFactions(gs)
	if len(got) != 2 {
		t.Fatalf("beklenen 2 aktif fraksiyon, got=%d (%v)", len(got), got)
	}
	if got[0] != "a" || got[1] != "c" {
		t.Fatalf("beklenen [a c], got=%v", got)
	}
}

func TestHandleDiplomacyInputScrollsOnPanelWheel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
	}
	for i := 0; i < 12; i++ {
		id := faction.FactionID("f" + itoa(i))
		gs.Factions[id] = &faction.Faction{ID: id, NameTR: "Devlet " + itoa(i)}
	}

	r := &Renderer{
		gs:                   gs,
		showDiplomacy:        true,
		diplomacyScroll:      0,
		diplomacyFocus:       0,
		diplomacyActionFocus: 0,
	}

	layout := diplomacyListLayoutForScreen()
	input := gameui.InputState{
		MouseX: layout.panelRect.X + 20,
		MouseY: layout.panelRect.Y + 20,
		WheelY: -1,
	}

	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 1 {
		t.Fatalf("wheel aşağı kaydırınca scroll 1 olmalı, got=%d", r.diplomacyScroll)
	}

	input.WheelY = -1
	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 2 {
		t.Fatalf("ardışık wheel aşağı kaydırmada scroll 2 olmalı, got=%d", r.diplomacyScroll)
	}

	input.WheelY = 1
	r.handleDiplomacyInput(input)
	if r.diplomacyScroll != 1 {
		t.Fatalf("wheel yukarı kaydırınca scroll 1 olmalı, got=%d", r.diplomacyScroll)
	}
}
