---
type: dev
tags: [progress, status, todo, known-issues, next-steps]
last_updated: 2026-05-11
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
| Ordu hareketi | ✅ | Komşuluk kısıtı, kara/deniz giriş kontrolü, savaş öncesi diplomasi kontrolü |
| Başlangıç orduları | ✅ | Her senaryonun `data/armies.json` dosyasından yükleniyor |
| Çarpışma motoru | ✅ | Birim gücü, arazi, teknoloji modları ve rastgele sonuç etkisi |
| Ordu detay paneli | ✅ | 20 slot, HP/deneyim çubukları, bölme/birleştirme aksiyonları |
| Ordu birleşme | ✅ | Dost bölgede otomatik veya panelden manuel birleşme, 20 birim limiti |
| Ordu bölme | ✅ | Seçili orduyu iki parçaya böler |
| Rakip ordu istihbaratı | ✅ | Menzildeki rakip orduda sayı ve yarım birim listesi görünür; menzil dışı detaylar gizlenir; emir verilemez |
| Çoklu ordu render | ✅ | Aynı bölgede ordular yan yana çizilir |
| Askeri kapasite | ✅ | Kara bölgesi başı 5 + kışla başı 5; ordu sayısı `ceil(kara_bölge/2)` |
| Asker alma | ✅ | Milis hızlı alım + belirli birim alımı; bina/teknoloji/altın/manpower kontrolü; JSON `turns_required` ile üretim kuyruğunda tamamlanır, tekrar tıklanınca iptal edilip altın iade edilir |
| Bina/birim hover bilgisi | ✅ | Kart tooltipleri maliyet, gereksinim, etki/istatistik ve görsel gösterir |
| Deniz birimi | ✅ | Liman ve kıyı koşuluyla filo/deniz birimi üretimini kuyruğa alma; tekrar tıklanınca iptal/iade |
| Ekonomi tick | ✅ | Vergi geliri, hasat modu, bina modları, ikincil mallar, tahıl bakım gideri |
| Vergi ayarlama | ✅ | Oyuncu bölgelerinde `.` / `,` ile ±5 |
| Bina inşası | ✅ | JSON bina tipleri, maliyet, arazi ve adet kısıtları; varsayılan 2 turluk üretim kuyruğu; kuyruktaki bina tekrar tıklanınca iptal/iade |
| Ticaret güzergahları | ✅ | `TradeRoutes` pasif gelir modeli var |
| Teknoloji ağacı | ✅ | Araştırma başlatma, tur sayacı, tamamlanan teknoloji efektleri, ağaç görünümü, seviye bazlı düzen, kategori renkleri, tamamlanmış teknoloji tick badge'leri, araştırma seçimi/değiştirme/vazgeçme, HUD'da aktif araştırma gösterimi, tur bitir uyarısı, tamamlanma mesajları event loguna ekleniyor |
| Diplomasi | ✅ | Savaş, barış, ittifak, ticaret; ilişki puanı ve duruş sistemi |
| Din diplomasisi | ✅ | Başlangıç ilişkileri din puanıyla kuruluyor; Sünni-Şii savaş başlıyor |
| Din dönüşümü | ✅ | Ele geçirilen bölgede 24 tur sonra yeni sahip dinine dönüşüm, memnuniyet -20 |
| Tarihsel olaylar | ✅ | JSON tetikleyici, tek seferlik olay işleme |
| Zafer koşulları | ✅ | `domination`, `economic`, `military`, `religious`, `conquer_city` kontrol ediliyor |
| AI turu | ✅ | Teknoloji, ekonomi, deniz, asker alma, konsolidasyon ve hedefe hareket |
| AI uzun menzilli hareket | ✅ | BFS ile uzaktaki hedefe doğru ilerleme |
| AI koalisyon | ✅ | Zorluk 3'te oyuncu 8+ bölgeyi geçince devreye girer |
| Kayıt/yükleme | ✅ | Autosave + slot1-3, metadata önizleme, silme |
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
| Edit mode bölge metadata editörü | ✅ | Inspector `Harita` sekmesinde region `name_tr`, `name`, `is_locked`, `unlock_turn` ve görsel Voronoi komşularından iki yönlü `neighbors` sync düzenlenir |
| Edit mode bölge ekleme/silme | ✅ | `Ctrl+Alt+sol` veya `Bolge Ekle` mevcut shape içinde yeni Voronoi seed region oluşturur; `Bolge Sil` seçili region'ı, komşu referanslarını ve o region'daki başlangıç ordularını kaldırır; undo/redo destekli |
| Edit mode geniş veri editörü | ✅ | Inspector `Veri` sekmesinde faction ekleme/düzenleme formu, faction silme, başlangıç kaynakları/playable/AI değeri, başlangıç diplomasi `stance/score`, başlangıç kara ordusu/donanma ekleme-silme ve seçili ordu/donanma birim sayıları düzenlenir; `Birim Tipi` dropdown'ı veri sekmesinde görünür; harita üstünde tüm ordu/donanma sayıları edit mode'da gizlenmeden görünür ve açık fraksiyon renklerinde kontrastlı metinle okunur; limanda demirli filolar liman anchor'ında, denize açılanlar deniz bölgesi anchor'ında çizilir; form `Kaydet` ve Ctrl+S `regions.json`, `factions.json`, `relations.json`, `armies.json` yazar |

