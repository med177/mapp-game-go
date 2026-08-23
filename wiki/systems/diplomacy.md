---
type: system
tags: [diplomacy, relations, stance, faction]
last_updated: 2026-08-23
related: [world/factions, systems/ai, architecture/state-management, dev/data-format]
---

# Diplomasi Sistemi

**Kaynak:** `internal/diplomacy/diplomacy.go`, `internal/diplomacy/offers.go`, `internal/diplomacy/vassalage.go`, `internal/diplomacy/peace_assessment.go`, `internal/diplomacy/alliance_strategy.go`, `internal/diplomacy/imperial.go`, `internal/diplomacy/imperial_politics.go`, `internal/faction/faction.go`, `internal/game/game.go`

`MilitaryPowerBreakdown()` kara ve donanma ordularının etkin güçlerini ayrı
hesaplar; `MilitaryPower()` bu iki değerin toplamıdır. Etkin güç, birimin saldırı
değeri ile can yüzdesi ve ordu morali etkisini içerir. Diplomatik güç kıyasları
ve devlet güç sıralaması bu toplamı kullanır.

## İmparatorluk Sistemi

HRE, bağımsız prensliklerin doğrudan tek sahibi olarak değil, ayrı bir siyasi kurum
olarak `GameState.Imperial` içinde tutulur. Gerçek vassallar hâlâ `OverlordID` ve
`SameRealm()` üzerinden otomatik geçiş/koalisyon kurallarını kullanır; imparatorluk
prensleri ise `ImperialMember` kaydıyla bağımsız kalır.

`AssessImperialWarCall()` üyeleri ortak savaşa çağırırken otorite, sadakat, özerklik,
askerî bağlılık, sınır tehdidi, mevcut savaşlar ve kaynak durumunu birlikte değerlendirir.
Katılım üç sonuca ayrılır: tam savaşa katılım, sınırlı altın/tahıl desteği veya tarafsız
kalma. Bağımsız üye savaşa girerse normal savaş relation/ledger kayıtları oluşturulur;
sınırlı destek veren üye savaş relation'ına geçirilmez.

`ImperialState.Authority` ve `ImperialMember.Loyalty` Diyet ve seçim akışının save-backed
temelidir. `AdvanceImperialPolitics()` periyodik Diyet'i çalıştırır; `HoldImperialElection()`
üyelerin `ElectorWeight` değerleri ve adayların askerî/diplomatik gücüyle yeni imparatoru
belirler. 1300 senaryosunda elektör ağırlıkları Altın Ferman öncesi değişken tutulur;
senaryo verisi `data/imperial.json` dosyasındadır.

HRE oyuncusu, oyun içindeki `İmparatorluk` HUD düğmesi veya `I` kısayoluyla
`internal/render/imperial_panel.go` panelini açar. Panel otoriteyi, mevcut imparatoru,
Diyet/seçim takvimini ve üyelerin sadakat/özerklik/askerî bağlılık değerlerini gösterir.
Üye satırından mevcut canonical diplomasi hedefi açılır. HRE savaş ilanında savaş
önizlemesi bağımsız üyeleri `İmparatorluk Üyeleri / Çağrılabilir Müttefikler` başlığıyla
gösterir; tam katılım, sınırlı kaynak desteği ve tarafsızlık ayrı statülerdir.

HRE oyuncusunun Diyet ve seçim kararları `ImperialState.PendingDecision` içinde save-backed
bekler. Diyet için merkezî otorite, prenslik imtiyazları veya askerî katkı seçenekleri;
seçim için geçerli aday seçenekleri sunulur. Karar verilmeden oyuncu turu bitiremez.
Oyuncu HRE değilse mevcut deterministik AI Diyet/seçim çözümü korunur.

## İlişki Yapısı

```go
type Relation struct {
    FactionA, FactionB FactionID
    Score   int              // -100 (düşman) → +100 (müttefik)
    Stance  DiplomaticStance
}
```

`RelationKey(a, b)` → her zaman sıralı `"a|b"` string'i üretir (çift kayıt önler).

## Kuşatma Vassallığı

