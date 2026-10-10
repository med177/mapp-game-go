# 298_war_of_five_kings — 100 Tur Simülasyon Raporu

- Tarih: 295/01 → 303/05
- Oyuncu devleti: Gece Nöbetçileri
- Seed: `298111`
- Zorluk: 3
- Karar politikası: Tüm devletler AI; AI, muharebe ve rastgele olay zarları aynı seed ile başlatılır; oyuncuya gelen barış/ittifak/ticaret/savaş çağrıları kabul edilir, teslimiyet ve kuşatma vassallığı teklifleri reddedilir; olay seçeneklerinde senaryo otomatik seçimi kullanılır.

## Nihai devlet durumu

| Devlet | Bölge | Ordu | Kara birimi | Donanma birimi | Üretim emri | Bina emri | Bina | Teknoloji | Altın | Tahıl | Durum |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Kuzey Krallığı | 23 (+4) | 13 | 220 | 0 | 166 | 135 | 29→180 | +11 | 842 | 18269 | ayakta |
| Martell | 22 (+2) | 6 | 64 | 13 | 209 | 219 | 20→251 | +15 | 963 | 15175 | ayakta |
| Lannister | 12 (+7) | 18 | 189 | 28 | 264 | 150 | 16→294 | +11 | 2673 | 20671 | ayakta |
| Tyrell | 11 (-1) | 9 | 85 | 32 | 202 | 251 | 18→246 | +13 | 592 | 44247 | ayakta |
| Baratheon (Fırtına Burnu) | 10 (-1) | 1 | 16 | 0 | 4 | 25 | 13→29 | +2 | 0 | 2447 | ayakta |
| Arryn | 7 (+0) | 5 | 53 | 2 | 112 | 98 | 14→112 | +6 | 568 | 6363 | ayakta |
| Braavos | 6 (+5) | 8 | 103 | 30 | 112 | 116 | 8→156 | +17 | 738 | 10206 | ayakta |
| Baratheon (Kralın Şehri) | 4 (-1) | 4 | 60 | 1 | 133 | 101 | 16→79 | +9 | 0 | 9188 | ayakta |
| Gece Nöbetçileri | 4 (+0) | 3 | 52 | 0 | 12 | 6 | 37→43 | +12 | 0 | 586 | ayakta |
| Greyjoy | 4 (-1) | 3 | 5 | 9 | 32 | 21 | 11→28 | +5 | 70 | 86 | ayakta |
| Baratheon (Ejderha Kayası) | 2 (-1) | 6 | 50 | 4 | 212 | 63 | 16→56 | +9 | 2470 | 1110 | ayakta |
| Targaryen | 2 (+1) | 4 | 40 | 10 | 210 | 29 | 9→58 | +20 | 1150 | 5747 | ayakta |
| Akgezenler | 2 (+1) | 1 | 6 | 0 | 0 | 0 | 0→11 | +4 | 126 | 460 | ayakta |
| Tully | 1 (-5) | 1 | 9 | 0 | 63 | 116 | 13→15 | +13 | 504 | 4290 | ayakta |
| Bolton | 0 (-1) | 0 | 0 | 0 | 2 | 5 | 8→0 | +2 | 65 | 543 | ELENDİ |
| Özgür Halk | 0 (-1) | 0 | 0 | 0 | 0 | 6 | 5→0 | +2 | 127 | 124 | ELENDİ |
| Özgür Şehirler | 0 (-6) | 0 | 0 | 0 | 5 | 51 | 11→0 | +12 | 6992 | 4509 | ELENDİ |
| Kardeşlik | 0 (-1) | 0 | 0 | 0 | 0 | 6 | 1→0 | +4 | 74 | 890 | ELENDİ |

Not: “Üretim emri”, simülasyon boyunca AI'nin kuyruğa aldığı birim üretim emirlerini; mevcut birim sayısı ise kayıp ve takviyeler sonrası nihai aktif birlikleri gösterir.

