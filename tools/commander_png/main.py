from PIL import Image
import zipfile
import io

image_path = "Tarihi Komutanlar Portre Arşivi.png"
output_zip = "portreler.zip"

# Görseldeki dosya isimleri (soldan sağa, yukarıdan aşağıya)
filenames = [
    "ottoman_kuyucu_murad_pasha.png", "ottoman_halil_pasha.png", "ottoman_kose_mihal.png",
    "ottoman_fazil_ahmed_pasha.png", "ottoman_kara_mustafa_pasha.png", "ottoman_fazil_mustafa_pasha.png", "ottoman_amcazade_huseyin.png",
    "ottoman_baltaci_mehmed.png", "ottoman_deli_huseyin.png", "ottoman_murtaza_pasha.png",
    "castile_kingdom_ambrosio_spinola.png", "castile_kingdom_juan_jose_austria.png", "castile_kingdom_eugenio_savoia.png", "castile_kingdom_villadarias.png",
    "castile_kingdom_pavia_commander.png", "france_turenne.png", "france_conde.png",
    "france_la_feuillade.png", "france_boufflers.png", "france_villars.png", "france_crillon.png",
    "france_montauban.png", "france_caumont.png", "france_talleyrand.png",
    "france_blanqui.png", "france_soubise.png", "france_villeneuve.png", "france_duras.png"
]

# Görseli yükle
img = Image.open(image_path)
img_width, img_height = img.size

cols, rows = 7, 4
cell_width = img_width // cols
cell_height = img_height // rows

with zipfile.ZipFile(output_zip, 'w') as zipf:
    for index, filename in enumerate(filenames):
        col = index % cols
        row = index // cols

        # Hücre koordinatlarını hesapla
        left = col * cell_width
        top = row * cell_height
        right = (col + 1) * cell_width
        
        # Sadece kare portre alanını kırp (yazıyı dışarıda bırakmak için yüksekliği genişliğe eşitliyoruz)
        portrait = img.crop((left, top, right, top + cell_width))
        
        # 96x96 boyutuna getir
        resized = portrait.resize((96, 96), Image.Resampling.LANCZOS)

        # Diske yazmadan doğrudan bellekte tutup zip dosyasına aktar
        img_byte_arr = io.BytesIO()
        resized.save(img_byte_arr, format='PNG')
        zipf.writestr(filename, img_byte_arr.getvalue())

print(f"{len(filenames)} adet 96x96 portre başarıyla {output_zip} dosyasına sıkıştırıldı.")