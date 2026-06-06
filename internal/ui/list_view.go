package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ListViewStyle struct {
	RowBG          color.RGBA
	SelectedRowBG  color.RGBA
	TextColor      color.RGBA
	SelectedText   color.RGBA
	MutedText      color.RGBA
	RowTextOffsetY float64
	TextVariant    TextVariant
}

type ListView struct {
	Rect        Rect
	Items       []string
	Selected    int
	Scroll      int
	RowHeight   float64
	VisibleRows int
	Enabled     bool
	pressX      float64
	pressY      float64
	pressScroll int
	pressIndex  int
	dragging    bool
	pressed     bool
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
	if !l.Rect.Hit(mx, my) || l.RowHeight <= 0 {
		return -1
	}
	row := int((my - l.Rect.Y) / l.RowHeight)
	idx := l.Scroll + row
	if row < 0 || row >= l.VisibleRows || idx < 0 || idx >= len(l.Items) {
		return -1
	}
	return idx
}

func (l *ListView) clampScroll() {
	maxScroll := len(l.Items) - l.VisibleRows
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
	end := l.Scroll + l.VisibleRows
	if end > len(l.Items) {
		end = len(l.Items)
	}
	for i := l.Scroll; i < end; i++ {
		ry := l.Rect.Y + float64(i-l.Scroll)*l.RowHeight
		bg := style.RowBG
		txt := style.TextColor
		if i == l.Selected {
			bg = style.SelectedRowBG
			txt = style.SelectedText
		}
		vector.FillRect(screen, float32(l.Rect.X), float32(ry), float32(l.Rect.W), float32(l.RowHeight-2), bg, false)
		text.Draw(screen, l.Items[i], l.Rect.X+8, ry+style.RowTextOffsetY, txt, style.TextVariant)
	}
	if len(l.Items) > l.VisibleRows {
		info := itoa(l.Scroll+1) + "-" + itoa(end) + "/" + itoa(len(l.Items))
		text.Draw(screen, info, l.Rect.X+8, l.Rect.Y+l.Rect.H+4, style.MutedText, style.TextVariant)
	}
}
