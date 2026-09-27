package ai

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

const (
	aiWarSupportNearDistance    = 1
	aiWarSupportMidDistance     = 3
	aiWarSupportFarDistance     = 5
	aiWarSupportMaxDistance     = 8
	aiHistoricalWarThreshold    = 42
	aiRapidConquestPowerPercent = 85
	aiReliableAllyCallPercent   = 70
)

type aiWarCoalitionRisk struct {
	TargetPower                     int
	TargetVassalPower               int
	AllyPower                       int
	AllyVassalPower                 int
	AllySupportPower                int
	DefenderPower                   int
	AttackerPower                   int
	AttackerVassalPower             int
	CertainAttackerAllyPower        int
	CertainAttackerAllyVassalPower  int
	CertainAttackerSupportPower     int
	ReliableAttackerAllyPower       int
	ReliableAttackerAllyVassalPower int
	ReliableAttackerSupportPower    int
	NearestAllyArmy                 int
	NearestAttackerAllyArmy         int
}

// aiWarCoalitionAssessment, savaş ilanı öncesinde iki tarafın etkin koalisyon
// gücünü ölçer. Savunan tarafta hedefin dış müttefikleri, saldıran tarafta ise
// AssessWarCall tarafından AutoJoin olarak işaretlenen kesin müttefikler ile
// çağrıyı güvenilir biçimde kabul edecek AI müttefikleri kullanılır. Müttefik
// katkısı, ordularının hedef ülke topraklarına olan rota mesafesiyle
// ağırlıklandırılır; böylece haritanın diğer ucundaki kuvvet tam cephe gücü
// gibi değerlendirilmez.
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
	assessment.AttackerPower = aiFactionMilitaryPowerAsSeenBy(gs, actorRoot, actorRoot)
	for _, vassalID := range diplomacy.VassalsOf(gs, actorRoot) {
		assessment.AttackerVassalPower += aiWarWeightedFactionPowerAsSeenBy(gs, actorRoot, vassalID, battlefield)
	}
	assessment.AttackerPower += assessment.AttackerVassalPower

	for _, allyRoot := range aiWarExternalAllies(gs, actorRoot, targetRoot) {
		if aiWarAllyJoinIsPending(gs, allyRoot) {
			continue
		}
		call := diplomacy.AssessWarCall(gs, actorRoot, allyRoot, targetRoot)
		if !call.AutoJoin && call.Chance < aiReliableAllyCallPercent {
			continue
		}
		allyPower, nearest := aiWarWeightedFactionPowerWithDistanceAsSeenBy(gs, actorRoot, allyRoot, battlefield)
		vassalPower := 0
		for _, vassalID := range diplomacy.VassalsOf(gs, allyRoot) {
			vassalPower += aiWarWeightedFactionPowerAsSeenBy(gs, actorRoot, vassalID, battlefield)
		}
		if call.AutoJoin {
			assessment.CertainAttackerAllyPower += allyPower
			assessment.CertainAttackerAllyVassalPower += vassalPower
			assessment.CertainAttackerSupportPower += allyPower + vassalPower
		} else {
			assessment.ReliableAttackerAllyPower += allyPower
			assessment.ReliableAttackerAllyVassalPower += vassalPower
			assessment.ReliableAttackerSupportPower += allyPower + vassalPower
		}
		assessment.AttackerPower += allyPower + vassalPower
		if nearest >= 0 && (assessment.NearestAttackerAllyArmy < 0 || nearest < assessment.NearestAttackerAllyArmy) {
			assessment.NearestAttackerAllyArmy = nearest
		}
	}

	assessment.TargetPower = aiFactionMilitaryPowerAsSeenBy(gs, actorRoot, targetRoot)
	for _, vassalID := range diplomacy.VassalsOf(gs, targetRoot) {
		assessment.TargetVassalPower += aiWarWeightedFactionPowerAsSeenBy(gs, actorRoot, vassalID, battlefield)
	}
	assessment.DefenderPower = assessment.TargetPower + assessment.TargetVassalPower

	for _, allyRoot := range aiWarExternalAllies(gs, targetRoot, actorRoot) {
		allyPower, nearest := aiWarWeightedFactionPowerWithDistanceAsSeenBy(gs, actorRoot, allyRoot, battlefield)
		vassalPower := 0
		for _, vassalID := range diplomacy.VassalsOf(gs, allyRoot) {
			vassalPower += aiWarWeightedFactionPowerAsSeenBy(gs, actorRoot, vassalID, battlefield)
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

func aiWarWeightedFactionPowerAsSeenBy(gs *state.GameState, observer, fid faction.FactionID, battlefield []world.RegionID) int {
	weighted := aiWarWeightedFactionPower(gs, fid, battlefield)
	if gs == nil || observer == fid || weighted <= 0 {
		return weighted
	}
	exact := aiFactionMilitaryPower(gs, fid)
	if exact <= 0 {
		return weighted
	}
	return weighted * aiFactionMilitaryPowerAsSeenBy(gs, observer, fid) / exact
}

func aiWarWeightedFactionPowerWithDistanceAsSeenBy(gs *state.GameState, observer, fid faction.FactionID, battlefield []world.RegionID) (int, int) {
	weighted, nearest := aiWarWeightedFactionPowerWithDistance(gs, fid, battlefield)
	if gs == nil || observer == fid || weighted <= 0 {
		return weighted, nearest
	}
	exact := aiFactionMilitaryPower(gs, fid)
	if exact <= 0 {
		return weighted, nearest
	}
	return weighted * aiFactionMilitaryPowerAsSeenBy(gs, observer, fid) / exact, nearest
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
		strength := aiArmyStrength(gs, armyRef)
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

const (
	aiRapidExpansionAbsoluteGain = 4
	aiRapidExpansionRelativeGain = 3
	aiRapidExpansionRelativeRate = 30
)

// aiFactionRapidExpansion bir fraksiyonun toplam büyüklüğünü değil, son kısa
// penceredeki ani toprak kazanımını tehdit kabul eder. Üç bölge küçük bir
// devlet için çok büyük bir sıçramaysa veya dört bölge mutlak olarak
// kazanılmışsa koalisyon baskısı açılabilir.
func aiFactionRapidExpansion(gs *state.GameState, fid faction.FactionID) bool {
	if gs == nil || fid == "" || gs.Factions[fid] == nil || gs.Factions[fid].IsEliminated {
		return false
	}
	gained := gs.RecentRegionGain(fid)
	if gained >= aiRapidExpansionAbsoluteGain {
		return true
	}
	if gained < aiRapidExpansionRelativeGain {
		return false
	}
	current := len(gs.LandRegionsOwnedBy(fid))
	return current > 0 && gained*100 >= current*aiRapidExpansionRelativeRate
}

// aiOverextensionWarScoreAdjustment, AI'nin kendi aşırı genişlemesini yeni
// saldırı kararına yansıtır. Agresif devletler daha yüksek riski göze alır;
// temkinli devletler aynı toprak baskısında savaş puanını düşürüp önce
// toparlanmayı seçer. Bu yalnız karar puanını etkiler, hukuki savaş kurallarını
// veya gerçek muharebe gücünü değiştirmez.
func aiOverextensionWarScoreAdjustment(gs *state.GameState, actor faction.FactionID) int {
	if gs == nil || actor == "" {
		return 0
	}
	self := gs.Factions[actor]
	if self == nil || self.IsEliminated {
		return 0
	}
	score := gs.OverextensionScore(actor)
	if score <= 0 {
		return 0
	}
	tolerance := 45 + self.AIAggressiveness/2
	if tolerance > 90 {
		tolerance = 90
	}
	if score <= tolerance {
		return 0
	}
	divisor := 2
	if self.AIAggressiveness >= 70 {
		divisor = 3
	}
	penalty := maxInt(1, (score-tolerance)/divisor)
	return -minInt(20, penalty)
}

// aiCoalitionWarCandidate, eski koalisyon yolunun herkesi hedefe yöneltmesi
// yerine normal savaş kararının bütün siyasi ve operasyonel kapılarını uygular.
// Böylece hızlı büyüyen hedefle müttefik veya ticaret ilişkisi güçlü bir devlet,
// uzak ve hazırlıksız bir devlet ya da kapasitesini doldurmuş bir devlet
// otomatik olarak koalisyona çekilmez.
func aiCoalitionWarCandidate(gs *state.GameState, actor, target faction.FactionID, ctx *StrategicContext) bool {
	if gs == nil || actor == "" || target == "" || actor == target {
		return false
	}
	self := gs.Factions[actor]
	other := gs.Factions[target]
	if self == nil || other == nil || self.IsEliminated || other.IsEliminated {
		return false
	}
	if diplomacy.DirectOverlord(gs, actor) != "" || diplomacy.SameRealm(gs, actor, target) {
		return false
	}
	if aiActiveWarCount(gs, actor) >= aiMaxConcurrentWars(gs, actor) || !aiWarCadenceAllows(gs, actor) {
		return false
	}
	rel := diplomacy.Relation(gs, actor, target)
	if rel == nil || rel.Stance == faction.StanceWar || rel.Stance == faction.StanceAllied {
		return false
	}
	if rel.Stance == faction.StanceTrade && rel.Score >= 15 {
		return false
	}
	if gs.TruceRemaining(actor, target) > 0 {
		return false
	}
	if ctx == nil {
		ctx = prepareStrategicContext(gs, actor)
	}
	return aiWarOpportunityScoreWithContext(gs, actor, target, rel, ctx) >= aiWarThresholdForDifficulty(gs)
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
	if aiEvaluateHistoricalWarOpportunity(gs, fid, strategicContext, steps) {
		return
	}
	bestScore := aiWarThresholdForDifficulty(gs)
	bestTarget := faction.FactionID("")
	recklessScore := -1
	recklessTarget := faction.FactionID("")
	recklessSelected := false
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
		if score < 0 {
			if nearMiss := aiRecklessWarCandidateScore(gs, fid, otherID, rel, strategicContext); nearMiss > recklessScore {
				recklessScore = nearMiss
				recklessTarget = otherID
			}
		}
	}

	if bestTarget == "" {
		if recklessTarget == "" || !aiRecklessWarRoll(gs, fid, self) {
			return
		}
		bestTarget = recklessTarget
		recklessSelected = true
	}
	result := diplomacy.ExecuteWarDeclaration(gs, fid, bestTarget, nil)
	if result.Applied {
		if recklessSelected {
			if ledger := gs.WarLedgerFor(fid, bestTarget); ledger != nil {
				ledger.RecklessDeclaration = true
			}
		}
		gs.QueueNavalContactForWar(fid, bestTarget)
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: bestTarget, Message: turnFactionName(gs, fid) + ": " + result.Message, WarDeclaration: &result})
	}
}

