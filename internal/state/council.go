package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
)

const (
	CouncilRoleHeadOfHouse         = "head_of_house"
	CouncilRoleTreasurer           = "treasurer"
	CouncilRoleDiplomat            = "diplomat"
	CouncilRoleHeir                = "heir"
	CouncilRoleMilitaryResponsible = army.CouncilRoleMilitaryResponsible
	CouncilRoleResearchResponsible = "research_responsible"
)

func commanderCouncilRoleAllowsArmyCommand(character *army.Commander) bool {
	return character == nil || character.AssignedCouncilRole == "" ||
		character.AssignedCouncilRole == CouncilRoleMilitaryResponsible
}

func (s *GameState) councilConfig() scenario.CharacterCouncilConfig {
	if s == nil {
		return scenario.DefaultCharacterCouncilConfig()
	}
	return s.CouncilConfig.WithDefaults()
}

func (s *GameState) validCouncilRole(role string) bool {
	_, ok := s.councilConfig().Role(role)
	return ok
}

func (s *GameState) councilCharacterAvailable(ownerID, roleID string, character *army.Commander) bool {
	return s != nil && ownerID != "" && character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter) &&
		character.OwnerID == ownerID && character.ActiveInYear(s.Year) &&
		(character.AssignedArmyID == "" || roleID == CouncilRoleMilitaryResponsible) &&
		(character.AssignedCouncilRole == "" || s.validCouncilRole(character.AssignedCouncilRole)) &&
		!s.DeadCommanderIDs[character.ID] && !s.DismissedCommanderIDs[character.ID] &&
		!s.ExiledCommanderIDs[character.ID] && character.CaptiveAtSettlementID == "" &&
		!s.CommanderUnavailable(character)
}

// CanAssignCharacterToCouncilRole atama veya UI disabled durumu için
// karakterin belirtilen senaryo görevine atanıp atanamayacağını döner.
func (s *GameState) CanAssignCharacterToCouncilRole(ownerID faction.FactionID, characterID, role string) bool {
	if s == nil || ownerID == "" || characterID == "" || !s.validCouncilRole(role) {
		return false
	}
	owner := s.Factions[ownerID]
	return owner != nil && !owner.IsEliminated && s.councilCharacterAvailable(string(ownerID), role, s.Commanders[characterID])
}

// AssignCharacterToCouncilRole, faction'ın aktif yönetici/karakterlerinden
// birini senaryo görevine atar veya mevcut görevinden geçirir.
func (s *GameState) AssignCharacterToCouncilRole(ownerID faction.FactionID, characterID, role string) bool {
	if !s.CanAssignCharacterToCouncilRole(ownerID, characterID, role) {
		return false
	}
	character := s.Commanders[characterID]
	for _, incumbent := range s.Commanders {
		if incumbent != nil && incumbent.OwnerID == string(ownerID) && incumbent.AssignedCouncilRole == role && incumbent.ID != characterID {
			incumbent.AssignedCouncilRole = ""
		}
	}
	character.AssignedCouncilRole = role
	return true
}

// UnassignCharacterFromCouncilRole karakteri konsey koltuğundan çıkarır.
func (s *GameState) UnassignCharacterFromCouncilRole(ownerID faction.FactionID, characterID string) bool {
	if s == nil || ownerID == "" || characterID == "" {
		return false
	}
	character := s.Commanders[characterID]
	if character == nil || character.OwnerID != string(ownerID) || character.AssignedCouncilRole == "" {
		return false
	}
	character.AssignedCouncilRole = ""
	return true
}

// CouncilRoleHolder deterministik olarak kullanılabilir görev sahibini döner.
func (s *GameState) CouncilRoleHolder(ownerID faction.FactionID, role string) *army.Commander {
	if s == nil || ownerID == "" || !s.validCouncilRole(role) {
		return nil
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return nil
	}
	var holder *army.Commander
	for _, character := range s.Commanders {
		if character == nil || character.AssignedCouncilRole != role ||
			!s.councilCharacterAvailableForEffect(string(ownerID), role, character) {
			continue
		}
		if holder == nil || character.ID < holder.ID {
			holder = character
		}
	}
	return holder
}

func (s *GameState) councilCharacterAvailableForEffect(ownerID, roleID string, character *army.Commander) bool {
	return character != nil && character.OwnerID == ownerID && character.ActiveInYear(s.Year) &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter) &&
		(character.AssignedArmyID == "" || roleID == CouncilRoleMilitaryResponsible) &&
		!s.DeadCommanderIDs[character.ID] &&
		!s.DismissedCommanderIDs[character.ID] && !s.ExiledCommanderIDs[character.ID] &&
		character.CaptiveAtSettlementID == "" && !s.CommanderUnavailable(character)
}

