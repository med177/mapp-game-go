---
type: dev
tags: [data, json, schema, assets]
last_updated: 2026-07-27
related: [architecture/state-management, architecture/shape-editor, world/regions, world/factions]
---

# JSON Veri Formatları

Tüm oyun tanım verisi her senaryo için `assets/scenarios/<senaryo_id>/data/` altında JSON olarak tutulur. Kod bu dosyaları `scenario.DataPath()` üzerinden okur — hiçbir tanım hardcode edilmez.

## Senaryo Yapısı

`assets/scenarios/scenarios.json` — yükleme sırası listesi:
```json
["1300_ottoman_rise", "1444_ottoman_empire", "1648_westphalia_peace", "1800_napoleon_rise"]
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
      "required_regions": ["constantinople"],
      "deadline_year": 1561,
      "deadline_month": 1
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

## imperial.json

HRE gibi bağımsız üyelerden oluşan üst siyasi kurumları tanımlar:

```json
{
  "empire_id": "hre",
  "emperor_id": "hre",
  "authority": 62,
  "next_diet_turn": 12,
  "election_due_turn": 97,
  "members": [
    {
      "faction_id": "milan_duchy",
      "status": "prince",
      "loyalty": 54,
      "autonomy": 88,
      "military_commitment": 48,
      "elector_weight": 1
    }
  ]
}
```

`status`: `elector`, `prince`, `free_city`, `order` veya `vassal` olabilir. `vassal`
üyeler yalnız bilgilendirme içindir; otomatik savaş/erişim davranışını `OverlordID`
belirler. `authority`, `loyalty`, `autonomy` ve `military_commitment` 0–100 aralığında
normalize edilir. `imperial.json` yoksa senaryo imparatorluk sistemi olmadan yüklenir.

Oyuncu HRE olarak oynarken Diyet veya seçim kararı bekliyorsa runtime save içinde
`imperial.pending_decision` alanı oluşur:

```json
{
  "pending_decision": {
    "kind": "diet",
    "created_turn": 12
  }
}
```

`kind` değeri `diet` veya `election` olabilir. Bu alan senaryo başlangıç verisinde
zorunlu değildir; karar tur çözümlemesinde oluşturulur ve karar verildiğinde silinir.

## trade_centers.json

Ticaret harita modunda kullanılacak tarihsel merkez düğümleri ve aralarındaki koridor graph'ı.

```json
{
  "centers": [
    { "id": "venice", "tier": "primary", "links": ["genoa", "constantinople"] },
    { "id": "constantinople", "tier": "primary", "links": ["venice", "aleppo"] },
    { "id": "aleppo", "tier": "secondary", "links": ["constantinople", "basra"] }
  ]
}
```

Alanlar:
- `id`: `regions.json` içindeki bölge ID'si
- `tier`: `primary` veya `secondary` (görsel vurgu seviyesi)
- `links`: merkezin doğrudan bağlı olduğu diğer merkez ID'leri

Kurallar:
- Sıra önemlidir; renderer merkezleri dosyadaki sırayla alır.
- Geçersiz, deniz veya `trade_capacity <= 0` olan merkezler atlanır.
- `links` içinde geçersiz/tekrarlı/self link girdileri temizlenir.
- Koridor akışı doğrudan her merkez çifti arasında değil, bu link graph'ı üzerindeki kısa yol boyunca dağıtılır.
- Dosya yoksa merkez listesi boş kalır (trade map çizimi yapılmaz).

## land_passages.json

Deniz veya başka bir görsel boşluk olsa bile iki kara bölgesini haritada özel
karasal geçiş olarak göstermek için kullanılır. Dosya doğrudan bir listedir:

```json
[
  {
    "from": "sicily",
    "to": "naples",
    "type": "strait",
    "move_cost": 1,
    "defense_bonus": 15,
    "start": [784, 509],
    "end": [770, 480]
  }
]
```

`from` ve `to` mevcut, deniz olmayan region ID'leri olmalıdır. Loader aynı iki
region arasındaki ters yönlü yinelenen kaydı tek kayda indirir. `type` şu anda
`strait` ile sınırlıdır; `move_cost` ve `defense_bonus` runtime state'te
korunur, hareket/savaş hesabına bağlanması sonraki fazdır. `start` ve `end`,
senaryo koordinatlarında çizginin tam `[x,y]` uçlarıdır; verilmezse eski kayıtlar
bölge/yerleşim anchor'ına geri döner. Edit mode `Shape` sekmesindeki `Geçiş Ekle`
butonu veya `P` ile ekleme modu açılır; önce başlangıç kara noktasına, sonra
bitiş kara noktasına tıklamak `strait / 1 / 15` ve iki koordinatlı kaydı oluşturur.
`Geçiş Düzenle` mevcut çizgiyi seçip uç noktalarını sürükleyerek `start/end`
değerlerini değiştirir. Seçili geçiş `Geçiş Sil` butonuyla veya `Delete` tuşuyla
silinir. Aynı `Shape` sekmesindeki `Komşu Ekle`, seçili kara bölgeyi kaynak
kabul edip haritadan seçilen hedef kara bölgeye iki yönlü `neighbors` bağlantısı
ekler; bu bağlantı deniz aralığı olsa bile kara ordusu hareket grafiğinde
kullanılır. `Ctrl+S` bu değişiklikleri `regions.json` ile birlikte yazar.

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

Yerleşim `type` değerleri serbest metindir; mevcut kullanım: `city`, `town`, `port`, `fortress`. `is_capital: true` ana yerleşimi belirtir ve ordu/etiket anchor'ı için önceliklidir. Her yerleşim `population` alanıyla kendi nüfusunu taşır; bölgenin `population` değeri yerleşim nüfusları ile `rural_population` alanının toplamıdır. `rural_population`, köyler ve yerleşim dışı kırsal nüfusu temsil eder.

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
  "cloth": 60,
  "research": {
    "completed": {
      "iron_weapons": true,
      "horse_breeding": true
    },
    "active_id": "",
    "turns_left": 0
  },
  "ai_aggressiveness": 62,
  "ai_expansion_targets": ["east_rome", "germiyan_bey"]
}
```

