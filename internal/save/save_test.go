package save

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"testing"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/victory"
	"mapp-game-go/internal/world"
)

func TestLoadScenarioBaseStateReadsAggressiveExpansionDuration(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	if got, want := gs.AggressiveExpansionLastTurns, 12; got != want {
		t.Fatalf("aşırı genişleme süresi = %d, %d bekleniyordu", got, want)
	}
}

func Test1300TradeCenterRelationsSeedAIMerchantAssignments(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	if got, want := gs.Regions["venice"].TradeCapacity, 16; got != want {
		t.Fatalf("Venedik merkez kapasitesi = %d, want %d", got, want)
	}
	if got, want := diplomacy.TradePartnerLimit(gs, faction.FactionID("genoa")), 8; got < want {
		t.Fatalf("Ceneviz merkez partner limiti = %d, want at least %d", got, want)
	}

	for _, fid := range []faction.FactionID{"venice", "genoa"} {
		foundRoute := false
		for _, route := range gs.TradeRoutes {
			if route != nil && route.FromFactionID == string(fid) && route.SuspendedTurns <= 0 {
				foundRoute = true
				break
			}
		}
		if !foundRoute {
			t.Fatalf("%s için başlangıç ticaret merkezi bağlantılı rota oluşturulmadı", fid)
		}

		ai.TakeTurn(gs, fid)
		assigned := false
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval || fleet.TradeRouteKey == "" {
				continue
			}
			for _, unit := range fleet.Units {
				if unit.TypeID == "merchant_ship" {
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		if !assigned {
			t.Fatalf("%s AI merchant filosuna başlangıç ticaret rotası atamadı", fid)
		}
	}
}

func Test1300LandTradeCenterRouteDoesNotAcceptMerchantFleet(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)

	var castileGranada *economy.TradeRoute
	for _, route := range gs.TradeRoutes {
		if route != nil && route.FromFactionID == "castile_kingdom" && route.ToFactionID == "granada_emirate" {
			castileGranada = route
			break
		}
	}
	if castileGranada == nil {
		t.Fatal("Kastilya-Gırnata ticaret rotası bulunamadı")
	}
	if got, ok := gs.MerchantTradeRouteTargetSeaRegion(castileGranada); ok || got != "" {
		t.Fatalf("Kastilya-Gırnata kara rotası deniz hedefi = (%q, %v), want boş", got, ok)
	}

	constantinopleAleppo := &economy.TradeRoute{FromFactionID: "east_rome", ToFactionID: "mamluk"}
	if got, ok := gs.MerchantTradeRouteTargetSeaRegion(constantinopleAleppo); ok || got != "" {
		t.Fatalf("Konstantiniyye-Halep doğrudan kara rotası deniz hedefi = (%q, %v), want boş", got, ok)
	}

	ai.TakeTurn(gs, faction.FactionID("castile_kingdom"))
	for _, fleet := range gs.Armies {
		if fleet == nil || fleet.OwnerID != "castile_kingdom" || !fleet.IsNaval || fleet.TradeRouteKey != castileGranada.AssignmentKey() {
			continue
		}
		for _, unit := range fleet.Units {
			if unit.TypeID == "merchant_ship" {
				t.Fatalf("Kastilya merchant filosuna kara rota atandı: filo=%s rota=%s", fleet.ID, fleet.TradeRouteKey)
			}
		}
	}
}

func Test1300MaritimeStatesStartWithTradeWeightedNavalComposition(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}

	type expectedComposition struct {
		warships  int
		merchants int
		deployed  int
	}
	expected := map[faction.FactionID]expectedComposition{
		"aragon":            {warships: 4, merchants: 2, deployed: 7},
		"aydin_bey":         {warships: 1, merchants: 1, deployed: 2},
		"candar_bey":        {warships: 1, merchants: 1, deployed: 2},
		"castile_kingdom":   {warships: 3, merchants: 2, deployed: 7},
		"cyprus_kingdom":    {warships: 2, merchants: 2, deployed: 4},
		"denmark_kingdom":   {warships: 1, merchants: 2, deployed: 3},
		"east_rome":         {warships: 1, merchants: 2, deployed: 8},
		"england":           {warships: 2, merchants: 3, deployed: 6},
		"flanders_county":   {warships: 1, merchants: 2, deployed: 3},
		"france":            {warships: 2, merchants: 2, deployed: 5},
		"florence_rep":      {warships: 0, merchants: 2, deployed: 2},
		"genoa":             {warships: 8, merchants: 6, deployed: 16},
		"granada_emirate":   {warships: 1, merchants: 2, deployed: 3},
		"hafsid_sultanate":  {warships: 1, merchants: 3, deployed: 4},
		"karesioglu_bey":    {warships: 2, merchants: 1, deployed: 3},
		"marinid_sultanate": {warships: 2, merchants: 3, deployed: 5},
		"mamluk":            {warships: 1, merchants: 3, deployed: 10},
		"mecca_sharifate":   {warships: 0, merchants: 1, deployed: 1},
		"mentese_bey":       {warships: 1, merchants: 1, deployed: 2},
		"naples_kingdom":    {warships: 2, merchants: 3, deployed: 6},
		"novgorod_rep":      {warships: 0, merchants: 1, deployed: 1},
		"portugal":          {warships: 2, merchants: 3, deployed: 6},
		"saruhan_bey":       {warships: 0, merchants: 1, deployed: 1},
		"trebizond_emp":     {warships: 1, merchants: 2, deployed: 3},
		"usfurid_emirate":   {warships: 0, merchants: 1, deployed: 1},
		"venice":            {warships: 16, merchants: 6, deployed: 24},
	}
	for fid, want := range expected {
		warships := 0
		merchants := 0
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval {
				continue
			}
			for _, unit := range fleet.Units {
				switch unit.TypeID {
				case "warship":
					warships++
				case "merchant_ship":
					merchants++
				}
			}
		}
		if warships != want.warships || merchants != want.merchants {
			t.Fatalf("%s başlangıç donanması = %d savaş, %d ticaret; want %d savaş, %d ticaret", fid, warships, merchants, want.warships, want.merchants)
		}
		if got := gs.DeployedNavalUnits(fid); got != want.deployed {
			t.Fatalf("%s başlangıç toplam gemi = %d, want %d", fid, got, want.deployed)
		}
		if got := gs.NavalCap(fid); got < want.deployed {
			t.Fatalf("%s başlangıç gemileri kapasiteyi aşıyor: %d/%d", fid, want.deployed, got)
		}
	}
}

