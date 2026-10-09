package render

import (
	"image/color"
	"sort"
	"strconv"

	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/state"
	"mapp-game-go/internal/world"

	gameui "mapp-game-go/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	negotiationKindRegion = iota
	negotiationKindResource
	negotiationKindArmy

	negotiationSelectedItemsHeight = 220.0
)

type negotiationSideState struct {
	kind           int
	option         int
	amount         int
	regionDropdown *gameui.Dropdown
	selectedList   gameui.ListView
	items          []state.DiplomaticTransfer
}

type negotiationPanelState struct {
	show         bool
	target       faction.FactionID
	requested    negotiationSideState
	offered      negotiationSideState
	counterIndex int
}

type negotiationOption struct {
	id       string
	label    string
	max      int
	isRegion bool
}

type negotiationSideLayout struct {
	panel  gameui.Rect
	kind   gameui.Button
	option gameui.Button
	minus  gameui.Button
	plus   gameui.Button
	amount gameui.Rect
	add    gameui.Button
}

type negotiationLayout struct {
	modal  gameui.Modal
	left   negotiationSideLayout
	right  negotiationSideLayout
	submit gameui.Button
	cancel gameui.Button
}

var negotiationStepperButtonStyle = func() gameui.ButtonStyle {
	style := tinyButtonStyle
	style.CornerRadius = 8
	return style
}()

func buildNegotiationLayout() negotiationLayout {
	w := minF(ScreenWidth-64, 1040)
	h := minF(ScreenHeight-28, 820)
	panelRect := gameui.AnchorRect(gameui.Rect{W: ScreenWidth, H: ScreenHeight}, w, h, gameui.AnchorCenter, gameui.AnchorMiddle, 0, 0)
	panel := gameui.NewPanel(panelRect.X, panelRect.Y, panelRect.W, panelRect.H)
	modal := gameui.NewModal(ScreenWidth, ScreenHeight, panel)
	columnW := (w - 56) / 2
	columnY := panelRect.Y + 76
	columnH := h - 156
	left := buildNegotiationSideLayout(gameui.Rect{X: panelRect.X + 18, Y: columnY, W: columnW, H: columnH})
	right := buildNegotiationSideLayout(gameui.Rect{X: panelRect.X + 38 + columnW, Y: columnY, W: columnW, H: columnH})
	footerY := panelRect.Y + h - 58
	return negotiationLayout{
		modal:  modal,
		left:   left,
		right:  right,
		submit: gameui.NewButton(panelRect.X+w-276, footerY, 132, 38, "Pazarlığı Gönder").WithIcon(gameui.IconSend),
		cancel: gameui.NewButton(panelRect.X+w-136, footerY, 118, 38, "Kapat").WithIcon(gameui.IconClose),
	}
}

func buildNegotiationSideLayout(panel gameui.Rect) negotiationSideLayout {
	// Tür ve seçim aynı satırda; miktar ve ekleme satırı bunun altında sabit kalır.
	controlY := negotiationControlY(panel)
	amountY := controlY + 36
	innerW := panel.W - 24
	buttonGap := 8.0
	kindW := (innerW - buttonGap) / 2
	kindX := panel.X + 12
	optionX := kindX + kindW + buttonGap
	minusX := panel.X + 12
	plusX := minusX + 32
	amountX := plusX + 34
	addW := minF(180, panel.W-190)
	addX := panel.X + panel.W - 12 - addW
	return negotiationSideLayout{
		panel:  panel,
		kind:   gameui.NewButton(kindX, controlY, kindW, 30, "Tür"),
		option: gameui.NewButton(optionX, controlY, kindW, 30, "Seçim"),
		minus:  gameui.NewButton(minusX, amountY+2, 26, 26, "").WithIcon(gameui.IconMinus),
		plus:   gameui.NewButton(plusX, amountY+2, 26, 26, "").WithIcon(gameui.IconPlus),
		amount: gameui.Rect{X: amountX, Y: amountY, W: 92, H: 30},
		add:    gameui.NewButton(addX, amountY, addW, 30, "Kalemi Ekle").WithIcon(gameui.IconCheck),
	}
}

