---
type: dev
tags: [data, json, schema, assets]
last_updated: 2026-05-11
related: [architecture/state-management, world/regions, world/factions]
---

# JSON Veri Formatları

Tüm oyun tanım verisi her senaryo için `assets/scenarios/<senaryo_id>/data/` altında JSON olarak tutulur. Kod bu dosyaları `scenario.DataPath()` üzerinden okur — hiçbir tanım hardcode edilmez.

## Senaryo Yapısı

`assets/scenarios/scenarios.json` — yükleme sırası listesi:
```json
["1300_ottoman_rise", "1444_constantinople"]
```

`assets/scenarios/<id>/scenario.json` — senaryo meta verisi:
```json
{
  "id": "1300_ottoman_rise",
  "name": "Osmanlı'nın Yükselişi",
  "description": "...",
  "year": 1300,
  "month": 3,
  "map": {
    "world_width": 2892,
    "world_height": 1440,
    "shape_offset_x": -530,
    "shape_offset_y": -180,
    "shape_scale_x": 2.025,
    "shape_scale_y": 2.025
  },
  "music": {
    "default_playlist": "campaign",
    "playlists": {
      "campaign": [
        { "file": "ottoman_theme_01.ogg", "weight": 3 },
        { "file": "anatolia_ambient_01.mp3", "weight": 1 }
      ]
    }
  },
  "victory_conditions": [
    {
      "id": "ottoman_rise",
      "title": "Osmanlı'nın Yükselişi",
      "type": "conquer_city",
      "target": "constantinople"
    }
  ]
}
```

`type` değerleri: `domination`, `economic`, `military`, `religious`, `conquer_city`

`map` alanı opsiyoneldir. Verilmeyen alanlar renderer'ın geriye dönük uyumlu varsayılanlarıyla tamamlanır. `world_width` / `world_height` arka plan PNG dünya boyutunu, `shape_offset_*` ve `shape_scale_*` ise `country_shapes.json` koordinatlarının world pikseline dönüşümünü belirler.

`music` alanı opsiyoneldir. `default_playlist` senaryo yüklendikten sonra başlatılacak listeyi belirtir; `playlists` içindeki dosya adları senaryonun `musics/` klasörüne göre çözülür. Desteklenen formatlar: `.ogg`, `.mp3`, `.wav`. `weight` eksik veya `0` ise `1` kabul edilir. Paylaşılan tıklama/uyarı efektleri bu alanın parçası değildir; `assets/sounds/` altından yüklenir.

---

## Veri Dosyaları (`data/` klasörü)

Her senaryo kendi bağımsız veri setini taşır — aşağıdaki şemalar her senaryo için geçerlidir.

## regions.json

Bölge listesi. Her kayıt:

```json
{
  "id": "london",
  "name_tr": "Londra",
  "owner_id": "england",
  "terrain": "plain",
  "neighbors": ["wessex", "east_anglia", "_sea_north"],
  "is_sea": false,
  "is_locked": false,
  "world_x": 490,
  "world_y": 260,
  "settlements": [
    {
      "id": "london",
      "name_tr": "Londra",
      "x": 490,
      "y": 260,
      "type": "city",
      "is_capital": true
    },
    {
      "id": "westminster",
      "name_tr": "Westminster",
      "x": 486,
      "y": 262,
      "type": "town"
    }
  ],
  "tax_rate": 50,
  "religion": "catholic",
  "base_gold_income": 60,
  "base_grain_output": 35,
  "trade_capacity": 5
}
```

`world_x` / `world_y` bölgenin Voronoi/raster bölünmesindeki merkezidir. Haritada görünen şehir noktaları ve isimleri `settlements[]` üzerinden çizilir. Yerleşim `x` / `y` değerleri aynı senaryo koordinat uzayındadır; renderer bu koordinatı gerçek region piksel alanı dışında bulursa log uyarısı basar ve aynı region içindeki en yakın piksele fallback yapar. `settlements` eksikse eski davranış korunur ve bölge adı `world_x/world_y` noktasından çizilir.

