---
name: mapp-ui-guidelines
description: Mapp UI ve render kodunu değiştirirken ortak widget, geometri, input ve state-render sınırlarını koru.
applyTo: "internal/ui/**/*.go,internal/render/**/*.go"
---

# UI geliştirme kuralları

- Ayrıntılı ekran rehberi için [`wiki/architecture/ui-screen-guide.md`](../../wiki/architecture/ui-screen-guide.md), framework durumu için [`wiki/architecture/ui-framework.md`](../../wiki/architecture/ui-framework.md) dosyasını kaynak kabul et.
- Yeni veya değiştirilen panel, modal, overlay, buton, liste, etiket ve form yüzeylerinde önce `internal/ui` primitive'lerini ve `internal/render` ortak helper'larını ara; aynı deseni elle kopyalama.
- Bir etkileşimli yüzey için tek geometry/rect kaynağı kullan. Draw, ölçüm, hover, cursor, hit-test ve input dispatch aynı geometriyi tüketmelidir.
- Modal veya overlay arka harita etkileşimini tüketmeli; input'u hem eski hem yeni yoldan işleyen çift path bırakma.
- Oyun kuralı, ekonomi, diplomasi, hareket veya state mutasyonunu çizim koduna koyma. Render sorgulasın; değişiklikleri oyun/state akışına yönlendir.
- `Update` ve `Draw` sıcak yollarında gereksiz image, slice/map veya widget allocation oluşturma; asset ve statik geometriyi uygun yerde cache'le.
- UI değişikliklerinde en azından ilgili paket testlerini çalıştır; geometri veya input davranışı değişirse headless geometri/regresyon testi ekle.
- UI değişikliği mimari davranışı etkiliyorsa ilgili wiki sayfasını ve özellik tamamlandıysa `wiki/dev/progress.md` dosyasını güncelle.
