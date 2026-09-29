package state

import (
	"testing"

	"mapp-game-go/internal/faction"
)

func TestMergePoliticalRelationsConsolidatesExternalRelations(t *testing.T) {
	resultID := faction.FactionID("result")
	sourceOneID := faction.FactionID("source_one")
	sourceTwoID := faction.FactionID("source_two")
	externalID := faction.FactionID("external")
	resultExternalKey := faction.RelationKey(resultID, externalID)
	sourceOneExternalKey := faction.RelationKey(sourceOneID, externalID)
	sourceTwoExternalKey := faction.RelationKey(sourceTwoID, externalID)

	gs := &GameState{
		Relations: map[string]*faction.Relation{
			resultExternalKey: {
				FactionA: resultID,
				FactionB: externalID,
				Score:    25,
				Stance:   faction.StancePeace,
			},
			sourceOneExternalKey: {
				FactionA: sourceOneID,
				FactionB: externalID,
				Score:    65,
				Stance:   faction.StanceAllied,
			},
			sourceTwoExternalKey: {
				FactionA: sourceTwoID,
				FactionB: externalID,
				Score:    90,
				Stance:   faction.StanceTrade,
			},
			faction.RelationKey(sourceOneID, sourceTwoID): {
				FactionA: sourceOneID,
				FactionB: sourceTwoID,
				Score:    80,
				Stance:   faction.StanceAllied,
			},
		},
		RelationOrder: []string{
			sourceOneExternalKey,
			resultExternalKey,
			sourceTwoExternalKey,
			faction.RelationKey(sourceOneID, sourceTwoID),
		},
	}

	gs.MergePoliticalRelations([]faction.FactionID{resultID, sourceOneID, sourceTwoID}, resultID)

	merged := gs.Relations[resultExternalKey]
	if merged == nil {
		t.Fatal("sonuç faction'ı için dış relation oluşturulmadı")
	}
	if merged.Stance != faction.StanceAllied || merged.Score != 65 {
		t.Fatalf("birleşik relation = %s/%d, want allied/65", merged.Stance, merged.Score)
	}
	if _, exists := gs.Relations[sourceOneExternalKey]; exists {
		t.Fatal("ilk kaynak faction relation'ı temizlenmedi")
	}
	if _, exists := gs.Relations[sourceTwoExternalKey]; exists {
		t.Fatal("ikinci kaynak faction relation'ı temizlenmedi")
	}
	if _, exists := gs.Relations[faction.RelationKey(sourceOneID, sourceTwoID)]; exists {
		t.Fatal("birleşen faction'lar arasındaki relation temizlenmedi")
	}
	if len(gs.RelationOrder) != 1 || gs.RelationOrder[0] != resultExternalKey {
		t.Fatalf("relation sırası = %#v, want [%s]", gs.RelationOrder, resultExternalKey)
	}
}
