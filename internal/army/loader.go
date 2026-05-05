package army

import (
	"encoding/json"
	"fmt"
	"os"
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
		m[t.ID] = t
	}
	return m, nil
}

// MakeUnits belirtilen tip ve sayıda yeni birim listesi oluşturur.
func MakeUnits(typeID string, count int) []Unit {
	units := make([]Unit, count)
	for i := range units {
		units[i] = Unit{TypeID: typeID, CurrentHP: 100, Experience: 0}
	}
	return units
}
