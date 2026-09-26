package render

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func displayedFactionPowerBreakdown(gs *state.GameState, fid faction.FactionID) (land, naval int, exact bool) {
	if gs == nil {
		return 0, 0, true
	}
	observer := gs.PlayerFactionID
	if observer == "" {
		return state.MilitaryPowerEstimateBreakdown(gs, fid, fid)
	}
	return state.MilitaryPowerEstimateBreakdown(gs, observer, fid)
}

func displayedFactionPower(gs *state.GameState, fid faction.FactionID) (power int, exact bool) {
	land, naval, exact := displayedFactionPowerBreakdown(gs, fid)
	return land + naval, exact
}

func perceivedPowerText(power int, exact bool) string {
	if !exact {
		return "~" + itoa(power)
	}
	return itoa(power)
}
