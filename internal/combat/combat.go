package combat

import (
	"math/rand"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/world"
)

type BattleStance string

const (
	BattleStanceAggressive BattleStance = "aggressive"
	BattleStanceBalanced   BattleStance = "balanced"
	BattleStanceDefensive  BattleStance = "defensive"
)

type BattleContext string

const (
	BattleContextLand       BattleContext = "land"
	BattleContextNaval      BattleContext = "naval"
	BattleContextAmphibious BattleContext = "amphibious"
)

type battleStanceConfig struct {
	LabelTR         string
	SummaryTR       string
	AttackMod       float64
	AttackerLossMod float64
	DefenderLossMod float64
}

type Preview struct {
	Stance               BattleStance
	StanceLabelTR        string
	StanceSummaryTR      string
	AttackStrength       int
	DefenseStrength      int
	WinChance            int
	LikelyOutcome        string
	AttackerHPDamageMin  int
	AttackerHPDamageMax  int
	AttackerHPExpected   int
	AttackerLossMin      int
	AttackerLossMax      int
	AttackerLossExpected int
	DefenderHPDamageMin  int
	DefenderHPDamageMax  int
	DefenderHPExpected   int
	DefenderLossMin      int
	DefenderLossMax      int
	DefenderLossExpected int
}

// TechMods savaşa etki eden teknoloji çarpanları.
type TechMods struct {
	AttackMod       float64 // kara saldırı çarpanı (ör. 0.10 = +10%)
	DefenseMod      float64 // kara savunma çarpanı
	NavalAttackMod  float64 // deniz saldırı çarpanı
	NavalDefenseMod float64 // deniz savunma çarpanı
}

// Result savaşın sonucunu özetler.
type Result struct {
	AttackerWins     bool
	AttackerLost     int
	DefenderLost     int
	AttackerHPDamage int
	DefenderHPDamage int
	Description      string
}

// ResolveBattle iki ordu arasındaki çarpışmayı hesaplar.
func ResolveBattle(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType) Result {
	return ResolveBattleWithPlan(atk, def, terrain, types, TechMods{}, TechMods{}, BattleStanceBalanced)
}

// ResolveBattleWithMods teknoloji modlarını dahil ederek savaşı hesaplar.
func ResolveBattleWithMods(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods) Result {
	return ResolveBattleWithPlan(atk, def, terrain, types, atkMods, defMods, BattleStanceBalanced)
}

// ResolveBattleWithPlan teknoloji ve saldırı duruşu modlarını dahil ederek savaşı hesaplar.
func ResolveBattleWithPlan(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods, stance BattleStance) Result {
	return ResolveBattleWithContextPlan(atk, def, terrain, types, atkMods, defMods, inferBattleContext(atk), stance)
}

// ResolveBattleWithContextPlan teknoloji, savaş tipi ve saldırı duruşu modlarını dahil ederek savaşı hesaplar.
func ResolveBattleWithContextPlan(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods, context BattleContext, stance BattleStance) Result {
	atkStr, defStr := battleStrengths(atk, def, terrain, types, atkMods, defMods, context, stance)
	outcome := resolveOutcome(atkStr, defStr, context, stance)
	atkHPBefore := 0
	defHPBefore := 0
	if atk != nil {
		atkHPBefore = totalUnitsHP(atk.Units)
	}
	if def != nil {
		defHPBefore = totalUnitsHP(def.Units)
	}

	atkDead := applyCasualties(atk, outcome.AttackerLossRatio)
	defDead := applyCasualties(def, outcome.DefenderLossRatio)
	atkHPAfter := 0
	defHPAfter := 0
	if atk != nil {
		atkHPAfter = totalUnitsHP(atk.Units)
	}
	if def != nil {
		defHPAfter = totalUnitsHP(def.Units)
	}
	if outcome.AttackerWins && len(def.Units) > 0 {
		defDead += len(def.Units)
		defHPAfter = 0
		def.Units = def.Units[:0]
	}

	return Result{
		AttackerWins:     outcome.AttackerWins,
		AttackerLost:     atkDead,
		DefenderLost:     defDead,
		AttackerHPDamage: atkHPBefore - atkHPAfter,
		DefenderHPDamage: defHPBefore - defHPAfter,
		Description:      outcome.Description,
	}
}

// PreviewBattleWithMods saldırı başlamadan önce aynı matematikle muhtemel sonucu özetler.
func PreviewBattleWithMods(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods, stance BattleStance) Preview {
	return PreviewBattleWithContextMods(atk, def, terrain, types, atkMods, defMods, inferBattleContext(atk), stance)
}

