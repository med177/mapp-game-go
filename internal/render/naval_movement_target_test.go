package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestArmyMovementTargetAtUsesTheTargetMarkerRadius(t *testing.T) {
	start := &world.Region{ID: "start", WorldX: 10, WorldY: 10}
	target := &world.Region{ID: "target", WorldX: 30, WorldY: 30}
	r := &Renderer{
		gs: &state.GameState{Regions: map[world.RegionID]*world.Region{
			start.ID:  start,
			target.ID: target,
		}},
		worldMap: &WorldMap{
			primarySettlement: map[world.RegionID][2]int{target.ID: {1000, 1000}},
		},
		camScale: 1,
	}
	armyUnit := &army.Army{ID: "army", RegionID: start.ID}
	reachability := state.MovementReachability{
		Nodes: map[world.RegionID]state.MovementRouteNode{
			start.ID:  {RegionID: start.ID},
			target.ID: {RegionID: target.ID},
		},
	}

	targetX, targetY := r.movementRegionScreenPos(target.ID, "")
	if got, ok := r.armyMovementTargetAt(float64(targetX)+10, float64(targetY), armyUnit, reachability); !ok || got != target.ID {
		t.Fatalf("hedef halkası içinde target = (%q, %v), want (%q, true)", got, ok, target.ID)
	}
	if got, ok := r.armyMovementTargetAt(float64(targetX)+23, float64(targetY), armyUnit, reachability); ok || got != "" {
		t.Fatalf("hedef halkası dışında target = (%q, %v), want (%q, false)", got, ok, "")
	}
}

func TestMovementPreviewSkipsCursorOverPanel(t *testing.T) {
	r := &Renderer{gs: &state.GameState{}}
	if r.movementPreviewCursorOverPanel(float64(ScreenWidth/2), float64(ScreenHeight/2)) {
		t.Fatal("panel yokken harita cursor'ı panel üzerinde kabul edildi")
	}

	r.showTech = true
	if !r.movementPreviewCursorOverPanel(float64(ScreenWidth/2), float64(ScreenHeight/2)) {
		t.Fatal("teknoloji paneli üzerindeki cursor hareket önizlemesini engellemedi")
	}
}

func TestNavalLandMoveTargetStyleUsesDiplomaticOwnership(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"vassal": {ID: "vassal", OverlordID: "player"},
			"ally":   {ID: "ally"},
			"enemy":  {ID: "enemy"},
			"war":    {ID: "war"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "war"):  {FactionA: "player", FactionB: "war", Stance: faction.StanceWar},
		},
	}
	fleet := &army.Army{OwnerID: "player", IsNaval: true}

	tests := []struct {
		name  string
		owner string
		want  [4]uint8
	}{
		{name: "kendi limanı", owner: "player", want: [4]uint8{80, 160, 255, 160}},
		{name: "vassal limanı", owner: "vassal", want: [4]uint8{80, 160, 255, 160}},
		{name: "müttefik limanı", owner: "ally", want: [4]uint8{80, 160, 255, 160}},
		{name: "yabancı liman", owner: "enemy", want: [4]uint8{220, 140, 30, 210}},
		{name: "savaş limanı", owner: "war", want: [4]uint8{220, 60, 60, 200}},
		{name: "sahipsiz liman", owner: "", want: [4]uint8{60, 220, 60, 200}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := navalLandMoveTargetStyle(gs, fleet, &world.Region{OwnerID: tt.owner})
			want := tt.want
			if [4]uint8{got.R, got.G, got.B, got.A} != want {
				t.Fatalf("renk = (%d, %d, %d, %d), want (%d, %d, %d, %d)", got.R, got.G, got.B, got.A, want[0], want[1], want[2], want[3])
			}
		})
	}
}

