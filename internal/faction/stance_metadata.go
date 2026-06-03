package faction

type DiplomaticStanceDef struct {
	Stance  DiplomaticStance
	NameTR  string
	BadgeTR string
}

var diplomaticStanceDefs = []DiplomaticStanceDef{
	{Stance: StancePeace, NameTR: "Barış", BadgeTR: "PEACE Barış"},
	{Stance: StanceWar, NameTR: "Savaş", BadgeTR: "WAR Savaş"},
	{Stance: StanceAllied, NameTR: "Müttefik", BadgeTR: "ALLY Müttefik"},
	{Stance: StanceTrade, NameTR: "Ticaret", BadgeTR: "TRADE Ticaret"},
}

var diplomaticStanceDefsByValue = func() map[DiplomaticStance]DiplomaticStanceDef {
	out := make(map[DiplomaticStance]DiplomaticStanceDef, len(diplomaticStanceDefs))
	for _, def := range diplomaticStanceDefs {
		out[def.Stance] = def
	}
	return out
}()

func AllDiplomaticStances() []DiplomaticStance {
	out := make([]DiplomaticStance, 0, len(diplomaticStanceDefs))
	for _, def := range diplomaticStanceDefs {
		out = append(out, def.Stance)
	}
	return out
}

func DiplomaticStanceLabelTR(stance DiplomaticStance) string {
	if def, ok := diplomaticStanceDefsByValue[stance]; ok {
		return def.NameTR
	}
	return string(stance)
}

func DiplomaticStanceBadgeTR(stance DiplomaticStance) string {
	if def, ok := diplomaticStanceDefsByValue[stance]; ok {
		return def.BadgeTR
	}
	return string(stance)
}

func NormalizeStance(stance DiplomaticStance) DiplomaticStance {
	if _, ok := diplomaticStanceDefsByValue[stance]; ok {
		return stance
	}
	return StancePeace
}

func NextDiplomaticStance(current DiplomaticStance) DiplomaticStance {
	options := AllDiplomaticStances()
	for i, option := range options {
		if option == current {
			return options[(i+1)%len(options)]
		}
	}
	return StancePeace
}
