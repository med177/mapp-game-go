package render

import "testing"

func TestTerrainAreaInspectorButtonsUseOrderedRows(t *testing.T) {
	name := editTerrainAreaInspectorButtonRect(editButtonRegionNameTR)
	typeButton := editTerrainAreaInspectorButtonRect(editButtonTerrainAreaType)
	paint := editTerrainAreaInspectorButtonRect(editButtonTerrainArea)

	if name[1] >= typeButton[1] || typeButton[1] >= paint[1] {
		t.Fatalf("arazi düğmeleri yukarıdan aşağıya sıralı değil: name=%v type=%v paint=%v", name, typeButton, paint)
	}
	if editRectButton(name, "").HitTest(name[0]+name[2]/2, name[1]+name[3]/2) == false {
		t.Fatal("arazi adı rect'i kendi merkezini yakalamıyor")
	}

	r := &Renderer{}
	if got := r.editTerrainAreaInspectorButtonAt(name[0]+name[2]/2, name[1]+name[3]/2); got != editButtonRegionNameTR {
		t.Fatalf("arazi adı hit-test sonucu = %v, want %v", got, editButtonRegionNameTR)
	}
	oldTypeRect := editInspectorButtonRect(editButtonRegionTerrain)
	if got := r.editTerrainAreaInspectorButtonAt(oldTypeRect[0]+oldTypeRect[2]/2, oldTypeRect[1]+oldTypeRect[3]/2); got != editButtonNone {
		t.Fatalf("kaldırılan genel arazi tipi düğmesi hâlâ hit-test ediliyor: %v", got)
	}
}
