---
type: system
tags: [economy, gold, tax, trade, buildings]
last_updated: 2026-07-22
related: [systems/seasons, systems/events, systems/ai, systems/combat, world/regions, architecture/game-loop, architecture/state-management]
---

# Ekonomi Sistemi

**Kaynak:** `internal/economy/economy.go`, `internal/economy/resources.go`, `internal/city/building.go`

## Kaynaklar

| Kaynak | Tür | Açıklama |
|---|---|---|
| Düka Altın | Birincil | Her şey altına çevrilir |
| Tahıl | İkincil | Ordu besleme, kıtlık riski |
| Demir | İkincil | Ordu kalitesi |
| Kereste | İkincil | Bina inşa |
| Taş | İkincil | Kuşatma/bina reçeteleri |
| Baharat | İkincil | Ticaret geliri |
| Kumaş | İkincil | Ticaret geliri |

Altın ve ikincil kaynaklar birlikte kullanılır; birim/bina üretiminde çoklu kaynak reçetesi zorunludur.

Kaynak adları ve fraksiyon alan eşlemeleri `internal/economy/resources.go` içinde `ResourceKind`/`ResourceDef` modeliyle merkezileştirilmiştir. UI metinleri, ticaret malları listesi ve `ResourceCost` formatlaması bu ortak tanımları kullanır; böylece `Altın/Tahıl/Demir/...` stringleri farklı paketlerde ayrı ayrı hardcode edilmez.

---

## Vergi Sistemi

Her bölgede `TaxRate` (0–100) ayarlanabilir.

Oyuncu: `.` tuşu +5, `,` tuşu -5 → `adjustTax()` — `internal/game/game.go:557`

| Vergi Oranı | Etkisi |
|---|---|
| Düşük (0–30) | Yüksek memnuniyet, isyan riski düşük |
| Orta (30–60) | Dengeli |
| Yüksek (60–100) | Fazla altın, memnuniyet düşer, isyan riski |

**İsyan:** `checkRebellions()` memnuniyet eşiğini kontrol eder → bölge kontrolü kaybedilebilir.

---

## Bina Gelir Etkileri

`assets/data/buildings.json`

| Bina | Tuş | Gelir Etkisi |
|---|---|---|
| Pazar (`market`) | 1 | +altın geliri |
| Çiftlik (`farm`) | 2 | +tahıl üretimi |
| Demirci (`barracks`) | 3 | +ordu eğitim hızı |
| Liman (`port`) | 4 | +deniz birimi, +ticaret |
| Surlar (`walls`) | 5 | +savunma bonusu |
| Tapınak/Kilise/Cami (`temple`) | 6 | +din etkisi, +memnuniyet |

Bina inşası `city.LoadBuildings()` ile yüklenen altın + kaynak reçetesini ister (`grain/iron/timber/stone_cost`).
Bina `MaxPerRegion` ile sınırlıdır.
Bazı binalar `RequiredTerrain` kısıtı taşır (ör. liman → kıyı).

---

## Ticaret Güzergahları

`TradeRoute` — `internal/economy/economy.go`

```go
type TradeRoute struct {
    FromFactionID string   `json:"from_faction_id"`
    ToFactionID   string   `json:"to_faction_id"`
    Good          GoodType `json:"good"`
    AmountPerTurn int      `json:"amount_per_turn"`
    GoldPerUnit   int      `json:"gold_per_unit"`
}
```

Ticaret anlaşması kurulunca aktif olur. `ApplyTradeRoutes()` her tur:
1. Kaynak fraksiyondan **mal çıkar** (yetersizse rota atlanır)
2. Hedef fraksiyona **mal ekler**
3. Hedef fraksiyondan **altın çıkar** (yetersizse rota atlanır)
4. Kaynak fraksiyona **altın ekler**

Hedef fraksiyonun `StrategicGrainDemand` değeri üç aylık güvenli rezerv hedefine kalan açığı gösterir. Yeni rota kurulurken hedefte bu sinyal pozitif, kaynakta `StrategicGrainSurplus` pozitifse kaynak-hedef rotası tahıl malına yönlendirilir; böylece ithalat mevcut rota transferi üzerinden gerçekleşir. Abluka rota hacmini azaltabilir; kaynakta stok veya hedefte altın yetersizse rota o tur çalışmaz.

→ Diplomasi anlaşmaları: [[systems/diplomacy]]