Aktif bir kara kuşatmasında kuşatılan devletin yalnız bir kara bölgesi kaldıysa
normal `propose_surrender` yerine bölge bağlı `propose_siege_vassalization`
kullanılır. Teklif iki yönde de üretilebilir: kuşatılan devlet kendi son toprağı
için vassal olmayı önerebilir veya kuşatan taraf vassallık isteyebilir. Kabul,
`ForceVassalizeAfterWar()` ile savaşı koalisyonlar arasında kapatır; bölge sahibi
değişmez, kuşatma temizlenir ve hedef devlet kuşatanın vassalı olur. Bu ayrı
aksiyon normal diplomasi elçi kotasını tüketmez, `RegionID` bazlı ret cooldown'u
ve teklif geçmişinde kendi sonucunu taşır.

Özgürleştirilen ardıl devlet, özgürleştiren oyuncu devletiyle doğrudan
`StanceAllied` relation'ı alır. Bu ilişki yeni devletin ilk diplomatik güvenlik
durumudur; `OverlordID` atanmaz, yani yeniden kurulan devlet bağımsız AI fraksiyonu
olarak davranır.

Savaş sonrası ardıl kararında `ForceReleaseAfterWar()` aynı barışı kapatıp ardılı
bağımsız ve doğrudan müttefik yapar. `Vassal Yap` seçeneği ise ardıl bölgeyi ardıl
fraksiyona devrettikten sonra `ForceVassalizeAfterWar()` ile `OverlordID` ve
vassallık turunu kaydeder. Bu kararlar diplomasi panelindeki mevcut `Vasallığı
Bitir / İlhak Et` yönetim aksiyonlarından ayrı tutulur.

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

Savaş ilanı `WarLedger`'a ilan eden ve savunan tarafı ayrı metadata olarak yazar.
Koalisyonun savunan tarafına sonradan katılan müttefikler de ilk ilan eden tarafın
yönünü değiştirmez; aktif savaş görünümünde ilan eden solda kalır
(`internal/diplomacy/war_call.go`, `internal/diplomacy/offers.go`).

| Duruş | Geçiş Koşulu | Puan Etkisi |
|---|---|---|
| `StancePeace` | Varsayılan / barış sonrası | Başlangıç 0; savaş sonrası -45 / -60 / -70 |
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
- İttifak için genel senaryolarda `Score >= 25`, `1300_ottoman_rise` senaryosunda ise `Score >= 40` gerekir; böylece aynı dinin varsayılan `25` puanı tek başına hemen ittifak açmaz. İki taraf arasında önceden değişmiş gerçek ilişki puanı, kara sınırı, fiili ticaret erişimi, ortak düşman veya ortak büyük tehditten en az biriyle diplomatik temas aranır. Buna ek olarak kara sınırı, fiili ticaret erişimi, ortak düşman veya ortak büyük tehditten en az biriyle gerçek coğrafi/stratejik bağ gerekir. Doğrudan sınır tehdidi mutlak blok değil, kabul şansını düşüren bir cezadır. Ortak düşman veya ortak büyük tehdit bu cezayı kısmen/tamamen telafi edebilir. Aynı din ayrıca doğrudan kabul şansı bonusu verir; yakın mezhep küçük bonus, sert mezhep ayrımı ise ek ceza üretir.
- İttifak asimetrik bir savaş çakışması oluşturamaz: teklif sahibi veya hedef devlet, diğer tarafın doğrudan müttefikiyle savaş halindeyse `allianceWarConflictBetween()` ittifak teklifini engeller. Bu kontrol oyuncu aksiyonunda, AI teklif üretiminde, doğrudan aksiyon geçidinde ve kuyruktaki teklifin kabulünde realm kökleri üzerinden tekrar kullanılır.
- Ticaret için `Score >= 15`, iki tarafın da kara bölgesi ve ortak helper ile hesaplanan toplam efektif ticaret kapasitesi `>= 4` olmalıdır. Pazar/liman/ambar/ibadethane ve ele geçirilmiş ticaret merkezi bu eşiğe katkı verir. Ana/ikincil merkez fethedildiğinde gelen `+2/+1` kapasite, mevcut anlaşmaların paylaştırılmış rota hacmini de sonraki ekonomi tick'inde artırabilir.
- Ticaret için aktif dış partner limiti (`4 + tam liman/pazar bölgeleri`) dolu olmamalıdır; tam maksimum liman ve pazar seviyesine sahip her sahip bölge `+1` partner verir. Bu limit yalnız teklif ekranında değil rota kurulumunda, aktif ilişki onarımında ve save/load temizliğinde de merkezi olarak zorlanır. Sıralama faction ID ile deterministiktir; limit aşan eski rota kayıtları aynı sırayla elenir. Doğrudan sınır tehdidi ise artık mutlak blok değil, kabul şansını düşüren bir cezadır.
- Ticaret anlaşması ayrıca bağlanabilir gerçek bir hat ister: ya iki realm arasında kesintisiz kara hattı, ya da her iki tarafta liman olup komşu deniz bölgeleri üzerinden bağlanabilen bir deniz hattı bulunmalıdır
- `StanceAllied` ile `TradeRoutes` artık ayrı kavramlardır; müttefiklik otomatik ticaret açmaz, ama müttefikken ayrıca ticaret anlaşması kurulabilir
- Zaten aynı duruştaysa tekrar kurulamaz; ancak `StanceTrade` duruşunda rota kaydı eksikse teklif akışı rotayı yeniden kurar
- Vassal-overlord bağı ayrı tutulur; iç realm relation'ları normalizasyonda `allied` çizgisine çekilir ve doğrudan overlord-vassal arasında kapasite/partner sınırından bağımsız iki yönlü ticaret rotası garanti edilir. İç-realm rota, dış partner sayısına ve dış rota kapasitesi paylaşımına girmez.

