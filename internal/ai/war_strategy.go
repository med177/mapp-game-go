package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiWarSupportNearDistance = 1
	aiWarSupportMidDistance  = 3
	aiWarSupportFarDistance  = 5
	aiWarSupportMaxDistance  = 8
)

type aiWarCoalitionRisk struct {
	TargetPower                    int
	TargetVassalPower              int
	AllyPower                      int
	AllyVassalPower                int
	AllySupportPower               int
	DefenderPower                  int
	AttackerPower                  int
	AttackerVassalPower            int
	CertainAttackerAllyPower       int
	CertainAttackerAllyVassalPower int
	CertainAttackerSupportPower    int
	NearestAllyArmy                int
	NearestAttackerAllyArmy        int
}

// aiWarCoalitionAssessment, savaş ilanı öncesinde iki tarafın etkin koalisyon
// gücünü ölçer. Savunan tarafta hedefin dış müttefikleri, saldıran tarafta ise
// AssessWarCall tarafından AutoJoin olarak işaretlenen kesin müttefikler
// kullanılır. Müttefik katkısı, ordularının hedef ülke topraklarına olan rota
// mesafesiyle ağırlıklandırılır; böylece haritanın diğer ucundaki kuvvet tam
// cephe gücü gibi değerlendirilmez.
func aiWarCoalitionAssessment(gs *state.GameState, actor, target faction.FactionID) aiWarCoalitionRisk {
	assessment := aiWarCoalitionRisk{NearestAllyArmy: -1, NearestAttackerAllyArmy: -1}
	if gs == nil || target == "" {
		return assessment
	}

	targetRoot := diplomacy.RealmRoot(gs, target)
	if targetRoot == "" {
		targetRoot = target
	}
	actorRoot := diplomacy.RealmRoot(gs, actor)
	if actorRoot == "" {
		actorRoot = actor
	}
	battlefield := aiWarBattlefieldRegions(gs, targetRoot)
	assessment.AttackerPower = diplomacy.MilitaryPower(gs, actorRoot)
	for _, vassalID := range diplomacy.VassalsOf(gs, actorRoot) {
		assessment.AttackerVassalPower += aiWarWeightedFactionPower(gs, vassalID, battlefield)
	}
	assessment.AttackerPower += assessment.AttackerVassalPower

	for _, allyRoot := range aiWarExternalAllies(gs, actorRoot, targetRoot) {
		if aiWarAllyJoinIsPending(gs, allyRoot) {
			continue
		}
		call := diplomacy.AssessWarCall(gs, actorRoot, allyRoot, targetRoot)
		if !call.AutoJoin {
			continue
		}
		allyPower, nearest := aiWarWeightedFactionPowerWithDistance(gs, allyRoot, battlefield)
		vassalPower := 0
		for _, vassalID := range diplomacy.VassalsOf(gs, allyRoot) {
			vassalPower += aiWarWeightedFactionPower(gs, vassalID, battlefield)
		}
		assessment.CertainAttackerAllyPower += allyPower
		assessment.CertainAttackerAllyVassalPower += vassalPower
		assessment.CertainAttackerSupportPower += allyPower + vassalPower
		assessment.AttackerPower += allyPower + vassalPower
		if nearest >= 0 && (assessment.NearestAttackerAllyArmy < 0 || nearest < assessment.NearestAttackerAllyArmy) {
			assessment.NearestAttackerAllyArmy = nearest
		}
	}

	assessment.TargetPower = diplomacy.MilitaryPower(gs, targetRoot)
	for _, vassalID := range diplomacy.VassalsOf(gs, targetRoot) {
		assessment.TargetVassalPower += aiWarWeightedFactionPower(gs, vassalID, battlefield)
	}
	assessment.DefenderPower = assessment.TargetPower + assessment.TargetVassalPower

	for _, allyRoot := range aiWarExternalAllies(gs, targetRoot, actorRoot) {
		allyPower, nearest := aiWarWeightedFactionPowerWithDistance(gs, allyRoot, battlefield)
		vassalPower := 0
		for _, vassalID := range diplomacy.VassalsOf(gs, allyRoot) {
			vassalPower += aiWarWeightedFactionPower(gs, vassalID, battlefield)
		}
		assessment.AllyPower += allyPower
		assessment.AllyVassalPower += vassalPower
		assessment.AllySupportPower += allyPower + vassalPower
		assessment.DefenderPower += allyPower + vassalPower
		if nearest >= 0 && (assessment.NearestAllyArmy < 0 || nearest < assessment.NearestAllyArmy) {
			assessment.NearestAllyArmy = nearest
		}
	}
	return assessment
}

