---
type: system
tags: [combat, battle, terrain, casualties]
last_updated: 2026-05-06
related: [systems/ai, world/regions, systems/tech-tree]
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

## calculateOutcome — ⚠️ Eksik Uygulama

`calculateOutcome()` fonksiyonu şu an **kullanıcı tarafından yazılmayı bekliyor.**

`internal/combat/combat.go:63-100` arası hazır iskelet:
- Parametre: `atkStr, defStr float64`
- Döner: `(kazandıMı bool, atkKayıpOranı float64, defKayıpOranı float64)`
- Düşünülecek tasarım kararları:
  - Tamamen deterministik mi? → `ratio >= 1.0 = zafer`
  - Hafif rastgele sonuç mu? → `rand.Float64()` ile varyasyon
  - Ezici zafer vs. dar zafer ayrımı mı? → farklı kayıp oranları

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
