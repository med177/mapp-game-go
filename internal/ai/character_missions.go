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
	for _, character := range characters {
		if character == nil || len(character.PendingAdventureAbilities) == 0 {
			continue
		}
		abilityID := aiBestAdventureAbility(gs, fid, character)
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
			target := aiCharacterMissionTarget(gs, fid, mission)
			if mission.TargetKind != scenario.CharacterMissionTargetNone && target == "" {
				continue
			}
			if ok, _ := gs.CanStartCharacterMissionTarget(fid, character.ID, mission.ID, target); !ok {
				continue
			}
			net := aiCharacterMissionNetValue(gs, fid, character, mission, target, context)
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

func aiBestAdventureAbility(gs *state.GameState, fid faction.FactionID, character *army.Commander) string {
	if gs == nil || character == nil {
		return ""
	}
	bestID := ""
	bestValue := float64(-1)
	for _, ability := range character.PendingAdventureAbilities {
		value := aiAdventureAbilityValue(gs, fid, ability.Scope, ability.Effect, ability.Modifier)
		if bestID == "" || value > bestValue {
			bestID = ability.ID
			bestValue = value
		}
	}
	return bestID
}

func aiCharacterMissionTarget(gs *state.GameState, fid faction.FactionID, mission scenario.CharacterMissionDefinition) string {
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
				value = aiFactionArmyStrength(gs, faction.FactionID(targetID))
			}
		case scenario.CharacterMissionTargetRegion:
			value = aiRegionArmyStrength(gs, targetID)
		}
		if value > bestValue {
			bestTarget = targetID
			bestValue = value
		}
	}
	return bestTarget
}

func aiCharacterMissionNetValue(
	gs *state.GameState,
	fid faction.FactionID,
	character *army.Commander,
	mission scenario.CharacterMissionDefinition,
	targetID string,
	context aiCouncilValueContext,
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
			aiAdventureAbilityChoicesValue(gs, fid, character, adventure.SuccessAbilityChoices)
		partialValue := float64(adventure.PartialGoldReward) +
			aiMissionExperienceValue(gs, fid, character, mission.RoleID, adventure.PartialCouncilExperience, context) +
			aiAdventureAbilityChoicesValue(gs, fid, character, adventure.PartialAbilityChoices)
		value += float64(success)*successValue/100 + float64(partial)*partialValue/100
		value -= float64(failure*adventure.FailureUnavailableTurns) * 5 / 100
	} else {
		value = float64(mission.CompletionEffects.GoldReward)
		value += aiMissionExperienceValue(gs, fid, character, mission.RoleID, mission.CompletionEffects.CouncilExperience, context)
		if mission.CompletionEffects.RelationBonusPoints > 0 && mission.TargetKind == scenario.CharacterMissionTargetFaction {
			value += float64(mission.CompletionEffects.RelationBonusPoints) * 2
		}
		if mission.CompletionEffects.MilitaryIntelTurns > 0 {
			value += aiFactionArmyStrength(gs, faction.FactionID(targetID)) *
				float64(mission.CompletionEffects.MilitaryIntelTurns) * 0.1
		}
		if mission.CompletionEffects.RegionIntelTurns > 0 {
			value += aiRegionArmyStrength(gs, targetID) *
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

func aiAdventureAbilityChoicesValue(gs *state.GameState, fid faction.FactionID, character *army.Commander, choices []scenario.AdventureAbilityChoice) float64 {
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
		value := aiAdventureAbilityValue(gs, fid, choice.Scope, choice.Effect, choice.Modifier)
		if value > bestValue {
			bestValue = value
		}
	}
	return bestValue
}

func aiAdventureAbilityValue(gs *state.GameState, fid faction.FactionID, scope, effect string, modifier float64) float64 {
	if gs == nil || modifier <= 0 {
		return 0
	}
	switch {
	case scope == "faction" && effect == "military_power":
		return aiFactionArmyStrength(gs, fid) * modifier * aiMissionAbilityHorizon
	case scope == "army" && effect == "army_power":
		return aiFactionArmyStrengthByType(gs, fid, false) * modifier * aiMissionAbilityHorizon
	case scope == "navy" && effect == "navy_power":
		return aiFactionArmyStrengthByType(gs, fid, true) * modifier * aiMissionAbilityHorizon
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

func aiFactionArmyStrength(gs *state.GameState, fid faction.FactionID) float64 {
	return aiFactionArmyStrengthByType(gs, fid, false) + aiFactionArmyStrengthByType(gs, fid, true)
}

func aiFactionArmyStrengthByType(gs *state.GameState, fid faction.FactionID, naval bool) float64 {
	if gs == nil {
		return 0
	}
	var strength float64
	for _, currentArmy := range aiSortedArmies(gs) {
		if currentArmy.OwnerID != string(fid) || currentArmy.IsNaval != naval {
			continue
		}
		strength += float64(currentArmy.TotalStrength(gs.UnitTypes))
	}
	return strength
}

func aiRegionArmyStrength(gs *state.GameState, regionID string) float64 {
	if gs == nil {
		return 0
	}
	var strength float64
	for _, currentArmy := range aiSortedArmies(gs) {
		if string(currentArmy.RegionID) == regionID {
			strength += float64(currentArmy.TotalStrength(gs.UnitTypes))
		}
	}
	return strength
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
	if mission.TargetKind == scenario.CharacterMissionTargetFaction {
		message = characterName + " " + missionName + " görevine " + turnFactionName(gs, faction.FactionID(targetID)) + " hedefine gönderiliyor."
	} else if mission.TargetKind == scenario.CharacterMissionTargetRegion {
		message = characterName + " " + missionName + " görevine " + turnRegionName(gs, world.RegionID(targetID)) + " hedefine gönderiliyor."
	}
	return message
}
