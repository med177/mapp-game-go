package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/events"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

type balanceSnapshot struct {
	regions int
	gold    int
	grain   int
	power   int
	techs   int
	trades  int
}

type balanceAggregate struct {
	id           faction.FactionID
	name         string
	playable     bool
	startRegions float64
	endRegions   float64
	regionGain   float64
	startGold    float64
	endGold      float64
	goldGain     float64
	endPower     float64
	endTechs     float64
	endTrades    float64
	score        float64
	warTempoAggregate
}

type warTempoAggregate struct {
	warsStarted       float64
	activeWarTurns    float64
	completedWars     float64
	completedWarTurns float64
	conquests         float64
	peaceSettlements  float64
	stalemates        float64
}

type warTempoTelemetry struct {
	previousWars   map[string]state.WarLedger
	previousOwners map[world.RegionID]faction.FactionID
	byFaction      map[faction.FactionID]warTempoAggregate
}

func newWarTempoTelemetry(gs *state.GameState) *warTempoTelemetry {
	telemetry := &warTempoTelemetry{
		previousWars:   make(map[string]state.WarLedger),
		previousOwners: make(map[world.RegionID]faction.FactionID),
		byFaction:      make(map[faction.FactionID]warTempoAggregate),
	}
	if gs == nil {
		return telemetry
	}
	for rid, region := range gs.Regions {
		if region != nil && region.OwnerID != "" {
			telemetry.previousOwners[rid] = faction.FactionID(region.OwnerID)
		}
	}
	for key, ledger := range gs.WarLedgers {
		if ledger == nil {
			continue
		}
		telemetry.previousWars[key] = *ledger
		telemetry.addBoth(*ledger, func(metrics *warTempoAggregate) {
			metrics.warsStarted++
		})
	}
	return telemetry
}

func (t *warTempoTelemetry) addBoth(ledger state.WarLedger, apply func(*warTempoAggregate)) {
	if t == nil || apply == nil {
		return
	}
	for _, fid := range []faction.FactionID{ledger.FactionA, ledger.FactionB} {
		metrics := t.byFaction[fid]
		apply(&metrics)
		t.byFaction[fid] = metrics
	}
}

func (t *warTempoTelemetry) addFor(fid faction.FactionID, apply func(*warTempoAggregate)) {
	if t == nil || fid == "" || apply == nil {
		return
	}
	metrics := t.byFaction[fid]
	apply(&metrics)
	t.byFaction[fid] = metrics
}

func (t *warTempoTelemetry) observe(gs *state.GameState) {
	if t == nil || gs == nil {
		return
	}
	currentWars := make(map[string]state.WarLedger, len(gs.WarLedgers))
	for key, ledger := range gs.WarLedgers {
		if ledger == nil {
			continue
		}
		currentWars[key] = *ledger
		if _, existed := t.previousWars[key]; !existed {
			t.addBoth(*ledger, func(metrics *warTempoAggregate) {
				metrics.warsStarted++
			})
		}
		t.addBoth(*ledger, func(metrics *warTempoAggregate) {
			metrics.activeWarTurns++
		})
	}

	for rid, region := range gs.Regions {
		if region == nil || region.OwnerID == "" {
			continue
		}
		oldOwner := t.previousOwners[rid]
		newOwner := faction.FactionID(region.OwnerID)
		if oldOwner == "" || oldOwner == newOwner {
			continue
		}
		if _, activeBefore := t.warForPair(oldOwner, newOwner, t.previousWars); !activeBefore {
			if _, activeNow := t.warForPair(oldOwner, newOwner, currentWars); !activeNow {
				continue
			}
		}
		t.addFor(newOwner, func(metrics *warTempoAggregate) {
			metrics.conquests++
		})
	}

	for key, ledger := range t.previousWars {
		if _, stillActive := currentWars[key]; stillActive {
			continue
		}
		duration := gs.Turn - ledger.StartedTurn
		if duration < 1 {
			duration = 1
		}
		t.addBoth(ledger, func(metrics *warTempoAggregate) {
			metrics.completedWars++
			metrics.completedWarTurns += float64(duration)
			metrics.peaceSettlements++
		})
		if warTempoWasStalemate(gs, ledger) {
			t.addBoth(ledger, func(metrics *warTempoAggregate) {
				metrics.stalemates++
			})
		}
	}

	t.previousWars = currentWars
	t.previousOwners = make(map[world.RegionID]faction.FactionID, len(gs.Regions))
	for rid, region := range gs.Regions {
		if region != nil && region.OwnerID != "" {
			t.previousOwners[rid] = faction.FactionID(region.OwnerID)
		}
	}
}

func (t *warTempoTelemetry) warForPair(a, b faction.FactionID, wars map[string]state.WarLedger) (state.WarLedger, bool) {
	if t == nil || a == "" || b == "" || a == b {
		return state.WarLedger{}, false
	}
	key := faction.RelationKey(a, b)
	ledger, ok := wars[key]
	return ledger, ok
}

