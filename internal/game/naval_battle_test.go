package game

import (
	"testing"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func newNavalBattleGame(attacker, defender *army.Army) *Game {
	return &Game{
		gs: &state.GameState{
			PlayerFactionID: "p1",
			Regions: map[world.RegionID]*world.Region{
				"sea_a":       {ID: "sea_a", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
				"sea_b":       {ID: "sea_b", IsSea: true, Neighbors: []world.RegionID{"sea_a", "sea_c", "enemy_coast"}},
				"sea_c":       {ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"sea_b"}},
				"enemy_coast": {ID: "enemy_coast", OwnerID: "p2", Neighbors: []world.RegionID{"sea_b"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				attacker.ID: attacker,
				defender.ID: defender,
			},
			Factions: map[faction.FactionID]*faction.Faction{
				"p1": {ID: "p1"},
				"p2": {ID: "p2"},
			},
			Relations: map[string]*faction.Relation{
				faction.RelationKey("p1", "p2"): {
					FactionA: "p1",
					FactionB: "p2",
					Stance:   faction.StanceWar,
				},
			},
			UnitTypes: map[string]*army.UnitType{
				"warship":   {ID: "warship", Category: army.CategoryNavalWar, Attack: 100, Defense: 100, Morale: 100},
				"transport": {ID: "transport", Category: army.CategoryNavalTrans, Attack: 1, Defense: 1, Morale: 1},
			},
		},
		renderer: &render.Renderer{},
	}
}

func TestNavalBattleLossSinksFleetAndEmbarkedArmy(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{
			ID:            "attacker_fleet",
			OwnerID:       "p1",
			RegionID:      "sea_a",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
		},
		&army.Army{
			ID:            "defender_fleet",
			OwnerID:       "p2",
			RegionID:      "sea_b",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
	)

	g.moveArmyToSettlementWithStanceAndNavalAttack("attacker_fleet", "sea_b", "", combat.BattleStanceBalanced, true)

	if _, ok := g.gs.Armies["attacker_fleet"]; ok {
		t.Fatal("deniz savaşını kaybeden filo state'ten kaldırılmalıydı")
	}
	if _, ok := g.gs.Armies["defender_fleet"]; !ok {
		t.Fatal("deniz savaşını kazanan filo korunmalıydı")
	}
}

func TestNavalBattleDefeatSinksDefenderFleetAndEmbarkedArmy(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{
			ID:            "attacker_fleet",
			OwnerID:       "p1",
			RegionID:      "sea_a",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
		&army.Army{
			ID:            "defender_fleet",
			OwnerID:       "p2",
			RegionID:      "sea_b",
			IsNaval:       true,
			MovePoints:    1,
			MaxMovePoints: 1,
			Units:         []army.Unit{{TypeID: "transport", CurrentHP: 100}},
			EmbarkedUnits: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
		},
	)

	g.moveArmyToSettlementWithStanceAndNavalAttack("attacker_fleet", "sea_b", "", combat.BattleStanceBalanced, true)

	if _, ok := g.gs.Armies["defender_fleet"]; ok {
		t.Fatal("deniz savaşını kaybeden savunma filosu state'ten kaldırılmalıydı")
	}
	if _, ok := g.gs.Armies["attacker_fleet"]; !ok {
		t.Fatal("deniz savaşını kazanan saldırı filosu korunmalıydı")
	}
}

func TestUnassignedNavalFleetsCanShareSeaWithoutBattle(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "attacker_fleet", OwnerID: "p1", RegionID: "sea_a", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "defender_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
	)

	g.moveArmy("attacker_fleet", "sea_b")
	if g.gs.PendingNavalContact == nil {
		t.Fatal("görevsiz filolar aynı denize girerken temas kaydı oluşmalıydı")
	}
	if g.gs.Armies["attacker_fleet"].RegionID != "sea_b" || g.gs.Armies["attacker_fleet"].MovePoints != 0 {
		t.Fatal("temas modalı açılmadan önce filo hedef denize taşınmalı ve hareket puanı harcanmalı")
	}
	g.resolveNavalContactChoice(2)
	if g.gs.Armies["attacker_fleet"].RegionID != "sea_b" || g.gs.Armies["defender_fleet"].RegionID != "sea_b" {
		t.Fatal("görevsiz filolar aynı denizde savaşmadan birlikte kalabilmeliydi")
	}
}

func TestPatrolLetsOutmatchedEnemyBlockadeWithdraw(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "patrol_fleet", OwnerID: "p1", RegionID: "sea_a", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea_b"}, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "blockade_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea_b"}, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}},
	)

	g.moveArmy("patrol_fleet", "sea_b")
	contact := g.gs.PendingNavalContact
	if contact == nil || contact.AttackerDecision != state.NavalContactUndecided || contact.DefenderDecision != state.NavalContactWithdraw {
		t.Fatalf("güçsüz abluka filosu geri çekilmeyi seçmeliydi: %+v", contact)
	}
	g.resolveNavalContactChoice(0)
	if len(g.gs.Armies) != 2 {
		t.Fatalf("abluka geri çekildiğinde savaş başlamamalıydı, filo sayısı=%d", len(g.gs.Armies))
	}
	if g.gs.Armies["patrol_fleet"].RegionID != "sea_b" || g.gs.Armies["blockade_fleet"].RegionID != "sea_c" {
		t.Fatal("geri çekilen abluka filosu komşu denize dönmeliydi")
	}
}

func TestNavalContactBattlesOnlyWhenBothClash(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{
			ID: "attacker_fleet", OwnerID: "p1", RegionID: "sea_a", IsNaval: true,
			MovePoints: 1, MaxMovePoints: 1,
			Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
		&army.Army{
			ID: "defender_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true,
			MovePoints: 1, MaxMovePoints: 1,
			NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea_b"},
			Units:        []army.Unit{{TypeID: "warship", CurrentHP: 100}},
		},
	)

	g.moveArmy("attacker_fleet", "sea_b")
	contact := g.gs.PendingNavalContact
	if contact == nil || contact.DefenderDecision != state.NavalContactClash {
		t.Fatalf("devriye filosu temas varsayılanında çatışmayı seçmeli: %+v", contact)
	}

	g.resolveNavalContactChoice(0)
	if g.gs.PendingNavalContact != nil {
		t.Fatal("iki taraf da çatışmayı seçtikten sonra temas kaydı temizlenmeli")
	}
	if len(g.gs.Armies) != 1 {
		t.Fatalf("iki taraf da çatışmayı seçtiğinde savaş çözülmeli, filo sayısı=%d", len(g.gs.Armies))
	}
}

func TestAssigningPatrolOrBlockadeInOccupiedSeaCreatesContactOnce(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "player_fleet", OwnerID: "p1", RegionID: "sea_b", IsNaval: true, MovePoints: 0, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "enemy_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 0, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
	)

	g.assignNavalMission("player_fleet", army.NavalMissionPatrol, "sea_b", "")
	if g.gs.PendingNavalContact == nil || g.gs.PendingNavalContact.Trigger != state.NavalContactMissionAssignment {
		t.Fatal("aynı denizde devriye görevi atanırken temas oluşmalı")
	}
	g.resolveNavalContactChoice(2)

	g.assignNavalMission("player_fleet", army.NavalMissionPatrol, "sea_b", "")
	if g.gs.PendingNavalContact != nil {
		t.Fatal("aynı görev tekrar atanırken temas yeniden açılmamalı")
	}
	g.assignNavalMission("player_fleet", army.NavalMissionBlockade, "sea_b", "")
	if g.gs.PendingNavalContact == nil {
		t.Fatal("aynı denizde yeni abluka görevi atanırken yeniden temas oluşmalı")
	}
}

