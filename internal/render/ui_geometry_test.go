package render

import (
	"math"
	"testing"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/save"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestInitialCameraScaleStartsCloserAndClampsToMax(t *testing.T) {
	oldScreenW, oldScreenH := ScreenWidth, ScreenHeight
	oldWorldW, oldWorldH := WorldW, WorldH
	defer func() {
		ScreenWidth = oldScreenW
		ScreenHeight = oldScreenH
		WorldW = oldWorldW
		WorldH = oldWorldH
	}()

	ScreenWidth, ScreenHeight = 1280, 720
	WorldW, WorldH = 2892, 1440

	minScale := minCameraScale()
	got := initialCameraScale()
	want := minScale * initialCameraZoomFactor
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("initialCameraScale fit ustune %.2fx yaklasmali: got=%.6f want=%.6f", initialCameraZoomFactor, got, want)
	}
	if got <= minScale {
		t.Fatalf("initialCameraScale fit scale'dan buyuk olmali: got=%.6f min=%.6f", got, minScale)
	}

	WorldW, WorldH = 120, 60
	if got := initialCameraScale(); math.Abs(got-float64(maxCameraZoomScale)) > 1e-9 {
		t.Fatalf("initialCameraScale max zoom'a clamp olmali: got=%.6f want=%.6f", got, float64(maxCameraZoomScale))
	}
}

func TestResetCameraFocusesPlayerCapitalOnInitialOpen(t *testing.T) {
	oldScreenW, oldScreenH := ScreenWidth, ScreenHeight
	oldWorldW, oldWorldH := WorldW, WorldH
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		ScreenWidth = oldScreenW
		ScreenHeight = oldScreenH
		WorldW = oldWorldW
		WorldH = oldWorldH
		shapeOffX = oldShapeOffX
		shapeOffY = oldShapeOffY
		shapeScaleX = oldShapeScaleX
		shapeScaleY = oldShapeScaleY
	}()

	ScreenWidth, ScreenHeight = 1280, 720
	WorldW, WorldH = 2892, 1440
	shapeOffX, shapeOffY = 0, 0
	shapeScaleX, shapeScaleY = 1, 1

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", CapitalSettlementID: "sogut"},
		},
		Regions: map[world.RegionID]*world.Region{
			"bithynia": {
				ID:      "bithynia",
				OwnerID: "osm",
				WorldX:  1090,
				WorldY:  468,
				Settlements: []world.Settlement{
					{ID: "sogut", NameTR: "Sogut", X: 1030, Y: 520, IsCenter: true},
				},
			},
		},
	}

	r := &Renderer{gs: gs}
	r.resetCamera()

	wantX, wantY := clampCameraCenter(wcX(1030), wcY(520), r.camScale)
	if math.Abs(r.camX-wantX) > 1e-9 || math.Abs(r.camY-wantY) > 1e-9 {
		t.Fatalf("kamera oyuncu baskentine odaklanmadi: got=(%.6f, %.6f) want=(%.6f, %.6f)", r.camX, r.camY, wantX, wantY)
	}
}

func TestResetCameraClampsInitialCapitalFocusNearMapEdge(t *testing.T) {
	oldScreenW, oldScreenH := ScreenWidth, ScreenHeight
	oldWorldW, oldWorldH := WorldW, WorldH
	oldShapeOffX, oldShapeOffY := shapeOffX, shapeOffY
	oldShapeScaleX, oldShapeScaleY := shapeScaleX, shapeScaleY
	defer func() {
		ScreenWidth = oldScreenW
		ScreenHeight = oldScreenH
		WorldW = oldWorldW
		WorldH = oldWorldH
		shapeOffX = oldShapeOffX
		shapeOffY = oldShapeOffY
		shapeScaleX = oldShapeScaleX
		shapeScaleY = oldShapeScaleY
	}()

	ScreenWidth, ScreenHeight = 1280, 720
	WorldW, WorldH = 2892, 1440
	shapeOffX, shapeOffY = 0, 0
	shapeScaleX, shapeScaleY = 1, 1

	gs := &state.GameState{
		PlayerFactionID: "ven",
		Factions: map[faction.FactionID]*faction.Faction{
			"ven": {ID: "ven", CapitalSettlementID: "venice"},
		},
		Regions: map[world.RegionID]*world.Region{
			"veneto": {
				ID:      "veneto",
				OwnerID: "ven",
				WorldX:  40,
				WorldY:  30,
				Settlements: []world.Settlement{
					{ID: "venice", NameTR: "Venedik", X: 24, Y: 18, IsCenter: true},
				},
			},
		},
	}

	r := &Renderer{gs: gs}
	r.resetCamera()

	wantX, wantY := clampCameraCenter(wcX(24), wcY(18), r.camScale)
	if math.Abs(r.camX-wantX) > 1e-9 || math.Abs(r.camY-wantY) > 1e-9 {
		t.Fatalf("kenar baskent focus'u clamp olmadi: got=(%.6f, %.6f) want=(%.6f, %.6f)", r.camX, r.camY, wantX, wantY)
	}
}

