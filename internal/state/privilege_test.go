package state

import (
	"testing"

	"mapp-game-go/internal/economy"
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
	if got := len(gs.LandRegionsVisibleTo("east_rome")); got != 2 {
		t.Fatalf("egemen devlet imtiyazlı alt bölgeyi görünür saymadı: got=%d", got)
	}
	if got := len(gs.LandRegionsVisibleTo("genoa")); got != 1 {
		t.Fatalf("kullanım sahibi imtiyazlı alt bölgeyi görünür saymadı: got=%d", got)
	}
	gs.PlayerFactionID = "east_rome"
	if !gs.CanRevokeMinorPrivilege(minor.ID) {
		t.Fatalf("egemen devlet imtiyazı kaldırabilir görünmüyor: %s", gs.MinorPrivilegeRevokeBlockReason(minor.ID))
	}
	gs.PlayerFactionID = "genoa"
	if gs.CanRevokeMinorPrivilege(minor.ID) {
		t.Fatal("kullanım sahibi imtiyaz kaldırma yetkisine sahip görünüyor")
	}
}

func TestPrivilegedMinorFollowsParentOwnerChangeForIncomeAndRevoke(t *testing.T) {
	parent := &world.Region{ID: "constantinople", OwnerID: "east_rome"}
	minor := &world.Region{
		ID: "galata", OwnerID: "genoa", IsMinorRegion: true,
		IsPrivileged: true, ParentRegionID: parent.ID,
	}
	gs := &GameState{
		PlayerFactionID: "venice",
		Regions: map[world.RegionID]*world.Region{
			parent.ID: parent,
			minor.ID:  minor,
		},
	}

	parent.OwnerID = "venice"
	if got := gs.SovereignOwnerID(minor); got != "venice" {
		t.Fatalf("üst bölge el değiştirince minor egemen sahibi = %q, venice bekleniyordu", got)
	}
	sovereign, operator := gs.RegionIncomeShares(minor, 100)
	if sovereign != 50 || operator != 50 {
		t.Fatalf("yeni egemenin minor geliri = (%d, %d), (50, 50) bekleniyordu", sovereign, operator)
	}
	if !gs.CanRevokeMinorPrivilege(minor.ID) {
		t.Fatalf("üst bölgeyi alan yeni devlet imtiyazı kaldırabilir görünmüyor: %s", gs.MinorPrivilegeRevokeBlockReason(minor.ID))
	}
	gs.PlayerFactionID = "east_rome"
	if gs.CanRevokeMinorPrivilege(minor.ID) {
		t.Fatal("eski üst bölge sahibi imtiyazı kaldırabilir görünmeye devam ediyor")
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

func TestPrivilegedTradeParticipantsAndValue(t *testing.T) {
	parent := &world.Region{ID: "galata_parent", OwnerID: "grantor"}
	minor := &world.Region{
		ID: "galata", OwnerID: "operator", IsMinorRegion: true,
		IsPrivileged: true, ParentRegionID: parent.ID,
	}
	gs := &GameState{
		Regions:        map[world.RegionID]*world.Region{parent.ID: parent, minor.ID: minor},
		BaseGoldValues: map[economy.GoodType]int{economy.GoodCloth: 8},
	}
	route := &economy.TradeRoute{
		FromFactionID: "grantor", ToFactionID: "operator",
		Good: economy.GoodCloth, AmountPerTurn: 4, IsPrivilegedMinor: true,
	}

	sovereign, operator, ok := gs.PrivilegedTradeParticipants(route.FromFactionID, route.ToFactionID)
	if !ok || sovereign != "grantor" || operator != "operator" {
		t.Fatalf("imtiyazlı rota tarafları = (%q, %q, %t), (grantor, operator, true) bekleniyordu", sovereign, operator, ok)
	}
	if got := gs.PrivilegedTradeRouteValue(route, 4); got != 32 {
		t.Fatalf("imtiyazlı rota değeri = %d, 32 bekleniyordu", got)
	}
	if got := PrivilegedTradeRoyalty(32); got != 8 {
		t.Fatalf("imtiyaz telifi = %d, 8 bekleniyordu", got)
	}
}
