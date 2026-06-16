package faction

import "mapp-game-go/internal/religion"

// FactionID fraksiyon benzersiz kimliği.
type FactionID string

// ResearchState bir fraksiyonun teknoloji araştırma durumu.
// tech paketi bu struct üzerinde çalışan yardımcı fonksiyonlar sağlar.
type ResearchState struct {
	Completed   map[string]bool `json:"completed"`
	PausedTurns map[string]int  `json:"paused_turns,omitempty"`
	ActiveID    string          `json:"active_id"`
	TurnsLeft   int             `json:"turns_left"`
}

// Faction oyundaki bir fraksiyonu temsil eder.
type Faction struct {
	ID           FactionID     `json:"id"`
	Name         string        `json:"name"`
	NameTR       string        `json:"name_tr"`
	Religion     religion.Type `json:"religion"`
	Color        [3]uint8      `json:"color"`
	IsPlayable   bool          `json:"is_playable"`
	IsEliminated bool          `json:"is_eliminated"`

	Gold   int `json:"gold"`
	Grain  int `json:"grain"`
	Iron   int `json:"iron"`
	Timber int `json:"timber"`
	Stone  int `json:"stone"`
	Spice  int `json:"spice"`
	Cloth  int `json:"cloth"`

	// Teknoloji araştırma durumu
	Research ResearchState `json:"research"`

	AIAggressiveness   int         `json:"ai_aggressiveness"`
	AIExpansionTargets []FactionID `json:"ai_expansion_targets,omitempty"`
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
