package tech

type CategoryDef struct {
	Category Category
	NameTR   string
}

var categoryDefs = []CategoryDef{
	{Category: CategoryMilitary, NameTR: "Askeri"},
	{Category: CategoryEconomy, NameTR: "Ekonomi"},
	{Category: CategoryDiplomacy, NameTR: "Diplomasi"},
	{Category: CategoryNaval, NameTR: "Denizcilik"},
	{Category: CategoryReligion, NameTR: "Din"},
}

var categoryDefsByValue = func() map[Category]CategoryDef {
	out := make(map[Category]CategoryDef, len(categoryDefs))
	for _, def := range categoryDefs {
		out[def.Category] = def
	}
	return out
}()

var categoryOrderIndex = func() map[Category]int {
	out := make(map[Category]int, len(categoryDefs))
	for i, def := range categoryDefs {
		out[def.Category] = i
	}
	return out
}()

func AllCategories() []Category {
	out := make([]Category, 0, len(categoryDefs))
	for _, def := range categoryDefs {
		out = append(out, def.Category)
	}
	return out
}

func CategoryLabelTR(category Category) string {
	if def, ok := categoryDefsByValue[category]; ok {
		return def.NameTR
	}
	return string(category)
}

func CategoryOrder(category Category) int {
	if order, ok := categoryOrderIndex[category]; ok {
		return order
	}
	return len(categoryDefs)
}
