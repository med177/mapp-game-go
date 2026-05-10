---
type: architecture
tags: [render, ebitengine, camera, input, ui]
last_updated: 2026-05-10
related: [game-loop, state-management, systems/combat]
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
| 3 | Hareket hedefleri (ordu komşuları) | `renderer.go` |
| 4 | Bölge etiketleri + şehir noktası; edit mode'da bölge merkezi işaretleri ve Voronoi debug overlay; etiketler stabil sıralanır ve çakışan metinler atlanır | `renderer.go` |
| 5 | Ordu ikonları; çizim sırası ekran konumu + ID ile deterministiktir | `renderer.go` |
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

Edit mode'da kullanılan yeniden kullanılabilir dropdown component. Sahip, arazi ve yerleşim tipi seçimlerinde kullanılır.

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

Harita, her fraksiyon sahipliği değişiminde `MarkDirty()` ile işaretlenir ve bir sonraki `Refresh()` çağrısında yeniden üretilir. Bölge poligonları `country_shapes.json`'dan gelir; renkler fraksiyon rengiyle doldurulur.

Deniz bölgeleri `internal/render/mapgen.go:buildSeaRegions` içinde kara pikselleri bariyer kabul eden multi-source BFS ile üretilir. Seed araması önce mevcut shape dönüşümlü koordinatı, sonuç çıkmazsa ham `world_x/world_y` koordinatını dener; bu, senaryo verisindeki deniz merkezlerinin dünya pikseli olarak tutulduğu durumlarda `_sea_*` seed uyarılarını engeller.

Deniz ve kara region raster alanlarından `WorldMap.RegionAnchor` hesaplanır. Deniz orduları ve deniz hareket hedefleri JSON merkez koordinatı yerine bu gerçek piksel anchor'ını kullanır; anchor, bölgenin kendi piksel alanı içinden seçildiği için kıyıda kara poligonunun kapattığı deniz bölgelerinde filo ikonları karanın üstüne düşmez.

Kara bölgelerde görünen şehir noktaları `regions.json` içindeki `settlements[]` alanından gelir. `WorldMap` her yerleşim için `SettlementAnchor` hesaplar; koordinat yanlışlıkla bölge dışına verilirse log uyarısı basılır ve aynı region içindeki en yakın piksele fallback yapılır. Ordu ikonları ve hareket hedefleri ana yerleşim (`is_capital`) anchor'ını kullanır, `world_x/world_y` ise bölge geometrisi için korunur.

Edit mode'da `world_x/world_y` merkezleri ayrı işaretlerle çizilir. Shift + sol sürükleme bu koordinatları değiştirir; Voronoi sınırları `WorldMap` raster cache'ine bağlı olduğu için sürükleme sırasında sadece merkez işareti güncellenir, fare bırakıldığında cache bir kez yeniden oluşturulur.

---

## Input Yönetimi

`HandleInput()` döner: `InputAction{Kind, ArmyID, TargetRegion, TargetFaction, BuildingID, Delta}`

**Just-pressed takibi:** `prevKeys`, `prevMouse` map'leri tutulur; `keyJustPressed()` / `mouseJustPressed()` bir frame'lik tetikleme sağlar.

**Tık öncelik sırası:**
1. Açık detay paneli kapatma düğmeleri (bölge/ordu)
2. Alt-orta aksiyon HUD butonları (diplomasi, teknoloji, tur bitir)
3. Olay logu akordiyonu: başlık butonu paneli daraltır/genişletir, kart X'i olayı kapatır, kart gövdesi detay popup açar
4. UI bölgesi (üst-sol durum paneli / sağ-üst tarih-menü HUD / alt-orta aksiyon HUD / sağ panel) → geçersiz say
5. Bölge paneli aksiyonları: vergi +/- düğmeleri, oluşturulabilir bina kartına tıklayarak inşa
6. Birim oluştur paneli (`recruit_panel.go:RecruitPanelHitTest`); kıyı olmayan bölgelerde deniz birimleri gösterilmez
7. Bölge/birim oluştur paneli boş alan tıklamaları → tüketilir, arkadaki haritaya düşmez
8. BÖL/BİRLEŞTİR butonları (seçili ordu varsa, `army_panel.go` hit-test)
9. Ordu ikonuna tıklama — `armyIconPositions()` üzerinden offset'li 14px yarıçap
10. Bölge seçimi (WorldMap pixel lookup)

