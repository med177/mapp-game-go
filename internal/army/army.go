package army

import (
	"strings"

	"mapp-game-go/internal/world"
)

const MaxArmySize = 20

// ArmyID ordu benzersiz kimliği.
type ArmyID string

// Army harita üzerindeki bir orduyu temsil eder.
type Army struct {
	ID                 ArmyID         `json:"id"`
	OwnerID            string         `json:"owner_id"` // fraksiyon ID
	RegionID           world.RegionID `json:"region_id"`
	DockedRegionID     world.RegionID `json:"docked_region_id,omitempty"`
	DockedSettlementID string         `json:"docked_settlement_id,omitempty"`
	Units              []Unit         `json:"units"`
	EmbarkedUnits      []Unit         `json:"embarked_units,omitempty"` // filo içindeki kara birimleri
	MovePoints         int            `json:"move_points"`              // bu turda kalan hareket puanı
	MaxMovePoints      int            `json:"max_move_points"`
	IsNaval            bool           `json:"is_naval"` // deniz ordusu mu?
	IsGarrison         bool           `json:"is_garrison,omitempty"`
	Commander          *Commander     `json:"commander,omitempty"`
	// EmbarkedCommander kara birlikleri filoda taşınırken komutan bağlantısını korur.
	EmbarkedCommander *Commander `json:"embarked_commander,omitempty"`

	// Pusu durumu: geçit bölgesinde bekliyorsa true
	InAmbush bool `json:"in_ambush"`

	// Bölgesel ikmal kapasitesi aşımı üst üste kaç tur sürdü.
	OverCapacityTurns int `json:"over_capacity_turns,omitempty"`

	// Limana uğramadan denizde, taşınan birliklerle geçirilen ardışık tur sayısı.
	TurnsWithoutPort int `json:"turns_without_port,omitempty"`
}

// AssignCommander komutanı bu orduya atar.
func (a *Army) AssignCommander(commander *Commander) bool {
	if a == nil || commander == nil {
		return false
	}
	commander.Normalize()
	if commander.OwnerID == "" {
		commander.OwnerID = a.OwnerID
	}
	if a.Commander != nil && a.Commander != commander && a.Commander.AssignedArmyID == a.ID {
		a.Commander.AssignedArmyID = ""
	}
	commander.AssignedArmyID = a.ID
	a.Commander = commander
	return true
}

// RemoveCommander komutanı ordudan ayırır ve ayrılan komutanı döner.
func (a *Army) RemoveCommander() *Commander {
	if a == nil {
		return nil
	}
	commander := a.Commander
	a.Commander = nil
	if commander != nil && commander.AssignedArmyID == a.ID {
		commander.AssignedArmyID = ""
	}
	return commander
}

// CommanderModifiers ordu savaş hesabında kullanılacak komutan çarpanlarını döner.
func (a *Army) CommanderModifiers() (attack, defense float64) {
	if a == nil || a.Commander == nil {
		return 0, 0
	}
	return a.Commander.AttackModifier(), a.Commander.DefenseModifier()
}

// CommanderMoraleModifier ordunun moral tabanlı çarpanını döner.
func (a *Army) CommanderMoraleModifier() float64 {
	if a == nil || a.Commander == nil {
		return 0
	}
	return a.Commander.MoraleModifier()
}

// CommanderMoveBonus ordunun komutanından gelen ekstra hareket puanını döner.
func (a *Army) CommanderMoveBonus() int {
	if a == nil || a.Commander == nil {
		return 0
	}
	return a.Commander.MoveBonus()
}

// CommanderSiegeBonuses ordunun komutanından gelen kuşatma ilerleme bonuslarını döner.
func (a *Army) CommanderSiegeBonuses() (progress, breach int) {
	if a == nil || a.Commander == nil {
		return 0, 0
	}
	return a.Commander.SiegeBonuses()
}

// RecordBattle sonucu bu ordunun komutanına aktarır.
func (a *Army) RecordBattle(won bool) CommanderProgress {
	if a == nil || a.Commander == nil {
		return CommanderProgress{}
	}
	return a.Commander.RecordBattle(won)
}

// Size ordu boyutunu döner.
func (a *Army) Size() int {
	return len(a.Units)
}

