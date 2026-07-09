---
type: system
tags: [diplomacy, relations, stance, faction]
last_updated: 2026-07-09
related: [world/factions, systems/ai, architecture/state-management]
---

# Diplomasi Sistemi

**Kaynak:** `internal/diplomacy/diplomacy.go`, `internal/faction/faction.go`, `internal/game/game.go`

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

---

## Diplomatik Duruşlar (DiplomaticStance)

| Duruş | Geçiş Koşulu | Puan Etkisi |
|---|---|---|
| `StancePeace` | Varsayılan / barış sonrası | Score = -20 |
| `StanceWar` | Savaş ilan edildiğinde | Score = -80 |
| `StanceTrade` | Ticaret anlaşması | Score +15 |
| `StanceAllied` | İttifak | Score +20 |

Diplomatik duruşların görünen adları, badge metinleri ve editörde dolaşım sırası `internal/faction/stance_metadata.go` içinde merkezileştirilmiştir (`DiplomaticStanceLabelTR`, `DiplomaticStanceBadgeTR`, `AllDiplomaticStances`, `NextDiplomaticStance`).

**Geçiş kısıtları:**
- Savaştayken ittifak veya ticaret kurulamaz
- İttifak için `Score >= 20` gerekir
- Ticaret için `Score >= 10`, iki tarafın da kara bölgesi ve toplam `trade_capacity >= 4` olmalıdır
- Ticaret için aktif partner limiti (`4`) dolu olmamalıdır; doğrudan sınır tehdidi ise artık mutlak blok değil, kabul şansını düşüren bir cezadır
- `StanceAllied` ile `TradeRoutes` artık ayrı kavramlardır; müttefiklik otomatik ticaret açmaz, ama müttefikken ayrıca ticaret anlaşması kurulabilir
- Zaten aynı duruştaysa tekrar kurulamaz; ancak `StanceTrade` duruşunda rota kaydı eksikse teklif akışı rotayı yeniden kurar
- Vassal-overlord bağı ayrı tutulur; iç realm relation'ları normalizasyonda `allied` çizgisine çekilir

---

## Oyuncu Diplomatik Aksiyonları

`internal/game/game.go`

| Aksiyon | Fonksiyon | Koşul |
|---|---|---|
| Savaş ilan et | `declareWar()` | Zaten savaşta değilse |
| Barış teklif et | `proposePeace()` | Savaş halinde gerekli; kabul için savaş baskısı + güç dengesi + ekonomik stres değerlendirilir, teklif sahibi `peace_relation_bonus` tech etkisi bu eşiği destekler |
| Heyet gönder | `improveRelations()` | Savaşta değil + `40` altın; ilişkiyi deterministik `+8` artırır |
| Hediye gönder | `sendGift()` | Savaşta değil + `120` altın; ilişkiyi deterministik `+15` artırır |
| İttifak kur | `proposeAlliance()` | Savaşta değil + `Score >= 20` + doğrudan sınır tehdidi olmamalı |
| Ticaret anlaşması | `proposeTrade()` | Savaşta değil + `Score >= 10` + iki tarafın da kara bölgesi ve yeterli ticaret kapasitesi var; aynı helper kabul şansını ve UI'daki engel nedenini birlikte üretir |
| Vassallık teklif et | `offerVassalization()` | Teklif eden zaten vassal değilse + hedef başka devlete bağlı değilse + `Score >= 55` + belirgin askeri/bölgesel üstünlük varsa |

Teklifler artık otomatik kabul edilmez; oyuncu ve AI aynı değerlendirme motorunu kullanır.

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
| `ApplyRelationDecay()` | Savaşta skor düşer; barış/ticaret/ittifakta ilişki yumuşar |
| Ortak düşman | +bonus (AI koalisyon mantığında) |
| Din bonusu/cezası | `religion.Relation(a,b)` — başlangıç skoru; +25 / -20 / -30 / -40 |

→ `applyRelationDecay` tur çözümleme sırası: [[architecture/game-loop]]

---

## Ticaret Entegrasyonu

`GameState.TradeRoutes` artık diplomasi motoru tarafından yönetilir.

- Ticaret anlaşması kabul edilince iki yönlü rota oluşturulur
- Aynı iki fraksiyon için rota çoğaltılmaz; mevcut çift önce temizlenir
- İttifak duruşu ticaretten bağımsızdır; müttefik iki devlet arasında rota varsa `StanceAllied` korunur
- Savaş ilanı veya barış kabulü iki taraf arasındaki aktif rotaları kapatır
- Save/load veya eski kayıt migrasyonu sırasında elenmiş fraksiyon, geçersiz relation veya duplicate rota kalmışsa `SanitizeTradeRoutes()` bunları temizler; böylece yıkılmış devlet adları trade listesine geri sızmaz
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

---

## AI Diplomasi Davranışı

`aiHandleDiplomacy()` ve `FormCoalitionAgainstPlayer()` — zorluk 3 koalisyon dahil aynı motoru kullanır

AI:

- uzun savaşta ve zayıf kaldığında barış dener
- ortak düşman + yeterli skor varsa ittifak dener
- barışta ve skor uygunsa ticaret açar
- vassal durumundaki AI bağımsız diplomasi ve savaş değerlendirmesi yapmaz
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

## Diplomasi Paneli

`internal/render/diplom.go`

- Panel iki adımdır: önce hedef devlet listesi, sonra teklif sayfası açılır.
- Teklif paneli artık çekirdek aksiyonların yanında `Heyet`, `Hediye` ve `Vassallık` seçeneklerini de gösterir.
- Hedef listesi artık panel gövdesi üzerinde mouse wheel ile kaydırılır; scroll sadece dar satır alanına değil panel bağlamına da bağlıdır.
- Hedef listesinde seçim `mouse down` anında değil, kısa click release anında kesinleşir; basılı tutup sürüklemek listeyi satır yüksekliği bazında kaydırır ve yanlışlıkla devlet seçmez.
- Liste kartları fraksiyon rengi accent şeridi, ilişki/duruş özeti ve görünür scrollbar ile çizilir; teklif sayfası aynı UI compose ailesindeki kart/panel çerçevesini kullanır.
- Teklif sayfasında `Hedef`, `Durum` ve `İlişki Skoru` blokları ayrı iç padding ile çizilir; vassal veya overlord hedeflerinde durum satırı düz stance yerine hiyerarşi etiketi gösterir. Alt footer butonları da ortak ikonlu `Button` primitive'inin `TextOffsetY` hizasına bağlandığı için `Geri` ve `Teklif Gönder` satırı artık icon/metin kaydırması üretmez.

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
