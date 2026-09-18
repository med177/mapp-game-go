package render

import "testing"

func TestNormalizeEditID(t *testing.T) {
	if got := normalizeEditID("  NeW_Region-ID  "); got != "new_region-id" {
		t.Fatalf("normalizeEditID() = %q, want new_region-id", got)
	}
}

func TestAppendFactionFormRuneNormalizesUppercaseID(t *testing.T) {
	r := &Renderer{editFactionForm: editFactionFormState{active: editFactionFieldID}}
	r.appendFactionFormRune('A')

	if got := r.editFactionForm.id; got != "a" {
		t.Fatalf("faction ID after uppercase input = %q, want a", got)
	}
}
