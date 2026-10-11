package game

import (
	"testing"

	"mapp-game-go/internal/events"
	"mapp-game-go/internal/state"
)

func TestCodexReasonLabelExplainsFlagClearedByOccurredEvent(t *testing.T) {
	const activeWarFlag = "war_of_five_kings_active"
	game := &Game{
		gs: &state.GameState{
			FiredEventIDs: map[string]bool{
				"war_resolution_baratheon": true,
			},
		},
		evts: []*events.Event{
			{
				ID:       "death_of_robert_298",
				NameTR:   "Kral Robert'ın Ölümü",
				SetFlags: []string{activeWarFlag},
			},
			{
				ID:         "war_resolution_baratheon",
				NameTR:     "Beş Kralın Savaşı: Kralın Şehri Zaferi",
				ClearFlags: []string{activeWarFlag},
			},
		},
	}

	got := game.codexReasonLabel("flag bekleniyor: "+activeWarFlag, nil)
	want := "etkin olmayan flag: " + activeWarFlag + " (gerçekleşen event temizledi: Beş Kralın Savaşı: Kralın Şehri Zaferi)"
	if got != want {
		t.Fatalf("flag açıklaması = %q, want %q", got, want)
	}
}

func TestCodexReasonLabelKeepsSetterForUnclearedFlag(t *testing.T) {
	game := &Game{
		gs: &state.GameState{FiredEventIDs: map[string]bool{}},
		evts: []*events.Event{{
			ID:       "death_of_robert_298",
			NameTR:   "Kral Robert'ın Ölümü",
			SetFlags: []string{"war_of_five_kings_active"},
		}},
	}

	got := game.codexReasonLabel("flag bekleniyor: war_of_five_kings_active", nil)
	if want := "beklenen event: Kral Robert'ın Ölümü"; got != want {
		t.Fatalf("flag açıklaması = %q, want %q", got, want)
	}
}