func TestCoreUIGeometryFitsCommonViewports(t *testing.T) {
	cases := []struct {
		w float64
		h float64
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = tc.w
		ScreenHeight = tc.h
		for _, btn := range buildBottomActionButtons(true) {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		for _, btn := range buildMapModeButtons() {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		assertButtonInside(t, tc.w, tc.h, buildTradeToggleButton())
		assertTradePanelInside(t, tc.w, tc.h)
		assertDiplomacyPanelInside(t, tc.w, tc.h)
		assertTechPanelInside(t, tc.w, tc.h)
		assertScreenStacksInside(t, tc.w, tc.h)
		assertButtonInside(t, tc.w, tc.h, buildTopDateHudMenuButton())
		assertButtonInside(t, tc.w, tc.h, buildEventDetailCloseButton())
		assertModalInside(t, tc.w, tc.h, buildConfirmDialogModal())
		assertModalInside(t, tc.w, tc.h, buildWarConfirmModal())
		assertModalInside(t, tc.w, tc.h, buildWarSummaryModal())
		assertModalInside(t, tc.w, tc.h, buildBattlePlanModal())
		assertModalInside(t, tc.w, tc.h, buildBattleReportModal())
		offerModal := buildDiplomacyOfferModal()
		assertModalInside(t, tc.w, tc.h, offerModal)
		acceptBtn, rejectBtn := buildDiplomacyOfferButtons()
		if acceptBtn.Label != "Kabul Et" {
			t.Fatalf("diplomasi teklifinde kabul düğmesi Kabul Et olmalı, got=%q", acceptBtn.Label)
		}
		assertButtonInside(t, tc.w, tc.h, acceptBtn)
		assertButtonInside(t, tc.w, tc.h, rejectBtn)
		if rejectBtn.X+rejectBtn.W > offerModal.Panel.Rect.X+offerModal.Panel.Rect.W/2 {
			t.Fatalf("diplomacy offer buttons should stay on the left side: modal=%+v accept=%+v reject=%+v", offerModal.Panel.Rect, acceptBtn, rejectBtn)
		}
		assertButtonInside(t, tc.w, tc.h, buildWarSummaryCloseButton())
		assertButtonInside(t, tc.w, tc.h, buildBattleReportCloseButton())
		assertButtonInside(t, tc.w, tc.h, buildBattleReportContinueButton())
		assertModalInside(t, tc.w, tc.h, buildVictoryDetailModal())
		assertModalInside(t, tc.w, tc.h, buildHistoricalEventModal())
		assertRectInside(t, tc.w, tc.h, aiDiagnosticPanelRect())
		for _, btn := range battlePlanCardRects() {
			assertRectInside(t, tc.w, tc.h, btn)
		}
		battleButtons, cancelBtn := buildBattlePlanButtons()
		for _, btn := range battleButtons {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		assertButtonInside(t, tc.w, tc.h, cancelBtn)
		for _, btn := range buildHistoricalEventChoiceButtons(2) {
			assertButtonInside(t, tc.w, tc.h, btn)
		}
		assertArmyPanelGeometry(t, tc.w, tc.h)
	}
}

func TestEmptyRegionDoesNotCreateSettlementLabel(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"ankara": {
					ID:     "ankara",
					NameTR: "Ankara",
				},
			},
		},
	}

	r.appendSettlementDraws(r.gs.Regions["ankara"])
	if got := len(r.regionLabelBuf); got != 0 {
		t.Fatalf("bos bolge icin yerlesim etiketi uretilmemeliydi, got=%d", got)
	}
}

