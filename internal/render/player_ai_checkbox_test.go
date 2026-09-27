package render

import "testing"

func TestPlayerAITurnCheckboxUsesSharedEndTurnGeometry(t *testing.T) {
	end := BottomButtonRects()[4]
	control := playerAITurnControlRect()
	rects := playerAITurnCheckboxRects()
	if control.W < 150 || control.H <= 0 {
		t.Fatalf("iki satırlı oyuncu AI kontrol alanı yetersiz: %+v", control)
	}
	if control.X <= float64(BottomButtonRects()[3][0]+BottomButtonRects()[3][2]) {
		t.Fatalf("AI kontrol alanı Diplomasi düğmesine yaslanıyor: control=%+v diplomacy=%v", control, BottomButtonRects()[3])
	}
	if control.X+control.W >= float64(end[0]) {
		t.Fatalf("AI kontrol alanı Tur Bitir düğmesine yaslanıyor: control=%+v end=%v", control, end)
	}

	for index, rect := range rects {
		if rect.W <= 0 || rect.H <= 0 {
			t.Fatalf("oyuncu AI checkbox recti geçersiz: index=%d rect=%+v", index, rect)
		}
		if rect.X < control.X || rect.X+rect.W > control.X+control.W {
			t.Fatalf("checkbox kontrol kartının dışına taşıyor: index=%d rect=%+v control=%+v", index, rect, control)
		}
		if !playerAITurnCheckboxHit(rect.X+rect.W/2, rect.Y+rect.H/2) {
			t.Fatalf("checkbox merkezinden tıklanamadı: index=%d", index)
		}
		if playerAITurnCheckboxIndex(rect.X+rect.W/2, rect.Y+rect.H/2) != index {
			t.Fatalf("checkbox satırı yanlış çözüldü: got=%d want=%d", playerAITurnCheckboxIndex(rect.X+rect.W/2, rect.Y+rect.H/2), index)
		}
		if playerAITurnCheckboxHit(rect.X-1, rect.Y+rect.H/2) {
			t.Fatalf("checkbox dışı tıklama kabul edildi: index=%d", index)
		}
	}
}
