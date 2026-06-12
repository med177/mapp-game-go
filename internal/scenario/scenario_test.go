package scenario

import "testing"

func TestVictoryOptionRegionTargetsPrefersRequiredRegions(t *testing.T) {
	opt := VictoryOptionDef{
		RequiredRegions: []string{"constantinople", "ankara"},
	}

	got := opt.RegionTargets()
	if len(got) != 2 || got[0] != "constantinople" || got[1] != "ankara" {
		t.Fatalf("required_regions bekleniyordu, got=%v", got)
	}
}

func TestVictoryOptionRegionTargetsSkipsEmptyValues(t *testing.T) {
	opt := VictoryOptionDef{
		RequiredRegions: []string{"paris", "", "flanders"},
	}

	got := opt.RegionTargets()
	if len(got) != 2 || got[0] != "paris" || got[1] != "flanders" {
		t.Fatalf("bos degerler filtrelenmeliydi, got=%v", got)
	}
}

func TestFilterVictoryOptionsForFaction(t *testing.T) {
	options := []VictoryOptionDef{
		{ID: "shared"},
		{ID: "ottoman_only", AllowedFactions: []string{"ottoman"}},
		{ID: "rome_only", AllowedFactions: []string{"east_rome"}},
	}

	got := FilterVictoryOptionsForFaction(options, "ottoman")
	if len(got) != 2 {
		t.Fatalf("2 sonuc bekleniyordu, got=%d", len(got))
	}
	if got[0].ID != "shared" || got[1].ID != "ottoman_only" {
		t.Fatalf("beklenmeyen filtre sonucu: %+v", got)
	}
}