func aiWarAllyJoinIsPending(gs *state.GameState, ally faction.FactionID) bool {
	if gs == nil || gs.PlayerFactionID == "" || ally == "" {
		return false
	}
	return diplomacy.SameRealm(gs, ally, gs.PlayerFactionID)
}

func aiWarExternalAllies(gs *state.GameState, target, actor faction.FactionID) []faction.FactionID {
	if gs == nil || target == "" {
		return nil
	}
	allies := make([]faction.FactionID, 0, 4)
	seen := make(map[faction.FactionID]struct{})
	for _, otherID := range aiSortedFactionIDs(gs) {
		if otherID == target || otherID == actor || diplomacy.SameRealm(gs, target, otherID) {
			continue
		}
		allyRoot := diplomacy.RealmRoot(gs, otherID)
		if allyRoot == "" {
			allyRoot = otherID
		}
		if allyRoot == target || allyRoot == actor || diplomacy.SameRealm(gs, target, allyRoot) {
			continue
		}
		if _, exists := seen[allyRoot]; exists {
			continue
		}
		rel := diplomacy.Relation(gs, target, allyRoot)
		if rel == nil || rel.Stance != faction.StanceAllied {
			continue
		}
		seen[allyRoot] = struct{}{}
		allies = append(allies, allyRoot)
	}
	return allies
}

func aiWarBattlefieldRegions(gs *state.GameState, target faction.FactionID) []world.RegionID {
	if gs == nil || target == "" {
		return nil
	}
	owners := map[faction.FactionID]struct{}{target: {}}
	for _, vassalID := range diplomacy.VassalsOf(gs, target) {
		owners[vassalID] = struct{}{}
	}
	regions := make([]world.RegionID, 0, len(gs.Regions))
	for _, region := range aiSortedRegions(gs) {
		if region.IsSea {
			continue
		}
		if _, ok := owners[faction.FactionID(region.OwnerID)]; ok {
			regions = append(regions, region.ID)
		}
	}
	return regions
}

func aiWarWeightedFactionPower(gs *state.GameState, fid faction.FactionID, battlefield []world.RegionID) int {
	power, _ := aiWarWeightedFactionPowerWithDistance(gs, fid, battlefield)
	return power
}

func aiWarWeightedFactionPowerWithDistance(gs *state.GameState, fid faction.FactionID, battlefield []world.RegionID) (int, int) {
	if gs == nil || fid == "" || len(battlefield) == 0 {
		return 0, -1
	}
	total := 0
	nearest := -1
	for _, armyRef := range aiSortedArmies(gs) {
		if armyRef == nil || armyRef.OwnerID != string(fid) {
			continue
		}
		strength := armyRef.TotalStrength(gs.UnitTypes)
		if strength <= 0 && gs.UnitTypes == nil {
			strength = len(armyRef.Units) * 10
		}
		if strength <= 0 {
			continue
		}
		distance := aiWarBattlefieldDistance(gs, armyRef.RegionID, battlefield)
		if distance < 0 {
			continue
		}
		if nearest < 0 || distance < nearest {
			nearest = distance
		}
		total += strength * aiWarSupportPercent(distance) / 100
	}
	return total, nearest
}

func aiWarSupportPercent(distance int) int {
	switch {
	case distance < 0:
		return 0
	case distance <= aiWarSupportNearDistance:
		return 100
	case distance <= aiWarSupportMidDistance:
		return 75
	case distance <= aiWarSupportFarDistance:
		return 50
	case distance <= aiWarSupportMaxDistance:
		return 25
	default:
		return 10
	}
}

func aiWarBattlefieldDistance(gs *state.GameState, start world.RegionID, battlefield []world.RegionID) int {
	if gs == nil || start == "" || len(battlefield) == 0 || gs.Regions[start] == nil {
		return -1
	}
	targets := make(map[world.RegionID]struct{}, len(battlefield))
	for _, regionID := range battlefield {
		targets[regionID] = struct{}{}
	}
	if _, ok := targets[start]; ok {
		return 0
	}
	type queuedRegion struct {
		id       world.RegionID
		distance int
	}
	queue := []queuedRegion{{id: start}}
	visited := map[world.RegionID]struct{}{start: {}}
	for head := 0; head < len(queue); head++ {
		current := gs.Regions[queue[head].id]
		if current == nil {
			continue
		}
		distance := queue[head].distance + 1
		for _, neighborID := range current.Neighbors {
			if _, ok := targets[neighborID]; ok {
				return distance
			}
			if _, ok := visited[neighborID]; ok || gs.Regions[neighborID] == nil {
				continue
			}
			visited[neighborID] = struct{}{}
			queue = append(queue, queuedRegion{id: neighborID, distance: distance})
		}
	}
	return -1
}

