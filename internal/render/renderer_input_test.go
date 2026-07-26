package render

import (
	"testing"

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
				"home_region":  {ID: "home_region", OwnerID: "player"},
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
	if r.selectMapRegionFromMapClick("home_region") || r.selectMapRegionFromMapClick("home_region") {
		t.Fatal("oyuncunun kendi bölgesine çift tıklama diplomasi açmamalı")
	}
	if r.showDiplomacy {
		t.Fatal("oyuncu bölgesinde diplomasi paneli açık kalmamalı")
	}
}
