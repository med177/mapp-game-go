"""
Natural Earth shapefile'dan ada poligonlarini cikartir ve manuel sekiller uretir.
country_shapes.json'a ekler/gunceller.

Kullanim: python tools/extract_islands.py
"""

import shapefile, json, math, sys

SHP_PATH    = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"
SHAPES_PATH = "assets/data/generated/country_shapes.json"

MAP_W, MAP_H = 1828, 997
def lon_to_px(lon): return 20.0 * lon + 476.0
def lat_to_py(lat): return -18.98 * lat + 1234.8


def simplify_dp(pts, tol):
    if len(pts) <= 2: return list(pts)
    d_max = idx = 0
    end = len(pts) - 1
    x0, y0 = pts[0]; x1, y1 = pts[end]
    dx, dy = x1-x0, y1-y0
    norm = math.sqrt(dx*dx + dy*dy)
    for i in range(1, end):
        d = (math.sqrt((pts[i][0]-x0)**2+(pts[i][1]-y0)**2) if norm == 0
             else abs(dy*pts[i][0]-dx*pts[i][1]+x1*y0-y1*x0)/norm)
        if d > d_max: d_max, idx = d, i
    if d_max > tol:
        return simplify_dp(pts[:idx+1], tol)[:-1] + simplify_dp(pts[idx:], tol)
    return [pts[0], pts[end]]


def ring_area(pts):
    n = len(pts); a = 0.0
    for i in range(n):
        j = (i+1) % n
        a += pts[i][0]*pts[j][1] - pts[j][0]*pts[i][1]
    return abs(a) / 2.0


def all_rings_for_iso(iso):
    sf = shapefile.Reader(SHP_PATH)
    fields = [f[0] for f in sf.fields[1:]]
    for sr in sf.shapeRecords():
        d = dict(zip(fields, sr.record))
        if d.get("ADM0_A3") != iso: continue
        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]
        rings = [shape.points[parts[i]:parts[i+1]]
                 for i in range(len(parts)-1) if parts[i+1]-parts[i] >= 4]
        return rings
    return []


def find_ring_near(rings, target_lon, target_lat, min_px_area=200):
    best, best_dist = None, float("inf")
    for ring in rings:
        cx = sum(p[0] for p in ring) / len(ring)
        cy = sum(p[1] for p in ring) / len(ring)
        dist = (cx-target_lon)**2 + (cy-target_lat)**2
        if dist < best_dist:
            px_ring = [(lon_to_px(p[0]), lat_to_py(p[1])) for p in ring]
            if ring_area(px_ring) < min_px_area: continue
            best_dist = dist; best = ring
    return best


def make_shape(new_id, name, lon_lat_ring, tol):
    px_ring = [(lon_to_px(lon), lat_to_py(lat)) for lon, lat in lon_lat_ring]
    simplified = simplify_dp(px_ring, tol)
    int_ring = [[max(0, min(MAP_W-1, round(x))), max(0, min(MAP_H-1, round(y)))]
                for x, y in simplified]
    if len(int_ring) < 3: return None
    cx = sum(p[0] for p in int_ring) / len(int_ring)
    cy = sum(p[1] for p in int_ring) / len(int_ring)
    print(f"  {new_id:6s}  {name:18s}  {len(lon_lat_ring):5d} -> {len(int_ring):3d} pt  "
          f"merkez=({cx:.0f},{cy:.0f})")
    return {"id": new_id, "name": name, "rings": [int_ring]}


# ─── Ada hedefleri (kaynak_iso, yeni_id, hedef_lon, hedef_lat, tolerans, isim) ─
ISLAND_TARGETS = [
    ("ITA", "SCL", 14.0,  37.5, 1.2, "Sicily"),
    ("ITA", "SAR",  8.8,  40.3, 1.5, "Sardinia"),
    ("FRA", "COR",  8.9,  42.1, 1.2, "Corsica"),
    ("GRC", "CRT", 24.9,  35.2, 1.2, "Crete"),
]

# ─── Manuel poligonlar (piksel uzayi, [lon/lat yorumu] icin hesaplandi) ────────

