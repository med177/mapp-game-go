package scenario

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	DefaultCouncilXPPerTurn   = 5
	DefaultCouncilXPPerEffect = 10
)

// CouncilEffects contains the supported numeric effects of a council role.
type CouncilEffects struct {
	IncomeMod             float64 `json:"income_mod,omitempty"`
	SatisfactionPoints    int     `json:"satisfaction_points,omitempty"`
	MilitaryPowerMod      float64 `json:"military_power_mod,omitempty"`
	CommanderPowerMod     float64 `json:"commander_power_mod,omitempty"`
	RelationBonusPoints   int     `json:"relation_bonus_points,omitempty"`
	ResourceProductionMod float64 `json:"resource_production_mod,omitempty"`
	ResearchSpeedMod      float64 `json:"research_speed_mod,omitempty"`
}

// CouncilRoleDefinition is a scenario-defined council duty and its level-scaled effect.
type CouncilRoleDefinition struct {
	ID                   string         `json:"id"`
	NameTR               string         `json:"name_tr"`
	NoCost               bool           `json:"no_cost,omitempty"`
	SetCommanderArmy     bool           `json:"set_commander_army,omitempty"`
	Effects              CouncilEffects `json:"effects,omitempty"`
	PerLevel             CouncilEffects `json:"per_level,omitempty"`
	TraitFallbackEffects CouncilEffects `json:"trait_fallback_effects,omitempty"`
}

// CharacterTraitDefinition names a trait and describes its additional council effects.
type CharacterTraitDefinition struct {
	NameTR         string                    `json:"name_tr"`
	CouncilEffects map[string]CouncilEffects `json:"council_effects,omitempty"`
}

// CharacterCouncilConfig is the scenario-specific character duty configuration.
type CharacterCouncilConfig struct {
	XPPerTurn   int                     `json:"xp_per_turn,omitempty"`
	XPPerEffect int                     `json:"xp_per_effect,omitempty"`
	Roles       []CouncilRoleDefinition `json:"roles"`
	// Traits supports older combined council files. New scenario data belongs
	// in character_traits.json and is merged at load time.
	Traits map[string]CharacterTraitDefinition `json:"traits,omitempty"`
}

type CharacterTraitsConfig struct {
	Traits map[string]CharacterTraitDefinition `json:"traits"`
}

func DefaultCharacterCouncilConfig() CharacterCouncilConfig {
	return CharacterCouncilConfig{
		XPPerTurn:   DefaultCouncilXPPerTurn,
		XPPerEffect: DefaultCouncilXPPerEffect,
		Roles: []CouncilRoleDefinition{
			{ID: "head_of_house", NameTR: "Hane Lideri", Effects: CouncilEffects{SatisfactionPoints: 1, MilitaryPowerMod: 1.01, RelationBonusPoints: 1}, PerLevel: CouncilEffects{SatisfactionPoints: 1, MilitaryPowerMod: 1.01, RelationBonusPoints: 1}},
			{ID: "treasurer", NameTR: "Hazine Sorumlusu", Effects: CouncilEffects{IncomeMod: 1.05, ResourceProductionMod: 1.05}, PerLevel: CouncilEffects{IncomeMod: 1.01, ResourceProductionMod: 1.01}},
			{ID: "diplomat", NameTR: "Diplomat", Effects: CouncilEffects{RelationBonusPoints: 3}, PerLevel: CouncilEffects{RelationBonusPoints: 1}},
			{ID: "heir", NameTR: "Veliaht", Effects: CouncilEffects{SatisfactionPoints: 1, RelationBonusPoints: 1}, PerLevel: CouncilEffects{SatisfactionPoints: 1, RelationBonusPoints: 1}},
			{ID: "military_responsible", NameTR: "Askeri Sorumlu", SetCommanderArmy: true, Effects: CouncilEffects{MilitaryPowerMod: 1.03, CommanderPowerMod: 1.03}, PerLevel: CouncilEffects{MilitaryPowerMod: 1.01, CommanderPowerMod: 1.01}, TraitFallbackEffects: CouncilEffects{MilitaryPowerMod: 0.99}},
			{ID: "research_responsible", NameTR: "Bilim Sorumlusu", Effects: CouncilEffects{ResearchSpeedMod: 1.05}, PerLevel: CouncilEffects{ResearchSpeedMod: 1.01}},
		},
		Traits: map[string]CharacterTraitDefinition{},
	}
}

func (c CharacterCouncilConfig) WithDefaults() CharacterCouncilConfig {
	defaults := DefaultCharacterCouncilConfig()
	if c.XPPerTurn <= 0 {
		c.XPPerTurn = defaults.XPPerTurn
	}
	if c.XPPerEffect <= 0 {
		c.XPPerEffect = defaults.XPPerEffect
	}
	if c.Roles == nil {
		c.Roles = defaults.Roles
	}
	if c.Traits == nil {
		c.Traits = defaults.Traits
	}
	return c
}

func LoadCharacterCouncilConfig(path string) (CharacterCouncilConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultCharacterCouncilConfig(), nil
	}
	if err != nil {
		return CharacterCouncilConfig{}, fmt.Errorf("karakter görev tanımı okunamadı: %w", err)
	}
	var config CharacterCouncilConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CharacterCouncilConfig{}, fmt.Errorf("karakter görev tanımı parse edilemedi: %w", err)
	}
	config = config.WithDefaults()
	if err := config.Validate(); err != nil {
		return CharacterCouncilConfig{}, err
	}
	return config, nil
}

func LoadCharacterTraitsConfig(path string) (CharacterTraitsConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return CharacterTraitsConfig{Traits: map[string]CharacterTraitDefinition{}}, nil
	}
	if err != nil {
		return CharacterTraitsConfig{}, fmt.Errorf("karakter nitelik tanımı okunamadı: %w", err)
	}
	var config CharacterTraitsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return CharacterTraitsConfig{}, fmt.Errorf("karakter nitelik tanımı parse edilemedi: %w", err)
	}
	if config.Traits == nil {
		config.Traits = map[string]CharacterTraitDefinition{}
	}
	return config, nil
}

func (c CharacterCouncilConfig) Validate() error {
	roleIDs := make(map[string]struct{}, len(c.Roles))
	for _, role := range c.Roles {
		if role.ID == "" || role.NameTR == "" {
			return fmt.Errorf("karakter görevlerinde id ve name_tr zorunludur")
		}
		if _, exists := roleIDs[role.ID]; exists {
			return fmt.Errorf("yinelenen karakter görevi %q", role.ID)
		}
		roleIDs[role.ID] = struct{}{}
	}
	for traitID, trait := range c.Traits {
		if traitID == "" || trait.NameTR == "" {
			return fmt.Errorf("karakter niteliklerinde id ve name_tr zorunludur")
		}
		for roleID := range trait.CouncilEffects {
			if _, exists := roleIDs[roleID]; !exists {
				return fmt.Errorf("nitelik %q bilinmeyen karakter görevine başvuruyor: %q", traitID, roleID)
			}
		}
	}
	return nil
}

func (c CharacterCouncilConfig) Role(roleID string) (CouncilRoleDefinition, bool) {
	for _, role := range c.Roles {
		if role.ID == roleID {
			return role, true
		}
	}
	return CouncilRoleDefinition{}, false
}