---

## Oyuncu Diplomatik Aksiyonları

`internal/game/game.go`

| Aksiyon | Fonksiyon | Koşul |
|---|---|---|
| Savaş ilan et | `declareWar()` | Zaten savaşta değilse; önce koalisyon önizlemesi açılır, hedefin vassalları ile iki tarafın çağrılabilir müttefikleri ve katılım ihtimali gösterilir |
| Barış teklif et | `proposePeace()` | Savaş halinde gerekli; 1300'de kalıcı savaş ledger'ı, objective, toprak/kayıp dengesi, süre, güç, ekonomi, çoklu savaş ve başkent tehdidi değerlendirilir. Diğer senaryolarda legacy savaş baskısı + güç + ekonomik stres modeli korunur. Teklif ekranı ve backend aynı `AssessPeaceProposal()` sonucunu kullanır |
| Heyet gönder | `improveRelations()` | `40` altın; savaş sırasında da gönderilebilir ve ilişkiyi duruşu bozmadan deterministik `+8` artırır |
| Hediye gönder | `sendGift()` | Savaşta değil + `120` altın; ilişkiyi deterministik `+15` artırır |
| İttifak kur | `proposeAlliance()` | Savaşta değil + genel senaryolarda `Score >= 25`, 1300'de `Score >= 40`; iki tarafın doğrudan müttefikleriyle mevcut savaş çakışması varsa oyuncu ve AI için teklif engellenir. Varsayılan din skorunun ötesinde diplomatik temas ve coğrafi/stratejik bağ gerekir. Kabul şansı ilişki puanı, doğrudan din uyumu bonusu, güç/bölge farkı, mevcut trade bağı, doğrudan sınır tehdidi cezası ve `ortak düşman / ortak büyük tehdit` bonuslarıyla değerlendirilir |
| Ticaret anlaşması | `proposeTrade()` | Savaşta değil + `Score >= 15` + iki tarafın da kara bölgesi ve yeterli ticaret kapasitesi var; ayrıca bağlanabilir kara/deniz ticaret hattı gerekir. Aynı helper kabul şansını ve UI'daki engel nedenini birlikte üretir |
| İttifakı bitir | `cancelAlliance()` | Dış devletle aktif ittifak varsa; mevcut ticaret rotaları korunur ve relation `trade/peace` durumuna iner |
| Ticareti bitir | `cancelTrade()` | Aktif ticaret rotası varsa; rotalar kaldırılır, mevcut ittifak korunur |
| Vassallık teklif et | `offerVassalization()` | Teklif eden zaten vassal değilse ve hedef başka devlete bağlı değilse; savaş duruşu teklifi göndermeyi engellemez, barışta mevcut `Score >= 55` ve askerî ön koşullar korunur |
| Vasallığı bitir | `releaseVassal()` | Yalnız oyuncunun doğrudan vassalında; devlet bağımsızlaşır, overlord ile ticaret anlaşması devam eder |
| Vassalı ilhak et | `annexVassal()` | Yalnız oyuncunun doğrudan vassalında ve onay sonrası; tüm bölgeler, kuvvetler, kaynaklar ve üretim emirleri oyuncuya devredilir, vassal fraksiyon elenir |

