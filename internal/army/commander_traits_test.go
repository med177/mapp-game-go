package army

import "testing"

func TestCharacterTraitLabelsAreLocalized(t *testing.T) {
	traits := []CharacterTrait{
		"assassin", "spy", "explorer", "queen", "intelligent", "beautiful",
		"manipulative", "sorceress", "shadowbinder", "mother", "resourceful",
		"loyal", "lover", "schemer", "vengeful", "fighter", "independent",
		"princess", "free_spirit", "mercenary", "scholar", "cunning", "goodwill",
		"advisor", "alchemist", "mysterious", "skilled", "cruel", "spoiled", "innocent",
	}
	for _, trait := range traits {
		if got := TraitLabelTR(trait); got == string(trait) {
			t.Errorf("TraitLabelTR(%q) returned its untranslated identifier", trait)
		}
	}
}
