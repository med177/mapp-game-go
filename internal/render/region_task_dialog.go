package render

import (
	"fmt"
	"image/color"

	"mapp-game-go/internal/army"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	regionTaskDialogW    = float32(660)
	regionTaskDialogH    = float32(220)
	regionTaskDialogBtnW = float32(140)
	regionTaskDialogBtnH = float32(42)
	regionTaskDialogGap  = float32(12)
)

func buildRegionTaskDialogModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, float64(regionTaskDialogW), float64(regionTaskDialogH), gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildRegionTaskDialogButtons(modal gameui.Modal, fortified bool) [4]gameui.Button {
	labels := [4]string{"Kuşatma", "Yağmala", "Pusu", "Vazgeç"}
	if !fortified {
		labels[0] = "Ele Geçir"
	}
	// Kuşatma ve pusu askerî harekâtı, yağma kaynak transferini,
	// vazgeç ise modalı kapatmayı temsil eder.
	icons := [4]gameui.IconID{gameui.IconSword, gameui.IconSell, gameui.IconSword, gameui.IconClose}
	totalW := regionTaskDialogBtnW*4 + regionTaskDialogGap*3
	startX := modal.Panel.Rect.X + (modal.Panel.Rect.W-float64(totalW))/2
	y := modal.Panel.Rect.Y + modal.Panel.Rect.H - float64(regionTaskDialogBtnH) - 20
	var buttons [4]gameui.Button
	for i := range buttons {
		x := startX + float64(i)*(float64(regionTaskDialogBtnW)+float64(regionTaskDialogGap))
		buttons[i] = gameui.NewButton(x, y, float64(regionTaskDialogBtnW), float64(regionTaskDialogBtnH), labels[i]).WithIcon(icons[i])
	}
	return buttons
}

func (r *Renderer) showRegionTaskDialog(attacker *army.Army, target *world.Region, fortified bool) {
	if r == nil || attacker == nil || target == nil {
		return
	}
	actions := [4]InputAction{
		{Kind: ActionStartSiege, ArmyID: attacker.ID, TargetRegion: target.ID},
		{Kind: ActionRaidRegion, ArmyID: attacker.ID, TargetRegion: target.ID},
		{Kind: ActionSetAmbush, ArmyID: attacker.ID, TargetRegion: target.ID},
		{},
	}
	if !fortified {
		actions[0].Kind = ActionCaptureRegion
	}
	modal := buildRegionTaskDialogModal()
	r.regionTaskDialog = regionTaskDialogState{
		show:    true,
		title:   "Bölge Görevi",
		message: fmt.Sprintf("%s üzerinde ordunun bu turdaki görevini seç.", target.NameTR),
		buttons: buildRegionTaskDialogButtons(modal, fortified),
		actions: actions,
	}
}

func regionTaskDialogButtonStyle(index int) gameui.ButtonStyle {
	switch index {
	case 0:
		return solidButtonStyle(color.RGBA{70, 140, 70, 240}, color.RGBA{120, 180, 120, 255}, ColorWhite, 10)
	case 1:
		return solidButtonStyle(color.RGBA{145, 95, 45, 235}, color.RGBA{190, 135, 75, 255}, ColorWhite, 10)
	case 2:
		return solidButtonStyle(color.RGBA{55, 92, 142, 240}, color.RGBA{112, 164, 202, 255}, ColorWhite, 10)
	default:
		return solidButtonStyle(color.RGBA{70, 70, 70, 220}, color.RGBA{120, 120, 120, 255}, ColorWhite, 10)
	}
}

func (r *Renderer) drawRegionTaskDialog(screen *ebiten.Image) {
	modal := buildRegionTaskDialogModal()
	gameui.DrawModal(screen, modal, standardModalStyle, nil, nil)
	drawUILabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 28}, r.regionTaskDialog.title, color.RGBA{255, 220, 100, 255}, gameui.TextLarge, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: modal.Panel.Rect.X + 24, Y: modal.Panel.Rect.Y + 62, W: modal.Panel.Rect.W - 48}, r.regionTaskDialog.message, color.RGBA{220, 220, 220, 255}, gameui.TextSmall, 17, 2)
	for i, btn := range r.regionTaskDialog.buttons {
		drawUIButtonWidget(screen, btn, regionTaskDialogButtonStyle(i))
	}
}

func (r *Renderer) regionTaskDialogHovering(fx, fy float64) bool {
	if r == nil || !r.regionTaskDialog.show {
		return false
	}
	for _, btn := range r.regionTaskDialog.buttons {
		if btn.Enabled && btn.HitTest(fx, fy) {
			return true
		}
	}
	return false
}

func (r *Renderer) handleRegionTaskDialogInput() InputAction {
	if r == nil || !r.regionTaskDialog.show {
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		for i, btn := range r.regionTaskDialog.buttons {
			if !btn.Enabled || !btn.HitTest(float64(mx), float64(my)) {
				continue
			}
			action := r.regionTaskDialog.actions[i]
			r.regionTaskDialog = regionTaskDialogState{}
			return action
		}
	}
	if r.keyJustPressed(ebiten.KeyEscape) || r.keyJustPressed(ebiten.KeyN) {
		r.regionTaskDialog = regionTaskDialogState{}
	}
	return InputAction{}
}
