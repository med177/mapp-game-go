"""Harita denetimi: tum bolge/fraksiyon durumunu raporla."""
import json
from collections import Counter, defaultdict

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)
with open('assets/data/factions.json', encoding='utf-8') as f:
    factions = json.load(f)

faction_ids = {f['id'] for f in factions}
land = [r for r in regions if not r.get('is_sea') and r['id'] != '' and not r['id'].startswith('_')]

# ── 1. Shape gruplarini listele ───────────────────────────────────────────────
print("=" * 70)
print("SHAPE BAZLI BOLGE DAGILIMI (kuzey->guney siralama)")
print("=" * 70)
by_shape = defaultdict(list)
for r in land:
    by_shape[r.get('shape_id', 'NONE')].append(r)

for shape in sorted(by_shape.keys()):
    regs = sorted(by_shape[shape], key=lambda x: (x['world_y'], x['world_x']))
    print(f"\n{shape} ({len(regs)} bolge):")
    for r in regs:
        owner = r.get('owner_id', '')
        sea_nb = [n for n in r.get('neighbors', []) if n.startswith('_sea')]
        sea_str = ' SEA:' + '+'.join(n.replace('_sea_','') for n in sea_nb) if sea_nb else ''
        print(f"  {r['id']:20} wx={r['world_x']:4} wy={r['world_y']:4}  {owner:20}{sea_str}")

# ── 2. Sahipsiz kara bolgeleri ─────────────────────────────────────────────────
print("\n" + "=" * 70)
neutral = [r for r in land if not r.get('owner_id')]
print(f"SAHIPSIZ KARA BOLGELERI ({len(neutral)}):")
for r in sorted(neutral, key=lambda x: x.get('shape_id','')):
    print(f"  {r['id']:22} shape={r.get('shape_id',''):6} wx={r['world_x']:4} wy={r['world_y']:4}")

# ── 3. Gecersiz owner ─────────────────────────────────────────────────────────
print("\n" + "=" * 70)
bad_owners = [(r['id'], r['owner_id']) for r in land if r.get('owner_id') and r['owner_id'] not in faction_ids]
if bad_owners:
    print(f"GECERSIZ OWNER ({len(bad_owners)}):")
    for rid, oid in bad_owners:
        print(f"  {rid}: {oid}")
else:
    print("TUM OWNER'LAR GECERLI")

# ── 4. Buyuk bosluklar olan shape'ler (tek bolge) ────────────────────────────
print("\n" + "=" * 70)
print("TEK BOLGELI SHAPE'LER (eksik bolge adayi):")
for shape, regs in sorted(by_shape.items()):
    if len(regs) == 1:
        r = regs[0]
        print(f"  {shape}: {r['id']} ({r.get('owner_id','neutral')})")

print(f"\nToplam: {len(land)} kara, {len(factions)} fraksiyon")
