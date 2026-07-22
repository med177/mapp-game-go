---
type: system
tags: [combat, battle, terrain, casualties]
last_updated: 2026-07-22
related: [systems/ai, systems/economy, world/regions, systems/tech-tree, architecture/render-pipeline, architecture/state-management]
---

# Çarpışma Sistemi

**Kaynak:** `internal/combat/combat.go`

## Genel Bakış

Savunma kuşatması oyuncu ordusu veya kuşatılmış yerleşim seçildiğinde `Kuşatma Emri` panelinde görünür. `Huruç başlat` kuşatılan bölgede kuşatan orduyla kara muharebesi çözer; savunmacı kazanırsa aynı bölgede kalır, kuşatmayı kaldırır ve en az 1 hareket hakkını korur, kaybederse bölgede kalır ve kuşatma sürer. `Teslim ol` düğmesi yalnız kuşatan AI'ın oyuncuya gönderdiği bölge bağlı `propose_surrender` teklifi varsa aktifleşir. Kabul edilen teslimiyette savunma orduları mümkünse en yakın kendi bölgesine, 0 hareket ve -15 moral ile geri çekilir; son toprakta teslim olan AI, oyuncu tarafından kabul edilirse doğrudan vassal olur ve bölge yerel devlette kalır (`internal/game/{game.go,siege.go}`, `internal/render/action.go`, `internal/diplomacy/offers.go`).

Tüm çarpışmalar harita üzerinde otomatik hesaplanır — ayrı taktik sahne yok. Ordu bir düşman bölgesine hareket edince `ResolveBattleWithPlan()` veya `ResolveBattleWithContextPlan()` tetiklenir; oyuncu kara, deniz ve çıkarma saldırılarında önce duruş seçer. Tahkimli kara bölgelerde ise akış artık doğrudan fetih değildir: önce kuşatma veya genel hücum kararı gerekir. Resolve tamamlanınca oyuncu tarafında render katmanı ayrı bir savaş raporu modalı açar; burada sonuç, duruş ve tarafların `Güç / Birim / HP` önce-sonra kırılımı gösterilir.

---

## Hesap Akışı

```
saldıranGücü = ordu.TotalStrength(types) × (1 + atkMods.AttackMod + komutanSaldırıModu)
savunucuGücü = ordu.TotalStrength(types) × terrainBonus × (1 + defMods.DefenseMod + komutanSavunmaModu)

calculateOutcome(saldıranGücü, savunucuGücü)
    → (kazandıMı bool, saldıranKayıpOranı float64, savunucuKayıpOranı float64)

applyCasualties(ordu, kayıpOranı) → HP hasarı + gerekirse gerçek birim kaybı
Result → saldıran/savunan kayıp birimi + toplam HP hasarı
```

---

## Arazi Bonusları (Savunucu)

`terrainBonus()` — `internal/combat/combat.go:48`

| Arazi | Savunucu Çarpanı |
|---|---|
| Dağ | ×1.8 |
| Geçit | ×1.5 |
| Orman | ×1.3 |
| Kıyı | ×1.1 |
| Ova / Diğer | ×1.0 |

→ Arazi tipleri: [[world/regions]]

---

## calculateOutcome — Uygulama

`calculateOutcome()` — `internal/combat/combat.go`

±%15 rastgele zar dalgalanması içerir; zayıf ordu nadir de olsa kazanabilir.

```
dice  := rand.Float64()*2 - 1) * 0.15   // [-0.15, +0.15]
ratio := (atkStr / (defStr + 1)) * (1 + dice)
```

| Koşul | Sonuç | Saldıran Kayıp | Savunucu Kayıp |
|---|---|---|---|
| `ratio > 1.5` | Ezici Zafer | %10 | %80 |
| `ratio >= 1.0` | Dar Zafer | %35 | %50 |
| `ratio >= 0.7` | Geri Çekilme | %50 | %30 |
| `ratio < 0.7` | Ağır Yenilgi | %80 | %10 |

`outcomeDescription()` sonuca göre `"Ezici Zafer"`, `"Dar Zafer"`, `"Geri Çekilme"`, `"Ağır Yenilgi"` metin üretir.

---

## Teknoloji Modları

`TechMods{AttackMod, DefenseMod}` — `internal/combat/combat.go:10`

`game.techModsFor()` oyuncu/AI için tamamlanan teknoloji etkilerini toplar:
- `InfantryAttackMod + CavalryAttackMod + SiegeAttackMod` → `AttackMod`
- `LandDefenseMod` → `DefenseMod`

→ Teknoloji efektleri: [[systems/tech-tree]]

## Komutan etkisi ve kariyer

