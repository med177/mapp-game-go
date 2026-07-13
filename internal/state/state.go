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

// MaxDiplomacyOffersPerTurn bir devletin bir turda gönderebileceği azami teklif sayısıdır.
const MaxDiplomacyOffersPerTurn = 3

// VictoryType zafer koşulu türü.
type VictoryType string

const (
	VictoryDomination   VictoryType = "domination"    // bölge sayısı + kritik şehirler
	VictoryEconomic     VictoryType = "economic"      // altın gelir hedefi
	VictoryMilitary     VictoryType = "military"      // ordu gücü + yenilgiler
	VictoryReligious    VictoryType = "religious"     // kutsal şehirleri tut
	VictoryConquerCity  VictoryType = "conquer_city"  // tek hedef bölgeyi ele geçir
	VictorySurviveTurns VictoryType = "survive_turns" // belirli tur sayısı boyunca ayakta kal
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
	TargetTurns        int              `json:"turns"`                // survive_turns
	DeadlineYear       int              `json:"deadline_year"`        // 0 = süresiz
	DeadlineMonth      int              `json:"deadline_month"`       // 1-12, 0 = yıl sonu
}

// DiplomaticOffer AI/oyuncu arasında bekleyen diplomatik teklif kaydıdır.
type DiplomaticOffer struct {
	FromFactionID        faction.FactionID `json:"from_faction_id"`
	ToFactionID          faction.FactionID `json:"to_faction_id"`
	Action               string            `json:"action"`
	CreatedTurn          int               `json:"created_turn"`
	Priority             int               `json:"priority,omitempty"`
	PriorityReason       string            `json:"priority_reason,omitempty"`
	WarDeclarerFactionID faction.FactionID `json:"war_declarer_faction_id,omitempty"`
	WarEnemyFactionID    faction.FactionID `json:"war_enemy_faction_id,omitempty"`
}

// DiplomaticOfferHistoryEntry çözümlenmiş diplomatik tekliflerin kısa geçmiş kaydıdır.
type DiplomaticOfferHistoryEntry struct {
	FromFactionID        faction.FactionID `json:"from_faction_id"`
	ToFactionID          faction.FactionID `json:"to_faction_id"`
	Action               string            `json:"action"`
	CreatedTurn          int               `json:"created_turn"`
	ResolvedTurn         int               `json:"resolved_turn"`
	Accepted             bool              `json:"accepted"`
	Applied              bool              `json:"applied"`
	Priority             int               `json:"priority,omitempty"`
	PriorityReason       string            `json:"priority_reason,omitempty"`
	ResultMessage        string            `json:"result_message,omitempty"`
	WarDeclarerFactionID faction.FactionID `json:"war_declarer_faction_id,omitempty"`
	WarEnemyFactionID    faction.FactionID `json:"war_enemy_faction_id,omitempty"`
}

type SiegeState struct {
	RegionID             world.RegionID `json:"region_id"`
	AttackerArmyID       army.ArmyID    `json:"attacker_army_id"`
	AttackerHomeRegionID world.RegionID `json:"attacker_home_region_id,omitempty"`
	DefenderArmyID       army.ArmyID    `json:"defender_army_id,omitempty"`
	AttackerFactionID    string         `json:"attacker_faction_id"`
	StartedTurn          int            `json:"started_turn"`
	TurnsElapsed         int            `json:"turns_elapsed"`
	FortLevel            int            `json:"fort_level"`
	BreachProgress       int            `json:"breach_progress"`
	BreachLevel          int            `json:"breach_level"`
}

// SiegeSurrenderTurns tahkimat seviyesine göre kuşatmanın kaç turda teslim olacağını döner.
func SiegeSurrenderTurns(fortLevel int) int {
	if fortLevel < 1 {
		fortLevel = 1
	}
	return 6 + fortLevel*4
}

