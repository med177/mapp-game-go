# 298_war_of_five_kings — 120 Tur Simülasyon Raporu

- Tarih: 295/01 → 305/01
- Oyuncu devleti: Gece Nöbetçileri
- Seed: `298113`
- Zorluk: 3
- Karar politikası: Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.

## Nihai devlet durumu

| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Martell | 27 (+7) | 12 | 103 | 5 | 316 | 186 | 20→276 | +14 | 2138 | 20189 | ayakta |
| Kuzey Krallığı | 26 (+7) | 8 | 98 | 0 | 166 | 131 | 29→264 | +10 | 583 | 21511 | ayakta |
| Tyrell | 10 (-2) | 13 | 164 | 26 | 229 | 248 | 18→212 | +13 | 353 | 34118 | ayakta |
| Lannister | 9 (+4) | 14 | 133 | 3 | 587 | 230 | 16→231 | +17 | 1789 | 29726 | ayakta |
| Braavos | 6 (+5) | 11 | 150 | 30 | 155 | 140 | 8→168 | +19 | 15263 | 9214 | ayakta |
| Kardeşlik | 5 (+4) | 13 | 90 | 18 | 154 | 89 | 1→131 | +22 | 1278 | 44106 | ayakta |
| Gece Nöbetçileri | 5 (+1) | 1 | 15 | 0 | 15 | 10 | 37→53 | +18 | 426 | 8112 | ayakta |
| Baratheon (Ejderha Kayası) | 4 (+1) | 5 | 45 | 10 | 176 | 67 | 16→99 | +20 | 1697 | 2033 | ayakta |
| Baratheon (Fırtına Burnu) | 4 (-7) | 3 | 1 | 3 | 9 | 32 | 13→37 | +10 | 489 | 3349 | ayakta |
| Tully | 3 (-3) | 4 | 35 | 14 | 172 | 72 | 13→74 | +18 | 2354 | 9484 | ayakta |
| Greyjoy | 3 (-2) | 8 | 33 | 7 | 102 | 19 | 11→23 | +7 | 750 | 777 | ayakta |
| Baratheon (Kralın Şehri) | 3 (-2) | 1 | 20 | 0 | 233 | 166 | 16→28 | +12 | 0 | 10365 | ayakta |
| Targaryen | 2 (+1) | 6 | 56 | 14 | 172 | 38 | 9→62 | +23 | 38390 | 7330 | ayakta |
| Akgezenler | 1 (+0) | 1 | 10 | 0 | 17 | 2 | 0→2 | +9 | 332 | 44 | ayakta |
| Bolton | 0 (-1) | 0 | 0 | 0 | 35 | 7 | 8→0 | +10 | 228 | 1306 | ELENDİ |
| Özgür Halk | 0 (-1) | 0 | 0 | 0 | 1 | 5 | 5→0 | +2 | 255 | 361 | ELENDİ |
| Özgür Şehirler | 0 (-6) | 0 | 0 | 0 | 1 | 35 | 11→0 | +5 | 4694 | 1440 | ELENDİ |
| Arryn | 0 (-7) | 0 | 0 | 0 | 76 | 100 | 14→0 | +7 | 86 | 2772 | ELENDİ |

Not: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.

## Toprak değişimleri

