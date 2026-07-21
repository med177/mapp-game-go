---
type: system
tags: [technology, research, effects, tree]
last_updated: 2026-07-21
related: [systems/combat, systems/economy, architecture/state-management, dev/data-format]
---

# Teknoloji Ağacı

**Kaynak:** `internal/tech/tech.go`, `internal/tech/category_metadata.go`, `assets/data/technologies.json`

## Araştırma Yapısı

```go
type Technology struct {
    ID           string
    NameTR       string
    TurnsRequired int
    GoldCost     int
    Category     string        // military | economy | diplomacy | naval | culture
    Requires []string           // teknoloji ağacı bağımlılıkları
    Effects      TechEffects
}
```

---

## Araştırma Durumu (Faction içinde)

```go
type ResearchState struct {
    Completed   map[string]bool  // tamamlanan teknoloji ID'leri
    PausedTurns map[string]int   // yarım bırakılan araştırmaların kalan turu
    ActiveID    string           // şu an araştırılan teknoloji
    TurnsLeft   int              // aktif araştırmada kalan tur
}
```

---

## Görselleştirme

Teknoloji paneli (`internal/render/tech_panel.go`) ağaç yapısında gösterilir:

- **Seviyeler:** Gereksinimlere göre hiyerarşik seviyeler (0 = temel teknolojiler)
- **Renk Kodlaması:**
  - Askeri: Kırmızımsı (200,100,100)
  - Ekonomi: Yeşil (100,200,100) 
  - Diplomasi: Mavi (100,100,200)
  - Denizcilik: Sarı (200,200,100)
- **Tamamlanmış Teknolojiler:** Kategori rengine sahip tick badge ile işaretlenir
- **Aktif Araştırma:** HUD'da gösterilir (isim + kalan tur)
- **Seçim Esnekliği:** İlk seçim sonrası vazgeçme/değiştirme mümkün; yarım kalan araştırma pause olur ve sonra altın tekrar düşmeden kaldığı yerden devam eder
- **Tur Bitir Davranışı:** Aktif araştırma yoksa ve uygun bir teknoloji varsa turn resolution sırasında ağaç sırasına göre sonraki bağlı teknoloji otomatik başlatılır; böylece her tur panel açıp kart seçme zorunluluğu kalkar. Tüm teknolojiler tamamlandıysa veya yalnız kilitli düğümler kaldıysa otomatik başlangıç olmaz
  - Din: Magenta (200,100,200)
- **Durum Göstergeleri:**
  - Tamamlandı: Yeşil
  - Araştırılıyor: Sarı
  - Kilitli: Gri
  - Kullanılabilir: Kategori rengi
- **Bağlantılar:** Gereksinim teknolojileri diyagonal değil, ortogonal akış-şeması çizgileriyle gösterilir; mevcut veri modelindeki `Requires[]` bağımlılıkları zorunlu olduğu için çizgiler solid görünür
- **Okunabilirlik:** Düğüm içinde yalnız teknoloji adı ve açık maliyet/tur satırı gösterilir; kategori adı kutu içinde tekrar etmez, panel altındaki renk legend'inde gösterilir
- **Etkileşim:** Düğüm tıklayarak araştırma başlatma
- **AI Görünürlüğü:** Bölge panelindeki sahip devlet adına tıklanınca açılan devlet paneli, rakip devletin aktif araştırmasını, tamamlanan teknoloji listesini ve kümülatif buff özetini gösterir

1300 senaryosu artık başlangıç 26 düğümle sınırlı değildir; orta ve ileri dönem için yeni askeri, ekonomik, diplomatik, denizcilik ve dinî alt dallar eklendi. Özellikle `market_gold_mod`, `peace_relation_bonus`, `naval_move_bonus`, `reveal_enemy_strength` ve `conversion_speed_mod` effect alanları artık sadece veri içinde tanımlı kalmaz, runtime'da karşılık bulur.

AI zaten aynı `ResearchState` ve `tech.StartResearch / tech.Tick / tech.PauseResearch` akışını kullanır; oyuncu ve AI için teknoloji ilerleme mantığı ayrışmaz.

Birim üretim kapıları `assets/scenarios/<id>/data/units.json` içindeki `required_tech` dizisini AND olarak değerlendirir. Birim, listelenen zincirin tüm halkaları tamamlanmadan oyuncu veya AI tarafından üretilemez.

`applyTechTicks(gs)` — her tur `TurnsLeft--`, `0` olunca teknoloji tamamlanır.

---

## Teknoloji Efektleri

`TechEffects` — `tech.ComputeEffects(completed, types)`

| Efekt Alanı | Kullanıldığı Yer |
|---|---|
| `InfantryAttackMod` | Çarpışma hesabı |
| `CavalryAttackMod` | Çarpışma hesabı |
| `SiegeAttackMod` | Çarpışma hesabı |
| `LandDefenseMod` | Çarpışma hesabı |
| `GoldIncomeMod` | Ekonomi tick |
| `PopGrowthMod` | Bölge gelişimi |
| `GrainMod`, `IronMod`, `TimberMod`, `StoneMod` | Kaynak üretim çarpanları |

→ Çarpışmaya etkisi: [[systems/combat]]

---

## Kategoriler

| Kategori | İçerik |
|---|---|
| `military` | Saldırı/savunma bonusları, yeni birim tipleri |
| `economy` | Gelir artışı, bina verimliliği |
| `diplomacy` | İlişki bonusları, müzakere kolaylığı |
| `naval` | Deniz hareketi, gemi kapasitesi |
| `culture` | Din etkisi, memnuniyet, özel bölge bonusları |

Kategori görünen adları ve panel sırası `internal/tech/category_metadata.go` içinde merkezileştirilmiştir (`CategoryLabelTR`, `AllCategories`, `CategoryOrder`). Teknoloji paneli kategori başlığını ve düğüm sıralamasını bu ortak metadata'dan alır.

---

## Bölge Bağımlılığı

Bazı teknolojiler belirli şehirlerin ele geçirilmesini gerektirir:
- Konstantinopolis → Doğu Roma mühendisliği dalı
- Kudüs → Haçlı/Cihad teknolojileri (planlanmış)

---

## UI

`T` tuşu → teknoloji paneli aç/kapat (`internal/render/tech_panel.go`)

Panel: araştırılabilir teknolojileri listeler, cursor ile seçim, Enter ile araştırma başlatır. Tur sonunda aktif araştırma boşsa uygun sonraki bağlı teknoloji otomatik seçilip başlatılır.
