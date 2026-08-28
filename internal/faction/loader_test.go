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

func TestBuildInitialRelationsDoesNotDeclareWarForEliminatedFaction(t *testing.T) {
	factions := map[FactionID]*Faction{
		"active_sunni": {ID: "active_sunni", Religion: religion.Sunni},
		"active_shia":  {ID: "active_shia", Religion: religion.Shia},
		"dead_shia":    {ID: "dead_shia", Religion: religion.Shia, IsEliminated: true},
	}

	relations := BuildInitialRelations(factions)
	if got := relations[RelationKey("active_sunni", "active_shia")].Stance; got != StanceWar {
		t.Fatalf("aktif Sünni-Şii çifti savaşta başlamalıydı: %s", got)
	}
	if got := relations[RelationKey("active_sunni", "dead_shia")].Stance; got != StancePeace {
		t.Fatalf("elenmiş Şii devlet savaş ilişkisine girmemeliydi: %s", got)
	}
}
