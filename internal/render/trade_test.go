package render

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestHandleTradePanelInputMarketSelectsFactionAndGoodOnClick(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	playerID := faction.FactionID("player")
	alphaID := faction.FactionID("alpha")
	betaID := faction.FactionID("beta")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID, NameTR: "Oyuncu", Gold: 200, Grain: 10, Iron: 10, Timber: 10},
			alphaID:  {ID: alphaID, NameTR: "Alfa", Gold: 200, Grain: 20, Iron: 20, Timber: 20},
			betaID:   {ID: betaID, NameTR: "Beta", Gold: 200, Grain: 30, Iron: 30, Timber: 30},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(playerID, alphaID): {FactionA: playerID, FactionB: alphaID, Score: 40, Stance: faction.StanceTrade},
			faction.RelationKey(playerID, betaID):  {FactionA: playerID, FactionB: betaID, Score: 40, Stance: faction.StanceTrade},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: string(playerID), ToFactionID: string(alphaID), Good: economy.GoodGrain, AmountPerTurn: 2, GoldPerUnit: 2},
			{FromFactionID: string(playerID), ToFactionID: string(betaID), Good: economy.GoodGrain, AmountPerTurn: 2, GoldPerUnit: 2},
		},
		MarketPrices: economy.CurrentMarketPrice{
			economy.GoodGrain:  2,
			economy.GoodIron:   5,
			economy.GoodTimber: 3,
			economy.GoodStone:  4,
			economy.GoodSpice:  12,
			economy.GoodCloth:  8,
		},
	}
	r := &Renderer{
		gs:                gs,
		showTrade:         true,
		tradeTab:          TradeTabMarket,
		tradeFactionFocus: 0,
		tradeGoodFocus:    0,
		tradeAmount:       5,
		tradeListFilter:   TradeListAll,
		tradeListSort:     TradeSortDistance,
	}

	layout := tradePanelLayout()
	factionClick := gameui.InputState{MouseX: layout.marketListRect.X + 16, MouseY: layout.marketListRect.Y + 28 + 8, LeftJustPressed: true}
	handleTradePanelInput(r, factionClick)
	if r.tradeFactionFocus != 1 {
		t.Fatalf("ikinci fraksiyon satiri tiklaninca focus 1 olmali, got=%d", r.tradeFactionFocus)
	}

	goodButtons := buildTradeGoodFilterButtons(layout)
	goodClick := gameui.InputState{MouseX: goodButtons[1].Button.X + goodButtons[1].Button.W/2, MouseY: goodButtons[1].Button.Y + goodButtons[1].Button.H/2, LeftJustPressed: true}
	handleTradePanelInput(r, goodClick)
	if r.tradeGoodFocus != 1 {
		t.Fatalf("ikinci mal satiri tiklaninca focus 1 olmali, got=%d", r.tradeGoodFocus)
	}
}

func TestHandleTradePanelInputMarketReturnsEmergencyGrainSale(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	playerID := faction.FactionID("player")
	targetID := faction.FactionID("alpha")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID: {ID: playerID, NameTR: "Oyuncu", Grain: 160},
			targetID: {ID: targetID, NameTR: "Alfa", Gold: 200, Grain: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {ID: "home", OwnerID: string(playerID), BaseGoldIncome: 100, TaxRate: 100, Satisfaction: 50},
		},
		GrainEconomy: map[faction.FactionID]state.GrainEconomyStatus{
			playerID: {FactionID: playerID, StorageCapacity: 100},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(playerID, targetID): {FactionA: playerID, FactionB: targetID, Score: 40, Stance: faction.StanceTrade},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: string(playerID), ToFactionID: string(targetID), Good: economy.GoodGrain, AmountPerTurn: 2, GoldPerUnit: 2},
		},
		MarketPrices: economy.CurrentMarketPrice{economy.GoodGrain: 10},
	}
	r := &Renderer{
		gs:                gs,
		showTrade:         true,
		tradeTab:          TradeTabMarket,
		tradeFactionFocus: 0,
		tradeGoodFocus:    0,
		tradeAmount:       5,
		tradeListFilter:   TradeListAll,
		tradeListSort:     TradeSortDistance,
	}

	layout := tradePanelLayout()
	btn := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, true)
	action := handleTradePanelInput(r, gameui.InputState{
		MouseX:          btn.X + btn.W/2,
		MouseY:          btn.Y + btn.H/2,
		LeftJustPressed: true,
	})
	if action.Kind != ActionEmergencyGrainSale || action.Delta != 5 {
		t.Fatalf("acil tahıl satış aksiyonu dönmeliydi, got=%+v", action)
	}
}

