package game

import (
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/scenario"
)

func TestLoad1455WarsOfTheRosesScenario(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	scenarioPath := filepath.Join(root, "assets", "scenarios", "1455_wars_of_the_roses")
	scenarios, err := scenario.LoadAll(filepath.Join(root, "assets", "scenarios"))
	if err != nil {
		t.Fatalf("senaryo index'i yüklenemedi: %v", err)
	}
	previousScenarioList := render.ScenarioList
	render.ScenarioList = scenarios
	t.Cleanup(func() { render.ScenarioList = previousScenarioList })

	gs, evts, err := loadScenarioData(scenarioPath, 2, nil)
	if err != nil {
		t.Fatalf("Güller Savaşı senaryosu yüklenemedi: %v", err)
	}
	if gs.ScenarioID != "1455_wars_of_the_roses" || gs.Year != 1455 || gs.Month != 5 {
		t.Fatalf("senaryo başlangıç state'i yanlış: id=%q tarih=%d/%d", gs.ScenarioID, gs.Year, gs.Month)
	}
	if len(gs.Regions) == 0 || len(gs.Factions) == 0 || len(evts) == 0 {
		t.Fatalf("senaryo verisi eksik yüklendi: regions=%d factions=%d events=%d", len(gs.Regions), len(gs.Factions), len(evts))
	}
}

func TestLoad1300LoadsImperialState(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	scenarioPath := filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")

	gs, _, err := loadScenarioData(scenarioPath, 2, nil)
	if err != nil {
		t.Fatalf("1300 senaryosu yüklenemedi: %v", err)
	}
	if gs.Imperial == nil || gs.Imperial.EmpireID != "hre" || gs.Imperial.Authority != 62 {
		t.Fatalf("1300 HRE state'i yüklenmedi: %+v", gs.Imperial)
	}
	for _, memberID := range []string{"flanders_county", "carniola_margraviate", "milan_duchy", "savoy_county", "teutonic_order", "austria_duchy", "bohemian_kingdom", "bavaria_duchy", "saxony_duchy", "brandenburg_margraviate"} {
		if gs.Imperial.Members[faction.FactionID(memberID)] == nil {
			t.Fatalf("HRE üyesi eksik: %s", memberID)
		}
	}
}
