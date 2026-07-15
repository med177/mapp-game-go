package army

import "testing"

func TestCommanderProgressionUnlocksTraits(t *testing.T) {
	c := NewCommander("c1", "Test Komutan")

	progress := c.RecordBattle(true)
	if progress.CurrentLevel != 2 || len(progress.NewTraits) != 1 || !c.HasTrait(CommanderTraitVeteran) {
		t.Fatalf("ilk zafer sonrası beklenen terfi/trait yok: %+v, commander=%+v", progress, c)
	}

	c.RecordBattle(true)
	if c.Level != 2 {
		t.Fatalf("200 XP sonrası seviye 2 bekleniyordu, got %d", c.Level)
	}
	c.RecordBattle(true)
	if c.Level != 3 || !c.HasTrait(CommanderTraitTactician) {
		t.Fatalf("300 XP sonrası taktikçi bekleniyordu: %+v", c)
	}
}

func TestCommanderLossProgressionAndModifiers(t *testing.T) {
	c := NewCommander("c1", "Test Komutan")
	c.RecordBattle(false)
	if c.Experience != CommanderLossXP || c.Victories != 0 || c.Battles != 1 {
		t.Fatalf("yenilgi kariyeri yanlış işlendi: %+v", c)
	}

	c.Experience = CommanderLevel4XP
	c.Normalize()
	attack, defense := c.AttackModifier(), c.DefenseModifier()
	if c.Level != 4 || !c.HasTrait(CommanderTraitDefender) || attack <= 0 || defense <= attack {
		t.Fatalf("seviye 4 komutan modifiyerleri yanlış: level=%d atk=%.2f def=%.2f traits=%v", c.Level, attack, defense, c.Traits)
	}
}

func TestCommanderBalanceMilestonesAndModifierCaps(t *testing.T) {
	lossOnly := NewCommander("c2", "Dayanıklı Komutan")
	for i := 0; i < 8; i++ {
		lossOnly.RecordBattle(false)
	}
	if lossOnly.Experience != 320 || lossOnly.Level != 3 {
		t.Fatalf("8 yenilgide kariyer temposu yanlış: xp=%d level=%d", lossOnly.Experience, lossOnly.Level)
	}

	c := NewCommander("c1", "Denge Komutanı")
	for i := 0; i < 8; i++ {
		c.RecordBattle(true)
	}
	if c.Experience != 800 || c.Level != 4 {
		t.Fatalf("8 zaferde seviye 4 bekleniyordu: xp=%d level=%d", c.Experience, c.Level)
	}
	if got := c.AttackModifier(); got != 0.06 {
		t.Fatalf("seviye 4 saldırı bonusu dengeli değil: got=%.2f", got)
	}
	if got := c.DefenseModifier(); got != 0.10 {
		t.Fatalf("seviye 4 savunma bonusu dengeli değil: got=%.2f", got)
	}

	c.RecordBattle(true)
	if c.Experience < CommanderLevel5XP || c.Level != 5 {
		t.Fatalf("9 zaferde maksimum seviye bekleniyordu: xp=%d level=%d", c.Experience, c.Level)
	}
	if got := c.AttackModifier(); got != 0.12 {
		t.Fatalf("maksimum saldırı bonusu tavanı aşıyor: got=%.2f", got)
	}
	if got := c.DefenseModifier(); got != 0.10 {
		t.Fatalf("maksimum savunma bonusu tavanı aşıyor: got=%.2f", got)
	}
	if got := c.MoraleModifier(); got != 0.13 {
		t.Fatalf("maksimum moral bonusu beklenenden farklı: got=%.2f", got)
	}
	if got := c.MoveBonus(); got != 1 {
		t.Fatalf("taktisyen hareket bonusu kayboldu: got=%d", got)
	}
	progress, breach := c.SiegeBonuses()
	if progress != 1 || breach != 1 {
		t.Fatalf("maksimum komutan kuşatma bonusları yanlış: progress=%d breach=%d", progress, breach)
	}
}

func TestArmyCommanderAssignment(t *testing.T) {
	a := &Army{}
	c := NewCommander("c1", "Test Komutan")
	if !a.AssignCommander(c) || a.Commander != c {
		t.Fatal("komutan orduya atanamadı")
	}
	if removed := a.RemoveCommander(); removed != c || a.Commander != nil {
		t.Fatal("komutan ordudan doğru ayrılmadı")
	}
}
