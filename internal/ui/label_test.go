package ui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type stubTextRenderer struct {
	widths map[string]float64
	draws  []stubDrawCall
}

type stubDrawCall struct {
	text    string
	x       float64
	y       float64
	variant TextVariant
}

func (s *stubTextRenderer) Measure(text string, variant TextVariant) float64 {
	if s.widths == nil {
		return 0
	}
	return s.widths[text]
}

func (s *stubTextRenderer) Draw(_ *ebiten.Image, text string, x, y float64, _ color.Color, variant TextVariant) {
	s.draws = append(s.draws, stubDrawCall{text: text, x: x, y: y, variant: variant})
}

func TestLabelDrawAlignments(t *testing.T) {
	screen := ebiten.NewImage(16, 16)
	renderer := &stubTextRenderer{widths: map[string]float64{"abc": 12}}

	labels := []Label{
		NewTextLabel(Rect{X: 10, Y: 4}, "abc", color.White, TextSmall, TextAlignStart),
		NewTextLabel(Rect{X: 10, Y: 8}, "abc", color.White, TextMedium, TextAlignCenter),
		NewTextLabel(Rect{X: 10, Y: 12}, "abc", color.White, TextLarge, TextAlignEnd),
		NewTextLabel(Rect{X: 20, Y: 16, W: 40}, "abc", color.White, TextSmall, TextAlignCenter),
	}
	for _, label := range labels {
		label.Draw(screen, renderer)
	}
	if len(renderer.draws) != 4 {
		t.Fatalf("expected 4 draw calls, got %d", len(renderer.draws))
	}
	if renderer.draws[0].x != 10 {
		t.Fatalf("start align x = %.1f, want 10", renderer.draws[0].x)
	}
	if renderer.draws[1].x != 4 {
		t.Fatalf("center point x = %.1f, want 4", renderer.draws[1].x)
	}
	if renderer.draws[2].x != -2 {
		t.Fatalf("end point x = %.1f, want -2", renderer.draws[2].x)
	}
	if renderer.draws[3].x != 34 {
		t.Fatalf("center rect x = %.1f, want 34", renderer.draws[3].x)
	}
	if renderer.draws[1].variant != TextMedium || renderer.draws[2].variant != TextLarge {
		t.Fatalf("expected variants to propagate, got %+v", renderer.draws)
	}
}
