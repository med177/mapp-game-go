package ai

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

// executeAITerritoryTask, düşman toprağında hareket sonrası kalmayı seçen AI
// ordularına oyuncudaki aynı görevleri verir. Pusu seçimi deterministiktir;
// böylece replay ve save/load sonrasında AI davranışı değişmez.
func executeAITerritoryTask(gs *state.GameState, a *army.Army, fid faction.FactionID) (TurnStep, bool) {
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
	// Her üçüncü AI turunda arazi uygunsa pusu, diğer turlarda yağma seçilir.
	// Böylece AI her iki görevi de kullanır ve pusu sürekli görünmezlikte kalmaz.
	if world.TerrainData[region.Terrain].AmbushBonus > 0 && gs.Turn%3 == 0 {
		if gs.SetAmbush(a, region) {
			return TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: a.ID, FocusRegion: region.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesinde pusu kurdu."}, true
		}
	}
	if gs.ApplyRaid(a, region) {
		return TurnStep{FactionID: fid, Kind: TurnStepInfo, ArmyID: a.ID, FocusRegion: region.ID, Message: turnFactionName(gs, fid) + " " + turnRegionName(gs, region.ID) + " bölgesini yağmaladı."}, true
	}
	return TurnStep{}, false
}
