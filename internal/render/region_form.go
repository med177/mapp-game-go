package render

import (
	"image/color"
	"strconv"
	"strings"

	"mapp-game-go/internal/religion"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

var editRegionFormFields = [...]struct {
	field editRegionFormField
	label string
}{
	{editRegionFieldNameTR, "Ad TR"},
	{editRegionFieldName, "Ad EN"},
	{editRegionFieldGold, "Altın geliri"},
	{editRegionFieldGrain, "Tahıl geliri"},
	{editRegionFieldIron, "Demir geliri"},
	{editRegionFieldTimber, "Kereste geliri"},
	{editRegionFieldStone, "Taş geliri"},
	{editRegionFieldSpice, "Baharat geliri"},
	{editRegionFieldCloth, "Kumaş geliri"},
	{editRegionFieldTradeCapacity, "Ticaret kapasitesi"},
	{editRegionFieldSatisfaction, "Memnuniyet"},
	{editRegionFieldTaxRate, "Vergi oranı"},
	{editRegionFieldPopulation, "Toplam nüfus"},
	{editRegionFieldRuralPopulation, "Kırsal nüfus"},
	{editRegionFieldReligion, "Din"},
	{editRegionFieldActiveEvent, "Aktif olay ID"},
	{editRegionFieldUnlockTurn, "Açılış turu"},
}

func editRegionFormRect() (float32, float32, float32, float32) {
	const w, h = float32(760), float32(600)
	return float32(ScreenWidth)/2 - w/2, float32(ScreenHeight)/2 - h/2, w, h
}

func editRegionFormHit(mx, my float64) bool {
	x, y, w, h := editRegionFormRect()
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
}

func editRegionFormSaveButton() gameui.Button {
	x, y, _, h := editRegionFormRect()
	return gameui.NewButton(float64(x)+18, float64(y+h-48), 150, 32, "Kaydet")
}

func editRegionFormCancelButton() gameui.Button {
	x, y, w, h := editRegionFormRect()
	return gameui.NewButton(float64(x+w-168), float64(y+h-48), 150, 32, "İptal")
}

func editRegionFormButtonHit(mx, my float64) bool {
	return editRegionFormSaveButton().HitTest(mx, my) || editRegionFormCancelButton().HitTest(mx, my)
}

func editRegionFormReligionButton() gameui.Button {
	rect := editRegionFormFieldRect(editRegionFieldReligion)
	return gameui.NewButton(rect[0], rect[1], rect[2], rect[3], "")
}

func (r *Renderer) editRegionFormInteractiveHit(mx, my float64) bool {
	if editRegionFormButtonHit(mx, my) || editRegionFormReligionButton().HitTest(mx, my) {
		return true
	}
	return r.editRegionReligionDropdown != nil && r.editRegionReligionDropdown.IsOpen() && r.editRegionReligionDropdown.HitTest(mx, my)
}

func editRegionFormFieldRect(field editRegionFormField) uiRect {
	x, y, w, _ := editRegionFormRect()
	index := int(field) - 1
	if index < 0 || index >= len(editRegionFormFields) {
		return uiRect{}
	}
	const (
		leftInset = 18.0
		gap       = 14.0
		fieldH    = 30.0
		rowGap    = 18.0
	)
	fieldW := (float64(w) - leftInset*2 - gap) / 2
	row := float64(index / 2)
	col := float64(index % 2)
	return uiRect{
		float64(x) + leftInset + col*(fieldW+gap),
		float64(y) + 70 + row*(fieldH+rowGap),
		fieldW,
		fieldH,
	}
}

func (r *Renderer) drawEditRegionForm(screen *ebiten.Image) {
	if !r.editRegionForm.show {
		return
	}
	x, y, w, h := editRegionFormRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{14, 18, 22, 248})
	drawPanelBorder(screen, x, y, w, h)
	DrawText(screen, "BÖLGE VERİLERİ", float64(x)+18, float64(y)+14, FaceLarge, ColorGold)
	DrawText(screen, "JSON alanlarını düzenle · Enter kaydetmez, aşağıdaki Kaydet düğmesini kullanır.",
		float64(x)+18, float64(y)+42, FaceSmall, ColorGray)

	for _, item := range editRegionFormFields {
		r.drawEditRegionFormField(screen, item.field, item.label, r.editRegionForm.values[item.field])
	}
	if r.editRegionForm.errorText != "" {
		DrawText(screen, r.editRegionForm.errorText, float64(x)+18, float64(y)+float64(h)-72, FaceSmall, ColorRed)
	}
	drawUIButtonWidget(screen, editRegionFormSaveButton(), tinyButtonStyle)
	drawUIButtonWidget(screen, editRegionFormCancelButton(), tinyButtonStyle)
	if r.editRegionReligionDropdown != nil {
		drawUIDropdown(screen, r.editRegionReligionDropdown)
	}
}

