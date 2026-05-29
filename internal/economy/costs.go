package economy

import (
	"fmt"
	"strings"

	"mapp-game-go/internal/faction"
)

type ResourceCost struct {
	Gold   int
	Grain  int
	Iron   int
	Timber int
	Stone  int
}

func (c ResourceCost) CanAfford(f *faction.Faction) bool {
	if f == nil {
		return false
	}
	return f.Gold >= c.Gold &&
		f.Grain >= c.Grain &&
		f.Iron >= c.Iron &&
		f.Timber >= c.Timber &&
		f.Stone >= c.Stone
}

func (c ResourceCost) Apply(f *faction.Faction) {
	if f == nil {
		return
	}
	f.Gold -= c.Gold
	f.Grain -= c.Grain
	f.Iron -= c.Iron
	f.Timber -= c.Timber
	f.Stone -= c.Stone
}

func (c ResourceCost) Refund(f *faction.Faction) {
	if f == nil {
		return
	}
	f.Gold += c.Gold
	f.Grain += c.Grain
	f.Iron += c.Iron
	f.Timber += c.Timber
	f.Stone += c.Stone
}

func (c ResourceCost) ShortTR() string {
	parts := make([]string, 0, 5)
	if c.Gold > 0 {
		parts = append(parts, fmt.Sprintf("%d Altın", c.Gold))
	}
	if c.Grain > 0 {
		parts = append(parts, fmt.Sprintf("%d Tahıl", c.Grain))
	}
	if c.Iron > 0 {
		parts = append(parts, fmt.Sprintf("%d Demir", c.Iron))
	}
	if c.Timber > 0 {
		parts = append(parts, fmt.Sprintf("%d Kereste", c.Timber))
	}
	if c.Stone > 0 {
		parts = append(parts, fmt.Sprintf("%d Taş", c.Stone))
	}
	if len(parts) == 0 {
		return "Bedava"
	}
	return strings.Join(parts, ", ")
}

