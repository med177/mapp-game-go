package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestNegotiationRegionOptionsExcludeAddedItemsAndRestoreRemovedItems(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"region_a": {ID: "region_a", OwnerID: "player", NameTR: "Bölge A"},
			"region_b": {ID: "region_b", OwnerID: "player", NameTR: "Bölge B"},
			"region_c": {ID: "region_c", OwnerID: "player", NameTR: "Bölge C"},
		},
	}
	side := &negotiationSideState{
		items: []state.DiplomaticTransfer{{Kind: "region", ID: "region_b", Amount: 1}},
	}

	options := negotiationOptionsForSide(gs, faction.FactionID("player"), side, negotiationKindRegion)
	if len(options) != 2 {
		t.Fatalf("eklenmiş bölge çıkarılınca seçenek sayısı = %d, want 2", len(options))
	}
	for _, option := range options {
		if option.id == "region_b" {
			t.Fatal("eklenmiş bölge dropdown seçeneklerinde kaldı")
		}
	}

	side.items = nil
	options = negotiationOptionsForSide(gs, faction.FactionID("player"), side, negotiationKindRegion)
	if len(options) != 3 {
		t.Fatalf("kalem kaldırılınca seçenek sayısı = %d, want 3", len(options))
	}
	for _, option := range options {
		if option.id == "region_b" {
			return
		}
	}
	t.Fatal("kaldırılan bölge dropdown seçeneklerine geri dönmedi")
}

func TestNegotiationRegionControlsDisableWhenAllRegionsAreAdded(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"region_a": {ID: "region_a", OwnerID: "player", NameTR: "Bölge A"},
		},
	}
	side := &negotiationSideState{
		kind:  negotiationKindRegion,
		items: []state.DiplomaticTransfer{{Kind: "region", ID: "region_a", Amount: 1}},
	}
	layout := negotiationSideLayout{
		option: gameui.NewButton(0, 0, 10, 10, "Seçim"),
		add:    gameui.NewButton(0, 0, 10, 10, "Ekle"),
	}

	configureNegotiationSideControls(gs, faction.FactionID("player"), side, &layout)
	if layout.option.Enabled || layout.add.Enabled {
		t.Fatalf("bütün bölgeler ekliyken kontroller etkin: option=%t add=%t", layout.option.Enabled, layout.add.Enabled)
	}
}

func TestNegotiationArmyDropdownSelectsUnitTypeAndDefaultAmount(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"army_a": {
				ID:      "army_a",
				OwnerID: "player",
				Units: []army.Unit{
					{TypeID: "archer"},
					{TypeID: "swordsman"},
					{TypeID: "swordsman"},
				},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"archer":    {ID: "archer", NameTR: "Okçu"},
			"swordsman": {ID: "swordsman", NameTR: "Kılıçlı"},
		},
	}
	side := &negotiationSideState{kind: negotiationKindArmy, amount: 1}
	panel := gameui.Rect{X: 10, Y: 20, W: 400, H: 500}
	dropdown := negotiationOptionDropdown(gs, "player", side, panel)
	if dropdown == nil {
		t.Fatal("asker türleri dropdown'ı oluşturulmadı")
	}
	if got := dropdown.Options(); len(got) != 2 || got[0] != "Okçu" || got[1] != "Kılıçlı" {
		t.Fatalf("asker türü seçenekleri = %v, want [Okçu Kılıçlı]", got)
	}

	dropdown.Toggle()
	rect := negotiationOptionDropdownRect(panel)
	input := gameui.InputState{
		MouseX:          rect.X + 12,
		MouseY:          rect.Y + 28 + 28 + 1,
		LeftJustPressed: true,
	}
	if !handleNegotiationOptionDropdownInput(input, dropdown, gs, "player", side) {
		t.Fatal("asker türü seçimi dropdown input'u tarafından işlenmedi")
	}
	if side.option != 1 || side.amount != 2 {
		t.Fatalf("asker seçimi/miktarı = %d/%d, want 1/2", side.option, side.amount)
	}
	if dropdown.IsOpen() {
		t.Fatal("asker türü seçildikten sonra dropdown açık kaldı")
	}
}

func TestNegotiationAmountShiftChangesByOneHundred(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 1000},
		},
	}
	side := &negotiationSideState{kind: negotiationKindResource, amount: 500}

	adjustNegotiationAmount(gs, "player", side, 1, false)
	if side.amount != 510 {
		t.Fatalf("normal kaynak artışı = %d, want 510", side.amount)
	}

	adjustNegotiationAmount(gs, "player", side, 1, true)
	if side.amount != 610 {
		t.Fatalf("Shift ile kaynak artışı = %d, want 610", side.amount)
	}

	adjustNegotiationAmount(gs, "player", side, -1, true)
	if side.amount != 510 {
		t.Fatalf("Shift ile kaynak azaltımı = %d, want 510", side.amount)
	}
}