func Test1300MaritimeMerchantFleetsHaveTradeRoutesForAI(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)

	merchantStates := make([]faction.FactionID, 0)
	for fid := range gs.Factions {
		hasMerchantFleet := false
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval {
				continue
			}
			for _, unit := range fleet.Units {
				if unit.TypeID == "merchant_ship" {
					hasMerchantFleet = true
					break
				}
			}
			if hasMerchantFleet {
				break
			}
		}
		if !hasMerchantFleet {
			continue
		}
		for _, route := range gs.TradeRoutes {
			if route != nil && route.FromFactionID == string(fid) && route.SuspendedTurns <= 0 && len(gs.MerchantTradeRouteSeaRegions(route)) > 0 {
				merchantStates = append(merchantStates, fid)
				break
			}
		}
	}
	sort.Slice(merchantStates, func(i, j int) bool { return merchantStates[i] < merchantStates[j] })
	for _, fid := range merchantStates {
		hasRoute := false
		for _, route := range gs.TradeRoutes {
			if route == nil || route.FromFactionID != string(fid) || route.SuspendedTurns > 0 {
				continue
			}
			if len(gs.MerchantTradeRouteSeaRegions(route)) > 0 {
				hasRoute = true
				break
			}
		}
		if !hasRoute {
			t.Fatalf("%s merchant filosu için başlangıçta kullanılabilir deniz ticaret rotası yok", fid)
		}

		ai.TakeTurn(gs, fid)
		assigned := false
		for _, fleet := range gs.Armies {
			if fleet == nil || fleet.OwnerID != string(fid) || !fleet.IsNaval || fleet.TradeRouteKey == "" {
				continue
			}
			for _, unit := range fleet.Units {
				if unit.TypeID == "merchant_ship" {
					assigned = true
					break
				}
			}
			if assigned {
				break
			}
		}
		if !assigned {
			t.Fatalf("%s AI merchant filosunu başlangıç ticaret rotasına atamadı", fid)
		}
	}
}