Edit mode'da oyun HUD/panelleri çizilmez; harita, minimap, üst edit HUD ve alt-sol bilgi HUD'u görünür. Sol tık settlement, bölge veya ordu seçer; settlement sürükleme koordinatı canlı taşır ve başka kara region'a bırakılan settlement o region'ın `settlements[]` listesine aktarılır. Alt + sol tık tıklanan kara bölgeye yeni settlement ekler, Delete seçili settlement'ı siler. Shift + sol sürükleme kara bölgenin `world_x/world_y` merkezini taşır ve fare bırakıldığında harita cache'ini yeniler. Alt-sol HUD'daki `Yerlesim Ekle`, `Tip`, `Ana Yap`, `Isim`, `Arazi`, `Sahip`, `Sil` ve `Kaydet` butonları aynı işlemleri doğrudan çalıştırır. `Tip`, `Arazi` ve `Sahip` inspector yanında kaydırılabilir dropdown açar; seçilen satır ilgili `type`, `terrain` veya `owner_id` değerini doğrudan yazar. F2/Enter seçili settlement adını düzenler, Ctrl+S `ActionSaveScenario` üretir.

Voronoi debug overlay `V` ile açılıp kapanır. Overlay `WorldMap.BoundaryPixels` ile seçili veya hover bölgenin gerçek raster sınırını camgöbeği piksellerle çizer. `WorldMap.VisualNeighbors` üzerinden raster sınır komşularını çıkarır ve JSON `neighbors` listesiyle karşılaştırır: yeşil çizgi görsel+JSON komşu, kırmızı çizgi sadece görsel komşu, gri çizgi sadece JSON komşudur. Sağ üst panel hover pixel'in `RegionAt` sonucunu, senaryo koordinatını ve seçili bölgenin visual/json komşu sayısını gösterir.

Edit mode'da `editDirty` true iken ESC doğrudan çıkmaz; genel onay modalı üç seçenekle açılır: `Kaydet` önce `ActionSaveScenarioAndGoMainMenu` üretir, kayıt başarılıysa ana menüye döner; `Kaydetmeden Cik` doğrudan `ActionGoMainMenu` üretir; `Iptal` modalı kapatır.

Menü ve üst paneller fareyle tamamlanabilir: senaryo/fraksiyon/zafer ve kayıt ekranlarında `Geri` düğmesi vardır; diplomasi ve teknoloji panelleri X düğmesiyle kapanır; kayıt silme onayı kart içi `Sil`/`İptal` düğmeleriyle yapılır. Ayarlar ekranında müzik/ses efektleri aç-kapat ve her ikisi için `0-100` arası ayrı seviye bulunur. Paylaşılan efektler `assets/sounds/` altından yüklenir; senaryo müziği `scenario.json` içindeki `music.default_playlist` ile başlar ve dosyaları senaryo `musics/` klasöründen okur. Oyun içi müzik HUD'u aktif parçayı gösterir ve `Dur/Cal` ile `Sonr` kontrollerini sunar; ESC menüsünde müzik aç/kapat ve müzik seviyesi hızlıca değiştirilebilir.

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

Aynı bölgede birden fazla ordu bulunabilir. `armyIconPositions()` (`renderer.go`) tüm orduları `RegionID`'ye göre gruplar, her grubun ikonlarını 26px aralıklarla yatayda ortalar. Hem `drawArmies` hem `handleLeftClick` hem de `cursor.go:inGameHovering` bu tek fonksiyonu kullanır — tutarsızlık riski yoktur.

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
