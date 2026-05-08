---
type: architecture
tags: [state, gamestate, serialize, save-load]
last_updated: 2026-05-08
related: [game-loop, render-pipeline]
---

# State Yönetimi

**Kaynak:** `internal/state/state.go`

## GameState Yapısı

`GameState` tüm oyun verisinin tek kaynağıdır. Save/load bu struct'ı JSON olarak serialize eder.

```go
type GameState struct {
    // Zaman
    Turn, Year, Month, StartYear int

    // Senaryo
    ScenarioID   string   // ör. "1300_ottoman_rise"
    ScenarioPath string   // senaryo klasörü tam yolu
    MapConfig    scenario.MapConfig

    // Oyuncu
    PlayerFactionID FactionID
    Difficulty      int       // 1=kolay 2=normal 3=zor
    DevelopmentMode bool

    Victory VictoryCondition

    // Dünya verisi
    Regions   map[RegionID]*Region
    Factions  map[FactionID]*Faction
    Armies    map[ArmyID]*Army
    ShapeData CountryShapeJSON           // json:"-"

    // Runtime-only (json:"-")
    UnitTypes          map[string]*UnitType
    BuildingTypes      map[string]*Building
    TechTypes          map[string]*Technology
    AvailableVictories []VictoryOptionDef  // scenario.json'dan

    // Zafer takibi
    EconomicVictoryTurns  int
    FactionsEliminated    int
    ReligiousVictoryTurns int

    // Diplomatik & ticaret
    Relations     map[string]*Relation
    TradeRoutes   []*TradeRoute
    FiredEventIDs map[string]bool

    ProductionQueue []ProductionOrder // devam eden bina/birim üretimleri
    NextProductionSeq int             // üretim ID sayacı
    NextArmySeq int                   // ordu ID üretici sayaç

    Phase    Phase
    WinnerID FactionID
}
```

`ProductionOrder`, bina ve birim üretimlerini kayıt dosyasına yazılan tur bazlı kuyruk olarak saklar. `kind` alanı `building` veya `unit`, `type_id` ise bina ID'si veya birim tipi ID'sidir. `turns_left` her tur çözümlemede azalır; sıfırlandığında üretim uygulanır.

---

## Runtime-Only Alanlar (`json:"-"`)

Bu alanlar JSON'a yazılmaz; oyun her başladığında assets'ten yeniden yüklenir:

| Alan | Yükleme kaynağı |
|---|---|
| `UnitTypes` | `assets/scenarios/<id>/data/units.json` |
| `BuildingTypes` | `assets/scenarios/<id>/data/buildings.json` |
| `TechTypes` | `assets/scenarios/<id>/data/technologies.json` |
| `ShapeData` | `assets/scenarios/<id>/data/country_shapes.json` |
| `AvailableVictories` | `assets/scenarios/<id>/scenario.json` |

**Neden bu ayrım?** Tanım verisi değişmez — onu kayıt dosyasına koymak gereksiz ve kırılgan. Sadece *durum* (kim neye sahip, ne araştırdı) kaydedilir.

`MapConfig` senaryo metadata'sından gelir ve kayıt dosyasına da yazılır. Böylece senaryo değiştiğinde aktif kaydın harita hizalama ayarı korunur.

---

## Yardımcı Metodlar

`CurrentSeason() Season` — `season.FromMonth(s.Month)` ile mevsimi döner → [[systems/seasons]]

`AdvanceTurn()` — `Turn++`, `Month++`, Ocak geçince `Year++`

`RegionsOwnedBy(fid) []*Region` — fraksiyon bölge listesi

`IsEliminated(fid) bool` — bölgesi yoksa `true`

`ManpowerCap(fid) int` — kara bölgesi başı 5 + kışlalı bölge başı +5 ek kapasite

`DeployedLandUnits(fid) int` — fraksiyonun aktif kara birim sayısı

`MaxLandArmies(fid) int` — `ceil(kara_bölge_sayısı / 2)` (minimum 1)

`CurrentLandArmies(fid) int` — fraksiyonun aktif kara ordu sayısı

---

## Veri Yükleme Akışı

`loadScenario()` — `internal/game/game.go`

Tüm yollar `gs.ScenarioPath` üzerinden senaryo klasörüne yönelir:

```
scenario.LoadAll("assets/scenarios")  → senaryo listesi
    ↓ senaryo seçilince
world.LoadRegions(scenario.DataPath("regions.json"))
world.LoadCountryShapes(scenario.DataPath("country_shapes.json"))
faction.LoadFactions(scenario.DataPath("factions.json"))
army.LoadUnitTypes(scenario.DataPath("units.json"))
city.LoadBuildings(scenario.DataPath("buildings.json"))
tech.LoadTechnologies(scenario.DataPath("technologies.json"))
faction.BuildInitialRelations()  → ilişki map'i (din bonusları dahil)
army.LoadArmies(scenario.DataPath("armies.json")) → başlangıç orduları
```

Kayıttan yüklemede `internal/save/save.go:loadFromPath` kayıt JSON'unu okur ve runtime tanım verilerinden `UnitTypes`, `BuildingTypes`, `TechTypes` alanlarını yeniden doldurur. Senaryo metadata'sı da tekrar okunur; eski kayıtta `MapConfig` yoksa `scenario.json` içindeki `map` alanı uygulanır ve `AvailableVictories` güncellenir. `ShapeData` şu an kayıttan yüklemede state'e geri yazılmıyor; takip işi [[dev/progress]] altında listeli.

---

## Zafer Koşulu Yapısı

```go
type VictoryCondition struct {
    Type               VictoryType      // domination | economic | military | religious
    TargetRegionCount  int              // domination: 20+ bölge
    RequiredRegions    []RegionID       // domination: constantinople, rome, paris, cairo, jerusalem
    TargetGoldIncome   int              // economic: mevcut kodda hazine eşiği gibi kontrol ediliyor
    GoldHoldTurns      int              // economic: kaç tur koru
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
PhaseScenarioSelect // senaryo seçim ekranı
PhaseFactionSelect  // fraksiyon seçim
PhaseVictorySelect  // zafer koşulu seçim
PhasePlayerTurn     // oyuncu aksiyonları
PhaseAITurn         // AI tur işlemi
PhaseTurnResolution // tur çözümleme
PhaseGameOver       // oyun sonu
PhasePauseMenu      // oyun içi duraklama menüsü (ESC)
PhaseLoadSelect     // kayıt slot seçim ekranı (yükleme)
PhaseSaveSelect     // kayıt slot seçim ekranı (kaydetme)
```

**ESC akışı (PhasePlayerTurn'de):**
- Seçili bölge/ordu/panel varsa → iptal et (faz değişmez)
- Hiçbir şey seçili değilse → `PhasePauseMenu`

→ Geçiş diyagramı için [[architecture/game-loop]]
