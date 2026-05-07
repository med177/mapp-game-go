"""
Tarihi fraksiyonlari ekle ve bolgelere 1300 donemi sahipligi ata.
Mimari: her siyasi varlık bir faction, birden cok bolge ayni faction'a ait olabilir.
"""
import json

with open('assets/data/factions.json', encoding='utf-8') as f:
    factions = json.load(f)

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

existing_ids = {f['id'] for f in factions}

# ── Yeni fraksiyonlar ────────────────────────────────────────────────
new_factions = [
    # ── Büyük İmparatorluklar / Krallıklar (oynanabilir) ────────────
    {
        "id": "byzantine",
        "name": "Byzantine Empire",
        "name_tr": "Bizans İmparatorluğu",
        "religion": "orthodox",
        "color": [150, 40, 180],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 700,
        "grain": 180,
        "iron": 80,
        "timber": 60,
        "spice": 70,
        "cloth": 90,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 40
    },
    {
        "id": "hre",
        "name": "Holy Roman Empire",
        "name_tr": "Kutsal Roma İmparatorluğu",
        "religion": "catholic",
        "color": [220, 195, 45],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 900,
        "grain": 300,
        "iron": 200,
        "timber": 180,
        "spice": 40,
        "cloth": 120,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 50
    },
    {
        "id": "serbian_empire",
        "name": "Serbian Empire",
        "name_tr": "Sırp İmparatorluğu",
        "religion": "orthodox",
        "color": [190, 35, 35],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 400,
        "grain": 220,
        "iron": 120,
        "timber": 100,
        "spice": 20,
        "cloth": 60,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 65
    },
    {
        "id": "bulgarian_empire",
        "name": "Second Bulgarian Empire",
        "name_tr": "İkinci Bulgar İmparatorluğu",
        "religion": "orthodox",
        "color": [50, 110, 200],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 350,
        "grain": 250,
        "iron": 100,
        "timber": 90,
        "spice": 20,
        "cloth": 50,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "hungarian_kingdom",
        "name": "Kingdom of Hungary",
        "name_tr": "Macaristan Krallığı",
        "religion": "catholic",
        "color": [210, 75, 35],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 600,
        "grain": 350,
        "iron": 150,
        "timber": 130,
        "spice": 30,
        "cloth": 80,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 60
    },
    {
        "id": "golden_horde",
        "name": "Golden Horde",
        "name_tr": "Altın Orda",
        "religion": "sunni",
        "color": [210, 175, 45],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 550,
        "grain": 200,
        "iron": 80,
        "timber": 50,
        "spice": 60,
        "cloth": 70,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 80
    },
    {
        "id": "ilkhanate",
        "name": "Ilkhanate",
        "name_tr": "İlhanlı Devleti",
        "religion": "sunni",
        "color": [145, 95, 45],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 650,
        "grain": 280,
        "iron": 130,
        "timber": 60,
        "spice": 100,
        "cloth": 90,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 70
    },
    {
        "id": "genoa",
        "name": "Republic of Genoa",
        "name_tr": "Ceneviz Cumhuriyeti",
        "religion": "catholic",
        "color": [215, 215, 215],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 900,
        "grain": 80,
        "iron": 90,
        "timber": 100,
        "spice": 120,
        "cloth": 140,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 45
    },
    {
        "id": "trebizond_emp",
        "name": "Empire of Trebizond",
        "name_tr": "Trabzon Rum İmparatorluğu",
        "religion": "orthodox",
        "color": [55, 175, 120],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 350,
        "grain": 120,
        "iron": 70,
        "timber": 90,
        "spice": 80,
        "cloth": 70,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 35
    },
    {
        "id": "naples_kingdom",
        "name": "Kingdom of Naples",
        "name_tr": "Napoli Krallığı (Anjou)",
        "religion": "catholic",
        "color": [75, 75, 200],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 450,
        "grain": 200,
        "iron": 80,
        "timber": 60,
        "spice": 50,
        "cloth": 70,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "florence_rep",
        "name": "Republic of Florence",
        "name_tr": "Floransa Cumhuriyeti",
        "religion": "catholic",
        "color": [200, 55, 55],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 1100,
        "grain": 100,
        "iron": 70,
        "timber": 60,
        "spice": 90,
        "cloth": 160,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 30
    },
    {
        "id": "polish_kingdom",
        "name": "Kingdom of Poland",
        "name_tr": "Polonya Krallığı",
        "religion": "catholic",
        "color": [200, 45, 45],
        "is_playable": True,
        "is_eliminated": False,
        "gold": 450,
        "grain": 300,
        "iron": 120,
        "timber": 150,
        "spice": 20,
        "cloth": 70,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "papal_states_f",
        "name": "Papal States",
        "name_tr": "Papalık Devletleri",
        "religion": "catholic",
        "color": [230, 230, 230],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 600,
        "grain": 150,
        "iron": 60,
        "timber": 50,
        "spice": 40,
        "cloth": 80,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 20
    },
    # ── Anadolu Beylikleri (AI kontrollü küçük devletler) ────────────
    {
        "id": "karaman_bey",
        "name": "Karamanids",
        "name_tr": "Karamanoğulları Beyliği",
        "religion": "sunni",
        "color": [175, 115, 45],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 250,
        "grain": 150,
        "iron": 60,
        "timber": 40,
        "spice": 30,
        "cloth": 40,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 65
    },
    {
        "id": "germiyan_bey",
        "name": "Germiyanids",
        "name_tr": "Germiyanoğulları Beyliği",
        "religion": "sunni",
        "color": [155, 125, 65],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 200,
        "grain": 120,
        "iron": 50,
        "timber": 40,
        "spice": 20,
        "cloth": 35,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "aydin_bey",
        "name": "Aydinids",
        "name_tr": "Aydınoğulları Beyliği",
        "religion": "sunni",
        "color": [95, 175, 135],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 220,
        "grain": 100,
        "iron": 40,
        "timber": 50,
        "spice": 35,
        "cloth": 50,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 60
    },
    {
        "id": "mentese_bey",
        "name": "Menteşeids",
        "name_tr": "Menteşeoğulları Beyliği",
        "religion": "sunni",
        "color": [75, 155, 165],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 180,
        "grain": 90,
        "iron": 35,
        "timber": 45,
        "spice": 30,
        "cloth": 40,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "hamid_bey",
        "name": "Hamidids",
        "name_tr": "Hamidoğulları Beyliği",
        "religion": "sunni",
        "color": [155, 135, 75],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 190,
        "grain": 110,
        "iron": 45,
        "timber": 40,
        "spice": 25,
        "cloth": 35,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "teke_bey",
        "name": "Teke Beylik",
        "name_tr": "Teke Beyliği",
        "religion": "sunni",
        "color": [115, 160, 85],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 170,
        "grain": 100,
        "iron": 40,
        "timber": 45,
        "spice": 25,
        "cloth": 35,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 55
    },
    {
        "id": "candar_bey",
        "name": "Candarids",
        "name_tr": "Candaroğulları Beyliği",
        "religion": "sunni",
        "color": [75, 125, 165],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 200,
        "grain": 130,
        "iron": 55,
        "timber": 70,
        "spice": 25,
        "cloth": 40,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 50
    },
    {
        "id": "eretna_bey",
        "name": "Eretnids",
        "name_tr": "Eretna Devleti",
        "religion": "sunni",
        "color": [165, 95, 75],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 220,
        "grain": 140,
        "iron": 65,
        "timber": 50,
        "spice": 30,
        "cloth": 45,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 60
    },
    {
        "id": "dulkadir_bey",
        "name": "Dulkadirids",
        "name_tr": "Dulkadiroğulları Beyliği",
        "religion": "sunni",
        "color": [155, 75, 75],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 190,
        "grain": 120,
        "iron": 55,
        "timber": 45,
        "spice": 35,
        "cloth": 40,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 60
    },
    {
        "id": "ramazan_bey",
        "name": "Ramazanids",
        "name_tr": "Ramazanoğulları / Kilikya",
        "religion": "catholic",
        "color": [185, 135, 55],
        "is_playable": False,
        "is_eliminated": False,
        "gold": 200,
        "grain": 130,
        "iron": 50,
        "timber": 45,
        "spice": 40,
        "cloth": 50,
        "tech_points": 0,
        "researched_techs": [],
        "ai_aggressiveness": 50
    },
]

