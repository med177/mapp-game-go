package game

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/season"
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

type economyTickReport struct {
	PlayerLogisticsAlerts []state.RegionLogisticsStatus
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
	s := gs.CurrentSeason()

	if s.IsWinter() {
		for _, a := range gs.Armies {
			a.ApplyWinterAttrition()
		}
	}

	movMod := s.MovementMod()
	for _, a := range gs.Armies {
		if !s.IsWinter() {
			if a.IsNaval {
				replenishDockedFleet(gs, a, friendlyReplenishHP)
			} else {
				a.ReplenishInFriendlyTerritory(gs.Regions, friendlyReplenishHP)
			}
		}
		mp := 2 * movMod / 100
		if mp < 1 {
			mp = 1
		}
		// Kara ve deniz tech bonusları hareket havuzuna eklenir.
		if f, ok := gs.Factions[faction.FactionID(a.OwnerID)]; ok && gs.TechTypes != nil {
			fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
			mp += fx.MoveBonus
			if a.IsNaval {
				mp += fx.NavalMoveBonus
			}
		}
		// Difficulty 3: AI fraksiyonlar +1 hareket puanı bonusu alır
		if gs.Difficulty >= 3 && a.OwnerID != string(gs.PlayerFactionID) {
			mp++
		}
		a.MaxMovePoints = mp
		a.ResetMovePoints()
	}
}

