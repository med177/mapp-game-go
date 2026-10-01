"""
1. DNK seklini gunceller: Funen (ring[8]) ucuncu ring olarak eklenir
   (Jutland ve Zelanda arasindaki bosluk dolar)
2. Balear Adalari (MAL=Mallorca, MEN=Menorca, IBZ=Ibiza) sekilleri eklenir
3. balearics bolgesi regions.json'a eklenir
Kullanim: python tools/fix_denmark_add_balearics.py
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


def get_rings_for_iso(iso):
    sf = shapefile.Reader(SHP_PATH)
    fields = [f[0] for f in sf.fields[1:]]
    for sr in sf.shapeRecords():
        if dict(zip(fields, sr.record)).get("ADM0_A3") != iso:
            continue
        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]
        return [shape.points[parts[i]:parts[i+1]]
                for i in range(len(parts)-1)]
    return []


def to_px_ring(lonlat_ring, tol):
    px = [(lon_to_px(lon), lat_to_py(lat)) for lon, lat in lonlat_ring]
    simp = simplify_dp(px, tol)
    return [[max(0, min(MAP_W-1, round(x))), max(0, min(MAP_H-1, round(y)))]
            for x, y in simp]


def centroid(px_ring):
    cx = sum(p[0] for p in px_ring) / len(px_ring)
    cy = sum(p[1] for p in px_ring) / len(px_ring)
    return round(cx), round(cy)


with open(SHAPES_PATH, "r", encoding="utf-8") as f:
    shape_file = json.load(f)
by_id = {s["id"]: i for i, s in enumerate(shape_file["shapes"])}

# ── 1. DNK: Funen (ring[8]) ekle ────────────────────────────────────────────
print("DNK — Funen (ring[8]) ekleniyor:")
dnk_rings = get_rings_for_iso("DNK")
funen_raw = dnk_rings[8]  # centroid (10.30, 55.35) = Funen
funen_px = to_px_ring(funen_raw, 1.5)
print(f"  Funen: {len(funen_raw)} -> {len(funen_px)} pt  merkez={centroid(funen_px)}")

if "DNK" in by_id:
    dnk_shape = shape_file["shapes"][by_id["DNK"]]
    # Daha once 2 ring varsa (Jutland + Zelanda), Funen'i ucuncu olarak ekle
    while len(dnk_shape["rings"]) < 2:
        dnk_shape["rings"].append([])  # placeholder
    if len(dnk_shape["rings"]) == 2:
        dnk_shape["rings"].append(funen_px)
        print(f"  DNK rings: {len(dnk_shape['rings'])} (Jutland + Zelanda + Funen)")
    else:
        # Guncelle: 3. ring'i degistir
        dnk_shape["rings"][2] = funen_px
        print(f"  DNK ring[2] guncellendi (Funen)")

# ── 2. Balear Adalari ─────────────────────────────────────────────────────────
print("\nBalear Adalari sekilleri:")
esp_rings = get_rings_for_iso("ESP")

# Mallorca: ring[15], Menorca: ring[16], Ibiza: ring[14]
island_targets = [
    ("MAL", "Mallorca", 15, 0.8),
    ("MEN", "Menorca",  16, 0.8),
    ("IBZ", "Ibiza",    14, 0.8),
]

bal_rings = []  # BAL shape icin tum adalar
for shape_id, name, ring_idx, tol in island_targets:
    raw = esp_rings[ring_idx]
    px = to_px_ring(raw, tol)
    cx, cy = centroid(px)
    print(f"  {shape_id} {name}: {len(raw)} -> {len(px)} pt  merkez=({cx},{cy})")
    bal_rings.append(px)
    # Her adayi ayri sekil olarak kaydet (Voronoi icin)
    entry = {"id": shape_id, "name": name, "rings": [px]}
    if shape_id in by_id:
        shape_file["shapes"][by_id[shape_id]] = entry
        print(f"    Guncellendi: {shape_id}")
    else:
        shape_file["shapes"].append(entry)
        by_id[shape_id] = len(shape_file["shapes"]) - 1
        print(f"    Eklendi   : {shape_id}")

shape_file["shapes"].sort(key=lambda s: s["id"])
with open(SHAPES_PATH, "w", encoding="utf-8") as f:
    json.dump(shape_file, f, indent=2, ensure_ascii=False)
print(f"\nToplam {len(shape_file['shapes'])} sekil -> {SHAPES_PATH}")

# ── 3. balearics bolgesi ─────────────────────────────────────────────────────
# Mallorca centroidu ana konum
mal_cx, mal_cy = centroid(to_px_ring(esp_rings[15], 0.8))

with open(REGIONS_PATH, "r", encoding="utf-8") as f:
    regions = json.load(f)
by_rid = {r["id"]: i for i, r in enumerate(regions)}

NEW_REGIONS = [
    {
        "id": "mallorca", "name": "Mallorca", "name_tr": "Mallorca",
        "shape_id": "MAL", "terrain": "coast", "owner_id": "",
        "neighbors": ["aragon", "_sea_med_west"],
        "world_x": mal_cx, "world_y": mal_cy,
        "is_sea": False, "is_locked": False, "unlock_turn": 0,
        "base_gold_income": 28, "base_grain_output": 22, "trade_capacity": 4,
        "satisfaction": 65, "tax_rate": 30, "population": 300,
        "religion": "catholic", "active_event_id": "",
    },
]

NEIGHBOR_ADD = {
    "aragon": ["mallorca"],
}

added = 0
for nr in NEW_REGIONS:
    if nr["id"] in by_rid:
        print(f"  ATLANDI (zaten var): {nr['id']}")
        continue
    regions.append(nr)
    by_rid[nr["id"]] = len(regions) - 1
    added += 1
    print(f"\nBolge eklendi: {nr['id']} wx={nr['world_x']} wy={nr['world_y']}")

nb_updated = 0
for rid, adds in NEIGHBOR_ADD.items():
    if rid not in by_rid:
        print(f"  UYARI: {rid} bulunamadi")
        continue
    nb = regions[by_rid[rid]].get("neighbors", [])
    changed = any(a not in nb for a in adds)
    for a in adds:
        if a not in nb:
            nb.append(a)
    if changed:
        regions[by_rid[rid]]["neighbors"] = nb
        nb_updated += 1
        print(f"Komsu guncellendi: {rid}")

with open(REGIONS_PATH, "w", encoding="utf-8") as f:
    json.dump(regions, f, indent=2, ensure_ascii=False)
print(f"\n{added} bolge eklendi, {nb_updated} komsu guncellendi. Toplam: {len(regions)}")
