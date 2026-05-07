---
type: dev
tags: [progress, status, todo, known-issues]
last_updated: 2026-05-07
# Combat sistemi güncellendi — calculateOutcome() implemente edildi
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
| Çarpışma motoru | ✅ | `calculateOutcome()` ±%15 rastgele dice, 4 sonuç kategorisi (ezici/dar zafer, yakın/ağır mağlubiyet) |
| Ekonomi tick | ✅ | Vergi geliri, ticaret güzergahları |
| Bina inşası | ✅ | 6 bina tipi, maliyet, kısıt |
| Teknoloji ağacı | ✅ | Araştırma sayacı, efekt hesabı |
| Mevsim mekaniği | ✅ | Kış hasarı, sonbahar bonusu |
| Diplomasi | ✅ | 4 duruş, puanlama, decay, din ilişkisi başlangıç bonus/ceza |
| AI teknoloji araştırma | ✅ | Askeri > ekonomi > deniz öncelikli, maliyet/tur optimize |
| AI ekonomi stratejisi | ✅ | Pazar (80% prio), çiftlik (düşük tahıl), sur (sınır) |
| AI deniz stratejisi | ✅ | Liman inşası, nakliye gemisi alımı (max 2 filo) |
| AI elite birim stratejisi | ✅ | Altın/teknolojiye göre: seçkin piyade > ağır süvari > piyade > süvari > milis |
| AI turu | ✅ | Stratejik hareket, koalisyon (zor), ittifak kurma |
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
| Ticaret güzergahları | ✅ | Fraksiyonlar arası pasif gelir, `TradeRoutes` |
| Din dönüşümü | ✅ | 24 turda bölge dini değişir, memnuniyet -20 |
| Din diplomasisi | ✅ | `ReligionRelation` aynı din +25, Sünni-Şii -40, farklı din -30 |

---

## Eksik / Planlanan 

| Özellik | Öncelik | Notlar |
|---|---|---|
| AI çoklu ordu konsolidasyonu | 🟡 Orta | AI orduları dağınık, ana ordu oluşturmuyor |
| AI uzun menzilli planlama | 🟢 Düşük | AI sadece komşu bölgelere hareket ediyor |
| Görsel mevsim değişimi | 🟢 Düşük | Arka plan swap |
| İkincil mal döngüsü | 🟢 Düşük | Tahıl/demir/kereste |
| Olay zincirleme | 🟢 Düşük | |
| Oyuncu tepki seçenekleri (A/B) | 🟢 Düşük | |
| `religion/` paketi ayrıştırması | 🟢 Düşük | `faction.go` ve `resolution.go` inline, ayrı paket yok |
| Ses efektleri | 🟢 Düşük | |
| ~~Tek seferlik kayıt → çoklu slot~~ | ✅ | Tamamlandı (2026-05-07) |

---

## Bilinen Sorunlar 🐛

| Sorun | Dosya | Açıklama |
|---|---|---|
| `game.exe` kök dizinde geçici | `game.exe` | `bin/game.exe` kalıcı olmalı |
| `religion/` paketi inline | `faction.go`, `resolution.go` | Din mantığı `faction` ve `game` paketlerine dağılmış |

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