## Dinamik Piyasa Fiyatlandırması

`ComputeMarketPrices()` her tur sonu tüm fraksiyonların stoklarına göre fiyatları günceller:

- **Arz artışı → fiyat düşer** (bol mal değersizleşir)
- **Arz azalışı → fiyat yükselir** (kıt mal pahalanır)
- Fiyat sınırları: basePrice × %25 (min) – basePrice × %300 (max)
- Her aktif fraksiyon varsayılan talep üretir (10 birim/mal); tahıl için fraksiyonların stratejik rezerv açığı da ek talep sinyali olarak fiyat hesabına dahil edilir.

Mevcut fiyatlar `GameState.MarketPrices`'ta tutulur (serialize edilmez, her tur yeniden hesaplanır).

## Pasif Ticaret Geliri

### Merchant gemisi rota katkısı (1300)

`TradeRoute` içindeki `MerchantAmountBonus` save'e yazılmayan runtime alanıdır.
`GameState.RefreshMerchantTradeBonuses()` her ekonomi çözümünden önce merchant
filolarını yeniden değerlendirir:

- Merchant gemisi aktif yönlü rotaya `+1 AmountPerTurn` ekler; rota başına üst sınır `+2`dir.
- Filo, rotanın uç fraksiyonlarından birine ait olmalı ve en az bir uçta aktif kıyısal
  trade center bulunmalıdır. İki uçta da merkez varsa tarihsel link grafiği bağlantısı
  aranır; filo merkez komşusu denizde değilse bonus verilmez.
- Askıdaki rota `ApplyTradeRoutes()` tarafından atlanır. Kaynak mal ya da hedef altın
  yetersizse merchant katkısı bedava gelir üretmez; rota o tur gerçekleşmez.
- AI merchant görevi `Army.TradeRouteKey` ile kalıcıdır; rota anahtarı `gönderen->alan`
  yönünü korur ve save/load sonrası yeniden bağlanabilir.

Her bölgenin `TradeCapacity` değerine göre pasif ticaret geliri hesaplanır:

```
tradeIncome = TradeCapacity × 2 × goldMod
```

Pazar (`gold_mod: 1.5`) ve Liman (`gold_mod: 1.3`) gibi binalar bu geliri artırır.

## Üretim Reçeteleri ve Lojistik

