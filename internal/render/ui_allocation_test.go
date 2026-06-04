package render

import "testing"

func TestCoreUIBuildersAvoidHeapAllocations(t *testing.T) {
	cases := []struct {
		name string
		fn   func()
	}{
		{"back_button", func() { _ = buildBackButton() }},
		{"war_modal", func() { _ = buildWarConfirmModal() }},
		{"war_buttons", func() { _, _ = buildWarConfirmButtons() }},
		{"event_detail_modal", func() { _ = buildEventDetailModal() }},
		{"event_detail_close", func() { _ = buildEventDetailCloseButton() }},
		{"event_codex_modal", func() { _ = buildEventCodexModal() }},
		{"event_codex_close", func() { _ = buildEventCodexCloseButton() }},
		{"diplomacy_offer_modal", func() { _ = buildDiplomacyOfferModal() }},
		{"diplomacy_offer_buttons", func() { _, _ = buildDiplomacyOfferButtons() }},
	}

	for _, tc := range cases {
		allocs := testing.AllocsPerRun(1000, tc.fn)
		if allocs != 0 {
			t.Fatalf("%s allocated %.2f times", tc.name, allocs)
		}
	}
}
