package render

import (
	"image/color"
	"testing"

	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/state"
)

func TestArmySheetGroupFollowsFactionReligion(t *testing.T) {
	gs := &state.GameState{
		Factions: map[faction.FactionID]*faction.Faction{
			"sunni":    {ID: "sunni", Religion: religion.Sunni},
			"shia":     {ID: "shia", Religion: religion.Shia},
			"catholic": {ID: "catholic", Religion: religion.Catholic},
			"orthodox": {ID: "orthodox", Religion: religion.Orthodox},
			"unknown":  {ID: "unknown"},
		},
	}

	checks := []struct {
		owner string
		want  armySheetGroup
	}{
		{owner: "sunni", want: armySheetGroupMuslim},
		{owner: "shia", want: armySheetGroupMuslim},
		{owner: "catholic", want: armySheetGroupChristian},
		{owner: "orthodox", want: armySheetGroupChristian},
		{owner: "unknown", want: armySheetGroupLegacy},
		{owner: "missing", want: armySheetGroupLegacy},
	}
	for _, check := range checks {
		if got := armySheetGroupForFaction(gs, check.owner); got != check.want {
			t.Fatalf("%s için army sheet grubu yanlış: got=%v want=%v", check.owner, got, check.want)
		}
	}
}

func TestArmyIconBorderColorUsesDiplomaticPalette(t *testing.T) {
	gs := &state.GameState{
		PlayerFactionID: "player",
		Factions: map[faction.FactionID]*faction.Faction{
			"player": {ID: "player"},
			"ally":   {ID: "ally"},
			"enemy":  {ID: "enemy"},
			"vassal": {ID: "vassal", OverlordID: "player"},
		},
		Relations: map[string]*faction.Relation{
			faction.RelationKey("player", "ally"):  {FactionA: "player", FactionB: "ally", Stance: faction.StanceAllied},
			faction.RelationKey("player", "enemy"): {FactionA: "player", FactionB: "enemy", Stance: faction.StanceWar},
		},
	}

	if got := armyIconBorderColor(gs, "player", false); got != borderColorPlayerRealm {
		t.Fatalf("player border rengi yanlış: got=%v want=%v", got, borderColorPlayerRealm)
	}
	if got := armyIconBorderColor(gs, "vassal", false); got != borderColorPlayerRealm {
		t.Fatalf("vassal border rengi yanlış: got=%v want=%v", got, borderColorPlayerRealm)
	}
	if got := armyIconBorderColor(gs, "ally", false); got != borderColorAlly {
		t.Fatalf("ally border rengi yanlış: got=%v want=%v", got, borderColorAlly)
	}
	if got := armyIconBorderColor(gs, "enemy", false); got != borderColorEnemy {
		t.Fatalf("enemy border rengi yanlış: got=%v want=%v", got, borderColorEnemy)
	}
	if got := armyIconBorderColor(gs, "enemy", true); got != (color.RGBA{255, 215, 0, 255}) {
		t.Fatalf("selected army border rengi sabit altın olmalı: got=%v", got)
	}
}
