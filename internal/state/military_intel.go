package state

import (
	"hash/fnv"
	"strconv"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const (
	// MilitaryPowerUncertaintyPercent gerçek askerî gücün iki tarafına uygulanır.
	// Değer yalnızca istihbarat ve arayüzde kullanılır; savaş çözümlemesi gerçeği
	// kullanmaya devam eder.
	MilitaryPowerUncertaintyPercent = 20
	militaryIntelRefreshTurns       = 1
)

// MilitaryPowerEstimate bir gözlemcinin hedef devlet hakkında bildiği tahmini
// toplam askerî gücü döndürür. Devlet kendi gücünü ve reveal_enemy_strength
// teknolojisine sahip olduğu hedefin gücünü tam bilir.
func MilitaryPowerEstimate(gs *GameState, observer, target faction.FactionID) int {
	land, naval, exact := militaryPowerBreakdownExact(gs, target)
	if exact || observer == target || HasMilitaryIntel(gs, observer) ||
		HasTemporaryMilitaryIntel(gs, observer, target) || debugRevealMilitaryPower(gs) {
		return land + naval
	}
	land, naval = applyMilitaryIntelError(gs, observer, target, "faction", land, naval)
	return land + naval
}

// MilitaryPowerEstimateBreakdown, diplomasi ekranındaki kara/deniz ayrımını
// aynı gözlem hatasıyla döndürür. Böylece toplam ve alt kalemler çelişmez.
func MilitaryPowerEstimateBreakdown(gs *GameState, observer, target faction.FactionID) (land, naval int, exact bool) {
	land, naval, exact = militaryPowerBreakdownExact(gs, target)
	if exact || observer == target || HasMilitaryIntel(gs, observer) ||
		HasTemporaryMilitaryIntel(gs, observer, target) || debugRevealMilitaryPower(gs) {
		return land, naval, true
	}
	land, naval = applyMilitaryIntelError(gs, observer, target, "faction", land, naval)
	return land, naval, false
}

// ArmyPowerEstimate tek bir düşman ordusu için aynı istihbarat modelini uygular.
// Ordu paneli ve savaş özeti gibi ekranlar bu yardımcıyı kullanmalıdır.
func ArmyPowerEstimate(gs *GameState, observer faction.FactionID, armyRef *army.Army) (power int, exact bool) {
	if armyRef == nil {
		return 0, true
	}
	var unitTypes map[string]*army.UnitType
	if gs != nil {
		unitTypes = gs.UnitTypes
	}
	power = armyRef.TotalStrength(unitTypes)
	target := faction.FactionID(armyRef.OwnerID)
	if gs == nil || observer == target || HasMilitaryIntel(gs, observer) ||
		HasTemporaryMilitaryIntel(gs, observer, target) ||
		HasTemporaryRegionMilitaryIntel(gs, observer, armyRef.RegionID) || debugRevealMilitaryPower(gs) {
		return power, true
	}
	land, _ := applyMilitaryIntelError(gs, observer, faction.FactionID(armyRef.OwnerID), "army:"+string(armyRef.ID), power, 0)
	return land, false
}

// ArmyDefenseEstimate, kuşatma/temas önizlemelerinde saldırı tahminiyle aynı
// gözlem katsayısını savunma değerine uygular.
func ArmyDefenseEstimate(gs *GameState, observer faction.FactionID, armyRef *army.Army) (power int, exact bool) {
	if armyRef == nil {
		return 0, true
	}
	unitTypes := map[string]*army.UnitType(nil)
	if gs != nil {
		unitTypes = gs.UnitTypes
	}
	power = armyRef.TotalDefense(unitTypes)
	target := faction.FactionID(armyRef.OwnerID)
	if gs == nil || observer == target || HasMilitaryIntel(gs, observer) ||
		HasTemporaryMilitaryIntel(gs, observer, target) ||
		HasTemporaryRegionMilitaryIntel(gs, observer, armyRef.RegionID) || debugRevealMilitaryPower(gs) {
		return power, true
	}
	land, _ := applyMilitaryIntelError(gs, observer, faction.FactionID(armyRef.OwnerID), "army-defense:"+string(armyRef.ID), power, 0)
	return land, false
}

func debugRevealMilitaryPower(gs *GameState) bool {
	return gs != nil && gs.DevelopmentMode && gs.DebugRevealMilitaryPower
}

// EnsureDecisionSeed eski save'ler ve test fixture'ları için geriye dönük
// güvenli bir seed üretir. Yeni kampanyalar bunu zaman tabanlı seed ile
// başlatır; bu fallback yalnızca kayıt seed taşımıyorsa kullanılır.
func (s *GameState) EnsureDecisionSeed() {
	if s == nil || s.DecisionSeed != 0 {
		return
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(s.ScenarioID + "|" + string(s.PlayerFactionID) + "|" + strconv.Itoa(s.Turn) + "|" + strconv.Itoa(s.Year)))
	s.DecisionSeed = h.Sum64()
	if s.DecisionSeed == 0 {
		s.DecisionSeed = 1
	}
}

// HasMilitaryIntel, observerın hedef devletlerin gücünü tam görmesini sağlayan
// teknolojiye sahip olup olmadığını kontrol eder.
func HasMilitaryIntel(gs *GameState, observer faction.FactionID) bool {
	if gs == nil || observer == "" || gs.TechTypes == nil {
		return false
	}
	fx := gs.Factions[observer]
	if fx == nil {
		return false
	}
	return tech.ComputeEffects(fx.Research.Completed, gs.TechTypes).RevealEnemyStrength
}

// HasTemporaryMilitaryIntel reports whether an observer has current, targeted
// intelligence about a faction's exact military strength.
func HasTemporaryMilitaryIntel(gs *GameState, observer, target faction.FactionID) bool {
	return gs != nil && observer != "" && target != "" &&
		gs.TemporaryMilitaryIntel[observer][target] > 0
}

// GrantTemporaryMilitaryIntel grants or refreshes targeted military intelligence.
func (s *GameState) GrantTemporaryMilitaryIntel(observer, target faction.FactionID, turns int) {
	if s == nil || observer == "" || target == "" || observer == target || turns <= 0 {
		return
	}
	if s.TemporaryMilitaryIntel == nil {
		s.TemporaryMilitaryIntel = make(map[faction.FactionID]map[faction.FactionID]int)
	}
	targets := s.TemporaryMilitaryIntel[observer]
	if targets == nil {
		targets = make(map[faction.FactionID]int)
		s.TemporaryMilitaryIntel[observer] = targets
	}
	if targets[target] < turns {
		targets[target] = turns
	}
}

// HasTemporaryRegionMilitaryIntel reports whether an observer has current
// exact military intelligence for a specific region.
func HasTemporaryRegionMilitaryIntel(gs *GameState, observer faction.FactionID, target world.RegionID) bool {
	return gs != nil && observer != "" && target != "" &&
		gs.TemporaryRegionMilitaryIntel[observer][target] > 0
}

// GrantTemporaryRegionMilitaryIntel grants or refreshes region-targeted
// military intelligence.
func (s *GameState) GrantTemporaryRegionMilitaryIntel(observer faction.FactionID, target world.RegionID, turns int) {
	if s == nil || observer == "" || target == "" || turns <= 0 || s.Regions[target] == nil {
		return
	}
	if s.TemporaryRegionMilitaryIntel == nil {
		s.TemporaryRegionMilitaryIntel = make(map[faction.FactionID]map[world.RegionID]int)
	}
	targets := s.TemporaryRegionMilitaryIntel[observer]
	if targets == nil {
		targets = make(map[world.RegionID]int)
		s.TemporaryRegionMilitaryIntel[observer] = targets
	}
	if targets[target] < turns {
		targets[target] = turns
	}
}

func (s *GameState) tickTemporaryMilitaryIntel() {
	if s == nil {
		return
	}
	for observer, targets := range s.TemporaryMilitaryIntel {
		for target, turns := range targets {
			if turns <= 1 {
				delete(targets, target)
				continue
			}
			targets[target] = turns - 1
		}
		if len(targets) == 0 {
			delete(s.TemporaryMilitaryIntel, observer)
		}
	}
	for observer, targets := range s.TemporaryRegionMilitaryIntel {
		for target, turns := range targets {
			if turns <= 1 {
				delete(targets, target)
				continue
			}
			targets[target] = turns - 1
		}
		if len(targets) == 0 {
			delete(s.TemporaryRegionMilitaryIntel, observer)
		}
	}
}

func militaryPowerBreakdownExact(gs *GameState, target faction.FactionID) (land, naval int, exact bool) {
	if gs == nil || target == "" {
		return 0, 0, true
	}
	for _, armyRef := range gs.Armies {
		if armyRef == nil || armyRef.OwnerID != string(target) {
			continue
		}
		power := armyRef.TotalStrength(gs.UnitTypes)
		if armyRef.IsNaval {
			naval += power
		} else {
			land += power
		}
	}
	return land, naval, false
}

func applyMilitaryIntelError(gs *GameState, observer, target faction.FactionID, scope string, land, naval int) (int, int) {
	factor := militaryIntelFactor(gs, observer, target, scope)
	land = scaleIntelValue(land, factor)
	naval = scaleIntelValue(naval, factor)
	return land, naval
}

func scaleIntelValue(value, factor int) int {
	if value <= 0 {
		return 0
	}
	result := value * factor / 100
	if result < 1 {
		return 1
	}
	return result
}

func militaryIntelFactor(gs *GameState, observer, target faction.FactionID, scope string) int {
	seed := uint64(0)
	turn := 0
	if gs != nil {
		seed = gs.DecisionSeed
		turn = gs.Turn / militaryIntelRefreshTurns
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(strconv.FormatUint(seed, 10)))
	_, _ = h.Write([]byte("|" + strconv.Itoa(turn) + "|" + string(observer) + "|" + string(target) + "|" + scope))
	return 100 - MilitaryPowerUncertaintyPercent + int(h.Sum32()%uint32(MilitaryPowerUncertaintyPercent*2+1))
}
