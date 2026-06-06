package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func repeatedUnits(typeID string, count int, hp int) []army.Unit {
	out := make([]army.Unit, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, army.Unit{TypeID: typeID, CurrentHP: hp})
	}
	return out
}

func TestCheckRegionUnlocksUnlocksTimedRegionAtTurn(t *testing.T) {
	gs := &state.GameState{
		Turn: 5,
		Regions: map[world.RegionID]*world.Region{
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("timed region açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksDoesNotUnlockTimedRegionEarlyByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 5},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if !gs.Regions["locked"].IsLocked {
		t.Fatal("timed region erken açıldı")
	}
	if len(unlocked) != 0 {
		t.Fatalf("erken unlock listesi boş olmalı, got=%v", unlocked)
	}
}

func TestCheckRegionUnlocksUnlocksDiscoveryRegionByAdjacency(t *testing.T) {
	gs := &state.GameState{
		Turn: 4,
		Regions: map[world.RegionID]*world.Region{
			"src":    {ID: "src", Neighbors: []world.RegionID{"locked"}},
			"locked": {ID: "locked", IsLocked: true, UnlockTurn: 0},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", RegionID: "src"},
		},
	}

	unlocked := checkRegionUnlocks(gs)

	if gs.Regions["locked"].IsLocked {
		t.Fatal("discovery region komşulukla açılmadı")
	}
	if len(unlocked) != 1 || unlocked[0] != "locked" {
		t.Fatalf("beklenen unlock listesi [locked], got=%v", unlocked)
	}
}

func TestApplyEconomyTickAddsTradeIncome(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", Religion: religion.Catholic, Gold: 10, Grain: 0, Spice: 10},
			"b": {ID: "b", Religion: religion.Catholic, Gold: 30, Grain: 0, Spice: 0},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TaxRate: 50, Satisfaction: 50},
			"b1": {ID: "b1", OwnerID: "b", TaxRate: 50, Satisfaction: 50},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b", Good: economy.GoodSpice, AmountPerTurn: 2, GoldPerUnit: 12},
		},
		Armies:    map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{},
	}

	applyEconomyTick(gs)

	// a: 10 (başlangıç) + 24 (2 spice * 12 gold, b'den ödeme) = 34, spice: 10 - 2 = 8
	if gs.Factions["a"].Gold != 34 {
		t.Fatalf("ticaret geliri altına eklenmedi, got=%d", gs.Factions["a"].Gold)
	}
	if gs.Factions["a"].Spice != 8 {
		t.Fatalf("ticaret rotası malı kaynaktan çıkarmalıydı, got=%d", gs.Factions["a"].Spice)
	}
	// b: 30 (başlangıç) - 24 (ticaret ödemesi) = 6, spice: 0 + 2 = 2
	if gs.Factions["b"].Gold != 6 {
		t.Fatalf("ticaret rotası hedeften altın çıkarmalıydı, got=%d", gs.Factions["b"].Gold)
	}
	if gs.Factions["b"].Spice != 2 {
		t.Fatalf("ticaret rotası malı hedefe eklemeliydi, got=%d", gs.Factions["b"].Spice)
	}
}

func TestApplyEconomyTickAppliesMarketGoldBonusToPassiveTrade(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {
				ID: "a",
				Research: faction.ResearchState{
					Completed: map[string]bool{"guilds": true},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a", TaxRate: 50, Satisfaction: 50, TradeCapacity: 5},
		},
		Armies: map[army.ArmyID]*army.Army{},
		TechTypes: map[string]*tech.Technology{
			"guilds": {ID: "guilds", Effects: tech.Effects{MarketGoldMod: 1}},
		},
	}

	applyEconomyTick(gs)

	if gs.Factions["a"].Gold != 22 {
		t.Fatalf("market bonus pasif ticaret gelirini ikiye katlamalıydı, got=%d", gs.Factions["a"].Gold)
	}
}

func TestApplySeasonEffectsAddsNavalMoveBonus(t *testing.T) {
	gs := &state.GameState{
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {
				ID: "player",
				Research: faction.ResearchState{
					Completed: map[string]bool{"navigation": true},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {ID: "fleet", OwnerID: "player", RegionID: "sea_1", IsNaval: true},
			"land":  {ID: "land", OwnerID: "player", RegionID: "land_1"},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_1":  {ID: "sea_1", IsSea: true},
			"land_1": {ID: "land_1", OwnerID: "player"},
		},
		TechTypes: map[string]*tech.Technology{
			"navigation": {ID: "navigation", Effects: tech.Effects{NavalMoveBonus: 1}},
		},
	}

	applySeasonEffects(gs)

	if gs.Armies["fleet"].MaxMovePoints != 3 {
		t.Fatalf("naval move bonus filoya +1 hareket vermeliydi, got=%d", gs.Armies["fleet"].MaxMovePoints)
	}
	if gs.Armies["land"].MaxMovePoints != 2 {
		t.Fatalf("naval move bonus kara ordusunu etkilememeliydi, got=%d", gs.Armies["land"].MaxMovePoints)
	}
}

func TestApplyReligionConversionUsesTechSpeedBonus(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {
				ID:       "owner",
				Religion: religion.Catholic,
				Research: faction.ResearchState{
					Completed: map[string]bool{"proselytism": true},
				},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "owner", Religion: string(religion.Sunni), Satisfaction: 60},
		},
		TechTypes: map[string]*tech.Technology{
			"proselytism": {ID: "proselytism", Effects: tech.Effects{ConversionSpeedMod: 1}},
		},
	}

	applyReligionConversion(gs)

	if gs.Regions["r1"].ConversionTurns != 2 {
		t.Fatalf("conversion speed bonus bir turda +2 ilerletmeliydi, got=%d", gs.Regions["r1"].ConversionTurns)
	}
}

