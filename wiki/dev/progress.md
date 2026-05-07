---
type: dev
tags: [progress, status, todo, known-issues]
last_updated: 2026-05-07
related: [HOME, architecture/render-pipeline]
---

# Geliştirme Durumu

## Tamamlanan Sistemler ✅

| Sistem | Durum | Notlar |
|---|---|---|
| Ebitengine kurulum | ✅ | `cmd/game/main.go`, 60 TPS |
| GameState merkezi yapı | ✅ | `internal/state/state.go` |
| Phase state machine | ✅ | 8 fase, tam geçiş |
| Harita render (2D) | ✅ | WorldMap cache, poligon doldurma |
| Bölge sistemi | ✅ | JSON'dan yükleme, komşuluk grafı |
| Fraksiyon sistemi | ✅ | 9 fraksiyon, renk, din |
| Ordu hareketi | ✅ | Komşuluk kısıtı, naval/kara ayrımı |
| Çarpışma motoru | ⚠️ | `calculateOutcome()` taslak — doldurul mast |
| Ekonomi tick | ✅ | Vergi geliri, ticaret güzergahları |
| Bina inşası | ✅ | 6 bina tipi, maliyet, kısıt |
| Teknoloji ağacı | ✅ | Araştırma sayacı, efekt hesabı |
| Mevsim mekaniği | ✅ | Kış hasarı, sonbahar bonusu |
| Diplomasi | ✅ | 4 duruş, puanlama, decay |
| AI turu | ✅ | Temel hareket mantığı, koalisyon (zor) |
| Tarihsel olaylar | ✅ | JSON tetikleyici, tek-seferlik |
| Zafer koşulları | ✅ | 4 tip, Check() döngüsü |
| Kayıt/yükleme | ✅ | 4 slot (autosave + slot1-3), slot silme, metadata önizleme |
| Ana menü / ayarlar | ✅ | Fraksiyon & zafer seçim ekranı; "Kayıttan Yükle" slot seçim ekranına açılır |
| Pause menüsü | ✅ | ESC ile açılır; Devam Et / Kaydet / Yükle / Ana Menü / Çıkış |
| Ordu detay paneli (20 slot) | ✅ | Boş slotlar silik, HP çubuğu, deneyim çubuğu, BÖLDÜR butonu |
| Askeri kapasite (manpower) sistemi | ✅ | Bölge başı 5 + kışla +5 birim; ordu sayısı = ceil(bölge/2); `state.go:ManpowerCap` |
| Ordu birleşme (merge) | ✅ | Dost bölgeye taşınınca ≤20 birimse otomatik birleşir; `game.go:tryMergeArmies` |
| Ordu bölme (split) | ✅ | "✂ BÖLDÜR" butonu, birimleri ikiye böler; `game.go:splitArmy` |
| Çoklu ordu yan yana render | ✅ | `renderer.go:armyIconPositions()` — aynı bölgedeki ordular 26px aralıkla |
| Minimap | ✅ | Sağ alt köşe, kamera konumu, ordu konumları gösterimi |
| Tüm menülerde parmak imleci | ✅ | Pause, load/save, settings, in-game dahil tüm fazlar |
| Vergi ayarlama | ✅ | . / , tuşları, ±5% |
| Deniz birimi | ✅ | Nakliye gemisi, liman koşulu |

---

## Eksik / Planlanan ⬜

| Özellik | Öncelik | Notlar |
|---|---|---|
| `calculateOutcome()` implementasyonu | 🔴 Kritik | Combat sistemi çalışmıyor |
| Din ceza/bonus sistemi | 🟡 Orta | Veri hazır, mantık yok |
| AI teknoloji araştırma kararı | 🟡 Orta | Sadece ordu hareketi var |
| AI bina inşa stratejisi | 🟡 Orta | |
| Görsel mevsim değişimi | 🟢 Düşük | Arka plan swap |
| İkincil mal döngüsü | 🟢 Düşük | Tahıl/demir/kereste |
| Olay zincirleme | 🟢 Düşük | |
| Oyuncu tepki seçenekleri (A/B) | 🟢 Düşük | |
| `religion/` paketi ayrıştırması | 🟢 Düşük | game.go'da inline |
| Sprite'lar | 🟢 Düşük | Şu an renkli poligonlar |
| Ses efektleri | 🟢 Düşük | |
| ~~Tek seferlik kayıt → çoklu slot~~ | ✅ | Tamamlandı (2026-05-07) |

---

## Bilinen Sorunlar 🐛

| Sorun | Dosya | Açıklama |
|---|---|---|
| `calculateOutcome` taslak | `internal/combat/combat.go:63` | İskelet var, uygulama yok |
| `religion/diplomacy` paket eksikliği | `internal/game/game.go` | Logic inline, paket yok |
| `game.exe` kök dizinde geçici | `game.exe` | `bin/game.exe` kalıcı olmalı |

---

## Araçlar (tools/)

| Araç | Amaç |
|---|---|
| `tools/centroids/main.go` | Bölge merkez koordinatları hesapla |
| `tools/populate_all_shapes.py` | Natural Earth'ten poligon üret |
| `tools/update_shapes_from_ne.py` | Şekilleri güncelle |
| `tools/fix_*.py` | Belirli bölge düzeltmeleri |
| `tools/add_regions*.py` | Yeni bölge ekleme |
| `tools/add_missing_countries.js` | Eksik ülke tamamlama |
