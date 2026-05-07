"""Türkiye/Anadolu, Ukrayna, Suriye, Irak bolgelerini 1300 donemina guncelle."""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    data = json.load(f)

def get_r(rid):
    return next((r for r in data if r['id'] == rid), None)

def set_field(rid, field, value):
    r = get_r(rid)
    if r: r[field] = value

def set_coords(rid, wx, wy):
    r = get_r(rid)
    if r:
        r['world_x'] = wx
        r['world_y'] = wy

def set_neighbors(rid, nbs):
    r = get_r(rid)
    if r: r['neighbors'] = nbs

def add_nb(rid, nb):
    r = get_r(rid)
    if r and nb not in r['neighbors']:
        r['neighbors'].append(nb)

def remove_nb(rid, nb):
    r = get_r(rid)
    if r and nb in r['neighbors']:
        r['neighbors'].remove(nb)

def replace_nb(rid, old, new):
    r = get_r(rid)
    if r:
        r['neighbors'] = [new if x == old else x for x in r['neighbors']]


# ════════════════════════════════════════════════════════════════
# 1. TÜRKİYE / ANADOLU
# ════════════════════════════════════════════════════════════════

# anatolia: owner=ottoman yanlis, Sultanate of Rum temsil ediyor
set_field('anatolia', 'owner_id', '')
set_field('anatolia', 'name_tr', 'Anadolu Selcuklu Sultanligi')
set_field('anatolia', 'name', 'Sultanate of Rum')
set_coords('anatolia', 1175, 462)
set_neighbors('anatolia', [
    'bithynia', 'candaroglu', 'eretna', 'hamit', 'karaman', '_sea_black'
])

# Constantinople: bithynia komsusu ekle
set_neighbors('constantinople', [
    'bithynia', 'bulgaria', 'greece', 'thrace',
    '_sea_aegean', '_sea_black'
])

# hamit: bithynia ekle
add_nb('hamit', 'bithynia')
remove_nb('hamit', 'anatolia')
remove_nb('hamit', 'constantinople')

# candaroglu: anatolia'yi guncelle
replace_nb('candaroglu', 'anatolia', 'anatolia')  # stays same id

# ramazanoglu: 1300'de Kilikya Ermeni Kralligi var bu alanda
set_field('ramazanoglu', 'name', 'Cilicia')
set_field('ramazanoglu', 'name_tr', 'Kilikya Ermeni Kralligi')
replace_nb('ramazanoglu', 'syria', 'aleppo')

# dulkadir: syria -> aleppo (dulkadir kuzey Suriye sinirina bakar)
replace_nb('dulkadir', 'syria', 'aleppo')
add_nb('dulkadir', 'mosul')

# akkoyunlu: 1300'de bu alan Ilhanli kontrolunde, ad guncel tut ama not
set_field('akkoyunlu', 'name_tr', 'Akkoyunlu (Ilhanli Topraklari)')
add_nb('akkoyunlu', 'mosul')


# ════════════════════════════════════════════════════════════════
# 2. UKRAYNA
# ════════════════════════════════════════════════════════════════

# Mevcut ukraine: dogu stepine tasinan Kiev bolgesi
set_coords('ukraine', 1160, 305)
set_field('ukraine', 'name', 'Ukrainian Steppe')
set_field('ukraine', 'name_tr', 'Ukrayna Bozkiri (Altin Orda)')
set_neighbors('ukraine', [
    'kiev', 'galicia', 'moldova', 'wallachia', 'crimea', 'moscow'
])

# Komsu guncellemeleri
replace_nb('belarus', 'ukraine', 'kiev')
replace_nb('moscow', 'ukraine', 'kiev')
add_nb('moscow', 'ukraine')
replace_nb('lithuania', 'ukraine', 'galicia')
add_nb('lithuania', 'kiev')
replace_nb('poland', 'ukraine', 'galicia')
replace_nb('mazovia', 'ukraine', 'galicia')
add_nb('mazovia', 'kiev')
replace_nb('moldova', 'ukraine', 'galicia')
add_nb('moldova', 'ukraine')


# ════════════════════════════════════════════════════════════════
# 3. SURİYE
# ════════════════════════════════════════════════════════════════

# syria: Samiye/Kusey Suriye -> guncelleme, Halep ayri
set_coords('syria', 1242, 580)
set_field('syria', 'name', 'Damascus')
set_field('syria', 'name_tr', 'Sam (Memluk)')
set_neighbors('syria', [
    'aleppo', 'lebanon', 'palestine', 'jordan', 'mosul', 'basra'
])
remove_nb('ramazanoglu', 'syria')

# lebanon: aleppo ekle
add_nb('lebanon', 'aleppo')


# ════════════════════════════════════════════════════════════════
# 4. IRAK
# ════════════════════════════════════════════════════════════════

# iraq: Bagdat merkezi
set_coords('iraq', 1345, 600)
set_field('iraq', 'name', 'Baghdad')
set_field('iraq', 'name_tr', 'Bagdat (Ilhanli)')
set_neighbors('iraq', [
    'mosul', 'basra', 'syria', 'persia_west', 'kuwait'
])
remove_nb('persia_west', 'iraq')
add_nb('persia_west', 'mosul')
add_nb('persia_west', 'basra')
add_nb('persia_west', 'iraq')


# ════════════════════════════════════════════════════════════════
# 5. YENİ BÖLGELER
# ════════════════════════════════════════════════════════════════

