package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestLandRegionAttritionPercentUsesTerrainAreaValue(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"parent": {ID: "parent"},
			"desert": {ID: "desert", IsTerrainArea: true, ParentRegionID: "parent", TerrainAreaID: "d1"},
		},
		TerrainAreas: []world.TerrainArea{
			{ID: "d1", ParentRegionID: "parent", AttritionCost: 15},
		},
	}
	if got := gs.LandRegionAttritionPercent(gs.Regions["desert"]); got != 15 {
		t.Fatalf("çöl alanı yıpranma yüzdesi hatalı: got=%d", got)
	}
	if got := gs.LandRegionAttritionPercent(gs.Regions["parent"]); got != 0 {
		t.Fatalf("normal bölge yıpranma almamalı: got=%d", got)
	}
}

func TestApplyLandRegionEntryAttritionDamagesLandArmyOnly(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"desert": {ID: "desert", IsTerrainArea: true, ParentRegionID: "parent", TerrainAreaID: "d1"},
			"sea":    {ID: "sea", IsSea: true},
		},
		TerrainAreas: []world.TerrainArea{
			{ID: "d1", ParentRegionID: "parent", AttritionCost: 20},
		},
	}
	landArmy := &army.Army{RegionID: "desert", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}}
	if lost := gs.ApplyLandRegionEntryAttrition(landArmy); lost != 0 || landArmy.Units[0].CurrentHP != 80 {
		t.Fatalf("kara ordusu çöl yıpranmasını almalıydı: lost=%d hp=%d", lost, landArmy.Units[0].CurrentHP)
	}

	navalArmy := &army.Army{RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}}
	if lost := gs.ApplyLandRegionEntryAttrition(navalArmy); lost != 0 || navalArmy.Units[0].CurrentHP != 100 {
		t.Fatalf("deniz ordusu arazi yıpranması almamalı: lost=%d hp=%d", lost, navalArmy.Units[0].CurrentHP)
	}
}
