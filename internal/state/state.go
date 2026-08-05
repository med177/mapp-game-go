package state

import (
	"sort"

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

// DiplomaticOfferRetryCooldownTurns reddedilen bir teklifin aynı aktör-hedef-
// aksiyon üçlüsü için zorunlu bekleme süresidir.
const DiplomaticOfferRetryCooldownTurns = 3

// civilianGrainPopulationUnit nüfusun aylık temel tahıl tüketim oranını taşır.
// 18 nüfus bir tahıl birimi tüketir; 1300 senaryosundaki üretim ve stoklar
// birlikte değerlendirildiğinde bu oran barışta küçük rezerv, savaşta açık
// oluşturacak tarihsel kıtlık baskısını korur.
const civilianGrainPopulationUnit = 18

// grainSaleGoldCapPercentOfTaxIncome acil/otomatik tahıl satışının bir turda
// üretebileceği altını mevcut temel vergi gelirine bağlar.
const grainSaleGoldCapPercentOfTaxIncome = 100

const (
	grainCivilianStorageMonths  = 6
	grainArmyStorageMonths      = 3
	grainMinimumStorageCapacity = 100
)

// GrainAidCost bir bölgeye tek seferlik sivil tahıl yardımı için gereken stoktur.
const GrainAidCost = 10

// GrainAidSatisfactionGain tahıl yardımının bölge memnuniyetine katkısıdır.
const GrainAidSatisfactionGain = 1

// AIDiagnosticHistoryEntry, geliştirme save'i yüklendikten sonra bir tam AI
// fazının karşılaştırmalı karar özetini taşır. Ayrıntılı runtime context yerine
// rapor için gerekli sabit alanlar tutulur.
type AIDiagnosticHistoryEntry struct {
	Turn                 int               `json:"turn"`
	FactionID            faction.FactionID `json:"faction_id"`
	PlanKind             AIObjectiveKind   `json:"plan_kind,omitempty"`
	PlanTargetFactionID  faction.FactionID `json:"plan_target_faction_id,omitempty"`
	TargetRegionID       world.RegionID    `json:"target_region_id,omitempty"`
	FrontCount           int               `json:"front_count"`
	ActiveWarCount       int               `json:"active_war_count"`
	ReservePercent       int               `json:"reserve_percent"`
	ReserveTargetPower   int               `json:"reserve_target_power"`
	ReserveAssignedPower int               `json:"reserve_assigned_power"`
	BlockReasons         []string          `json:"block_reasons,omitempty"`
}

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
	RegionID             world.RegionID    `json:"region_id,omitempty"`
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
	RegionID             world.RegionID    `json:"region_id,omitempty"`
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
	NavalLanding         bool           `json:"naval_landing,omitempty"`
	DefenderArmyID       army.ArmyID    `json:"defender_army_id,omitempty"`
	AttackerFactionID    string         `json:"attacker_faction_id"`
	StartedTurn          int            `json:"started_turn"`
	TurnsElapsed         int            `json:"turns_elapsed"`
	FortLevel            int            `json:"fort_level"`
	BreachProgress       int            `json:"breach_progress"`
	BreachLevel          int            `json:"breach_level"`
}

const (
	grainUpkeepStationaryPercent = 100
	grainUpkeepMovingPercent     = 150
	grainUpkeepGarrisonPercent   = 75
	grainUpkeepSiegeDefender     = 125
	grainUpkeepSiegeAttacker     = 200
	// Kendi toprağına bitişik kuşatmalar düzenli kara ikmalinden yararlanır.
	grainUpkeepSuppliedSiegeAttacker = 150

	capitalSupplyGraceDistance     = 2
	capitalSupplyDistanceStep      = 2
	capitalSupplyPenaltyPerStep    = 10
	capitalSupplyMaxPenalty        = 50
	capitalSupplyDisconnectedTax   = 50
	capitalSupplyFriendlyBorderTax = 10
)

// SiegeSurrenderTurns tahkimat seviyesine göre kuşatmanın kaç turda teslim olacağını döner.
func SiegeSurrenderTurns(fortLevel int) int {
	if fortLevel < 1 {
		fortLevel = 1
	}
	return 4 + fortLevel*2
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
	EventID                string         `json:"event_id"`
	RegionID               world.RegionID `json:"region_id"`
	TurnsLeft              int            `json:"turns_left"` // kaç tur daha görünür kalacak
	Type                   string         `json:"type"`       // plague, famine, blessing, revolt, notification
	LabelTR                string         `json:"label_tr"`   // kısa açıklama (tooltip için)
	GrainProductionPercent int            `json:"grain_production_percent,omitempty"`
	GrainDemandPercent     int            `json:"grain_demand_percent,omitempty"`
}

