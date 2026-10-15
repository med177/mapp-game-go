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
		{"id":"archive","name_tr":"Arşiv Seferi","allowed_characters":["chr_arya_stark"],"allowed_factions":["stark"],"blocked_factions":["braavos"],"requires_flags":["journey_unlocked"],"blocks_flags":["journey_closed"],"min_year":300,"duration_turns":2,"gold_cost":25,"completion_effects":{"council_experience":10}},
		{"id":"shared","name_tr":"Ortak Görev","allow_council_roles":true,"blocked_factions":["others"],"duration_turns":2},
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
	if len(loaded.CharacterCouncil.Missions) != 5 {
		t.Fatalf("yüklü misyon sayısı = %d, want 5", len(loaded.CharacterCouncil.Missions))
	}
	envoy, archive, shared, spy, recon := loaded.CharacterCouncil.Missions[0], loaded.CharacterCouncil.Missions[1], loaded.CharacterCouncil.Missions[2], loaded.CharacterCouncil.Missions[3], loaded.CharacterCouncil.Missions[4]
	if envoy.RoleID != "diplomat" || envoy.TargetKind != CharacterMissionTargetFaction || envoy.DurationTurns != 2 || envoy.GoldCost != 100 || envoy.CompletionEffects.RelationBonusPoints != 5 {
		t.Fatalf("role-linked mission = %+v", envoy)
	}
	if archive.RoleID != "" || archive.TargetKind != CharacterMissionTargetNone || archive.CompletionEffects.CouncilExperience != 10 ||
		len(archive.AllowedCharacters) != 1 || archive.AllowedCharacters[0] != "chr_arya_stark" ||
		len(archive.AllowedFactions) != 1 || archive.AllowedFactions[0] != "stark" ||
		len(archive.BlockedFactions) != 1 || archive.BlockedFactions[0] != "braavos" ||
		len(archive.RequiresFlags) != 1 || archive.RequiresFlags[0] != "journey_unlocked" ||
		len(archive.BlocksFlags) != 1 || archive.BlocksFlags[0] != "journey_closed" || archive.MinYear != 300 {
		t.Fatalf("standalone mission = %+v", archive)
	}
	if !shared.AllowCouncilRoles || len(shared.BlockedFactions) != 1 || shared.BlockedFactions[0] != "others" {
		t.Fatalf("shared mission = %+v", shared)
	}
	if spy.TargetKind != CharacterMissionTargetFaction || spy.CompletionEffects.MilitaryIntelTurns != 3 {
		t.Fatalf("spy mission = %+v", spy)
	}
	if recon.TargetKind != CharacterMissionTargetRegion || recon.CompletionEffects.RegionIntelTurns != 2 {
		t.Fatalf("region reconnaissance mission = %+v", recon)
	}
}

func TestCharacterMissionValidationRejectsRoleLinkedAllowCouncilRoles(t *testing.T) {
	config := CharacterCouncilConfig{
		Roles: []CouncilRoleDefinition{{ID: "diplomat", NameTR: "Diplomat"}},
		Missions: []CharacterMissionDefinition{{
			ID: "envoy", NameTR: "Elçilik", RoleID: "diplomat", AllowCouncilRoles: true, DurationTurns: 1,
		}},
	}
	if err := config.Validate(); err == nil {
		t.Fatal("role-linked mission with allow_council_roles was accepted")
	}
}