func assertArmyPanelGeometry(t *testing.T, screenW, screenH float64) {
	t.Helper()
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = screenW
	ScreenHeight = screenH

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:       "a1",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units: []army.Unit{
					{TypeID: "infantry", CurrentHP: 100},
					{TypeID: "infantry", CurrentHP: 100},
				},
			},
			"a2": {
				ID:       "a2",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units: []army.Unit{
					{TypeID: "infantry", CurrentHP: 100},
				},
			},
		},
	}

	layout := armyPanelGeometry()
	px, py, panelW := layout.panelX, layout.panelY, layout.panelW
	if px < 0 || py < 0 || px+panelW > float32(screenW) {
		t.Fatalf("army panel outside %.0fx%.0f viewport: px=%.1f py=%.1f w=%.1f", screenW, screenH, px, py, panelW)
	}

	splitBtn, ok := buildSplitArmyButton(gs, "a1")
	if !ok {
		t.Fatalf("split button bekleniyordu")
	}
	mergeBtn, ok := buildMergeArmyButton(gs, "a1")
	if !ok {
		t.Fatalf("merge button bekleniyordu")
	}
	commanderPortrait, ok := commanderPortraitHitRect(gs, "a1")
	if !ok {
		t.Fatalf("commander portrait hit rect bekleniyordu")
	}
	assertButtonInside(t, screenW, screenH, splitBtn)
	assertButtonInside(t, screenW, screenH, mergeBtn)
	assertRectInside(t, screenW, screenH, commanderPortrait)
	if splitBtn.Y < float64(py)+float64(armyPanelBtnY)-0.5 {
		t.Fatalf("split button header satirina cok yakin: panelY=%.1f btnY=%.1f", py, splitBtn.Y)
	}
	if mergeBtn.Y < float64(py)+float64(armyPanelBtnY)-0.5 {
		t.Fatalf("merge button header satirina cok yakin: panelY=%.1f btnY=%.1f", py, mergeBtn.Y)
	}
	if splitBtn.Y+splitBtn.H > float64(py)+float64(armyPanelHdrH)+0.5 {
		t.Fatalf("split button header alanini tasiyor: panelY=%.1f btn=%+v", py, splitBtn)
	}
	if mergeBtn.Y+mergeBtn.H > float64(py)+float64(armyPanelHdrH)+0.5 {
		t.Fatalf("merge button header alanini tasiyor: panelY=%.1f btn=%+v", py, mergeBtn)
	}
	if mergeBtn.X+mergeBtn.W >= splitBtn.X {
		t.Fatalf("BİRLEŞTİR solda, BÖL sağda olmalıydı: split=%+v merge=%+v", splitBtn, mergeBtn)
	}
	statusTextLeft := float64(px+panelW-armyPanelPadX) - MeasureText("Takviye aktif", FaceSmall)
	if mergeBtn.X+mergeBtn.W > statusTextLeft-8 {
		t.Fatalf("merge butonu sağ durum metnine giriyor: merge=%+v statusTextLeft=%.1f", mergeBtn, statusTextLeft)
	}
	if commanderPortrait.X < float64(layout.commanderX)-0.5 || commanderPortrait.X+commanderPortrait.W > float64(layout.commanderX+layout.commanderW)+0.5 {
		t.Fatalf("commander portrait commander kolonundan tasiyor: rect=%+v layout=%+v", commanderPortrait, layout)
	}
	if commanderPortrait.Y < float64(layout.commanderY)-0.5 || commanderPortrait.Y+commanderPortrait.H > float64(layout.commanderY+layout.commanderH)+0.5 {
		t.Fatalf("commander portrait commander kartindan tasiyor: rect=%+v layout=%+v", commanderPortrait, layout)
	}
	if commanderPortrait.X+commanderPortrait.W >= splitBtn.X {
		t.Fatalf("commander portrait birim aksiyonlariyla cakisiyor: portrait=%+v split=%+v", commanderPortrait, splitBtn)
	}
}

