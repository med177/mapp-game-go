package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestArmyPowerFooterShowsPointerOnPowerText(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 900

	const armyID army.ArmyID = "player-army"
	gs := &state.GameState{
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", Terrain: world.TerrainPlain},
		},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Attack: 10, Defense: 8, Morale: 5},
		},
		Armies: map[army.ArmyID]*army.Army{
			armyID: {ID: armyID, OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "infantry", CurrentHP: army.MaxUnitHP}}},
		},
	}
	r := &Renderer{gs: gs, showArmyDetailPanel: true, SelectedArmy: armyID}
	r.rebuildUILayers()

	rect, ok := armyPanelPowerFooterRect(gs, armyID, false)
	if !ok {
		t.Fatal("ordu gücü metni için hover alanı oluşturulmadı")
	}
	x, y := rect.X+rect.W/2, rect.Y+rect.H/2
	if !ArmyPanelInteractiveHit(x, y, gs, armyID) {
		t.Fatal("ordu gücü metninin merkezi hover hit-test'ine girmedi")
	}
	if got := r.cursorShapeAt(x, y); got != ebiten.CursorShapePointer {
		t.Fatalf("ordu gücü metni cursor şekli = %v, want pointer", got)
	}

	outsideX := rect.X - 2
	if got := r.cursorShapeAt(outsideX, y); got == ebiten.CursorShapePointer {
		t.Fatal("ordu gücü metninin dışı pointer cursor üretti")
	}
}
