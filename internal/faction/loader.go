package faction

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

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
	relations := BuildInitialRelations(factions)
	order := make([]string, 0, len(relations))
	for key := range relations {
		order = append(order, key)
	}
	sort.Strings(order)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return relations, order, nil
		}
		return nil, nil, fmt.Errorf("relations dosyası okunamadı: %w", err)
	}

	var list []*Relation
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, fmt.Errorf("relations JSON parse hatası: %w", err)
	}
	order = order[:0]
	for _, rel := range list {
		if rel == nil {
			continue
		}
		if factions[rel.FactionA] == nil || factions[rel.FactionB] == nil || rel.FactionA == rel.FactionB {
			continue
		}
		if rel.Stance == StanceWar && (factions[rel.FactionA].IsEliminated || factions[rel.FactionB].IsEliminated) {
			// Elenmiş ardıl devletler başlangıç diplomasi savaşlarına katılmaz.
			// İleride yeniden kurulduklarında ilişki normal diplomasi akışıyla açılır.
			rel.Stance = StancePeace
		}
		key := RelationKey(rel.FactionA, rel.FactionB)
		relations[key] = &Relation{
			FactionA: rel.FactionA,
			FactionB: rel.FactionB,
			Score:    rel.Score,
			Stance:   normalizeStance(rel.Stance),
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
				FactionA: a.ID,
				FactionB: b.ID,
				Score:    DefaultRelationScore(a, b),
				Stance:   StancePeace,
			}
		}
	}
	return relations
}

// DefaultRelationScore, relations.json içinde kaydı olmayan çiftlerin
// başlangıç puanını belirler. Özel tarihsel ilişkiler JSON'da açıkça tutulur.
func DefaultRelationScore(a, b *Faction) int {
	if a != nil && b != nil && a.Religion == b.Religion {
		return 25
	}
	return -30
}
