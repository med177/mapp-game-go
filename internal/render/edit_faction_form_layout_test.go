package render

import "testing"

func rectsOverlap(a, b uiRect) bool {
	return a[0] < b[0]+b[2] && a[0]+a[2] > b[0] &&
		a[1] < b[1]+b[3] && a[1]+a[3] > b[1]
}

func TestEditFactionFormLayoutHasNoOverlappingInteractiveRects(t *testing.T) {
	fields := make([]uiRect, 0, editFactionFieldAI-editFactionFieldNone)
	for field := editFactionFieldID; field <= editFactionFieldAI; field++ {
		fields = append(fields, editFactionFieldRect(field))
	}
	for i := range fields {
		for j := i + 1; j < len(fields); j++ {
			if rectsOverlap(fields[i], fields[j]) {
				t.Fatalf("faction form fields %d and %d overlap: %v / %v", i, j, fields[i], fields[j])
			}
		}
	}

	buttons := make([]uiRect, 0, editFactionFormBluePlus-editFactionFormSave+1)
	for kind := editFactionFormSave; kind <= editFactionFormBluePlus; kind++ {
		buttons = append(buttons, editFactionFormButtonRect(kind))
	}
	for i := range buttons {
		for j := i + 1; j < len(buttons); j++ {
			if rectsOverlap(buttons[i], buttons[j]) {
				t.Fatalf("faction form buttons %d and %d overlap: %v / %v", i, j, buttons[i], buttons[j])
			}
		}
	}
	for i, field := range fields {
		for j, button := range buttons {
			if rectsOverlap(field, button) {
				t.Fatalf("faction form field %d and button %d overlap: %v / %v", i, j, field, button)
			}
		}
	}
}