## Toprak değişimleri

- **Kuzey Krallığı**: +Taşlı Kıyı, Kraken Burnu, İkizler, Nehir Toprakları, The Crag, Üç Dişli Mızrak, Nehir Koşusu, Dehşet Kalesi, Harrenhal, Kralın Şehri; -The Crag, Nehir Toprakları, Nehir Koşusu, Harrenhal, Kralın Şehri, Üç Dişli Mızrak
- **Martell**: +Stoms End, Oldtown, Kral Ormanı; -Oldtown, Kral Ormanı
- **Lannister**: +Taç Toprakları, Duskendale, Kralın Şehri, Nehir Koşusu, Harrenhal, Üç Dişli Mızrak, Yengeç Pençesi Burnu; -Taç Toprakları, Kralın Şehri
- **Tyrell**: +Oldtown; -Oldtown
- **Baratheon (Fırtına Burnu)**: +Taç Toprakları; -Taç Toprakları, Stoms End
- **Braavos**: +Norvos, Myr, Essos Free Cities, Yunkai, İhtilaflı Topraklar; -Yunkai
- **Baratheon (Kralın Şehri)**: +Taç Toprakları, Oyuk Tepe, Kral Ormanı, Kralın Şehri, Nehir Toprakları; -Taç Toprakları, Duskendale, Kral Ormanı, Kralın Şehri
- **Greyjoy**: +—; -Kraken Burnu
- **Baratheon (Ejderha Kayası)**: +Kral Ormanı; -Kral Ormanı, Yengeç Pençesi Burnu
- **Targaryen**: +Meereen; -—
- **Akgezenler**: +Duvar'ın Ötesi; -—
- **Tully**: +Oyuk Tepe, Üç Dişli Mızrak, The Crag; -İkizler, Nehir Toprakları, The Crag, Üç Dişli Mızrak, Nehir Koşusu, Oyuk Tepe, Harrenhal
- **Bolton**: +—; -Dehşet Kalesi
- **Özgür Halk**: +—; -Duvar'ın Ötesi
- **Özgür Şehirler**: +İhtilaflı Topraklar; -Astapor, Norvos, Myr, Essos Free Cities, Yunkai, İhtilaflı Topraklar, Meereen
- **Kardeşlik**: +—; -Oyuk Tepe

## Savaşlar

