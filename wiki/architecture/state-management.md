---
type: architecture
tags: [state, gamestate, serialize, save-load]
last_updated: 2026-07-12
related: [game-loop, render-pipeline, shape-editor]
---

# State Yönetimi

**Kaynak:** `internal/state/state.go`

## GameState Yapısı

`GameState` tüm oyun verisinin tek kaynağıdır. Ancak save/load artık bu struct'ın ham snapshot'ını yazmaz; senaryo tanımını yeniden kurup yalnız kampanya sırasında değişen delta state'i serialize eder.

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
    ScenarioVictories  []VictoryOptionDef  // scenario.json'daki tam zafer listesi
    AvailableVictories []VictoryOptionDef  // oyuncu fraksiyonuna filtrelenmiş görünür liste
    RegionLogistics    map[RegionID]RegionLogisticsStatus
    ArmyLogistics      map[ArmyID]ArmyLogisticsStatus

    // Zafer takibi
    EconomicVictoryTurns  int
    FactionsEliminated    int
    ReligiousVictoryTurns int

    // Diplomatik & ticaret
    Relations     map[string]*Relation
    TradeRoutes   []*TradeRoute
    Sieges        map[RegionID]*SiegeState
    FiredEventIDs map[string]bool

    ProductionQueue []ProductionOrder // devam eden bina/birim üretimleri
    NextProductionSeq int             // üretim ID sayacı
    NextArmySeq int                   // ordu ID üretici sayaç

    Phase    Phase
    WinnerID FactionID
}
```

`ProductionOrder`, bina ve birim üretimlerini kayıt dosyasına yazılan tur bazlı kuyruk olarak saklar. `kind` alanı `building` veya `unit`, `type_id` ise bina ID'si veya birim tipi ID'sidir. `turns_left` her tur çözümlemede azalır; ancak bölge aktif kuşatma altındaysa bina ve birim emirleri duraklatılır, kuşatma kalkınca aynı sayaçtan devam eder; bölge el değiştirirse o bölgedeki üretim emirleri kuyruktan silinir; sıfırlandığında üretim uygulanır.

`SiegeState`, tahkimli düşman kara bölgesi üstündeki aktif kuşatmayı serialize eder. Kayıt; hedef bölgeyi, kuşatan orduyu, varsa içerideki savunucu orduyu, başlangıç turunu, geçen süreyi, o anki tahkimat seviyesini ve gedik ilerlemesini taşır. Böylece save/load sonrası kuşatma baskısı kaybolmaz.
`CanJoinActiveSiege(attacker, regionID)`, aynı fraksiyon, müttefik veya aynı vassal zincirindeki bir ordunun mevcut kuşatmaya normal hareketle destek verip veremeyeceğini döner; bu kural render ve game katmanında aynı relation/hiyerarşi verisinden okunur.
`SelectBattleDefender(attacker, target, navalSeaMove)` artık kara ve deniz için savunucu seçimini yalnız gerçekten savaş halindeki ordularla sınırlar; müttefik veya barış durumundaki ordular hedef bölgede dursa bile savaş planı/presolve akışına girmez.

`Army` state'i içinde artık `IsGarrison` alanı bulunur. Senaryo/save dosyalarındaki eski `army_garrison_*` veya `*_garrison` ID'leri load sırasında normalize edilerek bu bayrağa taşınır; böylece saha ordusu limiti ile sabit garnizon başlangıç birlikleri birbirine karışmaz.

Fraksiyon state'i artık ulusal başkent settlement'ını ve olası taşıma kuyruğunu da serialize eder:

- `CapitalSettlementID`
- `PendingCapitalSettlementID`
- `PendingCapitalTurns`

Fraksiyon state'i ayrıca vassallık zincirini de serialize eder:

- `OverlordID`

Bu alan `Relation.Stance` içine gömülmez; çünkü bir devlet aynı anda yalnız bir overlord'a bağlı olabilir ama diğer relation kayıtları ayrı kalır.

---

## Runtime-Only Alanlar (`json:"-"`)

Bu alanlar JSON'a yazılmaz; oyun her başladığında assets'ten yeniden yüklenir:

| Alan | Yükleme kaynağı |
|---|---|
| `UnitTypes` | `assets/scenarios/<id>/data/units.json` |
| `BuildingTypes` | `assets/scenarios/<id>/data/buildings.json` |
| `TechTypes` | `assets/scenarios/<id>/data/technologies.json` |
| `ShapeData` | `assets/scenarios/<id>/data/country_shapes.json` |
| `ScenarioVictories` | `assets/scenarios/<id>/scenario.json` içindeki tam zafer listesi |
| `AvailableVictories` | `ScenarioVictories` listesinin oyuncu fraksiyonuna göre filtrelenmiş kopyası |
| `RegionLogistics`, `ArmyLogistics` | tur çözümlemesinde üretilen geçici ikmal baskısı UI özeti |

**Neden bu ayrım?** Tanım verisi değişmez — onu kayıt dosyasına koymak gereksiz ve kırılgan. Sadece *durum* (kim neye sahip, ne araştırdı) kaydedilir.

`MapConfig`, `TradeCenters`, bölge adları/komşulukları/shape kimlikleri, fraksiyon adları/renkleri ve `region_shapes.json` kaynaklı region paint override verisi senaryodan gelir ve kayıt dosyasına yazılmaz. Kayıt yalnız sahiplik, ekonomi, araştırma, ordular, diplomasi ve benzeri mutable campaign state'i taşır.

Kompakt save formatı ayrıca şu sıkıştırmaları kullanır:

- `relations`: tam snapshot yerine yalnız baz senaryodan farklı relation delta'ları yazılır
- `regions`: yalnız değişen mutable alanlar serialize edilir; değişmeyen bölgeler tamamen atlanır
- `settlements`: tam liste yerine add/update/remove + order patch formatı kullanılır
- `armies.units`: aynı `type_id + hp + xp` kombinasyonları count ile stack'lenir
- disk katmanı: payload `zstd+base64` ile `state_zstd` alanında tutulur; envelope metadata'sı düz JSON kalır
- debug/dev katmanı: yalnız `GameState.DevelopmentMode=true` iken aynı slot için `saves/<slot>.debug.json` sidecar'ı yazılır; bu dosya sıkıştırılmamış ve açıklayıcı alan adlarıyla debug amaçlıdır, normal oyunda üretilmez

---

## Yardımcı Metodlar

`CurrentSeason() Season` — `season.FromMonth(s.Month)` ile mevsimi döner → [[systems/seasons]]

`AdvanceTurn()` — `Turn++`, `Month++`, Ocak geçince `Year++`

`SyncTimedRegionUnlocks() []RegionID` — `is_locked=true` ve `unlock_turn>0` olan bölgelerde aktif tur `unlock_turn` değerine ulaştıysa kilidi kaldırır; save/load ve tur ilerlemesinde senkron için kullanılır

`RegionsOwnedBy(fid) []*Region` — fraksiyon bölge listesi

`LandRegionsOwnedBy(fid) []*Region` — fraksiyonun yalnızca kara bölgeleri

`SelectBattleDefender(attacker, target, navalSeaMove)` — hedef bölgede saldıranı karşılayacak düşman orduyu deterministik seçer; kara savaşında en güçlü savunucuyu, deniz savaşında ise yalnız `StanceWar` ilişkisine sahip filoları dikkate alır. Savaş preview modalı ile gerçek resolve aynı savunucuyu kullansın diye render ve game katmanı bu helper üzerinden bağlanır.

`SiegeAt(regionID)` / `SiegeByArmy(armyID)` — aktif kuşatma kaydını bölge veya saldıran ordu üstünden döner; renderer, AI ve oyun mantığı aynı save verisini bu helper'larla okur.

`RegionProductionSummary(region) RegionProductionSummary` — seçili bölgenin efektif altın/mal üretimini hesaplar; bina çarpanları, arazi uzmanlaşması, mevsim ticaret/hasat etkileri ve sahip fraksiyonun ekonomi teknolojilerini UI önizlemesiyle paylaşır

`FindSettlementByID(settlementID)` — settlement ID'den region + settlement çözümlemesi yapar

`FactionCapital(fid)` — fraksiyonun geçerli başkent settlement ve bölgesini döner

`SetFactionCapital(fid, settlementID)` — başkenti anında değiştirir ve pending kuyruğu temizler

`StartCapitalMove(fid, settlementID, turns)` / `AdvanceCapitalMoves()` — 5 tur gibi gecikmeli başkent taşıma akışını yürütür

`NormalizeFactionCapitals()` — yükleme sonrası eksik/geçersiz başkentleri en yüksek getirili owned settlement'a normalize eder

`RegionLogisticsStatus` / `ArmyLogisticsStatus` — son turdaki bölgesel ikmal yükü, kapasite, aşım ve zayiat bilgisini render katmanına taşır; serialize edilmez.

`IsEliminated(fid) bool` — kara toprağı yoksa `true` (sadece deniz bölgesi kalan fraksiyonlar da elenir)

`ManpowerCap(fid) int` — kara bölgesi başı 5 + kışla seviyesi başına +5 ek kapasite

`DeployedLandUnits(fid) int` — fraksiyonun aktif kara birim sayısı

`MaxLandArmies(fid) int` — `ceil(kara_bölge_sayısı / 2)` (minimum 1)

`CurrentLandArmies(fid) int` — fraksiyonun aktif saha ordusu sayısı; `IsGarrison=true` kara orduları bu limite dahil edilmez

`LandUnitProductionLimit(region)` / `NavalUnitProductionLimit(region)` / `UnitProductionLimit(region, unitType)` — aynı tur tamamlanabilecek üretim adedinin ortak kaynak noktası; kara hattı `max(1,kışla seviyesi)`, deniz hattı `liman seviyesi`

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

Kayıttan yüklemede `internal/save/save.go:loadFromPath` önce kayıt JSON'unu okur, gerekiyorsa `state_zstd` payload'unu açar, `ScenarioID` üzerinden senaryo klasörünü çözer ve senaryo baz state'ini tekrar kurar. Ardından kayıt içindeki campaign delta bu baz state'in üstüne overlay edilir. Bu yüzden `UnitTypes`, `BuildingTypes`, `TechTypes`, `ShapeData`, `RegionOrder`, `FactionOrder`, `MapConfig`, `TradeCenters` ve tam `ScenarioVictories` listesi dosyadan değil yeniden senaryodan gelir. Yeni save formatı üst seviyede `kind` (`auto`, `quick`, `slot`), `game_version` ve slot kartları için düz `meta` alanı taşır; sıkıştırılmış gövde `state_zstd` içinde tutulur, eski düz `GameState` save'leri ve eski wrapper save'ler de geriye uyumlu okunur. Legacy save içinde bazı alanlar yoksa baz senaryonun varsayılanları korunur. `DevelopmentMode` açıksa save sırasında ana dosyaya ek olarak `*.debug.json` sidecar'ı da yazılır; bu yardımcı dosya yükleme için zorunlu değildir ve normal mod save alındığında aynı slotun eski debug sidecar'ı temizlenir. `Game.startLoadSlot()` save yüklendiğinde olay listesini (`events.json`) tekrar kurar; böylece ses/müzik, zafer seçimi ve olay akışı yeni oturumda da aktif senaryoyla tutarlı kalır. `ShapeData`, `country_shapes.json` içindeki ring + isim bilgisini tutar; edit mode shape paint işlemleri bu runtime veriyi günceller ve senaryo kaydında tekrar dosyaya yazar.

Load/startup sonunda `diplomacy.NormalizeVassalage()` çalışır. Böylece geçersiz `OverlordID` referansları temizlenir, realm içi relation kayıtları dost çizgiye çekilir ve vassalın üçüncü taraf trade/offer sızıntıları kapanır.

`Game` katmanında ayrıca serialize edilmeyen bir `pendingConquestDecisions` kuyruğu vardır. Bu runtime kuyruk, oyuncu savaşta bir devletin son kara toprağını düşürdüğünde battle report ile nihai ilhak/vassallık kararını birbirinden ayırmak için kullanılır; save/load veya yeni oyun başlangıcında temizlenir.

Kayıt slotları: `autosave`, `quicksave`, `slot1`, `slot2`, `slot3`. Oyun içinde `ActionSave` (Ctrl+S/S) `quicksave` slotuna yazar; `ActionEndTurn` ve araştırma onayından gelen `ActionConfirmEndTurn` AI turuna geçmeden hemen önce `autosave` slotuna yazar. Ana menüde `Devam Et`, `autosave` ile `quicksave` arasından en yeni olanı açar.

---

## Zafer Koşulu Yapısı

```go
type VictoryCondition struct {
    Type               VictoryType      // domination | economic | military | religious | conquer_city | survive_turns
    TargetRegionCount  int              // domination: 20+ bölge
    RequiredRegions    []RegionID       // domination: constantinople, rome, paris, cairo, jerusalem
    TargetGoldIncome   int              // economic: tur başı gelir eşiği
    GoldHoldTurns      int              // economic: kaç tur koru
    TargetArmyStrength int             // military: 200 güç puanı
    TargetDefeated     int             // military: 3 fraksiyon yenilgisi
    TargetTurns        int             // survive_turns: toplam tur eşiği
    DeadlineYear       int             // 0 = süresiz
    DeadlineMonth      int             // 1-12, 0 = yıl sonu
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
