---
type: dev
tags: [progress, status, todo, known-issues, next-steps]
last_updated: 2026-06-12
related: [HOME, architecture/game-loop, architecture/state-management, architecture/render-pipeline, systems/victory]
---

# Geliştirme Durumu

## Denetim Özeti (2026-05-08)

Proje artık oynanabilir dikey kesite yakın: ana menüden senaryo seçiliyor, fraksiyon ve zafer koşulu seçilip kampanya başlıyor, tur döngüsü çalışıyor, AI turu işleniyor, harita ve paneller render ediliyor, kayıt/yükleme slotları var.

Mevcut veri seti iki senaryoda da aynı genişlikte:

| Senaryo | Bölge | Deniz | Fraksiyon | Oynanabilir | Başlangıç ordusu |
|---|---:|---:|---:|---:|---:|
| `1300_ottoman_rise` | 210 | 52 | 45 | 30 | 49 |
| `1444_constantinople` | 210 | 52 | 45 | 30 | 49 |

Doğrulama: `go test ./...` WSL ortamında 2026-05-08 tarihinde başarıyla çalıştı.

## Tamamlanan Sistemler

| Sistem | Durum | Notlar |
|---|---|---|
| Ebitengine kurulum | ✅ | `cmd/game/main.go`, 60 TPS |
| GameState merkezi yapı | ✅ | `internal/state/state.go` |
| Phase state machine | ✅ | 12 faz: ana menü, ayarlar, senaryo, fraksiyon, zafer, oyun, AI, çözümleme, game over, pause, load, save |
| Senaryo sistemi | ✅ | `internal/scenario/scenario.go`; `assets/scenarios/scenarios.json` index + bağımsız senaryo klasörleri |
| Senaryo seçim ekranı | ✅ | `internal/render/scenario_select.go`, `PhaseScenarioSelect` |
| Harita render | ✅ | `WorldMap` cache, ülke/deniz şekilleri, sahiplik rengi, seçili bölge vurgusu |
| Senaryo bazlı harita hizalama | ✅ | `scenario.json` içindeki `map` alanı `WorldW/WorldH` ve shape offset/scale değerlerini belirler |
| Görsel mevsim değişimi | ✅ | `internal/render/mapgen.go:applyOwnership`; kış/ilkbahar/sonbahar tint |
| Bölge sistemi | ✅ | JSON'dan yükleme, komşuluk grafı, kilitli bölge alanları |
| Fraksiyon sistemi | ✅ | 45 fraksiyon, 30 oynanabilir, renk/din/kaynaklar |
| Din paketi | ✅ | `internal/religion`; `catholic`, `orthodox`, `sunni`, `shia` ilişki puanları |
| Ordu hareketi | ✅ | Komşuluk kısıtı, kara/deniz giriş kontrolü, savaş öncesi diplomasi kontrolü; donanmalar deniz bölgeleri arasında savaş ilanı olmadan dolaşır, deniz çatışması sadece `StanceWar` durumunda tetiklenir; AI dost bölgelerde bölgesel ikmal baskısını okuyup aşırı dolu kara bölgelerden daha rahat komşu bölgelere dağılabilir |
| Deniz taşıma akışı | ✅ | Kara ordusu uygun `transport` filosuna binebilir, filo `EmbarkedUnits` ile taşır, komşu dost/boş karaya çıkarma yapılır; oyuncu ve AI aynı kural setini kullanır |
| Boğaz deniz geçiş sürekliliği | ✅ | Senaryo verilerinde Marmara-Ege-Karadeniz deniz komşuluğu çift yönlü korunur; filolar `Ege -> Marmara -> Karadeniz` ve ters yönde komşuluk bazlı ilerleyebilir, bu köprü testi `internal/world/scenario_sea_adjacency_test.go` ile sabitlenmiştir |
| Amfibi savaş fazı | ✅ | Düşman kıyıya çıkarma savaş halinde aktif; çıkarma anı çatışması `combat` ile çözülür, başarılı çıkarma karaya ordu indirip sahiplik günceller, AI barışta çıkarma denemez |
| Başlangıç orduları | ✅ | Her senaryonun `data/armies.json` dosyasından yükleniyor |
| Çarpışma motoru | ✅ | Birim gücü, arazi, teknoloji modları ve rastgele sonuç etkisi |
| Savaş sonrası toparlanma | ✅ | Savaş, lojistik ve diğer HP kayıpları artık kısmi hasar bırakır; kara orduları kendi kara toprağında tur başına `+10 HP` ile %100'e kadar toparlanır, limana bağlı donanmalar da kendi veya müttefik limanında aynı hızla onarım alır |
| Ordu detay paneli | ✅ | 20 slot, HP/deneyim çubukları, bölme/birleştirme aksiyonları, dost toprakta toparlanan birimler için küçük `+` rozeti |
| Ordu birleşme | ✅ | Dost bölgede otomatik veya panelden manuel birleşme, 20 birim limiti |
| Ordu bölme | ✅ | Seçili orduyu iki parçaya böler |
| Rakip ordu istihbaratı | ✅ | Menzildeki rakip orduda sayı ve yarım birim listesi görünür; menzil dışı detaylar gizlenir; emir verilemez |
| Çoklu ordu render | ✅ | Aynı bölgede ordular yan yana çizilir |
| Askeri kapasite | ✅ | Kara bölgesi başı 5 + kışla başı 5; ordu sayısı `ceil(kara_bölge/2)` |
| Asker alma | ✅ | Milis hızlı alım + belirli birim alımı; bina/teknoloji/çoklu kaynak/manpower kontrolü; JSON `turns_required` ile üretim kuyruğunda tamamlanır, tekrar tıklanınca iptal edilip kaynaklar iade edilir; aynı bölgede mevcut ordu 20/20 doluysa üretim artık bloke olmaz, tamamlanan birlikler boş slotlu orduya eklenir veya gerekirse ikinci kara ordusu olarak spawn olur |
| Çoklu eğitim kuyruğu (Total War benzeri) | ✅ | Recruit panelinde birim bazında `- xN +` seçimi, kuyrukta aynı birim için ilk tamamlanma turu görünürlüğü ve tek tıkta çoklu (`xN`) üretim emri; bölgesel kapasite `max(1,population/100)+kışla` kuralıyla sınırlandırılır |
| Bina/birim hover bilgisi | ✅ | Kart tooltipleri maliyet, gereksinim, etki/istatistik ve görsel gösterir |
| Deniz birimi | ✅ | Liman ve kıyı koşuluyla filo/deniz birimi üretimini kuyruğa alma; tekrar tıklanınca iptal/iade |
| Liman docking akışı | ✅ | Limana bağlı donanma aynı deniz bölgesine hareket emri aldığında liman bağını bırakıp deniz merkezine çıkar; ayrıca komşu denizden sahibi olunan veya müttefik olunan port settlement içeren kara bölgesine dock olabilir, deniz `RegionID` korunur ve hareket puanı tüketir |
| Ekonomi tick | ✅ | Vergi geliri, hasat modu, bina modları, ikincil mallar, taş üretimi, tahıl bakım gideri ve tahıl açığında lojistik HP cezası; ayrıca bölge bazlı ikmal kapasitesi (efektif tahıl + yerleşim tamponu + sınırlı stok desteği) aşılınca aynı bölgede uzun süre yığılan kara orduları kademeli zayiat alır |
| Vergi ayarlama | ✅ | Oyuncu bölgelerinde `.` / `,` ile ±5 |
| Bina inşası | ✅ | JSON bina tipleri, maliyet, arazi ve adet kısıtları; varsayılan 2 turluk üretim kuyruğu; kuyruktaki bina tekrar tıklanınca iptal/iade; liman için kanonik uygunluk kuralı artık literal `terrain=coast` değil, deniz komşuluğunu okuyan `Region.IsCoastal` predicate'i |
| Kaynak reçete sistemi | ✅ | Birim ve bina üretiminde `grain/iron/timber/stone` tüketimi; UI maliyet satırı ve AI kararları bu modele bağlı |
| Bina seviye sistemi | ✅ | Binalar `max_per_region` kadar seviye alır (Lv1..LvN); panelde `Lv` ve kuyruk adedi görünür, inşa mesajları seviye geçişini (`LvX→LvY`) gösterir; kurulu bina kartları da tıklanabildiği için yükseltme/iptal akışı doğrudan kart üzerinden çalışır; manpower ve üretim kapasitesi kışla seviyesiyle artar |
| Ticaret güzergahları | ✅ | `TradeRoutes` pasif gelir modeli var |
| Teknoloji ağacı | ✅ | Araştırma başlatma, tur sayacı, tamamlanan teknoloji efektleri, ağaç görünümü, seviye bazlı düzen, kategori renkleri, tamamlanmış teknoloji tick badge'leri, araştırma seçimi/değiştirme/vazgeçme, HUD'da aktif araştırma gösterimi, tur bitir uyarısı ve tamamlanma mesajları event loguna ekleniyor; tur bitir araştırma uyarısı artık yalnız gerçekten araştırılabilir teknoloji kaldığında gösteriliyor; yarım bırakılan araştırmalar pause/resume ile kaldığı yerden sürüyor; 1300 senaryosunun research ağı yeni orta/ileri düğümlerle genişletildi ve daha önce boştaki `market_gold_mod`, `peace_relation_bonus`, `naval_move_bonus`, `reveal_enemy_strength`, `conversion_speed_mod` alanları runtime'a bağlandı; bölge panelindeki sahip devlet adına tıklanınca rakip devletin aktif araştırması, tamamlanan teknolojileri, malları ve ticaret özeti ayrı panelde görülebiliyor |
| Diplomasi | ✅ | `internal/diplomacy` ortak motoru ile savaş/barış/ittifak/ticaret; deterministik kabul-red, ilişki decay'i ve ticaret rotası senkronu |
| Diplomasi paneli modern akış | ✅ | Solda devlet seçimi + sağda teklif paneli; savaş/barış/ittifak/ticaret için muallak kabul olasılığı (%) ve durum göstergesi bulunur; teklif sayfasında üst bilgi blokları padding'li, footer icon+metin hizası ortak button primitive'iyle tutarlı çizilir |
| Elenen fraksiyon diplomasi temizliği | ✅ | Kara toprağı biten fraksiyonlar (sadece deniz bölgesi kalsa bile) elendiğinde tüm diplomasi ilişkileri, bekleyen `diplomatic_offers` kayıtları ve ticaret rotaları temizlenir; ayrıca save/load sonrası relation dışı veya elenmiş fraksiyona bağlı stale trade rotaları sanitize edilerek trade paneline geri sızmaları engellenir |
| Liman işgalinde donanma tahliyesi | ✅ | Bölge el değiştirince, ele geçirilen limana bağlı eski sahip filoları otomatik limandan çıkarılır ve en yakın deniz bölgesine bırakılır; ancak fetih bir fraksiyonun son kara toprağını da düşürürse kalan kara orduları ve donanmaları galip devlete devrolur ve yıkılış mesajı event log'a yazılır |
| Oyuncuya gelen diplomasi teklif paneli | ✅ | AI barış teklifleri `diplomatic_offers` kuyruğuna düşer; oyuncu modal anlaşma panelinden kabul/red verir, kabulde standart diplomasi motoru uygulanır |
| Din diplomasisi | ✅ | Başlangıç ilişkileri din puanıyla kuruluyor; Sünni-Şii savaş başlıyor |
| Din dönüşümü | ✅ | Ele geçirilen bölgede 24 tur sonra yeni sahip dinine dönüşüm, memnuniyet -20 |
| Tarihsel olaylar | ✅ | JSON tetikleyici, tek seferlik olay işleme; tarihsel modal içinde A/B kararları, choice prompt, ekonomi/diplomasi/ordu etkisi ve ayrı karar log kaydı; follow-up zincirler flag, bölge sahipliği, teknoloji ve diplomasi stance/score koşullarına bağlanabiliyor; event popup ve event detail log görünümü follow-up/koşul özetini gösteriyor; event log `Kodex` düğmesi pending historical zincirleri `Hazır/Takvim/Kilitli` statüsüyle filtreli açıyor, solda kısa özetli ve scroll'lu liste sağda detay kartı gösteriyor, event başına kalan ay + kritik eksik koşulu sunuyor ve liste artık modal dışına taşmıyor; tarihsel popup artık draw/input/cursor tarafında gerçek üst modal olarak işlendiği için altta bekleyen teklif/onay diyalogları choice butonlarını kilitlemiyor |
| Zafer koşulları | ✅ | `domination`, `economic`, `military`, `religious`, `conquer_city`, `survive_turns` kontrol ediliyor; senaryo hedefleri `allowed_factions` ile oyuncu fraksiyonuna göre filtreleniyor, tam liste `ScenarioVictories` içinde tutuluyor; zafer kartları `deadline_year/month` ile oyuncuya süre veriyor ve AI hedef tamamladığı için otomatik kazanamıyor |
| AI turu | ✅ | Teknoloji, ekonomi, deniz, asker alma, konsolidasyon, diplomasi taraması ve hedefe hareket; deniz hedefleri `aiSeaPressure()` ile savaş baskısına göre seçilir, filo limiti kıyı/savaş durumuna göre 1-3 arası dinamikleşir |
| AI uzun menzilli hareket | ✅ | BFS ile uzaktaki hedefe doğru ilerleme |
| AI koalisyon | ✅ | Zorluk 3'te oyuncu 8+ bölgeyi geçince devreye girer |
| Kayıt/yükleme | ✅ | Autosave + QuickSave + slot1-3, metadata önizleme, silme; tur bitirde autosave, oyun içi kaydetmede quicksave; save/load kartında silme onayı başlıkla çakışmayacak ayrı satır düzeni kullanır |
| Yükleme ekranı | ✅ | Senaryo ve kayıt yükleme sırasında gerçek zaman tabanlı hareketli spinner gösteriliyor; iş yükü ilk loading frame çizildikten sonra başlatıldığı ve yükleme adımları scheduler'a yield verdiği için loader animasyonu senaryo okunurken donmuyor; yükleme ekranı artık step-bazlı yüzde ve progress bar da gösteriyor; senaryo yüklemesi `faction_select`/`victory_select`e gidiyorsa ağır `WorldMap` cache'i loader bitiminde değil, harita ilk gerçekten gerektiğinde kuruluyor; zafer koşulu sonrası oyuncu turuna geçerken ve save load akışında `WorldMap` hazırlığı arka planda yapılıp loading ekranı altında tamamlanıyor |
| Ana menü / ayarlar | ✅ | Yeni oyun, autosave varsa devam et, kayıt yükleme, ayarlar, çıkış |
| Pause menüsü | ✅ | ESC ile açılır; devam, kaydet, yükle, ana menü, çıkış |
| Fare odaklı UI akışı | ✅ | Menü geri düğmeleri, teknoloji/diplomasi X kapatma, bölge/ordu panel kapatma, vergi/bina/asker aksiyonları fareyle yapılabilir |
| Olay paneli | ✅ | Sağ üst olay paneli daha fazla kayıt tutar, uzun liste mouse wheel ile kaydırılır |
| Minimap | ✅ | Sağ alt köşe, kamera ve ordu konumları |
| Üst-sol durum paneli | ✅ | Fraksiyon, kaynak, zafer ve ordu özeti haritanın üst-solunda ayrı panel; zafer/askeri özet kompakt iç kartlarla taşmadan çizilir, zafer HUD kartına tıklanınca merkezde detay popup açılır ve popup içinde sahip olunan/eksik hedefler checklist olarak gösterilir |
| Sağ-üst tarih/menü HUD | ✅ | Tarih, mevsim, tur ve duraklama menüsü butonu sağ üstte ayrı panel |
| Alt-orta aksiyon HUD | ✅ | Diplomasi, Teknoloji ve Tur Bitir butonları ayrı HUD içinde alt-ortada |
| Olay logu akordiyonu | ✅ | Panel daraltılıp genişletilir; uzun metinler wrap edilir; kartlar X ile kapanır, tıklanınca detay popup açılır |
| Info popup bildirimi | ✅ | Altın yetersiz gibi oyun içi uyarılar olay loguna yazılmaz, ayrı geçici popup olarak görünür |
| Kompakt UI taşma düzeltmeleri | ✅ | Genel onay modalı mesaj wrap eder; bölge panelinde üretim alanı artık yalnız Tahıl ile sınırlı değildir, efektif `Altın/Tahıl/Demir/Kereste/Taş/Baharat/Kumaş` satırları iki kolon grid halinde çizilir; sahip bilgisi başlığın hemen altında etiketsiz görünür; memnuniyet/vergi satırlarında yüzde metni, progress bar ve vergi `-/+` düğmeleri artık birbirine taşmaz; maksimum seviyedeki bina kartlarında alt satırdaki `Maks` yazısı kaldırılır ve uyarı durumu sol üst `Lv` rozetinin kırmızı arka planıyla verilir; kuyruktaki bina kartlarının `N Tur` göstergesi kontrastlı pill rozetine taşındığı için açık sprite üstünde kaybolmaz |
| Panel cursor hit-test düzeltmesi | ✅ | Sol alt bölge paneli, olay logu, alt HUD, kayıt slotları ve onay panellerinde parmak imleci sadece gerçek tıklanabilir alanlarda gösterilir |
| Ordu/yerleşim tıklama önceliği | ✅ | Aynı pikselde çakışan inputta seçim sırası artık `ordu/donanma etiketi > yerleşim > bölge`; hover hit-test de aynı helper ile eşlendi |
| UI framework migrasyonu (çekirdek + ekranlar) | ✅ | `internal/ui` altında `Widget`, `InputState`, `Manager`, ortak `Panel`, `Label`, `TextBox`, `Image/Icon`, `Tooltip`, `Button`, `Dropdown`, `ListView`, `Checkbox`, `RadioGroup`, `Modal`, `Overlay` ve layout yardımcıları eklendi; trade, diplomasi, teknoloji, pause/save-load, ana menü ve seçim ekranları, HUD küçük etkileşim yüzeyleri, recruit panel hit-test ailesi, ordu split/merge overlay'i, edit mode inspector/form yüzeyleri, shape yardım/pixel preview overlay'leri, diplomasi teklif diyaloğu ve confirm/war/event detail/historical modal aileleri ortak UI builder akışına taşındı; tema tokenları `internal/render/ui_theme.go` altında merkezileştirildi, ana menü/senaryo/fraksiyon/zafer/pause/kayıt slotlarında `Manager` tab focus kullanılmaya başlandı, seçim ekranı metinleri ortak `Label + TextRenderer` primitive'ine bağlandı, headless geometri + draw-call smoke ve allocation testleri eklendi; `ListView` artık press/release ayrımı ve drag threshold taşıyor, böylece diplomasi gibi uzun listelerde sürükleme scroll'u yanlış seçim üretmiyor; teknoloji panelinde kart görünümü ve hit-test akışı ayrıca `techCardComponent` seam'ine ayrıldı, çizim ve tıklama aynı rect/projection üzerinden yürütülüyor |
| Ortak ikon buton katmanı | ✅ | `internal/ui.Button` opsiyonel `IconID` desteği aldı; gerçek bitmap ikonlar `assets/ui/icons/*.png` altından yüklenip cache'leniyor, aynı primitive close/back/menu/kodex/müzik/diplomasi gönder yanında trade al/sat, savaş/diplomasi onayları, save/delete mini aksiyonları ve confirm dialog butonlarında tekrar kullanılıyor |
| Ses ve müzik | ✅ | `assets/sounds` global efektleri; senaryo `musics/` playlistleri `scenario.json` `music` alanından; ayarlarda ayrı müzik/ses seviyeleri; oyun içi müzik HUD'u ve ESC menüsü müzik kontrolleri |
| Development mode | ✅ | `DEV_MODE=true` ile `GameState.DevelopmentMode` |
| Render başlangıç log temizliği | ✅ | Boş senaryo path'inde shape dosyası okunmaz; deniz seed araması ham `world_x/world_y` fallback kullanır |
| Açılış kamera zoom ayarı | ✅ | İlk frame `resetCamera()` minimum sığdırma yerine `1.12x` yakın başlar; alttaki siyah boşluk azalır, kullanıcı isterse tekerlekle tekrar minimum zoom'a dönebilir |
| Deniz anchor ve çakışma stabilizasyonu | ✅ | Deniz orduları gerçek su piksel anchor'ına çizilir; ordu/etiket çizim sırası deterministik, çakışan etiket metinleri bastırılır |
| Çoklu yerleşim noktaları | ✅ | `regions.json` içinde `settlements[]`; ana yerleşim ordu/etiket anchor'ı, yakın zoom'da ek yerleşim noktaları/isimleri, bölge dışı koordinatta log + nearest-region fallback |
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
| Edit mode shape paint editor | ✅ | Inspector `Shape` sekmesi seçili kara region'ın `shape_id` verisini sağ mouse drag ile boya/sil düzenler; stroke sırasında yeşil/kırmızı canlı preview overlay ve yardım paneli görünür; stroke bitince mask contour'ları yeniden ring'e çevrilir, `ShapeData` + `Region.Shape` güncellenir, undo/redo world snapshot'a shape verisini de alır; `Kaydet` artık `country_shapes.json` da yazar |
| Ticaret yolu görsel sadeleştirme | ✅ | Harita üstü ticaret çizimi `A->B` ve `B->A` rotalarını tek koridorda birleştirir; `camScale < 0.85` iken yalnızca oyuncuya bağlı hatlar çizilir, etiketler yalnızca yakın zoom'da görünür |
| Harita modu (Normal/Ticaret) | ✅ | EU4 benzeri harita modu anahtarı eklendi; ticaret koridorları yalnızca `Ticaret` modunda çiziliyor, normal haritada çizgi karmaşası yok |
| Senaryo bazlı tarihsel ticaret merkezleri | ✅ | Trade map merkezleri senaryo `data/trade_centers.json` içindeki `tier` + `links` graph yapısından okunuyor; koridor akışı merkezler arasında doğrudan değil, link graph kısa yolu üzerinden dağıtılıyor; `off_map=true` ile sadece etiket ve bağlantı gösteren dış hat düğümleri (`name_tr`, `world_x`, `world_y`) de JSON’dan tanımlanabiliyor; `unlock_year` alanı sayesinde geç dönem Atlantik/Amerika hatları belirli yıldan önce tamamen gizli/pasif tutulabiliyor |
| Ticaret paneli ayrıştırması | ✅ | `Yeni Rota` sekmesi artık gerçek ticaret anlaşması adaylarını ve engel nedenlerini gösterir; manuel al/sat akışı ayrı `Pazar` sekmesine taşındı, müttefik devletlerle ticaret rota bazında bağımsız açılabiliyor; pazar sekmesinde fraksiyon/mal listeleri click anında satır resolve ettiği ve panel tam mouse state aldığı için seçimler tekrar güvenilir çalışıyor |

