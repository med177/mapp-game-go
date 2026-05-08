# AI Agent Environment & Capability Rules

**CRITICAL INSTRUCTION FOR ALL AGENTS:** 
Identify your environment before taking any action. Read the section below that applies to your operating environment.

---

## 1. Rules for Claude, Copilot, and Windows-based Agents
If you are operating directly on the **Windows host OS** (via VS Code, Cursor, Windsurf, or Visual Studio):
* **Environment:** You are in a Windows environment (PowerShell / CMD).
* **Paths:** Use standard Windows file paths (e.g., `C:\...`).
* **Execution:** Do NOT attempt to run Linux-specific bash scripts or WSL binaries. 
* **Focus:** Focus on standard code generation, IDE integrations, and cross-platform Node.js/Python scripts if automation is needed.

---

## 2. Rules for Codex (WSL / Ubuntu Sandbox)
If you are **Codex**, you are operating within an **Ubuntu WSL sandbox**. You have access to a specialized set of high-performance CLI tools. 

**CRITICAL RULE FOR CODEX:** Do NOT attempt to write custom Python or Bash scripts for tasks that can be natively handled by the tools listed below.

### A. Search, Navigation & System
* **`rg` (ripgrep):** Fast text and code searching.
* **`fd` (fd-find):** Use instead of `find` for file discovery.
* **`tree`:** Visualize directory structures.
* **`tldr`:** Run `tldr <command>` if you forget syntax.

### B. Data Parsing
* **`jq` & `yq`:** Use exclusively for parsing JSON and YAML files or API responses. 

### C. Networking & APIs
* **`curl` & `httpie`:** Test REST APIs or download resources. Prefer `http` for readable outputs.

### D. Databases & Caching
Connect directly to services to inspect schemas or test queries. Do not guess structures:
* **`postgresql-client` (psql):** For PostgreSQL / Supabase operations.
* **`mysql-client`:** For MySQL operations.
* **`redis-tools` (`redis-cli`):** For cache inspection.

### E. Development & Build Tools
* **`git`:** Version control and diffs.
* **`build-essential`:** C/C++ compilers and `make`.
* **`python3-venv`:** Create isolated Python virtual environments if needed.
* **`golang-go`:** Build or test Go services.

### F. Infrastructure & Mobile Debugging
* **`rclone`:** Use for cloud synchronization and backups.
* **`adb` (Android Debug Bridge):** Connect to emulators/devices. Pull `logcat` to debug mobile apps, analyze crashes, or inspect Bluetooth/plugin communications (e.g., ESC/POS printers) directly from the terminal.

### Operating Guidelines (Codex Only)
1. **Explore First:** Use `tree` or `fd` to understand context before making changes.
2. **Execute Directly:** Run `psql`, `adb`, or `rg` commands directly in the shell to gather information instead of asking the user.
3. **Fail Fast:** If a command fails, read the error output and adjust your syntax immediately.

# Role & Context
You are an expert game software engineer specializing in Go (Golang) and the Ebitengine (2D game engine). 
We are developing a Grand Strategy game featuring a "Total War" style topographical campaign map. 
The development environment is WSL (Ubuntu) on Windows.

# Technical Stack & Core Architecture
- Language: Go (strict typing, idiomatic conventions).
- Engine: Ebitengine (github.com/hajimehoshi/ebiten/v2).
- Domain: Grand strategy mechanics, topographical/isometric campaign maps, node/tile-based logic, pathfinding algorithms (A*, Dijkstra), and Entity-Component-System (ECS) or strictly decoupled state management.

# Strict Coding & Engine Guidelines
1. Performance & GC Optimization: Ebitengine's `Update` and `Draw` loops run continuously (typically 60 TPS). You MUST avoid dynamic memory allocations (e.g., creating new slices, structs, or strings) inside these loops to prevent Garbage Collection (GC) stutters. Reuse objects and pre-allocate buffers.
2. Ebitengine Best Practices: 
   - Cache `ebiten.Image` objects. Never load or create images within the game loop.
   - Batch draw calls where possible using `ebiten.DrawImageOptions`.
   - Calculate topographical map data, coordinates, and pathfinding asynchronously or cache them; do not recalculate static spatial data on every frame.
