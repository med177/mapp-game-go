package tech

import "testing"

func TestCategoryLabelTR(t *testing.T) {
	if got := CategoryLabelTR(CategoryNaval); got != "Denizcilik" {
		t.Fatalf("category label mismatch: got=%q", got)
	}
}

func TestCategoryOrder(t *testing.T) {
	if CategoryOrder(CategoryMilitary) >= CategoryOrder(CategoryReligion) {
		t.Fatalf("category order should preserve declared metadata sequence")
	}
}
