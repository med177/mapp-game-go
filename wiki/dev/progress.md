---
type: dev
tags: [progress, status, todo, known-issues, next-steps]
last_updated: 2026-07-22
related: [HOME, architecture/game-loop, architecture/state-management, architecture/render-pipeline, systems/victory]
---

# Geliştirme Durumu

- 2026-07-22: Tahıl ekonomisi yeniden dengeleme planı tamamlandı. Kasım ayındaki stabil rezerv fazlası nüfus yatırımına bağlandı; nüfus artışı sonraki tick'lerde sivil tahıl talebini artırıyor. Ordu morali save-backed `Army.Morale` alanına eklendi; stabil/uyarı/kritik/kıtlık seviyeleri sırasıyla `+1/-1/-3/-6` moral etkisi uyguluyor ve `TotalStrength()` üzerinden savaş/AI gücüne yansıyor. HUD/event detayında moral delta görünür; eski save'ler 100 başlangıç moraliyle uyumlu. Doğrulama: `go test ./...`.

- 2026-07-22: Tahıl ekonomisi Faz 5 tahıl ithalatı/stratejik talep alt adımı tamamlandı. Fraksiyonun üç aylık rezerv hedefine kalan açığı `StrategicDemand` olarak hesaplanıyor; tahıl piyasa fiyatına ek talep olarak yazılıyor, Pazar panelinde hedef bazında gösteriliyor ve yeni ticaret rotasında kaynakta rezerv fazlası varsa rota tahıla yönlendiriliyor. İthalat mevcut rota transferi üzerinden kaynak stok/altın koşullarıyla sınırlanıyor. Kapsam: `internal/{economy/economy.go,game/{game.go,resolution.go},state/state.go,diplomacy/diplomacy.go,render/trade.go}`; doğrulama: hedefli testler.

- 2026-07-22: Tahıl ekonomisi Faz 5 olay bağlama alt adımı tamamlandı. Kıtlık ve kötü hasat olayları aktif bölge süresince tahıl üretimini azaltıp sivil talebi artıran yüzde modifiyerleri taşıyor; kuraklık/hasat event adları aynı sözleşmeyle destekleniyor. Ekonomi tick'i, bölgesel ikmal, stratejik talep ve AI hesapları ortak `GameState` yardımcılarını kullanıyor. Mevcut düşman savaş gemisi tabanlı abluka ticaret ve liman ikmal kesintisiyle birlikte korunuyor. Kapsam: `internal/{events/events.go,state/state.go,game/resolution.go,ai/{ai.go,movement_strategy.go}}`, `assets/scenarios/*/data/events.json`; doğrulama: `go test ./internal/state ./internal/events ./internal/game ./internal/ai`.

- 2026-07-22: Tahıl ekonomisi Faz 6 denge ve kapanış adımı tamamlandı. 1300 senaryosu için 24 tur × 2 seed erken/orta/savaş tahıl raporu ve `1.0–4.0` üretim/sivil talep, `-1.0–2.5` net değişim/sivil talep kabul bantları eklendi; kıtlık oranı teşhis metriği olarak raporlanıyor. AI ve oyuncu lojistiği ortak `RegionMilitaryGrainProduction()` ve `EffectiveArmyGrainUpkeep()` seam'lerine geçirildi. Yeni tüketim modeli sonrası 42 tur × 4 seed altın bantları güncellendi; deterministik iki turluk replay ve `go test ./...` başarılı. Kapsam: `internal/{state/state.go,game/scenario_balance_test.go,ai/grain_economy_test.go}`, `wiki/{systems/{economy,ai}.md,architecture/state-management.md}`.

- 2026-07-21: Tahıl ekonomisi Faz 5 ordu yenileme alt adımı tamamlandı. Mevcut ücretsiz dost-toprak toparlanmasına ek olarak ekonomi tick'inde yalnız depo kapasitesi üzerindeki tahıl kullanılıyor; dost ve kuşatma dışı kara orduları faction/army ID sırasıyla ordu başına en fazla `+10 HP` alıyor. `1 HP = 1 tahıl`, rezerv kapasitesi korunuyor ve yenileme HP/tahıl metrikleri runtime raporlanıyor. Kapsam: `internal/{game/{resolution.go,game.go,resolution_test.go},state/state.go}`; doğrulama: `go test ./internal/game ./internal/state`.

- 2026-07-21: Tahıl ekonomisi Faz 5 otomatik ihracat alt adımı tamamlandı. Pazar paneline save/load ile korunan `Oto. İhracat` toggle'ı eklendi; açıkken kapasite üzeri tahıl aktif ve savaşta olmayan ticaret ağı partnerlerine faction ID sırasıyla güncel fiyatın `%60`'ı üzerinden satılıyor, alıcı altını yetersizse miktar sınırlanıyor. Runtime raporu otomatik ihraç edilen tahıl/altını gösteriyor. Kapsam: `internal/{economy/economy.go,state/{state.go,grain_aid_test.go},game/{resolution.go,game.go},save/compact.go,render/{action.go,trade.go,trade_test.go,ui_geometry_test.go}}`; doğrulama: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi Faz 5 acil satış alt adımı tamamlandı. Pazar sekmesine partner gerektirmeyen `ACİL TAHIL SAT` butonu eklendi; yalnızca depo kapasitesi üzerindeki tahıl satılıyor, güncel piyasa fiyatının `%70`'i uygulanıyor ve rezerv tabanı korunuyor. Kapsam: `internal/{economy/economy.go,state/state.go,game/game.go,render/{action.go,trade.go,trade_test.go,ui_geometry_test.go}}`; doğrulama: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi Faz 5 tahıl yardımı alt adımı tamamlandı. Kendi bölgesinin aksiyon bandındaki `Tahıl Yardımı` düğmesi 12 tahıl karşılığında memnuniyeti +10 artırıyor; bölge başına turda bir kez, kuşatma/başka sahip/yetersiz stok/yüksek memnuniyet kontrolleri state katmanında uygulanıyor. Kapsam: `internal/{state/{state.go,grain_aid_test.go},game/game.go,render/{action.go,panel.go,hover_tooltip.go,renderer_input.go,region_panel_activity_test.go}}`; doğrulama: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi Faz 5 nüfus büyümesi alt adımı tamamlandı. Kasım ekonomi tick'inde stabil rezervin depolama kapasitesi üzerindeki kısmı, memnuniyeti en az 60 olan ve kuşatma/isyan riski taşımayan bölgelere yıllık nüfus yatırımı olarak harcanıyor; nüfus artışı %1 (minimum 1), maliyet nüfus başına 2 tahıl. `GrainEconomyStatus` büyüme ve harcama metriklerini runtime olarak taşıyor. Kapsam: `internal/{game/{resolution.go,game.go,resolution_test.go},state/state.go}`; doğrulama: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi Faz 4 abluka alt adımı tamamlandı. Rota uçlarındaki denizlerdeki düşman savaş gemilerinden türetilen runtime `BlockadePercent` ticaret hacmini azaltıyor; limanlı bölgesinin yerleşim/rezerv ikmal tamponu da aynı kesintiyi alıyor. Save formatına yeni kalıcı alan eklenmedi. Kapsam: `internal/{economy/economy.go,state/merchant_trade.go,game/{resolution.go,game.go},ai/{ai.go,movement_strategy.go}}`; doğrulama: `go test ./internal/economy ./internal/state ./internal/game ./internal/ai`.

- 2026-07-21: Tahıl ekonomisi Faz 4 askeri ikmal alt adımı tamamlandı. Efektif ordu tahıl bakımı sabit `%100`, hareket `%150`, garnizon `%75`, kuşatma savunması `%125`, kuşatma saldırısı `%200` katsayılarıyla ortak `GameState.EffectiveArmyGrainUpkeep()` hesabına taşındı. Bölgesel kapasite artık sivil talep sonrası üretim fazlasını kullanıyor; yabancı bölgede yerel üretim desteği kesiliyor. AI hareket/geri çekilme/birleşme/bina yatırım lojistiği aynı hesabı kullanıyor. Runtime hareket takibi save formatını değiştirmiyor. Doğrulama: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi Faz 3 ambar alt adımı tamamlandı. Tüm senaryo `buildings.json` dosyalarına `granary` (`Ambar`) tanımı eklendi; kurulu seviye başına +100 tahıl depolama kapasitesi veriyor. Mevcut sprite seti korunarak çiftlik görseli fallback'i kullanılıyor ve bina tooltip'inde kapasite bonusu gösteriliyor. Kapsam: `internal/{city/building.go,game/resolution.go,render/{panel.go,hover_tooltip.go}}`, `assets/scenarios/*/data/buildings.json`; hedefli test: `go test ./internal/game ./internal/render`.

- 2026-07-21: Tahıl ekonomisi Faz 3'ün ilk iterasyonu tamamlandı. Depolama kapasitesi `6 × sivil talep + 3 × ordu bakımı`, minimum 100 tahıl olarak hesaplanıyor; kapasite üstü stok %2 yumuşak bozuluyor. HUD `stok / kapasite` gösteriyor, runtime-only alanlar save migration gerektirmiyor. Kapsam: `internal/{state/state.go,game/{resolution.go,game.go},render/panel.go}`; test: `go test ./internal/game ./internal/render`.

- 2026-07-21: Tahıl ekonomisi Faz 2'nin ilk iterasyonu tamamlandı. `GrainEconomyStatus` ile fraksiyonun toplam tahıl talebine göre stok-ay seviyesi hesaplanıyor; 3 ay altı uyarı, 1 ay altı kritik, açık oluşması kıtlık kabul ediliyor. Gelir/memnuniyet etkileri, deterministic açlık HP dağılımı, event log ve HUD'da `stok / ay` görünürlüğü eklendi. Kapsam: `internal/{state/state.go,game/{resolution.go,game.go},render/panel.go}`; test: `go test ./...`.

- 2026-07-21: Tahıl ekonomisi yeniden dengeleme planı `_PLANS/Grain_Economy_Rework_Plan.md` altında oluşturuldu. Faz 1 kapsamında pozitif nüfuslu bölgelerde `ceil(population / 20)` sivil tahıl tüketimi `applyEconomyTick()` akışına eklendi; sıfır nüfus legacy/test bölgeleri tüketim üretmiyor. Regression: `TestCivilianGrainDemandRoundsUpAndIgnoresEmptyPopulation`, `TestApplyEconomyTickConsumesCivilianGrain`; doğrulama: `go test ./...`.

- 2026-07-21: Ordu birim hover popup'ı recruit verilerinden ayrıştırıldı. Ordu kartlarında üretim maliyeti ve teknoloji/gereksinim satırları kaldırıldı; aynı tipin ordu içindeki birlik adedi, tur başı tahıl bakımı, savaş değerleri ve gerçek `CurrentHP` gösteriliyor. Kapsam: `internal/render/{army_panel.go,hover_tooltip.go}`, `wiki/architecture/render-pipeline.md`; test: `go test ./internal/render`.

- 2026-07-21: Ordu bilgi panelindeki oyuncu birim kartlarına hover ile recruit ekranındaki ortak birim detay popup'ı eklendi. Kart hit-test'i kategoriye göre görünen sıra ile hizalı; düşman ordularında gizli birim detayları açılmıyor. Kapsam: `internal/render/{army_panel.go,army_panel_test.go,hover_tooltip.go,renderer.go}`, `wiki/architecture/render-pipeline.md`; test: `go test ./internal/render`.

- 2026-07-21: Ordu detay panelinin işaretli alt bilgi bandına oyuncu ordusu için saldırı/savunma gücü eklendi. Düşman ordularında güç yalnız görünen birim tiplerinden hesaplanıyor; gizli birimler güç toplamına sızmıyor. Kapsam: `internal/render/{army_panel.go,army_panel_test.go}`, `wiki/architecture/render-pipeline.md`; test: `go test ./internal/render`.

- 2026-07-21: Birimlerin `required_tech` alanı tüm senaryolarda diziye geçirildi. Liste AND semantiğiyle çalışıyor; 1300 senaryosunda ağır süvari, kuşatma ve deniz birimlerinin teknoloji zincirleri ara halkalarıyla birlikte tanımlandı. Oyuncu recruitment, tooltip, AI recruitment ve AI araştırma önceliği aynı ortak birim gereksinimi kontrolünü kullanıyor. Regression: `internal/army/unit_test.go`; doğrulama: `go test ./...`.

- 2026-07-21: 1300 AI deniz görevinde aktif objective barıştaki bir devlete kilitlenmişken
  yüklü transport filolarının beklemesi düzeltildi. Filo artık mevcut savaş düşmanları
  arasındaki ulaşılabilir kıyıları deterministik biçimde tarıyor; en yakın kazanılabilir
  kıyıya retarget oluyor, barıştaki devlete savaş ilan etmiyor ve güçlü savunuculu hedefi
  atlıyor. Regression: `Test1300LoadedFleetRetargetsNeutralObjectiveToReachableWarCoast`;
  doğrulama: `go test ./...`.

- 2026-07-21: Devlet bilgi paneli açıkken bölge seçimi değiştiğinde panel artık kapanmıyor; seçilen bölgenin sahibi devlete otomatik geçiyor. Aynı devletin başka bölgesine geçişte mevcut panel scroll'u korunuyor, farklı devlette sıfırlanıyor. Sahipsiz/deniz bölgesinde eski devlet bilgisi gösterilmemesi için panel kapanıyor. Kapsam: `internal/render/{renderer_input.go,faction_panel_test.go}`; test: `go test ./internal/render`.

- 2026-07-21: Seçilen devlet bilgi panelinin `Durum` bölümüne devletin askeri gücü ve aktif devletler arasındaki güç sırası eklendi. HUD ve devlet paneli aynı `factionMilitaryPowerStanding` hesabını kullanıyor; panel scroll içerik yüksekliği yeni iki satıra göre güncellendi. Kapsam: `internal/render/{panel.go,panel_test.go}`, `wiki/architecture/render-pipeline.md`; test: `go test ./internal/render`.

- 2026-07-21: Üst-sol oyuncu devlet HUD'unda bayrak ve adın altına mevcut askeri güç ile aktif devletler arasındaki güç sırası eklendi. Sıralama mevcut `diplomacy.MilitaryPower` hesabını kullanıyor; elenmiş devletler dışarıda bırakılıyor ve eşit güçte faction ID'siyle deterministik sonuç üretiliyor. Kapsam: `internal/render/{panel.go,panel_test.go}`, `wiki/architecture/render-pipeline.md`; test: `go test ./internal/render`.

- 2026-07-21: Kuşatma başladıktan sonra bölgeye gelen aynı realm/müttefik orduların
  kuşatma katkısı her tur yeniden hesaplanıyor. Böylece başlangıçta kuşatma birimi
  olmayan ordunun kuşatmasına sonradan getirilen kuşatma ekipmanı `BreachProgress` ve
  gedik seviyesine katılıyor. Regression: `TestResolveSiegesUsesSiegeUnitArrivingAfterSiegeStarted`;
  doğrulama: `go test ./...`.

- 2026-07-20: Ordu detay panelindeki birim kartları state sırası korunarak kategoriye
  göre yan yana gruplanıyor; görünüm sırası piyade, süvari ve kuşatma şeklinde.
  Kapsam: `internal/render/army_panel.go`; test: `go test ./internal/render`.

- 2026-07-20: Diplomasi hedef listesindeki devlet adı kolonu daraltıldı; ilişki/durum
  kolonu sola taşınarak uzun ilişki etiketlerinin sağ panele taşması engellendi.
  Kapsam: `internal/render/diplom.go`; test: `go test ./internal/render ./internal/ui`.

- 2026-07-22: Diplomasi hedef listesindeki her devlet satırının en soluna senaryo
  bayrağı 40×40 px kare rozet olarak eklendi; eksik bayrakta baş harfi fallback'i
  korunuyor ve devlet adı alanı rozet sonrasına göre hizalanıyor. Kapsam:
  `internal/render/diplom.go`, `wiki/architecture/render-pipeline.md`; test:
  `go test ./internal/render`.

- 2026-07-22: AI turundaki `Bilgi` popup'ı `HAMLELER` paneliyle örtüşmeyecek şekilde
  panelin altına 16 px boşlukla taşındı; normal bildirim konumu da biraz aşağı alındı.
  Kapsam: `internal/render/{panel.go,renderer.go}`, `wiki/architecture/render-pipeline.md`;
  test: `go test ./internal/render`.

- 2026-07-22: Diplomasi listesindeki vassal devlet satırına ikinci tıklama dış
  overlord devletin teklif panelini açıyor; oyuncunun kendi vassalında mevcut
  vassal yönetim paneli korunuyor. Regression: `TestHandleDiplomacyInputVassalDoubleClickOpensOverlord`;
  doğrulama: `go test ./internal/render`.

