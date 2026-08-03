package game

import (
	"fmt"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

func (g *Game) raidRegion(armyID army.ArmyID, regionID world.RegionID) {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	a := g.gs.Armies[armyID]
	region := g.gs.Regions[regionID]
	if reason := g.gs.RaidBlockReason(a, region); reason != "" {
		g.renderer.ShowCombatResult(reason)
		return
	}
	if !g.gs.ApplyRaid(a, region) {
		g.renderer.ShowCombatResult("Yağmalama emri uygulanamadı.")
		return
	}
	g.renderer.MarkMapDirty()
	g.renderer.ShowCombatResult(fmt.Sprintf("%s bu tur yağmalandı. Vergi gelirinin %%%d'i ve üretimlerinin %%%d'i yağmalayana aktarılacak.", region.NameTR, 80, 50))
}

func (g *Game) setArmyAmbush(armyID army.ArmyID, regionID world.RegionID) {
	if g == nil || g.gs == nil || g.renderer == nil {
		return
	}
	a := g.gs.Armies[armyID]
	region := g.gs.Regions[regionID]
	if reason := g.gs.AmbushBlockReason(a, region); reason != "" {
		g.renderer.ShowCombatResult(reason)
		return
	}
	if !g.gs.SetAmbush(a, region) {
		g.renderer.ShowCombatResult("Pusu emri uygulanamadı.")
		return
	}
	g.renderer.MarkMapDirty()
	g.renderer.ShowCombatResult("Ordu pusuya yattı; düşman tarafından gizlendi.")
}
