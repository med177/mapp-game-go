package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestArmyIconPositionsKeepBesiegerLeftOfSplitPart(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"fort": {
				ID:      "fort",
				OwnerID: "p1",
				Settlements: []world.Settlement{
					{ID: "castle", Type: world.SettlementFortress},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"army_p1_9": {
				ID:       "army_p1_9",
				OwnerID:  "p1",
				RegionID: "fort",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
			"army_p1_10": {
				ID:       "army_p1_10",
				OwnerID:  "p1",
				RegionID: "fort",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:       "fort",
				AttackerArmyID: "army_p1_9",
				DefenderArmyID: "army_p1_10",
			},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{
				{Region: "fort", Index: 0}: {100, 100},
			},
			primarySettlement: map[world.RegionID][2]int{
				"fort": {100, 100},
			},
		},
	}

	positions := r.armyIconPositions()
	if len(positions) != 2 {
		t.Fatalf("iki ordu ikonu bekleniyordu, got=%d", len(positions))
	}
	if positions[0].ArmyID != "army_p1_9" || positions[1].ArmyID != "army_p1_10" {
		t.Fatalf("kuşatan ordu solda, ayrılan parça sağda olmalıydı: %+v", positions)
	}
	if got := positions[1].X - positions[0].X; got != 52 {
		t.Fatalf("kuşatma çiftinde kılıç rozeti için arada bir slot bırakılmalıydı: delta=%.1f", got)
	}
	if got := armySiegeBadgeCenterX(positions[0].X, positions[1].X, true); got != positions[0].X+26 {
		t.Fatalf("kılıç rozeti kuşatan ile savunan ordunun ortasında olmalıydı: got=%.1f", got)
	}
}

func TestArmyIconPositionsKeepSiegePairOnFortressAnchor(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"fort": {
				ID:      "fort",
				OwnerID: "p2",
				Settlements: []world.Settlement{
					{ID: "center", IsCenter: true},
					{ID: "castle", Type: world.SettlementFortress},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {
				ID:       "besieger",
				OwnerID:  "p1",
				RegionID: "fort",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
			"defender": {
				ID:       "defender",
				OwnerID:  "p2",
				RegionID: "fort",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:       "fort",
				AttackerArmyID: "besieger",
				DefenderArmyID: "defender",
			},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{
				{Region: "fort", Index: 0}: {100, 100},
				{Region: "fort", Index: 1}: {150, 100},
			},
			primarySettlement: map[world.RegionID][2]int{
				"fort": {100, 100},
			},
		},
	}

	positions := r.armyIconPositions()
	if len(positions) != 2 {
		t.Fatalf("iki kuşatma ordusu ikonu bekleniyordu, got=%d", len(positions))
	}
	if positions[0].ArmyID != "besieger" || positions[1].ArmyID != "defender" {
		t.Fatalf("kuşatma çifti kale anchor'ında kuşatan-savunmacı sırasını korumalıydı: %+v", positions)
	}
	if got := positions[1].X - positions[0].X; got != 52 {
		t.Fatalf("kuşatma çifti tek ortak marker grubunda 52 px ayrılmalıydı: delta=%.1f", got)
	}
	if got := armySiegeBadgeCenterX(positions[0].X, positions[1].X, true); got != positions[0].X+26 {
		t.Fatalf("kılıç rozeti kuşatma çiftinin ortak orta slotunda olmalıydı: got=%.1f", got)
	}
}

func TestArmyIconPositionsLeaveThreePixelsBetweenNavalMarkers(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"sea": {ID: "sea", IsSea: true},
		},
		Armies: map[army.ArmyID]*army.Army{
			"fleet_a": {ID: "fleet_a", OwnerID: "p1", RegionID: "sea", IsNaval: true},
			"fleet_b": {ID: "fleet_b", OwnerID: "p1", RegionID: "sea", IsNaval: true},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			regionAnchor: map[world.RegionID][2]int{"sea": {100, 100}},
		},
	}

	positions := r.armyIconPositions()
	if len(positions) != 2 {
		t.Fatalf("iki donanma ikonu bekleniyordu, got=%d", len(positions))
	}
	if got := positions[1].X - positions[0].X; got != 29 {
		t.Fatalf("yan yana donanma markerları arasında 3 px boşluk olmalıydı: delta=%.1f", got)
	}
}

