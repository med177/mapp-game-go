package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
)

func navalMissionPanelStateFixture() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		Armies: map[army.ArmyID]*army.Army{
			"war":       {ID: "war", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
			"transport": {ID: "transport", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "transport"}}},
			"enemy":     {ID: "enemy", OwnerID: "enemy", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
		},
	}
}

func TestNavalMissionButtonOnlyTargetsEligiblePlayerFleet(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := navalMissionPanelStateFixture()
	button := navalMissionButtonRect(armyPanelGeometry())
	x, y := button.X+button.W/2, button.Y+button.H/2
	if !navalMissionButtonHit(x, y, gs, "war") || !navalMissionButtonHit(x, y, gs, "transport") {
		t.Fatal("oyuncunun savaş/nakliye filosunda görev butonu hit olmalıydı")
	}
	if navalMissionButtonHit(x, y, gs, "enemy") {
		t.Fatal("düşman filosunda görev butonu hit olmamalıydı")
	}
}

func TestNavalMissionOptionsExposeWarshipAndTransportTasks(t *testing.T) {
	gs := navalMissionPanelStateFixture()
	options := navalMissionOptions(gs, gs.Armies["war"])
	if len(options) != 3 {
		t.Fatalf("savaş filosu için devriye, abluka ve escort bekleniyordu, %d seçenek var", len(options))
	}
	if options[0].kind != army.NavalMissionPatrol || options[1].kind != army.NavalMissionBlockade || options[2].kind != army.NavalMissionEscort {
		t.Fatalf("savaş filosu görevleri beklenen sırada değil: %+v", options)
	}
	gs.Armies["transport"].EmbarkedUnits = []army.Unit{{TypeID: "infantry"}}
	transportOptions := navalMissionOptions(gs, gs.Armies["transport"])
	if len(transportOptions) != 1 || transportOptions[0].kind != army.NavalMissionTransport {
		t.Fatalf("taşıma filosunda yalnız nakliye görevi bekleniyordu: %+v", transportOptions)
	}
}

func TestNavalMissionPanelStaysInsideViewport(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720

	layout := navalMissionPanelLayoutFor(12)
	if layout.panelX < 0 || layout.panelY < 0 || layout.panelX+layout.panelW > float32(ScreenWidth) || layout.panelY+layout.panelH > float32(ScreenHeight) {
		t.Fatalf("donanma görev paneli viewport dışına taştı: %+v", layout)
	}
}

func TestNavalMissionPanelTwoLineRowsHaveVerticalClearance(t *testing.T) {
	if navalMissionPanelRowH < 60 {
		t.Fatalf("iki satırlı görev seçenekleri için satır yüksekliği yetersiz: %.1f", navalMissionPanelRowH)
	}

	rowRectHeight := navalMissionPanelRowH - 10
	if rowRectHeight < 50 {
		t.Fatalf("görev satırının iç kutusu iki satırlı metne sığmıyor: %.1f", rowRectHeight)
	}
}

func TestNavalMissionPanelUsesSharedCloseIconButton(t *testing.T) {
	layout := navalMissionPanelLayoutFor(1)
	if layout.close.Label != "" || layout.close.Icon != gameui.IconClose {
		t.Fatalf("donanma görev paneli ortak kapatma ikonunu kullanmalı: %+v", layout.close)
	}
	if layout.close.IconSize != 13 {
		t.Fatalf("donanma görev kapatma ikonu diğer panellerdeki boyutu kullanmalı: %.1f", layout.close.IconSize)
	}
}

func TestNavalMissionPanelRowGeometryMatchesExpandedHeight(t *testing.T) {
	layout := navalMissionPanelLayoutFor(2)
	first := navalMissionPanelRowRect(layout, 0)
	second := navalMissionPanelRowRect(layout, 1)
	if first.H != float64(navalMissionPanelRowH-10) {
		t.Fatalf("ilk görev satırı yüksekliği ortak row geometry ile eşleşmiyor: %+v", first)
	}
	if second.Y-first.Y != float64(navalMissionPanelRowH) {
		t.Fatalf("görev satırları genişletilmiş row yüksekliğini kullanmıyor: first=%+v second=%+v", first, second)
	}
}