new_regions = [
    # ── Bithynia (Erken Osmanli beyligi, Bursa bolgesi) ──────────
    {
        "active_event_id": "",
        "base_gold_income": 35,
        "base_grain_output": 42,
        "id": "bithynia",
        "is_locked": False,
        "is_sea": False,
        "name": "Bithynia",
        "name_tr": "Bitinya (Erken Osmanli Beyligi)",
        "neighbors": ["constantinople", "anatolia", "hamit", "germiyan", "candaroglu", "_sea_black", "_sea_aegean"],
        "owner_id": "",
        "population": 380,
        "religion": "sunni",
        "satisfaction": 62,
        "shape_id": "TUR",
        "tax_rate": 28,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 1090,
        "world_y": 468
    },
    # ── Galicia-Volhynia (Bati Ukrayna) ──────────────────────────
    {
        "active_event_id": "",
        "base_gold_income": 30,
        "base_grain_output": 48,
        "id": "galicia",
        "is_locked": False,
        "is_sea": False,
        "name": "Galicia-Volhynia",
        "name_tr": "Galicya-Volhiyn Prensliği",
        "neighbors": ["poland", "mazovia", "silesia", "transylvania", "moldova", "kiev", "ukraine", "lithuania"],
        "owner_id": "",
        "population": 420,
        "religion": "orthodox",
        "satisfaction": 60,
        "shape_id": "UKR",
        "tax_rate": 28,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 992,
        "world_y": 262
    },
    # ── Kiev (Kiev Prensliği, Altin Orda vasali) ──────────────────
    {
        "active_event_id": "",
        "base_gold_income": 28,
        "base_grain_output": 45,
        "id": "kiev",
        "is_locked": False,
        "is_sea": False,
        "name": "Kiev",
        "name_tr": "Kiev Prensliği (Altin Orda)",
        "neighbors": ["galicia", "ukraine", "belarus", "lithuania", "mazovia", "moscow"],
        "owner_id": "",
        "population": 380,
        "religion": "orthodox",
        "satisfaction": 58,
        "shape_id": "UKR",
        "tax_rate": 25,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 1092,
        "world_y": 262
    },
    # ── Aleppo (Kuzey Suriye, Memluk) ────────────────────────────
    {
        "active_event_id": "",
        "base_gold_income": 55,
        "base_grain_output": 40,
        "id": "aleppo",
        "is_locked": False,
        "is_sea": False,
        "name": "Aleppo",
        "name_tr": "Halep (Memluk)",
        "neighbors": ["ramazanoglu", "dulkadir", "eretna", "damascus", "lebanon", "mosul"],
        "owner_id": "mamluk",
        "population": 480,
        "religion": "sunni",
        "satisfaction": 65,
        "shape_id": "SYR",
        "tax_rate": 35,
        "terrain": "plain",
        "trade_capacity": 4,
        "unlock_turn": 0,
        "world_x": 1222,
        "world_y": 547
    },
    # ── Mosul (Kuzey Irak, Ilhanli) ──────────────────────────────
    {
        "active_event_id": "",
        "base_gold_income": 40,
        "base_grain_output": 35,
        "id": "mosul",
        "is_locked": False,
        "is_sea": False,
        "name": "Mosul",
        "name_tr": "Musul (Ilhanli)",
        "neighbors": ["aleppo", "iraq", "basra", "armenia", "akkoyunlu", "dulkadir", "eretna"],
        "owner_id": "",
        "population": 360,
        "religion": "sunni",
        "satisfaction": 58,
        "shape_id": "IRQ",
        "tax_rate": 28,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 1328,
        "world_y": 564
    },
    # ── Basra (Guney Irak, Körfez kapisi) ────────────────────────
    {
        "active_event_id": "",
        "base_gold_income": 45,
        "base_grain_output": 38,
        "id": "basra",
        "is_locked": False,
        "is_sea": False,
        "name": "Basra",
        "name_tr": "Basra (Ilhanli)",
        "neighbors": ["iraq", "mosul", "persia_west", "kuwait", "_sea_persian"],
        "owner_id": "",
        "population": 320,
        "religion": "sunni",
        "satisfaction": 58,
        "shape_id": "IRQ",
        "tax_rate": 28,
        "terrain": "coast",
        "trade_capacity": 4,
        "unlock_turn": 0,
        "world_x": 1390,
        "world_y": 652
    },
]

# aleppo ve damascus icin syria referansini duzelt
# (aleppo neighbors'ta 'damascus' = 'syria' ID'si)
# aleppo'yu ekledikten sonra neighbors duzeltecegiz
data.extend(new_regions)

# aleppo neighbors 'damascus' -> 'syria' (ID degismedi)
for r in data:
    if r['id'] == 'aleppo':
        r['neighbors'] = ['ramazanoglu', 'dulkadir', 'eretna', 'syria', 'lebanon', 'mosul']

print(f"Yeni bolge eklendi: {len(new_regions)}")

# ════════════════════════════════════════════════════════════════
# 6. DOGRULAMA
# ════════════════════════════════════════════════════════════════
from collections import Counter
id_counts = Counter(r['id'] for r in data)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

print("\nBolge ozeti:")
for s, label in [('TUR','Anadolu'), ('UKR','Ukrayna'), ('SYR','Suriye'), ('IRQ','Irak'), ('BGR','Bulgaristan')]:
    regs = [r for r in data if r.get('shape_id') == s]
    print(f"  {label} ({s}) {len(regs)} bolge: {[r['id'] for r in regs]}")

print(f"\nToplam bolge: {len(data)}")

# ════════════════════════════════════════════════════════════════
# 7. KAYDET
# ════════════════════════════════════════════════════════════════
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
print("regions.json kaydedildi")
