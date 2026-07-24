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
	Spice  int
	Cloth  int
}

func (c ResourceCost) Amount(kind ResourceKind) int {
	switch kind {
	case ResourceGold:
		return c.Gold
	case ResourceGrain:
		return c.Grain
	case ResourceIron:
		return c.Iron
	case ResourceTimber:
		return c.Timber
	case ResourceStone:
		return c.Stone
	case ResourceSpice:
		return c.Spice
	case ResourceCloth:
		return c.Cloth
	default:
		return 0
	}
}

func (c ResourceCost) CanAfford(f *faction.Faction) bool {
	if f == nil {
		return false
	}
	for _, kind := range CostResourceKinds() {
		if FactionResourceAmount(f, kind) < c.Amount(kind) {
			return false
		}
	}
	return true
}

func (c ResourceCost) Apply(f *faction.Faction) {
	if f == nil {
		return
	}
	for _, kind := range CostResourceKinds() {
		AddFactionResource(f, kind, -c.Amount(kind))
	}
}

func (c ResourceCost) Refund(f *faction.Faction) {
	if f == nil {
		return
	}
	for _, kind := range CostResourceKinds() {
		AddFactionResource(f, kind, c.Amount(kind))
	}
}

func (c ResourceCost) ShortTR() string {
	parts := make([]string, 0, len(CostResourceKinds()))
	for _, kind := range CostResourceKinds() {
		if amount := c.Amount(kind); amount > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", amount, ResourceNameTR(kind)))
		}
	}
	if len(parts) == 0 {
		return "Bedava"
	}
	return strings.Join(parts, ", ")
}
