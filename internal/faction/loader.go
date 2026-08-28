package faction

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"mapp-game-go/internal/religion"
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
	relations := BuildInitialRelations(factions)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return relations, nil
		}
		return nil, fmt.Errorf("relations dosyası okunamadı: %w", err)
	}

	var list []*Relation
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("relations JSON parse hatası: %w", err)
	}
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
	}
	return relations, nil
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
			score := religion.Relation(a.Religion, b.Religion)

			stance := StancePeace
			// Sünni-Şii arasını baştan gergin başlat
			if !a.IsEliminated && !b.IsEliminated && ((a.Religion == religion.Sunni && b.Religion == religion.Shia) ||
				(a.Religion == religion.Shia && b.Religion == religion.Sunni)) {
				stance = StanceWar
			}

			key := RelationKey(a.ID, b.ID)
			relations[key] = &Relation{
				FactionA: a.ID,
				FactionB: b.ID,
				Score:    score,
				Stance:   stance,
			}
		}
	}
	return relations
}
