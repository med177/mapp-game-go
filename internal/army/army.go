package army

import "mapp-game-go/internal/world"

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

	// Pusu durumu: geçit bölgesinde bekliyorsa true
	InAmbush bool `json:"in_ambush"`

	// Bölgesel ikmal kapasitesi aşımı üst üste kaç tur sürdü.
	OverCapacityTurns int `json:"over_capacity_turns,omitempty"`
}

// Size ordu boyutunu döner.
func (a *Army) Size() int {
	return len(a.Units)
}

// CanAddUnit yeni birim eklenebilir mi?
func (a *Army) CanAddUnit() bool {
	return len(a.Units) < MaxArmySize
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
			if region == nil || region.IsSea || region.OwnerID != a.OwnerID {
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
