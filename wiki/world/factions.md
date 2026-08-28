---
type: world
tags: [factions, religion, diplomacy, starting-positions]
last_updated: 2026-08-06
related: [systems/diplomacy, world/regions, architecture/state-management]
---

# Fraksiyonlar

**Kaynak:** `internal/faction/faction.go`, `internal/faction/loader.go`, `internal/religion/religion.go`, `assets/scenarios/<id>/data/factions.json`

## Fraksiyon Verisi

Her aktif senaryo 45 fraksiyon içerir; oynanabilir roster senaryoya göre değişir (`is_playable=true`). Örneğin `1300_ottoman_rise` 30 oynanabilir fraksiyon açarken `1444_ottoman_empire` tarihsel hedefi olan 6 devleti, `1512_yavuz_selim` ise tarihsel hedefi olan 5 devleti açık bırakır. Başlangıç orduları fraksiyon dosyasında değil, aynı senaryonun `data/armies.json` dosyasında tutulur.

`1300_ottoman_rise` başlangıç filoları da aynı `armies.json` kaynağında tutulur.
Venedik ve Ceneviz savaş/ticaret filoları; Doğu Roma, Aragon, İngiltere ve Fransa
savaş-nakliye filoları; Portekiz nakliye/ticaret filoları; Memlük nakliye filosu
başlangıçta tarihsel ana limanlarına bağlıdır. Osmanlı, Safevî ve Rusya 1300 açılışında
donanmasız bırakılmıştır; bu devletlerde deniz gücü kıyı ve liman altyapısı geliştikçe
oyun içinde kurulacaktır.

1300 açılışında daha önce sahipsiz kalan Kuzey Afrika ve Körfez bölgeleri için
`marinid_sultanate`, `zayyanid_tlemcen`, `hafsid_sultanate`, `barqa_emirate`,
`usfurid_emirate` ve `hormuz_sultanate` veri fraksiyonları tanımlıdır. Bunlar tarihsel
dengeyi ve AI cephelerini doldurur, ancak başlangıç roster'ında oynanabilir değildir.
Berka, Memlük overlord'u olarak; yeni devletlerin ordu, stok, teknoloji ve AI hedefleri
aynı senaryo veri dosyalarında tutulur.

Hicaz doğrudan Memlük toprağı yerine `mecca_sharifate` adlı AI-only vassal ile modellenir;
Mekke Şerifliği Memlük üst-egemenliğini kabul eder, fakat yerel yönetim ve kutsal şehirlerin
korunması kendisinde kalır.

1300 Balkan başlangıcında Macar tacının doğrudan yönetimi dört çekirdek bölgeyle
(`hungary`, `alfold`, `slovakia`, `transylvania`) sınırlandırıldı. Sırbistan mevcut
`serbian_empire` devletine verildi; Slovenya ise Kutsal Roma İmparatorluğu'na bağlı
`carniola_margraviate` olarak ayrıldı. Hırvatistan ve Bosna'nın Macar tacıyla kişisel
birlik/vasallık ilişkisini göstermek için `croatian_kingdom` ve `bosnian_banate` AI-only
bağlı devletleri eklendi. Hırvat bağlı devletinin Kvarner, Hum ve Hersek sınırları için
başlangıç ordusu, komutanı, stokları ve Adriyatik hedefleri; Bosna ve Carniola için de
ayrı savunma orduları ve AI profilleri tanımlıdır.

---

## Faction Yapısı

```go
type Faction struct {
    ID           FactionID
    Name         string
    NameTR       string
    Religion     religion.Type // catholic | orthodox | sunni | shia
    Color        [3]uint8      // harita rengi (RGB)
    IsPlayable   bool
    IsEliminated bool
    OverlordID   FactionID
    VassalizedTurn int       // vassallık bağının kurulduğu toplam tur
    CapitalSettlementID        string
    PendingCapitalSettlementID string
    PendingCapitalTurns        int

    Gold   int
    Grain  int
    Iron   int
    Timber int
    Spice  int
    Cloth  int

    Research         ResearchState
    AIAggressiveness int           // AI saldırganlık düzeyi
    AIExpansionTargets []FactionID  // AI tarihsel genişleme hedefleri
}
```

`capital_settlement_id`, fraksiyonun ulusal başkent settlement'ını tutar. Bu alan artık bölgesel `settlement.is_capital` işaretinden ayrıdır: `is_capital` bir bölgenin ana yerleşimini/anchor noktasını belirlemeye devam ederken, ulusal başkent tekil olarak fraksiyon üstünde tutulur.