func negotiationControlY(panel gameui.Rect) float64 {
	return panel.Y + panel.H - 72
}

func negotiationHasOptions(gs *state.GameState, owner faction.FactionID, kind int) bool {
	if gs == nil {
		return false
	}
	switch kind {
	case negotiationKindRegion:
		for _, region := range gs.Regions {
			if region != nil && !region.IsSea && region.OwnerID == string(owner) {
				return true
			}
		}
	case negotiationKindResource:
		f := gs.Factions[owner]
		for _, resource := range economy.AllResourceKinds() {
			if economy.FactionResourceAmount(f, resource) > 0 {
				return true
			}
		}
	case negotiationKindArmy:
		for _, current := range gs.Armies {
			if current == nil || current.OwnerID != string(owner) || current.IsNaval || current.Commander != nil || current.EmbarkedCommander != nil {
				continue
			}
			if len(current.Units) > 0 {
				return true
			}
		}
	}
	return false
}

func configureNegotiationSideControls(gs *state.GameState, owner faction.FactionID, side *negotiationSideState, layout *negotiationSideLayout) {
	if side == nil || layout == nil {
		return
	}
	hasOptions := negotiationHasOptions(gs, owner, side.kind)
	layout.option.Enabled = hasOptions
	layout.minus.Enabled = hasOptions && side.kind != negotiationKindRegion
	layout.plus.Enabled = hasOptions && side.kind != negotiationKindRegion
	layout.add.Enabled = hasOptions
}

func negotiationSelectedItemHit(side *negotiationSideState, panel gameui.Rect, mx, my float64) bool {
	if side == nil || len(side.items) == 0 {
		return false
	}
	list := negotiationSelectedItemsListGeometry(panel)
	maxScroll := len(side.items) - list.VisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	list.Scroll = side.selectedList.Scroll
	if list.Scroll < 0 {
		list.Scroll = 0
	}
	if list.Scroll > maxScroll {
		list.Scroll = maxScroll
	}
	return list.ItemIndexAtCount(mx, my, len(side.items)) >= 0
}

func (r *Renderer) diplomacyNegotiationPointerHit(mx, my float64) bool {
	if r == nil || r.gs == nil || !r.negotiation.show {
		return false
	}
	layout := buildNegotiationLayout()
	layout.submit.Enabled = len(r.negotiation.requested.items)+len(r.negotiation.offered.items) > 0
	if layout.cancel.HitTest(mx, my) || (layout.submit.Enabled && layout.submit.HitTest(mx, my)) {
		return true
	}
	return negotiationSidePointerHit(r.gs, r.negotiation.target, &r.negotiation.requested, layout.left, mx, my) ||
		negotiationSidePointerHit(r.gs, r.gs.PlayerFactionID, &r.negotiation.offered, layout.right, mx, my)
}

func negotiationSidePointerHit(gs *state.GameState, owner faction.FactionID, side *negotiationSideState, layout negotiationSideLayout, mx, my float64) bool {
	configureNegotiationSideControls(gs, owner, side, &layout)
	if layout.kind.HitTest(mx, my) ||
		(layout.option.Enabled && layout.option.HitTest(mx, my)) ||
		(layout.minus.Enabled && layout.minus.HitTest(mx, my)) ||
		(layout.plus.Enabled && layout.plus.HitTest(mx, my)) ||
		(layout.add.Enabled && layout.add.HitTest(mx, my)) ||
		negotiationSelectedItemHit(side, layout.panel, mx, my) {
		return true
	}
	if side != nil && side.regionDropdown != nil && side.regionDropdown.IsOpen() {
		if _, ok := side.regionDropdown.GetSelectedOption(mx, my); ok {
			return true
		}
	}
	return false
}

func (r *Renderer) openDiplomacyNegotiation(target faction.FactionID) {
	if r == nil || r.gs == nil || target == "" || target == r.gs.PlayerFactionID {
		return
	}
	r.negotiation = negotiationPanelState{
		show:         true,
		target:       target,
		counterIndex: -1,
		requested:    negotiationSideState{kind: negotiationKindRegion, amount: 1},
		offered:      negotiationSideState{kind: negotiationKindRegion, amount: 1},
	}
}

