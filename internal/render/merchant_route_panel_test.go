package render

import (
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
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
	if merchantRoutePanelRowH < 56 {
		t.Fatalf("iki satırlı rota seçenekleri için satır yüksekliği yetersiz: %.1f", merchantRoutePanelRowH)
	}

	rowRectHeight := merchantRoutePanelRowH - 10
	if rowRectHeight < 46 {
		t.Fatalf("rota satırının iç kutusu iki satırlı metne sığmıyor: %.1f", rowRectHeight)
	}
}

func TestMerchantRoutePanelUsesSharedCloseIconButton(t *testing.T) {
	layout := merchantRoutePanelLayoutFor(1)
	if layout.close.Label != "" || layout.close.Icon != gameui.IconClose {
		t.Fatalf("merchant rota paneli ortak kapatma ikonunu kullanmalı: %+v", layout.close)
	}
	if layout.close.IconSize != 13 {
		t.Fatalf("merchant rota kapatma ikonu diğer panellerdeki boyutu kullanmalı: %.1f", layout.close.IconSize)
	}
}

func TestMerchantRoutePanelRowGeometryMatchesExpandedHeight(t *testing.T) {
	layout := merchantRoutePanelLayoutFor(2)
	first := merchantRoutePanelRowRect(layout, 0)
	second := merchantRoutePanelRowRect(layout, 1)
	if first.H != float64(merchantRoutePanelRowH-10) {
		t.Fatalf("ilk rota satırı yüksekliği ortak row geometry ile eşleşmiyor: %+v", first)
	}
	if second.Y-first.Y != float64(merchantRoutePanelRowH) {
		t.Fatalf("rota satırları genişletilmiş row yüksekliğini kullanmıyor: first=%+v second=%+v", first, second)
	}
}

func TestMerchantRouteSeaDisplayNameUsesTargetSeaRegion(t *testing.T) {
	gs := &state.GameState{
		Year: 1300,
		Regions: map[world.RegionID]*world.Region{
			"from":    {ID: "from", OwnerID: "from_faction", Neighbors: []world.RegionID{"aegean"}},
			"to":      {ID: "to", OwnerID: "to_faction", Neighbors: []world.RegionID{"marmara"}},
			"aegean":  {ID: "aegean", NameTR: "Ege Denizi", IsSea: true},
			"marmara": {ID: "marmara", NameTR: "Marmara Denizi", IsSea: true},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "from", Links: []world.RegionID{"to"}},
			{ID: "to", Links: []world.RegionID{"from"}},
		}},
	}
	route := &economy.TradeRoute{FromFactionID: "from_faction", ToFactionID: "to_faction"}

	if got := merchantRouteSeaDisplayName(gs, route); got != "Marmara Denizi" {
		t.Fatalf("rota denizleri yanlış gösteriliyor: %q", got)
	}
}

func TestMerchantRouteHighlightClearsWhenAnotherRegionIsSelected(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"target_sea": {ID: "target_sea", IsSea: true},
				"other":      {ID: "other"},
			},
		},
		merchantRouteHighlight: "target_sea",
	}

	r.selectMapRegion("target_sea")
	if r.merchantRouteHighlight != "target_sea" {
		t.Fatalf("hedef deniz seçildiğinde rota vurgusu korunmalıydı: %q", r.merchantRouteHighlight)
	}

	r.selectMapRegion("other")
	if r.merchantRouteHighlight != "" {
		t.Fatalf("başka bölge seçildiğinde rota vurgusu temizlenmeliydi: %q", r.merchantRouteHighlight)
	}
}

func TestMerchantRouteSelectionFocusesTargetAndClosesArmyPanel(t *testing.T) {
	gs := &state.GameState{
		Year: 1300,
		Regions: map[world.RegionID]*world.Region{
			"from":    {ID: "from", OwnerID: "from_faction", Neighbors: []world.RegionID{"aegean"}},
			"to":      {ID: "to", OwnerID: "to_faction", Neighbors: []world.RegionID{"marmara"}},
			"aegean":  {ID: "aegean", IsSea: true},
			"marmara": {ID: "marmara", IsSea: true},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "from", Links: []world.RegionID{"to"}},
			{ID: "to", Links: []world.RegionID{"from"}},
		}},
	}
	r := &Renderer{
		gs:                     gs,
		worldMap:               &WorldMap{regionAnchor: map[world.RegionID][2]int{"marmara": {700, 500}}},
		SelectedArmy:           "fleet",
		splitSelectedUnits:     map[int]bool{0: true},
		merchantRouteHighlight: "",
	}
	r.focusMerchantRouteTarget(&economy.TradeRoute{FromFactionID: "from_faction", ToFactionID: "to_faction"})

	if r.merchantRouteHighlight != "marmara" {
		t.Fatalf("rota seçimi hedef denizi işaretlemeliydi: %q", r.merchantRouteHighlight)
	}
	if r.camX != 700 || r.camY != 500 {
		t.Fatalf("rota seçimi kamerayı hedef deniz anchor'ına odaklamalıydı: got=(%.1f,%.1f)", r.camX, r.camY)
	}
	if r.SelectedArmy != "" {
		t.Fatalf("rota seçimi açık ordu panelinin seçimini temizlemeliydi: %q", r.SelectedArmy)
	}
	if len(r.splitSelectedUnits) != 0 {
		t.Fatalf("rota seçimi ordu bölme seçimini temizlemeliydi: %+v", r.splitSelectedUnits)
	}
}