func TestArmyPanelSplitButtonStaysRightmost(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"split_and_merge": {
				ID:       "split_and_merge",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units: []army.Unit{
					{TypeID: "infantry", CurrentHP: 100},
					{TypeID: "infantry", CurrentHP: 100},
				},
			},
			"merge_only": {
				ID:       "merge_only",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
			"merge_target": {
				ID:       "merge_target",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
		},
	}

	splitBtn, splitOK := buildSplitArmyButton(gs, "split_and_merge")
	mergeBtn, mergeOK := buildMergeArmyButton(gs, "split_and_merge")
	if !splitOK || !mergeOK {
		t.Fatalf("iki aksiyon da görünür olmalıydı: split=%v merge=%v", splitOK, mergeOK)
	}
	if splitBtn.X <= mergeBtn.X || splitBtn.X+splitBtn.W <= mergeBtn.X+mergeBtn.W {
		t.Fatalf("BÖL sağdaki düğme olmalıydı: split=%+v merge=%+v", splitBtn, mergeBtn)
	}

	mergeOnlyBtn, mergeOnlyOK := buildMergeArmyButton(gs, "merge_only")
	if !mergeOnlyOK {
		t.Fatal("yalnız BİRLEŞTİR aksiyonu görünür olmalıydı")
	}
	mergeOnlyTargets := FindMergeTargets(gs, "merge_only")
	if len(mergeOnlyTargets) != 2 {
		t.Fatalf("merge_only için iki ayrı hedef butonu bekleniyordu: got=%v", mergeOnlyTargets)
	}
	mergeOnlySecond, secondOK := buildMergeArmyButtonForTarget(gs, "merge_only", mergeOnlyTargets[1], 1, len(mergeOnlyTargets))
	if !secondOK || mergeOnlyBtn.X >= mergeOnlySecond.X {
		t.Fatalf("hedef butonları soldan sağa sıralanmalıydı: first=%+v second=%+v", mergeOnlyBtn, mergeOnlySecond)
	}
}

