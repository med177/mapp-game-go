package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ListViewStyle struct {
	RowBG            color.RGBA
	SelectedRowBG    color.RGBA
	TextColor        color.RGBA
	SelectedText     color.RGBA
	MutedText        color.RGBA
	RowTextOffsetY   float64
	TextVariant      TextVariant
	PaginationPrefix string
	DrawFrame        bool
	FrameBG          color.RGBA
	FrameBorder      color.RGBA
	FrameBorderWidth float32
	EmptyText        string
	EmptyTextColor   color.RGBA
	EmptyTextVariant TextVariant
	EmptyTextAlign   TextAlign
}

type ListViewInsets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

type ListView struct {
	Rect          Rect
	ContentInsets ListViewInsets
	Items         []string
	Selected      int
	Scroll        int
	RowHeight     float64
	VisibleRows   int
	Enabled       bool
	pressX        float64
	pressY        float64
	pressScroll   int
	pressIndex    int
	dragging      bool
	pressed       bool
}

const listViewDragThreshold = 8.0

func NewListView(x, y, w, h, rowHeight float64, visibleRows int, items []string) ListView {
	return ListView{
		Rect:        Rect{X: x, Y: y, W: w, H: h},
		Items:       append([]string(nil), items...),
		Selected:    -1,
		RowHeight:   rowHeight,
		VisibleRows: visibleRows,
		Enabled:     true,
	}
}

func (l ListView) HitTest(mx, my float64) bool {
	return l.Enabled && l.Rect.Hit(mx, my)
}

// ContentRect returns the inset viewport shared by row drawing and hit-testing.
func (l ListView) ContentRect() Rect {
	content := Rect{
		X: l.Rect.X + l.ContentInsets.Left,
		Y: l.Rect.Y + l.ContentInsets.Top,
		W: l.Rect.W - l.ContentInsets.Left - l.ContentInsets.Right,
		H: l.Rect.H - l.ContentInsets.Top - l.ContentInsets.Bottom,
	}
	if content.W < 0 {
		content.W = 0
	}
	if content.H < 0 {
		content.H = 0
	}
	return content
}

// ItemIndexAt returns the item index at a point, or -1 if it is outside a row.
func (l ListView) ItemIndexAt(mx, my float64) int {
	return l.ItemIndexAtCount(mx, my, len(l.Items))
}

// ItemIndexAtCount tests row geometry without requiring item labels to be built.
func (l ListView) ItemIndexAtCount(mx, my float64, itemCount int) int {
	content := l.ContentRect()
	if !content.Hit(mx, my) || l.RowHeight <= 0 {
		return -1
	}
	row := int((my - content.Y) / l.RowHeight)
	if row < 0 || row >= l.visibleRows() {
		return -1
	}
	index := l.Scroll + row
	if index < 0 || index >= itemCount {
		return -1
	}
	return index
}

func (l *ListView) HandleInput(input InputState) bool {
	if l == nil || !l.Enabled {
		return false
	}
	if input.WheelY != 0 && l.HitTest(input.MouseX, input.MouseY) {
		l.scroll(input.WheelY)
		return true
	}

	if input.LeftJustPressed {
		if !l.HitTest(input.MouseX, input.MouseY) {
			return false
		}
		l.pressed = true
		l.dragging = false
		l.pressX = input.MouseX
		l.pressY = input.MouseY
		l.pressScroll = l.Scroll
		l.pressIndex = l.itemIndexAt(input.MouseX, input.MouseY)
		return true
	}

	if l.pressed {
		if input.LeftPressed {
			dx := absF(input.MouseX - l.pressX)
			dy := absF(input.MouseY - l.pressY)
			if dx >= listViewDragThreshold || dy >= listViewDragThreshold {
				l.dragging = true
			}
			if l.dragging && l.RowHeight > 0 {
				rowDelta := int((l.pressY - input.MouseY) / l.RowHeight)
				l.Scroll = l.pressScroll + rowDelta
				l.clampScroll()
			}
			return true
		}
		if input.LeftJustReleased {
			pressedIndex := l.pressIndex
			dragging := l.dragging
			l.pressed = false
			l.dragging = false
			l.pressIndex = -1
			if dragging {
				return true
			}
			idx := l.itemIndexAt(input.MouseX, input.MouseY)
			if idx >= 0 && idx == pressedIndex {
				l.Selected = idx
				return true
			}
		}
	}
	return false
}

func (l ListView) Draw(_ *ebiten.Image, _ TextRenderer) {}

func (l *ListView) scroll(dy float64) {
	if dy > 0 {
		l.Scroll--
	} else if dy < 0 {
		l.Scroll++
	}
	l.clampScroll()
}

func (l ListView) itemIndexAt(mx, my float64) int {
	return l.ItemIndexAt(mx, my)
}

func (l ListView) visibleRows() int {
	if l.RowHeight <= 0 || l.VisibleRows <= 0 {
		return 0
	}
	rowsFit := int(l.ContentRect().H / l.RowHeight)
	if rowsFit < l.VisibleRows {
		return rowsFit
	}
	return l.VisibleRows
}

func (l *ListView) clampScroll() {
	maxScroll := len(l.Items) - l.visibleRows()
	if maxScroll < 0 {
		maxScroll = 0
	}
	if l.Scroll < 0 {
		l.Scroll = 0
	}
	if l.Scroll > maxScroll {
		l.Scroll = maxScroll
	}
}

func absF(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func DrawListView(screen *ebiten.Image, l ListView, style ListViewStyle, text TextRenderer) {
	if style.DrawFrame {
		vector.FillRect(screen, float32(l.Rect.X), float32(l.Rect.Y), float32(l.Rect.W), float32(l.Rect.H), style.FrameBG, false)
		if style.FrameBorderWidth > 0 {
			vector.StrokeRect(screen, float32(l.Rect.X), float32(l.Rect.Y), float32(l.Rect.W), float32(l.Rect.H), style.FrameBorderWidth, style.FrameBorder, false)
		}
	}
	content := l.ContentRect()
	visibleRows := l.visibleRows()
	end := l.Scroll + visibleRows
	if end > len(l.Items) {
		end = len(l.Items)
	}
	for i := l.Scroll; i < end; i++ {
		ry := content.Y + float64(i-l.Scroll)*l.RowHeight
		bg := style.RowBG
		txt := style.TextColor
		if i == l.Selected {
			bg = style.SelectedRowBG
			txt = style.SelectedText
		}
		vector.FillRect(screen, float32(content.X), float32(ry), float32(content.W), float32(l.RowHeight-2), bg, false)
		text.Draw(screen, l.Items[i], content.X+8, ry+style.RowTextOffsetY, txt, style.TextVariant)
	}
	if len(l.Items) == 0 && style.EmptyText != "" {
		emptyRect := content
		textY := content.Y + (content.H-12)/2
		text.Draw(screen, style.EmptyText, alignedTextX(text, emptyRect, style.EmptyText, style.EmptyTextVariant, style.EmptyTextAlign), textY, style.EmptyTextColor, style.EmptyTextVariant)
	}
	if len(l.Items) > visibleRows {
		info := itoa(l.Scroll+1) + "-" + itoa(end) + "/" + itoa(len(l.Items))
		if style.PaginationPrefix != "" {
			info = style.PaginationPrefix + " " + info
		}
		text.Draw(screen, info, content.X+8, content.Y+content.H+4, style.MutedText, style.TextVariant)
	}
}
