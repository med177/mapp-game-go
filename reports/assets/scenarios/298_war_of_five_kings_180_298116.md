# 298_war_of_five_kings — 174 Tur Simülasyon Raporu

- Tarih: 295/01 → 309/07
- Oyuncu devleti: Gece Nöbetçileri
- Seed: `298116`
- Zorluk: 3
- Karar politikası: Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.

- Durdurma nedeni: `oyun_bitti`

## Nihai devlet durumu

| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Martell | 24 (+4) | 24 | 254 | 14 | 381 | 249 | 20→339 | +14 | 1272 | 26655 | ayakta |
| Kuzey Krallığı | 22 (+3) | 5 | 66 | 1 | 416 | 203 | 29→305 | +13 | 240 | 30254 | ayakta |
| Akgezenler | 15 (+14) | 5 | 61 | 0 | 47 | 1 | 4→227 | +17 | 0 | 24403 | ayakta |
| Tyrell | 14 (+2) | 10 | 152 | 1 | 423 | 308 | 18→245 | +16 | 1051 | 53362 | ayakta |
| Braavos | 14 (+13) | 15 | 106 | 62 | 533 | 170 | 8→347 | +19 | 363 | 56732 | ayakta |
| Lannister | 13 (+8) | 22 | 342 | 70 | 991 | 353 | 16→372 | +18 | 28146 | 46422 | ayakta |
| Arryn | 7 (+0) | 5 | 93 | 0 | 98 | 86 | 14→100 | +8 | 286 | 944 | ayakta |
| Greyjoy | 5 (+0) | 5 | 46 | 12 | 159 | 31 | 11→63 | +11 | 0 | 199 | ayakta |
| Baratheon (Ejderha Kayası) | 4 (+1) | 9 | 95 | 19 | 346 | 80 | 16→96 | +18 | 348 | 2092 | ayakta |
| Targaryen | 1 (+0) | 4 | 35 | 4 | 172 | 38 | 9→30 | +23 | 332 | 8312 | ayakta |
| Baratheon (Fırtına Burnu) | 1 (-10) | 1 | 3 | 0 | 6 | 30 | 13→3 | +9 | 89 | 82 | ayakta |
| Özgür Halk | 0 (-4) | 0 | 0 | 0 | 126 | 21 | 20→0 | +8 | 260 | 415 | ELENDİ |
| Gece Nöbetçileri | 0 (-6) | 0 | 0 | 0 | 46 | 123 | 61→0 | +14 | 1681 | 5755 | ELENDİ |
| Özgür Şehirler | 0 (-6) | 0 | 0 | 0 | 5 | 56 | 11→0 | +8 | 3419 | 4188 | ELENDİ |
| Kardeşlik | 0 (-1) | 0 | 0 | 0 | 0 | 4 | 1→0 | +3 | 66 | 1071 | ELENDİ |
| Bolton | 0 (-1) | 0 | 0 | 0 | 6 | 7 | 8→0 | +3 | 7 | 804 | ELENDİ |
| Tully | 0 (-6) | 0 | 0 | 0 | 86 | 101 | 13→0 | +14 | 812 | 2666 | ELENDİ |
| Baratheon (Kralın Şehri) | 0 (-5) | 0 | 0 | 0 | 131 | 112 | 16→0 | +12 | 0 | 0 | ELENDİ |

Not: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.

## Toprak değişimleri

