---
type: architecture
tags: [render, editor, shapes, country-shapes, tooling]
last_updated: 2026-10-05
related: [architecture/render-pipeline, architecture/state-management, dev/data-format, dev/progress]
---

# Shape Editor

`Shape Kes` onayında kesim maskeleri yalnız ilgili shape ve polygon sınırları
için oluşturulur. Kesilen pikseller mevcut harita raster cache'inde yeni
bölgeye aktarılır; bu işlem tam shape rasterizasyonunu ve deniz BFS'ini yeniden
çalıştırmaz.

`Shape Birleştir` onayında iki shape'in mevcut raster cache'leri birleştirilir
ve yalnız birleşik piksel kümesi yeni ortak shape bölgelerine göre yeniden
bölgelendirilir. Diğer shape'ler ve deniz rasterı yeniden hesaplanmaz.

Shape boya commit'i de hedef shape için ayrı bir geometri rebuild yolu kullanır:
eski pikseller deniz baseline'ına döndürülür, yeni ring'ler scanline raster ile
oluşturulur ve diğer shape'ler korunur. Merkez değişikliklerinde geometri tekrar
rasterize edilmeden mevcut shape piksel cache'i kullanılır.

Shape sekmesindeki `Haritayı Yenile` düğmesi gerektiğinde tam `WorldMap`
oluşturur ve eski sınır, seçim, Voronoi ve etiket cache'lerini temizler. Worker
sonrasında yeni harita kabul edildiğinde aynı cache temizleme akışı otomatik
uygulanır.

Ülke ring'leri ve shape edit maskeleri scanline span'leriyle doldurulur. Böylece
her piksel için polygonun tüm kenarlarını tekrar test eden yol kullanılmaz.

`country_shapes.json` artık sadece dış araçlarla değil, oyun içi edit mode üzerinden de düzenlenebilir.

## Problem

Voronoi seed region düzenleme oyunda yapılabiliyordu; fakat gerçek kıyı/ülke alanını belirleyen `data/country_shapes.json` hâlâ elle veya `tools/` scriptleriyle değiştiriliyordu. Bu, küçük kıyı düzeltmeleri ve eksik ada/çıkıntı eklemelerini yavaşlatıyordu.

## MVP hedefi

Edit mode inspector içine üçüncü bir `Shape` sekmesi eklenir.

- seçili region'ın `shape_id` değeri okunur
- aynı `shape_id` paylaşan tüm region'ların ortak country shape'i düzenlenir
- sağ mouse ile boya/sil fırçası uygulanır
- mouse bırakılınca yalnız geçici önizleme tutulur; `Uygula` ile mask → ring dönüşümü yapılır
- `Ctrl+S` / `Kaydet` akışı `country_shapes.json` dosyasını da yazar
- shape değişiklikleri doğrudan runtime state'e yazılır ve `editDirty` ile kayda girer

## Veri akışı

1. `world.LoadCountryShapes()` JSON'u `GameState.ShapeData` içine yükler.
2. Shape tab açılınca seçili `shape_id` için raster mask oluşturulur.
3. Brush bu mask üzerinde add/erase yapar.
4. `Uygula` tıklanınca mask grid sınırlarından polygon ring'leri yeniden üretilir.
5. Yeni ring'ler hem `GameState.ShapeData.Shapes[shape_id]` hem ilgili `Region.Shape` alanlarına geri yazılır.
6. `rebuildEditWorldMap()` ile harita cache'i yeniden üretilir.
7. Senaryo kaydında `writeScenarioShapes()` ile `data/country_shapes.json` güncellenir.

Shape commit sonrasında worker mevcut harita cache'ini korur; yalnızca hedef
shape'in eski alanını deniz baseline'ına döndürür ve yeni ring'lerini rasterize
eder. Böylece her `Uygula` işleminde tüm shape'ler ile deniz BFS'i yeniden
hesaplanmaz.

