---
type: architecture
tags: [render, ebitengine, camera, input, ui]
last_updated: 2026-05-19
related: [game-loop, state-management, shape-editor, systems/combat]
---

# Render Pipeline

**Kaynak:** `internal/render/renderer.go`

## Renderer Yapısı

```go
type Renderer struct {
    gs       *state.GameState
    worldMap *WorldMap          // üretilmiş harita görüntüsü

    camX, camY float64          // dünya uzayında kamera merkezi
    camScale   float64          // zoom (0.25 – 3.0)

    SelectedRegion world.RegionID
    SelectedArmy   army.ArmyID

    showDiplomacy, showTech bool
    eventLog []string           // son 8 olay
    combatLog string            // 3 saniyelik bildirim

    // Duraklama menüsü
    pauseCursor int

    // Kayıt/yükleme slot seçim ekranı
    slotCursor        int
    saveSelectMode    bool
    pendingDeleteSlot string  // onay bekleyen slot adı
}
```

---

## Draw Katman Sırası

`Draw(screen)` — `internal/render/renderer.go:166`

| Sıra | Katman | Dosya |
|---|---|---|
| 0 | Özel tam ekranlar: ana menü, ayarlar, fraksiyon seçim, zafer seçim, game over | `main_menu.go`, `settings.go`, `faction_select.go`, `victory_select.go` |
| 0 | Yükleme ekranı: spinner + durum metni | `loading.go` |
| 0 | Kayıt slot seçim ekranları (PhaseLoadSelect / PhaseSaveSelect) | `load_select.go` |
| 0 | Duraklama menüsü (PhasePauseMenu) — harita altta, overlay üstte | `pause_menu.go` |
| 1 | Dünya haritası (WorldMap cache) | `mapgen.go`, `tile.go` |
| 2 | Seçim halkası (bölge) | `renderer.go` |
| 3 | Ticaret koridorları (çift yön rotalar tek hatta birleştirilir; uzak zoom'da yalnızca oyuncuya bağlı koridorlar çizilir) | `renderer.go` |
| 3 | Hareket hedefleri (ordu komşuları) | `renderer.go` |
| 4 | Bölge etiketleri + şehir noktası; edit mode'da bölge merkezi işaretleri, Voronoi debug overlay ve `Shape` sekmesi aktifken country shape outline/brush overlay'i; etiketler stabil sıralanır ve çakışan metinler atlanır | `renderer.go` |
| 5 | Ordu ikonları; çizim sırası ekran konumu + ID ile deterministiktir; edit mode'da tüm ordu/donanma birim sayıları görünür; ikon üstü sayı metni fraksiyon rengine göre kontrast uyarlamalıdır | `renderer.go` |
| 6 | UI panelleri (üst-sol durum paneli, sağ-üst tarih/menü HUD, alt-orta aksiyon HUD, bölge/ordu/minimap/event log) | `panel.go` |
| 6 | Ordu detay paneli — 20 slot ızgarası, boş slotlar silik | `army_panel.go` |
| 6 | Bölge üretim UI — bina kartlarında kuyruktaki inşaatın kalan tur etiketi ve tekrar tıklayınca iptal, birim kartlarında üretim turu | `panel.go`, `recruit_panel.go` |
| 6 | Olay logu akordiyonu — daralt/genişlet, wrap edilmiş kartlar, X ile kapatma, tıklayınca detay popup | `panel.go`, `renderer.go` |
| 6 | Edit mode alt-sol bilgi HUD'u — seçili bölge/settlement/ordu özeti ve edit butonları | `renderer.go` |
| 7 | Diplomasi paneli (Tab) — tam ekran overlay | `diplom.go` |
| 8 | Teknoloji paneli (T) — tam ekran ağaç görünümü | `tech_panel.go` |
| 9 | Info popup bildirimi (combatLog, olay loguna yazmaz) | `renderer.go`, `panel.go` |
| 10 | Savaş ilan ve genel onay diyalogları; genel onay mesajı açılışta satırlara bölünür, buton hitbox'ları aynı sabitlerden hesaplanır | `renderer.go`, `cursor.go` |
| 11 | Tarihsel olay popup | `renderer.go` |

---

## UI Components

### Dropdown Component

**Kaynak:** `internal/render/renderer.go:Dropdown`

Edit mode'da kullanılan yeniden kullanılabilir dropdown component. Sahip, arazi, yerleşim tipi ve veri editöründeki birim tipi seçimlerinde kullanılır.

```go
type Dropdown struct {
    x, y, w, h int
    options    []string
    selected   string
    scroll     int
    open       bool
}
```

**Metodlar:**
- `SetPosition(x, y float32)` — dropdown konumunu ayarlar
- `SetOptions(options []string, selected string)` — seçenekleri ve seçili değeri ayarlar
- `Toggle()` — aç/kapat
- `Close()` — kapat
- `IsOpen() bool` — açık mı kontrolü
- `HitTest(mx, my float64) bool` — fare pozisyonu dropdown içinde mi
- `Scroll(dy float64)` — tekerlek ile kaydırma
- `GetSelectedOption(mx, my float64) (int, bool)` — tıklanan seçeneği döndürür
- `Draw(screen *ebiten.Image)` — render

---

## Kamera Sistemi

**Koordinat sistemi:** Dünya uzayı `(WorldW × WorldH)` px, ekran uzayına dönüşüm:

`WorldW`, `WorldH`, `shape_offset_*` ve `shape_scale_*` aktif senaryonun `scenario.json` içindeki `map` alanından okunur. Alan eksikse renderer eski varsayılanları kullanır (`2892×1440`, offset `-530/-180`, scale `2.025/2.025`).

```
screenX = (worldX - camX + worldY * mapShearX) * camScale + ScreenWidth/2
screenY = (worldY - camY) * camScale * mapPitchY + ScreenHeight/2
```

`mapPitchY = 1.0`, `mapShearX = 0.0` → şu an düz 2D (izometrik bükme kapalı)

**Zoom:** Fare tekerleği ile fare pozisyonuna odaklanarak büyütür. Uzaklaşma limiti `internal/render/renderer.go:minCameraScale` üzerinden aktif senaryonun `world_width` / `world_height` değerlerinden gelen `WorldW` / `WorldH` boyutuna göre hesaplanır; oyuncu haritayı ekrana tamamen sığdıran ölçeğin altına inemez. Yakınlaşma üst sınırı `3.0`.

**Sürükleme:** Orta fare tuşu basılıyken dünya uzayı delta hesaplanır.

---

## WorldMap Cache

`WorldMap` — `internal/render/mapgen.go`

Harita, her fraksiyon sahipliği değişiminde `MarkDirty()` ile işaretlenir ve bir sonraki `Refresh()` çağrısında yeniden üretilir. Bölge poligonları normalde `country_shapes.json`'dan gelir; edit mode sırasında ise `GameState.ShapeData` içindeki anlık shape verisi önceliklidir, böylece paint edit sonrası `rebuildEditWorldMap()` doğrudan yeni sınırı gösterir.

Deniz bölgeleri `internal/render/mapgen.go:buildSeaRegions` içinde kara pikselleri bariyer kabul eden multi-source BFS ile üretilir. Seed araması önce mevcut shape dönüşümlü koordinatı, sonuç çıkmazsa ham `world_x/world_y` koordinatını dener; bu, senaryo verisindeki deniz merkezlerinin dünya pikseli olarak tutulduğu durumlarda `_sea_*` seed uyarılarını engeller.

Deniz ve kara region raster alanlarından `WorldMap.RegionAnchor` hesaplanır. Deniz orduları ve deniz hareket hedefleri JSON merkez koordinatı yerine bu gerçek piksel anchor'ını kullanır; anchor, bölgenin kendi piksel alanı içinden seçildiği için kıyıda kara poligonunun kapattığı deniz bölgelerinde filo ikonları karanın üstüne düşmez.

Kara bölgelerde görünen şehir noktaları `regions.json` içindeki `settlements[]` alanından gelir. `WorldMap` her yerleşim için `SettlementAnchor` hesaplar; koordinat yanlışlıkla bölge dışına verilirse log uyarısı basılır ve aynı region içindeki en yakın piksele fallback yapılır. Ordu ikonları ve hareket hedefleri ana yerleşim (`is_capital`) anchor'ını kullanır, `world_x/world_y` ise bölge geometrisi için korunur.

Edit mode'da `world_x/world_y` merkezleri ayrı işaretlerle çizilir. Kara ve deniz bölgesi odak noktaları farklı renktedir; deniz seçiliyken odak işareti kara seçiminden farklı mavi/camgöbeği tona döner. Shift + sol sürükleme bu koordinatları değiştirir; Voronoi sınırları `WorldMap` raster cache'ine bağlı olduğu için sürükleme sırasında sadece merkez işareti güncellenir, fare bırakıldığında cache bir kez yeniden oluşturulur.

---

## Input Yönetimi

`HandleInput()` döner: `InputAction{Kind, ArmyID, TargetRegion, TargetFaction, BuildingID, Delta}`

**Just-pressed takibi:** `prevKeys`, `prevMouse` map'leri tutulur; `keyJustPressed()` / `mouseJustPressed()` bir frame'lik tetikleme sağlar.

**Tık öncelik sırası:**
1. Açık detay paneli kapatma düğmeleri (bölge/ordu)
2. Alt-orta aksiyon HUD butonları (diplomasi, teknoloji, tur bitir)
3. Olay logu akordiyonu: başlık butonu paneli daraltır/genişletir, kart X'i olayı kapatır, kart gövdesi detay popup açar
4. UI bölgesi (üst-sol durum paneli / sağ-üst tarih-menü HUD / alt-orta aksiyon HUD / sağ panel) → geçersiz say
5. Bölge paneli aksiyonları: vergi +/- düğmeleri, oluşturulabilir bina kartına tıklayarak inşa; `is_locked=true` bölgelerde vergi/inşa/birim alımı hit-test'te kapatılır
6. Birim oluştur paneli (`recruit_panel.go:RecruitPanelHitTest`); kıyı olmayan bölgelerde deniz birimleri gösterilmez
7. Bölge/birim oluştur paneli boş alan tıklamaları → tüketilir, arkadaki haritaya düşmez
8. BÖL/BİRLEŞTİR butonları (seçili ordu varsa, `army_panel.go` hit-test)
9. Ordu ikonuna tıklama — `armyIconPositions()` üzerinden offset'li 14px yarıçap
10. Bölge seçimi (WorldMap pixel lookup)

Edit mode'da oyun HUD/panelleri çizilmez; harita, üst edit HUD ve alt-sol sekmeli inspector görünür. Sol tık settlement, bölge veya ordu seçer; settlement sürükleme koordinatı canlı taşır ve başka kara region'a bırakılan settlement o region'ın `settlements[]` listesine aktarılır. Alt + sol tık tıklanan kara bölgeye yeni settlement ekler; Ctrl + Alt + sol tık tıklanan bölgenin `shape_id` alanını paylaşan yeni Voronoi seed region oluşturur. Delete seçili settlement'ı, settlement seçili değilse seçili region'ı siler. Shift + sol sürükleme seçili bölgenin `world_x/world_y` merkezini taşır ve fare bırakıldığında harita cache'ini yeniler. Inspector `Harita` sekmesindeki `Yerlesim Ekle`, `Tip`, `Ana Yap`, `Isim`, `Arazi`, `Sahip`, `Ad TR`, `Ad EN`, `Kilit`, `-10 Tur`, `+10 Tur`, `Komsu Sync`, `Bolge Ekle`, `Bolge Sil`, `Yerlesim Sil` ve `Kaydet` butonları region/settlement metadata işlemlerini doğrudan çalıştırır. `+10/-10 Tur`, `unlock_turn` alanını değiştirir; `is_locked=true` ve `unlock_turn>0` ise bölge aktif tur o değere ulaştığında otomatik açılır. Deniz region'larında settlement işlemleri kapalı kalır ama bölge odaklı seçim, merkez taşıma, komşu sync, ekleme/silme ve owner/terrain düzenleme aynıdır; inspector bu seçimlerde açıkça `Deniz Bolgesi`, `Deniz bolgesinde yerlesim yok.` ve pasif `Denizde Yok` etiketi gösterir. Settlement odaklı pasif butonlar da bağlama göre `Tip Yok`, `Isim Yok`, `Silinmez` ya da settlement seçimi bekleniyorsa `Tip Sec`, `Isim Sec`, `Sil Sec` etiketine döner. `Shape` sekmesi seçili kara region'ın `shape_id` kaydını düzenler; sağ mouse drag ile paint/erase brush mask'e uygulanır, stroke sırasında ekleme/silme pikselleri canlı preview overlay ile gösterilir, mouse bırakılınca contour ring'leri yeniden üretilir, `GameState.ShapeData` ve ilgili `Region.Shape` alanları güncellenir, ardından harita cache'i yeniden kurulur. Sağ üst yardım paneli aktif `shape_id`, brush boyutu ve kontrol özetini gösterir. `Tip`, `Arazi` ve `Sahip` inspector yanında kaydırılabilir dropdown açar; seçilen satır ilgili `type`, `terrain` veya `owner_id` değerini doğrudan yazar. `Veri` sekmesi faction ekleme/düzenleme formu, faction silme, başlangıç kaynakları/playable/AI değeri, başlangıç kara ordusu/donanma ekleme-silme ve seçili ordu/donanma birim tip-sayılarını düzenler. Donanma ekleme liman tipli yerleşimin kara region'ından komşu deniz region'ına `is_naval: true` ordu yerleştirir. Faction formu ID, `name`, `name_tr`, din, renk, playable, kaynaklar, AI, hedef faction, diplomasi `stance` ve `score` alanlarını tek yerde toplar; formdaki `Kaydet` değişikliği uygular ve senaryo JSON dosyalarını yazar. `Kaydet` / `Ctrl+S` artık `regions.json`, `country_shapes.json`, `factions.json`, `relations.json` ve `armies.json` dosyalarını birlikte yazar. F2/Enter seçili settlement adını düzenler, Ctrl+S `ActionSaveScenario` üretir.

Voronoi debug overlay `V` ile açılıp kapanır. Overlay `WorldMap.BoundaryPixels` ile seçili veya hover bölgenin gerçek raster sınırını camgöbeği piksellerle çizer. `WorldMap.VisualNeighbors` üzerinden raster sınır komşularını çıkarır ve JSON `neighbors` listesiyle karşılaştırır: yeşil çizgi görsel+JSON komşu, kırmızı çizgi sadece görsel komşu, gri çizgi sadece JSON komşudur. Sağ üst panel hover pixel'in `RegionAt` sonucunu, senaryo koordinatını ve seçili bölgenin visual/json komşu sayısını gösterir. `Komsu Sync`, seçili region'ın görsel komşularını JSON `neighbors` listesine yazar; eklenen/çıkarılan her komşuda karşı region listesi de iki yönlü güncellenir.

Edit mode'da `editDirty` true iken ESC doğrudan çıkmaz; genel onay modalı üç seçenekle açılır: `Kaydet` önce `ActionSaveScenarioAndGoMainMenu` üretir, kayıt başarılıysa ana menüye döner; `Kaydetmeden Cik` doğrudan `ActionGoMainMenu` üretir; `Iptal` modalı kapatır.

Undo/redo edit mode içinde `editUndoStack` / `editRedoStack` ile tutulur. Settlement işlemleri yalnızca etkilenen region'ların `settlements[]` snapshot'ını alır; region center değişiklikleri sadece eski/yeni `world_x/world_y`, owner/terrain/type/name/lock/unlock değişiklikleri ilgili alan snapshot'ını tutar. Neighbor sync etkilenen tüm region `neighbors[]` listelerini snapshot'lar; region ekleme/silme, shape paint commit'i ve ordu/donanma ekleme-silme/birim sayısı değişiklikleri region map, order, başlangıç orduları ve `ShapeData` için dünya snapshot'ı kullanır; geniş veri editörü faction/army alanları için küçük alan command'leri üretir. `Ctrl+Z` undo, `Ctrl+Y` veya `Ctrl+Shift+Z` redo üretir; drag işlemleri command'i frame frame değil mouse bırakıldığında tek kez push eder.

Menü ve üst paneller fareyle tamamlanabilir: senaryo/fraksiyon/zafer ve kayıt ekranlarında `Geri` düğmesi vardır; diplomasi ve teknoloji panelleri X düğmesiyle kapanır; kayıt silme onayı kart içi `Sil`/`İptal` düğmeleriyle yapılır. Ayarlar ekranında müzik/ses efektleri aç-kapat ve her ikisi için `0-100` arası ayrı seviye bulunur. Paylaşılan efektler `assets/sounds/` altından yüklenir; senaryo müziği `scenario.json` içindeki `music.default_playlist` ile başlar ve dosyaları senaryo `musics/` klasöründen okur. Oyun içi müzik HUD'u aktif parçayı gösterir ve `Dur/Cal` ile `Sonr` kontrollerini sunar; ESC menüsünde müzik aç/kapat ve müzik seviyesi hızlıca değiştirilebilir.

Harita modu anahtarı alt-orta aksiyon HUD'unun üstündeki `Normal | Ticaret` segmentinde yer alır. `M` kısayolu veya bu segment ile mod değişir. Ticaret koridor çizimi yalnızca `Ticaret` modunda render edilir; `Normal` modda ticaret çizgileri tamamen gizlidir.

`Ticaret` modunda harita üstüne hafif desatüre/sisli bir overlay eklenir ve çizim tüm fraksiyon çiftleri arasında birebir mesh yerine `ticaret merkezi` odaklı yapılır: merkez düğümleri senaryo bazlı `data/trade_centers.json` dosyasından okunur, fraksiyonlar en yakın merkeze ince spoke ile bağlanır, ana ağ ise merkezler arası kavisli bezier glow/core koridorlar olarak çizilir. Merkezler arası akış doğrudan her çift arasında çizilmez; senaryoda tanımlı `links` graph'ı üzerindeki kısa yol boyunca dağıtılır (ör. Halep -> Konstantinopolis -> Venedik). Çizim sırası deterministik tutulduğu için frame-frame titreme/yanıp sönme engellenir.

Ticaret koridorları etkileşimlidir: koridor üzerine hover yapıldığında koridor focus moduna geçilir (arka ağ karartılır, seçili hat parlatılır) ve tooltip `merkez A ↔ merkez B`, `hacim/tur`, `bağlı fraksiyon` ve baskın emtia özeti gösterir; sol tık aynı bilgiyi kısa bildirim olarak yazar.

Merkez odak modu: Ticaret merkezi düğümlerinden birinin üzerine hover yapıldığında yalnız o merkeze bağlı koridorlar belirgin tutulur, diğer ağ düşük alpha ile geri plana atılır. Merkeze tıklama, bağlı koridor sayısı ve toplam hacim özetini verir.

Üst-sol durum paneli `internal/render/panel.go:185` içinde çizilir. Sağ taraftaki zafer hedefi ve askeri kapasite alanları, sabit ölçülü iç kartlar ve kendi ilerleme barı çizimiyle sınırlandırılır; böylece zafer barı askeri kapasite ayırıcısına veya panel sağ sınırına taşmaz.

Uzun sürebilen senaryo/kayıt yükleme işleri `PhaseLoading` ekranına geçer. `internal/game/game.go` yükleme işini arka planda başlatır; renderer bu sırada `loading.go` içindeki gerçek zaman tabanlı spinner'ı çizer ve sonuç hazır olduğunda state ana thread üzerinde uygulanır.

Rakip orduları seçilebilir ama emir verilemez. Renderer rakip ordusu için hareket hedefi çizmez ve sağ/sol tık hareket aksiyonu üretmez. Oyuncu ordularından birinin mevcut hareket menzilindeki rakip ordularda ikon birim sayısını gösterir; detay panelinde birimlerin yaklaşık yarısı görünür, kalanları `Gizli` kartlarıyla saklanır. Menzil dışındaki rakip ordularda birim sayısı ve hareket/birim detayları gizli kalır.

Bina ve birim kartlarında hover tooltip vardır. Tooltip maliyet, gereksinim, temel etki/istatistik ve kart görselini gösterir. Bölgeye uygun olmayan bina kartları render edilmez; liman son sıradadır ve kıyı olmayan bölgelerde görünmez.

Bölge bilgi panelinde parmak imleci panelin tamamında değil, yalnızca kapatma düğmesi, vergi `-/+` düğmeleri ve inşa edilebilir bina kartları üzerinde gösterilir. Oyun içi HUD/panel cursor davranışı gerçek etkileşim alanlarına bağlıdır: sağ üstte yalnızca `Menü`, alt HUD'da yalnızca üç aksiyon butonu, olay logunda toggle/kart/X, birim panelinde yalnızca birim kartları pointer üretir. Boş panel alanları tıklamayı tüketmeye devam eder ama clickable cursor üretmez.

---

## Bildirim Sistemi

```
ShowCombatResult(msg)          → combatLogTimer = 180 frame (~3 sn), ayrı info popup; eventLog'a eklemez
ShowHistoricalEvent(title,desc) → tam ekran popup, herhangi tuş/tık ile kapatılır
AddEvent(msg)                  → sağ olay logundaki kalıcı kart listesine ekler
```

`eventLog` maksimum 8 girdi tutar; yeniler öne eklenir, sondan taşanlar düşer.

---

## Ordu İkon Sistemi

Aynı bölgede birden fazla ordu bulunabilir. `armyIconPositions()` (`renderer.go`) kara ordularını `RegionID`/yerleşim anchor'ına göre, donanmaları ise sadece `docked_region_id` / `docked_settlement_id` doluyken bağlı liman anchor'ında; aksi halde deniz bölgesi anchor'ında gruplar. Dock state varsa renderer bunu doğrudan kullanır. Her grubun ikonlarını 26px aralıklarla yatayda ortalar. Hem `drawArmies` hem `handleLeftClick` hem de `cursor.go:inGameHovering` bu tek fonksiyonu kullanır — tutarsızlık riski yoktur.

```
Tek ordu  →  bölge merkezinde
İki ordu  →  merkez ±13px
Üç ordu   →  merkez -26px, 0px, +26px
```

---

## Minimap — Ordu Konumları

`panel.go:drawMinimapArmies` bölge sahiplik noktaları yerine orduların konumlarını çizer. Her ordu fraksiyon rengiyle dolu bir daire + ortada beyaz nokta olarak gösterilir; oyuncunun orduları altın kenarlıkla ayrışır.

---

## Ordu Bölme (Split)

`army_panel.go:DrawArmyDetailPanel` seçili ordunun panel başlığında "✂ BÖL" butonu gösterir (≥2 birim şartı). `SplitButtonHitTest()` hit-test fonksiyonu `renderer.go` ve `cursor.go` tarafından kullanılır. Buton tıklandığında `ActionSplitArmy` üretilir; `game.go:splitArmy()` birimleri ikiye böler ve yeni ordu oluşturur.

---

## İmleç Yönetimi (`cursor.go`)

`updateCursorShape()` her frame çalışır. Aşağıdaki fazlarda parmak imleci gösterilir:

| Faz | Koşul |
|---|---|
| PhaseMainMenu | `mainMenuHoverIndex >= 0` |
| PhaseFactionSelect | `factionCardHoverIndex >= 0` |
| PhaseVictorySelect | `victoryCardHoverIndex >= 0` |
| PhasePlayerTurn | `inGameHovering` (butonlar, ordu ikonları, BÖL butonu) |
| PhasePauseMenu | `pauseMenuHoverIndex >= 0` |
| PhaseLoadSelect / PhaseSaveSelect | `slotHoverIndex >= 0` |
| PhaseSettings | `settingsHovering` |

---

## Dosya Sorumlulukları

| Dosya | İçerik |
|---|---|
| `renderer.go` | Kamera, draw döngüsü, input ana yönlendirici, `armyIconPositions()` |
| `mapgen.go` | WorldMap cache, poligon doldurma |
| `tile.go` | Arazi renk/doku katmanı |
| `panel.go` | Alt bar, bölge/ordu/minimap/event log panelleri; event log kaydırma geometrisi; minimap'te ordu konumları |
| `army_panel.go` | Ordu detay paneli — 20 slot ızgara, HP çubuğu, BÖL butonu |
| `diplom.go` | Diplomasi paneli UI + input |
| `tech_panel.go` | Teknoloji ağacı paneli + input |
| `pause_menu.go` | Oyun içi duraklama menüsü (ESC) |
| `load_select.go` | Kayıt slot seçim ekranı (yükleme + kaydetme + silme) |
| `recruit_panel.go` | Birlik alım paneli |
| `action.go` | `InputAction` ve `ActionKind` tanımları |
| `font.go` | Font yükleme, `DrawText`, `MeasureText` |
| `assets.go` | Görsel varlık yükleme |
| `cursor.go` | İmleç şekli yönetimi (tüm fazlar) |
| `faction_select.go` | Fraksiyon seçim ekranı |
| `victory_select.go` | Zafer koşulu seçim ekranı |
| `main_menu.go` | Ana menü ("Devam et" → autosave yükleme, "Kayıttan Yükle" → slot seçim ekranı) |
| `settings.go` | Ayarlar ekranı |
