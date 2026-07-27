---
type: architecture
tags: [render, editor, shapes, country-shapes, tooling]
last_updated: 2026-07-27
related: [architecture/render-pipeline, architecture/state-management, dev/data-format, dev/progress]
---

# Shape Editor

`country_shapes.json` artık sadece dış araçlarla değil, oyun içi edit mode üzerinden de düzenlenebilir.

## Problem

Voronoi seed region düzenleme oyunda yapılabiliyordu; fakat gerçek kıyı/ülke alanını belirleyen `data/country_shapes.json` hâlâ elle veya `tools/` scriptleriyle değiştiriliyordu. Bu, küçük kıyı düzeltmeleri ve eksik ada/çıkıntı eklemelerini yavaşlatıyordu.

## MVP hedefi

Edit mode inspector içine üçüncü bir `Shape` sekmesi eklenir.

- seçili region'ın `shape_id` değeri okunur
- aynı `shape_id` paylaşan tüm region'ların ortak country shape'i düzenlenir
- sağ mouse ile boya/sil fırçası uygulanır
- mouse bırakılınca mask → ring dönüşümü yapılır
- `Ctrl+S` / `Kaydet` akışı `country_shapes.json` dosyasını da yazar
- undo/redo world snapshot içine shape verisini de alır

## Veri akışı

1. `world.LoadCountryShapes()` JSON'u `GameState.ShapeData` içine yükler.
2. Shape tab açılınca seçili `shape_id` için raster mask oluşturulur.
3. Brush bu mask üzerinde add/erase yapar.
4. Stroke bitince mask grid sınırlarından polygon ring'leri yeniden üretilir.
5. Yeni ring'ler hem `GameState.ShapeData.Shapes[shape_id]` hem ilgili `Region.Shape` alanlarına geri yazılır.
6. `rebuildEditWorldMap()` ile harita cache'i yeniden üretilir.
7. Senaryo kaydında `writeScenarioShapes()` ile `data/country_shapes.json` güncellenir.

`Bolge Boya/Sil` performans notu:
- Stroke sırasında `regionAt` canlı olarak güncellenir ama ağır `regionPx` dilim bakımı mouse hareketi başına yapılmaz; bu toplu indeks yenilemesi rebuild aşamasına bırakılır.
- `region_shapes.json` override'ları yüklendiğinde world map override öncesi `baseRegionAt` snapshot'ı alınır; edit mode erase baseline'ı ikinci bir tam world map kurmadan bundan okunur.
- `Shape/Bolge Boya/Sil` canlı preview'i stroke sırasında world-space önizleme görüntüsüne artımlı işlenir; her frame'de etkilenen piksel listesi ve yeni UI overlay'i yeniden çizilmez. Region tool büyük country mask'ini lazy tutar.
- Boyama koordinatı raster hücresinin bulunduğu aralığa `floor` ile atanır;
  önizleme ve fırça imleci aynı hücrenin merkezini (`x+0.5`, `y+0.5`) kullanır.
  Böylece işaretlenen nokta ile boyanan piksel arasında yarım hücrelik sol-üst
  kayma oluşmaz.
- Shape konturu da world raster üretimindeki aynı ölçekleme ve kesme sırasını
  kullanır; böylece kontur, renkli raster alanının kenarından ayrışmaz.

## UX kuralları

- Sol tık seçim davranışını korur.
- Shape düzenleme `Shape` sekmesinde ve **sağ mouse drag** ile yapılır; böylece region seçimiyle çakışmaz.
- `Boya` ve `Sil` modları inspector butonlarından değişir.
- Fırça yarıçapı inspector'dan artırılıp azaltılır; `1.00` altına iki ince kademe
  (`0.75` ve `0.50`) bulunur. Bu seviyeler rasterde tek piksel hassasiyetini
  korurken canlı imleç boyutunu küçültür.
- Brush stroke sırasında imleç yarıçapı ekranda gösterilir.
- Stroke sırasında eklenen alanlar yeşil, silinen alanlar kırmızı preview overlay ile gösterilir.
- Sağ üstte kısa yardım paneli seçili `shape_id`, mod ve kontrol şemasını gösterir.
- Stroke commit'i mouse bırakıldığında yapılır; bu sırada undo snapshot alınır.
- `Harita` sekmesindeki `ID` aksiyonu seçili region kimliğini mevcut değerle
  doldurur; Ctrl+A ile yeni kimlik girilebilir. Boş veya mevcut bir region ile
  çakışan ID reddedilir. Kabul edilen değişiklik region map anahtarını,
  komşuları, geçişleri, ordu/donanma konumlarını, paint override'larını ve
  editor seçim state'ini birlikte taşır; undo/redo world snapshot ile korunur.
- Deniz region'ları `shape_id` taşımadıkları için `Shape Boya/Sil` değil, aynı sekmedeki `Bolge Boya/Sil` aracıyla düzenlenir; bu akış `region_shapes.json` override katmanına yazar ve deniz bölgeleri arasında alan aktarımı yapar. Aynı araç kara region seçiliyken de ülke dış sınırının dışına taşan bölge genişletmelerini kalıcı override olarak saklar; sonraki stroke'larda baseline override öncesi world map'ten üretildiği için daha önce boyanan dış piksel tekrar fırçadan geçti diye kayıttan düşmez.

## Sınırlamalar

- MVP polygon vertex edit içermez.
- Delik (hole) semantiği için ayrı iç ring authoring UI yoktur; raster mask'ten çıkan bağlı sınırlar kaydedilir.
- Amaç önce küçük kıyı düzeltmeleri, eksik ada/parça ekleme ve kaba sınır boyamayı oyuna taşımaktır.

## Sonraki adımlar

- point/vertex seçip sürükleme
- lasso / fill tool
- shape diff preview
- ayrı island/component listesi
