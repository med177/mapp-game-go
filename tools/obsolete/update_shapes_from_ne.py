"""
Natural Earth shapefile'dan SVK, EST, LVA, BLR, NOR, SWE, FIN poligonlarını okur,
piksel uzayına dönüştürür, sadeleştirir ve country_shapes.json'ı günceller.

Kullanım: python tools/update_shapes_from_ne.py
"""

import shapefile
import json
import math
import sys

SHP_PATH  = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"
SHAPES_PATH = "assets/data/generated/country_shapes.json"

TARGET = {"SVK", "EST", "LVA", "BLR", "NOR", "SWE", "FIN"}

# Projeksiyon: 1920×1080 piksel uzayı
def lon_to_px(lon): return 20 * lon + 476
def lat_to_py(lat): return -18.98 * lat + 1234.8

MAP_W, MAP_H = 1920, 1080


def simplify_dp(pts, tolerance):
    """Douglas-Peucker çizgi sadeleştirme."""
    if len(pts) <= 2:
        return pts
    d_max, idx = 0, 0
    end = len(pts) - 1
    x0, y0 = pts[0]
    x1, y1 = pts[end]
    dx, dy = x1 - x0, y1 - y0
    norm = math.sqrt(dx * dx + dy * dy)
    for i in range(1, end):
        if norm == 0:
            d = math.sqrt((pts[i][0] - x0) ** 2 + (pts[i][1] - y0) ** 2)
        else:
            d = abs(dy * pts[i][0] - dx * pts[i][1] + x1 * y0 - y1 * x0) / norm
        if d > d_max:
            d_max, idx = d, i
    if d_max > tolerance:
        left  = simplify_dp(pts[:idx + 1], tolerance)
        right = simplify_dp(pts[idx:], tolerance)
        return left[:-1] + right
    return [pts[0], pts[end]]


def extract_shapes(shp_path, target_isos, tolerance_map):
    sf = shapefile.Reader(shp_path)
    fields = [f[0] for f in sf.fields[1:]]
    results = {}

    for sr in sf.shapeRecords():
        d = dict(zip(fields, sr.record))
        iso = d.get("ADM0_A3")
        if iso not in target_isos:
            continue

        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]

        # Tüm parçaları piksel uzayına çevir, MAP içinde olanları topla
        all_rings = []
        for i in range(len(parts) - 1):
            raw = shape.points[parts[i]:parts[i + 1]]
            px_ring = []
            for lon, lat in raw:
                px = lon_to_px(lon)
                py = lat_to_py(lat)
                # Harita sınırları dışındaki noktaları kırp (y<0 = harita üstü)
                px_ring.append((px, py))
            all_rings.append(px_ring)

        # En büyük parçayı al (adaları atla — kıta gövdesi)
        all_rings.sort(key=lambda r: _ring_area(r), reverse=True)
        main_ring = all_rings[0]

        tol = tolerance_map.get(iso, 2.0)
        simplified = simplify_dp(main_ring, tol)

        # Tamsayıya yuvarla, harita sınırlarına kırp
        int_ring = []
        for x, y in simplified:
            cx = max(0, min(MAP_W - 1, round(x)))
            cy = max(0, min(MAP_H - 1, round(y)))
            int_ring.append([cx, cy])

        results[iso] = int_ring
        print(f"  {iso}: {len(main_ring)} ham nokta -> {len(simplified)} sadelesstirilmis")

    return results


def _ring_area(pts):
    """Shoelace formülüyle poligon alanı (işaretsiz)."""
    n = len(pts)
    area = 0.0
    for i in range(n):
        j = (i + 1) % n
        area += pts[i][0] * pts[j][1]
        area -= pts[j][0] * pts[i][1]
    return abs(area) / 2.0


def main():
    # Her ülke için farklı sadeleştirme toleransı (büyük kıyı şeridine sahip ülkeler daha yüksek)
    tolerance_map = {
        "SVK": 1.5,
        "EST": 1.5,
        "LVA": 1.5,
        "BLR": 1.5,
        "NOR": 6.0,   # Fjordlu kıyı — daha agresif sadeleştirme
        "SWE": 4.0,
        "FIN": 4.0,
    }

    print("Shapefile okunuyor…")
    new_rings = extract_shapes(SHP_PATH, TARGET, tolerance_map)

    if len(new_rings) != len(TARGET):
        missing = TARGET - set(new_rings.keys())
        print(f"HATA: Şu ülkeler bulunamadı: {missing}", file=sys.stderr)
        sys.exit(1)

    # country_shapes.json güncelle
    with open(SHAPES_PATH, "r", encoding="utf-8") as f:
        shape_file = json.load(f)

    updated = 0
    added   = 0
    existing_ids = {s["id"]: i for i, s in enumerate(shape_file["shapes"])}

    for iso, ring in new_rings.items():
        entry = {"id": iso, "name": iso, "rings": [ring]}
        if iso in existing_ids:
            shape_file["shapes"][existing_ids[iso]] = entry
            updated += 1
            print(f"  Güncellendi : {iso}  ({len(ring)} nokta)")
        else:
            shape_file["shapes"].append(entry)
            added += 1
            print(f"  Eklendi     : {iso}  ({len(ring)} nokta)")

    with open(SHAPES_PATH, "w", encoding="utf-8") as f:
        json.dump(shape_file, f, indent=2)

    print(f"\n{updated} sekil guncellendi, {added} sekil eklendi -> {SHAPES_PATH}")


if __name__ == "__main__":
    main()
