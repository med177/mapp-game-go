package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/satisfaction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestSatisfactionDeltaPresentationUsesColoredDeltaAndSharedRect(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: "player", TaxRate: 30},
		},
	}
	region := gs.Regions["home"]
	positive := satisfaction.Calculate(gs, region)
	positive.Army = 8
	positive.Total = 8
	gs.Regions["home"].Satisfaction = 100

	rect, ok := regionSatisfactionDeltaRect(gs, "home")
	if !ok || rect.W <= 0 || rect.H <= 0 {
		t.Fatalf("memnuniyet delta alanı üretilmeliydi: rect=%+v ok=%t", rect, ok)
	}
	barX, _ := regionPanelTaxBarLayout(infoPanelX()+float32(panelPad), infoPanelW-float32(panelPad*2))
	if rect.X+rect.W >= float64(barX) {
		t.Fatalf("delta etiketi memnuniyet barına taşmamalı: rect=%+v barX=%v", rect, barX)
	}
	if !regionPanelInteractiveHitForTab(rect.X+rect.W/2, rect.Y+rect.H/2, gs, "home", regionPanelTabBuildings, 0) {
		t.Fatal("memnuniyet delta alanı cursor/input için etkileşimli olmalı")
	}
	if satisfactionDeltaText(8) != "+8" || satisfactionDeltaText(-5) != "-5" {
		t.Fatalf("delta metin formatı yanlış")
	}
	if satisfactionDeltaColor(8) == satisfactionDeltaColor(-5) {
		t.Fatalf("pozitif ve negatif delta renkleri ayrılmalı")
	}
	lines := satisfactionBreakdownLines(region, positive)
	joined := ""
	for _, line := range lines {
		joined += line.text + "\n"
	}
	if !strings.Contains(joined, "Vergi (%30): +0") || !strings.Contains(joined, "Toplam: +8") {
		t.Fatalf("popup hesap satırları eksik: %s", joined)
	}
}
