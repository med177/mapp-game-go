package render

import (
	"math"
	"strings"
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

func TestArmyTransportFooterShowsPointerAndTotalSupplyCapacity(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 900

	const armyID army.ArmyID = "player-fleet"
	units := make([]army.Unit, 16)
	for i := range units {
		units[i] = army.Unit{TypeID: "transport"}
	}
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
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 4},
			"dragon":    {ID: "dragon", Category: army.CategoryDragon, GrainUpkeep: 25},
		},
		Armies: map[army.ArmyID]*army.Army{
			armyID: {ID: armyID, OwnerID: "player", RegionID: "home", IsNaval: true, Units: units},
		},
	}
	r := &Renderer{gs: gs, showArmyDetailPanel: true, SelectedArmy: armyID}
	r.rebuildUILayers()

	rect, ok := armyTransportInfoButtonRect(gs, armyID)
	if !ok {
		t.Fatal("filo taşıma footer'ı için hover alanı oluşturulmadı")
	}
	x, y := rect.X+rect.W/2, rect.Y+rect.H/2
	if got := r.cursorShapeAt(x, y); got != ebiten.CursorShapePointer {
		t.Fatalf("filo taşıma footer cursor şekli = %v, want pointer", got)
	}
	footer := armyPanelFooterLayoutFor(gs, gs.Armies[armyID], armyPanelGeometry())
	button, ok := armyTransportInfoButtonFromFooter(footer)
	if !ok {
		t.Fatal("filo taşıma footer düğmesi oluşturulmadı")
	}
	gap := float64(footer.powerX) - (button.X + button.W)
	if math.Abs(gap-8) > 0.01 {
		t.Fatalf("taşıma düğmesiyle güç metni arasındaki boşluk = %.2f px, want 8 px", gap)
	}

	title, detail, ok := armyTransportInfoTooltipText(gs, gs.Armies[armyID])
	if !ok || title != "Filo Taşıma Kapasitesi" {
		t.Fatalf("filo taşıma tooltip sonucu = (%q, %q, %v)", title, detail, ok)
	}
	if want := "Toplam ikmal malı kapasitesi (5 tur): 2500 mal"; !strings.Contains(detail, want) {
		t.Fatalf("tooltip detayı %q, %q içermiyor", detail, want)
	}
	lines := strings.Split(detail, "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[1], "Mevcut yük:") {
		t.Fatalf("mevcut yük ikinci satırda değil: %q", detail)
	}
}
