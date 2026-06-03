package ui

type HAnchor int

const (
	AnchorLeft HAnchor = iota
	AnchorCenter
	AnchorRight
)

type VAnchor int

const (
	AnchorTop VAnchor = iota
	AnchorMiddle
	AnchorBottom
)

func AnchorRect(parent Rect, w, h float64, hAnchor HAnchor, vAnchor VAnchor, marginX, marginY float64) Rect {
	x := parent.X + marginX
	switch hAnchor {
	case AnchorCenter:
		x = parent.X + (parent.W-w)/2 + marginX
	case AnchorRight:
		x = parent.X + parent.W - w - marginX
	}

	y := parent.Y + marginY
	switch vAnchor {
	case AnchorMiddle:
		y = parent.Y + (parent.H-h)/2 + marginY
	case AnchorBottom:
		y = parent.Y + parent.H - h - marginY
	}

	return Rect{X: x, Y: y, W: w, H: h}
}