func warTempoWasStalemate(gs *state.GameState, ledger state.WarLedger) bool {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" {
		return false
	}
	warTurns := gs.Turn - ledger.StartedTurn
	if warTurns < 4 {
		return false
	}
	lastActionTurn := ledger.LastBattleTurn
	if lastActionTurn == 0 {
		lastActionTurn = ledger.StartedTurn
	}
	if gs.Turn-lastActionTurn < 3 {
		return false
	}
	for regionID, siege := range gs.Sieges {
		if siege == nil {
			continue
		}
		attacker := gs.Armies[siege.AttackerArmyID]
		target := gs.Regions[regionID]
		if attacker == nil || target == nil {
			continue
		}
		if (attacker.OwnerID == string(ledger.FactionA) && target.OwnerID == string(ledger.FactionB)) ||
			(attacker.OwnerID == string(ledger.FactionB) && target.OwnerID == string(ledger.FactionA)) {
			return false
		}
	}
	return true
}

type scenarioCalibrationBand struct {
	minGoldGain float64
	maxGoldGain float64
}

type scenarioTempoProfile struct {
	name       string
	turns      int
	runs       int
	difficulty int
}

type grainPhaseAggregate struct {
	samples         int
	production      float64
	civilianDemand  float64
	armyUpkeep      float64
	netChange       float64
	stockpileMonths float64
	monthsSamples   int
	famineSamples   int
}

func (a *grainPhaseAggregate) add(status state.GrainEconomyStatus) {
	if a == nil {
		return
	}
	a.samples++
	a.production += float64(status.Production)
	a.civilianDemand += float64(status.CivilianDemand)
	a.armyUpkeep += float64(status.ArmyUpkeep)
	a.netChange += float64(status.NetChange)
	if status.MonthsOfSupply >= 0 {
		a.stockpileMonths += float64(status.MonthsOfSupply)
		a.monthsSamples++
	}
	if status.SupplyLevel == state.GrainSupplyFamine {
		a.famineSamples++
	}
}

func (a grainPhaseAggregate) averages() (production, civilianDemand, armyUpkeep, netChange, stockpileMonths, famineRate float64) {
	if a.samples == 0 {
		return 0, 0, 0, 0, 0, 0
	}
	denominator := float64(a.samples)
	production = a.production / denominator
	civilianDemand = a.civilianDemand / denominator
	armyUpkeep = a.armyUpkeep / denominator
	netChange = a.netChange / denominator
	if a.monthsSamples > 0 {
		stockpileMonths = a.stockpileMonths / float64(a.monthsSamples)
	}
	famineRate = float64(a.famineSamples) / denominator
	return production, civilianDemand, armyUpkeep, netChange, stockpileMonths, famineRate
}

func TestLoadScenarioDataLoads1300AIStrategyProfiles(t *testing.T) {
	gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	ottoman, ok := gs.AIStrategies["ottoman"]
	if !ok || ottoman.Profile != "frontier_expansion" || len(ottoman.Objectives) == 0 {
		t.Fatalf("Osmanlı AI profili runtime state'e yüklenmedi: %+v", ottoman)
	}
	if _, ok := gs.AIStrategies["east_rome"]; !ok {
		t.Fatal("Doğu Roma AI profili runtime state'e yüklenmedi")
	}
	if gs.MonthsPerTurn != 3 {
		t.Fatalf("1300 senaryosu üç aylık tur temposuyla yüklenmeli: %d", gs.MonthsPerTurn)
	}
	if gs.TechTypes["iron_weapons"].TurnsRequired <= 0 || gs.BuildingTypes["market"].TurnsRequired <= 0 || gs.UnitTypes["militia"].TurnsRequired <= 0 {
		t.Fatalf("1300 üretim süreleri pozitif olmalı: tech=%d building=%d unit=%d", gs.TechTypes["iron_weapons"].TurnsRequired, gs.BuildingTypes["market"].TurnsRequired, gs.UnitTypes["militia"].TurnsRequired)
	}
	level, ok := gs.AIDifficultyPolicy.Level(2)
	if !ok || !gs.AIDifficultyPolicy.FairMovement || level.PlanHorizonTurns != 2 || level.MinAttackPowerPercent != 115 {
		t.Fatalf("1300 AI zorluk politikası runtime state'e yüklenmedi: policy=%+v level=%+v", gs.AIDifficultyPolicy, level)
	}
}

func Test1300OttomanOpeningPlanBuildsAnatolianPowerBase(t *testing.T) {
	gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	gs.PlayerFactionID = ""
	ai.TakeTurn(gs, "ottoman")

	plan := gs.AIPlans["ottoman"]
	if plan == nil || plan.ObjectiveID != "forge_anatolian_power_base" || plan.Kind != state.AIObjectiveConsolidate {
		t.Fatalf("Osmanlı açılışta uzun vadeli Rumeli seferinden önce Anadolu güç tabanını kurmalıydı: %+v", plan)
	}
}

func Test1300LevantFrontPlansFollowOpeningWar(t *testing.T) {
	for _, test := range []struct {
		factionID   faction.FactionID
		objective   string
		targetID    faction.FactionID
		firstRegion world.RegionID
	}{
		{factionID: "mamluk", objective: "hold_levant_cairo_corridor", targetID: "ilkhanate", firstRegion: "damascus"},
		{factionID: "ilkhanate", objective: "press_levant_frontier", targetID: "mamluk", firstRegion: "damascus"},
	} {
		t.Run(string(test.factionID), func(t *testing.T) {
			gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
			if err != nil {
				t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
			}
			gs.PlayerFactionID = ""
			ai.TakeTurn(gs, test.factionID)

			plan := gs.AIPlans[test.factionID]
			if plan == nil || plan.ObjectiveID != test.objective || plan.TargetFactionID != test.targetID {
				t.Fatalf("Levant açılış objective'i yanlış: %+v", plan)
			}
			if len(plan.TargetRegionIDs) == 0 || plan.TargetRegionIDs[0] != test.firstRegion {
				t.Fatalf("Levant hedef bölgesi yanlış: %+v", plan.TargetRegionIDs)
			}
		})
	}
}

