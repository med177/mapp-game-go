package save

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestCompactArmyRoundTripPreservesNavalMission(t *testing.T) {
	original := map[army.ArmyID]*army.Army{
		"fleet": {
			ID: "fleet", OwnerID: "player", IsNaval: true,
			Units: []army.Unit{{TypeID: "warship"}},
			NavalMission: &army.NavalMission{
				Kind: army.NavalMissionBlockade, TargetRegionID: world.RegionID("sea_aegean"),
			},
		},
	}
	restored := restoreArmiesFromSaveState(convertArmiesToSaveState(original))
	mission := restored["fleet"].NavalMission
	if mission == nil || mission.Kind != army.NavalMissionBlockade || mission.TargetRegionID != "sea_aegean" {
		t.Fatalf("donanma görevi compact round-trip sonrası korunmadı: %+v", mission)
	}
}
