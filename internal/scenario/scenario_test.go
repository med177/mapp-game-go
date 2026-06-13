package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"mapp-game-go/internal/faction"
)

func TestVictoryOptionRegionTargetsPrefersRequiredRegions(t *testing.T) {
	opt := VictoryOptionDef{
		RequiredRegions: []string{"constantinople", "ankara"},
	}

	got := opt.RegionTargets()
	if len(got) != 2 || got[0] != "constantinople" || got[1] != "ankara" {
		t.Fatalf("required_regions bekleniyordu, got=%v", got)
	}
}

func TestVictoryOptionRegionTargetsSkipsEmptyValues(t *testing.T) {
	opt := VictoryOptionDef{
		RequiredRegions: []string{"paris", "", "flanders"},
	}

	got := opt.RegionTargets()
	if len(got) != 2 || got[0] != "paris" || got[1] != "flanders" {
		t.Fatalf("bos degerler filtrelenmeliydi, got=%v", got)
	}
}

func TestFilterVictoryOptionsForFaction(t *testing.T) {
	options := []VictoryOptionDef{
		{ID: "shared"},
		{ID: "ottoman_only", AllowedFactions: []string{"ottoman"}},
		{ID: "rome_only", AllowedFactions: []string{"east_rome"}},
	}

	got := FilterVictoryOptionsForFaction(options, "ottoman")
	if len(got) != 2 {
		t.Fatalf("2 sonuc bekleniyordu, got=%d", len(got))
	}
	if got[0].ID != "shared" || got[1].ID != "ottoman_only" {
		t.Fatalf("beklenmeyen filtre sonucu: %+v", got)
	}
}

func assertHistoricalVictoryFactionsMatchPlayableRoster(t *testing.T, scenarioID string) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	scenarioPath := filepath.Join(root, "assets", "scenarios", scenarioID)

	var sc Scenario
	scenarioData, err := os.ReadFile(filepath.Join(scenarioPath, "scenario.json"))
	if err != nil {
		t.Fatalf("scenario.json okunamadi: %v", err)
	}
	if err := json.Unmarshal(scenarioData, &sc); err != nil {
		t.Fatalf("scenario.json parse edilemedi: %v", err)
	}

	var factions []faction.Faction
	factionsData, err := os.ReadFile(filepath.Join(scenarioPath, "data", "factions.json"))
	if err != nil {
		t.Fatalf("factions.json okunamadi: %v", err)
	}
	if err := json.Unmarshal(factionsData, &factions); err != nil {
		t.Fatalf("factions.json parse edilemedi: %v", err)
	}

	expected := make(map[string]struct{})
	for _, opt := range sc.VictoryConditions {
		for _, fid := range opt.AllowedFactions {
			if fid == "" {
				continue
			}
			expected[fid] = struct{}{}
		}
	}

	var got []string
	for _, f := range factions {
		if f.IsPlayable {
			got = append(got, string(f.ID))
		}
	}
	sort.Strings(got)

	want := make([]string, 0, len(expected))
	for fid := range expected {
		want = append(want, fid)
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("playable faction sayisi uyusmuyor: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("playable faction listesi uyusmuyor: got=%v want=%v", got, want)
		}
	}
}

func Test1444HistoricalVictoryFactionsAreOnlyPlayableFactions(t *testing.T) {
	assertHistoricalVictoryFactionsMatchPlayableRoster(t, "1444_ottoman_empire")
}

func Test1512HistoricalVictoryFactionsAreOnlyPlayableFactions(t *testing.T) {
	assertHistoricalVictoryFactionsMatchPlayableRoster(t, "1512_yavuz_selim")
}
