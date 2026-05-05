package render

import "image/color"

// BlendColors iki rengi t/255 oranında karıştırır.
func BlendColors(a, b color.RGBA, t uint8) color.RGBA {
	tF := float64(t) / 255
	aF := 1 - tF
	return color.RGBA{
		R: uint8(float64(a.R)*aF + float64(b.R)*tF),
		G: uint8(float64(a.G)*aF + float64(b.G)*tF),
		B: uint8(float64(a.B)*aF + float64(b.B)*tF),
		A: 255,
	}
}

// DimColor rengi belirli parlaklıkta koyulaştırır (128 = yarı parlak).
func DimColor(c color.RGBA, brightness uint8) color.RGBA {
	f := float64(brightness) / 255
	return color.RGBA{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: c.A,
	}
}
