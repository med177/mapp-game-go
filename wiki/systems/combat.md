---
type: system
tags: [combat, battle, terrain, casualties]
last_updated: 2026-06-20
related: [systems/ai, world/regions, systems/tech-tree, architecture/render-pipeline]
---

# Çarpışma Sistemi

**Kaynak:** `internal/combat/combat.go`

## Genel Bakış

Tüm çarpışmalar harita üzerinde otomatik hesaplanır — ayrı taktik sahne yok. Ordu bir düşman bölgesine hareket edince `ResolveBattleWithPlan()` veya `ResolveBattleWithContextPlan()` tetiklenir; oyuncu kara, deniz ve çıkarma saldırılarında önce duruş seçer. Tahkimli kara bölgelerde ise akış artık doğrudan fetih değildir: önce kuşatma veya genel hücum kararı gerekir.

---

## Hesap Akışı

```
saldıranGücü = ordu.TotalStrength(types) × (1 + atkMods.AttackMod)
savunucuGücü = ordu.TotalStrength(types) × terrainBonus × (1 + defMods.DefenseMod)

calculateOutcome(saldıranGücü, savunucuGücü)
    → (kazandıMı bool, saldıranKayıpOranı float64, savunucuKayıpOranı float64)

applyCasualties(ordu, kayıpOranı) → HP hasarı + gerekirse gerçek birim kaybı
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

`calculateOutcome()` — `internal/combat/combat.go`

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

## Savaş Duruşları ve Bağlamı

`BattleStance` — `internal/combat/combat.go`

Oyuncu saldırı başlatırken üç duruştan birini seçer:

| Duruş | Etki |
|---|---|
| `Agresif` | efektif saldırı gücünü artırır; saldıran ve savunan kayıp oranlarını yukarı çeker |
| `Dengeli` | mevcut modelin nötr hali |
| `Savunmacı` | efektif saldırı gücünü düşürür; kayıp oranlarını azaltır |

Sistem artık üç ayrı savaş bağlamı tanır:

| Bağlam | Açıklama |
|---|---|
| `land` | klasik kara ordusu saldırısı |
| `naval` | iki donanmanın deniz bölgesindeki çatışması |
| `amphibious` | nakliye filosunun düşman kıyıya asker çıkartırken girdiği savaş |

Her bağlam kendi duruş çarpanını ve Türkçe özetini kullanır. Yani `Agresif` kara, deniz ve çıkarma savaşlarında aynı isimle görünse de etkisi birebir kopya değildir.

Bu seçim hem gerçek resolve hattında hem de saldırı öncesi preview panelinde aynı helper'larla hesaplanır:

- `battleStrengths()` efektif güçleri çıkarır
- `resolveOutcome()` gerçek zar sonucunu duruş çarpanlarıyla birleştirir
- `PreviewBattleWithContextMods()` zar aralığını tarayıp muhtemel sonuç, zafer şansı ve tahmini kayıp penceresi üretir

Preview tarafındaki kayıp özeti artık iki katmanlıdır:

- `HP` kaybı: orta ölçekli çatışmalarda birim ölmese bile hasarın görünmesini sağlar
- `Birim` kaybı: gerçekten düşmesi beklenen birlik sayısını verir

Preview, gerçek savaştakiyle aynı arazi/teknoloji/duruş/savaş tipi matematiğini kullanır; yani panelde görülen güç ve gerçek resolve birbirinden kopmaz.

## Kuşatma Akışı

`internal/game/siege.go`

Tahkimli kara bölgesi (`fortress` settlement veya `walls` seviyesi) artık ayrı bir ara katman üretir:

1. Saldıran kara ordusunda en az bir `siege` kategorili birlik olmalı.
2. İlk temas anında oyuncu sağ tık sonrası `Kuşatma Kararı` modalında `Kuşatma Başlat` veya `Genel Hücum` seçer.
3. `Kuşatma Başlat` anında ordu hedefe girmez; `GameState.Sieges` içine kayıt yazılır ve ordu hareketi biter.
4. Gedik ilerlemesi artık kuşatma ekipmanı tier'i ile sınırlıdır: `fortLevel = 3` ise yalnız `Tier 3` kuşatma birimi yeni gedik ilerlemesi üretebilir. Daha düşük tier araçlar kuşatmayı sürdürür, savunucuyu yıpratır ama yeni gedik açamaz.
5. Her tur çözümlemesinde kuşatma baskısı savunucu orduya attrition uygular; gedik kapasitesi yetiyorsa ayrıca `BreachProgress` artar ve gedik seviyesi (`yok / küçük / büyük`) güncellenir.
6. Aktif kuşatma seçildiğinde renderer alt-ortada modal-dışı `Kuşatma Emri` paneli gösterir; oyuncu buradan `Genel Hücum` veya `Kuşatmayı Kaldır` seçebilir.
7. Kuşatmayı yapan ordu başka komşu bölgeye normal hareket emri alırsa bu hareket, eski kuşatmayı otomatik kaldırır; ayrı ikinci onay gerekmez.
8. Savunucu ordu çözülüp büyük gedik açılırsa tahkimat teslim olabilir; gedik açılamasa bile uzun aç bırakma kuşatması sonunda teslimiyet mümkün kalır. Oyuncu veya AI isterse kuşatma üstünden genel hücum da deneyebilir.

Kuşatma hücumunda savunana arazi bonusuna ek olarak tahkimat savunma çarpanı uygulanır. Gedik büyüdükçe bu bonus düşer; yani surlar kırıldıkça saha savaşı normal kara muharebesine yaklaşır.

---

## Hasar ve Toparlanma

Çarpışma sonucu artık ilk etapta tüm kaybı doğrudan birim silerek çözmez. `applyCasualties()` toplam HP havuzu üstünden hasarı dağıtır:

- orta kayıplarda birimler hayatta kalır ama hasarlı çıkar,
- ağır kayıplarda bazı birimler tamamen düşer,
- yaşayan birimlerin savaş katkısı `army.TotalStrength()` ve `TotalDefense()` içinde mevcut HP oranıyla ölçeklenir.

`internal/game/resolution.go`

- kara orduları sahip oldukları kara bölgede tur başına `+10 HP` toparlanır,
- toparlanma `CurrentHP < 100` olan birimlerde çalışır ve `%100`e kadar sürer,
- kış turunda önce attrition uygulanır, aynı sweep içinde ek ücretsiz toparlanma verilmez,
- donanmalar ve dost olmayan topraktaki kara orduları bu akıştan yararlanmaz.

## Savaş Sonrası Uygulama

`internal/game/game.go`

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

## Deniz Hareketi ve Çatışma Kuralı

Donanmalar deniz bölgeleri arasında savaş ilanı olmadan serbest hareket eder.
Denizde çatışma yalnızca iki fraksiyon arasında `StanceWar` varsa tetiklenir; barış/ittifak/trade durumunda aynı deniz bölgesine girildiğinde savaş açılmaz.
Savaş varsa oyuncu gemi-gemi çatışmasında da önce duruş seçer. Düşman kıyıya çıkarma sırasında savunan ordu varsa aynı modal `Çıkarma Muharebesi` başlığıyla açılır ve seçilen duruş `ActionDisembarkArmy` üzerinden oyun katmanına taşınır.

---

## Birim Gücü

`army.TotalStrength(types)` — her birimin `Attack + Morale/10` değeri mevcut HP oranıyla ağırlıklandırılır.

→ Birim tipleri: `assets/data/units.json`
