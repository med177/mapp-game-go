package render

import (
	"image/color"
	"strings"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestArmySpriteSetFollowsFactionReligion(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"sunni":    {ID: "sunni", Religion: religion.Sunni},
			"shia":     {ID: "shia", Religion: religion.Shia},
			"catholic": {ID: "catholic", Religion: religion.Catholic},
			"orthodox": {ID: "orthodox", Religion: religion.Orthodox},
			"unknown":  {ID: "unknown"},
		},
	}

	checks := []struct {
		owner string
		want  armySpriteSet
	}{
		{owner: "sunni", want: armySpriteSetEastern},
		{owner: "shia", want: armySpriteSetEastern},
		{owner: "catholic", want: armySpriteSetWestern},
		{owner: "orthodox", want: armySpriteSetWestern},
		{owner: "unknown", want: armySpriteSetWestern},
		{owner: "missing", want: armySpriteSetLegacy},
	}
	for _, check := range checks {
		if got := armySpriteSetForFaction(gs, check.owner); got != check.want {
			t.Fatalf("%s için army sprite grubu yanlış: got=%v want=%v", check.owner, got, check.want)
		}
	}
}

func TestUnitSpriteAssetNamesCoverAllUnitTypes(t *testing.T) {
	want := map[string]string{
		"militia":        "infantry_light.png",
		"infantry":       "infantry_medium.png",
		"elite_infantry": "infantry_heavy.png",
		"light_cavalry":  "cavalry_light.png",
		"cavalry":        "cavalry_medium.png",
		"heavy_cavalry":  "cavalry_heavy.png",
		"catapult":       "siege_trebuchet.png",
		"bombard":        "siege_mortar.png",
		"cannon":         "siege_cannon.png",
		"transport":      "ship_transport.png",
		"merchant_ship":  "ship_merchant.png",
		"warship":        "ship_war_galley.png",
	}
	for unitID, filename := range want {
		if got := unitSpriteAssetNames[unitID]; got != filename {
			t.Fatalf("%s sprite eşlemesi yanlış: got=%q want=%q", unitID, got, filename)
		}
	}
}

func TestArmyIconBorderColorUsesDiplomaticPalette(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
			"enemy":  {ID: "enemy"},
			"vassal": {ID: "vassal", OverlordID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}

	if got := armyIconBorderColor(gs, "player", false); got != borderColorPlayerRealm {
		t.Fatalf("player border rengi yanlış: got=%v want=%v", got, borderColorPlayerRealm)
	}
	if got := armyIconBorderColor(gs, "vassal", false); got != borderColorPlayerRealm {
		t.Fatalf("vassal border rengi yanlış: got=%v want=%v", got, borderColorPlayerRealm)
	}
	if got := armyIconBorderColor(gs, "ally", false); got != borderColorAlly {
		t.Fatalf("ally border rengi yanlış: got=%v want=%v", got, borderColorAlly)
	}
	if got := armyIconBorderColor(gs, "enemy", false); got != borderColorEnemy {
		t.Fatalf("enemy border rengi yanlış: got=%v want=%v", got, borderColorEnemy)
	}
	if got := armyIconBorderColor(gs, "enemy", true); got != (color.RGBA{255, 215, 0, 255}) {
		t.Fatalf("selected army border rengi sabit altın olmalı: got=%v", got)
	}
}

func TestArmyCommanderBadgeAlignsAboveLandAndNavalIcons(t *testing.T) {
	_, landY, landSize := armyCommanderBadgeRect(100, 100, false, false)
	if landY+landSize != 88 {
		t.Fatalf("kara komutan rozeti sayı karesinin üstünde bitmeli: bottom=%.1f", landY+landSize)
	}

	_, navalY, navalSize := armyCommanderBadgeRect(100, 100, true, false)
	if navalY+navalSize != 87 {
		t.Fatalf("deniz komutan rozeti dairenin üstünde bitmeli: bottom=%.1f", navalY+navalSize)
	}

	_, embarkedY, embarkedSize := armyCommanderBadgeRect(100, 100, true, true)
	if embarkedY+embarkedSize != 87 {
		t.Fatalf("taşıma rozeti olan filoda komutan portresi dairenin üstünde kalmalı: bottom=%.1f", embarkedY+embarkedSize)
	}
}

