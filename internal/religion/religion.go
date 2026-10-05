package religion

import (
	"encoding/json"
	"fmt"
	"os"
)

// Type fraksiyon dini.
type Type string

// Eski test ve sentetik state fixture'ları için kullanılan yaygın din kimlikleri.
// Aktif senaryoların din listesi religions.json içinden yüklenir.
const (
	Catholic   Type = "catholic"
	Orthodox   Type = "orthodox"
	Sunni      Type = "sunni"
	Shia       Type = "shia"
	Tengri     Type = "tengri"
	Pagan      Type = "pagan"
	Protestant Type = "protestant"
	Utraquist  Type = "utraquist"
)

// Def tek bir senaryo dininin metadata'sını tanımlar.
type Def struct {
	Type      Type   `json:"id"`
	Name      string `json:"name,omitempty"`
	NameTR    string `json:"name_tr"`
	Group     string `json:"group,omitempty"`
	SpriteSet string `json:"sprite_set,omitempty"`
}

// RelationDef iki din arasındaki senaryo bazlı ilişki katsayılarını tanımlar.
type RelationDef struct {
	ReligionA     Type   `json:"religion_a"`
	ReligionB     Type   `json:"religion_b"`
	Score         *int   `json:"relation_score,omitempty"`
	AllianceBonus *int   `json:"alliance_bonus,omitempty"`
	InitialStance string `json:"initial_stance,omitempty"`
}

type fileDefinition struct {
	Religions []Def         `json:"religions"`
	Relations []RelationDef `json:"relations,omitempty"`
}

type relationKey struct {
	a Type
	b Type
}

// Registry aktif senaryonun din tanımlarını ve ilişkilerini tutar.
type Registry struct {
	defs       []Def
	defsByType map[Type]Def
	relations  map[relationKey]RelationDef
}

var defaultDefs = []Def{
	{Type: Catholic, NameTR: "Katolik"},
	{Type: Orthodox, NameTR: "Ortodoks"},
	{Type: Sunni, NameTR: "Sünni İslam", SpriteSet: "eastern"},
	{Type: Shia, NameTR: "Şii İslam", SpriteSet: "eastern"},
	{Type: Tengri, NameTR: "Tengricilik"},
	{Type: Pagan, NameTR: "Paganlık"},
	{Type: Protestant, NameTR: "Protestanlık"},
	{Type: Utraquist, NameTR: "Utrakvistlik"},
}

var defaultRelations = []RelationDef{
	{ReligionA: Sunni, ReligionB: Shia, Score: intPtr(-40), AllianceBonus: intPtr(-8)},
	{ReligionA: Catholic, ReligionB: Orthodox, Score: intPtr(-20), AllianceBonus: intPtr(2)},
	{ReligionA: Catholic, ReligionB: Protestant, Score: intPtr(-20)},
	{ReligionA: Protestant, ReligionB: Utraquist, Score: intPtr(-20)},
}

func intPtr(value int) *int { return &value }

// NewRegistry doğrulanmış din tanımlarından runtime registry oluşturur.
func NewRegistry(defs []Def, relations []RelationDef) (*Registry, error) {
	if len(defs) == 0 {
		return nil, fmt.Errorf("religions JSON din listesi boş")
	}
	registry := &Registry{
		defs:       make([]Def, 0, len(defs)),
		defsByType: make(map[Type]Def, len(defs)),
		relations:  make(map[relationKey]RelationDef, len(relations)),
	}
	for _, def := range defs {
		if def.Type == "" {
			return nil, fmt.Errorf("religion id boş olamaz")
		}
		if _, exists := registry.defsByType[def.Type]; exists {
			return nil, fmt.Errorf("tekrarlanan religion id: %q", def.Type)
		}
		if def.NameTR == "" {
			def.NameTR = def.Name
		}
		if def.NameTR == "" {
			def.NameTR = string(def.Type)
		}
		registry.defs = append(registry.defs, def)
		registry.defsByType[def.Type] = def
	}
	for _, relation := range relations {
		if relation.ReligionA == "" || relation.ReligionB == "" || relation.ReligionA == relation.ReligionB {
			return nil, fmt.Errorf("geçersiz religion ilişkisi: %q/%q", relation.ReligionA, relation.ReligionB)
		}
		if _, ok := registry.defsByType[relation.ReligionA]; !ok {
			return nil, fmt.Errorf("religion ilişkisi tanımsız din içeriyor: %q", relation.ReligionA)
		}
		if _, ok := registry.defsByType[relation.ReligionB]; !ok {
			return nil, fmt.Errorf("religion ilişkisi tanımsız din içeriyor: %q", relation.ReligionB)
		}
		key := makeRelationKey(relation.ReligionA, relation.ReligionB)
		if _, exists := registry.relations[key]; exists {
			return nil, fmt.Errorf("tekrarlanan religion ilişkisi: %q/%q", relation.ReligionA, relation.ReligionB)
		}
		registry.relations[key] = relation
	}
	return registry, nil
}

