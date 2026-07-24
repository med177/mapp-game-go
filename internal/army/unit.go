package army

// UnitCategory birimin kategorisi.
type UnitCategory string

const (
	CategoryInfantry   UnitCategory = "infantry"
	CategoryCavalry    UnitCategory = "cavalry"
	CategorySiege      UnitCategory = "siege"
	CategoryNavalWar   UnitCategory = "naval_war"
	CategoryNavalTrans UnitCategory = "naval_trans"
	CategoryNavalTrade UnitCategory = "naval_trade"
)

// UnitTier birimin seviyesi (1=temel, 2=orta, 3=elit).
type UnitTier int

const MaxUnitHP = 100

// UnitType bir birim türünü tanımlar (JSON'dan yüklenir).
type UnitType struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	NameTR   string       `json:"name_tr"`
	Category UnitCategory `json:"category"`
	Tier     UnitTier     `json:"tier"`

	// MovementPoints bu birim tipinin tek başına taşıyabileceği tur başı hareket puanıdır.
	MovementPoints int `json:"movement_points"`

	// Savaş değerleri
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
	Morale  int `json:"morale"` // bozguna dayanıklılık
	HP      int `json:"hp"`     // başlangıç can puanı

	// Maliyet
	GoldCost      int `json:"gold_cost"`
	GrainCost     int `json:"grain_cost"`
	IronCost      int `json:"iron_cost"`
	TimberCost    int `json:"timber_cost"`
	StoneCost     int `json:"stone_cost"`
	SpiceCost     int `json:"spice_cost"`
	ClothCost     int `json:"cloth_cost"`
	GrainUpkeep   int `json:"grain_upkeep"` // tur başına bakım
	TurnsRequired int `json:"turns_required"`

	// Gereksinimler
	RequiredTech      []string `json:"required_tech"`       // tüm listedeki teknolojiler gerekir
	RequiredBldg      string   `json:"required_bldg"`       // gerekli bina ID
	RequiredBldgLevel int      `json:"required_bldg_level"` // 0/1 = Lv1, 2 = Lv2 ...

	// Denizde taşınabilir mi?
	Embarkable    bool `json:"embarkable"`
	CarryCapacity int  `json:"carry_capacity,omitempty"`
}

// RequiresTech birim tanımının belirli bir teknolojiyi istediğini bildirir.
func (t *UnitType) RequiresTech(techID string) bool {
	if t == nil || techID == "" {
		return false
	}
	for _, required := range t.RequiredTech {
		if required == techID {
			return true
		}
	}
	return false
}

// HasAllRequiredTechs birimin tüm teknoloji gereksinimlerinin tamamlanıp
// tamamlanmadığını kontrol eder. Liste AND semantiğine sahiptir.
func (t *UnitType) HasAllRequiredTechs(completed map[string]bool) bool {
	if t == nil {
		return false
	}
	for _, required := range t.RequiredTech {
		if required != "" && !completed[required] {
			return false
		}
	}
	return true
}

// MissingRequiredTechs tamamlanmamış teknoloji ID'lerini veri sırasıyla döner.
func (t *UnitType) MissingRequiredTechs(completed map[string]bool) []string {
	if t == nil {
		return nil
	}
	missing := make([]string, 0, len(t.RequiredTech))
	for _, required := range t.RequiredTech {
		if required != "" && !completed[required] {
			missing = append(missing, required)
		}
	}
	return missing
}

// BaseMovementPoints veri alanı olmayan eski birim tanımlarına geriye dönük
// güvenli varsayılan verir. Senaryo verileri bu alanı açıkça taşımalıdır.
func (t *UnitType) BaseMovementPoints() int {
	if t == nil {
		return 2
	}
	if t.MovementPoints > 0 {
		return t.MovementPoints
	}
	switch t.Category {
	case CategoryCavalry:
		return 3
	case CategorySiege:
		return 1
	case CategoryNavalWar, CategoryNavalTrans, CategoryNavalTrade:
		return 3
	default:
		return 2
	}
}

// Unit ordu içindeki tek bir birim örneğini temsil eder.
type Unit struct {
	TypeID     string `json:"type_id"`
	CurrentHP  int    `json:"current_hp"`
	Experience int    `json:"experience"` // 0-100, savaşlarla artar
}

func (u *Unit) HPPercent() float64 {
	hp := u.CurrentHP
	if hp < 0 {
		hp = 0
	}
	if hp > MaxUnitHP {
		hp = MaxUnitHP
	}
	return float64(hp) / float64(MaxUnitHP)
}

func (u *Unit) MissingHP() int {
	if u.CurrentHP >= MaxUnitHP {
		return 0
	}
	if u.CurrentHP < 0 {
		return MaxUnitHP
	}
	return MaxUnitHP - u.CurrentHP
}

// EffectiveAttack deneyim bonusunu dahil eder.
func (u *Unit) EffectiveAttack(types map[string]*UnitType) int {
	t, ok := types[u.TypeID]
	if !ok {
		return 0
	}
	bonus := u.Experience / 20 // her 20 XP için +1
	return t.Attack + bonus
}

// EffectiveDefense deneyim bonusunu dahil eder.
func (u *Unit) EffectiveDefense(types map[string]*UnitType) int {
	t, ok := types[u.TypeID]
	if !ok {
		return 0
	}
	bonus := u.Experience / 20
	return t.Defense + bonus
}
