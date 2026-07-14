package render

import (
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
			if regionContainsPoint(region, float64(raw.X), float64(raw.Y)) {
				t.Fatalf("beklenen raw nokta shape disinda olmaliydi: (%d,%d)", raw.X, raw.Y)
			}

			ax, ay, ok := wm.SettlementAnchor("slovenia", koperIdx)
			if !ok {
				t.Fatal("settlement anchor olusturulamadi")
			}
			rawAX := int(shapeOffX + float64(raw.X)*shapeScaleX)
			rawAY := int(shapeOffY + float64(raw.Y)*shapeScaleY)
			if ax == rawAX && ay == rawAY {
				t.Fatalf("anchor raw konumda kaldi: (%d,%d)", ax, ay)
			}
			if got := wm.RegionAt(ax, ay); got != "slovenia" {
				t.Fatalf("anchor yanlis bolgeye dustu: got=%q want=%q", got, "slovenia")
			}
		})
	}
}