func TestHandleTradePanelInputTogglesAutomaticGrainExport(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	r := &Renderer{
		gs:        &state.GameState{PlayerFactionID: "player"},
		showTrade: true,
		tradeTab:  TradeTabMarket,
	}
	layout := tradePanelLayout()
	btn := buildTradeAutoExportButton(layout, false)
	action := handleTradePanelInput(r, gameui.InputState{
		MouseX:          btn.X + btn.W/2,
		MouseY:          btn.Y + btn.H/2,
		LeftJustPressed: true,
	})
	if action.Kind != ActionToggleAutoGrainExport {
		t.Fatalf("otomatik ihracat toggle aksiyonu dönmeliydi, got=%+v", action)
	}
}

func TestSortedFactionsForMarketListsUnlinkedPeacefulStatesButExcludesEnemies(t *testing.T) {
	playerID := faction.FactionID("player")
	openMarketID := faction.FactionID("open_market")
	enemyID := faction.FactionID("enemy")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		Factions: map[faction.FactionID]*faction.Faction{
			playerID:     {ID: playerID, Grain: 10},
			openMarketID: {ID: openMarketID, Grain: 100},
			enemyID:      {ID: enemyID, Grain: 100},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(playerID, enemyID): {FactionA: playerID, FactionB: enemyID, Stance: faction.StanceWar},
		},
		MarketPrices: economy.CurrentMarketPrice{economy.GoodGrain: 2},
	}

	got := sortedFactionsForMarket(gs, 0, TradeListAll, TradeSortDistance)
	if len(got) != 1 || got[0] != openMarketID {
		t.Fatalf("rota olmadan barıştaki devlet listelenmeli, düşman dışarıda kalmalı: got=%v", got)
	}
}

func TestTradeRouteListFilterDefaultsToOwnedAndTogglesAllRoutes(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	playerID := faction.FactionID("player")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: string(playerID), ToFactionID: "ally"},
			{FromFactionID: "ally", ToFactionID: string(playerID)},
			{FromFactionID: "foreign_a", ToFactionID: "foreign_b"},
		},
	}

	owned := filteredTradeRoutes(gs, TradeRouteFilterOwned)
	if len(owned) != 2 {
		t.Fatalf("oyuncuya ait rota filtresi iki yönlü oyuncu rotalarını göstermeli, got=%d", len(owned))
	}

	all := filteredTradeRoutes(gs, TradeRouteFilterAll)
	if len(all) != 3 {
		t.Fatalf("tüm rotalar filtresi tüm rotaları göstermeli, got=%d", len(all))
	}

	r := &Renderer{gs: gs, showTrade: true, tradeTab: TradeTabRoutes}
	if r.tradeRouteFilter != TradeRouteFilterOwned {
		t.Fatalf("rota filtresinin sıfır değeri başlangıçta oyuncuya ait olmalı, got=%v", r.tradeRouteFilter)
	}
	layout := tradePanelLayout()
	buttons := buildTradeRouteFilterButtons(layout)
	action := handleTradePanelInput(r, gameui.InputState{
		MouseX:          buttons[0].Button.X + buttons[0].Button.W/2,
		MouseY:          buttons[0].Button.Y + buttons[0].Button.H/2,
		LeftJustPressed: true,
	})
	if action.Kind != ActionNone || r.tradeRouteFilter != TradeRouteFilterAll {
		t.Fatalf("tüm rotalar butonu filtreyi değiştirmeli, action=%+v filter=%v", action, r.tradeRouteFilter)
	}
}

func TestTradeQuantityButtonsUseOnlyTenStepsWithoutIcons(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	buttons, _, _ := buildTradeActionButtonsAt(tradeMarketActionCardRect, tradePanelLayout())
	if len(buttons) != 2 {
		t.Fatalf("miktar düğmeleri yalnızca -10 ve +10 içermeli, got=%d", len(buttons))
	}
	if buttons[0].Label != "-10" || buttons[1].Label != "+10" {
		t.Fatalf("miktar düğmeleri beklenmedik: %q, %q", buttons[0].Label, buttons[1].Label)
	}
	for _, button := range buttons {
		if button.Icon != gameui.IconNone {
			t.Fatalf("miktar düğmelerinde ikon olmamalı: label=%q icon=%v", button.Label, button.Icon)
		}
	}
}

