package render

import "testing"

func TestCollapsedEventLogKeepsButtonsInsetFromPanelBottom(t *testing.T) {
	layout := buildEventLogLayout(true)
	panelBottom := layout.panel.Rect.Y + layout.panel.Rect.H
	buttonBottom := layout.toggle.Y + layout.toggle.H
	if codexBottom := layout.codex.Y + layout.codex.H; codexBottom > buttonBottom {
		buttonBottom = codexBottom
	}

	if gap := panelBottom - buttonBottom; gap < eventLogPanelPad {
		t.Fatalf("collapsed event-log buttons have only %.1f px bottom padding, want at least %.1f px", gap, eventLogPanelPad)
	}
}
