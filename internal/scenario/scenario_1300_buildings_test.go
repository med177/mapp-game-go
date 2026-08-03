package scenario

import (
	"testing"

	"mapp-game-go/internal/world"
)

func buildingLevel(region *world.Region, buildingID string) int {
	if region == nil {
		return 0
	}
	level := 0
	for _, candidate := range region.Buildings {
		if candidate == buildingID {
			level++
		}
	}
	return level
}

func Test1300ScenarioSettlementInfrastructureHasMinimumBuildings(t *testing.T) {
	_, regions, factions := load1300IntegrityData(t)

	for regionID, region := range regions {
		if region == nil || region.IsSea {
			continue
		}

		hasPort := false
		hasFortress := false
		for _, settlement := range region.Settlements {
			hasPort = hasPort || settlement.Type == world.SettlementPort
			hasFortress = hasFortress || settlement.Type == world.SettlementFortress
		}
		if hasPort && buildingLevel(region, "port") < 1 {
			t.Errorf("liman yerleşimli bölgede en az 1. seviye liman olmalı: region=%s", regionID)
		}
		if hasFortress && buildingLevel(region, "walls") < 1 {
			t.Errorf("kale yerleşimli bölgede en az 1. seviye sur olmalı: region=%s", regionID)
		}
	}

	settlements := make(map[string]*world.Region)
	for _, region := range regions {
		if region == nil {
			continue
		}
		for i := range region.Settlements {
			settlement := &region.Settlements[i]
			settlements[settlement.ID] = region
		}
	}
	for factionID, faction := range factions {
		if faction == nil || faction.IsEliminated || faction.CapitalSettlementID == "" {
			continue
		}
		region := settlements[faction.CapitalSettlementID]
		if region == nil {
			t.Fatalf("başkent settlement bölgesi bulunamadı: faction=%s settlement=%s", factionID, faction.CapitalSettlementID)
		}
		for _, buildingID := range []string{"barracks", "granary", "temple", "market"} {
			if buildingLevel(region, buildingID) < 1 {
				t.Errorf("başkentte 1. seviye temel bina eksik: faction=%s region=%s building=%s", factionID, region.ID, buildingID)
			}
		}
	}
}

func Test1300ScenarioHistoricalStrongholdsHaveExpectedWallLevels(t *testing.T) {
	_, regions, _ := load1300IntegrityData(t)

	wantLevels := map[world.RegionID]int{
		"constantinople": 3,
		"austria":        3,
		"serbia":         3,
		"hungary":        3,
		"paris":          3,
		"egypt":          3,
		"rhodes":         3,
		"thrace":         2,
		"nis":            2,
		"vidin":          2,
		"bursa":          2,
		"nicomedia":      2,
		"sinop":          2,
		"bithynia":       2,
	}
	for regionID, want := range wantLevels {
		region := regions[regionID]
		if region == nil {
			t.Fatalf("tarihsel güçlü mevki bölgesi bulunamadı: %s", regionID)
		}
		if got := buildingLevel(region, "walls"); got < want {
			t.Errorf("tarihsel güçlü mevkide sur seviyesi düşük: region=%s got=%d want>=%d", regionID, got, want)
		}
	}
}
