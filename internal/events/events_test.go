package events

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestApplyUnitReinforcementsCreatesLandArmyAndDockedFleet(t *testing.T) {
	const ownerID = faction.FactionID("east_rome")
	capital := &world.Region{
		ID:      "constantinople",
		OwnerID: string(ownerID),
		Neighbors: []world.RegionID{
			"sea_of_marmara",
		},
		Settlements: []world.Settlement{{ID: "constantinople_main"}},
	}
	sea := &world.Region{ID: "sea_of_marmara", IsSea: true}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			capital.ID: capital,
			sea.ID:     sea,
		},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "constantinople_main"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
			"warship":  {ID: "warship", Category: army.CategoryNavalWar},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		UnitReinforcements: []UnitReinforcementEffect{
			{UnitType: "infantry", UnitCount: 5},
			{UnitType: "warship", UnitCount: 3},
		},
	})

	if len(gs.Armies) != 2 {
		t.Fatalf("%d ordu oluşturuldu, 2 bekleniyordu", len(gs.Armies))
	}
	var foundLand, foundFleet bool
	for _, current := range gs.Armies {
		switch {
		case !current.IsNaval:
			foundLand = len(current.Units) == 5 && current.Units[0].TypeID == "infantry" && current.RegionID == capital.ID
		case current.IsNaval:
			foundFleet = len(current.Units) == 3 && current.Units[0].TypeID == "warship" &&
				current.RegionID == sea.ID && current.DockedRegionID == capital.ID &&
				current.DockedSettlementID == "constantinople_main"
		}
	}
	if !foundLand || !foundFleet {
		t.Fatalf("takviyeler beklenen kara ordusu ve dock edilmiş filo olarak oluşturulmadı")
	}
}

func TestApplyChoiceAppliesChoiceReinforcement(t *testing.T) {
	const ownerID = faction.FactionID("ottoman")
	capital := &world.Region{
		ID:          "bithynia",
		OwnerID:     string(ownerID),
		Settlements: []world.Settlement{{ID: "sogut"}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{capital.ID: capital},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "sogut"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	_, ok := ApplyChoice(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		Choices:         []Choice{{Effect: Effect{UnitReinforcements: []UnitReinforcementEffect{{UnitType: "infantry", UnitCount: 3}}}}},
	}, 0)
	if !ok || len(gs.Armies) != 1 {
		t.Fatalf("seçim takviyesi uygulanmadı")
	}
	for _, current := range gs.Armies {
		if len(current.Units) != 3 || current.Units[0].TypeID != "infantry" {
			t.Fatalf("seçim sonrası yanlış birlik oluşturuldu")
		}
	}
}

func TestApplyBaseEventReinforcement(t *testing.T) {
	const ownerID = faction.FactionID("ottoman")
	capital := &world.Region{
		ID:          "bilecik_frontier",
		OwnerID:     string(ownerID),
		Settlements: []world.Settlement{{ID: "sogut"}},
	}
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{capital.ID: capital},
		Factions: map[faction.FactionID]*faction.Faction{
			ownerID: {ID: ownerID, CapitalSettlementID: "sogut"},
		},
		Armies: map[army.ArmyID]*army.Army{},
		UnitTypes: map[string]*army.UnitType{
			"infantry": {ID: "infantry", Category: army.CategoryInfantry},
		},
	}

	Apply(gs, &Event{
		Target:          "specific_faction",
		AffectedFaction: string(ownerID),
		UnitReinforcements: []UnitReinforcementEffect{
			{UnitType: "infantry", UnitCount: 3},
		},
	})

	if len(gs.Armies) != 1 {
		t.Fatalf("temel event takviyesi %d ordu oluşturdu, 1 bekleniyordu", len(gs.Armies))
	}
	for _, current := range gs.Armies {
		if current.RegionID != capital.ID || len(current.Units) != 3 || current.Units[0].TypeID != "infantry" {
			t.Fatalf("temel event takviyesi başkentte 3 piyade oluşturmadı")
		}
	}
}
