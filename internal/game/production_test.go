package game

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestCompleteBuildingAddsNamedPortSettlementToMinorRegion(t *testing.T) {
	minor := &world.Region{
		ID:            "sinop_castle",
		IsMinorRegion: true,
		Neighbors:     []world.RegionID{"black_sea"},
		WorldX:        100,
		WorldY:        200,
	}
	sea := &world.Region{
		ID:     "black_sea",
		IsSea:  true,
		WorldX: 130,
		WorldY: 200,
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			minor.ID: minor,
			sea.ID:   sea,
		},
		BuildingTypes: map[string]*city.Building{
			"port": {ID: "port", MaxPerRegion: 1},
		},
	}

	g := &Game{gs: gs}
	if !g.completeBuilding(minor, "port") {
		t.Fatal("minor bölgedeki liman binası tamamlanmadı")
	}
	if len(minor.Settlements) != 1 {
		t.Fatalf("liman tamamlanınca %d yerleşim oluştu, 1 bekleniyordu", len(minor.Settlements))
	}
	port := minor.Settlements[0]
	if port.Type != world.SettlementPort || port.NameTR != "Liman" || port.Name != "Port" {
		t.Fatalf("liman yerleşimi yanlış: %#v", port)
	}
	if port.X != 110 || port.Y != 200 {
		t.Fatalf("liman kıyı konumu = (%d, %d), (%d, %d) bekleniyordu", port.X, port.Y, 110, 200)
	}
}

func TestScenarioExportNeighborsRestoresSourceAreaLinks(t *testing.T) {
	region := &world.Region{
		ID:                "parent",
		Neighbors:         []world.RegionID{"neighbor"},
		AreaNeighborOrder: []world.RegionID{"area::first", "area::second"},
	}

	got := scenarioExportNeighbors(region)
	want := []world.RegionID{"neighbor", "area::first", "area::second"}
	if len(got) != len(want) {
		t.Fatalf("neighbor sayısı = %d, %d bekleniyordu: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("neighbor[%d] = %q, %q bekleniyordu", i, got[i], want[i])
		}
	}
}
