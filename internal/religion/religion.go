package religion

// Type fraksiyon dini.
type Type string

const (
	Catholic Type = "catholic"
	Orthodox Type = "orthodox"
	Sunni    Type = "sunni"
	Shia     Type = "shia"
)

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
	return -30
}