3. Separation of Concerns: Strictly separate Game State (business logic, map data, troop movements) from the View (rendering logic). The map data structure (grids, graphs, regions) must be independent of the Ebitengine drawing operations.
4. Environment Awareness: Code must compile and run flawlessly within WSL (Ubuntu). Avoid Windows-specific syscalls or paths unless guarded by Go build tags.

# Output Format & Behavior
- When writing code, provide context-aware, production-ready Go snippets.
- Briefly explain *why* a specific data structure (e.g., 1D array vs. 2D slice for spatial maps, spatial hashing for collision) was chosen for performance.
- Assume the user is an experienced developer. Skip basic Go tutorials. Focus on advanced game architecture, optimal mathematical calculations for map grids, and Ebiten-specific rendering techniques.

# Mapp Game Go (Proje Hafızası)

## Proje Özeti
Orta Çağ temalı (1300–1600) sıra tabanlı strateji oyunu.
Total War serisinin kampanya haritasına benzer oynanış — taktik savaş sahnesi yok,
tüm çarpışmalar harita üzerinde otomatik hesaplanır.
Tek oyunculu, karşısında stratejik yapay zeka var.

---

## Teknoloji Yığını
| Bileşen | Seçim | Neden |
|---|---|---|
| Dil | Go | Performans, sadelik, derleme hızı |
| Render | Ebitengine v2 | Saf Go, Windows desteği mükemmel, 2D/2.5D ideal |
| Veri Formatı | JSON | Bölge, birim, teknoloji, olay verileri için |
| Kayıt | JSON serialize | İnsan okunabilir kayıt dosyaları |
| Hedef Platform | Windows (önce), cross-platform mimari |

---

## Oyun Dünyası

### Harita
- **Kapsam:** Akdeniz havzası merkez; İngiltere kuzey sınır, Rusya doğu, İran/Safevi güneydoğu.
- **Dönem:** 1300–1600 (Osmanlı kuruluş ve yükseliş dönemi odak).
- **Bölge Sistemi:** Dinamik — oyun ilerledikçe yeni bölgeler keşfedilip açılır.
- **Perspektif:** İzometrik 2.5D.
- **Arazi Tipleri:**
  - Ova (serbest geçiş)
  - Orman (görüş kısıtlı, yavaş geçiş)
  - Dağ (geçilemez blok, sadece dar geçitler)
  - Deniz (sadece deniz kuvvetleriyle)
  - Geçit (dağlar arası tek yol → pusu noktası)
- **Görüş Açısı:** Arazi tipine göre kısıtlı; ormanda/dağda düşman görülmez, geçit yolunda pusu kurulabilir.

### Mevsimler
- 1 tur = 1 ay (12 tur = 1 yıl).
- Mevsimler haritada görsel olarak değişir (kar, yeşil, vb.).
- **Kış cezası:** Ordular her kış ayı birim kaybeder (soğuk hasarı).
- **İlkbahar bonusu:** Hareket bonusu.
- **Yaz:** Normal.
- **Sonbahar:** Hasat — vergi geliri artar.

---

## Fraksiyonlar

### Oynanabilir Başlangıç Fraksiyonları
| Fraksiyon | Din | Bölge | Özel Not |
|---|---|---|---|
| Osmanlı | Sünni İslam | Anadolu/Balkanlar | Başlangıç odak fraksiyonu |
| Venedik | Katolik | Kuzey İtalya / Adalar | Ticaret odaklı |
| Fransa | Katolik | Batı Avrupa | |
| İngiltere | Katolik | Britanya | |
| Memlük | Sünni İslam | Mısır/Levant | |
| Safevi | Şii İslam | İran/Azerbaycan | Osmanlı'nın doğu rakibi |
| Rusya | Ortodoks | Doğu Avrupa | |
| Aragon | Katolik | İber/Akdeniz | |
| Portekiz | Katolik | İber/Atlas kenarı | Deniz ticareti |

