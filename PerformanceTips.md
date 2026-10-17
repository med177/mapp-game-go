# Performance Tips

Bu belge, Mapp'te oyun mantığı ve UI geliştirilirken performans regresyonlarını
önlemek ve ölçülebilir darboğazları güvenle iyileştirmek için proje kurallarını
toplar. Tur geçişi optimizasyonlarının geçmişi için
[`_PLANS/tur_hizlandirma_plani.md`](_PLANS/tur_hizlandirma_plani.md),
tur sırası ve scheduler davranışı için
[`wiki/architecture/game-loop.md`](wiki/architecture/game-loop.md) kaynak
kabul edilir.

## Değişiklikten önce ölç

- Kullanıcının bildirdiği gecikmenin hangi aşamada olduğunu ayır: tur başlangıç
  hazırlığı, pazar, faction AI kararları/hamleleri, `resolveTurn()`, kayıt ve
  ekranda ilk ilerlemenin görünmesi aynı maliyet değildir.
- Önce aynı senaryo ve aynı save ile başlangıç ölçümü al. Gerçek autosave
  profilinde save'i her tekrar öncesi yeniden yükle; aynı rastgele seed'i,
  faction sırasını ve oyuncu AI ayarlarını kullan. Save dosyasını değiştiren
  yan etkileri engelle ve ölçüm öncesi/sonrası dosya metadata'sını kontrol et.
- Aşama sürelerini ayrı ayrı kaydet ve en az üç tekrardan medyanı raporla.
  Headless profil oyun mantığının maliyetini gösterir; çizim, GPU, gerçek
  `Update` başına gecikme ve algılanan ilk-kare süresini ölçtüğü varsayılmamalı.
- CPU profilinde `flat` ve `cum` değerlerini ayır. Kümülatif maliyetler
  üst üste biner; yüzde veya süreleri toplayıp toplam darboğaz gibi sunma.
  Profilin işaret ettiği fonksiyonu çağrı yolu ve gerçek veriyle doğrula.
- Mikrobenchmark, sentetik yükte değiştirilen algoritmanın maliyetini ve
  allocation'larını kıyaslamak için uygundur; gerçek oyun turu ölçümünün yerine
  geçmez. Önce/sonra karşılaştırmasında aynı fixture, state ve komut
  kullanılmalıdır.
- İlk makul sonuçla yetinme: ölçümü tekrarla, beklenmedik dalgalanmayı incele
  ve sonucu kullanıcıya hangi ölçüm sınırlarının dahil olduğunu belirterek sun.

## Sıcak yollarda hesaplamayı paylaş

- AI turunda faction, bölge, ordu, diplomatik ilişki veya ticaret grafiğini
  aday başına tekrar tarama. Hesap aynı state için geçerliyse kapsamı açık olan
  bir snapshot/index üret ve o karar turunda yeniden kullan.
- Savaş yorgunluğunu realm başına toplu snapshot ile hesapla; ortak düşman ve
  kara/deniz sınırları gibi faction'lar arasında paylaşılan verileri yeniden
  keşfetme. Snapshot'ın savaş, sahiplik veya ilişki değişince geçerliliğini
  yitirip yitirmediğini belirle.
- Vergi, pazar talebi ve bina yatırımı gibi aynı faction kararlarında ortak
  `StrategicContext` ve memnuniyet hesaplayıcısını kullan. Context içindeki
  cache'in hangi veriyi tuttuğunu açıkça sınırla; değişebilir bölge vergisi ve
  kaynak miktarı gibi değerleri eski snapshot'tan okumamaya dikkat et.
- Karakter hedeflemesinde her aday için ordu listesi taramak yerine ihtiyaç
  varsa faction/bölge askerî güçlerini tek geçişte indeksle. Böyle bir indeks
  hiçbir aday/ödül için gerekmiyorsa kurulum maliyetini ödeme.
- Birbirine bağlı sonuçları aynı çağrıda hesapla. Örneğin ticaret merkezi
  kapasite ve gelir bonusları aynı hacim artışını kullandığından artışı bir kez
  hesapla; gelir hesabında efektif kapasiteyi üretirken aynı hacim taramasını
  yeniden tetikleme.
- Senaryo/config varsayılanlarını veya sabit tabloları sıcak çağrıda yeniden
  kurma. Tamamlanmış ve değişmez config'i tekrar kullan; kısmi config için
  varsayılan uygulama ve eski veri davranışını koru.
- Arama pahalıysa önce ucuz, anlamı değiştirmeyen filtreleri uygula. Örneğin
  faction askeri güç yeteneği taşımayan karakterleri ölüm, uygunluk ve görev
  kontrollerine sokmadan ele. Filtreyi yalnız ilgili etki/kanalda gerçekten
  katkısı olmayan değerleri dışladığını test ederek ekle.

## Cache, state ve deterministiklik

- Önce canonical domain helper'ını ve mevcut hesaplama akışını kullan; aynı
  kuralı AI, çözümleme ve UI için ayrı ayrı yeniden yazma.
- Cache eklemeden önce girdileri, ömrünü ve invalidation noktalarını tanımla.
  Sahiplik, yıl, görev, ölüm, sürgün, esaret, event flag'i, ordu ataması veya
  pazar emri değişebiliyorsa uzun ömürlü cache yanlış sonuç üretebilir.
