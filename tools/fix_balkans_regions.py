"""Balkan bolgelerini 1300 donemina gore guncelle ve zenginlestir."""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    data = json.load(f)

def get_r(rid):
    return next((r for r in data if r['id'] == rid), None)

def set_field(rid, field, value):
    r = get_r(rid)
    if r:
        r[field] = value

def set_coords(rid, wx, wy):
    r = get_r(rid)
    if r:
        r['world_x'] = wx
        r['world_y'] = wy

def set_neighbors(rid, nbs):
    r = get_r(rid)
    if r:
        r['neighbors'] = nbs

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

# ── 1. rumeli (BGR) → vidin ─────────────────────────────────────────
vidin_r = get_r('rumeli')
if vidin_r:
    vidin_r['id'] = 'vidin'
    vidin_r['name'] = 'Vidin'
    vidin_r['name_tr'] = 'Vidin (Bulgar Kalesi)'
    vidin_r['world_x'] = 955
    vidin_r['world_y'] = 408
    vidin_r['neighbors'] = ['serbia', 'rascia', 'wallachia', 'bulgaria', 'macedonia']
    vidin_r['base_gold_income'] = 32
    vidin_r['base_grain_output'] = 38
    vidin_r['population'] = 420
    print("rumeli => vidin")

# rumeli referanslarini tum komsu listelerinde vidin yap
for r in data:
    replace_nb(r['id'], 'rumeli', 'vidin')

# ── 2. Bulgaria: Osmanli sahipligini kaldir, koordinat guncelle ──────
set_field('bulgaria', 'owner_id', '')
set_coords('bulgaria', 1015, 415)
set_field('bulgaria', 'name_tr', 'Bulgar Kralligi (Shishman)')
set_neighbors('bulgaria', [
    'wallachia', 'vidin', 'thrace', 'macedonia', 'constantinople', '_sea_black'
])
set_field('bulgaria', 'base_gold_income', 38)

# ── 3. Serbia: koordinat + komsu guncelle ────────────────────────────
set_coords('serbia', 882, 382)
set_field('serbia', 'name_tr', 'Sirbistan (Nemanjic)')
set_neighbors('serbia', [
    'hungary', 'alfold', 'transylvania', 'wallachia',
    'vidin', 'rascia', 'macedonia', 'kosovo'
])

# ── 4. Croatia: koordinat + dalmatia icin hazirlık ───────────────────
set_coords('croatia', 820, 366)
set_field('croatia', 'name_tr', 'Hırvatistan Kralligi')
set_neighbors('croatia', [
    'slovenia', 'hungary', 'alfold', 'rascia',
    'bosnia', 'dalmatia', 'venice', '_sea_adriatic'
])

# ── 5. Bosnia: koordinat + komsu guncelle ────────────────────────────
set_coords('bosnia', 833, 388)
set_field('bosnia', 'name_tr', 'Bosna Banliği')
set_neighbors('bosnia', ['croatia', 'dalmatia', 'rascia', 'hum', 'serbia'])

# ── 6. Albania: koordinat + split icin hazirlık ──────────────────────
set_coords('albania', 862, 447)
set_field('albania', 'name_tr', 'Arnavutluk (Kuzey)')
set_neighbors('albania', [
    'epirus', 'montenegro', 'rascia', 'hum', 'thessaly', '_sea_adriatic'
])

# ── 7. Greece: koordinat + parcala ──────────────────────────────────
set_coords('greece', 948, 492)
set_field('greece', 'name_tr', 'Atina Dukali (Latin)')
set_neighbors('greece', [
    'epirus', 'thessaly', 'morea', 'macedonia', 'thrace',
    '_sea_aegean', '_sea_adriatic'
])

# ── 8. Thrace: komsu listesi duzelt ─────────────────────────────────
set_neighbors('thrace', [
    'bulgaria', 'vidin', 'constantinople', 'greece', '_sea_aegean', '_sea_black'
])

# ── 9. Macedonia: vidin komsu ekle ───────────────────────────────────
set_neighbors('macedonia', ['serbia', 'rascia', 'vidin', 'bulgaria', 'kosovo', 'albania', 'epirus', 'greece', 'thessaly'])

# ── 10. Montenegro: hum ekle ─────────────────────────────────────────
add_nb('montenegro', 'hum')
add_nb('montenegro', 'rascia')
remove_nb('montenegro', 'serbia')

# ── 11. Kosovo: rascia ekle ──────────────────────────────────────────
replace_nb('kosovo', 'serbia', 'rascia')
add_nb('kosovo', 'serbia')
add_nb('kosovo', 'vidin')

# ── 12. Hungary ve alfold: rascia komsusu ────────────────────────────
add_nb('hungary', 'rascia')
remove_nb('hungary', 'serbia')
add_nb('alfold', 'rascia')
remove_nb('alfold', 'serbia')

# ── 13. Wallachia: dobruja ekle ──────────────────────────────────────
add_nb('wallachia', 'dobruja')

# ── 14. Venice: dalmatia ekle ─────────────────────────────────────────
add_nb('venice', 'dalmatia')

# ── 15. Slovenia: dalmatia ekle ──────────────────────────────────────
add_nb('slovenia', 'dalmatia')

# ── 16. Transylvania: komsu guncelle ────────────────────────────────
set_neighbors('transylvania', [
    'hungary', 'alfold', 'wallachia', 'moldova', 'dobruja', 'vidin', 'serbia'
])