func (s *GameState) CouncilRoleEffects(ownerID faction.FactionID, roleID string) scenario.CouncilEffects {
	holder := s.CouncilRoleHolder(ownerID, roleID)
	if holder == nil {
		return scenario.CouncilEffects{}
	}
	return s.CouncilRoleEffectsForCharacter(holder, roleID)
}

// CouncilRoleEffectsForCharacter returns the configured role, level, and trait
// effects for a character without requiring that character to hold the role.
func (s *GameState) CouncilRoleEffectsForCharacter(character *army.Commander, roleID string) scenario.CouncilEffects {
	if character == nil {
		return scenario.CouncilEffects{}
	}
	config := s.councilConfig()
	role, ok := config.Role(roleID)
	if !ok {
		return scenario.CouncilEffects{}
	}
	level := character.CouncilLevel
	if level < army.CommanderStartingLevel {
		level = army.CouncilLevelForExperience(character.CouncilExperience)
	}
	effects := addCouncilEffects(role.Effects, scaleCouncilEffects(role.PerLevel, level-1))
	hasPositiveTraitEffect := false
	for _, trait := range character.Traits {
		if definition, exists := config.Traits[string(trait)]; exists {
			traitEffects := definition.CouncilEffects[roleID]
			effects = addCouncilEffects(effects, traitEffects)
			if councilEffectsImproveFallbackChannel(traitEffects, role.TraitFallbackEffects) {
				hasPositiveTraitEffect = true
			}
		}
	}
	if !hasPositiveTraitEffect {
		effects = addCouncilEffects(effects, role.TraitFallbackEffects)
	}
	return effects
}

func councilEffectsImproveFallbackChannel(effects, fallback scenario.CouncilEffects) bool {
	return (fallback.IncomeMod != 0 && effects.IncomeMod > 1) ||
		(fallback.SatisfactionPoints != 0 && effects.SatisfactionPoints > 0) ||
		(fallback.MilitaryPowerMod != 0 && effects.MilitaryPowerMod > 1) ||
		(fallback.CommanderPowerMod != 0 && effects.CommanderPowerMod > 1) ||
		(fallback.RelationBonusPoints != 0 && effects.RelationBonusPoints > 0) ||
		(fallback.ResourceProductionMod != 0 && effects.ResourceProductionMod > 1) ||
		(fallback.ResearchSpeedMod != 0 && effects.ResearchSpeedMod > 1)
}

func (s *GameState) CouncilEffectsForFaction(ownerID faction.FactionID) scenario.CouncilEffects {
	var effects scenario.CouncilEffects
	for _, role := range s.councilConfig().Roles {
		effects = addCouncilEffects(effects, s.CouncilRoleEffects(ownerID, role.ID))
	}
	return effects
}

func addCouncilEffects(a, b scenario.CouncilEffects) scenario.CouncilEffects {
	return scenario.CouncilEffects{
		IncomeMod:             addCouncilMods(a.IncomeMod, b.IncomeMod),
		SatisfactionPoints:    a.SatisfactionPoints + b.SatisfactionPoints,
		MilitaryPowerMod:      addCouncilMods(a.MilitaryPowerMod, b.MilitaryPowerMod),
		CommanderPowerMod:     addCouncilMods(a.CommanderPowerMod, b.CommanderPowerMod),
		RelationBonusPoints:   a.RelationBonusPoints + b.RelationBonusPoints,
		ResourceProductionMod: addCouncilMods(a.ResourceProductionMod, b.ResourceProductionMod),
		ResearchSpeedMod:      addCouncilMods(a.ResearchSpeedMod, b.ResearchSpeedMod),
	}
}

func addCouncilMods(a, b float64) float64 {
	if a == 0 {
		return normalizeCouncilMod(b)
	}
	if b == 0 {
		return normalizeCouncilMod(a)
	}
	return normalizeCouncilMod(a) + (normalizeCouncilMod(b) - 1)
}

