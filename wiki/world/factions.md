---
type: world
tags: [factions, religion, diplomacy, starting-positions]
last_updated: 2026-07-31
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

`ai_expansion_targets` opsiyoneldir. Tanımlandığında AI diplomasi safhasında yalnız kara sınırı paylaştığı ve hala `peace` durumunda olan bu fraksiyonlara daha yüksek öncelik verir; `trade` veya `allied` ilişkiyi yine savaş için bozmaz.

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

Senaryo `relations.json` dosyası bu varsayılanları tarihsel başlangıç skorlarıyla ezer. AI'nın proaktif savaş hedefleri ise fraksiyon kaydındaki `ai_expansion_targets` alanında tutulur; örneğin 1300 senaryosunda Osmanlı için Doğu Roma, Germiyan, Karesi ve Ahiler hedeflenir.

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
