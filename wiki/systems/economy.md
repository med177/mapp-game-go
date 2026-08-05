---
type: system
tags: [economy, gold, tax, trade, buildings]
last_updated: 2026-08-05
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

Altın ve ikincil kaynaklar birlikte kullanılır; birim/bina üretiminde çoklu kaynak reçetesi zorunludur. `ResourceCost` artık altın, tahıl, demir, kereste, taş, baharat ve kumaşın tamamını affordability, ödeme, iade ve tooltip akışına taşır (`internal/economy/costs.go`, `resources.go`).

Kaynak adları ve fraksiyon alan eşlemeleri `internal/economy/resources.go` içinde `ResourceKind`/`ResourceDef` modeliyle merkezileştirilmiştir. UI metinleri, ticaret malları listesi ve `ResourceCost` formatlaması bu ortak tanımları kullanır; böylece `Altın/Tahıl/Demir/...` stringleri farklı paketlerde ayrı ayrı hardcode edilmez.

Devletin tur başı efektif üretimi `GameState.FactionProductionSummary()` ile bölge bazlı üretimlerden toplanır; kuşatma altındaki bölgeler üretime katkı vermez. Tahıl HUD değeri ayrıca `FactionGrainNetChange()` ile sivil talep ve ordu bakımını düşerek net stok değişimini gösterir.

Düşman toprağında bekleyen ordu `Yağmala` görevini verdiğinde `GameState.Raids`
aynı bölgeyi o tur için tek kez işaretler. Ekonomi tick'i hedefin efektif vergi
gelirinin `%80`'ini ve üretimlerinin `%50`'sini hedef devletten düşüp yağmalayan
fraksiyona aktarır; işlem sonrasında raid kaydı temizlenir. Aktif kuşatma yapan
ordu yağmalayamaz. `RaidLootPreview()` aynı hesabı görev rozetinin hover
tooltip'ine taşır; `RaidState.RaiderArmyID` yağma kazancını doğru marker'a
bağlar (`internal/state/raid_ambush.go`, `internal/game/resolution.go`,
`internal/render/army_task_status.go`).

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
| Pazar (`market`) | 1 | +altın geliri, tur başı memnuniyet `+1` |
| Çiftlik (`farm`) | 2 | +tahıl üretimi, tur başı memnuniyet `+1` |
| Kışla (`barracks`) | 3 | +ordu eğitim hızı, seviye başına tur başı memnuniyet `-1` |
| Liman (`port`) | 4 | +deniz birimi, +ticaret, tur başı memnuniyet `+1` |
| Surlar (`walls`) | 5 | +savunma bonusu, tur başı memnuniyet `+6` |
| Tapınak/Kilise/Cami (`temple`) | 6 | +din etkisi, tur başı memnuniyet `+10` |
| Ambar (`granary`) | — | +tahıl depolama kapasitesi |

Bina inşası `city.LoadBuildings()` ile yüklenen altın + kaynak reçetesini ister (`grain/iron/timber/stone/spice/cloth_cost`). Pazar, liman ve ibadet yeri gibi ticaret/kültür yapıları baharat veya kumaş tüketebilir; temel tarım ve savunma yapıları bölgesel hammaddelere dayanır.
Bina `MaxPerRegion` ile sınırlıdır.
Bazı binalar `RequiredTerrain` kısıtı taşır (ör. liman → kıyı).

Oyuncu, bölge panelindeki tamamlanmış bina seviyelerinden birini kırmızı `X`
düğmesiyle yıkabilir. Yıkım ortak onay modalından sonra `Game.demolishBuilding()`
ile uygulanır; kaynak iadesi yapılmaz, devam eden yükseltme kuyruğu varken yıkım
engellenir. Port binasının otomatik oluşturduğu liman yerleşimi, son port seviyesi
de kaldırıldığında temizlenir; senaryo ile tanımlı port yerleşimleri korunur.

1300 Osmanlı yükselişi senaryosunda `farm` üretim çarpanı `x1.30`, bölge başına üst sınırı 3 seviyedir. Bina çarpanları ekonomi tick'inde birlikte uygulanır; bu nedenle üç farm seviyesi güçlü bir tarım yatırımıdır ancak denge testindeki `1.0–4.0` üretim/sivil talep bandını aşmamalıdır.

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

