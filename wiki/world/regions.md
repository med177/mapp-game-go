---
type: world
tags: [regions, terrain, map, neighbors, coastal, succession]
last_updated: 2026-09-30
related: [systems/combat, world/factions, architecture/render-pipeline]
---

# Bölge Sistemi

**Kaynak:** `internal/world/region.go`, `internal/world/terrain.go`, `internal/city/building.go`, `assets/scenarios/<id>/data/regions.json`, `assets/scenarios/<id>/data/buildings.json`

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
    IsMinorRegion bool         // ana bölge içindeki fethedilebilir küçük alt bölge
    IsPrivileged bool           // kullanım OwnerID'de, egemenlik ana bölge sahibinde
    PrivilegeGrantedTurn int    // aktif imtiyazın verildiği kampanya turu
    ParentRegionID RegionID    // küçük alt bölgenin ana bölgesi
    IsLocked  bool             // henüz keşfedilmemiş
    WorldX, WorldY int         // harita koordinatı
    ShapeID string             // Natural Earth kaynak ID'si
    Settlements []Settlement   // görsel şehir/kasaba/kale noktaları

    Buildings    []string      // inşa edilmiş bina ID'leri
    TaxRate      int           // 0-60; world.ClampTaxRate ile sınırlandırılır
    Satisfaction int           // halk memnuniyeti
    Population      int // regions.json içindeki bağımsız bölge toplam nüfusu
    RuralPopulation int // eski save/oyun içi nüfus değişimleri için yardımcı alan

    Religion        string     // mevcut bölge dini
    ConversionTurns int        // din dönüşüm sayacı
    ActiveEventID   string
}
```

`is_minor_region=true` olan bölgeler normal `Region` olarak kalır; bu nedenle
kuşatılabilir, fethedilebilir ve sahiplik hesabına katılır. Edit Mode bu tip
bölgeleri ana bölgenin `ShapeID` rasterı üzerinde oluşturur; minor bölge
merkezleri Voronoi alan dağıtımına katılmaz. `Küçük Alt Bölge Ekle` sonrasında
çizilen poligon ana bölgenin rasterına paint override olarak işlenir ve kesin
alanı `region_shapes.json` içindeki `minor_polygons` altında poligon noktalarıyla
saklanır; runtime piksel override'ını bu geometriden üretir. Küçük alt bölgelerde yalnızca
`fortress` ve `port` yerleşimleri geçerlidir. Bina izni bina tanımındaki
`minor_regions` alanından gelir; `1300_ottoman_rise` senaryosunda `walls`,
`granary` ve `port` bu alanla işaretlidir. İmtiyazlı minor bölgelerde
`minor_regions=true` olan binaların tavanı senaryonun
`privileged_building_max_level` alanından okunur; 1300 senaryosunda bu değer 1'dir.
`parent_region_id` ilişki bilgisidir; alt bölgenin fethedilmesi ana bölgenin
sahipliğini değiştirmez. Edit Mode yeni minor kimliklerini `new_minor_region_*`
önekiyle üretir. Poligon çizimi tamamlanana kadar minor için merkez/odak işareti
gösterilmez; onaydan sonra poligonun geometrik merkezi `WorldX/WorldY` olarak
atanır.

`successor_faction_id`, edit mode'da `Ardıl Devlet` düğmesiyle atanır. Bölge oyuncu
tarafından fethedildiğinde bu fraksiyon `is_eliminated=true` ise bilgi panelinde
`Özgürleştir` görünür; aksiyon bölgeyi ardıl devlete verir ve devletin yeniden
kuruluş state'ini başlatır. Aynı koşullarda `Vassallaştır` da görünür; bu aksiyon
ardıl devleti yeniden kurup bölgeyi ona verir ve oyuncunun doğrudan vassalı yapar.
1300 senaryosunda başkent settlement'ı bulunan ve
sahibi eşleşen 68 bölge başlangıçta kendi sahibiyle işaretlidir.

Oyuncu ordusu bu metadata'yı taşıyan düşman kara bölgesini savaşla veya savaşsız
ele geçirdiğinde, ardıl fraksiyon `is_eliminated=true` ve topraksızsa savaş
raporu kapatıldıktan sonra `Ardıl Devlet Kararı` paneli açılır. `İlhak Et` bölgeyi
oyuncuya verir; `Serbest Bırak` ardıl devleti bağımsız müttefik olarak kurar;
`Vassal Yap` bölgeyi ardıla verip onu oyuncunun doğrudan vassalı yapar. Ardıl
fraksiyon hâlâ oyundaysa panel açılmadan bölge doğrudan ilhak edilir. Elenmiş
ardıl, iki bölgesel kurulum seçeneğinde düşük kaynak ve beş milisle yeniden
etkinleştirilir.

`is_privileged=true` yalnız küçük alt bölgelerde kullanılır. Bu durumda `OwnerID`
alt bölgenin kullanım/işletme sahibidir; egemen sahip ayrıca tutulmaz ve
`parent_region_id` ile bağlı ana bölgenin güncel `OwnerID` değerinden çözülür.
İmtiyazlı alt bölgenin parasal geliri, `parent_region_id` ile bağlı üst bölgenin
güncel sahibi ile kullanım sahibi arasında %50-%50 paylaşılır; üst bölge el
değiştirirse bu gelir payı ve `İmtiyazı Kaldır` yetkisi yeni egemen devlete geçer.
Alt bölgenin egemen sahibine yapılan savaş ilanı kullanım sahibiyle doğrudan
savaş başlatmaz; kullanım sahibiyle ilişkiyi 10 puan azaltır.
İmtiyazlı minor bölge her iki devletin bölge görünümünde yer alır. Egemen devlet
bölge panelindeki `İmtiyazı Kaldır` düğmesiyle hakkı sona erdirebilir; bu işlem
kullanım sahibiyle ilişkiyi 10 puan azaltır ve savaş başlatmaz. Kaldırma anında
minor bölge ana bölgenin egemen sahibine normal sahiplikle devredilir; bölgedeki
diğer devlet orduları kendi devletlerinin en yakın kara bölgesine, filoları
denize çıkarılır. Kullanım sahibinin başka egemen toprağı kalmamışsa devlet
elenir ve tüm askeri birimleri silinir.
AI kararları da bu ayrımı kullanır: egemen AI kendi %50 gelir payını bütçe ve
stratejik değer hesabına katar; işletmeci AI yalnız kendi %50 altın payını
yatırım getirisinde görür. Egemen AI, yüksek getirili ve ilişkisi yeterince
güçlü olmayan imtiyazları kaldırarak bölgeyi doğrudan yönetimine alabilir;
müttefik veya yüksek ilişkili işletmecilerin imtiyazlarını korur.
Yeni verilen imtiyazlar `PrivilegeGrantedTurn` ile damgalanır; bekleme süresi
senaryonun `minor_privilege_protection_turns` alanından okunur. Egemen AI,
işletmeciyle savaş yoksa ilk 30 kampanya turu boyunca bu imtiyazı ekonomik
gerekçeyle kaldıramaz. Egemen devlet işletmeciyle savaştaysa bu koruma
uygulanmaz ve imtiyaz savaş istisnası olarak hemen kaldırılabilir. Başlangıç
senaryosundaki imtiyazlar 1. turda verilmiş kabul edilir; damga compact campaign
save içinde korunur.
`is_privileged=false` olduğunda alt bölge normal bölge gibi doğrudan `OwnerID`
sahibine aittir.

Minor bölgelerde `successor_faction_id` Edit Mode tarafından otomatik doldurulmaz;
yeni minor kayıtlarında alan boş/eksik kalır. Kullanıcı `Ardıl Devlet` seçerse
alan o zaman açıkça doldurulabilir.

Kara birimi üretimi için bölgede en az bir `barracks` seviyesi gerekir; milis
de bu kuralın dışına çıkmaz. `port` bulunan fakat
`barracks` bulunmayan kıyı bölgelerinde yalnız deniz birimi üretim hattı açık
kalır. En az bir birim hattının bina gereksinimi sağlanıyorsa alt HUD'daki
üretim düğmesi aktiftir; teknoloji ve kaynak eksikleri üretim panelindeki kartta
gösterilir. Kışla ve limanın ikisi de yoksa düğme pasif olur.

`WorldX/WorldY` bölge geometrisi ve Voronoi ayrımı için korunur. Haritadaki şehir noktaları `Settlements` üzerinden çizilir; ana yerleşim `is_center` ile seçilir. Bu alan hiçbir yerleşimde yoksa runtime merkezi sırasıyla kale, şehir, kasaba ve liman tiplerinden seçer ve seçimi `is_center` olarak tamamlar. `settlements` eksikse renderer eski davranışa dönüp bölge adını `WorldX/WorldY` noktasından çizer.

Edit Mode'da seçili bölgeden `Yeni Bölge Ekle` ile oluşturulan yeni kayıt,
seçili bölgenin ekonomik üretimlerini (`base_*`), ticaret kapasitesini ve
nüfusunu kopyalamaz; bu değerler iki bölge arasında toplam korunacak şekilde
bölünür. Tek sayılarda kalan birim yeni bölgeye aktarılır. Kaynak bölgedeki
`population` ve `rural_population` alanları da bölünür; yerleşim kayıtları bu
işlemde değiştirilmez ve yeni bölgeye yerleşim ekleme işi Edit Mode'da ayrıca
yapılır.

```go
type Settlement struct {
    ID         string
    NameTR     string
    X, Y       int     // world_x/world_y ile aynı koordinat uzayı
    Type       string  // city, town, port, fortress
    IsCenter   bool
}

