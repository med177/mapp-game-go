---
type: system
tags: [ai, strategy, coalition, difficulty]
last_updated: 2026-08-06
related: [systems/combat, systems/diplomacy, systems/economy, systems/victory, architecture/game-loop, architecture/state-management]
---

# Yapay Zeka Sistemi

1300 senaryosunda AI ekonomi ve askeri üretim kararları bölgesel kaynak uzmanlaşmasını ve ortak `ResourceCost` sözleşmesini kullanır. Pazar/liman/ibadet yeri ile elit kara ve deniz birliklerindeki baharat/kumaş maliyetleri `internal/ai/{building_investment.go,recruitment_strategy.go,unit_composition.go,naval_mission.go}` üzerinden aynı affordability ve bütçe akışına bağlanır. Tahıl yatırımı, sivil tüketim ile ordu bakımının toplam açığına göre puanlanır; tamamlanmamış en fazla iki çiftlik aynı anda kuyruğa alınabilir. Tahıl açığı yaşayan devletler, açık pazarda savaşta olmadıkları ve kendi üç aylık rezervini koruyan satıcılardan tahıl alır; üretim kararındaki birim, bina, nakliye veya savaş gemisinin eksik tahıl, demir, kereste, taş, baharat ve kumaş maliyetleri `aiProcureStrategicResources()` tarafından otomatik çıkarılır ve yeterli altın kaldığı sürece aynı açık pazardan tamamlanır. Abluka altındaki limanlar, somut çıkarma görevi olmasa bile deniz tehdidi snapshot'ından seçilir; gerekli liman seviyesi ve `%110` savunma gücü tamamlanana kadar savaş gemisi üretimi planlanır. Askerî bütçe ve ilk kışla kararı bu rezervleri koruyacak şekilde çalışır.

**Kaynak:** `internal/ai/ai.go`, `internal/ai/turn_stepper.go`,
`internal/ai/strategic_plan.go`, `internal/ai/fronts.go`, `internal/ai/rally.go`,
`internal/ai/retreat.go`, `internal/ai/security.go`, `internal/ai/pathfinding.go`,
`internal/ai/budget.go`, `internal/ai/building_investment.go`,
`internal/ai/unit_composition.go`, `internal/ai/recruitment_region.go`,
`internal/ai/research_strategy.go`, `internal/ai/naval_mission.go`,
`internal/ai/naval_threat.go`, `internal/ai/naval_patrol.go`, `internal/ai/merchant_trade.go`,
`internal/ai/conquest_policy.go`,
`internal/ai/difficulty_policy.go`, `internal/scenario/ai_strategy.go`,
`internal/ai/grain_procurement.go`, `internal/ai/tax_policy.go`,
`internal/diplomacy/war_fatigue.go`

## Memnuniyet ve vergi politikası

AI her tur başında bölgelerinin vergi oranını memnuniyet ve bağımsız savaş
sayısına göre ayarlar. Savaş yorgunluğu projeksiyonu overlord/vassal realm'lerini
tek devlet sayar ve ekonomi tick'indeki gerçek `-2 × bağımsız düşman` etkisini
önceden hesaba katar. Projeksiyon 35'in altındaysa vergi `20` puan, 50'nin
altındaysa `10` puan azaltılır; amaç `Satisfaction < 30` isyan kontrolüne
gelmeden bölgeyi ve gelir tabanını korumaktır. Projeksiyon 75 veya üzerindeyse
vergi `10` puan artırılır; böylece güvenli memnuniyet seviyelerinde gelir
toplanır. Vergi değişimi yalnız AI'nin sahip olduğu kara bölgelerinde uygulanır.

1300 bina yatırım skoru da aynı savaş yorgunluğu projeksiyonunu kullanır.
Memnuniyet açığı büyüdükçe `temple` (`+10`) ve uygun savunma bölgelerinde
`walls` (`+6`) gibi istikrar sağlayan binalar öne çıkar; ibadet yeri için
`Satisfaction < 30` ek öncelik uygulanır.

Üretim tedarikinde askerî aday seçimi ile gerçek üretim önkoşulları birlikte
değerlendirilir. Tedarik öncesi kaynak baskısı puanı uygulanmadığı için eksik
demir, AI'yi demirsiz milise kaçırmaz; demir/kereste isteyen piyade veya kuşatma
adayının maliyeti pazar talebine dönüşür. Geçerli askerî üretim bölgesi kışla
eksikliği nedeniyle bulunamıyorsa ve stratejik kapasite/ordu limiti uygunsa,
kışlanın `ResourceCost` içindeki tüm eksik girdileri de aynı turda satın alınır.
`aiNeedsBarracksForMilitaryProduction()` bu kararı hem tedarik öncesi hem de
kışla kuyruğa alma adımında paylaşır; böylece kışla maliyeti satın alınmadan
kışla üretiminin kilitlenmesi önlenir. Tedarik, aktif rota gerektirmeyen açık
pazarda savaşta olmayan, stok güvenlik payı bırakan tedarikçilerden ve
`aiMinGoldReserve` korunarak yapılır.

Her AI turunun başında `RefreshMarketOrders()` aynı kararları açık pazar emir
defterine yazar. Satış arzı stratejik rezerv ve mevcut hedef maliyetleri
üzerindeki stok fazlasıdır; alım talebi eksik hammaddeyi ve üç aylık tahıl
rezerv açığını taşır. AI tedariki bu emirleri tükettiği için panelde görünen
arz/talep ile gerçek alım üst sınırı aynı state değeridir.

AI ordusu düşman toprağında görünür savunucu yoksa mevcut bölge görevi seçebilir.
Ana fetih planındaki hedef bölge görevle geciktirilmez; AI normal taarruz/kuşatma
akışına döner. Diğer bölgelerde karar deterministiktir: komşu düşman ordusunun
bir sonraki hamlede ulaşabildiği, pusu bonuslu arazi ve kuvvet dengesi pusu
puanını; gerçek `RaidLootPreview` altın/kaynak çıktısı yağma puanını belirler.
Belirgin biçimde güçlü karşı taarruz varken AI açıkta yağma veya pusu yapmaz.
Pusu orduları AI hedef seçimi ve normal bölge savunucu taramasında gizlidir;
düşman ordu bölgeye girdiğinde `SelectAmbushDefender` ile özel temas tetiklenir.
AI pusu tarafının çatışma bonusu da oyuncu ile aynı arazi `AmbushBonus`
değerinden gelir.

`scenario.json` içindeki `victory_conditions` de AI için stratejik girdidir.
Bir fraksiyona `allowed_factions` ile özel tanımlanmış tarihsel hedef varsa,
1300 dışı senaryolarda AI önce bu hedefin eksik ve kara sınırından erişilebilir
bölgelerinin sahibini genişleme planı yapar. Böyle bir hedefi olmayan devletler,
genel zafer koşullarındaki erişilebilir bölgesel hedefleri; sabit bölgesi olmayan
askerî zaferde ise en uygun kara komşusunu kullanır. Ekonomik ve hayatta kalma
hedefleri konsolidasyon planına dönüşür. 1300'de mevcut profil, erken dönem
tarih/yıl/event kapılarını korur; victory hedefi profil yoksa yedek yön olur ve
savaş fırsatı puanını besler. Bu niyet `victory:<option-id>` olarak save/load
arasında korunur; koalisyon gücü, ateşkes, lojistik ve kritik tehdit kurallarını
bypass etmez. `ScenarioVictories` yüklü tüm senaryolarda çalışır.

`assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`, başlangıçta
elimine olan Ragusa ve Burgonya dahil senaryodaki her fraksiyon için profil
taşır. Elimine devletler normal turda plan üretmese de özgürleşip yeniden
kurulduklarında aynı amaçlarını kullanır. `Test1300ScenarioAIStrategyReferencesExist`
bu tam kapsama ile hedef devlet/bölge referanslarını birlikte zorunlu tutar.

## Genel Yapı

AI diplomasi kapanışında `PeaceAssessment` savaş yorgunluğu ile altın, tahıl,
memnuniyet ve ilişki baskılarını raporlar. `AssessPeaceSettlement` AI-AI
barışını beyaz barış, bölge bırakma, tazminat veya vassallık sonucuna ayırır;
`ExecuteAIPeace` sonucu uygular. Oyuncuya gelen bekleyen barış teklifi, açık
seçim olmadan toprak veya altın kaybettirmez.

Barış kararı artık tüm senaryolarda ortak `AssessPeaceDesire()` akışından geçer.
İlk dört savaş turunda olağan teklif üretilmez; başkent tehdidi veya askerî
çöküş acil durum istisnasıdır. AI, `TerritorialClaims`, aktif expand planı ve
`WarLedger` hedefiyle düşmanın tuttuğu bölgeleri kontrol eder. Core işgali
acil durum yoksa barış kabulünü kapatır; normal claim değeri eşik ve skor
üzerinde baskı oluşturur. Savaş sonrası relation tabanı da hedef sonucuna göre
`-45`, `-60` veya `-70` olur; böylece AI savaşları bir-iki tur sonra barışla
ve ilişki puanını hızlıca sıfırlayarak kapatamaz.

Senaryo başında her devletin sahip olduğu bütün kara bölgeleri `core` kabul edilir.
Strateji `territorial_claims` kayıtları ile objective içindeki
`territorial_claims` bölgeleri kalıcı `TerritorialClaims` kayıtlarına dönüştürülür;
`expansion_targets` yalnız genel
AI savaş fallback'i için runtime uyumluluk hedeflerini sağlar;
böylece AI bir hedefin yalnızca başkentini değil, tanımlı başlangıç topraklarının
tamamını korumaya veya geri almaya çalışır.

`BuildAIDiagnosticSnapshot` planı, cephe hedeflerini, güç/tehdit değerlerini,
ordu rol dağılımını ve lojistik/yedek kuvvet bloklanma nedenlerini tek runtime
çıktısında toplar. Normal save state'ine yazılmaz; debug paneli ve senaryo tempo raporu için
gerçek AI karar context'inin aynısını kullanır.

Savaş ilanı fırsatı yalnız hedef devletin askerî gücüyle değerlendirilmez.
`internal/ai/war_strategy.go:aiWarCoalitionAssessment`, hedefin otomatik katılan
vassallarını ve hedefin dış müttefiklerini savunma koalisyonuna ekler. Saldıran
tarafta `AssessWarCall().AutoJoin` sonucu kesin olanlar ile çağrıyı en az `%70`
olasılıkla kabul edecek AI müttefikleri ve onların vassalları hesaba katılır;
oyuncuya gönderilen bekleyen katılım teklifi destek sayılmaz. Müttefik ve vassal
ordularının katkısı, ordunun hedefin kara bölgelerine komşuluk grafiğindeki
mesafesine göre `%100 / %75 / %50 / %25 / %10` ağırlıklandırılır. Ağırlıklı
savunma koalisyonu gücü saldırı eşiğini yükseltir; güvenilir saldıran müttefik
gücü ise saldırı gücünü artırır.

Aktif `expand` planı ayrıca doğrudan savaş yoludur: plan hedefiyle barıştaki AI,
normal fırsat taramasını beklemeden hedefin koalisyonu ve kendi lojistiği hazırsa
savaş açar. Toplam güç henüz normal `%115` eşiğinde değilse, yalnız hedeflenen
bölge doğrudan kara sınırındaysa ve AI sınır kuvveti karşı tarafın sınır gücünün
en az `%125`iyse sınırlı hızlı fetih için `%85` toplam güç eşiği uygulanır.
Bu istisna tahıl krizi, rally, kritik tehdit, hedef müttefikleri ve cephe kuvveti
kontrollerini atlamaz. Tek başına yetersiz hedef sahibi, önce hedefe sınırı veya
aynı hedef planı bulunan; ilişki ve ittifak değerlendirmesini geçen devlete
ittifak teklif eder. Kabul edilen AI ittifakı aynı turda güvenilir savaş çağrısı
gücü olarak yeniden değerlendirilir.

Barış döneminde savunma objective'i genişleme objective'ini otomatik olarak
gölgelemez. Kritik tehdit yoksa erişilebilir `expand` hedeflerine stratejik
öncelik bonusu verilir; savunma planında saldırı rolü atanmamış olsa bile sınır
kuvveti, güç ve koalisyon kontrollerini geçerse fırsat savaşı için kullanılabilir.
Lojistik uyarısı bu fırsat savaşlarını tamamen kilitlemez; yalnız gerçek tahıl
krizi yeni saldırıyı durdurur. Böylece AI sınırlarında pasif beklemek yerine
hazırlık, baskı ve sınırlı genişleme arasında geçiş yapar.

Kaynak/test: `internal/ai/strategic_plan.go`, `internal/ai/fronts.go`,
`internal/ai/fronts_test.go`, `internal/ai/rally_test.go`

Geliştirme modunda bu snapshot'lar `quicksave.debug.json` veya
`autosave.debug.json` sidecar'ındaki `state.ai_diagnostics` alanına eklenir.
Oyun içinde aynı veri `F3` ile açılan modalda gösterilir; `TAB` devletler arasında,
mouse tekeri teşhis satırları arasında geçiş yapar. Bu görünüm yalnız geliştirme
modunda aktiftir.

Geliştirme save'i yüklendiğinde `GameState` beş AI fazı boyunca geçici bir
`AIDiagnosticHistoryEntry` listesi toplar. Her kayıt tur, plan türü, hedef bölge,
aktif savaş/cephe sayısı, yedek kuvvet seviyesi ve bloklanma nedenlerini taşır.
Beşinci AI fazı tamamlanınca history, normal sıkıştırılmış save'e eklenmeden
`autosave.debug.json` içindeki `state.ai_diagnostic_history` alanına yazılır; F3
modalı seçili devletin bu beş turdaki satırlarını karşılaştırmalı olarak gösterir.

Her `PhaseAITurn` artık iki katmandan oluşur:

- `ai.TakeTurn()` hâlâ tam turu tek çağrıda çözebilen saf AI entrypoint'idir.
- `ai.TurnStepper` ise aynı mantığı adım adım açar; oyun döngüsü bunu kullanarak her AI devletini sırayla görünür işler.

Oyun katmanı AI fraksiyonlarını `FactionOrder` sırasıyla dolaşır, her fraksiyon için `TurnStepper.Step()` çağırır ve her step arasında kısa bekleme ekler. Böylece harita bir anda "snap" olmaz; yakın cephedeki hareketler tek tek görünür.

AI kararlarında fraksiyon, bölge, ordu, teknoloji ve konsolidasyon adayları ID sırasıyla değerlendirilir. Eşit puanlı seçimler Go map iterasyon sırasına bağlı değildir. `TakeTurn` ve gerçek oyun akışındaki `TurnStepper`, seçilen aksiyon hareket puanı tüketmeden geri dönerse ordunun kalan hareketini sıfırlar; engellenmiş hedef veya başarısız genel hücum aynı hedefi sonsuz kez yeniden seçemez.

