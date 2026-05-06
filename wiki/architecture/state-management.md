---
type: architecture
tags: [state, gamestate, serialize, save-load]
last_updated: 2026-05-06
related: [game-loop, render-pipeline]
---

# State Yönetimi

**Kaynak:** `internal/state/state.go`

## GameState Yapısı

`GameState` tüm oyun verisinin tek kaynağıdır. Save/load bu struct'ı JSON olarak serialize eder.

```go
type GameState struct {
    Turn, Year, Month, StartYear int          // Zaman
    PlayerFactionID               FactionID   // Oyuncu
    Difficulty                    int         // 1=kolay 2=normal 3=zor
    Victory                       VictoryCondition

    Regions   map[RegionID]*Region            // Dünya verisi
    Factions  map[FactionID]*Faction
    Armies    map[ArmyID]*Army
    ShapeData CountryShapeJSON                // json:"-" — kaydedilmez

    UnitTypes     map[string]*UnitType        // json:"-" — assets'ten yüklenir
    BuildingTypes map[string]*Building        // json:"-"
    TechTypes     map[string]*Technology      // json:"-"

    Relations   map[string]*Relation          // Diplomatik ilişkiler
    TradeRoutes []*TradeRoute
    FiredEventIDs map[string]bool             // Tekrar tetiklenmesin diye

    Phase    Phase                            // State machine pozisyonu
    WinnerID FactionID                        // Boşsa oyun devam ediyor
}
```

---

## Runtime-Only Alanlar (`json:"-"`)

Bu alanlar JSON'a yazılmaz; oyun her başladığında assets'ten yeniden yüklenir:

| Alan | Yükleme kaynağı |
|---|---|
| `UnitTypes` | `assets/data/units.json` |
| `BuildingTypes` | `assets/data/buildings.json` |
| `TechTypes` | `assets/data/technologies.json` |
| `ShapeData` | `assets/data/generated/country_shapes.json` |

**Neden bu ayrım?** Tanım verisi değişmez — onu kayıt dosyasına koymak gereksiz ve kırılgan. Sadece *durum* (kim neye sahip, ne araştırdı) kaydedilir.

---

## Yardımcı Metodlar

`CurrentSeason() Season` — `season.FromMonth(s.Month)` ile mevsimi döner → [[systems/seasons]]

`AdvanceTurn()` — `Turn++`, `Month++`, Ocak geçince `Year++`

`RegionsOwnedBy(fid) []*Region` — fraksiyon bölge listesi

`IsEliminated(fid) bool` — bölgesi yoksa `true`

---

## Veri Yükleme Akışı

`loadGameState()` — `internal/game/game.go:614`

```
world.LoadRegions()           → regions.json
world.LoadCountryShapes()     → generated/country_shapes.json
faction.LoadFactions()        → factions.json
army.LoadUnitTypes()          → units.json
city.LoadBuildings()          → buildings.json
tech.LoadTechnologies()       → technologies.json
faction.BuildInitialRelations() → ilişki map'i oluştur
buildStartingArmies()         → başlangıç orduları
```

---

## Zafer Koşulu Yapısı

```go
type VictoryCondition struct {
    Type               VictoryType      // domination | economic | military | religious
    TargetRegionCount  int              // domination: 20+ bölge
    RequiredRegions    []RegionID       // domination: constantinople, rome, paris, cairo, jerusalem
    TargetGoldIncome   int              // economic: 500 altın/tur
    GoldHoldTurns      int             // economic: 5 tur boyunca koru
    TargetArmyStrength int             // military: 200 güç puanı
    TargetDefeated     int             // military: 3 fraksiyon yenilgisi
    DeadlineTurn       int             // 0 = süresiz
}
```

Detaylar → [[systems/victory]]

---

## Phase Listesi

```go
PhaseMainMenu       // ana menü
PhaseSettings       // ayarlar ekranı
PhaseFactionSelect  // fraksiyon seçim
PhaseVictorySelect  // zafer koşulu seçim
PhasePlayerTurn     // oyuncu aksiyonları
PhaseAITurn         // AI tur işlemi
PhaseTurnResolution // tur çözümleme
PhaseGameOver       // oyun sonu
```

→ Geçiş diyagramı için [[architecture/game-loop]]
