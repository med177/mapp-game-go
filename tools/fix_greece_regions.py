"""
Yunanistan (GRC) poligonu Voronoi duzeltmesi.

Sorun:
  greece (948,492) GRC'nin en dogu merkezi => Chalcidice + Attika
  yarimalari tum dar bantlariyla 'greece' hucresine dusuyor.

Cozum:
  1. thessaly: (928,470) -> (920,462)  [daha kuzey = Selanik/Larissa]
  2. greece:   (948,492) -> (962,503)  [daha dogu+guney = Attika/Atina]
  3. Yeni: chalcidice (968,460) Bizans  [kuzeyden 'greece' yi keser]

Voronoi kontrol:
  chalcidice(968,460) - greece(962,503) midpoint (965,481.5):
    d2(C)=6.25+462.25=468.5  d2(G)=9+462.25=471.25  => C kazanir (araya giriyor)
  thessaly(920,462) - greece(962,503) midpoint (941,482.5):
    d2(T)=21^2+20.5^2=820+420=1240  d2(G)=21^2+20.5^2=1240
    d2(C)=27^2+22.5^2=729+506=1235  => C kazanir (yaklik ciddi)
    Pratik: C sadece 5 birim avantajli -> cok kucuk etki.
    Dolayisiyla thessaly-greece bant arasi chalcidice ile dolar ama
    thessaly-greece komsulugu da korunabilir.
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

def get_r(rid):
    return next((r for r in regions if r['id'] == rid), None)

def add_nb(rid, nb):
    r = get_r(rid)
    if r and nb not in r.get('neighbors', []):
        r['neighbors'].append(nb)

def remove_nb(rid, nb):
    r = get_r(rid)
    if r and nb in r.get('neighbors', []):
        r['neighbors'].remove(nb)

# ════════════════════════════════════════════════════════════════
# 1. MEVCUT MERKEZLERI YENIDEN KONUMLANDIR
# ════════════════════════════════════════════════════════════════

# thessaly: Selanik/Larissa = daha kuzey, biraz bati
t = get_r('thessaly')
if t:
    t['world_x'] = 920
    t['world_y'] = 462
    print(f"thessaly: ({t['world_x']},{t['world_y']}) guncellendi")

# greece (Atina Dukali): Attika yarimadasi = daha dogu+guney
g = get_r('greece')
if g:
    g['world_x'] = 962
    g['world_y'] = 503
    print(f"greece: ({g['world_x']},{g['world_y']}) guncellendi")

# ════════════════════════════════════════════════════════════════
# 2. CHALCIDICE: YENI BOLGE
# ════════════════════════════════════════════════════════════════
# Chalcidice (Halkidiki) yarimadasi: Bizans monastic cumhuriyeti (Mount Athos).
# (968,460) konumu:
#   - thessaly(920,462) kuzeyindeki dogu uzantisini alir
#   - greece(962,503) ile araya girerek dar banti keser
#   - thrace(TRA,1022,451) ile kesisimde cross-polygon komsusu
existing_ids = {r['id'] for r in regions}
if 'chalcidice' not in existing_ids:
    regions.append({
        "active_event_id": "",
        "base_gold_income": 30,
        "base_grain_output": 35,
        "id": "chalcidice",
        "is_locked": False,
        "is_sea": False,
        "name": "Chalcidice",
        "name_tr": "Halkidiki (Kutsal Dag / Bizans)",
        "neighbors": ["thessaly", "greece", "thrace", "macedonia", "_sea_aegean"],
        "owner_id": "byzantine",
        "population": 180,
        "religion": "orthodox",
        "satisfaction": 72,
        "shape_id": "GRC",
        "tax_rate": 20,
        "terrain": "coast",
        "trade_capacity": 2,
        "unlock_turn": 0,
        "world_x": 968,
        "world_y": 460
    })
    print("chalcidice eklendi")
else:
    print("chalcidice zaten var")

# ════════════════════════════════════════════════════════════════
# 3. KOMSU LISTELERI
# ════════════════════════════════════════════════════════════════

# thessaly: chalcidice ekle
add_nb('thessaly', 'chalcidice')

# greece: chalcidice ekle
#   thrace KALDIR (chalcidice araya girdi: thrace artik chalcidice uzerinden)
#   epirus KALDIR (thessaly araya giriyor)
add_nb('greece', 'chalcidice')
remove_nb('greece', 'thrace')
remove_nb('greece', 'epirus')

# thrace: chalcidice ekle
add_nb('thrace', 'chalcidice')

# macedonia: chalcidice ekle (cografi komsu, farkli polygon)
add_nb('macedonia', 'chalcidice')

# epirus: greece KALDIR (thessaly araya giriyor artik)
remove_nb('epirus', 'greece')

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

grc = sorted([r for r in regions if r.get('shape_id') == 'GRC'],
             key=lambda x: (x['world_y'], x['world_x']))
print("\nGRC poligonu (kuzey->guney):")
for r in grc:
    print(f"  {r['id']:12} wx={r['world_x']:4} wy={r['world_y']:4} "
          f"owner={r.get('owner_id','')[:15]:15} "
          f"nb={r.get('neighbors',[])}")

# ════════════════════════════════════════════════════════════════
# 5. KAYDET
# ════════════════════════════════════════════════════════════════
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
print(f"\nKaydedildi. Toplam: {len(regions)} bolge")