func replenishDockedFleet(gs *state.GameState, fleet *army.Army, amount int) int {
	if gs == nil || fleet == nil || !fleet.IsNaval || amount <= 0 || fleet.DockedRegionID == "" {
		return 0
	}
	dockedRegion := gs.Regions[fleet.DockedRegionID]
	if dockedRegion == nil || dockedRegion.IsSea || dockedRegion.OwnerID == "" {
		return 0
	}
	healAmount := amount
	if dockedRegion.OwnerID != fleet.OwnerID {
		key := faction.RelationKey(faction.FactionID(fleet.OwnerID), faction.FactionID(dockedRegion.OwnerID))
		rel, ok := gs.Relations[key]
		if !ok || rel.Stance != faction.StanceAllied {
			return 0
		}
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
	grainByFaction := make(map[string]int)
	ironByFaction := make(map[string]int)
	timberByFaction := make(map[string]int)
	stoneByFaction := make(map[string]int)
	spiceByFaction := make(map[string]int)
	clothByFaction := make(map[string]int)

	for _, r := range gs.Regions {
		if r.IsSea || r.OwnerID == "" {
			continue
		}
		if gs.SiegeAt(r.ID) != nil {
			// Kuşatma altındaki bölge devlete gelir/hammadde sağlamaz
			// ve halkın memnuniyeti kuşatma stresinden her tur azalır
			r.Satisfaction = clamp(r.Satisfaction-5, 0, 100)
			continue
		}

		// Bina çarpanları
		goldMod := 1.0
		grainMod := 1.0
		tradeCapMod := 1.0
		satBonus := 0
		for _, bid := range r.Buildings {
			if b, ok := gs.BuildingTypes[bid]; ok {
				goldMod *= b.GoldMod
				grainMod *= b.GrainMod
				tradeCapMod *= b.TradeCapacityMod
				satBonus += b.SatBonus
			}
		}

		income := int(float64(r.GoldIncome()) * goldMod * float64(harvestMod) / 100)
		grain := int(float64(r.BaseGrainOutput) * grainMod)
		iron := r.BaseIronOutput
		timber := r.BaseTimberOutput
		stone := r.BaseStoneOutput
		spice := r.BaseSpiceOutput
		cloth := r.BaseClothOutput
		grain, iron, timber, stone, spice, cloth = applyTerrainSpecialization(r.Terrain, grain, iron, timber, stone, spice, cloth)

		// Pasif ticaret geliri (TradeCapacity bazlı)
		// TradeCapacityMod: pazar ve liman gibi binalar ticaret kapasitesini artırır
		tradeIncome := economy.RegionTradeIncome(r.TradeCapacity, tradeCapMod)
		// Mevsimsel ticaret modu uygula
		tradeIncome = tradeIncome * s.TradeMod() / 100
		if fx, ok := effectsByFaction[r.OwnerID]; ok && fx.MarketGoldMod != 0 {
			tradeIncome = int(float64(tradeIncome) * (1.0 + fx.MarketGoldMod))
		}

		incomeByFaction[r.OwnerID] += income + tradeIncome
		grainByFaction[r.OwnerID] += grain
		ironByFaction[r.OwnerID] += iron
		timberByFaction[r.OwnerID] += timber
		stoneByFaction[r.OwnerID] += stone
		spiceByFaction[r.OwnerID] += spice
		clothByFaction[r.OwnerID] += cloth
		if bonus := gs.CapitalRegionBonus(r); bonus != (state.RegionProductionSummary{}) {
			incomeByFaction[r.OwnerID] += bonus.Gold
			grainByFaction[r.OwnerID] += bonus.Grain
			ironByFaction[r.OwnerID] += bonus.Iron
			timberByFaction[r.OwnerID] += bonus.Timber
			stoneByFaction[r.OwnerID] += bonus.Stone
			spiceByFaction[r.OwnerID] += bonus.Spice
			clothByFaction[r.OwnerID] += bonus.Cloth
		}

		// Vergi memnuniyet etkisi + bina bonusu
		delta := economy.TaxSatisfactionDelta(r.TaxRate) + satBonus
		r.Satisfaction = clamp(r.Satisfaction+delta, 0, 100)
	}

	// --- Ticaret rotalarını işlet (mal + altın transferi) ---
	tradeLogs := economy.ApplyTradeRoutes(gs.Factions, gs.TradeRoutes)
	for _, log := range tradeLogs {
		// Ticaret logları oyuncuya aitse göster
		if gs.PlayerFactionID != "" {
			_ = log // ileride oyuncuya bildirim gösterilebilir
		}
	}

	// Gerçek ordu bakım maliyetleri (UnitType.GrainUpkeep)
	upkeepByFaction := make(map[string]int)
	for _, a := range gs.Armies {
		upkeepByFaction[a.OwnerID] += a.TotalGrainUpkeep(gs.UnitTypes)
	}

	for fid, f := range gs.Factions {
		fidStr := string(fid)

		// Teknoloji bonusları
		fx := effectsByFaction[fidStr]

		// GoldPerRegion tech bonusu
		ownedCount := len(gs.RegionsOwnedBy(fid))
		techGold := fx.GoldPerRegion * ownedCount

		f.Gold += incomeByFaction[fidStr] + techGold
		netGrain := int(float64(grainByFaction[fidStr]) * (1.0 + fx.GrainMod))
		f.Grain += netGrain - upkeepByFaction[fidStr]
		f.Iron += int(float64(ironByFaction[fidStr]) * (1.0 + fx.IronMod))
		f.Timber += int(float64(timberByFaction[fidStr]) * (1.0 + fx.TimberMod))
		f.Stone += int(float64(stoneByFaction[fidStr]) * (1.0 + fx.StoneMod))
		f.Spice += spiceByFaction[fidStr]
		f.Cloth += clothByFaction[fidStr]

		// Memnuniyet tech bonusu tüm bölgelere
		if fx.SatisfactionBonus > 0 {
			for _, r := range gs.Regions {
				if r.OwnerID == fidStr {
					r.Satisfaction = clamp(r.Satisfaction+fx.SatisfactionBonus, 0, 100)
				}
			}
		}

		if f.Gold < 0 {
			f.Gold = 0
		}
		if f.Grain < 0 {
			applyGrainShortagePenalty(gs, fidStr, -f.Grain)
			f.Grain = 0
		}
	}

	for fid, f := range gs.Factions {
		if f == nil || f.IsEliminated || f.OverlordID == "" {
			continue
		}
		overlord := gs.Factions[f.OverlordID]
		if overlord == nil || overlord.IsEliminated {
			f.OverlordID = ""
			continue
		}
		income := incomeByFaction[string(fid)] + effectsByFaction[string(fid)].GoldPerRegion*len(gs.RegionsOwnedBy(fid))
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
	}

	report.PlayerLogisticsAlerts = applyRegionalLogisticsPressure(gs)

	// --- Dinamik piyasa fiyatlarını güncelle ---
	gs.MarketPrices = economy.ComputeMarketPrices(gs.Factions)
	return report
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
		if a.DockedRegionID != "" || len(a.EmbarkedUnits) == 0 {
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
	for _, a := range gs.Armies {
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

func applyRegionalLogisticsPressure(gs *state.GameState) []state.RegionLogisticsStatus {
	gs.RegionLogistics = make(map[world.RegionID]state.RegionLogisticsStatus)
	gs.ArmyLogistics = make(map[army.ArmyID]state.ArmyLogisticsStatus)
	if gs == nil {
		return nil
	}

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

	alerts := make([]state.RegionLogisticsStatus, 0)
	for rid, armiesInRegion := range armiesByRegion {
		region := gs.Regions[rid]
		if region == nil {
			continue
		}

		ownerID := armiesInRegion[0].OwnerID
		totalDemand := 0
		totalUnits := 0
		peakTurns := 0
		for _, a := range armiesInRegion {
			totalDemand += a.TotalGrainUpkeep(gs.UnitTypes)
			totalUnits += len(a.Units)
			if a.OverCapacityTurns > peakTurns {
				peakTurns = a.OverCapacityTurns
			}
		}
		if totalDemand <= 0 || totalUnits <= 0 {
			for _, a := range armiesInRegion {
				a.OverCapacityTurns = 0
			}
			continue
		}

		production := gs.RegionProductionSummary(region).Grain
		settlementBuffer := regionSettlementLogisticsBuffer(gs, region)
		reserveSupport := regionReserveSupport(gs, ownerID, production, settlementBuffer)
		capacity := production + settlementBuffer + reserveSupport
		if capacity < 4 {
			capacity = 4
		}
		overload := totalDemand - capacity

		regionStatus := state.RegionLogisticsStatus{
			RegionID:         rid,
			OwnerID:          ownerID,
			LocalProduction:  production,
			SettlementBuffer: settlementBuffer,
			ReserveSupport:   reserveSupport,
			Demand:           totalDemand,
			Capacity:         capacity,
			Overload:         overload,
			ArmyCount:        len(armiesInRegion),
		}

		if overload <= 0 {
			for _, a := range armiesInRegion {
				a.OverCapacityTurns = 0
			}
			gs.RegionLogistics[rid] = regionStatus
			continue
		}

		damagePerUnit := logisticsDamagePerUnit(totalDemand, capacity, overload, peakTurns+1)
		for _, a := range armiesInRegion {
			a.OverCapacityTurns++
			armyStatus := state.ArmyLogisticsStatus{
				ArmyID:            a.ID,
				RegionID:          rid,
				OwnerID:           a.OwnerID,
				Demand:            totalDemand,
				Capacity:          capacity,
				Overload:          overload,
				OverCapacityTurns: a.OverCapacityTurns,
				DamagePerUnit:     damagePerUnit,
			}
			unitsBefore := len(a.Units)
			totalDamage := 0
			survivors := a.Units[:0]
			for _, u := range a.Units {
				u.CurrentHP -= damagePerUnit
				totalDamage += damagePerUnit
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
		if settlement.IsCapital {
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

func regionReserveSupport(gs *state.GameState, ownerID string, production, settlementBuffer int) int {
	if gs == nil || ownerID == "" {
		return 0
	}
	f := gs.Factions[faction.FactionID(ownerID)]
	if f == nil || f.Grain <= 0 {
		return 0
	}
	cap := production/2 + settlementBuffer/2 + 4
	if cap < 4 {
		cap = 4
	}
	reserve := f.Grain / 10
	if reserve > cap {
		reserve = cap
	}
	return reserve
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

// checkRebellions isyan riski olan bölgeleri kontrol eder.
func checkRebellions(gs *state.GameState) {
	for _, r := range gs.Regions {
		if !r.IsRebellionRisk() {
			continue
		}
		hasGarrison := false
		for _, a := range gs.Armies {
			if a.RegionID == r.ID && !a.IsNaval {
				hasGarrison = true
				break
			}
		}
		// Surlar isyanı bastırır
		for _, bid := range r.Buildings {
			if bid == "walls" {
				hasGarrison = true
				break
			}
		}
		if !hasGarrison {
			r.OwnerID = ""
			gs.ClearProductionOrdersForRegion(r.ID)
			r.Satisfaction = 50
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

var _ = season.SeasonWinter
