package render

import (
	"strings"
	"testing"
)

func TestTopResourceHUDColumnsReduceLeftPaddingAndKeepGaps(t *testing.T) {
	originalWidth, originalHeight := ScreenWidth, ScreenHeight
	defer func() {
		ScreenWidth, ScreenHeight = originalWidth, originalHeight
	}()
	ScreenWidth, ScreenHeight = 1280, 720

	leftCol1, leftCol2, rightCol, leftColW, rightColW := topResourceHUDColumns()
	firstGap := leftCol2 - (leftCol1 + leftColW)
	secondGap := rightCol - (leftCol2 + leftColW)
	if leftCol1 >= 300 {
		t.Errorf("first resource column starts at %.1f; want it moved left of 300", leftCol1)
	}
	if firstGap < 18 || secondGap < 18 {
		t.Errorf("resource column gaps are %.1f and %.1f; want both at least 18", firstGap, secondGap)
	}
	if firstGap != secondGap {
		t.Errorf("resource column gaps differ: first %.1f, second %.1f", firstGap, secondGap)
	}
	if rightCol+rightColW > victoryProgressPanelRect().X-12 {
		t.Errorf("last resource column ends at %.1f, past its available edge %.1f", rightCol+rightColW, victoryProgressPanelRect().X-12)
	}

	textX := 104.0
	name := factionHUDNameText(strings.Repeat("Uzun Devlet Adı ", 8), textX)
	maxNameWidth := leftCol1 - textX - 12
	if got := MeasureText(name, FaceLarge); got > maxNameWidth {
		t.Errorf("trimmed faction name width %.1f exceeds available width %.1f", got, maxNameWidth)
	}
}