// GameState oyunun tüm anlık durumunu tutar. Save/load ham struct snapshot'ı
// yerine bu state'in mutable campaign alanlarını serialize eder ve senaryo baz
// state'i yükleme sırasında yeniden kurar.
type GameState struct {
	// Zaman
	Turn          int `json:"turn"`  // toplam tur sayısı (1'den başlar)
	Year          int `json:"year"`  // 1300-1600
	Month         int `json:"month"` // 1-12
	MonthsPerTurn int `json:"months_per_turn,omitempty"`
	StartYear     int `json:"start_year"`

	// Senaryo
	ScenarioID   string             `json:"scenario_id"`   // aktif senaryo ID'si
	ScenarioPath string             `json:"scenario_path"` // aktif senaryo klasörü
	MapConfig    scenario.MapConfig `json:"map"`           // aktif senaryonun harita hizalama ayarları

	// Oyuncu
	PlayerFactionID faction.FactionID `json:"player_faction_id"`
	AutoGrainExport bool              `json:"auto_grain_export,omitempty"`
	Difficulty      int               `json:"difficulty"` // 1=kolay, 2=normal, 3=zor

	// Development mode
	DevelopmentMode bool `json:"development_mode"`
	EditMode        bool `json:"edit_mode"`

	// Zafer koşulu
	Victory                 VictoryCondition `json:"victory"`
	SelectedVictoryOptionID string           `json:"selected_victory_option_id"`

	// Dünya verisi
	Regions                 map[world.RegionID]*world.Region       `json:"regions"`
	RegionOrder             []world.RegionID                       `json:"-"`
	LandPassages            []world.LandPassage                    `json:"land_passages,omitempty"`
	Factions                map[faction.FactionID]*faction.Faction `json:"factions"`
	FactionOrder            []faction.FactionID                    `json:"-"`
	Armies                  map[army.ArmyID]*army.Army             `json:"armies"`
	ArmyOrder               []army.ArmyID                          `json:"-"`
	Commanders              map[string]*army.Commander             `json:"commanders,omitempty"`
	CommanderArrivalNotices map[string]bool                        `json:"commander_arrival_notices,omitempty"`
	AIPlans                 map[faction.FactionID]*AIPlanState     `json:"ai_plans,omitempty"`
	// Imperial, bağımsız üyelerin HRE çağrı/otorite state'ini taşır. Gerçek
	// vassalların realm davranışı yine Faction.OverlordID ile belirlenir.
	Imperial  *ImperialState         `json:"imperial,omitempty"`
	ShapeData world.CountryShapeJSON `json:"-"`

	// Runtime-only (json:"-") — her başlangıçta assets'ten yüklenir
	AIStrategies       map[string]scenario.AIFactionStrategy    `json:"-"`
	AIDifficultyPolicy scenario.AIDifficultyPolicy              `json:"-"`
	UnitTypes          map[string]*army.UnitType                `json:"-"`
	BuildingTypes      map[string]*city.Building                `json:"-"`
	TechTypes          map[string]*tech.Technology              `json:"-"`
	CommanderTemplates map[string][]*army.Commander             `json:"-"`
	ScenarioVictories  []scenario.VictoryOptionDef              `json:"-"`
	AvailableVictories []scenario.VictoryOptionDef              `json:"-"`
	RegionLogistics    map[world.RegionID]RegionLogisticsStatus `json:"-"`
	ArmyLogistics      map[army.ArmyID]ArmyLogisticsStatus      `json:"-"`
	GrainEconomy       map[faction.FactionID]GrainEconomyStatus `json:"-"`
	GrainSaleGoldUsed  map[faction.FactionID]int                `json:"-"`

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
	// Aktif savaşların başlangıç durumu ve kalıcı kayıp/fetih sayaçları.
	WarLedgers map[string]*WarLedger `json:"war_ledgers,omitempty"`
	// Barış sonrası geçici ateşkes bitiş turları (relation key -> expiry turn).
	RecentTruces map[string]int `json:"recent_truces,omitempty"`
	// Geliştirme modunda save yükleme sonrası ilk AI turlarını karşılaştırmak
	// için tutulan geçici telemetri. Normal campaign payload'ına yazılmaz;
	// yalnız debug sidecar'a aktarılır.
	AIDiagnosticHistory            []AIDiagnosticHistoryEntry `json:"-"`
	AIDiagnosticCaptureTurnsRemain int                        `json:"-"`
	// Bekleyen diplomatik teklifler (ör. AI barış teklifi)
	DiplomaticOffers []DiplomaticOffer `json:"diplomatic_offers,omitempty"`
	// Çözümlenmiş diplomatik tekliflerin kısa geçmişi.
	DiplomaticOfferHistory []DiplomaticOfferHistoryEntry `json:"diplomatic_offer_history,omitempty"`
	// Turn içinde devlet başına gönderilen diplomasi teklif sayacı.
	DiplomacyOfferCounts map[faction.FactionID]int `json:"diplomacy_offer_counts,omitempty"`
	// Aktör-hedef-aksiyon bazında son reddedilen teklif turu.
	OfferRejectionTurns map[string]int `json:"diplomatic_offer_last_rejected_turns,omitempty"`

	// Ticaret güzergahları
	TradeRoutes  []*economy.TradeRoute          `json:"trade_routes"`
	TradeCenters world.TradeCenterConfig        `json:"trade_centers,omitempty"` // senaryo bazlı tarihsel ticaret merkezleri + link graph
	Sieges       map[world.RegionID]*SiegeState `json:"sieges,omitempty"`
	// Bu tur uygulanacak yağmalar. Ekonomi tick'inde hedef üretiminden düşülüp
	// yağmalayan fraksiyona aktarılır; aynı bölge aynı turda yalnız bir kez
	// yağmalanabilir.
	Raids map[world.RegionID]*RaidState `json:"raids,omitempty"`

	// Dinamik piyasa fiyatları (her tur sonu güncellenir)
	MarketPrices economy.CurrentMarketPrice `json:"-"`

	// Tur çözümlemesinde MovePoints sıfırlanmadan önce yakalanan hareket bilgisi.
	// Kalıcı değildir; yüklenen oyunda bir sonraki çözümleme başında yeniden üretilir.
	ArmyMoveUsage map[army.ArmyID]bool `json:"-"`

	// Bu tur oyuncu tarafından tahıl yardımı uygulanmış bölgeler; kalıcı değildir.
	GrainAidUsage map[world.RegionID]bool `json:"-"`

	// Devam eden üretimler
	ProductionQueue   []ProductionOrder `json:"production_queue"`
	NextProductionSeq int               `json:"next_production_seq"`

	// Sıradaki ordu ID üretmek için sayaç
	NextArmySeq      int `json:"next_army_seq"`
	NextCommanderSeq int `json:"next_commander_seq,omitempty"`

	// Oyun aşaması
	Phase Phase `json:"phase"`

	// Kazanan (boş = oyun devam ediyor)
	WinnerID faction.FactionID `json:"winner_id"`

	// Region paint overrides - edit modunda bölge boyama değişiklikleri (piksel indeksi -> bölge ID)
	RegionPaintOverrides map[int]world.RegionID `json:"region_paint_overrides,omitempty"`

	// Aktif bölge event ikonları (haritada birkaç tur görünür kalır)
	ActiveRegionEvents []RegionEventStatus `json:"active_region_events,omitempty"`

	// Geçici açık deniz temas kararı; temas çözülünce temizlenir ve save'e yazılmaz.
	PendingNavalContact *NavalContact `json:"-"`
	// Geçici kara temas kararı; temas çözülünce temizlenir ve save'e yazılmaz.
	PendingLandContact *LandContact `json:"-"`
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

// CivilianGrainDemand bir bölgenin tur başı sivil tahıl ihtiyacını döner.
// Population bölgenin kırsal ve yerleşim nüfuslarının toplamıdır; 18 nüfus bir
// tahıl birimi tüketir. Nüfusu olmayan legacy/test bölgeleri tüketim oluşturmaz.
func CivilianGrainDemand(region *world.Region) int {
	if region == nil || region.Population <= 0 {
		return 0
	}

	demand := (region.Population + civilianGrainPopulationUnit - 1) / civilianGrainPopulationUnit
	if demand < 1 {
		return 1
	}
	return demand
}

// CivilianGrainDemandForRegion aktif olayların geçici tüketim etkisini de
// uygulayarak bir bölgenin tur başı sivil tahıl ihtiyacını döner.
func (s *GameState) CivilianGrainDemandForRegion(region *world.Region) int {
	base := CivilianGrainDemand(region)
	if s == nil || region == nil || base <= 0 {
		return base
	}

	percent := 100 + s.RegionGrainDemandModifier(region.ID)
	if percent < 0 {
		percent = 0
	}
	demand := base * percent / 100
	if demand < 0 {
		return 0
	}
	return demand
}

// GrainStorageCapacity toplam sivil/ordu talebinden tahıl ambar kapasitesini
// hesaplar. Ekonomi tick'i ve HUD başlangıç görünümü aynı kapasite kuralını
// kullanır.
func GrainStorageCapacity(civilianDemand, armyUpkeep, storageBonus int) int {
	if civilianDemand < 0 {
		civilianDemand = 0
	}
	if armyUpkeep < 0 {
		armyUpkeep = 0
	}
	if storageBonus < 0 {
		storageBonus = 0
	}
	if civilianDemand+armyUpkeep <= 0 {
		return storageBonus
	}

	capacity := civilianDemand*grainCivilianStorageMonths + armyUpkeep*grainArmyStorageMonths + storageBonus
	if capacity < grainMinimumStorageCapacity {
		return grainMinimumStorageCapacity
	}
	return capacity
}

// GrainStorageCapacityForFaction ekonomi tick'i henüz çalışmadığında da HUD
// için fraksiyonun güncel tahıl ambar kapasitesini hesaplar.
func (s *GameState) GrainStorageCapacityForFaction(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	civilianDemand := 0
	storageBonus := 0
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || s.SiegeAt(region.ID) != nil {
			continue
		}
		civilianDemand += s.CivilianGrainDemandForRegion(region)
		for _, buildingID := range region.Buildings {
			if building := s.BuildingTypes[buildingID]; building != nil {
				storageBonus += building.StorageCapacity
			}
		}
	}
	armyUpkeep := 0
	for _, currentArmy := range s.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			armyUpkeep += s.EffectiveArmyGrainUpkeep(currentArmy)
		}
	}
	return GrainStorageCapacity(civilianDemand, armyUpkeep, storageBonus)
}

// RegionGrainProductionModifier aktif bölge olaylarının toplam üretim
// çarpanını yüzde puan olarak döner. Süresi bitmiş kayıtlar etkisizdir.
func (s *GameState) RegionGrainProductionModifier(regionID world.RegionID) int {
	if s == nil || regionID == "" {
		return 0
	}
	modifier := 0
	for _, event := range s.ActiveRegionEvents {
		if event.RegionID == regionID && event.TurnsLeft > 0 {
			modifier += event.GrainProductionPercent
		}
	}
	if modifier < -100 {
		return -100
	}
	if modifier > 200 {
		return 200
	}
	return modifier
}

// RegionGrainDemandModifier aktif bölge olaylarının toplam sivil tüketim
// çarpanını yüzde puan olarak döner. Süresi bitmiş kayıtlar etkisizdir.
func (s *GameState) RegionGrainDemandModifier(regionID world.RegionID) int {
	if s == nil || regionID == "" {
		return 0
	}
	modifier := 0
	for _, event := range s.ActiveRegionEvents {
		if event.RegionID == regionID && event.TurnsLeft > 0 {
			modifier += event.GrainDemandPercent
		}
	}
	if modifier < -100 {
		return -100
	}
	if modifier > 200 {
		return 200
	}
	return modifier
}

// RegionMilitaryGrainProduction, sivil talep karşılandıktan sonra aynı bölgede
// kara ordusu ikmaline kalabilecek efektif tahıl üretimini döner. Oyun ve AI
// lojistik hesapları bu ortak seam'i kullanır.
func (s *GameState) RegionMilitaryGrainProduction(region *world.Region) int {
	if s == nil || region == nil {
		return 0
	}
	production := s.RegionProductionSummary(region).Grain
	production -= s.CivilianGrainDemandForRegion(region)
	if production < 0 {
		return 0
	}
	return production
}

