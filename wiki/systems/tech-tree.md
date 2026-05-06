---
type: system
tags: [technology, research, effects, tree]
last_updated: 2026-05-06
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

## Araştırma Akışı

`tech.StartResearch(research, tech, gold)` — `internal/tech/tech.go`

Koşullar:
- Zaten başka araştırma aktif değil
- Gerekli altın mevcut
- Bağımlı teknolojiler tamamlanmış
- Gerekli bina bölgede var

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
