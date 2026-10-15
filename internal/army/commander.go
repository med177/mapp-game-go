package army

// CharacterTrait, karakterin sahip olduğu kalıcı özellik türünü belirtir.
type CharacterTrait string

// CharacterType, senaryo karakterinin oyun içindeki rolünü belirtir.
type CharacterType string

const (
	CharacterTypeAdministrator CharacterType = "administrator"
	CharacterTypeCommander     CharacterType = "commander"
	CharacterTypeCharacter     CharacterType = "character"
)

// CouncilRoleMilitaryResponsible allows its holder to command an army while serving.
const CouncilRoleMilitaryResponsible = "military_responsible"

// IsValid, karakter tipinin desteklenen değerlerden biri olup olmadığını söyler.
func (t CharacterType) IsValid() bool {
	switch t {
	case CharacterTypeAdministrator, CharacterTypeCommander, CharacterTypeCharacter:
		return true
	default:
		return false
	}
}

const (
	CharacterTraitVeteran   CharacterTrait = "veteran"
	CharacterTraitTactician CharacterTrait = "tactician"
	CharacterTraitDefender  CharacterTrait = "defender"
	CharacterTraitAggressor CharacterTrait = "aggressor"
)

const (
	CommanderStartingLevel = 1
	CommanderBossLevel     = 10
	CommanderWinXP         = 100
	CommanderLossXP        = 40
	CommanderLevel2XP      = 100
	CommanderLevel3XP      = 300
	CommanderLevel4XP      = 550
	CommanderLevel5XP      = 850
	DefaultPortraitAsset   = "default.png"
)

// CommanderProgress bir savaş sonrasında komutanın kazandığı ilerlemeyi özetler.
// Domain katmanı render paketine bağımlı kalmadan UI ve battle report bildirimlerini
// besler.
type CommanderProgress struct {
	XPGained      int
	PreviousLevel int
	CurrentLevel  int
	NewTraits     []CharacterTrait
}

// CommanderUnavailableStatus, tutsaklık/sürgün gibi ölüm dışı event flag'leri
// aktifken komutanın neden atanamadığını ve UI metnini tanımlar. Ölüm ayrı
// GameState.DeadCommanderIDs state'inde tutulur.
type CommanderUnavailableStatus struct {
	Flag           string `json:"flag"`
	ReasonTR       string `json:"reason_tr"`
	HideFromRoster bool   `json:"hide_from_roster,omitempty"`
	Permanent      bool   `json:"permanent,omitempty"`
}

// CommanderEffects komutanın orduya verdiği tüm hesap etkilerini tek yerde toplar.
type CommanderEffects struct {
	AttackMod          float64
	DefenseMod         float64
	MoraleMod          float64
	MoveBonus          int
	SiegeProgressBonus int
	SiegeBreachBonus   int
}

// Commander bir orduya atanmış komutanın kalıcı kariyer state'idir.
type Commander struct {
	ID                          string                       `json:"id"`
	Type                        CharacterType                `json:"type"`
	OwnerID                     string                       `json:"owner_id"`
	AssignedArmyID              ArmyID                       `json:"assigned_army_id,omitempty"`
	AssignedCouncilRole         string                       `json:"assigned_council_role,omitempty"`
	CouncilLevel                int                          `json:"council_level,omitempty"`
	CouncilExperience           int                          `json:"council_experience,omitempty"`
	ActiveMissionID             string                       `json:"active_mission_id,omitempty"`
	ActiveMissionTargetID       string                       `json:"active_mission_target_id,omitempty"`
	ActiveMissionTurnsRemaining int                          `json:"active_mission_turns_remaining,omitempty"`
	Name                        string                       `json:"name"`
	PortraitAsset               string                       `json:"portrait_asset,omitempty"`
	Level                       int                          `json:"level"`
	IsBoss                      bool                         `json:"is_boss,omitempty"`
	Experience                  int                          `json:"experience"`
	Battles                     int                          `json:"battles"`
	Victories                   int                          `json:"victories"`
	AdoptionCostGold            int                          `json:"adoption_cost_gold,omitempty"`
	GoldUpkeep                  int                          `json:"gold_upkeep,omitempty"`
	Traits                      []CharacterTrait             `json:"traits,omitempty"`
	UnavailableStatuses         []CommanderUnavailableStatus `json:"unavailable_statuses,omitempty"`
	InjuryTurnsOnDefeat         int                          `json:"injury_turns_on_defeat,omitempty"`
	UnavailableUntilTurn        int                          `json:"unavailable_until_turn,omitempty"`
	UnavailableReasonTR         string                       `json:"unavailable_reason_tr,omitempty"`
	CaptiveAtSettlementID       string                       `json:"captive_at_settlement_id,omitempty"`
	CaptivityReleaseFlags       []string                     `json:"captivity_release_flags,omitempty"`
	CaptivityKeepPenaltyApplied bool                         `json:"captivity_keep_penalty_applied,omitempty"`
	StartYear                   int                          `json:"start_year,omitempty"`
	EndYear                     int                          `json:"end_year,omitempty"`
}

