package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

// CompletedCharacterMission is a resolved save-backed task result for the game
// layer to apply through canonical diplomacy/economy APIs.
type CompletedCharacterMission struct {
	OwnerID       faction.FactionID
	CharacterID   string
	CharacterName string
	Mission       scenario.CharacterMissionDefinition
	TargetFaction faction.FactionID
	TargetRegion  world.RegionID
}

func (s *GameState) characterMissionDefinition(missionID string) (scenario.CharacterMissionDefinition, bool) {
	if s == nil || missionID == "" {
		return scenario.CharacterMissionDefinition{}, false
	}
	for _, mission := range s.CouncilConfig.Missions {
		if mission.ID == missionID {
			return mission, true
		}
	}
	return scenario.CharacterMissionDefinition{}, false
}

func characterMissionRoleMatches(character *army.Commander, mission scenario.CharacterMissionDefinition) bool {
	if character == nil {
		return false
	}
	if mission.RoleID == "" {
		return character.AssignedCouncilRole == ""
	}
	return character.AssignedCouncilRole == mission.RoleID
}

func (s *GameState) missionCharacterAvailable(ownerID string, character *army.Commander, mission scenario.CharacterMissionDefinition) bool {
	return s != nil && ownerID != "" && character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter || character.Type == army.CharacterTypeCommander) &&
		character.OwnerID == ownerID && character.ActiveInYear(s.Year) && character.AssignedArmyID == "" &&
		character.ActiveMissionID == "" && characterMissionRoleMatches(character, mission) &&
		!s.DeadCommanderIDs[character.ID] && !s.DismissedCommanderIDs[character.ID] &&
		!s.ExiledCommanderIDs[character.ID] && character.CaptiveAtSettlementID == "" &&
		!s.CommanderUnavailable(character)
}

// CharacterMissionOptions returns JSON-configured tasks that fit the
// character's current role/assignment. Config order is preserved for UI.
func (s *GameState) CharacterMissionOptions(ownerID faction.FactionID, characterID string) []scenario.CharacterMissionDefinition {
	if s == nil || ownerID == "" {
		return nil
	}
	owner := s.Factions[ownerID]
	character := s.Commanders[characterID]
	if owner == nil || owner.IsEliminated || character == nil || !character.ActiveInYear(s.Year) ||
		character.AssignedArmyID != "" || character.ActiveMissionID != "" ||
		s.DeadCommanderIDs[characterID] || s.DismissedCommanderIDs[characterID] ||
		s.ExiledCommanderIDs[characterID] || character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return nil
	}
	options := make([]scenario.CharacterMissionDefinition, 0, len(s.CouncilConfig.Missions))
	for _, mission := range s.CouncilConfig.Missions {
		if characterMissionRoleMatches(character, mission) {
			options = append(options, mission)
		}
	}
	return options
}