// OpenDiplomacyCounterOffer, bekleyen transfer teklifini ters yönde doldurup
// oyuncuya karşı teklif düzenleme ekranı açar.
func (r *Renderer) OpenDiplomacyCounterOffer(index int) {
	if r == nil || r.gs == nil || index < 0 || index >= len(r.gs.DiplomaticOffers) {
		return
	}
	offer := r.gs.DiplomaticOffers[index]
	if offer.Action != string(diplomacyActionTransfer()) || offer.ToFactionID != r.gs.PlayerFactionID {
		return
	}
	r.showDiplomacy = false
	r.negotiation = negotiationPanelState{
		show:         true,
		target:       offer.FromFactionID,
		counterIndex: index,
		requested:    negotiationSideState{kind: negotiationKindRegion, amount: 1, items: cloneNegotiationTransfers(offer.OfferedTransfers)},
		offered:      negotiationSideState{kind: negotiationKindRegion, amount: 1, items: cloneNegotiationTransfers(offer.RequestedTransfers)},
	}
}

func diplomacyActionTransfer() string {
	return "propose_transfer"
}

func (r *Renderer) CloseDiplomacyNegotiation() {
	if r == nil {
		return
	}
	r.negotiation = negotiationPanelState{}
	r.showDiplomacy = false
	r.diplomacyTargetFaction = ""
}

func negotiationFactionName(gs *state.GameState, id faction.FactionID) string {
	name := factionDisplayName(gs, string(id))
	if name == "" {
		return string(id)
	}
	return name
}

func negotiationKindLabel(kind int) string {
	switch kind {
	case negotiationKindResource:
		return "Kaynak / Altın"
	case negotiationKindArmy:
		return "Asker"
	default:
		return "Bölge"
	}
}

func negotiationOptions(gs *state.GameState, owner faction.FactionID, kind int) []negotiationOption {
	if gs == nil {
		return nil
	}
	options := make([]negotiationOption, 0)
	switch kind {
	case negotiationKindRegion:
		for _, region := range gs.LandRegionsOwnedBy(owner) {
			if region == nil {
				continue
			}
			label := region.NameTR
			if label == "" {
				label = string(region.ID)
			}
			options = append(options, negotiationOption{id: string(region.ID), label: label, max: 1, isRegion: true})
		}
		sort.Slice(options, func(i, j int) bool {
			if options[i].label == options[j].label {
				return options[i].id < options[j].id
			}
			return options[i].label < options[j].label
		})
	case negotiationKindResource:
		f := gs.Factions[owner]
		for _, kind := range economy.AllResourceKinds() {
			amount := economy.FactionResourceAmount(f, kind)
			if amount <= 0 {
				continue
			}
			options = append(options, negotiationOption{id: string(kind), label: economy.ResourceNameTR(kind), max: amount})
		}
	case negotiationKindArmy:
		counts := make(map[string]int)
		for _, current := range gs.Armies {
			if current == nil || current.OwnerID != string(owner) || current.IsNaval || current.Commander != nil || current.EmbarkedCommander != nil {
				continue
			}
			for _, unit := range current.Units {
				counts[unit.TypeID]++
			}
		}
		for id, count := range counts {
			label := id
			if unitType := gs.UnitTypes[id]; unitType != nil && unitType.NameTR != "" {
				label = unitType.NameTR
			}
			options = append(options, negotiationOption{id: id, label: label, max: count})
		}
		sort.Slice(options, func(i, j int) bool { return options[i].id < options[j].id })
	}
	return options
}

func negotiationCurrentOption(gs *state.GameState, owner faction.FactionID, side *negotiationSideState) (negotiationOption, bool) {
	options := negotiationOptions(gs, owner, side.kind)
	if len(options) == 0 {
		return negotiationOption{}, false
	}
	if side.option < 0 {
		side.option = 0
	}
	if side.option >= len(options) {
		side.option = len(options) - 1
	}
	option := options[side.option]
	if side.amount <= 0 {
		side.amount = negotiationDefaultAmount(side.kind, option.max)
	}
	if side.amount > option.max {
		side.amount = option.max
	}
	return option, true
}

