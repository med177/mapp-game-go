package render

import "testing"

func TestInitialMainMenuCursorPrefersContinueWhenSaveExists(t *testing.T) {
	if got := InitialMainMenuCursor(true, false); got != 1 {
		t.Fatalf("save varken Devam et seçili başlamalı: got=%d want=1", got)
	}
	if got := InitialMainMenuCursor(false, false); got != 0 {
		t.Fatalf("save yokken Yeni Oyun seçili başlamalı: got=%d want=0", got)
	}
}

func TestEditModeMenuItemIsOptionalAndBelowExit(t *testing.T) {
	withoutEdit := buildMenuItems(false, false, false)
	if len(withoutEdit) != 5 || withoutEdit[0].action != ActionNewGame {
		t.Fatalf("EDIT_MODE=false iken menü sırası yanlış: %+v", withoutEdit)
	}

	withEdit := buildMenuItems(false, false, true)
	if len(withEdit) != 6 || withEdit[5].label != "EDIT MODE" || withEdit[5].action != ActionEditMode || withEdit[4].action != ActionQuit {
		t.Fatalf("EDIT_MODE=true iken EDIT MODE Çıkış'ın altında olmalı: %+v", withEdit)
	}
	if got := InitialMainMenuCursor(false, true); got != 0 {
		t.Fatalf("EDIT_MODE=true iken kayıt yokken varsayılan seçim Yeni Oyun olmalı: got=%d want=0", got)
	}
	if got := InitialMainMenuCursor(true, true); got != 1 {
		t.Fatalf("EDIT_MODE=true ve kayıt varken varsayılan seçim Devam et olmalı: got=%d want=1", got)
	}

	exitY := mainMenuItemY(100, withEdit, 4)
	editY := mainMenuItemY(100, withEdit, 5)
	if got, want := editY-exitY, mainMenuButtonHeight+mainMenuButtonGap; got != want {
		t.Fatalf("EDIT MODE ile Çıkış arasında standart boşluk olmalı: got=%.1f want=%.1f", got, want)
	}
}

func TestMainMenuItemsStayBelowHeader(t *testing.T) {
	oldHeight := ScreenHeight
	defer func() { ScreenHeight = oldHeight }()
	ScreenHeight = 720

	separatorY := ScreenHeight/2 - 200 + 80
	if got := mainMenuItemStartY(6) - 6; got <= separatorY {
		t.Fatalf("altı menü satırı ayraçla çakışmamalı: buttonTop=%.1f separatorY=%.1f", got, separatorY)
	}
}
