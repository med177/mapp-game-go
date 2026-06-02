package ui

// Rect ortak geometri yüzeyidir.
type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func (r Rect) Hit(mx, my float64) bool {
	return mx >= r.X && mx <= r.X+r.W && my >= r.Y && my <= r.Y+r.H
}