func TestNavalContactWithdrawMovesFleetToAnotherSea(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "player_fleet", OwnerID: "p1", RegionID: "sea_b", IsNaval: true, MovePoints: 3, MaxMovePoints: 3, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "enemy_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 3, MaxMovePoints: 3, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
	)

	g.assignNavalMission("player_fleet", army.NavalMissionPatrol, "sea_b", "")
	if g.gs.PendingNavalContact == nil {
		t.Fatal("aynı denizde görev atanırken temas oluşmalı")
	}
	g.resolveNavalContactChoice(1)

	fleet := g.gs.Armies["player_fleet"]
	if fleet.RegionID == "sea_b" || fleet.MovePoints != 1 {
		t.Fatalf("geri çekilen filo başka denize geçip 2 hareket puanı harcamalı: region=%s mp=%d", fleet.RegionID, fleet.MovePoints)
	}
}

func TestNavalContactWithdrawPrefersUnoccupiedSea(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "player_fleet", OwnerID: "p1", RegionID: "sea_b", IsNaval: true, MovePoints: 3, MaxMovePoints: 3, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "enemy_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 3, MaxMovePoints: 3, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
	)
	g.gs.Regions["sea_a"].Neighbors = []world.RegionID{"sea_b"}
	g.gs.Regions["sea_b"].Neighbors = []world.RegionID{"sea_a", "sea_c", "enemy_coast"}
	g.gs.Regions["sea_c"] = &world.Region{ID: "sea_c", IsSea: true, Neighbors: []world.RegionID{"sea_b"}}
	g.gs.Armies["blocking_fleet"] = &army.Army{ID: "blocking_fleet", OwnerID: "p2", RegionID: "sea_a", IsNaval: true, MovePoints: 3, MaxMovePoints: 3, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}}

	g.assignNavalMission("player_fleet", army.NavalMissionPatrol, "sea_b", "")
	g.resolveNavalContactChoice(1)

	if fleet := g.gs.Armies["player_fleet"]; fleet.RegionID != "sea_c" {
		t.Fatalf("geri çekilme düşmansız denizi tercih etmeliydi: region=%s", fleet.RegionID)
	}
}

