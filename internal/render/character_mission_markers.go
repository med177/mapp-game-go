package render

import (
	"image/color"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const characterMissionMapMarkerSize = 28.0
const characterMissionMapMarkerLabelGap = 6.0

type characterMissionMapMarker struct {
	regionID       world.RegionID
	count          int
	characterID    string
	missionID      string
	turnsRemaining int
}

func characterMissionMapMarkerRect(cx, cy float64) gameui.Rect {
	size := characterMissionMapMarkerSize
	return gameui.Rect{X: cx - size/2, Y: cy - size/2, W: size, H: size}
}

func characterMissionMapTargetRegion(gs *state.GameState, mission scenario.CharacterMissionDefinition, targetID string) *world.Region {
	if gs == nil || targetID == "" {
		return nil
	}
	switch mission.TargetKind {
	case scenario.CharacterMissionTargetRegion:
		return gs.Regions[world.RegionID(targetID)]
	case scenario.CharacterMissionTargetFaction:
		targetFactionID := faction.FactionID(targetID)
		if region, _, _, ok := gs.FactionCapital(targetFactionID); ok {
			return region
		}
		var fallback *world.Region
		for _, region := range gs.Regions {
			if region == nil || region.OwnerID != targetID || region.IsSea || region.IsTerrainArea || region.IsLocked {
				continue
			}
			if fallback == nil || region.ID < fallback.ID {
				fallback = region
			}
		}
		return fallback
	default:
		return nil
	}
}

func (r *Renderer) characterMissionMapMarkers() []characterMissionMapMarker {
	if r == nil || r.gs == nil || r.gs.PlayerFactionID == "" || r.gs.Phase == state.PhaseEditMode {
		return nil
	}
	if r.characterMissionMarkerCacheValid && r.characterMissionMarkerFrame == r.renderFrame {
		return r.characterMissionMarkerBuf
	}
	markers := r.characterMissionMarkerBuf[:0]
	for _, character := range r.gs.Commanders {
		if character == nil || character.OwnerID != string(r.gs.PlayerFactionID) ||
			character.ActiveMissionID == "" || character.ActiveMissionTargetID == "" ||
			character.ActiveMissionTurnsRemaining <= 0 {
			continue
		}
		mission, ok := characterMissionDefinition(r.gs, character.ActiveMissionID)
		if !ok {
			continue
		}
		region := characterMissionMapTargetRegion(r.gs, mission, character.ActiveMissionTargetID)
		if region == nil {
			continue
		}

		index, insertAt := -1, len(markers)
		for i := range markers {
			if markers[i].regionID == region.ID {
				index = i
				break
			}
			if markers[i].regionID > region.ID {
				insertAt = i
				break
			}
		}
		if index >= 0 {
			markers[index].count++
			if character.ID < markers[index].characterID {
				markers[index].characterID = character.ID
				markers[index].missionID = mission.ID
				markers[index].turnsRemaining = character.ActiveMissionTurnsRemaining
			}
			continue
		}
		markers = append(markers, characterMissionMapMarker{})
		copy(markers[insertAt+1:], markers[insertAt:])
		markers[insertAt] = characterMissionMapMarker{
			regionID: region.ID, count: 1, characterID: character.ID,
			missionID: mission.ID, turnsRemaining: character.ActiveMissionTurnsRemaining,
		}
	}
	r.characterMissionMarkerBuf = markers
	r.characterMissionMarkerFrame = r.renderFrame
	r.characterMissionMarkerCacheValid = true
	return markers
}

func characterMissionDefinition(gs *state.GameState, missionID string) (scenario.CharacterMissionDefinition, bool) {
	if gs == nil || missionID == "" {
		return scenario.CharacterMissionDefinition{}, false
	}
	for _, mission := range gs.CouncilConfig.Missions {
		if mission.ID == missionID {
			return mission, true
		}
	}
	return scenario.CharacterMissionDefinition{}, false
}

func (r *Renderer) characterMissionMapMarkerCenter(regionID world.RegionID) (float64, float64, bool) {
	if r == nil || r.gs == nil {
		return 0, 0, false
	}
	region := r.gs.Regions[regionID]
	if region == nil {
		return 0, 0, false
	}
	primaryIndex := region.PrimarySettlementIndex()
	for _, label := range r.regionLabelStatic() {
		if label.Region == nil || label.Region.ID != regionID || label.Index != primaryIndex {
			continue
		}
		sx, sy := r.worldToScreen(label.WX, label.WY)
		cx, cy := characterMissionMapMarkerCenterAfterLabel(sx-label.W/2, sy+16, label.W, label.H)
		return cx, cy, true
	}
	x, y := r.regionScreenPos(region)
	return x + 22, y - 20, true
}

func characterMissionMapMarkerCenterAfterLabel(labelX, labelY, labelW, labelH float64) (float64, float64) {
	return labelX + labelW + characterMissionMapMarkerLabelGap + characterMissionMapMarkerSize/2,
		labelY + labelH/2
}

func (r *Renderer) drawCharacterMissionMapMarkers(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return
	}
	for _, marker := range r.characterMissionMapMarkers() {
		cx, cy, ok := r.characterMissionMapMarkerCenter(marker.regionID)
		if !ok || cx < -characterMissionMapMarkerSize || cx > ScreenWidth+characterMissionMapMarkerSize ||
			cy < -characterMissionMapMarkerSize || cy > ScreenHeight+characterMissionMapMarkerSize {
			continue
		}
		rect := characterMissionMapMarkerRect(cx, cy)
		vector.FillCircle(screen, float32(cx), float32(cy), float32(characterMissionMapMarkerSize/2), color.RGBA{38, 28, 12, 248}, true)
		vector.StrokeCircle(screen, float32(cx), float32(cy), float32(characterMissionMapMarkerSize/2-1), 2, color.RGBA{255, 214, 82, 255}, true)
		if !gameui.DrawIcon(screen, gameui.IconBook, rect.X+6, rect.Y+6, 16, color.RGBA{255, 238, 174, 255}) {
			DrawTextCentered(screen, "G", cx, cy-5, FaceSmall, color.RGBA{255, 238, 174, 255})
		}
		if marker.count > 1 {
			badgeX, badgeY := float32(rect.X+rect.W-2), float32(rect.Y+rect.H-2)
			vector.FillCircle(screen, badgeX, badgeY, 7, color.RGBA{117, 48, 24, 255}, true)
			drawMarkerBadgeText(screen, itoa(marker.count), float64(badgeX), float64(badgeY), 14, color.RGBA{255, 246, 224, 255})
		}
	}
}

