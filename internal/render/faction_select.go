package render

import (
	"image"
	"image/color"
	"path/filepath"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const factionGroupGap = 34.0
const factionGroupLabelH = 20.0
const factionGroupLabelPad = 6.0
const factionSelectViewportTop = 104.0
const factionSelectViewportBottom = 18.0
const factionSelectScrollStep = 72.0

const (
	factionCardFlagSize     = 94.0
	factionCardFlagRightGap = 14.0
	factionCardFlagTop      = 34.0
)

var (
	factionSelectBackgroundPath string
	factionSelectBackground     *ebiten.Image
)

func buildFactionCardButtons(gs *state.GameState, scroll float64) []gameui.Button {
	factions, historicalCount := selectableFactions(gs)
	cols := 3
	cardW := 350.0
	cardH := 138.0
	padX := 30.0
	padY := 12.0
	headerH := 70.0
	buttons := make([]gameui.Button, 0, len(factions))
	for i, fid := range factions {
		r := factionCardRectScrolled(i, historicalCount, len(factions), cols, cardW, cardH, padX, padY, headerH, scroll)
		label := ""
		if f := gs.Factions[fid]; f != nil {
			label = f.NameTR
		}
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, label))
	}
	return buttons
}

// DrawFactionSelect fraksiyon seçim ekranını çizer.
func DrawFactionSelect(screen *ebiten.Image, gs *state.GameState, cursor int, scroll float64) {
	if background := factionSelectBackgroundImage(gs); background != nil {
		drawUIImageCover(screen, background)
		// Senaryo görseli üzerindeki kart ve başlık metinlerini okunabilir tut.
		drawUIOverlay(screen, color.RGBA{0, 0, 0, 82})
		drawUIScreenChromeOverlay(screen, "MAPP — Devlet Seç", "Devlet seçmek için tıkla")
	} else {
		drawUIScreenChrome(screen, color.RGBA{10, 8, 5, 255}, "MAPP — Devlet Seç", "Devlet seçmek için tıkla")
	}

	factions, historicalCount := selectableFactions(gs)
	cols := 3
	cardW := float32(350)
	cardH := float32(138)
	headerH := 70.0

	drawBackButton(screen)

	layout := factionGroupLayoutScrolled(len(factions), historicalCount, cols, float64(cardW), float64(cardH), 30, 12, headerH, scroll)

	if layout.viewport.H <= 0 {
		return
	}
	body := screen.SubImage(image.Rect(int(layout.viewport.X), int(layout.viewport.Y), int(layout.viewport.X+layout.viewport.W), int(layout.viewport.Y+layout.viewport.H))).(*ebiten.Image)
	drawFactionGroupLabels(body, layout, len(factions), historicalCount)
	for i, fid := range factions {
		f := gs.Factions[fid]
		cell := factionCardRectScrolled(i, historicalCount, len(factions), cols, float64(cardW), float64(cardH), 30, 12, headerH, layout.scroll)
		if cell.Y+cell.H <= layout.viewport.Y || cell.Y >= layout.viewport.Y+layout.viewport.H {
			continue
		}
		x := float32(cell.X)
		y := float32(cell.Y)
		flagRect := factionCardFlagRect(cell)
		textW := flagRect.X - cell.X - 24

		fc := color.RGBA{f.Color[0], f.Color[1], f.Color[2], 255}
		bgCol := color.RGBA{22, 18, 12, 220}
		borderCol := color.RGBA{80, 65, 40, 200}
		if i == cursor {
			bgCol = color.RGBA{45, 36, 20, 240}
			borderCol = fc
		}

		drawUICardRect(body, cell, bgCol, borderCol, 2)

		// Renk şeridi
		drawUICardAccent(body, cell, 8, fc)

		// İsim
		nameCol := ColorWhite
		if i == cursor {
			nameCol = ColorYellow
		}
		drawUILabel(body, gameui.Rect{X: float64(x + 16), Y: float64(y + 12)}, f.NameTR, nameCol, gameui.TextLarge, gameui.TextAlignStart)
		drawFactionFlagBadge(body, f.ID, factionInitial(f.NameTR), flagRect.X, flagRect.Y, flagRect.W, fc, panelBorder)

		// Din
		drawUILabel(body, gameui.Rect{X: float64(x + 16), Y: float64(y + 36)}, gs.ActiveReligionRegistry().DisplayNameTR(f.Religion), ColorGray, gameui.TextSmall, gameui.TextAlignStart)

		// Bölge sayısı ve başlangıç altını
		regionCount := len(gs.RegionsVisibleTo(fid))
		drawUILabel(body, gameui.Rect{X: float64(x + 16), Y: float64(y + 54)}, itoa(regionCount)+" bölge", ColorGold, gameui.TextSmall, gameui.TextAlignStart)

		totalVictories, historicalVictories, generalVictories, featuredVictory := factionVictorySummary(gs, fid)
		victoryLine := itoa(totalVictories) + " zafer hedefi"
		if historicalVictories > 0 {
			victoryLine += "  |  " + itoa(historicalVictories) + " tarihsel"
		}
		if generalVictories > 0 {
			victoryLine += "  |  " + itoa(generalVictories) + " genel"
		}
		drawUILabel(body, gameui.Rect{X: float64(x + 16), Y: float64(y + 74), W: textW}, trimTextToWidth(victoryLine, FaceSmall, textW), color.RGBA{188, 176, 142, 235}, gameui.TextSmall, gameui.TextAlignStart)
		if featuredVictory != "" {
			drawUILabel(body, gameui.Rect{X: float64(x + 16), Y: float64(y + 94), W: textW}, trimTextToWidth("Öne çıkan: "+featuredVictory, FaceSmall, textW), color.RGBA{210, 188, 118, 235}, gameui.TextSmall, gameui.TextAlignStart)
		}
	}
	drawFactionSelectScrollbar(screen, layout)
}