// PreviewBattleWithContextMods saldırı başlamadan önce aynı matematikle muhtemel sonucu özetler.
func PreviewBattleWithContextMods(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods, context BattleContext, stance BattleStance) Preview {
	stance = NormalizeBattleStance(stance)
	context = NormalizeBattleContext(context)
	cfg := battleStanceSpec(context, stance)
	atkStr, defStr := battleStrengths(atk, def, terrain, types, atkMods, defMods, context, stance)
	buckets := previewOutcomeBuckets(atkStr, defStr, context, stance)

	preview := Preview{
		Stance:          stance,
		StanceLabelTR:   cfg.LabelTR,
		StanceSummaryTR: cfg.SummaryTR,
		AttackStrength:  int(atkStr + 0.5),
		DefenseStrength: int(defStr + 0.5),
		LikelyOutcome:   "Belirsiz",
	}
	var attackerLossExpected float64
	var defenderLossExpected float64
	var attackerHPExpected float64
	var defenderHPExpected float64

	if len(buckets) == 0 {
		return preview
	}

	maxProb := -1.0
	hasLossWindow := false
	for _, bucket := range buckets {
		if bucket.Probability <= 0 {
			continue
		}
		if bucket.AttackerWins {
			preview.WinChance += int(bucket.Probability*100 + 0.5)
		}
		if bucket.Probability > maxProb {
			maxProb = bucket.Probability
			preview.LikelyOutcome = bucket.Description
		}

		atkSummary := estimateLossSummary(atk, bucket.AttackerLossRatio, false)
		defSummary := estimateLossSummary(def, bucket.DefenderLossRatio, bucket.AttackerWins)
		atkLoss := atkSummary.LostUnits
		defLoss := defSummary.LostUnits
		attackerLossExpected += float64(atkLoss) * bucket.Probability
		defenderLossExpected += float64(defLoss) * bucket.Probability
		attackerHPExpected += float64(atkSummary.HPDamage) * bucket.Probability
		defenderHPExpected += float64(defSummary.HPDamage) * bucket.Probability

		if !hasLossWindow {
			preview.AttackerLossMin, preview.AttackerLossMax = atkLoss, atkLoss
			preview.DefenderLossMin, preview.DefenderLossMax = defLoss, defLoss
			preview.AttackerHPDamageMin, preview.AttackerHPDamageMax = atkSummary.HPDamage, atkSummary.HPDamage
			preview.DefenderHPDamageMin, preview.DefenderHPDamageMax = defSummary.HPDamage, defSummary.HPDamage
			hasLossWindow = true
			continue
		}
		if atkLoss < preview.AttackerLossMin {
			preview.AttackerLossMin = atkLoss
		}
		if atkLoss > preview.AttackerLossMax {
			preview.AttackerLossMax = atkLoss
		}
		if defLoss < preview.DefenderLossMin {
			preview.DefenderLossMin = defLoss
		}
		if defLoss > preview.DefenderLossMax {
			preview.DefenderLossMax = defLoss
		}
		if atkSummary.HPDamage < preview.AttackerHPDamageMin {
			preview.AttackerHPDamageMin = atkSummary.HPDamage
		}
		if atkSummary.HPDamage > preview.AttackerHPDamageMax {
			preview.AttackerHPDamageMax = atkSummary.HPDamage
		}
		if defSummary.HPDamage < preview.DefenderHPDamageMin {
			preview.DefenderHPDamageMin = defSummary.HPDamage
		}
		if defSummary.HPDamage > preview.DefenderHPDamageMax {
			preview.DefenderHPDamageMax = defSummary.HPDamage
		}
	}
	preview.AttackerHPExpected = int(attackerHPExpected + 0.5)
	preview.DefenderHPExpected = int(defenderHPExpected + 0.5)
	preview.AttackerLossExpected = int(attackerLossExpected + 0.5)
	preview.DefenderLossExpected = int(defenderLossExpected + 0.5)
	return preview
}

func NormalizeBattleStance(stance BattleStance) BattleStance {
	switch stance {
	case BattleStanceAggressive, BattleStanceBalanced, BattleStanceDefensive:
		return stance
	default:
		return BattleStanceBalanced
	}
}

func NormalizeBattleContext(context BattleContext) BattleContext {
	switch context {
	case BattleContextLand, BattleContextNaval, BattleContextAmphibious:
		return context
	default:
		return BattleContextLand
	}
}

