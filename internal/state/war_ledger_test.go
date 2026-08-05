package state

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

func TestWarLedgerTracksBothPerspectivesAndResetsPlanOnPeace(t *testing.T) {
	gs := &GameState{
		Turn: 5,
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a"},
			"a2": {ID: "a2", OwnerID: "a"},
			"b1": {ID: "b1", OwnerID: "b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
		},
		AIPlans: map[faction.FactionID]*AIPlanState{
			"a": {TargetFactionID: "b", RallyRegionID: "a1", RallyDeadlineTurn: 9, ReassessTurn: 12},
		},
	}

	ledger := gs.BeginWarLedger("b", "a")
	if ledger == nil || ledger.FactionA != "a" || ledger.FactionB != "b" || ledger.DeclarerFactionID != "b" || ledger.DefenderFactionID != "a" {
		t.Fatalf("ledger taraf sırası hatalı: %+v", ledger)
	}
	if ledger.StartedTurn != 5 || ledger.InitialRegionsA != 2 || ledger.InitialRegionsB != 1 {
		t.Fatalf("başlangıç snapshot'ı hatalı: %+v", ledger)
	}

	gs.RecordWarCasualties("b", "a", 2, 4)
	gs.RecordWarRegionCapture("b", "a")
	if ledger.CasualtiesA != 4 || ledger.CasualtiesB != 2 || ledger.RegionsCapturedB != 1 || ledger.LastBattleTurn != 5 {
		t.Fatalf("savaş sayaçları hatalı: %+v", ledger)
	}

	gs.EndWarLedger("a", "b")
	if gs.WarLedgerFor("a", "b") != nil {
		t.Fatal("barış sonrası aktif ledger kaldırılmadı")
	}
	plan := gs.AIPlans["a"]
	if plan.ReassessTurn != 5 || plan.RallyRegionID != "" || plan.RallyDeadlineTurn != 0 {
		t.Fatalf("barış sonrası plan/rally sıfırlanmadı: %+v", plan)
	}
}

func TestSyncWarLedgersMigratesActiveLegacyWarAtCurrentTurn(t *testing.T) {
	gs := &GameState{
		Turn: 17,
		Regions: map[world.RegionID]*world.Region{
			"a1": {ID: "a1", OwnerID: "a"},
			"b1": {ID: "b1", OwnerID: "b"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("a", "b"): {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
		},
	}

	gs.SyncWarLedgers()
	ledger := gs.WarLedgerFor("a", "b")
	if ledger == nil || ledger.StartedTurn != 17 {
		t.Fatalf("eski aktif savaş yükleme turunda başlatılmadı: %+v", ledger)
	}
}
