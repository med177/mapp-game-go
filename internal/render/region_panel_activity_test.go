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

func TestRegionPanelActivityAndDiplomacyUseVisiblePanelGeometry(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("enemy_region")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID:   {ID: regionID, NameTR: "Düşman Bölgesi", OwnerID: "enemy", Neighbors: []world.RegionID{"neighbor"}},
			"neighbor": {ID: "neighbor", NameTR: "Komşu"},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar"},
		},
	}

	if !regionActivityNeighborVisible(gs, gs.Regions[regionID]) {
		t.Fatal("komşu listesi geliştirme modu kapalıyken de görünür olmalı")
	}
	if got := regionActivityNeighborContentHeight(gs, gs.Regions[regionID]); got <= 8+devNeighborTitleHeight {
		t.Fatalf("komşu satırı içerik yüksekliğine eklenmedi: %.1f", got)
	}

	start := buildingGridStartY(gs, gs.Regions[regionID], false)
	end := buildingGridEndY(gs, gs.Regions[regionID], start)
	barY := float64(end) + 5
	btn := buildRegionDiplomacyButtons(gs, "enemy", infoPanelX()+float32(panelPad), float32(barY), infoPanelW-float32(panelPad*2), regionPanelActionBarHeight)
	expectedY := barY + (float64(regionPanelActionBarHeight)-btn.H)/2
	if btn.Y < expectedY-0.01 || btn.Y > expectedY+0.01 {
		t.Fatalf("diplomasi düğmesi aksiyon bandında ortalanmadı: %+v", btn)
	}
	if !regionDiplomacyButtonHit(btn.X+btn.W/2, btn.Y+btn.H/2, gs, regionID) {
		t.Fatal("çizilen diplomasi düğmesinin merkezi tıklanabilir olmalı")
	}
}
