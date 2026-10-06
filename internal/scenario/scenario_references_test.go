package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/world"
)

func Test1300ScenarioRegionAndSettlementReferences(t *testing.T) {
	dataDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data")

	regions, err := world.LoadRegions(filepath.Join(dataDir, "regions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := world.LoadRegionSettlements(filepath.Join(dataDir, "settlements.json"), regions); err != nil {
		t.Fatal(err)
	}

	settlements := make(map[string]world.RegionID)
	for regionID, region := range regions {
		for _, settlement := range region.Settlements {
			if previousRegion, duplicate := settlements[settlement.ID]; duplicate {
				t.Fatalf("aynı settlement ID birden fazla bölgede tanımlı: id=%s regions=%s,%s", settlement.ID, previousRegion, regionID)
			}
			settlements[settlement.ID] = regionID
		}
	}

	for _, fileName := range []string{"events.json", "ai_strategies.json", "factions.json"} {
		path := filepath.Join(dataDir, fileName)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s okunamadı: %v", fileName, err)
		}

		var document any
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatalf("%s parse edilemedi: %v", fileName, err)
		}

		walkScenarioReferenceValues(t, fileName, document, "$", regions, settlements)
	}
}

func walkScenarioReferenceValues(
	t *testing.T,
	fileName string,
	value any,
	path string,
	regions map[world.RegionID]*world.Region,
	settlements map[string]world.RegionID,
) {
	t.Helper()

	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			childPath := path + "." + key
			switch key {
			case "region_id", "destination_region_id":
				checkScenarioRegionReference(t, fileName, childPath, child, regions)
			case "regions", "region_ids", "requires_owned_regions", "requires_owned_regions_any", "requires_unowned_regions", "source_region_ids", "target_regions", "readiness_regions":
				checkScenarioRegionList(t, fileName, childPath, child, regions)
			case "capital_settlement_id", "blocks_capital_settlement_id":
				checkScenarioSettlementReference(t, fileName, childPath, child, settlements)
			}
			walkScenarioReferenceValues(t, fileName, child, childPath, regions, settlements)
		}
	case []any:
		for index, child := range typed {
			walkScenarioReferenceValues(t, fileName, child, fmt.Sprintf("%s[%d]", path, index), regions, settlements)
		}
	}
}

func checkScenarioRegionReference(
	t *testing.T,
	fileName, path string,
	value any,
	regions map[world.RegionID]*world.Region,
) {
	t.Helper()

	regionID, ok := value.(string)
	if !ok {
		t.Errorf("%s %s string region ID değil: %T", fileName, path, value)
		return
	}
	if regions[world.RegionID(regionID)] == nil {
		t.Errorf("%s %s mevcut olmayan region ID kullanıyor: %q", fileName, path, regionID)
	}
}

func checkScenarioRegionList(
	t *testing.T,
	fileName, path string,
	value any,
	regions map[world.RegionID]*world.Region,
) {
	t.Helper()

	regionIDs, ok := value.([]any)
	if !ok {
		t.Errorf("%s %s string region ID listesi değil: %T", fileName, path, value)
		return
	}
	for index, regionID := range regionIDs {
		checkScenarioRegionReference(t, fileName, fmt.Sprintf("%s[%d]", path, index), regionID, regions)
	}
}

func checkScenarioSettlementReference(
	t *testing.T,
	fileName, path string,
	value any,
	settlements map[string]world.RegionID,
) {
	t.Helper()

	settlementID, ok := value.(string)
	if !ok {
		t.Errorf("%s %s string settlement ID değil: %T", fileName, path, value)
		return
	}
	if _, ok := settlements[settlementID]; !ok {
		t.Errorf("%s %s mevcut olmayan settlement ID kullanıyor: %q", fileName, path, settlementID)
	}
}
