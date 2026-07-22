package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestBattleReportCommanderProgressTextIncludesPromotionAndTrait(t *testing.T) {
	text := battleReportCommanderProgressText(BattleReportCommanderProgress{
		SideLabel:     "Saldıran",
		Name:          "Murat Bey",
		XPGained:      100,
		PreviousLevel: 1,
		CurrentLevel:  2,
		NewTraits:     []string{"Savaş Tecrübesi"},
	})
	for _, want := range []string{"Murat Bey", "Saldıran", "+100 XP", "Lv1 → Lv2", "Savaş Tecrübesi"} {
		if !strings.Contains(text, want) {
			t.Fatalf("komutan ilerleme metni %q içinde %q yok", text, want)
		}
	}
}

func TestBattleReportCommanderEffectTextsHandlePresentAndMissingCommander(t *testing.T) {
	side := BattleReportSide{
		CommanderName:               "Osman Bey",
		CommanderBattleEffects:      "Saldırı +6%  |  Moral +8%",
		CommanderOperationalEffects: "Hareket +1  |  Kuşatma +1/+1",
	}
	for _, want := range []string{"Osman Bey", "Saldırı +6%", "Hareket +1"} {
		full := battleReportCommanderNameText(side) + " " + battleReportCommanderBattleText(side) + " " + battleReportCommanderOperationalText(side)
		if !strings.Contains(full, want) {
			t.Fatalf("komutan blok metni %q içinde %q yok", full, want)
		}
	}

	empty := BattleReportSide{}
	for _, want := range []string{"Komutan: Yok", "Muharebe: katkı yok.", "Operasyon: katkı yok."} {
		full := battleReportCommanderNameText(empty) + " " + battleReportCommanderBattleText(empty) + " " + battleReportCommanderOperationalText(empty)
		if !strings.Contains(full, want) {
			t.Fatalf("boş komutan blok metni %q içinde %q yok", full, want)
		}
	}
}

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
		diplomacyHistoryVisible:     true,
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
	if r.showDiplomacy || r.diplomacyFocus != 0 || r.diplomacyScroll != 0 || r.diplomacyActionFocus != 0 || r.diplomacyTargetFaction != "" || r.diplomacyOfferHistoryBrowse != "" || r.diplomacyHistoryVisible {
		t.Fatalf("diplomasi state sifirlanmaliydi: show=%t focus=%d scroll=%d action=%d target=%q browse=%q history=%t", r.showDiplomacy, r.diplomacyFocus, r.diplomacyScroll, r.diplomacyActionFocus, r.diplomacyTargetFaction, r.diplomacyOfferHistoryBrowse, r.diplomacyHistoryVisible)
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

func TestSelectPlayerCapitalRegionSelectsActiveCapitalRegion(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "capital_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital_region": {
				ID:      "capital_region",
				OwnerID: "player",
				Settlements: []world.Settlement{
					{ID: "capital_city", NameTR: "Başkent", IsCapital: true},
				},
			},
			"other_region": {ID: "other_region", OwnerID: "player"},
		},
	}
	r := &Renderer{
		gs:                       gs,
		SelectedRegion:           "other_region",
		SelectedArmy:             "army_1",
		selectedSettlementRegion: "other_region",
		selectedSettlementIndex:  0,
	}

	if !r.SelectPlayerCapitalRegion() {
		t.Fatal("oyuncu başkenti seçilebilmeliydi")
	}
	if r.SelectedRegion != "capital_region" {
		t.Fatalf("başkent bölgesi seçilmedi: got=%q", r.SelectedRegion)
	}
	if r.SelectedArmy != "" {
		t.Fatalf("eski ordu seçimi temizlenmeliydi: got=%q", r.SelectedArmy)
	}
	if r.selectedSettlementRegion != "" || r.selectedSettlementIndex != -1 {
		t.Fatalf("eski settlement seçimi temizlenmeliydi: region=%q index=%d", r.selectedSettlementRegion, r.selectedSettlementIndex)
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
