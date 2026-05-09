package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VictoryOptionDef bir senaryo için tek bir kazanma koşulunu tanımlar.
// UI metni (Title, Description, Detail) ve oyun mekaniği değerlerini (Type, hedefler) bir arada tutar.
type VictoryOptionDef struct {
	// Kimlik ve gösterim
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Detail      string `json:"detail"`

	// Kazanma türü: domination | economic | military | religious | conquer_city
	Type string `json:"type"`

	// Domination
	TargetRegionCount int      `json:"target_region_count"`
	RequiredRegions   []string `json:"required_regions"`

	// Economic
	TargetGoldIncome int `json:"target_gold_income"`
	GoldHoldTurns    int `json:"gold_hold_turns"`

	// Military
	TargetArmyStrength int `json:"target_army_strength"`
	TargetDefeated     int `json:"target_defeated"`

	// Conquer_city — tek hedef bölge
	Target string `json:"target"`

	// Ortak
	DeadlineTurn int `json:"deadline_turn"`
}

// MapConfig senaryonun arka plan ve shape hizalama ayarlarını tanımlar.
type MapConfig struct {
	WorldWidth   *int     `json:"world_width,omitempty"`
	WorldHeight  *int     `json:"world_height,omitempty"`
	ShapeOffsetX *float64 `json:"shape_offset_x,omitempty"`
	ShapeOffsetY *float64 `json:"shape_offset_y,omitempty"`
	ShapeScaleX  *float64 `json:"shape_scale_x,omitempty"`
	ShapeScaleY  *float64 `json:"shape_scale_y,omitempty"`
}

// MusicTrackDef senaryo musics/ klasöründeki bir playlist parçasını tanımlar.
type MusicTrackDef struct {
	File   string `json:"file"`
	Weight int    `json:"weight,omitempty"`
}

// MusicConfig senaryo bazlı müzik playlistlerini tanımlar.
type MusicConfig struct {
	DefaultPlaylist string                     `json:"default_playlist"`
	Playlists       map[string][]MusicTrackDef `json:"playlists"`
}

// Scenario oyun başında seçilebilen bir tarihsel senaryoyu tanımlar.
type Scenario struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     float64 `json:"version"`
	Author      string  `json:"author"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`

	MapConfig         MapConfig          `json:"map"`
	Music             MusicConfig        `json:"music"`
	VictoryConditions []VictoryOptionDef `json:"victory_conditions"`

	// Path: runtime-only, klasörün tam yolu (JSON'da yok)
	Path string `json:"-"`
}

// DataPath senaryo data/ klasöründeki bir dosyanın tam yolunu döner.
func (s *Scenario) DataPath(filename string) string {
	return filepath.Join(s.Path, "data", filename)
}

// AssetPath senaryo içindeki bir varlık alt klasörünün yolunu döner (maps, sprites, musics).
func (s *Scenario) AssetPath(subdir, filename string) string {
	return filepath.Join(s.Path, subdir, filename)
}

// LoadAll baseDir/scenarios.json index dosyasını okuyarak senaryoları sırayla yükler.
func LoadAll(baseDir string) ([]*Scenario, error) {
	indexPath := filepath.Join(baseDir, "scenarios.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("senaryo listesi okunamadı (%s): %w", indexPath, err)
	}

	var ids []string
	if err := json.Unmarshal(indexData, &ids); err != nil {
		return nil, fmt.Errorf("senaryo listesi parse hatası: %w", err)
	}

	var scenarios []*Scenario
	for _, id := range ids {
		metaPath := filepath.Join(baseDir, id, "scenario.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var s Scenario
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		s.Path = filepath.Join(baseDir, id)
		scenarios = append(scenarios, &s)
	}

	if len(scenarios) == 0 {
		return nil, fmt.Errorf("hiç senaryo yüklenemedi: %s", baseDir)
	}
	return scenarios, nil
}