// Load senaryonun data/religions.json dosyasını okur.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("religions dosyası okunamadı: %w", err)
	}
	var definition fileDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, fmt.Errorf("religions JSON parse hatası: %w", err)
	}
	return NewRegistry(definition.Religions, definition.Relations)
}

// DefaultRegistry sentetik state ve paket testleri için mevcut varsayılanları döner.
func DefaultRegistry() *Registry {
	defs := append([]Def(nil), defaultDefs...)
	relations := append([]RelationDef(nil), defaultRelations...)
	registry, err := NewRegistry(defs, relations)
	if err != nil {
		panic(err)
	}
	return registry
}

func makeRelationKey(a, b Type) relationKey {
	if a > b {
		return relationKey{a: b, b: a}
	}
	return relationKey{a: a, b: b}
}

// All senaryoda tanımlı dinleri JSON sırasıyla döner.
func (r *Registry) All() []Type {
	if r == nil {
		return nil
	}
	out := make([]Type, 0, len(r.defs))
	for _, def := range r.defs {
		out = append(out, def.Type)
	}
	return out
}

// DisplayNameTR dinin senaryodaki Türkçe adını döner.
func (r *Registry) DisplayNameTR(t Type) string {
	if r != nil {
		if def, ok := r.defsByType[t]; ok {
			return def.NameTR
		}
	}
	return string(t)
}

// Next editör seçiminde sonraki senaryo dinini döner.
func (r *Registry) Next(current Type) Type {
	options := r.All()
	for i, option := range options {
		if option == current {
			return options[(i+1)%len(options)]
		}
	}
	if len(options) == 0 {
		return ""
	}
	return options[0]
}

// Relation iki din arasındaki başlangıç ilişki puanını döner.
func (r *Registry) Relation(a, b Type) int {
	if a == b && a != "" {
		return 25
	}
	if r != nil {
		if relation, ok := r.relations[makeRelationKey(a, b)]; ok && relation.Score != nil {
			return *relation.Score
		}
	}
	return -30
}

// AllianceBonus din farkının ittifak teklifine etkisini döner.
func (r *Registry) AllianceBonus(a, b Type) int {
	if a == b && a != "" {
		return 8
	}
	if r != nil {
		if relation, ok := r.relations[makeRelationKey(a, b)]; ok && relation.AllianceBonus != nil {
			return *relation.AllianceBonus
		}
	}
	return -4
}

// InitialStance din çifti için editörün yeni ilişki varsayılanını döner.
func (r *Registry) InitialStance(a, b Type) string {
	if r != nil {
		if relation, ok := r.relations[makeRelationKey(a, b)]; ok {
			return relation.InitialStance
		}
	}
	return ""
}

// SpriteSet dinin birlik sprite grubunu döner.
func (r *Registry) SpriteSet(t Type) string {
	if r != nil {
		if def, ok := r.defsByType[t]; ok {
			return def.SpriteSet
		}
	}
	return ""
}

// ValidateTypes kullanılan faction ve bölge dinlerinin tanımlı olduğunu doğrular.
func (r *Registry) ValidateTypes(types []Type) error {
	if r == nil {
		return fmt.Errorf("religion registry yok")
	}
	for _, typ := range types {
		if typ == "" {
			continue
		}
		if _, ok := r.defsByType[typ]; !ok {
			return fmt.Errorf("tanımsız religion: %q", typ)
		}
	}
	return nil
}

// Paket düzeyi fonksiyonlar sentetik fixture'ların registry oluşturmadan
// çalışmasını sağlar. Aktif oyun state'i kendi registry'sini kullanır.
var defaultRegistry = DefaultRegistry()

func All() []Type                 { return defaultRegistry.All() }
func DisplayNameTR(t Type) string { return defaultRegistry.DisplayNameTR(t) }
func Next(current Type) Type      { return defaultRegistry.Next(current) }
func Relation(a, b Type) int      { return defaultRegistry.Relation(a, b) }
func AllianceBonus(a, b Type) int { return defaultRegistry.AllianceBonus(a, b) }
