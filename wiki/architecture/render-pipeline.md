---
type: architecture
tags: [render, ebitengine, camera, input, ui]
last_updated: 2026-05-06
related: [game-loop, state-management]
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
}
```

---

## Draw Katman Sırası

`Draw(screen)` — `internal/render/renderer.go:166`

| Sıra | Katman | Dosya |
|---|---|---|
| 0 | Özel ekranlar (menü, ayarlar, seçim, game over) | `main_menu.go`, `settings.go`, `faction_select.go`, `victory_select.go` |
| 1 | Dünya haritası (WorldMap cache) | `mapgen.go`, `tile.go` |
| 2 | Seçim halkası (bölge) | `renderer.go` |
| 3 | Hareket hedefleri (ordu komşuları) | `renderer.go` |
| 4 | Bölge etiketleri + şehir noktası | `renderer.go` |
| 5 | Ordu ikonları | `renderer.go` |
| 6 | UI panelleri (alt bar, bölge paneli, ordu paneli, minimap, event log) | `panel.go` |
| 7 | Diplomasi paneli (Tab) | `diplom.go` |
| 8 | Teknoloji paneli (T) | `tech_panel.go` |
| 9 | Bildirim mesajı (combatLog) | `renderer.go` |
| 10 | Tarihsel olay popup | `renderer.go` |

---

## Kamera Sistemi

**Koordinat sistemi:** Dünya uzayı `(WorldW × WorldH)` px, ekran uzayına dönüşüm:

```
screenX = (worldX - camX + worldY * mapShearX) * camScale + ScreenWidth/2
screenY = (worldY - camY) * camScale * mapPitchY + ScreenHeight/2
```

`mapPitchY = 1.0`, `mapShearX = 0.0` → şu an düz 2D (izometrik bükme kapalı)

**Zoom:** Fare tekerleği ile 0.25–3.0 arası, fare pozisyonuna odaklanarak büyütür.

**Sürükleme:** Orta fare tuşu basılıyken dünya uzayı delta hesaplanır.

---

## WorldMap Cache

`WorldMap` — `internal/render/mapgen.go`

Harita, her fraksiyon sahipliği değişiminde `MarkDirty()` ile işaretlenir ve bir sonraki `Refresh()` çağrısında yeniden üretilir. Bölge poligonları `country_shapes.json`'dan gelir; renkler fraksiyon rengiyle doldurulur.

---

## Input Yönetimi

`HandleInput()` döner: `InputAction{Kind, ArmyID, TargetRegion, TargetFaction, BuildingID, Delta}`

**Just-pressed takibi:** `prevKeys`, `prevMouse` map'leri tutulur; `keyJustPressed()` / `mouseJustPressed()` bir frame'lik tetikleme sağlar.

**Tık öncelik sırası:**
1. Alt panel butonları (tıklandı mı?)
2. UI bölgesi (alt bar / sağ panel) → geçersiz say
3. Ordu ikonuna tıklama (14px yarıçap)
4. Bölge seçimi (WorldMap pixel lookup)

---

## Bildirim Sistemi

```
ShowCombatResult(msg)          → combatLogTimer = 180 frame (~3 sn), eventLog'a da ekler
ShowHistoricalEvent(title,desc) → tam ekran popup, herhangi tuş/tık ile kapatılır
```

`eventLog` maksimum 8 girdi tutar; yeniler öne eklenir, sondan taşanlar düşer.

---

## Dosya Sorumlulukları

| Dosya | İçerik |
|---|---|
| `renderer.go` | Kamera, draw döngüsü, input ana yönlendirici |
| `mapgen.go` | WorldMap cache, poligon doldurma |
| `tile.go` | Arazi renk/doku katmanı |
| `panel.go` | Alt bar, bölge/ordu/minimap/event log panelleri |
| `diplom.go` | Diplomasi paneli UI + input |
| `tech_panel.go` | Teknoloji ağacı paneli + input |
| `action.go` | `InputAction` tanımları |
| `font.go` | Font yükleme, `DrawText`, `MeasureText` |
| `assets.go` | Görsel varlık yükleme |
| `cursor.go` | İmleç şekli yönetimi |
| `faction_select.go` | Fraksiyon seçim ekranı |
| `victory_select.go` | Zafer koşulu seçim ekranı |
| `main_menu.go` | Ana menü |
| `settings.go` | Ayarlar ekranı |
