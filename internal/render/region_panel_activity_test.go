package render

import (
	"testing"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestRegionPanelEventTabScrollAndSharedContentArea(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("test_region")
	gs := &state.GameState{
		DevelopmentMode: true,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, NameTR: "Test", OwnerID: "player", Neighbors: []world.RegionID{
				"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10",
				"n11", "n12", "n13", "n14", "n15", "n16", "n17", "n18", "n19", "n20",
			}},
		},
		BuildingTypes: map[string]*city.Building{
			"market":   {ID: "market", NameTR: "Pazar"},
			"farm":     {ID: "farm", NameTR: "Çiftlik"},
			"barracks": {ID: "barracks", NameTR: "Kışla"},
			"walls":    {ID: "walls", NameTR: "Surlar"},
			"temple":   {ID: "temple", NameTR: "İbadet Yeri"},
			"port":     {ID: "port", NameTR: "Liman"},
		},
	}
	for _, neighborID := range gs.Regions[regionID].Neighbors {
		gs.Regions[neighborID] = &world.Region{ID: neighborID, NameTR: string(neighborID)}
	}
	for i := 0; i < 3; i++ {
		gs.ActiveRegionEvents = append(gs.ActiveRegionEvents, state.RegionEventStatus{RegionID: regionID, EventID: "event", LabelTR: "Olay", TurnsLeft: 3})
	}

	start := buildingGridStartY(gs, gs.Regions[regionID], false)
	end := buildingGridEndY(gs, gs.Regions[regionID], start)
	viewport := regionPanelActivityViewport(gs, gs.Regions[regionID])
	if viewport.Y != float64(start) {
		t.Fatalf("olaylar viewport'u bina kartlarıyla aynı içerik başlangıcında olmalı: activityY=%.1f startY=%.1f", viewport.Y, start)
	}
	if viewport.Y+viewport.H > float64(end)+0.01 {
		t.Fatalf("olaylar viewport'u sekme içerik alanını aşmamalı: activityBottom=%.1f buildingEndY=%.1f", viewport.Y+viewport.H, end)
	}
	maxScroll := clampRegionPanelScroll(gs, regionID, 1<<20)
	if maxScroll <= 0 {
		t.Fatal("uzun olay/komsu listesi icin pozitif scroll araligi bekleniyordu")
	}
	if !regionPanelActivityHit(viewport.X+4, viewport.Y+4, gs, regionID) {
		t.Fatal("aktivite viewport'u mouse wheel hit-test'i kabul etmeli")
	}
	neighborToggleY := viewport.Y + 8 + 17 + 3*28 + 6
	if !regionActivityNeighborToggleHit(viewport.X+4, neighborToggleY+4, gs, gs.Regions[regionID], 0) {
		t.Fatal("olaylar sekmesindeki komşu daralt/genişlet başlığı tıklanabilir olmalı")
	}
	if expanded, collapsed := regionActivityNeighborContentHeightForExpanded(gs, gs.Regions[regionID], true), regionActivityNeighborContentHeightForExpanded(gs, gs.Regions[regionID], false); expanded <= collapsed {
		t.Fatalf("komşu görünümü daraltılınca içerik yüksekliği azalmalı: expanded=%.1f collapsed=%.1f", expanded, collapsed)
	}
}

func TestRegionPanelTabsAndActiveEventRowHit(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("tab_region")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID:   {ID: regionID, NameTR: "Sekme", OwnerID: "player", Neighbors: []world.RegionID{"neighbor"}},
			"neighbor": {ID: "neighbor", NameTR: "Komşu"},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar"},
		},
		ActiveRegionEvents: []state.RegionEventStatus{
			{RegionID: regionID, EventID: "evt", LabelTR: "Veba", TurnsLeft: 3, Type: "plague"},
		},
	}

	buildingTab, eventTab := regionPanelTabRects(gs, gs.Regions[regionID])
	if tab, ok := regionPanelTabHit(buildingTab.X+2, buildingTab.Y+2, gs, regionID); !ok || tab != regionPanelTabBuildings {
		t.Fatalf("binalar sekmesi hit-test'i başarısız: tab=%d ok=%t", tab, ok)
	}
	if tab, ok := regionPanelTabHit(eventTab.X+2, eventTab.Y+2, gs, regionID); !ok || tab != regionPanelTabEvents {
		t.Fatalf("olaylar sekmesi hit-test'i başarısız: tab=%d ok=%t", tab, ok)
	}

	viewport := regionPanelActivityViewport(gs, gs.Regions[regionID])
	idx, ok := regionActiveEventPanelHit(viewport.X+8, viewport.Y+26, gs, gs.Regions[regionID], 0)
	if !ok || idx != 0 {
		t.Fatalf("aktif olay satırı popup için tıklanabilir olmalı: idx=%d ok=%t", idx, ok)
	}
}

