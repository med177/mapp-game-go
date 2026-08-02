package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func navalMissionPanelStateFixture() *state.GameState {
	return &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"enemy":  {ID: "enemy"},
		},
		UnitTypes: map[string]*army.UnitType{
			"warship":   {ID: "warship", Category: army.CategoryNavalWar},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 10},
		},
		Regions: map[world.RegionID]*world.Region{
			"sea":         {ID: "sea", IsSea: true, Neighbors: []world.RegionID{"enemy_coast"}},
			"enemy_coast": {ID: "enemy_coast", OwnerID: "enemy", Neighbors: []world.RegionID{"sea"}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
		Armies: map[army.ArmyID]*army.Army{
			"war":       {ID: "war", OwnerID: "player", RegionID: "sea", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
			"transport": {ID: "transport", OwnerID: "player", IsNaval: true, Units: []army.Unit{{TypeID: "transport"}}},
			"enemy":     {ID: "enemy", OwnerID: "enemy", IsNaval: true, Units: []army.Unit{{TypeID: "warship"}}},
		},
	}
}

func TestNavalMissionButtonOnlyTargetsEligiblePlayerFleet(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720

	gs := navalMissionPanelStateFixture()
	button := navalMissionButtonRect(armyPanelGeometry(), false)
	x, y := button.X+button.W/2, button.Y+button.H/2
	if !navalMissionButtonHit(x, y, gs, "war") {
		t.Fatal("oyuncunun savaş filosunda görev butonu hit olmalıydı")
	}
	if navalMissionButtonHit(x, y, gs, "transport") {
		t.Fatal("sırf nakliye filosunda görev butonu olmamalıydı")
	}
	if navalMissionButtonHit(x, y, gs, "enemy") {
		t.Fatal("düşman filosunda görev butonu hit olmamalıydı")
	}
}

func TestNavalMissionOptionsExposeWarshipTasks(t *testing.T) {
	gs := navalMissionPanelStateFixture()
	options := navalMissionOptions(gs, gs.Armies["war"])
	if len(options) != 3 {
		t.Fatalf("savaş filosu için devriye, abluka ve escort bekleniyordu, %d seçenek var", len(options))
	}
	if options[0].kind != army.NavalMissionPatrol || options[1].kind != army.NavalMissionBlockade || options[2].kind != army.NavalMissionEscort {
		t.Fatalf("savaş filosu görevleri beklenen sırada değil: %+v", options)
	}
	if options[0].effect == "" || options[1].effect == "" || options[2].effect == "" {
		t.Fatalf("savaş filosu görev bonusları görünür olmalı: %+v", options)
	}
	if options[0].effect != "Etki: aynı denizdeki abluka filosuyla otomatik savaşır; ticaret ve lojistik kesintisini dengeler." {
		t.Fatalf("devriye bonusu beklenen açıklamayı taşımıyor: %q", options[0].effect)
	}
	if options[1].effect != "Etki: savaş gemisi başına -%50 ticaret; azami -%100." {
		t.Fatalf("abluka bonusu beklenen açıklamayı taşımıyor: %q", options[1].effect)
	}
	gs.Armies["transport"].EmbarkedUnits = []army.Unit{{TypeID: "infantry"}}
	transportOptions := navalMissionOptions(gs, gs.Armies["transport"])
	if len(transportOptions) != 0 {
		t.Fatalf("sırf nakliye filosunda görev seçeneği olmamalıydı: %+v", transportOptions)
	}
}

func TestNavalMissionTargetCircleUsesPointerHitRadius(t *testing.T) {
	if !navalMissionTargetCircleHit(100, 100, 108, 100) {
		t.Fatal("hedef yuvarlağının üzerinde cursor parmak olmalı")
	}
	if navalMissionTargetCircleHit(100, 100, 113, 100) {
		t.Fatal("hedef yuvarlağından uzak cursor parmak olmamalı")
	}
}

func TestNavalMissionOptionsHideBlockadeOutsideValidCurrentSea(t *testing.T) {
	gs := navalMissionPanelStateFixture()
	gs.Armies["war"].RegionID = "open_sea"
	gs.Regions["open_sea"] = &world.Region{ID: "open_sea", IsSea: true}
	options := navalMissionOptions(gs, gs.Armies["war"])
	for _, option := range options {
		if option.kind == army.NavalMissionBlockade {
			t.Fatalf("açık deniz konumunda abluka görünmemeli: %+v", option)
		}
	}
}

func TestNavalMissionPanelStaysInsideViewport(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() { ScreenWidth, ScreenHeight = oldW, oldH }()
	ScreenWidth, ScreenHeight = 1280, 720

	layout := navalMissionPanelLayoutFor(12)
	if layout.panelX < 0 || layout.panelY < 0 || layout.panelX+layout.panelW > float32(ScreenWidth) || layout.panelY+layout.panelH > float32(ScreenHeight) {
		t.Fatalf("donanma görev paneli viewport dışına taştı: %+v", layout)
	}
}

func TestNavalMissionPanelThreeLineRowsHaveVerticalClearance(t *testing.T) {
	if navalMissionPanelRowH < 80 {
		t.Fatalf("üç satırlı görev seçenekleri için satır yüksekliği yetersiz: %.1f", navalMissionPanelRowH)
	}

	rowRectHeight := navalMissionPanelRowH - 10
	if rowRectHeight < 60 {
		t.Fatalf("görev satırının iç kutusu bonus satırına sığmıyor: %.1f", rowRectHeight)
	}
}

func TestNavalMissionPanelClearButtonUsesCommanderStylePlacement(t *testing.T) {
	layout := navalMissionPanelLayoutFor(3)
	if layout.clear.Label != "Görevi Kaldır" {
		t.Fatalf("görev temizleme düğmesi beklenen etiketi taşımıyor: %+v", layout.clear)
	}
	if layout.clear.W != commanderPanelButtonW || layout.clear.H != commanderPanelButtonH {
		t.Fatalf("görev temizleme düğmesi ortak kaldırma düğmesi ölçüsünde olmalı: %+v", layout.clear)
	}
	if layout.clear.Y <= float64(layout.rowY) {
		t.Fatalf("görev temizleme düğmesi satırların altında olmalı: %+v", layout.clear)
	}
	if layout.panelW >= 720 {
		t.Fatalf("donanma görev paneli daraltılmalı: %.1f", layout.panelW)
	}
}

func TestNavalMissionPanelUsesSharedCloseIconButton(t *testing.T) {
	layout := navalMissionPanelLayoutFor(1)
	if layout.close.Label != "" || layout.close.Icon != gameui.IconClose {
		t.Fatalf("donanma görev paneli ortak kapatma ikonunu kullanmalı: %+v", layout.close)
	}
	if layout.close.IconSize != 13 {
		t.Fatalf("donanma görev kapatma ikonu diğer panellerdeki boyutu kullanmalı: %.1f", layout.close.IconSize)
	}
}

func TestNavalMissionPanelRowGeometryMatchesExpandedHeight(t *testing.T) {
	layout := navalMissionPanelLayoutFor(2)
	first := navalMissionPanelRowRect(layout, 0)
	second := navalMissionPanelRowRect(layout, 1)
	if first.H != float64(navalMissionPanelRowH-10) {
		t.Fatalf("ilk görev satırı yüksekliği ortak row geometry ile eşleşmiyor: %+v", first)
	}
	if second.Y-first.Y != float64(navalMissionPanelRowH) {
		t.Fatalf("görev satırları genişletilmiş row yüksekliğini kullanmıyor: first=%+v second=%+v", first, second)
	}
}
