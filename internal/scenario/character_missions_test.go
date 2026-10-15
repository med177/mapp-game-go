package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScenarioCharacterMissionsSupportsRoleLinkedAndStandaloneTasks(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("data klasörü oluşturulamadı: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "scenario.json"), []byte(`{"id":"test","name":"Test","period":"medieval"}`), 0o600); err != nil {
		t.Fatalf("scenario.json yazılamadı: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "character_council.json"), []byte(`{"roles":[{"id":"diplomat","name_tr":"Diplomat"}]}`), 0o600); err != nil {
		t.Fatalf("character_council.json yazılamadı: %v", err)
	}
	missions := `{"missions":[
		{"id":"envoy","name_tr":"Elçilik","role_id":"diplomat","target_kind":"faction","duration_turns":2,"gold_cost":100,"completion_effects":{"relation_bonus_points":5}},
		{"id":"archive","name_tr":"Arşiv Seferi","duration_turns":2,"gold_cost":25,"completion_effects":{"council_experience":10}},
		{"id":"spy","name_tr":"İstihbarat","target_kind":"faction","duration_turns":2,"gold_cost":50,"completion_effects":{"military_intel_turns":3}},
		{"id":"recon","name_tr":"Bölgesel Keşif","target_kind":"region","duration_turns":2,"gold_cost":75,"completion_effects":{"region_military_intel_turns":2}}
	]}`
	if err := os.WriteFile(filepath.Join(dataDir, "character_missions.json"), []byte(missions), 0o600); err != nil {
		t.Fatalf("character_missions.json yazılamadı: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("görevleri olan senaryo yüklenemedi: %v", err)
	}
	if len(loaded.CharacterCouncil.Missions) != 4 {
		t.Fatalf("yüklü misyon sayısı = %d, want 4", len(loaded.CharacterCouncil.Missions))
	}
	envoy, archive, spy, recon := loaded.CharacterCouncil.Missions[0], loaded.CharacterCouncil.Missions[1], loaded.CharacterCouncil.Missions[2], loaded.CharacterCouncil.Missions[3]
	if envoy.RoleID != "diplomat" || envoy.TargetKind != CharacterMissionTargetFaction || envoy.DurationTurns != 2 || envoy.GoldCost != 100 || envoy.CompletionEffects.RelationBonusPoints != 5 {
		t.Fatalf("role-linked mission = %+v", envoy)
	}
	if archive.RoleID != "" || archive.TargetKind != CharacterMissionTargetNone || archive.CompletionEffects.CouncilExperience != 10 {
		t.Fatalf("standalone mission = %+v", archive)
	}
	if spy.TargetKind != CharacterMissionTargetFaction || spy.CompletionEffects.MilitaryIntelTurns != 3 {
		t.Fatalf("spy mission = %+v", spy)
	}
	if recon.TargetKind != CharacterMissionTargetRegion || recon.CompletionEffects.RegionIntelTurns != 2 {
		t.Fatalf("region reconnaissance mission = %+v", recon)
	}
}

func TestCharacterMissionRegionIntelRequiresRegionTarget(t *testing.T) {
	config := CharacterCouncilConfig{Missions: []CharacterMissionDefinition{{
		ID: "recon", NameTR: "Bölgesel Keşif", TargetKind: CharacterMissionTargetRegion,
		DurationTurns: 2, CompletionEffects: CharacterMissionEffects{RegionIntelTurns: 2},
	}}}
	if err := config.Validate(); err != nil {
		t.Fatalf("region-targeted intel mission rejected: %v", err)
	}
	config.Missions[0].TargetKind = CharacterMissionTargetFaction
	if err := config.Validate(); err == nil {
		t.Fatal("region intel mission without region target was accepted")
	}
}

func TestLoadCharacterMissionsRejectsUnknownRoleAndInvalidDuration(t *testing.T) {
	config := CharacterCouncilConfig{
		Roles: []CouncilRoleDefinition{{ID: "diplomat", NameTR: "Diplomat"}},
		Missions: []CharacterMissionDefinition{{
			ID: "envoy", NameTR: "Elçilik", RoleID: "unknown",
			TargetKind: CharacterMissionTargetFaction, DurationTurns: 2,
			CompletionEffects: CharacterMissionEffects{RelationBonusPoints: 5},
		}},
	}

	if err := config.Validate(); err == nil {
		t.Fatal("unknown mission role was accepted")
	}

	config.Missions[0].RoleID = "diplomat"
	config.Missions[0].DurationTurns = 0
	if err := config.Validate(); err == nil {
		t.Fatal("mission with non-positive duration was accepted")
	}
}

func TestCharacterMissionMilitaryIntelRequiresFactionTarget(t *testing.T) {
	config := CharacterCouncilConfig{
		Roles: []CouncilRoleDefinition{{ID: "diplomat", NameTR: "Diplomat"}},
		Missions: []CharacterMissionDefinition{{
			ID: "spy", NameTR: "İstihbarat", TargetKind: CharacterMissionTargetFaction,
			DurationTurns: 2, CompletionEffects: CharacterMissionEffects{MilitaryIntelTurns: 2},
		}},
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("faction-targeted military intel mission rejected: %v", err)
	}
	config.Missions[0].TargetKind = CharacterMissionTargetNone
	if err := config.Validate(); err == nil {
		t.Fatal("military intel mission without faction target was accepted")
	}
}
