package game

import (
	"fmt"
	"strings"

	"mapp-game-go/internal/diplomacy"
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
		if owner != nil {
			owner.Gold += effects.GoldReward
		}

		relationRewardApplied := false
		if effects.RelationBonusPoints != 0 && owner != nil && result.TargetFaction != "" &&
			g.gs.Factions[result.TargetFaction] != nil {
			diplomacy.EnsureRelation(g.gs, result.OwnerID, result.TargetFaction)
			diplomacy.AddRelationScoreBoth(g.gs, result.OwnerID, result.TargetFaction, effects.RelationBonusPoints)
			relationRewardApplied = true
		}
		if effects.MilitaryIntelTurns > 0 && owner != nil && result.TargetFaction != "" &&
			g.gs.Factions[result.TargetFaction] != nil {
			g.gs.GrantTemporaryMilitaryIntel(result.OwnerID, result.TargetFaction, effects.MilitaryIntelTurns)
		}
		if effects.RegionIntelTurns > 0 && result.TargetRegion != "" && g.gs.Regions[result.TargetRegion] != nil {
			g.gs.GrantTemporaryRegionMilitaryIntel(result.OwnerID, result.TargetRegion, effects.RegionIntelTurns)
		}

		character := g.gs.Commanders[result.CharacterID]
		if character != nil {
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
		if result.TargetRegion != "" {
			if region := g.gs.Regions[result.TargetRegion]; region != nil && region.NameTR != "" {
				message += " (" + region.NameTR + ")"
			}
		}
		var rewards []string
		if relationRewardApplied {
			rewards = append(rewards, fmt.Sprintf("+%d ilişki puanı", effects.RelationBonusPoints))
		}
		if owner != nil && effects.GoldReward > 0 {
			rewards = append(rewards, fmt.Sprintf("+%d altın", effects.GoldReward))
		}
		if character != nil && effects.CouncilExperience > 0 {
			rewards = append(rewards, fmt.Sprintf("+%d konsey XP", effects.CouncilExperience))
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