- **Martell**: +Stoms End, Oldtown, Sharp Point, Sis Ormanı, Griffin Yuvası, Tarth, Kralın Şehri; -Oldtown, Stoms End
- **Kuzey Krallığı**: +Taşlı Kıyı, Flint Parmağı, Kraken Burnu, Deepwood Motte, Moat Cailin, Parmaklar, Batı Vadi, Büyük Wyk, Dehşet Kalesi, Ay Dağları, Wickenden, Kartal Yuvası; -Moat Cailin, Flint Parmağı, Deepwood Motte, Taşlı Kıyı, Kraken Burnu, Batı Vadi, Parmaklar
- **Tyrell**: +Oldtown, Kral Ormanı, Summerhall, Fırtına Diyarı, Hapstack Hall; -Oldtown, Summerhall, Altın Koruluk, Fırtına Diyarı, Hapstack Hall, Gümüş Tepe
- **Lannister**: +Harrenhal, Nehir Koşusu, Flint Parmağı, Deepwood Motte, The Crag, Taşlı Kıyı, Kraken Burnu, Taş Septi, Altın Koruluk, Gümüş Tepe; -Harrenhal, Flint Parmağı, Taş Septi, Deepwood Motte, Taşlı Kıyı, The Crag
- **Braavos**: +İhtilaflı Topraklar, Norvos, Essos Free Cities, Yunkai, Myr, Meereen; -Meereen
- **Kardeşlik**: +Nehir Koşusu, Üç Dişli Mızrak, Nehir Toprakları, Harrenhal, Yengeç Pençesi Burnu, Taç Toprakları, Duskendale; -Nehir Koşusu, Harrenhal, Taç Toprakları, Yengeç Pençesi Burnu
- **Gece Nöbetçileri**: +Duvar'ın Ötesi; -—
- **Baratheon (Ejderha Kayası)**: +Yengeç Pençesi Burnu, Fırtına Diyarı, Taç Toprakları, Faces Island; -Yengeç Pençesi Burnu, Sharp Point, Fırtına Diyarı
- **Baratheon (Fırtına Burnu)**: +Taç Toprakları, Stoms End, Fırtına Diyarı, Taş Septi, Summerhall, Hapstack Hall; -Taç Toprakları, Stoms End, Fırtına Diyarı, Hapstack Hall, Sis Ormanı, Griffin Yuvası, Gazap Burnu, Yağmur Ormanı, Taş Miğfer, Tarth, Summerhall, Taş Septi
- **Tully**: +Deepwood Motte, The Crag, Flint Parmağı, Tuz Tavaları; -Harrenhal, Nehir Koşusu, Üç Dişli Mızrak, Nehir Toprakları, The Crag, Deepwood Motte, Flint Parmağı
- **Greyjoy**: +Moat Cailin, Flint Parmağı, Taşlı Kıyı; -Kraken Burnu, Moat Cailin, Taşlı Kıyı, Flint Parmağı, Büyük Wyk
- **Baratheon (Kralın Şehri)**: +Taç Toprakları, Yengeç Pençesi Burnu, Sharp Point, Hapstack Hall, Fırtına Diyarı, Harrenhal, Gazap Burnu, Yağmur Ormanı, Taş Miğfer, Summerhall; -Taç Toprakları, Sharp Point, Harrenhal, Kral Ormanı, Summerhall, Fırtına Diyarı, Kralın Şehri, Hapstack Hall, Yengeç Pençesi Burnu, Faces Island, Duskendale
- **Targaryen**: +Meereen; -—
- **Akgezenler**: +Duvar'ın Ötesi; -Duvar'ın Ötesi
- **Bolton**: +—; -Dehşet Kalesi
- **Özgür Halk**: +—; -Duvar'ın Ötesi
- **Özgür Şehirler**: +—; -Astapor, Norvos, Essos Free Cities, Yunkai, Myr, Meereen
- **Arryn**: +Batı Vadi, Parmaklar; -Parmaklar, Batı Vadi, Ay Dağları, Tuz Tavaları, Martı Kasabası, Wickenden, Kartal Yuvası

## Savaşlar

| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |
|---|---:|---:|---|---:|---:|
| Kardeşlik — Tully | 1 | 16. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 4 | 10. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 7 | 11. tur | aktif | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 9 | 13. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 16 | 22. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 21 | 33. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 54 | 62. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kardeşlik | 54 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Greyjoy | 54 | 58. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Lannister | 54 | 58. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Martell | 54 | 77. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Gece Nöbetçileri | 62 | 66. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 62 | 118. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 68 | 77. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Greyjoy | 68 | 72. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 68 | 110. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tully | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Lannister | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Kardeşlik — Lannister | 80 | 85. tur | aktif | 1 / 1 | 0 / 0 |
| Greyjoy — Lannister | 80 | 93. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Gece Nöbetçileri | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Tully | 80 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 81 | 85. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Gece Nöbetçileri | 90 | 94. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Kuzey Krallığı | 90 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Targaryen | 90 | 94. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tully | 94 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 94 | 98. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kardeşlik | 94 | 117. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Martell | 94 | 98. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Gece Nöbetçileri | 94 | 98. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Kuzey Krallığı | 94 | 100. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 94 | 98. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tyrell | 94 | 99. tur | aktif | 3 / 0 | 41 / 60 |
| Martell — Tully | 94 | 98. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Gece Nöbetçileri | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 105 | 109. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 106 | 110. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Gece Nöbetçileri | 106 | 110. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Kuzey Krallığı | 106 | 110. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 106 | 113. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tyrell | 106 | devam ediyor | bitti | 0 / 0 | 0 / 0 |
| Martell — Tully | 106 | 110. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Gece Nöbetçileri | 112 | 116. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tully | 112 | 116. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tyrell | 112 | 116. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kardeşlik | 113 | 117. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Lannister | 113 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Lannister | 113 | 117. tur | bitti | 0 / 0 | 0 / 0 |
| Kardeşlik — Lannister | 113 | devam ediyor | bitti | 0 / 0 | 0 / 0 |
| Kardeşlik — Martell | 113 | devam ediyor | aktif | 2 / 1 | 0 / 3 |
| Greyjoy — Lannister | 113 | 117. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Martell | 115 | devam ediyor | aktif | 0 / 2 | 0 / 1 |
| Baratheon (Ejderha Kayası) — Kuzey Krallığı | 115 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Targaryen | 115 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Kardeşlik — Targaryen | 115 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Lannister | 117 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Martell | 117 | devam ediyor | aktif | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Targaryen | 117 | 119. tur | bitti | 0 / 0 | 0 / 0 |