Deniz temasında AI tarafı kararını temas modalı açılmadan hemen önce verir. Filo
gücü karşı tarafın gücünün `%125` eşiğini aşıyorsa ve en az bir hareket puanı
varsa `Geri Çekil` seçer; bu karar filoyu gerçek bir deniz komşusuna taşır ve
geri çekilme için 2 hareket puanı harcar; girişte harcanan puanla birlikte kalan
puan sıfıra kadar düşebilir. Hareket puanı kalmayan AI
geri çekilemez. Güçleri yakın görevsiz veya devriye filoları `Çatış`; abluka,
escort ve nakliye filoları normalde `Pozisyonu Koru` tutumunu kullanır. Geri
çekilme rotası önce düşman filosu olmayan deniz komşusunu seçer. Bu
rota düşmanın geldiği kaynak denizi de dışarıda bırakır; güvenli hedef yoksa AI
geri çekilme kararı vermez. Bu varsayılan, güçlü düşman karşısında geri
çekilme kararını engellemez.

Oyuncu filosu içermeyen savaş açılışı veya deniz hareketi teması
`PendingNavalContact` içinde bekletilmez. `ResolveAIOnlyNavalContact()` iki AI
filosunun çatışma/geri çekilme kararını aynı AI turunda çözer ve geçici temas
state'ini temizler; oyun scheduler'ı yalnız oyuncunun karar modalını gerektiren
temaslarda bekler (`internal/ai/naval_contact.go`, `internal/game/game.go`).

Hedef puanlama boyunca `moveScoreContext`, manpower doluluk durumunu, bölgedeki orduları ve lojistik özetlerini tek hareket kapsamında cache'ler. Hareket uygulandıktan sonra context atılır ve sonraki adım güncel state üzerinden yeniden kurulur.

1300'de her stratejik tur üç ay kapsar; AI plan ufku ve savaş ilanı sıklığı bu
takvim hızına göre yeniden ölçeklenmiştir. Normal zorluk iki tur (altı ay) ufukla
planını sık günceller, dört tur (bir yıl) savaş ilanı aralığını korur. AI'nin
toplanma, barış değerlendirmesi ve durgun savaş eşikleri de aynı gerçek zaman
anlamını koruyacak biçimde tur bazında kısaltılmıştır. Bina, teknoloji ve birlik
üretimi tur bazlı kalır; süreleri her senaryonun kendi `data/{buildings,technologies,units}.json`
dosyasında doğrudan iki katına çıkarılmıştır. Runtime'da gizli bir süre çarpanı yoktur.

`internal/game/scenario_balance_test.go`, yalnız `1300_ottoman_rise` için deterministik tempo harness'ıdır. `RUN_SCENARIO_TEMPO_REPORT=fast|medium|calibration` sırasıyla 12x2, 42x4 ve 120x8 kapsamını çalıştırır; `SCENARIO_TEMPO_TURNS/RUNS` ile kontrollü override, `SCENARIO_TEMPO_DIFFICULTY=1|2|3` ile zorluk karşılaştırması destekler. Go 1.25'in `rand.Seed` no-op varsayılanı test kapsamında `randseednop=0` ile kapatılır ve savaş zarları tur/fraksiyon/step scope'una ayrılır. Aynı seed'in iki turluk tam state replay testi ile benchmark da bu dosyadadır.

Tempo raporu her profil için devlet bazında `wars_started`, aktif savaş-turu,
tamamlanan savaşların ortalama süresi, fetih, barış ve stalemate sayaçlarını da
çıkarır. Bu telemetry yalnız test harness'inde tutulur; `GameState` ve save formatı
genişletilmez. 12 turluk tahıl bandı, 42 turluk medium tempo ve 120 turluk calibration
raporu aynı savaş davranış görünürlüğünü kullanır; 42 aylık altın bantları yalnız
medium profilde doğrulanır.
Hareket karar refactor'ı sonrasında replay testi yine geçti. CPU/allocasyon profiliyle
diplomasi tehdit snapshot'ı, paylaşılan ticaret erişimi ve state order cache'leri eklendi;
42x8 Normal ölçümü `60.416 sn`den `55.658 sn` test süresine indi (`58.568 sn` duvar saati).
Nihai 60 saniye hedefi karşılanmıştır.

2026-07-19 objective dikey dilimi ölçümünde fast profil `9.08 sn`, medium profil
`59.89 sn` sürdü. Medium 42x4 sonuçta Osmanlı ortalama `2.0 → 7.8` kara bölgesine
ulaştı; bu sayı tarihsel bir sonucu zorlayan kabul şartı değil, sonraki kalibrasyonlar
için yön/tempo referansıdır.

### 1300 Çok-Seed Kabul Bantları

`RUN_SCENARIO_TEMPO_REPORT=medium` 42 tur ve 4 seed çalıştırıldığında tempo raporu
ortalama altın kazanımını aşağıdaki sözleşmeyle doğrular:

| Fraksiyon grubu | 42 aylık altın kazanımı |
|---|---:|
| Memlük | `12.000–30.000` |
| İngiltere | `17.000–32.000` |
| HRE | `15.000–32.000` |
| Fransa | `12.000–30.000` |
| İlhanlı | `10.000–30.000` |
| Venedik | `9.000–22.000` |
| Osmanlı | `-2.000–6.000` |
| Safevî | `500–5.000` |

Bantlar tarihsel sonucu sabitlemez; ekonomi teknolojisi kalibrasyonu, aktif savaşlar,
bina/ordu harcaması ve bölge kazanımı birlikte ölçülür. Bant dışına çıkılması, yeni
AI/ekonomi değişikliğinin tempo incelemesi gerektirdiğini gösterir. Bant kontrolü
yalnız 42 tur/4 seed medium profilinde yapılır; 120 tur/8 seed `calibration` profili
uzun dönem birikimi raporlar ve 42 aylık toplam bantla karşılaştırılmaz. `fast`
profili 12 tur x 2 seed hızlı regresyon, iki turluk tam state replay testi ise
deterministiklik kontrolüdür.

Kaynak/test: `internal/game/scenario_balance_test.go` içindeki
`assert1300CalibrationBands`, `Test1300ScenarioTempoReport` ve
`Test1300ScenarioAITwoTurnReplayIsDeterministic`.

### 1300 Tahıl Ekonomisi Faz 6 Bantları

`Test1300ScenarioGrainEconomyBands`, 12 turu iki seed ile erken (`1–4`), orta
(`5–8`) ve savaş/ileri (`9–12`) pencerelerine ayırır. Her pencere ve büyük
fraksiyon için `production`, `civilian demand`, `army upkeep`, `net change`,
`stockpile months` ve kıtlık oranı raporlanır. Kabul sözleşmesi:

- üretim / sivil talep oranı: `1.0–4.0`
- net değişim / sivil talep oranı: `-1.0–2.5`
- kıtlık oranı: dengeyi teşhis eden rapor metriği; ithalat baskısı olan Venedik
  ve erken Osmanlı gibi profillerin negatif fazı başarısızlık sayılmaz

AI hareket/ikmal kararları `GameState.RegionMilitaryGrainProduction()` ve
`GameState.EffectiveArmyGrainUpkeep()` üzerinden oyuncu ekonomi tick'iyle aynı
üretim, sivil talep ve ordu bakım kurallarını kullanır. Bu parity için
`internal/ai/grain_economy_test.go` regresyonu vardır.

### Kalıcı Stratejik Plan

Yalnız `1300_ottoman_rise` senaryosunda her AI devletinin tur öncesinde bir `StrategicContext` üretilir. Bu runtime context manpower kapasitesi, konuşlandırılmış kara birimi, sahip olunan/sınır bölgeleri, aktif savaşlar ve ihtiyaç halinde hesaplanan devlet gücü, frontier gücü ve bölge değerini cache'ler; `GameState` veya save payload'ına yazılmaz.

Kalıcı karar `GameState.AIPlans[factionID]` içindeki `AIPlanState` kaydıdır:

- `kind`: `expand`, `defend` veya `consolidate`
- hedef devlet ve zorluk politikasına göre öncelik sıralı `3/4/5` hedef bölge
- plan başlangıcı ve zorluk politikasına göre `4/6/9` turluk yeniden değerlendirme tarihi
- devlet agresifliğinden türeyen commitment
- savaş sonrası vassallığa izin verilip verilmediği ve doğrudan ilhak edilecek stratejik bölgeler
- debug/kalibrasyon için karar nedeni

Senaryo AI config'inde `ally` objective'i diplomatik yönelim metadata'sı olarak
tanımlanabilir. Bu kayıtlar `LoadAIConfig()` tarafından kabul edilir ancak askeri
planlayıcıya `AIPlanState` olarak yazılmaz; böylece ittifak niyeti yanlışlıkla saldırı
veya yabancı bölge yürüyüşü hedefi haline gelmez.

1300 senaryosunun statik yönleri `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`
dosyasından yüklenir. `GameState.AIStrategies` runtime-only'dir ve save'e yazılmaz;
aynı dosyadaki `GameState.AIDifficultyPolicy` de runtime-only tutulur. İkisi save
yüklenirken senaryo baz state'iyle yeniden kurulur. Her objective bölgesel claim,
öncelik, commitment, readiness bölgeleri ve savaş sonrası düzen tercihini taşıyabilir;
claim bölgesinin güncel sahibi hedef devlet olarak çözülür.
Tarihsel hedefler genel olarak soft yönelimdir: mevcut güç, kara sınırı, frontier gücü ve
diplomasi güvenlik kontrollerini atlamaz. Yalnız geç veya anakronik hedefler `min_year`,
`max_year` ve isteğe bağlı event flag hard gate'leri arkasında tutulur. `max_year`
verilen yılın sonuna kadar kapsayıcıdır; son yıllarda objective puanına zaman baskısı
eklenir, sonraki yılda hedef artık seçilmez.

İlk dikey dilimde Osmanlı; Bitinya hattı, Anadolu beylikleri, Ankara koridoru,
Trakya/Konstantinopolis ve 1501 sonrası Safevi rekabeti yönlerini kullanır. Doğu Roma;
Konstantinopolis/Trakya ve Anadolu kıyı savunması ile Bitinya geri alma yönünü taşır.
Aktif objective hedef devleti savaş ilanı puanında, öncelikli bölgeleri ise yerel ve uzun
menzilli hareket puanında bonus alır. Hedef bölge tamamlandığında, hedef vassal realm'e
katıldığında, devlet elendiğinde veya `reassess_turn` geldiğinde plan yenilenir.
`expand` ve `consolidate` objective'leri bütün `territorial_claims` bölgeleri AI'nin
elindeyse tamamlanmış sayılır ve yeniden seçilmez. Sıradaki objective'in `min_year`,
`max_year` veya event kapısı henüz açılmadıysa AI geçici `consolidate:<faction_id>`
hazırlık planına geçer. `defend` objective'leri yalnız claim sahipliğine bakılarak
bitirilmez; çekirdek savunma niyeti aktif kalır. Objective seçimi güç farkını puanlar,
ancak seçimi tek başına güçle kilitlemez; savaş ilanı ve saldırı için güç, cephe ve
lojistik eşikleri ayrıca uygulanır.
Bir `consolidate` objective'i tamamlandıktan sonra claim bölgelerinden biri başka
bir devlete geçerse objective yeniden açılır; kaybedilen claim'lerin güncel sahipleri
dinamik hedef olarak seçilir ve plan `expand`/recovery türünde bu bölgeleri geri alma
hazırlığı yapar. Claim geri alındığında recovery planı yeniden tamamlanmış sayılır.
Profil bulunmayan 1300 devletleri `ai_expansion_targets`, aktif savaş ve konsolidasyon
fallback'ini kullanmaya devam eder.

### Anadolu Beylikleri Objective Kalibrasyonu

`1300_ottoman_rise` artık Anadolu'daki 13 küçük/orta aktörü tek bir genişleme profiliyle
çalıştırmaz. `ai_strategies.json` içindeki profiller; Ege'de Aydın-Menteşe-Saruhan ve
Karesi rekabetini, batı/orta hatta Germiyan-Hamid-Eşrefoğlu çekişmesini, Pontus'ta
Candar-Canik tamponunu ve güneyde Karaman-Ramazan-Dulkadir geçişini ayrı hedeflerle
tanımlar. Tek bölgeli Ahiler, Canik, Dulkadir, Eşrefoğlu, Karesi ve Ramazan önce kendi
geçit/çekirdek bölgelerini savunur; genişleme objective'i bulunan devletler ise komşu
beylik hedefi sürdüğü müddetçe ilgili sınır bölgesine yönelir.

Yerel rakiplerin seçili başlangıç ilişkileri `-10` seviyesinde tutulur. Bu, aynı mezhep
bonusunu tamamen silmeden Normal/Zor AI'nin objective hedefi için savaş eşiğini
geçebilmesini sağlar. Genişleme objective'lerinde `allow_vassalization` vassallık
zarını açar; sonuç, `TryResolvePostWarVassalization` üzerinden hedefin gücü, savaş
planı ve fetih anındaki zar sonucuna göre vassal veya ilhak olur. Savunma objective'leri ise
aktif savaş ve cephe rolleriyle birleşerek küçük devletlerin plansız uzak fetihlere
gitmesini engeller.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`,
`assets/scenarios/1300_ottoman_rise/data/relations.json`.
Test: `internal/scenario/scenario_1300_integrity_test.go` içindeki Anadolu profil
sözleşmesi ve genel AI/scenario testleri.

### Venedik ve Ceneviz Deniz-Ticaret Profilleri

Venedik `adriatic_merchant_thalassocracy` profiliyle Venedik, Girit, Kıbrıs ve kuzey
Kıbrıs hattını aynı savunma objective'inde toplar. Konstantinopolis/Trakya yönündeki
`restore_eastern_mediterranean_trade_gate_1340` seferi ancak 1340'tan sonra açılır;
böylece erken oyun ada savunması ve ticaret filosunu kurmadan uzak bir kuşatmaya
dönüşmez. Ceneviz `western_merchant_network` profili Cenova, Korsika ve Kırım ticaret
merkezlerini korur; Trabzon Karadeniz kapısı sonraki genişleme yönüdür.

Bu objective'ler doğrudan bedava filo üretmez. Mevcut `merchant_trade.go` akışı aktif
trade route'ları en az kapsanan merkezden doldurur, tehditli merkezde merchant gemisinden
önce `%110` escort eşiğini tamamlar ve yalnız gerekli liman seviyesini yükseltir. Böylece
profil, liman/ada/ticaret merkezi önceliğini gerçek üretim ve deniz tehdidi kapılarına
aktarır; Venedik-Ceneviz rekabeti başlangıçtaki `-10` ilişkiyle korunur.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`,
`internal/ai/merchant_trade.go`, `internal/ai/naval_threat.go`.