func factionCardFlagRect(card gameui.Rect) gameui.Rect {
	return gameui.Rect{
		X: card.X + card.W - factionCardFlagSize - factionCardFlagRightGap,
		Y: card.Y + factionCardFlagTop,
		W: factionCardFlagSize,
		H: factionCardFlagSize,
	}
}

func factionSelectBackgroundImage(gs *state.GameState) *ebiten.Image {
	path := ""
	if gs != nil && gs.ScenarioPath != "" {
		path = filepath.Join(gs.ScenarioPath, "scenario_bg.png")
	}
	if path == factionSelectBackgroundPath {
		return factionSelectBackground
	}

	factionSelectBackgroundPath = path
	factionSelectBackground = nil
	if path != "" {
		factionSelectBackground = tryLoadImage(path)
	}
	return factionSelectBackground
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
	return factionCardRectScrolled(index, historicalCount, total, cols, cardW, cardH, padX, padY, headerH, 0)
}

func factionCardRectScrolled(index, historicalCount, total, cols int, cardW, cardH, padX, padY, headerH, scroll float64) gameui.Rect {
	layout := factionGroupLayoutScrolled(total, historicalCount, cols, cardW, cardH, padX, padY, headerH, scroll)
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
	viewport        gameui.Rect
	contentHeight   float64
	scroll          float64
}

func factionGroupLayout(total, historicalCount, cols int, cardW, cardH, padX, padY, headerH float64) factionSelectLayout {
	return factionGroupLayoutScrolled(total, historicalCount, cols, cardW, cardH, padX, padY, headerH, 0)
}

func factionGroupLayoutScrolled(total, historicalCount, cols int, cardW, cardH, padX, padY, headerH, scroll float64) factionSelectLayout {
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

	baseX := ScreenWidth/2 - gridW/2
	viewport := gameui.Rect{X: baseX, Y: factionSelectViewportTop, W: gridW, H: ScreenHeight - factionSelectViewportTop - factionSelectViewportBottom}
	maxScroll := maxFloat64Value(blockH - viewport.H)
	scroll = clampFactionSelectScroll(scroll, maxScroll)
	baseY := viewport.Y - scroll
	layout := factionSelectLayout{}
	layout.viewport = viewport
	layout.contentHeight = blockH
	layout.scroll = scroll
	currentY := baseY

	if historicalRows > 0 {
		layout.historicalLabel = gameui.Rect{X: baseX, Y: currentY, W: gridW, H: factionGroupLabelH}
		currentY += factionGroupLabelH + 6
		layout.historicalGrid = gameui.Rect{
			X: baseX,
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
		layout.generalLabel = gameui.Rect{X: baseX, Y: currentY, W: gridW, H: factionGroupLabelH}
		currentY += factionGroupLabelH + 6
		layout.generalGrid = gameui.Rect{
			X: baseX,
			Y: currentY,
			W: gridW,
			H: cardH*float64(generalRows) + padY*float64(maxScreenInt(generalRows-1, 0)),
		}
	}

	return layout
}

func drawFactionGroupLabels(screen *ebiten.Image, layout factionSelectLayout, total, historicalCount int) {
	if total == 0 {
		return
	}
	if historicalCount > 0 && layout.historicalLabel.W > 0 {
		drawFactionGroupLabelBackdrop(screen, layout.historicalLabel)
		drawUIOutlinedLabel(screen, layout.historicalLabel, "Tarihsel Hedefi Olan Devletler", ColorGold, ownerLabelOutlineColor(ColorGold), gameui.TextMedium, gameui.TextAlignCenter)
	}
	if historicalCount < total && layout.generalLabel.W > 0 {
		drawFactionGroupLabelBackdrop(screen, layout.generalLabel)
		drawUIOutlinedLabel(screen, layout.generalLabel, "Genel Hedefi Olan Devletler", ColorGold, ownerLabelOutlineColor(ColorGold), gameui.TextMedium, gameui.TextAlignCenter)
	}
}

func clampFactionSelectScroll(scroll, maxScroll float64) float64 {
	if scroll < 0 {
		return 0
	}
	if scroll > maxScroll {
		return maxScroll
	}
	return scroll
}

func drawFactionSelectScrollbar(screen *ebiten.Image, layout factionSelectLayout) {
	maxScroll := maxFloat64Value(layout.contentHeight - layout.viewport.H)
	if maxScroll <= 0 {
		return
	}
	track := gameui.Rect{X: layout.viewport.X + layout.viewport.W + 10, Y: layout.viewport.Y, W: 5, H: layout.viewport.H}
	thumbH := track.H * layout.viewport.H / layout.contentHeight
	if thumbH < 28 {
		thumbH = 28
	}
	thumbY := track.Y + (track.H-thumbH)*layout.scroll/maxScroll
	drawUICardRect(screen, track, color.RGBA{25, 25, 45, 220}, color.RGBA{80, 80, 120, 200}, 1)
	drawUICardRect(screen, gameui.Rect{X: track.X, Y: thumbY, W: track.W, H: thumbH}, color.RGBA{180, 150, 60, 230}, color.RGBA{220, 190, 100, 240}, 1)
}

func drawFactionGroupLabelBackdrop(screen *ebiten.Image, label gameui.Rect) {
	drawUIPanelRect(screen, gameui.Rect{
		X: label.X - factionGroupLabelPad,
		Y: label.Y - factionGroupLabelPad/2,
		W: label.W + factionGroupLabelPad*2,
		H: label.H + factionGroupLabelPad,
	}, color.RGBA{0, 0, 0, 125}, color.RGBA{}, 0)
}
