package ui

type Axis int

const (
	AxisVertical Axis = iota + 1
	AxisHorizontal
)

func VBox(x, y, w, itemH, gap float64, count int) []Rect {
	out := make([]Rect, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, Rect{
			X: x,
			Y: y + float64(i)*(itemH+gap),
			W: w,
			H: itemH,
		})
	}
	return out
}

func HBox(x, y, h, itemW, gap float64, count int) []Rect {
	out := make([]Rect, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, Rect{
			X: x + float64(i)*(itemW+gap),
			Y: y,
			W: itemW,
			H: h,
		})
	}
	return out
}

func Grid(x, y, cellW, cellH, gapX, gapY float64, cols, rows int) []Rect {
	out := make([]Rect, 0, cols*rows)
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			out = append(out, Rect{
				X: x + float64(col)*(cellW+gapX),
				Y: y + float64(row)*(cellH+gapY),
				W: cellW,
				H: cellH,
			})
		}
	}
	return out
}
