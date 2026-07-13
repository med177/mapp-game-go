package render

import gameui "mapp-game-go/internal/ui"

func centeredStackRect(count int, itemW, itemH, gap, headerH float64) gameui.Rect {
	totalH := float64(count)*itemH + float64(maxScreenInt(count-1, 0))*gap
	return gameui.Rect{
		X: ScreenWidth/2 - itemW/2,
		Y: ScreenHeight/2 - (totalH+headerH)/2 + headerH,
		W: itemW,
		H: totalH,
	}
}

func stackItemRect(stack gameui.Rect, itemH, gap float64, index int) gameui.Rect {
	return gameui.Rect{
		X: stack.X,
		Y: stack.Y + float64(index)*(itemH+gap),
		W: stack.W,
		H: itemH,
	}
}

func centeredGridRect(cols, rows int, cellW, cellH, gapX, gapY, headerH float64) gameui.Rect {
	gridW := cellW*float64(cols) + gapX*float64(maxScreenInt(cols-1, 0))
	gridH := cellH*float64(rows) + gapY*float64(maxScreenInt(rows-1, 0))
	return gameui.Rect{
		X: ScreenWidth/2 - gridW/2,
		Y: ScreenHeight/2 - (gridH+headerH)/2 + headerH,
		W: gridW,
		H: gridH,
	}
}

func gridCellRect(grid gameui.Rect, cellW, cellH, gapX, gapY float64, col, row int) gameui.Rect {
	return gameui.Rect{
		X: grid.X + float64(col)*(cellW+gapX),
		Y: grid.Y + float64(row)*(cellH+gapY),
		W: cellW,
		H: cellH,
	}
}

func maxScreenInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minScreenInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