AI ayrıca `internal/ai/grain_procurement.go` üzerinden aktif rota grafiğine bağlı
devletlerden doğrudan tahıl satın alır. Bu tamamlayıcı alım, rota başına sabit transfer
beklemek yerine üç aylık kapasite hedefindeki açığı iki aylık pencereyle kapatır; tedarikçi
devletin güvenli rezervi korunur ve alıcının altını acil rezervin altına düşürülmez.

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
- Filo, rotanın uç fraksiyonlarından birine ait olmalı ve yönlü liman çiftinin hedef
  denizinde bulunmalıdır. Gemlik → Özi gibi rotalarda hedef deniz, Özi tarafındaki
  gerçek liman denizidir; bağlı tüm denizler aynı anda geçerli sayılmaz.
- Askıdaki rota `ApplyTradeRoutes()` tarafından atlanır. Kaynak mal ya da hedef altın
  yetersizse merchant katkısı bedava gelir üretmez; rota o tur gerçekleşmez.
- AI merchant görevi `Army.TradeRouteKey` ile kalıcıdır; rota anahtarı `gönderen->alan`
  yönünü korur ve save/load sonrası yeniden bağlanabilir.
- Oyuncu seçili merchant filosundaki `ROTA ATA` düğmesiyle aynı geçerli rota listesinden
  görev seçebilir veya görevi kaldırabilir. Atama `SetMerchantTradeRoute()` ile doğrulanır;
  filo hedef denizde değilse görev kayıtlı kalır ancak merchant bonusu filo doğru denize
  ulaşana kadar uygulanmaz. Aktif hedef denizde bonus kazanan merchant gemileri kış
  attrition'ından muaftır; aynı filodaki savaş ve nakliye gemileri normal kış hasarı alır.
  AI tarafı aynı `TradeRouteKey` modelini otomatik rota seçimi ve deniz hareketiyle kullanır.

Merchant rotası olmayan tarihsel merkez bağlantılı anlaşmalar panelden gizlenmez; aktif
limanlar arasında deniz yolu bulunuyorsa `MerchantTradeRoutePortPairs()` bu liman çiftini
üretir. Haritanın `Ticaret` modunda oyuncuya ait aktif anlaşmalar, seçilen liman çiftleri
arasında tek turuncu renkli, eşit uzunlukta çizgi/boşluklardan oluşan kesikli koridor ve liman uçlarındaki işaretlerle gösterilir
(`internal/state/merchant_trade.go`, `internal/render/trade_overlay.go`).

Her bölgenin ham `TradeCapacity` değeri, `GameState.EffectiveRegionTradeCapacity()`
ile önce bina çarpanları ve aktif ticaret merkezi bonusuyla ortak efektif değere
dönüştürülür. Aynı değer pasif gelir, diplomasi kapasitesi, rota hacmi, AI
ekonomi değerlendirmesi ve ticaret UI'ında kullanılır:

```
effectiveCapacity = round(baseCapacity × binaTradeCapacityModlari) + merkezHacimBonusu
tradeIncome = (effectiveCapacity × 2 + merkezGümrüğü) × mevsim × abluka × marketTeknolojisi
```

Pazar (`trade_capacity_mod: 1.45`) ve liman (`1.60`) ana büyütücülerdir. Ambar
(`1.05`) ile ibadethane (`1.03`) küçük fakat gerçek ticaret katkısı sağlar;
tekrar eden bina seviyeleri çarpanları üst üste uygular. 1300 ve 1455 senaryoları
primary ticaret merkezinin hacim 50'ye kadar tabanı `+2 kapasite / +4 altın`,
secondary merkezin tabanı `+1 kapasite / +2 altın`dır. Hacim 50'yi aştığında
primary merkez her 25, secondary merkez her 50 hacimde sınırsız olarak `+1
kapasite / +2 altın` büyür. Hacim, bina sonrası yerel kapasite ile aktif rota
hacimlerinin toplamıdır; merkez bonusu kendi hacmini yeniden büyütmez. Bu
doğrudan gelir mevsim, abluka ve pazar teknolojisi çarpanlarından geçer; dolayısıyla
merkezin kuşatılması/ablukası avantajı da azaltır. Tüm bonuslar bölgenin sahibi
değiştiğinde otomatik olarak yeni devlete geçer. Eski `trade_centers.json`
dosyalarında bu root alanlar yoksa loader aynı `+2/+1` ve `+4/+2` kalibrasyonunu
geriye uyumlu varsayılan olarak uygular.

