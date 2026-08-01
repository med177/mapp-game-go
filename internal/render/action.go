package render

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/combat"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// ActionKind renderer'dan gelen oyun eyleminin türü.
type ActionKind string

const (
	ActionNone                     ActionKind = ""
	ActionEndTurn                  ActionKind = "end_turn"
	ActionConfirmEndTurn           ActionKind = "confirm_end_turn"
	ActionSelectArmy               ActionKind = "select_army"
	ActionMoveArmy                 ActionKind = "move_army"
	ActionEmbarkArmy               ActionKind = "embark_army"
	ActionDisembarkArmy            ActionKind = "disembark_army"
	ActionAssignMerchantRoute      ActionKind = "assign_merchant_route"
	ActionClearMerchantRoute       ActionKind = "clear_merchant_route"
	ActionAssignNavalMission       ActionKind = "assign_naval_mission"
	ActionClearNavalMission        ActionKind = "clear_naval_mission"
	ActionStartSiege               ActionKind = "start_siege"
	ActionAssaultSiege             ActionKind = "assault_siege"
	ActionLiftSiege                ActionKind = "lift_siege"
	ActionProposeSiegeSurrender    ActionKind = "propose_siege_surrender"
	ActionSortieSiege              ActionKind = "sortie_siege"
	ActionSurrenderSiege           ActionKind = "surrender_siege"
	ActionRecruitUnit              ActionKind = "recruit_unit"
	ActionBuild                    ActionKind = "build"
	ActionDeclareWar               ActionKind = "declare_war"
	ActionProposePeace             ActionKind = "propose_peace"
	ActionImproveRelations         ActionKind = "improve_relations"
	ActionSendGift                 ActionKind = "send_gift"
	ActionGrainAid                 ActionKind = "grain_aid"
	ActionOfferVassalization       ActionKind = "offer_vassalization"
	ActionReleaseVassal            ActionKind = "release_vassal"
	ActionAnnexVassal              ActionKind = "annex_vassal"
	ActionSave                     ActionKind = "save"
	ActionLoad                     ActionKind = "load"
	ActionSelectFaction            ActionKind = "select_faction"
	ActionAdjustTax                ActionKind = "adjust_tax"      // Delta: +5 veya -5
	ActionResearch                 ActionKind = "research"        // BuildingID = tech ID
	ActionCancelResearch           ActionKind = "cancel_research" // teknoloji araştırmasını iptal et
	ActionCancelBuilding           ActionKind = "cancel_building" // bina inşaatını onaylı iptal
	ActionSelectVictory            ActionKind = "select_victory"  // BuildingID = VictoryType
	ActionProposeAlliance          ActionKind = "propose_alliance"
	ActionProposeTrade             ActionKind = "propose_trade"
	ActionCancelAlliance           ActionKind = "cancel_alliance"
	ActionCancelTrade              ActionKind = "cancel_trade"
	ActionRecruitNaval             ActionKind = "recruit_naval"
	ActionRecruitSpecific          ActionKind = "recruit_specific"     // BuildingID = unit type ID
	ActionCancelRecruitOrder       ActionKind = "cancel_recruit_order" // BuildingID = production order ID (tek emir iptal)
	ActionDeclareWarAndMove        ActionKind = "declare_war_and_move" // savaş ilan et + orduyu taşı
	ActionAnnexDefeatedFaction     ActionKind = "annex_defeated_faction"
	ActionVassalizeDefeatedFaction ActionKind = "vassalize_defeated_faction"
	ActionLiberateSuccessor        ActionKind = "liberate_successor"
	// Ardıl devlet kararları, genel vassal yönetiminden ve son düşman
	// toprağına ilişkin iki seçenekli karardan ayrı tutulur.
	ActionAnnexSuccessor     ActionKind = "annex_successor"
	ActionVassalizeSuccessor ActionKind = "vassalize_successor"
	ActionReleaseSuccessor   ActionKind = "release_successor"
	// Ana menü
	ActionNewGame                   ActionKind = "new_game"
	ActionEditMode                  ActionKind = "edit_mode"
	ActionContinue                  ActionKind = "continue"
	ActionOpenSettings              ActionKind = "open_settings"
	ActionQuit                      ActionKind = "quit"
	ActionSaveSettings              ActionKind = "save_settings"
	ActionBack                      ActionKind = "back"
	ActionResume                    ActionKind = "resume"           // duraklama menüsünden devam
	ActionGoMainMenu                ActionKind = "go_main_menu"     // oyundan ana menüye dön
	ActionLoadFromPause             ActionKind = "load_from_pause"  // duraklama menüsünden yükle
	ActionOpenPauseMenu             ActionKind = "open_pause_menu"  // duraklama menüsünü aç
	ActionOpenLoadSelect            ActionKind = "open_load_select" // kayıt seçim ekranını aç
	ActionOpenSaveSelect            ActionKind = "open_save_select" // slot seçerek kaydetme ekranını aç
	ActionSelectSave                ActionKind = "select_save"      // belirli slotu yükle/kaydet (BuildingID = slot adı)
	ActionDeleteSave                ActionKind = "delete_save"      // belirli slotu sil (BuildingID = slot adı)
	ActionSplitArmy                 ActionKind = "split_army"       // seçili orduyu ikiye böl
	ActionMergeArmies               ActionKind = "merge_armies"     // ArmyID'yi TargetArmyID ile birleştir
	ActionAssignCommander           ActionKind = "assign_commander"
	ActionUnassignCommander         ActionKind = "unassign_commander"
	ActionUnassignEmbarkedCommander ActionKind = "unassign_embarked_commander"
	ActionSelectScenario            ActionKind = "select_scenario"                // BuildingID = senaryo klasör yolu
	ActionSaveScenario              ActionKind = "save_scenario"                  // edit mode: aktif senaryo JSON verisini kaydet
	ActionSaveScenarioAndGoMainMenu ActionKind = "save_scenario_and_go_main_menu" // edit mode: kaydet ve ana menüye dön
	ActionToggleMusic               ActionKind = "toggle_music"
	ActionNextMusic                 ActionKind = "next_music"
	ActionAdjustMusic               ActionKind = "adjust_music" // Delta: müzik ses seviyesi

	// Ticaret paneli
	ActionOpenTradeView         ActionKind = "open_trade_view"
	ActionCloseTradeView        ActionKind = "close_trade_view"
	ActionCreateTradeRoute      ActionKind = "create_trade_route"   // BuildingID = mal tipi, TargetFaction = hedef, Delta = miktar
	ActionCancelTradeRoute      ActionKind = "cancel_trade_route"   // BuildingID = rota indeksi
	ActionOneTimeTrade          ActionKind = "one_time_trade"       // BuildingID = mal tipi, Delta = miktar
	ActionEmergencyGrainSale    ActionKind = "emergency_grain_sale" // Delta = satış miktarı
	ActionToggleAutoGrainExport ActionKind = "toggle_auto_grain_export"
	ActionTradeScroll           ActionKind = "trade_scroll"     // Delta: +1/-1
	ActionTradeTabSwitch        ActionKind = "trade_tab_switch" // Delta: hangi sekme
	ActionRespondDiplomacyOffer ActionKind = "respond_diplomacy_offer"
	ActionChooseHistoricalEvent ActionKind = "choose_historical_event"
	ActionOpenEventCodex        ActionKind = "open_event_codex"
	ActionScheduleCapitalMove   ActionKind = "schedule_capital_move"
	ActionOpenImperialPanel     ActionKind = "open_imperial_panel"
	ActionImperialDietChoice    ActionKind = "imperial_diet_choice"
	ActionImperialElectionVote  ActionKind = "imperial_election_vote"
)

// InputAction'da BuildingID bina inşa işlemleri için kullanılır.
// TargetFaction diplomasi işlemleri için kullanılır.

// InputAction renderer'ın bir çerçevede ürettiği tek oyun eylemi.
type InputAction struct {
	Kind         ActionKind
	ArmyID       army.ArmyID
	UnitIndices  []int // ActionSplitArmy için seçilerek ayrılacak fiziksel birim index'leri
	TargetArmyID army.ArmyID
	CommanderID  string
	TargetRegion world.RegionID
	// TargetSettlementID, denizden kara hedefinde liman ile merkez yerleşimi
	// ayırır. Boş bırakıldığında bölge tabanlı eski hareket semantiği korunur.
	TargetSettlementID string
	BuildingID         string
	Quantity           int
	TargetFaction      faction.FactionID
	WarAllies          []faction.FactionID
	Delta              int // AdjustTax için: +5 veya -5
	OfferIndex         int
	OfferAccepted      bool
	ChoiceIndex        int
	BattleStance       combat.BattleStance
}
