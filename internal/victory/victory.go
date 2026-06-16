package victory

import (
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const (
	defaultDominationTarget = 15
	defaultGoldThreshold    = 500
	defaultGoldHoldTurns    = 5
	defaultArmyStrength     = 200
	defaultDefeated         = 3
)

// Check her tur sonunda zafer/mağlubiyet koşullarını kontrol eder.
func Check(gs *state.GameState) {
	if gs.Phase == state.PhaseGameOver {
		return
	}

	playerRegions := gs.RegionsOwnedBy(gs.PlayerFactionID)

	// Oyuncu elendi mi?
	if len(playerRegions) == 0 && gs.Turn > 1 {
		gs.Phase = state.PhaseGameOver
		gs.WinnerID = ""
		return
	}

	// Seçilen zafer tipine göre kontrol
	switch gs.Victory.Type {
	case state.VictoryDomination, "":
		checkDomination(gs, playerRegions)
	case state.VictoryEconomic:
		checkEconomic(gs)
	case state.VictoryMilitary:
		checkMilitary(gs)
	case state.VictoryReligious:
		checkReligious(gs, playerRegions)
	case state.VictoryConquerCity:
		checkConquerCity(gs)
	case state.VictorySurviveTurns:
		checkSurviveTurns(gs)
	}

	if gs.Phase == state.PhaseGameOver || gs.VictoryAchieved {
		return
	}

	if deadlineExpired(gs) {
		gs.Phase = state.PhaseGameOver
		gs.WinnerID = ""
	}
}

func deadlineExpired(gs *state.GameState) bool {
	if gs == nil || gs.Victory.DeadlineYear <= 0 {
		return false
	}

	deadlineMonth := gs.Victory.DeadlineMonth
	if deadlineMonth <= 0 || deadlineMonth > 12 {
		deadlineMonth = 12
	}

	if gs.Year > gs.Victory.DeadlineYear {
		return true
	}
	if gs.Year < gs.Victory.DeadlineYear {
		return false
	}
	return gs.Month > deadlineMonth
}

// checkConquerCity gerekli tüm hedef bölgeler oyuncuya geçtiğinde zafer verir.
func checkConquerCity(gs *state.GameState) {
	if len(gs.Victory.RequiredRegions) == 0 {
		return
	}
	for _, targetID := range gs.Victory.RequiredRegions {
		region, ok := gs.Regions[targetID]
		if !ok || region.OwnerID != string(gs.PlayerFactionID) {
			return
		}
	}
	markPlayerVictory(gs)
}

// checkDomination bölge sayısına ve kritik bölgelere göre zafer kontrol eder.
func checkDomination(gs *state.GameState, playerRegions []*world.Region) {
	target := gs.Victory.TargetRegionCount
	if target == 0 {
		target = defaultDominationTarget
	}
	if len(playerRegions) < target {
		return
	}
	// Zorunlu bölgeler var mı?
	for _, rid := range gs.Victory.RequiredRegions {
		region, ok := gs.Regions[rid]
		if !ok || region.OwnerID != string(gs.PlayerFactionID) {
			return
		}
	}
	markPlayerVictory(gs)
}

// checkEconomic altın miktarını belirli tur süre boyunca koruma zaferini kontrol eder.
func checkEconomic(gs *state.GameState) {
	threshold := gs.Victory.TargetGoldIncome
	if threshold == 0 {
		threshold = defaultGoldThreshold
	}
	holdTurns := gs.Victory.GoldHoldTurns
	if holdTurns == 0 {
		holdTurns = defaultGoldHoldTurns
	}

	if gs.Factions[gs.PlayerFactionID] == nil {
		return
	}
	if CurrentGoldIncome(gs) >= threshold {
		gs.EconomicVictoryTurns++
		if gs.EconomicVictoryTurns >= holdTurns {
			markPlayerVictory(gs)
		}
	} else {
		gs.EconomicVictoryTurns = 0
	}
}

// CurrentGoldIncome oyuncunun mevcut tur başı altın gelirini hesaplar.
func CurrentGoldIncome(gs *state.GameState) int {
	if gs == nil {
		return 0
	}
	fid := gs.PlayerFactionID
	if gs.Factions[fid] == nil {
		return 0
	}

	income := 0
	season := gs.CurrentSeason()
	seasonMod := season.TradeMod()
	harvestMod := season.HarvestMod()
	var fx tech.Effects
	if gs.TechTypes != nil {
		fx = tech.ComputeEffects(gs.Factions[fid].Research.Completed, gs.TechTypes)
	}
	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) {
			continue
		}
		goldMod := 1.0
		tradeCapMod := 1.0
		for _, bid := range region.Buildings {
			if building, ok := gs.BuildingTypes[bid]; ok {
				goldMod *= building.GoldMod
				tradeCapMod *= building.TradeCapacityMod
			}
		}
		income += int(float64(region.GoldIncome()) * goldMod * float64(harvestMod) / 100)
		tradeIncome := economy.RegionTradeIncome(region.TradeCapacity, tradeCapMod)
		tradeIncome = tradeIncome * seasonMod / 100
		if fx.MarketGoldMod != 0 {
			tradeIncome = int(float64(tradeIncome) * (1.0 + fx.MarketGoldMod))
		}
		income += tradeIncome
	}

	for _, route := range gs.TradeRoutes {
		if route != nil && route.FromFactionID == string(fid) {
			income += route.GoldEarned()
		}
	}

	if gs.TechTypes != nil {
		income += fx.GoldPerRegion * len(gs.RegionsOwnedBy(fid))
	}

	return max(income, 0)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// checkMilitary ordu gücü ve fraksiyon yenilgisi sayısına göre zafer kontrol eder.
