package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type techCardViewModel struct {
	Title         string
	Summary       string
	CostText      string
	ProgressText  string
	Progress      float64
	NameColor     color.Color
	CostColor     color.Color
	CategoryColor color.RGBA
	IconColor     color.RGBA
	IconID        gameui.IconID
	IsActive      bool
	IsDone        bool
	IsUnlocked    bool
	CanResearch   bool
}

type techCardComponent struct {
	Rect  gameui.Rect
	Model techCardViewModel
}

func newTechCardComponent(node techNode, rect gameui.Rect, activeResearchID string, research faction.ResearchState) techCardComponent {
	isActive := activeResearchID == node.t.ID
	model := techCardViewModel{
		Title:         node.t.NameTR,
		Summary:       techEffectSummary(node.t),
		CostText:      fmt.Sprintf("%d altın  •  %d tur", node.t.GoldCost, node.t.TurnsRequired),
		NameColor:     techCardNameColor(node.unlocked, node.done),
		CostColor:     techTextCost,
		CategoryColor: techCardCategoryColor(node.t.Category, node.unlocked || node.done),
		IconColor:     techCardIconColor(node.t.Category, node.unlocked, node.done),
		IconID:        techCategoryIcon(node.t.Category),
		IsActive:      isActive,
		IsDone:        node.done,
		IsUnlocked:    node.unlocked,
		CanResearch:   node.unlocked && !node.done,
	}
	if model.IsDone {
		model.Summary = "✓ Tamamlandı"
	}
	if !node.unlocked && !node.done {
		model.CostColor = techTextCostLocked
	}
	if isActive {
		model.Progress, model.ProgressText = techCardProgress(node.t, research)
	}
	return techCardComponent{Rect: rect, Model: model}
}

func (c techCardComponent) HitTest(mx, my float64) bool {
	return c.Rect.Hit(mx, my)
}

func (c techCardComponent) Draw(screen *ebiten.Image) {
	c.drawFrame(screen)
	if c.Model.IsActive {
		c.drawGlow(screen)
	}
	c.drawContent(screen)
}

func (c techCardComponent) drawFrame(screen *ebiten.Image) {
	shadowRect := gameui.Rect{
		X: c.Rect.X + 3,
		Y: c.Rect.Y + 3,
		W: c.Rect.W,
		H: c.Rect.H,
	}
	drawRoundedRectF64(screen, shadowRect.X, shadowRect.Y, shadowRect.W, shadowRect.H, float64(techCardRadius), color.RGBA{0, 0, 0, 120})

	cardBg := techCardBgBase
	if c.Model.IsActive {
		cardBg = techCardBgActive
	} else if !c.Model.IsUnlocked {
		cardBg = techCardBgLocked
	}
	drawRoundedRectF64(screen, c.Rect.X, c.Rect.Y, c.Rect.W, c.Rect.H, float64(techCardRadius), cardBg)

	vector.FillRect(screen, float32(c.Rect.X+3), float32(c.Rect.Y+6), 5, float32(c.Rect.H-12), c.Model.CategoryColor, false)

	borderCol := techCardBorderGlow
	borderW := float32(1.2)
	switch {
	case c.Model.IsActive:
		borderCol = techCardBorderActive
		borderW = 2.5
	case c.Model.IsDone:
		borderCol = color.RGBA{140, 200, 140, 200}
		borderW = 2.0
	case c.Model.IsUnlocked:
		borderCol = color.RGBA{c.Model.CategoryColor.R, c.Model.CategoryColor.G, c.Model.CategoryColor.B, 180}
		borderW = 1.8
	}
	drawRoundedRectStrokeF64(screen, c.Rect.X, c.Rect.Y, c.Rect.W, c.Rect.H, float64(techCardRadius), float64(borderW), borderCol, cardBg)
}

func (c techCardComponent) drawGlow(screen *ebiten.Image) {
	for i := 0; i < 3; i++ {
		alpha := uint8(50 - i*15)
		glow := color.RGBA{techCardBorderActive.R, techCardBorderActive.G, techCardBorderActive.B, alpha}
		offset := float32(i + 1)
		drawRoundedRectStrokeF64(
			screen,
			c.Rect.X-float64(offset), c.Rect.Y-float64(offset),
			c.Rect.W+float64(offset*2), c.Rect.H+float64(offset*2),
			float64(techCardRadius+float32(i)), 1.5, glow, techCardBgActive,
		)
	}
}

