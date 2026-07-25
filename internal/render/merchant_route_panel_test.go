package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestMerchantRouteButtonOnlyTargetsPlayerMerchantFleet(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
			"warship":       {ID: "warship", Category: army.CategoryNavalWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"merchant": {ID: "merchant", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "merchant_ship"}}},
			"war":      {ID: "war", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
			"enemy":    {ID: "enemy", OwnerID: "enemy", IsNaval: true, Units: []army.Unit{{TypeID: "merchant_ship"}}},
		},
	}

	button := merchantRouteAssignmentButtonRect(armyPanelGeometry())
	if !merchantRouteButtonHit(button.X+button.W/2, button.Y+button.H/2, gs, "merchant") {
		t.Fatal("oyuncu merchant filosunda rota butonu hit olmalıydı")
	}
	if merchantRouteButtonHit(button.X+button.W/2, button.Y+button.H/2, gs, "war") {
		t.Fatal("savaş filosunda merchant rota butonu görünmemeliydi")
	}
	if merchantRouteButtonHit(button.X+button.W/2, button.Y+button.H/2, gs, "enemy") {
		t.Fatal("düşman merchant filosunda rota butonu görünmemeliydi")
	}
}

func TestMerchantRouteButtonHasFooterInset(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	layout := armyPanelGeometry()
	button := merchantRouteAssignmentButtonRect(layout)
	footerY := float64(layout.panelY + layout.panelH - siegeFooterH)
	footerBottom := float64(layout.panelY + layout.panelH)
	if button.Y <= footerY || button.Y+button.H >= footerBottom {
		t.Fatalf("merchant rota butonu footer içine dengeli oturmuyor: button=%+v footerY=%.1f bottom=%.1f", button, footerY, footerBottom)
	}
}

func TestMerchantRoutePanelStaysInsideViewport(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	layout := merchantRoutePanelLayoutFor(12)
	if layout.panelX < 0 || layout.panelY < 0 || layout.panelX+layout.panelW > float32(ScreenWidth) || layout.panelY+layout.panelH > float32(ScreenHeight) {
		t.Fatalf("merchant rota paneli viewport dışına taştı: %+v", layout)
	}
}

func TestMerchantRoutePanelTwoLineRowsHaveVerticalClearance(t *testing.T) {
	if merchantRoutePanelRowH < 46 {
		t.Fatalf("iki satırlı rota seçenekleri için satır yüksekliği yetersiz: %.1f", merchantRoutePanelRowH)
	}

	rowRectHeight := merchantRoutePanelRowH - 10
	if rowRectHeight < 38 {
		t.Fatalf("rota satırının iç kutusu iki satırlı metne sığmıyor: %.1f", rowRectHeight)
	}
}
