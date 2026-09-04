---
type: dev
tags: [progress, status, todo, known-issues, next-steps]
last_updated: 2026-09-04
related: [HOME, architecture/game-loop, architecture/state-management, architecture/render-pipeline, systems/victory]
---

# Geliştirme Durumu

- 2026-09-04: Hareket sonrası aynı açık denizde kalan düşman filoları için,
  oyuncunun kalan hareket puanıyla belirli filo ikonuna sağ tıklayarak ayrı
  deniz teması başlatabilmesi sağlandı. Regression: `TestPlayerCanContactAnotherFleetAfterMovementContact`;
  doğrulama: `go test ./... -count=1`.

- 2026-09-03: Vassal haraç oranı oyuncunun `Vassal Yönetimi` kartından `%0–50`
  arasında `%5` adımlarla ayarlanabilir hale getirildi. Haraç oranı ekonomi
  aktarımına ve vassal bölge memnuniyetine bağlandı; oran ve yapılandırma durumu
  save/load ile korunuyor. Regression: `TestApplyEconomyTickUsesConfiguredVassalTributeRate`,
  `TestCalculateIncludesVassalTributePressure`, `TestHandleDiplomacyInputAdjustsVassalTribute`.

- 2026-08-29: Aktif kuşatma bölgesinde bulunan, kuşatanla savaşta olmayan ve
  kuşatanın müttefiki/aynı realm'ı olmayan üçüncü devlet orduları, son geldikleri
  geçerli komşu bölgeye otomatik olarak geri çıkarılıyor. Orduların son hareket
  kaynağı `PreviousRegionID` ile save uyumlu biçimde tutuluyor. Regression:
  `TestResolveSiegesEvacuatesDefenderAlliesFromBesiegedRegion`; kapsam:
  `internal/army/army.go`, `internal/game/{game.go,land_contact.go,siege.go}`.

- 2026-08-28: 1300 senaryosuna Artuklu, Eretna, Kadı Burhaneddin, Karakoyunlu,
  Akkoyunlu, Celayirli, Muzafferî, Şirvanşah, Timur ve Afşar ardıl devletleri
  eklendi. Devletler başlangıçta elenmiş, gerçek senaryo bölgeleri
  `successor_faction_id` ile eşlenmiş ve tarih aralıkları faction metadata'sına
  işlendi; mevcut isyan, fetih sonrası karar ve özgürleştirme akışlarıyla
  oyunda yeniden kurulabilir. Regression:
  `Test1300ScenarioHistoricalSuccessorStatesAreMappedAndDated`.

- 2026-08-28: Kuşatan devletin hedef bölgeyle ortak kara sınırı olduğunda düzenli
  ikmal, savunma ordusu bulunmayan kuşatmalardaki doğrudan yıpranmayı `3`ten
  `2 HP/birim/tur`a indiriyor. Regression: `TestResolveSiegesReducesBesiegerAttritionWithOwnedBorderSupply`;
  kapsam: `internal/game/siege.go`.

- 2026-08-28: Oyuncunun kuşatması altındaki AI ordusu huruç yaptığında savaşın
  otomatik çözülmesi kaldırıldı. AI turu `TurnStepSortie` ile bekletiliyor ve
  oyuncuya `Çatış` veya `Kuşatmayı Kaldır` seçenekleri sunuluyor; AI-AI huruçları
  otomatik çözülmeye devam ediyor. Regression: `TestExecuteMoveDefersAISortieAgainstPlayerSiege`,
  `TestShowSortieDecisionOffersFightOrLiftSiege`.

- 2026-08-28: Sanal Rebel devletleri save/load sözleşmesine dahil edildi. Compact
  ve debug kayıtları `IsVirtual`, Rebel ordularının `IsRebel` ve
  `RebelAgainstID` alanlarını taşır. Eski kayıtlarda senaryo fraksiyon listesinde
  bulunmayan `rebel_<region>` sahipleri yükleme sırasında otomatik sanal fraksiyona
  dönüştürülür; Rebel ordularının eski sahibi de savaş ilişkisinden yeniden
  çıkarılır. Böylece isyan bölgesine ordu sevki ve savaş ilanı kayıt yükleme
  sonrasında da çalışır. Regression: `TestCompactSaveRestoresVirtualRebelFactionAndArmy`,
  `TestOldSaveMigratesMissingVirtualRebelFaction`,
  `TestDeclareWarAgainstVirtualRebelFaction`; kapsam:
  `internal/save/compact.go`, `internal/{save,diplomacy}/*_test.go`.

- 2026-08-26: Arazi alanlarına (çöl/dağ/göl/nehir/sık orman/bataklık) hareket
  maliyetine ek olarak yıpranma (%HP kaybı) değeri eklendi. Alan seçiliyken
  Harita sekmesindeki "Yıpranma" düğmesiyle 0→20 arası 5'lik adımlarla
  ayarlanır; oyuncu ve AI orduları bu alana her girişte `army.ApplyAttritionPercent`
  ile HP kaybeder. Arazi alanları hâlâ normal bölge gibi seçilebilir/komşulanabilir
  ama vergi, yerleşim, nüfus ve üretim sistemlerinden tamamen muaf kalır.
  Kapsam: `internal/world/terrain_area.go`, `internal/army/army.go`,
  `internal/state/movement.go`, `internal/game/game.go`, `internal/ai/ai.go`,
  `internal/render/shape_editor.go`.

- 2026-08-25: Arazi alanı boyama, arazi tipi, alan maliyeti ve silme
  kontrolleri Edit Mode **Harita** sekmesinde toplandı. Normal bölge tipi
  **Bölge Tipi** olarak Bölge sekmesinde kaldı; arazi alanları yalnız kendi
  `Arazi Tipi` listesindeki dağ, çöl, göl, nehir, sık orman ve bataklık
  seçeneklerini kullanıyor. Kapsam: `internal/render/map_editor.go`,
  `internal/render/shape_editor.go`.

- 2026-08-25: Arazi alanı boyama, maliyet ve silme kontrolleri Edit Mode
  **Bölge** sekmesine taşındı; `Komşu Ekle` de aynı sekmede yer alıyor.
  Normal bölgelerde dropdown etiketi `Bölge Tipi`, terrain child düğümlerinde
  `Arazi Tipi` olarak ayrıştırıldı. Kapsam: `internal/render/map_editor.go`,
  `internal/render/shape_editor.go`.

- 2026-08-25: Edit Mode arazi alanlarına seçili alanı silme ve bağımsız arazi
  tipi değiştirme kontrolleri eklendi. Alanlar parent devlet renginin tonlarıyla
  çiziliyor; runtime merkezi boyalı hücrelerin ortasında hesaplanıyor ve merkez
  sürükleme kapalı. Kapsam: `internal/render/shape_editor.go`,
  `internal/render/terrain_areas.go`, `internal/world/terrain_area.go`.

- 2026-08-25: Arazi alanlarının merkezleri sabitlendi ve Voronoi merkez
  adaylarından çıkarıldı. Alt alanlara bağımsız `dağ`, `çöl`, `göl`, `nehir`,
  `sık orman` ve `bataklık` arazi tipleri verilebiliyor; editör dropdown'u ve
  harita renkleri güncellendi. Kapsam: `internal/world/terrain.go`,
  `internal/world/terrain_area.go`, `internal/render/mapgen.go`.

- 2026-08-24: Arazi alanları artık oyun haritasında gerçek runtime alt-bölge
  düğümleri olarak seçilebilir. Parent bölge ve temas eden kardeş alanlarla
  komşuluk kurar; ordu sağ tıkla geçilebilir alanlara yönlendirilebilir ve
  normal seçim/marker akışı kullanılır. Arazi alanı paneli vergi, nüfus,
  üretim ve yerleşim işlemlerini içermez. Kapsam: `internal/world/terrain_area.go`,
  `internal/render/mapgen.go`, `internal/render/panel.go`, `internal/state/movement.go`.

- 2026-08-24: Yerleşimsiz dağ/çöl/sık orman gibi alt arazi alanları eklendi.
  Edit modda seçili kara bölgesinin mevcut ortak fırçasıyla çizilir; `0`
  geçişi engeller, `-1` ve `-2` ek hareket puanı tüketir. Alanlar
  `terrain_areas.json` içinde parent bölgeye bağlı hücreler olarak saklanır,
  oyun ve AI hareket kurallarına uygulanır. Kapsam: `internal/world/terrain_area.go`,
  `internal/state/movement.go`, `internal/render/terrain_areas.go`.

- 2026-08-24: Bölgesel vergi oranı oyuncu ve AI için en fazla `%60` olacak şekilde
  sınırlandı. Senaryo ve kayıt yükleme sırasında eski `%60` üzeri oranlar da
  normalize ediliyor. Regression: `TestAIAdjustTaxesProtectsUnrestAndRaisesHealthyRevenue`;
  kapsam: `internal/world/region.go`, `internal/world/loader.go`,
  `internal/game/game.go`, `internal/ai/tax_policy.go`, `internal/save/compact.go`.

- 2026-08-24: Üst HUD askerî kapasite kartı 1024 px genişliğindeki ekranlarda
  sağa taşmaması için 14 px sola alındı; `Savaşçı / Ordu / Donanma` değerleri
  kart ve görünür ekran sınırları içinde kalıyor. Kapsam: `internal/render/panel.go`.

- 2026-08-24: Kuşatma yıpranması artık yalnızca kuşatma başlangıcında kaydedilen
  savunma ordusu ID'sine bağlı değil. Her kuşatma turunda bölgedeki güncel,
  savaş halindeki savunma ordusu yeniden bulunup HP kaybı alıyor; savunma ordusu
  yoksa kuşatan ordu da yalnız düşük seviyeli, zamanla biriken `3 HP/birim/tur`
  kuşatma yıpranması alıyor.
  Regression: `TestResolveSiegesFindsDefenderThatWasNotPresentAtSiegeStart`,
  `TestResolveSiegesAppliesLowAttritionWhenRegionHasNoDefenderArmy`;
  kapsam: `internal/game/siege.go`.

- 2026-08-28: Kuşatma teslim süresi sur başına 2, ambar başına 1 tur olacak şekilde
  sadeleştirildi. Büyük gedikte teklif yarı süreyi beklemeden kabul edilir; diğer
  teklifler toplam sürenin yarısından sonra yüzde 50 zarla kabul edilebilir ve toplam
  süre dolduğunda zorunlu teslimiyet uygulanır. Eski sabit 3 tur AI kabul eşiği
  kaldırıldı. Kapsam: `internal/state/state.go`,
  `internal/game/{game.go,siege.go}`.

- 2026-08-23: Teknoloji tooltip'inde önkoşul olarak gösterilen ham teknoloji ID'leri
  yerelleştirilmiş `NameTR` etiketleriyle değiştirildi. Eksik teknoloji tanımları
  için ID geri dönüşü korunuyor. Regression: `TestTechRequirementLabelsUseTechnologyNames`;
  kapsam: `internal/render/tech_panel.go`, `tech_effect_summary.go`.

- 2026-08-23: AI ilişki bakım spam'i sınırlandı. Her fraksiyon tur başına en fazla
  bir heyet/hediye aksiyonu yapıyor; başarılı veya oyuncuya kuyruğa alınan bakım
  aksiyonu ilgili ilişkiye dört tur cooldown uyguluyor. Oyuncuya bekleyen ilişki
  bakım bildirimleri en fazla iki kayıtla sınırlanıyor. Regression:
  `TestAISendsDelegationButNeverGiftToStrategicTargetUnderDirectThreat`,
  `TestQueueOfferWithMetaLimitsPendingRelationshipNotifications`;
  kapsam: `internal/ai/diplomacy.go`, `internal/diplomacy/offers.go`,
  `internal/faction/faction.go`, `internal/save/compact.go`.

- 2026-08-23: Düşük memnuniyetli bölgelerde isyan artık sahipsiz bir durumdan
  ibaret değil. Nüfus, yerleşim/bina gelişmişliği ve tahıl ikmal seviyesine göre
  1–20 milislik `IsRebel` ordusu doğuyor; eski sahibin ordusu bölgeye gelirse
  isyan bastırılıyor. Bölgenin `SuccessorFactionID` değeri elenmiş geçerli bir
  fraksiyona işaret ediyorsa isyanın sonraki turdaki zaferi ardıl devleti kurup
  isyancı orduyu devlete bağlıyor. Regression: `TestCheckRebellionsSpawnsScaledRebelArmy`,
  `TestCheckRebellionsSuccessorFormsAfterRebellionWins`,
  `TestCheckRebellionsArmySuppressesExistingRevolt`.

- 2026-08-23: Teknoloji kartlarının üzerine gelindiğinde bilgi popup'ı açılıyor;
  oyuncu teknoloji etkisini, maliyetini, araştırma süresini ve tamamlanmamış
  önkoşulları tek yerde görebiliyor. Regression: `TestTechRequirementSummaryShowsMissingRegionsAndYear`,
  `TestTechCardComponentCarriesRequirementText`.

- 2026-08-23: Teknolojilere `required_regions` ve `min_year` önkoşulları eklendi.
  `cast_bronze_cannon`, Edirne/Trakya (`thrace`) ve Bursa'nın elde tutulmasını ve
  1420 yılına ulaşılmasını bekliyor; oyuncu, AI ve otomatik araştırma aynı uygunluk
  kontrolünü kullanıyor. Regression: `TestIsUnlockedForContextRequiresYearAndRegions`,
  `TestNextResearchableTechIDForContextSkipsUnavailableTech`.

- 2026-08-23: AI devletinin başkenti fethedildiğinde yeni başkent artık en yüksek
  gelirli değil, tamamlanmış bina seviyeleri başta olmak üzere yerleşim altyapısı
  ve nüfusu en gelişmiş kalan bölgeden seçiliyor. Eşit skorlar bölge ID'si ile
  deterministik çözülüyor. Regression: `TestBestCapitalSettlementPrefersMostDevelopedRegionOverHigherIncome`.

- 2026-08-23: Agresif veya savaş halindeki 1300 AI, nüfus tabanlı kara rezervine
  ulaştığında asker üretimini kesmiyor. Genişleme kuvveti hedefi agresifliğe göre
  artırılıyor; mevcut `ManpowerCap`, bütçe, kaynak, lojistik ve güvenlik sınırları
  korunuyor. Regression: `TestAIAggressiveExpansionKeepsBuildingArmyAfterReserveFloor`.

- 2026-08-23: Barış sonrası beş turluk ateşkes artık devletin barış yaptığı
  düşmana karşı müttefik veya imparatorluk savaş çağrısıyla yeniden savaşa
  girmesini de engelliyor. Doğrudan savaş ilanı, çağrı önizlemesi ve bekleyen
  savaş çağrısı çözümlemesi aynı kuralı kullanıyor. Regression:
  `TestWarCallBlockedByRecentPeaceWithEnemy`.

- 2026-08-23: Diplomasi teklif modalının sağ özet kartında uzun teklif eden
  devlet adları iki satıra kadar sarılıyor; durum ve ilişki satırları ek satır
  kadar aşağı taşınarak kart dışına taşma önlendi (`internal/render/renderer_dialogs.go`).

- 2026-08-23: Ordu bilgi paneline seçili birimleri `SİL` düğmesiyle terhis etme
  eklendi. Evet/Hayır modalı sonrasında birimler çıkarılıyor ve normal üretim
  maliyetlerinin kaynak bazında `%20`si iade ediliyor; boş ordular ile lojistik
  kayıtları temizleniyor. Regression: `TestDisbandArmyRefundsTwentyPercentAndKeepsUnselectedUnits`,
  `TestDisbandArmyRemovesEmptyArmyAndLogistics`, `TestDisbandButtonIsLeftOfArmyActions`.

- 2026-08-23: Tek saha ordusunun kuşatıcıyı yenmeye yetmediği durumda AI artık
  ulaşılabilir yakın orduları ortak `relief` rally noktasında topluyor. Toplam güç
  kuşatıcı gücünün `%110`una ulaştığında ordular mevcut lojistik/kapasite kurallarıyla
  birleşip sonraki tur kuşatmayı kaldırmaya çalışıyor. Regression:
  `TestFriendlyReliefRalliesMultipleArmiesWhenOneIsInsufficient`; doğrulama:
  `go test ./internal/ai`.

- 2026-08-23: AI'nin kuşatma altındaki başkente yakın ordusuyla müdahale etmemesi
  düzeltildi. `nearestReliefArmy()` artık yalnız ulaşılabilir ve kuşatıcıdan en az
  `%110` güçlü orduları değerlendiriyor; kritik savunmada en güçlü/tek saha ordusu
  da relief rolüne atanabiliyor. Üstün kuşatıcıya karşı zayıf ordu pasif savunmada
  bırakılıyor. Regression: `TestFriendlyReliefUsesStrongestArmyWhenItIsTheOnlyViableForce`,
  `TestFriendlyReliefDoesNotCommitToSuperiorBesieger`; doğrulama: `go test ./internal/ai`.

- 2026-08-23: Gelir ve tahıl hesabı popup'larındaki header/footer ayırıcılarının
  bitiş koordinatı popup genişliği yerine panelin mutlak X konumundan türetildi.
  Böylece çizgiler panel dışına taşmıyor (`internal/render/income_popup.go`).

- 2026-08-23: Ordu oluşturma paneli, alt ana aksiyon HUD'ının hemen üzerine
  `3 px` boşlukla hizalandı. Çizim ve hit-test aynı ortak Y geometrisini kullanıyor.
  Regression: `TestRecruitPanelSitsThreePixelsAboveBottomActionHUD`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-08-23: Üst HUD Tahıl hover alanı, çizilen Tahıl satırıyla hizalandı;
  önceki hit-test bir alt satırdaki Baharat alanına taşıyordu. Regression:
  `TestGrainHUDValueRectIsInteractive`; doğrulama:
  `go test ./internal/render -run 'Test(IncomeHUDValueRectIsInteractive|GrainHUDValueRectIsInteractive|GrainEconomyPopupUsesCurrentSnapshot|PlayerGoldEconomyStatusUsesNetChange)$' -count=1`.

- 2026-08-23: Savaş halindeki devletlere `Heyet` gönderme ve `Vassallık` teklifi
  diplomasi panelinde pasifleşmiyor. Heyet savaş duruşunu koruyarak ilişkiyi `+8`
  artırıyor; vassallık teklifinde savaş durumu gönderim kapısı değil, mevcut kabul
  değerlendirmesi korunuyor. Regression: `TestWartimeDelegationAndVassalizationRemainAvailable`;
  doğrulama: `go test ./internal/diplomacy -count=1`.

- 2026-08-21: Ordu/donanma marker'larının sağ-üst bonus rozetlerinde iki/üç
  karakterli değerlerin taşmasını önlemek için geniş etiketler 8 px `FaceMicro`
  fontuna otomatik geçirildi; rozet anchor ve hit-test geometrisi değişmedi.
  Regression: `TestMarkerBadgeCompactFaceFitsTwoDigitBonusValue`.

- 2026-08-21: Donanma sınırsız üretimden çıkarıldı. `GameState.NavalCap()` kara
  bölgesi, liman seviyesi ve nüfusa göre toplam gemi kapasitesi hesaplıyor;
  savaş, nakliye ve ticaret gemileri aktif/kuyrukta aynı kapasiteye dahil ediliyor.
  Oyuncu üretimi miktarı bu sınıra göre kırpıyor, AI deniz görevleri kapasite
  doluyken kaynak harcamadan duruyor. Regression: `TestNavalCapScalesWithRegionsPopulationAndPortLevels`,
  `TestNavalUnitsIncludingQueueCountsAllShipCategories`,
  `TestRecruitSpecificScalesNavalQueueToFactionNavalCap`.

- 2026-08-21: Donanma kapasitesi UI'a taşındı. Üst HUD askerî kapasite kartı
  `Savaşçı / Ordu / Donanma` satırlarını gösteriyor; liman tooltip'i sonraki
  seviyenin `+2 gemi` etkisini ve `mevcut → sonraki` toplam sınırı açıklıyor.
  Regression: `TestPortBuildingEffectLinesShowNavalCapacityIncrease`.

- 2026-08-21: Kara askerî kapasitesi de donanma ile aynı görünürlük sözleşmesine
  taşındı. Savaşçı sınırı temel `ceil(kara bölgesi/2) × 20` kuralına bağlı;
  ordu sınırındaki +1 slot savaşçı kapasitesini artırmıyor. Kışla tooltip'i
  bu ayrımı ve üretim hattı limitini,
  recruit paneli `Ordu`, aktif + kuyruktaki `Savaşçı` ve kıyıda `Donanma`
  değerlerini `mevcut/sınır` formatında gösteriyor; fraksiyon detayındaki
  askeri durum satırları da aynı formata geçirildi. Regression:
  `TestBarracksBuildingEffectLinesShowLandCapacityIncrease`,
  `TestManpowerCapTracksMaxLandArmies`.

- 2026-08-07: AI komutan ataması askerî birim varlığına bağlandı. Boş kara
  orduları ile yalnız ticaret/nakliye gemisi taşıyan filolar komutan almıyor;
  savaş gemisi içeren filolar komutan almaya devam ediyor. Askerî olmayan filoda
  eski komutan varsa havuza bırakılıyor ve yeni fallback üretmeden önce boşta
  komutan yeniden kullanılıyor. Regression: `TestEnsureFactionCommandersSkipsNonMilitaryFleetsAndReusesAvailableCommander`,
  `TestEnsureFactionCommandersAssignsWarshipButNotMerchantFleet`; doğrulama:
  `go test ./... -count=1`.

- 2026-08-07: Üst HUD Tahıl değeri altın gelir hesabına paralel hover popup'ına
  bağlandı. Popup üretim, halk tüketimi, ordu tüketimi, toplam tüketim, gerçek
  pazar satış arzı, otomatik ihracat ve net değişimi gösteriyor; Tahıl değeri ve
  popup hit-test'i aynı rect'i kullanıyor ve hover sırasında pointer cursor alıyor.
  Regression: `TestGrainHUDValueRectIsInteractive`,
  `TestGrainEconomyPopupUsesCurrentSnapshot`; kapsam:
  `internal/render/{income_popup.go,cursor.go,renderer.go}`.

- 2026-08-07: AI'nin açık genişleme hedeflerine, aktif `expand` planı
  hedeflerine ve claim bölgelerinin güncel sahiplerine `Heyet`/`Hediye`
  göndermesi sınırlandı. Normal durumda ikisi de engelleniyor; AI doğrudan
  tehdit altında ve müşkül durumdaysa yalnız `Heyet` gönderilebiliyor,
  `Hediye` hiçbir zaman gönderilmiyor. Stratejik hedef olmayan doğrudan
  tehditlerde savunma amaçlı yatıştırma davranışı korunuyor. Regression:
  `TestAIDoesNotRepairRelationsWithExpansionTarget`,
  `TestAIRejectsRelationRepairForCurrentClaimOwner`,
  `TestAIRejectsRelationRepairForActiveExpansionPlanTarget`,
  `TestAISendsDelegationButNeverGiftToStrategicTargetUnderDirectThreat`;
  kapsam:
  `internal/ai/{diplomacy.go,strategic_plan.go,ai_test.go}`.

- 2026-08-07: Diplomasi teklif panelindeki `Aktif İlişkiler` listeleri artık her
  kategorideki tüm devletleri gösteriyor. Üç satırlık sabit kesme kaldırıldı;
  kategori içerikleri ortak clipped viewport içinde çiziliyor, ihtiyaç halinde
  scrollbar ve mouse-wheel ile bağımsız ilişki scroll'u kullanılıyor. Regression:
  `TestDiplomacyRelationsPanelShowsAllEntriesWithIndependentScroll`; kapsam:
  `internal/render/{diplom.go,renderer.go,diplom_test.go}`.

- 2026-08-07: Oyuncuya gelen `Heyet` ve `Hediye` bildirim modalı 3 saniye açık
  kalıyor. Süre dolduğunda mevcut `Tamam` çözümü otomatik üretilerek bildirim
  kabul edilmiş gibi akış devam ediyor; diğer diplomasi teklifleri seçim
  beklemeye devam ediyor. Regression:
  `TestDiplomacyRelationshipNotificationAutoClosesAfterThreeSeconds`.

- 2026-09-04: Devlet başına doğrudan dış ittifak sayısı `MaxAlliances = 5` ile sınırlandı. Aynı vassal realm içindeki iç `StanceAllied` kayıtları kotaya dahil edilmiyor; doğrudan teklif ve kuyruk çözümü için regresyon testleri eklendi (`TestProposeAllianceRejectedWhenActorHasFiveAllies`, `TestQueuedAllianceOfferRejectedWhenTargetReachesFiveAllies`).

- 2026-08-07: İttifak-savaş çakışması iki yönlü ve oyuncu/AI ortak kural haline getirildi. Bir devlet, diğer tarafın doğrudan müttefikiyle savaş halindeyse ittifak kuramıyor. Bir devlete savaş ilan edildiğinde hedefin tarafsız kalan doğrudan müttefikleriyle ilişki `-25` düşüyor; saldıranın mevcut müttefiki olan hedef müttefikiyle ittifak bozulup ilişki `-35` düşüyor. Savaşa katılan müttefiklerde ek ceza uygulanmıyor. Regression: `TestProposeAllianceRejectedWhenTargetAllyIsAtWarWithPlayer`, `TestDeclareWarPenalizesTargetAlliesOnce`, `TestDeclareWarBreaksCrossAllianceWithTargetAlly`, `TestDeclareWarAgainstTargetAllyThatJoinsUsesWarRelationOnly`.

- 2026-08-07: Askerî kara/deniz birimlerine JSON tabanlı sabit `gold_upkeep`
  eklendi. Ekonomi tick'i her tur hazine gelirinden toplam ordu maaşını düşüyor;
  yetersiz hazine `GoldEconomyStatus.Shortage` ile birlikte HP yıpranması,
  moral kaybı ve tam açığa bağlı deterministik asker kaçağı üretiyor.
  Recruit tooltip/HUD bakım bilgisini gösteriyor. AI mevcut ordunun üç turluk
  maaş rezervini koruyor ve yeni birim seçiminde gelecekteki maaşı puanlıyor.
  Regression: `TestTotalGoldUpkeepIncludesLandAndEmbarkedUnits`,
  `TestApplyEconomyTickDeductsFixedGoldArmyUpkeep`,
  `TestApplyEconomyTickClampsUnpaidGoldUpkeepAtZero`,
  `TestGoldUpkeepShortageCausesAttritionAndDesertion`.

- 2026-08-07: Üst HUD Gelir değeri bakım düşülmüş net tur değişimine bağlandı.
  Vergi, pasif/rota ticareti, teknoloji, haraç, ganimet, ordu masrafı ve tur içi
  hediye transferleri `GoldEconomy` snapshot'ında ayrıştırılıyor. Gelir rakamı
  ortak hit-test rect'iyle pointer cursor alıyor ve hover popup'ında hesap
  kalemlerini gösteriyor.

- 2026-08-07: Seçili ordu/donanma panelinde tahıl ihtiyacının yanında toplam
  `Asker Maaşı: N/tur` gösterildi. Kara birlikleri, gemiler ve taşınan birlikler
  kanonik altın bakım hesabına dahil ediliyor. Regression:
  `TestArmyUpkeepSummaryShowsGrainAndGoldSalary`.

- 2026-08-07: `Test1300ScenarioResourceSpecializationsAndProductionCosts` sabit
  bölge, bina, birlik ve maliyet beklentilerinden çıkarıldı. Test artık ilgili
  JSON kayıtlarını keşfediyor; yalnız JSON'da mevcut olan kaynak, üretim maliyeti
  ve mod alanlarını loader çıktısıyla karşılaştırıyor. Sabit kaynak sayısı
  eşikleri ve belirli ID listeleri kaldırıldı.

- 2026-08-07: Kara ve deniz temasında `Pozisyonu Koru` artık tek başına
  çatışmayı iptal etmiyor. Taraflardan biri `Çatış` seçip diğeri `Geri Çekil`
  seçmediyse muharebe çözülüyor; iki taraf da `Pozisyonu Koru` seçerse temas
  savaşsız kapanıyor. Koru seçen tarafın temas savunma bonusu savaş planı
  önizlemesine ve gerçek resolve hesabına taşındı. Regression:
  `TestPlayerLandContactHoldStillBattlesWhenEnemyClashes`,
  `TestNavalContactBattlesWhenEitherSideClashes`,
  `TestBlockadeAndPlayerHoldCanShareSeaWithoutBattle`.

- 2026-08-07: AI'nin `Heyet`/`Hediye` harcamaları hazinenin son önceliğine taşındı.
  AI önce tahıl ve stratejik kaynak tedarikini, ardından araştırma, ekonomi,
  donanma ve ordu yatırımlarını tamamlıyor; ilişki onarımı yalnız kalan bütçeyle
  çalışıyor. `1300_ottoman_rise` bütçesinde bu harcama `FlexibleGold` ile sınırlı.
  Uygun harcama kararından sonra AI'nin gönderim için deterministik `%60` başarı
  zarı bulunuyor; başarısız sonuçta altın ve ilişki değişmiyor.

