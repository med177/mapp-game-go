package game

import (
	"fmt"

	"mapp-game-go/internal/ai"
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/render"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

type siegeTurnUpdate struct {
	Message string
	Detail  string
	Popup   bool
}

func ensureSiegeMap(gs *state.GameState) {
	if gs != nil && gs.Sieges == nil {
		gs.Sieges = make(map[world.RegionID]*state.SiegeState)
	}
}

func gameFactionsAtWar(gs *state.GameState, attackerOwnerID, defenderOwnerID string) bool {
	if gs == nil || attackerOwnerID == "" || defenderOwnerID == "" || attackerOwnerID == defenderOwnerID {
		return false
	}
	key := faction.RelationKey(faction.FactionID(attackerOwnerID), faction.FactionID(defenderOwnerID))
	rel, ok := gs.Relations[key]
	return ok && rel != nil && rel.Stance == faction.StanceWar
}

func regionsAdjacent(gs *state.GameState, from, to world.RegionID) bool {
	if gs == nil || from == "" || to == "" {
		return false
	}
	region := gs.Regions[from]
	if region == nil {
		return false
	}
	for _, neighbor := range region.Neighbors {
		if neighbor == to {
			return true
		}
	}
	return false
}

func siegeTechMod(gs *state.GameState, ownerID string) float64 {
	if gs == nil || ownerID == "" {
		return 0
	}
	f := gs.Factions[faction.FactionID(ownerID)]
	if f == nil || gs.TechTypes == nil {
		return 0
	}
	return tech.ComputeEffects(f.Research.Completed, gs.TechTypes).SiegeAttackMod
}

func canArmyStartSiege(gs *state.GameState, attacker *army.Army, targetRegion *world.Region) (bool, string) {
	if gs == nil || attacker == nil || targetRegion == nil {
		return false, ""
	}
	if attacker.IsNaval {
		return false, "Donanma kuşatma başlatamaz."
	}
	if !targetRegion.CanLandEnter() {
		return false, "Bu bölgeye kuşatma kurulamaz."
	}
	if targetRegion.OwnerID == "" || targetRegion.OwnerID == attacker.OwnerID {
		return false, "Kuşatma için düşman tahkimatı gerekli."
	}
	if !targetRegion.IsFortified() {
		return false, "Bu bölge tahkimli değil."
	}
	if !gameFactionsAtWar(gs, attacker.OwnerID, targetRegion.OwnerID) {
		return false, "Kuşatma için önce savaş halinde olmalısın."
	}
	if attacker.RegionID != targetRegion.ID && !regionsAdjacent(gs, attacker.RegionID, targetRegion.ID) {
		return false, "Kuşatma için hedef bölgeye komşu olmalısın."
	}
	if active := gs.SiegeByArmy(attacker.ID); active != nil && active.RegionID != targetRegion.ID {
		return false, "Bu ordu başka bir kuşatma yürütüyor. Önce onu kaldır."
	}
	if siege := gs.SiegeAt(targetRegion.ID); siege != nil && siege.AttackerArmyID != attacker.ID {
		if gs.CanJoinActiveSiege(attacker, targetRegion.ID) {
			return false, "Bu kuşatmaya destek için normal hareket kullan."
		}
		siegeArmy := gs.Armies[siege.AttackerArmyID]
		if siegeArmy == nil || siegeArmy.OwnerID != attacker.OwnerID {
			return false, "Bu bölge zaten başka bir ordu tarafından kuşatılıyor."
		}
	}
	return true, ""
}

func canArmyAssaultSiege(gs *state.GameState, attacker *army.Army, targetRegion *world.Region) (*state.SiegeState, bool, string) {
	if gs == nil || attacker == nil || targetRegion == nil {
		return nil, false, ""
	}
	if ok, reason := canArmyStartSiege(gs, attacker, targetRegion); ok {
		return gs.SiegeAt(targetRegion.ID), true, ""
	} else if siege := gs.SiegeAt(targetRegion.ID); siege != nil && siege.AttackerArmyID == attacker.ID {
		if attacker.RegionID != targetRegion.ID && !regionsAdjacent(gs, attacker.RegionID, targetRegion.ID) {
			return nil, false, "Genel hücum için kuşatma hattında kalmalısın."
		}
		if !gameFactionsAtWar(gs, attacker.OwnerID, targetRegion.OwnerID) {
			return nil, false, "Kuşatma artık geçersiz; savaş hali sona ermiş."
		}
		return siege, true, ""
	} else {
		return nil, false, reason
	}
}

func siegeBreachThresholds(fortLevel int) (int, int) {
	if fortLevel < 1 {
		fortLevel = 1
	}
	minor := 8 + fortLevel*2
	major := minor * 2
	return minor, major
}

func siegeBreachLevel(progress, fortLevel int) int {
	minorThreshold, majorThreshold := siegeBreachThresholds(fortLevel)
	switch {
	case progress >= majorThreshold:
		return 2
	case progress >= minorThreshold:
		return 1
	default:
		return 0
	}
}

func siegeDefenseBonus(fortLevel, breachLevel int) float64 {
	if fortLevel <= 0 {
		return 0
	}
	base := float64(fortLevel) * 0.14
	switch breachLevel {
	case 2:
		return base * 0.25
	case 1:
		return base * 0.55
	default:
		return base + 0.18
	}
}

func siegeAssaultCanCapture(breachLevel int) bool {
	return breachLevel > 0
}

func siegeAssaultAttackerDamage(fortLevel, breachLevel int) int {
	if fortLevel < 1 {
		fortLevel = 1
	}
	damage := fortLevel
	switch breachLevel {
	case 2:
		damage = fortLevel
	case 1:
		damage = fortLevel*2 + 1
	default:
		damage = fortLevel*2 + 6
	}
	if damage < 1 {
		damage = 1
	}
	if damage > 14 {
		damage = 14
	}
	return damage
}

// siegeForce, aktif kuşatmanın o turdaki toplam kuşatma katkısını taşır.
// Kuşatma başlangıcından sonra bölgeye gelen aynı realm/müttefik ordular da
// bu kuvvete dahil edilir; SiegeState yalnızca ana kuşatan ArmyID'sini
// serialize ettiği için destek orduları her tick'te canlı state'ten taranır.
type siegeForce struct {
	UnitScore     int
	HighestTier   int
	BreachPower   float64
	ProgressBonus int
	BreachBonus   int
}

// siegeForceForArmy tek bir ordunun kuşatma katkısını hesaplar. Gedik gücü,
// yalnızca ilgili sur seviyesine erişebilen kuşatma birimlerinden ve canlı HP
// oranlarından oluşur.
func siegeForceForArmy(gs *state.GameState, attacker *army.Army, fortLevel int) siegeForce {
	force := siegeForce{}
	if gs == nil || attacker == nil || attacker.IsNaval || len(attacker.Units) == 0 {
		return force
	}
	force.UnitScore = attacker.SiegeUnitScore(gs.UnitTypes)
	force.HighestTier = attacker.HighestSiegeTier(gs.UnitTypes)
	for _, unit := range attacker.Units {
		unitType := gs.UnitTypes[unit.TypeID]
		if unitType == nil || unitType.Category != army.CategorySiege ||
			fortLevel > int(unitType.Tier)+2 {
			continue
		}
		hp := unit.CurrentHP
		if hp < 0 {
			hp = 0
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		force.BreachPower += float64(2+int(unitType.Tier)) * float64(hp) / float64(army.MaxUnitHP)
	}
	force.ProgressBonus, force.BreachBonus = attacker.CommanderSiegeBonuses()
	return force
}

// activeSiegeForce, ana kuşatanın yanında aktif kuşatma bölgesine gelmiş
// destek ordularını da toplar. CanJoinActiveSiege ile aynı diplomasi kuralı
// kullanıldığı için ilgisiz üçüncü devletlerin birlikleri kuşatma gücüne
// sızamaz.
func activeSiegeForce(gs *state.GameState, siege *state.SiegeState, attacker *army.Army) siegeForce {
	fortLevel := 1
	if gs != nil && siege != nil {
		if target := gs.Regions[siege.RegionID]; target != nil {
			fortLevel = target.FortificationLevel()
		}
	}
	force := siegeForceForArmy(gs, attacker, fortLevel)
	if gs == nil || siege == nil || attacker == nil {
		return force
	}
	for candidateID, candidate := range gs.Armies {
		if candidateID == attacker.ID || candidate == nil || candidate.IsNaval ||
			candidate.RegionID != siege.RegionID || len(candidate.Units) == 0 {
			continue
		}
		if !gs.CanJoinActiveSiege(candidate, siege.RegionID) {
			continue
		}
		candidateForce := siegeForceForArmy(gs, candidate, fortLevel)
		force.UnitScore += candidateForce.UnitScore
		force.BreachPower += candidateForce.BreachPower
		if candidateForce.HighestTier > force.HighestTier {
			force.HighestTier = candidateForce.HighestTier
		}
		force.ProgressBonus += candidateForce.ProgressBonus
		force.BreachBonus += candidateForce.BreachBonus
	}
	return force
}

func siegeProgressGain(gs *state.GameState, attacker *army.Army, targetRegion *world.Region, defender *army.Army) int {
	if gs == nil || attacker == nil || targetRegion == nil {
		return 0
	}
	return siegeProgressGainForForce(gs, attacker, targetRegion, defender, siegeForceForArmy(gs, attacker, targetRegion.FortificationLevel()))
}

func siegeProgressGainForForce(gs *state.GameState, attacker *army.Army, targetRegion *world.Region, defender *army.Army, force siegeForce) int {
	if gs == nil || attacker == nil || targetRegion == nil {
		return 0
	}
	progress := force.UnitScore + 1 + int(siegeTechMod(gs, attacker.OwnerID)*10+0.5) + force.ProgressBonus
	progress -= targetRegion.FortificationLevel()
	if defender != nil {
		progress -= defender.TotalDefense(gs.UnitTypes) / 90
	}
	if progress < 1 {
		progress = 1
	}
	return progress
}

func siegeBreachGain(gs *state.GameState, attacker *army.Army, targetRegion *world.Region, defender *army.Army) float64 {
	if gs == nil || attacker == nil || targetRegion == nil {
		return 0
	}
	return siegeBreachGainForForce(gs, attacker, targetRegion, defender, siegeForceForArmy(gs, attacker, targetRegion.FortificationLevel()))
}

func siegeBreachGainForForce(gs *state.GameState, attacker *army.Army, targetRegion *world.Region, defender *army.Army, force siegeForce) float64 {
	if gs == nil || attacker == nil || targetRegion == nil || force.BreachPower <= 0 {
		return 0
	}
	fortLevel := targetRegion.FortificationLevel()
	if fortLevel < 1 {
		fortLevel = 1
	}
	gain := force.BreachPower + float64(int(siegeTechMod(gs, attacker.OwnerID)*8+0.5)+force.BreachBonus-fortLevel/2)
	if defender != nil {
		gain -= float64(defender.TotalDefense(gs.UnitTypes) / 120)
	}
	if gain < 0 {
		gain = 0
	}
	return gain
}

func siegeSurrenderTurns(fortLevel int) int {
	return state.SiegeSurrenderTurns(fortLevel)
}

func siegeAttritionDamage(progressGain, breachLevel, fortLevel int) int {
	damage := 2 + progressGain/2 + breachLevel*2 - fortLevel/2
	if damage < 2 {
		damage = 2
	}
	if damage > 12 {
		damage = 12
	}
	return damage
}

// siegeEmptyGarrisonAttritionDamage represents the small cost of maintaining
// a siege when the settlement has no field army left. It is deliberately much
// lower than defender attrition: there is no opposing army causing casualties,
// but disease, exhaustion, and siege logistics still wear the besieger down.
// Kuşatanın hedef bölgeyle ortak kara sınırı varsa düzenli ikmal bu maliyeti
// azaltır.
func siegeEmptyGarrisonAttritionDamage(supplied bool) int {
	if supplied {
		return 2
	}
	return 3
}

const (
	granaryAttritionReductionPerLevel = 10
	granaryAttritionReductionMax      = 30
)

// granaryAttritionReductionPercent bölgedeki ambar seviyesinin kuşatılan
// savunma ordusuna sağladığı yerel yıpranma azaltmasını döner. Ambar seviyesi
// Region.BuildingLevel ile tutulduğu için ekonomi ve kuşatma aynı bina state'ini
// kullanır; kuşatan ordu bu savunma bonusundan yararlanmaz.
func granaryAttritionReductionPercent(region *world.Region) int {
	if region == nil {
		return 0
	}
	level := region.BuildingLevel("granary")
	if level <= 0 {
		return 0
	}
	percent := level * granaryAttritionReductionPerLevel
	if percent > granaryAttritionReductionMax {
		return granaryAttritionReductionMax
	}
	return percent
}

func reduceAttritionDamageForGranary(damage, reductionPercent int) int {
	if damage <= 0 || reductionPercent <= 0 {
		return damage
	}
	if reductionPercent > 100 {
		reductionPercent = 100
	}
	damage = damage * (100 - reductionPercent) / 100
	if damage < 1 {
		damage = 1
	}
	return damage
}

func applyArmyFlatDamage(a *army.Army, damage int) (lostUnits int, totalHPDamage int) {
	if a == nil || damage <= 0 || len(a.Units) == 0 {
		return 0, 0
	}
	survivors := a.Units[:0]
	for _, unit := range a.Units {
		before := unit.CurrentHP
		unit.CurrentHP -= damage
		if unit.CurrentHP <= 0 {
			lostUnits++
			totalHPDamage += before
			continue
		}
		totalHPDamage += before - unit.CurrentHP
		survivors = append(survivors, unit)
	}
	a.Units = survivors
	return lostUnits, totalHPDamage
}

func (g *Game) virtualSiegeGarrison(targetRegion *world.Region) *army.Army {
	if g == nil || g.gs == nil || targetRegion == nil {
		return nil
	}
	unitTypeID := "militia"
	if _, ok := g.gs.UnitTypes[unitTypeID]; !ok {
		for id, ut := range g.gs.UnitTypes {
			if ut != nil && ut.Category == army.CategoryInfantry {
				unitTypeID = id
				break
			}
		}
	}
	fortLevel := targetRegion.FortificationLevel()
	unitCount := 1 + fortLevel
	if unitCount > 6 {
		unitCount = 6
	}
	units := make([]army.Unit, 0, unitCount)
	for i := 0; i < unitCount; i++ {
		units = append(units, army.Unit{TypeID: unitTypeID, CurrentHP: army.MaxUnitHP})
	}
	return &army.Army{
		OwnerID:    targetRegion.OwnerID,
		RegionID:   targetRegion.ID,
		Units:      units,
		MovePoints: 0,
	}
}

// siegeDefenderForTurn returns the living defensive army currently inside the
// besieged region. DefenderArmyID is only a save-friendly hint from siege
// creation; it can become stale when the army is destroyed, split, or replaced
// by a relieving force during a later turn.
func siegeDefenderForTurn(gs *state.GameState, siege *state.SiegeState, attacker *army.Army, regionID world.RegionID) *army.Army {
	if gs == nil || siege == nil || attacker == nil || regionID == "" {
		return nil
	}
	if defender := gs.Armies[siege.DefenderArmyID]; defender != nil &&
		defender.RegionID == regionID && len(defender.Units) > 0 &&
		gs.IsArmyDefendingSiegedRegion(defender) {
		return defender
	}

	// Re-select every tick so that a newly arrived owner/allied army receives
	// the same gradual siege attrition as the original garrison.
	return gs.SelectBattleDefender(attacker, regionID, false)
}

func (g *Game) clearSiege(regionID world.RegionID) {
	if g == nil || g.gs == nil || g.gs.Sieges == nil || regionID == "" {
		return
	}
	if _, exists := g.gs.Sieges[regionID]; exists && g.renderer != nil {
		g.renderer.MarkMapDirty()
	}
	delete(g.gs.Sieges, regionID)
}

// transferSiegeToRemainingArmy kuşatan ordu ayrılırken aynı fraksiyona ait
// başka bir kara ordusu hedef bölgede kaldıysa kuşatma kaydını ona devreder.
// Böylece kuşatma, kuşatan gücün son birimi bölgeden ayrılmadığı sürece
// yalnızca eski ArmyID değişti diye düşmez.
func (g *Game) transferSiegeToRemainingArmy(regionID world.RegionID, leavingArmyID army.ArmyID) bool {
	if g == nil || g.gs == nil || regionID == "" || leavingArmyID == "" {
		return false
	}
	siege := g.gs.SiegeAt(regionID)
	if siege == nil || siege.AttackerArmyID != leavingArmyID {
		return false
	}
	leavingArmy := g.gs.Armies[leavingArmyID]
	attackerFactionID := siege.AttackerFactionID
	if attackerFactionID == "" && leavingArmy != nil {
		attackerFactionID = leavingArmy.OwnerID
	}
	if attackerFactionID == "" {
		return false
	}
	for candidateID, candidate := range g.gs.Armies {
		if candidateID == leavingArmyID || candidate == nil || candidate.IsNaval ||
			candidate.OwnerID != attackerFactionID || candidate.RegionID != regionID ||
			len(candidate.Units) == 0 {
			continue
		}
		siege.AttackerArmyID = candidateID
		return true
	}
	return false
}

// releaseSiegeForArmyMovement kuşatmayı terk eden ordunun yerine aynı
// fraksiyondan kalan bir ordu varsa kaydı ona devreder; yoksa kuşatmayı
// kaldırır. Çağıran taraf bunu yalnız hareket doğrulandıktan sonra yapmalıdır.
func (g *Game) releaseSiegeForArmyMovement(regionID world.RegionID, leavingArmyID army.ArmyID) {
	if g.transferSiegeToRemainingArmy(regionID, leavingArmyID) {
		return
	}
	g.clearSiege(regionID)
}

func (g *Game) clearSiegesByArmy(armyID army.ArmyID) {
	if g == nil || g.gs == nil || g.gs.Sieges == nil || armyID == "" {
		return
	}
	for rid, siege := range g.gs.Sieges {
		if siege != nil && siege.AttackerArmyID == armyID {
			if g.renderer != nil {
				g.renderer.MarkMapDirty()
			}
			// Orduyu kuşatma öncesi bulunduğu bölgeye geri taşı
			if a := g.gs.Armies[armyID]; a != nil && siege.AttackerHomeRegionID != "" {
				a.RegionID = siege.AttackerHomeRegionID
			}
			delete(g.gs.Sieges, rid)
		}
	}
}

func (g *Game) startSiege(aid army.ArmyID, target world.RegionID) {
	g.startSiegeForArmy(aid, target, true)
}

func (g *Game) startSiegeForArmy(aid army.ArmyID, target world.RegionID, notify bool) bool {
	if g == nil || g.gs == nil {
		return false
	}
	attacker := g.gs.Armies[aid]
	targetRegion := g.gs.Regions[target]
	ok, reason := canArmyStartSiege(g.gs, attacker, targetRegion)
	if !ok {
		if notify && reason != "" && g.renderer != nil {
			g.renderer.ShowCombatResult(reason)
		}
		return false
	}
	ensureSiegeMap(g.gs)
	defender := g.gs.SelectBattleDefender(attacker, target, false)
	fortLevel := targetRegion.FortificationLevel()
	homeRegion := attacker.RegionID
	g.gs.Sieges[target] = &state.SiegeState{
		RegionID:             target,
		AttackerArmyID:       aid,
		AttackerHomeRegionID: homeRegion,
		AttackerFactionID:    attacker.OwnerID,
		StartedTurn:          g.gs.Turn,
		FortLevel:            fortLevel,
		GranaryLevel:         targetRegion.BuildingLevel("granary"),
	}
	if defender != nil {
		g.gs.Sieges[target].DefenderArmyID = defender.ID
	}
	attacker.RegionID = target
	attacker.MovePoints = 0
	if notify && g.renderer != nil {
		msg := fmt.Sprintf("%s kuşatıldı. Tahkimat seviyesi: %d.", targetRegion.NameTR, fortLevel)
		g.renderer.MarkMapDirty()
		g.renderer.ShowCombatResult(msg)
		g.renderer.AddEvent("[KUSATMA] " + msg)
	}
	return true
}

func (g *Game) liftSiege(aid army.ArmyID, target world.RegionID) {
	if g == nil || g.gs == nil {
		return
	}
	siege := g.gs.SiegeAt(target)
	if siege == nil || siege.AttackerArmyID != aid {
		if g.renderer != nil {
			g.renderer.ShowCombatResult("Bu ordu aktif bir kuşatma yürütmüyor.")
		}
		return
	}
	// Orduyu kuşatma öncesi bulunduğu bölgeye geri taşı
	if a := g.gs.Armies[aid]; a != nil && siege.AttackerHomeRegionID != "" {
		a.RegionID = siege.AttackerHomeRegionID
	}
	g.clearSiege(target)
	if region := g.gs.Regions[target]; region != nil && g.renderer != nil {
		msg := region.NameTR + " kuşatması kaldırıldı."
		g.renderer.MarkMapDirty()
		g.renderer.ShowCombatResult(msg)
		g.renderer.AddEvent("[KUSATMA] " + msg)
	}
}

func (g *Game) captureBesiegedRegion(attacker *army.Army, targetRegion *world.Region, showAfterBattleReport bool) (eliminationResult, bool) {
	if g == nil || attacker == nil || targetRegion == nil {
		return eliminationResult{}, false
	}
	attacker.RegionID = targetRegion.ID
	attacker.DockedRegionID = ""
	attacker.DockedSettlementID = ""
	attacker.MovePoints = 0
	g.clearSiege(targetRegion.ID)
	prompted := g.queueConquestDecision(faction.FactionID(attacker.OwnerID), targetRegion, showAfterBattleReport)
	collapse := eliminationResult{}
	if !prompted {
		vassalResult := ai.TryResolvePostWarVassalization(g.gs, faction.FactionID(attacker.OwnerID), targetRegion)
		if !vassalResult.Applied {
			collapse = g.applyConquestWithNavalEviction(targetRegion, attacker.OwnerID)
		}
	}
	return collapse, prompted
}

func (g *Game) assaultSiege(aid army.ArmyID, target world.RegionID) {
	g.assaultSiegeWithStance(aid, target, combat.BattleStanceBalanced)
}

func (g *Game) assaultSiegeWithStance(aid army.ArmyID, target world.RegionID, stance combat.BattleStance) bool {
	if g == nil || g.gs == nil {
		return false
	}
	attacker := g.gs.Armies[aid]
	targetRegion := g.gs.Regions[target]
	siege, ok, reason := canArmyAssaultSiege(g.gs, attacker, targetRegion)
	if !ok {
		if reason != "" && g.renderer != nil {
			g.renderer.ShowCombatResult(reason)
		}
		return false
	}
	stance = combat.NormalizeBattleStance(stance)
	fortLevel := targetRegion.FortificationLevel()
	breachLevel := 0
	if siege != nil {
		fortLevel = siege.FortLevel
		breachLevel = siege.BreachLevel
	}

	defender := g.gs.SelectBattleDefender(attacker, target, false)
	virtualDefense := false
	if defender == nil {
		defender = g.virtualSiegeGarrison(targetRegion)
		virtualDefense = true
	}
	attackerBefore := snapshotBattleArmy(attacker, g.gs.UnitTypes)
	defenderBefore := snapshotBattleArmy(defender, g.gs.UnitTypes)
	atkMods := techModsFor(g.gs, attacker.OwnerID)
	defMods := combat.TechMods{}
	if virtualDefense {
		defMods = techModsFor(g.gs, targetRegion.OwnerID)
	} else {
		defMods = techModsFor(g.gs, defender.OwnerID)
	}
	defenderLabel := "Savunma Hattı"
	if virtualDefense {
		defenderLabel = "Garnizon"
	}
	defMods.DefenseMod += siegeDefenseBonus(fortLevel, breachLevel)
	result := combat.ResolveBattleWithContextPlan(attacker, defender, targetRegion.Terrain, g.gs.UnitTypes, atkMods, defMods, combat.BattleContextLand, stance)
	g.recordCommanderBattle(attacker, defender, nil, result.AttackerWins)
	var collapse eliminationResult
	extraDamage := siegeAssaultAttackerDamage(fortLevel, breachLevel)
	if extraDamage > 0 {
		extraLost, _ := applyArmyFlatDamage(attacker, extraDamage)
		result.AttackerLost += extraLost
	}
	defenderOwnerID := targetRegion.OwnerID
	if !virtualDefense && defender != nil {
		defenderOwnerID = defender.OwnerID
	}
	g.gs.RecordWarCasualties(faction.FactionID(attacker.OwnerID), faction.FactionID(defenderOwnerID), result.AttackerLost, result.DefenderLost)

	if result.AttackerWins {
		prompted := false
		if !virtualDefense && len(defender.Units) == 0 {
			g.gs.RemoveArmy(defender.ID)
		}
		if !siegeAssaultCanCapture(breachLevel) {
			if len(attacker.Units) == 0 {
				g.gs.RemoveArmy(aid)
				g.clearSiegesByArmy(aid)
			}
			g.presentBattleReport(g.makeBattleReport(
				render.BattleSceneSiege,
				targetRegion.NameTR,
				stance,
				result.Description,
				"Surlar zorlandı fakat gedik olmadan içeri girilemedi.",
				"Hücum Gücü",
				defenderLabel,
				g.factionNameTR(attacker.OwnerID),
				g.factionNameTR(targetRegion.OwnerID),
				attackerBefore,
				attacker,
				defenderBefore,
				defender,
			))
			return false
		}
		if len(attacker.Units) > 0 {
			collapse, prompted = g.captureBesiegedRegion(attacker, targetRegion, true)
			if g.renderer != nil {
				if !prompted {
					g.renderer.MarkMapDirty()
				}
			}
		} else {
			g.gs.RemoveArmy(aid)
			g.clearSiege(target)
		}
		outcomeDetail := "Tahkimat düştü ve bölge ele geçirildi."
		if prompted {
			outcomeDetail = "Tahkimat düştü; ilhak ya da vassallık için savaş sonrası karar bekleniyor."
		}
		if len(attacker.Units) == 0 {
			outcomeDetail = "Tahkimat yarıldı fakat hücum gücü tükendi; bölge alınamadı."
		}
		g.presentBattleReport(g.makeBattleReport(
			render.BattleSceneSiege,
			targetRegion.NameTR,
			stance,
			result.Description,
			outcomeDetail,
			"Hücum Gücü",
			defenderLabel,
			g.factionNameTR(attacker.OwnerID),
			g.factionNameTR(targetRegion.OwnerID),
			attackerBefore,
			attacker,
			defenderBefore,
			defender,
		))
		g.announceElimination(collapse)
		return true
	}

	if len(attacker.Units) == 0 {
		g.gs.RemoveArmy(aid)
		g.clearSiegesByArmy(aid)
	}
	g.presentBattleReport(g.makeBattleReport(
		render.BattleSceneSiege,
		targetRegion.NameTR,
		stance,
		result.Description,
		"Genel hücum püskürtüldü; tahkimat elde kaldı.",
		"Hücum Gücü",
		defenderLabel,
		g.factionNameTR(attacker.OwnerID),
		g.factionNameTR(targetRegion.OwnerID),
		attackerBefore,
		attacker,
		defenderBefore,
		defender,
	))
	return false
}

func (g *Game) resolveSieges() []siegeTurnUpdate {
	if g == nil || g.gs == nil || len(g.gs.Sieges) == 0 {
		return nil
	}
	updates := make([]siegeTurnUpdate, 0, len(g.gs.Sieges))
	for regionID, siege := range g.gs.Sieges {
		if siege == nil {
			delete(g.gs.Sieges, regionID)
			continue
		}
		targetRegion := g.gs.Regions[regionID]
		attacker := g.gs.Armies[siege.AttackerArmyID]
		g.evacuateNonBelligerentArmiesFromSiege(regionID, siege, attacker)
		if targetRegion == nil || attacker == nil || targetRegion.OwnerID == "" || targetRegion.OwnerID == attacker.OwnerID || (attacker.RegionID != regionID && !regionsAdjacent(g.gs, attacker.RegionID, regionID)) || !gameFactionsAtWar(g.gs, attacker.OwnerID, targetRegion.OwnerID) {
			// Kuşatma geçersiz → orduyu homeRegion'a geri taşı
			if attacker != nil && siege.AttackerHomeRegionID != "" {
				attacker.RegionID = siege.AttackerHomeRegionID
			}
			delete(g.gs.Sieges, regionID)
			continue
		}

		siege.TurnsElapsed++
		siege.FortLevel = targetRegion.FortificationLevel()
		siege.GranaryLevel = targetRegion.BuildingLevel("granary")
		defender := siegeDefenderForTurn(g.gs, siege, attacker, regionID)
		if defender != nil {
			siege.DefenderArmyID = defender.ID
		}
		force := activeSiegeForce(g.gs, siege, attacker)
		progressGain := siegeProgressGainForForce(g.gs, attacker, targetRegion, defender, force)
		breachGain := siegeBreachGainForForce(g.gs, attacker, targetRegion, defender, force)
		oldBreach := siege.BreachLevel
		totalBreachProgress := siege.BreachProgressRemainder + breachGain
		wholeBreachProgress := int(totalBreachProgress)
		siege.BreachProgress += wholeBreachProgress
		siege.BreachProgressRemainder = totalBreachProgress - float64(wholeBreachProgress)
		siege.BreachLevel = siegeBreachLevel(siege.BreachProgress, siege.FortLevel)

		if defender != nil {
			damage := siegeAttritionDamage(progressGain, siege.BreachLevel, siege.FortLevel)
			damage = reduceAttritionDamageForGranary(damage, granaryAttritionReductionPercent(targetRegion))
			lostUnits, totalHPDamage := applyArmyFlatDamage(defender, damage)
			g.gs.RecordWarAttritionCasualties(faction.FactionID(attacker.OwnerID), faction.FactionID(targetRegion.OwnerID), 0, lostUnits)
			if len(defender.Units) == 0 {
				g.gs.RemoveArmy(defender.ID)
				defender = nil
				siege.DefenderArmyID = ""
			}
			if totalHPDamage > 0 {
				updates = append(updates, siegeTurnUpdate{
					Message: fmt.Sprintf("%s kuşatması savunanlara baskı uyguluyor.", targetRegion.NameTR),
					Detail:  fmt.Sprintf("%s kuşatmasında savunan ordu %d HP ve %d birim kaybetti.", targetRegion.NameTR, totalHPDamage, lostUnits),
					Popup:   attacker.OwnerID == string(g.gs.PlayerFactionID) || targetRegion.OwnerID == string(g.gs.PlayerFactionID),
				})
			}
		} else {
			// A vacant settlement still imposes a small, gradual cost on the
			// besieger. This is not fortress combat and must not use the full
			// defender attrition formula.
			damage := siegeEmptyGarrisonAttritionDamage(
				g.gs.HasOwnedLandSupplyBorder(targetRegion.ID, attacker.OwnerID),
			)
			lostUnits, totalHPDamage := applyArmyFlatDamage(attacker, damage)
			g.gs.RecordWarAttritionCasualties(
				faction.FactionID(attacker.OwnerID),
				faction.FactionID(targetRegion.OwnerID),
				lostUnits,
				0,
			)
			if totalHPDamage > 0 {
				updates = append(updates, siegeTurnUpdate{
					Message: fmt.Sprintf("%s kuşatması kuşatan orduyu yıpratıyor.", targetRegion.NameTR),
					Detail:  fmt.Sprintf("%s kuşatmasında savunma ordusu bulunmadığı halde kuşatan ordu %d HP ve %d birim kaybetti.", targetRegion.NameTR, totalHPDamage, lostUnits),
					Popup:   attacker.OwnerID == string(g.gs.PlayerFactionID),
				})
			}
		}

		if siege.BreachLevel > oldBreach {
			breachLabel := "küçük gedik"
			if siege.BreachLevel >= 2 {
				breachLabel = "büyük gedik"
			}
			updates = append(updates, siegeTurnUpdate{
				Message: fmt.Sprintf("%s surlarında %s açıldı.", targetRegion.NameTR, breachLabel),
				Detail:  fmt.Sprintf("%s kuşatması %d turdur sürüyor. İlerleme: %d, tahkimat: %d.", targetRegion.NameTR, siege.TurnsElapsed, siege.BreachProgress, siege.FortLevel),
				Popup:   attacker.OwnerID == string(g.gs.PlayerFactionID) || targetRegion.OwnerID == string(g.gs.PlayerFactionID),
			})
		}

		if defender == nil && (siege.BreachLevel >= 2 || siege.TurnsElapsed >= siege.TotalSurrenderTurns()) {
			collapse, prompted := g.captureBesiegedRegion(attacker, targetRegion, false)
			if g.renderer != nil {
				if !prompted {
					g.renderer.MarkMapDirty()
				}
			}
			delete(g.gs.Sieges, regionID)
			msg := fmt.Sprintf("%s kuşatma sonrası teslim oldu.", targetRegion.NameTR)
			if siege.BreachLevel < 2 {
				msg = fmt.Sprintf("%s uzun kuşatma sonrası açlıkla teslim oldu.", targetRegion.NameTR)
			}
			detail := fmt.Sprintf("%s kuşatması %d tur sonunda teslimiyetle sonuçlandı.", targetRegion.NameTR, siege.TurnsElapsed)
			if siege.BreachLevel < 2 {
				detail = fmt.Sprintf("%s kuşatması %d tur sürdü; surlarda gedik açılamasa da açlık teslimiyet getirdi.", targetRegion.NameTR, siege.TurnsElapsed)
			}
			if prompted {
				detail += " Nihai düzen için ilhak veya vassallık kararı bekleniyor."
			}
			updates = append(updates, siegeTurnUpdate{
				Message: msg,
				Detail:  detail,
				Popup:   attacker.OwnerID == string(g.gs.PlayerFactionID) || targetRegion.OwnerID == string(g.gs.PlayerFactionID),
			})
			g.announceElimination(collapse)
		}
	}
	return updates
}

// evacuateNonBelligerentArmiesFromSiege, kuşatanla savaşta olmayan üçüncü
// devlet ordularının kuşatma hattında çatışmaya zorlanmasını önler. Bu ordular
// kuşatma bölgesine son hareketlerinde geldikleri komşu bölgeye döner;
// önceki konum bilinmiyorsa veya artık geçerli değilse sessizce bırakılır.
func (g *Game) evacuateNonBelligerentArmiesFromSiege(regionID world.RegionID, siege *state.SiegeState, besieger *army.Army) int {
	if g == nil || g.gs == nil || siege == nil || besieger == nil || regionID == "" {
		return 0
	}
	count := 0
	for _, candidate := range g.gs.Armies {
		if candidate == nil || candidate.ID == besieger.ID || candidate.IsNaval ||
			candidate.RegionID != regionID || candidate.PreviousRegionID == "" ||
			candidate.PreviousRegionID == regionID {
			continue
		}
		if gameFactionsAtWar(g.gs, candidate.OwnerID, besieger.OwnerID) {
			continue
		}
		// Kuşatanın aynı realm/müttefik destekçileri savaşta görünmese de
		// kuşatma gücünün parçasıdır; yalnızca karşı tarafın müttefikleri veya
		// ilgisiz devletler geri gönderilir.
		if g.gs.CanJoinActiveSiege(candidate, regionID) {
			continue
		}
		if !regionsAdjacent(g.gs, candidate.PreviousRegionID, regionID) {
			continue
		}
		previous := candidate.RegionID
		candidate.RegionID = candidate.PreviousRegionID
		candidate.PreviousRegionID = previous
		candidate.DockedRegionID = ""
		candidate.DockedSettlementID = ""
		candidate.InAmbush = false
		count++
		if g.renderer != nil {
			g.renderer.MarkMapDirty()
		}
	}
	return count
}
