package ai

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestAITerritoryTaskCanSetAmbushAndKeepArmyHidden(t *testing.T) {
	gs := &state.GameState{
		Turn: 3,
		Regions: map[world.RegionID]*world.Region{
			"pass": {ID: "pass", OwnerID: "enemy", Terrain: world.TerrainPass},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ai_army": {ID: "ai_army", OwnerID: "ai", RegionID: "pass", MovePoints: 2},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("ai", "enemy"): {FactionA: "ai", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}
	step, handled := executeAITerritoryTask(gs, gs.Armies["ai_army"], "ai")
	if !handled || !gs.Armies["ai_army"].InAmbush {
		t.Fatalf("AI pusu görevi uygulanmalıydı: handled=%v army=%+v", handled, gs.Armies["ai_army"])
	}
	if step.Message == "" || gs.Armies["ai_army"].MovePoints != 0 {
		t.Fatalf("AI pusu görevi hareketi tüketip kayıt üretmeli: step=%+v army=%+v", step, gs.Armies["ai_army"])
	}
}