added = 0
for nf in new_factions:
    if nf['id'] not in existing_ids:
        factions.append(nf)
        added += 1
    else:
        print(f"ATLA (zaten var): {nf['id']}")

print(f"Eklenen fraksiyon: {added}, toplam: {len(factions)}")

# ── Bölge sahipligi atamalari (1300 donemi) ──────────────────────────
# {bolge_id: faction_id}
ownership = {
    # Bizans
    "constantinople":   "byzantine",
    "greece":           "byzantine",
    "thessaly":         "byzantine",
    # Mora Prensliği Latin, ama Byzantine'a yakın - bırakalım neutral
    # "morea": "",

    # Kutsal Roma İmparatorluğu
    "brandon":       "hre",
    "saxony":           "hre",
    "bavaria":          "hre",
    "westphalia":       "hre",
    "thuringia":        "hre",
    "palatinate":       "hre",
    "pomerania":        "hre",
    "austria":          "hre",
    "styria":           "hre",
    "tyrol":            "hre",
    "bohemia":          "hre",
    "moravia":          "hre",
    "switzerland":      "hre",
    "lorraine":         "hre",
    "alsace":           "hre",
    "luxembourg":       "hre",

    # Sırp İmparatorluğu
    "serbia":           "serbian_empire",
    "rascia":           "serbian_empire",
    "kosovo":           "serbian_empire",
    "macedonia":        "serbian_empire",

    # Bulgar İmparatorluğu
    "bulgaria":         "bulgarian_empire",
    "vidin":            "bulgarian_empire",

    # Macaristan Krallığı
    "hungary":          "hungarian_kingdom",
    "alfold":           "hungarian_kingdom",
    "transylvania":     "hungarian_kingdom",
    "croatia":          "hungarian_kingdom",
    "slovenia":         "hungarian_kingdom",
    "slovakia":         "hungarian_kingdom",

    # Venedik (venice bölgesi + Dalmaçya, Kıbrıs)
    "venice":           "venice",
    "dalmatia":         "venice",
    "cyprus":           "venice",

    # Ceneviz
    "genoa":            "genoa",
    "crimea":           "genoa",   # Kaffa = Ceneviz ticaret kolonisi

    # Floransa
    "florence":         "florence_rep",
    "siena":            "florence_rep",

    # Papalık
    "papal_states":     "papal_states_f",
    "ferrara":          "papal_states_f",

    # Napoli Krallığı (Anjou)
    "naples":           "naples_kingdom",
    "puglia":           "naples_kingdom",

    # Fransa (mevcut faction)
    "paris":            "france",
    "normandy":         "france",
    "brittany":         "france",
    "anjou":            "france",
    "champagne":        "france",
    "burgundy":         "france",
    "provence":         "france",
    "languedoc":        "france",
    "savoy":            "france",

    # İngiltere (mevcut faction)
    "london":           "england",
    "wessex":           "england",
    "east_anglia":      "england",
    "mercia":           "england",
    "yorkshire":        "england",
    "lancashire":       "england",
    "wales":            "england",

    # Aragon (mevcut)
    "aragon":           "aragon",
    "castile":          "aragon",     # İberya 1300 karmaşık, oyun sadeleştirmesi
    "navarre":          "aragon",
    "sardinia":         "aragon",
    "sicily":           "aragon",

    # Portekiz (mevcut)
    "portugal":         "portugal",

    # Polonya Krallığı
    "poland":           "polish_kingdom",
    "mazovia":          "polish_kingdom",
    "silesia":          "polish_kingdom",
    "galicia":          "polish_kingdom",

    # Moskova Büyük Prensliği (russia faction mevcut)
    "moscow":           "russia",
    "novgorod":         "russia",

    # Altın Orda
    "ukraine":          "golden_horde",
    "kiev":             "golden_horde",
    # crimea: genoa (yukarıda atandı - Kaffa kolonisi)

    # Memlük (mevcut)
    "egypt":            "mamluk",
    "syria":            "mamluk",     # Şam / Damascus
    "aleppo":           "mamluk",
    "palestine":        "mamluk",
    "hejaz":            "mamluk",
    "lebanon":          "mamluk",
    "jordan":           "mamluk",
    "israel":           "mamluk",

    # İlhanlı
    "iraq":             "ilkhanate",  # Bağdat
    "mosul":            "ilkhanate",
    "basra":            "ilkhanate",
    "akkoyunlu":        "ilkhanate",  # 1300'de İlhanlı toprağı
    "eretna":           "eretna_bey", # oyun basitleştirmesi
    "persia_north":     "ilkhanate",
    "persia_west":      "ilkhanate",

    # Safevi (mevcut) - İran güneyi
    "persia_south":     "safavid",   # oyun başlangıç dengesi için

    # Trabzon İmparatorluğu
    "trebizond":        "trebizond_emp",

    # Anadolu Beylikleri
    "bithynia":         "ottoman",
    "karaman":          "karaman_bey",
    "germiyan":         "germiyan_bey",
    "aydinoglu":        "aydin_bey",
    "mentese":          "mentese_bey",
    "hamit":            "hamid_bey",
    "teke":             "teke_bey",
    "candaroglu":       "candar_bey",
    "dulkadir":         "dulkadir_bey",
    "ramazanoglu":      "ramazan_bey",
    "anatolia":         "",           # Selçuklu çöküyor, neutral

    # Gürcistan
    "georgia":          "",           # Bağımsız ama zayıf, neutral

    # Ermenistan / Azerbaycan
    "armenia":          "",
    "azerbaijan":       "",
}

