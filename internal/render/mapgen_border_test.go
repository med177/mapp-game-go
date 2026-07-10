package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestDrawRegionBordersHighlightsPlayerRealmAndAlliedRealms(t *testing.T) {
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
		dispPixels: make([]byte, WorldW*WorldH*4),
		regionAt:   make([]uint16, WorldW*WorldH),
		regionIDs:  []world.RegionID{"", "player_land", "vassal_land", "ally_land", "enemy_land", "ally_vassal_land"},
	}
	base := color.RGBA{100, 100, 100, 255}
	for i := 0; i < WorldW*WorldH; i++ {
		wm.dispPixels[i*4] = base.R
		wm.dispPixels[i*4+1] = base.G
		wm.dispPixels[i*4+2] = base.B
		wm.dispPixels[i*4+3] = base.A
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

	wm.drawRegionBorders(gs, "", MapModeNormal)
	row := 3
	assertBorderPixel(t, wm, 3, row, blendedBorderColor(base, color.RGBA{35, 22, 10, 255}, 52))
	assertBorderPixel(t, wm, 5, row, base)
	assertBorderPixel(t, wm, 6, row, color.RGBA{255, 205, 74, 255})
	assertBorderPixel(t, wm, 7, row, base)
	assertBorderPixel(t, wm, 8, row, base)
	assertBorderPixel(t, wm, 8, 1, color.RGBA{82, 210, 166, 255})
	assertBorderPixel(t, wm, 10, row, color.RGBA{218, 62, 54, 255})
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

func assertBorderPixel(t *testing.T, wm *WorldMap, x, y int, want color.RGBA) {
	t.Helper()
	i := (y*WorldW + x) * 4
	got := color.RGBA{wm.dispPixels[i], wm.dispPixels[i+1], wm.dispPixels[i+2], wm.dispPixels[i+3]}
	if got != want {
		t.Fatalf("(%d,%d) sınır rengi yanlış: got=%v want=%v", x, y, got, want)
	}
}

func blendedBorderColor(base, target color.RGBA, alpha byte) color.RGBA {
	return color.RGBA{
		R: blend(base.R, target.R, alpha),
		G: blend(base.G, target.G, alpha),
		B: blend(base.B, target.B, alpha),
		A: 255,
	}
}
