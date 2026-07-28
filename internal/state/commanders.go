package state

import (
	"fmt"
	"sort"

	"mapp-game-go/internal/army"
)

const InitialPlayerCommanderPool = 3

var initialCommanderNames = []string{
	"Murat Bey",
	"Selim Bey",
	"Ayşe Hanım",
}

// SyncCommanderLinks havuzdaki komutanlarla orduların pointer bağlantısını yeniden kurar.
func (s *GameState) SyncCommanderLinks() {
	if s == nil {
		return
	}
	if s.Commanders == nil {
		s.Commanders = make(map[string]*army.Commander)
	}
	for _, commander := range s.Commanders {
		if commander == nil || commander.AssignedArmyID == "" {
			continue
		}
		s.repairGeneratedCommanderPortrait(commander)
		assigned := s.Armies[commander.AssignedArmyID]
		if assigned == nil || (assigned.Commander != commander && assigned.EmbarkedCommander != commander) {
			commander.AssignedArmyID = ""
		}
	}
	for aid, currentArmy := range s.Armies {
		if currentArmy == nil {
			continue
		}
		s.syncCommanderPointer(currentArmy, aid, currentArmy.Commander, false)
		s.syncCommanderPointer(currentArmy, aid, currentArmy.EmbarkedCommander, true)
	}
}

func (s *GameState) syncCommanderPointer(currentArmy *army.Army, aid army.ArmyID, commander *army.Commander, embarked bool) {
	if currentArmy == nil || commander == nil {
		return
	}
	commander.Normalize()
	if commander.ID == "" {
		commander.ID = fmt.Sprintf("commander_%s_%s", commander.OwnerID, aid)
	}
	if commander.OwnerID == "" {
		commander.OwnerID = currentArmy.OwnerID
	}
	s.repairGeneratedCommanderPortrait(commander)
	if pooled := s.Commanders[commander.ID]; pooled != nil {
		if pooled.AssignedArmyID != "" && pooled.AssignedArmyID != aid {
			if embarked {
				currentArmy.EmbarkedCommander = nil
			} else {
				currentArmy.Commander = nil
			}
			return
		}
		commander = pooled
		if embarked {
			currentArmy.EmbarkedCommander = pooled
		} else {
			currentArmy.Commander = pooled
		}
	} else {
		s.Commanders[commander.ID] = commander
	}
	commander.AssignedArmyID = aid
}

// RemoveArmy komutan bağlantılarını serbest bırakıp orduyu state'ten kaldırır.
func (s *GameState) RemoveArmy(armyID army.ArmyID) *army.Army {
	if s == nil || armyID == "" {
		return nil
	}
	current := s.Armies[armyID]
	if current == nil {
		return nil
	}
	for _, commander := range []*army.Commander{current.Commander, current.EmbarkedCommander} {
		if commander != nil && commander.AssignedArmyID == armyID {
			commander.AssignedArmyID = ""
		}
	}
	delete(s.Armies, armyID)
	return current
}

// NormalizeEmptyArmies, birim ve taşınmış birlik taşımayan artık ordu
// kayıtlarını state'ten kaldırır. Savaş/lojistik sırasında boşalan eski save
// kayıtları askeri gücü zaten artırmaz; map'te kalmaları ise AI'nin kapasite,
// komutan ve UI hesaplarını gereksiz yere kirletir.
func (s *GameState) NormalizeEmptyArmies() int {
	if s == nil || len(s.Armies) == 0 {
		return 0
	}
	removed := 0
	for id, current := range s.Armies {
		if current == nil || len(current.Units) > 0 || len(current.EmbarkedUnits) > 0 {
			continue
		}
		if s.RemoveArmy(id) != nil {
			removed++
		}
	}
	return removed
}

// TransferArmyOwnership ordu ve taşıdığı komutanların sahipliğini birlikte değiştirir.
func (s *GameState) TransferArmyOwnership(current *army.Army, ownerID string) {
	if current == nil || ownerID == "" {
		return
	}
	current.OwnerID = ownerID
	for _, commander := range []*army.Commander{current.Commander, current.EmbarkedCommander} {
		if commander != nil {
			commander.OwnerID = ownerID
		}
	}
}