1300 başlangıç grafiğinde her oynanabilir devlet en az bir kendi ticaret merkezi
düğümüne sahiptir. Osmanlı `Bithynia → Konstantiniyye`, Aragon `Katalonya →
Ceneviz/Portekiz`, HRE `Hollanda → Flandre/Danimarka` ve Moskova `Moskova →
Novgorod` bağlantılarıyla ana ağa girer. Tarihsel merkez linkleri çift yönlü
saklanır; böylece data/export denetimi ile görsel rota grafiği ayrışmaz.
Norveç (Oslo bölgesi) `Danimarka` üzerinden; İsveç (Stockholm bölgesi) ise
`Danimarka` ve `Novgorod` üzerinden Kuzey Denizi-Baltık ticaret ağına bağlanır.
Fas `Cezayir/Portekiz`, Girit `Venedik/Konstantiniyye/Mısır` ve Rodos
`Konstantiniyye/Mısır` koridorlarıyla Akdeniz ve Atlantik ağına katılır.

Efektif kapasite, aktif dış ticaret anlaşmaları arasında gerçek bir ortak havuzdur.
`RebalanceTradeRouteCapacities()` her rota kurulumu, yükleme temizliği ve ekonomi
tick'inde devletin kapasitesini partnerlerine eşit paylaştırır; bölünmeyen kalan
birimler faction ID sırasındaki ilk anlaşmalara gider. İki yönlü rotanın temel
`AmountPerTurn` değeri iki tarafın paylarından düşük olanıdır ve anlaşma başına
en fazla `4` kalır. Böylece daha fazla anlaşma imzalamak tek tek rotaları
zayıflatır; bina, fetih veya merkez bonusu kapasiteyi artırdığında rota hacmi
bir sonraki tick'te yeniden büyür. Merchant bonusu bu temel kapasite havuzunu
tüketmez.

## Üretim Reçeteleri ve Lojistik

