---
type: system
tags: [diplomacy, relations, stance, faction]
last_updated: 2026-07-22
related: [world/factions, systems/ai, architecture/state-management]
---

# Diplomasi Sistemi

**Kaynak:** `internal/diplomacy/diplomacy.go`, `internal/diplomacy/peace_assessment.go`, `internal/diplomacy/alliance_strategy.go`, `internal/faction/faction.go`, `internal/game/game.go`

## İlişki Yapısı

```go
type Relation struct {
    FactionA, FactionB FactionID
    Score   int              // -100 (düşman) → +100 (müttefik)
    Stance  DiplomaticStance
}
```

`RelationKey(a, b)` → her zaman sıralı `"a|b"` string'i üretir (çift kayıt önler).

Vassallık relation duruşu olarak değil, doğrudan fraksiyon üstünde `OverlordID` alanıyla tutulur. Böylece:

- bir devlet yalnız tek bir overlord'a bağlı olur
- realm içi dostluk `StanceAllied` ile korunurken hiyerarşi ayrı kalır
- vassal için üçüncü taraf diplomasi yasağı relation katmanına zorla sığdırılmaz

1300 senaryosunda AI savaş sonrası aynı modeli kullanır. Anadolu beylikleri objective'i
aktifken son toprağında yenilen zayıf ve dış müttefiksiz hedef,
`ForceVassalizeAfterWar()` üzerinden vassal bırakılabilir. Direnç eşiğini aşan veya
objective tarafından stratejik ilhak bölgesi sayılan hedefte normal fetih sürer; AI için
ayrı bir vassallık state'i ya da relation duruşu oluşturulmaz.

---

## Diplomatik Duruşlar (DiplomaticStance)

| Duruş | Geçiş Koşulu | Puan Etkisi |
|---|---|---|
| `StancePeace` | Varsayılan / barış sonrası | Score = -20 |
| `StanceWar` | Savaş ilan edildiğinde | Score = -80 |
| `StanceTrade` | Ticaret anlaşması | Score +15 |
| `StanceAllied` | İttifak | Score +20 |

Diplomatik duruşların görünen adları, badge metinleri ve editörde dolaşım sırası `internal/faction/stance_metadata.go` içinde merkezileştirilmiştir (`DiplomaticStanceLabelTR`, `DiplomaticStanceBadgeTR`, `AllDiplomaticStances`, `NextDiplomaticStance`).

1300 senaryosunda ittifak kabulü yalnız ilişki ve din puanından oluşmaz. Hedef AI,
teklif sahibinin ortak düşman/büyük tehdide katkısını, tampon konumunu, tehdit cephesinde
bulunan gerçek ordu gücünü, ticaret erişimini ve partner güç/bölge katkısını kendi
perspektifinden değerlendirir. Teklif sahibi AI de aynı bileşenleri ters perspektifte
asgari girişim eşiğinden geçirir. Aktif objective çakışması hard block'tur; statik
gelecek genişleme hedefi ortak tehditle aşılabilen yumuşak cezadır.

**Geçiş kısıtları:**
- Savaştayken ittifak veya ticaret kurulamaz
- İttifak için `Score >= 25` gerekir; ayrıca salt din bazlı varsayılan skor tek başına yeterli sayılmaz. İki taraf arasında önceden değişmiş gerçek ilişki puanı, kara sınırı, fiili ticaret erişimi, ortak düşman veya ortak büyük tehditten en az biriyle diplomatik temas aranır. Buna ek olarak kara sınırı, fiili ticaret erişimi, ortak düşman veya ortak büyük tehditten en az biriyle gerçek coğrafi/stratejik bağ gerekir. Doğrudan sınır tehdidi mutlak blok değil, kabul şansını düşüren bir cezadır. Ortak düşman veya ortak büyük tehdit bu cezayı kısmen/tamamen telafi edebilir. Aynı din ayrıca doğrudan kabul şansı bonusu verir; yakın mezhep küçük bonus, sert mezhep ayrımı ise ek ceza üretir.
- Ticaret için `Score >= 15`, iki tarafın da kara bölgesi ve toplam `trade_capacity >= 4` olmalıdır
- Ticaret için aktif partner limiti (`4`) dolu olmamalıdır; doğrudan sınır tehdidi ise artık mutlak blok değil, kabul şansını düşüren bir cezadır
- Ticaret anlaşması ayrıca bağlanabilir gerçek bir hat ister: ya iki realm arasında kesintisiz kara hattı, ya da her iki tarafta liman olup komşu deniz bölgeleri üzerinden bağlanabilen bir deniz hattı bulunmalıdır
- `StanceAllied` ile `TradeRoutes` artık ayrı kavramlardır; müttefiklik otomatik ticaret açmaz, ama müttefikken ayrıca ticaret anlaşması kurulabilir
- Zaten aynı duruştaysa tekrar kurulamaz; ancak `StanceTrade` duruşunda rota kaydı eksikse teklif akışı rotayı yeniden kurar
- Vassal-overlord bağı ayrı tutulur; iç realm relation'ları normalizasyonda `allied` çizgisine çekilir ve doğrudan overlord-vassal arasında kapasite/partner sınırından bağımsız iki yönlü ticaret rotası garanti edilir