## Ticaret ve ittifak hareketleri

- 1. tur: **başladı** — Arryn ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Arryn ↔ Martell (grain)
- 1. tur: **başladı** — Arryn ↔ Tyrell (grain)
- 1. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (grain)
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
- 1. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (grain)
- 1. tur: **başladı** — Kuzey Krallığı ↔ Braavos (grain)
- 1. tur: **başladı** — Tully ↔ Martell (grain)
- 1. tur: **başladı** — Tully ↔ Tyrell (grain)
- 1. tur: **başladı** — Tyrell ↔ Arryn (grain)
- 1. tur: **başladı** — Tyrell ↔ Tully (grain)
- 2. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tully (grain)
- 2. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tyrell (grain)
- 2. tur: **başladı** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 2. tur: **başladı** — Tyrell ↔ Baratheon (Kralın Şehri) (grain)
- 11. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (grain)
- 11. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 20. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kardeşlik (spice)
- 20. tur: **başladı** — Kardeşlik ↔ Baratheon (Ejderha Kayası) (grain)
- 20. tur: **başladı** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 20. tur: **başladı** — Gece Nöbetçileri ↔ Greyjoy (timber)
- 39. tur: **başladı** — Braavos ↔ Targaryen (spice)
- 39. tur: **başladı** — Targaryen ↔ Braavos (spice)
- 40. tur: **başladı** — Kardeşlik ↔ Kuzey Krallığı (grain)
- 40. tur: **başladı** — Kuzey Krallığı ↔ Kardeşlik (iron)
- 50. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (spice)
- 50. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 54. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 54. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri (grain)
- 54. tur: **başladı** — Kardeşlik ↔ Gece Nöbetçileri (grain)
- 54. tur: **başladı** — Kardeşlik ↔ Targaryen (grain)
- 54. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 54. tur: **başladı** — Martell ↔ Targaryen (spice)
- 54. tur: **başladı** — Gece Nöbetçileri ↔ Baratheon (Fırtına Burnu) (cloth)
- 54. tur: **başladı** — Gece Nöbetçileri ↔ Kardeşlik (cloth)
- 54. tur: **başladı** — Targaryen ↔ Kardeşlik (spice)
- 54. tur: **başladı** — Targaryen ↔ Martell (spice)
- 54. tur: **bitti** — Baratheon (Fırtına Burnu) ↔ Lannister (grain)
- 54. tur: **bitti** — Baratheon (Fırtına Burnu) ↔ Martell (grain)
- 54. tur: **bitti** — Lannister ↔ Baratheon (Fırtına Burnu) (grain)
- 54. tur: **bitti** — Martell ↔ Baratheon (Fırtına Burnu) (spice)
- 55. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 55. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 56. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu) (iron)
- 56. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Baratheon (Ejderha Kayası) (spice)
- 56. tur: **başladı** — Braavos ↔ Greyjoy (spice)
- 56. tur: **başladı** — Greyjoy ↔ Braavos (timber)
- 57. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tully (spice)
- 57. tur: **başladı** — Braavos ↔ Greyjoy (spice)
- 57. tur: **başladı** — Greyjoy ↔ Braavos (timber)
- 57. tur: **başladı** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 58. tur: **başladı** — Braavos ↔ Greyjoy (spice)
- 58. tur: **başladı** — Greyjoy ↔ Braavos (timber)
- 58. tur: **başladı** — Greyjoy ↔ Tyrell (timber)
- 58. tur: **başladı** — Tyrell ↔ Greyjoy (spice)
- 58. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tyrell (grain)
- 58. tur: **bitti** — Tyrell ↔ Baratheon (Kralın Şehri) (grain)
- 59. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 59. tur: **başladı** — Greyjoy ↔ Tyrell (timber)
- 59. tur: **başladı** — Martell ↔ Tully (spice)
- 59. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 59. tur: **başladı** — Tully ↔ Martell (grain)
- 59. tur: **başladı** — Tyrell ↔ Greyjoy (spice)
- 60. tur: **başladı** — Greyjoy ↔ Tyrell (timber)
- 60. tur: **başladı** — Tyrell ↔ Greyjoy (spice)
- 61. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Lannister (spice)
- 61. tur: **başladı** — Lannister ↔ Baratheon (Kralın Şehri) (iron)
- 62. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 62. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 63. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (spice)
- 63. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 64. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 64. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 64. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 64. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 65. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 65. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 65. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 65. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 66. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 66. tur: **başladı** — Martell ↔ Targaryen (spice)
- 66. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 66. tur: **başladı** — Targaryen ↔ Martell (spice)
- 67. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 67. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 68. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 68. tur: **başladı** — Greyjoy ↔ Tyrell (iron)
- 68. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 68. tur: **başladı** — Tyrell ↔ Greyjoy (grain)
- 70. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 70. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 73. tur: **başladı** — Greyjoy ↔ Tyrell (iron)
- 73. tur: **başladı** — Tyrell ↔ Greyjoy (grain)
- 75. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (cloth)
- 75. tur: **başladı** — Greyjoy ↔ Tyrell (iron)
- 75. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 75. tur: **başladı** — Tyrell ↔ Greyjoy (grain)
- 76. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (cloth)
- 76. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 77. tur: **başladı** — Martell ↔ Tyrell (spice)
- 77. tur: **başladı** — Tyrell ↔ Martell (spice)
- 79. tur: **başladı** — Greyjoy ↔ Tyrell (timber)
- 79. tur: **başladı** — Tyrell ↔ Greyjoy (grain)
- 80. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tully (spice)
- 80. tur: **bitti** — Martell ↔ Tully (spice)
- 80. tur: **bitti** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 80. tur: **bitti** — Tully ↔ Martell (grain)
- 85. tur: **başladı** — Kardeşlik ↔ Targaryen (grain)
- 85. tur: **başladı** — Targaryen ↔ Kardeşlik (cloth)
- 86. tur: **başladı** — Kardeşlik ↔ Targaryen (grain)
- 86. tur: **başladı** — Targaryen ↔ Kardeşlik (cloth)
- 87. tur: **başladı** — Kardeşlik ↔ Targaryen (grain)
- 87. tur: **başladı** — Targaryen ↔ Kardeşlik (cloth)
- 89. tur: **başladı** — Kardeşlik ↔ Kuzey Krallığı (grain)
- 89. tur: **başladı** — Kuzey Krallığı ↔ Kardeşlik (iron)
- 90. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 90. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (iron)
- 90. tur: **başladı** — Kardeşlik ↔ Kuzey Krallığı (grain)
- 90. tur: **başladı** — Kuzey Krallığı ↔ Kardeşlik (iron)
- 91. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 91. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (iron)
- 93. tur: **başladı** — Kardeşlik ↔ Targaryen (grain)
- 93. tur: **başladı** — Targaryen ↔ Kardeşlik (cloth)
- 95. tur: **başladı** — Kardeşlik ↔ Tully (grain)
- 95. tur: **başladı** — Greyjoy ↔ Tully (timber)
- 95. tur: **başladı** — Tully ↔ Kardeşlik (grain)
- 95. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 96. tur: **başladı** — Arryn ↔ Kardeşlik (spice)
- 96. tur: **başladı** — Kardeşlik ↔ Arryn (grain)
- 96. tur: **başladı** — Greyjoy ↔ Tully (timber)
- 96. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 97. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (grain)
- 97. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 97. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 97. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 98. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 98. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 99. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 99. tur: **başladı** — Martell ↔ Kuzey Krallığı (spice)
- 99. tur: **başladı** — Kuzey Krallığı ↔ Martell (iron)
- 99. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 100. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 100. tur: **başladı** — Martell ↔ Kuzey Krallığı (spice)
- 100. tur: **başladı** — Kuzey Krallığı ↔ Martell (iron)
- 100. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 101. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 101. tur: **başladı** — Martell ↔ Kuzey Krallığı (spice)
- 101. tur: **başladı** — Kuzey Krallığı ↔ Martell (iron)
- 101. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 102. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 102. tur: **başladı** — Martell ↔ Kuzey Krallığı (spice)
- 102. tur: **başladı** — Kuzey Krallığı ↔ Martell (iron)
- 102. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 103. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 103. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 104. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 104. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 105. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 105. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 106. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 106. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 110. tur: **başladı** — Greyjoy ↔ Tully (iron)
- 110. tur: **başladı** — Tully ↔ Greyjoy (grain)
- 111. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Tully (cloth)
- 111. tur: **başladı** — Tully ↔ Baratheon (Fırtına Burnu) (grain)
- 112. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Greyjoy (cloth)
- 112. tur: **başladı** — Greyjoy ↔ Baratheon (Fırtına Burnu) (iron)
- 113. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Lannister (spice)
- 113. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 113. tur: **başladı** — Lannister ↔ Baratheon (Kralın Şehri) (iron)
- 113. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (iron)
- 113. tur: **bitti** — Arryn ↔ Kardeşlik (spice)
- 113. tur: **bitti** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 113. tur: **bitti** — Kardeşlik ↔ Arryn (grain)
- 113. tur: **bitti** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 114. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (grain)
- 114. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (iron)
- 115. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 115. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (iron)
- 115. tur: **bitti** — Kardeşlik ↔ Targaryen (grain)
- 115. tur: **bitti** — Targaryen ↔ Kardeşlik (cloth)
- 116. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 116. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 117. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 117. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 117. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kardeşlik (spice)
- 117. tur: **başladı** — Kardeşlik ↔ Baratheon (Kralın Şehri) (grain)
- 118. tur: **başladı** — Arryn ↔ Baratheon (Ejderha Kayası) (spice)
- 118. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Arryn (spice)
- 119. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Tully (cloth)
- 119. tur: **başladı** — Tully ↔ Baratheon (Fırtına Burnu) (grain)
- 120. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Tully (cloth)
- 120. tur: **başladı** — Tully ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 1. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 1. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Lannister ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 2. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 2. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 2. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 4. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 5. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 6. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 8. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 9. tur: **ittifak kuruldu** — Kardeşlik ↔ Tyrell
- 13. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 14. tur: **ittifak kuruldu** — Kardeşlik ↔ Lannister
- 16. tur: **ittifak bitti** — Arryn ↔ Kardeşlik
- 16. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 16. tur: **ittifak kuruldu** — Arryn ↔ Greyjoy
- 16. tur: **ittifak kuruldu** — Arryn ↔ Tully
- 16. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 16. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 18. tur: **ittifak kuruldu** — Arryn ↔ Gece Nöbetçileri
- 18. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 19. tur: **ittifak bitti** — Kardeşlik ↔ Lannister
- 19. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 20. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 20. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 20. tur: **ittifak kuruldu** — Kardeşlik ↔ Lannister
- 20. tur: **ittifak kuruldu** — Kardeşlik ↔ Martell
- 20. tur: **ittifak kuruldu** — Kardeşlik ↔ Kuzey Krallığı
- 24. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kardeşlik
- 24. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 29. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 31. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Greyjoy
- 32. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 32. tur: **ittifak kuruldu** — Braavos ↔ Kuzey Krallığı
- 34. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 36. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 36. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 36. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Tully
- 37. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kardeşlik
- 37. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 37. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 37. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 38. tur: **ittifak bitti** — Kardeşlik ↔ Lannister
- 39. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 40. tur: **ittifak bitti** — Kardeşlik ↔ Kuzey Krallığı
- 40. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 40. tur: **ittifak kuruldu** — Kardeşlik ↔ Lannister
- 40. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 41. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 42. tur: **ittifak bitti** — Arryn ↔ Lannister
- 42. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 46. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 47. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 48. tur: **ittifak kuruldu** — Kardeşlik ↔ Tully
- 49. tur: **ittifak bitti** — Kardeşlik ↔ Tully
- 50. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 51. tur: **ittifak bitti** — Kardeşlik ↔ Lannister
- 52. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 52. tur: **ittifak kuruldu** — Lannister ↔ Tyrell
- 54. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 54. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 56. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 57. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 58. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 58. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Lannister
- 61. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 61. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 61. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 61. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 61. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 61. tur: **ittifak kuruldu** — Kardeşlik ↔ Lannister
- 62. tur: **ittifak bitti** — Arryn ↔ Gece Nöbetçileri
- 62. tur: **ittifak bitti** — Arryn ↔ Tully
- 62. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Lannister
- 62. tur: **ittifak bitti** — Braavos ↔ Kuzey Krallığı
- 62. tur: **ittifak bitti** — Kardeşlik ↔ Lannister
- 62. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 62. tur: **ittifak kuruldu** — Arryn ↔ Akgezenler
- 62. tur: **ittifak kuruldu** — Kardeşlik ↔ Tully
- 63. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 64. tur: **ittifak kuruldu** — Braavos ↔ Greyjoy
- 64. tur: **ittifak kuruldu** — Kardeşlik ↔ Lannister
- 65. tur: **ittifak kuruldu** — Greyjoy ↔ Tyrell
- 66. tur: **ittifak bitti** — Arryn ↔ Akgezenler
- 67. tur: **ittifak bitti** — Greyjoy ↔ Tyrell
- 68. tur: **ittifak kuruldu** — Arryn ↔ Tully
- 68. tur: **ittifak kuruldu** — Bolton ↔ Gece Nöbetçileri
- 69. tur: **ittifak bitti** — Kardeşlik ↔ Lannister
- 69. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 69. tur: **ittifak kuruldu** — Kardeşlik ↔ Greyjoy
- 70. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 70. tur: **ittifak kuruldu** — Martell ↔ Targaryen
- 75. tur: **ittifak bitti** — Arryn ↔ Tully
- 75. tur: **ittifak kuruldu** — Arryn ↔ Tyrell
- 77. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 78. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 78. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 83. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Targaryen
- 84. tur: **ittifak bitti** — Arryn ↔ Kardeşlik
- 84. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Greyjoy
- 84. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 84. tur: **ittifak bitti** — Kardeşlik ↔ Tully
- 84. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 84. tur: **ittifak kuruldu** — Bolton ↔ Greyjoy
- 85. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Kardeşlik
- 89. tur: **ittifak kuruldu** — Arryn ↔ Bolton
- 90. tur: **ittifak bitti** — Bolton ↔ Gece Nöbetçileri
- 91. tur: **ittifak kuruldu** — Bolton ↔ Akgezenler
- 92. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Ejderha Kayası)
- 93. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 94. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 94. tur: **ittifak bitti** — Bolton ↔ Greyjoy
- 94. tur: **ittifak bitti** — Greyjoy ↔ Tully
- 94. tur: **ittifak kuruldu** — Tully ↔ Tyrell
- 95. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 95. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 98. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 98. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 102. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 106. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 110. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 111. tur: **ittifak bitti** — Arryn ↔ Greyjoy
- 111. tur: **ittifak bitti** — Arryn ↔ Tyrell
- 112. tur: **ittifak bitti** — Arryn ↔ Baratheon (Ejderha Kayası)
- 112. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 112. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 112. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 112. tur: **ittifak kuruldu** — Arryn ↔ Akgezenler
- 113. tur: **ittifak bitti** — Kardeşlik ↔ Martell
- 113. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 113. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 114. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 114. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Martell
- 114. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 114. tur: **ittifak kuruldu** — Kardeşlik ↔ Tully
- 114. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 116. tur: **ittifak bitti** — Arryn ↔ Akgezenler
- 116. tur: **ittifak kuruldu** — Braavos ↔ Gece Nöbetçileri
- 117. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 117. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Kardeşlik
- 117. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 117. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Martell
- 117. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 117. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kardeşlik
- 119. tur: **ittifak kuruldu** — Kardeşlik ↔ Kuzey Krallığı

