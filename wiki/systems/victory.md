---
type: system
tags: [victory, win-condition, game-over]
last_updated: 2026-05-06
related: [architecture/state-management, architecture/game-loop]
---

# Zafer Sistemi

**Kaynak:** `internal/victory/victory.go`, `internal/state/state.go:14`

## 4 Zafer Tipi

Oyun başında oyuncu bir zafer koşulu seçer (`PhaseVictorySelect`).

### 1. Toprak Hakimiyeti (`domination`)

```
TargetRegionCount = 20
RequiredRegions   = [constantinople, rome, paris, cairo, jerusalem]
```

20+ bölge **ve** kritik şehirlerin tümünü aynı anda tut.

### 2. Ekonomik Güç (`economic`)

```
TargetGoldIncome = 500  (tur başına net altın geliri)
GoldHoldTurns    = 5    (bu geliri 5 tur boyunca koru)
```

`EconomicVictoryTurns` sayacı `GameState`'te tutulur.

### 3. Askeri Üstünlük (`military`)

```
TargetArmyStrength = 200  (toplam birim gücü)
TargetDefeated     = 3    (elenen fraksiyon sayısı)
```

`FactionsEliminated` sayacı `GameState`'te tutulur.

### 4. Dinî Zafer (`religious`)

```
RequiredRegions = [jerusalem, rome, mecca]
```

3 kutsal şehri aynı anda kontrol et.

`ReligiousVictoryTurns` sayacı `GameState`'te tutulur.

---

## Kontrol Akışı

`victory.Check(gs)` her tur çözümlemesinin sonuna yakın çağrılır.

- Kazanma koşulu sağlandıysa: `gs.WinnerID = gs.PlayerFactionID`, `gs.Phase = PhaseGameOver`
- AI fraksiyonu son bölgesini kaybederse: `checkEliminations()` → `IsEliminated()` kontrolü

**Son şans mekaniği:** Son bölge düşene kadar oyun bitmez.

---

## Zafer Koşulu Uygulama

`applyVictoryChoice(vtype)` — `internal/game/game.go:566`

Seçilen tipe göre `VictoryCondition` struct'ı doldurulur ve `gs.Victory`'ye yazılır.

---

## Deadline

`VictoryCondition.DeadlineTurn = 0` → süresiz. İleride belirli yıla kadar koşul opsiyonu eklenebilir (örn. 1500 yılına kadar domination).