func TestMergeButtonsCarryTargetAndResultingUnitCount(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Armies: map[army.ArmyID]*army.Army{
			"source": {
				ID: "source", OwnerID: "osm", RegionID: "ankara",
				Units: []army.Unit{{TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"}},
			},
			"target_a": {
				ID: "target_a", OwnerID: "osm", RegionID: "ankara",
				Units: make([]army.Unit, 7),
			},
			"target_b": {
				ID: "target_b", OwnerID: "osm", RegionID: "ankara",
				Units: make([]army.Unit, 5),
			},
		},
	}
	for _, target := range gs.Armies {
		for i := range target.Units {
			target.Units[i].TypeID = "infantry"
		}
	}

	targets := FindMergeTargets(gs, "source")
	if len(targets) != 2 || targets[0] != "target_a" || targets[1] != "target_b" {
		t.Fatalf("birleştirme hedefleri deterministik olmalıydı: %v", targets)
	}
	first, firstOK := buildMergeArmyButtonForTarget(gs, "source", targets[0], 0, len(targets))
	second, secondOK := buildMergeArmyButtonForTarget(gs, "source", targets[1], 1, len(targets))
	if !firstOK || !secondOK {
		t.Fatal("iki ayrı birleştirme butonu bekleniyordu")
	}
	if first.Label != "->7" || second.Label != "->5" {
		t.Fatalf("buton etiketleri hedef orduların mevcut birim sayısını göstermeliydi: first=%q second=%q", first.Label, second.Label)
	}
	if targetID, ok := MergeButtonTargetAt(first.X+first.W/2, first.Y+first.H/2, gs, "source"); !ok || targetID != targets[0] {
		t.Fatalf("ilk buton target_a ordusunu seçmeliydi: got=%q ok=%v", targetID, ok)
	}
	if targetID, ok := MergeButtonTargetAt(second.X+second.W/2, second.Y+second.H/2, gs, "source"); !ok || targetID != targets[1] {
		t.Fatalf("ikinci buton target_b ordusunu seçmeliydi: got=%q ok=%v", targetID, ok)
	}
}

func assertScreenStacksInside(t *testing.T, screenW, screenH float64) {
	t.Helper()
	scenarioStack := centeredStackRect(3, 560, 130, 16, 20)
	if scenarioStack.X < 0 || scenarioStack.Y < 0 || scenarioStack.X+scenarioStack.W > screenW || scenarioStack.Y+scenarioStack.H > screenH {
		t.Fatalf("scenario stack outside %.0fx%.0f viewport: %+v", screenW, screenH, scenarioStack)
	}
	factionGrid := centeredGridRect(3, 2, 350, 110, 30, 12, 70)
	if factionGrid.X < 0 || factionGrid.Y < 0 || factionGrid.X+factionGrid.W > screenW || factionGrid.Y+factionGrid.H > screenH {
		t.Fatalf("faction grid outside %.0fx%.0f viewport: %+v", screenW, screenH, factionGrid)
	}
	settingsStack := centeredStackRect(6, 500, 56, 4, 40)
	if settingsStack.X < 0 || settingsStack.Y < 0 || settingsStack.X+settingsStack.W > screenW || settingsStack.Y+settingsStack.H > screenH {
		t.Fatalf("settings stack outside %.0fx%.0f viewport: %+v", screenW, screenH, settingsStack)
	}
	cardW, cardH := victoryCardDimensions()
	victoryLayout := victoryLayout(4, 2, cardW, cardH, 12, 80)
	victoryStack := victoryLayout.generalStack
	if victoryLayout.historicalStack.W > victoryStack.W {
		victoryStack = victoryLayout.historicalStack
	}
	if victoryStack.X < 0 || victoryLayout.historicalLabel.Y < 0 || victoryStack.X+victoryStack.W > screenW || victoryStack.Y+victoryStack.H > screenH {
		t.Fatalf("victory stack outside %.0fx%.0f viewport: historical=%+v general=%+v", screenW, screenH, victoryLayout.historicalStack, victoryLayout.generalStack)
	}
	loadStack := centeredStackRect(5, 480, 88, 14, 0)
	if loadStack.X < 0 || loadStack.Y < 0 || loadStack.X+loadStack.W > screenW || loadStack.Y+loadStack.H > screenH {
		t.Fatalf("load stack outside %.0fx%.0f viewport: %+v", screenW, screenH, loadStack)
	}
}

func assertDiplomacyPanelInside(t *testing.T, screenW, screenH float64) {
	t.Helper()
	list := diplomacyListLayoutForScreen()
	if list.panelRect.X < 0 || list.panelRect.Y < 0 || list.panelRect.X+list.panelRect.W > screenW || list.panelRect.Y+list.panelRect.H > screenH {
		t.Fatalf("diplomacy list panel outside %.0fx%.0f viewport: %+v", screenW, screenH, list)
	}
	if list.historyRect.W > 0 {
		assertRectInside(t, screenW, screenH, list.historyRect)
		assertButtonInside(t, screenW, screenH, buildDiplomacySideViewButton(list.historyRect, false))
	}
	offer := diplomacyOfferLayoutForScreen()
	if offer.panelRect.X < 0 || offer.panelRect.Y < 0 || offer.panelRect.X+offer.panelRect.W > screenW || offer.panelRect.Y+offer.panelRect.H > screenH {
		t.Fatalf("diplomacy offer panel outside %.0fx%.0f viewport: %+v", screenW, screenH, offer)
	}
	if offer.historyRect.W > 0 {
		assertRectInside(t, screenW, screenH, offer.historyRect)
		assertButtonInside(t, screenW, screenH, buildDiplomacySideViewButton(offer.historyRect, true))
		management := buildDiplomacyVassalManagementLayout()
		assertRectInside(t, screenW, screenH, management.panelRect)
		assertButtonInside(t, screenW, screenH, management.releaseButton)
		assertButtonInside(t, screenW, screenH, management.annexButton)
		buttons := buildDiplomacyHistoryFilterButtons(offer.historyRect, diplomacyHistoryDirectionAll, ActionNone)
		actionRowBottom := buttons[3].Button.Y + buttons[3].Button.H
		firstCardY := diplomacyOfferHistoryCardRect(offer.historyRect, 0).Y
		if firstCardY < actionRowBottom+4 {
			t.Fatalf("diplomacy history cards start too close to filter buttons in %.0fx%.0f viewport: firstCardY=%.1f actionRowBottom=%.1f layout=%+v", screenW, screenH, firstCardY, actionRowBottom, offer)
		}
	}
	statusSecondLineY := offer.statusRect.Y + 8 + 24
	actionLabelY := offer.actionsRect.Y - 18
	if statusSecondLineY+14 > actionLabelY {
		t.Fatalf("diplomacy offer status and action label overlap in %.0fx%.0f viewport: statusSecondLineY=%.1f actionLabelY=%.1f layout=%+v", screenW, screenH, statusSecondLineY, actionLabelY, offer)
	}
	assertButtonInside(t, screenW, screenH, buildDiplomacyCloseButton())
	assertButtonInside(t, screenW, screenH, buildDiplomacyBackButton())
	assertButtonInside(t, screenW, screenH, buildDiplomacySendButton())
	for _, btn := range buildDiplomacyActionButtons(nil, "") {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
}

func assertTechPanelInside(t *testing.T, screenW, screenH float64) {
	t.Helper()
	layout := techPanelLayoutForScreen()
	if layout.panelRect.X < 0 || layout.panelRect.Y < 0 || layout.panelRect.X+layout.panelRect.W > screenW || layout.panelRect.Y+layout.panelRect.H > screenH {
		t.Fatalf("tech panel outside %.0fx%.0f viewport: %+v", screenW, screenH, layout)
	}
	assertButtonInside(t, screenW, screenH, buildTechCloseButton())
}

func TestTechTreeSideBlankClickClosesPanel(t *testing.T) {
	oldScreenW, oldScreenH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldScreenW, oldScreenH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
		},
		TechTypes: map[string]*tech.Technology{
			"root": {ID: "root", NameTR: "Kök", Category: tech.CategoryMilitary},
		},
	}

	for _, side := range []string{"sol", "sağ"} {
		r := &Renderer{gs: gs, showTech: true}
		f := gs.Factions[gs.PlayerFactionID]
		layout := techPanelLayoutForScreen()
		treeData := r.buildLaidOutTechTree(f)
		originX, _ := techTreeViewOrigin(layout.treeRect, treeData.contentW)
		flowX := layout.treeRect.X + originX
		clickX := flowX - 8
		if side == "sağ" {
			clickX = flowX + treeData.contentW + 8
		}

		r.handleTechInput(f, gameui.InputState{
			MouseX:          clickX,
			MouseY:          layout.treeRect.Y + 20,
			LeftJustPressed: true,
		})
		if r.showTech {
			t.Fatalf("teknoloji paneli %s flow boşluğuna tıklayınca kapanmalı", side)
		}
	}
}

