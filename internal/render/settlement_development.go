package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	settlementDevelopmentBarWidth    = 62.0
	settlementDevelopmentBarHeight   = 7.0
	settlementDevelopmentIconGap     = 4.0
	settlementDevelopmentIconSize    = 16.0
	settlementDevelopmentLayoutWidth = 82.0
)

func settlementLabelLayout(labelWidth, labelHeight float64, development bool) (float64, float64) {
	if !development {
		return labelWidth, labelHeight
	}
	if labelWidth < settlementDevelopmentLayoutWidth {
		labelWidth = settlementDevelopmentLayoutWidth
	}
	return labelWidth, labelHeight + 2 + settlementDevelopmentBarHeight
}

func settlementBuildingDevelopment(gs *state.GameState, region *world.Region) float64 {
	if gs == nil || region == nil {
		return 0
	}

	builtLevels, maxLevels := 0, 0
	for buildingID, building := range gs.BuildingTypes {
		if building == nil || !region.AllowsBuilding(buildingID, building.MinorRegions) {
			continue
		}
		maxLevel := gs.BuildingLevelCap(region, buildingID)
		if maxLevel <= 0 {
			continue
		}
		maxLevels += maxLevel

		level := 0
		for _, builtID := range region.Buildings {
			if builtID == buildingID {
				level++
			}
		}
		if level > maxLevel {
			level = maxLevel
		}
		builtLevels += level
	}
	if maxLevels == 0 {
		return 0
	}
	return float64(builtLevels) / float64(maxLevels)
}

func hasActiveBuildingUpgrade(gs *state.GameState, region *world.Region) bool {
	if gs == nil || region == nil || gs.PlayerFactionID == "" || region.OwnerID != string(gs.PlayerFactionID) {
		return false
	}
	for _, order := range gs.ProductionQueue {
		if order.Kind == "building" &&
			order.FactionID == string(gs.PlayerFactionID) &&
			order.RegionID == region.ID &&
			order.TurnsLeft > 0 {
			return true
		}
	}
	return false
}

func hasAffordableBuildingOption(gs *state.GameState, region *world.Region) bool {
	if !regionBuildingActionsAvailable(gs, region) || gs.SiegeAt(region.ID) != nil {
		return false
	}
	faction := gs.Factions[gs.PlayerFactionID]
	if faction == nil {
		return false
	}
	for buildingID, building := range gs.BuildingTypes {
		if building == nil || !buildingVisibleByRegionRules(gs, region, buildingID, building) {
			continue
		}
		maxLevel := gs.BuildingLevelCap(region, buildingID)
		if maxLevel <= region.BuildingLevel(buildingID) {
			continue
		}
		queued := false
		for _, order := range gs.ProductionQueue {
			if order.Kind == "building" && order.RegionID == region.ID && order.TypeID == buildingID {
				queued = true
				break
			}
		}
		if queued {
			continue
		}
		if buildingCost(gs, region, building).CanAfford(faction) {
			return true
		}
	}
	return false
}

func settlementDevelopmentStateKey(gs *state.GameState) uint64 {
	if gs == nil || gs.PlayerFactionID == "" {
		return 0
	}
	key := borderHashString(string(gs.PlayerFactionID))
	for regionID, region := range gs.Regions {
		if region == nil || region.OwnerID != string(gs.PlayerFactionID) {
			continue
		}
		for _, settlement := range region.Settlements {
			if settlement.IsCenter {
				key ^= borderHashString(string(regionID))
				if hasActiveBuildingUpgrade(gs, region) {
					key ^= borderHashString(string(regionID) + "|building")
				}
				break
			}
		}
	}
	return key
}

func settlementDevelopmentBackground(gs *state.GameState, region *world.Region) color.RGBA {
	if hasAffordableBuildingOption(gs, region) {
		return color.RGBA{190, 45, 38, 255}
	}
	return color.RGBA{0, 0, 0, 255}
}

func (r *Renderer) drawSettlementDevelopmentBar(screen *ebiten.Image, region *world.Region, centerX, y float32) {
	if screen == nil {
		return
	}
	progress := settlementBuildingDevelopment(r.gs, region)
	x := centerX - settlementDevelopmentBarWidth/2
	vector.FillRect(screen, x, y, settlementDevelopmentBarWidth, settlementDevelopmentBarHeight, settlementDevelopmentBackground(r.gs, region), false)
	if progress > 0 {
		fillWidth := float32(settlementDevelopmentBarWidth) * float32(progress)
		vector.FillRect(screen, x, y, fillWidth, settlementDevelopmentBarHeight, color.RGBA{35, 190, 66, 255}, false)
	}
	vector.StrokeRect(screen, x, y, settlementDevelopmentBarWidth, settlementDevelopmentBarHeight, 1, color.RGBA{0, 0, 0, 255}, false)
	if hasActiveBuildingUpgrade(r.gs, region) {
		drawSettlementWorkingIcon(screen, centerX+settlementDevelopmentBarWidth/2+settlementDevelopmentIconGap, y+(settlementDevelopmentBarHeight-settlementDevelopmentIconSize)/2)
	}
}

func drawSettlementWorkingIcon(screen *ebiten.Image, x, y float32) {
	if screen == nil {
		return
	}
	ensureSettlementMarkerSprites()
	if settlementMarkerWorkingSprite == nil {
		return
	}
	bounds := settlementMarkerWorkingSprite.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return
	}
	op := newLinearImageDrawOptions()
	op.GeoM.Scale(settlementDevelopmentIconSize/float64(bounds.Dx()), settlementDevelopmentIconSize/float64(bounds.Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(settlementMarkerWorkingSprite, op)
}
