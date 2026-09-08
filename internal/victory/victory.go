package victory

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// GoldEconomyPreview mevcut state'ten bir sonraki ekonomi tick'inin altın
// özetini yan etkisiz olarak hesaplar. Tur içinde değişen vergi, haraç,
// abluka, yağma ve ticaret anlaşmaları HUD/panellerde hemen görünmelidir;
// GoldEconomy map'i ise yalnızca son tamamlanan tick'in sonucudur.
func GoldEconomyPreview(gs *state.GameState, fid faction.FactionID) state.GoldEconomyStatus {
	if gs == nil || fid == "" || gs.Factions[fid] == nil {
		return state.GoldEconomyStatus{}
	}

	status := state.GoldEconomyStatus{FactionID: fid}
	season := gs.CurrentSeason()
	seasonTrade := season.TradeMod()
	harvest := season.HarvestMod()
	var effects tech.Effects
	if gs.TechTypes != nil {
		effects = tech.ComputeEffects(gs.Factions[fid].Research.Completed, gs.TechTypes)
	}
	baseIncomeByFaction := make(map[faction.FactionID]int)
	baseIncomeByFaction[fid] = 0

	for _, region := range gs.Regions {
		if region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID == "" || gs.SiegeAt(region.ID) != nil {
			continue
		}
		owner := faction.FactionID(region.OwnerID)
		retention := gs.RegionBlockadeOutputRetentionPercent(region)
		goldMod := 1.0
		for _, buildingID := range region.Buildings {
			if building := gs.BuildingTypes[buildingID]; building != nil {
				goldMod *= building.GoldMod
			}
		}
		tax := state.ScaleBlockadeOutputForEconomy(int(float64(region.GoldIncome())*goldMod*float64(harvest)/100), retention)
		trade := gs.BaseRegionTradeIncome(region) * seasonTrade / 100
		trade = state.ScaleBlockadeOutputForEconomy(trade, retention)
		tradeBase := (gs.BaseRegionTradeIncome(region) - gs.RegionTradeCenterIncome(region)) * seasonTrade / 100
		tradeBase = state.ScaleBlockadeOutputForEconomy(tradeBase, retention)
		if ownerEffects, ok := gs.Factions[owner]; ok && ownerEffects != nil && gs.TechTypes != nil {
			marketMod := tech.ComputeEffects(ownerEffects.Research.Completed, gs.TechTypes).MarketGoldMod
			trade = int(float64(trade) * (1 + marketMod))
			tradeBase = int(float64(tradeBase) * (1 + marketMod))
		}
		center := trade - tradeBase
		capital := gs.CapitalRegionBonus(region).Gold
		regionTotal := tax + trade + capital
		if raid := gs.Raids[region.ID]; raid != nil && raid.Turn == gs.Turn && raid.RaiderFactionID != owner {
			loot := gs.RaidLootPreview(region).Gold
			regionTotal -= loot
			if raid.RaiderFactionID == fid {
				status.RaidIncome += loot
			}
		}
		baseIncomeByFaction[owner] += regionTotal
		if owner != fid {
			continue
		}
		status.TaxIncome += tax
		status.TradeIncome += tradeBase
		status.TradeCenterIncome += center
		status.CapitalIncome += capital
	}

	blockadeLoot := gs.BlockadeLootForFaction(fid).Gold
	status.BlockadeIncome = blockadeLoot
	status.TechnologyIncome = effects.GoldPerRegion * len(gs.RegionsOwnedBy(fid))
	status.TradePowerIncome = gs.TradePowerCommerceIncome(fid)
	supplyPercent := 100
	if grainStatus, ok := gs.GrainEconomy[fid]; ok {
		supplyPercent = goldIncomeSupplyPercent(grainStatus.SupplyLevel)
	}
	status.TaxIncome = status.TaxIncome * supplyPercent / 100
	status.TradeIncome = status.TradeIncome * supplyPercent / 100
	status.TradeCenterIncome = status.TradeCenterIncome * supplyPercent / 100
	status.CapitalIncome = status.CapitalIncome * supplyPercent / 100
	status.TechnologyIncome = status.TechnologyIncome * supplyPercent / 100
	status.BlockadeIncome = status.BlockadeIncome * supplyPercent / 100
	// Yağma transferi arz cezasından etkilenmez.
	status.Income = status.TaxIncome + status.TradeIncome + status.TradeCenterIncome + status.CapitalIncome + status.TechnologyIncome + status.BlockadeIncome + status.RaidIncome

	// Aktif rotaların beklenen başarılı transferlerini toplar. Bu okuma,
	// anlaşma kurulduğu anda rotanın gelecek tur katkısını görünür kılar.
	type routeResource struct {
		fid  faction.FactionID
		good economy.GoodType
	}
	availableGoods := make(map[routeResource]int)
	availableGold := make(map[faction.FactionID]int, len(gs.Factions))
	for ownerID, owner := range gs.Factions {
		if owner == nil {
			continue
		}
		availableGold[ownerID] = owner.Gold
		for _, good := range []economy.GoodType{economy.GoodGrain, economy.GoodIron, economy.GoodTimber, economy.GoodStone, economy.GoodSpice, economy.GoodCloth} {
			if kind, ok := economy.GoodToResourceKind(good); ok {
				availableGoods[routeResource{fid: ownerID, good: good}] = economy.FactionResourceAmount(owner, kind)
			}
		}
	}
	for _, route := range gs.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.EffectiveAmountPerTurn() <= 0 {
			continue
		}
		fromID := faction.FactionID(route.FromFactionID)
		toID := faction.FactionID(route.ToFactionID)
		amountPerTurn := route.EffectiveAmountPerTurn()
		cost := route.GoldEarned()
		resourceKey := routeResource{fid: fromID, good: route.Good}
		if availableGoods[resourceKey] < amountPerTurn || availableGold[toID] < cost {
			continue
		}
		availableGoods[resourceKey] -= amountPerTurn
		availableGold[fromID] += cost
		availableGold[toID] -= cost
		amount := cost
		if route.FromFactionID == string(fid) {
			status.TradeRouteIncome += amount
		}
		if route.ToFactionID == string(fid) {
			status.TradeRouteExpense += amount
			customs := amount * (economy.TradeRouteCustomsRatePercent + gs.TradePowerSharePercent(fid)/10) / 100
			status.TradeRouteCustomsIncome += customs
			availableGold[toID] += customs
		}
	}

	if f := gs.Factions[fid]; f != nil && f.OverlordID != "" {
		status.TributePaid = tributePreview(gs, fid, tributeBaseIncome(gs, fid, baseIncomeByFaction[fid]))
	}
	for childID, child := range gs.Factions {
		if child != nil && child.OverlordID == fid {
			status.TributeIncome += tributePreview(gs, childID, tributeBaseIncome(gs, childID, baseIncomeByFaction[childID]))
		}
	}
	status.Upkeep = gs.FactionGoldUpkeep(fid)
	status.BuildingUpkeep = gs.FactionBuildingGoldUpkeep(fid)
	status.NetChange = status.Income + status.TradeRouteIncome - status.TradeRouteExpense + status.TradeRouteCustomsIncome + status.TributeIncome - status.TributePaid - status.Upkeep - status.BuildingUpkeep
	return status
}

