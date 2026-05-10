package faction

import (
	"encoding/json"
	"fmt"
	"os"

	"mapp-game-go/internal/religion"
)

// LoadFactions assets/data/factions.json dosyasını okur ve map döner.
func LoadFactions(path string) (map[FactionID]*Faction, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("factions dosyası okunamadı: %w", err)
	}

	var list []*Faction
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("factions JSON parse hatası: %w", err)
	}

	result := make(map[FactionID]*Faction, len(list))
	for _, f := range list {
		result[f.ID] = f
	}
	return result, nil
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
	switch stance {
	case StanceWar, StancePeace, StanceAllied, StanceTrade:
		return stance
	default:
		return StancePeace
	}
}

// BuildInitialRelations fraksiyonlar arasındaki başlangıç diplomatik ilişkilerini oluşturur.
func BuildInitialRelations(factions map[FactionID]*Faction) map[string]*Relation {
	relations := make(map[string]*Relation)

	ids := make([]FactionID, 0, len(factions))
	for id := range factions {
		ids = append(ids, id)
	}

	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := factions[ids[i]]
			b := factions[ids[j]]
			score := religion.Relation(a.Religion, b.Religion)

			stance := StancePeace
			// Sünni-Şii arasını baştan gergin başlat
			if (a.Religion == religion.Sunni && b.Religion == religion.Shia) ||
				(a.Religion == religion.Shia && b.Religion == religion.Sunni) {
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
