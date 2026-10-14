package state

import (
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

// NeutralCharacterOwnerID senaryo verisinde henüz bir fraksiyona bağlı olmayan
// karakterlerin owner_id değeridir.
const NeutralCharacterOwnerID = "none"

func isRosterCharacter(character *army.Commander) bool {
	return character != nil && (character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter)
}

func characterDutySortPriority(character *army.Commander) int {
	switch character.AssignedCouncilRole {
	case CouncilRoleHeadOfHouse:
		return 0
	case CouncilRoleTreasurer:
		return 1
	case CouncilRoleDiplomat:
		return 2
	case "":
		if character.AssignedArmyID != "" {
			return 3
		}
		return 5
	default:
		return 4
	}
}

// CharacterRoster bir fraksiyona ait aktif yönetici ve karakter kayıtlarını
// görev önceliği, ardından ada göre deterministik sırada döner. Askerî
// komutanlar kendi roster'larında kalır.
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
	sort.Slice(roster, func(i, j int) bool {
		priorityI, priorityJ := characterDutySortPriority(roster[i]), characterDutySortPriority(roster[j])
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

// NeutralCharacterCandidates tarafsız havuzdan oyuncunun fraksiyonuna
// katılabilecek, aktif ve şu anda başka bir göreve bağlı olmayan kayıtları döner.
func (s *GameState) NeutralCharacterCandidates() []*army.Commander {
	if s == nil {
		return nil
	}
	candidates := make([]*army.Commander, 0)
	for _, character := range s.Commanders {
		if !isRosterCharacter(character) || character.OwnerID != NeutralCharacterOwnerID ||
			!character.ActiveInYear(s.Year) || character.AssignedArmyID != "" ||
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

// AdoptNeutralCharacter tarafsız bir karakteri oyuncunun fraksiyonuna ücretsiz
// bağlar. Şablon değiştirilmez; sahiplik runtime karakter state'inde tutulur ve
// mevcut commanders save alanı üzerinden korunur.
func (s *GameState) AdoptNeutralCharacter(characterID string) bool {
	if s == nil || characterID == "" || s.PlayerFactionID == "" || s.Factions[s.PlayerFactionID] == nil {
		return false
	}
	character := s.Commanders[characterID]
	if !isRosterCharacter(character) || character.OwnerID != NeutralCharacterOwnerID ||
		!character.ActiveInYear(s.Year) || character.AssignedArmyID != "" ||
		s.DeadCommanderIDs[characterID] || s.DismissedCommanderIDs[characterID] ||
		s.ExiledCommanderIDs[characterID] || character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return false
	}
	character.OwnerID = string(s.PlayerFactionID)
	return true
}

// ReleaseCharacterToNeutralPool aktif karakterin atamalarını kaldırıp tarafsız
// aday havuzuna geri bırakır.
func (s *GameState) ReleaseCharacterToNeutralPool(ownerID faction.FactionID, characterID string) bool {
	if s == nil || ownerID == "" || ownerID == faction.FactionID(NeutralCharacterOwnerID) ||
		characterID == "" || s.Factions[ownerID] == nil {
		return false
	}
	character := s.Commanders[characterID]
	if !isRosterCharacter(character) || character.OwnerID != string(ownerID) ||
		!character.ActiveInYear(s.Year) || s.DeadCommanderIDs[characterID] ||
		s.DismissedCommanderIDs[characterID] || s.ExiledCommanderIDs[characterID] ||
		character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return false
	}
	for _, currentArmy := range s.Armies {
		if currentArmy == nil {
			continue
		}
		if currentArmy.Commander == character || (currentArmy.Commander != nil && currentArmy.Commander.ID == characterID) {
			currentArmy.Commander = nil
		}
		if currentArmy.EmbarkedCommander == character || (currentArmy.EmbarkedCommander != nil && currentArmy.EmbarkedCommander.ID == characterID) {
			currentArmy.EmbarkedCommander = nil
		}
	}
	character.AssignedArmyID = ""
	character.AssignedCouncilRole = ""
	character.OwnerID = NeutralCharacterOwnerID
	return true
}
