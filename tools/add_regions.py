"""
regions.json'a yeni bolgeler ekler ve komsu listelerini gunceller.
Kullanim: python tools/add_regions.py
"""
import json, sys

REGIONS_PATH = "assets/data/regions.json"

# ── Eklenecek yeni bolgeler ──────────────────────────────────────────────────
NEW_REGIONS = [
    {
        "id": "thrace", "name": "Thrace", "name_tr": "Trakya",
        "terrain": "plain", "owner_id": "",
        "neighbors": ["bulgaria", "constantinople", "greece", "_sea_aegean", "_sea_black"],
        "world_x": 1026, "world_y": 447,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 30, "base_grain_output": 25, "trade_capacity": 2,
        "satisfaction": 60, "tax_rate": 30, "population": 400,
        "religion": "orthodox", "active_event_id": "", "shape_id": "TUR",
    },
    {
        "id": "jordan", "name": "Jordan", "name_tr": "\u00dcrd\u00fcn",
        "terrain": "desert", "owner_id": "",
        "neighbors": ["palestine", "syria", "iraq", "hejaz"],
        "world_x": 1206, "world_y": 645,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 18, "base_grain_output": 12, "trade_capacity": 2,
        "satisfaction": 60, "tax_rate": 30, "population": 250,
        "religion": "sunni", "active_event_id": "", "shape_id": "JOR",
    },
    {
        "id": "azerbaijan", "name": "Azerbaijan", "name_tr": "Azerbaycan",
        "terrain": "plain", "owner_id": "",
        "neighbors": ["georgia", "armenia", "persia_north", "_sea_caspian"],
        "world_x": 1426, "world_y": 452,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 28, "base_grain_output": 22, "trade_capacity": 2,
        "satisfaction": 60, "tax_rate": 30, "population": 450,
        "religion": "sunni", "active_event_id": "", "shape_id": "AZE",
    },
    {
        "id": "armenia", "name": "Armenia", "name_tr": "Ermenistan",
        "terrain": "mountain", "owner_id": "",
        "neighbors": ["georgia", "azerbaijan", "persia_north", "anatolia"],
        "world_x": 1366, "world_y": 462,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 22, "base_grain_output": 18, "trade_capacity": 2,
        "satisfaction": 60, "tax_rate": 30, "population": 300,
        "religion": "orthodox", "active_event_id": "", "shape_id": "ARM",
    },
]

# ── Komsu listesi guncellemeleri (mevcut bolgelere eklenecek) ────────────────
NEIGHBOR_UPDATES = {
    # Trakya'nin komsu olarak eklendigi mevcut bolgeler
    "bulgaria":       {"add": ["thrace"]},
    "constantinople": {"add": ["thrace"]},
    "greece":         {"add": ["thrace"]},
    # Urdun'un komsu olarak eklendigi mevcut bolgeler
    "palestine":      {"add": ["jordan"]},
    "syria":          {"add": ["jordan"]},
    "iraq":           {"add": ["jordan"]},
    "hejaz":          {"add": ["jordan"]},
    # Azerbaycan / Ermenistan'in komsu olarak eklendigi mevcut bolgeler
    "georgia":        {"add": ["azerbaijan", "armenia"]},
    "persia_north":   {"add": ["azerbaijan", "armenia"]},
    "anatolia":       {"add": ["armenia"]},
    "trebizond":      {"add": ["armenia"]},
}

# ── Guncelleme ────────────────────────────────────────────────────────────────
with open(REGIONS_PATH, "r", encoding="utf-8") as f:
    regions = json.load(f)

existing_ids = {r["id"] for r in regions}

added = 0
for nr in NEW_REGIONS:
    if nr["id"] in existing_ids:
        print(f"  ATLANDI (zaten var): {nr['id']}")
        continue
    regions.append(nr)
    added += 1
    print(f"  Eklendi: {nr['id']} ({nr['name_tr']})")

updated_neighbors = 0
for region in regions:
    upd = NEIGHBOR_UPDATES.get(region["id"])
    if not upd:
        continue
    nb = region.get("neighbors", [])
    changed = False
    for new_nb in upd.get("add", []):
        if new_nb not in nb:
            nb.append(new_nb)
            changed = True
    if changed:
        region["neighbors"] = nb
        updated_neighbors += 1
        print(f"  Komsu guncellendi: {region['id']} -> {nb}")

with open(REGIONS_PATH, "w", encoding="utf-8") as f:
    json.dump(regions, f, indent=2, ensure_ascii=False)

print(f"\n{added} bolge eklendi, {updated_neighbors} bolgenin komsu listesi guncellendi.")
print(f"Toplam bolge sayisi: {len(regions)}")