// aiNavalWarReady yalnız mevcut stratejik planın somut bir deniz çıkarma
// görevi üretebildiği durumda kara sınırı olmayan hedefe savaş izni verir.
// Böylece her kıyı devleti rastgele denizaşırı savaş açmaz; transport/liman
// hattı kurulabilen tarihsel hedefler ise diplomasi katmanında kilitlenmez.
func aiNavalWarReady(ctx *StrategicContext, target faction.FactionID) bool {
	if ctx == nil || ctx.gs == nil || target == "" || ctx.navalMission == nil {
		return false
	}
	mission := ctx.navalMission
	if mission.Kind != aiNavalMissionAssault || mission.TargetFactionID != target || mission.EmbarkArmyID == "" || mission.EmbarkRegionID == "" || mission.EmbarkSeaRegionID == "" || mission.LandingSeaRegionID == "" {
		return false
	}
	landing := ctx.gs.Regions[mission.TargetRegionID]
	if landing == nil || landing.IsSea || landing.OwnerID != string(target) || !landing.IsCoastal(ctx.gs.Regions) {
		return false
	}
	armyRef := ctx.gs.Armies[mission.EmbarkArmyID]
	if armyRef == nil || armyRef.IsNaval || len(armyRef.Units) == 0 {
		return false
	}
	self := ctx.gs.Factions[ctx.FactionID]
	if self == nil {
		return false
	}
	transportType := ctx.gs.UnitTypes["transport"]
	if transportType == nil || transportType.CarryCapacity <= 0 || !transportType.HasAllRequiredTechs(self.Research.Completed) {
		return false
	}
	return aiSeaRouteDistance(ctx.gs, mission.EmbarkSeaRegionID, mission.LandingSeaRegionID) >= 0 && aiNavalWarPortReady(ctx.gs, ctx.FactionID, mission.EmbarkRegionID)
}

func aiNavalWarPortReady(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) bool {
	if gs == nil || fid == "" || regionID == "" {
		return false
	}
	region := gs.Regions[regionID]
	return region != nil && region.OwnerID == string(fid) && !region.IsSea && aiNavalEmbarkPortViable(gs, fid, region)
}

// aiEvaluateWarOpportunitiesWithSteps selects at most one opportunistic war
// target after diplomacy has resolved peace/alliance/trade actions.
func aiEvaluateWarOpportunitiesWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	if gs == nil || !aiProactiveWarEnabled(gs) {
		return
	}
	self := gs.Factions[fid]
	if self == nil || self.IsEliminated {
		return
	}
	if diplomacy.DirectOverlord(gs, fid) != "" {
		return
	}
	if aiActiveWarCount(gs, fid) >= aiMaxConcurrentWars(gs, fid) || !aiWarCadenceAllows(gs, fid) {
		return
	}

	strategicContext := prepareStrategicContext(gs, fid)
	bestScore := aiWarThresholdForDifficulty(gs)
	bestTarget := faction.FactionID("")
	for _, otherID := range aiSortedFactionIDs(gs) {
		other := gs.Factions[otherID]
		if otherID == fid || other == nil || other.IsEliminated {
			continue
		}
		if overlord := diplomacy.DirectOverlord(gs, otherID); overlord != "" && overlord != fid {
			continue
		}
		rel := diplomacy.Relation(gs, fid, otherID)
		if rel == nil || rel.Stance != faction.StancePeace {
			continue
		}
		score := aiWarOpportunityScoreWithContext(gs, fid, otherID, rel, strategicContext)
		if score > bestScore {
			bestScore = score
			bestTarget = otherID
		}
	}

	if bestTarget == "" {
		return
	}
	result := diplomacy.Execute(gs, fid, bestTarget, diplomacy.ActionDeclareWar)
	if result.Applied || result.Accepted {
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: bestTarget, Message: turnFactionName(gs, fid) + ": " + result.Message})
	}
}

