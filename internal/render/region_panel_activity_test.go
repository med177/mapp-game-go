package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestRegionActivityPanelScrollAndBuildingOrder(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("test_region")
	gs := &state.GameState{
		DevelopmentMode: true,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, NameTR: "Test", OwnerID: "player", Neighbors: []world.RegionID{"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10"}},
		},
		BuildingTypes: map[string]*city.Building{
			"market":   {ID: "market", NameTR: "Pazar"},
			"farm":     {ID: "farm", NameTR: "Çiftlik"},
			"barracks": {ID: "barracks", NameTR: "Kışla"},
			"walls":    {ID: "walls", NameTR: "Surlar"},
			"temple":   {ID: "temple", NameTR: "İbadet Yeri"},
			"port":     {ID: "port", NameTR: "Liman"},
		},
	}
	for _, neighborID := range gs.Regions[regionID].Neighbors {
		gs.Regions[neighborID] = &world.Region{ID: neighborID, NameTR: string(neighborID)}
	}
	for i := 0; i < 3; i++ {
		gs.ActiveRegionEvents = append(gs.ActiveRegionEvents, state.RegionEventStatus{RegionID: regionID, EventID: "event", LabelTR: "Olay", TurnsLeft: 3})
	}

	start := buildingGridStartY(gs, gs.Regions[regionID], false)
	end := buildingGridEndY(gs, gs.Regions[regionID], start)
	viewport := regionPanelActivityViewport(gs, gs.Regions[regionID])
	if viewport.Y <= float64(end) {
		t.Fatalf("aktivite alani binalarin altinda olmali: activityY=%.1f buildingEndY=%.1f", viewport.Y, end)
	}
	maxScroll := clampRegionPanelScroll(gs, regionID, 1<<20)
	if maxScroll <= 0 {
		t.Fatal("uzun olay/komsu listesi icin pozitif scroll araligi bekleniyordu")
	}
	if !regionPanelActivityHit(viewport.X+4, viewport.Y+4, gs, regionID) {
		t.Fatal("aktivite viewport'u mouse wheel hit-test'i kabul etmeli")
	}
}
