# 298_war_of_five_kings — 120 Tur Simülasyon Raporu

- Tarih: 295/01 → 305/01
- Oyuncu devleti: Gece Nöbetçileri
- Seed: `298114`
- Zorluk: 3
- Karar politikası: Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.

## Nihai devlet durumu

| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Kuzey Krallığı | 28 (+9) | 14 | 240 | 1 | 70 | 65 | 29→166 | +8 | 0 | 1545 | ayakta |
| Martell | 23 (+3) | 11 | 109 | 20 | 251 | 216 | 20→252 | +16 | 504 | 38441 | ayakta |
| Tyrell | 16 (+4) | 13 | 171 | 21 | 257 | 288 | 18→342 | +15 | 971 | 65128 | ayakta |
| Lannister | 11 (+6) | 26 | 192 | 37 | 339 | 188 | 16→301 | +19 | 3340 | 32635 | ayakta |
| Baratheon (Fırtına Burnu) | 9 (-2) | 4 | 18 | 2 | 5 | 31 | 13→28 | +11 | 1 | 2420 | ayakta |
| Braavos | 8 (+7) | 12 | 67 | 45 | 404 | 139 | 8→227 | +19 | 14271 | 31090 | ayakta |
| Greyjoy | 4 (-1) | 3 | 20 | 2 | 56 | 16 | 11→24 | +5 | 189 | 468 | ayakta |
| Gece Nöbetçileri | 4 (+0) | 1 | 15 | 0 | 19 | 12 | 37→49 | +18 | 173 | 5808 | ayakta |
| Baratheon (Kralın Şehri) | 2 (-3) | 2 | 20 | 2 | 99 | 91 | 16→35 | +12 | 0 | 804 | ayakta |
| Akgezenler | 2 (+1) | 1 | 17 | 0 | 0 | 1 | 0→12 | +5 | 0 | 593 | ayakta |
| Baratheon (Ejderha Kayası) | 1 (-2) | 5 | 26 | 7 | 132 | 42 | 16→30 | +17 | 4002 | 710 | ayakta |
| Arryn | 1 (-6) | 1 | 3 | 0 | 15 | 30 | 14→5 | +3 | 569 | 670 | ELENDİ |
| Tully | 1 (-5) | 3 | 0 | 19 | 81 | 141 | 13→12 | +18 | 162 | 300 | ayakta |
| Bolton | 0 (-1) | 0 | 0 | 0 | 2 | 5 | 8→0 | +2 | 33 | 1052 | ELENDİ |
| Özgür Halk | 0 (-1) | 0 | 0 | 0 | 1 | 6 | 5→0 | +2 | 168 | 352 | ELENDİ |
| Targaryen | 0 (-1) | 0 | 0 | 0 | 18 | 20 | 9→0 | +11 | 16558 | 843 | ELENDİ |
| Özgür Şehirler | 0 (-6) | 0 | 0 | 0 | 2 | 46 | 11→0 | +11 | 125 | 2251 | ELENDİ |
| Kardeşlik | 0 (-1) | 0 | 0 | 0 | 0 | 4 | 1→0 | +3 | 58 | 1008 | ELENDİ |

Not: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.

## Toprak değişimleri