`internal/army/commander.go` içindeki `Commander`, doğrudan `Army.Commander` alanında
taşınır. Komutan bir savaşa katıldığında zaferde `+100 XP`, yenilgide `+40 XP` alır;
ilk zaferle seviye 2 ve `Savaş Tecrübesi`, 300 XP'de `Taktisyen`, 550 XP'de
`Savunma Uzmanı`, 850 XP'de `Saldırgan` trait'i açılır. Bu eşikler komutanın yaklaşık
9 zaferde maksimum seviyeye ulaşmasını sağlar.

Trait'ler `Army.CommanderModifiers()` ve ek komutan helper'ları üzerinden aynı
`battleStrengths()` hattına bağlanır. Böylece komutan etkisi hem gerçek resolve hem de
savaş öncesi preview'da aynı hesapla uygulanır. Bonus tavanı maksimum seviyede saldırı
için `%12`, savunma için `%10`, moral için `%13` olacak şekilde sınırlıdır. Yeni
uzmanlıklar mevcut trait'lerden türetilir:

| Trait | Etki |
|---|---|
| `Savaş Tecrübesi` | saldırı `%2`, savunma `%2`, moral `%8` |
| `Taktisyen` | saldırı `%4`, savunma `%2`, hareket `+1` |
| `Savunma Uzmanı` | savunma `%6`, moral `%5` |
| `Saldırgan` | saldırı `%6`, kuşatma ilerleme `+1`, gedik kazanımı `+1` |

Bu yüzden komutan etkisi artık sadece savaş gücüyle sınırlı değil: tur başı hareket
havuzu, kuşatma tick'lerindeki `BreachProgress` / gedik kazanımı ve moral tabanlı savaş
dayanıklılığı da aynı kariyer hattından beslenir. Birleşik savunmada XP, sanal birleşik
orduya değil savaşa katılan gerçek savunucu orduların komutanlarına yazılır.

Oyuncu fraksiyonu için başlangıçta üç kişilik komutan havuzu oluşturulur. Ordu detay
panelindeki `KOMUTAN ATA` / `KOMUTAN DEĞİŞTİR` aksiyonu, boşta olan komutanları gösteren
modalı açar; atama ve ayırma işlemleri `GameState.Commanders` havuzunu ve
`Army.Commander` bağlantısını birlikte günceller. Ordu birleşmesi, garnizon dönüşümü
ve save/load akışları da bu tekil atamayı korur. Kara ordusu nakliye filosuna bindiğinde
komutan `EmbarkedCommander` olarak korunur; çıkarma savaşı ve başarılı karaya çıkış
aynı kariyer bağlantısını kullanır. AI tur prelude'u ise her aktif saha
ordusu için (garnizon hariç) deterministik bir komutan üretip atar; deniz filoları da
naval savaşlara katıldıkları için aynı kariyer hattını kullanır. Savaş raporu ve olay
günlüğü detayı, katılan komutanların kazandığı XP'yi, seviye artışını ve yeni trait'leri
ayrı `Komutan gelişimi` satırlarında gösterir.

## Savaş Duruşları ve Bağlamı

`BattleStance` — `internal/combat/combat.go`

Oyuncu saldırı başlatırken üç duruştan birini seçer:

| Duruş | Etki |
|---|---|
| `Agresif` | efektif saldırı gücünü artırır; saldıran ve savunan kayıp oranlarını yukarı çeker |
| `Dengeli` | mevcut modelin nötr hali |
| `Savunmacı` | efektif saldırı gücünü düşürür; kayıp oranlarını azaltır |

Sistem artık üç ayrı savaş bağlamı tanır:

| Bağlam | Açıklama |
|---|---|
| `land` | klasik kara ordusu saldırısı |
| `naval` | iki donanmanın deniz bölgesindeki çatışması |
| `amphibious` | nakliye filosunun düşman kıyıya asker çıkartırken girdiği savaş |

Her bağlam kendi duruş çarpanını ve Türkçe özetini kullanır. Yani `Agresif` kara, deniz ve çıkarma savaşlarında aynı isimle görünse de etkisi birebir kopya değildir.

Bu seçim hem gerçek resolve hattında hem de saldırı öncesi preview panelinde aynı helper'larla hesaplanır:

- `battleStrengths()` efektif güçleri çıkarır
- `resolveOutcome()` gerçek zar sonucunu duruş çarpanlarıyla birleştirir
- `PreviewBattleWithContextMods()` zar aralığını tarayıp muhtemel sonuç, zafer şansı ve tahmini kayıp penceresi üretir

Preview tarafındaki kayıp özeti artık iki katmanlıdır:

