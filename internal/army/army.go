package army

import "mapp-game-go/internal/world"

const MaxArmySize = 20

// ArmyID ordu benzersiz kimliği.
type ArmyID string

// Army harita üzerindeki bir orduyu temsil eder.
type Army struct {
	ID         ArmyID          `json:"id"`
	OwnerID    string          `json:"owner_id"` // fraksiyon ID
	RegionID   world.RegionID  `json:"region_id"`
	Units      []Unit          `json:"units"`
	MovePoints int             `json:"move_points"` // bu turda kalan hareket puanı
	MaxMovePoints int          `json:"max_move_points"`
	IsNaval    bool            `json:"is_naval"`  // deniz ordusu mu?

	// Pusu durumu: geçit bölgesinde bekliyorsa true
	InAmbush   bool `json:"in_ambush"`
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
		total += u.EffectiveAttack(types) + t.Morale/10
	}
	return total
}

// TotalDefense ordunun toplam savunma gücünü hesaplar.
func (a *Army) TotalDefense(types map[string]*UnitType) int {
	total := 0
	for _, u := range a.Units {
		total += u.EffectiveDefense(types)
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

// ResetMovePoints tur başında hareket puanlarını sıfırlar.
func (a *Army) ResetMovePoints() {
	a.MovePoints = a.MaxMovePoints
	a.InAmbush = false
}
