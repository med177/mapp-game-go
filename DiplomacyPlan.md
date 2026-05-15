# Diplomasi V1 Yeniden Planı

## Özet

Mevcut `DiplomacyPlan.md` kısmen eskimiş durumda; diplomasi panelinde ilişki skoru ve duruş gösterimi zaten var. Yeni plan, mevcut sistemi tekrar anlatmak yerine gerçekten eksik olan üç alanı kapatacak: merkezi diplomasi kural katmanı, AI teklif/değerlendirme davranışı ve diplomasi-ekonomi entegrasyonu.

## Temel Yaklaşım

- Mevcut `declareWar/proposePeace/proposeAlliance/proposeTrade` mantığını `internal/game/game.go` içinde büyütmek yerine ortak bir diplomasi kural katmanına taşı.
- Bu katman teklif değerlendirmesi, ilişki skor güncellemesi, `stance` geçiş kuralları ve `TradeRoutes` senkronunu tek yerden yönetsin.
- `allied` duruşu ilk sürümde yalnızca diplomatik durum olarak kalsın; hareket, savaş veya bonus kuralları eklenmesin.
- UI tarafında yeni panel tasarımı yapılmasın; mevcut diplomasi paneli korunup yalnızca yeni sonuç mesajları ve gerekiyorsa durum etiketleri mevcut akışa uyarlansın.

## Uygulama Adımları

1. `internal/diplomacy` paketi ekle.
   - Ortak veri tipi ve helper'ları burada topla: teklif türü, karar sonucu, skor clamp, geçerli `stance` geçişleri, mevcut ilişkinin okunması/yazılması.
   - Oyuncu ve AI aynı karar motorunu kullansın; kurallar `game.go` ve `ai.go` içinde kopyalanmasın.

2. Teklif değerlendirme kurallarını tanımla.
   - `war` ilanı doğrudan uygulanır; skor `-80`.
   - `peace` teklifi yalnızca savaş halindeyken geçerli olsun; kabul için ilişki skoru, güç dengesi ve uzun süren savaş baskısı hesaba katılsın.
   - `trade` teklifi savaşta reddedilsin; kabul için negatif ilişki alt sınırı ve iki tarafın da en az bir kara bölgesine sahip olması yeterli olsun.
   - `alliance` teklifi savaşta reddedilsin; kabul için pozitif ilişki eşiği, taraflardan birinin diğerine doğrudan tehdit oluşturmaması ve iki tarafın da elenmemiş olması gereksinim olsun.
   - Kararlar deterministik olsun; bu sürümde rastgele diplomasi zarları eklenmesin.

3. Diplomasiyi ekonomiyle bağla.
   - Ticaret anlaşması imzalandığında `TradeRoutes` otomatik oluşturulsun; tekrar teklif mevcut rotayı çoğaltmasın.
   - Barış veya savaş durumuna geçildiğinde iki taraf arasındaki aktif ticaret yolları kapatılsın.
   - Tur çözümlemesindeki pasif ticaret geliri mevcut `TradeRoutes` üzerinden işlemeye devam etsin; ekstra "anlık altın ver" hack'i eklenmesin.
   - Rota üretimi başlangıçta basit kalsın: iki yönlü, sabit mallı, sabit miktarlı anlaşma; dinamik piyasa sistemi bu sprintte yok.

4. AI diplomasi davranışını ekle.
   - `ai.TakeTurn` başında veya sonunda diplomasi taraması yap.
   - AI savaşta kaldığı ve güç dengesi kötüleştiği hedeflerle barış arasın.
   - İyi ilişki + ortak tehdit gördüğü AI'larla ittifak arasın.
   - Barışta ve yeterli ilişki skorunda ticaret anlaşması arasın.
   - Zor mod koalisyon mantığı korunup yeni diplomasi motorunu kullanacak şekilde yeniden yönlendirilsin; özel-case relation yazımı bırakılmasın.

5. Dokümantasyon ve takip kayıtlarını güncelle.
   - `wiki/systems/diplomacy.md` yeni kabul/red kuralları, ticaret rotası üretimi ve AI davranışıyla güncellensin.
   - `wiki/systems/ai.md` diplomasi safhasını yeni akışla anlatsın.
   - `wiki/dev/progress.md` tamamlanan diplomasi maddelerini güncellesin ve "otomatik kabul" bilinen sorununu kapatsın.

## Arayüz ve Tip Değişiklikleri

- Yeni iç paket: `internal/diplomacy`.
- `GameState.Relations` ve `GameState.TradeRoutes` korunur; save formatı kırılmaz.
- Yeni iç tipler yalnızca uygulama içi olur; mevcut JSON şemasını bu sprintte değiştirme.
- Gerekirse `Relation` üzerinde save uyumluluğunu bozmayacak ek alanlar yerine runtime hesap tercih et.

## Test Planı

- Diplomasi kural testleri:
  - savaş dışı barış teklifi reddi
  - düşük ilişkiyle ittifak reddi
  - savaşta ticaret reddi
  - geçerli tekliflerde skor ve `stance` güncellemesi
- Ticaret entegrasyon testleri:
  - ticaret anlaşmasının tekil rota üretmesi
  - savaş/barış geçişinde rota kapanması
  - tur çözümlemesinde rota gelirinin altına yansıması
