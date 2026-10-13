package army

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadCharacterTemplates senaryo characters.json dosyasındaki kayıtları fraksiyon ID'sine göre indeksler.
// Eski senaryoların commanders.json dosyası da geriye dönük olarak okunur.
func LoadCharacterTemplates(path string) (map[string][]*Commander, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) && filepath.Base(path) == "characters.json" {
		data, err = os.ReadFile(filepath.Join(filepath.Dir(path), "commanders.json"))
	}
	if err != nil {
		if os.IsNotExist(err) {
			return map[string][]*Commander{}, nil
		}
		return nil, fmt.Errorf("komutan şablonları okunamadı: %w", err)
	}

	var list []*Commander
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("characters JSON parse edilemedi: %w", err)
	}

	result := make(map[string][]*Commander)
	for _, commander := range list {
		if commander == nil || commander.ID == "" || commander.OwnerID == "" || commander.Name == "" {
			continue
		}
		if commander.Type != "" && !commander.Type.IsValid() {
			continue
		}
		commander.AssignedArmyID = ""
		commander.Normalize()
		result[commander.OwnerID] = append(result[commander.OwnerID], commander)
	}
	return result, nil
}

// LoadCommanderTemplates eski çağrı noktaları için uyumluluk sarmalayıcısıdır.
func LoadCommanderTemplates(path string) (map[string][]*Commander, error) {
	return LoadCharacterTemplates(path)
}
