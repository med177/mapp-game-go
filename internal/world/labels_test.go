package world

import "testing"

func TestTerrainLabelTR(t *testing.T) {
	if got := TerrainForest.LabelTR(); got != "Orman" {
		t.Fatalf("terrain label mismatch: got=%q", got)
	}
}

func TestSettlementTypeLabelTR(t *testing.T) {
	if got := SettlementPort.LabelTR(); got != "Liman" {
		t.Fatalf("settlement label mismatch: got=%q", got)
	}
}