- **Kuzey Krallığı**: +Taşlı Kıyı, Flint Parmağı, Dehşet Kalesi, Üç Dişli Mızrak, Kraken Burnu, Batı Vadi, Parmaklar, Ay Dağları, Kartal Yuvası, Tuz Tavaları; -Flint Parmağı, Tuz Tavaları
- **Martell**: +Stoms End, Oldtown, Yengeç Pençesi Burnu, Kralın Şehri; -Oldtown, Stoms End, Kralın Şehri
- **Tyrell**: +Oldtown, Taç Toprakları, Kral Ormanı, Sharp Point, Duskendale, Kralın Şehri, Tuz Tavaları, Wickenden, Parlaksu Diyarı, Sis Ormanı; -Oldtown, Taç Toprakları, Parlaksu Diyarı, Tuz Tavaları, Kralın Şehri
- **Lannister**: +Harrenhal, Oyuk Tepe, Flint Parmağı, Taş Septi, The Crag, Nehir Toprakları, Nehir Koşusu, İkizler; -Harrenhal, Oyuk Tepe, Flint Parmağı, Taş Septi
- **Baratheon (Fırtına Burnu)**: +Taç Toprakları, Stoms End, Fırtına Diyarı, Gazap Burnu; -Stoms End, Taç Toprakları, Fırtına Diyarı, Gazap Burnu, Sis Ormanı
- **Braavos**: +İhtilaflı Topraklar, Norvos, Essos Free Cities, Myr, Meereen, Yunkai, Pentos (Özgür Şehir); -—
- **Greyjoy**: +Kraken Burnu; -Kraken Burnu
- **Baratheon (Kralın Şehri)**: +Gazap Burnu, Sis Ormanı, Üç Dişli Mızrak, Harrenhal, Parlaksu Diyarı, Taç Toprakları; -Taç Toprakları, Duskendale, Kral Ormanı, Kralın Şehri, Gazap Burnu, Üç Dişli Mızrak, Harrenhal, Parlaksu Diyarı, Sis Ormanı
- **Akgezenler**: +Duvar'ın Ötesi; -—
- **Baratheon (Ejderha Kayası)**: +Taç Toprakları, Duskendale, Kral Ormanı, Fırtına Diyarı, Kralın Şehri, Gazap Burnu; -Taç Toprakları, Fırtına Diyarı, Kral Ormanı, Sharp Point, Yengeç Pençesi Burnu, Gazap Burnu, Duskendale, Kralın Şehri
- **Arryn**: +Wickenden; -Batı Vadi, Martı Kasabası, Parmaklar, Ay Dağları, Tuz Tavaları, Wickenden, Kartal Yuvası
- **Tully**: +Oyuk Tepe, Harrenhal, Taş Septi, Martı Kasabası; -Harrenhal, Oyuk Tepe, Üç Dişli Mızrak, Taş Septi, The Crag, Nehir Toprakları, Nehir Koşusu, İkizler
- **Bolton**: +—; -Dehşet Kalesi
- **Özgür Halk**: +—; -Duvar'ın Ötesi
- **Targaryen**: +—; -Pentos (Özgür Şehir)
- **Özgür Şehirler**: +—; -Astapor, Norvos, Essos Free Cities, Myr, Meereen, Yunkai
- **Kardeşlik**: +—; -Oyuk Tepe

## Savaşlar

| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |
|---|---:|---:|---|---:|---:|
| Kardeşlik — Tully | 1 | 17. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 2 | 7. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 7 | 11. tur | aktif | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 9 | 13. tur | aktif | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 12 | 28. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Gece Nöbetçileri | 19 | 23. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 21 | 33. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Gece Nöbetçileri | 29 | 33. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Bolton | 41 | 58. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 41 | 101. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Greyjoy | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Greyjoy | 41 | 58. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Martell | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Tully | 41 | 58. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Gece Nöbetçileri | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Kuzey Krallığı | 41 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Kuzey Krallığı | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tully | 41 | 52. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 48 | 50. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Targaryen | 48 | 51. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Greyjoy | 51 | 55. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Greyjoy | 51 | 55. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Lannister | 51 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Martell | 51 | 55. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Tyrell | 51 | 55. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 51 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Fırtına Burnu) | 60 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 60 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 60 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 60 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Lannister | 60 | 64. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tully | 70 | 89. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Greyjoy | 70 | 74. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Gece Nöbetçileri | 70 | 74. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 70 | 83. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Lannister | 74 | 81. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Martell | 74 | 81. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Lannister | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Gece Nöbetçileri | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 74 | 96. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Tully | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Tully — Tyrell | 74 | 78. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Tyrell | 82 | 94. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Baratheon (Fırtına Burnu) | 82 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Kuzey Krallığı | 82 | 86. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 82 | 86. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Tyrell | 97 | 101. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tyrell | 97 | 118. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Akgezenler | 99 | 101. tur | bitti | 0 / 0 | 0 / 0 |
| Martell — Akgezenler | 99 | 103. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 99 | devam ediyor | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tyrell | 102 | 119. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Tyrell | 107 | 111. tur | bitti | 0 / 0 | 0 / 0 |

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
- 1. tur: **başladı** — Kuzey Krallığı ↔ Braavos (grain)
- 1. tur: **başladı** — Tully ↔ Martell (grain)
- 1. tur: **başladı** — Tully ↔ Tyrell (grain)
- 1. tur: **başladı** — Tyrell ↔ Arryn (grain)
- 1. tur: **başladı** — Tyrell ↔ Tully (grain)
- 2. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (grain)
- 2. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tully (grain)
- 2. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (grain)
- 2. tur: **başladı** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 9. tur: **başladı** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 9. tur: **başladı** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 10. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (grain)
- 10. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 23. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 23. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 26. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Greyjoy (iron)
- 26. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Gece Nöbetçileri (iron)
- 26. tur: **başladı** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 26. tur: **başladı** — Gece Nöbetçileri ↔ Baratheon (Ejderha Kayası) (timber)
- 41. tur: **bitti** — Baratheon (Ejderha Kayası) ↔ Greyjoy (iron)
- 41. tur: **bitti** — Greyjoy ↔ Baratheon (Ejderha Kayası) (iron)
- 41. tur: **bitti** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 41. tur: **bitti** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 43. tur: **başladı** — Braavos ↔ Targaryen (spice)
- 43. tur: **başladı** — Targaryen ↔ Braavos (spice)
- 48. tur: **bitti** — Braavos ↔ Targaryen (spice)
- 48. tur: **bitti** — Targaryen ↔ Braavos (spice)
- 51. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Greyjoy (grain)
- 51. tur: **bitti** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 54. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Tyrell (spice)
- 54. tur: **başladı** — Tyrell ↔ Baratheon (Kralın Şehri) (spice)
- 56. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 56. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 59. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 59. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 60. tur: **bitti** — Arryn ↔ Baratheon (Fırtına Burnu) (grain)
- 60. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tully (grain)
- 60. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Tyrell (spice)
- 60. tur: **bitti** — Baratheon (Fırtına Burnu) ↔ Arryn (grain)
- 60. tur: **bitti** — Baratheon (Fırtına Burnu) ↔ Lannister (grain)
- 60. tur: **bitti** — Lannister ↔ Baratheon (Fırtına Burnu) (grain)
- 60. tur: **bitti** — Tully ↔ Baratheon (Kralın Şehri) (grain)
- 60. tur: **bitti** — Tyrell ↔ Baratheon (Kralın Şehri) (spice)
- 61. tur: **başladı** — Lannister ↔ Gece Nöbetçileri (iron)
- 61. tur: **başladı** — Gece Nöbetçileri ↔ Lannister (iron)
- 62. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Lannister (iron)
- 62. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Gece Nöbetçileri (spice)
- 62. tur: **başladı** — Lannister ↔ Baratheon (Ejderha Kayası) (iron)
- 62. tur: **başladı** — Gece Nöbetçileri ↔ Baratheon (Kralın Şehri) (iron)
- 64. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Lannister (spice)
- 64. tur: **başladı** — Lannister ↔ Baratheon (Kralın Şehri) (iron)
- 68. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (grain)
- 68. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (timber)
- 69. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (grain)
- 69. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (timber)
- 70. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (grain)
- 70. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (timber)
- 70. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Gece Nöbetçileri (spice)
- 70. tur: **bitti** — Gece Nöbetçileri ↔ Baratheon (Kralın Şehri) (iron)
- 71. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (grain)
- 71. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Tully (spice)
- 71. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (timber)
- 71. tur: **başladı** — Tully ↔ Baratheon (Fırtına Burnu) (grain)
- 74. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı (grain)
- 74. tur: **başladı** — Kuzey Krallığı ↔ Baratheon (Ejderha Kayası) (timber)
- 74. tur: **bitti** — Lannister ↔ Gece Nöbetçileri (iron)
- 74. tur: **bitti** — Martell ↔ Tully (spice)
- 74. tur: **bitti** — Gece Nöbetçileri ↔ Lannister (iron)
- 74. tur: **bitti** — Tully ↔ Martell (grain)
- 74. tur: **bitti** — Tully ↔ Tyrell (grain)
- 74. tur: **bitti** — Tyrell ↔ Tully (grain)
- 78. tur: **başladı** — Gece Nöbetçileri ↔ Tyrell (iron)
- 78. tur: **başladı** — Tyrell ↔ Gece Nöbetçileri (spice)
- 81. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu) (spice)
- 81. tur: **başladı** — Baratheon (Fırtına Burnu) ↔ Baratheon (Ejderha Kayası) (spice)
- 82. tur: **başladı** — Lannister ↔ Tyrell (spice)
- 82. tur: **başladı** — Tyrell ↔ Lannister (grain)
- 82. tur: **bitti** — Arryn ↔ Tyrell (grain)
- 82. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı (spice)
- 82. tur: **bitti** — Kuzey Krallığı ↔ Baratheon (Kralın Şehri) (timber)
- 82. tur: **bitti** — Tyrell ↔ Arryn (grain)
- 83. tur: **başladı** — Braavos ↔ Gece Nöbetçileri (spice)
- 83. tur: **başladı** — Gece Nöbetçileri ↔ Braavos (iron)
- 96. tur: **başladı** — Lannister ↔ Tully (spice)
- 96. tur: **başladı** — Tully ↔ Lannister (grain)
- 114. tur: **başladı** — Greyjoy ↔ Tyrell (iron)
- 114. tur: **başladı** — Tyrell ↔ Greyjoy (spice)
- 115. tur: **başladı** — Greyjoy ↔ Tyrell (iron)
- 115. tur: **başladı** — Tyrell ↔ Greyjoy (spice)
- 1. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 1. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 2. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 2. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 2. tur: **ittifak kuruldu** — Kardeşlik ↔ Tyrell
- 2. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 3. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Kardeşlik
- 6. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 9. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 9. tur: **ittifak kuruldu** — Arryn ↔ Gece Nöbetçileri
- 9. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 10. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 15. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 16. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 16. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 16. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 17. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 18. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 19. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 19. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 19. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 20. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 20. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Tully
- 21. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 24. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Tully
- 25. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 25. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 25. tur: **ittifak kuruldu** — Arryn ↔ Greyjoy
- 25. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 27. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 27. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 28. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 28. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 32. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 33. tur: **ittifak bitti** — Arryn ↔ Lannister
- 34. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 34. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı
- 38. tur: **ittifak bitti** — Arryn ↔ Lannister
- 39. tur: **ittifak bitti** — Arryn ↔ Tully
- 40. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 40. tur: **ittifak kuruldu** — Arryn ↔ Kuzey Krallığı
- 40. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak kuruldu** — Greyjoy ↔ Martell
- 41. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 43. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 45. tur: **ittifak kuruldu** — Martell ↔ Gece Nöbetçileri
- 46. tur: **ittifak bitti** — Arryn ↔ Lannister
- 46. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 46. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 47. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 49. tur: **ittifak bitti** — Lannister ↔ Tully
- 49. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 49. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 51. tur: **ittifak bitti** — Arryn ↔ Greyjoy
- 51. tur: **ittifak bitti** — Greyjoy ↔ Martell
- 51. tur: **ittifak kuruldu** — Lannister ↔ Tyrell
- 53. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 55. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 55. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tyrell
- 57. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 58. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 58. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 59. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 63. tur: **ittifak bitti** — Arryn ↔ Lannister
- 63. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 64. tur: **ittifak kuruldu** — Arryn ↔ Tully
- 65. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 67. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tully
- 68. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 68. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı
- 70. tur: **ittifak bitti** — Arryn ↔ Tully
- 70. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Tully
- 70. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tully
- 74. tur: **ittifak kuruldu** — Martell ↔ Tyrell
- 81. tur: **ittifak bitti** — Lannister ↔ Martell
- 81. tur: **ittifak bitti** — Lannister ↔ Tyrell
- 81. tur: **ittifak bitti** — Martell ↔ Tyrell
- 82. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 82. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 83. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 84. tur: **ittifak bitti** — Arryn ↔ Lannister
- 84. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 85. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 86. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 87. tur: **ittifak bitti** — Greyjoy ↔ Tully
- 91. tur: **ittifak kuruldu** — Braavos ↔ Greyjoy
- 95. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı
- 95. tur: **ittifak kuruldu** — Greyjoy ↔ Tully
- 96. tur: **ittifak bitti** — Greyjoy ↔ Tully
- 96. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 96. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Tully
- 96. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 96. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 97. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tyrell
- 109. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 111. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Lannister
- 118. tur: **ittifak bitti** — Lannister ↔ Martell
- 118. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 119. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 119. tur: **ittifak kuruldu** — Lannister ↔ Martell