func TestNavalLandMoveTargetSettlementDistinguishesLandingCenter(t *testing.T) {
	port := world.Settlement{Type: world.SettlementPort}
	center := world.Settlement{Type: world.SettlementCity, IsCenter: true}
	nonCenter := world.Settlement{Type: world.SettlementCity}

	if !navalLandMoveTargetSettlement(port, false) || !navalLandMoveTargetSettlement(port, true) {
		t.Fatal("liman marker'ı taşıma durumundan bağımsız görünür olmalı")
	}
	if !navalLandMoveTargetSettlement(center, true) {
		t.Fatal("gemide ordu varken merkez settlement çıkarma hedefi olmalı")
	}
	if navalLandMoveTargetSettlement(center, false) || navalLandMoveTargetSettlement(nonCenter, true) {
		t.Fatal("çıkarma olmayan veya merkez olmayan settlement hedef olarak işaretlenmemeli")
	}
}

func TestNavalLandingTargetTooltipText(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
			"enemy":  {ID: "enemy"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"): {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
		},
	}
	fleet := &army.Army{OwnerID: "player", IsNaval: true, EmbarkedUnits: []army.Unit{{TypeID: "infantry"}}}
	if _, _, ok := navalLandingTargetTooltipText(gs, fleet, nil); ok {
		t.Fatal("hedef bölge yokken sevk/çıkarma tooltip'i gösterilmemeli")
	}
	for _, test := range []struct {
		name, owner, title string
	}{
		{name: "kendi bölgesi", owner: "player", title: "Ordu sevk noktası"},
		{name: "müttefik bölgesi", owner: "ally", title: "Ordu sevk noktası"},
		{name: "düşman bölgesi", owner: "enemy", title: "Çıkarma noktası"},
	} {
		t.Run(test.name, func(t *testing.T) {
			title, detail, ok := navalLandingTargetTooltipText(gs, fleet, &world.Region{OwnerID: test.owner})
			if !ok || title != test.title || detail == "" {
				t.Fatalf("tooltip = (%q, %q, %v), beklenen başlık %q", title, detail, ok, test.title)
			}
		})
	}
}

func TestDockedFleetSamePortIsNoOpOnlyWhenEmpty(t *testing.T) {
	fleet := &army.Army{
		IsNaval:        true,
		DockedRegionID: "port",
	}
	if !dockedFleetSamePortIsNoOp(fleet, "port") {
		t.Fatal("boş docked filo kendi limanına yeniden konuşlanma no-op'u olmalı")
	}

	fleet.EmbarkedUnits = []army.Unit{{TypeID: "infantry"}}
	if dockedFleetSamePortIsNoOp(fleet, "port") {
		t.Fatal("kara birimi taşıyan filo kendi limanında çıkarma akışını engellememeli")
	}
	if dockedFleetSamePortIsNoOp(fleet, "başka_liman") {
		t.Fatal("başka liman no-op kabul edilmemeli")
	}
}

func TestEmbarkTargetRequiresMovementPoints(t *testing.T) {
	gs := &state.GameState{
		UnitTypes: map[string]*army.UnitType{
			"infantry":  {ID: "infantry", Embarkable: true},
			"transport": {ID: "transport", Category: army.CategoryNavalTrans, CarryCapacity: 1},
		},
		Regions: map[world.RegionID]*world.Region{
			"coast": {ID: "coast", Neighbors: []world.RegionID{"sea"}},
			"sea":   {ID: "sea", IsSea: true},
		},
	}
	fleet := &army.Army{
		OwnerID:  "player",
		IsNaval:  true,
		RegionID: "sea",
		Units:    []army.Unit{{TypeID: "transport"}},
	}
	selected := &army.Army{
		OwnerID:  "player",
		RegionID: "coast",
		Units:    []army.Unit{{TypeID: "infantry"}},
	}

	selected.MovePoints = 0
	if embarkableFleetForSelectedArmy(gs, selected, fleet) {
		t.Fatal("hareket hakkı olmayan ordu için nakliye hedefi uygun kabul edildi")
	}

	selected.MovePoints = 1
	if !embarkableFleetForSelectedArmy(gs, selected, fleet) {
		t.Fatal("hareket hakkı olan ordu için geçerli nakliye hedefi reddedildi")
	}
}