- Birim ve bina inşası `gold_cost` yanında `grain_cost`, `iron_cost`, `timber_cost`, `stone_cost`, `spice_cost` ve `cloth_cost` tüketebilir. Binalardaki tahıl maliyeti, inşaatta çalışan işçilerin iaşesini temsil eder ve emir kuyruğa alındığında peşin düşülür; emir iptal edilirse diğer kaynaklarla birlikte iade edilir. 1300/1455 ham verisinde bina tahıl maliyetleri yaklaşık `%20` artırılarak inşaatın sivil tahıl rezervi üzerindeki küçük baskısı korunmuştur. Temel kara birlikleri tahıl/demir eksenini korurken elit birlikler ve deniz birlikleri kumaş; ticaret gemileri ayrıca baharat kullanır.
- Ordu bakımında `grain_upkeep` temel alınır. Sabit ordu `%100`, o tur hareket etmiş ordu `%150`, garnizon `%75`, kuşatma saldırganı `%200`, kuşatma savunmacısı/destekçisi `%125` tüketir. Hareket bilgisi `ArmyMoveUsage` runtime-only yakalanır; save formatına alan eklenmez.
- Dost toprakta mevcut ücretsiz toparlanmaya ek olarak, ekonomi tick'inde yalnız `StorageCapacity` üzerindeki tahıl kara ordusu yenilemesine ayrılabilir. Her 1 HP yenileme 1 tahıl harcar; ordu başına/turuna en fazla `+10 HP` verilir. Ordular faction/army ID sırasıyla işlenir, kuşatma altındaki savunmacılar ve düşman toprakları kapsam dışıdır. Rezerv kapasitesi korunur; harcanan tahıl ve yenilenen HP `GrainEconomyStatus` içinde raporlanır.
- Pozitif nüfuslu her sahip bölge, kırsal nüfus ile yerleşim nüfuslarının toplamı olan `Population` üzerinden `ceil(population / 18)` tahıl/tur sivil tüketim üretir. Bu tüketim fraksiyonun ortak tahıl havuzundan, bölge üretimi ve ticaret girdilerinden sonra, ordu bakımı uygulanmadan önce düşülür. `Population <= 0` olan legacy/test bölgeleri tüketim oluşturmaz. Böylece yerleşim nüfusu eksik veya kırsal köyler göz ardı edilmeden, 1300 senaryosunda barışta küçük rezerv ve savaşta belirgin tahıl açığı oluşur.
- Aktif hasat/kıtlık/kuraklık olayları `RegionEventStatus` içindeki geçici yüzde modifiyerleriyle tahıl üretimini ve sivil tüketimi etkiler. Etki; ekonomi tick'i, bölgesel ordu lojistiği, stratejik ithalat talebi ve AI değerlendirmesinde ortak state yardımcıları üzerinden uygulanır; olay süresi bitince normal değerler geri gelir.
- 1300 tahıl denge raporu erken/orta/savaş pencerelerinde büyük fraksiyonların üretim/sivil talep oranını `0.75–4.0`, net değişim/sivil talep oranını `-1.0–2.5` bandında doğrular. Kıtlık oranı ayrıca raporlanır; erken Osmanlı ve Venedik gibi ithalat baskısı yaşayan profillerde negatif dönem kabul edilir ve stratejik tahıl talebi üretir.
- `GameState.GrainEconomy` runtime snapshot'ı fraksiyon başına üretim, sivil talep, ordu bakımı, net değişim, stok ve stokun kaç ay yeteceğini taşır. Toplam talebin 3 aydan az stoğa oranı uyarı, 1 aydan azı kritik, mevcut tur talebi karşılanamadığında kıtlık sayılır.
- Uyarı seviyesinde gelir %5 ve memnuniyet 1, kritik seviyede gelir %10 ve memnuniyet 2, kıtlık seviyesinde gelir %25 ve memnuniyet 4 azalır. Gerçek stok açığı varsa mevcut ordu HP cezası yalnız o tick'in hesaplanan açık miktarı kadar uygulanır; ordu sırası deterministiktir.
- Her ekonomi turunda bölgenin vergi, bina, tahıl, teknoloji, savaş, genişleme ve ordu etkileri tek bir memnuniyet deltası olarak toplanır; sonuç `0–100` aralığına sınırlandırılır. Kışla her kurulu seviye için `-1`, pazar/çiftlik/liman `+1` uygular.
- Aralık ekonomi turunda yıl sonu yıpranması olarak tüm sahipli kara bölgelerine ek `-1` memnuniyet uygulanır. Bu etki yalnızca yılda bir kez çalışır ve diğer memnuniyet etkileriyle aynı delta içinde toplanır.
- Bir fraksiyon herhangi bir devletle savaş halindeyse savaş yorgunluğu nedeniyle sahip olduğu tüm kara bölgeleri `-1` alır. Fraksiyon 20'den fazla kara bölgesine sahipse yozlaşma nedeniyle tüm kara bölgeleri ayrıca `-1` alır.
- Bir bölgede sahibine ait kara orduları varsa toplam `TotalStrength / 10` kadar, en fazla `+10`, memnuniyet bonusu verilir. Düşman ordusu bu bölge istikrar bonusuna dahil değildir.
- Tahıl arzı ordunun kalıcı moraline de bağlanır: stabil seviyede her ekonomi tick'inde `+1`, uyarı/kritik/kıtlık seviyelerinde sırasıyla `-1/-3/-6` uygulanır. Moral `1–100` arasında tutulur; 100 moral nötr, 50 moral yaklaşık `%15` toplam savaş gücü kaybı üretir. Gerçekleşen toplam değişim `GrainEconomyStatus.ArmyMoraleDelta` ile HUD/event detayına taşınır ve uygulama Army ID sırasıyla deterministiktir.
- Depolama kapasitesi `6 × sivil talep + 3 × ordu bakımı` olarak hesaplanır; talep varsa minimum kapasite 100'dür. Kapasite üstündeki stok her ekonomi tick'inde fazlanın %2'si oranında, en az 1 tahıl olacak şekilde bozulur. `StorageCapacity` ve `Spoiled` runtime snapshot alanlarıdır; save migration gerektirmez.
- `granary` / `Ambar` binası her kurulu seviye için +100 tahıl depolama kapasitesi verir. Bina tüm senaryolarda veri tanımı olarak bulunur; özel sprite yoksa mevcut çiftlik sprite'ı görsel fallback olarak kullanılır.
- 1300 AI ekonomi puanlamasında ilk ambar, stok/kapasite/üretim-bakım açığı varsa çiftlik ve pazarın önüne çıkar. Ambar kapasitesi yalnız stok saklamaz; `applyRegionalLogisticsPressure()` ve kapasite üzeri tahıl yenilemesi üzerinden orduların toparlanabilirliğini de artırır.
- Bölgesel lojistikte ambar, ekonomi tick'i sonrası elde kalan tahıldan `min(kalan stok, ambar kapasitesi)` kadarını bölgeye aktarılabilir askerî rezerv yapar. Bu tahıl ikinci kez tüketilmez; genel ordu bakımında zaten düşülmüş stokun bölgesel dağıtım kapasitesini temsil eder. Aynı fraksiyonun bölgeleri sınırlı rezervi deterministik sırayla paylaşır ve başkent önce gelir.
- Kuşatılan bölgedeki her `granary` seviyesi savunucu ordunun kuşatma kaynaklı doğrudan HP hasarını ve bölgesel ikmal açığı hasarını `%10` azaltır; toplam azaltma `%30` ile sınırlıdır. Bu yerel ambar dayanıklılığı kuşatan orduya verilmez.
- Kara orduları ayrıca bölge bazlı ikmal kapasitesine tabidir. Yerel askeri kapasite, bölge üretiminden önce sivil talep düşüldükten sonraki fazlalık + yerleşim/ticaret tamponu + fraksiyon stokundan sınırlı destek olarak hesaplanır. Yabancı/düşman bölgede yerel üretim desteği yoktur. Efektif ordu talebi kapasiteyi karşılamazsa aynı bölgede bekleyen ordular turdan tura artan HP zayiatı alır. AI hareket, geri çekilme, birleşme ve bina yatırımındaki lojistik tahminlerde `GameState.EffectiveArmyGrainUpkeep()` kullanır.
- Kara ordusu toparlanması bu ikmal kararından sonra yapılır; talep kapasiteyi aşıyorsa hasarlı ordu aynı tur HP geri kazanmaz. Kapasite yeterliyse ücretsiz hız `2 + 2 × (Çiftlik seviyesi + Ambar seviyesi)` HP/birim/turdur. Depo kapasitesi üzerindeki tahılın bedelli yenileme tavanı da aynı bölgesel hızdır; böylece bina olmayan bölgede büyük fraksiyon stoğu hızlı toparlanma yaratmaz.
- Ordu veya filo farklı konuma taşındığında eski konuma ait `ArmyLogisticsStatus` temizlenir. Kara ordusunun bölgeye özgü `OverCapacityTurns` sayacı hedef bölgede sıfırdan başlar; filonun açık deniz yolculuk süresini taşıyan `TurnsWithoutPort` sayacı korunur. Böylece harita marker'ındaki `!` yalnız mevcut konumdaki güncel yıpranma kaydını gösterir.
- Kara ordusunun bölgesel ikmal talebi, başkentten yalnız kendi kara bölgeleri üzerinden kurulabilen ikmal hattına göre de ölçeklenir. Başkente yakın iki bölgelik hat cezasızdır; daha uzak hatlarda yerel yıpranma baskısı kademeli artar, geçerli başkente kara bağlantısı olmayan ordular en yüksek ek yükü alır. Düşman tahkimatını kendi, aynı realm/vassal ya da müttefik kara bölgesine bitişik yerde kuşatan ordu düzenli sınır ikmaliyle kuşatma bakımını `%200` yerine `%150` sayar.

