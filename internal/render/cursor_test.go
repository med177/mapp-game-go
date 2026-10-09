package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestDiplomacyCounterOfferButtonIsClickableAndShowsPointer(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 900

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ai":     {ID: "ai"},
			"player": {ID: "player"},
		},
		DiplomaticOffers: []state.DiplomaticOffer{{
			FromFactionID:      "ai",
			ToFactionID:        "player",
			Action:             string(diplomacy.ActionProposeTransfer),
			RequestedTransfers: []state.DiplomaticTransfer{{Kind: "resource", ID: "gold", Amount: 15}},
			OfferedTransfers:   []state.DiplomaticTransfer{{Kind: "resource", ID: "grain", Amount: 1}},
		}},
	}
	r := &Renderer{gs: gs, showDiplomacy: true}
	r.rebuildUILayers()
	counterBtn := buildDiplomacyOfferCounterButton()
	centerX := counterBtn.X + counterBtn.W/2
	centerY := counterBtn.Y + counterBtn.H/2

	if layer, ok := r.uiLayers.TopAt(centerX, centerY); !ok || layer.ID != uiLayerDiplomacyOffer {
		t.Fatalf("Karşı Teklif düğmesindeki üst UI katmanı = %+v, want %q", layer, uiLayerDiplomacyOffer)
	}
	if got := r.cursorShapeAt(centerX, centerY); got != ebiten.CursorShapePointer {
		t.Fatalf("Karşı Teklif düğmesi cursor şekli = %v, want pointer", got)
	}
	r.handleDiplomacyOfferInputState(0, gameui.InputState{
		MouseX:          centerX,
		MouseY:          centerY,
		LeftJustPressed: true,
	})
	if !r.negotiation.show || r.negotiation.target != "ai" || r.negotiation.counterIndex != 0 {
		t.Fatalf("Karşı Teklif tıklaması pazarlık ekranını açmadı: %+v", r.negotiation)
	}
	if r.showDiplomacy {
		t.Fatal("Karşı teklif açılırken diplomasi paneli arkada açık kaldı")
	}
	if _, ok := r.playerDiplomacyOfferIndex(); ok {
		t.Fatal("karşı teklif düzenlenirken gelen teklif penceresi yeniden gösteriliyor")
	}
	r.rebuildUILayers()
	layout := buildNegotiationLayout()
	panelCenterX := layout.modal.Panel.Rect.X + layout.modal.Panel.Rect.W/2
	panelCenterY := layout.modal.Panel.Rect.Y + layout.modal.Panel.Rect.H/2
	if layer, ok := r.uiLayers.TopAt(panelCenterX, panelCenterY); !ok || layer.ID != uiLayerDiplomacyNegotiation {
		t.Fatalf("Pazarlık panelindeki üst UI katmanı = %+v, want %q", layer, uiLayerDiplomacyNegotiation)
	}
	if got := r.negotiation.requested.items; len(got) != 1 || got[0].ID != "grain" || got[0].Amount != 1 {
		t.Fatalf("Karşı teklif talep kalemleri = %+v, want offered grain transfer", got)
	}
	if got := r.negotiation.offered.items; len(got) != 1 || got[0].ID != "gold" || got[0].Amount != 15 {
		t.Fatalf("Karşı teklif verilen kalemleri = %+v, want requested gold transfer", got)
	}
}

func TestCursorShapeAtUsesSharedEditModeUILayer(t *testing.T) {
	originalHeight := ScreenHeight
	defer func() { ScreenHeight = originalHeight }()
	ScreenHeight = 900

	r := &Renderer{
		gs:               &state.GameState{Phase: state.PhaseEditMode},
		editInspectorTab: editInspectorData,
	}
	r.rebuildUILayers()

	tab := buildEditInspectorTabButton(editInspectorData, "")
	if got := r.cursorShapeAt(tab.X+tab.W/2, tab.Y+tab.H/2); got != ebiten.CursorShapePointer {
		t.Fatalf("Edit Mode sekmesi cursor şekli = %v, want pointer", got)
	}

	x, y, _, _ := editInspectorRect()
	if got := r.cursorShapeAt(float64(x)+10, float64(y)+60); got != ebiten.CursorShapeDefault {
		t.Fatalf("Edit Mode inspector boş alanı cursor şekli = %v, want default", got)
	}

	if got := r.cursorShapeAt(float64(ScreenWidth-10), float64(ScreenHeight/2)); got != ebiten.CursorShapeDefault {
		t.Fatalf("Edit Mode harita alanı cursor şekli = %v, want default", got)
	}
}

