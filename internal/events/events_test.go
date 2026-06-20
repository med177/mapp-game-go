package events

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

func TestTickReturnsEventWithoutApplyingEffects(t *testing.T) {
	gs := &state.GameState{
		Turn:            7,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 100, Grain: 100},
		},
		Regions: map[world.RegionID]*world.Region{
			"r1": {ID: "r1", OwnerID: "player", Satisfaction: 60},
		},
	}
	evts := []*Event{{
		ID:          "trade_boom",
		NameTR:      "Ticaret Patlaması",
		Target:      "player_faction",
		MinTurn:     6,
		Probability: 1,
		GoldDelta:   50,
		SatDelta:    10,
	}}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "trade_boom" {
		t.Fatalf("tetiklenen event bekleniyordu, got=%v", evt)
	}
	if gs.Factions["player"].Gold != 100 {
		t.Fatalf("Tick artık effect uygulamamalıydı, gold=%d", gs.Factions["player"].Gold)
	}
	if gs.Regions["r1"].Satisfaction != 60 {
		t.Fatalf("Tick artık satisfaction değiştirmemeliydi, got=%d", gs.Regions["r1"].Satisfaction)
	}
}

func TestApplyChoiceUpdatesFactionArmyAndRelations(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", Gold: 200, Grain: 120, Religion: religion.Catholic},
			"ally":   {ID: "ally", Gold: 50, Grain: 40, Religion: religion.Catholic},
			"rival":  {ID: "rival", Gold: 80, Grain: 50, Religion: religion.Sunni},
		},
		Regions: map[world.RegionID]*world.Region{
			"cap": {ID: "cap", OwnerID: "player", Satisfaction: 55},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "player", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "ally", FactionB: "player", Score: 10, Stance: faction.StancePeace},
			faction.RelationKey("player", "rival"): {FactionA: "player", FactionB: "rival", Score: -20, Stance: faction.StanceWar},
		},
		TechTypes: map[string]*tech.Technology{
			"trade_routes": {ID: "trade_routes"},
			"banking":      {ID: "banking", TurnsRequired: 4},
		},
	}
	evt := &Event{
		ID:     "policy",
		Target: "player_faction",
		Choices: []Choice{{
			LabelTR: "Karantina",
			Effect: Effect{
				Target:            "player_faction",
				SatDelta:          8,
				GoldDelta:         -30,
				GrainDelta:        -10,
				ArmyHPMod:         0.9,
				RelationDeltaAll:  -5,
				CompleteTechs:     []string{"trade_routes"},
				StartResearchTech: "banking",
				Relations: []RelationEffect{{
					FactionID:  "ally",
					Stance:     "trade",
					ScoreDelta: 7,
				}},
				SetFlags: []string{"plague_quarantine"},
			},
		}},
	}

	choice, ok := ApplyChoice(gs, evt, 0)
	if !ok || choice.LabelTR != "Karantina" {
		t.Fatalf("choice apply başarısız, choice=%+v ok=%v", choice, ok)
	}
	if gs.Factions["player"].Gold != 170 || gs.Factions["player"].Grain != 110 {
		t.Fatalf("ekonomi etkisi uygulanmadı, gold=%d grain=%d", gs.Factions["player"].Gold, gs.Factions["player"].Grain)
	}
	if gs.Regions["cap"].Satisfaction != 63 {
		t.Fatalf("memnuniyet 63 olmalıydı, got=%d", gs.Regions["cap"].Satisfaction)
	}
	if hp := gs.Armies["a1"].Units[0].CurrentHP; hp != 90 {
		t.Fatalf("ordu hp 90 olmalıydı, got=%d", hp)
	}
	if score := gs.Relations[faction.RelationKey("player", "ally")].Score; score != 12 {
		t.Fatalf("ally relation 12 olmalıydı, got=%d", score)
	}
	if score := gs.Relations[faction.RelationKey("player", "rival")].Score; score != -25 {
		t.Fatalf("rival relation -25 olmalıydı, got=%d", score)
	}
	if stance := gs.Relations[faction.RelationKey("player", "ally")].Stance; stance != faction.StanceTrade {
		t.Fatalf("ally stance trade olmalıydı, got=%s", stance)
	}
	if !gs.FiredEventIDs["flag:plague_quarantine"] {
		t.Fatalf("choice flag set edilmeliydi")
	}
	if !gs.Factions["player"].Research.Completed["trade_routes"] {
		t.Fatalf("trade_routes tech tamamlanmalıydı")
	}
	if gs.Factions["player"].Research.ActiveID != "banking" || gs.Factions["player"].Research.TurnsLeft != 4 {
		t.Fatalf("banking aktif araştırma başlamalıydı, got=%s/%d", gs.Factions["player"].Research.ActiveID, gs.Factions["player"].Research.TurnsLeft)
	}
}

func TestTickRequiresAndBlocksFlags(t *testing.T) {
	gs := &state.GameState{
		Year:  1454,
		Month: 1,
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {ID: "ottoman", Research: faction.ResearchState{Completed: map[string]bool{}}},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "ottoman"},
		},
		FiredEventIDs: map[string]bool{
			"flag:capital_rebuild": true,
		},
	}
	evts := []*Event{
		{
			ID:              "blocked_event",
			NameTR:          "Blocked",
			HistoricalYear:  1454,
			HistoricalMonth: 1,
			RequiresFlags:   []string{"arsenal_push"},
		},
		{
			ID:                   "allowed_event",
			NameTR:               "Allowed",
			HistoricalYear:       1454,
			HistoricalMonth:      1,
			RequiresFlags:        []string{"capital_rebuild"},
			BlocksFlags:          []string{"arsenal_push"},
			Target:               "specific_faction",
			AffectedFaction:      "ottoman",
			RequiresOwnedRegions: []world.RegionID{"constantinople"},
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed_event" {
		t.Fatalf("flag koşullarına göre allowed_event bekleniyordu, got=%v", evt)
	}
}

