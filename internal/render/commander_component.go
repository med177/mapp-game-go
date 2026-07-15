package render

import (
	"fmt"
	"image/color"
	"strings"

	"mapp-game-go/internal/army"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type commanderTraitBadgeStyle struct {
	Label  string
	BG     color.RGBA
	Border color.RGBA
	Text   color.RGBA
}

type commanderEffectLine struct {
	Text  string
	Color color.RGBA
}

type commanderTraitBadgeOptions struct {
	MaxRows int
}

type commanderCardOptions struct {
	Role            string
	EmptySummary    string
	EmptyEffectText string
	ExtraLine       string
	ShowEffectText  bool
	MaxTraitRows    int
	BottomInset     float64
}

type commanderCompactStripOptions struct {
	PortraitAsset      string
	Name               string
	EmptyName          string
	BattleEffects      string
	EmptyBattleEffects string
	OperationalEffects string
	EmptyOperational   string
}

func commanderSummaryHeaderTexts(role string, commander *army.Commander) (topLabel, rightLabel string) {
	role = strings.TrimSpace(role)
	if commander == nil {
		if role == "" {
			role = "Komutan"
		}
		return role, "Atanmadı"
	}
	name := strings.TrimSpace(commander.Name)
	if name == "" {
		name = role
	}
	if role == "" {
		role = "Komutan"
	}
	return role, name
}

func commanderTraitsSummary(commander *army.Commander) string {
	if commander == nil || len(commander.Traits) == 0 {
		return "Yok"
	}
	traits := make([]string, 0, len(commander.Traits))
	for _, trait := range commander.Traits {
		traits = append(traits, army.TraitLabelTR(trait))
	}
	return strings.Join(traits, ", ")
}

func commanderTraitBadge(trait army.CommanderTrait) commanderTraitBadgeStyle {
	switch trait {
	case army.CommanderTraitVeteran:
		return commanderTraitBadgeStyle{
			Label:  "Tecrübe",
			BG:     color.RGBA{70, 48, 16, 230},
			Border: color.RGBA{216, 168, 72, 255},
			Text:   color.RGBA{255, 236, 176, 255},
		}
	case army.CommanderTraitTactician:
		return commanderTraitBadgeStyle{
			Label:  "Taktik",
			BG:     color.RGBA{32, 58, 88, 230},
			Border: color.RGBA{104, 168, 228, 255},
			Text:   color.RGBA{218, 236, 255, 255},
		}
	case army.CommanderTraitDefender:
		return commanderTraitBadgeStyle{
			Label:  "Savunma",
			BG:     color.RGBA{34, 74, 48, 230},
			Border: color.RGBA{108, 196, 136, 255},
			Text:   color.RGBA{222, 250, 226, 255},
		}
	case army.CommanderTraitAggressor:
		return commanderTraitBadgeStyle{
			Label:  "Saldırı",
			BG:     color.RGBA{96, 42, 30, 235},
			Border: color.RGBA{232, 126, 86, 255},
			Text:   color.RGBA{255, 226, 206, 255},
		}
	default:
		return commanderTraitBadgeStyle{
			Label:  army.TraitLabelTR(trait),
			BG:     color.RGBA{58, 58, 58, 230},
			Border: color.RGBA{146, 146, 146, 255},
			Text:   color.RGBA{238, 238, 238, 255},
		}
	}
}

func commanderOverflowBadge(hiddenCount int) commanderTraitBadgeStyle {
	return commanderTraitBadgeStyle{
		Label:  fmt.Sprintf("+%d", hiddenCount),
		BG:     color.RGBA{46, 46, 46, 220},
		Border: color.RGBA{124, 124, 124, 240},
		Text:   color.RGBA{232, 232, 232, 255},
	}
}

func commanderCardEffectLines(commander *army.Commander) []commanderEffectLine {
	if commander == nil {
		return nil
	}
	lines := make([]commanderEffectLine, 0, 5)
	if attack := commander.AttackModifier(); attack > 0 {
		lines = append(lines, commanderEffectLine{
			Text:  fmt.Sprintf("Saldırı +%.0f%%", attack*100),
			Color: color.RGBA{155, 210, 150, 255},
		})
	}
	if defense := commander.DefenseModifier(); defense > 0 {
		lines = append(lines, commanderEffectLine{
			Text:  fmt.Sprintf("Savunma +%.0f%%", defense*100),
			Color: color.RGBA{155, 210, 150, 255},
		})
	}
	if morale := commander.MoraleModifier(); morale > 0 {
		lines = append(lines, commanderEffectLine{
			Text:  fmt.Sprintf("Moral +%.0f%%", morale*100),
			Color: color.RGBA{145, 185, 220, 255},
		})
	}
	if move := commander.MoveBonus(); move > 0 {
		lines = append(lines, commanderEffectLine{
			Text:  fmt.Sprintf("Hareket +%d", move),
			Color: color.RGBA{145, 185, 220, 255},
		})
	}
	progress, breach := commander.SiegeBonuses()
	if progress > 0 || breach > 0 {
		lines = append(lines, commanderEffectLine{
			Text:  fmt.Sprintf("Kuşatma +%d/+%d", progress, breach),
			Color: color.RGBA{145, 185, 220, 255},
		})
	}
	return lines
}

func commanderSummaryDividerY(portraitY float64, effectLineCount int) float64 {
	infoY := portraitY + float64(armyPanelCommanderPortrait) + 14
	if effectLineCount <= 0 {
		return infoY
	}
	lastEffectY := portraitY + 64 + float64(effectLineCount-1)*12
	minDividerY := lastEffectY + 16
	if minDividerY > infoY {
		return minDividerY
	}
	return infoY
}

func drawCommanderTraitBadge(screen *ebiten.Image, style commanderTraitBadgeStyle, x, y, w, h float64) {
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), style.BG, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, style.Border, false)
	DrawTextCentered(screen, style.Label, x+w/2, y+2, FaceSmall, style.Text)
}