- 2026-07-20: Tur bitimindeki AI hamle paneli 620×180 px düzene büyütüldü; o an hamle yapan devletin faction ID'si render durumunda taşınıyor ve ülke adının solunda sarı iç çerçevesiz 128×128 px kare bayrak rozeti gösteriliyor. AI adım mesajı aynı anda genel `Bilgi` popup'ında tekrarlanmıyor; tur başında önceki oyuncu bildirimi temizleniyor. Asset yoksa baş harfi fallback'i korunuyor. Kapsam: `internal/game/game.go`, `internal/render/{panel.go,renderer.go}`, `wiki/architecture/render-pipeline.md`.

- 2026-07-20: Bölge panelindeki aktif olaylar ve komşu listesi çerçevesi görünürken
  metinlerin boş kalmasına neden olan `SubImage` koordinat hatası düzeltildi. İçerik
  origin'i viewport ekran koordinatlarına taşındı; scroll ve çizim sözleşmesi için
  regression testi eklendi. Doğrulama: `go test ./...`.

- 2026-07-20: Üst-sol durum HUD'unda oyuncu devletinin baş harfi yerine, aktif senaryoda
  faction ID'siyle eşleşen `sprites/flags/<faction-id>.png` bayrağı kare rozet içinde
  gösteriliyor; dosya bulunamazsa aynı kare zeminde baş harfi fallback'i korunuyor. Aynı
  bayrak rozeti devlet bilgi paneli başlığında da devlet isminin soluna ve bölge bilgi
  paneli çerçevesinin hemen üstüne bitişik sahiplik kimlik rozeti olarak eklendi; bölge
  ve devlet adlarının eski sol başlık konumu korundu. Bayrak asset'leri yol bazlı
  cache'leniyor ve senaryo değişiminde temizleniyor. Kapsam:
  `internal/render/{panel.go,renderer.go}`; test: `go test ./internal/render`.

- 2026-07-20: Faz 0 event veri bütünlüğü tamamlandı. `events.json` içindeki etkilenen
  fraksiyonlar, diplomatik ilişki hedefleri, sahip olunması gereken bölgeler ve teknoloji
  koşulları ile seçim etkilerindeki teknoloji/fraksiyon referansları senaryo testinde
  doğrulanıyor. Test-local JSON sözleşmesi kullanılarak `scenario -> events -> state ->
  scenario` import döngüsü engellendi.

- 2026-07-20: Faz 7 Anadolu AI profilleri tamamlandı. 13 beylik için yerel rekabet,
  çekirdek/geçit savunması ve savaş sonrası vassallık/ilhak objective'leri senaryo
  verisine işlendi; seçili Ege, Marmara, Pontus ve Toros komşulukları `-10` rekabet
  ilişkisiyle başlangıçta ayrıştırıldı. `scenario_1300_integrity_test.go` profil ve
  objective referanslarını doğruluyor; Normal fast iki seed ve medium 42 tur x 4 seed
  tempo profili (`27.57 sn`) başarılı.

- 2026-07-20: Faz 7 Venedik/Ceneviz deniz-ticaret profilleri tamamlandı. Venedik için
  Adriyatik-Girit-Kıbrıs savunması ve Konstantinopolis kapısı; Ceneviz için Cenova-
  Korsika-Kırım ticaret ağı ve Trabzon kapısı objective'leri eklendi. Mevcut merchant,
  liman ve escort güvenlik akışı veri profilleriyle birleşiyor; bölgesel profil testi iki
  deniz cumhuriyetini de doğruluyor.

- 2026-07-20: Faz 7 Memlük/İlhanlı ana cephesi tamamlandı. Açılış savaşı Mosul-Bağdat-
  Şam-Halep hattında karşılıklı expand objective'lerine bağlandı; Memlük'ün Kahire-
  Levant koridoru, İlhanlı'nın Bağdat-Musul-Malatya-Azerbaycan çekirdeği savunma
  objective'leri ve readiness bölgeleri tanımlandı. Senaryo plan testi iki devletin
  hedef devlet/bölge seçimini doğruluyor; Normal fast tempo testi başarılı.

- 2026-07-20: Faz 7 Balkan profilleri tamamlandı. Sırbistan, Bulgaristan, Epir,
  Arnavutluk, Atina ve Eflak için yakın sınır/geçit savunması ve düşük öncelikli yerel
  genişleme objective'leri işlendi. Altı devletin açılış planı savunma objective'ine
  yöneliyor; Normal fast tempo testi başarılı.

- 2026-07-20: Faz 7 Rusya/Altın Orda/Baltık profilleri tamamlandı. Rusya'nın Moskova-
  Nijni Novgorod-Dağıstan konsolidasyonu ve Ukrayna bozkırı yönü, Altın Orda'nın
  Kiev-Rusya/Litvanya baskısı, Teuton Tarikatı'nın Konigsberg-Litvanya cephesi ve
  liman savunması, Novgorod'un ticaret kapısı savunması ve Litvanya'nın Belarus-Kiev
  yönü senaryo AI verisine işlendi. Beş profilin ilk objective sözleşmesi senaryo
  bütünlük ve açılış plan testleriyle doğrulandı.

- 2026-07-20: Faz 7 İngiltere-Fransa tarihsel savaş kilidi tamamlandı. 1300'de iki
  devlet barışta başlıyor ve genel expansion target'ları temizleniyor; İngiltere'nin
  Kanal/ada, Fransa'nın kraliyet çekirdeği konsolidasyon profilleri aktif kalıyor.
  Mayıs 1337 event'i savaşı başlatıp `hundred_years_war_started` bayrağını yazıyor;
  yıl ve event bayrağı hard gate'ini geçen iki kıta objective'i sonraki planlamada
  devreye giriyor. Diplomasi, profil ve game geçiş testi başarılı.

- 2026-07-20: Faz 7 Safevî survival/yükseliş profili tamamlandı. 1300'de Güney İran
  çekirdeğini koruyan konsolidasyon objective'i aktif; Ocak 1501 `safavid_rise_1501`
  olayı `safavid_rise` bayrağıyla kaynak/teknoloji takviyesi veriyor. Bayrak sonrası
  Safevîlerin Azerbaycan-Batı/Kuzey İran ve Mezopotamya genişleme objective'i, Osmanlı'nın
  doğu rekabetiyle aynı tarihsel hard gate'e bağlandı. Safevî geçiş testi başarılı.

- 2026-07-20: Erken ekonomi teknolojileri kalibre edildi. Ticaret Yolları, Bankacılık,
  Loncalar, Tahrir Defterleri, Kervansaray Ağı ve Darphane Standardı'nın bölge/pazar
  getirileri kademeli olarak düşürüldü; önkoşul, araştırma maliyeti ve askeri teknoloji
  değerleri korundu. Normal medium 42x4 ölçümünde büyük devletlerin 42 aylık altın
  birikimi `27–31 bin` bandından yaklaşık `20–25 bin` bandına indi. Senaryo teknoloji
  sözleşmesi testi eklendi.

- 2026-07-20: Faz 8 çok-seed kabul bantları test sözleşmesine alındı. Medium 42x4 tempo
  raporu büyük/orta devletlerin altın kazanımını fraksiyon gruplarına göre doğruluyor;
  fast 12x2 hızlı regresyon ve iki turluk tam state replay deterministiklik kapsamı
  korunuyor. `assert1300CalibrationBands` yalnız medium veya daha geniş profillerde
  çalışarak normal test akışını yavaşlatmıyor.

- 2026-07-20: 1300 AI diplomasi taramasında ilişki skoru 25 altındaki barış çiftleri
  stratejik ittifak değerlendirmesine sokulmuyor. Bu güvenli kısa devre ile 42 tur x 8
  seed tempo `103.6 sn`den `101.6 sn`ye indi; nihai 60 saniye hedefi için optimizasyon
  çalışması devam ediyor.

- 2026-07-20: Faz 1 refactor'ın ilk diliminde faction/region/army sıralama yardımcıları
  `internal/ai/ordering.go` dosyasına taşındı. Sıralı map erişimi, nil davranışı ve
  deterministik ID sırası için test eklendi; AI ve game davranış sözleşmesi korundu.

- 2026-07-20: Diplomasi karar döngüsü (`aiHandleDiplomacyWithSteps`) `internal/ai/diplomacy.go`
  dosyasına ayrıldı. Barış, ittifak, ticaret ve vassal diplomasi geçitleri korunurken
  `ai.go` ortak tur orkestrasyonu ve wrapper'larla sınırlandı; AI/game ve replay testleri
  başarılı.

- 2026-07-20: Fırsat savaşı aday taraması ve hedef puanlaması `internal/ai/war_strategy.go`
  dosyasına ayrıldı. Güç, cephe, sınır, objective ve expansion target kararları aynı
  çağrı sözleşmesiyle çalışıyor; hareket/çarpışma katmanına dokunulmadı.

- 2026-07-20: Üretim/recruitment orkestrasyonu `internal/ai/recruitment_strategy.go`
  dosyasına taşındı. Kışla, manpower, bütçe ve üretim kuyruğu davranışı korunurken
  birim seçimi ve recruitment bölgesi değerlendirmesi ayrı yardımcı katmanlarda kaldı.

- 2026-07-20: Araştırma başlatma ve ekonomi bina wrapper'ları
  `internal/ai/economy_research.go` dosyasına taşındı. Aktif araştırma/rezerv kontrolü,
  bütçe tüketimi ve bina yatırım stratejisi çağrıları korunarak AI orkestrasyonu küçültüldü.

- 2026-07-20: Deniz stratejisi giriş wrapper'ları `internal/ai/naval_strategy.go`
  dosyasına ayrıldı. Legacy kıyı üretimi ile 1300 görev/merchant stratejik context geçidi
  korunurken deniz görev hesaplayıcıları mevcut modüllerde bırakıldı.

- 2026-07-20: Kuşatma state/başlatma, sanal garnizon ve tahkimat savunma katsayısı
  `internal/ai/siege_strategy.go` dosyasına ayrıldı. Hareket hedefi seçimi ve combat
  resolve davranışı değiştirilmedi.

- 2026-07-20: Uzun menzilli hareket hedefleme `internal/ai/movement_strategy.go`
  dosyasına ayrıldı. 1300 ağırlıklı rota ve legacy BFS seçimi ile stratejik rol geçidi
  korunurken komşu skorlaması ve state uygulaması aynı kaldı; pathfinding testleri geçti.

- 2026-07-20: Komşu hareket hedefi, lojistik score context'i, embark puanı ve `scoreMove`
  kararları da `internal/ai/movement_strategy.go` dosyasına taşındı. `ai.go` hareketin
  gerçek state/çarpışma uygulamasına daraltıldı; AI ve game testleri başarılı.

- 2026-07-20: Faz 1 doğrulaması tamamlandı. `Test1300ScenarioAITwoTurnReplayIsDeterministic`
  aynı seed ile iki tam state akışını eşit doğruladı. Refactor sonrası 42x8 Normal tempo
  `101.57 sn` test süresinde (`105.07 sn` duvar saati) tamamlandı; önceki `101.6 sn`
  referansına göre regresyon yok. Nihai 60 saniye optimizasyon maddesi açık.

- 2026-07-20: CPU/allocasyon profiliyle diplomasi hot path'i optimize edildi. Tehdit
  kontrolü tek snapshot kullanıyor, ittifak perspektifleri ticaret erişimini paylaşıyor;
  `FactionOrder`/`RegionOrder` doğrudan, dinamik ordular `ArmyOrder` runtime cache'iyle
  okunuyor. 42x8 Normal tempo `60.416 sn`den `55.658 sn` test süresine indi; duvar saati
  `58.568 sn` ve nihai 60 saniye hedefi karşılandı.

- 2026-07-20: Faz 3 deniz rol entegrasyonu tamamlandı. `AIArmyRoleTransport` ve
  `AIArmyRoleEscort`, kara ordusu rollerinin bulunduğu `ArmyAssignments` map'ine taşındı;
  aktif naval mission anchor'larıyla transport/escort filoları ayrıştırılıyor, merchant
  filoları rol dışı kalıyor. `TestNavalAssignmentsUseTransportAndEscortRoles` eklendi.

- 2026-07-20: `1300_ottoman_rise` açılış diplomasisi tarihsel cephelere göre düzeltildi.
  Osmanlı-Doğu Roma, Memlük-İlhanlı, Aragon-Kastilya, Aragon-Napoli, İngiltere-İskoçya
  ve Fransa-HRE savaşta; İngiltere-Fransa 1337 event'ine kadar barışta, Aragon-Granada
  müttefik, Venedik-Ceneviz barışta bırakıldı. Yeni `flanders_county` Katolik fraksiyonu Flandre bölgesinin sahibi
  ve HRE vassalıdır; Flandre-Fransa savaşı HRE-Fransa kök savaşıyla aynı koalisyona
  bağlanır. Cephe orduları ve başlangıç teknoloji seviyeleri yeniden atandı; ilişki
  loader'ında deterministik faction ID sırası sabitlendi.

- 2026-07-20: `1300_ottoman_rise` başlangıç donanması tarihsel profil ve oyun
  ölçeğine göre seed edildi. Venedik/Ceneviz savaş ve ticaret filoları; Doğu Roma,
  Aragon, İngiltere ve Fransa savaş-nakliye filoları; Portekiz nakliye/ticaret
  filoları; Memlük nakliye filosu ile başlıyor. Londra, Normandiya, Portekiz,
  Sicilya ve Mısır'a geçerli dock için başlangıç `port` binası eklendi. Osmanlı,
  Safevî ve Rusya açılışta donanmasız tutuldu. `internal/scenario` bütünlük testi
  başlangıç filolarının sahip, deniz, liman ve birim kategorilerini doğruluyor.

- 2026-07-20: `1300_ottoman_rise` AI merchant ticaret akışı tamamlandı. Merchant
  filoları `TradeRouteKey` ile save uyumlu yönlü rotaya atanıyor; aktif kıyısal trade
  center uçlarından en az biri ile gerçek deniz konumu doğrulanmadan hacim bonusu verilmiyor.
  Her gemi rota başına `+1`, en fazla `+2` throughput sağlıyor; askıdaki rota veya
  mal/altın yetersizliği normal ticaret transfer kapısından geçiyor. Venedik/Ceneviz
  eksik slotları üretim kuyruğuna alıyor, liman seviyesi ve ilk merchant maliyeti için
  kaynak rezervi korunuyor; tehditli merkezler aynı `%110` escort eşiğiyle önce
  savunuluyor. Merchant, transport ve warship filoları konsolidasyon/production
  tamamlanma katmanında ayrıştırıldı. Normal fast 12x2 `6.71 sn`, Osmanlı `2 → 3`
  bölge/güç `238`, deniz birimi ilk beş tur `0`, yedinci turda `2`. Test kapsamı:
  `internal/ai/merchant_trade_test.go`, `internal/state/merchant_trade_test.go`,
  `internal/save/save_test.go`, `internal/game/production_naval_test.go`.

- 2026-07-19: `1300_ottoman_rise` AI deniz görevlerine efektif tehdit haritası,
  güvenli rota ve `%110` filo gücü kapısı eklendi. Düşman güç hesabı gerçek deniz
  muharebesindeki `TotalStrength`, teknoloji ve komutan etkilerini kullanıyor. Rotalar
  maksimum tehdit, toplam tehdit, mesafe ve ID sırasıyla deterministik seçildiği için
  uzun güvenli hat kısa tehlikeli hatta tercih ediliyor. Zorunlu tehditli adımda gerçek
  filo stack'i düşmanın `%110`una ulaşmadan ilerlemiyor. Görev rotası veya çıkış limanı
  tehditliyse mevcut/pending filo gücü aynı eşiğe erişene kadar port/escort yatırımı
  öncelikleniyor. Context tehdidi tur başına, hareket tehdidi rota çağrısı başına bir
  kez hesaplanıyor. Normal fast 12x2 `6.65 sn`; Osmanlı `2 → 3`, güç `238`, deniz
  birimi altıncı turda `1`. Kapsam: `internal/ai/naval_threat.go`,
  `internal/ai/naval_mission.go`, strategic context bağlantıları ve
  `internal/ai/naval_threat_test.go`; doğrulama `go test -count=1 ./...`.

