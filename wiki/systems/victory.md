---
type: system
tags: [victory, win-condition, game-over]
last_updated: 2026-06-13
related: [architecture/state-management, architecture/game-loop]
---

# Zafer Sistemi

**Kaynak:** `internal/victory/victory.go`, `internal/state/state.go:14`

## Zafer Tipleri

Oyun başında oyuncu bir zafer koşulu seçer (`PhaseVictorySelect`).

### 1. Toprak Hakimiyeti (`domination`)

```
TargetRegionCount = 20
RequiredRegions   = [constantinople, papal_states, egypt, paris, palestine]
```

20+ bölge **ve** kritik şehirlerin tümünü aynı anda tut.

### 2. Ekonomik Güç (`economic`)

```
TargetGoldIncome = 500
GoldHoldTurns    = 5    (bu geliri 5 tur boyunca koru)
```

`EconomicVictoryTurns` sayacı `GameState`'te tutulur. Ekonomik zafer artık oyuncunun mevcut hazinesini değil, o turdaki toplam altın gelirini kontrol eder. Hesaba bölge vergi geliri, bina `gold_mod` çarpanları, aktif ticaret rotaları ve teknoloji `gold_per_region` bonusları dahildir.

### 3. Askeri Üstünlük (`military`)

```
TargetArmyStrength = 200  (toplam birim gücü)
TargetDefeated     = 3    (elenen fraksiyon sayısı)
```

`FactionsEliminated` sayacı `GameState`'te tutulur.

### 4. Dinî Zafer (`religious`)

```
RequiredRegions = [palestine, papal_states, yemen]
```

3 kutsal şehri 12 tur boyunca kontrol et.

`ReligiousVictoryTurns` sayacı `GameState`'te tutulur.

### 5. Hayatta Kalma (`survive_turns`)

````
Turns = 80
````

Oyuncu belirlenen toplam tur sayısına kadar elenmeden ayakta kalırsa zafer kazanır. İlerleme doğrudan `GameState.Turn` üzerinden izlenir.

---

## Kontrol Akışı

`victory.Check(gs)` her tur çözümlemesinin sonuna yakın çağrılır.

- Kazanma koşulu sağlandıysa: `gs.WinnerID = gs.PlayerFactionID`, `gs.Phase = PhaseGameOver`
- AI fraksiyonu son bölgesini kaybederse: `checkEliminations()` → `IsEliminated()` kontrolü

**Son şans mekaniği:** Son bölge düşene kadar oyun bitmez.

---

### Senaryo Özel Hedef (`conquer_city`)

Senaryo JSON'larında `conquer_city` tipi `required_regions` listesini kullanır.

`applyVictoryChoice()` hedef bölge listesini `VictoryCondition.RequiredRegions` içine yazar. `internal/victory/victory.go` bu tipte listedeki tüm hedef bölgeler oyuncuya geçtiğinde zafer verir.

Senaryo hedefleri gerçek `regions.json` ID'leriyle eşleşmelidir; kısa kodlar (`CON`, `ROM`, vb.) kullanılmaz.

### Fraksiyon Bazlı Görünürlük

Senaryo `victory_conditions` kayıtları opsiyonel `allowed_factions` alanı taşıyabilir.

- Alan boşsa: hedef tüm oynanabilir fraksiyonlara gösterilir.
- Alan doluysa: yalnız listelenen fraksiyonlar `PhaseVictorySelect` ekranında bu kartı görür.

Tam senaryo listesi `GameState.ScenarioVictories` içinde saklanır; seçim ekranına gösterilen filtrelenmiş kopya `GameState.AvailableVictories` alanına yazılır. Save/load sırasında `scenario.json` tekrar okunup filtre yeniden uygulanır.

## Zafer Popup

Oyun içi üst-sol zafer HUD kartına tıklanınca merkezde ayrı bir modal açılır.

- Başlık ve açıklama seçilen zafer kartının metadata'sından gelir (`SelectedVictoryOptionID`).
- Bölge tabanlı zaferlerde popup içinde yeşil `✓` / kırmızı `✗` checklist görünür.
- Ekonomik, askerî ve `survive_turns` hedeflerinde checklist durum satırları eşiklere göre güncellenir.

---

## Zafer Koşulu Uygulama

`applyVictoryChoice(optionID)` — `internal/game/game.go`

Seçilen tipe göre `VictoryCondition` struct'ı doldurulur ve `gs.Victory`'ye yazılır.

---

## Deadline

Her zafer kartı opsiyonel olarak `deadline_year` + `deadline_month` taşır.

- Deadline tanımlıysa hedef yalnız oyuncu için geçerlidir.
- Oyuncu hedefi son ay bitmeden tamamlarsa `VictoryAchieved` set edilir ve oyun devam eder.
- Oyuncu deadline ayını da geçirirse oyun kaybedilir.
- AI tarafı zafer kartı seçmez ve hedef tamamladığı için otomatik kazanmaz.