- AI testleri:
  - zayıf AI'nin savaşta barış araması
  - uygun ilişkide ticaret açması
  - koalisyonun yeni diplomasi motoru üzerinden ittifak kurması
- Regresyon:
  - mevcut save/load akışı `relations` ve `trade_routes` ile çalışmaya devam etmeli
  - `go test ./...` temiz geçmeli

## Varsayımlar

- Bu sprintte teklif ekranı, bekleyen diplomatik mesaj kutusu veya çok turlu müzakere sistemi eklenmeyecek.
- İttifakın askeri geçiş izni ve ortak savaş bonusu sonraki faza bırakılacak.
- Ticaret rotaları harita üstü pathfinding ile değil, soyut anlaşma modeliyle üretilecek.
- Mevcut `DiplomacyPlan.md` tamamen bu yeni planla değiştirilecek; eski "UI'de ilişki bilgisi yok" ve "ittifak/ticaret ikonları eksik" maddeleri kaldırılacak.

---

# Amfibi Savaş V1 Planı (Ayrı Plan)

## Özet

Bu plan diplomasi planından bağımsızdır ve yalnızca denizden düşman kıyısına çıkarma + çıkarma anı çatışma kurallarını kapsar.

## Kapsam

- Mevcut taşıma akışı (embark/disembark) korunur.
- Yeni faz: düşman sahipli kara bölgesine denizden çıkarma yapılabilsin.
- Çıkarma sırasında savaş çözümlemesi tek yerden, mevcut combat motoruyla çalışsın.
- Save formatı korunur; `GameState` JSON şeması kırılmaz.

## Uygulama Adımları

1. ✅ Çıkarma geçerlilik kurallarını netleştir.
   - Filo `EmbarkedUnits` taşıyor olmalı.
   - Hedef kara bölgesi filo ile komşu deniz bölgesinde olmalı.
   - Hedef düşmansa savaş hali zorunlu olmalı (`relations` kontrolü).

2. ✅ Çıkarma çatışması çözümünü ekle.
   - Hedefte düşman ordu varsa `combat.ResolveBattleWithMods` ile çöz.
   - Kazanırsa kara ordusu karaya iner ve bölge ele geçirilir.
   - Kaybederse çıkarma ordusu kaybedilir; filo denizde kalır.

3. ✅ Hedefte düşman ordu yoksa çıkarma işleyişini tanımla.
   - Taşınan birimler yeni kara ordusu olarak hedefte doğar.
   - Bölge düşmansa savaş hali varsa sahiplik güncellenir.
   - Filo cargo temizlenir, deniz bölgesinde kalır.

4. ✅ UI ve mesajları güncelle.
   - Mevcut hareket akışı korunur, ayrı buton eklenmez.
   - Çıkarma sonucu için net bilgi mesajları gösterilir (başarılı/başarısız).

5. ✅ AI entegrasyonu.
   - AI, savaşta olduğu kıyı hedefleri için filo çıkarma denesin.
   - Uygun olmayan hedeflerde (savaş yok, güç yetersiz) gereksiz çıkarma denemesin.

## Durum

- Genel durum: ✅ **Tamamlandı (V1)**
- Doğrulama: `go test ./...` temiz geçti.

## V2 Backlog (İyileştirme)

### P1 (Sonraki Sprint)

1. Deniz savaşı katmanı
   - Filo vs filo çarpışması (taşıma filosu korunma/kaçış kuralları).
   - Çıkarma öncesi deniz üstünlüğü kontrolü.

2. Çıkarma denge ayarları
   - Çıkarma turunda geçici saldırı/savunma modifiyeri.
   - Kıyı arazi tipine göre farklı çıkarma zorluğu.

### P2 (Orta Vadeli)

3. AI amfibi strateji derinliği
   - Çoklu filo koordinasyonu ve eşzamanlı çıkarma.
   - Uzak hedef puanlama (kritik liman, kutsal şehir, ekonomik merkez).

4. Lojistik ve sürdürülebilirlik
   - Çıkarma sonrası ikmal hattı kontrolü.
   - İkmal kesilirse attrition/moral cezası.

### P3 (Polish)

5. UI/telemetri geliştirmeleri
   - Çıkarma önizleme paneli (tahmini güç karşılaştırması).
   - Event log’da amfibi operasyon filtreleme etiketi.

## Test Planı

- Oyun kural testleri:
  - savaş hali yokken düşman kıyıya çıkarma reddi
  - düşman orduya karşı başarılı çıkarma
  - düşman orduya karşı başarısız çıkarma
  - düşman ordusuz kıyıya başarılı çıkarma ve sahiplik güncellemesi
- AI testleri:
  - savaşta uygun kıyıya çıkarma girişimi
  - barışta çıkarma girişimi yapmaması
- Regresyon:
  - mevcut embark/disembark dost/boş kıyı davranışı bozulmamalı
  - `go test ./...` temiz geçmeli

## Varsayımlar

- Bu fazda deniz savaşları (filo vs filo) genişletilmeyecek.
- Çıkarma için ekstra moral/ikmal sistemi eklenmeyecek.
- Çok turlu çıkarma hazırlığı yok; tek tur karar modeli korunacak.
