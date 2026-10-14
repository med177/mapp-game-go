package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
)

// NeutralCharacterOwnerID senaryo verisinde henüz bir fraksiyona bağlı olmayan
// karakterlerin owner_id değeridir.
const NeutralCharacterOwnerID = "none"

func isRosterCharacter(character *army.Commander) bool {
	return character != nil && (character.Type == army.CharacterTypeAdministrator ||
		character.Type == army.CharacterTypeCharacter || character.Type == army.CharacterTypeCommander)
}

func isHouseholdDutyCharacter(character *army.Commander) bool {
	return character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter)
}

func characterDutySortPriority(character *army.Commander, roles []scenario.CouncilRoleDefinition) int {
	if character.AssignedCouncilRole == "" {
		if character.AssignedArmyID != "" {
			return len(roles) + 1
		}
		return len(roles) + 2
	}
	for index, role := range roles {
		if role.ID == character.AssignedCouncilRole {
			return index
		}
	}
	return len(roles)
}

// CharacterRoster bir fraksiyona ait aktif hane karakterleri ve askerî
// komutanları görev önceliği, ardından ada göre deterministik sırada döner.
func (s *GameState) CharacterRoster(ownerID string) []*army.Commander {
	if s == nil || ownerID == "" {
		return nil
	}
	roster := make([]*army.Commander, 0)
	for _, character := range s.Commanders {
		if !isRosterCharacter(character) || character.OwnerID != ownerID || !character.ActiveInYear(s.Year) ||
			s.DeadCommanderIDs[character.ID] || s.DismissedCommanderIDs[character.ID] {
			continue
		}
		roster = append(roster, character)
	}
	roles := s.councilConfig().Roles
	sort.Slice(roster, func(i, j int) bool {
		priorityI, priorityJ := characterDutySortPriority(roster[i], roles), characterDutySortPriority(roster[j], roles)
		if priorityI != priorityJ {
			return priorityI < priorityJ
		}
		if roster[i].Name == roster[j].Name {
			return roster[i].ID < roster[j].ID
		}
		return roster[i].Name < roster[j].Name
	})
	return roster
}

// NeutralCharacterCandidates tüm fraksiyonların hanesine katılabileceği
// tarafsız havuzdaki aktif ve şu anda başka bir göreve bağlı olmayan kayıtları döner.
func (s *GameState) NeutralCharacterCandidates() []*army.Commander {
	if s == nil {
		return nil
	}
	candidates := make([]*army.Commander, 0)
	for _, character := range s.Commanders {
		if !isHouseholdDutyCharacter(character) || character.OwnerID != NeutralCharacterOwnerID ||
			!character.ActiveInYear(s.Year) || character.AssignedArmyID != "" || character.AssignedCouncilRole != "" ||
			s.DeadCommanderIDs[character.ID] || s.DismissedCommanderIDs[character.ID] ||
			s.ExiledCommanderIDs[character.ID] || character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
			continue
		}
		candidates = append(candidates, character)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Name == candidates[j].Name {
			return candidates[i].ID < candidates[j].ID
		}
		return candidates[i].Name < candidates[j].Name
	})
	return candidates
}

// AdoptNeutralCharacter oyuncunun fraksiyonuna katma işlemini kolaylaştıran
// uyumluluk sarmalayıcısıdır.
func (s *GameState) AdoptNeutralCharacter(characterID string) bool {
	if s == nil {
		return false
	}
	return s.AdoptNeutralCharacterForFaction(s.PlayerFactionID, characterID)
}

