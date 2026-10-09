package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func relationshipRepairTestState(gold int) (*state.GameState, *faction.Relation) {
	actor := faction.FactionID("actor")
	target := faction.FactionID("target")
	rel := &faction.Relation{
		FactionA:  actor,
		FactionB:  target,
		ScoreAToB: 20,
		ScoreBToA: 20,
		Stance:    faction.StanceTrade,
	}
	return &state.GameState{
		Turn: 1,
		DiplomacyConfig: scenario.DiplomacyConfig{
			RelationImprovementGoldCost: 1000,
			RelationImprovementBonus:    5,
			GiftGoldCost:                2500,
			GiftReceiverGold:            1000,
			GiftRelationBonus:           12,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor, Gold: gold},
			target: {ID: target, Gold: 5000},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(actor, target): rel,
		},
		TradeRoutes: []*economy.TradeRoute{{
			FromFactionID: string(actor),
			ToFactionID:   string(target),
			AmountPerTurn: 1,
		}},
	}, rel
}

func TestAIGiftTreasuryGateRejectsHalfTreasuryGift(t *testing.T) {
	gs, rel := relationshipRepairTestState(5000)
	actor := gs.Factions[rel.FactionA]

	if aiGiftTreasurySafe(actor, 2500, nil) {
		t.Fatal("2.500 altınlık hediye, 5.000 altın hazinede güvenli kabul edildi")
	}

	// Deterministik başarı zarı bu turu seçmeyebilir; iki durumda da güvenli
	// olmayan hediye uygulanmamalı, en fazla heyet uygulanmalıdır.
	aiHandleRelationshipRepairWithBudget(gs, rel.FactionA, rel.FactionB, rel, nil)
	if actor.Gold == 2500 || rel.ScoreFrom(rel.FactionA) == 32 {
		t.Fatal("AI güvenli olmayan hediyeyi uyguladı")
	}
}

func TestAIGiftTreasuryGateAllowsGiftWithLargeReserve(t *testing.T) {
	actor := &faction.Faction{ID: "actor", Gold: 10000}
	if !aiGiftTreasurySafe(actor, 2500, nil) {
		t.Fatal("yeterli hazine rezervi olan AI için hediye gereksiz yere engellendi")
	}
}

func resourceNegotiationTestState(relationScore int) *state.GameState {
	actor := faction.FactionID("actor")
	target := faction.FactionID("target")
	player := faction.FactionID("player")
	return &state.GameState{
		Turn:            4,
		DecisionSeed:    17,
		PlayerFactionID: player,
		Factions: map[faction.FactionID]*faction.Faction{
			actor:  {ID: actor, NameTR: "AI", Grain: 300},
			target: {ID: target, NameTR: "Hedef", Iron: 100},
			player: {ID: player, NameTR: "Oyuncu"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey(actor, target): {
				FactionA: actor, FactionB: target,
				ScoreAToB: relationScore, ScoreBToA: relationScore,
				Stance: faction.StancePeace,
			},
			faction.RelationKey(actor, player): {
				FactionA: actor, FactionB: player,
				ScoreAToB: relationScore, ScoreBToA: relationScore,
				Stance: faction.StancePeace,
			},
		},
	}
}

func TestAIResourceNegotiationQueuesOfferForPlayer(t *testing.T) {
	gs := resourceNegotiationTestState(20)
	gs.MarketPrices = economy.CurrentMarketPrice{
		economy.GoodGrain: 6,
		economy.GoodIron:  10,
	}
	gs.Factions["player"].Iron = 100
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("AI oyuncuya kaynak pazarlığı önermedi")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("bekleyen pazarlık sayısı = %d, 1 bekleniyordu", len(gs.DiplomaticOffers))
	}
	offer := gs.DiplomaticOffers[0]
	if offer.Action != string(diplomacy.ActionProposeTransfer) || offer.ToFactionID != "player" {
		t.Fatalf("beklenmeyen pazarlık teklifi: %+v", offer)
	}
	if len(offer.RequestedTransfers) != 1 || offer.RequestedTransfers[0].ID != "iron" || offer.RequestedTransfers[0].Amount != 20 {
		t.Fatalf("AI ihtiyacını karşılayan kaynak talebi oluşturmadı: %+v", offer.RequestedTransfers)
	}
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].ID != "grain" || offer.OfferedTransfers[0].Amount <= 0 {
		t.Fatalf("AI fazla kaynağını karşı teklif olarak sunmadı: %+v", offer.OfferedTransfers)
	}
	if offer.OfferedTransfers[0].Amount < 30 || offer.OfferedTransfers[0].Amount > 33 {
		t.Fatalf("AI pazarlığı açık pazar fiyatından %%10'dan fazla indirimli/değerli yaptı: %+v", offer.OfferedTransfers[0])
	}
	if gs.Factions["actor"].Grain != 300 || gs.Factions["actor"].Iron != 0 {
		t.Fatal("oyuncu karar vermeden pazarlık kaynakları aktarıldı")
	}
	if len(steps) != 1 || steps[0].Kind != TurnStepDiplomacy {
		t.Fatalf("oyuncuya pazarlık turn step'i bildirilmedi: %+v", steps)
	}
}