func drawCommanderTraitBadges(screen *ebiten.Image, commander *army.Commander, x, y, maxWidth float64, opts commanderTraitBadgeOptions) float64 {
	const (
		badgeH = 18.0
		gapX   = 6.0
		gapY   = 6.0
		padX   = 8.0
	)
	maxRows := opts.MaxRows
	if maxRows <= 0 {
		maxRows = 99
	}
	maxX := x + maxWidth

	badgeWidth := func(label string) float64 {
		w := MeasureText(label, FaceSmall) + padX*2
		if w > maxWidth {
			return maxWidth
		}
		return w
	}

	renderOverflow := func(cursorX, cursorY float64, hiddenCount int) float64 {
		if hiddenCount <= 0 {
			return cursorY + badgeH
		}
		style := commanderOverflowBadge(hiddenCount)
		w := badgeWidth(style.Label)
		if cursorX > x && cursorX+w > maxX {
			cursorX = x
			cursorY += badgeH + gapY
		}
		drawCommanderTraitBadge(screen, style, cursorX, cursorY, w, badgeH)
		return cursorY + badgeH
	}

	if commander == nil || len(commander.Traits) == 0 {
		style := commanderTraitBadgeStyle{
			Label:  "Özellik yok",
			BG:     color.RGBA{42, 42, 42, 220},
			Border: color.RGBA{108, 108, 108, 240},
			Text:   color.RGBA{216, 216, 216, 255},
		}
		drawCommanderTraitBadge(screen, style, x, y, badgeWidth(style.Label), badgeH)
		return y + badgeH
	}

	cursorX := x
	cursorY := y
	row := 1
	for i, trait := range commander.Traits {
		style := commanderTraitBadge(trait)
		w := badgeWidth(style.Label)
		if cursorX > x && cursorX+w > maxX {
			if row >= maxRows {
				return renderOverflow(cursorX, cursorY, len(commander.Traits)-i)
			}
			cursorX = x
			cursorY += badgeH + gapY
			row++
		}
		drawCommanderTraitBadge(screen, style, cursorX, cursorY, w, badgeH)
		cursorX += w + gapX
	}
	return cursorY + badgeH
}

