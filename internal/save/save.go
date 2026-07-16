package save

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/buildinfo"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/diplomacy"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

const saveDir = "saves"
const autoSavePath = "saves/autosave.json"

var scenarioBaseDir = filepath.Join("assets", "scenarios")

// GameVersion save dosyasına yazılan oyun sürümü bilgisidir.
// Varsayılan değer buildinfo üzerinden gelir; testler veya özel akışlar bunu override edebilir.
var GameVersion = buildinfo.SaveVersion()

// SaveKind save dosyasının türünü belirtir.
type SaveKind string

const (
	SaveKindAuto  SaveKind = "auto"
	SaveKindQuick SaveKind = "quick"
	SaveKindSlot  SaveKind = "slot"
)

// slotDefs tüm kayıt slotlarını tanımlar; sıra UI'da gösterim sırasıdır.
var slotDefs = []struct {
	name        string
	displayName string
	path        string
}{
	{"autosave", "Otomatik Kayıt", "saves/autosave.json"},
	{"quicksave", "Hızlı Kayıt", "saves/quicksave.json"},
	{"slot1", "Kayıt 1", "saves/slot1.json"},
	{"slot2", "Kayıt 2", "saves/slot2.json"},
	{"slot3", "Kayıt 3", "saves/slot3.json"},
}

func slotPath(slotName string) (string, bool) {
	for _, def := range slotDefs {
		if def.name == slotName {
			return def.path, true
		}
	}
	return "", false
}

func kindForSlotName(slotName string) SaveKind {
	switch slotName {
	case "autosave":
		return SaveKindAuto
	case "quicksave":
		return SaveKindQuick
	case "slot1", "slot2", "slot3":
		return SaveKindSlot
	default:
		return SaveKindAuto
	}
}

type saveEnvelope struct {
	Kind          SaveKind     `json:"kind"`
	GameVersion   string       `json:"game_version"`
	Meta          saveMetadata `json:"meta"`
	StateEncoding string       `json:"state_encoding,omitempty"`
	StateZstd     string       `json:"state_zstd,omitempty"`
}

type saveEnvelopeRaw struct {
	Kind          SaveKind        `json:"kind,omitempty"`
	GameVersion   string          `json:"game_version,omitempty"`
	Meta          saveMetadata    `json:"meta,omitempty"`
	StateEncoding string          `json:"state_encoding,omitempty"`
	StateZstd     string          `json:"state_zstd,omitempty"`
	State         json.RawMessage `json:"state,omitempty"`
}

type debugSaveEnvelope struct {
	Kind        SaveKind                `json:"kind"`
	GameVersion string                  `json:"game_version"`
	Meta        saveMetadata            `json:"meta"`
	State       legacyCampaignSaveState `json:"state"`
}

type saveMetadata struct {
	ScenarioID      string            `json:"scenario_id,omitempty"`
	ScenarioPath    string            `json:"scenario_path,omitempty"`
	PlayerFactionID faction.FactionID `json:"player_faction_id,omitempty"`
	FactionName     string            `json:"faction_name,omitempty"`
	Turn            int               `json:"turn,omitempty"`
	Year            int               `json:"year,omitempty"`
}

// SaveSlot bir kayıt slotunun metadata'sını taşır.
type SaveSlot struct {
	Name        string
	DisplayName string
	Path        string
	Kind        SaveKind
	GameVersion string
	Exists      bool
	FactionName string
	Turn        int
	Year        int
	ModTime     time.Time
}

