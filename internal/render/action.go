package render

import (
	"mapp-game-go/internal/army"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/world"
)

// ActionKind renderer'dan gelen oyun eyleminin türü.
type ActionKind string

const (
	ActionNone           ActionKind = ""
	ActionEndTurn        ActionKind = "end_turn"
	ActionSelectArmy     ActionKind = "select_army"
	ActionMoveArmy       ActionKind = "move_army"
	ActionRecruitUnit    ActionKind = "recruit_unit"
	ActionBuild          ActionKind = "build"
	ActionDeclareWar     ActionKind = "declare_war"
	ActionProposePeace   ActionKind = "propose_peace"
	ActionSave           ActionKind = "save"
	ActionLoad           ActionKind = "load"
	ActionSelectFaction  ActionKind = "select_faction"
	ActionAdjustTax      ActionKind = "adjust_tax"     // Delta: +5 veya -5
	ActionResearch       ActionKind = "research"       // BuildingID = tech ID
	ActionSelectVictory  ActionKind = "select_victory" // BuildingID = VictoryType
	ActionProposeAlliance ActionKind = "propose_alliance"
	ActionProposeTrade    ActionKind = "propose_trade"
	ActionRecruitNaval   ActionKind = "recruit_naval"
	// Ana menü
	ActionNewGame       ActionKind = "new_game"
	ActionContinue      ActionKind = "continue"
	ActionOpenSettings  ActionKind = "open_settings"
	ActionQuit          ActionKind = "quit"
	ActionSaveSettings  ActionKind = "save_settings"
	ActionBack          ActionKind = "back"
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
}
