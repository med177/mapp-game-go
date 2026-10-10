# 298_war_of_five_kings — 120 Tur Simülasyon Raporu

- Tarih: 295/01 → 305/01
- Oyuncu devleti: Gece Nöbetçileri
- Seed: `298111`
- Zorluk: 3
- Karar politikası: Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.

## Nihai devlet durumu

| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Martell | 23 (+3) | 9 | 113 | 17 | 245 | 244 | 20→276 | +17 | 550 | 14323 | ayakta |
| Kuzey Krallığı | 18 (-1) | 10 | 91 | 1 | 155 | 139 | 29→175 | +11 | 534 | 14143 | ayakta |
| Tyrell | 12 (+0) | 12 | 100 | 40 | 231 | 276 | 18→278 | +14 | 205 | 60723 | ayakta |
| Arryn | 10 (+3) | 10 | 125 | 2 | 109 | 119 | 14→141 | +8 | 470 | 7271 | ayakta |
| Baratheon (Fırtına Burnu) | 10 (-1) | 4 | 32 | 0 | 4 | 25 | 13→29 | +2 | 0 | 521 | ayakta |
| Tully | 8 (+2) | 7 | 60 | 39 | 304 | 195 | 13→219 | +18 | 17806 | 57286 | ayakta |
| Braavos | 7 (+6) | 11 | 117 | 33 | 155 | 119 | 8→184 | +17 | 710 | 13604 | ayakta |
| Baratheon (Ejderha Kayası) | 6 (+3) | 9 | 52 | 26 | 392 | 81 | 16→174 | +17 | 4778 | 7612 | ayakta |
| Lannister | 5 (+0) | 10 | 135 | 23 | 110 | 130 | 16→144 | +17 | 20871 | 10077 | ayakta |
| Gece Nöbetçileri | 4 (+0) | 1 | 20 | 0 | 12 | 6 | 37→43 | +14 | 308 | 3360 | ayakta |
| Greyjoy | 3 (-2) | 2 | 38 | 0 | 63 | 23 | 11→23 | +8 | 260 | 568 | ayakta |
| Targaryen | 2 (+1) | 5 | 50 | 14 | 212 | 31 | 9→60 | +23 | 31688 | 8906 | ayakta |
| Akgezenler | 2 (+1) | 1 | 6 | 0 | 2 | 0 | 0→11 | +4 | 153 | 584 | ayakta |
| Bolton | 0 (-1) | 0 | 0 | 0 | 2 | 6 | 8→0 | +7 | 10 | 981 | ELENDİ |
| Özgür Halk | 0 (-1) | 0 | 0 | 0 | 0 | 6 | 5→0 | +2 | 127 | 124 | ELENDİ |
| Özgür Şehirler | 0 (-6) | 0 | 0 | 0 | 6 | 51 | 11→0 | +11 | 6849 | 4455 | ELENDİ |
| Kardeşlik | 0 (-1) | 0 | 0 | 0 | 0 | 6 | 1→0 | +4 | 74 | 890 | ELENDİ |
| Baratheon (Kralın Şehri) | 0 (-5) | 0 | 0 | 0 | 255 | 126 | 16→0 | +10 | 156 | 4519 | ELENDİ |

Not: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.

## Toprak değişimleri

