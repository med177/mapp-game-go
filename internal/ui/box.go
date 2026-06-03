package ui

// Box rect üstünde akış tabanlı layout üretmek için hafif bir yardımcıdır.
// Widget primitive'lerini ortak tutmak tek başına yeterli olmadığından,
// ekran kompozisyonlarının da aynı kutu kurallarıyla kurulmasını sağlar.
type Box struct {
	Rect Rect
}

func BoxFromRect(r Rect) Box {
	return Box{Rect: r}
}

func (b Box) Inset(all float64) Box {
	return b.InsetXY(all, all)
}

func (b Box) InsetXY(x, y float64) Box {
	return Box{
		Rect: Rect{
			X: b.Rect.X + x,
			Y: b.Rect.Y + y,
			W: maxBoxFloat(0, b.Rect.W-2*x),
			H: maxBoxFloat(0, b.Rect.H-2*y),
		},
	}
}

func (b Box) CutTop(h, gap float64) (Rect, Box) {
	if h < 0 {
		h = 0
	}
	if h > b.Rect.H {
		h = b.Rect.H
	}
	top := Rect{X: b.Rect.X, Y: b.Rect.Y, W: b.Rect.W, H: h}
	restY := b.Rect.Y + h + gap
	restH := b.Rect.H - h - gap
	if restH < 0 {
		restH = 0
	}
	return top, Box{Rect: Rect{X: b.Rect.X, Y: restY, W: b.Rect.W, H: restH}}
}

func (b Box) CutBottom(h, gap float64) (Rect, Box) {
	if h < 0 {
		h = 0
	}
	if h > b.Rect.H {
		h = b.Rect.H
	}
	bottomY := b.Rect.Y + b.Rect.H - h
	bottom := Rect{X: b.Rect.X, Y: bottomY, W: b.Rect.W, H: h}
	restH := b.Rect.H - h - gap
	if restH < 0 {
		restH = 0
	}
	return bottom, Box{Rect: Rect{X: b.Rect.X, Y: b.Rect.Y, W: b.Rect.W, H: restH}}
}

func (b Box) CutLeft(w, gap float64) (Rect, Box) {
	if w < 0 {
		w = 0
	}
	if w > b.Rect.W {
		w = b.Rect.W
	}
	left := Rect{X: b.Rect.X, Y: b.Rect.Y, W: w, H: b.Rect.H}
	restX := b.Rect.X + w + gap
	restW := b.Rect.W - w - gap
	if restW < 0 {
		restW = 0
	}
	return left, Box{Rect: Rect{X: restX, Y: b.Rect.Y, W: restW, H: b.Rect.H}}
}

func (b Box) CutRight(w, gap float64) (Rect, Box) {
	if w < 0 {
		w = 0
	}
	if w > b.Rect.W {
		w = b.Rect.W
	}
	rightX := b.Rect.X + b.Rect.W - w
	right := Rect{X: rightX, Y: b.Rect.Y, W: w, H: b.Rect.H}
	restW := b.Rect.W - w - gap
	if restW < 0 {
		restW = 0
	}
	return right, Box{Rect: Rect{X: b.Rect.X, Y: b.Rect.Y, W: restW, H: b.Rect.H}}
}

func (b Box) SplitColumns(gap float64, weights ...float64) []Rect {
	return splitAxis(b.Rect, gap, true, weights...)
}

func (b Box) SplitRows(gap float64, weights ...float64) []Rect {
	return splitAxis(b.Rect, gap, false, weights...)
}

func splitAxis(r Rect, gap float64, horizontal bool, weights ...float64) []Rect {
	if len(weights) == 0 {
		return nil
	}
	totalWeight := 0.0
	for _, w := range weights {
		if w > 0 {
			totalWeight += w
		}
	}
	if totalWeight <= 0 {
		return nil
	}
	out := make([]Rect, 0, len(weights))
	span := r.H
	startX, startY := r.X, r.Y
	if horizontal {
		span = r.W
	}
	totalGap := gap * float64(len(weights)-1)
	usable := span - totalGap
	if usable < 0 {
		usable = 0
	}
	cursorX, cursorY := startX, startY
	for i, weight := range weights {
		size := 0.0
		if weight > 0 {
			size = usable * (weight / totalWeight)
		}
		if horizontal {
			out = append(out, Rect{X: cursorX, Y: startY, W: size, H: r.H})
			cursorX += size
			if i < len(weights)-1 {
				cursorX += gap
			}
		} else {
			out = append(out, Rect{X: startX, Y: cursorY, W: r.W, H: size})
			cursorY += size
			if i < len(weights)-1 {
				cursorY += gap
			}
		}
	}
	return out
}

func maxBoxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
