package scenario

import (
	"encoding/json"
	"fmt"
	"os"
)

// AIStrategyConfig bir senaryonun statik AI profillerini taşır. Bu veri
// campaign save'ine yazılmaz; senaryo baz state'i kurulurken tekrar yüklenir.
type AIStrategyConfig struct {
	DifficultyPolicy AIDifficultyPolicy  `json:"difficulty_policy,omitempty"`
	Factions         []AIFactionStrategy `json:"factions"`
}

// AIDifficultyPolicy zorluk seviyesini kaynak hilesinden çok karar kalitesiyle
// ayıran, senaryo bazlı runtime politikadır.
type AIDifficultyPolicy struct {
	FairMovement bool                         `json:"fair_movement,omitempty"`
	Levels       map[string]AIDifficultyLevel `json:"levels,omitempty"`
}

// AIDifficultyLevel tek bir zorluk seviyesinin planlama ve küçük başlangıç
// tamponı parametrelerini taşır.
type AIDifficultyLevel struct {
	PlanHorizonTurns       int  `json:"plan_horizon_turns,omitempty"`
	PlanTargetRegionLimit  int  `json:"plan_target_region_limit,omitempty"`
	PathSearchDepth        int  `json:"path_search_depth,omitempty"`
	PlanMoveBonusPercent   int  `json:"plan_move_bonus_percent,omitempty"`
	ProactiveWar           bool `json:"proactive_war,omitempty"`
	WarThreshold           int  `json:"war_threshold,omitempty"`
	MinAttackPowerPercent  int  `json:"min_attack_power_percent,omitempty"`
	WarCadenceTurns        int  `json:"war_cadence_turns,omitempty"`
	MaxConcurrentWars      int  `json:"max_concurrent_wars,omitempty"`
	PlayerTargetScoreBonus int  `json:"player_target_score_bonus,omitempty"`
	StartGoldBuffer        int  `json:"start_gold_buffer,omitempty"`
	StartGrainBuffer       int  `json:"start_grain_buffer,omitempty"`
}

// Level sayısal oyun zorluğunun tanımlı politikasını döner.
func (p AIDifficultyPolicy) Level(difficulty int) (AIDifficultyLevel, bool) {
	if len(p.Levels) == 0 {
		return AIDifficultyLevel{}, false
	}
	level, ok := p.Levels[fmt.Sprintf("%d", difficulty)]
	return level, ok
}

// LoadedAIConfig parse edilmiş profil index'i ile zorluk politikasını bir arada
// taşır; ikisi de save'e yazılmadan senaryo baz state'inde tutulur.
type LoadedAIConfig struct {
	Strategies       map[string]AIFactionStrategy
	DifficultyPolicy AIDifficultyPolicy
}

// AIFactionStrategy tek bir devletin uzun vadeli yönelimini ve sıralı
// objective tanımlarını içerir.
type AIFactionStrategy struct {
	FactionID string `json:"faction_id"`
	Profile   string `json:"profile,omitempty"`
	// NavalFocus, denizci devletlerin kıyı sayısından bağımsız olarak büyük
	// savaş filosu kurmasını sağlayan senaryo-verisi odak bayrağıdır.
	NavalFocus        bool                    `json:"naval_focus,omitempty"`
	ExpansionTargets  []string                `json:"expansion_targets,omitempty"`
	TerritorialClaims []AITerritorialClaimDef `json:"territorial_claims,omitempty"`
	Objectives        []AIObjectiveDef        `json:"objectives"`
}

// AITerritorialClaimDef AI stratejisindeki tarihsel claim ağırlığını taşır.
// Başlangıç core'ları burada tutulmaz; onlar başlangıç sahipliğinden türetilir.
type AITerritorialClaimDef struct {
	RegionID string `json:"region_id"`
	Value    int    `json:"value"`
}

// AIObjectiveDef tarihsel yönü puanlamaya ekler. MinYear, MaxYear ve
// RequiredEventFlags hard gate'tir; MaxYear verilen yılın sonuna kadar
// geçerlidir. Diğer alanlar AI'yi yönlendirir ama güç, cephe ve diplomasi
// güvenlik kontrollerini geçersiz kılmaz.
type AIObjectiveDef struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	// TerritorialClaims objective'in bölgesel hedeflerini ve ağırlıklarını tek
	// yerde taşır. Bölgenin o anki sahibi ayrıca yazılmaz; AI bunu runtime'da
	// bölgenin OwnerID değerinden bulur.
	TerritorialClaims []AITerritorialClaimDef `json:"territorial_claims,omitempty"`
	// Eski scenario verileri için geriye uyumluluk alanları. Yeni veride
	// TerritorialClaims kullanılmalıdır.
	TargetFactions     []string `json:"target_factions,omitempty"`
	TargetRegions      []string `json:"target_regions,omitempty"`
	Priority           int      `json:"priority"`
	Commitment         int      `json:"commitment,omitempty"`
	MinYear            int      `json:"min_year,omitempty"`
	MaxYear            int      `json:"max_year,omitempty"`
	RequiredEventFlags []string `json:"required_event_flags,omitempty"`
	ReadinessRegions   []string `json:"readiness_regions,omitempty"`
	AllowVassalization bool     `json:"allow_vassalization,omitempty"`
}

