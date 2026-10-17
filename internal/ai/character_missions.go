package ai

import (
	"math"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const aiMissionAbilityHorizon = 5

type aiCharacterMissionChoice struct {
	mission scenario.CharacterMissionDefinition
	target  string
	net     float64
}

func aiManageCharacterMissions(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || fid == "" || (fid == gs.PlayerFactionID && !gs.AIControlsPlayerFaction) {
		return
	}
	owner := gs.Factions[fid]
	if owner == nil || owner.IsEliminated {
		return
	}

	characters := gs.CharacterRoster(string(fid))
	sort.Slice(characters, func(i, j int) bool { return characters[i].ID < characters[j].ID })
	if len(characters) == 0 {
		return
	}
	var armyStrengths *aiCharacterMissionArmyStrengths
	getArmyStrengths := func() *aiCharacterMissionArmyStrengths {
		if armyStrengths == nil {
			armyStrengths = newAICharacterMissionArmyStrengths(gs)
		}
		return armyStrengths
	}
	for _, character := range characters {
		if character == nil || len(character.PendingAdventureAbilities) == 0 {
			continue
		}
		var abilityArmyStrengths *aiCharacterMissionArmyStrengths
		if aiPendingAdventureAbilitiesUseArmyStrength(character.PendingAdventureAbilities) {
			abilityArmyStrengths = getArmyStrengths()
		}
		abilityID := aiBestAdventureAbility(gs, fid, character, abilityArmyStrengths)
		if abilityID == "" {
			addTurnStep(steps, TurnStep{
				FactionID: fid,
				Kind:      TurnStepInfo,
				Message:   character.Name + " için macera yeteneği seçilemedi: geçerli seçenek bulunamadı.",
			})
			continue
		}
		if chosen, reason := gs.ChooseCharacterAdventureAbilityForFaction(fid, character.ID, abilityID); !chosen {
			addTurnStep(steps, TurnStep{
				FactionID: fid,
				Kind:      TurnStepInfo,
				Message:   character.Name + " için macera yeteneği seçilemedi: " + reason,
			})
		}
	}

	context := aiNewCouncilValueContext(gs, fid, nil)
	for _, character := range characters {
		if character == nil || owner.Gold < aiMinGoldReserve {
			continue
		}
		var best aiCharacterMissionChoice
		found := false
		for _, mission := range gs.CharacterMissionOptions(fid, character.ID) {
			if aiMissionUsesArmyStrength(mission) {
				getArmyStrengths()
			}
			target := aiCharacterMissionTarget(gs, fid, mission, armyStrengths)
			if mission.TargetKind != scenario.CharacterMissionTargetNone && target == "" {
				continue
			}
			if ok, _ := gs.CanStartCharacterMissionTarget(fid, character.ID, mission.ID, target); !ok {
				continue
			}
			net := aiCharacterMissionNetValue(gs, fid, character, mission, target, context, armyStrengths)
			if net <= 0 {
				continue
			}
			if !found || net > best.net || (net == best.net && mission.ID < best.mission.ID) {
				best = aiCharacterMissionChoice{mission: mission, target: target, net: net}
				found = true
			}
		}
		if !found || owner.Gold-best.mission.GoldCost < aiMinGoldReserve {
			continue
		}
		started, reason := gs.StartCharacterMissionTarget(fid, character.ID, best.mission.ID, best.target)
		if !started {
			addTurnStep(steps, TurnStep{
				FactionID: fid,
				Kind:      TurnStepInfo,
				Message:   character.Name + " için " + best.mission.ID + " görevi başlatılamadı: " + reason,
			})
			continue
		}
		addTurnStep(steps, TurnStep{
			FactionID: fid,
			Kind:      TurnStepInfo,
			Message:   aiCharacterMissionStartMessage(gs, character, best.mission, best.target),
		})
	}
}

func aiBestAdventureAbility(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	armyStrengths *aiCharacterMissionArmyStrengths,
) string {
	if gs == nil || character == nil {
		return ""
	}
	bestID := ""
	bestValue := float64(-1)
	for _, ability := range character.PendingAdventureAbilities {
		value := aiAdventureAbilityValue(gs, fid, ability.Scope, ability.Effect, ability.Modifier, armyStrengths)
		if bestID == "" || value > bestValue {
			bestID = ability.ID
			bestValue = value
		}
	}
	return bestID
}

func aiCharacterMissionTarget(
	gs *state.GameState,
	fid faction.FactionID,
	mission scenario.CharacterMissionDefinition,
	armyStrengths *aiCharacterMissionArmyStrengths,
) string {
	if gs == nil || mission.TargetKind == scenario.CharacterMissionTargetNone {
		return ""
	}
	targets := gs.CharacterMissionTargetIDs(fid, mission.ID)
	if len(targets) == 0 {
		return ""
	}
	sort.Strings(targets)
	bestTarget := targets[0]
	bestValue := math.Inf(-1)
	for _, targetID := range targets {
		value := float64(0)
		switch mission.TargetKind {
		case scenario.CharacterMissionTargetFaction:
			if mission.CompletionEffects.RelationBonusPoints > 0 {
				score, _ := relationScore(gs, string(fid), targetID)
				value = float64(-score)
			} else {
				value = armyStrengths.factionStrength(faction.FactionID(targetID))
			}
		case scenario.CharacterMissionTargetRegion:
			value = armyStrengths.regionStrength(targetID)
		}
		if value > bestValue {
			bestTarget = targetID
			bestValue = value
		}
	}
	return bestTarget
}

func aiMissionUsesArmyStrength(mission scenario.CharacterMissionDefinition) bool {
	if mission.TargetKind == scenario.CharacterMissionTargetRegion ||
		(mission.TargetKind == scenario.CharacterMissionTargetFaction && mission.CompletionEffects.RelationBonusPoints <= 0) ||
		mission.CompletionEffects.MilitaryIntelTurns > 0 ||
		mission.CompletionEffects.RegionIntelTurns > 0 {
		return true
	}
	if mission.Adventure == nil {
		return false
	}
	for _, choice := range mission.Adventure.SuccessAbilityChoices {
		if aiAdventureAbilityUsesArmyStrength(choice.Scope, choice.Effect) {
			return true
		}
	}
	for _, choice := range mission.Adventure.PartialAbilityChoices {
		if aiAdventureAbilityUsesArmyStrength(choice.Scope, choice.Effect) {
			return true
		}
	}
	return false
}

func aiAdventureAbilityUsesArmyStrength(scope, effect string) bool {
	return (scope == "faction" && effect == "military_power") ||
		(scope == "army" && effect == "army_power") ||
		(scope == "navy" && effect == "navy_power")
}

func aiPendingAdventureAbilitiesUseArmyStrength(abilities []army.CharacterAbility) bool {
	for _, ability := range abilities {
		if aiAdventureAbilityUsesArmyStrength(ability.Scope, ability.Effect) {
			return true
		}
	}
	return false
}

func aiCharacterMissionNetValue(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	mission scenario.CharacterMissionDefinition,
	targetID string,
	context aiCouncilValueContext,
	armyStrengths *aiCharacterMissionArmyStrengths,
) float64 {
	if gs == nil || character == nil {
		return 0
	}
	isAdventure := mission.Type == scenario.CharacterMissionTypeAdventure && mission.Adventure != nil
	value := float64(0)
	if isAdventure {
		success, partial, failure := gs.CharacterAdventureChances(character, mission.Adventure)
		adventure := mission.Adventure
		successValue := float64(adventure.SuccessGoldReward) +
			aiMissionExperienceValue(gs, fid, character, mission.RoleID, adventure.SuccessCouncilExperience, context) +
			aiAdventureAbilityChoicesValue(gs, fid, character, adventure.SuccessAbilityChoices, armyStrengths)
		partialValue := float64(adventure.PartialGoldReward) +
			aiMissionExperienceValue(gs, fid, character, mission.RoleID, adventure.PartialCouncilExperience, context) +
			aiAdventureAbilityChoicesValue(gs, fid, character, adventure.PartialAbilityChoices, armyStrengths)
		value += float64(success)*successValue/100 + float64(partial)*partialValue/100
		value -= float64(failure*adventure.FailureUnavailableTurns) * 5 / 100
	} else {
		value = float64(mission.CompletionEffects.GoldReward)
		value += aiMissionExperienceValue(gs, fid, character, mission.RoleID, mission.CompletionEffects.CouncilExperience, context)
		if mission.CompletionEffects.RelationBonusPoints > 0 && mission.TargetKind == scenario.CharacterMissionTargetFaction {
			value += float64(mission.CompletionEffects.RelationBonusPoints) * 2
		}
		if mission.CompletionEffects.MilitaryIntelTurns > 0 {
			value += armyStrengths.factionStrength(faction.FactionID(targetID)) *
				float64(mission.CompletionEffects.MilitaryIntelTurns) * 0.1
		}
		if mission.CompletionEffects.RegionIntelTurns > 0 {
			value += armyStrengths.regionStrength(targetID) *
				float64(mission.CompletionEffects.RegionIntelTurns) * 0.1
		}
	}
	value -= float64(mission.GoldCost)
	value -= aiMissionCouncilOpportunityCost(gs, fid, character, mission, context) * float64(mission.DurationTurns)
	return value
}

func aiMissionExperienceValue(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	roleID string,
	experience int,
	context aiCouncilValueContext,
) float64 {
	if gs == nil || character == nil || experience <= 0 {
		return 0
	}
	roles := gs.CouncilConfig.WithDefaults().Roles
	if roleID != "" {
		for _, role := range roles {
			if role.ID == roleID {
				roles = []scenario.CouncilRoleDefinition{role}
				break
			}
		}
	}
	improved := *character
	improved.RecordCouncilExperience(experience)
	bestDelta := 0
	for _, role := range roles {
		before := aiCouncilRoleValue(gs, role.ID, character, gs.CouncilRoleEffectsForCharacter(character, role.ID), context)
		after := aiCouncilRoleValue(gs, role.ID, &improved, gs.CouncilRoleEffectsForCharacter(&improved, role.ID), context)
		if delta := after - before; delta > bestDelta {
			bestDelta = delta
		}
	}
	return float64(bestDelta * aiMissionAbilityHorizon)
}

func aiAdventureAbilityChoicesValue(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	choices []scenario.AdventureAbilityChoice,
	armyStrengths *aiCharacterMissionArmyStrengths,
) float64 {
	if gs == nil || character == nil {
		return 0
	}
	learned := make(map[string]struct{}, len(character.AdventureAbilities))
	for _, ability := range character.AdventureAbilities {
		learned[ability.ID] = struct{}{}
	}
	bestValue := float64(0)
	for _, choice := range choices {
		if _, exists := learned[choice.ID]; exists {
			continue
		}
		value := aiAdventureAbilityValue(gs, fid, choice.Scope, choice.Effect, choice.Modifier, armyStrengths)
		if value > bestValue {
			bestValue = value
		}
	}
	return bestValue
}

func aiAdventureAbilityValue(
	gs *state.GameState,
	fid faction.FactionID,
	scope, effect string,
	modifier float64,
	armyStrengths *aiCharacterMissionArmyStrengths,
) float64 {
	if gs == nil || modifier <= 0 {
		return 0
	}
	switch {
	case scope == "faction" && effect == "military_power":
		return armyStrengths.factionStrength(fid) * modifier * aiMissionAbilityHorizon
	case scope == "army" && effect == "army_power":
		return armyStrengths.factionStrengthByType(fid, false) * modifier * aiMissionAbilityHorizon
	case scope == "navy" && effect == "navy_power":
		return armyStrengths.factionStrengthByType(fid, true) * modifier * aiMissionAbilityHorizon
	default:
		return 0
	}
}

func aiMissionCouncilOpportunityCost(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	mission scenario.CharacterMissionDefinition,
	context aiCouncilValueContext,
) float64 {
	if gs == nil || character == nil || mission.RoleID == "" || character.AssignedCouncilRole != mission.RoleID {
		return 0
	}
	effects := gs.CouncilRoleEffectsForCharacter(character, mission.RoleID)
	return float64(maxInt(0, aiCouncilRoleValue(gs, mission.RoleID, character, effects, context)))
}

type aiCharacterMissionArmyStrengths struct {
	landByFaction  map[faction.FactionID]float64
	navalByFaction map[faction.FactionID]float64
	byRegion       map[string]float64
}

func newAICharacterMissionArmyStrengths(gs *state.GameState) *aiCharacterMissionArmyStrengths {
	strengths := &aiCharacterMissionArmyStrengths{
		landByFaction:  make(map[faction.FactionID]float64),
		navalByFaction: make(map[faction.FactionID]float64),
		byRegion:       make(map[string]float64),
	}
	if gs == nil {
		return strengths
	}
	for _, currentArmy := range aiSortedArmies(gs) {
		fid := faction.FactionID(currentArmy.OwnerID)
		strength := float64(currentArmy.TotalStrength(gs.UnitTypes))
		if currentArmy.IsNaval {
			strengths.navalByFaction[fid] += strength
		} else {
			strengths.landByFaction[fid] += strength
		}
		strengths.byRegion[string(currentArmy.RegionID)] += strength
	}
	return strengths
}

func (s *aiCharacterMissionArmyStrengths) factionStrength(fid faction.FactionID) float64 {
	if s == nil {
		return 0
	}
	return s.factionStrengthByType(fid, false) + s.factionStrengthByType(fid, true)
}

func (s *aiCharacterMissionArmyStrengths) factionStrengthByType(fid faction.FactionID, naval bool) float64 {
	if s == nil {
		return 0
	}
	if naval {
		return s.navalByFaction[fid]
	}
	return s.landByFaction[fid]
}

func (s *aiCharacterMissionArmyStrengths) regionStrength(regionID string) float64 {
	if s == nil {
		return 0
	}
	return s.byRegion[regionID]
}

func aiCharacterMissionStartMessage(gs *state.GameState, character *army.Commander, mission scenario.CharacterMissionDefinition, targetID string) string {
	characterName := character.ID
	if character.Name != "" {
		characterName = character.Name
	}
	missionName := mission.NameTR
	if missionName == "" {
		missionName = mission.ID
	}
	message := characterName + " " + missionName + " görevine gönderiliyor."
	switch mission.TargetKind {
	case scenario.CharacterMissionTargetFaction:
		message = characterName + " " + missionName + " görevine " + turnFactionName(gs, faction.FactionID(targetID)) + " hedefine gönderiliyor."
	case scenario.CharacterMissionTargetRegion:
		message = characterName + " " + missionName + " görevine " + turnRegionName(gs, world.RegionID(targetID)) + " hedefine gönderiliyor."
	}
	return message
}
