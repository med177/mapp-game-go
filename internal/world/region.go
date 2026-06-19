package world

// RegionID bölge benzersiz kimliği.
type RegionID string

// Region harita üzerindeki tek bir bölgeyi temsil eder.
type Region struct {
	ID        RegionID    `json:"id"`
	Name      string      `json:"name"`
	NameTR    string      `json:"name_tr"`
	Terrain   TerrainType `json:"terrain"`
	OwnerID   string      `json:"owner_id"`
	Neighbors []RegionID  `json:"neighbors"`

	// Dünya haritası koordinatları (renderer WorldW×WorldH dünya uzayı)
	WorldX int `json:"world_x"`
	WorldY int `json:"world_y"`

	// Natural Earth kaynaklı ülke sınırı ID'si (ISO_A3).
	ShapeID string `json:"shape_id,omitempty"`

	// Settlements bölge içindeki şehir/kasaba/kaleleri temsil eder.
	// X/Y koordinatları world_x/world_y ile aynı senaryo koordinat uzayındadır.
	Settlements []Settlement `json:"settlements,omitempty"`

	// Shape[poligon_idx][nokta_idx]
	Shape [][][2]float32 `json:"-"`

	// Deniz bölgesi mi? Oynanabilir kara bölgesi değildir.
	IsSea bool `json:"is_sea"`

	IsLocked   bool `json:"is_locked"`
	UnlockTurn int  `json:"unlock_turn"`

	// Ekonomi
	BaseGoldIncome   int `json:"base_gold_income"`
	BaseGrainOutput  int `json:"base_grain_output"`
	BaseIronOutput   int `json:"base_iron_output"`
	BaseTimberOutput int `json:"base_timber_output"`
	BaseStoneOutput  int `json:"base_stone_output"`
	BaseSpiceOutput  int `json:"base_spice_output"`
	BaseClothOutput  int `json:"base_cloth_output"`
	TradeCapacity    int `json:"trade_capacity"`

	// Durum
	Satisfaction int `json:"satisfaction"` // 0-100
	TaxRate      int `json:"tax_rate"`     // 0-100 yüzde
	Population   int `json:"population"`

	Religion string `json:"religion"`
	// ConversionTurns: sahip fraksiyon dini bölgeyle uyuşmuyorsa her tur artar.
	// 24 tura ulaşınca din değişir (~2 yıl).
	ConversionTurns int    `json:"conversion_turns,omitempty"`
	ActiveEventID   string `json:"active_event_id"`

	// İnşa edilmiş bina ID listesi
	Buildings []string `json:"buildings"`
}

type SettlementType string

const (
	SettlementCity     SettlementType = "city"
	SettlementTown     SettlementType = "town"
	SettlementFortress SettlementType = "fortress"
	SettlementPort     SettlementType = "port"
)

type SettlementTypeDef struct {
	Type   SettlementType
	NameTR string
}

var settlementTypeDefs = []SettlementTypeDef{
	{Type: SettlementCity, NameTR: "Şehir"},
	{Type: SettlementTown, NameTR: "Kasaba"},
	{Type: SettlementFortress, NameTR: "Kale"},
	{Type: SettlementPort, NameTR: "Liman"},
}

var settlementTypeDefsByValue = func() map[SettlementType]SettlementTypeDef {
	out := make(map[SettlementType]SettlementTypeDef, len(settlementTypeDefs))
	for _, def := range settlementTypeDefs {
		out[def.Type] = def
	}
	return out
}()

func AllSettlementTypes() []string {
	out := make([]string, 0, len(settlementTypeDefs))
	for _, def := range settlementTypeDefs {
		out = append(out, string(def.Type))
	}
	return out
}

func AllSettlementTypeValues() []SettlementType {
	out := make([]SettlementType, 0, len(settlementTypeDefs))
	for _, def := range settlementTypeDefs {
		out = append(out, def.Type)
	}
	return out
}

func (t SettlementType) LabelTR() string {
	if def, ok := settlementTypeDefsByValue[t]; ok {
		return def.NameTR
	}
	return string(t)
}

type Settlement struct {
	ID        string         `json:"id"`
	Name      string         `json:"name,omitempty"`
	NameTR    string         `json:"name_tr"`
	X         int            `json:"x"`
	Y         int            `json:"y"`
	Type      SettlementType `json:"type,omitempty"`
	IsCapital bool           `json:"is_capital,omitempty"`
}

// IsCoastal komşularda deniz olan kara bölgesiyse true döner.
func (r *Region) IsCoastal(allRegions map[RegionID]*Region) bool {
	if r.IsSea {
		return false
	}
	for _, nid := range r.Neighbors {
		if n, ok := allRegions[nid]; ok && n.IsSea {
			return true
		}
	}
	return false
}

// HasPort yerleşim tipi veya inşa edilmiş bina üzerinden bölgenin liman erişimi olup olmadığını döner.
func (r *Region) HasPort() bool {
	if r == nil || r.IsSea {
		return false
	}
	for _, settlement := range r.Settlements {
		if settlement.Type == SettlementPort {
			return true
		}
	}
	for _, buildingID := range r.Buildings {
		if buildingID == "port" {
			return true
		}
	}
	return false
}

func (r *Region) HasBuilding(buildingID string) bool {
	if r == nil || r.IsSea || buildingID == "" {
		return false
	}
	for _, id := range r.Buildings {
		if id == buildingID {
			return true
		}
	}
	return false
}

func (r *Region) HasPortBuilding() bool {
	return r.HasBuilding("port")
}

// CanNavalEnter bir naval ordunun bu bölgeye girebilip giremeyeceğini döner.
// Naval ordular sadece deniz bölgelerine girer.
func (r *Region) CanNavalEnter() bool {
	return r.IsSea
}

// CanLandEnter bir kara ordusunun bu bölgeye girebilip giremeyeceğini döner.
func (r *Region) CanLandEnter() bool {
	return !r.IsSea && !r.IsLocked
}
func (r *Region) GoldIncome() int {
	base := r.BaseGoldIncome * r.TaxRate / 100
	satisfactionMod := r.Satisfaction - 50
	adjusted := base + (base*satisfactionMod)/200
	if adjusted < 0 {
		return 0
	}
	return adjusted
}

// IsRebellionRisk isyan riski eşiğini kontrol eder.
func (r *Region) IsRebellionRisk() bool {
	return r.Satisfaction < 30
}

// ApplyConquest bölge el değiştirdiğinde memnuniyet ve sahiplik günceller.
// Farklı din → ekstra memnuniyet cezası.
func (r *Region) ApplyConquest(newOwnerID, newOwnerReligion string) {
	r.OwnerID = newOwnerID
	r.Satisfaction -= 10
	if newOwnerReligion != "" && newOwnerReligion != r.Religion {
		r.Satisfaction -= 15 // din farkı cezası
	}
	if r.Satisfaction < 0 {
		r.Satisfaction = 0
	}
}
