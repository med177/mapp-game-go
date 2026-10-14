package ai

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

// aiManageNeutralCharacters adopts at most one shared-pool candidate per turn
// when the faction has an unfilled council role.
func aiManageNeutralCharacters(gs *state.GameState, fid faction.FactionID) {
	if gs == nil || fid == "" || (fid == gs.PlayerFactionID && !gs.AIControlsPlayerFaction) {
		return
	}
	owner := gs.Factions[fid]
	if owner == nil || owner.IsEliminated {
		return
	}
	candidates := gs.NeutralCharacterCandidates()
	if len(candidates) == 0 {
		return
	}

	for _, role := range gs.CouncilConfig.WithDefaults().Roles {
		if gs.CouncilRoleHolder(fid, role.ID) != nil {
			continue
		}
		for _, candidate := range candidates {
			if candidate == nil || !gs.AdoptNeutralCharacterForFaction(fid, candidate.ID) {
				continue
			}
			if gs.AssignCharacterToCouncilRole(fid, candidate.ID, role.ID) {
				return
			}
			gs.ReleaseCharacterToNeutralPool(fid, candidate.ID)
		}
	}
}