func TestAIContactClashResolvesWhenEnteringPlayerSea(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "ai_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 0, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "player_fleet", OwnerID: "p1", RegionID: "sea_b", IsNaval: true, MovePoints: 0, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}},
	)
	contact := g.gs.BeginNavalContact(g.gs.Armies["ai_fleet"], g.gs.Armies["player_fleet"], "sea_b", "sea_a", state.NavalContactMovement)
	if contact == nil {
		t.Fatal("AI filosu oyuncu filosuna girdiğinde temas oluşmalı")
	}
	contact.AttackerDecision = state.NavalContactClash
	contact.DefenderDecision = state.NavalContactClash

	step := ai.ResolveNavalContactBattle(g.gs, "ai_fleet", "sea_b")
	if step.Message == "" {
		t.Fatal("oyuncu çatışmayı kabul ettiğinde AI deniz savaşı çözülmeli")
	}
}

func TestUnassignedEnemyContactWithdrawsWhenOutmatched(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "attacker_fleet", OwnerID: "p1", RegionID: "sea_a", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "defender_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "transport", CurrentHP: 100}}},
	)

	g.moveArmy("attacker_fleet", "sea_b")
	if g.gs.PendingNavalContact == nil || g.gs.PendingNavalContact.DefenderDecision != state.NavalContactWithdraw {
		t.Fatalf("güçsüz görevsiz AI filosu geri çekilmeyi seçmeliydi: %+v", g.gs.PendingNavalContact)
	}
	g.resolveNavalContactChoice(0)
	if len(g.gs.Armies) != 2 || g.gs.Armies["defender_fleet"].RegionID != "sea_c" || g.gs.Armies["defender_fleet"].MovePoints != 0 {
		t.Fatalf("geri çekilen görevsiz AI filosu komşu denize dönüp hareket puanı harcamalı: %+v", g.gs.Armies["defender_fleet"])
	}
}

func TestUnassignedEnemyContactClashesWhenComparable(t *testing.T) {
	g := newNavalBattleGame(
		&army.Army{ID: "attacker_fleet", OwnerID: "p1", RegionID: "sea_a", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
		&army.Army{ID: "defender_fleet", OwnerID: "p2", RegionID: "sea_b", IsNaval: true, MovePoints: 1, MaxMovePoints: 1, Units: []army.Unit{{TypeID: "warship", CurrentHP: 100}}},
	)

	g.moveArmy("attacker_fleet", "sea_b")
	if g.gs.PendingNavalContact == nil || g.gs.PendingNavalContact.DefenderDecision != state.NavalContactClash {
		t.Fatalf("güçleri yakın görevsiz AI filosu çatışmayı seçmeliydi: %+v", g.gs.PendingNavalContact)
	}
	g.resolveNavalContactChoice(0)
	if len(g.gs.Armies) != 1 {
		t.Fatalf("iki taraf da çatışmayı seçtiğinde muharebe çözülmeli, filo sayısı=%d", len(g.gs.Armies))
	}
}
