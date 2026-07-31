package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
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
