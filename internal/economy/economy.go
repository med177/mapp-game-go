package economy

import (
	"mapp-game-go/internal/faction"
)

// GoodType ticari mal türü.
type GoodType string

const (
	GoodGrain  GoodType = "grain"
	GoodIron   GoodType = "iron"
	GoodTimber GoodType = "timber"
	GoodStone  GoodType = "stone"
	GoodSpice  GoodType = "spice"
	GoodCloth  GoodType = "cloth"
)

// BaseGoldValue malların altın karşılığı (birim başına) — referans fiyat.
var BaseGoldValue = map[GoodType]int{
	GoodGrain:  2,
	GoodIron:   5,
	GoodTimber: 3,
	GoodStone:  4,
	GoodSpice:  12,
	GoodCloth:  8,
}

// EmergencySalePricePercent acil pazar satışında güncel fiyatın uygulanan oranıdır.
const EmergencySalePricePercent = 70

// AutomaticExportPricePercent otomatik tahıl ihracatında uygulanan düşük fiyat oranıdır.
const AutomaticExportPricePercent = 60

// EmergencySaleUnitPrice düşük fiyatlı doğrudan pazar satışının birim fiyatını döner.
func EmergencySaleUnitPrice(marketPrice int) int {
	if marketPrice <= 0 {
		return 0
	}
	price := marketPrice * EmergencySalePricePercent / 100
	if price < 1 {
		price = 1
	}
	return price
}

// AutomaticExportUnitPrice otomatik ihracatın düşük birim fiyatını döner.
func AutomaticExportUnitPrice(marketPrice int) int {
	if marketPrice <= 0 {
		return 0
	}
	price := marketPrice * AutomaticExportPricePercent / 100
	if price < 1 {
		price = 1
	}
	return price
}

// CurrentMarketPrice tur başı güncellenen dinamik piyasa fiyatlarını tutar.
// Fiyatlar, tüm fraksiyonların toplam stok miktarına göre dalgalanır.
type CurrentMarketPrice map[GoodType]int

// ComputeMarketPrices arz-talep dengesine göre tüm malların güncel fiyatını hesaplar.
// formül: fiyat = basePrice * (1 + (talepFaktörü - arzFaktörü) / arzFaktörü)
// Arz arttıkça fiyat düşer, arz azaldıkça fiyat yükselir.
// Minimum fiyat basePrice'ın %25'i, maksimum %300'ü.
func ComputeMarketPrices(factions map[faction.FactionID]*faction.Faction) CurrentMarketPrice {
	return ComputeMarketPricesWithStrategicDemand(factions, nil)
}

// ComputeMarketPricesWithStrategicDemand, kıtlık yaşayan fraksiyonların
// ithalat ihtiyacını tahıl fiyatına ek talep olarak yansıtır. demandByFaction
// değerleri fraksiyonun üç aylık rezerv hedefine kalan tahıl açığıdır.
func ComputeMarketPricesWithStrategicDemand(
	factions map[faction.FactionID]*faction.Faction,
	demandByFaction map[faction.FactionID]int,
) CurrentMarketPrice {
	prices := make(CurrentMarketPrice, len(BaseGoldValue))

	// Toplam arzı hesapla (tüm fraksiyonların stokları)
	totalSupply := map[GoodType]int{}
	for _, f := range factions {
		if f == nil || f.IsEliminated {
			continue
		}
		for _, good := range TradeGoods() {
			totalSupply[good] += getGoodAmount(f, good)
		}
	}

	// Aktif fraksiyon sayısı (talep göstergesi)
	activeFactions := 0
	for _, f := range factions {
		if f != nil && !f.IsEliminated {
			activeFactions++
		}
	}
	if activeFactions < 1 {
		activeFactions = 1
	}

	for good, basePrice := range BaseGoldValue {
		supply := totalSupply[good]
		if supply <= 0 {
			supply = 1 // sıfıra bölmeyi önle
		}

		// Talep faktörü: her aktif fraksiyon ~10 birim talep eder (varsayılan)
		demandFactor := float64(activeFactions * 10)
		if good == GoodGrain {
			for _, demand := range demandByFaction {
				if demand > 0 {
					// Sinyali yumuşat: her 5 tahıl açık talebi fiyat talebine
					// bir birim ekler; fiyat yine mevcut alt/üst sınırlara tabidir.
					demandFactor += float64(demand / 5)
				}
			}
		}
		supplyFactor := float64(supply)

		// Fiyat = basePrice * (demandFactor / supplyFactor)
		ratio := demandFactor / supplyFactor
		price := int(float64(basePrice) * ratio)

		// Fiyat sınırları
		minPrice := basePrice / 4
		if minPrice < 1 {
			minPrice = 1
		}
		maxPrice := basePrice * 3

		if price < minPrice {
			price = minPrice
		}
		if price > maxPrice {
			price = maxPrice
		}

		prices[good] = price
	}

	return prices
}

