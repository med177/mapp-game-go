---
type: system
tags: [economy, gold, tax, trade, buildings]
last_updated: 2026-05-06
related: [systems/seasons, world/regions, architecture/game-loop]
---

# Ekonomi Sistemi

**Kaynak:** `internal/economy/economy.go`, `internal/city/building.go`

## Kaynaklar

| Kaynak | Tür | Açıklama |
|---|---|---|
| Düka Altın | Birincil | Her şey altına çevrilir |
| Tahıl | İkincil | Ordu besleme, kıtlık riski |
| Demir | İkincil | Ordu kalitesi |
| Kereste | İkincil | Bina inşa |
| Baharat | İkincil | Ticaret geliri |
| Kumaş | İkincil | Ticaret geliri |

Şu an altın mekanik olarak aktif; ikincil mallar veri modelinde var ama gelir hesabına tam entegre edilmemiştir.

---

## Vergi Sistemi

Her bölgede `TaxRate` (0–100) ayarlanabilir.

Oyuncu: `.` tuşu +5, `,` tuşu -5 → `adjustTax()` — `internal/game/game.go:557`

| Vergi Oranı | Etkisi |
|---|---|
| Düşük (0–30) | Yüksek memnuniyet, isyan riski düşük |
| Orta (30–60) | Dengeli |
| Yüksek (60–100) | Fazla altın, memnuniyet düşer, isyan riski |

**İsyan:** `checkRebellions()` memnuniyet eşiğini kontrol eder → bölge kontrolü kaybedilebilir.

---

## Bina Gelir Etkileri

`assets/data/buildings.json`

| Bina | Tuş | Gelir Etkisi |
|---|---|---|
| Pazar (`market`) | 1 | +altın geliri |
| Çiftlik (`farm`) | 2 | +tahıl üretimi |
| Demirci (`barracks`) | 3 | +ordu eğitim hızı |
| Liman (`port`) | 4 | +deniz birimi, +ticaret |
| Surlar (`walls`) | 5 | +savunma bonusu |
| Tapınak/Kilise/Cami (`temple`) | 6 | +din etkisi, +memnuniyet |

Bina inşası `city.LoadBuildings()` ile yüklenen `Building.GoldCost` kadar altın ister.
Bina `MaxPerRegion` ile sınırlıdır.
Bazı binalar `RequiredTerrain` kısıtı taşır (ör. liman → kıyı).

---

## Ticaret Güzergahları

`TradeRoute` — `internal/economy/economy.go`

```go
type TradeRoute struct {
    FactionA, FactionB FactionID
    GoldPerTurn        int
}
```

Ticaret anlaşması kurulunca aktif olur. `applyEconomyTick()` her tur bu geliri fraksiyonlara ekler.

→ Diplomasi anlaşmaları: [[systems/diplomacy]]

---

## Sonbahar Gelir Bonusu

Sonbahar aylarında (9, 10, 11) `applyEconomyTick()` gelir çarpanı uygular.

→ [[systems/seasons]]

---

## Eksik / Planlanan

- [ ] İkincil mal üretim/tüketim döngüsü
- [ ] Piyasa fiyatı dalgalanması
- [ ] Kıtlık mekaniği (tahıl sıfırlandığında)
- [ ] Ekonomik zafer sayacı (500 altın/tur × 5 tur) tam bağlantısı
