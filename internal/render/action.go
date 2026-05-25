package render

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// ActionKind renderer'dan gelen oyun eyleminin türü.
type ActionKind string

const (
	ActionNone              ActionKind = ""
	ActionEndTurn           ActionKind = "end_turn"
	ActionConfirmEndTurn    ActionKind = "confirm_end_turn"
	ActionSelectArmy        ActionKind = "select_army"
	ActionMoveArmy          ActionKind = "move_army"
	ActionRecruitUnit       ActionKind = "recruit_unit"
	ActionBuild             ActionKind = "build"
	ActionDeclareWar        ActionKind = "declare_war"
	ActionProposePeace      ActionKind = "propose_peace"
	ActionSave              ActionKind = "save"
	ActionLoad              ActionKind = "load"
	ActionSelectFaction     ActionKind = "select_faction"
	ActionAdjustTax         ActionKind = "adjust_tax"      // Delta: +5 veya -5
	ActionResearch          ActionKind = "research"        // BuildingID = tech ID
	ActionCancelResearch    ActionKind = "cancel_research" // teknoloji araştırmasını iptal et
	ActionSelectVictory     ActionKind = "select_victory"  // BuildingID = VictoryType
	ActionProposeAlliance   ActionKind = "propose_alliance"
	ActionProposeTrade      ActionKind = "propose_trade"
	ActionRecruitNaval      ActionKind = "recruit_naval"
	ActionRecruitSpecific   ActionKind = "recruit_specific"     // BuildingID = unit type ID
	ActionDeclareWarAndMove ActionKind = "declare_war_and_move" // savaş ilan et + orduyu taşı
	// Ana menü
	ActionNewGame                   ActionKind = "new_game"
	ActionContinue                  ActionKind = "continue"
	ActionOpenSettings              ActionKind = "open_settings"
	ActionQuit                      ActionKind = "quit"
	ActionSaveSettings              ActionKind = "save_settings"
	ActionBack                      ActionKind = "back"
	ActionResume                    ActionKind = "resume"                         // duraklama menüsünden devam
	ActionGoMainMenu                ActionKind = "go_main_menu"                   // oyundan ana menüye dön
	ActionLoadFromPause             ActionKind = "load_from_pause"                // duraklama menüsünden yükle
	ActionOpenPauseMenu             ActionKind = "open_pause_menu"                // duraklama menüsünü aç
	ActionOpenLoadSelect            ActionKind = "open_load_select"               // kayıt seçim ekranını aç
	ActionOpenSaveSelect            ActionKind = "open_save_select"               // slot seçerek kaydetme ekranını aç
	ActionSelectSave                ActionKind = "select_save"                    // belirli slotu yükle/kaydet (BuildingID = slot adı)
	ActionDeleteSave                ActionKind = "delete_save"                    // belirli slotu sil (BuildingID = slot adı)
	ActionSplitArmy                 ActionKind = "split_army"                     // seçili orduyu ikiye böl
	ActionMergeArmies               ActionKind = "merge_armies"                   // iki orduyu birleştir (ArmyID + TargetRegion içindeki diğeri)
	ActionSelectScenario            ActionKind = "select_scenario"                // BuildingID = senaryo klasör yolu
	ActionSaveScenario              ActionKind = "save_scenario"                  // edit mode: aktif senaryo JSON verisini kaydet
	ActionSaveScenarioAndGoMainMenu ActionKind = "save_scenario_and_go_main_menu" // edit mode: kaydet ve ana menüye dön
	ActionToggleMusic               ActionKind = "toggle_music"
	ActionNextMusic                 ActionKind = "next_music"
	ActionAdjustMusic               ActionKind = "adjust_music" // Delta: müzik ses seviyesi

	// Ticaret paneli
	ActionOpenTradeView    ActionKind = "open_trade_view"
	ActionCloseTradeView   ActionKind = "close_trade_view"
	ActionCreateTradeRoute ActionKind = "create_trade_route" // BuildingID = mal tipi, TargetFaction = hedef, Delta = miktar
	ActionCancelTradeRoute ActionKind = "cancel_trade_route" // BuildingID = rota indeksi
	ActionOneTimeTrade     ActionKind = "one_time_trade"     // BuildingID = mal tipi, Delta = miktar
	ActionTradeScroll      ActionKind = "trade_scroll"       // Delta: +1/-1
	ActionTradeTabSwitch   ActionKind = "trade_tab_switch"   // Delta: hangi sekme
	ActionRespondDiplomacyOffer ActionKind = "respond_diplomacy_offer"
)

// InputAction'da BuildingID bina inşa işlemleri için kullanılır.
// TargetFaction diplomasi işlemleri için kullanılır.

// InputAction renderer'ın bir çerçevede ürettiği tek oyun eylemi.
type InputAction struct {
	Kind          ActionKind
	ArmyID        army.ArmyID
	TargetRegion  world.RegionID
	BuildingID    string
	TargetFaction faction.FactionID
	Delta         int // AdjustTax için: +5 veya -5
	OfferIndex    int
	OfferAccepted bool
}