// TradeRoute iki fraksiyon arasındaki aktif ticaret güzergahını tutar.
type TradeRoute struct {
	FromFactionID  string   `json:"from_faction_id"`
	ToFactionID    string   `json:"to_faction_id"`
	Good           GoodType `json:"good"`
	AmountPerTurn  int      `json:"amount_per_turn"`
	GoldPerUnit    int      `json:"gold_per_unit"`   // anlaşmadaki sabit fiyat (dinamik değil)
	SuspendedTurns int      `json:"suspended_turns"` // korsan/olay nedeniyle kaç tur askıda (0=aktif)

	// MerchantAmountBonus save'e yazılmaz; her ekonomi çözümünde filoların
	// aktif görevi ve gerçek deniz konumundan yeniden hesaplanır.
	MerchantAmountBonus int `json:"-"`

	// BlockadePercent düşman savaş filolarının rota hacmine uyguladığı runtime
	// kesintidir; save'e yazılmaz ve her ekonomi çözümünde yeniden hesaplanır.
	BlockadePercent int `json:"-"`
}

const MaxMerchantAmountBonusPerRoute = 2
const MaxTradeRouteBlockadePercent = 100

// TradeRouteAssignmentKey rota yeniden üretildiğinde de değişmeyen görev anahtarıdır.
func TradeRouteAssignmentKey(fromFactionID, toFactionID string) string {
	if fromFactionID == "" || toFactionID == "" || fromFactionID == toFactionID {
		return ""
	}
	return fromFactionID + "->" + toFactionID
}

func (t *TradeRoute) AssignmentKey() string {
	if t == nil {
		return ""
	}
	return TradeRouteAssignmentKey(t.FromFactionID, t.ToFactionID)
}

// EffectiveAmountPerTurn merchant filosu katkısı dahil gerçek rota hacmidir.
func (t *TradeRoute) EffectiveAmountPerTurn() int {
	if t == nil {
		return 0
	}
	bonus := t.MerchantAmountBonus
	if bonus < 0 {
		bonus = 0
	}
	if bonus > MaxMerchantAmountBonusPerRoute {
		bonus = MaxMerchantAmountBonusPerRoute
	}
	amount := t.AmountPerTurn + bonus
	blockade := t.BlockadePercent
	if blockade < 0 {
		blockade = 0
	}
	if blockade > MaxTradeRouteBlockadePercent {
		blockade = MaxTradeRouteBlockadePercent
	}
	return amount * (MaxTradeRouteBlockadePercent - blockade) / MaxTradeRouteBlockadePercent
}

// GoldEarned bu güzergahtan tur başına altın kazancını döner (satan taraf için).
func (t *TradeRoute) GoldEarned() int {
	return t.EffectiveAmountPerTurn() * t.GoldPerUnit
}

// ApplyTradeRoutes tüm aktif ticaret rotalarını bir tur işletir.
// Kaynak fraksiyondan mal çıkar, hedef fraksiyona mal ekler.
// Hedef fraksiyondan altın çıkar, kaynak fraksiyona altın ekler.
// Yetersiz mal veya altın durumunda rota o tur için atlanır.
func ApplyTradeRoutes(factions map[faction.FactionID]*faction.Faction, routes []*TradeRoute) []string {
	var logs []string

	for _, tr := range routes {
		if tr == nil || tr.SuspendedTurns > 0 {
			continue
		}
		srcFaction := factions[faction.FactionID(tr.FromFactionID)]
		dstFaction := factions[faction.FactionID(tr.ToFactionID)]

		if srcFaction == nil || dstFaction == nil {
			continue
		}
		if srcFaction.IsEliminated || dstFaction.IsEliminated {
			continue
		}

		// Kaynak fraksiyonda yeterli mal var mı?
		amount := tr.EffectiveAmountPerTurn()
		if amount <= 0 {
			continue
		}
		available := getGoodAmount(srcFaction, tr.Good)
		if available < amount {
			logs = append(logs, tr.FromFactionID+" yetersiz "+string(tr.Good)+" — ticaret rotası atlandı")
			continue
		}

		// Hedef fraksiyonda yeterli altın var mı?
		totalCost := tr.GoldEarned()
		if dstFaction.Gold < totalCost {
			logs = append(logs, tr.ToFactionID+" yetersiz altın — ticaret rotası atlandı")
			continue
		}

		// Mal transferi: kaynaktan çıkar, hedefe ekle
		addGoodAmount(srcFaction, tr.Good, -amount)
		addGoodAmount(dstFaction, tr.Good, amount)

		// Altın transferi: hedeften çıkar, kaynağa ekle
		dstFaction.Gold -= totalCost
		srcFaction.Gold += totalCost
	}

	return logs
}

