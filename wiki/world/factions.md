---
type: world
tags: [factions, religion, diplomacy, starting-positions]
last_updated: 2026-05-08
related: [systems/diplomacy, world/regions, architecture/state-management]
---

# Fraksiyonlar

**Kaynak:** `internal/faction/faction.go`, `internal/faction/loader.go`, `internal/religion/religion.go`, `assets/scenarios/<id>/data/factions.json`

## Fraksiyon Verisi

Her aktif senaryo 45 fraksiyon içeriyor; 30 tanesi oynanabilir (`is_playable=true`). Başlangıç orduları fraksiyon dosyasında değil, aynı senaryonun `data/armies.json` dosyasında tutulur.

---

## Faction Yapısı

```go
type Faction struct {
    ID           FactionID
    Name         string
    NameTR       string
    Religion     religion.Type // catholic | orthodox | sunni | shia
    Color        [3]uint8      // harita rengi (RGB)
    IsPlayable   bool
    IsEliminated bool

    Gold   int
    Grain  int
    Iron   int
    Timber int
    Spice  int
    Cloth  int

    Research         ResearchState
    AIAggressiveness int           // AI saldırganlık düzeyi
}
```

---

## Din Sistemi

`religion.Type` — `internal/religion/religion.go`

| Din | Sabit | Fraksiyonlar |
|---|---|---|
| Katolik | `catholic` | Fransa, İngiltere, Venedik, Aragon, Portekiz |
| Ortodoks | `orthodox` | Rusya |
| Sünni İslam | `sunni` | Osmanlı, Memlük |
| Şii İslam | `shia` | Safevi |

**`religion.Relation(a, b religion.Type) int`** — `internal/religion/religion.go`

| Kombinasyon | Puan |
|---|---|
| Aynı din | +25 |
| Sünni ↔ Şii | -40 |
| Katolik ↔ Ortodoks | -20 |
| Diğer farklı din | -30 |

Bu puan `BuildInitialRelations()` sırasında ilişki skorlarına eklenir.

**Mezhep değişimi:** Ele geçirilen bölge `ConversionTurns` sayacıyla 24 turda yeni sahip dinine geçer, memnuniyet -20 uygular. → [[world/regions]]

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — tüm çiftlerin skoru `religion.Relation()` sonucuyla başlatılır. Sünni-Şii çiftleri başlangıçta savaş duruşu alır, diğer çiftler barışta başlar.

Tarihsel düşmanlıklar (Osmanlı–Safevi, İngiltere–Fransa vb.) şu an `factions.json`'a hardcode edilmemiş; `initial_relations` alanı eklenebilir.

→ İlişki sistemi: [[systems/diplomacy]]

---

## Eleme Koşulu

`gs.IsEliminated(fid)` → `len(RegionsOwnedBy(fid)) == 0`

Bir fraksiyon tüm bölgelerini kaybedince `checkEliminations()` tarafından tespit edilir ve `FactionsEliminated` sayacı artar.
