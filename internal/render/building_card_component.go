package render

import (
	"image/color"

	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type buildingCardComponent struct {
	ID          string
	Name        string
	Building    *city.Building
	Rect        gameui.Rect
	ImageRect   gameui.Rect
	SpriteRect  gameui.Rect
	LabelY      float64
	Level       int
	MaxLevel    int
	QueuedCount int
	TurnsLeft   int
	CanAfford   bool
}

var (
	lastBuildingGridRegionID world.RegionID
	lastBuildingGridCards    []buildingCardComponent
)

func cacheBuildingGridComponents(regionID world.RegionID, cards []buildingCardComponent) {
	lastBuildingGridRegionID = regionID
	lastBuildingGridCards = cards
}

func lastDrawnBuildingGridHit(mx, my float64, rid world.RegionID) (string, bool) {
	if rid == "" || lastBuildingGridRegionID != rid || lastBuildingGridCards == nil {
		return "", false
	}
	for _, card := range lastBuildingGridCards {
		if card.HitTest(mx, my) {
			return card.ID, true
		}
	}
	return "", true
}

func buildBuildingCardComponents(gs *state.GameState, region *world.Region, panelX, startY, panelW float32) []buildingCardComponent {
	if gs == nil || region == nil {
		return nil
	}

	builtCount := make(map[string]int, len(region.Buildings))
	for _, bid := range region.Buildings {
		builtCount[bid]++
	}
	queuedSet := make(map[string]int)
	queuedTurnsMin := make(map[string]int)
	for _, order := range gs.ProductionQueue {
		if order.Kind == "building" && order.RegionID == region.ID {
			queuedSet[order.TypeID]++
			if queuedTurnsMin[order.TypeID] == 0 || order.TurnsLeft < queuedTurnsMin[order.TypeID] {
				queuedTurnsMin[order.TypeID] = order.TurnsLeft
			}
		}
	}

	display := visibleBuildingIDs(gs, region)
	cards := make([]buildingCardComponent, 0, len(display))
	for i, bid := range display {
		b, hasDef := gs.BuildingTypes[bid]
		name := bid
		maxLevel := 1
		canAfford := false
		if hasDef {
			name = b.NameTR
			if b.MaxPerRegion > 0 {
				maxLevel = b.MaxPerRegion
			}
			if f := gs.Factions[gs.PlayerFactionID]; f != nil {
				cost := economy.ResourceCost{
					Gold:   b.GoldCost,
					Grain:  b.GrainCost,
					Iron:   b.IronCost,
					Timber: b.TimberCost,
					Stone:  b.StoneCost,
				}
				canAfford = cost.CanAfford(f)
			}
		}

		rect, imageRect, labelY := buildingCardLayout(panelX, startY, panelW, i)
		spriteRect := imageRect
		if buildingSheet != nil {
			spriteRect = buildingSpriteDrawRect(bid, imageRect)
		}
		cards = append(cards, buildingCardComponent{
			ID:          bid,
			Name:        name,
			Building:    b,
			Rect:        rect,
			ImageRect:   imageRect,
			SpriteRect:  spriteRect,
			LabelY:      labelY,
			Level:       builtCount[bid],
			MaxLevel:    maxLevel,
			QueuedCount: queuedSet[bid],
			TurnsLeft:   queuedTurnsMin[bid],
			CanAfford:   canAfford,
		})
	}
	return cards
}

func buildingCardLayout(panelX, startY, panelW float32, index int) (gameui.Rect, gameui.Rect, float64) {
	const cols = 3
	pad := float32(panelPad)
	availW := panelW - pad*2
	slotW := availW / float32(cols)
	cardH := buildingGridSpriteH + buildingGridNameH
	rowH := cardH + buildingGridRowGap

	col := index % cols
	row := index / cols
	x := panelX + pad + float32(col)*slotW
	y := startY + float32(row)*rowH
	w := slotW - 3

	rect := gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(cardH)}
	imageRect := gameui.Rect{X: rect.X, Y: rect.Y, W: rect.W, H: float64(buildingGridSpriteH)}
	labelY := imageRect.Y + imageRect.H + 3
	return rect, imageRect, labelY
}

func (c buildingCardComponent) HitTest(mx, my float64) bool {
	return c.SpriteRect.Hit(mx, my)
}

