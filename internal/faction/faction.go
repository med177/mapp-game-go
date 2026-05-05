package faction

// Religion fraksiyon dini.
type Religion string

const (
	ReligionCatholic Religion = "catholic"
	ReligionOrthodox Religion = "orthodox"
	ReligionSunni    Religion = "sunni"
	ReligionShia     Religion = "shia"
)

// ReligionRelation iki mezhep arasındaki diplomatik çarpanı döner (-50..+30).
func ReligionRelation(a, b Religion) int {
	if a == b {
		return 25
	}
	if (a == ReligionSunni && b == ReligionShia) || (a == ReligionShia && b == ReligionSunni) {
		return -40
	}
	if (a == ReligionCatholic && b == ReligionOrthodox) || (a == ReligionOrthodox && b == ReligionCatholic) {
		return -20
	}
	return -30
}

// FactionID fraksiyon benzersiz kimliği.
type FactionID string

// ResearchState bir fraksiyonun teknoloji araştırma durumu.
// tech paketi bu struct üzerinde çalışan yardımcı fonksiyonlar sağlar.
type ResearchState struct {
	Completed map[string]bool `json:"completed"`
	ActiveID  string          `json:"active_id"`
	TurnsLeft int             `json:"turns_left"`
}

// Faction oyundaki bir fraksiyonu temsil eder.
type Faction struct {
	ID           FactionID `json:"id"`
	Name         string    `json:"name"`
	NameTR       string    `json:"name_tr"`
	Religion     Religion  `json:"religion"`
	Color        [3]uint8  `json:"color"`
	IsPlayable   bool      `json:"is_playable"`
	IsEliminated bool      `json:"is_eliminated"`

	Gold   int `json:"gold"`
	Grain  int `json:"grain"`
	Iron   int `json:"iron"`
	Timber int `json:"timber"`
	Spice  int `json:"spice"`
	Cloth  int `json:"cloth"`

	// Teknoloji araştırma durumu
	Research ResearchState `json:"research"`

	AIAggressiveness int `json:"ai_aggressiveness"`
}

// DiplomaticStance iki fraksiyon arasındaki ilişki durumu.
type DiplomaticStance string

const (
	StanceWar    DiplomaticStance = "war"
	StancePeace  DiplomaticStance = "peace"
	StanceAllied DiplomaticStance = "allied"
	StanceTrade  DiplomaticStance = "trade"
)

// Relation iki fraksiyon arasındaki tam ilişkiyi tutar.
type Relation struct {
	FactionA FactionID        `json:"faction_a"`
	FactionB FactionID        `json:"faction_b"`
	Score    int              `json:"score"`
	Stance   DiplomaticStance `json:"stance"`
}

// RelationKey iki fraksiyon için sıralı anahtar üretir.
func RelationKey(a, b FactionID) string {
	if a < b {
		return string(a) + "|" + string(b)
	}
	return string(b) + "|" + string(a)
}