- **Martell**: +Stoms End, Oldtown, Moat Cailin, Kraken Burnu; -Oldtown, Moat Cailin
- **Kuzey Krallığı**: +Taşlı Kıyı, Dehşet Kalesi, Tümülüs Toprakları, Moat Cailin, Flint Parmağı; -Boyun, Moat Cailin, Tümülüs Toprakları, Flint Parmağı, Deepwood Motte
- **Tyrell**: +Oldtown, Büyük Wyk; -Oldtown
- **Arryn**: +Boyun, Moat Cailin, Tümülüs Toprakları, Üç Dişli Mızrak, Deepwood Motte; -Moat Cailin, Tümülüs Toprakları, Üç Dişli Mızrak
- **Baratheon (Fırtına Burnu)**: +Taç Toprakları; -Taç Toprakları, Stoms End
- **Tully**: +Oyuk Tepe, Flint Parmağı, Üç Dişli Mızrak, Taç Toprakları, Harrenhal; -Üç Dişli Mızrak, Harrenhal, Flint Parmağı
- **Braavos**: +Norvos, Myr, Essos Free Cities, İhtilaflı Topraklar, Yunkai, Sharp Point; -—
- **Baratheon (Ejderha Kayası)**: +Yengeç Pençesi Burnu, Duskendale, Kralın Şehri, Kral Ormanı, Faces Island; -Yengeç Pençesi Burnu, Sharp Point
- **Greyjoy**: +Moat Cailin; -Moat Cailin, Kraken Burnu, Büyük Wyk
- **Targaryen**: +Meereen; -—
- **Akgezenler**: +Duvar'ın Ötesi; -—
- **Bolton**: +—; -Dehşet Kalesi
- **Özgür Halk**: +—; -Duvar'ın Ötesi
- **Özgür Şehirler**: +İhtilaflı Topraklar; -Astapor, Norvos, Myr, Essos Free Cities, İhtilaflı Topraklar, Yunkai, Meereen
- **Kardeşlik**: +—; -Oyuk Tepe
- **Baratheon (Kralın Şehri)**: +Taç Toprakları, Harrenhal, Yengeç Pençesi Burnu, Sharp Point; -Taç Toprakları, Harrenhal, Duskendale, Kralın Şehri, Sharp Point, Kral Ormanı, Faces Island

## Savaşlar

| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |
|---|---:|---:|---|---:|---:|
| Kardeşlik — Tully | 1 | 17. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 1 | 7. tur | aktif | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 2 | 6. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 12 | 19. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 13 | 25. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 24 | 28. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 39 | 43. tur | aktif | 0 / 0 | 0 / 0 |
| Arryn — Bolton | 41 | 59. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 41 | 59. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Bolton | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Kuzey Krallığı | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Greyjoy | 41 | 59. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Tully | 41 | 60. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Gece Nöbetçileri | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Kuzey Krallığı | 41 | 59. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tully | 41 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Akgezenler — Kuzey Krallığı | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Akgezenler — Tully | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 71 | 77. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tully | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Baratheon (Kralın Şehri) | 74 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Gece Nöbetçileri | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 74 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tully | 74 | 93. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 85 | 99. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Kuzey Krallığı | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Greyjoy | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Martell | 85 | 90. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Gece Nöbetçileri | 85 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Kuzey Krallığı | 85 | 90. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Tyrell | 85 | 91. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Braavos | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kuzey Krallığı | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Targaryen | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 105 | 107. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 107 | 112. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 117 | devam ediyor | bitti | 0 / 0 | 0 / 0 |

## Ticaret ve ittifak hareketleri