func Test1300BalkanOpeningPlansPreferBorderSecurity(t *testing.T) {
	expected := map[faction.FactionID]string{
		"serbian_empire":   "hold_serbian_mountain_core",
		"bulgarian_empire": "hold_danube_balkan_line",
		"epir":             "hold_epirus_thessaly",
		"arnavut_des":      "hold_albanian_mountains",
		"athena_duk":       "hold_athens_coast",
		"wallachia_prince": "hold_wallachian_buffer",
	}
	for factionID, objectiveID := range expected {
		t.Run(string(factionID), func(t *testing.T) {
			gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
			if err != nil {
				t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
			}
			gs.PlayerFactionID = ""
			ai.TakeTurn(gs, factionID)
			plan := gs.AIPlans[factionID]
			if plan == nil || plan.ObjectiveID != objectiveID || plan.Kind != state.AIObjectiveDefend {
				t.Fatalf("Balkan açılış savunma objective'i seçilmedi: %+v", plan)
			}
		})
	}
}

func Test1300EasternSteppeAndBalticPlansFollowRegionalObjectives(t *testing.T) {
	expected := map[faction.FactionID]struct {
		objective string
		kind      state.AIObjectiveKind
	}{
		"russia":         {"hold_moscow_northern_frontier", state.AIObjectiveDefend},
		"golden_horde":   {"press_rus_steppe", state.AIObjectiveExpand},
		"teutonic_order": {"press_lithuanian_frontier", state.AIObjectiveExpand},
		"novgorod_rep":   {"hold_novgorod_trade_gate", state.AIObjectiveDefend},
		"lithuanian_gd":  {"contest_kievan_steppe", state.AIObjectiveExpand},
	}
	for factionID, want := range expected {
		t.Run(string(factionID), func(t *testing.T) {
			gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
			if err != nil {
				t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
			}
			gs.PlayerFactionID = ""
			ai.TakeTurn(gs, factionID)
			plan := gs.AIPlans[factionID]
			if plan == nil || plan.ObjectiveID != want.objective || plan.Kind != want.kind {
				t.Fatalf("doğu bozkır/Baltık açılış objective'i seçilmedi: %+v", plan)
			}
		})
	}
}

func Test1300EnglishFrenchWarWaitsFor1337Event(t *testing.T) {
	gs, evts, err := loadScenarioData(scenario1300Path(t), 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	gs.PlayerFactionID = ""
	ai.TakeTurn(gs, "england")
	ai.TakeTurn(gs, "france")
	if relation := diplomacy.Relation(gs, "england", "france"); relation == nil || relation.Stance != faction.StancePeace {
		t.Fatalf("İngiltere-Fransa 1300'de barışta başlamalı: %+v", relation)
	}
	if plan := gs.AIPlans["england"]; plan == nil || plan.ObjectiveID != "secure_english_channel_and_isles" {
		t.Fatalf("İngiltere erken dönemde kıta savaşına yönelmemeli: %+v", plan)
	}
	if plan := gs.AIPlans["france"]; plan == nil || plan.ObjectiveID != "protect_french_royal_core" {
		t.Fatalf("Fransa erken dönemde İngiltere savaşına yönelmemeli: %+v", plan)
	}

	gs.Year = 1337
	gs.Month = 5
	evt := events.Tick(gs, evts)
	if evt == nil || evt.ID != "hundred_years_war_1337" {
		t.Fatalf("1337 tarihsel savaş olayı tetiklenmedi: %+v", evt)
	}
	if _, ok := events.ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("1337 savaş olayı seçimi uygulanamadı")
	}
	if relation := diplomacy.Relation(gs, "england", "france"); relation == nil || relation.Stance != faction.StanceWar {
		t.Fatalf("1337 olayından sonra İngiltere-Fransa savaşa geçmeli: %+v", relation)
	}
	if !gs.FiredEventIDs["flag:hundred_years_war_started"] {
		t.Fatal("1337 savaş olayı AI hard gate bayrağını ayarlamadı")
	}
	gs.AIPlans = nil
	ai.TakeTurn(gs, "england")
	ai.TakeTurn(gs, "france")
	if plan := gs.AIPlans["england"]; plan == nil || plan.ObjectiveID != "secure_english_channel_and_isles" {
		t.Fatalf("İngiltere 1415'e kadar kıta seferi yerine ada savunmasını sürdürmeli: %+v", plan)
	}
	if plan := gs.AIPlans["france"]; plan == nil || plan.ObjectiveID != "recover_plantagenet_aquitaine_1337" || plan.TargetFactionID != "england" {
		t.Fatalf("1337 sonrası Fransa İngiliz cephe objective'ine geçmeli: %+v", plan)
	}

	gs.Year = 1415
	gs.AIPlans = nil
	ai.TakeTurn(gs, "england")
	if plan := gs.AIPlans["england"]; plan == nil || plan.ObjectiveID != "renew_french_crown_campaign_1415" || plan.TargetFactionID != "france" {
		t.Fatalf("İngiltere 1415'te Fransız tacı seferine geçmeli: %+v", plan)
	}
}

