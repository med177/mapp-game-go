package render

import "testing"

func TestNearestEventCodexEntrySkipsLockedEvents(t *testing.T) {
	r := &Renderer{}
	r.eventCodexEntries[EventCodexPlayer] = []EventCodexEntry{
		{EventID: "locked_state_event", Title: "Kilitli state olayı", Status: "Kilitli", TurnsUntil: 0},
		{EventID: "first_calendar_event", Title: "Takvimdeki ilk olay", Status: "Takvim", TurnsUntil: 7},
		{EventID: "later_calendar_event", Title: "Daha sonraki olay", Status: "Takvim", TurnsUntil: 12},
	}

	got, ok := r.nearestEventCodexEntry()
	if !ok {
		t.Fatal("kilitli olaylar atlandıktan sonra takvim olayı bulunamadı")
	}
	if got.EventID != "first_calendar_event" {
		t.Fatalf("HUD yanlış olayı seçti: got %q, want %q", got.EventID, "first_calendar_event")
	}
}

func TestNearestEventCodexEntryHidesWhenOnlyLockedEventsExist(t *testing.T) {
	r := &Renderer{}
	r.eventCodexEntries[EventCodexPlayer] = []EventCodexEntry{
		{EventID: "locked_state_event", Title: "Kilitli state olayı", Status: "Kilitli", TurnsUntil: 0},
		{EventID: "fired_event", Title: "Gerçekleşmiş olay", Status: "Gerçekleşti", TurnsUntil: 0},
	}

	if got, ok := r.nearestEventCodexEntry(); ok {
		t.Fatalf("yaklaşan olay yokken HUD kayıt döndürdü: %+v", got)
	}
}

func TestNearestGeneralEventCodexEntrySkipsPlayerAndLockedEvents(t *testing.T) {
	r := &Renderer{}
	r.eventCodexEntries[EventCodexAll] = []EventCodexEntry{
		{EventID: "player_event", Title: "Oyuncu olayı", Status: "Takvim", TurnsUntil: 2},
		{EventID: "locked_general", Title: "Kilitli genel olay", Status: "Kilitli", TurnsUntil: 1},
		{EventID: "first_general", Title: "İlk genel olay", Status: "Hazir", TurnsUntil: 4},
		{EventID: "later_general", Title: "Sonraki genel olay", Status: "Takvim", TurnsUntil: 8},
	}
	r.eventCodexEntries[EventCodexPlayer] = []EventCodexEntry{
		{EventID: "player_event", Title: "Oyuncu olayı", Status: "Takvim", TurnsUntil: 2},
	}

	got, ok := r.nearestGeneralEventCodexEntry()
	if !ok {
		t.Fatal("genel yaklaşan olay bulunamadı")
	}
	if got.EventID != "first_general" {
		t.Fatalf("HUD yanlış genel olayı seçti: got %q, want %q", got.EventID, "first_general")
	}
}

func TestNearestGeneralEventCodexEntryHidesWhenNoGeneralEventsExist(t *testing.T) {
	r := &Renderer{}
	r.eventCodexEntries[EventCodexAll] = []EventCodexEntry{
		{EventID: "player_event", Title: "Oyuncu olayı", Status: "Takvim", TurnsUntil: 2},
		{EventID: "locked_general", Title: "Kilitli genel olay", Status: "Kilitli", TurnsUntil: 1},
	}
	r.eventCodexEntries[EventCodexPlayer] = []EventCodexEntry{
		{EventID: "player_event", Title: "Oyuncu olayı", Status: "Takvim", TurnsUntil: 2},
	}

	if got, ok := r.nearestGeneralEventCodexEntry(); ok {
		t.Fatalf("genel yaklaşan olay yokken HUD kayıt döndürdü: %+v", got)
	}
}
