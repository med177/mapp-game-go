---
type: architecture
tags: [game-loop, phases, ebitengine, turn-system]
last_updated: 2026-05-08
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
PhaseScenarioSelect
    ↓ SenaryoSeç
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

**Ayrıca:** `PhaseSettings` (ana menüden, ana menüye döner) · `PhasePauseMenu` (ESC ile)

---

## Tur Çözümleme Sırası

`resolveTurn()` — `internal/game/game.go:230`

1. `applySeasonEffects(gs)` — kış hasarı, ilkbahar bonusu → [[systems/seasons]]
2. `applyEconomyTick(gs)` — vergi geliri, ticaret → [[systems/economy]]
3. `applyTechTicks(gs)` — aktif araştırma ilerleme sayacı → [[systems/tech-tree]]
4. `applyProductionTicks()` — bina ve birim üretim kuyruğunu ilerletir; tamamlanan oyuncu üretimleri popup/event log bildirimi üretir
5. `applyReligionConversion(gs)` — ele geçirilmiş bölgelerde yavaş din dönüşümü
6. `checkRegionUnlocks(gs)` — kilitli bölgeleri açma koşulları
7. `checkRebellions(gs)` — düşük memnuniyet → isyan kontrolü
8. `checkEliminations(gs)` — bölgesi kalmayan fraksiyon elenir
9. `applyRelationDecay(gs)` — ilişki puanlarını sıfıra doğru çekme
10. `victory.Check(gs)` — zafer/yenilgi koşulu kontrolü → [[systems/victory]]
11. `events.Tick(gs, evts)` — tarihsel olayları tetikle → [[systems/events]]
12. `gs.AdvanceTurn()` — ay/yıl ilerlet

---

## Oyuncu Aksiyonları (PhasePlayerTurn)

| Aksiyon | Tetikleyici | Açıklama |
|---|---|---|
| `ActionEndTurn` | Enter/Space | AI turuna geç |
| `ActionMoveArmy` | Sağ tık | Orduyu komşu bölgeye taşı / savaş |
| `ActionRecruitUnit` | R | Seçili bölgede milis eğitimini üretim kuyruğuna al |
| `ActionRecruitNaval` | N | Kıyı bölgede nakliye gemisi üretimini kuyruğa al (liman gerekli) |
| `ActionBuild` | 1-6 | market/farm/barracks/port/walls/temple inşaatını kuyruğa al |
| `ActionResearch` | Tech panelinden | Teknoloji araştır |
| `ActionDeclareWar` | Diplomasi paneli | Savaş ilan et |
| `ActionProposePeace` | Diplomasi paneli | Barış teklif et |
| `ActionProposeAlliance` | Diplomasi paneli | İttifak kur |
| `ActionProposeTrade` | Diplomasi paneli | Ticaret anlaşması |
| `ActionSave` / `ActionLoad` | S / L | Kaydet / Yükle |
| `ActionAdjustTax` | . / , | Vergi ±5% |

---

## Başlangıç Orduları

`army.LoadArmies()` — `internal/army/loader.go`

Başlangıç orduları artık kodda üretilmiyor. Her senaryo `data/armies.json` dosyasında ordu ID'si, sahip fraksiyon, başlangıç bölgesi ve birim sayımlarını tanımlar; yükleyici `count` değerlerini tek tek `army.Unit` kayıtlarına açar.