- 2026-08-07: Bölge panelindeki memnuniyet barının yüzde geometrisi korunarak
  yanında tur başı `+N`/`-N` etkisi gösterildi. Etiket pozitifse yeşil, negatifse
  kırmızı çiziliyor; hover cursor'ı parmağa dönüşüyor ve ortak breakdown popup'ı
  vergi, bina, tahıl, teknoloji, savaş, genişleme, ordu, yıllık yıpranma ve
  kuşatma bileşenlerini ayrıntılı gösteriyor.

- 2026-08-07: AI'nin vergi artırma/azaltma ve `Heyet`/`Hediye` ilişki bakımı
  HAMLELER akışından çıkarıldı. State mutasyonları korunurken bu düşük öncelikli
  adımlar artık görsel bekleme üretmiyor. Regression:
  `TestAIAdjustTaxesDoesNotCreateVisibleTurnSteps`,
  `TestAIRelationRepairDoesNotCreateVisibleTurnStep`.

- 2026-08-07: Vergi-memnuniyet dengesi yeniden kalibre edildi. %30 vergi nötr
  bırakıldı; altındaki her tam 10 puan `+5`, üstündeki her tam 10 puan `-10`
  memnuniyet etkisi veriyor. Böylece %100 vergi tur başına `-70` alıyor ve AI'nin
  vergi düşürüp isyanı önleme davranışı gerçek bir maliyet sinyali görüyor. İbadet
  yeri bonusu `+5`, kışla cezası `-2` olarak senaryo bina verisine güncellendi.
  Bağımsız savaş başına savaş yorgunluğu da `-3`e yükseltildi; ekonomi, AI ve HUD
  aynı ortak savaş sayımı üzerinden bu değeri kullanıyor.

- 2026-08-06: Başkent toparlanma katsayısı bölge sahibinden ayrıştırılarak
  `Army.OwnerID` faction'ına bağlandı. Oyuncu ordusu vassal başkentinde veya
  vassal ordusu overlord başkentinde `×2` alamıyor; AI recovery skoru da aynı
  owner-aware helper'ı kullanıyor. Regression: `TestArmyReplenishmentHPDoublesAtFactionCapital`,
  `TestArmyPanelCapitalReplenishmentUsesSecondBadgeOnTheLeft`.

- 2026-08-06: Aktif kuşatma altındaki bölgeye gelen AI müttefikinin savunmayı
  yendikten sonra bölgeyi kendi adına fethetmesi engellendi. İlk kuşatmacı
  bölgenin fetih hakkını koruyor; destek ordusu yalnız kuşatma desteği olarak
  kaydediliyor ve aktif kuşatma kaydı değişmiyor. Regression:
  `TestExecuteMoveAlliedSiegeSupportCannotConquerBesiegedRegion`; doğrulama:
  `go test ./internal/ai ./internal/game ./internal/state -count=1`.

- 2026-08-06: AI'nin kendi çıkarına hizmet eden aktif/bağlanabilir ticaret, ittifak ve güvenlik ilişkilerinde `Heyet`/`Hediye` kullanması eklendi. Heyet ve hediye AI-AI arasında doğrudan çözülüyor; oyuncuya gönderildiğinde mevcut diplomasi modalı yalnız `Tamam` düğmeli bildirim olarak açılıyor. Altın rezervi ve tur içi diplomasi kotası korunuyor. Regression: `TestAIUsesDelegationToReachTradeRelationThreshold`, `TestAIUsesGiftForActiveTradeRelation`, `TestAIQueuesGiftToPlayerAsDiplomacyNotification`, `TestDiplomacyRelationshipNotificationText`.

- 2026-08-06: Üst durum HUD'ına oyuncunun bağımsız savaş sayısından türetilen
  `Savaş Yorgunluğu: -N` etiketi eklendi. Etiket `Ticaret Rotası` göstergesinin
  soluna hizalanıyor; `0` yeşil, negatif değer kırmızı çiziliyor. Regression:
  `TestWarFatigueHUDTextUsesGreenZeroAndRedPenalty`.

- 2026-08-06: AI memnuniyet yönetimi eklendi. Bağımsız savaş realm'leri başına
  `-2` savaş yorgunluğu AI bina ve vergi projeksiyonlarına bağlandı; memnuniyet
  projeksiyonu `35` altına indiğinde vergi `20`, `50` altına indiğinde `10`
  puan azaltılıyor. Memnuniyet `75+` seviyesinde vergi `10` puan artırılıyor.
  Overlord/vassal realm'leri ayrı savaş sayılmıyor. Regression:
  `TestAIAdjustTaxesProtectsUnrestAndRaisesHealthyRevenue`,
  `TestAIAdjustTaxesAccountsForIndependentWarFatigue`.

- 2026-08-06: Diplomasi hedef listesine oyuncunun oynadığı devlet de eklendi.
  Kendi satırı teklif paneli açmıyor; güç ve ekonomik sıralamalara dahil olmaya
  devam ediyor. Regression: `TestOpenDiplomacyTargetRejectsPlayerFaction`.

- 2026-08-06: Diplomasi teklif panelinde pasif eylemlerin altında pasiflik
  nedeni yeniden gösteriliyor. `ActionBlockReason` ile hesaplanan gerçek engel
  metni, `PASİF` etiketinin altındaki ortak açıklama satırında çiziliyor;
  aktif eylemlerin kabul olasılığı metinleri korunuyor. Doğrulama:
  `go test ./internal/render`.

- 2026-08-06: Diplomasi teklif butonlarının açıklama satırına Heyet ve Hediye
  ödeme akışı sağa hizalı olarak eklendi: `Karşı devlete ödeme gitmez` ve
  `Karşı devletin hazinesine 80 altın gider`. Metinler ortak buton geometrisinde
  sol taraftaki ilişki/kabul açıklamasıyla çakışmayacak şekilde yerleştiriliyor.

- 2026-08-06: Diplomasi hedef listesine `Hazine` (`Gelir/Altın`) sütunu ve
  `Ekonomik Sıralama` düğmesi eklendi. Ekonomik sıralama önce tur başı brüt
  geliri, eşitlikte mevcut hazineyi kullanıyor. Regression:
  `TestDiplomacyEconomicSortUsesIncomeThenTreasury`.

- 2026-08-06: Diplomasi hedef listesindeki askerî güç değeri artık yalnız toplamı
  değil `ordu/donanma` kırılımını gösteriyor; güç sıralaması toplam etkin güçten
  hesaplanmaya devam ediyor. Ortak `MilitaryPowerBreakdown` helper'ı kullanıldı.

- 2026-08-06: Başkentte `×2` toparlanma uygulanan kara ordularının birim
  kartlarında, mevcut yeşil `+` rozetinin soluna ikinci aynı rozet eklendi.
  Donanma kartları ve normal bölgeler tek rozet davranışını koruyor. Regression:
  `TestArmyPanelCapitalReplenishmentUsesSecondBadgeOnTheLeft`.

- 2026-08-06: Üst HUD'a `Ticaret Rotası: kullanılan/limit` göstergesi eklendi.
  Çift yönlü rota kayıtları tek partner sayılır; açık partner slotu kırmızı,
  partner limiti dolu durum yeşil görünür. Gösterge dinamik liman+pazar bonuslarını
  da kullanır ve `Elçi` göstergesinin solundaki ortak banda sağdan hizalanır.
  Regression: `TestTradeRouteHUDTextShowsOpenPartnerSlotsInRed`,
  `TestTradeRouteHUDTextShowsFullPartnerLimitInGreen`.

- 2026-08-06: AI bina yatırım puanlaması ticaret hedeflerini doğrudan dikkate
  alıyor. Maksimum pazar seviyesi `+2` rota hacmi tavanı, maksimum liman+pazar
  kombinasyonu `+1` partner bonusu olarak skorlanıyor; liman adayları yalnız kıyı
  bölgelerinde açılıyor. Regression: `TestAIBuildingInvestmentTargetsMaximumMarketTradeVolume`,
  `TestAIBuildingInvestmentTargetsMaximumPortForFullTradeRegion`; doğrulama:
  `go test ./...`.

- 2026-08-06: Ticaret partneri ve anlaşma hacmi bina gelişimine bağlandı. Temel `4`
  dış partner limiti, aynı bölgede maksimum liman+pazar seviyesine ulaşan her bölge
  için `+1`; temel `4` rota hacmi tavanı ise devletin her maksimum pazar bölgesi
  için `+2` alıyor. Dinamik limitler rota kurulumu, teklif değerlendirmesi,
  `SanitizeTradeRoutes()`, kapasite yeniden dengelemesi ve ticaret paneliyle ortak
  kullanılıyor. Regression: `TestTradePartnerLimitGrowsPerFullyDevelopedPortMarketRegion`,
  `TestTradeRouteAmountLimitGrowsPerFullyDevelopedMarketRegion`; doğrulama:
  `go test ./internal/...`.

- 2026-08-06: Sahip olunan ulusal başkent bölgesindeki kara ordusu toparlanması
  normal bölgesel değerin `×2` katına çıkarıldı. Çiftlik/ambar tabanlı ücretsiz
  toparlanma ve kapasite üzeri tahılla finanse edilen tavan aynı canonical helper'ı
  kullandığı için başkent bonusu iki akışta da tutarlı; normal bölgeler ve AI
  recovery skorları değişmedi. Regression:
  `TestArmyReplenishmentHPDoublesAtFactionCapital`,
  `TestApplyEconomyTickDoublesArmyReplenishmentAtFactionCapital`.

- 2026-08-06: Piyasa fiyatındaki arz tanımı düzeltildi. `Toplam stok` artık açık
  pazar arzı gibi kullanılmıyor; fiyat ve fiyat ekranındaki `Pazar Arzı` sütunu
  kalan `SellOffers` toplamını kullanıyor. Satış arzı sıfırken tahıl fiyatı stok
  fazlası nedeniyle taban fiyata düşmüyor; save/yeni oyun/AI turu/tur sonu
  akışları ortak fiyat yenileme yardımcısını kullanıyor. Regression:
  `TestMarketPriceUsesOpenMarketSupplyInsteadOfReservedStock`,
  `TestOpenMarketSupplyExcludesReservedFactionStock`; doğrulama:
  `go test ./... -count=1`.

- 2026-08-06: Ekonomi tick'i sonunda tahılı sıfır olan fraksiyonların tüm kara
  bölgelerinde memnuniyet tur başına `-5` azaltılıyor; düşük memnuniyet aynı tur
  çözümünde isyan kontrolüne giriyor. Regression:
  `TestApplyEconomyTickAppliesZeroGrainPenaltyToAllOwnedRegions`.

- 2026-08-06: Pazar miktar kontrolü oyuncunun mevcut altın değerine bağlandı.
  Seçili malın fiyatını aşacak `+10` artışı pasifleşiyor; eski miktar değeri de
  altınla karşılanabilir üst sınıra kırpılıyor. Çizim, input ve cursor hit-test
  aynı ortak buton durumunu kullanıyor. Regression:
  `TestTradeMarketPlusTenStopsAtGoldValueLimit`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-08-06: Tamamlanmış `consolidate` objective'lerinde claim kaybı recovery
  akışına bağlandı. Başka devletin aldığı claim bölgeleri güncel sahiplerine göre
  dinamik `expand` hedefi olur; AI aynı objective kimliğiyle geri alma planı kurar.
  Claim geri alındığında recovery kapanır. Regression:
  `TestConsolidationObjectiveRecoversLostClaimFromCurrentOwner`.

- 2026-08-06: AI strateji objective tamamlanması eklendi. `expand` ve
  `consolidate` objective'lerinin tüm claim bölgeleri AI devletinin elindeyse
  objective tamamlanıyor ve tekrar seçilmiyor. Sonraki tarih/event kapısı henüz
  açılmadıysa AI geçici `consolidate:<faction_id>` hazırlık planına geçiyor;
  `defend` objective'leri yalnız claim sahipliğiyle kapanmıyor. Regression:
  `TestScenarioObjectiveCompletionUsesClaimsButKeepsDefendIntent` ve 1300
  İngiltere/Safevî senaryo plan testleri.

- 2026-08-06: Gedik oluşan kuşatmaların settlement marker'ı, marker seçili
  olmasa da görülebilen yeşil dış halka ile işaretleniyor. Kapsam:
  `internal/render/renderer.go`.

- 2026-08-06: Saldıran ve savunan kuşatma panellerinde kuşatmanın kaç turdur
  sürdüğü sağ üst durum etiketinde gösteriliyor. Kapsam:
  `internal/render/renderer_dialogs.go`.

- 2026-08-06: Seçili bölge ve devlet detay panellerinin ortak genişliği 305 px'den
  315 px'e çıkarıldı. Bina kartları, sekmeler, aksiyon alanı ve devlet paneli
  içeriği aynı ortak ölçüden türetilerek genişletildi; input/hit-test ve scroll
  geometrisi korunuyor. Kapsam: `internal/render/panel.go`.

- 2026-08-06: Edit Mode bina kuralları otomatikleştirildi. Kale yerleşimlerine
  `walls`, liman yerleşimlerine `port`, ulusal başkent bölgelerine
  `barracks/granary/temple/market` eksikse ekleniyor; yerleşim tipi, taşınması,
  merkez/başkent değişimi ve Edit Mode yüklemesi bu ortak senkronizasyondan
  geçiyor. 1300 senaryosundaki mevcut altyapı eksikleri tamamlandı. Regression:
  `Test1300ScenarioSettlementInfrastructureHasMinimumBuildings`,
  `TestEditModeSetsSelectedSettlementAsFactionCapital`,
  `TestEditModeSettlementTypeFillsRequiredBuildingAndUndo`.

- 2026-08-06: `DEV_MODE=true` iken devlet bilgi panelinin en altına AI strateji
  teşhisi eklendi. Aktif objective, profil, min/max yıl aralığı ve claim
  bölgelerinin güncel sahipleri gösteriliyor; diğer objective'ler de tarih ve
  claim sayısıyla listeleniyor. Bölüm mevcut panel scroll'u ve ortak UI
  yardımcılarını kullanıyor. Regression: `TestFactionAIDebugIsDevOnlyAndExcludesPlayer`,
  `TestFactionAIDebugUsesActiveObjectiveAndDynamicClaimOwner`.

- 2026-08-06: AI barış kararları tüm senaryolarda ortak savaş değerlendirmesine
  bağlandı. İlk dört savaş turunda olağan barış kapatıldı; `TerritorialClaims`,
  aktif expand hedefleri ve `WarLedger` hedefi düşmanın tuttuğu core/claim
  bölgeleriyle karşılaştırılıyor. Core işgali acil durum dışı barışı engelliyor,
  claim değeri barış eşiğini yükseltiyor. Savaş sonrası ilişki artık `-45/-60/-70`
  tabanından toparlanıyor. Regression: `TestPeaceAssessmentRespectsUnresolvedTerritorialClaims`,
  `TestPeaceAssessmentBlocksUnresolvedCoreOutsideEmergency`, hedefli `internal/diplomacy`
  ve `internal/ai` testleri.

- 2026-08-06: Bölgesel hedefler devlet adına sabitlenmekten çıkarıldı. Claim'ler
  `ai_strategies.json` içindeki `territorial_claims` ve objective bölgelerinden
  materialize ediliyor; AI hedef devleti her tur claim bölgesinin güncel
  sahibinden dinamik seçiyor. `ai_expansion_targets` verileri de strateji
  `expansion_targets` alanına taşındı ve runtime uyumluluk için faction state'ine
  kopyalanıyor.

- 2026-08-06: AI objective zaman sınırı eklendi. `min_year` başlangıç kapısı,
  kapsayıcı `max_year` son geçerlilik yılıdır; son yıllarda objective puanına
  zaman baskısı eklenir ve `max_year` sonrasında hedef kapanır. Bölgesel hedefler
  objective içindeki `territorial_claims` alanında tek yerde tutulur.

- 2026-08-05: Komutan geliş bildirimindeki çoklu kart listesi modal dışına
  taşmayacak şekilde panel-local viewport'a alındı. İki görünür satırdan sonraki
  komutanlar mouse wheel ve scrollbar ile kaydırılıyor; `SubImage` clipping ile
  kartlar devam metnine veya modal sınırına çizilmiyor. Regression:
  `TestCommanderArrivalScrollViewportAndClamp`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-08-05: Ana menüden devam et veya kayıt slotundan yükleme akışında, yükleme
  ekranı kayıt metadata'sından çözülen senaryonun `scenario_bg.png` görselini de
  kullanıyor. Senaryo seçimi akışındaki mevcut arka plan davranışı korunuyor.
  Regression: `ScenarioPathForSlot`, `TestLoadingBackgroundLoadsFromScenarioDirectory`;
  doğrulama: `go test ./internal/save ./internal/game ./internal/render -count=1`.

- 2026-08-05: Ticaret paneli görsel yerleşimi düzeltildi. Aday ve pazar listelerinde
  yinelenen/taşan sayfalama satırı kaldırıldı ve ortak `ListView` sayfalaması için
  footer boşluğu ayrıldı. Pazar işlem düğmeleri kart içinde kompakt bir gruba
  taşındı; rota özeti değerleri etiketlerine yaklaştırıldı. Regression:
  `TestTradeMarketActionButtonsStayGroupedInsideActionCard`,
  `TestCoreUIGeometryFitsCommonViewports`; doğrulama: `go test ./internal/render`.

- 2026-08-05: Harita ordu çizimi iki geçişe ayrıldı; tüm komutan portreleri
  marker ve görev rozetlerinden önce, ortak marker/rozet katmanı sonra çiziliyor.
  Böylece yakın kara veya deniz marker'larında komutan resimleri rozetlerin
  üstüne çıkmıyor. Regression: `go test ./internal/render`.

- 2026-08-05: Pazar panelindeki ayrı mal listesi kaldırıldı. Tahıl, demir,
  kereste, taş, baharat ve kumaş üstte ortak mal filtreleri olarak seçiliyor;
  seçilen mala göre tek devlet listesi güncelleniyor. Her satırda `Stok`,
  `Satış arzı`, `Alım talebi` ve `Fiyat` birlikte gösteriliyor; işlem
  kartı aynı seçili malı kullanıyor. Regression:
  `TestHandleTradePanelInputMarketSelectsFactionAndGoodOnClick`,
  `TestCoreUIGeometryFitsCommonViewports`; doğrulama: `go test ./internal/render`.

- 2026-08-05: Ticaret paneli düğmeleri ortak dikey merkezleme hesabına alındı;
  miktar kontrolü ikonsuz `-10` / `+10` adımlarına indirildi. İşlem kartındaki
  `Miktar` / `Tutar` satırı büyütüldü, pazar satırlarından oyuncunun zaten HUD'da
  görünen kendi stoku kaldırıldı ve panel yeni açılışta `Pazar` sekmesini gösteriyor.
  Regression: `TestTradeQuantityButtonsUseOnlyTenStepsWithoutIcons`,
  `TestTradeButtonsUseCommonVerticalCentering`, `TestTradePanelOpensOnMarketTab`.

- 2026-08-05: Açık pazar işlemleri AI devletlerinin tur başında belirlediği
  satış fazlası ve alım talebi kotalarına bağlandı. Ticaret paneli ham stoktan
  ayrı olarak `Satış arzı` ve `Alım talebi` gösteriyor; oyuncu yalnız satıcının
  arzını satın alabiliyor veya hedefin altınla karşılayabileceği talebe satış
  yapabiliyor. Başarılı transferler kota bakiyesini azaltıyor. Regression:
  `TestMarketOrdersClampTransactionsToOfferAndGoldBackedDemand`,
  `TestMarketOrdersCloneIsDeepCopy`; doğrulama: `go test ./internal/state
  ./internal/ai ./internal/game ./internal/save`, ticaret render testleri.

- 2026-08-05: Mevcut Rotalar sekmesine `Tüm Rotalar` / `Sadece Bize Ait`
  filtreleri eklendi. Varsayılan görünüm oyuncunun rota uçlarından birine ait
  olduğu rotaları gösteriyor; filtre butonları ortak ticaret paneli geometrisi,
  hit-test ve input akışını kullanıyor. Gereksiz kırmızı `İptal` başlığı kaldırıldı.
  Regression: `TestTradeRouteListFilterDefaultsToOwnedAndTogglesAllRoutes`;
  doğrulama: `go test ./internal/render`.

- 2026-08-05: `Sadece Bize Ait` rota görünümünün altına tur başına toplam
  `Gelir`, `Gider` ve `Net Fark` özeti eklendi. Değerler rota yönü ve ortak
  `GoldEarned()` hesabından türetiliyor; askıdaki rotalar dışarıda tutuluyor.
  Regression: `TestFactionTradeStatsCalculatesRouteIncomeExpenseAndNet`.

- 2026-08-05: Pazar erişimi Ticaret Haritası üst HUD'ından alt aksiyon HUD'ına
  taşındı. Alt bölüm `Ordu → Pazar → Diplomasi → Teknoloji → Tur Bitir` sırasıyla
  beş butona genişletildi; Pazar, Diplomasi'nin hemen solunda diğer alt HUD
  butonlarıyla aynı ölçüyü ve ortak çizim/hit-test geometrisini kullanıyor.

- 2026-08-05: 1300 AI tahıl yatırımı yalnız ordu bakımını değil sivil tüketimi
  de hesaplamaya başladı; bu yüzden nüfusu yüksek fakat az ordulu devletler
  çiftliği doğru zamanda pazarın önüne alıyor. Aynı kıtlığa aşırı bina kuyruğu
  açılmaması için en fazla iki tamamlanmamış çiftlik tutuluyor. Stratejik tahıl
  eksiğinde AI, aktif rota aramadan savaşta olmadığı ve satıcının üç aylık
  rezervini koruyan tüm devletlerin açık pazarından aynı turda alım yapabiliyor.
  Oyuncu da `Pazar` sekmesinde aktif rota şartı olmadan aynı savaş dışı devletlerle
  alım/satım yapabiliyor. Regression:
  `TestAIBuildingInvestmentCountsCivilianDemandForFarmPriority`,
  `TestAIProcuresGrainFromOpenMarket`,
  `TestAIOpenMarketNeverBuysFromEnemy`,
  `TestCanPlayerOneTimeTradeWithUsesOpenMarketButExcludesEnemies`,
  `TestAIGrainProcurementUsesReserveInsteadOfStorageCapacityAsSupplierLimit`;
  doğrulama: `go test ./internal/ai ./internal/game ./internal/render`.

- 2026-08-05: Diplomasi teklif panelinde pasif aksiyonların uzun blok sebebi
  satır içinden kaldırıldı; yalnız etiket ve `PASİF` göstergesi kalıyor. Ortak
  `Box` yerleşimi footer'ı önce ayırıyor; iki satırlı durum kartı ve aktif teklif
  ayrıntıları alt çerçevelerinden güvenli boşlukla çiziliyor. Regression:
  `TestCoreUIGeometryFitsCommonViewports`; doğrulama: `go test ./internal/render`.

- 2026-08-05: Seçili devlet detay panelinde güç sırası bölgelerin üstüne büyük,
  kalın altın vurgu olarak taşındı. Askerî güç artık `Kara / Deniz Gücü` olarak
  ayrı görünür; sıralama ikisinin toplamını kullanır. Devlet başlığı sabit
  turkuaz, bayrak rozeti ise fraksiyon rengindedir. Regression:
  `TestMilitaryPowerBreakdownIncludesNavalStrength`,
  `TestFactionMilitaryPowerBreakdownLabelSeparatesLandAndNavalStrength`.

- 2026-08-05: Aktif Savaşlar paneli alfabetik taraf düzeni yerine savaş ilanı
  yönünü gösteriyor: ilk ilan eden solda, savunan sağda; kayıp/güç/ordu değerleri
  aynı yönle eşleşiyor. Yön bilgisi `WarLedger` ile save/load'a eklendi ve savunan
  ittifaka sonradan katılımda korunuyor. Eski save'ler yön metadata'sı taşımadığında
  korunan ilişki taraf sırası kullanılır. Regression: `TestCollectActiveWarSummariesShowsTurnsStrengthAndArmyCounts`,
  `TestResolveAcceptedWarJoinOfferAddsPlayerToWar`; doğrulama:
  `go test ./internal/state ./internal/diplomacy ./internal/render ./internal/save`.

- 2026-08-05: 1300 AI kara birimi seçimi, mevcut plan kompozisyon oranlarını
  (`expand %55/%25/%20`, `defend %75/%15/%10`, fallback `%65/%25/%10`) korurken
  aynı kategori içindeki güç/tahıl verimini ayrıca puanlamaya başladı. Yakın tahıl
  bakımı karşılığında belirgin savaş değeri veren elit piyade, teknoloji, üretim
  hattı ve bütçe uygunsa milise tercih ediliyor; tahıl darlığında mutlak bakım
  cezası korunuyor. Plan/cephe üzerindeki farklı tahkimli düşman bölgeleri
  sayılıyor ve en fazla üç birliklik kuşatma kolu, sahadaki-bekleyen kuşatma
  kapasitesi tamamlanana kadar zorunlu üretim açığı oluşturuyor. Regression:
  `TestAIStrategicRecruitmentPrefersEliteWithSuperiorPowerPerGrain`,
  `TestAIFortifiedCampaignBuildsSiegeCorpsForMultipleTargets`; doğrulama:
  `go test ./internal/ai`.

- 2026-08-05: Aktif tarihsel/stratejik `expand` hedefleri için savaş kararı
  yalnız tek devletin anlık toplam gücüne bağlanmaktan çıkarıldı. Hedefin
  vassal ve müttefiklerinden oluşan savunma koalisyonuna karşı, otomatik katılan
  ve savaş çağrısını en az `%70` olasılıkla kabul eden yakın AI müttefiklerinin
  mesafe-ağırlıklı gücü sayılıyor. Yetersiz hedef sahibi önce hedefe baskı
  yapabilecek devletle ittifak arıyor; ittifakın ardından ortak güç eşiği geçerse
  aynı tur savaş açabiliyor. Doğrudan hedef bölgesinde `%125` yerel sınır
  üstünlüğü bulunan planlar, lojistik ve diğer güvenlik kapıları korunarak `%85`
  toplam koalisyon eşiğiyle hızlı fetih açılışı yapabiliyor. Regression:
  `TestAIHistoricalPlanCanOpenWarWithReliableNewAlly`,
  `TestAIHistoricalPlanUsesRapidConquestOnlyWithBorderSuperiority`,
  `TestAIWarAssessmentIncludesReliableAttackingAlly`; doğrulama:
  `go test ./internal/ai`.

- 2026-08-05: AI askerî tabanı nüfus ve kıyı ölçeğine bağlandı. Kara rezervi
  `1 birlik / 200 nüfus` temelinden, plan/savaş tehdidi çarpanlarıyla ve gerçek
  manpower sınırıyla türetiliyor; eksik kuvvet, saldırı rotası henüz oluşmamış olsa
  bile güvenli-ikmalli iç üretim hattından tamamlanıyor. Genel filo hedefi her iki
  kıyı için bir savaş gemisi; `ai_strategies.json`daki `naval_focus` ile işaretlenen
  denizci devletler kıyı başına iki savaş gemisi ve en az altı gemilik ana filo kurup
  deniz bütçesine `%35` ayırıyor. Hedef teknoloji → liman → savaş gemisi zincirini ve
  eksik reçete girdilerinin ticaret ağından alımını tetikliyor. `expand` planındaki
  hedef devletin aktif+bekleyen kara gücünün `%135`i de ayrı bir kara birlik hedefi
  üretiyor; hedef devlet güçlendikçe tarihsel fetih hazırlığı büyüyor. Regression:
  `TestAIForceRequirementsScaleLandReserveWithPopulationAndPlan`,
  `TestAIExpansionReserveExceedsHistoricalTargetProjectedLandPower`,
  `TestAIReserveRecruitmentQueuesOnlyPopulationBasedShortfall`,
  `TestAINavalReserveBuildsPortBeforeWarship`; doğrulama: `go test ./internal/ai`.

- 2026-08-05: AI geri çekilme anchor'ları yeni ikmal-toparlanma modeline bağlandı.
  Kapasitesi aşılmayan adaylarda Çiftlik/Ambar seviyesinin gerçek toparlanma hızı
  puana katılıyor; kısa rota maliyetine rağmen ağır yıpranmış ordu daha hızlı
  iyileşeceği güvenli bölgeyi seçebiliyor. AI ikmal tahmini, ambarın yerel stok
  desteğini de tur çözümlemesindeki öncelikle hesaba katıyor. Regression:
  `TestRecoveryAnchorPrefersFasterFarmAndGranaryRegion`; doğrulama:
  `go test ./internal/ai ./internal/game ./internal/state`.

