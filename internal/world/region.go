package world

import "sort"

// RegionID bölge benzersiz kimliği.
type RegionID string

// SortedRegionIDs returns a sorted copy of region IDs without changing the
// order used by the runtime graph.
func SortedRegionIDs(ids []RegionID) []RegionID {
	result := append([]RegionID(nil), ids...)
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

// MaxTaxRate oyuncu ve AI için izin verilen azami bölgesel vergi oranıdır.
const MaxTaxRate = 60

// ClampTaxRate vergi oranını oyun kurallarındaki [0, MaxTaxRate] aralığına indirger.
func ClampTaxRate(rate int) int {
	if rate < 0 {
		return 0
	}
	if rate > MaxTaxRate {
		return MaxTaxRate
	}
	return rate
}

// Region harita üzerindeki tek bir bölgeyi temsil eder.
type Region struct {
	ID      RegionID    `json:"id"`
	Name    string      `json:"name"`
	NameTR  string      `json:"name_tr"`
	Terrain TerrainType `json:"terrain"`
	OwnerID string      `json:"owner_id"`
	// SuccessorFactionID, bölge fethedildikten sonra yeniden kurulabilecek
	// tarihsel devletin fraksiyon kimliğidir.
	SuccessorFactionID string     `json:"successor_faction_id,omitempty"`
	Neighbors          []RegionID `json:"neighbors"`
	// AreaNeighborOrder, regions.json yüklenirken görülen area:: komşularının
	// kaynak sırasını korur. Arazi alanları runtime'da yeniden üretildiği için
	// bu bilgi JSON'a yazılmaz; Edit Mode kaydında gereksiz sıra farklarını
	// önlemek için kullanılır.
	AreaNeighborOrder []RegionID `json:"-"`

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
	// IsMinorRegion, ana bir bölgenin içinde yer alan ve normal bölge gibi
	// fethedilebilen küçük stratejik alt bölgeyi işaretler. Hangi binaların
	// kullanılabileceği bina tanımlarındaki minor_regions alanından gelir.
	IsMinorRegion bool `json:"is_minor_region,omitempty"`
	// IsPrivileged, küçük alt bölgenin kullanım hakkının OwnerID'de kalırken
	// egemenliğinin ParentRegionID ile bağlı ana bölge sahibinden çözülmesini
	// sağlar. false olduğunda OwnerID hem kullanım hem egemenlik sahibidir.
	IsPrivileged bool `json:"is_privileged,omitempty"`
	// PrivilegeGrantedTurn, aktif imtiyazın verildiği kampanya turudur. Senaryo
	// başlangıcındaki imtiyazlar yükleyici tarafından 1. tur olarak damgalanır.
	// Bu alan senaryo export'una değil, kampanya save'ine aittir.
	PrivilegeGrantedTurn int `json:"-"`
	// IsTerrainArea marks a runtime child region painted inside a parent region.
	// These regions are navigable but have no economy or settlements.
	IsTerrainArea bool `json:"-"`
	// ParentRegionID, küçük alt bölgenin kalıcı ana bölgesini veya terrain
	// alanının runtime geometrik parent bilgisini taşır. Terrain alanları
	// scenario export'unda ayrıca filtrelendiği için bu alan normal alt
	// bölgelerde kalıcı, terrain runtime düğümlerinde geçicidir.
	ParentRegionID RegionID `json:"parent_region_id,omitempty"`
	TerrainAreaID  string   `json:"-"`

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
	TaxRate      int `json:"tax_rate"`     // 0-60 yüzde
	// Population kırsal nüfus ile yerleşim nüfuslarının toplamıdır.
	Population      int `json:"population"`
	RuralPopulation int `json:"rural_population"`

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
	ID         string         `json:"id"`
	Name       string         `json:"name,omitempty"`
	NameTR     string         `json:"name_tr"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	Type       SettlementType `json:"type,omitempty"`
	IsCenter   bool           `json:"is_center,omitempty"`
	Population int            `json:"population"`
}

// SettlementPositionClear, verilen koordinatın mevcut yerleşim marker'larına
// yeterince uzak olup olmadığını döner. Marker ölçüsü render katmanına ait
// olduğu için açıklık çağıran tarafından world-coordinate birimiyle verilir.
func (r *Region) SettlementPositionClear(x, y, clearance int) bool {
	if r == nil {
		return false
	}
	if clearance < 0 {
		clearance = 0
	}
	clearanceSquared := clearance * clearance
	for _, settlement := range r.Settlements {
		dx := settlement.X - x
		dy := settlement.Y - y
		if dx*dx+dy*dy < clearanceSquared {
			return false
		}
	}
	return true
}

const MinorRegionDefaultBaseGoldIncome = 10

// NewMinorRegionFromParent, Edit Mode'un yeni küçük alt bölge için kullandığı
// düşük gelirli başlangıç modelini üretir. Alt bölge normal Region olarak
// kaldığı için savaş, sahiplik ve gelir akışları tarafından işlenebilir.
func NewMinorRegionFromParent(id RegionID, parent *Region, worldX, worldY int) *Region {
	if parent == nil || parent.IsSea || parent.IsTerrainArea || id == "" {
		return nil
	}
	taxRate := parent.TaxRate
	if taxRate == 0 {
		taxRate = 45
	}
	return &Region{
		ID:             id,
		Name:           "New Minor Region",
		NameTR:         "Yeni Küçük Alt Bölge",
		Terrain:        parent.Terrain,
		OwnerID:        parent.OwnerID,
		Neighbors:      []RegionID{parent.ID},
		WorldX:         worldX,
		WorldY:         worldY,
		ShapeID:        parent.ShapeID,
		IsMinorRegion:  true,
		ParentRegionID: parent.ID,
		IsLocked:       parent.IsLocked,
		UnlockTurn:     parent.UnlockTurn,
		BaseGoldIncome: MinorRegionDefaultBaseGoldIncome,
		Satisfaction:   70,
		TaxRate:        taxRate,
		Religion:       parent.Religion,
		Buildings:      []string{},
		Settlements:    []Settlement{},
	}
}

// AllowsSettlementType, küçük alt bölge kuralını domain katmanında tutar.
// Normal bölgelerde mevcut tüm yerleşim tipleri geçerlidir.
func (r *Region) AllowsSettlementType(settlementType SettlementType) bool {
	if r == nil || r.IsSea {
		return false
	}
	if !r.IsMinorRegion {
		return true
	}
	return settlementType == SettlementFortress || settlementType == SettlementPort
}

// AllowsBuilding, bina tanımındaki minor_regions işaretine göre küçük alt
// bölgede inşaata izin verilip verilmediğini döner. Mevcut geçerli binalar
// korunur; eski kayıtlardaki geçersiz içerik bu helper tarafından geriye dönük
// silinmez.
func (r *Region) AllowsBuilding(buildingID string, minorRegions bool) bool {
	if r == nil || r.IsSea || buildingID == "" {
		return false
	}
	if !r.IsMinorRegion {
		return true
	}
	return minorRegions
}

// BuildingLevelCap, bina tanımındaki bölge tavanını bölgenin özel kurallarıyla
// birleştirir. İmtiyazlı minor bölgelerde minor_regions ile işaretli binalar
// senaryonun privileged_building_max_level değerine kadar geliştirilebilir;
// diğer bölgelerde senaryonun bina tanımı aynen korunur.
func (r *Region) BuildingLevelCap(buildingID string, configuredMax int, minorRegions bool, privilegedMaxLevel int) int {
	if r != nil && r.IsMinorRegion && r.IsPrivileged && minorRegions && privilegedMaxLevel > 0 {
		if configuredMax <= 0 || privilegedMaxLevel < configuredMax {
			return privilegedMaxLevel
		}
	}
	return configuredMax
}

// AllowedSettlementTypes, Edit Mode dropdown'ının bölgeye özel seçeneklerini
// üretir; UI settlement kuralını tekrar etmez.
func (r *Region) AllowedSettlementTypes() []SettlementType {
	if r != nil && r.IsMinorRegion {
		return []SettlementType{SettlementFortress, SettlementPort}
	}
	return AllSettlementTypeValues()
}

// PrimarySettlementIndex returns the center settlement index. Explicitly
// marked centers win; otherwise the fallback order is fortress, city, town,
// then port. Ties retain the source order.
func (r *Region) PrimarySettlementIndex() int {
	if r == nil || len(r.Settlements) == 0 {
		return -1
	}
	for i, settlement := range r.Settlements {
		if settlement.IsCenter {
			return i
		}
	}
	best := 0
	for i := 1; i < len(r.Settlements); i++ {
		if settlementPriority(r.Settlements[i].Type) > settlementPriority(r.Settlements[best].Type) {
			best = i
		}
	}
	return best
}

func settlementPriority(settlementType SettlementType) int {
	switch settlementType {
	case SettlementFortress:
		return 4
	case SettlementCity:
		return 3
	case SettlementTown:
		return 2
	case SettlementPort:
		return 1
	default:
		return 0
	}
}

// EnsurePrimarySettlement persists the fallback center on a region that has
// no explicit center. Existing center definitions are preserved.
func (r *Region) EnsurePrimarySettlement() bool {
	idx := r.PrimarySettlementIndex()
	if idx < 0 || r.Settlements[idx].IsCenter {
		return false
	}
	r.Settlements[idx].IsCenter = true
	return true
}

// SettlementPopulation bölgedeki yerleşimlerin toplam nüfusunu döner.
func (r *Region) SettlementPopulation() int {
	if r == nil || r.IsTerrainArea {
		return 0
	}
	total := 0
	for _, settlement := range r.Settlements {
		if settlement.Population > 0 {
			total += settlement.Population
		}
	}
	return total
}

// RecalculatePopulation, kırsal ve yerleşim nüfuslarını bölge toplamına bağlar.
// Eski senaryo/kayıt verisinde bileşenler bulunmuyorsa mevcut Population değeri
// geriye dönük uyumluluk için kırsal nüfus kabul edilir.
func (r *Region) RecalculatePopulation() int {
	if r == nil || r.IsTerrainArea {
		return 0
	}
	if r.RuralPopulation < 0 {
		r.RuralPopulation = 0
	}
	settlementPopulation := r.SettlementPopulation()
	if r.RuralPopulation == 0 && r.Population > settlementPopulation {
		r.RuralPopulation = r.Population - settlementPopulation
	}
	r.Population = r.RuralPopulation + settlementPopulation
	return r.Population
}

// AddPopulation büyümeyi kırsal nüfusa ekler; yerleşim nüfusları korunur.
func (r *Region) AddPopulation(amount int) {
	if r == nil || r.IsTerrainArea || amount <= 0 {
		return
	}
	settlementPopulation := r.SettlementPopulation()
	if r.RuralPopulation == 0 && r.Population > settlementPopulation {
		r.RuralPopulation = r.Population - settlementPopulation
	}
	r.RuralPopulation += amount
	r.RecalculatePopulation()
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

func (r *Region) BuildingLevel(buildingID string) int {
	if r == nil || r.IsSea || buildingID == "" {
		return 0
	}
	level := 0
	for _, id := range r.Buildings {
		if id == buildingID {
			level++
		}
	}
	return level
}

// BuildingLevels, bölgedeki tamamlanmış bina seviyelerini bina ID'sine göre
// döndürür. Aynı bina ID'sinin tekrarı seviye olarak sayılır.
func (r *Region) BuildingLevels() map[string]int {
	levels := make(map[string]int)
	if r == nil || r.IsSea {
		return levels
	}
	for _, buildingID := range r.Buildings {
		if buildingID != "" {
			levels[buildingID]++
		}
	}
	return levels
}

func (r *Region) HasFortressSettlement() bool {
	if r == nil || r.IsSea {
		return false
	}
	for _, settlement := range r.Settlements {
		if settlement.Type == SettlementFortress {
			return true
		}
	}
	return false
}

func (r *Region) FortificationLevel() int {
	if r == nil || r.IsSea {
		return 0
	}
	level := r.BuildingLevel("walls")
	if r.HasFortressSettlement() {
		level++
	}
	return level
}

func (r *Region) IsFortified() bool {
	return r.FortificationLevel() > 0
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
	if r == nil || r.IsTerrainArea {
		return 0
	}
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
	return r != nil && !r.IsTerrainArea && r.Satisfaction < 30
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
