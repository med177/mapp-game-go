package army

import "testing"

func TestAirCarrierTransportCapacity(t *testing.T) {
	types := map[string]*UnitType{
		"dragon": {
			ID:            "dragon",
			Category:      CategoryDragon,
			CarryCapacity: 1,
		},
		"soldier": {
			ID:       "soldier",
			Category: CategoryInfantry,
		},
	}
	carrier := &Army{
		Units: []Unit{{TypeID: "dragon"}},
	}

	if got := carrier.TransportCapacity(types); got != 1 {
		t.Fatalf("air carrier capacity = %d, want 1", got)
	}
	if !carrier.CanEmbarkUnits(types, 1) {
		t.Fatal("air carrier should accept one unit")
	}
	if carrier.CanEmbarkUnits(types, 2) {
		t.Fatal("air carrier should reject units beyond capacity")
	}

	landArmy := &Army{Units: []Unit{{TypeID: "soldier"}}}
	if got := landArmy.TransportCapacity(types); got != 0 {
		t.Fatalf("land army capacity = %d, want 0", got)
	}

	types["wagon"] = &UnitType{ID: "wagon", Category: CategoryInfantry, CarryCapacity: 2}
	landCarrier := &Army{Units: []Unit{{TypeID: "wagon"}}}
	if !landCarrier.IsTransportCarrier(types) || landCarrier.TransportCapacity(types) != 2 {
		t.Fatal("transport capacity should work for every unit category")
	}
}
