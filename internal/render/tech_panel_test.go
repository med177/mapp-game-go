package render

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"
)

func TestLayoutTechTreeWrapsAfterFourColumns(t *testing.T) {
	levels := [][]techNode{{
		{t: nil}, {t: nil}, {t: nil}, {t: nil}, {t: nil},
	}}

	_, _ = layoutTechTree(levels, techTreeMaxCols)

	for i := 1; i < 4; i++ {
		if levels[0][i].y != levels[0][0].y {
			t.Fatalf("ilk dort dugum ayni satirda kalmali, node0=%.1f node%d=%.1f", levels[0][0].y, i, levels[0][i].y)
		}
	}
	if levels[0][4].y <= levels[0][0].y {
		t.Fatalf("besinci dugum alt satira inmeliydi, firstRowY=%.1f wrappedY=%.1f", levels[0][0].y, levels[0][4].y)
	}
}

func TestTechTreeViewOriginCentersNarrowContent(t *testing.T) {
	treeRect := techPanelLayout{treeRect: gameui.Rect{W: 1000, H: 600}}.treeRect
	x, y := techTreeViewOrigin(treeRect, 700)
	if x <= 0 {
		t.Fatalf("dar icerik yatay merkezlenmeliydi, got x=%.1f", x)
	}
	if y != 0 {
		t.Fatalf("dikey origin sabit kalmali, got y=%.1f", y)
	}
}

func TestOrderTechTreeLevelsPrefersParentNeighborhood(t *testing.T) {
	levels := [][]techNode{
		{
			{t: &tech.Technology{ID: "a"}},
			{t: &tech.Technology{ID: "b"}},
		},
		{
			{t: &tech.Technology{ID: "child_b", Requires: []string{"b"}}},
			{t: &tech.Technology{ID: "child_a", Requires: []string{"a"}}},
		},
	}

	orderTechTreeLevels(levels)

	if levels[1][0].t.ID != "child_a" || levels[1][1].t.ID != "child_b" {
		t.Fatalf("child nodes should reorder toward parent order, got %s then %s", levels[1][0].t.ID, levels[1][1].t.ID)
	}
}

func TestTechEffectSummaryIncludesReadableBuffs(t *testing.T) {
	summary := techEffectSummary(&tech.Technology{
		Effects: tech.Effects{
			GoldPerRegion:       3,
			MarketGoldMod:       0.25,
			RevealEnemyStrength: true,
		},
	})

	if summary == "" {
		t.Fatal("summary should not be empty")
	}
	if summary != "Bölge altını +3  •  Ticaret +%25  •  Tam istihbarat" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}

func TestFactionCompletedTechPreviewCapsAndReportsHidden(t *testing.T) {
	gs := &state.GameState{
		TechTypes: map[string]*tech.Technology{
			"a": {ID: "a", NameTR: "A", Category: tech.CategoryMilitary},
			"b": {ID: "b", NameTR: "B", Category: tech.CategoryEconomy},
			"c": {ID: "c", NameTR: "C", Category: tech.CategoryNaval},
		},
	}
	f := &faction.Faction{Research: faction.ResearchState{Completed: map[string]bool{"a": true, "b": true, "c": true}}}

	names, hidden := factionCompletedTechPreview(gs, f, 2)
	if len(names) != 2 {
		t.Fatalf("iki teknoloji preview'da görünmeliydi, got=%d", len(names))
	}
	if hidden != 1 {
		t.Fatalf("bir teknoloji gizli kalmalıydı, got=%d", hidden)
	}
	if names[0] != "A" {
		t.Fatalf("kategori sırasına göre ilk teknoloji A olmalıydı, got=%q", names[0])
	}
}

func TestFactionTradeStatsCountsActiveRoutesAndExports(t *testing.T) {
	gs := &state.GameState{
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b", AmountPerTurn: 2, GoldPerUnit: 5},
			{FromFactionID: "b", ToFactionID: "a", AmountPerTurn: 1, GoldPerUnit: 7},
			{FromFactionID: "a", ToFactionID: "c", AmountPerTurn: 3, GoldPerUnit: 4, SuspendedTurns: 2},
		},
	}

	stats := factionTradeStats(gs, "a")
	if stats.RouteCount != 2 {
		t.Fatalf("aktif rota sayısı 2 olmalıydı, got=%d", stats.RouteCount)
	}
	if stats.SuspendedCount != 1 {
		t.Fatalf("askıdaki rota sayısı 1 olmalıydı, got=%d", stats.SuspendedCount)
	}
	if stats.PartnerCount != 1 {
		t.Fatalf("aktif partner sayısı 1 olmalıydı, got=%d", stats.PartnerCount)
	}
	if stats.ExportGold != 10 {
		t.Fatalf("ihracat geliri 10 olmalıydı, got=%d", stats.ExportGold)
	}
}
