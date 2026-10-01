"""
KON (Kaliningrad/Konigsberg), CRM (Kirim), TRA (Trakya) icin GERCEK NE verisi kullanir:
- KON: RUS poligonundaki Kaliningrad ring'i (ayri exclave)
- CRM: UKR ana ring'inin Kirim bbox'iyla Sutherland-Hodgman kırpması
- TRA: TUR ana ring'inin Trakya bbox'iyla Sutherland-Hodgman kırpması

Kullanim: python tools/fix_real_shapes.py
"""

import shapefile, json, math, sys

SHP_PATH    = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"
SHAPES_PATH = "assets/data/generated/country_shapes.json"

MAP_W, MAP_H = 1828, 997
def lon_to_px(lon): return 20.0 * lon + 476.0
def lat_to_py(lat): return -18.98 * lat + 1234.8


# ── Douglas-Peucker sadeleştirme ─────────────────────────────────────────────
def simplify_dp(pts, tol):
    if len(pts) <= 2:
        return list(pts)
    d_max = idx = 0
    end = len(pts) - 1
    x0, y0 = pts[0];  x1, y1 = pts[end]
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


# ── Sutherland-Hodgman bbox kırpma (lon/lat uzayında) ───────────────────────
def clip_polygon_bbox(poly, lon_min, lat_min, lon_max, lat_max):
    """Sutherland-Hodgman ile poligonu eksene paralel bbox ile kirpar."""
    def clip_edge(pts, inside, intersect):
        out = []
        if not pts:
            return out
        prev = pts[-1]
        prev_in = inside(prev)
        for curr in pts:
            curr_in = inside(curr)
            if curr_in:
                if not prev_in:
                    out.append(intersect(prev, curr))
                out.append(curr)
            elif prev_in:
                out.append(intersect(prev, curr))
            prev, prev_in = curr, curr_in
        return out

    def lerp(a, b, t):
        return (a[0] + t*(b[0]-a[0]), a[1] + t*(b[1]-a[1]))

    def ix_left(a, b):
        t = (lon_min - a[0]) / (b[0] - a[0]) if b[0] != a[0] else 0
        return lerp(a, b, t)
    def ix_right(a, b):
        t = (lon_max - a[0]) / (b[0] - a[0]) if b[0] != a[0] else 0
        return lerp(a, b, t)
    def ix_bottom(a, b):
        t = (lat_min - a[1]) / (b[1] - a[1]) if b[1] != a[1] else 0
        return lerp(a, b, t)
    def ix_top(a, b):
        t = (lat_max - a[1]) / (b[1] - a[1]) if b[1] != a[1] else 0
        return lerp(a, b, t)

    p = list(poly)
    p = clip_edge(p, lambda pt: pt[0] >= lon_min, ix_left)
    p = clip_edge(p, lambda pt: pt[0] <= lon_max, ix_right)
    p = clip_edge(p, lambda pt: pt[1] >= lat_min, ix_bottom)
    p = clip_edge(p, lambda pt: pt[1] <= lat_max, ix_top)
    return p


# ── NE'den ring çekme yardımcısı ─────────────────────────────────────────────
def get_rings(iso):
    sf = shapefile.Reader(SHP_PATH)
    fields = [f[0] for f in sf.fields[1:]]
    for sr in sf.shapeRecords():
        if dict(zip(fields, sr.record)).get("ADM0_A3") != iso:
            continue
        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]
        return [shape.points[parts[i]:parts[i+1]]
                for i in range(len(parts)-1) if parts[i+1]-parts[i] >= 3]
    return []


def to_px_ring(lonlat_ring, tol):
    px = [(lon_to_px(lon), lat_to_py(lat)) for lon, lat in lonlat_ring]
    simp = simplify_dp(px, tol)
    return [[max(0, min(MAP_W-1, round(x))), max(0, min(MAP_H-1, round(y)))]
            for x, y in simp]


def ring_centroid(px_ring):
    cx = sum(p[0] for p in px_ring) / len(px_ring)
    cy = sum(p[1] for p in px_ring) / len(px_ring)
    return cx, cy


# ─────────────────────────────────────────────────────────────────────────────

