package army

import "testing"

func TestArmyBaseMovePointsUsesSlowestUnit(t *testing.T) {
	types := map[string]*UnitType{
		"inf":       {ID: "inf", Category: CategoryInfantry, MovementPoints: 2},
		"cav":       {ID: "cav", Category: CategoryCavalry, MovementPoints: 3},
		"siege":     {ID: "siege", Category: CategorySiege, MovementPoints: 1},
		"transport": {ID: "transport", Category: CategoryNavalTrans, MovementPoints: 3},
	}

	cases := []struct {
		name string
		army *Army
		want int
	}{
		{name: "süvari", army: &Army{Units: []Unit{{TypeID: "cav"}}}, want: 3},
		{name: "piyade", army: &Army{Units: []Unit{{TypeID: "inf"}}}, want: 2},
		{name: "kuşatma", army: &Army{Units: []Unit{{TypeID: "siege"}}}, want: 1},
		{name: "karışık ordu", army: &Army{Units: []Unit{{TypeID: "cav"}, {TypeID: "inf"}, {TypeID: "siege"}}}, want: 1},
		{name: "filo taşınan birlikten etkilenmez", army: &Army{IsNaval: true, Units: []Unit{{TypeID: "transport"}}, EmbarkedUnits: []Unit{{TypeID: "siege"}}}, want: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.army.BaseMovePoints(types); got != tc.want {
				t.Fatalf("taban hareket puanı yanlış: got=%d want=%d", got, tc.want)
			}
		})
	}
}