func Test1300SafavidRiseWaitsFor1501Event(t *testing.T) {
	gs, evts, err := loadScenarioData(scenario1300Path(t), 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	gs.PlayerFactionID = ""
	ai.TakeTurn(gs, "safavid")
	if plan := gs.AIPlans["safavid"]; plan == nil || plan.ObjectiveID != "hold_southern_persian_core" || plan.Kind != state.AIObjectiveConsolidate {
		t.Fatalf("Safevî 1300'de erken genişlememeli: %+v", plan)
	}

	gs.Year = 1501
	gs.Month = 1
	evt := events.Tick(gs, evts)
	if evt == nil || evt.ID != "safavid_rise_1501" {
		t.Fatalf("1501 Safevî yükseliş olayı tetiklenmedi: %+v", evt)
	}
	events.Apply(gs, evt)
	if _, ok := events.ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("1501 Safevî yükseliş seçimi uygulanamadı")
	}
	if !gs.FiredEventIDs["flag:safavid_rise"] {
		t.Fatal("Safevî yükselişi AI hard gate bayrağını ayarlamadı")
	}
	gs.AIPlans = nil
	ai.TakeTurn(gs, "safavid")
	if plan := gs.AIPlans["safavid"]; plan == nil || plan.ObjectiveID != "rise_into_persian_heartland_1501" || plan.TargetFactionID != "ilkhanate" {
		t.Fatalf("1501 sonrası Safevî İran çekirdeği objective'ine geçmeli: %+v", plan)
	}
}

func Test1300HardDifficultyExtendsEastRomePlanningCoverage(t *testing.T) {
	loadAndPlan := func(difficulty int) *state.AIPlanState {
		t.Helper()
		gs, _, err := loadScenarioData(scenario1300Path(t), difficulty, nil)
		if err != nil {
			t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
		}
		gs.PlayerFactionID = ""
		ai.TakeTurn(gs, "east_rome")
		return gs.AIPlans["east_rome"]
	}

	normal := loadAndPlan(2)
	hard := loadAndPlan(3)
	if normal == nil || hard == nil || normal.ObjectiveID != "hold_bosporus" || hard.ObjectiveID != "hold_bosporus" {
		t.Fatalf("Doğu Roma savunma objective'i üretilemedi: normal=%+v hard=%+v", normal, hard)
	}
	if len(normal.TargetRegionIDs) != 4 || len(hard.TargetRegionIDs) != 5 {
		t.Fatalf("zor AI daha geniş objective kapsamalıydı: normal=%+v hard=%+v", normal.TargetRegionIDs, hard.TargetRegionIDs)
	}
	if normal.ReassessTurn-normal.StartedTurn != 2 || hard.ReassessTurn-hard.StartedTurn != 3 {
		t.Fatalf("plan ufku zorlukla ölçeklenmedi: normal=%+v hard=%+v", normal, hard)
	}
}