func TestCharacterMissionValidationRejectsInvalidAllowedCharacters(t *testing.T) {
	config := CharacterCouncilConfig{Missions: []CharacterMissionDefinition{{
		ID: "archive", NameTR: "Arşiv Seferi", DurationTurns: 1,
		AllowedCharacters: []string{"chr_arya_stark", "chr_arya_stark"},
	}}}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with duplicate allowed character IDs was accepted")
	}
	config.Missions[0].AllowedCharacters = []string{""}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with an empty allowed character ID was accepted")
	}
	config.Missions[0].AllowedCharacters = nil
	config.Missions[0].AllowedFactions = []string{"stark", "stark"}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with duplicate allowed faction IDs was accepted")
	}
	config.Missions[0].AllowedFactions = []string{""}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with an empty allowed faction ID was accepted")
	}
	config.Missions[0].AllowedFactions = nil
	config.Missions[0].BlockedFactions = []string{"stark", "stark"}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with duplicate blocked faction IDs was accepted")
	}
	config.Missions[0].BlockedFactions = []string{""}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with an empty blocked faction ID was accepted")
	}
	config.Missions[0].BlockedFactions = nil
	config.Missions[0].RequiresFlags = []string{"ready", "ready"}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with duplicate required flags was accepted")
	}
	config.Missions[0].RequiresFlags = []string{""}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with an empty required flag was accepted")
	}
	config.Missions[0].RequiresFlags = nil
	config.Missions[0].BlocksFlags = []string{"closed", "closed"}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with duplicate blocked flags was accepted")
	}
	config.Missions[0].BlocksFlags = []string{""}
	if err := config.Validate(); err == nil {
		t.Fatal("mission with an empty blocked flag was accepted")
	}
	config.Missions[0].BlocksFlags = nil
	config.Missions[0].MinYear = -1
	if err := config.Validate(); err == nil {
		t.Fatal("mission with a negative min_year was accepted")
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

func TestLoadCharacterAdventureMissionAndLegacyMissionType(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.Mkdir(dataDir, 0o700); err != nil {
		t.Fatalf("data klasörü oluşturulamadı: %v", err)
	}
	files := map[string]string{
		"scenario.json":          `{"id":"test","name":"Test","period":"medieval"}`,
		"character_council.json": `{"roles":[]}`,
		"character_missions.json": `{"missions":[
			{"id":"old","name_tr":"Eski görev","duration_turns":1},
			{"id":"adventure","name_tr":"Macera","type":"adventure","duration_turns":2,
			 "adventure":{"success_chance_percent":50,"partial_success_chance_percent":30,
			 "relevant_traits":["explorer"],"trait_bonus_percent":15,"level_bonus_percent":5,
			 "failure_unavailable_turns":2,
			 "success_ability_choices":[{"id":"fieldcraft","name_tr":"Saha Keşfi",
			 "scope":"faction","effect":"military_power","modifier":0.02}]}}
		]}`,
	}
	for filename, contents := range files {
		if err := os.WriteFile(filepath.Join(dataDir, filename), []byte(contents), 0o600); err != nil {
			t.Fatalf("%s yazılamadı: %v", filename, err)
		}
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("macera görevi içeren senaryo yüklenemedi: %v", err)
	}
	if len(loaded.CharacterCouncil.Missions) != 2 {
		t.Fatalf("yüklü görev sayısı = %d, want 2", len(loaded.CharacterCouncil.Missions))
	}
	legacy, adventure := loaded.CharacterCouncil.Missions[0], loaded.CharacterCouncil.Missions[1]
	if legacy.Type != CharacterMissionTypeStandard || legacy.TargetKind != CharacterMissionTargetNone {
		t.Fatalf("eski görev varsayılanları = %+v", legacy)
	}
	if adventure.Type != CharacterMissionTypeAdventure || adventure.Adventure == nil ||
		adventure.Adventure.SuccessChancePercent != 50 || adventure.Adventure.PartialSuccessChancePercent != 30 ||
		len(adventure.Adventure.SuccessAbilityChoices) != 1 ||
		adventure.Adventure.SuccessAbilityChoices[0].ID != "fieldcraft" ||
		adventure.Adventure.SuccessAbilityChoices[0].Scope != "faction" ||
		adventure.Adventure.SuccessAbilityChoices[0].Effect != "military_power" ||
		adventure.Adventure.SuccessAbilityChoices[0].Modifier != 0.02 {
		t.Fatalf("yüklenen macera tanımı = %+v", adventure)
	}

	loaded.CharacterCouncil.Missions[1].Adventure.SuccessChancePercent = 80
	loaded.CharacterCouncil.Missions[1].Adventure.PartialSuccessChancePercent = 30
	if err := loaded.CharacterCouncil.Validate(); err == nil {
		t.Fatal("toplam başarı/kısmi başarı olasılığı %100'ü aşan görev kabul edildi")
	}
}