func goldIncomeSupplyPercent(level state.GrainSupplyLevel) int {
	switch level {
	case state.GrainSupplyFamine:
		return 75
	case state.GrainSupplyCritical:
		return 90
	case state.GrainSupplyWarning:
		return 95
	default:
		return 100
	}
}

func tributeBaseIncome(gs *state.GameState, fid faction.FactionID, base int) int {
	if gs == nil {
		return base
	}
	if owner := gs.Factions[fid]; owner != nil && gs.TechTypes != nil {
		base += tech.ComputeEffects(owner.Research.Completed, gs.TechTypes).GoldPerRegion * len(gs.RegionsOwnedBy(fid))
	}
	return base + gs.BlockadeLootForFaction(fid).Gold
}

func tributePreview(gs *state.GameState, fid faction.FactionID, income int) int {
	f := gs.Factions[fid]
	if f == nil || f.OverlordID == "" || income <= 0 {
		return 0
	}
	rate := f.TributeRate
	if !f.TributeRateConfigured {
		rate = diplomacy.VassalTributeRatePercent()
	}
	tribute := income * diplomacy.ClampVassalTributeRate(rate) / 100
	if tribute > f.Gold {
		return f.Gold
	}
	return tribute
}

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
	return GoldIncomeForFaction(gs, gs.PlayerFactionID)
}

// GoldIncomeForFaction seçilen devletin mevcut tur başı brüt altın gelirini
// hesaplar. Oyuncu HUD'u ve devlet bilgi paneli aynı gelir hesabını kullanır.
func GoldIncomeForFaction(gs *state.GameState, fid faction.FactionID) int {
	if gs == nil || fid == "" || gs.Factions[fid] == nil {
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
		if region == nil || region.IsSea || region.IsTerrainArea || region.OwnerID != string(fid) {
			continue
		}
		goldMod := 1.0
		for _, bid := range region.Buildings {
			if building, ok := gs.BuildingTypes[bid]; ok {
				goldMod *= building.GoldMod
			}
		}
		retention := gs.RegionBlockadeOutputRetentionPercent(region)
		income += state.ScaleBlockadeOutputForEconomy(int(float64(region.GoldIncome())*goldMod*float64(harvestMod)/100), retention)
		tradeIncome := gs.BaseRegionTradeIncome(region)
		tradeIncome = tradeIncome * seasonMod / 100
		tradeIncome = state.ScaleBlockadeOutputForEconomy(tradeIncome, retention)
		if fx.MarketGoldMod != 0 {
			tradeIncome = int(float64(tradeIncome) * (1.0 + fx.MarketGoldMod))
		}
		income += tradeIncome
	}
	loot := gs.BlockadeLootForFaction(fid)
	income += loot.Gold

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