Teklifler artık otomatik kabul edilmez; oyuncu ve AI aynı değerlendirme motorunu kullanır.
Barışta vassallık için `Score >= 55` teklifin ilişki ön koşuludur; savaş sırasında bu
ön koşul teklif gönderimini pasifleştirmez, kabul kararı yine
`AssessVassalizationProposal()` ile değerlendirilir. Hedef, teklif sahibinin
doğrudan sınır tehdidi altında değilse askerî gücün ve kara bölgesinin en az `5x` olmasını
birlikte arar. Böylece ilişki puanı tek başına veya sınırlı üstünlükle uzak/zayıf hedefler
otomatik olarak vassal olmaz (`internal/diplomacy/vassalage.go:136`).

Barış oranı deterministik kabul eşiği üzerinden gösterilir; `AssessPeaceProposal()`
hedefin kabul edecek taraf perspektifini kullanır. Değerlendirme kabul etmiyorsa
oran `%100` olamaz; kabul ediyorsa `%100` kesin kabul anlamına gelir ve panel bunu
`Kesin kabul` olarak etiketler. Böylece renderer'daki yaklaşık bölge/ilişki formülü
ile `Execute()` içindeki gerçek barış kararı birbirinden ayrılmaz
(`internal/diplomacy/peace_assessment.go`, `internal/render/diplom.go`).

Bekleyen normal teklif kuyruğa alındığı anda gönderen fraksiyonun tur içi diplomasi
kotası tüketilir. Kuşatma bağlamındaki `propose_surrender` teklifi bu kotadan
istisnadır; oyuncu ve AI bu özel teklifi elçi hakkı doluyken de gönderebilir. Oyuncu
teklifi kabul ettiğinde `ResolveOffer()` güncel ilişki ve geçerlilik
koşullarını yeniden kontrol eder, ancak aynı teklif için kotayı ikinci kez tüketmez.
İttifak teklifinde kabul kararı ayrıca yeniden şans/strateji hesabına sokulmaz; teklif
sonrasında değişebilen ortak tehdit veya ticaret koşulları, oyuncunun geçerli teklifi
kabul etmesini tek başına engellemez. Savaş, elenme, aynı realm'e dönüşme veya teklif
taraflarından birinin mevcut müttefikiyle karşı cepheye düşmesi ise
teklifi geçersiz kılar.
Bu ayrım, gönderenin üçüncü ve son elçi hakkıyla oluşturduğu tekliflerin kabulde
yanlışlıkla uygulanamamasını önler. İlgili uygulama seam'leri
`internal/diplomacy/offers.go:ResolveOffer()` ve `internal/diplomacy/diplomacy.go:execute()`
ile regression testleri `TestResolveQueuedAllianceOfferDoesNotSpendQuotaTwice` ve
`TestResolveQueuedAllianceOfferKeepsTermsAfterStrategicStateChanges`'dır.

Savaş sırasında aynı oyuncuya hem barış hem de kuşatma teslimiyeti teklifi bekliyorsa
`BestOfferIndex()` barış teklifini teslimiyetten önce seçer; teslimiyetin daha yüksek
öncelik puanı bu sıra kuralını geçersiz kılamaz. Oyuncu barışı kabul ettiğinde
`setPeaceBetweenCoalitions()` aynı savaşın bekleyen `propose_surrender` kayıtlarını
temizler. Böylece teslimiyet yalnızca barış reddedildiğinde veya barış teklifi
oluşmadığında sonraki karar olarak gösterilir. Regression:
`TestPeaceOfferPrecedesSiegeSurrenderAndAcceptanceSkipsIt`.

Barış kabul edildiğinde `setPeaceBetweenCoalitions()` yalnız relation ve ateşkes
kayıtlarını güncellemez: iki koalisyonun birbirinin kara bölgesinde kalan tüm
orduları, hareket puanına bakılmadan `EvacuateArmiesFromPeaceTerritory()` ile en
yakın kendi, aynı realm veya resmî müttefik bölgesine çekilir. Terk edilen
kuşatmalar ile bu savaşa ait bekleyen kara/deniz temasları da aynı anda temizlenir.
Denizden başlayan kuşatmada önce nakliye filosuna yeniden binme tercihi korunur;
nakliye yoksa aynı güvenli bölge seçimi uygulanır. Regression:
`TestAcceptedPeaceEvacuatesLandArmyRegardlessOfMovePoints`.

