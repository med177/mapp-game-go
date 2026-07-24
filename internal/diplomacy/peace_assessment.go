package diplomacy

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
)

const (
	peaceDesireThreshold   = 42
	peaceMinimumWarTurns   = 3
	peaceOfferCooldown     = 3
	peaceEmergencyDiscount = 10
)

// PeaceAssessment 1300 senaryosundaki bir savaş tarafının barış isteğini
// açıklar. Score yükseldikçe savaşı bitirme baskısı artar.
type PeaceAssessment struct {
	Score            int
	Threshold        int
	WarTurns         int
	Eligible         bool
	Emergency        bool
	CapitalThreat    bool
	MilitaryCollapse bool
	ObjectiveDone    bool
	Stalemate        bool
	OwnLosses        int
	EnemyLosses      int
	OwnRegionsLost   int
	RegionsGained    int
}

func (a PeaceAssessment) ShouldPropose() bool {
	return a.Eligible && (a.Score >= a.Threshold || a.Stalemate)
}

// AssessPeaceDesire aktif savaşı actor perspektifinden değerlendirir. Bu yeni
// model yalnızca 1300_ottoman_rise için kullanılır.
func AssessPeaceDesire(gs *state.GameState, actor, opponent faction.FactionID) PeaceAssessment {
	assessment := PeaceAssessment{Threshold: peaceDesireThreshold}
	if gs == nil || actor == "" || opponent == "" || actor == opponent || !IsWar(gs, actor, opponent) {
		return assessment
	}
	gs.SyncWarLedgers()
	ledger := gs.WarLedgerFor(actor, opponent)
	if ledger == nil {
		return assessment
	}
	assessment.WarTurns = max(0, gs.Turn-ledger.StartedTurn)
	assessment.OwnLosses, assessment.EnemyLosses = warCasualtiesFor(ledger, actor)
	actorCaptured, opponentCaptured := warCapturesFor(ledger, actor)
	assessment.RegionsGained = actorCaptured
	assessment.OwnRegionsLost = opponentCaptured

	initialActor, _ := warInitialRegionsFor(ledger, actor)
	currentActor := landRegionCount(gs, actor)
	assessment.OwnRegionsLost = max(assessment.OwnRegionsLost, max(0, initialActor-currentActor))

	assessment.Score += min(24, max(0, assessment.WarTurns-peaceMinimumWarTurns)*3)
	assessment.Score += assessment.OwnRegionsLost * 9
	assessment.Score -= assessment.RegionsGained * 7
	if assessment.OwnLosses > assessment.EnemyLosses {
		assessment.Score += min(20, (assessment.OwnLosses-assessment.EnemyLosses)*3)
	} else if assessment.EnemyLosses > assessment.OwnLosses {
		assessment.Score -= min(14, (assessment.EnemyLosses-assessment.OwnLosses)*2)
	}

	actorPower := MilitaryPower(gs, actor)
	opponentPower := MilitaryPower(gs, opponent)
	switch {
	case actorPower == 0 && opponentPower > 0:
		assessment.Score += 45
		assessment.MilitaryCollapse = true
	case opponentPower > 0 && actorPower*100 < opponentPower*55:
		assessment.Score += 38
		assessment.MilitaryCollapse = true
	case opponentPower > 0 && actorPower*100 < opponentPower*80:
		assessment.Score += 22
	case opponentPower > 0 && actorPower < opponentPower:
		assessment.Score += 10
	case actorPower > 0 && actorPower*100 > opponentPower*150:
		assessment.Score -= 18
	case actorPower > 0 && actorPower*100 > opponentPower*120:
		assessment.Score -= 10
	}

	assessment.Score += economicStress(gs, actor)
	if extraWars := activeWarCount(gs, actor) - 1; extraWars > 0 {
		assessment.Score += min(24, extraWars*12)
	}

	assessment.CapitalThreat = capitalUnderWarThreat(gs, actor, opponent)
	if assessment.CapitalThreat {
		assessment.Score += 40
	}
	if capitalUnderWarThreat(gs, opponent, actor) {
		assessment.Score -= 15
	}

	assessment.ObjectiveDone = warObjectiveCompleted(gs, actor, opponent)
	if assessment.ObjectiveDone {
		assessment.Score += 28
	} else if plan := gs.AIPlans[actor]; plan != nil && plan.TargetFactionID == opponent {
		commitmentPenalty := 8 + plan.Commitment/10
		assessment.Score -= min(20, commitmentPenalty)
	}

	if assessment.WarTurns >= 6 && (ledger.LastBattleTurn == 0 || gs.Turn-ledger.LastBattleTurn >= 4) {
		assessment.Score += 10
	}

	assessment.Emergency = assessment.CapitalThreat || assessment.MilitaryCollapse
	if assessment.Emergency {
		assessment.Threshold -= peaceEmergencyDiscount
	}
	assessment.Eligible = assessment.WarTurns >= peaceMinimumWarTurns || assessment.Emergency
	if ledger.LastPeaceOfferTurn > 0 && gs.Turn-ledger.LastPeaceOfferTurn < peaceOfferCooldown {
		assessment.Eligible = false
	}
	assessment.Stalemate = isWarStalemate(gs, actor, opponent, ledger)
	return assessment
}