## Elenen devletler

- 17. tur: **Kardeşlik**
- 33. tur: **Özgür Halk**
- 50. tur: **Özgür Şehirler**
- 51. tur: **Targaryen**
- 69. tur: **Bolton**
- 101. tur: **Arryn**

## Tur checkpoint'leri

### 9. tur (295/10)

1. Martell — 20 bölge, 3 ordu, 32 kara birimi, güç 305
2. Stark — 19 bölge, 3 ordu, 51 kara birimi, güç 672
3. Tyrell — 12 bölge, 2 ordu, 27 kara birimi, güç 442
4. Baratheon (Fırtına Burnu) — 12 bölge, 2 ordu, 14 kara birimi, güç 116
5. Arryn — 7 bölge, 2 ordu, 38 kara birimi, güç 524

### 19. tur (296/08)

1. Martell — 21 bölge, 3 ordu, 39 kara birimi, güç 281
2. Stark — 19 bölge, 5 ordu, 74 kara birimi, güç 632
3. Tyrell — 12 bölge, 2 ordu, 34 kara birimi, güç 547
4. Baratheon (Fırtına Burnu) — 11 bölge, 1 ordu, 14 kara birimi, güç 207
5. Arryn — 7 bölge, 2 ordu, 38 kara birimi, güç 374

