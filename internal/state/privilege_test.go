package state

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestPrivilegedMinorUsesParentOwnerAndSplitsIncome(t *testing.T) {
	parent := &world.Region{ID: "constantinople", OwnerID: "east_rome"}
	minor := &world.Region{
		ID: "galata", OwnerID: "genoa", IsMinorRegion: true,
		IsPrivileged: true, ParentRegionID: parent.ID,
	}
	gs := &GameState{Regions: map[world.RegionID]*world.Region{
		parent.ID: parent,
		minor.ID:  minor,
	}}

	if got := gs.SovereignOwnerID(minor); got != "east_rome" {
		t.Fatalf("imtiyazlı alt bölgenin egemen sahibi: got=%q want=east_rome", got)
	}
	sovereign, operator := gs.RegionIncomeShares(minor, 100)
	if sovereign != 50 || operator != 50 {
		t.Fatalf("imtiyaz geliri eşit paylaşılmadı: sovereign=%d operator=%d", sovereign, operator)
	}
	if got := len(gs.LandRegionsOwnedBy("east_rome")); got != 2 {
		t.Fatalf("ana devlet imtiyazlı alt bölgeyi toprağı saymadı: got=%d", got)
	}
	if got := len(gs.LandRegionsOwnedBy("genoa")); got != 0 {
		t.Fatalf("kullanım sahibi imtiyazlı alt bölgeyi egemen toprağı saydı: got=%d", got)
	}
}

func TestNonPrivilegedMinorRemainsOwnedByOwnerID(t *testing.T) {
	parent := &world.Region{ID: "constantinople", OwnerID: "east_rome"}
	minor := &world.Region{
		ID: "galata", OwnerID: "genoa", IsMinorRegion: true,
		ParentRegionID: parent.ID,
	}
	gs := &GameState{Regions: map[world.RegionID]*world.Region{
		parent.ID: parent,
		minor.ID:  minor,
	}}

	if got := gs.SovereignOwnerID(minor); got != "genoa" {
		t.Fatalf("imtiyazsız alt bölgenin sahibi ana bölgeye taşındı: got=%q", got)
	}
	sovereign, operator := gs.RegionIncomeShares(minor, 100)
	if sovereign != 100 || operator != 0 {
		t.Fatalf("imtiyazsız alt bölge geliri bölündü: sovereign=%d operator=%d", sovereign, operator)
	}
}