---

## Oyuncu Diplomatik Aksiyonları

`internal/game/game.go`

| Aksiyon | Fonksiyon | Koşul |
|---|---|---|
| Savaş ilan et | `declareWar()` | Zaten savaşta değilse; önce koalisyon önizlemesi açılır, hedefin vassalları ile iki tarafın çağrılabilir müttefikleri ve katılım ihtimali gösterilir |
| Barış teklif et | `proposePeace()` | Savaş halinde gerekli; 1300'de kalıcı savaş ledger'ı, objective, toprak/kayıp dengesi, süre, güç, ekonomi, çoklu savaş ve başkent tehdidi değerlendirilir. Diğer senaryolarda legacy savaş baskısı + güç + ekonomik stres modeli korunur |
| Heyet gönder | `improveRelations()` | Savaşta değil + `40` altın; ilişkiyi deterministik `+8` artırır |
| Hediye gönder | `sendGift()` | Savaşta değil + `120` altın; ilişkiyi deterministik `+15` artırır |
| İttifak kur | `proposeAlliance()` | Savaşta değil + `Score >= 25`; varsayılan din skorunun ötesinde diplomatik temas ve coğrafi/stratejik bağ gerekir. Kabul şansı ilişki puanı, doğrudan din uyumu bonusu, güç/bölge farkı, mevcut trade bağı, doğrudan sınır tehdidi cezası ve `ortak düşman / ortak büyük tehdit` bonuslarıyla değerlendirilir |
| Ticaret anlaşması | `proposeTrade()` | Savaşta değil + `Score >= 15` + iki tarafın da kara bölgesi ve yeterli ticaret kapasitesi var; ayrıca bağlanabilir kara/deniz ticaret hattı gerekir. Aynı helper kabul şansını ve UI'daki engel nedenini birlikte üretir |
| İttifakı bitir | `cancelAlliance()` | Dış devletle aktif ittifak varsa; mevcut ticaret rotaları korunur ve relation `trade/peace` durumuna iner |
| Ticareti bitir | `cancelTrade()` | Aktif ticaret rotası varsa; rotalar kaldırılır, mevcut ittifak korunur |
| Vassallık teklif et | `offerVassalization()` | Teklif eden zaten vassal değilse + hedef başka devlete bağlı değilse + `Score >= 55` + belirgin askeri/bölgesel üstünlük varsa |
| Vasallığı bitir | `releaseVassal()` | Yalnız oyuncunun doğrudan vassalında; devlet bağımsızlaşır, overlord ile ticaret anlaşması devam eder |
| Vassalı ilhak et | `annexVassal()` | Yalnız oyuncunun doğrudan vassalında ve onay sonrası; tüm bölgeler, kuvvetler, kaynaklar ve üretim emirleri oyuncuya devredilir, vassal fraksiyon elenir |

Teklifler artık otomatik kabul edilmez; oyuncu ve AI aynı değerlendirme motorunu kullanır.

Savaş ilanı artık `ExecuteWarDeclaration()` üstünden ayrı bir koalisyon akışı izler:

- Hedefin mevcut vassalları savaşa kesin katılır.
- Hem oyuncu hem AI tarafında dış müttefikler için ayrı `AssessWarCall()` değerlendirmesi yapılır.
- Oyuncu, savaş önizleme modalında hangi müttefiklerini çağıracağını checkbox ile seçer.
- Seçilip de çağrıya gelmeyen müttefiğin ittifakı bozulur; ilişki puanı `-10` düşer.
- Aynı deterministik helper savunan tarafın müttefikleri için de kullanılır; bu yüzden modalda görülen olasılık savaş resolve anındaki gerçek çağrı sonucuyla aynı kaynaktan beslenir.
- Resolve tamamlanınca render tarafında ayrı bir `Savaş Özeti` modalı açılır; burada gerçekten katılan coalition üyeleri, katılmayan müttefikler ve iki tarafın toplam askeri gücü gösterilir.

Diplomasi panelinin sağ kolonu seçili devletin güncel diplomatik ağını gösterir:

- `Savaşta` ve `İttifaklar` listeleri fraksiyonun tüm aktif relation kayıtlarından üretilir.
- `Ticaret Anlaşmaları` yalnız aktif, askıya alınmamış iki taraflı `TradeRoutes` kayıtlarını gösterir; bu yüzden ittifak ile ticaret birbirine karıştırılmaz.
- Yeni ticaret rotası kurulurken hedef fraksiyonun üç aylık tahıl rezerv açığı ve kaynak fraksiyonun kapasite üstü tahıl fazlası değerlendirilir; ikisi de pozitifse ilgili yön `GoodGrain` olarak oluşturulur ve sonraki ekonomi tick'lerinde normal altın/stok kontrolleriyle tahıl ithalatı gerçekleşir.
- Aynı vassal realm içindeki normalizasyon kaynaklı `StanceAllied` kayıtları dış ittifak sayılmaz; overlord veya bağlı devlet sayısı üst bilgide ayrıca gösterilir.
- Teklif geçmişi sağ kolonda sürekli yer kaplamaz; `Geçmiş` düğmesiyle açılır ve `İlişkiler` düğmesiyle güncel ağa dönülür.
- Standart teklif düğmeleri `ActionBlockReason()` sonucuna göre aktif veya `PASİF` çizilir; pasif düğmeler fare ve klavye odağına alınmaz. Dış devletle ilişki kurulmuşsa aynı `İttifak / Ticaret` düğmeleri `İttifakı Bitir / Ticareti Bitir` işlemine dönüşür ve alt aksiyon `Anlaşmayı Bitir` olur. Savaş `Barış`, vassallık ise `Vasallığı Bitir` yoluyla sona erdirilir. Doğrudan oyuncu vassalında sağ-alt `Vassal Yönetimi` kartı ayrıca onaylı `Vasallığı Bitir / İlhak Et` eylemlerini gösterir.
- `Savaş` aksiyonu artık doğrudan submit edilmez; teklif sayfasından veya harita üstü saldırı girişiminden sonra özel savaş önizleme modalı açılır. Bu modal iki cepheyi yan yana gösterir: kesin katılacak vassallar ve zaten savaşta olan müttefikler üstte, çağrılabilir müttefikler ise olasılık etiketiyle altta listelenir.

---

## İlişki Puanı Değişimleri

| Olay | Puan Değişimi |
|---|---|
| Savaş ilanı | -80 (sabit) |
| Barış | -20 (sıfırlama) |
| Heyet | +8 |
| Hediye | +15 |
| Ticaret | +15 |
| İttifak | +20 |
| Oyuncunun normal diplomasi teklifini reddetme | -3 |
| `ApplyRelationDecay()` | Savaşta skor düşer; barış/ticaret yumuşar, desteklenmeyen ittifaklar ise aşınır |
| Ortak düşman | +bonus (AI koalisyon mantığında) |
| Din bonusu/cezası | `religion.Relation(a,b)` — başlangıç skoru; +25 / -20 / -30 / -40 |

→ `applyRelationDecay` tur çözümleme sırası: [[architecture/game-loop]]

### Teklif Retleri ve Tekrar Denemeler

