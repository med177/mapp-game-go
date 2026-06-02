package render

import (
	"image/color"

	gameui "mapp-game-go/internal/ui"
)

var menuButtonStyle = gameui.ButtonStyle{
	BG:             color.RGBA{32, 28, 18, 220},
	Border:         panelBorder,
	Text:           ColorGold,
	DisabledBG:     color.RGBA{22, 20, 16, 180},
	DisabledBorder: color.RGBA{45, 38, 25, 160},
	DisabledText:   color.RGBA{90, 82, 60, 180},
	TextOffsetY:    8,
	BorderWidth:    1,
}

var tinyButtonStyle = gameui.ButtonStyle{
	BG:             color.RGBA{34, 26, 15, 230},
	Border:         panelBorder,
	Text:           ColorGold,
	DisabledBG:     color.RGBA{18, 16, 12, 180},
	DisabledBorder: color.RGBA{45, 38, 25, 160},
	DisabledText:   color.RGBA{85, 78, 62, 190},
	TextOffsetY:    2,
	BorderWidth:    1,
}

var slotMiniButtonStyle = gameui.ButtonStyle{
	BG:             color.RGBA{45, 45, 45, 230},
	Border:         panelBorder,
	Text:           ColorWhite,
	DisabledBG:     color.RGBA{24, 24, 24, 180},
	DisabledBorder: color.RGBA{45, 38, 25, 160},
	DisabledText:   color.RGBA{120, 120, 120, 180},
	TextOffsetY:    5,
	BorderWidth:    1,
}

var dropdownStyle = gameui.DropdownStyle{
	PanelBG:       color.RGBA{16, 20, 24, 242},
	Border:        panelBorder,
	TitleColor:    ColorGold,
	RowBG:         color.RGBA{28, 24, 18, 220},
	SelectedRowBG: color.RGBA{86, 64, 24, 238},
	RowText:       ColorWhite,
	SelectedText:  ColorGold,
	MutedText:     ColorGray,
	TitleOffsetY:  8,
	RowOffsetY:    5,
	BorderWidth:   1,
}

var standardModalStyle = gameui.ModalStyle{
	Overlay: color.RGBA{0, 0, 0, 140},
	Panel: gameui.PanelStyle{
		BG:          color.RGBA{12, 10, 8, 245},
		Border:      color.RGBA{110, 90, 50, 255},
		BorderWidth: 2,
	},
}

var eventDetailModalStyle = gameui.ModalStyle{
	Overlay: color.RGBA{0, 0, 0, 120},
	Panel: gameui.PanelStyle{
		BG:          panelBg,
		Border:      panelBorder,
		BorderWidth: 1,
	},
}

var historicalEventModalStyle = gameui.ModalStyle{
	Overlay: color.RGBA{2, 1, 5, 210},
	Panel: gameui.PanelStyle{
		BG:          color.RGBA{15, 10, 25, 245},
		Border:      color.RGBA{180, 140, 50, 255},
		BorderWidth: 2.5,
	},
}

var shapeHelpPanelStyle = gameui.PanelStyle{
	BG:          color.RGBA{16, 20, 24, 218},
	Border:      color.RGBA{170, 140, 70, 255},
	BorderWidth: 1,
}

func solidButtonStyle(bg, border, text color.RGBA, textOffsetY float64) gameui.ButtonStyle {
	style := menuButtonStyle
	style.BG = bg
	style.Border = border
	style.Text = text
	style.TextOffsetY = textOffsetY
	return style
}