func negotiationDefaultAmount(kind, max int) int {
	if kind == negotiationKindRegion {
		return 1
	}
	if kind == negotiationKindArmy {
		if max >= 3 {
			return 3
		}
		return max
	}
	if max >= 10 {
		return 10
	}
	return max
}

func cycleNegotiationPicker(gs *state.GameState, owner faction.FactionID, side *negotiationSideState, kindDelta int) {
	if side == nil || kindDelta == 0 {
		return
	}
	for step := 0; step < 3; step++ {
		side.kind = (side.kind + kindDelta + 3) % 3
		options := negotiationOptions(gs, owner, side.kind)
		if len(options) == 0 {
			continue
		}
		side.option = 0
		if side.regionDropdown != nil {
			side.regionDropdown.Close()
		}
		side.amount = negotiationDefaultAmount(side.kind, options[0].max)
		return
	}
	side.option = 0
	side.amount = 1
}

func cycleNegotiationOption(gs *state.GameState, owner faction.FactionID, side *negotiationSideState) {
	if side == nil {
		return
	}
	options := negotiationOptions(gs, owner, side.kind)
	if len(options) == 0 {
		return
	}
	if side.option < 0 || side.option >= len(options) {
		side.option = 0
	} else {
		side.option = (side.option + 1) % len(options)
	}
	side.amount = negotiationDefaultAmount(side.kind, options[side.option].max)
}

func adjustNegotiationAmount(gs *state.GameState, owner faction.FactionID, side *negotiationSideState, delta int) {
	option, ok := negotiationCurrentOption(gs, owner, side)
	if !ok || option.isRegion {
		return
	}
	step := 1
	if side.kind == negotiationKindResource {
		step = 10
	}
	side.amount += delta * step
	if side.amount < 1 {
		side.amount = 1
	}
	if side.amount > option.max {
		side.amount = option.max
	}
}

func addNegotiationItem(gs *state.GameState, owner faction.FactionID, side *negotiationSideState) {
	if len(side.items) >= 6 {
		return
	}
	option, ok := negotiationCurrentOption(gs, owner, side)
	if !ok {
		return
	}
	for i := range side.items {
		if side.items[i].Kind == negotiationTransferKind(side.kind) && side.items[i].ID == option.id {
			side.items[i].Amount += side.amount
			if side.items[i].Amount > option.max {
				side.items[i].Amount = option.max
			}
			return
		}
	}
	side.items = append(side.items, state.DiplomaticTransfer{Kind: negotiationTransferKind(side.kind), ID: option.id, Amount: side.amount})
}

func negotiationTransferKind(kind int) string {
	switch kind {
	case negotiationKindResource:
		return "resource"
	case negotiationKindArmy:
		return "army"
	default:
		return "region"
	}
}

func cloneNegotiationTransfers(items []state.DiplomaticTransfer) []state.DiplomaticTransfer {
	return append([]state.DiplomaticTransfer(nil), items...)
}

func negotiationSelectedItemsListRect(panel gameui.Rect) gameui.Rect {
	return negotiationSelectedItemsListGeometry(panel).ContentRect()
}

func negotiationSelectedItemsFrameRect(panel gameui.Rect) gameui.Rect {
	return gameui.Rect{X: panel.X + 12, Y: panel.Y + 58, W: panel.W - 24, H: negotiationSelectedItemsHeight + 32}
}

func negotiationSelectedItemsListGeometry(panel gameui.Rect) gameui.ListView {
	list := gameui.ListView{
		Rect:          negotiationSelectedItemsFrameRect(panel),
		ContentInsets: gameui.ListViewInsets{Top: 8, Right: 8, Bottom: 24, Left: 8},
		RowHeight:     28,
		VisibleRows:   int(negotiationSelectedItemsHeight) / 28,
		Enabled:       true,
	}
	return list
}

