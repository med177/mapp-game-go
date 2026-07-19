package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const strategicAllianceAcceptanceFloor = 12

// StrategicAllianceAssessment 1300 senaryosunda actor'un target ile ittifaktan
// elde edeceği stratejik değeri bileşenlerine ayırır.
type StrategicAllianceAssessment struct {
	Score                   int
	ThreatValue             int
	BufferValue             int
	FrontSupportValue       int
	TradeValue              int
	PartnerSupportValue     int
	ExpansionTensionPenalty int
	ActiveObjectiveConflict bool
	CommonEnemy             bool
	SharedMajorThreat       bool
}

// AssessStrategicAlliance değerlendirmeyi actor perspektifinden yapar. Teklif
// sahibinin girişim kararı ve hedef AI'nin kabul kararı aynı bileşenleri ters
// perspektiflerden okuyabilir.
func AssessStrategicAlliance(gs *state.GameState, actor, target faction.FactionID) StrategicAllianceAssessment {
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || actor == "" || target == "" || actor == target {
		return StrategicAllianceAssessment{}
	}
	return assessStrategicAlliance(gs, actor, target, HasCommonEnemy(gs, actor, target), HasSharedMajorThreat(gs, actor, target))
}

func assessStrategicAlliance(gs *state.GameState, actor, target faction.FactionID, commonEnemy, sharedMajorThreat bool) StrategicAllianceAssessment {
	assessment := StrategicAllianceAssessment{}
	if gs == nil || gs.ScenarioID != "1300_ottoman_rise" || actor == "" || target == "" || actor == target {
		return assessment
	}
	assessment.ActiveObjectiveConflict = activeAllianceObjectiveConflict(gs, actor, target)
	assessment.CommonEnemy = commonEnemy
	assessment.SharedMajorThreat = sharedMajorThreat
	if assessment.CommonEnemy {
		assessment.ThreatValue += 20
	}
	if assessment.SharedMajorThreat {
		assessment.ThreatValue += 18
	}

	threats := allianceThreatsAgainst(gs, actor, target)
	assessment.BufferValue = allianceBufferValue(gs, target, threats)
	assessment.FrontSupportValue = allianceFrontSupportValue(gs, target, threats)
	assessment.TradeValue = allianceTradeValue(gs, actor, target)
	assessment.PartnerSupportValue = alliancePartnerSupportValue(gs, target)
	if staticAllianceExpansionTension(gs, actor, target) {
		assessment.ExpansionTensionPenalty = 18
	}
	assessment.Score = assessment.ThreatValue +
		assessment.BufferValue +
		assessment.FrontSupportValue +
		assessment.TradeValue +
		assessment.PartnerSupportValue -
		assessment.ExpansionTensionPenalty
	return assessment
}

func activeAllianceObjectiveConflict(gs *state.GameState, a, b faction.FactionID) bool {
	if gs == nil {
		return false
	}
	if plan := gs.AIPlans[a]; plan != nil && plan.TargetFactionID == b {
		return true
	}
	if plan := gs.AIPlans[b]; plan != nil && plan.TargetFactionID == a {
		return true
	}
	return false
}

func staticAllianceExpansionTension(gs *state.GameState, a, b faction.FactionID) bool {
	return factionTargets(gs, a, b) || factionTargets(gs, b, a)
}

func factionTargets(gs *state.GameState, actor, target faction.FactionID) bool {
	if gs == nil {
		return false
	}
	f := gs.Factions[actor]
	if f == nil {
		return false
	}
	for _, targetID := range f.AIExpansionTargets {
		if targetID == target {
			return true
		}
	}
	return false
}

func allianceThreatsAgainst(gs *state.GameState, actor, candidate faction.FactionID) map[faction.FactionID]struct{} {
	threats := make(map[faction.FactionID]struct{})
	potentialThreats := make(map[faction.FactionID]struct{})
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(candidate) {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" {
				continue
			}
			potentialThreats[faction.FactionID(neighbor.OwnerID)] = struct{}{}
		}
	}
	for otherID := range potentialThreats {
		other := gs.Factions[otherID]
		if otherID == actor || otherID == candidate || other == nil || other.IsEliminated {
			continue
		}
		if IsWar(gs, actor, otherID) || isMajorThreatTo(gs, otherID, actor) {
			threats[otherID] = struct{}{}
		}
	}
	return threats
}

func allianceBufferValue(gs *state.GameState, candidate faction.FactionID, threats map[faction.FactionID]struct{}) int {
	if gs == nil || len(threats) == 0 {
		return 0
	}
	borderedThreats := 0
	for threatID := range threats {
		if sharesBorder(gs, candidate, threatID) {
			borderedThreats++
		}
	}
	if borderedThreats == 0 {
		return 0
	}
	return min(16, 7+borderedThreats*3+landRegionCount(gs, candidate)/2)
}

func allianceFrontSupportValue(gs *state.GameState, candidate faction.FactionID, threats map[faction.FactionID]struct{}) int {
	if gs == nil || len(threats) == 0 {
		return 0
	}
	frontPower := 0
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.IsNaval || armyRef.OwnerID != string(candidate) {
			continue
		}
		region := gs.Regions[armyRef.RegionID]
		if region == nil {
			continue
		}
		atThreatFront := false
		for _, neighborID := range region.Neighbors {
			neighbor := gs.Regions[neighborID]
			if neighbor == nil {
				continue
			}
			if _, ok := threats[faction.FactionID(neighbor.OwnerID)]; ok {
				atThreatFront = true
				break
			}
		}
		if !atThreatFront {
			continue
		}
		if gs.UnitTypes != nil {
			frontPower += armyRef.TotalStrength(gs.UnitTypes)
		} else {
			frontPower += len(armyRef.Units) * 10
		}
	}
	return min(16, frontPower/15)
}

func allianceTradeValue(gs *state.GameState, actor, target faction.FactionID) int {
	value := 0
	if HasTradeRouteBetween(gs, actor, target) {
		value += 10
	} else if CanEstablishTradeRoute(gs, actor, target) {
		value += 5
	}
	value += min(6, totalTradeCapacity(gs, target)/4)
	return value
}

func alliancePartnerSupportValue(gs *state.GameState, target faction.FactionID) int {
	if gs == nil {
		return 0
	}
	return min(14, MilitaryPower(gs, target)/20) + min(8, landRegionCount(gs, target)*2)
}
