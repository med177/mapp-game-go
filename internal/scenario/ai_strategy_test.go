package scenario

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAIConfigPreservesFactionOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ai_strategies.json")
	data := []byte(`{
  "factions": [
    {"faction_id": "second", "objectives": []},
    {"faction_id": "first", "objectives": []}
  ]
}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("AI strateji fixture'ı yazılamadı: %v", err)
	}

	config, err := LoadAIConfig(path)
	if err != nil {
		t.Fatalf("AI stratejileri yüklenemedi: %v", err)
	}
	want := []string{"second", "first"}
	if len(config.StrategyOrder) != len(want) {
		t.Fatalf("strateji sıra uzunluğu = %d, beklenen %d", len(config.StrategyOrder), len(want))
	}
	for i := range want {
		if config.StrategyOrder[i] != want[i] {
			t.Fatalf("strateji sırası[%d] = %q, beklenen %q", i, config.StrategyOrder[i], want[i])
		}
	}
}
