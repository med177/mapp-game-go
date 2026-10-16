package render

import (
	"image"
	"path/filepath"
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/scenario"
	gameui "mapp-game-go/internal/ui"
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

func TestCommanderArrivalPortraitIsSquareAndPreservesImageAspectRatio(t *testing.T) {
	card := gameui.Rect{X: 100, Y: 200, W: commanderArrivalCardW, H: commanderArrivalCardH}
	frame := commanderArrivalPortraitRect(card)
	if frame.W != frame.H {
		t.Fatalf("portrait frame is not square: %+v", frame)
	}
	if frame.X < card.X || frame.Y < card.Y || frame.X+frame.W > card.X+card.W || frame.Y+frame.H > card.Y+card.H {
		t.Fatalf("portrait frame extends beyond card: frame=%+v card=%+v", frame, card)
	}

	imageRect := commanderPortraitImageRect(image.Rect(0, 0, 128, 64), frame)
	if imageRect.W/imageRect.H != 2 {
		t.Fatalf("portrait aspect ratio = %.2f, want 2.00", imageRect.W/imageRect.H)
	}
	if imageRect.X != frame.X || imageRect.Y != frame.Y+(frame.H-imageRect.H)/2 {
		t.Fatalf("portrait image is not centered inside its square frame: image=%+v frame=%+v", imageRect, frame)
	}
}
