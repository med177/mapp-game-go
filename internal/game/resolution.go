package game

import (
	"fmt"
	"sort"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/satisfaction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const friendlyReplenishHP = 10
const (
	embarkedVoyageGraceTurns = 3
	embarkedVoyageBaseDamage = 4
	embarkedVoyageStepDamage = 2
	embarkedVoyageMaxDamage  = 12
)

// civilianGrainDemand bir bölgenin tur başı sivil tahıl ihtiyacını hesaplar.
// Nüfusu olmayan legacy/test bölgeleri ekonomi havuzundan tüketim yapmaz.
func civilianGrainDemand(region *world.Region) int {
	return state.CivilianGrainDemand(region)
}

const (
	grainWarningMonths          = 3
	grainCriticalMonths         = 1
	grainCivilianStorageMonths  = 6
	grainArmyStorageMonths      = 3
	grainMinimumStorageCapacity = 100
)

const grainSpoilagePercent = 2

const (
	grainPopulationGrowthMonth = 11
	grainPerPopulationGrowth   = 2
)

func grainStorageCapacity(civilianDemand, armyUpkeep, storageBonus int) int {
	return state.GrainStorageCapacity(civilianDemand, armyUpkeep, storageBonus)
}

func grainSpoilage(stockpile, capacity int) int {
	if stockpile <= capacity || capacity <= 0 {
		return 0
	}
	excess := stockpile - capacity
	spoiled := excess * grainSpoilagePercent / 100
	if spoiled < 1 {
		spoiled = 1
	}
	return spoiled
}

// grainEconomyStatus mevcut stok ve tur talebinden oyuncuya/AI'ye gösterilecek
// tahıl rezerv durumunu üretir. Negatif stok, tüketim sonrası kıtlık miktarıdır;
// gerçek faction stoku ekonomi tick'inde sıfıra clamp edilir.
func grainEconomyStatus(fid faction.FactionID, stockpile, production, civilianDemand, armyUpkeep, storageBonus int) state.GrainEconomyStatus {
	if stockpile < 0 {
		stockpile = 0
	}
	if production < 0 {
		production = 0
	}
	if civilianDemand < 0 {
		civilianDemand = 0
	}
	if armyUpkeep < 0 {
		armyUpkeep = 0
	}

	totalDemand := civilianDemand + armyUpkeep
	rawBalance := stockpile + production - totalDemand
	status := state.GrainEconomyStatus{
		FactionID:       fid,
		Production:      production,
		CivilianDemand:  civilianDemand,
		ArmyUpkeep:      armyUpkeep,
		TotalDemand:     totalDemand,
		NetChange:       production - totalDemand,
		Stockpile:       rawBalance,
		StorageCapacity: grainStorageCapacity(civilianDemand, armyUpkeep, storageBonus),
		MonthsOfSupply:  -1,
	}
	if rawBalance < 0 {
		status.Shortage = -rawBalance
		status.Stockpile = 0
	}
	if totalDemand > 0 {
		status.MonthsOfSupply = status.Stockpile / totalDemand
		switch {
		case status.Shortage > 0:
			status.SupplyLevel = state.GrainSupplyFamine
		case status.MonthsOfSupply < grainCriticalMonths:
			status.SupplyLevel = state.GrainSupplyCritical
		case status.MonthsOfSupply < grainWarningMonths:
			status.SupplyLevel = state.GrainSupplyWarning
		}
	}
	return status
}

func grainGoldIncomePercent(level state.GrainSupplyLevel) int {
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

// factionWarFatigueByID her faction için savaş yorgunluğu cezasını hesaplar.
// Aynı realm içindeki overlord ve vassallar tek bağımsız devlet sayılır;
// savaşan karşı realm sayısı başına ceza uygulanır.
func factionWarFatigueByID(gs *state.GameState) map[string]int {
	fatigueByFaction := make(map[string]int)
	if gs == nil {
		return fatigueByFaction
	}
	for fid := range gs.Factions {
		fatigueByFaction[string(fid)] = diplomacy.IndependentWarSatisfactionPenalty(gs, fid)
	}
	// Legacy/test state'lerinde ilişki tarafı Factions map'inde bulunmayabilir;
	// bu durumda bölge sahibi ID'si için de sonucu erişilebilir tut.
	for _, relation := range gs.Relations {
		if relation == nil || relation.Stance != faction.StanceWar {
			continue
		}
		for _, fid := range []faction.FactionID{relation.FactionA, relation.FactionB} {
			if fid == "" {
				continue
			}
			fatigueByFaction[string(fid)] = diplomacy.IndependentWarSatisfactionPenalty(gs, fid)
		}
	}
	return fatigueByFaction
}

// regionArmySatisfactionBonus bölgedeki sahibine ait kara ordularının toplam
// gücünü istikrar bonusuna çevirir. 100 güç +10 verir; bonus +10 ile sınırlıdır.
func regionArmySatisfactionBonus(gs *state.GameState, region *world.Region) int {
	return satisfaction.ArmyStabilityBonus(gs, region)
}

func grainArmyMoraleDelta(level state.GrainSupplyLevel) int {
	switch level {
	case state.GrainSupplyFamine:
		return -6
	case state.GrainSupplyCritical:
		return -3
	case state.GrainSupplyWarning:
		return -1
	default:
		return 1
	}
}

// applyGrainArmyMorale, sivil memnuniyeti ve ordu HP kaybından ayrı olarak,
// tahıl arzını ordunun kalıcı moral state'ine bağlar. ID sırası replay
// determinism'i korur; aynı tick'te her ordu yalnız bir kez etkilenir.
func applyGrainArmyMorale(gs *state.GameState, ownerID string, level state.GrainSupplyLevel) int {
	if gs == nil || ownerID == "" {
		return 0
	}
	delta := grainArmyMoraleDelta(level)
	if delta == 0 {
		return 0
	}
	armyIDs := gs.ArmyOrder
	if len(armyIDs) != len(gs.Armies) {
		armyIDs = make([]army.ArmyID, 0, len(gs.Armies))
		for aid := range gs.Armies {
			armyIDs = append(armyIDs, aid)
		}
		sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
	}

	changed := 0
	for _, aid := range armyIDs {
		a := gs.Armies[aid]
		if a == nil || a.OwnerID != ownerID || len(a.Units) == 0 {
			continue
		}
		changed += a.ApplyMoraleDelta(delta)
	}
	return changed
}

type economyTickReport struct {
	PlayerLogisticsAlerts []state.RegionLogisticsStatus
	PlayerGrainStatus     state.GrainEconomyStatus
	PlayerGoldStatus      state.GoldEconomyStatus
}

type navalVoyageAlert struct {
	FleetID          army.ArmyID
	RegionID         world.RegionID
	TurnsWithoutPort int
	DamagePerUnit    int
	UnitsAffected    int
	UnitsLost        int
	TotalHPDamage    int
}

type eliminationResult struct {
	FactionID         faction.FactionID
	SuccessorID       faction.FactionID
	TransferredArmies int
	TransferredFleets int
}

func eliminateFaction(gs *state.GameState, fid, successor faction.FactionID) eliminationResult {
	if gs == nil || fid == "" {
		return eliminationResult{}
	}
	f := gs.Factions[fid]
	if f == nil || f.IsEliminated {
		return eliminationResult{}
	}

	result := eliminationResult{FactionID: fid, SuccessorID: successor}
	f.IsEliminated = true
	if gs.AIPlans != nil {
		delete(gs.AIPlans, fid)
	}
	if successor != "" && successor != fid {
		for otherID, other := range gs.Factions {
			if other == nil || otherID == fid {
				continue
			}
			if other.OverlordID == fid {
				other.OverlordID = successor
			}
		}
	} else {
		for otherID, other := range gs.Factions {
			if other == nil || otherID == fid {
				continue
			}
			if other.OverlordID == fid {
				other.OverlordID = ""
			}
		}
	}
	f.OverlordID = ""
	f.CapitalSettlementID = ""
	f.PendingCapitalSettlementID = ""
	f.PendingCapitalTurns = 0

	transferOwnership := successor != "" && successor != fid && gs.Factions[successor] != nil
	for aid, a := range gs.Armies {
		if a == nil || a.OwnerID != string(fid) {
			continue
		}
		if transferOwnership {
			gs.TransferArmyOwnership(a, string(successor))
			if a.IsNaval {
				result.TransferredFleets++
			} else {
				result.TransferredArmies++
			}
			continue
		}
		gs.RemoveArmy(aid)
	}

	for key, rel := range gs.Relations {
		if rel == nil || rel.FactionA == fid || rel.FactionB == fid {
			delete(gs.Relations, key)
		}
	}

	if len(gs.DiplomaticOffers) > 0 {
		offers := gs.DiplomaticOffers[:0]
		for _, offer := range gs.DiplomaticOffers {
			if offer.FromFactionID == fid || offer.ToFactionID == fid {
				continue
			}
			offers = append(offers, offer)
		}
		gs.DiplomaticOffers = offers
	}

	if len(gs.TradeRoutes) > 0 {
		routes := gs.TradeRoutes[:0]
		fidStr := string(fid)
		for _, route := range gs.TradeRoutes {
			if route == nil || route.FromFactionID == fidStr || route.ToFactionID == fidStr {
				continue
			}
			routes = append(routes, route)
		}
		gs.TradeRoutes = routes
	}

	return result
}

// techModsFor bir fraksiyonun araştırdığı teknolojilerden savaş modlarını hesaplar.
func techModsFor(gs *state.GameState, ownerID string) combat.TechMods {
	return combat.TechModsFor(gs, ownerID)
}

// checkRegionUnlocks kilidi kalkan bölgeleri açar.
// UnlockTurn atanmış bölgeler sadece zamanla açılır; UnlockTurn=0 ise keşif tipi
// kilit sayılır ve komşuya ulaşan ordu ile açılabilir.
func checkRegionUnlocks(gs *state.GameState) []world.RegionID {
	unlocked := gs.SyncTimedRegionUnlocks()
	for _, r := range gs.Regions {
		if !r.IsLocked || r.UnlockTurn > 0 {
			continue
		}
		// Komşuya ulaşan ordu kilidi açar
		for _, a := range gs.Armies {
			src, ok := gs.Regions[a.RegionID]
			if !ok {
				continue
			}
			for _, nid := range src.Neighbors {
				if nid == r.ID {
					r.IsLocked = false
					unlocked = append(unlocked, r.ID)
					break
				}
			}
			if !r.IsLocked {
				break
			}
		}
	}
	return unlocked
}

// applyTechTicks tüm fraksiyonların aktif araştırmalarını bir tur ilerletir.
// Tamamlanan teknolojileri (fraksiyonID, techID) çiftleri olarak döner.
func applyTechTicks(gs *state.GameState) []struct {
	factionID string
	techID    string
} {
	var completed []struct {
		factionID string
		techID    string
	}
	for fid, f := range gs.Factions {
		if f.IsEliminated {
			continue
		}
		if completedID := tech.Tick(&f.Research); completedID != "" {
			completed = append(completed, struct {
				factionID string
				techID    string
			}{string(fid), completedID})
		}
	}
	return completed
}

// applySeasonEffects mevsim etkilerini tüm ordulara uygular.
func applySeasonEffects(gs *state.GameState) {
	gs.ArmyMoveUsage = make(map[army.ArmyID]bool, len(gs.Armies))
	s := gs.CurrentSeason()

	if s.IsWinter() {
		for _, a := range gs.Armies {
			applyWinterAttrition(gs, a)
		}
	}

	for _, a := range gs.Armies {
		if a.MaxMovePoints > 0 {
			gs.ArmyMoveUsage[a.ID] = a.MovePoints >= 0 && a.MovePoints < a.MaxMovePoints
		}
		if !s.IsWinter() && a.IsNaval {
			replenishDockedFleet(gs, a, friendlyReplenishHP)
		}
		// Mevsim etkisi en yavaş birimin tabanına uygulanır; diğer hareket
		// bonusları state katmanındaki ortak hesapta eklenir.
		mp := gs.ArmyMaxMovePoints(a)
		a.MaxMovePoints = mp
		a.ResetMovePoints()
	}
}

func applyWinterAttrition(gs *state.GameState, a *army.Army) (lost int) {
	if a == nil {
		return 0
	}
	if a.IsNaval && gs.CanFleetAvoidSeaAttrition(a) {
		return 0
	}
	route := merchantRouteForFleet(gs, a)
	merchantRouteActive := a.IsNaval && route != nil && gs.MerchantFleetTradeRouteBonus(a, route) > 0
	if !merchantRouteActive {
		return a.ApplyWinterAttrition()
	}

	surviving := a.Units[:0]
	for _, u := range a.Units {
		if isMerchantTradeUnit(gs, u) {
			surviving = append(surviving, u)
			continue
		}
		u.CurrentHP = u.CurrentHP * 90 / 100
		if u.CurrentHP <= 0 {
			lost++
			continue
		}
		surviving = append(surviving, u)
	}
	a.Units = surviving
	return lost
}

func merchantRouteForFleet(gs *state.GameState, fleet *army.Army) *economy.TradeRoute {
	if gs == nil || fleet == nil || fleet.TradeRouteKey == "" {
		return nil
	}
	for _, route := range gs.TradeRoutes {
		if route != nil && route.AssignmentKey() == fleet.TradeRouteKey {
			return route
		}
	}
	return nil
}

func isMerchantTradeUnit(gs *state.GameState, u army.Unit) bool {
	if u.TypeID == "merchant_ship" {
		return true
	}
	unitType := gs.UnitTypes[u.TypeID]
	return unitType != nil && unitType.Category == army.CategoryNavalTrade
}

func replenishDockedFleet(gs *state.GameState, fleet *army.Army, amount int) int {
	if gs == nil || fleet == nil || !fleet.IsNaval || amount <= 0 || fleet.DockedRegionID == "" {
		return 0
	}
	dockedRegion := gs.Regions[fleet.DockedRegionID]
	if dockedRegion == nil || dockedRegion.IsSea || dockedRegion.OwnerID == "" || !dockedRegion.HasPort() || !gs.CanArmyReplenishIn(fleet) {
		return 0
	}
	healAmount := amount
	if dockedRegion.OwnerID != fleet.OwnerID {
		healAmount = amount / 2
		if healAmount < 1 {
			healAmount = 1
		}
	}
	healedUnits := 0
	for i := range fleet.Units {
		if fleet.Units[i].CurrentHP >= army.MaxUnitHP {
			continue
		}
		fleet.Units[i].CurrentHP += healAmount
		if fleet.Units[i].CurrentHP > army.MaxUnitHP {
			fleet.Units[i].CurrentHP = army.MaxUnitHP
		}
		healedUnits++
	}
	for i := range fleet.EmbarkedUnits {
		if fleet.EmbarkedUnits[i].CurrentHP >= army.MaxUnitHP {
			continue
		}
		fleet.EmbarkedUnits[i].CurrentHP += healAmount
		if fleet.EmbarkedUnits[i].CurrentHP > army.MaxUnitHP {
			fleet.EmbarkedUnits[i].CurrentHP = army.MaxUnitHP
		}
		healedUnits++
	}
	return healedUnits
}

// applyEconomyTick tur başında her fraksiyonun ekonomisini günceller.
// Artık ticaret rotalarını işletir, mal transferi yapar ve piyasa fiyatlarını günceller.
func applyEconomyTick(gs *state.GameState) economyTickReport {
	report := economyTickReport{}
	s := gs.CurrentSeason()
	harvestMod := s.HarvestMod()
	effectsByFaction := make(map[string]tech.Effects, len(gs.Factions))
	if gs.TechTypes != nil {
		for fid, f := range gs.Factions {
			effectsByFaction[string(fid)] = tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
		}
	}

	incomeByFaction := make(map[string]int)
	taxIncomeByFaction := make(map[string]int)
	tradeIncomeByFaction := make(map[string]int)
	capitalIncomeByFaction := make(map[string]int)
	grainByFaction := make(map[string]int)
	civilianGrainDemandByFaction := make(map[string]int)
	ironByFaction := make(map[string]int)
	timberByFaction := make(map[string]int)
	stoneByFaction := make(map[string]int)
	spiceByFaction := make(map[string]int)
	clothByFaction := make(map[string]int)
	raidLootByFaction := make(map[string]state.RegionProductionSummary)
	storageCapacityByFaction := make(map[string]int)
	gs.GrainEconomy = make(map[faction.FactionID]state.GrainEconomyStatus, len(gs.Factions))
	gs.GoldEconomy = make(map[faction.FactionID]state.GoldEconomyStatus, len(gs.Factions))

	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID == "" {
			continue
		}
		if gs.SiegeAt(r.ID) != nil {
			// Kuşatma altındaki bölge devlete gelir/hammadde sağlamaz
			// ve halkın memnuniyeti kuşatma stresinden her tur azalır
			continue
		}

		// Bina çarpanları
		goldMod := 1.0
		grainMod := 1.0
		for _, bid := range r.Buildings {
			if b, ok := gs.BuildingTypes[bid]; ok {
				goldMod *= b.GoldMod
				grainMod *= b.GrainMod
				storageCapacityByFaction[r.OwnerID] += b.StorageCapacity
			}
		}

		blockadeRetention := gs.RegionBlockadeOutputRetentionPercent(r)
		income := state.ScaleBlockadeOutputForEconomy(int(float64(r.GoldIncome())*goldMod*float64(harvestMod)/100), blockadeRetention)
		grain := int(float64(r.BaseGrainOutput) * grainMod)
		iron := r.BaseIronOutput
		timber := r.BaseTimberOutput
		stone := r.BaseStoneOutput
		spice := r.BaseSpiceOutput
		cloth := r.BaseClothOutput
		grain, iron, timber, stone, spice, cloth = applyTerrainSpecialization(r.Terrain, grain, iron, timber, stone, spice, cloth)
		grain = state.ScaleBlockadeOutputForEconomy(grain, blockadeRetention)
		iron = state.ScaleBlockadeOutputForEconomy(iron, blockadeRetention)
		timber = state.ScaleBlockadeOutputForEconomy(timber, blockadeRetention)
		stone = state.ScaleBlockadeOutputForEconomy(stone, blockadeRetention)
		spice = state.ScaleBlockadeOutputForEconomy(spice, blockadeRetention)
		cloth = state.ScaleBlockadeOutputForEconomy(cloth, blockadeRetention)

		// Pasif ticaret geliri ortak efektif kapasite üzerinden hesaplanır.
		tradeIncome := gs.BaseRegionTradeIncome(r)
		// Mevsimsel ticaret modu uygula
		tradeIncome = tradeIncome * s.TradeMod() / 100
		tradeIncome = state.ScaleBlockadeOutputForEconomy(tradeIncome, blockadeRetention)
		if fx, ok := effectsByFaction[r.OwnerID]; ok && fx.MarketGoldMod != 0 {
			tradeIncome = int(float64(tradeIncome) * (1.0 + fx.MarketGoldMod))
		}

		incomeByFaction[r.OwnerID] += income + tradeIncome
		taxIncomeByFaction[r.OwnerID] += income
		tradeIncomeByFaction[r.OwnerID] += tradeIncome
		civilianGrainDemandByFaction[r.OwnerID] += gs.CivilianGrainDemandForRegion(r)
		ironByFaction[r.OwnerID] += iron
		timberByFaction[r.OwnerID] += timber
		stoneByFaction[r.OwnerID] += stone
		spiceByFaction[r.OwnerID] += spice
		clothByFaction[r.OwnerID] += cloth
		capitalGrainBonus := 0
		if bonus := gs.CapitalRegionBonus(r); bonus != (state.RegionProductionSummary{}) {
			incomeByFaction[r.OwnerID] += bonus.Gold
			capitalIncomeByFaction[r.OwnerID] += bonus.Gold
			capitalGrainBonus = bonus.Grain
			ironByFaction[r.OwnerID] += bonus.Iron
			timberByFaction[r.OwnerID] += bonus.Timber
			stoneByFaction[r.OwnerID] += bonus.Stone
			spiceByFaction[r.OwnerID] += bonus.Spice
			clothByFaction[r.OwnerID] += bonus.Cloth
		}
		productionPercent := 100 + gs.RegionGrainProductionModifier(r.ID)
		if productionPercent < 0 {
			productionPercent = 0
		}
		grain = (grain + capitalGrainBonus) * productionPercent / 100
		if raid := gs.Raids[r.ID]; raid != nil && raid.Turn == gs.Turn && raid.RaiderFactionID != faction.FactionID(r.OwnerID) {
			loot := gs.RaidLootPreview(r)
			income -= loot.Gold
			grain -= loot.Grain
			iron -= loot.Iron
			timber -= loot.Timber
			stone -= loot.Stone
			spice -= loot.Spice
			cloth -= loot.Cloth
			incomeByFaction[r.OwnerID] -= loot.Gold
			ironByFaction[r.OwnerID] -= loot.Iron
			timberByFaction[r.OwnerID] -= loot.Timber
			stoneByFaction[r.OwnerID] -= loot.Stone
			spiceByFaction[r.OwnerID] -= loot.Spice
			clothByFaction[r.OwnerID] -= loot.Cloth
			raidLootByFaction[string(raid.RaiderFactionID)] = addProductionSummary(raidLootByFaction[string(raid.RaiderFactionID)], loot)
		}
		grainByFaction[r.OwnerID] += grain

	}

	// --- Ticaret rotalarını işlet (mal + altın transferi) ---
	goldBeforeEconomy := make(map[string]int, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f != nil {
			goldBeforeEconomy[string(fid)] = f.Gold
		}
	}
	diplomacy.RebalanceTradeRouteCapacities(gs)
	gs.RefreshTradeRouteBlockades()
	gs.RefreshMerchantTradeBonuses()
	tradeLogs, tradeTransfers := economy.ApplyTradeRoutesWithTransfers(gs.Factions, gs.TradeRoutes)
	tradeRouteIncomeByFaction := make(map[string]int)
	tradeRouteExpenseByFaction := make(map[string]int)
	for _, transfer := range tradeTransfers {
		tradeRouteIncomeByFaction[string(transfer.FromFactionID)] += transfer.Amount
		tradeRouteExpenseByFaction[string(transfer.ToFactionID)] += transfer.Amount
	}
	for _, log := range tradeLogs {
		// Ticaret logları oyuncuya aitse göster
		if gs.PlayerFactionID != "" {
			_ = log // ileride oyuncuya bildirim gösterilebilir
		}
	}

	// Gerçek ordu bakım maliyetleri (UnitType.GrainUpkeep/GoldUpkeep)
	upkeepByFaction := make(map[string]int)
	goldUpkeepByFaction := make(map[string]int)
	for _, a := range gs.Armies {
		upkeepByFaction[a.OwnerID] += gs.EffectiveArmyGrainUpkeep(a)
		goldUpkeepByFaction[a.OwnerID] += gs.EffectiveArmyGoldUpkeep(a)
	}

	for fid, f := range gs.Factions {
		fidStr := string(fid)
		loot := gs.BlockadeLootForFaction(fid)

		// Teknoloji bonusları
		fx := effectsByFaction[fidStr]

		// GoldPerRegion tech bonusu
		ownedCount := len(gs.RegionsOwnedBy(fid))
		techGold := fx.GoldPerRegion * ownedCount

		raidLoot := raidLootByFaction[fidStr]
		netGrain := int(float64(grainByFaction[fidStr])*(1.0+fx.GrainMod)) + loot.Grain + raidLoot.Grain
		civilianDemand := civilianGrainDemandByFaction[fidStr]
		status := grainEconomyStatus(fid, f.Grain, netGrain, civilianDemand, upkeepByFaction[fidStr], storageCapacityByFaction[fidStr])
		goldBefore := goldBeforeEconomy[fidStr]
		goldIncome := (incomeByFaction[fidStr] + techGold + loot.Gold) * grainGoldIncomePercent(status.SupplyLevel) / 100
		// Yağmalanan vergi transferi doğrudan yağmalayan devlete geçer; hedef
		// devletin tahıl arz cezasından etkilenmez.
		goldIncome += raidLoot.Gold
		f.Gold += goldIncome
		if f.Gold < 0 {
			f.Gold = 0
		}
		goldUpkeep := goldUpkeepByFaction[fidStr]
		paidGoldUpkeep := goldUpkeep
		if paidGoldUpkeep > f.Gold {
			paidGoldUpkeep = f.Gold
		}
		f.Gold -= paidGoldUpkeep
		goldStatus := state.GoldEconomyStatus{
			FactionID:         fid,
			Income:            goldIncome,
			TaxIncome:         taxIncomeByFaction[fidStr] * grainGoldIncomePercent(status.SupplyLevel) / 100,
			TradeIncome:       tradeIncomeByFaction[fidStr] * grainGoldIncomePercent(status.SupplyLevel) / 100,
			CapitalIncome:     capitalIncomeByFaction[fidStr] * grainGoldIncomePercent(status.SupplyLevel) / 100,
			TechnologyIncome:  techGold * grainGoldIncomePercent(status.SupplyLevel) / 100,
			BlockadeIncome:    loot.Gold * grainGoldIncomePercent(status.SupplyLevel) / 100,
			RaidIncome:        raidLoot.Gold,
			TradeRouteIncome:  tradeRouteIncomeByFaction[fidStr],
			TradeRouteExpense: tradeRouteExpenseByFaction[fidStr],
			Upkeep:            goldUpkeep,
			GoldBefore:        goldBefore,
			GoldAfter:         f.Gold,
			PaidUpkeep:        paidGoldUpkeep,
			Shortage:          goldUpkeep - paidGoldUpkeep,
		}
		if ledger := gs.GoldTurnLedger[fid]; ledger.Turn == gs.Turn {
			goldStatus.GiftIncome = ledger.GiftIncome
			goldStatus.GiftExpense = ledger.GiftExpense
		}
		goldStatus.NetChange = goldStatus.Income + goldStatus.TradeRouteIncome - goldStatus.TradeRouteExpense - goldUpkeep
		if goldStatus.Shortage > 0 {
			applyGoldUpkeepShortagePenalty(gs, fidStr, goldStatus.Upkeep, goldStatus.Shortage, &goldStatus)
		}
		gs.GoldEconomy[fid] = goldStatus
		f.Grain = status.Stockpile
		f.Iron += int(float64(ironByFaction[fidStr])*(1.0+fx.IronMod)) + loot.Iron + raidLoot.Iron
		f.Timber += int(float64(timberByFaction[fidStr])*(1.0+fx.TimberMod)) + loot.Timber + raidLoot.Timber
		f.Stone += int(float64(stoneByFaction[fidStr])*(1.0+fx.StoneMod)) + loot.Stone + raidLoot.Stone
		f.Spice += spiceByFaction[fidStr] + loot.Spice + raidLoot.Spice
		f.Cloth += clothByFaction[fidStr] + loot.Cloth + raidLoot.Cloth

		if status.Shortage > 0 {
			applyGrainShortagePenalty(gs, fidStr, status.Shortage)
		}
		status.ArmyMoraleDelta = applyGrainArmyMorale(gs, fidStr, status.SupplyLevel)
		status.Spoiled = grainSpoilage(f.Grain, status.StorageCapacity)
		if status.Spoiled > 0 {
			f.Grain -= status.Spoiled
		}
		status.Stockpile = f.Grain
		if status.TotalDemand > 0 {
			status.MonthsOfSupply = status.Stockpile / status.TotalDemand
		}
		status.StrategicDemand = state.StrategicGrainDemandFromStockpile(status.Stockpile, status.TotalDemand)
		gs.GrainEconomy[fid] = status
	}
	for regionID, raid := range gs.Raids {
		if raid == nil || raid.Turn == gs.Turn {
			delete(gs.Raids, regionID)
		}
	}

	// Vergi, bina, tahıl, teknoloji, savaş, genişleme ve ordu etkilerini tek
	// delta içinde birleştir. Nihai değer yalnızca burada 0-100 aralığına çekilir.
	satisfactionCalculator := satisfaction.NewCalculator(gs)
	for _, r := range gs.Regions {
		if r == nil || r.IsSea || r.OwnerID == "" {
			continue
		}
		breakdown := satisfactionCalculator.ForRegion(r)
		r.Satisfaction = clamp(r.Satisfaction+breakdown.Total, 0, 100)
	}
	report.PlayerLogisticsAlerts = applyRegionalLogisticsPressure(gs)
	applyGrainFundedArmyReplenishment(gs)
	if autoSold, autoGold := gs.ApplyAutomaticGrainExport(); autoSold > 0 {
		fid := gs.PlayerFactionID
		status := gs.GrainEconomy[fid]
		status.AutoExportSold = autoSold
		status.AutoExportGold = autoGold
		if player := gs.Factions[fid]; player != nil {
			status.Stockpile = player.Grain
			if status.TotalDemand > 0 {
				status.MonthsOfSupply = status.Stockpile / status.TotalDemand
			}
			status.StrategicDemand = state.StrategicGrainDemandFromStockpile(status.Stockpile, status.TotalDemand)
		}
		gs.GrainEconomy[fid] = status
	}
	applyGrainFundedPopulationGrowth(gs)

	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated || f.OverlordID == "" {
			continue
		}
		overlord := gs.Factions[f.OverlordID]
		if overlord == nil || overlord.IsEliminated {
			f.OverlordID = ""
			continue
		}
		income := incomeByFaction[string(fid)] + effectsByFaction[string(fid)].GoldPerRegion*len(gs.RegionsOwnedBy(fid)) + gs.BlockadeLootForFaction(fid).Gold
		if income <= 0 {
			continue
		}
		tribute := income * diplomacy.VassalTributeRatePercent() / 100
		if tribute <= 0 {
			continue
		}
		if tribute > f.Gold {
			tribute = f.Gold
		}
		f.Gold -= tribute
		overlord.Gold += tribute
		vassalStatus := gs.GoldEconomy[fid]
		vassalStatus.TributePaid += tribute
		vassalStatus.NetChange -= tribute
		gs.GoldEconomy[fid] = vassalStatus
		overlordStatus := gs.GoldEconomy[f.OverlordID]
		overlordStatus.TributeIncome += tribute
		overlordStatus.NetChange += tribute
		gs.GoldEconomy[f.OverlordID] = overlordStatus
	}
	for fid, status := range gs.GoldEconomy {
		if f := gs.Factions[fid]; f != nil {
			status.GoldAfter = f.Gold
			gs.GoldEconomy[fid] = status
		}
	}

	if gs.PlayerFactionID != "" {
		report.PlayerGrainStatus = gs.GrainEconomy[gs.PlayerFactionID]
		report.PlayerGoldStatus = gs.GoldEconomy[gs.PlayerFactionID]
	}
	gs.ArmyMoveUsage = nil

	// --- Dinamik piyasa fiyatlarını açık pazar arzı ve stratejik tahılla güncelle ---
	refreshMarketPrices(gs)
	return report
}

