package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestZoomInLoopLevelsUseFinalFiveZoomSteps(t *testing.T) {
	threshold := maxCameraZoomScale / (cameraZoomStep * cameraZoomStep * cameraZoomStep * cameraZoomStep)
	if !isZoomInLoopLevel(threshold) {
		t.Fatal("son beş zoom seviyesinin alt sınırı loop seviyesi kabul edilmedi")
	}
	if isZoomInLoopLevel(threshold - 0.01) {
		t.Fatal("son beş zoom seviyesinin altındaki ölçek loop seviyesi kabul edildi")
	}
}

func TestTurnTechHudMetricHitRectsMatchDisplayedTexts(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}
	fatigueRect, routeRect := turnTechHudMetricRects(gs)
	if fatigueRect.W <= 0 || routeRect.W <= 0 {
		t.Fatalf("HUD metin rectleri oluşturulmadı: fatigue=%+v route=%+v", fatigueRect, routeRect)
	}
	if !turnTechHudWarFatigueHit(gs, fatigueRect.X+fatigueRect.W/2, fatigueRect.Y+fatigueRect.H/2) {
		t.Fatal("savaş yorgunluğu metin rectinin merkezi tıklanabilir değil")
	}
	if !turnTechHudTradeRouteHit(gs, routeRect.X+routeRect.W/2, routeRect.Y+routeRect.H/2) {
		t.Fatal("ticaret rotası metin rectinin merkezi tıklanabilir değil")
	}
	if turnTechHudTradeRouteHit(gs, routeRect.X-4, routeRect.Y+routeRect.H/2) {
		t.Fatal("ticaret rotası rectinin dışı tıklanabilir kabul edildi")
	}
}

func TestArmyOrganizationHUDValueRectMatchesDisplayedValue(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions:         map[world.RegionID]*world.Region{},
		Armies:          map[army.ArmyID]*army.Army{},
	}
	for i := 0; i < 8; i++ {
		id := world.RegionID("region_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "player"}
	}
	for i := 0; i < 6; i++ {
		id := army.ArmyID("army_" + string(rune('a'+i)))
		gs.Armies[id] = &army.Army{ID: id, OwnerID: "player", Units: []army.Unit{{TypeID: "infantry"}}}
	}

	rect := armyOrganizationHUDValueRect(gs)
	if rect.W <= 0 || rect.H <= 0 {
		t.Fatalf("ordu organizasyonu değer recti geçersiz: %+v", rect)
	}
	if !rect.Hit(rect.X+rect.W/2, rect.Y+rect.H/2) {
		t.Fatal("ordu organizasyonu değerinin merkezi hover alanında değil")
	}
	if armyOrganizationHUDValueRect(gs).Hit(rect.X-4, rect.Y+rect.H/2) {
		t.Fatal("ordu organizasyonu değer rectinin dışı hover alanı kabul edildi")
	}
}

func TestManpowerHUDValuesShareRightEdgeAndStayInsideCard(t *testing.T) {
	values := []string{"130/375", "12/12", "114/147"}
	wantRight := float64(manpowerHUDCardX+manpowerHUDCardW) - manpowerHUDCardPad
	for row, value := range values {
		displayed, rect := manpowerHUDRowValueRect(float64(row*22), value)
		if displayed != value {
			t.Fatalf("%q kapasite değeri gereksiz kırpıldı: %q", value, displayed)
		}
		if rect.X+rect.W != wantRight {
			t.Fatalf("%q sağ kenarı = %.2f, want %.2f", value, rect.X+rect.W, wantRight)
		}
		if rect.X < float64(manpowerHUDCardX) || rect.X+rect.W > float64(manpowerHUDCardX+manpowerHUDCardW) {
			t.Fatalf("%q kart dışına taşıyor: rect=%+v", value, rect)
		}
	}
}

func TestCoastalTargetDialogButtonPresentation(t *testing.T) {
	embark := decorateConfirmDialogButton(gameui.NewButton(0, 0, 100, 40, "Gemiye Bin"), "Gemiye Bin", "accept")
	if embark.Icon != gameui.IconLoad {
		t.Fatalf("gemiye bin butonu ikonu = %q, want %q", embark.Icon, gameui.IconLoad)
	}

	send := decorateConfirmDialogButton(gameui.NewButton(0, 0, 100, 40, "Sevk Et"), "Sevk Et", "accept")
	if send.Icon != gameui.IconExport {
		t.Fatalf("sevk butonu ikonu = %q, want %q", send.Icon, gameui.IconExport)
	}

	dock := decorateConfirmDialogButton(gameui.NewButton(0, 0, 100, 40, "Limana Gir"), "Limana Gir", "third")
	if dock.Icon != gameui.IconDock {
		t.Fatalf("limana gir butonu ikonu = %q, want %q", dock.Icon, gameui.IconDock)
	}

	style := confirmDialogThirdButtonStyle("Limana Gir")
	wantBG := color.RGBA{55, 92, 142, 240}
	if style.BG != wantBG {
		t.Fatalf("limana gir buton arka planı = %#v, want %#v", style.BG, wantBG)
	}
}