- Tercihen çağrı/faction/tur kapsamlı read-only snapshot kullan. Daha uzun
  ömürlü cache gerekiyorsa her mutation yolunun invalidation'ını ve save/load
  sonrası davranışı kanıtla; türetilmiş runtime verisini gereksiz yere save
  şemasına ekleme.
- Performans değişikliği karar sırasını, tie-break'i, map iterasyon
  deterministikliğini, rastgele sayı tüketimini veya AI hamle sırasını
  değiştirmemeli. Harita/map sırasına güvenen yeni seçim ekleme; mevcut
  deterministik sıralamayı koru.
- Oyun state'ini değiştiren AI kararlarını paralelleştirme. Yalnızca salt-okunur
  ve girdileri sabitlenmiş ön hesaplamaları paralelleştirmeyi düşün; yarış,
  sonuç sırası, deterministiklik ve ana goroutine/UI sınırlarını ayrıca test et.
- Hızlı turda UI yakınlık taraması gibi yalnızca görsel olan ve hızlı modda
  gerekli olmayan işleri atlamak mümkündür; oyun kuralı veya karar sonucuna
  etkisi olmadığını doğrula.

## Tur geçişi ve UI yanıtlılığı

- Tur başlatma, pazar ve AI hazırlığını “Tur Bitir” input'unun senkron
  çağrısında bloklama. Önce yeni faz/durum görünür olsun; işi ana oyun
  döngüsündeki aşamalı state machine üzerinden ilerlet ve kullanıcıya aşama
  durumu bildir.
- Hızlı turda `Update` başına hem süre hem adım sayısı bütçesi kullanmak,
  değişken iş yüklerinde UI yanıtlılığını korur. Bir `Step()` bölünemezse tek
  adım süre bütçesini aşabilir; limiti mutlak üst sınırmış gibi belgeleme.
- Normal hızda hamlelerin görünürlüğünü ve mevcut bekleme/animasyon semantiğini
  koru. Hızlı mod optimizasyonları normal modun davranışını değiştirmemeli.
- Ağır işi goroutine'e taşımak otomatik çözüm değildir. State mutasyonu ve
  render ortaklaşıyorsa veri yarışları, tutarlı olmayan snapshot ve farklı
  karar sıraları doğurabilir. Ana döngüde küçük adımlara bölmek daha güvenli
  seçenek olabilir.
- `Update`/`Draw` içinde tekrar tekrar görüntü, dilim, map veya closure
  oluşturma; sabit geometriyi ve statik asset'leri cache'le. Cache'i yalnız
  invalidation koşulları doğruysa tut.

## Regresyon doğrulaması

- Algoritmik optimizasyonda önce/sonra sonuçlarını aynı fixture üzerinde
  karşılaştır. Hedef seçimi, bonus, gelir, kaynak miktarı, rol sahibi ve
  faction sırası gibi kullanıcıya görünen sonuçları assertion ile sabitle.
- İlgili değişebilir state için invalidation testleri ekle: en azından görev
  başladı/bitti, karakter öldü veya devre dışı kaldı, sahiplik değişti ya da
  yeni pazar/rota state'i oluştu durumlarından uygun olanları kapsa.
- Save/state davranışı etkileniyorsa eski kayıt varsayılanlarını ve
  serialize/deserialize akışını doğrula. Runtime-only optimizasyonda save
  şemasının değişmediğini gözet.
- Dar kapsamlı test ve benchmark sonrası, tur akışı veya ortak state helper'ı
  değiştiyse doğrudan etkilenen paket testlerini çalıştır. UI güncellendiyse
  ilgili editör tanılarını da kontrol et.
- Sonuçları aşama bazında, medyan ve yüzde farkıyla bildir; benchmark'ın
  sentetik/gerçek, headless/görsel ve hangi kayıt/faction ayarlarıyla
  çalıştırıldığını ekle. Doğrulanmamış kazancı vaat etme.

## Projede uygulanmış örnekler

- Savaş yorgunluğu, faction askerî güçleri ve stratejik kaynak gereksinimleri
  ortak snapshot/index ile tekrar taramaları azaltır.
- AI faction turunda memnuniyet hesaplayıcısı ve pazar hazırlığında üretilen
  `StrategicContext` sonraki AI kararında yeniden kullanılır; güncel ve
  değişebilir state değerleri gerektiğinde yeniden okunur.
- Konsey rol sahipleri faction başına tek karakter geçişinde seçilir; rol
  config'i ve konsey default'ları her iç sorguda yeniden hazırlanmaz.
- Ticaret merkezi hacim artışı kapasite ve gelir bonusu için paylaşılır.
- Pazar hazırlığı ve AI adımları aşamalı ilerler; hızlı tur `Update` başına
  bütçelenir. AI state'i değiştirdiğinden tam paralel faction AI uygulanmaz.
- Gerçek autosave profili, sentetik benchmark'lardan ayrı olarak bu
  iyileştirmelerin birikimli etkisini ölçmek için kullanılır. Geçmiş adımlar,
  koşullar ve ölçümler `_PLANS/tur_hizlandirma_plani.md` ve
  `wiki/systems/ai.md` içinde tutulur.