// TurnsUntilSurrender kuşatmanın teslim olmasına kaç tur kaldığını döner.
// Negatif değerler 0 olarak döner.
func (s *SiegeState) TurnsUntilSurrender() int {
	if s == nil {
		return 0
	}
	total := SiegeSurrenderTurns(s.FortLevel)
	remaining := total - s.TurnsElapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RegionEventStatus bir bölgede aktif olan event görünürlük kaydını tutar.
// Event choice sonrası veya otomatik çözümlenen event'ler sonrası haritada
// birkaç tur boyunca ikon gösterimi için kullanılır.
type RegionEventStatus struct {
	EventID   string         `json:"event_id"`
	RegionID  world.RegionID `json:"region_id"`
	TurnsLeft int            `json:"turns_left"` // kaç tur daha görünür kalacak
	Type      string         `json:"type"`       // plague, famine, blessing, revolt, notification
	LabelTR   string         `json:"label_tr"`   // kısa açıklama (tooltip için)
}

// GameState oyunun tüm anlık durumunu tutar. Save/load ham struct snapshot'ı
// yerine bu state'in mutable campaign alanlarını serialize eder ve senaryo baz
// state'i yükleme sırasında yeniden kurar.
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
	Victory                 VictoryCondition `json:"victory"`
	SelectedVictoryOptionID string           `json:"selected_victory_option_id"`

	// Dünya verisi
	Regions      map[world.RegionID]*world.Region       `json:"regions"`
	RegionOrder  []world.RegionID                       `json:"-"`
	Factions     map[faction.FactionID]*faction.Faction `json:"factions"`
	FactionOrder []faction.FactionID                    `json:"-"`
	Armies       map[army.ArmyID]*army.Army             `json:"armies"`
	ShapeData    world.CountryShapeJSON                 `json:"-"`

	// Runtime-only (json:"-") — her başlangıçta assets'ten yüklenir
	UnitTypes          map[string]*army.UnitType                `json:"-"`
	BuildingTypes      map[string]*city.Building                `json:"-"`
	TechTypes          map[string]*tech.Technology              `json:"-"`
	ScenarioVictories  []scenario.VictoryOptionDef              `json:"-"`
	AvailableVictories []scenario.VictoryOptionDef              `json:"-"`
	RegionLogistics    map[world.RegionID]RegionLogisticsStatus `json:"-"`
	ArmyLogistics      map[army.ArmyID]ArmyLogisticsStatus      `json:"-"`

	// Zafer takibi
	EconomicVictoryTurns  int  `json:"economic_victory_turns"`
	FactionsEliminated    int  `json:"factions_eliminated"`
	ReligiousVictoryTurns int  `json:"religious_victory_turns"`
	VictoryAchieved       bool `json:"victory_achieved"`
	VictoryAchievedTurn   int  `json:"victory_achieved_turn"`

	// Tetiklenmiş tek seferlik olay ID'leri
	FiredEventIDs map[string]bool `json:"fired_event_ids"`

	// Diplomatik ilişkiler (key: RelationKey)
	Relations map[string]*faction.Relation `json:"relations"`
	// Bekleyen diplomatik teklifler (ör. AI barış teklifi)
	DiplomaticOffers []DiplomaticOffer `json:"diplomatic_offers,omitempty"`
	// Çözümlenmiş diplomatik tekliflerin kısa geçmişi.
	DiplomaticOfferHistory []DiplomaticOfferHistoryEntry `json:"diplomatic_offer_history,omitempty"`
	// Turn içinde devlet başına gönderilen diplomasi teklif sayacı.
	DiplomacyOfferCounts map[faction.FactionID]int `json:"diplomacy_offer_counts,omitempty"`

	// Ticaret güzergahları
	TradeRoutes  []*economy.TradeRoute          `json:"trade_routes"`
	TradeCenters world.TradeCenterConfig        `json:"trade_centers,omitempty"` // senaryo bazlı tarihsel ticaret merkezleri + link graph
	Sieges       map[world.RegionID]*SiegeState `json:"sieges,omitempty"`

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

	// Aktif bölge event ikonları (haritada birkaç tur görünür kalır)
	ActiveRegionEvents []RegionEventStatus `json:"active_region_events,omitempty"`
}

