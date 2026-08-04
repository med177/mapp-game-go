package army

// CommanderTrait ordunun komutanından gelen kalıcı uzmanlık türünü belirtir.
type CommanderTrait string

const (
	CommanderTraitVeteran   CommanderTrait = "veteran"
	CommanderTraitTactician CommanderTrait = "tactician"
	CommanderTraitDefender  CommanderTrait = "defender"
	CommanderTraitAggressor CommanderTrait = "aggressor"
)

const (
	CommanderStartingLevel = 1
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
	NewTraits     []CommanderTrait
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
	ID             string           `json:"id"`
	OwnerID        string           `json:"owner_id"`
	AssignedArmyID ArmyID           `json:"assigned_army_id,omitempty"`
	Name           string           `json:"name"`
	PortraitAsset  string           `json:"portrait_asset,omitempty"`
	Level          int              `json:"level"`
	Experience     int              `json:"experience"`
	Battles        int              `json:"battles"`
	Victories      int              `json:"victories"`
	Traits         []CommanderTrait `json:"traits,omitempty"`
	StartYear      int              `json:"start_year,omitempty"`
	EndYear        int              `json:"end_year,omitempty"`
}

// ActiveInYear, senaryo komutanının verilen yılda tarihsel olarak görevde
// olup olmadığını döner. EndYear doğum/ölüm aralığının üst sınırıdır; komutan
// EndYear başladığında artık atanamaz.
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
		Name:  name,
		Level: CommanderStartingLevel,
	}
}

// Normalize eski veya eksik save verisinden gelen komutan state'ini güvenli hale getirir.
func (c *Commander) Normalize() {
	if c == nil {
		return
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
	c.syncProgression(nil)
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
func (c *Commander) HasTrait(trait CommanderTrait) bool {
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
	if c.HasTrait(CommanderTraitVeteran) {
		effects.AttackMod += 0.02
		effects.DefenseMod += 0.02
		effects.MoraleMod += 0.08
	}
	if c.HasTrait(CommanderTraitTactician) {
		effects.AttackMod += 0.04
		effects.DefenseMod += 0.02
		effects.MoveBonus++
	}
	if c.HasTrait(CommanderTraitDefender) {
		effects.DefenseMod += 0.06
		effects.MoraleMod += 0.05
	}
	if c.HasTrait(CommanderTraitAggressor) {
		effects.AttackMod += 0.06
		effects.SiegeProgressBonus++
		effects.SiegeBreachBonus++
	}
	return effects
}

// TraitLabelTR trait'lerin oyuncuya gösterilecek adını döner.
func TraitLabelTR(trait CommanderTrait) string {
	switch trait {
	case CommanderTraitVeteran:
		return "Savaş Tecrübesi"
	case CommanderTraitTactician:
		return "Taktisyen"
	case CommanderTraitDefender:
		return "Savunma Uzmanı"
	case CommanderTraitAggressor:
		return "Saldırgan"
	default:
		return string(trait)
	}
}

func (c *Commander) syncProgression(existing []CommanderTrait) []CommanderTrait {
	if c == nil {
		return nil
	}
	if c.Experience >= CommanderLevel5XP {
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

	newTraits := make([]CommanderTrait, 0, 1)
	addTrait := func(trait CommanderTrait, requiredLevel int) {
		if c.Level < requiredLevel || c.HasTrait(trait) {
			return
		}
		c.Traits = append(c.Traits, trait)
		if existing != nil {
			for _, old := range existing {
				if old == trait {
					return
				}
			}
		}
		newTraits = append(newTraits, trait)
	}

	addTrait(CommanderTraitVeteran, 2)
	addTrait(CommanderTraitTactician, 3)
	addTrait(CommanderTraitDefender, 4)
	addTrait(CommanderTraitAggressor, 5)
	return newTraits
}