Oyuncunun reddettiği barış, ittifak, ticaret veya vassallık teklifleri
`GameState.OfferRejectionTurns` içinde aktör-hedef-aksiyon anahtarıyla tutulur.
İlişki skoru her normal ret için `-3` azalır. Aynı teklif üç tur boyunca yeniden
gönderilmez; bekleme bitince AI her tur `internal/ai/ai.go:aiDiplomacyOfferRoll`
üzerinden %35 tekrar deneme zarı atar. Bu kayıt compact ve debug/legacy save
akışlarında korunur. Savaş çağrısının mevcut ittifak bozma ve `-10` ilişki sonucu
bu kurala dahil değildir.

---

## Ticaret Entegrasyonu

`GameState.TradeRoutes` artık diplomasi motoru tarafından yönetilir.

- Ticaret anlaşması kabul edilince iki yönlü rota oluşturulur
- Aynı iki fraksiyon için rota çoğaltılmaz; mevcut çift önce temizlenir
- İttifak duruşu ticaretten bağımsızdır; müttefik iki devlet arasında rota varsa `StanceAllied` korunur
- Savaş ilanı veya barış kabulü iki taraf arasındaki aktif rotaları kapatır
- Save/load veya eski kayıt migrasyonu sırasında elenmiş fraksiyon, geçersiz relation, duplicate rota veya artık bağlanamayan dış ticaret hattı kalmışsa `SanitizeTradeRoutes()` bunları temizler; böylece yıkılmış devlet adları ya da kopmuş rota kayıtları trade listesine geri sızmaz
- Vassal olan devletin üçüncü taraf trade rotaları `NormalizeVassalage()` veya doğrudan vassallık kabul akışında kapatılır
- Rotalar soyut anlaşma modelidir; harita üstü pathfinding ile üretilmez

Rota detayları:

- Mal türü gönderen fraksiyonun en değerli mevcut stokuna göre seçilir
- `AmountPerTurn`, iki tarafın toplam `trade_capacity` değerinden türetilir
- Altın getirisi tur çözümlemesinde `TradeRoute.GoldEarned()` ile hesaplanır

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — `internal/faction/loader.go`

Tüm fraksiyon çiftleri için skor `internal/religion.Relation()` sonucuyla başlatılır. Varsayılan duruş barıştır; Sünni-Şii çiftleri başlangıçta savaş durumuna alınır.

`1300_ottoman_rise` senaryo override'ı bu varsayılanı gerçek 1300 cepheleriyle düzeltir:
Osmanlı-Doğu Roma, Memlük-İlhanlı, Aragon-Kastilya, Aragon-Napoli, İngiltere-Fransa,
İngiltere-İskoçya ve Fransa-HRE savaşta başlar. Aragon-Granada müttefik kalır;
Venedik-Ceneviz ile Doğu Roma-Bulgaristan barışta bırakılır. Flandre, HRE'nin vassalı
olduğu için Flandre-Fransa düşmanlığı HRE-Fransa kök savaşıyla birlikte koalisyona
katılır; overlord-vassal arasındaki iç ticaret ve geçiş garantisi korunur. İlişki çiftleri
loader'da sıralı faction ID'leriyle üretildiğinden save/replay yönü deterministiktir.

---

## AI Diplomasi Davranışı

`aiHandleDiplomacy()` ve `FormCoalitionAgainstPlayer()` — zorluk 3 koalisyon dahil aynı motoru kullanır

AI:

- 1300'de savaşın ilk üç turunda normal barış denemez; objective tamamlanması, toprak ve
  birlik kaybı, süre/durgunluk, güç ve ekonomik stres, çoklu savaş ve başkent tehdidi
  barış baskısını belirler. Başkent tehdidi veya askerî çöküş erken kapıyı aşabilir;
  reddedilen teklif aynı savaşta üç tur cooldown'a girer
- ittifakta artık sadece `ortak düşman` sert filtresine bakmaz; aynı alliance assessment helper'ını kullanır ve `ortak büyük tehdit` gördüğünde de teklif açabilir
- AI dış ittifak açarken artık stratejik bağ, müttefik kapasitesi, `ai_expansion_targets` gerilimi ve hedefin somut katkısını da dikkate alır; ortak tehdit yoksa uzak/alakasız, tarihsel hedef olan veya büyük güç için gerçek askeri/stratejik fayda üretmeyen küçük devlete ittifak spam atmaz
- barışta skor ve bağlanabilir kara/deniz hattı uygunsa ticaret açar
- vassal durumundaki AI bağımsız diplomasi ve savaş değerlendirmesi yapmaz
- dış ittifakın ortak tehdit/ticaret/sınır dayanağı kalmazsa relation skoru otomatik şişmez; AI yeterince zayıflayan veya artık anlamlı fayda üretmeyen ittifakı bozabilir
- koalisyon anında oyuncuya savaş açıp diğer AI'larla ittifak kurmaya çalışır

