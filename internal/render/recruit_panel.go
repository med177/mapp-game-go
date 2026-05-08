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
	"transport", "merchant_ship", "warship",
}

// unitSpriteLoc sprite sheet içindeki hücre konumunu tanımlar.
type unitSpriteLoc struct {
	row, col int
}

var unitSpriteLocs = map[string]unitSpriteLoc{
	"militia":        {0, 0},
	"infantry":       {0, 1},
	"elite_infantry": {0, 2},
	"light_cavalry":  {1, 0},
	"cavalry":        {1, 1},
	"heavy_cavalry":  {1, 2},
	"catapult":       {2, 0},
	"bombard":        {2, 1},
	"cannon":         {2, 2},
	"transport":      {3, 0},
	"merchant_ship":  {3, 1},
	"warship":        {3, 2},
}

// unitSpriteRect görüntü boyutuna göre sprite koordinatlarını döner.
// Görüntü 4 satır × 3 sütun eşit hücrelerden oluşur.
func unitSpriteRect(id string, sheet *ebiten.Image) image.Rectangle {
	loc, ok := unitSpriteLocs[id]
	if !ok {
		return image.Rectangle{}
	}
	cellW := sheet.Bounds().Dx() / 3
	cellH := sheet.Bounds().Dy() / 4
	x0 := loc.col * cellW
	y0 := loc.row * cellH
	return image.Rect(x0, y0, x0+cellW, y0+cellH)
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
	if !RecruitPanelBoundsHit(mx, my, gs, rid) {
		return ""
	}
	px := float64(recruitPanelX())
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW)
	ph := float64(recruitPanelH)

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
	display := visibleUnitIDs(gs, gs.Regions[rid])
	if idx >= len(display) {
		return ""
	}
	return display[idx]
}

func RecruitPanelBoundsHit(mx, my float64, gs *state.GameState, rid world.RegionID) bool {
	if !RecruitPanelVisible(gs, rid) {
		return false
	}
	px := float64(recruitPanelX())
	py := float64(recruitPanelY())
	pw := float64(recruitPanelW)
	ph := float64(recruitPanelH)
	return mx >= px && mx <= px+pw && my >= py && my <= py+ph
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
	titleW := MeasureText("BİRİM OLUŞTUR", FaceSmall)
	DrawText(screen, "BİRİM OLUŞTUR",
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

	display := visibleUnitIDs(gs, region)
	for i, uid := range display {
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

func visibleUnitIDs(gs *state.GameState, region *world.Region) []string {
	showNaval := region != nil && region.IsCoastal(gs.Regions)
	ids := make([]string, 0, len(unitDisplayOrder))
	for _, uid := range unitDisplayOrder {
		utype := gs.UnitTypes[uid]
		if utype == nil {
			continue
		}
		if utype.RequiredBldg == "port" && !showNaval {
			continue
		}
		ids = append(ids, uid)
	}
	return ids
}
