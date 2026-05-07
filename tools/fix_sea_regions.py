"""
Deniz bölgeleri düzeltmesi:
  1. world_x/world_y koordinatlarını harita içine taşı (mevcut bazıları ekran dışında)
  2. Deniz-deniz komşulukları ekle (bidirectional)
  3. Her deniz bölgesine komşu kıyı kara bölgelerini listesine ekle (bidirectional)

Koordinat dönüşümü (mapgen.go):
  px = -530 + wx * 2.025   →  geçerli wx: ~262-1690
  py = -180 + wy * 2.025   →  geçerli wy: ~89-800
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

def get_r(rid):
    return next((r for r in regions if r['id'] == rid), None)

def add_nb(rid, nb):
    r = get_r(rid)
    if r is not None and nb not in r.get('neighbors', []):
        r.setdefault('neighbors', []).append(nb)

# ═══════════════════════════════════════════════════════════════
# 1. DENİZ BÖLGE KOORDİNATLARINI DÜZELTİLMİŞ DEĞERLERE SET ET
# ═══════════════════════════════════════════════════════════════
# Hedef: her bölge kendi denizinin haritada görünen orta noktasında
sea_coords = {
    '_sea_atlantic':  (280, 490),   # Atlantik — haritanın sol kenarında ince şerit
    '_sea_irish':     (400, 240),   # İrlanda Denizi — İngiltere ile İrlanda arası
    '_sea_north':     (555, 195),   # Kuzey Denizi — İngiltere/Danimarka arası
    '_sea_baltic':    (845, 140),   # Baltık Denizi
    '_sea_med_west':  (570, 635),   # Batı Akdeniz
    '_sea_adriatic':  (770, 540),   # Adriyatik / Orta Akdeniz
    '_sea_med_east':  (1010, 700),  # Doğu Akdeniz
    '_sea_aegean':    (975, 615),   # Ege Denizi
    '_sea_black':     (1145, 455),  # Karadeniz
    '_sea_red':       (1200, 755),  # Kızıl Deniz
    '_sea_caspian':   (1565, 500),  # Hazar Denizi (kapalı)
    '_sea_persian':   (1515, 755),  # Basra Körfezi
}

updated = 0
for rid, (wx, wy) in sea_coords.items():
    r = get_r(rid)
    if r:
        r['world_x'] = wx
        r['world_y'] = wy
        r['is_sea'] = True
        updated += 1
        print(f"  {rid}: ({wx}, {wy})")
    else:
        print(f"UYARI: {rid} bulunamadı")
print(f"Koordinat güncellenen: {updated}")

# ═══════════════════════════════════════════════════════════════
# 2. DENİZ-DENİZ KOMŞULUKLAR (her ikisine de ekle)
# ═══════════════════════════════════════════════════════════════
sea_connections = [
    ('_sea_atlantic',  '_sea_irish'),
    ('_sea_atlantic',  '_sea_north'),
    ('_sea_atlantic',  '_sea_med_west'),
    ('_sea_irish',     '_sea_north'),
    ('_sea_north',     '_sea_baltic'),
    ('_sea_med_west',  '_sea_adriatic'),
    ('_sea_med_west',  '_sea_med_east'),
    ('_sea_adriatic',  '_sea_med_east'),
    ('_sea_adriatic',  '_sea_aegean'),
    ('_sea_med_east',  '_sea_aegean'),
    ('_sea_med_east',  '_sea_red'),
    ('_sea_aegean',    '_sea_black'),
    ('_sea_red',       '_sea_persian'),
    # Hazar — kapalı deniz, bağlantısı yok
]

for a, b in sea_connections:
    add_nb(a, b)
    add_nb(b, a)
print(f"Deniz-deniz bağlantı eklendi: {len(sea_connections)} çift")

# ═══════════════════════════════════════════════════════════════
# 3. KARA BÖLGE ↔ DENİZ KOMŞULUĞUNU BİDİRECTIONAL YAP
# Her kara bölgesi zaten _sea_* referans ediyor;
# şimdi deniz bölgelerinin de o kara bölgelerini listesinde olmasını sağla.
# ═══════════════════════════════════════════════════════════════
sea_ids = {r['id'] for r in regions if r.get('is_sea')}
coast_added = 0
for land in regions:
    if land.get('is_sea'):
        continue
    for nid in land.get('neighbors', []):
        if nid in sea_ids:
            sea_r = get_r(nid)
            lid = land['id']
            if sea_r is not None and lid not in sea_r.get('neighbors', []):
                sea_r.setdefault('neighbors', []).append(lid)
                coast_added += 1
print(f"Kiyi <-> deniz ters baglanti eklendi: {coast_added}")

# ═══════════════════════════════════════════════════════════════
# 4. DOĞRULAMA
# ═══════════════════════════════════════════════════════════════
from collections import Counter
ids = Counter(r['id'] for r in regions)
dupes = {k: v for k, v in ids.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID doğrulama: temiz")

print("\nDeniz bölge komşuları:")
for r in regions:
    if r.get('is_sea'):
        sea_nb  = [n for n in r.get('neighbors', []) if n.startswith('_sea')]
        land_nb = [n for n in r.get('neighbors', []) if not n.startswith('_sea')]
        print(f"  {r['id']:20} deniz:{sea_nb}  kıyı sayısı:{len(land_nb)}")

# ═══════════════════════════════════════════════════════════════
# 5. KAYDET
# ═══════════════════════════════════════════════════════════════
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
print("\nKaydedildi.")
