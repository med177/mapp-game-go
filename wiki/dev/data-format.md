---
type: dev
tags: [data, json, schema, assets]
last_updated: 2026-10-14
related: [architecture/state-management, architecture/shape-editor, world/regions, world/factions, architecture/render-pipeline]
---

# JSON Veri Formatları

Tüm oyun tanım verisi her senaryo için `assets/scenarios/<senaryo_id>/data/` altında JSON olarak tutulur. Kod bu dosyaları `scenario.DataPath()` üzerinden okur — hiçbir tanım hardcode edilmez.

## Senaryo Yapısı

`assets/scenarios/scenarios.json` — yükleme sırası listesi:
```json
["1300_ottoman_rise", "1444_ottoman_empire", "1648_westphalia_peace", "1800_napoleon_rise"]
```

`assets/scenarios/<id>/data/scenario.json` — senaryo meta verisi:
```json
{
  "id": "1300_ottoman_rise",
  "name": "Osmanlı'nın Yükselişi",
  "description": "...",
  "year": 1300,
  "month": 3,
  "map": {
    "shape_set_id": "akdeniz_havzasi",
    "background_image": "world_map_background.png",
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

`victory_conditions` içindeki `allowed_factions` eksik veya boşsa hedef tüm
fraksiyonlara açıktır; doluysa yalnızca listelenen fraksiyonlara gösterilir.

`map` alanı opsiyoneldir. Verilmeyen alanlar renderer'ın geriye dönük uyumlu varsayılanlarıyla tamamlanır. `background_image`, senaryonun `maps/` klasöründeki arka plan dosyasının yalnızca dosya adını belirtir; `.png`, `.jpg` ve `.jpeg` desteklenir. Alan verilmezse `world_map_background.png` kullanılır. `world_width` / `world_height` arka plan resminin dünya boyutunu, `shape_offset_*` ve `shape_scale_*` ise `country_shapes.json` koordinatlarının world pikseline dönüşümünü belirler.

`music` alanı opsiyoneldir. `default_playlist` senaryo yüklendikten sonra başlatılacak listeyi belirtir; `playlists` içindeki dosya adları senaryonun `musics/` klasörüne göre çözülür. Desteklenen formatlar: `.ogg`, `.mp3`, `.wav`. `weight` eksik veya `0` ise `1` kabul edilir. Paylaşılan tıklama/uyarı efektleri bu alanın parçası değildir; `assets/sounds/` altından yüklenir.

`diplomacy.peace_periods` belirli fraksiyon çiftleri arasında savaş ilanını kısıtlar. `blocked_factions` içindeki iki fraksiyon arasındaki savaşlar etkilenir; daha fazla fraksiyon eklemek listedeki tüm çiftleri kapsar. Yalnızca tek bir çifti sınırlamak için iki kimlikli ayrı bir kayıt kullanın. `min_turns` varsa bu asgari tur eşiği geçilmelidir. `war_declaration_requires_event`, gerçekleşmesi beklenen event'in **ID** değeridir; `FiredEventIDs[event_id]` kaydı kalıcıdır ve event'in `clear_flags` etkilerinden etkilenmez. `war_declaration_requires_event_flag` ve `war_declaration_unblocked_by_event_flag` ise değişebilir `flag:<id>` durumlarını kontrol eder; flag'ler sonradan temizlenebilir. Bu nedenle tek seferlik tarihsel açılışlar için event ID alanı tercih edilmelidir. Tüm alanlar opsiyoneldir. Bu kurallar mevcut savaşları bitirmez; başlangıçta savaşta olan çiftler için ilişkiler dosyasındaki `stance` değeri de `peace` olmalıdır.

```json
{
  "diplomacy": {
    "peace_periods": [
      {
        "min_turns": 10,
        "war_declaration_requires_event": "major_war_started_event",
        "blocked_factions": ["faction_a", "faction_b"]
      }
    ]
  }
}
```

---

## Veri Dosyaları (`data/` klasörü)

Her senaryo kendi bağımsız veri setini taşır — aşağıdaki şemalar her senaryo için geçerlidir.

## characters.json

Karakter kayıtları `type` alanıyla sınıflandırılır. Desteklenen değerler
`administrator`, `commander` ve `character`'dır. `administrator` ve `commander`
tipleri kara ordusu veya savaş gemisi içeren filolara atanabilir; `character`
tipi askeri komutan roster'ında gösterilmez ve atanamaz. Eski senaryolarda
`type` eksikse kayıt geriye dönük uyumluluk için `commander` kabul edilir.
Portre dosya adları `sprites/characters/` altından çözülür.
`assigned_army_id` ve `assigned_council_role` kampanya/runtime atamalarıdır;
senaryo template'inde bulunsalar bile karakter yükleyicisi bu alanları temizler.
Oyuncu atamaları yalnız save'deki runtime karakter kaydında tutulur.

Komutan tipindeki karakter şablonlarında `start_year` zorunlu başlangıç yılını, `end_year` ise
doğal ölüm sınırını belirtir. Komutan `start_year` yılında aktif olur; `end_year`
başladığında ölümü `GameState.DeadCommanderIDs` içine kaydedilir ve komutan
runtime roster'ından çıkarılır. `end_year` eksikse tarih kaynaklı doğal ölüm yoktur.
Event veya savaş sırasında daha erken gerçekleşen ölümler de aynı ölüm state'ini
kullanır. Yeni senaryo verilerinde `unavailable_statuses` kullanmayın: tutsaklık
`CaptiveAtSettlementID`, sürgün ise `ExiledCommanderIDs` state'inde tutulur. Alan
eski senaryo kayıtları için geriye dönük uyumluluk amacıyla okunmaya devam eder.
`administrator` ve `character` kayıtlarında isteğe bağlı `adoption_cost_gold`,
tarafsız havuzdan haneye alımın tek seferlik altın ücretini; `gold_upkeep` ise
hane üyesi başına turda ödenecek altın bakımını belirtir. `commander` kayıtlarında
`adoption_cost_gold` şablonun haneye alım ücretini, `gold_upkeep` ise karakter
konsey görevine atandığında tur başına ödenecek bakımı belirtir; görevsiz askerî
komutan için bu karakter bakımı alınmaz. Senaryo karakter verilerinde bakım
tutarları seviye başına 2 altın kuralını izler. Eksik alanlar sıfır kabul edilir.
`character_council.json` içindeki `no_cost: true` bu bakımı görev atanmış
karakterler için kaldırır.
Ekonomi çözümlemesinde hane bakımı önce, ordu maaşı ve bina bütçesi daha sonra
tahsil edilir; ödenemeyen hane bakımı ordu maaşı açığı cezasına dahil edilmez.

```json
{
  "id": "neutral_character_example",
  "type": "character",
  "owner_id": "none",
  "name": "Örnek Karakter",
  "level": 2,
  "adoption_cost_gold": 500,
  "gold_upkeep": 4,
  "traits": ["aggressor"]
}
```

## character_council.json

Senaryoya özgü konsey görevlerini tanımlar.
Dosya isteğe bağlıdır; bulunmadığında eski senaryolar yerleşik görev tanımlarını
kullanır. `roles` içindeki her görev benzersiz bir `id` ve Türkçe `name_tr`
taşımalıdır. `effects` başlangıç bonusunu, `per_level` ise 1. seviye üzerindeki
her seviye için ek bonusu belirler. İsteğe bağlı `no_cost: true`, bu göreve
atanmış karakterin tur başı `gold_upkeep` bakımını kaldırır; tarafsız havuzdan
alımın tek seferlik `adoption_cost_gold` ücretini etkilemez.
İsteğe bağlı `set_commander_army: true`, bu görevi yürütürken karakterin aynı
zamanda bir orduya komuta etmesine izin verir. Alan yoksa varsayılan `false` olur;
alana izin veren bir göreve atanmış ordu komutanı için de görev ve ordu komutanlığı
birlikte sürer. Bu alan, atanmış komutanın ordu uygunluğunu ve göreve atanabilirken
mevcut ordu atamasını koruyabilmesini aynı anda belirler.

Desteklenen etki anahtarları:

- `income_mod`: fraksiyonun gelir kaynaklarına çarpan bonusu (`1.05` = +%5).
- `satisfaction_points`: fraksiyona ait bölgelerin memnuniyet tur değişimine puan bonusu.
- `military_power_mod`: kara ve deniz ordularının saldırı/savunma gücüne çarpan bonusu (`1.03` = +%3).
- `commander_power_mod`: Askeri Sorumlu'nun komuta ettiği ordu/donanmanın gücüne ek çarpan bonusu.
- `relation_bonus_points`: Heyet diplomasi eylemine ek ilişki puanı.
- `resource_production_mod`: fraksiyonun tahıl, demir, kereste, taş, baharat ve kumaş üretimine çarpan bonusu.
- `research_speed_mod`: yeni başlatılan araştırmanın hızına çarpan bonusu.

`*_mod` alanları `1.0` tabanlıdır: `1.0` etkisiz, `1.05` +%5 ve `0.95` -%5
anlamına gelir. Bonuslar birleştirilirken çarpanların `1.0` üzerindeki farkları
toplanır; böylece birden fazla görev/nitelik etkisi eskisi gibi yüzde puanı
olarak eklenir. `per_level` içindeki mod değeri, 1. seviyenin üzerindeki her
seviye başına eklenecek farkı belirtir (`1.01` seviye başına +%1).
Negatif değerli nitelik etkileri de desteklenir: örneğin Askeri Sorumlu rolünde
`military_power_mod: 0.98`, o karakterin görev bonusundan 2 yüzde puanı düşürür.
Puan alanlarında negatif tamsayılar aynı şekilde ilgili puan bonusunu azaltır.
Bir rol, isteğe bağlı `trait_fallback_effects` ile eşleşen olumlu niteliği
olmayan karaktere tek seferlik varsayılan nitelik cezası tanımlayabilir. Örneğin
Askeri Sorumlu rolündeki `{"military_power_mod": 0.99}`, hiçbir niteliği askerî
güç bonusu sağlamayan karakterin toplam bonusundan 1 yüzde puanı düşürür; bu
ceza karakterin kaç tane uyumsuz niteliği olduğuna göre katlanmaz.

`xp_per_turn`, göreve atanmış karakterin her tur kazandığı deneyimi;
`xp_per_effect`, görev bonusu fiilen uygulandığında (gelir, memnuniyet,
üretim, diplomasi, araştırma başlangıcı veya muharebe) kazandığı deneyimi
belirler. Konsey seviyesi komutan deneyim eşiklerini kullanır
(100/300/550/850 XP ile 2–5. seviyeler); XP ve seviye karakterin kampanya
kaydında saklanır.
Örnek:

```json
{
  "xp_per_turn": 5,
  "xp_per_effect": 10,
  "roles": [
    {
      "id": "treasurer",
      "name_tr": "Hazine Sorumlusu",
      "effects": {
        "income_mod": 1.05,
        "resource_production_mod": 1.05
      },
      "per_level": {
        "income_mod": 1.01,
        "resource_production_mod": 1.01
      }
    }
  ]
}
```

## character_traits.json

Senaryoya özgü karakter niteliklerinin Türkçe adlarını ve konsey bonuslarını
tanımlar. Dosya isteğe bağlıdır. `traits` anahtarının altında kayıtlar karakter
şablonlarındaki `traits` ID'leriyle eşleşir; `council_effects` altındaki rol
ID'leri aynı senaryonun `character_council.json` dosyasında tanımlı olmalıdır.
Bonuslar görev etkilerine eklenir.

```json
{
  "traits": {
    "scholar": {
      "name_tr": "Bilgin",
      "council_effects": {
        "treasurer": {
          "income_mod": 1.02,
          "resource_production_mod": 1.02
        }
      }
    }
  }
}
```

### regions.json ardıl devlet alanı

Kara bölgesinde isteğe bağlı `successor_faction_id` alanı, fetih sonrası yeniden
kurulabilecek fraksiyonun ID'sini taşır:

```json
{
  "id": "constantinople",
  "owner_id": "east_rome",
  "successor_faction_id": "east_rome"
}
```

Bu alan edit mode `Ardıl Devlet` düğmesiyle yazılır; compact save karşılığı `sf`dir.

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
  "sources": [
    { "region_id": "silk_road", "main_route": true, "links": [{ "region_id": "azerbaijan", "type": "land" }] }
  ],
  "centers": [
    { "region_id": "venice", "tier": "primary", "links": [{ "region_id": "genoa", "type": "sea" }, { "region_id": "constantinople", "type": "sea" }] },
    { "region_id": "constantinople", "tier": "primary", "links": [{ "region_id": "venice", "type": "sea" }, { "region_id": "aleppo", "type": "land" }] },
    { "region_id": "aleppo", "tier": "secondary", "links": [{ "region_id": "constantinople", "type": "land" }, { "region_id": "basra", "type": "land" }] }
  ]
}
```