func negotiationSelectedItemsList(gs *state.GameState, side *negotiationSideState, panel gameui.Rect) *gameui.ListView {
	if gs == nil || side == nil {
		return nil
	}
	list := &side.selectedList
	geometry := negotiationSelectedItemsListGeometry(panel)
	list.Rect = geometry.Rect
	list.ContentInsets = geometry.ContentInsets
	list.RowHeight = geometry.RowHeight
	list.VisibleRows = geometry.VisibleRows
	list.Items = make([]string, 0, len(side.items))
	for _, item := range side.items {
		label := negotiationItemLabel(gs, item)
		if item.Kind == "region" {
			label += " (1)"
		} else {
			label += " (" + strconv.Itoa(item.Amount) + ")"
		}
		list.Items = append(list.Items, label)
	}
	list.Selected = -1
	list.Enabled = true
	maxScroll := len(list.Items) - list.VisibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if list.Scroll > maxScroll {
		list.Scroll = maxScroll
	}
	if list.Scroll < 0 {
		list.Scroll = 0
	}
	return list
}

func negotiationRegionDropdownRect(panel gameui.Rect) gameui.Rect {
	topBound := panel.Y + 56
	maxBottom := negotiationControlY(panel) - 8
	availableH := maxBottom - topBound
	visibleRows := int((availableH - 38) / 28)
	if visibleRows > 8 {
		visibleRows = 8
	}
	if visibleRows < 1 {
		visibleRows = 1
	}
	h := 28 + float64(visibleRows)*28 + 10
	return gameui.Rect{
		X: panel.X + 12,
		Y: maxBottom - h,
		W: panel.W - 24,
		H: h,
	}
}

func negotiationRegionDropdown(gs *state.GameState, owner faction.FactionID, side *negotiationSideState, panel gameui.Rect) *gameui.Dropdown {
	if gs == nil || side == nil {
		return nil
	}
	options := negotiationOptions(gs, owner, negotiationKindRegion)
	labels := make([]string, 0, len(options))
	for _, option := range options {
		labels = append(labels, option.label)
	}
	if len(labels) == 0 {
		return nil
	}
	if side.option < 0 {
		side.option = 0
	}
	if side.option >= len(labels) && len(labels) > 0 {
		side.option = len(labels) - 1
	}
	selected := ""
	if side.option >= 0 && side.option < len(labels) {
		selected = labels[side.option]
	}
	rect := negotiationRegionDropdownRect(panel)
	visibleRows := int((rect.H - 38) / 28)
	if visibleRows < 1 {
		visibleRows = 1
	}
	if side.regionDropdown == nil {
		side.regionDropdown = gameui.NewDropdown(rect.X, rect.Y, rect.W, rect.H, "Bölgeler — seçmek için tıklayın", 28, 28, visibleRows)
	}
	side.regionDropdown.SetPosition(rect.X, rect.Y)
	if !side.regionDropdown.IsOpen() {
		side.regionDropdown.SetOptions(labels, selected)
	}
	return side.regionDropdown
}

func negotiationRegionListStyle() gameui.ListViewStyle {
	return gameui.ListViewStyle{
		RowBG:            color.RGBA{31, 25, 17, 230},
		SelectedRowBG:    color.RGBA{72, 91, 49, 220},
		TextColor:        ColorWhite,
		SelectedText:     ColorWhite,
		MutedText:        ColorGray,
		RowTextOffsetY:   8,
		TextVariant:      gameui.TextSmall,
		PaginationPrefix: "Bölgeler:",
	}
}

func negotiationSelectedItemListStyle() gameui.ListViewStyle {
	style := negotiationRegionListStyle()
	style.PaginationPrefix = "Kalemler:"
	style.DrawFrame = true
	style.FrameBG = color.RGBA{12, 11, 9, 220}
	style.FrameBorder = color.RGBA{83, 67, 39, 210}
	style.FrameBorderWidth = 1
	style.EmptyText = "Henüz kalem eklenmedi."
	style.EmptyTextColor = ColorGray
	style.EmptyTextVariant = gameui.TextSmall
	style.EmptyTextAlign = gameui.TextAlignStart
	return style
}

