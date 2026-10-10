package render

import "testing"

func TestSlotCardDetailsRowKeepsLongFactionSeparateFromTurn(t *testing.T) {
	card := slotCardLayout{X: 120, Y: 80, W: slotCardW, H: slotCardH}
	factionRect, turnRect := slotCardDetailsRowRects(card)

	if factionRect.X < card.X || factionRect.X+factionRect.W > turnRect.X {
		t.Fatalf("faction details overlap turn details: faction=%+v turn=%+v", factionRect, turnRect)
	}
	if turnRect.X+turnRect.W > card.X+card.W {
		t.Fatalf("turn details extend beyond the card: turn=%+v card=%+v", turnRect, card)
	}

	label := "Devlet: Venedik Cumhuriyeti"
	if got := MeasureText(label, FaceSmall); got > factionRect.W {
		t.Fatalf("faction label width = %.1f, available width = %.1f", got, factionRect.W)
	}
}

func TestSlotDeleteButtonDoesNotCoverDetails(t *testing.T) {
	card := slotCardLayout{X: 120, Y: 80, W: slotCardW, H: slotCardH}
	factionRect, turnRect := slotCardDetailsRowRects(card)
	deleteRect := slotDeleteButtonRect(card.X, card.Y)

	if deleteRect[1] < turnRect.Y+FaceSmall.Size {
		t.Fatalf("delete button starts before turn details end: delete=%v turn=%+v", deleteRect, turnRect)
	}
	if deleteRect[1]+deleteRect[3] > card.Y+card.H {
		t.Fatalf("delete button extends beyond the card: delete=%v card=%+v", deleteRect, card)
	}
	if factionRect.X+factionRect.W > turnRect.X {
		t.Fatalf("faction details overlap turn details: faction=%+v turn=%+v", factionRect, turnRect)
	}
}