func TestAIResourceNegotiationSkipsResourceTargetAlreadyHasInExcess(t *testing.T) {
	gs := resourceNegotiationTestState(100)
	gs.Factions["target"].Grain = 100_000
	var steps []TurnStep

	if aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("AI, 100.000 tahılı olan devlete fazladan tahıl satarak pazarlık teklif etmemeliydi")
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("stok fazlasına rağmen gereksiz kaynak pazarlığı kuyruğa girdi: %+v", gs.DiplomaticOffers)
	}
}

func TestAIStrategicRegionClaimCreatesPaidTransferOffer(t *testing.T) {
	gs := resourceNegotiationTestState(100)
	gs.Factions["actor"].Gold = 5000
	gs.Regions = map[world.RegionID]*world.Region{
		"claimed-region": {ID: "claimed-region", OwnerID: "player", BaseGoldIncome: 20, TaxRate: 50, BaseGrainOutput: 20, BaseIronOutput: 3, Population: 10000},
	}
	gs.Factions["actor"].TerritorialClaims = []faction.TerritorialClaim{{RegionID: "claimed-region", Value: 85}}
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{}, &steps) {
		t.Fatal("AI stratejik claim bölgesi için satın alma teklifi oluşturmadı")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("bölge teklifi kuyruğa eklenmedi: %+v", gs.DiplomaticOffers)
	}
	offer := gs.DiplomaticOffers[0]
	if len(offer.RequestedTransfers) != 1 || offer.RequestedTransfers[0].Kind != "region" || offer.RequestedTransfers[0].ID != "claimed-region" {
		t.Fatalf("AI hedeflediği bölgeyi istemedi: %+v", offer.RequestedTransfers)
	}
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].Kind != "resource" || offer.OfferedTransfers[0].ID != "gold" || offer.OfferedTransfers[0].Amount <= 0 {
		t.Fatalf("bölge karşılığında altın ödemesi teklif edilmedi: %+v", offer.OfferedTransfers)
	}
	requestedValue := diplomacy.TransferListValue(gs, "player", offer.RequestedTransfers)
	offeredValue := diplomacy.TransferListValue(gs, "actor", offer.OfferedTransfers)
	if offeredValue < requestedValue {
		t.Fatalf("bölge teklifi en az 10 turluk efektif geliri karşılamıyor: requested=%d offered=%d", requestedValue, offeredValue)
	}
}