Müttefik/vassal sınır ikmali ücretsiz değildir. Destekçi devletin tur sonu tahılından, her desteklenen ordu için taban bakımın aynı realm/vassal durumda `%20`si, bağımsız müttefik durumda yaklaşık `%34`ü düşülür. Bir devlet en az `max(20 tahıl, bir aylık toplam talep)` rezervini koruyamıyorsa ileri ikmal göndermez; ordu normal başkent uzaklığı/yıpranma hesabına geri döner. AI de aynı uygunluk kontrolünü kullanır. Birden çok istek varsa ordular ID sırasıyla paylaştırılır. `ArmyLogisticsStatus`, bölge ikmal satırı ve tur olay kaydı destekçiyi ve bedeli gösterir; `GrainEconomyStatus.FriendlySupplyGrainSpent` ise destekçinin toplam harcamasını taşır.
- Limanlı bölgenin komşu denizinde düşman savaş gemisi varsa abluka oluşur: savaş gemisi başına ilgili ticaret rotası hacmi ve limanın yerleşim/rezerv ikmal tamponu %50 azalır; iki veya daha fazla gemi rotayı/tamponu tamamen keser. Buna ek olarak `%50` liman ablukası bölgesel vergi, yerel ticaret ve mal üretimini `%75` seviyesinde bırakır; `%100` abluka `%50` seviyesinde bırakır. Ablukacı, etkili gemi katkısına göre hedef bölgenin ablukasız yerel çıktısının `%50` durumda `%5`'ini, `%100` durumda `%10`'unu altın ve mal olarak alır. Abluka yüzdesi runtime state'ten her ekonomi tick'inde yeniden türetilir; `RegionProductionSummary()` aynı oranı UI önizlemesine taşır.
- Kasım ekonomi tick'inde arz seviyesi stabil olan fraksiyon, yalnızca depolama kapasitesini aşan tahılını nüfus yatırımında kullanır. Memnuniyeti en az 60 olan, isyan riski taşımayan ve kuşatma altında olmayan bölgelerde toplam bölge nüfusunun %1'i (minimum 1) kırsal nüfusa eklenir; her 1 nüfus artışı 2 tahıl harcar. Bu harcama `GrainEconomyStatus` içinde runtime raporlanır; sonraki tick'lerde artan toplam nüfus `ceil(population / 18)` sivil talebe dönüşerek büyümeyi aynı tahıl döngüsüne bağlar ve rezerv kapasite tabanı korunur.
- Bölge panelindeki `Tahıl Yardımı` aksiyonu kendi bölgesine 12 tahıl aktararak memnuniyeti +10 artırır; bölge başına turda bir kez kullanılabilir. Kuşatma altındaki, oyuncuya ait olmayan, zaten yüksek memnuniyetli veya yetersiz stoklu bölgeler yardım alamaz. Tur ilerlediğinde yardım kullanım kilidi runtime olarak sıfırlanır.
- Pazar ekranındaki `ACİL TAHIL SAT` aksiyonu ticaret partneri gerektirmez. Yalnızca fraksiyonun `StorageCapacity` üstündeki tahıl satılabilir; satış fiyatı güncel tahıl piyasa fiyatının %70'idir ve minimum 1 altın/tahıl olarak hesaplanır. Ancak bu satışın bir turdaki toplam altın getirisi, kuşatma dışı temel vergi gelirinin %100'ü ile sınırlıdır; `GrainSaleGoldUsed` runtime haritası tur sonunda sıfırlanır. Böylece tahıl acil nakit sağlayabilir ama verginin yerini alamaz.
- Pazar sekmesindeki `Oto. İhracat` toggle'ı açıkken ekonomi tick'i kapasite üstü tahılı aktif, savaşta olmayan ticaret ağı partnerlerine otomatik satar. Partnerler faction ID sırasıyla işlenir, fiyat güncel tahıl piyasa fiyatının %60'ıdır ve alıcının altını yetersizse satış miktarı düşürülür. Otomatik ihracat da aynı vergi geliri üst sınırını paylaşır; tercih `GameState.AutoGrainExport` olarak save'e yazılır, gerçekleşen ihracat tahılını ve altınını `GrainEconomyStatus` raporlar.
- Üretim emri iptalinde altınla birlikte diğer kaynaklar da iade edilir.