- 2026-07-19: `1300_ottoman_rise` deniz taşıma AI'sı somut görev ve kapasite açığı
  modeline taşındı. Aktif genişleme/savunma objective'inde güvenli kara yolu olmayan
  kıyı hedefi için taşınacak ordu, çıkış limanı, çıkış denizi ve hedef kıyı
  deterministik seçiliyor. Erişilebilir mevcut boş kapasite ile aynı hattaki pending
  transport kapasitesi birlikte sayılıyor; yalnız açık kadar gemi üretiliyor. Liman ve
  escort yatırımı aynı göreve bağlandı. Seçilen ordu limana, boş transport çıkış
  denizine, yüklenmiş filo objective kıyısına ve görev savaş gemileri taşıma hattına
  yöneliyor; görevsiz filolar uzak denizde dolaşmıyor. Model 1300'e özel, diğer
  senaryoların legacy davranışı korunuyor. Normal fast 12x2 `6.39 sn`, Osmanlı `2 → 3`,
  güç `238`; deniz birimi ilk beş tur `0`, altıncı turda somut görevle `1`. Kapsam:
  `internal/ai/naval_mission.go`, `internal/ai/{ai.go,fronts.go,strategic_plan.go}` ve
  `internal/ai/naval_mission_test.go`; doğrulama `go test -count=1 ./...`.

- 2026-07-19: `1300_ottoman_rise` ittifak AI'sı tehdit, tampon devlet, cephe desteği,
  ticaret ve objective çakışması tabanlı ortak değerlendirmeye taşındı. Teklif sahibi
  ile hedef AI aynı bileşenleri ayrı perspektiften okuyor. Aktif objective çakışması
  ittifakı kesin engelliyor; mevcut ittifakta hedef savaşı kilitleyen trade stance'i de
  kapanıyor. Statik expansion target `-18` aşılabilir gerilim cezası; ortak düşman/büyük
  tehdit, adayın tehdit sınırındaki tampon konumu, bu cephedeki gerçek ordu gücü, ticaret
  hattı/kapasitesi ve partner katkısı değer üretiyor. Tehdit ve fayda kaybolduğunda düşük
  değerli ittifak retention eşiğiyle çözülebiliyor. Model 1300'e özel, legacy senaryolar
  korunuyor. Normal fast 12x2 `6.25 sn`, Osmanlı `2 → 3`, güç `228`. Kapsam:
  `internal/diplomacy/alliance_strategy.go`, `internal/ai/ai.go` ve iki paketin stratejik
  ittifak testleri.

- 2026-07-19: `1300_ottoman_rise` için save/load uyumlu aktif savaş `WarLedger` state'i
  ve amaç odaklı barış kararı eklendi. Savaş başlangıç turu/bölge snapshot'ı, iki tarafın
  birlik kayıpları ve fetihleri, son muharebe ve teklif turu compact + legacy/debug
  save'de korunuyor; eski save savaşları yükleme turunda sıfır sayaçla göç ediyor.
  Savaş, çıkarma, huruç, genel hücum, kuşatma yıpratması ve fetih executor'ları sayaçları
  besliyor. Barış skoru objective ilerlemesi, toprak/kayıp dengesi, süre/durgunluk, güç,
  ekonomi, çoklu savaş ve başkent tehdidini birleştiriyor; ilk üç tur normal teklif
  kapalı, askerî çöküş/başkent tehdidi acil istisna ve teklif cooldown'u üç tur. AI-AI
  tarafları aynı modeli ayrı perspektiften geçiriyor; oyuncuya teklif kuyruklanıyor.
  Diğer senaryolar legacy davranışı koruyor. Ölçümler: Normal fast 12x2 `9.64 sn`,
  Osmanlı `2 → 3`, güç `232`; Zor fast 12x2 `7.56 sn`, Osmanlı `2 → 3`, güç `222`;
  Normal medium 42x4 `69.17 sn`, Osmanlı ortalama `2 → 4.2`, güç `424`. Kapsam:
  `internal/state/war_ledger.go`, `internal/diplomacy/peace_assessment.go`, savaş/fetih
  executor'ları ve save overlay zinciri.

- 2026-07-19: `1300_ottoman_rise` araştırma AI'sına plan ve darboğaz tabanlı teknoloji
  puanlaması eklendi. Genişleme askerî/hareket/kuşatma, savunma kara savunması/istikrar,
  konsolidasyon ekonomi/tahıl/istikrar yönelimini kullanıyor. Gerçek teknoloji efektleri,
  12 turluk üretim getirisi, stok ve bakım baskısı, din farkı, aktif savaş, kıyı erişimi,
  doğrudan birim açılımı, sonraki teknoloji, maliyet ve süre birlikte değerlendiriliyor.
  Aktif araştırma korunuyor, player-only istihbarat AI değeri üretmiyor ve legacy
  senaryolar değişmiyor. Ölçümler: Normal fast 12x2 `9.76 sn`, Osmanlı `2 → 3`, güç
  `270`; Normal medium 42x4 `68.84 sn`, Osmanlı ortalama `2 → 5.5`, güç `466`; Zor
  fast 12x2 `7.80 sn`, Osmanlı `2 → 3`, güç `254`. Büyük devletlerin daha erken ekonomi
  teknolojileriyle `27–31 bin` altın biriktirmesi Faz 7 kalibrasyon notuna alındı;
  sonraki ekonomi veri kalibrasyonuyla bu bant yaklaşık `20–25 bin` seviyesine çekildi.
  Kapsam: `internal/ai/{research_strategy.go,ai.go}` ve
  `internal/ai/research_strategy_test.go`.
- 2026-07-19: `1300_ottoman_rise` kara üretimine stratejik recruitment bölgesi seçimi
  eklendi. Aday hatlar kalan throughput, kışla seviyesi, kuyruk, güvenli ağırlıklı
  Dijkstra cephe mesafesi ve mevcut + pending + yeni birlik sonrası lojistik boşluğuyla
  puanlanıyor. Kuşatma, yabancı ordu, gerçek isyan riski, kritik tehdit cephesi veya
  projekte ikmal aşımı adayı eliyor. Kuşatma desteği ilgili eksik hücum cephesine yakın
  üretiliyor. Model 1300'e özel, deterministik ve runtime-only; legacy senaryo davranışı
  korunuyor. Ölçümler: Normal fast 12x2 `9.56 sn`, Osmanlı `2 → 3`, güç `278`; Normal
  medium 42x4 `67.44 sn`, Osmanlı ortalama `2 → 6.5`, güç `568`; Zor fast 12x2
  `7.74 sn`, Osmanlı `2 → 3`, güç `269`. Kapsam:
  `internal/ai/{recruitment_region.go,unit_composition.go,ai.go}` ve
  `internal/ai/recruitment_region_test.go`.
- 2026-07-19: `1300_ottoman_rise` kara üretimine plan bazlı ordu kompozisyonu eklendi.
  Genişleme `%55/%25/%20`, savunma `%75/%15/%10`, konsolidasyon `%65/%25/%10`
  piyade/süvari/kuşatma hedefi kullanıyor. Haritadaki, gemideki ve kuyruktaki birlikler
  birlikte sayılıyor; bileşim açığı gerçek saldırı/savunma/moral, hedef arazisi, düşman
  profili, maliyet, hammadde ve tahıl bakım baskısı ile üretim süresine karşı puanlanıyor.
  Tahkimli hücum hedefinde kuşatma desteği eksik hücum orduları önceliklendiriliyor.
  Model runtime-only ve 1300'e özel; diğer senaryolar legacy sırayı koruyor. Ölçümler:
  Normal fast 12x2 `9.08 sn`, Osmanlı `2 → 3`, güç `292`; Normal medium 42x4
  `62.53 sn`, Osmanlı ortalama `2 → 5`, güç `733`; Zor fast 12x2 `7.39 sn`, Osmanlı
  `2 → 3`, güç `290`. Kapsam: `internal/ai/{unit_composition.go,ai.go}` ve
  `internal/ai/unit_composition_test.go`.
- 2026-07-19: `1300_ottoman_rise` ekonomi AI'sına bina yatırım puanlaması eklendi.
  Pazar, çiftlik, sur ve ibadet yeri adayları gerçek marjinal 12 turluk üretim ROI'si,
  güncel hammadde fiyatı, tahıl/ordu bakım açığı, stok fırsat maliyeti, cephe ve kritik
  merkez tehdidi, objective/rally, memnuniyet, süre, seviye ve bölgesel kuyrukla
  karşılaştırılıyor. `80` yatırım eşiğini aşmayan aday ekonomi bütçesini tüketmiyor;
  pay sonraki donanma/ordu kategorisine geçiyor. Tur başına tek bina ve diğer
  senaryoların legacy sırası korunuyor; kullanılmayan `buildingPlan.prio` kaldırıldı.
  Ölçümler: Normal fast 12x2 `8.91 sn`, Osmanlı `2 → 3`, güç `268`; Normal medium
  42x4 `62.12 sn`, Osmanlı ortalama `2 → 5.8`, güç `657`; Zor fast 12x2 `7.22 sn`,
  Osmanlı `2 → 3`, güç `267`. Kapsam: `internal/ai/{building_investment.go,ai.go}` ve
  `internal/ai/building_investment_test.go`.
- 2026-07-19: `1300_ottoman_rise` AI için plan türüne bağlı runtime harcama bütçesi
  eklendi. Acil rezerv devlet büyüklüğü, efektif aylık altın, aktif savaş ve kritik
  merkez tehdidinden türetiliyor; `80–420` aralığında korunuyor. Rezerv üstü altın
  genişleme, savaş/savunma veya konsolidasyon profiline göre ordu, ekonomi, araştırma
  ve donanma soft paylarına ayrılıyor. Kullanılmayan pay aynı tur sonraki kategoriye
  aktarılıyor; kara devletinde donanma payı oransal yeniden dağıtılıyor. Model save'e
  yazılmıyor ve diğer senaryoların sabit `80` rezerv davranışı korunuyor. Ölçümler:
  Normal fast 12x2 `8.87 sn`, Osmanlı `2 → 3`, güç `288`; Normal medium 42x4
  `62.22 sn`, Osmanlı ortalama `2 → 5`, güç `670`; Zor fast 12x2 `7.35 sn`, Osmanlı
  `2 → 3`, güç `289`. Kapsam: `internal/ai/{budget.go,ai.go,strategic_plan.go}` ve
  `internal/ai/budget_test.go`.
- 2026-07-19: `1300_ottoman_rise` kara AI'sı için ortak ağırlıklı Dijkstra rota motoru
  eklendi. Arazi `MoveCost`, diplomatik erişim, yerel savaş düşmanı gücü ve öngörülen
  ikmal aşımı rota maliyetine katılıyor. Genel rota aynı-realm/müttefik transitini
  kullanırken sahipsiz ve savaş düşmanı bölgesini terminal sayıyor; barıştaki üçüncü
  tarafı geçiş hattı yapmıyor. `retreat/security` yalnız güvenli kendi toprağında kalıyor.
  Rotalar AI turunda cache'leniyor ve deterministik eşitlik çözümü kullanıyor. Gerçek
  hareket puanı ve save şeması değişmedi; diğer senaryolar BFS fallback kullanıyor.
  Ölçümler: Normal fast 12x2 `9.53 sn`, Osmanlı `2 → 3`, güç `272`; Normal medium
  42x4 `66.26 sn`, Osmanlı `2 → 5.8`, güç `665`; Zor fast 12x2 `7.41 sn`, Osmanlı
  `2 → 3`, güç `284`. Kapsam: `internal/ai/pathfinding.go`, `internal/ai/{ai,fronts,
  retreat,security,strategic_plan}.go`, `internal/ai/pathfinding_test.go`.
- 2026-07-19: `1300_ottoman_rise` için memnuniyet/din tabanlı fetih sonrası `security`
  rolü eklendi. Sursuz ve sabit garnizonsuz aynı din bölgesi memnuniyet `<35`, farklı
  din bölgesi `<45` olduğunda en küçük uygun saha ordusunu çağırıyor. `<30` gerçek isyan
  riskinde ordu aynı tur erişebilir olmalı; tek saha ordulu devlet yalnız bu acil eşikte
  kuvvet ayırıyor. Security dost ve işgal edilmemiş kara rotası kullanıyor, anchor'da
  eşik düzelene kadar kalıyor. Kritik cephe defense, relief ve aktif siege korunuyor;
  kullanılan stratejik rezerv savaş hazırlığı sayacından düşüyor, ağır yıpranmada
  retreat öncelik kazanıyor. Yeni save alanı yok; rol runtime-only. Ölçümler: Normal
  fast 12x2 `9.38 sn`, Osmanlı `2 → 3`; Normal medium 42x4 `64.17 sn`, Osmanlı
  `2 → 5.8`, güç `664`; Zor fast 12x2 `7.20 sn`, Osmanlı `2 → 3`. Kapsam:
  `internal/ai/{security.go,fronts.go,ai.go}`, `internal/ai/security_test.go`.
- 2026-07-19: `1300_ottoman_rise` kara AI'sına güvenli geri çekilme/takviye rolü
  eklendi. Açık arazi ordusu ağırlıklı tam güç oranı `%45`in altına indiğinde veya
  aynı/komşu bölgelerdeki savaş düşmanı gücü en az `%135` olduğunda `retreat` rolü
  alıyor. En yakın kuşatılmamış, yabancı ordusuz, komşu düşman tehdidi olmayan ve varış
  sonrası ikmali aşılmayan dost bölgeye yalnız dost kara hattından ilerliyor. Aktif
  kuşatma yalnız ikmal aşımı ile komşu relief gücünün kuşatan gücün `%150`sini aşması
  birlikte gerçekleşirse devrediliyor/kaldırılıyor; güvenli kaçış yoksa korunuyor.
  Kuşatma withdrawal adımı `TakeTurn` ve `TurnStepper` akışlarında hareketten önce
  çalışıyor. Ölçümler: Normal fast 12x2 `9.37 sn`, Osmanlı `2 → 3`; Normal medium 42x4
  `63.65 sn`, Osmanlı `2 → 5.8`, güç `664`; Zor fast 12x2 `7.03 sn`, Osmanlı `2 → 3`.
  Kapsam: `internal/ai/{retreat.go,fronts.go,ai.go,turn_stepper.go}` ve
  `internal/ai/retreat_test.go`.
- 2026-07-19: `1300_ottoman_rise` AI için kalıcı koordineli rally hazırlığı eklendi.
  En az iki hücum/kuşatma ordusu varsa güvenli dost hedef sınırı ikmal, yerel güç,
  tahkimat, başkent ve stratejik değerle seçiliyor. Ordular burada toplam hücum gücünün
  `%60`ı ile zorluğun hedef frontier oranını birlikte karşılayana kadar bekliyor; en
  fazla üç tur sonra serbest kalıyor. Aktif rally proaktif savaşı erteliyor, tek hücum
  ordusu bulunan küçük devlet bekletilmiyor, kuşatılan/geçersiz rally iptal ediliyor.
  `AIPlanState.RallyRegionID/RallyDeadlineTurn` compact/legacy/debug save hattında
  korunurken güç ve roller runtime-only kalıyor. Ölçümler: Normal fast 12x2 `9.21 sn`,
  Osmanlı `2 → 3`; Normal medium 42x4 `63.18 sn`, Osmanlı `2 → 6`, güç `666`; Zor
  fast 12x2 `7.25 sn`, Osmanlı `2 → 3`. Memlük medium büyümesi `+8.5`ten `+5`e indi.
  Kapsam: `internal/state/ai_plan.go`, `internal/ai/{rally.go,fronts.go}`, save roundtrip
  ve `internal/ai/rally_test.go`.
- 2026-07-19: `1300_ottoman_rise` AI için düşman eksenli cephe snapshot'ı, dinamik
  rezerv ve kara ordusu görevleri eklendi. `AIFront` dost/düşman sınır bölgelerini,
  gerçek cephe güçlerini, savaş/objective ilişkisini ve başkent/kritik merkez tehdidini
  runtime'da hesaplıyor. Mobil ordular her AI turunda `assault`, `siege`, `defense`,
  `reserve` veya `relief` rolü alıyor; aktif savaş barıştaki objective'ten önce
  sonuçlandırılıyor. Rezerv normalde mobil gücün `%15`i, kritik tehditte `%30`u; tek
  saha ordusu dondurulmuyor ve en güçlü stack aktif tutuluyor. Yeni savaş saldırı gücü
  ve rezerv hazır değilse veya kritik merkez tehdit altındaysa erteleniyor. Roller
  standart hareket/diplomasi kurallarını aşmıyor ve save'e yazılmıyor. Ölçümler: Normal
  fast 12x2 `9.06 sn`, Osmanlı `2 → 3`; Normal medium 42x4 `60.31 sn`, Osmanlı ortalama
  `2 → 6`, güç `672`; Zor fast 12x2 `6.96 sn`, Osmanlı `2 → 3`. Kapsam:
  `internal/ai/{fronts.go,strategic_plan.go,ai.go,turn_stepper.go}`; testler:
  `internal/ai/fronts_test.go` ve hedef paket regresyonları.