| WSL / Windows build hattı | ✅ | `wiki/dev/build-setup.md` içinde Ebitengine için Ubuntu paketleri, `go test ./...` doğrulaması ve `GOOS=windows GOARCH=amd64 go build -o bin/game.exe ./cmd/game` akışı belgelendi |

## Bilinen Sorunlar

| Öncelik | Sorun | Dosya | Etki |
|---|---|---|---|
| 🟢 Düşük | Kök dizinde geçici `game.exe` olabilir | `game.exe` | Kalıcı çıktı `bin/game.exe` olmalı |

## Sonraki Adım Planı

1. **Event görünürlüğü:** Choice sonuçlarını bölge bazlı ikon veya kısa süreli status etkisiyle daha görünür yap.
2. **AI eskort mantığı:** Transport filolarına savaş gemisi eşlik ettir.
3. **Event görünürlük/state izi:** Follow-up event koşullarını oyuncuya UI üzerinde önceden okunur hale getir.

## Yakın Sprint Önerisi

İlk sprintin hedefi "seçilen kampanya hedefi güvenilir çalışıyor ve kayıt yükleme bozmuyor" olmalı:

| Sıra | İş | Kabul Kriteri |
|---|---|---|
| 1 | Event görünürlüğü | Karar sonrası statü ikonları veya bölge marker'ları var |
| 2 | Naval escort | AI transport yanında escort gemisi de üretiyor |
| 3 | Event zinciri | Karar A/B farklı follow-up event setleri açıyor |

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