func TestNavalCoastalTargetDialogKeepsBothPortIntentActions(t *testing.T) {
	const targetRegion = world.RegionID("target")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			targetRegion: {
				ID:        targetRegion,
				OwnerID:   "player",
				Buildings: []string{"port"},
				Settlements: []world.Settlement{
					{ID: "target-port", Type: world.SettlementPort},
				},
			},
		},
	}
	fleet := &army.Army{
		ID:            "fleet",
		OwnerID:       "player",
		IsNaval:       true,
		EmbarkedUnits: []army.Unit{{TypeID: "infantry"}},
	}
	r := &Renderer{gs: gs}

	if !r.openNavalCoastalTargetDialog(fleet, gs.Regions[targetRegion], targetRegion, "target-port") {
		t.Fatal("geçerli liman hedefi kıyı niyet modalını açmadı")
	}
	if !r.confirmDialog.show || r.confirmDialog.acceptLabel != "Sevk Et" || r.confirmDialog.thirdLabel != "Limana Gir" {
		t.Fatalf("kıyı modalı beklenen düğmeleri göstermedi: %+v", r.confirmDialog)
	}
	if r.confirmDialog.pendingAction.Kind != ActionDisembarkArmy {
		t.Fatalf("sevk aksiyonu = %q, want %q", r.confirmDialog.pendingAction.Kind, ActionDisembarkArmy)
	}
	if got := r.confirmDialog.thirdAction.TargetSettlementID; got != "target-port" {
		t.Fatalf("liman aksiyonunun settlement hedefi = %q, want target-port", got)
	}
}

func TestPostWarConquestChoiceUsesRequestedIcons(t *testing.T) {
	annex := decorateConfirmDialogButton(gameui.NewButton(0, 0, 100, 40, "İlhak Et"), "İlhak Et", "accept")
	if annex.Icon != gameui.IconLogin {
		t.Fatalf("ilhak butonu ikonu = %q, want %q", annex.Icon, gameui.IconLogin)
	}

	vassalize := decorateConfirmDialogButton(gameui.NewButton(0, 0, 100, 40, "Vassal Yap"), "Vassal Yap", "decline")
	if vassalize.Icon != gameui.IconExit {
		t.Fatalf("vassal butonu ikonu = %q, want %q", vassalize.Icon, gameui.IconExit)
	}
}

func TestEventLogCodexButtonUsesBookIcon(t *testing.T) {
	button := buildEventLogCodexButton()
	if button.Icon != gameui.IconBook {
		t.Fatalf("kodex butonu ikonu = %q, want %q", button.Icon, gameui.IconBook)
	}
}

