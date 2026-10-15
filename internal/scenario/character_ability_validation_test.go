package scenario

import "testing"

func TestAdventureAbilityChoiceEffectValidation(t *testing.T) {
	config := CharacterCouncilConfig{Missions: []CharacterMissionDefinition{{
		ID: "voyage", NameTR: "Sefer", Type: CharacterMissionTypeAdventure,
		TargetKind: CharacterMissionTargetNone, DurationTurns: 1,
		Adventure: &CharacterAdventureDefinition{
			SuccessAbilityChoices: []AdventureAbilityChoice{{
				ID: "naval_strategy", NameTR: "Denizcilik", Scope: "navy", Effect: "navy_power", Modifier: 0.03,
			}},
		},
	}}}
	if err := config.Validate(); err != nil {
		t.Fatalf("valid ability effect rejected: %v", err)
	}

	config.Missions[0].Adventure.SuccessAbilityChoices[0].Effect = "army_power"
	if err := config.Validate(); err == nil {
		t.Fatal("ability effect incompatible with its scope was accepted")
	}
}
