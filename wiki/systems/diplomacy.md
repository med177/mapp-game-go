---
type: system
tags: [diplomacy, relations, stance, faction]
last_updated: 2026-05-25
related: [world/factions, systems/ai, architecture/state-management]
---

# Diplomasi Sistemi

**Kaynak:** `internal/diplomacy/diplomacy.go`, `internal/faction/faction.go`, `internal/game/game.go`

## İlişki Yapısı

```go
type Relation struct {
    FactionA, FactionB FactionID
    Score   int              // -100 (düşman) → +100 (müttefik)
    Stance  DiplomaticStance
}
```

`RelationKey(a, b)` → her zaman sıralı `"a|b"` string'i üretir (çift kayıt önler).

---

## Diplomatik Duruşlar (DiplomaticStance)

| Duruş | Geçiş Koşulu | Puan Etkisi |
|---|---|---|
| `StancePeace` | Varsayılan / barış sonrası | Score = -20 |
| `StanceWar` | Savaş ilan edildiğinde | Score = -80 |
| `StanceTrade` | Ticaret anlaşması | Score +15 |
| `StanceAllied` | İttifak | Score +20 |

**Geçiş kısıtları:**
- Savaştayken ittifak veya ticaret kurulamaz
- İttifak için `Score >= 20` gerekir
- Ticaret için `Score >= -35` gerekir ve iki tarafın da kara bölgesi olmalıdır
- Zaten aynı duruştaysa tekrar kurulamaz

---

## Oyuncu Diplomatik Aksiyonları

`internal/game/game.go`

| Aksiyon | Fonksiyon | Koşul |
|---|---|---|
| Savaş ilan et | `declareWar()` | Zaten savaşta değilse |
| Barış teklif et | `proposePeace()` | Savaş halinde gerekli; kabul için savaş baskısı + güç dengesi + ekonomik stres değerlendirilir |
| İttifak kur | `proposeAlliance()` | Savaşta değil + `Score >= 20` + doğrudan sınır tehdidi olmamalı |
| Ticaret anlaşması | `proposeTrade()` | Savaşta değil + `Score >= -35` + iki tarafın da kara bölgesi var |

Teklifler artık otomatik kabul edilmez; oyuncu ve AI aynı değerlendirme motorunu kullanır.

---

## İlişki Puanı Değişimleri

| Olay | Puan Değişimi |
|---|---|
| Savaş ilanı | -80 (sabit) |
| Barış | -20 (sıfırlama) |
| Ticaret | +15 |
| İttifak | +20 |
| `ApplyRelationDecay()` | Savaşta skor düşer; barış/ticaret/ittifakta ilişki yumuşar |
| Ortak düşman | +bonus (AI koalisyon mantığında) |
| Din bonusu/cezası | `religion.Relation(a,b)` — başlangıç skoru; +25 / -20 / -30 / -40 |

→ `applyRelationDecay` tur çözümleme sırası: [[architecture/game-loop]]

---

## Ticaret Entegrasyonu

`GameState.TradeRoutes` artık diplomasi motoru tarafından yönetilir.

- Ticaret anlaşması kabul edilince iki yönlü rota oluşturulur
- Aynı iki fraksiyon için rota çoğaltılmaz; mevcut çift önce temizlenir
- Savaş ilanı veya barış kabulü iki taraf arasındaki aktif rotaları kapatır
- Rotalar soyut anlaşma modelidir; harita üstü pathfinding ile üretilmez

Rota detayları:

- Mal türü gönderen fraksiyonun en değerli mevcut stokuna göre seçilir
- `AmountPerTurn`, iki tarafın toplam `trade_capacity` değerinden türetilir
- Altın getirisi tur çözümlemesinde `TradeRoute.GoldEarned()` ile hesaplanır

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — `internal/faction/loader.go`

Tüm fraksiyon çiftleri için skor `internal/religion.Relation()` sonucuyla başlatılır. Varsayılan duruş barıştır; Sünni-Şii çiftleri başlangıçta savaş durumuna alınır.

---

## AI Diplomasi Davranışı

`aiHandleDiplomacy()` ve `FormCoalitionAgainstPlayer()` — zorluk 3 koalisyon dahil aynı motoru kullanır

AI:

- uzun savaşta ve zayıf kaldığında barış dener
- ortak düşman + yeterli skor varsa ittifak dener
- barışta ve skor uygunsa ticaret açar
- koalisyon anında oyuncuya savaş açıp diğer AI'larla ittifak kurmaya çalışır

→ Detaylar: [[systems/ai]]

---

## Elenen Fraksiyon Temizliği

`internal/game/resolution.go:255` içindeki `checkEliminations()` artık bir fraksiyonun bölgesi kalmadığında:

- `IsEliminated=true` işaretler
- o fraksiyona ait tüm orduları kaldırır
- `GameState.Relations` içindeki o fraksiyonu içeren tüm diplomasi kayıtlarını siler

Bu sayede elenen devletler diğer devletlerle diplomasi verisi taşımaya devam etmez.

---

## Gelen Teklif Paneli (Oyuncu)

AI artık oyuncuya doğrudan barış sonucu dayatmaz. Savaş baskısı şartı oluştuğunda teklif `GameState.DiplomaticOffers` kuyruğuna eklenir:

- kaynak: `internal/ai/ai.go:87`
- kuyruk/çözümleme: `internal/diplomacy/offers.go`
- UI paneli: `internal/render/renderer.go` (`drawDiplomacyOfferDialog`, `handleDiplomacyOfferInput`)

Oyuncu teklif geldiğinde `Kabul Et` veya `Reddet` yanıtı verir; kabulde standart diplomasi motoru (`Execute`) çalışır, redde ise teklif kuyruktan düşer ve savaş sürer.

---

## Eksik / Planlanan

- [ ] Bekleyen diplomatik teklif kuyruğu / çok adımlı müzakere
- [ ] İttifak için ortak geçiş hakkı veya askeri bonuslar
- [ ] Ticaret için dinamik piyasa / rota pathfinding
- [x] `internal/religion` paketi ayrıştırıldı
