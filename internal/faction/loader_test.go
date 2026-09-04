package faction

import (
	"testing"

	"mapp-game-go/internal/religion"
)

func TestApplyHistoricalChangeUsesLatestChangeAtOrBeforeYear(t *testing.T) {
	f := &Faction{
		Name:   "Beylik",
		NameTR: "Beylik",
		HistoricalChanges: []HistoricalChange{
			{Year: 1350, Flag: "state.png", Name: "State", NameTR: "Devlet"},
			{Year: 1460, Flag: "empire.png", Name: "Empire", NameTR: "İmparatorluk"},
		},
	}

	if f.ApplyHistoricalChange(1349) {
		t.Fatal("erken yılda tarihsel değişiklik uygulanmamalı")
	}
	if !f.ApplyHistoricalChange(1350) || f.NameTR != "Devlet" || f.Flag != "state.png" {
		t.Fatalf("1350 değişikliği uygulanmadı: %+v", f)
	}
	if !f.ApplyHistoricalChange(1460) || f.NameTR != "İmparatorluk" || f.Flag != "empire.png" {
		t.Fatalf("1460 değişikliği uygulanmadı: %+v", f)
	}
}

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