func checkMilitary(gs *state.GameState) {
	// Elenmiş fraksiyon sayısını güncelle
	eliminated := 0
	for fid, f := range gs.Factions {
		if fid != gs.PlayerFactionID && f.IsEliminated {
			eliminated++
		}
	}
	gs.FactionsEliminated = eliminated

	targetStr := gs.Victory.TargetArmyStrength
	if targetStr == 0 {
		targetStr = defaultArmyStrength
	}
	targetDefeated := gs.Victory.TargetDefeated
	if targetDefeated == 0 {
		targetDefeated = defaultDefeated
	}

	// Oyuncunun toplam ordu gücünü hesapla
	totalStr := 0
	for _, a := range gs.Armies {
		if a.OwnerID == string(gs.PlayerFactionID) {
			totalStr += a.TotalStrength(gs.UnitTypes)
		}
	}

	if totalStr >= targetStr && eliminated >= targetDefeated {
		markPlayerVictory(gs)
	}
}

// checkReligious kutsal şehirlerin oyuncu tarafından belirli tur süre tutulması zaferini kontrol eder.
func checkReligious(gs *state.GameState, _ []*world.Region) {
	if len(gs.Victory.RequiredRegions) == 0 {
		return
	}

	allHeld := true
	for _, rid := range gs.Victory.RequiredRegions {
		region, ok := gs.Regions[rid]
		if !ok || region.OwnerID != string(gs.PlayerFactionID) {
			allHeld = false
			break
		}
	}

	if allHeld {
		gs.ReligiousVictoryTurns++
		// 12 tur (~1 yıl) kutsal şehirleri tutmak = zafer
		if gs.ReligiousVictoryTurns >= 12 {
			markPlayerVictory(gs)
		}
	} else {
		gs.ReligiousVictoryTurns = 0
	}
}

func checkSurviveTurns(gs *state.GameState) {
	target := gs.Victory.TargetTurns
	if target == 0 {
		target = 60
	}
	if gs.Turn >= target {
		markPlayerVictory(gs)
	}
}

func markPlayerVictory(gs *state.GameState) {
	if gs == nil || gs.VictoryAchieved {
		return
	}
	gs.VictoryAchieved = true
	gs.VictoryAchievedTurn = gs.Turn
	gs.WinnerID = gs.PlayerFactionID
}
