package render

import (
	"image"
	"image/color"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Sprite sheet yükleyici ─────────────────────────────────────────────

var (
	armySheet       *ebiten.Image
	armySheetLoaded bool
)

func ensureArmySheet() {
	if armySheetLoaded {
		return
	}
	armySheetLoaded = true
	armySheet = tryLoadImage(ActiveScenarioPath + "/sprites/army.png")
}

// unitDisplayOrder panel slotlarının sırasını belirler (3 sütun × 4 satır).
var unitDisplayOrder = []string{
	"militia", "infantry", "elite_infantry",
	"light_cavalry", "cavalry", "heavy_cavalry",
	"catapult", "bombard", "cannon",
	"warship", "transport", "merchant_ship",
}

// unitSpriteLoc sprite sheet içindeki hücre konumunu tanımlar.
type unitSpriteLoc struct {
	row, col  int
	largeGrid bool // true = 3-sütunlu büyük ızgara, false = 6-sütunlu küçük ızgara
}

var unitSpriteLocs = map[string]unitSpriteLoc{
	"militia":        {0, 0, true},
	"infantry":       {0, 1, true},
	"elite_infantry": {0, 2, true},
	"light_cavalry":  {1, 0, true},
	"cavalry":        {1, 1, true},
	"heavy_cavalry":  {1, 2, true},
	"catapult":       {2, 0, false},
	"bombard":        {2, 1, false},
	"cannon":         {2, 2, false},
	"warship":        {2, 3, false},
	"transport":      {2, 4, false},
	"merchant_ship":  {2, 5, false},
}

// unitSpriteRect görüntü boyutuna göre sprite koordinatlarını döner.
// Üst 2 satır: 3 sütun × ~38% yükseklik (piyade/süvari).
// Alt satır: 6 sütun × kalan alan (kuşatma/deniz).
func unitSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
	loc, ok := unitSpriteLocs[id]
	if !ok {
		return image.Rectangle{}
	}
	W := float64(sheet.Bounds().Dx())
	H := float64(sheet.Bounds().Dy())

	largeRowH := H * 0.37
	largeColW := W / 3
	smallRowY := H * 0.73
	smallRowH := H - smallRowY
	smallColW := W / 6

	var x0, y0, x1, y1 float64
	if loc.largeGrid {
		x0 = float64(loc.col) * largeColW
		y0 = float64(loc.row) * largeRowH
		x1 = x0 + largeColW
		y1 = y0 + largeRowH*0.87 // alt label alanını kırp
	} else {
		x0 = float64(loc.col) * smallColW
		y0 = smallRowY
		x1 = x0 + smallColW
		y1 = smallRowY + smallRowH*0.83
	}
	return image.Rect(int(x0), int(y0), int(x1), int(y1))
}

// ── Panel layout sabitleri ─────────────────────────────────────────────

const (
	recruitPanelW = infoPanelW
	recruitPanelH = infoPanelH
)

func recruitPanelX() float32 { return infoPanelX() + infoPanelW + 5 }
func recruitPanelY() float32 { return infoPanelY() }

// RecruitPanelVisible oyuncunun kendi bölgesi seçiliyken true döner.
func RecruitPanelVisible(gs *state.GameState, rid world.RegionID) bool {
	if rid == "" {
		return false
	}
	r, ok := gs.Regions[rid]
	return ok && !r.IsSea && r.OwnerID == string(gs.PlayerFactionID)
}

// RecruitPanelHitTest fare koordinatına denk gelen unit ID'sini döner; boş = tıklama yok.
func RecruitPanelHitTest(mx, my float64, gs *state.GameState, rid world.RegionID) string {
	if !RecruitPanelVisible(gs, rid) {
		return ""
	}
	px := float64(recruitPanelX())
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW)
	ph := float64(recruitPanelH)

	if mx < px || mx > px+pw || my < py || my > py+ph {
		return ""
	}

	const headerH = 28.0
	pad := panelPad
	availW := pw - pad*2
	slotW := availW / 3
	slotH := (ph - headerH - pad) / 4

	relX := mx - px - pad
	relY := my - py - headerH
	if relX < 0 || relY < 0 {
		return ""
	}

	col := int(relX / slotW)
	row := int(relY / slotH)
	if col < 0 || col > 2 || row < 0 || row > 3 {
		return ""
	}
	idx := row*3 + col
	if idx >= len(unitDisplayOrder) {
		return ""
	}
	return unitDisplayOrder[idx]
}