- `HP` kaybı: orta ölçekli çatışmalarda birim ölmese bile hasarın görünmesini sağlar
- `Birim` kaybı: gerçekten düşmesi beklenen birlik sayısını verir

Preview, gerçek savaştakiyle aynı arazi/teknoloji/duruş/savaş tipi matematiğini kullanır; yani panelde görülen güç ve gerçek resolve birbirinden kopmaz.

Gerçek resolve çıktısı artık sadece `AttackerLost / DefenderLost` değil, `AttackerHPDamage / DefenderHPDamage` alanlarını da taşır. Bu veri kuşatma hücumu ve çıkarma dahil tüm oyuncu savaş raporlarında kullanılır; yani sonuç ekranındaki HP düşüşü preview tahminiyle aynı kavram üstünden gelir.

## Kuşatma Akışı

`internal/game/siege.go`

Tahkimli kara bölgesi (`fortress` settlement veya `walls` seviyesi) artık ayrı bir ara katman üretir:

1. Tahkimli kara bölgesine kuşatma başlatmak için artık orduda `siege` kategorili birlik zorunlu değil; normal kara orduları da aktif kuşatma kurabilir.
2. İlk temas anında oyuncu sağ tık sonrası `Kuşatma Kararı` modalında `Kuşatma Başlat` veya, kuşatma birimi varsa, `Genel Hücum` seçer.
3. `Kuşatma Başlat` anında ordu hedefe girmez; `GameState.Sieges` içine kayıt yazılır ve ordu hareketi biter.
4. Gedik ilerlemesi artık kuşatma ekipmanı tier'i ile kale seviyesi birlikte dikkate alınarak hesaplanır: yüksek tahkimatlar düşük tier araçlarla da zorlanabilir, ama ilerleme çok daha yavaş olur. Daha düşük tier araçlar kuşatmayı sürdürür, savunucuyu yıpratır ve ancak uzun sürede gedik üretir.
5. Her tur çözümlemesinde kuşatma baskısı savunucu orduya attrition uygular; gedik kapasitesi yetiyorsa ayrıca `BreachProgress` artar ve gedik seviyesi (`yok / küçük / büyük`) güncellenir.
6. Aktif kuşatma seçildiğinde renderer alt-ortada modal-dışı `Kuşatma Emri` paneli gösterir; oyuncu buradan `Genel Hücum` veya `Kuşatmayı Kaldır` seçebilir.
7. Aktif kuşatmaya aynı fraksiyon ya da müttefik fraksiyon destek için normal hareketle girebilir; bu giriş ayrı bir kuşatma başlatmaz. Destek orduları aktif bölgedeki canlı state'ten her kuşatma tick'inde taranır; sonradan getirilen kuşatma birimleri de `BreachProgress` ve gedik kazanımını güçlendirir. İlgisiz fraksiyonlar yeni kuşatma hamlesi yapamaz.
8. Kuşatmayı yapan ordu başka komşu bölgeye normal hareket emri alırsa bu hareket, eski kuşatmayı otomatik kaldırır; ayrı ikinci onay gerekmez.
9. Savunucu ordu çözülüp büyük gedik açılırsa tahkimat teslim olabilir; gedik açılamasa bile uzun aç bırakma kuşatması sonunda teslimiyet mümkün kalır.
10. Kuşatma birimi yoksa da `Genel Hücum` yapılabilir; ancak gedik yoksa tahkimat doğrudan ele geçirilemez ve kale bekleme / aç bırakma / teslimiyet hattında kalır.
11. `Genel Hücum`, gedik yokken tahkimatı doğrudan düşüremez; saldırı savunucuyu yıpratsa bile kale elde tutulur. En az küçük gedik açıldıysa hücum fetih üretebilir.
12. Genel hücumda saldıran taraf ayrıca sur tırmanışı ve dar giriş baskısı kaynaklı ek zayiat alır; bu bedel gedik yokken en ağır, küçük gedikte daha düşük, büyük gedikte en düşüktür.
13. Zaten başka bir devlet tarafından kuşatılmış tahkimli bölgeye üçüncü devlet yeni kuşatma başlatamaz; ancak bölgeye giriş hakkı varsa kuşatma yapan düşman orduya karşı savaş açabilir ve o ordu yenilirse kuşatma kalkar. Böyle bir savaş allied / same realm geçişinde gerçekleşiyorsa sahiplik değişmez, yalnız kuşatma kaldırılır.
14. Kuşatılan bölgede bölge sahibine veya onun müttefikine ait bir kara ordusu varsa, bu ordu komşu bölgeye çıkmadan önce kuşatan orduya karşı huruç savaşı yapmak zorundadır. Huruç kazanılırsa kuşatma kalkar ve uygun dost/sahipsiz hedefe ilerlenir; kaybedilirse kalan birlikler kuşatılan bölgede kalır ve hareket puanı tükenir.
15. Saldıran oyuncu da kuşatma panelindeki `Teslimiyet Teklifi` düğmesiyle AI savunmacıya çağrı gönderebilir. AI kabulü kuşatma baskısı, gedik, süre ve güç dengesine göre çözülür; bu özel kuşatma teklifi oyuncu veya AI'ın normal elçi kotasını azaltmaz.
16. AI kuşatan, kuşatma baskısı ve gedik ilerlemesi yeterli olduğunda oyuncuya teslimiyet talebi gönderebilir; AI savunmacı da ağır baskı veya son toprak koşulunda oyuncuya teslim olmayı teklif edebilir. Teklifler `DiplomaticOffers` içinde bölge kimliğiyle saklanır, modal ve kuşatma paneli aynı çözümleyiciyi kullanır.