// aiRecklessWarCandidateScore, normal savaş hazırlığına takılan ama siyasi ve
// coğrafi olarak makul bir hedefi seçer. Bu yol yalnızca nadir bir zarla açılır;
// diplomasi ActionBlockReason yine bütün hukuki kuralları korur.
func aiRecklessWarCandidateScore(gs *state.GameState, actor, target faction.FactionID, rel *faction.Relation, ctx *StrategicContext) int {
	if gs == nil || rel == nil || rel.Stance != faction.StancePeace || actor == target {
		return -1
	}
	if !aiSharesLandBorder(gs, actor, target) || diplomacy.SameRealm(gs, actor, target) || diplomacy.DirectOverlord(gs, target) != "" {
		return -1
	}
	if ctx != nil && ctx.CriticalThreat {
		return -1
	}
	if rel.Score > 35 || aiFrontierPower(gs, actor, target) <= 0 {
		return -1
	}
	actorPower := aiFactionMilitaryPowerAsSeenBy(gs, actor, actor)
	targetPower := aiFactionMilitaryPowerAsSeenBy(gs, actor, target)
	if actorPower <= 0 || targetPower <= 0 || actorPower*100 >= targetPower*110 {
		return -1
	}
	score := 50 - rel.Score
	score += minInt(24, aiBestBorderTargetValue(gs, actor, target)/12)
	score += minInt(15, gs.Factions[actor].AIAggressiveness/5)
	if targetPower > actorPower*2 {
		score -= 20
	}
	return score
}

