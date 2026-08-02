---
type: architecture
tags: [state, gamestate, serialize, save-load]
last_updated: 2026-08-02
related: [game-loop, systems/events, systems/economy, systems/diplomacy, render-pipeline, shape-editor]
---

# State Yönetimi

**Kaynak:** `internal/state/state.go`, `internal/state/war_ledger.go`

## GameState Yapısı

`GameState` tüm oyun verisinin tek kaynağıdır. Ancak save/load artık bu struct'ın ham snapshot'ını yazmaz; senaryo tanımını yeniden kurup yalnız kampanya sırasında değişen delta state'i serialize eder.

`world.Region.SuccessorFactionID` senaryo metadata'sıdır ve compact save'de `sf`
delta alanıyla korunur. Böylece edit mode ataması, fetih sonrası özgürleştirme ve
save/load aynı bölge state'ini kullanır. Özgürleştirme runtime'da ardıl fraksiyonun
`IsEliminated` bayrağını kaldırır, kaynak/ordu/ittifak başlangıcını kurar ve başkenti
`NormalizeFactionCapitals()` ile yeniden belirler.

Geliştirme modunda save yükleme sonrası ilk beş AI fazı için
`AIDiagnosticHistory` ve `AIDiagnosticCaptureTurnsRemain` geçici runtime alanları
kullanılır. Bu alanlar normal compact campaign payload'ına girmez; beşinci fazdan
sonra yalnız debug sidecar'a `state.ai_diagnostic_history` olarak yazılır. Böylece
oynanabilir save formatı büyümeden AI plan/hedef/cephe/rezerv değişimi
karşılaştırılabilir.

Kaynak HUD'u için `FactionProductionSummary()` kuşatma dışı bölgelerin efektif üretimini toplar. `FactionGrainNetChange()` bu toplamdan sivil tahıl talebi ile orduların efektif tahıl bakımını çıkarır; böylece HUD'daki negatif net tahıl göstergesi ekonomi kurallarıyla aynı state hesaplarına dayanır. Bölge nüfusu `Region.RecalculatePopulation()` ile kırsal nüfus ve yerleşim nüfuslarının toplamından türetilir; `CivilianGrainDemand()` bu toplamı kullanır.

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
    Regions      map[RegionID]*Region
    LandPassages []world.LandPassage
    Factions  map[FactionID]*Faction
    Armies    map[ArmyID]*Army
    Commanders map[string]*Commander
    AIPlans map[FactionID]*AIPlanState
    Imperial *ImperialState // HRE otoritesi, üyelik ve seçim state'i
    WarLedgers map[string]*WarLedger
    RecentTruces map[string]int // relation key -> ateşkes bitiş turu
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
    OfferRejectionTurns map[string]int // normal ve bölge bazlı ret retry cooldown state'i
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

`ImperialState`, HRE gibi üst siyasi kurumları bağımsız üye fraksiyonlardan ayırır.
`EmpireID` kurumu, `EmperorID` mevcut seçilmiş hükümdarı, `Authority` merkezî
meşruiyeti, `Members` ise sadakat/özerklik/askerî bağlılık ve elektör ağırlıklarını
tutar. Bu alan compact save (`im`) ve debug/legacy save akışlarında korunur;
`data/imperial.json` yalnız senaryo başlangıç state'ini sağlar.

`ImperialState.PendingDecision`, HRE oyuncusunun çözmesi gereken Diyet veya seçim
kararını (`ImperialDecisionKind`) taşır. Bu state tur çözümünden sonra paneli açmak,
oyuncu kararını zorunlu kılmak ve save/load sonrasında modalı geri kurmak için kullanılır.
AI kontrollü HRE'de pending state oluşturulmaz; otomatik siyasi çözüm korunur.

`ProductionOrder`, bina ve birim üretimlerini kayıt dosyasına yazılan tur bazlı kuyruk olarak saklar. `kind` alanı `building` veya `unit`, `type_id` ise bina ID'si veya birim tipi ID'sidir. `turns_left` her tur çözümlemede azalır; ancak bölge aktif kuşatma altındaysa bina ve birim emirleri duraklatılır, kuşatma kalkınca aynı sayaçtan devam eder; bölge el değiştirirse o bölgedeki üretim emirleri kuyruktan silinir; sıfırlandığında üretim uygulanır.

