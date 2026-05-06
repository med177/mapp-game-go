"""
PNG uzerine sekil cercevelerini cizer ve karsilastirma icin kaydeder.
Kullanim: python tools/debug_alignment.py
"""
from PIL import Image, ImageDraw
import json

SHAPES_PATH = "assets/data/generated/country_shapes.json"
PNG_PATH    = "assets/maps/world_map_background.png"
OUT_PATH    = "assets/maps/debug_alignment.png"

# Kalibrasyon sabitleri - mapgen.go ile ayni tutulmali
SHAPE_W, SHAPE_H = 1828, 997
BG_PX0, BG_PY0 = 22.0, 67.0
BG_PX1, BG_PY1 = 2794.0, 1579.0

SCALE_X = (BG_PX1 - BG_PX0) / SHAPE_W
SCALE_Y = (BG_PY1 - BG_PY0) / SHAPE_H

def shape_to_png(sx, sy):
    return BG_PX0 + sx * SCALE_X, BG_PY0 + sy * SCALE_Y

def lon_to_sx(lon): return 20.0 * lon + 476.0
def lat_to_sy(lat): return -18.98 * lat + 1234.8

img = Image.open(PNG_PATH).convert("RGBA")
PNG_W, PNG_H = img.size

overlay = Image.new("RGBA", (PNG_W, PNG_H), (0, 0, 0, 0))
draw = ImageDraw.Draw(overlay)

with open(SHAPES_PATH, "r", encoding="utf-8") as f:
    shapes = json.load(f)["shapes"]

for shape in shapes:
    for ring in shape["rings"]:
        pts = [shape_to_png(p[0], p[1]) for p in ring]
        if len(pts) >= 2:
            draw.line(pts + [pts[0]], fill=(255, 80, 80, 200), width=2)

for lon in range(-20, 75, 10):
    px = BG_PX0 + lon_to_sx(lon) * SCALE_X
    if 0 <= px <= PNG_W:
        draw.line([(px, 0), (px, PNG_H)], fill=(80, 80, 255, 80), width=1)
        label = str(lon) + ("E" if lon >= 0 else "W")
        draw.text((px+2, 5), label, fill=(80, 80, 255, 180))

for lat in range(60, 10, -10):
    py = BG_PY0 + lat_to_sy(lat) * SCALE_Y
    if 0 <= py <= PNG_H:
        draw.line([(0, py), (PNG_W, py)], fill=(80, 80, 255, 80), width=1)
        draw.text((5, py+2), str(lat)+"N", fill=(80, 80, 255, 180))

draw.rectangle([BG_PX0, BG_PY0, BG_PX1, BG_PY1], outline=(255,255,0,200), width=3)

result = Image.alpha_composite(img, overlay)
result.save(OUT_PATH)
result.resize((1408, 768), Image.LANCZOS).save("assets/maps/debug_alignment_small.png")
print("OK: " + OUT_PATH)
