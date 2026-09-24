package render

import (
	"testing"
	"time"

	"mapp-game-go/internal/state"
)

func TestAnimatedArmyScreenPosStopsMovementSoundWhenMarkerReturns(t *testing.T) {
	r := &Renderer{
		gs:       &state.GameState{ScenarioPath: t.TempDir()},
		camScale: 1,
		armyMovementAnimation: armyMovementAnimation{
			armyID:       "army",
			soundName:    "army_move",
			fromX:        10,
			fromY:        20,
			toX:          30,
			toY:          40,
			startedAt:    time.Now().Add(-time.Second),
			duration:     500 * time.Millisecond,
			active:       true,
			finalStep:    true,
			visualActive: true,
		},
	}

	if _, _, ok := r.animatedArmyScreenPos("army"); ok {
		t.Fatal("tamamlanan animasyon marker konumu döndürdü")
	}
	if r.armyMovementAnimation.active {
		t.Fatal("marker normal konuma dönerken animasyon aktif kaldı")
	}
	if r.armyMovementAnimation.soundName != "" {
		t.Fatalf("marker geri görünürken hareket sesi açık kaldı: %q", r.armyMovementAnimation.soundName)
	}
}

func TestIntermediateArmyMovementKeepsMovementSound(t *testing.T) {
	r := &Renderer{
		gs: &state.GameState{ScenarioPath: t.TempDir()},
		armyMovementAnimation: armyMovementAnimation{
			soundName:    "army_move",
			active:       true,
			finalStep:    false,
			visualActive: true,
		},
	}

	r.completeArmyMovementAnimation()
	if r.armyMovementAnimation.soundName != "army_move" {
		t.Fatalf("ara geçişte hareket sesi kesildi: %q", r.armyMovementAnimation.soundName)
	}
	if !r.armyMovementAnimation.visualActive {
		t.Fatal("ara geçişte hareket markerı gizli kalmadı")
	}
}