func assertTradePanelInside(t *testing.T, screenW, screenH float64) {
	t.Helper()
	layout := tradePanelLayout()
	const eps = 0.5
	if layout.panelRect.X < 0 || layout.panelRect.Y < 0 || layout.panelRect.X+layout.panelRect.W > screenW || layout.panelRect.Y+layout.panelRect.H > screenH {
		t.Fatalf("trade panel outside %.0fx%.0f viewport: %+v", screenW, screenH, layout)
	}
	if layout.rightListRect.X+layout.rightListRect.W > layout.panelRect.X+layout.panelRect.W-8+eps {
		t.Fatalf("trade panel right column overflow in %.0fx%.0f viewport: %+v", screenW, screenH, layout)
	}
	cardX, cardY, cardW, cardH := tradeActionCardRect(layout, len(tradeSelectableGoods()))
	if float64(cardX) < layout.rightListRect.X-eps || float64(cardX+cardW) > layout.panelRect.X+layout.panelRect.W-8+eps {
		t.Fatalf("trade action card overflow in %.0fx%.0f viewport: layout=%+v card=(%.1f,%.1f,%.1f,%.1f)", screenW, screenH, layout, cardX, cardY, cardW, cardH)
	}
	if float64(cardY) < layout.rightListRect.Y+float64(tradeGoodsListHeight(len(tradeSelectableGoods())))-eps {
		t.Fatalf("trade action card overlaps goods list in %.0fx%.0f viewport: layout=%+v card=(%.1f,%.1f,%.1f,%.1f)", screenW, screenH, layout, cardX, cardY, cardW, cardH)
	}
	for _, btn := range buildTradeTabButtons() {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	for _, btn := range buildTradeFilterButtons(layout) {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	for _, btn := range buildTradeSortButtons(layout) {
		assertButtonInside(t, screenW, screenH, btn.Button)
	}
	assertButtonInside(t, screenW, screenH, buildTradeAutoExportButton(layout, false))
	qtyButtons, buyBtn, sellBtn := buildTradeActionButtons(layout, len(tradeSelectableGoods()))
	for _, btn := range qtyButtons {
		assertButtonInside(t, screenW, screenH, btn)
	}
	assertButtonInside(t, screenW, screenH, buyBtn)
	assertButtonInside(t, screenW, screenH, sellBtn)
	assertButtonInside(t, screenW, screenH, buildTradeEmergencyGrainSaleButton(layout, len(tradeSelectableGoods()), true))
}

func TestTradeFilterPredicates(t *testing.T) {
	if !isTradeSellerForPlayer(1) {
		t.Fatal("stocku olan fraksiyon seller sayilmali")
	}
	if isTradeSellerForPlayer(0) {
		t.Fatal("stogu olmayan fraksiyon seller sayilmamali")
	}
	if !isTradeBuyerForPlayer(20, 5, 2, 50) {
		t.Fatal("oyuncudan daha az stogu olan ve odeyebilen fraksiyon buyer sayilmali")
	}
	if isTradeBuyerForPlayer(20, 25, 2, 50) {
		t.Fatal("oyuncudan fazla stogu olan fraksiyon buyer sayilmamali")
	}
	if isTradeBuyerForPlayer(0, 0, 2, 50) {
		t.Fatal("oyuncuda mal yoksa buyer olusmamali")
	}
}

func TestMainMenuRenderSmokeCommonViewports(t *testing.T) {
	cases := []struct {
		w int
		h int
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = float64(tc.w)
		ScreenHeight = float64(tc.h)
		screen := ebiten.NewImage(tc.w, tc.h)
		DrawMainMenu(screen, 0, true, true, true, 0)
	}
}

func TestSelectionScreensRenderSmokeCommonViewports(t *testing.T) {
	scenarios := []*scenario.Scenario{
		{Name: "1300 Anadolu", Description: "Anadolu beylikleri, Doğu Roma kalıntıları ve yükselen Osmanlı arasında geçen yoğun başlangıç.", Year: 1300, Month: 3},
		{Name: "Doğu Akdeniz", Description: "Levant, Mısır ve deniz ticaret merkezleri etrafında şekillenen ekonomik mücadele.", Year: 1350, Month: 9},
	}
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı", Religion: religion.Sunni, Color: [3]uint8{180, 60, 40}, IsPlayable: true},
			"ven": {ID: "ven", NameTR: "Venedik", Religion: religion.Catholic, Color: [3]uint8{60, 120, 180}, IsPlayable: true},
			"mem": {ID: "mem", NameTR: "Memlük", Religion: religion.Sunni, Color: [3]uint8{160, 140, 60}, IsPlayable: true},
		},
		AvailableVictories: []scenario.VictoryOptionDef{
			{Title: "Toprak Hakimiyeti", Description: "Belirli sayıda bölgeyi elde tut.", Detail: "20 bölge ve kritik şehirler"},
			{Title: "Ekonomik Güç", Description: "Altın gelirini uzun süre koru.", Detail: "500 altın gelir, 5 tur"},
		},
	}
	oldSlots := SaveSlots
	SaveSlots = []save.SaveSlot{
		{Name: "autosave", DisplayName: "Otomatik Kayıt", Exists: true, FactionName: "Osmanlı", Turn: 12, Year: 1301, ModTime: time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC)},
		{Name: "slot1", DisplayName: "Kayıt 1", Exists: false},
	}
	defer func() { SaveSlots = oldSlots }()

	cases := []struct {
		w int
		h int
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = float64(tc.w)
		ScreenHeight = float64(tc.h)
		screen := ebiten.NewImage(tc.w, tc.h)
		DrawScenarioSelect(screen, scenarios, 0)
		DrawFactionSelect(screen, gs, 0)
		DrawVictorySelect(screen, gs, 0)
		DrawSlotSelectScreen(screen, 0, false, "")
		DrawSlotSelectScreen(screen, 0, false, "autosave")
		DrawSettingsScreen(screen, DefaultSettings(), 0)
	}
}

