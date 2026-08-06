// Package satisfaction, bölge memnuniyetinin tur başı değişimini ortaklaştırır.
package satisfaction

import (
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// Breakdown, bir bölgenin ekonomi turundaki memnuniyet hesabının bileşenlerini
// taşır. Alanlar doğrudan uygulanan deltadır; negatif değerler cezayı gösterir.
type Breakdown struct {
	Valid         bool
	Annual        int
	Siege         int
	Tax           int
	Buildings     int
	Grain         int
	Technology    int
	WarFatigue    int
	Overextension int
	Army          int
	Total         int
}

// Calculator, aynı ekonomi tick'inde çok sayıda bölgenin hesabını yaparken
// ortak savaş ve kara bölgesi sayaçlarını yeniden kullanır.
type Calculator struct {
	gs             *state.GameState
	warFatigueByID map[faction.FactionID]int
	landCountByID  map[faction.FactionID]int
}

// NewCalculator bir ekonomi tick'i veya tek bölge UI önizlemesi için hesap
// bağlamı oluşturur.
func NewCalculator(gs *state.GameState) *Calculator {
	calculator := &Calculator{
		gs:             gs,
		warFatigueByID: make(map[faction.FactionID]int),
		landCountByID:  make(map[faction.FactionID]int),
	}
	if gs == nil {
		return calculator
	}
	for fid := range gs.Factions {
		calculator.warFatigueByID[fid] = diplomacy.IndependentWarSatisfactionPenalty(gs, fid)
	}
	for _, region := range gs.Regions {
		if region != nil && !region.IsSea && region.OwnerID != "" {
			calculator.landCountByID[factionID(region.OwnerID)]++
		}
	}
	return calculator
}

// Calculate bölgenin mevcut state'e göre bir sonraki ekonomi turunda alacağı
// memnuniyet değişimini hesaplar. Ekonomi çözümlemesi de aynı helper'ı kullanır.
func Calculate(gs *state.GameState, region *world.Region) Breakdown {
	return NewCalculator(gs).ForRegion(region)
}

// ForRegion hesap bağlamındaki tek bir bölgenin memnuniyet deltasıdır.
func (calculator *Calculator) ForRegion(region *world.Region) Breakdown {
	var breakdown Breakdown
	if calculator == nil || calculator.gs == nil || region == nil || region.IsSea || region.OwnerID == "" {
		return breakdown
	}
	gs := calculator.gs
	breakdown.Valid = true

	if gs.CurrentTurnIncludesMonth(12) {
		breakdown.Annual = -1
	}
	if gs.SiegeAt(region.ID) != nil {
		breakdown.Siege = -5
	} else {
		breakdown.Tax = economy.TaxSatisfactionDelta(region.TaxRate)
		for _, buildingID := range region.Buildings {
			if building := gs.BuildingTypes[buildingID]; building != nil {
				breakdown.Buildings += building.SatBonus
			}
		}
		breakdown.Grain = grainDelta(gs, region.OwnerID)
	}

	if owner := gs.Factions[factionID(region.OwnerID)]; owner != nil {
		if effects := tech.ComputeEffects(owner.Research.Completed, gs.TechTypes); effects.SatisfactionBonus > 0 {
			breakdown.Technology = effects.SatisfactionBonus
		}
	}
	warFatigue := calculator.warFatigueByID[factionID(region.OwnerID)]
	if _, known := calculator.warFatigueByID[factionID(region.OwnerID)]; !known {
		warFatigue = diplomacy.IndependentWarSatisfactionPenalty(gs, factionID(region.OwnerID))
		calculator.warFatigueByID[factionID(region.OwnerID)] = warFatigue
	}
	if warFatigue > 0 {
		breakdown.WarFatigue = -warFatigue
	}
	if calculator.landCountByID[factionID(region.OwnerID)] > 20 {
		breakdown.Overextension = -1
	}
	breakdown.Army = ArmyStabilityBonus(gs, region)
	breakdown.Total = breakdown.Annual + breakdown.Siege + breakdown.Tax + breakdown.Buildings + breakdown.Grain + breakdown.Technology + breakdown.WarFatigue + breakdown.Overextension + breakdown.Army
	return breakdown
}

// ArmyStabilityBonus bölgedeki sahibine ait kara ordularının istikrar katkısını
// hesaplar. 100 güç +10 verir ve üst sınır +10'dur.
func ArmyStabilityBonus(gs *state.GameState, region *world.Region) int {
	if gs == nil || region == nil || region.OwnerID == "" || gs.Armies == nil {
		return 0
	}
	strength := 0
	for _, currentArmy := range gs.Armies {
		if currentArmy == nil || currentArmy.IsNaval || currentArmy.OwnerID != region.OwnerID || currentArmy.RegionID != region.ID || len(currentArmy.Units) == 0 {
			continue
		}
		strength += currentArmy.TotalStrength(gs.UnitTypes)
	}
	bonus := strength / 10
	if bonus > 10 {
		return 10
	}
	return bonus
}

func grainDelta(gs *state.GameState, ownerID string) int {
	status, ok := gs.GrainEconomy[factionID(ownerID)]
	penalty := 0
	if ok {
		switch status.SupplyLevel {
		case state.GrainSupplyWarning:
			penalty = 1
		case state.GrainSupplyCritical:
			penalty = 2
		case state.GrainSupplyFamine:
			penalty = 4
		}
	}
	if owner := gs.Factions[factionID(ownerID)]; owner != nil && owner.Grain == 0 {
		penalty = 5
	}
	return -penalty
}

func factionID(id string) faction.FactionID {
	return faction.FactionID(id)
}
