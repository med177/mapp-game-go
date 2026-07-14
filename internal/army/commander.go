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
	if c == nil {
		return 0
	}
	var modifier float64
	if c.HasTrait(CommanderTraitVeteran) {
		modifier += 0.02
	}
	if c.HasTrait(CommanderTraitTactician) {
		modifier += 0.04
	}
	if c.HasTrait(CommanderTraitAggressor) {
		modifier += 0.06
	}
	return modifier
}

// DefenseModifier komutanın ordunun savunma gücüne eklediği çarpanı döner.
func (c *Commander) DefenseModifier() float64 {
	if c == nil {
		return 0
	}
	var modifier float64
	if c.HasTrait(CommanderTraitVeteran) {
		modifier += 0.02
	}
	if c.HasTrait(CommanderTraitTactician) {
		modifier += 0.02
	}
	if c.HasTrait(CommanderTraitDefender) {
		modifier += 0.06
	}
	return modifier
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
