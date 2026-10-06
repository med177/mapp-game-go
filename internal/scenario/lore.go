package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FactionLore bir faction bilgi panelinde gösterilen senaryo metadatasıdır.
type FactionLore struct {
	TitleTR      string `json:"title_tr,omitempty"`
	HistoryTR    string `json:"history_tr,omitempty"`
	ImportanceTR string `json:"importance_tr,omitempty"`
	Image        string `json:"image,omitempty"`
}

// SettlementLore bir yerleşim bilgi panelinde gösterilen senaryo metadatasıdır.
type SettlementLore struct {
	TitleTR      string `json:"title_tr,omitempty"`
	HistoryTR    string `json:"history_tr,omitempty"`
	ImportanceTR string `json:"importance_tr,omitempty"`
	Image        string `json:"image,omitempty"`
}

func LoadLore(scenarioPath string) (map[string]FactionLore, map[string]SettlementLore, error) {
	factionLore := make(map[string]FactionLore)
	settlementLore := make(map[string]SettlementLore)
	if scenarioPath == "" {
		return factionLore, settlementLore, nil
	}
	if err := loadLoreFile(filepath.Join(scenarioPath, "wiki", "faction_history.json"), &factionLore); err != nil {
		return nil, nil, fmt.Errorf("faction lore okunamadı: %w", err)
	}
	if err := loadLoreFile(filepath.Join(scenarioPath, "wiki", "settlement_info.json"), &settlementLore); err != nil {
		return nil, nil, fmt.Errorf("settlement lore okunamadı: %w", err)
	}
	return factionLore, settlementLore, nil
}

func loadLoreFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}
