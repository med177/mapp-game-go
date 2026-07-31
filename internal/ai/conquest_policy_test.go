package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func conquestPolicyTestState() *state.GameState {
	return &state.GameState{
		ScenarioID:      "1300_ottoman_rise",
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"ottoman":       {ID: "ottoman", AIAggressiveness: 62},
			"beylik":        {ID: "beylik", AIAggressiveness: 50},
			"external_ally": {ID: "external_ally"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ottoman_home": {ID: "ottoman_home", OwnerID: "ottoman"},
			"beylik_last":  {ID: "beylik_last", OwnerID: "beylik"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ottoman_army": {ID: "ottoman_army", OwnerID: "ottoman", RegionID: "ottoman_home", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
		},
		UnitTypes: map[string]*army.UnitType{
			"inf": {ID: "inf", Attack: 10, Defense: 10, Morale: 50},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ottoman", "beylik"): {FactionA: "ottoman", FactionB: "beylik", Stance: faction.StanceWar, Score: -80},
		},
		AIPlans: map[faction.FactionID]*state.AIPlanState{
			"ottoman": {
				ObjectiveID:        "unite_anatolian_beyliks",
				Kind:               state.AIObjectiveExpand,
				TargetFactionID:    "beylik",
				TargetRegionIDs:    []world.RegionID{"beylik_last"},
				AllowVassalization: true,
			},
		},
	}
}

func TestTryResolvePostWarVassalizationLeavesWeakIsolatedBeylikIntact(t *testing.T) {
	gs := conquestPolicyTestState()
	result := TryResolvePostWarVassalization(gs, "ottoman", gs.Regions["beylik_last"])

	if !result.Applied {
		t.Fatalf("zayıf ve yalıtılmış beylik vassal bırakılmalıydı: %+v", result)
	}
	if gs.Regions["beylik_last"].OwnerID != "beylik" || gs.Factions["beylik"].OverlordID != "ottoman" {
		t.Fatalf("vassallık yerel sahipliği korumalıydı: owner=%s overlord=%s", gs.Regions["beylik_last"].OwnerID, gs.Factions["beylik"].OverlordID)
	}
}

func TestTryResolvePostWarVassalizationKeepsStrategicRegionForAnnexation(t *testing.T) {
	gs := conquestPolicyTestState()
	gs.AIPlans["ottoman"].AnnexRegionIDs = []world.RegionID{"beylik_last"}

	if result := TryResolvePostWarVassalization(gs, "ottoman", gs.Regions["beylik_last"]); result.Applied {
		t.Fatalf("stratejik bölge doğrudan ilhak için bırakılmalıydı: %+v", result)
	}
	if gs.Factions["beylik"].OverlordID != "" {
		t.Fatal("stratejik hedef yanlışlıkla vassal yapıldı")
	}
}

func TestTryResolvePostWarVassalizationRejectsActiveSuccessorMetadata(t *testing.T) {
	gs := conquestPolicyTestState()
	gs.Factions["successor"] = &faction.Faction{ID: "successor", NameTR: "Ardıl"}
	gs.Regions["beylik_last"].SuccessorFactionID = "successor"

	if result := TryResolvePostWarVassalization(gs, "ottoman", gs.Regions["beylik_last"]); result.Applied {
		t.Fatalf("aktif ardıl devlet metadata'sında AI vassallık uygulamamalıydı: %+v", result)
	}
	if got := gs.Factions["beylik"].OverlordID; got != "" {
		t.Fatalf("aktif ardıl metadata'sında savunmacı vassal yapılmamalıydı: %q", got)
	}
}

func TestTryResolvePostWarVassalizationRejectsResistantOrAlliedTarget(t *testing.T) {
	for _, test := range []struct {
		name  string
		setup func(*state.GameState)
	}{
		{name: "dirençli", setup: func(gs *state.GameState) { gs.Factions["beylik"].AIAggressiveness = aiVassalResistanceAggressiveness }},
		{name: "dış müttefikli", setup: func(gs *state.GameState) {
			gs.Relations[faction.RelationKey("beylik", "external_ally")] = &faction.Relation{FactionA: "beylik", FactionB: "external_ally", Stance: faction.StanceAllied, Score: 70}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			gs := conquestPolicyTestState()
			test.setup(gs)
			if result := TryResolvePostWarVassalization(gs, "ottoman", gs.Regions["beylik_last"]); result.Applied {
				t.Fatalf("dirençli veya diplomatik destekli hedef vassal olmamalıydı: %+v", result)
			}
		})
	}
}
