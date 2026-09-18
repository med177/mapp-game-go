package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestAdvanceImperialPoliticsSkipsLockedElection(t *testing.T) {
	const (
		hre     = faction.FactionID("hre")
		austria = faction.FactionID("austria_duchy")
		bohemia = faction.FactionID("bohemian_kingdom")
	)
	gs := &state.GameState{
		Turn:            120,
		PlayerFactionID: hre,
		Factions: map[faction.FactionID]*faction.Faction{
			hre:     {ID: hre},
			austria: {ID: austria},
			bohemia: {ID: bohemia},
		},
		Imperial: &state.ImperialState{
			EmpireID:        hre,
			EmperorID:       austria,
			ElectionDueTurn: 97,
			ElectionLocked:  true,
			Members:         map[faction.FactionID]*state.ImperialMember{},
		},
	}

	report := AdvanceImperialPolitics(gs)
	if report.Pending || gs.Imperial.PendingDecision != nil {
		t.Fatal("kilitli HRE seçiminde bekleyen seçim kararı oluşturulmamalı")
	}
	if gs.Imperial.ElectionDueTurn != 97 {
		t.Fatalf("kilitli seçim takvimi değişti: %d", gs.Imperial.ElectionDueTurn)
	}
	if result := HoldImperialElection(gs); result.WinnerID != "" {
		t.Fatalf("kilitli HRE seçimi kazanan üretti: %s", result.WinnerID)
	}
}