- 2026-08-05: Merchant ticaret gemileri yalnızca kendi fraksiyonunun ihracat
  yönündeki rotalara atanabilir hale getirildi. Karşı tarafın ihracat rotası
  artık merchant panelinde görünmüyor ve elle/legacy atanmış olsa bile hacim
  bonusu üretmiyor. Regression: `TestMerchantTradeRoutesForFleetFiltersByOwnerAndActiveCenter`;
  doğrulama: `go test ./...`.

- 2026-08-05: Merchant bonusu artık filo başına sabit `+2` değil, gemi başına
  `+1` hacim olarak hesaplanıyor. Bonus rota panelinde görünen ilgili rota
  kapasitesiyle sınırlı; merchant üretimi bu kapasiteye kadar tek filoya ekleniyor
  ve harita kalabalığı azalıyor.
  Regression: `TestMerchantTradeBonusUsesAssignmentLocationAndRouteCap`,
  `TestNavalProductionKeepsMerchantAndMilitaryFleetsSeparate`; doğrulama:
  `go test ./...`.

- 2026-08-05: Aynı rotanın merchant kapasitesi ortak atama sayısıyla dolduruluyor.
  Kapasitesi dolu rota yeni filoya atanamıyor; merchant rota modalı dolu satırı
  pasif gösterip tıklamayı engelliyor, mevcut aktif filo satırını koruyor.
  Regression: `TestMerchantTradeRouteCapacityBlocksAdditionalFleetAssignment`,
  `TestMerchantRouteOptionDisabledWhenCapacityFull`; doğrulama: `go test ./...`.

- 2026-08-05: Merchant rota paneli ve ordu footer'ı artık bonusun verilen/maksimum
  payını `+2/4` biçiminde gösteriyor. Böylece aynı rotadaki başka filoların
  kapasiteyi tüketmesi görünür hale geliyor. Regression:
  `TestMerchantFleetTradeRouteCapacityBonusShowsSharedRouteUsage`; doğrulama:
  `go test ./...`.

- 2026-08-05: Görevi atanmış fakat hedef denize ulaşmamış merchant filoları
  artık kapasiteyi tüketmiyor; marker üzerindeki bonus rozeti `+N` yerine `X`
  göstererek görevin beklemede olduğunu belirtiyor. Regression:
  `TestMerchantTradeBonusForArmyOnlyShowsActiveTargetSeaBonus`; doğrulama:
  `go test ./...`.

- 2026-08-05: Kara ordusu toparlanması sabit `+10 HP` yerine bölgesel ikmal
  sonucuna ve bina seviyelerine bağlandı. İkmal talebi kapasiteyi aştığında
  toparlanma sıfır ve lojistik yıpranması sürer; yeterli ikmalde hız
  `2 + 2 × (Çiftlik + Ambar seviyesi)` HP/birim/turdur. Kapasite üstü tahılla
  gelen bedelli yenileme de aynı tavanı kullanır. Regression:
  `TestApplyEconomyTickReplenishesFriendlyLandArmyByFarmAndGranaryLevel`,
  `TestApplyEconomyTickDoesNotReplenishArmyWhenRegionalSupplyIsOverloaded`;
  doğrulama: `go test ./internal/game ./internal/state ./internal/render`.

- 2026-08-05: Barış kabul edildiğinde karşı tarafın kara bölgesinde bulunan
  ordular artık hareket puanından bağımsız olarak en yakın kendi, vassal veya
  müttefik bölgesine zorunlu çekiliyor. Terk edilen kuşatmalar ve geçersiz
  bekleyen temaslar temizleniyor; deniz çıkarması için önce nakliye filosuna
  binme tercihi korunuyor. Tur sonuna ayrıca hatalı/legacy save konumlarını
  temizleyen erişim denetimi eklendi; savaş, ittifak ve aynı realm geçerli,
  diğer yabancı kara konumları zorunlu tahliye ediliyor. Ayrı askerî geçiş izni
  henüz plan notudur; eklendiğinde bu denetime dahil edilecek. Regression:
  `TestAcceptedPeaceEvacuatesLandArmyRegardlessOfMovePoints`,
  `TestEvacuateArmiesFromPeaceTerritoryUsesNearestAlliedLand`,
  `TestEvacuateArmiesWithoutLandAccessRepairsStalePeaceOccupation`;
  doğrulama: `go test ./...`.

- 2026-08-05: `scenario.json` içindeki `victory_conditions` AI stratejik
  planına bağlandı. Tarihsel fraksiyon hedefleri eksik/erişilebilir bölgelerin
  sahibini öncelikli genişleme hedefi yapıyor; özel hedefi olmayan devletler
  genel zafer koşullarını kullanıyor. Ekonomik ve hayatta kalma hedefleri
  konsolidasyona, askerî hedefler uygun kara komşusuna yönlendiriyor. Bu plan
  savaş fırsatı puanına ek proaktivite veriyor ve 1300 dışındaki, zafer verisi
  yüklü senaryolarda da çalışıyor. 1300'de mevcut profilin tarih/yıl/event
  açılış kapıları korunuyor. Regression:
  `TestHistoricalVictoryConditionOverridesProfilePlan`,
  `TestGeneralVictoryConditionGuidesFactionWithoutHistoricalGoal`,
  `TestEnsureStrategicPlanUsesVictoryConditionsInAnyScenario`,
  `TestVictoryPlanAllowsOpportunityWarOnMildPeace`.

- 2026-08-05: AI barış dönemindeki pasiflik düzeltildi. Kritik tehdit yokken
  genişleme objective'leri savunma objective'leri tarafından tamamen
  gölgelenmiyor; savunma planındaki yeterli sınır orduları fırsat savaşı
  açabiliyor. Uyarı seviyesindeki lojistik eksikliği savaşı kilitlemiyor,
  gerçek tahıl krizi hâlâ saldırıyı durduruyor. Regression:
  `TestStrategicWarReadyUsesBorderForceDuringDefensivePlan`,
  `TestStrategicWarLogisticsGatePreservesOpeningTempo`; doğrulama:
  `go test ./...`.

- 2026-08-05: 1300 senaryosunda demir üretimi olmayan kara bölgelerine `2 demir/tur`
  taban üretimi verildi. Böylece demir yatağı bulunmayan küçük ve kıyı devletleri
  temel askerî birlikleri üretebilirken, mevcut 4–20 demirlik uzmanlaşmış bölgeler
  yüksek üretim merkezleri olarak kaldı. Regression:
  `Test1300LandRegionsHaveBaselineIronForMilitaryProduction`.

- 2026-08-04: Ticaret haritasındaki rota filosu marker'ları sadeleştirildi.
  Ticaret gemisi ve görev rozetleri korunurken komutan portresi yalnız bu
  overlay çağrısında kapatıldı; normal harita ordu/donanma görünümü portreyi
  göstermeye devam ediyor.

- 2026-08-05: Ticaret merkezi bonusları sabit `+2/+4` değerlerinde kalmıyor.
  Hacim 50'yi aştığında primary merkez her 25, secondary merkez her 50 hacimde
  kapasiteye +1 ve gümrük gelirine +2 ekliyor; üst sınır kaldırıldı. Tabela hacmi
  ve ekonomi aynı `TradeCenterVolume()` hesabını kullanıyor. Regression:
  `TestTradeCenterBenefitsGrowWithVolumeWithoutUpperLimit`.

- 2026-08-04: Ticaret haritasında ana ve ikincil merkezlerin görsel hiyerarşisi
  ayrıştırıldı. İkincil merkez tabelaları daha küçük/düşük kontrastlı; hacimsiz
  merkez hatları ince sürekli gri; aktif hatlar parlak kalıyor. Ana merkez tabelası
  96×96 kaynaklı `trade_center.png` ikonunu çerçeveli sol karede gösteriyor ve
  daha büyük çiziliyor. Regression: `TestTradeCenterIconLoadsAtNativeResolution`.

- 2026-08-04: 1300 senaryosundaki komutan şablonları artık `start_year` ve
  `end_year` ile tarihsel olarak ortaya çıkıyor. Aktif komutanlar seçim listesi
  ve atama state'i tarafından ortak aralık helper'ıyla filtreleniyor; `end_year`
  başladığında atanmış komutan da ordudan emekli ediliyor. Oyuncu, ilk göründükleri
  yılda portre, seviye, özellik ve görev aralığını gösteren ortak modal popup alıyor.
  Regression: `TestCommanderAvailabilityAddsArrivalsAndRetiresExpiredAssignments`,
  `TestCommanderActiveYearHasExclusiveEndBoundary`; doğrulama: `go test ./...`.

- 2026-08-04: 1300 ticaret merkezi grafiği denetlendi ve oynanabilir devlet
  erişimi tamamlandı. Osmanlı için Bithynia, Aragon için Katalonya, HRE için
  Hollanda ve Moskova için Moskova ikincil merkezleri ana ağa eklendi; eski
  tek yönlü linkler karşılıklanarak graf tamamlandı. Ticaret haritası yalnız
  ana/ikincil merkezleri etiketliyor; devlet başkentleri en yakın merkeze ince
  sabit çizgiyle bağlanıyor. Regression:
  `Test1300PlayableFactionsOwnConnectedTradeCenters`.

- 2026-08-04: Norveç/Oslo ve İsveç/Stockholm bölgeleri 1300 ticaret haritasında
  düğümsüz kaldığı için görünür bağlantı oluşturmuyordu. Norveç `Danimarka`ya;
  İsveç `Danimarka` ve `Novgorod`a bağlanan ikincil merkezler olarak eklendi.

- 2026-08-04: Fas, Girit ve Rodos da ticaret merkezi grafiği dışında kalmıştı.
  Fas `Cezayir/Portekiz`, Girit `Venedik/Konstantiniyye/Mısır`, Rodos ise
  `Konstantiniyye/Mısır` koridorlarıyla ikincil merkez olarak bağlandı.

- 2026-08-04: Edit Mode harita fırçasında yanlışlıkla yapılan sınır/bölge
  tıklamalarını düzeltmek için `Shift+sol tık` ve `Shift+drag` aktif boya/sil
  işlemini tersine çeviriyor. Cursor preview ve yardım metni de ters işlem
  durumunu gösteriyor; regression: `TestEditShapeBrushShiftReversesPaintAction`.

- 2026-08-04: AI eksik üretim kaynağı tedariki askerî önkoşullarla eşlendi. Eksik
  demir artık tedarik öncesi aday puanında piyadeyi cezalandırıp demirsiz milisin
  seçilmesine yol açmıyor. Geçerli askerî üretim bölgesi kışla eksikliği nedeniyle
  bulunamıyorsa, kışlanın demir/kereste dahil tüm `ResourceCost` girdileri aktif
  ticaret ağından alınarak gerçek kışla kuyruğuyla aynı karar sözleşmesi kullanılıyor.
  Regression: `TestAIProcuresBarracksInputsBeforeMilitaryProduction`,
  `TestAIProcurementDoesNotChooseMilitiaToAvoidMissingIron`;
  doğrulama: `go test ./internal/ai`.

- 2026-08-04: Ticaret merkezleri fethedildiğinde kapasite bonusuna ek olarak
  görünür ve veri tanımlı gümrük geliri sağlıyor: ana merkez `+2 kapasite,
  +4 altın/tur`; ikincil merkez `+1 kapasite, +2 altın/tur`. Gümrük normal
  mevsim, abluka ve pazar teknolojisi çarpanlarını izliyor. Eski merkez JSON'ları
  aynı değerleri loader migration'ıyla alıyor. Ticaret haritası
  etiket/tıklama bilgisi bu faydaları gösteriyor; Edit Mode snapshot'ı yeni
  alanları koruyor. Regression: `TestApplyEconomyTickAddsTradeCenterCustomsIncome`.

- 2026-08-04: Dış ticarette `4` partner sınırı artık yalnız teklif kontrolü
  değil; rota kurma, ilişki onarma ve save/load sanitize yollarının ortak
  kuralı. Efektif ticaret kapasitesi aktif dış anlaşmalara paylaştırılıyor,
  rota başına taban hacim iki tarafın payından düşük olanla ve `4` üst sınırıyla
  belirleniyor. Yeni rota paneli kullanılan/toplam kapasiteyi gösteriyor.
  Regression: `TestEnsureTradeRoutesForActiveRelationsEnforcesPartnerLimit`,
  `TestRebalanceTradeRouteCapacitiesSharesEffectiveCapacity`.

- 2026-08-04: Harita input sözleşmesi düzeltildi. Tek başına `S` artık hızlı
  kayıt almıyor ve kamerayı aşağı kaydırıyor; hızlı kayıt yalnızca `Ctrl+S`
  ile tetikleniyor (`internal/render/renderer_input.go`).

- 2026-08-04: Ticaret kapasitesi tek kanonik state hesabına taşındı.
  Pazar/liman yanında ambar `x1.05` ve ibadethane `x1.03` katkı veriyor;
  primary/secondary ticaret merkezleri senaryo verisinden sırasıyla `+2/+1`
  efektif kapasite sağlıyor. Aynı değer pasif ticaret geliri, diplomasi eşiği,
  rota hacmi, AI değerlendirmesi ve ticaret UI'ında kullanılıyor; merkez fethedilince
  bonus yeni sahibine geçiyor. Regression: `TestEffectiveRegionTradeCapacityUsesBuildingsAndTradeCenter`,
  `TestEffectiveFactionTradeCapacityFollowsConquest`,
  `TestAssessTradeProposalUsesEffectiveBuildingCapacity`.

- 2026-08-04: Komutan atama listesindeki satır seçimi, `Yeni Komutan`
  düğmesinin fare edge state'ini erkenden tüketmesi nedeniyle çalışmıyordu.
  Komutan paneli tıklamayı frame başına bir kez okuyup kapatma, üretim,
  ayırma ve liste hit-test'lerinde aynı değeri kullanıyor; satır ataması yeniden
  tetikleniyor.

- 2026-08-04: Seçili ordu panelinin üst bilgi alanı ferahlatıldı. Başlık,
  lojistik/tahıl ve takviye/ikmal satırları artık ayrı dikey aralıklarda; BÖL ve
  BİRLEŞTİR düğmeleri bu satırların altına taşındı. Panel yüksekliği artırıldı,
  çizim ve hit-test ortak geometri sabitlerini kullanmaya devam ediyor.
  Regression: `TestCoreUIGeometryFitsCommonViewports` içindeki ordu paneli
  aksiyon/üst durum aralığı kontrolleri; doğrulama: `go test ./internal/render`.

- 2026-08-04: Kuşatılan bölgedeki ambar seviyesi savunma dayanıklılığına bağlandı.
  Her `granary` seviyesi savunucu ordunun doğrudan kuşatma HP yıpranmasını ve
  bölgesel ikmal açığı hasarını `%10` azaltıyor; toplam bonus `%30` ile sınırlı
  ve kuşatan ordu bu bonusu almıyor. Regression: `TestResolveSiegesGranaryReducesDefenderAttrition`,
  `TestSiegeGranaryReducesDefenderLogisticsAttrition`.

- 2026-08-04: Oyuncu komutan atama modalından `Yeni Komutan` seçip adını
  yazarak havuza yeni aday ekleyebiliyor. Yeni adaylar varsayılan portreyle,
  kalıcı sıra sayısından deterministik türetilen rastgele başlangıç XP'si ve
  buna bağlı Taktisyen/Savunma gibi uzmanlıklarla oluşuyor; save/load sonrası
  profil değişmiyor. Oyuncu ve AI runtime aday üretiminde aynı `500 altın +
  100 tahıl` maliyetini öder; oyuncu düğmesi yetersiz kaynakta pasif kalır,
  AI ise komutansız orduyu bekletir. Ad girişi ortak
  `gameui.Modal`/`TextBox`/`Button` bileşenleriyle modal önceliği, hit-test ve
  cursor sözleşmesini koruyor.
  Regression: `TestRecruitPlayerCommanderUsesEnteredNameAndDeterministicProgression`,
  `TestRecruitPlayerCommanderRejectsBlankAndTooLongNames`; doğrulama:
  `go test ./internal/state ./internal/render -count=1`.

- 2026-08-04: Seçili kara ordusu ve donanma marker'larına, komutan portresini
  de kapsayan yuvarlatılmış köşeli kesik altın seçim çerçevesi eklendi. Çerçeve
  görev ve deniz rozetlerinden sonra çizildiği için görünür kalıyor; input
  alanları değişmedi. Regression:
  `TestArmySelectionIndicatorRectCoversCommanderAndMarkerForLandAndNaval`,
  `TestArmySelectionIndicatorRectWithoutCommanderStaysCompact`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-08-04: Kuşatma açlık teslim süresi tahkimat seviyesine göre 10/14/18
  turdan 6/8/10 tura indirildi. Kuşatan ordu hedefin yanında kendi kara
  bölgesine sahipse bakım yükü `%200` yerine `%150`; başkentten uzak ya da kara
  ikmal hattı kopuksa bölgesel yıpranma talebi kademeli olarak daha yüksek.
  Ortak hesap bölgesel yıpranma ve AI lojistiğine birlikte giriyor; fraksiyon
  toplam tahıl gideri değişmiyor.
  Regression: `TestSiegeSurrenderTurnsAreShorterByFortLevel`,
  `TestEffectiveArmyGrainUpkeepUsesBorderSupplyAndCapitalDistance`.

- 2026-08-04: Müttefik veya aynı realm içindeki vassal kara bölgesine bitişik
  düşman cepheleri de ileri ikmal noktası sayılıyor. Bu sınır kuşatma yükünü
  `%150`ye indiriyor; kendi başkentine kara hattı olmasa bile yakın dost sınırı
  bölgesel yıpranma cezasını `%10`la sınırlandırıyor.

- 2026-08-04: Müttefik/vassal ileri ikmali bedelli hale getirildi. Aynı realm
  destekçi taban bakımın `%20`sini, bağımsız müttefik yaklaşık `%34`ünü tahıl
  olarak harcar; bir aylık talep veya 20 tahıllık güvenlik rezervinin altına
  düşmeye izin verilmez. Yetersiz stokta avantaj kapanır. Ordu/bölge lojistik
  yüzeyleri ve `[IKMAL]` tur olayları kaynak devleti ile harcanan tahılı
  gösterir. Regression: `TestFriendlyFrontlineSupplyCostsProviderGrainAndPrefersVassal`.

- 2026-08-04: Tahkimli bölgeye doğrudan hareket emri, renderer karar modalını
  atladığında artık genel uyarıda takılmıyor; `ShowSiegeDecision` üzerinden
  Kuşatma Başlat / Genel Hücum seçeneklerini açıyor. Kuşatma birimi olmayan
  orduların karar verilmeden hareket puanı veya konumu değişmiyor. Regression:
  `TestMoveArmyToFortifiedRegionOpensSiegeDecisionWithoutSiegeUnit`.

- 2026-08-04: 1300 senaryosunun zafer havuzu tarihsel dönüm noktalarına göre
  kalibre edildi. Oynanabilir 10 devletin her birine fraksiyona özel kart eklendi;
  genel havuz daha yüksek eşikli toprak, ekonomi, askerî, dinî ve 20 yıllık beka
  koşullarıyla genişletildi. Eski ortak 1561 son tarihleri ilgili hedeflerin
  1341–1517 aralığındaki tarihsel eşikleriyle değiştirildi. Üçten fazla kartlı
  zafer grupları ortak kart rect'iyle iki sütunda çizilerek 1280×720'de seçilebilir
  kaldı. Regression: `Test1300PlayableFactionsHaveHistoricalVictoryOption`,
  `TestCoreUIGeometryFitsCommonViewports`.

- 2026-08-04: 1300 ve 1455 senaryoları üç aylık (mevsimlik) stratejik tura geçti:
  dört tur bir yılı temsil ediyor. Bina, birlik ve teknoloji kuyrukları tur bazlı
  kaldı; süreleri ilgili senaryo veri dosyalarında doğrudan iki katına çıkarıldı.
  AI plan/savaş,
  toplanma ve barış zamanlaması gerçek takvim temposunu koruyacak şekilde
  ölçeklendi; yıllık ekonomi etkileri takvim penceresine bağlandı. Tarihsel olaylar
  artık üç aylık pencereyi kullanıyor ve aynı penceredeki ikinci olay bir sonraki
  turda güvenli biçimde işleniyor. Regression: `TestQuarterlyTurnCoversCalendarRangeAndAdvancesThreeMonths`,
  `TestTickCarriesSecondHistoricalEventIntoNextQuarter`, `TestScenarioTurnDurationsAreExplicitlyScaledInData`.

- 2026-08-04: 1300'deki büyük devletlerin AI genişleme hedefleri tarihsel uzun dönem
  eşiklerine taşındı. Osmanlı 1354 Rumeli/1453 Konstantinopolis, İngiltere 1415 Fransız
  tacı, Fransa 1337 Aquitaine, Kutsal Roma 1311 İtalya, Memlük 1320 Bağdat-Musul,
  Rusya 1478 Novgorod-Kırım, Venedik 1340 Doğu Akdeniz, Aragon 1416 Napoli, Portekiz
  1415 Fas ve Safevî 1501 İran çekirdeği ekseninde plan yapar. Açılışta konsolidasyon
  veya savunma seçildiği ve her uzun hedefin yıl/bölge eşiği regression testinde
  doğrulanır: `Test1300MajorPowersUseHistoricalLongHorizonObjectives`.

- 2026-08-04: 1300 senaryosundaki tüm 72 devlet için AI strateji profili
  tanımlandı. Eksik İber, İtalya, Orta Avrupa, Kafkasya, Balkan ve Britanya
  devletleri kendi savunma çekirdekleri ile sınırlı genişleme hedeflerini aldı;
  başlangıçta elimine Ragusa ile Burgonya da yeniden kurulduğunda kullanacakları
  profil ve geri alma hedefleriyle kapsandı. `Test1300ScenarioAIStrategyReferencesExist`
  artık her fraksiyon için en az bir profile/amaç tanımı zorunluluğu getiriyor.

- 2026-08-04: AI'nin yeni yağma/pusu görevleri amaç ve risk temelli hale getirildi.
  Normal zorlukta mevcut proaktif savaş, ekonomi, araştırma, donanma ve üretim
  planının yanında; AI artık ana fetih hedefini görevle geciktirmiyor, yalnız
  komşu düşman kuvvetinin yaklaşabildiği uygun arazide pusu kuruyor ve gerçek
  yağma ganimeti yüksek düşman bölgelerini yağmalıyor. Belirgin karşı taarruz
  üstünlüğünde normal güvenlik/geri çekilme akışına dönüyor. Regression:
  `TestAITerritoryTaskCanSetAmbushAndKeepArmyHidden`,
  `TestAITerritoryTaskRaidsValuableRegionWithoutAmbushOpportunity`,
  `TestAITerritoryTaskDoesNotDelayPrimaryConquestTarget`.

- 2026-08-03: 1300 başlangıç senaryosunun bina altyapısı tarihsel yerleşim ve devlet
  yapısına göre dolduruldu. Port yerleşimleri en az 1. seviye liman, kale yerleşimleri
  en az 1. seviye sur, devlet başkentleri ise 1. seviye kışla/ambar/ibadethane/pazar
  ile başlıyor. Konstantiniyye, Viyana, Belgrad, Buda, Paris, Kahire ve Rodos 3.
  seviye; Edirne, Niş-Semendire, Vidin, Bursa, İzmit, Sinop ve Bitinya 2. seviye
  surla işaretlendi. Regression: `Test1300ScenarioSettlementInfrastructureHasMinimumBuildings`,
  `Test1300ScenarioHistoricalStrongholdsHaveExpectedWallLevels`; doğrulama:
  `go test ./...`.

- 2026-08-04: Son kara toprağı kuşatılan devlet için teslimiyet yerine bölgeye
  bağlı `propose_siege_vassalization` akışı eklendi. AI aynı baskı eşiğinde
  vassallık teklif ediyor; oyuncu da kuşatan tarafta `Vassallık Teklifi`,
  kuşatılan tarafta `Vassallığı Kabul Et` görüyor. Kabulde bölge sahibi
  değişmeden hedef devlet kuşatanın vassalı oluyor, savaş ve kuşatma bitiyor.
  Regression: `TestAcceptedLastRegionSiegeVassalizationKeepsRegionAndEndsSiege`,
  `TestPlayerCanOfferSiegeVassalizationToLastRegion`,
  `TestAILastRegionSiegeOffersVassalization`; hedefli diplomasi/AI/game/render
  testleri geçti.

- 2026-08-03: Üst oyuncu HUD'unda Gelir/Altın sütunu sağa yaslanarak Kereste/Taş
  değerleriyle örtüşmesi giderildi. Kaynak, gelir, altın, ambar ve askeri güç
  değerleri Türkçe binlik ayıracıyla (`10.000`) gösteriliyor; formatlı uzun
  değerlerin satır içinde taşmaması için ortak KeyValue satırında HUD'a özel
  sıkı aralık kullanılıyor. Regression: `TestFormatNumberTR`,
  `TestTopResourceHUDColumnsKeepIncomeSeparate`; doğrulama:
  `go test ./internal/render ./internal/ui`.

- 2026-08-03: Filo teması beklenirken kamera dikey olarak yeniden çerçeveleniyor;
  gerçek deniz anchor'ı modalın altında alt-orta alana taşınıyor. Böylece sahte
  ikinci bir marker yerine temas eden gerçek filo ikonları, komutan portreleri
  ve temas halkası görünür kalıyor. Temas kapandığında önceki kamera konumu
  geri yükleniyor. Regression: `TestNavalContactCameraTargetLeavesMapBelowModal`;
  doğrulama: `go test ./internal/render -run 'TestNavalContact' -count=1`.

- 2026-08-03: Filo temas kartlarında `Birim`, `Saldırı / Savunma`, `Moral`,
  `Hareket` ve `Görev / Komutan` değerleri ayrıştırıldı. Toplam `GÜÇ`, kartın
  en altında ayraçlı ve daha büyük vurgu alanında gösteriliyor.

- 2026-08-03: Açık deniz filo temas modalı, genel karar akışını koruyarak iki
  filonun devlet, birim/güç, savunma/moral, hareket hakkı, görev ve komutan
  durumlarını karşılaştırmalı kartlarda gösteriyor. Modal üst HUD/`HAMLELER`
  panelinin altına taşındı; temas denizi haritada seçili kalıyor ve iki filo
  marker'ı temas halkasıyla vurgulanıyor. Geri çekil düğmesinin mevcut hareket
  hakkı koşulu korunuyor. Regression: `TestNavalContactDialogShowsFleetComparisonAndUpperPlacement`,
  `TestNavalContactWithdrawButtonCanBeDisabled`; doğrulama: `go test ./...`.

- 2026-08-03: Temas sonrası düşman bölgesinde `Pozisyonu Koru` seçen oyuncu
  ordusu artık aynı bölgeye sağ tıklayarak görev alabiliyor. Tahkimli hedefte
  mevcut `Kuşatma Kararı` açılıyor; tahkimatsız ve savunma ordusu olmayan
  hedefte görev modalından `Ele Geçir` seçilince bölge doğrudan oyuncu toprağına
  katılıyor. Düşman ordusu varsa `Kara Muharebesi` planı açılıyor. Görevler için
  genel onay penceresinden ayrılmış dört seçenekli modal kullanılıyor; `Vazgeç`
  ile aksiyon almadan kapanıyor ve buton ikonları göreve göre anlamlandırılıyor.
  Aynı-bölge görev göstergesi seçili ordu üzerinde altın görev rozetiyle
  çiziliyor; bölge üzerinde cursor parmağa dönüyor. Pusu ve aktif yağma için
  markerın sağ-üstüne görev rozetleri eklendi; hover tooltip'i pusu etkisini
  veya yağmanın gerçek altın/kaynak kazancını gösteriyor. Rozet taşıyan yan yana
  kara marker gruplarında aralık rozetlerin çakışmaması için genişletildi.
  Aynı görev menüsüne `Yağmala` ve `Pusu` da eklendi. Yağma bölge başına turda
  bir kez uygulanıp ekonomi tick'inde verginin %80'ini ve üretimin %50'sini
  yağmalayan devlete aktarır. Pusu ordusu düşman görünürlüğünden çıkarılır;
  hedefe girişte özel kara teması açılır ve hareketli tarafın geri çekilmesi
  ile `Pozisyonu Koru` seçimi kapatılır. AI aynı görev ve görünürlük kurallarını
  kullanır. Regression:
  `TestCurrentRegionTaskOpensSiegeForFortifiedHeldEnemyRegion`,
  `TestCurrentRegionTaskOffersDirectCaptureForUnfortifiedHeldEnemyRegion`,
  `TestCurrentRegionTaskOpensBattlePlanWhenUnfortifiedRegionHasEnemyArmy`,
  `TestCaptureUnfortifiedRegionDirectlyAnnexesHeldEnemyRegion`; doğrulama:
  `go test ./internal/game ./internal/render ./internal/state ./internal/ai -count=1`.

