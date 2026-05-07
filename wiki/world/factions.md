---
type: world
tags: [factions, religion, diplomacy, starting-positions]
last_updated: 2026-05-07
related: [systems/diplomacy, world/regions, architecture/state-management]
---

# Fraksiyonlar

**Kaynak:** `internal/faction/faction.go`, `assets/data/factions.json`

## 9 Oynanabilir Fraksiyon

| ID | Ad | Din | Başlangıç Bölgesi | Başlangıç Ordusu |
|---|---|---|---|---|
| `ottoman` | Osmanlı | Sünni İslam | `anatolia` | 5 milis + 2 süvari |
| `france` | Fransa | Katolik | `france` | 4 milis + 1 süvari |
| `england` | İngiltere | Katolik | `england` | 4 milis |
| `venice` | Venedik | Katolik | `venice` | 3 milis + 1 süvari |
| `mamluk` | Memlük | Sünni İslam | `cairo` | 4 milis + 1 süvari |
| `safavid` | Safevi | Şii İslam | `tabriz` | 4 milis + 1 süvari |
| `russia` | Rusya | Ortodoks | `moscow` | 4 milis |
| `aragon` | Aragon | Katolik | `aragon` | 3 milis + 1 süvari |
| `portugal` | Portekiz | Katolik | `portugal` | 3 milis |

---

## Faction Yapısı

```go
type Faction struct {
    ID           FactionID
    Name         string
    NameTR       string
    Religion     Religion      // catholic | orthodox | sunni | shia
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

`Religion` — `internal/faction/faction.go`

| Din | Sabit | Fraksiyonlar |
|---|---|---|
| Katolik | `catholic` | Fransa, İngiltere, Venedik, Aragon, Portekiz |
| Ortodoks | `orthodox` | Rusya |
| Sünni İslam | `sunni` | Osmanlı, Memlük |
| Şii İslam | `shia` | Safevi |

**`ReligionRelation(a, b Religion) int`** — `internal/faction/faction.go:14`

| Kombinasyon | Puan |
|---|---|
| Aynı din | +25 |
| Sünni ↔ Şii | -40 |
| Katolik ↔ Ortodoks | -20 |
| Diğer farklı din | -30 |

Bu puan `BuildInitialRelations()` sırasında ilişki skorlarına eklenir.

**Mezhep değişimi:** Ele geçirilen bölge `ConversionProgress` sayacıyla 24 turda yeni sahip dinine geçer, memnuniyet -20 uygular. → [[world/regions]]

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — tüm çiftler `StancePeace` başlar. Skor, `ReligionRelation()` sonucuyla başlatılır (din bonusu/cezası dahil).

Tarihsel düşmanlıklar (Osmanlı–Safevi, İngiltere–Fransa vb.) şu an `factions.json`'a hardcode edilmemiş; `initial_relations` alanı eklenebilir.

→ İlişki sistemi: [[systems/diplomacy]]

---

## Eleme Koşulu

`gs.IsEliminated(fid)` → `len(RegionsOwnedBy(fid)) == 0`

Bir fraksiyon tüm bölgelerini kaybedince `checkEliminations()` tarafından tespit edilir ve `FactionsEliminated` sayacı artar.
