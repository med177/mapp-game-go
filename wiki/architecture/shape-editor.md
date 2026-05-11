---
type: architecture
tags: [render, editor, shapes, country-shapes, tooling]
last_updated: 2026-05-11
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

## UX kuralları

- Sol tık seçim davranışını korur.
- Shape düzenleme `Shape` sekmesinde ve **sağ mouse drag** ile yapılır; böylece region seçimiyle çakışmaz.
- `Boya` ve `Sil` modları inspector butonlarından değişir.
- Fırça yarıçapı inspector'dan artırılıp azaltılır.
- Brush stroke sırasında imleç yarıçapı ekranda gösterilir.
- Stroke sırasında eklenen alanlar yeşil, silinen alanlar kırmızı preview overlay ile gösterilir.
- Sağ üstte kısa yardım paneli seçili `shape_id`, mod ve kontrol şemasını gösterir.
- Stroke commit'i mouse bırakıldığında yapılır; bu sırada undo snapshot alınır.

## Sınırlamalar

- MVP polygon vertex edit içermez.
- Delik (hole) semantiği için ayrı iç ring authoring UI yoktur; raster mask'ten çıkan bağlı sınırlar kaydedilir.
- Amaç önce küçük kıyı düzeltmeleri, eksik ada/parça ekleme ve kaba sınır boyamayı oyuna taşımaktır.

## Sonraki adımlar

- point/vertex seçip sürükleme
- lasso / fill tool
- shape diff preview
- ayrı island/component listesi
