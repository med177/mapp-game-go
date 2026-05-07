---
type: system
tags: [diplomacy, relations, stance, faction]
last_updated: 2026-05-07
related: [world/factions, systems/ai, architecture/state-management]
---

# Diplomasi Sistemi

**Kaynak:** `internal/faction/faction.go`, `internal/game/game.go:220–302`

## İlişki Yapısı

```go
type Relation struct {
    FactionA, FactionB FactionID
    Score   int              // -100 (düşman) → +100 (müttefik)
    Stance  DiplomaticStance
}
```

`RelationKey(a, b)` → her zaman alfabetik sırada `"a_b"` string'i üretir (çift kayıt önler).

---

## Diplomatik Duruşlar (DiplomaticStance)

| Duruş | Geçiş Koşulu | Puan Etkisi |
|---|---|---|
| `StancePeace` | Varsayılan / barış sonrası | Score = -20 |
| `StanceWar` | Savaş ilan edildiğinde | Score = -80 |
| `StanceTrade` | Ticaret anlaşması | Score +15 |
| `StanceAllied` | İttifak | Score +30 |

**Geçiş kısıtları:**
- Savaştayken ittifak veya ticaret kurulamaz
- İttifak için `Score >= -20` gerekir
- Zaten aynı duruştaysa tekrar kurulamaz

---

## Oyuncu Diplomatik Aksiyonları

`internal/game/game.go`

| Aksiyon | Fonksiyon | Koşul |
|---|---|---|
| Savaş ilan et | `declareWar()` | Zaten savaşta değilse |
| Barış teklif et | `proposePeace()` | Savaş halinde gerekli |
| İttifak kur | `proposeAlliance()` | Savaşta değil + Score ≥ -20 |
| Ticaret anlaşması | `proposeTrade()` | Savaşta değil |

> Şu an tüm teklifler otomatik kabul edilir (basit versiyon). İleride AI kabul/red mantığı eklenecek.

---

## İlişki Puanı Değişimleri

| Olay | Puan Değişimi |
|---|---|
| Savaş ilanı | -80 (sabit) |
| Barış | -20 (sıfırlama) |
| Ticaret | +15 |
| İttifak | +30 |
| `applyRelationDecay()` | Her tur sıfıra doğru çekme |
| Ortak düşman | +bonus (AI koalisyon mantığında) |
| Din bonusu/cezası | `ReligionRelation(a,b)` — başlangıç skoru; +25 / -20 / -30 / -40 |

→ `applyRelationDecay` tur çözümleme sırası: [[architecture/game-loop]]

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — `internal/faction/faction.go`

Tüm fraksiyon çiftleri için `Stance = StancePeace, Score = 0` başlatılır. Tarihsel düşmanlıklar JSON'dan veya başlangıç bonuslarıyla eklenebilir.

---

## AI Diplomasi Davranışı

`FormCoalitionAgainstPlayer()` — zorluk 3'te aktif

Oyuncu `coalitionThreshold = 8` bölgeyi geçince AI fraksiyonları aralarında ittifak kurmaya çalışır.

→ Detaylar: [[systems/ai]]

---

## Eksik / Planlanan

- [ ] AI kabul/red mantığı (ilişki skoruna göre — şu an tüm teklifler otomatik kabul)
- [ ] Komşu bölge tehdidi algısı → puan cezası
- [ ] `religion/` paketi ayrıştırılması (şu an `faction.go` + `resolution.go` içinde inline)
