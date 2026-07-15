package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestScoutedEnemyRevealCountUsesSeventyFivePercentForSiegeIntel(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, false, 0.75); got != 6 {
		t.Fatalf("8 birim için %%75 görünürlük 6 olmalıydı, got=%d", got)
	}
	if got := scoutedEnemyRevealCount(1, false, 0.75); got != 1 {
		t.Fatalf("tek birim için en az 1 görünürlük korunmalıydı, got=%d", got)
	}
}

func TestCommanderPortraitHitRectStaysInsideCommanderColumn(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {ID: "a1", OwnerID: "osm", RegionID: "ankara", Units: []army.Unit{{TypeID: "infantry", CurrentHP: 100}}},
		},
	}

	rect, ok := commanderPortraitHitRect(gs, "a1")
	if !ok {
		t.Fatal("commander portrait rect bekleniyordu")
	}
	layout := armyPanelGeometry()
	if rect.X < float64(layout.commanderX) || rect.X+rect.W > float64(layout.commanderX+layout.commanderW) {
		t.Fatalf("portrait rect commander kolonunu tasiyor: rect=%+v layout=%+v", rect, layout)
	}
	if rect.Y < float64(layout.commanderY) || rect.Y+rect.H > float64(layout.commanderY+layout.commanderH) {
		t.Fatalf("portrait rect commander kartini tasiyor: rect=%+v layout=%+v", rect, layout)
	}
}

func TestScoutedEnemyRevealCountUsesFullIntelWhenEnabled(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, true, 0.75); got != 8 {
		t.Fatalf("tam istihbaratta tüm birimler görünmeliydi, got=%d", got)
	}
}

func TestArmyPanelBoundsHitCoversWholePanel(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:       "a1",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units: []army.Unit{
					{TypeID: "infantry", CurrentHP: 100},
				},
			},
		},
	}

	rect, ok := armyDetailPanelRect(gs, "a1")
	if !ok {
		t.Fatal("ordu panel rect'i bekleniyordu")
	}
	if !ArmyPanelBoundsHit(rect.X+rect.W/2, rect.Y+rect.H/2, gs, "a1") {
		t.Fatalf("panel merkezi hit olmalıydı: rect=%+v", rect)
	}
	if ArmyPanelBoundsHit(rect.X-4, rect.Y-4, gs, "a1") {
		t.Fatalf("panel dışı hit olmamalıydı: rect=%+v", rect)
	}
}

func TestArmyPanelInteractiveHitIgnoresEmptyPanelArea(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a1": {
				ID:       "a1",
				OwnerID:  "osm",
				RegionID: "ankara",
				Units:    []army.Unit{{TypeID: "infantry", CurrentHP: 100}},
			},
		},
	}

	layout := armyPanelGeometry()
	fx := float64(layout.gridX + cardW*4)
	fy := float64(layout.gridY + cardH + 20)
	if !ArmyPanelBoundsHit(fx, fy, gs, "a1") {
		t.Fatal("secilen nokta panel bounds icinde olmaliydi")
	}
	if ArmyPanelInteractiveHit(fx, fy, gs, "a1") {
		t.Fatalf("bos panel alani interactive sayilmamali: x=%.1f y=%.1f", fx, fy)
	}
}

func TestEnemyArmyPanelBoundsHitUsesIntelAwareRects(t *testing.T) {
	oldW, oldH := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth = oldW
		ScreenHeight = oldH
	}()
	ScreenWidth = 1280
	ScreenHeight = 720

	gs := &state.GameState{
		PlayerFactionID: "osm",
		Factions: map[faction.FactionID]*faction.Faction{
			"osm": {ID: "osm", NameTR: "Osmanlı"},
			"byz": {ID: "byz", NameTR: "Bizans"},
		},
		Regions: map[world.RegionID]*world.Region{
			"ankara": {ID: "ankara", NameTR: "Ankara", OwnerID: "osm", Neighbors: []world.RegionID{"bursa"}},
			"bursa":  {ID: "bursa", NameTR: "Bursa", OwnerID: "byz", Neighbors: []world.RegionID{"ankara"}},
			"izmir":  {ID: "izmir", NameTR: "İzmir", OwnerID: "byz"},
		},
		Relations: map[string]*faction.Relation{},
		Armies: map[army.ArmyID]*army.Army{
			"player": {ID: "player", OwnerID: "osm", RegionID: "ankara", MovePoints: 2, Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"near":   {ID: "near", OwnerID: "byz", RegionID: "bursa", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
			"far":    {ID: "far", OwnerID: "byz", RegionID: "izmir", Units: []army.Unit{{TypeID: "inf", CurrentHP: 100}}},
		},
	}

	nearRect, ok := armyDetailPanelRect(gs, "near")
	if !ok {
		t.Fatal("yakın düşman için rect bekleniyordu")
	}
	farRect, ok := armyDetailPanelRect(gs, "far")
	if !ok {
		t.Fatal("uzak düşman için rect bekleniyordu")
	}
	if nearRect != farRect {
		t.Fatalf("düşman panelleri artık aynı ortak geometriyi paylaşmalı: near=%+v far=%+v", nearRect, farRect)
	}
	if !ArmyPanelBoundsHit(nearRect.X+nearRect.W/2, nearRect.Y+nearRect.H/2, gs, "near") {
		t.Fatalf("yakın düşman panel merkezi hit olmalıydı: rect=%+v", nearRect)
	}
	if !ArmyPanelBoundsHit(farRect.X+farRect.W/2, farRect.Y+farRect.H/2, gs, "far") {
		t.Fatalf("uzak düşman panel merkezi hit olmalıydı: rect=%+v", farRect)
	}
}