Alanlar:
- `sources`: ana kaynak yollarının merkezlerden ayrı tutulduğu dizi; bu kayıtlar otomatik olarak off-map kabul edilir ve runtime yüklemesi bunları merkez grafiğine dahil eder
- `id`: `regions.json` içindeki bölge ID'si
- `tier`: `primary` veya `secondary` (görsel vurgu seviyesi)
- `main_route`: merkezin ana tarihsel yol olarak görsel ve akış kurallarına tabi olduğunu belirtir
- `links`: görsel olarak merkezden çıkan oku `{ "region_id": "...", "type": "land|sea" }` nesnesiyle tutar; normal merkezlerde tek kayıt ekonomik olarak iki yönlü bağlantı kabul edilir

Kurallar:
- Sıra önemlidir; renderer merkezleri dosyadaki sırayla alır.
- Geçersiz, deniz veya `trade_capacity <= 0` olan merkezler atlanır.
- `links` içinde geçersiz/tekrarlı/self link girdileri temizlenir.
- `sources` bağlantıları yalnız listelenen yönde mal taşır; hedef merkezden kaynağa ters akış oluşturulmaz.
- Koridor akışı doğrudan her merkez çifti arasında değil, bu yönlü link graph'ı üzerindeki kısa yol boyunca dağıtılır.
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

