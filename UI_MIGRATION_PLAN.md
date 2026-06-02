# UI Migration Plan (Tüm Ekranlar)

## Hedef
Ebitengine üzerinde tüm ekranlarda ortak bir `internal/ui` katmanına geçerek:
- hitbox/üst üste binme/tıklama tüketimi hatalarını azaltmak,
- kod tekrarını düşürmek,
- yeni ekran geliştirme hızını artırmak,
- bakım maliyetini düşürmek.

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

## Aşama 3: Pilot Migrasyon (Trade) (2-3 gün)
1. Trade panelini tamamen yeni UI katmanına taşı
2. `internal/render/trade.go` içindeki manuel hit-test bloklarını kaldır
3. Kabul kriterleri:
   - tüm butonlar tıklanır
   - hover cursor doğru
   - panel açıkken arka harita etkileşimi yok
   - farklı çözünürlükte overlap yok

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
2. Görsel regresyon screenshot testleri:
   - 1280x720
   - 1600x900
   - 1920x1080
3. Performans kontrolü:
   - Draw/Update içinde gereksiz allocation yok
4. CI kalite kapısı:
   - `go test ./...`
   - UI paket testleri
   - görsel smoke test

## Aşama 6: Dokümantasyon (1-2 gün)
1. `wiki/architecture/render-pipeline` güncelle
2. Yeni `wiki/architecture/ui-framework` sayfası ekle
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
1. Plan kapsamındaki ana ekran migrasyonu tamamlandı:
   - Trade
   - Diplomasi
   - Teknoloji
   - Pause
   - Save/Load
   - MainMenu
   - ScenarioSelect
   - FactionSelect
   - VictorySelect
2. Ortak `internal/ui` primitive seti genişletildi:
   - `Panel`
   - `Label`
   - `Button`
   - `Dropdown`
   - `ListView`
   - `Checkbox`
   - `RadioGroup`
   - `Modal`
   - `Overlay`
3. HUD ve özel overlay yüzeyleri ortak builder/widget hattına alındı:
   - alt aksiyon HUD
   - harita modu düğmeleri
   - recruit panel hit-test ailesi
   - ordu split/merge overlay'i
   - edit mode inspector/form yüzeyleri
   - confirm / war confirm / event detail / historical event modal aileleri
   - oyuncuya gelen diplomasi teklif diyaloğu
   - shape yardım overlay paneli
4. Kalite sonrası ek durum:
   - ortak button/dropdown/modal/shape yardım stilleri `internal/render/ui_theme.go` altında merkezileştirilmeye başladı
   - `internal/ui.Manager` için focus sırası testleri eklendi
   - modal builder sıcak path'inde gereksiz children slice allocation'ı kaldırıldı
5. Kalan dar alan:
   - shape paint canlı preview çizimi ortak `Overlay` primitive'i içinde çalışıyor, ancak piksel seviyeli brush çizimi doğası gereği render-spesifik kalıyor