`GameState.CollectDefenders()` birleşik savunmaya katılan gerçek orduları `ArmyID` sırasıyla toplar. Böylece 20 birim sınırına giren kompozisyon, kaynak ordu ID listesi ve sonrasındaki kayıp dağıtımı map iterasyon sırasından bağımsızdır; aynı state ve aynı savaş zarı aynı sonucu üretir.

`SiegeState`, tahkimli düşman kara bölgesi üstündeki aktif kuşatmayı serialize eder. Kayıt; hedef bölgeyi, kuşatan orduyu, varsa içerideki savunucu orduyu, başlangıç turunu, geçen süreyi, o anki tahkimat seviyesini ve gedik ilerlemesini taşır. Denizden tahkimli kıyıya inerek başlatılan kuşatmalar `NavalLanding` ile işaretlenir. Barış bu kuşatmayı kapattığında `EvacuateNavalLandingSiegesAfterPeace()` hedefe en yakın toplam yeterli kapasitedeki dost nakliye filolarına birlikleri geri yükler; uygun filo yoksa orduyu en yakın kendi kara bölgesine taşır. Böylece save/load sonrası kuşatma baskısı kaybolmaz ve barışta düşman bölgesinde kara ordusu bırakılmaz.

`WarLedger`, `RelationKey` ile aynı sıralı taraf anahtarında yalnız aktif savaşın kalıcı
sonuç state'ini tutar: başlangıç turu, iki tarafın başlangıç kara bölgesi sayısı, tamamen
kaybedilen birlikler, ele geçirilen bölge sayıları, son muharebe turu ve son barış teklifi
turu. Aktif savaşın runtime AI hedef kilidi de `TargetRegionID`/`TargetLockedTurn` ile
kısa süreli olarak save'e yazılır; hedef geçersizleşirse veya kilit süresi dolarsa AI
stratejik skorla yeni hedef seçer. `BeginWarLedger()` savaş geçişinde snapshot alır; muharebe/fetih executor'ları
sayaçları günceller, `EndWarLedger()` barışta kaydı kaldırıp hedef planın rally state'ini
temizler. `SyncWarLedgers()` eski save veya doğrudan stance düzenleyen legacy yolları
aktif ilişkilerle uzlaştırır; ledger taşımayan eski save'deki savaş yükleme turunda sıfır
sayaçla başlar.
Barış çözümünde `RecordTruce()` aynı relation key için altı tur sonrasını
`RecentTruces` içine yazar; bu alan compact save'e alınır ve eski save'lerde boş kabul
edilir. `TruceRemaining()` süresi dolmuş kaydı etkisiz sayar.
`CanJoinActiveSiege(attacker, regionID)`, aynı fraksiyon, müttefik veya aynı vassal zincirindeki bir ordunun mevcut kuşatmaya normal hareketle destek verip veremeyeceğini döner; bu kural render ve game katmanında aynı relation/hiyerarşi verisinden okunur.
`IsArmyDefendingSiegedRegion(candidate)`, aktif kuşatma altındaki bölgede bölge sahibi veya onun müttefiki olarak duran kara ordusunu ortak savunmacı state'i olarak tanımlar. Bu predicate huruç zorunluluğu ile kuşatma altı iyileşme engelinin aynı state kuralını kullanmasını sağlar.
`SelectBattleDefender(attacker, target, navalSeaMove)` artık kara ve deniz için savunucu seçimini yalnız gerçekten savaş halindeki ordularla sınırlar; müttefik veya barış durumundaki ordular hedef bölgede dursa bile savaş planı/presolve akışına girmez.