# ── 17. Yeni bolgeler ────────────────────────────────────────────────
new_regions = [
    # Rascia (Raska) - Sirbistan'in tarihi kalbi
    {
        "active_event_id": "",
        "base_gold_income": 28,
        "base_grain_output": 40,
        "id": "rascia",
        "is_locked": False,
        "is_sea": False,
        "name": "Rascia",
        "name_tr": "Raska (Sirp Kalbi)",
        "neighbors": ["serbia", "croatia", "dalmatia", "bosnia", "hum", "montenegro", "kosovo"],
        "owner_id": "",
        "population": 480,
        "religion": "orthodox",
        "satisfaction": 65,
        "shape_id": "SRB",
        "tax_rate": 28,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 858,
        "world_y": 412
    },
    # Dalmatia (Dalmasya) - Adriatik kiyisi
    {
        "active_event_id": "",
        "base_gold_income": 55,
        "base_grain_output": 25,
        "id": "dalmatia",
        "is_locked": False,
        "is_sea": False,
        "name": "Dalmatia",
        "name_tr": "Dalmasya",
        "neighbors": ["croatia", "rascia", "hum", "venice", "slovenia", "_sea_adriatic"],
        "owner_id": "",
        "population": 380,
        "religion": "catholic",
        "satisfaction": 65,
        "shape_id": "HRV",
        "tax_rate": 32,
        "terrain": "coast",
        "trade_capacity": 4,
        "unlock_turn": 0,
        "world_x": 792,
        "world_y": 396
    },
    # Hum (Herzegovina) - Bosna guneyindeki prenslik
    {
        "active_event_id": "",
        "base_gold_income": 22,
        "base_grain_output": 30,
        "id": "hum",
        "is_locked": False,
        "is_sea": False,
        "name": "Hum",
        "name_tr": "Hum (Hersek)",
        "neighbors": ["bosnia", "rascia", "dalmatia", "montenegro", "_sea_adriatic"],
        "owner_id": "",
        "population": 320,
        "religion": "catholic",
        "satisfaction": 62,
        "shape_id": "BIH",
        "tax_rate": 25,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 848,
        "world_y": 415
    },
    # Epirus - Guney Arnavutluk / Kuzey Epir
    {
        "active_event_id": "",
        "base_gold_income": 30,
        "base_grain_output": 35,
        "id": "epirus",
        "is_locked": False,
        "is_sea": False,
        "name": "Epirus",
        "name_tr": "Epir Despotlugu",
        "neighbors": ["albania", "greece", "thessaly", "macedonia", "_sea_adriatic", "_sea_aegean"],
        "owner_id": "",
        "population": 300,
        "religion": "orthodox",
        "satisfaction": 62,
        "shape_id": "ALB",
        "tax_rate": 25,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 882,
        "world_y": 470
    },
    # Thessaly - Orta Yunanistan
    {
        "active_event_id": "",
        "base_gold_income": 32,
        "base_grain_output": 42,
        "id": "thessaly",
        "is_locked": False,
        "is_sea": False,
        "name": "Thessaly",
        "name_tr": "Selanik/Tesalya",
        "neighbors": ["epirus", "greece", "morea", "macedonia", "_sea_aegean"],
        "owner_id": "",
        "population": 340,
        "religion": "orthodox",
        "satisfaction": 60,
        "shape_id": "GRC",
        "tax_rate": 25,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 928,
        "world_y": 470
    },
    # Morea (Peloponnese) - Mora Prensliği (Latin)
    {
        "active_event_id": "",
        "base_gold_income": 38,
        "base_grain_output": 40,
        "id": "morea",
        "is_locked": False,
        "is_sea": False,
        "name": "Morea",
        "name_tr": "Mora Prensliği (Latin)",
        "neighbors": ["greece", "thessaly", "_sea_aegean", "_sea_med_east"],
        "owner_id": "",
        "population": 350,
        "religion": "catholic",
        "satisfaction": 60,
        "shape_id": "GRC",
        "tax_rate": 25,
        "terrain": "coast",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 918,
        "world_y": 518
    },
    # Dobruja - Karadeniz kiyisi, Romen-Bulgar siniri
    {
        "active_event_id": "",
        "base_gold_income": 28,
        "base_grain_output": 45,
        "id": "dobruja",
        "is_locked": False,
        "is_sea": False,
        "name": "Dobruja",
        "name_tr": "Dobrica (Karadeniz Kiyisi)",
        "neighbors": ["wallachia", "transylvania", "bulgaria", "_sea_black"],
        "owner_id": "",
        "population": 280,
        "religion": "orthodox",
        "satisfaction": 60,
        "shape_id": "ROU",
        "tax_rate": 25,
        "terrain": "plain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 1042,
        "world_y": 395
    },
]

data.extend(new_regions)
print(f"Yeni bolge eklendi: {len(new_regions)}")

# ── Dogrulama ────────────────────────────────────────────────────────
from collections import Counter
id_counts = Counter(r['id'] for r in data)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

balkan_shapes = ['SRB','BGR','BIH','HRV','ALB','GRC','ROU','MDA','TRA','KOS','MNE','MKD','SVN']
print("\nBalkan poligon ozeti:")
for s in balkan_shapes:
    regs = [r for r in data if r.get('shape_id') == s]
    if regs:
        print(f"  {s} ({len(regs)}): {[r['id'] for r in regs]}")

print(f"\nToplam bolge: {len(data)}")

# ── Kaydet ──────────────────────────────────────────────────────────
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
print("regions.json kaydedildi")
