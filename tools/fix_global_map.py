"""
Kapsamli harita duzeltmesi: yeni fraksiyonlar, eksik sahip atamalari,
Rusya poligonu icin yeni bolgeler.

Sorunlar:
  1. RUS poligonu: 2 bolge, devasa Voronoi boslugu -> Novgorod + N.Novgorod eklendi
  2. 14 bölgede owner eksik veya yanlis
  3. Savoy owner=france yanlis (1300'de bagimsiz kontluk)
  4. Baltik bölgeleri (Koenigsberg, Latviya, Estonya) sahipsiz
  5. Avrupa/Kafkasya/Ege'de eksik sahipler
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)
with open('assets/data/factions.json', encoding='utf-8') as f:
    factions = json.load(f)

def get_r(rid):
    return next((r for r in regions if r['id'] == rid), None)

def set_owner(rid, owner):
    r = get_r(rid)
    if r:
        r['owner_id'] = owner
        return True
    return False

def add_nb(rid, nb):
    r = get_r(rid)
    if r and nb not in r.get('neighbors', []):
        r['neighbors'].append(nb)

faction_ids = {f['id'] for f in factions}

# ════════════════════════════════════════════════════════════════
# 1. YENI FRAKSIYONLAR
# ════════════════════════════════════════════════════════════════
new_factions = [
    # Novgorod Cumhuriyeti - 1300'de Rusya'nin en guclu devleti
    {
        "id": "novgorod_rep",
        "name": "Republic of Novgorod",
        "name_tr": "Novgorod Cumhuriyeti",
        "religion": "orthodox",
        "color": [60, 120, 160],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 500, "grain": 300, "iron": 80, "timber": 200,
        "spice": 30, "cloth": 90,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 45
    },
    # Toton Sovalyeleri - Prusya ve Livonya'yi yoneten dini askeri duzen
    {
        "id": "teutonic_order",
        "name": "Teutonic Order",
        "name_tr": "Toton Sovalyeleri",
        "religion": "catholic",
        "color": [220, 220, 220],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 400, "grain": 220, "iron": 120, "timber": 160,
        "spice": 10, "cloth": 50,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 70
    },
    # Gurcistan Kralligi - Kafkasya'nin onemli Hiristiyan devleti
    {
        "id": "georgia_kingdom",
        "name": "Kingdom of Georgia",
        "name_tr": "Gurcistan Kralligi",
        "religion": "orthodox",
        "color": [180, 60, 60],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 320, "grain": 200, "iron": 70, "timber": 90,
        "spice": 40, "cloth": 50,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 55
    },
    # Mora Prensliği (Akhaia Prensligi) - Yunanistan'daki Latin Haclı devleti
    {
        "id": "achaea_principality",
        "name": "Principality of Achaea",
        "name_tr": "Akhaia Prensligi (Mora, Latin)",
        "religion": "catholic",
        "color": [200, 160, 80],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 260, "grain": 180, "iron": 50, "timber": 60,
        "spice": 30, "cloth": 60,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 45
    },
    # Savoy Kontlugu - Fransa ve HRE arasinda bagimsiz alpli devlet
    {
        "id": "savoy_county",
        "name": "County of Savoy",
        "name_tr": "Savoy Kontlugu",
        "religion": "catholic",
        "color": [220, 60, 60],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 250, "grain": 150, "iron": 60, "timber": 80,
        "spice": 20, "cloth": 50,
        "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 40
    },
]

added_factions = 0
for nf in new_factions:
    if nf['id'] not in faction_ids:
        factions.append(nf)
        faction_ids.add(nf['id'])
        added_factions += 1
print(f"Eklenen fraksiyon: {added_factions}")

# ════════════════════════════════════════════════════════════════
# 2. RUS POLIGONU: YENI BOLGELER
# ════════════════════════════════════════════════════════════════
# RUS poligonunda sadece moscow (1293,177) ve kazan (1687,115) var.
# Bu devasa boslugu doldurmak icin 2 yeni bolge ekleniyor.

new_regions = [
    # ── Novgorod: Bati Rusya'nin dominanti, Hansa ortagi ─────────────────
    # Moscow (1293,177) ile Latvia/Estonia arasindaki boslugu doldurur.
    # Tarihsel: Novgorod Cumhuriyeti 862-1478, Bati Rusya'nin en zengin devleti.
    {
        "active_event_id": "",
        "base_gold_income": 120,
        "base_grain_output": 60,
        "id": "novgorod",
        "is_locked": False,
        "is_sea": False,
        "name": "Novgorod",
        "name_tr": "Novgorod Cumhuriyeti",
        "neighbors": ["moscow", "koenigsberg", "latvia", "estonia", "lithuania",
                      "_sea_baltic"],
        "owner_id": "novgorod_rep",
        "population": 800,
        "religion": "orthodox",
        "satisfaction": 70,
        "shape_id": "RUS",
        "tax_rate": 30,
        "terrain": "forest",
        "trade_capacity": 5,
        "unlock_turn": 0,
        "world_x": 1155,
        "world_y": 135
    },
    # ── Nizhny Novgorod: Volga-Oka kavSagi, Moscow ile Kazan arasi ─────────
    # Moscow (1293,177) ile Kazan (1687,115) arasindaki 394px boslugu kapatir.
    # Tarihsel: Nizhny Novgorod Prensliği 1341'de kuruldu, ama bölge 1300'de
    # Moscow'nun nüfuzunda; Volga ticaret merkezi.
    {
        "active_event_id": "",
        "base_gold_income": 55,
        "base_grain_output": 70,
        "id": "nizhny_novgorod",
        "is_locked": False,
        "is_sea": False,
        "name": "Nizhny Novgorod",
        "name_tr": "Nizhny Novgorod (Volga-Oka Kavşağı)",
        "neighbors": ["moscow", "kazan"],
        "owner_id": "russia",
        "population": 320,
        "religion": "orthodox",
        "satisfaction": 62,
        "shape_id": "RUS",
        "tax_rate": 25,
        "terrain": "forest",
        "trade_capacity": 3,
        "unlock_turn": 0,
        "world_x": 1440,
        "world_y": 185
    },
]

existing_ids = {r['id'] for r in regions}
added_regions = 0
for nr in new_regions:
    if nr['id'] not in existing_ids:
        regions.append(nr)
        existing_ids.add(nr['id'])
        added_regions += 1
print(f"Eklenen bolge: {added_regions}")

# RUS komsu guncelleme
add_nb('moscow', 'novgorod')
add_nb('moscow', 'nizhny_novgorod')
add_nb('kazan', 'nizhny_novgorod')
add_nb('koenigsberg', 'novgorod')
add_nb('latvia', 'novgorod')
add_nb('estonia', 'novgorod')
add_nb('lithuania', 'novgorod')

# ════════════════════════════════════════════════════════════════
# 3. SAHIP ATAMALARI
# ════════════════════════════════════════════════════════════════
owner_fixes = {
    # Rusya
    "kazan":        "golden_horde",      # 1300'de Altin Orda kontrolu (Tatarlar)

    # Dogu Avrupa
    "dobruja":      "bulgarian_empire",  # Dobrogea 1300'de Bulgar topragi
    "moldova":      "golden_horde",      # Bogdan/Moldova 1300'de Altin Orda etkisinde

    # Kafkasya
    "georgia":      "georgia_kingdom",   # Gurcistan Kralligi
    "azerbaijan":   "ilkhanate",         # Ilhanlilar'in kalbi (Tebriz baskent)
    # armenia: cekismeli bölge (Mongol/Ermeni), neutral birakiliyor

    # Yunanistan
    "morea":        "achaea_principality",  # Mora/Akhaia Prensligi (Latin Hacli)
    "crete":        "venice",            # Venedik 1204'ten beri Girit'i yonetiyor

    # Bati Avrupa
    "savoy":        "savoy_county",      # DUZELTME: france degil, bagimsiz kontluk
    "corsica":      "genoa",             # Ceneviz Korsikasi (1284'ten)
    "ireland":      "england",           # Anglo-Norman lordluklar, Ingiltere etkisi
    "finland":      "denmark_kingdom",   # 1300'de Isvec egemenliginde (Isvec->Danimarka)

    # Baltik
    "koenigsberg":  "teutonic_order",    # Toton Sovalyeleri Prusyasi
    "latvia":       "teutonic_order",    # Livonya Duzeni (Toton kolu)
    "estonia":      "teutonic_order",    # Livonya Duzeni / Danimarka etkisi
}

fixed = 0
for rid, owner in owner_fixes.items():
    if set_owner(rid, owner):
        fixed += 1
    else:
        print(f"UYARI: {rid} bulunamadi")
print(f"Sahip atamalari: {fixed}/{len(owner_fixes)}")

# ════════════════════════════════════════════════════════════════
# 4. DOGRULAMA
# ════════════════════════════════════════════════════════════════
from collections import Counter

faction_ids_final = {f['id'] for f in factions}
bad = [(r['id'], r['owner_id']) for r in regions
       if r.get('owner_id') and r['owner_id'] not in faction_ids_final]
if bad:
    print(f"UYARI gecersiz owner ({len(bad)}): {bad}")
else:
    print("Tum owner_id'ler gecerli")

id_counts = Counter(r['id'] for r in regions)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

land = [r for r in regions if not r.get('is_sea') and not r['id'].startswith('_')]
neutral = [r for r in land if not r.get('owner_id')]
print(f"\nRUS poligonu:")
for r in sorted([r for r in regions if r.get('shape_id') == 'RUS'], key=lambda x: x['world_x']):
    print(f"  {r['id']:20} wx={r['world_x']:4} wy={r['world_y']:4}  {r.get('owner_id','')}")
print(f"\nToplam: {len(land)} kara bolge")
print(f"Sahipli: {len(land)-len(neutral)}, Sahipsiz: {len(neutral)}")
print(f"Fraksiyon: {len(factions)}")

# ════════════════════════════════════════════════════════════════
# 5. KAYDET
# ════════════════════════════════════════════════════════════════
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
with open('assets/data/factions.json', 'w', encoding='utf-8') as f:
    json.dump(factions, f, ensure_ascii=False, indent=2)
print("Kaydedildi.")
