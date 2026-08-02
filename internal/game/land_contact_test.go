package game

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func newLandContactGame(playerAttacker bool) *Game {
	playerOwner := "p1"
	enemyOwner := "p2"
	attackerOwner := enemyOwner
	defenderOwner := playerOwner
	if playerAttacker {
		attackerOwner = playerOwner
		defenderOwner = enemyOwner
	}
	return &Game{
		gs: &state.GameState{
			PlayerFactionID: faction.FactionID(playerOwner),
			Regions: map[world.RegionID]*world.Region{
				"source": {ID: "source", OwnerID: attackerOwner, Neighbors: []world.RegionID{"front"}},
				"front":  {ID: "front", OwnerID: defenderOwner, Neighbors: []world.RegionID{"source", "rear"}},
				"rear":   {ID: "rear", OwnerID: defenderOwner, Neighbors: []world.RegionID{"front"}},
			},
			Armies: map[army.ArmyID]*army.Army{
				"attacker": {ID: "attacker", OwnerID: attackerOwner, RegionID: "source", MovePoints: 2, MaxMovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
				"defender": {ID: "defender", OwnerID: defenderOwner, RegionID: "front", MovePoints: 2, MaxMovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}, {TypeID: "inf", CurrentHP: 100}}},
			},
			Factions: map[faction.FactionID]*faction.Faction{
				"p1": {ID: "p1", NameTR: "Oyuncu"},
				"p2": {ID: "p2", NameTR: "Düşman"},
			},
			Relations: map[string]*faction.Relation{
				faction.RelationKey("p1", "p2"): {FactionA: "p1", FactionB: "p2", Stance: faction.StanceWar},
			},
			UnitTypes: map[string]*army.UnitType{
				"inf": {ID: "inf", Category: army.CategoryInfantry, Attack: 12, Defense: 10, Morale: 50},
			},
		},
		renderer: &render.Renderer{},
	}
}

func TestPlayerLandMovementCreatesContactBeforeBattle(t *testing.T) {
	g := newLandContactGame(true)
	g.moveArmyWithStance("attacker", "front", combat.BattleStanceBalanced)

	contact := g.gs.PendingLandContact
	if contact == nil || contact.PlayerArmyID != "attacker" {
		t.Fatalf("oyuncu kara ordusu hedefteki düşmanla temas popup'ına girmeli: %+v", contact)
	}
	if g.gs.Armies["attacker"].RegionID != "front" || g.gs.Armies["attacker"].MovePoints != 1 || g.gs.Armies["defender"].RegionID != "front" {
		t.Fatal("kara temasında popup açılırken saldıran ordu hedef bölgede ve hareketi tüketilmiş görünmeli")
	}

	g.resolveLandContactChoice(2)
	if g.gs.PendingLandContact != nil || g.gs.Armies["attacker"].RegionID != "front" {
		t.Fatalf("pozisyonu koru temasını savaşa çevirmeden emri sonlandırmalı: contact=%+v attacker=%+v", g.gs.PendingLandContact, g.gs.Armies["attacker"])
	}
}

func TestPlayerLandContactClashResolvesBattle(t *testing.T) {
	g := newLandContactGame(true)
	g.moveArmyWithStance("attacker", "front", combat.BattleStanceBalanced)
	if g.gs.PendingLandContact == nil {
		t.Fatal("savaş öncesi kara teması oluşmalı")
	}
	g.resolveLandContactChoice(0)

	if g.gs.PendingLandContact != nil {
		t.Fatal("çatış kararı sonrası kara teması temizlenmeli")
	}
	attacker := g.gs.Armies["attacker"]
	if attacker == nil || attacker.RegionID != "front" || attacker.MovePoints != 1 {
		t.Fatalf("iki taraf da çatışınca kazanan oyuncu ordusu hedefe ilerlemeli: %+v", attacker)
	}
}

func TestPlayerLandAttackerWithdrawReturnsToSourceAfterEntry(t *testing.T) {
	g := newLandContactGame(true)
	g.moveArmyWithStance("attacker", "front", combat.BattleStanceBalanced)
	if g.gs.PendingLandContact == nil {
		t.Fatal("geri çekilme öncesi kara teması oluşmalı")
	}
	g.resolveLandContactChoice(1)

	attacker := g.gs.Armies["attacker"]
	if g.gs.PendingLandContact != nil || attacker.RegionID != "source" || attacker.MovePoints != 1 {
		t.Fatalf("saldıran geri çekilince kaynak bölgeye dönmeli: contact=%+v attacker=%+v", g.gs.PendingLandContact, attacker)
	}
}

func TestPlayerLandMovementCreatesContactBeforeFortifiedSiege(t *testing.T) {
	g := newLandContactGame(true)
	g.gs.Regions["front"].Buildings = []string{"walls"}

	g.moveArmyWithStance("attacker", "front", combat.BattleStanceBalanced)

	if g.gs.PendingLandContact == nil {
		t.Fatal("tahkimli hedefte kuşatma başlatmadan önce kara teması oluşmalı")
	}
	if g.gs.Sieges != nil && g.gs.SiegeAt("front") != nil {
		t.Fatal("temas kararı verilmeden tahkimli hedefte doğrudan kuşatma başlamamalı")
	}
}

func TestPlayerLandDefenderCanWithdrawFromContact(t *testing.T) {
	g := newLandContactGame(false)
	g.beginLandContact(g.gs.Armies["attacker"], g.gs.Armies["defender"], "front", "source", state.LandContactMovement, true)
	if g.gs.PendingLandContact == nil || g.gs.PendingLandContact.PlayerArmyID != "defender" {
		t.Fatal("AI saldırısı oyuncu savunucuyla kara teması oluşturmalı")
	}
	g.resolveLandContactChoice(1)

	if g.gs.PendingLandContact != nil {
		t.Fatal("savunucu geri çekilince temas temizlenmeli")
	}
	if got := g.gs.Armies["defender"]; got.RegionID != "rear" || got.MovePoints != 1 {
		t.Fatalf("oyuncu savunucusu güvenli komşuya çekilmeli ve 1 hareket harcamalı: %+v", got)
	}
	if got := g.gs.Armies["attacker"]; got.RegionID != "front" {
		t.Fatalf("temas popup'ı açıldığında AI saldıranı hedefte görünmeli: %+v", got)
	}
}