# Trakya — Avrupa Turkiyesi (TUR seklinden bagımsız ozel poligon)
TRA_RING = [
    (996, 439), (1008, 435), (1022, 431), (1036, 432),
    (1048, 437), (1060, 447), (1062, 458), (1048, 463),
    (1024, 464), (1004, 462), (996, 456), (994, 448),
    (996, 439),
]

# Kirim Yarimadasi — UKR ana ringi'nden bagımsız
CRM_RING = [
    (1128, 364), (1156, 362), (1186, 366), (1204, 372),
    (1208, 386), (1196, 394), (1168, 397), (1148, 394),
    (1132, 387), (1124, 374), (1128, 364),
]

# Konigsberg / Dogu Prusya — Kaliningrad + Kuzey Polonya
KON_RING = [
    (848, 183), (884, 178), (924, 182), (936, 196),
    (928, 220), (904, 228), (868, 228), (848, 214),
    (844, 198), (848, 183),
]


def main():
    with open(SHAPES_PATH, "r", encoding="utf-8") as f:
        shape_file = json.load(f)

    by_id = {s["id"]: i for i, s in enumerate(shape_file["shapes"])}

    def upsert(entry):
        if entry["id"] in by_id:
            shape_file["shapes"][by_id[entry["id"]]] = entry
            print(f"    Guncellendi: {entry['id']}")
        else:
            shape_file["shapes"].append(entry)
            by_id[entry["id"]] = len(shape_file["shapes"]) - 1
            print(f"    Eklendi    : {entry['id']}")

    # ── Ada sekilleri ──────────────────────────────────────────────
    print("Ada sekilleri NE'den cikartiliyor:")
    for iso, new_id, tlon, tlat, tol, name in ISLAND_TARGETS:
        rings = all_rings_for_iso(iso)
        if not rings:
            print(f"  HATA: {iso} bulunamadi", file=sys.stderr); continue
        ring = find_ring_near(rings, tlon, tlat)
        if ring is None:
            print(f"  HATA: {new_id} ringi bulunamadi", file=sys.stderr); continue
        entry = make_shape(new_id, name, ring, tol)
        if entry: upsert(entry)

    # ── Danimarka — Zealand ikincil halka ────────────────────────
    print("\nDanimarka Sjaelland (Zealand) ikincil halka ekleniyor:")
    dnk_rings = all_rings_for_iso("DNK")
    zld_ring = find_ring_near(dnk_rings, 11.8, 55.6)
    if zld_ring:
        px = [(lon_to_px(p[0]), lat_to_py(p[1])) for p in zld_ring]
        simplified = simplify_dp(px, 1.5)
        int_ring = [[max(0, min(MAP_W-1, round(x))), max(0, min(MAP_H-1, round(y)))]
                    for x, y in simplified]
        if "DNK" in by_id:
            dnk = shape_file["shapes"][by_id["DNK"]]
            if len(dnk["rings"]) < 2:
                dnk["rings"].append(int_ring)
                print(f"  DNK Sjaelland: {len(zld_ring)} -> {len(int_ring)} pt eklendi")

    # ── Manuel poligonlar ────────────────────────────────────────
    print("\nManuel poligonlar ekleniyor:")

    def manual(new_id, name, ring_px):
        int_ring = [[max(0, min(MAP_W-1, x)), max(0, min(MAP_H-1, y))] for x, y in ring_px]
        cx = sum(p[0] for p in int_ring) / len(int_ring)
        cy = sum(p[1] for p in int_ring) / len(int_ring)
        print(f"  {new_id:6s}  {name:18s}  {len(int_ring)} pt  merkez=({cx:.0f},{cy:.0f})")
        return {"id": new_id, "name": name, "rings": [int_ring]}

    upsert(manual("TRA", "Thrace",  TRA_RING))
    upsert(manual("CRM", "Crimea",  CRM_RING))
    upsert(manual("KON", "Konigsberg", KON_RING))

    shape_file["shapes"].sort(key=lambda s: s["id"])
    with open(SHAPES_PATH, "w", encoding="utf-8") as f:
        json.dump(shape_file, f, indent=2, ensure_ascii=False)
    print(f"\nToplam {len(shape_file['shapes'])} sekil -> {SHAPES_PATH}")


if __name__ == "__main__":
    main()