func (r *Renderer) drawEditRegionFormField(screen *ebiten.Image, field editRegionFormField, label, value string) {
	rect := editRegionFormFieldRect(field)
	if field == editRegionFieldReligion {
		drawUIButtonWidget(screen, editRegionFormReligionButton(), tinyButtonStyle)
		DrawText(screen, religion.DisplayNameTR(religion.Type(value)), rect[0]+8, rect[1]+7, FaceSmall, ColorWhite)
		gameui.NewLabel(rect[0], rect[1]-16, label, ColorGray).Draw(screen, renderText)
		return
	}
	box := gameui.NewTextBox(rect[0], rect[1], rect[2], rect[3], "")
	box.Value = value
	box.Focused = r.editRegionForm.active == field
	gameui.DrawTextBox(screen, box, editRegionFormTextBoxStyle(), renderText)
	gameui.NewLabel(rect[0], rect[1]-16, label, ColorGray).Draw(screen, renderText)
}

func editRegionFormTextBoxStyle() gameui.TextBoxStyle {
	return gameui.TextBoxStyle{
		BG:          color.RGBA{28, 32, 38, 235},
		Border:      color.RGBA{120, 105, 60, 210},
		Focused:     ColorGold,
		Text:        ColorWhite,
		Placeholder: ColorGray,
		BorderWidth: 1,
		TextOffsetX: 8,
		TextOffsetY: 7,
		TextVariant: gameui.TextSmall,
	}
}

func (r *Renderer) openEditRegionForm() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	values := [editRegionFormFieldCount]string{}
	values[editRegionFieldNameTR] = region.NameTR
	values[editRegionFieldName] = region.Name
	values[editRegionFieldGold] = strconv.Itoa(region.BaseGoldIncome)
	values[editRegionFieldGrain] = strconv.Itoa(region.BaseGrainOutput)
	values[editRegionFieldIron] = strconv.Itoa(region.BaseIronOutput)
	values[editRegionFieldTimber] = strconv.Itoa(region.BaseTimberOutput)
	values[editRegionFieldStone] = strconv.Itoa(region.BaseStoneOutput)
	values[editRegionFieldSpice] = strconv.Itoa(region.BaseSpiceOutput)
	values[editRegionFieldCloth] = strconv.Itoa(region.BaseClothOutput)
	values[editRegionFieldTradeCapacity] = strconv.Itoa(region.TradeCapacity)
	values[editRegionFieldSatisfaction] = strconv.Itoa(region.Satisfaction)
	values[editRegionFieldTaxRate] = strconv.Itoa(region.TaxRate)
	values[editRegionFieldPopulation] = strconv.Itoa(region.Population)
	values[editRegionFieldRuralPopulation] = strconv.Itoa(region.RuralPopulation)
	values[editRegionFieldReligion] = region.Religion
	values[editRegionFieldActiveEvent] = region.ActiveEventID
	values[editRegionFieldUnlockTurn] = strconv.Itoa(region.UnlockTurn)
	r.editRegionForm = editRegionFormState{
		show:     true,
		regionID: region.ID,
		active:   editRegionFieldNameTR,
		values:   values,
	}
	options := religion.All()
	displayOptions := make([]string, len(options))
	for i, option := range options {
		displayOptions[i] = religion.DisplayNameTR(option)
	}
	r.editRegionReligionDropdown.SetOptions(displayOptions, religion.DisplayNameTR(religion.Type(region.Religion)))
	r.editRegionReligionDropdown.Close()
}

