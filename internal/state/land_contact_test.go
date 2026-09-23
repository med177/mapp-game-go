package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func TestLandContactAmbushCannotBeDeclinedByEnteringArmy(t *testing.T) {
	from := &world.Region{ID: "from"}
	target := &world.Region{ID: "target"}
	attacker := &army.Army{ID: "attacker", OwnerID: "enemy", RegionID: from.ID}
	ambusher := &army.Army{ID: "ambusher", OwnerID: "player", RegionID: target.ID, InAmbush: true}
	gs := &GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			from.ID:   from,
			target.ID: target,
		},
		Armies: map[army.ArmyID]*army.Army{
			attacker.ID: attacker,
			ambusher.ID: ambusher,
		},
	}
	contact := gs.BeginLandContact(attacker, ambusher, target.ID, from.ID, LandContactMovement)
	if contact == nil {
		t.Fatal("pusu kara teması oluşturulamadı")
	}

	// Eski/ara karar state'inde AI saldıranı geri çekilmeye ayarlasa bile,
	// pusuya giren taraf oyuncu Çatış dediğinde bu karar korunmamalıdır.
	contact.AttackerDecision = LandContactWithdraw
	if !gs.LandContactDecisionForPlayer(contact, LandContactClash) {
		t.Fatal("oyuncunun pusu temasındaki çatış kararı uygulanmadı")
	}
	if contact.AttackerDecision != LandContactClash {
		t.Fatalf("pusuya giren taraf geri çekilmeyi korudu: %q", contact.AttackerDecision)
	}
	if !gs.LandContactWillClash(contact) {
		t.Fatal("pusuya giren taraf reddedemediği halde temas çatışmaya dönüşmedi")
	}
}

func TestLandContactRetreatRegionRejectsBlockedTerrainArea(t *testing.T) {
	current := &world.Region{
		ID:        "izmir",
		Neighbors: []world.RegionID{"area::ege_mountains"},
	}
	blocked := &world.Region{
		ID:            "area::ege_mountains",
		IsTerrainArea: true,
		TerrainAreaID: "ege_mountains",
	}
	armyRef := &army.Army{ID: "defender", OwnerID: "byzantium", RegionID: current.ID}
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			current.ID: current,
			blocked.ID: blocked,
		},
		TerrainAreas: []world.TerrainArea{{ID: "ege_mountains", MoveCost: 0}},
	}

	if retreat := gs.LandContactRetreatRegion(armyRef); retreat != "" {
		t.Fatalf("blocked terrain area was selected as retreat target: %s", retreat)
	}
}
