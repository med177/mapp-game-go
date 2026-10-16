package tech

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/faction"
)

func TestTechnologyFactionAvailability(t *testing.T) {
	var technology Technology
	if err := json.Unmarshal([]byte(`{"blocked_factions":["free_folk","nights_watch"]}`), &technology); err != nil {
		t.Fatalf("blocked_factions yüklenemedi: %v", err)
	}

	if technology.IsAvailableToFaction("free_folk") || technology.IsAvailableToFaction("nights_watch") {
		t.Fatal("engellenen fraksiyonlar teknolojiye erişememeliydi")
	}
	if !technology.IsAvailableToFaction("stark") {
		t.Fatal("yalnızca engellenen fraksiyonları olan teknoloji diğer fraksiyonlara açık olmalıydı")
	}

	technology.AllowedFactions = []string{"free_folk", "stark"}
	if technology.IsAvailableToFaction("free_folk") {
		t.Fatal("blocked_factions, allowed_factions listesinden öncelikli olmalıydı")
	}
	if !technology.IsAvailableToFaction("stark") || technology.IsAvailableToFaction("lannister") {
		t.Fatal("allowed_factions mevcut sınırlama davranışını korumalıydı")
	}
}

func TestTechnologyReligionAvailability(t *testing.T) {
	var technology Technology
	if err := json.Unmarshal([]byte(`{"allowed_factions":["baratheon_dragonstone"],"allowed_religions":["lord_of_light"]}`), &technology); err != nil {
		t.Fatalf("allowed_religions yüklenemedi: %v", err)
	}

	if !technology.IsAvailableToFactionAndReligion("baratheon_dragonstone", "lord_of_light") {
		t.Fatal("uygun faction ve dine teknoloji açılmalıydı")
	}
	if technology.IsAvailableToFactionAndReligion("baratheon_dragonstone", "faith_of_the_seven") {
		t.Fatal("aynı faction'ın farklı dini teknolojiye erişememeliydi")
	}
	if technology.IsAvailableToFactionAndReligion("stark", "lord_of_light") {
		t.Fatal("uygun dini olsa da faction kısıtlaması korunmalıydı")
	}
	if technology.IsAvailableToFaction("baratheon_dragonstone") {
		t.Fatal("din bağlamı olmayan sorgu dine özel teknolojiye erişememeliydi")
	}
}

func TestComputeEffectsIncludesNearbyArmyIntel(t *testing.T) {
	techs := map[string]*Technology{
		"nearby": {ID: "nearby", Effects: Effects{RevealNearbyArmyIntel: true}},
	}
	effects := ComputeEffects(map[string]bool{"nearby": true}, techs)
	if !effects.RevealNearbyArmyIntel {
		t.Fatal("yakın ordu istihbaratı etkisi tamamlanmış teknolojiden birleştirilmedi")
	}
	if effects.RevealEnemyStrength {
		t.Fatal("yakın ordu istihbaratı global düşman gücü açığına dönüşmemeliydi")
	}
}

func TestIsUnlockedForFactionReligionContext(t *testing.T) {
	technology := &Technology{ID: "old_gods_customs", AllowedReligions: []string{"old_gods"}}
	research := &faction.ResearchState{Completed: map[string]bool{}}

	if !IsUnlockedForFactionReligionContext(research, technology, "stark", "old_gods", 298, nil) {
		t.Fatal("Eski Tanrılar inancındaki faction teknolojiye erişebilmeliydi")
	}
	if IsUnlockedForFactionReligionContext(research, technology, "lannister", "faith_of_the_seven", 298, nil) {
		t.Fatal("Yedi İnancı'ndaki faction Eski Tanrılar teknolojisine erişememeliydi")
	}
}

func TestNextResearchableTechIDRespectsFactionReligion(t *testing.T) {
	research := &faction.ResearchState{Completed: map[string]bool{}}
	allTechs := map[string]*Technology{
		"old_gods_customs": {
			ID: "old_gods_customs", GoldCost: 10, AllowedReligions: []string{"old_gods"},
		},
		"faith_of_the_seven": {
			ID: "faith_of_the_seven", GoldCost: 10, AllowedReligions: []string{"faith_of_the_seven"},
		},
	}

	id, ok := NextResearchableTechIDForFactionReligionContextWithModifiers(
		research, allTechs, "stark", "old_gods", 100, 298, nil, 0,
	)
	if !ok || id != "old_gods_customs" {
		t.Fatalf("Stark araştırma seçimi = %q, %t; Eski Tanrılar teknolojisi bekleniyordu", id, ok)
	}
}
