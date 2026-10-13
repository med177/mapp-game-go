package render

import (
	"path/filepath"
	"testing"
)

func TestCommanderPortraitPathUsesCharactersDirectory(t *testing.T) {
	previousScenarioPath := ActiveScenarioPath
	ActiveScenarioPath = filepath.Join("testdata", "scenario")
	t.Cleanup(func() { ActiveScenarioPath = previousScenarioPath })

	want := filepath.Join(ActiveScenarioPath, "sprites", "characters", "commander.png")
	if got := commanderPortraitPath("commander.png"); got != want {
		t.Fatalf("commanderPortraitPath() = %q, want %q", got, want)
	}
	if got := commanderPortraitPath(filepath.Join("commanders", "commander.png")); got != want {
		t.Fatalf("legacy commander path = %q, want %q", got, want)
	}
}
