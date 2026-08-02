package state

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func navalMissionStateFixture() *GameState {
	return &GameState{
		PlayerFactionID: "player",
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
			"infantry":  {ID: "infantry", Category: army.CategoryInfantry, Embarkable: true},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea":    {ID: "sea", IsSea: true},
			"coast":  {ID: "coast", OwnerID: "enemy", Neighbors: []world.RegionID{"sea"}},
			"inland": {ID: "inland"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"war-fleet": {
				ID: "war-fleet", OwnerID: "player", RegionID: "sea", IsNaval: true,
				Units: []army.Unit{{TypeID: "warship"}},
			},
			"transport-fleet": {
				ID: "transport-fleet", OwnerID: "player", IsNaval: true,
				Units: []army.Unit{{TypeID: "transport"}},
			},
		},
	}
}

func TestCanAssignNavalMissionValidatesPlayerFleetRoles(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["coast"].Neighbors = []world.RegionID{"sea"}
	gs.Regions["sea"].Neighbors = []world.RegionID{"coast"}
	gs.Armies["transport-fleet"].EmbarkedUnits = []army.Unit{{TypeID: "infantry"}}

	cases := []struct {
		name    string
		mission army.NavalMission
	}{
		{name: "devriye", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}},
		{name: "abluka", mission: army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}},
		{name: "escort", mission: army.NavalMission{Kind: army.NavalMissionEscort, TargetFleetID: "transport-fleet"}},
		{name: "nakliye", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if ok, reason := gs.AssignNavalMission(map[bool]army.ArmyID{true: "war-fleet", false: "transport-fleet"}[tc.name == "devriye" || tc.name == "abluka" || tc.name == "escort"], tc.mission); !ok {
				t.Fatalf("görev atanamadı: %s", reason)
			}
		})
	}
	if gs.Armies["war-fleet"].NavalMission == nil || gs.Armies["transport-fleet"].NavalMission == nil {
		t.Fatal("geçerli görevler filolarda saklanmadı")
	}
}

func TestCanAssignNavalMissionRejectsInvalidTargetsAndRoles(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["coast"].Neighbors = []world.RegionID{"sea"}

	invalid := []struct {
		name    string
		fleetID army.ArmyID
		mission army.NavalMission
	}{
		{name: "savaş gemisi olmayan filo devriye atamaz", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}},
		{name: "kara bölgesi devriye hedefi olamaz", fleetID: "war-fleet", mission: army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "inland"}},
		{name: "nakliye için taşınan ordu gerekir", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "coast"}},
		{name: "nakliye iç bölgeye atanamaz", fleetID: "transport-fleet", mission: army.NavalMission{Kind: army.NavalMissionTransport, TargetRegionID: "inland"}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			if ok, reason := gs.CanAssignNavalMission(tc.fleetID, tc.mission); ok {
				t.Fatalf("geçersiz görev kabul edildi: %s", reason)
			}
		})
	}
}

func TestNavalBlockadeRequiresEnemyCoastalSea(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["coast"].Neighbors = []world.RegionID{"sea"}
	if ok, reason := gs.CanAssignNavalMission("war-fleet", army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}); !ok {
		t.Fatalf("düşman kıyısına komşu deniz abluka için geçerli olmalı: %s", reason)
	}
	gs.Regions["open_sea"] = &world.Region{ID: "open_sea", IsSea: true}
	if ok, reason := gs.CanAssignNavalMission("war-fleet", army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "open_sea"}); ok {
		t.Fatal("okyanus ortası düşman kıyısına bağlı olmayan deniz abluka kabul etmemeli")
	} else if reason == "" {
		t.Fatal("geçersiz abluka hedefi açıklama döndürmeli")
	}
}

func TestPatrolAndBlockadeMustUseCurrentOpenSea(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["other_sea"] = &world.Region{ID: "other_sea", IsSea: true}

	for _, kind := range []army.NavalMissionKind{army.NavalMissionPatrol, army.NavalMissionBlockade} {
		if ok, reason := gs.CanAssignNavalMission("war-fleet", army.NavalMission{Kind: kind, TargetRegionID: "other_sea"}); ok {
			t.Fatalf("%s görevi filonun bulunduğu deniz dışına atanamamalı", kind)
		} else if reason == "" {
			t.Fatalf("%s görevi geçersiz hedef için neden döndürmeli", kind)
		}
	}
	gs.Armies["war-fleet"].DockedRegionID = "coast"
	if ok, reason := gs.CanAssignNavalMission("war-fleet", army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}); ok {
		t.Fatal("limanda bağlı filo devriye görevi alamamalı")
	} else if reason == "" {
		t.Fatal("limandaki filo için neden döndürülmeli")
	}
}

func TestClearNavalMissionRemovesPlayerTask(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Regions["sea"].IsSea = true
	if ok, reason := gs.AssignNavalMission("war-fleet", army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}); !ok {
		t.Fatalf("test görevi atanamadı: %s", reason)
	}
	if !gs.ClearNavalMission("war-fleet") || gs.Armies["war-fleet"].NavalMission != nil {
		t.Fatal("görev temizlenmedi")
	}
}