## Elenen devletler

- 22. tur: **Özgür Şehirler**
- 33. tur: **Özgür Halk**
- 107. tur: **Bolton**
- 118. tur: **Arryn**

## Tur checkpoint'leri

### 9. tur (295/10)

1. Stark — 20 bölge, 4 ordu, 42 kara birimi, güç 536
2. Martell — 20 bölge, 3 ordu, 32 kara birimi, güç 383
3. Tyrell — 12 bölge, 2 ordu, 31 kara birimi, güç 564
4. Baratheon (Fırtına Burnu) — 12 bölge, 2 ordu, 18 kara birimi, güç 270
5. Arryn — 7 bölge, 2 ordu, 27 kara birimi, güç 438

### 19. tur (296/08)

1. Martell — 21 bölge, 2 ordu, 35 kara birimi, güç 371
2. Stark — 20 bölge, 5 ordu, 60 kara birimi, güç 576
3. Tyrell — 12 bölge, 3 ordu, 39 kara birimi, güç 670
4. Baratheon (Fırtına Burnu) — 11 bölge, 2 ordu, 20 kara birimi, güç 211
5. Arryn — 7 bölge, 2 ordu, 33 kara birimi, güç 570

### 29. tur (297/06)

1. Martell — 21 bölge, 3 ordu, 37 kara birimi, güç 478
2. Stark — 20 bölge, 7 ordu, 97 kara birimi, güç 831
3. Tyrell — 12 bölge, 6 ordu, 40 kara birimi, güç 708
4. Baratheon (Fırtına Burnu) — 11 bölge, 2 ordu, 23 kara birimi, güç 303
5. Arryn — 7 bölge, 2 ordu, 36 kara birimi, güç 586

