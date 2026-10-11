package save

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

func TestDiplomacyDisabledChangePersistsAcrossSaveFormats(t *testing.T) {
	fid := faction.FactionID("others")
	base := &faction.Faction{ID: fid, DiplomacyDisabled: true}
	current := &faction.Faction{ID: fid, DiplomacyDisabled: false}

	delta, ok := makeFactionSaveState(current, base)
	if !ok || delta.DiplomacyDisabled == nil || *delta.DiplomacyDisabled {
		t.Fatalf("compact save false değişikliğini saklamadı: %#v", delta.DiplomacyDisabled)
	}
	data, err := json.Marshal(delta)
	if err != nil {
		t.Fatalf("compact faction delta serileştirilemedi: %v", err)
	}
	var decoded factionSaveState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("compact faction delta çözülemedi: %v", err)
	}
	restored := &faction.Faction{ID: fid, DiplomacyDisabled: true}
	applyFactionSaveState(restored, decoded)
	if restored.DiplomacyDisabled {
		t.Fatal("compact save yüklenince diplomacy_disabled kapalı kaldı")
	}

	debugState := makeDebugCampaignSaveState(&state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{fid: current},
	})
	debugFaction := debugState.Factions[fid]
	if debugFaction.DiplomacyDisabled == nil || *debugFaction.DiplomacyDisabled {
		t.Fatalf("debug save false değişikliğini saklamadı: %#v", debugFaction.DiplomacyDisabled)
	}
	legacyConverted := convertLegacyCampaignSaveState(debugState)
	legacyDelta := legacyConverted.Factions[fid]
	if legacyDelta.DiplomacyDisabled == nil || *legacyDelta.DiplomacyDisabled {
		t.Fatalf("debug/legacy dönüşümü false değerini kaybetti: %#v", legacyDelta.DiplomacyDisabled)
	}
}

func TestLegacySaveWithoutDiplomacyDisabledPreservesScenarioDefault(t *testing.T) {
	fid := faction.FactionID("others")
	legacy := convertLegacyCampaignSaveState(legacyCampaignSaveState{
		Factions: map[faction.FactionID]legacyFactionSaveState{
			fid: {},
		},
	})
	loaded := &faction.Faction{ID: fid, DiplomacyDisabled: true}
	applyFactionSaveState(loaded, legacy.Factions[fid])
	if !loaded.DiplomacyDisabled {
		t.Fatal("diplomacy_disabled alanı olmayan eski save senaryo varsayılanını değiştirdi")
	}
}
