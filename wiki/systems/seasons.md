---
type: system
tags: [seasons, time, month, year, weather]
last_updated: 2026-06-06
related: [architecture/game-loop, systems/economy, systems/diplomacy]
---

# Mevsim Sistemi

**Kaynak:** `internal/season/season.go`

## Zaman Yapısı

- **1 tur = 1 ay**
- **12 tur = 1 yıl**
- Başlangıç: Mart 1300

`gs.AdvanceTurn()` → `Month++`; Ocak geçince `Year++`

---

## Mevsimler

`season.FromMonth(month)` — `internal/season/season.go`

| Aylar | Mevsim | Efekt |
|---|---|---|
| 12, 1, 2 | Kış | Ordular her tur birim kaybeder (soğuk hasarı) |
| 3, 4, 5 | İlkbahar | Hareket bonusu |
| 6, 7, 8 | Yaz | Normal |
| 9, 10, 11 | Sonbahar | Hasat → vergi geliri artar |

---

## Tur Çözümlemesindeki Uygulama

`applySeasonEffects(gs)` — tur çözümleme sırasında **ilk** çalışır.

- Kış: Her ordu için birim hasar kontrolü
- İlkbahar: `MovePoints` bonusu (şu an taslak — detay `season.go`'da)
- Sonbahar: Gelir çarpanı
- Kış dışı turlar: Kara orduları kendi kara toprağında toparlanır; donanmalar ise kendi veya müttefik limanına bağlı (`DockedRegionID`) durumdaysa aynı `+10 HP` toparlanmayı alır

→ Çözümleme sırası için [[architecture/game-loop]]

---

## Görsel Değişim

Harita arka planı mevsime göre değiştirilmesi planlanıyor (kar kaplı, yeşil, vb.) — şu an statik arka plan.
