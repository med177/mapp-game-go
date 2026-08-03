package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func raidAmbushTestState() *GameState {
	return &GameState{
		Turn:            4,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"enemy": {ID: "enemy", OwnerID: "enemy", Terrain: world.TerrainPass},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"}, "enemy": {ID: "enemy"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player_army": {ID: "player_army", OwnerID: "player", RegionID: "enemy", MovePoints: 2},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}
}

func TestRaidCanBeAppliedOncePerRegionAndTurn(t *testing.T) {
	gs := raidAmbushTestState()
	region := gs.Regions["enemy"]
	a := gs.Armies["player_army"]
	if !gs.ApplyRaid(a, region) {
		t.Fatal("ilk yağma uygulanmalıydı")
	}
	if gs.CanRaid(a, region) {
		t.Fatal("aynı bölge aynı turda ikinci kez yağmalanmamalı")
	}
	if a.MovePoints != 0 {
		t.Fatalf("yağma hareket puanını tüketmeli, got=%d", a.MovePoints)
	}
	if got := gs.Raids[region.ID].RaiderArmyID; got != a.ID {
		t.Fatalf("yağma kaydı doğru orduyu taşımalı: got=%q want=%q", got, a.ID)
	}
}

func TestAmbushIsHiddenFromBattleSelectionButTriggersSpecialSelection(t *testing.T) {
	gs := raidAmbushTestState()
	gs.Armies["enemy_army"] = &army.Army{ID: "enemy_army", OwnerID: "enemy", RegionID: "enemy", InAmbush: true}
	attacker := gs.Armies["player_army"]
	if got := gs.SelectBattleDefender(attacker, "enemy", false); got != nil {
		t.Fatalf("gizli pusu ordusu normal seçimde görünmemeli: %+v", got)
	}
	if got := gs.SelectAmbushDefender(attacker, "enemy", false); got == nil || got.ID != "enemy_army" {
		t.Fatalf("gizli pusu ordusu özel temas seçiminde bulunmalı: %+v", got)
	}
	contact := gs.BeginLandContact(attacker, gs.Armies["enemy_army"], "enemy", "source", LandContactMovement)
	if contact == nil || contact.AmbushArmyID != "enemy_army" {
		t.Fatalf("temas pusu ordusunu işaretlemeli: %+v", contact)
	}
	if gs.LandContactDecisionForPlayer(contact, LandContactHold) {
		t.Fatal("pusu temasında pozisyonu koru seçilememeli")
	}
}
