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
		Year:     1454,
		Month:    1,
		Factions: map[faction.FactionID]*faction.Faction{},
		Regions:  map[world.RegionID]*world.Region{},
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
			ID:              "allowed_event",
			NameTR:          "Allowed",
			HistoricalYear:  1454,
			HistoricalMonth: 1,
			RequiresFlags:   []string{"capital_rebuild"},
			BlocksFlags:     []string{"arsenal_push"},
		},
	}

	evt := Tick(gs, evts)
	if evt == nil || evt.ID != "allowed_event" {
		t.Fatalf("flag koşullarına göre allowed_event bekleniyordu, got=%v", evt)
	}
}
