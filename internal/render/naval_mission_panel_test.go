package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestNavalMissionOptionsShowFishingOnlyBesideOwnedCoast(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 30},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: "player", Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "sea",
				Units: []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
			},
		},
	}

	hasFishingOption := func() bool {
		for _, option := range navalMissionOptions(gs, gs.Armies["fleet"]) {
			if option.kind == army.NavalMissionFishing {
				return true
			}
		}
		return false
	}

	if !hasFishingOption() {
		t.Fatal("fishing option missing for non-warship with configured yield beside player's coast")
	}
	if !playerNavalMissionEligible(gs, gs.Armies["fleet"]) {
		t.Fatal("fishing-capable non-warship fleet should have access to the mission panel")
	}
	gs.Regions["coast"].OwnerID = "other"
	if hasFishingOption() {
		t.Fatal("fishing option should be hidden without player's adjacent coast")
	}
}

func TestFishingBonusBadgeSuppressesArmyMarkerSummaryAtBadgeOnly(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"fishing_ship": {ID: "fishing_ship", Category: army.CategoryNavalTrade, FishingGrainIncome: 30},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", OwnerID: "player", Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"coast"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID: "fleet", OwnerID: "player", IsNaval: true, RegionID: "sea",
				NavalMission: &army.NavalMission{Kind: army.NavalMissionFishing, TargetRegionID: "sea"},
				Units:        []army.Unit{{TypeID: "fishing_ship", CurrentHP: army.MaxUnitHP}},
			},
		},
	}
	r := &Renderer{
		gs:       gs,
		camScale: 1,
		armyIconWorldBuf: []armyIconPos{{
			ArmyID: "fleet",
			X:      0,
			Y:      0,
		}},
	}
	r.armyIconLayoutKey = r.armyIconStateKey()
	r.armyIconLayoutValid = true

	badge := navalMissionBonusBadgeRect(float32(ScreenWidth/2), float32(ScreenHeight/2))
	if !r.armyMarkerSummarySuppressedAt(badge.X+badge.W/2, badge.Y+badge.H/2) {
		t.Fatal("fishing income badge should suppress army marker summary tooltip")
	}
	if r.armyMarkerSummarySuppressedAt(ScreenWidth/2, ScreenHeight/2) {
		t.Fatal("army marker summary should remain available outside the bonus badge")
	}
}