Kuşatma hücumunda savunana arazi bonusuna ek olarak tahkimat savunma çarpanı uygulanır. Gedik büyüdükçe bu bonus düşer; yani surlar kırıldıkça saha savaşı normal kara muharebesine yaklaşır. Aynı anda saldıranın ekstra hücum zayiatı da azalır; küçük gedik hâlâ pahalı bir baskınken büyük gedik daha düşük bedelli bir yarma fırsatı sayılır.

---

## Hasar ve Toparlanma

Çarpışma sonucu artık ilk etapta tüm kaybı doğrudan birim silerek çözmez. `applyCasualties()` toplam HP havuzu üstünden hasarı dağıtır:

- orta kayıplarda birimler hayatta kalır ama hasarlı çıkar,
- ağır kayıplarda bazı birimler tamamen düşer,
- yaşayan birimlerin savaş katkısı `army.TotalStrength()` ve `TotalDefense()` içinde mevcut HP oranıyla ölçeklenir.

`internal/game/resolution.go`

- kara orduları sahip oldukları kara bölgede tur başına ücretsiz `+10 HP` toparlanır; ekonomi tick'inde depo kapasitesini aşan tahılla ordu başına ek `+10 HP`'ye kadar yenileme yapılabilir (`1 HP = 1 tahıl`),
- toparlanma `CurrentHP < 100` olan birimlerde çalışır ve `%100`e kadar sürer,
- kış turunda önce attrition uygulanır, aynı sweep içinde ek ücretsiz toparlanma verilmez,
- donanmalar ve dost olmayan topraktaki kara orduları bu akıştan yararlanmaz.
- aktif kuşatma altındaki savunmacı ordular, bölge sahibi kendi toprağı olsa bile toparlanmaz; `GameState.IsArmyDefendingSiegedRegion()` hareket, iyileşme ve UI tarafındaki ortak predicate'tir.

## Savaş Sonrası Uygulama

`internal/game/game.go`

```
if saldıranKazandı:
    düşman ordusu → temizlendi (birimsiz kaldıysa)
    saldıran ordu → hedef bölgeye taşındı
    targetRegion.ApplyConquest(ownerID, religion) → sahiplik değişti
    a.MovePoints--
else:
    saldıran ordu → yerinde kaldı (birimsiz kaldıysa silinir)
```

`ApplyConquest` bölgeyi yeni fraksiyona devreder ve din dönüşüm sayacını başlatır.

## Deniz Hareketi ve Çatışma Kuralı

Donanmalar deniz bölgeleri arasında savaş ilanı olmadan serbest hareket eder.
Denizde çatışma yalnızca iki fraksiyon arasında `StanceWar` varsa tetiklenir; barış/ittifak/trade durumunda aynı deniz bölgesine girildiğinde savaş açılmaz.
Savaş varsa oyuncu gemi-gemi çatışmasında da önce duruş seçer. Düşman kıyıya çıkarma sırasında savunan ordu varsa aynı modal `Çıkarma Muharebesi` başlığıyla açılır ve seçilen duruş `ActionDisembarkArmy` üzerinden oyun katmanına taşınır.

---

## Birim Gücü

`army.TotalStrength(types)` — her birimin `Attack + Morale/10` değeri mevcut HP oranıyla ağırlıklandırılır ve ordu `Army.Morale` çarpanından geçirilir. Ordu morali 100'de nötrdür; 50 moral toplam gücü yaklaşık %15, minimum moral ise yaklaşık %30 azaltır. Tahıl arzı bu state'i ekonomi tick'inde günceller; eski save'lerde eksik moral alanı 100 kabul edilir.

→ Birim tipleri: `assets/data/units.json`