// CharacterMissionTargetIDs returns eligible targets in deterministic display
// order for the configured mission target kind.
func (s *GameState) CharacterMissionTargetIDs(ownerID faction.FactionID, missionID string) []string {
	mission, ok := s.characterMissionDefinition(missionID)
	if !ok || s == nil || ownerID == "" {
		return nil
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return nil
	}
	switch mission.TargetKind {
	case scenario.CharacterMissionTargetFaction:
		targets := make([]string, 0, len(s.Factions))
		for targetID, target := range s.Factions {
			if target == nil || targetID == ownerID || target.IsEliminated || target.IsVirtual {
				continue
			}
			if relation := s.Relations[faction.RelationKey(ownerID, targetID)]; relation != nil && relation.Stance == faction.StanceWar {
				continue
			}
			targets = append(targets, string(targetID))
		}
		sort.Slice(targets, func(i, j int) bool {
			left, right := s.Factions[faction.FactionID(targets[i])], s.Factions[faction.FactionID(targets[j])]
			leftName, rightName := targets[i], targets[j]
			if left != nil && left.NameTR != "" {
				leftName = left.NameTR
			}
			if right != nil && right.NameTR != "" {
				rightName = right.NameTR
			}
			if leftName == rightName {
				return targets[i] < targets[j]
			}
			return leftName < rightName
		})
		return targets
	case scenario.CharacterMissionTargetRegion:
		targets := make([]string, 0, len(s.Regions))
		for regionID, region := range s.Regions {
			if region == nil || region.IsSea || region.IsTerrainArea || region.IsLocked ||
				region.OwnerID == "" || faction.FactionID(region.OwnerID) == ownerID {
				continue
			}
			targetOwner := s.Factions[faction.FactionID(region.OwnerID)]
			if targetOwner == nil || targetOwner.IsEliminated || targetOwner.IsVirtual {
				continue
			}
			if relation := s.Relations[faction.RelationKey(ownerID, faction.FactionID(region.OwnerID))]; relation != nil &&
				relation.Stance == faction.StanceWar {
				continue
			}
			targets = append(targets, string(regionID))
		}
		sort.Slice(targets, func(i, j int) bool {
			left, right := s.Regions[world.RegionID(targets[i])], s.Regions[world.RegionID(targets[j])]
			leftName, rightName := targets[i], targets[j]
			if left != nil && left.NameTR != "" {
				leftName = left.NameTR
			}
			if right != nil && right.NameTR != "" {
				rightName = right.NameTR
			}
			if leftName == rightName {
				return targets[i] < targets[j]
			}
			return leftName < rightName
		})
		return targets
	default:
		return nil
	}
}

// CharacterMissionTargets is retained for existing faction-target callers.
func (s *GameState) CharacterMissionTargets(ownerID faction.FactionID, missionID string) []faction.FactionID {
	mission, ok := s.characterMissionDefinition(missionID)
	if !ok || mission.TargetKind != scenario.CharacterMissionTargetFaction {
		return nil
	}
	ids := s.CharacterMissionTargetIDs(ownerID, missionID)
	targets := make([]faction.FactionID, len(ids))
	for i, id := range ids {
		targets[i] = faction.FactionID(id)
	}
	return targets
}

// CanStartCharacterMission validates a JSON-defined task without mutating
// state. A role-linked mission requires its incumbent; a role-free mission
// requires a character with no council assignment.
func (s *GameState) CanStartCharacterMission(ownerID faction.FactionID, characterID, missionID string, targetID faction.FactionID) (bool, string) {
	return s.CanStartCharacterMissionTarget(ownerID, characterID, missionID, string(targetID))
}

// CanStartCharacterMissionTarget validates a mission using its configured
// faction, region, or no-target input.
func (s *GameState) CanStartCharacterMissionTarget(ownerID faction.FactionID, characterID, missionID, targetID string) (bool, string) {
	if s == nil || ownerID == "" || characterID == "" || missionID == "" {
		return false, "Geçersiz karakter görevi."
	}
	owner := s.Factions[ownerID]
	mission, ok := s.characterMissionDefinition(missionID)
	character := s.Commanders[characterID]
	if owner == nil || owner.IsEliminated || !ok || !s.missionCharacterAvailable(string(ownerID), character, mission) {
		return false, "Karakter bu görev için uygun değil."
	}
	if owner.Gold < mission.GoldCost {
		return false, "Görev için yeterli altın yok."
	}
	switch mission.TargetKind {
	case scenario.CharacterMissionTargetNone:
		if targetID != "" {
			return false, "Bu görev hedef kabul etmiyor."
		}
	case scenario.CharacterMissionTargetFaction:
		targetFactionID := faction.FactionID(targetID)
		if targetFactionID == "" || targetFactionID == ownerID {
			return false, "Görev için başka bir fraksiyon seçilmeli."
		}
		target := s.Factions[targetFactionID]
		if target == nil || target.IsEliminated || target.IsVirtual {
			return false, "Görev hedefi bulunamadı."
		}
		if relation := s.Relations[faction.RelationKey(ownerID, targetFactionID)]; relation != nil && relation.Stance == faction.StanceWar {
			return false, "Savaş halindeki fraksiyona bu görev gönderilemez."
		}
	case scenario.CharacterMissionTargetRegion:
		region := s.Regions[world.RegionID(targetID)]
		if region == nil || region.IsSea || region.IsTerrainArea || region.IsLocked ||
			region.OwnerID == "" || faction.FactionID(region.OwnerID) == ownerID {
			return false, "Bölge görev hedefi geçersiz."
		}
		targetOwnerID := faction.FactionID(region.OwnerID)
		targetOwner := s.Factions[targetOwnerID]
		if targetOwner == nil || targetOwner.IsEliminated || targetOwner.IsVirtual {
			return false, "Görev bölgesinin sahibi bulunamadı."
		}
		if relation := s.Relations[faction.RelationKey(ownerID, targetOwnerID)]; relation != nil &&
			relation.Stance == faction.StanceWar {
			return false, "Savaş halindeki fraksiyonun bölgesine bu görev gönderilemez."
		}
	default:
		return false, "Görev hedef türü geçersiz."
	}
	return true, ""
}