- Birim üretimi `gold_cost` yanında `grain_cost`, `iron_cost`, `timber_cost`, `stone_cost` tüketir.
- Ordu bakımında `grain_upkeep` temel alınır. Sabit ordu `%100`, o tur hareket etmiş ordu `%150`, garnizon `%75`, kuşatma saldırganı `%200`, kuşatma savunmacısı/destekçisi `%125` tüketir. Hareket bilgisi `ArmyMoveUsage` runtime-only yakalanır; save formatına alan eklenmez.
- Dost toprakta mevcut ücretsiz toparlanmaya ek olarak, ekonomi tick'inde yalnız `StorageCapacity` üzerindeki tahıl kara ordusu yenilemesine ayrılabilir. Her 1 HP yenileme 1 tahıl harcar; ordu başına/turuna en fazla `+10 HP` verilir. Ordular faction/army ID sırasıyla işlenir, kuşatma altındaki savunmacılar ve düşman toprakları kapsam dışıdır. Rezerv kapasitesi korunur; harcanan tahıl ve yenilenen HP `GrainEconomyStatus` içinde raporlanır.
- Pozitif nüfuslu her sahip bölge `ceil(population / 20)` tahıl/tur sivil tüketim üretir. Bu tüketim fraksiyonun ortak tahıl havuzundan, bölge üretimi ve ticaret girdilerinden sonra, ordu bakımı uygulanmadan önce düşülür. `Population <= 0` olan legacy/test bölgeleri tüketim oluşturmaz.
- Aktif hasat/kıtlık/kuraklık olayları `RegionEventStatus` içindeki geçici yüzde modifiyerleriyle tahıl üretimini ve sivil tüketimi etkiler. Etki; ekonomi tick'i, bölgesel ordu lojistiği, stratejik ithalat talebi ve AI değerlendirmesinde ortak state yardımcıları üzerinden uygulanır; olay süresi bitince normal değerler geri gelir.
- 1300 Faz 6 denge raporu erken/orta/savaş pencerelerinde büyük fraksiyonların üretim/sivil talep oranını `1.0–4.0`, net değişim/sivil talep oranını `-1.0–2.5` bandında doğrular. Kıtlık oranı ayrıca raporlanır; erken Osmanlı ve Venedik gibi ithalat baskısı yaşayan profillerde negatif dönem kabul edilir ve stratejik tahıl talebi üretir.
- `GameState.GrainEconomy` runtime snapshot'ı fraksiyon başına üretim, sivil talep, ordu bakımı, net değişim, stok ve stokun kaç ay yeteceğini taşır. Toplam talebin 3 aydan az stoğa oranı uyarı, 1 aydan azı kritik, mevcut tur talebi karşılanamadığında kıtlık sayılır.
- Uyarı seviyesinde gelir %5 ve memnuniyet 1, kritik seviyede gelir %10 ve memnuniyet 2, kıtlık seviyesinde gelir %25 ve memnuniyet 4 azalır. Gerçek stok açığı varsa mevcut ordu HP cezası yalnız o tick'in hesaplanan açık miktarı kadar uygulanır; ordu sırası deterministiktir.
- Tahıl arzı ordunun kalıcı moraline de bağlanır: stabil seviyede her ekonomi tick'inde `+1`, uyarı/kritik/kıtlık seviyelerinde sırasıyla `-1/-3/-6` uygulanır. Moral `1–100` arasında tutulur; 100 moral nötr, 50 moral yaklaşık `%15` toplam savaş gücü kaybı üretir. Gerçekleşen toplam değişim `GrainEconomyStatus.ArmyMoraleDelta` ile HUD/event detayına taşınır ve uygulama Army ID sırasıyla deterministiktir.
- Depolama kapasitesi `6 × sivil talep + 3 × ordu bakımı` olarak hesaplanır; talep varsa minimum kapasite 100'dür. Kapasite üstündeki stok her ekonomi tick'inde fazlanın %2'si oranında, en az 1 tahıl olacak şekilde bozulur. `StorageCapacity` ve `Spoiled` runtime snapshot alanlarıdır; save migration gerektirmez.
- `granary` / `Ambar` binası her kurulu seviye için +100 tahıl depolama kapasitesi verir. Bina tüm senaryolarda veri tanımı olarak bulunur; özel sprite yoksa mevcut çiftlik sprite'ı görsel fallback olarak kullanılır.
- Kara orduları ayrıca bölge bazlı ikmal kapasitesine tabidir. Yerel askeri kapasite, bölge üretiminden önce sivil talep düşüldükten sonraki fazlalık + yerleşim/ticaret tamponu + fraksiyon stokundan sınırlı destek olarak hesaplanır. Yabancı/düşman bölgede yerel üretim desteği yoktur. Efektif ordu talebi kapasiteyi karşılamazsa aynı bölgede bekleyen ordular turdan tura artan HP zayiatı alır. AI hareket, geri çekilme, birleşme ve bina yatırımındaki lojistik tahminlerde `GameState.EffectiveArmyGrainUpkeep()` kullanır.
- Limanlı bölgenin komşu denizinde düşman savaş gemisi varsa abluka oluşur: savaş gemisi başına ilgili ticaret rotası hacmi ve limanın yerleşim/rezerv ikmal tamponu %50 azalır; iki veya daha fazla gemi rotayı/tamponu tamamen keser. Abluka yüzdesi runtime state'ten her ekonomi tick'inde yeniden türetilir.
- Kasım ekonomi tick'inde arz seviyesi stabil olan fraksiyon, yalnızca depolama kapasitesini aşan tahılını nüfus yatırımında kullanır. Memnuniyeti en az 60 olan, isyan riski taşımayan ve kuşatma altında olmayan bölgelerde nüfusun %1'i (minimum 1) büyütülür; her 1 nüfus artışı 2 tahıl harcar. Bu harcama `GrainEconomyStatus` içinde runtime raporlanır; sonraki tick'lerde artan nüfus `ceil(population / 20)` sivil talebe dönüşerek büyümeyi aynı tahıl döngüsüne bağlar ve rezerv kapasite tabanı korunur.
- Bölge panelindeki `Tahıl Yardımı` aksiyonu kendi bölgesine 12 tahıl aktararak memnuniyeti +10 artırır; bölge başına turda bir kez kullanılabilir. Kuşatma altındaki, oyuncuya ait olmayan, zaten yüksek memnuniyetli veya yetersiz stoklu bölgeler yardım alamaz. Tur ilerlediğinde yardım kullanım kilidi runtime olarak sıfırlanır.
- Pazar ekranındaki `ACİL TAHIL SAT` aksiyonu ticaret partneri gerektirmez. Yalnızca fraksiyonun `StorageCapacity` üstündeki tahıl satılabilir; satış fiyatı güncel tahıl piyasa fiyatının %70'idir ve minimum 1 altın/tahıl olarak hesaplanır. Böylece acil nakit ihtiyacı karşılanırken savaş rezervi korunur.
- Pazar sekmesindeki `Oto. İhracat` toggle'ı açıkken ekonomi tick'i kapasite üstü tahılı aktif, savaşta olmayan ticaret ağı partnerlerine otomatik satar. Partnerler faction ID sırasıyla işlenir, fiyat güncel tahıl piyasa fiyatının %60'ıdır ve alıcının altını yetersizse satış miktarı düşürülür. Tercih `GameState.AutoGrainExport` olarak save'e yazılır; `GrainEconomyStatus` gerçekleşen ihracat tahılını ve altınını runtime raporlar.
- Üretim emri iptalinde altınla birlikte diğer kaynaklar da iade edilir.

