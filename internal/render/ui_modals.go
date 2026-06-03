package render

import gameui "mapp-game-go/internal/ui"

func buildConfirmDialogModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, float64(confirmDialogW), float64(confirmDialogH), gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildConfirmDialogButtons(state confirmDialogState) (gameui.Button, gameui.Button, gameui.Button, bool) {
	modal := buildConfirmDialogModal()
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - float64(confirmDialogBtnH) - 16
	if state.thirdLabel != "" {
		saveX, discardX, cancelX := confirmDialogThreeButtonXs(float32(modal.Panel.Rect.X))
		return gameui.NewButton(float64(saveX), btnY, float64(confirmDialogBtnW), float64(confirmDialogBtnH), state.acceptLabel),
			gameui.NewButton(float64(discardX), btnY, float64(confirmDialogBtnW), float64(confirmDialogBtnH), state.thirdLabel),
			gameui.NewButton(float64(cancelX), btnY, float64(confirmDialogBtnW), float64(confirmDialogBtnH), state.declineLabel),
			true
	}
	yesX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - float64(confirmDialogBtnW) - 10
	noX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 + 10
	return gameui.NewButton(yesX, btnY, float64(confirmDialogBtnW), float64(confirmDialogBtnH), state.acceptLabel),
		gameui.Button{},
		gameui.NewButton(noX, btnY, float64(confirmDialogBtnW), float64(confirmDialogBtnH), state.declineLabel),
		false
}

func buildWarConfirmModal() gameui.Modal {
	const dlgW, dlgH = 380.0, 130.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildWarConfirmButtons() (gameui.Button, gameui.Button) {
	const btnW, btnH = 110.0, 36.0
	modal := buildWarConfirmModal()
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 16
	yesX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - btnW - 10
	noX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 + 10
	return gameui.NewButton(yesX, btnY, btnW, btnH, "Savas Ilan Et"),
		gameui.NewButton(noX, btnY, btnW, btnH, "Hayir")
}

func buildEventDetailModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 620, 300, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildEventDetailCloseButton() gameui.Button {
	modal := buildEventDetailModal()
	return gameui.NewButton(modal.Panel.Rect.X+modal.Panel.Rect.W-30-12, modal.Panel.Rect.Y+10, 30, 26, "X")
}

func buildHistoricalEventModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 680, 240, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildDiplomacyOfferModal() gameui.Modal {
	const dlgW, dlgH = 520.0, 190.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildDiplomacyOfferButtons() (gameui.Button, gameui.Button) {
	const btnW, btnH = 120.0, 36.0
	modal := buildDiplomacyOfferModal()
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 16
	acceptX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - btnW - 12
	rejectX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 + 12
	return gameui.NewButton(acceptX, btnY, btnW, btnH, "Kabul Et"),
		gameui.NewButton(rejectX, btnY, btnW, btnH, "Reddet")
}