- 1. tur: **başladı** — Arryn ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Arryn ↔ Martell (grain)
- 1. tur: **başladı** — Arryn ↔ Tyrell (grain)
- 1. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Arryn (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Lannister (grain)
- 1. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Tully (grain)
- 1. tur: **başladı** — Braavos ↔ Lannister (spice)
- 1. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 1. tur: **başladı** — Lannister ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Lannister ↔ Braavos (grain)
- 1. tur: **başladı** — Martell ↔ Arryn (spice)
- 1. tur: **başladı** — Martell ↔ Tully (spice)
- 1. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (grain)
- 1. tur: **başladı** — Kuzey Krallığı ↔ Braavos (grain)
- 1. tur: **başladı** — Tully ↔ Baratheon (Fırtına Burnu) (grain)
- 1. tur: **başladı** — Tully ↔ Martell (grain)
- 1. tur: **başladı** — Tully ↔ Tyrell (grain)
- 1. tur: **başladı** — Tyrell ↔ Arryn (grain)
- 1. tur: **başladı** — Tyrell ↔ Tully (grain)
- 4. tur: **başladı** — Lannister ↔ Tyrell (iron)
- 4. tur: **başladı** — Tyrell ↔ Lannister (grain)
- 7. tur: **başladı** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 7. tur: **başladı** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 9. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Martell (grain)
- 9. tur: **başladı** — Martell ↔ Baratheon (Kralın Şehri) (spice)
- 23. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 23. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 26. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Özgür Şehirler (spice)
- 26. tur: **başladı** — Özgür Şehirler ↔ Baratheon (Ejderha Kayası) (spice)
- 28. tur: **başladı** — Braavos ↔ Targaryen (spice)
- 28. tur: **başladı** — Targaryen ↔ Braavos (spice)
- 29. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 29. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 35. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 35. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 36. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 36. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 38. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 38. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 40. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 40. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 41. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 41. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 41. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 41. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 41. tur: **bitti** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 41. tur: **bitti** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 44. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 44. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 45. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 45. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 45. tur: **başladı** — Özgür Şehirler ↔ Gece Nöbetçileri (grain)
- 45. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 45. tur: **başladı** — Gece Nöbetçileri ↔ Özgür Şehirler (cloth)
- 45. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 48. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 48. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 50. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 50. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 54. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 54. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 58. tur: **başladı** — Arryn ↔ Tyrell (iron)
- 58. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 58. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 58. tur: **başladı** — Tyrell ↔ Arryn (spice)
- 60. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 60. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 61. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 61. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 64. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 64. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 65. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 65. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (grain)
- 66. tur: **başladı** — Arryn ↔ Tyrell (iron)
- 66. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 66. tur: **başladı** — Gece Nöbetçileri ↔ Targaryen (cloth)
- 66. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (grain)
- 66. tur: **başladı** — Targaryen ↔ Gece Nöbetçileri (spice)
- 66. tur: **başladı** — Tyrell ↔ Arryn (spice)
- 70. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 70. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 71. tur: **başladı** — Arryn ↔ Tyrell (spice)
- 71. tur: **başladı** — Tyrell ↔ Arryn (spice)
- 72. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri (cloth)
- 72. tur: **başladı** — Martell ↔ Kuzey Krallığı (spice)
- 72. tur: **başladı** — Gece Nöbetçileri ↔ Baratheon (Fırtına Burnu) (cloth)
- 72. tur: **başladı** — Kuzey Krallığı ↔ Martell (iron)
- 73. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 73. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 73. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 73. tur: **başladı** — Martell ↔ Tully (spice)
- 73. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 73. tur: **başladı** — Tully ↔ Martell (grain)
- 75. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 75. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 75. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 75. tur: **başladı** — Martell ↔ Tully (spice)
- 75. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 75. tur: **başladı** — Tully ↔ Martell (grain)
- 76. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tyrell (cloth)
- 76. tur: **başladı** — Tyrell ↔ Baratheon (Kralın Şehri) (spice)
- 78. tur: **başladı** — Arryn ↔ Lannister (spice)
- 78. tur: **başladı** — Lannister ↔ Arryn (iron)
- 79. tur: **başladı** — Arryn ↔ Lannister (spice)
- 79. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 79. tur: **başladı** — Lannister ↔ Arryn (iron)
- 79. tur: **başladı** — Martell ↔ Tully (spice)
- 79. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (grain)
- 79. tur: **başladı** — Tully ↔ Martell (grain)
- 80. tur: **başladı** — Arryn ↔ Lannister (spice)
- 80. tur: **başladı** — Lannister ↔ Arryn (iron)
- 81. tur: **başladı** — Arryn ↔ Lannister (spice)
- 81. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 81. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 81. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 81. tur: **başladı** — Lannister ↔ Arryn (iron)
- 81. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (grain)
- 83. tur: **başladı** — Arryn ↔ Lannister (spice)
- 83. tur: **başladı** — Lannister ↔ Arryn (iron)
- 83. tur: **başladı** — Martell ↔ Tully (spice)
- 83. tur: **başladı** — Tully ↔ Martell (grain)
- 86. tur: **başladı** — Braavos ↔ Lannister (grain)
- 86. tur: **başladı** — Lannister ↔ Braavos (iron)
- 88. tur: **başladı** — Arryn ↔ Lannister (spice)
- 88. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 88. tur: **başladı** — Lannister ↔ Arryn (iron)
- 88. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (grain)
- 89. tur: **başladı** — Braavos ↔ Lannister (grain)
- 89. tur: **başladı** — Lannister ↔ Braavos (iron)
- 92. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (spice)
- 92. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 94. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 94. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 94. tur: **başladı** — Kuzey Krallığı ↔ Targaryen (iron)
- 94. tur: **başladı** — Targaryen ↔ Kuzey Krallığı (cloth)
- 99. tur: **başladı** — Gece Nöbetçileri ↔ Tully (cloth)
- 99. tur: **başladı** — Tully ↔ Gece Nöbetçileri (grain)
- 100. tur: **başladı** — Lannister ↔ Tyrell (cloth)
- 100. tur: **başladı** — Tyrell ↔ Lannister (spice)
- 102. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 102. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 105. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tyrell (cloth)
- 105. tur: **bitti** — Tyrell ↔ Baratheon (Kralın Şehri) (spice)
- 112. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (spice)
- 112. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (iron)
- 118. tur: **başladı** — Martell ↔ Tyrell (spice)
- 118. tur: **başladı** — Tyrell ↔ Martell (spice)
- 1. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 1. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 1. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 1. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 2. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 3. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tully
- 4. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 4. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 6. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 6. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 7. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 7. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 7. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 10. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 10. tur: **ittifak kuruldu** — Kardeşlik ↔ Tyrell
- 13. tur: **ittifak kuruldu** — Arryn ↔ Gece Nöbetçileri
- 14. tur: **ittifak bitti** — Arryn ↔ Kardeşlik
- 14. tur: **ittifak kuruldu** — Arryn ↔ Greyjoy
- 16. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 16. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 16. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 18. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tully
- 19. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 19. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 19. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 19. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 19. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 20. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 22. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 22. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 22. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Tully
- 25. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 25. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 25. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 25. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 25. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 27. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 28. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Ejderha Kayası)
- 29. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Tully
- 31. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 31. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 35. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Lannister
- 35. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Greyjoy
- 36. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Tyrell
- 37. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Lannister
- 37. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 37. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 38. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 41. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 42. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 43. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 46. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 47. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 48. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 50. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 51. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 53. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 56. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 58. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 58. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 58. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 61. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 62. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 63. tur: **ittifak kuruldu** — Braavos ↔ Gece Nöbetçileri
- 64. tur: **ittifak bitti** — Lannister ↔ Tully
- 68. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 68. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 71. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 71. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 71. tur: **ittifak kuruldu** — Braavos ↔ Kuzey Krallığı
- 72. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 72. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Lannister
- 73. tur: **ittifak kuruldu** — Martell ↔ Kuzey Krallığı
- 74. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 74. tur: **ittifak bitti** — Greyjoy ↔ Tully
- 74. tur: **ittifak bitti** — Lannister ↔ Tully
- 74. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 74. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 74. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 74. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tyrell
- 75. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 76. tur: **ittifak bitti** — Arryn ↔ Lannister
- 76. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 77. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 77. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 78. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 79. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Targaryen
- 85. tur: **ittifak kuruldu** — Martell ↔ Tyrell
- 90. tur: **ittifak bitti** — Martell ↔ Tyrell
- 91. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 94. tur: **ittifak bitti** — Lannister ↔ Tully
- 94. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Tully
- 95. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Targaryen
- 96. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 99. tur: **ittifak kuruldu** — Arryn ↔ Tyrell
- 99. tur: **ittifak kuruldu** — Martell ↔ Tully
- 100. tur: **ittifak kuruldu** — Arryn ↔ Tully
- 102. tur: **ittifak bitti** — Arryn ↔ Tully
- 103. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 103. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 105. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 105. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 105. tur: **ittifak bitti** — Lannister ↔ Tully
- 105. tur: **ittifak bitti** — Martell ↔ Kuzey Krallığı
- 105. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Targaryen
- 108. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 114. tur: **ittifak bitti** — Lannister ↔ Tully
- 116. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 116. tur: **ittifak kuruldu** — Tully ↔ Tyrell
- 117. tur: **ittifak kuruldu** — Lannister ↔ Tully