func addProductionSummary(a, b state.RegionProductionSummary) state.RegionProductionSummary {
	a.Gold += b.Gold
	a.Grain += b.Grain
	a.Iron += b.Iron
	a.Timber += b.Timber
	a.Stone += b.Stone
	a.Spice += b.Spice
	a.Cloth += b.Cloth
	return a
}

// applyGrainFundedArmyReplenishment, mevcut ücretsiz toparlanmaya ek olarak,
// yalnızca depo kapasitesini aşan tahılla dost kara ordularını yeniler. Ordu ve
// birim sırası deterministiktir; böylece sınırlı fazla tahıl her replay'de aynı
// ordulara gider. Bir HP toparlanması bir tahıl tüketir ve ordu başına/turuna
// en fazla o bölgenin çiftlik/ambar kaynaklı toparlanma hızı kadar ek iyileşme
// yapılır.
func applyGrainFundedArmyReplenishment(gs *state.GameState) {
	if gs == nil || gs.CurrentSeason().IsWinter() || len(gs.GrainEconomy) == 0 {
		return
	}

	factionIDs := make([]faction.FactionID, 0, len(gs.Factions))
	for fid := range gs.Factions {
		factionIDs = append(factionIDs, fid)
	}
	sort.Slice(factionIDs, func(i, j int) bool { return factionIDs[i] < factionIDs[j] })

	armyIDs := make([]army.ArmyID, 0, len(gs.Armies))
	for aid := range gs.Armies {
		armyIDs = append(armyIDs, aid)
	}
	sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })

	for _, fid := range factionIDs {
		f := gs.Factions[fid]
		status, ok := gs.GrainEconomy[fid]
		if f == nil || f.IsEliminated || !ok {
			continue
		}
		budget := f.Grain - status.StorageCapacity
		if budget <= 0 {
			continue
		}

		for _, aid := range armyIDs {
			if budget <= 0 {
				break
			}
			a := gs.Armies[aid]
			if a == nil || a.OwnerID != string(fid) || a.IsNaval ||
				!gs.CanArmyReplenishIn(a) || gs.IsArmyDefendingSiegedRegion(a) ||
				gs.ArmyReplenishmentHP(a) <= 0 {
				continue
			}
			regionSupply, supplied := gs.RegionLogistics[a.RegionID]
			if !supplied || regionSupply.Overload > 0 {
				continue
			}

			armyBudget := gs.ArmyReplenishmentHP(a)
			for i := range a.Units {
				if budget <= 0 || armyBudget <= 0 {
					break
				}
				missing := a.Units[i].MissingHP()
				if missing <= 0 {
					continue
				}
				heal := missing
				if heal > armyBudget {
					heal = armyBudget
				}
				if heal > budget {
					heal = budget
				}
				a.Units[i].CurrentHP += heal
				budget -= heal
				armyBudget -= heal
				status.ReplenishmentHP += heal
				status.ReplenishmentGrainSpent += heal
			}
		}

		if status.ReplenishmentGrainSpent <= 0 {
			continue
		}
		f.Grain -= status.ReplenishmentGrainSpent
		status.NetChange -= status.ReplenishmentGrainSpent
		status.Stockpile = f.Grain
		if status.TotalDemand > 0 {
			status.MonthsOfSupply = status.Stockpile / status.TotalDemand
		}
		status.StrategicDemand = state.StrategicGrainDemandFromStockpile(status.Stockpile, status.TotalDemand)
		gs.GrainEconomy[fid] = status
	}
}