## Bölge Uzmanlaşması

- Ova: tahıl üretimi artar.
- Orman: kereste üretimi artar.
- Dağ/geçit: demir ve taş üretimi artar.
- `base_stone_output` olmayan senaryolarda dağ/geçit bölgeleri fallback taş üretimi sağlar.

### 1300 Osmanlı Yükselişi Kaynak Profili

`assets/scenarios/1300_ottoman_rise/data/regions.json` coğrafi uzmanlaşmayı veri seviyesinde taşır. Tahıl verimli ovalar ve nehir havzalarında; kereste orman, kuzey ve kıyı hinterlandlarında; taş ve demir dağ/geçit/maden hatlarında yoğunlaşır. Baharat Mısır-Kızıldeniz, Şam-Halep, Basra ve Akdeniz ticaret düğümlerine; kumaş Bursa, Konstantiniyye, Selanik, Flandre ve İtalyan şehirlerine dağıtılmıştır. Böylece ticaret malları artık yalnız başlangıç stoğundan değil, ele geçirilen bölgesel üretim merkezlerinden de elde edilir.

Senaryo verisi için bütün kara bölgelerinde toplam başlangıç üretimi yaklaşık `tahıl 5948`, `demir 287`, `kereste 1098`, `taş 632`, `baharat 679`, `kumaş 1342` seviyesindedir. İngiltere ve Fransa'daki aşırı tahıl üretimi aşağı çekilmiş, 1300 birliklerinin `grain_upkeep` değerleri artırılmıştır. Üretim uzmanlaşması `Test1300ScenarioResourceSpecializationsAndProductionCosts`, ekonomi sürdürülebilirliği ise `Test1300ScenarioGrainEconomyBands` ile korunur.

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
- [x] Nüfus bazlı temel sivil tahıl tüketimi (`ceil(population / 18)`)
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
