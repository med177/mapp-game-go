package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func currentRegionTaskTestRenderer(fortified bool) *Renderer {
	target := &world.Region{ID: "enemy_region", OwnerID: "enemy", NameTR: "Düşman Bölgesi"}
	if fortified {
		target.Buildings = []string{"walls"}
	}
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			target.ID: target,
		},
		Armies: map[army.ArmyID]*army.Army{
			"attacker": {ID: "attacker", OwnerID: "player", RegionID: target.ID},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}
	return &Renderer{gs: gs}
}

func TestCurrentRegionTaskOpensSiegeForFortifiedHeldEnemyRegion(t *testing.T) {
	r := currentRegionTaskTestRenderer(true)
	attacker := r.gs.Armies["attacker"]
	target := r.gs.Regions[attacker.RegionID]

	if !r.openCurrentRegionArmyTask(attacker, target) {
		t.Fatal("tahkimli düşman bölgesinde aynı bölge görevi açılmalıydı")
	}
	if !r.regionTaskDialog.show || r.regionTaskDialog.buttons[0].Label != "Kuşatma" || r.regionTaskDialog.actions[0].Kind != ActionStartSiege {
		t.Fatalf("tahkimli hedefte görev modali açılmalıydı: %+v", r.regionTaskDialog)
	}
	if r.regionTaskDialog.buttons[3].Label != "Vazgeç" || r.regionTaskDialog.actions[3].Kind != ActionNone {
		t.Fatal("görev modalında vazgeç seçeneği bulunmalı ve oyun aksiyonu üretmemeli")
	}
}

func TestCurrentRegionArmyTaskHoveringUsesCurrentRegionTaskAvailability(t *testing.T) {
	oldWorldW, oldWorldH := WorldW, WorldH
	defer func() {
		WorldW, WorldH = oldWorldW, oldWorldH
	}()
	WorldW, WorldH = 1, 1

	r := currentRegionTaskTestRenderer(false)
	r.SelectedArmy = "attacker"
	r.worldMap = &WorldMap{
		regionAt:  []uint16{1},
		regionIDs: []world.RegionID{"", "enemy_region"},
	}
	r.camScale = 1

	if !r.currentRegionArmyTaskHovering(ScreenWidth/2, ScreenHeight/2) {
		t.Fatal("görev verilebilir mevcut bölge üzerinde cursor hover durumu etkin olmalı")
	}
}

func TestCurrentRegionArmyTaskIndicatorHiddenForBesiegingArmy(t *testing.T) {
	r := currentRegionTaskTestRenderer(true)
	attacker := r.gs.Armies["attacker"]
	target := r.gs.Regions[attacker.RegionID]
	r.gs.Sieges = map[world.RegionID]*state.SiegeState{
		target.ID: {RegionID: target.ID, AttackerArmyID: attacker.ID},
	}

	if r.currentRegionArmyTaskIndicatorVisible(attacker, target) {
		t.Fatal("aktif kuşatma yürüten orduda aynı-bölge görev göstergesi görünmemeli")
	}
}

func TestCurrentRegionTaskOffersDirectCaptureForUnfortifiedHeldEnemyRegion(t *testing.T) {
	r := currentRegionTaskTestRenderer(false)
	attacker := r.gs.Armies["attacker"]
	target := r.gs.Regions[attacker.RegionID]

	if !r.openCurrentRegionArmyTask(attacker, target) {
		t.Fatal("tahkimatsız düşman bölgesinde ele geçirme görevi açılmalıydı")
	}
	if !r.regionTaskDialog.show || r.regionTaskDialog.buttons[0].Label != "Ele Geçir" || r.regionTaskDialog.actions[0].Kind != ActionCaptureRegion {
		t.Fatalf("tahkimatsız hedefte ele geçir aksiyonu bekleniyordu: %+v", r.regionTaskDialog)
	}
}

func TestCurrentRegionTaskOpensBattlePlanWhenUnfortifiedRegionHasEnemyArmy(t *testing.T) {
	r := currentRegionTaskTestRenderer(false)
	target := r.gs.Regions["enemy_region"]
	r.gs.Armies["attacker"].MovePoints = 1
	r.gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "enemy", RegionID: target.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
	}
	r.gs.UnitTypes = map[string]*army.UnitType{
		"infantry": {ID: "infantry", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50},
	}

	if !r.openCurrentRegionArmyTask(r.gs.Armies["attacker"], target) {
		t.Fatal("düşman ordusu varken aynı bölge çatışma görevi açılmalıydı")
	}
	if !r.battlePlan.show || r.battlePlan.actionKind != ActionMoveArmy || !r.battlePlan.contactResolved {
		t.Fatalf("düşman ordusu varken ele geçirme yerine temas çözülmüş savaş planı açılmalıydı: %+v", r.battlePlan)
	}
}

