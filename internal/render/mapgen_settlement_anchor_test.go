package render

import (
	"os"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSloveniaKoperSettlementAnchorStaysInsideShape(t *testing.T) {
	scenarios := []string{
		filepath.Join("..", "..", "assets", "scenarios", "1444_ottoman_empire"),
		filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"),
	}

	for _, scenarioPath := range scenarios {
		t.Run(filepath.Base(scenarioPath), func(t *testing.T) {
			skipIfScenarioDirMissing(t, scenarioPath)

			regions, _, err := world.LoadRegionsWithOrder(filepath.Join(scenarioPath, "data", "regions.json"))
			if err != nil {
				t.Fatalf("regions yuklenemedi: %v", err)
			}
			if err := world.LoadRegionSettlements(filepath.Join(scenarioPath, "data", "settlements.json"), regions); err != nil {
				t.Fatalf("settlements yuklenemedi: %v", err)
			}
			shapeData, err := world.LoadCountryShapes(filepath.Join(scenarioPath, "data", "country_shapes.json"), regions)
			if err != nil {
				t.Fatalf("shapes yuklenemedi: %v", err)
			}

			gs := &state.GameState{
				Regions:   regions,
				ShapeData: shapeData,
			}

			wm := NewWorldMap(gs)
			region := regions["slovenia"]
			if region == nil {
				t.Fatal("slovenia bolgesi bulunamadi")
			}
			var koperIdx = -1
			for i, settlement := range region.Settlements {
				if settlement.ID == "slovenia_koper" {
					koperIdx = i
					break
				}
			}
			if koperIdx < 0 {
				t.Fatal("slovenia_koper settlement bulunamadi")
			}

			raw := region.Settlements[koperIdx]
			rawInsideShape := regionContainsPoint(region, float64(raw.X), float64(raw.Y))

			ax, ay, ok := wm.SettlementAnchor("slovenia", koperIdx)
			if !ok {
				t.Fatal("settlement anchor olusturulamadi")
			}
			rawAX := int(shapeOffX + float64(raw.X)*shapeScaleX)
			rawAY := int(shapeOffY + float64(raw.Y)*shapeScaleY)
			if !rawInsideShape && ax == rawAX && ay == rawAY {
				t.Fatalf("anchor raw konumda kaldi: (%d,%d)", ax, ay)
			}
			if got := wm.RegionAt(ax, ay); got != "slovenia" {
				t.Fatalf("anchor yanlis bolgeye dustu: got=%q want=%q", got, "slovenia")
			}
		})
	}
}

func Test1300AzerbaijanShamakhiSettlementAnchorStaysInsideRegion(t *testing.T) {
	scenarioPath := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise")
	skipIfScenarioDirMissing(t, scenarioPath)

	regions, _, err := world.LoadRegionsWithOrder(filepath.Join(scenarioPath, "data", "regions.json"))
	if err != nil {
		t.Fatalf("regions yuklenemedi: %v", err)
	}
	if err := world.LoadRegionSettlements(filepath.Join(scenarioPath, "data", "settlements.json"), regions); err != nil {
		t.Fatalf("settlements yuklenemedi: %v", err)
	}
	shapeData, err := world.LoadCountryShapes(filepath.Join(scenarioPath, "data", "country_shapes.json"), regions)
	if err != nil {
		t.Fatalf("shapes yuklenemedi: %v", err)
	}

	region := regions["azerbaijan"]
	if region == nil {
		t.Fatal("azerbaijan bolgesi bulunamadi")
	}
	settlementIdx := -1
	for i, settlement := range region.Settlements {
		if settlement.ID == "azerbaijan_shamakhi" {
			settlementIdx = i
			if !regionContainsPoint(region, float64(settlement.X), float64(settlement.Y)) {
				t.Fatalf("shamakhi raw koordinati shape disinda: (%d,%d)", settlement.X, settlement.Y)
			}
			break
		}
	}
	if settlementIdx < 0 {
		t.Fatal("azerbaijan_shamakhi yerlesimi bulunamadi")
	}

	gs := &state.GameState{Regions: regions, ShapeData: shapeData}
	wm := NewWorldMap(gs)
	ax, ay, ok := wm.SettlementAnchor("azerbaijan", settlementIdx)
	if !ok {
		t.Fatal("shamakhi settlement anchor olusturulamadi")
	}
	if got := wm.RegionAt(ax, ay); got != "azerbaijan" {
		t.Fatalf("shamakhi anchor yanlis bolgeye dustu: got=%q want=%q", got, "azerbaijan")
	}
}

func Test1300MeccaSettlementAnchorUsesConfiguredCoordinate(t *testing.T) {
	scenarioPath := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise")
	skipIfScenarioDirMissing(t, scenarioPath)

	regions, _, err := world.LoadRegionsWithOrder(filepath.Join(scenarioPath, "data", "regions.json"))
	if err != nil {
		t.Fatalf("regions yuklenemedi: %v", err)
	}
	if err := world.LoadRegionSettlements(filepath.Join(scenarioPath, "data", "settlements.json"), regions); err != nil {
		t.Fatalf("settlements yuklenemedi: %v", err)
	}
	shapeData, err := world.LoadCountryShapes(filepath.Join(scenarioPath, "data", "country_shapes.json"), regions)
	if err != nil {
		t.Fatalf("shapes yuklenemedi: %v", err)
	}

	region := regions["hejaz"]
	if region == nil {
		t.Fatal("hejaz bolgesi bulunamadi")
	}
	settlementIdx := -1
	for i, settlement := range region.Settlements {
		if settlement.ID == "hejaz_mecca" {
			settlementIdx = i
			if !regionContainsPoint(region, float64(settlement.X), float64(settlement.Y)) {
				t.Fatalf("mekke koordinati shape disinda: (%d,%d)", settlement.X, settlement.Y)
			}
			break
		}
	}
	if settlementIdx < 0 {
		t.Fatal("hejaz_mecca yerlesimi bulunamadi")
	}

	gs := &state.GameState{Regions: regions, ShapeData: shapeData}
	wm := NewWorldMap(gs)
	settlement := region.Settlements[settlementIdx]
	rawX := int(shapeOffX + float64(settlement.X)*shapeScaleX)
	rawY := int(shapeOffY + float64(settlement.Y)*shapeScaleY)
	if got := wm.RegionAt(rawX, rawY); got != "hejaz" {
		t.Fatalf("mekke raw koordinati yanlis bolgeye dustu: got=%q want=%q", got, "hejaz")
	}
	ax, ay, ok := wm.SettlementAnchor("hejaz", settlementIdx)
	if !ok {
		t.Fatal("mekke settlement anchor olusturulamadi")
	}
	if ax != rawX || ay != rawY {
		t.Fatalf("mekke anchor fallback/duzeltme uyguladi: got=(%d,%d) raw=(%d,%d)", ax, ay, rawX, rawY)
	}
}

func skipIfScenarioDirMissing(t *testing.T, scenarioPath string) {
	t.Helper()

	info, err := os.Stat(scenarioPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("senaryo dizini mevcut değil: %s", scenarioPath)
		}
		t.Fatalf("senaryo dizini kontrol edilemedi: %v", err)
	}
	if !info.IsDir() {
		t.Skipf("senaryo yolu dizin değil: %s", scenarioPath)
	}
}
