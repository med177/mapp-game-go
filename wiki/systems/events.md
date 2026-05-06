---
type: system
tags: [events, historical, trigger, notification]
last_updated: 2026-05-06
related: [world/regions, architecture/game-loop]
---

# Tarihsel Olaylar Sistemi

**Kaynak:** `internal/events/events.go`, `assets/data/events.json`

## Olay Yapısı

Olaylar JSON'dan yüklenir ve `events.LoadEvents()` ile `[]*Event` listesine dönüştürülür.

Her olay bir kez tetiklenir: `gs.FiredEventIDs` map'i ile takip edilir.

---

## Tetikleme Koşulları

`events.Tick(gs, evts)` — tur çözümleme sırasında çağrılır.

Tetikleme kriterleri:
- **Yıl/ay:** Gerçek tarihe yakın dönem (`Year >= X`)
- **Bölge sahipliği:** Belirli bölgeler belirli fraksiyona ait olmalı
- **Fraksiyon durumu:** Güç, altın, bölge sayısı eşikleri

---

## Olay Tipleri

| Tip | Efekt |
|---|---|
| Veba | Bölge nüfus/üretim düşüşü, komşulara yayılma riski |
| Kıtlık | Tahıl üretimi sıfır, isyan riski artar |
| Taht krizi | Fraksiyon içi isyan veya geçici zayıflık |
| Suikast | Lider/komutan kaybı |
| Dini hareket | Reformasyon, mezhep çatışması |
| Keşif | Yeni bölge açılımı tetikleyici |

---

## Bildirim

Olay tetiklendiğinde:
- `renderer.ShowCombatResult("📜 " + olay.Adı + ": " + açıklama)` → kısa bildirim
- Büyük olaylar için `ShowHistoricalEvent(başlık, açıklama)` → tam ekran popup

---

## Eklenecek / Planlanmış

- [ ] Olayların zincirleme tetiklenmesi (veba → kıtlık)
- [ ] Bölge ikonu gösterimi (harita üzerinde ❗ ikon)
- [ ] Oyuncu tepki seçenekleri (A / B karar pop-up)