func TestNegotiationArmyDropdownShowsCommanderBoundTypesAsDisabled(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"army_free": {
				ID:      "army_free",
				OwnerID: "player",
				Units:   []army.Unit{{TypeID: "archer"}},
			},
			"army_commanded": {
				ID:        "army_commanded",
				OwnerID:   "player",
				Commander: &army.Commander{ID: "commander"},
				Units:     []army.Unit{{TypeID: "cavalry"}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"archer":  {ID: "archer", NameTR: "Okçu"},
			"cavalry": {ID: "cavalry", NameTR: "Akıncılar"},
		},
	}
	side := &negotiationSideState{kind: negotiationKindArmy, amount: 1}
	options := negotiationOptionsForSide(gs, "player", side, negotiationKindArmy)
	if len(options) != 2 {
		t.Fatalf("asker türü seçeneği sayısı = %d, want 2", len(options))
	}
	if options[0].id != "archer" || options[0].disabled {
		t.Fatalf("aktarılabilir asker seçeneği hatalı: %+v", options[0])
	}
	if options[1].id != "cavalry" || !options[1].disabled || options[1].label != "Akıncılar (Komutana Bağlı)" {
		t.Fatalf("komutana bağlı asker seçeneği pasif açıklamayla gösterilmedi: %+v", options[1])
	}

	panel := gameui.Rect{X: 10, Y: 20, W: 400, H: 500}
	dropdown := negotiationOptionDropdown(gs, "player", side, panel)
	if dropdown == nil || !dropdown.OptionDisabled(1) {
		t.Fatal("komutana bağlı asker türü dropdown'da pasif değil")
	}
	dropdown.Toggle()
	rect := negotiationOptionDropdownRect(panel)
	input := gameui.InputState{
		MouseX:          rect.X + 12,
		MouseY:          rect.Y + 28 + 28 + 1,
		LeftJustPressed: true,
	}
	if !handleNegotiationOptionDropdownInput(input, dropdown, gs, "player", side) {
		t.Fatal("pasif satır tıklaması dropdown tarafından tüketilmedi")
	}
	if side.option != 0 || len(side.items) != 0 || !dropdown.IsOpen() {
		t.Fatalf("pasif asker tipi seçilebilir hale geldi: option=%d items=%v open=%t", side.option, side.items, dropdown.IsOpen())
	}

	side.option = 1
	addNegotiationItem(gs, "player", side)
	if len(side.items) != 0 {
		t.Fatalf("komutana bağlı asker kalemi pazarlığa eklendi: %+v", side.items)
	}
}

func TestNegotiationArmyDropdownRemainsOpenableWhenOnlyCommanderBoundTypesExist(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"army_commanded": {
				ID:        "army_commanded",
				OwnerID:   "player",
				Commander: &army.Commander{ID: "commander"},
				Units:     []army.Unit{{TypeID: "cavalry"}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"cavalry": {ID: "cavalry", NameTR: "Akıncılar"},
		},
	}
	side := &negotiationSideState{kind: negotiationKindArmy, amount: 1}
	layout := negotiationSideLayout{
		option: gameui.Button{Enabled: true},
		add:    gameui.Button{Enabled: true},
	}

	configureNegotiationSideControls(gs, "player", side, &layout)
	if !layout.option.Enabled || layout.add.Enabled {
		t.Fatalf("yalnız komutanlı birlik varken seçim/ekleme durumu yanlış: option=%t add=%t", layout.option.Enabled, layout.add.Enabled)
	}
	if dropdown := negotiationOptionDropdown(gs, "player", side, gameui.Rect{X: 10, Y: 20, W: 400, H: 500}); dropdown == nil || len(dropdown.Options()) != 1 {
		t.Fatal("yalnız komutanlı birlikler için açıklama dropdown'da gösterilmedi")
	}
}