def main():
    with open(SHAPES_PATH, "r", encoding="utf-8") as f:
        shape_file = json.load(f)
    by_id = {s["id"]: i for i, s in enumerate(shape_file["shapes"])}

    def upsert(entry):
        if entry["id"] in by_id:
            shape_file["shapes"][by_id[entry["id"]]] = entry
            print(f"  Guncellendi: {entry['id']}")
        else:
            shape_file["shapes"].append(entry)
            by_id[entry["id"]] = len(shape_file["shapes"]) - 1
            print(f"  Eklendi    : {entry['id']}")

    # ── KON: Kaliningrad (RUS ring[1]) ───────────────────────────────────────
    print("KON — Kaliningrad / Konigsberg:")
    rus_rings = get_rings("RUS")
    # Kaliningrad: centroid lon~21.4, lat~54.9 olan ring
    kal_ring = None
    for ring in rus_rings:
        cx = sum(p[0] for p in ring) / len(ring)
        cy = sum(p[1] for p in ring) / len(ring)
        if 19 < cx < 23 and 53 < cy < 56:
            kal_ring = ring
            break
    if kal_ring is None:
        print("  HATA: Kaliningrad ring'i bulunamadi", file=sys.stderr)
    else:
        px_ring = to_px_ring(kal_ring, 1.2)
        cx, cy = ring_centroid(px_ring)
        print(f"  {len(kal_ring)} -> {len(px_ring)} pt  merkez=({cx:.0f},{cy:.0f})")
        upsert({"id": "KON", "name": "Konigsberg", "rings": [px_ring]})

    # ── CRM: Kırım yarımadası — RUS ring[2] (tam yarımada poligonu) ─────────
    # NE admin_0'da Kırım, RUS kaydında ring[2] olarak saklanıyor (lat 44.38-46.22)
    # UKR ring[0] sadece kuzey kıyısı/isthmusu içeriyor.
    print("\nCRM — Kirim Yarimadasi (RUS ring[2]):")
    rus_all_rings = get_rings("RUS")
    # Kırım yarımadası: centroid lon~34°E, lat~45°N olan ring
    crm_ring = None
    for ring in rus_all_rings:
        lons = [p[0] for p in ring]
        lats = [p[1] for p in ring]
        cx = sum(lons) / len(lons)
        cy = sum(lats) / len(lats)
        # bbox içinde yoğunluk kontrolü: lon 32-37, lat 44-47
        in_box = sum(1 for p in ring if 32 <= p[0] <= 37 and 44 <= p[1] <= 47)
        min_lat = min(lats)
        # Kaliningrad'ı dışla (lon ~21), Kerc çıkıntısı ring[0]'ı dışla (çok büyük)
        # Kırım ringi: çoğu nokta bbox içinde, min lat < 45.5, boyut 200-600 arası
        if in_box >= 200 and min_lat < 45.5 and 100 < len(ring) < 800:
            crm_ring = ring
            break
    if crm_ring is None:
        print("  HATA: Kirim ringi RUS'ta bulunamadi", file=sys.stderr)
    else:
        px_ring = to_px_ring(crm_ring, 1.0)
        cx, cy = ring_centroid(px_ring)
        print(f"  {len(crm_ring)} -> {len(px_ring)} pt  merkez=({cx:.0f},{cy:.0f})")
        upsert({"id": "CRM", "name": "Crimea", "rings": [px_ring]})

    # ── TRA: Trakya — TUR ring[1] (Avrupa Türkiyesi, lat 42.1°N'e kadar) ────
    # ring[0] sadece lat 41.24°N'e kadar çıkıyor; ring[1] lat 42.10°N'e kadar.
    print("\nTRA — Trakya (Avrupa Turkiyesi, TUR ring[1]):")
    tur_rings = get_rings("TUR")
    if len(tur_rings) < 2:
        print("  HATA: TUR ring[1] bulunamadi", file=sys.stderr)
    else:
        # ring[1] Trakya bölgesi için en kapsamlı ring
        # Birden fazla ring varsa Trakya bbox'ında en fazla nokta olan ring[1]'i seç
        thrace_bbox = (25.0, 40.0, 30.5, 42.5)
        best_ring = None
        best_count = 0
        for ring in tur_rings:
            cnt = sum(1 for p in ring
                      if thrace_bbox[0] <= p[0] <= thrace_bbox[2]
                      and thrace_bbox[1] <= p[1] <= thrace_bbox[3])
            # Anadolu ana gövdesini dışla (çok büyük ring)
            if cnt > best_count and len(ring) < 2000:
                best_count = cnt
                best_ring = ring
        if best_ring is None or best_count < 10:
            print("  HATA: uygun Trakya ringi bulunamadi", file=sys.stderr)
        else:
            tra_clipped = clip_polygon_bbox(best_ring, 25.0, 40.0, 30.5, 42.5)
            if len(tra_clipped) < 3:
                print("  HATA: kirpma sonucu bos", file=sys.stderr)
            else:
                px_ring = to_px_ring(tra_clipped, 0.8)
                cx, cy = ring_centroid(px_ring)
                print(f"  {len(best_ring)} -> kirpma {len(tra_clipped)} -> {len(px_ring)} pt  merkez=({cx:.0f},{cy:.0f})")
                upsert({"id": "TRA", "name": "Thrace", "rings": [px_ring]})

    shape_file["shapes"].sort(key=lambda s: s["id"])
    with open(SHAPES_PATH, "w", encoding="utf-8") as f:
        json.dump(shape_file, f, indent=2, ensure_ascii=False)
    print(f"\nToplam {len(shape_file['shapes'])} sekil -> {SHAPES_PATH}")


if __name__ == "__main__":
    main()
