package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"
)

func TestCollectActiveWarSummariesShowsTurnsStrengthAndArmyCounts(t *testing.T) {
	gs := &state.GameState{
		Turn: 9,
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A Devleti"},
			"b": {ID: "b", NameTR: "B Devleti"},
		},
		Regions: map[world.RegionID]*world.Region{
			"a-region": {ID: "a-region", OwnerID: "a"},
			"b-region": {ID: "b-region", OwnerID: "b"},
		},
		Armies: map[army.ArmyID]*army.Army{
			"a-army": {ID: "a-army", OwnerID: "a", Units: []army.Unit{{TypeID: "inf"}, {TypeID: "inf"}}},
			"b-army": {ID: "b-army", OwnerID: "b", Units: []army.Unit{{TypeID: "inf"}}, EmbarkedUnits: []army.Unit{{TypeID: "inf"}}},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("b", "a"): {FactionA: "b", FactionB: "a", Stance: faction.StanceWar},
		},
		WarLedgers: map[string]*state.WarLedger{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", DeclarerFactionID: "b", DefenderFactionID: "a", StartedTurn: 4, CasualtiesA: 3, CasualtiesB: 5},
		},
	}

	wars := collectActiveWarSummaries(gs, nil)
	if len(wars) != 1 {
		t.Fatalf("tek aktif savaş bekleniyordu, got=%d", len(wars))
	}
	war := wars[0]
	if war.FactionANameTR != "B Devleti" || war.FactionBNameTR != "A Devleti" {
		t.Fatalf("taraf adları yanlış: %+v", war)
	}
	if war.Turns != 5 || war.ArmiesA != 1 || war.ArmiesB != 1 || war.UnitsA != 2 || war.UnitsB != 2 {
		t.Fatalf("savaş süresi/ordu sayıları yanlış: %+v", war)
	}
	if war.CasualtiesA != 5 || war.CasualtiesB != 3 {
		t.Fatalf("kayıplar yanlış: %+v", war)
	}
}

func TestCollectActiveWarSummariesUsesRelationDirectionForLegacyLedger(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A Devleti"},
			"b": {ID: "b", NameTR: "B Devleti"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "b", FactionB: "a", Stance: faction.StanceWar},
		},
		WarLedgers: map[string]*state.WarLedger{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b"},
		},
	}

	wars := collectActiveWarSummaries(gs, nil)
	if len(wars) != 1 || wars[0].FactionA != "b" || wars[0].FactionB != "a" {
		t.Fatalf("legacy ilişki yönü korunmalıydı: %+v", wars)
	}
}

func TestActiveWarsPanelDoesNotCoverWarHUDButton(t *testing.T) {
	button := buildActiveWarsHUDButton()
	panel := activeWarsPanelRect()
	if panel.Hit(button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("aktif savaş paneli açılmadan önce HUD savaş düğmesinin üzerine taşmamalı")
	}
	if !activeWarsHudButtonHit(button.X+button.W/2, button.Y+button.H/2) {
		t.Fatal("aktif savaş HUD düğmesi kendi merkezinde tıklanabilir olmalı")
	}
}

func TestActiveWarsHUDButtonUsesUtilitySlotAfterMusic(t *testing.T) {
	oldWidth := ScreenWidth
	ScreenWidth = 1920
	defer func() { ScreenWidth = oldWidth }()

	musicX, _, musicW, _ := musicHudRect()
	button := buildActiveWarsHUDButton()
	if button.X != float64(musicX+musicW+activeWarsHudButtonGap) {
		t.Fatalf("aktif savaş düğmesi müzik HUD'unun sağındaki yardımcı slota yerleşmeli: musicRight=%.0f buttonX=%.0f", musicX+musicW, button.X)
	}
	if button.W != activeWarsHudButtonSize || button.H != activeWarsHudButtonSize || button.Y != activeWarsHudButtonTop {
		t.Fatalf("aktif savaş düğmesi boyut/margin yanlış: %+v", button)
	}
}

func TestActiveWarsPanelStaysLeftOfEventLog(t *testing.T) {
	panel := activeWarsPanelRect()
	if panel.X+panel.W > float64(evLogX()) {
		t.Fatalf("aktif savaş paneli olaylar paneliyle yatayda çakışmamalı: panelRight=%.0f eventLogX=%.0f", panel.X+panel.W, evLogX())
	}
}

func TestActiveWarRowKeepsCenteredTextBetweenFactionFlags(t *testing.T) {
	row := gameui.Rect{X: 100, Y: 40, W: activeWarsPanelW - activeWarsPanelPad*2, H: activeWarRowH}
	leftFlag, center, rightFlag := activeWarRowContentRects(row)
	if leftFlag.W != activeWarFlagSize || rightFlag.W != activeWarFlagSize || leftFlag.H != rightFlag.H {
		t.Fatalf("bayrak alanları kare olmalı: left=%+v right=%+v", leftFlag, rightFlag)
	}
	if center.X <= leftFlag.X+leftFlag.W || center.X+center.W >= rightFlag.X {
		t.Fatalf("metin alanı iki bayrağın arasında kalmalı: left=%+v center=%+v right=%+v", leftFlag, center, rightFlag)
	}
}

func TestActiveWarsPanelLeavesOutsideMapPointAvailable(t *testing.T) {
	panel := activeWarsPanelRect()
	mapX := panel.X - 24
	mapY := panel.Y + 120
	if activeWarsPanelHit(mapX, mapY) {
		t.Fatal("panel dışındaki harita noktası aktif savaş paneli tarafından tüketilmemeli")
	}
}

func TestActiveWarRowAtReturnsClickedVisibleWar(t *testing.T) {
	viewport := activeWarsPanelViewport()
	wars := []ActiveWarSummary{{FactionA: "a", FactionB: "b"}, {FactionA: "c", FactionB: "d"}}
	row := activeWarRowRect(viewport, 1)
	if got := activeWarRowAt(row.X+row.W/2, row.Y+row.H/2, wars, 0); got != 1 {
		t.Fatalf("ikinci görünen savaş satırı seçilmeliydi: got=%d", got)
	}
}

func TestActiveWarRepresentativeRegionPrefersCapitalThenDeterministicRegion(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"z-region": {ID: "z-region", OwnerID: "faction"},
			"a-region": {ID: "a-region", OwnerID: "faction"},
		},
	}
	if region := activeWarRepresentativeRegion(gs, "faction"); region == nil || region.ID != "a-region" {
		t.Fatalf("başkent yokken deterministik ilk bölge seçilmeliydi: %+v", region)
	}
}