### 39. tur (298/04)

1. Martell — 21 bölge, 6 ordu, 42 kara birimi, güç 553
2. Stark — 19 bölge, 9 ordu, 126 kara birimi, güç 923
3. Tyrell — 11 bölge, 8 ordu, 46 kara birimi, güç 728
4. Baratheon (Fırtına Burnu) — 11 bölge, 7 ordu, 31 kara birimi, güç 405
5. Braavos — 7 bölge, 5 ordu, 26 kara birimi, güç 654

### 49. tur (299/02)

1. Martell — 22 bölge, 6 ordu, 41 kara birimi, güç 550
2. Kuzey Krallığı — 19 bölge, 7 ordu, 92 kara birimi, güç 668
3. Tyrell — 11 bölge, 6 ordu, 56 kara birimi, güç 720
4. Baratheon (Fırtına Burnu) — 10 bölge, 4 ordu, 34 kara birimi, güç 444
5. Braavos — 7 bölge, 5 ordu, 34 kara birimi, güç 650

### 59. tur (299/12)

1. Martell — 23 bölge, 8 ordu, 44 kara birimi, güç 779
2. Kuzey Krallığı — 18 bölge, 6 ordu, 61 kara birimi, güç 571
3. Tyrell — 11 bölge, 9 ordu, 69 kara birimi, güç 1093
4. Baratheon (Fırtına Burnu) — 9 bölge, 2 ordu, 0 kara birimi, güç 12
5. Baratheon (Kralın Şehri) — 8 bölge, 10 ordu, 112 kara birimi, güç 1229