func (c buildingCardComponent) Draw(screen *ebiten.Image) {
	isBuilt := c.Level > 0
	isQueued := c.QueuedCount > 0
	isMaxLevel := c.Level >= c.MaxLevel

	cardBg := color.RGBA{250, 250, 250, 240}
	borderCol := color.RGBA{160, 160, 160, 220}
	switch {
	case isBuilt:
		cardBg = color.RGBA{255, 255, 255, 245}
		borderCol = color.RGBA{150, 130, 85, 230}
	case isQueued:
		cardBg = color.RGBA{248, 248, 248, 240}
		borderCol = color.RGBA{145, 145, 145, 220}
	case c.CanAfford:
		cardBg = color.RGBA{252, 252, 252, 242}
		borderCol = color.RGBA{165, 165, 165, 220}
	}
	drawUICardRect(screen, c.Rect, cardBg, borderCol, 1)
	vector.StrokeLine(screen,
		float32(c.ImageRect.X), float32(c.ImageRect.Y+c.ImageRect.H),
		float32(c.ImageRect.X+c.ImageRect.W), float32(c.ImageRect.Y+c.ImageRect.H),
		1, color.RGBA{210, 210, 210, 170}, false,
	)

	if buildingSheet != nil {
		c.drawSprite(screen, isBuilt)
	}

	nameCol := color.RGBA{75, 65, 50, 220}
	switch {
	case isBuilt:
		nameCol = ColorGold
	case isQueued:
		nameCol = color.RGBA{210, 190, 120, 230}
	case c.CanAfford:
		nameCol = color.RGBA{170, 145, 85, 230}
	}
	DrawTextCentered(screen, c.Name, c.Rect.X+c.Rect.W/2, c.LabelY, FaceSmall, nameCol)

	if isBuilt {
		c.drawLevelBadge(screen, isMaxLevel)
	}
	if isQueued {
		c.drawQueuedBadge(screen)
	}
}

func (c buildingCardComponent) drawSprite(screen *ebiten.Image, isBuilt bool) {
	r := buildingSpriteRect(c.ID, buildingSheet)
	sub := buildingSheet.SubImage(r).(*ebiten.Image)
	op := &ebiten.DrawImageOptions{}
	scale := c.SpriteRect.W / float64(r.Dx())
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(c.SpriteRect.X, c.SpriteRect.Y)
	if !isBuilt {
		if c.CanAfford {
			op.ColorScale.Scale(0.65, 0.65, 0.65, 0.9)
		} else {
			op.ColorScale.Scale(0.35, 0.35, 0.35, 0.85)
		}
	}
	screen.DrawImage(sub, op)
}

func buildingSpriteDrawRect(bid string, imageRect gameui.Rect) gameui.Rect {
	if buildingSheet == nil {
		return imageRect
	}
	r := buildingSpriteRect(bid, buildingSheet)
	fitW := imageRect.W - 2
	fitH := imageRect.H - 2
	scale := fitW / float64(r.Dx())
	if hScale := fitH / float64(r.Dy()); hScale < scale {
		scale = hScale
	}
	drawW := float64(r.Dx()) * scale
	drawH := float64(r.Dy()) * scale
	return gameui.Rect{
		X: imageRect.X + imageRect.W/2 - drawW/2,
		Y: imageRect.Y + imageRect.H/2 - drawH/2,
		W: drawW,
		H: drawH,
	}
}

func (c buildingCardComponent) drawLevelBadge(screen *ebiten.Image, isMaxLevel bool) {
	vector.StrokeRect(screen, float32(c.ImageRect.X)+1, float32(c.ImageRect.Y)+1, float32(c.ImageRect.W)-2, float32(c.ImageRect.H)-2, 1, color.RGBA{160, 130, 50, 120}, false)
	lvText := "Lv" + itoa(c.Level)
	lvX := c.ImageRect.X + 6
	lvY := c.ImageRect.Y + 4
	lvW := float32(MeasureText(lvText, FaceSmall) + 8)
	lvH := float32(14)
	lvBg := color.RGBA{18, 14, 8, 225}
	lvBorder := color.RGBA{170, 140, 75, 220}
	if isMaxLevel {
		lvBg = color.RGBA{150, 34, 34, 235}
		lvBorder = color.RGBA{225, 120, 120, 235}
	}
	drawUICardRect(screen, gameui.Rect{X: lvX - 3, Y: lvY - 2, W: float64(lvW), H: float64(lvH)}, lvBg, lvBorder, 1)
	DrawText(screen, lvText, lvX, lvY, FaceSmall, color.RGBA{255, 245, 220, 250})
}

func (c buildingCardComponent) drawQueuedBadge(screen *ebiten.Image) {
	qLabel := itoa(c.TurnsLeft) + " Tur"
	if c.QueuedCount > 1 {
		qLabel = "x" + itoa(c.QueuedCount) + " " + qLabel
	}
	qTextW := MeasureText(qLabel, FaceSmall)
	qPadX := float32(8)
	qBadgeH := float32(18)
	qBadgeW := float32(qTextW) + qPadX*2
	maxBadgeW := float32(c.ImageRect.W) - 10
	if qBadgeW > maxBadgeW {
		qBadgeW = maxBadgeW
	}
	qx := c.ImageRect.X + (c.ImageRect.W-float64(qBadgeW))/2
	qy := c.ImageRect.Y + c.ImageRect.H - float64(qBadgeH) - 8
	drawUICardRect(screen,
		gameui.Rect{X: qx, Y: qy, W: float64(qBadgeW), H: float64(qBadgeH)},
		color.RGBA{28, 20, 10, 232},
		color.RGBA{214, 176, 92, 235},
		1,
	)
	vector.StrokeRect(screen,
		float32(qx)+1, float32(qy)+1, qBadgeW-2, qBadgeH-2, 1,
		color.RGBA{255, 232, 170, 72}, false,
	)
	DrawTextCentered(screen, qLabel, qx+float64(qBadgeW)/2, qy+2, FaceSmall, color.RGBA{255, 238, 188, 250})
}