### Din Sistemi
- **Mezhepler:** Katolik, Ortodoks, Sünni İslam, Şii İslam.
- Aynı mezhep → diplomatik ilişki bonusu, ittifak kolaylığı.
- Farklı mezhep → diplomatik ceza, savaş ilanı kolaylaşır.
- **Mezhep değişimi:** Ele geçirilen bölgede yıllara yayılır, isyan riski doğurur.

---

## Diplomasi Sistemi
- **Eylemler:** İttifak kur, Savaş ilan et, Barış müzakere et, Ticaret anlaşması imzala.
- Fraksiyon ilişkileri puanlama sistemi (-100 düşman → +100 müttefik).
- Ortak düşman → ilişki bonusu.
- Din farkı → kalıcı ceza çarpanı.
- Ele geçirilmiş bölgeler → komşu fraksiyonlara tehdit algısı.

---

## Ekonomi

### Kaynaklar
- **Birincil:** Düka Altın (her şey altına çevrilebilir).
- **İkincil mallar:** Tahıl, Demir, Kereste, Baharat, Kumaş.
- Mallar ticaret anlaşmalarıyla fraksiyonlar arası satılabilir.
- **Dönüşüm:** Mallar belirli oranda altına çevrilir (piyasa fiyatı dalgalanabilir).

### Şehir & Kale Geliştirme
- Her bölgede bir merkez şehir/kale var.
- Binalar üretim ve kapasiteyi artırır:
  - Pazar → ticaret geliri
  - Çiftlik → tahıl üretimi
  - Demirci → demir, birlik kalitesi
  - Kışla → ordu eğitim hızı
  - Liman → deniz birimi üretimi, ticaret kapasitesi
  - Katedrals/Cami/Kilise → din etkisi, halk memnuniyeti
  - Surlar → savunma bonusu

### Vergi Sistemi
- Her bölgede vergi oranı ayarlanabilir (0–100%).
- Düşük vergi → yüksek halk memnuniyeti, yavaş üretim kayıpları yok.
- Yüksek vergi → yüksek altın, isyan riski, üretim düşüşü.
- **İsyan mekaniği:** Memnuniyet eşiği altına düşerse bölge isyan edebilir → ordu olmadan kontrol kaybolur.

---

## Ordu Sistemi

### Kara Birlikleri (3 kategori × 3 çeşit)
| Kategori | Temel | Orta | Elit |
|---|---|---|---|
| Piyade | Milis | Piyade | Yeniçeri/Şövalye |
| Süvari | Hafif Süvari | Süvari | Ağır Süvari |
| Topçu | Mancınık | Bombarda | Top |

### Deniz Birlikleri (3 çeşit)
- Savaş gemisi (muharebe)
- Nakliye gemisi (ordu taşıma)
- Ticaret gemisi (pasif gelir)

### Ordu Mekaniği
- Bir ordu en fazla **20 birim** taşır.
- Harita üzerinde ikon (taş/mühür görünümü) olarak temsil edilir.
- **Çarpışma:** Otomatik hesaplanır. Faktörler:
  - Birim sayısı × birim gücü
  - Arazi tipi çarpanı (savunma bonusu)
  - Hücum/savunma durumu
  - Komutan bonusu (ileride)
  - Mevsim cezası/bonusu
  - Pusu (geçit noktasında hazır bekleyen = ciddi atak bonusu)

### Hareket
- Belirli yollar/güzergahlar üzerinden ilerlenebilir (Total War sefer haritası gibi).
- Arazi tiplerine göre hareket puanı tüketilir.
- Dağ geçitleri tek yol → stratejik tıkama noktaları.

---

## Teknoloji Ağacı
- Araştırmalar tur harcamasıyla ilerler (altın veya üretim puanı).
- **Bina bağımlılığı:** Bazı teknolojiler belirli binaların varlığını gerektirir.
- **Bölge bağımlılığı:** Belirli şehirler/bölgeler ele geçirilince yeni teknoloji dalları açılır (örn: Konstantinopolis → Bizans mühendisliği).
- Kategoriler: Askeri, Ekonomi, Diplomasi, Denizcilik, Din/Kültür.

