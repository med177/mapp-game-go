package state

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func newWarLedgerTestState() *GameState {
	return &GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a", NameTR: "A"},
			"b": {ID: "b", NameTR: "B"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {
				FactionA: "a",
				FactionB: "b",
				Stance:   faction.StanceWar,
			},
		},
	}
}

func TestRecordWarCasualtiesByTypeKeepsSortedSideAndTypeTotals(t *testing.T) {
	gs := newWarLedgerTestState()
	gs.Turn = 12
	gs.RecordWarCasualtiesByType("a", "b", 3, 1, false, true)

	ledger := gs.WarLedgerFor("a", "b")
	if ledger == nil {
		t.Fatal("savaş ledger'ı oluşturulmadı")
	}
	if ledger.CasualtiesA != 3 || ledger.CasualtiesB != 1 {
		t.Fatalf("toplam kayıplar = %d/%d, want 3/1", ledger.CasualtiesA, ledger.CasualtiesB)
	}
	if ledger.CasualtiesArmyA != 3 || ledger.CasualtiesFleetA != 0 || ledger.CasualtiesArmyB != 0 || ledger.CasualtiesFleetB != 1 {
		t.Fatalf("tür kayıpları = ordu %d/%d filo %d/%d, want ordu 3/0 filo 0/1", ledger.CasualtiesArmyA, ledger.CasualtiesArmyB, ledger.CasualtiesFleetA, ledger.CasualtiesFleetB)
	}
	if ledger.LastBattleTurn != 12 {
		t.Fatalf("son muharebe turu = %d, want 12", ledger.LastBattleTurn)
	}

	gs = newWarLedgerTestState()
	gs.Turn = 13
	gs.RecordWarCasualtiesByType("b", "a", 4, 2, true, false)
	ledger = gs.WarLedgerFor("a", "b")
	if ledger.CasualtiesA != 2 || ledger.CasualtiesB != 4 || ledger.CasualtiesArmyA != 2 || ledger.CasualtiesFleetB != 4 {
		t.Fatalf("ters saldırı tarafları yanlış eşlendi: toplam %d/%d orduA %d filoB %d", ledger.CasualtiesA, ledger.CasualtiesB, ledger.CasualtiesArmyA, ledger.CasualtiesFleetB)
	}
}

func TestRecordWarCasualtiesLegacyDefaultsToArmyLosses(t *testing.T) {
	gs := newWarLedgerTestState()
	gs.RecordWarCasualties("a", "b", 2, 1)

	ledger := gs.WarLedgerFor("a", "b")
	if ledger == nil {
		t.Fatal("savaş ledger'ı oluşturulmadı")
	}
	if ledger.CasualtiesArmyA != 2 || ledger.CasualtiesArmyB != 1 || ledger.CasualtiesFleetA != 0 || ledger.CasualtiesFleetB != 0 {
		t.Fatalf("legacy kayıp dağılımı = ordu %d/%d filo %d/%d", ledger.CasualtiesArmyA, ledger.CasualtiesArmyB, ledger.CasualtiesFleetA, ledger.CasualtiesFleetB)
	}
}

func TestRecentRegionGainTracksShortExpansionWindow(t *testing.T) {
	gs := &GameState{Turn: 1}
	gs.RecordRegionAcquisition("fast", "old_a")
	gs.RecordRegionAcquisition("fast", "old_b")

	if got := gs.RecentRegionGain("fast"); got != 2 {
		t.Fatalf("yakın dönem bölge kazanımı = %d, want 2", got)
	}

	gs.Turn = 1 + RecentFactionExpansionWindowTurns
	if got := gs.RecentRegionGain("fast"); got != 0 {
		t.Fatalf("süresi dolmuş genişleme kaydı = %d, want 0", got)
	}

	gs.RecordRegionAcquisition("fast", "old_c")
	if got := gs.RecentRegionGain("fast"); got != 1 {
		t.Fatalf("yeni genişleme penceresi kazanımı = %d, want 1", got)
	}
}

func TestOverextensionScoreUsesRecentAndRelativeExpansion(t *testing.T) {
	gs := &GameState{
		Turn:    1,
		Regions: make(map[world.RegionID]*world.Region),
	}
	for i := 0; i < 10; i++ {
		id := world.RegionID("owned_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: "fast"}
	}
	for i := 0; i < 3; i++ {
		gs.RecordRegionAcquisition("fast", faction.FactionID("old_"+string(rune('a'+i))))
	}

	if got, want := gs.OverextensionScore("fast"), 60; got != want {
		t.Fatalf("aşırı genişleme değeri = %d, want %d", got, want)
	}

	gs.Turn = 1 + RecentFactionExpansionWindowTurns
	if got := gs.OverextensionScore("fast"); got != 0 {
		t.Fatalf("süresi dolan genişleme sonrası aşırı genişleme = %d, want 0", got)
	}
}
