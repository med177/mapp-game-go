package render

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestDiplomacyCacheReusesAndInvalidatesFactionOrder(t *testing.T) {
	gs := &state.GameState{
		Turn:            1,
		Year:            1300,
		Month:           1,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {NameTR: "Oyuncu"},
			"alpha":  {NameTR: "Alpha"},
			"beta":   {NameTR: "Beta"},
		},
	}
	r := &Renderer{gs: gs}

	first := r.cachedDiplomacyFactions(diplomacyListSortAlphabetical)
	second := r.cachedDiplomacyFactions(diplomacyListSortAlphabetical)
	if len(first) != 3 || len(second) != 3 {
		t.Fatalf("cache fraksiyon listesini korumadı: first=%v second=%v", first, second)
	}
	if &first[0] != &second[0] {
		t.Fatal("aynı state için sıralı diplomasi listesi yeniden oluşturuldu")
	}

	gs.Factions["beta"].IsEliminated = true
	r.invalidateDiplomacyCache()
	updated := r.cachedDiplomacyFactions(diplomacyListSortAlphabetical)
	if len(updated) != 2 {
		t.Fatalf("cache invalidasyonundan sonra elenmiş devlet listede kaldı: %v", updated)
	}
}