func TestAIStrategicRegionClaimSkipsOfferBelowTenTurnIncome(t *testing.T) {
	gs := resourceNegotiationTestState(100)
	gs.Factions["actor"].Gold = 500
	gs.Regions = map[world.RegionID]*world.Region{
		"claimed-region": {ID: "claimed-region", OwnerID: "player", BaseGoldIncome: 20, TaxRate: 50, BaseGrainOutput: 20, BaseIronOutput: 3, Population: 10000},
	}
	gs.Factions["actor"].TerritorialClaims = []faction.TerritorialClaim{{RegionID: "claimed-region", Value: 85}}
	var steps []TurnStep

	if aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{}, &steps) {
		t.Fatal("AI, 10 turluk bölge getirisini karşılayamayan ödemeyle satın alma teklifi göndermemeliydi")
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("yetersiz bölge bedeliyle teklif kuyruğa girdi: %+v", gs.DiplomaticOffers)
	}
}

func TestAIStrategicRegionClaimCanOfferAnotherUnclaimedRegion(t *testing.T) {
	gs := resourceNegotiationTestState(100)
	gs.Regions = map[world.RegionID]*world.Region{
		"claimed-region": {ID: "claimed-region", OwnerID: "player", BaseGoldIncome: 15, TaxRate: 50, BaseGrainOutput: 20, Population: 10000},
		"actor-home":     {ID: "actor-home", OwnerID: "actor", BaseGoldIncome: 15, TaxRate: 50, BaseGrainOutput: 20, Population: 10000},
		"spare-region":   {ID: "spare-region", OwnerID: "actor", BaseGoldIncome: 15, TaxRate: 50, BaseGrainOutput: 20, Population: 10000},
	}
	gs.Factions["actor"].TerritorialClaims = []faction.TerritorialClaim{
		{RegionID: "claimed-region", Value: 85},
		{RegionID: "actor-home", Value: 100, Core: true},
	}
	gs.Factions["actor"].Gold = 0
	gs.Factions["actor"].Grain = 0
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{}, &steps) {
		t.Fatal("AI uygun bölge takası teklifi oluşturmadı")
	}
	offer := gs.DiplomaticOffers[0]
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].Kind != "region" || offer.OfferedTransfers[0].ID != "spare-region" {
		t.Fatalf("AI başkent/claim olmayan bölgesini takas için sunmadı: %+v", offer.OfferedTransfers)
	}
}

func mercenaryTradeTestState() *state.GameState {
	gs := resourceNegotiationTestState(100)
	gs.Regions = map[world.RegionID]*world.Region{
		"actor-region":  {ID: "actor-region", OwnerID: "actor", Population: 10000},
		"player-region": {ID: "player-region", OwnerID: "player", Population: 10000},
	}
	gs.UnitTypes = map[string]*army.UnitType{
		"infantry": {ID: "infantry", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 10, Tier: 1, GoldCost: 100, GrainCost: 10},
	}
	return gs
}

func addMercenaryTestArmy(gs *state.GameState, owner faction.FactionID, id army.ArmyID, region world.RegionID, amount int) {
	units := make([]army.Unit, amount)
	for index := range units {
		units[index] = army.Unit{TypeID: "infantry", CurrentHP: army.MaxUnitHP}
	}
	gs.Armies[id] = &army.Army{ID: id, OwnerID: string(owner), RegionID: region, Units: units}
}

func TestAIMercenaryOfferKeepsMostArmyAndTargetsMilitaryShortage(t *testing.T) {
	gs := mercenaryTradeTestState()
	gs.Factions["player"].Gold = 1000
	gs.Armies = make(map[army.ArmyID]*army.Army)
	addMercenaryTestArmy(gs, "actor", "actor-army", "actor-region", 8)
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{}, &steps) {
		t.Fatal("AI ordusunun güvenli fazlası için paralı asker teklifi oluşturmadı")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("paralı asker teklifi kuyruğa eklenmedi: %+v", gs.DiplomaticOffers)
	}
	offer := gs.DiplomaticOffers[0]
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].Kind != "army" || offer.OfferedTransfers[0].Amount != 2 {
		t.Fatalf("AI askerî rezervinin çoğunu koruyarak en fazla iki asker sunmalıydı: %+v", offer.OfferedTransfers)
	}
	if len(offer.RequestedTransfers) != 1 || offer.RequestedTransfers[0].Kind != "resource" || offer.RequestedTransfers[0].ID != "gold" {
		t.Fatalf("paralı asker karşılığında ödeme istenmedi: %+v", offer.RequestedTransfers)
	}
}