func Test1300OpeningEconomyCoversUpkeepAfterMerchantAssignments(t *testing.T) {
	gs, err := loadScenarioBaseState("1300_ottoman_rise", filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise"))
	if err != nil {
		t.Fatalf("loadScenarioBaseState() error = %v", err)
	}
	diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	for _, fid := range gs.FactionOrder {
		ai.AssignOpeningMerchantFleets(gs, fid)
	}

	const openingRunwayTurns = 10
	for _, fid := range gs.FactionOrder {
		regions := gs.RegionsOwnedBy(fid)
		if len(regions) == 0 {
			continue
		}
		owner := gs.Factions[fid]
		status := victory.GoldEconomyPreview(gs, fid)
		projectedGold := 0
		if owner != nil {
			projectedGold = owner.Gold + status.NetChange*openingRunwayTurns
		}
		if status.Income <= 0 || status.NetChange < 0 || projectedGold < 0 {
			t.Fatalf("%s açılış ekonomi dayanıklılığı yetersiz: net=%d, 10 tur sonrası hazine=%d (bölge=%d, gelir=%d, ordu bakımı=%d, bina bakımı=%d)",
				fid, status.NetChange, projectedGold, len(regions), status.Income, status.Upkeep, status.BuildingUpkeep)
		}
		t.Logf("%s açılış ekonomisi: net=%d, 10 tur sonrası hazine=%d, gelir=%d, ordu bakımı=%d, bina bakımı=%d",
			fid, status.NetChange, projectedGold, status.Income, status.Upkeep, status.BuildingUpkeep)
	}
}

func TestCampaignSaveStateDoesNotPersistScenarioDerivedStartYear(t *testing.T) {
	saved := campaignSaveState{
		Turn:       7,
		Year:       1306,
		ScenarioID: "1300_ottoman_rise",
	}

	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal campaign save state: %v", err)
	}
	if _, ok := fields["sy"]; ok {
		t.Fatal("scenario-derived start year compact save'e yazıldı")
	}
}

func TestCampaignSaveStatePreservesMinorPrivilegeRevocation(t *testing.T) {
	base := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: true}
	current := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: false}

	saved, ok := makeRegionSaveState(current, base)
	if !ok || saved.IsPrivileged == nil || *saved.IsPrivileged {
		t.Fatalf("imtiyaz kaldırma compact save'e yazılmadı: %+v", saved.IsPrivileged)
	}

	restored := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: true}
	applyRegionSaveState(restored, saved)
	if restored.IsPrivileged {
		t.Fatal("imtiyaz kaldırma compact save'den geri yüklenmedi")
	}
}

func TestCampaignSaveStatePreservesMinorPrivilegeGrantedTurn(t *testing.T) {
	base := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: true, PrivilegeGrantedTurn: 1}
	current := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: true, PrivilegeGrantedTurn: 7}

	saved, ok := makeRegionSaveState(current, base)
	if !ok || saved.PrivilegeGrantedTurn == nil || *saved.PrivilegeGrantedTurn != 7 {
		t.Fatalf("imtiyaz başlangıç turu compact save'e yazılmadı: %+v", saved.PrivilegeGrantedTurn)
	}
	restored := &world.Region{ID: "galata", IsMinorRegion: true, IsPrivileged: true, PrivilegeGrantedTurn: 1}
	applyRegionSaveState(restored, saved)
	if restored.PrivilegeGrantedTurn != 7 {
		t.Fatalf("imtiyaz başlangıç turu geri yüklenmedi: %d", restored.PrivilegeGrantedTurn)
	}
}

func TestCompactSavePreservesNavalSupplyCargo(t *testing.T) {
	armies := map[army.ArmyID]*army.Army{
		"fleet": {
			ID:      "fleet",
			OwnerID: "player",
			IsNaval: true,
			SupplyCargo: economy.ResourceCost{
				Grain: 240,
				Iron:  12,
			},
			NavalMission: &army.NavalMission{
				Kind:         army.NavalMissionSupplyArmy,
				TargetArmyID: "army",
			},
		},
	}

	saved := convertArmiesToSaveState(armies)
	restored := restoreArmiesFromSaveState(saved)["fleet"]
	if restored == nil {
		t.Fatal("ikmal filosu save'den geri yüklenmedi")
	}
	if restored.SupplyCargo.Grain != 240 || restored.SupplyCargo.Iron != 12 {
		t.Fatalf("ikmal kargosu korunmadı: %+v", restored.SupplyCargo)
	}
	if restored.NavalMission == nil || restored.NavalMission.TargetArmyID != "army" {
		t.Fatalf("ikmal görevi korunmadı: %+v", restored.NavalMission)
	}
}

