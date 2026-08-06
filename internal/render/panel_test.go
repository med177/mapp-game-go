package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestDiplomacyOfferQuotaHUDText(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}

	text, col := diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 3/3" {
		t.Fatalf("başlangıçta tam hak görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{232, 190, 100, 255}) {
		t.Fatalf("tam hak rengi farklıydı, got=%v", col)
	}

	gs.DiplomacyOfferCounts = map[faction.FactionID]int{"player": 2}
	text, col = diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 1/3" {
		t.Fatalf("iki teklif sonrası kalan hak 1/3 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 130, 60, 255}) {
		t.Fatalf("tek hak rengi farklıydı, got=%v", col)
	}

	gs.DiplomacyOfferCounts["player"] = 3
	text, col = diplomacyOfferQuotaHUDText(gs)
	if text != "Elçi 0/3" {
		t.Fatalf("hak bitince 0/3 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 90, 90, 255}) {
		t.Fatalf("hak bitti rengi farklıydı, got=%v", col)
	}
}

func TestTradeRouteHUDTextShowsOpenPartnerSlotsInRed(t *testing.T) {
	gs := &state.GameState{PlayerFactionID: "player"}

	text, col := tradeRouteHUDText(gs)
	if text != "Ticaret Rotası: 0/4" {
		t.Fatalf("boş partner limiti görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 90, 90, 255}) {
		t.Fatalf("boş rota slotları kırmızı görünmeliydi, got=%v", col)
	}
}

func TestTradeRouteHUDTextShowsFullPartnerLimitInGreen(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		TradeRoutes: []*economy.TradeRoute{
			{FromFactionID: "player", ToFactionID: "a"},
			{FromFactionID: "a", ToFactionID: "player"},
			{FromFactionID: "player", ToFactionID: "b"},
			{FromFactionID: "b", ToFactionID: "player"},
			{FromFactionID: "player", ToFactionID: "c"},
			{FromFactionID: "c", ToFactionID: "player"},
			{FromFactionID: "player", ToFactionID: "d"},
			{FromFactionID: "d", ToFactionID: "player"},
		},
	}

	text, col := tradeRouteHUDText(gs)
	if text != "Ticaret Rotası: 4/4" {
		t.Fatalf("dolu partner limiti görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{100, 220, 100, 255}) {
		t.Fatalf("dolu rota limiti yeşil görünmeliydi, got=%v", col)
	}
}

func TestWarFatigueHUDTextUsesGreenZeroAndRedPenalty(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
	}

	text, col := warFatigueHUDText(gs)
	if text != "Savaş Yorgunluğu: 0" {
		t.Fatalf("barışta savaş yorgunluğu 0 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{100, 220, 100, 255}) {
		t.Fatalf("sıfır savaş yorgunluğu yeşil görünmeliydi, got=%v", col)
	}

	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("player", "enemy"): {
			FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar,
		},
	}
	text, col = warFatigueHUDText(gs)
	if text != "Savaş Yorgunluğu: -3" {
		t.Fatalf("tek bağımsız savaşta savaş yorgunluğu -3 görünmeliydi, got=%q", text)
	}
	if col != (color.RGBA{220, 90, 90, 255}) {
		t.Fatalf("negatif savaş yorgunluğu kırmızı görünmeliydi, got=%v", col)
	}
}

func TestPlayerMilitaryPowerStandingRanksActiveFactions(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"strong": {ID: "strong"},
			"weak":   {ID: "weak"},
			"dead":   {ID: "dead", IsEliminated: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "player", Units: []army.Unit{{}, {}}},
			"strong-army": {ID: "strong-army", OwnerID: "strong", Units: []army.Unit{{}, {}, {}}},
			"dead-army":   {ID: "dead-army", OwnerID: "dead", Units: []army.Unit{{}, {}, {}, {}}},
		},
	}

	power, rank, count := playerMilitaryPowerStanding(gs)
	if power != 20 || rank != 2 || count != 3 {
		t.Fatalf("oyuncu askeri standing yanlis: power=%d rank=%d count=%d", power, rank, count)
	}
}

func TestPlayerMilitaryPowerStandingUsesFactionIDForTies(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "z-player",
		Factions: map[faction.FactionID]*faction.Faction{
			"a-state":  {ID: "a-state"},
			"z-player": {ID: "z-player"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-army": {ID: "player-army", OwnerID: "z-player", Units: []army.Unit{{}}},
			"other-army":  {ID: "other-army", OwnerID: "a-state", Units: []army.Unit{{}}},
		},
	}

	_, rank, count := playerMilitaryPowerStanding(gs)
	if rank != 2 || count != 2 {
		t.Fatalf("esit guc tie-break sirasi yanlis: rank=%d count=%d", rank, count)
	}
}

func TestFactionMilitaryPowerStandingRanksSelectedFaction(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":   {ID: "player"},
			"selected": {ID: "selected"},
			"strong":   {ID: "strong"},
			"dead":     {ID: "dead", IsEliminated: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"selected-army": {ID: "selected-army", OwnerID: "selected", Units: []army.Unit{{}, {}}},
			"strong-army":   {ID: "strong-army", OwnerID: "strong", Units: []army.Unit{{}, {}, {}}},
			"dead-army":     {ID: "dead-army", OwnerID: "dead", Units: []army.Unit{{}, {}, {}, {}}},
		},
	}

	power, rank, count := factionMilitaryPowerStanding(gs, "selected")
	if power != 20 || rank != 2 || count != 3 {
		t.Fatalf("secili faction askeri standing yanlis: power=%d rank=%d count=%d", power, rank, count)
	}
}

func TestFactionMilitaryPowerBreakdownLabelSeparatesLandAndNavalStrength(t *testing.T) {
	gs := &state.GameState{
		Armies: map[army.ArmyID]*army.Army{
			"land":  {ID: "land", OwnerID: "selected", Units: []army.Unit{{}, {}}},
			"fleet": {ID: "fleet", OwnerID: "selected", IsNaval: true, Units: []army.Unit{{}}},
		},
	}
	if got := factionMilitaryPowerBreakdownLabel(gs, "selected"); got != "20 / 10" {
		t.Fatalf("kara/deniz güç etiketi yanlış: %q", got)
	}
}

func TestRegionGrainProductionDisplayShowsMilitaryRemainderAndTotal(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"bursa": {ID: "bursa", OwnerID: "player", Population: 1000, BaseGrainOutput: 65},
		},
	}
	region := gs.Regions["bursa"]

	if got := regionGrainProductionDisplayValue(gs, region, state.RegionProductionSummary{Grain: 65}); got != "+9/65" {
		t.Fatalf("tahıl üretiminde sivil tüketim sonrası kalan/toplam görünmeliydi, got=%q", got)
	}
}
