package ai

import (
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func rapidExpansionTestState(fid faction.FactionID) *state.GameState {
	gs := &state.GameState{
		Turn:            1,
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			fid:      {ID: fid, NameTR: "Hızlı büyüyen devlet"},
		},
		Regions: make(map[world.RegionID]*world.Region),
	}
	for i := 0; i < 8; i++ {
		id := world.RegionID("fast_region_" + string(rune('a'+i)))
		gs.Regions[id] = &world.Region{ID: id, OwnerID: string(fid)}
	}
	return gs
}

func TestAIFactionRapidExpansionIncludesAIControlledFaction(t *testing.T) {
	gs := rapidExpansionTestState("ai_rival")
	for i := 0; i < aiRapidExpansionAbsoluteGain; i++ {
		gs.RecordRegionAcquisition("ai_rival", faction.FactionID("old_"+string(rune('a'+i))))
	}

	if !aiFactionRapidExpansion(gs, "ai_rival") {
		t.Fatal("hızlı toprak kazanan AI devleti koalisyon hedefi olarak algılanmadı")
	}
	if aiFactionRapidExpansion(gs, "player") {
		t.Fatal("büyüme kaydı olmayan oyuncu yanlışlıkla hızlı genişleyen hedef sayıldı")
	}
}

func TestAICoalitionWarCandidateRejectsAlliedRapidExpansionTarget(t *testing.T) {
	gs := rapidExpansionTestState("ai_rival")
	gs.Relations = map[string]*faction.Relation{
		faction.RelationKey("actor", "ai_rival"): {
			FactionA: "actor", FactionB: "ai_rival",
			Stance: faction.StanceAllied, Score: 80,
		},
	}
	gs.Factions["actor"] = &faction.Faction{ID: "actor", NameTR: "Aktör"}
	for i := 0; i < aiRapidExpansionAbsoluteGain; i++ {
		gs.RecordRegionAcquisition("ai_rival", faction.FactionID("old_"+string(rune('a'+i))))
	}

	if aiCoalitionWarCandidate(gs, "actor", "ai_rival", nil) {
		t.Fatal("müttefik ilişki hızlı büyüyen AI hedefine karşı koalisyon savaşı açtı")
	}
}

func TestAIOverextensionAdjustmentRespectsAggressiveness(t *testing.T) {
	gs := rapidExpansionTestState("actor")
	for i := 0; i < aiRapidExpansionAbsoluteGain; i++ {
		gs.RecordRegionAcquisition("actor", faction.FactionID("old_"+string(rune('a'+i))))
	}

	gs.Factions["actor"].AIAggressiveness = 25
	cautious := aiOverextensionWarScoreAdjustment(gs, "actor")
	gs.Factions["actor"].AIAggressiveness = 80
	aggressive := aiOverextensionWarScoreAdjustment(gs, "actor")
	if cautious >= aggressive {
		t.Fatalf("temkinli AI aşırı genişleme cezası = %d, agresif = %d", cautious, aggressive)
	}
}