// applyGrainFundedPopulationGrowth yılda bir kez Kasım ayını kapsayan turda, yalnızca depolama kapasitesini
// aşan ve stabil rezervden tahıl harcayarak nüfusu büyütür. Bölge sırası ve
// fraksiyon sırası deterministiktir; savaş rezervi kapasite tabanına kadar korunur.
func applyGrainFundedPopulationGrowth(gs *state.GameState) {
	if gs == nil || !gs.CurrentTurnIncludesMonth(grainPopulationGrowthMonth) || len(gs.GrainEconomy) == 0 {
		return
	}

	factionIDs := make([]faction.FactionID, 0, len(gs.Factions))
	for fid := range gs.Factions {
		factionIDs = append(factionIDs, fid)
	}
	sort.Slice(factionIDs, func(i, j int) bool { return factionIDs[i] < factionIDs[j] })

	for _, fid := range factionIDs {
		f := gs.Factions[fid]
		status := gs.GrainEconomy[fid]
		if f == nil || f.IsEliminated || status.SupplyLevel != state.GrainSupplyStable {
			continue
		}
		budget := f.Grain - status.StorageCapacity
		if budget < grainPerPopulationGrowth {
			continue
		}

		regions := make([]*world.Region, 0)
		for _, region := range gs.Regions {
			if region == nil || region.IsSea || region.OwnerID != string(fid) || region.Population <= 0 {
				continue
			}
			if region.Satisfaction < 60 || region.IsRebellionRisk() || gs.SiegeAt(region.ID) != nil {
				continue
			}
			regions = append(regions, region)
		}
		sort.Slice(regions, func(i, j int) bool { return regions[i].ID < regions[j].ID })

		for _, region := range regions {
			growth := region.Population / 100
			if growth < 1 {
				growth = 1
			}
			cost := growth * grainPerPopulationGrowth
			if cost > budget {
				break
			}
			region.AddPopulation(growth)
			f.Grain -= cost
			budget -= cost
			status.PopulationGrowth += growth
			status.GrowthGrainSpent += cost
			if budget < grainPerPopulationGrowth {
				break
			}
		}

		if status.GrowthGrainSpent > 0 {
			status.NetChange -= status.GrowthGrainSpent
			status.Stockpile = f.Grain
			if status.TotalDemand > 0 {
				status.MonthsOfSupply = status.Stockpile / status.TotalDemand
			}
			status.StrategicDemand = state.StrategicGrainDemandFromStockpile(status.Stockpile, status.TotalDemand)
			gs.GrainEconomy[fid] = status
		}
	}
}