- 2026-08-03: Senaryo yüklenirken gösterilen yükleme ekranına da seçilen senaryonun
  `scenario_bg.png` arka planı eklendi. Yeni senaryo yolu yükleme süresince renderer
  state'inde taşınıyor; asset yoksa koyu fallback korunuyor. Regression:
  `TestLoadingBackgroundLoadsFromScenarioDirectory`; doğrulama:
  `go test ./internal/render -run TestLoadingBackgroundLoadsFromScenarioDirectory -count=1`.

- 2026-08-03: Yükleme ekranındaki spinner, mesaj, ilerleme çubuğu, yüzde ve
  bekleme metni ekranın alt-orta bölümüne taşındı; üst dekoratif çizgiler ve
  yükleme ilerleme mantığı korunuyor.

- 2026-08-03: Fraksiyon seçim kartlarına, devlet başlığının altında sağ tarafa
  senaryoya ait bayraklar eklendi. Kartın tıklanabilir alanı korunurken metin
  genişliği bayrak alanına göre kırpılıyor; eksik bayraklarda mevcut baş harf
  fallback'i kullanılıyor. Regression: `TestFactionCardFlagRectStaysUnderTitleInsideCard`;
  doğrulama: `go test ./internal/render -run 'Test(FactionCardFlagRect|FactionSelectBackground|SelectionScreensRenderSmoke)' -count=1`.

- 2026-08-03: Ayarlar ekranına `Tam Ekran`/`Pencereli` ekran modu eklendi. Seçim
  anında uygulanıyor, `saves/settings.json` içine kaydediliyor ve sonraki açılışta
  yükleniyor; F11 kısayolu da aynı ayarla senkron kalıyor. Doğrulama:
  `go test ./internal/render ./internal/game`.

- 2026-08-03: Yeni Oyun → Senaryo Seç akışından sonra açılan oynanabilir devlet
  seçim ekranı, seçilen senaryo dizinindeki `scenario_bg.png` görselini arka plan
  olarak kullanıyor. Görsel cover ölçekleniyor, okunabilirlik için hafif overlay
  uygulanıyor ve asset yoksa koyu fallback korunuyor. Regression:
  `TestFactionSelectBackgroundLoadsFromScenarioDirectory`; doğrulama:
  `go test ./internal/render -run 'Test(FactionSelectBackground|SelectionScreensRenderSmoke)' -count=1`.

- 2026-08-03: Ana menüye `assets/images/main_menu_bg.png` arka planı eklendi.
  Görsel bir kez cache'leniyor, farklı ekran oranlarında cover ölçekleme ile
  menünün arkasında kalıyor; canlı renkler korunurken menü ekseninde siyah,
  kenarlara doğru saydamlaşan focus gradient'i kullanılıyor. Aynı görsel senaryo
  seçim ekranında da kartların arkasında hafif overlay ile kullanılıyor. Yükleme
  başarısız olursa koyu fallback korunuyor.
  Doğrulama: `go test ./internal/render -run 'Test(MainMenu|SelectionScreensRenderSmoke)' -count=1`.

- 2026-08-03: 1300 AI'nin aktif savaşta eksik üretim kaynağı nedeniyle üretimi
  tamamen durdurması düzeltildi. `aiProcureStrategicResources()` seçilen birim,
  bina, nakliye veya savaş gemisinin `ResourceCost` maliyetinden tüm eksik tahıl,
  demir, kereste, taş, baharat ve kumaşı otomatik çıkarıp aktif ticaret ağından
  satın alıyor; yeterli altın yoksa yalnız karşılanabilen kısmı alıyor. Abluka
  altındaki limanlar somut çıkarma görevi olmasa da `%110` deniz savunma eşiğine
  kadar savaş gemisi kuyruğu açıyor ve deniz tehdidinde ilgili araştırmalar öne
  alınıyor. Regression: `TestAIProcuresEveryMissingProductionResource`,
  `TestAIProcuresMilitaryIronFromConnectedTradeNetwork`,
  `Test1300ThreatenedPortQueuesNavalDefenseWithoutMission`; doğrulama:
  `go test ./internal/ai`.

- 2026-08-03: Oyuncuya gelen barış teklifi modalına kabul sonrası ateşkes uyarısı
  eklendi. Modal, barış kabul edilirse altı tur boyunca aynı devlete yeniden savaş
  ilan edilemeyeceğini ortak `PostPeaceTruceTurns` sabitinden okuyarak gösteriyor;
  diğer diplomasi teklifleri etkilenmiyor. Regression: `TestDiplomacyOfferTruceNoticeTR`;
  doğrulama: `go test ./internal/render ./internal/diplomacy ./internal/state`.

- 2026-08-03: Ordu veya filo yıpranmış konumdan farklı bir bölgeye taşındığında eski `ArmyLogisticsStatus` temizleniyor; kara ordusunun bölgesel `OverCapacityTurns` sayacı hedefte yeniden başlıyor, filonun `TurnsWithoutPort` yolculuk sayacı korunuyor. Böylece marker üzerindeki `!` eski konumun yıpranmasını taşımıyor. Regression: `TestMovingArmyClearsPreviousRegionLogisticsWarning`, `TestMovingFleetClearsPreviousSeaAttritionWarning`; doğrulama: hedefli paket testleri ve `go test ./...`.

- 2026-08-03: Savaş ilanı sonrası akış sıralaması düzeltildi. `Savaş Özeti`
  artık ilan uygulandığı anda modal olarak öne çıkıyor; bekleyen ordu hareketi
  ve savaş açılış teması özet kapanana kadar erteleniyor, ardından normal akış
  devam ediyor. Regression: `TestWarDeclarationShowsSummaryBeforeQueuedMovementContinues`;
  doğrulama: `go test ./internal/game ./internal/render`.

- 2026-08-03: Savunucu ordusu bulunmayan tahkimli rakip bölgelerde de savaş ilanı
  uygulandığı anda `Savaş Özeti` gösteriliyor. `Kuşatma Kararı` artık özeti
  örtmeden, özet kapandıktan sonra açılıyor. Regression:
  `TestWarDeclarationWithoutDefenderShowsSummaryBeforeSiegeDecision`,
  `TestFinalizeWarConfirmDefersSiegeDecisionUntilWarIsDeclared`.

- 2026-08-03: Kuşatma kılıç rozeti hizası düzeltildi. Kuşatan ve kuşatılan
  ordular, merkez ve kale yerleşim anchor'ları farklı olduğunda bile aynı kale
  marker grubunda çiziliyor; böylece kılıç rozeti bir kuşatmada marker'ın sağına,
  diğerinde merkezine kaymıyor. Regression: `TestArmyIconPositionsKeepSiegePairOnFortressAnchor`;
  doğrulama: `go test ./internal/render -run 'TestArmyIconPositions|TestArmySiegeBadge' -count=1`.

- 2026-08-02: Ordu bilgi paneli arkada kaldığında askeri birim kartı hover
  popup'ının görünmeye devam etmesi düzeltildi. Popup yalnız ordu paneli üst
  etkileşim katmanındaysa çiziliyor; diplomasi, Merchant Route, kuşatma ve
  modal/overlay panelleri açıkken arka panelin hover'ı bastırılıyor. Yalnız
  kendi yüzeyinde input tüketen Aktif Savaşlar paneli açıkken harita ve ordu
  paneli aktif kalıyor.
  Regression: `TestArmyPanelTooltipOnlyActiveOnTopLayer`; doğrulama:
  `go test ./internal/render/...`.

- 2026-08-02: Ticaret Haritasında pozitif hacim bonusu üreten merchant donanmaları
  artık kendi yuvarlak marker'ı ve sarı `+N` rozetiyle görünür. Her marker,
  atandığı `TradeRouteKey` koridorunun en yakın bezier noktasına ince altın/cyan
  connector ile bağlanır; koridoru görünmeyen uzak zoom rotalarında bağlantısız
  marker çizilmez. Rozet hover'ı normal haritadaki aynı ticaret bonusu popup'ını
  gösterir. Regression: `TestTradeBonusFleetVisualsFilterToActiveAssignedFleets`,
  `TestTradeBonusFleetConnectsToItsAssignedCorridor`; doğrulama:
  `go test ./...`.

- 2026-08-02: Kara ordusu temas akışı donanma temasına paralel hale getirildi.
  Düşman kara ordusuna verilen hareket emri artık doğrudan savaşa girmeden önce
  `Düşman Ordusu Tespit Edildi` modalında `Çatış / Geri Çekil / Pozisyonu Koru`
  seçeneklerini sunuyor. İki taraf da `Çatış` seçerse mevcut kara muharebesi
  planı ve seçilen saldırı duruşu korunuyor; savunmacı geri çekilirse güvenli
  dost/boş komşu kara bölgesine taşınıyor. AI `%135` güç farkı ve güvenli rota
  varsa geri çekiliyor. Tahkimli hedeflerde temas sonrası mevcut kuşatma akışı
  korunuyor. Regression:
  `TestPlayerLandMovementCreatesContactBeforeBattle`,
  `TestPlayerLandContactClashResolvesBattle`,
  `TestPlayerLandDefenderCanWithdrawFromContact`,
  `TestLandContactOutmatchedAIWithdraws`,
  `TestLandContactClashOpensBattlePlan`; doğrulama:
  `go test ./internal/...`.

- 2026-08-02: Tahkimli kara hedefleri de kuşatma başlatmadan önce kara temasına
  alındı. Temas popup'ı açılırken saldıran ordu hedef bölgede görünür ve hareket
  puanını tüketir; `Çatış` seçilince açık arazideki muharebe planı yerine
  tahkimli hedefte mevcut `Kuşatma Kararı` modalı açılıyor; `Kuşatma Başlat` veya
  `Genel Hücum` seçimi normal kuşatma çözümüne devrediliyor. Aktif kuşatma destek/
  huruç akışı korunuyor. Regression:
  `TestPlayerLandMovementCreatesContactBeforeFortifiedSiege`,
  `TestExecuteMoveCreatesContactBeforeSiegeOnFortifiedTargetWithDefender`,
  `TestFortifiedLandContactClashOpensSiegeDecision`; doğrulama:
  `go test ./...`.

- 2026-08-29: Tahkimli bölgede düşman ordusuyla kara teması çatışmaya
  dönüştüğünde kuşatma kararının araya girmesi düzeltildi. `Çatış` seçimi artık
  tahkimat seviyesinden bağımsız olarak `ShowLandContactBattlePlan` kara
  muharebesi planına gider; kuşatma kararı yalnızca hedefte düşman ordusu
  bulunmayan normal hareket akışında açılır. Regression:
  `TestFortifiedLandContactClashOpensBattlePlan`; doğrulama:
  `go test ./internal/game ./internal/render -count=1`.

- 2026-08-02: Görevi atanmış ancak görev bölgesine ulaşmamış oyuncu filolarında
  sağ-üstte siyah borderlı gri dairesel bekleyen-görev rozeti gösteriliyor.
  Hedefe ulaşınca mevcut görev bonus rozeti korunuyor; çizim, cursor ve seçim
  aynı rozet geometry'sini kullanıyor. Regression:
  `TestNavalMissionPendingBadgeOnlyShowsBeforeMissionRegion`,
  `TestNavalMissionPendingBadgeSharesMissionBadgeGeometry`; doğrulama:
  `go test ./internal/render`.

- 2026-08-02: Escort görevi atanmış savaş filosu, koruduğu nakliye filosu başka
  bir açık deniz bölgesine başarıyla taşındığında kendi hareket puanı yeterliyse
  aynı denize otomatik olarak takip ediyor. Escort hareketi normal filo hareketi
  doğrulamasını ve kendi hareket puanı tüketimini kullanıyor; puanı olmayan escort
  bulunduğu denizde kalıyor. Regression: `TestMovingEscortedFleetMovesEscortWithAvailableMovement`,
  `TestMovingEscortedFleetLeavesEscortWhenMovementIsInsufficient`; doğrulama:
  `go test ./internal/game`.

- 2026-08-02: Abluka veya devriye görevli filo farklı bir denize, limana ya da
  temas geri çekilmesiyle başka bir konuma taşındığında görev otomatik temizleniyor.
  Merkezi hareket çözümü ve deniz teması geri çekilmesi aynı state helper'ını
  kullanıyor. Regression: `TestMovingPatrolOrBlockadeFleetClearsMission`.

- 2026-08-02: Escort görev satırı artık hedef nakliye filosunun dahili ID'sini
  göstermiyor. Yalnızca savaş filosuyla aynı açık denizde bulunan uygun nakliye
  filoları `Escort` seçeneği olarak listeleniyor; state katmanı da farklı deniz
  veya liman konumundaki hedefleri reddediyor. Regression:
  `TestEscortRequiresSameOpenSeaAsTransport`,
  `TestNavalMissionOptionsOnlyShowSameSeaTransportEscort`.

- 2026-08-02: Devriye ve abluka görevleri artık hedef bölge seçimi istemiyor;
  görev panelinde yalnızca filonun bulunduğu açık denizde atanabiliyor ve doğrudan
  mevcut deniz bölgesini hedefliyor. Abluka seçeneği yalnız savaş halindeki düşman
  kıyısına komşu denizde görünür; limandaki filoda bu görevler gizlenir. Nakliye
  hedef daireleri korunarak bu dairelerin üzerinde cursor pointer olur. Regression:
  `TestPatrolAndBlockadeMustUseCurrentOpenSea`,
  `TestNavalMissionOptionsHideBlockadeOutsideValidCurrentSea`,
  `TestNavalMissionTargetCircleUsesPointerHitRadius`.

- 2026-08-02: Abluka altındaki düşman kıyı şeridi overlay'i koyu gri yerine
  koyu kırmızı olarak güncellendi; abluka yüzdesi ve kıyı kaplama geometrisi
  değişmedi. Regression: `TestSelectedMapBorderUsesThreePixelStroke`.

- 2026-08-02: Abluka görevi yalnızca savaş halindeki düşmanın kıyı kara
  bölgelerine komşu denizlerde atanabilir hale getirildi. Okyanus ortası hedefler
  state doğrulamasında reddediliyor; hedef seçimindeki merkez yuvarlakları da
  aynı kuralla filtreleniyor ve üzerlerinde cursor parmağa dönüşüyor. Regression:
  `TestNavalBlockadeRequiresEnemyCoastalSea`,
  `TestNavalMissionTargetCircleUsesPointerHitRadius`.

- 2026-08-02: Denizden çıkarma ile başlatılan düşman kale kuşatması barışla sona
  erdiğinde kuşatan ordu artık düşman kıyısında bırakılmıyor. `NavalLanding`
  işaretli kuşatmalar ortak barış akışında hedefe en yakın yeterli nakliye
  filolarına geri yükleniyor; nakliye yoksa en yakın kendi kara bölgesine
  çekiliyor. Oyuncu ve AI çıkarma yolları aynı state çözümünü kullanıyor.
  Regression: `TestEvacuateNavalLandingSiegeReembarksAtNearestTransportFleet`,
  `TestEvacuateNavalLandingSiegeRetreatsToNearestOwnedRegionWithoutFleet`,
  `TestAcceptedPeaceEvacuatesNavalLandingSiege`; doğrulama:
  `go test ./internal/state ./internal/diplomacy ./internal/game ./internal/ai`.

- 2026-08-02: Abluka filosu bonus rozeti hover edildiğinde `%5/%10` ganimet
  oranının hedef kıyı üretimindeki altın karşılığı `Gelir katkısı (ganimet):
  +N altın/tur` biçiminde gösteriliyor. Tooltip, tek filo katkısını state
  katmanındaki `BlockadeLootGoldForFleet()` hesabından alıyor. Regression:
  `TestNavalMissionBonusTooltipIncludesTargetAndEffect`,
  `TestRegionProductionAndBlockadeLootFollowRetentionRates`; doğrulama:
  `go test ./...`.

- 2026-08-02: AI müttefikinin, oyuncunun zaten savaşta olduğu hedefe açtığı savaş
  için gereksiz `Savaşa Katılım Çağrısı` modalı göstermesi düzeltildi. Savaş
  çağrısı kuyruğu realm kökleri üzerinden mevcut savaşı kontrol ediyor; eski
  geçersiz çağrılar da modal seçimine alınmıyor. Regression:
  `TestExecuteWarDeclarationDoesNotQueuePlayerAlreadyAtWarWithTarget`,
  `TestBestOfferIndexSkipsWarJoinOfferForExistingWar`; doğrulama:
  `go test ./internal/diplomacy`.

- 2026-08-02: Rotaya atanmış ticaret filolarının sarı `+N` rozeti görev
  rozetleriyle aynı sağ-üst konuma taşındı. Rozet hover edildiğinde rota hacim
  bonusu ve abluka sonrası tur başına altın katkısı tooltip'te gösteriliyor;
  çizim, cursor ve hit-test ortak rozet geometry'sini kullanıyor. Regression:
  `TestNavalArmyBadgesShareUpperRightAnchor`,
  `TestMerchantTradeBonusForArmyOnlyShowsActiveTargetSeaBonus`; doğrulama:
  `go test ./...`.

- 2026-08-02: Deniz temasındaki `Geri Çekil` seçeneği, oyuncu filosunun hareket
  puanı yoksa disabled hale getirildi. Renderer butonu pasif çiziyor; state
  katmanı da geçersiz geri çekilme kararını reddediyor. Regression:
  `TestNavalContactWithdrawRequiresMovementPoint`,
  `TestNavalContactWithdrawButtonCanBeDisabled`.

- 2026-08-02: Deniz temasında AI filo gücü ve hareket puanına göre karar veriyor.
  Güçleri yakın görevsiz AI filosu `Çatış` seçiyor; karşı taraf `%125` veya daha
  güçlüyse ve hareket hakkı varsa `Geri Çekil` seçip komşu denize dönüyor.
  Abluka filosunun normal `Pozisyonu Koru` varsayılanı, güçlü düşman karşısında
  geri çekilmeyi engellemiyor. Regression:
  `TestUnassignedEnemyContactWithdrawsWhenOutmatched`,
  `TestUnassignedEnemyContactClashesWhenComparable`.

- 2026-08-02: Hareket puanı olmayan AI filosu geri çekilme kararı veremiyor;
  AI-AI temasında seçilen geri çekilme gerçek deniz hareketine ve hareket puanı
  düşümüne bağlandı. Geri çekilme maliyeti 2 hareket puanı olarak sabitlendi;
  kalan puan daha azsa sıfırlanıyor. Regression:
  `TestAIContactRetreatMovesWeakFleetBackAfterEntry`,
  `TestAIContactCannotChooseWithdrawWithoutMovementPoint`,
  `TestAIContactRetreatPrefersUnoccupiedSea`.

- 2026-08-02: Oyuncu filosu düşman filosunun bulunduğu denize giderken temas
  modalı artık hareketten önce açılmıyor. Filo önce hedef denize taşınıyor ve
  hareket puanını harcıyor; ardından `Düşman Filo Tespit Edildi` modalı açılıyor.
  Temas sonrası çatışma planı bu harcama bilgisini koruyor.

- 2026-08-02: Filo zaten düşman filosuyla aynı denizdeyken yeni `Devriye` veya
  `Abluka` görevi atanması temas akışına bağlandı. Aynı görev tekrar atanırsa
  yeni modal açılmıyor; görev değişirse temas yeniden başlıyor. Regression:
  `TestAssigningPatrolOrBlockadeInOccupiedSeaCreatesContactOnce`.

- 2026-08-02: Deniz temasında `Geri Çekil` seçimi artık filoyu gerçek bir deniz
  komşusuna taşıyor ve 2 hareket puanı harcıyor. Hareket sırasında hedef denize
  giriş puanı zaten harcandıysa kalan puan sıfıra kadar düşürülüyor; rota
  düşmansız deniz komşusunu tercih ediyor ve düşmanın geldiği kaynak denizi
  dışarıda bırakıyor. Güvenli hedef yoksa geri çekilme pasif kalıyor. Regression:
  `TestNavalContactWithdrawPrefersUnoccupiedSea`.
  Regression: `TestNavalContactWithdrawMovesFleetToAnotherSea`.

- 2026-08-02: Düşman denizine hareket sırasında `Deniz Muharebesi Planı` artık
  temas modalından önce açılmıyor. Filo önce hedef denize hareket ediyor, iki
  taraf da `Çatış` seçerse savaş duruşu planı açılıyor. Regression:
  `TestNavalContactClashOpensBattlePlan`.

- 2026-08-02: Port binası tamamlandığında oluşturulan otomatik `Liman`
  settlement'ı artık ülke shape'inin rastgele kenarına değil, ilgili kara
  bölgesi ile komşu deniz bölgesinin gerçek raster kıyı sınırına yerleşiyor.
  Aynı `ShapeID`'yi paylaşan alt bölgelerde uzak/dağlık ülke kenarı fallback'te
  kullanılmıyor. Regression: `TestCoastalSettlementPointUsesActualLandSeaRasterBoundary`,
  `TestAutoPortSettlementPointRejectsSharedCountryShape`,
  `TestCompleteBuildingPortCreatesPortSettlement`; doğrulama:
  `go test ./internal/game ./internal/render ./internal/world`.

- 2026-08-02: Bölge bilgi panelindeki tamamlanmış ve kuyruğu boş bina kartlarına
  sağ üstte kırmızı `X` yıkım düğmesi eklendi. Tıklama ortak confirm modalını açıyor;
  onay sonrası oyuncunun kendi bölgesinden bir bina seviyesi kaldırılıyor, kaynak
  iadesi yapılmıyor ve portun otomatik yerleşim temizliği korunuyor. Regression:
  `TestBuildingDemolishButtonIsAtCardTopRightAndHitTestable`,
  `TestBuildingDemolitionOpensConfirmationDialog`,
  `TestDemolishBuildingRemovesOneCompletedLevelWithoutRefund`; doğrulama:
  `go test ./internal/render ./internal/game`.

- 2026-08-02: Deniz temas olayı eklendi. Aynı denize giren düşman filoları veya
  savaş açılışında zaten aynı denizde bulunan filolar önce `Düşman Filo Tespit
  Edildi` modalında karşılıklı karar verir. `Çatış` yalnız iki taraf da seçerse
  savaşı başlatır; `Geri Çekil` ve `Pozisyonu Koru` savaşsız temas çözümü üretir.
  Devriye varsayılan olarak çatışır, abluka ve görevsiz filolar pozisyonunu korur.
  Kapsam: `internal/state/naval_contact.go`, `internal/game/naval_contact.go`,
  `internal/render/action.go`, `internal/ai/ai.go`; regression:
  `TestNavalContactBattlesOnlyWhenBothClash`, `TestNavalContactMissionDefaults`,
  `TestQueueNavalContactForWar`.

- 2026-08-02: Devlet bilgi panelinin `Kaynaklar` bölümüne seçilen devletin
  tur başı brüt altın geliri `Gelir +N/tur` satırı olarak eklendi. Oyuncu HUD'u
  ve devlet paneli aynı `victory.GoldIncomeForFaction()` hesabını kullanıyor;
  gelir mevcut bölge üretimi, ticaret rotaları, abluka etkisi, ganimet ve
  teknoloji bonuslarıyla birlikte gösteriliyor. Regression: `TestCurrentGoldIncomeIncludesRegionsTradeAndTech`;
  doğrulama: `go test ./internal/victory ./internal/render`.

- 2026-08-02: Deniz filo görevleri ile doğrudan saldırı ayrıştırıldı. Görevsiz
  filolar aynı denizde savaşmıyor; `Saldır` emri savaş planı onayından sonra
  doğrudan deniz savaşı başlatıyor. `Devriye` yalnız görevli düşman `Abluka`
  filosunu otomatik yakalıyor, `Abluka` sadece ekonomik baskı kuruyor ve
  `Escort` yalnız bağlı nakliye filosunu koruyor. Eski görevsiz abluka ve
  görevsiz deniz çarpışması kaldırıldı. Regression:
  `TestUnassignedNavalFleetsCanShareSeaWithoutBattle`,
  `TestPatrolAutomaticallyEngagesEnemyBlockade`,
  `TestNavalFleetsAutoEngageOnlyPatrolAndBlockadePair`,
  `TestUnassignedFleetDoesNotCreateBlockade`.

- 2026-08-01: Abluka ekonomisi uygulandı. Limanlı bölgede `%50` abluka vergi,
  yerel ticaret ve bölgesel mal üretimini `%75` seviyesinde bırakıyor; `%100`
  abluka `%50` seviyesine indiriyor. Ablukacı etkili katkısına göre `%50` için
  `%5`, `%100` için `%10` altın ve mal loot'u alıyor. Bölge paneli anında
  `sonraki tur yerel -%X` bilgisini, üst HUD ise azaltılmış üretim/geliri gösteriyor.
  Kapsam: `internal/{state/{state.go,blockade_economy.go},game/resolution.go,victory/victory.go,render/panel.go}`;
  regression: `TestRegionBlockadeEconomicEffectUsesApprovedRetentionAndLootRates`,
  `TestRegionProductionAndBlockadeLootFollowRetentionRates`,
  `TestApplyEconomyTickAppliesBlockadeOutputAndBlockaderLoot`.

- 2026-08-01: Abluka altındaki liman bölgelerinin kara-deniz kıyı sınırı kalın
  koyu gri overlay ile gösteriliyor. `RegionBlockadePercent()` sonucu toplam kıyı
  uzunluğuna uygulanıyor; `%50` yarıyı, `%100` tamamını boyuyor. Cache imzası
  abluka yüzdesini izliyor. Regression: `TestBlockadeCoastlineUsesBlockadePercentOfTotalCoast`;
  doğrulama: `go test ./internal/render ./internal/state -count=1`.

- 2026-08-01: Kendi, vassal ve müttefik limanlı kara bölgelerine komşu deniz
  hücreleri donanmalar için güvenli kıyı sayılıyor. Bu hücrelerde kış gemi
  yıpranması ve embarked sefer yıpranması uygulanmıyor; limansız dost kıyılarda
  normal yıpranma sürüyor. Güvenli bölgeden çıkışta eski limansızlık sayacı
  taşınmıyor. Ortak kapı
  `GameState.CanFleetAvoidSeaAttrition()` ile iki çözümleme akışında kullanılıyor.
  Regression: `TestFriendlyCoastalSeaPreventsNavalAttrition`; doğrulama:
  `go test ./internal/game ./internal/state -run 'Test(FriendlyCoastalSeaPreventsNavalAttrition|ApplySeasonEffectsProtectsActiveMerchantShipsFromWinterAttrition|ApplyEmbarkedVoyageAttrition)' -count=1`.

- 2026-08-01: Bölge panelinde ardıl devlet elenmişken `Tahıl Yardımı` ve
  `Özgürleştir` düğmeleri aynı aksiyon bandında ayrıştırıldı; tahıl yardımı sola,
  özgürleştirme sağa hizalandı. Özgürleştirme hover'ı ardıl devlet adını gösterir.
  Regression: `TestOwnRegionActionButtonsStaySeparatedForLiberation`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-08-01: Escort bonus rozeti daireden düz üst kenarlı, altı eğri birleşen
  kalkan siluetine dönüştürüldü; yeşil-haki dolgu ve mevcut 20 px rozet alanı
  korundu. Doğrulama: `go test ./internal/render -run 'TestNavalMission.*(Badge|Color)|TestNavalMissionBonus' -count=1`.

- 2026-08-01: Escort görev bonus rozeti sarıdan yeşil-haki renge çevrildi;
  devriye, abluka ve ticaret bonus renkleri korunuyor. Regression:
  `TestNavalMissionEscortBadgeUsesKhakiGreen`; doğrulama:
  `go test ./internal/render -run 'TestNavalMission.*Badge' -count=1`.

- 2026-08-01: Ordu taşıyan donanmanın kara hedefleri settlement bazında ayrıştırıldı.
  Merkez yerleşim kare çıkarma border'ı, limanlar koyu mavi yuvarlak docking
  hedefi olarak aynı anda gösteriliyor; liman tıklaması `TargetSettlementID`
  ile dock state'ini koruyor, merkez tıklaması mevcut çıkarma action'ını
  kullanıyor. Regression: `TestNavalLandMoveTargetSettlementShowsPortsAndLandingCenters`,
  `TestMoveArmyExplicitPortTargetKeepsEmbarkedUnitsDocked`; doğrulama:
  `go test ./internal/game ./internal/render -count=1`.

- 2026-08-01: Donanma görev ikonları sadeleştirildi. Devriye/escort görev harfi
  kareleri kaldırıldı; hedefteki `+N`/yüzde bonus dairesi taşınan ordu rozetiyle
  aynı üst-sağ anchor'a hizalandı ve açık renkli dış border kaldırıldı.
  Regression: `TestNavalArmyBadgesShareUpperRightAnchor`,
  `TestNavalMissionBonusBadgeTextContrast`;
  doğrulama: `go test ./internal/render -run 'TestNavalMission|TestNavalArmyBadges|TestArmyCommanderBadge' -count=1`.