| Taraflar | Başlangıç | Bitiş | Durum | Bölge kazanımı | Kayıp |
|---|---:|---:|---|---:|---:|
| Kardeşlik — Tully | 1 | 17. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 1 | 7. tur | aktif | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 2 | 6. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 12 | 19. tur | bitti | 0 / 0 | 0 / 0 |
| Özgür Halk — Akgezenler | 13 | 25. tur | bitti | 0 / 0 | 0 / 0 |
| Braavos — Özgür Şehirler | 24 | 29. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Bolton | 41 | 60. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Kuzey Krallığı | 41 | 66. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Greyjoy | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Greyjoy | 41 | 52. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Tully | 41 | 60. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Gece Nöbetçileri | 41 | 45. tur | bitti | 0 / 0 | 0 / 0 |
| Greyjoy — Kuzey Krallığı | 41 | 52. tur | bitti | 0 / 0 | 0 / 0 |
| Kuzey Krallığı — Tully | 41 | 76. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Fırtına Burnu) — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Bolton — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Akgezenler — Kuzey Krallığı | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Akgezenler — Tully | 57 | 61. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Kralın Şehri) | 64 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Lannister | 64 | 76. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Martell | 64 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Kralın Şehri) — Tully | 64 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Akgezenler | 64 | 68. tur | bitti | 0 / 0 | 0 / 0 |
| Arryn — Baratheon (Ejderha Kayası) | 67 | 71. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Lannister | 67 | 88. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Martell | 67 | 74. tur | bitti | 0 / 0 | 0 / 0 |
| Baratheon (Ejderha Kayası) — Tully | 67 | 71. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Kuzey Krallığı | 67 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Lannister — Tully | 76 | 84. tur | bitti | 0 / 0 | 0 / 0 |
| Gece Nöbetçileri — Akgezenler | 95 | 99. tur | bitti | 0 / 0 | 0 / 0 |

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
- 35. tur: **başladı** — Özgür Şehirler ↔ Gece Nöbetçileri (spice)
- 35. tur: **başladı** — Gece Nöbetçileri ↔ Özgür Şehirler (cloth)
- 37. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 37. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 39. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 39. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 41. tur: **bitti** — Greyjoy ↔ Gece Nöbetçileri (iron)
- 41. tur: **bitti** — Gece Nöbetçileri ↔ Greyjoy (grain)
- 42. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 42. tur: **başladı** — Braavos ↔ Kuzey Krallığı (spice)
- 42. tur: **başladı** — Kuzey Krallığı ↔ Braavos (iron)
- 42. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 43. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 43. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 47. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Targaryen (spice)
- 47. tur: **başladı** — Targaryen ↔ Baratheon (Ejderha Kayası) (spice)
- 48. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Greyjoy (spice)
- 48. tur: **başladı** — Greyjoy ↔ Baratheon (Kralın Şehri) (iron)
- 55. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Özgür Şehirler (spice)
- 55. tur: **başladı** — Özgür Şehirler ↔ Baratheon (Ejderha Kayası) (spice)
- 57. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Özgür Şehirler (spice)
- 57. tur: **başladı** — Özgür Şehirler ↔ Baratheon (Ejderha Kayası) (spice)
- 61. tur: **başladı** — Arryn ↔ Tyrell (iron)
- 61. tur: **başladı** — Tyrell ↔ Arryn (spice)
- 64. tur: **başladı** — Martell ↔ Gece Nöbetçileri (spice)
- 64. tur: **başladı** — Gece Nöbetçileri ↔ Martell (cloth)
- 64. tur: **bitti** — Baratheon (Kralın Şehri) ↔ Martell (grain)
- 64. tur: **bitti** — Martell ↔ Baratheon (Kralın Şehri) (spice)
- 66. tur: **başladı** — Baratheon (Ejderha Kayası) ↔ Baratheon (Kralın Şehri) (spice)
- 66. tur: **başladı** — Baratheon (Kralın Şehri) ↔ Baratheon (Ejderha Kayası) (grain)
- 68. tur: **başladı** — Gece Nöbetçileri ↔ Targaryen (cloth)
- 68. tur: **başladı** — Targaryen ↔ Gece Nöbetçileri (spice)
- 76. tur: **başladı** — Kuzey Krallığı ↔ Tully (iron)
- 76. tur: **başladı** — Tully ↔ Kuzey Krallığı (grain)
- 1. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 1. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 1. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Lannister
- 1. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Tyrell
- 1. tur: **ittifak kuruldu** — Gece Nöbetçileri ↔ Kuzey Krallığı
- 2. tur: **ittifak kuruldu** — Özgür Şehirler ↔ Targaryen
- 3. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tully
- 4. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tully
- 4. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 4. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 6. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 6. tur: **ittifak kuruldu** — Arryn ↔ Kardeşlik
- 7. tur: **ittifak bitti** — Arryn ↔ Kuzey Krallığı
- 7. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tully
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
- 37. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Kuzey Krallığı
- 38. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Tyrell
- 40. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 40. tur: **ittifak bitti** — Kuzey Krallığı ↔ Tully
- 40. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 43. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 44. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 45. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 50. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 52. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 53. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 54. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Martell
- 54. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 55. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Lannister
- 55. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 56. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 56. tur: **ittifak bitti** — Lannister ↔ Tully
- 56. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri
- 57. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 57. tur: **ittifak kuruldu** — Arryn ↔ Lannister
- 57. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 57. tur: **ittifak kuruldu** — Lannister ↔ Tully
- 58. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 59. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 59. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Tully
- 59. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Martell
- 60. tur: **ittifak bitti** — Özgür Şehirler ↔ Targaryen
- 61. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Bolton
- 62. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Bolton
- 62. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri
- 62. tur: **ittifak kuruldu** — Arryn ↔ Baratheon (Kralın Şehri)
- 62. tur: **ittifak kuruldu** — Braavos ↔ Gece Nöbetçileri
- 63. tur: **ittifak bitti** — Arryn ↔ Baratheon (Kralın Şehri)
- 67. tur: **ittifak bitti** — Arryn ↔ Baratheon (Ejderha Kayası)
- 67. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Akgezenler
- 67. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 68. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Akgezenler
- 68. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Akgezenler
- 69. tur: **ittifak bitti** — Lannister ↔ Martell
- 70. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Targaryen
- 70. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 71. tur: **ittifak kuruldu** — Braavos ↔ Kuzey Krallığı
- 73. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tyrell
- 75. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 75. tur: **ittifak kuruldu** — Lannister ↔ Martell
- 76. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 76. tur: **ittifak bitti** — Greyjoy ↔ Tully
- 76. tur: **ittifak bitti** — Lannister ↔ Tully
- 76. tur: **ittifak bitti** — Gece Nöbetçileri ↔ Tully
- 76. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 76. tur: **ittifak kuruldu** — Kuzey Krallığı ↔ Tully
- 77. tur: **ittifak bitti** — Arryn ↔ Greyjoy
- 77. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Baratheon (Fırtına Burnu)
- 77. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Kuzey Krallığı
- 79. tur: **ittifak kuruldu** — Arryn ↔ Greyjoy
- 83. tur: **ittifak bitti** — Arryn ↔ Lannister
- 84. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 89. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 90. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 91. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 93. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 94. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 94. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri
- 95. tur: **ittifak bitti** — Baratheon (Kralın Şehri) ↔ Tyrell
- 96. tur: **ittifak kuruldu** — Baratheon (Kralın Şehri) ↔ Tyrell
- 97. tur: **ittifak kuruldu** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 99. tur: **ittifak bitti** — Baratheon (Ejderha Kayası) ↔ Baratheon (Fırtına Burnu)
- 99. tur: **ittifak bitti** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri
- 100. tur: **ittifak kuruldu** — Baratheon (Fırtına Burnu) ↔ Gece Nöbetçileri