// getGoodAmount fraksiyonun belirli bir maldan kaç birime sahip olduğunu döner.
func getGoodAmount(f *faction.Faction, good GoodType) int {
	if kind, ok := GoodToResourceKind(good); ok {
		return FactionResourceAmount(f, kind)
	}
	return 0
}

// addGoodAmount fraksiyonun belirli bir malını amount kadar artırır/azaltır.
func addGoodAmount(f *faction.Faction, good GoodType, amount int) {
	if kind, ok := GoodToResourceKind(good); ok {
		AddFactionResource(f, kind, amount)
	}
}

// TransferGoods iki fraksiyon arasında tek seferlik mal takası yapar.
// fromID'den toID'ye amount kadar mal gider, toID'den fromID'ye gold gider.
// Dinamik piyasa fiyatını kullanır.
func TransferGoods(
	factions map[faction.FactionID]*faction.Faction,
	fromID, toID faction.FactionID,
	good GoodType,
	amount int,
	prices CurrentMarketPrice,
) bool {
	src := factions[fromID]
	dst := factions[toID]
	if src == nil || dst == nil {
		return false
	}

	available := getGoodAmount(src, good)
	if available < amount {
		return false
	}

	price := prices[good]
	return TransferGoodsAtUnitPrice(factions, fromID, toID, good, amount, price)
}

// TransferGoodsAtUnitPrice iki fraksiyon arasında belirli birim fiyatla mal takası yapar.
// Otomatik ihracat gibi piyasa dışı indirimli işlemler bu ortak transfer yolunu kullanır.
func TransferGoodsAtUnitPrice(
	factions map[faction.FactionID]*faction.Faction,
	fromID, toID faction.FactionID,
	good GoodType,
	amount, price int,
) bool {
	src := factions[fromID]
	dst := factions[toID]
	if src == nil || dst == nil || amount <= 0 || price <= 0 {
		return false
	}
	totalCost := amount * price
	if getGoodAmount(src, good) < amount || dst.Gold < totalCost {
		return false
	}

	addGoodAmount(src, good, -amount)
	addGoodAmount(dst, good, amount)
	dst.Gold -= totalCost
	src.Gold += totalCost
	return true
}

// TaxLevel vergi oranından memnuniyet etkisini hesaplar.
// Dönen değer: memnuniyet değişimi (negatif = düşüş).
func TaxSatisfactionDelta(taxRate int) int {
	switch {
	case taxRate <= 20:
		return 5 // çok düşük vergi → halk mutlu
	case taxRate <= 40:
		return 2
	case taxRate <= 60:
		return 0 // dengeli
	case taxRate <= 80:
		return -3
	default:
		return -8 // yüksek vergi → isyan riski
	}
}

// RegionTradeIncome bir bölgenin ticaret kapasitesine göre pasif ticaret gelirini hesaplar.
// TradeCapacity değeri kullanılır, goldMod çarpanı (pazar/liman binalarından gelir) uygulanır.
// goldMod: 1.0 = normal, 1.5 = pazar bonusu, 1.3 = liman bonusu vb.
func RegionTradeIncome(tradeCapacity int, tradeCapMod float64) int {
	baseTradeIncome := tradeCapacity * 2 // her birim kapasite 2 altın
	if baseTradeIncome < 0 {
		baseTradeIncome = 0
	}
	return int(float64(baseTradeIncome) * tradeCapMod)
}

// ── Korsanlık ──────────────────────────────────────────────────────────

