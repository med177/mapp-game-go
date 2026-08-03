package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// executeAITerritoryTask, eski çağrı noktaları için stratejik bağlamı olmayan
// görev seçimi girişidir.
func executeAITerritoryTask(gs *state.GameState, a *army.Army, fid faction.FactionID) (TurnStep, bool) {
	return executeAITerritoryTaskWithStrategicContext(gs, a, fid, nil)
}

// executeAITerritoryTaskWithStrategicContext, temas sonrasında düşman
// toprağında kalan AI ordusuna oyuncudaki aynı yağma/pusu görevlerini verir.
// Karar rastgele tur numarasına değil, plan hedefi, ganimet değeri, arazi ve
// yaklaşan düşman kuvvetine dayanır; böylece replay/save-load davranışı da
// deterministik kalır.
func executeAITerritoryTaskWithStrategicContext(gs *state.GameState, a *army.Army, fid faction.FactionID, ctx *StrategicContext) (TurnStep, bool) {
	if gs == nil || a == nil || a.IsNaval || a.RegionID == "" {
		return TurnStep{}, false
	}
	region := gs.Regions[a.RegionID]
	if region == nil || region.OwnerID == "" || region.OwnerID == a.OwnerID || !diplomacy.IsWar(gs, faction.FactionID(a.OwnerID), faction.FactionID(region.OwnerID)) || gs.SiegeByArmy(a.ID) != nil {
		return TurnStep{}, false
	}
	if a.InAmbush {
		a.MovePoints = 0
		return TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: a.ID, FocusRegion: region.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesinde pusuda bekliyor."}, true
	}
	if aiEnemyArmyInRegion(gs, a.OwnerID, region.ID) != nil {
		return TurnStep{}, false
	}
	// Ana fetih hedefinde beklemek, kuşatma/taarruz planını gereksiz yere
	// geciktirir. Bu durumda ordu normal hareket kararına geri döner.
	if aiTerritoryTaskIsPrimaryConquestTarget(gs, fid, region.ID) {
		return TurnStep{}, false
	}

	ownPower, approachingEnemyPower := aiTerritoryTaskPowerBalance(gs, a, region)
	// Düşman karşı taarruzu belirgin biçimde üstünse AI ne yağma uğruna açıkta
	// kalır ne de pusuya yatar; normal geri çekilme/güvenlik ataması devreye girer.
	if approachingEnemyPower > 0 && ownPower > 0 && approachingEnemyPower*100 > ownPower*135 {
		return TurnStep{}, false
	}

	ambushScore := aiTerritoryTaskAmbushScore(region, ownPower, approachingEnemyPower, ctx)
	raidScore := aiTerritoryTaskRaidScore(gs, region, approachingEnemyPower)
	if ambushScore > raidScore && ambushScore > 0 {
		if gs.SetAmbush(a, region) {
			return TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: a.ID, FocusRegion: region.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesinde pusu kurdu."}, true
		}
	}
	if raidScore > 0 && gs.ApplyRaid(a, region) {
		return TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: a.ID, FocusRegion: region.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesini yağmaladı."}, true
	}
	return TurnStep{}, false
}

func aiTerritoryTaskIsPrimaryConquestTarget(gs *state.GameState, fid faction.FactionID, regionID world.RegionID) bool {
	if gs == nil || regionID == "" {
		return false
	}
	plan := gs.AIPlans[fid]
	if plan == nil || plan.Kind != state.AIObjectiveExpand {
		return false
	}
	for _, targetID := range plan.TargetRegionIDs {
		if targetID == regionID {
			return true
		}
	}
	return false
}

// aiTerritoryTaskPowerBalance yalnız komşu düşman ordularını sayar: bu ordular
// bir sonraki harekette pusu/yağma yapılan bölgeye gerçekçi olarak ulaşabilir.
func aiTerritoryTaskPowerBalance(gs *state.GameState, a *army.Army, region *world.Region) (ownPower, approachingEnemyPower int) {
	if gs == nil || a == nil || region == nil {
		return 0, 0
	}
	ownPower = a.TotalStrength(gs.UnitTypes)
	if ownPower <= 0 {
		ownPower = len(a.Units)
	}
	for _, neighborID := range region.Neighbors {
		for _, candidate := range aiSortedArmies(gs) {
			if candidate == nil || candidate.IsNaval || candidate.RegionID != neighborID || candidate.OwnerID == a.OwnerID {
				continue
			}
			if !diplomacy.IsWar(gs, faction.FactionID(a.OwnerID), faction.FactionID(candidate.OwnerID)) {
				continue
			}
			power := candidate.TotalStrength(gs.UnitTypes)
			if power <= 0 {
				power = len(candidate.Units)
			}
			if power > approachingEnemyPower {
				approachingEnemyPower = power
			}
		}
	}
	return ownPower, approachingEnemyPower
}

func aiTerritoryTaskAmbushScore(region *world.Region, ownPower, approachingEnemyPower int, ctx *StrategicContext) int {
	if region == nil || approachingEnemyPower <= 0 {
		return 0
	}
	bonus := world.TerrainData[region.Terrain].AmbushBonus
	if bonus <= 0 {
		return 0
	}
	score := bonus + 35
	if ownPower >= approachingEnemyPower {
		score += 15
	}
	if ctx != nil && ctx.CriticalThreat {
		score += 10
	}
	return score
}

func aiTerritoryTaskRaidScore(gs *state.GameState, region *world.Region, approachingEnemyPower int) int {
	if gs == nil || region == nil {
		return 0
	}
	loot := gs.RaidLootPreview(region)
	score := loot.Gold*3 + loot.Grain + loot.Iron*2 + loot.Timber + loot.Stone + loot.Spice*3 + loot.Cloth*2
	if score <= 0 {
		return 0
	}
	// Yaklaşan kuvvet varsa pusu fırsatı daha kıymetlidir; yine de yüksek
	// değerli ekonomik hedefler yağmalanabilir.
	if approachingEnemyPower > 0 {
		score -= 20
	}
	return score
}
