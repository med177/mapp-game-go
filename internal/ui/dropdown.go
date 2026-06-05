package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type DropdownStyle struct {
	PanelBG       color.RGBA
	Border        color.RGBA
	TitleColor    color.RGBA
	RowBG         color.RGBA
	SelectedRowBG color.RGBA
	RowText       color.RGBA
	SelectedText  color.RGBA
	MutedText     color.RGBA
	TitleOffsetY  float64
	RowOffsetY    float64
	TextVariant   TextVariant
	BorderWidth   float32
}

type Dropdown struct {
	open        bool
	scroll      int
	options     []string
	selected    string
	x, y, w, h  float64
	title       string
	headerH     float64
	rowH        float64
	visibleRows int
}

func NewDropdown(x, y, w, h float64, title string, headerH, rowH float64, visibleRows int) *Dropdown {
	return &Dropdown{
		x:           x,
		y:           y,
		w:           w,
		h:           h,
		title:       title,
		headerH:     headerH,
		rowH:        rowH,
		visibleRows: visibleRows,
	}
}

func (d *Dropdown) SetPosition(x, y float64) { d.x, d.y = x, y }

func (d *Dropdown) SetOptions(options []string, selected string) {
	d.options = make([]string, len(options))
	copy(d.options, options)
	d.selected = selected
	d.scroll = 0
}

func (d *Dropdown) Toggle() {
	d.open = !d.open
	if d.open {
		d.scroll = 0
	}
}

func (d *Dropdown) Close() {
	d.open = false
	d.scroll = 0
}

func (d *Dropdown) IsOpen() bool { return d.open }

func (d *Dropdown) HitTest(mx, my float64) bool {
	return mx >= d.x && mx <= d.x+d.w && my >= d.y && my <= d.y+d.h
}

func (d *Dropdown) Scroll(dy float64) {
	if dy > 0 {
		d.scroll--
	} else if dy < 0 {
		d.scroll++
	}
	maxScroll := len(d.options) - d.visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if d.scroll < 0 {
		d.scroll = 0
	}
	if d.scroll > maxScroll {
		d.scroll = maxScroll
	}
}

func (d *Dropdown) GetSelectedOption(mx, my float64) (int, bool) {
	if mx < d.x+8 || mx > d.x+d.w-8 {
		return 0, false
	}
	startY := d.y + d.headerH
	if my < startY {
		return 0, false
	}
	row := int((my - startY) / d.rowH)
	if row < 0 || row >= d.visibleRows {
		return 0, false
	}
	idx := d.scroll + row
	if idx < 0 || idx >= len(d.options) {
		return 0, false
	}
	return idx, true
}

func (d *Dropdown) Options() []string { return d.options }

func (d *Dropdown) OptionAt(idx int) string {
	if idx < 0 || idx >= len(d.options) {
		return ""
	}
	return d.options[idx]
}

func DrawDropdown(screen *ebiten.Image, d *Dropdown, style DropdownStyle, text TextRenderer) {
	if d == nil || !d.open {
		return
	}
	vector.FillRect(screen, float32(d.x), float32(d.y), float32(d.w), float32(d.h), style.PanelBG, false)
	vector.StrokeRect(screen, float32(d.x), float32(d.y), float32(d.w), float32(d.h), style.BorderWidth, style.Border, false)
	text.Draw(screen, d.title, d.x+10, d.y+style.TitleOffsetY, style.TitleColor, style.TextVariant)

	rowX := d.x + 8
	rowY := d.y + d.headerH
	rowW := d.w - 16
	for i := 0; i < d.visibleRows; i++ {
		optionIndex := d.scroll + i
		if optionIndex >= len(d.options) {
			break
		}
		option := d.options[optionIndex]
		oy := rowY + float64(i)*d.rowH
		bg := style.RowBG
		txt := style.RowText
		if option == d.selected {
			bg = style.SelectedRowBG
			txt = style.SelectedText
		}
		vector.FillRect(screen, float32(rowX), float32(oy), float32(rowW), float32(d.rowH-2), bg, false)
		text.Draw(screen, option, rowX+8, oy+style.RowOffsetY, txt, style.TextVariant)
	}
	if len(d.options) > d.visibleRows {
		info := itoa(d.scroll+1) + "-" + itoa(minInt(d.scroll+d.visibleRows, len(d.options))) + "/" + itoa(len(d.options))
		text.Draw(screen, info, d.x+d.w-68, d.y+style.TitleOffsetY, style.MutedText, style.TextVariant)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
