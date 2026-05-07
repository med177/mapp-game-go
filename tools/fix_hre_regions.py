"""HRE bölge düzenleyici: duplicate temizle, yeni alt-bölgeler ekle."""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    data = json.load(f)

# ── 1. Duplicate kayıtları sil ───────────────────────────────────────
# idx 128=burgundy-dup, 129=brittany-dup, 130=bavaria-dup, 131=saxony-dup
remove_indices = {128, 129, 130, 131}
data = [r for i, r in enumerate(data) if i not in remove_indices]
print(f"Duplicate silindi, kalan: {len(data)}")

# ── 2. brandon (7-char, wx=728) → pomerania ─────────────────────────
for r in data:
    if r['id'] == 'brandon' and r['world_x'] == 728:
        r['id'] = 'pomerania'
        r['name'] = 'Pomerania'
        r['name_tr'] = 'Pomeranya'
        r['neighbors'] = ['brandon', 'silesia', 'mazovia', 'konigsberg', '_sea_baltic']
        print("brandon => pomerania")
        break

# ── 3. Komşu listesi yardımcıları ───────────────────────────────────
def add_neighbors(region_id, new_nbs):
    for r in data:
        if r['id'] == region_id:
            for n in new_nbs:
                if n not in r['neighbors']:
                    r['neighbors'].append(n)
            return

def replace_neighbor(region_id, old_nb, new_nb):
    for r in data:
        if r['id'] == region_id:
            r['neighbors'] = [new_nb if x == old_nb else x for x in r['neighbors']]
            return

def set_coords(region_id, wx, wy):
    for r in data:
        if r['id'] == region_id:
            r['world_x'] = wx
            r['world_y'] = wy
            return

# brandon referansını pomerania yap
replace_neighbor('konigsberg', 'brandon', 'pomerania')

# Austria
set_coords('austria', 782, 322)
add_neighbors('austria', ['styria', 'tyrol', 'moravia'])

# Bohemia
set_coords('bohemia', 758, 280)
add_neighbors('bohemia', ['moravia', 'silesia'])

# Poland
set_coords('poland', 865, 228)
add_neighbors('poland', ['silesia', 'mazovia'])

# Hungary
add_neighbors('hungary', ['alfold'])

# Holland
add_neighbors('holland', ['friesland'])

# Champagne
add_neighbors('champagne', ['lorraine'])

# Palatinate
add_neighbors('palatinate', ['lorraine', 'alsace'])

# Bavaria
add_neighbors('bavaria', ['tyrol'])

# Switzerland
add_neighbors('switzerland', ['alsace', 'tyrol'])

# Slovenia
add_neighbors('slovenia', ['styria'])

# Slovakia
add_neighbors('slovakia', ['moravia', 'silesia'])

# Brandenburg (11-char)
add_neighbors('brandon', ['pomerania'])

# Saxony
add_neighbors('saxony', ['pomerania', 'silesia'])

# Transylvania
add_neighbors('transylvania', ['alfold'])

# Burgundy — lorraine/alsace ile bağlantı
add_neighbors('burgundy', ['lorraine', 'alsace'])

# Luxembourg — lorraine ile bağlantı
add_neighbors('luxembourg', ['lorraine'])

print("Komşu güncellemeleri tamam")

