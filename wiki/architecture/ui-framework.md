---
type: architecture
tags: [ui, render, input, widgets]
last_updated: 2026-06-05
related: [architecture/render-pipeline, architecture/game-loop, architecture/state-management, architecture/ui-screen-guide, dev/progress]
---

# UI Framework Migration Plan

## Hedef
Ebitengine üzerinde tüm ekranlarda ortak bir `internal/ui` katmanına geçerek:
- hitbox/üst üste binme/tıklama tüketimi hatalarını azaltmak,
- kod tekrarını düşürmek,
- yeni ekran geliştirme hızını artırmak,
- bakım maliyetini düşürmek.

## Mevcut Durum
1. `internal/ui` içinde çekirdek sözleşmeler (`Widget`, `InputState`, `Manager`) ve test edilebilir focus sırası eklendi.
2. Ortak `Button`, `Dropdown`, `Panel`, `Label`, `TextBox`, `Image/Icon`, `Tooltip`, `ListView`, `Checkbox`, `RadioGroup`, `Modal`, `Overlay` ve temel layout yardımcıları (`VBox`, `HBox`, `Grid`, `AnchorRect`, `Box`) oluşturuldu. `Label` artık sadece small text saran ince wrapper değil; `TextVariant` (small/medium/large) ve `TextAlign` (start/center/end) taşıyan ortak text primitive'i olarak menü/seçim ekranlarındaki manuel başlık/açıklama/hint çizimlerini de besliyor. Aynı katmanda `WrappedLabel` (satır saran metin bloğu) ve `OutlinedLabel` (gölgeli/outlined kısa etiket) primitive'leri de eklendi; modal açıklamaları ve ikon üstü sayaçlar bu yoldan çiziliyor.
3. `render` katmanında back/menu/mini/tiny buton çizimleri, edit mode dropdown'ları, trade, diplomasi, teknoloji, pause/save-load, ana menü ve seçim ekranları ortak bileşen yüzeylerine bağlandı.
4. HUD, recruit paneli, ordu split/merge overlay'i, edit mode inspector/form etkileşim yüzeyleri ve modal aileleri ortak button/panel/overlay geometry builder'larına taşındı.
5. Ortak button, dropdown, modal, HUD ve shape yardım paneli stilleri `internal/render/ui_theme.go` altında toplanmaya başladı.
6. Tam ekran menü/seçim aileleri için ortak screen chrome ve kart compose helper'ları (`internal/render/ui_compose.go`) kullanılmaya başlandı.
7. Ana menü ve kayıt/yükleme slot ekranlarında `Tab` focus geçişi `internal/ui.Manager` üzerinden çalışır.
8. `internal/render/ui_bridge.go` artık tek bir ortak `TextRenderer` sağlar; widget ailesi font varyantını UI katmanından ister, render katmanı yalnız face eşlemesini yapar.

## Hedef Mimari
1. `internal/ui` altında ortak UI framework
2. Tek input dispatcher (frame başına 1 kez input okuma + consume zinciri)
3. Ortak widget seti + layout + tema
4. Ekranların kademeli migrasyonu
5. Render/GameState ayrımı korunarak entegrasyon

## Aşama 0: Envanter ve Kurallar (1-2 gün)
1. Tüm ekranları envanterle:
   - MainMenu
   - Settings
   - ScenarioSelect
   - FactionSelect
   - VictorySelect
   - InGame HUD
   - Trade
   - Diplomacy
   - Tech
   - Pause
   - Save/Load
   - EditMode popup/panelleri
2. Her ekran için etkileşim matrisi çıkar:
   - tıklama alanları
   - keyboard shortcut
   - hover/tooltip
   - scroll/focus
   - modal davranış
3. UI kurallarını dondur:
   - z-order
   - event consume
   - focus/blur
   - clip/overflow
   - çözünürlük ölçekleme

## Aşama 1: UI Çekirdeği (3-5 gün)
1. `UIManager`:
   - `BeginFrame(input)`
   - `Dispatch()`
   - `Draw()`
   - layer stack
   - modal stack
2. `InputSnapshot`:
   - mouse pos
   - just-pressed/released
   - wheel
   - key durumları
   - text input
3. `Widget` temel arayüzü:
   - `ID`
   - `Bounds`
   - `Visible`
   - `Enabled`
   - `Z`
   - `HandleInput`
   - `Update`
   - `Draw`