func scaleCouncilEffects(effects scenario.CouncilEffects, scale int) scenario.CouncilEffects {
	return scenario.CouncilEffects{
		IncomeMod:             scaleCouncilMod(effects.IncomeMod, scale),
		SatisfactionPoints:    effects.SatisfactionPoints * scale,
		MilitaryPowerMod:      scaleCouncilMod(effects.MilitaryPowerMod, scale),
		CommanderPowerMod:     scaleCouncilMod(effects.CommanderPowerMod, scale),
		RelationBonusPoints:   effects.RelationBonusPoints * scale,
		ResourceProductionMod: scaleCouncilMod(effects.ResourceProductionMod, scale),
		ResearchSpeedMod:      scaleCouncilMod(effects.ResearchSpeedMod, scale),
	}
}

func scaleCouncilMod(mod float64, scale int) float64 {
	return 1 + (normalizeCouncilMod(mod)-1)*float64(scale)
}

func normalizeCouncilMod(mod float64) float64 {
	if mod == 0 {
		return 1
	}
	return mod
}

func (s *GameState) CouncilIncomeBonus(ownerID faction.FactionID, income int) int {
	if income <= 0 {
		return 0
	}
	return int(float64(income)*(s.CouncilEffectsForFaction(ownerID).IncomeMod-1) + 1e-9)
}

func (s *GameState) CouncilSatisfactionBonus(ownerID faction.FactionID) int {
	if s == nil {
		return 0
	}
	return s.CouncilEffectsForFaction(ownerID).SatisfactionPoints
}

func (s *GameState) DiplomatRelationBonus(ownerID faction.FactionID) int {
	if s == nil {
		return 0
	}
	return s.CouncilEffectsForFaction(ownerID).RelationBonusPoints
}

func (s *GameState) CouncilMilitaryPowerMod(ownerID faction.FactionID) float64 {
	if s == nil {
		return 1
	}
	return normalizeCouncilMod(s.CouncilEffectsForFaction(ownerID).MilitaryPowerMod)
}

func (s *GameState) CouncilCommanderPowerMod(ownerID faction.FactionID) float64 {
	if s == nil {
		return 1
	}
	return normalizeCouncilMod(s.CouncilRoleEffects(ownerID, CouncilRoleMilitaryResponsible).CommanderPowerMod)
}

func (s *GameState) CouncilResourceProductionMod(ownerID faction.FactionID) float64 {
	if s == nil {
		return 1
	}
	return normalizeCouncilMod(s.CouncilEffectsForFaction(ownerID).ResourceProductionMod)
}

func (s *GameState) CouncilResearchSpeedMod(ownerID faction.FactionID) float64 {
	if s == nil {
		return 0
	}
	return normalizeCouncilMod(s.CouncilEffectsForFaction(ownerID).ResearchSpeedMod) - 1
}

func (s *GameState) RecordCouncilEffectExperience(ownerID faction.FactionID, effectType string) {
	if s == nil {
		return
	}
	config := s.councilConfig()
	for _, role := range config.Roles {
		holder := s.CouncilRoleHolder(ownerID, role.ID)
		if holder == nil {
			continue
		}
		effects := s.CouncilRoleEffects(ownerID, role.ID)
		switch effectType {
		case "income_mod":
			if effects.IncomeMod != 0 && effects.IncomeMod != 1 {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		case "satisfaction_points":
			if effects.SatisfactionPoints != 0 {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		case "military_power_mod", "commander_power_mod":
			if (effects.MilitaryPowerMod != 0 && effects.MilitaryPowerMod != 1) ||
				(effects.CommanderPowerMod != 0 && effects.CommanderPowerMod != 1) {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		case "relation_bonus_points":
			if effects.RelationBonusPoints != 0 {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		case "resource_production_mod":
			if effects.ResourceProductionMod != 0 && effects.ResourceProductionMod != 1 {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		case "research_speed_mod":
			if effects.ResearchSpeedMod != 0 && effects.ResearchSpeedMod != 1 {
				holder.RecordCouncilExperience(config.XPPerEffect)
			}
		}
	}
}

func (s *GameState) RecordCouncilTurnExperience() {
	if s == nil {
		return
	}
	config := s.councilConfig()
	for ownerID := range s.Factions {
		for _, role := range config.Roles {
			if holder := s.CouncilRoleHolder(ownerID, role.ID); holder != nil {
				holder.RecordCouncilExperience(config.XPPerTurn)
			}
		}
	}
}

// RecordCouncilBattleExperience awards duty XP to factions that used military effects in battle.
func (s *GameState) RecordCouncilBattleExperience(ownerIDs ...faction.FactionID) {
	for _, ownerID := range ownerIDs {
		s.RecordCouncilEffectExperience(ownerID, "military_power_mod")
	}
}