Savaş ilanı artık `ExecuteWarDeclaration()` üstünden ayrı bir koalisyon akışı izler:

- Hedefin mevcut vassalları savaşa kesin katılır.
- Hem oyuncu hem AI tarafında dış müttefikler için ayrı `AssessWarCall()` değerlendirmesi yapılır.
- Oyuncu, savaş önizleme modalında hangi müttefiklerini çağıracağını checkbox ile seçer.
- Savaş ilanından önce hedefin doğrudan dış müttefikleri snapshot alınarak ilişki sonucu uygulanır. Tarafsız kalan her müttefikle ilişki `-25` düşer; saldıranın aynı zamanda müttefiki olan hedef müttefikiyle ittifak bozulur ve ilişki `-35` düşer. Savaşa katılan müttefiklerde bu ceza ikinci kez uygulanmaz; ilişki doğrudan `Savaş / -80` olur.
- Seçilip de çağrıya gelmeyen müttefiğin ittifakı bozulur; ilişki puanı `-10` düşer.
- Aynı deterministik helper savunan tarafın müttefikleri için de kullanılır; bu yüzden modalda görülen olasılık savaş resolve anındaki gerçek çağrı sonucuyla aynı kaynaktan beslenir.
- Barıştan sonraki beş turluk `PostPeaceTruceTurns` ateşkesi, doğrudan savaş ilanının yanı sıra müttefik veya imparatorluk savaş çağrılarında da uygulanır; devlet barış yaptığı düşmana karşı başka bir müttefikinin savaşına katılamaz (`AssessWarCall()`, `AssessImperialWarCall()`).
- Savaş ilanı uygulandığı anda render tarafında ayrı bir `Savaş Özeti` modalı açılır; burada gerçekten katılan coalition üyeleri, katılmayan müttefikler ve iki tarafın toplam askeri gücü gösterilir. Özet kapanmadan hareket, temas ve diğer savaş devam akışları çalıştırılmaz; özet kapandıktan sonra bekleyen normal aksiyon sürdürülür.

AI savaş ilanı öncesinde iki tarafın koalisyon gücünü projekte eder. Hedefin
vassalları, hedefin dış müttefikleri ve bu müttefiklerin vassalları savunma
tarafına eklenir. Saldıran AI tarafında yalnız `AssessWarCall().AutoJoin` ile
kesin katılacak dış müttefikler ve onların vassalları saldırı gücüne eklenir;
oyuncunun bekleyen savaş katılım teklifi kesin destek olarak sayılmaz.
Müttefik ordularının hedef kara bölgelerine rota mesafesi katkıyı azaltır;
yakın müttefik gücü minimum saldırı eşiğine veya saldırı gücüne tam ağırlıkla,
uzak kuvvet ise daha düşük ağırlıkla girer. Bu hesap `AI` karar katmanında
`internal/ai/war_strategy.go:aiWarCoalitionAssessment` ile yapılır; savaş
çözümündeki koalisyon katılımı ve vassal otomatik katılım kuralları değişmez.

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
| Bir devletin doğrudan müttefikine savaş ilanı | Diğer doğrudan müttefiklerle -25 |
| Mevcut müttefikin müttefikine savaş ilanı | İttifak bozulur, ilişki -35 |
| Barış | Savaş sonucuna göre -45 / -60 / -70; düşmanlık anında sıfırlanmaz |
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
- `AmountPerTurn`, iki tarafın toplam `trade_capacity` değerinden türetilir; temel
  anlaşma tavanı `4` olup gönderen ve alan devletin her maksimum seviyedeki pazar
  bölgesi için kendi tavanına `+2` eklenir
- Altın getirisi tur çözümlemesinde `TradeRoute.GoldEarned()` ile hesaplanır

---

## Başlangıç İlişkileri

`faction.BuildInitialRelations(factions)` — `internal/faction/loader.go`