func TestNegotiationCaptivePickerSeparatesRequestedAndOfferedCommanders(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"player_hold": {ID: "player_hold", OwnerID: "player", Settlements: []world.Settlement{{ID: "player_hold_settlement"}}},
			"target_hold": {ID: "target_hold", OwnerID: "target", Settlements: []world.Settlement{{ID: "target_hold_settlement"}}},
		},
		Commanders: map[string]*army.Commander{
			"target_commander": {ID: "target_commander", OwnerID: "target", Name: "Hedef Komutan", CaptiveAtSettlementID: "player_hold_settlement"},
			"player_commander": {ID: "player_commander", OwnerID: "player", Name: "Oyuncu Komutan", CaptiveAtSettlementID: "target_hold_settlement"},
			"other_commander":  {ID: "other_commander", OwnerID: "other", Name: "Üçüncü Taraf", CaptiveAtSettlementID: "player_hold_settlement"},
		},
	}

	offered := &negotiationSideState{kind: negotiationKindCaptive, counterparty: "target", amount: 1}
	options := negotiationOptionsForSide(gs, "player", offered, negotiationKindCaptive)
	if len(options) != 1 || options[0].id != "target_commander" || options[0].label != "Hedef Komutan" {
		t.Fatalf("verilecek tutsak seçenekleri = %+v, yalnız hedef komutanı bekleniyordu", options)
	}
	addNegotiationItem(gs, "player", offered)
	if len(offered.items) != 1 || offered.items[0].Kind != "captive" || offered.items[0].ID != "target_commander" || offered.items[0].Amount != 1 {
		t.Fatalf("verilen tutsak kalemi yanlış eklendi: %+v", offered.items)
	}
	if options = negotiationOptionsForSide(gs, "player", offered, negotiationKindCaptive); len(options) != 0 {
		t.Fatalf("eklenmiş tutsak tekrar seçilebilir: %+v", options)
	}

	requested := &negotiationSideState{kind: negotiationKindCaptive, counterparty: "player", amount: 1}
	options = negotiationOptionsForSide(gs, "target", requested, negotiationKindCaptive)
	if len(options) != 1 || options[0].id != "player_commander" {
		t.Fatalf("talep edilebilir tutsak seçenekleri = %+v, oyuncu komutanı bekleniyordu", options)
	}

	layout := negotiationSideLayout{minus: gameui.Button{Enabled: true}, plus: gameui.Button{Enabled: true}, add: gameui.Button{Enabled: true}}
	configureNegotiationSideControls(gs, "target", requested, &layout)
	if !layout.add.Enabled || layout.minus.Enabled || layout.plus.Enabled {
		t.Fatalf("tutsak tekil kalem kontrolleri yanlış: add=%t minus=%t plus=%t", layout.add.Enabled, layout.minus.Enabled, layout.plus.Enabled)
	}
}

func TestNegotiationMapNavigationUsesMiddleDragAndLeftClickDoesNotMoveCamera(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 900

	r := &Renderer{
		gs:       &state.GameState{Phase: state.PhaseEditMode},
		camX:     float64(WorldW) / 2,
		camY:     float64(WorldH) / 2,
		camScale: minCameraScale() * 2,
	}
	panel := gameui.Rect{X: 800, Y: 100, W: 400, H: 700}
	initialX, initialY, initialScale := r.camX, r.camY, r.camScale

	r.handleNegotiationMapDrag(100, 200, true, panel)
	r.handleNegotiationMapDrag(160, 230, true, panel)
	if r.camX == initialX && r.camY == initialY {
		t.Fatal("orta tuşla harita sürükleme kamerayı hareket ettirmedi")
	}
	r.handleNegotiationMapDrag(160, 230, false, panel)

	beforePanelDragX, beforePanelDragY := r.camX, r.camY
	r.handleNegotiationMapDrag(900, 300, true, panel)
	r.handleNegotiationMapDrag(950, 350, true, panel)
	if r.camX != beforePanelDragX || r.camY != beforePanelDragY {
		t.Fatal("pazarlık paneli üzerinde orta tuşla sürükleme kamerayı hareket ettirdi")
	}
	r.handleNegotiationMapDrag(950, 350, false, panel)
	r.handleNegotiationMapNavigation(gameui.InputState{MouseX: 300, MouseY: 300, LeftPressed: true, LeftJustPressed: true}, panel)
	if r.camX != beforePanelDragX || r.camY != beforePanelDragY {
		t.Fatal("haritaya sol tıklama kamerayı hareket ettirdi")
	}
	r.handleNegotiationMapNavigation(gameui.InputState{MouseX: 300, MouseY: 300, WheelY: 1}, panel)
	if r.camScale <= initialScale {
		t.Fatalf("harita alanında zoom ölçeği artmadı: başlangıç=%f güncel=%f", initialScale, r.camScale)
	}
	zoomedScale := r.camScale
	r.handleNegotiationMapNavigation(gameui.InputState{MouseX: 900, MouseY: 300, WheelY: 1}, panel)
	if r.camScale != zoomedScale {
		t.Fatal("pazarlık paneli üzerindeki tekerlek kamera zoom'unu değiştirdi")
	}
}

