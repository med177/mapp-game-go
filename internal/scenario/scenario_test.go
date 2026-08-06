package scenario

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
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

func TestCalendarMonthsPerTurnUsesScenarioValue(t *testing.T) {
	sc := Scenario{TurnMonths: 3}
	if got := sc.CalendarMonthsPerTurn(); got != 3 {
		t.Fatalf("üç aylık takvim ayarı kayboldu: %d", got)
	}
	if got := (Scenario{}).CalendarMonthsPerTurn(); got != 1 {
		t.Fatalf("eski senaryo uyumluluğu bir ay olmalı: %d", got)
	}
}

func TestScenarioProductionDurationsAreValidInData(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	scenariosRoot := filepath.Join(root, "assets", "scenarios")
	entries, err := os.ReadDir(scenariosRoot)
	if err != nil {
		t.Fatalf("senaryo dizini okunamadı: %v", err)
	}
	scenarioPaths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			scenarioPaths = append(scenarioPaths, filepath.Join(scenariosRoot, entry.Name()))
		}
	}
	sort.Strings(scenarioPaths)
	if len(scenarioPaths) == 0 {
		t.Fatal("test edilecek senaryo dizini bulunamadı")
	}

	for _, scenarioPath := range scenarioPaths {
		scenarioID := filepath.Base(scenarioPath)
		t.Run(scenarioID, func(t *testing.T) {
			dataPath := filepath.Join(scenarioPath, "data")

			buildings, err := city.LoadBuildings(filepath.Join(dataPath, "buildings.json"))
			if err != nil {
				t.Fatalf("%s binaları yüklenemedi: %v", scenarioID, err)
			}
			technologies, err := tech.LoadTechnologies(filepath.Join(dataPath, "technologies.json"))
			if err != nil {
				t.Fatalf("%s teknolojileri yüklenemedi: %v", scenarioID, err)
			}
			units, err := army.LoadUnitTypes(filepath.Join(dataPath, "units.json"))
			if err != nil {
				t.Fatalf("%s birlikleri yüklenemedi: %v", scenarioID, err)
			}
			for id, building := range buildings {
				if building.TurnsRequired <= 0 {
					t.Errorf("%s/%s bina üretim süresi pozitif olmalı: %d", scenarioID, id, building.TurnsRequired)
				}
			}
			for id, technology := range technologies {
				if technology.TurnsRequired <= 0 {
					t.Errorf("%s/%s araştırma süresi pozitif olmalı: %d", scenarioID, id, technology.TurnsRequired)
				}
			}
			for id, unit := range units {
				if unit.TurnsRequired <= 0 {
					t.Errorf("%s/%s birlik üretim süresi pozitif olmalı: %d", scenarioID, id, unit.TurnsRequired)
				}
			}
		})
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
	skipIfScenarioDirMissing(t, scenarioPath)

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

func skipIfScenarioDirMissing(t *testing.T, scenarioPath string) {
	t.Helper()

	info, err := os.Stat(scenarioPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("senaryo dizini mevcut değil: %s", scenarioPath)
		}
		t.Fatalf("senaryo dizini kontrol edilemedi: %v", err)
	}
	if !info.IsDir() {
		t.Skipf("senaryo yolu dizin değil: %s", scenarioPath)
	}
}

func Test1444HistoricalVictoryFactionsAreOnlyPlayableFactions(t *testing.T) {
	assertHistoricalVictoryFactionsMatchPlayableRoster(t, "1444_ottoman_empire")
}

func Test1512HistoricalVictoryFactionsAreOnlyPlayableFactions(t *testing.T) {
	assertHistoricalVictoryFactionsMatchPlayableRoster(t, "1512_yavuz_selim")
}