`from` ve `to` mevcut, deniz olmayan region ID'leri olmalıdır. `name` isteğe
bağlı harita etiketidir. Loader aynı iki region arasındaki ters yönlü yinelenen
kaydı tek kayda indirir. `type` `strait`, `bridge`, `mountain_pass` veya
`fortified_crossing` olabilir. `move_cost` geçişten giriş hareket maliyetidir;
`defense_bonus` geçiş üzerinden yapılan kara temasında savunmacıya yüzde bonus
verir. `start` ve `end`,
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

Yerleşim `type` değerleri serbest metindir; mevcut kullanım: `city`, `town`,
`port`, `fortress`. `is_center: true` ana yerleşimi belirtir ve ordu/etiket
anchor'ı için önceliklidir. Yerleşim kayıtları nüfus taşımaz; eski dosyalardaki
`population` alanı geriye dönük olarak okunmaz ve kaydetmede yeniden yazılmaz.
Bölge nüfusunun tek senaryo kaynağı `regions.json` içindeki `population`
alanıdır. `rural_population` yalnız eski save/oyun içi nüfus değişimleri için
yardımcı uyumluluk alanıdır.

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
  "ai_expansion_targets": ["east_rome", "germiyan_bey"],
  "territorial_claims": [
    { "region_id": "bithynia", "value": 100, "core": true },
    { "region_id": "constantinople", "value": 80 }
  ]
}
```

Din değerleri `internal/religion` sabitleriyle eşleşir: `catholic`, `orthodox`,
`sunni`, `shia`, `tengri`, `pagan`, `protestant`, `utraquist`.

`ai_expansion_targets` artık senaryo runtime uyumluluğu için türetilen alandır.
1300 senaryosunda kaynak liste `ai_strategies.json` içindeki `expansion_targets`
alanıdır; bu hedefler base state kurulurken faction runtime'ına kopyalanır.

`territorial_claims` opsiyoneldir. AI stratejisindeki her kayıt `region_id` ve
`value` (`1..100`) taşır. `core` bilgisi stratejiye yazılmaz; başlangıçta sahip
olunan kara bölgelerinden türetilir. Core bölgeleri, başkent tehdidi veya askerî
çöküş gibi acil durum yoksa AI'nin o bölge düşmandayken olağan barışı kabul
etmesini engeller. Normal claim değeri barış eşiğini yükseltir; aktif
`AIPlanState.TargetRegionIDs`, `WarLedger.TargetRegionID` ve tarihsel
`expansion_targets` claim kaynağı değildir; claim'ler bölge bazında strateji
`territorial_claims` ve objective hedeflerinden üretilir.

Senaryo base state'i yüklenirken sahibi olan her başlangıç kara bölgesi ilgili
devlet için otomatik olarak `core: true` ve `value: 100` claim'ine eklenir.
`ai_strategies.json` içindeki `territorial_claims` ve objective içindeki
`territorial_claims` alanları ilgili devletin claim listesinde birleştirilir.
Objective `territorial_claims` bölgeyi ve claim ağırlığını tek
yerde taşır; bölgenin güncel sahibi dinamik hedef devlettir. Bu üretim yeni oyun
ve save/load base state'inde aynıdır;
fetih sonrası mevcut sahiplik değişikliği core listesini geriye dönük değiştirmez.

`diplomacy_disabled`, `trade_disabled` ve `war_fatigue_exempt` opsiyonel boolean alanlardır.
`diplomacy_disabled`, faction'ın taraf olduğu tüm normal diplomasi aksiyonlarını
ve ticareti, ayrıca doğrudan savaş ilanını kapatır; bu faction'la savaş
ilişkisi senaryo event etkileriyle kurulabilir. Bu faction barış öneremez veya
AI barış teklifini kabul edemez; diğer ilişki türleri de açılamaz.
`trade_disabled`, fraksiyonun ticaret işlemlerini kapatır. `war_fatigue_exempt`
ise bu fraksiyonun taraf olduğu savaşları her iki tarafın memnuniyet savaş
yorgunluğu cezası hesabından çıkarır; diğer savaşlar normal hesaplanır.

`research.completed` teknoloji ID'lerini anahtar, tamamlanma durumunu `true` değer
olarak taşıyan bir nesnedir. Dizi biçimi desteklenmez; bu sözleşme yükleme sırasında
`faction.ResearchState.Completed map[string]bool` alanına doğrudan eşlenir.

---

## ai_strategies.json

Senaryoya özgü uzun vadeli AI yönlerini taşıyan opsiyonel dosyadır. Dosya yoksa
fraksiyonlar `ai_strategies.json.expansion_targets` ve genel AI fallback'ini kullanır;
eski `factions.json.ai_expansion_targets` alanı runtime uyumluluğu için türetilir.
`1300_ottoman_rise` ve `298_war_of_five_kings` senaryolarında kullanılır.

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
          "territorial_claims": [
            { "region_id": "aydinoglu", "value": 92 },
            { "region_id": "germiyan", "value": 92 },
            { "region_id": "kutahya", "value": 92 },
            { "region_id": "sivrihisar", "value": 92 }
          ],
          "priority": 92,
          "commitment": 66,
          "min_year": 1300,
          "max_year": 1453,
          "allow_vassalization": true
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
- `territorial_claims`: objective'in bölgesel hedefleri ve ağırlıkları. Hedef
  devlet yazılmaz; AI her değerlendirmede claim bölgesinin güncel sahibini bulur.
- `priority`: aynı anda uygulanabilir objective'ler arasındaki taban öncelik.
- `commitment`: plan save state'ine taşınan `25..90` kararlılık değeri.
- `readiness_regions`: sahip olunan her kayıt için hazırlık bonusu, eksikler için soft
  ceza üretir; objective'i tek başına kapatmaz.
- `min_year`: geç/anakronik hedefi verilen yıla kadar kapatan hard gate.
- `max_year`: objective'in geçerli olduğu son yıl. Verilen yılın tamamında
  geçerlidir; sonraki yılda hard gate kapanır. Son yıllarda AI puanına zaman
  baskısı eklenir.
- `required_event_flags`: event motorunun `set_flags` etkileriyle açılan hard gate listesi.
- `blocked_event_flags`: listelenen event bayraklarından biri set edildiğinde objective'i
  kapatan hard gate listesi. Örneğin bir istila hedefi, savunma hattı kırıldıktan sonra
  yerini güneye ilerleme hedeflerine bırakabilir.
- `allow_vassalization`: uygun son bölge ele geçirildiğinde vassallık zarının
  kullanılmasına izin verir. Uygunluk kontrollerini geçen durumda zar başarılıysa
  hedef vassal kalır; başarısızsa bölge doğrudan ilhak edilir.

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

Başlangıç diplomasi ilişkileri. Dosyada bulunmayan faction çiftleri runtime'da din
temelli varsayılanlarla üretilir; bu varsayılan çiftler Edit Mode kaydında JSON'a
eklenmez. Böylece dosya yalnızca özel olarak tanımlanmış veya editörde açıkça
değiştirilmiş ilişkileri taşır. Edit Mode'da değişiklik yapılmadan kaydetmek,
ilişki kayıtlarının sırasını ve içeriğini değiştirmez.

```json
[
  {
    "faction_a": "ottoman",
    "faction_b": "venice",
    "score_a_to_b": -20,
    "score_b_to_a": -20,
    "stance": "peace"
  },
  {
    "faction_a": "venice",
    "faction_b": "east_rome",
    "score_a_to_b": 35,
    "score_b_to_a": 20,
    "stance": "trade"
  }
]
```

`stance` değerleri: `war`, `peace`, `allied`, `trade`. İlişki puanları editörde
`-100..100` aralığında tutulur.
`score_a_to_b` ve `score_b_to_a` iki yönün puanlarını ayrı ayrı tutar. Eski
senaryolardaki tekil `score` alanı geriye dönük olarak iki yöne birlikte
uygulanır; yeni kayıtlar yönlü alanları yazar.

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
  "required_buildings": [
    { "id": "barracks", "level": 1 }
  ],
  "embarkable": true
}
```

