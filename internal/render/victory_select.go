package render

import (
	"image/color"
	"strings"

	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const victoryGroupGap = 26.0
const victoryGroupLabelH = 20.0

func victoryCardDimensions() (float64, float64) {
	return 780.0, 126.0
}

func buildVictoryCardButtons(gs *state.GameState) []gameui.Button {
	opts, historicalCount := orderedVictoryOptions(gs)
	cardW, cardH := victoryCardDimensions()
	gap := 12.0
	headerH := 80.0
	buttons := make([]gameui.Button, 0, len(opts))
	for i, opt := range opts {
		r := victoryCardRect(i, len(opts), historicalCount, cardW, cardH, gap, headerH)
		buttons = append(buttons, gameui.NewButton(r.X, r.Y, r.W, r.H, opt.Title))
	}
	return buttons
}

func orderedVictoryOptions(gs *state.GameState) ([]scenario.VictoryOptionDef, int) {
	if gs == nil {
		return nil, 0
	}
	historical := make([]scenario.VictoryOptionDef, 0, len(gs.AvailableVictories))
	general := make([]scenario.VictoryOptionDef, 0, len(gs.AvailableVictories))
	for _, opt := range gs.AvailableVictories {
		if len(opt.AllowedFactions) > 0 {
			historical = append(historical, opt)
			continue
		}
		general = append(general, opt)
	}
	ordered := make([]scenario.VictoryOptionDef, 0, len(gs.AvailableVictories))
	ordered = append(ordered, historical...)
	ordered = append(ordered, general...)
	return ordered, len(historical)
}

type victorySelectLayout struct {
	historicalLabel gameui.Rect
	historicalStack gameui.Rect
	generalLabel    gameui.Rect
	generalStack    gameui.Rect
}

func victoryLayout(total, historicalCount int, cardW, cardH, gap, headerH float64) victorySelectLayout {
	generalCount := total - historicalCount
	totalH := 0.0
	if historicalCount > 0 {
		totalH += victoryGroupLabelH + 6
		totalH += float64(historicalCount)*cardH + float64(maxScreenInt(historicalCount-1, 0))*gap
	}
	if generalCount > 0 {
		if totalH > 0 {
			totalH += victoryGroupGap
		}
		totalH += victoryGroupLabelH + 6
		totalH += float64(generalCount)*cardH + float64(maxScreenInt(generalCount-1, 0))*gap
	}
	stackX := ScreenWidth/2 - cardW/2
	startY := ScreenHeight/2 - (totalH+headerH)/2 + headerH
	layout := victorySelectLayout{}
	currentY := startY
	if historicalCount > 0 {
		layout.historicalLabel = gameui.Rect{X: stackX, Y: currentY, W: cardW, H: victoryGroupLabelH}
		currentY += victoryGroupLabelH + 6
		stackH := float64(historicalCount)*cardH + float64(maxScreenInt(historicalCount-1, 0))*gap
		layout.historicalStack = gameui.Rect{X: stackX, Y: currentY, W: cardW, H: stackH}
		currentY += stackH
	}
	if generalCount > 0 {
		if historicalCount > 0 {
			currentY += victoryGroupGap
		}
		layout.generalLabel = gameui.Rect{X: stackX, Y: currentY, W: cardW, H: victoryGroupLabelH}
		currentY += victoryGroupLabelH + 6
		stackH := float64(generalCount)*cardH + float64(maxScreenInt(generalCount-1, 0))*gap
		layout.generalStack = gameui.Rect{X: stackX, Y: currentY, W: cardW, H: stackH}
	}
	return layout
}

func victoryCardRect(index, total, historicalCount int, cardW, cardH, gap, headerH float64) gameui.Rect {
	layout := victoryLayout(total, historicalCount, cardW, cardH, gap, headerH)
	if historicalCount > 0 && index < historicalCount {
		return stackItemRect(layout.historicalStack, cardH, gap, index)
	}
	return stackItemRect(layout.generalStack, cardH, gap, index-historicalCount)
}

func victoryAudienceBadge(opt scenario.VictoryOptionDef) (string, color.RGBA, color.RGBA) {
	if len(opt.AllowedFactions) > 0 {
		return "Tarihsel Hedef", color.RGBA{82, 58, 26, 235}, color.RGBA{206, 168, 90, 235}
	}
	return "Genel Hedef", color.RGBA{32, 46, 66, 225}, color.RGBA{110, 156, 214, 225}
}

func victorySelectSubtitle(gs *state.GameState) string {
	if gs == nil || gs.PlayerFactionID == "" {
		return "Nasıl kazanmak istiyorsun?"
	}
	if f := gs.Factions[gs.PlayerFactionID]; f != nil && f.NameTR != "" {
		return f.NameTR + " için zafer yolunu seç"
	}
	return "Nasıl kazanmak istiyorsun?"
}

func regionDisplayName(gs *state.GameState, regionID string) string {
	if gs != nil && gs.Regions != nil {
		if region, ok := gs.Regions[world.RegionID(regionID)]; ok && region != nil {
			if region.NameTR != "" {
				return region.NameTR
			}
			if region.Name != "" {
				return region.Name
			}
		}
	}
	return regionID
}

func regionDisplayNames(gs *state.GameState, regionTargets []string) []string {
	regionNames := make([]string, 0, len(regionTargets))
	for _, regionID := range regionTargets {
		regionNames = append(regionNames, regionDisplayName(gs, regionID))
	}
	return regionNames
}

func formatVictoryDeadline(year, month int) string {
	if year <= 0 {
		return ""
	}
	if month <= 0 || month > 12 {
		month = 12
	}
	return "Son tarih: " + itoa(year) + "/" + pad2(month)
}

func pad2(v int) string {
	if v < 10 {
		return "0" + itoa(v)
	}
	return itoa(v)
}

func victoryTargetSummary(gs *state.GameState, opt scenario.VictoryOptionDef) string {
	regionNames := regionDisplayNames(gs, opt.RegionTargets())
	deadline := formatVictoryDeadline(opt.DeadlineYear, opt.DeadlineMonth)

	appendDeadline := func(parts []string) string {
		if deadline != "" {
			parts = append(parts, deadline)
		}
		if len(parts) > 0 {
			return strings.Join(parts, "  |  ")
		}
		return opt.Detail
	}

	switch opt.Type {
	case "conquer_city":
		if len(regionNames) > 0 {
			return appendDeadline([]string{"Hedef bölgeler: " + strings.Join(regionNames, ", ")})
		}
	case "domination":
		parts := make([]string, 0, 2)
		if opt.TargetRegionCount > 0 {
			parts = append(parts, "Bölge hedefi: "+itoa(opt.TargetRegionCount))
		}
		if len(regionNames) > 0 {
			parts = append(parts, "Kilit merkezler: "+strings.Join(regionNames, ", "))
		}
		return appendDeadline(parts)
	case "economic":
		parts := make([]string, 0, 2)
		if opt.TargetGoldIncome > 0 {
			parts = append(parts, "Gelir: "+itoa(opt.TargetGoldIncome)+" altın")
		}
		if opt.GoldHoldTurns > 0 {
			parts = append(parts, "Koruma: "+itoa(opt.GoldHoldTurns)+" tur")
		}
		return appendDeadline(parts)
	case "military":
		parts := make([]string, 0, 2)
		if opt.TargetArmyStrength > 0 {
			parts = append(parts, "Ordu gücü: "+itoa(opt.TargetArmyStrength))
		}
		if opt.TargetDefeated > 0 {
			parts = append(parts, "Yenilgi: "+itoa(opt.TargetDefeated)+" fraksiyon")
		}
		return appendDeadline(parts)
	case "religious":
		if len(regionNames) > 0 {
			return appendDeadline([]string{"Kutsal bölgeler: " + strings.Join(regionNames, ", ")})
		}
	case "survive_turns":
		if opt.Turns > 0 {
			return appendDeadline([]string{"Hayatta kalma süresi: " + itoa(opt.Turns) + " tur"})
		}
	}

	if deadline != "" {
		return appendDeadline([]string{opt.Detail})
	}
	return opt.Detail
}

func currentVictoryOption(gs *state.GameState) (scenario.VictoryOptionDef, bool) {
	if gs == nil {
		return scenario.VictoryOptionDef{}, false
	}
	if gs.SelectedVictoryOptionID != "" {
		for _, opt := range gs.ScenarioVictories {
			if opt.ID == gs.SelectedVictoryOptionID {
				return opt, true
			}
		}
	}
	for _, opt := range gs.ScenarioVictories {
		if opt.Type == string(gs.Victory.Type) {
			return opt, true
		}
	}
	return scenario.VictoryOptionDef{}, false
}

func activeVictoryTargetSummary(gs *state.GameState) string {
	if gs == nil {
		return ""
	}
	regionTargets := make([]string, 0, len(gs.Victory.RequiredRegions))
	for _, regionID := range gs.Victory.RequiredRegions {
		regionTargets = append(regionTargets, string(regionID))
	}
	regionNames := regionDisplayNames(gs, regionTargets)

	switch gs.Victory.Type {
	case state.VictoryConquerCity:
		if len(regionNames) > 0 {
			return strings.Join(regionNames, ", ")
		}
	case state.VictoryDomination, "":
		if len(regionNames) > 0 {
			return "Kilit merkezler: " + strings.Join(regionNames, ", ")
		}
	case state.VictoryEconomic:
		if gs.Victory.GoldHoldTurns > 0 {
			return itoa(gs.Victory.GoldHoldTurns) + " tur koruma"
		}
		return "Gelir eşiğini koru"
	case state.VictoryMilitary:
		if gs.Victory.TargetDefeated > 0 {
			return itoa(gs.Victory.TargetDefeated) + " fraksiyon yenilgisi"
		}
		return "Ordu gücünü büyüt"
	case state.VictoryReligious:
		if len(regionNames) > 0 {
			return strings.Join(regionNames, ", ")
		}
	case state.VictorySurviveTurns:
		if gs.Victory.TargetTurns > 0 {
			return itoa(gs.Victory.TargetTurns) + " tur hayatta kal"
		}
	}
	return ""
}

// DrawVictorySelect zafer koşulu seçim ekranını çizer.
// Seçenekler gs.AvailableVictories'ten okunur — hardcode değil.
func DrawVictorySelect(screen *ebiten.Image, gs *state.GameState, cursor int) {
	opts, historicalCount := orderedVictoryOptions(gs)
	drawUIScreenChrome(screen, color.RGBA{10, 10, 20, 255}, "ZAFER KOŞULUNU SEÇ", victorySelectSubtitle(gs))

	cardW, cardH := victoryCardDimensions()
	gap := 12.0
	headerH := 80.0
	layout := victoryLayout(len(opts), historicalCount, cardW, cardH, gap, headerH)

	drawBackButton(screen)

	if len(opts) == 0 {
		drawUILabel(screen, gameui.Rect{X: 0, Y: ScreenHeight / 2, W: ScreenWidth}, "Bu senaryo için zafer koşulu tanımlanmamış.", ColorGray, gameui.TextMedium, gameui.TextAlignCenter)
		return
	}

	if historicalCount > 0 {
		drawUILabel(screen, layout.historicalLabel, "Tarihsel Hedefler", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	}
	if historicalCount < len(opts) {
		drawUILabel(screen, layout.generalLabel, "Genel Hedefler", ColorGray, gameui.TextMedium, gameui.TextAlignStart)
	}

	for i, opt := range opts {
		rect := victoryCardRect(i, len(opts), historicalCount, cardW, cardH, gap, headerH)
		y := rect.Y

		bg := color.RGBA{25, 25, 45, 220}
		border := color.RGBA{80, 80, 120, 200}
		if i == cursor {
			bg = color.RGBA{50, 45, 90, 240}
			border = color.RGBA{200, 160, 60, 255}
		}

		drawUICardRect(screen, rect, bg, border, 2)

		titleCol := ColorWhite
		if i == cursor {
			titleCol = ColorYellow
		}
		badgeLabel, badgeBG, badgeBorder := victoryAudienceBadge(opt)
		badgeW := MeasureText(badgeLabel, FaceSmall) + 18
		badgeRect := gameui.Rect{X: rect.X + rect.W - badgeW - 16, Y: y + 12, W: badgeW, H: 20}
		drawUICardRect(screen, badgeRect, badgeBG, badgeBorder, 1)
		drawUILabel(screen, gameui.Rect{X: badgeRect.X, Y: badgeRect.Y + 2, W: badgeRect.W}, badgeLabel, ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)

		titleMaxW := badgeRect.X - rect.X - 34
		title := trimTextToWidth(opt.Title, FaceMed, titleMaxW)
		drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 14, W: titleMaxW}, title, titleCol, gameui.TextLarge, gameui.TextAlignStart)
		drawUIWrappedLabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 42, W: rect.W - 36}, opt.Description, ColorGray, gameui.TextMedium, 19, 2)
		targetSummary := victoryTargetSummary(gs, opt)
		drawUIWrappedLabel(screen, gameui.Rect{X: rect.X + 18, Y: y + 84, W: rect.W - 36}, targetSummary, color.RGBA{140, 120, 80, 220}, gameui.TextSmall, 16, 2)
	}

	footerY := layout.generalStack.Y + layout.generalStack.H
	if historicalCount == len(opts) {
		footerY = layout.historicalStack.Y + layout.historicalStack.H
	}
	drawUILabel(screen, gameui.Rect{X: 0, Y: footerY + 20, W: ScreenWidth}, "Zafer koşulunu seçmek için tıkla", ColorGray, gameui.TextSmall, gameui.TextAlignCenter)
}

