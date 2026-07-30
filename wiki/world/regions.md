---
type: world
tags: [regions, terrain, map, neighbors, coastal, succession]
last_updated: 2026-07-30
related: [systems/combat, world/factions, architecture/render-pipeline]
---

# Bölge Sistemi

**Kaynak:** `internal/world/region.go`, `internal/world/terrain.go`, `assets/scenarios/<id>/data/regions.json`

## Region Yapısı

```go
type Region struct {
    ID        RegionID
    NameTR    string
    OwnerID   string           // fraksiyon ID veya ""
    SuccessorFactionID string   // fetih sonrası yeniden kurulabilecek devlet
    Terrain   TerrainType
    Neighbors []RegionID       // komşu bölge listesi

    IsSea     bool             // deniz bölgesi
    IsLocked  bool             // henüz keşfedilmemiş
    WorldX, WorldY int         // harita koordinatı
    ShapeID string             // Natural Earth kaynak ID'si
    Settlements []Settlement   // görsel şehir/kasaba/kale noktaları

    Buildings    []string      // inşa edilmiş bina ID'leri
    TaxRate      int           // 0-100
    Satisfaction int           // halk memnuniyeti
    Population      int // RuralPopulation + SettlementPopulation()
    RuralPopulation int // yerleşim dışındaki köy/kırsal nüfus

    Religion        string     // mevcut bölge dini
    ConversionTurns int        // din dönüşüm sayacı
    ActiveEventID   string
}
```

`successor_faction_id`, edit mode'da `Ardıl Devlet` düğmesiyle atanır. Bölge oyuncu
tarafından fethedildiğinde bu fraksiyon `is_eliminated=true` ise bilgi panelinde
`Özgürleştir` görünür; aksiyon bölgeyi ardıl devlete verir ve devletin yeniden
kuruluş state'ini başlatır. 1300 senaryosunda başkent settlement'ı bulunan ve
sahibi eşleşen 68 bölge başlangıçta kendi sahibiyle işaretlidir.

Oyuncu ordusu bu metadata'yı taşıyan düşman kara bölgesini savaşla veya savaşsız
ele geçirdiğinde, savaş raporu kapatıldıktan sonra `Ardıl Devlet Kararı` paneli
açılır. `İlhak Et` bölgeyi oyuncuya verir; `Serbest Bırak` ardıl devleti bağımsız
müttefik olarak kurar; `Vassal Yap` bölgeyi ardıla verip onu oyuncunun doğrudan
vassalı yapar. Elenmiş ardıl, iki bölgesel kurulum seçeneğinde düşük kaynak ve
beş milisle yeniden etkinleştirilir.

`WorldX/WorldY` bölge geometrisi ve Voronoi ayrımı için korunur. Haritadaki şehir noktaları `Settlements` üzerinden çizilir; ana yerleşim `is_capital` ile seçilir. `settlements` eksikse renderer eski davranışa dönüp bölge adını `WorldX/WorldY` noktasından çizer.

```go
type Settlement struct {
    ID         string
    NameTR     string
    X, Y       int     // world_x/world_y ile aynı koordinat uzayı
    Type       string  // city, town, port, fortress
    IsCapital  bool
    Population int
}

`Region.Population` artık doğrudan bağımsız bir nüfus kaynağı değildir; `RuralPopulation` ile bölgedeki tüm `Settlement.Population` değerlerinin toplamıdır. Kırsal pay; köyler, mezralar ve yerleşim dışındaki tarımsal nüfusu temsil eder. Tahıl sivil tüketimi bu toplam bölge nüfusu üzerinden hesaplanır. Eski kayıtlarda bileşen alanları yoksa mevcut toplam nüfus kırsal nüfus olarak göç edilir.
```

Yerleşim koordinatı yanlışlıkla bölge raster alanının dışına düşerse render cache yüklenirken uyarı loglanır ve nokta aynı region içindeki en yakın piksele taşınır.