`Region.Population` senaryo verisindeki bağımsız bölge toplam nüfusudur. Yerleşim
noktaları yalnızca harita/başkent/stratejik anchor olarak tutulur; `Settlement`
içinde nüfus alanı yoktur ve `settlements.json` içindeki eski `population`
alanları yüklenirken yok sayılır. Nüfus artışı ve olay etkileri doğrudan bölge
toplamına uygulanır. Eski save'lerde yalnız toplam nüfus varsa yardımcı
`RuralPopulation` alanı bu toplamla başlatılır.
```

Yerleşim koordinatı yanlışlıkla bölge raster alanının dışına düşerse render cache yüklenirken uyarı loglanır ve nokta aynı region içindeki en yakın piksele taşınır.

Kıyı bölgesinde `port` binası tamamlandığında, bölgede henüz `type=port` settlement yoksa oyun bu bölge için denize yakın yeni bir `Liman` yerleşimi üretir. Kıyı adayı aynı bölgedeki mevcut settlement marker'larından en az marker açıklığı kadar ayrılır; tercih edilen nokta doluysa aynı kıyı doğrultusundaki başka bir aday seçilir. Bu akış minor bölgeler için de aynıdır; Edit Mode'da minor alan kıyıya boyandığında görsel deniz komşuluğu state'e aktarılır. Böylece liman binası sadece ekonomi/üretim değil, dock edilen filonun görünür anchor noktası için de tekil veri kaynağı olur.

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

## Boyanmış Arazi Alanları

`terrain_areas.json` içindeki boyanmış alanlar runtime'da `area::<id>` biçiminde
geçici terrain düğümleri olarak oluşturulur. Bu düğümler normal keşif veya
zaman bazlı bölge unlock akışına katılmaz. Geçiş kuralı yalnızca alanın
`move_cost` değeridir: `0` geçilmez, sıfır dışındaki değerler geçilebilir ve
ek hareket maliyetini belirtir. Save yüklenirken terrain düğümünün kilit durumu
senaryo `TerrainAreas` verisinden yeniden türetilir; eski save içindeki normal
bölge kilidi bu değeri değiştiremez.

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