- **Martell**: +Stoms End, Oldtown, Ayazçam Yaylası, Faces Island, Soğuksu Kıyısı, Donuk Ayaz Çanağı, Kralın Şehri; -Oldtown, Stoms End, Ayazçam Yaylası, Soğuksu Kıyısı, Donuk Ayaz Çanağı
- **Kuzey Krallığı**: +Moat Cailin, Taşlı Kıyı, Kraken Burnu, Dehşet Kalesi, Soğuksu Kıyısı, Donuk Ayaz Çanağı, Kırağıkoru Sırtı, Ayazçam Yaylası, Taç Toprakları, Harrenhal, Deepwood Motte, Uzun Göl, Karhold, Boyun, Ayı Adası, Köstebek Kasabası; -Moat Cailin, Soğuksu Kıyısı, Ayazçam Yaylası, Kırağıkoru Sırtı, Donuk Ayaz Çanağı, Harrenhal, Taç Toprakları, Karhold, Uzun Göl, Deepwood Motte, Ayı Adası, Boyun, Kraken Burnu
- **Akgezenler**: +Ürperen Zirveler, Donuk Ayaz Çanağı, Akçam Sırtları, Kırağıkoru Sırtı, Ayazçam Yaylası, Kışiz Kıyısı, Soğuksu Kıyısı, Soluk Yel Sınırı, Duvar'ın Ötesi, Duvarın Batısı, Duvarın Doğusu, Brandon Hediyesi, Köstebek Kasabası, Hediye, Kraliçe Tacı; -Ayazçam Yaylası, Soğuksu Kıyısı, Donuk Ayaz Çanağı, Kışiz Kıyısı, Akçam Sırtları, Ürperen Zirveler, Kırağıkoru Sırtı, Soluk Yel Sınırı, Duvar'ın Ötesi, Köstebek Kasabası
- **Tyrell**: +Oldtown, Kral Ormanı, Summerhall, Reach, Kralın Şehri, Fırtına Diyarı, Griffin Yuvası, Taş Miğfer, Sis Ormanı, Yağmur Ormanı, Hapstack Hall, Gazap Burnu, Tumbleton; -Oldtown, Reach, Kral Ormanı, Kralın Şehri, Çimenli Vadi, Ashford, Summerhall, Altın Koruluk, Eski Meşe, Sis Ormanı
- **Braavos**: +Norvos, İhtilaflı Topraklar, Myr, Meereen, Essos Free Cities, Yunkai, Gazap Burnu, Yengeç Pençesi Burnu, Duskendale, Kral Ormanı, Harrenhal, Kralın Şehri, Reach, Çimenli Vadi, Ashford, Summerhall, Pentos (Özgür Şehir), Altın Koruluk, Ayı Adası, Eski Meşe; -Meereen, Yengeç Pençesi Burnu, Harrenhal, Gazap Burnu, Kralın Şehri, Summerhall, Ayı Adası
- **Lannister**: +Harrenhal, Oyuk Tepe, Üç Dişli Mızrak, The Crag, Nehir Koşusu, Nehir Toprakları, İkizler, Kışiz Kıyısı, Akçam Sırtları, Ürperen Zirveler, Soğuksu Kıyısı, Taç Toprakları, Karhold, Deepwood Motte, Boyun, Kraken Burnu, Kralın Şehri; -Harrenhal, Oyuk Tepe, Ürperen Zirveler, Soğuksu Kıyısı, Akçam Sırtları, Kışiz Kıyısı, Deepwood Motte, Taç Toprakları, Kralın Şehri, Karhold, Boyun, Kraken Burnu
- **Greyjoy**: +Moat Cailin, Kraken Burnu; -Moat Cailin, Kraken Burnu
- **Baratheon (Ejderha Kayası)**: +Kral Ormanı, Fırtına Diyarı, Yengeç Pençesi Burnu, Soğuksu Kıyısı, Sis Ormanı; -Kral Ormanı, Yengeç Pençesi Burnu, Fırtına Diyarı, Soğuksu Kıyısı
- **Targaryen**: +Meereen, Kralın Şehri, Taç Toprakları, Uzun Göl; -Pentos (Özgür Şehir), Kralın Şehri, Taç Toprakları, Uzun Göl
- **Baratheon (Fırtına Burnu)**: +Taç Toprakları, Stoms End, Gazap Burnu, Summerhall; -Stoms End, Gazap Burnu, Fırtına Diyarı, Summerhall, Griffin Yuvası, Taş Miğfer, Sis Ormanı, Yağmur Ormanı, Hapstack Hall, Tumbleton
- **Özgür Halk**: +Soluk Yel Sınırı, Kışiz Kıyısı, Kırağıkoru Sırtı, Ayazçam Yaylası; -Kırağıkoru Sırtı, Ayazçam Yaylası, Kışiz Kıyısı, Soğuksu Kıyısı, Soluk Yel Sınırı, Duvar'ın Ötesi
- **Gece Nöbetçileri**: +Soluk Yel Sınırı, Uzun Göl, Duvar'ın Ötesi; -Soluk Yel Sınırı, Duvarın Batısı, Uzun Göl, Duvarın Doğusu, Brandon Hediyesi, Köstebek Kasabası, Hediye, Kraliçe Tacı, Duvar'ın Ötesi
- **Özgür Şehirler**: +İhtilaflı Topraklar, Astapor; -Astapor, Norvos, İhtilaflı Topraklar, Myr, Meereen, Essos Free Cities, Yunkai
- **Kardeşlik**: +—; -Oyuk Tepe
- **Bolton**: +—; -Dehşet Kalesi
- **Tully**: +Oyuk Tepe, Harrenhal; -Harrenhal, Oyuk Tepe, Üç Dişli Mızrak, The Crag, Nehir Koşusu, Nehir Toprakları, İkizler
- **Baratheon (Kralın Şehri)**: +Taç Toprakları, Gazap Burnu, Yengeç Pençesi Burnu, Reach, Harrenhal; -Taç Toprakları, Kral Ormanı, Reach, Harrenhal, Gazap Burnu, Yengeç Pençesi Burnu, Faces Island, Duskendale, Kralın Şehri

## Savaşlar

| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |
|---|---:|---:|---|---:|---:|
| Kardeşlik — Tully | 1 | 16. tur | bitti | 0 / 0 | 0 / 0 |
| Tully — Tyrell | 1 | 5. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 2 | 9. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 14 | 18. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 24 | 28. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Gece Nöbetçileri | 27 | 33. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Tully | 27 | 31. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 34 | 41. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Fırtına Burnu) | 50 | 54. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 50 | 54. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 50 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Tyrell | 50 | 140. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Akgezenler | 61 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Akgezenler | 61 | devam ediyor | aktif | 1 / 1 | 4 / 0 |
| Baratheon (Fırtına Burnu) — Akgezenler | 61 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Greyjoy — Akgezenler | 61 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Lannister — Akgezenler | 61 | devam ediyor | aktif | 5 / 5 | 44 / 20 |
| Martell — Akgezenler | 61 | devam ediyor | aktif | 2 / 3 | 3 / 0 |
| Gece Nöbetçileri — Akgezenler | 61 | devam ediyor | bitti | 1 / 9 | 68 / 37 |
| Akgezenler — Kuzey Krallığı | 61 | devam ediyor | aktif | 4 / 7 | 10 / 73 |
| Akgezenler — Targaryen | 61 | devam ediyor | aktif | 0 / 1 | 0 / 2 |
| Akgezenler — Tyrell | 61 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Arryn — Martell | 63 | 67. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Özgür Halk | 63 | 67. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Martell | 63 | 69. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 63 | 92. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Martell | 63 | 67. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Kralın Şehri) | 77 | 83. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Braavos | 77 | 92. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Özgür Halk | 77 | 81. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Martell | 77 | 81. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kuzey Krallığı | 77 | 81. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Martell | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Kuzey Krallığı | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Kuzey Krallığı | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tyrell | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Özgür Halk | 89 | 92. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Martell | 89 | 92. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kuzey Krallığı | 89 | 92. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Tyrell | 95 | 106. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tyrell | 95 | 99. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Özgür Halk | 99 | 103. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Gece Nöbetçileri | 99 | 105. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Gece Nöbetçileri | 99 | 103. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 102 | 108. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Kuzey Krallığı | 102 | 106. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Kuzey Krallığı | 102 | 106. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Kuzey Krallığı | 102 | 108. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Kuzey Krallığı | 102 | 115. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Kuzey Krallığı | 102 | 106. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Martell | 105 | 109. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Targaryen | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Targaryen | 105 | 124. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Targaryen | 105 | 109. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Ejderha Kayası) | 109 | 113. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Targaryen | 109 | 115. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Tyrell | 112 | 117. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Tyrell | 112 | 121. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Martell | 115 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Targaryen | 115 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Ejderha Kayası) | 153 | 157. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Tyrell | 153 | 157. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Tyrell | 153 | 157. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tyrell | 153 | 157. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tyrell | 153 | 157. tur | bitti | 0 / 0 | 0 / 0 |

## Ticaret ve ittifak hareketleri

