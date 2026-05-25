package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/city"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/season"
	"mapp-game-go/internal/tech"
	"mapp-game-go/internal/world"
)

// VictoryType zafer koşulu türü.
type VictoryType string

const (
	VictoryDomination  VictoryType = "domination"   // bölge sayısı + kritik şehirler
	VictoryEconomic    VictoryType = "economic"     // altın gelir hedefi
	VictoryMilitary    VictoryType = "military"     // ordu gücü + yenilgiler
	VictoryReligious   VictoryType = "religious"    // kutsal şehirleri tut
	VictoryConquerCity VictoryType = "conquer_city" // tek hedef bölgeyi ele geçir
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

// DiplomaticOffer AI/oyuncu arasında bekleyen diplomatik teklif kaydıdır.
type DiplomaticOffer struct {
	FromFactionID faction.FactionID `json:"from_faction_id"`
	ToFactionID   faction.FactionID `json:"to_faction_id"`
	Action        string            `json:"action"`
	CreatedTurn   int               `json:"created_turn"`
}

// GameState oyunun tüm anlık durumunu tutar. Save/load bu struct'ı serialize eder.
type GameState struct {
	// Zaman
	Turn      int `json:"turn"`  // toplam tur sayısı (1'den başlar)
	Year      int `json:"year"`  // 1300-1600
	Month     int `json:"month"` // 1-12
	StartYear int `json:"start_year"`

	// Senaryo
	ScenarioID   string             `json:"scenario_id"`   // aktif senaryo ID'si
	ScenarioPath string             `json:"scenario_path"` // aktif senaryo klasörü
	MapConfig    scenario.MapConfig `json:"map"`           // aktif senaryonun harita hizalama ayarları

	// Oyuncu
	PlayerFactionID faction.FactionID `json:"player_faction_id"`
	Difficulty      int               `json:"difficulty"` // 1=kolay, 2=normal, 3=zor

	// Development mode
	DevelopmentMode bool `json:"development_mode"`
	EditMode        bool `json:"edit_mode"`

	// Zafer koşulu
	Victory VictoryCondition `json:"victory"`

	// Dünya verisi
	Regions     map[world.RegionID]*world.Region       `json:"regions"`
	RegionOrder []world.RegionID                       `json:"-"`
	Factions    map[faction.FactionID]*faction.Faction `json:"factions"`
	Armies      map[army.ArmyID]*army.Army             `json:"armies"`
	ShapeData   world.CountryShapeJSON                 `json:"-"`

	// Runtime-only (json:"-") — her başlangıçta assets'ten yüklenir
	UnitTypes          map[string]*army.UnitType   `json:"-"`
	BuildingTypes      map[string]*city.Building   `json:"-"`
	TechTypes          map[string]*tech.Technology `json:"-"`
	AvailableVictories []scenario.VictoryOptionDef `json:"-"`

	// Zafer takibi
	EconomicVictoryTurns  int `json:"economic_victory_turns"`
	FactionsEliminated    int `json:"factions_eliminated"`
	ReligiousVictoryTurns int `json:"religious_victory_turns"`
	VictoryAchieved       bool `json:"victory_achieved"`
	VictoryAchievedTurn   int  `json:"victory_achieved_turn"`

	// Tetiklenmiş tek seferlik olay ID'leri
	FiredEventIDs map[string]bool `json:"fired_event_ids"`

	// Diplomatik ilişkiler (key: RelationKey)
	Relations map[string]*faction.Relation `json:"relations"`
	// Bekleyen diplomatik teklifler (ör. AI barış teklifi)
	DiplomaticOffers []DiplomaticOffer `json:"diplomatic_offers,omitempty"`

	// Ticaret güzergahları
	TradeRoutes []*economy.TradeRoute  `json:"trade_routes"`
	TradeCenters world.TradeCenterConfig `json:"trade_centers,omitempty"` // senaryo bazlı tarihsel ticaret merkezleri + link graph

	// Dinamik piyasa fiyatları (her tur sonu güncellenir)
	MarketPrices economy.CurrentMarketPrice `json:"-"`

	// Devam eden üretimler
	ProductionQueue   []ProductionOrder `json:"production_queue"`
	NextProductionSeq int               `json:"next_production_seq"`

	// Sıradaki ordu ID üretmek için sayaç
	NextArmySeq int `json:"next_army_seq"`

	// Oyun aşaması
	Phase Phase `json:"phase"`

	// Kazanan (boş = oyun devam ediyor)
	WinnerID faction.FactionID `json:"winner_id"`

	// Region paint overrides - edit modunda bölge boyama değişiklikleri (piksel indeksi -> bölge ID)
	RegionPaintOverrides map[int]world.RegionID `json:"region_paint_overrides,omitempty"`
}

// ProductionOrder bina ve birim üretimlerinin tur bazlı kuyruğunu tutar.
type ProductionOrder struct {
	ID        string         `json:"id"`
	Kind      string         `json:"kind"` // "building" veya "unit"
	FactionID string         `json:"faction_id"`
	RegionID  world.RegionID `json:"region_id"`
	TypeID    string         `json:"type_id"`
	TurnsLeft int            `json:"turns_left"`
}

// Phase oyun aşaması.
type Phase string

const (
	PhaseMainMenu       Phase = "main_menu"       // ana menü
	PhaseSettings       Phase = "settings"        // ayarlar ekranı
	PhaseScenarioSelect Phase = "scenario_select" // senaryo seçim ekranı
	PhaseFactionSelect  Phase = "faction_select"
	PhaseVictorySelect  Phase = "victory_select"
	PhasePlayerTurn     Phase = "player_turn"
	PhaseAITurn         Phase = "ai_turn"
	PhaseTurnResolution Phase = "resolution"
	PhaseGameOver       Phase = "game_over"
	PhaseLoading        Phase = "loading"
	PhasePauseMenu      Phase = "pause_menu"  // oyun içi duraklama menüsü
	PhaseLoadSelect     Phase = "load_select" // kayıt seçim ekranı
	PhaseSaveSelect     Phase = "save_select" // slot seçerek kaydetme ekranı
	PhaseEditMode       Phase = "edit_mode"   // senaryo veri düzenleme modu
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

// SyncTimedRegionUnlocks aktif tur UnlockTurn'a ulaşmış kilitli bölgeleri açar.
// UnlockTurn=0 olan bölgeler zaman bazlı değil, başka sistemlerle açılır.
func (s *GameState) SyncTimedRegionUnlocks() []world.RegionID {
	unlocked := make([]world.RegionID, 0)
	for _, r := range s.Regions {
		if r == nil || !r.IsLocked || r.UnlockTurn <= 0 {
			continue
		}
		if s.Turn >= r.UnlockTurn {
			r.IsLocked = false
			unlocked = append(unlocked, r.ID)
		}
	}
	return unlocked
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

// LandRegionsOwnedBy bir fraksiyonun sahip olduğu kara bölgelerini döner.
func (s *GameState) LandRegionsOwnedBy(fid faction.FactionID) []*world.Region {
	var result []*world.Region
	for _, r := range s.Regions {
		if r.OwnerID == string(fid) && !r.IsSea {
			result = append(result, r)
		}
	}
	return result
}

// IsEliminated bir fraksiyon elenmiş mi kontrol eder.
func (s *GameState) IsEliminated(fid faction.FactionID) bool {
	return len(s.LandRegionsOwnedBy(fid)) == 0
}

// ── Askeri Kapasite ───────────────────────────────────────────────────────

const (
	ManpowerPerRegion   = 5 // kara bölgesi başına temel birim kapasitesi
	ManpowerBarracksAdd = 5 // kışlası olan bölgenin ekstra kapasitesi
)

// ManpowerCap bir fraksiyonun toplam kara birimi kapasitesini döner.
func (s *GameState) ManpowerCap(fid faction.FactionID) int {
	cap := 0
	for _, r := range s.Regions {
		if r.OwnerID != string(fid) || r.IsSea {
			continue
		}
		cap += ManpowerPerRegion
		for _, bid := range r.Buildings {
			if bid == "barracks" {
				cap += ManpowerBarracksAdd
				break
			}
		}
	}
	return cap
}

// DeployedLandUnits bir fraksiyonun aktif kara ordu birim sayısını döner.
func (s *GameState) DeployedLandUnits(fid faction.FactionID) int {
	total := 0
	for _, a := range s.Armies {
		if a.OwnerID == string(fid) && !a.IsNaval {
			total += len(a.Units)
		}
	}
	return total
}

// MaxLandArmies bir fraksiyonun sahip olabileceği maksimum kara ordu sayısını döner.
// Her 2 kara bölgesi için 1 ordu; minimum 1.
func (s *GameState) MaxLandArmies(fid faction.FactionID) int {
	landCount := 0
	for _, r := range s.Regions {
		if r.OwnerID == string(fid) && !r.IsSea {
			landCount++
		}
	}
	max := (landCount + 1) / 2 // ceil(landCount/2)
	if max < 1 {
		max = 1
	}
	return max
}

// CurrentLandArmies bir fraksiyonun aktif kara ordu sayısını döner.
func (s *GameState) CurrentLandArmies(fid faction.FactionID) int {
	count := 0
	for _, a := range s.Armies {
		if a.OwnerID == string(fid) && !a.IsNaval {
			count++
		}
	}
	return count
}