func TestArmySelectionIndicatorRectCoversCommanderAndMarkerForLandAndNaval(t *testing.T) {
	for _, check := range []struct {
		name    string
		isNaval bool
	}{
		{name: "kara", isNaval: false},
		{name: "deniz", isNaval: true},
	} {
		t.Run(check.name, func(t *testing.T) {
			rect := armySelectionIndicatorRect(100, 100, check.isNaval, true)
			portraitX, portraitY, portraitSize := armyCommanderBadgeRect(100, 100, check.isNaval, false)
			if rect.X >= float64(portraitX) || rect.Y >= float64(portraitY) || rect.X+rect.W <= float64(portraitX+portraitSize) || rect.Y+rect.H <= float64(portraitY+portraitSize) {
				t.Fatalf("seçim çerçevesi komutan portresini sarmalı: rect=%+v portrait=(%.1f, %.1f, %.1f)", rect, portraitX, portraitY, portraitSize)
			}
			if !rect.Hit(100, 100) {
				t.Fatalf("seçim çerçevesi marker merkezini kapsamalı: %+v", rect)
			}
		})
	}
}

func TestArmySelectionIndicatorRectWithoutCommanderStaysCompact(t *testing.T) {
	withoutCommander := armySelectionIndicatorRect(100, 100, false, false)
	withCommander := armySelectionIndicatorRect(100, 100, false, true)
	if withoutCommander.H >= withCommander.H {
		t.Fatalf("komutansız seçim çerçevesi portreli çerçeveden kısa olmalı: without=%+v with=%+v", withoutCommander, withCommander)
	}
}

func TestNavalArmyBadgesShareUpperRightAnchor(t *testing.T) {
	badge := navalEmbarkedArmyBadgeRect(100, 100)
	if badge.X <= 100 || badge.Y >= 100 {
		t.Fatalf("taşınan ordu rozeti markerın sağ üstünde olmalı: %+v", badge)
	}
	if badge.W != 16 || badge.H != 16 {
		t.Fatalf("taşınan ordu rozeti 16 px kare olmalı: %+v", badge)
	}
	if badge.X+badge.W/2 != 114 || badge.Y+badge.H/2 != 86 {
		t.Fatalf("taşınan ordu rozeti ortak üst-sağ anchor'da olmalı: %+v", badge)
	}

	bonus := navalMissionBonusBadgeRect(100, 100)
	if bonus.X+bonus.W/2 != badge.X+badge.W/2 || bonus.Y+bonus.H/2 != badge.Y+badge.H/2 {
		t.Fatalf("bonus rozeti taşınan ordu rozetiyle aynı anchor'ı paylaşmalı: embarked=%+v bonus=%+v", badge, bonus)
	}
	merchant := merchantTradeBonusBadgeRect(100, 100)
	if merchant.X+merchant.W/2 != bonus.X+bonus.W/2 || merchant.Y+merchant.H/2 != bonus.Y+bonus.H/2 {
		t.Fatalf("ticaret rozeti diğer görev rozetleriyle aynı anchor'ı paylaşmalı: merchant=%+v bonus=%+v", merchant, bonus)
	}
}

func TestNavalDamageBadgeUsesUpperLeftAnchor(t *testing.T) {
	x, y := navalDamageBadgeCenter(100, 100)
	if x >= 100 || y >= 100 {
		t.Fatalf("zayiat rozeti sol üstte olmalı: x=%.1f y=%.1f", x, y)
	}
	if x != 86 || y != 86 {
		t.Fatalf("zayiat rozeti ortak sol-üst konumda olmalı: x=%.1f y=%.1f", x, y)
	}
}

func TestArmyDamageBadgeUsesUpperLeftAnchor(t *testing.T) {
	x, y := armyDamageBadgeCenter(100, 100)
	if x >= 100 || y >= 100 {
		t.Fatalf("zayiat rozeti sol üstte olmalı: x=%.1f y=%.1f", x, y)
	}
	if x != 86 || y != 86 {
		t.Fatalf("kara ordusu zayiat rozeti donanmayla aynı sol-üst konumda olmalı: x=%.1f y=%.1f", x, y)
	}
}

func TestNavalEmbarkedArmyBadgeFollowsFleetVisibility(t *testing.T) {
	fleet := &army.Army{EmbarkedUnits: make([]army.Unit, 17)}
	if got := navalEmbarkedArmyBadgeText(fleet, false); got != "?" {
		t.Fatalf("görünmeyen düşman taşıma sayısı ? olmalı: got=%q", got)
	}
	if got := navalEmbarkedArmyBadgeText(fleet, true); got != "17" {
		t.Fatalf("görünür taşıma sayısı gerçek adet olmalı: got=%q", got)
	}
}

