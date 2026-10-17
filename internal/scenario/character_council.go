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
	ID                    string         `json:"id"`
	NameTR                string         `json:"name_tr"`
	NoCost                bool           `json:"no_cost,omitempty"`
	SetCommanderArmy      bool           `json:"set_commander_army,omitempty"`
	AllowedCharacterTypes []string       `json:"allowed_character_types,omitempty"`
	Effects               CouncilEffects `json:"effects,omitempty"`
	PerLevel              CouncilEffects `json:"per_level,omitempty"`
	TraitFallbackEffects  CouncilEffects `json:"trait_fallback_effects,omitempty"`
}

func (r CouncilRoleDefinition) AllowsCharacterType(characterType string) bool {
	if len(r.AllowedCharacterTypes) == 0 {
		return true
	}
	for _, allowedType := range r.AllowedCharacterTypes {
		if allowedType == characterType {
			return true
		}
	}
	return false
}

// CharacterTraitDefinition names a trait and describes its additional council effects.
type CharacterTraitDefinition struct {
	NameTR         string                    `json:"name_tr"`
	CouncilEffects map[string]CouncilEffects `json:"council_effects,omitempty"`
}

// CharacterCouncilConfig is the scenario-specific character duty configuration.
type CharacterCouncilConfig struct {
	XPPerTurn   int                          `json:"xp_per_turn,omitempty"`
	XPPerEffect int                          `json:"xp_per_effect,omitempty"`
	Roles       []CouncilRoleDefinition      `json:"roles"`
	Missions    []CharacterMissionDefinition `json:"-"`
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
			{ID: "treasurer", NameTR: "Hazine Sorumlusu", Effects: CouncilEffects{IncomeMod: 1.05}, PerLevel: CouncilEffects{IncomeMod: 1.01}},
			{ID: "diplomat", NameTR: "Diplomat", Effects: CouncilEffects{RelationBonusPoints: 3}, PerLevel: CouncilEffects{RelationBonusPoints: 1}},
			{ID: "heir", NameTR: "Veliaht", Effects: CouncilEffects{SatisfactionPoints: 1, RelationBonusPoints: 1}, PerLevel: CouncilEffects{SatisfactionPoints: 1, RelationBonusPoints: 1}},
			{ID: "military_responsible", NameTR: "Askeri Sorumlu", SetCommanderArmy: true, AllowedCharacterTypes: []string{"commander", "administrator"}, Effects: CouncilEffects{MilitaryPowerMod: 1.03, CommanderPowerMod: 1.03}, PerLevel: CouncilEffects{MilitaryPowerMod: 1.01, CommanderPowerMod: 1.01}, TraitFallbackEffects: CouncilEffects{MilitaryPowerMod: 0.99}},
			{ID: "research_responsible", NameTR: "Bilim Sorumlusu", Effects: CouncilEffects{ResearchSpeedMod: 1.05}, PerLevel: CouncilEffects{ResearchSpeedMod: 1.01}},
		},
		Traits: map[string]CharacterTraitDefinition{},
	}
}