Din değerleri `internal/religion` sabitleriyle eşleşir: `catholic`, `orthodox`, `sunni`, `shia`.

`ai_expansion_targets` opsiyoneldir ve fraksiyon ID listesi taşır. Normal/zor zorlukta AI bu hedefleri fırsatçı savaş değerlendirmesinde önceliklendirir; hedefin yine kara sınırı paylaşması, ilişkinin `peace` olması ve güç kıyasından geçmesi gerekir.

`research.completed` teknoloji ID'lerini anahtar, tamamlanma durumunu `true` değer
olarak taşıyan bir nesnedir. Dizi biçimi desteklenmez; bu sözleşme yükleme sırasında
`faction.ResearchState.Completed map[string]bool` alanına doğrudan eşlenir.

---

## ai_strategies.json

Senaryoya özgü uzun vadeli AI yönlerini taşıyan opsiyonel dosyadır. Dosya yoksa
fraksiyonlar `factions.json.ai_expansion_targets` ve genel AI fallback'ini kullanır.
İlk kullanım yalnız `1300_ottoman_rise` senaryosundadır.

```json
{
  "difficulty_policy": {
    "fair_movement": true,
    "levels": {
      "1": {
        "plan_horizon_turns": 4,
        "plan_target_region_limit": 3,
        "path_search_depth": 5,
        "plan_move_bonus_percent": 70,
        "war_threshold": 82,
        "min_attack_power_percent": 130,
        "war_cadence_turns": 12,
        "max_concurrent_wars": 1
      },
      "2": {
        "plan_horizon_turns": 6,
        "plan_target_region_limit": 4,
        "path_search_depth": 8,
        "plan_move_bonus_percent": 100,
        "proactive_war": true,
        "war_threshold": 70,
        "min_attack_power_percent": 115,
        "war_cadence_turns": 10,
        "max_concurrent_wars": 1
      },
      "3": {
        "plan_horizon_turns": 9,
        "plan_target_region_limit": 5,
        "path_search_depth": 12,
        "plan_move_bonus_percent": 125,
        "proactive_war": true,
        "war_threshold": 65,
        "min_attack_power_percent": 100,
        "war_cadence_turns": 7,
        "max_concurrent_wars": 2,
        "player_target_score_bonus": 4,
        "start_gold_buffer": 80,
        "start_grain_buffer": 30
      }
    }
  },
  "factions": [
    {
      "faction_id": "ottoman",
      "profile": "frontier_expansion",
      "objectives": [
        {
          "id": "unite_anatolian_beyliks",
          "kind": "expand",
          "target_factions": ["karesioglu_bey", "germiyan_bey", "ahiler"],
          "target_regions": ["aydinoglu", "germiyan", "kutahya", "sivrihisar"],
          "priority": 92,
          "commitment": 66,
          "allow_vassalization": true,
          "annex_region_ids": ["kutahya"]
        }
      ]
    }
  ]
}
```

