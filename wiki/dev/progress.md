---
type: dev
tags: [progress, status, todo, known-issues, next-steps]
last_updated: 2026-05-29
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
| Ordu hareketi | ✅ | Komşuluk kısıtı, kara/deniz giriş kontrolü, savaş öncesi diplomasi kontrolü; donanmalar deniz bölgeleri arasında savaş ilanı olmadan dolaşır, deniz çatışması sadece `StanceWar` durumunda tetiklenir |
| Deniz taşıma akışı | ✅ | Kara ordusu uygun `transport` filosuna binebilir, filo `EmbarkedUnits` ile taşır, komşu dost/boş karaya çıkarma yapılır; oyuncu ve AI aynı kural setini kullanır |
| Amfibi savaş fazı | ✅ | Düşman kıyıya çıkarma savaş halinde aktif; çıkarma anı çatışması `combat` ile çözülür, başarılı çıkarma karaya ordu indirip sahiplik günceller, AI barışta çıkarma denemez |
| Başlangıç orduları | ✅ | Her senaryonun `data/armies.json` dosyasından yükleniyor |
| Çarpışma motoru | ✅ | Birim gücü, arazi, teknoloji modları ve rastgele sonuç etkisi |
| Ordu detay paneli | ✅ | 20 slot, HP/deneyim çubukları, bölme/birleştirme aksiyonları |
| Ordu birleşme | ✅ | Dost bölgede otomatik veya panelden manuel birleşme, 20 birim limiti |
| Ordu bölme | ✅ | Seçili orduyu iki parçaya böler |
| Rakip ordu istihbaratı | ✅ | Menzildeki rakip orduda sayı ve yarım birim listesi görünür; menzil dışı detaylar gizlenir; emir verilemez |
| Çoklu ordu render | ✅ | Aynı bölgede ordular yan yana çizilir |
| Askeri kapasite | ✅ | Kara bölgesi başı 5 + kışla başı 5; ordu sayısı `ceil(kara_bölge/2)` |
| Asker alma | ✅ | Milis hızlı alım + belirli birim alımı; bina/teknoloji/çoklu kaynak/manpower kontrolü; JSON `turns_required` ile üretim kuyruğunda tamamlanır, tekrar tıklanınca iptal edilip kaynaklar iade edilir |
| Çoklu eğitim kuyruğu (Total War benzeri) | ✅ | Recruit panelinde birim bazında `- xN +` seçimi, kuyrukta aynı birim için ilk tamamlanma turu görünürlüğü ve tek tıkta çoklu (`xN`) üretim emri; bölgesel kapasite `max(1,population/100)+kışla` kuralıyla sınırlandırılır |
| Bina/birim hover bilgisi | ✅ | Kart tooltipleri maliyet, gereksinim, etki/istatistik ve görsel gösterir |
| Deniz birimi | ✅ | Liman ve kıyı koşuluyla filo/deniz birimi üretimini kuyruğa alma; tekrar tıklanınca iptal/iade |
| Limandan denize çıkış (undock) | ✅ | Limana bağlı donanma aynı deniz bölgesine hareket emri aldığında liman bağını bırakıp deniz merkezine çıkar; hareket puanı tüketir |
| Ekonomi tick | ✅ | Vergi geliri, hasat modu, bina modları, ikincil mallar, taş üretimi, tahıl bakım gideri ve tahıl açığında lojistik HP cezası |
| Vergi ayarlama | ✅ | Oyuncu bölgelerinde `.` / `,` ile ±5 |
| Bina inşası | ✅ | JSON bina tipleri, maliyet, arazi ve adet kısıtları; varsayılan 2 turluk üretim kuyruğu; kuyruktaki bina tekrar tıklanınca iptal/iade |
| Kaynak reçete sistemi | ✅ | Birim ve bina üretiminde `grain/iron/timber/stone` tüketimi; UI maliyet satırı ve AI kararları bu modele bağlı |
| Bina seviye sistemi | ✅ | Binalar `max_per_region` kadar seviye alır (Lv1..LvN); panelde `Lv` ve kuyruk adedi görünür, inşa mesajları seviye geçişini (`LvX→LvY`) gösterir; kurulu bina kartları da tıklanabildiği için yükseltme/iptal akışı doğrudan kart üzerinden çalışır; manpower ve üretim kapasitesi kışla seviyesiyle artar |
| Ticaret güzergahları | ✅ | `TradeRoutes` pasif gelir modeli var |
| Teknoloji ağacı | ✅ | Araştırma başlatma, tur sayacı, tamamlanan teknoloji efektleri, ağaç görünümü, seviye bazlı düzen, kategori renkleri, tamamlanmış teknoloji tick badge'leri, araştırma seçimi/değiştirme/vazgeçme, HUD'da aktif araştırma gösterimi, tur bitir uyarısı, tamamlanma mesajları event loguna ekleniyor |
| Diplomasi | ✅ | `internal/diplomacy` ortak motoru ile savaş/barış/ittifak/ticaret; deterministik kabul-red, ilişki decay'i ve ticaret rotası senkronu |
| Diplomasi paneli modern akış | ✅ | Solda devlet seçimi + sağda teklif paneli; savaş/barış/ittifak/ticaret için muallak kabul olasılığı (%) ve durum göstergesi bulunur |
| Elenen fraksiyon diplomasi temizliği | ✅ | Kara toprağı biten fraksiyonlar (sadece deniz bölgesi kalsa bile) elendiğinde kara orduları + donanmaları kaldırılır, tüm diplomasi ilişkileri silinir ve diplomasi panelinde artık listelenmez |
| Liman işgalinde donanma tahliyesi | ✅ | Bölge el değiştirince, ele geçirilen limana bağlı eski sahip filoları otomatik limandan çıkarılır ve en yakın deniz bölgesine bırakılır |
| Oyuncuya gelen diplomasi teklif paneli | ✅ | AI barış teklifleri `diplomatic_offers` kuyruğuna düşer; oyuncu modal anlaşma panelinden kabul/red verir, kabulde standart diplomasi motoru uygulanır |
| Din diplomasisi | ✅ | Başlangıç ilişkileri din puanıyla kuruluyor; Sünni-Şii savaş başlıyor |
| Din dönüşümü | ✅ | Ele geçirilen bölgede 24 tur sonra yeni sahip dinine dönüşüm, memnuniyet -20 |
| Tarihsel olaylar | ✅ | JSON tetikleyici, tek seferlik olay işleme |
| Zafer koşulları | ✅ | `domination`, `economic`, `military`, `religious`, `conquer_city` kontrol ediliyor |
| AI turu | ✅ | Teknoloji, ekonomi, deniz, asker alma, konsolidasyon, diplomasi taraması ve hedefe hareket |
| AI uzun menzilli hareket | ✅ | BFS ile uzaktaki hedefe doğru ilerleme |
| AI koalisyon | ✅ | Zorluk 3'te oyuncu 8+ bölgeyi geçince devreye girer |
| Kayıt/yükleme | ✅ | Autosave + QuickSave + slot1-3, metadata önizleme, silme; tur bitirde autosave, oyun içi kaydetmede quicksave |
| Yükleme ekranı | ✅ | Senaryo ve kayıt yükleme sırasında gerçek zaman tabanlı hareketli spinner gösteriliyor |
| Ana menü / ayarlar | ✅ | Yeni oyun, autosave varsa devam et, kayıt yükleme, ayarlar, çıkış |
| Pause menüsü | ✅ | ESC ile açılır; devam, kaydet, yükle, ana menü, çıkış |
| Fare odaklı UI akışı | ✅ | Menü geri düğmeleri, teknoloji/diplomasi X kapatma, bölge/ordu panel kapatma, vergi/bina/asker aksiyonları fareyle yapılabilir |
| Olay paneli | ✅ | Sağ üst olay paneli daha fazla kayıt tutar, uzun liste mouse wheel ile kaydırılır |
| Minimap | ✅ | Sağ alt köşe, kamera ve ordu konumları |
| Üst-sol durum paneli | ✅ | Fraksiyon, kaynak, zafer ve ordu özeti haritanın üst-solunda ayrı panel; zafer/askeri özet kompakt iç kartlarla taşmadan çizilir |
| Sağ-üst tarih/menü HUD | ✅ | Tarih, mevsim, tur ve duraklama menüsü butonu sağ üstte ayrı panel |
| Alt-orta aksiyon HUD | ✅ | Diplomasi, Teknoloji ve Tur Bitir butonları ayrı HUD içinde alt-ortada |
| Olay logu akordiyonu | ✅ | Panel daraltılıp genişletilir; uzun metinler wrap edilir; kartlar X ile kapanır, tıklanınca detay popup açılır |
| Info popup bildirimi | ✅ | Altın yetersiz gibi oyun içi uyarılar olay loguna yazılmaz, ayrı geçici popup olarak görünür |
| Kompakt UI taşma düzeltmeleri | ✅ | Genel onay modalı mesaj wrap eder; bölge panelinde memnuniyet/vergi barları metin, buton ve alt çizgiyle çakışmaz |
| Panel cursor hit-test düzeltmesi | ✅ | Sol alt bölge paneli, olay logu, alt HUD, kayıt slotları ve onay panellerinde parmak imleci sadece gerçek tıklanabilir alanlarda gösterilir |
| Ses ve müzik | ✅ | `assets/sounds` global efektleri; senaryo `musics/` playlistleri `scenario.json` `music` alanından; ayarlarda ayrı müzik/ses seviyeleri; oyun içi müzik HUD'u ve ESC menüsü müzik kontrolleri |
| Development mode | ✅ | `DEV_MODE=true` ile `GameState.DevelopmentMode` |
| Render başlangıç log temizliği | ✅ | Boş senaryo path'inde shape dosyası okunmaz; deniz seed araması ham `world_x/world_y` fallback kullanır |
| Deniz anchor ve çakışma stabilizasyonu | ✅ | Deniz orduları gerçek su piksel anchor'ına çizilir; ordu/etiket çizim sırası deterministik, çakışan etiket metinleri bastırılır |
| Çoklu yerleşim noktaları | ✅ | `regions.json` içinde `settlements[]`; ana yerleşim ordu/etiket anchor'ı, yakın zoom'da ek yerleşim noktaları/isimleri, bölge dışı koordinatta log + nearest-region fallback |
| Settlement edit mode | ✅ | `.env` `EDIT_MODE=true`; senaryo seçince harita editörü açılır, alt-sol bilgi/aksiyon HUD'u, settlement ekleme/silme, tip/capital değiştirme, bölge terrain/owner değiştirme, sürükleme, bölge arası taşıma, isim düzenleme, Shift+sürükle ile bölge merkezi taşıma ve Ctrl+S ile `regions.json` kaydı |
| Dropdown component | ✅ | `internal/render/renderer.go:Dropdown`; edit mode'da sahip/arazi/yerleşim tipi seçimlerinde yeniden kullanılabilir dropdown, scroll ve tam içerik desteği |
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
| Senaryo bazlı tarihsel ticaret merkezleri | ✅ | Trade map merkezleri senaryo `data/trade_centers.json` içindeki `tier` + `links` graph yapısından okunuyor; koridor akışı merkezler arasında doğrudan değil, link graph kısa yolu üzerinden dağıtılıyor |

