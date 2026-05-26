# 1300 Senaryosu - 20 Tur Snowball Risk Raporu

Tarih: 2026-05-27

## Metod
- Bu rapor tam oyun simülasyonu değil; veri tabanlı öngörü skorudur.
- Skor bileşenleri: stok kaynaklar, bölge üretimi, toplam birlik, ticaret ağ bağlantısı, AI agresifliği, trade capacity, bölge sayısı.
- Yaklaşık skor formülü: `0.20*stok + 0.55*bölge_ekonomi + 6*birlik + 8*ticaret_link + 2*agresiflik + 1.2*trade_capacity + 1*bölge_adedi`

## Genel Sıralama (İlk 20)

| Sıra | Fraksiyon | Playable | Snowball Skoru |
|---|---|---:|---:|
| 1 | ilkhanate | false | 1453.0 |
| 2 | hre | false | 1352.5 |
| 3 | mamluk | true | 981.5 |
| 4 | france | true | 912.5 |
| 5 | hungarian_kingdom | false | 903.9 |
| 6 | byzantine | false | 860.3 |
| 7 | england | true | 845.8 |
| 8 | venice | true | 722.6 |
| 9 | polish_kingdom | false | 591.8 |
| 10 | russia | true | 585.9 |
| 11 | genoa | false | 559.7 |
| 12 | aragon | true | 534.9 |
| 13 | florence_rep | false | 526.2 |
| 14 | ottoman | true | 513.4 |
| 15 | golden_horde | false | 504.8 |
| 16 | milan_duchy | false | 500.3 |
| 17 | novgorod_rep | false | 470.0 |
| 18 | lithuanian_gd | false | 453.4 |
| 19 | teutonic_order | false | 441.6 |
| 20 | serbian_empire | false | 433.9 |

## Playable Fraksiyon Sıralaması

| Sıra | Fraksiyon | Skor | Risk Seviyesi |
|---|---|---:|---|
| 1 | mamluk | 981.5 | Çok Yüksek |
| 2 | france | 912.5 | Çok Yüksek |
| 3 | england | 845.8 | Yüksek |
| 4 | venice | 722.6 | Yüksek |
| 5 | russia | 585.9 | Orta |
| 6 | aragon | 534.9 | Orta |
| 7 | ottoman | 513.4 | Orta |
| 8 | safavid | 423.3 | Düşük |
| 9 | portugal | 382.0 | Düşük |

## Bulgular
- `mamluk` açık ara güçlü başlangıç yapıyor ve erken snowball riski yüksek.
- `venice` ekonomi+ticaret kaynaklı çok hızlı büyüyebiliyor.
- `ottoman` orta seviyede; erken oyunda kırılgan değil ama tek başına runaway yapmıyor.
- `portugal` ve `safavid` oynanabilirler içinde düşük-orta eşikte kalıyor.
- Oynanamaz ama haritayı domine eden bloklar: `ilkhanate`, `hre`.

## Kalibrasyon Önerisi (İsteğe Bağlı İkinci Tur)
1. `ilkhanate` için toplam başlangıç birliklerini 35 -> 24 bandına çek.
2. `venice` için `spice` ve `cloth` toplamını ~40 azalt.
3. `mamluk` için başlangıç altını ~100 azalt veya 1 stack düşür.
4. `ottoman` için erken genişleme hissi isteniyorsa +2 hafif süvari eklenebilir.