- 2026-08-01: Mavi görev ve sarı ticaret bonus rozetleri tüm donanma marker'ları
  ve komutan portrelerinden sonra ön-plan geçişinde çiziliyor; komşu marker veya
  portreler rozetleri kapatamıyor. Kara ordusu ve donanmadaki zayiat `!` rozeti
  ortak sol-üst konuma taşındı. Regression: `TestNavalDamageBadgeUsesUpperLeftAnchor`,
  `TestArmyDamageBadgeUsesUpperLeftAnchor`;
  doğrulama: `go test ./internal/render -run 'TestNavalMission|TestNavalArmyBadges|TestNavalDamageBadge|TestArmyDamageBadge|TestArmyCommanderBadge' -count=1`.

- 2026-08-01: Aynı gruptaki yan yana donanma marker'ları 26 px yerine 29 px
  merkez aralığıyla diziliyor; 26 px marker çapı korunurken aralarında 3 px
  görsel boşluk bırakılıyor. Kara ordu gruplarının 26 px düzeni değişmedi.
  Regression: `TestArmyIconPositionsLeaveThreePixelsBetweenNavalMarkers`;
  doğrulama: `go test ./internal/render -run 'TestArmyIconPositions|TestNavalMission|TestNavalArmyBadges|TestNavalDamageBadge' -count=1`.

- 2026-08-01: Düşman filosunun taşınan ordu rozeti filo istihbarat görünürlüğüne
  bağlandı. Filo sayısı `?` görünüyorsa taşınan birlik sayısı da `?`; tam
  istihbaratta gerçek adet gösteriliyor. Regression:
  `TestNavalEmbarkedArmyBadgeFollowsFleetVisibility`;
  doğrulama: `go test ./internal/render -run 'TestNavalEmbarkedArmyBadge|TestArmyIconPositions|TestNavalMission' -count=1`.

- 2026-08-01: Oyuncunun kendi filosundaki taşınan ordu karesi hover tooltip'ine
  bağlandı. Tooltip `Nakliye Görevi` başlığını ve `Taşınan ordu N birim`
  ayrıntısını gösteriyor; hit-test mevcut taşınan ordu rozetiyle aynı.
  Regression: `TestNavalEmbarkedArmyTooltipText`;
  doğrulama: `go test ./internal/render -run 'TestNavalEmbarkedArmyTooltip|TestNavalMission' -count=1`.

- 2026-08-01: Donanma hareket hedeflerindeki liman settlement işareti açık mavi
  kareden koyu mavi daireye dönüştürüldü; daire, filonun hedef limana dock
  olacağını çıkarma merkezinden görsel olarak ayırıyor.

- 2026-08-01: Bölge seçildiğinde, seçili bölgenin `IsCenter` settlement
  marker'ı sarı bir seçim halkasıyla çevreleniyor. Aynı bölgedeki diğer
  settlement marker'ları ve seçili olmayan bölgelerin merkezleri vurgulanmıyor;
  Edit Mode settlement seçimi de index bazında doğru marker'a sınırlandı.
  Regression: `TestSettlementSelectionOverlayTargetsSelectedRegionCenter`;
  doğrulama: `go test ./internal/render -count=1`.

- 2026-08-01: Seçili donanmanın komşu kara bölgesi hareket hedefleri settlement
  türüne göre ayrıştırıldı. Ordu taşımayan filo yalnız liman settlement'larını,
  `EmbarkedUnits` taşıyan filo liman docking ve merkez çıkarma settlement'larını
  ayrı işaretliyor; bölge merkezinde yanlış hedef halkası çizilmiyor.
  Regression: `TestNavalLandMoveTargetSettlementShowsPortsAndLandingCenters`;
  doğrulama: `go test ./internal/render ./internal/game ./internal/state`.

- 2026-08-01: Oyuncu donanma görevleri mekanik olarak ayrıştırıldı. Abluka
  görevi hedef denizde ticaret rotası ve liman lojistiği kesintisi üretirken,
  Devriye aynı denizdeki düşman abluka gücünü dengeleyip dost ticaret/ikmali
  koruyor. Escort görevine aynı denizdeki nakliye filosu için yüzde 15 savunma
  bonusu eklendi; çoklu escort yüzde 30 ile sınırlandı. Regression:
  `TestPatrolCountersEnemyBlockadeForTradeRoute`,
  `TestPatrolAndBlockadeMissionsHaveDifferentCommerceEffects`,
  `TestPatrolProtectsPortLogisticsFromEnemyBlockade`,
  `TestNavalEscortDefenseBonusAppliesOnlyToSameSeaTransport`.

- 2026-08-01: Donanma görevlerinin etkileri oyuncu UI'ında görünür hale getirildi.
  Görev seçim satırları artık her rolün gerçek bonusunu gösteriyor; hedefe ulaşan
  oyuncu filosu haritada görev türüne özel renkli `HEDEFTE` marker'ı ile
  `DEVRİYE`, `ABLUKA`, `ESCORT` veya `NAKLİYE` etkisini tekrar gösteriyor.
  Regression: `TestNavalMissionPanelThreeLineRowsHaveVerticalClearance`,
  `TestNavalMissionReachedRegionOnlyShowsActiveTarget`;
  doğrulama: `go test ./internal/render -count=1`.

- 2026-08-01: Donanma görev paneli sadeleştirildi. Liste içindeki ayrı görev
  temizleme satırı kaldırıldı; aktif görev için panel footer'ına `Komutanı Ayır`
  stiliyle kırmızı `Görevi Kaldır` düğmesi eklendi. Panel genişliği 600 px'e,
  görev satırları 80 px'e ayarlandı; çizim, cursor ve input aynı layout/button
  geometry'sini kullanıyor. Regression: `TestNavalMissionPanelClearButtonUsesCommanderStylePlacement`;
  doğrulama: `go test ./internal/render -run '^TestNavalMissionPanel|^TestOverlayPanel' -count=1`.

- 2026-08-01: Donanma bonuslarının harita gösterimi sadeleştirildi. Büyük hedef
  bölge marker'ı kaldırıldı; hedefe ulaşan filonun ikonuna küçük dairesel bonus
  rozeti eklendi (`50/100`, `+1`, `15`). Rozet hover'ında ortak tooltip ile hedef
  ve gerçek etki açıklanıyor; badge hit-test'i ordu ikon seçim akışıyla uyumlu.
  Regression: `TestNavalMissionBonusBadgeUsesCompactActiveValues`;
  doğrulama: `go test ./internal/render -count=1`.

- 2026-08-01: Saf nakliye filoları görev UI'ından çıkarıldı; bu filolarda `GÖREV`
  butonu ve görev durumu çizilmiyor, mevcut embark/disembark mekaniği korunuyor.
  Taşınan ordunun merkez sayısı marker içinden kaldırıldı; birlik sayısı sağ
  üstte kare rozetle, komutan portresi doğrudan filo dairesi üzerinde çiziliyor.
  Donanma kara hedefleri settlement hit-test'iyle eşleştirildi: çıkarma merkezi
  dikdörtgen, normal dock limanı koyu mavi daire olarak seçiliyor.
  Regression: `TestNavalEmbarkedArmyBadgeStaysAtUpperRight`,
  `TestNavalLandMoveTargetSettlementUsesPortsUntilLanding`;
  doğrulama: `go test ./internal/render -count=1`.

- 2026-08-01: Donanma görevlerinin harita üzerindeki ayrı `A/D/E` kare rol
  rozetleri kaldırıldı. Bonus daireleri görev etkisini taşımaya devam ediyor;
  açık renkli dış border kullanılmıyor ve metin görev türüne göre kontrast
  alıyor. Regression: `TestNavalMissionBonusBadgeTextContrast`;
  doğrulama: `go test ./internal/render -run '^TestNavalMission' -count=1`.

- 2026-08-01: Taşıma marker'ında filo birim sayısı yuvarlak marker içine geri
  getirildi. Taşınan ordu sayısı karesi merchant bonus rozetiyle aynı düşey
  hizaya alınıp marker'ın sağ üstüne taşındı; komutan portresi marker üzerinde
  kaldı. Regression: `TestNavalEmbarkedArmyBadgeStaysAtUpperRight`;
  doğrulama: `go test ./internal/render -count=1`.

- 2026-08-01: Ordu detay panelindeki komutan kartına, ana komutan mevcut
  olduğunda kırmızı `Komutanı Ayır` düğmesi eklendi. Düğme yalnız oyuncunun
  ordusunda görünür; çizim, hit-test ve cursor aynı ortak `gameui.Button`
  rect'inden türetilir ve mevcut komutan havuza iade akışını kullanır.
  Regression: `TestArmyPanelCommanderUnassignButtonUsesPrimaryCommanderAndOwnArmy`;
  doğrulama: `go test ./internal/render -run
  '^Test(ArmyPanelCommanderUnassignButtonUsesPrimaryCommanderAndOwnArmy|CommanderPanel|ArmyPanelInteractiveHit)' -count=1`.

- 2026-08-01: Merchant rota modalı satırlarına, filonun gitmesi gereken geçerli
  hedef deniz bölgesi mavi etiketle eklendi. Etiket doğrudan
  `MerchantTradeRouteSeaRegions()` state yardımcısından türetiliyor; satır alanına
  göre kırpılıyor. Regression:
  `TestMerchantRouteSeaDisplayNameUsesTargetSeaRegion`,
  `TestMerchantRouteSeaDisplayNameHandlesMissingRouteSea`; doğrulama:
  `go test ./internal/render -run '^TestMerchantRoute' -count=1`.

- 2026-08-01: Merchant rota satırı seçildiğinde satırdaki hedef deniz haritada
  seçili deniz vurgusuyla işaretleniyor. Vurgu panel kapanınca ve başka filo
  seçilince korunuyor; oyuncu başka bir bölge seçtiğinde temizleniyor. Açık
  ordu paneli kapanıyor ve hedef normal bölge seçiminden farklı altın/cyan
  reticle ile gösteriliyor; kamera hedef denizin gerçek raster anchor'ına
  odaklanıyor. Regression:
  `TestMerchantRouteHighlightClearsWhenAnotherRegionIsSelected`,
  `TestMerchantRouteSelectionFocusesTargetAndClosesArmyPanel`;
  doğrulama: `go test ./internal/render -run '^TestMerchantRoute' -count=1`.

- 2026-08-01: Hedef denizde bonus kazanan merchant filolarının yuvarlak harita
  ikonunun sol üstüne bonus miktarını (`+1`/`+2`) gösteren küçük altın rozet
  eklendi. Rozet yalnız `MerchantFleetTradeRouteBonus()` pozitif olduğunda
  çiziliyor. Regression: `TestMerchantTradeBonusForArmyOnlyShowsActiveTargetSeaBonus`;
  doğrulama: `go test ./internal/render -run '^TestMerchant' -count=1`.

- 2026-08-01: Merchant rotaları artık bağlı tüm ticaret merkezi denizlerini
  birleştirmek yerine gerçek yönlü liman çiftinin hedef denizini kullanıyor.
  Gemlik → Özi örneği `black_open_4` / `Karadeniz Açık 4` olarak doğrulandı;
  merchant bonusu, AI hareketi ve abluka aynı tek hedef denizden türetiliyor.
  AI merchant üretimi rakip hedef denizinde değil kendi uç limanında yapıyor.
  Regression: `TestMerchantTradeRouteUsesDestinationPortSeaForGemlikOzi`;
  doğrulama: `go test ./internal/state ./internal/ai ./internal/render`.

- 2026-08-01: Hedef denizde bonus kazanan merchant gemileri kış attrition'ından
  muaf tutuldu; aynı filodaki savaş veya nakliye gemileri normal kış hasarını
  almaya devam ediyor. Hedef denizden uzaktaki veya rotasız merchant gemileri
  muaf değil. Regression:
  `TestApplySeasonEffectsProtectsActiveMerchantShipsFromWinterAttrition`;
  doğrulama: `go test ./internal/game`.

- 2026-08-01: Merchant rota, Donanma Görevi ve Aktif Savaşlar panellerinin
  cursor hover'ı ortak satır/kapatma düğmesi hit-test'lerine bağlandı. Aynı
  paneller için son açılanı öne alan overlay stack'i çizim, input ve cursor
  önceliğini birlikte yönetiyor; panel kapanınca alttaki açık panel görünür.
  Regression: `TestOverlayPanelOrderBringsLastOpenedPanelToFront`,
  `TestOverlayPanelCursorUsesFrontPanelSurface`; doğrulama:
  `go test ./internal/render -run '^Test(OverlayPanel|MerchantRoutePanel|NavalMissionPanel|ActiveWars)' -count=1`.

- 2026-08-01: AI savaş ilanı kararı artık yalnız hedef devletin gücüne bakmıyor.
  Hedefin vassalları, dış müttefikleri ve müttefik vassalları savunma koalisyonu
  gücüne dahil ediliyor; müttefik ordularının hedef kara bölgelerine mesafesi
  katkıyı `%100 / %75 / %50 / %25 / %10` kademeleriyle ağırlıklandırıyor.
  Yakın güçlü koalisyon savaş ilanını engellerken uzak müttefik tam cephe gücü
  sayılmıyor. Regression: `TestAIWarAssessmentIncludesAlliedAndVassalMilitaryPower`,
  `TestAIWarDecisionUsesAlliedDistanceAndCoalitionPower`; doğrulama:
  `go test ./internal/ai ./internal/diplomacy -count=1`.

- 2026-08-01: Savaş açmayı düşünen AI’nin saldıran koalisyon hesabı genişletildi.
  `AssessWarCall().AutoJoin` ile kesin katılacak dış müttefikler ve onların
  vassalları, hedef kara bölgelerine uzaklıklarına göre AI’nin etkin saldırı
  gücüne ekleniyor; oyuncunun bekleyen savaş katılımı kesin destek sayılmıyor.
  Regression: `TestAIWarAssessmentIncludesCertainAttackingAlly`,
  `TestAIWarDecisionUsesCertainAttackingAllyDistance`.

- 2026-08-01: `settlements.json` dışa aktarımı yalnızca kara region kayıtlarını yazacak
  şekilde düzenlendi; deniz region'larının 83 boş settlement satırı temizlendi.
  Edit Mode settlement ekleme ve taşıma yolları deniz region'larını filtreliyor.
  Regression: `TestWriteScenarioSettlementsSkipsSeaRegions`,
  `TestEditModeDoesNotAddSettlementToSeaRegion`; doğrulama:
  `go test ./internal/game ./internal/render -run 'TestWriteScenarioSettlementsSkipsSeaRegions|TestEditModeDoesNotAddSettlementToSeaRegion' -count=1`.

- 2026-07-31: Minimap'teki oyuncu ve düşman ordu/donanma ikonları kaldırıldı;
  minimap dünya görünümü ile kameranın ekrandaki alanını gösteren viewport
  dikdörtgenini göstermeye devam ediyor. Kapsam: `internal/render/panel.go`;
  doğrulama: `go test ./internal/render`.

- 2026-07-31: Aktif kuşatma altındaki kuşatan orduya sonradan gelen destek ordusu,
  aynı settlement anchor'ında kuşatanın soluna hizalanıyor; kuşatma rozeti için
  kuşatan-savunmacı arası özel slot ve mevcut hit-test sırası korunuyor.
  Regression: `TestArmyIconPositionsPutSiegeSupportLeftOfBesieger`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-07-31: Diplomasi teklif panelinde tıklanan teklif türü artık seçili durumunu
  koruyor ve etkin seçili düğme altın-sarı border ile görsel olarak vurgulanıyor.
  `Teklif Gönder` öncesi aksiyon seçimi netleştirildi. Regression:
  `TestDiplomacyActionSelectionUsesHighlightedBorder`; doğrulama:
  `go test ./internal/render -count=1`.

- 2026-07-31: Oyuncu donanma görevlerinin ilk dikey dilimi tamamlandı. Savaş ve
  nakliye filoları için `GÖREV` paneli, devriye/abluka/escort/nakliye atama,
  hedef seçiminde harita işaretleri, görev rozeti, temizleme/yeniden atama ve
  compact save/load desteği eklendi. Tur başında görevli filolar deterministik
  deniz rotasıyla otomatik ilerliyor; nakliye kıyıya ulaştığında mevcut çıkarma
  akışı çalışıyor. Regression: `go test ./internal/game ./internal/state
  ./internal/save ./internal/render`.

- 2026-07-31: Oyuncuyla savaş halindeki realm'lerin açık deniz filoları haritada
  bulundukları deniz bölgesini düşük opaklıklı kırmızı tehdit overlay'i ve kırmızı
  border ile işaretliyor. Docked filolar işaretlenmiyor; filo konumu harita cache
  imzasına bağlandığı için hareket sonrası eski kırmızı alan temizleniyor; deniz
  seçildiğinde kara kıyısındaki kalın sarı `Selected` border önceliğini koruyor.
  Regression:
  `TestEnemyNavalRegionSetMarksOnlyOpenWarFleets`,
  `TestEnemyNavalRegionUsesEnemyBorderStyle`.

- 2026-07-31: AI deniz operasyonları tamamlandı. Limanda üretilen/senaryodan gelen
  donanmalar ilk AI adımında kanonik dock temizliğiyle denize çıkıyor; görevsiz savaş
  gemileri düşman denizi, tehditli liman ve aktif ticaret hattı arasında `Patrol`
  rolüyle hareket ediyor, görevli filolar `Escort` rolünü koruyor. Kara sınırı olmayan
  savaş ilanları somut çıkarma görevi, kıyı hedefi, port, deniz rotası ve transport
  kapasitesi hazır olmadan açılmıyor. F3 AI teşhisine filo/docked/görev/engel satırı
  eklendi. Regression: `Test1300DockedMerchantFleetUndocksBeforeTradeRouteMove`,
  `Test1300WarshipPatrolMovesTowardActiveTradeSea`,
  `Test1300NavalWarReadyRequiresConcreteAssaultMission`,
  `TestAIDiagnosticSnapshotExposesNavalState`; doğrulama:
  `go test ./internal/ai ./internal/render -count=1`.

- 2026-07-31: Vassallık tekliflerinin kabul kapısı sıkılaştırıldı. `Score >= 55` artık
  tek başına yeterli değil; hedef AI doğrudan sınır tehdidi altında olmalı veya teklif
  sahibi en az `5x` askerî güç ve `5x` kara bölgesine sahip olmalı.
  Yönlü tehdit hesabı zayıf hedefin güçlü devlete yanlışlıkla vassal olmasını engelliyor.
  Regression: `TestAssessVassalizationDoesNotAcceptRelationScoreAlone`,
  `TestAssessVassalizationRequiresRegionalSuperiorityWithMilitarySuperiority`,
  `TestAssessVassalizationAcceptsDirectFrontierThreatWithoutRegionalSuperiority`;
  doğrulama: `go test ./internal/diplomacy -count=1`.

- 2026-07-31: Aynı bölgede mevcut ordunun yanına gelen oyuncu veya müttefik
  ordusu artık ikon grubunda mevcut ordunun soluna yerleşiyor. Grup bazlı ilk
  görülme sırası korunurken kuşatma çifti yerleşimi değişmedi. Regression:
  `TestArmyIconPositionsPutNewArrivalLeftOfExistingArmy`; doğrulama:
  `go test ./internal/render -run 'TestArmyIconPositions|Test.*Army.*Icon' -count=1 -v`.

- 2026-07-31: Komutan atama panelindeki boş komutan listesi panel-local viewport'a
  taşındı. Uzun listeler mouse-wheel ve ok tuşlarıyla satır bazında kayıyor;
  `SubImage` clipping'i panel dışına taşmayı engelliyor, scrollbar yalnızca liste
  viewport'a sığmadığında çiziliyor ve görünmeyen satırlar tıklanamıyor. Regression:
  `TestCommanderPanelListViewportAndScrollClamp`,
  `TestCommanderPanelRowHitOnlyUsesVisibleRows`; doğrulama:
  `go test ./internal/render`.

- 2026-07-30: Oyuncu ordusu ardıl devlet metadata'sı dolu bir düşman bölgesini
  ele geçirdiğinde savaş raporu kapanana kadar fetih state'i bekletiliyor; ardından
  `İlhak Et`, `Serbest Bırak` veya `Vassal Yap` seçenekli karar paneli açılıyor.
  Serbest bırakma ardılı bağımsız müttefik, vassal seçeneği doğrudan oyuncu vassalı
  yapıyor; elenmiş ardıllar düşük kaynak ve beş milisle yeniden kuruluyor. Kapsam:
  `internal/game/conquest_decision.go`, `internal/game/game.go`,
  `internal/diplomacy/vassalage.go`, `internal/render/{action.go,renderer_dialogs.go}`;
  regression: `TestSuccessorMetadataQueuesThreeWayDecisionAfterBattle`,
  `TestSuccessorDecisionReleaseTransfersRegionAndDefeatsPreviousOwner`,
  `TestSuccessorDecisionVassalizesEliminatedSuccessor`,
  `TestHideBattleReportShowsQueuedThreeChoiceSuccessorDialog`.

- 2026-07-30: Ana menüde `EDIT MODE` düğmesi `Çıkış`ın altına taşındı ve araya
  standart bir menü satırı boşluğu eklendi. Çizim, fare hit-test'i ve klavye
  cursor'u ortak geometriyi kullanıyor; başlangıç seçimi kayıt durumuna göre
  `Yeni Oyun`/`Devam et` olarak korunuyor. Regression:
  `TestEditModeMenuItemIsOptionalAndBelowExit`; doğrulama:
  `go test ./internal/render`.

- 2026-07-29: Edit Mode mouse sözleşmesi ayrıştırıldı. Settlement konumu sağ
  tık sürüklemeyle, Shape/Bölge boya-sil fırçaları sol tık sürüklemeyle
  çalışıyor; sol tık seçim ve diğer editor aksiyonları için korunuyor.
- 2026-07-29: Edit Mode Shape/Bölge boya-sil işlemleri ertelendi. Mouse bırakılınca
  yalnız geçici boya önizlemesi kalıyor; aktif araç `Uygula` ile hesaplama, harita
  yenileme ve undo snapshot'ını tek seferde çalıştırıyor.
- 2026-07-30: Edit Mode boya/sil araçlarında aktif araç kilidi eklendi. Sınır/Bölge
  boya veya sil araçlarından biri seçilince yalnız seçili düğme yeşil `Uygula`
  olarak kalıyor; diğer üç düğme pasifleşiyor. Uygula sonrası dört araç yeniden
  erişilebilir oluyor. Regression: `TestShapeInspectorToolButtonsLockOtherToolsUntilApply`;
  doğrulama: `go test ./internal/render`.
- 2026-07-30: Edit Mode hassas shape boyama kalıcılığı düzeltildi. `country_shapes.json`
  artık ring noktalarını tam sayıya yuvarlamadan ondalık koordinatla yazıyor; böylece
  ölçekli dünya pikseli sınırları kapatıp açtıktan sonra kaymıyor.

- 2026-07-29: Edit Mode ordu/filo aksiyonu `Bu Devlete Ata` olarak yeniden
  adlandırıldı. Kara orduları için seçili bölge, docked filolar için
  `DockedRegionID` ile gösterilen liman bölgesi sahiplik kaynağıdır; aksiyon
  yalnız ordu sahibi ile bölge sahibi farklıysa aktifleşir. Regression:
  `TestSelectedArmyOwnerAssignmentUsesDockedRegion`.

- 2026-07-29: Edit Mode `Donanma Ekle` pasiflik hatası düzeltildi. Donanma
  uygun kıyı bölgesindeki `port` yerleşimi veya `port` binası üzerinden
  oluşturulabiliyor; filo komşu deniz bölgesine ve tercih edilen liman
  settlement'ına dock ediliyor. Regression:
  `TestEditFleetCanUsePortSettlementWithoutPortBuilding`.

- 2026-07-29: Edit Mode inspector görsel olarak yeniden düzenlendi. `Yerleşim
  Birimi`, `Bölge`, `Devlet`, `Harita` ve `Veri` sekmeleriyle settlement/ordu,
  region, faction ve shape/geçit araçları ayrıştırıldı; ortak
  `Değişiklikleri Kaydet` düğmesi tüm sekmelerde panelin en altında sabitlendi.
  Dinamik sekme/action rect'leri ortak hit-test geometriyle çakışmayacak şekilde
  düzenlendi. Regression: `TestEditInspectorTabsAndSaveAreaStaySeparated`;
  doğrulama: `go test ./internal/render`.

- 2026-07-29: Ardıl devlet/özgürleştirme mekaniği eklendi. Edit mode bölge
  inspector'ında `Ardıl Devlet` dropdown'ı `successor_faction_id` yazar ve undo/redo
  ile korunur. Fethedilen bölgenin ardıl fraksiyonu elenmişse bölge panelinde
  `Özgürleştir` çıkar; aksiyon devleti düşük kaynaklarla, beş milisle ve
  özgürleştiriciyle müttefik olarak normal AI akışına döndürür. 1300 senaryosunda
  başkent sahibi eşleşen 68 bölge işaretlendi. Regression:
  `TestLiberateSuccessorRevivesFactionWithMilitiaAndAlliance`,
  `TestLiberationButtonOnlyAppearsForEliminatedSuccessor`,
  `TestEditModeAssignsSuccessorWithUndoRedo`,
  `Test1300ScenarioCapitalRegionsHaveSuccessorFaction`.

- 2026-07-31: Ardıl devlet metadata'sı olan fetihlerde karar kapısı düzeltildi.
  `GameState.CanRestoreSuccessorAtRegion()` yalnız `is_eliminated=true` ve
  topraksız ardıllar için `İlhak Et / Serbest Bırak / Vassal Yap` panelini açar;
  aktif ardıl devlet bulunan bölge oyuncu ve AI akışlarında doğrudan ilhak edilir.
  Regression: `TestActiveSuccessorMetadataDirectlyAnnexesWithoutDecision`,
  `TestTryResolvePostWarVassalizationRejectsActiveSuccessorMetadata`.

- 2026-07-29: 1300 senaryosunda kara bölgelerinin settlement merkezleri tekilleştirildi;
  Constantine için Annaba merkez olarak tanımlandı. Iberia veri grafiğinde Aragon
  bölgesinin sahibi düzeltildi, Zaragoza Aragon'a taşındı, Almeria Granada'ya ve
  Guadalajara Toledo'ya bağlandı; Catalonia'nın merkezi Barcelona oldu. Regression:
  `Test1300ScenarioLandSettlementCentersAreUnique`,
  `Test1300IberianSettlementOwnershipGraph`; doğrulama: `go test ./...`.

- 2026-07-29: Edit Mode Harita sekmesine `Başkent Yap` aksiyonu eklendi. Seçili settlement'ın sahibi olan fraksiyonun `capital_settlement_id` alanını anında güncelliyor, bekleyen başkent taşımasını temizliyor ve undo/redo ile geri alınabiliyor; mevcut `Ana Yap` yalnızca bölgesel `is_capital` işaretini değiştirmeye devam ediyor. Regression: `TestEditModeSetsSelectedSettlementAsFactionCapital`.

- 2026-08-06: Edit Mode'da bir settlement `Ana Yap` ile bölgesel merkez veya
  `Başkent Yap` ile ulusal başkent seçildiğinde, bölge sahibi boş değilse
  `successor_faction_id` otomatik olarak owner kimliğine eşitleniyor. Settlement
  ekleme, silme ve bölge değiştirme sırasında oluşan yeni merkezler de aynı kuralı
  kullanıyor; successor alanı settlement undo/redo snapshot'ında korunuyor.
  Senaryo bütünlük testi artık her başkentte successor
  zorunluluğu aramıyor; yalnızca mevcut successor referanslarının geçerli
  fraksiyonlara işaret ettiğini denetliyor. Regression:
  `TestEditModeSettlementCapitalAssignsOwnerAsSuccessorWithUndoRedo`,
  `Test1300ScenarioSuccessorFactionReferencesAreOptionalAndValid`.

- 2026-07-29: `1300_ottoman_rise` senaryosunda tüm 68 devletin başlangıç tahıl
  stokları, mevcut başlangıç ordularının canonical `EffectiveArmyGrainUpkeep`
  hesabına göre en az 12 turluk askerî bakımı karşılayacak şekilde güncellendi;
  garrison indirimi ve filolar hesaba dahil. Regression:
  `TestLoad1300StartingGrainCoversTwelveArmyTurns`; doğrulama: hedefli game/state/AI
  testleri geçti; doğrulama sonrasında full suite de geçti.

- 2026-07-29: Edit mode bölge ID değişimi artık komşuluk, geçit, ordu, paint,
  AI objective ve trade-center ID/link referanslarını aynı state işlemi içinde
  taşıyor; undo/redo snapshot'ları da AI/ticaret verisini kapsıyor. Senaryo
  kaydı `ai_strategies.json` ve `trade_centers.json` dosyalarını güncel runtime
  state'ten yeniden yazıyor. 1300 verisinde artık mevcut bölge grafiğinde
  bulunmayan `ragusa` AI hedefi `dalmatia` ile düzeltildi; toprağı olmayan Ragusa
  devleti `is_eliminated` olarak işaretlendi. Koper anchor testi güncel shape
  geometrisini kabul ederken shape dışı anchor fallback'ini koruyor.
  Regression: `TestRenameRegionIDUpdatesEditorReferences`,
  `TestTradeCenterVisualFollowsEditedRegionMetadata`,
  `TestWriteScenarioEditDataWritesAIStrategiesAndTradeCenters`; doğrulama:
  `go test ./...`.