Kıyı bölgesinde `port` binası tamamlandığında, bölgede henüz `type=port` settlement yoksa oyun bu bölge için denize yakın yeni bir `Liman` yerleşimi üretir. Böylece liman binası sadece ekonomi/üretim değil, dock edilen filonun görünür anchor noktası için de tekil veri kaynağı olur.

`1300_ottoman_rise` başlangıç verisinde Londra, Normandiya, Portekiz, Sicilya ve
Mısır'a tarihsel başlangıç filolarının dock edilebilmesi için birinci seviye `port`
binası tanımlıdır. Bu limanlar `data/armies.json` içindeki ilgili filo ve port
settlement kayıtlarıyla birlikte doğrulanır.

Flandre bölgesi (`flanders`) 1300 açılışında doğrudan HRE yerine `flanders_county`
tarafından yönetilir. Bu sahiplik, yerel vergi/ticaret akışını korurken HRE vassallığı
ve Flandre limanının savunma görevlerini veri modelinde görünür kılar.

1300 açılışındaki sahipsiz kara bölgeleri tarihsel devletlere atanmıştır: Fas Merînîlere;
Batı/Orta Cezayir Tlemsen Zeyyânîlerine; Konstantin, Tunus ve Trablus Hafsîlere;
Berka Memlük bağlısı Berka Emirliği'ne; Bahreyn-Katar Usfûrîlere; Hürmüz kıyısı
Hürmüz Sultanlığı'na; Malta Aragon'a; Ermenistan ve Basra/Kuveyt İlhanlılara aittir.
Orta Cezayir, Konstantin bölgesi olarak ayrılmış; Annaba ve Biskra yerleşimleri bu yeni
bölgeye taşınmıştır. Başlangıç orduları `data/armies.json`, üretim ve stok değerleri
ise `regions.json` ile `factions.json` içinde doğrulanır.
Hicaz ise Mekke Şerifliği'ne verilmiş, Memlük vassallığı korunmuştur.
Arab Çölü (`arabian_desert`) ise Memlüklerden çıkarılmış, 1300 için sahipsiz/çekişmeli
bir keşif ve geçiş bölgesi olarak bırakılmıştır; bu alanı doğrudan yöneten tek bir
merkezî devlet yoktur.

1300 Balkan düzeltmesinde `serbia` Sırp devletine, `slovenia` Kranj Marklığı'na,
`croatia` ve `kvarner` (Kvarner) Hırvat Krallığı'na, `bosnia` Bosna Banlığı'na;
`hum` ve `herzegovina` (Hersek) ise 14. yüzyıl başındaki Šubić etkisini temsil eden
Hırvat bağlı devletine verildi. `kvarner` için Senj, `herzegovina` için
Trebinye başlangıç yerleşimleri eklendi. Böylece Macaristan'ın doğrudan renk alanı
çekirdek havza ve tarihsel Macar krallığı bölgeleriyle sınırlı kalırken, kişisel birlik
ve yerel banlık ilişkileri harita üzerinde görünürdür.

1300 senaryosundaki otomatik `new_region_*` kayıtları anlamlı coğrafi ID'lere taşındı.
Kara tarafında `welsh_marches`, `scania`, `algarve`, `toledo`, `luxor`, `raqqa`,
`podolia` ve `north_caucasus`; deniz tarafında `sea_of_marmara`, `bosphorus`,
`western_black_sea`, `cretan_sea`, `alboran_sea`, `levantine_sea` ve ilgili körfez/
boğaz kayıtları kullanılır. Aynı koordinattaki komşusuz iki sahte deniz kaydı kaldırıldı;
Scania ve Ragusa'ya da başlangıç yerleşimleri eklendi.

---

## Arazi Tipleri

`internal/world/terrain.go`

| Tip | Geçiş | Savunma Bonusu | Görüş |
|---|---|---|---|
| `TerrainPlain` (Ova) | Serbest | ×1.0 | Tam |
| `TerrainForest` (Orman) | Yavaş | ×1.3 | Kısıtlı |
| `TerrainMountain` (Dağ) | Geçilemez (geçit hariç) | ×1.8 | Yok |
| `TerrainPass` (Geçit) | Tek yol | ×1.5 | Kısıtlı |
| `TerrainCoast` (Kıyı) | Normal | ×1.1 | Normal |
| `TerrainSea` (Deniz) | Sadece deniz ordusu | — | — |