func aiRecklessWarRoll(gs *state.GameState, fid faction.FactionID, self *faction.Faction) bool {
	if gs == nil || self == nil {
		return false
	}
	chance := 1
	if self.AIAggressiveness >= 60 {
		chance++
	}
	if self.AIAggressiveness >= 80 {
		chance++
	}
	return aiDecisionRoll(gs, fid, "", "reckless_war") < chance
}

// aiEvaluateHistoricalWarOpportunity gives an active expansion plan a direct
// path to war. Generic opportunity scanning still handles every other target,
// but a historical objective may use a lower declaration threshold once its
// coalition and frontier are genuinely ready.
func aiEvaluateHistoricalWarOpportunity(gs *state.GameState, fid faction.FactionID, ctx *StrategicContext, steps *[]TurnStep) bool {
	if gs == nil || fid == "" {
		return false
	}
	plan := gs.AIPlans[fid]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID == "" || plan.TargetFactionID == fid {
		return false
	}
	target := plan.TargetFactionID
	rel := diplomacy.Relation(gs, fid, target)
	if rel == nil || rel.Stance != faction.StancePeace {
		return false
	}
	if score := aiWarOpportunityScoreWithContext(gs, fid, target, rel, ctx); score < aiHistoricalWarThreshold {
		return false
	}
	result := diplomacy.ExecuteWarDeclaration(gs, fid, target, nil)
	if !result.Applied {
		return false
	}
	gs.QueueNavalContactForWar(fid, target)
	addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepDiplomacy, TargetFaction: target, Message: turnFactionName(gs, fid) + " tarihsel hedefi için " + result.Message, WarDeclaration: &result})
	return true
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
	isVictoryTarget := false
	if plan := gs.AIPlans[actor]; plan != nil {
		isVictoryTarget = plan.Kind == state.AIObjectiveExpand && plan.TargetFactionID == target && aiPlanIsVictoryObjective(plan)
	}
	maxPeaceScore := -20
	if isVictoryTarget {
		maxPeaceScore = 25
	} else if isPlanTarget {
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
	requiredPowerPercent := aiMinAttackPowerPercent(gs)
	if isPlanTarget && aiRapidConquestOpportunity(gs, actor, target, coalition) {
		requiredPowerPercent = aiRapidConquestPowerPercent
	}
	if coalition.AttackerPower <= 0 || (coalition.DefenderPower > 0 && coalition.AttackerPower*100 < coalition.DefenderPower*requiredPowerPercent) || !aiStrategicWarReady(strategicContext, target) {
		return -1
	}
	frontierPower := aiFrontierPowerAsSeenBy(gs, actor, actor, target)
	if frontierPower <= 0 && sharesLandBorder {
		return -1
	}
	targetFrontierPower := aiFrontierPowerAsSeenBy(gs, actor, target, actor)

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
		if requiredPowerPercent == aiRapidConquestPowerPercent {
			score += 16
		}
	}
	if isVictoryTarget {
		score += 12
	}
	if !sharesLandBorder {
		// Deniz aşırı savaş kara sınırı puanını taşımadığı için, yalnızca
		// gerçek bir deniz görevi hazırsa kontrollü bir hazırlık bonusu alır.
		score += 12
	}
	score += aiOverextensionWarScoreAdjustment(gs, actor)
	if target == gs.PlayerFactionID {
		score -= 18
		score += aiPlayerTargetScoreBonus(gs)
	}
	return score
}

// aiRapidConquestOpportunity permits a bounded historical opening war even
// when total coalition power is not yet 115%: the attacking state must already
// dominate the actual border force and have a concrete plan region on that
// border. It never bypasses logistics, rally, or defender-alliance checks.
func aiRapidConquestOpportunity(gs *state.GameState, actor, target faction.FactionID, coalition aiWarCoalitionRisk) bool {
	if gs == nil || !aiSharesLandBorder(gs, actor, target) || coalition.DefenderPower <= 0 {
		return false
	}
	plan := gs.AIPlans[actor]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != target {
		return false
	}
	frontierPower := aiFrontierPowerAsSeenBy(gs, actor, actor, target)
	targetFrontierPower := aiFrontierPowerAsSeenBy(gs, actor, target, actor)
	if frontierPower <= 0 || (targetFrontierPower > 0 && frontierPower*100 < targetFrontierPower*125) {
		return false
	}
	for _, regionID := range plan.TargetRegionIDs {
		region := gs.Regions[regionID]
		if region == nil || region.IsSea || region.OwnerID != string(target) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == string(actor) {
				return true
			}
		}
	}
	return false
}