func TestNegotiationMapHighlightFollowsRequestedOrOfferedSide(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"target_region": {ID: "target_region", OwnerID: "target", NameTR: "Hedef Bölge"},
			"player_region": {ID: "player_region", OwnerID: "player", NameTR: "Oyuncu Bölgesi"},
		},
	}
	r := &Renderer{
		gs: gs,
		negotiation: negotiationPanelState{
			show:   true,
			target: "target",
			requested: negotiationSideState{
				kind: negotiationKindRegion, amount: 1,
			},
			offered: negotiationSideState{
				kind: negotiationKindRegion, amount: 1,
			},
		},
	}
	if got := r.negotiationMapHighlightRegion(); got != "target_region" {
		t.Fatalf("varsayılan harita vurgusu = %q, hedef bölgesi bekleniyordu", got)
	}
	r.negotiation.mapFocusSide = 1
	if got := r.negotiationMapHighlightRegion(); got != "player_region" {
		t.Fatalf("verilenler tarafındaki harita vurgusu = %q, oyuncu bölgesi bekleniyordu", got)
	}
}

func TestNegotiationMapRegionSelectionTogglesForCorrectOwner(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"target_region": {ID: "target_region", OwnerID: "target"},
			"player_region": {ID: "player_region", OwnerID: "player"},
			"other_region":  {ID: "other_region", OwnerID: "other"},
			"sea_region":    {ID: "sea_region", OwnerID: "target", IsSea: true},
		},
	}
	requested := &negotiationSideState{kind: negotiationKindResource}
	offered := &negotiationSideState{kind: negotiationKindArmy}

	if !toggleNegotiationRegion(gs, "target", requested, "target_region") {
		t.Fatal("karşı tarafa ait bölge talebe eklenmedi")
	}
	if requested.kind != negotiationKindRegion || len(requested.items) != 1 || requested.items[0].ID != "target_region" {
		t.Fatalf("talep edilen bölge kalemi yanlış: kind=%d items=%+v", requested.kind, requested.items)
	}
	if !toggleNegotiationRegion(gs, "target", requested, "target_region") || len(requested.items) != 0 {
		t.Fatalf("bölgeye tekrar tıklayınca talepten kaldırılmadı: %+v", requested.items)
	}
	if !toggleNegotiationRegion(gs, "player", offered, "player_region") || len(offered.items) != 1 || offered.items[0].ID != "player_region" {
		t.Fatalf("oyuncuya ait bölge teklife eklenmedi: %+v", offered.items)
	}
	if toggleNegotiationRegion(gs, "target", requested, "other_region") || toggleNegotiationRegion(gs, "target", requested, "sea_region") {
		t.Fatal("başka devletin veya denizin bölgesi pazarlığa seçilebildi")
	}
}

func TestKeepDiplomacyNegotiationOpenPreservesSelectedTransfers(t *testing.T) {
	selected := []state.DiplomaticTransfer{{Kind: "region", ID: "target_region", Amount: 1}}
	r := &Renderer{negotiation: negotiationPanelState{
		show:         true,
		target:       "target",
		counterIndex: 2,
		requested:    negotiationSideState{items: selected},
	}}

	r.KeepDiplomacyNegotiationOpen("Pazarlık reddedildi.")

	if !r.negotiation.show || r.negotiation.counterIndex != -1 {
		t.Fatalf("ret sonrasında pazarlık state'i açık değil veya eski index korunmuş: %+v", r.negotiation)
	}
	if len(r.negotiation.requested.items) != 1 || r.negotiation.requested.items[0].ID != "target_region" {
		t.Fatalf("ret sonrasında seçili bölge kayboldu: %+v", r.negotiation.requested.items)
	}
	if r.negotiation.statusMessage != "Pazarlık reddedildi." {
		t.Fatalf("ret mesajı panel state'ine yazılmadı: %q", r.negotiation.statusMessage)
	}
}

func TestOpenDiplomacyNegotiationClosesDiplomacyPanel(t *testing.T) {
	r := &Renderer{
		gs:                     &state.GameState{PlayerFactionID: "player"},
		showDiplomacy:          true,
		diplomacyTargetFaction: "target",
	}

	r.openDiplomacyNegotiation("target")

	if !r.negotiation.show || r.negotiation.target != "target" {
		t.Fatalf("pazarlık hedefi açılmadı: %+v", r.negotiation)
	}
	if r.showDiplomacy || r.diplomacyTargetFaction != "" {
		t.Fatalf("pazarlık açılırken Diplomasi paneli açık kaldı: show=%t target=%q", r.showDiplomacy, r.diplomacyTargetFaction)
	}
	r.rebuildUILayers()
	layout := buildNegotiationLayout()
	centerX := layout.modal.Panel.Rect.X + layout.modal.Panel.Rect.W/2
	centerY := layout.modal.Panel.Rect.Y + layout.modal.Panel.Rect.H/2
	if layer, ok := r.uiLayers.TopAt(centerX, centerY); !ok || layer.ID != uiLayerDiplomacyNegotiation {
		t.Fatalf("Diplomasi kapanınca en üst katman pazarlık değil: %+v", layer)
	}
}