type RegionLogisticsStatus struct {
	RegionID                 world.RegionID
	OwnerID                  string
	LocalProduction          int
	SettlementBuffer         int
	GranarySupport           int
	ReserveSupport           int
	BlockadePercent          int
	Demand                   int
	Capacity                 int
	Overload                 int
	ArmyCount                int
	UnitsAffected            int
	UnitsLost                int
	TotalHPDamage            int
	PeakOverloadTurns        int
	FriendlySupplyArmies     int
	FriendlySupplyGrainSpent int
}

type ArmyLogisticsStatus struct {
	ArmyID                   army.ArmyID
	RegionID                 world.RegionID
	OwnerID                  string
	Demand                   int
	Capacity                 int
	Overload                 int
	OverCapacityTurns        int
	DamagePerUnit            int
	UnitsAffected            int
	UnitsLost                int
	TotalHPDamage            int
	FriendlySupplyFactionID  faction.FactionID
	FriendlySupplyRegionID   world.RegionID
	FriendlySupplyGrainSpent int
	FriendlySupplySameRealm  bool
}

// FriendlySupplySupport bir dost devletin cephedeki orduya yaptığı ücretli
// ileri ikmal katkısını taşır. Runtime-only durumdur; her ekonomi tick'inde
// yeniden belirlenir.
type FriendlySupplySupport struct {
	ArmyID            army.ArmyID
	ProviderFactionID faction.FactionID
	ProviderRegionID  world.RegionID
	GrainSpent        int
	SameRealm         bool
}

// GrainSupplyLevel fraksiyonun mevcut tahıl rezervinin şiddetini bildirir.
// Runtime görünürlüğü için kullanılır; save'e doğrudan yazılmaz.
type GrainSupplyLevel int

const (
	GrainSupplyStable GrainSupplyLevel = iota
	GrainSupplyWarning
	GrainSupplyCritical
	GrainSupplyFamine
)

// GrainEconomyStatus ekonomi tick'inin tahıl üretim/tüketim sonucunu taşır.
// Stockpile ve MonthsOfSupply tick sonrasındaki gerçek rezervi temsil eder.
type GrainEconomyStatus struct {
	FactionID                faction.FactionID
	Production               int
	CivilianDemand           int
	ArmyUpkeep               int
	StrategicDemand          int
	ReplenishmentHP          int
	ReplenishmentGrainSpent  int
	PopulationGrowth         int
	GrowthGrainSpent         int
	FriendlySupplyGrainSpent int
	// ArmyMoraleDelta bu ekonomi tick'inde fraksiyon ordularında gerçekleşen
	// toplam moral değişimidir; negatif değer ikmal kaynaklı kaybı gösterir.
	ArmyMoraleDelta int
	AutoExportSold  int
	AutoExportGold  int
	TotalDemand     int
	NetChange       int
	Stockpile       int
	StorageCapacity int
	Spoiled         int
	MonthsOfSupply  int
	Shortage        int
	SupplyLevel     GrainSupplyLevel
}

const (
	strategicGrainReserveMonths      = 3
	strategicGrainReserveCapacityMin = 100
)

// StrategicGrainDemandFromStockpile üç aylık rezerv hedefinin mevcut stok
// tarafından karşılanmayan kısmını döner.
func StrategicGrainDemandFromStockpile(stockpile, totalDemand int) int {
	if stockpile < 0 {
		stockpile = 0
	}
	if totalDemand <= 0 {
		return 0
	}
	target := totalDemand * strategicGrainReserveMonths
	if stockpile >= target {
		return 0
	}
	return target - stockpile
}

// StrategicGrainDemand mevcut talep ve stoktan türeyen, fraksiyonun üç aylık
// güvenli rezerv hedefine ulaşmak için ithal etmesi gereken tahıl miktarıdır.
// Runtime ekonomi snapshot'ı yoksa talep bölge/ordu state'inden yeniden hesaplanır.
func (s *GameState) StrategicGrainDemand(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	f := s.Factions[fid]
	if f == nil || f.IsEliminated {
		return 0
	}
	totalDemand := 0
	if status, ok := s.GrainEconomy[fid]; ok {
		totalDemand = status.TotalDemand
	} else {
		for _, region := range s.Regions {
			if region != nil && !region.IsSea && region.OwnerID == string(fid) {
				totalDemand += s.CivilianGrainDemandForRegion(region)
			}
		}
		for _, a := range s.Armies {
			if a != nil && a.OwnerID == string(fid) {
				totalDemand += s.EffectiveArmyGrainUpkeep(a)
			}
		}
	}
	if totalDemand <= 0 {
		return 0
	}
	return StrategicGrainDemandFromStockpile(f.Grain, totalDemand)
}

