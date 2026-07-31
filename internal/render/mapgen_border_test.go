package render

import (
	"testing"

	"mapp-game-go/internal/army"
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

func TestSelectedBorderStyleCoversBothSidesOfRegionBoundary(t *testing.T) {
	oldW, oldH := WorldW, WorldH
	WorldW, WorldH = 4, 4
	defer func() {
		WorldW, WorldH = oldW, oldH
	}()

	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"upper_left": {ID: "upper_left", OwnerID: "other"},
			"selected":   {ID: "selected", OwnerID: "player"},
		},
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"other":  {ID: "other"},
		},
	}
	wm := &WorldMap{
		regionAt:  make([]uint16, WorldW*WorldH),
		regionIDs: []world.RegionID{"", "upper_left", "selected"},
	}
	for y := 0; y < 2; y++ {
		wm.regionAt[y*WorldW] = 1
		wm.regionAt[y*WorldW+1] = 1
		wm.regionAt[y*WorldW+2] = 2
		wm.regionAt[y*WorldW+3] = 2
	}
	for y := 2; y < WorldH; y++ {
		for x := 0; x < WorldW; x++ {
			wm.regionAt[y*WorldW+x] = 2
		}
	}

	wm.rebuildBorderSegments(gs)
	wm.updateBorderStyles(gs, "selected", MapModeNormal)

	if !hasBorderSegment(wm, 2, 0, 2, 2, mapBorderStyleSelected) {
		t.Fatal("seçili bölgenin sağ tarafında kalan dikey sınır sarı seçili stil almalı")
	}
	if !hasBorderSegment(wm, 0, 2, 2, 2, mapBorderStyleSelected) {
		t.Fatal("seçili bölgenin alt tarafında kalan yatay sınır sarı seçili stil almalı")
	}
}

func TestSelectedMapBorderUsesThreePixelStroke(t *testing.T) {
	if got := mapBorderStyleStrokeWidth(mapBorderStyleSelected); got != 3 {
		t.Fatalf("seçili border kalınlığı yanlış: got=%v want=3", got)
	}
	if got := mapBorderStyleStrokeWidth(mapBorderStyleStrong); got != mapBorderStrokeWidth {
		t.Fatalf("normal border kalınlığı değişmemeli: got=%v want=%v", got, mapBorderStrokeWidth)
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

func TestEnemyNavalRegionSetMarksOnlyOpenWarFleets(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
			"ally":   {ID: "ally"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_open":   {ID: "enemy_open", OwnerID: "enemy", RegionID: "sea_red", IsNaval: true},
			"enemy_docked": {ID: "enemy_docked", OwnerID: "enemy", RegionID: "sea_red", DockedRegionID: "port", IsNaval: true},
			"ally_open":    {ID: "ally_open", OwnerID: "ally", RegionID: "sea_blue", IsNaval: true},
		},
	}

	marked := enemyNavalRegionSet(gs)
	if !marked["sea_red"] {
		t.Fatal("savaş halindeki açık düşman filosunun deniz bölgesi işaretlenmeliydi")
	}
	if marked["sea_blue"] {
		t.Fatal("müttefik filosunun deniz bölgesi düşman olarak işaretlenmemeliydi")
	}
	if got := len(marked); got != 1 {
		t.Fatalf("yalnız açık düşman filosunun bölgesi işaretlenmeliydi: %+v", marked)
	}

	baseSignature := borderDiplomacySignature(gs)
	gs.Armies["enemy_open"].RegionID = "sea_other"
	if movedSignature := borderDiplomacySignature(gs); movedSignature == baseSignature {
		t.Fatal("filo deniz bölgesi değişince harita cache imzası değişmeliydi")
	}
}

func TestEnemyNavalRegionUsesEnemyBorderStyle(t *testing.T) {
	oldW, oldH := WorldW, WorldH
	WorldW, WorldH = 3, 3
	defer func() {
		WorldW, WorldH = oldW, oldH
	}()

	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea_red":     {ID: "sea_red", IsSea: true},
			"player_land": {ID: "player_land", OwnerID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"enemy_fleet": {ID: "enemy_fleet", OwnerID: "enemy", RegionID: "sea_red", IsNaval: true},
		},
	}
	wm := &WorldMap{
		regionAt:  make([]uint16, WorldW*WorldH),
		regionIDs: []world.RegionID{"", "sea_red", "player_land"},
	}
	for y := 0; y < WorldH; y++ {
		wm.regionAt[y*WorldW] = 1
		wm.regionAt[y*WorldW+1] = 2
		wm.regionAt[y*WorldW+2] = 2
	}

	wm.rebuildBorderSegments(gs)
	wm.updateBorderStyles(gs, "", MapModeNormal)
	for _, style := range wm.borderStyles {
		if style == mapBorderStyleEnemy {
			goto enemyBorderFound
		}
	}
	t.Fatalf("düşman filosu bulunan deniz hücresinin sınırı enemy stilinde olmalıydı: %+v", wm.borderStyles)

enemyBorderFound:
	wm.updateBorderStyles(gs, "sea_red", MapModeNormal)
	for _, style := range wm.borderStyles {
		if style == mapBorderStyleSelected {
			return
		}
	}
	t.Fatalf("seçili düşman denizinin kara sınırı sarı selected stilini korumalıydı: %+v", wm.borderStyles)
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