// DrawRecruitPanel birim seçim ızgarasını bölge panelinin sağına çizer.
func DrawRecruitPanel(screen *ebiten.Image, gs *state.GameState, rid world.RegionID) {
	if !RecruitPanelVisible(gs, rid) {
		return
	}
	region := gs.Regions[rid]
	f := gs.Factions[gs.PlayerFactionID]

	ensureArmySheet()

	px := recruitPanelX()
	py := recruitPanelY()
	pw := recruitPanelW
	ph := recruitPanelH

	// Arka plan ve çerçeve
	vector.FillRect(screen, px, py, pw, ph, panelBg, false)
	drawPanelBorder(screen, px, py, pw, ph)
	vector.FillRect(screen, px, py, pw, 3, panelBorder, false)

	// Başlık
	titleW := MeasureText("ASKER AL", FaceSmall)
	DrawText(screen, "ASKER AL",
		float64(px)+float64(pw)/2-titleW/2, float64(py)+8,
		FaceSmall, color.RGBA{200, 170, 90, 220})
	sepY := py + 24
	vector.StrokeLine(screen, px+12, sepY, px+float32(pw)-12, sepY, 1, panelBorder, false)

	// Bölge binaları
	hasBarracks, hasPort := false, false
	for _, bid := range region.Buildings {
		switch bid {
		case "barracks":
			hasBarracks = true
		case "port":
			hasPort = true
		}
	}

	// Izgara boyutları
	const cols = 3
	const headerH = float32(28)
	pad := float32(panelPad)
	availW := float32(pw) - pad*2
	slotW := availW / float32(cols)
	slotH := (float32(ph) - headerH - pad) / 4
	spriteH := slotH * 0.62
	nameY_off := spriteH + 2
	costY_off := nameY_off + 13

	for i, uid := range unitDisplayOrder {
		col := i % cols
		row := i / cols

		sx := px + pad + float32(col)*slotW
		sy := py + headerH + float32(row)*slotH
		innerW := slotW - 3

		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}

		// Kullanılabilirlik kontrolü
		var needsBuilding bool
		switch utype.RequiredBldg {
		case "barracks":
			needsBuilding = !hasBarracks
		case "port":
			needsBuilding = !hasPort
		}

		needsTech := utype.RequiredTech != "" &&
			(f == nil || !f.Research.Completed[utype.RequiredTech])

		canAfford := f != nil && f.Gold >= utype.GoldCost
		fullyAvail := !needsBuilding && !needsTech

		// Slot arka plan
		slotBg := color.RGBA{20, 16, 12, 200}
		borderCol := color.RGBA{55, 45, 30, 200}
		if fullyAvail {
			slotBg = color.RGBA{38, 30, 15, 235}
			borderCol = panelBorder
		}
		vector.FillRect(screen, sx, sy, innerW, spriteH, slotBg, false)
		vector.StrokeRect(screen, sx, sy, innerW, spriteH, 1, borderCol, false)

		// Sprite
		if armySheet != nil {
			r := unitSpriteRect(uid, armySheet)
			if !r.Empty() {
				sub := armySheet.SubImage(r).(*ebiten.Image)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(
					float64(innerW-2)/float64(r.Dx()),
					float64(spriteH-2)/float64(r.Dy()),
				)
				op.GeoM.Translate(float64(sx+1), float64(sy+1))
				switch {
				case needsBuilding:
					op.ColorScale.Scale(0.25, 0.25, 0.25, 1.0)
				case needsTech:
					op.ColorScale.Scale(0.45, 0.45, 0.45, 1.0)
				case !canAfford:
					op.ColorScale.Scale(0.65, 0.45, 0.45, 1.0)
				}
				screen.DrawImage(sub, op)

				// Kilit ikonu — teknoloji eksik
				if needsTech && !needsBuilding {
					DrawTextCentered(screen, "🔒",
						float64(sx)+float64(innerW)/2, float64(sy)+float64(spriteH)/2-8,
						FaceMed, color.RGBA{200, 180, 80, 220})
				}
			}
		}

		// Birim adı
		nameX := float64(sx) + float64(innerW)/2
		nameCol := ColorGold
		if !fullyAvail {
			nameCol = color.RGBA{80, 70, 55, 190}
		}
		DrawTextCentered(screen, utype.NameTR, nameX, float64(sy)+float64(nameY_off), FaceSmall, nameCol)

		// Maliyet
		costStr := itoa(utype.GoldCost) + " ✦"
		costCol := color.RGBA{180, 160, 60, 220}
		if !fullyAvail {
			costCol = color.RGBA{70, 62, 48, 180}
		} else if !canAfford {
			costCol = ColorRed
		}
		DrawTextCentered(screen, costStr, nameX, float64(sy)+float64(costY_off), FaceSmall, costCol)
	}
}