---

## Olaylar Sistemi
- Olaylar **gerçek tarihe yakın** tetiklenir (tarih ve bölgeye göre).
- **Olay Türleri:**
  - Veba (bölge nüfus/üretim düşüşü, komşulara yayılabilir)
  - Kıtlık (tahıl üretimi sıfır, isyan riski)
  - Taht krizi (fraksiyon içi isyan veya zayıflık)
  - Suikast (lider/komutan kaybı)
  - Dini hareketler (Reformasyon, mezhep çatışması)
  - Keşif olayları (yeni bölge açılımı tetikleyici)
- Olaylar bölge ve harita üzerinde net ikonla gösterilir.

---

## Zafer Koşulları (Oyun başında seçilir)
| Tip | Koşul |
|---|---|
| Toprak Hakimiyeti | X tarihe kadar belirli kritik bölgelerle birlikte 20+ bölge tut |
| Ekonomik Güç | Y miktarı altın gelire ulaş ve 5 tur boyunca koru |
| Askeri Üstünlük | Z adet ordu birimi oluştur, 3 büyük fraksiyon yenilgisi |
| Dinî Zafer | Kutsal şehirleri (Kudüs, Roma, Mekke) aynı anda tut |

### Kaybetme
- Son bölge düşene kadar oyun bitmez (son şans mekaniği).

---

## Yapay Zeka
- **Strateji:** Önce kendinden zayıf komşuları hedefler.
- **Fırsatçı:** Rakip fraksiyonun isyanı/zayıflığı → saldırı fırsatı.
- **Ekonomik:** Kaynak bölgeleri öncelikli ele geçirir.
- **Diplomatik:** Tehdit altındayken ittifak kurar.
- **3 Zorluk Seviyesi:**
  - Kolay: Yavaş büyüme, pasif AI.
  - Normal: Dengeli strateji.
  - Zor: Kaynak bonusu + agresif koalisyon kuruluyor.

---

## Kayıt Sistemi
- JSON tabanlı kayıt dosyaları (`saves/` klasörü).
- Birden fazla kayıt slotu.
- Kaldığı turdan devam etme.
- Otomatik kayıt (her tur sonu opsiyonel).

---