func TestArmyTaskStatusBadgeUsesUpperRightAnchor(t *testing.T) {
	badge := armyTaskStatusBadgeRect(100, 100)
	if badge.X <= 100 || badge.Y >= 100 {
		t.Fatalf("görev rozeti markerın sağ üstünde olmalı: %+v", badge)
	}
	if badge.X+badge.W/2 != 114 || badge.Y+badge.H/2 != 86 {
		t.Fatalf("görev rozeti ortak sağ-üst anchor'da olmalı: %+v", badge)
	}
}

func TestCurrentRegionArmyTaskBadgeUsesLowerLeftAnchor(t *testing.T) {
	x, y := currentRegionArmyTaskBadgeCenter(100, 100)
	if x != 84 || y != 116 {
		t.Fatalf("aynı-bölge görev kılıcı markerın alt-soluna taşınmalı: x=%.1f y=%.1f", x, y)
	}
}

func TestArmyIconPositionsSeparateAdjacentTaskBadges(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"enemy": {
				ID: "enemy", OwnerID: "enemy",
				Settlements: []world.Settlement{{ID: "center", IsCenter: true}},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ambusher": {ID: "ambusher", OwnerID: "player", RegionID: "enemy", InAmbush: true},
			"other":    {ID: "other", OwnerID: "player", RegionID: "enemy"},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			settlementAnchor:  map[settlementAnchorKey][2]int{{Region: "enemy", Index: 0}: {100, 100}},
			primarySettlement: map[world.RegionID][2]int{"enemy": {100, 100}},
		},
	}

	positions := r.armyIconPositions()
	if len(positions) != 2 {
		t.Fatalf("iki kara ordusu bekleniyordu, got=%d", len(positions))
	}
	if got := positions[1].X - positions[0].X; got != armyTaskStatusIconStep {
		t.Fatalf("görev rozeti taşıyan marker grubunda aralık büyütülmeli: got=%.1f want=%.1f", got, armyTaskStatusIconStep)
	}
	left := armyTaskStatusBadgeRect(positions[0].X, positions[0].Y)
	right := armyTaskStatusBadgeRect(positions[1].X, positions[1].Y)
	if left.X+left.W > right.X {
		t.Fatalf("yan yana görev rozetleri üst üste binmemeli: left=%+v right=%+v", left, right)
	}
}

func TestArmyTaskStatusUsesRaidArmyIDAndShowsLootTooltip(t *testing.T) {
	gs := &state.GameState{
		Turn:            4,
		Month:           6,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"enemy": {ID: "enemy", OwnerID: "enemy", NameTR: "Düşman", BaseGoldIncome: 100, TaxRate: 100, Satisfaction: 50, BaseGrainOutput: 20},
		},
		Armies: map[army.ArmyID]*army.Army{
			"raider": {ID: "raider", OwnerID: "player", RegionID: "enemy"},
			"other":  {ID: "other", OwnerID: "player", RegionID: "enemy"},
		},
		Raids: map[world.RegionID]*state.RaidState{
			"enemy": {RegionID: "enemy", RaiderFactionID: "player", RaiderArmyID: "raider", Turn: 4},
		},
	}

	if activeRaidForArmy(gs, gs.Armies["raider"]) == nil {
		t.Fatal("yağma rozeti yağmayı başlatan orduya bağlanmalı")
	}
	if activeRaidForArmy(gs, gs.Armies["other"]) != nil {
		t.Fatal("aynı bölgedeki başka ordu yağma rozeti almamalı")
	}
	title, detail, ok := armyTaskStatusTooltipText(gs, gs.Armies["raider"])
	if !ok || title != "Yağmalama" || !strings.Contains(detail, "+80 altın") || !strings.Contains(detail, "+10 tahıl") {
		t.Fatalf("yağma tooltip'i gerçek kazancı göstermeli: title=%q detail=%q ok=%t", title, detail, ok)
	}
}

func TestArmyTaskStatusShowsAmbushBadgeAndTooltip(t *testing.T) {
	gs := &state.GameState{
		Turn:            4,
		PlayerFactionID: "player",
		Regions: map[world.RegionID]*world.Region{
			"enemy": {ID: "enemy", OwnerID: "enemy", NameTR: "Düşman"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"ambusher": {ID: "ambusher", OwnerID: "player", RegionID: "enemy", InAmbush: true},
		},
	}

	if !armyTaskStatusVisible(gs, gs.Armies["ambusher"]) {
		t.Fatal("oyuncunun pusu ordusu görev rozeti göstermeli")
	}
	title, detail, ok := armyTaskStatusTooltipText(gs, gs.Armies["ambusher"])
	if !ok || title != "Pusu" || !strings.Contains(detail, "düşmandan gizleniyor") {
		t.Fatalf("pusu tooltip'i görev etkisini göstermeli: title=%q detail=%q ok=%t", title, detail, ok)
	}
}
