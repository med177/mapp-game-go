package game

import (
	"path/filepath"
	"runtime"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

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

func TestLoad1300StartingGrainAndArmyUpkeepArePositive(t *testing.T) {
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

	// Başlangıç orduları hareketsizdir; garrison indirimi ve runtime'ın
	// canonical bakım hesabını kullanırız. Güncel ekonomi modeli stokun sabit
	// 12 tur bakımını garanti etmez; kapasite üç aylık ikmal rezervine göre
	// hesaplanır.
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
		upkeep := gs.EffectiveArmyGrainUpkeep(candidate)
		if upkeep <= 0 {
			t.Errorf("başlangıç ordusunun tahıl bakımı pozitif olmalı: army=%s upkeep=%d", armyID, upkeep)
		}
		armyUpkeep[ownerID] += upkeep
	}

	for factionID, definition := range gs.Factions {
		if definition == nil || armyUpkeep[factionID] == 0 {
			continue
		}
		if definition.Grain <= 0 {
			t.Errorf("başlangıç tahıl stoku pozitif olmalı: faction=%s grain=%d army_upkeep=%d", factionID, definition.Grain, armyUpkeep[factionID])
		}
		if capacity := gs.GrainStorageCapacityForFaction(factionID); capacity <= 0 {
			t.Errorf("başlangıç tahıl ambar kapasitesi pozitif olmalı: faction=%s capacity=%d", factionID, capacity)
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