## Proje Dizin Yapısı
```
mapp-game-go/
├── CLAUDE.md
├── AGENTS.md
├── go.mod
├── go.sum
├── game.exe               # Kök dizindeki geçici build çıktısı
├── cmd/
│   └── game/
│       └── main.go        # Uygulama giriş noktası
├── bin/
│   └── game.exe           # Kalıcı build çıktısı
├── assets/
│   ├── maps/              # Harita görselleri
│   │   ├── world_map_background.png
│   │   ├── mini-map.png
│   │   └── debug_alignment*.png
│   └── data/
│       ├── regions.json
│       ├── factions.json
│       ├── units.json
│       ├── technologies.json
│       ├── buildings.json
│       ├── events.json
│       └── generated/
│           └── country_shapes.json   # Ülke poligon verileri (üretilmiş)
├── internal/
│   ├── game/              # Ana oyun döngüsü, tur yönetimi
│   │   ├── game.go
│   │   └── resolution.go
│   ├── state/             # Merkezi oyun durumu (GameState)
│   │   └── state.go
│   ├── world/             # Harita, bölge, arazi, görüş
│   │   ├── region.go
│   │   ├── terrain.go
│   │   └── loader.go
│   ├── faction/           # Fraksiyon verisi, ilişkiler
│   │   ├── faction.go
│   │   └── loader.go
│   ├── army/              # Ordu, birlik, hareket
│   │   ├── army.go
│   │   ├── unit.go
│   │   └── loader.go
│   ├── combat/            # Çarpışma hesaplama motoru
│   │   └── combat.go
│   ├── economy/           # Kaynak, vergi, ticaret
│   │   └── economy.go
│   ├── city/              # Şehir, kale, bina sistemi
│   │   └── building.go
│   ├── tech/              # Teknoloji ağacı
│   │   └── tech.go
│   ├── events/            # Tarihsel olaylar motoru
│   │   └── events.go
│   ├── ai/                # Yapay zeka stratejisi
│   │   └── ai.go
│   ├── season/            # Mevsim mekaniği
│   │   └── season.go
│   ├── victory/           # Zafer koşulları kontrolü
│   │   └── victory.go
│   ├── save/              # Kayıt/yükleme
│   │   └── save.go
│   └── render/            # Ebitengine render katmanı
│       ├── renderer.go    # Ana render döngüsü
│       ├── assets.go      # Görsel varlık yükleyici
│       ├── font.go        # Font yönetimi
│       ├── tile.go        # Harita tile render
│       ├── mapgen.go      # Harita üretimi/şekiller
│       ├── panel.go       # UI panelleri
│       ├── cursor.go      # Fare imleci
│       ├── action.go      # Oyuncu aksiyonları UI
│       ├── diplom.go      # Diplomasi paneli
│       ├── faction_select.go  # Fraksiyon seçim ekranı
│       ├── main_menu.go   # Ana menü
│       ├── settings.go    # Ayarlar paneli
│       ├── tech_panel.go  # Teknoloji ağacı paneli
│       └── victory_select.go  # Zafer koşulu seçimi
├── tools/                 # Harita/veri üretim araçları (Python/JS)
│   ├── centroids/
│   │   └── main.go        # Bölge merkezleri hesaplama (Go)
│   ├── add_regions*.py
│   ├── populate_all_shapes.py
│   ├── update_shapes_from_ne.py
│   ├── fix_*.py
│   ├── extract_islands.py
│   ├── debug_alignment.py
│   └── add_missing_countries.js
├── _REFERENCE/            # Tasarım referans görselleri ve şekil verileri
│   ├── *.png / *.jpg / *.webp
│   └── ne_10m_admin_0_countries/   # Natural Earth ülke sınırları
└── saves/                 # Oyun kayıt dosyaları
```

---

## Geliştirme Öncelik Sırası
1. Proje iskelet + Ebitengine kurulumu
2. Harita render motoru (izometrik grid, arazi tipleri)
3. Bölge sistemi + fraksiyon sahipliği
4. Tur sistemi + mevsim mekaniği
5. Ordu hareketi (yol ağı üzerinde)
6. Çarpışma hesaplama
7. Ekonomi + vergi + bina
8. Diplomasi
9. Teknoloji ağacı
10. Din sistemi
11. Olaylar motoru
12. Yapay zeka
13. Zafer koşulları
14. Kayıt/yükleme
15. UI polish + ses

---

## Geliştirme Notları
- Kodlama dili: Go
- Render: Ebitengine v2 (`github.com/hajimehoshi/ebiten/v2`)
- Her sistem kendi `internal/` paketi altında izole edilmeli
- Veri (bölge, birim, olay) JSON dosyalarından okunmalı — hardcode edilmemeli
- Kayıt formatı da JSON (insan okunabilir, debug kolaylığı)
- AI için ilk aşamada `internal/ai` paketi soyut `Strategy` interface üzerinden çalışmalı

---

## Harita Voronoi Render Kuralları

### Temel İlke
Her bölgenin haritadaki renk alanı **saf koordinat mesafesine** göre belirlenir (`world_x`, `world_y`).
`internal/render/mapgen.go` → `nearestShapeRegion()`: her piksel için en yakın merkezi bulur.
**Komşu (neighbors) listesi render'ı ETKİLEMEZ** — sadece ordu hareketi ve ticaret için kullanılır.

### Voronoi Sınır Formülü
İki bölge A ve B arasındaki görsel sınır, her eksen için orta noktada geçer:
```
sınır_wy = (A.world_y + B.world_y) / 2
```
A, B'den **daha küçük wy** değerine sahipse (daha kuzeyde), A'nın Voronoi hücresi kuzeye uzanır.

