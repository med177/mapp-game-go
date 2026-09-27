package scenario

import (
	"path/filepath"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestValidateAIReferencesRejectsRemovedRegionAndFaction(t *testing.T) {
	config := LoadedAIConfig{Strategies: map[string]AIFactionStrategy{
		"owner": {
			FactionID:         "owner",
			ExpansionTargets:  []string{"removed_target"},
			TerritorialClaims: []AITerritorialClaimDef{{RegionID: "removed_region", Value: 50}},
		},
	}}
	regions := map[world.RegionID]*world.Region{"kept": {ID: "kept"}}
	factions := map[faction.FactionID]*faction.Faction{"owner": {ID: "owner"}}
	if err := ValidateAIReferences(config, regions, factions); err == nil {
		t.Fatal("removed AI references were accepted")
	}
}

func TestValidateFactionTerritorialClaimsRejectsRemovedRegion(t *testing.T) {
	regions := map[world.RegionID]*world.Region{"kept": {ID: "kept"}}
	factions := map[faction.FactionID]*faction.Faction{
		"owner": {ID: "owner", TerritorialClaims: []faction.TerritorialClaim{{RegionID: "removed"}}},
	}
	if err := ValidateFactionTerritorialClaims(regions, factions); err == nil {
		t.Fatal("removed faction claim was accepted")
	}
}

func Test1300ScenarioAIReferencesAndUpdatedAnatolianTarget(t *testing.T) {
	dataDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data")
	config, err := LoadAIConfig(filepath.Join(dataDir, "ai_strategies.json"))
	if err != nil {
		t.Fatal(err)
	}
	regions, err := world.LoadRegions(filepath.Join(dataDir, "regions.json"))
	if err != nil {
		t.Fatal(err)
	}
	factions, err := faction.LoadFactions(filepath.Join(dataDir, "factions.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateAIReferences(config, regions, factions); err != nil {
		t.Fatal(err)
	}

	strategy, ok := config.Strategies["karaman_bey"]
	if !ok || len(strategy.Objectives) == 0 {
		t.Fatal("Karaman AI stratejisi bulunamadı")
	}
	if got := strategy.Objectives[0].TerritorialClaims[0].RegionID; got != "teke" {
		t.Fatalf("Karaman güncel kara hedefini kullanmıyor: got=%s", got)
	}
	if _, ok := config.Strategies["papal_states"]; !ok {
		t.Fatal("Papalık stratejisi güncel faction kimliğiyle yüklenmedi")
	}
	if _, stale := config.Strategies["papal_states_f"]; stale {
		t.Fatal("eski papalık faction kimliği hâlâ stratejilerde")
	}
}
