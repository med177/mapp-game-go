package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestFactionPrimaryRegionUsesConfiguredFactionCapital(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"owner": {ID: "owner", CapitalSettlementID: "capital_city"},
		},
		Regions: map[world.RegionID]*world.Region{
			"capital": {
				ID: "capital", OwnerID: "owner", TradeCapacity: 1,
				Settlements: []world.Settlement{{ID: "capital_city", IsCenter: true}},
			},
			"richer_region": {
				ID: "richer_region", OwnerID: "owner", TradeCapacity: 12,
				Settlements: []world.Settlement{{ID: "richer_city", IsCenter: true}},
			},
		},
	}
	r := &Renderer{gs: gs}

	if got := r.factionPrimaryRegion("owner"); got == nil || got.ID != "capital" {
		t.Fatalf("ticaret bağlantısı ulusal başkenti kullanmalı: got=%+v", got)
	}
}

func TestTradeCenterIconLoadsAtNativeResolution(t *testing.T) {
	loadTradeCenterIcon()
	if tradeCenterIcon == nil {
		t.Fatal("primary ticaret merkezi ikonu yüklenemedi")
	}
	bounds := tradeCenterIcon.Bounds()
	if bounds.Dx() != 96 || bounds.Dy() != 96 {
		t.Fatalf("ticaret merkezi ikonu 96x96 olmalı: %dx%d", bounds.Dx(), bounds.Dy())
	}
}
