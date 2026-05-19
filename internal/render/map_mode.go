package render

type MapMode int

const (
	MapModeNormal MapMode = iota
	MapModeTrade
)

func (m MapMode) LabelTR() string {
	switch m {
	case MapModeTrade:
		return "Ticaret"
	default:
		return "Normal"
	}
}

func (m MapMode) Next() MapMode {
	switch m {
	case MapModeTrade:
		return MapModeNormal
	default:
		return MapModeTrade
	}
}