// ActiveInYear, senaryo komutanının verilen yılda tarihsel olarak yaşayıp
// görevde olup olmadığını döner. EndYear doğal ölüm yılının ilk günüdür;
// komutan bu yıl başladığında artık aktif değildir.
func (c *Commander) ActiveInYear(year int) bool {
	if c == nil {
		return false
	}
	if c.StartYear != 0 && year < c.StartYear {
		return false
	}
	return c.EndYear == 0 || year < c.EndYear
}

// NewCommander yeni bir komutanı başlangıç seviyesiyle oluşturur.
func NewCommander(id, name string) *Commander {
	return &Commander{
		ID:    id,
		Type:  CharacterTypeCommander,
		Name:  name,
		Level: CommanderStartingLevel,
	}
}

// CanCommandArmy, karakterin kara ordusu veya donanma komutanlığına atanıp atanamayacağını söyler.
// Tipi eksik eski kayıtlar komutan kabul edilir.
func (c *Commander) CanCommandArmy() bool {
	return c != nil && (c.Type == "" || c.Type == CharacterTypeAdministrator || c.Type == CharacterTypeCommander)
}

// Normalize eski veya eksik save verisinden gelen komutan state'ini güvenli hale getirir.
func (c *Commander) Normalize() {
	if c == nil {
		return
	}
	if c.Type == "" {
		c.Type = CharacterTypeCommander
	}
	if c.Level < CommanderStartingLevel {
		c.Level = CommanderStartingLevel
	}
	if c.Experience < 0 {
		c.Experience = 0
	}
	if c.Battles < 0 {
		c.Battles = 0
	}
	if c.Victories < 0 {
		c.Victories = 0
	}
	if c.Victories > c.Battles {
		c.Victories = c.Battles
	}
	c.NormalizeCouncilProgress()
	c.NormalizeMissionState()
	c.syncProgression(nil)
}

func (c *Commander) NormalizeCouncilProgress() {
	if c == nil {
		return
	}
	if c.CouncilExperience < 0 {
		c.CouncilExperience = 0
	}
	if c.CouncilLevel > 0 || c.CouncilExperience > 0 {
		c.CouncilLevel = CouncilLevelForExperience(c.CouncilExperience)
	}
}

func (c *Commander) NormalizeMissionState() {
	if c == nil {
		return
	}
	if c.ActiveMissionTurnsRemaining < 0 {
		c.ActiveMissionTurnsRemaining = 0
	}
	if c.ActiveMissionID == "" {
		c.ActiveMissionTargetID = ""
		c.ActiveMissionTurnsRemaining = 0
	}
}

func CouncilLevelForExperience(experience int) int {
	switch {
	case experience >= CommanderLevel5XP:
		return 5
	case experience >= CommanderLevel4XP:
		return 4
	case experience >= CommanderLevel3XP:
		return 3
	case experience >= CommanderLevel2XP:
		return 2
	default:
		return CommanderStartingLevel
	}
}

func (c *Commander) RecordCouncilExperience(amount int) (previousLevel, currentLevel int) {
	if c == nil || amount <= 0 {
		return 0, 0
	}
	c.NormalizeCouncilProgress()
	previousLevel = c.CouncilLevel
	if previousLevel == 0 {
		previousLevel = CommanderStartingLevel
	}
	c.CouncilExperience += amount
	c.CouncilLevel = CouncilLevelForExperience(c.CouncilExperience)
	return previousLevel, c.CouncilLevel
}

// RecordBattle komutanın katıldığı bir savaşı kariyerine işler.
// Kazanılan savaş daha yüksek XP verir; trait'ler seviye eşiklerinde otomatik açılır.
func (c *Commander) RecordBattle(won bool) CommanderProgress {
	if c == nil {
		return CommanderProgress{}
	}
	c.Normalize()
	previousLevel := c.Level
	c.Battles++
	if won {
		c.Victories++
	}
	xp := CommanderLossXP
	if won {
		xp = CommanderWinXP
	}
	c.Experience += xp
	newTraits := c.syncProgression(nil)
	return CommanderProgress{
		XPGained:      xp,
		PreviousLevel: previousLevel,
		CurrentLevel:  c.Level,
		NewTraits:     newTraits,
	}
}

// HasTrait komutanın belirtilen uzmanlığı kazanıp kazanmadığını söyler.
func (c *Commander) HasTrait(trait CharacterTrait) bool {
	if c == nil {
		return false
	}
	for _, current := range c.Traits {
		if current == trait {
			return true
		}
	}
	return false
}

// AttackModifier komutanın ordunun saldırı gücüne eklediği çarpanı döner.
func (c *Commander) AttackModifier() float64 {
	return c.Effects().AttackMod
}

// DefenseModifier komutanın ordunun savunma gücüne eklediği çarpanı döner.
func (c *Commander) DefenseModifier() float64 {
	return c.Effects().DefenseMod
}

