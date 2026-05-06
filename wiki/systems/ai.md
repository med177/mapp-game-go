---
type: system
tags: [ai, strategy, coalition, difficulty]
last_updated: 2026-05-06
related: [systems/combat, systems/diplomacy, architecture/game-loop]
---

# Yapay Zeka Sistemi

**Kaynak:** `internal/ai/ai.go`

## Genel Yapı

Her `PhaseAITurn`'de tüm AI fraksiyonları sırayla işlenir:

```go
for fid := range gs.Factions {
    if fid == gs.PlayerFactionID { continue }
    ai.TakeTurn(gs, fid)
}
```

`TakeTurn` iki iş yapar:
1. Zorluk 3 ise → `FormCoalitionAgainstPlayer()`
2. Fraksiyonun tüm ordularını hareket ettir → `moveArmy()`

---

## Ordu Hareketi Mantığı

`moveArmy()` her ordunun hareketi için çağrılır. Seçim önceliği (düşük indeksten yükseğe):

1. Komşu düşman bölgesi varsa → saldır
2. Savunmasız (ordusu olmayan) rakip bölge varsa → ele geçir
3. Sınır güçlendirme / bekleme

AI de oyuncu ile aynı `combat.ResolveBattleWithMods()` fonksiyonunu kullanır; haksız avantajı yoktur (zorluk 3 hariç başlangıç bonusu).

---

## Koalisyon Mantığı

`FormCoalitionAgainstPlayer()` — zorluk 3'te her AI turunun başında çalışır.

**Tetikleme koşulu:** Oyuncunun bölge sayısı `coalitionThreshold = 8`'i geçmesi.

**Etki:** AI fraksiyonlar birbiriyle ittifak kurmaya çalışır → oyuncuya karşı koordineli hareket.

→ İttifak mekanizması: [[systems/diplomacy]]

---

## Teknoloji Modları

`aiTechMods()` — AI de oyuncuyla aynı teknoloji bonuslarını hesaplar:

```go
fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
return TechMods{
    AttackMod:  fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod,
    DefenseMod: fx.LandDefenseMod,
}
```

---

## Zorluk Seviyeleri

| Seviye | Fark |
|---|---|
| 1 (Kolay) | Pasif AI, yavaş büyüme |
| 2 (Normal) | Dengeli strateji |
| 3 (Zor) | +300 başlangıç altın, +100 tahıl; koalisyon mantığı aktif |

Zorluk 3 başlangıç bonusu `resetToNewGame()` içinde uygulanır — `internal/game/game.go:337`

---

## Eksik / Planlanan

- [ ] Ekonomik önceliklendirme (kaynak bölgeleri hedefleme)
- [ ] Diplomatik fırsatçılık (zayıf fraksiyona saldırı)
- [ ] Deniz birimi kullanımı
- [ ] Teknoloji araştırma kararları
- [ ] Bina inşa stratejisi
