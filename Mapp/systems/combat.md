---
type: system
tags: [combat, battle, terrain, casualties]
last_updated: 2026-05-07
related: [systems/ai, world/regions, systems/tech-tree, architecture/render-pipeline]
---

# Çarpışma Sistemi

**Kaynak:** `internal/combat/combat.go`

## Genel Bakış

Tüm çarpışmalar harita üzerinde otomatik hesaplanır — ayrı taktik sahne yok. Ordu bir düşman bölgesine hareket edince `ResolveBattleWithMods()` tetiklenir.

---

## Hesap Akışı

```
saldıranGücü = ordu.TotalStrength(types) × (1 + atkMods.AttackMod)
savunucuGücü = ordu.TotalStrength(types) × terrainBonus × (1 + defMods.DefenseMod)

calculateOutcome(saldıranGücü, savunucuGücü)
    → (kazandıMı bool, saldıranKayıpOranı float64, savunucuKayıpOranı float64)

applyCasualties(ordu, kayıpOranı) → gerçek birim kayıpları
```

---

## Arazi Bonusları (Savunucu)

`terrainBonus()` — `internal/combat/combat.go:48`

| Arazi | Savunucu Çarpanı |
|---|---|
| Dağ | ×1.8 |
| Geçit | ×1.5 |
| Orman | ×1.3 |
| Kıyı | ×1.1 |
| Ova / Diğer | ×1.0 |

→ Arazi tipleri: [[world/regions]]

---

## calculateOutcome — Uygulama

`calculateOutcome()` — `internal/combat/combat.go:85`

±%15 rastgele zar dalgalanması içerir; zayıf ordu nadir de olsa kazanabilir.

```
dice  := rand.Float64()*2 - 1) * 0.15   // [-0.15, +0.15]
ratio := (atkStr / (defStr + 1)) * (1 + dice)
```

| Koşul | Sonuç | Saldıran Kayıp | Savunucu Kayıp |
|---|---|---|---|
| `ratio > 1.5` | Ezici Zafer | %10 | %80 |
| `ratio >= 1.0` | Dar Zafer | %35 | %50 |
| `ratio >= 0.7` | Geri Çekilme | %50 | %30 |
| `ratio < 0.7` | Ağır Yenilgi | %80 | %10 |

`outcomeDescription()` sonuca göre `"Ezici Zafer"`, `"Dar Zafer"`, `"Geri Çekilme"`, `"Ağır Yenilgi"` metin üretir.

---

## Teknoloji Modları

`TechMods{AttackMod, DefenseMod}` — `internal/combat/combat.go:10`

`game.techModsFor()` oyuncu/AI için tamamlanan teknoloji etkilerini toplar:
- `InfantryAttackMod + CavalryAttackMod + SiegeAttackMod` → `AttackMod`
- `LandDefenseMod` → `DefenseMod`

→ Teknoloji efektleri: [[systems/tech-tree]]

---

## Savaş Sonrası Uygulama

`internal/game/game.go:515`

```
if saldıranKazandı:
    düşman ordusu → temizlendi (birimsiz kaldıysa)
    saldıran ordu → hedef bölgeye taşındı
    targetRegion.ApplyConquest(ownerID, religion) → sahiplik değişti
    a.MovePoints--
else:
    saldıran ordu → yerinde kaldı (birimsiz kaldıysa silinir)
```

`ApplyConquest` bölgeyi yeni fraksiyona devreder ve din dönüşüm sayacını başlatır.

---

## Birim Gücü

`army.TotalStrength(types)` — her birimin `types[unit.TypeID].Attack` değerlerinin toplamı ve mevcut HP oranıyla ağırlıklandırılır.

→ Birim tipleri: `assets/data/units.json`