// StrategicGrainSurplus kapasite üstündeki, ticaretle güvenle ihraç edilebilecek
// tahıl miktarını döner. Runtime snapshot yoksa minimum 100'lük legacy rezervi kullanır.
func (s *GameState) StrategicGrainSurplus(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	f := s.Factions[fid]
	if f == nil || f.IsEliminated {
		return 0
	}
	capacity := strategicGrainReserveCapacityMin
	if status, ok := s.GrainEconomy[fid]; ok && status.StorageCapacity > 0 {
		capacity = status.StorageCapacity
	}
	if f.Grain <= capacity {
		return 0
	}
	return f.Grain - capacity
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

// CalendarMonthsPerTurn bir stratejik turun temsil ettiği takvim ayı sayısını
// döner. Eski save/test state'leri alanı taşımadığında bir aylık davranış korunur.
func (s *GameState) CalendarMonthsPerTurn() int {
	if s != nil && s.MonthsPerTurn > 0 && s.MonthsPerTurn <= 12 {
		return s.MonthsPerTurn
	}
	return 1
}

// CurrentTurnEndDate aktif turun kapsadığı son takvim ayını döner.
func (s *GameState) CurrentTurnEndDate() (year, month int) {
	if s == nil {
		return 0, 0
	}
	year, month = s.Year, s.Month
	if month < 1 || month > 12 {
		month = 1
	}
	for remaining := s.CalendarMonthsPerTurn() - 1; remaining > 0; remaining-- {
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	return year, month
}

// HistoricalDateOccursThisTurn tarihsel yıl/ayın aktif stratejik turun kapsadığı
// takvim aralığında olup olmadığını bildirir. month=0 aynı yılın herhangi bir
// ayını temsil eder.
func (s *GameState) HistoricalDateOccursThisTurn(year, month int) bool {
	if s == nil || year <= 0 || s.Year > year {
		return false
	}
	endYear, endMonth := s.CurrentTurnEndDate()
	if year > endYear {
		return false
	}
	if month <= 0 {
		return year >= s.Year && year <= endYear
	}
	if month > 12 {
		return false
	}
	startMonth := s.Month
	if startMonth < 1 || startMonth > 12 {
		startMonth = 1
	}
	startAbs := s.Year*12 + startMonth - 1
	endAbs := endYear*12 + endMonth - 1
	targetAbs := year*12 + month - 1
	return targetAbs >= startAbs && targetAbs <= endAbs
}

// CurrentTurnIncludesMonth aktif turun takvim aralığının verilen ayı içerip
// içermediğini döner. Yıllık etkiler, tur uzunluğundan bağımsız olarak bununla
// yalnız bir kez uygulanır.
func (s *GameState) CurrentTurnIncludesMonth(month int) bool {
	if s == nil || month < 1 || month > 12 {
		return false
	}
	currentMonth := s.Month
	if currentMonth < 1 || currentMonth > 12 {
		currentMonth = 1
	}
	for remaining := s.CalendarMonthsPerTurn(); remaining > 0; remaining-- {
		if currentMonth == month {
			return true
		}
		currentMonth++
		if currentMonth > 12 {
			currentMonth = 1
		}
	}
	return false
}

// AdvanceTurn turu bir ileri alır, senaryonun takvim ayı hızına göre ay/yıl günceller.
func (s *GameState) AdvanceTurn() {
	s.Turn++
	for remaining := s.CalendarMonthsPerTurn(); remaining > 0; remaining-- {
		s.Month++
		if s.Month > 12 {
			s.Month = 1
			s.Year++
		}
	}
	s.RetireExpiredCommanders()
	s.ResetDiplomacyOfferCounts()
	s.GrainAidUsage = nil
	s.GrainSaleGoldUsed = nil
}

// GrainAidBlockReason tahıl yardımının neden uygulanamayacağını döner.
func (s *GameState) GrainAidBlockReason(rid world.RegionID) string {
	if s == nil || rid == "" {
		return "Geçersiz yardım bölgesi."
	}
	region := s.Regions[rid]
	if region == nil || region.IsSea {
		return "Deniz veya bilinmeyen bölgeye tahıl yardımı yapılamaz."
	}
	if region.OwnerID != string(s.PlayerFactionID) {
		return "Tahıl yardımı yalnız kendi bölgelerine yapılabilir."
	}
	if region.IsLocked {
		return "Kilitli bölgeye tahıl yardımı yapılamaz."
	}
	if s.SiegeAt(rid) != nil {
		return "Kuşatma altındaki bölgeye tahıl yardımı ulaştırılamaz."
	}
	if s.GrainAidUsage != nil && s.GrainAidUsage[rid] {
		return "Bu bölgeye bu tur zaten tahıl yardımı yapıldı."
	}
	if region.Satisfaction >= 90 {
		return "Bu bölgede tahıl yardımına ihtiyaç yok."
	}
	f := s.Factions[s.PlayerFactionID]
	if f == nil || f.Grain < GrainAidCost {
		return "Tahıl yardımı için 12 tahıl gerekiyor."
	}
	return ""
}

// CanApplyGrainAid UI ve input katmanının ortak yardım uygunluk kontrolüdür.
func (s *GameState) CanApplyGrainAid(rid world.RegionID) bool {
	return s.GrainAidBlockReason(rid) == ""
}

// ApplyGrainAid kendi bölgesinin tahıl karşılığında memnuniyetini artırır.
// Yardım state üzerinde tek bir kanonik mutasyon noktasıdır.
func (s *GameState) ApplyGrainAid(rid world.RegionID) bool {
	if !s.CanApplyGrainAid(rid) {
		return false
	}
	region := s.Regions[rid]
	f := s.Factions[s.PlayerFactionID]
	f.Grain -= GrainAidCost
	region.Satisfaction += GrainAidSatisfactionGain
	if region.Satisfaction > 100 {
		region.Satisfaction = 100
	}
	if s.GrainAidUsage == nil {
		s.GrainAidUsage = make(map[world.RegionID]bool)
	}
	s.GrainAidUsage[rid] = true
	return true
}

// TaxIncomeForFaction kuşatma altındaki bölgeleri dışarıda bırakarak fraksiyonun
// mevcut temel vergi gelirini döner. Ticaret, teknoloji ve mevsim bonusları bu
// limite dahil değildir; tahıl satışı doğrudan vergi gelirinin yerine geçmez.
func (s *GameState) TaxIncomeForFaction(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	total := 0
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || s.SiegeAt(region.ID) != nil {
			continue
		}
		total += scaleBlockadeOutput(region.GoldIncome(), s.RegionBlockadeOutputRetentionPercent(region))
	}
	return total
}

// GrainSaleGoldBudget bu turda acil/otomatik tahıl satışında kullanılabilecek
// kalan altın bütçesini döner. Bütçe temel vergi gelirinin %100'ü ile sınırlıdır.
func (s *GameState) GrainSaleGoldBudget(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	cap := s.TaxIncomeForFaction(fid) * grainSaleGoldCapPercentOfTaxIncome / 100
	used := s.GrainSaleGoldUsed[fid]
	if cap <= used {
		return 0
	}
	return cap - used
}

// RecordGrainSaleGold satışın bu turdaki vergi gelirine bağlı bütçesini tüketir.
func (s *GameState) RecordGrainSaleGold(fid faction.FactionID, gold int) {
	if s == nil || fid == "" || gold <= 0 {
		return
	}
	budget := s.GrainSaleGoldBudget(fid)
	if gold > budget {
		gold = budget
	}
	if gold <= 0 {
		return
	}
	if s.GrainSaleGoldUsed == nil {
		s.GrainSaleGoldUsed = make(map[faction.FactionID]int)
	}
	s.GrainSaleGoldUsed[fid] += gold
}

// grainExcessStock depolama kapasitesinin üzerindeki tahıl miktarını döner.
// GrainEconomy henüz oluşmadıysa küçük devletler için temel 100 rezervi korunur.
func (s *GameState) grainExcessStock(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}
	f := s.Factions[fid]
	if f == nil || f.Grain <= 0 {
		return 0
	}
	capacity := s.GrainEconomy[fid].StorageCapacity
	if capacity <= 0 {
		capacity = 100
	}
	limit := f.Grain - capacity
	if limit < 0 {
		return 0
	}
	return limit
}

// EmergencyGrainSaleLimit depolama kapasitesinin üzerindeki tahıldan bu tur
// satılabilecek miktarı döner. Miktar ayrıca vergi gelirine bağlı altın bütçesi
// ile sınırlandırılır; böylece satış tek başına vergi gelirinin üstüne çıkamaz.
func (s *GameState) EmergencyGrainSaleLimit() int {
	if s == nil || s.PlayerFactionID == "" {
		return 0
	}
	price := s.EmergencyGrainSaleUnitPrice()
	if price <= 0 {
		return 0
	}
	limit := s.grainExcessStock(s.PlayerFactionID)
	byBudget := s.GrainSaleGoldBudget(s.PlayerFactionID) / price
	if byBudget < limit {
		limit = byBudget
	}
	return limit
}

// EmergencyGrainSaleUnitPrice acil tahıl satışının güncel birim fiyatını döner.
func (s *GameState) EmergencyGrainSaleUnitPrice() int {
	if s == nil {
		return 0
	}
	price := s.MarketPrices[economy.GoodGrain]
	if price <= 0 {
		price = economy.BaseGoldValue[economy.GoodGrain]
	}
	return economy.EmergencySaleUnitPrice(price)
}

// ApplyEmergencyGrainSale fazla tahılı doğrudan pazara satar; rezerv kapasitesini korur.
func (s *GameState) ApplyEmergencyGrainSale(amount int) (sold, gold int) {
	if s == nil || amount <= 0 {
		return 0, 0
	}
	f := s.Factions[s.PlayerFactionID]
	if f == nil {
		return 0, 0
	}
	sold = amount
	if limit := s.EmergencyGrainSaleLimit(); sold > limit {
		sold = limit
	}
	price := s.EmergencyGrainSaleUnitPrice()
	if sold <= 0 || price <= 0 {
		return 0, 0
	}
	f.Grain -= sold
	gold = sold * price
	f.Gold += gold
	s.RecordGrainSaleGold(s.PlayerFactionID, gold)
	return sold, gold
}