func TestRegionPanelActivityAndDiplomacyUseVisiblePanelGeometry(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("enemy_region")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID:   {ID: regionID, NameTR: "Düşman Bölgesi", OwnerID: "enemy", Neighbors: []world.RegionID{"neighbor"}},
			"neighbor": {ID: "neighbor", NameTR: "Komşu"},
		},
		BuildingTypes: map[string]*city.Building{
			"market": {ID: "market", NameTR: "Pazar"},
		},
	}

	if !regionActivityNeighborVisible(gs, gs.Regions[regionID]) {
		t.Fatal("komşu listesi geliştirme modu kapalıyken de görünür olmalı")
	}
	if got := regionActivityNeighborContentHeight(gs, gs.Regions[regionID]); got <= 8+devNeighborTitleHeight {
		t.Fatalf("komşu satırı içerik yüksekliğine eklenmedi: %.1f", got)
	}

	start := buildingGridStartY(gs, gs.Regions[regionID], false)
	end := buildingGridEndY(gs, gs.Regions[regionID], start)
	barY := float64(end) + 5
	btn := buildRegionDiplomacyButtons(gs, "enemy", infoPanelX()+float32(panelPad), float32(barY), infoPanelW-float32(panelPad*2), regionPanelActionBarHeight)
	expectedY := float64(barY) + (float64(regionPanelActionBarHeight)-btn.H)/2
	if btn.Y < expectedY-0.01 || btn.Y > expectedY+0.01 {
		t.Fatalf("diplomasi düğmesi aksiyon bandında ortalanmadı: %+v", btn)
	}
	if !regionDiplomacyButtonHit(btn.X+btn.W/2, btn.Y+btn.H/2, gs, regionID) {
		t.Fatal("çizilen diplomasi düğmesinin merkezi tıklanabilir olmalı")
	}
}

func TestOwnRegionGrainAidButtonUsesRegionActionBarGeometry(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("home")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: state.GrainAidCost},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "player", Satisfaction: 50},
		},
	}

	start := buildingGridStartY(gs, gs.Regions[regionID], false)
	end := buildingGridEndY(gs, gs.Regions[regionID], start)
	barY := end + 5
	btn := buildRegionGrainAidButton(gs, regionID, infoPanelX()+float32(panelPad), float32(barY), infoPanelW-float32(panelPad*2), regionPanelActionBarHeight)
	expectedY := float64(barY) + (float64(regionPanelActionBarHeight)-btn.H)/2
	if btn.Y < expectedY-0.01 || btn.Y > expectedY+0.01 {
		t.Fatalf("tahıl yardım düğmesi aksiyon bandında ortalanmadı: %+v", btn)
	}
	if !regionGrainAidButtonHit(btn.X+btn.W/2, btn.Y+btn.H/2, gs, regionID) {
		t.Fatal("çizilen tahıl yardım düğmesinin merkezi tıklanabilir olmalı")
	}
}

func TestOwnRegionActionButtonsStaySeparatedForLiberation(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth = 1280
	ScreenHeight = 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	regionID := world.RegionID("former_capital")
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":    {ID: "player", Grain: state.GrainAidCost},
			"successor": {ID: "successor", IsEliminated: true},
		},
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "player", SuccessorFactionID: "successor", Satisfaction: 50},
		},
	}

	barY := float32(regionPanelActionBarY(gs, gs.Regions[regionID], regionPanelTabBuildings))
	barX := infoPanelX() + float32(panelPad)
	barW := infoPanelW - float32(panelPad*2)
	grainBtn := buildRegionGrainAidButton(gs, regionID, barX, barY, barW, regionPanelActionBarHeight)
	liberateBtn := buildRegionLiberateButton(barX, barY, barW, regionPanelActionBarHeight)

	if grainBtn.X+grainBtn.W > liberateBtn.X {
		t.Fatalf("aksiyon düğmeleri üst üste biniyor: grain=%+v liberate=%+v", grainBtn, liberateBtn)
	}
	if !regionGrainAidButtonHitForTab(grainBtn.X+grainBtn.W/2, grainBtn.Y+grainBtn.H/2, gs, regionID, regionPanelTabBuildings) {
		t.Fatal("tahıl yardım düğmesinin merkezi tıklanabilir olmalı")
	}
	if !regionLiberateButtonHitForTab(liberateBtn.X+liberateBtn.W/2, liberateBtn.Y+liberateBtn.H/2, gs, regionID, regionPanelTabBuildings) {
		t.Fatal("özgürleştir düğmesinin merkezi tıklanabilir olmalı")
	}
}

func TestRegionActivityContentUsesViewportScreenCoordinates(t *testing.T) {
	viewport := gameui.Rect{X: 12, Y: 640, W: 281, H: 120}
	const scroll = 28.0

	contentX, contentY := regionActivityContentOrigin(viewport, scroll)
	contentCenterX := regionActivityContentCenterX(viewport)
	if contentX != 24 || contentY != 620 || contentCenterX != 152.5 {
		t.Fatalf("aktivite içeriği viewport ekran koordinatına taşınmadı: x=%.1f y=%.1f center=%.1f", contentX, contentY, contentCenterX)
	}
}
