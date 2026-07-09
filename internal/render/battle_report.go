package render

import (
	"fmt"
	"image/color"
	"strings"

	"mapp-game-go/internal/audio"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type BattleScene string

const (
	BattleSceneLand       BattleScene = "land"
	BattleSceneNaval      BattleScene = "naval"
	BattleSceneAmphibious BattleScene = "amphibious"
	BattleSceneSiege      BattleScene = "siege"
)

type BattleReportSide struct {
	Label          string
	Faction        string
	StrengthBefore int
	StrengthAfter  int
	UnitsBefore    int
	UnitsAfter     int
	UnitsLost      int
	HPBefore       int
	HPAfter        int
	HPDamage       int
}

type BattleReport struct {
	Scene         BattleScene
	RegionName    string
	Title         string
	Outcome       string
	OutcomeDetail string
	StanceLabel   string
	Attacker      BattleReportSide
	Defender      BattleReportSide
}

type battleReportState struct {
	show bool
	data BattleReport
}

type battleReportLayout struct {
	panelRect    gameui.Rect
	headerRect   gameui.Rect
	titleRect    gameui.Rect
	closeRect    gameui.Rect
	artRect      gameui.Rect
	summaryRect  gameui.Rect
	attackerRect gameui.Rect
	defenderRect gameui.Rect
	footerRect   gameui.Rect
}

var (
	battleReportArtCache = map[BattleScene]*ebiten.Image{}
	battleReportArtTried = map[BattleScene]bool{}
)

func battleReportSceneLabelTR(scene BattleScene) string {
	switch scene {
	case BattleSceneNaval:
		return "Deniz Muharebesi"
	case BattleSceneAmphibious:
		return "Çıkarma Muharebesi"
	case BattleSceneSiege:
		return "Kuşatma Hücumu"
	default:
		return "Kara Muharebesi"
	}
}

func battleReportDefaultTitleTR(scene BattleScene) string {
	return battleReportSceneLabelTR(scene) + " Raporu"
}

func battleReportSoundKey(scene BattleScene) string {
	switch scene {
	case BattleSceneNaval:
		return "battle_naval"
	case BattleSceneAmphibious:
		return "battle_amphibious"
	case BattleSceneSiege:
		return "battle_siege"
	default:
		return "battle_land"
	}
}

func battleReportImageCandidates(scene BattleScene) []string {
	sceneName := string(scene)
	if sceneName == "" {
		sceneName = string(BattleSceneLand)
	}
	return []string{
		fmt.Sprintf("assets/ui/battle_%s.png", sceneName),
		fmt.Sprintf("assets/ui/battle_%s.jpg", sceneName),
		fmt.Sprintf("assets/ui/battle_%s.jpeg", sceneName),
		"assets/ui/battle_generic.png",
		"assets/ui/battle_generic.jpg",
		"assets/ui/battle_generic.jpeg",
		fmt.Sprintf("assets/battles/%s.png", sceneName),
		fmt.Sprintf("assets/battles/%s.jpg", sceneName),
		fmt.Sprintf("assets/battles/%s.jpeg", sceneName),
		"assets/battles/generic.png",
		"assets/battles/generic.jpg",
		"assets/battles/generic.jpeg",
	}
}

func battleReportPrimaryImageHint(scene BattleScene) string {
	sceneName := string(scene)
	if sceneName == "" {
		sceneName = string(BattleSceneLand)
	}
	return "assets/ui/battle_" + sceneName + ".png"
}

func battleReportArt(scene BattleScene) *ebiten.Image {
	if img, ok := battleReportArtCache[scene]; ok {
		return img
	}
	if battleReportArtTried[scene] {
		return nil
	}
	battleReportArtTried[scene] = true
	for _, candidate := range battleReportImageCandidates(scene) {
		if img := tryLoadImage(candidate); img != nil {
			battleReportArtCache[scene] = img
			return img
		}
	}
	return nil
}

func (r *Renderer) ShowBattleReport(report BattleReport) {
	if r == nil {
		return
	}
	if report.Scene == "" {
		report.Scene = BattleSceneLand
	}
	if report.Title == "" {
		report.Title = battleReportDefaultTitleTR(report.Scene)
	}
	r.battleReport = battleReportState{
		show: true,
		data: report,
	}
	r.combatLogTimer = 0
	soundKey := battleReportSoundKey(report.Scene)
	if !audio.HasSound(soundKey) {
		soundKey = "combat"
	}
	audio.PlaySound(soundKey)
}

func (r *Renderer) HideBattleReport() {
	if r == nil {
		return
	}
	r.battleReport = battleReportState{}
}

func buildBattleReportModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 980, 568, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func battleReportHeaderRects() (gameui.Rect, gameui.Rect, gameui.Rect, gameui.Box) {
	modal := buildBattleReportModal()
	panelRect := modal.Panel.Rect
	box := gameui.BoxFromRect(panelRect).Inset(20)
	headerRect, rest := box.CutTop(42, 16)
	closeRect, titleBox := gameui.BoxFromRect(headerRect).CutRight(30, 12)
	return panelRect, titleBox.Rect, closeRect, rest
}

func buildBattleReportLayout() battleReportLayout {
	panelRect, titleRect, closeRect, rest := battleReportHeaderRects()
	topRow, rest := rest.CutTop(232, 18)
	topCols := gameui.BoxFromRect(topRow).SplitColumns(20, 0.40, 0.60)
	bottomRow, footerBox := rest.CutTop(176, 18)
	bottomCols := gameui.BoxFromRect(bottomRow).SplitColumns(18, 0.50, 0.50)
	layout := battleReportLayout{
		panelRect:  panelRect,
		headerRect: gameui.Rect{X: titleRect.X, Y: titleRect.Y, W: titleRect.W + 12 + closeRect.W, H: titleRect.H},
		titleRect:  titleRect,
		closeRect:  closeRect,
		footerRect: footerBox.Rect,
	}
	if len(topCols) == 2 {
		layout.artRect = topCols[0]
		layout.summaryRect = topCols[1]
	}
	if len(bottomCols) == 2 {
		layout.attackerRect = bottomCols[0]
		layout.defenderRect = bottomCols[1]
	}
	return layout
}

func buildBattleReportCloseButton() gameui.Button {
	_, _, closeRect, _ := battleReportHeaderRects()
	btn := gameui.NewButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildBattleReportContinueButton() gameui.Button {
	const (
		btnW = 184.0
		btnH = 34.0
	)
	modal := buildBattleReportModal()
	x := modal.Panel.Rect.X + (modal.Panel.Rect.W-btnW)/2
	y := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 20
	return gameui.NewButton(x, y, btnW, btnH, "Devam Et").WithIcon(gameui.IconCheck)
}

func buildBattleReportOkButton() gameui.Button {
	const (
		btnW = 184.0
		btnH = 34.0
	)
	modal := buildBattleReportModal()
	x := modal.Panel.Rect.X + (modal.Panel.Rect.W-btnW)/2
	y := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 20
	return gameui.NewButton(x, y, btnW, btnH, "Tamam").WithIcon(gameui.IconCheck)
}

func battleReportPopupHit(fx, fy float64) bool {
	return buildBattleReportModal().Panel.Rect.Hit(fx, fy)
}

func battleReportCloseHit(fx, fy float64) bool {
	return buildBattleReportCloseButton().HitTest(fx, fy)
}

func battleReportContinueHit(fx, fy float64) bool {
	return buildBattleReportContinueButton().HitTest(fx, fy)
}

func battleReportStrengthText(side BattleReportSide) string {
	if side.StrengthBefore == side.StrengthAfter {
		return itoa(side.StrengthAfter)
	}
	return itoa(side.StrengthBefore) + " -> " + itoa(side.StrengthAfter)
}

func battleReportUnitsText(side BattleReportSide) string {
	return fmt.Sprintf("%d -> %d  |  Kayıp %d", side.UnitsBefore, side.UnitsAfter, side.UnitsLost)
}

func battleReportHPText(side BattleReportSide) string {
	return fmt.Sprintf("%d -> %d  |  Hasar %d", side.HPBefore, side.HPAfter, side.HPDamage)
}

func drawBattleReportArt(screen *ebiten.Image, rect gameui.Rect, report BattleReport) {
	drawRoundedRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), 10, color.RGBA{24, 18, 12, 238})
	vector.StrokeRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), 1.5, color.RGBA{108, 86, 54, 255}, false)
	if img := battleReportArt(report.Scene); img != nil {
		bounds := img.Bounds()
		if bounds.Dx() > 0 && bounds.Dy() > 0 {
			scale := min(rect.W/float64(bounds.Dx()), rect.H/float64(bounds.Dy()))
			drawW := float64(bounds.Dx()) * scale
			drawH := float64(bounds.Dy()) * scale
			x := rect.X + (rect.W-drawW)/2
			y := rect.Y + (rect.H-drawH)/2
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(x, y)
			screen.DrawImage(img, op)
			return
		}
	}
	sceneLabel := battleReportSceneLabelTR(report.Scene)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 20, W: rect.W - 36}, sceneLabel, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignCenter)
	drawUIWrappedLabel(screen, gameui.Rect{X: rect.X + 20, Y: rect.Y + 74, W: rect.W - 40}, "Bu yüzeye sahne resmi ekleyebilirsin.", color.RGBA{220, 214, 202, 255}, gameui.TextMedium, 20, 2)
	drawUIWrappedLabel(screen, gameui.Rect{X: rect.X + 20, Y: rect.Y + 122, W: rect.W - 40}, "Önerilen dosya: "+battleReportPrimaryImageHint(report.Scene), color.RGBA{152, 186, 210, 255}, gameui.TextSmall, 17, 3)
}

