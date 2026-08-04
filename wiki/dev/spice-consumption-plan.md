---
type: dev
tags: [economy, spice, upkeep, consumption]
last_updated: 2026-08-05
related: [systems/economy, world/regions, dev/progress, architecture/game-loop]
---

# 1300 Baharat Sürekli Tüketim Planı

## Amaç

1300 senaryosunda baharatı yalnızca bina ve birim üretiminde kullanılan tek
seferlik bir kaynak olmaktan çıkarıp, halk ve ordunun düzenli gıda/ikmal
sepetinin düşük miktarlı bir parçası yapmak.

Tahıl temel gıda olarak kalır. Baharat daha az miktarda tüketilir ve doğal
üretim bölgeleri, başlangıç stoğu ve ticaret yoluyla desteklenir.

## 1300 Mevcut Durum Denetimi

- `regions.json` içinde `base_spice_output > 0` olan 70 bölge var.
- Ham toplam baharat üretimi 687/tur.
- En yüksek ham üretim Memlük bölgelerinde 150/tur, İlhanlı bölgelerinde
  104/tur, Venedik'te 55/tur ve Doğu Roma'da 42/tur.
- Baharat sahibi olunan bölge üretiminden her tur otomatik oluşur; yalnızca
  satın alınan bir kaynak değildir.
- Üretim arazi uzmanlaşması, abluka, yağma ve başkent bonusu akışlarından geçer.
- Ticaret sistemi `GoodSpice` destekliyor; mevcut stok ticaret rotalarıyla
  sürekli artırılabilir veya başka devlete aktarılabilir.
- Başlangıç faction stoklarında baharat değerleri mevcut.
- `farm` yalnızca tahıl üretimini `grain_mod: 1.3` ile artırıyor; baharatı
  artıran bir bina veya baharat üretim bonusu yok.
- `market`, `port` ve `temple` baharat maliyeti kullanabiliyor ama baharat
  üretmiyor.

Sonuç: 1300'te baharat hem sürekli üretilen hem de ticaretle alınabilen bir
kaynaktır. Ancak üretim, oyuncunun istediği bölgeye çiftlik kurarak artırılamaz;
doğal baharat bölgelerini ele geçirmek veya ticaret ağına bağlanmak gerekir.

## Önerilen Veri Modeli

### Ordu bakımı

`internal/army.UnitType` içine yeni alan:

```json
"grain_upkeep": 8,
"spice_upkeep": 1
```

- `spice_upkeep` birlik başına tur tüketimidir.
- Eski unit JSON'larında alan yoksa `0` kabul edilir.
- Tahıl bakımındaki hareket, garnizon ve kuşatma katsayıları ilk iterasyonda
  baharat bakımına da uygulanır.

### Sivil tüketim

Sivil baharat talebi nüfustan türetilir; bölgelere sabit tüketim yazılmaz.
1300 senaryo ayarına şu oran eklenir:

```json
"civilian_spice_population_unit": 100
```

Başlangıç dengesi: 100 nüfus başına 1 baharat/tur. Böylece baharat zorunlu
olur ama tahılla aynı miktarda tüketilmez.

### Kıtlık davranışı

Baharat eksikliği tahıl kıtlığından ayrı değerlendirilir:

- Tahıl eksikliği: açlık, ciddi memnuniyet ve ordu kaybı.
- Baharat eksikliği: öncelikle memnuniyet ve ordu morali/ikmal verimi cezası.
- Uzun süreli baharat açlığına asker kaybı eklenmesi ancak denge testinden
  sonra değerlendirilir.

## Uygulama Aşamaları

1. `UnitType.SpiceUpkeep` alanını ve JSON sözleşmesini ekle.
2. Ordu için `TotalSpiceUpkeep` ve hareket/kuşatma etkili ortak bakım hesabını
   ekle.
3. Bölge nüfusundan sivil baharat talebini hesapla.
4. Tur çözümünde üretim + ticaret + yağma sonrasında sivil ve askerî baharat
   tüketimini faction stoğuna uygula.
5. Baharat kıtlığı seviyesini ve tahıldan bağımsız etkilerini ekle.
6. HUD, tur özeti ve bölge/ordu tooltip'lerinde baharat üretim-tüketim-net
   değişimini göster.
7. AI'nin öngörülen baharat bakımını rezerv hesabına katmasını ve eksik
   baharatı aktif ticaret ağından almasını sağla.
8. 1300 başlangıç stoklarını ve üretim oranlarını simülasyon testiyle kalibre
   et; mevcut tahıl ekonomi bantlarını gevşetme.

## Test Planı

- `spice_upkeep` JSON yükleme ve eksik alanın geriye uyumlu olarak `0` olması.
- Sabit, hareketli, garnizon ve kuşatma ordularında baharat bakım hesabı.
- Nüfus oranına göre sivil baharat talebi.
- Aynı turda üretim, ticaret, yağma ve tüketim sıralaması.
- Baharat kıtlığının tahıl kıtlığından ayrı uygulanması.
- 1300 toplam baharat üretimi ve faction stoklarının denge bandı.
- AI'nin baharat açığını tespit edip ticaret yoluyla kapatması.

## Karar

İlk uygulama için önerilen değerler:

- `civilian_spice_population_unit: 100`
- temel askerlerde `spice_upkeep: 0 veya 1`
- elit/deniz/uzun ikmal gerektiren birliklerde daha yüksek değer
- baharat açığında önce yumuşak memnuniyet/moral cezası

Bu plan yalnız 1300 senaryosunu kapsar; eski senaryolar için veri veya kural
değişikliği bu planın kapsamında değildir.
