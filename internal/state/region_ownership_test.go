package state

import (
	"testing"

	"mapp-game-go/internal/world"
)

func TestLandRegionsOwnedByIgnoresTerrainAreaFragments(t *testing.T) {
	base := &world.Region{ID: "lazio", OwnerID: "papal_states"}
	terrain := &world.Region{
		ID: "area::alps::lazio", OwnerID: "papal_states", IsTerrainArea: true,
	}
	sea := &world.Region{ID: "tyrrhenian_sea", OwnerID: "papal_states", IsSea: true}
	gs := &GameState{Regions: map[world.RegionID]*world.Region{
		base.ID:    base,
		terrain.ID: terrain,
		sea.ID:     sea,
	}}

	owned := gs.LandRegionsOwnedBy("papal_states")
	if len(owned) != 1 || owned[0] != base {
		t.Fatalf("terrain-area veya deniz bölgesi stratejik kara sayıldı: got=%v", owned)
	}
}

func TestIsEliminatedWhenOnlyTerrainAreaRemains(t *testing.T) {
	terrain := &world.Region{
		ID: "area::alps::lazio", OwnerID: "papal_states", IsTerrainArea: true,
	}
	gs := &GameState{Regions: map[world.RegionID]*world.Region{terrain.ID: terrain}}

	if !gs.IsEliminated("papal_states") {
		t.Fatal("yalnız terrain-area parçası kalan devlet elenmiş sayılmadı")
	}
}