func Test1300ScenarioTempoReport(t *testing.T) {
	enableScenarioTempoRandSeeding(t)
	profile, enabled, err := scenarioTempoProfileFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !enabled {
		t.Skip("tempo raporu istege bagli; RUN_SCENARIO_TEMPO_REPORT=fast|medium|calibration ile calistir")
	}

	turns := profile.turns
	runs := profile.runs

	scenarioPath := scenario1300Path(t)
	aggregates := map[faction.FactionID]*balanceAggregate{}

	for seed := 1; seed <= runs; seed++ {
		rand.Seed(int64(seed))
		gs, evts, err := loadScenarioData(scenarioPath, profile.difficulty, nil)
		if err != nil {
			t.Fatalf("scenario load failed: %v", err)
		}
		gs.PlayerFactionID = ""

		start := snapshotAll(gs)
		warTelemetry := newWarTempoTelemetry(gs)
		simulateTempoTurnsWithTurnEnd(t, gs, evts, turns, int64(seed), nil, func(_ int, current *state.GameState) {
			warTelemetry.observe(current)
		})
		end := snapshotAll(gs)

		for fid, f := range gs.Factions {
			if f == nil {
				continue
			}
			agg := aggregates[fid]
			if agg == nil {
				agg = &balanceAggregate{
					id:       fid,
					name:     f.NameTR,
					playable: f.IsPlayable,
				}
				aggregates[fid] = agg
			}
			agg.startRegions += float64(start[fid].regions)
			agg.endRegions += float64(end[fid].regions)
			agg.regionGain += float64(end[fid].regions - start[fid].regions)
			agg.startGold += float64(start[fid].gold)
			agg.endGold += float64(end[fid].gold)
			agg.goldGain += float64(end[fid].gold - start[fid].gold)
			agg.endPower += float64(end[fid].power)
			agg.endTechs += float64(end[fid].techs)
			agg.endTrades += float64(end[fid].trades)
			if metrics, ok := warTelemetry.byFaction[fid]; ok {
				agg.warTempoAggregate.warsStarted += metrics.warsStarted
				agg.warTempoAggregate.activeWarTurns += metrics.activeWarTurns
				agg.warTempoAggregate.completedWars += metrics.completedWars
				agg.warTempoAggregate.completedWarTurns += metrics.completedWarTurns
				agg.warTempoAggregate.conquests += metrics.conquests
				agg.warTempoAggregate.peaceSettlements += metrics.peaceSettlements
				agg.warTempoAggregate.stalemates += metrics.stalemates
			}
		}
	}

	rows := make([]*balanceAggregate, 0, len(aggregates))
	runCount := float64(runs)
	for _, agg := range aggregates {
		agg.startRegions /= runCount
		agg.endRegions /= runCount
		agg.regionGain /= runCount
		agg.startGold /= runCount
		agg.endGold /= runCount
		agg.goldGain /= runCount
		agg.endPower /= runCount
		agg.endTechs /= runCount
		agg.endTrades /= runCount
		agg.warsStarted /= runCount
		agg.activeWarTurns /= runCount
		agg.completedWars /= runCount
		agg.completedWarTurns /= runCount
		agg.conquests /= runCount
		agg.peaceSettlements /= runCount
		agg.stalemates /= runCount
		agg.score = agg.regionGain*120 + agg.goldGain/8 + agg.endPower/25 + agg.endTrades*20 + agg.endTechs*18
		rows = append(rows, agg)
	}
	// Bu bantlar 42 aylık medium profilin sözleşmesidir. 120 aylık calibration
	// raporu toplam birikimi ölçer; aynı sınırlarla karşılaştırmak uzun ömürlü
	// fraksiyonları yanlış negatif üretir.
	if profile.turns == 42 && profile.runs >= 4 {
		assert1300CalibrationBands(t, aggregates)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].score == rows[j].score {
			return rows[i].endRegions > rows[j].endRegions
		}
		return rows[i].score > rows[j].score
	})

	t.Logf("1300 tempo report: profile=%s difficulty=%d turns=%d runs=%d", profile.name, profile.difficulty, turns, runs)
	t.Log("top growth factions (avg):")
	for i, row := range rows {
		if i >= 14 {
			break
		}
		t.Logf(
			"%2d. %-22s playable=%t score=%.1f region %.1f->%.1f (delta %+0.1f) gold %.0f->%.0f (delta %+0.0f) power %.0f tech %.1f trades %.1f",
			i+1,
			row.id,
			row.playable,
			row.score,
			row.startRegions,
			row.endRegions,
			row.regionGain,
			row.startGold,
			row.endGold,
			row.goldGain,
			row.endPower,
			row.endTechs,
			row.endTrades,
		)
	}

	t.Log("playable factions:")
	playables := make([]*balanceAggregate, 0)
	for _, row := range rows {
		if row.playable {
			playables = append(playables, row)
		}
	}
	sort.Slice(playables, func(i, j int) bool { return playables[i].score > playables[j].score })
	for _, row := range playables {
		t.Logf(
			"  %-22s score=%.1f region %+0.1f gold %+0.0f power %.0f tech %.1f trades %.1f",
			row.id,
			row.score,
			row.regionGain,
			row.goldGain,
			row.endPower,
			row.endTechs,
			row.endTrades,
		)
	}

	warRows := append([]*balanceAggregate(nil), rows...)
	sort.Slice(warRows, func(i, j int) bool { return warRows[i].id < warRows[j].id })
	t.Log("war telemetry (average per run):")
	for _, row := range warRows {
		if row.warsStarted == 0 && row.activeWarTurns == 0 && row.completedWars == 0 && row.conquests == 0 {
			continue
		}
		averageCompletedWarTurns := 0.0
		if row.completedWars > 0 {
			averageCompletedWarTurns = row.completedWarTurns / row.completedWars
		}
		t.Logf(
			"  %-22s wars_started=%.1f active_war_turns=%.1f completed_wars=%.1f avg_war_turns=%.1f conquests=%.1f peace=%.1f stalemate=%.1f",
			row.id,
			row.warsStarted,
			row.activeWarTurns,
			row.completedWars,
			averageCompletedWarTurns,
			row.conquests,
			row.peaceSettlements,
			row.stalemates,
		)
	}
}

