package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSettlementPanelHitRequiresVisiblePanel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"morea": {
					ID: "morea",
					Settlements: []world.Settlement{
						{Name: "Mora"},
					},
				},
			},
		},
		SelectedRegion: "morea",
	}

	mx := float64(settlementPanelX() + 20)
	my := float64(settlementPanelY() + 20)

	if r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken hit-test map alani tuketmemeli")
	}
	if r.settlementPanelCloseHit(mx, my) {
		t.Fatal("yerlesim paneli kapaliyken kapatma butonu aktif olmamali")
	}

	r.selectSettlement("morea", 0)

	if !r.settlementPanelHit(mx, my) {
		t.Fatal("yerlesim paneli acikken hit-test aktif olmali")
	}
}

func TestBuildingGridHitTestUsesDrawnSpriteBoundsOnly(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	oldCachedRegion := lastBuildingGridRegionID
	oldCachedCards := lastBuildingGridCards
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
		lastBuildingGridRegionID = oldCachedRegion
		lastBuildingGridCards = oldCachedCards
	}()

	regionID := world.RegionID("bitinya")
	region := &world.Region{
		ID:        regionID,
		NameTR:    "Bitinya",
		OwnerID:   "osm",
		Terrain:   world.TerrainPlain,
		Buildings: []string{"market"},
	}
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Regions: map[world.RegionID]*world.Region{
			regionID: region,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 1},
		},
	}

	startY := buildingGridStartY(gs, region, false)
	cards := buildBuildingCardComponents(gs, region, infoPanelX(), startY, infoPanelW)
	if len(cards) != 1 {
		t.Fatalf("1 bina karti bekleniyordu, got=%d", len(cards))
	}
	card := cards[0]

	if got := BuildingGridHitTest(card.ImageRect.X+card.ImageRect.W/2, card.ImageRect.Y+card.ImageRect.H/2, gs, regionID, false); got != "market" {
		t.Fatalf("resim merkezinde market bekleniyordu, got=%q", got)
	}
	if got := BuildingGridHitTest(card.Rect.X+card.Rect.W/2, card.LabelY+1, gs, regionID, false); got != "" {
		t.Fatalf("label alani hit olmamali, got=%q", got)
	}
	if got := BuildingGridHitTest(card.Rect.X+card.Rect.W/2, card.Rect.Y+card.Rect.H+3, gs, regionID, false); got != "" {
		t.Fatalf("kart disi bosluk hit olmamali, got=%q", got)
	}

	card.SpriteRect = card.ImageRect
	card.SpriteRect.Y += 20
	card.SpriteRect.H -= 20
	cacheBuildingGridComponents(regionID, []buildingCardComponent{card})
	if got := BuildingGridHitTest(card.ImageRect.X+card.ImageRect.W/2, card.ImageRect.Y+4, gs, regionID, false); got != "" {
		t.Fatalf("son cizilen sprite disindaki ust alan hit olmamali, got=%q", got)
	}
	if got := BuildingGridHitTest(card.SpriteRect.X+card.SpriteRect.W/2, card.SpriteRect.Y+card.SpriteRect.H/2, gs, regionID, false); got != "market" {
		t.Fatalf("son cizilen sprite merkezinde market bekleniyordu, got=%q", got)
	}
}
