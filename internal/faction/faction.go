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

// TerritorialClaim bir fraksiyonun belirli bir bölge üzerindeki tarihsel veya
// stratejik talebini taşır. Base state yüklenirken AI stratejisi ve başlangıç
// sahipliğinden materialize edilir; barış değerlendirmesi bunu mevcut sahiplik
// ve aktif AI planıyla birlikte yorumlar.
type TerritorialClaim struct {
	RegionID string `json:"region_id"`
	Value    int    `json:"value"`
	Core     bool   `json:"core,omitempty"`
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
	// IsVirtual, isyan sırasında otomatik oluşturulan; diplomasi ve ticarete
	// kapalı, yalnız askeri hedef olarak var olan sanal isyancı devleti işaretler.
	IsVirtual  bool      `json:"is_virtual,omitempty"`
	OverlordID FactionID `json:"overlord_id,omitempty"`
	// TributeRate, vassalın overlord'a ödediği gelir payıdır (0-50).
	TributeRate           int  `json:"tribute_rate,omitempty"`
	TributeRateConfigured bool `json:"tribute_rate_configured,omitempty"`
	// VassalizedTurn, vassallık bağının kurulduğu toplam turu tutar. Bu alan
	// ilhak bekleme süresinin save/load sonrasında da korunmasını sağlar.
	VassalizedTurn int `json:"vassalized_turn,omitempty"`
	// CapitalSettlementID fraksiyonun aktif başkent settlement'ını tutar.
	CapitalSettlementID string `json:"capital_settlement_id,omitempty"`
	// HistoricalStartYear ve HistoricalEndYear, senaryo tarihçesinde devletin
	// tarihsel olarak ortaya çıktığı ve sona erdiği yılları taşır. Bu metadata
	// ardıl devletin oyun içi kuruluş koşulunu tek başına değiştirmez; kuruluş
	// yine bölgenin successor_faction_id akışıyla çözülür.
	HistoricalStartYear int `json:"historical_start_year,omitempty"`
	HistoricalEndYear   int `json:"historical_end_year,omitempty"`
	// PendingCapitalSettlementID başkent taşıma kuyruğundaki hedef settlement'tır.
	PendingCapitalSettlementID string `json:"pending_capital_settlement_id,omitempty"`
	// PendingCapitalTurns kalan başkent taşıma turudur.
	PendingCapitalTurns int `json:"pending_capital_turns,omitempty"`

	Gold   int `json:"gold"`
	Grain  int `json:"grain"`
	Iron   int `json:"iron"`
	Timber int `json:"timber"`
	Stone  int `json:"stone"`
	Spice  int `json:"spice"`
	Cloth  int `json:"cloth"`

	// Teknoloji araştırma durumu
	Research ResearchState `json:"research"`

	AIAggressiveness   int                `json:"ai_aggressiveness"`
	AIExpansionTargets []FactionID        `json:"ai_expansion_targets,omitempty"`
	TerritorialClaims  []TerritorialClaim `json:"territorial_claims,omitempty"`
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
	FactionA                 FactionID        `json:"faction_a"`
	FactionB                 FactionID        `json:"faction_b"`
	Score                    int              `json:"score"`
	Stance                   DiplomaticStance `json:"stance"`
	NextAIRelationRepairTurn int              `json:"next_ai_relation_repair_turn,omitempty"`
}

// RelationKey iki fraksiyon için sıralı anahtar üretir.
func RelationKey(a, b FactionID) string {
	if a < b {
		return string(a) + "|" + string(b)
	}
	return string(b) + "|" + string(a)
}