func BattleStanceLabelTR(stance BattleStance) string {
	return battleStanceSpec(BattleContextLand, stance).LabelTR
}

func BattleStanceSummaryTR(stance BattleStance) string {
	return battleStanceSpec(BattleContextLand, stance).SummaryTR
}

func BattleContextLabelTR(context BattleContext) string {
	switch NormalizeBattleContext(context) {
	case BattleContextNaval:
		return "Deniz Muharebesi"
	case BattleContextAmphibious:
		return "Çıkarma Muharebesi"
	default:
		return "Kara Muharebesi"
	}
}

func battleStanceSpec(context BattleContext, stance BattleStance) battleStanceConfig {
	context = NormalizeBattleContext(context)
	switch NormalizeBattleStance(stance) {
	case BattleStanceAggressive:
		switch context {
		case BattleContextNaval:
			return battleStanceConfig{
				LabelTR:         "Agresif",
				SummaryTR:       "Yakın borda baskısını artırır, filoyu daha fazla yıpratır.",
				AttackMod:       0.14,
				AttackerLossMod: 1.18,
				DefenderLossMod: 1.14,
			}
		case BattleContextAmphibious:
			return battleStanceConfig{
				LabelTR:         "Agresif",
				SummaryTR:       "Sahil başını zorlar, çıkarma kaybını da yükseltir.",
				AttackMod:       0.12,
				AttackerLossMod: 1.24,
				DefenderLossMod: 1.12,
			}
		default:
			return battleStanceConfig{
				LabelTR:         "Agresif",
				SummaryTR:       "Zafer baskısını artırır, karşı darbeyi ağırlaştırır.",
				AttackMod:       0.18,
				AttackerLossMod: 1.20,
				DefenderLossMod: 1.10,
			}
		}
	case BattleStanceDefensive:
		switch context {
		case BattleContextNaval:
			return battleStanceConfig{
				LabelTR:         "Savunmacı",
				SummaryTR:       "Hat düzenini korur, filo kaybını azaltır ama baskıyı düşürür.",
				AttackMod:       -0.08,
				AttackerLossMod: 0.82,
				DefenderLossMod: 0.92,
			}
		case BattleContextAmphibious:
			return battleStanceConfig{
				LabelTR:         "Savunmacı",
				SummaryTR:       "Daha düzenli iner, kaybı azaltır ama kıyı baskısını düşürür.",
				AttackMod:       -0.12,
				AttackerLossMod: 0.80,
				DefenderLossMod: 0.86,
			}
		default:
			return battleStanceConfig{
				LabelTR:         "Savunmacı",
				SummaryTR:       "Daha temkinli ilerler, kaybı düşürür ama baskıyı azaltır.",
				AttackMod:       -0.10,
				AttackerLossMod: 0.78,
				DefenderLossMod: 0.88,
			}
		}
	default:
		switch context {
		case BattleContextNaval:
			return battleStanceConfig{
				LabelTR:         "Dengeli",
				SummaryTR:       "Top ve manevra arasında dengeli hat tutar.",
				AttackMod:       0,
				AttackerLossMod: 1,
				DefenderLossMod: 1,
			}
		case BattleContextAmphibious:
			return battleStanceConfig{
				LabelTR:         "Dengeli",
				SummaryTR:       "İniş temposu ile güvenliği dengede tutar.",
				AttackMod:       0,
				AttackerLossMod: 1,
				DefenderLossMod: 1,
			}
		default:
			return battleStanceConfig{
				LabelTR:         "Dengeli",
				SummaryTR:       "Standart düzen; risk ve baskı dengede tutulur.",
				AttackMod:       0,
				AttackerLossMod: 1,
				DefenderLossMod: 1,
			}
		}
	}
}

type outcomeSpec struct {
	AttackerWins      bool
	AttackerLossRatio float64
	DefenderLossRatio float64
	Description       string
}

type outcomeBucket struct {
	Probability float64
	outcomeSpec
}

