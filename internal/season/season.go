package season

// Season mevsimi tanımlar.
type Season string

const (
	SeasonSpring Season = "spring" // ilkbahar: 3,4,5
	SeasonSummer Season = "summer" // yaz: 6,7,8
	SeasonAutumn Season = "autumn" // sonbahar: 9,10,11
	SeasonWinter Season = "winter" // kış: 12,1,2
)

// FromMonth ay numarasından (1-12) mevsim döner.
func FromMonth(month int) Season {
	switch month {
	case 3, 4, 5:
		return SeasonSpring
	case 6, 7, 8:
		return SeasonSummer
	case 9, 10, 11:
		return SeasonAutumn
	default: // 12, 1, 2
		return SeasonWinter
	}
}

// MovementMod mevsime göre hareket puanı çarpanı döner (yüzde).
// 110 = %10 bonus, 80 = %20 ceza.
func (s Season) MovementMod() int {
	switch s {
	case SeasonSpring:
		return 110
	case SeasonSummer:
		return 100
	case SeasonAutumn:
		return 95
	case SeasonWinter:
		return 70
	}
	return 100
}

// HarvestMod sonbaharda vergi geliri bonusu.
func (s Season) HarvestMod() int {
	if s == SeasonAutumn {
		return 120 // %20 bonus
	}
	return 100
}

// TradeMod mevsime göre ticaret geliri çarpanı döner (yüzde).
// İlkbahar: %110 (+10), Yaz: %100, Sonbahar: %110 (+10), Kış: %70 (-30)
func (s Season) TradeMod() int {
	switch s {
	case SeasonSpring:
		return 110
	case SeasonSummer:
		return 100
	case SeasonAutumn:
		return 110
	case SeasonWinter:
		return 70
	}
	return 100
}

// PirateMod mevsime göre korsan baskını olasılık çarpanı (yüzde).
// Kışın deniz trafiği azalır → korsan az, yazın artar → korsan çok.
func (s Season) PirateMod() int {
	switch s {
	case SeasonSpring:
		return 80
	case SeasonSummer:
		return 120
	case SeasonAutumn:
		return 90
	case SeasonWinter:
		return 50
	}
	return 100
}

// IsWinter kış mevsimi mi?
func (s Season) IsWinter() bool {
	return s == SeasonWinter
}

// DisplayName Türkçe görünen isim.
func (s Season) DisplayName() string {
	switch s {
	case SeasonSpring:
		return "İlkbahar"
	case SeasonSummer:
		return "Yaz"
	case SeasonAutumn:
		return "Sonbahar"
	case SeasonWinter:
		return "Kış"
	}
	return ""
}
