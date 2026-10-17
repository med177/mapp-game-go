package render

import (
	"image/color"
	"math"
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSettlementLabelLayoutIncludesDevelopmentBar(t *testing.T) {
	if gotW, gotH := settlementLabelLayout(80, 16, true); gotW != 82 || gotH != 25 {
		t.Fatalf("development label layout = (%v, %v), want (82, 25)", gotW, gotH)
	}
	if gotW, gotH := settlementLabelLayout(80, 16, false); gotW != 80 || gotH != 16 {
		t.Fatalf("plain label layout = (%v, %v), want (80, 16)", gotW, gotH)
	}
}

func TestHasActiveBuildingUpgrade(t *testing.T) {
	region := &world.Region{ID: "region", OwnerID: "player"}
	gs := &state.GameState{
		PlayerFactionID: "player",
		ProductionQueue: []state.ProductionOrder{{
			Kind: "building", FactionID: "player", RegionID: "region", TurnsLeft: 1,
		}},
	}
	if !hasActiveBuildingUpgrade(gs, region) {
		t.Fatal("active player building order should show working icon")
	}

	tests := []struct {
		name  string
		order state.ProductionOrder
		owner string
	}{
		{name: "unit order", order: state.ProductionOrder{Kind: "unit", FactionID: "player", RegionID: "region", TurnsLeft: 1}, owner: "player"},
		{name: "another region", order: state.ProductionOrder{Kind: "building", FactionID: "player", RegionID: "other", TurnsLeft: 1}, owner: "player"},
		{name: "another faction", order: state.ProductionOrder{Kind: "building", FactionID: "other", RegionID: "region", TurnsLeft: 1}, owner: "player"},
		{name: "finished order", order: state.ProductionOrder{Kind: "building", FactionID: "player", RegionID: "region", TurnsLeft: 0}, owner: "player"},
		{name: "lost region", order: state.ProductionOrder{Kind: "building", FactionID: "player", RegionID: "region", TurnsLeft: 1}, owner: "other"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs.ProductionQueue[0] = test.order
			region.OwnerID = test.owner
			if hasActiveBuildingUpgrade(gs, region) {
				t.Fatal("unexpected active building order")
			}
		})
	}
}

func TestHasAffordableBuildingOption(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 100},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", GoldCost: 10, MaxPerRegion: 2},
		},
	}
	region := &world.Region{
		ID: "region", OwnerID: "player", Buildings: []string{"market"},
	}
	if !hasAffordableBuildingOption(gs, region) {
		t.Fatal("affordable building below its level cap should mark the bar red")
	}
	if got, want := settlementDevelopmentBackground(gs, region), (color.RGBA{190, 45, 38, 255}); got != want {
		t.Fatalf("upgrade background = %v, want %v", got, want)
	}

	unbuiltRegion := &world.Region{ID: "unbuilt", OwnerID: "player"}
	if !hasAffordableBuildingOption(gs, unbuiltRegion) {
		t.Fatal("affordable first construction should mark the bar red")
	}

	tests := []struct {
		name     string
		gold     int
		region   *world.Region
		queued   []state.ProductionOrder
		maxLevel int
	}{
		{name: "insufficient resources", gold: 0, region: region, maxLevel: 2},
		{name: "at maximum level", gold: 100, region: &world.Region{ID: "region", OwnerID: "player", Buildings: []string{"market", "market"}}, maxLevel: 2},
		{name: "already queued", gold: 100, region: region, queued: []state.ProductionOrder{{Kind: "building", RegionID: "region", TypeID: "market"}}, maxLevel: 2},
		{name: "locked region", gold: 100, region: &world.Region{ID: "region", OwnerID: "player", IsLocked: true, Buildings: []string{"market"}}, maxLevel: 2},
		{name: "unbuilt but unaffordable", gold: 0, region: unbuiltRegion, maxLevel: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gs.Factions["player"].Gold = test.gold
			gs.BuildingTypes["market"].MaxPerRegion = test.maxLevel
			gs.ProductionQueue = test.queued
			if hasAffordableBuildingOption(gs, test.region) {
				t.Fatal("unexpected buildable or upgradeable building")
			}
			if got, want := settlementDevelopmentBackground(gs, test.region), (color.RGBA{0, 0, 0, 255}); got != want {
				t.Fatalf("background = %v, want %v", got, want)
			}
		})
	}
}

func TestSettlementBuildingDevelopment(t *testing.T) {
	gs := &state.GameState{
		BuildingTypes: map[string]*city.Building{
			"market":  {ID: "market", MaxPerRegion: 2},
			"granary": {ID: "granary", MaxPerRegion: 3},
		},
	}
	region := &world.Region{
		Buildings: []string{"market", "market", "market", "granary", "unknown"},
	}

	if got, want := settlementBuildingDevelopment(gs, region), 3.0/5.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("settlementBuildingDevelopment() = %v, want %v", got, want)
	}
}

func TestSettlementBuildingDevelopmentWithoutAvailableBuildings(t *testing.T) {
	gs := &state.GameState{BuildingTypes: map[string]*city.Building{}}
	region := &world.Region{Buildings: []string{"unknown"}}

	if got := settlementBuildingDevelopment(gs, region); got != 0 {
		t.Fatalf("settlementBuildingDevelopment() = %v, want 0", got)
	}
}

func TestSettlementBuildingDevelopmentExcludesUnavailableBuildings(t *testing.T) {
	gs := &state.GameState{
		BuildingTypes: map[string]*city.Building{
			"market":  {ID: "market", MaxPerRegion: 2},
			"granary": {ID: "granary", MaxPerRegion: 3, MinorRegions: true},
		},
	}
	region := &world.Region{
		IsMinorRegion: true,
		Buildings:     []string{"market", "granary"},
	}

	if got, want := settlementBuildingDevelopment(gs, region), 1.0/3.0; math.Abs(got-want) > 1e-9 {
		t.Fatalf("settlementBuildingDevelopment() = %v, want %v", got, want)
	}
}