func TestAIMercenaryPurchaseOffersPaymentForMissingArmy(t *testing.T) {
	gs := mercenaryTradeTestState()
	gs.Factions["actor"].Gold = 1000
	gs.Armies = make(map[army.ArmyID]*army.Army)
	addMercenaryTestArmy(gs, "player", "player-army", "player-region", 8)
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{}, &steps) {
		t.Fatal("AI askerî kuvvet açığı için asker satın alma teklifi oluşturmadı")
	}
	if len(gs.DiplomaticOffers) != 1 {
		t.Fatalf("asker satın alma teklifi kuyruğa eklenmedi: %+v", gs.DiplomaticOffers)
	}
	offer := gs.DiplomaticOffers[0]
	if offer.ToFactionID != "player" || len(offer.RequestedTransfers) != 1 || offer.RequestedTransfers[0].Kind != "army" {
		t.Fatalf("AI oyuncudan asker istemedi: %+v", offer)
	}
	if len(offer.OfferedTransfers) != 1 || offer.OfferedTransfers[0].Kind != "resource" || offer.OfferedTransfers[0].ID != "gold" {
		t.Fatalf("asker satın almak için dengeli ödeme sunulmadı: %+v", offer.OfferedTransfers)
	}
	productionCost := aiGoldEquivalentCost(gs, aiUnitResourceCost(gs.UnitTypes[offer.RequestedTransfers[0].ID])) * offer.RequestedTransfers[0].Amount
	maximumPrice := productionCost * aiMercenaryPurchaseCostPercent / 100
	offeredValue := diplomacy.TransferListValue(gs, "actor", offer.OfferedTransfers)
	if offeredValue <= 0 || offeredValue > maximumPrice {
		t.Fatalf("AI askerleri üretim maliyetinin en az %%20 altında almaya çalışmalı: cost=%d cap=%d offered=%d", productionCost, maximumPrice, offeredValue)
	}
}

func TestAIResourceNegotiationResolvesAIToAIOffer(t *testing.T) {
	gs := resourceNegotiationTestState(20)
	gs.Factions["target"].Grain = 0
	var steps []TurnStep

	if !aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("AI-AI kaynak pazarlığı uygulanmadı")
	}
	if got := gs.Factions["actor"].Iron; got != 20 {
		t.Fatalf("AI'nin demir stoku = %d, 20 bekleniyordu", got)
	}
	if got := gs.Factions["target"].Iron; got != 80 {
		t.Fatalf("hedef demir stoku = %d, 80 bekleniyordu", got)
	}
	if gs.Factions["actor"].Grain >= 300 || gs.Factions["target"].Grain <= 0 {
		t.Fatalf("AI-AI pazarlığında tahıl aktarımı olmadı: actor=%d target=%d", gs.Factions["actor"].Grain, gs.Factions["target"].Grain)
	}
	if len(gs.DiplomaticOffers) != 0 || len(steps) != 1 {
		t.Fatalf("AI-AI pazarlığı kuyruğu/turn step'i beklenmedik: offers=%d steps=%d", len(gs.DiplomaticOffers), len(steps))
	}
}

func TestAIResourceNegotiationDoesNotOverpayWhenTargetRelationIsLow(t *testing.T) {
	gs := resourceNegotiationTestState(0)
	var steps []TurnStep

	if aiHandleResourceNegotiationWithSteps(gs, "actor", economy.ResourceCost{Iron: 20}, &steps) {
		t.Fatal("düşük ilişki puanında AI kabul olasılığı sağlamayan pahalı bir teklif yaptı")
	}
	if len(gs.DiplomaticOffers) != 0 {
		t.Fatalf("reddedilmesi beklenen pazarlık kuyruğa girdi: %+v", gs.DiplomaticOffers)
	}
}