func (r *Renderer) drawCharacterMissionMarkerHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return
	}
	mx, my := ebiten.CursorPosition()
	if _, ok := r.uiLayers.TopAt(float64(mx), float64(my)); ok {
		return
	}
	for _, marker := range r.characterMissionMapMarkers() {
		cx, cy, ok := r.characterMissionMapMarkerCenter(marker.regionID)
		if !ok || !characterMissionMapMarkerRect(cx, cy).Hit(float64(mx), float64(my)) {
			continue
		}
		character := r.gs.Commanders[marker.characterID]
		mission, missionOK := characterMissionDefinition(r.gs, marker.missionID)
		if character == nil || !missionOK {
			return
		}
		detail := character.Name + ": " + mission.NameTR + " · " + itoa(marker.turnsRemaining) + " tur kaldı"
		if marker.count > 1 {
			detail += " · +" + itoa(marker.count-1) + " başka görev"
		}
		x, y, w, h := tooltipRect(float64(mx), float64(my), 330, 76)
		drawTooltipBox(screen, x, y, w, h)
		drawUILabel(screen, gameui.Rect{X: x + 12, Y: y + 8, W: w - 24, H: 20},
			"Karakter görevi", ColorGold, gameui.TextMedium, gameui.TextAlignStart)
		drawUIWrappedLabel(screen, gameui.Rect{X: x + 12, Y: y + 32, W: w - 24, H: h - 38},
			detail, ColorWhite, gameui.TextSmall, 16, 2)
		return
	}
}