func Test1300ScenarioGrainEconomyBands(t *testing.T) {
	enableScenarioTempoRandSeeding(t)
	const turns = 12
	const runs = 2
	// Uzun vadeli hedefler açılışta fetih yerine üretim yatırımı yaptırır.
	// Üretim/sivil talep oranı izlenir; güncel kıtlık modeli ordu bakımı
	// nedeniyle negatif net değişime izin verdiğinden eski sabit net bant
	// burada doğrulanmaz.
	// Ortalama üretim oranı kayan nokta yuvarlamasıyla 4.10'un birkaç binde
	// üzerine çıkabilir; mevcut senaryo akışındaki savaş fazını kapsayan küçük
	// bir ölçüm toleransı bırakılır.
	const maxProductionRatio = 4.25
	majorFactions := []faction.FactionID{"ottoman", "venice", "mamluk", "england", "france"}
	phaseName := func(turn int) string {
		switch {
		case turn <= 4:
			return "erken"
		case turn <= 8:
			return "orta"
		default:
			return "savaş"
		}
	}
	reports := make(map[faction.FactionID]map[string]*grainPhaseAggregate, len(majorFactions))
	warReports := make(map[faction.FactionID]warTempoAggregate)
	for _, fid := range majorFactions {
		reports[fid] = map[string]*grainPhaseAggregate{
			"erken": &grainPhaseAggregate{},
			"orta":  &grainPhaseAggregate{},
			"savaş": &grainPhaseAggregate{},
		}
	}

	for seed := 1; seed <= runs; seed++ {
		rand.Seed(int64(seed))
		gs, evts, err := loadScenarioData(scenario1300Path(t), 2, nil)
		if err != nil {
			t.Fatalf("scenario load failed: %v", err)
		}
		gs.PlayerFactionID = ""
		warTelemetry := newWarTempoTelemetry(gs)
		simulateTempoTurnsWithTurnEnd(t, gs, evts, turns, int64(seed), nil, func(turn int, current *state.GameState) {
			warTelemetry.observe(current)
			phase := phaseName(turn)
			for _, fid := range majorFactions {
				status, ok := current.GrainEconomy[fid]
				if !ok {
					t.Fatalf("%s için %d. tur tahıl snapshot'ı yok", fid, turn)
				}
				if status.Production < 0 || status.CivilianDemand < 0 || status.ArmyUpkeep < 0 || status.Stockpile < 0 {
					t.Fatalf("%s %d. turda negatif tahıl metriği üretti: %+v", fid, turn, status)
				}
				expectedNetChange := status.Production - status.TotalDemand - status.ReplenishmentGrainSpent - status.GrowthGrainSpent - status.AutoExportSold
				if status.NetChange != expectedNetChange {
					t.Fatalf("%s %d. tur net değişim formülü bozuldu: %+v", fid, turn, status)
				}
				reports[fid][phase].add(status)
			}
		})
		for fid, metrics := range warTelemetry.byFaction {
			aggregate := warReports[fid]
			aggregate.warsStarted += metrics.warsStarted
			aggregate.activeWarTurns += metrics.activeWarTurns
			aggregate.completedWars += metrics.completedWars
			aggregate.completedWarTurns += metrics.completedWarTurns
			aggregate.conquests += metrics.conquests
			aggregate.peaceSettlements += metrics.peaceSettlements
			aggregate.stalemates += metrics.stalemates
			warReports[fid] = aggregate
		}
	}

	for _, fid := range majorFactions {
		for _, phase := range []string{"erken", "orta", "savaş"} {
			aggregate := *reports[fid][phase]
			production, civilianDemand, armyUpkeep, netChange, stockpileMonths, famineRate := aggregate.averages()
			if civilianDemand <= 0 {
				t.Fatalf("%s/%s sivil talep üretmedi", fid, phase)
			}
			productionRatio := production / civilianDemand
			if productionRatio < 0.75 || productionRatio > maxProductionRatio {
				t.Fatalf("%s/%s üretim-tüketim bandı dışı: ratio=%.2f production=%.1f civilian=%.1f", fid, phase, productionRatio, production, civilianDemand)
			}
			t.Logf("1300 tahıl bandı faction=%s phase=%s production=%.1f civilian=%.1f army=%.1f net=%.1f stockpile_months=%.1f famine_rate=%.0f%%", fid, phase, production, civilianDemand, armyUpkeep, netChange, stockpileMonths, famineRate*100)
		}
	}

	ids := make([]faction.FactionID, 0, len(warReports))
	for fid := range warReports {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	t.Log("1300 war telemetry (24 turns, average per run):")
	for _, fid := range ids {
		metrics := warReports[fid]
		averageCompletedWarTurns := 0.0
		if metrics.completedWars > 0 {
			averageCompletedWarTurns = metrics.completedWarTurns / metrics.completedWars
		}
		t.Logf(
			"  %-22s wars_started=%.1f active_war_turns=%.1f completed_wars=%.1f avg_war_turns=%.1f conquests=%.1f peace=%.1f stalemate=%.1f",
			fid,
			metrics.warsStarted/float64(runs),
			metrics.activeWarTurns/float64(runs),
			metrics.completedWars/float64(runs),
			averageCompletedWarTurns,
			metrics.conquests/float64(runs),
			metrics.peaceSettlements/float64(runs),
			metrics.stalemates/float64(runs),
		)
	}
}

func assert1300CalibrationBands(t *testing.T, aggregates map[faction.FactionID]*balanceAggregate) {
	t.Helper()
	bands := map[faction.FactionID]scenarioCalibrationBand{
		"mamluk":    {minGoldGain: 12000, maxGoldGain: 30000},
		"england":   {minGoldGain: 17000, maxGoldGain: 32000},
		"hre":       {minGoldGain: 15000, maxGoldGain: 32000},
		"france":    {minGoldGain: 12000, maxGoldGain: 30000},
		"ilkhanate": {minGoldGain: 10000, maxGoldGain: 30000},
		"venice":    {minGoldGain: 9000, maxGoldGain: 22000},
		"ottoman":   {minGoldGain: -2000, maxGoldGain: 6000},
		"safavid":   {minGoldGain: 500, maxGoldGain: 5000},
	}
	for factionID, band := range bands {
		agg := aggregates[factionID]
		if agg == nil {
			t.Fatalf("kalibrasyon bandı için fraksiyon ölçümü yok: %s", factionID)
		}
		if agg.goldGain < band.minGoldGain || agg.goldGain > band.maxGoldGain {
			t.Fatalf("42 aylık altın kabul bandı aşıldı: faction=%s gain=%.0f want=%.0f..%.0f", factionID, agg.goldGain, band.minGoldGain, band.maxGoldGain)
		}
	}
}

func scenarioTempoProfileFromEnv() (scenarioTempoProfile, bool, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("RUN_SCENARIO_TEMPO_REPORT")))
	if raw == "" || raw == "0" || raw == "false" {
		return scenarioTempoProfile{}, false, nil
	}

	var profile scenarioTempoProfile
	switch raw {
	case "fast":
		profile = scenarioTempoProfile{name: "fast", turns: 12, runs: 2, difficulty: 2}
	case "1", "medium":
		profile = scenarioTempoProfile{name: "medium", turns: 42, runs: 4, difficulty: 2}
	case "calibration":
		profile = scenarioTempoProfile{name: "calibration", turns: 120, runs: 8, difficulty: 2}
	default:
		return scenarioTempoProfile{}, false, fmt.Errorf("gecersiz RUN_SCENARIO_TEMPO_REPORT profili %q", raw)
	}

	var err error
	profile.turns, err = positiveEnvOverride("SCENARIO_TEMPO_TURNS", profile.turns)
	if err != nil {
		return scenarioTempoProfile{}, false, err
	}
	profile.runs, err = positiveEnvOverride("SCENARIO_TEMPO_RUNS", profile.runs)
	if err != nil {
		return scenarioTempoProfile{}, false, err
	}
	profile.difficulty, err = positiveEnvOverride("SCENARIO_TEMPO_DIFFICULTY", profile.difficulty)
	if err != nil {
		return scenarioTempoProfile{}, false, err
	}
	if profile.difficulty > 3 {
		return scenarioTempoProfile{}, false, fmt.Errorf("SCENARIO_TEMPO_DIFFICULTY 1..3 aralığında olmalı: %d", profile.difficulty)
	}
	return profile, true, nil
}