func applyEmbarkedVoyageAttrition(gs *state.GameState) []navalVoyageAlert {
	if gs == nil {
		return nil
	}
	if gs.ArmyLogistics == nil {
		gs.ArmyLogistics = make(map[army.ArmyID]state.ArmyLogisticsStatus)
	}
	alerts := make([]navalVoyageAlert, 0)
	for aid, a := range gs.Armies {
		if a == nil || !a.IsNaval {
			continue
		}
		if a.DockedRegionID != "" || gs.CanFleetAvoidSeaAttrition(a) || len(a.EmbarkedUnits) == 0 {
			a.TurnsWithoutPort = 0
			continue
		}
		a.TurnsWithoutPort++
		if a.TurnsWithoutPort <= embarkedVoyageGraceTurns {
			continue
		}

		damagePerUnit := embarkedVoyageBaseDamage + (a.TurnsWithoutPort-embarkedVoyageGraceTurns-1)*embarkedVoyageStepDamage
		if damagePerUnit > embarkedVoyageMaxDamage {
			damagePerUnit = embarkedVoyageMaxDamage
		}
		unitsBefore := len(a.EmbarkedUnits)
		unitsLost := 0
		totalDamage := 0
		survivors := a.EmbarkedUnits[:0]
		for _, u := range a.EmbarkedUnits {
			u.CurrentHP -= damagePerUnit
			totalDamage += damagePerUnit
			if u.CurrentHP <= 0 {
				unitsLost++
				continue
			}
			survivors = append(survivors, u)
		}
		a.EmbarkedUnits = survivors
		gs.ArmyLogistics[aid] = state.ArmyLogisticsStatus{
			ArmyID:            aid,
			RegionID:          a.RegionID,
			OwnerID:           a.OwnerID,
			OverCapacityTurns: a.TurnsWithoutPort,
			DamagePerUnit:     damagePerUnit,
			UnitsAffected:     unitsBefore,
			UnitsLost:         unitsLost,
			TotalHPDamage:     totalDamage,
		}
		alerts = append(alerts, navalVoyageAlert{
			FleetID:          aid,
			RegionID:         a.RegionID,
			TurnsWithoutPort: a.TurnsWithoutPort,
			DamagePerUnit:    damagePerUnit,
			UnitsAffected:    unitsBefore,
			UnitsLost:        unitsLost,
			TotalHPDamage:    totalDamage,
		})
	}
	return alerts
}