// RegionProductionSummary bir bölgenin tur başı efektif ekonomik katkısını özetler.
// Bu hesap UI önizlemeleri için ekonomi çözümlemesindeki bina, arazi, mevsim ve
// sahip fraksiyon teknoloji çarpanlarını aynı sırayla uygular.
type RegionProductionSummary struct {
	Gold   int
	Grain  int
	Iron   int
	Timber int
	Stone  int
	Spice  int
	Cloth  int
}

type RegionLogisticsStatus struct {
	RegionID          world.RegionID
	OwnerID           string
	LocalProduction   int
	SettlementBuffer  int
	ReserveSupport    int
	Demand            int
	Capacity          int
	Overload          int
	ArmyCount         int
	UnitsAffected     int
	UnitsLost         int
	TotalHPDamage     int
	PeakOverloadTurns int
}

type ArmyLogisticsStatus struct {
	ArmyID            army.ArmyID
	RegionID          world.RegionID
	OwnerID           string
	Demand            int
	Capacity          int
	Overload          int
	OverCapacityTurns int
	DamagePerUnit     int
	UnitsAffected     int
	UnitsLost         int
	TotalHPDamage     int
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

// ClearProductionOrdersForRegion belirtilen bölgedeki tüm üretim emirlerini kuyruktan siler.
// Bölge el değiştirince mevcut yapı/inşaat ve eğitim emirleri artık devam etmez.
func (s *GameState) ClearProductionOrdersForRegion(regionID world.RegionID) int {
	if s == nil || regionID == "" || len(s.ProductionQueue) == 0 {
		return 0
	}
	remaining := s.ProductionQueue[:0]
	removed := 0
	for _, order := range s.ProductionQueue {
		if order.RegionID == regionID {
			removed++
			continue
		}
		remaining = append(remaining, order)
	}
	if removed == 0 {
		return 0
	}
	for i := len(remaining); i < len(s.ProductionQueue); i++ {
		s.ProductionQueue[i] = ProductionOrder{}
	}
	s.ProductionQueue = remaining
	return removed
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
	s.ResetDiplomacyOfferCounts()
}

// ResetDiplomacyOfferCounts mevcut tur teklif sayaçlarını sıfırlar.
func (s *GameState) ResetDiplomacyOfferCounts() {
	if s == nil || len(s.DiplomacyOfferCounts) == 0 {
		s.DiplomacyOfferCounts = nil
		return
	}
	s.DiplomacyOfferCounts = nil
}

// DiplomacyOfferQuotaUsed belirtilen fraksiyonun bu tur kullandığı teklif sayısını döner.
func (s *GameState) DiplomacyOfferQuotaUsed(fid faction.FactionID) int {
	if s == nil || fid == "" || len(s.DiplomacyOfferCounts) == 0 {
		return 0
	}
	return s.DiplomacyOfferCounts[fid]
}

// DiplomacyOfferQuotaRemaining belirtilen fraksiyonun bu tur kalan teklif hakkını döner.
func (s *GameState) DiplomacyOfferQuotaRemaining(fid faction.FactionID) int {
	if s == nil {
		return MaxDiplomacyOffersPerTurn
	}
	remaining := MaxDiplomacyOffersPerTurn - s.DiplomacyOfferQuotaUsed(fid)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanSpendDiplomacyOfferQuota belirtilen fraksiyon için en az bir teklif hakkı olup olmadığını döner.
func (s *GameState) CanSpendDiplomacyOfferQuota(fid faction.FactionID) bool {
	return s != nil && fid != "" && s.DiplomacyOfferQuotaRemaining(fid) > 0
}

// SpendDiplomacyOfferQuota belirtilen fraksiyonun teklif hakkını bir artırır.
func (s *GameState) SpendDiplomacyOfferQuota(fid faction.FactionID) bool {
	if !s.CanSpendDiplomacyOfferQuota(fid) {
		return false
	}
	if s.DiplomacyOfferCounts == nil {
		s.DiplomacyOfferCounts = make(map[faction.FactionID]int, 4)
	}
	s.DiplomacyOfferCounts[fid]++
	return true
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

// SelectBattleDefender hedef bölgede saldıranı karşılayacak düşman orduyu deterministik seçer.
func (s *GameState) SelectBattleDefender(attacker *army.Army, target world.RegionID, navalSeaMove bool) *army.Army {
	if s == nil || attacker == nil {
		return nil
	}
	var best *army.Army
	bestPower := -1
	for _, candidate := range s.Armies {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID {
			continue
		}
		key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(candidate.OwnerID))
		rel, exists := s.Relations[key]
		if !exists || rel == nil || rel.Stance != faction.StanceWar {
			continue
		}
		power := 0
		if s.UnitTypes != nil {
			power = candidate.TotalStrength(s.UnitTypes)
		}
		if best == nil || power > bestPower || (power == bestPower && string(candidate.ID) < string(best.ID)) {
			best = candidate
			bestPower = power
		}
	}
	return best
}

// CollectDefenders hedef bölgede saldırana karşı savaşacak TÜM düşman ordularını
// (düşmanın müttefikleri dahil) tek bir birleşik orduda toplar.
// Dönen ordu sanaldır — gerçek Army map'ine eklenmez, sadece savaş simülasyonu içindir.
// Ayrıca birleştirilen gerçek ordu ID'lerinin listesini döner ki kayıplar dağıtılabilsin.
func (s *GameState) CollectDefenders(attacker *army.Army, target world.RegionID, navalSeaMove bool) (combined *army.Army, sourceIDs []army.ArmyID) {
	if s == nil || attacker == nil {
		return nil, nil
	}
	var units []army.Unit
	for _, candidate := range s.Armies {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID {
			continue
		}
		// Deniz savaşında sadece savaş halindekiler; kara savaşında savaş halindeki herkes
		if navalSeaMove {
			key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(candidate.OwnerID))
			rel, exists := s.Relations[key]
			if !exists || rel == nil || rel.Stance != faction.StanceWar {
				continue
			}
		} else {
			// Kara hedefinde: hedef bölge sahibiyle savaş halindeysek,
			// bölgedeki saldırana savaş açmış TÜM ordular savunmaya katılır
			if candidate.OwnerID == "" {
				continue
			}
			key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(candidate.OwnerID))
			rel, exists := s.Relations[key]
			if !exists || rel == nil || rel.Stance != faction.StanceWar {
				continue
			}
		}
		units = append(units, candidate.Units...)
		sourceIDs = append(sourceIDs, candidate.ID)
	}
	if len(units) == 0 {
		return nil, nil
	}
	// 20 birim sınırına uygula
	if len(units) > 20 {
		units = units[:20]
	}
	combined = &army.Army{
		OwnerID: attacker.OwnerID, // geçici, sadece simülasyon için
		Units:   units,
	}
	return combined, sourceIDs
}