Yerleşim `type` değerleri serbest metindir; mevcut kullanım: `city`, `town`, `port`, `fortress`. `is_capital: true` ana yerleşimi belirtir ve ordu/etiket anchor'ı için önceliklidir.

---

## factions.json

```json
{
  "id": "ottoman",
  "name_tr": "Osmanlı",
  "religion": "sunni",
  "color": [220, 80, 40],
  "gold": 200,
  "grain": 200,
  "iron": 100,
  "timber": 80,
  "spice": 50,
  "cloth": 60
}
```

Din değerleri `internal/religion` sabitleriyle eşleşir: `catholic`, `orthodox`, `sunni`, `shia`.

---

## relations.json

Başlangıç diplomasi ilişkileri. Dosya yoksa tüm faction çiftleri din temelli varsayılanlarla üretilir.

```json
[
  {
    "faction_a": "ottoman",
    "faction_b": "venice",
    "score": -20,
    "stance": "peace"
  },
  {
    "faction_a": "venice",
    "faction_b": "byzantine",
    "score": 35,
    "stance": "trade"
  }
]
```

`stance` değerleri: `war`, `peace`, `allied`, `trade`. `score` aralığı editörde `-100..100` olarak tutulur.

---

## units.json

```json
{
  "id": "militia",
  "name_tr": "Milis",
  "gold_cost": 60,
  "grain_upkeep": 2,
  "turns_required": 1,
  "attack": 10,
  "defense": 8,
  "hp": 100,
  "category": "infantry",
  "tier": 1,
  "required_tech": "",
  "required_bldg": "barracks",
  "embarkable": true
}
```

`turns_required` üretim kuyruğunda kaç tur sonra birimin ordu/filoya ekleneceğini belirler. Eksik bırakılırsa yükleyici geriye dönük uyumluluk için `1` kabul eder.

---

## buildings.json

```json
{
  "id": "market",
  "name_tr": "Pazar",
  "gold_cost": 120,
  "turns_required": 2,
  "max_per_region": 1,
  "required_terrain": "",
  "effects": { "gold_income": 30 }
}
```

`turns_required` bina inşaatının kaç tur süreceğini belirler. Eksik bırakılırsa yükleyici geriye dönük uyumluluk için `2` kabul eder.

---

## technologies.json

```json
{
  "id": "improved_infantry",
  "name_tr": "Gelişmiş Piyade",
  "category": "military",
  "turns_required": 4,
  "gold_cost": 80,
  "required_techs": [],
  "required_building": "barracks",
  "effects": { "infantry_attack_mod": 0.10 }
}
```

---

## events.json

```json
{
  "id": "black_death_1347",
  "name_tr": "Kara Veba",
  "description_tr": "Veba Akdeniz'i kasıp kavurdu.",
  "trigger": {
    "year_min": 1347,
    "region_ids": ["constantinople", "venice"],
    "faction_id": ""
  },
  "effects": {
    "population_mod": -0.3,
    "production_mod": -0.4,
    "spread_chance": 0.2
  }
}
```

---

## armies.json

Başlangıç orduları senaryo verisidir:

```json
{
  "id": "army_ottoman_1",
  "owner_id": "ottoman",
  "region_id": "bithynia",
  "is_naval": false,
  "units": [
    { "type_id": "militia", "count": 5 },
    { "type_id": "light_cavalry", "count": 2 }
  ]
}
```

`is_naval` opsiyoneldir; eksikse `false` kabul edilir. Donanmalar `is_naval: true` ile deniz region'larında tutulur. `internal/army/loader.go` `count` değerlerini `army.Unit` listesine açar.

---

## country_shapes.json

`tools/populate_all_shapes.py` tarafından Natural Earth `ne_10m_admin_0_countries` şekillerinden üretilir. Manuel düzenleme **yapma** — araçla yeniden üret.

Format: bölge ID → poligon nokta dizisi `[[x, y], ...]` (dünya koordinatları).

> **Not:** Eski `assets/data/generated/country_shapes.json` yolu artık kullanılmıyor. Her senaryo kendi `data/country_shapes.json` dosyasına sahip.