`turns_required` üretim kuyruğunda kaç tur sonra birimin ordu/filoya ekleneceğini belirler. Eksik bırakılırsa yükleyici geriye dönük uyumluluk için `1` kabul eder.

`required_tech` dizisi AND semantiğine sahiptir; listedeki tüm teknoloji ID'leri tamamlanmadan birim üretilemez. Teknoloji zincirinin ara adımları da açıkça yazılır (ör. top için `gunpowder` ve `cast_bronze_cannon`).

`required_buildings` dizisi de AND semantiğine sahiptir. Her kayıt bina ID'sini ve gereken seviyeyi taşır; liste boşsa bina şartı yoktur.

Kuşatma birimleri için opsiyonel `siege_breach_multiplier`, `siege_breach_max_fort_level` ve `siege_defense_pressure` alanları kullanılabilir. İlk alan gedik gücünü çarpar; ikincisi `tier + 2` varsayılanından daha yüksek tahkimatlara erişim verir. `siege_defense_pressure`, tam HP'li birimin kuşatma hücumunda savunma bonusundan düşülecek yüzde puanını belirtir ve canlı HP ile orantılanır.

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

`allowed_factions` ve `blocked_factions` opsiyoneldir. `allowed_factions` doluysa
teknoloji yalnızca listedeki faction ID'leri tarafından araştırılabilir; boş
veya eksikse tüm faction'lara açıktır. `blocked_factions` listesinde bulunan
fraksiyonlar, izin listesinde olsalar bile araştırmayı yapamaz. `min_year`
opsiyonel tarih kilididir. Faction, belirtilen yıldan önce teknolojiye
başlayamaz. Bu alanlar oyuncu, AI ve event kaynaklı araştırma başlatma
akışlarında uygulanır.