func aiWarOpportunityScore(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation) int {
	return aiWarOpportunityScoreWithContext(gs, actor, target, rel, prepareStrategicContext(gs, actor))
}

func aiWarOpportunityScoreWithContext(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation, strategicContext *StrategicContext) int {
	self := gs.Factions[actor]
	other := gs.Factions[target]
	if self == nil || other == nil || rel == nil {
		return -1
	}
	isExpansionTarget := aiHasExpansionTarget(self, target)
	isPlanTarget := aiPlanTargetsFaction(gs, actor, target)
	maxPeaceScore := -20
	if isPlanTarget {
		maxPeaceScore = 20
	} else if isExpansionTarget {
		maxPeaceScore = 10
	} else if self.AIAggressiveness >= 70 {
		maxPeaceScore = -10
	}
	sharesLandBorder := aiSharesLandBorder(gs, actor, target)
	if rel.Score > maxPeaceScore {
		return -1
	}
	if !sharesLandBorder && !aiNavalWarReady(strategicContext, target) {
		return -1
	}

	coalition := aiWarCoalitionAssessment(gs, actor, target)
	if coalition.AttackerPower <= 0 || (coalition.DefenderPower > 0 && coalition.AttackerPower*100 < coalition.DefenderPower*aiMinAttackPowerPercent(gs)) || !aiStrategicWarReady(strategicContext, target) {
		return -1
	}
	frontierPower := aiFrontierPower(gs, actor, target)
	if frontierPower <= 0 && sharesLandBorder {
		return -1
	}
	targetFrontierPower := aiFrontierPower(gs, target, actor)

	score := 20
	if coalition.DefenderPower == 0 {
		score += 30
	} else {
		score += minInt(30, maxInt(0, (coalition.AttackerPower-coalition.DefenderPower)/12))
	}
	if targetFrontierPower == 0 {
		score += 16
	} else if frontierPower > targetFrontierPower {
		score += minInt(22, (frontierPower-targetFrontierPower)/10+8)
	} else {
		score -= 18
	}
	score += minInt(18, maxInt(0, -rel.Score/2))
	if rel.Score > 0 {
		score -= rel.Score
	}
	selfRegions := len(gs.LandRegionsOwnedBy(actor))
	targetRegions := len(gs.LandRegionsOwnedBy(target))
	if targetRegions <= 2 {
		score += 12
	}
	if coalition.AllySupportPower > 0 {
		// Savunma koalisyonunun cepheye ulaşabilen kısmı, hedefin kendi
		// gücünden ayrı bir risk olarak puanı aşağı çeker. Güç zaten yukarıdaki
		// saldırı eşiğine dahil edildiği için burada ikinci kez sert bir blok
		// oluşturulmaz; yakın müttefik ile uzak müttefik arasındaki fark korunur.
		score -= minInt(24, coalition.AllySupportPower/12)
		score -= minInt(12, coalition.AllyVassalPower/10)
	}
	if coalition.CertainAttackerSupportPower > 0 {
		// Kesin katılacak saldıran müttefikler, hedefe ulaşabilecekleri ölçüde
		// saldırı gücünü artırır. Uzak destek, karar eşiğini yapay biçimde
		// aşmaması için aynı mesafe ağırlığıyla zaten azaltılmıştır.
		score += minInt(18, coalition.CertainAttackerSupportPower/12)
	}
	if selfRegions >= targetRegions {
		score += 8
	}
	if gs.DeployedLandUnits(actor) >= gs.ManpowerCap(actor) {
		score += 8
	}
	score += minInt(15, aiBestBorderTargetValue(gs, actor, target)/15)
	if self.Religion != other.Religion {
		score += 6
	} else {
		score -= 6
	}
	score += (self.AIAggressiveness - 45) / 2
	if isExpansionTarget {
		score += 18
		if rel.Score <= 0 {
			score += 6
		}
		if self.AIAggressiveness >= 60 {
			score += 4
		}
	}
	if isPlanTarget {
		commitment := 50
		if plan := gs.AIPlans[actor]; plan != nil {
			commitment = plan.Commitment
		}
		score += minInt(36, 12+commitment/3)
	}
	if !sharesLandBorder {
		// Deniz aşırı savaş kara sınırı puanını taşımadığı için, yalnızca
		// gerçek bir deniz görevi hazırsa kontrollü bir hazırlık bonusu alır.
		score += 12
	}
	if target == gs.PlayerFactionID {
		score -= 18
		score += aiPlayerTargetScoreBonus(gs)
	}
	return score
}