- 2026-07-19: `1300_ottoman_rise` için veri güdümlü zorluk politikası eklendi.
  Kolay/Normal/Zor artık büyük kaynak hilesi yerine plan ufku (`4/6/9`), objective
  hedef kapsamı (`3/4/5`), yol arama derinliği (`5/8/12`), hareket hedefi ağırlığı,
  savaş risk eşiği/kadansı ve eşzamanlı savaş kapasitesi (`1/1/2`) ile ayrışıyor.
  Kolay proaktif savaş açmıyor; Zor denk güçte daha uzun plan kurup iki cephe
  taşıyabiliyor. 1300'de oyuncu ve AI aynı hareket kurallarını kullanıyor; eski Zor
  `+300 altın/+100 tahıl` bonusu küçük `+80/+30` başlangıç tamponına indirildi.
  Config taşımayan diğer senaryoların legacy davranışı korunuyor. Zor fast tempo 12x2
  ölçümü `6.84 sn` sürdü ve Osmanlı iki seed'de de `2 → 5` kara bölgesine ulaştı.
  Kapsam: `internal/{scenario,ai,state,game,save}`, `data/ai_strategies.json`; hedef
  testler: `go test -count=1 ./internal/scenario ./internal/ai ./internal/state ./internal/game ./internal/save`.
- 2026-07-19: `1300_ottoman_rise` için veri güdümlü Osmanlı/Doğu Roma AI objective
  dikey dilimi eklendi. `data/ai_strategies.json`; Bitinya, Anadolu beylikleri, Ankara
  koridoru, Trakya/Konstantinopolis, 1501 sonrası Safevi rekabeti ve Doğu Roma Boğaz
  savunması/Bitinya geri alma yönlerini tanımlıyor. Tarihsel hedefler mevcut güç ve
  frontier kontrollerini atlamayan soft savaş/hareket bonusları verir; yalnız geç
  hedefler yıl/event hard gate'i kullanır. Anadolu beyliklerinde son toprak sonrası
  sonuç hibrittir: zayıf ve dış müttefiksiz hedef vassal kalabilir, dirençli veya
  `annex_region_ids` ile stratejik sayılan hedef doğrudan ilhak edilir. Politika açık
  arazi, savaşsız işgal, çıkarma, genel hücum ve kuşatma tesliminde ortaktır. Yeni
  profil/reference, objective gate/puan, save alanı ve vassallık integration testleri
  eklendi; fast tempo 12x2 ölçümü `9.08 sn`, medium tempo 42x4 ölçümü `59.89 sn`
  geçti. Medium sonuçta Osmanlı ortalama `2.0 → 7.8` kara bölgesine çıktı.
- 2026-07-19: `1300_ottoman_rise` AI için hibrit kalıcı stratejik plan temeli eklendi. `GameState.AIPlans`, objective türü, hedef devlet/bölge öncelikleri, başlangıç ve yeniden değerlendirme turu, commitment ve karar nedenini save/load arasında koruyor; compact `state_zstd`, legacy payload ve dev debug sidecar hatları destekleniyor. Runtime `StrategicContext` manpower, deployed unit, sınır/savaş, askeri güç, frontier güç ve bölge değeri cache'lerini her AI turunda yeniden kuruyor ve save'e yazmıyor. İlk plan seçimi senaryodaki `ai_expansion_targets`, aktif savaş veya konsolidasyon fallback'inden üretiliyor; hedef geçersizleşince ya da altı tur dolunca yenileniyor, devlet elenince siliniyor. Kapsam: `internal/state/ai_plan.go`, `internal/ai/strategic_plan.go`, `internal/save/compact.go`, `internal/game/resolution.go`; testler: `internal/ai/strategic_plan_test.go`, `TestAIPlanStateRoundTripKeepsDurableIntent`, eliminasyon plan temizliği.
- 2026-07-18: `1300_ottoman_rise` AI Faz 0 ölçüm/determinizm altyapısı tamamlandı. Tempo raporu `fast` (12x2), `medium` (42x4), `calibration` (120x8) profillerine ayrıldı; tur süre/state sayaçları ve `Benchmark1300AIRound` eklendi. AI hedef puanlamasında hareket-scope cache ve ID tabanlı karar sıraları kullanılıyor. Başarısız genel hücum veya uygulanamayan hedefin hareket puanı tüketmeden sonsuz seçilmesi hem `TakeTurn` hem `TurnStepper` için engellendi. Birleşik savunmacı ve event bölge sıraları deterministik hale getirildi; aynı seed'li iki turluk replay testi geçiyor. Ölçümler: fast `9.0 sn`, medium `56.5 sn`, 42x8 kabul kapsamı `107.6 sn` (önceki koşu 10 dakika/timeout). Kapsam: `internal/{ai,state,events,game}`, `wiki/{systems/ai.md,systems/events.md,architecture/state-management.md}`.
- 2026-07-18: `1300_ottoman_rise` AI iyileştirme çalışmasının Faz 0 veri bütünlüğü tamamlandı. Ahiler'in hayalet `ankara` garnizonu `sivrihisar`a taşındı; Osmanlı yükseliş zaferindeki geçersiz `ankara` hedefi `cankiri`, dinî zaferdeki `mecca` hedefi `hejaz` yapıldı. Tekrar eden settlement ID'leri global olarak tekilleştirildi, etkilenen başkent referansları güncellendi ve İlhanlı başkenti Tebriz (`azerbaijan_main`) olarak sabitlendi. Army, victory, faction, capital, settlement ID ve trade çapraz referanslarını koruyan `internal/scenario/scenario_1300_integrity_test.go` eklendi; testler: `go test -count=1 ./internal/scenario ./internal/world`.
- 2026-07-18: Kuşatma ordusunun üstündeki kılıç rozeti kare dolgu/çerçeveden çıkarıldı; aynı ölçüde beyaz daire arka planına geçirildi. Kapsam: `internal/render/renderer.go`, `wiki/architecture/render-pipeline.md`, test: `go test ./internal/render`.
- 2026-07-18: Yerleşim marker ikonlarının beyaz daire arka planına göre dikey hizası düzeltildi; ikon ve daire artık aynı merkez koordinatını kullanıyor. Kapsam: `internal/render/renderer.go`, `wiki/architecture/render-pipeline.md`, test: `go test ./internal/render`.
- 2026-07-18: Kuşatma emri paneli seçili ordu detay paneli görünür kalacak şekilde ordu panelinin üstündeki boş alana taşındı; buton önceliği korunurken panel dışındaki input/cursor ordu paneline geçebiliyor. Kapsam: `internal/render/{renderer_dialogs.go,renderer_input.go,cursor.go}`.
- 2026-07-18: Eğitim kuyruğu kartlarına birim adı eklendi; tüm kart footer'ları tam genişlikte opak beyaz çizilerek kenarlardaki sprite sızıntısı kaldırıldı. Kapsam: `internal/render/recruit_panel.go`.
- 2026-07-18: Ordu komutan kartındaki sağ bilgi bloğu portreyle üst hizaya getirildi; rol, seviye, savaş ve bonus satırları 6px yukarı alındı. Kapsam: `internal/render/commander_component.go`.
- 2026-07-18: Askeri sprite seçimi düzeltildi; yalnızca Sünni/Şii fraksiyonlar `eastern_army`, diğer tanımlı dinler `western_army` kullanıyor. Kapsam: `internal/render/recruit_panel.go`, `renderer_army_icon_test.go`.
- 2026-07-17: Kuşatma altındaki bölge sahibi veya müttefik ordular için huruç hareketi eklendi. Komşu dost/sahipsiz hedefe çıkmadan önce kuşatanla savaş planı açılıyor; zaferde kuşatma kalkıp ordu çıkıyor, yenilgide kalan birlikler kayıplarıyla kuşatılan bölgede kalıyor ve hareket/iyileşme duruyor. AI aynı state kuralını uyguluyor; ortak `GameState.IsArmyDefendingSiegedRegion()` predicate'i hareket, tur sonu toparlanma ve panel görünümünü hizalıyor. Kapsam: `internal/state/state.go`, `internal/game/{game.go,resolution.go,siege_test.go}`, `internal/ai/{ai.go,siege_test.go}`, `internal/render/{renderer_dialogs.go,renderer_input.go,army_panel.go}`, `wiki/{systems/combat.md,systems/ai.md,architecture/state-management.md,architecture/render-pipeline.md}`, testler: `go test ./internal/state ./internal/game ./internal/ai ./internal/render`.
- 2026-07-17: Bölge bilgi panelinin altındaki aktif olay/komşu viewport'u artık komşuları geliştirme moduna bağlı olmadan gösteriyor; uzun içerik için yükseklik ve çizim hesabı eşitlendi. Diplomasi aksiyon düğmesinin etiketi dikey olarak ortalandı ve düğmenin çizim/hit-test geometrisi gerçek bina/stat yerleşimiyle hizalandı. Kapsam: `internal/render/panel.go`, `internal/render/region_panel_activity_test.go`, testler: `go test ./...`.
- 2026-07-17: Birim kartı sprite'ları fraksiyon dinine göre ayrıldı; Sünni/Şii ordular `muslim_army.png`, Katolik/Ortodoks ordular `christian_army.png` kullanıyor ve eski senaryolar için `army.png` fallback'i korunuyor. Recruit/tooltip oyuncu dinini, ordu detay paneli ordu sahibinin dinini baz alıyor. Kapsam: `internal/render/{recruit_panel.go,army_panel.go,hover_tooltip.go,renderer.go}`, `internal/render/renderer_army_icon_test.go`.
- 2026-07-17: Ordu ve recruit birim kartları yaklaşık `%20` büyütüldü; 10×2 slot düzeni korundu, sprite’lar uniform ölçeklenip kart içinde kırpılarak daha yakın gösterildi. Kapsam: `internal/render/{army_panel.go,recruit_panel.go}`, mevcut ortak viewport geometry testleri.
- 2026-07-17: Birim kartları 210×360 sprite oranına göre yeniden boyutlandırıldı; sprite artık üstten başlayarak tamamı çiziliyor, isim/HP/progress ve kuyruk iptal butonu görselin üzerine bindiriliyor. Kapsam: `internal/render/{army_panel.go,recruit_panel.go,hover_tooltip.go}`.
- 2026-07-17: Birim kartlarında sprite üzerine gelen alt etiket alanı kart çerçevesi içinde opak beyaz footer olarak çiziliyor; yazı ve HP/progress okunabilirliği artırıldı. Kapsam: `internal/render/{army_panel.go,recruit_panel.go}`.
- 2026-07-17: 1300 senaryosunda askeri birim görselleri tek sheet yerine birim türü başına ayrı asset'lere taşındı; doğu/batı sprite setleri din grubuna göre seçiliyor ve eski senaryolar için sheet fallback'i yalnızca yeni asset klasörleri yoksa yükleniyor. Kapsam: `assets/scenarios/1300_ottoman_rise/sprites/{eastern_army,western_army}`, `internal/render/{recruit_panel.go,army_panel.go,hover_tooltip.go,renderer.go}`.

## Denetim Özeti (2026-05-08)

Proje artık oynanabilir dikey kesite yakın: ana menüden senaryo seçiliyor, fraksiyon ve zafer koşulu seçilip kampanya başlıyor, tur döngüsü çalışıyor, AI turu işleniyor, harita ve paneller render ediliyor, kayıt/yükleme slotları var.

Mevcut veri seti iki senaryoda da aynı genişlikte:

| Senaryo | Bölge | Deniz | Fraksiyon | Oynanabilir | Başlangıç ordusu |
|---|---:|---:|---:|---:|---:|
| `1300_ottoman_rise` | 210 | 52 | 45 | 30 | 49 |

Doğrulama: `go test ./...` WSL ortamında 2026-05-08 tarihinde başarıyla çalıştı.

## Son Güncellemeler