// handleVictorySelectInput zafer seçim ekranı girişini işler.
func (r *Renderer) handleVictorySelectInput(input gameui.InputState) InputAction {
	opts, _ := orderedVictoryOptions(r.gs)
	n := len(opts)
	if n == 0 {
		if r.keyJustPressed(ebiten.KeyEscape) {
			return InputAction{Kind: ActionBack}
		}
		return InputAction{}
	}
	buttons := buildVictoryCardButtons(r.gs)

	for i, btn := range buttons {
		if btn.HitTest(input.MouseX, input.MouseY) {
			r.factionCursor = i
			break
		}
	}

	if r.keyJustPressed(ebiten.KeyArrowDown) {
		r.factionCursor = (r.factionCursor + 1) % n
	}
	if r.keyJustPressed(ebiten.KeyArrowUp) {
		r.factionCursor = (r.factionCursor - 1 + n) % n
	}
	if r.keyJustPressed(ebiten.KeyTab) {
		next := focusButtonIndex(buttons, r.factionCursor, ebiten.IsKeyPressed(ebiten.KeyShift))
		if next >= 0 && next < n {
			r.factionCursor = next
		}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		return InputAction{Kind: ActionSelectVictory, BuildingID: opts[r.factionCursor].ID}
	}
	if input.LeftJustPressed {
		if buildBackButton().HandleInput(input) {
			r.factionCursor = 0
			return InputAction{Kind: ActionBack}
		}
		for i, btn := range buttons {
			if btn.HandleInput(input) {
				return InputAction{Kind: ActionSelectVictory, BuildingID: opts[i].ID}
			}
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.factionCursor = 0
		return InputAction{Kind: ActionBack}
	}
	return InputAction{}
}

// VictoryOptionByID gs.AvailableVictories içinden ID'ye göre seçenek bulur.
func VictoryOptionByID(gs *state.GameState, id string) (scenario.VictoryOptionDef, bool) {
	for _, v := range gs.AvailableVictories {
		if v.ID == id {
			return v, true
		}
	}
	return scenario.VictoryOptionDef{}, false
}
