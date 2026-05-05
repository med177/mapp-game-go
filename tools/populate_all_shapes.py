"""
Natural Earth shapefile'dan harita kapsamindaki TUM ulkeleri okur,
piksel uzayina donusturur, sadelesstirir ve country_shapes.json'u tamamen doldurur.

Kullanim: python tools/populate_all_shapes.py
"""

import shapefile
import json
import math
import sys
import os

SHP_PATH    = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"
SHAPES_PATH = "assets/data/generated/country_shapes.json"

# Projeksiyon: px = 20*lon + 476, py = -18.98*lat + 1234.8  (1828x997 uzayi)
def lon_to_px(lon): return 20.0 * lon + 476.0
def lat_to_py(lat): return -18.98 * lat + 1234.8

MAP_W, MAP_H = 1828, 997

# Harita bolge siniri (lon/lat cinsinden)
MAP_LON_MIN, MAP_LON_MAX = -15.0, 65.0
MAP_LAT_MIN, MAP_LAT_MAX = 12.0,  70.0

# Natural Earth ISO -> oyun icindeki ISO mapping
# (NE kodlari farkli olan ulkeler icin)
ISO_REMAP = {
    "PSX": "PSE",   # Palestine
    "SDS": "SSD",   # South Sudan
}

# Kiyisi uzun olan ulkeler icin daha agresif sadelesstirime
TOLERANCE_MAP = {
    "NOR": 5.0,
    "SWE": 3.5,
    "FIN": 3.5,
    "GBR": 3.5,
    "GRC": 3.5,
    "IRL": 3.0,
    "RUS": 4.0,
    "TUR": 2.5,
    "SAU": 2.5,
    "KAZ": 3.0,
    "IRN": 2.5,
    "DZA": 2.5,
    "MAR": 2.0,
    "LBY": 2.0,
    "EGY": 2.0,
}
DEFAULT_TOLERANCE = 1.2


# ─── Douglas-Peucker ─────────────────────────────────────────────────────────

def simplify_dp(pts, tolerance):
    if len(pts) <= 2:
        return list(pts)
    d_max, idx = 0.0, 0
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


def ring_area(pts):
    n = len(pts)
    a = 0.0
    for i in range(n):
        j = (i + 1) % n
        a += pts[i][0] * pts[j][1]
        a -= pts[j][0] * pts[i][1]
    return abs(a) / 2.0


# ─── Ana donusum ─────────────────────────────────────────────────────────────

def extract_all_shapes(shp_path):
    sf = shapefile.Reader(shp_path)
    fields = [f[0] for f in sf.fields[1:]]
    results = {}

    for sr in sf.shapeRecords():
        d    = dict(zip(fields, sr.record))
        iso  = d.get("ADM0_A3", "")
        name = d.get("ADMIN", iso)

        # ISO yeniden adlandir
        iso = ISO_REMAP.get(iso, iso)

        # Bbox kontrolu: harita kapsaminda mi?
        b = sr.shape.bbox
        lon_min, lat_min, lon_max, lat_max = b
        if lon_max < MAP_LON_MIN or lon_min > MAP_LON_MAX:
            continue
        if lat_max < MAP_LAT_MIN or lat_min > MAP_LAT_MAX:
            continue

        shape = sr.shape
        parts = list(shape.parts) + [len(shape.points)]

        # Tum parcalari piksel uzayina donustur
        all_rings = []
        for i in range(len(parts) - 1):
            raw = shape.points[parts[i]:parts[i + 1]]
            # Cok kucuk adalari atla (ham nokta < 4)
            if len(raw) < 4:
                continue
            px_ring = [(lon_to_px(lon), lat_to_py(lat)) for lon, lat in raw]
            # Harita icinde en az bir noktasi olmali
            in_map = any(
                0 <= px <= MAP_W and 0 <= py <= MAP_H
                for px, py in px_ring
            )
            if not in_map:
                continue
            all_rings.append(px_ring)

        if not all_rings:
            continue

        # En buyuk parcayi sec (kita govdesi)
        all_rings.sort(key=ring_area, reverse=True)
        main_ring = all_rings[0]

        tol = TOLERANCE_MAP.get(iso, DEFAULT_TOLERANCE)
        simplified = simplify_dp(main_ring, tol)

        # Tamsayiya yuvarla, harita sinirlarindan cok uzaga cikmasina izin ver
        # (poligon sinir noktalarinin biraz disari cikmasi normal)
        int_ring = [[round(x), round(y)] for x, y in simplified]

        # Minimum nokta sayisi kontrolu
        if len(int_ring) < 3:
            continue

        results[iso] = {"name": name, "ring": int_ring}
        print(f"  {iso:6s}  {name[:30]:30s}  {len(main_ring):5d} -> {len(int_ring):3d} pt")

    return results


def main():
    if not os.path.exists(SHP_PATH):
        print(f"HATA: {SHP_PATH} bulunamadi", file=sys.stderr)
        sys.exit(1)

    print("Natural Earth shapefile okunuyor...")
    print(f"  Projeksiyon: px=20*lon+476, py=-18.98*lat+1234.8")
    print(f"  Harita siniri: lon [{MAP_LON_MIN},{MAP_LON_MAX}] lat [{MAP_LAT_MIN},{MAP_LAT_MAX}]")
    print()

    extracted = extract_all_shapes(SHP_PATH)

    print(f"\n{len(extracted)} ulke basariyla islendi.")

    # Mevcut dosyayi oku (korunmasi gereken metadata icin)
    if os.path.exists(SHAPES_PATH):
        with open(SHAPES_PATH, "r", encoding="utf-8") as f:
            existing = json.load(f)
        # Mevcut shape sayisini goster
        print(f"Mevcut dosyada {len(existing['shapes'])} shape vardi.")
    else:
        existing = {"shapes": []}

    # Mevcut shapeleri dict'e al (korunacak olanlar icin)
    old_map = {s["id"]: s for s in existing["shapes"]}

    # Yeni shape listesini olustur:
    # extracted'dan gelenler + extracted'da olmayan mevcut shapeler (korunur)
    new_shapes = []
    updated = added = kept = 0

    # Once extracted'dan gelenleri yaz
    for iso, data in sorted(extracted.items()):
        entry = {
            "id":    iso,
            "name":  data["name"],
            "rings": [data["ring"]]
        }
        if iso in old_map:
            updated += 1
        else:
            added += 1
        new_shapes.append(entry)

    # Extracted'da olmayan mevcut shapeleri koru
    for iso, entry in old_map.items():
        if iso not in extracted:
            new_shapes.append(entry)
            kept += 1
            print(f"  KORUNDU (NE'de yok): {iso}")

    # Alfabetik sirala
    new_shapes.sort(key=lambda s: s["id"])

    out = {"shapes": new_shapes}
    with open(SHAPES_PATH, "w", encoding="utf-8") as f:
        json.dump(out, f, indent=2, ensure_ascii=False)

    print(f"\nSonuc: {updated} guncellendi, {added} eklendi, {kept} korundu")
    print(f"Toplam {len(new_shapes)} shape -> {SHAPES_PATH}")


if __name__ == "__main__":
    main()