func TestNavalEscortDefenseBonusAppliesOnlyToSameSeaTransport(t *testing.T) {
	gs := navalMissionStateFixture()
	gs.Armies["war-fleet"].RegionID = "sea"
	gs.Armies["transport-fleet"].RegionID = "sea"
	gs.Armies["war-fleet"].NavalMission = &army.NavalMission{
		Kind: army.NavalMissionEscort, TargetFleetID: "transport-fleet",
	}
	if got := gs.NavalEscortDefenseBonus([]army.ArmyID{"war-fleet", "transport-fleet"}, "sea"); got != 0.15 {
		t.Fatalf("aynı denizde escort %%15 savunma bonusu vermeli: %.2f", got)
	}
	gs.Armies["war-fleet"].RegionID = "other-sea"
	if got := gs.NavalEscortDefenseBonus([]army.ArmyID{"war-fleet", "transport-fleet"}, "sea"); got != 0 {
		t.Fatalf("farklı denizde escort bonusu uygulanmamalı: %.2f", got)
	}
}

func TestNavalFleetsAutoEngageOnlyPatrolAndBlockadePair(t *testing.T) {
	gs := &GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {
				FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar,
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player": {ID: "player", OwnerID: "player", RegionID: "sea", IsNaval: true},
			"enemy":  {ID: "enemy", OwnerID: "enemy", RegionID: "sea", IsNaval: true},
		},
	}

	if gs.NavalFleetsAutoEngage(gs.Armies["player"], gs.Armies["enemy"]) {
		t.Fatal("görevsiz iki filo aynı denizde otomatik savaş başlatmamalı")
	}
	gs.Armies["player"].NavalMission = &army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: "sea"}
	gs.Armies["enemy"].NavalMission = &army.NavalMission{Kind: army.NavalMissionBlockade, TargetRegionID: "sea"}
	if !gs.NavalFleetsAutoEngage(gs.Armies["player"], gs.Armies["enemy"]) {
		t.Fatal("devriye aynı denizde düşman ablukasını otomatik yakalamalı")
	}
	gs.Armies["player"].NavalMission = &army.NavalMission{Kind: army.NavalMissionEscort, TargetRegionID: "sea"}
	if gs.NavalFleetsAutoEngage(gs.Armies["player"], gs.Armies["enemy"]) {
		t.Fatal("escort görevi düşman ablukasına kendiliğinden saldırmamalı")
	}
}

func TestNavalContactMissionDefaults(t *testing.T) {
	patrol := &army.Army{ID: "patrol", IsNaval: true, NavalMission: &army.NavalMission{Kind: army.NavalMissionPatrol}}
	blockade := &army.Army{ID: "blockade", IsNaval: true, NavalMission: &army.NavalMission{Kind: army.NavalMissionBlockade}}
	free := &army.Army{ID: "free", IsNaval: true}

	if got := navalContactDefaultDecision(patrol, NavalContactMovement); got != NavalContactClash {
		t.Fatalf("devriye varsayılanı çatış olmalı: %q", got)
	}
	if got := navalContactDefaultDecision(blockade, NavalContactMovement); got != NavalContactHold {
		t.Fatalf("abluka varsayılanı pozisyonu koru olmalı: %q", got)
	}
	if got := navalContactDefaultDecision(free, NavalContactMovement); got != NavalContactClash {
		t.Fatalf("görevsiz düşman filosu temas varsayılanında çatışmalı: %q", got)
	}
}

func TestQueueNavalContactForWarFindsSameSeaFleets(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea":      {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"safe_sea"}},
			"safe_sea": {ID: "safe_sea", IsSea: true, Neighbors: []world.RegionID{"sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-fleet": {ID: "player-fleet", OwnerID: "player", RegionID: "sea", IsNaval: true},
			"enemy-fleet":  {ID: "enemy-fleet", OwnerID: "enemy", RegionID: "sea", IsNaval: true},
		},
	}

	contact := gs.QueueNavalContactForWar("player", "enemy")
	if contact == nil || contact.Trigger != NavalContactWarOpening || contact.PlayerArmyID != "player-fleet" {
		t.Fatalf("savaş açılışında aynı denizdeki filolar temasa alınmalı: %+v", contact)
	}
	if contact.AttackerDecision != NavalContactUndecided && contact.DefenderDecision != NavalContactUndecided {
		t.Fatal("oyuncu filosunun kararı temas açılışında kararsız kalmalı")
	}
}

func TestNavalContactWithdrawRequiresMovementPoint(t *testing.T) {
	gs := &GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea":      {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"safe_sea"}},
			"safe_sea": {ID: "safe_sea", IsSea: true, Neighbors: []world.RegionID{"sea"}},
		},
		Armies: map[army.ArmyID]*army.Army{
			"player-fleet": {ID: "player-fleet", OwnerID: "player", RegionID: "sea", IsNaval: true},
			"enemy-fleet":  {ID: "enemy-fleet", OwnerID: "enemy", RegionID: "sea", IsNaval: true},
		},
	}
	contact := gs.BeginNavalContact(gs.Armies["player-fleet"], gs.Armies["enemy-fleet"], "sea", "", NavalContactWarOpening)
	if contact == nil {
		t.Fatal("deniz teması oluşturulmalı")
	}
	if gs.NavalContactDecisionForPlayer(contact, NavalContactWithdraw) {
		t.Fatal("hareket puanı olmayan filo geri çekilmeyi seçememeli")
	}
	gs.Armies["player-fleet"].MovePoints = 1
	if !gs.NavalContactDecisionForPlayer(contact, NavalContactWithdraw) {
		t.Fatal("hareket puanı olan filo geri çekilmeyi seçebilmeli")
	}
}