// MoveCommanderIntoFleet kara ordusu filoya binerken komutanı taşınan birliklerle
// birlikte korur. Filoda zaten taşınan bir kara komutanı varsa ikinci komutan havuza
// bırakılır; birleşen kara birlikleri tek bir komutanla temsil edilir.
func (s *GameState) MoveCommanderIntoFleet(armyID, fleetID army.ArmyID) {
	if s == nil {
		return
	}
	landArmy := s.Armies[armyID]
	fleet := s.Armies[fleetID]
	if landArmy == nil || fleet == nil || !fleet.IsNaval || landArmy.Commander == nil {
		return
	}
	commander := landArmy.Commander
	landArmy.Commander = nil
	if fleet.EmbarkedCommander != nil {
		commander.AssignedArmyID = ""
		return
	}
	fleet.EmbarkedCommander = commander
	commander.AssignedArmyID = fleet.ID
}

// MoveEmbarkedCommanderToArmy yeni oluşan kara ordusuna taşınan komutanı bağlar.
func (s *GameState) MoveEmbarkedCommanderToArmy(fleetID, armyID army.ArmyID) {
	if s == nil {
		return
	}
	fleet := s.Armies[fleetID]
	landArmy := s.Armies[armyID]
	if fleet == nil || landArmy == nil || fleet.EmbarkedCommander == nil {
		return
	}
	commander := fleet.EmbarkedCommander
	fleet.EmbarkedCommander = nil
	if landArmy.Commander != nil {
		commander.AssignedArmyID = ""
		return
	}
	landArmy.Commander = commander
	commander.AssignedArmyID = landArmy.ID
}

// ReleaseEmbarkedCommander taşınan kara birlikleri kaybolduğunda komutanı havuza bırakır.
func (s *GameState) ReleaseEmbarkedCommander(fleetID army.ArmyID) {
	if s == nil {
		return
	}
	fleet := s.Armies[fleetID]
	if fleet == nil || fleet.EmbarkedCommander == nil {
		return
	}
	commander := fleet.EmbarkedCommander
	fleet.EmbarkedCommander = nil
	if commander.AssignedArmyID == fleet.ID {
		commander.AssignedArmyID = ""
	}
}

// AmphibiousCommander çıkarma savaşında kullanılacak kara komutanını döner.
// Filo komutanı yalnızca donanmayı yönetir; taşıma göreviyle ilgisi olmayan
// filo komutanı çıkarma savaşına veya kara ordusunun bonuslarına katılmaz.
func (s *GameState) AmphibiousCommander(fleetID army.ArmyID) *army.Commander {
	if s == nil {
		return nil
	}
	fleet := s.Armies[fleetID]
	if fleet == nil {
		return nil
	}
	if fleet.EmbarkedCommander != nil {
		return fleet.EmbarkedCommander
	}
	return nil
}

// InitializePlayerCommanders mevcut komutanları korur ve oyuncu için ilk boş havuzu oluşturur.
func (s *GameState) InitializePlayerCommanders() {
	if s == nil || s.PlayerFactionID == "" {
		return
	}
	s.ensureCommanderPool(string(s.PlayerFactionID), InitialPlayerCommanderPool)
}

func (s *GameState) ensureCommanderPool(ownerID string, desired int) {
	if s == nil || ownerID == "" || desired <= 0 {
		return
	}
	s.SyncCommanderLinks()
	owned := 0
	for _, commander := range s.Commanders {
		if commander != nil && commander.OwnerID == ownerID {
			s.repairGeneratedCommanderPortrait(commander)
			owned++
		}
	}
	for owned < desired {
		if template := s.nextCommanderTemplate(ownerID); template != nil {
			s.Commanders[template.ID] = cloneCommanderTemplate(template, ownerID)
			owned++
			continue
		}

		s.NextCommanderSeq++
		id := fmt.Sprintf("commander_%s_%d", ownerID, s.NextCommanderSeq)
		name := fmt.Sprintf("Komutan %d", s.NextCommanderSeq)
		if ownerID == string(s.PlayerFactionID) && owned < len(initialCommanderNames) {
			name = initialCommanderNames[owned]
		}
		s.Commanders[id] = &army.Commander{
			ID:            id,
			OwnerID:       ownerID,
			Name:          name,
			PortraitAsset: army.DefaultPortraitAsset,
			Level:         army.CommanderStartingLevel,
		}
		owned++
	}
}

func (s *GameState) nextCommanderTemplate(ownerID string) *army.Commander {
	if s == nil || s.CommanderTemplates == nil {
		return nil
	}
	for _, template := range s.CommanderTemplates[ownerID] {
		if template == nil || template.ID == "" {
			continue
		}
		if _, exists := s.Commanders[template.ID]; !exists {
			return template
		}
	}
	return nil
}

