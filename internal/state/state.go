package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/season"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// VictoryType zafer koşulu türü.
type VictoryType string

const (
	VictoryDomination VictoryType = "domination" // bölge sayısı + kritik şehirler
	VictoryEconomic   VictoryType = "economic"   // altın gelir hedefi
	VictoryMilitary   VictoryType = "military"   // ordu gücü + yenilgiler
	VictoryReligious  VictoryType = "religious"  // kutsal şehirleri tut
)

// VictoryCondition seçilen zafer koşulunu tutar.
type VictoryCondition struct {
	Type               VictoryType      `json:"type"`
	TargetRegionCount  int              `json:"target_region_count"`  // domination
	RequiredRegions    []world.RegionID `json:"required_regions"`     // domination
	TargetGoldIncome   int              `json:"target_gold_income"`   // economic
	GoldHoldTurns      int              `json:"gold_hold_turns"`      // economic — kaç tur sürdürülmeli
	TargetArmyStrength int              `json:"target_army_strength"` // military
	TargetDefeated     int              `json:"target_defeated"`      // military — kaç fraksiyon yenilgisi
	DeadlineTurn       int              `json:"deadline_turn"`        // 0 = süresiz
}

// GameState oyunun tüm anlık durumunu tutar. Save/load bu struct'ı serialize eder.
type GameState struct {
	// Zaman
	Turn      int `json:"turn"`  // toplam tur sayısı (1'den başlar)
	Year      int `json:"year"`  // 1300-1600
	Month     int `json:"month"` // 1-12
	StartYear int `json:"start_year"`

	// Oyuncu
	PlayerFactionID faction.FactionID `json:"player_faction_id"`
	Difficulty      int               `json:"difficulty"` // 1=kolay, 2=normal, 3=zor

	// Zafer koşulu
	Victory VictoryCondition `json:"victory"`

	// Dünya verisi
	Regions   map[world.RegionID]*world.Region       `json:"regions"`
	Factions  map[faction.FactionID]*faction.Faction `json:"factions"`
	Armies    map[army.ArmyID]*army.Army             `json:"armies"`
	ShapeData world.CountryShapeJSON                 `json:"-"`

	// Runtime-only (json:"-") — her başlangıçta assets'ten yüklenir
	UnitTypes     map[string]*army.UnitType   `json:"-"`
	BuildingTypes map[string]*city.Building   `json:"-"`
	TechTypes     map[string]*tech.Technology `json:"-"`

	// Zafer takibi
	EconomicVictoryTurns  int `json:"economic_victory_turns"`
	FactionsEliminated    int `json:"factions_eliminated"`
	ReligiousVictoryTurns int `json:"religious_victory_turns"`

	// Tetiklenmiş tek seferlik olay ID'leri
	FiredEventIDs map[string]bool `json:"fired_event_ids"`

	// Diplomatik ilişkiler (key: RelationKey)
	Relations map[string]*faction.Relation `json:"relations"`

	// Ticaret güzergahları
	TradeRoutes []*economy.TradeRoute `json:"trade_routes"`

	// Sıradaki ordu ID üretmek için sayaç
	NextArmySeq int `json:"next_army_seq"`

	// Oyun aşaması
	Phase Phase `json:"phase"`

	// Kazanan (boş = oyun devam ediyor)
	WinnerID faction.FactionID `json:"winner_id"`
}

// Phase oyun aşaması.
type Phase string

const (
	PhaseMainMenu       Phase = "main_menu" // ana menü
	PhaseSettings       Phase = "settings"  // ayarlar ekranı
	PhaseFactionSelect  Phase = "faction_select"
	PhaseVictorySelect  Phase = "victory_select"
	PhasePlayerTurn     Phase = "player_turn"
	PhaseAITurn         Phase = "ai_turn"
	PhaseTurnResolution Phase = "resolution"
	PhaseGameOver       Phase = "game_over"
)

// CurrentSeason mevcut mevsimi döner.
func (s *GameState) CurrentSeason() season.Season {
	return season.FromMonth(s.Month)
}

// AdvanceTurn turu bir ileri alır, ay/yıl günceller.
func (s *GameState) AdvanceTurn() {
	s.Turn++
	s.Month++
	if s.Month > 12 {
		s.Month = 1
		s.Year++
	}
}

// RegionsOwnedBy bir fraksiyonun sahip olduğu bölge listesini döner.
func (s *GameState) RegionsOwnedBy(fid faction.FactionID) []*world.Region {
	var result []*world.Region
	for _, r := range s.Regions {
		if r.OwnerID == string(fid) {
			result = append(result, r)
		}
	}
	return result
}

// IsEliminated bir fraksiyon elenmiş mi kontrol eder.
func (s *GameState) IsEliminated(fid faction.FactionID) bool {
	return len(s.RegionsOwnedBy(fid)) == 0
}