`allowed_religions` opsiyoneldir. Doluysa teknoloji yalnızca listedeki dinlerden
birine sahip faction'lara görünür ve araştırılabilir. `allowed_factions` ile
birlikte kullanıldığında iki koşul da sağlanmalıdır. Alan eksik veya boşsa din
kısıtlaması uygulanmaz.

Teknoloji `effects` alanında `gold_upkeep_mod` ve `grain_upkeep_mod` pozitif
oranlar olarak ordu bakımını azaltır. `research_speed_mod` pozitif oranı
araştırma başlatılırken gereken tur sayısını, `research_cost_mod` pozitif oranı
ise altın maliyetini azaltır; devam eden araştırmanın kayıtlı `turns_left` değeri
değiştirilmez.

---

## events.json

Olaylardaki rastgele tetiklenme oranı `random_probability` alanında tutulur.
`0` rastgele tetiklenmeyi kapatır; tarih veya state koşullu tetiklenmeleri
engellemez. Eski senaryo dosyalarındaki `probability` alanı geriye dönük olarak
okunur; yeni kayıtlar `random_probability` yazar.

Tarihsel event'in hedef tarihi geçmiş olsa bile henüz tetiklenmediyse diğer
koşulları sağlandığı ilk turda çalışabilir. `one_shot: true`, event ilk kez
çalıştıktan sonra tekrarını engeller; geç kalmış event'i kendiliğinden iptal
etmez. Önemli karakterlerin hâlâ hayatta olmasını gerektiren event'ler
`requires_alive_commanders` listesinde `characters.json` içindeki komutan ID'lerini
kullanabilir. Koşul, komutanın başlangıç yılına gelinmediyse veya ölüm state'i
set edildiyse başarısız olur. `end_year` doğal ölümü kaydeder; event/choice
effect'indeki `commander_deaths` listesi ise erken ölümü anında kaydeder ve
komutanı ordularından/runtime roster'ından çıkarır. Ölüm state'i save/load
boyunca `DeadCommanderIDs` içinde saklanır; ölümü commander şablonundaki bir
status flag'ine bağlamak gerekmez. `commander_deaths` event kökünde verilirse
bütün seçimlerde, choice effect'inde verilirse yalnız o seçimde uygulanır.
`commander_exiles` / `commander_returns` event veya choice effect'leri komutanı
ölüm state'ine dokunmadan sürgüne gönderir ya da geri döndürür; sürgündeki kişi
alive requirement'ını karşılamaya devam eder. `blocks_flags` belirli bir event
sonucunu ayrıca engellemek için kullanılabilir.

