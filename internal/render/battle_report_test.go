package render

import (
	"testing"

	"mapp-game-go/internal/state"
)

func TestBattleReportSceneMappings(t *testing.T) {
	tests := []struct {
		scene     BattleScene
		wantTitle string
		wantSound string
		wantHint  string
	}{
		{BattleSceneLand, "Kara Muharebesi Raporu", "battle_land", "assets/ui/battle_land.png"},
		{BattleSceneNaval, "Deniz Muharebesi Raporu", "battle_naval", "assets/ui/battle_naval.png"},
		{BattleSceneAmphibious, "Çıkarma Muharebesi Raporu", "battle_amphibious", "assets/ui/battle_amphibious.png"},
		{BattleSceneSiege, "Kuşatma Hücumu Raporu", "battle_siege", "assets/ui/battle_siege.png"},
	}

	for _, tc := range tests {
		if got := battleReportDefaultTitleTR(tc.scene); got != tc.wantTitle {
			t.Fatalf("%s icin title mismatch: got=%q want=%q", tc.scene, got, tc.wantTitle)
		}
		if got := battleReportSoundKey(tc.scene); got != tc.wantSound {
			t.Fatalf("%s icin sound mismatch: got=%q want=%q", tc.scene, got, tc.wantSound)
		}
		if got := battleReportPrimaryImageHint(tc.scene); got != tc.wantHint {
			t.Fatalf("%s icin hint mismatch: got=%q want=%q", tc.scene, got, tc.wantHint)
		}
	}
}

func TestPrepareForTurnAdvanceClosesPanelsAndResetsMapMode(t *testing.T) {
	r := &Renderer{
		SelectedRegion:              "ankara",
		SelectedArmy:                "ordu_1",
		selectedFactionPanel:        "ai_1",
		selectedSettlementRegion:    "ankara",
		selectedSettlementIndex:     2,
		devNeighborListExpanded:     true,
		showRecruitPanel:            true,
		recruitUnitID:               "militia",
		recruitQty:                  4,
		showDiplomacy:               true,
		diplomacyFocus:              3,
		diplomacyScroll:             2,
		diplomacyActionFocus:        1,
		diplomacyTargetFaction:      "ai_2",
		diplomacyOfferHistoryBrowse: "ai_3",
		showTech:                    true,
		techCursor:                  5,
		techDragging:                true,
		showTrade:                   true,
		tradeScroll:                 7,
		tradeFactionFocus:           2,
		tradeGoodFocus:              1,
		tradeHoverIdx:               4,
		tradeCenterIdx:              3,
		mapMode:                     MapModeTrade,
		showEventCodex:              true,
		eventDetail:                 "olay detayi",
		showVictoryDetail:           true,
		victoryDetailScroll:         96,
	}

	r.PrepareForTurnAdvance()

	if r.SelectedRegion != "" || r.SelectedArmy != "" || r.selectedFactionPanel != "" {
		t.Fatalf("secimler temizlenmeliydi: region=%q army=%q faction=%q", r.SelectedRegion, r.SelectedArmy, r.selectedFactionPanel)
	}
	if r.selectedSettlementRegion != "" || r.selectedSettlementIndex != -1 {
		t.Fatalf("settlement paneli kapanmaliydi: region=%q idx=%d", r.selectedSettlementRegion, r.selectedSettlementIndex)
	}
	if r.showRecruitPanel || r.recruitUnitID != "" || r.recruitQty != 1 {
		t.Fatalf("recruit state sifirlanmaliydi: show=%t unit=%q qty=%d", r.showRecruitPanel, r.recruitUnitID, r.recruitQty)
	}
	if r.showDiplomacy || r.diplomacyFocus != 0 || r.diplomacyScroll != 0 || r.diplomacyActionFocus != 0 || r.diplomacyTargetFaction != "" || r.diplomacyOfferHistoryBrowse != "" {
		t.Fatalf("diplomasi state sifirlanmaliydi: show=%t focus=%d scroll=%d action=%d target=%q browse=%q", r.showDiplomacy, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyActionFocus, r.diplomacyTargetFaction, r.diplomacyOfferHistoryBrowse)
	}
	if r.showTech || r.techCursor != 0 || r.techDragging {
		t.Fatalf("tech state sifirlanmaliydi: show=%t cursor=%d dragging=%t", r.showTech, r.techCursor, r.techDragging)
	}
	if r.showTrade || r.tradeScroll != 0 || r.tradeFactionFocus != 0 || r.tradeGoodFocus != 0 || r.tradeHoverIdx != -1 || r.tradeCenterIdx != -1 || r.mapMode != MapModeNormal {
		t.Fatalf("trade/map state sifirlanmaliydi: show=%t scroll=%d faction=%d good=%d hover=%d center=%d mode=%v", r.showTrade, r.tradeScroll, r.tradeFactionFocus, r.tradeGoodFocus, r.tradeHoverIdx, r.tradeCenterIdx, r.mapMode)
	}
	if r.showEventCodex || r.eventDetail != "" || r.showVictoryDetail || r.victoryDetailScroll != 0 {
		t.Fatalf("popup state sifirlanmaliydi: codex=%t detail=%q victory=%t scroll=%.1f", r.showEventCodex, r.eventDetail, r.showVictoryDetail, r.victoryDetailScroll)
	}
	if r.devNeighborListExpanded {
		t.Fatal("debug neighbor paneli de kapanmaliydi")
	}
}

func TestWorldInputLockedByPhase(t *testing.T) {
	r := &Renderer{gs: &state.GameState{Phase: state.PhasePlayerTurn}}
	if r.worldInputLockedByPhase() {
		t.Fatal("oyuncu turunda dunya inputu kilitli olmamali")
	}
	r.gs.Phase = state.PhaseAITurn
	if !r.worldInputLockedByPhase() {
		t.Fatal("AI turunda dunya inputu kilitli olmali")
	}
	r.gs.Phase = state.PhaseTurnResolution
	if !r.worldInputLockedByPhase() {
		t.Fatal("turn resolution fazinda dunya inputu kilitli olmali")
	}
}
