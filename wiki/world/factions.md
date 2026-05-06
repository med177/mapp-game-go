---
type: world
tags: [factions, religion, diplomacy, starting-positions]
last_updated: 2026-05-06
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
    ID       FactionID
    NameTR   string
    Religion Religion      // Catholic | Orthodox | SunniIslam | ShiaIslam
    Color    [3]uint8      // harita rengi (RGB)
    Gold     int
    Grain    int

    Research ResearchState
}
```

---

## Din Sistemi

`Religion` — `internal/faction/faction.go`

| Din | Fraksiyonlar |
|---|---|
| `Catholic` | Fransa, İngiltere, Venedik, Aragon, Portekiz |
| `Orthodox` | Rusya |
| `SunniIslam` | Osmanlı, Memlük |
| `ShiaIslam` | Safevi |

**Diplomatik etkiler:**
- Aynı din → ilişki bonusu (planlanmış, şu an uygulanmamış)
- Farklı din → kalıcı ceza çarpanı (planlanmış)

**Mezhep değişimi:** Ele geçirilen bölge `ConversionProgress` sayacıyla yıllar içinde yeni sahip dinine geçer. → [[world/regions]]

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — tüm çiftler `StancePeace, Score=0` başlar.

Tarihsel düşmanlıklar (Osmanlı–Safevi, Osmanlı–Memlük, İngiltere–Fransa) şu an JSON'a hardcode edilmemiş; ileride `factions.json`'a `initial_relations` alanı eklenebilir.

→ İlişki sistemi: [[systems/diplomacy]]

---

## Eleme Koşulu

`gs.IsEliminated(fid)` → `len(RegionsOwnedBy(fid)) == 0`

Bir fraksiyon tüm bölgelerini kaybedince `checkEliminations()` tarafından tespit edilir ve `FactionsEliminated` sayacı artar.
