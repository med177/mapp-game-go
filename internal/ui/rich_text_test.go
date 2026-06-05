package ui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestWrappedLabelDrawWrapsAndClampsLines(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{
		"bir":            12,
		"iki":            12,
		"uc":             8,
		"bir iki":        28,
		"iki uc":         24,
		"bir iki uc":     40,
		"uzun":           30,
		"uzunu":          36,
		"uzunuz":         42,
		"uzunuzu":        48,
		"uzunuzun":       54,
		"uzunuzunk":      60,
		"uzunuzunkel":    78,
		"uzunuzunkelim":  84,
		"uzunuzunkelime": 90,
	}}
	label := NewWrappedLabel(Rect{X: 5, Y: 7, W: 30}, "bir iki uc", color.White, TextSmall, 10)
	label.Draw(screen, renderer)
	if got := len(renderer.draws); got != 2 {
		t.Fatalf("expected 2 wrapped draws, got %d", got)
	}
	if renderer.draws[0].text != "bir iki" || renderer.draws[1].text != "uc" {
		t.Fatalf("unexpected wrapped lines: %+v", renderer.draws)
	}

	renderer.draws = nil
	clamped := NewWrappedLabel(Rect{X: 0, Y: 0, W: 30}, "bir iki uc", color.White, TextSmall, 10)
	clamped.MaxLines = 1
	clamped.Draw(screen, renderer)
	if got := len(renderer.draws); got != 1 || renderer.draws[0].text != "bir iki" {
		t.Fatalf("expected single clamped line, got %+v", renderer.draws)
	}
}

func TestOutlinedLabelDrawsOutlineAndFill(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{"abc": 12}}
	label := NewOutlinedLabel(Rect{X: 10, Y: 6}, "abc", color.White, color.Black, TextMedium, TextAlignCenter)
	label.Draw(screen, renderer)
	if got := len(renderer.draws); got != 5 {
		t.Fatalf("expected 5 draw calls for outline+fill, got %d", got)
	}
	if renderer.draws[4].x != 4 {
		t.Fatalf("expected centered fill x=4, got %.1f", renderer.draws[4].x)
	}
	if renderer.draws[4].variant != TextMedium {
		t.Fatalf("expected medium variant, got %v", renderer.draws[4].variant)
	}
}

func TestRichTextBlockDrawsPerLineStyle(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{"a": 6, "b": 6}}
	block := NewRichTextBlock(Rect{X: 20, Y: 10, W: 20}, []RichTextLine{
		{Text: "a", Color: color.White, Variant: TextSmall, Align: TextAlignStart},
		{Text: "b", Color: color.White, Variant: TextLarge, Align: TextAlignCenter},
	}, 12)
	block.Draw(screen, renderer)
	if got := len(renderer.draws); got != 2 {
		t.Fatalf("expected 2 draw calls, got %d", got)
	}
	if renderer.draws[0].x != 20 || renderer.draws[0].y != 10 {
		t.Fatalf("unexpected first line position: %+v", renderer.draws[0])
	}
	if renderer.draws[1].x != 27 || renderer.draws[1].y != 22 {
		t.Fatalf("unexpected second line position: %+v", renderer.draws[1])
	}
	if renderer.draws[1].variant != TextLarge {
		t.Fatalf("expected large variant, got %v", renderer.draws[1].variant)
	}
}

func TestKeyValueRowDrawsLabelAndValue(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{"L": 6, "Value": 20}}
	row := NewKeyValueRow(Rect{X: 10, Y: 5, W: 80}, "L", "Value")
	row.LabelVariant = TextSmall
	row.ValueVariant = TextMedium
	row.Draw(screen, renderer)
	if got := len(renderer.draws); got != 2 {
		t.Fatalf("expected 2 draw calls, got %d", got)
	}
	if renderer.draws[0].x != 10 || renderer.draws[1].x != 70 {
		t.Fatalf("unexpected positions: %+v", renderer.draws)
	}
}

func TestTableRowDrawsWeightedColumns(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{"A": 6, "B": 6}}
	row := NewTableRow(Rect{X: 0, Y: 10, W: 100}, []TableCell{
		{Text: "A", Color: color.White, Variant: TextSmall, Align: TextAlignStart, Weight: 1},
		{Text: "B", Color: color.White, Variant: TextSmall, Align: TextAlignCenter, Weight: 1},
	}, 0)
	row.Draw(screen, renderer)
	if got := len(renderer.draws); got != 2 {
		t.Fatalf("expected 2 draw calls, got %d", got)
	}
	if renderer.draws[0].x != 0 {
		t.Fatalf("first cell x=%.1f, want 0", renderer.draws[0].x)
	}
	if renderer.draws[1].x != 72 {
		t.Fatalf("second cell x=%.1f, want 72", renderer.draws[1].x)
	}
}
