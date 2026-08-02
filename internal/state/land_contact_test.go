package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestLandContactWithdrawRequiresSafeDefenderRoute(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"source": {ID: "source", OwnerID: "enemy", Neighbors: []world.RegionID{"front"}},
			"front":  {ID: "front", OwnerID: "player", Neighbors: []world.RegionID{"source", "rear"}},
			"rear":   {ID: "rear", OwnerID: "player", Neighbors: []world.RegionID{"front"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"attacker": {ID: "attacker", OwnerID: "enemy", RegionID: "source", MovePoints: 1},
			"player":   {ID: "player", OwnerID: "player", RegionID: "front", MovePoints: 1},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("enemy", "player"): {FactionA: "enemy", FactionB: "player", Stance: faction.StanceWar},
		},
	}
	contact := gs.BeginLandContact(gs.Armies["attacker"], gs.Armies["player"], "front", "source", LandContactMovement)
	if contact == nil || !gs.LandContactHasSafeWithdrawal(contact) {
		t.Fatalf("oyuncu savunucunun güvenli geri çekilme seçeneği olmalı: contact=%+v", contact)
	}
	if !gs.LandContactDecisionForPlayer(contact, LandContactWithdraw) || contact.DefenderDecision != LandContactWithdraw {
		t.Fatalf("geri çekilme kararı savunucuya yazılmalı: %+v", contact)
	}
}