func TestMerchantRouteSeaDisplayNameHandlesMissingRouteSea(t *testing.T) {
	if got := merchantRouteSeaDisplayName(nil, nil); got != "Bilinmiyor" {
		t.Fatalf("eksik rota denizi için fallback yanlış: %q", got)
	}
}

func TestMerchantRouteOptionDisabledWhenCapacityFull(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "player", ToFactionID: "partner", AmountPerTurn: 2}
	gs := &state.GameState{
		UnitTypes:   map[string]*army.UnitType{"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade}},
		TradeRoutes: []*economy.TradeRoute{route},
		Armies: map[army.ArmyID]*army.Army{
			"assigned": {ID: "assigned", OwnerID: "player", IsNaval: true, TradeRouteKey: route.AssignmentKey(), Units: []army.Unit{{TypeID: "merchant_ship"}, {TypeID: "merchant_ship"}}},
			"incoming": {ID: "incoming", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "merchant_ship"}}},
		},
	}
	r := &Renderer{gs: gs, merchantRouteArmy: "incoming"}
	if r.merchantRouteOptionEnabled(route) {
		t.Fatal("kapasitesi dolu rota panelde etkin görünmemeliydi")
	}
	r.merchantRouteArmy = "assigned"
	if !r.merchantRouteOptionEnabled(route) {
		t.Fatal("mevcut filonun aktif rotası pasif görünmemeliydi")
	}
}

func TestMerchantTradeBonusForArmyOnlyShowsActiveTargetSeaBonus(t *testing.T) {
	route := &economy.TradeRoute{FromFactionID: "from_faction", ToFactionID: "to_faction", AmountPerTurn: 2, GoldPerUnit: 5}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"from":    {ID: "from", OwnerID: "from_faction", Neighbors: []world.RegionID{"aegean"}},
			"to":      {ID: "to", OwnerID: "to_faction", Neighbors: []world.RegionID{"marmara"}},
			"aegean":  {ID: "aegean", IsSea: true},
			"marmara": {ID: "marmara", IsSea: true},
		},
		TradeCenters: world.TradeCenterConfig{Centers: []world.TradeCenterDef{
			{ID: "from", Links: []world.RegionID{"to"}},
			{ID: "to", Links: []world.RegionID{"from"}},
		}},
		TradeRoutes: []*economy.TradeRoute{route},
		UnitTypes: map[string]*army.UnitType{
			"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		},
	}
	active := &army.Army{
		ID: "active", OwnerID: "from_faction", IsNaval: true, RegionID: "marmara",
		TradeRouteKey: route.AssignmentKey(), Units: []army.Unit{{TypeID: "merchant_ship"}, {TypeID: "merchant_ship"}},
	}
	away := *active
	away.ID = "away"
	away.RegionID = "aegean"
	away.TradeRouteKey = ""
	gs.Armies = map[army.ArmyID]*army.Army{"active": active, "away": &away}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{regionAnchor: map[world.RegionID][2]int{
			"marmara": {100, 100},
		}},
	}

	if got := r.merchantTradeBonusForArmy(active); got != 2 {
		t.Fatalf("aktif hedef denizde merchant bonusu +2 olmalıydı, got=%d", got)
	}
	if got := r.merchantTradeBonusForArmy(&army.Army{IsNaval: true, RegionID: "marmara", Units: []army.Unit{{TypeID: "merchant_ship"}}}); got != 0 {
		t.Fatalf("rotasız filo bonus rozeti üretmemeliydi, got=%d", got)
	}
	if got := r.merchantTradeBonusForArmy(&away); got != 0 {
		t.Fatalf("hedef denizden uzaktaki filo bonus rozeti üretmemeliydi, got=%d", got)
	}
	pending := *active
	pending.ID = "pending"
	pending.RegionID = "aegean"
	if !r.merchantTradeAssignmentPendingForArmy(&pending) {
		t.Fatal("hedef denize ulaşmamış atanmış merchant filosu bekleyen görev olarak işaretlenmeliydi")
	}
	positions := r.armyIconPositions()
	var activePos armyIconPos
	activeFound := false
	for _, pos := range positions {
		if pos.ArmyID == active.ID {
			activePos = pos
			activeFound = true
			break
		}
	}
	if !activeFound {
		t.Fatal("aktif merchant filosunun harita konumu bulunmalı")
	}
	badge := merchantTradeBonusBadgeRect(activePos.X, activePos.Y)
	if aid, ok := r.merchantTradeBonusHitAt(badge.X+badge.W/2, badge.Y+badge.H/2); !ok || aid != active.ID {
		t.Fatalf("ticaret rozeti hover hit-test'i aktif filoyu bulmalı: aid=%q hit=%t", aid, ok)
	}
	title, detail, ok := merchantTradeBonusTooltipText(gs, active)
	if !ok || title != "Ticaret rotası bonusu" || !strings.Contains(detail, "+2 mal/tur") || !strings.Contains(detail, "+10 altın/tur") {
		t.Fatalf("ticaret rozeti tooltip'i bonus ve tur başı geliri göstermeli: title=%q detail=%q ok=%t", title, detail, ok)
	}
	route.BlockadePercent = 50
	_, detail, ok = merchantTradeBonusTooltipText(gs, active)
	if !ok || !strings.Contains(detail, "+5 altın/tur") {
		t.Fatalf("abluka altındaki ticaret geliri tooltip'te kesintili görünmeli: detail=%q ok=%t", detail, ok)
	}
}
