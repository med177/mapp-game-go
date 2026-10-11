package tech

import (
	"encoding/json"
	"testing"
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
