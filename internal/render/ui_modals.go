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
	return buildConfirmDialogModalFor(confirmDialogState{})
}

func buildConfirmDialogModalFor(state confirmDialogState) gameui.Modal {
	modalW := float64(confirmDialogW)
	modalH := float64(confirmDialogH)
	if state.secondLabel != "" && state.fourthLabel != "" {
		modalW = float64(confirmDialogFourChoiceW)
	}
	if state.spacious {
		modalW = float64(confirmDialogSpaciousW)
		modalH = float64(confirmDialogSpaciousH)
	}
	anchor := gameui.AnchorMiddle
	if state.navalContact != nil {
		modalW = float64(navalContactDialogW)
		modalH = float64(navalContactDialogH)
		anchor = gameui.AnchorTop
	}
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, modalW, modalH, gameui.AnchorCenter, anchor, 0, 0)
	if state.navalContact != nil {
		// Üst HUD/"Hamleler" kartının altında kalır; temas denizi modalın
		// altında görünür ve merkezdeki map marker'ı kapatılmaz.
		_, hudY, _, hudH := turnTechHudRect()
		// `drawAITurnOverlay` HAMLELER kartı HUD'ın altında 40 px boşlukla
		// başlar ve 180 px sürer; temas modalı bu kartın altına yerleşir.
		y := float64(hudY + hudH + 40 + 180 + 12)
		if y+modalH > ScreenHeight-24 {
			y = ScreenHeight - modalH - 24
		}
		rect.Y = y
	}
	if state.landContact != nil {
		modalW = float64(landContactDialogW)
		modalH = float64(landContactDialogH)
		_, hudY, _, _ := bottomActionHudRect()
		rect = gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, modalW, modalH, gameui.AnchorCenter, gameui.AnchorTop, 0, float64(hudY)-20-modalH)
	}
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildConfirmDialogButtons(state confirmDialogState) (gameui.Button, gameui.Button, gameui.Button, bool) {
	modal := buildConfirmDialogModalFor(state)
	btnW := float64(confirmDialogBtnW)
	if state.spacious {
		btnW = float64(confirmDialogSpaciousBtnW)
	}
	if state.declineAction.Kind == ActionLiftSiege {
		btnW = 160
	}
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - float64(confirmDialogBtnH) - 18
	if state.thirdLabel != "" {
		saveX, discardX, cancelX := confirmDialogThreeButtonXs(float32(modal.Panel.Rect.X), btnW, modal.Panel.Rect.W)
		third := gameui.NewButton(float64(discardX), btnY, btnW, float64(confirmDialogBtnH), state.thirdLabel)
		third.Enabled = !state.thirdDisabled
		decline := gameui.NewButton(float64(cancelX), btnY, btnW, float64(confirmDialogBtnH), state.declineLabel)
		decline.Enabled = !state.declineDisabled
		return gameui.NewButton(float64(saveX), btnY, btnW, float64(confirmDialogBtnH), state.acceptLabel),
			third,
			decline,
			true
	}
	if state.messageOnly {
		okX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - btnW/2
		return gameui.NewButton(okX, btnY, btnW, float64(confirmDialogBtnH), state.acceptLabel),
			gameui.Button{}, gameui.Button{}, false
	}
	yesX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 - btnW - 10
	noX := modal.Panel.Rect.X + modal.Panel.Rect.W/2 + 10
	return gameui.NewButton(yesX, btnY, btnW, float64(confirmDialogBtnH), state.acceptLabel),
		gameui.Button{},
		gameui.NewButton(noX, btnY, btnW, float64(confirmDialogBtnH), state.declineLabel),
		false
}

func buildConfirmDialogFourButtons(state confirmDialogState) [4]gameui.Button {
	modal := buildConfirmDialogModalFor(state)
	btnW := (modal.Panel.Rect.W - 48) / 4
	gap := 8.0
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - float64(confirmDialogBtnH) - 18
	startX := modal.Panel.Rect.X + (modal.Panel.Rect.W-(btnW*4+gap*3))/2
	labels := [4]string{state.acceptLabel, state.secondLabel, state.thirdLabel, state.fourthLabel}
	buttons := [4]gameui.Button{}
	for i, label := range labels {
		buttons[i] = gameui.NewButton(startX+float64(i)*(btnW+gap), btnY, btnW, float64(confirmDialogBtnH), label)
	}
	buttons[1].Enabled = !state.secondDisabled
	buttons[2].Enabled = !state.thirdDisabled
	buttons[3].Enabled = !state.fourthDisabled
	return buttons
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
		topY  = 126.0
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
	filtersRect, bodyBox := rest.CutTop(42, 12)
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
	return gameui.NewCloseButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H)
}

func buildEventCodexCloseButton() gameui.Button {
	_, _, closeRect, _ := eventCodexHeaderRects()
	return gameui.NewCloseButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H)
}

func buildVictoryDetailCloseButton() gameui.Button {
	_, _, closeRect, _ := victoryDetailHeaderRects()
	return gameui.NewCloseButton(closeRect.X, closeRect.Y, closeRect.W, closeRect.H)
}

func buildEventCodexFilterButtons() []gameui.Button {
	layout := buildEventCodexLayout()
	const (
		btnW = 118.0
		btnH = 30.0
		gap  = 8.0
	)
	labels := []string{"Tümü", "Hazır", "Takvim", "Kilitli", "Oyuncu", "Gerçekleşen"}
	buttons := make([]gameui.Button, 0, len(labels))
	startX := layout.filtersRect.X
	y := layout.filtersRect.Y
	for i, label := range labels {
		x := startX + float64(i)*(btnW+gap)
		buttons = append(buttons, gameui.NewButton(x, y, btnW, btnH, label))
	}
	return buttons
}