func TestApplyEconomyTickRegionalLogisticsAttritionEscalates(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 40},
		},
		Regions: map[world.RegionID]*world.Region{
			"gelibolu": {ID: "gelibolu", OwnerID: "player", TaxRate: 50, Satisfaction: 50, BaseGrainOutput: 4},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a2": {ID: "a2", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a3": {ID: "a3", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
			"a4": {ID: "a4", OwnerID: "player", RegionID: "gelibolu", Units: repeatedUnits("inf", 5, 100)},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
	}

	report1 := applyEconomyTick(gs)
	firstStatus := gs.ArmyLogistics["a1"]
	if firstStatus.TotalHPDamage == 0 || firstStatus.DamagePerUnit == 0 {
		t.Fatal("ilk tur bölgesel ikmal zayiatı bekleniyordu")
	}
	if gs.Armies["a1"].OverCapacityTurns != 1 {
		t.Fatalf("ilk tur over-capacity sayacı 1 olmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if len(report1.PlayerLogisticsAlerts) != 1 {
		t.Fatalf("oyuncu için tek lojistik uyarısı bekleniyordu, got=%d", len(report1.PlayerLogisticsAlerts))
	}

	report2 := applyEconomyTick(gs)
	secondStatus := gs.ArmyLogistics["a1"]
	if secondStatus.DamagePerUnit <= firstStatus.DamagePerUnit {
		t.Fatalf("uzun süreli yığılmada zayiat artmalı, first=%d second=%d", firstStatus.DamagePerUnit, secondStatus.DamagePerUnit)
	}
	if gs.Armies["a1"].OverCapacityTurns != 2 {
		t.Fatalf("ikinci tur over-capacity sayacı 2 olmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if len(report2.PlayerLogisticsAlerts) != 1 || report2.PlayerLogisticsAlerts[0].Overload <= 0 {
		t.Fatalf("ikinci tur da oyuncu lojistik uyarısı sürmeliydi, got=%+v", report2.PlayerLogisticsAlerts)
	}
}

func TestApplyEconomyTickRegionalLogisticsResetsWhenCapacityRecovers(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Grain: 20},
		},
		Regions: map[world.RegionID]*world.Region{
			"safe": {ID: "safe", OwnerID: "player", TaxRate: 50, Satisfaction: 50, BaseGrainOutput: 12},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", RegionID: "safe", Units: repeatedUnits("inf", 1, 100), OverCapacityTurns: 3},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", GrainUpkeep: 2},
		},
	}

	report := applyEconomyTick(gs)

	if gs.Armies["a1"].OverCapacityTurns != 0 {
		t.Fatalf("kapasite yeterliyken sayaç sıfırlanmalı, got=%d", gs.Armies["a1"].OverCapacityTurns)
	}
	if status := gs.RegionLogistics["safe"]; status.Overload > 0 {
		t.Fatalf("güvenli bölgede overload olmamalı, got=%+v", status)
	}
	if _, ok := gs.ArmyLogistics["a1"]; ok {
		t.Fatal("kapasite yeterliyken orduya zayiat kaydı yazılmamalı")
	}
	if len(report.PlayerLogisticsAlerts) != 0 {
		t.Fatalf("kapasite yeterliyken oyuncu uyarısı olmamalı, got=%+v", report.PlayerLogisticsAlerts)
	}
}

