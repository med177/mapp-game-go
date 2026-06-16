package tech

import (
	"encoding/json"
	"os"
	"sort"

	"mapp-game-go/internal/faction"
)

// Category teknoloji kategorisi.
type Category string

const (
	CategoryMilitary  Category = "military"
	CategoryEconomy   Category = "economy"
	CategoryDiplomacy Category = "diplomacy"
	CategoryNaval     Category = "naval"
	CategoryReligion  Category = "religion"
)

// Effects bir teknolojinin oyun mekanikleri üzerindeki etkilerini tanımlar.
type Effects struct {
	InfantryAttackMod   float64 `json:"infantry_attack_mod"`
	CavalryAttackMod    float64 `json:"cavalry_attack_mod"`
	SiegeAttackMod      float64 `json:"siege_attack_mod"`
	NavalAttackMod      float64 `json:"naval_attack_mod"`
	NavalDefenseMod     float64 `json:"naval_defense_mod"`
	LandDefenseMod      float64 `json:"land_defense_mod"`
	GoldPerRegion       int     `json:"gold_per_region"`
	GrainMod            float64 `json:"grain_mod"`
	IronMod             float64 `json:"iron_mod"`
	TimberMod           float64 `json:"timber_mod"`
	StoneMod            float64 `json:"stone_mod"`
	MarketGoldMod       float64 `json:"market_gold_mod"`
	PeaceRelationBonus  int     `json:"peace_relation_bonus"`
	RevealEnemyStrength bool    `json:"reveal_enemy_strength"`
	NavalMoveBonus      int     `json:"naval_move_bonus"`
	MoveBonus           int     `json:"move_bonus"`
	SatisfactionBonus   int     `json:"satisfaction_bonus"`
	ConversionSpeedMod  float64 `json:"conversion_speed_mod"`
}

// Technology bir araştırılabilir teknolojiyi tanımlar.
type Technology struct {
	ID            string   `json:"id"`
	NameTR        string   `json:"name_tr"`
	Category      Category `json:"category"`
	DescriptionTR string   `json:"description_tr"`
	GoldCost      int      `json:"gold_cost"`
	TurnsRequired int      `json:"turns_required"`
	Requires      []string `json:"requires"`
	Effects       Effects  `json:"effects"`
}

// IsUnlocked tüm gereksinimlerin tamamlanıp tamamlanmadığını kontrol eder.
func IsUnlocked(rs *faction.ResearchState, t *Technology) bool {
	if rs.Completed == nil {
		return len(t.Requires) == 0
	}
	for _, req := range t.Requires {
		if !rs.Completed[req] {
			return false
		}
	}
	return true
}

func CanStartResearch(rs *faction.ResearchState, t *Technology, gold int) bool {
	if !IsUnlocked(rs, t) || (rs.Completed != nil && rs.Completed[t.ID]) {
		return false
	}
	if rs.ActiveID != "" && rs.ActiveID != t.ID {
		return false
	}
	if rs.PausedTurns != nil && rs.PausedTurns[t.ID] > 0 {
		return true
	}
	return gold >= t.GoldCost
}

func PauseResearch(rs *faction.ResearchState) {
	if rs == nil || rs.ActiveID == "" || rs.TurnsLeft <= 0 {
		return
	}
	if rs.PausedTurns == nil {
		rs.PausedTurns = make(map[string]int)
	}
	rs.PausedTurns[rs.ActiveID] = rs.TurnsLeft
	rs.ActiveID = ""
	rs.TurnsLeft = 0
}

// StartResearch araştırmayı başlatır; başarısızsa false döner.
func StartResearch(rs *faction.ResearchState, t *Technology, gold *int) bool {
	if !IsUnlocked(rs, t) || (rs.Completed != nil && rs.Completed[t.ID]) {
		return false
	}
	if rs.ActiveID != "" && rs.ActiveID != t.ID {
		return false
	}
	if rs.Completed == nil {
		rs.Completed = make(map[string]bool)
	}
	if rs.PausedTurns != nil {
		if pausedTurns, ok := rs.PausedTurns[t.ID]; ok && pausedTurns > 0 {
			rs.ActiveID = t.ID
			rs.TurnsLeft = pausedTurns
			delete(rs.PausedTurns, t.ID)
			return true
		}
	}
	if *gold < t.GoldCost {
		return false
	}
	*gold -= t.GoldCost
	rs.ActiveID = t.ID
	rs.TurnsLeft = t.TurnsRequired
	return true
}

