package faction

import (
	"testing"

	"mapp-game-go/internal/religion"
)

func TestApplyHistoricalChangeUpdatesReligionFromEffectiveYear(t *testing.T) {
	f := &Faction{
		Religion: religion.Sunni,
		HistoricalChanges: []HistoricalChange{
			{Year: 1501, Religion: religion.Shia},
		},
	}

	if changed := f.ApplyHistoricalChange(1500); changed {
		t.Fatal("1501 değişikliği 1500 yılında uygulanmamalı")
	}
	if f.Religion != religion.Sunni {
		t.Fatalf("1500 dininin Sünni kalması bekleniyordu, %q bulundu", f.Religion)
	}

	if changed := f.ApplyHistoricalChange(1501); !changed {
		t.Fatal("1501 değişikliği uygulanmalı")
	}
	if f.Religion != religion.Shia {
		t.Fatalf("1501 dininin Şii olması bekleniyordu, %q bulundu", f.Religion)
	}

	if changed := f.ApplyHistoricalChange(1502); changed {
		t.Fatal("aynı tarihsel değişiklik ikinci kez değişiklik üretmemeli")
	}
}