func battleStrengths(atk, def *army.Army, terrain world.TerrainType, types map[string]*army.UnitType, atkMods, defMods TechMods, context BattleContext, stance BattleStance) (float64, float64) {
	atkAttackMod := atkMods.AttackMod
	defDefenseMod := defMods.DefenseMod
	commanderAttackMod, _ := atk.CommanderModifiers()
	_, commanderDefenseMod := def.CommanderModifiers()
	attackerMoraleMod := atk.CommanderMoraleModifier()
	defenderMoraleMod := def.CommanderMoraleModifier()
	if atk.IsNaval {
		atkAttackMod = atkMods.NavalAttackMod
		defDefenseMod = defMods.NavalDefenseMod
	}
	cfg := battleStanceSpec(context, stance)
	atkStr := float64(atk.TotalStrength(types)) * (1.0 + atkAttackMod + cfg.AttackMod + commanderAttackMod + attackerMoraleMod)
	defStr := float64(def.TotalStrength(types)) * terrainBonus(terrain) * (1.0 + defDefenseMod + commanderDefenseMod + defenderMoraleMod)
	if atkStr < 1 {
		atkStr = 1
	}
	if defStr < 1 {
		defStr = 1
	}
	return atkStr, defStr
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
	spec := baseOutcomeForRatio(ratio)
	return spec.AttackerWins, spec.AttackerLossRatio, spec.DefenderLossRatio
}

func resolveOutcome(atkStr, defStr float64, context BattleContext, stance BattleStance) outcomeSpec {
	attackerWins, atkLoss, defLoss := calculateOutcome(atkStr, defStr)
	return applyBattleStance(baseOutcomeSpec(attackerWins, atkLoss, defLoss), context, stance)
}

func baseOutcomeForRatio(ratio float64) outcomeSpec {
	if ratio > 1.5 {
		return outcomeSpec{AttackerWins: true, AttackerLossRatio: 0.10, DefenderLossRatio: 0.80, Description: "Ezici Zafer"}
	} else if ratio >= 1.0 {
		return outcomeSpec{AttackerWins: true, AttackerLossRatio: 0.35, DefenderLossRatio: 0.50, Description: "Dar Zafer"}
	} else if ratio >= 0.7 {
		return outcomeSpec{AttackerWins: false, AttackerLossRatio: 0.50, DefenderLossRatio: 0.30, Description: "Geri Çekilme"}
	}
	return outcomeSpec{AttackerWins: false, AttackerLossRatio: 0.80, DefenderLossRatio: 0.10, Description: "Ağır Yenilgi"}
}

func baseOutcomeSpec(wins bool, atkLoss, defLoss float64) outcomeSpec {
	return outcomeSpec{
		AttackerWins:      wins,
		AttackerLossRatio: atkLoss,
		DefenderLossRatio: defLoss,
		Description:       outcomeDescription(wins, atkLoss, defLoss),
	}
}

func applyBattleStance(spec outcomeSpec, context BattleContext, stance BattleStance) outcomeSpec {
	cfg := battleStanceSpec(context, stance)
	spec.AttackerLossRatio = clamp01(spec.AttackerLossRatio * cfg.AttackerLossMod)
	spec.DefenderLossRatio = clamp01(spec.DefenderLossRatio * cfg.DefenderLossMod)
	return spec
}

func previewOutcomeBuckets(atkStr, defStr float64, context BattleContext, stance BattleStance) []outcomeBucket {
	baseRatio := atkStr / (defStr + 1)
	if baseRatio <= 0 {
		return []outcomeBucket{{Probability: 1, outcomeSpec: applyBattleStance(baseOutcomeForRatio(0), context, stance)}}
	}
	buckets := []struct {
		min  float64
		max  float64
		inf  bool
		spec outcomeSpec
	}{
		{min: 1.5, inf: true, spec: baseOutcomeForRatio(1.6)},
		{min: 1.0, max: 1.5, spec: baseOutcomeForRatio(1.2)},
		{min: 0.7, max: 1.0, spec: baseOutcomeForRatio(0.8)},
		{max: 0.7, spec: baseOutcomeForRatio(0.6)},
	}
	result := make([]outcomeBucket, 0, len(buckets))
	for _, bucket := range buckets {
		prob := probabilityForRatioRange(baseRatio, bucket.min, bucket.max, bucket.inf)
		if prob <= 0 {
			continue
		}
		result = append(result, outcomeBucket{
			Probability: prob,
			outcomeSpec: applyBattleStance(bucket.spec, context, stance),
		})
	}
	return result
}

func inferBattleContext(atk *army.Army) BattleContext {
	if atk != nil && atk.IsNaval {
		return BattleContextNaval
	}
	return BattleContextLand
}