`pending_capital_settlement_id` ve `pending_capital_turns`, oyuncunun veya bir event'in başlattığı başkent taşıma kuyruğunu saklar. Sayaç sıfırlandığında yeni settlement resmen başkent olur.

`overlord_id`, fraksiyon başka bir devlete bağlıysa o overlord'u gösterir. Bu alan:

- vassallık zincirini save/load içinde taşır
- diplomasi UI'sında hiyerarşik durum etiketini besler
- ekonomi tick'indeki altın haracı ve savaş coalition yayılımı için kullanılır

## Ardıl Devletin Yeniden Kurulması

Bir bölgedeki `successor_faction_id`, o bölge özgürleştirildiğinde yeniden oyuna
alınacak mevcut fraksiyonu gösterir. Fraksiyon `is_eliminated=true` ve başka kara
bölgesi yoksa oyuncu `Özgürleştir` aksiyonuyla devleti etkinleştirir. Yeniden kurulan
devlet düşük başlangıç kaynakları, aynı bölgede beş `militia` birimi ve onu
özgürleştiren devletle `StanceAllied` ilişkisi alır; ardından normal AI tur sırasına
katılır. Ardıl fraksiyon hâlâ aktifse bu metadata vassallık/serbest bırakma
seçeneği üretmez; fethedilen bölge doğrudan ilhak edilir.

1300 senaryosunda tarihsel ardıl havuzu `factions.json` içinde başlangıçta
elenmiş olarak tutulur. Artuklular (Musul çekirdeği), Eretna (Kayseri), Kadı
Burhaneddin (Sivas), Karakoyunlu (Van), Akkoyunlu (Diyarbekir), Celayirliler
(Bağdat), Muzafferiler (İsfahan çekirdeği), Şirvanşahlar (Şamahı), Timur
İmparatorluğu (senaryodaki Meşhed çekirdeği) ve Afşar Beyliği (Erzurum)
`historical_start_year`/`historical_end_year` alanlarıyla tarih aralıklarını
taşır. Bu yıllar tarihçe metadata'sıdır; devletin fiilî kuruluşu bölgedeki
`successor_faction_id` üzerinden isyan, fetih sonrası karar veya özgürleştirme
akışıyla gerçekleşir.
Kuruluş kaynakları faction kaydındaki mevcut `gold`, `grain`, `iron`, `timber`,
`stone`, `spice` ve `cloth` alanlarından korunur. Bu nedenle isyanla kurulan
devletler de özgürleştirilen devletlerle aynı başlangıç kaynaklarını alır.
Tarihsel başlangıç tarihine son 10 yıl kala ardıl bölgenin memnuniyeti kademeli
olarak azalır. Bitiş tarihine son 10 yıl kala daha güçlü bir çözülme cezası
uygulanır; bitiş tarihi geçtikten sonra ceza tur başına `-4`te sabitlenir. Bu
ceza, mevcut memnuniyet/isyan eşiğiyle birlikte çalışır; tarih tek başına
devleti otomatik olarak silmez, ancak isyanı ve toprak kaybını daha olası hâle
getirir.

`ai_expansion_targets` runtime uyumluluk alanıdır. 1300 senaryosunda kaynak
`ai_strategies.json` içindeki `expansion_targets` alanıdır; AI'nin bölgesel savaş
hedefi ise claim edilen bölgenin güncel sahibinden dinamik olarak türetilir.

---

## Din Sistemi

`religion.Type` — `internal/religion/religion.go`

| Din | Sabit | Fraksiyonlar |
|---|---|---|
| Katolik | `catholic` | Fransa, İngiltere, Venedik, Aragon, Portekiz |
| Ortodoks | `orthodox` | Rusya |
| Sünni İslam | `sunni` | Osmanlı, Memlük |
| Şii İslam | `shia` | Safevi |

**`religion.Relation(a, b religion.Type) int`** — `internal/religion/religion.go`

| Kombinasyon | Puan |
|---|---|
| Aynı din | +25 |
| Sünni ↔ Şii | -40 |
| Katolik ↔ Ortodoks | -20 |
| Diğer farklı din | -30 |

Bu puan `BuildInitialRelations()` sırasında ilişki skorlarına eklenir.

Dinlerin görünen Türkçe adları ve editörde/UI'da dolaşım sırası artık `internal/religion/religion.go` içindeki metadata üzerinden (`DisplayNameTR`, `All`, `Next`) merkezi olarak yönetilir; render katmanı aynı mapping'i tekrar etmez.

