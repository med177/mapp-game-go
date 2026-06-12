package render

import (
	"image/color"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const factionGroupGap = 34.0
const factionGroupLabelH = 20.0

func buildFactionCardButtons(gs *state.GameState) []gameui.Button {
	factions, historicalCount := selectableFactions(gs)
	cols := 3
	cardW := 350.0
	cardH := 138.0
	padX := 30.0
	padY := 12.0
	headerH := 70.0
	buttons := make([]gameui.Button, 0, len(factions))
	for i, fid := range factions {
		r := factionCardRect(i, historicalCount, len(factions), cols, cardW, cardH, padX, padY, headerH)
		label := ""
		if f := gs.Factions[fid]; f != nil {
			label = f.NameTR
		}
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, label))
	}
	return buttons
}

// DrawFactionSelect fraksiyon seçim ekranını çizer.
func DrawFactionSelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	drawUIScreenChrome(screen, color.RGBA{10, 8, 5, 255}, "MAPP — Fraksiyon Seç", "Fraksiyon kartını seçmek için tıkla")

	factions, historicalCount := selectableFactions(gs)
	cols := 3
	cardW := float32(350)
	cardH := float32(138)
	headerH := 70.0

	drawBackButton(screen)

	drawFactionGroupLabels(screen, len(factions), historicalCount, cols, float64(cardW), float64(cardH), 30, 12, headerH)

	for i, fid := range factions {
		f := gs.Factions[fid]
		cell := factionCardRect(i, historicalCount, len(factions), cols, float64(cardW), float64(cardH), 30, 12, headerH)
		x := float32(cell.X)
		y := float32(cell.Y)

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if i == cursor {
			bgCol = color.RGBA{45, 36, 20, 240}
			borderCol = fc
		}

		drawUICardRect(screen, cell, bgCol, borderCol, 2)

		// Renk şeridi
		drawUICardAccent(screen, cell, 8, fc)

		// İsim
		nameCol := ColorWhite
		if i == cursor {
			nameCol = ColorYellow
		}
		drawUILabel(screen, gameui.Rect{X: float64(x + 16), Y: float64(y + 12)}, f.NameTR, nameCol, gameui.TextLarge, gameui.TextAlignStart)

		// Din
		drawUILabel(screen, gameui.Rect{X: float64(x + 16), Y: float64(y + 36)}, religion.DisplayNameTR(f.Religion), ColorGray, gameui.TextSmall, gameui.TextAlignStart)

		// Bölge sayısı ve başlangıç altını
		regionCount := len(gs.RegionsOwnedBy(fid))
		drawUILabel(screen, gameui.Rect{X: float64(x + 16), Y: float64(y + 54)}, itoa(regionCount)+" bölge", ColorGold, gameui.TextSmall, gameui.TextAlignStart)

		totalVictories, historicalVictories, generalVictories, featuredVictory := factionVictorySummary(gs, fid)
		victoryLine := itoa(totalVictories) + " zafer hedefi"
		if historicalVictories > 0 {
			victoryLine += "  |  " + itoa(historicalVictories) + " tarihsel"
		}
		if generalVictories > 0 {
			victoryLine += "  |  " + itoa(generalVictories) + " genel"
		}
		drawUILabel(screen, gameui.Rect{X: float64(x + 16), Y: float64(y + 74), W: float64(cardW - 32)}, trimTextToWidth(victoryLine, FaceSmall, float64(cardW-32)), color.RGBA{188, 176, 142, 235}, gameui.TextSmall, gameui.TextAlignStart)
		if featuredVictory != "" {
			drawUILabel(screen, gameui.Rect{X: float64(x + 16), Y: float64(y + 94), W: float64(cardW - 32)}, trimTextToWidth("Öne çıkan: "+featuredVictory, FaceSmall, float64(cardW-32)), color.RGBA{210, 188, 118, 235}, gameui.TextSmall, gameui.TextAlignStart)
		}
	}
}

func factionVictorySummary(gs *state.GameState, fid faction.FactionID) (total, historical, general int, featured string) {
	if gs == nil {
		return 0, 0, 0, ""
	}

	visible := scenario.FilterVictoryOptionsForFaction(gs.ScenarioVictories, string(fid))
	total = len(visible)
	for _, opt := range visible {
		if len(opt.AllowedFactions) > 0 {
			historical++
			if featured == "" {
				featured = opt.Title
			}
			continue
		}
		general++
		if featured == "" {
			featured = opt.Title
		}
	}
	return total, historical, general, featured
}