- 2026-07-29: Edit mode başlangıcı ana menüde ayrı bir `EDIT MODE` seçeneğine
  taşındı. Seçenek yalnız `.env` içindeki `EDIT_MODE=true` iken görünür; önce
  senaryo seçimini açıp seçilen senaryoyu normal edit haritasına yükler. `Yeni
  Oyun`, `EDIT_MODE=true` olsa bile normal fraksiyon/zafer seçim akışını korur.
  Regression: `TestEditModeMenuItemIsOptionalAndAboveNewGame`,
  `TestScenarioLoadEditModeIsExplicit`; doğrulama: `go test ./internal/render ./internal/game`.

- 2026-07-29: 1300 senaryosunda tahılsız AI devletleri için aktif ticaret ağı üzerinden
  otomatik tahıl tedariki eklendi; alım üç aylık kapasite hedefi ve iki aylık pencereyle
  sınırlı, tedarikçi güvenli fazlasını koruyor. AI bina yatırımında ilk `granary` ambarı
  öne alındı, üretim bütçesi iki aylık operasyonel tahıl rezervini koruyor ve askerî bütçe
  oranları konsolidasyonda `%35`, genişlemede `%55`, savaş/savunmada `%70` oldu. Boş
  ordu kayıtları load ve tur sonu normalizasyonuyla temizleniyor; birimsiz devletlerin
  ilk kışlayı manpower doluluğu beklemeden kurması sağlandı. Regression: `internal/ai/strategy_regression_test.go`,
  `internal/state/commanders_test.go`; doğrulama: hedefli AI/state/save/game testleri geçti.

- 2026-07-30: Aktif savaş düğmesi müzik HUD'unun sağındaki ayrılmış yardımcı slota taşındı. Aktif savaş paneli olaylar panelinin soluna kaydırılarak iki panelin üst üste binmesi kaldırıldı; yeni slot ve panel sınırı için regression testleri eklendi. Doğrulama: `go test ./internal/render`.
- 2026-07-30: Aktif savaş paneli satırları 570 px genişliğe çıkarıldı. Her satırın iki ucuna 50×50 px faction bayrak alanı, mevcut savaş bilgileri için ortalanmış metin kolonu eklendi; ortak bayrak asset/cache ve baş harf fallback'i kullanılıyor. Regression: `TestActiveWarRowKeepsCenteredTextBetweenFactionFlags`; doğrulama: `go test ./internal/render`.
- 2026-07-30: Aktif savaş HUD düğmesi 36×36 px’e büyütüldü, üstten 5 px margin aldı ve hover durumunda parmak imleci gösterecek şekilde ortak in-game cursor hit-test’ine bağlandı. Regression: `TestActiveWarsHUDButtonUsesUtilitySlotAfterMusic`; doğrulama: `go test ./internal/render`.
- 2026-07-30: Aktif savaş paneli kapalıyken eski panel alanının harita tıklamalarını engellemesi düzeltildi. Açık panelde savaş satırı tıklaması, iki tarafın başkentleri arasındaki harita odağına kamerayı taşıyor; satır ve deterministik bölge fallback testleri eklendi. Regression: `TestActiveWarRowAtReturnsClickedVisibleWar`, `TestActiveWarRepresentativeRegionPrefersCapitalThenDeterministicRegion`; doğrulama: `go test ./internal/render`.

- 2026-07-28: Save varsa ana menü ilk açılışında varsayılan seçim `Yeni Oyun`
  yerine `Devam et` satırına alındı. Save yoksa başlangıç seçimi değişmeden
  `Yeni Oyun` olarak kalır. Regression: `TestInitialMainMenuCursorPrefersContinueWhenSaveExists`;
  doğrulama: `go test ./internal/render ./internal/game`.

- 2026-07-28: Barış teklif panelindeki oran ile gerçek kabul kararı ayrışıyordu;
  özellikle 1300 senaryosunda ekran eski yaklaşık formülle `%100` gösterirken
  hedefin barış değerlendirmesi teklifi reddedebiliyordu. `AssessPeaceProposal()`
  hedef-perspektifli ortak seam olarak renderer ve `Execute()` akışına bağlandı;
  kabul edilmeyen teklif artık `%100` gösterilmiyor, kabul edilen `%100` ise
  `Kesin kabul` etiketi taşıyor. Regression:
  `TestPeaceProposalAssessmentMatchesExecutionAndNeverShowsFalseCertainty`,
  `TestDiplomacyPeaceChanceUsesRealAcceptanceRules`; doğrulama: hedefli
  diplomasi/render testleri.

- 2026-07-28: Deniz savaşını kaybeden donanmalar artık kısmi zayiatla denizde
  kalmıyor; filo batıyor ve üzerindeki `EmbarkedUnits` kara ordusu da state'ten
  birlikte siliniyor. Birleşik savunmadaki tüm yenilen filolar ve taşınan
  komutan bağlantıları `GameState.RemoveArmy()` üzerinden temizleniyor. Regression:
  `TestNavalBattleLossSinksFleetAndEmbarkedArmy`,
  `TestNavalBattleDefeatSinksDefenderFleetAndEmbarkedArmy`; doğrulama:
  `go test ./internal/game -run 'TestNavalBattle' -count=1`.

- 2026-07-28: Üst tarih HUD'una aktif zorluk seviyesi üçüncü satırda
  (`Zorluk: Kolay/Normal/Zor`) gösterildi. Etiket ayarlar ekranıyla ortak
  yardımcıdan üretiliyor; doğrulama: `go test ./internal/render -count=1`.

- 2026-07-28: Yeni vassal yapılan devletler artık en az 12 tamamlanmış tur
  geçmeden ilhak edilemiyor. Vassallık başlangıç turu save/load ile korunuyor;
  ilhak yönetim kartı aynı diplomasi kapısından beslenerek süre dolana kadar
  pasifleşiyor. Regression: `TestAnnexVassalRequiresTwelveCompletedTurns`,
  `TestVassalizationStoresStartTurnForAnnexationCooldown`; doğrulama: hedefli
  diplomasi/oyun testleri.

- 2026-07-28: Oyuncunun kendi üretime uygun kara bölgesine çift tıklama, harita
  bölgesini seçtikten sonra alt HUD'daki `Ordu` butonuyla aynı recruit paneli
  state geçişini çalıştırıyor. Yabancı bölge çift tıklamasının diplomasi akışı
  korunuyor. Ortak `toggleRecruitPanelFromBottomAction()` helper'ı ve regression:
  `TestMapRegionDoubleClickOpensDiplomacyForForeignRegion`; doğrulama:
  `go test ./internal/render -run 'TestMapRegionDoubleClick|TestSelectMapRegion' -count=1`.

- 2026-07-28: Ordu detay panelindeki manuel birleştirme tek hedef yerine aynı
  konumdaki her dost ordu için ayrı, hedef ordunun mevcut birim sayısını gösteren
  `->N` düğmesi kullanıyor. Düğme tıklaması
  hedef ArmyID'sini taşıyor; hover'da hedef ordunun küçük birim kompozisyonu
  popup'ı gösteriliyor. Regression: `TestMergeButtonsCarryTargetAndResultingUnitCount`,
  `TestMergeArmiesUsesExplicitTargetArmy`; hedefli render/game testleri geçti.

- 2026-07-28: Savaş ilanı onay panelinde vassal katılımcı kartlarının not satırı
  kart altına taşıyordu; kart yüksekliği satır aralığıyla eşitlendi ve not satırı
  içeri alındı. Regression: `TestWarConfirmListViewportsDoNotOverlap`; doğrulama:
  `go test ./internal/render -run 'TestWarConfirm' -count=1`.

- 2026-07-28: Pause menüsünde müzik seviyesi azaltma tıklaması düzeltildi. Sol ok
  artık `-10`, sağ ok `+10` üretir; keyboard ok tuşlarının `-5/+5` davranışı korunur.
  Regression: `TestPauseMusicVolumeClickUsesArrowSide`; doğrulama: `go test ./... -count=1`
  (mevcut 1300 scenario fixture hataları dışında).

- 2026-07-28: Seçili bölge sınır vurgusu düzeltildi. Sınır segmentinin seçili
  bölgeyi `a` veya `b` tarafında taşımasından bağımsız olarak tüm çevre sarı
  stile atanıyor; seçili kontur 3 px, diğer harita sınırları mevcut 1.25 px
  kalınlığını koruyor. Regression: `TestSelectedBorderStyleCoversBothSidesOfRegionBoundary`,
  `TestSelectedMapBorderUsesThreePixelStroke`; doğrulama: `go test ./internal/render -count=1`.

- 2026-07-28: Sınır boya/sil maskesi dünya piksel çözünürlüğüne taşındı. En küçük
  fırça artık görünen nokta/çizgi karesini doğrudan boyuyor; tek pikselin ring
  dönüşümünden sonra aynı yerde kalması regression testiyle doğrulanıyor:
  `TestSingleShapeWorldPixelRoundTripsThroughRings`.

- 2026-07-28: Yerleşim marker ve etiketleri zoom seviyesine göre LOD filtresine
  bağlandı. Uzakta yalnız başkent/kale, orta yakınlıkta liman/şehir, yakında
  kasabalar dahil tüm yerleşimler görünür. LOD eşikleri `1.25` ve `1.8` olarak
  ayarlandı; gizli yerleşimler hover ve tıklama hedefi de oluşturmaz. Regression:
  `TestSettlementVisibilityUsesZoomTiers`.

- 2026-07-28: Shape ve Bölge boyama fırçalarının yarıçap ölçüsü ortak dünya
  pikseline getirildi. Shape maskesi ölçekli mesafe hesabı kullanıyor; böylece
  aynı fırça kademesi Shape seçiminde gereksiz büyümüyor. Regression:
  `TestShapeBrushRadiusUsesWorldPixelUnits`.

- 2026-07-27: Edit mode shape konturu ile renkli raster alanı hizalandı. Kontur
  ve raster aynı world-pixel dönüşümünü kullanıyor; Voronoi debug sınırları piksel
  merkezine çiziliyor. Regression: `TestShapeOutlineUsesRasterizedWorldBoundary`.

- 2026-07-27: Edit mode boya/sil araçlarında dünya koordinatı ile raster hücre
  merkezi hizalandı. Shape ve bölge boyama `floor` ile doğru hücreyi seçiyor;
  preview ve cursor hücre merkezine snap oluyor. Regression: `TestPaintCoordinatesUseContainingCellAndItsCenter`; doğrulama: `go test ./... -count=1`.

- 2026-07-27: Edit mode `Shape/Bolge Boya/Sil` preview performansı düzeltildi.
  Her frame'de piksel listesi ve `Overlay` üretmek yerine stroke değişimleri tek
  world-space preview image'a artımlı yazılıyor; region stroke sırasında geçici
  piksel map/list allocation'ı kaldırıldı. Doğrulama: `go test ./internal/render -count=1`.

- 2026-07-27: Edit mode `Sınır Boya/Sil` ve `Bölge Boya/Sil` araçlarında fırça
  yarıçapının `1.00` altına iki ince kademe (`0.75`, `0.50`) eklendi. Canlı
  cursor, shape mask ve region override stroke aynı float yarıçapı kullanıyor;
  regression: `TestShapeBrushSupportsTwoFineStepsBelowOnePixelRadius`.

- 2026-07-27: Edit mode Voronoi debug paneli artık ordu ikonlarından sonra çiziliyor;
  ordu kareleri panelin arka planı ve metinleri üzerine binmiyor. Sınır ve merkez
  işaretleri harita katmanında bırakıldı; doğrulama: `go test ./internal/render` ve
  `go test ./... -count=1`.

- 2026-07-27: İttifak teklifleri mevcut müttefiklerin savaşlarıyla uyumlu hale getirildi;
  bir devlet, müttefikinin o anda savaşta olduğu hedefe ittifak teklifi gönderemiyor.
  Kural AI teklif üretiminde, doğrudan `ActionBlockReason()` geçidinde ve bekleyen
  teklifin kabulünde ortak `allianceWarConflictBetween()` helper'ıyla uygulanıyor.
  1300'de ittifak ilişki eşiği `25`ten `40`a çıkarıldı; aynı dinin varsayılan puanı
  artık hemen ittifak kurdurmuyor. Kastilya-Portekiz, Leon-Kastilya,
  Osmanlı-Karamanoğulları, Ceneviz-Venedik ve diğer tarihsel sürtüşmeli çiftlerin
  başlangıç relation skorları da senaryo verisinde düşürüldü. Regression:
  `TestProposeAllianceRejectedAgainstCurrentAllyWarEnemy`,
  `TestQueuedAllianceOfferExpiresWhenAllyWarConflictAppears`,
  `Test1300AllianceRequiresMoreThanReligiousBaseline`; doğrulama: hedefli diplomasi,
  AI ve senaryo integrity testleri. Aynı 1300 veri paketindeki kalan bütünlük
  hataları da kapatıldı: Bosna ve Flandre üst-devlet kayıtları tamamlandı; eksik
  devlet başkentleri mevcut ve doğru sahipliğe bağlı settlement kimlikleriyle
  dolduruldu. Doğrulama: `go test ./internal/scenario -count=1` ve
  `go test ./... -count=1`.

- 2026-07-27: Edit mode `Harita` inspector'ına region `ID` değiştirme aksiyonu
  eklendi. Mevcut ID önden dolduruluyor, Ctrl+A ile değiştirilebiliyor; boş ve
  çakışan ID'ler reddediliyor. Rename sırasında region map anahtarıyla birlikte
  komşular, karasal geçişler, kara/deniz ordularının konumları, paint override'ları
  ve seçim state'i güncelleniyor; world snapshot undo/redo bu değişikliği koruyor.
  Regression: `TestRenameRegionIDUpdatesEditorReferences`; doğrulama:
  `go test ./internal/render`.

- 2026-07-27: 1300 senaryosunda Mekke'nin shape sınırı dışındaki `y=827` yerleşim
  koordinatı `y=797` olarak düzeltildi. Anchor artık `hejaz` bölgesindeki ham
  koordinatı kullanıyor; `Test1300MeccaSettlementAnchorUsesConfiguredCoordinate`
  regression testi fallback'e geri dönüşü yakalıyor. Doğrulama: `go test ./internal/render`.

- 2026-07-27: HRE oyuncu akışı tamamlandı. HRE seçildiğinde alt HUD'daki
  `İmparatorluk` düğmesi veya `I` kısayolu otorite, imparator, siyasi takvim ve
  bağımsız üye listesini açıyor; üye satırları mevcut diplomasi akışına bağlanıyor.
  Diyet ve imparatorluk seçimi HRE oyuncusunda pending karar olarak panelde açılıyor,
  karar verilmeden tur bitmiyor ve compact save/load içinde korunuyor. Savaş önizlemesi
  imparatorluk üyelerinin tam katılım/sınırlı destek/tarafsızlık sonuçlarını ayrı
  başlıkta gösteriyor. Regression: `TestImperialPanelAvailableOnlyForHREPlayer`,
  `TestImperialPanelLayoutDrawsAtCampaignResolutions`,
  `TestPlayerImperialPoliticsCreatesAndResolvesDietDecision`,
  `TestPlayerImperialElectionUsesSelectedValidCandidate`.

- 2026-07-27: İmparatorluk üyesi bölgelerin bilgi paneline, sahip devlet
  bayrağının sağında bağlı imparatorluğun küçük bayrak rozeti eklendi. Rozet
  yalnız `ImperialState.Members` içindeki sahiplerde gösteriliyor; doğrudan
  imparatorluk bölgelerinde ikinci bayrak çizilmiyor. Regression:
  `TestRegionImperialBadgeUsesMemberEmpireAndSitsBesideOwnerFlag`.

- 2026-07-26: HRE için bağımsız imparatorluk kurumu eklendi. `GameState.Imperial`
  ve 1300 `data/imperial.json` artık otorite, Diyet takvimi, üyelik sadakati/özerkliği
  ve elektör ağırlıklarını save-backed taşıyor. `AssessImperialWarCall()` mevcut savaş
  önizlemesi ve koalisyon akışına bağlandı: imparatorluk üyeleri tam savaşa katılabiliyor,
  sınırlı altın/tahıl desteği verebiliyor veya tarafsız kalabiliyor; gerçek vassalların
  `SameRealm` otomatik katılımı korunuyor. `AdvanceImperialPolitics()` periyodik Diyet'i
  ve `HoldImperialElection()` imparator seçimini çözüyor. Regression:
  `TestImperialWarPreviewIncludesIndependentMembers`,
  `TestImperialWarCallJoinsHighCommitmentMember`,
  `TestImperialWarCallCanSendLimitedSupportWithoutEnteringWar`,
  `TestImperialElectionUpdatesEmperorAndResetsAuthority`; doğrulama:
  `go test ./internal/state ./internal/diplomacy ./internal/save ./internal/game ./internal/scenario`.

- 2026-07-26: 1300 HRE haritasındaki çekirdek siyasi parçalanma veri seviyesinde
  görünür hale getirildi. Avusturya, Bohemya, Bavyera, Saksonya ve Brandenburg artık
  HRE'nin doğrudan sahipliği yerine ayrı imparatorluk üyeleri olarak kendi başkent,
  ordu, komutan, ekonomi ve AI hedeflerine sahip; HRE yalnızca kalan imparatorluk
  bölgelerini ve gerçek vassalı Flandre'yi tutuyor. Regression:
  `TestLoad1300LoadsImperialState`, `Test1300CoreImperialMembersHaveStartingCommandersAndArmies`;
  doğrulama: `go test ./internal/scenario ./internal/game ./internal/diplomacy`.

- 2026-07-26: Bölge bilgi paneli ilk açıldığında komşu listesi artık varsayılan
  olarak `Tümünü Göster` durumunda geliyor. Aynı bölge içindeki kullanıcı daraltması
  korunuyor; yeni bölge seçiminde genişletilmiş varsayılana dönülüyor. Regression:
  `TestSelectMapRegionDefaultsToExpandedNeighborList`; doğrulama: `go test ./internal/render`.

- 2026-07-26: 1300 senaryosundaki 37 otomatik `new_region_*` kaydı anlamlı coğrafi
  kara/deniz ID ve görünen adlara taşındı; Marmara-Boğaz-Karadeniz deniz bağlantısı
  semantic ID'lerle korundu. Aynı koordinattaki iki komşusuz sahte deniz kaydı ile boş
  settlement girdileri kaldırıldı; Scania (Lund/Malmö) ve Ragusa (Dubrovnik/Kotor)
  yerleşimleri eklendi. Regression: `Test1300ScenarioRegionNamesAndIDsAreSemantic`,
  `Test1300ScenarioEveryLandRegionHasSettlement`, `TestScenarioSeaAdjacency_MarmaraBridgesAegeanAndBlackSea`;
  doğrulama: `go test ./internal/scenario ./internal/world`.

- 2026-07-26: 1300 senaryosunda Macaristan'ın doğrudan sahiplikleri tarihsel olarak
  ayrıştırıldı. `serbia` Sırp devletine, `slovenia` HRE vassalı `carniola_margraviate`'e;
  `croatia`/Kvarner/Hum/Hersek Hırvat vassalına ve `bosnia` `bosnian_banate`'e verildi.
  Üç bağlı devlet için kaynak-üretim stokları, teknoloji, başlangıç orduları, komutanlar,
  diplomasi ve AI hedefleri eklendi; Kvarner-Senj ve Hersek-Trebinye yerleşimleriyle boş
  bölge kayıtları tamamlandı. Regression: `Test1300HistoricalUnownedRegionsAreAssignedToNewStates`,
  `Test1300ScenarioArmyReferencesExist`, `Test1300ScenarioCapitalSettlementsExist` ve
  `Test1300ScenarioProfilesCoverRegionalObjectives`; doğrulama: `go test ./internal/scenario`.

- 2026-07-26: 1300 senaryosunda sahipsiz kara bölgeleri tarihsel devletlerle dolduruldu;
  merkezî devlet kontrolü olmayan Arab Çölü istisna olarak çekişmeli bölge bırakıldı.
  Merînî, Zeyyânî Tlemsen, Hafsî, Berka, Usfûrî ve Hürmüz devletleri AI-only olarak
  eklendi; Orta Cezayir Konstantin olarak bölündü, Annaba/Biskra yerleşimleri taşındı,
  Malta Aragon'a, Ermenistan/Basra İlhanlılara ve Körfez bölgeleri ilgili sultanlıklara
  bağlandı. Başlangıç kara orduları, Hürmüz filosu, komutanlar, stok/üretim değerleri,
  diplomasi ve AI hedefleri senaryo JSON'larına işlendi. Hicaz doğrudan Memlük yerine
  Memlük vassalı Mekke Şerifliği'ne verildi; Arab Çölü sahipsiz/çekişmeli bırakıldı. Regression:
  `Test1300HistoricalUnownedRegionsAreAssignedToNewStates`,
  `Test1300ScenarioStartingNaviesAreDockedAtHistoricalPorts`; doğrulama:
  `go test ./internal/scenario ./internal/diplomacy ./internal/game -run '1300|Replay'`.

- 2026-07-26: Shape edit araçları toggle davranışına geçirildi. Aktif `Shape
  Boya/Sil` veya `Bolge Boya/Sil` butonuna ikinci tıklama aracı kapatıyor;
  `>` seçimi, canlı fırça önizlemesi ve brush cursor'u temizleniyor.

- 2026-07-26: Özel karasal geçiş veri modeli eklendi. Senaryo başına
  `data/land_passages.json` dosyası `from/to/type/move_cost/defense_bonus`
  alanlarına ek olarak isteğe bağlı `start/end` `[x,y]` uç noktalarını
  taşıyor; 1300 senaryosuna Sicilya-Napoli, Konstantiniyye-Bitinya,
  Gelibolu-Atikhisar, Danimarka-İsveç, Ulster-İskoçya ve Fas-Gırnata geçişleri
  eklendi. Renderer bu bağlantıları kalın kesikli çizgi ve uç noktalarıyla
  gösteriyor; edit mode `Shape` sekmesindeki `Geçiş Ekle`/`Geçiş Düzenle`/
  `Geçiş Sil`/`Komşu Ekle` butonları veya `P`/`Delete` ile tam koordinatlı
  boğaz geçişi ekleyip mevcut uçları ayarlayabiliyor; `Komşu Ekle` iki yönlü
  deniz üstü kara hareket bağlantısı yazıyor. Hareket/savaş entegrasyonu sonraki fazda.

- 2026-07-26: Harita bölge ve deniz sınırları raster harita dokusundan ayrıldı. `regionAt` kenarları cache'lenmiş yatay/dikey kontur parçalarına sıkıştırılıyor; diplomasi, seçili bölge ve ticaret modu renkleri `map_borders.go` üzerinden antialiased screen-space mesh olarak çiziliyor. Zoom sırasında sınır kalınlaşması/basamaklanması giderildi; edit mode region assignment değişiklikleri kontur cache'ini yeniliyor. Regression: `TestVectorBorderStylesHighlightPlayerRealmAndAlliedRealms`; doğrulama: `go test ./internal/render` ve `go test ./...`.

- 2026-07-31: `Normal`/`Ticaret` harita modu düğmeleri minimap'in hemen üstüne taşındı; Ticaret modundaki `Pazar` düğmesi aynı HUD kümesinin üst satırında kalıyor. Çizim ve tıklama geometrisi ortak helper'lardan okunuyor, `Pazar` HUD hit-test ve ticaret overlay engellemesine dahil ediliyor. Regression: `TestMapModeHudSitsAboveMinimapWithTradeToggle`; doğrulama: `go test ./internal/render -run 'Test(MapModeHudSitsAboveMinimapWithTradeToggle|CoreUIGeometryFitsCommonViewports)$'`.

- 2026-07-26: Oyuncuya ait olmayan kara bölgesine aynı bölge içinde 400 ms
  içinde çift tıklanınca, bölge bilgi panelindeki `Diplomasi` düğmesiyle aynı
  `openDiplomacyTarget()` akışı kullanılarak ilgili devletin teklif paneli
  doğrudan açılıyor. İlk tıklama yalnız bölgeyi seçiyor; oyuncu bölgeleri,
  deniz/sahipsiz bölgeler açılımı tetiklemiyor. Yerleşim etiketi tıklaması da
  region-ID tabanlı ortak seçim akışına bağlandı. Regression:
  `TestMapRegionDoubleClickOpensDiplomacyForForeignRegion`; doğrulama:
  `go test ./internal/render`.

- 2026-07-26: Vektör sınır overlay'inin performans sorunu düzeltildi. Screen-space mesh yalnız kamera/ekran/harita durumu değiştiğinde hazırlanıyor; hazır transparan overlay statik framelerde tekrar kullanılıyor, ekran dışı segmentler eleniyor ve uzak zoom'da bir pikselden küçük yardımcı/deniz sınırları çizilmiyor. DirectX dinamik buffer baskısı azaltıldı. Doğrulama: `go test ./internal/render`, `go test ./...`, `GOOS=windows GOARCH=amd64 go build -o /tmp/mapp-game-go-vector-mesh-check.exe ./cmd/game`.

- 2026-07-25: Bölge panelinde nüfus satırı sonrasında oluşan dikey geometri kayması düzeltildi. Vergi `+/-` düğmeleri artık vergi satırıyla hizalı; `BİNALAR/OLAYLAR` sekmeleri de nüfus satırının altında çizim ve hit-test ile birlikte aşağı taşınıyor. Regression: `TestRegionTaxButtonsUseCompactAlignedRects`, `TestRegionTaxInteractiveBarStopsBeforeDecreaseButton`; doğrulama: `go test ./...`.

- 2026-07-25: Bölge bilgi paneline toplam nüfus ve kırsal/yerleşim kırılımı eklendi. Gösterim, tahıl tüketiminde kullanılan `Region.Population` toplamını ve `SettlementPopulation()` dağılımını doğrudan state'ten okuyor. Regression: `TestRegionPopulationDisplayTextIncludesTotalAndBreakdown`; doğrulama: `go test ./internal/render` ve `go test ./...`.

- 2026-07-25: Nüfus modeli yerleşim/kırsal ayrımına taşındı. `Settlement.Population` yerleşim nüfusunu, `Region.RuralPopulation` köyler ve yerleşim dışı kırsal nüfusu tutuyor; `Region.Population` bu bileşenlerin toplamından oluşuyor. 1300 ve 1455 ham verileri mevcut toplam nüfusu koruyacak şekilde dağıtıldı; sivil tahıl tüketimi toplam bölge nüfusunu kullanıyor. Eski save'ler yalnız toplam nüfusu taşıdığında değer kırsal nüfusa göç ediliyor. Regression: `Test1300ScenarioSettlementPopulationLeavesRuralPopulation`, `TestRegionRecalculatePopulationCombinesRuralAndSettlements`; doğrulama: `go test ./...`.

- 2026-07-25: Bina inşaatlarının işçi iaşesi artık senaryo verisinde daha görünür bir tahıl maliyeti taşıyor. `1300_ottoman_rise` ve `1455_wars_of_the_roses` binalarında tahıl gereksinimi yaklaşık `%20` artırıldı; mevcut ödeme/iade, AI bütçe ve tooltip akışları aynı `ResourceCost` üzerinden çalışmaya devam ediyor. Regression: `Test1300ScenarioResourceSpecializationsAndProductionCosts`; doğrulama: `go test ./internal/scenario ./internal/game ./internal/ai ./internal/render`.

- 2026-07-25: Tahıl ekonomisi tarihsel kıtlık baskısı için yeniden dengelendi. Acil tahıl satışı, otomatik ihracat ve oyuncunun doğrudan tahıl satışı tur başına kuşatma dışı temel vergi gelirinin `%100`'üyle sınırlandı; sivil tüketim `ceil(population/18)` oldu. `1300_ottoman_rise` ham verisinde birlik tahıl bakımları yaklaşık `%33` artırıldı, İngiltere ve Fransa tahıl üretimi aşağı çekildi. Regression: `TestGrainSaleBudgetTracksTaxIncomeAndResetsEachTurn`, `Test1300ScenarioGrainEconomyBands`; doğrulama: `go test ./internal/state ./internal/game ./internal/render ./internal/economy`.

- 2026-07-25: Komutan portresi harita üzerindeki donanma ikonlarına da eklendi. Filo komutanı veya taşınan kara komutanı, deniz ikonunun üstünde gösteriliyor; nakliye filosundaki `EmbarkedUnits` rozeti varsa portre onun üstünde konumlanıyor. Regression: `TestArmyCommanderBadgeAlignsAboveLandAndNavalIcons`; doğrulama: `go test ./internal/render`.

