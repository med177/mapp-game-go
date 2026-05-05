"""
Buyuk bolge guncellemesi: ada/yarimada/ulke bolgelerini ekler,
mevcut bölgelerin shape_id ve komsu listelerini duzenler.
Kullanim: python tools/add_regions3.py
"""
import json

REGIONS_PATH = "assets/data/regions.json"

# ── 1. Mevcut bölge düzeltmeleri ─────────────────────────────────────────────
SHAPE_FIXES = {
    "thrace": "TRA",   # TUR paylaşimından bağımsız özel poligona geç
    "crimea": "CRM",   # UKR büyük poligonundan yarımada poligonuna geç
}

# Kırım'ın eski komsu listesini tamamen yeni listeyle değiştir
NEIGHBOR_REPLACE = {
    "crimea": ["ukraine", "_sea_black"],
}

# ── 2. Yeni bölgeler ──────────────────────────────────────────────────────────
def region(id, name_tr, shape, x, y, terrain, religion, neighbors,
           gold=25, grain=20, trade=2, pop=400):
    return {
        "id": id, "name": id.replace("_", " ").title(), "name_tr": name_tr,
        "shape_id": shape, "terrain": terrain, "owner_id": "",
        "neighbors": neighbors,
        "world_x": x, "world_y": y,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": gold, "base_grain_output": grain,
        "trade_capacity": trade, "satisfaction": 60, "tax_rate": 30,
        "population": pop, "religion": religion, "active_event_id": "",
    }

NEW_REGIONS = [
    # ── Ukrayna (UKR ana polügonu — Kırım hariç) ──────────────────────────
    region("ukraine", "Ukrayna", "UKR", 1080, 250, "plain", "orthodox",
           ["poland", "lithuania", "wallachia", "moldova", "crimea", "moscow", "belarus"],
           gold=28, grain=35, trade=2, pop=600),

    # ── Königsberg / Doğu Prusya ──────────────────────────────────────────
    region("konigsberg", "Königsberg", "KON", 883, 201, "plain", "catholic",
           ["poland", "lithuania", "_sea_baltic"],
           gold=22, grain=28, trade=2, pop=300),

    # ── Adalar ────────────────────────────────────────────────────────────
    region("sicily",   "Sicilya",  "SCL", 761, 518, "coast",    "catholic",
           ["naples", "tunis", "_sea_med_west"],
           gold=32, grain=28, trade=3, pop=450),
    region("sardinia", "Sardinya", "SAR", 655, 473, "mountain", "catholic",
           ["_sea_med_west"],
           gold=20, grain=22, trade=2, pop=250),
    region("corsica",  "Korsika",  "COR", 656, 434, "mountain", "catholic",
           ["france_south", "_sea_med_west"],
           gold=18, grain=18, trade=2, pop=200),
    region("crete",    "Girit",    "CRT", 973, 565, "coast",    "orthodox",
           ["_sea_aegean", "_sea_med_east"],
           gold=30, grain=24, trade=3, pop=350),
    region("malta",    "Malta",    "MLT", 765, 554, "coast",    "catholic",
           ["tunis", "_sea_med_west"],
           gold=22, grain=12, trade=4, pop=150),

    # ── Levant ────────────────────────────────────────────────────────────
    region("israel",   "İsrail",  "ISR", 1178, 628, "plain", "sunni",
           ["lebanon", "palestine", "jordan", "egypt", "_sea_med_east"],
           gold=25, grain=18, trade=3, pop=300),

    # ── Kuzey Afrika / Boynuz Afrika ─────────────────────────────────────
    region("sudan",    "Sudan",   "SDN", 1087, 940, "desert", "sunni",
           ["egypt", "eritrea", "ethiopia"],
           gold=15, grain=14, trade=2, pop=350),
    region("eritrea",  "Eritre",  "ERI", 1272, 940, "coast",  "orthodox",
           ["sudan", "ethiopia", "_sea_red"],
           gold=12, grain=10, trade=2, pop=200),
    region("ethiopia", "Etiyopya","ETH", 1230, 990, "mountain","orthodox",
           ["sudan", "eritrea"],
           gold=18, grain=22, trade=2, pop=500),
    region("djibouti", "Cibuti",  "DJI", 1330, 990, "coast",  "sunni",
           ["eritrea", "ethiopia", "_sea_red"],
           gold=10, grain=6,  trade=3, pop=80),
    region("yemen",    "Yemen",   "YEM", 1391, 920, "mountain","sunni",
           ["hejaz", "oman", "_sea_red"],
           gold=20, grain=14, trade=3, pop=400),

    # ── Körfez ────────────────────────────────────────────────────────────
    region("oman",    "Umman",   "OMN", 1607, 820, "desert", "sunni",
           ["hejaz", "uae", "yemen", "_sea_persian"],
           gold=22, grain=8,  trade=4, pop=300),
    region("uae",     "BAE",     "ARE", 1571, 760, "desert", "sunni",
           ["oman", "qatar", "hejaz", "_sea_persian"],
           gold=20, grain=6,  trade=3, pop=200),
    region("qatar",   "Katar",   "QAT", 1498, 750, "desert", "sunni",
           ["uae", "bahrain", "hejaz", "_sea_persian"],
           gold=18, grain=5,  trade=3, pop=120),
    region("bahrain", "Bahreyn", "BHR", 1486, 736, "coast",  "sunni",
           ["qatar", "hejaz", "_sea_persian"],
           gold=20, grain=5,  trade=4, pop=100),
]