## Elenen devletler

- 17. tur: **Kardeşlik**
- 25. tur: **Özgür Halk**
- 61. tur: **Özgür Şehirler**
- 82. tur: **Bolton**
- 107. tur: **Baratheon (Kralın Şehri)**

## Tur checkpoint'leri

### 9. tur (295/10)

1. Martell — 20 bölge, 3 ordu, 32 kara birimi, güç 383
2. Stark — 19 bölge, 4 ordu, 52 kara birimi, güç 641
3. Tyrell — 12 bölge, 2 ordu, 31 kara birimi, güç 564
4. Baratheon (Fırtına Burnu) — 11 bölge, 2 ordu, 24 kara birimi, güç 246
5. Arryn — 7 bölge, 2 ordu, 20 kara birimi, güç 153

### 19. tur (296/08)

1. Martell — 21 bölge, 2 ordu, 35 kara birimi, güç 373
2. Stark — 19 bölge, 5 ordu, 78 kara birimi, güç 753
3. Tyrell — 12 bölge, 2 ordu, 36 kara birimi, güç 639
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 27 kara birimi, güç 380
5. Arryn — 7 bölge, 3 ordu, 21 kara birimi, güç 304

### 29. tur (297/06)

1. Martell — 21 bölge, 4 ordu, 38 kara birimi, güç 388
2. Stark — 19 bölge, 5 ordu, 79 kara birimi, güç 810
3. Tyrell — 12 bölge, 5 ordu, 43 kara birimi, güç 746
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 31 kara birimi, güç 288
5. Arryn — 7 bölge, 2 ordu, 20 kara birimi, güç 289