- 2026-07-25: Donanma konumları liman ve açık deniz olarak mekanik düzeyinde ayrıştırıldı. Docked filonun kanonik konumu `DockedSettlementID` üzerinden `Army.LocationID()` ile okunuyor; yalnız `IsAtSea()` olan filolar deniz savunmasına, abluka hesabına ve AI deniz tehdidine katılıyor. Deniz hareketi dock bağını temizliyor; manuel/AI filo birleştirme ve liman deniz üretimi de settlement konumunu eşleştiriyor. Regression: `TestNavalLocationSeparatesPortFromSea`, `TestSelectBattleDefenderIgnoresDockedFleetAtSea`, `Test1300NavalThreatSnapshotMarksApproachedPort`; doğrulama: hedefli `go test` kontrolleri `internal/army`, `internal/state`, `internal/ai`, `internal/game` ve `internal/render` paketlerinde geçti.

- 2026-07-25: Save'den yüklemede açık denizdeki donanmaların komşu sahip limana yanlışlıkla taşınması düzeltildi. `loadFromPath()` artık legacy otomatik docking migrasyonunu çalıştırmıyor; kayıtlı `RegionID` ve dock alanları korunuyor. Legacy docking yalnız başlangıç senaryosundaki eksik dock verisini uyumlulaştırıyor. Regression: `TestLoadFromPathKeepsOpenSeaFleetUndocked`; doğrulama: `go test ./internal/save ./internal/army ./internal/state ./internal/ai ./internal/game`.

- 2026-07-25: Donanma çıkarma akışı tahkimli kıyılarda düzeltildi. Kale/duvar bulunan düşman kıyısına çıkarma artık amfibi savaş veya anında fetih üretmiyor; kara ordusu karaya inerek aktif kuşatma başlatıyor. Otomatik embark sırasında kara komutanı `EmbarkedCommander` olarak filoya taşınıyor; çıkarma preview'si, savaş hesabı ve karaya çıkan ordu filo komutanı yerine kara komutanının bonuslarını kullanıyor. Regression: `TestMoveArmyDisembarkEnemyFortressStartsSiegeWithoutBattle`, `TestMoveArmyDisembarkEnemyFortressWithoutDefenderStillStartsSiege`, `TestMoveArmyEmbarkTransfersLandCommanderInsteadOfFleetCommander`, `TestAIMoveArmyDisembarksToEnemyFortressAndStartsSiege`; doğrulama: `go test ./internal/state ./internal/game ./internal/render ./internal/ai`.

- 2026-07-25: Merchant filo footer'ına gemi katkısı ve konum doğrulaması eklendi. Seçili filonun gemi sayısı, rotaya uyguladığı `+N hacim/tur` bonusu ve rota denizinde olup olmadığı aynı satırda gösteriliyor; hesap `MerchantFleetTradeRouteBonus()` ile state çözümündeki iki gemi sınırıyla ortaklaştırıldı. Regression: `TestMerchantTradeBonusUsesAssignmentLocationAndRouteCap`, `TestMerchantTradeBonusRequiresActiveConnectedCenterSea`; doğrulama: `go test ./internal/state ./internal/render`.

- 2026-07-25: Ticaret koridoru hover bilgisi oyuncu akışıyla genişletildi. Liman rotalarında etkin hacim, emtia, oyuncunun verdiği/alacağı kaynak, tur başına altın geliri veya altın ödemesi ve askı durumu gösteriliyor; tooltip yüksekliği içeriğe göre büyüyor. Değerler `EffectiveAmountPerTurn()` ve rota birim fiyatından türetiliyor; doğrulama: `go test ./internal/render`.

- 2026-07-25: Limanlar arası ticaret koridorlarının çizimi sadeleştirildi. Oyuncu rotaları tek turuncu renkte çiziliyor; dash/gap uzunlukları eğrinin ve rotanın toplam uzunluğundan bağımsız olarak 12/10 px sabit tutuluyor. Regression kapsamı `drawDashedTradeCurve` yay uzunluğu örneklemesini kullanıyor; doğrulama: `go test ./internal/render`.

- 2026-07-25: Merchant filo footer'ındaki görsel sıkışma düzeltildi. `ROTA ATA` butonu artık footer alt sınırına yapışmıyor; 22 px yüksekliğinde, üst-alt inset'li çiziliyor ve aktif rota metni butonun gerçek bitişinden sonra başlıyor. Regression: `TestMerchantRouteButtonHasFooterInset`; doğrulama: `go test ./internal/render`.

- 2026-07-25: Merchant rota modalındaki görsel örtüşme düzeltildi. İki satırlı seçenekler 48 px satır yüksekliği ve daha güvenli iç boşlukla çiziliyor; modal karartması güçlendirilerek alttaki ordu panelinin `ROTA ATA` butonu ve metinleri arka plana alınıyor. Regression: `TestMerchantRoutePanelTwoLineRowsHaveVerticalClearance`; doğrulama: `go test ./internal/render`.

- 2026-08-01: Merchant rota modalındaki iki satırlı seçenek alanları 58 px satır yüksekliğine çıkarıldı; rota adı ve mal/gelir satırı artık iç kutuya taşmadan sığması için gerçek genişliğe göre kırpılıyor. Modal çerçevesi, başlık/alt metinler, rota satırları ve kapatma düğmesi ortak UI component/compose yüzeylerine taşındı; kapatma düğmesi diğer panellerdeki `IconClose` + `tinyButtonStyle` sözleşmesini kullanıyor. Regression: `TestMerchantRoutePanelTwoLineRowsHaveVerticalClearance`, `TestMerchantRoutePanelUsesSharedCloseIconButton`, `TestMerchantRoutePanelRowGeometryMatchesExpandedHeight`; doğrulama: `go test ./internal/render`.

- 2026-08-01: Donanma görev modalındaki iki satırlı seçenek alanları 62 px satır yüksekliğine çıkarıldı; görev adı/açıklaması gerçek satır genişliğine göre taşmasız kırpılıyor. Panel çerçevesi, başlık/alt metinler, görev satırları ve kapatma düğmesi ortak UI component/compose yüzeylerine taşındı; kapatma düğmesi `IconClose` + `tinyButtonStyle` kullanıyor. Regression: `TestNavalMissionPanelTwoLineRowsHaveVerticalClearance`, `TestNavalMissionPanelUsesSharedCloseIconButton`, `TestNavalMissionPanelRowGeometryMatchesExpandedHeight`; doğrulama: `go test ./internal/render`.

- 2026-07-25: Merchant rota filtresindeki yapısal boşluk düzeltildi. Tarihsel ticaret merkezi olmayan aktif anlaşmalar, tarafların sahip olduğu bağlı limanlar ve deniz bağlantısı üzerinden `ROTA ATA` panelinde listeleniyor; `MerchantTradeRoutePortPairs()` ortak uç modeliyle oyuncunun kendi aktif anlaşmaları `Ticaret` haritasında turuncu renkli kesikli liman-liman koridor olarak çiziliyor. Regression: `TestMerchantTradeRouteUsesConnectedOwnedPortsWithoutHistoricalCenters`; doğrulama: `go test ./internal/state ./internal/render`.

- 2026-07-25: Aktif kuşatma ikonları savaş ilişkisini gösterecek şekilde yeniden dizildi. Kuşatan ordu karesi solda, yuvarlak kılıç rozeti ortadaki ayrı slotta, savunan/rakip ordu karesi sağda çiziliyor; kuşatan ve savunmacı merkezleri arasında 52 px boşluk bırakılıyor. Regression: `TestArmyIconPositionsKeepBesiegerLeftOfSplitPart`; doğrulama: `go test ./internal/render`.

- 2026-07-25: Harita üzerindeki kara ordularında komutan portresi görünür hale getirildi. Atanmış `Army.Commander` için portre, birim sayısı karesinin hemen üstünde sayı karesinden biraz büyük 28×28 px rozet olarak çiziliyor; sayı karesi ve donanma ikonlarının mevcut düzeni korunuyor. Kapsam: `internal/render/{renderer.go,renderer_army_icon_test.go}`, `wiki/architecture/render-pipeline.md`; regression: `TestArmyCommanderBadgeSitsAboveAndExceedsCountSquare`; doğrulama: `go test ./internal/render`.

- 2026-07-25: Bölge panelindeki oyuncu vergi barı yatayda genişletildi. Bar normal seviye barıyla aynı x konumundan başlıyor, `-` butonundan önce yalnızca 4 px boşluk bırakıyor ve yüksekliği değişmiyor. Regression: `TestRegionTaxInteractiveBarStopsBeforeDecreaseButton`; doğrulama: `go test ./internal/render`.

- 2026-07-24: AI savaş/diplomasi planının Faz 6 teşhis geçmişi tamamlandı. Geliştirme
  save'i yüklendiğinde ilk beş AI fazının plan, hedef, cephe, aktif savaş, yedek ve
  bloklanma özeti deterministik sırayla toplanıyor; beşinci faz sonrası
  `autosave.debug.json` içindeki `state.ai_diagnostic_history` alanına yazılıyor ve
  F3 modalında devlet bazında gösteriliyor. Normal compact save payload'ı bu geçici
  geçmişi taşımıyor. Regression: `TestRecordAIDiagnosticRoundCapturesSortedAIHistory`
  ve debug sidecar save doğrulaması; doğrulama: `go test ./internal/ai ./internal/save ./internal/render ./internal/diplomacy ./internal/state`.
- 2026-07-24: 1300 tempo kabul ölçümü ayrıştırıldı. 42 aylık altın bantları yalnız
  medium profilinde doğrulanıyor; 120 aylık calibration raporu uzun dönem teşhisi
  olarak kalıyor. Böylece Venedik'in 120 ayda biriken altını yanlışlıkla 42 aylık
  banda sokulmuyor. `RUN_SCENARIO_TEMPO_REPORT=medium` ve `calibration` profilleri
  doğrulandı.
- 2026-07-24: AI savaş/diplomasi planının son telemetry adımı tamamlandı. 24, 42 ve
  120 tur tempo raporları devlet bazında savaş başlangıcı, aktif savaş-ay,
  tamamlanan savaş süresi, fetih, barış ve stalemate ortalamalarını logluyor.
  Telemetry test harness'inde kalıyor; save/state şeması değişmiyor. Doğrulama:
  `Test1300ScenarioGrainEconomyBands` tekil, `RUN_SCENARIO_TEMPO_REPORT=medium`
  ve `Test1300ScenarioAITwoTurnReplayIsDeterministic`.

- 2026-07-24: AI savaş/diplomasi planının Faz 3 mobilizasyon ilk adımı tamamlandı.
  1300 senaryosunda ilk 24 tur açılış temposu korunuyor; sonrasında yeni savaş ilanı
  altın acil rezervi ve iki aylık operasyon tahıl stoğuyla sınırlandırılıyor. Aktif
  savaşta kritik/kıtlık seviyesi saldırı rolünü savunma/ikmal rolüne çeviriyor; uyarı
  seviyesi cepheyi kilitlemiyor. `GrainEconomyStatus` yoksa aynı state talep fallback'i
  kullanılıyor. Regression: `TestStrategicWarReadinessUsesGoldAndGrainReserves`,
  `TestStrategicWarLogisticsGatePreservesOpeningTempo`; ayrıca mature ana cephedeki
  tahkimli hedef için `55/25/20`, tahkimatsız hedef için `60/25/15`
  piyade/süvari/kuşatma kompozisyonu ve kritik tehditte `75/15/10` savunma
  kompozisyonu seçiliyor. Birden fazla aktif veya pozitif tehditli cephede stratejik
  yedek oranı `%25`e, kritik tehditle `%30`a çıkıyor. Regression:
  `TestAICompositionTargetFollowsMaturePrimaryFront`, `TestMultipleActiveFrontsRaiseReserveWithoutCriticalThreat`;
  doğrulama: `go test ./internal/ai` ve
  `go test ./internal/game -run '^Test1300ScenarioGrainEconomyBands$'`.

- 2026-07-24: Faz 4 ortak cephe koordinasyonunun ilk dilimi tamamlandı. Aynı realm
  vassal/overlord ve aktif savaşa katılmış müttefik orduları ilgili savaş cephesinin
  dost gücüne dahil ediliyor; katılımcılardan birinin geçerli `WarLedger` hedef kilidi
  diğer katılımcılara aktarılıyor. Barıştaki müttefikler ve savaşa katılmamış ordular
  hesaba katılmıyor; komuta yetkisi fraksiyonlarda kalıyor. Regression:
  `TestSameRealmWarFrontSharesTargetAndFriendlyPower`; relief görevleri de aktif
  savaşa katılmış müttefik/vassal bölgelerine genişletildi; doğrulama:
  `go test ./internal/ai ./internal/game`.

- 2026-07-24: Faz 5 geçici ateşkes adımı tamamlandı. Barış çözümünde taraf çifti için
  save-backed altı turluk `RecentTruces` kaydı tutuluyor; ateşkes sürerken savaş ilanı
  engelleniyor, süresi dolunca yeniden açılıyor. Compact save/load ve koalisyon barışı
  korunuyor. Regression: `TestAcceptedPeaceCreatesTemporaryTruce`,
  `TestCompactCampaignStatePreservesWarLedger`; doğrulama: `go test ./internal/diplomacy ./internal/save`.

- 2026-07-24: Faz 5 savaş yorgunluğu görünürlük adımı tamamlandı. `PeaceAssessment`
  artık savaş süresi/kayıpları, altın, tahıl, memnuniyet ve ilişki baskılarını ayrı
  türetilmiş alanlarda raporluyor. İlk iterasyonda skor davranışı açılış dengesini
  değiştirmemesi için bu alanlar açıklama/telemetry olarak tutuluyor. Regression:
  `TestPeaceAssessmentReportsWarExhaustionPressures`; doğrulama:
  `go test ./internal/diplomacy`.

- 2026-07-24: AI savaş/diplomasi iyileştirme planının Faz 1 temeli başladı ve tamamlandı.
  `PeaceAssessment` actor perspektifinden `WarScore` ile fetih, kayıp, mevcut bölge,
  başkent tehdidi ve expand objective ilerlemesini birleştiriyor; `ObjectiveHeld`/
  `ObjectiveTotal` hedef ilerlemesini raporluyor. Savaş skoru save'e yazılmıyor,
  güncel state'ten türetiliyor. Regression: `TestPeaceAssessmentReportsWarScoreAndObjectiveProgress`.

- 2026-07-24: AI savaş/diplomasi planının Faz 2 hedef seçimi başladı. Aktif
  savunma/konsolidasyon cephelerinde `AIFront.TargetRegionID`; stratejik değer,
  başkent, kuşatma, düşman savunma gücü ve dost erişimiyle deterministik seçiliyor;
  recruitment aynı hedefi kullanıyor. Ana saldırı cephesi seçimi ve dört turluk
  `WarLedger` hedef kilidi eklendi; diğer cepheler ikincil savunmada kalıyor. Mevcut
  expand objective sırası açılış tahıl temposunu korumak için değiştirilmedi. Regression:
  `TestFrontTargetPrefersStrategicValueOverFirstRegion`, `TestWarFrontTargetStaysLockedForShortWindow`;
  doğrulama: `go test ./...`.

- 2026-07-24: AI savaş temposu yeniden değerlendirildi. Savunma/konsolidasyon planına
  sahip devletlerin aktif savaş cephelerinde kritik tehdit yoksa tek saha ordusu
  kontrollü `assault`/`siege` rolüne atanıyor; yeni savaşlarda 12 turluk seferberlik
  korunuyor. 12 turdan uzun ve son 8 turu muharebesiz/kuşatmasız savaşlar stalemate
  barış kapısını açıyor; `defend/consolidate` planları savaş hedefi tamamlanmış sayılmıyor.
  Eski `difficulty=0` kayıtları Normal'e göç ediyor. Regression: `internal/ai`,
  `internal/diplomacy`, `internal/save`; denge: `Test1300ScenarioGrainEconomyBands`.

- 2026-07-24: Hasarlı donanmalar kendi limanına bağlandığında tur çözümündeki tamirat durumu artık ordu paneline de yansıyor. `Army.CanReplenishIn()` dock edilmiş kendi limanını tanıyor; hasarlı gemi kartlarının sağ üstünde mevcut yeşil `+` rozeti ve başlıkta `Takviye aktif` görünür. Açık deniz, düşman limanı ve limansız kıyı göstergeleri kapalı kalıyor. Regression: `TestNavalCanReplenishInOwnPort`, `TestArmyPanelReplenishmentBadgeActivatesForDamagedFleetInOwnPort`; doğrulama: `go test ./internal/army ./internal/render ./internal/game`.

- 2026-07-24: Ordu bölme ve birleştirme sonrasında, o turda henüz hareket edilmemişse hareket havuzu yeni kompozisyonun en yavaş birimine göre yeniden hesaplanıyor. Daha önce hareket edilmiş ordularda puan iadesi yapılmıyor. Regression: `TestSplitArmyRefreshesUnmovedPartsBySlowestUnit`, `TestMergeArmiesRefreshesUnmovedResultBySlowestUnit`, `TestSplitArmyDoesNotRefundMovementAlreadyUsedThisTurn`; doğrulama: `go test ./...`.

- 2026-07-24: Bölgesel tahıl lojistiği ambar rezerviyle düzeltildi. Mevcut fraksiyon stoku, bölgedeki ambar kapasitesi kadar askerî ikmale aktarılabiliyor; aynı devletin bölgeleri rezervi deterministik biçimde paylaşıyor ve başkent önceliği korunuyor. Bölge panelinde tahıl `+kalan/toplam` formatında, ordu panelinde tur başı tahıl ihtiyacı gösteriliyor. Kapsam: `internal/{game/{resolution.go,resolution_test.go},state/state.go,render/{panel.go,panel_test.go,army_panel.go}}`, `wiki/systems/economy.md`.

- 2026-07-24: Savaş ilanı koalisyon önizleme modalında uzun kesin katılanlar ve çağrılabilir müttefik listelerinin birbirinin üzerine çizilmesi düzeltildi. Sol ve sağ cephedeki iki liste alanı ayrı viewport ve bağımsız mouse-wheel scroll state'i kullanıyor; her listenin scrollbar'ı yalnız kendi taşan içeriğinde gösteriliyor ve checkbox hit-test'i yalnız görünür satırlara uygulanıyor. Regression: `TestWarConfirmListViewportsDoNotOverlap`, `TestWarConfirmScrollTargetIsLocalToViewport`; doğrulama: `go test ./internal/render -run 'TestWarConfirm'`.

- 2026-07-24: Bölge panelindeki vergi `+/-` düğmeleri 18×16 px'e küçültüldü ve içerik sağına hizalandı. Etkileşimli vergi barı iki uçtan da düğme genişliği kadar içeri alındı; tam uzunluklu barın butonların altında kalmasına neden olan çift çizim kaldırıldı. Regression: `TestRegionTaxButtonsUseCompactAlignedRects`, `TestRegionTaxInteractiveBarStopsBeforeDecreaseButton`; doğrulama: `go test ./internal/render`.

- 2026-07-23: `1300_ottoman_rise` bölge kaynakları tarihsel/coğrafi uzmanlaşmaya göre yeniden dağıtıldı. Tahıl ovalar ve nehir havzalarında, kereste orman/kuzey hatlarında, taş-demir dağ ve maden bölgelerinde, baharat Mısır-Levant-Basra/Akdeniz ticaret düğümlerinde, kumaş ise Bursa/Konstantiniyye/Selanik/Flandre/İtalya tekstil merkezlerinde yoğunlaştırıldı. Bina ve birlik veri modellerine `spice_cost`/`cloth_cost` eklendi; ortak ödeme, iade, AI bütçe, üretim kuyruğu ve tooltip akışları bu kaynakları kullanıyor. Regression: `TestResourceCostAppliesSpiceAndCloth`, `Test1300ScenarioResourceSpecializationsAndProductionCosts`, `Test1300ScenarioGrainEconomyBands`; doğrulama: hedefli paket testleri ve full `go test ./...`.

- 2026-07-23: Hareket puanı bitmiş seçili kara ordusunda dost nakliye filosunun harita üzerindeki `BIN` göstergesi artık çizilmiyor. Görsel aksiyon koşulu sağ tık input kapısıyla hizalandı. Regression: `TestEmbarkPromptRequiresSelectedArmyMovementPoints`; doğrulama: `go test ./internal/render`.

- 2026-07-23: Memnuniyet cezaları oynanabilirlik için yeniden dengelendi. Savaş yorgunluğu `-3` yerine `-1/tur`, 20+ bölge yozlaşması `-5` yerine `-1/tur`, kışla cezası ise seviye başına `-5` yerine `-1/tur` oldu. Böylece üç olumsuz durum birlikte `-3/tur` seviyesinde kalıyor; mevcut vergi, bina, ordu ve yıllık `-1` etkileriyle birlikte uzun vadeli baskı korunuyor ancak memnuniyetin tek yılda çökmesi engelleniyor. Regression: `TestApplyEconomyTickCombinesSatisfactionModifiers`; doğrulama: `go test ./internal/game ./internal/economy ./internal/state`.

- 2026-07-23: Her 12 turda bir yıl sonu yıpranması eklendi. Aralık ekonomi turunda tüm sahipli kara bölgeleri, diğer memnuniyet artı/eksileriyle aynı toplamsal delta içinde `-1` alıyor; yıl içindeki aylarda tekrar uygulanmıyor. Regression: `TestApplyEconomyTickAppliesAnnualSatisfactionDecayAtYearEnd`; doğrulama: `go test ./internal/game ./internal/economy ./internal/state`.

- 2026-07-23: Bölge memnuniyeti tur çözümünde toplamsal delta olarak yeniden düzenlendi. Kışla seviye başına `-5`, çiftlik/pazar/liman `+1`; savaş halindeki devletin tüm bölgelerine `-3` savaş yorgunluğu, 20'den fazla bölgeye sahip devletin tüm bölgelerine `-5` yozlaşma ve bölgedeki sahibine ait kara ordusu gücüne göre en fazla `+10` istikrar bonusu uygulanıyor. Vergi, bina, tahıl kıtlığı, teknoloji, kuşatma ve diğer mevcut etkiler tek hesapta birleştiriliyor. Regression: `TestApplyEconomyTickCombinesSatisfactionModifiers`, `TestRegionArmySatisfactionBonusScalesAndCaps`; doğrulama: `go test ./internal/game ./internal/economy ./internal/state`.

- 2026-07-23: Oyuncu ordularının dost bölgeye hareketinde otomatik birleşme kaldırıldı. Hareket eden ordu, hedefteki aynı fraksiyon ordusuyla ayrı kalıyor; oyuncu isterse ordu panelindeki `BİRLEŞTİR` aksiyonunu kullanarak birleşmeyi başlatıyor. Aktif kuşatma desteği ve kuşatma sonrası yerleşme de otomatik birleşme yapmıyor. Regression: `TestMoveArmyWithStanceDoesNotAutoMergeFriendlyArmy`; doğrulama: `go test ./internal/game`.

- 2026-07-23: Oyuncunun vassal devletlerine ait kara ve deniz orduları artık savaş sisi/istihbarat koşullarından bağımsız olarak kendi orduları gibi tam görünür. Harita ikonunda gerçek birim sayısı, ordu panelinde tüm birim kartları ve kart hover'ında ayrıntılı birlik bilgileri gösteriliyor; vassal ordularının AI tarafından yönetilen hareket ve komuta yetkileri korunuyor. Kapsam: `internal/render/{renderer.go,army_panel.go,army_panel_test.go}`, `wiki/architecture/render-pipeline.md`; regression: `TestPlayerCanSeeArmyDetailsIncludesVassalChain`, `TestArmyPanelUnitHoverIDUsesDisplayedCardOrder`; doğrulama: hedefli `go test ./internal/render -run 'Test(PlayerCanSeeArmyDetailsIncludesVassalChain|ArmyPanelUnitHoverIDUsesDisplayedCardOrder|ArmyIconBorderColorUsesDiplomaticPalette)$'` başarılı.

- 2026-07-23: Savaş halindeki AI teklif akışı düzeltildi. Aynı oyuncuya barış ve kuşatma teslimiyeti teklifleri birlikte oluştuğunda barış kuyruğa ve oyuncu kararına önce geliyor; barış kabul edilirse aynı savaşın teslimiyet teklifi otomatik atlanıyor. Kapsam: `internal/{ai/diplomacy.go,diplomacy/{offers.go,vassalage.go}}`; regression: `TestAIPeaceOfferIsQueuedBeforeSiegeSurrenderOffer`, `TestPeaceOfferPrecedesSiegeSurrenderAndAcceptanceSkipsIt`; doğrulama: `go test ./internal/diplomacy ./internal/ai ./internal/game`.

- 2026-07-23: Ordu panelinde birlik kartlarına tıklayarak seçili birlikleri ayrı bir orduya bölme eklendi. Seçim yokken eski ortadan bölme korunuyor; seçilen kartlar altın çerçeveyle gösteriliyor, fiziksel birim index'leri oyun katmanına aktarılıyor ve ana ordunun tamamen boşaltılması engelleniyor. Regression: `TestSplitArmyWithSelectedUnitsMovesOnlyThoseUnits`, `TestSplitSelectionRequiresOneUnitToRemain`; doğrulama: `go test ./internal/game -run 'Test(SplitArmyWithSelectedUnits|SplitBesiegingArmyKeepsSiegeWithRemainingUnit)$'` ve ilgili render testleri.

- 2026-07-23: Altın yetersizliği nedeniyle alt HUD'daki `Ordu` butonunun pasifleşmesi düzeltildi. Oyuncuya ait uygun bölgede, Kışla/birim gereksinimleri sağlandığında panel açılabilir kalıyor; kaynak eksikliği kart üzerinde gösteriliyor ve üretim emrinde state/game katmanında doğrulanıyor. Regression: `TestRecruitPanelButtonRemainsEnabledWhenGoldIsInsufficient`; doğrulama: `go test ./internal/render -run 'TestRecruitPanel(ButtonRemainsEnabledWhenGoldIsInsufficient|DisabledReasonUsesResourceShortage)$'`.

- 2026-07-22: Üst HUD'daki aktif araştırma adı tıklanabilir hale getirildi; tıklama alt HUD'daki `Teknoloji` düğmesiyle aynı şekilde teknoloji panelini açıyor. Regression: `TestTurnTechHudTechHitUsesResearchRow`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Teknoloji ağacının ortadaki flow içeriğinin sağ ve sol boşluklarına tıklama, üstteki teknoloji paneli kapatma düğmesiyle aynı kapanış davranışına bağlandı. Sekme, kart, scroll ve sürükleme hit-test'leri korunuyor. Regression: `TestTechTreeSideBlankClickClosesPanel`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Bekleyen AI diplomasi teklifleri oyuncu tarafından kabul edilirken gönderenin
  tur içi elçi kotasının ikinci kez tüketilmesi düzeltildi. Teklif kuyruğa alınırken kota
  korunuyor; kabul sırasında ilişki/stance/stratejik geçerlilik yeniden doğrulanıyor fakat
  aynı teklif tekrar ücretlendirilmiyor. Regression: `TestResolveQueuedAllianceOfferDoesNotSpendQuotaTwice`;
  doğrulama: `go test ./internal/diplomacy`.

- 2026-07-22: Bekleyen AI ittifak teklifleri kabul sırasında ortak tehdit/ticaret gibi
  teklif sonrası değişebilen stratejik koşullarla yeniden reddedilmiyor. Kabulde yalnızca
  tarafların geçerliliği, savaş ve aynı realm kontrolleri korunuyor. Regression:
  `TestResolveQueuedAllianceOfferKeepsTermsAfterStrategicStateChanges`.

- 2026-07-22: Savunma panelinden başarılı huruç sonrasında savunmacı ordu aynı bölgede kalırken en az 1 hareket hakkını koruyor. Regression: `TestSortieSiegeActionResolvesInPlace`; doğrulama: `go test ./internal/game -run '^TestSortieSiegeActionResolvesInPlace$'`.

- 2026-07-22: Kuşatılan oyuncu ordusu veya yerleşimi seçildiğinde saldıran kuşatma paneline karşılık gelen savunma paneli açılıyor. Panel tahkimat, ilerleme, gedik, teslim süresi, kuşatan/savunan gücü bilgilerini gösteriyor; `Huruç başlat` aynı bölgede huruç savaşını, `Teslim ol` ise savunma ordularının kaldırılması ve kuşatana fetih akışını çalıştırıyor. Kapsam: `internal/{render/{action.go,renderer_dialogs.go,renderer_input.go},game/{game.go,siege.go}}`, testler: `internal/{render/war_confirm_test.go,game/siege_test.go}`; doğrulama: hedefli render/game testleri geçti.

