package combat

import (
	"math/rand"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

// TechMods savaşa etki eden teknoloji çarpanları.
type TechMods struct {
	AttackMod  float64 // toplam saldırı çarpanı (ör. 0.10 = +10%)
	DefenseMod float64 // toplam savunma çarpanı
}

// Result savaşın sonucunu özetler.
type Result struct {
	AttackerWins bool
	AttackerLost int
	DefenderLost int
	Description  string
}

// ResolveBattle iki ordu arasındaki çarpışmayı hesaplar.
func ResolveBattle(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType) Result {
	return ResolveBattleWithMods(atk, def, terrain, types, TechMods{}, TechMods{})
}

// ResolveBattleWithMods teknoloji modlarını dahil ederek savaşı hesaplar.
func ResolveBattleWithMods(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods) Result {
	atkStr := float64(atk.TotalStrength(types)) * (1.0 + atkMods.AttackMod)
	defStr := float64(def.TotalStrength(types)) * terrainBonus(terrain) * (1.0 + defMods.DefenseMod)

	attackerWins, atkLoss, defLoss := calculateOutcome(atkStr, defStr)

	atkDead := applyCasualties(atk, atkLoss)
	defDead := applyCasualties(def, defLoss)

	return Result{
		AttackerWins: attackerWins,
		AttackerLost: atkDead,
		DefenderLost: defDead,
		Description:  outcomeDescription(attackerWins, atkLoss, defLoss),
	}
}

// terrainBonus savunucuya araziye göre güç çarpanı uygular.
func terrainBonus(t world.TerrainType) float64 {
	switch t {
	case world.TerrainMountain:
		return 1.8
	case world.TerrainPass:
		return 1.5
	case world.TerrainForest:
		return 1.3
	case world.TerrainCoast:
		return 1.1
	default:
		return 1.0
	}
}

// calculateOutcome güç oranına göre savaşın kazananını ve kayıp oranlarını belirler.
// atkStr, defStr: iki ordunun net güç değerleri (savunucu için arazi bonusu zaten uygulandı).
// Döner: (saldıran kazandı mı, saldıranın kayıp oranı [0–1], savunucunun kayıp oranı [0–1])
//
// Bu fonksiyon savaşın nasıl hissettireceğini doğrudan belirler.
// Burayı sen yaz! Düşünülecek seçenekler:
//   - Tamamen deterministik mi (ratio >= 1.0 = zafer) yoksa hafif rastgele mi?
//   - Dar zafer vs. ezici zafer ayrımı yapılsın mı? (kayıp oranları farklı olabilir)
//   - Sadece ratio'ya göre mi yoksa mutlak güç farkına göre de mi?
//
// Örnek çerçeve:
//
//	ratio := atkStr / (defStr + 1)
//	if ratio > 1.5 { // ezici zafer → saldıran az kayıp, savunucu yok edilir
//	    return true, 0.10, 0.80
//	} else if ratio >= 1.0 { // dar zafer → her iki taraf da kayıp verir
//	    return true, 0.35, 0.50
//	} else if ratio >= 0.7 { // yakın mağlubiyet
//	    return false, 0.50, 0.30
//	} else { // ezici mağlubiyet
//	    return false, 0.80, 0.10
//	}
func calculateOutcome(atkStr, defStr float64) (attackerWins bool, atkCasualtyRatio, defCasualtyRatio float64) {
	// ±%15 aralığında rastgele güç dalgalanması — zayıf ordu nadiren kazanabilir
	dice := (rand.Float64()*2 - 1) * 0.15 // [-0.15, +0.15]
	ratio := (atkStr / (defStr + 1)) * (1 + dice)

	if ratio > 1.5 {
		return true, 0.10, 0.80
	} else if ratio >= 1.0 {
		return true, 0.35, 0.50
	} else if ratio >= 0.7 {
		return false, 0.50, 0.30
	}
	return false, 0.80, 0.10
}

// applyCasualties ordudaki birim sayısını ratio kadar azaltır.
// Ratio doğrudan "ölen birim oranı" olarak yorumlanır (HP fraksiyonu değil).
func applyCasualties(a *army.Army, ratio float64) (lost int) {
	n := len(a.Units)
	toKill := int(float64(n)*ratio + 0.5) // yuvarla
	if toKill > n {
		toKill = n
	}
	a.Units = a.Units[:n-toKill]
	return toKill
}

func outcomeDescription(wins bool, atkLoss, defLoss float64) string {
	if wins {
		if atkLoss <= 0.15 {
			return "Ezici Zafer"
		}
		return "Dar Zafer"
	}
	if defLoss <= 0.15 {
		return "Ağır Yenilgi"
	}
	return "Geri Çekilme"
}