// Tick aktif araştırmayı ilerletir; tamamlanırsa tech ID'sini döner.
func Tick(rs *faction.ResearchState) string {
	if rs.ActiveID == "" {
		return ""
	}
	rs.TurnsLeft--
	if rs.TurnsLeft <= 0 {
		if rs.Completed == nil {
			rs.Completed = make(map[string]bool)
		}
		id := rs.ActiveID
		rs.Completed[id] = true
		rs.ActiveID = ""
		rs.TurnsLeft = 0
		return id
	}
	return ""
}

// NextResearchableTechID araştırılabilir ve ödenebilir teknolojiler arasından
// ağaç sırasına göre ilk uygun tech ID'sini döner.
func NextResearchableTechID(rs *faction.ResearchState, allTechs map[string]*Technology, gold int) (string, bool) {
	if rs == nil || allTechs == nil || len(rs.PausedTurns) > 0 {
		return "", false
	}

	candidates := make([]string, 0, len(allTechs))
	for id, t := range allTechs {
		if t == nil {
			continue
		}
		if rs.Completed != nil && rs.Completed[id] {
			continue
		}
		if !IsUnlocked(rs, t) {
			continue
		}
		if gold < t.GoldCost {
			continue
		}
		candidates = append(candidates, id)
	}
	if len(candidates) == 0 {
		return "", false
	}

	levels := make(map[string]int, len(candidates))
	visiting := make(map[string]bool, len(candidates))
	sort.Slice(candidates, func(i, j int) bool {
		left := allTechs[candidates[i]]
		right := allTechs[candidates[j]]
		leftLevel := techLevel(left, allTechs, levels, visiting)
		rightLevel := techLevel(right, allTechs, levels, visiting)
		if leftLevel != rightLevel {
			return leftLevel < rightLevel
		}
		if left.Category != right.Category {
			return CategoryOrder(left.Category) < CategoryOrder(right.Category)
		}
		return left.ID < right.ID
	})
	return candidates[0], true
}

func techLevel(t *Technology, allTechs map[string]*Technology, memo map[string]int, visiting map[string]bool) int {
	if t == nil {
		return 0
	}
	if level, ok := memo[t.ID]; ok {
		return level
	}
	if visiting[t.ID] {
		return 0
	}
	visiting[t.ID] = true
	defer delete(visiting, t.ID)

	if len(t.Requires) == 0 {
		memo[t.ID] = 0
		return 0
	}

	maxReqLevel := 0
	for _, reqID := range t.Requires {
		req, ok := allTechs[reqID]
		if !ok || req == nil {
			continue
		}
		reqLevel := techLevel(req, allTechs, memo, visiting)
		if reqLevel > maxReqLevel {
			maxReqLevel = reqLevel
		}
	}
	level := maxReqLevel + 1
	memo[t.ID] = level
	return level
}

// ComputeEffects tamamlanmış teknolojilerin kümülatif etkilerini hesaplar.
func ComputeEffects(completed map[string]bool, allTechs map[string]*Technology) Effects {
	var total Effects
	for id := range completed {
		t, ok := allTechs[id]
		if !ok {
			continue
		}
		e := t.Effects
		total.InfantryAttackMod += e.InfantryAttackMod
		total.CavalryAttackMod += e.CavalryAttackMod
		total.SiegeAttackMod += e.SiegeAttackMod
		total.NavalAttackMod += e.NavalAttackMod
		total.NavalDefenseMod += e.NavalDefenseMod
		total.LandDefenseMod += e.LandDefenseMod
		total.GoldPerRegion += e.GoldPerRegion
		total.GrainMod += e.GrainMod
		total.IronMod += e.IronMod
		total.TimberMod += e.TimberMod
		total.StoneMod += e.StoneMod
		total.MarketGoldMod += e.MarketGoldMod
		total.PeaceRelationBonus += e.PeaceRelationBonus
		if e.RevealEnemyStrength {
			total.RevealEnemyStrength = true
		}
		total.NavalMoveBonus += e.NavalMoveBonus
		total.MoveBonus += e.MoveBonus
		total.SatisfactionBonus += e.SatisfactionBonus
		total.ConversionSpeedMod += e.ConversionSpeedMod
	}
	return total
}

// LoadTechnologies JSON dosyasından teknoloji listesini yükler.
func LoadTechnologies(path string) (map[string]*Technology, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var list []*Technology
	if err := json.NewDecoder(f).Decode(&list); err != nil {
		return nil, err
	}

	out := make(map[string]*Technology, len(list))
	for _, t := range list {
		out[t.ID] = t
	}
	return out, nil
}