// MoraleModifier komutanın ordunun moral tabanlı savaş dayanıklılığına eklediği çarpanı döner.
func (c *Commander) MoraleModifier() float64 {
	return c.Effects().MoraleMod
}

// MoveBonus komutanın tur başı hareket havuzuna eklediği puanı döner.
func (c *Commander) MoveBonus() int {
	return c.Effects().MoveBonus
}

// SiegeBonuses komutanın kuşatma ilerleme ve gedik kazanımına verdiği bonusları döner.
func (c *Commander) SiegeBonuses() (progress, breach int) {
	effects := c.Effects()
	return effects.SiegeProgressBonus, effects.SiegeBreachBonus
}

// Effects komutanın tüm savaş ve operasyon etkilerini tek yerde hesaplar.
func (c *Commander) Effects() CommanderEffects {
	if c == nil {
		return CommanderEffects{}
	}
	effects := CommanderEffects{}
	if c.HasTrait(CharacterTraitVeteran) {
		effects.AttackMod += 0.02
		effects.DefenseMod += 0.02
		effects.MoraleMod += 0.08
	}
	if c.HasTrait(CharacterTraitTactician) {
		effects.AttackMod += 0.04
		effects.DefenseMod += 0.02
		effects.MoveBonus++
	}
	if c.HasTrait(CharacterTraitDefender) {
		effects.DefenseMod += 0.06
		effects.MoraleMod += 0.05
	}
	if c.HasTrait(CharacterTraitAggressor) {
		effects.AttackMod += 0.06
		effects.SiegeProgressBonus++
		effects.SiegeBreachBonus++
	}
	if c.IsBoss {
		effects.AttackMod += 0.25
		effects.DefenseMod += 0.25
		effects.MoraleMod += 0.20
		effects.MoveBonus += 2
		effects.SiegeProgressBonus += 2
		effects.SiegeBreachBonus += 2
	}
	return effects
}

// TraitLabelTR trait'lerin oyuncuya gösterilecek adını döner.
func TraitLabelTR(trait CharacterTrait) string {
	switch trait {
	case CharacterTraitVeteran:
		return "Gazi"
	case CharacterTraitTactician:
		return "Taktisyen"
	case CharacterTraitDefender:
		return "Savunmacı"
	case CharacterTraitAggressor:
		return "Saldırgan"
	case "assassin":
		return "Suikastçı"
	case "spy":
		return "Casus"
	case "explorer":
		return "Kaşif"
	case "queen":
		return "Kraliçe"
	case "intelligent":
		return "Zeki"
	case "beautiful":
		return "Güzel"
	case "manipulative":
		return "Manipülatif"
	case "sorceress":
		return "Büyücü"
	case "shadowbinder":
		return "Karanlık"
	case "mother":
		return "Korumacı"
	case "resourceful":
		return "Becerikli"
	case "loyal":
		return "Sadık"
	case "lover":
		return "Sevgili"
	case "schemer":
		return "Entrikacı"
	case "vengeful":
		return "Kindar"
	case "fighter":
		return "Savaşçı"
	case "independent":
		return "Bağımsız"
	case "princess":
		return "Prenses"
	case "free_spirit":
		return "Özgür Ruhlu"
	case "mercenary":
		return "Paralı Asker"
	case "scholar":
		return "Bilgin"
	case "cunning":
		return "Kurnaz"
	case "goodwill":
		return "Saf Kalpli"
	case "advisor":
		return "Danışman"
	case "alchemist":
		return "Simyacı"
	case "mysterious":
		return "Gizemli"
	case "skilled":
		return "Yetenekli"
	case "cruel":
		return "Zalim"
	case "spoiled":
		return "Şımarık"
	case "innocent":
		return "Masum"
	default:
		return string(trait)
	}
}

func (c *Commander) syncProgression(existing []CharacterTrait) []CharacterTrait {
	if c == nil {
		return nil
	}
	if c.IsBoss {
		c.Level = CommanderBossLevel
	} else if c.Experience >= CommanderLevel5XP {
		c.Level = 5
	} else if c.Experience >= CommanderLevel4XP {
		c.Level = 4
	} else if c.Experience >= CommanderLevel3XP {
		c.Level = 3
	} else if c.Experience >= CommanderLevel2XP {
		c.Level = 2
	} else {
		c.Level = CommanderStartingLevel
	}

	newTraits := make([]CharacterTrait, 0, 1)
	addTrait := func(trait CharacterTrait, requiredLevel int) {
		if c.Level < requiredLevel || c.HasTrait(trait) {
			return
		}
		c.Traits = append(c.Traits, trait)
		for _, old := range existing {
			if old == trait {
				return
			}
		}
		newTraits = append(newTraits, trait)
	}

	addTrait(CharacterTraitVeteran, 2)
	addTrait(CharacterTraitTactician, 3)
	addTrait(CharacterTraitDefender, 4)
	addTrait(CharacterTraitAggressor, 5)
	return newTraits
}