Tüm fraksiyon çiftleri için skor `internal/religion.Relation()` sonucuyla başlatılır. Varsayılan duruş barıştır; Sünni-Şii çiftleri başlangıçta savaş durumuna alınır.

`1300_ottoman_rise` senaryo override'ı bu varsayılanı gerçek 1300 cepheleriyle düzeltir:
Osmanlı-Doğu Roma, Memlük-İlhanlı, Aragon-Kastilya, Aragon-Napoli, İngiltere-Fransa,
İngiltere-İskoçya ve Fransa-HRE savaşta başlar. Aragon-Granada müttefik kalır;
Venedik-Ceneviz ile Doğu Roma-Bulgaristan barışta bırakılır. İttifak eşiği ve
başlangıç sürtüşme puanları Kastilya-Portekiz, Leon-Kastilya,
Osmanlı-Karamanoğulları, Macaristan/Venedik ve benzeri çiftlerin hemen müttefik
olmasını engeller. Flandre, HRE'nin vassalı
olduğu için Flandre-Fransa düşmanlığı HRE-Fransa kök savaşıyla birlikte koalisyona
katılır; overlord-vassal arasındaki iç ticaret ve geçiş garantisi korunur. İlişki çiftleri
loader'da sıralı faction ID'leriyle üretildiğinden save/replay yönü deterministiktir.

---

## Barış sonucu ve ateşkes

`AssessPeaceSettlement` savaş skoruna göre barışı beyaz barış, bölge bırakma,
altın tazminatı veya vassallık olarak sınıflandırır. `ExecuteAIPeace` AI-AI
barışlarında sonucu uygular. Oyuncunun mevcut barış aksiyonu ise seçim yapılmadan
toprak veya altın kaybettirmemek için varsayılan olarak beyaz barıştır.

Barış sonrası taraflara beş turluk save-backed ateşkes verilir. Bu bilgi
`GameState.RecentTruces` içinde relation key ve bitiş turu olarak tutulur;
`ActionDeclareWar` ateşkes bitene kadar bloklanır.

Barış kabul edildiğinde koalisyon taraflarının denizden başlattığı aktif
`NavalLanding` kuşatmaları da kapanır. Kuşatan kara ordusu hedefe en yakın ve
toplam kapasitesi yeterli dost nakliye filolarına yeniden bindirilir; yeterli
nakliye yoksa en yakın kendi kara bölgesine çekilir. Bu tahliye oyuncu barışı,
bekleyen barış teklifi ve AI-AI barışı için ortak `setPeaceBetweenCoalitions()`
akışında uygulanır.

## AI Diplomasi Davranışı

`aiHandleDiplomacy()` ve `FormCoalitionAgainstPlayer()` — zorluk 3 koalisyon dahil aynı motoru kullanır

AI:

- Tüm senaryolarda savaşın ilk dört turunda normal barış denemez; objective tamamlanması, toprak ve
  birlik kaybı, süre/durgunluk, güç ve ekonomik stres, çoklu savaş ve başkent tehdidi
  barış baskısını belirler. Başkent tehdidi veya askerî çöküş erken kapıyı aşabilir;
  12 turu geçen ve son 8 turda muharebe/aktif kuşatma üretmeyen savaşlar otomatik
  `stalemate` kabul edilerek barış teklifine açılır; reddedilen teklif aynı savaşta üç
  tur cooldown'a girer
- `warObjectiveCompleted()` yalnız `expand` objective'ini savaş hedefi tamamlanması
  sayar. `defend` veya `consolidate` planı, savaşın kendisini bitirmiş gibi barış
  baskısını yanlışlıkla azaltamaz; uzun durgunlukta iki taraf da aynı stalemate kapısından
  barışı kabul edebilir
- `PeaceAssessment.WarScore`, actor perspektifinden türetilen `-100..100` arası savaş
  sonucudur. Fetih/kayıp ledger'ı, mevcut bölge değişimi, başkent tehdidi ve `expand`
  planındaki elde tutulan hedef bölgeleri birleştirir. `ObjectiveHeld` ve
  `ObjectiveTotal` değerleri hedef ilerlemesini ayrıca görünür kılar; bu skor save'e
  yazılmaz, her değerlendirmede güncel state'ten yeniden hesaplanır.