func negotiationItemLabel(gs *state.GameState, item state.DiplomaticTransfer) string {
	switch item.Kind {
	case "region":
		if region := gs.Regions[world.RegionID(item.ID)]; region != nil {
			if region.NameTR != "" {
				return "Bölge: " + region.NameTR
			}
		}
		return "Bölge: " + item.ID
	case "resource":
		return economy.ResourceNameTR(economy.ResourceKind(item.ID))
	case "army":
		if unitType := gs.UnitTypes[item.ID]; unitType != nil && unitType.NameTR != "" {
			return unitType.NameTR
		}
		return item.ID
	default:
		return item.ID
	}
}

func drawNegotiationSide(screen *ebiten.Image, gs *state.GameState, layout negotiationSideLayout, side *negotiationSideState, owner faction.FactionID, title string, accent color.RGBA) {
	drawUIPanelFrame(screen, layout.panel, color.RGBA{18, 14, 10, 242}, accent, 1.5, 4)
	drawUILabel(screen, gameui.Rect{X: layout.panel.X + 12, Y: layout.panel.Y + 8, W: layout.panel.W - 24}, title, accent, gameui.TextMedium, gameui.TextAlignStart)
	drawUILabel(screen, gameui.Rect{X: layout.panel.X + 12, Y: layout.panel.Y + 32, W: layout.panel.W - 24}, negotiationFactionName(gs, owner), ColorWhite, gameui.TextSmall, gameui.TextAlignStart)

	if selectedList := negotiationSelectedItemsList(gs, side, layout.panel); selectedList != nil {
		gameui.DrawListView(screen, *selectedList, negotiationSelectedItemListStyle(), renderText)
	}
	option, ok := negotiationCurrentOption(gs, owner, side)
	kindLabel := negotiationKindLabel(side.kind)
	if !ok {
		kindLabel += " (uygun kalem yok)"
	}
	layout.kind.Label = "Tür: " + kindLabel
	layout.option.Label = "Seçim: " + option.label
	if !ok {
		layout.option.Label = "Seçim: Yok"
	}
	layout.option.Label = trimTextToWidth(layout.option.Label, FaceSmall, layout.option.W-18)
	var regionDropdown *gameui.Dropdown
	if side.kind == negotiationKindRegion {
		regionDropdown = negotiationRegionDropdown(gs, owner, side, layout.panel)
	}
	configureNegotiationSideControls(gs, owner, side, &layout)
	layout.minus.Enabled = ok && !option.isRegion
	layout.plus.Enabled = ok && !option.isRegion
	drawUIButtonWidget(screen, layout.kind, tinyButtonStyle)
	drawUIButtonWidget(screen, layout.option, tinyButtonStyle)
	drawUIButtonWidget(screen, layout.minus, negotiationStepperButtonStyle)
	drawUIButtonWidget(screen, layout.plus, negotiationStepperButtonStyle)
	amountLabel := "Miktar: -"
	if ok {
		amountLabel = "Miktar: " + strconv.Itoa(side.amount)
	}
	drawUICardRect(screen, layout.amount, color.RGBA{31, 25, 17, 230}, color.RGBA{83, 67, 39, 210}, 1)
	drawUILabel(screen, gameui.Rect{X: layout.amount.X, Y: layout.amount.Y + 9, W: layout.amount.W}, amountLabel, ColorWhite, gameui.TextSmall, gameui.TextAlignCenter)
	drawUIButtonWidget(screen, layout.add, applyTinyButtonStyle)
	if regionDropdown != nil && regionDropdown.IsOpen() {
		gameui.DrawDropdown(screen, regionDropdown, dropdownStyle, renderText)
	}
}

