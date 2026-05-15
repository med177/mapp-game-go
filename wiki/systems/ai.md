---
type: system
tags: [ai, strategy, coalition, difficulty]
last_updated: 2026-05-15
related: [systems/combat, systems/diplomacy, architecture/game-loop]
---

# Yapay Zeka Sistemi

**Kaynak:** `internal/ai/ai.go`

## Genel Yapı

Her `PhaseAITurn`'de tüm AI fraksiyonları sırayla işlenir:

```go
for fid := range gs.Factions {
    if fid == gs.PlayerFactionID { continue }
    ai.TakeTurn(gs, fid)
}
```

`TakeTurn` sırasıyla şu adımları yapar:
1. Zorluk 3 ise → `FormCoalitionAgainstPlayer()`
2. Diplomasi taraması → `aiHandleDiplomacy()`
3. Teknoloji araştırma → `aiResearch()`
4. Ekonomik bina inşası → `aiEconomyBuild()`
5. Deniz stratejisi → `aiNavalStrategy()`
6. Birim alımı + kışla inşası → `aiRecruitAndBuild()`
7. Ordu hareketi → `moveArmy()` (her ordu için)

---

## Ordu Hareketi Mantığı

`moveArmy()` → `chooseBestMove()` → `scoreMove()` → `executeMove()`

`scoreMove()` hedef bölge için puan hesaplar:

| Koşul | Puan |
|---|---|
| Kendi bölgesi | 0 (hareket etme) |
| Barış/ittifak/ticaret halindeki bölge | -1 (atla) |
| Savaş halinde + üstün güç | 95 |
| Daha güçlü düşman var | -1 |
| Sahipsiz bölge (kapasite doluysa) | 70 |
| Sahipsiz bölge | 50 |
| Düşman bölgesi, kapasite dolu + savaş | 100 |
| Düşman bölgesi, savaş | 90 |

`atCapacity` — `DeployedLandUnits >= ManpowerCap` ise fetih yaparak kapasite genişletme önceliklenir.

AI de oyuncu ile aynı `combat.ResolveBattleWithMods()` kullanır.

---

## Koalisyon Mantığı

`FormCoalitionAgainstPlayer()` — zorluk 3'te her AI turunun başında çalışır.

**Tetikleme koşulu:** Oyuncunun bölge sayısı `coalitionThreshold = 8`'i geçmesi.

**Etki:** AI fraksiyon oyuncuya savaş açar ve aynı diplomasi motoru üzerinden diğer AI'larla ittifak kurmaya çalışır.

→ İttifak mekanizması: [[systems/diplomacy]]

---

## Diplomasi Safhası

`aiHandleDiplomacy()` her AI turunda ilişkileri tarar:

- `war` ilişkisinde skor çok düşmüşse veya AI askeri/bölgesel olarak gerideyse barış teklif eder
- `peace` ilişkisinde ortak düşman ve yeterli skor varsa ittifak dener
- `peace` ilişkisinde skor nötr veya pozitifse ticaret dener

AI ve oyuncu aynı `internal/diplomacy` motorunu kullandığı için:

- kabul/red kuralları tutarlıdır
- ticaret rotaları aynı şekilde açılıp kapanır
- AI barışta olan veya ticaret yaptığı hedefe saldırmaz

---

## Teknoloji Modları

`aiTechMods()` — AI de oyuncuyla aynı teknoloji bonuslarını hesaplar:

```go
fx := tech.ComputeEffects(f.Research.Completed, gs.TechTypes)
return TechMods{
    AttackMod:  fx.InfantryAttackMod + fx.CavalryAttackMod + fx.SiegeAttackMod,
    DefenseMod: fx.LandDefenseMod,
}
```

---

## Zorluk Seviyeleri

| Seviye | Fark |
|---|---|
| 1 (Kolay) | Pasif AI, yavaş büyüme |
| 2 (Normal) | Dengeli strateji |
| 3 (Zor) | +300 başlangıç altın, +100 tahıl; koalisyon mantığı aktif |