func (c techCardComponent) drawContent(screen *ebiten.Image) {
	if c.Model.IconID != gameui.IconNone {
		gameui.DrawIcon(screen, c.Model.IconID, c.Rect.X+10, c.Rect.Y+5, 20, c.Model.IconColor)
	}

	nameOffsetX := 0.0
	if c.Model.IconID != gameui.IconNone {
		nameOffsetX = 19
	}
	nameRightInset := 12.0
	if c.Model.IsDone {
		nameRightInset = 44
	}
	nameRect := gameui.Rect{X: c.Rect.X + 12 + nameOffsetX, Y: c.Rect.Y + 10, W: c.Rect.W - 24 - nameOffsetX - nameRightInset}
	drawUIWrappedLabelAligned(screen, nameRect, c.Model.Title, c.Model.NameColor, gameui.TextMedium, 17, 2, gameui.TextAlignCenter)

	buffRect := gameui.Rect{X: c.Rect.X + 14, Y: c.Rect.Y + 48, W: c.Rect.W - 28}
	drawUIWrappedLabelAligned(screen, buffRect, c.Model.Summary, color.RGBA{210, 210, 225, 230}, gameui.TextSmall, 14, 3, gameui.TextAlignCenter)

	if c.Model.IsActive {
		c.drawProgress(screen)
	} else {
		costRect := gameui.Rect{X: c.Rect.X + 12, Y: c.Rect.Y + c.Rect.H - 20, W: c.Rect.W - 24}
		drawUILabel(screen, costRect, c.Model.CostText, c.Model.CostColor, gameui.TextSmall, gameui.TextAlignCenter)
	}

	if c.Model.IsDone {
		badgeW := 28.0
		badgeH := 20.0
		badgeX := c.Rect.X + c.Rect.W - badgeW - 6
		badgeY := c.Rect.Y + 6
		drawRoundedRectF64(screen, badgeX, badgeY, badgeW, badgeH, 4, techBadgeDoneBG)
		drawRoundedRectStrokeF64(screen, badgeX, badgeY, badgeW, badgeH, 4, 1, techBadgeDoneBorder, techBadgeDoneBG)
		gameui.DrawIcon(screen, gameui.IconCheck, badgeX+6, badgeY+2, 14, color.RGBA{200, 245, 200, 255})
	}

	if c.Model.IsActive {
		badgeW := 30.0
		badgeH := 20.0
		badgeX := c.Rect.X + 6
		badgeY := c.Rect.Y + c.Rect.H - badgeH - 6
		drawRoundedRectF64(screen, badgeX, badgeY, badgeW, badgeH, 4, techBadgeActiveBG)
		drawRoundedRectStrokeF64(screen, badgeX, badgeY, badgeW, badgeH, 4, 1, techBadgeActiveBorder, techBadgeActiveBG)
		gameui.DrawIcon(screen, gameui.IconPlay, badgeX+7, badgeY+2, 14, color.RGBA{255, 235, 180, 255})
	}
}

func (c techCardComponent) drawProgress(screen *ebiten.Image) {
	progressBarY := c.Rect.Y + c.Rect.H - 22
	progressBarH := 6.0
	barX := c.Rect.X + 12
	barW := c.Rect.W - 24
	vector.FillRect(screen, float32(barX), float32(progressBarY), float32(barW), float32(progressBarH), techProgressBarBG, false)
	if c.Model.Progress > 0 {
		fillW := float32(barW) * float32(c.Model.Progress)
		vector.FillRect(screen, float32(barX), float32(progressBarY), fillW, float32(progressBarH), techProgressBarFill, false)
	}
	vector.StrokeRect(screen, float32(barX), float32(progressBarY), float32(barW), float32(progressBarH), 1, color.RGBA{130, 110, 50, 180}, false)
	drawUILabel(screen, gameui.Rect{X: barX, Y: progressBarY - 15, W: barW}, c.Model.ProgressText,
		color.RGBA{255, 210, 60, 240}, gameui.TextSmall, gameui.TextAlignCenter)
}

func techCardCategoryColor(category tech.Category, available bool) color.RGBA {
	catColor := techCategoryColors[category]
	if !available {
		return color.RGBA{80, 80, 90, 200}
	}
	return catColor
}

func techCardIconColor(category tech.Category, unlocked, done bool) color.RGBA {
	iconColor := techCategoryIconColors[category]
	if !unlocked && !done {
		return color.RGBA{90, 90, 100, 180}
	}
	return iconColor
}

func techCardNameColor(unlocked, done bool) color.Color {
	if unlocked || done {
		return techTextUnlocked
	}
	return techTextLocked
}

func techCardProgress(t *tech.Technology, research faction.ResearchState) (progress float64, label string) {
	if research.TurnsLeft <= 0 {
		return 0, ""
	}
	totalTurns := float64(t.TurnsRequired)
	if totalTurns <= 0 {
		totalTurns = float64(research.TurnsLeft * 2)
	}
	progress = 1.0 - float64(research.TurnsLeft)/totalTurns
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	return progress, fmt.Sprintf("%d tur kaldı", research.TurnsLeft)
}