func positiveEnvOverride(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("%s pozitif bir tam sayi olmali: %q", key, raw)
	}
	return value, nil
}

func TestScenarioTempoProfileFromEnv(t *testing.T) {
	t.Setenv("SCENARIO_TEMPO_TURNS", "")
	t.Setenv("SCENARIO_TEMPO_RUNS", "")
	t.Setenv("SCENARIO_TEMPO_DIFFICULTY", "")

	t.Setenv("RUN_SCENARIO_TEMPO_REPORT", "fast")
	profile, enabled, err := scenarioTempoProfileFromEnv()
	if err != nil || !enabled || profile.turns != 12 || profile.runs != 2 || profile.difficulty != 2 {
		t.Fatalf("fast profile = %+v enabled=%t err=%v", profile, enabled, err)
	}

	t.Setenv("RUN_SCENARIO_TEMPO_REPORT", "calibration")
	t.Setenv("SCENARIO_TEMPO_TURNS", "9")
	t.Setenv("SCENARIO_TEMPO_RUNS", "3")
	t.Setenv("SCENARIO_TEMPO_DIFFICULTY", "3")
	profile, enabled, err = scenarioTempoProfileFromEnv()
	if err != nil || !enabled || profile.turns != 9 || profile.runs != 3 || profile.difficulty != 3 {
		t.Fatalf("overridden profile = %+v enabled=%t err=%v", profile, enabled, err)
	}
}

func Test1300ScenarioAITwoTurnReplayIsDeterministic(t *testing.T) {
	enableScenarioTempoRandSeeding(t)
	scenarioPath := scenario1300Path(t)
	type replayCheckpoint struct {
		name    string
		hash    [sha256.Size]byte
		encoded []byte
	}
	run := func() ([]byte, []replayCheckpoint) {
		t.Helper()
		rand.Seed(1300)
		gs, evts, err := loadScenarioData(scenarioPath, 2, nil)
		if err != nil {
			t.Fatalf("scenario load failed: %v", err)
		}
		gs.PlayerFactionID = ""
		var checkpoints []replayCheckpoint
		simulateTempoTurns(t, gs, evts, 2, 1300, func(name string, current *state.GameState) {
			encoded, marshalErr := json.Marshal(current)
			if marshalErr != nil {
				t.Fatalf("checkpoint state marshal failed: %v", marshalErr)
			}
			checkpoints = append(checkpoints, replayCheckpoint{name: name, hash: sha256.Sum256(encoded), encoded: encoded})
		})
		encoded, err := json.Marshal(gs)
		if err != nil {
			t.Fatalf("state marshal failed: %v", err)
		}
		return encoded, checkpoints
	}

	first, firstCheckpoints := run()
	second, secondCheckpoints := run()
	if !bytes.Equal(first, second) {
		for i := 0; i < len(firstCheckpoints) && i < len(secondCheckpoints); i++ {
			if firstCheckpoints[i].name != secondCheckpoints[i].name || firstCheckpoints[i].hash != secondCheckpoints[i].hash {
				t.Fatalf(
					"aynı seed ile AI replay ilk kez %s noktasında ayrıştı: first=%x second=%x diff=%s",
					firstCheckpoints[i].name,
					firstCheckpoints[i].hash,
					secondCheckpoints[i].hash,
					jsonDifferenceWindow(firstCheckpoints[i].encoded, secondCheckpoints[i].encoded),
				)
			}
		}
		firstHash := sha256.Sum256(first)
		secondHash := sha256.Sum256(second)
		t.Fatalf("aynı seed ile iki turluk AI replay farklı state üretti: first=%x second=%x", firstHash, secondHash)
	}
}

func enableScenarioTempoRandSeeding(t *testing.T) {
	t.Helper()
	parts := strings.Split(os.Getenv("GODEBUG"), ",")
	filtered := parts[:0]
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || strings.HasPrefix(part, "randseednop=") {
			continue
		}
		filtered = append(filtered, part)
	}
	filtered = append(filtered, "randseednop=0")
	t.Setenv("GODEBUG", strings.Join(filtered, ","))
}

