package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAIStrategiesValidatesAndIndexesFactionProfiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ai_strategies.json")
	data := []byte(`{"factions":[{"faction_id":"ottoman","profile":"frontier_expansion","naval_focus":true,"objectives":[{"id":"secure_bithynia","kind":"expand","target_factions":["east_rome"],"priority":100}]}]}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("AI strateji fixture yazılamadı: %v", err)
	}

	strategies, err := LoadAIStrategies(path)
	if err != nil {
		t.Fatalf("AI stratejileri yüklenemedi: %v", err)
	}
	got := strategies["ottoman"]
	if got.Profile != "frontier_expansion" || !got.NavalFocus || len(got.Objectives) != 1 || got.Objectives[0].ID != "secure_bithynia" {
		t.Fatalf("AI strateji profili yanlış yüklendi: %+v", got)
	}
}

func TestLoadAIStrategiesAllowsMissingOptionalFile(t *testing.T) {
	strategies, err := LoadAIStrategies(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || len(strategies) != 0 {
		t.Fatalf("opsiyonel dosya yokluğu genel AI fallback olmalı: strategies=%+v err=%v", strategies, err)
	}
}

func TestLoadAIConfigIncludesDifficultyPolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ai_strategies.json")
	data := []byte(`{"difficulty_policy":{"fair_movement":true,"levels":{"3":{"plan_horizon_turns":9,"plan_target_region_limit":5,"path_search_depth":12,"plan_move_bonus_percent":125,"proactive_war":true,"war_threshold":65,"min_attack_power_percent":100,"war_cadence_turns":7,"max_concurrent_wars":2,"start_gold_buffer":80,"start_grain_buffer":30}}},"factions":[]}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("AI config fixture yazılamadı: %v", err)
	}

	config, err := LoadAIConfig(path)
	if err != nil {
		t.Fatalf("AI config yüklenemedi: %v", err)
	}
	level, ok := config.DifficultyPolicy.Level(3)
	if !ok || !config.DifficultyPolicy.FairMovement || level.PlanHorizonTurns != 9 || level.PathSearchDepth != 12 || level.StartGoldBuffer != 80 || level.StartGrainBuffer != 30 {
		t.Fatalf("zorluk politikası yanlış yüklendi: policy=%+v level=%+v", config.DifficultyPolicy, level)
	}
}

func TestLoadAIConfigAllowsAllianceObjectiveMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ai_strategies.json")
	data := []byte(`{"factions":[{"faction_id":"york","objectives":[{"id":"channel_backing","kind":"ally","target_factions":["flanders_county"],"priority":48}]}]}`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("AI alliance fixture yazılamadı: %v", err)
	}

	config, err := LoadAIConfig(path)
	if err != nil {
		t.Fatalf("ally objective metadata'sı reddedilmemeli: %v", err)
	}
	if got := config.Strategies["york"].Objectives[0].Kind; got != "ally" {
		t.Fatalf("ally objective türü korunmadı: %q", got)
	}
}
