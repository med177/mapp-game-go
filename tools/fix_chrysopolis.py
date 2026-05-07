"""
Hrisopolis (Chrysopolis) ekle: Bithynia'yi Karadeniz'den kesin olarak kes.

Sorun:
  Nicomedia (1078, 450) x ekseni farkli, Bithynia (1090, 468) ile kuzey
  kiyisi arasinda hala Voronoi bogazi kaliyor.

Cozum:
  Chrysopolis (1090, 445): Bithynia'nin tam kuzeyinde.
  Gecis wy = (445+468)/2 = 456.5
  => Kiyinin (wy~445-455) hepsi Chrysopolis (Bizans) rengiyle gosterilir.

Tarihsel: Hrisopolis (modern Uskudar/Kadikoy Asian yakasi) 1326'ya dek Bizans.
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

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


# ─── Yeni bolge ──────────────────────────────────────────────────────────────
chrysopolis = {
    "active_event_id": "",
    "base_gold_income": 35,
    "base_grain_output": 20,
    "id": "chrysopolis",
    "is_locked": False,
    "is_sea": False,
    "name": "Chrysopolis",
    "name_tr": "Hrisopolis (Bizans Asya Yakasi)",
    # Nicomedia (1078,450) ile Paphlagonia (1142,452) arasinda,
    # Bithynia (1090,468) nin tam kuzeyinde.
    "world_x": 1090,
    "world_y": 445,
    "shape_id": "TUR",
    "owner_id": "byzantine",
    "neighbors": ["nicomedia", "paphlagonia", "bithynia", "constantinople",
                  "_sea_black"],
    "population": 180,
    "religion": "orthodox",
    "satisfaction": 65,
    "tax_rate": 28,
    "terrain": "coast",
    "trade_capacity": 3,
    "unlock_turn": 0
}
regions.append(chrysopolis)
print("chrysopolis eklendi")

# ─── Komsu guncelleme ────────────────────────────────────────────────────────

# bithynia: chrysopolis ekle (nicomedia/paphlagonia zaten var)
add_nb('bithynia', 'chrysopolis')

# nicomedia: chrysopolis ekle, paphlagonia KALDIR
#   (chrysopolis araya girdi: Nicomedia-Paphlagonia midpoint'i chrysopolis kazaniyor)
add_nb('nicomedia', 'chrysopolis')
remove_nb('nicomedia', 'paphlagonia')
remove_nb('paphlagonia', 'nicomedia')  # karsilikli kaldir

# paphlagonia: chrysopolis ekle
add_nb('paphlagonia', 'chrysopolis')

# constantinople: chrysopolis ekle
#   (cografi: Hrisopolis tam karsisinda)
add_nb('constantinople', 'chrysopolis')

# ─── Dogrulama ───────────────────────────────────────────────────────────────
from collections import Counter
id_counts = Counter(r['id'] for r in regions)
dupes = {k: v for k, v in id_counts.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

# Kuzey-guney siralama dogrulamasi
tur = [r for r in regions if r.get('shape_id') == 'TUR' and not r.get('is_sea')]
print("\nTUR polygon (kuzey->guney):")
for r in sorted(tur, key=lambda x: (x['world_y'], x['world_x'])):
    print(f"  {r['id']:15} wx={r['world_x']:4} wy={r['world_y']:4} owner={r.get('owner_id','')[:15]}")

# ─── Kaydet ──────────────────────────────────────────────────────────────────
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
print(f"\nKaydedildi. Toplam: {len(regions)} bolge")