### 29. tur (297/06)

1. Martell — 21 bölge, 3 ordu, 39 kara birimi, güç 413
2. Stark — 19 bölge, 7 ordu, 102 kara birimi, güç 894
3. Tyrell — 12 bölge, 6 ordu, 43 kara birimi, güç 672
4. Baratheon (Fırtına Burnu) — 11 bölge, 1 ordu, 14 kara birimi, güç 210
5. Arryn — 7 bölge, 2 ordu, 38 kara birimi, güç 385

### 39. tur (298/04)

1. Martell — 21 bölge, 8 ordu, 42 kara birimi, güç 551
2. Stark — 19 bölge, 7 ordu, 131 kara birimi, güç 1049
3. Baratheon (Fırtına Burnu) — 12 bölge, 5 ordu, 17 kara birimi, güç 232
4. Tyrell — 11 bölge, 7 ordu, 47 kara birimi, güç 653
5. Arryn — 7 bölge, 1 ordu, 18 kara birimi, güç 33

### 49. tur (299/02)

1. Martell — 22 bölge, 7 ordu, 46 kara birimi, güç 572
2. Kuzey Krallığı — 19 bölge, 8 ordu, 136 kara birimi, güç 979
3. Tyrell — 11 bölge, 5 ordu, 57 kara birimi, güç 646
4. Baratheon (Fırtına Burnu) — 11 bölge, 3 ordu, 17 kara birimi, güç 231
5. Tully — 7 bölge, 6 ordu, 45 kara birimi, güç 511

### 59. tur (299/12)

1. Martell — 22 bölge, 8 ordu, 51 kara birimi, güç 808
2. Kuzey Krallığı — 19 bölge, 9 ordu, 145 kara birimi, güç 1502
3. Tyrell — 11 bölge, 9 ordu, 43 kara birimi, güç 844
4. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 17 kara birimi, güç 249
5. Braavos — 8 bölge, 7 ordu, 38 kara birimi, güç 1457

### 69. tur (300/10)

1. Martell — 22 bölge, 7 ordu, 57 kara birimi, güç 899
2. Kuzey Krallığı — 21 bölge, 10 ordu, 165 kara birimi, güç 1114
3. Tyrell — 11 bölge, 7 ordu, 42 kara birimi, güç 823
4. Baratheon (Fırtına Burnu) — 9 bölge, 3 ordu, 17 kara birimi, güç 221
5. Braavos — 8 bölge, 10 ordu, 45 kara birimi, güç 1647

### 79. tur (301/08)

1. Martell — 23 bölge, 8 ordu, 63 kara birimi, güç 805
2. Kuzey Krallığı — 21 bölge, 7 ordu, 106 kara birimi, güç 894
3. Tyrell — 13 bölge, 8 ordu, 77 kara birimi, güç 1128
4. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 17 kara birimi, güç 236
5. Braavos — 8 bölge, 10 ordu, 46 kara birimi, güç 1864

