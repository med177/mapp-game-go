package diplomacy

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestIndependentWarSatisfactionPenaltiesUsesLandSeaAndRemoteWarCosts(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a":      {ID: "a"},
			"b":      {ID: "b"},
			"c":      {ID: "c"},
			"d":      {ID: "d"},
			"vassal": {ID: "vassal", OverlordID: "a"},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_land":     {ID: "a_land", OwnerID: "a", Neighbors: []world.RegionID{"b_land", "shared_sea"}},
			"b_land":     {ID: "b_land", OwnerID: "b", Neighbors: []world.RegionID{"a_land", "shared_sea"}},
			"c_land":     {ID: "c_land", OwnerID: "c"},
			"d_land":     {ID: "d_land", OwnerID: "d", Neighbors: []world.RegionID{"shared_sea"}},
			"shared_sea": {ID: "shared_sea", IsSea: true},
		},
		Relations: map[string]*faction.Relation{
			"a|b": {FactionA: "a", FactionB: "b", Stance: faction.StanceWar},
			"a|c": {FactionA: "a", FactionB: "c", Stance: faction.StanceWar},
			"a|d": {FactionA: "a", FactionB: "d", Stance: faction.StanceWar},
		},
	}

	penalties := IndependentWarSatisfactionPenalties(gs)
	if got, want := penalties["a"], 6; got != want {
		t.Fatalf("toplu savaş yorgunluğu: got=%d want=%d", got, want)
	}
	if got, want := penalties["vassal"], 6; got != want {
		t.Fatalf("vassal realm savaş yorgunluğu: got=%d want=%d", got, want)
	}
	if got, want := IndependentWarSatisfactionPenalty(gs, "a"), 6; got != want {
		t.Fatalf("tekil helper savaş yorgunluğu: got=%d want=%d", got, want)
	}
	for fid := range gs.Factions {
		if got, want := penalties[fid], legacyIndependentWarSatisfactionPenalty(gs, fid); got != want {
			t.Fatalf("toplu/tekil savaş yorgunluğu ayrıştı fid=%s: got=%d want=%d", fid, got, want)
		}
	}
}

func legacyIndependentWarSatisfactionPenalty(gs *state.GameState, fid faction.FactionID) int {
	realm := RealmRoot(gs, fid)
	if realm == "" {
		realm = fid
	}
	penalty := 0
	for opponent := range independentWarRealms(gs, realm) {
		switch {
		case realmsShareLandBorder(gs, realm, opponent):
			penalty += WarFatigueSatisfactionPenalty
		case realmsShareSeaBorder(gs, realm, opponent):
			penalty += warFatigueSeaBorderPenalty
		default:
			penalty += warFatigueRemotePenalty
		}
	}
	return penalty
}

func TestIndependentWarSatisfactionPenaltiesIgnoresInternalRealmWar(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"overlord": {ID: "overlord"},
			"vassal":   {ID: "vassal", OverlordID: "overlord"},
		},
		Relations: map[string]*faction.Relation{
			"overlord|vassal": {FactionA: "overlord", FactionB: "vassal", Stance: faction.StanceWar},
		},
	}

	penalties := IndependentWarSatisfactionPenalties(gs)
	if got := penalties["overlord"]; got != 0 {
		t.Fatalf("aynı realm içindeki savaş cezalandırıldı: got=%d", got)
	}
	if got := IndependentWarSatisfactionPenalty(gs, "vassal"); got != 0 {
		t.Fatalf("aynı realm vassal savaş cezası aldı: got=%d", got)
	}
}

func TestSharedMajorThreatsMatchesPairwiseHelper(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"a": {ID: "a"},
			"b": {ID: "b"},
			"c": {ID: "c"},
			"d": {ID: "d"},
		},
		Regions: map[world.RegionID]*world.Region{
			"a_land": {ID: "a_land", OwnerID: "a", Neighbors: []world.RegionID{"c_land"}},
			"b_land": {ID: "b_land", OwnerID: "b", Neighbors: []world.RegionID{"c_land"}},
			"c_1":    {ID: "c_1", OwnerID: "c"},
			"c_2":    {ID: "c_2", OwnerID: "c"},
			"c_3":    {ID: "c_3", OwnerID: "c"},
			"c_land": {ID: "c_land", OwnerID: "c", Neighbors: []world.RegionID{"a_land", "b_land"}},
			"d_land": {ID: "d_land", OwnerID: "d"},
		},
	}

	all := SharedMajorThreats(gs, "a")
	for targetID := range gs.Factions {
		if targetID == "a" {
			continue
		}
		if got, want := all[targetID], HasSharedMajorThreat(gs, "a", targetID); got != want {
			t.Fatalf("paylaşılan tehdit sonucu ayrıştı hedef=%s: got=%t want=%t", targetID, got, want)
		}
	}
	if !all["b"] {
		t.Fatal("c fraksiyonu a ve b için ortak büyük tehdit olarak bulunamadı")
	}
	if all["d"] {
		t.Fatal("sınırı olmayan d fraksiyonu ortak büyük tehdit sayıldı")
	}
}
