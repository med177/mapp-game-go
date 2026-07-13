package render

import (
	"testing"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"
)

func TestShouldPromptWarConfirmForMoveSkipsAlliesAndPromptsHostiles(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player", NameTR: "Oyuncu"},
			"ally":   {ID: "ally", NameTR: "Müttefik"},
			"enemy":  {ID: "enemy", NameTR: "Düşman"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StancePeace},
		},
	}

	attacker := &army.Army{OwnerID: "player"}

	if shouldPromptWarConfirmForMove(gs, attacker, &world.Region{OwnerID: "ally"}) {
		t.Fatal("müttefik bölge için savaş ilanı penceresi açılmamalıydı")
	}

	if !shouldPromptWarConfirmForMove(gs, attacker, &world.Region{OwnerID: "enemy"}) {
		t.Fatal("düşman ama savaşta olmayan bölge için savaş ilanı penceresi açılmalıydı")
	}

	gs.Relations[faction.RelationKey("player", "enemy")].Stance = faction.StanceWar
	if shouldPromptWarConfirmForMove(gs, attacker, &world.Region{OwnerID: "enemy"}) {
		t.Fatal("savaş halindeki bölge için ek savaş ilanı penceresi açılmamalıydı")
	}
}