func drawCommanderSummaryCard(screen *ebiten.Image, commander *army.Commander, x, y, w, h float64, opts commanderCardOptions) float64 {
	vector.FillRect(screen, float32(x), float32(y), float32(w), float32(h), color.RGBA{20, 16, 10, 215}, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{90, 72, 38, 220}, false)

	titleX := x + float64(armyPanelCommanderCardPad)
	titleY := y + float64(armyPanelCommanderSectionY)
	headerTop, headerRight := commanderSummaryHeaderTexts(opts.Role, commander)

	portraitX := x + float64(armyPanelCommanderCardPad)
	portraitY := y + 28
	drawCommanderPortrait(screen, commander, portraitX, portraitY, float64(armyPanelCommanderPortrait), float64(armyPanelCommanderPortrait))

	textX := portraitX + float64(armyPanelCommanderPortrait) + 10
	textW := w - (textX - x) - float64(armyPanelCommanderCardPad)
	if textW < 40 {
		textW = 40
	}
	if commander == nil {
		DrawText(screen, trimTextToWidth(headerRight, FaceMed, w-float64(armyPanelCommanderCardPad*2)), titleX, titleY, FaceMed, ColorWhite)
		DrawText(screen, trimTextToWidth(headerTop, FaceSmall, textW), textX, portraitY+6, FaceSmall, ColorGold)
		DrawText(screen, trimTextToWidth(opts.EmptySummary, FaceSmall, textW), textX, portraitY+34, FaceSmall, ColorGray)
		if opts.EmptyEffectText != "" {
			DrawText(screen, trimTextToWidth(opts.EmptyEffectText, FaceSmall, textW), textX, portraitY+54, FaceSmall, color.RGBA{145, 185, 220, 255})
		}
	} else {
		DrawText(screen, trimTextToWidth(headerRight, FaceMed, w-float64(armyPanelCommanderCardPad*2)), titleX, titleY, FaceMed, ColorWhite)
		DrawText(screen, trimTextToWidth(headerTop, FaceSmall, textW), textX, portraitY+6, FaceSmall, ColorGold)
		DrawText(screen, fmt.Sprintf("Seviye %d  |  %d XP", commander.Level, commander.Experience), textX, portraitY+28, FaceSmall, ColorGray)
		DrawText(screen, fmt.Sprintf("Savaş %d  |  Zafer %d", commander.Battles, commander.Victories), textX, portraitY+48, FaceSmall, ColorGray)
		effectLines := commanderCardEffectLines(commander)
		for i, line := range effectLines {
			DrawText(screen, trimTextToWidth(line.Text, FaceTiny, textW), textX, portraitY+64+float64(i)*12, FaceTiny, line.Color)
		}
	}

	infoY := commanderSummaryDividerY(portraitY, len(commanderCardEffectLines(commander)))
	vector.StrokeLine(screen, float32(x)+armyPanelCommanderCardPad, float32(infoY), float32(x+w)-armyPanelCommanderCardPad, float32(infoY), 1, color.RGBA{72, 56, 30, 180}, false)
	infoTextX := x + float64(armyPanelCommanderCardPad)
	infoTextW := w - float64(armyPanelCommanderCardPad*2)
	contentBottomY := y + h - opts.BottomInset
	if contentBottomY <= infoY+24 {
		contentBottomY = y + h
	}
	DrawText(screen, "Uzmanlıklar", infoTextX, infoY+8, FaceSmall, ColorGold)
	badgesBottomY := drawCommanderTraitBadges(screen, commander, infoTextX, infoY+24, infoTextW, commanderTraitBadgeOptions{MaxRows: opts.MaxTraitRows})
	nextTextY := badgesBottomY + 8
	if opts.ShowEffectText && nextTextY+14 <= contentBottomY {
		if commander == nil {
			DrawText(screen, trimTextToWidth(opts.EmptyEffectText, FaceSmall, infoTextW), infoTextX, nextTextY, FaceSmall, ColorGray)
		} else {
			DrawText(screen, trimTextToWidth(commanderEffectSummary(commander), FaceSmall, infoTextW), infoTextX, nextTextY, FaceSmall, color.RGBA{190, 182, 150, 235})
		}
		nextTextY += 18
	}
	if opts.ExtraLine != "" && nextTextY+14 <= contentBottomY {
		DrawText(screen, trimTextToWidth(opts.ExtraLine, FaceSmall, infoTextW), infoTextX, nextTextY, FaceSmall, color.RGBA{190, 172, 126, 235})
	}
	return badgesBottomY
}

func drawCommanderCompactStrip(screen *ebiten.Image, x, y, w, h float64, opts commanderCompactStripOptions) {
	drawRoundedRect(screen, float32(x), float32(y), float32(w), float32(h), 8, color.RGBA{28, 22, 16, 232})
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{90, 72, 42, 255}, false)

	portraitX := x + 8
	portraitY := y + 4
	portraitW := 44.0
	portraitH := 54.0
	drawCommanderPortraitAsset(screen, opts.PortraitAsset, portraitX, portraitY, portraitW, portraitH)

	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = opts.EmptyName
	}
	battleEffects := strings.TrimSpace(opts.BattleEffects)
	if battleEffects == "" {
		battleEffects = opts.EmptyBattleEffects
	}
	operationalEffects := strings.TrimSpace(opts.OperationalEffects)
	if operationalEffects == "" {
		operationalEffects = opts.EmptyOperational
	}

	textX := portraitX + portraitW + 10
	textW := x + w - textX - 8
	drawUILabel(screen, gameui.Rect{X: textX, Y: y + 8, W: textW}, trimTextToWidth(name, FaceSmall, textW), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: textX, Y: y + 26, W: textW}, trimTextToWidth(battleEffects, FaceSmall, textW), color.RGBA{155, 210, 150, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: textX, Y: y + 44, W: textW}, trimTextToWidth(operationalEffects, FaceSmall, textW), color.RGBA{145, 185, 220, 255}, gameui.TextSmall, gameui.TextAlignStart)
}