func TestInfoPopupClickDismissesOnlyInsidePopup(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}

	tests := []struct {
		name  string
		timer int
		mx    float64
		my    float64
		want  bool
	}{
		{name: "popup ici", timer: 10, mx: 150, my: 120, want: true},
		{name: "harita disi", timer: 10, mx: 50, my: 120, want: false},
		{name: "sure bitmis", timer: 0, mx: 150, my: 120, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := infoPopupClickDismisses(tt.timer, popup, tt.mx, tt.my, true, false)
			if got != tt.want {
				t.Fatalf("popup kapatma = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfoPopupClickDismissesRequiresNewClick(t *testing.T) {
	popup := gameui.Rect{X: 100, Y: 100, W: 200, H: 80}
	if infoPopupClickDismisses(10, popup, 150, 120, true, true) {
		t.Fatal("basili tutulmus eski tiklama popup'i kapatti")
	}
}

func TestInfoPopupUsesTopRightSharedRect(t *testing.T) {
	originalWidth := ScreenWidth
	defer func() { ScreenWidth = originalWidth }()
	ScreenWidth = 1680

	r := &Renderer{combatLog: "Donanma limana konuşlandı."}
	popup := r.infoPopupRect()
	activeWarsButton := buildActiveWarsHUDButton()
	wantX := activeWarsButton.X + activeWarsButton.W + float64(infoPopupActiveWarsGap)
	if popup.X != wantX || popup.Y != float64(infoPopupTop) || popup.W != float64(infoPopupW) {
		t.Fatalf("bildirim recti = %+v, sağ üst recti bekleniyordu: x=%v y=%v w=%v", popup, wantX, infoPopupTop, infoPopupW)
	}
}

func TestInfoPopupDismissClearsMessageAndTimer(t *testing.T) {
	r := &Renderer{}
	r.ShowInfo("Kısa bilgi")
	if r.combatLogTimer != infoPopupAutoCloseFrames {
		t.Fatalf("bildirim süresi = %d, want %d kare", r.combatLogTimer, infoPopupAutoCloseFrames)
	}

	r.dismissInfoPopup()
	if r.combatLog != "" || r.combatLogTimer != 0 {
		t.Fatalf("bildirim kapatılmadı: message=%q timer=%d", r.combatLog, r.combatLogTimer)
	}
}

func TestMapRegionDoubleClickUsesTerrainParentIdentity(t *testing.T) {
	parentID := world.RegionID("parent")
	fragmentID := world.RegionID("area::forest::parent")
	r := &Renderer{gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
		parentID:   {ID: parentID, OwnerID: "target"},
		fragmentID: {ID: fragmentID, IsTerrainArea: true, ParentRegionID: parentID},
	}}}

	if r.mapRegionDoubleClicked(fragmentID) {
		t.Fatal("first terrain click unexpectedly counted as double click")
	}
	if !r.mapRegionDoubleClicked(parentID) {
		t.Fatal("terrain fragment and its parent were not treated as the same map click target")
	}
}

func TestToggleArmyDetailPanel(t *testing.T) {
	r := &Renderer{}

	r.toggleArmyDetailPanel()
	if !r.showArmyDetailPanel || !r.armyDetailPanelPreference {
		t.Fatal("marker sağ tıklaması ordu detay panelini açmadı")
	}

	r.toggleArmyDetailPanel()
	if r.showArmyDetailPanel || r.armyDetailPanelPreference {
		t.Fatal("aynı marker'a ikinci sağ tıklama açık ordu detay panelini kapatmadı")
	}
}

func TestDisbandButtonUsesTheSameGeometryForDrawAndHitTest(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"army": {ID: "army", OwnerID: "player", Units: []army.Unit{{TypeID: "infantry"}}},
		},
	}
	selected := map[int]bool{0: true}

	button, ok := buildDisbandArmyButton(gs, "army", selected)
	if !ok {
		t.Fatal("seçili birim varken Sil düğmesi oluşturulamadı")
	}
	if button.Icon != gameui.IconTrash {
		t.Fatalf("Sil düğmesi ikonu = %q, want %q", button.Icon, gameui.IconTrash)
	}
	mx := button.X + button.W/2
	my := button.Y + button.H/2
	if !DisbandButtonHitTest(mx, my, gs, "army", selected) {
		t.Fatal("Sil düğmesinin merkezi hit-test tarafından yakalanmadı")
	}
	if !ArmyPanelInteractiveHit(mx, my, gs, "army", selected) {
		t.Fatal("Sil düğmesinin merkezi panel cursor/input hit-test'i tarafından yakalanmadı")
	}
}

func TestMergeArmyButtonUsesForwardIconAndTargetCount(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"source": {ID: "source", OwnerID: "player", RegionID: "region", Units: []army.Unit{{TypeID: "infantry"}}},
			"target": {ID: "target", OwnerID: "player", RegionID: "region", Units: []army.Unit{
				{TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"}, {TypeID: "infantry"},
			}},
		},
	}

	button, ok := buildMergeArmyButton(gs, "source")
	if !ok {
		t.Fatal("birleştirme butonu oluşturulamadı")
	}
	if button.Label != "4" {
		t.Fatalf("birleştirme buton etiketi = %q, want %q", button.Label, "4")
	}
	if button.Icon != gameui.IconForward {
		t.Fatalf("birleştirme butonu ikonu = %q, want %q", button.Icon, gameui.IconForward)
	}
}

func TestSplitArmyButtonHiddenForFleetWithEmbarkedArmy(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:       "fleet",
				OwnerID:  "player",
				RegionID: "sea",
				IsNaval:  true,
				Units: []army.Unit{
					{TypeID: "transport"},
					{TypeID: "transport"},
				},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry"}},
			},
		},
	}

	if _, ok := buildSplitArmyButton(gs, "fleet"); ok {
		t.Fatal("ordu taşıyan filoda BÖL butonu gösterilmemeli")
	}
}