- `PeaceAssessment` ayrıca `WarExhaustion`, `GoldPressure`, `GrainPressure`,
  `SatisfactionPressure` ve `RelationshipPressure` alanlarıyla barış baskısının
  açıklamasını taşır. Savaş süresi/kayıpları, mevcut altın ve tahıl seviyesi, sahipli
  bölgelerin ortalama memnuniyet açığı ve savaş ilişkisinin negatif skoru ayrı raporlanır;
  bunlar save'e yazılmaz, güncel state'ten türetilir.
- `Faction.TerritorialClaims` içindeki core bölgeleri ve claim değerleri mevcut
  sahiplikle karşılaştırılır. Düşmanın elindeki core, acil durum yoksa barış
  kapısını kapatır; normal claim'ler barış eşiğini yükseltir. Aktif expand planı,
  savaş ledger hedefi ve strateji objective'lerinden materialize edilen claim'ler
  aynı değerlendirmeye katılır. Claim'in hedef devleti sabit değildir; AI bölgenin
  güncel sahibini savaş hedefi olarak kullanır.
- Base state kurulurken başlangıçta sahip olunan tüm kara bölgeleri otomatik core'dur.
  Strateji `territorial_claims` kayıtları ve objective içindeki `territorial_claims`/
  ilhak listeleri claim'e eklenir; claim bölgeye aittir, o anki sahibine değil. Bu yüzden savaşta
  yalnız başkent değil, devletin tanımlı hedef topraklarının tamamı barış kararını etkiler.
- Savaş kapandığında ilişki `-20` ile nötrlenmez: çözüme ulaşmış hedefte `-45`,
  normal beyaz barışta `-60`, çözümsüz core varken `-70` tabanı kullanılır.
  Barıştaki doğal `+1/tur` iyileşmesi ve ücretli heyet/hediye aksiyonları bu
  savaş hafızasını kademeli olarak onarır.
- ittifakta artık sadece `ortak düşman` sert filtresine bakmaz; aynı alliance assessment helper'ını kullanır ve `ortak büyük tehdit` gördüğünde de teklif açabilir
- AI dış ittifak açarken artık stratejik bağ, müttefik kapasitesi, `ai_expansion_targets` gerilimi ve hedefin somut katkısını da dikkate alır; ortak tehdit yoksa uzak/alakasız, tarihsel hedef olan veya büyük güç için gerçek askeri/stratejik fayda üretmeyen küçük devlete ittifak spam atmaz
- barışta skor ve bağlanabilir kara/deniz hattı uygunsa ticaret açar
- AI, kendi çıkarı olan ve genişleme hedefi olmayan barışçıl ilişkilerde ticaret/ittifak/güvenlik eşiğini yükseltmek için aynı `Heyet` ve `Hediye` aksiyonlarını kullanır; heyet `40` altın karşılığında `+8`, hediye `120` altın karşılığında `+15` verir ve hediyenin `80` altını alıcıya aktarılır. Bu harcamalar AI hazinesinde tahıl/kaynak tedariki ile araştırma, ekonomi, donanma ve ordu yatırımlarından sonra gelir; `1300_ottoman_rise` bütçesinde yalnız bu önceliklerden arta kalan altın kullanılabilir. Harcama kararı sonrasında deterministik `%60` başarı zarı atılır; başarısız zar teklif veya ödeme üretmez. Aynı fraksiyon turunda en fazla bir ilişki bakım aksiyonu yapar ve uygulanan bakım aksiyonu ilgili ilişkiyi dört tur cooldown'a alır; oyuncuya gönderilen teklif de bu korumayı kuyruğa girişte başlatır. AI-AI işlemleri anında, oyuncuya gönderilenler ise `DiplomaticOffers` kuyruğunda yalnız `Tamam` bildirimiyle çözülür.
- vassal durumundaki AI bağımsız diplomasi ve savaş değerlendirmesi yapmaz
- barış kabul edildiğinde taraf çifti `GameState.RecentTruces` içinde beş turluk
  save-backed ateşkes alır; bu sürede `ActionDeclareWar` engellenir, süre bitince
  savaş ilanı yeniden açılır. Koalisyon barışı taraf çiftlerinin her biri için aynı
  kaydı üretir
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
- Vassallık kabulünde `VassalizedTurn` kaydedilir; doğrudan overlord, bağın
  kurulmasından itibaren en az 12 tamamlanmış tur geçmeden vassalı ilhak edemez.
  Bekleme süresi dolana kadar aynı `ActionBlockReason()` kapısı yönetim kartını
  ve backend aksiyonunu pasif tutar.
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

