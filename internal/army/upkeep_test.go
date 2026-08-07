package army

import "testing"

func TestTotalGoldUpkeepIncludesLandAndEmbarkedUnits(t *testing.T) {
	a := &Army{
		Units:         []Unit{{TypeID: "infantry"}, {TypeID: "cavalry"}},
		EmbarkedUnits: []Unit{{TypeID: "militia"}},
	}
	types := map[string]*UnitType{
		"militia":  {ID: "militia", GoldUpkeep: 1},
		"infantry": {ID: "infantry", GoldUpkeep: 2},
		"cavalry":  {ID: "cavalry", GoldUpkeep: 3},
	}

	if got, want := a.TotalGoldUpkeep(types), 6; got != want {
		t.Fatalf("altın bakımı hatalı: got=%d want=%d", got, want)
	}
}