// AdoptNeutralCharacterForFaction, yeterli altın varsa tarafsız karakteri
// belirtilen fraksiyona bağlar ve tek seferlik katılım ücretini tahsil eder.
func (s *GameState) AdoptNeutralCharacterForFaction(ownerID faction.FactionID, characterID string) bool {
	if s == nil || characterID == "" || ownerID == "" || ownerID == faction.FactionID(NeutralCharacterOwnerID) {
		return false
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return false
	}
	character := s.Commanders[characterID]
	if !isHouseholdDutyCharacter(character) || character.OwnerID != NeutralCharacterOwnerID ||
		!character.ActiveInYear(s.Year) || character.AssignedArmyID != "" || character.AssignedCouncilRole != "" ||
		s.DeadCommanderIDs[characterID] || s.DismissedCommanderIDs[characterID] ||
		s.ExiledCommanderIDs[characterID] || character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) ||
		character.AdoptionCostGold < 0 || character.GoldUpkeep < 0 || owner.Gold < character.AdoptionCostGold {
		return false
	}
	owner.Gold -= character.AdoptionCostGold
	character.OwnerID = string(ownerID)
	return true
}

// FactionCharacterGoldUpkeep returns the per-turn gold upkeep of a faction's
// active household characters, including members temporarily unavailable for duties.
func (s *GameState) FactionCharacterGoldUpkeep(ownerID string) int {
	if s == nil || ownerID == "" {
		return 0
	}
	councilConfig := s.councilConfig()
	total := 0
	for _, character := range s.Commanders {
		if !isHouseholdDutyCharacter(character) || character.OwnerID != ownerID ||
			!character.ActiveInYear(s.Year) || s.DeadCommanderIDs[character.ID] ||
			s.DismissedCommanderIDs[character.ID] || character.GoldUpkeep <= 0 {
			continue
		}
		if role, ok := councilConfig.Role(character.AssignedCouncilRole); ok && role.NoCost {
			continue
		}
		total += character.GoldUpkeep
	}
	return total
}

// ReleaseCharacterToNeutralPool aktif karakterin atamalarını kaldırıp tarafsız
// aday havuzuna geri bırakır.
func (s *GameState) ReleaseCharacterToNeutralPool(ownerID faction.FactionID, characterID string) bool {
	if s == nil || ownerID == "" || ownerID == faction.FactionID(NeutralCharacterOwnerID) ||
		characterID == "" || s.Factions[ownerID] == nil {
		return false
	}
	character := s.Commanders[characterID]
	if !isHouseholdDutyCharacter(character) || character.OwnerID != string(ownerID) ||
		!character.ActiveInYear(s.Year) || s.DeadCommanderIDs[characterID] ||
		s.DismissedCommanderIDs[characterID] || s.ExiledCommanderIDs[characterID] ||
		character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return false
	}
	s.moveCharacterToNeutralPool(character)
	return true
}

// ReleaseFactionCharactersToNeutralPool returns living household characters to
// the shared pool when their faction is eliminated. Captive or temporarily
// unavailable characters remain unavailable until their status clears.
func (s *GameState) ReleaseFactionCharactersToNeutralPool(ownerID faction.FactionID) int {
	if s == nil || ownerID == "" || ownerID == faction.FactionID(NeutralCharacterOwnerID) {
		return 0
	}
	released := 0
	for _, character := range s.Commanders {
		if !isHouseholdDutyCharacter(character) || character.OwnerID != string(ownerID) ||
			!character.ActiveInYear(s.Year) || s.DeadCommanderIDs[character.ID] ||
			s.DismissedCommanderIDs[character.ID] {
			continue
		}
		s.moveCharacterToNeutralPool(character)
		released++
	}
	return released
}

func (s *GameState) moveCharacterToNeutralPool(character *army.Commander) {
	if s == nil || character == nil {
		return
	}
	for _, currentArmy := range s.Armies {
		if currentArmy == nil {
			continue
		}
		if currentArmy.Commander == character || (currentArmy.Commander != nil && currentArmy.Commander.ID == character.ID) {
			currentArmy.Commander = nil
		}
		if currentArmy.EmbarkedCommander == character || (currentArmy.EmbarkedCommander != nil && currentArmy.EmbarkedCommander.ID == character.ID) {
			currentArmy.EmbarkedCommander = nil
		}
	}
	character.AssignedArmyID = ""
	character.AssignedCouncilRole = ""
	character.OwnerID = NeutralCharacterOwnerID
}