### 69. tur (300/10)

1. Martell — 24 bölge, 8 ordu, 54 kara birimi, güç 909
2. Kuzey Krallığı — 17 bölge, 6 ordu, 57 kara birimi, güç 743
3. Tyrell — 11 bölge, 10 ordu, 82 kara birimi, güç 1115
4. Baratheon (Kralın Şehri) — 9 bölge, 10 ordu, 94 kara birimi, güç 1445
5. Baratheon (Fırtına Burnu) — 8 bölge, 3 ordu, 6 kara birimi, güç 56

### 79. tur (301/08)

1. Martell — 25 bölge, 9 ordu, 74 kara birimi, güç 1164
2. Kuzey Krallığı — 19 bölge, 6 ordu, 76 kara birimi, güç 969
3. Baratheon (Kralın Şehri) — 13 bölge, 6 ordu, 97 kara birimi, güç 1252
4. Tyrell — 12 bölge, 12 ordu, 105 kara birimi, güç 1494
5. Lannister — 8 bölge, 10 ordu, 58 kara birimi, güç 1170

### 89. tur (302/06)

1. Martell — 25 bölge, 10 ordu, 80 kara birimi, güç 1069
2. Kuzey Krallığı — 18 bölge, 5 ordu, 83 kara birimi, güç 879
3. Baratheon (Kralın Şehri) — 13 bölge, 5 ordu, 87 kara birimi, güç 1027
4. Tyrell — 12 bölge, 14 ordu, 140 kara birimi, güç 1526
5. Lannister — 9 bölge, 14 ordu, 89 kara birimi, güç 1108

