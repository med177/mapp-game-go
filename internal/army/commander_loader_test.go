package army

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCharacterTemplatesDefaultsLegacyTypeAndKeepsCharacterTypes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "characters.json")
	data := `[
		{"id":"administrator","type":"administrator","owner_id":"faction","name":"Administrator","assigned_council_role":"treasurer"},
		{"id":"commander","type":"commander","owner_id":"faction","name":"Commander"},
		{"id":"character","type":"character","owner_id":"faction","name":"Character"},
		{"id":"legacy","owner_id":"faction","name":"Legacy"}
	]`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	templates, err := LoadCharacterTemplates(path)
	if err != nil {
		t.Fatalf("LoadCharacterTemplates() error = %v", err)
	}
	got := templates["faction"]
	if len(got) != 4 {
		t.Fatalf("loaded template count = %d, want 4", len(got))
	}
	wantTypes := map[string]CharacterType{
		"administrator": CharacterTypeAdministrator,
		"commander":     CharacterTypeCommander,
		"character":     CharacterTypeCharacter,
		"legacy":        CharacterTypeCommander,
	}
	for _, character := range got {
		if want := wantTypes[character.ID]; character.Type != want {
			t.Errorf("character %q type = %q, want %q", character.ID, character.Type, want)
		}
		if character.AssignedCouncilRole != "" {
			t.Errorf("template character %q retained runtime council role %q", character.ID, character.AssignedCouncilRole)
		}
	}
}

func TestLoadCharacterTemplatesFallsBackToLegacyFilename(t *testing.T) {
	dir := t.TempDir()
	legacyPath := filepath.Join(dir, "commanders.json")
	if err := os.WriteFile(legacyPath, []byte(`[{"id":"legacy","owner_id":"faction","name":"Legacy"}]`), 0o600); err != nil {
		t.Fatal(err)
	}

	templates, err := LoadCharacterTemplates(filepath.Join(dir, "characters.json"))
	if err != nil {
		t.Fatalf("LoadCharacterTemplates() error = %v", err)
	}
	if got := templates["faction"]; len(got) != 1 || got[0].Type != CharacterTypeCommander {
		t.Fatalf("legacy templates = %#v, want one commander-typed record", got)
	}
}
