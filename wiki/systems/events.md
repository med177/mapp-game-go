---
type: system
tags: [events, historical, trigger, notification]
last_updated: 2026-06-03
related: [world/regions, architecture/game-loop, architecture/render-pipeline]
---

# Tarihsel Olaylar Sistemi

**Kaynak:** `internal/events/events.go`, `assets/scenarios/*/data/events.json`

## Olay Yapısı

Olaylar JSON'dan yüklenir ve `events.LoadEvents()` ile `[]*Event` listesine dönüştürülür.

Her olay `events.Tick()` içinde yalnızca **tespit edilir**; uygulama sonrasında `events.Apply()` / `events.ApplyChoice()` ile yapılır. Tek seferlik event'ler `gs.FiredEventIDs` içinde takip edilir.

---

## Tetikleme Koşulları

`events.Tick(gs, evts)` — tur çözümleme sırasında çağrılır.

Tetikleme kriterleri:
- **Yıl/ay:** `historical_year` + opsiyonel `historical_month`
- **Rastgele olay:** `probability > 0` ve `min_turn` eşiği
- **Tek seferlik olay:** `one_shot=true` ise tekrar tetiklenmez

---

## Etki Modeli

Base event etkisi eski alanlarla (`target`, `sat_delta`, `gold_delta`, `grain_delta`, `army_hp_mod`, `affected_faction`) tutulur.

Yeni choice katmanı:

```json
{
  "choice_prompt_tr": "Devlet nasıl yanıt verecek?",
  "choices": [
    {
      "label_tr": "Karantina Uygula",
      "desc_tr": "Ticaret daralır ama kayıplar sınırlanır.",
      "ai_weight": 8,
      "effect": {
        "target": "player_faction",
        "gold_delta": -180,
        "sat_delta": 8,
        "army_hp_mod": 0.92,
        "relation_delta_all": -5
      }
    }
  ]
}
```

`relation_delta_all`, etkilenen fraksiyonun tüm aktif ilişkilerine doğrudan score delta uygular; bu olay seçimlerinin diplomasi etkisini taşır.

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
- `renderer.ShowCombatResult("OLAY: ...")` → kısa bildirim
- Tarihsel event varsa `ShowHistoricalEvent(...)` → tam ekran popup
- Choice varsa historical modal içinde A/B butonları açılır ve sonuç ayrı `KARAR` event log kaydı üretir

---

## Mevcut Choice Örnekleri

- `black_death_1347` → `Karantina Uygula` / `Limanları Açık Tut`
- `printing_press_1455` → `Matbaayı Destekle` / `Sansür Uygula`
- `reformation_1517` → `Tolerans Politikası` / `Baskı Uygula`

## Eklenecek / Planlanmış

- [ ] Olayların zincirleme tetiklenmesi (veba → kıtlık)
- [ ] Bölge ikonu gösterimi (harita üzerinde ❗ ikon)
- [ ] Choice sonuçlarını minimap/bölge ikonlarıyla görünür kıl
