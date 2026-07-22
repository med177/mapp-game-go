---
type: architecture
tags: [state, gamestate, serialize, save-load]
last_updated: 2026-07-22
related: [game-loop, systems/events, systems/economy, render-pipeline, shape-editor]
---

# State Yönetimi

**Kaynak:** `internal/state/state.go`, `internal/state/war_ledger.go`

## GameState Yapısı

`GameState` tüm oyun verisinin tek kaynağıdır. Ancak save/load artık bu struct'ın ham snapshot'ını yazmaz; senaryo tanımını yeniden kurup yalnız kampanya sırasında değişen delta state'i serialize eder.

Kaynak HUD'u için `FactionProductionSummary()` kuşatma dışı bölgelerin efektif üretimini toplar. `FactionGrainNetChange()` bu toplamdan sivil tahıl talebi ile orduların efektif tahıl bakımını çıkarır; böylece HUD'daki negatif net tahıl göstergesi ekonomi kurallarıyla aynı state hesaplarına dayanır.

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
    Commanders map[string]*Commander
    AIPlans map[FactionID]*AIPlanState
    WarLedgers map[string]*WarLedger
    ShapeData CountryShapeJSON           // json:"-"

    // Runtime-only (json:"-")
    AIStrategies       map[string]AIFactionStrategy
    AIDifficultyPolicy AIDifficultyPolicy
    UnitTypes          map[string]*UnitType
    BuildingTypes      map[string]*Building
    TechTypes          map[string]*Technology
    ScenarioVictories  []VictoryOptionDef  // scenario.json'daki tam zafer listesi
    AvailableVictories []VictoryOptionDef  // oyuncu fraksiyonuna filtrelenmiş görünür liste
    RegionLogistics    map[RegionID]RegionLogisticsStatus
    ArmyLogistics      map[ArmyID]ArmyLogisticsStatus
    GrainEconomy       map[FactionID]GrainEconomyStatus
    AutoGrainExport    bool // oyuncunun save'lenen otomatik ihracat tercihi
    ArmyMoveUsage      map[ArmyID]bool // ekonomi tick'i öncesi runtime hareket snapshot'ı
    GrainAidUsage      map[RegionID]bool // tur içi tahıl yardımı kilidi

    // Zafer takibi
    EconomicVictoryTurns  int
    FactionsEliminated    int
    ReligiousVictoryTurns int

    // Diplomatik & ticaret
    Relations     map[string]*Relation
    OfferRejectionTurns map[string]int // reddedilen tekliflerin retry cooldown state'i
    DiplomaticOffers []DiplomaticOffer // normal ve bölge bağlı kuşatma teklifleri
    TradeRoutes   []*TradeRoute
    Sieges        map[RegionID]*SiegeState
    FiredEventIDs map[string]bool

    ProductionQueue []ProductionOrder // devam eden bina/birim üretimleri
    NextProductionSeq int             // üretim ID sayacı
    NextArmySeq int                   // ordu ID üretici sayaç
    NextCommanderSeq int              // komutan ID üretici sayaç

    Phase    Phase
    WinnerID FactionID
}
```

`ProductionOrder`, bina ve birim üretimlerini kayıt dosyasına yazılan tur bazlı kuyruk olarak saklar. `kind` alanı `building` veya `unit`, `type_id` ise bina ID'si veya birim tipi ID'sidir. `turns_left` her tur çözümlemede azalır; ancak bölge aktif kuşatma altındaysa bina ve birim emirleri duraklatılır, kuşatma kalkınca aynı sayaçtan devam eder; bölge el değiştirirse o bölgedeki üretim emirleri kuyruktan silinir; sıfırlandığında üretim uygulanır.

`GameState.CollectDefenders()` birleşik savunmaya katılan gerçek orduları `ArmyID` sırasıyla toplar. Böylece 20 birim sınırına giren kompozisyon, kaynak ordu ID listesi ve sonrasındaki kayıp dağıtımı map iterasyon sırasından bağımsızdır; aynı state ve aynı savaş zarı aynı sonucu üretir.

`SiegeState`, tahkimli düşman kara bölgesi üstündeki aktif kuşatmayı serialize eder. Kayıt; hedef bölgeyi, kuşatan orduyu, varsa içerideki savunucu orduyu, başlangıç turunu, geçen süreyi, o anki tahkimat seviyesini ve gedik ilerlemesini taşır. Böylece save/load sonrası kuşatma baskısı kaybolmaz.

`WarLedger`, `RelationKey` ile aynı sıralı taraf anahtarında yalnız aktif savaşın kalıcı
sonuç state'ini tutar: başlangıç turu, iki tarafın başlangıç kara bölgesi sayısı, tamamen
kaybedilen birlikler, ele geçirilen bölge sayıları, son muharebe turu ve son barış teklifi
turu. `BeginWarLedger()` savaş geçişinde snapshot alır; muharebe/fetih executor'ları
sayaçları günceller, `EndWarLedger()` barışta kaydı kaldırıp hedef planın rally state'ini
temizler. `SyncWarLedgers()` eski save veya doğrudan stance düzenleyen legacy yolları
aktif ilişkilerle uzlaştırır; ledger taşımayan eski save'deki savaş yükleme turunda sıfır
sayaçla başlar.
`CanJoinActiveSiege(attacker, regionID)`, aynı fraksiyon, müttefik veya aynı vassal zincirindeki bir ordunun mevcut kuşatmaya normal hareketle destek verip veremeyeceğini döner; bu kural render ve game katmanında aynı relation/hiyerarşi verisinden okunur.
`IsArmyDefendingSiegedRegion(candidate)`, aktif kuşatma altındaki bölgede bölge sahibi veya onun müttefiki olarak duran kara ordusunu ortak savunmacı state'i olarak tanımlar. Bu predicate huruç zorunluluğu ile kuşatma altı iyileşme engelinin aynı state kuralını kullanmasını sağlar.
`SelectBattleDefender(attacker, target, navalSeaMove)` artık kara ve deniz için savunucu seçimini yalnız gerçekten savaş halindeki ordularla sınırlar; müttefik veya barış durumundaki ordular hedef bölgede dursa bile savaş planı/presolve akışına girmez.

`Army` state'i içinde artık `IsGarrison` alanı bulunur. Senaryo/save dosyalarındaki eski `army_garrison_*` veya `*_garrison` ID'leri load sırasında normalize edilerek bu bayrağa taşınır; böylece saha ordusu limiti ile sabit garnizon başlangıç birlikleri birbirine karışmaz.

`Army.Commander` alanı komutanın kalıcı kariyer state'ini taşır. Komutan ID'si, adı,
seviyesi, XP'si, savaş/zafer sayıları ve trait listesi compact save içindeki `ar.*.c`
alanına yazılır. Komutan pointer'ı kopyalanırken trait slice'ı da bağımsız kopyalandığı
için edit-mode undo/redo ve save/load aynı komutan state'ini paylaşan yanlış referanslar
üretmez. `GameState.Commanders` aynı komutanın iki orduya atanmasını engelleyen
canonical havuzdur; `SyncCommanderLinks()` yükleme sonrasında ordu pointer'larını bu
havuzdaki nesnelere bağlar. Oyuncu havuzu ve ordu panelindeki atama/ayırma modalı
`InitializePlayerCommanders()`, `AssignCommanderToArmy()` ve
`UnassignCommanderFromArmy()` üzerinden çalışır.
Senaryo başlangıç şablonları `data/commanders.json` dosyasından okunur. Her kayıt
`id`, `owner_id`, `name`, başlangıç `level`/`experience`/`traits` ve ileride portre
yüklemek için `portrait_asset` alanlarını taşıyabilir. Şablon runtime komutanına
clone edilir; savaşlarda değişen XP, seviye ve atamalar save state içindeki
`GameState.Commanders` havuzunda tutulur.
AI tarafında `EnsureFactionCommanders(ownerID)` aktif saha ordusu sayısına göre
havuzu büyütür ve komutansız orduları deterministik ID sırasıyla doldurur; garnizonlar
bu otomatik atamanın dışındadır.

Ordu yaşam döngüsü de komutan bağlantısını state katmanında korur. `RemoveArmy()`
silinen ordunun normal veya nakliye filosunda taşınan komutanını havuza geri bırakır;
`TransferArmyOwnership()` fetih ya da fraksiyon eliminasyonu sonrası hem ordu hem de
komutan `OwnerID` alanlarını birlikte günceller. Kara ordusu filoya bindiğinde komutan
`Army.EmbarkedCommander` alanında korunur; `AmphibiousCommander()` çıkarma savaşında
bu komutanı kullanır, başarılı karaya çıkışta yeni orduya geri bağlar ve başarısız
çıkarma veya iptal durumunda havuza serbest bırakır.

Ordu hareket havuzu da runtime state'te birim kompozisyonundan türetilir.
`Army.BaseMovePoints(UnitTypes)` kendi `Units` listesindeki en düşük
`UnitType.MovementPoints` değerini seçer; bu nedenle yalnız süvari `3`, yalnız
piyade `2`, yalnız kuşatma/topçu `1`, karışık kara ordusu ise en yavaş birim kadar
ilerler. Filo hesabında `EmbarkedUnits` dikkate alınmaz.
`GameState.ArmyMaxMovePoints()` bu tabana mevcut mevsim çarpanını uygular ve
ardından komutan/teknoloji bonuslarını ekler. 1300 senaryosunun `fair_movement`
politikası oyuncu ve AI hesabını eşitler; config taşımayan eski senaryolarda Zor AI'nin
legacy `+1` hareketi korunur. `RefreshArmyMovePoints()` ilk senaryo ve save/load
senkronizasyonunda kullanılır.

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

`RegionLogisticsStatus` / `ArmyLogisticsStatus` — son turdaki bölgesel ikmal yükü, kapasite, abluka yüzdesi, aşım ve zayiat bilgisini render katmanına taşır; serialize edilmez.

`GrainEconomyStatus` / `GameState.GrainEconomy` — son ekonomi tick'inde fraksiyon bazlı tahıl üretimi, sivil tüketimi, ordu bakımı, ordu yenilemesi, ordu `ArmyMoraleDelta` değişimi, stratejik ithalat ihtiyacı, nüfus büyümesi ve otomatik ihracat için harcanan tahıl, net değişim, stok-ay seviyesi ve açık bilgisini render/event bildirimlerine taşır; runtime-only olduğu için save'e yazılmaz.

`GameState.ArmyMoveUsage` — `applySeasonEffects()` hareket puanlarını yenilemeden önce ordunun o tur hareket edip etmediğini geçici olarak yakalar. `GameState.EffectiveArmyGrainUpkeep()` bu snapshot'ı kuşatma ve garnizon katsayılarıyla birleştirir; ekonomi, bölgesel lojistik ve AI aynı efektif talebi kullanır. Alan serialize edilmez.

`GameState.GrainAidUsage` / `CanApplyGrainAid()` / `ApplyGrainAid()` — oyuncunun bölge panelinden yaptığı tahıl yardımını bölge başına turda bir kez sınırlar; 12 tahıl karşılığında memnuniyeti +10 artırır. Kullanım haritası `AdvanceTurn()` içinde sıfırlanır ve save'e yazılmaz.

`EmergencyGrainSaleLimit()` / `EmergencyGrainSaleUnitPrice()` / `ApplyEmergencyGrainSale()` — pazar partneri gerektirmeyen acil tahıl satışını yönetir. Yalnızca fraksiyon depo kapasitesi üzerindeki miktar satılır; `economy.EmergencySaleUnitPrice()` güncel fiyatın %70 indirimli değerini üretir. Bu işlem kalıcı yeni state alanı eklemez.

`GameState.AutoGrainExport` / `ApplyAutomaticGrainExport()` — Pazar sekmesindeki tercihi ve ekonomi tick'inde kapasite üzeri tahılın aktif, savaşta olmayan ticaret ağı partnerlerine faction ID sırasıyla %60 fiyatla satışını yönetir. Alıcı altını yetersizse miktar alıcının bütçesiyle sınırlanır; tercih compact save alanında korunur, gerçekleşen miktar ve altın runtime `GrainEconomyStatus` içinde raporlanır.

`applyGrainFundedArmyReplenishment()` — mevcut ücretsiz dost-toprak toparlanmasına ek olarak kapasite üstü tahılı dost ve kuşatma dışı kara ordularına aktarır. Faction/army ID sırası deterministiktir, ordu başına en fazla +10 HP verilir ve 1 HP başına 1 tahıl tüketilir; rezerv kapasitesi altına inilmez.

`GameState.StrategicGrainDemand()` / `StrategicGrainSurplus()` — fraksiyonun üç aylık güvenli rezerv hedefindeki açığı ve kapasite üstü ihraç edilebilir stoku hesaplar. Diplomasi yeni rota kurarken bu iki sinyalle hedefteki tahıl ihtiyacını kaynak fazlasına bağlar; sinyal save'e yazılmaz ve her ekonomi tick'inde yeniden türetilir.

`RegionEventStatus` içindeki `GrainProductionPercent` ve `GrainDemandPercent`, aktif hasat/kıtlık/kuraklık olaylarının geçici bölgesel tahıl etkileridir. `RegionGrainProductionModifier()`, `RegionGrainDemandModifier()` ve `CivilianGrainDemandForRegion()` bu kayıtları toplar; alanlar `ActiveRegionEvents` ile compact save/load içinde korunur, süre dolunca `TickActiveRegionEvents()` tarafından temizlenir.

`GameState.RegionMilitaryGrainProduction()` bölgesel efektif tahıl üretiminden aktif sivil talebi düşer. Oyun lojistiği ve AI hareket/recruitment lojistiği bu ortak helper'ı; ordu talebi için de `EffectiveArmyGrainUpkeep()` metodunu kullanır. Böylece oyuncu ve AI aynı tahıl tüketim kurallarından sapmaz.

`GrainStorageCapacity()` ve `GameState.GrainStorageCapacityForFaction()` sivil nüfus talebi, efektif ordu bakımı ve ambar bina bonusunu aynı `6 ay sivil + 3 ay ordu`, minimum 100 kapasite kuralında birleştirir. İkinci helper ekonomi tick'i oluşmadan HUD'un başlangıçta da doğru ambar kapasitesini gösterebilmesini sağlar.

`Army.Morale` ordunun kalıcı ikmal moralidir. `CurrentMorale()` eski kayıt veya fixture'larda eksik alanı 100 başlangıç morali olarak normalize eder; `ApplyMoraleDelta()` değeri 1–100 aralığında tutar. Compact save/load içindeki `mo` alanıyla taşınır ve `Army.TotalStrength()` içinde savaş/AI güç değerlendirmelerine uygulanır.

`TradeRoute.BlockadePercent` — rota uçlarındaki denizlerde bulunan düşman savaş gemilerinden türetilen geçici hacim kesintisidir. `RefreshTradeRouteBlockades()` ve `RegionBlockadePercent()` konum/savaş state'inden her ekonomi tick'inde yeniden hesaplar; save migration gerektirmez.

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

Kayıttan yüklemede `internal/save/save.go:loadFromPath` önce kayıt JSON'unu okur, gerekiyorsa `state_zstd` payload'unu açar, `ScenarioID` üzerinden senaryo klasörünü çözer ve senaryo baz state'ini tekrar kurar. Ardından kayıt içindeki campaign delta bu baz state'in üstüne overlay edilir. Bu yüzden `UnitTypes`, `BuildingTypes`, `TechTypes`, `ShapeData`, `RegionOrder`, `FactionOrder`, `MapConfig`, `TradeCenters` ve tam `ScenarioVictories` listesi dosyadan değil yeniden senaryodan gelir. `ArmyOrder`, dinamik ordu map'i ilk ihtiyaçta ID sırasına göre oluşturulan runtime-only cache'tir; save formatına yazılmaz ve ordu ekleme/kaldırma sonrası uzunluğu değiştiğinde yeniden kurulur. Yeni save formatı üst seviyede `kind` (`auto`, `quick`, `slot`), `game_version` ve slot kartları için düz `meta` alanı taşır; sıkıştırılmış gövde `state_zstd` içinde tutulur, eski düz `GameState` save'leri ve eski wrapper save'ler de geriye uyumlu okunur. `AIPlans` içindeki objective, target, commitment, reassess, rally bölgesi ve rally deadline alanları mutable campaign delta olarak saklanır; `StrategicContext` güç/lojistik/yol cache'leri, düşman eksenli `AIFront` kayıtları, dinamik rezerv hedefi, rally hazırlık gücü ve ordu rol atamaları runtime-only olduğu için yüklemede ve her AI turunda yeniden hesaplanır. Legacy save içinde AI plan alanı yoksa boş state korunur ve ilgili devletin sonraki AI turunda plan üretilir. `DevelopmentMode` açıksa save sırasında ana dosyaya ek olarak `*.debug.json` sidecar'ı da yazılır; bu yardımcı dosya yükleme için zorunlu değildir ve normal mod save alındığında aynı slotun eski debug sidecar'ı temizlenir. `Game.startLoadSlot()` save yüklendiğinde olay listesini (`events.json`) tekrar kurar; böylece ses/müzik, zafer seçimi ve olay akışı yeni oturumda da aktif senaryoyla tutarlı kalır. `ShapeData`, `country_shapes.json` içindeki ring + isim bilgisini tutar; edit mode shape paint işlemleri bu runtime veriyi günceller ve senaryo kaydında tekrar dosyaya yazar.

Load/startup sonunda `diplomacy.NormalizeVassalage()` çalışır. Böylece geçersiz `OverlordID` referansları temizlenir, realm içi relation kayıtları dost çizgiye çekilir ve vassalın üçüncü taraf trade/offer sızıntıları kapanır.

`AIPlans` mutable campaign niyetidir; objective kimliği/türü, hedef devlet ve bölge
öncelikleri, commitment, yeniden değerlendirme turu, rally bölgesi/deadline ile
vassallık/stratejik ilhak tercihlerini save/load arasında korur. Buna karşılık `AIStrategies` ve
`AIDifficultyPolicy`, `ai_strategies.json` dosyasından gelen statik senaryo verileridir
ve runtime-only tutulur. Save yüklemesinde baz senaryoyla yeniden kurulurlar; böylece
statik profil ile plan/risk/hareket politikası save payload'ında tekrar edilmez ve eski
save'ler sonraki AI turunda güncel senaryo konfigürasyonunu kullanabilir.

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