func TestCheckEliminationsRemovesArmiesAndRelations(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
		},
		Regions: map[world.RegionID]*world.Region{
			"b1": {ID: "b1", OwnerID: "b"},
			"c1": {ID: "c1", OwnerID: "c"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "a"},
			"b1": {ID: "b1", OwnerID: "b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
			faction.RelationKey("a", "c"): {FactionA: "a", FactionB: "c", Stance: faction.StancePeace},
			faction.RelationKey("b", "c"): {FactionA: "b", FactionB: "c", Stance: faction.StanceAllied},
		},
		DiplomaticOffers: []state.DiplomaticOffer{
			{FromFactionID: "a", ToFactionID: "b", Action: "propose_peace"},
			{FromFactionID: "b", ToFactionID: "c", Action: "propose_trade"},
		},
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "a", ToFactionID: "b"},
			{FromFactionID: "b", ToFactionID: "c"},
		},
	}

	checkEliminations(gs)

	if !gs.Factions["a"].IsEliminated {
		t.Fatal("bölgesi kalmayan fraksiyon elenmiş işaretlenmeli")
	}
	if _, ok := gs.Armies["a1"]; ok {
		t.Fatal("elenen fraksiyonun ordusu temizlenmeliydi")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "b")]; ok {
		t.Fatal("elenen fraksiyonun ilişkileri temizlenmeliydi (a|b)")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "c")]; ok {
		t.Fatal("elenen fraksiyonun ilişkileri temizlenmeliydi (a|c)")
	}
	if _, ok := gs.Relations[faction.RelationKey("b", "c")]; !ok {
		t.Fatal("elenmeyen fraksiyonlar arası ilişki korunmalı")
	}
	if len(gs.DiplomaticOffers) != 1 || gs.DiplomaticOffers[0].FromFactionID != "b" {
		t.Fatalf("elenen fraksiyonun teklifleri temizlenmeliydi, got=%+v", gs.DiplomaticOffers)
	}
	if len(gs.TradeRoutes) != 1 || gs.TradeRoutes[0].FromFactionID != "b" || gs.TradeRoutes[0].ToFactionID != "c" {
		t.Fatalf("elenen fraksiyonun ticaret rotalari temizlenmeliydi, got=%+v", gs.TradeRoutes)
	}
}

func TestCheckEliminationsRemovesSeaOnlyFactionWithFleets(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_a":  {ID: "sea_a", OwnerID: "a", IsSea: true},
			"land_b": {ID: "land_b", OwnerID: "b", IsSea: false},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a_fleet": {ID: "a_fleet", OwnerID: "a", IsNaval: true},
			"a_land":  {ID: "a_land", OwnerID: "a", IsNaval: false},
			"b_land":  {ID: "b_land", OwnerID: "b", IsNaval: false},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
		},
	}

	checkEliminations(gs)

	if !gs.Factions["a"].IsEliminated {
		t.Fatal("sadece deniz bölgesi kalan fraksiyon elenmeli")
	}
	if _, ok := gs.Armies["a_fleet"]; ok {
		t.Fatal("elenen fraksiyonun donanması temizlenmeliydi")
	}
	if _, ok := gs.Armies["a_land"]; ok {
		t.Fatal("elenen fraksiyonun kara ordusu temizlenmeliydi")
	}
	if _, ok := gs.Relations[faction.RelationKey("a", "b")]; ok {
		t.Fatal("elenen fraksiyonun diplomasi ilişkileri temizlenmeliydi")
	}
}

func TestApplySeasonEffectsReplenishesFriendlyLandArmy(t *testing.T) {
	gs := &state.GameState{
		Month:           4,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home":      {ID: "home", OwnerID: "player"},
			"enemy":     {ID: "enemy", OwnerID: "enemy"},
			"ally_port": {ID: "ally_port", OwnerID: "ally"},
			"sea":       {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"home_army":   {ID: "home_army", OwnerID: "player", RegionID: "home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 65}}},
			"enemy_army":  {ID: "enemy_army", OwnerID: "player", RegionID: "enemy", Units: []army.Unit{{TypeID: "inf", CurrentHP: 65}}},
			"fleet":       {ID: "fleet", OwnerID: "player", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 65}}},
			"docked_self": {ID: "docked_self", OwnerID: "player", RegionID: "sea", DockedRegionID: "home", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 65}}},
			"docked_ally": {ID: "docked_ally", OwnerID: "player", RegionID: "sea", DockedRegionID: "ally_port", IsNaval: true, Units: []army.Unit{{TypeID: "transport", CurrentHP: 65}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ally", "player"): {FactionA: "ally", FactionB: "player", Stance: faction.StanceAllied},
		},
	}

	applySeasonEffects(gs)

	if got := gs.Armies["home_army"].Units[0].CurrentHP; got != 75 {
		t.Fatalf("dost topraktaki ordu iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["enemy_army"].Units[0].CurrentHP; got != 65 {
		t.Fatalf("dost olmayan toprakta iyilesme olmamali, got=%d", got)
	}
	if got := gs.Armies["fleet"].Units[0].CurrentHP; got != 65 {
		t.Fatalf("donanma kara takviyesi almamali, got=%d", got)
	}
	if got := gs.Armies["docked_self"].Units[0].CurrentHP; got != 75 {
		t.Fatalf("kendi limanındaki donanma iyilesmeli, got=%d", got)
	}
	if got := gs.Armies["docked_ally"].Units[0].CurrentHP; got != 70 {
		t.Fatalf("müttefik limanındaki donanma yari hizda iyilesmeli, got=%d", got)
	}
}