Donanma konumu iki katmanlıdır: `Army.RegionID` deniz rotası için kullanılan deniz ankrajını korur; filo limandaysa gerçek eş-konum `DockedSettlementID` (eski veride `DockedRegionID` fallback'i) olan `Army.LocationID()` ile okunur. `Army.IsAtSea()` yalnız dock bağı olmayan filoları açık deniz savaşı, abluka ve AI deniz tehdidi hesaplarına dahil eder. Deniz hareketi dock bağını temizleyerek hedef deniz bölgesini tekrar kanonik konum yapar. Save yükleme bu state'i olduğu gibi korur; eski dock migrasyonu yalnız başlangıç senaryosundaki eksik dock verisi için çalışır.

Oyuncu filosunun kalıcı görevi `Army.NavalMission` içinde tutulur. `patrol` ve
`blockade` filonun o an bulunduğu açık deniz bölgesini, `escort` aynı açık deniz bölgesindeki aynı devlete ait nakliye filosunu, `transport`
ise taşınan kara ordusu ile kıyı kara bölgesini hedefler. Atama ve temizleme
`GameState.CanAssignNavalMission()`, `AssignNavalMission()` ve
`ClearNavalMission()` kapılarından geçer; compact save/load `nm` alanıyla bu
görevi korur. `internal/game/player_naval_mission.go` her yeni turda hareket
puanı yenilendikten sonra yalnız nakliye filosunu deterministik deniz BFS'iyle
hedef kıyıya yaklaştırır; nakliye hedef kıyıya ulaştığında mevcut
çıkarma/komşuluk çözümünü kullanır. Devriye ve abluka görevi mevcut denizde
kalır, görev ataması filoyu başka bir denize hareket ettirmez.
Bu görevlerden birini taşıyan filo manuel hareket, temas geri çekilmesi veya
liman bağlantısı nedeniyle gerçek konumunu değiştirirse
`ClearNavalMissionAfterRelocation()` görevi otomatik temizler; böylece görev
eski deniz bölgesinde ekonomik veya devriye etkisi üretmeye devam etmez.
Devriye, abluka ve escort yalnızca ilgili filolar aynı açık denizdeyken görev
panelinden atanır; panel
hedef haritası açmaz ve tıklanan görev doğrudan mevcut `RegionID`'yi hedefler.
Limana bağlı filo için bu iki görev seçeneği gösterilmez. Devriye açık denizde
atanabilir; `blockade` hedefi ise yalnızca savaş halindeki düşmanın kıyı kara bölgelerine
komşu denizler arasından seçilebilir; açık deniz hedefleri state kapısından
reddedilir.
Görev etkileri de konuma ve göreve göre ayrıdır: görevsiz filolar aynı denizde
savaş başlatmaz. Açık denizdeki `blockade` filoları rota/liman kesintisi üretir;
`patrol` filoları aynı denizdeki dost rota ve liman üzerindeki hedef denize
atanmış düşman `blockade` filosunu otomatik yakalar; `blockade` görevi tek
başına savaş başlatmaz. Açık deniz saldırısı yalnız savaş planı onayındaki
doğrudan saldırı emriyle başlar. `escort` görevi,
aynı denizde savunmaya katılan nakliye filosuna yüzde 15 deniz savunması verir;
çoklu escort bonusu yüzde 30 ile sınırlıdır. Devriye veya escort görevi verilen
oyuncu savaş gemisi hedef denizde abluka kesintisi sayılmaz.

Açık denizde düşman filo tespit edildiğinde geçici `GameState.PendingNavalContact`
kaydı oluşturulur. Kayıt iki filo, deniz bölgesi, temas nedeni ve iki tarafın
kararını taşır; save'e yazılmaz. Aynı denize hareket ve savaş açıldığı anda zaten
aynı denizde bulunan filolar için temas üretilebilir. Filo zaten düşmanla aynı
denizdeyken yeni `Devriye` veya `Abluka` görevi atanması da temas üretir; aynı
görevin tekrar atanması yeni temas oluşturmaz. Oyuncuya ortak üç seçenekli
modal açılır: `Çatış`, `Geri çekil`, `Pozisyonu koru`. Devriye ve görevsiz filo
varsayılan olarak `Çatış`, abluka filosu `Pozisyonu koru` seçer; savaş yalnız iki tarafın
kararı da `Çatış` olduğunda çözülür. Savaş açılışı temasında AI tarafı varsayılan
olarak `Çatış` seçer. `QueueNavalContactForWar()` ve
`NavalContactDecisionForPlayer()` bu geçici kararı merkezi state kapısından yönetir.
Oyuncu filosunun hareket puanı yoksa `Geri çekil` seçeneği modalda pasif kalır ve
state katmanı da bu kararı kabul etmez. Geçerli bir geri çekilme, 2 hareket puanı
harcar; filonun kalan puanı 2'den azsa sıfıra indirilir. Hedef komşu denizler
arasında düşman filosu olmayan bölge varsa geri çekilme rotası onu seçer ve
temasın geldiği kaynak denizi dışarıda bırakır. Güvenli hedef yoksa state kapısı
geri çekilme kararını reddeder.

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
komutan `OwnerID` alanlarını birlikte günceller. Kara ordusu filoya bindiğinde, hem
ayrı embark aksiyonunda hem de denize hareket ederek yapılan otomatik embark'ta kara
komutanı `Army.EmbarkedCommander` alanında korunur; `AmphibiousCommander()` yalnız bu
komutanı döndürür. Filo komutanı çıkarma savaşına katılmaz. Başarılı karaya çıkışta
kara komutanı yeni orduya geri bağlanır; başarısız çıkarma veya iptal durumunda havuza
serbest bırakılır.

Ordu hareket havuzu da runtime state'te birim kompozisyonundan türetilir.
`Army.BaseMovePoints(UnitTypes)` kendi `Units` listesindeki en düşük
`UnitType.MovementPoints` değerini seçer; bu nedenle yalnız süvari `3`, yalnız
piyade `2`, yalnız kuşatma/topçu `1`, karışık kara ordusu ise en yavaş birim kadar
ilerler. Filo hesabında `EmbarkedUnits` dikkate alınmaz.
`GameState.ArmyMaxMovePoints()` bu tabana mevcut mevsim çarpanını uygular ve
ardından komutan/teknoloji bonuslarını ekler. 1300 senaryosunun `fair_movement`
politikası oyuncu ve AI hesabını eşitler; config taşımayan eski senaryolarda Zor AI'nin
legacy `+1` hareketi korunur. `RefreshArmyMovePoints()` ilk senaryo ve save/load
senkronizasyonunda kullanılır. `RefreshArmyMovePointsAfterCompositionChange()` split
ve merge sonrası kompozisyonu yeniden hesaplar; ordu o turda hareket etmemişse yeni
hareket havuzunu tamamen kullanılabilir yapar, hareket etmişse kalan puanı iade etmez.

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

`LandPassages` senaryo verisinden yüklenen, iki kara bölgesi arasındaki özel
geçiş kayıtlarını taşır. Bu fazda render ve edit mode tarafından kullanılır;
oyuncu hareketi ile savaş çözümü henüz bu alanı tüketmez. Senaryo yüklemesinde
`data/land_passages.json`, edit mode kaydında aynı dosya kullanılır; normal
save/load ise alanı campaign state içinde korur.

`RegionsOwnedBy(fid) []*Region` — fraksiyon bölge listesi

`LandRegionsOwnedBy(fid) []*Region` — fraksiyonun yalnızca kara bölgeleri

`SelectBattleDefender(attacker, target, navalSeaMove)` — hedef bölgede saldıranı karşılayacak düşman orduyu deterministik seçer; kara savaşında en güçlü savunucuyu, deniz savaşında ise yalnız `StanceWar` ilişkisine sahip filoları dikkate alır. Savaş preview modalı ile gerçek resolve aynı savunucuyu kullansın diye render ve game katmanı bu helper üzerinden bağlanır.

`SiegeAt(regionID)` / `SiegeByArmy(armyID)` — aktif kuşatma kaydını bölge veya saldıran ordu üstünden döner; renderer, AI ve oyun mantığı aynı save verisini bu helper'larla okur.

`RegionProductionSummary(region) RegionProductionSummary` — seçili bölgenin efektif altın/mal üretimini hesaplar; bina çarpanları, arazi uzmanlaşması, mevsim ticaret/hasat etkileri ve sahip fraksiyonun ekonomi teknolojilerini UI önizlemesiyle paylaşır

`RegionBlockadeEconomicEffect(region)` liman ablukasını ekonomi katsayılarına çevirir: `%50` abluka bölgesel vergi, yerel ticaret ve kaynak üretiminin `%75`'ini bırakır; `%100` abluka `%50` bırakır. `RegionProductionSummary()` ve `applyEconomyTick()` aynı retention hesabını kullanır. `BlockadeLootForFaction()` ise etkili savaş gemisi katkısını deterministik paylaşarak ablukacıya sırasıyla `%5` veya `%10` oranında altın ve mal aktarır; kuşatma altındaki limanlar ekonomi tick'inde olduğu gibi loot tabanına dahil edilmez. `UnblockedRegionProductionSummary()` yalnız bu loot tabanı için kullanılır. `BlockadeLootGoldForFleet()` aynı zinciri tek bir abluka filosunun tooltip'te gösterilecek altın katkısına indirger.

`FindSettlementByID(settlementID)` — settlement ID'den region + settlement çözümlemesi yapar

`Region.SettlementPopulation()` / `Region.RecalculatePopulation()` — yerleşim nüfuslarını toplar ve `Population = RuralPopulation + yerleşim toplamı` sözleşmesini korur. Eski save'lerde yalnız toplam nüfus varsa fark kırsal nüfusa göç edilir.

`FactionCapital(fid)` — fraksiyonun geçerli başkent settlement ve bölgesini döner

`SetFactionCapital(fid, settlementID)` — başkenti anında değiştirir ve pending kuyruğu temizler

`StartCapitalMove(fid, settlementID, turns)` / `AdvanceCapitalMoves()` — 5 tur gibi gecikmeli başkent taşıma akışını yürütür

`NormalizeFactionCapitals()` — yükleme sonrası eksik/geçersiz başkentleri en yüksek getirili owned settlement'a normalize eder

`RegionLogisticsStatus` / `ArmyLogisticsStatus` — son turdaki bölgesel ikmal yükü, kapasite, abluka yüzdesi, aşım ve zayiat bilgisini render katmanına taşır; serialize edilmez.

`GrainEconomyStatus` / `GameState.GrainEconomy` — son ekonomi tick'inde fraksiyon bazlı tahıl üretimi, sivil tüketimi, ordu bakımı, ordu yenilemesi, ordu `ArmyMoraleDelta` değişimi, stratejik ithalat ihtiyacı, nüfus büyümesi ve otomatik ihracat için harcanan tahıl, net değişim, stok-ay seviyesi ve açık bilgisini render/event bildirimlerine taşır; runtime-only olduğu için save'e yazılmaz.

`GameState.TaxIncomeForFaction()` ve `GrainSaleGoldBudget()` — kuşatma dışı temel vergi gelirini ve o turda tahıl satışından kullanılabilecek kalan altın bütçesini hesaplar. Acil satış, otomatik ihracat ve oyuncunun doğrudan tahıl satışı aynı bütçeyi paylaşır; `GrainSaleGoldUsed` tur sonunda sıfırlanır ve save'e yazılmaz.

`GameState.ArmyMoveUsage` — `applySeasonEffects()` hareket puanlarını yenilemeden önce ordunun o tur hareket edip etmediğini geçici olarak yakalar. `GameState.EffectiveArmyGrainUpkeep()` bu snapshot'ı kuşatma ve garnizon katsayılarıyla birleştirir; ekonomi, bölgesel lojistik ve AI aynı efektif talebi kullanır. Alan serialize edilmez.

`GameState.GrainAidUsage` / `CanApplyGrainAid()` / `ApplyGrainAid()` — oyuncunun bölge panelinden yaptığı tahıl yardımını bölge başına turda bir kez sınırlar; 12 tahıl karşılığında memnuniyeti +10 artırır. Kullanım haritası `AdvanceTurn()` içinde sıfırlanır ve save'e yazılmaz.

`EmergencyGrainSaleLimit()` / `EmergencyGrainSaleUnitPrice()` / `ApplyEmergencyGrainSale()` — pazar partneri gerektirmeyen acil tahıl satışını yönetir. Yalnızca fraksiyon depo kapasitesi üzerindeki miktar satılır; `economy.EmergencySaleUnitPrice()` güncel fiyatın %70 indirimli değerini üretir ve miktar kalan vergi bazlı satış bütçesiyle sınırlanır.

`GameState.AutoGrainExport` / `ApplyAutomaticGrainExport()` — Pazar sekmesindeki tercihi ve ekonomi tick'inde kapasite üzeri tahılın aktif, savaşta olmayan ticaret ağı partnerlerine faction ID sırasıyla %60 fiyatla satışını yönetir. Alıcı altını yetersizse miktar alıcının bütçesiyle sınırlanır; tercih compact save alanında korunur, gerçekleşen miktar ve altın runtime `GrainEconomyStatus` içinde raporlanır.

`GameState.CanArmyReplenishIn()` / `ReplenishArmyInFriendlyTerritory()` — kendi, müttefik veya aynı realm içindeki vassal bölgesinde bulunan ordunun dost ikmal uygunluğunu ve ücretsiz HP uygulamasını ortaklaştırır; aktif kuşatma istisnası çözümleme katmanında korunur. `applyGrainFundedArmyReplenishment()` mevcut ücretsiz dost-toprak toparlanmasına ek olarak kapasite üstü tahılı dost ve kuşatma dışı kara ordularına aktarır. Faction/army ID sırası deterministiktir, ordu başına en fazla +10 HP verilir ve 1 HP başına 1 tahıl tüketilir; rezerv kapasitesi altına inilmez.

`GameState.CanFleetAvoidSeaAttrition()` — açık denizdeki filonun komşu deniz bölgesine bağlı limanlı kara bölgeleri arasında filo sahibinin kendi toprağını, aynı realm içindeki vassal toprağını veya müttefik toprağını arar. Bu kapı kış donanma yıpranması ile embarked sefer yıpranmasında birlikte kullanılır; limanlı dost kıyı komşuluğunda gemi ve taşınan birlik hasar almaz, limansız kıyıda normal yıpranma sürer.

`GameState.StrategicGrainDemand()` / `StrategicGrainSurplus()` — fraksiyonun üç aylık güvenli rezerv hedefindeki açığı ve kapasite üstü ihraç edilebilir stoku hesaplar. Diplomasi yeni rota kurarken bu iki sinyalle hedefteki tahıl ihtiyacını kaynak fazlasına bağlar; sinyal save'e yazılmaz ve her ekonomi tick'inde yeniden türetilir.

`RegionEventStatus` içindeki `GrainProductionPercent` ve `GrainDemandPercent`, aktif hasat/kıtlık/kuraklık olaylarının geçici bölgesel tahıl etkileridir. `RegionGrainProductionModifier()`, `RegionGrainDemandModifier()` ve `CivilianGrainDemandForRegion()` bu kayıtları toplar; alanlar `ActiveRegionEvents` ile compact save/load içinde korunur, süre dolunca `TickActiveRegionEvents()` tarafından temizlenir.

`GameState.RegionMilitaryGrainProduction()` bölgesel efektif tahıl üretiminden aktif sivil talebi düşer. Oyun lojistiği ve AI hareket/recruitment lojistiği bu ortak helper'ı; ordu talebi için de `EffectiveArmyGrainUpkeep()` metodunu kullanır. Böylece oyuncu ve AI aynı tahıl tüketim kurallarından sapmaz.

`applyRegionalLogisticsPressure()` bölgedeki ambar seviyesini, ekonomi tick'i sonrası fraksiyon stokundan bölgeye aktarılabilir rezerv olarak değerlendirir. Destek `min(kalan stok, ambar kapasitesi)` ile sınırlıdır; aynı fraksiyonun bölgeleri rezervi deterministik sırada paylaşır ve başkent önceliği kullanır. `RegionLogisticsStatus.GranarySupport` bu geçici katkıyı UI teşhisi için taşır.

`GrainStorageCapacity()` ve `GameState.GrainStorageCapacityForFaction()` sivil nüfus talebi, efektif ordu bakımı ve ambar bina bonusunu aynı `6 ay sivil + 3 ay ordu`, minimum 100 kapasite kuralında birleştirir. İkinci helper ekonomi tick'i oluşmadan HUD'un başlangıçta da doğru ambar kapasitesini gösterebilmesini sağlar.

`Army.Morale` ordunun kalıcı ikmal moralidir. `CurrentMorale()` eski kayıt veya fixture'larda eksik alanı 100 başlangıç morali olarak normalize eder; `ApplyMoraleDelta()` değeri 1–100 aralığında tutar. Compact save/load içindeki `mo` alanıyla taşınır ve `Army.TotalStrength()` içinde savaş/AI güç değerlendirmelerine uygulanır.

`GameState.NormalizeEmptyArmies()` birim veya taşınmış birlik içermeyen artık ordu
kayıtlarını kaldırır ve varsa komutanı havuza bırakır. Save yüklemede legacy garrison
normalizasyonundan sonra, tur çözümünün sonunda da çalışır; böylece eski edit/save
kayıtlarındaki boş ordu nesneleri AI manpower/komutan/UI hesaplarını kirletmez.

`TradeRoute.BlockadePercent` — rota uçlarındaki denizlerde bulunan, açık denizdeki düşman savaş gemilerinden türetilen geçici hacim kesintisidir. `RefreshTradeRouteBlockades()` ve `RegionBlockadePercent()` konum/savaş state'inden her ekonomi tick'inde yeniden hesaplar; limana bağlı filo bu deniz bölgesinde sayılmaz ve save migration gerektirmez.

Merchant filo görevi `Army.TradeRouteKey` ile kalıcı tutulur. `MerchantTradeRoutesForFleet()` yalnız filonun sahibi olan fraksiyonun uçlarında bulunan, aktif ve geçerli ticaret merkezi denizine sahip rotaları döndürür; `SetMerchantTradeRoute()` oyuncu UI'sından gelen atamayı aynı state doğrulamasından geçirir. Rota anahtarı save/load ile korunur, merchant hacim bonusu ise ekonomi tick'inde gerçek filo konumu ve görevinden yeniden türetilir (`internal/state/merchant_trade.go`).

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

Save göçünde `difficulty` yalnız `1..3` aralığında geçerli kabul edilir. Eski kayıtlar
bu alanı `0` taşıyabiliyorsa `internal/save/compact.go:applyCampaignSaveState()` değeri
Normal (`2`) olarak düzeltir. Yeni oyun reset'i ise artık boş/önceki `GameState` değerini
değil ayarlar ekranındaki `renderer.CurrentSettings.Difficulty` değerini kullanır; böylece
oyuncu zorluk seçimi ilk turda AI politikasına gerçekten ulaşır.

`Game` katmanında ayrıca serialize edilmeyen bir `pendingConquestDecisions` kuyruğu vardır. Bu runtime kuyruk, oyuncu savaşta bir devletin son kara toprağını düşürdüğünde battle report ile nihai ilhak/vassallık kararını birbirinden ayırmak için kullanılır. `SuccessorFactionID` taşıyan fetihlerde yalnız `GameState.CanRestoreSuccessorAtRegion()` true ise üçlü ardıl kararını (`İlhak Et`, `Serbest Bırak`, `Vassal Yap`) taşır; aktif ardıl devlet metadata'sı doğrudan ilhaka gider. Save/load veya yeni oyun başlangıcında temizlenir. Ardıl kararları `internal/game/conquest_decision.go`, renderer modalı ise `internal/render/renderer_dialogs.go` içinde tutulur.

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