func applyGrainShortagePenalty(gs *state.GameState, ownerID string, shortage int) {
	if shortage <= 0 {
		return
	}
	remaining := shortage
	armyIDs := gs.ArmyOrder
	if len(armyIDs) != len(gs.Armies) {
		armyIDs = make([]army.ArmyID, 0, len(gs.Armies))
		for aid := range gs.Armies {
			armyIDs = append(armyIDs, aid)
		}
		sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
	}
	for _, aid := range armyIDs {
		a := gs.Armies[aid]
		if a == nil {
			continue
		}
		if a.OwnerID != ownerID || len(a.Units) == 0 {
			continue
		}
		for i := range a.Units {
			if remaining <= 0 {
				return
			}
			// Tahıl açığında önce HP erir; lojistik krizi hissedilir.
			damage := 10
			if remaining > 6 {
				damage = 15
			}
			a.Units[i].CurrentHP -= damage
			if a.Units[i].CurrentHP < 0 {
				a.Units[i].CurrentHP = 0
			}
			remaining--
		}
	}
}

// applyGoldUpkeepShortagePenalty, maaş ödenemediğinde önce tüm ordularda
// yıpranma ve moral kaybı, ardından açığın şiddetine göre sınırlı asker kaçağı
// uygular. Ordu ID sırası replay determinism'i korur; isyan bu aşamanın
// kapsamı dışındadır.
func applyGoldUpkeepShortagePenalty(gs *state.GameState, ownerID string, upkeep, shortage int, status *state.GoldEconomyStatus) {
	if gs == nil || ownerID == "" || upkeep <= 0 || shortage <= 0 || status == nil {
		return
	}
	severity := shortage * 100 / upkeep
	if severity > 100 {
		severity = 100
	}
	damage := 5 + severity/20 // kısmi açıkta 5, tam açıkta 10 HP/tur
	moraleDelta := -5
	if severity >= 50 {
		moraleDelta = -10
	}
	if severity >= 100 {
		moraleDelta = -15
	}

	armyIDs := sortedArmyIDsForOwner(gs, ownerID)
	totalUnits := 0
	for _, aid := range armyIDs {
		a := gs.Armies[aid]
		if a == nil {
			continue
		}
		totalUnits += len(a.Units) + len(a.EmbarkedUnits)
		status.ArmyMoraleDelta += a.ApplyMoraleDelta(moraleDelta)
		lost, hpDamage := applyArmyFlatDamage(a, damage)
		status.UnitsLost += lost
		status.AttritionHPDamage += hpDamage
		if len(a.EmbarkedUnits) > 0 {
			remaining, embarkedLost, embarkedDamage := applyFlatDamageToUnits(a.EmbarkedUnits, damage)
			a.EmbarkedUnits = remaining
			status.UnitsLost += embarkedLost
			status.AttritionHPDamage += embarkedDamage
		}
	}

	// Tamamen ödenmeyen bir aylık maaş, gücün yaklaşık %10'unu kaçırabilir;
	// kısmi açıklar aynı turda yalnız HP/moral yıpranması oluşturur. En az dört
	// mevcut birimde tam açık bir kaçak garantilenir, tek kişilik ordular korunur.
	desertions := totalUnits * severity / 1000
	if severity >= 100 && totalUnits >= 4 && desertions == 0 {
		desertions = 1
	}
	maxDesertions := totalUnits - 1
	if desertions > maxDesertions {
		desertions = maxDesertions
	}
	if desertions <= 0 {
		return
	}
	for i := len(armyIDs) - 1; i >= 0 && desertions > 0; i-- {
		a := gs.Armies[armyIDs[i]]
		if a == nil {
			continue
		}
		removed := removeArmyUnitsForDesertion(a, desertions)
		desertions -= removed
		status.DesertedUnits += removed
		status.UnitsLost += removed
	}
}

