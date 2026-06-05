package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type KeyValueRow struct {
	Rect         Rect
	Label        string
	Value        string
	LabelColor   color.Color
	ValueColor   color.Color
	LabelVariant TextVariant
	ValueVariant TextVariant
	Gap          float64
	ValueAlign   TextAlign
	Visible      bool
}

func NewKeyValueRow(rect Rect, label, value string) KeyValueRow {
	return KeyValueRow{
		Rect:         rect,
		Label:        label,
		Value:        value,
		LabelColor:   color.White,
		ValueColor:   color.White,
		LabelVariant: TextSmall,
		ValueVariant: TextMedium,
		Gap:          26,
		ValueAlign:   TextAlignEnd,
		Visible:      true,
	}
}

func (r KeyValueRow) HitTest(_, _ float64) bool     { return false }
func (r KeyValueRow) HandleInput(_ InputState) bool { return false }

func (r KeyValueRow) Draw(screen *ebiten.Image, text TextRenderer) {
	if !r.Visible {
		return
	}
	text.Draw(screen, r.Label, r.Rect.X, r.Rect.Y, r.LabelColor, r.LabelVariant)
	valueW := text.Measure(r.Value, r.ValueVariant)
	valueX := r.Rect.X + r.Rect.W - valueW
	minValueX := r.Rect.X + text.Measure(r.Label, r.LabelVariant) + r.Gap
	if r.ValueAlign == TextAlignStart {
		valueX = minValueX
	} else if valueX < minValueX {
		valueX = minValueX
	}
	text.Draw(screen, r.Value, valueX, r.Rect.Y, r.ValueColor, r.ValueVariant)
}

type TableCell struct {
	Text    string
	Color   color.Color
	Variant TextVariant
	Align   TextAlign
	Weight  float64
}

type TableRow struct {
	Rect    Rect
	Cells   []TableCell
	Gap     float64
	Visible bool
}

func NewTableRow(rect Rect, cells []TableCell, gap float64) TableRow {
	out := make([]TableCell, len(cells))
	copy(out, cells)
	return TableRow{Rect: rect, Cells: out, Gap: gap, Visible: true}
}

func (r TableRow) HitTest(_, _ float64) bool     { return false }
func (r TableRow) HandleInput(_ InputState) bool { return false }

func (r TableRow) Draw(screen *ebiten.Image, text TextRenderer) {
	if !r.Visible || len(r.Cells) == 0 {
		return
	}
	totalWeight := 0.0
	for _, cell := range r.Cells {
		if cell.Weight > 0 {
			totalWeight += cell.Weight
		}
	}
	if totalWeight <= 0 {
		totalWeight = float64(len(r.Cells))
	}
	totalGap := r.Gap * float64(len(r.Cells)-1)
	usable := r.Rect.W - totalGap
	if usable < 0 {
		usable = 0
	}
	x := r.Rect.X
	for _, cell := range r.Cells {
		weight := cell.Weight
		if weight <= 0 {
			weight = 1
		}
		w := usable * (weight / totalWeight)
		drawX := alignedTextX(text, Rect{X: x, Y: r.Rect.Y, W: w}, cell.Text, cell.Variant, cell.Align)
		text.Draw(screen, cell.Text, drawX, r.Rect.Y, cell.Color, cell.Variant)
		x += w + r.Gap
	}
}