### Bölge Denize/Kıyıya Ulaşıyor Sorunu
**Belirti:** Bir bölgenin rengi deniz/kıyı piksellerine kadar uzanıyor.
**Neden:** O bölge ile deniz arasında başka hiçbir bölge merkezi yok.
**Çözüm:** Araya yeni bir bölge ekle:
1. Sorunlu bölgenin `world_x` değerine yakın, `world_y` değerinden daha küçük (daha kuzey) bir nokta seç
2. Geçiş çizgisi = `(yeni_wy + sorunlu_wy) / 2` — bu değerin kıyı wy'sinden **büyük** olduğunu doğrula
3. Bölgeyi tarihsel olarak doğru bir fraksiyon/isimle ekle
4. Komşu listelerini **her iki yönde** güncelle
5. `neighbors` listesinden `_sea_black` / `_sea_aegean` gibi deniz komşularını kaldır (artık kıyıya ulaşamadığı için)

### Uygulama Örneği (1300 Anadolu)
```
Bithynia    (wx=1090, wy=468) → Karadeniz'e ulaşıyordu
Chrysopolis (wx=1090, wy=445) eklendi → geçiş: wy=456.5
=> wy<456: Chrysopolis (Bizans), wy>456: Bithynia (Osmanlı) ✓

Germiyan    (wx=1084, wy=498) → Akdeniz'e ulaşıyordu
Lycia       (wx=1082, wy=516) eklendi → geçiş: wy=507
=> wy<507: Germiyan, wy>507: Lycia (Teke) ✓
```

### Voronoi Komşuluk Testi
İki bölge A ve B'nin arasına C eklenecekse, A-B artık komşu **değildir**:
```
midpoint = ((A.wx+B.wx)/2, (A.wy+B.wy)/2)
d²(C, midpoint) < d²(A, midpoint) → C araya giriyor, A-B neighbors'tan kaldır
```

---

## Wiki Bakım Kuralları

Proje, `wiki/` dizininde Obsidian uyumlu bir LLM wiki'si tutar.

### Yapı
```
wiki/
├── HOME.md               ← Ana index
├── architecture/         ← Teknik mimari sayfaları
├── systems/              ← Oyun sistemleri (combat, economy, vb.)
├── world/                ← Dünya içeriği (bölgeler, fraksiyonlar)
└── dev/                  ← Geliştirme durumu, veri formatları
```

### Sayfa Formatı (zorunlu)
Her sayfa YAML frontmatter ile başlar:
```markdown
---
type: architecture | system | world | dev | index
tags: [tag1, tag2]
last_updated: YYYY-MM-DD
related: [other/page, systems/combat]
---
```
Sayfalar arası bağlantı: `[[systems/combat]]` (Obsidian wikilink sözdizimi).
Kod konumları: `internal/game/game.go:160` formatıyla belirt.

### Ne Zaman Güncellenir
| Değişiklik | Güncellenmesi gereken wiki sayfası |
|---|---|
| `internal/state/state.go` değişti | `architecture/state-management` |
| `internal/combat/combat.go` değişti | `systems/combat` |
| `internal/ai/ai.go` değişti | `systems/ai` |
| `internal/faction/faction.go` değişti | `world/factions`, `systems/diplomacy` |
| `internal/render/` değişti | `architecture/render-pipeline` |
| `internal/economy/economy.go` değişti | `systems/economy` |
| `internal/tech/tech.go` değişti | `systems/tech-tree` |
| `internal/victory/victory.go` değişti | `systems/victory` |
| `internal/season/season.go` değişti | `systems/seasons` |
| `internal/events/events.go` değişti | `systems/events` |
| Yeni sistem paketi eklendi | Yeni wiki sayfası aç + HOME.md'ye ekle |
| Özellik tamamlandı / bug çözüldü | `dev/progress.md` güncelle |

### Kurallar
- `last_updated` tarihini her düzenlemede güncelle
- Yeni bağlantı eklersen hedef sayfanın `related` listesine de ekle
- Tamamlanan özelliği `dev/progress.md`'de ✅ yap
- Eksik/planlanan özelliği `dev/progress.md`'de ⬜ listesine ekle