// ApplyAutomaticGrainExport aktif ticaret ağı partnerlerine kapasite üstü tahılı
// düşük fiyatla satar. Partner sırası faction ID ile deterministiktir.
func (s *GameState) ApplyAutomaticGrainExport() (sold, gold int) {
	if s == nil || !s.AutoGrainExport || s.PlayerFactionID == "" {
		return 0, 0
	}
	limit := s.grainExcessStock(s.PlayerFactionID)
	price := economy.AutomaticExportUnitPrice(s.MarketPrices[economy.GoodGrain])
	if price <= 0 {
		price = economy.AutomaticExportUnitPrice(economy.BaseGoldValue[economy.GoodGrain])
	}
	if price <= 0 || limit <= 0 {
		return 0, 0
	}
	byBudget := s.GrainSaleGoldBudget(s.PlayerFactionID) / price
	if byBudget <= 0 {
		return 0, 0
	}
	if byBudget < limit {
		limit = byBudget
	}

	partnersSet := make(map[faction.FactionID]struct{})
	for _, route := range s.TradeRoutes {
		if route == nil || route.SuspendedTurns > 0 || route.AmountPerTurn <= 0 {
			continue
		}
		var partner faction.FactionID
		switch {
		case faction.FactionID(route.FromFactionID) == s.PlayerFactionID:
			partner = faction.FactionID(route.ToFactionID)
		case faction.FactionID(route.ToFactionID) == s.PlayerFactionID:
			partner = faction.FactionID(route.FromFactionID)
		default:
			continue
		}
		if partner == "" || partner == s.PlayerFactionID {
			continue
		}
		f := s.Factions[partner]
		if f == nil || f.IsEliminated {
			continue
		}
		if relation := s.Relations[faction.RelationKey(s.PlayerFactionID, partner)]; relation != nil && relation.Stance == faction.StanceWar {
			continue
		}
		partnersSet[partner] = struct{}{}
	}

	partners := make([]faction.FactionID, 0, len(partnersSet))
	for partner := range partnersSet {
		partners = append(partners, partner)
	}
	sort.Slice(partners, func(i, j int) bool { return partners[i] < partners[j] })

	remaining := limit
	for _, partner := range partners {
		if remaining <= 0 {
			break
		}
		buyer := s.Factions[partner]
		amount := buyer.Gold / price
		if amount > remaining {
			amount = remaining
		}
		if amount <= 0 {
			continue
		}
		if !economy.TransferGoodsAtUnitPrice(s.Factions, s.PlayerFactionID, partner, economy.GoodGrain, amount, price) {
			continue
		}
		sold += amount
		gold += amount * price
		s.RecordGrainSaleGold(s.PlayerFactionID, amount*price)
		remaining -= amount
	}
	return sold, gold
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

// DiplomaticOfferRejectionKey aynı aktörün aynı hedefe aynı aksiyonu tekrar
// denemesini izlemek için yönlü anahtar üretir.
func DiplomaticOfferRejectionKey(from, to, action string) string {
	return from + "|" + to + "|" + action
}

// DiplomaticOfferRegionRejectionKey kuşatma gibi aynı aktör-hedef-aksiyon
// altında birden fazla bölgeye ait teklifleri birbirinden ayırır.
func DiplomaticOfferRegionRejectionKey(from, to, action string, regionID world.RegionID) string {
	return DiplomaticOfferRejectionKey(from, to, action) + "|region=" + string(regionID)
}

// MarkDiplomaticOfferRejected son reddedilen teklif turunu kaydeder.
func (s *GameState) MarkDiplomaticOfferRejected(from, to, action string) {
	if s == nil || from == "" || to == "" || action == "" {
		return
	}
	if s.OfferRejectionTurns == nil {
		s.OfferRejectionTurns = make(map[string]int, 4)
	}
	s.OfferRejectionTurns[DiplomaticOfferRejectionKey(from, to, action)] = s.Turn
}

// MarkDiplomaticOfferRejectedForRegion yalnız belirtilen kuşatma bölgesindeki
// teklifin tekrarını bekletir; diğer bölgelerdeki teklifler etkilenmez.
func (s *GameState) MarkDiplomaticOfferRejectedForRegion(from, to, action string, regionID world.RegionID) {
	if s == nil || from == "" || to == "" || action == "" || regionID == "" {
		return
	}
	if s.OfferRejectionTurns == nil {
		s.OfferRejectionTurns = make(map[string]int, 4)
	}
	s.OfferRejectionTurns[DiplomaticOfferRegionRejectionKey(from, to, action, regionID)] = s.Turn
}

// DiplomaticOfferRegionRetryBlocked aynı kuşatma bölgesine ait teklifin
// bölgesel ret cooldown'u içindeyse true döner.
func (s *GameState) DiplomaticOfferRegionRetryBlocked(from, to, action string, regionID world.RegionID, cooldownTurns int) bool {
	if s == nil || from == "" || to == "" || action == "" || regionID == "" || cooldownTurns <= 0 || len(s.OfferRejectionTurns) == 0 {
		return false
	}
	lastRejected, ok := s.OfferRejectionTurns[DiplomaticOfferRegionRejectionKey(from, to, action, regionID)]
	return ok && s.Turn-lastRejected < cooldownTurns
}

// DiplomaticOfferRetryBlocked reddedilen teklifin bekleme süresi dolmadıysa
// true döner. Kayıtlı ret yoksa teklif engellenmez.
func (s *GameState) DiplomaticOfferRetryBlocked(from, to, action string, cooldownTurns int) bool {
	if s == nil || cooldownTurns <= 0 || len(s.OfferRejectionTurns) == 0 {
		return false
	}
	lastRejected, ok := s.OfferRejectionTurns[DiplomaticOfferRejectionKey(from, to, action)]
	return ok && s.Turn-lastRejected < cooldownTurns
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
		// Pusu ordusu normal keşif ve hedef seçiminde görünmezdir. Bölgeye
		// giren hareketli ordu için özel SelectAmbushDefender çağrısı bunu
		// temas tetikleyicisi olarak ayrıca bulur.
		if candidate.InAmbush {
			continue
		}
		if navalSeaMove && candidate.IsDocked() {
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

// SelectAmbushDefender, hedef bölgedeki gizli pusu ordusunu deterministik
// olarak seçer. Bu helper yalnız hedefe giriş anındaki temas kontrolünde
// kullanılmalıdır; normal düşman görüşü pusu ordusunu görmez.
func (s *GameState) SelectAmbushDefender(attacker *army.Army, target world.RegionID, navalSeaMove bool) *army.Army {
	if s == nil || attacker == nil {
		return nil
	}
	var best *army.Army
	bestPower := -1
	for _, candidate := range s.Armies {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID || !candidate.InAmbush {
			continue
		}
		if navalSeaMove || candidate.IsNaval {
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

// ArmyHiddenFrom reports whether an army's pusu stance hides it from the
// opposing faction. The owning faction always retains visibility.
func (s *GameState) ArmyHiddenFrom(candidate *army.Army, observer faction.FactionID) bool {
	return candidate != nil && candidate.InAmbush && candidate.OwnerID != string(observer)
}

// SelectNavalAutoEngagementDefender, yalnız devriye-abluka görevi çifti
// otomatik karşılaşma oluşturduğunda savaşacak düşman filoyu seçer. Görevsiz
// filolar aynı denizde bulunabilir ancak bu seçimden dışarıda kalır.
func (s *GameState) SelectNavalAutoEngagementDefender(attacker *army.Army, target world.RegionID) *army.Army {
	if s == nil || attacker == nil || !attacker.IsAtSea() {
		return nil
	}
	var best *army.Army
	bestPower := -1
	for _, candidate := range s.Armies {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID || !candidate.IsAtSea() {
			continue
		}
		key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(candidate.OwnerID))
		rel, exists := s.Relations[key]
		if !exists || rel == nil || rel.Stance != faction.StanceWar || !s.NavalFleetsAutoEngageAtSea(attacker, candidate, target) {
			continue
		}
		power := 0
		if s.UnitTypes != nil {
			power = candidate.TotalStrength(s.UnitTypes)
		}
		if best == nil || power > bestPower || power == bestPower && candidate.ID < best.ID {
			best = candidate
			bestPower = power
		}
	}
	return best
}

// SelectNavalPatrolDefender, AI devriyesinin hedef denizdeki açık abluka
// filosunu yakalayıp yakalayamayacağını kontrol eder. AI devriyesi kalıcı
// oyuncu görevi taşımadığı için burada saldıran filonun devriye niyeti geçici
// bir kopya görevle ifade edilir; gerçek filo state'i değiştirilmez.
func (s *GameState) SelectNavalPatrolDefender(attacker *army.Army, target world.RegionID) *army.Army {
	if s == nil || attacker == nil {
		return nil
	}
	patrol := *attacker
	patrolMission := army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: target}
	patrol.NavalMission = &patrolMission
	return s.SelectNavalAutoEngagementDefender(&patrol, target)
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
	candidates := make([]*army.Army, 0, len(s.Armies))
	for _, candidate := range s.Armies {
		if candidate != nil {
			candidates = append(candidates, candidate)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	for _, candidate := range candidates {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID {
			continue
		}
		if navalSeaMove && candidate.IsDocked() {
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

// CollectNavalAutoEngagementDefenders, devriye ile yakalanan düşman abluka
// filosunun aynı görev çatışmasına dahil olan filolarını toplar. Görevsiz
// filolar bu otomatik savaşta savunma hattına eklenmez.
func (s *GameState) CollectNavalAutoEngagementDefenders(attacker *army.Army, target world.RegionID) (combined *army.Army, sourceIDs []army.ArmyID) {
	if s == nil || attacker == nil || !attacker.IsAtSea() {
		return nil, nil
	}
	candidates := make([]*army.Army, 0, len(s.Armies))
	for _, candidate := range s.Armies {
		if candidate != nil {
			candidates = append(candidates, candidate)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	var units []army.Unit
	for _, candidate := range candidates {
		if candidate == nil || candidate.RegionID != target || candidate.OwnerID == attacker.OwnerID || !candidate.IsAtSea() {
			continue
		}
		key := faction.RelationKey(faction.FactionID(attacker.OwnerID), faction.FactionID(candidate.OwnerID))
		rel, exists := s.Relations[key]
		if !exists || rel == nil || rel.Stance != faction.StanceWar || !s.NavalFleetsAutoEngageAtSea(attacker, candidate, target) {
			continue
		}
		units = append(units, candidate.Units...)
		sourceIDs = append(sourceIDs, candidate.ID)
	}
	if len(units) == 0 {
		return nil, nil
	}
	if len(units) > army.MaxArmySize {
		units = units[:army.MaxArmySize]
	}
	return &army.Army{OwnerID: attacker.OwnerID, Units: units}, sourceIDs
}

// CollectNavalPatrolDefenders, AI devriyesinin yalnız abluka görevi taşıyan
// düşman filolarını yakaladığı otomatik deniz savaşını kurar.
func (s *GameState) CollectNavalPatrolDefenders(attacker *army.Army, target world.RegionID) (*army.Army, []army.ArmyID) {
	if s == nil || attacker == nil {
		return nil, nil
	}
	patrol := *attacker
	patrolMission := army.NavalMission{Kind: army.NavalMissionPatrol, TargetRegionID: target}
	patrol.NavalMission = &patrolMission
	return s.CollectNavalAutoEngagementDefenders(&patrol, target)
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
			s.RemoveArmy(id)
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

// EffectiveArmyGrainUpkeep ordunun bu turdaki temel tahıl bakım ihtiyacını
// hareket ve kuşatma yüküyle birlikte hesaplar. Bu değer fraksiyon ekonomisi
// ve toplam stok güvenliği için kullanılır.
func (s *GameState) EffectiveArmyGrainUpkeep(a *army.Army) int {
	return s.armyGrainUpkeep(a, false, false)
}

// RegionalArmyGrainDemand bölgesel ikmal baskısında kullanılan tahıl talebini
// döner. Başkentten uzak/kopuk hatların ve sınır ikmali alan kuşatmaların
// etkisi yalnız yerel yıpranmaya girer; fraksiyonun toplam tahıl giderini
// yapay olarak şişirmez.
func (s *GameState) RegionalArmyGrainDemand(a *army.Army) int {
	return s.armyGrainUpkeep(a, true, s.ExternalFriendlySupplyAvailable(a))
}

// RegionalArmyGrainDemandWithExternalSupply gerçek tur çözümünde kullanılır.
// Müttefik/vassal ikmali yalnız destekçi tahılı gerçekten ödediyse etkindir.
func (s *GameState) RegionalArmyGrainDemandWithExternalSupply(a *army.Army, externalSupplyActive bool) int {
	return s.armyGrainUpkeep(a, true, externalSupplyActive)
}

func (s *GameState) armyGrainUpkeep(a *army.Army, includeRegionalSupply, externalSupplyActive bool) int {
	if s == nil || a == nil {
		return 0
	}
	base := a.TotalGrainUpkeep(s.UnitTypes)
	if base <= 0 {
		return 0
	}

	percent := grainUpkeepStationaryPercent
	if siege := s.SiegeByArmy(a.ID); siege != nil {
		percent = grainUpkeepSiegeAttacker
		if includeRegionalSupply && (s.HasOwnedLandSupplyBorder(siege.RegionID, a.OwnerID) || (externalSupplyActive && s.hasExternalFriendlyLandSupplyBorder(siege.RegionID, a.OwnerID))) {
			percent = grainUpkeepSuppliedSiegeAttacker
		}
	} else {
		for _, siege := range s.Sieges {
			if siege == nil {
				continue
			}
			if siege.DefenderArmyID == a.ID || s.IsArmyDefendingSiegedRegion(a) {
				percent = grainUpkeepSiegeDefender
				break
			}
		}
		if percent == grainUpkeepStationaryPercent && a.IsGarrison {
			percent = grainUpkeepGarrisonPercent
		}
	}

	if percent == grainUpkeepStationaryPercent {
		moved := false
		if s.ArmyMoveUsage != nil {
			moved = s.ArmyMoveUsage[a.ID]
		} else if a.MaxMovePoints > 0 {
			moved = a.MovePoints >= 0 && a.MovePoints < a.MaxMovePoints
		}
		if moved {
			percent = grainUpkeepMovingPercent
		}
	}
	if includeRegionalSupply {
		percent += s.capitalSupplyPenaltyPercent(a, externalSupplyActive)
	}

	upkeep := base * percent / 100
	if upkeep < 1 {
		return 1
	}
	return upkeep
}

// HasOwnedLandSupplyBorder hedef kara bölgesinin, belirtilen devlete ait bir
// kara bölgesiyle doğrudan sınırı olup olmadığını döner. Kuşatmada bu sınır,
// ordunun ülkesinden düzenli kara ikmali aldığı anlamına gelir.
func (s *GameState) HasOwnedLandSupplyBorder(regionID world.RegionID, ownerID string) bool {
	if s == nil || regionID == "" || ownerID == "" {
		return false
	}
	region := s.Regions[regionID]
	if region == nil || region.IsSea {
		return false
	}
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor != nil && !neighbor.IsSea && neighbor.OwnerID == ownerID {
			return true
		}
	}
	return false
}

// HasFriendlyLandSupplyBorder hedef kara bölgesinin belirtilen devlete ait,
// aynı realm içindeki ya da ittifaklı bir kara bölgesiyle sınırı olup olmadığını
// döner. Böyle bir bölge sahadaki ordu için ileri ikmal noktasıdır.
func (s *GameState) HasFriendlyLandSupplyBorder(regionID world.RegionID, ownerID string) bool {
	if s == nil || regionID == "" || ownerID == "" {
		return false
	}
	region := s.Regions[regionID]
	if region == nil || region.IsSea {
		return false
	}
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor != nil && !neighbor.IsSea && s.canFactionReplenishIn(ownerID, neighbor.OwnerID) {
			return true
		}
	}
	return false
}

func (s *GameState) hasExternalFriendlyLandSupplyBorder(regionID world.RegionID, ownerID string) bool {
	if s == nil || regionID == "" || ownerID == "" {
		return false
	}
	region := s.Regions[regionID]
	if region == nil || region.IsSea {
		return false
	}
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || neighbor.OwnerID == ownerID {
			continue
		}
		if s.canFactionReplenishIn(ownerID, neighbor.OwnerID) {
			return true
		}
	}
	return false
}

// ExternalFriendlySupplyQuote komşu müttefik veya aynı realm vassal bölgesinin
// sağlayabileceği ileri ikmalin kaynağını ve tur başı tahıl bedelini döner.
// Aynı realm ikmali bağımsız müttefik ikmalinden daha verimlidir.
func (s *GameState) ExternalFriendlySupplyQuote(a *army.Army) (FriendlySupplySupport, bool) {
	if s == nil || a == nil || a.IsNaval || a.OwnerID == "" || a.RegionID == "" {
		return FriendlySupplySupport{}, false
	}
	region := s.Regions[a.RegionID]
	if region == nil || region.IsSea || region.OwnerID == a.OwnerID {
		return FriendlySupplySupport{}, false
	}
	best := FriendlySupplySupport{ArmyID: a.ID}
	found := false
	for _, neighborID := range region.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || neighbor.OwnerID == a.OwnerID || !s.canFactionReplenishIn(a.OwnerID, neighbor.OwnerID) {
			continue
		}
		sameRealm := stateSameRealm(s, faction.FactionID(a.OwnerID), faction.FactionID(neighbor.OwnerID))
		candidate := FriendlySupplySupport{
			ArmyID:            a.ID,
			ProviderFactionID: faction.FactionID(neighbor.OwnerID),
			ProviderRegionID:  neighbor.ID,
			SameRealm:         sameRealm,
		}
		if !found || (candidate.SameRealm && !best.SameRealm) || (candidate.SameRealm == best.SameRealm && candidate.ProviderRegionID < best.ProviderRegionID) {
			best = candidate
			found = true
		}
	}
	if !found {
		return FriendlySupplySupport{}, false
	}
	base := a.TotalGrainUpkeep(s.UnitTypes)
	if base <= 0 {
		return FriendlySupplySupport{}, false
	}
	divisor := 3 // bağımsız müttefik kendi konvoyu/harcaması için daha fazla öder.
	if best.SameRealm {
		divisor = 5
	}
	best.GrainSpent = (base + divisor - 1) / divisor
	if best.GrainSpent < 1 {
		best.GrainSpent = 1
	}
	return best, true
}

// ExternalFriendlySupplyAvailable destekçinin kendi asgari tahıl rezervini
// bozmadan bu tur bir ileri ikmal konvoyunu ödeyip ödeyemediğini döner. AI
// tahminleri ile gerçek turn çözümlemesi aynı uygunluk kuralını kullanır.
func (s *GameState) ExternalFriendlySupplyAvailable(a *army.Army) bool {
	supply, ok := s.ExternalFriendlySupplyQuote(a)
	if !ok || supply.GrainSpent <= 0 {
		return false
	}
	provider := s.Factions[supply.ProviderFactionID]
	if provider == nil || provider.IsEliminated {
		return false
	}
	reserve := 20
	if status, ok := s.GrainEconomy[supply.ProviderFactionID]; ok && status.TotalDemand > reserve {
		reserve = status.TotalDemand
	}
	return provider.Grain-supply.GrainSpent >= reserve
}

// ArmySupplyDistanceFromCapital sadece aynı devlete ait kara bölgeleri
// kullanarak başkentten orduya ulaşılan en kısa ikmal hattını döner. Düşman
// bölgesindeki ordu, kendi toprağına komşuysa bu son sınır geçişi de hesaba
// katılır. Geçerli başkent veya kara hattı yoksa ok false döner.
func (s *GameState) ArmySupplyDistanceFromCapital(a *army.Army) (distance int, ok bool) {
	if s == nil || a == nil || a.IsNaval || a.OwnerID == "" || a.RegionID == "" {
		return 0, false
	}
	capital, _, _, capitalOK := s.FactionCapital(faction.FactionID(a.OwnerID))
	current := s.Regions[a.RegionID]
	if !capitalOK || capital == nil || current == nil || current.IsSea {
		return 0, false
	}

	type supplyNode struct {
		regionID world.RegionID
		distance int
	}
	queue := []supplyNode{{regionID: capital.ID}}
	visited := map[world.RegionID]bool{capital.ID: true}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.regionID == current.ID {
			return node.distance, true
		}
		if current.OwnerID != a.OwnerID && regionsAreNeighbors(s.Regions[node.regionID], current.ID) {
			return node.distance + 1, true
		}
		region := s.Regions[node.regionID]
		if region == nil {
			continue
		}
		for _, neighborID := range region.Neighbors {
			neighbor := s.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.OwnerID != a.OwnerID || visited[neighborID] {
				continue
			}
			visited[neighborID] = true
			queue = append(queue, supplyNode{regionID: neighborID, distance: node.distance + 1})
		}
	}
	return 0, false
}

// CapitalSupplyPenaltyPercent uzak veya kopuk kara ikmalinin tahıl bakımına
// eklediği yüzdeyi döner. Başkente yakın iki bölgelik hat cezasızdır; daha
// uzun hatlar kademeli artar. Başkente kara bağlantısı olmayan ordular, ancak
// geçerli bir başkent varsa, en yüksek ikmal cezasını alır.
func (s *GameState) CapitalSupplyPenaltyPercent(a *army.Army) int {
	return s.capitalSupplyPenaltyPercent(a, true)
}

func (s *GameState) capitalSupplyPenaltyPercent(a *army.Army, externalSupplyActive bool) int {
	if s == nil || a == nil || a.IsNaval || a.OwnerID == "" {
		return 0
	}
	if _, _, _, capitalOK := s.FactionCapital(faction.FactionID(a.OwnerID)); !capitalOK {
		return 0
	}
	distance, connected := s.ArmySupplyDistanceFromCapital(a)
	if !connected {
		if externalSupplyActive && s.hasExternalFriendlyLandSupplyBorder(a.RegionID, a.OwnerID) {
			return capitalSupplyFriendlyBorderTax
		}
		return capitalSupplyDisconnectedTax
	}
	if distance <= capitalSupplyGraceDistance {
		return 0
	}
	penalty := ((distance - capitalSupplyGraceDistance + capitalSupplyDistanceStep - 1) / capitalSupplyDistanceStep) * capitalSupplyPenaltyPerStep
	if penalty > capitalSupplyMaxPenalty {
		return capitalSupplyMaxPenalty
	}
	return penalty
}

func regionsAreNeighbors(region *world.Region, neighborID world.RegionID) bool {
	if region == nil || neighborID == "" {
		return false
	}
	for _, id := range region.Neighbors {
		if id == neighborID {
			return true
		}
	}
	return false
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

// IsArmyDefendingSiegedRegion aktif kuşatma altındaki bölgede savunmacı
// tarafta duran bir kara ordusu olup olmadığını döner. Bölge sahibi orduları
// ile bölge sahibinin aynı realm içindeki veya müttefik orduları savunmacı
// kabul edilir; kuşatan tarafın orduları bu kapsama girmez.
func (s *GameState) IsArmyDefendingSiegedRegion(candidate *army.Army) bool {
	if s == nil || candidate == nil || candidate.IsNaval || candidate.OwnerID == "" || candidate.RegionID == "" {
		return false
	}
	region := s.Regions[candidate.RegionID]
	if region == nil || region.IsSea || region.OwnerID == "" {
		return false
	}
	siege := s.SiegeAt(candidate.RegionID)
	if siege == nil || siege.AttackerArmyID == "" || siege.AttackerArmyID == candidate.ID {
		return false
	}
	siegeArmy := s.Armies[siege.AttackerArmyID]
	if siegeArmy == nil || siegeArmy.OwnerID == "" || siegeArmy.OwnerID == candidate.OwnerID {
		return false
	}
	if candidate.OwnerID == region.OwnerID {
		return true
	}
	if stateSameRealm(s, faction.FactionID(candidate.OwnerID), faction.FactionID(region.OwnerID)) {
		return true
	}
	key := faction.RelationKey(faction.FactionID(candidate.OwnerID), faction.FactionID(region.OwnerID))
	rel := s.Relations[key]
	return rel != nil && rel.Stance == faction.StanceAllied
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

// CanArmyReplenishIn ordunun bulunduğu kara bölgesinde veya bağlı olduğu
// limanda toparlanıp toparlanamayacağını döner. Kendi ve müttefik bölgelerine
// ek olarak aynı realm içindeki vassal bölgeleri de dost ikmal alanıdır.
func (s *GameState) CanArmyReplenishIn(a *army.Army) bool {
	if s == nil || a == nil || a.OwnerID == "" {
		return false
	}

	regionID := a.RegionID
	if a.IsNaval {
		regionID = a.DockedRegionID
	}
	region := s.Regions[regionID]
	if region == nil || region.IsSea || region.OwnerID == "" {
		return false
	}
	if a.IsNaval && !region.HasPort() {
		return false
	}
	return s.canFactionReplenishIn(a.OwnerID, region.OwnerID)
}

// ReplenishArmyInFriendlyTerritory, ücretsiz kara ordusu toparlanmasının
// konum kararını state'teki vassal/ittifak ilişkileriyle birlikte uygular.
func (s *GameState) ReplenishArmyInFriendlyTerritory(a *army.Army, amount int) int {
	if s == nil || a == nil || a.IsNaval || !s.CanArmyReplenishIn(a) {
		return 0
	}
	return a.Replenish(amount)
}

// ArmyReplenishmentHP, dost bir kara bölgesindeki bir ordunun ikmal yeterliyse
// her hasarlı birimine uygulayabileceği ücretsiz toparlanma miktarını döner.
// Çiftlik düzenli üretimi, ambar ise eldeki tahılın korunup dağıtılmasını
// temsil eder; ikisi de aynı lineer toparlanma katkısını sağlar.
func (s *GameState) ArmyReplenishmentHP(a *army.Army) int {
	if s == nil || a == nil || a.IsNaval || !s.CanArmyReplenishIn(a) {
		return 0
	}
	return s.RegionArmyReplenishmentHP(s.Regions[a.RegionID])
}

// RegionArmyReplenishmentHP, ikmal yeterliyse bölgenin çiftlik ve ambar
// seviyelerinden türeyen kara ordusu toparlanma hızını döner. AI hedef
// değerlendirmesi ile tur çözümlemesi aynı katsayıyı bu helper üzerinden alır.
func (s *GameState) RegionArmyReplenishmentHP(region *world.Region) int {
	if s == nil || region == nil || region.IsSea {
		return 0
	}
	const baseHP = 2
	const buildingLevelHP = 2
	return baseHP + buildingLevelHP*(region.BuildingLevel("farm")+region.BuildingLevel("granary"))
}

// ClearArmyLogisticsAfterRelocation eski konumda üretilmiş lojistik yıpranma
// snapshot'ını farklı konuma geçen ordudan ayırır. Kara ordusunun aşırı yük
// sayacı da bölgeye özgü olduğu için hedef bölgede yeniden başlatılır; deniz
// filosunun TurnsWithoutPort sayacı ise açık deniz yolculuğunun toplam süresini
// temsil ettiğinden korunur.
func (s *GameState) ClearArmyLogisticsAfterRelocation(a *army.Army, previousLocation string) bool {
	if s == nil || a == nil || previousLocation == "" || a.LocationID() == previousLocation {
		return false
	}

	cleared := false
	if s.ArmyLogistics != nil {
		if _, ok := s.ArmyLogistics[a.ID]; ok {
			delete(s.ArmyLogistics, a.ID)
			cleared = true
		}
	}
	if !a.IsNaval && a.OverCapacityTurns != 0 {
		a.OverCapacityTurns = 0
		cleared = true
	}
	return cleared
}

// CanFleetAvoidSeaAttrition, filonun bulunduğu deniz bölgesine komşu en az
// bir limanlı kara bölgesinin filo sahibine ait, aynı realm içinde veya
// müttefik olup olmadığını döner. Liman, denizdeki güvenli ikmal hattının
// zorunlu parçasıdır.
func (s *GameState) CanFleetAvoidSeaAttrition(a *army.Army) bool {
	if s == nil || a == nil || !a.IsAtSea() || a.OwnerID == "" {
		return false
	}

	seaRegion := s.Regions[a.RegionID]
	if seaRegion == nil || !seaRegion.IsSea {
		return false
	}
	for _, neighborID := range seaRegion.Neighbors {
		neighbor := s.Regions[neighborID]
		if neighbor == nil || neighbor.IsSea || neighbor.OwnerID == "" || !neighbor.HasPort() {
			continue
		}
		if s.canFactionReplenishIn(a.OwnerID, neighbor.OwnerID) {
			return true
		}
	}
	return false
}

func (s *GameState) canFactionReplenishIn(armyOwner, regionOwner string) bool {
	if s == nil || armyOwner == "" || regionOwner == "" {
		return false
	}
	if armyOwner == regionOwner || stateSameRealm(s, faction.FactionID(armyOwner), faction.FactionID(regionOwner)) {
		return true
	}
	rel := s.Relations[faction.RelationKey(faction.FactionID(armyOwner), faction.FactionID(regionOwner))]
	return rel != nil && rel.Stance == faction.StanceAllied
}

// RegionProductionSummary hesaplanan efektif bölge üretimini döner.
func (s *GameState) regionProductionSummary(region *world.Region, applyBlockade bool) RegionProductionSummary {
	if s == nil || region == nil || region.IsSea || region.OwnerID == "" {
		return RegionProductionSummary{}
	}
	blockadeRetention := 100
	if applyBlockade {
		blockadeRetention = s.RegionBlockadeOutputRetentionPercent(region)
	}

	goldMod := 1.0
	grainMod := 1.0
	for _, bid := range region.Buildings {
		if b, ok := s.BuildingTypes[bid]; ok && b != nil {
			goldMod *= b.GoldMod
			grainMod *= b.GrainMod
		}
	}

	out := RegionProductionSummary{
		Gold:   scaleBlockadeOutput(int(float64(region.GoldIncome())*goldMod*float64(s.CurrentSeason().HarvestMod())/100), blockadeRetention),
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
	out.Grain = scaleBlockadeOutput(out.Grain, blockadeRetention)
	out.Iron = scaleBlockadeOutput(out.Iron, blockadeRetention)
	out.Timber = scaleBlockadeOutput(out.Timber, blockadeRetention)
	out.Stone = scaleBlockadeOutput(out.Stone, blockadeRetention)
	out.Spice = scaleBlockadeOutput(out.Spice, blockadeRetention)
	out.Cloth = scaleBlockadeOutput(out.Cloth, blockadeRetention)

	tradeIncome := s.BaseRegionTradeIncome(region) * s.CurrentSeason().TradeMod() / 100
	tradeIncome = scaleBlockadeOutput(tradeIncome, blockadeRetention)
	if fx, ok := s.Factions[faction.FactionID(region.OwnerID)]; ok && fx != nil && s.TechTypes != nil {
		tradeIncome = int(float64(tradeIncome) * (1.0 + tech.ComputeEffects(fx.Research.Completed, s.TechTypes).MarketGoldMod))
	}
	out.Gold += tradeIncome

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

	productionPercent := 100 + s.RegionGrainProductionModifier(region.ID)
	if productionPercent < 0 {
		productionPercent = 0
	}
	out.Grain = out.Grain * productionPercent / 100

	return out
}

// RegionProductionSummary, abluka etkisi dahil bölgenin sonraki tur üretim
// önizlemesini döner. Gerçek ekonomi tick'i ile UI/AI aynı helper zincirini
// kullanır.
func (s *GameState) RegionProductionSummary(region *world.Region) RegionProductionSummary {
	return s.regionProductionSummary(region, true)
}

// UnblockedRegionProductionSummary abluka uygulanmadan önceki yerel çıktıyı
// döner. Abluka loot'u, hedef bölgenin gerçekten üretebildiği mevcut tabanını
// değil, ablukasız ekonomik değerini baz alır.
func (s *GameState) UnblockedRegionProductionSummary(region *world.Region) RegionProductionSummary {
	return s.regionProductionSummary(region, false)
}

// FactionProductionSummary devletin kuşatma altında olmayan bölgelerinin
// efektif tur başı üretimini toplar. HUD ve ekonomi önizlemeleri aynı bina,
// mevsim, arazi ve teknoloji hesaplarını kullanır.
func (s *GameState) FactionProductionSummary(fid faction.FactionID) RegionProductionSummary {
	if s == nil || fid == "" {
		return RegionProductionSummary{}
	}

	var out RegionProductionSummary
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || s.SiegeAt(region.ID) != nil {
			continue
		}
		production := s.RegionProductionSummary(region)
		out.Gold += production.Gold
		out.Grain += production.Grain
		out.Iron += production.Iron
		out.Timber += production.Timber
		out.Stone += production.Stone
		out.Spice += production.Spice
		out.Cloth += production.Cloth
	}
	loot := s.BlockadeLootForFaction(fid)
	out.Gold += loot.Gold
	out.Grain += loot.Grain
	out.Iron += loot.Iron
	out.Timber += loot.Timber
	out.Stone += loot.Stone
	out.Spice += loot.Spice
	out.Cloth += loot.Cloth
	return out
}

// FactionGrainNetChange devletin bir turdaki tahıl üretimi ile sivil ve ordu
// tüketimi arasındaki farkı döner. Negatif sonuç stokun azalacağını gösterir.
func (s *GameState) FactionGrainNetChange(fid faction.FactionID) int {
	if s == nil || fid == "" {
		return 0
	}

	net := s.FactionProductionSummary(fid).Grain
	for _, region := range s.Regions {
		if region == nil || region.IsSea || region.OwnerID != string(fid) || s.SiegeAt(region.ID) != nil {
			continue
		}
		net -= s.CivilianGrainDemandForRegion(region)
	}
	for _, currentArmy := range s.Armies {
		if currentArmy != nil && currentArmy.OwnerID == string(fid) {
			net -= s.EffectiveArmyGrainUpkeep(currentArmy)
		}
	}
	return net
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

// CanRestoreSuccessorAtRegion, bir bölgenin ardıl devlet karar paneline
// girebilmesi için ardıl fraksiyonun gerçekten oyundan elenmiş ve kara toprağı
// kalmamış olması gerektiğini doğrular.
func (s *GameState) CanRestoreSuccessorAtRegion(region *world.Region) bool {
	if s == nil || region == nil || region.SuccessorFactionID == "" {
		return false
	}
	successorID := faction.FactionID(region.SuccessorFactionID)
	successor := s.Factions[successorID]
	return successor != nil && successor.IsEliminated && len(s.LandRegionsOwnedBy(successorID)) == 0
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