### 39. tur (298/04)

1. Martell — 22 bölge, 6 ordu, 41 kara birimi, güç 504
2. Stark — 19 bölge, 8 ordu, 103 kara birimi, güç 801
3. Tyrell — 11 bölge, 8 ordu, 45 kara birimi, güç 731
4. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 34 kara birimi, güç 230
5. Arryn — 7 bölge, 2 ordu, 20 kara birimi, güç 274

### 49. tur (299/02)

1. Martell — 22 bölge, 7 ordu, 42 kara birimi, güç 582
2. Kuzey Krallığı — 19 bölge, 6 ordu, 73 kara birimi, güç 695
3. Tyrell — 11 bölge, 5 ordu, 48 kara birimi, güç 761
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 34 kara birimi, güç 267
5. Arryn — 8 bölge, 5 ordu, 47 kara birimi, güç 412

### 59. tur (299/12)

1. Martell — 22 bölge, 7 ordu, 42 kara birimi, güç 779
2. Kuzey Krallığı — 17 bölge, 6 ordu, 74 kara birimi, güç 1011
3. Tyrell — 11 bölge, 6 ordu, 50 kara birimi, güç 1031
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 34 kara birimi, güç 407
5. Arryn — 9 bölge, 4 ordu, 51 kara birimi, güç 862

### 69. tur (300/10)

1. Martell — 22 bölge, 7 ordu, 41 kara birimi, güç 812
2. Kuzey Krallığı — 16 bölge, 4 ordu, 61 kara birimi, güç 847
3. Tyrell — 11 bölge, 8 ordu, 59 kara birimi, güç 1474
4. Baratheon (Fırtına Burnu) — 10 bölge, 1 ordu, 14 kara birimi, güç 105
5. Arryn — 9 bölge, 4 ordu, 60 kara birimi, güç 1001

### 79. tur (301/08)

1. Martell — 21 bölge, 6 ordu, 49 kara birimi, güç 831
2. Kuzey Krallığı — 16 bölge, 7 ordu, 89 kara birimi, güç 1056
3. Tyrell — 11 bölge, 7 ordu, 61 kara birimi, güç 1736
4. Arryn — 10 bölge, 5 ordu, 69 kara birimi, güç 1103
5. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 22 kara birimi, güç 274

### 89. tur (302/06)

1. Martell — 22 bölge, 7 ordu, 61 kara birimi, güç 1078
2. Kuzey Krallığı — 18 bölge, 10 ordu, 102 kara birimi, güç 1218
3. Tyrell — 11 bölge, 7 ordu, 71 kara birimi, güç 1589
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 26 kara birimi, güç 362
5. Arryn — 9 bölge, 5 ordu, 75 kara birimi, güç 1298

### 99. tur (303/04)

1. Martell — 22 bölge, 10 ordu, 74 kara birimi, güç 1135
2. Kuzey Krallığı — 18 bölge, 2 ordu, 17 kara birimi, güç 120
3. Tyrell — 12 bölge, 7 ordu, 78 kara birimi, güç 1734
4. Arryn — 11 bölge, 6 ordu, 103 kara birimi, güç 1373
5. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 29 kara birimi, güç 444

### 109. tur (304/02)

