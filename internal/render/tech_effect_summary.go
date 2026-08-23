package render

import (
	"strings"

	"mapp-game-go/internal/tech"
)

func techRequirementLabels(requirementIDs []string, allTechs map[string]*tech.Technology) string {
	labels := make([]string, 0, len(requirementIDs))
	for _, requirementID := range requirementIDs {
		label := requirementID
		if requiredTech, ok := allTechs[requirementID]; ok && requiredTech != nil && requiredTech.NameTR != "" {
			label = requiredTech.NameTR
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, ", ")
}

func techRequirementSummary(t *tech.Technology, currentYear int, ownedRegions map[string]bool, regionNames map[string]string) string {
	if t == nil || (len(t.RequiredRegions) == 0 && t.MinYear <= 0) {
		return ""
	}
	parts := make([]string, 0, 2)
	if len(t.RequiredRegions) > 0 {
		regions := make([]string, 0, len(t.RequiredRegions))
		for _, regionID := range t.RequiredRegions {
			if regionID == "" {
				continue
			}
			name := regionNames[regionID]
			if name == "" {
				name = regionID
			}
			mark := "✕"
			if ownedRegions[regionID] {
				mark = "✓"
			}
			regions = append(regions, mark+name)
		}
		if len(regions) > 0 {
			parts = append(parts, "Bölgeler: "+strings.Join(regions, ", "))
		}
	}
	if t.MinYear > 0 {
		yearText := "En erken yıl: " + itoa(t.MinYear)
		if currentYear >= t.MinYear {
			yearText = "Yıl: ✓" + itoa(t.MinYear)
		}
		parts = append(parts, yearText)
	}
	return strings.Join(parts, "  •  ")
}

func techEffectSummary(t *tech.Technology) string {
	if t == nil {
		return ""
	}
	return techEffectsSummary(t.Effects, t.DescriptionTR)
}

func techEffectsSummary(e tech.Effects, fallback string) string {
	parts := make([]string, 0, 3)
	if e.InfantryAttackMod > 0 {
		parts = append(parts, "Piyade +%"+itoa(int(e.InfantryAttackMod*100)))
	}
	if e.CavalryAttackMod > 0 {
		parts = append(parts, "Süvari +%"+itoa(int(e.CavalryAttackMod*100)))
	}
	if e.SiegeAttackMod > 0 {
		parts = append(parts, "Kuşatma +%"+itoa(int(e.SiegeAttackMod*100)))
	}
	if e.NavalAttackMod > 0 {
		parts = append(parts, "Deniz atk +%"+itoa(int(e.NavalAttackMod*100)))
	}
	if e.NavalDefenseMod > 0 {
		parts = append(parts, "Deniz sav +%"+itoa(int(e.NavalDefenseMod*100)))
	}
	if e.LandDefenseMod > 0 {
		parts = append(parts, "Kara sav +%"+itoa(int(e.LandDefenseMod*100)))
	}
	if e.GoldPerRegion > 0 {
		parts = append(parts, "Bölge altını +"+itoa(e.GoldPerRegion))
	}
	if e.GrainMod > 0 {
		parts = append(parts, "Tahıl +%"+itoa(int(e.GrainMod*100)))
	}
	if e.IronMod > 0 {
		parts = append(parts, "Demir +%"+itoa(int(e.IronMod*100)))
	}
	if e.TimberMod > 0 {
		parts = append(parts, "Kereste +%"+itoa(int(e.TimberMod*100)))
	}
	if e.StoneMod > 0 {
		parts = append(parts, "Taş +%"+itoa(int(e.StoneMod*100)))
	}
	if e.MarketGoldMod > 0 {
		parts = append(parts, "Ticaret +%"+itoa(int(e.MarketGoldMod*100)))
	}
	if e.PeaceRelationBonus > 0 {
		parts = append(parts, "Barış +"+itoa(e.PeaceRelationBonus))
	}
	if e.NavalMoveBonus > 0 {
		parts = append(parts, "Deniz hareket +"+itoa(e.NavalMoveBonus))
	}
	if e.MoveBonus > 0 {
		parts = append(parts, "Hareket +"+itoa(e.MoveBonus))
	}
	if e.SatisfactionBonus > 0 {
		parts = append(parts, "Memnuniyet +"+itoa(e.SatisfactionBonus))
	}
	if e.ConversionSpeedMod > 0 {
		parts = append(parts, "Dönüşüm +"+itoa(int(e.ConversionSpeedMod)))
	}
	if e.RevealEnemyStrength {
		parts = append(parts, "Tam istihbarat")
	}
	if len(parts) == 0 {
		return strings.TrimSpace(fallback)
	}
	if len(parts) > 3 {
		parts = parts[:3]
	}
	return strings.Join(parts, "  •  ")
}
