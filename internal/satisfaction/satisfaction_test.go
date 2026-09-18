package satisfaction

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestArmyStabilityBonusUsesVassalAndAllyArmies(t *testing.T) {
	const regionID = world.RegionID("region")

	tests := []struct {
		name       string
		armyOwner  faction.FactionID
		relation   faction.DiplomaticStance
		overlordID faction.FactionID
		want       int
	}{
		{name: "owner", armyOwner: "player", want: 10},
		{name: "vassal", armyOwner: "vassal", overlordID: "player", want: 10},
		{name: "ally", armyOwner: "ally", relation: faction.StanceAllied, want: 7},
		{name: "enemy", armyOwner: "enemy", relation: faction.StanceWar, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &state.GameState{
				Regions: map[world.RegionID]*world.Region{
					regionID: {ID: regionID, OwnerID: "player"},
				},
				Factions: map[faction.FactionID]*faction.Faction{
					"player": {ID: "player"},
					"vassal": {ID: "vassal", OverlordID: tt.overlordID},
					"ally":   {ID: "ally"},
					"enemy":  {ID: "enemy"},
				},
				Relations: map[string]*faction.Relation{},
				Armies: map[army.ArmyID]*army.Army{
					"army": {
						ID:       "army",
						OwnerID:  string(tt.armyOwner),
						RegionID: regionID,
						Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
					},
				},
				UnitTypes: map[string]*army.UnitType{
					"infantry": {ID: "infantry", Attack: 100},
				},
			}
			if tt.relation != "" {
				gs.Relations[faction.RelationKey("player", tt.armyOwner)] = &faction.Relation{
					FactionA: "player",
					FactionB: tt.armyOwner,
					Stance:   tt.relation,
				}
			}

			if got := ArmyStabilityBonus(gs, gs.Regions[regionID]); got != tt.want {
				t.Fatalf("ArmyStabilityBonus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestArmyStabilityBonusCombinesWeightedStrengthBeforeCap(t *testing.T) {
	const regionID = world.RegionID("region")
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			regionID: {ID: regionID, OwnerID: "player"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {
				FactionA: "player",
				FactionB: "ally",
				Stance:   faction.StanceAllied,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"owner": {
				ID:       "owner",
				OwnerID:  "player",
				RegionID: regionID,
				Units:    []army.Unit{{TypeID: "owner_infantry", CurrentHP: 100}},
			},
			"ally": {
				ID:       "ally",
				OwnerID:  "ally",
				RegionID: regionID,
				Units:    []army.Unit{{TypeID: "ally_infantry", CurrentHP: 100}},
			},
		},
		UnitTypes: map[string]*army.UnitType{
			"owner_infantry": {ID: "owner_infantry", Attack: 55},
			"ally_infantry":  {ID: "ally_infantry", Attack: 60},
		},
	}

	// 55 tam güç + 60 müttefik gücünün %75'i = 100; ayrı ayrı tam sayıya
	// çevrilseydi 5 + 4 olurdu. Önce toplam ağırlıklandırılır, sonra bonus
	// hesaplanır.
	if got := ArmyStabilityBonus(gs, gs.Regions[regionID]); got != 10 {
		t.Fatalf("ArmyStabilityBonus() = %d, want capped bonus 10", got)
	}
}