// CountsTowardArmyLimit ordu kara saha ordusu limitine dahil mi?
func (a *Army) CountsTowardArmyLimit() bool {
	return a != nil && !a.IsNaval && !a.IsGarrison && len(a.Units) > 0
}

// CanAddUnit yeni birim eklenebilir mi?
func (a *Army) CanAddUnit() bool {
	return len(a.Units) < MaxArmySize
}

// CanEmbark ordudaki tüm kara birimlerinin deniz taşımaya uygun olup olmadığını döner.
func (a *Army) CanEmbark(types map[string]*UnitType) bool {
	if a == nil || a.IsNaval || len(a.Units) == 0 {
		return false
	}
	for _, u := range a.Units {
		ut, ok := types[u.TypeID]
		if !ok || ut == nil || !ut.Embarkable {
			return false
		}
	}
	return true
}

// EmbarkBlockerNames orduda deniz taşımayı engelleyen birim tiplerini döner.
func (a *Army) EmbarkBlockerNames(types map[string]*UnitType) []string {
	if a == nil || len(a.Units) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(a.Units))
	names := make([]string, 0, len(a.Units))
	for _, u := range a.Units {
		ut, ok := types[u.TypeID]
		if ok && ut != nil && ut.Embarkable {
			continue
		}
		name := u.TypeID
		if ok && ut != nil && ut.NameTR != "" {
			name = ut.NameTR
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

// EmbarkedCount filonun taşıdığı kara birimi sayısını döner.
func (a *Army) EmbarkedCount() int {
	if a == nil {
		return 0
	}
	return len(a.EmbarkedUnits)
}

// TransportCapacity filodaki nakliye gemilerinin toplam kapasitesini döner.
func (a *Army) TransportCapacity(types map[string]*UnitType) int {
	if a == nil || !a.IsNaval {
		return 0
	}
	total := 0
	for _, u := range a.Units {
		t, ok := types[u.TypeID]
		if !ok || t == nil || t.Category != CategoryNavalTrans || t.CarryCapacity <= 0 {
			continue
		}
		total += t.CarryCapacity
	}
	if total > MaxArmySize {
		total = MaxArmySize
	}
	return total
}

// AvailableTransportCapacity filoda kalan taşıma slotunu döner.
func (a *Army) AvailableTransportCapacity(types map[string]*UnitType) int {
	capacity := a.TransportCapacity(types) - a.EmbarkedCount()
	if capacity < 0 {
		return 0
	}
	return capacity
}

// CanEmbarkUnits filonun verilen sayıda kara birimini taşıyıp taşıyamadığını döner.
func (a *Army) CanEmbarkUnits(types map[string]*UnitType, count int) bool {
	if count <= 0 {
		return false
	}
	return a.AvailableTransportCapacity(types) >= count
}

// TotalStrength ordunun toplam saldırı gücünü hesaplar.
func (a *Army) TotalStrength(types map[string]*UnitType) int {
	total := 0
	for _, u := range a.Units {
		t, ok := types[u.TypeID]
		if !ok {
			continue
		}
		power := u.EffectiveAttack(types) + t.Morale/10
		scaled := int(float64(power) * u.HPPercent())
		if scaled < 1 {
			scaled = 1
		}
		total += scaled
	}
	return total
}

// TotalDefense ordunun toplam savunma gücünü hesaplar.
func (a *Army) TotalDefense(types map[string]*UnitType) int {
	total := 0
	for _, u := range a.Units {
		power := int(float64(u.EffectiveDefense(types)) * u.HPPercent())
		if power < 1 {
			power = 1
		}
		total += power
	}
	return total
}

// TotalGrainUpkeep ordudaki tüm birimlerin tur başı tahıl bakım yükünü döner.
func (a *Army) TotalGrainUpkeep(types map[string]*UnitType) int {
	total := 0
	for _, u := range a.Units {
		if t, ok := types[u.TypeID]; ok && t != nil {
			total += t.GrainUpkeep
		}
	}
	return total
}

func (a *Army) SiegeUnitScore(types map[string]*UnitType) int {
	if a == nil {
		return 0
	}
	score := 0
	for _, u := range a.Units {
		t, ok := types[u.TypeID]
		if !ok || t == nil || t.Category != CategorySiege {
			continue
		}
		score += 2 + int(t.Tier)
	}
	return score
}

func (a *Army) HighestSiegeTier(types map[string]*UnitType) int {
	if a == nil {
		return 0
	}
	best := 0
	for _, u := range a.Units {
		t, ok := types[u.TypeID]
		if !ok || t == nil || t.Category != CategorySiege {
			continue
		}
		if tier := int(t.Tier); tier > best {
			best = tier
		}
	}
	return best
}

func (a *Army) CanBreachFortification(types map[string]*UnitType, fortLevel int) bool {
	if fortLevel <= 0 {
		return true
	}
	return a.HighestSiegeTier(types) >= fortLevel
}

func (a *Army) HasSiegeUnits(types map[string]*UnitType) bool {
	return a.SiegeUnitScore(types) > 0
}

// ApplyWinterAttrition kış erozyonu — her birim %10 HP kaybeder.
func (a *Army) ApplyWinterAttrition() (lost int) {
	surviving := a.Units[:0]
	for _, u := range a.Units {
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

func (a *Army) HasDamagedUnits() bool {
	for _, u := range a.Units {
		if u.CurrentHP < MaxUnitHP {
			return true
		}
	}
	return false
}

func (a *Army) CanReplenishIn(regions map[world.RegionID]*world.Region) bool {
	if a == nil || a.IsNaval {
		return false
	}
	region := regions[a.RegionID]
	return region != nil && !region.IsSea && region.OwnerID == a.OwnerID
}

func (a *Army) ReplenishInFriendlyTerritory(regions map[world.RegionID]*world.Region, amount int) (healedUnits int) {
	if !a.CanReplenishIn(regions) || amount <= 0 {
		return 0
	}
	for i := range a.Units {
		if a.Units[i].CurrentHP >= MaxUnitHP {
			continue
		}
		a.Units[i].CurrentHP += amount
		if a.Units[i].CurrentHP > MaxUnitHP {
			a.Units[i].CurrentHP = MaxUnitHP
		}
		healedUnits++
	}
	return healedUnits
}

// ResetMovePoints tur başında hareket puanlarını sıfırlar.
func (a *Army) ResetMovePoints() {
	a.MovePoints = a.MaxMovePoints
	a.InAmbush = false
}

// LooksLikeGarrisonID legacy scenario/save verilerindeki garnizon ID desenlerini tanır.
func LooksLikeGarrisonID(id ArmyID) bool {
	name := strings.ToLower(string(id))
	return strings.HasPrefix(name, "army_garrison_") ||
		strings.HasSuffix(name, "_garrison") ||
		strings.Contains(name, "_garrison_")
}

// NormalizeLegacyGarrisons explicit state'i olmayan eski garnizon ordularını
// yeni `IsGarrison` alanına taşır.
func NormalizeLegacyGarrisons(armies map[ArmyID]*Army) {
	for id, a := range armies {
		if a == nil || a.IsNaval || a.IsGarrison {
			continue
		}
		if LooksLikeGarrisonID(id) {
			a.IsGarrison = true
		}
	}
}

// InitializeLegacyFleetDocking eski veri dosyalarında dock bilgisi olmayan
// filolar için komşu sahipli limandan başlangıç dock konumu türetir.
func InitializeLegacyFleetDocking(armies map[ArmyID]*Army, regions map[world.RegionID]*world.Region) {
	for _, a := range armies {
		if a == nil || !a.IsNaval || a.DockedRegionID != "" {
			continue
		}
		seaRegion := regions[a.RegionID]
		if seaRegion == nil || !seaRegion.IsSea {
			continue
		}
		for _, nid := range seaRegion.Neighbors {
			region := regions[nid]
			if region == nil || region.IsSea || region.OwnerID != a.OwnerID || !region.HasPortBuilding() {
				continue
			}
			for _, settlement := range region.Settlements {
				if settlement.Type != world.SettlementPort {
					continue
				}
				a.DockedRegionID = region.ID
				a.DockedSettlementID = settlement.ID
				break
			}
			if a.DockedRegionID != "" {
				break
			}
		}
	}
}