func sortedArmyIDsForOwner(gs *state.GameState, ownerID string) []army.ArmyID {
	if gs == nil {
		return nil
	}
	ids := make([]army.ArmyID, 0, len(gs.Armies))
	for aid, a := range gs.Armies {
		if a != nil && a.OwnerID == ownerID && (len(a.Units) > 0 || len(a.EmbarkedUnits) > 0) {
			ids = append(ids, aid)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func applyFlatDamageToUnits(units []army.Unit, damage int) ([]army.Unit, int, int) {
	if damage <= 0 || len(units) == 0 {
		return units, 0, 0
	}
	survivors := units[:0]
	lost := 0
	totalDamage := 0
	for _, unit := range units {
		before := unit.CurrentHP
		unit.CurrentHP -= damage
		if unit.CurrentHP <= 0 {
			lost++
			totalDamage += before
			continue
		}
		totalDamage += before - unit.CurrentHP
		survivors = append(survivors, unit)
	}
	return survivors, lost, totalDamage
}

func removeArmyUnitsForDesertion(a *army.Army, count int) int {
	if a == nil || count <= 0 {
		return 0
	}
	removed := 0
	landCount := count
	if landCount > len(a.Units) {
		landCount = len(a.Units)
	}
	a.Units = a.Units[:len(a.Units)-landCount]
	removed += landCount
	remaining := count - removed
	if remaining > len(a.EmbarkedUnits) {
		remaining = len(a.EmbarkedUnits)
	}
	if remaining > 0 {
		a.EmbarkedUnits = a.EmbarkedUnits[:len(a.EmbarkedUnits)-remaining]
		removed += remaining
	}
	return removed
}

func applyRegionalLogisticsPressure(gs *state.GameState) []state.RegionLogisticsStatus {
	if gs == nil {
		return nil
	}
	gs.RegionLogistics = make(map[world.RegionID]state.RegionLogisticsStatus)
	gs.ArmyLogistics = make(map[army.ArmyID]state.ArmyLogisticsStatus)
	friendlySupplies := allocateFriendlyFrontlineSupply(gs)

	armiesByRegion := make(map[world.RegionID][]*army.Army)
	for _, a := range gs.Armies {
		if a == nil || a.IsNaval || len(a.Units) == 0 {
			continue
		}
		region := gs.Regions[a.RegionID]
		if region == nil || region.IsSea {
			continue
		}
		armiesByRegion[a.RegionID] = append(armiesByRegion[a.RegionID], a)
	}
	for _, armiesInRegion := range armiesByRegion {
		sort.Slice(armiesInRegion, func(i, j int) bool {
			return armiesInRegion[i].ID < armiesInRegion[j].ID
		})
	}
	regionIDs := make([]world.RegionID, 0, len(armiesByRegion))
	for regionID := range armiesByRegion {
		regionIDs = append(regionIDs, regionID)
	}
	// Sınırlı stok desteği deterministik dağıtılır; başkent önce gelir.
	sort.Slice(regionIDs, func(i, j int) bool {
		left := gs.Regions[regionIDs[i]]
		right := gs.Regions[regionIDs[j]]
		leftCapital := gs.IsCapitalRegion(left)
		rightCapital := gs.IsCapitalRegion(right)
		if leftCapital != rightCapital {
			return leftCapital
		}
		return regionIDs[i] < regionIDs[j]
	})
	availableReserveByFaction := make(map[string]int, len(gs.Factions))
	for fid, f := range gs.Factions {
		if f != nil && f.Grain > 0 {
			availableReserveByFaction[string(fid)] = f.Grain
		}
	}

	alerts := make([]state.RegionLogisticsStatus, 0)
	for _, rid := range regionIDs {
		armiesInRegion := armiesByRegion[rid]
		region := gs.Regions[rid]
		if region == nil {
			continue
		}

		ownerID := armiesInRegion[0].OwnerID
		totalDemand := 0
		totalUnits := 0
		peakTurns := 0
		for _, a := range armiesInRegion {
			_, externalSupplyActive := friendlySupplies[a.ID]
			totalDemand += gs.RegionalArmyGrainDemandWithExternalSupply(a, externalSupplyActive)
			totalUnits += len(a.Units)
			if a.OverCapacityTurns > peakTurns {
				peakTurns = a.OverCapacityTurns
			}
		}
		if totalUnits <= 0 {
			continue
		}
		if totalDemand <= 0 {
			regionStatus := state.RegionLogisticsStatus{
				RegionID:  rid,
				OwnerID:   ownerID,
				Capacity:  4,
				ArmyCount: len(armiesInRegion),
			}
			for _, a := range armiesInRegion {
				a.OverCapacityTurns = 0
				if !gs.CurrentSeason().IsWinter() && !gs.IsArmyDefendingSiegedRegion(a) {
					gs.ReplenishArmyInFriendlyTerritory(a, gs.ArmyReplenishmentHP(a))
				}
			}
			gs.RegionLogistics[rid] = regionStatus
			continue
		}

		militaryProduction := gs.RegionMilitaryGrainProduction(region)
		if region.OwnerID != ownerID {
			militaryProduction = 0
		}
		settlementBuffer := regionSettlementLogisticsBuffer(gs, region)
		blockadePercent := gs.RegionBlockadePercent(region, ownerID)
		settlementBuffer = settlementBuffer * (100 - blockadePercent) / 100
		availableReserve := availableReserveByFaction[ownerID]
		granarySupport := minInt(availableReserve, regionGranaryStorageCapacity(gs, region))
		availableReserve -= granarySupport
		reserveSupport := regionReserveSupport(availableReserve, militaryProduction, settlementBuffer)
		availableReserveByFaction[ownerID] = availableReserve - reserveSupport
		capacity := militaryProduction + settlementBuffer + granarySupport + reserveSupport
		if capacity < 4 {
			capacity = 4
		}
		overload := totalDemand - capacity

		regionStatus := state.RegionLogisticsStatus{
			RegionID:         rid,
			OwnerID:          ownerID,
			LocalProduction:  militaryProduction,
			SettlementBuffer: settlementBuffer,
			GranarySupport:   granarySupport,
			ReserveSupport:   reserveSupport,
			BlockadePercent:  blockadePercent,
			Demand:           totalDemand,
			Capacity:         capacity,
			Overload:         overload,
			ArmyCount:        len(armiesInRegion),
		}
		for _, a := range armiesInRegion {
			if supply, ok := friendlySupplies[a.ID]; ok {
				regionStatus.FriendlySupplyArmies++
				regionStatus.FriendlySupplyGrainSpent += supply.GrainSpent
			}
		}

		if overload <= 0 {
			for _, a := range armiesInRegion {
				a.OverCapacityTurns = 0
				if supply, ok := friendlySupplies[a.ID]; ok {
					gs.ArmyLogistics[a.ID] = friendlySupplyArmyLogisticsStatus(a, regionStatus, supply)
				}
				if !gs.CurrentSeason().IsWinter() && !gs.IsArmyDefendingSiegedRegion(a) {
					gs.ReplenishArmyInFriendlyTerritory(a, gs.ArmyReplenishmentHP(a))
				}
			}
			gs.RegionLogistics[rid] = regionStatus
			continue
		}

		damagePerUnit := logisticsDamagePerUnit(totalDemand, capacity, overload, peakTurns+1)
		for _, a := range armiesInRegion {
			a.OverCapacityTurns++
			armyDamagePerUnit := damagePerUnit
			if gs.IsArmyDefendingSiegedRegion(a) {
				armyDamagePerUnit = reduceAttritionDamageForGranary(armyDamagePerUnit, granaryAttritionReductionPercent(region))
			}
			armyStatus := state.ArmyLogisticsStatus{
				ArmyID:            a.ID,
				RegionID:          rid,
				OwnerID:           a.OwnerID,
				Demand:            totalDemand,
				Capacity:          capacity,
				Overload:          overload,
				OverCapacityTurns: a.OverCapacityTurns,
				DamagePerUnit:     armyDamagePerUnit,
			}
			if supply, ok := friendlySupplies[a.ID]; ok {
				armyStatus.FriendlySupplyFactionID = supply.ProviderFactionID
				armyStatus.FriendlySupplyRegionID = supply.ProviderRegionID
				armyStatus.FriendlySupplyGrainSpent = supply.GrainSpent
				armyStatus.FriendlySupplySameRealm = supply.SameRealm
			}
			unitsBefore := len(a.Units)
			totalDamage := 0
			survivors := a.Units[:0]
			for _, u := range a.Units {
				u.CurrentHP -= armyDamagePerUnit
				totalDamage += armyDamagePerUnit
				if u.CurrentHP <= 0 {
					armyStatus.UnitsLost++
					continue
				}
				survivors = append(survivors, u)
			}
			a.Units = survivors
			armyStatus.UnitsAffected = unitsBefore
			armyStatus.TotalHPDamage = totalDamage
			gs.ArmyLogistics[a.ID] = armyStatus
			if len(a.Units) == 0 {
				gs.RemoveArmy(a.ID)
			}

			regionStatus.UnitsAffected += armyStatus.UnitsAffected
			regionStatus.UnitsLost += armyStatus.UnitsLost
			regionStatus.TotalHPDamage += armyStatus.TotalHPDamage
			if armyStatus.OverCapacityTurns > regionStatus.PeakOverloadTurns {
				regionStatus.PeakOverloadTurns = armyStatus.OverCapacityTurns
			}
		}
		gs.RegionLogistics[rid] = regionStatus
		if ownerID == string(gs.PlayerFactionID) {
			alerts = append(alerts, regionStatus)
		}
	}

	return alerts
}

// allocateFriendlyFrontlineSupply ikmal alan orduları ArmyID sırasıyla ele alır;
// böylece aynı sınırlı müttefik/vassal rezervine birden fazla ordu talip olsa da
// sonuç kayıt yükleme ve AI turu arasında deterministik kalır.
func allocateFriendlyFrontlineSupply(gs *state.GameState) map[army.ArmyID]state.FriendlySupplySupport {
	if gs == nil {
		return nil
	}
	armyIDs := make([]army.ArmyID, 0, len(gs.Armies))
	for aid, a := range gs.Armies {
		if a != nil && !a.IsNaval && len(a.Units) > 0 {
			armyIDs = append(armyIDs, aid)
		}
	}
	sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
	supplies := make(map[army.ArmyID]state.FriendlySupplySupport)
	for _, aid := range armyIDs {
		a := gs.Armies[aid]
		supply, ok := gs.ExternalFriendlySupplyQuote(a)
		if !ok || supply.GrainSpent <= 0 {
			continue
		}
		provider := gs.Factions[supply.ProviderFactionID]
		if provider == nil || provider.IsEliminated {
			continue
		}
		reserve := friendlySupplyReserve(gs, supply.ProviderFactionID)
		if provider.Grain-supply.GrainSpent < reserve {
			continue
		}
		provider.Grain -= supply.GrainSpent
		if provider.Grain < 0 {
			provider.Grain = 0
		}
		status := gs.GrainEconomy[supply.ProviderFactionID]
		status.FriendlySupplyGrainSpent += supply.GrainSpent
		status.Stockpile = provider.Grain
		if status.TotalDemand > 0 {
			status.MonthsOfSupply = status.Stockpile / status.TotalDemand
		}
		status.StrategicDemand = state.StrategicGrainDemandFromStockpile(status.Stockpile, status.TotalDemand)
		gs.GrainEconomy[supply.ProviderFactionID] = status
		supplies[aid] = supply
	}
	return supplies
}

func friendlySupplyReserve(gs *state.GameState, fid faction.FactionID) int {
	const minimumReserve = 20
	if gs == nil {
		return minimumReserve
	}
	reserve := minimumReserve
	if status, ok := gs.GrainEconomy[fid]; ok && status.TotalDemand > reserve {
		reserve = status.TotalDemand
	}
	return reserve
}

func friendlySupplyArmyLogisticsStatus(a *army.Army, region state.RegionLogisticsStatus, supply state.FriendlySupplySupport) state.ArmyLogisticsStatus {
	return state.ArmyLogisticsStatus{
		ArmyID:                   a.ID,
		RegionID:                 region.RegionID,
		OwnerID:                  a.OwnerID,
		Demand:                   region.Demand,
		Capacity:                 region.Capacity,
		Overload:                 region.Overload,
		FriendlySupplyFactionID:  supply.ProviderFactionID,
		FriendlySupplyRegionID:   supply.ProviderRegionID,
		FriendlySupplyGrainSpent: supply.GrainSpent,
		FriendlySupplySameRealm:  supply.SameRealm,
	}
}

func regionSettlementLogisticsBuffer(gs *state.GameState, region *world.Region) int {
	buffer := 0
	for _, settlement := range region.Settlements {
		switch settlement.Type {
		case world.SettlementCity:
			buffer += 8
		case world.SettlementTown:
			buffer += 5
		case world.SettlementFortress:
			buffer += 6
		case world.SettlementPort:
			buffer += 6
		default:
			buffer += 4
		}
		if settlement.IsCenter {
			buffer += 4
		}
	}
	if gs != nil && gs.IsCapitalRegion(region) {
		buffer += state.CapitalRegionLogisticsBonus
	}
	if tc := region.TradeCapacity / 2; tc > 0 {
		if tc > 6 {
			tc = 6
		}
		buffer += tc
	}
	return buffer
}

func regionReserveSupport(availableGrain, production, settlementBuffer int) int {
	if availableGrain <= 0 {
		return 0
	}
	cap := production/2 + settlementBuffer/2 + 4
	if cap < 4 {
		cap = 4
	}
	reserve := availableGrain / 10
	if reserve > cap {
		reserve = cap
	}
	return reserve
}

func regionGranaryStorageCapacity(gs *state.GameState, region *world.Region) int {
	if gs == nil || region == nil {
		return 0
	}
	capacity := 0
	for _, buildingID := range region.Buildings {
		building := gs.BuildingTypes[buildingID]
		if building != nil && building.StorageCapacity > 0 {
			capacity += building.StorageCapacity
		}
	}
	return capacity
}

func logisticsDamagePerUnit(totalDemand, capacity, overload, nextTurn int) int {
	if totalDemand <= 0 || overload <= 0 {
		return 0
	}
	ratio := overload * 100 / max(1, totalDemand)
	damage := 2 + ratio/12
	if capacity <= 0 {
		damage += 3
	}
	if nextTurn > 1 {
		damage += (nextTurn - 1) * 3
	}
	if damage < 3 {
		damage = 3
	}
	if damage > 18 {
		damage = 18
	}
	return damage
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func applyTerrainSpecialization(
	terrain world.TerrainType,
	grain, iron, timber, stone, spice, cloth int,
) (int, int, int, int, int, int) {
	switch terrain {
	case world.TerrainPlain:
		grain = grain * 120 / 100
	case world.TerrainForest:
		timber = timber * 130 / 100
	case world.TerrainMountain:
		iron = iron * 125 / 100
		if stone <= 0 {
			stone = 1 + iron/3
		}
		stone = stone * 140 / 100
	case world.TerrainPass:
		if stone <= 0 {
			stone = 1
		}
		stone = stone * 120 / 100
	}
	return grain, iron, timber, stone, spice, cloth
}

const (
	rebellionPopulationPerUnit = 100
	rebellionBaseUnits         = 1
)

// rebellionUnitCount, halkın nüfusunu yerleşim/bina gelişmişliği ve eski
// sahibin tahıl ikmaliyle isyancı asker gücüne çevirir. Bu birimler üretim
// kuyruğundan geçmez; isyanın o anda sahada ortaya çıkan kuvvetidir.
func rebellionUnitCount(gs *state.GameState, region *world.Region) int {
	if region == nil {
		return 0
	}
	count := rebellionBaseUnits + region.Population/rebellionPopulationPerUnit
	count += len(region.Settlements) / 2
	count += len(region.Buildings) / 2
	if status, ok := gs.GrainEconomy[faction.FactionID(region.OwnerID)]; ok {
		switch status.SupplyLevel {
		case state.GrainSupplyWarning:
			count = count * 90 / 100
		case state.GrainSupplyCritical:
			count = count * 75 / 100
		case state.GrainSupplyFamine:
			count = count * 60 / 100
		}
	}
	if count < 1 {
		count = 1
	}
	if count > army.MaxArmySize {
		count = army.MaxArmySize
	}
	return count
}

func rebellionArmyID(gs *state.GameState, regionID world.RegionID) army.ArmyID {
	gs.NextArmySeq++
	id := army.ArmyID(fmt.Sprintf("army_rebel_%s_%d", regionID, gs.NextArmySeq))
	for gs.Armies[id] != nil {
		gs.NextArmySeq++
		id = army.ArmyID(fmt.Sprintf("army_rebel_%s_%d", regionID, gs.NextArmySeq))
	}
	return id
}

// formSuccessorFromRebellion, yalnız gerçekten elenmiş ve bölgede ardıl
// metadata'sı bulunan devletleri yeniden oyuna sokar.
func formSuccessorFromRebellion(gs *state.GameState, region *world.Region, rebel *army.Army) bool {
	if gs == nil || region == nil || rebel == nil || !gs.CanRestoreSuccessorAtRegion(region) {
		return false
	}
	successorID := faction.FactionID(region.SuccessorFactionID)
	successor := gs.Factions[successorID]
	successor.IsEliminated = false
	successor.OverlordID = ""
	region.OwnerID = string(successorID)
	rebel.OwnerID = string(successorID)
	rebel.IsRebel = false
	rebel.RebelAgainstID = ""
	gs.NormalizeFactionCapitals()
	return true
}

// checkRebellions isyan riskindeki bölgeleri kontrol eder. İlk kontrolde
// isyancı ordu doğar; sonraki turlarda eski sahibin ordusu bastırır, sahipsiz
// kalan ardıl devlet ise isyanı kazanıp bölgeyi kurar.
func checkRebellions(gs *state.GameState) {
	if gs == nil {
		return
	}
	if gs.Armies == nil {
		gs.Armies = make(map[army.ArmyID]*army.Army)
	}
	regionIDs := make([]world.RegionID, 0, len(gs.Regions))
	for id := range gs.Regions {
		regionIDs = append(regionIDs, id)
	}
	sort.Slice(regionIDs, func(i, j int) bool { return regionIDs[i] < regionIDs[j] })
	for _, regionID := range regionIDs {
		r := gs.Regions[regionID]
		if r == nil || r.IsSea {
			continue
		}
		var rebel *army.Army
		hasOwnerArmy := false
		for _, current := range gs.Armies {
			if current == nil || current.IsNaval || current.RegionID != r.ID {
				continue
			}
			if current.IsRebel {
				rebel = current
			} else if r.OwnerID != "" && current.OwnerID == r.OwnerID {
				hasOwnerArmy = true
			}
		}
		if rebel != nil {
			if hasOwnerArmy {
				gs.RemoveArmy(rebel.ID)
				r.OwnerID = rebel.RebelAgainstID
				r.Satisfaction = 50
				continue
			}
			if formSuccessorFromRebellion(gs, r, rebel) {
				continue
			}
			continue
		}
		if !r.IsRebellionRisk() || r.OwnerID == "" || hasOwnerArmy {
			continue
		}
		// Surlar isyanın o turda başlamasını engeller.
		if r.BuildingLevel("walls") > 0 {
			continue
		}
		formerOwner := r.OwnerID
		unitCount := rebellionUnitCount(gs, r)
		r.OwnerID = ""
		gs.ClearProductionOrdersForRegion(r.ID)
		r.Satisfaction = 50
		id := rebellionArmyID(gs, r.ID)
		rebelOwner := r.SuccessorFactionID
		if rebelOwner == "" {
			rebelOwner = "rebel_" + string(r.ID)
		}
		gs.Armies[id] = &army.Army{
			ID: id, OwnerID: rebelOwner, RegionID: r.ID,
			Units:         army.MakeUnits("militia", unitCount),
			MaxMovePoints: army.DefaultArmyMovePoints, MovePoints: army.DefaultArmyMovePoints,
			IsRebel: true, RebelAgainstID: formerOwner,
		}
	}
}

// checkEliminations kara toprağı kalmayan fraksiyonları elendi olarak işaretler ve ordularını temizler.
func checkEliminations(gs *state.GameState) {
	for fid, f := range gs.Factions {
		if f.IsEliminated {
			continue
		}
		if len(gs.LandRegionsOwnedBy(fid)) == 0 {
			eliminateFaction(gs, fid, "")
		}
	}
}

// applyRelationDecay savaş halindeki ilişkileri kötüleştirir, barış halindekini iyileştirir.
func applyRelationDecay(gs *state.GameState) {
	diplomacy.ApplyRelationDecay(gs)
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// applyReligionConversion her tur sahip olunan bölgelerin din dönüşümünü ilerletir.
// Sahibin dini bölgenin dininden farklıysa ConversionTurns artar.
// 24 turda din değişir ve memnuniyet cezası uygulanır.
func applyReligionConversion(gs *state.GameState) {
	const conversionThreshold = 24

	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID == "" {
			r.ConversionTurns = 0
			continue
		}
		ownerRel := ownerReligionStr(gs, r.OwnerID)
		if ownerRel == "" || ownerRel == r.Religion {
			r.ConversionTurns = 0
			continue
		}

		step := 1
		if ownerFaction := gs.Factions[faction.FactionID(r.OwnerID)]; ownerFaction != nil && gs.TechTypes != nil {
			fx := tech.ComputeEffects(ownerFaction.Research.Completed, gs.TechTypes)
			step += int(fx.ConversionSpeedMod)
		}
		if step < 1 {
			step = 1
		}
		r.ConversionTurns += step
		if r.ConversionTurns >= conversionThreshold {
			r.Religion = ownerRel
			r.ConversionTurns = 0
			// Din değişimi halk memnuniyetini düşürür
			r.Satisfaction -= 20
			if r.Satisfaction < 0 {
				r.Satisfaction = 0
			}
		}
	}
}

// ownerReligionStr bir fraksiyonun dinini döner; bulunamazsa "".
func ownerReligionStr(gs *state.GameState, ownerID string) string {
	for fid, f := range gs.Factions {
		if string(fid) == ownerID {
			return string(f.Religion)
		}
	}
	return ""
}