func selectableFactions(gs *state.GameState) ([]faction.FactionID, int) {
	if gs == nil {
		return nil, 0
	}

	orderedPlayable := make([]faction.FactionID, 0, len(gs.Factions))
	seen := make(map[faction.FactionID]struct{}, len(gs.Factions))
	for _, fid := range gs.FactionOrder {
		if f := gs.Factions[fid]; f != nil && f.IsPlayable {
			orderedPlayable = append(orderedPlayable, fid)
			seen[fid] = struct{}{}
		}
	}
	for fid, f := range gs.Factions {
		if !f.IsPlayable {
			continue
		}
		if _, ok := seen[fid]; ok {
			continue
		}
		orderedPlayable = append(orderedPlayable, fid)
	}

	var historicalFids []faction.FactionID
	var generalOnlyFids []faction.FactionID
	for _, fid := range orderedPlayable {
		_, historical, _, _ := factionVictorySummary(gs, fid)
		if historical > 0 {
			historicalFids = append(historicalFids, fid)
		} else {
			generalOnlyFids = append(generalOnlyFids, fid)
		}
	}

	ordered := make([]faction.FactionID, 0, len(historicalFids)+len(generalOnlyFids))
	ordered = append(ordered, historicalFids...)
	ordered = append(ordered, generalOnlyFids...)
	return ordered, len(historicalFids)
}

func factionCardRect(index, historicalCount, total, cols int, cardW, cardH, padX, padY, headerH float64) gameui.Rect {
	layout := factionGroupLayout(total, historicalCount, cols, cardW, cardH, padX, padY, headerH)
	if historicalCount > 0 && index < historicalCount {
		col := index % cols
		row := index / cols
		return gridCellRect(layout.historicalGrid, cardW, cardH, padX, padY, col, row)
	}
	generalIndex := index - historicalCount
	col := generalIndex % cols
	row := generalIndex / cols
	return gridCellRect(layout.generalGrid, cardW, cardH, padX, padY, col, row)
}

type factionSelectLayout struct {
	historicalGrid  gameui.Rect
	generalGrid     gameui.Rect
	historicalLabel gameui.Rect
	generalLabel    gameui.Rect
}

func factionGroupLayout(total, historicalCount, cols int, cardW, cardH, padX, padY, headerH float64) factionSelectLayout {
	generalCount := total - historicalCount
	historicalRows := 0
	if historicalCount > 0 {
		historicalRows = (historicalCount + cols - 1) / cols
	}
	generalRows := 0
	if generalCount > 0 {
		generalRows = (generalCount + cols - 1) / cols
	}

	gridW := cardW*float64(cols) + padX*float64(maxScreenInt(cols-1, 0))
	blockH := 0.0
	if historicalRows > 0 {
		blockH += factionGroupLabelH + 6
		blockH += cardH*float64(historicalRows) + padY*float64(maxScreenInt(historicalRows-1, 0))
	}
	if generalRows > 0 {
		if blockH > 0 {
			blockH += factionGroupGap
		}
		blockH += factionGroupLabelH + 6
		blockH += cardH*float64(generalRows) + padY*float64(maxScreenInt(generalRows-1, 0))
	}

	base := gameui.Rect{
		X: ScreenWidth/2 - gridW/2,
		Y: ScreenHeight/2 - (blockH+headerH)/2 + headerH,
		W: gridW,
	}
	layout := factionSelectLayout{}
	currentY := base.Y

	if historicalRows > 0 {
		layout.historicalLabel = gameui.Rect{X: base.X, Y: currentY, W: 320, H: factionGroupLabelH}
		currentY += factionGroupLabelH + 6
		layout.historicalGrid = gameui.Rect{
			X: base.X,
			Y: currentY,
			W: gridW,
			H: cardH*float64(historicalRows) + padY*float64(maxScreenInt(historicalRows-1, 0)),
		}
		currentY += layout.historicalGrid.H
	}

	if generalRows > 0 {
		if historicalRows > 0 {
			currentY += factionGroupGap
		}
		layout.generalLabel = gameui.Rect{X: base.X, Y: currentY, W: 320, H: factionGroupLabelH}
		currentY += factionGroupLabelH + 6
		layout.generalGrid = gameui.Rect{
			X: base.X,
			Y: currentY,
			W: gridW,
			H: cardH*float64(generalRows) + padY*float64(maxScreenInt(generalRows-1, 0)),
		}
	}

	return layout
}

func drawFactionGroupLabels(screen *ebiten.Image, total, historicalCount, cols int, cardW, cardH, padX, padY, headerH float64) {
	if total == 0 {
		return
	}
	layout := factionGroupLayout(total, historicalCount, cols, cardW, cardH, padX, padY, headerH)
	if historicalCount > 0 && layout.historicalLabel.W > 0 {
		drawUILabel(screen, layout.historicalLabel, "Tarihsel Hedefi Olanlar", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	}
	if historicalCount < total && layout.generalLabel.W > 0 {
		drawUILabel(screen, layout.generalLabel, "Genel Hedefliler", ColorGray, gameui.TextMedium, gameui.TextAlignStart)
	}
}