Barış teklifi modalı kabulün ek sonucunu da açıkça belirtir: kabul edilirse
`PostPeaceTruceTurns` değerine göre beş tur boyunca teklif gönderen devlete
yeniden savaş ilan edilemez. Bu bildirim yalnız barış tekliflerinde gösterilir;
ittifak, ticaret ve savaşa katılım çağrılarında modal metni değişmez.

AI savaş ilanı sırasında oyuncu tarafında aktif bir ittifak varsa aynı kuyruk artık `join_war_call` tipiyle kullanılır:

- Müttefik AI başka bir devlete savaş ilan ettiğinde oyuncuya `Savaşa Katılım Çağrısı` gelir
- Oyuncu hedef devletle zaten savaş halindeyse bu çağrı kuyruğa alınmaz; eski/geçersiz bir çağrı da modal seçimine sokulmaz
- Bir AI oyuncunun müttefikine savaş ilan ettiğinde bu kez savunan müttefik oyuncuyu kendi safına çağırır
- Oyuncu kabul etmeden savaşa çekilmez; kabulde oyuncu realm'i ilgili düşman koalisyonuyla savaşa girer
- Reddederse çağrıyı yapan müttefikle ittifak bozulur ve ilişki puanı `-10` düşer
- Teklif AI turu içinde doğduysa sıra makinesi oyuncu cevabını bekler; kabul edilen çağrı aynı aktif AI deklaratörünün turunu kapatır

## Diplomasi Paneli

`internal/render/diplom.go`

- Panel iki adımdır: önce hedef devlet listesi, sonra teklif sayfası açılır.
- Hedef devlet listesi her satırda `Askeri güç` değerini `ordu/donanma` biçiminde, `Hazine` değerini `Gelir/Altın` biçiminde ve aktif devletler arasındaki `Güç sırası` (`X/Y`) değerini gösterir. Güç sırası bu iki askerî değerin toplamına göre hesaplanır. Listenin üstündeki `Alfabetik`, `İlişki`, `Güç Sıralaması` ve `Ekonomik Sıralama` düğmeleri listeyi sırasıyla varsayılan ID alfabetiğine, oyuncuyla olan ilişki puanı azalan düzene, standing sırası artan düzene veya brüt gelir azalan ve eşitlikte hazine azalan düzene göre yeniden düzenler. `İlişki` sıralamasında aynı ilişki puanına sahip hedefler içinde oyuncuyla kara sınırı paylaşan devletler önce gelir; kalan eşitlik faction ID'siyle çözülür.
- Oyuncunun oynadığı devlet de bu listede kendi güç/ekonomi sırasıyla görünür. Satır `Oyuncu / Kendi devletin` olarak gösterilir; kendi devleti teklif hedefi olmadığı için satır çift tıklaması teklif paneli açmaz.
- Teklif paneli artık çekirdek aksiyonların yanında `Heyet`, `Hediye` ve `Vassallık` seçeneklerini de gösterir.
- Teklif türü düğmelerine tıklanınca seçilen aksiyon `diplomacyActionFocus` ile korunur; seçili ve etkin düğme 2 px altın-sarı border ile vurgulanır, böylece `Teklif Gönder` öncesi hangi teklifin seçildiği net görünür.
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
- Dock edilmiş filo deniz rota ankrajı olan `RegionID` değerini korur; oyun mekaniğindeki kanonik konumu `LocationID()` ile `DockedSettlementID`, açık denize çıkınca ise deniz bölgesi ID'sidir. Render tarafı docked settlement anchor'ını kullanır
- İttifak biter ya da liman el değiştirirse `sanitizeDockedFleets()` bu bağı düşürür ve filoyu en yakın deniz bölgesine çıkarır

---

## Eksik / Planlanan

- [ ] Bekleyen diplomatik teklif kuyruğu / çok adımlı müzakere
- [ ] İttifak için ortak geçiş hakkı veya askeri bonuslar
- [ ] Ticaret için dinamik piyasa / rota pathfinding
- [x] `internal/religion` paketi ayrıştırıldı