### 99. tur (303/04)

1. Martell — 25 bölge, 9 ordu, 81 kara birimi, güç 1097
2. Kuzey Krallığı — 20 bölge, 5 ordu, 66 kara birimi, güç 979
3. Tyrell — 12 bölge, 14 ordu, 187 kara birimi, güç 1587
4. Baratheon (Kralın Şehri) — 11 bölge, 4 ordu, 80 kara birimi, güç 832
5. Lannister — 9 bölge, 12 ordu, 120 kara birimi, güç 1684

### 109. tur (304/02)

1. Martell — 27 bölge, 10 ordu, 84 kara birimi, güç 1101
2. Kuzey Krallığı — 21 bölge, 6 ordu, 82 kara birimi, güç 715
3. Tyrell — 11 bölge, 13 ordu, 171 kara birimi, güç 1163
4. Lannister — 9 bölge, 13 ordu, 126 kara birimi, güç 1087
5. Baratheon (Kralın Şehri) — 8 bölge, 5 ordu, 71 kara birimi, güç 613

### 119. tur (304/12)

1. Martell — 27 bölge, 11 ordu, 99 kara birimi, güç 1643
2. Kuzey Krallığı — 26 bölge, 6 ordu, 95 kara birimi, güç 1349
3. Tyrell — 10 bölge, 14 ordu, 184 kara birimi, güç 1550
4. Lannister — 9 bölge, 12 ordu, 129 kara birimi, güç 1431
5. Braavos — 6 bölge, 11 ordu, 150 kara birimi, güç 2284

