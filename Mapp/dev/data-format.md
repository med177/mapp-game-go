---
type: dev
tags: [data, json, schema, assets]
last_updated: 2026-05-07
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
  "victory_conditions": [
    {
      "id": "ottoman_rise",
      "title": "Osmanlı'nın Yükselişi",
      "type": "conquer_city",
      "target": "CON"
    }
  ]
}
```

`type` değerleri: `domination`, `economic`, `military`, `religious`, `conquer_city`

---

## Veri Dosyaları (`data/` klasörü)

Her senaryo kendi bağımsız veri setini taşır — aşağıdaki şemalar her senaryo için geçerlidir.

## regions.json

Bölge listesi. Her kayıt:

```json
{
  "id": "anatolia",
  "name_tr": "Anadolu",
  "owner_id": "ottoman",
  "terrain": "plain",
  "neighbors": ["constantinople", "armenia", "black_sea"],
  "is_sea": false,
  "is_locked": false,
  "world_x": 412.5,
  "world_y": 218.3,
  "tax_rate": 50,
  "religion": "sunni_islam"
}
```

---

## factions.json

```json
{
  "id": "ottoman",
  "name_tr": "Osmanlı",
  "religion": "SunniIslam",
  "color": [220, 80, 40],
  "gold": 200,
  "grain": 50
}
```

---

## units.json

```json
{
  "id": "militia",
  "name_tr": "Milis",
  "attack": 10,
  "defense": 8,
  "hp": 100,
  "move_cost": 1,
  "category": "infantry",
  "tier": 1
}
```

---

## buildings.json

```json
{
  "id": "market",
  "name_tr": "Pazar",
  "gold_cost": 120,
  "max_per_region": 1,
  "required_terrain": "",
  "effects": { "gold_income": 30 }
}
```

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

## country_shapes.json

`tools/populate_all_shapes.py` tarafından Natural Earth `ne_10m_admin_0_countries` şekillerinden üretilir. Manuel düzenleme **yapma** — araçla yeniden üret.

Format: bölge ID → poligon nokta dizisi `[[x, y], ...]` (dünya koordinatları).

> **Not:** Eski `assets/data/generated/country_shapes.json` yolu artık kullanılmıyor. Her senaryo kendi `data/country_shapes.json` dosyasına sahip.
