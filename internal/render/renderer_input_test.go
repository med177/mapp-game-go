package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

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