func TestDiplomacyNegotiationCursorOnlyTargetsInteractiveControls(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 900

	targetRegion := world.RegionID("target-region")
	playerRegion := world.RegionID("player-region")
	gs := &state.GameState{
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			targetRegion: {ID: targetRegion, OwnerID: "target", NameTR: "Hedef Bölgesi"},
			playerRegion: {ID: playerRegion, OwnerID: "player", NameTR: "Oyuncu Bölgesi"},
		},
	}
	r := &Renderer{
		gs:            gs,
		showDiplomacy: true,
		negotiation: negotiationPanelState{
			show:   true,
			target: "target",
			requested: negotiationSideState{
				kind:   negotiationKindRegion,
				amount: 1,
			},
			offered: negotiationSideState{
				kind:   negotiationKindRegion,
				amount: 1,
			},
		},
	}
	layout := buildNegotiationLayout()

	assertPointer := func(name string, x, y float64, want ebiten.CursorShapeType) {
		t.Helper()
		if got := r.cursorShapeAt(x, y); got != want {
			t.Errorf("%s cursor şekli = %v, want %v", name, got, want)
		}
	}
	assertPointer("Tür düğmesi", layout.left.kind.X+layout.left.kind.W/2, layout.left.kind.Y+layout.left.kind.H/2, ebiten.CursorShapePointer)
	assertPointer("Bölge seçim düğmesi", layout.left.option.X+layout.left.option.W/2, layout.left.option.Y+layout.left.option.H/2, ebiten.CursorShapePointer)
	assertPointer("Bölge türündeki devre dışı miktar düğmesi", layout.left.minus.X+layout.left.minus.W/2, layout.left.minus.Y+layout.left.minus.H/2, ebiten.CursorShapeDefault)
	assertPointer("Boş gönder düğmesi", layout.submit.X+layout.submit.W/2, layout.submit.Y+layout.submit.H/2, ebiten.CursorShapeDefault)
	assertPointer("Kapat düğmesi", layout.cancel.X+layout.cancel.W/2, layout.cancel.Y+layout.cancel.H/2, ebiten.CursorShapePointer)
	assertPointer("Panel boş alanı", layout.left.panel.X+layout.left.panel.W/2, layout.left.panel.Y+layout.left.panel.H/2, ebiten.CursorShapeDefault)

	dropdown := negotiationRegionDropdown(gs, "target", &r.negotiation.requested, layout.left.panel)
	if dropdown == nil {
		t.Fatal("bölge dropdown'ı oluşturulmadı")
	}
	dropdown.Toggle()
	listRect := negotiationRegionDropdownRect(layout.left.panel)
	assertPointer("Açık dropdown seçeneği", listRect.X+12, listRect.Y+34, ebiten.CursorShapePointer)

	r.negotiation.requested.items = []state.DiplomaticTransfer{{Kind: "region", ID: string(targetRegion), Amount: 1}}
	selectedItems := negotiationSelectedItemsListRect(layout.left.panel)
	assertPointer("Seçili kalem satırı", selectedItems.X+8, selectedItems.Y+10, ebiten.CursorShapePointer)
}

func TestNegotiationSelectedItemsListIsInsetFromFrame(t *testing.T) {
	panel := buildNegotiationLayout().left.panel
	frame := negotiationSelectedItemsFrameRect(panel)
	list := negotiationSelectedItemsListRect(panel)
	if list.X <= frame.X || list.Y <= frame.Y || list.X+list.W >= frame.X+frame.W || list.Y+list.H >= frame.Y+frame.H {
		t.Fatalf("seçili kalem listesi çerçevenin içinde değil: frame=%+v list=%+v", frame, list)
	}
}

func TestInGameHoveringIgnoresHiddenRecruitPanelCards(t *testing.T) {
	rid := world.RegionID("owned")
	gs := &state.GameState{
		Phase:           state.PhasePlayerTurn,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			rid: {ID: rid, OwnerID: "player"},
		},
		UnitTypes: map[string]*army.UnitType{
			"militia":  {ID: "militia"},
			"infantry": {ID: "infantry"},
			"cavalry":  {ID: "cavalry"},
		},
		UnitTypeOrder: []string{"militia", "infantry", "cavalry"},
	}
	r := &Renderer{gs: gs, SelectedRegion: rid, mapMode: MapModeNormal}
	buttons := buildRecruitUnitCardButtons(gs, rid)
	if len(buttons) < 3 {
		t.Fatalf("test kartları oluşturulamadı: %d", len(buttons))
	}
	mx := buttons[2].X + buttons[2].W/2
	my := buttons[2].Y + buttons[2].H/2

	if !RecruitPanelInteractiveHit(mx, my, gs, rid) {
		t.Fatal("kart merkezi görünür Kışla paneli için hit-test vermedi")
	}
	if r.inGameHovering(mx, my) {
		t.Fatal("Kışla paneli kapalıyken görünmeyen kart cursor hover alanı oluşturdu")
	}

	r.showRecruitPanel = true
	if !r.inGameHovering(mx, my) {
		t.Fatal("Kışla paneli açıkken görünen kart cursor hover alanı oluşturmadı")
	}
}
