package world

import (
	"path/filepath"
	"testing"
)

func TestScenarioSeaAdjacency_MarmaraBridgesAegeanAndBlackSea(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		regionPath string
		marmaraID  RegionID
		aegeanID   RegionID
		blackSeaID RegionID
	}{
		{
			name:       "1300 ottoman rise",
			regionPath: filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "regions.json"),
			marmaraID:  "sea_of_marmara",
			aegeanID:   "aegean_sea",
			blackSeaID: "black_sea",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			regions, err := LoadRegions(tc.regionPath)
			if err != nil {
				t.Fatalf("regions yuklenemedi: %v", err)
			}

			assertNeighborBothWays(t, regions, tc.marmaraID, tc.aegeanID)
			assertNeighborBothWays(t, regions, tc.marmaraID, tc.blackSeaID)
		})
	}
}

func assertNeighborBothWays(t *testing.T, regions map[RegionID]*Region, a, b RegionID) {
	t.Helper()

	ra := regions[a]
	if ra == nil {
		t.Fatalf("region bulunamadi: %s", a)
	}
	rb := regions[b]
	if rb == nil {
		t.Fatalf("region bulunamadi: %s", b)
	}

	if !hasNeighbor(ra, b) {
		t.Fatalf("%s komsularinda %s olmali", a, b)
	}
	if !hasNeighbor(rb, a) {
		t.Fatalf("%s komsularinda %s olmali", b, a)
	}
}

func hasNeighbor(region *Region, target RegionID) bool {
	for _, neighbor := range region.Neighbors {
		if neighbor == target {
			return true
		}
	}
	return false
}
