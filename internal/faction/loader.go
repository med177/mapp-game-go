package faction

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"mapp-game-go/internal/religion"
)

type relationDefinition struct {
	FactionA  FactionID        `json:"faction_a"`
	FactionB  FactionID        `json:"faction_b"`
	Score     *int             `json:"score"`
	ScoreAToB *int             `json:"score_a_to_b"`
	ScoreBToA *int             `json:"score_b_to_a"`
	Stance    DiplomaticStance `json:"stance"`
}

// LoadFactionsWithOrder assets/data/factions.json dosyasını okur, map ve dosya sırasını döner.
func LoadFactionsWithOrder(path string) (map[FactionID]*Faction, []FactionID, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("factions dosyası okunamadı: %w", err)
	}

	var list []*Faction
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, fmt.Errorf("factions JSON parse hatası: %w", err)
	}

	result := make(map[FactionID]*Faction, len(list))
	order := make([]FactionID, 0, len(list))
	for _, f := range list {
		if f == nil {
			continue
		}
		result[f.ID] = f
		order = append(order, f.ID)
	}
	return result, order, nil
}

// LoadFactions assets/data/factions.json dosyasını okur ve map döner.
func LoadFactions(path string) (map[FactionID]*Faction, error) {
	result, _, err := LoadFactionsWithOrder(path)
	return result, err
}

// LoadRelations başlangıç diplomasi ilişkilerini JSON'dan okur.
// Dosya yoksa din temelli varsayılan ilişkiler döner.
func LoadRelations(path string, factions map[FactionID]*Faction) (map[string]*Relation, error) {
	result, _, err := LoadRelationsWithOrder(path, factions)
	return result, err
}

// LoadRelationsWithOrder ilişkileri JSON'dan yükler ve kaynak dosyadaki geçerli
// ilişki sırasını ayrıca döner. Edit Mode kaydında bu sıra korunmalıdır.
func LoadRelationsWithOrder(path string, factions map[FactionID]*Faction) (map[string]*Relation, []string, error) {
	return LoadRelationsWithOrderForRegistry(path, factions, religion.DefaultRegistry())
}

// LoadRelationsWithOrderForRegistry ilişkileri aktif senaryonun din registry'siyle yükler.
func LoadRelationsWithOrderForRegistry(path string, factions map[FactionID]*Faction, registry *religion.Registry) (map[string]*Relation, []string, error) {
	relations := BuildInitialRelationsForRegistry(factions, registry)
	// Varsayılan runtime ilişkileri kaynak ilişkisi değildir. Dosya yoksa veya
	// dosyada bulunmuyorsa kaydetme sırasında JSON'a yazılmamalıdır.
	order := make([]string, 0)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return relations, order, nil
		}
		return nil, nil, fmt.Errorf("relations dosyası okunamadı: %w", err)
	}

	var list []*relationDefinition
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, fmt.Errorf("relations JSON parse hatası: %w", err)
	}
	for _, rel := range list {
		if rel == nil {
			continue
		}
		if factions[rel.FactionA] == nil || factions[rel.FactionB] == nil || rel.FactionA == rel.FactionB {
			continue
		}
		key := RelationKey(rel.FactionA, rel.FactionB)
		scoreAToB, scoreBToA := 0, 0
		if rel.Score != nil {
			scoreAToB = *rel.Score
			scoreBToA = *rel.Score
		}
		if rel.ScoreAToB != nil {
			scoreAToB = *rel.ScoreAToB
		}
		if rel.ScoreBToA != nil {
			scoreBToA = *rel.ScoreBToA
		}
		relations[key] = &Relation{
			FactionA:  rel.FactionA,
			FactionB:  rel.FactionB,
			ScoreAToB: scoreAToB,
			ScoreBToA: scoreBToA,
			Stance:    normalizeStance(rel.Stance),
		}
		order = append(order, key)
	}
	return relations, order, nil
}

func normalizeStance(stance DiplomaticStance) DiplomaticStance {
	return NormalizeStance(stance)
}

// BuildInitialRelations fraksiyonlar arasındaki başlangıç diplomatik ilişkilerini oluşturur.
func BuildInitialRelations(factions map[FactionID]*Faction) map[string]*Relation {
	return BuildInitialRelationsForRegistry(factions, religion.DefaultRegistry())
}

// BuildInitialRelationsForRegistry başlangıç puanlarını senaryo din verisinden üretir.
func BuildInitialRelationsForRegistry(factions map[FactionID]*Faction, registry *religion.Registry) map[string]*Relation {
	relations := make(map[string]*Relation)

	ids := make([]FactionID, 0, len(factions))
	for id := range factions {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := factions[ids[i]]
			b := factions[ids[j]]

			key := RelationKey(a.ID, b.ID)
			relations[key] = &Relation{
				FactionA:  a.ID,
				FactionB:  b.ID,
				ScoreAToB: DefaultRelationScoreForRegistry(a, b, registry),
				ScoreBToA: DefaultRelationScoreForRegistry(a, b, registry),
				Stance:    StancePeace,
			}
		}
	}
	return relations
}

// DefaultRelationScore, relations.json içinde kaydı olmayan çiftlerin
// başlangıç puanını belirler. Özel tarihsel ilişkiler JSON'da açıkça tutulur.
func DefaultRelationScore(a, b *Faction) int {
	return DefaultRelationScoreForRegistry(a, b, religion.DefaultRegistry())
}

// DefaultRelationScoreForRegistry relations.json içinde kaydı olmayan çiftin
// başlangıç puanını aktif senaryonun din verisinden belirler.
func DefaultRelationScoreForRegistry(a, b *Faction, registry *religion.Registry) int {
	if a != nil && b != nil && a.Religion == b.Religion {
		return 25
	}
	if a == nil || b == nil {
		return -30
	}
	return registry.Relation(a.Religion, b.Religion)
}