func isWarStalemate(gs *state.GameState, actor, opponent faction.FactionID, ledger *state.WarLedger) bool {
	if gs == nil || ledger == nil || gs.ScenarioID != "1300_ottoman_rise" {
		return false
	}
	warTurns := gs.Turn - ledger.StartedTurn
	if warTurns < 12 {
		return false
	}
	lastActionTurn := ledger.LastBattleTurn
	if lastActionTurn == 0 {
		lastActionTurn = ledger.StartedTurn
	}
	if gs.Turn-lastActionTurn < 8 {
		return false
	}
	for regionID, siege := range gs.Sieges {
		if siege == nil {
			continue
		}
		attacker := gs.Armies[siege.AttackerArmyID]
		target := gs.Regions[regionID]
		if attacker == nil || target == nil {
			continue
		}
		if (attacker.OwnerID == string(actor) && target.OwnerID == string(opponent)) ||
			(attacker.OwnerID == string(opponent) && target.OwnerID == string(actor)) {
			return false
		}
	}
	return true
}

func warCasualtiesFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.CasualtiesA, ledger.CasualtiesB
	}
	return ledger.CasualtiesB, ledger.CasualtiesA
}

func warCapturesFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.RegionsCapturedA, ledger.RegionsCapturedB
	}
	return ledger.RegionsCapturedB, ledger.RegionsCapturedA
}

func warInitialRegionsFor(ledger *state.WarLedger, actor faction.FactionID) (int, int) {
	if ledger == nil {
		return 0, 0
	}
	if actor == ledger.FactionA {
		return ledger.InitialRegionsA, ledger.InitialRegionsB
	}
	return ledger.InitialRegionsB, ledger.InitialRegionsA
}

func activeWarCount(gs *state.GameState, actor faction.FactionID) int {
	count := 0
	for _, rel := range gs.Relations {
		if rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		if rel.FactionA == actor || rel.FactionB == actor {
			count++
		}
	}
	return count
}

func capitalUnderWarThreat(gs *state.GameState, actor, opponent faction.FactionID) bool {
	if gs == nil || actor == "" || opponent == "" {
		return false
	}
	f := gs.Factions[actor]
	if f == nil || f.CapitalSettlementID == "" {
		return landRegionCount(gs, actor) <= 1 && MilitaryPower(gs, actor) == 0
	}
	capitalRegion, _, _, ok := gs.FindSettlementByID(f.CapitalSettlementID)
	if !ok || capitalRegion == nil || capitalRegion.OwnerID != string(actor) {
		return true
	}
	if siege := gs.SiegeAt(capitalRegion.ID); siege != nil && siege.AttackerFactionID == string(opponent) {
		return true
	}
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.IsNaval || armyRef.OwnerID != string(opponent) {
			continue
		}
		if armyRef.RegionID == capitalRegion.ID {
			return true
		}
		for _, neighborID := range capitalRegion.Neighbors {
			if armyRef.RegionID == neighborID {
				return true
			}
		}
	}
	return false
}

func warObjectiveCompleted(gs *state.GameState, actor, opponent faction.FactionID) bool {
	if gs == nil {
		return false
	}
	plan := gs.AIPlans[actor]
	if plan == nil || plan.Kind != state.AIObjectiveExpand || plan.TargetFactionID != opponent || len(plan.TargetRegionIDs) == 0 {
		return false
	}
	for _, regionID := range plan.TargetRegionIDs {
		region := gs.Regions[regionID]
		if region != nil && region.OwnerID == string(opponent) {
			return false
		}
	}
	return true
}