func TestTradeMarketPlusTenStopsAtGoldValueLimit(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 1280, 720
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	player := &faction.Faction{Gold: 257}
	layout := tradePanelLayout()

	underLimit, _, _ := buildTradeMarketActionButtons(layout, 240, player, 1)
	if !underLimit[1].Enabled {
		t.Fatal("+10, altın sınırına ulaşılabiliyorsa etkin olmalı")
	}

	atLimit, _, _ := buildTradeMarketActionButtons(layout, 250, player, 1)
	if atLimit[1].Enabled {
		t.Fatal("+10, miktarın değeri mevcut altını aşacaksa pasif olmalı")
	}
	if got := clampTradeAmountToGold(315, player, 1); got != 257 {
		t.Fatalf("miktar altınla karşılanabilecek üst sınıra kırpılmalı: got=%d want=257", got)
	}
}

func TestTradeMarketActionButtonsStayGroupedInsideActionCard(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	ScreenWidth, ScreenHeight = 2048, 1244
	defer func() {
		ScreenWidth, ScreenHeight = oldW, oldH
	}()

	layout := tradePanelLayout()
	cardX, cardY, cardW, cardH := tradeMarketActionCardRect(layout)
	qty, buy, sell := buildTradeActionButtonsAt(tradeMarketActionCardRect, layout)
	emergency := buildTradeEmergencyGrainSaleButtonAt(tradeMarketActionCardRect, layout, true)

	if buy.X != qty[1].X+qty[1].W+18 {
		t.Fatalf("al/sat grubu miktar kontrollerinden kopuk olmamalı: qty=%+v buy=%+v", qty, buy)
	}
	if sell.X != buy.X+buy.W+14 || emergency.X != buy.X {
		t.Fatalf("işlem düğmeleri aynı kompakt kolon grubunda olmalı: buy=%+v sell=%+v emergency=%+v", buy, sell, emergency)
	}
	if qty[0].X <= float64(cardX+12) || sell.X+sell.W >= float64(cardX+cardW-12) {
		t.Fatalf("işlem grubu kart içinde görünür iç boşlukla kalmalı: card=(%.1f,%.1f,%.1f,%.1f) qty=%+v sell=%+v", cardX, cardY, cardW, cardH, qty[0], sell)
	}
}

func TestTradeButtonsUseCommonVerticalCentering(t *testing.T) {
	if got := tradeButtonStyle(true).TextOffsetY; got != 0 {
		t.Fatalf("ticaret sekmeleri ortak dikey merkezlemeyi kullanmalı: offset=%.1f", got)
	}
	if got := tradeButtonStyle(false).TextOffsetY; got != 0 {
		t.Fatalf("pasif ticaret düğmeleri ortak dikey merkezlemeyi kullanmalı: offset=%.1f", got)
	}
}

func TestTradePanelOpensOnMarketTab(t *testing.T) {
	r := &Renderer{tradeTab: TradeTabRoutes}
	r.toggleTradePanel()
	if !r.showTrade || r.tradeTab != TradeTabMarket {
		t.Fatalf("ticaret paneli ilk açılışta Pazar sekmesini göstermeli: show=%v tab=%v", r.showTrade, r.tradeTab)
	}
	r.toggleTradePanel()
	if r.showTrade {
		t.Fatal("ticaret paneli ikinci toggle'da kapanmalı")
	}
}

func TestFactionTradeStatsCalculatesRouteIncomeExpenseAndNet(t *testing.T) {
	playerID := faction.FactionID("player")
	gs := &state.GameState{
		PlayerFactionID: playerID,
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: string(playerID), ToFactionID: "buyer", AmountPerTurn: 2, GoldPerUnit: 5},
			{FromFactionID: "seller", ToFactionID: string(playerID), AmountPerTurn: 3, GoldPerUnit: 4},
			{FromFactionID: string(playerID), ToFactionID: "suspended", AmountPerTurn: 9, GoldPerUnit: 9, SuspendedTurns: 2},
		},
	}

	stats := factionTradeStats(gs, playerID)
	if stats.ExportGold != 10 || stats.ImportGold != 12 || stats.NetGold != -2 {
		t.Fatalf("rota finans özeti beklenmedik: %+v", stats)
	}
	income, expense := tradeRouteFinancialTotals(gs)
	if income != 10 || expense != 12 {
		t.Fatalf("rota finans toplamları beklenmedik: gelir=%d gider=%d", income, expense)
	}
	if stats.RouteCount != 2 || stats.SuspendedCount != 1 {
		t.Fatalf("aktif/askıdaki rota sayısı beklenmedik: %+v", stats)
	}
}
