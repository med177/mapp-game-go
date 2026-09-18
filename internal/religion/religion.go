package religion

// Type fraksiyon dini.
type Type string

const (
	Catholic   Type = "catholic"
	Orthodox   Type = "orthodox"
	Sunni      Type = "sunni"
	Shia       Type = "shia"
	Tengri     Type = "tengri"
	Pagan      Type = "pagan"
	Protestant Type = "protestant"
	Utraquist  Type = "utraquist"
)

type Def struct {
	Type   Type
	NameTR string
}

var defs = []Def{
	{Type: Catholic, NameTR: "Katolik"},
	{Type: Orthodox, NameTR: "Ortodoks"},
	{Type: Sunni, NameTR: "Sünni İslam"},
	{Type: Shia, NameTR: "Şii İslam"},
	{Type: Tengri, NameTR: "Tengricilik"},
	{Type: Pagan, NameTR: "Paganlık"},
	{Type: Protestant, NameTR: "Protestanlık"},
	{Type: Utraquist, NameTR: "Utrakvistlik"},
}

var defsByType = func() map[Type]Def {
	out := make(map[Type]Def, len(defs))
	for _, def := range defs {
		out[def.Type] = def
	}
	return out
}()

func All() []Type {
	out := make([]Type, 0, len(defs))
	for _, def := range defs {
		out = append(out, def.Type)
	}
	return out
}

func DisplayNameTR(t Type) string {
	if def, ok := defsByType[t]; ok {
		return def.NameTR
	}
	return string(t)
}

func Next(current Type) Type {
	options := All()
	for i, option := range options {
		if option == current {
			return options[(i+1)%len(options)]
		}
	}
	return Catholic
}

// Relation iki mezhep arasındaki diplomatik çarpanı döner (-50..+30).
func Relation(a, b Type) int {
	if a == b {
		return 25
	}
	if (a == Sunni && b == Shia) || (a == Shia && b == Sunni) {
		return -40
	}
	if (a == Catholic && b == Orthodox) || (a == Orthodox && b == Catholic) {
		return -20
	}
	if (a == Catholic && b == Protestant) || (a == Protestant && b == Catholic) {
		return -20
	}
	if (a == Protestant && b == Utraquist) || (a == Utraquist && b == Protestant) {
		return -20
	}
	return -30
}