// LoadAIConfig opsiyonel ai_strategies.json dosyasını yükler. Dosya
// bulunmayan senaryolar mevcut genel AI davranışını kullanmaya devam eder.
func LoadAIConfig(path string) (LoadedAIConfig, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return LoadedAIConfig{}, nil
	}
	if err != nil {
		return LoadedAIConfig{}, fmt.Errorf("AI stratejileri okunamadı: %w", err)
	}

	var config AIStrategyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return LoadedAIConfig{}, fmt.Errorf("AI stratejileri parse edilemedi: %w", err)
	}
	for difficulty, level := range config.DifficultyPolicy.Levels {
		if difficulty != "1" && difficulty != "2" && difficulty != "3" {
			return LoadedAIConfig{}, fmt.Errorf("AI zorluk seviyesi geçersiz: %s", difficulty)
		}
		if level.PlanHorizonTurns <= 0 || level.PlanTargetRegionLimit <= 0 || level.PathSearchDepth <= 0 || level.PlanMoveBonusPercent <= 0 || level.WarThreshold <= 0 || level.MinAttackPowerPercent <= 0 || level.WarCadenceTurns <= 0 || level.MaxConcurrentWars <= 0 {
			return LoadedAIConfig{}, fmt.Errorf("AI zorluk politikası eksik veya geçersiz: level=%s", difficulty)
		}
		if level.StartGoldBuffer < 0 || level.StartGrainBuffer < 0 {
			return LoadedAIConfig{}, fmt.Errorf("AI başlangıç tamponı negatif olamaz: level=%s", difficulty)
		}
	}

	strategies := make(map[string]AIFactionStrategy, len(config.Factions))
	for _, strategy := range config.Factions {
		if strategy.FactionID == "" {
			return LoadedAIConfig{}, fmt.Errorf("AI stratejisinde faction_id boş")
		}
		if _, duplicate := strategies[strategy.FactionID]; duplicate {
			return LoadedAIConfig{}, fmt.Errorf("AI stratejisi birden fazla tanımlanmış: %s", strategy.FactionID)
		}
		claimIDs := make(map[string]struct{}, len(strategy.TerritorialClaims))
		for _, claim := range strategy.TerritorialClaims {
			if claim.RegionID == "" || claim.Value < 1 || claim.Value > 100 {
				return LoadedAIConfig{}, fmt.Errorf("AI territorial claim geçersiz: faction=%s region=%s value=%d", strategy.FactionID, claim.RegionID, claim.Value)
			}
			if _, duplicate := claimIDs[claim.RegionID]; duplicate {
				return LoadedAIConfig{}, fmt.Errorf("AI territorial claim birden fazla tanımlanmış: faction=%s region=%s", strategy.FactionID, claim.RegionID)
			}
			claimIDs[claim.RegionID] = struct{}{}
		}
		objectiveIDs := make(map[string]struct{}, len(strategy.Objectives))
		for _, objective := range strategy.Objectives {
			if objective.MaxYear > 0 && objective.MinYear > 0 && objective.MaxYear < objective.MinYear {
				return LoadedAIConfig{}, fmt.Errorf("AI objective max_year min_year'dan önce: faction=%s objective=%s", strategy.FactionID, objective.ID)
			}
			claimIDs := make(map[string]struct{}, len(objective.TerritorialClaims))
			for _, claim := range objective.TerritorialClaims {
				if claim.RegionID == "" || claim.Value < 1 || claim.Value > 100 {
					return LoadedAIConfig{}, fmt.Errorf("AI objective territorial claim geçersiz: faction=%s objective=%s region=%s value=%d", strategy.FactionID, objective.ID, claim.RegionID, claim.Value)
				}
				if _, duplicate := claimIDs[claim.RegionID]; duplicate {
					return LoadedAIConfig{}, fmt.Errorf("AI objective territorial claim birden fazla tanımlanmış: faction=%s objective=%s region=%s", strategy.FactionID, objective.ID, claim.RegionID)
				}
				claimIDs[claim.RegionID] = struct{}{}
			}
			if objective.ID == "" {
				return LoadedAIConfig{}, fmt.Errorf("AI objective kimliği boş: faction=%s", strategy.FactionID)
			}
			// ally, diplomatik yönelim metadata'sıdır; askeri planlayıcı bunu
			// doğrudan bir saldırı/savunma planına çevirmemelidir.
			if objective.Kind != "expand" && objective.Kind != "defend" && objective.Kind != "consolidate" && objective.Kind != "ally" {
				return LoadedAIConfig{}, fmt.Errorf("AI objective türü geçersiz: faction=%s objective=%s kind=%s", strategy.FactionID, objective.ID, objective.Kind)
			}
			if _, duplicate := objectiveIDs[objective.ID]; duplicate {
				return LoadedAIConfig{}, fmt.Errorf("AI objective birden fazla tanımlanmış: faction=%s objective=%s", strategy.FactionID, objective.ID)
			}
			objectiveIDs[objective.ID] = struct{}{}
		}
		strategies[strategy.FactionID] = strategy
	}
	return LoadedAIConfig{Strategies: strategies, DifficultyPolicy: config.DifficultyPolicy}, nil
}

// LoadAIStrategies yalnız profil index'ine ihtiyaç duyan doğrulama ve araçlar
// için geriye uyumlu kısa yoldur.
func LoadAIStrategies(path string) (map[string]AIFactionStrategy, error) {
	config, err := LoadAIConfig(path)
	return config.Strategies, err
}