`Yeni Kara Sınırı` düğmesi yalnızca seçili deniz bölgesinde aktif olur. Düğme
küçük bir modal açar; editör modal içinde sırasıyla `Shape ID` ve `Shape Adı`
ister. Her iki adımda Enter veya `OK` ile ilerlenebilir, Escape/`İptal` ile
vazgeçilebilir. ID boşluk içeremez ve mevcut bir
shape ile çakışamaz. Onaylandığında son tıklanan deniz hücresinde küçük bir
başlangıç halkasıyla yeni bir `IsSea=false` kara bölgesi ve ona bağlı shape
oluşturulur. Yeni bölge hemen `Sınır Boya/Sil` araçlarıyla genişletilebilir;
shape, bölge ve bölge sırası değişiklikleri doğrudan runtime state'e yazılır ve
`Kaydet` ile `country_shapes.json` içine yazılır.

`Shape Kes` aracı, seçili kara shape'inin içinden sol tıkla kapatılan bir poligon
seçer. Seçim dünya raster hücreleriyle mevcut shape maskesine kısaltılır; seçilen
hücreler yeni shape'e aktarılır ve ana shape'den çıkarılır. `Shape ID` ve `Shape Adı`
modalından alınan bilgiler `ShapeData`'ya yazılır. Yeni shape'in haritada
kullanılabilmesi için kaynak bölgeden türetilen yeni bir kara bölgesi de oluşturulur;
kesim ana shape'in tamamını kapsıyorsa işlem reddedilir. Poligon, ilk noktaya
tıklanarak veya `Kes ve Kaydet` düğmesiyle modalı açar; sağ tık/Escape taslağı
iptal eder.

`Shape Birleştir` aracıyla mevcut seçili kara shape'i birinci/hedef olarak alınır.
Mod açıldıktan sonra haritadaki ikinci kara shape'e tıklanır. İki shape'in raster
maskeleri birleşik ring'lere çevrilir; ikinci shape'e bağlı bölgelerin `ShapeID`
değeri hedef shape'e taşınır ve ikinci shape kaydı silinir. Escape, sağ tık veya
`Birleştirmeyi İptal` düğmesi seçim modunu kapatır.

`Bolge Boya/Sil` performans notu:
- Stroke sırasında `regionAt` canlı olarak güncellenir ama ağır `regionPx` dilim bakımı mouse hareketi başına yapılmaz; bu toplu indeks yenilemesi rebuild aşamasına bırakılır.
- `region_shapes.json` içindeki `minor_polygons` yüklendiğinde world map
  override öncesi `baseRegionAt` snapshot'ı alınır; poligonlar yalnızca bu
  temel raster üzerinde runtime piksel atamalarına çevrilir. Edit mode erase
  baseline'ı ikinci bir tam world map kurmadan bundan okunur.
- Yeni minor bölge kimlikleri `new_minor_region_*` önekiyle üretilir. Çizim
  tamamlanmadan merkez işareti oluşturulmaz; poligon commit edildiğinde merkez
  geometriden hesaplanıp senaryo koordinatlarına çevrilir.
- Bölge sekmesindeki `İmtiyazlı: Evet/Hayır` düğmesi yalnız minor bölgelerde
  görünür. `Evet` durumunda egemen sahip parent bölgeden çözülür, mevcut
  `OwnerID` kullanım sahibi olarak kalır; ayrı bir privilege-owner alanı
  oluşturulmaz.
- `Shape/Bolge Boya/Sil` canlı preview'i stroke sırasında world-space önizleme görüntüsüne artımlı işlenir; her frame'de etkilenen piksel listesi ve yeni UI overlay'i yeniden çizilmez. Region tool büyük country mask'ini lazy tutar.
- Boyama koordinatı raster hücresinin bulunduğu aralığa `floor` ile atanır;
  önizleme ve fırça imleci aynı hücrenin merkezini (`x+0.5`, `y+0.5`) kullanır.
  Böylece işaretlenen nokta ile boyanan piksel arasında yarım hücrelik sol-üst
  kayma oluşmaz.
- Shape konturu da world raster üretimindeki aynı ölçekleme ve kesme sırasını
  kullanır; böylece kontur, renkli raster alanının kenarından ayrışmaz.
