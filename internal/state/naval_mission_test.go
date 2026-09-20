package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestConvertInvalidNavalBlockadesToPatrol(t *testing.T) {
	gs := navalMissionTransitionState()
	blockade := gs.Armies["blockade"]

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 0 {
		t.Fatalf("geçerli abluka için dönüşüm sayısı = %d, want 0", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionBlockade {
		t.Fatalf("geçerli abluka görevi değişmemeli, got %q", blockade.NavalMission.Kind)
	}

	gs.Regions["enemy_port"].OwnerID = "player"
	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("geçersiz abluka için dönüşüm sayısı = %d, want 1", got)
	}
	if blockade.NavalMission.Kind != army.NavalMissionPatrol {
		t.Fatalf("geçersiz abluka devriyeye dönüşmeli, got %q", blockade.NavalMission.Kind)
	}
	if blockade.NavalMission.TargetRegionID != "sea" {
		t.Fatalf("devriye hedefi mevcut deniz olmalı, got %q", blockade.NavalMission.TargetRegionID)
	}
}

func TestConvertInvalidNavalBlockadesKeepsUnrelatedMissions(t *testing.T) {
	gs := navalMissionTransitionState()
	gs.Armies["patrol"] = &army.Army{
		ID:       "patrol",
		OwnerID:  "player",
		RegionID: "sea",
		IsNaval:  true,
		NavalMission: &army.NavalMission{
			Kind:           army.NavalMissionPatrol,
			TargetRegionID: "sea",
		},
	}
	gs.Regions["enemy_port"].OwnerID = "player"

	if got := gs.ConvertInvalidNavalBlockadesToPatrol(); got != 1 {
		t.Fatalf("dönüşüm sayısı = %d, want 1", got)
	}
	if got := gs.Armies["patrol"].NavalMission.Kind; got != army.NavalMissionPatrol {
		t.Fatalf("mevcut devriye görevi korunmalı, got %q", got)
	}
}

func navalMissionTransitionState() *GameState {
	return &GameState{
		Regions: map[world.RegionID]*world.Region{
			"enemy_port": {
				ID:          "enemy_port",
				OwnerID:     "enemy",
				Neighbors:   []world.RegionID{"sea"},
				Settlements: []world.Settlement{{ID: "enemy_harbor", Type: world.SettlementPort}},
			},
			"sea": {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"enemy_port"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"blockade": {
				ID:       "blockade",
				OwnerID:  "player",
				RegionID: "sea",
				IsNaval:  true,
				NavalMission: &army.NavalMission{
					Kind:           army.NavalMissionBlockade,
					TargetRegionID: "sea",
				},
			},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player",
				FactionB: "enemy",
				Stance:   faction.StanceWar,
			},
		},
	}
}
