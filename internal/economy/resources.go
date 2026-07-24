package economy

import "mapp-game-go/internal/faction"

// ResourceKind oyundaki ortak kaynak/ticaret kalemlerini tek model altında toplar.
type ResourceKind string

const (
	ResourceGold   ResourceKind = "gold"
	ResourceGrain  ResourceKind = "grain"
	ResourceIron   ResourceKind = "iron"
	ResourceTimber ResourceKind = "timber"
	ResourceStone  ResourceKind = "stone"
	ResourceSpice  ResourceKind = "spice"
	ResourceCloth  ResourceKind = "cloth"
)

// ResourceDef bir kaynağın görünen adı ve oyun içi eşlemelerini taşır.
type ResourceDef struct {
	Kind        ResourceKind
	NameTR      string
	LowerNameTR string
	TradeGood   GoodType
}

var resourceDefs = []ResourceDef{
	{Kind: ResourceGold, NameTR: "Altın", LowerNameTR: "altın"},
	{Kind: ResourceGrain, NameTR: "Tahıl", LowerNameTR: "tahıl", TradeGood: GoodGrain},
	{Kind: ResourceIron, NameTR: "Demir", LowerNameTR: "demir", TradeGood: GoodIron},
	{Kind: ResourceTimber, NameTR: "Kereste", LowerNameTR: "kereste", TradeGood: GoodTimber},
	{Kind: ResourceStone, NameTR: "Taş", LowerNameTR: "taş", TradeGood: GoodStone},
	{Kind: ResourceSpice, NameTR: "Baharat", LowerNameTR: "baharat", TradeGood: GoodSpice},
	{Kind: ResourceCloth, NameTR: "Kumaş", LowerNameTR: "kumaş", TradeGood: GoodCloth},
}

var resourceDefsByKind = func() map[ResourceKind]ResourceDef {
	m := make(map[ResourceKind]ResourceDef, len(resourceDefs))
	for _, def := range resourceDefs {
		m[def.Kind] = def
	}
	return m
}()

var resourceKindByGood = func() map[GoodType]ResourceKind {
	m := make(map[GoodType]ResourceKind, len(resourceDefs))
	for _, def := range resourceDefs {
		if def.TradeGood != "" {
			m[def.TradeGood] = def.Kind
		}
	}
	return m
}()

func AllResourceKinds() []ResourceKind {
	return []ResourceKind{
		ResourceGold,
		ResourceGrain,
		ResourceIron,
		ResourceTimber,
		ResourceStone,
		ResourceSpice,
		ResourceCloth,
	}
}

// CostResourceKinds inşa/üretim maliyetlerinde kullanılan kaynak sırasını döner.
func CostResourceKinds() []ResourceKind {
	return []ResourceKind{
		ResourceGold,
		ResourceGrain,
		ResourceIron,
		ResourceTimber,
		ResourceStone,
		ResourceSpice,
		ResourceCloth,
	}
}

// TradeGoods ticaret ekranında dolaşılan mal sırasını döner.
func TradeGoods() []GoodType {
	return []GoodType{
		GoodGrain,
		GoodIron,
		GoodTimber,
		GoodStone,
		GoodSpice,
		GoodCloth,
	}
}

func ResourceDefByKind(kind ResourceKind) ResourceDef {
	if def, ok := resourceDefsByKind[kind]; ok {
		return def
	}
	return ResourceDef{Kind: kind, NameTR: string(kind), LowerNameTR: string(kind)}
}

func ResourceNameTR(kind ResourceKind) string {
	return ResourceDefByKind(kind).NameTR
}

func ResourceLowerNameTR(kind ResourceKind) string {
	return ResourceDefByKind(kind).LowerNameTR
}

func GoodToResourceKind(good GoodType) (ResourceKind, bool) {
	kind, ok := resourceKindByGood[good]
	return kind, ok
}

func GoodNameTR(good GoodType) string {
	if kind, ok := GoodToResourceKind(good); ok {
		return ResourceNameTR(kind)
	}
	return string(good)
}

func GoodLowerNameTR(good GoodType) string {
	if kind, ok := GoodToResourceKind(good); ok {
		return ResourceLowerNameTR(kind)
	}
	return string(good)
}

func ResourceInvalidCountMessageTR(kind ResourceKind) string {
	return ResourceNameTR(kind) + " sayısı geçersiz."
}

func FormatResourceAmountTR(kind ResourceKind, amount int) string {
	return ResourceNameTR(kind) + " " + itoa(amount)
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func FactionResourceAmount(f *faction.Faction, kind ResourceKind) int {
	if f == nil {
		return 0
	}
	switch kind {
	case ResourceGold:
		return f.Gold
	case ResourceGrain:
		return f.Grain
	case ResourceIron:
		return f.Iron
	case ResourceTimber:
		return f.Timber
	case ResourceStone:
		return f.Stone
	case ResourceSpice:
		return f.Spice
	case ResourceCloth:
		return f.Cloth
	default:
		return 0
	}
}

func AddFactionResource(f *faction.Faction, kind ResourceKind, amount int) {
	if f == nil {
		return
	}
	switch kind {
	case ResourceGold:
		f.Gold += amount
	case ResourceGrain:
		f.Grain += amount
	case ResourceIron:
		f.Iron += amount
	case ResourceTimber:
		f.Timber += amount
	case ResourceStone:
		f.Stone += amount
	case ResourceSpice:
		f.Spice += amount
	case ResourceCloth:
		f.Cloth += amount
	}
}
