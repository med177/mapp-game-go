package render

import (
	"fmt"
	"image/color"
	"strings"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const armyTaskStatusIconStep = armyIconStep

func (r *Renderer) armyIconStepForTaskStatus(aids []army.ArmyID, fallback float32) float32 {
	if r == nil || r.gs == nil {
		return fallback
	}
	for _, aid := range aids {
		if armyTaskStatusVisible(r.gs, r.gs.Armies[aid]) {
			return armyTaskStatusIconStep
		}
	}
	return fallback
}

func armyTaskStatusBadgeRect(cx, cy float32) gameui.Rect {
	// Kara görev rozetleri donanma görev rozetleriyle aynı sağ-üst anchor,
	// boyut ve hit-test geometrisini paylaşır.
	return navalMissionBonusBadgeRect(cx, cy)
}

func activeRaidForArmy(gs *state.GameState, a *army.Army) *state.RaidState {
	if gs == nil || a == nil || a.IsNaval || gs.Raids == nil || gs.PlayerFactionID == "" || gs.ArmyHiddenFrom(a, gs.PlayerFactionID) {
		return nil
	}
	raid := gs.Raids[a.RegionID]
	if raid == nil || raid.Turn != gs.Turn || string(raid.RaiderFactionID) != a.OwnerID {
		return nil
	}
	// RaiderArmyID yeni kayıtlar için kesin eşleşmedir. Eski kayıtlar bu alanı
	// taşımadığından, aynı bölgede tek geçerli oyuncu ordusu varsayımıyla geriye
	// dönük görünürlük korunur.
	if raid.RaiderArmyID != "" && raid.RaiderArmyID != a.ID {
		return nil
	}
	return raid
}

func armyTaskStatusVisible(gs *state.GameState, a *army.Army) bool {
	if gs == nil || a == nil || a.IsNaval || gs.PlayerFactionID == "" || gs.ArmyHiddenFrom(a, gs.PlayerFactionID) {
		return false
	}
	return a.InAmbush || activeRaidForArmy(gs, a) != nil
}

func drawArmyTaskStatusBadge(screen *ebiten.Image, gs *state.GameState, a *army.Army, cx, cy float32) {
	if !armyTaskStatusVisible(gs, a) {
		return
	}
	badge := armyTaskStatusBadgeRect(cx, cy)
	badgeCX := float32(badge.X + badge.W/2)
	badgeCY := float32(badge.Y + badge.H/2)
	if a.InAmbush {
		vector.FillCircle(screen, badgeCX, badgeCY, 10, color.RGBA{8, 8, 8, 245}, false)
		vector.FillCircle(screen, badgeCX, badgeCY, 8.5, color.RGBA{156, 156, 156, 245}, false)
		return
	}
	drawGoldPlusBadge(screen, cx, cy, "+")
}

func (r *Renderer) drawArmyTaskStatusBadges(screen *ebiten.Image, positions []armyIconPos) {
	if r == nil || r.gs == nil {
		return
	}
	for _, pos := range positions {
		a := r.gs.Armies[pos.ArmyID]
		drawArmyTaskStatusBadge(screen, r.gs, a, pos.X, pos.Y)
	}
}

func (r *Renderer) armyTaskStatusBadgeHitAt(mx, my float64) (army.ArmyID, bool) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return "", false
	}
	positions := r.armyIconPositions()
	for i := len(positions) - 1; i >= 0; i-- {
		pos := positions[i]
		if !armyTaskStatusVisible(r.gs, r.gs.Armies[pos.ArmyID]) {
			continue
		}
		if armyTaskStatusBadgeRect(pos.X, pos.Y).Hit(mx, my) {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func raidLootParts(loot state.RegionProductionSummary) []string {
	parts := make([]string, 0, 7)
	if loot.Gold > 0 {
		parts = append(parts, "+"+itoa(loot.Gold)+" altın")
	}
	if loot.Grain > 0 {
		parts = append(parts, "+"+itoa(loot.Grain)+" tahıl")
	}
	if loot.Iron > 0 {
		parts = append(parts, "+"+itoa(loot.Iron)+" demir")
	}
	if loot.Timber > 0 {
		parts = append(parts, "+"+itoa(loot.Timber)+" kereste")
	}
	if loot.Stone > 0 {
		parts = append(parts, "+"+itoa(loot.Stone)+" taş")
	}
	if loot.Spice > 0 {
		parts = append(parts, "+"+itoa(loot.Spice)+" baharat")
	}
	if loot.Cloth > 0 {
		parts = append(parts, "+"+itoa(loot.Cloth)+" kumaş")
	}
	return parts
}

func armyTaskStatusTooltipText(gs *state.GameState, a *army.Army) (string, string, bool) {
	if !armyTaskStatusVisible(gs, a) {
		return "", "", false
	}
	if a.InAmbush {
		return "Pusu", "Ordu bu bölgede düşmandan gizleniyor. Düşman buraya hareket ederse temas otomatik çatışmaya dönüşür ve pusu bonusu uygulanır.", true
	}
	raid := activeRaidForArmy(gs, a)
	if raid == nil {
		return "", "", false
	}
	region := gs.Regions[raid.RegionID]
	if region == nil {
		return "Yağmalama", "Bu tur yağma emri uygulandı.", true
	}
	regionName := region.NameTR
	if regionName == "" {
		regionName = region.Name
	}
	if regionName == "" {
		regionName = "Bölge"
	}
	parts := raidLootParts(gs.RaidLootPreview(region))
	detail := fmt.Sprintf("%s bölgesinde bu tur yağma yapıldı. Kazanç: %s.", regionName, strings.Join(parts, ", "))
	if len(parts) == 0 {
		detail = fmt.Sprintf("%s bölgesinde bu tur yağma yapıldı; aktarılabilir gelir veya üretim oluşmadı.", regionName)
	}
	return "Yağmalama", detail, true
}

func (r *Renderer) drawArmyTaskStatusHoverTooltip(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.mapMode == MapModeTrade {
		return
	}
	mx, my := ebiten.CursorPosition()
	aid, ok := r.armyTaskStatusBadgeHitAt(float64(mx), float64(my))
	if !ok {
		return
	}
	title, detail, ok := armyTaskStatusTooltipText(r.gs, r.gs.Armies[aid])
	if !ok {
		return
	}
	const tooltipW = 390.0
	const tooltipH = 128.0
	x, y, w, h := tooltipRect(float64(mx), float64(my), tooltipW, tooltipH)
	drawTooltipBox(screen, x, y, w, h)
	drawUILabel(screen, gameui.Rect{X: x + 12, Y: y + 9, W: w - 24, H: 20}, title, ColorGold, gameui.TextMedium, gameui.TextAlignStart)
	drawUIWrappedLabel(screen, gameui.Rect{X: x + 12, Y: y + 34, W: w - 24, H: h - 42}, detail, ColorWhite, gameui.TextSmall, 17, 5)
}