Arazi ve yerleşim tiplerinin Türkçe görünen etiketleri artık paket içinde tutulur: `TerrainType.LabelTR()` ve `SettlementType.LabelTR()`. UI panelleri bu değerleri doğrudan `internal/world` metadata'sından alır.

→ Savunma bonusu çarpışmaya etkisi: [[systems/combat]]

---

## Hareket Kuralları

`CanLandEnter()` — kara orduları deniz bölgesine giremez
`CanNavalEnter()` — deniz orduları sadece deniz bölgelerine girer
`IsCoastal()` — komşu bölgeler arasında deniz varsa `true` → gemi inşa koşulu
`HasPortBuilding()` — docking için gerekli operasyonel liman binası

---

## Ele Geçirme

`ApplyConquest(ownerID, religion)` — savaş sonrası sahiplik transferi

1. `OwnerID = ownerID` → sahip değişir
2. Memnuniyet -10 düşer
3. Saldıranın dini bölgeden farklıysa ekstra -15 memnuniyet cezası uygulanır
4. Din dönüşümü tur çözümlemede `ConversionTurns` ile ilerler; 24 tur sonunda bölge dini yeni sahibin dinine döner

---

## Komşuluk Grafı

`Neighbors []RegionID` — hem kara hem deniz komşuları içerir.

Ordu hareketi bu listeyle kısıtlanır: sadece direkt komşuya hareket.

---

## Kilit Sistemi

`IsLocked = true` olan bölgeler haritada görünmez/girilemez. `checkRegionUnlocks()` belirli koşullarda (bölge yakınlaşması, teknoloji, tarih) `IsLocked = false` yapar.

---

## Kritik Bölgeler

Zafer koşulları ve olaylar için referans alınan bölgeler:

| Bölge ID | Önem |
|---|---|
| `constantinople` | Domination + Doğu Roma teknoloji dalı |
| `papal_states` | Domination + Dini zafer (Roma bölgesi temsili) |
| `palestine` | Domination + Dini zafer (Kudüs bölgesi temsili) |
| `egypt` | Domination (Kahire/Mısır bölgesi temsili) |
| `paris` | Domination (Fransa başkenti) |
| `london` | Domination (İngiltere başkenti) |
| `yemen` | Dini zafer için Arap yarımadası temsili |

Not: Senaryo zafer hedefleri `regions.json` içindeki gerçek ID'leri kullanır; örnek: `constantinople`, `papal_states`, `egypt`, `palestine`.

---

## 1300'lü Yıllar Tarihi Bölgeler

### İngiltere Krallığı (6 bölge)
- `london` — Başkent, yüksek gelir (60)
- `yorkshire` — Kuzey, tahıl üretimi (50)
- `lancashire` — Kuzeybatı, dağlık (30)
- `mercia` — Orta, ormanlık (45)
- `east_anglia` — Doğu, tahıl ambarı (40)
- `wessex` — Güneybatı, verimli ovalar (35)

### Fransa Krallığı (8 bölge)
- `paris` — Başkent, Île-de-France (70)
- `normandy` — Normandiya Dükalığı, kıyı (45)
- `brittany` — Bretonya, yarımada (35)
- `anjou` — Anjou Kontluğu, Loire vadisi (40)
- `champagne` — Şampanya, ticaret merkezi (50)
- `burgundy` — Burgonya Dükalığı (55)
- `provence` — Provence, Akdeniz kıyısı (50)
- `languedoc` — Languedoc, Toulouse (45)

### Kutsal Roma İmparatorluğu (6 prenslik)
- `brandenburg` — Brandenburg Markgrafluğu, kuzeydoğu
- `saxony` — Saksonya Dükalığı, kuzey orta
- `bavaria` — Bavyera Dükalığı, güney
- `westphalia` — Vestfalya, batı (Ren bölgesi)
- `thuringia` — Turingiya, orta
- `palatinate` — Palatinate, orta-batı
