package render

import "testing"

func TestConfirmDialogModalLeavesMessageGapAboveButtons(t *testing.T) {
	for _, lineCount := range []int{2, 3, 8} {
		state := confirmDialogState{
			messageLines: make([]string, lineCount),
			acceptLabel:  "Evet",
			declineLabel: "Hayır",
		}
		modal := buildConfirmDialogModalFor(state)
		accept, _, _, _ := buildConfirmDialogButtons(state)
		messageBottom := modal.Panel.Rect.Y + confirmDialogMessageY + float64(lineCount)*confirmDialogMessageLineStep
		gap := accept.Y - messageBottom

		if gap < confirmDialogMessageButtonGap {
			t.Errorf("%d message lines leave %.1fpx above buttons, want at least %.1fpx",
				lineCount, gap, confirmDialogMessageButtonGap)
		}
		if lineCount >= 3 && gap != confirmDialogMessageButtonGap {
			t.Errorf("%d message lines leave %.1fpx above buttons, want exactly %.1fpx",
				lineCount, gap, confirmDialogMessageButtonGap)
		}
	}
}
