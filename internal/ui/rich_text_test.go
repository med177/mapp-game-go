package ui

import "testing"

func TestRichTextBlockSkipsEmptySpacingLines(t *testing.T) {
	block := NewRichTextBlock(Rect{}, []RichTextLine{{}}, 20)

	// Boş satır yalnızca aralık içindir; renderer veya ekran olmadan da
	// güvenle atlanabilmelidir.
	block.Draw(nil, nil)
}
