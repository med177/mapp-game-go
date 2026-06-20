---
type: dev
tags: [capital, settlements, conquest, events, ui]
last_updated: 2026-06-20
related: [world/regions, systems/events, architecture/game-loop, architecture/render-pipeline, dev/progress]
---

# Başkent Sistemi Planı

## Amaç

Yerleşim tabanlı gerçek bir başkent sistemi kurmak:

- Başkent artık bölge değil, belirli bir `settlement` olacak.
- Her fraksiyon kendi başkent settlement'ını `Faction.capital_settlement_id` ile tutacak.
- Başkentin bulunduğu bölge ek avantajlar sağlayacak.
- Başkent fethedildiğinde ekonomi, teknoloji ve otomatik yeniden atama sonuçları doğacak.
- Oyuncu settlement panelinden 5 tur süren bir taşıma süreci başlatabilecek.
- Event'ler de aynı taşıma kuyruğunu tetikleyebilecek.
- Haritadaki tüm başkent settlement'ları yıldız işaretiyle ayırt edilecek.

## Veri Modeli

### Faction

`internal/faction/faction.go`

- `CapitalSettlementID string`
- `PendingCapitalSettlementID string`
- `PendingCapitalTurns int`

Bu modelde kalıcı truth fraksiyon üstündedir. `Settlement.is_capital` görsel/yardımcı veri olarak korunur; bölgesel ana settlement anlamını taşır, ulusal başkent truth'u olmaz.

### Normalizasyon

Yükleme sonrası tek yardımcı akış:

- `Faction.capital_settlement_id` doluysa onu kullan
- boşsa sahip olunan kara bölgeler içinden en yüksek getirili bölgenin ana settlement'ını seç
- pending hedef artık geçersiz/elde değilse kuyruğu temizle

## Oyun Kuralları

### Başkent Bölgesi Bonusları

Başkent settlement'ın bulunduğu bölgeye ek avantaj verilecek:

- ekonomi katkısı
- lojistik / depo katkısı
- ileride başka sistemlere genişletilebilecek tekil helper üzerinden hesap

İlk iterasyonda bonus mantığı merkezi helper'da tutulacak; ham `settlement.IsCapital` kontrolüne dağılmayacak.

### Başkent Düşüşü

Bir fraksiyon başkent settlement'ını içeren bölgeyi kaybederse:

1. Başkent deposundan ele geçen kaynakların belirli bölümü fethedene aktarılır.
2. Savunanın tamamlanmış ama fethedenin sahip olmadığı teknolojilerinin yarısı fethedene anında açılır.
3. Savunanın başkenti otomatik yeniden atanır.

### Otomatik Yeniden Atama

Başkentini kaybeden fraksiyon için yeni başkent:

- sahip olduğu kara bölgeler arasından
- en yüksek efektif getiriye sahip bölge seçilerek
- o bölgenin merkez / ana settlement'ına atanır

Fraksiyonun hiç kara bölgesi kalmadıysa başkent boş kalabilir; eliminasyon akışı bunu ayrıca temizler.

## Başkent Taşıma

### Oyuncu Kararı

Settlement panelinde uygun koşullarda:

- `Başkent Yap` aksiyonu görünecek
- seçim anında doğrudan taşımayacak
- 5 turluk `pending` relocation başlatacak

### Event Kararı

Event effect modeli, doğrudan anlık başkent değişimi yerine aynı pending relocation sistemini başlatabilecek:

- hedef settlement ID
- 5 tur varsayılan bekleme

### Tur Çözümleme

Her tur sonunda:

- `PendingCapitalTurns` azalır
- sıra bittiğinde başkent resmen yeni settlement'a taşınır

## UI ve Render

### Harita

`internal/render/renderer.go`

- Tüm başkent settlement noktalarına mevcut işaretin yanında yıldız çizilecek.
- Bu işaret tüm fraksiyonlar için görünür olacak.

### Settlement Panel

`internal/render/panel.go`

- yerleşimin mevcut başkent olup olmadığı gösterilecek
- pending relocation varsa kalan tur bilgisi gösterilecek
- oyuncunun settlement seçimi üzerinden başkent taşıma aksiyonu sunulacak

## Teknik Temas Noktaları

### Fetih

`internal/game/game.go:applyConquestWithNavalEviction(...)`

Başkent kaybına bağlı:

- loot transferi
- teknoloji transferi
- otomatik yeni başkent ataması

burada koordine edilecek.

### Kuşatma

`internal/game/siege.go`

Kuşatma zaten fetih helper'ına düştüğü için ayrı bir ikinci başkent akışı yazılmayacak.

### Eventler

`internal/events/events.go`

Event effect modeli yeni pending-capital alanlarını taşıyacak.

### Save/Load

`internal/save/save.go`

Yeni fraksiyon alanları JSON save/load zincirine doğal olarak akacak.

## Test Planı

- Başkent yükleme / normalizasyon testi
- Başkent bölgesi bonus testi
- Başkent fetih loot testi
- Başkent fetih teknoloji kazanımı testi
- Otomatik yeniden atama testi
- 5 tur sonra relocation tamamlama testi
- Settlement panel / render yıldız davranışı testi

## Notlar

- Önceki save dosyaları için geri uyumluluk hedeflenmeyecek.
- Senaryo dosyalarında `capital_settlement_id` alanı eksik olsa bile runtime normalizasyonla sistem ayağa kaldırılacak.
- İlk iterasyonda teknoloji transferi yalnız tamamlanmış teknolojiler üzerinden çalışacak.
