"""
Ikinci tur bolge eklemeleri: Lubnan, Kuveyt, Kibris, Flandre, Hollanda, Luksemburg.
Kullanim: python tools/add_regions2.py
"""
import json

REGIONS_PATH = "assets/data/regions.json"

NEW_REGIONS = [
    {
        "id": "lebanon", "name": "Lebanon", "name_tr": "L\u00fcbnan",
        "terrain": "mountain", "owner_id": "",
        "neighbors": ["syria", "palestine", "_sea_med_east"],
        "world_x": 1194, "world_y": 591,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 22, "base_grain_output": 18, "trade_capacity": 3,
        "satisfaction": 60, "tax_rate": 30, "population": 300,
        "religion": "catholic", "active_event_id": "", "shape_id": "LBN",
    },
    {
        "id": "kuwait", "name": "Kuwait", "name_tr": "Basra",
        "terrain": "desert", "owner_id": "",
        "neighbors": ["iraq", "hejaz", "_sea_persian"],
        "world_x": 1431, "world_y": 678,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 20, "base_grain_output": 8, "trade_capacity": 3,
        "satisfaction": 60, "tax_rate": 30, "population": 200,
        "religion": "sunni", "active_event_id": "", "shape_id": "KWT",
    },
    {
        "id": "cyprus", "name": "Cyprus", "name_tr": "K\u0131br\u0131s",
        "terrain": "coast", "owner_id": "",
        "neighbors": ["_sea_med_east"],
        "world_x": 1137, "world_y": 571,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 30, "base_grain_output": 20, "trade_capacity": 4,
        "satisfaction": 65, "tax_rate": 30, "population": 350,
        "religion": "catholic", "active_event_id": "", "shape_id": "CYP",
    },
    {
        "id": "flanders", "name": "Flanders", "name_tr": "Flandre",
        "terrain": "plain", "owner_id": "",
        "neighbors": ["france_north", "hre_west", "hre_north", "holland", "luxembourg", "_sea_north"],
        "world_x": 568, "world_y": 273,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 40, "base_grain_output": 25, "trade_capacity": 4,
        "satisfaction": 65, "tax_rate": 30, "population": 700,
        "religion": "catholic", "active_event_id": "", "shape_id": "BEL",
    },
    {
        "id": "holland", "name": "Holland", "name_tr": "Hollanda",
        "terrain": "coast", "owner_id": "",
        "neighbors": ["flanders", "hre_north", "_sea_north"],
        "world_x": 582, "world_y": 250,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 38, "base_grain_output": 20, "trade_capacity": 4,
        "satisfaction": 65, "tax_rate": 30, "population": 600,
        "religion": "catholic", "active_event_id": "", "shape_id": "NLD",
    },
    {
        "id": "luxembourg", "name": "Luxembourg", "name_tr": "L\u00fcksemburg",
        "terrain": "plain", "owner_id": "",
        "neighbors": ["france_north", "hre_west", "flanders"],
        "world_x": 597, "world_y": 289,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 22, "base_grain_output": 15, "trade_capacity": 2,
        "satisfaction": 65, "tax_rate": 30, "population": 200,
        "religion": "catholic", "active_event_id": "", "shape_id": "LUX",
    },
]

NEIGHBOR_UPDATES = {
    "syria":       {"add": ["lebanon"]},
    "palestine":   {"add": ["lebanon"]},
    "iraq":        {"add": ["kuwait"]},
    "hejaz":       {"add": ["kuwait"]},
    "france_north":{"add": ["flanders", "holland"]},
    "hre_west":    {"add": ["flanders", "luxembourg"]},
    "hre_north":   {"add": ["flanders", "holland"]},
}

with open(REGIONS_PATH, "r", encoding="utf-8") as f:
    regions = json.load(f)

existing = {r["id"] for r in regions}
added = 0
for nr in NEW_REGIONS:
    if nr["id"] in existing:
        print(f"  ATLANDI: {nr['id']}")
        continue
    regions.append(nr)
    added += 1
    print(f"  Eklendi: {nr['id']} ({nr['name_tr']})")

updated = 0
for r in regions:
    upd = NEIGHBOR_UPDATES.get(r["id"])
    if not upd:
        continue
    nb = r.get("neighbors", [])
    changed = False
    for new_nb in upd.get("add", []):
        if new_nb not in nb:
            nb.append(new_nb)
            changed = True
    if changed:
        r["neighbors"] = nb
        updated += 1
        print(f"  Komsu guncellendi: {r['id']}")

with open(REGIONS_PATH, "w", encoding="utf-8") as f:
    json.dump(regions, f, indent=2, ensure_ascii=False)

print(f"\n{added} bolge eklendi, {updated} komsu guncellendi. Toplam: {len(regions)}")
