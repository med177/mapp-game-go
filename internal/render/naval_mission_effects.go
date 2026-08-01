package render

import (
	"image/color"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
)

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
		return color.RGBA{220, 178, 62, 245}
	case army.NavalMissionTransport:
		return color.RGBA{232, 154, 54, 245}
	default:
		return ColorGold
	}
}

// navalMissionBonusBadgeRect, hedefte bonus kazanan filonun ikon yanındaki
// küçük dairesel rozeti için çizim, hover ve input geometry'sini paylaşır.
func navalMissionBonusBadgeRect(cx, cy float32) gameui.Rect {
	return gameui.Rect{X: float64(cx + 25), Y: float64(cy - 27), W: 20, H: 20}
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
		return "Devriye bonusu", targetName + " denizinde " + badge + " düşman abluka gemisini dengeliyor.", true
	case army.NavalMissionBlockade:
		return "Abluka bonusu", targetName + " denizinde ticarete toplam -%" + badge + " kesinti uyguluyor.", true
	case army.NavalMissionEscort:
		return "Escort bonusu", targetName + " denizindeki nakliye filosuna +%15 deniz savunması veriyor.", true
	default:
		return "", "", false
	}
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