func (r *Renderer) handleEditRegionFormInput() InputAction {
	form := &r.editRegionForm
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.editRegionReligionDropdown != nil && r.editRegionReligionDropdown.IsOpen() {
			r.editRegionReligionDropdown.Close()
			return InputAction{}
		}
		form.show = false
		return InputAction{}
	}
	if r.editRegionReligionDropdown != nil && r.editRegionReligionDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editRegionReligionDropdown.HitTest(fx, fy) {
			r.editRegionReligionDropdown.Scroll(wheelY)
			return InputAction{}
		}
		if r.mouseJustPressed(ebiten.MouseButtonLeft) {
			if idx, ok := r.editRegionReligionDropdown.GetSelectedOption(fx, fy); ok {
				options := religion.All()
				if idx < len(options) {
					form.values[editRegionFieldReligion] = string(options[idx])
				}
				r.editRegionReligionDropdown.Close()
				return InputAction{}
			}
			if !r.editRegionReligionDropdown.HitTest(fx, fy) {
				r.editRegionReligionDropdown.Close()
				return InputAction{}
			}
		}
		return InputAction{}
	}
	if !editRegionFormHit(fx, fy) {
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if editRegionFormSaveButton().HitTest(fx, fy) {
			r.saveEditRegionForm()
			return InputAction{}
		}
		if editRegionFormCancelButton().HitTest(fx, fy) {
			form.show = false
			return InputAction{}
		}
		if editRegionFormReligionButton().HitTest(fx, fy) {
			rect := editRegionFormFieldRect(editRegionFieldReligion)
			r.editRegionReligionDropdown.SetPosition(rect[0], rect[1])
			r.editRegionReligionDropdown.Toggle()
			form.active = editRegionFieldNone
			return InputAction{}
		}
		for _, item := range editRegionFormFields {
			rect := editRegionFormFieldRect(item.field)
			if (gameui.Rect{X: rect[0], Y: rect[1], W: rect[2], H: rect[3]}).Hit(fx, fy) {
				form.active = item.field
				form.errorText = ""
				return InputAction{}
			}
		}
	}
	if form.active == editRegionFieldNone {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(form.values[form.active]) > 0 {
		form.values[form.active] = trimLastRune(form.values[form.active])
	}
	chars := ebiten.AppendInputChars(nil)
	if len(chars) > 0 {
		limit := 64
		if form.active >= editRegionFieldGold && form.active <= editRegionFieldUnlockTurn {
			limit = 12
		}
		form.values[form.active] = limitStringRunes(form.values[form.active]+string(chars), limit)
	}
	return InputAction{}
}

func (r *Renderer) saveEditRegionForm() {
	form := &r.editRegionForm
	region := r.gs.Regions[form.regionID]
	if region == nil {
		form.errorText = "Bölge bulunamadı."
		return
	}
	parse := func(field editRegionFormField) (int, bool) {
		value, err := strconv.Atoi(strings.TrimSpace(form.values[field]))
		return value, err == nil
	}
	ints := make(map[editRegionFormField]int, 15)
	for _, field := range []editRegionFormField{
		editRegionFieldGold, editRegionFieldGrain, editRegionFieldIron, editRegionFieldTimber,
		editRegionFieldStone, editRegionFieldSpice, editRegionFieldCloth, editRegionFieldTradeCapacity,
		editRegionFieldSatisfaction, editRegionFieldTaxRate, editRegionFieldPopulation,
		editRegionFieldRuralPopulation, editRegionFieldUnlockTurn,
	} {
		value, ok := parse(field)
		if !ok {
			form.errorText = "Sayısal alanlardan biri geçersiz."
			return
		}
		ints[field] = value
	}
	if ints[editRegionFieldSatisfaction] < 0 || ints[editRegionFieldSatisfaction] > 100 ||
		ints[editRegionFieldTaxRate] < 0 || ints[editRegionFieldTaxRate] > world.MaxTaxRate ||
		ints[editRegionFieldPopulation] < 0 || ints[editRegionFieldRuralPopulation] < 0 ||
		ints[editRegionFieldUnlockTurn] < 0 {
		form.errorText = "Memnuniyet/vergi/nüfus/açılış değeri geçersiz."
		return
	}

	before := r.worldSnapshot()
	region.NameTR = form.values[editRegionFieldNameTR]
	region.Name = form.values[editRegionFieldName]
	region.BaseGoldIncome = ints[editRegionFieldGold]
	region.BaseGrainOutput = ints[editRegionFieldGrain]
	region.BaseIronOutput = ints[editRegionFieldIron]
	region.BaseTimberOutput = ints[editRegionFieldTimber]
	region.BaseStoneOutput = ints[editRegionFieldStone]
	region.BaseSpiceOutput = ints[editRegionFieldSpice]
	region.BaseClothOutput = ints[editRegionFieldCloth]
	region.TradeCapacity = ints[editRegionFieldTradeCapacity]
	region.Satisfaction = ints[editRegionFieldSatisfaction]
	region.TaxRate = ints[editRegionFieldTaxRate]
	region.Population = ints[editRegionFieldPopulation]
	region.RuralPopulation = ints[editRegionFieldRuralPopulation]
	region.Religion = strings.TrimSpace(form.values[editRegionFieldReligion])
	region.ActiveEventID = strings.TrimSpace(form.values[editRegionFieldActiveEvent])
	region.UnlockTurn = ints[editRegionFieldUnlockTurn]
	r.rebuildEditWorldMap()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
	form.show = false
	r.editRegionReligionDropdown.Close()
}