func drawBattleReportSideCard(screen *ebiten.Image, rect gameui.Rect, accent color.RGBA, side BattleReportSide) {
	drawRoundedRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), 10, color.RGBA{22, 16, 12, 236})
	vector.StrokeRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), float32(rect.H), 1.5, accent, false)
	vector.FillRect(screen, float32(rect.X), float32(rect.Y), float32(rect.W), 4, accent, false)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 16, W: rect.W - 36}, side.Label, accent, gameui.TextLarge, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 44, W: rect.W - 36}, side.Faction, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 82, W: rect.W - 36}, "Güç: "+battleReportStrengthText(side), color.RGBA{228, 224, 214, 255}, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 114, W: rect.W - 36}, "Birim: "+battleReportUnitsText(side), color.RGBA{228, 224, 214, 255}, gameui.TextSmall, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: rect.X + 18, Y: rect.Y + 140, W: rect.W - 36}, "HP: "+battleReportHPText(side), color.RGBA{206, 198, 180, 255}, gameui.TextSmall, gameui.TextAlignStart)
}

func drawBattleReportPopup(screen *ebiten.Image, report BattleReport) {
	modal := buildBattleReportModal()
	gameui.DrawModal(screen, modal, eventDetailModalStyle, nil, nil)

	layout := buildBattleReportLayout()
	drawUIPanelTopBar(screen, layout.panelRect, 3, panelBorder)

	title := report.Title
	if title == "" {
		title = battleReportDefaultTitleTR(report.Scene)
	}
	DrawText(screen, title, layout.titleRect.X, layout.titleRect.Y+6, FaceLarge, ColorGold)

	metaParts := make([]string, 0, 3)
	if region := strings.TrimSpace(report.RegionName); region != "" {
		metaParts = append(metaParts, region)
	}
	if stance := strings.TrimSpace(report.StanceLabel); stance != "" {
		metaParts = append(metaParts, "Duruş: "+stance)
	}
	if outcome := strings.TrimSpace(report.Outcome); outcome != "" {
		metaParts = append(metaParts, outcome)
	}
	drawUILabel(screen, gameui.Rect{X: layout.titleRect.X, Y: layout.titleRect.Y + 30, W: layout.headerRect.W - 48}, strings.Join(metaParts, " | "), ColorGray, gameui.TextSmall, gameui.TextAlignStart)

	closeBtn := buildBattleReportCloseButton()
	drawUIButtonWidget(screen, closeBtn, tinyButtonStyle)

	drawBattleReportArt(screen, layout.artRect, report)

	drawRoundedRect(screen, float32(layout.summaryRect.X), float32(layout.summaryRect.Y), float32(layout.summaryRect.W), float32(layout.summaryRect.H), 10, color.RGBA{24, 19, 14, 236})
	vector.StrokeRect(screen, float32(layout.summaryRect.X), float32(layout.summaryRect.Y), float32(layout.summaryRect.W), float32(layout.summaryRect.H), 1.5, color.RGBA{130, 105, 55, 255}, false)
	drawUILabel(screen, gameui.Rect{X: layout.summaryRect.X + 18, Y: layout.summaryRect.Y + 18, W: layout.summaryRect.W - 36}, report.Outcome, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: layout.summaryRect.X + 18, Y: layout.summaryRect.Y + 56, W: layout.summaryRect.W - 36}, report.OutcomeDetail, color.RGBA{232, 226, 214, 255}, gameui.TextMedium, 20, 4)

	drawBattleReportSideCard(screen, layout.attackerRect, color.RGBA{216, 146, 82, 255}, report.Attacker)
	drawBattleReportSideCard(screen, layout.defenderRect, color.RGBA{126, 182, 126, 255}, report.Defender)

	okBtn := buildBattleReportOkButton()
	drawUIButtonWidget(screen, okBtn, solidButtonStyle(color.RGBA{76, 108, 66, 240}, color.RGBA{120, 156, 102, 255}, ColorWhite, 10))
}