## Bilinen Sorunlar

| Öncelik | Sorun | Dosya | Etki |
|---|---|---|---|
| 🟡 Orta | Olaylar oyuncu seçimi sunmuyor | `internal/events/events.go` | Olay sistemi tek yönlü etki uyguluyor, A/B kararları yok |
| 🟢 Düşük | Kök dizinde geçici `game.exe` olabilir | `game.exe` | Kalıcı çıktı `bin/game.exe` olmalı |

## Sonraki Adım Planı

1. **Olay seçenekleri:** Tarihsel olay popup'larına A/B seçenekleri ve farklı etkiler ekle.
2. **WSL build notu:** Ebitengine için X11 ve ALSA paketlerini geliştirici dokümantasyonuna ekle; Windows build komutunu ayrıca doğrula.
3. **AI strateji derinliği:** Uzun menzilli amfibi hedefleme ve çoklu filo koordinasyonunu geliştir.

## Yakın Sprint Önerisi

İlk sprintin hedefi "seçilen kampanya hedefi güvenilir çalışıyor ve kayıt yükleme bozmuyor" olmalı:

| Sıra | İş | Kabul Kriteri |
|---|---|---|
| 1 | Olay seçenekleri | Tarihsel olaylarda A/B kararları ve farklı sonuçlar var |
| 2 | WSL bağımlılık notu | `go test ./...` için eksik native paketler wiki/README'de listelenmiş |
| 3 | AI amfibi derinlik | AI zayıf çıkarma denemelerini azaltıp hedef seçimini daha tutarlı yapıyor |

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
