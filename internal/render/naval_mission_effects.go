package render

import (
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var escortShieldPath vector.Path

// navalMissionReachedRegion, görev bonusunun artık uygulanabileceği hedef
// konumunu döndürür. Hedefe ulaşmamış görevler haritada yalnızca filo rozetiyle
// görünür; bu etiket hedefteki gerçek etkiyi temsil eder.
func navalMissionReachedRegion(gs *state.GameState, fleet *army.Army) *world.Region {
	if gs == nil || fleet == nil || !fleet.IsNaval || fleet.NavalMission == nil {
		return nil
	}

	mission := fleet.NavalMission
	switch mission.Kind {
	case army.NavalMissionPatrol, army.NavalMissionBlockade:
		if !fleet.IsAtSea() || fleet.RegionID != mission.TargetRegionID {
			return nil
		}
		region := gs.Regions[mission.TargetRegionID]
		if region == nil || !region.IsSea {
			return nil
		}
		return region
	case army.NavalMissionEscort:
		escortTarget := gs.Armies[mission.TargetFleetID]
		if escortTarget == nil || !fleet.IsAtSea() || !escortTarget.IsAtSea() ||
			fleet.RegionID != escortTarget.RegionID {
			return nil
		}
		region := gs.Regions[fleet.RegionID]
		if region == nil || !region.IsSea {
			return nil
		}
		return region
	case army.NavalMissionTransport:
		// Nakliye görevi kara bölgesine girince otomatik çıkarma yaptığı için
		// görev state'i çoğu zaman aynı çözümleme içinde temizlenir. Buradaki
		// işaret, çıkarma öncesi hedef kıyının deniz komşusuna varılan kısa
		// bekleme durumunu da görünür kılar.
		targetLand := gs.Regions[mission.TargetRegionID]
		if targetLand == nil || targetLand.IsSea || !fleet.IsAtSea() || len(fleet.EmbarkedUnits) == 0 {
			return nil
		}
		for _, neighborID := range targetLand.Neighbors {
			if neighborID != fleet.RegionID {
				continue
			}
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && neighbor.IsSea {
				return targetLand
			}
		}
	}
	return nil
}

func navalMissionReachedLabel(kind army.NavalMissionKind) string {
	switch kind {
	case army.NavalMissionPatrol:
		return "HEDEFTE • DEVRİYE: 1 AB. GEMİ"
	case army.NavalMissionBlockade:
		return "HEDEFTE • ABLUKA: -%50/GEMİ"
	case army.NavalMissionEscort:
		return "HEDEFTE • ESCORT: +%15 SAV."
	case army.NavalMissionTransport:
		return "HEDEFTE • NAKLİYE: ÇIKARMA"
	default:
		return "HEDEFTE • DONANMA GÖREVİ"
	}
}

func navalMissionReachedColor(kind army.NavalMissionKind) color.RGBA {
	switch kind {
	case army.NavalMissionPatrol:
		return color.RGBA{72, 185, 218, 245}
	case army.NavalMissionBlockade:
		return color.RGBA{232, 78, 78, 245}
	case army.NavalMissionEscort:
		return color.RGBA{128, 150, 72, 245}
	case army.NavalMissionTransport:
		return color.RGBA{232, 154, 54, 245}
	default:
		return ColorGold
	}
}

// navalMissionBonusBadgeRect, hedefte bonus kazanan filonun ikon yanındaki
// küçük dairesel rozeti için çizim, hover ve input geometry'sini paylaşır.
func navalMissionBonusBadgeRect(cx, cy float32) gameui.Rect {
	const badgeSize = 20.0
	centerX, centerY := navalUpperRightBadgeCenter(cx, cy)
	return gameui.Rect{X: centerX - badgeSize/2, Y: centerY - badgeSize/2, W: badgeSize, H: badgeSize}
}

// navalMissionPendingBadge, görevi atanmış ancak henüz görev konumuna
// ulaşmamış oyuncu filosunu gösterir. Hedefteki bonus rozetiyle aynı anchor'ı
// paylaşır; böylece görev durumu filo marker'ından bağımsız ama tutarlı kalır.
func navalMissionPendingBadge(gs *state.GameState, fleet *army.Army) bool {
	return gs != nil && fleet != nil && fleet.IsNaval &&
		fleet.OwnerID == string(gs.PlayerFactionID) && fleet.NavalMission != nil &&
		navalMissionReachedRegion(gs, fleet) == nil
}

func navalMissionPendingBadgeRect(cx, cy float32) gameui.Rect {
	return navalMissionBonusBadgeRect(cx, cy)
}

// merchantTradeBonusBadgeRect, ticaret rotası rozetini diğer deniz görevi
// rozetleriyle aynı üst-sağ anchor'a bağlar. Çizim, hit-test ve hover bu
// geometry'yi birlikte kullanır.
func merchantTradeBonusBadgeRect(cx, cy float32) gameui.Rect {
	return navalMissionBonusBadgeRect(cx, cy)
}

func appendEscortShieldPath(path *vector.Path, cx, cy, halfWidth, sideBottom, bottom float32) {
	path.Reset()
	path.MoveTo(cx-halfWidth, cy-halfWidth)
	path.LineTo(cx+halfWidth, cy-halfWidth)
	path.LineTo(cx+halfWidth, cy+sideBottom)
	path.QuadTo(cx+halfWidth, cy+sideBottom+4, cx+halfWidth*0.62, cy+bottom-2)
	path.QuadTo(cx+halfWidth*0.30, cy+bottom, cx, cy+bottom)
	path.QuadTo(cx-halfWidth*0.30, cy+bottom, cx-halfWidth*0.62, cy+bottom-2)
	path.QuadTo(cx-halfWidth, cy+sideBottom+4, cx-halfWidth, cy+sideBottom)
	path.Close()
}

func drawNavalMissionEscortBadge(screen *ebiten.Image, cx, cy float32, badgeColor color.RGBA) {
	appendEscortShieldPath(&escortShieldPath, cx, cy, 10, -2, 10)
	var drawOptions vector.DrawPathOptions
	drawOptions.AntiAlias = true
	drawOptions.ColorScale.ScaleWithColor(color.RGBA{22, 24, 30, 245})
	vector.FillPath(screen, &escortShieldPath, nil, &drawOptions)

	appendEscortShieldPath(&escortShieldPath, cx, cy, 8.5, -1.5, 8.5)
	drawOptions.ColorScale.Reset()
	drawOptions.ColorScale.ScaleWithColor(badgeColor)
	vector.FillPath(screen, &escortShieldPath, nil, &drawOptions)
}

func navalMissionBonusBadgeTextColor(kind army.NavalMissionKind) color.RGBA {
	text := color.RGBA{35, 25, 15, 255}
	if kind == army.NavalMissionBlockade {
		text = color.RGBA{255, 255, 255, 255}
	}
	return text
}

func navalMissionWarshipCount(gs *state.GameState, fleet *army.Army) int {
	if gs == nil || fleet == nil {
		return 0
	}
	count := 0
	for _, unit := range fleet.Units {
		if unitType := gs.UnitTypes[unit.TypeID]; unitType != nil && unitType.Category == army.CategoryNavalWar {
			count++
		}
	}
	return count
}

// navalMissionBonusBadge, yalnız hedefe ulaşıp gerçek bonus üreten filolara
// küçük ikon rozeti verir. Büyük bölge marker'ı yerine bu bilgi filo üzerinde
// tutulur; ayrıntı hover tooltip'inde gösterilir.
func navalMissionBonusBadge(gs *state.GameState, fleet *army.Army) (string, color.RGBA, bool) {
	if gs == nil || fleet == nil || fleet.OwnerID != string(gs.PlayerFactionID) || navalMissionReachedRegion(gs, fleet) == nil {
		return "", color.RGBA{}, false
	}
	switch fleet.NavalMission.Kind {
	case army.NavalMissionPatrol:
		return "+" + itoa(navalMissionWarshipCount(gs, fleet)), navalMissionReachedColor(fleet.NavalMission.Kind), true
	case army.NavalMissionBlockade:
		percent := navalMissionWarshipCount(gs, fleet) * 50
		if percent > 100 {
			percent = 100
		}
		return itoa(percent), navalMissionReachedColor(fleet.NavalMission.Kind), true
	case army.NavalMissionEscort:
		return "15", navalMissionReachedColor(fleet.NavalMission.Kind), true
	default:
		return "", color.RGBA{}, false
	}
}

func (r *Renderer) navalMissionBonusHitAt(mx, my float64) (army.ArmyID, bool) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return "", false
	}
	positions := r.armyIconPositions()
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		fleet := r.gs.Armies[pos.ArmyID]
		if _, _, ok := navalMissionBonusBadge(r.gs, fleet); !ok {
			continue
		}
		if navalMissionBonusBadgeRect(pos.X, pos.Y).Hit(mx, my) {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func (r *Renderer) navalMissionPendingHitAt(mx, my float64) (army.ArmyID, bool) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return "", false
	}
	positions := r.armyIconPositions()
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		fleet := r.gs.Armies[pos.ArmyID]
		if !navalMissionPendingBadge(r.gs, fleet) {
			continue
		}
		if navalMissionPendingBadgeRect(pos.X, pos.Y).Hit(mx, my) {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func (r *Renderer) merchantTradeBonusHitAt(mx, my float64) (army.ArmyID, bool) {
	if r == nil || r.gs == nil {
		return "", false
	}
	if r.mapMode == MapModeTrade && !r.tradeOverlayVisible() {
		return "", false
	}
	positions := r.armyIconPositions()
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		fleet := r.gs.Armies[pos.ArmyID]
		if r.merchantTradeBonusForArmy(fleet) <= 0 {
			continue
		}
		if r.mapMode == MapModeTrade {
			if _, _, connected := r.tradeRouteConnectionPoint(fleet.TradeRouteKey, float64(pos.X), float64(pos.Y)); !connected {
				continue
			}
		}
		if merchantTradeBonusBadgeRect(pos.X, pos.Y).Hit(mx, my) {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func navalMissionBonusTooltipText(gs *state.GameState, fleet *army.Army) (string, string, bool) {
	badge, _, ok := navalMissionBonusBadge(gs, fleet)
	if !ok || fleet.NavalMission == nil {
		return "", "", false
	}
	target := navalMissionReachedRegion(gs, fleet)
	targetName := "hedef bölgede"
	if target != nil {
		if target.NameTR != "" {
			targetName = target.NameTR
		} else if target.Name != "" {
			targetName = target.Name
		}
	}
	switch fleet.NavalMission.Kind {
	case army.NavalMissionPatrol:
		return "Devriye bonusu", targetName + " denizinde " + badge + " görevli düşman abluka gemisini dengeliyor.", true
	case army.NavalMissionBlockade:
		lootGold := gs.BlockadeLootGoldForFleet(fleet)
		return "Abluka bonusu", targetName + " denizinde ticarete toplam -%" + badge + " kesinti uyguluyor. Gelir katkısı (ganimet): +" + itoa(lootGold) + " altın/tur.", true
	case army.NavalMissionEscort:
		return "Escort bonusu", targetName + " denizindeki nakliye filosuna +%15 deniz savunması veriyor.", true
	default:
		return "", "", false
	}
}

func merchantTradeBonusTooltipText(gs *state.GameState, fleet *army.Army) (string, string, bool) {
	if gs == nil || fleet == nil {
		return "", "", false
	}
	route := merchantRouteForKey(gs, fleet.TradeRouteKey)
	bonus := gs.MerchantFleetTradeRouteBonus(fleet, route)
	if route == nil || bonus <= 0 {
		return "", "", false
	}
	income := bonus * route.GoldPerUnit
	if route.BlockadePercent > 0 {
		income = income * (economy.MaxTradeRouteBlockadePercent - route.BlockadePercent) / economy.MaxTradeRouteBlockadePercent
	}
	detail := "Rota bonusu: +" + itoa(bonus) + " mal/tur • Gelir katkısı: +" + itoa(income) + " altın/tur"
	return "Ticaret rotası bonusu", detail, true
}

func navalEmbarkedArmyTooltipText(fleet *army.Army) (string, string, bool) {
	if fleet == nil || !fleet.IsNaval || len(fleet.EmbarkedUnits) == 0 {
		return "", "", false
	}
	return "Nakliye Görevi", "Taşınan ordu " + itoa(len(fleet.EmbarkedUnits)) + " birim", true
}

// drawNavalMissionBonusHoverTooltip, yalnız küçük bonus rozeti üzerine
// gelindiğinde ayrıntıyı gösterir; haritanın üzerinde kalıcı büyük yazı çizmez.
func (r *Renderer) drawNavalMissionBonusHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return
	}
	mx, my := ebiten.CursorPosition()
	aid, ok := r.navalMissionBonusHitAt(float64(mx), float64(my))
	if !ok {
		return
	}
	title, detail, ok := navalMissionBonusTooltipText(r.gs, r.gs.Armies[aid])
	if !ok {
		return
	}
	const tooltipW = 360.0
	const tooltipH = 82.0
	x, y, w, h := tooltipRect(float64(mx), float64(my), tooltipW, tooltipH)
	drawTooltipBox(screen, x, y, w, h)
	drawUILabel(screen, gameui.Rect{X: x + 12, Y: y + 9, W: w - 24, H: 20}, title, ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: x + 12, Y: y + 34, W: w - 24, H: h - 42}, detail, ColorWhite, gameui.TextSmall, 17, 2)
}

