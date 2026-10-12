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
