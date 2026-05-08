---
type: system
tags: [victory, win-condition, game-over]
last_updated: 2026-05-08
related: [architecture/state-management, architecture/game-loop]
---

# Zafer Sistemi

**Kaynak:** `internal/victory/victory.go`, `internal/state/state.go:14`

## Zafer Tipleri

Oyun başında oyuncu bir zafer koşulu seçer (`PhaseVictorySelect`).

### 1. Toprak Hakimiyeti (`domination`)

```
TargetRegionCount = 20
RequiredRegions   = [constantinople, rome, paris, cairo, jerusalem]
```

20+ bölge **ve** kritik şehirlerin tümünü aynı anda tut.

### 2. Ekonomik Güç (`economic`)

```
TargetGoldIncome = 500
GoldHoldTurns    = 5    (bu geliri 5 tur boyunca koru)
```

`EconomicVictoryTurns` sayacı `GameState`'te tutulur. Dikkat: mevcut kod `TargetGoldIncome` alanını isim/metin aksine tur başı gelir olarak değil, fraksiyonun mevcut altını (`f.Gold`) olarak kontrol ediyor. Bu uyumsuzluk [[dev/progress]] içinde kritik takip işidir.

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

3 kutsal şehri 12 tur boyunca kontrol et.

`ReligiousVictoryTurns` sayacı `GameState`'te tutulur.

---

## Kontrol Akışı

`victory.Check(gs)` her tur çözümlemesinin sonuna yakın çağrılır.

- Kazanma koşulu sağlandıysa: `gs.WinnerID = gs.PlayerFactionID`, `gs.Phase = PhaseGameOver`
- AI fraksiyonu son bölgesini kaybederse: `checkEliminations()` → `IsEliminated()` kontrolü

**Son şans mekaniği:** Son bölge düşene kadar oyun bitmez.

---

### Senaryo Özel Hedef (`conquer_city`)

Senaryo JSON'larında `conquer_city` tipi var ve `applyVictoryChoice()` bunu tek hedef bölgeyi `RequiredRegions` listesine çevirerek state'e yazar. Ancak `internal/victory/victory.go` henüz bu tipi `Check()` içinde ele almıyor. Bu yüzden `conquer_city` şu an seçilebilir ama kazanımı tetiklemeyen kritik eksiktir.

Ek veri riski: mevcut senaryo hedefleri `CON`, `ROM`, `CAI`, `PAR`, `JER`, `MEC` gibi kısa ID'ler kullanıyor; `regions.json` içinde bölge ID'leri `constantinople`, `paris`, `london` gibi uzun formda. Zafer sistemi düzeltilirken senaryo hedefleri de gerçek region ID'leriyle eşitlenmeli.

---

## Zafer Koşulu Uygulama

`applyVictoryChoice(optionID)` — `internal/game/game.go`

Seçilen tipe göre `VictoryCondition` struct'ı doldurulur ve `gs.Victory`'ye yazılır.

---

## Deadline

`VictoryCondition.DeadlineTurn = 0` → süresiz. İleride belirli yıla kadar koşul opsiyonu eklenebilir (örn. 1500 yılına kadar domination).