// StartCharacterMission charges the JSON-configured one-time cost and stores
// the in-progress task on the character's existing save-backed commander entry.
func (s *GameState) StartCharacterMission(ownerID faction.FactionID, characterID, missionID string, targetID faction.FactionID) (bool, string) {
	return s.StartCharacterMissionTarget(ownerID, characterID, missionID, string(targetID))
}

// StartCharacterMissionTarget charges the configured cost and stores the
// target's string ID, which may identify either a faction or region.
func (s *GameState) StartCharacterMissionTarget(ownerID faction.FactionID, characterID, missionID, targetID string) (bool, string) {
	if ok, reason := s.CanStartCharacterMissionTarget(ownerID, characterID, missionID, targetID); !ok {
		return false, reason
	}
	mission, _ := s.characterMissionDefinition(missionID)
	character := s.Commanders[characterID]
	owner := s.Factions[ownerID]
	owner.Gold -= mission.GoldCost
	character.ActiveMissionID = mission.ID
	character.ActiveMissionTargetID = string(targetID)
	character.ActiveMissionTurnsRemaining = mission.DurationTurns
	return true, ""
}

// TickCharacterMissions advances each active mission once per resolved turn.
// It returns completed tasks in character ID order and clears their runtime state.
func (s *GameState) TickCharacterMissions() []CompletedCharacterMission {
	if s == nil || len(s.Commanders) == 0 {
		return nil
	}
	ids := make([]string, 0, len(s.Commanders))
	for id, character := range s.Commanders {
		if character != nil && character.ActiveMissionID != "" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	completed := make([]CompletedCharacterMission, 0)
	for _, id := range ids {
		character := s.Commanders[id]
		mission, ok := s.characterMissionDefinition(character.ActiveMissionID)
		if !ok {
			character.ActiveMissionID = ""
			character.ActiveMissionTargetID = ""
			character.ActiveMissionTurnsRemaining = 0
			continue
		}
		if character.ActiveMissionTurnsRemaining > 1 {
			character.ActiveMissionTurnsRemaining--
			continue
		}
		completed = append(completed, CompletedCharacterMission{
			OwnerID:       faction.FactionID(character.OwnerID),
			CharacterID:   character.ID,
			CharacterName: character.Name,
			Mission:       mission,
			TargetFaction: missionTargetFaction(mission, character.ActiveMissionTargetID),
			TargetRegion:  missionTargetRegion(mission, character.ActiveMissionTargetID),
		})
		character.ActiveMissionID = ""
		character.ActiveMissionTargetID = ""
		character.ActiveMissionTurnsRemaining = 0
	}
	return completed
}

func missionTargetFaction(mission scenario.CharacterMissionDefinition, targetID string) faction.FactionID {
	if mission.TargetKind != scenario.CharacterMissionTargetFaction {
		return ""
	}
	return faction.FactionID(targetID)
}

func missionTargetRegion(mission scenario.CharacterMissionDefinition, targetID string) world.RegionID {
	if mission.TargetKind != scenario.CharacterMissionTargetRegion {
		return ""
	}
	return world.RegionID(targetID)
}