func (r *Renderer) drawMerchantTradeBonusHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || (r.mapMode == MapModeTrade && !r.tradeOverlayVisible()) {
		return
	}
	mx, my := ebiten.CursorPosition()
	aid, ok := r.merchantTradeBonusHitAt(float64(mx), float64(my))
	if !ok {
		return
	}
	title, detail, ok := merchantTradeBonusTooltipText(r.gs, r.gs.Armies[aid])
	if !ok {
		return
	}
	const tooltipW = 360.0
	const tooltipH = 82.0
	x, y, w, h := tooltipRect(float64(mx), float64(my), tooltipW, tooltipH)
	drawTooltipBox(screen, x, y, w, h)
	drawUILabel(screen, gameui.Rect{X: x + 12, Y: y + 9, W: w - 24, H: 20}, title, ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: x + 12, Y: y + 34, W: w - 24, H: h - 42}, detail, ColorWhite, gameui.TextSmall, 17, 2)
}

func (r *Renderer) drawNavalEmbarkedArmyHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade || r.gs.PlayerFactionID == "" {
		return
	}
	mx, my := ebiten.CursorPosition()
	aid, ok := r.embarkedArmyHitAt(float64(mx), float64(my))
	if !ok {
		return
	}
	fleet := r.gs.Armies[aid]
	if fleet == nil || fleet.OwnerID != string(r.gs.PlayerFactionID) {
		return
	}
	title, detail, ok := navalEmbarkedArmyTooltipText(fleet)
	if !ok {
		return
	}
	const tooltipW = 360.0
	const tooltipH = 82.0
	x, y, w, h := tooltipRect(float64(mx), float64(my), tooltipW, tooltipH)
	drawTooltipBox(screen, x, y, w, h)
	drawUILabel(screen, gameui.Rect{X: x + 12, Y: y + 9, W: w - 24, H: 20}, title, ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: x + 12, Y: y + 34, W: w - 24, H: h - 42}, detail, ColorWhite, gameui.TextSmall, 17, 2)
}
