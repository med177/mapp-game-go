package world

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadRegions assets/data/regions.json dosyasını okur ve map döner.
func LoadRegions(path string) (map[RegionID]*Region, error) {
	result, _, err := LoadRegionsWithOrder(path)
	return result, err
}

func LoadRegionsWithOrder(path string) (map[RegionID]*Region, []RegionID, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("regions dosyası okunamadı: %w", err)
	}

	var list []*Region
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, nil, fmt.Errorf("regions JSON parse hatası: %w", err)
	}

	result := make(map[RegionID]*Region, len(list))
	order := make([]RegionID, 0, len(list))
	for _, r := range list {
		result[r.ID] = r
		order = append(order, r.ID)
	}
	return result, order, nil
}

// countryShapeEntry JSON dosyasındaki tek bir ülke girişini temsil eder.
type countryShapeEntry struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Rings [][][2]float32 `json:"rings"`
}

// countryShapesFile JSON dosyasının kök yapısı.
type countryShapesFile struct {
	Shapes []countryShapeEntry `json:"shapes"`
}

// ShapeBounds poligon koordinatlarının sınır değerlerini tutar.
type ShapeBounds struct {
	MinX, MinY, MaxX, MaxY float32
}

// CountryShapeJSON işlenmiş harita poligon verilerini tutar.
type CountryShapeJSON struct {
	Shapes map[string][][][2]float32
	Names  map[string]string
	Bounds ShapeBounds
}

// LoadCountryShapes poligon verilerini JSON'dan okur, bölgelere atar ve sınırları hesaplar.
func LoadCountryShapes(path string, regions map[RegionID]*Region) (CountryShapeJSON, error) {
	var result CountryShapeJSON

	data, err := os.ReadFile(path)
	if err != nil {
		return result, fmt.Errorf("shapes dosyası okunamadı: %w", err)
	}

	var file countryShapesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return result, fmt.Errorf("shapes JSON parse hatası: %w", err)
	}

	// id → rings map'i oluştur
	shapeMap := make(map[string][][][2]float32, len(file.Shapes))
	nameMap := make(map[string]string, len(file.Shapes))
	for _, entry := range file.Shapes {
		shapeMap[entry.ID] = entry.Rings
		if entry.Name != "" {
			nameMap[entry.ID] = entry.Name
		}
	}
	result.Shapes = shapeMap
	result.Names = nameMap

	// Tüm koordinatların sınırlarını hesapla
	first := true
	for _, rings := range shapeMap {
		for _, ring := range rings {
			for _, pt := range ring {
				if first {
					result.Bounds.MinX = pt[0]
					result.Bounds.MaxX = pt[0]
					result.Bounds.MinY = pt[1]
					result.Bounds.MaxY = pt[1]
					first = false
				} else {
					if pt[0] < result.Bounds.MinX {
						result.Bounds.MinX = pt[0]
					}
					if pt[0] > result.Bounds.MaxX {
						result.Bounds.MaxX = pt[0]
					}
					if pt[1] < result.Bounds.MinY {
						result.Bounds.MinY = pt[1]
					}
					if pt[1] > result.Bounds.MaxY {
						result.Bounds.MaxY = pt[1]
					}
				}
			}
		}
	}

	// Bölgelere poligon ata
	for _, region := range regions {
		if rings, ok := shapeMap[region.ShapeID]; ok {
			region.Shape = rings
		}
	}

	return result, nil
}
