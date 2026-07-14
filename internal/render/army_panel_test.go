package render

import (
	"testing"

	"mapp-game-go/internal/army"
)

func TestCommanderActionLabelShowsEmbarkedCommander(t *testing.T) {
	landCommander := army.NewCommander("land_cmd", "Kara Komutanı")
	fleetCommander := army.NewCommander("fleet_cmd", "Filo Komutanı")

	if got := commanderActionLabel(&army.Army{EmbarkedCommander: landCommander}); got != "KARA KOMUTANI" {
		t.Fatalf("yalnız taşınan komutan için yanlış aksiyon etiketi: %q", got)
	}
	if got := commanderActionLabel(&army.Army{Commander: fleetCommander, EmbarkedCommander: landCommander}); got != "KOMUTANLAR" {
		t.Fatalf("iki komutanlı filo için yanlış aksiyon etiketi: %q", got)
	}
}

func TestScoutedEnemyRevealCountUsesSeventyFivePercentForSiegeIntel(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, false, 0.75); got != 6 {
		t.Fatalf("8 birim için %%75 görünürlük 6 olmalıydı, got=%d", got)
	}
	if got := scoutedEnemyRevealCount(1, false, 0.75); got != 1 {
		t.Fatalf("tek birim için en az 1 görünürlük korunmalıydı, got=%d", got)
	}
}

func TestScoutedEnemyRevealCountUsesFullIntelWhenEnabled(t *testing.T) {
	if got := scoutedEnemyRevealCount(8, true, 0.75); got != 8 {
		t.Fatalf("tam istihbaratta tüm birimler görünmeliydi, got=%d", got)
	}
}
