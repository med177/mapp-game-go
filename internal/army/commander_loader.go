package army

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadCommanderTemplates senaryo commanders.json dosyasındaki başlangıç
// komutanlarını fraksiyon ID'sine göre indeksler. Bu kayıtlar runtime kariyer
// state'inden ayrıdır; oyun başında clone edilerek GameState.Commanders'a alınır.
func LoadCommanderTemplates(path string) (map[string][]*Commander, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string][]*Commander{}, nil
		}
		return nil, fmt.Errorf("komutan şablonları okunamadı: %w", err)
	}

	var list []*Commander
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("commanders JSON parse edilemedi: %w", err)
	}

	result := make(map[string][]*Commander)
	for _, commander := range list {
		if commander == nil || commander.ID == "" || commander.OwnerID == "" || commander.Name == "" {
			continue
		}
		commander.AssignedArmyID = ""
		commander.Normalize()
		result[commander.OwnerID] = append(result[commander.OwnerID], commander)
	}
	return result, nil
}