// DistributeDefenderLosses birleşik savunma ordusuna verilen kayıpları
// kaynak ordulara orantılı olarak dağıtır.
func (s *GameState) DistributeDefenderLosses(sourceIDs []army.ArmyID, totalLost int) {
	if s == nil || len(sourceIDs) == 0 || totalLost <= 0 {
		return
	}
	remaining := totalLost
	for _, id := range sourceIDs {
		a := s.Armies[id]
		if a == nil || len(a.Units) == 0 {
			continue
		}
		canLose := len(a.Units)
		lose := (totalLost * canLose) / (totalLost + canLose) // basit orantı
		if lose > canLose {
			lose = canLose
		}
		if lose > remaining {
			lose = remaining
		}
		if lose <= 0 {
			continue
		}
		a.Units = a.Units[:len(a.Units)-lose]
		remaining -= lose
		if len(a.Units) == 0 {
			delete(s.Armies, id)
		}
		if remaining <= 0 {
			break
		}
	}
}

func (s *GameState) SiegeAt(regionID world.RegionID) *SiegeState {
	if s == nil || s.Sieges == nil || regionID == "" {
		return nil
	}
	return s.Sieges[regionID]
}

func (s *GameState) SiegeByArmy(armyID army.ArmyID) *SiegeState {
	if s == nil || s.Sieges == nil || armyID == "" {
		return nil
	}
	for _, siege := range s.Sieges {
		if siege != nil && siege.AttackerArmyID == armyID {
			return siege
		}
	}
	return nil
}

