package economy

// GoodType ticari mal türü.
type GoodType string

const (
	GoodGrain  GoodType = "grain"
	GoodIron   GoodType = "iron"
	GoodTimber GoodType = "timber"
	GoodSpice  GoodType = "spice"
	GoodCloth  GoodType = "cloth"
)

// BaseGoldValue malların altın karşılığı (birim başına).
var BaseGoldValue = map[GoodType]int{
	GoodGrain:  2,
	GoodIron:   5,
	GoodTimber: 3,
	GoodSpice:  12,
	GoodCloth:  8,
}

// TradeRoute iki fraksiyon arasındaki aktif ticaret güzergahını tutar.
type TradeRoute struct {
	FromFactionID string   `json:"from_faction_id"`
	ToFactionID   string   `json:"to_faction_id"`
	Good          GoodType `json:"good"`
	AmountPerTurn int      `json:"amount_per_turn"`
	GoldPerUnit   int      `json:"gold_per_unit"` // anlaşmadaki fiyat
}

// GoldEarned bu güzergahtan tur başına altın kazancını döner.
func (t *TradeRoute) GoldEarned() int {
	return t.AmountPerTurn * t.GoldPerUnit
}

// TaxLevel vergi oranından memnuniyet etkisini hesaplar.
// Dönen değer: memnuniyet değişimi (negatif = düşüş).
func TaxSatisfactionDelta(taxRate int) int {
	switch {
	case taxRate <= 20:
		return 5  // çok düşük vergi → halk mutlu
	case taxRate <= 40:
		return 2
	case taxRate <= 60:
		return 0  // dengeli
	case taxRate <= 80:
		return -3
	default:
		return -8 // yüksek vergi → isyan riski
	}
}