Zorluk 3 başlangıç bonusu `resetToNewGame()` içinde uygulanır — `internal/game/game.go:337`

---

## Teknoloji Araştırma (`aiResearch`)

Aktif araştırma yoksa başlatır. Öncelik sırası:

| Kategori | Puan | Ek bonus |
|---|---|---|
| `military` | 100 | Saldırı efektleri varsa +20 |
| `economy` | 70 | `gold_per_region` varsa +15 |
| `naval` | 50 | — |
| `diplomacy` | 40 | — |
| `religion` | 30 | — |

Kısa süreli teknolojilere `TurnsRequired / 2` azaltma uygulanır.

---

## Ekonomik Bina (`aiEconomyBuild`)

Her tur en fazla bir bina inşa eder. Öncelik:
1. **Pazar** (prio 80) — her zaman uygun
2. **Çiftlik** (prio 60) — `BaseGrainOutput < 20` bölgelere
3. **Sur** (prio 50) — sınır bölgelerine (komşuda farklı fraksiyon varsa)

---

## Deniz Stratejisi (`aiNavalStrategy`)

Kıyı bölgesi varsa:
1. Limansız kıyı bölgesine liman inşa et
2. `fleetLimit = 2` — en fazla 2 adet deniz filosu oluştur
3. Filo oluşturulurken nakliye gemisi (`transport`) konur

---

## AI Deniz Taşıma Akışı

AI artık kara ordularını nakliye filosuna bindirip indirebilir:

- Kara ordusu `chooseBestMove()` içinde komşu deniz bölgesini, o denizde uygun `transport` filosu varsa ve karşı kıyıda pozitif hedef skoru varsa seçer.
- `executeMove()` kara → deniz geçişinde birimleri filonun `EmbarkedUnits` alanına taşır ve kara ordusunu haritadan kaldırır.
- Donanma `EmbarkedUnits` taşıyorsa komşu kara bölgesine çıkarma (`disembark`) yapar; yeni kara ordusu üretilir.
- Düşman kıyıya çıkarma yalnızca savaş halindeyken yapılır; barışta AI çıkarma denemez.
- Düşman kıyıda ordu varsa AI çıkarma hedeflemesinde güç kıyası yapar; zayıfsa çıkarma girişimini atlar.
- Çıkarma savaşı yine `combat.ResolveBattleWithMods()` ile çözülür; kazanırsa çıkarma ordusu karaya iner ve bölge el değiştirir.

Kaynak kod:
- `internal/ai/ai.go:377`
- `internal/ai/ai.go:438`
- `internal/ai/ai.go:666`

Testler:
- `internal/ai/ai_test.go:67`
- `internal/ai/ai_test.go:119`
- `internal/ai/ai_test.go:172`
- `internal/ai/ai_test.go:221`

---

## Birim Alımı (`aiRecruitAndBuild`)

Manpower sıkışıksa önce kışla inşa eder. Sonra `aiSelectBestUnit()` ile birim seçer:

| Öncelik | Birim | Koşul |
|---|---|---|
| 1 | `elite_infantry` | Altın ≥ 350 + rezerv, teknoloji tamamlandıysa |
| 2 | `heavy_cavalry` | Altın ≥ 450 + rezerv, teknoloji tamamlandıysa |
| 3 | `infantry` | Altın ≥ 180 + rezerv |
| 4 | `cavalry` | Altın ≥ 300 + rezerv |
| 5 | `light_cavalry` | Altın ≥ 200 + rezerv |
| 6 | `cannon/bombard` | Altın ≥ 650 + rezerv, savaş halinde |
| 7 | `militia` | Varsayılan |

---

## Eksik / Planlanan

- [ ] AI çoklu ordu konsolidasyonu (dağınık ordular ana orduya katılsın)
- [ ] AI uzun menzilli planlama (sadece komşu değil, stratejik hedef)
- [ ] Diplomasi teklif önceliklerini tehdit seviyesi ve teknoloji farkıyla daha da zenginleştir
