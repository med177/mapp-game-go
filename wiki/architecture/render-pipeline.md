---
type: architecture
tags: [render, ebitengine, camera, input, ui]
last_updated: 2026-05-08
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
| 4 | Bölge etiketleri + şehir noktası | `renderer.go` |
| 5 | Ordu ikonları | `renderer.go` |
| 6 | UI panelleri (üst-sol durum paneli, sağ-üst tarih/menü HUD, alt-orta aksiyon HUD, bölge/ordu/minimap/event log) | `panel.go` |
| 6 | Ordu detay paneli — 20 slot ızgarası, boş slotlar silik | `army_panel.go` |
| 6 | Olay logu akordiyonu — daralt/genişlet, wrap edilmiş kartlar, X ile kapatma, tıklayınca detay popup | `panel.go`, `renderer.go` |
| 7 | Diplomasi paneli (Tab) — tam ekran overlay | `diplom.go` |
| 8 | Teknoloji paneli (T) — tam ekran ağaç görünümü | `tech_panel.go` |
| 9 | Info popup bildirimi (combatLog, olay loguna yazmaz) | `renderer.go`, `panel.go` |
| 10 | Savaş ilan onay diyalogu | `renderer.go` |
| 11 | Tarihsel olay popup | `renderer.go` |

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

Menü ve üst paneller fareyle tamamlanabilir: senaryo/fraksiyon/zafer ve kayıt ekranlarında `Geri` düğmesi vardır; diplomasi ve teknoloji panelleri X düğmesiyle kapanır; kayıt silme onayı kart içi `Sil`/`İptal` düğmeleriyle yapılır.

Uzun sürebilen senaryo/kayıt yükleme işleri `PhaseLoading` ekranına geçer. `internal/game/game.go` yükleme işini arka planda başlatır; renderer bu sırada `loading.go` içindeki gerçek zaman tabanlı spinner'ı çizer ve sonuç hazır olduğunda state ana thread üzerinde uygulanır.

Düşman orduları seçilebilir ama emir verilemez. Renderer düşman ordusu için hareket hedefi çizmez ve sağ/sol tık hareket aksiyonu üretmez. Oyuncu ordularından birinin mevcut hareket menzilindeki düşman ordularda ikon birim sayısını gösterir; detay panelinde birimlerin yaklaşık yarısı görünür, kalanları `Gizli` kartlarıyla saklanır. Menzil dışındaki düşman ordularda birim sayısı ve hareket/birim detayları gizli kalır.

Bina ve birim kartlarında hover tooltip vardır. Tooltip maliyet, gereksinim, temel etki/istatistik ve kart görselini gösterir. Bölgeye uygun olmayan bina kartları render edilmez; liman son sıradadır ve kıyı olmayan bölgelerde görünmez.

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
