package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestInciteRevoltRaisesVassalRelationAndLowersOverlordRelationEveryTime(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":   {ID: "player", Gold: 1300, NameTR: "Oyuncu"},
			"vassal":   {ID: "vassal", OverlordID: "overlord", NameTR: "Vassal"},
			"overlord": {ID: "overlord", NameTR: "Sahip"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "vassal"):   {FactionA: "player", FactionB: "vassal", Score: 20, Stance: faction.StancePeace},
			faction.RelationKey("vassal", "overlord"): {FactionA: "vassal", FactionB: "overlord", Score: 50, Stance: faction.StanceAllied},
		},
		DiplomacyConfig: scenario.DiplomacyConfig{
			InciteRevoltGoldCost:        600,
			InciteRevoltRelationBonus:   10,
			InciteRevoltOverlordPenalty: 10,
			InciteRevoltVassalThreshold: 100,
			InciteRevoltOwnerThreshold:  -100,
		},
	}

	for i, wantGold := range []int{700, 100} {
		result := Execute(gs, "player", "vassal", ActionInciteRevolt)
		if !result.Applied {
			t.Fatalf("%d. teşvik uygulanmadı: %s", i+1, result.Message)
		}
		if got := gs.Factions["player"].Gold; got != wantGold {
			t.Fatalf("%d. teşvikte altın: got=%d want=%d", i+1, got, wantGold)
		}
		if got := gs.Factions["vassal"].Gold; got != (i+1)*600 {
			t.Fatalf("%d. teşvikte vassal hazinesi: got=%d want=%d", i+1, got, (i+1)*600)
		}
		if got := Relation(gs, "player", "vassal").Score; got != 20+(i+1)*10 {
			t.Fatalf("%d. teşvikte vassal ilişkisi: got=%d", i+1, got)
		}
		if got := Relation(gs, "vassal", "overlord").Score; got != 50-(i+1)*10 {
			t.Fatalf("%d. teşvikte sahip ilişkisi: got=%d", i+1, got)
		}
	}
}

func TestInciteRevoltBreaksVassalageAtConfiguredThresholds(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":   {ID: "player", Gold: 10},
			"vassal":   {ID: "vassal", OverlordID: "overlord", Gold: 0},
			"overlord": {ID: "overlord"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "vassal"):   {FactionA: "player", FactionB: "vassal", Stance: faction.StancePeace},
			faction.RelationKey("vassal", "overlord"): {FactionA: "vassal", FactionB: "overlord", Score: 40, Stance: faction.StanceAllied},
		},
		Regions: map[world.RegionID]*world.Region{
			"vassal_region_a": {ID: "vassal_region_a", OwnerID: "vassal", Satisfaction: 80},
			"vassal_region_b": {ID: "vassal_region_b", OwnerID: "vassal", Satisfaction: 35},
			"foreign_region":  {ID: "foreign_region", OwnerID: "overlord", Satisfaction: 70},
		},
		DiplomacyConfig: scenario.DiplomacyConfig{
			InciteRevoltGoldCost:        1,
			InciteRevoltRelationBonus:   10,
			InciteRevoltOverlordPenalty: 10,
			InciteRevoltVassalThreshold: 50,
			InciteRevoltOwnerThreshold:  -20,
		},
	}

	for i := 0; i < 5; i++ {
		gs.ResetDiplomacyOfferCounts()
		if result := Execute(gs, "player", "vassal", ActionInciteRevolt); !result.Applied {
			t.Fatalf("eşik öncesi %d. teşvik uygulanmadı: %s", i+1, result.Message)
		}
	}
	if gs.Factions["vassal"].OverlordID != "overlord" {
		t.Fatal("eşik oluşmadan vassallık kaldırıldı")
	}
	gs.ResetDiplomacyOfferCounts()

	result := Execute(gs, "player", "vassal", ActionInciteRevolt)
	if !result.Applied || gs.Factions["vassal"].OverlordID != "" {
		t.Fatalf("eşikte vassal isyanı başlamadı: applied=%v message=%s", result.Applied, result.Message)
	}
	if relation := Relation(gs, "vassal", "overlord"); relation == nil || relation.Stance != faction.StanceWar {
		t.Fatal("isyan sonrası vassal ile sahibi savaşa girmedi")
	}
	if got := gs.Regions["vassal_region_a"].Satisfaction; got != 20 {
		t.Fatalf("vassal bölgesinin memnuniyeti azaltılmadı: got=%d want=20", got)
	}
	if got := gs.Regions["vassal_region_b"].Satisfaction; got != 0 {
		t.Fatalf("vassal bölgesinin memnuniyeti sıfırın altına indi veya yanlış azaltıldı: got=%d want=0", got)
	}
	if got := gs.Regions["foreign_region"].Satisfaction; got != 70 {
		t.Fatalf("sahibin bölgesinin memnuniyeti değişti: got=%d want=70", got)
	}
}

func TestInciteRevoltRequiresExternalActorAndVassalTarget(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":  {ID: "player", Gold: 1000},
			"vassal":  {ID: "vassal", OverlordID: "player"},
			"foreign": {ID: "foreign", Gold: 0},
		},
	}

	if reason := ActionBlockReason(gs, "player", "foreign", ActionInciteRevolt); reason == "" {
		t.Fatal("vassal olmayan hedef için teşvik engellenmedi")
	}
	if reason := ActionBlockReason(gs, "player", "vassal", ActionInciteRevolt); reason == "" {
		t.Fatal("kendi vassalı için teşvik engellenmedi")
	}
}

func TestInciteRevoltRemainsAvailableWhenVassalRelationIsMaxed(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "venice",
		Factions: map[faction.FactionID]*faction.Faction{
			"venice":   {ID: "venice", Gold: 1000},
			"flanders": {ID: "flanders", OverlordID: "hre"},
			"hre":      {ID: "hre"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("venice", "flanders"): {
				FactionA: "flanders", FactionB: "venice", Score: 100, Stance: faction.StanceTrade,
			},
		},
		DiplomacyConfig: scenario.DiplomacyConfig{
			InciteRevoltGoldCost:        600,
			InciteRevoltRelationBonus:   10,
			InciteRevoltOverlordPenalty: 10,
			InciteRevoltVassalThreshold: 150,
			InciteRevoltOwnerThreshold:  -100,
		},
	}

	if reason := ActionBlockReason(gs, "venice", "flanders", ActionInciteRevolt); reason != "" {
		t.Fatalf("maksimum ilişkide dış vassala teşvik engellendi: %s", reason)
	}
	result := Execute(gs, "venice", "flanders", ActionInciteRevolt)
	if !result.Applied {
		t.Fatalf("maksimum ilişkide dış vassala teşvik uygulanmadı: %s", result.Message)
	}
}
