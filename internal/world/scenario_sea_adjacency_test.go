package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScenarioSeaAdjacency_MarmaraBridgesAegeanAndBlackSea(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		regionPath string
		bridgePath []RegionID
	}{
		{
			name:       "1300 ottoman rise",
			regionPath: filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "data", "regions.json"),
			bridgePath: []RegionID{"aegean_sea", "new_region_238", "new_region_255", "black_sea"},
		},
		{
			name:       "1444 ottoman empire",
			regionPath: filepath.Join("..", "..", "assets", "scenarios", "1444_ottoman_empire", "data", "regions.json"),
			bridgePath: []RegionID{"aegean_sea", "sea_of_marmara", "black_sea"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			skipIfScenarioDirMissing(t, filepath.Dir(filepath.Dir(tc.regionPath)))

			regions, err := LoadRegions(tc.regionPath)
			if err != nil {
				t.Fatalf("regions yuklenemedi: %v", err)
			}

			for i := 0; i < len(tc.bridgePath)-1; i++ {
				assertNeighborBothWays(t, regions, tc.bridgePath[i], tc.bridgePath[i+1])
			}
		})
	}
}

func skipIfScenarioDirMissing(t *testing.T, scenarioPath string) {
	t.Helper()

	info, err := os.Stat(scenarioPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("senaryo dizini mevcut değil: %s", scenarioPath)
		}
		t.Fatalf("senaryo dizini kontrol edilemedi: %v", err)
	}
	if !info.IsDir() {
		t.Skipf("senaryo yolu dizin değil: %s", scenarioPath)
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