`specific_faction` event veya choice etkisinde opsiyonel `trade_disabled`
boolean alanı hedef fraksiyonun ticaret kilidini değiştirir. `false`, ticareti
açar; alanın bulunmaması mevcut durumu değiştirmez. Değişiklik kompakt ve debug
kayıtlarında saklanır; eski kayıtlarda alan yoksa senaryo varsayılanı korunur.
`diplomacy_disabled` da event veya choice etkisinde opsiyonel boolean olarak
kullanılabilir. `true` normal diplomasi aksiyonlarını kapatır, `false` yeniden
açar; bu değişiklik kompakt ve debug kaydında saklanır. Alan verilmezse mevcut
değer korunur ve eski save'lerde senaryo varsayılanı geçerlidir.

Yaklaşan olay HUD'unda `player_faction` ve oyuncunun `affected_faction` değeriyle
eşleşen `specific_faction` event'leri `Özel` listesine alınır. `all_factions` ve
diğer hedefler yalnızca `Genel` listede gösterilir. Bu sınıflandırma event'in
tetiklenmesini, etkilerini veya normal olay bildirimlerini değiştirmez.

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

Tarihsel olarak `tools/obsolete/populate_all_shapes.py` tarafından Natural Earth `ne_10m_admin_0_countries` şekillerinden üretiliyordu. Bu eski akış artık aktif değildir; küçük kıyı/sınır düzeltmeleri edit mode `Shape` sekmesinden oyun içi paint editor ile yapılabilir.

Format: `{"id": string, "shapes": [{"id": string, "name": string, "rings": [[[x, y], ...]]}]}`. Dosyanın kök `id` değeri bir shape-set kimliğidir; senaryo bunu `map.shape_set_id` ile belirtir ve yükleyici iki değeri eşleştirir. `rings` içindeki koordinatlar shape/scenario uzayındadır ve ondalıklı olabilir; aktif senaryonun `map.shape_offset_*` ve `map.shape_scale_*` alanlarıyla world pikseline dönüştürülür. Edit mode kayıt akışı koordinatları virgülden sonra tek basamakla yazar; tam sayıya yuvarlama dünya pikseli sınırını kaydırabilir.

> **Not:** Eski `assets/data/generated/country_shapes.json` dosyası kaldırılmıştır. Her senaryo kendi `data/country_shapes.json` dosyasına sahiptir; eski üretim araçları `tools/obsolete/` altında yalnızca arşiv olarak tutulur.
