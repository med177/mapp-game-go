"""
NE admin_0 shapefile'daki tum kayitlari tarar:
- Kırım bbox (32-37, 44-47) icinde veri olan ISO kodlari
- Trakya bbox (25-31, 40-43) icinde veri olan ISO kodlari
"""
import shapefile

SHP_PATH = "_REFERENCE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp"

def points_in_bbox(ring, lon_min, lat_min, lon_max, lat_max):
    return sum(1 for lon, lat in ring
               if lon_min <= lon <= lon_max and lat_min <= lat <= lat_max)

sf = shapefile.Reader(SHP_PATH)
fields = [f[0] for f in sf.fields[1:]]

print("=== Kırım bbox (lon 32-37, lat 44-47) ===")
for sr in sf.shapeRecords():
    d = dict(zip(fields, sr.record))
    shape = sr.shape
    parts = list(shape.parts) + [len(shape.points)]
    for i in range(len(parts)-1):
        ring = shape.points[parts[i]:parts[i+1]]
        cnt = points_in_bbox(ring, 32, 44, 37, 47)
        if cnt >= 3:
            lons = [p[0] for p in ring if 32 <= p[0] <= 37 and 44 <= p[1] <= 47]
            lats = [p[1] for p in ring if 32 <= p[0] <= 37 and 44 <= p[1] <= 47]
            print(f"  ISO={d.get('ADM0_A3'):6s} SOVEREIGNT={d.get('SOVEREIGNT','')[:20]:20s} "
                  f"ring[{i}] pts_in_bbox={cnt:4d}  "
                  f"lat=[{min(lats):.2f},{max(lats):.2f}] lon=[{min(lons):.2f},{max(lons):.2f}]")

print()
print("=== Trakya bbox (lon 25-31, lat 40-43) ===")
for sr in sf.shapeRecords():
    d = dict(zip(fields, sr.record))
    shape = sr.shape
    parts = list(shape.parts) + [len(shape.points)]
    for i in range(len(parts)-1):
        ring = shape.points[parts[i]:parts[i+1]]
        cnt = points_in_bbox(ring, 25, 40, 31, 43)
        if cnt >= 3:
            lons = [p[0] for p in ring if 25 <= p[0] <= 31 and 40 <= p[1] <= 43]
            lats = [p[1] for p in ring if 25 <= p[0] <= 31 and 40 <= p[1] <= 43]
            print(f"  ISO={d.get('ADM0_A3'):6s} SOVEREIGNT={d.get('SOVEREIGNT','')[:20]:20s} "
                  f"ring[{i}] pts_in_bbox={cnt:4d}  "
                  f"lat=[{min(lats):.2f},{max(lats):.2f}] lon=[{min(lons):.2f},{max(lons):.2f}]")
