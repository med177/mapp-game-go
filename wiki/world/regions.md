---
type: world
tags: [regions, terrain, map, neighbors, coastal]
last_updated: 2026-05-08
related: [systems/combat, world/factions, architecture/render-pipeline]
---

# Bölge Sistemi

**Kaynak:** `internal/world/region.go`, `internal/world/terrain.go`, `assets/scenarios/<id>/data/regions.json`

## Region Yapısı

```go
type Region struct {
    ID        RegionID
    NameTR    string
    OwnerID   string           // fraksiyon ID veya ""
    Terrain   TerrainType
    Neighbors []RegionID       // komşu bölge listesi

    IsSea     bool             // deniz bölgesi
    IsLocked  bool             // henüz keşfedilmemiş
    WorldX, WorldY int         // harita koordinatı
    ShapeID string             // Natural Earth kaynak ID'si

    Buildings    []string      // inşa edilmiş bina ID'leri
    TaxRate      int           // 0-100
    Satisfaction int           // halk memnuniyeti
    Population   int

    Religion        string     // mevcut bölge dini
    ConversionTurns int        // din dönüşüm sayacı
    ActiveEventID   string
}
```

---

## Arazi Tipleri

`internal/world/terrain.go`

| Tip | Geçiş | Savunma Bonusu | Görüş |
|---|---|---|---|
| `TerrainPlain` (Ova) | Serbest | ×1.0 | Tam |
| `TerrainForest` (Orman) | Yavaş | ×1.3 | Kısıtlı |
| `TerrainMountain` (Dağ) | Geçilemez (geçit hariç) | ×1.8 | Yok |
| `TerrainPass` (Geçit) | Tek yol | ×1.5 | Kısıtlı |
| `TerrainCoast` (Kıyı) | Normal | ×1.1 | Normal |
| `TerrainSea` (Deniz) | Sadece deniz ordusu | — | — |

→ Savunma bonusu çarpışmaya etkisi: [[systems/combat]]

---

## Hareket Kuralları

`CanLandEnter()` — kara orduları deniz bölgesine giremez
`CanNavalEnter()` — deniz orduları sadece deniz bölgelerine girer
`IsCoastal()` — komşu bölgeler arasında deniz varsa `true` → gemi inşa koşulu

---

## Ele Geçirme

`ApplyConquest(ownerID, religion)` — savaş sonrası sahiplik transferi

1. `OwnerID = ownerID` → sahip değişir
2. Memnuniyet -10 düşer
3. Saldıranın dini bölgeden farklıysa ekstra -15 memnuniyet cezası uygulanır
4. Din dönüşümü tur çözümlemede `ConversionTurns` ile ilerler; 24 tur sonunda bölge dini yeni sahibin dinine döner

---

## Komşuluk Grafı

`Neighbors []RegionID` — hem kara hem deniz komşuları içerir.

Ordu hareketi bu listeyle kısıtlanır: sadece direkt komşuya hareket.

---

## Kilit Sistemi

`IsLocked = true` olan bölgeler haritada görünmez/girilemez. `checkRegionUnlocks()` belirli koşullarda (bölge yakınlaşması, teknoloji, tarih) `IsLocked = false` yapar.

---

## Kritik Bölgeler

Zafer koşulları ve olaylar için referans alınan bölgeler:

| Bölge ID | Önem |
|---|---|
| `constantinople` | Domination + Bizans teknoloji dalı |
| `rome` | Domination + Dini zafer |
| `jerusalem` | Domination + Dini zafer |
| `cairo` | Domination |
| `paris` | Domination (Fransa başkenti) |
| `london` | Domination (İngiltere başkenti) |
| `mecca` | Dini zafer |

Not: Senaryo zafer hedefleri şu an bazı yerlerde `CON`, `ROM`, `JER` gibi kısa ID'ler kullanıyor. Bunlar `regions.json` ID'leriyle eşleşmeli; takip işi [[dev/progress]] altında.

---

## 1300'lü Yıllar Tarihi Bölgeler

### İngiltere Krallığı (6 bölge)
- `london` — Başkent, yüksek gelir (60)
- `yorkshire` — Kuzey, tahıl üretimi (50)
- `lancashire` — Kuzeybatı, dağlık (30)
- `mercia` — Orta, ormanlık (45)
- `east_anglia` — Doğu, tahıl ambarı (40)
- `wessex` — Güneybatı, verimli ovalar (35)

### Fransa Krallığı (8 bölge)
- `paris` — Başkent, Île-de-France (70)
- `normandy` — Normandiya Dükalığı, kıyı (45)
- `brittany` — Bretonya, yarımada (35)
- `anjou` — Anjou Kontluğu, Loire vadisi (40)
- `champagne` — Şampanya, ticaret merkezi (50)
- `burgundy` — Burgonya Dükalığı (55)
- `provence` — Provence, Akdeniz kıyısı (50)
- `languedoc` — Languedoc, Toulouse (45)

### Kutsal Roma İmparatorluğu (6 prenslik)
- `brandenburg` — Brandenburg Markgrafluğu, kuzeydoğu
- `saxony` — Saksonya Dükalığı, kuzey orta
- `bavaria` — Bavyera Dükalığı, güney
- `westphalia` — Vestfalya, batı (Ren bölgesi)
- `thuringia` — Turingiya, orta
- `palatinate` — Palatinate, orta-batı
