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

// UnitType bir birim türünü tanımlar (JSON'dan yüklenir).
type UnitType struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	NameTR   string       `json:"name_tr"`
	Category UnitCategory `json:"category"`
	Tier     UnitTier     `json:"tier"`

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
	GrainUpkeep   int `json:"grain_upkeep"` // tur başına bakım
	TurnsRequired int `json:"turns_required"`

	// Gereksinimler
	RequiredTech      string `json:"required_tech"`       // "" = gerek yok
	RequiredBldg      string `json:"required_bldg"`       // gerekli bina ID
	RequiredBldgLevel int    `json:"required_bldg_level"` // 0/1 = Lv1, 2 = Lv2 ...

	// Denizde taşınabilir mi?
	Embarkable bool `json:"embarkable"`
}

// Unit ordu içindeki tek bir birim örneğini temsil eder.
type Unit struct {
	TypeID     string `json:"type_id"`
	CurrentHP  int    `json:"current_hp"`
	Experience int    `json:"experience"` // 0-100, savaşlarla artar
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