- 2026-07-22: Güller Savaşı (`1455_wars_of_the_roses`) senaryosunun açılmasını engelleyen iki veri sözleşmesi uyumsuzluğu düzeltildi. `factions.json` içindeki `research.completed` alanları `map[string]bool` formatına geçirildi; `ally` AI objective türü diplomatik metadata olarak kabul edilip askeri planlayıcıdan ayrıştırıldı. Regression: `TestLoad1455WarsOfTheRosesScenario`, `TestLoadAIConfigAllowsAllianceObjectiveMetadata`; doğrulama: `go test ./internal/game ./internal/ai`, 1455 JSON shape kontrolü.

- 2026-07-22: Tur sonu AI hamleleri ve turn resolution tamamlandığında oyuncu turu aktif başkent bölgesi seçili olarak açılıyor; mevcut başkent geçersizse seçim zorlanmıyor. Kapsam: `internal/{game/game.go,render/renderer.go}`, regression: `TestSelectPlayerCapitalRegionSelectsActiveCapitalRegion`; doğrulama: `go test ./internal/render -run '^TestSelectPlayerCapitalRegionSelectsActiveCapitalRegion$'`, `go test ./internal/game`.

- 2026-07-22: Teknoloji paneli açıkken tekerlek ve fare inputunun arka plandaki haritaya sızması düzeltildi. Teknoloji paneli inputu kamera işleme adımından önce tüketiyor; ağaç scroll/drag yalnız panel pan state'ini, panel dışı tıklamalar ise hiçbir harita seçimini değiştirmiyor. Kapsam: `internal/render/{renderer_input.go,tech_panel.go}`, `wiki/architecture/render-pipeline.md`; doğrulama: `go test ./...`.

- 2026-07-22: `OLAYLAR` sekmesindeki `Komşu (...) [Daralt] / [Tümünü Göster]` başlığı için eksik kara-bölge hit-test'i düzeltildi. Toggle artık komşu görünümünü değiştiriyor; içerik yüksekliği, scrollbar ve scroll clamp'i daraltılmış/genişletilmiş listeye göre güncelleniyor. Regression: `TestRegionPanelEventTabScrollAndSharedContentArea`; doğrulama: `go test ./...`.

- 2026-07-22: `OLAYLAR` sekmesi açıkken gizli bina kartlarının hover tooltip üretmesi düzeltildi. Hover çözümlemesi aktif bölge paneli sekmesine bağlandı; bina tooltip'i yalnız `BİNALAR` sekmesinde çalışıyor. Regression: `TestNonOwnedBuildingCardsAreNotActionableOrHoverable`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Bölge bilgi panelindeki bina ve aktif olay/komşu içeriği iki sekmeli ortak alana taşındı. `BİNALAR` sekmesi bina kartlarını, `OLAYLAR` sekmesi aktif olaylar ile komşuları gösteriyor; olay satırı popup tıklaması, panel-local scrollbar ve wheel davranışı korunuyor. Seçili bölge değişince sekme ve scroll sıfırlanıyor. Kapsam: `internal/render/{panel.go,renderer.go,renderer_input.go,cursor.go}`, `wiki/{architecture/render-pipeline.md,systems/events.md}`; regression: `TestRegionPanelEventTabScrollAndSharedContentArea`, `TestRegionPanelTabsAndActiveEventRowHit`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Diplomasi hedef listesine devletlerin `Askeri güç` ve aktif devletler arasındaki `Güç sırası` değerleri eklendi. Liste sıralama düğmeleri artık `Alfabetik`, `İlişki` ve `Güç Sıralaması`; `İlişki` modu oyuncuyla olan ilişki puanını büyükten küçüğe sıralıyor, eşit puanlarda oyuncuyla kara sınırı paylaşan devletleri öne alıyor. Seçim aynı paneldeki liste sırasını değiştiriyor, focus/scroll yeni sıraya göre sıfırlanıyor. Kapsam: `internal/render/{diplom.go,renderer_input.go}`, `wiki/{systems/diplomacy.md,architecture/render-pipeline.md}`; regression: `TestDiplomacyFactionRelationSortPrefersAdjacentFactionOnScoreTie`, `TestDiplomacyListSortButtonsUpdateRendererState`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Recruit birim kartlarının alt etiket footer'ları üretilebilirlik durumuna göre renklendirildi: hazır kartlar yeşil, bina/teknoloji eksikleri amber/mavi, kaynak yetersizliği kırmızı; birim adı ve tur süresi metinleri siyah tutuluyor. Kuyruk kartlarının aktif/bekleyen ayrımı korunuyor. Kapsam: `internal/render/recruit_panel.go`, `wiki/architecture/render-pipeline.md`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Recruit ve ordu birim detay hover popup'larındaki görseller oran korunarak 50 px daha yüksek çiziliyor; genişlik ve popup metin alanı buna göre artırıldı. Kapsam: `internal/render/hover_tooltip.go`, `wiki/architecture/render-pipeline.md`; doğrulama: `go test ./...`.

- 2026-07-22: Bölge tıklamalarının askeri birim üretim panelini açması kaldırıldı; panel artık yalnızca alt HUD'daki `Ordu` butonuyla açılıyor. Bölge seçiminde açık panel kapanıyor ve bu davranış için regression testi eklendi. Kapsam: `internal/render/{renderer.go,renderer_input.go,renderer_input_test.go}`, `wiki/architecture/render-pipeline.md`; doğrulama: `go test ./...`.

- 2026-07-22: Askeri birim üretim paneli hover popup'ında maliyet satırları artık mevcut stok/gerekli miktar oranı yerine yalnız gerekli miktarı gösteriyor; kaynak yetersizse ilgili satır kırmızı `eksik` uyarısı taşıyor. Bina tooltip maliyetlerinin mevcut/gerekli formatı korunuyor. Kapsam: `internal/render/{hover_tooltip.go,recruit_panel_test.go}`, `wiki/architecture/render-pipeline.md`; doğrulama: `go test ./internal/render`.

- 2026-07-22: Üst-sol devlet HUD'unda Tahıl, Kereste, Demir ve Taş satırları mevcut stokla birlikte tur başı değişimi `+üretim/mevcut` formatında gösteriyor; negatif tahıl neti `-` işareti ve kırmızı renkle görünür. Üretim toplamı kuşatma altındaki bölgeleri dışarıda bırakan ortak state helper'larından hesaplanıyor. Kapsam: `internal/{state/state.go,render/panel.go}`, testler: `internal/{state/faction_production_test.go,render/resource_hud_test.go}`; doğrulama: `go test ./internal/state ./internal/render`.

- 2026-07-22: Bölge bina kartlarında kaynakları yeterli, maksimum seviyeye ulaşmamış ve kuyruğu boş olan binaların isim etiketi yeşil gösteriliyor; oyuncu inşa edebileceği seçenekleri kart ızgarasında doğrudan ayırt edebiliyor. Kapsam: `internal/render/building_card_component.go`, `wiki/architecture/render-pipeline.md`.

- 2026-07-22: Oyuncuya ait olmayan bölge seçildiğinde bina kartı hover popup'ı kapatıldı ve kaynaklar oyuncuda yeterli olsa bile bina etiketi altın sarısı tutuldu; bina hover, tıklama ve uygunluk görünümü ortak oyuncu sahipliği kontrolüne bağlandı. Regression: `TestNonOwnedBuildingCardsAreNotActionableOrHoverable`; doğrulama: `go test ./...`.

- 2026-07-22: Oyuncunun reddettiği AI barış/ittifak/ticaret/vassallık teklifleri artık aynı aktör-hedef-aksiyon için üç tur cooldown'a giriyor; sonrasında AI tur bazlı deterministik %35 retry zarı atıyor. Her normal ret ilişkiyi `-3` düşürüyor ve cooldown kaydı compact/debug/legacy save-load ile korunuyor. Savaş çağrısının mevcut ittifak bozma sonucu korunuyor. Kapsam: `internal/{state,diplomacy,ai,save}`, testler: `go test ./internal/diplomacy ./internal/ai ./internal/save`, `go test ./internal/game -run '^Test1300ScenarioGrainEconomyBands$'`.

- 2026-07-22: AI devletlerinin teknoloji tamamlanma mesajları oyuncu bildirimlerinden ayrıştırıldı. AI araştırması state içinde ilerlemeye devam ediyor; ancak AI teknolojileri kısa popup ve oyuncunun olay günlüğüne yazılmıyor. Kapsam: `internal/game/game.go`, `wiki/systems/tech-tree.md`; regression: `internal/game/technology_notification_test.go`.

- 2026-07-23: Oyuncu merchant filosu rota atama akışı tamamlandı. Seçili merchant filosunun ordu paneline `ROTA ATA` düğmesi eklendi; aktif ve filonun sahibiyle ilişkili rotalar modal listeden seçilebiliyor, görev kaldırılabiliyor ve yanlış denizdeki filo için bekleme durumu gösteriliyor. `TradeRouteKey` state doğrulaması AI'nin mevcut otomatik atama/hareket mantığıyla ortak tutuldu. Kapsam: `internal/{state/merchant_trade.go,game/game.go,render/{action.go,army_panel.go,merchant_route_panel.go,renderer.go,renderer_input.go}}`; testler: `go test ./internal/state -run 'TestMerchantTrade'`, `go test ./internal/render -run 'Test(ArmyPanel|MerchantRoute)'`.

- 2026-07-23: Nakliye filosu ordu paneli footer'ına taşıma doluluğu eklendi. Nakliye kapasitesi olan filolarda `Taşıma: dolu/kapasite` metni `Güç` bilgisinin solunda gösteriliyor; kapasitesiz filolarda alan boş kalıyor. Kapsam: `internal/render/army_panel.go`, `internal/army/army.go`; regression: `TestArmyPanelTransportFooterTextShowsLoadAndCapacity`, `TestArmyPanelTransportFooterTextHiddenWithoutTransportCapacity`.

- 2026-07-22: Tahıl ekonomisi yeniden dengeleme planı tamamlandı. Kasım ayındaki stabil rezerv fazlası nüfus yatırımına bağlandı; nüfus artışı sonraki tick'lerde sivil tahıl talebini artırıyor. Ordu morali save-backed `Army.Morale` alanına eklendi; stabil/uyarı/kritik/kıtlık seviyeleri sırasıyla `+1/-1/-3/-6` moral etkisi uyguluyor ve `TotalStrength()` üzerinden savaş/AI gücüne yansıyor. HUD/event detayında moral delta görünür; eski save'ler 100 başlangıç moraliyle uyumlu. Doğrulama: `go test ./...`.

- 2026-07-22: Tahıl ekonomisi Faz 5 tahıl ithalatı/stratejik talep alt adımı tamamlandı. Fraksiyonun üç aylık rezerv hedefine kalan açığı `StrategicDemand` olarak hesaplanıyor; tahıl piyasa fiyatına ek talep olarak yazılıyor, Pazar panelinde hedef bazında gösteriliyor ve yeni ticaret rotasında kaynakta rezerv fazlası varsa rota tahıla yönlendiriliyor. İthalat mevcut rota transferi üzerinden kaynak stok/altın koşullarıyla sınırlanıyor. Kapsam: `internal/{economy/economy.go,game/{game.go,resolution.go},state/state.go,diplomacy/diplomacy.go,render/trade.go}`; doğrulama: hedefli testler.

- 2026-07-22: Tahıl ekonomisi Faz 5 olay bağlama alt adımı tamamlandı. Kıtlık ve kötü hasat olayları aktif bölge süresince tahıl üretimini azaltıp sivil talebi artıran yüzde modifiyerleri taşıyor; kuraklık/hasat event adları aynı sözleşmeyle destekleniyor. Ekonomi tick'i, bölgesel ikmal, stratejik talep ve AI hesapları ortak `GameState` yardımcılarını kullanıyor. Mevcut düşman savaş gemisi tabanlı abluka ticaret ve liman ikmal kesintisiyle birlikte korunuyor. Kapsam: `internal/{events/events.go,state/state.go,game/resolution.go,ai/{ai.go,movement_strategy.go}}`, `assets/scenarios/*/data/events.json`; doğrulama: `go test ./internal/state ./internal/events ./internal/game ./internal/ai`.

- 2026-07-22: Tahıl ekonomisi Faz 6 denge ve kapanış adımı tamamlandı. 1300 senaryosu için 24 tur × 2 seed erken/orta/savaş tahıl raporu ve `1.0–4.0` üretim/sivil talep, `-1.0–2.5` net değişim/sivil talep kabul bantları eklendi; kıtlık oranı teşhis metriği olarak raporlanıyor. AI ve oyuncu lojistiği ortak `RegionMilitaryGrainProduction()` ve `EffectiveArmyGrainUpkeep()` seam'lerine geçirildi. Yeni tüketim modeli sonrası 42 tur × 4 seed altın bantları güncellendi; deterministik iki turluk replay ve `go test ./...` başarılı. Kapsam: `internal/{state/state.go,game/scenario_balance_test.go,ai/grain_economy_test.go}`, `wiki/{systems/{economy,ai}.md,architecture/state-management.md}`.

- 2026-07-23: 1300 tahıl denge testi düzeltildi. Üç seviyeye kadar üst üste uygulanan `farm` çarpanı `x1.65`ten `x1.30`a indirildi; Osmanlı savaş fazındaki `7.89` üretim/sivil talep oranı `3.41`e düşerek `1.0–4.0` bandına girdi. Doğrulama: `go test ./internal/game -run '^Test1300ScenarioGrainEconomyBands$'`.

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
- 2026-08-06: AI objective'lerinden `annex_region_ids` kaldırıldı. Vassallık artık
  `allow_vassalization` uygunluk kontrollerini geçen fetihte yüzde 50 deterministik
  zarla kararlaştırılıyor; zar başarısızsa normal ilhak yapılıyor. Böylece bölge
  stratejisi `territorial_claims` altında, savaş sonrası siyasi sonuç ise runtime
  kararında tek akıştan yönetiliyor.

- 2026-07-19: `1300_ottoman_rise` için veri güdümlü Osmanlı/Doğu Roma AI objective
  dikey dilimi eklendi. `data/ai_strategies.json`; Bitinya, Anadolu beylikleri, Ankara
  koridoru, Trakya/Konstantinopolis, 1501 sonrası Safevi rekabeti ve Doğu Roma Boğaz
  savunması/Bitinya geri alma yönlerini tanımlıyor. Tarihsel hedefler mevcut güç ve
  frontier kontrollerini atlamayan soft savaş/hareket bonusları verir; yalnız geç
  hedefler yıl/event hard gate'i kullanır. Anadolu beyliklerinde son toprak sonrası
  sonuç hibrittir: uygunluk kontrollerini geçen hedef fetih anındaki yüzde 50 zarla
  vassal kalabilir; zar başarısızsa doğrudan ilhak edilir. Politika açık arazi,
  savaşsız işgal, çıkarma, genel hücum ve kuşatma tesliminde ortaktır. Yeni
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
| Kuşatma sistemi | ✅ | Tahkimli kara bölgeler (`fortress` settlement veya `walls`) artık savaşsız anında düşmez. Normal kara orduları da kuşatma kurabilir; kuşatma birimi olmayan orduda `Genel Hücum` kapalıdır ve kale yalnız aç bırakma / teslimiyet süreciyle düşebilir. Sağ tık akışı savaş ilanından sonra `Kuşatma Kararı` modalına bağlanır; `Kuşatma Başlat`, `Genel Hücum` (uygun ordularda), aktif kuşatmada `Kuşatmayı Kaldır` seçenekleri vardır. Kuşatma durumu `GameState.Sieges` içinde kaydedilir, kuşatan ordu render'da kuşatılan bölgenin üstünde görünür ve bölge kuşatanın rengiyle taralı muallak overlay alır. Seçili kuşatma ordusunda alt-orta `Kuşatma Emri` paneli ve ordu üstünde kılıç rozeti görünür; başka komşu bölgeye verilen normal hareket emri eski kuşatmayı otomatik kaldırır. Gedik açılması artık kuşatma birimi adedi, canlı HP, ekipman tier'i ve kale seviyesi birlikte hesaplanır; T1 mancınık T3, T2 bombarda T4, T3 top T5 sura kadar etkilidir, uyumsuz düşük tier araçlar gedik gücüne katkı vermez. Genel hücum artık gedik yokken tahkimatı doğrudan ele geçiremez; ayrıca saldıran taraf gedik küçüldükçe daha ağır hücum zayiatı öder. Aktif kuşatmaya yalnız aynı fraksiyon ya da müttefik devlet destek için katılabilir; ilgisiz üçüncü devletler ikinci bir kuşatma hamlesi yapamaz. Kuşatma altındaki bölgeye üçüncü devlet yeni kuşatma başlatamaz; fakat bölgeye giriş hakkı olan bir ordu, kuşatma yapan düşman orduyu savaşta yenerek kuşatmayı kaldırabilir. AI de tahkimli hedeflerde doğrudan fetih yerine kuşatma açar |
| Deniz taşıma akışı | ✅ | Kara ordusu uygun `transport` filosuna binebilir, filo `EmbarkedUnits` ile taşır, komşu dost/boş karaya çıkarma yapılır; oyuncu ve AI aynı kural setini kullanır. Sahipsiz kıyıya amfibi çıkarmada bölge artık otomatik sahiplenilir; bug'lı eski save'lerde sahipsiz kalmış ama tek taraflı işgal altında duran kara bölgeleri yükleme/tur çözümlemesinde toparlanır. `transport` birimleri artık `carry_capacity` ile gerçek slot kapasitesi taşır, aynı filo kapasite yettikçe birden fazla kara birimini biriktirebilir ve oyuncu seçili orduyla dost nakliye filosuna sağ tıklayınca `Gemiye Bin` onayı alır. Gemide birlik taşıyan filo limana uğramadan 3 turdan fazla açık denizde kalırsa taşınan birlikler her tur artan zayiat alır |
| Boğaz deniz geçiş sürekliliği | ✅ | Senaryo verilerinde Marmara-Ege-Karadeniz deniz komşuluğu çift yönlü korunur; filolar `Ege -> Marmara -> Karadeniz` ve ters yönde komşuluk bazlı ilerleyebilir, bu köprü testi `internal/world/scenario_sea_adjacency_test.go` ile sabitlenmiştir |
| Amfibi savaş fazı | ✅ | Düşman kıyıya çıkarma savaş halinde aktif; çıkarma anı çatışması `combat` ile çözülür, başarılı çıkarma karaya ordu indirip sahiplik günceller, AI barışta çıkarma denemez |
| Başlangıç orduları | ✅ | Her senaryonun `data/armies.json` dosyasından yükleniyor |
| Çarpışma motoru | ✅ | Birim gücü, arazi, teknoloji modları ve rastgele sonuç etkisi; saldırı duruşu (`agresif/dengeli/savunmacı`) gerçek savaş sonucu ve saldırı öncesi preview hesabında aynı combat helper'larıyla işlenir. `land/naval/amphibious` bağlamları ayrı stance çarpanları ve açıklama metinleri taşır; muhtemel kayıp paneli bu bağlama göre hesaplanır |
| Komutan kariyeri | ✅ | `Army.Commander` çekirdeği, dengelenmiş XP/seviye/trait ilerlemesi, savaş gücüne saldırı-savunma etkisi, save/load, üç kişilik oyuncu havuzu, oyuncunun isimli-rastgele XP aday üretimi ve ordu panelinden atama/ayırma, `500 altın + 100 tahıl` maliyetli AI/oyuncu runtime komutan üretimi, birleşme-garnizon yaşam döngüsü ve savaş raporu/olay günlüğü kariyer bildirimi hazır |
| Savaş sonrası toparlanma | ✅ | Savaş, lojistik ve diğer HP kayıpları artık kısmi hasar bırakır; kara orduları kendi, müttefik veya vassal realm toprağında tur başına `+10 HP` ile %100'e kadar toparlanır, limana bağlı donanmalar da kendi, müttefik veya vassal limanında onarım alır; birim kartındaki `+` rozeti aynı ortak uygunluk kuralını kullanır |
| Ordu detay paneli | ✅ | 20 slot, HP/deneyim çubukları, bölme/birleştirme aksiyonları, dost toprakta toparlanan birimler için küçük `+` rozeti |
| Ordu birleşme | ✅ | Dost bölgede yalnızca panelden manuel birleşme, 20 birim limiti; hareket orduları otomatik birleşmez |
| Ordu bölme | ✅ | Seçili orduyu iki parçaya böler |
| Rakip ordu istihbaratı | ✅ | Menzildeki rakip orduda sayı ve yarım birim listesi görünür; menzil dışı detaylar gizlenir; emir verilemez |
| Çoklu ordu render | ✅ | Aynı bölgede ordular yan yana çizilir |
| Askeri kapasite | ✅ | Ordu sayısı `ceil(kara_bölge/2) + 1`, savaşçı sınırı temel `ceil(kara_bölge/2) × 20`; +1 ordu slotu savaşçı kapasitesine eklenmez; scenario/save kökenli `garrison` orduları artık bu saha ordusu limitine sayılmaz, hareket/split/merge ile sahaya çıktıklarında normal orduya dönüşür |
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
| Ticaret güzergahları | ✅ | `TradeRoutes` pasif gelir modeli var; AI merchant filoları rotaya otomatik atanıyor, oyuncu merchant filosu seçili ordu panelindeki `ROTA ATA` modalından aktif rotayı seçip görevi kaldırabiliyor |
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
| Çoklu yerleşim noktaları | ✅ | `regions.json` içinde `settlements[]`; ana yerleşim ordu/etiket anchor'ı, zoom LOD'una göre uzak görünümde başkent/kale, orta görünümde liman/şehir, yakın görünümde kasaba dahil tüm yerleşim noktaları/isimleri, görünür marker ile ortak hit-test, bölge dışı koordinatta log + nearest-region fallback; `port` settlement'lar liman simgesi, `fortress` settlement'lar kale simgesiyle ayrışır |
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

## Son AI savaş/diplomasi ilerlemesi

- ✅ Aktif savaşlara stratejik hedef, hedef kilidi, seferberlik/lojistik rezervi ve
  müttefik-vassal ortak cephe katkısı eklendi.
- ✅ Barış sonrası altı turluk ateşkes ve savaş yorgunluğu/ekonomik baskı metrikleri
  eklendi.
- ✅ AI-AI barışları beyaz barış, bölge bırakma, tazminat veya vassallık sonucuna
  ayrılabiliyor; oyuncu barışı seçim yapılmadan güvenli beyaz barış olarak kalıyor.
- ✅ Geliştirme modunda `F3` AI teşhis modalı plan, cephe, hedef, roller ve
  bloklanma nedenlerini gösteriyor; aynı snapshot debug save sidecar'ına yazılıyor.

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
- 2026-07-22: Kuşatma teslimiyeti iki yönlü diplomasi teklifine bağlandı. AI kuşatan baskı yeterliyse oyuncuya bölge kimliği taşıyan teslim olma talebi gönderebiliyor; AI savunmacı da ağır kuşatmada oyuncuya teslimiyet teklif edebiliyor. Teklif modalı `Kabul Et` etiketi kullanıyor, mesaj alanı genişletilmiş dikey panelde daha fazla satır gösteriyor ve bölge bağlı tekliflerde kamera `RegionID` üzerinden kuşatılan alana odaklanıyor. Savunma kuşatma panelindeki `Teslim ol` yalnız gerçek teklif geldiğinde aktifleşiyor. Kabulde savunma orduları mümkünse en yakın dost bölgeye moral kaybıyla çekiliyor; AI'ın son toprağı için oyuncu kabulü doğrudan vassallık kuruyor. Kapsam: `internal/{state/state.go,diplomacy/{diplomacy.go,offers.go,quota.go},ai/diplomacy.go,game/game.go,render/{renderer_dialogs.go,renderer_input.go,ui_modals.go}}`, testler: `internal/{ai/siege_test.go,game/siege_test.go,render/war_confirm_test.go}`.
- 2026-07-22: Saldıran kuşatma paneline `Teslimiyet Teklifi` düğmesi eklendi. Oyuncu aktif kuşatmadan AI savunmacıya teklif gönderebiliyor; AI baskı, gedik, süre ve güç dengesine göre aynı tur kabul veya ret veriyor. Kabul mevcut teslimiyet/fetih/vassallık çözümleyicisine, ret diplomasi geçmişi ve kota akışına bağlandı. Üçlü panel düğmeleri çakışmayacak şekilde daraltıldı. Kapsam: `internal/{render/{action.go,renderer.go,renderer_dialogs.go,renderer_input.go},game/game.go}`, testler: `internal/{game/siege_test.go,render/war_confirm_test.go}`.
- 2026-07-22: Genel hücum için kuşatma birimi zorunluluğu kaldırıldı. Kuşatma birimi olmayan kara ordusu da kuşatma kararından veya aktif kuşatma panelinden `Genel Hücum` başlatabilir; gedik yokken tahkimatın doğrudan ele geçirilmesini engelleyen mevcut gedik kuralı korunur. AI aktif kuşatmada aynı şekilde genel hücum çözebilir. Kapsam: `internal/{game/{game.go,siege.go},ai/{ai.go,movement_strategy.go},render/{renderer_dialogs.go,renderer_input.go}}`, testler: `internal/{game/siege_test.go,render/war_confirm_test.go}`.
- 2026-07-22: Kuşatma teslimiyet teklifleri normal diplomasi elçi kotasından çıkarıldı. Oyuncu ve AI, elçi hakkı dolu olsa bile geçerli kuşatmalar için teslimiyet teklifi gönderebilir; aynı bölge tekrar teklif ve bölge bazlı ret cooldown kuralları korunur.
- 2026-07-22: Reddedilen kuşatma teslimiyet teklifi aynı turda aynı bölgeye yeniden gönderilemiyor. Saldıran kuşatma panelindeki `Teslimiyet Teklifi` düğmesi yalnız ilgili bölge için pasifleşiyor; diğer bölgeler ve sonraki turdaki cooldown akışı etkilenmiyor. Regression: `TestRejectedSurrenderOfferCannotRepeatInSameRegionThisTurn`, `TestSelectedSiegeSurrenderOfferDisabledAfterSameTurnRejection`.
- 2026-08-01: Taşınan ordu biriminin sağ üst karesi, ticaret görevinin `+2` rozetindeki iki katmanlı ince border stiline uyarlandı. Koyu dış zemin, yaklaşık 1 px altın iç kenar ve koyu küçük metin kullanılıyor; filo gemi sayısı da ana dairesindeki yerini koruyor. Regression: `go test ./internal/render -count=1`.
- 2026-08-01: Taşınan ordu karesi tıklanabilir yapıldı. Tıklama filoyu mekanik olarak seçili tutup aynı geometriyi kullanan taşınan kara ordusu bilgi panelini açıyor; panel yalnız bilgi gösteriyor. Rozetin içi siyah, metni beyaz ve border'ı ince sarı olarak güncellendi. Regression: `TestEmbarkedArmyBadgeHitSelectsTransportedArmyView`, `go test ./internal/render -count=1`.
- 2026-08-01: Filo panelindeki `Taşıma: X/Y` kapasite metni ortak buton yüzeyine taşındı. Tıklama taşınan kara ordusu bilgi panelini açıyor; çizim, hit-test ve cursor aynı footer geometrisini kullanıyor. Regression: `TestArmyTransportFooterTextUsesInfoButtonGeometry`, `go test ./internal/render -count=1`.
- 2026-08-01: Taşıma butonunun görünümü düzeltildi. `gameui.Button` ortak dikey merkezleme hesabı etkinleştirildi, sabit text offset kaldırıldı ve etiket çevresine 10 px yatay padding eklendi. Regression: `TestArmyTransportFooterTextUsesInfoButtonGeometry`, `go test ./internal/render -count=1`.
- 2026-08-01: Taşınan birlik içeren filolarda `Taşıma: X/Y` butonu yeşil aktif taşıma rengine alındı. Footer ile buton arasında üst-alt 4 px margin bırakıldı; boşluk merchant/görev footer butonlarıyla hizalandı. Regression: `TestArmyTransportFooterTextUsesInfoButtonGeometry`, `go test ./internal/render -count=1`.
- 2026-08-01: Filonun taşıdığı kara ordusu rozeti 16 px'e büyütüldü; çift siyah/sarı katman yerine yalnız sarı border ve siyah iç dolgu kullanılıyor. Birlik sayısı outlinesız beyaz metin olarak kalıyor. Regression: `TestNavalEmbarkedArmyBadgeStaysAtUpperRight`, `TestEmbarkedArmyBadgeHitSelectsTransportedArmyView`, `go test ./internal/render -count=1`.
- 2026-08-06: Yan yana ordu marker'larının ortak merkez aralığı 40 px'e çıkarıldı; komutan portreleri artık bitişmiyor, üst görev/bonus rozetleri komşu portrelerin üzerine taşmıyor. Kuşatma, donanma ve farklı grupların yeniden dağıtım geometrileri aynı adımı kullanıyor. Regression: `TestArmyIconSpacingSeparatesCommanderPortraitsAndBadges`, `go test ./internal/render -run 'TestArmyIcon|TestArmyCommanderBadge|TestArmySelectionIndicator|TestNavalArmyBadges|TestNavalDamageBadge|TestArmyDamageBadge' -count=1`.
