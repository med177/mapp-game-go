package economy

import (
	"fmt"
	"math"
	"strings"

	"mapp-game-go/internal/city"
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

// BuildingCostAtLevel, JSON'daki bina maliyetini hedef seviyeye göre bileşik
// olarak artırır. Seviye 1 her zaman taban maliyettir; eksik veya 1'in altındaki
// çarpanlar loader tarafından 1.0'a normalize edilir.
func BuildingCostAtLevel(building *city.Building, level int) ResourceCost {
	if building == nil {
		return ResourceCost{}
	}
	if level < 1 {
		level = 1
	}
	multiplier := building.UpgradeCostMultiplier
	if multiplier < 1.0 {
		multiplier = 1.0
	}
	factor := math.Pow(multiplier, float64(level-1))
	scale := func(base int) int {
		if base <= 0 {
			return 0
		}
		return int(math.Round(float64(base) * factor))
	}
	return ResourceCost{
		Gold:   scale(building.GoldCost),
		Grain:  scale(building.GrainCost),
		Iron:   scale(building.IronCost),
		Timber: scale(building.TimberCost),
		Stone:  scale(building.StoneCost),
		Spice:  scale(building.SpiceCost),
		Cloth:  scale(building.ClothCost),
	}
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

// RefundPercent, maliyetin verilen yüzde kadarını her kaynak için aşağı
// yuvarlayarak iade eder. Üretim maliyeti olmayan kaynaklar değişmez.
func (c ResourceCost) RefundPercent(f *faction.Faction, percent int) {
	if f == nil || percent <= 0 {
		return
	}
	if percent > 100 {
		percent = 100
	}
	for _, kind := range CostResourceKinds() {
		amount := c.Amount(kind) * percent / 100
		if amount > 0 {
			AddFactionResource(f, kind, amount)
		}
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