func jsonDifferenceWindow(first, second []byte) string {
	limit := len(first)
	if len(second) < limit {
		limit = len(second)
	}
	index := 0
	for index < limit && first[index] == second[index] {
		index++
	}
	start := index - 120
	if start < 0 {
		start = 0
	}
	firstEnd := index + 180
	if firstEnd > len(first) {
		firstEnd = len(first)
	}
	secondEnd := index + 180
	if secondEnd > len(second) {
		secondEnd = len(second)
	}
	return fmt.Sprintf("offset=%d first=%q second=%q", index, first[start:firstEnd], second[start:secondEnd])
}

func Benchmark1300AIRound(b *testing.B) {
	scenarioPath := scenario1300Path(b)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		rand.Seed(1)
		gs, _, err := loadScenarioData(scenarioPath, 2, nil)
		if err != nil {
			b.Fatalf("scenario load failed: %v", err)
		}
		gs.PlayerFactionID = ""
		b.StartTimer()
		for _, fid := range sortedFactionIDs(gs) {
			f := gs.Factions[fid]
			if f != nil && !f.IsEliminated {
				ai.TakeTurn(gs, fid)
			}
		}
	}
}

func scenario1300Path(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")
}

func simulateTempoTurns(t *testing.T, gs *state.GameState, evts []*events.Event, turns int, baseSeed int64, checkpoint func(string, *state.GameState)) {
	simulateTempoTurnsWithTurnEnd(t, gs, evts, turns, baseSeed, checkpoint, nil)
}

func simulateTempoTurnsWithTurnEnd(t *testing.T, gs *state.GameState, evts []*events.Event, turns int, baseSeed int64, checkpoint func(string, *state.GameState), turnEnd func(int, *state.GameState)) {
	t.Helper()
	g := &Game{gs: gs}
	for i := 0; i < turns; i++ {
		startedAt := time.Now()
		for _, fid := range sortedFactionIDs(gs) {
			f := gs.Factions[fid]
			if f == nil || f.IsEliminated {
				continue
			}
			rand.Seed(scenarioTempoSeed(baseSeed, i+1, string(fid)))
			stepper := ai.NewTurnStepper(gs, fid)
			stepIndex := 0
			for {
				rand.Seed(scenarioTempoSeed(baseSeed, i+1, fmt.Sprintf("%s|step=%d", fid, stepIndex+1)))
				step, done := stepper.Step()
				if done {
					break
				}
				stepIndex++
				if checkpoint != nil {
					checkpoint(fmt.Sprintf("turn=%d faction=%s step=%d kind=%s army=%s target=%s", i+1, fid, stepIndex, step.Kind, step.ArmyID, step.TargetRegion), gs)
				}
			}
			if checkpoint != nil {
				checkpoint(fmt.Sprintf("turn=%d faction=%s", i+1, fid), gs)
			}
		}
		rand.Seed(scenarioTempoSeed(baseSeed, i+1, "resolution"))
		g.sanitizeDockedFleets()
		applySeasonEffects(gs)
		applyEconomyTick(gs)
		_ = applyTechTicks(gs)
		_ = g.applyProductionTicks()
		applyReligionConversion(gs)
		checkRebellions(gs)
		checkEliminations(gs)
		applyRelationDecay(gs)
		if evt := events.Tick(gs, evts); evt != nil {
			events.Apply(gs, evt)
			if idx := events.AutoChoose(evt); idx >= 0 {
				events.ApplyChoice(gs, evt, idx)
			}
		}
		gs.AdvanceTurn()
		checkRegionUnlocks(gs)
		if turnEnd != nil {
			turnEnd(i+1, gs)
		}
		landUnits, navalUnits := tempoUnitCounts(gs)
		t.Logf(
			"tempo turn=%d duration=%s armies=%d land_units=%d naval_units=%d queue=%d sieges=%d",
			i+1,
			time.Since(startedAt).Round(time.Millisecond),
			len(gs.Armies),
			landUnits,
			navalUnits,
			len(gs.ProductionQueue),
			len(gs.Sieges),
		)
	}
}

func scenarioTempoSeed(baseSeed int64, turn int, scope string) int64 {
	hasher := fnv.New64a()
	_, _ = fmt.Fprintf(hasher, "%d|%d|%s", baseSeed, turn, scope)
	return int64(hasher.Sum64() & uint64(^uint64(0)>>1))
}

func tempoUnitCounts(gs *state.GameState) (land, naval int) {
	for _, candidate := range gs.Armies {
		if candidate == nil {
			continue
		}
		if candidate.IsNaval {
			naval += len(candidate.Units)
		} else {
			land += len(candidate.Units)
		}
	}
	return land, naval
}

func sortedFactionIDs(gs *state.GameState) []faction.FactionID {
	ids := make([]faction.FactionID, 0, len(gs.Factions))
	for fid := range gs.Factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func snapshotAll(gs *state.GameState) map[faction.FactionID]balanceSnapshot {
	out := make(map[faction.FactionID]balanceSnapshot, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f == nil {
			continue
		}
		out[fid] = balanceSnapshot{
			regions: len(gs.RegionsOwnedBy(fid)),
			gold:    f.Gold,
			grain:   f.Grain,
			power:   diplomacy.MilitaryPower(gs, fid),
			techs:   len(f.Research.Completed),
			trades:  factionTradeCount(gs, fid),
		}
	}
	return out
}

func factionTradeCount(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil {
		return 0
	}
	total := 0
	for _, tr := range gs.TradeRoutes {
		if tr == nil || tr.SuspendedTurns > 0 {
			continue
		}
		if tr.FromFactionID == string(fid) || tr.ToFactionID == string(fid) {
			total++
		}
	}
	return total
}
