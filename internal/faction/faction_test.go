package faction

import (
	"encoding/json"
	"testing"

	"mapp-game-go/internal/religion"
)

func TestFactionDiplomacyDisabledLoadsFromJSON(t *testing.T) {
	var disabled Faction
	if err := json.Unmarshal([]byte(`{"id":"others","diplomacy_disabled":true}`), &disabled); err != nil {
		t.Fatalf("diplomasi ayarı yüklenemedi: %v", err)
	}
	if !disabled.DiplomacyDisabled {
		t.Fatal("diplomacy_disabled=true faction modeline yüklenmedi")
	}

	var legacy Faction
	if err := json.Unmarshal([]byte(`{"id":"legacy"}`), &legacy); err != nil {
		t.Fatalf("eski faction JSON'u yüklenemedi: %v", err)
	}
	if legacy.DiplomacyDisabled {
		t.Fatal("alanı olmayan eski faction diplomasiye kapalı olmamalı")
	}
}

func TestDiplomacyDisabledFactionCannotTrade(t *testing.T) {
	f := &Faction{DiplomacyDisabled: true}
	if f.CanTrade() {
		t.Fatal("diplomasiye kapalı faction ticaret yapabilmeli")
	}
}

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
