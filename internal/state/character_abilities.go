package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
)

type AdventureAbilityEffects struct {
	ArmyPowerMod  float64
	NavalPowerMod float64
}

func (s *GameState) adventureAbilityCouncilEffects(ownerID faction.FactionID) scenario.CouncilEffects {
	if s == nil || ownerID == "" {
		return scenario.CouncilEffects{}
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return scenario.CouncilEffects{}
	}

	var militaryPowerMod float64
	for _, character := range s.Commanders {
		if !s.characterCanProvideAdventureAbility(ownerID, character) {
			continue
		}
		for index, ability := range character.AdventureAbilities {
			if ability.ID == "" {
				continue
			}
			if hasEarlierAdventureAbilityID(character, index, ability.ID) {
				continue
			}
			if ability.Scope == "faction" && ability.Effect == "military_power" && ability.Modifier > 0 {
				militaryPowerMod += ability.Modifier
			}
		}
	}
	return scenario.CouncilEffects{MilitaryPowerMod: 1 + militaryPowerMod}
}

func (s *GameState) AdventureAbilityEffectsForArmy(a *army.Army) AdventureAbilityEffects {
	if s == nil || a == nil || a.ID == "" || a.Commander == nil ||
		a.Commander.AssignedArmyID != a.ID || a.Commander.OwnerID != a.OwnerID ||
		!s.characterCanProvideAdventureAbility(faction.FactionID(a.OwnerID), a.Commander) {
		return AdventureAbilityEffects{}
	}

	effects := AdventureAbilityEffects{}
	for index, ability := range a.Commander.AdventureAbilities {
		if ability.ID == "" || ability.Modifier <= 0 {
			continue
		}
		if hasEarlierAdventureAbilityID(a.Commander, index, ability.ID) {
			continue
		}
		if a.IsNaval && ability.Scope == "navy" && ability.Effect == "navy_power" {
			effects.NavalPowerMod += ability.Modifier
		} else if !a.IsNaval && ability.Scope == "army" && ability.Effect == "army_power" {
			effects.ArmyPowerMod += ability.Modifier
		}
	}
	return effects
}

func hasEarlierAdventureAbilityID(character *army.Commander, index int, id string) bool {
	if character == nil || index <= 0 || index > len(character.AdventureAbilities) || id == "" {
		return false
	}
	for _, ability := range character.AdventureAbilities[:index] {
		if ability.ID == id {
			return true
		}
	}
	return false
}

func (s *GameState) characterCanProvideAdventureAbility(ownerID faction.FactionID, character *army.Commander) bool {
	if s == nil || ownerID == "" || character == nil || character.ID == "" {
		return false
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated ||
		character.OwnerID != string(ownerID) || !character.ActiveInYear(s.Year) ||
		character.ActiveMissionID != "" || s.DeadCommanderIDs[character.ID] ||
		s.DismissedCommanderIDs[character.ID] || s.ExiledCommanderIDs[character.ID] ||
		character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return false
	}
	switch character.Type {
	case army.CharacterTypeAdministrator, army.CharacterTypeCharacter, army.CharacterTypeCommander:
		return true
	default:
		return false
	}
}