Aktif savaşta liman ablukası görüldüğünde AI yalnız ticaret rotasının merchant
escort kararını beklemez. `aiProcureStrategicResources()` o turdaki gerçek üretim
maliyetlerinden tüm eksik ticari kaynakları aktif ticaret ağından satın alır;
`aiProcureMilitaryIron()` yalnız geriye dönük uyumluluk sarmalayıcısıdır.
`aiProduceNavalDefenseAtThreatenedPort()` ise liman seviyesi ve savaş gemisi
teknolojisi uygunsa tehdit gücünün `%110` eşiğine kadar savaş gemisi kuyruğu açar.
Kıyı/abluka tehdidi araştırma seçiminde de deniz teknolojisi ve savaş gemisi
açılımına öncelik verir. Böylece yüksek altın stoğu, eksik bir tek kaynak
nedeniyle sonsuza kadar kullanılmayan bir üretim bütçesine dönüşmez.

### Memlük ve İlhanlı Levant-Mezopotamya Cephesi

Memlük `levant_sultanate_frontier` profili 1320'ye kadar Şam-Halep-Ürdün-Mısır
koridorunu ve Kahire ticaret merkezini cephe rezerviyle korur. Ardından açılan
`break_ilkhanate_mesopotamian_front_1320` karşı taarruzu Bağdat ve Musul'u hedefler;
readiness bölgeleri Şam ve Halep'tir. Böylece başlangıç savaşı doğrudan, düşük maliyetli
bir sınır baskınıyla sonuçlanmaz.

İlhanlı `eastern_imperial_frontier` profili Şam/Halep/Ürdün yönünde baskı kurar;
Bağdat, Musul, Malatya ve Azerbaycan savunma çekirdeğidir. Böylece başlangıçtaki
`ilkhanate|mamluk` savaş ilişkisi yalnız genel fırsat savaşı olarak kalmaz, iki tarafın
cephe orduları, rally ve rezerv kararlarıyla aynı hedef devlet/bölge objective'ine
bağlanır. `TryResolvePostWarVassalization` bu iki büyük devlet arasında kullanılmaz;
stratejik bölgeler doğrudan ilhak listesinde tutulur.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`,
`assets/scenarios/1300_ottoman_rise/data/armies.json`.
Test: `internal/game/scenario_balance_test.go` içindeki Levant açılış plan testi.

### Balkan Devletleri ve Osmanlı Tehdidi

Sırp, Bulgar, Epir, Arnavut, Atina ve Eflak profilleri öncelikle yakın çekirdek ve
geçitlerini savunur. Sırbistan için Niş-Kosova-Raşka; Bulgaristan için Bulgaristan-
Vidin-Dobruca; Epir ve Arnavutluk için dağ geçitleri; Atina için kıyı; Eflak için
Wallachia-Oltenia-Besarabya tamponu korunur. Bu savunma objective'leri Osmanlı'yı
doğrudan komşu kabul etmek yerine Doğu Roma, Tuna ve Macar cephelerini büyüyen tehdit
göstergesi olarak kullanır.

Güvenlik planı bozulmadan yerel genişleme düşük öncelikli soft objective olarak kalır:
Sırbistan-Arnavutluk, Bulgaristan-Aşağı Tuna, Epir-Atina, Arnavutluk-Epir ve Eflak-
Dobruca yönleri. Böylece küçük devletler ilk turda uzak fethe kilitlenmez; sınır ordusu,
rezerv ve retreat kararları mevcut `AIFront`/`security` katmanında çalışmaya devam eder.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`.
Test: `internal/game/scenario_balance_test.go` içindeki Balkan açılış plan testi.

### Rusya, Altın Orda ve Baltık Cephesi

Rusya `moscow_consolidation` profiliyle Moskova, Nijni Novgorod, yeni doğu Rus
çekirdeği ve Dağıstan hattını güvenceye alır. `gather_rus_and_reach_black_sea_1478`
hedefi Novgorod ve Kırım'ı birlikte kapsar; 1478'e kadar kapalı olduğundan Rusya erken
oyunda Altın Orda sınırına bedelsiz bir bozkır akınıyla yönelmez. Altın Orda
`steppe_hegemony` profili Kiev-Ukrayna bozkırını ana cephe yapar, Rusya ve Litvanya
yönündeki baskıyı önceliklendirir ve Moldova/Kiev hattını savunma rezerviyle tutar.

Teuton Tarikatı `baltic_crusader_frontier` profili Konigsberg, Letonya ve Estonya
limanlarını korurken Litvanya sınırına baskı kurar. Novgorod `northern_trade_survival`
profili tek merkezli ticaret kapısını Teuton, Altın Orda ve Litvanya tehditlerine karşı
korur. Litvanya `eastern_baltic_expansion` profili Belarus üzerinden Kiev yönünü soft
genişleme hedefi olarak izler; Teuton ve Altın Orda baskısı arttığında Litvanya çekirdeği
savunması önceliği korur.

Bu yönelimler mevcut başlangıç orduları, negatif sınır ilişkileri ve kara/liman
altyapısıyla eşleşir; yeni bir savaş veya bedava kuvvet üretmez. Objective önceliği,
readiness bölgeleri, hedef devletler ve savaş sonrası ilhak listeleri `StrategicPlan`
katmanındaki güvenlik/cephe kontrollerinden geçer.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`,
`assets/scenarios/1300_ottoman_rise/data/regions.json`,
`assets/scenarios/1300_ottoman_rise/data/armies.json`.
Test: `internal/scenario/scenario_1300_integrity_test.go` profil sözleşmesi ve
`internal/game/scenario_balance_test.go` doğu bozkır/Baltık açılış plan testi.

### İngiltere-Fransa ve 1337 Tarihsel Savaş Kilidi

1300 başlangıcında İngiltere ile Fransa `-20` ilişki skoruyla barıştadır; iki devletin
`ai_expansion_targets` alanı boş bırakıldığı için genel fırsat savaşı taraması Yüz Yıl
Savaşı'nı erkenden başlatamaz. İngiltere'nin ilk yönelimi Kanal ve ada çekirdeğini
toparlamak, Fransa'nın ilk yönelimi ise Paris-Normandiya kraliyet çekirdeğini korumaktır.

`hundred_years_war_1337` tarihsel olayı Mayıs 1337'de tetiklenir. Olayın otomatik veya
oyuncu seçimiyle uygulanan kararı `hundred_years_war_started` bayrağını yazar ve
İngiltere-Fransa ilişkisini savaşa çevirir. Sonraki stratejik plan değerlendirmesinde
Fransa'nın Plantagenet Aquitaine'ini geri alma objective'i `min_year: 1337` ve event
bayrağı hard gate'ini geçerek açılır. İngiltere'nin Normandiya-Anjou-Paris seferi ise
aynı bayrağa ek olarak 1415'e kadar kapalıdır. Hard-gate objective'leri açıldığında
verilen aktivasyon bonusu, eski konsolidasyon planının yeni tarihsel cepheyi
gölgelemesini önler.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/events.json`,
`assets/scenarios/1300_ottoman_rise/data/relations.json`,
`assets/scenarios/1300_ottoman_rise/data/factions.json`,
`assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`.
Test: `internal/game/scenario_balance_test.go` içindeki
`Test1300EnglishFrenchWarWaitsFor1337Event`.

### 1300 Büyük Devletlerde Uzak Hedef Eşikleri

1300'deki ana devletlerin genişleme amaçları tek hamlede erişilebilen başlangıç
komşularından çıkarıldı. Osmanlı önce Bitinya-Bilecik güç tabanını kurar, 1354'te
Rumeli köprübaşını, 1453'te Konstantinopolis'i hedefler. Kutsal Roma İmparatorluğu
1311'den itibaren Milano-Floransa-Venedik yönünde imparatorluk otoritesini; Aragon
1416'dan itibaren Napoli-Apulya tacını; Portekiz de 1415'ten itibaren Fas köprübaşını
hedefler. Safevîlerin mevcut 1501 İran çekirdeği hedefi aynı uzun dönem kuralını
korur.