func TestTickRequiresTechAndRelationConditions(t *testing.T) {
	gs := &state.GameState{
		Year:  1494,
		Month: 4,
		Factions: map[faction.FactionID]*faction.Faction{
			"aragon":          {ID: "aragon", Research: faction.ResearchState{Completed: map[string]bool{"navigation": true}}},
			"castile_kingdom": {ID: "castile_kingdom"},
			"portugal":        {ID: "portugal"},
		},
		Regions: map[world.RegionID]*world.Region{
			"granada": {ID: "granada", OwnerID: "aragon"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("aragon", "castile_kingdom"): {
				FactionA: "aragon",
				FactionB: "castile_kingdom",
				Score:    22,
				Stance:   faction.StanceAllied,
			},
			faction.RelationKey("aragon", "portugal"): {
				FactionA: "aragon",
				FactionB: "portugal",
				Score:    -60,
				Stance:   faction.StanceWar,
			},
		},
		FiredEventIDs: map[string]bool{
			"flag:atlantic_expedition": true,
			"flag:iberian_market":      true,
		},
	}
	evts := []*Event{
		{
			ID:                   "blocked_by_tech",
			NameTR:               "BlockedByTech",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"atlantic_expedition"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			BlocksTechs:          []string{"navigation"},
		},
		{
			ID:                   "blocked_by_relation",
			NameTR:               "BlockedByRelation",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"iberian_market"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			RelationRequirements: []RelationRequirement{{
				FactionID:     "portugal",
				BlocksStances: []string{string(faction.StanceWar)},
			}},
		},
		{
			ID:                   "allowed_with_relation_and_tech",
			NameTR:               "Allowed",
			HistoricalYear:       1494,
			HistoricalMonth:      4,
			Target:               "specific_faction",
			AffectedFaction:      "aragon",
			RequiresFlags:        []string{"atlantic_expedition"},
			RequiresOwnedRegions: []world.RegionID{"granada"},
			RequiresTechs:        []string{"navigation"},
			RelationRequirements: []RelationRequirement{{
				FactionID:    "castile_kingdom",
				AnyOfStances: []string{string(faction.StancePeace), string(faction.StanceTrade), string(faction.StanceAllied)},
				MinScore:     10,
			}},
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed_with_relation_and_tech" {
		t.Fatalf("tech ve relation koşullarına göre allowed event bekleniyordu, got=%v", evt)
	}
}

func TestTickBlocksCapitalSettlementWhenAlreadyCurrentOrPending(t *testing.T) {
	gs := &state.GameState{
		Year:  1453,
		Month: 5,
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman": {
				ID:                         "ottoman",
				CapitalSettlementID:        "edirne",
				PendingCapitalSettlementID: "ghost",
				PendingCapitalTurns:        2,
				Research:                   faction.ResearchState{Completed: map[string]bool{}},
			},
		},
		Regions: map[world.RegionID]*world.Region{
			"constantinople": {ID: "constantinople", OwnerID: "ottoman"},
		},
	}
	evts := []*Event{
		{
			ID:                        "blocked_current",
			NameTR:                    "BlockedCurrent",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "edirne",
		},
		{
			ID:                        "blocked_pending",
			NameTR:                    "BlockedPending",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "ghost",
		},
		{
			ID:                        "allowed",
			NameTR:                    "Allowed",
			HistoricalYear:            1453,
			HistoricalMonth:           5,
			Target:                    "specific_faction",
			AffectedFaction:           "ottoman",
			RequiresOwnedRegions:      []world.RegionID{"constantinople"},
			BlocksCapitalSettlementID: "constantinople_main",
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed" {
		t.Fatalf("capital block koşullarına göre allowed event bekleniyordu, got=%v", evt)
	}
}

func TestApplyChoiceStartsCapitalMove(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", CapitalSettlementID: "old_cap"},
		},
		Regions: map[world.RegionID]*world.Region{
			"home": {
				ID:      "home",
				OwnerID: "player",
				Settlements: []world.Settlement{
					{ID: "old_cap", NameTR: "Eski Başkent", IsCapital: true},
					{ID: "new_cap", NameTR: "Yeni Başkent"},
				},
			},
		},
	}
	evt := &Event{
		ID:     "move_capital",
		Target: "player_faction",
		Choices: []Choice{{
			LabelTR: "Taşı",
			Effect: Effect{
				Target:              "player_faction",
				CapitalSettlementID: "new_cap",
				CapitalMoveTurns:    3,
			},
		}},
	}

	if _, ok := ApplyChoice(gs, evt, 0); !ok {
		t.Fatal("choice uygulanmalıydı")
	}
	f := gs.Factions["player"]
	if f.PendingCapitalSettlementID != "new_cap" || f.PendingCapitalTurns != 3 {
		t.Fatalf("pending başkent taşıma başlamalıydı, got=%+v", f)
	}
}
