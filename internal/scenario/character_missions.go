package scenario

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	CharacterMissionTargetNone    = "none"
	CharacterMissionTargetFaction = "faction"
	CharacterMissionTargetRegion  = "region"
	CharacterMissionTypeStandard  = "standard"
	CharacterMissionTypeAdventure = "adventure"
	CharacterAdventureSuccess     = "success"
	CharacterAdventurePartial     = "partial"
	CharacterAdventureFailure     = "failure"
)

type AdventureAbilityChoice struct {
	ID       string  `json:"id"`
	NameTR   string  `json:"name_tr"`
	Scope    string  `json:"scope,omitempty"`
	Effect   string  `json:"effect,omitempty"`
	Modifier float64 `json:"modifier,omitempty"`
}

type CharacterAdventureDefinition struct {
	SuccessChancePercent        int                      `json:"success_chance_percent"`
	PartialSuccessChancePercent int                      `json:"partial_success_chance_percent,omitempty"`
	RelevantTraits              []string                 `json:"relevant_traits,omitempty"`
	TraitBonusPercent           int                      `json:"trait_bonus_percent,omitempty"`
	LevelBonusPercent           int                      `json:"level_bonus_percent,omitempty"`
	FailureUnavailableTurns     int                      `json:"failure_unavailable_turns,omitempty"`
	SuccessGoldReward           int                      `json:"success_gold_reward,omitempty"`
	PartialGoldReward           int                      `json:"partial_gold_reward,omitempty"`
	SuccessCouncilExperience    int                      `json:"success_council_experience,omitempty"`
	PartialCouncilExperience    int                      `json:"partial_council_experience,omitempty"`
	SuccessAbilityChoices       []AdventureAbilityChoice `json:"success_ability_choices,omitempty"`
	PartialAbilityChoices       []AdventureAbilityChoice `json:"partial_ability_choices,omitempty"`
}

// CharacterMissionEffects describes data-driven rewards applied when a task
// completes. Relation changes require a faction target; other rewards do not.
type CharacterMissionEffects struct {
	RelationBonusPoints int `json:"relation_bonus_points,omitempty"`
	GoldReward          int `json:"gold_reward,omitempty"`
	CouncilExperience   int `json:"council_experience,omitempty"`
	MilitaryIntelTurns  int `json:"military_intel_turns,omitempty"`
	RegionIntelTurns    int `json:"region_military_intel_turns,omitempty"`
}

// CharacterMissionDefinition is a scenario-defined, timed character task.
// RoleID links a task to one council role; AllowCouncilRoles lets a role-free
// task also be used by characters assigned to any council role.
type CharacterMissionDefinition struct {
	ID                string                        `json:"id"`
	NameTR            string                        `json:"name_tr"`
	Type              string                        `json:"type,omitempty"`
	RoleID            string                        `json:"role_id,omitempty"`
	AllowCouncilRoles bool                          `json:"allow_council_roles,omitempty"`
	AllowedCharacters []string                      `json:"allowed_characters,omitempty"`
	AllowedFactions   []string                      `json:"allowed_factions,omitempty"`
	BlockedFactions   []string                      `json:"blocked_factions,omitempty"`
	RequiresFlags     []string                      `json:"requires_flags,omitempty"`
	BlocksFlags       []string                      `json:"blocks_flags,omitempty"`
	TargetKind        string                        `json:"target_kind,omitempty"`
	MinYear           int                           `json:"min_year,omitempty"`
	DurationTurns     int                           `json:"duration_turns"`
	GoldCost          int                           `json:"gold_cost,omitempty"`
	CompletionEffects CharacterMissionEffects       `json:"completion_effects,omitempty"`
	Adventure         *CharacterAdventureDefinition `json:"adventure,omitempty"`
}

type CharacterMissionsConfig struct {
	Missions []CharacterMissionDefinition `json:"missions"`
}

// LoadCharacterMissionsConfig loads optional per-scenario character tasks.
// Scenarios without character_missions.json remain compatible.
func LoadCharacterMissionsConfig(path string) (CharacterMissionsConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return CharacterMissionsConfig{}, nil
	}
	if err != nil {
		return CharacterMissionsConfig{}, fmt.Errorf("karakter görevleri okunamadı: %w", err)
	}
	var config CharacterMissionsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CharacterMissionsConfig{}, fmt.Errorf("karakter görevleri parse edilemedi: %w", err)
	}
	for index := range config.Missions {
		if config.Missions[index].TargetKind == "" {
			config.Missions[index].TargetKind = CharacterMissionTargetNone
		}
		if config.Missions[index].Type == "" {
			config.Missions[index].Type = CharacterMissionTypeStandard
		}
	}
	return config, nil
}
