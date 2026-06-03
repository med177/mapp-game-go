---
type: architecture
tags: [ui, widgets, screens, guide]
last_updated: 2026-06-03
related: [architecture/ui-framework, architecture/render-pipeline, dev/progress]
---

# UI Screen Guide

## Amaç
Yeni ekran veya panel eklerken manuel hitbox, ayrı cursor hesabı ve kopya geometri üretmeden ortak `internal/ui` hattını kullanmak.

## Kural Seti
1. Input frame başına bir kez toplanmalı.
2. Aynı yüzey için draw, hover ve click geometri kaynağı tek olmalı.
3. Modal ekranlar arka harita etkileşimini tüketmeli.
4. Ekran dosyasında ham `mx >= x && ...` zinciri son çare olmalı.
5. Yeni etkileşimli yüzeyler mümkünse `internal/ui` primitive’leriyle temsil edilmeli.

## Tercih Edilen Primitive'ler
1. `Button`
2. `Panel`
3. `Label`
4. `Dropdown`
5. `ListView`
6. `Checkbox`
7. `RadioGroup`
8. `Modal`
9. `Overlay`
10. `TextBox`
11. `Tooltip`
12. `Image` / `Icon`
13. `AnchorRect`

## Uygulama Akışı
1. Geometri builder yaz:
   - örnek: `buildFooPanel()`, `buildFooButtons()`
2. Draw kodunu bu builder çıktılarından besle.
3. Cursor hover kararını aynı builder `.HitTest()` ile ver.
4. Click/input akışını aynı builder `.HitTest()` ile çöz.
5. Arka plan etkileşimi kapanması gerekiyorsa `Panel` veya `Modal` hit-test’ini erken tüket.
6. Klavye focus gerekiyorsa `internal/ui.Manager` ve `Focusable` sırasını kullan.

## Örnek Desen
```go
func buildExampleModal() gameui.Modal {
	panel := gameui.NewPanel(320, 180, 420, 180)
	return gameui.NewModal(ScreenWidth, ScreenHeight, panel)
}

func buildExampleButtons() (gameui.Button, gameui.Button) {
	modal := buildExampleModal()
	return gameui.NewButton(modal.Panel.Rect.X+40, modal.Panel.Rect.Y+120, 140, 32, "Onayla"),
		gameui.NewButton(modal.Panel.Rect.X+240, modal.Panel.Rect.Y+120, 140, 32, "İptal")
}
```

## Test Beklentisi
1. Yeni primitive eklendiyse `internal/ui` test ekle.
2. Render seam değiştiyse en az:
   - `go test ./internal/ui ./internal/render ./internal/game`
3. Ekran ailesi geniş etkiliyse:
   - `go test ./...`
4. Çözünürlük riski varsa headless geometri smoke testi ekle.
5. Ana ekran çizimi değiştiyse headless draw-call smoke testi ekle; piksel okuma Ebitengine'de oyun başlamadan çalışmaz.
6. Sıcak UI builder değiştiyse `testing.AllocsPerRun` ile allocation eşiği koy.

## Ne Zaman Yeni Primitive Eklenmeli
1. Aynı çizim/hit-test deseni en az iki yerde tekrar ediyorsa
2. Modal veya overlay ailesi ayrı ayrı kopyalanıyorsa
3. Metin/panel/button grubu yeni ekranlarda tekrar kullanılıyorsa

## Ne Zaman Render-Spesifik Kalabilir
1. Saf görsel debug overlay
2. Piksel bazlı brush preview
3. Harita üstü yoğun, frame-özel çizim katmanları

Bu alanlarda bile mümkünse hit-test ve kapatma davranışı ortak UI yüzeyleriyle ayrılmalı.