func TestBottomArmyActionUsesContextualSelection(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", OwnerID: "player"},
			"naval": {ID: "naval", OwnerID: "player", IsNaval: true},
		},
	}

	tests := []struct {
		name                               string
		selectedArmy                       army.ArmyID
		showArmyDetail                     bool
		showRecruit                        bool
		recruitEnabled                     bool
		wantLabel                          string
		wantEnabled, wantActive, wantNaval bool
	}{
		{name: "bölge", showRecruit: true, recruitEnabled: true, wantLabel: "Kışla", wantEnabled: true, wantActive: true},
		{name: "kara ordusu", selectedArmy: "land", wantLabel: "Ordu", wantEnabled: true},
		{name: "donanma", selectedArmy: "naval", showArmyDetail: true, wantLabel: "Donanma", wantEnabled: true, wantActive: true, wantNaval: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, enabled, active, naval, _ := bottomArmyAction(gs, tt.selectedArmy, tt.showArmyDetail, tt.showRecruit, tt.recruitEnabled)
			if label != tt.wantLabel || enabled != tt.wantEnabled || active != tt.wantActive || naval != tt.wantNaval {
				t.Fatalf("alt HUD aksiyonu = (%q, %v, %v, %v), want (%q, %v, %v, %v)", label, enabled, active, naval, tt.wantLabel, tt.wantEnabled, tt.wantActive, tt.wantNaval)
			}
			buttons := buildBottomActionButtons(label, enabled)
			if buttons[0].Label != tt.wantLabel || buttons[0].Enabled != tt.wantEnabled {
				t.Fatalf("alt HUD düğmesi = (%q, %v), want (%q, %v)", buttons[0].Label, buttons[0].Enabled, tt.wantLabel, tt.wantEnabled)
			}
			wantLabels := []string{tt.wantLabel, "Teknoloji", "Pazar", "Diplomasi", "Tur Bitir >"}
			for i, want := range wantLabels {
				if buttons[i].Label != want {
					t.Fatalf("alt HUD sıra[%d] = %q, want %q", i, buttons[i].Label, want)
				}
			}
		})
	}
}

func TestRecruitPanelStateForRegionFollowsPreference(t *testing.T) {
	ownedRegion := world.RegionID("owned")
	enemyRegion := world.RegionID("enemy")
	r := &Renderer{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Regions: map[world.RegionID]*world.Region{
				ownedRegion: {ID: ownedRegion, OwnerID: "player"},
				enemyRegion: {ID: enemyRegion, OwnerID: "enemy"},
			},
		},
		recruitPanelPreference: true,
	}

	if !r.recruitPanelStateForRegion(ownedRegion) {
		t.Fatal("açık Kışla tercihi oyuncunun uygun bölgesine taşınmadı")
	}
	if r.recruitPanelStateForRegion(enemyRegion) {
		t.Fatal("Kışla paneli düşman bölgesinde açık kabul edildi")
	}

	r.recruitPanelPreference = false
	if r.recruitPanelStateForRegion(ownedRegion) {
		t.Fatal("Kışla tercihi kapatıldıktan sonra panel açık kaldı")
	}
}

func TestRecruitPanelButtonDisabledWhenRegionHasNoMilitaryBuilding(t *testing.T) {
	region := &world.Region{ID: "minor", OwnerID: "player", IsMinorRegion: true}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions:         map[world.RegionID]*world.Region{region.ID: region},
		Factions:        map[faction.FactionID]*faction.Faction{"player": {ID: "player"}},
		UnitTypes: map[string]*army.UnitType{"militia": {
			ID: "militia", Category: army.CategoryInfantry,
			RequiredBuildings: []army.BuildingRequirement{{ID: "barracks", Level: 1}},
		}},
		UnitTypeOrder: []string{"militia"},
	}
	if RecruitPanelButtonEnabled(gs, region.ID) {
		t.Fatal("kışla ve limanı olmayan bölgede Kışla düğmesi aktif göründü")
	}

	region.Buildings = []string{"port"}
	gs.Regions["black_sea"] = &world.Region{ID: "black_sea", IsSea: true}
	region.Neighbors = []world.RegionID{"black_sea"}
	gs.UnitTypes["transport"] = &army.UnitType{
		ID: "transport", Category: army.CategoryNavalTrans,
		RequiredBuildings: []army.BuildingRequirement{{ID: "port", Level: 1}},
	}
	gs.UnitTypeOrder = []string{"militia", "transport"}
	if !RecruitPanelButtonEnabled(gs, region.ID) {
		t.Fatal("yalnızca limanı olan bölgede deniz üretim paneli açılabilir olmalıydı")
	}
}
