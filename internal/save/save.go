package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const saveDir = "saves"
const autoSavePath = "saves/autosave.json"

var scenarioBaseDir = filepath.Join("assets", "scenarios")

// slotDefs tüm kayıt slotlarını tanımlar; sıra UI'da gösterim sırasıdır.
var slotDefs = []struct {
	name        string
	displayName string
	path        string
}{
	{"autosave", "Otomatik Kayıt", "saves/autosave.json"},
	{"slot1", "Kayıt 1", "saves/slot1.json"},
	{"slot2", "Kayıt 2", "saves/slot2.json"},
	{"slot3", "Kayıt 3", "saves/slot3.json"},
}

// SaveSlot bir kayıt slotunun metadata'sını taşır.
type SaveSlot struct {
	Name        string
	DisplayName string
	Path        string
	Exists      bool
	FactionName string
	Turn        int
	Year        int
	ModTime     time.Time
}

// metaFields sadece metadata okumak için minimal struct.
type metaFields struct {
	Turn            int    `json:"turn"`
	Year            int    `json:"year"`
	PlayerFactionID string `json:"player_faction_id"`
	Factions        map[string]struct {
		NameTR string `json:"name_tr"`
	} `json:"factions"`
}

// ListSlots tüm slotların mevcut durumunu döner.
func ListSlots() []SaveSlot {
	slots := make([]SaveSlot, len(slotDefs))
	for i, def := range slotDefs {
		s := SaveSlot{
			Name:        def.name,
			DisplayName: def.displayName,
			Path:        def.path,
		}
		fi, err := os.Stat(def.path)
		if err == nil {
			s.Exists = true
			s.ModTime = fi.ModTime()
			if data, err := os.ReadFile(def.path); err == nil {
				var m metaFields
				if json.Unmarshal(data, &m) == nil {
					s.Turn = m.Turn
					s.Year = m.Year
					if f, ok := m.Factions[m.PlayerFactionID]; ok {
						s.FactionName = f.NameTR
					}
				}
			}
		}
		slots[i] = s
	}
	return slots
}

// AnySlotExists en az bir kayıt dosyası olup olmadığını döner.
func AnySlotExists() bool {
	for _, def := range slotDefs {
		if _, err := os.Stat(def.path); err == nil {
			return true
		}
	}
	return false
}

// SaveToSlot oyun durumunu isimli slota yazar.
func SaveToSlot(gs *state.GameState, slotName string) error {
	path := autoSavePath
	for _, def := range slotDefs {
		if def.name == slotName {
			path = def.path
			break
		}
	}
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("kayıt dizini oluşturulamadı: %w", err)
	}
	data, err := json.MarshalIndent(gs, "", "  ")
	if err != nil {
		return fmt.Errorf("serileştirme hatası: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}
	return nil
}

// LoadSlot isimli slottan oyun durumunu yükler.
func LoadSlot(slotName string) (*state.GameState, error) {
	path := autoSavePath
	for _, def := range slotDefs {
		if def.name == slotName {
			path = def.path
			break
		}
	}
	return loadFromPath(path)
}

// Save otomatik kayıt slotuna yazar (geriye dönük uyumluluk).
func Save(gs *state.GameState) error {
	return SaveToSlot(gs, "autosave")
}

// Load otomatik kayıt slotundan yükler (geriye dönük uyumluluk).
func Load() (*state.GameState, error) {
	return LoadSlot("autosave")
}

// DeleteSlot isimli slot dosyasını siler.
func DeleteSlot(slotName string) error {
	for _, def := range slotDefs {
		if def.name == slotName {
			if err := os.Remove(def.path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("kayıt silinemedi: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("bilinmeyen slot: %s", slotName)
}

// SaveExists otomatik kayıt dosyasının var olup olmadığını kontrol eder.
func SaveExists() bool {
	_, err := os.Stat(autoSavePath)
	return err == nil
}

func loadFromPath(path string) (*state.GameState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kayıt dosyası bulunamadı (%s): %w", filepath.Base(path), err)
	}
	var gs state.GameState
	if err := json.Unmarshal(data, &gs); err != nil {
		return nil, fmt.Errorf("kayıt dosyası okunamadı: %w", err)
	}
	army.InitializeLegacyFleetDocking(gs.Armies, gs.Regions)
	gs.SyncTimedRegionUnlocks()
	ensureScenarioIdentity(&gs)
	applyScenarioMetadata(&gs)
	if gs.ScenarioPath == "" {
		return nil, fmt.Errorf("senaryo yolu çözümlenemedi")
	}

	dp := func(f string) string { return gs.ScenarioPath + "/data/" + f }
	if _, order, err := world.LoadRegionsWithOrder(dp("regions.json")); err == nil {
		gs.RegionOrder = order
	}

	unitTypes, err := army.LoadUnitTypes(dp("units.json"))
	if err != nil {
		return nil, err
	}
	gs.UnitTypes = unitTypes

	buildingTypes, err := city.LoadBuildings(dp("buildings.json"))
	if err != nil {
		return nil, err
	}
	gs.BuildingTypes = buildingTypes

	techTypes, err := tech.LoadTechnologies(dp("technologies.json"))
	if err != nil {
		return nil, err
	}
	gs.TechTypes = techTypes

	tradeCenters, err := world.LoadTradeCenters(dp("trade_centers.json"), gs.Regions)
	if err != nil {
		return nil, err
	}
	gs.TradeCenters = tradeCenters

	shapeData, err := world.LoadCountryShapes(dp("country_shapes.json"), gs.Regions)
	if err != nil {
		return nil, err
	}
	gs.ShapeData = shapeData
	return &gs, nil
}

func ensureScenarioIdentity(gs *state.GameState) {
	if gs == nil {
		return
	}
	if gs.ScenarioID == "" && gs.ScenarioPath != "" {
		gs.ScenarioID = filepath.Base(gs.ScenarioPath)
	}
	if gs.ScenarioPath == "" && gs.ScenarioID != "" {
		candidate := filepath.Join(scenarioBaseDir, gs.ScenarioID)
		if _, err := os.Stat(filepath.Join(candidate, "scenario.json")); err == nil {
			gs.ScenarioPath = candidate
		}
	}
}

func applyScenarioMetadata(gs *state.GameState) {
	if gs.ScenarioPath == "" {
		return
	}
	data, err := os.ReadFile(filepath.Join(gs.ScenarioPath, "scenario.json"))
	if err != nil {
		return
	}
	var sc scenario.Scenario
	if err := json.Unmarshal(data, &sc); err != nil {
		return
	}
	if mapConfigEmpty(gs.MapConfig) {
		gs.MapConfig = sc.MapConfig
	}
	gs.AvailableVictories = sc.VictoryConditions
}

func mapConfigEmpty(cfg scenario.MapConfig) bool {
	return cfg.WorldWidth == nil &&
		cfg.WorldHeight == nil &&
		cfg.ShapeOffsetX == nil &&
		cfg.ShapeOffsetY == nil &&
		cfg.ShapeScaleX == nil &&
		cfg.ShapeScaleY == nil
}