- Shape maskesi dünya piksel çözünürlüğünde tutulur; commit sırasında hücre
  sınırları tekrar shape senaryo koordinatlarına çevrilir. Böylece en küçük
  fırça doğrudan görünen nokta/çizgi karesini boyar veya siler.

## UX kuralları

- Sol tık seçim davranışını korur.
- Arazi alanı ekleme ve düzenleme düğmeleri `Arazi` sekmesinde toplanır; haritadan
  bir arazi alanı seçildiğinde bu sekme otomatik olarak aktifleşir. `Harita`
  sekmesi shape ve karasal geçiş araçlarını barındırır.
- Shape düzenleme `Shape` sekmesinde ve **sağ mouse drag** ile yapılır; böylece region seçimiyle çakışmaz.
- `Boya` ve `Sil` modları inspector butonlarından değişir.
- Fırça yarıçapı inspector'dan artırılıp azaltılır; `1.00` altına iki ince kademe
  (`0.75` ve `0.50`) bulunur. Yarıçap değeri dünya pikseli cinsindendir; bu
  seviyeler bölge ve shape araçlarında aynı fiziksel fırça boyutunu hedefler.
  Harita editinde `Ctrl+tekerlek` ile de büyütülüp küçültülebilir.
- Brush stroke sırasında imleç yarıçapı ekranda gösterilir.
- Stroke sırasında eklenen alanlar yeşil, silinen alanlar kırmızı preview overlay ile gösterilir.
- Sağ üstte kısa yardım paneli seçili `shape_id`, mod ve kontrol şemasını gösterir.
- Stroke bırakıldığında yalnız önizleme ve bekleyen mask değişikliği tutulur; aktif
  aracın `Uygula` düğmesi commit ve harita yenilemesini tek seferde üretir.
- Dört boya/sil aracından biri seçildiğinde yalnız seçili araç `Uygula` olarak
  kalır; diğer üçü disabled görünür ve input tüketmez. `Uygula` düğmesi yeşil
  çizilir. Uygulama sonrası aktif araç kapanır ve dört araç yeniden erişilebilir
  olur (`internal/render/shape_editor.go`, `map_editor.go`).
- `country_shapes.json` yazımı ring koordinatlarını virgülden sonra tek basamakla
  korur. Editörün dünya-piksel sınırları ölçekli shape koordinatına çevrildiği
  için koordinatları tekrar tam sayıya yuvarlamak yeniden açılışta hassas
  boyama/silme piksellerini kaydırır; kayıt yuvarlaması mevcut raster hücresini
  koruyan en yakın tek ondalık değeri seçer.
- `Harita` sekmesindeki `ID` aksiyonu seçili region kimliğini mevcut değerle
  doldurur; Ctrl+A ile yeni kimlik girilebilir. Boş veya mevcut bir region ile
  çakışan ID reddedilir. Kabul edilen değişiklik region map anahtarını,
  komşuları, geçişleri, ordu/donanma konumlarını, paint override'larını ve
  editor seçim state'i ile birlikte runtime state'te tutulur.
- Deniz region'larında `Yeni Kara Sınırı` düğmesi aktif olur; yeni `shape_id` ve ad girildiğinde seçilen deniz pikselinde küçük bir başlangıç halkası oluşturulur. Yeni shape'e bağlanan deniz bölgesi `Sınır Boya` ile genişletilebilir. `Bölge Boya/Sil` aracı commit sırasında ilgili ülke shape'ini günceller; `region_shapes.json` ise minor bölge poligonları için kullanılır. Kara region'larda `Yeni Kara Sınırı` pasiftir.

## Sınırlamalar

- MVP polygon vertex edit içermez.
- Delik (hole) semantiği için ayrı iç ring authoring UI yoktur; raster mask'ten çıkan bağlı sınırlar kaydedilir.
- Amaç önce küçük kıyı düzeltmeleri, eksik ada/parça ekleme ve kaba sınır boyamayı oyuna taşımaktır.

## Sonraki adımlar

- point/vertex seçip sürükleme
- lasso / fill tool
- shape diff preview
- ayrı island/component listesi
