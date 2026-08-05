package game

import (
	"testing"

	"mapp-game-go/internal/army"
)

func TestNavalProductionKeepsMerchantAndMilitaryFleetsSeparate(t *testing.T) {
	unitTypes := map[string]*army.UnitType{
		"merchant_ship": {ID: "merchant_ship", Category: army.CategoryNavalTrade},
		"warship":       {ID: "warship", Category: army.CategoryNavalWar},
	}
	merchantFleet := &army.Army{
		IsNaval: true, TradeRouteKey: "venice->mamluk",
		Units: []army.Unit{{TypeID: "merchant_ship", CurrentHP: 100}},
	}
	if navalFleetAcceptsCompletedUnit(merchantFleet, unitTypes["warship"], unitTypes) {
		t.Fatal("savaş gemisi merchant görev filosuna eklenmemeliydi")
	}
	if !navalFleetAcceptsCompletedUnit(merchantFleet, unitTypes["merchant_ship"], unitTypes) {
		t.Fatal("ikinci merchant gemisi aynı görev filosuna eklenebilmeliydi")
	}
	merchantFleet.Units = append(merchantFleet.Units, army.Unit{TypeID: "merchant_ship", CurrentHP: 100})
	if !navalFleetAcceptsCompletedUnit(merchantFleet, unitTypes["merchant_ship"], unitTypes) {
		t.Fatal("üçüncü merchant gemisi aynı görev filosuna eklenebilmeliydi")
	}
	merchantFleet.Units = append(merchantFleet.Units, army.Unit{TypeID: "merchant_ship", CurrentHP: 100})
	if !navalFleetAcceptsCompletedUnit(merchantFleet, unitTypes["merchant_ship"], unitTypes) {
		t.Fatal("dördüncü merchant gemisi aynı görev filosuna eklenebilmeliydi")
	}
	merchantFleet.Units = append(merchantFleet.Units, army.Unit{TypeID: "merchant_ship", CurrentHP: 100})
	if navalFleetAcceptsCompletedUnit(merchantFleet, unitTypes["merchant_ship"], unitTypes) {
		t.Fatal("rota sınırını aşan beşinci merchant gemisi eklenmemeliydi")
	}
}
