package world

import (
	"path/filepath"
	"testing"
)

func TestLoad1300TradeCentersLoadsHistoricalFlows(t *testing.T) {
	regions, err := LoadRegions(filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "regions.json"))
	if err != nil {
		t.Fatalf("LoadRegions() error = %v", err)
	}
	config, err := LoadTradeCenters(filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "trade_centers.json"), regions)
	if err != nil {
		t.Fatalf("LoadTradeCenters() error = %v", err)
	}
	if got, want := len(config.HistoricalFlows), 11; got != want {
		t.Fatalf("historical flow count = %d, want %d", got, want)
	}
	for _, flow := range config.HistoricalFlows {
		if flow.StartYear == 0 || flow.AmountPerTurn <= 0 || flow.GoldIncomePerTurn <= 0 {
			t.Fatalf("invalid historical flow: %+v", flow)
		}
	}
	for _, center := range config.Centers {
		if center.ID == "cape_route" {
			if len(center.CompetitionImpacts) != 2 || center.UnlockYear != 1498 {
				t.Fatalf("cape competition data = %+v", center)
			}
			return
		}
	}
	t.Fatal("cape_route center was not loaded")
}
