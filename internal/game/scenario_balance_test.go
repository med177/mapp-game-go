package game

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

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

func Test1300ScenarioTempoReport(t *testing.T) {
	if os.Getenv("RUN_SCENARIO_TEMPO_REPORT") != "1" {
		t.Skip("tempo raporu istege bagli; RUN_SCENARIO_TEMPO_REPORT=1 ile calistir")
	}

	const (
		turns = 42
		runs  = 8
	)

	scenarioPath := scenario1300Path(t)
	aggregates := map[faction.FactionID]*balanceAggregate{}

	for seed := 1; seed <= runs; seed++ {
		rand.Seed(int64(seed))
		gs, evts, err := loadScenarioData(scenarioPath, 2, nil)
		if err != nil {
			t.Fatalf("scenario load failed: %v", err)
		}
		gs.PlayerFactionID = ""

		start := snapshotAll(gs)
		simulateTempoTurns(gs, evts, turns)
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
	for _, agg := range aggregates {
		agg.startRegions /= runs
		agg.endRegions /= runs
		agg.regionGain /= runs
		agg.startGold /= runs
		agg.endGold /= runs
		agg.goldGain /= runs
		agg.endPower /= runs
		agg.endTechs /= runs
		agg.endTrades /= runs
		agg.score = agg.regionGain*120 + agg.goldGain/8 + agg.endPower/25 + agg.endTrades*20 + agg.endTechs*18
		rows = append(rows, agg)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].score == rows[j].score {
			return rows[i].endRegions > rows[j].endRegions
		}
		return rows[i].score > rows[j].score
	})

	t.Logf("1300 tempo report: turns=%d runs=%d", turns, runs)
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

func scenario1300Path(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")
}

func simulateTempoTurns(gs *state.GameState, evts []*events.Event, turns int) {
	g := &Game{gs: gs}
	for i := 0; i < turns; i++ {
		for _, fid := range sortedFactionIDs(gs) {
			f := gs.Factions[fid]
			if f == nil || f.IsEliminated {
				continue
			}
			ai.TakeTurn(gs, fid)
		}
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
	}
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
