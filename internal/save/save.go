package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
)

const saveDir = "saves"
const autoSavePath = "saves/autosave.json"

// Save oyun durumunu JSON dosyasına yazar.
func Save(gs *state.GameState) error {
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("kayıt dizini oluşturulamadı: %w", err)
	}
	data, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return fmt.Errorf("serileştirme hatası: %w", err)
	}
	if err := os.WriteFile(autoSavePath, data, 0644); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}
	return nil
}

// Load JSON dosyasından oyun durumunu yükler ve runtime alanlarını doldurur.
func Load(unitTypesPath, buildingsPath string) (*state.GameState, error) {
	data, err := os.ReadFile(autoSavePath)
	if err != nil {
		return nil, fmt.Errorf("kayıt dosyası bulunamadı (%s): %w", filepath.Base(autoSavePath), err)
	}
	var gs state.GameState
	if err := json.Unmarshal(data, &gs); err != nil {
		return nil, fmt.Errorf("kayıt dosyası okunamadı: %w", err)
	}

	// Runtime alanları yeniden yükle (json:"-" olanlar)
	unitTypes, err := army.LoadUnitTypes(unitTypesPath)
	if err != nil {
		return nil, err
	}
	gs.UnitTypes = unitTypes

	buildingTypes, err := city.LoadBuildings(buildingsPath)
	if err != nil {
		return nil, err
	}
	gs.BuildingTypes = buildingTypes

	techTypes, err := tech.LoadTechnologies("assets/data/technologies.json")
	if err != nil {
		return nil, err
	}
	gs.TechTypes = techTypes

	return &gs, nil
}

// SaveExists otomatik kayıt dosyasının var olup olmadığını kontrol eder.
func SaveExists() bool {
	_, err := os.Stat(autoSavePath)
	return err == nil
}
