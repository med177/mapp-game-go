package ui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type labelTestTextRenderer struct {
	drawnY []float64
	height float64
}

func (r *labelTestTextRenderer) Measure(string, TextVariant) float64 { return 0 }
func (r *labelTestTextRenderer) MeasureHeight(TextVariant) float64   { return r.height }
func (r *labelTestTextRenderer) Draw(_ *ebiten.Image, _ string, _, y float64, _ color.Color, _ TextVariant) {
	r.drawnY = append(r.drawnY, y)
}

func TestLabelVerticalAlignment(t *testing.T) {
	for _, tc := range []struct {
		name  string
		align TextVerticalAlign
		wantY float64
	}{
		{name: "top default", align: TextVerticalAlignTop, wantY: 20},
		{name: "center", align: TextVerticalAlignCenter, wantY: 35},
		{name: "bottom", align: TextVerticalAlignBottom, wantY: 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			renderer := &labelTestTextRenderer{height: 10}
			label := NewTextLabel(Rect{X: 10, Y: 20, W: 100, H: 40}, "test", color.White, TextSmall, TextAlignStart)
			label.VerticalAlign = tc.align
			label.Draw(nil, renderer)

			if len(renderer.drawnY) != 1 || renderer.drawnY[0] != tc.wantY {
				t.Fatalf("draw y = %v, want %v", renderer.drawnY, tc.wantY)
			}
		})
	}
}

func TestOutlinedLabelVerticalAlignmentAppliesToTextAndOutline(t *testing.T) {
	renderer := &labelTestTextRenderer{height: 10}
	label := NewOutlinedLabel(Rect{X: 10, Y: 20, W: 100, H: 40}, "test", color.White, color.Black, TextSmall, TextAlignStart)
	label.Label.VerticalAlign = TextVerticalAlignCenter
	label.Draw(nil, renderer)

	if len(renderer.drawnY) != 5 {
		t.Fatalf("draw calls = %d, want 5 (four outline offsets and text)", len(renderer.drawnY))
	}
	for i, y := range renderer.drawnY {
		if y != 35 && y != 34 && y != 36 {
			t.Errorf("draw call %d y = %.1f, want centered text or outline offset", i, y)
		}
	}
}