func TestSlotPendingDeleteLayoutDoesNotOverlap(t *testing.T) {
	_, nameBottom, promptTop, promptBottom := slotPendingDeleteTextBounds(100)
	if nameBottom > promptTop {
		t.Fatalf("pending delete metinleri üst üste biniyor: nameBottom=%.1f promptTop=%.1f", nameBottom, promptTop)
	}
	if promptBottom > 100+slotCardH {
		t.Fatalf("pending delete sorusu kart dışına taşıyor: promptBottom=%.1f cardBottom=%.1f", promptBottom, 100+slotCardH)
	}
}

func TestPanelFamilyRenderSmokeCommonViewports(t *testing.T) {
	regionID := world.RegionID("ankara")
	seaID := world.RegionID("_sea_aegean")
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Turn:            12,
		Year:            1301,
		Month:           3,
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {
				ID:         "osm",
				NameTR:     "Osmanlı",
				Religion:   religion.Sunni,
				Color:      [3]uint8{180, 60, 40},
				IsPlayable: true,
				Gold:       300,
				Grain:      120,
				Iron:       40,
				Timber:     55,
				Stone:      22,
				Research:   faction.ResearchState{Completed: map[string]bool{}},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {
				ID:              regionID,
				NameTR:          "Ankara",
				Terrain:         world.TerrainPlain,
				OwnerID:         "osm",
				Neighbors:       []world.RegionID{seaID},
				WorldX:          100,
				WorldY:          100,
				BaseGoldIncome:  12,
				BaseGrainOutput: 8,
				Satisfaction:    65,
				TaxRate:         15,
				Population:      240,
				Religion:        string(religion.Sunni),
				Buildings:       []string{"market"},
				Settlements: []world.Settlement{
					{ID: "ankara_city", NameTR: "Ankara", X: 100, Y: 100, Type: world.SettlementCity, IsCenter: true},
				},
			},
			seaID: {
				ID:        seaID,
				NameTR:    "Ege Denizi",
				Terrain:   world.TerrainSea,
				IsSea:     true,
				Neighbors: []world.RegionID{regionID},
				WorldX:    120,
				WorldY:    120,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:            "a1",
				OwnerID:       "osm",
				RegionID:      regionID,
				MovePoints:    4,
				MaxMovePoints: 6,
				Units: []army.Unit{
					{TypeID: "militia", CurrentHP: 100},
				},
			},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar", MaxPerRegion: 1},
			"farm":   {ID: "farm", NameTR: "Çiftlik", MaxPerRegion: 1},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":  {ID: "militia", NameTR: "Milis", TurnsRequired: 1},
			"infantry": {ID: "infantry", NameTR: "Piyade", TurnsRequired: 2},
		},
		ProductionQueue: []state.ProductionOrder{
			{Kind: "building", RegionID: regionID, TypeID: "farm", TurnsLeft: 3},
		},
		Relations: map[string]*faction.Relation{},
	}

	cases := []struct {
		w int
		h int
	}{
		{1280, 720},
		{1600, 900},
		{1920, 1080},
	}
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()

	for _, tc := range cases {
		ScreenWidth = float64(tc.w)
		ScreenHeight = float64(tc.h)
		screen := ebiten.NewImage(tc.w, tc.h)
		DrawBottomPanel(screen, gs, true, true, "", true, true, false, MapModeNormal)
		DrawRegionPanel(screen, gs, regionID)
		DrawArmyPanel(screen, gs, "a1")
		DrawSettlementPanel(screen, gs, gs.Regions[regionID], &gs.Regions[regionID].Settlements[0])
		DrawSeaRegionPanel(screen, gs, gs.Regions[seaID], false)
		DrawRecruitPanel(screen, gs, regionID, "", 0)
	}
}