// metaFields sadece metadata okumak için minimal struct.
type metaFields struct {
	ScenarioID      string `json:"scenario_id"`
	ScenarioPath    string `json:"scenario_path,omitempty"`
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
			Kind:        kindForSlotName(def.name),
		}
		fi, err := os.Stat(def.path)
		if err == nil {
			s.Exists = true
			s.ModTime = fi.ModTime()
			if data, err := os.ReadFile(def.path); err == nil {
				if meta, version, ok, err := readSaveMetadata(data); err == nil && ok {
					s.GameVersion = version
					s.Turn = meta.Turn
					s.Year = meta.Year
					s.FactionName = meta.FactionName
					if s.FactionName == "" {
						s.FactionName = factionNameFromScenario(meta.ScenarioID, meta.ScenarioPath, meta.PlayerFactionID)
					}
				} else {
					payload, version, _, err := splitSavePayload(data)
					if err == nil {
						s.GameVersion = version
						var m metaFields
						if json.Unmarshal(payload, &m) == nil {
							s.Turn = m.Turn
							s.Year = m.Year
							if f, ok := m.Factions[m.PlayerFactionID]; ok {
								s.FactionName = f.NameTR
							}
							if s.FactionName == "" {
								s.FactionName = factionNameFromScenario(m.ScenarioID, m.ScenarioPath, faction.FactionID(m.PlayerFactionID))
							}
						}
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
	path, ok := slotPath(slotName)
	if !ok {
		path = autoSavePath
	}
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return fmt.Errorf("kayıt dizini oluşturulamadı: %w", err)
	}
	savedState, err := makeCampaignSaveState(gs)
	if err != nil {
		return fmt.Errorf("kayıt hazırlanamadı: %w", err)
	}
	stateEncoding, stateZstd, err := encodeCompressedStatePayload(savedState)
	if err != nil {
		return fmt.Errorf("kayıt sıkıştırılamadı: %w", err)
	}
	payload := saveEnvelope{
		Kind:          kindForSlotName(slotName),
		GameVersion:   GameVersion,
		Meta:          makeSaveMetadata(gs),
		StateEncoding: stateEncoding,
		StateZstd:     stateZstd,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("serileştirme hatası: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("dosya yazılamadı: %w", err)
	}
	if gs != nil && gs.DevelopmentMode {
		if err := writeDebugSidecar(path, debugSaveEnvelope{
			Kind:        payload.Kind,
			GameVersion: payload.GameVersion,
			Meta:        payload.Meta,
			State:       makeDebugCampaignSaveState(gs),
		}); err != nil {
			log.Printf("Save debug sidecar yazılamadı (%s): %v", path, err)
		}
	} else if err := removeDebugSidecar(path); err != nil {
		log.Printf("Eski save debug sidecar temizlenemedi (%s): %v", path, err)
	}
	return nil
}

func writeDebugSidecar(path string, payload debugSaveEnvelope) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(debugSidecarPath(path), data, 0644)
}

func removeDebugSidecar(path string) error {
	if err := os.Remove(debugSidecarPath(path)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func debugSidecarPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + ".debug.json"
	}
	return strings.TrimSuffix(path, ext) + ".debug" + ext
}

// LoadSlot isimli slottan oyun durumunu yükler.
func LoadSlot(slotName string) (*state.GameState, error) {
	path, ok := slotPath(slotName)
	if !ok {
		path = autoSavePath
	}
	return loadFromPath(path)
}

// LatestContinueSlot en yeni autosave/quicksave slotunu döner.
func LatestContinueSlot() (string, bool) {
	var newestName string
	var newestModTime time.Time
	found := false

	for _, slotName := range []string{"autosave", "quicksave"} {
		path, ok := slotPath(slotName)
		if !ok {
			continue
		}
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !found || fi.ModTime().After(newestModTime) {
			newestName = slotName
			newestModTime = fi.ModTime()
			found = true
		}
	}

	return newestName, found
}

// ContinueSaveExists autosave ya da quicksave içinde en az bir kayıt olup olmadığını kontrol eder.
func ContinueSaveExists() bool {
	_, ok := LatestContinueSlot()
	return ok
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
	path, ok := slotPath(slotName)
	if ok {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("kayıt silinemedi: %w", err)
		}
		return nil
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
	payload, _, _, err := splitSavePayload(data)
	if err != nil {
		return nil, fmt.Errorf("kayıt dosyası okunamadı: %w", err)
	}
	saved, err := decodeCampaignSaveState(payload)
	if err != nil {
		return nil, fmt.Errorf("kayıt dosyası okunamadı: %w", err)
	}
	if saved.ScenarioID == "" && saved.ScenarioPath == "" {
		return nil, fmt.Errorf("senaryo yolu çözümlenemedi")
	}
	gs, err := loadScenarioBaseState(saved.ScenarioID, saved.ScenarioPath)
	if err != nil {
		return nil, err
	}
	applyCampaignSaveState(gs, saved)
	army.NormalizeLegacyGarrisons(gs.Armies)
	army.InitializeLegacyFleetDocking(gs.Armies, gs.Regions)
	gs.RefreshArmyMovePoints(false)
	gs.SyncTimedRegionUnlocks()
	gs.NormalizeFactionCapitals()
	gs.AvailableVictories = scenario.FilterVictoryOptionsForFaction(gs.ScenarioVictories, string(gs.PlayerFactionID))
	diplomacy.NormalizeVassalage(gs)
	if gs.TradeRoutes == nil {
		gs.TradeRoutes = []*economy.TradeRoute{}
	}
	diplomacy.SanitizeTradeRoutes(gs)
	if len(gs.TradeRoutes) == 0 {
		diplomacy.EnsureTradeRoutesForActiveRelations(gs)
	}
	gs.MarketPrices = economy.ComputeMarketPrices(gs.Factions)
	return gs, nil
}

func splitSavePayload(data []byte) ([]byte, string, bool, error) {
	var env saveEnvelopeRaw
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, "", false, err
	}
	if env.StateZstd != "" {
		payload, err := decodeCompressedStatePayload(env.StateEncoding, env.StateZstd)
		if err != nil {
			return nil, "", false, err
		}
		return payload, env.GameVersion, true, nil
	}
	if len(env.State) > 0 {
		return env.State, string(env.GameVersion), true, nil
	}
	return data, "", false, nil
}

func readSaveMetadata(data []byte) (saveMetadata, string, bool, error) {
	var env saveEnvelopeRaw
	if err := json.Unmarshal(data, &env); err != nil {
		return saveMetadata{}, "", false, err
	}
	if env.GameVersion == "" && env.Kind == "" && env.StateZstd == "" && len(env.State) == 0 && env.Meta == (saveMetadata{}) {
		return saveMetadata{}, "", false, nil
	}
	return env.Meta, env.GameVersion, true, nil
}

func makeSaveMetadata(gs *state.GameState) saveMetadata {
	meta := saveMetadata{
		ScenarioID:      gs.ScenarioID,
		ScenarioPath:    saveScenarioPath(gs.ScenarioID, gs.ScenarioPath),
		PlayerFactionID: gs.PlayerFactionID,
		Turn:            gs.Turn,
		Year:            gs.Year,
	}
	if fx := gs.Factions[gs.PlayerFactionID]; fx != nil {
		meta.FactionName = fx.NameTR
	}
	if meta.FactionName == "" {
		meta.FactionName = factionNameFromScenario(gs.ScenarioID, gs.ScenarioPath, gs.PlayerFactionID)
	}
	return meta
}

func loadScenarioBaseState(scenarioID, savedScenarioPath string) (*state.GameState, error) {
	scenarioPath := resolveScenarioPath(scenarioID, savedScenarioPath)
	if scenarioPath == "" {
		return nil, fmt.Errorf("senaryo yolu çözümlenemedi")
	}

	sc, err := loadScenarioDefinition(scenarioPath)
	if err != nil {
		return nil, err
	}

	dp := func(f string) string { return filepath.Join(scenarioPath, "data", f) }

	regions, regionOrder, err := world.LoadRegionsWithOrder(dp("regions.json"))
	if err != nil {
		return nil, err
	}
	if err := world.LoadRegionSettlements(dp("settlements.json"), regions); err != nil {
		return nil, err
	}

	shapeData, err := world.LoadCountryShapes(dp("country_shapes.json"), regions)
	if err != nil {
		log.Printf("Ülke sınırları yüklenemedi: %v", err)
	}

	factions, factionOrder, err := faction.LoadFactionsWithOrder(dp("factions.json"))
	if err != nil {
		return nil, err
	}
	relations, err := faction.LoadRelations(dp("relations.json"), factions)
	if err != nil {
		return nil, err
	}

	unitTypes, err := army.LoadUnitTypes(dp("units.json"))
	if err != nil {
		log.Printf("Birim tipleri yüklenemedi: %v", err)
	}
	commanderTemplates, err := army.LoadCommanderTemplates(dp("commanders.json"))
	if err != nil {
		log.Printf("Komutan şablonları yüklenemedi: %v", err)
		commanderTemplates = map[string][]*army.Commander{}
	}
	buildingTypes, err := city.LoadBuildings(dp("buildings.json"))
	if err != nil {
		log.Printf("Binalar yüklenemedi: %v", err)
	}
	techTypes, err := tech.LoadTechnologies(dp("technologies.json"))
	if err != nil {
		log.Printf("Teknolojiler yüklenemedi: %v", err)
	}

	armies, err := army.LoadArmies(dp("armies.json"), unitTypes)
	if err != nil {
		log.Printf("Ordular yüklenemedi: %v", err)
		armies = map[army.ArmyID]*army.Army{}
	}
	army.NormalizeLegacyGarrisons(armies)

	tradeCenters, err := world.LoadTradeCenters(dp("trade_centers.json"), regions)
	if err != nil {
		log.Printf("Ticaret merkezleri yüklenemedi: %v", err)
	}

	gs := &state.GameState{
		Turn:               1,
		Year:               sc.Year,
		Month:              sc.Month,
		StartYear:          sc.Year,
		Phase:              state.PhasePlayerTurn,
		ScenarioID:         scenarioIDFromPath(scenarioPath),
		ScenarioPath:       scenarioPath,
		MapConfig:          sc.MapConfig,
		Regions:            regions,
		RegionOrder:        regionOrder,
		Factions:           factions,
		FactionOrder:       factionOrder,
		Armies:             armies,
		ShapeData:          shapeData,
		UnitTypes:          unitTypes,
		CommanderTemplates: commanderTemplates,
		BuildingTypes:      buildingTypes,
		TechTypes:          techTypes,
		ScenarioVictories:  sc.VictoryConditions,
		AvailableVictories: scenario.FilterVictoryOptionsForFaction(sc.VictoryConditions, ""),
		Relations:          relations,
		TradeCenters:       tradeCenters,
		NextArmySeq:        len(armies),
		FiredEventIDs:      map[string]bool{},
	}
	return gs, nil
}

func resolveScenarioPath(scenarioID, savedScenarioPath string) string {
	if savedScenarioPath != "" {
		if _, err := os.Stat(filepath.Join(savedScenarioPath, "scenario.json")); err == nil {
			return savedScenarioPath
		}
	}
	if scenarioID == "" {
		return ""
	}
	candidate := filepath.Join(scenarioBaseDir, scenarioID)
	if _, err := os.Stat(filepath.Join(candidate, "scenario.json")); err == nil {
		return candidate
	}
	return ""
}

func loadScenarioDefinition(scenarioPath string) (*scenario.Scenario, error) {
	data, err := os.ReadFile(filepath.Join(scenarioPath, "scenario.json"))
	if err != nil {
		return nil, err
	}
	var sc scenario.Scenario
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, err
	}
	sc.Path = scenarioPath
	return &sc, nil
}

func scenarioIDFromPath(scenarioPath string) string {
	if scenarioPath == "" {
		return ""
	}
	return filepath.Base(scenarioPath)
}

func factionNameFromScenario(scenarioID, savedScenarioPath string, factionID faction.FactionID) string {
	if factionID == "" {
		return ""
	}
	scenarioPath := resolveScenarioPath(scenarioID, savedScenarioPath)
	if scenarioPath == "" {
		return ""
	}
	factions, err := faction.LoadFactions(filepath.Join(scenarioPath, "data", "factions.json"))
	if err != nil {
		return ""
	}
	if fx := factions[factionID]; fx != nil {
		return fx.NameTR
	}
	return ""
}
