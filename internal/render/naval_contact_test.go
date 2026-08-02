package render

import "testing"

func TestNavalContactWithdrawButtonCanBeDisabled(t *testing.T) {
	r := &Renderer{}
	r.ShowThreeChoiceDialogWithThirdEnabled(
		"Düşman Filo Tespit Edildi",
		"Temas",
		"Çatış",
		"Geri Çekil",
		"Pozisyonu Koru",
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 0},
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 1},
		InputAction{Kind: ActionResolveNavalContact, ChoiceIndex: 2},
		false,
	)
	_, third, _, hasThird := buildConfirmDialogButtons(r.confirmDialog)
	if !hasThird || third.Enabled {
		t.Fatal("hareket puanı olmayan filo için geri çekil seçeneği pasif olmalı")
	}
}