// CanJoinActiveSiege aktif bir kuşatmaya mevcut ordunun destek için katılıp katılamayacağını döner.
// Aynı fraksiyon ya da müttefik fraksiyonlar destek verebilir; kuşatmayı başlatan ordu hariç tutulur.
func (s *GameState) CanJoinActiveSiege(attacker *army.Army, regionID world.RegionID) bool {
	if s == nil || attacker == nil || s.Sieges == nil || regionID == "" || attacker.OwnerID == "" {
		return false
	}
	siege := s.Sieges[regionID]
	if siege == nil || siege.AttackerArmyID == "" || siege.AttackerArmyID == attacker.ID {
		return false
	}
	siegeArmy := s.Armies[siege.AttackerArmyID]
	if siegeArmy == nil || siegeArmy.OwnerID == "" {
		return false
	}
	if siegeArmy.OwnerID == attacker.OwnerID {
		return true
	}
	if stateSameRealm(s, faction.FactionID(attacker.OwnerID), faction.FactionID(siegeArmy.OwnerID)) {
		return true
	}
	key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(siegeArmy.OwnerID))
	rel := s.Relations[key]
	return rel != nil && rel.Stance == faction.StanceAllied
}

// CanEnterActiveSiegedRegion aktif kuşatma altındaki bölgeye girilebileceğini döner.
// Kuşatma destekçileri ve kuşatan tarafa karşı savaşta olan ordular girebilir.
func (s *GameState) CanEnterActiveSiegedRegion(attacker *army.Army, regionID world.RegionID) bool {
	if s == nil || attacker == nil || s.Sieges == nil || regionID == "" || attacker.OwnerID == "" {
		return false
	}
	siege := s.Sieges[regionID]
	if siege == nil || siege.AttackerArmyID == "" {
		return false
	}
	if siege.AttackerArmyID == attacker.ID {
		return true
	}
	if s.CanJoinActiveSiege(attacker, regionID) {
		return true
	}
	siegeArmy := s.Armies[siege.AttackerArmyID]
	if siegeArmy == nil || siegeArmy.OwnerID == "" {
		return false
	}
	key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(siegeArmy.OwnerID))
	rel := s.Relations[key]
	return rel != nil && rel.Stance == faction.StanceWar
}

func stateDirectOverlord(s *GameState, fid faction.FactionID) faction.FactionID {
	if s == nil || fid == "" {
		return ""
	}
	f := s.Factions[fid]
	if f == nil || f.IsEliminated || f.OverlordID == "" || f.OverlordID == fid {
		return ""
	}
	overlord := s.Factions[f.OverlordID]
	if overlord == nil || overlord.IsEliminated {
		return ""
	}
	return f.OverlordID
}

func stateRealmRoot(s *GameState, fid faction.FactionID) faction.FactionID {
	if fid == "" {
		return ""
	}
	current := fid
	seen := map[faction.FactionID]struct{}{}
	for {
		overlord := stateDirectOverlord(s, current)
		if overlord == "" {
			return current
		}
		if _, exists := seen[overlord]; exists {
			return current
		}
		seen[overlord] = struct{}{}
		current = overlord
	}
}

func stateSameRealm(s *GameState, a, b faction.FactionID) bool {
	return a != "" && b != "" && stateRealmRoot(s, a) == stateRealmRoot(s, b)
}