## Bilinen Sorunlar

| Öncelik | Sorun | Dosya | Etki |
|---|---|---|---|
| 🔴 Kritik | Ekonomik zafer metni gelir diyor, kod hazineyi kontrol ediyor | `internal/victory/victory.go:83`, `internal/render/panel.go:837` | Oyuncu hedefi yanlış anlar; `TargetGoldIncome` isim/metin/kod uyumsuz |
| 🟠 Yüksek | Save load shape datasını state'e geri yazmıyor | `internal/save/save.go` | `ShapeData` kayıt yüklemede yeniden doldurulmuyor; render şu an dosyadan tekrar okuyarak haritayı kurtarıyor ama state eksik kalıyor |
| 🟠 Yüksek | Başlangıç zor zorluk bonusu oyuncu seçilmeden uygulanıyor | `internal/game/game.go:499` | `PlayerFactionID` boş olduğu için tüm fraksiyonlar AI bonusu alıyor; oyuncu seçilince bu bonus oyuncuda da kalabilir |
| 🟡 Orta | Deniz taşıma mekaniği yok | `internal/game/game.go:700` | Kara ordusu denize giremiyor; nakliye gemisi üretiliyor ama ordu taşıma akışı henüz yok |
| 🟡 Orta | Diplomasi teklifleri otomatik kabul | `internal/game/game.go` | AI kabul/red, pazarlık ve tehdit hesabı yok |
| 🟡 Orta | Olaylar oyuncu seçimi sunmuyor | `internal/events/events.go` | Olay sistemi tek yönlü etki uyguluyor, A/B kararları yok |
| 🟢 Düşük | Kök dizinde geçici `game.exe` olabilir | `game.exe` | Kalıcı çıktı `bin/game.exe` olmalı |

## Sonraki Adım Planı

1. **Ekonomik zafer kararını netleştir:** `TargetGoldIncome` gerçekten tur başı gelir mi, mevcut hazine mi ölçmeli? Kod, UI ve senaryo metni aynı anlama çekilmeli.
2. **Kayıt/yükleme bütünlüğü:** `LoadSlot` içinde senaryo metadata, `ShapeData`, `AvailableVictories` ve ses/senaryo asset yolu tutarlı şekilde geri yüklenmeli.
3. **Zorluk bonusu sıralaması:** Zor mod AI bonusunu fraksiyon seçildikten sonra, oyuncu hariç uygulanacak hale getir.
4. **Deniz taşıma akışı:** Nakliye gemisine kara ordusu bindirme/indirme, deniz geçişi ve kıyıdan çıkarma kurallarını ekle.
5. **AI ve diplomasi derinliği:** Otomatik kabul yerine ilişki, güç dengesi, ortak düşman ve komşu tehdit algısına göre kabul/red skoru ekle.
6. **Olay seçenekleri:** Tarihsel olay popup'larına A/B seçenekleri ve farklı etkiler ekle.
7. **Linux/WSL build notu:** Ebitengine için X11 ve ALSA paketlerini geliştirici dokümantasyonuna ekle; Windows build komutunu ayrıca doğrula.

## Yakın Sprint Önerisi

İlk sprintin hedefi "seçilen kampanya hedefi güvenilir çalışıyor ve kayıt yükleme bozmuyor" olmalı:

| Sıra | İş | Kabul Kriteri |
|---|---|---|
| 1 | Ekonomik zafer kararını netleştir | UI, JSON alanı ve `victory.Check` aynı şeyi ölçüyor |
| 2 | Zor mod bonus fix | Oyuncu seçilen fraksiyon AI başlangıç bonusu almıyor |
| 3 | Save load state tamamlama | Slot yükleyince harita, zafer hedefi, runtime tanımlar ve senaryo assetleri tutarlı |
| 4 | WSL bağımlılık notu | `go test ./...` için eksik native paketler wiki/README'de listelenmiş |

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