Alanlar:

- `difficulty_policy.fair_movement`: `true` ise AI ve oyuncu aynı hareket havuzunu
  kullanır; zor AI'ye ayrı `+1` hareket verilmez.
- `difficulty_policy.levels`: `1`, `2`, `3` için eksiksiz zorluk tanımlarıdır.
- `plan_horizon_turns`: kalıcı planın varsayılan yeniden değerlendirme aralığı.
- `plan_target_region_limit`: bir objective'ten plana alınabilecek azami hedef bölge.
- `path_search_depth`: uzun menzilli hareket aramasının azami derinliği.
- `plan_move_bonus_percent`: objective kaynaklı hareket puanının seviye çarpanı.
- `proactive_war`: barışta fırsat savaşı değerlendirmesini açar.
- `war_threshold`: savaş adayının geçmesi gereken asgari toplam skor.
- `min_attack_power_percent`: saldıranın hedef gücüne göre asgari güç yüzdesi.
- `war_cadence_turns`: proaktif savaş taramasının taban tur aralığı.
- `max_concurrent_wars`: devletin aynı anda taşıyabileceği azami savaş/cephe sayısı.
- `player_target_score_bonus`: oyuncu hedef olduğunda eklenen zorluk skoru.
- `start_gold_buffer` / `start_grain_buffer`: yeni oyunda AI'ye verilen küçük başlangıç
  tamponı; 1300 senaryosunda yalnız Zor seviyede `80/30` değerindedir.
- `faction_id`: `factions.json` içindeki devlet ID'si; dosyada tekil olmalıdır.
- `profile`: telemetri ve sonraki profile bağlı ağırlıklar için strateji etiketi.
- `objectives[].id`: devlet içinde tekil kalıcı objective kimliği.
- `kind`: `expand`, `defend`, `consolidate` veya diplomatik yönelim metadata'sı olan `ally`.
  `ally` objective'leri askeri stratejik plan üretmez; saldırı hedefi gibi yorumlanmaz.
- `target_factions` / `target_regions`: soft yönelim verilecek hedefler.
- `priority`: aynı anda uygulanabilir objective'ler arasındaki taban öncelik.
- `commitment`: plan save state'ine taşınan `25..90` kararlılık değeri.
- `readiness_regions`: sahip olunan her kayıt için hazırlık bonusu, eksikler için soft
  ceza üretir; objective'i tek başına kapatmaz.
- `min_year`: geç/anakronik hedefi verilen yıla kadar kapatan hard gate.
- `required_event_flags`: event motorunun `set_flags` etkileriyle açılan hard gate listesi.
- `allow_vassalization`: son toprağında yenilen uygun hedefin vassal bırakılmasına izin verir.
- `annex_region_ids`: vassallık açık olsa da stratejik sayılıp doğrudan ilhak edilecek bölgeler.