# ── 4. Yeni bölgeler ────────────────────────────────────────────────
new_regions = [
    {
        "active_event_id": "",
        "base_gold_income": 35,
        "base_grain_output": 30,
        "id": "styria",
        "is_locked": False,
        "is_sea": False,
        "name": "Styria",
        "name_tr": "Steiermark",
        "neighbors": ["austria", "tyrol", "slovenia", "hungary", "venice"],
        "owner_id": "",
        "population": 320,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "AUT",
        "tax_rate": 30,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 775,
        "world_y": 352
    },
    {
        "active_event_id": "",
        "base_gold_income": 30,
        "base_grain_output": 28,
        "id": "tyrol",
        "is_locked": False,
        "is_sea": False,
        "name": "Tyrol",
        "name_tr": "Tirol",
        "neighbors": ["austria", "styria", "bavaria", "switzerland", "venice"],
        "owner_id": "",
        "population": 280,
        "religion": "catholic",
        "satisfaction": 66,
        "shape_id": "AUT",
        "tax_rate": 28,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 700,
        "world_y": 338
    },
    {
        "active_event_id": "",
        "base_gold_income": 38,
        "base_grain_output": 45,
        "id": "moravia",
        "is_locked": False,
        "is_sea": False,
        "name": "Moravia",
        "name_tr": "Moravya",
        "neighbors": ["bohemia", "austria", "hungary", "slovakia", "silesia", "poland"],
        "owner_id": "",
        "population": 420,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "CZE",
        "tax_rate": 32,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 808,
        "world_y": 292
    },
    {
        "active_event_id": "",
        "base_gold_income": 40,
        "base_grain_output": 50,
        "id": "silesia",
        "is_locked": False,
        "is_sea": False,
        "name": "Silesia",
        "name_tr": "Silezya",
        "neighbors": ["poland", "saxony", "bohemia", "moravia", "mazovia", "brandon", "pomerania"],
        "owner_id": "",
        "population": 450,
        "religion": "catholic",
        "satisfaction": 64,
        "shape_id": "POL",
        "tax_rate": 33,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 792,
        "world_y": 265
    },
    {
        "active_event_id": "",
        "base_gold_income": 35,
        "base_grain_output": 48,
        "id": "mazovia",
        "is_locked": False,
        "is_sea": False,
        "name": "Mazovia",
        "name_tr": "Mazovya",
        "neighbors": ["poland", "silesia", "pomerania", "konigsberg", "lithuania", "_sea_baltic"],
        "owner_id": "",
        "population": 380,
        "religion": "catholic",
        "satisfaction": 63,
        "shape_id": "POL",
        "tax_rate": 30,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 910,
        "world_y": 215
    },
    {
        "active_event_id": "",
        "base_gold_income": 42,
        "base_grain_output": 35,
        "id": "friesland",
        "is_locked": False,
        "is_sea": False,
        "name": "Friesland",
        "name_tr": "Frizelya",
        "neighbors": ["holland", "denmark", "_sea_north", "_sea_baltic"],
        "owner_id": "",
        "population": 300,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "NLD",
        "tax_rate": 30,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 610,
        "world_y": 212
    },
    {
        "active_event_id": "",
        "base_gold_income": 38,
        "base_grain_output": 40,
        "id": "lorraine",
        "is_locked": False,
        "is_sea": False,
        "name": "Lorraine",
        "name_tr": "Lorraine",
        "neighbors": ["champagne", "burgundy", "alsace", "westphalia", "luxembourg", "palatinate"],
        "owner_id": "",
        "population": 360,
        "religion": "catholic",
        "satisfaction": 64,
        "shape_id": "FRA",
        "tax_rate": 32,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 615,
        "world_y": 275
    },
    {
        "active_event_id": "",
        "base_gold_income": 45,
        "base_grain_output": 38,
        "id": "alsace",
        "is_locked": False,
        "is_sea": False,
        "name": "Alsace",
        "name_tr": "Alsas",
        "neighbors": ["lorraine", "palatinate", "switzerland", "burgundy"],
        "owner_id": "",
        "population": 340,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "FRA",
        "tax_rate": 33,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 638,
        "world_y": 295
    },
    {
        "active_event_id": "",
        "base_gold_income": 32,
        "base_grain_output": 60,
        "id": "alfold",
        "is_locked": False,
        "is_sea": False,
        "name": "Alföld",
        "name_tr": "Macar Ovası",
        "neighbors": ["hungary", "transylvania", "wallachia", "serbia", "croatia"],
        "owner_id": "",
        "population": 400,
        "religion": "catholic",
        "satisfaction": 62,
        "shape_id": "HUN",
        "tax_rate": 28,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 905,
        "world_y": 348
    },
]

data.extend(new_regions)
print(f"Yeni bolge eklendi: {len(new_regions)}, toplam: {len(data)}")

# ── 5. Doğrulama: duplicate ID kalmadı mı? ─────────────────────────
from collections import Counter
id_counts = Counter(r['id'] for r in data)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI - Hala duplicate ID var: {dupes}")
else:
    print("ID dogrulama: duplicate yok")

# ── 6. Kaydet ───────────────────────────────────────────────────────
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
print("regions.json kaydedildi")