## Elenen devletler

- 17. tur: **Kardeşlik**
- 25. tur: **Özgür Halk**
- 60. tur: **Özgür Şehirler**
- 67. tur: **Bolton**

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

1. Martell — 21 bölge, 5 ordu, 39 kara birimi, güç 396
2. Stark — 19 bölge, 5 ordu, 79 kara birimi, güç 810
3. Tyrell — 12 bölge, 5 ordu, 43 kara birimi, güç 746
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 31 kara birimi, güç 288
5. Arryn — 7 bölge, 2 ordu, 20 kara birimi, güç 289

### 39. tur (298/04)

1. Martell — 22 bölge, 6 ordu, 41 kara birimi, güç 507
2. Stark — 19 bölge, 7 ordu, 99 kara birimi, güç 679
3. Tyrell — 11 bölge, 8 ordu, 47 kara birimi, güç 756
4. Baratheon (Fırtına Burnu) — 10 bölge, 3 ordu, 34 kara birimi, güç 312
5. Arryn — 7 bölge, 3 ordu, 21 kara birimi, güç 303

### 49. tur (299/02)

1. Kuzey Krallığı — 22 bölge, 6 ordu, 76 kara birimi, güç 533
2. Martell — 22 bölge, 7 ordu, 41 kara birimi, güç 528
3. Tyrell — 11 bölge, 5 ordu, 46 kara birimi, güç 667
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 35 kara birimi, güç 367
5. Arryn — 7 bölge, 2 ordu, 11 kara birimi, güç 130

