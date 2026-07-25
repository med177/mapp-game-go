package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestVectorBorderStylesHighlightPlayerRealmAndAlliedRealms(t *testing.T) {
	oldW, oldH := WorldW, WorldH
	WorldW, WorldH = 12, 7
	defer func() {
		WorldW, WorldH = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player":      {ID: "player"},
			"vassal":      {ID: "vassal", OverlordID: "player"},
			"ally":        {ID: "ally"},
			"ally_vassal": {ID: "ally_vassal", OverlordID: "ally"},
			"enemy":       {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"player_land":      {ID: "player_land", OwnerID: "player"},
			"vassal_land":      {ID: "vassal_land", OwnerID: "vassal"},
			"ally_land":        {ID: "ally_land", OwnerID: "ally"},
			"enemy_land":       {ID: "enemy_land", OwnerID: "enemy"},
			"ally_vassal_land": {ID: "ally_vassal_land", OwnerID: "ally_vassal"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}
	wm := &WorldMap{
		regionAt:  make([]uint16, WorldW*WorldH),
		regionIDs: []world.RegionID{"", "player_land", "vassal_land", "ally_land", "enemy_land", "ally_vassal_land"},
	}
	for y := 1; y < WorldH-1; y++ {
		for x := 1; x <= 3; x++ {
			wm.regionAt[y*WorldW+x] = 1
		}
		for x := 4; x <= 6; x++ {
			wm.regionAt[y*WorldW+x] = 2
		}
		for x := 7; x <= 9; x++ {
			wm.regionAt[y*WorldW+x] = 3
		}
		wm.regionAt[y*WorldW+10] = 4
	}

	realms, affiliations := buildBorderDiplomacyContext(gs, wm.regionIDs)
	if realms[2] != "player" || affiliations[2] != borderAffiliationPlayerRealm {
		t.Fatalf("oyuncu vassalı oyuncu realm konturuna dahil olmalı: realm=%q affiliation=%d", realms[2], affiliations[2])
	}
	if realms[5] != "ally" || affiliations[5] != borderAffiliationAlly {
		t.Fatalf("müttefik vassalı müttefik konturuna dahil olmalı: realm=%q affiliation=%d", realms[5], affiliations[5])
	}
	if affiliations[4] != borderAffiliationEnemy {
		t.Fatalf("savaştaki devlet düşman konturu almalı: affiliation=%d", affiliations[4])
	}

	wm.rebuildBorderSegments(gs)
	wm.updateBorderStyles(gs, "", MapModeNormal)
	if !hasBorderSegment(wm, 4, 1, 4, 6, mapBorderStyleSubtle) {
		t.Fatal("oyuncu realm içindeki vassal sınırı sıkıştırılmış subtle kontur olarak bulunamadı")
	}
	if !hasBorderSegment(wm, 7, 1, 7, 6, mapBorderStylePlayerRealm) {
		t.Fatal("oyuncu-ally sınırı oyuncu realm rengiyle bulunamadı")
	}
	if !hasBorderSegment(wm, 10, 1, 10, 6, mapBorderStyleEnemy) {
		t.Fatal("ally-enemy sınırı düşman rengiyle bulunamadı")
	}
	if !hasBorderSegment(wm, 7, 1, 10, 1, mapBorderStyleAlly) {
		t.Fatal("ally kıyı sınırı ally rengiyle bulunamadı")
	}
	if len(wm.borderSegments) >= (WorldW*WorldH)/2 {
		t.Fatalf("kontur parçaları yeterince sıkıştırılmamış: %d", len(wm.borderSegments))
	}
}

func TestBorderDiplomacySignatureChangesWithAllianceAndVassalage(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"other":  {ID: "other"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "other"): {FactionA: "player", FactionB: "other", Stance: faction.StancePeace},
		},
	}
	base := borderDiplomacySignature(gs)
	gs.Relations[faction.RelationKey("player", "other")].Stance = faction.StanceAllied
	allied := borderDiplomacySignature(gs)
	if allied == base {
		t.Fatal("ittifak değişimi sınır cache imzasını değiştirmeli")
	}
	gs.Factions["other"].OverlordID = "player"
	if vassal := borderDiplomacySignature(gs); vassal == allied {
		t.Fatal("vassallık değişimi sınır cache imzasını değiştirmeli")
	}
}

func TestFarZoomSkipsMinorBorderStyles(t *testing.T) {
	for _, style := range []uint8{mapBorderStyleSubtle, mapBorderStyleTradeSubtle, mapBorderStyleSea} {
		if shouldDrawMapBorderStyle(style, 0.7) {
			t.Fatalf("uzak zoom'da stil çizilmemeli: %d", style)
		}
	}
	for _, style := range []uint8{mapBorderStyleSelected, mapBorderStylePlayerRealm, mapBorderStyleAlly, mapBorderStyleEnemy, mapBorderStyleStrong} {
		if !shouldDrawMapBorderStyle(style, 0.7) {
			t.Fatalf("uzak zoom'da önemli stil korunmalı: %d", style)
		}
	}
	if !shouldDrawMapBorderStyle(mapBorderStyleSubtle, 1.0) {
		t.Fatal("yakın zoom'da subtle sınır korunmalı")
	}
}

func hasBorderSegment(wm *WorldMap, x1, y1, x2, y2 float32, style uint8) bool {
	for i, segment := range wm.borderSegments {
		if i >= len(wm.borderStyles) || wm.borderStyles[i] != style {
			continue
		}
		if segment.x1 == x1 && segment.y1 == y1 && segment.x2 == x2 && segment.y2 == y2 {
			return true
		}
	}
	return false
}