type historicalEventLayout struct {
	modal      gameui.Modal
	descRect   gameui.Rect
	promptRect gameui.Rect
	infoRects  []gameui.Rect
	buttons    []gameui.Button
}

func buildHistoricalEventLayout(title, desc, prompt string, choices []HistoricalEventChoice) historicalEventLayout {
	const (
		minW       = 760.0
		maxW       = 1100.0
		horizontal = 30.0
		choiceGap  = 16.0
		buttonH    = 44.0
	)

	panelW := maxF(minW, minF(maxW, ScreenWidth-80))
	contentW := panelW - horizontal*2
	titleLines := gameui.WrappedLineCount(renderText, title, contentW, gameui.TextLarge)
	if titleLines < 1 {
		titleLines = 1
	}
	descLines := gameui.WrappedLineCount(renderText, desc, contentW, gameui.TextMedium)
	if descLines < 1 {
		descLines = 1
	}

	contentY := 28.0 + 26.0 + float64(titleLines)*30.0 + 12.0 + float64(descLines)*22.0
	layout := historicalEventLayout{descRect: gameui.Rect{X: horizontal, Y: contentY - float64(descLines)*22.0, W: contentW}}
	if len(choices) > 0 {
		promptLines := gameui.WrappedLineCount(renderText, prompt, contentW, gameui.TextMedium)
		if promptLines < 1 {
			promptLines = 1
		}
		contentY += 18
		layout.promptRect = gameui.Rect{X: horizontal, Y: contentY, W: contentW}
		contentY += float64(promptLines)*20.0 + 20.0

		choiceW := (contentW - choiceGap*float64(len(choices)-1)) / float64(len(choices))
		maxInfoH := 0.0
		layout.infoRects = make([]gameui.Rect, len(choices))
		for i, choice := range choices {
			infoLines := historicalChoiceInfoLineCount(choice, choiceW)
			infoH := float64(infoLines) * 16.0
			if infoH > maxInfoH {
				maxInfoH = infoH
			}
			layout.infoRects[i] = gameui.Rect{X: horizontal + float64(i)*(choiceW+choiceGap), Y: contentY, W: choiceW, H: infoH}
		}
		contentY += maxInfoH + 16.0
		buttonY := contentY
		btnW := choiceW
		layout.buttons = make([]gameui.Button, len(choices))
		for i := range choices {
			layout.buttons[i] = gameui.NewButton(horizontal+float64(i)*(btnW+choiceGap), buttonY, btnW, buttonH, "")
		}
		contentY += buttonH + 24.0
	} else {
		contentY += 24.0
	}

	panelH := maxF(260, contentY)
	panelH = minF(panelH, ScreenHeight-32)
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, panelW, panelH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	layout.modal = gameui.NewModal(ScreenWidth, ScreenHeight, panel)
	for i := range layout.buttons {
		layout.buttons[i].X += rect.X
		layout.buttons[i].Y += rect.Y
	}
	layout.descRect.X += rect.X
	layout.descRect.Y += rect.Y
	layout.promptRect.X += rect.X
	layout.promptRect.Y += rect.Y
	for i := range layout.infoRects {
		layout.infoRects[i].X += rect.X
		layout.infoRects[i].Y += rect.Y
	}
	return layout
}

func historicalChoiceInfoLineCount(choice HistoricalEventChoice, width float64) int {
	count := 0
	for _, value := range []string{choice.Desc, choice.Effect, choice.FollowUp, choice.Conditions} {
		if value == "" {
			continue
		}
		count += gameui.WrappedLineCount(renderText, value, width, gameui.TextSmall)
	}
	return count
}

func buildHistoricalEventModal(title, desc, prompt string, choices []HistoricalEventChoice) gameui.Modal {
	return buildHistoricalEventLayout(title, desc, prompt, choices).modal
}

func buildCommanderArrivalModal() gameui.Modal {
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, 1000, 420, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildHistoricalEventChoiceButtons(title, desc, prompt string, choices []HistoricalEventChoice) []gameui.Button {
	return buildHistoricalEventLayout(title, desc, prompt, choices).buttons
}

func buildDiplomacyOfferModal() gameui.Modal {
	const dlgW, dlgH = 760.0, 460.0
	rect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, dlgW, dlgH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(rect.X, rect.Y, rect.W, rect.H)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildDiplomacyOfferButtons() (gameui.Button, gameui.Button) {
	return buildDiplomacyOfferButtonsWithAcceptLabel("Kabul Et")
}

func buildDiplomacyOfferButtonsWithAcceptLabel(acceptLabel string) (gameui.Button, gameui.Button) {
	const btnW, btnH = 120.0, 36.0
	modal := buildDiplomacyOfferModal()
	// Offer dialog'da sağ özet panelinden uzak kalmak için butonları sol blokta tut.
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 12
	acceptX := modal.Panel.Rect.X + 16
	rejectX := acceptX + btnW + 12
	return gameui.NewButton(acceptX, btnY, btnW, btnH, acceptLabel).WithIcon(gameui.IconCheck),
		gameui.NewButton(rejectX, btnY, btnW, btnH, "Reddet").WithIcon(gameui.IconClose)
}

func buildDiplomacyOfferNoticeButton() gameui.Button {
	const btnW, btnH = 120.0, 36.0
	modal := buildDiplomacyOfferModal()
	btnY := modal.Panel.Rect.Y + modal.Panel.Rect.H - btnH - 12
	return gameui.NewButton(modal.Panel.Rect.X+16, btnY, btnW, btnH, "Tamam").WithIcon(gameui.IconCheck)
}
