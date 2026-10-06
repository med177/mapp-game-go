package render

import (
	"image/color"
	"strconv"
	"strings"

	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

type landPassageFormField int

const (
	landPassageFieldNone landPassageFormField = iota
	landPassageFieldName
	landPassageFieldMoveCost
	landPassageFieldDefenseBonus
)

type landPassageFormState struct {
	show        bool
	editing     int
	from        world.RegionID
	to          world.RegionID
	start       [2]int
	end         [2]int
	active      landPassageFormField
	name        string
	passageType world.LandPassageType
	moveCost    string
	defense     string
	errorText   string
}

const (
	landPassageFormW = 560.0
	landPassageFormH = 350.0
)

func landPassageFormRect() gameui.Rect {
	return gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, landPassageFormW, landPassageFormH, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
}

func landPassageFormFieldRect(row int) gameui.Rect {
	r := landPassageFormRect()
	return gameui.Rect{X: r.X + 170, Y: r.Y + 72 + float64(row)*42, W: r.W - 194, H: 30}
}

func landPassageFormTypeButton() gameui.Button {
	r := landPassageFormRect()
	return gameui.NewButton(r.X+170, r.Y+198, r.W-194, 30, "")
}

func landPassageFormOKButton() gameui.Button {
	r := landPassageFormRect()
	return gameui.NewButton(r.X+r.W-178, r.Y+r.H-44, 76, 30, "Uygula")
}

func landPassageFormCancelButton() gameui.Button {
	r := landPassageFormRect()
	return gameui.NewButton(r.X+r.W-94, r.Y+r.H-44, 70, 30, "İptal")
}

func landPassageTypeLabel(value world.LandPassageType) string {
	switch value {
	case world.LandPassageBridge:
		return "Köprü"
	case world.LandPassageMountainPass:
		return "Dağ geçidi"
	case world.LandPassageFortifiedCrossing:
		return "Tahkimli geçiş"
	default:
		return "Boğaz / sığ geçiş"
	}
}

func nextLandPassageType(value world.LandPassageType) world.LandPassageType {
	switch value {
	case world.LandPassageStrait:
		return world.LandPassageBridge
	case world.LandPassageBridge:
		return world.LandPassageMountainPass
	case world.LandPassageMountainPass:
		return world.LandPassageFortifiedCrossing
	default:
		return world.LandPassageStrait
	}
}

func (r *Renderer) openNewLandPassageForm(from, to world.RegionID, start, end [2]int) {
	r.editLandPassageForm = landPassageFormState{
		show: true, editing: -1, from: from, to: to, start: start, end: end,
		passageType: world.LandPassageStrait, moveCost: "1", defense: "15",
		active: landPassageFieldName,
	}
}

func (r *Renderer) openSelectedLandPassageForm() {
	index := r.editLandPassageSelected
	if index < 0 || index >= len(r.gs.LandPassages) {
		return
	}
	p := r.gs.LandPassages[index]
	r.editLandPassageForm = landPassageFormState{
		show: true, editing: index, from: p.From, to: p.To, passageType: p.Type,
		name: p.Name, moveCost: strconv.Itoa(p.MoveCost), defense: strconv.Itoa(p.DefenseBonus),
		active: landPassageFieldName,
	}
}

func (r *Renderer) drawLandPassageForm(screen *ebiten.Image) {
	if !r.editLandPassageForm.show {
		return
	}
	rct := landPassageFormRect()
	gameui.DrawModal(screen, gameui.NewModal(ScreenWidth, ScreenHeight, gameui.NewPanel(rct.X, rct.Y, rct.W, rct.H)), standardModalStyle, renderText, func() {
		DrawText(screen, "GEÇİŞ ÖZELLİKLERİ", rct.X+24, rct.Y+24, FaceLarge, ColorGold)
		DrawText(screen, string(r.editLandPassageForm.from)+"  →  "+string(r.editLandPassageForm.to), rct.X+24, rct.Y+48, FaceSmall, ColorGray)
		drawLandPassageFormLabel(screen, rct.X+24, rct.Y+92, "Etiket")
		drawLandPassageFormLabel(screen, rct.X+24, rct.Y+134, "Hareket maliyeti")
		drawLandPassageFormLabel(screen, rct.X+24, rct.Y+176, "Savunma bonusu (%)")
		drawLandPassageFormLabel(screen, rct.X+24, rct.Y+218, "Geçiş tipi")
		drawLandPassageTextBox(screen, landPassageFormFieldRect(0), r.editLandPassageForm.name, r.editLandPassageForm.active == landPassageFieldName, "Geçiş adı")
		drawLandPassageTextBox(screen, landPassageFormFieldRect(1), r.editLandPassageForm.moveCost, r.editLandPassageForm.active == landPassageFieldMoveCost, "1 veya daha büyük")
		drawLandPassageTextBox(screen, landPassageFormFieldRect(2), r.editLandPassageForm.defense, r.editLandPassageForm.active == landPassageFieldDefenseBonus, "0-100")
		typeButton := landPassageFormTypeButton()
		typeButton.Label = landPassageTypeLabel(r.editLandPassageForm.passageType)
		drawUIButtonWidget(screen, typeButton, tinyButtonStyle)
		if r.editLandPassageForm.errorText != "" {
			DrawText(screen, r.editLandPassageForm.errorText, rct.X+24, rct.Y+rct.H-68, FaceSmall, ColorRed)
		}
		drawUIButtonWidget(screen, landPassageFormOKButton(), tinyButtonStyle)
		drawUIButtonWidget(screen, landPassageFormCancelButton(), tinyButtonStyle)
	})
}

func drawLandPassageFormLabel(screen *ebiten.Image, x, y float64, label string) {
	DrawText(screen, label, x, y, FaceSmall, ColorGray)
}

func drawLandPassageTextBox(screen *ebiten.Image, rect gameui.Rect, value string, focused bool, placeholder string) {
	box := gameui.NewTextBox(rect.X, rect.Y, rect.W, rect.H, placeholder)
	box.Value = value
	box.Focused = focused
	box.MaxLen = 64
	gameui.DrawTextBox(screen, box, gameui.TextBoxStyle{BG: color.RGBA{28, 32, 38, 245}, Border: color.RGBA{120, 105, 60, 210}, Focused: ColorGold, Text: ColorWhite, Placeholder: ColorGray, BorderWidth: 1, TextOffsetX: 8, TextOffsetY: 7, TextVariant: gameui.TextSmall}, renderText)
}

func (r *Renderer) handleLandPassageFormInput() InputAction {
	form := &r.editLandPassageForm
	if r.keyJustPressed(ebiten.KeyEscape) {
		form.show = false
		return InputAction{}
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		mx, my := cursorX(), cursorY()
		if landPassageFormCancelButton().HitTest(mx, my) {
			form.show = false
			return InputAction{}
		}
		if landPassageFormOKButton().HitTest(mx, my) {
			r.commitLandPassageForm()
			return InputAction{}
		}
		if landPassageFormTypeButton().HitTest(mx, my) {
			form.passageType = nextLandPassageType(form.passageType)
			return InputAction{}
		}
		for row, field := range []landPassageFormField{landPassageFieldName, landPassageFieldMoveCost, landPassageFieldDefenseBonus} {
			if landPassageFormFieldRect(row).Hit(mx, my) {
				form.active = field
				form.errorText = ""
				return InputAction{}
			}
		}
	}
	if form.active == landPassageFieldNone {
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) {
		value := r.landPassageActiveValue()
		if len(value) > 0 {
			r.setLandPassageActiveValue(value[:len(value)-1])
		}
	}
	if editCtrlPressed() && r.keyJustPressed(ebiten.KeyV) {
		if text, ok := readEditClipboardText(); ok {
			r.setLandPassageActiveValue(string(appendEditText([]rune(r.landPassageActiveValue()), text, 64)))
		}
		return InputAction{}
	}
	chars := ebiten.AppendInputChars(nil)
	if len(chars) > 0 {
		value := r.landPassageActiveValue() + string(chars)
		if form.active != landPassageFieldName {
			value = strings.Map(func(ch rune) rune {
				if ch >= '0' && ch <= '9' {
					return ch
				}
				return -1
			}, value)
		}
		r.setLandPassageActiveValue(limitStringRunes(value, 64))
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		r.commitLandPassageForm()
	}
	return InputAction{}
}

func (r *Renderer) landPassageActiveValue() string {
	switch r.editLandPassageForm.active {
	case landPassageFieldMoveCost:
		return r.editLandPassageForm.moveCost
	case landPassageFieldDefenseBonus:
		return r.editLandPassageForm.defense
	default:
		return r.editLandPassageForm.name
	}
}

func (r *Renderer) setLandPassageActiveValue(value string) {
	switch r.editLandPassageForm.active {
	case landPassageFieldMoveCost:
		r.editLandPassageForm.moveCost = value
	case landPassageFieldDefenseBonus:
		r.editLandPassageForm.defense = value
	default:
		r.editLandPassageForm.name = value
	}
}

func (r *Renderer) commitLandPassageForm() {
	form := &r.editLandPassageForm
	moveCost, moveErr := strconv.Atoi(strings.TrimSpace(form.moveCost))
	defense, defenseErr := strconv.Atoi(strings.TrimSpace(form.defense))
	if moveErr != nil || moveCost < 1 || defenseErr != nil || defense < 0 || defense > 100 {
		form.errorText = "Hareket maliyeti 1+, savunma bonusu 0-100 olmalı."
		return
	}
	name := strings.TrimSpace(form.name)
	if form.editing >= 0 && form.editing < len(r.gs.LandPassages) {
		p := &r.gs.LandPassages[form.editing]
		p.Name, p.Type, p.MoveCost, p.DefenseBonus = name, form.passageType, moveCost, defense
		r.editLandPassageSelected = form.editing
	} else {
		start, end := form.start, form.end
		r.gs.LandPassages = append(r.gs.LandPassages, world.LandPassage{From: form.from, To: form.to, Name: name, Type: form.passageType, MoveCost: moveCost, DefenseBonus: defense, Start: &start, End: &end})
		r.editLandPassageSelected = len(r.gs.LandPassages) - 1
	}
	r.editDirty = true
	form.show = false
	r.editLandPassageMessage = "geçiş kaydedildi"
}
