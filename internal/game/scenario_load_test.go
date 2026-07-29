package game

import (
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
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

func TestLoad1300StartingGrainCoversTwelveArmyTurns(t *testing.T) {
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

	// Başlangıç orduları hareketsizdir; garrison indirimi ve filolar dahil
	// olmak üzere runtime'ın canonical bakım hesabını kullanırız.
	armyUpkeep := make(map[faction.FactionID]int, len(gs.Factions))
	for armyID, candidate := range gs.Armies {
		if candidate == nil {
			continue
		}
		ownerID := faction.FactionID(candidate.OwnerID)
		if gs.Factions[ownerID] == nil {
			t.Errorf("tahıl bakım hesabı bilinmeyen devlete bağlı: army=%s faction=%s", armyID, ownerID)
			continue
		}
		armyUpkeep[ownerID] += gs.EffectiveArmyGrainUpkeep(candidate)
	}

	const reserveTurns = 12
	for factionID, definition := range gs.Factions {
		if definition == nil {
			continue
		}
		required := armyUpkeep[factionID] * reserveTurns
		if definition.Grain < required {
			t.Errorf("başlangıç tahıl stoku 12 tur asker bakımını karşılamıyor: faction=%s grain=%d required=%d army_upkeep=%d", factionID, definition.Grain, required, armyUpkeep[factionID])
		}
	}
}

func TestScenarioLoadEditModeIsExplicit(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	scenarioPath := filepath.Join(root, "assets", "scenarios", "1300_ottoman_rise")
	t.Setenv("EDIT_MODE", "true")

	normal, _, err := loadScenarioData(scenarioPath, 2, nil)
	if err != nil {
		t.Fatalf("normal senaryo yüklenemedi: %v", err)
	}
	if normal.EditMode || normal.Phase != state.PhaseFactionSelect {
		t.Fatalf("Yeni Oyun EDIT_MODE=true olsa bile normal başlamalı: edit=%v phase=%s", normal.EditMode, normal.Phase)
	}

	edit, _, err := loadScenarioDataForMode(scenarioPath, 2, true, nil)
	if err != nil {
		t.Fatalf("edit mode senaryosu yüklenemedi: %v", err)
	}
	if !edit.EditMode || edit.Phase != state.PhaseEditMode {
		t.Fatalf("EDIT MODE butonu senaryoyu edit modda başlatmalı: edit=%v phase=%s", edit.EditMode, edit.Phase)
	}
}