func TestCompactSavePreservesDecisionSeedAndRecklessWarLedger(t *testing.T) {
	const key = "alpha|beta"
	saved := campaignSaveState{
		DecisionSeed:             987654321,
		DevelopmentMode:          true,
		DebugRevealMilitaryPower: true,
		WarLedgers: map[string]*state.WarLedger{
			key: {FactionA: "alpha", FactionB: "beta", RecklessDeclaration: true},
		},
	}
	gs := &state.GameState{}
	applyCampaignSaveState(gs, saved)
	if gs.DecisionSeed != saved.DecisionSeed {
		t.Fatalf("karar seed'i korunmadı: got=%d want=%d", gs.DecisionSeed, saved.DecisionSeed)
	}
	if !gs.DebugRevealMilitaryPower {
		t.Fatal("geliştirme güç görünümü save'den dönmedi")
	}
	if ledger := gs.WarLedgers[key]; ledger == nil || !ledger.RecklessDeclaration {
		t.Fatalf("riskli savaş işareti save'den dönmedi: %+v", ledger)
	}
}

func TestCompactSavePreservesRecentFactionExpansion(t *testing.T) {
	saved := campaignSaveState{
		Turn:       9,
		ScenarioID: "1300_ottoman_rise",
		RecentFactionExpansion: map[faction.FactionID]state.FactionExpansionRecord{
			"ai_rival": {
				WindowStartTurn: 7,
				RegionsGained:   4,
				TurnGains:       map[int]int{7: 2, 8: 2},
			},
		},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("genişleme kaydı marshal edilemedi: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("genişleme kaydı save'den çözülemedi: %v", err)
	}

	gs := &state.GameState{}
	applyCampaignSaveState(gs, decoded)
	got := gs.RecentFactionExpansion["ai_rival"]
	if got.WindowStartTurn != 7 || got.RegionsGained != 4 || got.TurnGains[7] != 2 || got.TurnGains[8] != 2 {
		t.Fatalf("yakın dönem genişleme kaydı korunmadı: %+v", got)
	}
}

func TestCompactSavePreservesRegionAttackRelationTrend(t *testing.T) {
	scoreAToB := 14
	scoreBToA := 23
	stance := encodeStance(faction.StancePeace)
	passiveAToB := -6
	passiveBToA := 4
	saved := campaignSaveState{
		Turn:               9,
		ScenarioID:         "1300_ottoman_rise",
		FactionAttackTurns: map[faction.FactionID]int{"attacker": 9},
		FactionAttackTargetTurns: map[faction.FactionID]map[faction.FactionID]int{
			"attacker": {"neighbor": 9},
		},
		RelationTrendAppliedTurn: 9,
		Relations: map[string]relationSaveState{
			"attacker|neighbor": {
				ScoreAToB:           &scoreAToB,
				ScoreBToA:           &scoreBToA,
				Stance:              &stance,
				PassiveModifierAToB: &passiveAToB,
				PassiveModifierBToA: &passiveBToA,
			},
		},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("ilişki trendi save'e yazılamadı: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("ilişki trendi save'den çözülemedi: %v", err)
	}

	gs := &state.GameState{}
	applyCampaignSaveState(gs, decoded)
	if gs.FactionAttackTurns["attacker"] != 9 || gs.FactionAttackTargetTurns["attacker"]["neighbor"] != 9 || gs.RelationTrendAppliedTurn != 9 {
		t.Fatalf("saldırı trend metadata'sı korunmadı: turns=%v targets=%v applied=%d", gs.FactionAttackTurns, gs.FactionAttackTargetTurns, gs.RelationTrendAppliedTurn)
	}
	rel := gs.Relations["attacker|neighbor"]
	if rel == nil || rel.ScoreFrom("attacker") != scoreAToB || rel.ScoreFrom("neighbor") != scoreBToA || rel.Stance != faction.StancePeace || rel.PassiveModifierFrom("attacker") != passiveAToB || rel.PassiveModifierFrom("neighbor") != passiveBToA {
		t.Fatalf("ilişki trendi korunmadı: %+v", rel)
	}
}

func TestCampaignSaveStateRestoresOtherIncomeDelta(t *testing.T) {
	const fid = faction.FactionID("portugal")
	base := &faction.Faction{ID: fid}
	current := &faction.Faction{ID: fid, OtherIncomeDelta: 25}

	savedFaction, ok := makeFactionSaveState(current, base)
	if !ok || savedFaction.OtherIncomeDelta == nil || *savedFaction.OtherIncomeDelta != 25 {
		t.Fatalf("other income delta compact save'e yazılmadı: %+v", savedFaction.OtherIncomeDelta)
	}

	restored := &faction.Faction{ID: fid}
	applyFactionSaveState(restored, savedFaction)
	if restored.OtherIncomeDelta != 25 {
		t.Fatalf("other income delta save'den yüklenmedi: %d", restored.OtherIncomeDelta)
	}
}

func TestCampaignSaveStateRefreshesSelectedVictoryFromScenario(t *testing.T) {
	option := scenario.VictoryOptionDef{
		ID:                   "updated_goal",
		Type:                 "economic",
		TargetGoldIncome:     777,
		GoldHoldTurns:        4,
		RequiredRegions:      []string{"bursa"},
		RequiredTradeCenters: []string{"venice"},
	}
	gs := &state.GameState{
		ScenarioVictories:       []scenario.VictoryOptionDef{option},
		Victory:                 state.VictoryCondition{Type: state.VictoryMilitary, TargetArmyStrength: 10},
		SelectedVictoryOptionID: "updated_goal",
	}
	saved := campaignSaveState{
		ScenarioID:              "1300_ottoman_rise",
		SelectedVictoryOptionID: "updated_goal",
		Victory:                 state.VictoryCondition{Type: state.VictoryMilitary, TargetArmyStrength: 10},
	}
	applyCampaignSaveState(gs, saved)

	if gs.Victory.Type != state.VictoryEconomic || gs.Victory.TargetGoldIncome != 777 || gs.Victory.GoldHoldTurns != 4 {
		t.Fatalf("güncel senaryo zaferi uygulanmadı: %+v", gs.Victory)
	}
	if len(gs.Victory.RequiredRegions) != 1 || gs.Victory.RequiredRegions[0] != "bursa" {
		t.Fatalf("güncel bölge hedefleri uygulanmadı: %+v", gs.Victory.RequiredRegions)
	}
	if len(gs.Victory.RequiredTradeCenters) != 1 || gs.Victory.RequiredTradeCenters[0] != "venice" {
		t.Fatalf("güncel ticaret merkezi hedefleri uygulanmadı: %+v", gs.Victory.RequiredTradeCenters)
	}
}

func TestCampaignSaveStateRestoresDismissedCommanderIDs(t *testing.T) {
	saved := campaignSaveState{
		ScenarioID:            "1300_ottoman_rise",
		DismissedCommanderIDs: map[string]bool{"commander_template": true},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("decode campaign save state: %v", err)
	}

	gs := &state.GameState{}
	applyCampaignSaveState(gs, decoded)
	if !gs.DismissedCommanderIDs["commander_template"] {
		t.Fatal("dismissed commander ID was not restored")
	}
}

func TestCampaignSaveStateRestoresChangedCapital(t *testing.T) {
	base := &faction.Faction{ID: "genoa", CapitalSettlementID: "capital_a"}
	current := *base
	current.CapitalSettlementID = "capital_b"
	delta, ok := makeFactionSaveState(&current, base)
	if !ok || delta.CapitalSettlementID == nil || *delta.CapitalSettlementID != "capital_b" {
		t.Fatalf("capital değişikliği save delta'sına yazılmadı: %+v", delta)
	}

	saved := campaignSaveState{
		ScenarioID: "1300_ottoman_rise",
		Factions:   map[faction.FactionID]factionSaveState{"genoa": delta},
	}
	payload, err := json.Marshal(saved)
	if err != nil {
		t.Fatalf("marshal campaign save state: %v", err)
	}
	decoded, err := decodeCampaignSaveState(payload)
	if err != nil {
		t.Fatalf("decode campaign save state: %v", err)
	}

	restored := &faction.Faction{ID: "genoa", CapitalSettlementID: "capital_a"}
	applyFactionSaveState(restored, decoded.Factions["genoa"])
	if restored.CapitalSettlementID != "capital_b" {
		t.Fatalf("yüklenen başkent = %q, want capital_b", restored.CapitalSettlementID)
	}
}
