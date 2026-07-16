---
type: system
tags: [seasons, time, month, year, weather]
last_updated: 2026-07-16
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
- Hareket havuzu `ArmyMaxMovePoints()` ile hesaplanır: önce ordudaki en yavaş
  birimin `UnitType.MovementPoints` değeri alınır, sonra mevsim çarpanı uygulanır;
  komutan, teknoloji ve zorluk bonusları bu iklimlendirilmiş tabana eklenir.
- Senaryo hareket değerleri: süvari `3`, piyade `2`, kuşatma/topçu `1`;
  karışık kara ordusu her zaman en düşük değeri kullanır. Taşınan kara birlikleri
  filonun hızını etkilemez.
- İlkbahar `%110`, yaz `%100`, sonbahar `%95`, kış `%70` hareket çarpanı uygular;
  sonuç en az `1` puandır.
- Sonbahar: Gelir çarpanı
- Kış dışı turlar: Kara orduları kendi kara toprağında toparlanır; donanmalar ise kendi veya müttefik limanına bağlı (`DockedRegionID`) durumdaysa toparlanır. Kendi limanında gemiler ve taşınan kara birlikleri `+10 HP`, müttefik limanında ise `+5 HP` alır
- Gemide kara birimi taşıyan filolar limana uğramadan uzun süre açık denizde kalırsa `turns_without_port` sayacı işler; 3 turluk emniyet penceresinden sonra taşınan birlikler her tur artan HP zayiatı alır ve limana bağlanınca sayaç sıfırlanır

→ Çözümleme sırası için [[architecture/game-loop]]

---

## Görsel Değişim

Harita arka planı mevsime göre değiştirilmesi planlanıyor (kar kaplı, yeşil, vb.) — şu an statik arka plan.