4. Event modeli:
   - `OnClick`
   - `OnHover`
   - `OnChange`
   - `OnSubmit`
   - consume edilen event alttaki widget'a geçmez

## Aşama 2: Temel Bileşenler (4-6 gün)
1. Primitive bileşenler:
   - `Panel`
   - `Label`
   - `Button`
   - `Icon/Image`
2. Form bileşenleri:
   - `TextBox`
   - `Checkbox`
   - `RadioGroup`
   - `Dropdown/Combo`
3. Veri bileşenleri:
   - `ListView` (selection + scroll + row renderer)
4. Yardımcılar:
   - `Tooltip`
   - `Modal`
   - `ConfirmDialog`
5. Layout:
   - `VBox`
   - `HBox`
   - `Grid`
   - `Anchor`
6. Tema:
   - font/renk/spacing tokenları tek noktada
   - render tarafındaki ortak stiller `internal/render/ui_theme.go` içinde tutulur

## Aşama 3: Pilot Migrasyon (Trade) (2-3 gün)
1. Trade panelini tamamen yeni UI katmanına taşı
2. `internal/render/trade.go` içindeki manuel hit-test bloklarını kaldır
3. Kabul kriterleri:
   - tüm butonlar tıklanır
   - hover cursor doğru
   - panel açıkken arka harita etkileşimi yok
   - farklı çözünürlükte overlap yok
4. Ekran kompozisyonu primitive seviyesinde değil container/layout seviyesinde de ortak olmalı:
   - panel içi başlık/sekme/kolon/aksiyon kartı alanları `Box` cut/split akışıyla kurulmalı
   - sabit `x/y` zinciri yerine slot tabanlı rect türetimi kullanılmalı
5. Tam ekran seçim ekranlarında da aynı yaklaşım uygulanmalı:
   - üst/alt chrome bantları, başlık ve alt bilgi satırı ortak helper ile çizilmeli
   - seçim kartları ortak card rect helper'larıyla çizilmeli

## Aşama 4: Kademeli Ekran Geçişi (2-3 hafta)
1. Öncelik 1 (yüksek bug riski):
   - Diplomasi
   - Tech
   - Pause
   - Save/Load
2. Öncelik 2:
   - MainMenu
   - Settings
   - Scenario/Faction/Victory select
3. Öncelik 3:
   - InGame HUD parçaları
   - EditMode panel/popup
4. Her ekran sonrası zorunlu:
   - eski input path temizliği
   - çift-path bırakmama

## Aşama 5: Test ve Kalite Kapısı (sürekli)
1. Headless UI etkileşim testleri:
   - click
   - hover
   - scroll
   - focus
   - modal consume
2. Görsel regresyon smoke testleri:
   - 1280x720
   - 1600x900
   - 1920x1080
   - headless ortamda temel UI yüzeyleri için geometri smoke testi
   - ana menü için headless draw-call smoke testi
3. Performans kontrolü:
   - Draw/Update içinde gereksiz allocation yok
   - ortak modal builder'ları sıcak path'te children slice ayırmaz
   - çekirdek modal/button builder'ları heap allocation üretmez
4. CI kalite kapısı:
   - `go test ./...`
   - UI paket testleri
   - görsel smoke test

## Aşama 6: Dokümantasyon (1-2 gün)
1. `wiki/architecture/render-pipeline` güncelle
2. `wiki/architecture/ui-framework` sayfasını güncel tut
3. "Yeni ekran geliştirme rehberi" hazırla
4. Kod standardı belirle:
   - ekran dosyalarında manuel hitbox yasak
   - UI framework kullanımı zorunlu

## Riskler ve Önlemler
1. Risk: Eski ve yeni input path çakışması
   - Önlem: ekran bazlı feature flag + geçiş sonrası eski path sil
2. Risk: çözünürlükte taşma
   - Önlem: layout + clip + screenshot regresyon
3. Risk: dağınık state erişimi
   - Önlem: ViewModel katmanı, doğrudan `GameState` mutasyonunu sınırlama

## Teslimat Stratejisi
1. Sprint 1:
   - UI çekirdeği
   - temel bileşenler
   - Trade migrasyonu
2. Sprint 2:
   - Diplomasi
   - Tech
   - Pause
   - Save/Load
3. Sprint 3:
   - kalan menüler
   - HUD/edit panelleri
   - temizlik + dokümantasyon tamamlanması

