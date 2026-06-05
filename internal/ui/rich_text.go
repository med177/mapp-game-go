package ui

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type RichTextLine struct {
	Text    string
	Color   color.Color
	Variant TextVariant
	Align   TextAlign
}

type RichTextBlock struct {
	Rect     Rect
	Lines    []RichTextLine
	LineStep float64
	Visible  bool
}

func NewRichTextBlock(rect Rect, lines []RichTextLine, lineStep float64) RichTextBlock {
	out := make([]RichTextLine, len(lines))
	copy(out, lines)
	return RichTextBlock{
		Rect:     rect,
		Lines:    out,
		LineStep: lineStep,
		Visible:  true,
	}
}

func (b RichTextBlock) HitTest(_, _ float64) bool     { return false }
func (b RichTextBlock) HandleInput(_ InputState) bool { return false }

func (b RichTextBlock) Draw(screen *ebiten.Image, text TextRenderer) {
	if !b.Visible {
		return
	}
	for i, line := range b.Lines {
		x := alignedTextX(text, b.Rect, line.Text, line.Variant, line.Align)
		text.Draw(screen, line.Text, x, b.Rect.Y+float64(i)*b.LineStep, line.Color, line.Variant)
	}
}

type WrappedLabel struct {
	Rect     Rect
	Text     string
	Color    color.Color
	Variant  TextVariant
	LineStep float64
	MaxLines int
	Ellipsis bool
	Align    TextAlign
	Visible  bool
}

func NewWrappedLabel(rect Rect, text string, col color.Color, variant TextVariant, lineStep float64) WrappedLabel {
	return WrappedLabel{
		Rect:     rect,
		Text:     text,
		Color:    col,
		Variant:  variant,
		LineStep: lineStep,
		Align:    TextAlignStart,
		Visible:  true,
	}
}

func (w WrappedLabel) HitTest(_, _ float64) bool     { return false }
func (w WrappedLabel) HandleInput(_ InputState) bool { return false }

func (w WrappedLabel) Draw(screen *ebiten.Image, text TextRenderer) {
	if !w.Visible || w.Text == "" || w.Rect.W <= 0 {
		return
	}
	lines := wrapLines(text, w.Text, w.Rect.W, w.Variant)
	if w.MaxLines > 0 && len(lines) > w.MaxLines {
		lines = lines[:w.MaxLines]
		if w.Ellipsis && len(lines) > 0 {
			lines[len(lines)-1] = trimWrappedLine(text, lines[len(lines)-1]+"...", w.Rect.W, w.Variant)
		}
	}
	for i, line := range lines {
		x := alignedTextX(text, w.Rect, line, w.Variant, w.Align)
		text.Draw(screen, line, x, w.Rect.Y+float64(i)*w.LineStep, w.Color, w.Variant)
	}
}

type OutlinedLabel struct {
	Label        Label
	OutlineColor color.Color
	Offsets      [][2]float64
}

func NewOutlinedLabel(rect Rect, text string, fill color.Color, outline color.Color, variant TextVariant, align TextAlign) OutlinedLabel {
	return OutlinedLabel{
		Label: Label{
			Rect:    rect,
			Text:    text,
			Color:   fill,
			Variant: variant,
			Align:   align,
			Visible: true,
		},
		OutlineColor: outline,
		Offsets: [][2]float64{
			{-1, 0},
			{1, 0},
			{0, -1},
			{0, 1},
		},
	}
}

func (o OutlinedLabel) HitTest(mx, my float64) bool       { return o.Label.HitTest(mx, my) }
func (o OutlinedLabel) HandleInput(input InputState) bool { return o.Label.HandleInput(input) }

func (o OutlinedLabel) Draw(screen *ebiten.Image, text TextRenderer) {
	if !o.Label.Visible {
		return
	}
	baseX := alignedTextX(text, o.Label.Rect, o.Label.Text, o.Label.Variant, o.Label.Align)
	for _, off := range o.Offsets {
		text.Draw(screen, o.Label.Text, baseX+off[0], o.Label.Rect.Y+off[1], o.OutlineColor, o.Label.Variant)
	}
	text.Draw(screen, o.Label.Text, baseX, o.Label.Rect.Y, o.Label.Color, o.Label.Variant)
}

func alignedTextX(text TextRenderer, rect Rect, label string, variant TextVariant, align TextAlign) float64 {
	x := rect.X
	if align == TextAlignStart {
		return x
	}
	w := text.Measure(label, variant)
	switch align {
	case TextAlignCenter:
		if rect.W > 0 {
			return rect.X + (rect.W-w)/2
		}
		return rect.X - w/2
	case TextAlignEnd:
		if rect.W > 0 {
			return rect.X + rect.W - w
		}
		return rect.X - w
	default:
		return x
	}
}

func wrapLines(text TextRenderer, s string, maxWidth float64, variant TextVariant) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 4)
	line := words[0]
	for _, word := range words[1:] {
		candidate := line + " " + word
		if text.Measure(candidate, variant) <= maxWidth {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = word
	}
	lines = append(lines, line)

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if text.Measure(line, variant) <= maxWidth {
			out = append(out, line)
			continue
		}
		out = append(out, splitWrappedWord(text, line, maxWidth, variant)...)
	}
	return out
}

func splitWrappedWord(text TextRenderer, word string, maxWidth float64, variant TextVariant) []string {
	runes := []rune(word)
	if len(runes) == 0 {
		return nil
	}
	lines := make([]string, 0, 2)
	start := 0
	for start < len(runes) {
		end := start + 1
		for end <= len(runes) && text.Measure(string(runes[start:end]), variant) <= maxWidth {
			end++
		}
		if end == start+1 {
			end = start + 1
		} else {
			end--
		}
		lines = append(lines, string(runes[start:end]))
		start = end
	}
	return lines
}

func trimWrappedLine(text TextRenderer, line string, maxWidth float64, variant TextVariant) string {
	if text.Measure(line, variant) <= maxWidth {
		return line
	}
	runes := []rune(line)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		candidate := string(runes)
		if text.Measure(candidate, variant) <= maxWidth {
			return candidate
		}
	}
	return ""
}
