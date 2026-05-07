"""
Anadolu sinirlarini tarihi hatlara yakinlastir.

Eklenen bolgeler:
1. nicomedia  (wx=1078, wy=450) - Bizans, Bithynia kuzeyini keser -> Karadeniz erisimi kesilir
2. paphlagonia(wx=1142, wy=452) - Candaroglu, Karadeniz kiyisi ortasi
3. lycia      (wx=1082, wy=516) - Teke, Germiyan gueney tampon -> Akdeniz erisimi kesilir
4. saruhan    (wx=1057, wy=480) - Saruhanogullari, Aydinoglu-Germiyan arasi (eksik beylik)

Eklenen fraksiyon:
- saruhan_bey: Saruhanogullari Beyligi
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)
with open('assets/data/factions.json', encoding='utf-8') as f:
    factions = json.load(f)

def get_r(rid):
    return next((r for r in regions if r['id'] == rid), None)

def add_nb(rid, nb):
    r = get_r(rid)
    if r and nb not in r['neighbors']:
        r['neighbors'].append(nb)

def remove_nb(rid, nb):
    r = get_r(rid)
    if r and nb in r['neighbors']:
        r['neighbors'].remove(nb)


# ════════════════════════════════════════════════════════════════
# 1. SARUHAN FRAKSIYONU (eksikti, bolgeyle birlikte ekleniyor)
# ════════════════════════════════════════════════════════════════
faction_ids = {f['id'] for f in factions}
if 'saruhan_bey' not in faction_ids:
    factions.append({
        "id": "saruhan_bey",
        "name": "Principality of Saruhan",
        "name_tr": "Saruhanogullari Beyligi",
        "religion": "sunni",
        "color": [160, 120, 60],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 200, "grain": 120, "iron": 50, "timber": 40, "spice": 20, "cloth": 45,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 55
    })
    print("saruhan_bey fraksiyonu eklendi")
else:
    print("saruhan_bey zaten mevcut")


# ════════════════════════════════════════════════════════════════
# 2. YENI BOLGELER
# ════════════════════════════════════════════════════════════════
new_regions = [
    # ── Nicomedia: Bizans'in Anadolu kuzey kiyisi (Izmit/Kocaeli) ────────
    # Bithynia'nin (1090,468) kuzeyine oturarak Karadeniz erisimini keser.
    # Tarihsel: Nikomedia 1337'ye kadar Bizans elinde kaldi.
    {
        "active_event_id": "",
        "base_gold_income": 42,
        "base_grain_output": 30,
        "id": "nicomedia",
        "is_locked": False,
        "is_sea": False,
        "name": "Nicomedia",
        "name_tr": "Nikomedia (Bizans Sinir Kenti)",
        "neighbors": ["bithynia", "paphlagonia", "candaroglu", "constantinople",
                      "_sea_black", "_sea_aegean"],
        "owner_id": "byzantine",
        "population": 280,
        "religion": "orthodox",
        "satisfaction": 62,
        "shape_id": "TUR",
        "tax_rate": 30,
        "terrain": "coast",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 1078,
        "world_y": 450
    },
    # ── Paphlagonia: Candaroglu'nun bati Karadeniz kiyisi (Bolu/Kastamonu) ─
    # Nikomedia ile Kastamonu arasi: Bithynia'yi Karadeniz'den tamamen keser.
    {
        "active_event_id": "",
        "base_gold_income": 28,
        "base_grain_output": 35,
        "id": "paphlagonia",
        "is_locked": False,
        "is_sea": False,
        "name": "Paphlagonia",
        "name_tr": "Paflagonia (Candarogullari Bati Kiyisi)",
        "neighbors": ["nicomedia", "bithynia", "candaroglu", "_sea_black"],
        "owner_id": "candar_bey",
        "population": 220,
        "religion": "sunni",
        "satisfaction": 60,
        "shape_id": "TUR",
        "tax_rate": 28,
        "terrain": "coast",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 1142,
        "world_y": 452
    },
    # ── Lycia: Germiyan ile Akdeniz arasi tampon (Teke topraklari) ──────────
    # Germiyan'in (1084,498) gueney Voronoi hucresini keser.
    # Tarihsel: Likya (Mugla/Fethiye ardiyolu) Tekeogillarinin nufuz alaninday di.
    {
        "active_event_id": "",
        "base_gold_income": 28,
        "base_grain_output": 28,
        "id": "lycia",
        "is_locked": False,
        "is_sea": False,
        "name": "Lycia",
        "name_tr": "Likya (Teke Beyligi Kuzeyi)",
        "neighbors": ["germiyan", "mentese", "hamit", "teke"],
        "owner_id": "teke_bey",
        "population": 220,
        "religion": "sunni",
        "satisfaction": 60,
        "shape_id": "TUR",
        "tax_rate": 25,
        "terrain": "mountain",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 1082,
        "world_y": 516
    },
    # ── Saruhan: Saruhanogillari Beyligi (Manisa/Izmir hinterlandi) ──────────
    # Aydinoglu (1025,486) ile Germiyan (1084,498) arasinday di, eksikti.
    # Tarihsel: Saruhanogillari Manisa merkezli, Ege kiyisina ulas an guclu beylik.
    {
        "active_event_id": "",
        "base_gold_income": 45,
        "base_grain_output": 35,
        "id": "saruhan",
        "is_locked": False,
        "is_sea": False,
        "name": "Saruhan",
        "name_tr": "Saruhanogullari Beyligi",
        "neighbors": ["aydinoglu", "germiyan", "bithynia", "mentese", "_sea_aegean"],
        "owner_id": "saruhan_bey",
        "population": 300,
        "religion": "sunni",
        "satisfaction": 62,
        "shape_id": "TUR",
        "tax_rate": 30,
        "terrain": "plain",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 1057,
        "world_y": 480
    },
]

regions.extend(new_regions)
print(f"Yeni bolge eklendi: {len(new_regions)}")


# ════════════════════════════════════════════════════════════════
# 3. KOMSU LISTELERI GUNCELLEME
# ════════════════════════════════════════════════════════════════

# bithynia: nicomedia + paphlagonia + saruhan ekle
#   _sea_black KALDIR (nicomedia/paphlagonia kesti)
#   candaroglu KALDIR (paphlagonia araya girdi, Voronoi analizi)
add_nb('bithynia', 'nicomedia')
add_nb('bithynia', 'paphlagonia')
add_nb('bithynia', 'saruhan')
remove_nb('bithynia', '_sea_black')
remove_nb('bithynia', 'candaroglu')

# candaroglu: nicomedia + paphlagonia ekle
add_nb('candaroglu', 'nicomedia')
add_nb('candaroglu', 'paphlagonia')

# constantinople: nicomedia ekle
add_nb('constantinople', 'nicomedia')

# germiyan: lycia + saruhan ekle
#   aydinoglu KALDIR (saruhan araya girdi)
#   candaroglu KALDIR (paphlagonia araya girdi)
add_nb('germiyan', 'lycia')
add_nb('germiyan', 'saruhan')
remove_nb('germiyan', 'aydinoglu')
remove_nb('germiyan', 'candaroglu')

# aydinoglu: saruhan ekle, germiyan kaldir
add_nb('aydinoglu', 'saruhan')
remove_nb('aydinoglu', 'germiyan')

# mentese: lycia + saruhan ekle
add_nb('mentese', 'lycia')
add_nb('mentese', 'saruhan')

# hamit: lycia ekle
add_nb('hamit', 'lycia')

# teke: lycia ekle
add_nb('teke', 'lycia')


# ════════════════════════════════════════════════════════════════
# 4. DOGRULAMA
# ════════════════════════════════════════════════════════════════
from collections import Counter
id_counts = Counter(r['id'] for r in regions)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

faction_ids_final = {f['id'] for f in factions}
bad = [(r['id'], r['owner_id']) for r in regions
       if r.get('owner_id') and r['owner_id'] not in faction_ids_final]
if bad:
    print(f"UYARI gecersiz owner: {bad}")
else:
    print("Tum owner_id'ler gecerli")

tur = [r for r in regions if r.get('shape_id') == 'TUR']
print(f"\nTUR polygon - {len(tur)} bolge:")
for r in sorted(tur, key=lambda x: (x['world_y'], x['world_x'])):
    print(f"  {r['id']:15} wx={r['world_x']:4} wy={r['world_y']:4} | {r['name_tr'][:35]}")

print(f"\nToplam: {len(regions)} bolge, {len(factions)} fraksiyon")


# ════════════════════════════════════════════════════════════════
# 5. KAYDET
# ════════════════════════════════════════════════════════════════
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
with open('assets/data/factions.json', 'w', encoding='utf-8') as f:
    json.dump(factions, f, ensure_ascii=False, indent=2)
print("Kaydedildi.")