func cloneCommanderTemplate(source *army.Commander, ownerID string) *army.Commander {
	if source == nil {
		return nil
	}
	clone := *source
	clone.OwnerID = ownerID
	clone.AssignedArmyID = ""
	clone.Traits = append([]army.CommanderTrait(nil), source.Traits...)
	clone.Normalize()
	return &clone
}

func (s *GameState) repairGeneratedCommanderPortrait(commander *army.Commander) {
	if s == nil || commander == nil || commander.PortraitAsset != "" {
		return
	}
	if s.isTemplateCommander(commander.OwnerID, commander.ID) {
		return
	}
	commander.PortraitAsset = army.DefaultPortraitAsset
}

func (s *GameState) isTemplateCommander(ownerID, commanderID string) bool {
	if s == nil || ownerID == "" || commanderID == "" || s.CommanderTemplates == nil {
		return false
	}
	for _, template := range s.CommanderTemplates[ownerID] {
		if template != nil && template.ID == commanderID {
			return true
		}
	}
	return false
}

// EnsureFactionCommanders AI fraksiyonunun aktif ordularına deterministik şekilde
// komutan üretir ve boşta kalan komutanları atar. Garnizonlar sahra ordusuna
// dönüşene kadar komutan tüketmez; deniz filoları ise savaşabildiği için kapsamdadır.
func (s *GameState) EnsureFactionCommanders(ownerID string) {
	if s == nil || ownerID == "" {
		return
	}
	s.SyncCommanderLinks()
	armyIDs := make([]army.ArmyID, 0)
	for aid, currentArmy := range s.Armies {
		if currentArmy == nil || currentArmy.OwnerID != ownerID || currentArmy.IsGarrison {
			continue
		}
		armyIDs = append(armyIDs, aid)
	}
	sort.Slice(armyIDs, func(i, j int) bool { return armyIDs[i] < armyIDs[j] })
	s.ensureCommanderPool(ownerID, len(armyIDs))
	available := s.AvailableCommanders(ownerID)
	for _, aid := range armyIDs {
		currentArmy := s.Armies[aid]
		if currentArmy == nil || currentArmy.Commander != nil || len(available) == 0 {
			continue
		}
		commander := available[0]
		if !s.AssignCommanderToArmy(commander.ID, aid) {
			continue
		}
		available = available[1:]
	}
}

// AvailableCommanders boşta olan komutanları deterministik sırada döner.
func (s *GameState) AvailableCommanders(ownerID string) []*army.Commander {
	if s == nil || ownerID == "" {
		return nil
	}
	s.SyncCommanderLinks()
	available := make([]*army.Commander, 0, len(s.Commanders))
	for _, commander := range s.Commanders {
		if commander == nil || commander.OwnerID != ownerID || commander.AssignedArmyID != "" {
			continue
		}
		available = append(available, commander)
	}
	sort.Slice(available, func(i, j int) bool {
		if available[i].Name == available[j].Name {
			return available[i].ID < available[j].ID
		}
		return available[i].Name < available[j].Name
	})
	return available
}

// AssignCommanderToArmy tekil bir komutanı orduya atar.
func (s *GameState) AssignCommanderToArmy(commanderID string, armyID army.ArmyID) bool {
	if s == nil || armyID == "" {
		return false
	}
	s.SyncCommanderLinks()
	currentArmy := s.Armies[armyID]
	commander := s.Commanders[commanderID]
	if currentArmy == nil || commander == nil || commander.OwnerID != currentArmy.OwnerID {
		return false
	}
	if commander.AssignedArmyID != "" && commander.AssignedArmyID != armyID {
		return false
	}
	if currentArmy.Commander != nil && currentArmy.Commander != commander {
		currentArmy.Commander.AssignedArmyID = ""
	}
	commander.AssignedArmyID = armyID
	currentArmy.Commander = commander
	return true
}

// UnassignCommanderFromArmy komutanı ordudan ayırıp havuza geri bırakır.
func (s *GameState) UnassignCommanderFromArmy(armyID army.ArmyID) bool {
	if s == nil {
		return false
	}
	currentArmy := s.Armies[armyID]
	if currentArmy == nil || currentArmy.Commander == nil {
		return false
	}
	currentArmy.Commander.AssignedArmyID = ""
	currentArmy.Commander = nil
	return true
}