func TestVictoryChecklistEntriesReflectOwnership(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "osm",
		Victory: state.VictoryCondition{
			Type:            state.VictoryConquerCity,
			RequiredRegions: []world.RegionID{"constantinople", "ankara"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"osm":     {ID: "osm", NameTR: "Osmanlı"},
			"karaman": {ID: "karaman", NameTR: "Karamanoğulları"},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", NameTR: "Konstantinopolis", OwnerID: "osm"},
			"ankara":         {ID: "ankara", NameTR: "Ankara", OwnerID: "karaman"},
		},
	}

	lines, _ := victoryChecklistEntries(gs)
	if len(lines) != 2 {
		t.Fatalf("2 checklist satiri bekleniyordu, got=%d", len(lines))
	}
	if lines[0] != "✓ Konstantinopolis" {
		t.Fatalf("ilk satir beklenmedik: %q", lines[0])
	}
	if lines[1] != "✗ Ankara (Karamanoğulları)" {
		t.Fatalf("ikinci satir beklenmedik: %q", lines[1])
	}
}

func assertButtonInside(t *testing.T, screenW, screenH float64, btn gameui.Button) {
	t.Helper()
	if btn.X < 0 || btn.Y < 0 || btn.X+btn.W > screenW || btn.Y+btn.H > screenH {
		t.Fatalf("button %q outside %.0fx%.0f viewport: %+v", btn.Label, screenW, screenH, btn)
	}
}

func assertRectInside(t *testing.T, screenW, screenH float64, rect gameui.Rect) {
	t.Helper()
	if rect.X < 0 || rect.Y < 0 || rect.X+rect.W > screenW || rect.Y+rect.H > screenH {
		t.Fatalf("rect outside %.0fx%.0f viewport: %+v", screenW, screenH, rect)
	}
}

func assertModalInside(t *testing.T, screenW, screenH float64, modal gameui.Modal) {
	t.Helper()
	p := modal.Panel.Rect
	if p.X < 0 || p.Y < 0 || p.X+p.W > screenW || p.Y+p.H > screenH {
		t.Fatalf("modal panel outside %.0fx%.0f viewport: %+v", screenW, screenH, p)
	}
}