func (c CharacterCouncilConfig) WithDefaults() CharacterCouncilConfig {
	if c.XPPerTurn > 0 && c.XPPerEffect > 0 && c.Roles != nil && c.Traits != nil {
		return c
	}
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
		allowedCharacterTypes := make(map[string]struct{}, len(role.AllowedCharacterTypes))
		for _, characterType := range role.AllowedCharacterTypes {
			switch characterType {
			case "administrator", "commander", "character":
			default:
				return fmt.Errorf("karakter görevi %q geçersiz allowed_character_types girdisi içeriyor: %q", role.ID, characterType)
			}
			if _, exists := allowedCharacterTypes[characterType]; exists {
				return fmt.Errorf("karakter görevi %q yinelenen allowed_character_types girdisi içeriyor: %q", role.ID, characterType)
			}
			allowedCharacterTypes[characterType] = struct{}{}
		}
	}
	missionIDs := make(map[string]struct{}, len(c.Missions))
	for _, mission := range c.Missions {
		if mission.ID == "" || mission.NameTR == "" {
			return fmt.Errorf("karakter görevlerinde id ve name_tr zorunludur")
		}
		if _, exists := missionIDs[mission.ID]; exists {
			return fmt.Errorf("yinelenen karakter misyonu %q", mission.ID)
		}
		missionIDs[mission.ID] = struct{}{}
		if mission.RoleID != "" {
			if mission.AllowCouncilRoles {
				return fmt.Errorf("role_id ile allow_council_roles birlikte kullanılamaz: %q", mission.ID)
			}
			if _, exists := roleIDs[mission.RoleID]; !exists {
				return fmt.Errorf("misyon %q bilinmeyen karakter rolüne başvuruyor: %q", mission.ID, mission.RoleID)
			}
		}
		allowedCharacters := make(map[string]struct{}, len(mission.AllowedCharacters))
		for _, characterID := range mission.AllowedCharacters {
			if characterID == "" {
				return fmt.Errorf("misyon %q boş allowed_characters girdisi içeriyor", mission.ID)
			}
			if _, exists := allowedCharacters[characterID]; exists {
				return fmt.Errorf("misyon %q yinelenen allowed_characters girdisi içeriyor: %q", mission.ID, characterID)
			}
			allowedCharacters[characterID] = struct{}{}
		}
		allowedFactions := make(map[string]struct{}, len(mission.AllowedFactions))
		for _, factionID := range mission.AllowedFactions {
			if factionID == "" {
				return fmt.Errorf("misyon %q boş allowed_factions girdisi içeriyor", mission.ID)
			}
			if _, exists := allowedFactions[factionID]; exists {
				return fmt.Errorf("misyon %q yinelenen allowed_factions girdisi içeriyor: %q", mission.ID, factionID)
			}
			allowedFactions[factionID] = struct{}{}
		}
		blockedFactions := make(map[string]struct{}, len(mission.BlockedFactions))
		for _, factionID := range mission.BlockedFactions {
			if factionID == "" {
				return fmt.Errorf("misyon %q boş blocked_factions girdisi içeriyor", mission.ID)
			}
			if _, exists := blockedFactions[factionID]; exists {
				return fmt.Errorf("misyon %q yinelenen blocked_factions girdisi içeriyor: %q", mission.ID, factionID)
			}
			blockedFactions[factionID] = struct{}{}
		}
		for field, flags := range map[string][]string{
			"requires_flags": mission.RequiresFlags,
			"blocks_flags":   mission.BlocksFlags,
		} {
			seenFlags := make(map[string]struct{}, len(flags))
			for _, flag := range flags {
				if flag == "" {
					return fmt.Errorf("misyon %q boş %s girdisi içeriyor", mission.ID, field)
				}
				if _, exists := seenFlags[flag]; exists {
					return fmt.Errorf("misyon %q yinelenen %s girdisi içeriyor: %q", mission.ID, field, flag)
				}
				seenFlags[flag] = struct{}{}
			}
		}
		if mission.DurationTurns <= 0 || mission.GoldCost < 0 || mission.MinYear < 0 {
			return fmt.Errorf("misyon %q pozitif duration_turns, negatif olmayan gold_cost ve min_year gerektirir", mission.ID)
		}
		switch mission.Type {
		case "":
			if mission.Adventure != nil {
				return fmt.Errorf("misyon %q adventure tanımı içeremez", mission.ID)
			}
		case CharacterMissionTypeStandard:
			if mission.Adventure != nil {
				return fmt.Errorf("standart misyon %q adventure tanımı içeremez", mission.ID)
			}
		case CharacterMissionTypeAdventure:
			if mission.Adventure == nil || mission.TargetKind != CharacterMissionTargetNone {
				return fmt.Errorf("macera misyonu %q hedef almamalı ve adventure tanımı içermelidir", mission.ID)
			}
			if mission.CompletionEffects != (CharacterMissionEffects{}) {
				return fmt.Errorf("macera misyonu %q ödülleri adventure sonuçlarında tanımlanmalıdır", mission.ID)
			}
			adventure := mission.Adventure
			if adventure.SuccessChancePercent < 0 || adventure.SuccessChancePercent > 100 ||
				adventure.PartialSuccessChancePercent < 0 || adventure.PartialSuccessChancePercent > 100 ||
				adventure.SuccessChancePercent+adventure.PartialSuccessChancePercent > 100 ||
				adventure.TraitBonusPercent < 0 || adventure.LevelBonusPercent < 0 ||
				adventure.FailureUnavailableTurns < 0 || adventure.SuccessGoldReward < 0 ||
				adventure.PartialGoldReward < 0 || adventure.SuccessCouncilExperience < 0 ||
				adventure.PartialCouncilExperience < 0 {
				return fmt.Errorf("macera misyonu %q olasılık, ceza ve ödülleri geçersiz", mission.ID)
			}
			if err := validateAdventureAbilityChoices(mission.ID, adventure.SuccessAbilityChoices); err != nil {
				return err
			}
			if err := validateAdventureAbilityChoices(mission.ID, adventure.PartialAbilityChoices); err != nil {
				return err
			}
		default:
			return fmt.Errorf("misyon %q geçersiz type: %q", mission.ID, mission.Type)
		}
		switch mission.TargetKind {
		case CharacterMissionTargetNone, CharacterMissionTargetFaction, CharacterMissionTargetRegion:
		default:
			return fmt.Errorf("misyon %q geçersiz target_kind: %q", mission.ID, mission.TargetKind)
		}

		if mission.CompletionEffects.GoldReward < 0 || mission.CompletionEffects.CouncilExperience < 0 ||
			mission.CompletionEffects.MilitaryIntelTurns < 0 || mission.CompletionEffects.RegionIntelTurns < 0 {
			return fmt.Errorf("misyon %q completion_effect ödülleri negatif olamaz", mission.ID)
		}
		if mission.TargetKind != CharacterMissionTargetFaction &&
			(mission.CompletionEffects.RelationBonusPoints != 0 || mission.CompletionEffects.MilitaryIntelTurns != 0) {
			return fmt.Errorf("misyon %q relation_bonus_points ve military_intel_turns için faction hedefi gerektirir", mission.ID)
		}
		if mission.TargetKind != CharacterMissionTargetRegion && mission.CompletionEffects.RegionIntelTurns != 0 {
			return fmt.Errorf("misyon %q region_military_intel_turns için region hedefi gerektirir", mission.ID)
		}
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

func validateAdventureAbilityChoices(missionID string, choices []AdventureAbilityChoice) error {
	ids := make(map[string]struct{}, len(choices))
	for _, ability := range choices {
		if ability.ID == "" || ability.NameTR == "" {
			return fmt.Errorf("macera misyonu %q yetenek seçeneklerinde id ve name_tr zorunludur", missionID)
		}
		if _, exists := ids[ability.ID]; exists {
			return fmt.Errorf("macera misyonu %q yinelenen yetenek seçeneği %q içeriyor", missionID, ability.ID)
		}
		ids[ability.ID] = struct{}{}
		if ability.Scope == "" && ability.Effect == "" && ability.Modifier == 0 {
			continue
		}
		validEffect := false
		switch ability.Scope {
		case "faction":
			validEffect = ability.Effect == "military_power"
		case "army":
			validEffect = ability.Effect == "army_power"
		case "navy":
			validEffect = ability.Effect == "navy_power"
		}
		if !validEffect || ability.Modifier <= 0 {
			return fmt.Errorf("macera misyonu %q yeteneği %q için scope, effect ve pozitif modifier uyumsuz", missionID, ability.ID)
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