- 1. tur: **başladı** — Arryn ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Arryn ↔ Martell (grain)
- 1. tur: **başladı** — Arryn ↔ Tyrell (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Arryn (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Lannister (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Martell (grain)
- 1. tur: **başladı** — Braavos ↔ Lannister (spice)
- 1. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 1. tur: **başladı** — Lannister ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Lannister ↔ Braavos (grain)
- 1. tur: **başladı** — Martell ↔ Arryn (spice)
- 1. tur: **başladı** — Martell ↔ Baratheon (Fırtına Burnu) (spice)
- 1. tur: **başladı** — Martell ↔ Tully (spice)
- 1. tur: **başladı** — Gece Nöbetçileri ↔ Kuzey Krallığı (grain)
- 1. tur: **başladı** — Kuzey Krallığı ↔ Braavos (grain)
- 1. tur: **başladı** — Kuzey Krallığı ↔ Gece Nöbetçileri (grain)
- 1. tur: **başladı** — Tully ↔ Martell (grain)
- 1. tur: **başladı** — Tyrell ↔ Arryn (grain)
- 2. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tully (grain)
- 2. tur: **başladı** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 9. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 9. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 10. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (grain)
- 10. tur: **başladı** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 10. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 10. tur: **başladı** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 23. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 23. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 26. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Braavos (spice)
- 26. tur: **başladı** — Braavos ↔ Baratheon (Ejderha Kayası) (spice)
- 27. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Gece Nöbetçileri (spice)
- 27. tur: **başladı** — Gece Nöbetçileri ↔ Baratheon (Ejderha Kayası) (timber)
- 50. tur: **başladı** — Arryn ↔ Tully (cloth)
- 50. tur: **başladı** — Tully ↔ Arryn (grain)
- 50. tur: **bitti** — Arryn ↔ Baratheon (Fırtına Burnu) (grain)
- 50. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tully (grain)
- 50. tur: **bitti** — Baratheon (Fırtına Burnu) ↔ Arryn (grain)
- 50. tur: **bitti** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 58. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 58. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 63. tur: **başladı** — Arryn ↔ Targaryen (spice)
- 63. tur: **başladı** — Targaryen ↔ Arryn (spice)
- 63. tur: **bitti** — Arryn ↔ Martell (grain)
- 63. tur: **bitti** — Martell ↔ Arryn (spice)
- 66. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (spice)
- 66. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (spice)
- 67. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (spice)
- 67. tur: **başladı** — Braavos ↔ Targaryen (spice)
- 67. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 67. tur: **başladı** — Martell ↔ Tully (spice)
- 67. tur: **başladı** — Targaryen ↔ Braavos (spice)
- 67. tur: **başladı** — Tully ↔ Martell (grain)
- 70. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Lannister (spice)
- 70. tur: **başladı** — Lannister ↔ Baratheon (Fırtına Burnu) (iron)
- 74. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 74. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 77. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 77. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 79. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (spice)
- 79. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Lannister (spice)
- 79. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 79. tur: **başladı** — Lannister ↔ Baratheon (Fırtına Burnu) (iron)
- 86. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (spice)
- 86. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Lannister (spice)
- 86. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 86. tur: **başladı** — Lannister ↔ Baratheon (Fırtına Burnu) (iron)
- 86. tur: **başladı** — Martell ↔ Targaryen (spice)
- 86. tur: **başladı** — Targaryen ↔ Martell (spice)
- 87. tur: **başladı** — Martell ↔ Targaryen (spice)
- 87. tur: **başladı** — Targaryen ↔ Martell (spice)
- 88. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Lannister (spice)
- 88. tur: **başladı** — Lannister ↔ Baratheon (Kralın Şehri) (iron)
- 89. tur: **başladı** — Martell ↔ Targaryen (spice)
- 89. tur: **başladı** — Targaryen ↔ Martell (spice)
- 94. tur: **başladı** — Martell ↔ Tyrell (spice)
- 94. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 94. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 94. tur: **başladı** — Tyrell ↔ Martell (spice)
- 95. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 95. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 96. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 96. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 97. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 97. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 98. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 98. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 101. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 101. tur: **başladı** — Lannister ↔ Martell (iron)
- 101. tur: **başladı** — Martell ↔ Lannister (spice)
- 101. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (iron)
- 102. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 102. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 102. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (spice)
- 102. tur: **başladı** — Lannister ↔ Martell (grain)
- 102. tur: **başladı** — Martell ↔ Lannister (spice)
- 102. tur: **başladı** — Gece Nöbetçileri ↔ Tyrell (iron)
- 102. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (cloth)
- 102. tur: **başladı** — Tyrell ↔ Gece Nöbetçileri (spice)
- 102. tur: **bitti** — Gece Nöbetçileri ↔ Kuzey Krallığı (iron)
- 102. tur: **bitti** — Kuzey Krallığı ↔ Gece Nöbetçileri (iron)
- 103. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (spice)
- 103. tur: **başladı** — Lannister ↔ Martell (grain)
- 103. tur: **başladı** — Martell ↔ Lannister (spice)
- 103. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (spice)
- 104. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (spice)
- 104. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (spice)
- 105. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 105. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (spice)
- 109. tur: **bitti** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 109. tur: **bitti** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 118. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (cloth)
- 118. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (cloth)
- 121. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (cloth)
- 121. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (cloth)
- 124. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Targaryen (cloth)
- 124. tur: **başladı** — Targaryen ↔ Baratheon (Fırtına Burnu) (cloth)
- 132. tur: **başladı** — Greyjoy ↔ Lannister (iron)
- 132. tur: **başladı** — Lannister ↔ Greyjoy (spice)
- 133. tur: **başladı** — Greyjoy ↔ Lannister (iron)
- 133. tur: **başladı** — Lannister ↔ Greyjoy (spice)
- 137. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 137. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 138. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 138. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 143. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (cloth)
- 143. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 153. tur: **bitti** — Arryn ↔ Baratheon (Ejderha Kayası) (cloth)
- 153. tur: **bitti** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 164. tur: **başladı** — Braavos ↔ Kuzey Krallığı (cloth)
- 164. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 165. tur: **başladı** — Braavos ↔ Kuzey Krallığı (cloth)
- 165. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 1. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 1. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 1. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Kardeşlik ↔ Tyrell
- 2. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 2. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 3. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 4. tur: **ittifak kuruldu** — Lannister ↔ Tyrell
- 5. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 6. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 7. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 12. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 15. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kardeşlik
- 16. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 16. tur: **ittifak kuruldu** — Arryn ↔ Greyjoy
- 16. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 16. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Tully
- 19. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 19. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 21. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 22. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Tully
- 24. tur: **ittifak kuruldu** — Tully ↔ Tyrell
- 25. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 25. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 28. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 28. tur: **ittifak kuruldu** — Özgür Halk ↔ Kuzey Krallığı
- 29. tur: **ittifak bitti** — Özgür Halk ↔ Kuzey Krallığı
- 29. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 30. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 31. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 33. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 35. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 37. tur: **ittifak bitti** — Arryn ↔ Lannister
- 37. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 39. tur: **ittifak bitti** — Arryn ↔ Tully
- 40. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 40. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 40. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 40. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 43. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 46. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 50. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 52. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 53. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 55. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 57. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 57. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 58. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 58. tur: **ittifak bitti** — Lannister ↔ Tully
- 58. tur: **ittifak kuruldu** — Bolton ↔ Gece Nöbetçileri
- 59. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 61. tur: **ittifak kuruldu** — Arryn ↔ Gece Nöbetçileri
- 61. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 62. tur: **ittifak kuruldu** — Arryn ↔ Özgür Halk
- 62. tur: **ittifak kuruldu** — Özgür Halk ↔ Martell
- 62. tur: **ittifak kuruldu** — Özgür Halk ↔ Kuzey Krallığı
- 62. tur: **ittifak kuruldu** — Özgür Halk ↔ Targaryen
- 62. tur: **ittifak kuruldu** — Martell ↔ Tyrell
- 63. tur: **ittifak bitti** — Lannister ↔ Martell
- 64. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 64. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 67. tur: **ittifak kuruldu** — Braavos ↔ Kuzey Krallığı
- 68. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 71. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 76. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Özgür Halk
- 78. tur: **ittifak bitti** — Arryn ↔ Lannister
- 82. tur: **ittifak kuruldu** — Braavos ↔ Lannister
- 85. tur: **ittifak bitti** — Özgür Halk ↔ Kuzey Krallığı
- 85. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 86. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Özgür Halk
- 86. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Targaryen
- 88. tur: **ittifak kuruldu** — Lannister ↔ Gece Nöbetçileri
- 88. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Targaryen
- 89. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 89. tur: **ittifak kuruldu** — Arryn ↔ Targaryen
- 92. tur: **ittifak kuruldu** — Martell ↔ Targaryen
- 93. tur: **ittifak bitti** — Martell ↔ Tyrell
- 94. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 98. tur: **ittifak bitti** — Braavos ↔ Kuzey Krallığı
- 99. tur: **ittifak bitti** — Arryn ↔ Özgür Halk
- 99. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Özgür Halk
- 99. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Özgür Halk
- 99. tur: **ittifak bitti** — Özgür Halk ↔ Targaryen
- 99. tur: **ittifak bitti** — Lannister ↔ Gece Nöbetçileri
- 99. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Targaryen
- 99. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Lannister
- 100. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Özgür Halk
- 100. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Özgür Halk
- 100. tur: **ittifak kuruldu** — Lannister ↔ Gece Nöbetçileri
- 101. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Lannister
- 102. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Targaryen
- 102. tur: **ittifak bitti** — Martell ↔ Targaryen
- 102. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Targaryen
- 103. tur: **ittifak bitti** — Arryn ↔ Lannister
- 103. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Braavos
- 104. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 105. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Targaryen
- 109. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 110. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 112. tur: **ittifak bitti** — Braavos ↔ Lannister
- 113. tur: **ittifak kuruldu** — Braavos ↔ Lannister
- 116. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Targaryen
- 117. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Braavos
- 120. tur: **ittifak bitti** — Braavos ↔ Lannister
- 124. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Targaryen
- 128. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Targaryen
- 129. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri
- 135. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Braavos
- 138. tur: **ittifak kuruldu** — Arryn ↔ Tyrell
- 139. tur: **ittifak kuruldu** — Braavos ↔ Gece Nöbetçileri
- 153. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı
- 168. tur: **ittifak kuruldu** — Arryn ↔ Braavos

## Elenen devletler

- 16. tur: **Kardeşlik**
- 41. tur: **Özgür Şehirler**
- 67. tur: **Tully**
- 69. tur: **Bolton**
- 92. tur: **Baratheon (Kralın Şehri)**
- 107. tur: **Özgür Halk**
- 174. tur: **Gece Nöbetçileri**

## Tur checkpoint'leri

### 9. tur (295/10)

1. Martell — 20 bölge, 3 ordu, 32 kara birimi, güç 319
2. Stark — 19 bölge, 3 ordu, 55 kara birimi, güç 680
3. Tyrell — 12 bölge, 2 ordu, 26 kara birimi, güç 427
4. Baratheon (Fırtına Burnu) — 12 bölge, 1 ordu, 14 kara birimi, güç 179
5. Arryn — 7 bölge, 2 ordu, 32 kara birimi, güç 555

### 19. tur (296/08)

1. Martell — 21 bölge, 2 ordu, 37 kara birimi, güç 308
2. Stark — 19 bölge, 6 ordu, 94 kara birimi, güç 857
3. Tyrell — 12 bölge, 3 ordu, 34 kara birimi, güç 525
4. Baratheon (Fırtına Burnu) — 11 bölge, 1 ordu, 15 kara birimi, güç 185
5. Arryn — 7 bölge, 2 ordu, 32 kara birimi, güç 382

### 29. tur (297/06)

1. Martell — 21 bölge, 5 ordu, 42 kara birimi, güç 442
2. Stark — 18 bölge, 5 ordu, 83 kara birimi, güç 771
3. Tyrell — 12 bölge, 6 ordu, 44 kara birimi, güç 613
4. Baratheon (Fırtına Burnu) — 10 bölge, 1 ordu, 15 kara birimi, güç 126
5. Arryn — 7 bölge, 2 ordu, 40 kara birimi, güç 407

### 39. tur (298/04)

1. Martell — 21 bölge, 5 ordu, 47 kara birimi, güç 564
2. Stark — 19 bölge, 8 ordu, 119 kara birimi, güç 914
3. Tyrell — 11 bölge, 4 ordu, 53 kara birimi, güç 597
4. Baratheon (Fırtına Burnu) — 11 bölge, 4 ordu, 18 kara birimi, güç 167
5. Arryn — 7 bölge, 3 ordu, 41 kara birimi, güç 210

### 49. tur (299/02)

1. Martell — 21 bölge, 7 ordu, 51 kara birimi, güç 558
2. Kuzey Krallığı — 20 bölge, 9 ordu, 131 kara birimi, güç 941
3. Tyrell — 12 bölge, 6 ordu, 57 kara birimi, güç 690
4. Baratheon (Fırtına Burnu) — 9 bölge, 2 ordu, 18 kara birimi, güç 185
5. Arryn — 7 bölge, 3 ordu, 44 kara birimi, güç 626

### 59. tur (299/12)

1. Martell — 21 bölge, 7 ordu, 57 kara birimi, güç 914
2. Kuzey Krallığı — 20 bölge, 8 ordu, 138 kara birimi, güç 1586
3. Tyrell — 12 bölge, 7 ordu, 76 kara birimi, güç 1308
4. Baratheon (Fırtına Burnu) — 8 bölge, 0 ordu, 0 kara birimi, güç 0
5. Lannister — 7 bölge, 11 ordu, 112 kara birimi, güç 1568

### 69. tur (300/10)

1. Kuzey Krallığı — 22 bölge, 10 ordu, 154 kara birimi, güç 1483
2. Martell — 21 bölge, 5 ordu, 42 kara birimi, güç 748
3. Tyrell — 14 bölge, 7 ordu, 77 kara birimi, güç 1298
4. Lannister — 11 bölge, 16 ordu, 126 kara birimi, güç 1720
5. Akgezenler — 8 bölge, 4 ordu, 59 kara birimi, güç 2630

### 79. tur (301/08)

1. Kuzey Krallığı — 23 bölge, 10 ordu, 159 kara birimi, güç 1467
2. Martell — 21 bölge, 5 ordu, 45 kara birimi, güç 735
3. Lannister — 14 bölge, 15 ordu, 145 kara birimi, güç 1757
4. Tyrell — 14 bölge, 8 ordu, 58 kara birimi, güç 624
5. Braavos — 8 bölge, 7 ordu, 60 kara birimi, güç 1949

### 89. tur (302/06)

1. Kuzey Krallığı — 24 bölge, 10 ordu, 160 kara birimi, güç 1359
2. Martell — 24 bölge, 6 ordu, 55 kara birimi, güç 743
3. Lannister — 15 bölge, 13 ordu, 147 kara birimi, güç 1977
4. Tyrell — 13 bölge, 6 ordu, 52 kara birimi, güç 547
5. Braavos — 9 bölge, 8 ordu, 76 kara birimi, güç 2418

### 99. tur (303/04)

1. Kuzey Krallığı — 24 bölge, 11 ordu, 168 kara birimi, güç 1655
2. Martell — 22 bölge, 6 ordu, 53 kara birimi, güç 713
3. Lannister — 12 bölge, 12 ordu, 126 kara birimi, güç 1454
4. Tyrell — 12 bölge, 3 ordu, 14 kara birimi, güç 66
5. Braavos — 11 bölge, 11 ordu, 79 kara birimi, güç 2302

### 109. tur (304/02)

1. Martell — 22 bölge, 7 ordu, 61 kara birimi, güç 755
2. Kuzey Krallığı — 18 bölge, 8 ordu, 114 kara birimi, güç 805
3. Braavos — 15 bölge, 12 ordu, 88 kara birimi, güç 2573
4. Lannister — 15 bölge, 10 ordu, 115 kara birimi, güç 1881
5. Akgezenler — 10 bölge, 7 ordu, 95 kara birimi, güç 2173

### 119. tur (304/12)

1. Martell — 22 bölge, 8 ordu, 78 kara birimi, güç 1104
2. Lannister — 17 bölge, 15 ordu, 183 kara birimi, güç 4246
3. Kuzey Krallığı — 17 bölge, 7 ordu, 89 kara birimi, güç 991
4. Braavos — 15 bölge, 10 ordu, 94 kara birimi, güç 3296
5. Akgezenler — 10 bölge, 6 ordu, 95 kara birimi, güç 3641

### 129. tur (305/10)

1. Martell — 24 bölge, 9 ordu, 99 kara birimi, güç 1313
2. Kuzey Krallığı — 18 bölge, 8 ordu, 113 kara birimi, güç 1237
3. Lannister — 16 bölge, 32 ordu, 355 kara birimi, güç 5008
4. Braavos — 15 bölge, 11 ordu, 96 kara birimi, güç 3477
5. Akgezenler — 10 bölge, 4 ordu, 67 kara birimi, güç 2682

### 139. tur (306/08)

1. Martell — 24 bölge, 15 ordu, 130 kara birimi, güç 1729
2. Kuzey Krallığı — 20 bölge, 11 ordu, 137 kara birimi, güç 1417
3. Braavos — 15 bölge, 13 ordu, 99 kara birimi, güç 3341
4. Tyrell — 15 bölge, 5 ordu, 71 kara birimi, güç 1053
5. Lannister — 14 bölge, 25 ordu, 362 kara birimi, güç 5230

### 149. tur (307/06)

1. Martell — 24 bölge, 18 ordu, 172 kara birimi, güç 2123
2. Kuzey Krallığı — 20 bölge, 9 ordu, 150 kara birimi, güç 1116
3. Braavos — 15 bölge, 11 ordu, 97 kara birimi, güç 3464
4. Tyrell — 15 bölge, 11 ordu, 106 kara birimi, güç 1497
5. Lannister — 14 bölge, 23 ordu, 362 kara birimi, güç 4699

### 159. tur (308/04)

1. Martell — 24 bölge, 23 ordu, 206 kara birimi, güç 2160
2. Kuzey Krallığı — 20 bölge, 12 ordu, 165 kara birimi, güç 1547
3. Braavos — 15 bölge, 9 ordu, 100 kara birimi, güç 3366
4. Tyrell — 14 bölge, 7 ordu, 126 kara birimi, güç 1616
5. Lannister — 13 bölge, 22 ordu, 342 kara birimi, güç 5178

### 169. tur (309/02)

1. Martell — 24 bölge, 24 ordu, 241 kara birimi, güç 2123
2. Kuzey Krallığı — 21 bölge, 5 ordu, 63 kara birimi, güç 449
3. Braavos — 14 bölge, 12 ordu, 104 kara birimi, güç 3044
4. Tyrell — 14 bölge, 8 ordu, 145 kara birimi, güç 1416
5. Lannister — 13 bölge, 22 ordu, 342 kara birimi, güç 4782

### 174. tur (309/07)

1. Martell — 24 bölge, 24 ordu, 254 kara birimi, güç 2564
2. Kuzey Krallığı — 22 bölge, 5 ordu, 66 kara birimi, güç 772
3. Akgezenler — 15 bölge, 5 ordu, 61 kara birimi, güç 2517
4. Braavos — 14 bölge, 15 ordu, 106 kara birimi, güç 3284
5. Tyrell — 14 bölge, 10 ordu, 152 kara birimi, güç 1811

## Tetiklenen olaylar

- 6. tur: **Kraliyet Borçlarının Gölgesi** (`crown_debt_crisis_295`)
- 14. tur: **Kuzeyin Kış Hazırlıkları** (`northern_preparations_296`)
- 20. tur: **Lannister Altını ve Tacın Borçları** (`lannister_gold_and_crown_296`)
- 27. tur: **Ejderha Kayası'nda Seferberlik** (`dragonstone_muster_297`)
- 37. tur: **Jon Arryn'in Ölümü ve Şüpheler** (`jon_arryn_investigation_298`)
- 38. tur: **Catelyn Tyrion'u Yakalıyor** (`catelyn_captures_tyrion_298`)
- 38. tur: **Eddard Stark Kral Eli Oluyor** (`ned_stark_named_hand_298`)
- 38. tur: **Sansa ve Joffrey'nin Nişanı** (`sansa_joffrey_betrothal_298`)
- 39. tur: **Kral Robert'ın Ölümü** (`death_of_robert_298`)
- 39. tur: **Renly'nin Güneydeki İttifak Arayışı** (`renly_southern_alliance_297`)
- 40. tur: **Ned Stark'ın Tutuklanması** (`ned_arrested_after_robert_298`)
- 40. tur: **Vadi'nin Siyasi Tutumu** (`vale_political_stance_298`)
- 42. tur: **Ned Stark'ın Kaderi** (`ned_execution_298`)
- 43. tur: **Nehir Toprakları Savaşının Yükü** (`lannister_riverlands_war_burden_298`)
- 46. tur: **Kuzeyden Gelen Alametler** (`winter_omens_298`)
- 52. tur: **Robb'un Kuzey Kralı İlanı** (`robb_king_in_the_north_299`)
- 53. tur: **Jaime Lannister Esir Alındı** (`jaime_captured_299`)
- 53. tur: **Renly'nin Gölgeyle Öldürülmesi** (`renly_shadow_assassination_299`)
- 55. tur: **Kara Su Savaşı** (`battle_of_the_blackwater_299`)
- 56. tur: **Edmure ve Roslin'in Evliliği** (`edmure_roslin_marriage_299`)
- 56. tur: **Greyjoy Saldırısı** (`greyjoy_rebellion_299`)
- 56. tur: **Margaery ve Joffrey'nin Evliliği** (`tyrell_joffrey_marriage_299`)
- 57. tur: **Kızıl Düğün İhaneti** (`red_wedding_event`)
- 58. tur: **Kuzey Baskınının Sonuçları** (`greyjoy_northern_raid_aftermath_299`)
- 58. tur: **İttifakın Savaş Bedeli** (`lannister_tyrell_alliance_war_cost_299`)
- 58. tur: **Riverrun Savunması** (`riverrun_defense_299`)
- 58. tur: **Kızıl Düğün'den Sonra Tullylerin Kararı** (`tully_after_red_wedding_299`)
- 58. tur: **Vadi Kapılarını Kapatıyor** (`vale_neutrality_after_red_wedding_299`)
- 60. tur: **Kuzeyde Kış Sertleşiyor** (`winter_storms_299`)
- 61. tur: **Jaime'nin Kaderi** (`jaime_fate_300`)
- 61. tur: **Buzun Altındaki Ordu** (`others_awaken_300`)
- 62. tur: **Demir Donanma Seferi** (`iron_fleet_campaign_300`)
- 62. tur: **Mor Düğün** (`purple_wedding_300`)
- 62. tur: **Riverrun Kuşatmasının Sonucu** (`riverrun_siege_outcome_300`)
- 64. tur: **Blackfish'in Son Direnişi** (`blackfish_resistance_300`)
- 64. tur: **Reach Savaş Seferberliği** (`tyrell_war_mobilization_cost_300`)
- 65. tur: **Greyjoy Savaş Konseyi** (`greyjoy_succession_300`)
- 65. tur: **Lannister Savaş Yıpranması** (`lannister_war_exhaustion_300`)
- 66. tur: **Vadi'de Veraset Krizi** (`vale_succession_crisis_300`)
- 68. tur: **Gece Nöbeti'nin Duvar Hazırlıkları** (`night_watch_wall_preparations_300`)
- 68. tur: **Batı İttifakının Çatlağı** (`tyrell_lannister_tension_300`)
- 68. tur: **Tywin Lannister'ın Ölümü** (`tywin_death_300`)
- 69. tur: **Demir Adalar'ın İkmal Krizi** (`ironborn_supply_crisis_300`)
- 69. tur: **Tyrion'un Kaderi** (`tyrion_fate_300`)
- 70. tur: **Mance Rayder'ın Yürüyüşü** (`mance_rayder_invasion_300`)
- 75. tur: **Duvar Savaşının Eşiğinde** (`wall_pressure_301`)
- 77. tur: **Akgezenlere Karşı Son Hazırlık** (`night_watch_against_others_301`)
- 78. tur: **Duvar Savunmasının Bedeli** (`wall_breach_after_defense_301`)
- 107. tur: **Özgür Halkın Ardından** (`night_watch_after_free_folk_defeat`)
- 114. tur: **Daenerys'in Meereen Seferi** (`daenerys_meereen_campaign_301`)
- 116. tur: **Lannister Zaferinin Düzeni** (`postwar_settlement_lannister`)
- 116. tur: **Reach'in Savaş Sonrası Konsolidasyonu** (`tyrell_postwar_consolidation_301`)
- 116. tur: **Beş Kralın Savaşı: Lannister Zaferi** (`war_of_five_kings_resolution_lannister`)
- 129. tur: **Targaryenlerin Westeros'a Dönüşü** (`targaryen_return_302`)

## Set edilen flag'ler

- 37. tur: `jon_arryn_path_chosen`
- 37. tur: `royal_party_travelling_north`
- 38. tur: `ned_stark_named_hand`
- 38. tur: `sansa_joffrey_betrothal`
- 38. tur: `tully_lannister_conflict_escalated`
- 38. tur: `tyrion_captured`
- 38. tur: `tyrion_delivered_to_the_eyrie`
- 39. tur: `baratheon_claimants_war_started`
- 39. tur: `joffrey_claim_supported`
- 39. tur: `robert_dead`
- 39. tur: `war_of_five_kings_active`
- 40. tur: `arryn_neutrality_declared`
- 40. tur: `ned_arrested`
- 40. tur: `ned_betrayed`
- 42. tur: `ned_executed`
- 43. tur: `lannister_riverlands_burden`
- 46. tur: `white_walkers_stirring`
- 52. tur: `king_in_the_north`
- 52. tur: `northern_independence_declared`
- 53. tur: `jaime_captured`
- 53. tur: `jaime_hostage_held`
- 53. tur: `renly_dead`
- 53. tur: `renly_host_joined_stannis`
- 55. tur: `blackwater_lannister_victory`
- 56. tur: `frey_marriage_agreed`
- 56. tur: `greyjoy_campaign_started`
- 56. tur: `greyjoy_raids_north`
- 56. tur: `greyjoy_rebellion_started`
- 56. tur: `highgarden_joined_war`
- 56. tur: `margaery_joffrey_married`
- 56. tur: `tully_frey_truce`
- 56. tur: `tyrell_lannister_alliance`
- 57. tur: `northern_resistance_continues`
- 57. tur: `red_wedding_resisted`
- 57. tur: `stark_betrayed`
- 58. tur: `blackfish_resistance`
- 58. tur: `greyjoy_northern_raid_success`
- 58. tur: `lannister_tyrell_war_burden`
- 58. tur: `riverrun_defense_prepared`
- 58. tur: `riverrun_resistance`
- 58. tur: `riverrun_siege_started`
- 58. tur: `vale_refused_stark_aid`
- 60. tur: `long_winter_begun`
- 61. tur: `jaime_held_through_war`
- 61. tur: `others_awakened`
- 62. tur: `iron_fleet_expanded`
- 62. tur: `purple_wedding_completed`
- 62. tur: `riverrun_held_against_siege`
- 62. tur: `tyrion_on_trial`
- 64. tur: `riverlands_recapture_campaign`
- 64. tur: `tyrell_war_mobilized`
- 65. tur: `balon_independence_line`
- 65. tur: `lannister_war_exhausted`
- 66. tur: `arryn_independent_regency`
- 68. tur: `cersei_controls_lannisters`
- 68. tur: `tyrell_alliance_broken`
- 68. tur: `tyrell_alliance_preserved`
- 68. tur: `tywin_dead`
- 68. tur: `wall_defense_hardened`
- 68. tur: `wall_defense_prepared`
- 69. tur: `ironborn_war_tax_imposed`
- 69. tur: `tyrion_exiled_to_essos`
- 70. tur: `mance_invasion_started`
- 78. tur: `wall_breached`
- 114. tur: `meereen_campaign_started`
- 114. tur: `targaryen_dragons_grown`
- 114. tur: `targaryen_return_ready`
- 116. tur: `postwar_lannister_order`
- 116. tur: `tyrell_postwar_consolidation`
- 116. tur: `war_of_five_kings_resolved`
- 116. tur: `war_of_five_kings_settlement_complete`
- 116. tur: `war_of_five_kings_winner_lannister`
- 129. tur: `targaryen_restoration_started`
