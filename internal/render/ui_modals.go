package render

import gameui "mapp-game-go/internal/ui"

type eventDetailLayout struct {
	panelRect   gameui.Rect
	headerRect  gameui.Rect
	titleRect   gameui.Rect
	closeRect   gameui.Rect
	filtersRect gameui.Rect
	bodyRect    gameui.Rect
	listRect    gameui.Rect
	detailRect  gameui.Rect
}

func eventDetailHeaderRects() (gameui.Rect, gameui.Rect, gameui.Rect, gameui.Box) {
	modal := buildEventDetailModal()
	panelRect := modal.Panel.Rect
	box := gameui.BoxFromRect(panelRect).Inset(18)
	headerRect, rest := box.CutTop(28, 12)
	closeRect, titleBox := gameui.BoxFromRect(headerRect).CutRight(30, 12)
	return panelRect, titleBox.Rect, closeRect, rest
}

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
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 700, 420, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildEventDetailLayout() eventDetailLayout {
	panelRect, titleRect, closeRect, rest := eventDetailHeaderRects()
	filtersRect, bodyBox := rest.CutTop(28, 22)
	cols := bodyBox.SplitColumns(18, 0.36, 0.64)
	layout := eventDetailLayout{
		panelRect:   panelRect,
		headerRect:  gameui.Rect{X: titleRect.X, Y: titleRect.Y, W: titleRect.W + 12 + closeRect.W, H: titleRect.H},
		titleRect:   titleRect,
		closeRect:   closeRect,
		filtersRect: filtersRect,
		bodyRect:    bodyBox.Rect,
	}
	if len(cols) == 2 {
		layout.listRect = cols[0]
		layout.detailRect = cols[1]
	}
	return layout
}

func buildEventDetailCloseButton() gameui.Button {
	_, _, closeRect, _ := eventDetailHeaderRects()
	return gameui.NewButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H, "X")
}

func buildEventCodexFilterButtons() []gameui.Button {
	layout := buildEventDetailLayout()
	const (
		btnW = 92.0
		btnH = 28.0
		gap  = 10.0
	)
	labels := []string{"Tümü", "Hazır", "Takvim", "Kilitli"}
	buttons := make([]gameui.Button, 0, len(labels))
	startX := layout.filtersRect.X
	y := layout.filtersRect.Y
	for i, label := range labels {
		x := startX + float64(i)*(btnW+gap)
		buttons = append(buttons, gameui.NewButton(x, y, btnW, btnH, label))
	}
	return buttons
}

func buildHistoricalEventModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 760, 420, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildHistoricalEventChoiceButtons(count int) []gameui.Button {
	if count <= 0 {
		return nil
	}
	modal := buildHistoricalEventModal()
	const (
		btnW = 260.0
		btnH = 36.0
		gap  = 16.0
	)
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 24
	totalW := float64(count)*btnW + float64(max(0, count-1))*gap
	startX := modal.Panel.Rect.X + (modal.Panel.Rect.W-totalW)/2
	buttons := make([]gameui.Button, 0, count)
	for i := 0; i < count; i++ {
		x := startX + float64(i)*(btnW+gap)
		buttons = append(buttons, gameui.NewButton(x, btnY, btnW, btnH, ""))
	}
	return buttons
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
