package game

import (
	"fmt"
	"strings"

	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/scenario"
)

func (g *Game) resolveCharacterMissionCompletions() {
	if g == nil || g.gs == nil {
		return
	}
	completed := g.gs.TickCharacterMissions()
	if len(completed) == 0 {
		return
	}
	for _, result := range completed {
		effects := result.Mission.CompletionEffects
		owner := g.gs.Factions[result.OwnerID]
		character := g.gs.Commanders[result.CharacterID]
		isAdventure := result.Mission.Type == scenario.CharacterMissionTypeAdventure && result.Mission.Adventure != nil
		adventureMessage := ""
		if isAdventure {
			adventure := result.Mission.Adventure
			switch result.AdventureOutcome {
			case scenario.CharacterAdventureSuccess:
				if owner != nil {
					owner.Gold += adventure.SuccessGoldReward
				}
				if character != nil {
					character.RecordCouncilExperience(adventure.SuccessCouncilExperience)
				}
				adventureMessage = "başarıyla tamamlandı"
			case scenario.CharacterAdventurePartial:
				if owner != nil {
					owner.Gold += adventure.PartialGoldReward
				}
				if character != nil {
					character.RecordCouncilExperience(adventure.PartialCouncilExperience)
				}
				adventureMessage = "kısmi başarıyla tamamlandı"
			default:
				if character != nil && adventure.FailureUnavailableTurns > 0 {
					g.gs.InjureCommander(character.ID, adventure.FailureUnavailableTurns, "Macera sırasında yaralandı")
				}
				adventureMessage = "başarısız oldu"
			}
		} else if owner != nil {
			owner.Gold += effects.GoldReward
		}

		relationRewardApplied := false
		if !isAdventure && effects.RelationBonusPoints != 0 && owner != nil && result.TargetFaction != "" &&
			g.gs.Factions[result.TargetFaction] != nil {
			diplomacy.EnsureRelation(g.gs, result.OwnerID, result.TargetFaction)
			diplomacy.AddRelationScoreBoth(g.gs, result.OwnerID, result.TargetFaction, effects.RelationBonusPoints)
			relationRewardApplied = true
		}
		if !isAdventure && effects.MilitaryIntelTurns > 0 && owner != nil && result.TargetFaction != "" &&
			g.gs.Factions[result.TargetFaction] != nil {
			g.gs.GrantTemporaryMilitaryIntel(result.OwnerID, result.TargetFaction, effects.MilitaryIntelTurns)
		}
		if !isAdventure && effects.RegionIntelTurns > 0 && result.TargetRegion != "" && g.gs.Regions[result.TargetRegion] != nil {
			g.gs.GrantTemporaryRegionMilitaryIntel(result.OwnerID, result.TargetRegion, effects.RegionIntelTurns)
		}

		if !isAdventure && character != nil {
			character.RecordCouncilExperience(effects.CouncilExperience)
		}
		if result.OwnerID != g.gs.PlayerFactionID || g.renderer == nil {
			continue
		}

		characterName := result.CharacterName
		if character != nil && character.Name != "" {
			characterName = character.Name
		}
		if characterName == "" {
			characterName = result.CharacterID
		}
		missionName := result.Mission.NameTR
		if missionName == "" {
			missionName = result.Mission.ID
		}
		message := fmt.Sprintf("%s görevini tamamladı: %s", characterName, missionName)
		if isAdventure {
			message = fmt.Sprintf("%s macerası %s: %s", characterName, adventureMessage, missionName)
		}
		if result.TargetRegion != "" {
			if region := g.gs.Regions[result.TargetRegion]; region != nil && region.NameTR != "" {
				message += " (" + region.NameTR + ")"
			}
		}
		var rewards []string
		if relationRewardApplied {
			rewards = append(rewards, fmt.Sprintf("+%d ilişki puanı", effects.RelationBonusPoints))
		}
		if isAdventure && owner != nil {
			reward := 0
			switch result.AdventureOutcome {
			case scenario.CharacterAdventureSuccess:
				reward = result.Mission.Adventure.SuccessGoldReward
			case scenario.CharacterAdventurePartial:
				reward = result.Mission.Adventure.PartialGoldReward
			}
			if reward > 0 {
				rewards = append(rewards, fmt.Sprintf("+%d altın", reward))
			}
		} else if owner != nil && effects.GoldReward > 0 {
			rewards = append(rewards, fmt.Sprintf("+%d altın", effects.GoldReward))
		}
		if isAdventure && character != nil {
			experience := 0
			switch result.AdventureOutcome {
			case scenario.CharacterAdventureSuccess:
				experience = result.Mission.Adventure.SuccessCouncilExperience
			case scenario.CharacterAdventurePartial:
				experience = result.Mission.Adventure.PartialCouncilExperience
			}
			if experience > 0 {
				rewards = append(rewards, fmt.Sprintf("+%d konsey XP", experience))
			}
		} else if character != nil && effects.CouncilExperience > 0 {
			rewards = append(rewards, fmt.Sprintf("+%d konsey XP", effects.CouncilExperience))
		}
		if isAdventure && character != nil {
			if result.AdventureOutcome == scenario.CharacterAdventureFailure && result.Mission.Adventure.FailureUnavailableTurns > 0 {
				rewards = append(rewards, fmt.Sprintf("%d tur yaralı", result.Mission.Adventure.FailureUnavailableTurns))
			}
			if len(character.PendingAdventureAbilities) > 0 {
				rewards = append(rewards, "yetenek seçimi bekliyor")
			}
		}
		if effects.MilitaryIntelTurns > 0 && result.TargetFaction != "" {
			rewards = append(rewards, fmt.Sprintf("%d tur askerî istihbarat", effects.MilitaryIntelTurns))
		}
		if effects.RegionIntelTurns > 0 && result.TargetRegion != "" {
			rewards = append(rewards, fmt.Sprintf("%d tur bölgesel askerî istihbarat", effects.RegionIntelTurns))
		}
		if len(rewards) > 0 {
			message += " (" + strings.Join(rewards, ", ") + ")"
		}
		message += "."
		g.renderer.ShowCombatResult(message)
		g.renderer.AddEvent("[KARAKTER GÖREVİ] " + message)
	}
	if g.renderer != nil {
		g.renderer.RefreshCharacterRosterPanel()
	}
}