1. Martell — 22 bölge, 11 ordu, 96 kara birimi, güç 1263
2. Kuzey Krallığı — 18 bölge, 4 ordu, 45 kara birimi, güç 568
3. Tyrell — 12 bölge, 9 ordu, 95 kara birimi, güç 1742
4. Arryn — 10 bölge, 7 ordu, 112 kara birimi, güç 1478
5. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 32 kara birimi, güç 297

### 119. tur (304/12)

1. Martell — 22 bölge, 10 ordu, 113 kara birimi, güç 1930
2. Kuzey Krallığı — 18 bölge, 11 ordu, 82 kara birimi, güç 1261
3. Tyrell — 12 bölge, 11 ordu, 104 kara birimi, güç 1917
4. Arryn — 10 bölge, 10 ordu, 122 kara birimi, güç 1540
5. Baratheon (Fırtına Burnu) — 10 bölge, 4 ordu, 32 kara birimi, güç 389

### 120. tur (305/01)

1. Martell — 23 bölge, 9 ordu, 113 kara birimi, güç 1712
2. Kuzey Krallığı — 18 bölge, 10 ordu, 91 kara birimi, güç 1113
3. Tyrell — 12 bölge, 12 ordu, 100 kara birimi, güç 1546
4. Arryn — 10 bölge, 10 ordu, 125 kara birimi, güç 1540
5. Baratheon (Fırtına Burnu) — 10 bölge, 4 ordu, 32 kara birimi, güç 325

## Tetiklenen olaylar

- 6. tur: **Kraliyet Borçlarının Gölgesi** (`crown_debt_crisis_295`)
- 14. tur: **Kuzeyin Kış Hazırlıkları** (`northern_preparations_296`)
- 20. tur: **Lannister Altını ve Tacın Borçları** (`lannister_gold_and_crown_296`)
- 27. tur: **Ejderha Kayası'nda Seferberlik** (`dragonstone_muster_297`)
- 37. tur: **Jon Arryn'in Ölümü ve Şüpheler** (`jon_arryn_investigation_298`)
- 38. tur: **Catelyn Tyrion'u Yakalıyor** (`catelyn_captures_tyrion_298`)
- 39. tur: **Kral Robert'ın Ölümü** (`death_of_robert_298`)
- 40. tur: **Özgür Halkın Ardından** (`night_watch_after_free_folk_defeat`)
- 41. tur: **Vadi'nin Siyasi Tutumu** (`vale_political_stance_298`)
- 43. tur: **Nehir Toprakları Savaşının Yükü** (`lannister_riverlands_war_burden_298`)
- 46. tur: **Kuzeyden Gelen Alametler** (`winter_omens_298`)
- 49. tur: **Özgür Şehirlerin Tarafsızlığı** (`free_cities_nonintervention_299`)
- 52. tur: **Robb'un Kuzey Kralı İlanı** (`robb_king_in_the_north_299`)
- 55. tur: **Kara Su Savaşı** (`battle_of_the_blackwater_299`)
- 56. tur: **Greyjoy Saldırısı** (`greyjoy_rebellion_299`)
- 57. tur: **Edmure ve Roslin'in Evliliği** (`edmure_roslin_marriage_299`)
- 58. tur: **Margaery ve Joffrey'nin Evliliği** (`tyrell_joffrey_marriage_299`)
- 59. tur: **Kızıl Düğün İhaneti** (`red_wedding_event`)
- 60. tur: **Kızıl Düğün'den Sonra Tullylerin Kararı** (`tully_after_red_wedding_299`)
- 61. tur: **Kuzeyde Kış Sertleşiyor** (`winter_storms_299`)
- 62. tur: **Buzun Altındaki Ordu** (`others_awaken_300`)
- 65. tur: **Lannister Savaş Yıpranması** (`lannister_war_exhaustion_300`)
- 66. tur: **Vadi'de Veraset Krizi** (`vale_succession_crisis_300`)
- 68. tur: **Nehir Toprakları'nın Savaş Bedeli** (`riverlands_war_burden_300`)
- 69. tur: **Gece Nöbeti'nin Duvar Hazırlıkları** (`night_watch_wall_preparations_300`)
- 77. tur: **Akgezenlere Karşı Son Hazırlık** (`night_watch_against_others_301`)
- 114. tur: **Daenerys'in Meereen Seferi** (`daenerys_meereen_campaign_301`)
