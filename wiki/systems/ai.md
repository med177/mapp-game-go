---
type: system
tags: [ai, strategy, coalition, difficulty]
last_updated: 2026-07-15
related: [systems/combat, systems/diplomacy, architecture/game-loop]
---

# Yapay Zeka Sistemi

**Kaynak:** `internal/ai/ai.go`, `internal/ai/turn_stepper.go`

## Genel Yapı

Her `PhaseAITurn` artık iki katmandan oluşur:

- `ai.TakeTurn()` hâlâ tam turu tek çağrıda çözebilen saf AI entrypoint'idir.
- `ai.TurnStepper` ise aynı mantığı adım adım açar; oyun döngüsü bunu kullanarak her AI devletini sırayla görünür işler.

Oyun katmanı AI fraksiyonlarını `FactionOrder` sırasıyla dolaşır, her fraksiyon için `TurnStepper.Step()` çağırır ve her step arasında kısa bekleme ekler. Böylece harita bir anda "snap" olmaz; yakın cephedeki hareketler tek tek görünür.

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

### Lojistik Farkındalığı

AI artık dost kara bölgelerini sadece diplomasi/savaş açısından değil, ikmal baskısı açısından da puanlar.

- Kaynak bölge aşırı doluysa (`grain_upkeep` toplamı bölgenin efektif tahıl + yerleşim tamponu + sınırlı stok desteğini aşıyorsa) `scoreMove()` dost komşular arasında baskıyı azaltan bölgeye pozitif puan verir.
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

`aiHandleDiplomacy()` her AI turunda ilişkileri tarar:

- `war` ilişkisinde skor çok düşmüşse veya AI askeri/bölgesel olarak gerideyse barış teklif eder
- `peace` ilişkisinde ittifak için artık sadece skor yetmez; kara sınırı, aktif ticaret, ortak düşman veya ortak büyük tehditten en az biriyle gerçek coğrafi/stratejik bağ aranır. AI ayrıca küçük devletlerde düşük müttefik tavanı uygular, `ai_expansion_targets` ile çakışan hedeflere ortak tehdit yoksa ittifak teklif etmez ve büyük gücün tek bölgeli/zayıf devlete attığı alliance teklifi için ayrıca gerçek askeri veya stratejik fayda arar
- `peace` ilişkisinde skor yeterliyse ve bağlanabilir kara/deniz hattı varsa ticaret dener; salt uzak ve nötr devletlere sırf kapasite var diye trade açmaz
- vassal durumundaki AI bağımsız diplomasi açmaz; overlord'u olmayan devletler ise başka bir overlord'a bağlı hedeflerle doğrudan müzakere başlatmaz
- `allied` ilişkide ortak tehdit kalmamış, ticaret/sınır bağı yok olmuş, tarihsel genişleme hedefiyle çatışan ya da büyük güç için artık anlamlı katkı üretmeyen zayıf ittifaklar AI tarafından iptal edilebilir
- normal/zor zorlukta, ticaret/ittifak ilişkisini bozmadan, sınır komşusu zayıf bir hedefe karşı güç ve cephe üstünlüğü varsa tek bir fırsat savaşı açabilir
- bekleyen barış, ittifak ve ticaret teklifleri teknoloji farkı ve uzun vadeli tehdit baskısına göre önceliklenir; oyuncuya daha baskı altındaki teklif önce gösterilir ve prompt içinde tür ile kısa sebep bilgisi görünür. İttifak teklifinde AI ayrıca turn + taraf kimliğine bağlı deterministik hafif rastgelelik kullanır; böylece aynı uygun çerçevede her tur mekanik olarak sabit spam yerine bazen teklif açar, ama yüksek olasılıklı ortak tehdit senaryoları yine güvenilir biçimde görünür. Desteksiz dış ittifakların relation skoru ise artık her tur otomatik yükselmez; destek yoksa yavaşça aşınır.

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

`ai_expansion_targets` rastgele saldırganlık değildir: hedef olmayan fraksiyonlarda eski negatif ilişki kapısı korunur, ticaret/ittifak ilişkileri bozulmaz ve güçsüz AI daha güçlü hedefe savaş açmaz. 1300 senaryosunda Osmanlı için Doğu Roma, Germiyan, Karesi ve Ahiler; Reconquista, Yüz Yıl Savaşları ve doğu bozkır cepheleri için ilgili tarihsel hedefler bu listeyle tanımlıdır.

AI ve oyuncu aynı `internal/diplomacy` motorunu kullandığı için:

- kabul/red kuralları tutarlıdır
- ticaret rotaları aynı şekilde açılıp kapanır
- AI ticaret yaptığı veya müttefik olduğu hedefe saldırmaz
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

| Seviye | Fark |
|---|---|
| 1 (Kolay) | Pasif AI, yavaş büyüme |
| 2 (Normal) | Dengeli strateji |
| 3 (Zor) | +300 başlangıç altın, +100 tahıl; koalisyon mantığı aktif |

Zorluk 3 başlangıç bonusu `resetToNewGame()` içinde uygulanır — `internal/game/game.go:337`

---

## Teknoloji Araştırma (`aiResearch`)

Aktif araştırma yoksa başlatır. Öncelik sırası:

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

