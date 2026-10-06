package ai

import (
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

// aiResearch selects and starts one technology when the faction is idle.
func aiResearch(gs *state.GameState, fid faction.FactionID) {
	aiResearchWithSteps(gs, fid, nil)
}

func aiResearchWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiResearchWithBudgetAndSteps(gs, fid, nil, steps)
}

func aiResearchWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) {
	aiResearchWithStrategicContextAndSteps(gs, fid, budget, nil, steps)
}

func aiResearchWithStrategicContextAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, strategicContext *StrategicContext, steps *[]TurnStep) {
	f := gs.Factions[fid]
	if f.IsEliminated || gs.TechTypes == nil || f.Research.ActiveID != "" {
		return
	}
	if budget == nil && f.Gold < aiTechReserve {
		return
	}
	best := aiSelectResearchTechnology(gs, f, budget, strategicContext)
	if best == nil {
		return
	}
	researchEffects := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
	researchCost := tech.EffectiveResearchGoldCost(best, researchEffects.ResearchCostMod)
	if tech.StartResearchWithModifiers(&f.Research, best, &f.Gold, researchEffects.ResearchSpeedMod, researchEffects.ResearchCostMod) {
		if budget != nil {
			budget.consume(aiBudgetResearch, researchCost)
		}
		addTurnStep(steps, TurnStep{FactionID: fid, Kind: TurnStepResearch, Message: turnFactionName(gs, fid) + " " + best.NameTR + " araştırmasını başlattı."})
	}
}

// aiEconomyBuild, bütçe ve kaynaklar elverdiği sürece her uygun bina türü için
// tek bir kademe yatırım emri verir. Aynı bina türünde bekleyen emir varsa
// strateji o türü atlar; farklı bina türleri aynı bölgede paralel ilerleyebilir.
func aiEconomyBuild(gs *state.GameState, fid faction.FactionID) {
	aiEconomyBuildWithSteps(gs, fid, nil)
}

func aiEconomyBuildWithSteps(gs *state.GameState, fid faction.FactionID, steps *[]TurnStep) {
	aiEconomyBuildWithBudgetAndSteps(gs, fid, nil, steps)
}

func aiEconomyBuildWithBudgetAndSteps(gs *state.GameState, fid faction.FactionID, budget *aiBudget, steps *[]TurnStep) {
	aiEconomyBuildWithStrategicContextAndSteps(gs, fid, budget, nil, steps)
}