## Bölge Uzmanlaşması

- Ova: tahıl üretimi artar.
- Orman: kereste üretimi artar.
- Dağ/geçit: demir ve taş üretimi artar.
- `base_stone_output` olmayan senaryolarda dağ/geçit bölgeleri fallback taş üretimi sağlar.

## UI Üretim Önizlemesi

`GameState.RegionProductionSummary()` seçili bir bölgenin efektif altın ve mal katkısını UI için hesaplar.

- Vergi + memnuniyet bazlı altın geliri, bina çarpanları ve mevsimsel hasat/ticaret modları uygulanır.
- Tahıl/demir/kereste/taş/baharat/kumaş üretimi arazi uzmanlaşması sonrası gösterilir.
- Sahip fraksiyonun ekonomi teknolojileri varsa aynı önizlemeye dahil edilir.
- Bölge bilgi paneli bu helper ile beslendiği için görünen üretim satırları ekonomi çözüm mantığıyla daha yakındır.

## Tek Seferlik Mal Transferi

`TransferGoods()` dinamik piyasa fiyatını kullanarak iki fraksiyon arasında anlık takas yapar.
Kullanım senaryosu: diplomasi panelinde oyuncunun elindeki malları satması.

---

## Sonbahar Gelir Bonusu

Sonbahar aylarında (9, 10, 11) `applyEconomyTick()` gelir çarpanı uygular.

→ [[systems/seasons]]

---

## Eksik / Planlanan

- [x] İkincil mal üretim/tüketim döngüsü
- [x] Piyasa fiyatı dalgalanması
- [x] Nüfus bazlı temel sivil tahıl tüketimi (`ceil(population / 20)`)
- [x] Tahıl stok-ay göstergesi ve kademeli kıtlık görünürlüğü
- [x] Temel stok kapasitesi, kapasite üstü bozulma ve HUD `stok / kapasite` görünümü
- [x] Kıtlık mekaniği (tahıl sıfırlandığında lojistik ceza)
- [x] Kıtlıkta nüfus artışı durur; arz seviyesine göre ordu morali etkisi uygulanır
- [x] Ambar/depo binalarından kapasite bonusu
- [x] Liman ticareti ve bölgesel ikmal için düşman savaş gemisi ablukası
- [x] Kademeli sivil kıtlık eşikleri ve stok ay görünürlüğü
- [x] Tahıl depolama kapasitesi ve bozulma
- [x] Stabil rezerv fazlasını yıllık nüfus büyümesine bağla
- [x] Bölge bazlı tahıl yardımıyla memnuniyet/isyan rahatlatma
- [x] Depo kapasitesi üzerindeki tahıl için acil piyasa satışı
- [x] Düşük fiyatlı otomatik tahıl ihracatı
- [x] Hasat/kıtlık/kuraklık olaylarının üretim ve tüketim modeline bağlanması; mevcut düşman savaş gemisi tabanlı abluka kesintisinin bölgesel ikmal ve ticaret akışında korunması
- [ ] Tahılın diğer alternatif harcama alanları
- [ ] Ekonomik zafer sayacı (500 altın/tur × 5 tur) tam bağlantısı
