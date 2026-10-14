package state

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
)

const (
	CouncilRoleHeadOfHouse         = "head_of_house"
	CouncilRoleTreasurer           = "treasurer"
	CouncilRoleDiplomat            = "diplomat"
	CouncilRoleHeir                = "heir"
	CouncilRoleMilitaryResponsible = "military_responsible"
	TreasurerTaxIncomeBonusPercent = 5
	DiplomatRelationBonusPoints    = 3
)

func validCouncilRole(role string) bool {
	return role == CouncilRoleHeadOfHouse || role == CouncilRoleTreasurer || role == CouncilRoleDiplomat ||
		role == CouncilRoleHeir || role == CouncilRoleMilitaryResponsible
}

func (s *GameState) councilCharacterAvailable(ownerID string, character *army.Commander) bool {
	return s != nil && ownerID != "" && character != nil &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter) &&
		character.OwnerID == ownerID && character.ActiveInYear(s.Year) &&
		character.AssignedArmyID == "" && (character.AssignedCouncilRole == "" || validCouncilRole(character.AssignedCouncilRole)) &&
		!s.DeadCommanderIDs[character.ID] && !s.DismissedCommanderIDs[character.ID] &&
		!s.ExiledCommanderIDs[character.ID] && character.CaptiveAtSettlementID == "" &&
		!s.CommanderUnavailable(character)
}

// CanAssignCharacterToCouncilRole atama veya UI disabled durumu için
// karakterin belirtilen konsey koltuğuna atanıp atanamayacağını döner.
func (s *GameState) CanAssignCharacterToCouncilRole(ownerID faction.FactionID, characterID, role string) bool {
	if s == nil || ownerID == "" || characterID == "" || !validCouncilRole(role) {
		return false
	}
	owner := s.Factions[ownerID]
	return owner != nil && !owner.IsEliminated && s.councilCharacterAvailable(string(ownerID), s.Commanders[characterID])
}

// AssignCharacterToCouncilRole, faction'ın aktif yönetici/karakterlerinden
// birini konsey görevine atar veya mevcut görevinden geçirir. Dolu koltuğun eski
// sahibi boşa çıkar.
func (s *GameState) AssignCharacterToCouncilRole(ownerID faction.FactionID, characterID, role string) bool {
	if !s.CanAssignCharacterToCouncilRole(ownerID, characterID, role) {
		return false
	}
	character := s.Commanders[characterID]
	for _, incumbent := range s.Commanders {
		if incumbent != nil && incumbent.OwnerID == string(ownerID) && incumbent.AssignedCouncilRole == role && incumbent.ID != characterID {
			incumbent.AssignedCouncilRole = ""
		}
	}
	character.AssignedCouncilRole = role
	return true
}

// UnassignCharacterFromCouncilRole karakteri konsey koltuğundan çıkarır.
func (s *GameState) UnassignCharacterFromCouncilRole(ownerID faction.FactionID, characterID string) bool {
	if s == nil || ownerID == "" || characterID == "" {
		return false
	}
	character := s.Commanders[characterID]
	if character == nil || character.OwnerID != string(ownerID) || character.AssignedCouncilRole == "" {
		return false
	}
	character.AssignedCouncilRole = ""
	return true
}

// CouncilRoleHolder konsey koltuğunun kullanılabilir mevcut sahibini döner.
// Eski/hatalı save'de aynı role birden fazla karakter yazılmışsa ID sırasındaki
// ilk geçerli kayıt deterministik olarak seçilir.
func (s *GameState) CouncilRoleHolder(ownerID faction.FactionID, role string) *army.Commander {
	if s == nil || ownerID == "" || !validCouncilRole(role) {
		return nil
	}
	owner := s.Factions[ownerID]
	if owner == nil || owner.IsEliminated {
		return nil
	}
	var holder *army.Commander
	for _, character := range s.Commanders {
		if character == nil || character.AssignedCouncilRole != role ||
			!s.councilCharacterAvailableForEffect(string(ownerID), character) {
			continue
		}
		if holder == nil || character.ID < holder.ID {
			holder = character
		}
	}
	return holder
}

func (s *GameState) councilCharacterAvailableForEffect(ownerID string, character *army.Commander) bool {
	return character != nil && character.OwnerID == ownerID && character.ActiveInYear(s.Year) &&
		(character.Type == army.CharacterTypeAdministrator || character.Type == army.CharacterTypeCharacter) &&
		character.AssignedArmyID == "" && !s.DeadCommanderIDs[character.ID] &&
		!s.DismissedCommanderIDs[character.ID] && !s.ExiledCommanderIDs[character.ID] &&
		character.CaptiveAtSettlementID == "" && !s.CommanderUnavailable(character)
}

// TreasurerTaxIncomeBonus yalnız bölgesel vergi gelirinin %5'ini üretir;
// ticaret, haraç ve diğer gelir kategorileri bu bonusa dahil değildir.
func (s *GameState) TreasurerTaxIncomeBonus(ownerID faction.FactionID, taxIncome int) int {
	if taxIncome <= 0 || s.CouncilRoleHolder(ownerID, CouncilRoleTreasurer) == nil {
		return 0
	}
	return taxIncome * TreasurerTaxIncomeBonusPercent / 100
}

// DiplomatRelationBonus, oyuncunun atadığı diplomatın Heyet hamlesine eklediği
// tek seferlik ilişki puanını döner. Hedef, normal diplomasi eyleminde seçilir.
func (s *GameState) DiplomatRelationBonus(ownerID faction.FactionID) int {
	if s == nil || s.CouncilRoleHolder(ownerID, CouncilRoleDiplomat) == nil {
		return 0
	}
	return DiplomatRelationBonusPoints
}