func TestArmyIconPositionsPutNewArrivalLeftOfExistingArmy(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"target": {
				ID:      "target",
				OwnerID: "p1",
				Settlements: []world.Settlement{
					{ID: "target_center", IsCenter: true},
				},
			},
			"other": {
				ID:      "other",
				OwnerID: "p1",
				Settlements: []world.Settlement{
					{ID: "other_center", IsCenter: true},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			// ID sırası geliş sırasını temsil etmiyor; test özellikle bu
			// durumda konuma ilk gelen ordunun korunmasını doğrular.
			"army_zulu": {
				ID:       "army_zulu",
				OwnerID:  "p1",
				RegionID: "target",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
			"army_alpha": {
				ID:       "army_alpha",
				OwnerID:  "p1",
				RegionID: "other",
				Units:    []army.Unit{{TypeID: "inf"}},
			},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{
				{Region: "target", Index: 0}: {100, 100},
				{Region: "other", Index: 0}:  {200, 100},
			},
			primarySettlement: map[world.RegionID][2]int{
				"target": {100, 100},
				"other":  {200, 100},
			},
		},
	}

	// İlk frame'de target'taki ordu mevcut ordudur; diğer ordu başka
	// bölgededir ve henüz target grubuna dahil değildir.
	r.armyIconPositions()
	gs.Armies["army_alpha"].RegionID = "target"

	positions := r.armyIconPositions()
	if len(positions) != 2 {
		t.Fatalf("target grubunda iki ordu ikonu bekleniyordu, got=%d", len(positions))
	}
	if positions[0].ArmyID != "army_alpha" || positions[1].ArmyID != "army_zulu" {
		t.Fatalf("yeni gelen ordu mevcut ordunun solunda olmalıydı: %+v", positions)
	}
}

func TestArmyIconPositionsPutSiegeSupportLeftOfBesieger(t *testing.T) {
	gs := &state.GameState{
		Regions: map[world.RegionID]*world.Region{
			"fort": {
				ID:      "fort",
				OwnerID: "p2",
				Settlements: []world.Settlement{
					{ID: "castle", Type: world.SettlementFortress},
				},
			},
			"other": {
				ID:      "other",
				OwnerID: "p1",
				Settlements: []world.Settlement{
					{ID: "other_center", IsCenter: true},
				},
			},
		},
		Armies: map[army.ArmyID]*army.Army{
			"besieger": {
				ID:       "besieger",
				OwnerID:  "p1",
				RegionID: "fort",
				Units:    repeatedTestUnits(10),
			},
			"defender": {
				ID:       "defender",
				OwnerID:  "p2",
				RegionID: "fort",
				Units:    repeatedTestUnits(4),
			},
			"support": {
				ID:       "support",
				OwnerID:  "p1",
				RegionID: "other",
				Units:    repeatedTestUnits(1),
			},
		},
		Sieges: map[world.RegionID]*state.SiegeState{
			"fort": {
				RegionID:       "fort",
				AttackerArmyID: "besieger",
				DefenderArmyID: "defender",
			},
		},
	}
	r := &Renderer{
		gs: gs,
		worldMap: &WorldMap{
			settlementAnchor: map[settlementAnchorKey][2]int{
				{Region: "fort", Index: 0}:  {100, 100},
				{Region: "other", Index: 0}: {200, 100},
			},
			primarySettlement: map[world.RegionID][2]int{
				"fort":  {100, 100},
				"other": {200, 100},
			},
		},
	}

	// Kuşatan ve savunmacı daha önce aynı gruptaydı; destek daha sonra gelir.
	r.armyIconPositions()
	gs.Armies["support"].RegionID = "fort"

	positions := r.armyIconPositions()
	if len(positions) != 3 {
		t.Fatalf("üç ordu ikonu bekleniyordu, got=%d", len(positions))
	}
	if positions[0].ArmyID != "support" || positions[1].ArmyID != "besieger" || positions[2].ArmyID != "defender" {
		t.Fatalf("destek ordusu kuşatanın solunda, savunmacı sağında olmalıydı: %+v", positions)
	}
	if got := positions[1].X - positions[0].X; got != 26 {
		t.Fatalf("destek ile kuşatan bitişik slotlarda olmalıydı: delta=%.1f", got)
	}
}

func repeatedTestUnits(count int) []army.Unit {
	units := make([]army.Unit, count)
	for i := range units {
		units[i] = army.Unit{TypeID: "inf"}
	}
	return units
}