### 120. tur (305/01)

1. Martell — 27 bölge, 12 ordu, 103 kara birimi, güç 1475
2. Kuzey Krallığı — 26 bölge, 8 ordu, 98 kara birimi, güç 1208
3. Tyrell — 10 bölge, 13 ordu, 164 kara birimi, güç 1144
4. Lannister — 9 bölge, 14 ordu, 133 kara birimi, güç 1269
5. Braavos — 6 bölge, 11 ordu, 150 kara birimi, güç 2071

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
- 39. tur: **Özgür Halkın Ardından** (`night_watch_after_free_folk_defeat`)
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
- 58. tur: **Kızıl Düğün'den Sonra Tullylerin Kararı** (`tully_after_red_wedding_299`)
- 58. tur: **Vadi Kapılarını Kapatıyor** (`vale_neutrality_after_red_wedding_299`)
- 60. tur: **Kuzeyde Kış Sertleşiyor** (`winter_storms_299`)
- 61. tur: **Buzun Altındaki Ordu** (`others_awaken_300`)
- 62. tur: **Demir Donanma Seferi** (`iron_fleet_campaign_300`)
- 62. tur: **Mor Düğün** (`purple_wedding_300`)
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
- 74. tur: **Taç Topraklarının Savaş Yorgunluğu** (`crownlands_postwar_exhaustion_300`)
- 74. tur: **Kralın Şehri Zaferinin Düzeni** (`postwar_settlement_baratheon_king`)
- 74. tur: **Reach'in Savaş Sonrası Konsolidasyonu** (`tyrell_postwar_consolidation_301`)
- 74. tur: **Beş Kralın Savaşı: Kralın Şehri Zaferi** (`war_of_five_kings_resolution_baratheon_king`)
- 77. tur: **Akgezenlere Karşı Son Hazırlık** (`night_watch_against_others_301`)
- 78. tur: **Duvar Savunmasının Bedeli** (`wall_breach_after_defense_301`)
- 114. tur: **Daenerys'in Meereen Seferi** (`daenerys_meereen_campaign_301`)

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
- 39. tur: `wall_defense_hardened`
- 39. tur: `wall_defense_prepared`
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
- 58. tur: `riverrun_resistance`
- 58. tur: `vale_refused_stark_aid`
- 60. tur: `long_winter_begun`
- 61. tur: `others_awakened`
- 62. tur: `iron_fleet_expanded`
- 62. tur: `purple_wedding_completed`
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
- 69. tur: `ironborn_war_tax_imposed`
- 69. tur: `tyrion_exiled_to_essos`
- 74. tur: `crownlands_postwar_exhausted`
- 74. tur: `postwar_crown_order`
- 74. tur: `tyrell_postwar_consolidation`
- 74. tur: `war_of_five_kings_resolved`
- 74. tur: `war_of_five_kings_settlement_complete`
- 74. tur: `war_of_five_kings_winner_baratheon_king`
- 77. tur: `boss_defeated:cmd_night_king`
- 78. tur: `wall_breached`
- 114. tur: `meereen_campaign_started`
- 114. tur: `targaryen_dragons_grown`
- 114. tur: `targaryen_return_ready`