// RegionProductionSummary hesaplanan efektif bölge üretimini döner.
func (s *GameState) RegionProductionSummary(region *world.Region) RegionProductionSummary {
	if s == nil || region == nil || region.IsSea || region.OwnerID == "" {
		return RegionProductionSummary{}
	}

	goldMod := 1.0
	grainMod := 1.0
	tradeCapMod := 1.0
	for _, bid := range region.Buildings {
		if b, ok := s.BuildingTypes[bid]; ok && b != nil {
			goldMod *= b.GoldMod
			grainMod *= b.GrainMod
			tradeCapMod *= b.TradeCapacityMod
		}
	}

	out := RegionProductionSummary{
		Gold:   int(float64(region.GoldIncome()) * goldMod * float64(s.CurrentSeason().HarvestMod()) / 100),
		Grain:  int(float64(region.BaseGrainOutput) * grainMod),
		Iron:   region.BaseIronOutput,
		Timber: region.BaseTimberOutput,
		Stone:  region.BaseStoneOutput,
		Spice:  region.BaseSpiceOutput,
		Cloth:  region.BaseClothOutput,
	}
	out.Grain, out.Iron, out.Timber, out.Stone, out.Spice, out.Cloth = applyRegionTerrainSpecialization(
		region.Terrain, out.Grain, out.Iron, out.Timber, out.Stone, out.Spice, out.Cloth,
	)

	out.Gold += economy.RegionTradeIncome(region.TradeCapacity, tradeCapMod) * s.CurrentSeason().TradeMod() / 100

	if fx, ok := s.Factions[faction.FactionID(region.OwnerID)]; ok && fx != nil && s.TechTypes != nil {
		effects := tech.ComputeEffects(fx.Research.Completed, s.TechTypes)
		out.Gold += effects.GoldPerRegion
		out.Grain = int(float64(out.Grain) * (1.0 + effects.GrainMod))
		out.Iron = int(float64(out.Iron) * (1.0 + effects.IronMod))
		out.Timber = int(float64(out.Timber) * (1.0 + effects.TimberMod))
		out.Stone = int(float64(out.Stone) * (1.0 + effects.StoneMod))
	}

	if bonus := s.CapitalRegionBonus(region); bonus != (RegionProductionSummary{}) {
		out.Gold += bonus.Gold
		out.Grain += bonus.Grain
		out.Iron += bonus.Iron
		out.Timber += bonus.Timber
		out.Stone += bonus.Stone
		out.Spice += bonus.Spice
		out.Cloth += bonus.Cloth
	}

	return out
}

func applyRegionTerrainSpecialization(
	terrain world.TerrainType,
	grain, iron, timber, stone, spice, cloth int,
) (int, int, int, int, int, int) {
	switch terrain {
	case world.TerrainPlain:
		grain = grain * 120 / 100
	case world.TerrainForest:
		timber = timber * 130 / 100
	case world.TerrainMountain:
		iron = iron * 125 / 100
		if stone <= 0 {
			stone = 1 + iron/3
		}
		stone = stone * 140 / 100
	case world.TerrainPass:
		if stone <= 0 {
			stone = 1
		}
		stone = stone * 120 / 100
	}
	return grain, iron, timber, stone, spice, cloth
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

func buildingLevel(region *world.Region, buildingID string) int {
	if region == nil || buildingID == "" {
		return 0
	}
	level := 0
	for _, bid := range region.Buildings {
		if bid == buildingID {
			level++
		}
	}
	return level
}

// LandUnitProductionLimit bir kara bölgesinin tur başına tamamlayabileceği kara birimi adedini döner.
// Milis gibi bina gereksinimi olmayan temel birlik akışını kırmamak için taban değer 1 korunur.
func LandUnitProductionLimit(region *world.Region) int {
	if region == nil || region.IsSea {
		return 0
	}
	limit := buildingLevel(region, "barracks")
	if limit < 1 {
		limit = 1
	}
	return limit
}

// NavalUnitProductionLimit bir kara bölgesinin tur başına tamamlayabileceği deniz birimi adedini döner.
func NavalUnitProductionLimit(region *world.Region) int {
	if region == nil || region.IsSea {
		return 0
	}
	return buildingLevel(region, "port")
}

// UnitProductionLimit birim tipine göre ilgili kışla/liman üretim hattının tur limitini döner.
func UnitProductionLimit(region *world.Region, unitType *army.UnitType) int {
	if unitType != nil && unitType.RequiredBldg == "port" {
		return NavalUnitProductionLimit(region)
	}
	return LandUnitProductionLimit(region)
}

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
		if a != nil && a.OwnerID == string(fid) && a.CountsTowardArmyLimit() {
			count++
		}
	}
	return count
}
