package diplomacy

type ActionDef struct {
	Action     Action
	LabelTR    string
	Quick      bool
	Contextual bool
}

var actionDefs = []ActionDef{
	{Action: ActionDeclareWar, LabelTR: "Savaş", Quick: true},
	{Action: ActionJoinWarCall, LabelTR: "Savaşa Katılım", Contextual: true},
	{Action: ActionProposePeace, LabelTR: "Barış", Quick: true},
	{Action: ActionProposeAlliance, LabelTR: "İttifak", Quick: true},
	{Action: ActionProposeTrade, LabelTR: "Ticaret", Quick: true},
	{Action: ActionProposeSurrender, LabelTR: "Teslimiyet", Contextual: true},
	{Action: ActionProposeSiegeVassalization, LabelTR: "Kuşatma Vassallığı", Contextual: true},
	{Action: ActionCancelAlliance, LabelTR: "İttifakı Bitir", Contextual: true},
	{Action: ActionCancelTrade, LabelTR: "Ticareti Bitir", Contextual: true},
	{Action: ActionImproveRelations, LabelTR: "Heyet"},
	{Action: ActionSendGift, LabelTR: "Hediye"},
	{Action: ActionOfferVassalization, LabelTR: "Vassallık"},
	{Action: ActionReleaseVassal, LabelTR: "Vasallığı Bitir", Contextual: true},
	{Action: ActionAnnexVassal, LabelTR: "İlhak Et", Contextual: true},
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
		if def.Contextual {
			continue
		}
		out = append(out, def.Action)
	}
	return out
}

func VassalManagementActions() []Action {
	out := make([]Action, 0, 2)
	for _, def := range actionDefs {
		if def.Action == ActionReleaseVassal || def.Action == ActionAnnexVassal {
			out = append(out, def.Action)
		}
	}
	return out
}

func QuickActions() []Action {
	out := make([]Action, 0, len(actionDefs))
	for _, def := range actionDefs {
		if def.Quick {
			out = append(out, def.Action)
		}
	}
	return out
}

func ActionLabelTR(action Action) string {
	if def, ok := actionDefsByValue[action]; ok {
		return def.LabelTR
	}
	return string(action)
}