# ── 3. Mevcut bölge komşu eklemeleri ─────────────────────────────────────────
NEIGHBOR_ADD = {
    # Ukrayna eklenmesiyle
    "poland":     ["konigsberg", "ukraine"],
    "lithuania":  ["konigsberg", "ukraine"],
    "wallachia":  ["ukraine"],
    "moldova":    ["ukraine"],
    "moscow":     ["ukraine"],
    "belarus":    ["ukraine"],
    # Adalar
    "naples":     ["sicily"],
    "tunis":      ["sicily", "malta"],
    "france_south":["corsica"],
    # Levant
    "palestine":  ["israel"],
    "jordan":     ["israel"],
    "egypt":      ["israel", "sudan"],
    # Güney Arabistan / Körfez
    "hejaz":      ["yemen", "oman", "uae", "qatar", "bahrain"],
    "persia_south":["oman", "uae"],
    "kuwait":     ["uae"],
}

# ─────────────────────────────────────────────────────────────────────────────

with open(REGIONS_PATH, "r", encoding="utf-8") as f:
    regions = json.load(f)

by_id = {r["id"]: i for i, r in enumerate(regions)}

# 1. shape_id düzeltmeleri
for rid, new_shape in SHAPE_FIXES.items():
    if rid in by_id:
        old = regions[by_id[rid]].get("shape_id", "")
        regions[by_id[rid]]["shape_id"] = new_shape
        print(f"  shape_id: {rid}: {old} → {new_shape}")

# 2. Komşu listesi değiştirme
for rid, new_nb in NEIGHBOR_REPLACE.items():
    if rid in by_id:
        old = regions[by_id[rid]].get("neighbors", [])
        regions[by_id[rid]]["neighbors"] = new_nb
        print(f"  neighbors replaced: {rid}: {old} → {new_nb}")

# 3. Yeni bölgeler
added = 0
for nr in NEW_REGIONS:
    if nr["id"] in by_id:
        print(f"  ATLANDI (zaten var): {nr['id']}")
        continue
    regions.append(nr)
    by_id[nr["id"]] = len(regions) - 1
    added += 1
    print(f"  Eklendi: {nr['id']} ({nr['name_tr']})")

# 4. Komşu eklemeleri
nb_updated = 0
for rid, adds in NEIGHBOR_ADD.items():
    if rid not in by_id:
        print(f"  UYARI: {rid} bulunamadı (komşu eklenemedi)")
        continue
    nb = regions[by_id[rid]].get("neighbors", [])
    changed = any(a not in nb for a in adds)
    for a in adds:
        if a not in nb:
            nb.append(a)
    if changed:
        regions[by_id[rid]]["neighbors"] = nb
        nb_updated += 1

with open(REGIONS_PATH, "w", encoding="utf-8") as f:
    json.dump(regions, f, indent=2, ensure_ascii=False)

print(f"\n{added} bölge eklendi, {nb_updated} komşu güncellendi. Toplam: {len(regions)}")