func TestCurrentRegionTaskOpensBattlePlanAgainstEnemyArmyInOwnedFortifiedRegion(t *testing.T) {
	r := currentRegionTaskTestRenderer(true)
	target := r.gs.Regions["enemy_region"]
	target.OwnerID = "player"
	r.gs.Armies["attacker"].MovePoints = 1
	r.gs.Armies["defender"] = &army.Army{
		ID: "defender", OwnerID: "enemy", RegionID: target.ID,
		Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
	}
	r.gs.UnitTypes = map[string]*army.UnitType{
		"infantry": {ID: "infantry", Category: army.CategoryInfantry, Attack: 10, Defense: 10, Morale: 50},
	}

	if !r.openCurrentRegionArmyTask(r.gs.Armies["attacker"], target) {
		t.Fatal("kendi tahkimatlı bölgesindeki düşman orduya saldırı planı açılmalıydı")
	}
	if !r.battlePlan.show || !r.battlePlan.contactResolved || r.battlePlan.battleContext != combat.BattleContextLand {
		t.Fatalf("kendi bölgesindeki düşman için kara muharebesi planı kurulmadı: %+v", r.battlePlan)
	}
	if r.regionTaskDialog.show || r.confirmDialog.show {
		t.Fatal("kendi tahkimatlı bölgesindeki düşman saldırısı kuşatma/görev modalı açmamalıydı")
	}
}

func TestSelectMapRegionDoesNotOpenRecruitPanel(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"bursa": {ID: "bursa", OwnerID: "player"},
			},
		},
		SelectedRegion:   "ankara",
		showRecruitPanel: true,
		recruitUnitID:    "militia",
		recruitQty:       3,
	}

	r.selectMapRegion("bursa")
	r.selectMapRegion("bursa")

	if r.showRecruitPanel {
		t.Fatal("bölge seçimi recruit panelini açmamalı")
	}
	if r.recruitUnitID != "" || r.recruitQty != 1 {
		t.Fatalf("bölge seçimi recruit seçimini temizlemeli: unit=%q qty=%d", r.recruitUnitID, r.recruitQty)
	}
}

func TestEmbarkedArmyBadgeHitSelectsTransportedArmyView(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet": {
				ID:            "fleet",
				OwnerID:       "player",
				RegionID:      "sea",
				IsNaval:       true,
				Units:         []army.Unit{{TypeID: "transport"}},
				EmbarkedUnits: []army.Unit{{TypeID: "infantry"}},
			},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			regionAnchor: map[world.RegionID][2]int{
				"sea": {100, 100},
			},
		},
	}

	positions := r.armyIconPositions()
	if len(positions) != 1 {
		t.Fatalf("tek filo ikonu bekleniyordu, got=%d", len(positions))
	}
	badge := navalEmbarkedArmyBadgeRect(positions[0].X, positions[0].Y)
	aid, ok := r.embarkedArmyHitAt(badge.X+badge.W/2, badge.Y+badge.H/2)
	if !ok || aid != "fleet" {
		t.Fatalf("taşınan ordu karesi filoyu taşıdığı ordu görünümüne yönlendirmeli: aid=%q hit=%t", aid, ok)
	}
}

func TestSelectMapRegionDefaultsToExpandedNeighborList(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			Regions: map[world.RegionID]*world.Region{
				"bursa":  {ID: "bursa"},
				"ankara": {ID: "ankara"},
			},
		},
		SelectedRegion:          "ankara",
		devNeighborListExpanded: false,
	}

	r.selectMapRegion("bursa")
	if !r.devNeighborListExpanded {
		t.Fatal("yeni bölge seçildiğinde komşu listesi varsayılan olarak genişletilmiş gelmeli")
	}

	r.devNeighborListExpanded = false
	r.selectMapRegion("bursa")
	if r.devNeighborListExpanded {
		t.Fatal("aynı bölge yeniden seçildiğinde kullanıcının daraltma tercihi korunmalı")
	}
}

func TestMapRegionDoubleClickOpensDiplomacyForForeignRegion(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{
			PlayerFactionID: "player",
			Factions: map[faction.FactionID]*faction.Faction{
				"player": {ID: "player", NameTR: "Oyuncu"},
				"enemy":  {ID: "enemy", NameTR: "Düşman"},
			},
			Regions: map[world.RegionID]*world.Region{
				"enemy_region": {ID: "enemy_region", OwnerID: "enemy"},
				"home_region":  {ID: "home_region", OwnerID: "player", Buildings: []string{"barracks"}},
			},
			UnitTypes: map[string]*army.UnitType{
				"infantry": {ID: "infantry", RequiredBldg: "barracks", RequiredBldgLevel: 1},
			},
		},
	}

	if r.selectMapRegionFromMapClick("enemy_region") {
		t.Fatal("ilk yabancı bölge tıklaması diplomasi açmamalı")
	}
	if r.showDiplomacy {
		t.Fatal("ilk tıklamada diplomasi paneli açılmamalı")
	}

	if !r.selectMapRegionFromMapClick("enemy_region") {
		t.Fatal("ikinci aynı bölge tıklaması çift tıklama olarak algılanmalı")
	}
	if !r.showDiplomacy || r.diplomacyTargetFaction != "enemy" {
		t.Fatalf("yabancı bölge çift tıklaması diplomasi hedefini açmalı: show=%t target=%q", r.showDiplomacy, r.diplomacyTargetFaction)
	}

	r.CloseDiplomacyPanel()
	if r.selectMapRegionFromMapClick("home_region") {
		t.Fatal("ilk oyuncu bölgesi tıklaması Ordu panelini açmamalı")
	}
	if !r.selectMapRegionFromMapClick("home_region") {
		t.Fatal("oyuncunun kendi bölgesine çift tıklama Ordu kısayolunu tetiklemeli")
	}
	if r.showDiplomacy || !r.showRecruitPanel {
		t.Fatalf("oyuncu bölgesi çift tıklaması Ordu davranışını açmalı: diplomacy=%t recruit=%t", r.showDiplomacy, r.showRecruitPanel)
	}
}
