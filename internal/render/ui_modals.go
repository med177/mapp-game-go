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

type eventCodexLayout struct {
	panelRect   gameui.Rect
	headerRect  gameui.Rect
	titleRect   gameui.Rect
	closeRect   gameui.Rect
	filtersRect gameui.Rect
	listRect    gameui.Rect
	detailRect  gameui.Rect
}

type victoryDetailLayout struct {
	panelRect  gameui.Rect
	headerRect gameui.Rect
	titleRect  gameui.Rect
	closeRect  gameui.Rect
	bodyRect   gameui.Rect
	scrollRect gameui.Rect
	scrollbar  gameui.Rect
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
	const dlgW, dlgH = 960.0, 560.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildWarConfirmButtons() (gameui.Button, gameui.Button) {
	const btnW, btnH = 176.0, 38.0
	modal := buildWarConfirmModal()
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 16
	yesX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - btnW - 10
	noX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 + 10
	return gameui.NewButton(yesX, btnY, btnW, btnH, "Savaş İlan Et").WithIcon(gameui.IconSword),
		gameui.NewButton(noX, btnY, btnW, btnH, "İptal").WithIcon(gameui.IconClose)
}

func buildBattlePlanModal() gameui.Modal {
	const dlgW, dlgH = 860.0, 483.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func battlePlanCardRects() [3]gameui.Rect {
	modal := buildBattlePlanModal()
	const (
		cardW = 252.0
		cardH = 301.0
		gap   = 18.0
		topY  = 114.0
	)
	totalW := cardW*3 + gap*2
	startX := modal.Panel.Rect.X + (modal.Panel.Rect.W-totalW)/2
	y := modal.Panel.Rect.Y + topY
	return [3]gameui.Rect{
		{X: startX, Y: y, W: cardW, H: cardH},
		{X: startX + cardW + gap, Y: y, W: cardW, H: cardH},
		{X: startX + (cardW+gap)*2, Y: y, W: cardW, H: cardH},
	}
}

func buildBattlePlanButtons() ([3]gameui.Button, gameui.Button) {
	const (
		btnW = 164.0
		btnH = 32.0
	)
	rects := battlePlanCardRects()
	var buttons [3]gameui.Button
	for i, rect := range rects {
		x := rect.X + (rect.W-btnW)/2
		y := rect.Y + rect.H - btnH - 10
		buttons[i] = gameui.NewButton(x, y, btnW, btnH, "")
	}
	modal := buildBattlePlanModal()
	cancelBtn := gameui.NewButton(modal.Panel.Rect.X+modal.Panel.Rect.W/2-70, modal.Panel.Rect.Y+modal.Panel.Rect.H-48, 140, 32, "İptal")
	return buttons, cancelBtn
}

func buildEventDetailModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 700, 420, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func eventCodexHeaderRects() (gameui.Rect, gameui.Rect, gameui.Rect, gameui.Box) {
	modal := buildEventCodexModal()
	panelRect := modal.Panel.Rect
	box := gameui.BoxFromRect(panelRect).Inset(20)
	headerRect, rest := box.CutTop(30, 14)
	closeRect, titleBox := gameui.BoxFromRect(headerRect).CutRight(30, 14)
	return panelRect, titleBox.Rect, closeRect, rest
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

func buildEventCodexModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 980, 620, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func victoryDetailHeaderRects() (gameui.Rect, gameui.Rect, gameui.Rect, gameui.Box) {
	modal := buildVictoryDetailModal()
	panelRect := modal.Panel.Rect
	box := gameui.BoxFromRect(panelRect).Inset(20)
	headerRect, rest := box.CutTop(30, 16)
	closeRect, titleBox := gameui.BoxFromRect(headerRect).CutRight(30, 12)
	return panelRect, titleBox.Rect, closeRect, rest
}

func buildVictoryDetailModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 760, 460, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildVictoryDetailLayout() victoryDetailLayout {
	panelRect, titleRect, closeRect, rest := victoryDetailHeaderRects()
	scrollRect := rest.Rect
	scrollbar := gameui.Rect{
		X: scrollRect.X + scrollRect.W - 10,
		Y: scrollRect.Y,
		W: 6,
		H: scrollRect.H,
	}
	return victoryDetailLayout{
		panelRect:  panelRect,
		headerRect: gameui.Rect{X: titleRect.X, Y: titleRect.Y, W: titleRect.W + 12 + closeRect.W, H: titleRect.H},
		titleRect:  titleRect,
		closeRect:  closeRect,
		bodyRect: gameui.Rect{
			X: scrollRect.X,
			Y: scrollRect.Y,
			W: scrollRect.W - 18,
			H: scrollRect.H,
		},
		scrollRect: scrollRect,
		scrollbar:  scrollbar,
	}
}

func buildEventCodexLayout() eventCodexLayout {
	panelRect, titleRect, closeRect, rest := eventCodexHeaderRects()
	filtersRect, bodyBox := rest.CutTop(30, 20)
	cols := bodyBox.SplitColumns(22, 0.38, 0.62)
	layout := eventCodexLayout{
		panelRect:   panelRect,
		headerRect:  gameui.Rect{X: titleRect.X, Y: titleRect.Y, W: titleRect.W + 14 + closeRect.W, H: titleRect.H},
		titleRect:   titleRect,
		closeRect:   closeRect,
		filtersRect: filtersRect,
	}
	if len(cols) == 2 {
		layout.listRect = cols[0]
		layout.detailRect = cols[1]
	}
	return layout
}

func buildEventDetailCloseButton() gameui.Button {
	_, _, closeRect, _ := eventDetailHeaderRects()
	btn := gameui.NewButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildEventCodexCloseButton() gameui.Button {
	_, _, closeRect, _ := eventCodexHeaderRects()
	btn := gameui.NewButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildVictoryDetailCloseButton() gameui.Button {
	_, _, closeRect, _ := victoryDetailHeaderRects()
	btn := gameui.NewButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H, "").WithIcon(gameui.IconClose)
	btn.IconSize = 13
	return btn
}

func buildEventCodexFilterButtons() []gameui.Button {
	layout := buildEventCodexLayout()
	const (
		btnW = 132.0
		btnH = 30.0
		gap  = 12.0
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
	const dlgW, dlgH = 760.0, 344.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildDiplomacyOfferButtons() (gameui.Button, gameui.Button) {
	const btnW, btnH = 120.0, 36.0
	modal := buildDiplomacyOfferModal()
	// Offer dialog'da sağ özet panelinden uzak kalmak için butonları sol blokta tut.
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 12
	acceptX := modal.Panel.Rect.X + 16
	rejectX := acceptX + btnW + 12
	return gameui.NewButton(acceptX, btnY, btnW, btnH, "Kabul Et").WithIcon(gameui.IconCheck),
		gameui.NewButton(rejectX, btnY, btnW, btnH, "Reddet").WithIcon(gameui.IconClose)
}