**Mezhep değişimi:** Ele geçirilen bölge `ConversionTurns` sayacıyla 24 turda yeni sahip dinine geçer, memnuniyet -20 uygular. → [[world/regions]]

---

## Başlangıç İlişkileri ve Hedefler

`faction.BuildInitialRelations(factions)` — tüm çiftlerin skoru `religion.Relation()` sonucuyla başlatılır. Sünni-Şii çiftleri başlangıçta savaş duruşu alır, diğer çiftler barışta başlar.

Senaryo `relations.json` dosyası bu varsayılanları tarihsel başlangıç skorlarıyla ezer. AI'nın proaktif genişleme hedefleri 1300'de `ai_strategies.json` içindeki `expansion_targets` alanından yüklenir; faction üzerindeki `AIExpansionTargets` yalnız runtime uyumluluk görünümüdür.

`TerritorialClaims` aynı faction kaydındaki bölgesel talepleri taşır. Senaryo
başlangıcında devletin sahip olduğu tüm kara bölgeleri otomatik `core: true`,
`value: 100` olarak eklenir. `ai_strategies.json` içindeki `territorial_claims`
ve objective içindeki `territorial_claims` normal claim olarak eklenir; açık strateji claim'i
aynı bölgedeki otomatik objective değerini geçersiz kılar. Claim bölgeye aittir,
o anki sahibine değil.
`value` barış kararındaki stratejik ağırlıktır. Barış değerlendirmesi yalnız
düşmanın hâlen tuttuğu claim'leri hesaba katar; fetih sonrası yeni sahiplik
otomatik olarak yeni core üretmez. Üretim `internal/scenario/territorial_claims.go`
ile hem yeni oyun hem save/load base state'inde uygulanır.

Runtime'da vassallık kabul edilirse hedef fraksiyonun `overlord_id` alanı doldurulur; üçüncü taraf diplomasi kapatılır ve realm içindeki fraksiyonlar dost çizgiye normalize edilir.

`1300_ottoman_rise` başlangıcında `flanders_county`, Flandre bölgesini yöneten Katolik
bir alt devlettir ve `overlord_id: "hre"` ile Kutsal Roma İmparatorluğu'na bağlıdır.
Flandre'nin ticaret ve liman kapasitesi korunurken dış savaş ilişkisi HRE kök realm'i
üzerinden koalisyona katılır; HRE ile garantili iç realm ticareti dış ittifak sayılmaz.

1300 senaryosunda HRE'nin bağımsız imparatorluk üyeleri `data/imperial.json` içinde
ayrıca tanımlıdır. Avusturya, Bohemya, Bavyera, Saksonya ve Brandenburg çekirdek
prenslikleri ile Milano, Savoy ve Töton Şövalyeleri `OverlordID` almadan imparatorluk
çağrısı alır; bu nedenle çağrıya katılabilir, sınırlı kaynak desteği verebilir veya
tarafsız kalabilirler. Beş çekirdek prensliğin ayrıca başlangıç ordusu, komutan şablonu
ve kendi AI savunma hedefi vardır. Flandre ve Kranj ise gerçek vassal olarak otomatik realm
katılımını korur.

→ İlişki sistemi: [[systems/diplomacy]]

---

## Eleme Koşulu

`gs.IsEliminated(fid)` → `len(RegionsOwnedBy(fid)) == 0`

Bir fraksiyon tüm bölgelerini kaybedince `checkEliminations()` tarafından tespit edilir ve `FactionsEliminated` sayacı artar.

## Başkent Davranışı

- Her fraksiyonun aynı anda tek aktif başkenti vardır.
- Senaryo verisinde `capital_settlement_id` yoksa runtime yükleme akışı fraksiyonun en yüksek getirili kara bölgesindeki ana settlement'ı başkent olarak seçer.
- Başkent bölgesi ek gelir, stok ve lojistik avantajı alır.
- Başkent fethedilirse savunan fraksiyonun hazine/hammadde stoklarının bir bölümü fethedene geçer.
- Başkent fetheden taraf, savunanın sahip olduğu ama kendisinde olmayan tamamlanmış teknolojilerin yaklaşık yarısını anında açar.
- Başkent kaybeden ama hayatta kalan devletin yeni başkenti otomatik olarak en yüksek getirili bölgesinin merkez settlement'ına atanır.
