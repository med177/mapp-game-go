package ui

import "testing"

func TestDropdownDisabledOptionsAreTrackedAndReset(t *testing.T) {
	dropdown := NewDropdown(0, 0, 120, 80, "Seçenekler", 20, 20, 2)
	dropdown.SetOptions([]string{"Aktif", "Pasif"}, "Aktif")
	dropdown.SetDisabledOptions([]bool{false, true})

	if dropdown.OptionDisabled(0) {
		t.Fatal("aktif dropdown seçeneği devre dışı işaretlendi")
	}
	if !dropdown.OptionDisabled(1) {
		t.Fatal("pasif dropdown seçeneği devre dışı işaretlenmedi")
	}
	if dropdown.OptionDisabled(-1) || dropdown.OptionDisabled(2) {
		t.Fatal("geçersiz seçenek index'i tutarsız sonuç verdi")
	}
	if index, ok := dropdown.GetSelectedOption(12, 41); !ok || index != 1 {
		t.Fatalf("pasif satırın hit-test sonucu = %d, %t; 1, true bekleniyordu", index, ok)
	}

	dropdown.SetOptions([]string{"Yeni"}, "Yeni")
	if dropdown.OptionDisabled(0) {
		t.Fatal("yeni seçenekler önceki disabled durumunu devraldı")
	}
}