- 2026-07-16: Seçili kuşatma ordusunun kuşatma özeti yalnız `Kuşatma Emri` paneline taşındı. Hedef, durum, tahkimat, ilerleme, gedik, teslim süresi ve genel hücum uygunluğu ayrı bilgi kartlarında daha okunaklı gösteriliyor; alt ordu panelindeki tekrar eden kuşatma footer'ı kaldırıldı. Kapsam: `internal/render/{renderer.go,renderer_dialogs.go,army_panel.go}`, testler: `go test ./...`.
- 2026-07-17: Kuşatma ordusu bölünüp parçalar birleştiğinde, kuşatılan bölgede aynı fraksiyona ait en az bir canlı kara ordusu kalıyorsa kuşatma artık korunuyor ve kayıt hayatta kalan parçaya devrediliyor. Split sonrası ayrılan parça doğrudan seçili bırakılıyor; ikon hit-test sırası da kuşatma ordusunu ayrılan parçadan deterministik ayırıyor. Kapsam: `internal/game/{game.go,siege.go,siege_test.go}`, `internal/render/{renderer.go,army_siege_split_test.go}`.
- 2026-07-16: Ordu hareket havuzu birim kompozisyonuna bağlandı. `units.json` içindeki `movement_points` değeri okunuyor; yalnız süvari orduları `3`, piyade orduları `2`, kuşatma/topçu orduları `1` puan alıyor ve karışık ordular en yavaş birime göre ilerliyor. Mevsim çarpanı önce bu tabana uygulanıyor, komutan/teknoloji ve legacy senaryo zorluk bonusu sonrasında ekleniyor; 1300 senaryosunda `fair_movement` oyuncu/AI hesabını eşitliyor. Taşınan kara birlikleri filonun hızını etkilemiyor. Kapsam: `internal/army/{unit.go,army.go,loader.go}`, `internal/state/movement.go`, `internal/game/{resolution.go,game.go}`, `assets/scenarios/*/data/units.json`, testler: `go test ./internal/army ./internal/game ./internal/ai ./internal/save`.
- 2026-07-16: Aktif event bölge marker'ları ana harita ve minimap'ten kaldırıldı. Event bilgisi artık yalnız seçilen bölgenin bilgi panelindeki `AKTİF OLAYLAR` bölümünde olay adı, tip ve kalan tur ile gösteriliyor; marker'a bağlı tıklama/hover akışı da kaldırıldı. Kapsam: `internal/render/{panel.go,renderer.go,renderer_input.go,cursor.go}`, `wiki/{systems/events.md,architecture/render-pipeline.md}`, test: `go test ./internal/render`.
- 2026-07-15: 1300 senaryosunun `commanders.json` verisi tarihsel fraksiyon liderleriyle dolduruldu. 46 fraksiyon ID’si için döneme uygun lider/komutan kayıtları ve oyun içi başlangıç trait’leri eklendi; Osmanlı havuzunda Osman Gazi, Konur Alp ve Akça Koca bulunuyor. 1300 öncesi/sonrası sınırı belirgin olan Arnavut Despotluğu, Canik, Dulkadir, Ramazan ve Eflak için anakronik isim eklenmedi; bu kayıtlar fallback havuzunu kullanıyor. Portre asset yolları her kayıt için hazır bırakıldı. Kapsam: `assets/scenarios/1300_ottoman_rise/data/commanders.json`.
- 2026-07-15: 1300 senaryosunda Osmanlı, Fransa, Kutsal Roma İmparatorluğu, İngiltere, Venedik, Moskova Rusyası ve Ceneviz komutan havuzları genişletildi. Toplam 73 kayıt içinde bu yedi havuz 4–8 komutan aralığına çıkarıldı; her yeni kayıt başlangıç trait’i ve gelecekte kullanılacak `portrait_asset` yolunu taşıyor. Kapsam: `assets/scenarios/1300_ottoman_rise/data/commanders.json`.
- 2026-07-15: Komutan atama UI’ı portre ve doğrudan seçim desteği kazandı. `KOMUTAN ATA` paneli artık yalnız boşta komutanları listeliyor; portre kartına tıklama veya klavye seçimiyle belirli komutan atanıyor, `portrait_asset` görselleri senaryo `sprites/` altında cache'leniyor ve eksik asset için yer tutucu gösteriliyor. Kapsam: `internal/render/{commander_panel.go,renderer.go}`, `internal/render/army_panel_test.go`, testler: `go test ./...`.
- 2026-07-15: Ordu detay paneli komutan odaklı iki kolon düzenine taşındı. Sol sütunda komutan portresi, seviye/XP, savaş-zafer istatistikleri, saldırı-savunma katkısı ve trait özeti; sağ sütunda birlik kartları gösteriliyor. Komutan aksiyonu da sol kartın içine alındı, panel genişliği yeni yerleşime göre büyütüldü. Kapsam: `internal/render/army_panel.go`, `internal/render/ui_geometry_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render -run 'Test(CommanderActionLabel|ScoutedEnemyRevealCount|UIGeometry|ArmyPanel)'`.
- 2026-07-15: Ordu paneli komutan aksiyonu görünürlük ve input davranışı güncellendi. `KOMUTAN ATA / KOMUTAN DEĞİŞTİR` butonu artık `BÖL` ile aynı aktif buton stilini kullanıyor; ayrıca panelin boş alanı tıklamayı haritaya düşürmediği için ordu paneli yalnız kapatma düğmesiyle veya yeni seçim akışıyla kapanıyor. Kapsam: `internal/render/{army_panel.go,renderer_input.go,cursor.go}`, `internal/render/army_panel_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render -run 'Test(CommanderActionLabel|ScoutedEnemyRevealCount|ArmyPanelBoundsHit|UIGeometry|ArmyPanel)'`.
- 2026-07-15: Görünür düşman orduların paneline de komutan bilgisi eklendi. Basit `Rakip Ordu` paneli ve kısmi/tam istihbaratlı düşman birlik paneli artık ayrı komutan strip’i çiziyor; atanmışsa portre, isim, trait/meta özeti ile muharebe ve operasyon katkıları, yoksa komutansız durumu görünüyor. Hit-test rect’i de yeni yüksekliklerle eşitlendi. Kapsam: `internal/render/army_panel.go`, `internal/render/army_panel_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-15: Düşman ordu paneli oyuncu ordusuyla aynı iki kolon tasarıma taşındı. Komutan solda, birlik alanı sağda; tam/kısmi istihbaratta gerçek veya gizli kartlar aynı ızgarada çiziliyor. Bu değişiklikle overlap/hit sözleşmesi `armyPanelGeometry()` üstünde birleşti ve `KOMUTAN ATA` düğmesi de daha parlak altın stile alındı. Kapsam: `internal/render/army_panel.go`, `internal/render/army_panel_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-15: Komutan uzmanlıkları hareket, kuşatma ve morale bağlandı. Mevcut trait’lerden `Taktisyen` artık tur başı hareket `+1`, `Saldırgan` kuşatma ilerleme ve gedik kazanımı `+1/+1`, `Savaş Tecrübesi` ile `Savunma Uzmanı` toplamda moral dayanıklılığı artırıyor; bu etkiler combat resolve/preview, turn başı hareket reset’i ve kuşatma tick’lerinde gerçek hesaba giriyor. Ordu ve komutan panelleri de yeni katkıları gösteriyor. Kapsam: `internal/army/{commander.go,army.go}`, `internal/combat/combat.go`, `internal/game/{resolution.go,siege.go}`, `internal/render/{army_panel.go,commander_panel.go}`, `wiki/{systems/combat.md,architecture/render-pipeline.md}`, testler: `go test ./internal/army -run 'Test(Commander|ArmyCommanderAssignment)'`, `go test ./internal/combat -run 'Test(Commander|PreviewBattle|ApplyCasualties)'`, `go test ./internal/game -run 'Test(ApplySeasonEffectsAdds(Commander|Naval)MoveBonus|CommanderSiegeBonusesIncreaseProgressAndBreachGain)'`, `go test ./internal/render -run 'Test(CommanderActionLabel|ArmyPanelBoundsHit)'`.
- 2026-07-15: Komutan etkileri savaş ve kuşatma karar UI’ında görünür hale getirildi. `Savaş Planı` modalı artık saldıran ve savunan komutanın gerçek muharebe bonuslarını (`saldırı/savunma/moral`) üst bantta gösteriyor; `Kuşatma Kararı` mesajı ve aktif `Kuşatma Emri` paneli de komutanın `moral / hareket / kuşatma` katkılarını yazıyor. Aynı değişiklikte çıkarma savaşı preview’si gerçek `EmbarkedCommander` üzerinden hesaplanacak şekilde düzeltildi. Kapsam: `internal/render/{renderer.go,renderer_dialogs.go,commander_panel.go,ui_modals.go}`, `internal/render/{renderer_naval_transport_test.go,war_confirm_test.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render -run 'Test(OpenBattlePlanUses|OpenSiegeDecisionIncludesCommanderOperationalSummary)'`.
- 2026-07-15: Komutan trait gösterimi rozet tasarımına taşındı. Oyuncu ordu paneli, görülen düşman ordu paneli ve komutan detay profilinde `Savaş Tecrübesi / Taktisyen / Savunma Uzmanı / Saldırgan` artık renk kodlu kompakt badge olarak çiziliyor; etki özeti bu rozetlerin altına yeniden hizalandığı için örtüşme giderildi. Kapsam: `internal/render/{army_panel.go,commander_panel.go,commander_panel_test.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-15: Komutan UI çizimi ortak komponent helper’larına toplandı. Yeni `commander_component.go`, trait badge stillerini ve komutan özet kartı chrome’unu tek yerde tutuyor; oyuncu ordusu, düşman ordusu, komutan detay profili ve komutan atama listesindeki dar satırlar aynı tasarım kodunu paylaşıyor. Liste satırları tek satıra sığmayan uzmanlıklar için ortak `+N` overflow rozeti kullanıyor. Kapsam: `internal/render/{army_panel.go,commander_panel.go,commander_component.go,commander_panel_test.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-15: Savaş raporu komutan bloğu da ortak commander component katmanına bağlandı. Battle report taraf kartlarındaki portre + isim + muharebe/operasyon satırları artık `commander_component.go` içindeki kompakt strip helper’ını kullanıyor; böylece ordu paneli, komutan paneli ve savaş raporu aynı çizim dilini paylaşıyor. Kapsam: `internal/render/{battle_report.go,commander_component.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-15: Ordu paneli komutan kartındaki taşma düzeltildi. Alttaki tekrar effect summary satırı panel varyantında kapatıldı, trait rozetleri tek satıra sınırlandı ve oyuncu kartında `KOMUTAN ATA / DEĞİŞTİR` butonu için alt boşluk reserve edildi; böylece rozet/metin/buton çakışması giderildi. Kapsam: `internal/render/{army_panel.go,commander_component.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-16: Ordu paneli komutan kartı okunurluğu tekrar sıkılaştırıldı. Komutan katkıları tek uzun satır yerine dar alana sığan ayrı satırlara bölündü; saldırı/savunma/moral/hareket/kuşatma değerleri artık yatay taşmıyor. Uzmanlık rozetleri CamelCase etikete çevrildi ve badge içinde dikey ortalandı. `Komutan Ata / Komutanı Değiştir` butonu da daha parlak dolgu, koyu yazı ve hafif gölgeyle belirginleştirildi. Kapsam: `internal/render/{army_panel.go,commander_component.go,army_panel_test.go,commander_panel_test.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-16: Komutan kartı ayırıcı çizgisi dinamik hale getirildi. Komutanın katkı satırı sayısı arttığında separator artık sabit Y’de kalmayıp son katkı satırının altına iniyor; böylece `Hareket` veya `Kuşatma` satırlarının üzerine binmiyor. Kapsam: `internal/render/{commander_component.go,commander_panel_test.go}`, testler: `go test ./internal/render`.
- 2026-07-16: Ordu panelindeki komutan aksiyon butonu kaldırıldı. `Komutan Ata / Değiştir` işlevi artık doğrudan komutan portresine tıklamaya bağlı; ayrı alt buton render’ı, button hit-test’i ve geometri sözleşmesi portre rect’iyle sadeleştirildi. Kapsam: `internal/render/{army_panel.go,renderer_input.go,ui_geometry_test.go,army_panel_test.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-16: Komutan atama listesindeki satır sıkışıklığı azaltıldı. Sol listedeki row yüksekliği artırıldı, satır içi komutan portresi 56px’ten 64px’e büyütüldü ve isim/XP/uzmanlık konumları yeni yüksekliğe göre yeniden hizalandı. Kapsam: `internal/render/commander_panel.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`.
- 2026-07-16: Cursor hover alanları daraltıldı. Komutan atama panelinde pointer yalnız close/ayırma düğmeleri ve komutan listesi satırlarında görünür; ordu panelinde ise pointer artık tüm panel bounds’unda kalmaz, yalnız gerçek tıklanabilir alanlarda (`BÖL`, `BİRLEŞTİR`, komutan portresi) görünür. Kapsam: `internal/render/{cursor.go,commander_panel.go,army_panel_test.go}`, testler: `go test ./internal/render`.
- 2026-07-16: Ordu panelindeki `BÖL` ve `BİRLEŞTİR` butonları başlık bandındaki `Hareket` ve `Takviye aktif` metinlerini kapatmayacak şekilde daha sola alındı. Buton rect’leri için sağda 92 px rezervasyon yapıldı ve geometri regresyon testi eklendi. Kapsam: `internal/render/{army_panel.go,ui_geometry_test.go}`, testler: `go test ./internal/render`.
- 2026-07-21: Ordu paneli aksiyon sırası düzeltildi. `BÖL` artık her durumda en sağdaki düğmede, `BİRLEŞTİR` görünürse onun solunda çiziliyor; çizim ve hit-test aynı geometri helper’larını kullanıyor. Kapsam: `internal/render/{army_panel.go,ui_geometry_test.go}`, `wiki/architecture/render-pipeline.md`, test: `go test ./internal/render`.
- 2026-07-16: Ordu paneli komutan kartı çerçevesi profil satırları ve uzmanlık taşma rozeti için genişletildi. Komutan kolonu 260 px’e çıkarıldı, kart yüksekliğine 12 px ek pay verildi ve seviye/XP ile savaş/zafer satırları kullanılabilir metin genişliğine göre kırpılıyor. Kapsam: `internal/render/{army_panel.go,commander_component.go,commander_panel_test.go}`, testler: `go test ./internal/render`.
- 2026-07-15: Komutan portre yolu kısa JSON formatına göre düzeltildi. `portrait_asset` artık yalnız dosya adıysa otomatik `sprites/commanders/` altında çözülüyor; eski `commanders/...` relative yollar da çalışmaya devam ediyor. Bu düzeltme sayesinde VS Code debug akışında 1300 Osmanlı komutan atama panelindeki `Yok` placeholder'ı gerçek portrelerle değişiyor. Kapsam: `internal/render/commander_panel.go`, `internal/render/commander_panel_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render -run TestCommanderPortraitPath`.
- 2026-07-15: Komutan başlangıç verileri senaryo dışı hardcode olmaktan çıkarıldı. `data/commanders.json` şablonları isim, başlangıç trait'i ve gelecekteki portre asset yolunu taşıyor; scenario/save-base load akışları bunları runtime komutan havuzuna clone ediyor, dosya yoksa fallback isimlendirme korunuyor. Kapsam: `internal/army/{commander.go,commander_loader.go}`, `internal/state`, `internal/game/game.go`, `internal/save/save.go`, `assets/scenarios/1300_ottoman_rise/data/commanders.json`, testler: `go test ./...`.
- 2026-07-15: Nakliye filosundaki taşınan kara komutanı UI’a bağlandı. Filo başlığında taşınan komutan görünür; komutan paneli filo komutanı ile kara komutanını ayrı profillerde gösterir ve taşınan komutanı filodan ayırma aksiyonu sunar. Kapsam: `internal/render/{army_panel.go,commander_panel.go,action.go}`, `internal/game/game.go`, test: `go test ./...`.
- 2026-07-15: Komutan yaşam döngüsü tamamlandı. Ordu silme işlemleri `GameState.RemoveArmy()` üzerinden komutanı havuza bırakıyor; eliminasyon/fetih sonrası devralınan ordularda komutan `OwnerID` aktarılıyor. Nakliye filosu `EmbarkedCommander` ile kara komutanını koruyor, çıkarma sonrası yeni orduya geri bağlıyor ve başarısız çıkarmada serbest bırakıyor; çıkarma savaşı artık gerçek komutanın bonus/XP hattını kullanıyor. Kapsam: `internal/army`, `internal/state`, `internal/game`, `internal/ai`, `internal/save`, testler: `go test ./internal/state ./internal/game ./internal/ai ./internal/save`.
- 2026-07-15: Savaş raporuna komutan kariyer bildirimi eklendi. Her gerçek katılımcı için kazanılan XP, seviye değişimi ve açılan trait'ler modalda `Komutan gelişimi` bölümünde ve savaş olayının detayında gösteriliyor; birleşik savunmada her gerçek savunucu komutanı ayrı raporlanıyor. Kapsam: `internal/army/commander.go`, `internal/game/{commander.go,battle_report.go,game.go}`, `internal/render/battle_report.go`, testler: `go test ./internal/army ./internal/game ./internal/render`.
- 2026-07-15: Savaş raporu komutan katkılarını da görünür gösteriyor. Taraf kartlarına ayrı `Komutan` bloğu eklendi; komutan portresi, adı, muharebe etkileri (`saldırı/savunma/moral`) ve operasyon etkileri (`hareket/kuşatma`) artık savaş sonucu istatistikleriyle aynı kartta okunabiliyor. Veri game katmanında battle snapshot’a taşındı, böylece yok olan savunucu ordularda bile komutan özeti ve portresi raporda korunuyor. Kapsam: `internal/game/battle_report.go`, `internal/render/{battle_report.go,commander_panel.go}`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/game -run TestBattleReportSideCarriesCommanderEffectSummaries`, `go test ./internal/render -run 'TestBattleReportCommander(Progress|EffectTexts)'`.
- 2026-07-15: Komutan kariyer dengesi ayarlandı. Seviye eşikleri `100/300/550/850 XP` olarak kademelendirildi; maksimum kariyer yaklaşık 9 zaferde tamamlanıyor. Trait bonusları maksimumda saldırı `%12`, savunma `%10` olacak şekilde yumuşatıldı ve milestone/modifier cap testleri eklendi. Kapsam: `internal/army/commander.go`, `internal/{combat,save}/**_test.go`, testler: `go test ./internal/army ./internal/combat ./internal/save`.
- 2026-07-14: `Çetin Kış` gibi `all_armies` hedefli event harita izleri artık deniz bölgelerinde çizilmiyor; yeni kayıtlar kara bölgeleriyle sınırlanıyor, mevcut save'lerdeki deniz event kayıtları render/minimap/hit-test tarafında gizleniyor. Marker anchor'larındaki world-pixel koordinatına hatalı biçimde ikinci kez shape dönüşümü uygulanması kaldırıldı; çizim ile hit-test aynı ekran-anchor kontratını kullanıyor. Aynı bölgedeki çoklu event stack'i de bölge dışına taşarsa en yakın bölge pikseline geri clamp ediliyor. Kapsam: `internal/events/events.go`, `internal/render/{renderer.go,panel.go}`, testler: `go test ./internal/events -run TestAffectedRegionIDsAllArmiesSkipsSeaRegions`, `go test ./internal/render -run 'Test(EventIcon|ActiveRegionEvent|ClampActiveRegionEvent|MinimapEvent)'`.
- 2026-07-14: Yerleşim üstündeki aktif event marker'lar artık ordu sayı karesiyle aynı Y noktasını paylaşmıyor; marker daha yukarı taşınıyor ve input önceliği army > event > settlement olarak işliyor. Kapsam: `internal/render/renderer.go`, `internal/render/renderer_input.go`, `internal/render/event_visibility_test.go`, `wiki/architecture/render-pipeline.md`, testler: `go test ./internal/render`, `go test ./...`.
- 2026-07-13: Devlet ismi üzerinden açılan devlet bilgi paneline scroll'lu diplomasi özeti eklendi; panel artık doğrudan üst devlet, vassal, ittifak, ticaret ve düşman listelerini gösteriyor ve uzun içerikte mouse wheel ile kaydırılabiliyor. Kapsam: `internal/render/panel.go`, `internal/render/renderer.go`, `internal/render/renderer_input.go`, test: `go test ./...`.
- 2026-07-13: Kuşatma altındaki tahkimli bölgelere üçüncü devletlerin yeni kuşatma başlatması artık bloklanıyor; ancak bölgeye giriş hakkı olan bir ordu, kuşatma yapan düşman orduyu savaşla yenebiliyorsa kuşatma kalkıyor ve müttefik/same realm bölgelerde sahiplik değişmeden kalıyor. AI aynı kuralı kullanıyor: tahkimli hedefte doğrudan fetih yerine önce kuşatma kuruyor, aktif kuşatmaya dışarıdan destek ise ayrı hareket akışıyla çözülüyor. Kapsam: `internal/game`, `internal/ai`, `internal/render`, `wiki/{architecture,systems}`, testler: `go test ./...`.
- 2026-07-12: Save/load mimarisi ham `GameState` snapshot'ından kampanya delta + senaryo overlay modeline geçirildi. Kayıt dosyaları artık statik senaryo tanımlarını (`map`, `trade_centers`, bölge/fraksiyon isimleri, shape kimlikleri ve `region_shapes.json` kaynaklı region paint override verisi vb.) tekrar yazmıyor; yüklemede senaryo baz state'i yeniden kurulup mutable alanlar üstüne uygulanıyor. Format ayrıca relation delta, region delta, settlement patch, stacked army unit kaydı ve `state_zstd` (`zstd+base64`) sıkıştırması kullanıyor. Slot kartı için gereken metadata düz JSON `meta` alanında ayrı tutuluyor. `DEV_MODE=true` olduğunda aynı slot için `*.debug.json` sidecar'ı da açıklayıcı alan adlarıyla yazılıyor; normal mod save alındığında debug sidecar bırakılmıyor. Legacy save'lerde eksik alanlar senaryo varsayılanlarını ezmiyor. Kapsam: `internal/save/save.go`, `internal/save/compact.go`, `internal/save/save_test.go`, `wiki/architecture/state-management.md`, testler: `go test ./internal/save`, `go test ./internal/render ./internal/game ./internal/save`, `go test ./...`.
- 2026-07-11: 1300 senaryosunda boş kalan Aksaray ve Arab Çölü bölgelerine döneme uygun yerleşimler eklendi; artık gerçek bir bölgeye bağlı olmayan boş `ankara` kaydı temizlendi. Kara bölgelerinin yerleşimsiz kalmasını ve settlement kayıtlarının bilinmeyen bölgelere bağlanmasını engelleyen senaryo veri testi eklendi. Kapsam: `assets/scenarios/1300_ottoman_rise/data/settlements.json`, test: `go test ./internal/world -run Test1300ScenarioEveryLandRegionHasSettlement`.
- 2026-07-11: AI savaş ilanı sonrası oyuncu müttefik çağrısı akışı eklendi; müttefik AI saldırı başlatınca veya müttefik AI saldırıya uğrayınca oyuncuya `Savaşa Katılım Çağrısı` düşüyor. Kabulde oyuncu ilgili savaşa katılıyor, ret halinde ittifak bozulup ilişki skoru düşüyor; aktif AI deklaratörü oyuncu cevabını bekliyor. Kapsam: `internal/diplomacy`, `internal/game`, `internal/render`, `wiki/systems/{diplomacy,ai}.md`, testler: `go test ./internal/diplomacy`, `go test ./internal/game -run 'TestAcceptedWarJoinOfferEndsDeclarerAITurn|TestAcceptedOfferEndsCurrentAIFactionTurn|TestRejectedOfferAppendsDiplomacyHistory|TestPendingPlayerDiplomacyOfferUsesPriority'`, `go test ./internal/render -run 'TestDiplomacyOfferActionLabelTRWarJoinCall|TestDiplomacyOfferMessageTRWarJoinCall|TestHandleDiplomacyOfferInput'`.
- 2026-07-11: Savaş ilanı onaylandıktan sonra ayrı `Savaş Özeti` modalı eklendi; savaşa gerçekten kimlerin katıldığı, çağrıyı reddeden müttefikler ve iki tarafın toplam askeri gücü gösteriliyor. Eğer aynı akışta battle report oluşursa battle report bu özet kapanana kadar kuyruğa alınıyor. Kapsam: `internal/render`, `internal/game`, `internal/diplomacy`, testler: `go test ./internal/render ./internal/game ./internal/diplomacy`.
- 2026-07-11: Savaş ilanı artık koalisyon önizleme modalından geçiyor; hedefin vassalları, iki tarafın çağrılabilir müttefikleri ve katılım olasılıkları gösteriliyor. Oyuncu kendi müttefiklerini tek tek çağırabiliyor; çağrıyı reddeden müttefiğin ittifakı bozulup ilişki skoru düşüyor. Kapsam: `internal/diplomacy`, `internal/render`, `internal/game`, testler: `go test ./internal/diplomacy ./internal/render ./internal/game`.
- 2026-07-11: AI ittifak mantığı coğrafi/stratejik yakınlık filtresi, AI müttefik kapasitesi ve `ai_expansion_targets` gerilimiyle sıkılaştırıldı; kara sınırı, aktif ticaret, ortak düşman veya ortak büyük tehdit olmadan alliance açılmıyor. Desteksiz dış ittifakların relation skoru artık otomatik şişmiyor, anlamını kaybeden alliance AI tarafından bozulabiliyor. Kapsam: `internal/ai`, `internal/diplomacy`, `wiki/systems/{ai,diplomacy}.md`, testler: `go test ./internal/ai ./internal/diplomacy`.
- 2026-07-11: Dış diplomasi için temas ve hat şartı sertleştirildi; alliance artık `Score >= 25` ve gerçek diplomatik temas istiyor, trade ise `Score >= 15` ve bağlanabilir kara/deniz ticaret hattı olmadan açılamıyor. `SanitizeTradeRoutes()` kopmuş dış rotaları da temizliyor. Kapsam: `internal/diplomacy`, `internal/ai`, `wiki/systems/{ai,diplomacy}.md`, testler: `go test ./internal/diplomacy ./internal/ai`.
- 2026-07-11: AI alliance teklifinde “anlamlı fayda” filtresi eklendi; büyük güçler tek bölgeli/zayıf devlete artık otomatik alliance atmıyor, yalnız gerçek askeri katkı, sınır tamponu veya ortak tehdit menfaati varsa teklif açıyor. Aynı mantık faydasız mevcut alliance’ların bozulmasına da yardım ediyor. Kapsam: `internal/ai`, `wiki/systems/{ai,diplomacy}.md`, testler: `go test ./internal/ai ./internal/diplomacy ./internal/game`.
- 2026-07-11: Dost veya müttefik kara bölgesinde duran müttefik ordular artık `SelectBattleDefender` tarafından savunucu seçilmiyor; bu yüzden oyuncu kendi/dost toprağına hareket ederken yanlış `Savaş Planı` modalı açılmıyor. Kapsam: `internal/state/state.go`, testler: `go test ./internal/state ./internal/render`.
- 2026-07-12: Haritada seçilen settlement etiketi artık altın tonla vurgulanıyor; yoğun kümelerde mouse ile seçilen yerleşim adı daha belirgin görünüyor. Kapsam: `internal/render/renderer.go`, `wiki/architecture/render-pipeline.md`.

## Tamamlanan Sistemler

| Sistem | Durum | Notlar |
|---|---|---|
| Ebitengine kurulum | ✅ | `cmd/game/main.go`, 60 TPS |
| GameState merkezi yapı | ✅ | `internal/state/state.go` |
| Phase state machine | ✅ | 12 faz: ana menü, ayarlar, senaryo, fraksiyon, zafer, oyun, AI, çözümleme, game over, pause, load, save |
| Senaryo sistemi | ✅ | `internal/scenario/scenario.go`; `assets/scenarios/scenarios.json` index + bağımsız senaryo klasörleri |
| Senaryo seçim ekranı | ✅ | `internal/render/scenario_select.go`, `PhaseScenarioSelect` |
| Harita render | ✅ | `WorldMap` cache, ülke/deniz şekilleri, sahiplik rengi ve seçili bölge vurgusu; normal modda oyuncu+vassal realm dış konturu tek piksel keskin altın, müttefik realm'ler doygun turkuaz-yeşil, savaştaki düşman realm'leri doygun kırmızı çizilir ve tüm tarafların vassalları kendi realm konturuna dahil edilir. Ortak sınırlar iki taraftan üst üste boyanmaz. Bölge/realm içi idari sınırlar aynı tek piksel kalınlıkta fakat araziye düşük oranlı blend ile daha soluktur; diplomasi imzası ittifak/vassallık/savaş değişince cache'i otomatik yeniler |
| Senaryo bazlı harita hizalama | ✅ | `scenario.json` içindeki `map` alanı `WorldW/WorldH` ve shape offset/scale değerlerini belirler |
| Görsel mevsim değişimi | ✅ | `internal/render/mapgen.go:applyOwnership`; kış/ilkbahar/sonbahar tint |
| Bölge sistemi | ✅ | JSON'dan yükleme, komşuluk grafı, kilitli bölge alanları |
| Fraksiyon sistemi | ✅ | 45 fraksiyon, senaryo bazlı oynanabilir roster; 1444 senaryosunda yalnız tarihsel hedefi olan 6 fraksiyon, 1512 senaryosunda ise yalnız tarihsel hedefi olan 5 fraksiyon açılıyor |
| Din paketi | ✅ | `internal/religion`; `catholic`, `orthodox`, `sunni`, `shia` ilişki puanları |
| Ordu hareketi | ✅ | Komşuluk kısıtı, kara/deniz giriş kontrolü, savaş öncesi diplomasi kontrolü; kara ordusu düşman kara ordusuna girmeden önce `Savaş Planı` modalında `Agresif / Dengeli / Savunmacı` duruş seçer ve bu seçim gerçek resolve'e taşınır. Aynı modal artık savaş halindeki düşman donanma üstüne girildiğinde `Deniz Muharebesi`, savunan ordu olan düşman kıyıya çıkarma sırasında ise `Çıkarma Muharebesi` başlığıyla açılır; seçilen duruş sırasıyla `ActionMoveArmy` ve `ActionDisembarkArmy` üzerinden çözülür. Resolve tamamlanınca oyuncu için ayrı `Savaş Raporu` modalı açılır; sonuç, duruş, `Güç/Birim/HP` kırılımı ve kullanıcı sonradan ekleyebileceği `assets/ui/battle_<scene>.png` ile `assets/sounds/battle_<scene>.wav` hook'ları aynı yüzeydedir. Hedef bölgede savunan ordu yoksa ama sahiplik el değişiyorsa, aynı rapor bu kez `Direniş Görülmedi` sonucu ile yine açılır. Donanmalar deniz bölgeleri arasında savaş ilanı olmadan dolaşır, deniz çatışması sadece `StanceWar` durumunda tetiklenir; AI dost bölgelerde bölgesel ikmal baskısını okuyup aşırı dolu kara bölgelerden daha rahat komşu bölgelere dağılabilir |
| Kuşatma sistemi | ✅ | Tahkimli kara bölgeler (`fortress` settlement veya `walls`) artık savaşsız anında düşmez. Normal kara orduları da kuşatma kurabilir; kuşatma birimi olmayan orduda `Genel Hücum` kapalıdır ve kale yalnız aç bırakma / teslimiyet süreciyle düşebilir. Sağ tık akışı savaş ilanından sonra `Kuşatma Kararı` modalına bağlanır; `Kuşatma Başlat`, `Genel Hücum` (uygun ordularda), aktif kuşatmada `Kuşatmayı Kaldır` seçenekleri vardır. Kuşatma durumu `GameState.Sieges` içinde kaydedilir, kuşatan ordu render'da kuşatılan bölgenin üstünde görünür ve bölge kuşatanın rengiyle taralı muallak overlay alır. Seçili kuşatma ordusunda alt-orta `Kuşatma Emri` paneli ve ordu üstünde kılıç rozeti görünür; başka komşu bölgeye verilen normal hareket emri eski kuşatmayı otomatik kaldırır. Gedik açılması artık kuşatma ekipmanı tier'i ile kale seviyesi birlikte hesaplanır; düşük tier araçlar aç bırakma kuşatmasını sürdürebilir, yüksek seviye surlarda bile çok yavaş gedik açabilir. Genel hücum artık gedik yokken tahkimatı doğrudan ele geçiremez; ayrıca saldıran taraf gedik küçüldükçe daha ağır hücum zayiatı öder. Aktif kuşatmaya yalnız aynı fraksiyon ya da müttefik devlet destek için katılabilir; ilgisiz üçüncü devletler ikinci bir kuşatma hamlesi yapamaz. Kuşatma altındaki bölgeye üçüncü devlet yeni kuşatma başlatamaz; fakat bölgeye giriş hakkı olan bir ordu, kuşatma yapan düşman orduyu savaşta yenerek kuşatmayı kaldırabilir. AI de tahkimli hedeflerde doğrudan fetih yerine kuşatma açar |
| Deniz taşıma akışı | ✅ | Kara ordusu uygun `transport` filosuna binebilir, filo `EmbarkedUnits` ile taşır, komşu dost/boş karaya çıkarma yapılır; oyuncu ve AI aynı kural setini kullanır. Sahipsiz kıyıya amfibi çıkarmada bölge artık otomatik sahiplenilir; bug'lı eski save'lerde sahipsiz kalmış ama tek taraflı işgal altında duran kara bölgeleri yükleme/tur çözümlemesinde toparlanır. `transport` birimleri artık `carry_capacity` ile gerçek slot kapasitesi taşır, aynı filo kapasite yettikçe birden fazla kara birimini biriktirebilir ve oyuncu seçili orduyla dost nakliye filosuna sağ tıklayınca `Gemiye Bin` onayı alır. Gemide birlik taşıyan filo limana uğramadan 3 turdan fazla açık denizde kalırsa taşınan birlikler her tur artan zayiat alır |
| Boğaz deniz geçiş sürekliliği | ✅ | Senaryo verilerinde Marmara-Ege-Karadeniz deniz komşuluğu çift yönlü korunur; filolar `Ege -> Marmara -> Karadeniz` ve ters yönde komşuluk bazlı ilerleyebilir, bu köprü testi `internal/world/scenario_sea_adjacency_test.go` ile sabitlenmiştir |
| Amfibi savaş fazı | ✅ | Düşman kıyıya çıkarma savaş halinde aktif; çıkarma anı çatışması `combat` ile çözülür, başarılı çıkarma karaya ordu indirip sahiplik günceller, AI barışta çıkarma denemez |
| Başlangıç orduları | ✅ | Her senaryonun `data/armies.json` dosyasından yükleniyor |
| Çarpışma motoru | ✅ | Birim gücü, arazi, teknoloji modları ve rastgele sonuç etkisi; saldırı duruşu (`agresif/dengeli/savunmacı`) gerçek savaş sonucu ve saldırı öncesi preview hesabında aynı combat helper'larıyla işlenir. `land/naval/amphibious` bağlamları ayrı stance çarpanları ve açıklama metinleri taşır; muhtemel kayıp paneli bu bağlama göre hesaplanır |
| Komutan kariyeri | ✅ | `Army.Commander` çekirdeği, dengelenmiş XP/seviye/trait ilerlemesi, savaş gücüne saldırı-savunma etkisi, save/load, üç kişilik oyuncu havuzu, ordu panelinden atama/ayırma, AI saha ordularına deterministik komutan üretimi, birleşme-garnizon yaşam döngüsü ve savaş raporu/olay günlüğü kariyer bildirimi hazır |
| Savaş sonrası toparlanma | ✅ | Savaş, lojistik ve diğer HP kayıpları artık kısmi hasar bırakır; kara orduları kendi kara toprağında tur başına `+10 HP` ile %100'e kadar toparlanır, limana bağlı donanmalar da kendi veya müttefik limanında aynı hızla onarım alır |
| Ordu detay paneli | ✅ | 20 slot, HP/deneyim çubukları, bölme/birleştirme aksiyonları, dost toprakta toparlanan birimler için küçük `+` rozeti |
| Ordu birleşme | ✅ | Dost bölgede otomatik veya panelden manuel birleşme, 20 birim limiti |
| Ordu bölme | ✅ | Seçili orduyu iki parçaya böler |
| Rakip ordu istihbaratı | ✅ | Menzildeki rakip orduda sayı ve yarım birim listesi görünür; menzil dışı detaylar gizlenir; emir verilemez |
| Çoklu ordu render | ✅ | Aynı bölgede ordular yan yana çizilir |
| Askeri kapasite | ✅ | Kara bölgesi başı 5 + kışla başı 5; ordu sayısı `ceil(kara_bölge/2)`; scenario/save kökenli `garrison` orduları artık bu saha ordusu limitine sayılmaz, hareket/split/merge ile sahaya çıktıklarında normal orduya dönüşür |
| Asker alma | ✅ | Milis hızlı alım + belirli birim alımı; bina/teknoloji/çoklu kaynak/manpower kontrolü; JSON `turns_required` ile üretim kuyruğunda tamamlanır, tekrar tıklanınca iptal edilip kaynaklar iade edilir; aynı bölgede mevcut ordu 20/20 doluysa üretim artık bloke olmaz, tamamlanan birlikler boş slotlu orduya eklenir veya gerekirse ikinci kara ordusu olarak spawn olur; kuşatma altındaki bölgedeki bina ve birim üretim emirleri duraklatılır, kuşatma kalkınca kaldığı tur sayısından devam eder; bölge kaybedilince o bölgedeki üretim emirleri kuyruktan silinir |
| Çoklu eğitim kuyruğu (Total War benzeri) | ✅ | Recruit panelinde birim bazında `- xN +` seçimi, kuyrukta aynı birim için ilk tamamlanma turu görünürlüğü ve tek tıkta çoklu (`xN`) üretim emri; aynı tur tamamlanabilen kara birimi sayısı `max(1,kışla seviyesi)`, deniz birimi sayısı `liman seviyesi` ile sınırlandırılır ve kıyı bölgelerinde bu iki hat birbirinden bağımsız işler. Kapasite dışındaki emirler beklerken `TurnsLeft` düşmez; yalnız o tur aktif kışla/liman slotuna giren emirler ilerler |
| Bina/birim hover bilgisi | ✅ | Kart tooltipleri maliyet, gereksinim, etki/istatistik ve görsel gösterir |
| Deniz birimi | ✅ | Liman ve kıyı koşuluyla filo/deniz birimi üretimini kuyruğa alma; tekrar tıklanınca iptal/iade |
| Liman docking akışı | ✅ | Limana bağlı donanma aynı deniz bölgesine hareket emri aldığında liman bağını bırakıp deniz merkezine çıkar; ayrıca komşu denizden sahibi olunan veya müttefik olunan ve `port` binası tamamlanmış kara bölgesine dock olabilir, deniz `RegionID` korunur ve hareket puanı tüketir. Liman binası tamamlanınca bölgede port settlement yoksa otomatik `Liman` yerleşimi açılır; bu nokta varsa bölge shape'inin gerçek kıyı sınırına yaslanır, shape yoksa merkez-deniz fallback'i kullanılır. Kara orduları dock edilmiş nakliye filosuna limandayken de binebilir ve dock edilmiş aynı kıyı bölgesine tekrar sağ tıklanınca gemideki birlikler karaya indirilebilir. Embark akışındaki `Gemiye Bin` onayına paralel olarak dost/boş kara indirmelerinde `Karaya In` onayı ve hedef üstünde `IN` rozeti görünür; bu onay askerleri doğrudan karaya indirir, liman uygunsa bile gemiyi otomatik dock etmez. Dock edilmiş filoda hem gemiler hem de taşınan kara birlikleri kendi limanında tam, müttefik limanında yarım hızla iyileşir |
| Ekonomi tick | ✅ | Vergi geliri, hasat modu, bina modları, ikincil mallar, taş üretimi, tahıl bakım gideri ve tahıl açığında lojistik HP cezası; ayrıca bölge bazlı ikmal kapasitesi (efektif tahıl + yerleşim tamponu + sınırlı stok desteği) aşılınca aynı bölgede uzun süre yığılan kara orduları kademeli zayiat alır |
| Vergi ayarlama | ✅ | Oyuncu bölgelerinde `.` / `,` ile ±5 |
| Bina inşası | ✅ | JSON bina tipleri, maliyet, arazi ve adet kısıtları; varsayılan üretim kuyruğu; kuyruktaki bina tekrar tıklanınca iptal/iade; kuşatma altındaki bölgede inşaat ilerlemez ve kuşatma kaldırılınca kuyruk kaldığı `TurnsLeft` değerinden devam eder; bölge kaybedilince o bölgedeki inşa emirleri de silinir; liman için kanonik uygunluk kuralı artık literal `terrain=coast` değil, deniz komşuluğunu okuyan `Region.IsCoastal` predicate'i; `walls` artık senaryo verilerinde daha pahalı ve daha uzun inşa ediliyor |
| Kaynak reçete sistemi | ✅ | Birim ve bina üretiminde `grain/iron/timber/stone` tüketimi; UI maliyet satırı ve AI kararları bu modele bağlı |
| Bina seviye sistemi | ✅ | Binalar `max_per_region` kadar seviye alır (Lv1..LvN); panelde `Lv` ve kuyruk adedi görünür, inşa mesajları seviye geçişini (`LvX→LvY`) gösterir; kurulu bina kartları da tıklanabildiği için yükseltme/iptal akışı doğrudan kart üzerinden çalışır; manpower ve üretim kapasitesi kışla seviyesiyle artar |
| Ticaret güzergahları | ✅ | `TradeRoutes` pasif gelir modeli var |
| Teknoloji ağacı | ✅ | Araştırma başlatma, tur sayacı, tamamlanan teknoloji efektleri, ağaç görünümü, seviye bazlı düzen, kategori renkleri, tamamlanmış teknoloji tick badge'leri, araştırma seçimi/değiştirme/vazgeçme, HUD'da aktif araştırma gösterimi, tur bitir uyarısı ve tamamlanma mesajları event loguna ekleniyor; tur bitir araştırma uyarısı artık yalnız gerçekten araştırılabilir teknoloji kaldığında gösteriliyordu ve artık aktif araştırma boş kaldığında turn resolution sonraki bağlı teknolojiyi otomatik başlatıyor; yarım bırakılan araştırmalar pause/resume ile kaldığı yerden sürüyor; 1300 senaryosunun research ağı yeni orta/ileri düğümlerle genişletildi ve daha önce boştaki `market_gold_mod`, `peace_relation_bonus`, `naval_move_bonus`, `reveal_enemy_strength`, `conversion_speed_mod` alanları runtime'a bağlandı; bölge panelindeki sahip devlet adına tıklanınca rakip devletin aktif araştırması, tamamlanan teknolojileri, malları ve ticaret özeti ayrı panelde görülebiliyor |
| Diplomasi | ✅ | `internal/diplomacy` ortak motoru ile savaş/barış/ittifak/ticaret; deterministik kabul-red, ilişki decay'i ve ticaret rotası senkronu |
| İlişki geliştirme ve vassallık | ✅ | Diplomasiye `Heyet`, `Hediye` ve `Vassallık` aksiyonları eklendi; vassal state'ler üçüncü taraflarla doğrudan diplomasi kuramıyor, overlord savaşa girince coalition olarak savaşa çekiliyor, ekonomi tick'inde altın haracı ödüyor ve vassallık kurulduğu anda overlord ile iki yönlü ticaret anlaşması kazanıyor. Aynı realm içindeki devletler arasında kara geçişi, dost kıyıya çıkarma, liman kullanımı ve kuşatma desteği için ayrıca savaş ilanı gerekmiyor. Oyuncunun doğrudan vassalı diplomasi panelinden onayla bağımsız bırakılabilir veya tüm bölgeleri, kuvvetleri, kaynakları ve üretim emirleri devralınarak ilhak edilebilir. Vassal bölgeler bölge bilgi panelinde `Bağlı: <overlord>` satırıyla, oyuncuya bağlı olanlarda ayrıca `Haraç: +X altın/tur` satırıyla ve harita marker'ındaki küçük rozetle ayırt ediliyor. Oyuncu bir devletin son kara toprağını düşürdüğünde battle report sonrası `Savaş Sonrası Düzen` modalı açılıyor ve `İlhak Et` veya `Vassal Yap` seçimiyle fetih kapanıyor |
| Diplomasi paneli modern akış | ✅ | Solda devlet seçimi + sağda teklif paneli; savaş/barış/ittifak/ticaret için kabul olasılığı (%) ve durum göstergesi bulunur. Standart teklif düğmeleri gerçek `ActionBlockReason` sonucuna göre aktif veya `PASİF` çizilir; pasif olanlar fare ve klavyeyle seçilemez. Kurulu dış ittifak ve ticarette aynı düğmeler `İttifakı Bitir / Ticareti Bitir` işlemine dönüşür; ittifak iptali ticareti, ticaret iptali ittifakı korur. Savaş mevcut `Barış`, vassallık `Vasallığı Bitir` akışıyla kapanır. Sağ kolon seçili devletin savaş, dış ittifak ve aktif ticaret ortaklarını listeler; vassal/overlord hiyerarşisini ayrıca belirtir ve doğrudan oyuncu vassalında onaylı `Vasallığı Bitir / İlhak Et` yönetim kartı gösterir. Geçmiş sürekli açık değildir, `Geçmiş / İlişkiler` düğmesiyle aynı kolonda değiştirilir. Bölge panelindeki eski hızlı aksiyonlar tek `Diplomasi` butonuna indirildi. Ticaret yüzdesi gerçek kabul motoruyla aynı helper'dan beslenir; ittifak değerlendirmesinde doğrudan sınır tehdidi ortak düşman veya ortak büyük tehditle dengelenebilir, aynı din kabul bonusu verir. AI aynı helper üzerinden oyuncuya ittifak teklifleri açabilir. `Savaş` aksiyonu ise artık özel koalisyon önizleme modalı açar; hedef vassalları ve iki tarafın müttefik çağrıları burada gösterilir, oyuncu kendi çağrı listesini checkbox ile belirler ve gelmeyen müttefiğin ittifakı otomatik düşer. Resolve sonrasında ayrıca `Savaş Özeti` modalı açılarak gerçek katılımcılar ve toplam güç dengesi gösterilir. History görünümündeki filtreler ve kartlar hover'da değil yalnız tıklamada etkileşime girer |
| Elenen fraksiyon diplomasi temizliği | ✅ | Kara toprağı biten fraksiyonlar (sadece deniz bölgesi kalsa bile) elendiğinde tüm diplomasi ilişkileri, bekleyen `diplomatic_offers` kayıtları ve ticaret rotaları temizlenir; ayrıca save/load sonrası relation dışı veya elenmiş fraksiyona bağlı stale trade rotaları sanitize edilerek trade paneline geri sızmaları engellenir |
| Fetih sonrası bölge tahliyesi | ✅ | Bölge el değiştirince, o bölgede kalan düşman kara orduları artık en yakın kendi kara bölgesine geri çekilir; limandaki yabancı filolar otomatik olarak en yakın deniz bölgesine çıkarılır; ancak fetih bir fraksiyonun son kara toprağını da düşürürse kalan kara orduları ve donanmaları galip devlete devrolur ve yıkılış mesajı event log'a yazılır |
| Oyuncuya gelen diplomasi teklif paneli | ✅ | AI barış teklifleri `diplomatic_offers` kuyruğuna düşer; oyuncu modal anlaşma panelinden kabul/red verir, kabulde standart diplomasi motoru uygulanır. Teklif AI turu sırasında geldiyse sıra makinesi oyuncu cevabını bekler; kabul halinde teklif sahibi aktif AI o tur yeni saldırı/ileri hareket üretmeden kapanır; kabul/red edilen çözümler `DiplomaticOfferHistory` içinde saklanır, modalda kısa history kartlarıyla görünür ve history kartına tıklanınca ilgili fraksiyonun teklif sayfası açılır |
| Tur bitir UI cleanup ve faz input kilidi | ✅ | `ActionEndTurn` başarılı olduğunda renderer tüm oyuncu panellerini ve seçim state'ini kapatıp haritayı `Normal` moda döndürür; `PhaseAITurn` ve `PhaseTurnResolution` sırasında genel harita/HUD tıklamaları bloke edilir, yalnız teklif/tarihsel event modal seçenekleri etkileşim alır |
| Din diplomasisi | ✅ | Başlangıç ilişkileri din puanıyla kuruluyor; Sünni-Şii savaş başlıyor |
| Din dönüşümü | ✅ | Ele geçirilen bölgede 24 tur sonra yeni sahip dinine dönüşüm, memnuniyet -20 |
| Tarihsel olaylar | ✅ | JSON tetikleyici, tek seferlik olay işleme; tarihsel modal içinde A/B kararları, choice prompt, ekonomi/diplomasi/ordu etkisi ve ayrı karar log kaydı; follow-up zincirler flag, bölge sahipliği, teknoloji ve diplomasi stance/score koşullarına bağlanabiliyor; event popup ve event detail log görünümü follow-up/koşul özetini gösteriyor; event log `Kodex` düğmesi pending historical zincirleri `Hazır/Takvim/Kilitli` statüsüyle filtreli açıyor, solda kısa özetli ve scroll'lu liste sağda detay kartı gösteriyor, event başına kalan ay + kritik eksik koşulu sunuyor ve liste artık modal dışına taşmıyor; tarihsel popup artık draw/input/cursor tarafında gerçek üst modal olarak işlendiği için altta bekleyen teklif/onay diyalogları choice butonlarını kilitlemiyor |
| Zafer koşulları | ✅ | `domination`, `economic`, `military`, `religious`, `conquer_city`, `survive_turns` kontrol ediliyor; senaryo hedefleri `allowed_factions` ile oyuncu fraksiyonuna göre filtreleniyor, tam liste `ScenarioVictories` içinde tutuluyor; zafer kartları `deadline_year/month` ile oyuncuya süre veriyor ve AI hedef tamamladığı için otomatik kazanamıyor |
| AI turu | ✅ | Teknoloji, ekonomi, deniz, asker alma, konsolidasyon, diplomasi taraması, fırsatçı savaş ilanı ve hedefe hareket; AI bina/birim/transport üretimini anlık state mutasyonu yerine `ProductionQueue` üzerinden açar ve pending emirleri manpower/filo/region queue sınırlarında hesaba katar; kara recruit artık kışla throughput'u dolu bölgeye kör yazılmayıp serbest hatta dağıtılır, deniz üretimi de liman throughput'u dolu kıyıyı atlayıp mümkünse başka serbest limana yönelir; transport hattı olan savaşçı AI aynı tur escort warship de kuyruklayabilir ve escort üretimi artık tek deniz hattına kilitlenmek yerine çoklu cephe/baskı koşullarına göre birkaç fronta yayılabilir; bekleyen barış, ittifak ve ticaret teklifleri teknoloji farkı ve uzun vadeli tehdit baskısına göre önceliklenir ve prompt içinde tür ile kısa sebep bilgisi görünür; 1300'de Normal/Zor sınır komşusu zayıf hedeflere karşı güç/cephe üstünlüğü varsa proaktif savaş değerlendirebilir, Zor en fazla iki eşzamanlı savaşı taşıyabilir; `ai_expansion_targets` tanımlı tarihsel hedefler ise daha erken ve yüksek öncelikle değerlendirilir. Deniz hedefleri `aiSeaPressure()` ile savaş baskısına göre seçilir, filo limiti kıyı/savaş durumuna göre 1-3 arası dinamikleşir. AI fazı artık `TurnStepper` ile devlet devlet ve adım adım çözülür; oyuncuya yakın hamlelerde kamera odaklanır, uzak hamlelerde üst-orta AI overlay'i akışı korur ve tur sonunda kamera eski konuma döner |
| AI uzun menzilli hareket | ✅ | BFS ile uzaktaki hedefe doğru ilerleme |
| AI koalisyon | ✅ | Zorluk 3'te oyuncu 8+ bölgeyi geçince devreye girer |
| Kayıt/yükleme | ✅ | Autosave + QuickSave + slot1-3, metadata önizleme, silme; save dosyaları `kind` (`auto`/`quick`/`slot`), `game_version` ve düz `meta` wrapper'ı taşıyor. Gövde artık tam senaryo snapshot'ı değil, senaryo üstüne uygulanan campaign delta state'i saklıyor; relation'lar delta, region'lar delta, settlement'lar patch, army unit'leri stack formatında encode ediliyor ve payload `state_zstd` alanında `zstd+base64` sıkıştırmasıyla tutuluyor. `DEV_MODE=true` iken ek olarak aynı slot için okunabilir `*.debug.json` sidecar'ı yazılıyor; normal modda yazılmıyor ve eski sidecar temizleniyor. Yüklemede baz senaryo tekrar kuruluyor. Tur bitirde autosave, oyun içi kaydetmede quicksave; ana menüde Devam Et en yeni autosave/quicksave kaydını açıyor; save/load kartında silme onayı başlıkla çakışmayacak ayrı satır düzeni kullanır |
| Yükleme ekranı | ✅ | Senaryo ve kayıt yükleme sırasında gerçek zaman tabanlı hareketli spinner gösteriliyor; iş yükü ilk loading frame çizildikten sonra başlatıldığı ve yükleme adımları scheduler'a yield verdiği için loader animasyonu senaryo okunurken donmuyor; yükleme ekranı artık step-bazlı yüzde ve progress bar da gösteriyor; senaryo yüklemesi `faction_select`/`victory_select`e gidiyorsa ağır `WorldMap` cache'i loader bitiminde değil, harita ilk gerçekten gerektiğinde kuruluyor; zafer koşulu sonrası oyuncu turuna geçerken ve save load akışında `WorldMap` hazırlığı arka planda yapılıp loading ekranı altında tamamlanıyor |
| Ana menü / ayarlar | ✅ | Yeni oyun, en yeni autosave/quicksave ile devam et, kayıt yükleme, ayarlar, çıkış |
| Pause menüsü | ✅ | ESC ile açılır; devam, kaydet, yükle, ana menü, çıkış |
| Fare odaklı UI akışı | ✅ | Menü geri düğmeleri, teknoloji/diplomasi X kapatma, bölge/ordu panel kapatma, vergi/bina/asker aksiyonları fareyle yapılabilir |
| Olay paneli | ✅ | Sağ üst olay paneli daha fazla kayıt tutar, uzun liste mouse wheel ile kaydırılır |
| Minimap | ✅ | Sağ alt köşe, kamera ve ordu konumları |
| Üst-sol durum paneli | ✅ | Fraksiyon, kaynak, zafer ve ordu özeti haritanın üst-solunda ayrı panel; zafer/askeri özet kompakt iç kartlarla taşmadan çizilir, zafer HUD kartına tıklanınca merkezde detay popup açılır ve popup içinde sahip olunan/eksik hedefler checklist olarak gösterilir |
| Sağ-üst tarih/menü HUD | ✅ | Tarih, mevsim, tur ve duraklama menüsü butonu sağ üstte ayrı panel |
| Alt-orta aksiyon HUD | ✅ | Diplomasi, Teknoloji ve Tur Bitir butonları ayrı HUD içinde alt-ortada |
| Olay logu akordiyonu | ✅ | Panel daraltılıp genişletilir; uzun metinler wrap edilir; kartlar X ile kapanır, tıklanınca detay popup açılır |
| Info popup bildirimi | ✅ | Altın yetersiz gibi oyun içi uyarılar olay loguna yazılmaz, ayrı geçici popup olarak görünür |
| Event görünürlüğü | ✅ | Choice sonrası aktif event'ler harita üzerinde kara bölgesi ikonuyla görünür kalır; deniz bölgeleri event marker üretmez ve eski save'lerden gelen deniz event kayıtları ana harita, minimap ve hit-test tarafında gizlenir; aktif bölge event ikonlarına tıklanınca detay popup açılır ve bölge/tip/kalan tur/state izi okunur; detay popup artık başlık/kaynak/gövde ayrımıyla `[OLAY]`, `[KARAR]` ve harita izini birbirinden ayırır |
| Zafer detay popup scroll | ✅ | Hedef popup'ı artık sabit satır bloklarıyla taşmıyor; içerik gerçek satır yüksekliğine göre akıyor, mouse wheel ile kaydırılıyor ve uzun checklist/not için scrollbar çiziyor |
| Kompakt UI taşma düzeltmeleri | ✅ | Genel onay modalı mesaj wrap eder; bölge panelinde üretim alanı artık yalnız Tahıl ile sınırlı değildir, efektif `Altın/Tahıl/Demir/Kereste/Taş/Baharat/Kumaş` satırları iki kolon grid halinde çizilir; sahip bilgisi başlığın hemen altında etiketsiz görünür; sahip adı artık fraksiyon rengini korurken adaptif outline ile çizildiği için koyu devlet renklerinde zemine karışmaz; memnuniyet/vergi satırlarında yüzde metni, progress bar ve vergi `-/+` düğmeleri artık birbirine taşmaz; maksimum seviyedeki bina kartlarında alt satırdaki `Maks` yazısı kaldırılır ve uyarı durumu sol üst `Lv` rozetinin kırmızı arka planıyla verilir; kuyruktaki bina kartlarının `N Tur` göstergesi kontrastlı pill rozetine taşındığı için açık sprite üstünde kaybolmaz; recruit kuyruğunda aynı tur tamamlanacak emirler artık daha parlak, bekleyenler daha soluk kart stiliyle ayrılır; zafer seçim kartları genişletilip yükseltildiği ve açıklama/hedef satırları wrap edildiği için uzun tarihsel hedeflerde metin kart dışına taşmaz; teknoloji ağacı kartları da daha geniş/yüksek hale getirilip başlık ve effect özeti wrap edildiği için uzun teknoloji adları ile üç parçalı buff özetleri artık köşe rozeti veya kart sınırıyla çakışmaz |
| Teknoloji ikon görünürlüğü | ✅ | Teknoloji kartı kategori ikonları 20 px, üst filtre sekmesi ikonları 22 px çizilir; başlık ve sekme etiketleri ikonlarla çakışmayacak şekilde hizalanır |
| Panel cursor hit-test düzeltmesi | ✅ | Sol alt bölge paneli, olay logu, alt HUD, kayıt slotları ve onay panellerinde parmak imleci sadece gerçek tıklanabilir alanlarda gösterilir |
| Bölge paneli olay/komşu scroll düzeni | ✅ | Bina grid'i önce, Diplomasi ve diğer bölge aksiyonları ayrı bar içinde hemen altında çizilir; aktif olaylar ve komşular panelin altındaki local viewport'ta scrollbar ve mouse-wheel scroll ile taşmadan gösterilir |
| Ordu/yerleşim tıklama önceliği | ✅ | Aynı pikselde çakışan inputta seçim sırası artık `ordu/donanma etiketi > yerleşim > bölge`; hover hit-test de aynı helper ile eşlendi |
| UI framework migrasyonu (çekirdek + ekranlar) | ✅ | `internal/ui` altında `Widget`, `InputState`, `Manager`, ortak `Panel`, `Label`, `TextBox`, `Image/Icon`, `Tooltip`, `Button`, `Dropdown`, `ListView`, `Checkbox`, `RadioGroup`, `Modal`, `Overlay` ve layout yardımcıları eklendi; trade, diplomasi, teknoloji, pause/save-load, ana menü ve seçim ekranları, HUD küçük etkileşim yüzeyleri, recruit panel hit-test ailesi, ordu split/merge overlay'i, edit mode inspector/form yüzeyleri, shape yardım/pixel preview overlay'leri, diplomasi teklif diyaloğu ve confirm/war/event detail/historical modal aileleri ortak UI builder akışına taşındı; tema tokenları `internal/render/ui_theme.go` altında merkezileştirildi, ana menü/senaryo/fraksiyon/zafer/pause/kayıt slotlarında `Manager` tab focus kullanılmaya başlandı, seçim ekranı metinleri ortak `Label + TextRenderer` primitive'ine bağlandı, headless geometri + draw-call smoke ve allocation testleri eklendi; `ListView` artık press/release ayrımı ve drag threshold taşıyor, böylece diplomasi gibi uzun listelerde sürükleme scroll'u yanlış seçim üretmiyor; teknoloji panelinde kart görünümü ve hit-test akışı ayrıca `techCardComponent` seam'ine ayrıldı, çizim ve tıklama aynı rect/projection üzerinden yürütülüyor |
| Renderer sorumluluk ayrımı | ✅ | Monolitik `renderer.go`; yaşam döngüsü/draw orkestrasyonu, input, modal akışları, map editor ve ticaret overlay'i olarak `renderer_input.go`, `renderer_dialogs.go`, `map_editor.go` ve `trade_overlay.go` dosyalarına davranış değiştirmeden ayrıldı |
| Ortak ikon buton katmanı | ✅ | `internal/ui.Button` opsiyonel `IconID` desteği aldı; gerçek bitmap ikonlar `assets/ui/icons/*.png` altından yüklenip cache'leniyor, aynı primitive close/back/menu/kodex/müzik/diplomasi gönder yanında trade al/sat, savaş/diplomasi onayları, save/delete mini aksiyonları ve confirm dialog butonlarında tekrar kullanılıyor |
| Ses ve müzik | ✅ | `assets/sounds` global efektleri; senaryo `musics/` playlistleri `scenario.json` `music` alanından; ayarlarda ayrı müzik/ses seviyeleri; oyun içi müzik HUD'u ve ESC menüsü müzik kontrolleri |
| Development mode | ✅ | `DEV_MODE=true` ile `GameState.DevelopmentMode` |
| Render başlangıç log temizliği | ✅ | Boş senaryo path'inde shape dosyası okunmaz; deniz seed araması ham `world_x/world_y` fallback kullanır |
| Açılış kamera zoom ayarı | ✅ | İlk frame `resetCamera()` minimum sığdırma yerine `1.55x` yakın başlar; kampanya daha yakın açılır, sıkışık yerleşim kümeleri daha rahat ayrılır ve fare tekerleğiyle `5.5x` seviyesine kadar daha derin yakınlaşma yapılabilir. Yeni oyun veya save load sonrası oyuncu fraksiyonunun geçerli başkent settlement'ı varsa kamera bu noktaya odaklanır; kenara yakın başkentlerde ilk viewport clamp'lenir |
| Deniz anchor ve çakışma stabilizasyonu | ✅ | Deniz orduları gerçek su piksel anchor'ına çizilir; ordu/etiket çizim sırası deterministik, çakışan etiket metinleri bastırılır |
| Çoklu yerleşim noktaları | ✅ | `regions.json` içinde `settlements[]`; ana yerleşim ordu/etiket anchor'ı, yakın zoom'da ek yerleşim noktaları/isimleri, bölge dışı koordinatta log + nearest-region fallback; `port` settlement'lar liman simgesi, `fortress` settlement'lar kale simgesiyle ayrışır |
| Settlement edit mode | ✅ | `.env` `EDIT_MODE=true`; senaryo seçince harita editörü açılır, alt-sol bilgi/aksiyon HUD'u, settlement ekleme/silme, tip/capital değiştirme, bölge terrain/owner değiştirme, sürükleme, bölge arası taşıma, isim düzenleme, Shift+sürükle ile bölge merkezi taşıma ve Ctrl+S ile `regions.json` kaydı |
| Dropdown component | ✅ | `internal/ui/dropdown.go`; edit mode'da sahip/arazi/yerleşim tipi seçimlerinde yeniden kullanılabilir dropdown, scroll ve tam içerik desteği |
| Edit mode Voronoi debug overlay | ✅ | Edit mode'da `V` ile aç/kapatılır; seçili/hover bölgenin raster/Voronoi sınırını ve görsel komşularını JSON `neighbors` ile karşılaştırır, merkezler arası çizgiler ve hover koordinat paneli gösterir |
| Edit mode dirty exit uyarısı | ✅ | `editDirty` true iken ESC ile çıkışta ortak modal açılır; `Kaydet`, `Kaydetmeden Cik`, `Iptal` seçenekleriyle kayıp veri engellenir |
| Edit mode cleanup | ✅ | `Tip`, `Arazi`, `Sahip` butonları dropdown davranışına göre adlandırıldı; eski cycle helper'ları kaldırıldı |
| Edit mode undo/redo | ✅ | `Ctrl+Z` undo, `Ctrl+Y` veya `Ctrl+Shift+Z` redo; settlement ekle/sil/taşı/bölge arası taşı, region center, owner/terrain/type/capital/name değişiklikleri küçük snapshot command'leriyle geri alınır |
| Zaman kilitli bölge açılışı | ✅ | `is_locked=true` ve `unlock_turn>0` olan region aktif tur eşik değerine gelince otomatik açılır; unlock bildirimi gösterilir; load/save sonrası geçmiş unlock'lar senkronlanır |
| Edit mode bölge metadata editörü | ✅ | Inspector `Harita` sekmesinde region `name_tr`, `name`, `is_locked`, `unlock_turn` ve görsel Voronoi komşularından iki yönlü `neighbors` sync düzenlenir; deniz region seçiminde inspector `Deniz Bolgesi`, yerleşim olmadığını ve pasif `Denizde Yok` buton etiketini açıkça gösterir; settlement odaklı pasif butonlar da bağlama göre `Tip Yok` / `Isim Yok` / `Silinmez` ya da `Tip Sec` / `Isim Sec` / `Sil Sec` etiketine döner; kara/deniz odak noktası renkleri edit modda ayrıdır |
| Edit mode bölge ekleme/silme | ✅ | `Ctrl+Alt+sol` veya `Bolge Ekle` mevcut shape içinde yeni Voronoi seed region oluşturur; kara ve deniz region'ları seçilip merkezleri taşınabilir, çoğaltılabilir ve silinebilir; `Bolge Sil` seçili region'ı, komşu referanslarını ve o region'daki başlangıç ordularını kaldırır; undo/redo destekli |
| Edit mode geniş veri editörü | ✅ | Inspector `Veri` sekmesinde faction ekleme/düzenleme formu, faction silme, başlangıç kaynakları/playable/AI değeri, başlangıç diplomasi `stance/score`, başlangıç kara ordusu/donanma ekleme-silme ve seçili ordu/donanma birim sayıları düzenlenir; `Birim Tipi` dropdown'ı veri sekmesinde görünür; harita üstünde tüm ordu/donanma sayıları edit mode'da gizlenmeden görünür ve açık fraksiyon renklerinde kontrastlı metinle okunur; limanda demirli filolar liman anchor'ında, denize açılanlar deniz bölgesi anchor'ında çizilir; form `Kaydet` ve Ctrl+S `regions.json`, `factions.json`, `relations.json`, `armies.json` yazar |
| Edit mode shape paint editor | ✅ | Inspector `Shape` sekmesi seçili kara region'ın `shape_id` verisini sağ mouse drag ile boya/sil düzenler; stroke sırasında yeşil/kırmızı canlı preview overlay ve yardım paneli görünür; stroke bitince mask contour'ları yeniden ring'e çevrilir, `ShapeData` + `Region.Shape` güncellenir, undo/redo world snapshot'a shape verisini de alır; `Kaydet` artık `country_shapes.json` da yazar. Aynı sekmedeki `Bolge Boya/Sil` aracı kara veya deniz region'larında `region_shapes.json` override katmanına kalıcı yazar; ülke dış sınırının dışına taşan boyamalar ile deniz alanı dağılımı restart sonrası korunur ve sonraki stroke'lar eski override piksellerini yanlışlıkla düşürmez. Region tool canlı preview'i stroke başlangıcına göre ayrı overlay çizer ve shape session'ı lazy tuttuğu için yoğun boyama sonrası merkez taşıma/geçişler daha akıcıdır |
| Ticaret yolu görsel sadeleştirme | ✅ | Harita üstü ticaret çizimi `A->B` ve `B->A` rotalarını tek koridorda birleştirir; `camScale < 0.85` iken yalnızca oyuncuya bağlı hatlar çizilir, etiketler yalnızca yakın zoom'da görünür |
| Harita modu (Normal/Ticaret) | ✅ | EU4 benzeri harita modu anahtarı eklendi; ticaret koridorları yalnızca `Ticaret` modunda çiziliyor, normal haritada çizgi karmaşası yok |
| Senaryo bazlı tarihsel ticaret merkezleri | ✅ | Trade map merkezleri senaryo `data/trade_centers.json` içindeki `tier` + `links` graph yapısından okunuyor; koridor akışı merkezler arasında doğrudan değil, link graph kısa yolu üzerinden dağıtılıyor; `off_map=true` ile sadece etiket ve bağlantı gösteren dış hat düğümleri (`name_tr`, `world_x`, `world_y`) de JSON’dan tanımlanabiliyor; `unlock_year` alanı sayesinde geç dönem Atlantik/Amerika hatları belirli yıldan önce tamamen gizli/pasif tutulabiliyor |
| Ticaret paneli ayrıştırması | ✅ | `Yeni Rota` sekmesi artık gerçek ticaret anlaşması adaylarını ve engel nedenlerini gösterir; manuel al/sat akışı ayrı `Pazar` sekmesine taşındı, müttefik devletlerle ticaret rota bazında bağımsız açılabiliyor; pazar sekmesinde fraksiyon/mal listeleri click anında satır resolve ettiği ve panel tam mouse state aldığı için seçimler tekrar güvenilir çalışıyor |
| Başkent sistemi | ✅ | Ulusal başkent artık fraksiyon üstünde `capital_settlement_id` ile tutulur; başkent bölgesi ek üretim/lojistik bonusu alır; başkent fethedilince hazine-hammadde stoğunun yarısı ve eksik teknolojilerin yaklaşık yarısı fethedene geçer; savunan için yeni başkent en yüksek getirili bölgenin merkez settlement'ına atanır; settlement panelinden 5 turluk taşıma kuyruğu başlatılabilir; tüm başkent settlement'ları haritada yıldız rozetiyle görünür |

| WSL / Windows build hattı | ✅ | `wiki/dev/build-setup.md` içinde Ebitengine için Ubuntu paketleri, `go test ./...` doğrulaması ve `GOOS=windows GOARCH=amd64 go build -o bin/game.exe ./cmd/game` akışı belgelendi |

## Bilinen Sorunlar

| Öncelik | Sorun | Dosya | Etki |
|---|---|---|---|
| 🟢 Düşük | Kök dizinde geçici `game.exe` olabilir | `game.exe` | Kalıcı çıktı `bin/game.exe` olmalı |

## Sonraki Adım Planı

1. **Diplomasi history ayrıntı paneli:** Filtrelenen kayıtlar için hover açıklaması ve ekstra debug etiketi düşün.

## Yakın Sprint Önerisi

İlk sprintin hedefi "seçilen kampanya hedefi güvenilir çalışıyor ve kayıt yükleme bozmuyor" olmalı:

| Sıra | İş | Kabul Kriteri |
|---|---|---|
| 1 | Diplomasi teklif geçmişi | Kabul/red edilen teklifler kısa bir history panelinde görülebiliyor |

## Araçlar

| Araç | Amaç |
|---|---|
| `tools/centroids/main.go` | Bölge merkez koordinatları hesapla |
| `tools/populate_all_shapes.py` | Natural Earth'ten poligon üret |
| `tools/update_shapes_from_ne.py` | Şekilleri güncelle |
| `tools/fix_*.py` | Belirli bölge düzeltmeleri |
| `tools/add_regions*.py` | Yeni bölge ekleme |
| `tools/add_missing_countries.js` | Eksik ülke tamamlama |
| `tools/audit_map.py` | Harita/veri denetimi |
