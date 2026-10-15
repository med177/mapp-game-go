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
)

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
// RoleID is optional: linked tasks require that council role; role-free tasks
// require an unassigned character.
type CharacterMissionDefinition struct {
	ID                string                  `json:"id"`
	NameTR            string                  `json:"name_tr"`
	RoleID            string                  `json:"role_id,omitempty"`
	TargetKind        string                  `json:"target_kind,omitempty"`
	DurationTurns     int                     `json:"duration_turns"`
	GoldCost          int                     `json:"gold_cost,omitempty"`
	CompletionEffects CharacterMissionEffects `json:"completion_effects,omitempty"`
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
	}
	return config, nil
}
