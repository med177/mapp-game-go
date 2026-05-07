"""
'anatolia' bolgesini 'konya' olarak yeniden adlandir ve Karamanogullari'na ver.

Tarihsel gerekce (1300 donemi):
  - Anadolu Selcuklu Sultanligi 1307'de tamamen coktu.
  - Moğol Ilhanlilari nominal hakim olsa da Karamanogullari
    Selcuk mirasini devraldi ve Konya'yi elinde tuttu.
  - 'anatolia' isimli bolge 1300 haritasinda yanlis temsil,
    yerini 'konya' (Karamanoğulları merkezi) almali.

Yapilan degisiklikler:
  - id: anatolia -> konya
  - name: Konya
  - name_tr: Konya (Karamanogullari)
  - owner_id: "" -> karaman_bey
  - world_y: 462 -> 472  (biraz guneye, Selcuk kalbi)
  - Komsu listesinden _sea_black kaldir (Karaman Karadeniz'e ulasmaz)
  - Tum regions.json icindeki 'anatolia' komsu referanslari 'konya' olur
"""
import json

with open('assets/data/regions.json', encoding='utf-8') as f:
    regions = json.load(f)

# ─── 1. Ana bölgeyi güncelle ─────────────────────────────────────────────────
r = next((r for r in regions if r['id'] == 'anatolia'), None)
if not r:
    print("HATA: anatolia bulunamadi!")
    exit(1)

r['id']       = 'konya'
r['name']     = 'Konya'
r['name_tr']  = 'Konya (Karamanogullari)'
r['owner_id'] = 'karaman_bey'
r['world_y']  = 472          # biraz guneye cek
# _sea_black kaldir (Karaman Karadeniz'e ulasmiyordu)
if '_sea_black' in r['neighbors']:
    r['neighbors'].remove('_sea_black')
    print("anatolia->konya: _sea_black komsulugundan cikarildi")

print(f"anatolia -> konya donusturuldu (wx={r['world_x']}, wy={r['world_y']})")

# ─── 2. Tüm komşu referanslarını güncelle ────────────────────────────────────
updated = 0
for region in regions:
    nb = region.get('neighbors', [])
    if 'anatolia' in nb:
        region['neighbors'] = ['konya' if n == 'anatolia' else n for n in nb]
        updated += 1
        print(f"  {region['id']:20} komsusu anatolia->konya")
print(f"Komsu referansi guncellenen: {updated}")

# ─── 3. Doğrulama ────────────────────────────────────────────────────────────
from collections import Counter
ids = Counter(r['id'] for r in regions)
dupes = {k: v for k, v in ids.items() if v > 1}
if dupes:
    print(f"UYARI duplicate: {dupes}")
else:
    print("ID dogrulama: temiz")

# anatolia referansi kalmamali
leftover = [r['id'] for r in regions if 'anatolia' in r.get('neighbors', [])]
if leftover:
    print(f"UYARI hala 'anatolia' komsulugu var: {leftover}")
else:
    print("'anatolia' referansi temizlendi")

# konya ve karaman yan yana kontrol
konya  = next(r for r in regions if r['id'] == 'konya')
karaman = next(r for r in regions if r['id'] == 'karaman')
print(f"\nKaramanogullari bolgeleri:")
print(f"  konya:   wx={konya['world_x']:4} wy={konya['world_y']:4}  nb={konya['neighbors']}")
print(f"  karaman: wx={karaman['world_x']:4} wy={karaman['world_y']:4}  nb={karaman['neighbors']}")

# ─── 4. Kaydet ───────────────────────────────────────────────────────────────
with open('assets/data/regions.json', 'w', encoding='utf-8') as f:
    json.dump(regions, f, ensure_ascii=False, indent=2)
print("\nKaydedildi.")
