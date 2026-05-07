"""İtalya bölgelerini 1300 dönemine göre güncelle ve genişlet."""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    data = json.load(f)

def get_region(rid):
    return next((r for r in data if r['id'] == rid), None)

def set_field(rid, field, value):
    r = get_region(rid)
    if r:
        r[field] = value

def set_coords(rid, wx, wy):
    r = get_region(rid)
    if r:
        r['world_x'] = wx
        r['world_y'] = wy

def set_neighbors(rid, nbs):
    r = get_region(rid)
    if r:
        r['neighbors'] = nbs

def add_neighbor(rid, nb):
    r = get_region(rid)
    if r and nb not in r['neighbors']:
        r['neighbors'].append(nb)

def remove_neighbor(rid, nb):
    r = get_region(rid)
    if r and nb in r['neighbors']:
        r['neighbors'].remove(nb)

# ── 1. Mevcut ITA bölgelerini 1300 donemina uygun guncelle ────────────

# Venice: koordinat ve komsu guncelleme
set_coords('venice', 728, 358)
set_neighbors('venice', [
    'verona', 'ferrara', 'switzerland', 'austria',
    'hungary', 'slovenia', 'croatia', '_sea_adriatic'
])
set_field('venice', 'name_tr', 'Venedik Cumhuriyeti')

# Milan: Visconti donemi (1295+), koordinat guncelle
set_coords('milan', 668, 372)
set_neighbors('milan', [
    'genoa', 'verona', 'savoy', 'switzerland', 'florence'
])
set_field('milan', 'name_tr', 'Milano (Visconti)')
set_field('milan', 'base_gold_income', 85)

# Genoa: Ceneviz Cumhuriyeti, Akdeniz ticaret gucu
set_coords('genoa', 633, 385)
set_neighbors('genoa', [
    'milan', 'florence', 'siena', 'savoy', '_sea_med_west'
])
set_field('genoa', 'terrain', 'coast')

# Florence: Toskana cumhuriyeti, bankacilk merkezi
set_coords('florence', 697, 413)
set_neighbors('florence', [
    'milan', 'verona', 'genoa', 'siena', 'papal_states', '_sea_med_west'
])
set_field('florence', 'name_tr', 'Floransa Cumhuriyeti')
set_field('florence', 'base_gold_income', 90)

# Papal States: Orta Italya, Lazio/Umbria/Marche
set_coords('papal_states', 752, 444)
set_neighbors('papal_states', [
    'ferrara', 'florence', 'siena', 'naples', 'puglia',
    '_sea_med_west', '_sea_adriatic'
])
set_field('papal_states', 'name_tr', 'Papalık Devletleri')
set_field('papal_states', 'satisfaction', 78)

# Naples: Angevin Kingdom of Naples (1300'de hala Anjouv hanedanı)
set_coords('naples', 793, 474)
set_neighbors('naples', [
    'papal_states', 'puglia', 'sicily', 'tunis',
    '_sea_med_west', '_sea_med_east'
])
set_field('naples', 'name_tr', 'Napoli Kralligi (Anjou)')
set_field('naples', 'owner_id', '')
set_field('naples', 'base_gold_income', 50)
set_field('naples', 'base_grain_output', 55)

# ── 2. Yeni 1300 donemi Italya bolgeleri ─────────────────────────────
new_regions = [
    {
        "active_event_id": "",
        "base_gold_income": 70,
        "base_grain_output": 45,
        "id": "verona",
        "is_locked": False,
        "is_sea": False,
        "name": "Verona",
        "name_tr": "Verona (Scaligeri)",
        "neighbors": ["milan", "venice", "ferrara", "florence", "genoa"],
        "owner_id": "",
        "population": 550,
        "religion": "catholic",
        "satisfaction": 66,
        "shape_id": "ITA",
        "tax_rate": 35,
        "terrain": "plain",
        "trade_capacity": 4,
        "unlock_turn": 0,
        "world_x": 698,
        "world_y": 373
    },
    {
        "active_event_id": "",
        "base_gold_income": 55,
        "base_grain_output": 50,
        "id": "ferrara",
        "is_locked": False,
        "is_sea": False,
        "name": "Ferrara",
        "name_tr": "Ferrara (Este)",
        "neighbors": ["venice", "verona", "milan", "papal_states"],
        "owner_id": "",
        "population": 420,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "ITA",
        "tax_rate": 32,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 735,
        "world_y": 395
    },
    {
        "active_event_id": "",
        "base_gold_income": 65,
        "base_grain_output": 38,
        "id": "siena",
        "is_locked": False,
        "is_sea": False,
        "name": "Siena",
        "name_tr": "Siena Cumhuriyeti",
        "neighbors": ["florence", "genoa", "papal_states", "_sea_med_west"],
        "owner_id": "",
        "population": 480,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "ITA",
        "tax_rate": 33,
        "terrain": "plain",
        "trade_capacity": 4,
        "unlock_turn": 0,
        "world_x": 710,
        "world_y": 432
    },
    {
        "active_event_id": "",
        "base_gold_income": 35,
        "base_grain_output": 60,
        "id": "puglia",
        "is_locked": False,
        "is_sea": False,
        "name": "Apulia",
        "name_tr": "Apulia (Puglia)",
        "neighbors": ["naples", "papal_states", "_sea_adriatic", "_sea_med_east"],
        "owner_id": "",
        "population": 450,
        "religion": "catholic",
        "satisfaction": 62,
        "shape_id": "ITA",
        "tax_rate": 28,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 855,
        "world_y": 462
    },
]

data.extend(new_regions)
print(f"Yeni ITA bolgesi eklendi: {len(new_regions)}")

# ── 3. Savoy: ITA poligonuna tasimak yerine komsu listesini duzelt ───
# (Savoy FRA poligonunda kaliyor ama komsu listesi guncelleniyor)
set_neighbors('savoy', ['genoa', 'milan', 'switzerland', 'provence', 'burgundy'])
# Switzerland: milan'i ekle
add_neighbor('switzerland', 'milan')
# Slovenia: venice komsusu (zaten var, kontrol)

# ── 4. Dogrulama ─────────────────────────────────────────────────────
from collections import Counter
id_counts = Counter(r['id'] for r in data)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

ita = [r for r in data if r.get('shape_id') == 'ITA']
print(f"\nITA poligonu - {len(ita)} bolge:")
for r in ita:
    print(f"  {r['id']:15} wx={r['world_x']:4} wy={r['world_y']:4} | {r['name_tr']}")

print(f"\nToplam bolge: {len(data)}")

# ── 5. Kaydet ────────────────────────────────────────────────────────
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
print("regions.json kaydedildi")
