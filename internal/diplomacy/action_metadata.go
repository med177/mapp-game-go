package diplomacy

type ActionDef struct {
	Action  Action
	LabelTR string
}

var actionDefs = []ActionDef{
	{Action: ActionDeclareWar, LabelTR: "Savaş"},
	{Action: ActionProposePeace, LabelTR: "Barış"},
	{Action: ActionProposeAlliance, LabelTR: "İttifak"},
	{Action: ActionProposeTrade, LabelTR: "Ticaret"},
}

var actionDefsByValue = func() map[Action]ActionDef {
	out := make(map[Action]ActionDef, len(actionDefs))
	for _, def := range actionDefs {
		out[def.Action] = def
	}
	return out
}()

func VisibleActions() []Action {
	out := make([]Action, 0, len(actionDefs))
	for _, def := range actionDefs {
		out = append(out, def.Action)
	}
	return out
}

func ActionLabelTR(action Action) string {
	if def, ok := actionDefsByValue[action]; ok {
		return def.LabelTR
	}
	return string(action)
}