Her tur en fazla bir bina üretim emri açar. Bina anında `Region.Buildings` listesine eklenmez; AI de `GameState.ProductionQueue` içine `kind=building` emri yazar ve tamamlanma `applyProductionTicks()` sırasında olur. Öncelik:
1. **Pazar** (prio 80) — her zaman uygun
2. **Çiftlik** (prio 60) — `BaseGrainOutput < 20` bölgelere
3. **Sur** (prio 50) — sınır bölgelerine (komşuda farklı fraksiyon varsa)

AI bina inşasında yalnız altını değil, bina reçetesindeki `grain/iron/timber/stone` maliyetlerini de kontrol eder. Kurulu seviye + pending seviye `max_per_region` sınırını aşamaz; inşa süresi oyuncu tarafındaki seviye modeliyle uyumlu şekilde mevcut seviye ve queued seviye arttıkça uzar.

---

## Deniz Stratejisi (`aiNavalStrategy`)

Kıyı bölgesi varsa:
1. Limansız kıyı bölgesine liman üretim emri aç
2. Filo limiti artık dinamiktir:
   `1` temel + `1` ek kıyı baskısı (3+ kıyı bölgesi) + savaşta `1` ek filo, üst sınır `3`
   Aynı denize bağlı birden fazla pending gemi emri, filo limiti hesabında tek yaklaşan filo olarak değerlendirilir.
3. Yeni `transport` emri, ilk bulunan limana değil `aiSeaPressure()` skoru en yüksek deniz hattına bağlı limanda kuyruğa girer
4. Aynı turda transport hattı ya da mevcut transport filosu olan savaşçı AI, deniz baskısı ve birden fazla cephe tespit ederse uygun limit dahilinde birden çok eskort `warship` emri de kuyruklar
5. Üretim tamamlandığında mevcut `completeNavalUnit()` akışı komşu denizde filo oluşturur veya mevcut filoya birim ekler
6. AI artık liman hattı doygunsa aynı kıyıya kör yeni emir yığmaz; mümkünse başka serbest liman hattına dağılır, hepsi doluysa yeni deniz emri açmaz

---

## AI Deniz Taşıma Akışı

AI artık kara ordularını nakliye filosuna bindirip indirebilir:

- Kara ordusu `chooseBestMove()` içinde komşu deniz bölgesini, o denizde uygun `transport` filosu varsa ve `aiEmbarkScore()` pozitifse seçer.
- `executeMove()` kara → deniz geçişinde birimleri filonun `EmbarkedUnits` alanına taşır ve kara ordusunu haritadan kaldırır.
- Donanma `EmbarkedUnits` taşıyorsa komşu kara bölgesine çıkarma (`disembark`) yapar; yeni kara ordusu üretilir.
- Hedef kara bölgesi sahipsizse başarılı çıkarma sonrası bölge artık AI sahipliğine yazılır; eski bug'lı save'lerde sahipsiz kalmış ama tek taraflı işgal altında olan kara bölgeleri yükleme/tur çözümlemesinde toparlanır.
- Düşman kıyıya çıkarma yalnızca savaş halindeyken yapılır; barışta AI çıkarma denemez.
- Düşman kıyıda ordu varsa AI çıkarma hedeflemesinde güç kıyası yapar; zayıfsa çıkarma girişimini atlar.
- Çıkarma savaşı yine `combat.ResolveBattleWithMods()` ile çözülür; kazanırsa çıkarma ordusu karaya iner ve bölge el değiştirir.
- Boş deniz hareketi kör yapılmaz; `aiSeaPressure()` düşman kıyı yoğunluğu, boş/sahipsiz kıyı fırsatı, mevcut dost filo yoğunluğu ve taşıma yükünü birlikte skorlar.

Kaynak kod:
- `internal/ai/ai.go:377`
- `internal/ai/ai.go:438`
- `internal/ai/ai.go:666`

Testler:
- `internal/ai/ai_test.go:67`
- `internal/ai/ai_test.go:119`
- `internal/ai/ai_test.go:172`
- `internal/ai/ai_test.go:221`

---

## Birim Alımı (`aiRecruitAndBuild`)

AI birim alımında sadece altın değil, birim reçetesindeki kaynakları da tüketir; `aiMinGoldReserve` korunurken diğer kaynaklar yetersizse alım yapılmaz. Birimler artık anında orduya eklenmez; `GameState.ProductionQueue` içine `kind=unit` emri yazılır ve çözümleme fazında tamamlanır.

Manpower sıkışıksa önce kışla inşa eder. Sonra `aiSelectBestUnit()` ile birim seçer:

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
- Mevcut ordu doluysa üretim tamamlandığında oyuncu ile aynı `completeLandUnit()` yolu yeni ordu oluşturabilir.

---

## Eksik / Planlanan

- [x] AI çoklu ordu konsolidasyonu (dağınık ordular ana orduya katılsın; ancak lojistik baskı altındaki kara bölgede konsolidasyon durur)
- [x] AI fırsatçı savaş ilanı (sınır, güç dengesi ve ilişki skoruna göre sınırlı proaktif savaş)
- [x] Diplomasi teklif önceliklerini teknoloji farkı ve uzun vadeli tehdit seviyesiyle zenginleştir
- [x] Transport yanında savaş gemisi escort üretimini de filo bileşimine kat
- [x] Escort üretimini çoklu deniz baskısı ve birden fazla cepheye göre ölçekle