Statik config save'e yazılmaz. Senaryo başlangıcında ve save baz state'i kurulurken
`scenario.LoadAIConfig()` ile yüklenir; seçilmiş dinamik objective ise
`GameState.AIPlans` içinde serialize edilir. Dinamik planın güvenli dost sınır hazırlığı
başlatılmışsa `rally_region_id` ve `rally_deadline_turn` da aynı kayıtta korunur; rol,
toplanan güç ve yol cache'leri save'e yazılmaz.

Aktif savaşların dinamik sonucu compact campaign state'te `wl`, legacy/debug sidecar'da
`war_ledgers` alanında tutulur. Her sıralı faction çifti `faction_a`, `faction_b`,
`started_turn`, `initial_regions_a/b`, `casualties_a/b`, `regions_captured_a/b`,
`last_battle_turn` ve `last_peace_offer_turn` alanlarını taşır. Bu kayıt yalnız aktif
savaş içindir; barışta silinir. Alanı taşımayan eski save yüklenirse aktif `war`
ilişkileri mevcut yükleme turundan sıfır sayaçla başlatılır.

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
    "faction_b": "east_rome",
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
  "required_tech": [],
  "required_bldg": "barracks",
  "embarkable": true
}
```

`turns_required` üretim kuyruğunda kaç tur sonra birimin ordu/filoya ekleneceğini belirler. Eksik bırakılırsa yükleyici geriye dönük uyumluluk için `1` kabul eder.

`required_tech` dizisi AND semantiğine sahiptir; listedeki tüm teknoloji ID'leri tamamlanmadan birim üretilemez. Teknoloji zincirinin ara adımları da açıkça yazılır (ör. top için `gunpowder` ve `cast_bronze_cannon`).

`carry_capacity` sadece `category = "naval_trans"` birimlerinde kullanılır. Her nakliye gemisinin aynı anda taşıyabildiği kara birimi slot sayısını belirtir. Filo toplam kapasitesi, filodaki tüm nakliye gemilerinin `carry_capacity` toplamıdır; ancak mevcut oyun kuralı gereği toplam taşınan kara birimi sayısı yine `MaxArmySize` sınırını aşmaz.

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
`required_terrain` genel binalar için literal arazi filtresidir; ancak `port` için kanonik kural `internal/world/region.go:122` içindeki `Region.IsCoastal`, yani denize komşu kara bölgesi olmaktır.

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

`is_naval` opsiyoneldir; eksikse `false` kabul edilir. Donanmaların `region_id` alanı deniz rotası ankrajıdır. Limanda bekleyen donanmalarda `docked_region_id` ve gerçek liman `docked_settlement_id` bulunur; `Army.LocationID()` limanda settlement ID'sini, denizde `region_id` değerini kanonik konum olarak döndürür. `internal/army/loader.go` `count` değerlerini `army.Unit` listesine açar.

---

## country_shapes.json

`tools/populate_all_shapes.py` tarafından Natural Earth `ne_10m_admin_0_countries` şekillerinden üretilir. Büyük toplu üretimler hâlâ araç tarafında yapılır; küçük kıyı/sınır düzeltmeleri edit mode `Shape` sekmesinden oyun içi paint editor ile yapılabilir.

Format: `{"shapes": [{"id": string, "name": string, "rings": [[[x, y], ...]]}]}`. `rings` içindeki koordinatlar shape/scenario uzayındadır; aktif senaryonun `map.shape_offset_*` ve `map.shape_scale_*` alanlarıyla world pikseline dönüştürülür.

> **Not:** Eski `assets/data/generated/country_shapes.json` yolu artık kullanılmıyor. Her senaryo kendi `data/country_shapes.json` dosyasına sahip.