Bu eşikler hedefe ulaşıldığını garanti etmez: normal zorlukta AI önce ekonomi,
teknoloji, deniz lojistiği ve cephe gücünü kurar; hedef yılı geldikten sonra da savaş
ilanı, kuşatma, yağma veya pusu kararı mevcut güç/risk kontrollerinden geçer.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`.
Test: `Test1300MajorPowersUseHistoricalLongHorizonObjectives`.

### Safevîler: Erken Survival, 1501 Sonrası Yükseliş

Safevî profili 1300'de yalnız Güney İran'daki tek bölgelik Erdebil çekirdeğini korur;
erken dönemde İlhanlı veya Osmanlı'ya karşı genel expansion target'ı bulunmadığı için
zayıf bir tarikat devleti anakronik fetih savaşına sürüklenmez. İlk objective
`hold_southern_persian_core`, readiness ve yüksek commitment ile iç konsolidasyonu,
ordu rezervini ve geçiş güvenliğini öne alır.

`safavid_rise_1501` olayı Ocak 1501'de Safevî devleti ilanını temsil eder. Seçim/AI
uygulaması `safavid_rise` bayrağını, başlangıç kaynak takviyesini ve kompozit yay
teknolojisini verir. Bayrak ve yıl hard gate'i açıldıktan sonra `rise_into_persian_
heartland_1501` objective'i Azerbaycan, Batı/Kuzey İran ve Mezopotamya yönünü İlhanlı
hedefiyle açar; readiness bölgesi Güney İran, stratejik ilhak listesi ise İran çekirdeğidir.
Osmanlı'nın doğu rekabeti objective'i de aynı event bayrağına bağlı olduğundan iki taraf
1501 öncesi Safevî savaşına yönelmez.

Kaynak: `assets/scenarios/1300_ottoman_rise/data/events.json`,
`assets/scenarios/1300_ottoman_rise/data/ai_strategies.json`.
Test: `internal/game/scenario_balance_test.go` içindeki
`Test1300SafavidRiseWaitsFor1501Event`.

1300 açılışında Flandre için ayrı `vassal_trade_defense` profili kullanılır. Profil,
Flandre liman/ticaret gelirini korumayı, Fransa cephesindeki yerel savunmayı ve HRE'nin
Holland hattından gönderdiği yardımı önceliklendirir. Vassal AI üçüncü tarafla bağımsız
diplomasi başlatmaz; HRE'nin savaşı ve ticaret garantisi realm koalisyon kurallarından
gelir.

### Dinamik Acil Rezerv ve Harcama Bütçesi

`internal/ai/budget.go`, yalnız `1300_ottoman_rise` AI turlarında save'e yazılmayan bir
harcama bütçesi üretir. Acil altın rezervi şu formüldür:

`40 + sahip olunan kara bölgesi*8 + aktif savaş*30 + min(120, efektif aylık altın/3)`

Başkent veya kritik merkez tehdidi rezervi `+40` artırır. Sonuç en az `80`, en fazla
`420` altındır. Efektif aylık altın, ekonomi tick'iyle aynı
`RegionProductionSummary()` kaynağından gelir; böylece vergi ve memnuniyet etkisi ham
bölge değerinden tahmin edilmez. Rezervin altına indirecek hiçbir AI harcamasına izin
verilmez.

Rezerv üstündeki harcanabilir altın plan durumuna göre soft paylara ayrılır:

| Plan durumu | Ordu | Ekonomi | Araştırma | Donanma |
|---|---:|---:|---:|---:|
| Genişleme | %55 | %20 | %15 | %10 |
| Aktif savaş veya savunma | %70 | %10 | %10 | %10 |
| Konsolidasyon | %35 | %35 | %20 | %10 |

Kategoriler `araştırma → ekonomi → donanma → ordu` sırasıyla çalışır. Bir kategorinin
kullanamadığı pay aynı tur esnek havuza bırakılır; sonraki kategoriler kendi payıyla bu
havuzu birlikte kullanabilir. Ordu son tüketici olduğu için savaş hazırlığı gerekli
olduğunda atıl kalan yatırım bütçesini kullanabilir. Kıyısı olmayan devlette donanma
kategorisi oluşturulmaz ve pay kalan kategorilere kendi oranlarında dağıtılır. Bu model
runtime-only'dir; compact/legacy save şeması değişmez ve diğer senaryolar mevcut sabit
`80` altın rezerv/hardcoded harcama sırasını korur.

Bütçe dilimi sonrası ölçümde Normal fast 12x2 `8.87 sn`, Osmanlı `2 → 3`, güç `288`;
Normal medium 42x4 `62.22 sn`, Osmanlı ortalama `2 → 5`, güç `670`; Zor fast 12x2
`7.35 sn`, Osmanlı `2 → 3`, güç `289` sonucunu verdi.

AI bütçesi tahıl maliyetli bir emir sonrasında `OperationalGrainReserve` altına
inemez; hedef rezerv normalde iki aylık toplam talep, kritik tehditte en az bir
buçuk aylık talep olarak korunur. `aiProcureGrain()` üç aylık kapasite hedefi ile
iki aylık satın alma penceresini karşılaştırır, kaynakta da güvenli fazlalık bırakır.
Bağlı ticaret ağında doğrudan rota olmasa bile aktif rota grafiği üzerinden tedarikçi
bulunur. Böylece Venedik, Ceneviz ve Arnavutluk gibi yüksek altınlı ancak tahılsız
devletler üretim ve askerî harcama öncesi otomatik olarak stok toplamaya başlar.

### Bina Yatırım Puanlaması

1300 senaryosunda ekonomi bütçesi artık sabit `farm → market → walls` taramasını
kullanmaz. `internal/ai/building_investment.go`, her uygun bölgedeki ambar, pazar,
çiftlik, kıyı limanı, sur ve ibadet yeri adayını aynı skorda karşılaştırır.
Kışla ordu bütçesinde kalır; liman hem genel ticaret yatırımı hem de özel donanma
gereksinimleri tarafından değerlendirilebilir.

Skorun bileşenleri:

- **ROI:** Bina bölgeye sanal olarak eklenir; mevcut ve sonraki
  `RegionProductionSummary()` farkı 12 tur için hesaplanır. Altın doğrudan, tahıl güncel
  piyasa fiyatı ve devletin üretim/bakım baskısıyla altın eşdeğerine çevrilir. Maliyet
  de tüm hammaddelerin piyasa karşılığıyla hesaplanır.
- **Kaynak darboğazı:** Tahıl üretimi ordu bakımını karşılamıyorsa veya stok güvenlik
  eşiğinin altındaysa çiftlik yükselir. Bir aday mevcut demir/kereste/taş/tahıl stokunun
  büyük bölümünü tüketecekse fırsat maliyeti cezası alır.
- **Ambar ve askerî dayanıklılık:** İlk `granary` yatırımı doğrudan üretim vermese de
  depolama kapasitesi ve lojistik toparlanma sağladığı için güçlü başlangıç bonusu alır;
  tahıl stoğu, kapasitesi veya üretim-bakım dengesi zayıf devletlerde ambar öne çıkar.
- **Tehdit ve objective:** Aktif savaş, kritik cephe, başkent, defend hedefi ve yüksek
  yerel tehdit surları öne çıkarır. Expand rally/hedef sınırı çiftlik ve pazarı;
  consolidate planı uzun vadeli pazar, çiftlik ve istikrar yatırımlarını destekler.
- **İstikrar:** Bina memnuniyet bonusu mevcut açığa göre değerlenir; gerçek isyan
  eşiğindeki bölgede ibadet yeri acil bonus alır.
- **Ticaret kapasitesi hedefi:** Maksimum seviyeye yaklaşan pazar yatırımları
  kademeli ticaret skoru alır; son pazar seviyesi anlaşma başına `+2` hacim
  tavanını açtığı için güçlü ek bonus alır. Kıyı limanı seviyeleri aynı bölgede
  maksimum pazarla birlikte tamamlandığında `+1` dış partner limiti sağlar.
- **Fırsat maliyeti:** Uzun inşa süresi, aynı binanın seviyesi ve bölgedeki mevcut bina
  kuyruğu skoru düşürür.

Skor `80`in altındaysa AI sırf ekonomi payını tüketmek için bina kurmaz; kullanılmayan
altın aynı tur donanma/ordu kategorisine aktarılır. Tur başına tek bina sınırı korunur.
Eşitlikler sırasıyla toplam skor, ROI, objective, tehdit, kısa süre, region ID ve bina
ID ile deterministik çözülür. Kuşatma altındaki bölgede yeni yatırım başlatılmaz.
Diğer senaryoların eski farm/market/walls koşulları değişmemiştir.

Bu dilim sonrası Normal fast 12x2 `8.91 sn`, Osmanlı `2 → 3`, güç `268`; Normal medium
42x4 `62.12 sn`, Osmanlı ortalama `2 → 5.8`, güç `657`; Zor fast 12x2 `7.22 sn`,
Osmanlı `2 → 3`, güç `267` ölçüldü.

### Plan Bazlı Kara Ordu Kompozisyonu

`internal/ai/unit_composition.go`, yalnız 1300 senaryosunda üretilecek kara birimini
aktif stratejik plana bağlar:

| Plan | Piyade | Süvari | Kuşatma |
|---|---:|---:|---:|
| Genişleme | %55 | %25 | %20 |
| Savunma | %75 | %15 | %10 |
| Konsolidasyon/fallback | %65 | %25 | %10 |

Kompozisyon hesabına haritadaki kara orduları, filolarda taşınan birlikler ve bekleyen
kara üretim emirleri birlikte girer. Her uygun aday önce hedef orana göre kapatacağı
açıkla, ardından saldırı/savunma/moral değeri, altın eşdeğeri toplam maliyet, hammadde
stok baskısı, tahıl bakımı, **güç/tahıl verimi** ve üretim süresiyle puanlanır. Böylece
milisle yakın bakım taşıyan fakat çok daha yüksek savaş değeri veren elit birlik, gerekli
teknoloji, üretim hattı ve bütçe varsa tercih edilir; tahıl krizi ise mutlak bakım cezasını
artırmaya devam eder. Mevcut teknoloji, gerekli bina ve ordu bütçesi kontrolleri aynen
uygulanır.

Savaş bağlamı oyunun gerçek mekaniklerinden türetilir. Genişleme saldırıyı, savunma
savunma değerini öne çıkarır; hedef dağ/geçit/orman ise saldıran tarafta saldırı ve
moral, dost savunma hedefinde savunma ve moral ağırlığı artar. Düşman ordularının toplam
saldırı/savunma profili karşı ağırlığı değiştirir. Birim kategorilerine oyunda olmayan
arazi veya karşı-birim bonusları eklenmez. Tahkimli objective ya da aktif savaş hedefi
varsa, kuşatma desteği olmayan `assault/siege` orduları ile kuyruktaki kuşatma üretimi
karşılaştırılır; açık varsa kuşatma birimi güçlü öncelik alır. Plan ve aktif cephelerdeki
farklı tahkimli hedefler ayrıca sayılır: en fazla üç birliklik bir kuşatma kolu hedeflenir;
bu sayı, sahadaki ve kuyruktaki kuşatma birimleriyle kapatılana kadar piyade/süvari
oranından önce gelir.

Seçim skor ve bağlayıcı alanlarla deterministiktir. Model runtime-only'dir; save şeması
değişmez. Diğer senaryolar sabit elite piyade/ağır süvari/piyade sırasını kullanmaya
devam eder.

Bu dilim sonrası Normal fast 12x2 `9.08 sn`, Osmanlı `2 → 3`, güç `292`; Normal medium
42x4 `62.53 sn`, Osmanlı ortalama `2 → 5`, güç `733`; Zor fast 12x2 `7.39 sn`,
Osmanlı `2 → 3`, güç `290` ölçüldü.

### Stratejik Recruitment Bölgesi

`internal/ai/recruitment_region.go`, 1300 senaryosunda seçilen kara biriminin hangi
bölgede üretileceğini ortak bir skorla belirler. Adayın gerekli kışla seviyesi ve ordu
slotu bulunmalı, o tur kara üretim hattında boş throughput kalmalıdır.

Skor şu sinyalleri birlikte kullanır:

- kalan throughput ve kışla seviyesi,
- aynı üretim hattındaki ve bölgedeki toplam pending kuyruk,
- planın rally, savunma veya ilgili savaş cephesi anchor'ına ağırlıklı Dijkstra maliyeti,
- mevcut ordular, pending birlikler ve yeni aday sonrasındaki tahıl lojistiği boşluğu,
- komşu savaş düşmanı gücünün oluşturduğu güvenlik cezası.

Kuşatma altındaki, yabancı ordu barındıran, `Satisfaction < 30` gerçek isyan eşiğindeki
veya kritik/başkent tehdidi taşıyan cephenin parçası olan bölge doğrudan elenir. Mevcut,
pending ve yeni birliklerin toplam tahıl bakımı yerel kapasiteyi aşacaksa üretim hattı
da kullanılmaz. Kuşatma biriminde önce kuşatma desteği eksik `assault/siege` ordusunun
`FrontFactionID` cephesi seçilir; böylece ekipman genel bir arka bölgeye değil ilgili
hücum hattına yakın çıkar.

Rota hesabı mevcut güvenli dost-toprak Dijkstra motorunu ve AI turu route cache'ini
kullanır. Skor eşitliğinde rota, lojistik boşluk, throughput, seviye, kuyruk ve region
ID sırasıyla deterministik bağ kırıcıdır. Model runtime-only'dir. Diğer senaryolar
mevcut remaining-capacity, kışla seviyesi ve region ID seçimini korur.

Bu dilim sonrası Normal fast 12x2 `9.56 sn`, Osmanlı `2 → 3`, güç `278`; Normal medium
42x4 `67.44 sn`, Osmanlı ortalama `2 → 6.5`, güç `568`; Zor fast 12x2 `7.74 sn`,
Osmanlı `2 → 3`, güç `269` ölçüldü.

### Plan ve Darboğaz Bazlı Araştırma

`internal/ai/research_strategy.go`, yalnız 1300 senaryosunda araştırılabilir ve bütçeye
uygun teknolojileri aktif stratejik planla puanlar:

- **Genişleme:** askerî saldırı, hareket, tahkimli hedefte kuşatma ve ihtiyaç duyulan
  birim açılımı öne çıkar.
- **Savunma:** gerçek `land_defense_mod`, ekonomi, memnuniyet ve diplomatik nefes alanı
  daha değerlidir.
- **Konsolidasyon:** altın/tahıl/hammadde üretimi, pazar getirisi, istikrar, din dönüşümü
  ve barış ilişkisi ağırlık kazanır.

Ekonomi değeri mevcut bölge sayısı, ticaret kapasitesi ve gerçek üretim üzerinden 12
turluk marjinal getiriyle hesaplanır; tahıl bakım açığı ile düşük demir/kereste/taş
stoku kaynak modlarının faydasını yükseltir. Memnuniyet ve din farkı yalnız sahip olunan
bölgelerden türetilir. Kıyısız devlet saf deniz teknolojisinde ceza alır; ancak
`move_bonus` veya ekonomi etkisi gibi gerçekten kara devletine yarayan efektler yine
değerlendirilir.

Bir teknoloji `units.json` içindeki birimi doğrudan açıyorsa, birimin kategori açığı,
savaş değeri, bina erişimi ve tahkimli hedefte kuşatma desteği hesaba katılır. Adayın
tamamlanmasıyla diğer tüm önkoşulları sağlanacak bir sonraki teknoloji de küçük zincir
bonusu alır. Maliyet ve süre skoru düşürür. Mevcut savaş motoru saldırı modlarını toplu
uyguladığı için piyade/süvari/kuşatma saldırı efektlerine oyunda olmayan ayrı karşı-birim
faydaları uydurulmaz. Yalnız oyuncu ordu panelini açan `reveal_enemy_strength` AI için
puan üretmez.

Aktif araştırma yarıda bırakılmaz. Eşitlik; toplam skor, birim açılımı, gerçek efekt,
sonraki teknoloji, süre/maliyet ve teknoloji ID ile deterministik çözülür. Model
runtime-only'dir; diğer senaryolar mevcut sabit kategori sırasını korur.

Bu dilim sonrası ekonomi getirileri de senaryo verisinde kalibre edildi: Ticaret Yolları
bölge bonusu `+2`, Bankacılık `+3` ve `%10` pazar, Loncalar `%75` pazar, Tahrir
Defterleri `+3`, Kervansaray Ağı `+2` ve `%15` pazar, Darphane Standardı `+4` verir.
Araştırma maliyetleri, önkoşul zinciri ve askeri/deniz teknoloji değerleri değişmedi.
Normal medium 42x4 ölçümünde büyük devletler yaklaşık `20–25 bin` altın biriktirdi;
önceki `27–31 bin` bandına göre ekonomi hâlâ büyümeyi finanse ediyor ancak tek başına
sınırsız hazine birikimi üretmiyor. `Test1300EarlyEconomyTechnologyValuesAreCalibrated`
bu veri sözleşmesini korur.

### Cephe, Dinamik Rezerv ve Ordu Rolleri

`internal/ai/fronts.go`, 1300 senaryosundaki her AI turunda sınır bölgelerini komşu
devlet ekseninde `AIFront` kayıtlarına ayırır. Her cephe şunları taşır:

- dost/düşman sınır bölgeleri ve stratejik anchor,
- iki tarafın sınırdaki gerçek ordu gücü,
- savaş ve aktif objective ilişkisi,
- başkent, savunma objective'i veya kuşatma kaynaklı kritik tehdit,
- deterministik tehdit skoru.

Bu snapshot'tan mobil kara ordularına save'e yazılmayan görevler atanır:

| Rol | Davranış |
|---|---|
| `assault` | Aktif savaş cephesini, savaş yoksa kalıcı objective hedefini izler |
| `siege` | Aktif kuşatmayı korur; kuşatma birimli objective ordusunu tahkimat hedefine yöneltir |
| `defense` | Tehdit altındaki dost cephe/merkeze yaklaşır ve plansız fetih üretmez |
| `reserve` | Başkent veya kritik merkez anchor'ına çekilir, yabancı toprağa girmez |
| `relief` | Dost/aynı realm kuşatmasını kaldırmak için kuşatılan bölgeye yönelir |
| `retreat` | Yıpranmış/ezilen orduyu dost kara hattından güvenli ikmal anchor'ına çeker |
| `security` | İsyan riski taşıyan bölgeyi en küçük uygun saha ordusuyla güvenceye alır |

Normal durumda rezerv hedefi toplam mobil kara gücünün yaklaşık `%15`idir. Başkent,
savunma objective'i veya dost kritik merkez aktif tehdit/kuşatma altındaysa hedef `%30`a
çıkar. Stack granülerliği nedeniyle atanan güç hedefi aşabilir; küçük devletin bütün
ordusunu dondurmamak için tek saha ordusunda rezerv ayrılmaz ve çok ordulu devlette en
güçlü stack aktif bırakılır. Aktif kuşatma ve relief görevi rezervden önce atanır.
Birden fazla aktif savaş veya pozitif tehdit skoru taşıyan ama henüz kritik olmayan
cepheler varsa oran `%25`e çıkar; savaşlar yatıştığında `%15` tabanına döner. Bu yedek
oranı uzun savaşta recovery/ikmal için güvenli güç bırakır ve başkent tehdidindeki
`%30` kuralını ezmez.

Rol bonusları yalnız standart hareket/diplomasi kurallarının zaten geçerli saydığı
hamleleri sıralar; barıştaki yabancı hedefi veya yasak geçişi açamaz. Kalıcı objective
barıştaki başka bir devleti gösterirken aktif savaş varsa hücum orduları önce mevcut
savaşı sonuçlandırır. Yeni proaktif savaş da saldırı rolü gücü ile rezerv hedefi hazır
değilse veya kritik merkez tehdit altındaysa ertelenir. Kuşatma birimi olmadan kuşatma
başlatılabilir ve aktif kuşatmada genel hücum da seçilebilir; kuşatma birimi kuşatma
ilerlemesini hızlandırır, ancak genel hücum için zorunlu değildir.

Savunma veya konsolidasyon objective'i aktif bir savaşı pasif savunmaya kilitlemez.
`assignAIArmyRoles()` bu durumda kritik/başkent tehdidi yoksa en uygun tek saha ordusunu
aktif savaşın düşman sınırına `assault` veya kuşatma birimi varsa `siege` rolüyle gönderir;
diğer savaş cepheleri savunmada tutulur. Yeni başlayan savaşlarda önce 12 turluk
seferberlik penceresi korunur; savaş ledger'ı bu eşiği geçtiğinde cephe yeniden hücum
edebilir. Genişleme objective'leri kendi mevcut hücum akışını kullanmaya devam eder.
Bu ayrım, kayıtlı savaşların uzun süre sonuçsuz kalmasını düzeltirken 1300 açılışındaki
genişleme ve tahıl tempo bantlarını bozmaz.

Savaş hazırlığı ekonomiyle de sınırlıdır. 1300 senaryosunda ilk 24 tur tarihsel
açılış temposunu koruyan bir hazırlık penceresidir; sonrasında yeni proaktif savaş
ancak `prepareAIBudget()` tarafından hesaplanan altın acil rezervi ve en az iki aylık
operasyonel tahıl stoku korunuyorsa açılır. Runtime `GrainEconomyStatus` kritik/kıtlık
seviyesindeyse kapı kapanır. Aktif savaşlarda uyarı seviyesi cepheyi durdurmaz; kritik
ve kıtlık seviyesinde saldırı rolleri savunma/ikmal rolüne çekilir. Bu kural mevcut
`StrategicGrainDemand()`, `EffectiveArmyGrainUpkeep()` ve bölgesel ikmal hesaplarını
yeniden kullanır; ekonomi tick'i henüz çalışmamış save/ilk tur için aynı state fallback'i
uygulanır. Regression: `TestStrategicWarReadinessUsesGoldAndGrainReserves`,
`TestStrategicWarLogisticsGatePreservesOpeningTempo`.

Aktif savunma/konsolidasyon savaşlarında hedef artık `EnemyRegions[0]` değildir.
`AIFront.TargetRegionID`, bölgenin ekonomik/stratejik değeri, başkent olması, kuşatma
durumu, hedefteki savunma gücü, dost erişimi ve mevcut expand plan hedefleriyle
deterministik seçilir. Mevcut `expand` objective'lerinin öncelik sırası açılış temposunu
korumak için aynen kullanılır; yeni hedef skoru özellikle planı savaşı kilitleyen
devletlerin fallback saldırı rotasına uygulanır. Recruitment da aynı cephe hedefini
kullanarak ordunun yanlış sınırda toplanmasını önler.

Üretim kompozisyonu da bu bağlamı okur. Mature ana cephede tahkimli hedef varsa
`55/25/20` piyade/süvari/kuşatma oranı, tahkimatsız hedefte `60/25/15` oranı kullanılır;
dost kritik tehditte `75/15/10` savunma oranına dönülür. Ana saldırı dışındaki aktif
cepheler bu override'ı almaz; genel `AIPlanState` oranları geçerli kalır. Böylece
kuşatma birimi doğru cephe için artırılırken tüm devletin rastgele topçu yığması
engellenir. Kaynak: `internal/ai/unit_composition.go`.

### Müttefik ve Vassal Ortak Cephesi

`buildAIFronts()` aynı realm içindeki overlord/vassal ordularını, ayrıca aktif savaşta
olan doğrudan müttefikleri ortak cephe gücüne dahil eder. Bu katkı yalnız ilgili
ordunun kendi düşmanla `war` ilişkisi varsa geçerlidir; barıştaki veya savaşa
katılmamış müttefik ordu yapay biçimde cephe gücü sayılmaz. Ortak katılımcılardan
birinin `WarLedger.TargetRegionID` kilidi varsa hedef, diğer katılımcının aynı düşman
cephesinde geçerli hedef listesinde bulunması koşuluyla paylaşılır. Böylece vassal ve
overlord farklı sınır bölgelerine dağılmaz; her fraksiyonun hareket/komuta yetkisi
ise kendi AI turunda kalır. Regression: `TestSameRealmWarFrontSharesTargetAndFriendlyPower`.
Relief hedefleri de aynı kuralla vassal veya savaşa katılmış müttefik bölgesine
genişletilir; kuşatma yapan aynı-realm orduya karşı yardım görevi üretilmez ve hedef
sahibi ile kuşatan arasında gerçek `war` ilişkisi aranır. Böylece müttefik kuşatması
boşta kalmazken barıştaki ortakların orduları yanlışlıkla savaşa çekilmez.

AI'nin denizden tahkimli kıyıya indirerek başlattığı kuşatma `NavalLanding` olarak
işaretlenir; barış kabulünde kara ordusu en yakın yeterli nakliye filosuna geri
yüklenir, nakliye yoksa en yakın kendi kara bölgesine çekilir.

### Geri Çekilme ve Takviye

`internal/ai/retreat.go`, yalnız `1300_ottoman_rise` stratejik context'inde mobil kara
ordusunun mevcut rolünü geçici `retreat` rolüyle ezebilir. Açık arazide iki tetikleyici
vardır:

- birimlerin deneyim dahil tam-can saldırı gücüne göre ağırlıklı mevcut güç oranı
  `%45`in altındadır; tam `%45` geri çekilmez,
- aynı veya komşu bölgelerdeki savaş düşmanı kara gücü ordu gücünün en az `%135`idir.

Recovery anchor, AI'nin kendi kara bölgesidir; kuşatma altında veya yabancı ordu
barındıran, komşusunda savaş düşmanı bulunan ve ordunun varışından sonra ikmal
kapasitesi aşılacak bölgeler aday olmaz. Aday puanı ikmal boşluğu, aynı bölgedeki
dost güç, tahkimat, başkent ve gerçek toparlanma hızını birlikte içerir; hız
`2 + 2 × (Çiftlik + Ambar seviyesi)` ile tur çözümlemesindeki ortak helper'dan gelir.
Kısa rota maliyeti puandan düşülür; ancak ağır yıpranmış ordu, bir adım daha uzaktaki
yüksek çiftlik/ambar seviyeli güvenli ikmal bölgesini seçebilir. AI kapasite tahmini
ambarın yerel stok önceliğini de hesaba katar. Rota yalnız AI'nin kendi, kuşatılmamış
ve yabancı kara ordusu bulunmayan transit bölgelerinden geçer. Ordu anchor'a ulaşınca
hareket etmez; mevcut konsolidasyon ve tur sonu takviye/iyileşme kuralları burada
çalışır. Güvenli anchor veya dost rota yoksa mevcut görev korunur.

Aktif kuşatmanın terk edilmesi daha sıkıdır. Kuşatan ordunun bölgesel ikmal aşımı veya
`OverCapacityTurns` kaydı bulunmalı ve kuşatma hedefinin komşularındaki savaş düşmanı
yardım gücü kuşatan gücün `%150`sini **aşmalıdır**; tam `%150` kuşatmayı bozmaz. Her iki
koşul ile güvenli recovery hattı birlikte yoksa siege rolü korunur. Geri çekilme
başladığında aynı fraksiyondan hedef bölgede kalan ilk uygun ordu kuşatmayı deterministik
devralır; yoksa kuşatma kaldırılır. `TurnStepper`, bu devri/kaldırmayı normal hareketten
önce görünür bir step olarak üretir. AI'nin başlattığı yeni kuşatmalar da mevcut
`AttackerHomeRegionID` alanını doldurur.

### Fetih Sonrası Güvenlik

`internal/ai/security.go`, yeni bir işgal süresi veya save alanı eklemeden mevcut bölge
state'ini kullanır. Surları (`BuildingLevel("walls") > 0`) veya aynı devlete ait sabit
`IsGarrison` ordusu bulunan bölge mevcut isyan motorunda zaten korunduğu için mobil
security görevi üretmez. Diğer AI-owned kara bölgelerinde:

- bölge dini devlet diniyle aynıysa `Satisfaction < 35`,
- din farklıysa `Satisfaction < 45`

security ihtiyacı oluşturur. Önce gerçek `Region.IsRebellionRisk()` bölgeleri, sonra
daha düşük memnuniyet, din farkı, stratejik değer ve region ID sırasıyla ele alınır.
Her hedef için aktif kuşatma, relief, retreat veya kritik düşman cephesi savunmasında
olmayan en küçük güçlü mobil kara ordusu seçilir; güç eşitliğinde mesafe ve ArmyID
deterministik bağ kırıcıdır.

Gerçek isyan riski (`Satisfaction < 30`) aynı tur çözüleceği için seçilen ordunun mevcut
hareket puanıyla hedefe ulaşabilmesi gerekir. Önleyici `%30–34` veya din farkında
`%30–44` bandında hazırlık birden fazla tura yayılabilir. Tek saha ordusu yalnız `<30`
acil eşiğinde security rolüne alınır. Security ordusu hedefe yalnız dost, kuşatılmamış ve
yabancı kara ordusu bulunmayan bölgelerden gider; anchor'a ulaşınca eşik düzelene kadar
ayrılmaz.

Security için stratejik rezerv kullanılırsa `ReserveAssignedPower` aynı güç kadar
azaltılır; AI eksilmiş rezervle yeni savaş açamaz. Kritik cephe defense, aktif siege ve
relief görevleri security tarafından bozulmaz. `applyRetreatAssignments()` daha sonra
çalıştığı için ağır yıpranmış/ezilen security ordusunun hayatta kalma geri çekilmesi
önceliklidir. Memnuniyet `35/45` eşiğine ulaştığında rol sonraki AI turunda otomatik
bırakılır; rol save'e yazılmaz.

### Ağırlıklı Kara Rotaları

`internal/ai/pathfinding.go`, yalnız `1300_ottoman_rise` stratejik kara AI'sında mevcut
uzun menzilli BFS seçimini deterministik Dijkstra maliyet alanıyla değiştirir. Aynı motor
cephe rolü mesafeleri, rally, reserve, relief, retreat ve security hedeflerine giden ilk
adımı da üretir. Gerçek hareket uygulaması hâlâ komşu bölge başına mevcut hareket puanı
kuralını kullanır; aşağıdaki değerler yalnız AI rota tercih ağırlıklarıdır:

- arazi: `TerrainData.MoveCost × 10` (`plain/coast=10`, `forest=20`, `pass=30`),
- aynı-realm/müttefik transit: `+5`,
- sahipsiz veya savaş düşmanı terminal hedef: `+10`,
- girilebilir aktif kuşatma terminali: `+20`,
- yerel savaş düşmanı gücü orduyla eşitse `+40`, oranla ölçekli ve en fazla `+80`,
- öngörülen ikmal açığının her birimi `+6`, en fazla `+60`.

Kilitli kara, dağ ve deniz geçilemez. Genel rota kendi, aynı-realm ve müttefik toprağını
transit kullanabilir; sahipsiz veya savaş düşmanı bölgesinde ilerleme sonlanır, barış/
ticaret ilişkisindeki üçüncü taraf toprağı kullanılamaz. `retreat/security` politikası
yalnız AI'nin kendi, kuşatılmamış ve yabancı kara ordusu bulunmayan bölgelerini kabul
eder. Uzun menzilli arama `PathSearchDepth` değerini hop sınırı olarak korur.

Hesaplanan alanlar `StrategicContext` içinde ordu, başlangıç bölgesi, politika ve hop
sınırı anahtarıyla yalnız o AI turunda cache'lenir. Eşit sonuçlar sırasıyla toplam
maliyet, hop sayısı, region ID ve ilk adımla çözülür. Save alanı eklenmez. Diğer
senaryolar geriye dönük BFS yolunu kullanır.

### Koordineli Rally Hazırlığı

Bir genişleme planında rezerv/defense/relief dışında en az iki `assault` veya `siege`
ordusu varsa `internal/ai/rally.go` koordineli hazırlık başlatır. Rally bölgesi:

- aktif savaş hedefi varsa o devletle, yoksa plan hedefiyle doğrudan sınır paylaşır,
- AI'nin kendi kara toprağıdır,
- kuşatma altında değildir ve içinde yabancı ordu bulunmaz,
- ikmal boşluğu, yerel dost/düşman gücü, tahkimat, başkent ve stratejik değerle
  deterministik puanlanır.

Seçilen `RallyRegionID` ile `RallyDeadlineTurn`, `AIPlanState` içinde compact ve legacy
save hatlarına yazılır. Böylece kayıt yüklemek üç turluk hazırlığı baştan başlatmaz.
Rally aktifken hücum/kuşatma rolleri yalnız rally noktasına yaklaşan yasal dost bölge
hamlelerini seçer; noktaya ulaşan ordu hazırlık tamamlanana kadar sınırı geçmez.

Hazırlığın erken tamamlanması için en az iki ordunun rally bölgesinde bulunması ve
toplanan gücün şu iki eşiğin büyüğünü karşılaması gerekir:

- atanmış toplam hücum gücünün `%60`ı,
- hedefin frontier gücünün zorluk seviyesindeki asgari saldırı oranı (`%130/%115/%100`).

Eşik karşılanınca deadline mevcut tura çekilir ve ordular objective hedefine serbest
bırakılır. Eşik karşılanmasa bile üçüncü turda bekleme sona erer; standart toplam güç,
frontier ve savaş riski kontrolleri yine korunur. Aktif rally varken yeni proaktif savaş
ilan edilmez. Tek hücum ordusu bulunan küçük devlet rally yüzünden bekletilmez. Rally
bölgesi kuşatılır veya hedefle sınırı kaybederse kalıcı kayıt iptal edilip uygun yeni
bir nokta aranır.

Retreat sonrası tempo ölçümünde Normal fast 12x2 `9.37 sn` ve Osmanlı `2 → 3`; Normal
medium 42x4 `63.65 sn`, Osmanlı ortalama `2 → 5.8`, güç `664`; Zor fast 12x2 `7.03 sn`
ve Osmanlı `2 → 3` sonucu verdi. Rally tabanına göre orta profildeki fark yalnız `-0.2`
bölge ve `-2` güçtür; Memlük büyümesi `+5.0 → +4.8` olmuştur.

Security sonrası ölçümde Normal fast 12x2 `9.38 sn`, Normal medium 42x4 `64.17 sn`,
Zor fast 12x2 `7.20 sn` sürdü. Osmanlı sonuçları sırasıyla `2 → 3`, ortalama
`2 → 5.8`/güç `664` ve `2 → 3` olarak retreat tabanıyla aynı kaldı.

Ağırlıklı Dijkstra sonrası Normal fast 12x2 `9.53 sn`, Normal medium 42x4 `66.26 sn`,
Zor fast 12x2 `7.41 sn` sürdü. Osmanlı sırasıyla `2 → 3`/güç `272`, ortalama
`2 → 5.8`/güç `665` ve `2 → 3`/güç `284` sonucunu verdi; orta profil bölge kazanımı
korunurken rota maliyetinin süre etkisi yaklaşık `%3` kaldı.

### AI Savaş Sonrası Düzen

AI'nin `allow_vassalization` işaretli genişleme objective'i hibrit sonuç kullanır. Son
kara toprağında yenilen hedef:

- dış müttefiki yoksa,
- saldıran belirgin askeri üstünlüğe sahipse,
- hedef agresifliği direnç eşiğinin altındaysa,
- fetih anındaki vassallık zarı başarılıysa

`diplomacy.ForceVassalizeAfterWar()` ile vassal bırakılır ve yerel bölge sahipliği
korunur. Zar başarısızsa veya uygunluk kontrollerinden biri geçmezse doğrudan
fetih/ilhak akışı sürer. Bu karar açık arazi savaşı, savaşsız işgal, çıkarma,
genel hücum ve kuşatma teslimi çıkışlarında aynı politika helper'ından geçer. Hedef
bölgede ardıl fraksiyon hâlâ aktifse `CanRestoreSuccessorAtRegion()` false olur ve
AI vassallık kararı uygulanmadan doğrudan ilhak akışı korunur.

`TakeTurn` sırasıyla şu adımları yapar:
1. Zorluk 3 ise → `FormCoalitionAgainstPlayer()`
2. Diplomasi taraması ve fırsatçı savaş değerlendirmesi → `aiHandleDiplomacy()`
3. Teknoloji araştırma → `aiResearch()`
4. Ekonomik bina inşası → `aiEconomyBuild()`
5. Deniz stratejisi → `aiNavalStrategy()`
6. Birim alımı + kışla inşası → `aiRecruitAndBuild()`
7. Ordu konsolidasyonu → `aiConsolidateArmies()`
8. Komutan havuzu üretimi ve boşta kalan saha ordularına atama → `EnsureFactionCommanders()`
9. Ordu hareketi → `moveArmy()` (her ordu için)

`TurnStepper` aynı zinciri iki faza ayırır:

1. `runTurnPrelude()` — diplomasi, araştırma, ekonomi, deniz, recruit/build, konsolidasyon ve komutan ataması
2. Hareket fazı — her `Step()` çağrısı tek bir `executeMove()` sonucu döndürür

Bu step sonucu `TurnStep{Kind, FocusRegion, Message}` biçimindedir. Oyun katmanı:

- oyuncuya yakın (`<= 3` komşuluk derinliği) aksiyonlarda kamerayı `FocusRegion` üzerine taşır,
- uzak aksiyonlarda yalnız AI overlay'i günceller,
- diplomasi/savaş/fetih gibi önemli adımları event log'a da yazar,
- oyuncuya bekleyen diplomasi teklifi varsa step çözmeyi tamamen durdurur,
- teklif kabul edilirse aktif AI fraksiyonunun kalan turunu kapatır; böylece aynı AI o tur içinde barıştan sonra yeni saldırı veya ileri savaş hamlesi üretmez. Aynı kural, AI savaş ilanı sonrası oyuncuya düşen müttefik savaş çağrısı kabulünde de deklaratör AI için geçerlidir.

---

## Ordu Hareketi Mantığı

`moveArmy()` / `TurnStepper.Step()` → `chooseBestMove()` → `scoreMove()` → `executeMove()`

`scoreMove()` hedef bölge için puan hesaplar:

| Koşul | Puan |
|---|---|
| Kendi bölgesi | 0 (hareket etme) |
| Barış/ittifak/ticaret halindeki bölge | -1 (atla) |
| Savaş halinde + üstün güç | 95 |
| Daha güçlü düşman var | -1 |
| Sahipsiz bölge (kapasite doluysa) | 70 |
| Sahipsiz bölge | 50 |
| Düşman bölgesi, kapasite dolu + savaş | 100 |
| Düşman bölgesi, savaş | 90 |

`atCapacity` — `DeployedLandUnits >= ManpowerCap` ise fetih yaparak kapasite genişletme önceliklenir.

Tahkimli hedefte başka bir ordunun aktif kuşatması varsa AI yeni kuşatma açmaz; ancak bölgeye giriş hakkı olan ve bölgedeki besieger düşman ordusunu savaşta yenebilen hedefler `scoreMove()` tarafından savaş adayı olarak puanlanır. Böylece AI, oyuncuyla aynı kuşatma kuralına tabi olur: önce kuşatma, istisnai olarak da kuşatma yapan orduyu savaşla kaldırma.

Kuşatılan bölgede duran bölge sahibi veya müttefik AI ordusu da normal hareketten önce huruç savaşı yapar. `executeMove()` bu savaşı kaynak bölgenin arazisiyle çözer; zaferde kuşatmayı kaldırıp dost/sahipsiz hedefe çıkar, yenilgide mevcut kayıplarıyla aynı bölgede bırakır ve hareketini bitirir.

### Lojistik Farkındalığı

AI artık dost kara bölgelerini sadece diplomasi/savaş açısından değil, ikmal baskısı açısından da puanlar.

AI hareket executor'ü de konum değişiminde `GameState.ClearArmyLogisticsAfterRelocation()` çağırır; eski bölgenin `!` uyarısı ve kara ordusunun bölgeye özgü aşım sayacı yeni konuma taşınmaz.

- Kaynak bölge aşırı doluysa (`grain_upkeep` toplamı bölgenin efektif tahıl + yerleşim tamponu + sınırlı stok desteğini aşıyorsa) `scoreMove()` dost komşular arasında baskıyı azaltan bölgeye pozitif puan verir.
- Aynı canlı kapasite modeli ambarın bölgeye ayırdığı stok desteğini içerir; böylece AI, oyuncunun gördüğü ikmal aşımından kaçınır ve iyileşmek için yalnız güvenli değil, çiftlik/ambar seviyesi yüksek bölgeleri de tercih eder.
- Komşu dost bölgeye geçiş sonrası aşım sıfırlanıyorsa ek bonus verilir; baskıyı daha kötü yapacak dost hedefler cezalandırılır.
- Böylece AI küçük ve tahılı zayıf bölgede yığılmış orduları savaş yokken bile daha rahat komşu bölgelere dağıtabilir.
- Aynı lojistik hesabı `aiConsolidateArmies()` ve hareket sonrası `tryMergeAIArmies()` için de kullanılır; aşırı dolu kara bölgede AI artık körlemesine ordu birleştirme yapmaz.

AI de oyuncu ile aynı `combat.ResolveBattleWithMods()` kullanır.

### AI Komutanları

`GameState.EnsureFactionCommanders(ownerID)`, AI tur öncesi üretim ve konsolidasyon
tamamlandıktan sonra çalışır. Her aktif saha ordusu için (garnizonlar hariç, deniz
filoları dahil) havuzda bir komutan bulunmasını sağlar ve komutansız ordulara ID
sırasına göre atama yapar. Mevcut komutan kariyerleri korunur; birleşme veya save/load
sonrasında `SyncCommanderLinks()` aynı komutanın iki orduya bağlanmasını engeller.
AI nakliye akışında kara komutanı filoda taşınır; çıkarma savaşında geçici kara
ordusuna uygulanır ve başarılı karaya çıkışta yeni orduya aktarılır.

`executeMove()` artık stepper için açıklamalı sonuç döndürür:

- `TurnStepMove`
- `TurnStepBattle`
- `TurnStepEmbark`
- `TurnStepDisembark`
- `TurnStepConquest`

Bu sayede hareket state'i gerçek zamanda akarken aynı anda UI mesajı ve yakınlık filtresi üretilebilir.

---

## Koalisyon Mantığı

`FormCoalitionAgainstPlayer()` — zorluk 3'te her AI turunun başında çalışır.

**Tetikleme koşulu:** Oyuncunun bölge sayısı `coalitionThreshold = 8`'i geçmesi.

**Etki:** AI fraksiyon oyuncuya savaş açar ve aynı diplomasi motoru üzerinden diğer AI'larla ittifak kurmaya çalışır.

→ İttifak mekanizması: [[systems/diplomacy]]

---

## Diplomasi Safhası

`aiHandleDiplomacy()` her AI turunda ilişkileri tarar. Karar döngüsü
`internal/ai/diplomacy.go` içinde tutulur; map tabanlı faction, region ve
army sıralama yardımcıları `internal/ai/ordering.go` içinde tutulur; böylece AI karar
döngüsü deterministik ID sırasını tek bir modülden kullanır.
Fırsat savaşı aday taraması ve hedef puanlaması `internal/ai/war_strategy.go` içindedir;
hareket ve çarpışma uygulaması bu karar modülünden ayrı kalır.
Üretim/recruitment orkestrasyonu `internal/ai/recruitment_strategy.go` içindedir;
kışla yatırımı ve manpower kuyruğu, birim seçimi ile bölge skorlamasından ayrıdır.
Araştırma ve ekonomi wrapper katmanı `internal/ai/economy_research.go` içindedir;
teknoloji seçimi `research_strategy.go`, bina ROI değerlendirmesi ise
`building_investment.go` tarafından yapılır.
Deniz stratejisi giriş katmanı `internal/ai/naval_strategy.go` içindedir; 1300 görev ve
merchant üretimi `naval_mission.go` ile `merchant_trade.go` modüllerine devredilir.
Kuşatma state ve başlatma yardımcıları `internal/ai/siege_strategy.go` içindedir;
uzun menzilli hareket hedefleme ve rota seçimi `internal/ai/movement_strategy.go`
içindedir; komşu hareket skoru ve combat resolve bu karar katmanlarından ayrıdır.
`1300_ottoman_rise` için ağırlıklı kara rotası, diğer senaryolar için geriye dönük BFS
seçimi kullanılır. Her iki yol da stratejik rol bonusunu ve deterministik bölge sırasını
korur. Komşu hedef seçimi, lojistik context önbelleği, denizden karaya çıkış puanı ve
`scoreMove` kararları da aynı dosyada tutulur; gerçek hareket puanı tüketimi, savaş ve
state değişikliği ortak `ai.go` akışında kalır.
Deniz filoları da aynı `ArmyAssignments` modelinde `transport` veya `escort` rolü alır;
aktif görev varsa anchor çıkış/iniş denizinden türetilir, merchant filoları kara görev
rollerine karıştırılmaz.

1300'de ilişki skoru `40` altındaki barış çiftleri, aynı dinin varsayılan `25` puanı
ittifak için yeterli olmadığı için pahalı stratejik ittifak değerlendirmesine girmeden
elenir. Diğer senaryolarda genel `25` eşiği korunur. AI ayrıca mevcut müttefiki hedefle
savaş halinde olan bir devlete teklif göndermez; aynı savaş çakışması doğrudan diplomasi
geçidinde ve kuyruk çözümünde de yeniden doğrulanır.

- `war` ilişkisinde 1300 senaryosu taraf çifti başına kalıcı `WarLedger` okur. Objective
  tamamlanması, fethedilen/kaybedilen bölgeler, muharebe ve kuşatma kayıpları, savaş
  süresi ve durgunluk, askeri güç oranı, altın/tahıl stresi, eşzamanlı savaş sayısı ve
  başkent tehdidi ortak barış baskısı üretir. İlk üç savaş turunda normal teklif açılmaz;
  başkent tehdidi veya askerî çöküş acil istisnadır. Aynı savaşta reddedilen teklif üç
  tur cooldown uygular. Diğer senaryolar mevcut legacy güç/bölge/skor kararını korur
- `peace` ilişkisinde ittifak için artık sadece skor yetmez; 1300'de skorun `40` olması, kara sınırı, aktif ticaret, ortak düşman veya ortak büyük tehditten en az biriyle gerçek coğrafi/stratejik bağ aranır. AI ayrıca küçük devletlerde düşük müttefik tavanı uygular, `ai_expansion_targets` ile çakışan hedeflere ortak tehdit yoksa ittifak teklif etmez, mevcut müttefikin savaş düşmanına teklif göndermez ve büyük gücün tek bölgeli/zayıf devlete attığı alliance teklifi için ayrıca gerçek askeri veya stratejik fayda arar
- `peace` ilişkisinde skor yeterliyse ve bağlanabilir kara/deniz hattı varsa ticaret dener; salt uzak ve nötr devletlere sırf kapasite var diye trade açmaz
- Somut ticaret çıkarı (aktif veya bağlanabilir hat), anlamlı ittifak faydası, ortak düşman/tehdit ya da sınır gerilimi olan ve genişleme hedefi olmayan barışçıl hedeflerde ilişkiyi ücretli olarak onarır. Skor `15` altındaysa altın rezervini koruyarak `Heyet` (`40` altın, `+8`) gönderir; aktif ticaret veya stratejik ittifak için daha yüksek skor gerekiyorsa `Hediye` (`120` altın, `+15`, alıcıya `80` altın) kullanır. Her aksiyon tur içi diplomasi kotasını tüketir.
- Bu aksiyonlar AI-AI arasında doğrudan çözülür. Hedef oyuncuysa barış/ittifak teklifleriyle aynı `DiplomaticOffers` kuyruğuna girer; oyuncuya yalnız `Tamam` düğmesi olan bildirim modalı gösterilir ve onaylanana kadar ilişki ile altın state'i değişmez (`internal/ai/diplomacy.go`, `internal/render/renderer_dialogs.go`).
- vassal durumundaki AI bağımsız diplomasi açmaz; overlord'u olmayan devletler ise başka bir overlord'a bağlı hedeflerle doğrudan müzakere başlatmaz
- `allied` ilişkide ortak tehdit kalmamış, ticaret/sınır bağı yok olmuş, tarihsel genişleme hedefiyle çatışan ya da büyük güç için artık anlamlı katkı üretmeyen zayıf ittifaklar AI tarafından iptal edilebilir
- normal/zor zorlukta, ticaret/ittifak ilişkisini bozmadan, sınır komşusu zayıf bir hedefe karşı güç ve cephe üstünlüğü varsa tek bir fırsat savaşı açabilir
- bekleyen barış, ittifak ve ticaret teklifleri teknoloji farkı ve uzun vadeli tehdit baskısına göre önceliklenir; oyuncuya daha baskı altındaki teklif önce gösterilir ve prompt içinde tür ile kısa sebep bilgisi görünür. İttifak teklifinde AI ayrıca turn + taraf kimliğine bağlı deterministik hafif rastgelelik kullanır; böylece aynı uygun çerçevede her tur mekanik olarak sabit spam yerine bazen teklif açar, ama yüksek olasılıklı ortak tehdit senaryoları yine güvenilir biçimde görünür. Oyuncu bir teklifi reddettiğinde aktör-hedef-aksiyon bazında üç tur cooldown uygulanır; ardından aynı zar mekanizması %35 tekrar deneme şansı verir. Ret ayrıca ilişkiyi `-3` düşürür. Desteksiz dış ittifakların relation skoru ise artık her tur otomatik yükselmez; destek yoksa yavaşça aşınır.
- Aynı oyuncuyla savaşta barış ve kuşatma anlaşması koşulları birlikte oluşursa önce barış teklifini kuyruğa alır; barış kabul edilirse geçersiz teslimiyet veya kuşatma vassallığı teklifi diplomasi kuyruğundan temizlenir, ret edilirse kuşatma teklifi sonraki karar olarak kalır.
- Aktif kuşatmalarda AI, oyuncu hedefliyse iki yönlü `propose_surrender` teklifi değerlendirebilir: kuşatan AI yeterli gedik/baskı ve güç üstünlüğünde oyuncudan teslim ister; kuşatılan AI ağır baskıdaysa oyuncuya teslim olmayı teklif eder. Kuşatılanın son kara toprağı varsa aynı eşik `propose_siege_vassalization` üretir; kabulde bölge sahibi değişmez, devlet kuşatanın vassalı olur, savaş ve kuşatma biter. Teklif `RegionID` ile aktif kuşatmaya bağlanır, normal diplomasi elçi kotasını tüketmez ve ret cooldown'u bölge bazında tutulur. Normal teslimiyet kabulünde savunmacı ordu geri çekilir.

Fırsatçı savaş kararı `aiEvaluateWarOpportunities()` ile sınırlanır:

- Kolay zorlukta çalışmaz.
- AI mevcut savaş limitini doldurduysa yeni savaş açmaz.
- Yalnız kara sınırı paylaştığı ve `peace` durumundaki hedefleri değerlendirir.
- Başka bir overlord'a bağlı hedefleri doğrudan savaş adayı yapmaz; aktif savaş sayısı da artık realm root bazında tutulur, böylece bir overlord'un vassal savaş kayıtları aynı savaşı yapay olarak çoğaltmaz.
- Aktif `trade` veya `allied` ilişkiyi savaş için bozmaz.
- `factions.json` içindeki `ai_expansion_targets` listesinde bulunan tarihsel hedefler daha geniş ilişki eşiğiyle aday olur ve skor bonusu alır.
- Hedef listesi olan AI fraksiyonları ilk turlarda savaş cadence beklemez; sonraki turlarda da cadence aralığı daha kısadır.
- Askeri güç, cephe gücü, ilişki skoru, hedef bölge değeri, bölge sayısı, din farkı, AI agresifliği ve oyuncu hedef cezası birlikte puanlanır.
- En iyi hedef eşik üstündeyse standart `diplomacy.Execute(..., ActionDeclareWar)` yolu kullanılır; hareket/fetih yine `scoreMove()` ve `executeMove()` zincirinden geçer.

`ai_expansion_targets` rastgele saldırganlık değildir: hedef olmayan fraksiyonlarda eski negatif ilişki kapısı korunur, ticaret/ittifak ilişkileri bozulmaz ve güçsüz AI daha güçlü hedefe savaş açmaz. 1300 senaryosunda Osmanlı için Doğu Roma, Germiyan, Karesi ve Ahiler; Reconquista, Yüz Yıl Savaşları ve doğu bozkır cepheleri için tarihsel hedefler profil objective'leriyle tanımlıdır. Yüz Yıl Savaşı hedefleri ayrıca 1337 event bayrağı açılana kadar genel fırsat savaşı listesinden çıkarılmıştır.

AI ve oyuncu aynı `internal/diplomacy` motorunu kullandığı için:

- kabul/red kuralları tutarlıdır
- ticaret rotaları aynı şekilde açılıp kapanır
- AI ticaret yaptığı veya müttefik olduğu hedefe saldırmaz

1300'de AI-AI barışı tek taraflı uygulanmaz: teklif sahibi önce kendi barış baskısı
eşiğini, hedef AI ise aynı değerlendirmeyi karşı perspektiften geçmelidir. Oyuncu
hedefse teklif `DiplomaticOffers` kuyruğuna girer ve oyuncu yanıtına kadar savaş sürer.
Barış kabul edildiğinde ilgili AI objective'i aynı tur yeniden değerlendirmeye açılır ve
kalıcı rally bölgesi/deadline temizlenir.

### 1300 Stratejik İttifak Değerlendirmesi

`internal/diplomacy/alliance_strategy.go`, ittifak faydasını teklif sahibi ve hedef AI
için ayrı perspektiften hesaplar. `AssessAllianceProposal()` hedefin kabul değerini,
`aiShouldAttemptAllianceOffer()` teklif sahibinin girişim değerini okur:

- ortak düşman `+20`, ortak büyük tehdit `+18` stratejik tehdit değeri üretir;
- adayın actor'u tehditten ayıran gerçek sınır konumu tampon değeridir;
- aday ordusunun bu tehdit sınırında bulunan gerçek gücü cephe desteği sayılır;
- aktif/bağlanabilir ticaret hattı, trade capacity, aday güç ve bölge katkısı değer ekler;
- statik `AIExpansionTargets` çakışması `-18` yumuşak cezadır ve ortak tehditçe aşılabilir;
- iki taraftan birinin aktif `AIPlanState.TargetFactionID` alanı diğerini gösteriyorsa
  ittifak kesin engellenir.

Aktif objective sonradan mevcut müttefiki hedeflerse AI ittifakı bitirir; korunmuş trade
stance'i hedef savaşı kilitlemesin diye aynı çiftin ticaret anlaşmasını da kapatır. Statik
gelecek hedefi tek başına bu sert temizliği yapmaz. Tehdit/fayda kaybolduğunda retention
eşiğinin altındaki veya müttefik tavanını aşan düşük değerli ittifaklar çözülür. Bu model
yalnız `1300_ottoman_rise` için aktiftir.
- AI savaş ilanında hem saldıran hem savunan taraftaki oyuncu müttefikleri otomatik çekilmez; önce oyuncuya savaş çağrısı modalı düşer

---

## Teknoloji Modları

`aiTechMods()` — AI de oyuncuyla aynı teknoloji bonuslarını hesaplar:

```go
fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
return TechMods{
    AttackMod:  fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod,
    DefenseMod: fx.LandDefenseMod,
}
```

AI savaşları da oyuncu ile aynı `Army.Commander` modifiyerlerini kullanır. Savaş
sonucunda XP ve zafer/yenilgi sayısı gerçek AI ordusunun komutanına yazılır; birleşik
savunmada sanal savunma stack'i yerine kaynak orduların komutanları ilerletilir.

---

## Zorluk Seviyeleri

1300 senaryosunda zorluk, `ai_strategies.json.difficulty_policy` üzerinden karar
kalitesi ve risk iştahıyla ayrışır:

| Seviye | Plan ufku | Hedef bölge | Yol arama | Plan hareket etkisi | Savaş politikası |
|---|---:|---:|---:|---:|---|
| 1 (Kolay) | 4 tur | 3 | 5 derinlik | `%70` | Proaktif savaş yok; en az `%130` güç; en fazla 1 savaş |
| 2 (Normal) | 6 tur | 4 | 8 derinlik | `%100` | Eşik 70; en az `%115` güç; 10 tur taban kadans; en fazla 1 savaş |
| 3 (Zor) | 9 tur | 5 | 12 derinlik | `%125` | Eşik 65; denk güç yeterli; 7 tur taban kadans; en fazla 2 savaş |

Zor seviye koalisyon mantığını korur ve oyuncuya karşı hedef skoruna yalnız `+4`
ekler. Büyük ekonomi/hareket hilesi yoktur: tüm seviyelerde AI ve oyuncu aynı hareket
hesabını kullanır. Zor AI yalnız yeni oyun başlangıcında küçük `+80 altın/+30 tahıl`
tamponı alır; Normal ve Kolay kaynak bonusu almaz. Bu politika yalnız 1300 senaryosuna
aittir. `difficulty_policy` bulunmayan eski senaryoların mevcut `+1` zor AI hareketi ve
`+300/+100` başlangıç davranışı geriye uyumluluk için korunur.

Rally katmanı sonrasında Normal fast 12x2 `9.21 sn` ve Osmanlı `2 → 3`, Normal medium
42x4 `63.18 sn` ve Osmanlı ortalama `2 → 6` sonucunu verdi. Zor fast 12x2 `7.25 sn`
sürdü ve Osmanlı `2 → 3` bölgeye ulaştı. Çok ordulu Memlük'ün 42 turluk ortalama
genişlemesi `+8.5`ten `+5`e indi. Bu sayılar tarihsel sonuç şartı değil, güvenlik ile
genişleme temposu arasındaki sonraki kalibrasyonların referansıdır.

---

## Teknoloji Araştırma (`aiResearch`)

Aktif araştırma yoksa başlatır; devam eden araştırmayı yarıda kesmez. 1300 senaryosunda
seçim yukarıdaki **Plan ve Darboğaz Bazlı Araştırma** modeliyle yapılır. Aşağıdaki sabit
puanlar yalnız diğer senaryoların legacy davranışıdır:

| Kategori | Puan | Ek bonus |
|---|---|---|
| `military` | 100 | Saldırı efektleri varsa +20 |
| `economy` | 70 | `gold_per_region` varsa +15 |
| `naval` | 50 | — |
| `diplomacy` | 40 | — |
| `religion` | 30 | — |

Kısa süreli teknolojilere `TurnsRequired / 2` azaltma uygulanır.

---

## Ekonomik Bina (`aiEconomyBuild`)

Her tur en fazla bir bina üretim emri açar. Bina anında `Region.Buildings` listesine
eklenmez; `GameState.ProductionQueue` içine `kind=building` emri yazılır ve
`applyProductionTicks()` sırasında tamamlanır. 1300 senaryosunda seçim yukarıdaki
**Bina Yatırım Puanlaması** ile yapılır. Diğer senaryolar geriye uyumluluk için ilk
uygun adayı `çiftlik → pazar → sur` sırasıyla seçen legacy akışı korur.

Her iki akış da yalnız altını değil, bina reçetesindeki `grain/iron/timber/stone`
maliyetlerini kontrol eder. Kurulu seviye + pending seviye `max_per_region` sınırını
aşamaz; inşa süresi mevcut ve kuyruktaki seviyeler arttıkça uzar.

---

## Deniz Stratejisi (`aiNavalStrategy`)

`1300_ottoman_rise`, `internal/ai/naval_mission.go` içindeki runtime-only görev modelini
kullanır. Kıyı sahibi olmak tek başına liman veya transport yatırımı açmaz:

1. Aktif `expand` objective'inin hedef bölgesi ya da aynı hedef devletin kıyısı ve
   `defend` objective'indeki dost kıyı adayları çıkarılır.
2. Uygun bir saha ordusunun hedefe güvenli kara rotası varsa deniz görevi kurulmaz.
3. Kara yolu yoksa taşınabilir, kuşatma/sabit güvenlik/geri çekilme görevi olmayan bir
   ordu; ulaşabildiği dost çıkış kıyısı ve deniz BFS'iyle erişilen hedef kıyı seçilir.
4. Seçilen ordunun birim sayısından, çıkış denizine ulaşabilen mevcut filoların boş
   `TransportCapacity` değeri ve aynı hatta kuyruktaki transport kapasitesi düşülür.
   Yalnız kalan açık kadar gemi siparişi verilir; yeterli boş kapasite varken yeni
   transport açılmaz.
5. Gerekli port seviyesi yoksa liman yalnız seçilen çıkış kıyısında kurulur. Escort
   değerlendirmesi de ancak bu somut görev varken ve aynı çıkış hattında çalışır.
6. Üretim tamamlandığında mevcut `completeNavalUnit()` akışı, sipariş limanının ilk
   komşu denizinde filo oluşturur veya mevcut filoya birim ekler. Görev hesabı da aynı
   çıkış denizini kullandığı için pending kapasite yanlış hatta sayılmaz.

### 1300 Deniz Tehdit Haritası ve Güvenlik Eşiği

`internal/ai/naval_threat.go`, her stratejik context üretiminde savaş halindeki düşman
filolarını deniz bölgesi bazında toplar. Güç, muharebe motoruyla aynı tabandan gelir:
`Army.TotalStrength()` üzerine ilgili tarafın `NavalAttackMod`/`NavalDefenseMod`, komutan
saldırı/savunma ve moral etkileri uygulanır. `StrategicContext.NavalThreats` düşman ve
dost gücünü; `ThreatenedPortIDs` ise kendi denizi veya bir komşu denizde savaş düşmanı
filo bulunan limanları taşır. Bunlar runtime-only'dir ve save'e yazılmaz.

Görev rotası klasik en kısa BFS değildir. Deterministik rota etiketi şu sırayla
karşılaştırılır:

1. rota üzerindeki en yüksek düşman filo gücü,
2. rotadaki toplam düşman filo gücü,
3. deniz adımı sayısı,
4. ilk adım ve bölge ID'si.

Bu nedenle tehditli kısa koridor yerine daha uzun ama tehdit taşımayan rota seçilir.
Güvenli alternatif yoksa yüklü transport veya öncü savaş filosu, gireceği denizdeki
düşman efektif savunma gücünün en az `%110`u kadar efektif saldırı gücüne sahip olana
kadar bekler. Aynı bölgede birleşebilen savaş gemileri tur öncesi mevcut filo
konsolidasyonu ile gerçek stack'e katılır; ayrı escort filoları görev hattına yönelip
tehdidi önce temizler.

Görev rotasının maksimum tehdidi veya çıkış limanına bir adım mesafedeki deniz tehdidi
pozitifse escort ihtiyacı da aynı `%110` hedefinden hesaplanır. Erişilebilir mevcut
filolar ve aynı deniz ağındaki pending gemiler projected güce katılır. Açık varsa görev
limanı önce gerekli port seviyesine çıkarılır, ardından yalnız güç açığını kapatacak
sayıda `warship` emri kuyruğa yazılır.

Tehdit haritası her rota kenarında filo taramaz: stratejik planlama sırasında context
başına bir kez cache'lenir. Filo hareketi state'i değiştirebildiği için hareket rotası
güncel tehdit map'ini çağrı başına bir kez yeniden kurar.

Diğer senaryolar legacy davranışı korur: kıyı/savaş durumuna göre `1–3` filo limiti,
`aiSeaPressure()` ile liman seçimi ve transport bulunan baskılı hatlarda escort üretimi.

### 1300 Savaş Filo Devriyesi ve Liman Çıkışı

Savaş gemisi somut bir çıkarma görevi taşımıyorsa artık boşta kalmaz. `Patrol` rolü
ile düşman filosu görülen denizleri, tehdit altındaki liman yaklaşımlarını ve kendi
aktif ticaret rotalarının denizlerini hedefler. Rota seçiminde tehdit haritası,
güvenlik eşiği ve deterministik BFS kullanılır; filo rastgele yabancı kıyıya
gönderilmez. Somut çıkarma görevi varsa aynı filo `Escort` rolünde kalır.

Üretim veya senaryo yüklemesi filoyu limana bağlı oluşturduğunda deniz `RegionID`si
korunur, fakat ilk AI emri aynı deniz ankrajına yöneltilen özel bir çıkış adımıdır.
Bu adım `executeMove()` içindeki kanonik dock temizliğini çalıştırır; sonraki turda
filo devriye, escort veya ticaret rotası hareketine devam eder. F3 AI teşhis ekranı
filo toplamını, limanda bekleyen sayısını, aktif deniz görevini ve ilgili engelleri
gösterir.

Denizaşırı savaş ilanı da artık yalnızca somut çıkarma görevi hazırsa açılır:
hedef kıyı, çıkış limanı, deniz rotası, transport teknolojisi/kapasitesi ve
taşınacak saha ordusu birlikte doğrulanır. Kara sınırı olmayan hedef için bu kapı
geçilmeden savaş fırsatı puanlanmaz.

---

## 1300 Merchant Ticaret Filosu

`internal/ai/merchant_trade.go`, `internal/state/merchant_trade.go` ve
`internal/economy/economy.go` birlikte merchant gemisini gerçek ticaret rotası
throughput'una bağlar:

- Her merchant gemisi aktif `gönderen->alan` rotasına `+1` hacim ekler; merchant
  bonusunun üst sınırı rota panelinde görünen o rotanın hacmidir. Tek merchant filosu
  bu kapasiteye kadar gemiyi birlikte taşıyabilir. Katkı `TradeRoute.MerchantAmountBonus` olarak
  runtime hesaplanır ve save'e yazılmaz.
- Filo görevi `Army.TradeRouteKey` ile compact/legacy/debug save'lerde korunur.
  Rota yeniden kurulduğunda mal türü veya geçici rota nesnesi değil, yönlü fraksiyon
  anahtarı kullanılır; askıya alınan ya da silinen rota bonus üretmez.
- Rota hacmi en az bir uçta aktif kıyısal trade center varsa ve filo o merkezin
  komşu deniz hücresindeyse uygulanır. İki uçta da merkez varsa merkezler
  `trade_centers.json` link graph'ında bağlı olmalıdır. Kaynak mal veya alıcı altını yetmiyorsa normal ekonomi
  kapısı rotayı atlar; merchant gemisi bedava altın üretmez.
- Merchant filoları savaş/nakliye filolarıyla birleştirilmez. Venedik ve Ceneviz
  aktif maritime rotalarda önce en az kapsanan hattı seçer; gerekli liman seviyesini
  kurar, açık merchant slotlarını üretir ve farklı yönlü rotalara deterministik
  atama yapar.
- Ticaret merkezine yaklaşan düşman filo tehdidi varsa merchant üretiminden önce
  aynı `%110` escort eşiğiyle `warship` açığı kapatılır. Bütçe, ilk liman yükseltmesi
  ve bir merchant gemisinin gerçek hammadde maliyetini ekonomi yatırımlarına karşı
  rezerve eder.

Kaynak kodu: `internal/ai/merchant_trade.go`, `internal/state/merchant_trade.go`,
`internal/economy/economy.go`, `internal/game/production.go`.

Testler: `internal/ai/merchant_trade_test.go`, `internal/state/merchant_trade_test.go`,
`internal/save/save_test.go`, `internal/game/production_naval_test.go`.

## Oyuncu Donanma Görevleri

AI'nin runtime deniz görevlerinden ayrı olarak oyuncu savaş filoları artık kalıcı
`patrol`, `blockade` ve `escort` görevleri alabilir. Saf nakliye filosuna oyuncu
görevi atanmaz; mevcut taşıma/çıkarma hareketi korunur. Savaş filosu
devriye/abluka için deniz bölgesi seçer; escort satırı aynı devlete ait bir
nakliye filosunu izler. Görev seçimi `GÖREV` butonundan yapılır, geçerli hedefler
haritada renkli işaretlerle gösterilir ve görev değiştirilebilir veya
kaldırılabilir.

Atama yalnız state doğrulamasından sonra kaydedilir. Görev seçim satırları
görevin ekonomik/askerî etkisini doğrudan gösterir; hedefe ulaşan filo da
ikon yanındaki küçük bonus rozetiyle aynı etkiyi tekrar görünür kılar. Rozet
hover'ında hedef bölge ve uygulanan etki tooltip'te ayrıntılı gösterilir.
Her tur başında görevli
oyuncu filosu komşuluk grafiğinde deterministik BFS ile hedefe yaklaşır;
Görev, filo ikonunda ilgili rol rozetiyle ve panel footer'ında hedef metniyle
görünür; Abluka'nın ayrıca `A` kare rozeti yoktur, kırmızı yüzde rozeti yeterlidir.
Görevlerin ekonomik/askerî rolleri ayrıdır: Abluka görevi hedef denizdeki savaş
gemilerini ticaret rotası ve liman ikmal kesintisine dahil eder; bir gemi yüzde
50, iki veya fazlası yüzde 100 kesinti üretir. Devriye görevi dost rota/liman
üzerindeki düşman abluka gücünü aynı sayıdaki savaş gemisi kadar azaltır ve
devriye gemisi kendi başına abluka kesintisi üretmez. Escort, aynı denizdeki
nakliye savunmasına yüzde 15 deniz savunması ekler; toplam bonus yüzde 30 ile
sınırlıdır.
Kaynak/test: `internal/army/army.go`, `internal/state/naval_mission.go`,
`internal/game/player_naval_mission.go`, `internal/render/naval_mission_panel.go`,
`internal/save/compact.go`; `internal/{state,game,render,save}/*naval_mission*_test.go`.

## AI Deniz Taşıma Akışı

AI kara ordularını nakliye filosuna bindirip indirebilir:

- 1300 görevinde seçilmiş kara ordusu güvenli dost rotayla çıkış limanına gider; yeterli
  tek filo kapasitesi hazırsa çıkış denizini seçip gemiye biner.
- Boş transport filoları görev çıkış denizinde toplanır. Yüklenmiş filo sonraki turda
  aktif plandan yeniden tanınır, deniz BFS'iyle objective kıyısına gider ve uygun hedefe
  çıkar. Görev savaş gemileri de taşıma hattına yaklaşır.
- Aktif objective'in hedefi savaş state'i değiştikten sonra barışta kalmışsa, yüklü filo
  bu hedefte kilitlenmez; mevcut savaş düşmanları arasındaki ulaşılabilir kıyıları tarar,
  en yakın çıkarılabilir hedefe retarget olur ve savunucu gücü üstün olan kıyıyı atlar.
  Bu fallback yeni savaş ilan etmez ve yalnız runtime mission state'ini değiştirir.
- Somut görev yoksa boş 1300 transport filosu uzak deniz veya rastgele yabancı kıyı aramaz;
  görev taşımayan savaş gemileri ise `Patrol` rolüyle güvenlik ve ticaret denizlerini
  devriye gezer.
  Eski save'den yük taşıyan ama objective'i kalmayan filo, yalnız komşu güvenli dost
  kıyıya tahliye yapar.
- Diğer senaryolarda kara ordusu `chooseBestMove()` içinde komşu deniz bölgesini, o
  denizde uygun `transport` filosu ve pozitif `aiEmbarkScore()` varsa seçen legacy
  davranışı sürdürür.
- `executeMove()` kara → deniz geçişinde birimleri filonun `EmbarkedUnits` alanına taşır ve kara ordusunu haritadan kaldırır.
- Donanma `EmbarkedUnits` taşıyorsa komşu kara bölgesine çıkarma (`disembark`) yapar; yeni kara ordusu üretilir. Hedef düşman tahkimatlı kıyıysa çıkarma savaşı/fetih yerine bu orduyla kuşatma başlatılır; mevcut kuşatma varsa yasal destek olarak katılır.
- Hedef kara bölgesi sahipsizse başarılı çıkarma sonrası bölge artık AI sahipliğine yazılır; eski bug'lı save'lerde sahipsiz kalmış ama tek taraflı işgal altında olan kara bölgeleri yükleme/tur çözümlemesinde toparlanır.
- Düşman kıyıya çıkarma yalnızca savaş halindeyken yapılır; barışta AI çıkarma denemez.
- Düşman kıyıda ordu varsa AI çıkarma hedeflemesinde güç kıyası yapar; zayıfsa çıkarma girişimini atlar.
- Tahkimatsız düşman kıyısındaki çıkarma savaşı `combat.ResolveBattleWithMods()` ile çözülür; kazanırsa çıkarma ordusu karaya iner ve bölge el değiştirir. Çıkarma savaşında kullanılan komutan, filoda taşınan kara ordusunun komutanıdır; filo komutanı bu savaşa dahil edilmez.
- Legacy boş deniz hareketinde `aiSeaPressure()` düşman kıyı yoğunluğu, boş/sahipsiz
  kıyı fırsatı, mevcut dost filo yoğunluğu ve taşıma yükünü birlikte skorlar.

Kaynak kod:
- `internal/ai/naval_mission.go`
- `internal/ai/naval_threat.go`
- `internal/ai/ai.go`

Testler:
- `internal/ai/naval_mission_test.go`
- `internal/ai/naval_threat_test.go`
- `internal/ai/ai_test.go`

---

## Birim Alımı (`aiRecruitAndBuild`)

AI birim alımında sadece altın değil, birim reçetesindeki kaynakları da tüketir; `aiMinGoldReserve` korunurken diğer kaynaklar yetersizse alım yapılmaz. Birimler artık anında orduya eklenmez; `GameState.ProductionQueue` içine `kind=unit` emri yazılır ve çözümleme fazında tamamlanır.

1300'de bu akışın başlangıç noktası artık yalnız mevcut ordu sayısı değildir.
`aiForceRequirements()` her devletin sahip olduğu kara bölgelerinin toplam nüfusundan
`1 birlik / 200 nüfus` temel kara rezervi çıkarır; genişleme planında hedef `%25`,
savaş/kritik tehditte `%50` yükselir ve her durumda gerçek `ManpowerCap` ile sınırlanır.
Bekleyen üretim emirleri hedefe dahil edilir. Devlet bu tabanın altındaysa AI önce
güvenli, ikmali karşılanabilen bir kışla hattını bulur; saldırı rally rotası henüz
kurulmadıysa bile iç bölgede rezerv yetiştirebilir. Uygun birim veya kışla maliyetinin
tahıl, demir, kereste, taş, baharat ya da kumaş girdisi eksikse aynı ortak ticaret ağı
tedarik zinciri bunu üretimden önce satın almaya çalışır.

Aktif plan `expand` ve `target_faction_id` taşıyorsa nüfus tabanı tek başına yeterli
sayılmaz. AI, hedef devletin aktif ve üretim kuyruğundaki **kara** gücünü hesaplar;
kendi planlanan kara gücü bu değerin en az `%135`ine ulaşana kadar birlik hedefini
yükseltir. Hedef devlet daha çok birlik yetiştirdikçe bu eşik sonraki AI turunda yeniden
hesaplanır. Böylece örneğin Fransa'nın HRE topraklarına yönelik tarihsel planı, HRE'nin
gerçek kara gücüne karşı hazırlık yapar. Deniz birimleri bu kara seferi hesabına dahil
edilmez; deniz hazırlığı aşağıdaki ayrı filo hedefiyle yönetilir.

Kıyı savunması da aynı şekilde coğrafi taban taşır: her iki kıyı bölgesi için en az bir
`naval_war` birimi hedeflenir; savaş/deniz tehdidinde hedef iki katına çıkar. Merchant
ve nakliye gemileri bu sayıyı karşılamaz. `ai_strategies.json` içindeki
`naval_focus: true` ise bu genel oranı, kıyı başına iki savaş gemisi ve en az altı
gemilik ana filo hedefiyle değiştirir; ayrıca deniz bütçesini `%35`e çıkarır. Böylece
Venedik, Ceneviz, İngiltere ve Portekiz gibi devletler kodda hardcode edilmeden senaryo
verisinden gerçek deniz gücü olarak davranır. Savaş gemisi eksikse AI önce gerekli deniz
teknolojisini araştırma skorunda yükseltir, sonra uygun kıyıda gereken liman seviyesini
kurar ve son olarak savaş gemisi kuyruğunu açar. Bu sıradaki ilk gerçek maliyet,
`aiProcureStrategicResources()` için kaynak talebine dönüşür.

Manpower sıkışıksa önce kışla inşa eder. Hiç kara birimi ve hiç kışlası olmayan devlet,
manpower kapasitesi doluluğa yaklaşmamış olsa bile ilk kışlayı kuyruğa alır; böylece
boş/legacy ordu kaydı AI'nin askerî üretime başlamasını kilitlemez. Sonra
`aiSelectBestUnit()` ile birim seçer:

1300 senaryosunda seçim yukarıdaki **Plan Bazlı Kara Ordu Kompozisyonu** modeliyle
yapılır. Aşağıdaki sabit sıra yalnız diğer senaryoların legacy davranışıdır:

| Öncelik | Birim | Koşul |
|---|---|---|
| 1 | `elite_infantry` | Altın ≥ 350 + rezerv, teknoloji tamamlandıysa |
| 2 | `heavy_cavalry` | Altın ≥ 450 + rezerv, teknoloji tamamlandıysa |
| 3 | `infantry` | Altın ≥ 180 + rezerv |
| 4 | `cavalry` | Altın ≥ 300 + rezerv |
| 5 | `light_cavalry` | Altın ≥ 200 + rezerv |
| 6 | `cannon/bombard` | Altın ≥ 650 + rezerv, savaş halinde |
| 7 | `militia` | Varsayılan |

Queue davranışı:

- Pending kara birimleri manpower hesabına dahil edilir.
- AI gerekli teknoloji, bina ve bina seviyesi olmayan birimi seçmez.
- Bölge başına pending üretim emri `20` ile sınırlıdır.
- AI kara recruit seçiminde yalnız ilk uygun bölgeyi almaz; kışla throughput'u dolu bölgeyi atlayıp aynı tur kapasitesi kalan başka uygun bölgeye dağılır. Tüm uygun kışla hatları doluysa yeni kara emri açmaz.
- 1300 senaryosunda uygun hatlar ayrıca stratejik recruitment bölgesi skoru ile
  sıralanır; diğer senaryolarda remaining-capacity/kışla seviyesi seçimi korunur.
- Mevcut ordu doluysa üretim tamamlandığında oyuncu ile aynı `completeLandUnit()` yolu yeni ordu oluşturabilir.

---

## Eksik / Planlanan

- [x] AI çoklu ordu konsolidasyonu (dağınık ordular ana orduya katılsın; ancak lojistik baskı altındaki kara bölgede konsolidasyon durur)
- [x] AI fırsatçı savaş ilanı (sınır, güç dengesi ve ilişki skoruna göre sınırlı proaktif savaş)
- [x] Diplomasi teklif önceliklerini teknoloji farkı ve uzun vadeli tehdit seviyesiyle zenginleştir
- [x] Transport yanında savaş gemisi escort üretimini de filo bileşimine kat
- [x] Escort üretimini çoklu deniz baskısı ve birden fazla cepheye göre ölçekle
- [x] 1300 transport/liman/escort üretimini somut denizaşırı göreve ve gerçek kapasite açığına bağla
- [x] 1300'de görevsiz filoların uzak deniz dolaşımını durdur
- [x] 1300 deniz rotalarını düşman filo gücü ve `%110` görev filosu güvenlik eşiğine bağla
- [x] Tehdit edilen görev limanını ihtiyaç kadar escort üretiminde önceliklendir