### 59. tur (299/12)

1. Kuzey Krallığı — 25 bölge, 7 ordu, 87 kara birimi, güç 1199
2. Martell — 22 bölge, 7 ordu, 40 kara birimi, güç 709
3. Tyrell — 11 bölge, 6 ordu, 50 kara birimi, güç 982
4. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 35 kara birimi, güç 245
5. Arryn — 7 bölge, 4 ordu, 31 kara birimi, güç 393

### 69. tur (300/10)

1. Kuzey Krallığı — 26 bölge, 7 ordu, 93 kara birimi, güç 1251
2. Martell — 23 bölge, 6 ordu, 43 kara birimi, güç 706
3. Tyrell — 11 bölge, 7 ordu, 61 kara birimi, güç 1341
4. Baratheon (Fırtına Burnu) — 10 bölge, 0 ordu, 0 kara birimi, güç 0
5. Lannister — 8 bölge, 14 ordu, 149 kara birimi, güç 1925

### 79. tur (301/08)

1. Kuzey Krallığı — 27 bölge, 6 ordu, 117 kara birimi, güç 1536
2. Martell — 22 bölge, 6 ordu, 46 kara birimi, güç 809
3. Tyrell — 11 bölge, 7 ordu, 64 kara birimi, güç 1545
4. Baratheon (Fırtına Burnu) — 10 bölge, 1 ordu, 1 kara birimi, güç 8
5. Lannister — 7 bölge, 12 ordu, 140 kara birimi, güç 1900

### 89. tur (302/06)

1. Kuzey Krallığı — 23 bölge, 9 ordu, 168 kara birimi, güç 1447
2. Martell — 22 bölge, 7 ordu, 58 kara birimi, güç 946
3. Lannister — 12 bölge, 13 ordu, 163 kara birimi, güç 2347
4. Tyrell — 11 bölge, 9 ordu, 69 kara birimi, güç 1648
5. Baratheon (Fırtına Burnu) — 10 bölge, 2 ordu, 7 kara birimi, güç 98

### 99. tur (303/04)

1. Kuzey Krallığı — 23 bölge, 11 ordu, 212 kara birimi, güç 1474
2. Martell — 22 bölge, 9 ordu, 67 kara birimi, güç 849
3. Lannister — 12 bölge, 17 ordu, 182 kara birimi, güç 2072
4. Tyrell — 11 bölge, 8 ordu, 81 kara birimi, güç 1643
5. Baratheon (Fırtına Burnu) — 10 bölge, 1 ordu, 15 kara birimi, güç 129

### 100. tur (303/05)

1. Kuzey Krallığı — 23 bölge, 13 ordu, 220 kara birimi, güç 1524
2. Martell — 22 bölge, 6 ordu, 64 kara birimi, güç 1020
3. Lannister — 12 bölge, 18 ordu, 189 kara birimi, güç 2190
4. Tyrell — 11 bölge, 9 ordu, 85 kara birimi, güç 1635
5. Baratheon (Fırtına Burnu) — 10 bölge, 1 ordu, 16 kara birimi, güç 178

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
- 68. tur: **Gece Nöbeti'nin Duvar Hazırlıkları** (`night_watch_wall_preparations_300`)
- 76. tur: **Beş Kralın Savaşı: Stark Zaferi** (`war_of_five_kings_resolution_stark`)
- 77. tur: **Kuzey Zaferinin Düzeni** (`postwar_settlement_stark`)
- 78. tur: **Taç Topraklarının Savaş Yorgunluğu** (`crownlands_postwar_exhaustion_300`)
- 79. tur: **Reach'in Savaş Sonrası Konsolidasyonu** (`tyrell_postwar_consolidation_301`)
- 80. tur: **Akgezenlere Karşı Son Hazırlık** (`night_watch_against_others_301`)
