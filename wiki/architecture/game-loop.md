---
type: architecture
tags: [game-loop, phases, ebitengine, turn-system]
last_updated: 2026-05-06
related: [state-management, render-pipeline]
---

# Oyun Döngüsü & Phase State Machine

**Kaynak:** `internal/game/game.go`

## Ebitengine Entegrasyonu

`Game` struct, Ebitengine'in `ebiten.Game` interface'ini uygular:

```
Update()  → 60 TPS — oyun mantığı
Draw()    → her frame — render
Layout()  → pencere boyutu bildirir
```

`Game` içinde üç bileşen bulunur:
- `gs *state.GameState` — tüm oyun verisi
- `renderer *render.Renderer` — görsel katman
- `evts []*events.Event` — yüklenmiş tarihsel olaylar listesi

---

## Phase State Machine

```
PhaseMainMenu
    ↓ YeniOyun
PhaseFactionSelect
    ↓ FraksiyonSeç
PhaseVictorySelect
    ↓ ZaferKoşuluSeç
PhasePlayerTurn ←──────────────────────┐
    ↓ TurSonu                          │
PhaseAITurn                            │
    ↓ (tüm AI fraksiyonlar işlendi)    │
PhaseTurnResolution                    │
    ↓ (çözüm tamamlandı)               │
    ├─ oyun devam → PhasePlayerTurn ───┘
    └─ oyun bitti → PhaseGameOver
```

**Ayrıca:** `PhaseSettings` (ana menüden, ana menüye döner)

---

## Tur Çözümleme Sırası

`resolveTurn()` — `internal/game/game.go:160`

1. `applySeasonEffects(gs)` — kış hasarı, ilkbahar bonusu → [[systems/seasons]]
2. `applyEconomyTick(gs)` — vergi geliri, ticaret → [[systems/economy]]
3. `applyTechTicks(gs)` — aktif araştırma ilerleme sayacı → [[systems/tech-tree]]
4. `applyReligionConversion(gs)` — ele geçirilmiş bölgelerde yavaş din dönüşümü
5. `checkRegionUnlocks(gs)` — kilitli bölgeleri açma koşulları
6. `checkRebellions(gs)` — düşük memnuniyet → isyan kontrolü
7. `checkEliminations(gs)` — bölgesi kalmayan fraksiyon elenir
8. `applyRelationDecay(gs)` — ilişki puanlarını sıfıra doğru çekme
9. `victory.Check(gs)` — zafer/yenilgi koşulu kontrolü → [[systems/victory]]
10. `events.Tick(gs, evts)` — tarihsel olayları tetikle → [[systems/events]]
11. `gs.AdvanceTurn()` — ay/yıl ilerlet

---

## Oyuncu Aksiyonları (PhasePlayerTurn)

| Aksiyon | Tetikleyici | Açıklama |
|---|---|---|
| `ActionEndTurn` | Enter/Space | AI turuna geç |
| `ActionMoveArmy` | Sağ tık | Orduyu komşu bölgeye taşı / savaş |
| `ActionRecruitUnit` | R | Seçili bölgede milis al (60 altın) |
| `ActionRecruitNaval` | N | Kıyı bölgede nakliye gemisi (200 altın, liman gerekli) |
| `ActionBuild` | 1-6 | market/farm/barracks/port/walls/temple |
| `ActionResearch` | Tech panelinden | Teknoloji araştır |
| `ActionDeclareWar` | Diplomasi paneli | Savaş ilan et |
| `ActionProposePeace` | Diplomasi paneli | Barış teklif et |
| `ActionProposeAlliance` | Diplomasi paneli | İttifak kur |
| `ActionProposeTrade` | Diplomasi paneli | Ticaret anlaşması |
| `ActionSave` / `ActionLoad` | S / L | Kaydet / Yükle |
| `ActionAdjustTax` | . / , | Vergi ±5% |

---

## Başlangıç Orduları

`buildStartingArmies()` — `internal/game/game.go:667`

Her fraksiyon başlangıçta kendi bölgesinde 3–7 birimlik bir orduyla başlar:
- Osmanlı: Anadolu, 5 milis + 2 süvari
- Fransa: France, 4 milis + 1 süvari
- İngiltere: England, 4 milis
- Venedik: Venice, 3 milis + 1 süvari
- Memlük: Cairo, 4 milis + 1 süvari
- Safevi: Tabriz, 4 milis + 1 süvari
- Rusya: Moscow, 4 milis
- Aragon: Aragon, 3 milis + 1 süvari
- Portekiz: Portugal, 3 milis
