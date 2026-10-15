package state

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/world"
)

// CompletedCharacterMission is a resolved save-backed task result for the game
// layer to apply through canonical diplomacy/economy APIs.
type CompletedCharacterMission struct {
	OwnerID          faction.FactionID
	CharacterID      string
	CharacterName    string
	Mission          scenario.CharacterMissionDefinition
	TargetFaction    faction.FactionID
	TargetRegion     world.RegionID
	AdventureOutcome string
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

func characterMissionMatches(
	ownerID string,
	firedEventIDs map[string]bool,
	year int,
	character *army.Commander,
	mission scenario.CharacterMissionDefinition,
) bool {
	if character == nil || (mission.MinYear > 0 && year < mission.MinYear) {
		return false
	}
	if mission.RoleID == "" {
		if character.AssignedCouncilRole != "" && !mission.AllowCouncilRoles {
			return false
		}
	} else if character.AssignedCouncilRole != mission.RoleID {
		return false
	}
	if len(mission.AllowedCharacters) > 0 {
		allowed := false
		for _, allowedCharacterID := range mission.AllowedCharacters {
			if character.ID == allowedCharacterID {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	if len(mission.AllowedFactions) > 0 {
		allowed := false
		for _, allowedFactionID := range mission.AllowedFactions {
			if ownerID == allowedFactionID {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}
	for _, blockedFactionID := range mission.BlockedFactions {
		if ownerID == blockedFactionID {
			return false
		}
	}
	for _, flag := range mission.RequiresFlags {
		if !firedEventIDs["flag:"+flag] {
			return false
		}
	}
	for _, flag := range mission.BlocksFlags {
		if firedEventIDs["flag:"+flag] {
			return false
		}
	}
	return true
}

func (s *GameState) CharacterAdventureChances(character *army.Commander, adventure *scenario.CharacterAdventureDefinition) (success, partial, failure int) {
	if adventure == nil {
		return 0, 0, 100
	}
	success = adventure.SuccessChancePercent
	if character != nil {
		level := character.CouncilLevel
		if level <= 0 {
			level = army.CommanderStartingLevel
		}
		hasRelevantTrait := false
		for _, trait := range character.Traits {
			for _, relevant := range adventure.RelevantTraits {
				if strings.EqualFold(strings.TrimSpace(string(trait)), strings.TrimSpace(relevant)) {
					hasRelevantTrait = true
					break
				}
			}
			if hasRelevantTrait {
				break
			}
		}
		if hasRelevantTrait {
			success = addAdventureChance(success, adventure.TraitBonusPercent)
		}
		if adventure.LevelBonusPercent > 0 && success < 100 {
			remaining := 100 - success
			levelsNeeded := remaining / adventure.LevelBonusPercent
			if remaining%adventure.LevelBonusPercent != 0 {
				levelsNeeded++
			}
			if level >= levelsNeeded {
				success = 100
			} else {
				success += level * adventure.LevelBonusPercent
			}
		}
	}
	if success > 100 {
		success = 100
	}
	if success < 0 {
		success = 0
	}
	partial = adventure.PartialSuccessChancePercent
	if partial > 100-success {
		partial = 100 - success
	}
	failure = 100 - success - partial
	return success, partial, failure
}

func addAdventureChance(chance, bonus int) int {
	if bonus >= 100-chance {
		return 100
	}
	return chance + bonus
}

func characterAdventureRoll(s *GameState, character *army.Commander, missionID string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(s.ScenarioID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(character.OwnerID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(character.ID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(missionID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(strconv.Itoa(s.Turn)))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(strconv.Itoa(character.CouncilExperience)))
	return int(hash.Sum32() % 100)
}

func adventureOutcome(roll, success, partial int) string {
	if roll < success {
		return scenario.CharacterAdventureSuccess
	}
	if roll < success+partial {
		return scenario.CharacterAdventurePartial
	}
	return scenario.CharacterAdventureFailure
}

func availableAdventureAbilityChoices(character *army.Commander, choices []scenario.AdventureAbilityChoice) []army.CharacterAbility {
	if character == nil || len(choices) == 0 {
		return nil
	}
	learned := make(map[string]struct{}, len(character.AdventureAbilities))
	for _, ability := range character.AdventureAbilities {
		learned[ability.ID] = struct{}{}
	}
	available := make([]army.CharacterAbility, 0, len(choices))
	for _, choice := range choices {
		if _, exists := learned[choice.ID]; exists {
			continue
		}
		available = append(available, army.CharacterAbility{
			ID: choice.ID, NameTR: choice.NameTR, Scope: choice.Scope, Effect: choice.Effect, Modifier: choice.Modifier,
		})
	}
	return available
}

func (s *GameState) missionCharacterAvailable(ownerID string, character *army.Commander, mission scenario.CharacterMissionDefinition) bool {
	return s != nil && ownerID != "" && character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter || character.Type == army.CharacterTypeCommander) &&
		character.OwnerID == ownerID && character.ActiveInYear(s.Year) && character.AssignedArmyID == "" &&
		character.ActiveMissionID == "" && len(character.PendingAdventureAbilities) == 0 &&
		characterMissionMatches(ownerID, s.FiredEventIDs, s.Year, character, mission) &&
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
		len(character.PendingAdventureAbilities) > 0 ||
		s.DeadCommanderIDs[characterID] || s.DismissedCommanderIDs[characterID] ||
		s.ExiledCommanderIDs[characterID] || character.CaptiveAtSettlementID != "" || s.CommanderUnavailable(character) {
		return nil
	}
	options := make([]scenario.CharacterMissionDefinition, 0, len(s.CouncilConfig.Missions))
	for _, mission := range s.CouncilConfig.Missions {
		if characterMissionMatches(string(ownerID), s.FiredEventIDs, s.Year, character, mission) {
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
	character.ActiveMissionOutcomeRoll = 0
	character.ActiveMissionOutcome = ""
	if mission.Type == scenario.CharacterMissionTypeAdventure {
		character.ActiveMissionOutcomeRoll = characterAdventureRoll(s, character, mission.ID)
		success, partial, _ := s.CharacterAdventureChances(character, mission.Adventure)
		character.ActiveMissionOutcome = adventureOutcome(character.ActiveMissionOutcomeRoll, success, partial)
	}
	return true, ""
}

// ChooseCharacterAdventureAbility stores one earned ability and clears the
// pending choice so it cannot be applied twice.
func (s *GameState) ChooseCharacterAdventureAbility(ownerID faction.FactionID, characterID, abilityID string) (bool, string) {
	if s == nil || ownerID != s.PlayerFactionID {
		return false, "Geçersiz yetenek seçimi."
	}
	return s.ChooseCharacterAdventureAbilityForFaction(ownerID, characterID, abilityID)
}

// ChooseCharacterAdventureAbilityForFaction stores an earned ability for any
// faction controlled by the game, including factions managed by the AI.
func (s *GameState) ChooseCharacterAdventureAbilityForFaction(ownerID faction.FactionID, characterID, abilityID string) (bool, string) {
	if s == nil || ownerID == "" || characterID == "" || abilityID == "" {
		return false, "Geçersiz yetenek seçimi."
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return false, "Karakterin fraksiyonu bulunamadı."
	}
	character := s.Commanders[characterID]
	if character == nil || character.OwnerID != string(ownerID) {
		return false, "Karakter bu yeteneği seçemez."
	}
	for _, choice := range character.PendingAdventureAbilities {
		if choice.ID != abilityID {
			continue
		}
		for _, learned := range character.AdventureAbilities {
			if learned.ID == abilityID {
				return false, "Karakter bu yeteneğe zaten sahip."
			}
		}
		character.AdventureAbilities = append(character.AdventureAbilities, choice)
		character.PendingAdventureAbilities = nil
		return true, ""
	}
	return false, "Bu yetenek seçeneği artık kullanılamıyor."
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
			character.ActiveMissionOutcomeRoll = 0
			character.ActiveMissionOutcome = ""
			continue
		}
		if character.ActiveMissionTurnsRemaining > 1 {
			character.ActiveMissionTurnsRemaining--
			continue
		}
		result := CompletedCharacterMission{
			OwnerID:       faction.FactionID(character.OwnerID),
			CharacterID:   character.ID,
			CharacterName: character.Name,
			Mission:       mission,
			TargetFaction: missionTargetFaction(mission, character.ActiveMissionTargetID),
			TargetRegion:  missionTargetRegion(mission, character.ActiveMissionTargetID),
		}
		if mission.Type == scenario.CharacterMissionTypeAdventure && mission.Adventure != nil {
			result.AdventureOutcome = character.ActiveMissionOutcome
			if result.AdventureOutcome == "" {
				success, partial, _ := s.CharacterAdventureChances(character, mission.Adventure)
				result.AdventureOutcome = adventureOutcome(character.ActiveMissionOutcomeRoll, success, partial)
			}
			var choices []scenario.AdventureAbilityChoice
			switch result.AdventureOutcome {
			case scenario.CharacterAdventureSuccess:
				choices = mission.Adventure.SuccessAbilityChoices
			case scenario.CharacterAdventurePartial:
				choices = mission.Adventure.PartialAbilityChoices
			}
			character.PendingAdventureAbilities = availableAdventureAbilityChoices(character, choices)
		}
		completed = append(completed, result)
		character.ActiveMissionID = ""
		character.ActiveMissionTargetID = ""
		character.ActiveMissionTurnsRemaining = 0
		character.ActiveMissionOutcomeRoll = 0
		character.ActiveMissionOutcome = ""
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
