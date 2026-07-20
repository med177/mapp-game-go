package faction

import (
	"testing"

	"mapp-game-go/internal/religion"
)

func TestBuildInitialRelationsUsesCanonicalFactionOrder(t *testing.T) {
	factions := map[FactionID]*Faction{
		"flanders_county": {ID: "flanders_county", Religion: religion.Catholic},
		"hre":             {ID: "hre", Religion: religion.Catholic},
		"ottoman":         {ID: "ottoman", Religion: religion.Sunni},
		"east_rome":       {ID: "east_rome", Religion: religion.Orthodox},
	}

	for key, relation := range BuildInitialRelations(factions) {
		if relation == nil || relation.FactionA > relation.FactionB {
			t.Fatalf("ilişki yönü kanonik değil: key=%s relation=%+v", key, relation)
		}
	}
}