// ApplyPirateRaids rastgele korsan baskını olayları üretir.
// Aktif ticaret rotalarını askıya alır. Deniz bölgelerinde filo olan rotalar korunur.
// Dönen değer: oyuncuya gösterilecek olay mesajları.
// baseChance: 0.0-1.0 arası temel olasılık (varsayılan 0.08 = %8).
// pirateMod: mevsimsel çarpan (season.PirateMod).
func ApplyPirateRaids(
	factions map[faction.FactionID]*faction.Faction,
	routes []*TradeRoute,
	baseChance float64,
	pirateMod int,
	armiesInSea func(faction.FactionID) int, // callback: deniz bölgelerindeki ordu sayısı
) []string {
	var logs []string
	if baseChance <= 0 {
		baseChance = 0.08
	}
	chance := baseChance * float64(pirateMod) / 100.0

	for _, tr := range routes {
		if tr.SuspendedTurns > 0 {
			tr.SuspendedTurns--
			continue
		}
		// Rastgele korsan baskını
		// Burada uint64 kullanmıyoruz, basit bir olasılık kontrolü yapıyoruz
		// Gerçek rastgelelik game loop tarafından sağlanır
		_ = chance

		// Deniz güvenliği kontrolü: hedef veya kaynak fraksiyonun deniz bölgelerinde ordusu varsa korunur
		srcNaval := armiesInSea(faction.FactionID(tr.FromFactionID))
		dstNaval := armiesInSea(faction.FactionID(tr.ToFactionID))
		if srcNaval > 0 || dstNaval > 0 {
			continue // filo koruması var
		}

		// Basit olasılık (fraksiyon sayısına göre normalize)
		activeCount := 0
		for _, f := range factions {
			if f != nil && !f.IsEliminated {
				activeCount++
			}
		}
		if activeCount <= 0 {
			activeCount = 1
		}
		// Her rota için aktif fraksiyon başına %2 temel risk
		risk := 0.02 * float64(activeCount) * float64(pirateMod) / 100.0
		if risk > 0.3 {
			risk = 0.3 // max %30
		}

		// risk olasılığıyla rota 4-8 tur askıya alınır
		// (gerçek rastgelelik için game loop tarafından kontrol edilir)
		_ = risk

		// Askıya alma mantığı resolution.go'da rastgelelik ile yapılacak
	}

	return logs
}

// SuspendRoute bir ticaret rotasını geçici olarak askıya alır.
func SuspendRoute(tr *TradeRoute, turns int) {
	if tr == nil || turns <= 0 {
		return
	}
	tr.SuspendedTurns = turns
}

// ── Kaynak Dönüşümü (İmalathane) ──────────────────────────────────────

// ProcessedGoodType işlenmiş mal türleri.
type ProcessedGoodType string

const (
	ProcessedArmor  ProcessedGoodType = "armor"  // Demir → Zırh (savaş bonusu)
	ProcessedShip   ProcessedGoodType = "ship"   // Kereste → Gemi (donanma inşası)
	ProcessedLuxury ProcessedGoodType = "luxury" // Kumaş → Lüks mal (yüksek fiyat)
)

// ProcessGoodReq bir işlenmiş mal için gereken ham mal ve miktar.
type ProcessGoodReq struct {
	InputGood  GoodType          `json:"input_good"`
	InputQty   int               `json:"input_qty"` // gereken ham mal miktarı
	OutputType ProcessedGoodType `json:"output_type"`
	OutputQty  int               `json:"output_qty"` // üretilen işlenmiş mal miktarı
	BuildReq   string            `json:"build_req"`  // gereken bina ID'si
}

// ProcessGoods varsayılan dönüşüm tarifleri.
var ProcessGoods = []ProcessGoodReq{
	{InputGood: GoodIron, InputQty: 2, OutputType: ProcessedArmor, OutputQty: 1, BuildReq: "workshop"},
	{InputGood: GoodTimber, InputQty: 3, OutputType: ProcessedShip, OutputQty: 1, BuildReq: "workshop"},
	{InputGood: GoodCloth, InputQty: 2, OutputType: ProcessedLuxury, OutputQty: 1, BuildReq: "workshop"},
}

// ProcessedGoodValue işlenmiş malların piyasa değeri.
var ProcessedGoodValue = map[ProcessedGoodType]int{
	ProcessedArmor:  20,
	ProcessedShip:   15,
	ProcessedLuxury: 30,
}

// ApplyProduction bir bölgedeki imalathanede ham malları işlenmiş ürüne dönüştürür.
// Başarılı olursa true döner.
func ApplyProduction(f *faction.Faction, req ProcessGoodReq, regionHasBuilding func(string) bool) bool {
	if f == nil {
		return false
	}
	// Bina kontrolü
	if !regionHasBuilding(req.BuildReq) {
		return false
	}
	// Yeterli hammadde var mı?
	switch req.InputGood {
	case GoodIron:
		if f.Iron < req.InputQty {
			return false
		}
		f.Iron -= req.InputQty
	case GoodTimber:
		if f.Timber < req.InputQty {
			return false
		}
		f.Timber -= req.InputQty
	case GoodCloth:
		if f.Cloth < req.InputQty {
			return false
		}
		f.Cloth -= req.InputQty
	default:
		return false
	}
	// İşlenmiş malı ekle (şimdilik altına çeviriyoruz)
	value := ProcessedGoodValue[req.OutputType] * req.OutputQty
	f.Gold += value
	return true
}
