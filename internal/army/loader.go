package army

import (
	"encoding/json"
	"fmt"
	"os"

	"mapp-game-go/internal/world"
)

// LoadUnitTypes birim tiplerini JSON'dan yükler ve ID'ye göre indeksler.
func LoadUnitTypes(path string) (map[string]*UnitType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("birim tipleri okunamadı: %w", err)
	}
	var list []*UnitType
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("birim tipleri parse edilemedi: %w", err)
	}
	m := make(map[string]*UnitType, len(list))
	for _, t := range list {
		if t.TurnsRequired <= 0 {
			t.TurnsRequired = 1
		}
		m[t.ID] = t
	}
	return m, nil
}

// armySpecJSON armies.json'daki tek ordu tanımını temsil eder.
type armySpecJSON struct {
	ID      string          `json:"id"`
	OwnerID string          `json:"owner_id"`
	Region  world.RegionID  `json:"region_id"`
	Units   []unitCountJSON `json:"units"`
}

type unitCountJSON struct {
	TypeID string `json:"type_id"`
	Count  int    `json:"count"`
}

// LoadArmies armies.json'dan başlangıç ordularını yükler.
func LoadArmies(path string) (map[ArmyID]*Army, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ordular okunamadı: %w", err)
	}
	var specs []armySpecJSON
	if err := json.Unmarshal(data, &specs); err != nil {
		return nil, fmt.Errorf("ordular parse edilemedi: %w", err)
	}
	armies := make(map[ArmyID]*Army, len(specs))
	for _, s := range specs {
		var units []Unit
		for _, uc := range s.Units {
			units = append(units, MakeUnits(uc.TypeID, uc.Count)...)
		}
		id := ArmyID(s.ID)
		armies[id] = &Army{
			ID:            id,
			OwnerID:       s.OwnerID,
			RegionID:      s.Region,
			Units:         units,
			MovePoints:    2,
			MaxMovePoints: 2,
		}
	}
	return armies, nil
}

// MakeUnits belirtilen tip ve sayıda yeni birim listesi oluşturur.
func MakeUnits(typeID string, count int) []Unit {
	units := make([]Unit, count)
	for i := range units {
		units[i] = Unit{TypeID: typeID, CurrentHP: 100, Experience: 0}
	}
	return units
}