→ Detaylar: [[systems/ai]]

---

## Elenen Fraksiyon Temizliği

`internal/game/resolution.go:255` içindeki `checkEliminations()` artık bir fraksiyonun bölgesi kalmadığında:

- `IsEliminated=true` işaretler
- o fraksiyona ait tüm orduları kaldırır
- `GameState.Relations` içindeki o fraksiyonu içeren tüm diplomasi kayıtlarını siler
- o fraksiyona ait `TradeRoutes` ve `DiplomaticOffers` kayıtlarını da siler
- overlord ise bağlı vassal zincirini successor'a devreder veya serbest bırakır

Bu sayede elenen devletler diğer devletlerle diplomasi verisi taşımaya devam etmez.

---

## Vassallık Kuralları

- Vassallık kabul edilince hedef fraksiyonun `OverlordID` alanı doldurulur.
- Aynı realm içindeki relation kayıtları `NormalizeVassalage()` ile dost çizgiye çekilir.
- Vassal, overlord dışındaki devletlerle doğrudan savaş/barış/ittifak/ticaret/vassallık diplomasisi kuramaz.
- Üçüncü taraf devletler de vassal ile doğrudan diplomasi kuramaz; muhatap overlord'dur.
- Overlord savaş açarsa veya savaşa çekilirse vassal coalition olarak aynı savaşa girer.
- Ekonomi tick'inde vassal altın gelirinin `%20` kadarını overlord'a haraç olarak aktarır.
- Aynı realm içindeki devletler (`overlord`, doğrudan vassal ve aynı kök zincirdeki bağlı devletler) kara geçişi, dost kıyıya çıkarma, liman kullanımı ve mevcut kuşatmaya destek için ayrıca ittifak veya savaş ilanı gerektirmez.
- UI tarafında vassal sahibi bölgeler bölge bilgi panelinde `Bağlı: <overlord>` satırıyla ve ana haritadaki yerleşim marker'ı üstündeki küçük rozetle görünür hale getirilir; oyuncu kendi vassal bölgelerini seçtiğinde aynı blokta o devletten gelen `Haraç: +X altın/tur` satırı da görünür, böylece hiyerarşi ve ekonomik bağlılık diplomasi ekranı açılmadan da okunabilir.
- Oyuncu, savaşta bir devletin son kara toprağını düşürdüğünde fetih artık otomatik ilhakla kapanmak zorunda değildir; battle report sonrası açılan `Savaş Sonrası Düzen` modalında `İlhak Et` veya `Vassal Yap` seçilir.
- Bu savaş-sonrası vassallık akışı hedef devletin o son bölgesini yerel yönetimde bırakır, ama realm ilişkisini hemen normalize eder; yani savaş biter, hedef overlord'a bağlanır ve overlord'un aktif savaşlarına coalition olarak çekilir.

---

## Gelen Teklif Paneli (Oyuncu)

AI artık oyuncuya doğrudan barış sonucu dayatmaz. Savaş baskısı şartı oluştuğunda teklif `GameState.DiplomaticOffers` kuyruğuna eklenir:

- kaynak: `internal/ai/ai.go:87`
- kuyruk/çözümleme: `internal/diplomacy/offers.go`
- UI paneli: `internal/render/renderer.go` (`drawDiplomacyOfferDialog`, `handleDiplomacyOfferInput`)

Oyuncu teklif geldiğinde `Kabul Et` veya `Reddet` yanıtı verir; kabulde standart diplomasi motoru (`Execute`) çalışır, redde ise teklif kuyruktan düşer ve savaş sürer.

AI savaş ilanı sırasında oyuncu tarafında aktif bir ittifak varsa aynı kuyruk artık `join_war_call` tipiyle kullanılır:

- Müttefik AI başka bir devlete savaş ilan ettiğinde oyuncuya `Savaşa Katılım Çağrısı` gelir
- Bir AI oyuncunun müttefikine savaş ilan ettiğinde bu kez savunan müttefik oyuncuyu kendi safına çağırır
- Oyuncu kabul etmeden savaşa çekilmez; kabulde oyuncu realm'i ilgili düşman koalisyonuyla savaşa girer
- Reddederse çağrıyı yapan müttefikle ittifak bozulur ve ilişki puanı `-10` düşer
- Teklif AI turu içinde doğduysa sıra makinesi oyuncu cevabını bekler; kabul edilen çağrı aynı aktif AI deklaratörünün turunu kapatır

## Diplomasi Paneli

`internal/render/diplom.go`

- Panel iki adımdır: önce hedef devlet listesi, sonra teklif sayfası açılır.
- Hedef devlet listesi her satırda `Askeri güç` ve aktif devletler arasındaki `Güç sırası` (`X/Y`) değerlerini gösterir. Listenin üstündeki `Alfabetik`, `İlişki` ve `Güç Sıralaması` düğmeleri listeyi sırasıyla varsayılan ID alfabetiğine, oyuncuyla olan ilişki puanı azalan düzene veya standing sırası artan düzene göre yeniden düzenler. `İlişki` sıralamasında aynı ilişki puanına sahip hedefler içinde oyuncuyla kara sınırı paylaşan devletler önce gelir; kalan eşitlik faction ID'siyle çözülür.
- Teklif paneli artık çekirdek aksiyonların yanında `Heyet`, `Hediye` ve `Vassallık` seçeneklerini de gösterir.
- Hedef listesi artık panel gövdesi üzerinde mouse wheel ile kaydırılır; scroll sadece dar satır alanına değil panel bağlamına da bağlıdır.
- Hedef listesinde seçim `mouse down` anında değil, kısa click release anında kesinleşir; basılı tutup sürüklemek listeyi satır yüksekliği bazında kaydırır ve yanlışlıkla devlet seçmez.
- Liste kartları fraksiyon rengi accent şeridi, ilişki/duruş özeti ve görünür scrollbar ile çizilir; teklif sayfası aynı UI compose ailesindeki kart/panel çerçevesini kullanır.
- Teklif sayfasında `Hedef`, `Durum` ve `İlişki Skoru` blokları ayrı iç padding ile çizilir; vassal veya overlord hedeflerinde durum satırı düz stance yerine hiyerarşi etiketi gösterir. Alt footer butonları da ortak ikonlu `Button` primitive'inin `TextOffsetY` hizasına bağlandığı için `Geri` ve `Teklif Gönder` satırı artık icon/metin kaydırması üretmez.
- Bölge bilgi panelindeki devlet adı üzerinden açılan devlet detay yüzeyi de aynı diplomasi verisini özetler; üst devlet, vassal, ittifak, ticaret ve düşman listeleri tek scroll alanında gösterilir, böylece teklif paneline girmeden ilişkiler okunabilir.

## Müttefik Liman Kullanımı

`internal/game/game.go`, `internal/render/renderer.go`

- Donanmalar komşu deniz bölgelerine ek olarak, yalnız `port` binası tamamlanmış komşu kara bölgesine docking emri alabilir
- Docking yalnız iki durumda geçerlidir: bölge oyuncunun/devletin kendi toprağıysa veya iki fraksiyon arasında `StanceAllied` varsa
- Hedef kara bölgesinde `port` settlement olsa bile bina yoksa filo oraya taşınamaz; ancak cargo taşıyorsa kıyıya saldırı veya çıkarma yine yapılabilir
- Dock edilmiş filo deniz `RegionID` değerini korur, ama `DockedRegionID` ve `DockedSettlementID` üzerinden liman anchor'ında çizilir
- İttifak biter ya da liman el değiştirirse `sanitizeDockedFleets()` bu bağı düşürür ve filoyu en yakın deniz bölgesine çıkarır

---

## Eksik / Planlanan

- [ ] Bekleyen diplomatik teklif kuyruğu / çok adımlı müzakere
- [ ] İttifak için ortak geçiş hakkı veya askeri bonuslar
- [ ] Ticaret için dinamik piyasa / rota pathfinding
- [x] `internal/religion` paketi ayrıştırıldı
