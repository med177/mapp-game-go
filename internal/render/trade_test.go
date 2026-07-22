package render

import (
	"testing"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
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
	factionClick := gameui.InputState{MouseX: layout.leftListRect.X + 16, MouseY: layout.leftListRect.Y + 28 + 8, LeftJustPressed: true}
	handleTradePanelInput(r, factionClick)
	if r.tradeFactionFocus != 1 {
		t.Fatalf("ikinci fraksiyon satiri tiklaninca focus 1 olmali, got=%d", r.tradeFactionFocus)
	}

	goodClick := gameui.InputState{MouseX: layout.rightListRect.X + 16, MouseY: layout.rightListRect.Y + 28 + 8, LeftJustPressed: true}
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
	btn := buildTradeEmergencyGrainSaleButton(layout, len(tradeSelectableGoods()), true)
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
