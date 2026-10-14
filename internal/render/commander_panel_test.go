package render

import (
	"path/filepath"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/scenario"
)

func TestCharacterTraitDisplayNameUsesScenarioDefinition(t *testing.T) {
	config := scenario.CharacterCouncilConfig{
		Traits: map[string]scenario.CharacterTraitDefinition{
			"administrator": {NameTR: "İdareci"},
		},
	}
	if got := characterTraitDisplayName(army.CharacterTrait("administrator"), config); got != "İdareci" {
		t.Fatalf("scenario trait label = %q, want %q", got, "İdareci")
	}
	if got := characterTraitDisplayName(army.CharacterTraitVeteran, config); got != "Gazi" {
		t.Fatalf("built-in trait fallback = %q, want %q", got, "Gazi")
	}
}

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
