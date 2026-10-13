package state

import "mapp-game-go/internal/faction"

// HasTransitAgreement iki devlet arasında aktif geçiş hakkı olup olmadığını döner.
func (s *GameState) HasTransitAgreement(a, b faction.FactionID) bool {
	if s == nil || a == "" || b == "" || a == b {
		return false
	}
	for _, agreement := range s.TransitAgreements {
		if (agreement.FactionA == a && agreement.FactionB == b) ||
			(agreement.FactionA == b && agreement.FactionB == a) {
			return true
		}
	}
	return false
}

// AddTransitAgreement mevcut değilse iki taraflı geçiş anlaşmasını kaydeder.
func (s *GameState) AddTransitAgreement(a, b faction.FactionID) bool {
	if s == nil || a == "" || b == "" || a == b || s.HasTransitAgreement(a, b) {
		return false
	}
	if a > b {
		a, b = b, a
	}
	s.TransitAgreements = append(s.TransitAgreements, DiplomaticTransitAgreement{
		FactionA: a,
		FactionB: b,
	})
	return true
}

// RemoveTransitAgreement iki taraflı geçiş anlaşmasını kaldırır.
func (s *GameState) RemoveTransitAgreement(a, b faction.FactionID) bool {
	if s == nil || a == "" || b == "" || a == b {
		return false
	}
	kept := s.TransitAgreements[:0]
	removed := false
	for _, agreement := range s.TransitAgreements {
		if (agreement.FactionA == a && agreement.FactionB == b) ||
			(agreement.FactionA == b && agreement.FactionB == a) {
			removed = true
			continue
		}
		kept = append(kept, agreement)
	}
	s.TransitAgreements = kept
	return removed
}
