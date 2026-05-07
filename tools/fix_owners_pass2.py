"""Kalan sahiplik duzeltmeleri + eksik kucuk fraksiyonlar."""
import json

with open('assets/data/factions.json', encoding='utf-8') as f:
    factions = json.load(f)
with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

existing_ids = {f['id'] for f in factions}

# ── Eksik kucuk fraksiyonlar ─────────────────────────────────────────
extra_factions = [
    {"id": "lithuanian_gd",  "name": "Grand Duchy of Lithuania", "name_tr": "Litvanya Büyük Prensliği",
     "religion": "catholic", "color": [80, 150, 80], "is_playable": True, "is_eliminated": False,
     "gold": 350, "grain": 250, "iron": 100, "timber": 160, "spice": 15, "cloth": 55,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 65},

    {"id": "wallachia_prince", "name": "Principality of Wallachia", "name_tr": "Eflak Prensliği",
     "religion": "orthodox", "color": [170, 90, 40], "is_playable": False, "is_eliminated": False,
     "gold": 200, "grain": 180, "iron": 60, "timber": 80, "spice": 10, "cloth": 30,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 45},

    {"id": "granada_emirate", "name": "Emirate of Granada", "name_tr": "Gırnata Emirliği",
     "religion": "sunni", "color": [180, 50, 180], "is_playable": True, "is_eliminated": False,
     "gold": 400, "grain": 130, "iron": 70, "timber": 50, "spice": 60, "cloth": 80,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 50},

    {"id": "milan_duchy",  "name": "Duchy of Milan", "name_tr": "Milano Dükalığı (Visconti)",
     "religion": "catholic", "color": [160, 90, 160], "is_playable": True, "is_eliminated": False,
     "gold": 600, "grain": 160, "iron": 120, "timber": 80, "spice": 50, "cloth": 100,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 55},

    {"id": "denmark_kingdom", "name": "Kingdom of Denmark", "name_tr": "Danimarka Krallığı",
     "religion": "catholic", "color": [200, 40, 40], "is_playable": True, "is_eliminated": False,
     "gold": 350, "grain": 200, "iron": 80, "timber": 120, "spice": 20, "cloth": 60,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 50},

    {"id": "castile_kingdom", "name": "Kingdom of Castile", "name_tr": "Kastilya Krallığı",
     "religion": "catholic", "color": [210, 170, 40], "is_playable": True, "is_eliminated": False,
     "gold": 500, "grain": 200, "iron": 100, "timber": 70, "spice": 40, "cloth": 80,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 60},

    {"id": "scotland_kingdom", "name": "Kingdom of Scotland", "name_tr": "İskoçya Krallığı",
     "religion": "catholic", "color": [50, 90, 170], "is_playable": True, "is_eliminated": False,
     "gold": 220, "grain": 150, "iron": 70, "timber": 100, "spice": 10, "cloth": 40,
     "tech_points": 0, "researched_techs": [], "ai_aggressiveness": 55},
]

added = 0
for nf in extra_factions:
    if nf['id'] not in existing_ids:
        factions.append(nf)
        existing_ids.add(nf['id'])
        added += 1
print(f"Ek fraksiyon eklendi: {added}, toplam: {len(factions)}")

# ── Owner atamalari: ID encoding sorunu olan bolgeleri bul ve ata ─────
def set_owner(region_id_or_prefix, new_owner, use_prefix=False):
    for r in regions:
        if use_prefix:
            if r['id'].startswith(region_id_or_prefix):
                r['owner_id'] = new_owner
                return r['id']
        else:
            if r['id'] == region_id_or_prefix:
                r['owner_id'] = new_owner
                return r['id']
    return None

# Brandenburg - prefix ile bul
result = set_owner('brandenbur', 'hre', use_prefix=True)
print(f"Brandenburg: {result} -> hre")

# Kalan önemli bölge atamaları
extra_owners = {
    # HRE bölgeleri (prefix sorunlari olabilir, ID ile dene)
    "flanders":         "hre",
    "holland":          "hre",
    "friesland":        "hre",
    "denmark":          "denmark_kingdom",
    "friesland":        "hre",

    # İtalya
    "milan":            "milan_duchy",
    "verona":           "milan_duchy",     # Verona, Visconti'nin nüfuz alaninday di

    # İberya
    "castile":          "castile_kingdom",
    "granada":          "granada_emirate",
    "navarre":          "aragon",

    # İskandinav
    "norway":           "denmark_kingdom", # 1300'de Norveç Danimarka etkisinde
    "sweden":           "denmark_kingdom", # Kalmar Birliği öncesi

    # Doğu Avrupa
    "wallachia":        "wallachia_prince",
    "lithuania":        "lithuanian_gd",
    "belarus":          "lithuanian_gd",  # Beyaz Rusya Litvanya etkisinde

    # Trakya / çekişmeli
    "thrace":           "byzantine",      # 1300'de Bizans elinde

    # Boşna
    "bosnia":           "",               # Bağımsız Ban - neutral
    "hum":              "",               # Neutral

    # İskoçya
    "scotland":         "scotland_kingdom",

    # Kuzey Afrika - Mamlük/bağımsız
    "tripolitania":     "",
    "algiers":          "",
    "morocco":          "",
    "tunis":            "",
}

updated = 0
for rid, owner in extra_owners.items():
    for r in regions:
        if r['id'] == rid:
            r['owner_id'] = owner
            updated += 1
            break

print(f"Ek owner ataması: {updated} bolge")

# ── Dogrulama ────────────────────────────────────────────────────────
faction_ids = {f['id'] for f in factions}
bad = [(r['id'], r['owner_id']) for r in regions
       if r.get('owner_id') and r['owner_id'] not in faction_ids]
if bad:
    print(f"UYARI gecersiz owner: {bad}")
else:
    print("Tum owner_id'ler gecerli")

from collections import Counter
owned = [r for r in regions if r.get('owner_id')]
neutral = [r for r in regions if not r.get('owner_id') and not r.get('is_sea')]
print(f"\nSahipli: {len(owned)}, Neutral: {len(neutral)}, Deniz: {sum(1 for r in regions if r.get('is_sea'))}")

# ── Kaydet ──────────────────────────────────────────────────────────
with open('assets/data/factions.json', 'w', encoding='utf-8') as f:
    json.dump(factions, f, ensure_ascii=False, indent=2)
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
print(f"Kaydedildi. Toplam: {len(factions)} fraksiyon, {len(regions)} bolge")
