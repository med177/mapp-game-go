"""
Kuzey Kibris (CYN ring[0]) ve Rodos (GRC ring[8]) sekilleri ekler,
ardından regions.json'a ilgili bolgeleri ekler.
Kullanim: python tools/add_north_cyprus_rhodes.py
"""
import shapefile, json, math

SHP_PATH    = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"
SHAPES_PATH = "assets/data/generated/country_shapes.json"
REGIONS_PATH = "assets/data/regions.json"

MAP_W, MAP_H = 1828, 997
def lon_to_px(lon): return 20.0 * lon + 476.0
def lat_to_py(lat): return -18.98 * lat + 1234.8


def simplify_dp(pts, tol):
    if len(pts) <= 2:
        return list(pts)
    d_max = idx = 0
    end = len(pts) - 1
    x0, y0 = pts[0]; x1, y1 = pts[end]
    dx, dy = x1 - x0, y1 - y0
    norm = math.sqrt(dx * dx + dy * dy)
    for i in range(1, end):
        d = (math.sqrt((pts[i][0]-x0)**2 + (pts[i][1]-y0)**2) if norm == 0
             else abs(dy*pts[i][0] - dx*pts[i][1] + x1*y0 - y1*x0) / norm)
        if d > d_max:
            d_max, idx = d, i
    if d_max > tol:
        return simplify_dp(pts[:idx+1], tol)[:-1] + simplify_dp(pts[idx:], tol)
    return [pts[0], pts[end]]


def get_ring(iso, ring_index):
    sf = shapefile.Reader(SHP_PATH)
    fields = [f[0] for f in sf.fields[1:]]
    for sr in sf.shapeRecords():
        if dict(zip(fields, sr.record)).get("ADM0_A3") != iso:
            continue
        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]
        rings = [shape.points[parts[i]:parts[i+1]]
                 for i in range(len(parts)-1)]
        if ring_index < len(rings):
            return rings[ring_index]
    return None


def to_px_ring(lonlat_ring, tol):
    px = [(lon_to_px(lon), lat_to_py(lat)) for lon, lat in lonlat_ring]
    simp = simplify_dp(px, tol)
    return [[max(0, min(MAP_W-1, round(x))), max(0, min(MAP_H-1, round(y)))]
            for x, y in simp]


def ring_centroid(px_ring):
    cx = sum(p[0] for p in px_ring) / len(px_ring)
    cy = sum(p[1] for p in px_ring) / len(px_ring)
    return round(cx), round(cy)


# ── Şekilleri country_shapes.json'a ekle ────────────────────────────────────

with open(SHAPES_PATH, "r", encoding="utf-8") as f:
    shape_file = json.load(f)
by_id = {s["id"]: i for i, s in enumerate(shape_file["shapes"])}

def upsert_shape(entry):
    if entry["id"] in by_id:
        shape_file["shapes"][by_id[entry["id"]]] = entry
        print(f"  Sekil guncellendi: {entry['id']}")
    else:
        shape_file["shapes"].append(entry)
        by_id[entry["id"]] = len(shape_file["shapes"]) - 1
        print(f"  Sekil eklendi   : {entry['id']}")

# Kuzey Kibris: CYN ring[0]
print("Kuzey Kibris (CYN ring[0]):")
kky_ring = get_ring("CYN", 0)
if kky_ring:
    px = to_px_ring(kky_ring, 0.8)
    cx, cy = ring_centroid(px)
    print(f"  {len(kky_ring)} -> {len(px)} pt  merkez=({cx},{cy})")
    upsert_shape({"id": "CYN", "name": "North Cyprus", "rings": [px]})
else:
    print("  HATA: CYN bulunamadi")
    cx, cy = 1149, 560

# Rodos: GRC ring[8]
print("Rodos (GRC ring[8]):")
rho_ring = get_ring("GRC", 8)
if rho_ring:
    px = to_px_ring(rho_ring, 0.8)
    cx_r, cy_r = ring_centroid(px)
    print(f"  {len(rho_ring)} -> {len(px)} pt  merkez=({cx_r},{cy_r})")
    upsert_shape({"id": "RHO", "name": "Rhodes", "rings": [px]})
else:
    print("  HATA: GRC ring[8] bulunamadi")
    cx_r, cy_r = 1035, 548

shape_file["shapes"].sort(key=lambda s: s["id"])
with open(SHAPES_PATH, "w", encoding="utf-8") as f:
    json.dump(shape_file, f, indent=2, ensure_ascii=False)
print(f"Toplam {len(shape_file['shapes'])} sekil -> {SHAPES_PATH}\n")

# ── Bölgeleri regions.json'a ekle ────────────────────────────────────────────

with open(REGIONS_PATH, "r", encoding="utf-8") as f:
    regions = json.load(f)
by_rid = {r["id"]: i for i, r in enumerate(regions)}

# Kuzey Kibris centroid (CYN ring[0])
kky_cx, kky_cy = ring_centroid(to_px_ring(get_ring("CYN", 0), 0.8)) if get_ring("CYN", 0) else (1149, 560)
rho_cx, rho_cy = ring_centroid(to_px_ring(get_ring("GRC", 8), 0.8)) if get_ring("GRC", 8) else (1035, 548)

NEW_REGIONS = [
    {
        "id": "north_cyprus", "name": "North Cyprus", "name_tr": "Kuzey Kıbrıs",
        "shape_id": "CYN", "terrain": "coast", "owner_id": "",
        "neighbors": ["cyprus", "_sea_med_east"],
        "world_x": kky_cx, "world_y": kky_cy,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 20, "base_grain_output": 15, "trade_capacity": 3,
        "satisfaction": 60, "tax_rate": 30, "population": 200,
        "religion": "sunni", "active_event_id": "",
    },
    {
        "id": "rhodes", "name": "Rhodes", "name_tr": "Rodos",
        "shape_id": "RHO", "terrain": "coast", "owner_id": "",
        "neighbors": ["_sea_aegean", "_sea_med_east"],
        "world_x": rho_cx, "world_y": rho_cy,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 28, "base_grain_output": 18, "trade_capacity": 4,
        "satisfaction": 65, "tax_rate": 30, "population": 250,
        "religion": "catholic", "active_event_id": "",
    },
]

NEIGHBOR_ADD = {
    "cyprus": ["north_cyprus"],
}

added = 0
for nr in NEW_REGIONS:
    if nr["id"] in by_rid:
        print(f"  ATLANDI (zaten var): {nr['id']}")
        continue
    regions.append(nr)
    by_rid[nr["id"]] = len(regions) - 1
    added += 1
    print(f"  Bolge eklendi: {nr['id']} ({nr['name_tr']}) wx={nr['world_x']} wy={nr['world_y']}")

nb_updated = 0
for rid, adds in NEIGHBOR_ADD.items():
    if rid not in by_rid:
        continue
    nb = regions[by_rid[rid]].get("neighbors", [])
    changed = any(a not in nb for a in adds)
    for a in adds:
        if a not in nb:
            nb.append(a)
    if changed:
        regions[by_rid[rid]]["neighbors"] = nb
        nb_updated += 1
        print(f"  Komsu guncellendi: {rid}")

with open(REGIONS_PATH, "w", encoding="utf-8") as f:
    json.dump(regions, f, indent=2, ensure_ascii=False)

print(f"\n{added} bolge eklendi, {nb_updated} komsu guncellendi. Toplam: {len(regions)}")