### 89. tur (302/06)

1. Kuzey Krallığı — 24 bölge, 7 ordu, 122 kara birimi, güç 1504
2. Martell — 23 bölge, 6 ordu, 48 kara birimi, güç 365
3. Tyrell — 14 bölge, 11 ordu, 92 kara birimi, güç 1545
4. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 17 kara birimi, güç 231
5. Braavos — 8 bölge, 9 ordu, 48 kara birimi, güç 1973

### 99. tur (303/04)

1. Kuzey Krallığı — 26 bölge, 11 ordu, 167 kara birimi, güç 2265
2. Martell — 23 bölge, 8 ordu, 65 kara birimi, güç 761
3. Tyrell — 16 bölge, 12 ordu, 109 kara birimi, güç 1697
4. Lannister — 11 bölge, 12 ordu, 135 kara birimi, güç 2322
5. Baratheon (Fırtına Burnu) — 9 bölge, 4 ordu, 18 kara birimi, güç 243

### 109. tur (304/02)

1. Kuzey Krallığı — 28 bölge, 15 ordu, 208 kara birimi, güç 2424
2. Martell — 23 bölge, 9 ordu, 86 kara birimi, güç 856
3. Tyrell — 14 bölge, 12 ordu, 141 kara birimi, güç 1661
4. Lannister — 11 bölge, 19 ordu, 210 kara birimi, güç 2775
5. Baratheon (Fırtına Burnu) — 10 bölge, 4 ordu, 18 kara birimi, güç 254

### 119. tur (304/12)

1. Kuzey Krallığı — 28 bölge, 15 ordu, 240 kara birimi, güç 4226
2. Martell — 23 bölge, 10 ordu, 104 kara birimi, güç 1273
3. Tyrell — 16 bölge, 12 ordu, 170 kara birimi, güç 2547
4. Lannister — 11 bölge, 23 ordu, 240 kara birimi, güç 3093
5. Baratheon (Fırtına Burnu) — 9 bölge, 4 ordu, 18 kara birimi, güç 315

### 120. tur (305/01)

1. Kuzey Krallığı — 28 bölge, 14 ordu, 240 kara birimi, güç 3617
2. Martell — 23 bölge, 11 ordu, 109 kara birimi, güç 1163
3. Tyrell — 16 bölge, 13 ordu, 171 kara birimi, güç 2176
4. Lannister — 11 bölge, 26 ordu, 192 kara birimi, güç 2877
5. Baratheon (Fırtına Burnu) — 9 bölge, 4 ordu, 18 kara birimi, güç 275

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
- 49. tur: **Özgür Şehirlerin Tarafsızlığı** (`free_cities_nonintervention_299`)
- 52. tur: **Riverrun Savunması** (`riverrun_defense_299`)
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
- 68. tur: **Nehir Toprakları'nın Savaş Bedeli** (`riverlands_war_burden_300`)
- 68. tur: **Batı İttifakının Çatlağı** (`tyrell_lannister_tension_300`)
- 68. tur: **Tywin Lannister'ın Ölümü** (`tywin_death_300`)
- 69. tur: **Demir Adalar'ın İkmal Krizi** (`ironborn_supply_crisis_300`)
- 69. tur: **Tyrion'un Kaderi** (`tyrion_fate_300`)
- 77. tur: **Akgezenlere Karşı Son Hazırlık** (`night_watch_against_others_301`)
- 78. tur: **Duvar Savunmasının Bedeli** (`wall_breach_after_defense_301`)

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
- 49. tur: `free_cities_nonintervention`
- 52. tur: `king_in_the_north`
- 52. tur: `northern_independence_declared`
- 52. tur: `riverrun_defense_prepared`
- 52. tur: `riverrun_siege_started`
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
- 68. tur: `riverlands_rationed`
- 68. tur: `tyrell_alliance_broken`
- 68. tur: `tyrell_alliance_preserved`
- 68. tur: `tywin_dead`
- 69. tur: `ironborn_war_tax_imposed`
- 69. tur: `tyrion_exiled_to_essos`
- 78. tur: `wall_breached`