func (r *Renderer) drawDiplomacyNegotiation(screen *ebiten.Image) {
	if r == nil || r.gs == nil || !r.negotiation.show {
		return
	}
	layout := buildNegotiationLayout()
	layout.submit.Enabled = len(r.negotiation.requested.items)+len(r.negotiation.offered.items) > 0
	gameui.DrawModal(screen, layout.modal, standardModalStyle, nil, nil)
	drawUILabel(screen, gameui.Rect{X: layout.modal.Panel.Rect.X + 20, Y: layout.modal.Panel.Rect.Y + 16, W: layout.modal.Panel.Rect.W - 40}, "Diplomatik Pazarlık", ColorGold, gameui.TextLarge, gameui.TextAlignCenter)
	drawUIMutedText(screen, layout.modal.Panel.Rect.X+20, layout.modal.Panel.Rect.Y+46, negotiationFactionName(r.gs, r.negotiation.target)+" ile karşılıklı talep ve teklif oluştur.")
	drawNegotiationSide(screen, r.gs, layout.left, &r.negotiation.requested, r.negotiation.target, "Talep Edilen", color.RGBA{210, 150, 78, 255})
	drawNegotiationSide(screen, r.gs, layout.right, &r.negotiation.offered, r.gs.PlayerFactionID, "Verilen", color.RGBA{88, 174, 112, 255})
	drawUIButtonWidget(screen, layout.submit, solidButtonStyle(color.RGBA{70, 126, 74, 240}, color.RGBA{125, 184, 124, 255}, ColorWhite, 9))
	drawUIButtonWidget(screen, layout.cancel, solidButtonStyle(color.RGBA{78, 68, 56, 240}, color.RGBA{132, 114, 88, 255}, ColorWhite, 9))
}

func handleNegotiationRegionDropdownInput(input gameui.InputState, dropdown *gameui.Dropdown, side *negotiationSideState) bool {
	if dropdown == nil || !dropdown.IsOpen() {
		return false
	}
	if input.WheelY != 0 && dropdown.HitTest(input.MouseX, input.MouseY) {
		dropdown.Scroll(input.WheelY)
		return true
	}
	if !input.LeftJustPressed {
		return dropdown.HitTest(input.MouseX, input.MouseY)
	}
	if index, ok := dropdown.GetSelectedOption(input.MouseX, input.MouseY); ok {
		side.option = index
		side.amount = 1
		dropdown.Close()
		return true
	}
	dropdown.Close()
	return true
}