## Done Kriteri
- Tüm ekranlarda input tek merkezden yönetiliyor
- Manuel dağınık hit-test kodu minimuma indirilmiş
- Kritik ekranlarda overlap/click-through bugları kapanmış
- CI testleri ve görsel smoke kontrolleri geçiyor

## Güncel Durum
1. Planın ekran migrasyonu kısmı tamamlandı:
   - Trade
   - Diplomasi
   - Teknoloji
   - Pause
   - Save/Load
   - MainMenu
   - ScenarioSelect
   - FactionSelect
   - VictorySelect
   - Settings
2. HUD içindeki küçük interaktif yüzeyler ortak builder'lara taşındı:
   - alt aksiyon HUD
   - harita modu düğmeleri
   - pazar toggle
   - tarih HUD menü düğmesi
   - müzik kontrol düğmeleri
   - event log toggle/kapatma
   - bölge paneli close/tax/hızlı diplomasi mini butonları
3. Layout'tan sonra screen compose da ortaklaştı:
   - `drawUIOverlay`
   - `drawUIPanelRect`
   - `drawUIPanelTopBar`
   - `drawUIPanelTitle`
   - `drawUISectionLabel`
   - `drawUIMutedText`
   - `drawUIInfoBlock`
   - `drawUIScreenChrome`
   - `drawUICardRect`
   - `drawUICardAccent`
4. Aynı compose helper'ları artık HUD/panel ailesine de uygulanıyor:
   - top status HUD
   - bottom action HUD
   - map mode / music / turn-tech mini kartları
   - event log paneli
   - region / army / sea / settlement bilgi panelleri
   - recruit panel ana çerçevesi ve queue bölümü
3. Genişletilmiş migrasyonla ortak geometri kullanan ek alanlar:
   - recruit panel close / kart / kuyruk iptal hit-test ailesi
   - ordu split/merge overlay aksiyonları
4. Trade ekranı artık yalnız ortak widget primitive'lerini değil, ortak kutu layout helper'larını da kullanır:
   - header, close button, tab strip, kontrol satırı, iki kolon ve action card `internal/ui/box.go` üstünden slotlara bölünür
   - overlap ve spacing regresyonları için `internal/render/ui_geometry_test.go` geometri smoke testleri çalışır
5. Diplomasi ve teknoloji ekranları da ortak kutu layout helper'larına taşındı:
   - diplomasi list/offer panellerinde title, list, action button ve footer alanları `Box` cut/split ile türetilir
   - teknoloji panelinde header, aktif araştırma satırı, tree body, hint ve close slotları aynı ortak layout akışını kullanır
6. Pause, save/load, settings ve seçim ekranları da ortak ekran layout helper'larına taşındı:
   - `internal/render/screen_layouts.go` içindeki centered stack/grid helper'ları senaryo, fraksiyon, zafer, kayıt slotu ve ayar satırı yerleşimlerini tek merkezleme kuralına bağlar
   - pause menüsü de panel/title/items/footer slotlarına ayrılır; manuel merkezleme formülleri ekran dosyalarından kademeli olarak çıkarılır
7. Render kompozisyonu da kademeli olarak tekilleştiriliyor:
   - `internal/render/ui_compose.go` ortak overlay, panel ve üst şerit çizim helper'larını içerir
   - trade, diplomasi, teknoloji ve pause ekranları artık aynı overlay/panel çizim helper'larını paylaşır; primitive ortaklığına ek olarak ekran chrome'u da ortaklaşır
8. Metin kompozisyonu da ortak helper'lara taşınmaya başladı:
   - panel başlığı, section label, muted hint ve kısa bilgi blokları `internal/render/ui_compose.go` içindeki text helper'larıyla çizilir
   - trade/diplomasi/teknoloji/pause ekranlarında serbest `DrawText` çağrıları kademeli azaltılır; ekran bazlı typography drift riski düşer
   - edit mode inspector/form tab, aksiyon ve form hit-test ailesi
   - genel onay, savaş ilan onayı, event detail ve historical event modal yüzeyleri
   - oyuncuya gelen diplomasi teklif diyaloğu
   - edit mode shape yardım overlay paneli
   - shape paint stroke preview overlay primitive'i
4. Halen tam widget ağacına taşınmamış ama ortak geometriyle çalışan alanlar:
   - shape paint canlı preview ve brush overlay katmanı