updated = 0
for r in regions:
    if r['id'] in ownership:
        old = r.get('owner_id', '')
        new = ownership[r['id']]
        if old != new:
            r['owner_id'] = new
            updated += 1

print(f"owner_id guncellendi: {updated} bolge")

# ── Germiyan koordinat duzeltmesi ────────────────────────────────────
# Germiyan inland olmali, wy=521 cok guney/kiyiya yakin
for r in regions:
    if r['id'] == 'germiyan':
        r['world_y'] = 498   # 521 -> 498 (iceri tasindi)
        print(f"germiyan wy: 521 -> 498")

# ── Dogrulama ────────────────────────────────────────────────────────
from collections import Counter
faction_ids = {f['id'] for f in factions}
bad_owners = [(r['id'], r['owner_id']) for r in regions
              if r.get('owner_id') and r['owner_id'] not in faction_ids]
if bad_owners:
    print(f"UYARI - gecersiz owner_id: {bad_owners}")
else:
    print("owner_id dogrulama: tum owner'lar gecerli faction")

# Fraksiyon basvuru ozeti
owned = [r for r in regions if r.get('owner_id')]
print(f"\nSahip atanmis bolgeler: {len(owned)} / {len(regions)}")
by_faction = Counter(r['owner_id'] for r in owned)
print("Fraksiyona gore bolge dagilimi (en fazla sahip):")
for fid, cnt in by_faction.most_common(15):
    print(f"  {fid:20} {cnt}")

# ── Kaydet ──────────────────────────────────────────────────────────
with open('assets/data/factions.json', 'w', encoding='utf-8') as f:
    json.dump(factions, f, ensure_ascii=False, indent=2)

with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)

print(f"\nfactions.json ({len(factions)} fraksiyon) ve regions.json kaydedildi")