func (r *Renderer) handleDiplomacyNegotiationInput() InputAction {
	if r == nil || r.gs == nil || !r.negotiation.show {
		return InputAction{}
	}
	layout := buildNegotiationLayout()
	layout.submit.Enabled = len(r.negotiation.requested.items)+len(r.negotiation.offered.items) > 0
	configureNegotiationSideControls(r.gs, r.negotiation.target, &r.negotiation.requested, &layout.left)
	configureNegotiationSideControls(r.gs, r.gs.PlayerFactionID, &r.negotiation.offered, &layout.right)
	mxi, myi := ebiten.CursorPosition()
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftWasPressed := r.prevMouse[ebiten.MouseButtonLeft]
	r.prevMouse[ebiten.MouseButtonLeft] = leftPressed
	_, wheelY := ebiten.Wheel()
	input := gameui.InputState{
		MouseX:           float64(mxi),
		MouseY:           float64(myi),
		LeftPressed:      leftPressed,
		LeftJustPressed:  leftPressed && !leftWasPressed,
		LeftJustReleased: !leftPressed && leftWasPressed,
		WheelY:           wheelY,
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		if r.negotiation.requested.regionDropdown != nil && r.negotiation.requested.regionDropdown.IsOpen() {
			r.negotiation.requested.regionDropdown.Close()
			return InputAction{}
		}
		if r.negotiation.offered.regionDropdown != nil && r.negotiation.offered.regionDropdown.IsOpen() {
			r.negotiation.offered.regionDropdown.Close()
			return InputAction{}
		}
		r.negotiation = negotiationPanelState{}
		return InputAction{}
	}
	if layout.cancel.HandleInput(input) {
		r.negotiation = negotiationPanelState{}
		return InputAction{}
	}
	if layout.submit.HandleInput(input) {
		if len(r.negotiation.requested.items) == 0 && len(r.negotiation.offered.items) == 0 {
			return InputAction{}
		}
		return InputAction{
			Kind:                 ActionNegotiateTransfer,
			TargetFaction:        r.negotiation.target,
			OfferIndex:           r.negotiation.counterIndex,
			CounterOffer:         r.negotiation.counterIndex >= 0,
			NegotiationRequested: cloneNegotiationTransfers(r.negotiation.requested.items),
			NegotiationOffered:   cloneNegotiationTransfers(r.negotiation.offered.items),
		}
	}
	requestOwner := r.negotiation.target
	offerOwner := r.gs.PlayerFactionID
	if list := negotiationSelectedItemsList(r.gs, &r.negotiation.requested, layout.left.panel); list != nil {
		handled := list.HandleInput(input)
		if input.LeftJustPressed {
			if index := list.ItemIndexAt(input.MouseX, input.MouseY); index >= 0 {
				r.negotiation.requested.items = append(r.negotiation.requested.items[:index], r.negotiation.requested.items[index+1:]...)
				return InputAction{}
			}
		}
		if handled {
			return InputAction{}
		}
	}
	if list := negotiationSelectedItemsList(r.gs, &r.negotiation.offered, layout.right.panel); list != nil {
		handled := list.HandleInput(input)
		if input.LeftJustPressed {
			if index := list.ItemIndexAt(input.MouseX, input.MouseY); index >= 0 {
				r.negotiation.offered.items = append(r.negotiation.offered.items[:index], r.negotiation.offered.items[index+1:]...)
				return InputAction{}
			}
		}
		if handled {
			return InputAction{}
		}
	}
	if r.negotiation.requested.kind == negotiationKindRegion {
		if dropdown := negotiationRegionDropdown(r.gs, requestOwner, &r.negotiation.requested, layout.left.panel); handleNegotiationRegionDropdownInput(input, dropdown, &r.negotiation.requested) {
			return InputAction{}
		}
	}
	if r.negotiation.offered.kind == negotiationKindRegion {
		if dropdown := negotiationRegionDropdown(r.gs, offerOwner, &r.negotiation.offered, layout.right.panel); handleNegotiationRegionDropdownInput(input, dropdown, &r.negotiation.offered) {
			return InputAction{}
		}
	}
	if layout.left.kind.HandleInput(input) {
		cycleNegotiationPicker(r.gs, requestOwner, &r.negotiation.requested, 1)
		return InputAction{}
	}
	if layout.left.option.HandleInput(input) {
		if r.negotiation.requested.kind == negotiationKindRegion {
			if dropdown := negotiationRegionDropdown(r.gs, requestOwner, &r.negotiation.requested, layout.left.panel); dropdown != nil {
				dropdown.Toggle()
			}
		} else {
			cycleNegotiationOption(r.gs, requestOwner, &r.negotiation.requested)
		}
		return InputAction{}
	}
	if layout.left.minus.HandleInput(input) {
		adjustNegotiationAmount(r.gs, requestOwner, &r.negotiation.requested, -1)
		return InputAction{}
	}
	if layout.left.plus.HandleInput(input) {
		adjustNegotiationAmount(r.gs, requestOwner, &r.negotiation.requested, 1)
		return InputAction{}
	}
	if layout.left.add.HandleInput(input) {
		addNegotiationItem(r.gs, requestOwner, &r.negotiation.requested)
		return InputAction{}
	}
	if layout.right.kind.HandleInput(input) {
		cycleNegotiationPicker(r.gs, offerOwner, &r.negotiation.offered, 1)
		return InputAction{}
	}
	if layout.right.option.HandleInput(input) {
		if r.negotiation.offered.kind == negotiationKindRegion {
			if dropdown := negotiationRegionDropdown(r.gs, offerOwner, &r.negotiation.offered, layout.right.panel); dropdown != nil {
				dropdown.Toggle()
			}
		} else {
			cycleNegotiationOption(r.gs, offerOwner, &r.negotiation.offered)
		}
		return InputAction{}
	}
	if layout.right.minus.HandleInput(input) {
		adjustNegotiationAmount(r.gs, offerOwner, &r.negotiation.offered, -1)
		return InputAction{}
	}
	if layout.right.plus.HandleInput(input) {
		adjustNegotiationAmount(r.gs, offerOwner, &r.negotiation.offered, 1)
		return InputAction{}
	}
	if layout.right.add.HandleInput(input) {
		addNegotiationItem(r.gs, offerOwner, &r.negotiation.offered)
	}
	return InputAction{}
}
