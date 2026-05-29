---
type: system
tags: [technology, research, effects, tree]
last_updated: 2026-05-29
related: [systems/combat, systems/economy, architecture/state-management]
---

# Teknoloji Ağacı

**Kaynak:** `internal/tech/tech.go`, `assets/data/technologies.json`

## Araştırma Yapısı

```go
type Technology struct {
    ID           string
    NameTR       string
    TurnsRequired int
    GoldCost     int
    Category     string        // military | economy | diplomacy | naval | culture
    RequiredTechs []string     // bağımlılıklar
    RequiredBuilding string    // gerekli bina ID
    Effects      TechEffects
}
```

---

## Araştırma Durumu (Faction içinde)

```go
type ResearchState struct {
    ActiveID   string           // şu an araştırılan teknoloji
    Progress   int              // geçen tur sayısı
    Completed  map[string]bool  // tamamlanan teknoloji ID'leri
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
- **Seçim Esnekliği:** İlk seçim sonrası vazgeçme/değiştirme mümkün
- **Tur Bitir Uyarısı:** Aktif araştırma yoksa panel açılır ve uyarı verilir
  - Din: Magenta (200,100,200)
- **Durum Göstergeleri:**
  - Tamamlandı: Yeşil
  - Araştırılıyor: Sarı
  - Kilitli: Gri
  - Kullanılabilir: Kategori rengi
- **Bağlantılar:** Gereksinim teknolojileri arasındaki çizgiler
- **Etkileşim:** Düğüm tıklayarak araştırma başlatma

`applyTechTicks(gs)` — her tur `Progress++`, `TurnsRequired`'e ulaşınca tamamlanır.

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

---

## Bölge Bağımlılığı

Bazı teknolojiler belirli şehirlerin ele geçirilmesini gerektirir:
- Konstantinopolis → Bizans mühendisliği dalı
- Kudüs → Haçlı/Cihad teknolojileri (planlanmış)

---

## UI

`T` tuşu → teknoloji paneli aç/kapat (`internal/render/tech_panel.go`)

Panel: araştırılabilir teknolojileri listeler, cursor ile seçim, Enter ile araştırma başlatır.
