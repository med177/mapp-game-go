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
}

type scenarioTempoProfile struct {
	name       string
	turns      int
	runs       int
	difficulty int
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
	level, ok := gs.AIDifficultyPolicy.Level(2)
	if !ok || !gs.AIDifficultyPolicy.FairMovement || level.PlanHorizonTurns != 6 || level.MinAttackPowerPercent != 115 {
		t.Fatalf("1300 AI zorluk politikası runtime state'e yüklenmedi: policy=%+v level=%+v", gs.AIDifficultyPolicy, level)
	}
}

func Test1300OttomanOpeningPlanUsesBithynianDirection(t *testing.T) {
	gs, _, err := loadScenarioData(scenario1300Path(t), 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	gs.PlayerFactionID = ""
	ai.TakeTurn(gs, "ottoman")

	plan := gs.AIPlans["ottoman"]
	if plan == nil || plan.ObjectiveID != "secure_bithynia" || plan.TargetFactionID != "east_rome" {
		t.Fatalf("Osmanlı açılış yönü Bitinya hattında kalmalıydı: %+v", plan)
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
	if normal.ReassessTurn-normal.StartedTurn != 6 || hard.ReassessTurn-hard.StartedTurn != 9 {
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
		simulateTempoTurns(t, gs, evts, turns, int64(seed), nil)
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
		agg.score = agg.regionGain*120 + agg.goldGain/8 + agg.endPower/25 + agg.endTrades*20 + agg.endTechs*18
		rows = append(rows, agg)
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