func probabilityForRatioRange(baseRatio, minRatio, maxRatio float64, maxInfinite bool) float64 {
	const diceMin = -0.15
	const diceMax = 0.15
	const diceSpan = diceMax - diceMin

	lower := diceMin
	if minRatio > 0 {
		bound := minRatio/baseRatio - 1
		if bound > lower {
			lower = bound
		}
	}
	upper := diceMax
	if !maxInfinite && maxRatio > 0 {
		bound := maxRatio/baseRatio - 1
		if bound < upper {
			upper = bound
		}
	}
	if upper <= lower {
		return 0
	}
	prob := (upper - lower) / diceSpan
	if prob < 0 {
		return 0
	}
	if prob > 1 {
		return 1
	}
	return prob
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// applyCasualties ordudaki birim sayısını ratio kadar azaltır.
// Ratio doğrudan "ölen birim oranı" olarak yorumlanır (HP fraksiyonu değil).
func applyCasualties(a *army.Army, ratio float64) (lost int) {
	n := len(a.Units)
	if n == 0 || ratio <= 0 {
		return 0
	}
	totalHP := 0
	for i := range a.Units {
		hp := a.Units[i].CurrentHP
		if hp < 1 {
			continue
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		totalHP += hp
	}
	if totalHP == 0 {
		a.Units = a.Units[:0]
		return n
	}

	damageBudget := int(float64(totalHP)*ratio + 0.5)
	if damageBudget <= 0 {
		return 0
	}
	if damageBudget > totalHP {
		damageBudget = totalHP
	}

	spreadDamage := damageBudget / n
	if spreadDamage > 60 {
		spreadDamage = 60
	}
	if spreadDamage > 0 {
		for i := range a.Units {
			a.Units[i].CurrentHP -= spreadDamage
		}
		damageBudget -= spreadDamage * n
	}

	start := 0
	if n > 1 {
		start = rand.Intn(n)
	}
	for damageBudget > 0 {
		target := -1
		targetHP := 0
		for step := 0; step < n; step++ {
			idx := (start + step) % n
			hp := a.Units[idx].CurrentHP
			if hp <= 0 {
				continue
			}
			if target == -1 || hp < targetHP {
				target = idx
				targetHP = hp
			}
		}
		if target == -1 {
			break
		}
		chunk := targetHP
		if damageBudget < chunk {
			chunk = damageBudget
		}
		a.Units[target].CurrentHP -= chunk
		damageBudget -= chunk
	}

	surviving := a.Units[:0]
	for _, u := range a.Units {
		if u.CurrentHP <= 0 {
			lost++
			continue
		}
		surviving = append(surviving, u)
	}
	a.Units = surviving
	return lost
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

type lossSummary struct {
	LostUnits int
	HPDamage  int
}

func estimateLossSummary(a *army.Army, ratio float64, clearSurvivors bool) lossSummary {
	if a == nil {
		return lossSummary{}
	}
	units := make([]army.Unit, len(a.Units))
	copy(units, a.Units)
	totalHP := totalUnitsHP(units)
	lost := estimateCasualties(units, ratio)
	remainingHP := totalUnitsHP(units)
	if clearSurvivors && len(units)-lost > 0 {
		lost = len(units)
		remainingHP = 0
	}
	return lossSummary{
		LostUnits: lost,
		HPDamage:  totalHP - remainingHP,
	}
}

func totalUnitsHP(units []army.Unit) int {
	totalHP := 0
	for i := range units {
		hp := units[i].CurrentHP
		if hp < 1 {
			continue
		}
		if hp > army.MaxUnitHP {
			hp = army.MaxUnitHP
		}
		totalHP += hp
	}
	return totalHP
}

func estimateCasualties(units []army.Unit, ratio float64) int {
	n := len(units)
	if n == 0 || ratio <= 0 {
		return 0
	}
	totalHP := totalUnitsHP(units)
	if totalHP == 0 {
		return n
	}

	damageBudget := int(float64(totalHP)*ratio + 0.5)
	if damageBudget <= 0 {
		return 0
	}
	if damageBudget > totalHP {
		damageBudget = totalHP
	}

	spreadDamage := damageBudget / n
	if spreadDamage > 60 {
		spreadDamage = 60
	}
	if spreadDamage > 0 {
		for i := range units {
			units[i].CurrentHP -= spreadDamage
		}
		damageBudget -= spreadDamage * n
	}

	for damageBudget > 0 {
		target := -1
		targetHP := 0
		for i := range units {
			hp := units[i].CurrentHP
			if hp <= 0 {
				continue
			}
			if target == -1 || hp < targetHP {
				target = i
				targetHP = hp
			}
		}
		if target == -1 {
			break
		}
		chunk := targetHP
		if damageBudget < chunk {
			chunk = damageBudget
		}
		units[target].CurrentHP -= chunk
		damageBudget -= chunk
	}

	lost := 0
	for i := range units {
		if units[i].CurrentHP <= 0 {
			lost++
		}
	}
	return lost
}
