package render

import (
	"image/color"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/economy"
	"mapp-game-go/internal/faction"
	"mapp-game-go/internal/religion"
	"mapp-game-go/internal/scenario"
	"mapp-game-go/internal/state"
	gameui "mapp-game-go/internal/ui"
	"mapp-game-go/internal/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (r *Renderer) drawEditModeHud(screen *ebiten.Image) {
	const panelW, panelH = float32(620), float32(132)
	x, y := float32(18), float32(18)
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 20, 24, 220})
	drawPanelBorder(screen, x, y, panelW, panelH)

	title := "EDIT MODE"
	if r.editDirty {
		title += " *"
	}
	DrawText(screen, title, float64(x)+14, float64(y)+10, FaceMed, ColorGold)
	help := "Sol: sec | Sag surukle: yerlesim tasi | Alt+sol: ekle | Ctrl+Alt+sol: bolge | Shift+sol: merkez | Ctrl+Z/Y"
	DrawText(screen, trimTextToWidth(help, FaceSmall, float64(panelW)-28),
		float64(x)+14, float64(y)+36, FaceSmall, ColorWhite)

	info := "Secili: yok"
	if region, ok := r.gs.Regions[r.editSelectedRegion]; ok &&
		r.editSelectedSettlement >= 0 && r.editSelectedSettlement < len(region.Settlements) {
		s := region.Settlements[r.editSelectedSettlement]
		info = region.NameTR + " / " + s.NameTR + "  (" + itoa(s.X) + "," + itoa(s.Y) + ")"
	} else if region, ok := r.gs.Regions[r.editSelectedRegion]; ok && region != nil {
		info = "Merkez: " + region.NameTR + "  (" + itoa(region.WorldX) + "," + itoa(region.WorldY) + ")"
	}
	DrawText(screen, info, float64(x)+14, float64(y)+58, FaceSmall, ColorGray)
	debugState := "Voronoi debug: kapali"
	if r.editVoronoiDebug {
		debugState = "Voronoi debug: acik"
	}
	historyState := "Geri/Ileri: " + itoa(len(r.editUndoStack)) + "/" + itoa(len(r.editRedoStack))
	if r.editRenaming {
		DrawText(screen, r.editTextLabel()+": "+string(r.editTextRunes), float64(x)+14, float64(y)+80, FaceSmall, ColorGold)
		if r.editTextError != "" {
			DrawText(screen, r.editTextError, float64(x)+14, float64(y)+100, FaceSmall, ColorRed)
		}
	} else {
		DrawText(screen, debugState+"   "+historyState+"   V: debug   Esc: ana menu", float64(x)+14, float64(y)+80, FaceSmall, ColorGray)
	}
	passageState := "P: karasal geçiş modu kapalı"
	if r.editLandPassageMode {
		passageState = "Geçiş ekleme açık"
		if r.editLandPassageFrom != "" {
			passageState += "   İlk: " + string(r.editLandPassageFrom) + "  (bitiş noktasına tıkla)"
		} else {
			passageState += "   Başlangıç noktasına tıkla"
		}
	} else if r.editLandPassageAdjustMode {
		passageState = "Geçiş düzenleme açık"
		if r.editLandPassageSelected >= 0 {
			passageState += "   Seçili geçiş: " + itoa(r.editLandPassageSelected+1)
		}
	} else if r.editNeighborAddMode {
		passageState = "Komşu ekleme açık"
		if r.editNeighborAddFrom != "" {
			passageState += "   Kaynak: " + string(r.editNeighborAddFrom) + "  (hedef kara bölgesine tıkla)"
		}
	}
	if r.editLandPassageMessage != "" {
		passageState += "   " + r.editLandPassageMessage
	}
	if r.editNeighborAddMessage != "" {
		passageState += "   " + r.editNeighborAddMessage
	}
	if !(r.editRenaming && r.editTextError != "") {
		DrawText(screen, passageState, float64(x)+14, float64(y)+100, FaceSmall, ColorGold)
	}
}

func (r *Renderer) drawEditInspector(screen *ebiten.Image) {
	x, y, w, h := editInspectorRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{16, 20, 24, 226})
	drawPanelBorder(screen, x, y, w, h)

	DrawText(screen, "EDITOR", float64(x)+14, float64(y)+10, FaceMed, ColorGold)
	r.drawEditInspectorTab(screen, editInspectorSettlement, "Yerleşim Birimi")
	r.drawEditInspectorTab(screen, editInspectorRegion, "Bölge")
	r.drawEditInspectorTab(screen, editInspectorFaction, "Devlet")
	r.drawEditInspectorTab(screen, editInspectorMap, "Harita")
	r.drawEditInspectorTab(screen, editInspectorData, "Veri")
	ly := float64(y) + 82

	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape {
		r.drawEditShapeInspector(screen, ly)
		drawEditInspectorSaveButton(screen)
		return
	}

	if r.editInspectorTab == editInspectorData {
		r.drawEditScenarioDataInspector(screen, ly)
		drawEditInspectorSaveButton(screen)
		return
	}

	if r.editInspectorTab == editInspectorFaction {
		r.drawEditDataInspector(screen, ly)
		drawEditInspectorSaveButton(screen)
		return
	}

	region := r.gs.Regions[r.editSelectedRegion]
	if r.SelectedArmy != "" {
		if a, ok := r.gs.Armies[r.SelectedArmy]; ok && a != nil {
			DrawText(screen, "Ordu: "+string(a.ID), float64(x)+14, ly, FaceSmall, ColorWhite)
			ly += 18
			DrawText(screen, "Bolge: "+string(a.RegionID), float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			DrawText(screen, "Birim: "+itoa(len(a.Units))+" / 20", float64(x)+14, ly, FaceSmall, ColorGray)
			if r.editInspectorTab == editInspectorSettlement {
				r.drawEditSettlementButtons(screen, region)
			} else {
				r.drawEditRegionButtons(screen, region)
			}
			drawEditInspectorSaveButton(screen)
			return
		}
	}

	if region == nil {
		DrawText(screen, "Haritadan bir bolge veya yerlesim sec.", float64(x)+14, ly, FaceSmall, ColorGray)
		r.drawEditRegionButtons(screen, nil)
		drawEditInspectorSaveButton(screen)
		return
	}

	name := region.NameTR
	if name == "" {
		name = region.Name
	}
	regionKind := "Kara Bolgesi"
	ownerLabel := region.OwnerID
	settlementLabel := itoa(len(region.Settlements))
	if region.IsSea {
		regionKind = "Deniz Bolgesi"
		if ownerLabel == "" {
			ownerLabel = "-"
		}
		settlementLabel = "yok"
	}
	DrawText(screen, name, float64(x)+14, ly, FaceSmall, ColorWhite)
	ly += 18
	DrawText(screen, "ID: "+string(region.ID), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Tur: "+regionKind+"   Sahip: "+ownerLabel+"   Arazi: "+string(region.Terrain), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	successorLabel := region.SuccessorFactionID
	if successorLabel == "" {
		successorLabel = "-"
	}
	DrawText(screen, "Ardil Devlet: "+successorLabel, float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Merkez: "+itoa(region.WorldX)+","+itoa(region.WorldY)+"   Yerlesim: "+settlementLabel, float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 22
	DrawText(screen, "Kilit: "+editBoolLabel(region.IsLocked)+"   Acilis: "+itoa(region.UnlockTurn)+"   Komsu: "+itoa(len(region.Neighbors)), float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 20

	if r.hasEditSelection() {
		settlement := region.Settlements[r.editSelectedSettlement]
		sName := settlement.NameTR
		if sName == "" {
			sName = settlement.Name
		}
		DrawText(screen, "Secili yerlesim: "+sName, float64(x)+14, ly, FaceSmall, ColorGold)
		ly += 18
		DrawText(screen, settlement.ID+"  "+string(settlement.Type)+"  nüfus "+itoa(settlement.Population)+"  "+itoa(settlement.X)+","+itoa(settlement.Y),
			float64(x)+14, ly, FaceSmall, ColorGray)
		if settlement.IsCenter {
			ly += 18
			DrawText(screen, "Ana yerlesim", float64(x)+14, ly, FaceSmall, ColorGray)
		}
	} else if region.IsSea {
		DrawText(screen, "Deniz bolgesinde yerlesim yok.", float64(x)+14, ly, FaceSmall, ColorGray)
	} else {
		DrawText(screen, "Yerlesim secili degil.", float64(x)+14, ly, FaceSmall, ColorGray)
	}

	if r.editInspectorTab == editInspectorSettlement {
		r.drawEditSettlementButtons(screen, region)
	} else {
		r.drawEditRegionButtons(screen, region)
	}
	drawUIDropdown(screen, r.editOwnerDropdown)
	drawUIDropdown(screen, r.editSuccessorDropdown)
	drawUIDropdown(screen, r.editTerrainDropdown)
	drawUIDropdown(screen, r.editSettlementTypeDropdown)
	drawUIDropdown(screen, r.editUnitTypeDropdown)
	drawEditInspectorSaveButton(screen)
}

func (r *Renderer) drawEditSettlementButtons(screen *ebiten.Image, region *world.Region) {
	canAdd := canAddSettlementToRegion(region)
	canSettlement := r.hasEditSelection()
	addSettlementLabel := "Yerlesim Ekle"
	settlementTypeLabel := "Yerlesim Tipi"
	renameSettlementLabel := "Yerlesim Adı"
	deleteSettlementLabel := "Yerlesim Sil"
	if region != nil && region.IsSea {
		addSettlementLabel = "Denizde Yok"
		settlementTypeLabel = "Tip Yok"
		renameSettlementLabel = "Isim Yok"
		deleteSettlementLabel = "Silinmez"
	} else if !canSettlement {
		settlementTypeLabel = "Tip Sec"
		renameSettlementLabel = "Isim Sec"
		deleteSettlementLabel = "Sil Sec"
	}
	drawEditInspectorButton(screen, editButtonAddSettlement, addSettlementLabel, canAdd)
	drawEditInspectorButton(screen, editButtonSettlementType, settlementTypeLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonSetCenterSettlement, "Merkez Yap", canSettlement)
	drawEditInspectorButton(screen, editButtonRenameSettlement, renameSettlementLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonDeleteSettlement, deleteSettlementLabel, canSettlement)
	r.drawEditArmyButtons(screen, region)
}

func (r *Renderer) drawEditArmyButtons(screen *ebiten.Image, region *world.Region) {
	drawEditInspectorButton(screen, editButtonAddArmy, "Ordu Ekle", r.canAddEditLandArmy(region))
	drawEditInspectorButton(screen, editButtonAddFleet, "Donanma Ekle", r.canAddEditFleet(region))
	drawEditInspectorButton(screen, editButtonDeleteArmy, "Ordu Sil", r.SelectedArmy != "")
	unitTypeLabel := "Birim Tipi"
	if r.editSelectedUnitType != "" {
		unitTypeLabel = "Birim Tipi: " + r.editSelectedUnitType
	}
	drawEditInspectorButton(screen, editButtonArmyUnitType, unitTypeLabel, r.SelectedArmy != "")
	drawEditInspectorButton(screen, editButtonArmyUnitMinus, "Birim -", r.canRemoveSelectedArmyUnit())
	drawEditInspectorButton(screen, editButtonArmyUnitPlus, "Birim +", r.canAddSelectedArmyUnit())
	drawEditInspectorButton(screen, editButtonArmyOwnerFromRegion, "Bu Devlete Ata", r.canAssignSelectedArmyToRegionOwner())
}

func (r *Renderer) drawEditRegionButtons(screen *ebiten.Image, region *world.Region) {
	canRegion := region != nil
	drawEditInspectorButton(screen, editButtonAddRegion, "Yeni Bölge Ekle", canRegion)
	drawEditInspectorButton(screen, editButtonDeleteRegion, "Bölgeyi Sil", canRegion)
	terrainLabel := "Bölge Tipi"
	if region != nil && region.IsTerrainArea {
		terrainLabel = "Arazi Tipi"
	}
	drawEditInspectorButton(screen, editButtonRegionTerrain, terrainLabel, canRegion)
	drawEditInspectorButton(screen, editButtonRegionNameTR, "Ad TR", canRegion)
	drawEditInspectorButton(screen, editButtonRegionName, "Ad EN", canRegion)
	drawEditInspectorButton(screen, editButtonRegionID, "ID", canRegion)
	drawEditInspectorButton(screen, editButtonRegionLock, "Kilit", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockMinus, "-10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockPlus, "+10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonSyncNeighbors, "Komşu Sync", canRegion)
	drawEditInspectorButton(screen, editButtonAddNeighbor, "Komşu Ekle", canRegion)
}

func drawEditInspectorSaveButton(screen *ebiten.Image) {
	rect := editInspectorButtonRect(editButtonSaveScenario)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), "Değişiklikleri Kaydet", true)
}

func (r *Renderer) drawEditInspectorTab(screen *ebiten.Image, tab editInspectorTab, label string) {
	rect := editInspectorTabRect(tab)
	active := r.editInspectorTab == tab || (tab == editInspectorMap && r.editInspectorTab == editInspectorShape)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), label, active)
}

func (r *Renderer) drawEditDataInspector(screen *ebiten.Image, ly float64) {
	x, _, _, _ := editInspectorRect()
	region := r.gs.Regions[r.editSelectedRegion]
	f := r.selectedEditFaction()

	DrawText(screen, "DEVLET VE ORDU", float64(x)+14, ly, FaceSmall, ColorGold)
	ly += 22
	if f == nil {
		DrawText(screen, "Sahipli bolge veya ordu sec.", float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 20
	} else {
		name := f.NameTR
		if name == "" {
			name = f.Name
		}
		DrawText(screen, "Fraksiyon: "+name+" ["+string(f.ID)+"]", float64(x)+14, ly, FaceSmall, ColorWhite)
		ly += 18
		DrawText(screen, economy.FormatResourceAmountTR(economy.ResourceGold, f.Gold)+"  "+economy.FormatResourceAmountTR(economy.ResourceGrain, f.Grain)+"  "+economy.FormatResourceAmountTR(economy.ResourceIron, f.Iron), float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 18
		DrawText(screen, economy.FormatResourceAmountTR(economy.ResourceTimber, f.Timber)+"  "+economy.FormatResourceAmountTR(economy.ResourceSpice, f.Spice)+"  "+economy.FormatResourceAmountTR(economy.ResourceCloth, f.Cloth), float64(x)+14, ly, FaceSmall, ColorGray)
		ly += 18
		DrawText(screen, "Playable: "+editBoolLabel(f.IsPlayable)+"  AI: "+itoa(f.AIAggressiveness), float64(x)+14, ly, FaceSmall, ColorGray)
	}
	ly += 24

	if r.SelectedArmy != "" {
		if a := r.gs.Armies[r.SelectedArmy]; a != nil {
			r.ensureEditSelectedUnitType(a)
			DrawText(screen, "Ordu: "+string(a.ID), float64(x)+14, ly, FaceSmall, ColorGold)
			ly += 18
			kind := "Kara"
			if a.IsNaval {
				kind = "Donanma"
			}
			DrawText(screen, "Tip: "+kind+"  Sahip: "+a.OwnerID+"  Bolge: "+string(a.RegionID), float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			DrawText(screen, "Birim: "+itoa(len(a.Units))+" / "+itoa(army.MaxArmySize)+"  Secili: "+r.editSelectedUnitType, float64(x)+14, ly, FaceSmall, ColorGray)
			ly += 18
			r.drawEditArmyUnitCounts(screen, a, float64(x)+14, ly)
		}
	} else {
		DrawText(screen, "Ordu secili degil.", float64(x)+14, ly, FaceSmall, ColorGray)
	}

	canRegion := region != nil
	drawEditInspectorButton(screen, editButtonRegionOwner, "Bölge Sahibi Belirle", canRegion)
	drawEditInspectorButton(screen, editButtonRegionSuccessor, "Ardıl Devlet Belirle", canRegion && !region.IsSea)
	drawEditInspectorButton(screen, editButtonSetFactionCapital, "Başkent Yap", r.canSetSelectedFactionCapital())
	drawEditInspectorButton(screen, editButtonAddFaction, "Yeni Devlet Ekle", true)
	drawEditInspectorButton(screen, editButtonEditFaction, "Devleti Düzenle", f != nil)
	drawEditInspectorButton(screen, editButtonDeleteFaction, "Devleti Sil", f != nil)
	drawUIDropdown(screen, r.editUnitTypeDropdown)
	drawUIDropdown(screen, r.editOwnerDropdown)
	drawUIDropdown(screen, r.editSuccessorDropdown)
}

func (r *Renderer) drawEditScenarioDataInspector(screen *ebiten.Image, ly float64) {
	x, _, _, _ := editInspectorRect()
	DrawText(screen, "SENARYO VERİLERİ", float64(x)+14, ly, FaceSmall, ColorGold)
	ly += 24
	DrawText(screen, "Bu senaryodaki düzenlemeler geçici olarak tutulur.", float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 18
	DrawText(screen, "Kaydet düğmesi tüm sekmelerde panelin altındadır.", float64(x)+14, ly, FaceSmall, ColorGray)
	ly += 24
	DrawText(screen, "Undo: "+itoa(len(r.editUndoStack))+"   Redo: "+itoa(len(r.editRedoStack)), float64(x)+14, ly, FaceSmall, ColorWhite)
	ly += 22
	if r.editDirty {
		DrawText(screen, "Durum: Kaydedilmemiş değişiklikler var.", float64(x)+14, ly, FaceSmall, ColorGold)
	} else {
		DrawText(screen, "Durum: Tüm değişiklikler kayıtlı.", float64(x)+14, ly, FaceSmall, ColorGray)
	}
}

func (r *Renderer) drawEditArmyUnitCounts(screen *ebiten.Image, a *army.Army, x, y float64) {
	if len(a.Units) == 0 {
		DrawText(screen, "Birim yok.", x, y, FaceSmall, ColorGray)
		return
	}
	// Devlet sekmesinin düğmeleriyle çakışmaması için ayrıntılı dağılımı
	// satır satır büyütmek yerine tek satırlık özet tutuyoruz.
	DrawText(screen, "Birim toplamı: "+itoa(len(a.Units)), x, y, FaceSmall, ColorGray)
}

func (r *Renderer) drawEditFactionForm(screen *ebiten.Image) {
	if !r.editFactionForm.show {
		return
	}
	x, y, w, h := editFactionFormRect()
	drawRoundedRect(screen, x, y, w, h, 8, color.RGBA{14, 18, 22, 244})
	drawPanelBorder(screen, x, y, w, h)
	title := "FACTION EKLE"
	if !r.editFactionForm.create {
		title = "FACTION DUZENLE"
	}
	DrawText(screen, title, float64(x)+18, float64(y)+14, FaceLarge, ColorGold)

	r.drawFactionFormField(screen, editFactionFieldID, "ID", r.editFactionForm.id)
	r.drawFactionFormField(screen, editFactionFieldNameTR, "Ad TR", r.editFactionForm.nameTR)
	r.drawFactionFormField(screen, editFactionFieldName, "Ad EN", r.editFactionForm.name)
	r.drawFactionFormField(screen, editFactionFieldGold, economy.ResourceNameTR(economy.ResourceGold), r.editFactionForm.gold)
	r.drawFactionFormField(screen, editFactionFieldGrain, economy.ResourceNameTR(economy.ResourceGrain), r.editFactionForm.grain)
	r.drawFactionFormField(screen, editFactionFieldIron, economy.ResourceNameTR(economy.ResourceIron), r.editFactionForm.iron)
	r.drawFactionFormField(screen, editFactionFieldTimber, economy.ResourceNameTR(economy.ResourceTimber), r.editFactionForm.timber)
	r.drawFactionFormField(screen, editFactionFieldSpice, economy.ResourceNameTR(economy.ResourceSpice), r.editFactionForm.spice)
	r.drawFactionFormField(screen, editFactionFieldCloth, economy.ResourceNameTR(economy.ResourceCloth), r.editFactionForm.cloth)
	r.drawFactionFormField(screen, editFactionFieldAI, "AI", r.editFactionForm.ai)

	drawEditFactionFormButton(screen, editFactionFormReligion, "Din: "+religion.DisplayNameTR(r.editFactionForm.religion))
	drawEditFactionFormButton(screen, editFactionFormPlayable, "Playable: "+editBoolLabel(r.editFactionForm.playable))
	relationTitle := "Iliski: yok"
	if r.editFactionForm.relationTarget != "" {
		relationTitle = "Iliski: " + string(r.editFactionForm.relationTarget)
	}
	drawEditFactionFormButton(screen, editFactionFormRelationTarget, relationTitle)
	drawEditFactionFormButton(screen, editFactionFormRelationStance, "Durum: "+faction.DiplomaticStanceLabelTR(r.editFactionForm.relationStance))
	drawEditFactionFormButton(screen, editFactionFormRelationScoreMinus, "Skor -10")
	drawEditFactionFormButton(screen, editFactionFormRelationScorePlus, "Skor +10")
	DrawText(screen, "Skor: "+r.editFactionForm.relationScore, float64(x)+18, float64(y)+304, FaceSmall, ColorGray)

	col := r.editFactionForm.color
	preview := editFactionFormColorPreviewRect()
	vector.FillRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), color.RGBA{col[0], col[1], col[2], 255}, false)
	vector.StrokeRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), 1, ColorGold, false)
	DrawText(screen, "Renk "+itoa(int(col[0]))+","+itoa(int(col[1]))+","+itoa(int(col[2])), float64(x)+338, float64(y)+332, FaceSmall, ColorGray)
	drawEditFactionFormButton(screen, editFactionFormRedMinus, "R-")
	drawEditFactionFormButton(screen, editFactionFormRedPlus, "R+")
	drawEditFactionFormButton(screen, editFactionFormGreenMinus, "G-")
	drawEditFactionFormButton(screen, editFactionFormGreenPlus, "G+")
	drawEditFactionFormButton(screen, editFactionFormBlueMinus, "B-")
	drawEditFactionFormButton(screen, editFactionFormBluePlus, "B+")

	if r.editFactionForm.errorText != "" {
		DrawText(screen, r.editFactionForm.errorText, float64(x)+18, float64(y)+float64(h)-74, FaceSmall, ColorRed)
	}
	drawEditFactionFormButton(screen, editFactionFormSave, "Kaydet")
	drawEditFactionFormButton(screen, editFactionFormCancel, "Iptal")
}

func drawEditFactionFormButton(screen *ebiten.Image, kind editFactionFormButton, label string) {
	x, y, w, h := rectXYWH(editFactionFormButtonRect(kind))
	drawTinyPanelButton(screen, x, y, w, h, label, true)
}

func editRectButton(r uiRect, label string) gameui.Button {
	return gameui.NewButton(r[0], r[1], r[2], r[3], label)
}

func (r *Renderer) drawFactionFormField(screen *ebiten.Image, field editFactionFormField, label, value string) {
	rect := editFactionFieldRect(field)
	col := color.RGBA{28, 32, 38, 235}
	if r.editFactionForm.active == field {
		col = color.RGBA{44, 48, 54, 245}
	}
	vector.FillRect(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), col, false)
	vector.StrokeRect(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), 1, color.RGBA{120, 105, 60, 210}, false)
	DrawText(screen, label, rect[0], rect[1]-16, FaceSmall, ColorGray)
	DrawText(screen, value, rect[0]+8, rect[1]+7, FaceSmall, ColorWhite)
}

func rectXYWH(rect uiRect) (float32, float32, float32, float32) {
	return float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3])
}

func editFactionFormRect() (float32, float32, float32, float32) {
	const w, h = float32(640), float32(520)
	return float32(ScreenWidth)/2 - w/2, float32(ScreenHeight)/2 - h/2, w, h
}

func editFactionFormHit(mx, my float64) bool {
	x, y, w, h := editFactionFormRect()
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
}

func editFactionFieldRect(field editFactionFormField) uiRect {
	x, y, _, _ := editFactionFormRect()
	left := float64(x) + 18
	right := float64(x) + 338
	top := float64(y) + 78
	const fw, fh, gap = float64(284), float64(30), float64(24)
	row := func(n int) float64 { return top + float64(n)*(fh+gap) }
	switch field {
	case editFactionFieldID:
		return uiRect{left, row(0), fw, fh}
	case editFactionFieldNameTR:
		return uiRect{right, row(0), fw, fh}
	case editFactionFieldName:
		return uiRect{left, row(1), fw, fh}
	case editFactionFieldGold:
		return uiRect{right, row(1), fw/2 - 6, fh}
	case editFactionFieldGrain:
		return uiRect{right + fw/2 + 6, row(1), fw/2 - 6, fh}
	case editFactionFieldIron:
		return uiRect{left, row(2), fw/2 - 6, fh}
	case editFactionFieldTimber:
		return uiRect{left + fw/2 + 6, row(2), fw/2 - 6, fh}
	case editFactionFieldSpice:
		return uiRect{right, row(2), fw/2 - 6, fh}
	case editFactionFieldCloth:
		return uiRect{right + fw/2 + 6, row(2), fw/2 - 6, fh}
	case editFactionFieldAI:
		return uiRect{left, row(3), fw/2 - 6, fh}
	default:
		return uiRect{}
	}
}

func buildEditFactionFieldButton(field editFactionFormField, label string) gameui.Button {
	return editRectButton(editFactionFieldRect(field), label)
}

type editFactionFormButton int

const (
	editFactionFormSave editFactionFormButton = iota
	editFactionFormCancel
	editFactionFormReligion
	editFactionFormPlayable
	editFactionFormRelationTarget
	editFactionFormRelationStance
	editFactionFormRelationScoreMinus
	editFactionFormRelationScorePlus
	editFactionFormRedMinus
	editFactionFormRedPlus
	editFactionFormGreenMinus
	editFactionFormGreenPlus
	editFactionFormBlueMinus
	editFactionFormBluePlus
)

func editFactionFormButtonRect(kind editFactionFormButton) uiRect {
	x, y, w, h := editFactionFormRect()
	right := float64(x) + 338
	switch kind {
	case editFactionFormReligion:
		return uiRect{right, float64(y) + 240, 136, 28}
	case editFactionFormPlayable:
		return uiRect{right + 148, float64(y) + 240, 136, 28}
	case editFactionFormRelationTarget:
		return uiRect{float64(x) + 18, float64(y) + 240, 284, 28}
	case editFactionFormRelationStance:
		return uiRect{float64(x) + 18, float64(y) + 272, 136, 28}
	case editFactionFormRelationScoreMinus:
		return uiRect{float64(x) + 166, float64(y) + 272, 64, 28}
	case editFactionFormRelationScorePlus:
		return uiRect{float64(x) + 238, float64(y) + 272, 64, 28}
	case editFactionFormRedMinus:
		return uiRect{right, float64(y) + 382, 42, 26}
	case editFactionFormRedPlus:
		return uiRect{right + 48, float64(y) + 382, 42, 26}
	case editFactionFormGreenMinus:
		return uiRect{right + 100, float64(y) + 382, 42, 26}
	case editFactionFormGreenPlus:
		return uiRect{right + 148, float64(y) + 382, 42, 26}
	case editFactionFormBlueMinus:
		return uiRect{right + 200, float64(y) + 382, 42, 26}
	case editFactionFormBluePlus:
		return uiRect{right + 248, float64(y) + 382, 42, 26}
	case editFactionFormSave:
		return uiRect{float64(x) + float64(w) - 264, float64(y) + float64(h) - 52, 116, 32}
	case editFactionFormCancel:
		return uiRect{float64(x) + float64(w) - 136, float64(y) + float64(h) - 52, 116, 32}
	default:
		return uiRect{}
	}
}

func buildEditFactionFormButton(kind editFactionFormButton, label string) gameui.Button {
	return editRectButton(editFactionFormButtonRect(kind), label)
}

func editFactionFormColorPreviewRect() uiRect {
	x, y, _, _ := editFactionFormRect()
	return uiRect{float64(x) + 338, float64(y) + 352, 284, 22}
}

func drawEditInspectorButton(screen *ebiten.Image, kind editInspectorButton, label string, active bool) {
	rect := editInspectorButtonRect(kind)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), label, active)
}

type editInspectorButton int

const (
	editButtonNone editInspectorButton = iota
	editButtonAddSettlement
	editButtonSettlementType
	editButtonSetCenterSettlement
	editButtonRenameSettlement
	editButtonSetFactionCapital
	editButtonRegionTerrain
	editButtonRegionOwner
	editButtonRegionSuccessor
	editButtonRegionNameTR
	editButtonRegionName
	editButtonRegionID
	editButtonRegionLock
	editButtonUnlockMinus
	editButtonUnlockPlus
	editButtonSyncNeighbors
	editButtonAddRegion
	editButtonDeleteRegion
	editButtonDeleteSettlement
	editButtonSaveScenario
	editButtonShapePaint
	editButtonShapeErase
	editButtonShapeRegionPaint
	editButtonShapeRegionErase
	editButtonShapeBrushMinus
	editButtonShapeBrushPlus
	editButtonLandPassageAdd
	editButtonLandPassageAdjust
	editButtonLandPassageDelete
	editButtonAddNeighbor
	editButtonTerrainArea
	editButtonTerrainAreaType
	editButtonTerrainAreaCost
	editButtonTerrainAreaAttrition
	editButtonTerrainAreaDelete
	editButtonAddFaction
	editButtonEditFaction
	editButtonDeleteFaction
	editButtonAddArmy
	editButtonAddFleet
	editButtonDeleteArmy
	editButtonArmyUnitType
	editButtonArmyUnitMinus
	editButtonArmyUnitPlus
	editButtonArmyOwnerFromRegion
)

func editInspectorRect() (float32, float32, float32, float32) {
	// 550 px yükseklik, üstteki Edit Mode yardım HUD'ı ile çakışmayı önler.
	// İçerik sekmelere dağıtıldığı için önceki dar panelden daha dengeli
	// ve beş sekmenin okunabildiği 440 px genişlik kullanılır.
	const w, h = float32(440), float32(550)
	return 18, float32(ScreenHeight) - h - 18, w, h
}

func editInspectorHit(mx, my float64) bool {
	x, y, w, h := editInspectorRect()
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
}

func editInspectorButtonRect(kind editInspectorButton) uiRect {
	x, y, w, h := editInspectorRect()
	const bh, gap = float64(26), float64(6)
	left := float64(x) + 14
	bw := float64(w) - 28
	colW := (bw - gap) / 2
	right := left + colW + gap
	row := func(index int) float64 {
		return float64(y) + float64(h) - 298 + float64(index)*(bh+gap)
	}
	full := func(index int) uiRect {
		return uiRect{left, row(index), bw, bh}
	}
	leftRect := func(index int) uiRect {
		return uiRect{left, row(index), colW, bh}
	}
	rightRect := func(index int) uiRect {
		return uiRect{right, row(index), colW, bh}
	}
	switch kind {
	case editButtonAddSettlement:
		return leftRect(0)
	case editButtonSettlementType:
		return rightRect(0)
	case editButtonSetCenterSettlement:
		return leftRect(1)
	case editButtonRenameSettlement:
		return rightRect(1)
	case editButtonDeleteSettlement:
		return leftRect(2)
	case editButtonAddRegion:
		return leftRect(0)
	case editButtonDeleteRegion:
		return rightRect(0)
	case editButtonRegionTerrain:
		return rightRect(1)
	case editButtonRegionNameTR:
		return leftRect(2)
	case editButtonRegionName:
		return rightRect(2)
	case editButtonRegionID:
		return leftRect(3)
	case editButtonRegionLock:
		return rightRect(3)
	case editButtonUnlockMinus:
		return leftRect(4)
	case editButtonUnlockPlus:
		return rightRect(4)
	case editButtonSyncNeighbors:
		return leftRect(5)
	case editButtonAddNeighbor:
		return rightRect(5)
	case editButtonSetFactionCapital:
		return leftRect(1)
	case editButtonSaveScenario:
		return uiRect{left, float64(y) + float64(h) - 42, bw, 32}
	case editButtonRegionSuccessor:
		return rightRect(0)
	case editButtonRegionOwner:
		return leftRect(0)
	case editButtonShapePaint:
		return leftRect(0)
	case editButtonShapeErase:
		return rightRect(0)
	case editButtonShapeRegionPaint:
		return leftRect(1)
	case editButtonShapeRegionErase:
		return rightRect(1)
	case editButtonShapeBrushMinus:
		return leftRect(2)
	case editButtonShapeBrushPlus:
		return rightRect(2)
	case editButtonLandPassageAdd:
		return leftRect(3)
	case editButtonLandPassageAdjust:
		return rightRect(3)
	case editButtonLandPassageDelete:
		return leftRect(4)
	case editButtonTerrainArea:
		return leftRect(6)
	case editButtonTerrainAreaType:
		return full(5)
	case editButtonTerrainAreaCost:
		return rightRect(6)
	case editButtonTerrainAreaAttrition:
		return leftRect(7)
	case editButtonTerrainAreaDelete:
		return rightRect(7)
	case editButtonAddFaction:
		return rightRect(1)
	case editButtonEditFaction:
		return leftRect(2)
	case editButtonDeleteFaction:
		return rightRect(2)
	case editButtonAddArmy:
		return leftRect(3)
	case editButtonAddFleet:
		return rightRect(3)
	case editButtonDeleteArmy:
		return leftRect(4)
	case editButtonArmyUnitType:
		return full(5)
	case editButtonArmyUnitMinus:
		return leftRect(6)
	case editButtonArmyUnitPlus:
		return rightRect(6)
	case editButtonArmyOwnerFromRegion:
		return full(7)
	default:
		return uiRect{}
	}
}

func editInspectorTabRect(tab editInspectorTab) uiRect {
	x, y, _, _ := editInspectorRect()
	const th, gap = float64(30), float64(5)
	left := float64(x) + 82
	widths := [...]float64{112, 54, 54, 58, 54}
	index := -1
	switch tab {
	case editInspectorSettlement:
		index = 0
	case editInspectorRegion:
		index = 1
	case editInspectorFaction:
		index = 2
	case editInspectorMap:
		index = 3
	case editInspectorData:
		index = 4
	default:
		return uiRect{}
	}
	for i := 0; i < index; i++ {
		left += widths[i] + gap
	}
	return uiRect{left, float64(y) + 9, widths[index], th}
}

func buildEditInspectorTabButton(tab editInspectorTab, label string) gameui.Button {
	return editRectButton(editInspectorTabRect(tab), label)
}

func buildEditInspectorActionButton(kind editInspectorButton, label string) gameui.Button {
	return editRectButton(editInspectorButtonRect(kind), label)
}

func editInspectorButtonAt(mx, my float64) editInspectorButton {
	if kind := editMapInspectorButtonAt(mx, my); kind != editButtonNone {
		return kind
	}
	return editDataInspectorButtonAt(mx, my)
}

func editMapInspectorButtonAt(mx, my float64) editInspectorButton {
	// Geriye dönük yardımcı: eski çağrılar hem bölge hem yerleşim
	// düğmelerini bu fonksiyon üzerinden arıyordu. Çakışan gizli sekme
	// düğmelerinin yanlış action üretmemesi için açık liste kullanılır.
	if buildEditInspectorActionButton(editButtonSetFactionCapital, "").HitTest(mx, my) {
		return editButtonSetFactionCapital
	}
	if kind := editRegionInspectorButtonAt(mx, my); kind != editButtonNone {
		return kind
	}
	if kind := editSettlementInspectorButtonAt(mx, my); kind != editButtonNone {
		return kind
	}
	return editButtonNone
}

func editSettlementInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range [...]editInspectorButton{
		editButtonAddSettlement,
		editButtonSettlementType,
		editButtonSetCenterSettlement,
		editButtonRenameSettlement,
		editButtonDeleteSettlement,
		editButtonAddArmy,
		editButtonAddFleet,
		editButtonDeleteArmy,
		editButtonArmyUnitType,
		editButtonArmyUnitMinus,
		editButtonArmyUnitPlus,
		editButtonArmyOwnerFromRegion,
	} {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func editRegionInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range [...]editInspectorButton{
		editButtonAddRegion,
		editButtonDeleteRegion,
		editButtonRegionTerrain,
		editButtonRegionNameTR,
		editButtonRegionName,
		editButtonRegionID,
		editButtonRegionLock,
		editButtonUnlockMinus,
		editButtonUnlockPlus,
		editButtonSyncNeighbors,
		editButtonAddNeighbor,
	} {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func editFactionInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range [...]editInspectorButton{
		editButtonRegionOwner,
		editButtonRegionSuccessor,
		editButtonSetFactionCapital,
		editButtonAddFaction,
		editButtonEditFaction,
		editButtonDeleteFaction,
	} {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func editShapeInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range [...]editInspectorButton{
		editButtonShapePaint,
		editButtonShapeErase,
		editButtonShapeRegionPaint,
		editButtonShapeRegionErase,
		editButtonShapeBrushMinus,
		editButtonShapeBrushPlus,
		editButtonLandPassageAdd,
		editButtonLandPassageAdjust,
		editButtonLandPassageDelete,
		editButtonAddNeighbor,
		editButtonTerrainArea,
		editButtonTerrainAreaType,
		editButtonTerrainAreaCost,
		editButtonTerrainAreaAttrition,
		editButtonTerrainAreaDelete,
	} {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func editDataInspectorButtonAt(mx, my float64) editInspectorButton {
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	return editButtonNone
}

func (r *Renderer) editInspectorActiveButtonAt(mx, my float64) editInspectorButton {
	if buildEditInspectorTabButton(editInspectorSettlement, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorRegion, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorFaction, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorMap, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorData, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape {
		kind := editShapeInspectorButtonAt(mx, my)
		if isEditShapeToolButton(kind) {
			active := r.activeEditShapeToolButton()
			if active != editButtonNone && kind != active {
				return editButtonNone
			}
			if active == editButtonNone && r.gs != nil && !r.editShapeToolButtonAvailable(kind) {
				return editButtonNone
			}
		}
		return kind
	}
	if r.editInspectorTab == editInspectorData {
		return editDataInspectorButtonAt(mx, my)
	}
	if r.editInspectorTab == editInspectorSettlement {
		return editSettlementInspectorButtonAt(mx, my)
	}
	if r.editInspectorTab == editInspectorFaction {
		return editFactionInspectorButtonAt(mx, my)
	}
	return editRegionInspectorButtonAt(mx, my)
}

const (
	editOwnerDropdownVisibleRows = 10
	editOwnerDropdownRowH        = float32(24)
	editOwnerDropdownHeaderH     = float32(30)
)

func editOwnerDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func editTerrainDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func editSettlementTypeDropdownRect() (float32, float32, float32, float32) {
	x, y, w, _ := editInspectorRect()
	dropW := float32(292)
	dropH := editOwnerDropdownHeaderH + editOwnerDropdownRowH*editOwnerDropdownVisibleRows + 10
	return x + w + 8, y, dropW, dropH
}

func (r *Renderer) updateEditDropdownPositions() {
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editOwnerDropdown.SetPosition(float64(dx), float64(dy))
	r.editSuccessorDropdown.SetPosition(float64(dx), float64(dy))
	r.editTerrainDropdown.SetPosition(float64(dx), float64(dy))
	r.editSettlementTypeDropdown.SetPosition(float64(dx), float64(dy))
	r.editUnitTypeDropdown.SetPosition(float64(dx), float64(dy))
}

func editMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (r *Renderer) drawEditRegionCenters(screen *ebiten.Image) {
	for _, region := range r.gs.Regions {
		if region == nil || region.IsLocked {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		col := color.RGBA{80, 220, 255, 190}
		if region.IsSea {
			col = color.RGBA{120, 210, 255, 210}
		}
		if region.ID == r.editSelectedRegion && r.editSelectedSettlement < 0 {
			if region.IsSea {
				col = color.RGBA{70, 235, 255, 245}
			} else {
				col = color.RGBA{255, 190, 45, 240}
			}
		}
		x, y := float32(sx), float32(sy)
		vector.StrokeCircle(screen, x, y, 6, 1.5, col, true)
		vector.StrokeLine(screen, x-8, y, x+8, y, 1.5, col, true)
		vector.StrokeLine(screen, x, y-8, x, y+8, 1.5, col, true)
	}
}

func (r *Renderer) drawEditVoronoiDebug(screen *ebiten.Image) {
	if !r.editVoronoiDebug {
		return
	}
	rid := r.editSelectedRegion
	if rid == "" {
		mx, my := ebiten.CursorPosition()
		rid = r.editRegionAt(float64(mx), float64(my))
	}
	r.editVoronoiDebugRegion = rid
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}

	r.editVisualNeighborBuf = r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf)
	r.editBoundaryPixelBuf = r.worldMap.BoundaryPixels(rid, r.editBoundaryPixelBuf)
	r.drawEditVoronoiBoundary(screen, r.editBoundaryPixelBuf)

	cx, cy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	for _, nrid := range r.editVisualNeighborBuf {
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{90, 220, 125, 205}
		if !regionHasNeighbor(region, nrid) {
			col = color.RGBA{235, 80, 80, 220}
		}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1.5, col, true)
		mx, my := (cx+nx)/2, (cy+ny)/2
		vector.FillRect(screen, float32(mx)-3, float32(my)-3, 6, 6, col, true)
	}

	for _, nrid := range region.Neighbors {
		if visualNeighborContains(r.editVisualNeighborBuf, nrid) {
			continue
		}
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{180, 180, 180, 150}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1, col, true)
	}

	vector.StrokeCircle(screen, float32(cx), float32(cy), 12, 2.5, color.RGBA{255, 220, 70, 245}, true)
}

func (r *Renderer) drawEditVoronoiLegendOverlay(screen *ebiten.Image) {
	if !r.editVoronoiDebug {
		return
	}
	r.drawEditVoronoiLegend(screen, r.editVoronoiDebugRegion, r.editVisualNeighborBuf)
}

func (r *Renderer) drawEditVoronoiBoundary(screen *ebiten.Image, pixels []int) {
	step := 1
	if r.camScale < 0.8 {
		step = 2
	}
	if r.camScale < 0.45 {
		step = 4
	}
	size := float32(2)
	if r.camScale >= 1.25 {
		size = 3
	}
	col := color.RGBA{80, 210, 255, 215}
	for i := 0; i < len(pixels); i += step {
		pIdx := pixels[i]
		wx := float64(pIdx%WorldW) + 0.5
		wy := float64(pIdx/WorldW) + 0.5
		sx, sy := r.worldToScreen(wx, wy)
		if sx < -4 || sx > ScreenWidth+4 || sy < -4 || sy > ScreenHeight+4 {
			continue
		}
		vector.FillRect(screen, float32(sx)-size/2, float32(sy)-size/2, size, size, col, true)
	}
}

func (r *Renderer) drawEditVoronoiLegend(screen *ebiten.Image, rid world.RegionID, visual []world.RegionID) {
	const panelW, panelH = float32(360), float32(104)
	x := float32(ScreenWidth) - panelW - 18
	y := float32(18)
	drawRoundedRect(screen, x, y, panelW, panelH, 8, color.RGBA{16, 20, 24, 218})
	drawPanelBorder(screen, x, y, panelW, panelH)
	DrawText(screen, "VORONOI DEBUG", float64(x)+12, float64(y)+10, FaceSmall, ColorGold)
	DrawText(screen, "camgobegi: raster sinir", float64(x)+12, float64(y)+31, FaceSmall, ColorGray)
	DrawText(screen, "yesil: gorunen+JSON   kirmizi: sadece gorunen", float64(x)+12, float64(y)+48, FaceSmall, ColorGray)

	mx, my := ebiten.CursorPosition()
	wx, wy := r.screenToWorld(float64(mx), float64(my))
	hover := r.worldMap.RegionAt(int(wx), int(wy))
	sx, sy := scenarioCoordsFromWorld(wx, wy)
	DrawText(screen, "Hover: "+string(hover)+"  "+itoa(sx)+","+itoa(sy), float64(x)+12, float64(y)+68, FaceSmall, ColorWhite)
	if rid != "" {
		region := r.gs.Regions[rid]
		jsonCount := 0
		if region != nil {
			jsonCount = len(region.Neighbors)
		}
		DrawText(screen, "Secili: "+string(rid)+"  visual/json: "+itoa(len(visual))+"/"+itoa(jsonCount),
			float64(x)+12, float64(y)+85, FaceSmall, ColorWhite)
	}
}

func regionHasNeighbor(region *world.Region, rid world.RegionID) bool {
	for _, nrid := range region.Neighbors {
		if nrid == rid {
			return true
		}
	}
	return false
}

func visualNeighborContains(neighbors []world.RegionID, rid world.RegionID) bool {
	for _, nrid := range neighbors {
		if nrid == rid {
			return true
		}
	}
	return false
}

func (r *Renderer) pushEditCommand(cmd editCommand) {
	if cmd.undo == nil || cmd.redo == nil {
		return
	}
	r.editUndoStack = append(r.editUndoStack, cmd)
	r.editRedoStack = r.editRedoStack[:0]
	r.editDirty = true
}

func (r *Renderer) undoEditCommand() {
	if len(r.editUndoStack) == 0 {
		return
	}
	last := len(r.editUndoStack) - 1
	cmd := r.editUndoStack[last]
	r.editUndoStack = r.editUndoStack[:last]
	cmd.undo(r)
	r.editRedoStack = append(r.editRedoStack, cmd)
	r.editDirty = true
}

func (r *Renderer) redoEditCommand() {
	if len(r.editRedoStack) == 0 {
		return
	}
	last := len(r.editRedoStack) - 1
	cmd := r.editRedoStack[last]
	r.editRedoStack = r.editRedoStack[:last]
	cmd.redo(r)
	r.editUndoStack = append(r.editUndoStack, cmd)
	r.editDirty = true
}

func cloneSettlements(settlements []world.Settlement) []world.Settlement {
	if settlements == nil {
		return nil
	}
	clone := make([]world.Settlement, len(settlements))
	copy(clone, settlements)
	return clone
}

func (r *Renderer) settlementSnapshot(rid world.RegionID) editRegionSettlementsSnapshot {
	region := r.gs.Regions[rid]
	if region == nil {
		return editRegionSettlementsSnapshot{Region: rid}
	}
	return editRegionSettlementsSnapshot{
		Region:             rid,
		Settlements:        cloneSettlements(region.Settlements),
		Buildings:          cloneStringSlice(region.Buildings),
		SuccessorFactionID: region.SuccessorFactionID,
	}
}

func uniqueSettlementSnapshots(snaps []editRegionSettlementsSnapshot) []editRegionSettlementsSnapshot {
	out := snaps[:0]
	for _, snap := range snaps {
		seen := false
		for _, existing := range out {
			if existing.Region == snap.Region {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, snap)
		}
	}
	return out
}

func (r *Renderer) restoreSettlementSnapshots(snaps []editRegionSettlementsSnapshot) {
	for _, snap := range snaps {
		region := r.gs.Regions[snap.Region]
		if region == nil {
			continue
		}
		region.Settlements = cloneSettlements(snap.Settlements)
		region.Buildings = cloneStringSlice(snap.Buildings)
		region.SuccessorFactionID = snap.SuccessorFactionID
	}
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editRenaming = false
	r.worldMap.RebuildSettlementAnchors(r.gs)
}

func (r *Renderer) pushSettlementSnapshots(before, after []editRegionSettlementsSnapshot, selectedRegion world.RegionID, selectedSettlement int) {
	before = uniqueSettlementSnapshots(before)
	after = uniqueSettlementSnapshots(after)
	if len(before) == 0 || len(after) == 0 || settlementSnapshotsEqual(before, after) {
		return
	}
	beforeCopy := cloneSettlementSnapshots(before)
	afterCopy := cloneSettlementSnapshots(after)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreSettlementSnapshots(beforeCopy)
			rr.editSelectedRegion = selectedRegion
			rr.editSelectedSettlement = -1
		},
		redo: func(rr *Renderer) {
			rr.restoreSettlementSnapshots(afterCopy)
			rr.editSelectedRegion = selectedRegion
			rr.editSelectedSettlement = selectedSettlement
		},
	})
}

func cloneSettlementSnapshots(snaps []editRegionSettlementsSnapshot) []editRegionSettlementsSnapshot {
	out := make([]editRegionSettlementsSnapshot, len(snaps))
	for i, snap := range snaps {
		out[i] = editRegionSettlementsSnapshot{
			Region:             snap.Region,
			Settlements:        cloneSettlements(snap.Settlements),
			Buildings:          cloneStringSlice(snap.Buildings),
			SuccessorFactionID: snap.SuccessorFactionID,
		}
	}
	return out
}

func settlementSnapshotsEqual(a, b []editRegionSettlementsSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Region != b[i].Region ||
			!settlementsEqual(a[i].Settlements, b[i].Settlements) ||
			!stringSlicesEqual(a[i].Buildings, b[i].Buildings) ||
			a[i].SuccessorFactionID != b[i].SuccessorFactionID {
			return false
		}
	}
	return true
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func settlementsEqual(a, b []world.Settlement) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func editUndoPressed() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	return ctrl && !shift
}

func editRedoPressed() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	return ctrl && shift
}

func (r *Renderer) handleEditModeInput() InputAction {
	if r.editRenaming {
		return r.handleEditRenameInput()
	}
	if r.editFactionForm.show {
		return r.handleEditFactionFormInput()
	}

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	leftPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	leftJustPressed := r.mouseJustPressed(ebiten.MouseButtonLeft)
	rightPressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	rightJustPressed := r.mouseJustPressed(ebiten.MouseButtonRight)

	if r.editOwnerDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editOwnerDropdown.HitTest(fx, fy) {
			r.editOwnerDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editSuccessorDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editSuccessorDropdown.HitTest(fx, fy) {
			r.editSuccessorDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editTerrainDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editTerrainDropdown.HitTest(fx, fy) {
			r.editTerrainDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editSettlementTypeDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editSettlementTypeDropdown.HitTest(fx, fy) {
			r.editSettlementTypeDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if r.editUnitTypeDropdown.IsOpen() {
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 && r.editUnitTypeDropdown.HitTest(fx, fy) {
			r.editUnitTypeDropdown.Scroll(wheelY)
			return InputAction{}
		}
	}

	if !r.editOwnerDropdown.IsOpen() && !r.editSuccessorDropdown.IsOpen() && !r.editTerrainDropdown.IsOpen() && !r.editSettlementTypeDropdown.IsOpen() && !r.editUnitTypeDropdown.IsOpen() {
		r.handleCamera()
	}

	if r.editShapePainting && !leftPressed {
		r.finishShapePaintStroke()
		return InputAction{}
	}

	if r.keyJustPressed(ebiten.KeyF11) {
		r.toggleFullscreen()
	}
	if r.keyJustPressed(ebiten.KeyV) {
		r.editVoronoiDebug = !r.editVoronoiDebug
	}
	if r.keyJustPressed(ebiten.KeyP) {
		r.toggleEditLandPassageMode()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyZ) {
		if editRedoPressed() {
			r.redoEditCommand()
			return InputAction{}
		}
		if editUndoPressed() {
			r.undoEditCommand()
			return InputAction{}
		}
	}
	if r.keyJustPressed(ebiten.KeyY) && editUndoPressed() {
		r.redoEditCommand()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editOwnerDropdown.Close()
		r.editSuccessorDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		r.editUnitTypeDropdown.Close()
		if r.editDirty {
			r.showEditExitConfirm()
			return InputAction{}
		}
		return InputAction{Kind: ActionGoMainMenu}
	}
	if r.keyJustPressed(ebiten.KeyS) && (ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)) {
		return InputAction{Kind: ActionSaveScenario}
	}
	if r.keyJustPressed(ebiten.KeyDelete) && r.editLandPassageAdjustMode {
		r.deleteSelectedLandPassage()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyDelete) && !r.hasEditSelection() && r.editSelectedRegion != "" {
		r.deleteSelectedRegion()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyDelete) && r.hasEditSelection() {
		r.deleteSelectedSettlement()
		return InputAction{}
	}
	if (r.keyJustPressed(ebiten.KeyF2) || r.keyJustPressed(ebiten.KeyEnter)) && r.hasEditSelection() {
		r.beginEditRename(editTextSettlementNameTR)
		return InputAction{}
	}

	if r.editLandPassageAdjustMode && r.editLandPassageDragEndpoint >= 0 {
		if leftPressed {
			r.updateEditLandPassageDrag(fx, fy)
			return InputAction{}
		}
		r.finishEditLandPassageDrag()
		return InputAction{}
	}
	if r.editDraggingSettlement {
		if rightPressed {
			r.moveSelectedSettlementTo(fx, fy)
			return InputAction{}
		}
		r.finishSettlementDrag()
		r.editDraggingSettlement = false
		return InputAction{}
	}
	inspectorOverlayOpen := r.editOwnerDropdown.IsOpen() ||
		r.editSuccessorDropdown.IsOpen() ||
		r.editTerrainDropdown.IsOpen() ||
		r.editSettlementTypeDropdown.IsOpen() ||
		r.editUnitTypeDropdown.IsOpen()
	if !r.editShapePainting && rightJustPressed && !inspectorOverlayOpen && !editInspectorHit(fx, fy) && !r.editShapeHelpPanelHit(fx, fy) {
		if rid, idx, ok := r.editSettlementAt(fx, fy); ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = idx
			r.editDraggingRegion = false
			r.editRenaming = false
			r.beginSettlementDrag(rid)
			r.editDraggingSettlement = true
			return InputAction{}
		}
	}
	if leftJustPressed {
		if action, ok := r.handleEditInspectorClick(fx, fy); ok {
			return action
		}
	}
	if r.editLandPassageMode && leftJustPressed {
		r.handleEditLandPassageClick(fx, fy)
		return InputAction{}
	}
	if r.editLandPassageAdjustMode && leftJustPressed {
		r.handleEditLandPassageAdjustClick(fx, fy)
		return InputAction{}
	}
	if r.editNeighborAddMode && leftJustPressed {
		r.handleEditNeighborAddClick(fx, fy)
		return InputAction{}
	}

	if (r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape || r.editInspectorTab == editInspectorRegion) && leftJustPressed && r.editShapeHelpPanelHit(fx, fy) {
		return InputAction{}
	}

	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape || r.editInspectorTab == editInspectorRegion {
		if leftJustPressed && r.beginShapePaintStroke(fx, fy) {
			return InputAction{}
		}
		if r.editShapePainting {
			r.continueShapePaintStroke(fx, fy)
			return InputAction{}
		}
	}

	if r.editDraggingRegion && !leftPressed {
		r.finishRegionCenterDrag()
		r.editDraggingRegion = false
		r.rebuildEditWorldMap()
	}

	if leftJustPressed {
		r.editSuccessorDropdown.Close()
		if editModifierPressed() {
			rid := r.editRegionAt(fx, fy)
			if rid != "" {
				r.editOwnerDropdown.Close()
				r.editTerrainDropdown.Close()
				r.editSettlementTypeDropdown.Close()
				r.editUnitTypeDropdown.Close()
				r.editSelectedRegion = rid
				r.setEditFactionFromRegion(rid)
				r.editSelectedSettlement = -1
				r.editDraggingRegion = true
				r.editDraggingSettlement = false
				r.editRenaming = false
				r.beginRegionCenterDrag(rid)
				r.moveSelectedRegionCenterTo(fx, fy)
				return InputAction{}
			}
		}
		if editAddModifierPressed() {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			if editCreateRegionModifierPressed() {
				r.addRegionAt(fx, fy)
				return InputAction{}
			}
			r.addSettlementAt(fx, fy)
			return InputAction{}
		}

		if aid, ok := r.editArmyAt(fx, fy); ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = aid
			if a := r.gs.Armies[aid]; a != nil {
				r.editSelectedRegion = a.RegionID
				r.syncSelectedTerrainArea(a.RegionID)
				r.setEditFactionFromArmy(a)
				r.ensureEditSelectedUnitType(a)
			}
			r.editSelectedSettlement = -1
			r.editDraggingSettlement = false
			r.editDraggingRegion = false
			r.editRenaming = false
			return InputAction{}
		}

		rid, idx, ok := r.editSettlementAt(fx, fy)
		if ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.syncSelectedTerrainArea(rid)
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = idx
			r.editDraggingSettlement = false
			r.editDraggingRegion = false
			return InputAction{}
		}
		if rid := r.editRegionAt(fx, fy); rid != "" {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.syncSelectedTerrainArea(rid)
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = -1
			r.editRenaming = false
			r.editDraggingRegion = false
			r.editDraggingSettlement = false
			return InputAction{}
		}
		r.editOwnerDropdown.Close()
		r.editTerrainDropdown.Close()
		r.editSettlementTypeDropdown.Close()
		r.editUnitTypeDropdown.Close()
		r.SelectedArmy = ""
		r.editSelectedRegion = ""
		r.editSelectedFaction = ""
		r.editSelectedSettlement = -1
		r.editRenaming = false
		r.editDraggingRegion = false
	}

	if r.editDraggingRegion {
		r.moveSelectedRegionCenterTo(fx, fy)
		return InputAction{}
	}

	return InputAction{}
}

func (r *Renderer) syncSelectedTerrainArea(rid world.RegionID) {
	r.editTerrainAreaSelected = -1
	region := r.gs.Regions[rid]
	if region == nil || !region.IsTerrainArea {
		return
	}
	for i := range r.gs.TerrainAreas {
		if r.gs.TerrainAreas[i].ID == region.TerrainAreaID {
			r.editTerrainAreaSelected = i
			r.editTerrainAreaMoveCost = r.gs.TerrainAreas[i].MoveCost
			r.editTerrainAreaAttritionCost = r.gs.TerrainAreas[i].AttritionCost
			return
		}
	}
}

func (r *Renderer) handleEditInspectorClick(fx, fy float64) (InputAction, bool) {
	if r.editOwnerDropdown.IsOpen() {
		if idx, ok := r.editOwnerDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedRegionOwner(r.editOwnerDropdown.OptionAt(idx))
			r.editOwnerDropdown.Close()
			return InputAction{}, true
		}
		if r.editOwnerDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editOwnerDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editSuccessorDropdown.IsOpen() {
		if idx, ok := r.editSuccessorDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedRegionSuccessor(r.editSuccessorDropdown.OptionAt(idx))
			r.editSuccessorDropdown.Close()
			return InputAction{}, true
		}
		if r.editSuccessorDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editSuccessorDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editTerrainDropdown.IsOpen() {
		if idx, ok := r.editTerrainDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedRegionTerrain(world.TerrainType(r.editTerrainDropdown.OptionAt(idx)))
			r.editTerrainDropdown.Close()
			return InputAction{}, true
		}
		if r.editTerrainDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editTerrainDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editSettlementTypeDropdown.IsOpen() {
		if idx, ok := r.editSettlementTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedSettlementType(r.editSettlementTypeDropdown.OptionAt(idx))
			r.editSettlementTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editSettlementTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editSettlementTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	if r.editUnitTypeDropdown.IsOpen() {
		if idx, ok := r.editUnitTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedEditArmyUnitType(r.editUnitTypeDropdown.OptionAt(idx))
			r.editUnitTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editUnitTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editUnitTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	if !editInspectorHit(fx, fy) {
		return InputAction{}, false
	}
	if buildEditInspectorTabButton(editInspectorSettlement, "Yerleşim Birimi").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorSettlement
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorRegion, "Bölge").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorRegion
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorFaction, "Devlet").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorFaction
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorMap, "Harita").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorMap
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorData, "Veri").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorData
		return InputAction{}, true
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(fx, fy) {
		return InputAction{Kind: ActionSaveScenario}, true
	}
	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape {
		return r.handleEditShapeInspectorClick(fx, fy)
	}
	if r.editInspectorTab == editInspectorData {
		return r.handleEditDataInspectorClick(fx, fy)
	}
	if r.editInspectorTab == editInspectorFaction {
		return r.handleEditFactionInspectorClick(fx, fy)
	}
	if r.editInspectorTab == editInspectorSettlement {
		switch editSettlementInspectorButtonAt(fx, fy) {
		case editButtonAddSettlement:
			r.addSettlementToSelectedRegion()
		case editButtonSettlementType:
			if r.hasEditSelection() {
				r.toggleEditSettlementTypeDropdown()
			}
		case editButtonSetCenterSettlement:
			if r.hasEditSelection() {
				r.setSelectedSettlementCapital()
			}
		case editButtonRenameSettlement:
			if r.hasEditSelection() {
				r.beginEditRename(editTextSettlementNameTR)
			}
		case editButtonDeleteSettlement:
			if r.hasEditSelection() {
				r.deleteSelectedSettlement()
			}
		case editButtonAddArmy:
			r.addEditLandArmy()
		case editButtonAddFleet:
			r.addEditFleet()
		case editButtonDeleteArmy:
			r.deleteSelectedArmy()
		case editButtonArmyUnitType:
			r.toggleEditUnitTypeDropdown()
		case editButtonArmyUnitMinus:
			r.removeSelectedArmyUnit()
		case editButtonArmyUnitPlus:
			r.addSelectedArmyUnit()
		case editButtonArmyOwnerFromRegion:
			r.setSelectedArmyOwnerFromRegion()
		}
		return InputAction{}, true
	}
	if r.hasEditSelection() && buildEditInspectorActionButton(editButtonSetFactionCapital, "").HitTest(fx, fy) {
		r.setSelectedFactionCapital()
		return InputAction{}, true
	}
	switch editRegionInspectorButtonAt(fx, fy) {
	case editButtonRegionTerrain:
		r.toggleEditTerrainDropdown()
	case editButtonRegionNameTR:
		r.beginEditRename(editTextRegionNameTR)
	case editButtonRegionName:
		r.beginEditRename(editTextRegionName)
	case editButtonRegionID:
		r.beginEditRename(editTextRegionID)
	case editButtonRegionLock:
		r.toggleSelectedRegionLock()
	case editButtonUnlockMinus:
		r.adjustSelectedRegionUnlockTurn(-10)
	case editButtonUnlockPlus:
		r.adjustSelectedRegionUnlockTurn(10)
	case editButtonSyncNeighbors:
		r.syncSelectedRegionNeighborsFromVisual()
	case editButtonAddNeighbor:
		r.toggleEditNeighborAddMode()
	case editButtonAddRegion:
		r.addRegionNearSelected()
	case editButtonDeleteRegion:
		r.deleteSelectedRegion()
	case editButtonDeleteSettlement:
		if r.hasEditSelection() {
			r.deleteSelectedSettlement()
		}
	}
	return InputAction{}, true
}

func (r *Renderer) handleEditFactionInspectorClick(fx, fy float64) (InputAction, bool) {
	switch editFactionInspectorButtonAt(fx, fy) {
	case editButtonRegionOwner:
		r.toggleEditOwnerDropdown()
	case editButtonRegionSuccessor:
		r.toggleEditSuccessorDropdown()
	case editButtonSetFactionCapital:
		r.setSelectedFactionCapital()
	case editButtonAddFaction:
		r.openFactionCreateForm()
	case editButtonEditFaction:
		r.openFactionEditForm()
	case editButtonDeleteFaction:
		r.deleteSelectedFaction()
	}
	return InputAction{}, true
}

func (r *Renderer) handleEditDataInspectorClick(fx, fy float64) (InputAction, bool) {
	if r.editUnitTypeDropdown.IsOpen() {
		if idx, ok := r.editUnitTypeDropdown.GetSelectedOption(fx, fy); ok {
			r.setSelectedEditArmyUnitType(r.editUnitTypeDropdown.OptionAt(idx))
			r.editUnitTypeDropdown.Close()
			return InputAction{}, true
		}
		if r.editUnitTypeDropdown.HitTest(fx, fy) {
			return InputAction{}, true
		}
		if !editInspectorHit(fx, fy) {
			r.editUnitTypeDropdown.Close()
			return InputAction{}, false
		}
	}
	switch editDataInspectorButtonAt(fx, fy) {
	case editButtonAddFaction:
		r.openFactionCreateForm()
	case editButtonEditFaction:
		r.openFactionEditForm()
	case editButtonDeleteFaction:
		r.deleteSelectedFaction()
	case editButtonAddArmy:
		r.addEditLandArmy()
	case editButtonAddFleet:
		r.addEditFleet()
	case editButtonDeleteArmy:
		r.deleteSelectedArmy()
	case editButtonArmyUnitType:
		r.toggleEditUnitTypeDropdown()
	case editButtonArmyUnitMinus:
		r.removeSelectedArmyUnit()
	case editButtonArmyUnitPlus:
		r.addSelectedArmyUnit()
	case editButtonArmyOwnerFromRegion:
		r.setSelectedArmyOwnerFromRegion()
	case editButtonSaveScenario:
		return InputAction{Kind: ActionSaveScenario}, true
	}
	return InputAction{}, true
}

func (r *Renderer) toggleEditOwnerDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		r.editOwnerDropdown.Close()
		return
	}

	r.editSuccessorDropdown.Close()
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editOwnerDropdown.SetPosition(float64(dx), float64(dy))
	r.editOwnerDropdown.SetOptions(editOwnerOptions(r.gs.Factions), region.OwnerID)
	r.editOwnerDropdown.Toggle()
}

func (r *Renderer) toggleEditSuccessorDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea {
		r.editSuccessorDropdown.Close()
		return
	}
	r.editOwnerDropdown.Close()
	r.editTerrainDropdown.Close()
	r.editSettlementTypeDropdown.Close()
	r.editUnitTypeDropdown.Close()
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editSuccessorDropdown.SetPosition(float64(dx), float64(dy))
	r.editSuccessorDropdown.SetOptions(editOwnerOptions(r.gs.Factions), region.SuccessorFactionID)
	r.editSuccessorDropdown.Toggle()
}

func (r *Renderer) toggleEditTerrainDropdown() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		r.editTerrainDropdown.Close()
		return
	}

	dx, dy, _, _ := editTerrainDropdownRect()
	r.editTerrainDropdown.SetPosition(float64(dx), float64(dy))
	terrainOptions := editRegionTerrainOptions()
	if region.IsTerrainArea {
		terrainOptions = editTerrainAreaOptions()
	}
	stringOptions := make([]string, len(terrainOptions))
	for i, t := range terrainOptions {
		stringOptions[i] = string(t)
	}
	r.editTerrainDropdown.SetOptions(stringOptions, string(region.Terrain))
	r.editTerrainDropdown.Toggle()
}

func (r *Renderer) toggleEditSettlementTypeDropdown() {
	if !r.hasEditSelection() {
		r.editSettlementTypeDropdown.Close()
		return
	}

	dx, dy, _, _ := editSettlementTypeDropdownRect()
	r.editSettlementTypeDropdown.SetPosition(float64(dx), float64(dy))
	region := r.gs.Regions[r.editSelectedRegion]
	settlement := region.Settlements[r.editSelectedSettlement]
	r.editSettlementTypeDropdown.SetOptions(world.AllSettlementTypes(), string(settlement.Type))
	r.editSettlementTypeDropdown.Toggle()
}

func (r *Renderer) hasEditSelection() bool {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	return ok && region != nil && r.editSelectedSettlement >= 0 &&
		r.editSelectedSettlement < len(region.Settlements)
}

func (r *Renderer) beginEditRename(target editTextTarget) {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	switch target {
	case editTextSettlementNameTR:
		if !r.hasEditSelection() {
			return
		}
	case editTextRegionNameTR:
	case editTextRegionName:
	case editTextRegionID:
	default:
		return
	}
	r.editTextTarget = target
	r.editTextError = ""
	r.editTextRunes = r.editTextRunes[:0]
	if target == editTextRegionID {
		r.editTextRunes = append(r.editTextRunes, []rune(string(region.ID))...)
	}
	r.editRenaming = true
	r.editDraggingSettlement = false
}

func (r *Renderer) handleEditRenameInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editRenaming = false
		r.editTextTarget = editTextNone
		r.editTextError = ""
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		r.commitEditRename()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(r.editTextRunes) > 0 {
		r.editTextRunes = r.editTextRunes[:len(r.editTextRunes)-1]
	}
	if r.editTextTarget == editTextRegionID && r.keyJustPressed(ebiten.KeyA) && editUndoPressed() {
		r.editTextRunes = r.editTextRunes[:0]
		return InputAction{}
	}
	r.editTextRunes = ebiten.AppendInputChars(r.editTextRunes)
	if len(r.editTextRunes) > 64 {
		r.editTextRunes = r.editTextRunes[:64]
	}
	return InputAction{}
}

func (r *Renderer) commitEditRename() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		r.editRenaming = false
		r.editTextTarget = editTextNone
		r.editTextError = ""
		return
	}
	newName := strings.TrimSpace(string(r.editTextRunes))
	rid := region.ID
	switch r.editTextTarget {
	case editTextRegionID:
		newID := world.RegionID(newName)
		if newID == "" {
			r.editTextError = "ID bos olamaz."
			return
		}
		if strings.IndexFunc(newName, unicode.IsSpace) >= 0 {
			r.editTextError = "ID bosluk icermemeli."
			return
		}
		if newID != rid && r.gs.Regions[newID] != nil {
			r.editTextError = "Bu region ID zaten var."
			return
		}
		if newID != rid {
			before := r.worldSnapshot()
			r.renameRegionID(rid, newID)
			after := r.worldSnapshot()
			r.pushWorldSnapshotCommand(before, after)
		}
		r.editRenaming = false
		r.editTextTarget = editTextNone
		r.editTextError = ""
		return
	case editTextSettlementNameTR:
		if !r.hasEditSelection() {
			break
		}
		idx := r.editSelectedSettlement
		oldName := region.Settlements[idx].NameTR
		if newName != "" && oldName != newName {
			region.Settlements[idx].NameTR = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) {
					rr.setSettlementNameTR(rid, idx, oldName)
				},
				redo: func(rr *Renderer) {
					rr.setSettlementNameTR(rid, idx, newName)
				},
			})
			r.editDirty = true
		}
	case editTextRegionNameTR:
		oldName := region.NameTR
		if newName != "" && oldName != newName {
			region.NameTR = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) { rr.setRegionNameTR(rid, oldName) },
				redo: func(rr *Renderer) { rr.setRegionNameTR(rid, newName) },
			})
			r.editDirty = true
		}
	case editTextRegionName:
		oldName := region.Name
		if newName != "" && oldName != newName {
			region.Name = newName
			r.pushEditCommand(editCommand{
				undo: func(rr *Renderer) { rr.setRegionName(rid, oldName) },
				redo: func(rr *Renderer) { rr.setRegionName(rid, newName) },
			})
			r.editDirty = true
		}
	}
	r.editRenaming = false
	r.editTextTarget = editTextNone
	r.editTextError = ""
}

func (r *Renderer) editTextLabel() string {
	switch r.editTextTarget {
	case editTextRegionNameTR:
		return "Bolge Ad TR"
	case editTextRegionName:
		return "Bolge Ad EN"
	case editTextRegionID:
		return "Bolge ID"
	default:
		return "Isim"
	}
}

func (r *Renderer) editSettlementAt(fx, fy float64) (world.RegionID, int, bool) {
	bestDist := float64(14 * 14)
	var bestRegion world.RegionID
	bestIndex := -1
	for rid, region := range r.gs.Regions {
		if region == nil || region.IsSea {
			continue
		}
		for i := range region.Settlements {
			ax, ay, ok := r.worldMap.SettlementAnchor(rid, i)
			if !ok {
				continue
			}
			sx, sy := r.worldToScreen(float64(ax), float64(ay))
			dx, dy := fx-sx, fy-sy
			dist := dx*dx + dy*dy
			if dist <= bestDist {
				bestDist = dist
				bestRegion = rid
				bestIndex = i
			}
		}
	}
	return bestRegion, bestIndex, bestIndex >= 0
}

func (r *Renderer) editRegionAt(fx, fy float64) world.RegionID {
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	if region, ok := r.gs.Regions[rid]; ok && region != nil {
		return rid
	}
	return ""
}

func (r *Renderer) editArmyAt(fx, fy float64) (army.ArmyID, bool) {
	armyPositions := r.armyIconPositions()
	for i := len(armyPositions) - 1; i >= 0; i-- {
		pos := armyPositions[i]
		dx := fx - float64(pos.X)
		dy := fy - float64(pos.Y)
		if dx*dx+dy*dy < 14*14 {
			return pos.ArmyID, true
		}
	}
	return "", false
}

func (r *Renderer) beginRegionCenterDrag(rid world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil || region.IsTerrainArea {
		r.editRegionDragStart = nil
		return
	}
	r.editRegionDragStart = &editRegionCenterSnapshot{
		Region: rid,
		X:      region.WorldX,
		Y:      region.WorldY,
	}
}

func (r *Renderer) finishRegionCenterDrag() {
	start := r.editRegionDragStart
	r.editRegionDragStart = nil
	if start == nil {
		return
	}
	region := r.gs.Regions[start.Region]
	if region == nil || region.IsTerrainArea || (region.WorldX == start.X && region.WorldY == start.Y) {
		return
	}
	begin := *start
	end := editRegionCenterSnapshot{Region: start.Region, X: region.WorldX, Y: region.WorldY}
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreRegionCenter(begin)
		},
		redo: func(rr *Renderer) {
			rr.restoreRegionCenter(end)
		},
	})
}

func (r *Renderer) restoreRegionCenter(snapshot editRegionCenterSnapshot) {
	region := r.gs.Regions[snapshot.Region]
	if region == nil || region.IsTerrainArea {
		return
	}
	region.WorldX = snapshot.X
	region.WorldY = snapshot.Y
	r.editSelectedRegion = snapshot.Region
	r.editSelectedSettlement = -1
	r.editDraggingRegion = false
	r.editDraggingSettlement = false
	r.rebuildEditWorldMap()
}

func (r *Renderer) beginSettlementDrag(rid world.RegionID) {
	r.editSettlementDragStart = r.editSettlementDragStart[:0]
	r.editSettlementDragStart = append(r.editSettlementDragStart, r.settlementSnapshot(rid))
}

func (r *Renderer) ensureSettlementDragSnapshot(rid world.RegionID) {
	for _, snap := range r.editSettlementDragStart {
		if snap.Region == rid {
			return
		}
	}
	r.editSettlementDragStart = append(r.editSettlementDragStart, r.settlementSnapshot(rid))
}

func (r *Renderer) finishSettlementDrag() {
	if len(r.editSettlementDragStart) == 0 {
		return
	}
	before := cloneSettlementSnapshots(r.editSettlementDragStart)
	after := make([]editRegionSettlementsSnapshot, 0, len(before)+1)
	for _, snap := range before {
		after = append(after, r.settlementSnapshot(snap.Region))
	}
	if r.editSelectedRegion != "" {
		after = append(after, r.settlementSnapshot(r.editSelectedRegion))
	}
	r.pushSettlementSnapshots(before, after, r.editSelectedRegion, r.editSelectedSettlement)
	r.editSettlementDragStart = r.editSettlementDragStart[:0]
}

func (r *Renderer) moveSelectedSettlementTo(fx, fy float64) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || r.editSelectedSettlement < 0 ||
		r.editSelectedSettlement >= len(region.Settlements) {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	newX, newY := scenarioCoordsFromWorld(wx, wy)
	targetRegionID := r.worldMap.RegionAt(int(wx), int(wy))
	if targetRegion, ok := r.gs.Regions[targetRegionID]; ok && targetRegion != nil &&
		!targetRegion.IsSea && targetRegion.ID != region.ID {
		r.transferSelectedSettlement(targetRegion.ID, newX, newY)
		return
	}
	region.Settlements[r.editSelectedSettlement].X = newX
	region.Settlements[r.editSelectedSettlement].Y = newY
	r.worldMap.UpdateSettlementAnchor(r.gs, r.editSelectedRegion, r.editSelectedSettlement)
	r.editDirty = true
}

func (r *Renderer) moveSelectedRegionCenterTo(fx, fy float64) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsTerrainArea {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	newX, newY := scenarioCoordsFromWorld(wx, wy)
	if region.WorldX == newX && region.WorldY == newY {
		return
	}
	region.WorldX = newX
	region.WorldY = newY
	r.editDirty = true
}

func (r *Renderer) addSettlementAt(fx, fy float64) {
	wx, wy := r.screenToWorld(fx, fy)
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	x, y := scenarioCoordsFromWorld(wx, wy)
	r.addSettlement(rid, x, y)
}

func (r *Renderer) addSettlementToSelectedRegion() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || !canAddSettlementToRegion(region) {
		return
	}
	r.addSettlement(region.ID, region.WorldX, region.WorldY)
}

func (r *Renderer) addSettlement(rid world.RegionID, x, y int) {
	region, ok := r.gs.Regions[rid]
	if !ok || !canAddSettlementToRegion(region) {
		return
	}
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}

	name := region.NameTR
	if name == "" {
		name = region.Name
	}
	if len(region.Settlements) > 0 {
		name += " " + itoa(len(region.Settlements)+1)
	}
	settlement := world.Settlement{
		ID:       nextSettlementID(region),
		NameTR:   name,
		X:        x,
		Y:        y,
		Type:     "city",
		IsCenter: len(region.Settlements) == 0,
	}
	region.Settlements = append(region.Settlements, settlement)
	region.RecalculatePopulation()
	syncRegionSuccessorToOwner(region)
	world.EnsureRequiredSettlementBuildings(region, r.gs.IsCapitalRegion(region))
	r.editSelectedRegion = rid
	r.editSelectedSettlement = len(region.Settlements) - 1
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.worldMap.UpdateSettlementAnchor(r.gs, rid, r.editSelectedSettlement)
	r.editDirty = true
	after := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
	r.pushSettlementSnapshots(before, after, rid, r.editSelectedSettlement)
}

func (r *Renderer) deleteSelectedSettlement() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	rid := region.ID
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
	removedCapital := region.Settlements[r.editSelectedSettlement].IsCenter
	removedPopulation := region.Settlements[r.editSelectedSettlement].Population
	region.Settlements = append(region.Settlements[:r.editSelectedSettlement], region.Settlements[r.editSelectedSettlement+1:]...)
	region.RuralPopulation += removedPopulation
	region.RecalculatePopulation()
	if removedCapital {
		ensurePrimarySettlement(region)
		syncRegionSuccessorToOwner(region)
	}
	r.editSelectedSettlement = -1
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.worldMap.RebuildSettlementAnchors(r.gs)
	r.editDirty = true
	after := []editRegionSettlementsSnapshot{r.settlementSnapshot(rid)}
	r.pushSettlementSnapshots(before, after, rid, -1)
}

func (r *Renderer) addRegionAt(fx, fy float64) {
	sourceID := r.editRegionAt(fx, fy)
	if sourceID == "" {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	x, y := scenarioCoordsFromWorld(wx, wy)
	r.addRegionFromSource(sourceID, x, y)
}

func (r *Renderer) addRegionNearSelected() {
	source := r.gs.Regions[r.editSelectedRegion]
	if source == nil {
		return
	}
	r.addRegionFromSource(source.ID, source.WorldX+12, source.WorldY+12)
}

func (r *Renderer) addRegionFromSource(sourceID world.RegionID, x, y int) {
	source := r.gs.Regions[sourceID]
	if source == nil {
		return
	}
	before := r.worldSnapshot()
	rid := nextRegionID(r.gs)
	nameNo := itoa(len(r.gs.Regions) + 1)
	region := &world.Region{
		ID:                 rid,
		Name:               "New Region " + nameNo,
		NameTR:             "Yeni Bolge " + nameNo,
		Terrain:            source.Terrain,
		OwnerID:            source.OwnerID,
		SuccessorFactionID: source.SuccessorFactionID,
		WorldX:             x,
		WorldY:             y,
		ShapeID:            source.ShapeID,
		IsSea:              source.IsSea,
		IsLocked:           source.IsLocked,
		UnlockTurn:         source.UnlockTurn,
		BaseGoldIncome:     source.BaseGoldIncome,
		BaseGrainOutput:    source.BaseGrainOutput,
		BaseIronOutput:     source.BaseIronOutput,
		BaseTimberOutput:   source.BaseTimberOutput,
		BaseSpiceOutput:    source.BaseSpiceOutput,
		BaseClothOutput:    source.BaseClothOutput,
		TradeCapacity:      source.TradeCapacity,
		Satisfaction:       source.Satisfaction,
		TaxRate:            source.TaxRate,
		Population:         source.Population,
		RuralPopulation:    source.RuralPopulation,
		Religion:           source.Religion,
		ActiveEventID:      source.ActiveEventID,
		Buildings:          cloneStringSlice(source.Buildings),
	}
	if region.Terrain == "" {
		if region.IsSea {
			region.Terrain = world.TerrainSea
		} else {
			region.Terrain = world.TerrainPlain
		}
	}
	if region.Satisfaction == 0 {
		region.Satisfaction = 70
	}
	if region.TaxRate == 0 {
		region.TaxRate = 45
	}
	r.gs.Regions[rid] = region
	r.insertRegionOrderAfter(sourceID, rid)
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
	r.SelectedArmy = ""
	r.rebuildEditWorldMap()
	visual := r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf[:0])
	r.applyVisualNeighbors(rid, visual)
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) deleteSelectedRegion() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	if region.IsTerrainArea {
		r.deleteSelectedTerrainArea()
		return
	}
	before := r.worldSnapshot()
	rid := region.ID
	for _, other := range r.gs.Regions {
		removeNeighborID(other, rid)
	}
	delete(r.gs.Regions, rid)
	r.removeRegionFromOrder(rid)
	for aid, a := range r.gs.Armies {
		if a != nil && a.RegionID == rid {
			r.gs.RemoveArmy(aid)
		}
	}
	r.editSelectedRegion = ""
	r.editSelectedSettlement = -1
	r.SelectedArmy = ""
	r.rebuildEditWorldMap()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementCapital() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	before := []editRegionSettlementsSnapshot{r.settlementSnapshot(region.ID)}
	changed := false
	for i := range region.Settlements {
		isCapital := i == r.editSelectedSettlement
		if region.Settlements[i].IsCenter != isCapital {
			region.Settlements[i].IsCenter = isCapital
			changed = true
		}
	}
	if changed {
		r.worldMap.RebuildSettlementAnchors(r.gs)
	}
	successorChanged := syncRegionSuccessorToOwner(region)
	infrastructureChanged := world.EnsureRequiredSettlementBuildings(region, r.gs.IsCapitalRegion(region))
	if changed || successorChanged || infrastructureChanged {
		r.editDirty = true
		after := []editRegionSettlementsSnapshot{r.settlementSnapshot(region.ID)}
		r.pushSettlementSnapshots(before, after, region.ID, r.editSelectedSettlement)
	}
}

func (r *Renderer) canSetSelectedFactionCapital() bool {
	if !r.hasEditSelection() {
		return false
	}
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil || region.IsSea || region.OwnerID == "" {
		return false
	}
	return r.gs.Factions[faction.FactionID(region.OwnerID)] != nil
}

func (r *Renderer) setSelectedFactionCapital() {
	if !r.canSetSelectedFactionCapital() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	settlement := &region.Settlements[r.editSelectedSettlement]
	fid := faction.FactionID(region.OwnerID)
	f := r.gs.Factions[fid]
	before := r.worldSnapshot()
	if f.CapitalSettlementID == settlement.ID && f.PendingCapitalSettlementID == "" {
		successorChanged := setRegionSuccessorToOwner(region)
		infrastructureChanged := world.EnsureRequiredSettlementBuildings(region, true)
		if !successorChanged && !infrastructureChanged {
			return
		}
		after := r.worldSnapshot()
		r.pushWorldSnapshotCommand(before, after)
		return
	}

	if !r.gs.SetFactionCapital(fid, settlement.ID) {
		return
	}
	setRegionSuccessorToOwner(region)
	world.EnsureRequiredSettlementBuildings(region, true)
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
}

func (r *Renderer) setSelectedRegionTerrain(terrain world.TerrainType) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		return
	}
	if region.Terrain == terrain {
		return
	}
	if region.IsTerrainArea {
		for i := range r.gs.TerrainAreas {
			if r.gs.TerrainAreas[i].ID != region.TerrainAreaID {
				continue
			}
			before := r.worldSnapshot()
			r.gs.TerrainAreas[i].Terrain = terrain
			r.rebuildEditWorldMap()
			after := r.worldSnapshot()
			r.pushWorldSnapshotCommand(before, after)
			r.editDirty = true
			return
		}
	}
	rid := region.ID
	old := region.Terrain
	region.Terrain = terrain
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.setRegionTerrainValue(rid, old)
		},
		redo: func(rr *Renderer) {
			rr.setRegionTerrainValue(rid, terrain)
		},
	})
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementType(typ string) {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	settlement := &region.Settlements[r.editSelectedSettlement]
	st := world.SettlementType(typ)
	if settlement.Type == st {
		return
	}
	before := r.worldSnapshot()
	settlement.Type = st
	world.EnsureRequiredSettlementBuildings(region, r.gs.IsCapitalRegion(region))
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionOwner(ownerID string) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		return
	}
	if region.OwnerID == ownerID {
		return
	}
	rid := region.ID
	old := region.OwnerID
	region.OwnerID = ownerID
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.worldMap.MarkDirty()
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.setRegionOwnerValue(rid, old)
		},
		redo: func(rr *Renderer) {
			rr.setRegionOwnerValue(rid, ownerID)
		},
	})
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionSuccessor(successorID string) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea || region.SuccessorFactionID == successorID {
		return
	}
	rid := region.ID
	old := region.SuccessorFactionID
	region.SuccessorFactionID = successorID
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setRegionSuccessorValue(rid, old) },
		redo: func(rr *Renderer) { rr.setRegionSuccessorValue(rid, successorID) },
	})
}

func (r *Renderer) setRegionSuccessorValue(rid world.RegionID, successorID string) {
	region := r.gs.Regions[rid]
	if region == nil || region.IsTerrainArea {
		return
	}
	region.SuccessorFactionID = successorID
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setSettlementNameTR(rid world.RegionID, index int, name string) {
	region := r.gs.Regions[rid]
	if region == nil || index < 0 || index >= len(region.Settlements) {
		return
	}
	region.Settlements[index].NameTR = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = index
}

func (r *Renderer) setRegionNameTR(rid world.RegionID, name string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.NameTR = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setRegionName(rid world.RegionID, name string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.Name = name
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

// renameRegionID, editorde bir bölgenin map anahtarını değiştirirken aynı ID'yi
// taşıyan editör/runtime referanslarını da birlikte günceller. Senaryo
// kayıtlarının regions, settlements, land_passages, armies, AI stratejileri,
// ticaret merkezleri ve region_shapes
// dosyaları bu state alanlarından üretildiği için bu alanların ayrışmasına izin
// verilmez.
func (r *Renderer) renameRegionID(oldID, newID world.RegionID) {
	if r == nil || r.gs == nil || oldID == "" || newID == "" || oldID == newID {
		return
	}
	region := r.gs.Regions[oldID]
	if region == nil || r.gs.Regions[newID] != nil {
		return
	}

	delete(r.gs.Regions, oldID)
	region.ID = newID
	r.gs.Regions[newID] = region

	for _, candidate := range r.gs.Regions {
		if candidate == nil {
			continue
		}
		for i, neighborID := range candidate.Neighbors {
			if neighborID == oldID {
				candidate.Neighbors[i] = newID
			}
		}
	}
	for i, rid := range r.gs.RegionOrder {
		if rid == oldID {
			r.gs.RegionOrder[i] = newID
		}
	}
	for i := range r.gs.LandPassages {
		if r.gs.LandPassages[i].From == oldID {
			r.gs.LandPassages[i].From = newID
		}
		if r.gs.LandPassages[i].To == oldID {
			r.gs.LandPassages[i].To = newID
		}
	}
	for _, a := range r.gs.Armies {
		if a == nil {
			continue
		}
		if a.RegionID == oldID {
			a.RegionID = newID
		}
		if a.DockedRegionID == oldID {
			a.DockedRegionID = newID
		}
	}
	for factionID, strategy := range r.gs.AIStrategies {
		for j := range strategy.TerritorialClaims {
			if strategy.TerritorialClaims[j].RegionID == string(oldID) {
				strategy.TerritorialClaims[j].RegionID = string(newID)
			}
		}
		for i := range strategy.Objectives {
			replaceRegionIDInStrings(strategy.Objectives[i].TargetRegions, oldID, newID)
			replaceRegionIDInStrings(strategy.Objectives[i].ReadinessRegions, oldID, newID)
			for j := range strategy.Objectives[i].TerritorialClaims {
				if strategy.Objectives[i].TerritorialClaims[j].RegionID == string(oldID) {
					strategy.Objectives[i].TerritorialClaims[j].RegionID = string(newID)
				}
			}
		}
		r.gs.AIStrategies[factionID] = strategy
	}
	for i := range r.gs.TradeCenters.Centers {
		center := &r.gs.TradeCenters.Centers[i]
		if center.ID == oldID {
			center.ID = newID
		}
		for j := range center.Links {
			if center.Links[j] == oldID {
				center.Links[j] = newID
			}
		}
	}
	for pIdx, rid := range r.gs.RegionPaintOverrides {
		if rid == oldID {
			r.gs.RegionPaintOverrides[pIdx] = newID
		}
	}
	for pIdx, rid := range r.editRegionPaintOverrides {
		if rid == oldID {
			r.editRegionPaintOverrides[pIdx] = newID
		}
	}

	if r.editSelectedRegion == oldID {
		r.editSelectedRegion = newID
	}
	if r.SelectedRegion == oldID {
		r.SelectedRegion = newID
	}
	if r.selectedSettlementRegion == oldID {
		r.selectedSettlementRegion = newID
	}
	if r.lastMapRegionClickID == oldID {
		r.lastMapRegionClickID = newID
	}
	if r.editLandPassageFrom == oldID {
		r.editLandPassageFrom = newID
	}
	if r.editNeighborAddFrom == oldID {
		r.editNeighborAddFrom = newID
	}
	r.rebuildEditWorldMap()
}

func (r *Renderer) toggleSelectedRegionLock() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	rid := region.ID
	old := region.IsLocked
	region.IsLocked = !region.IsLocked
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setRegionLockValue(rid, old) },
		redo: func(rr *Renderer) { rr.setRegionLockValue(rid, !old) },
	})
	r.editDirty = true
}

func (r *Renderer) setRegionLockValue(rid world.RegionID, locked bool) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.IsLocked = locked
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) adjustSelectedRegionUnlockTurn(delta int) {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	old := region.UnlockTurn
	next := old + delta
	if next < 0 {
		next = 0
	}
	if old == next {
		return
	}
	rid := region.ID
	region.UnlockTurn = next
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setRegionUnlockTurn(rid, old) },
		redo: func(rr *Renderer) { rr.setRegionUnlockTurn(rid, next) },
	})
	r.editDirty = true
}

func (r *Renderer) setRegionUnlockTurn(rid world.RegionID, turn int) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.UnlockTurn = turn
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setSettlementTypeValue(rid world.RegionID, index int, typ world.SettlementType) {
	region := r.gs.Regions[rid]
	if region == nil || index < 0 || index >= len(region.Settlements) {
		return
	}
	region.Settlements[index].Type = typ
	r.editSelectedRegion = rid
	r.editSelectedSettlement = index
}

func (r *Renderer) setRegionTerrainValue(rid world.RegionID, terrain world.TerrainType) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.Terrain = terrain
	r.editSelectedRegion = rid
	r.editSelectedSettlement = -1
}

func (r *Renderer) setRegionOwnerValue(rid world.RegionID, ownerID string) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	region.OwnerID = ownerID
	r.editSelectedRegion = rid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedSettlement = -1
	r.worldMap.MarkDirty()
}

func (r *Renderer) syncSelectedRegionNeighborsFromVisual() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	visual := r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
	before := r.neighborSnapshot(region.ID, visual)
	r.applyVisualNeighbors(region.ID, visual)
	after := r.neighborSnapshot(region.ID, visual)
	if neighborSnapshotsEqual(before, after) {
		return
	}
	rid := region.ID
	beforeCopy := cloneNeighborSnapshots(before)
	afterCopy := cloneNeighborSnapshots(after)
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(beforeCopy)
			rr.editSelectedRegion = rid
			rr.editSelectedSettlement = -1
		},
		redo: func(rr *Renderer) {
			rr.restoreNeighborSnapshots(afterCopy)
			rr.editSelectedRegion = rid
			rr.editSelectedSettlement = -1
		},
	})
	r.editDirty = true
}

func (r *Renderer) worldSnapshot() editWorldSnapshot {
	return editWorldSnapshot{
		Regions:              cloneRegionMap(r.gs.Regions),
		RegionOrder:          cloneRegionIDSlice(r.gs.RegionOrder),
		LandPassages:         cloneLandPassages(r.gs.LandPassages),
		Factions:             cloneFactionMap(r.gs.Factions),
		AIStrategies:         cloneAIStrategyMap(r.gs.AIStrategies),
		TradeCenters:         cloneTradeCenterConfig(r.gs.TradeCenters),
		Armies:               cloneArmyMap(r.gs.Armies),
		Relations:            cloneRelationMap(r.gs.Relations),
		ShapeData:            cloneCountryShapeJSON(r.gs.ShapeData),
		RegionPaintOverrides: cloneRegionPaintOverrides(r.editRegionPaintOverrides),
		TerrainAreas:         cloneTerrainAreas(r.gs.TerrainAreas),
		Selected:             r.editSelectedRegion,
		Settlement:           r.editSelectedSettlement,
		Faction:              r.editSelectedFaction,
		Army:                 r.SelectedArmy,
		Player:               r.gs.PlayerFactionID,
	}
}

func (r *Renderer) pushWorldSnapshotCommand(before, after editWorldSnapshot) {
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.restoreWorldSnapshot(before) },
		redo: func(rr *Renderer) { rr.restoreWorldSnapshot(after) },
	})
}

func (r *Renderer) restoreWorldSnapshot(snapshot editWorldSnapshot) {
	r.gs.Regions = cloneRegionMap(snapshot.Regions)
	r.gs.RegionOrder = cloneRegionIDSlice(snapshot.RegionOrder)
	r.gs.LandPassages = cloneLandPassages(snapshot.LandPassages)
	r.gs.TerrainAreas = cloneTerrainAreas(snapshot.TerrainAreas)
	r.gs.Factions = cloneFactionMap(snapshot.Factions)
	r.gs.AIStrategies = cloneAIStrategyMap(snapshot.AIStrategies)
	r.gs.TradeCenters = cloneTradeCenterConfig(snapshot.TradeCenters)
	r.gs.Armies = cloneArmyMap(snapshot.Armies)
	r.gs.Relations = cloneRelationMap(snapshot.Relations)
	r.gs.ShapeData = cloneCountryShapeJSON(snapshot.ShapeData)
	r.editRegionPaintOverrides = cloneRegionPaintOverrides(snapshot.RegionPaintOverrides)
	// Region paint overrides'ı oyun durumuna da senkronize et
	if len(r.editRegionPaintOverrides) > 0 {
		if r.gs.RegionPaintOverrides == nil {
			r.gs.RegionPaintOverrides = make(map[int]world.RegionID)
		}
		for k, v := range r.editRegionPaintOverrides {
			r.gs.RegionPaintOverrides[k] = v
		}
	} else {
		r.gs.RegionPaintOverrides = nil
	}
	r.editSelectedRegion = snapshot.Selected
	r.editSelectedSettlement = snapshot.Settlement
	r.editSelectedFaction = snapshot.Faction
	r.SelectedArmy = snapshot.Army
	r.gs.PlayerFactionID = snapshot.Player
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editLandPassageFrom = ""
	r.editLandPassageMode = false
	r.editLandPassageAdjustMode = false
	r.editLandPassageStart = [2]int{}
	r.editLandPassageStartSet = false
	r.editLandPassageSelected = -1
	r.editLandPassageDragEndpoint = -1
	r.editLandPassageDragBefore = nil
	r.editLandPassageDragChanged = false
	r.editLandPassageMessage = ""
	r.editNeighborAddMode = false
	r.editNeighborAddFrom = ""
	r.editNeighborAddMessage = ""
	r.editShapePainting = false
	r.editShapePaintPending = false
	r.editShapeStrokeBefore = nil
	r.editShapePendingBefore = nil
	r.editShapePendingAffectsLandShapes = false
	r.editRenaming = false
	r.rebuildEditWorldMap()
}

func cloneRegionMap(src map[world.RegionID]*world.Region) map[world.RegionID]*world.Region {
	dst := make(map[world.RegionID]*world.Region, len(src))
	for rid, region := range src {
		if region == nil {
			continue
		}
		copyRegion := *region
		copyRegion.Neighbors = cloneRegionIDSlice(region.Neighbors)
		copyRegion.Settlements = cloneSettlements(region.Settlements)
		copyRegion.Buildings = cloneStringSlice(region.Buildings)
		if len(region.Shape) > 0 {
			copyRegion.Shape = make([][][2]float32, len(region.Shape))
			for i := range region.Shape {
				copyRegion.Shape[i] = make([][2]float32, len(region.Shape[i]))
				copy(copyRegion.Shape[i], region.Shape[i])
			}
		}
		dst[rid] = &copyRegion
	}
	return dst
}

func cloneTerrainAreas(src []world.TerrainArea) []world.TerrainArea {
	if len(src) == 0 {
		return nil
	}
	dst := make([]world.TerrainArea, len(src))
	for i, area := range src {
		dst[i] = area
		dst[i].Cells = append([][2]int(nil), area.Cells...)
	}
	return dst
}

func cloneArmyMap(src map[army.ArmyID]*army.Army) map[army.ArmyID]*army.Army {
	dst := make(map[army.ArmyID]*army.Army, len(src))
	for aid, a := range src {
		if a == nil {
			continue
		}
		copyArmy := *a
		copyArmy.Units = make([]army.Unit, len(a.Units))
		copy(copyArmy.Units, a.Units)
		copyArmy.EmbarkedUnits = make([]army.Unit, len(a.EmbarkedUnits))
		copy(copyArmy.EmbarkedUnits, a.EmbarkedUnits)
		if a.Commander != nil {
			commander := *a.Commander
			commander.Traits = append([]army.CommanderTrait(nil), a.Commander.Traits...)
			copyArmy.Commander = &commander
		}
		if a.EmbarkedCommander != nil {
			commander := *a.EmbarkedCommander
			commander.Traits = append([]army.CommanderTrait(nil), a.EmbarkedCommander.Traits...)
			copyArmy.EmbarkedCommander = &commander
		}
		dst[aid] = &copyArmy
	}
	return dst
}

func cloneFactionMap(src map[faction.FactionID]*faction.Faction) map[faction.FactionID]*faction.Faction {
	dst := make(map[faction.FactionID]*faction.Faction, len(src))
	for fid, f := range src {
		if f == nil {
			continue
		}
		copyFaction := *f
		if f.Research.Completed != nil {
			copyFaction.Research.Completed = make(map[string]bool, len(f.Research.Completed))
			for id, done := range f.Research.Completed {
				copyFaction.Research.Completed[id] = done
			}
		}
		dst[fid] = &copyFaction
	}
	return dst
}

func cloneRelationMap(src map[string]*faction.Relation) map[string]*faction.Relation {
	dst := make(map[string]*faction.Relation, len(src))
	for key, rel := range src {
		if rel == nil {
			continue
		}
		copyRel := *rel
		dst[key] = &copyRel
	}
	return dst
}

func cloneAIStrategyMap(src map[string]scenario.AIFactionStrategy) map[string]scenario.AIFactionStrategy {
	if src == nil {
		return nil
	}
	dst := make(map[string]scenario.AIFactionStrategy, len(src))
	for id, strategy := range src {
		copyStrategy := strategy
		copyStrategy.TerritorialClaims = append([]scenario.AITerritorialClaimDef(nil), strategy.TerritorialClaims...)
		copyStrategy.Objectives = make([]scenario.AIObjectiveDef, len(strategy.Objectives))
		for i, objective := range strategy.Objectives {
			copyObjective := objective
			copyObjective.TargetFactions = cloneStringSlice(objective.TargetFactions)
			copyObjective.TargetRegions = cloneStringSlice(objective.TargetRegions)
			copyObjective.TerritorialClaims = append([]scenario.AITerritorialClaimDef(nil), objective.TerritorialClaims...)
			copyObjective.ReadinessRegions = cloneStringSlice(objective.ReadinessRegions)
			copyObjective.RequiredEventFlags = cloneStringSlice(objective.RequiredEventFlags)
			copyStrategy.Objectives[i] = copyObjective
		}
		dst[id] = copyStrategy
	}
	return dst
}

func cloneTradeCenterConfig(src world.TradeCenterConfig) world.TradeCenterConfig {
	if src.Centers == nil {
		return world.TradeCenterConfig{}
	}
	dst := world.TradeCenterConfig{
		PrimaryTradeCapacityBonus:   src.PrimaryTradeCapacityBonus,
		SecondaryTradeCapacityBonus: src.SecondaryTradeCapacityBonus,
		PrimaryTradeIncomeBonus:     src.PrimaryTradeIncomeBonus,
		SecondaryTradeIncomeBonus:   src.SecondaryTradeIncomeBonus,
		Centers:                     make([]world.TradeCenterDef, len(src.Centers)),
	}
	for i, center := range src.Centers {
		dst.Centers[i] = center
		dst.Centers[i].Links = cloneRegionIDSlice(center.Links)
	}
	return dst
}

func replaceRegionIDInStrings(ids []string, oldID, newID world.RegionID) {
	for i := range ids {
		if world.RegionID(ids[i]) == oldID {
			ids[i] = string(newID)
		}
	}
}

func cloneRegionIDSlice(src []world.RegionID) []world.RegionID {
	if src == nil {
		return nil
	}
	dst := make([]world.RegionID, len(src))
	copy(dst, src)
	return dst
}

func cloneStringSlice(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func (r *Renderer) insertRegionOrderAfter(after, rid world.RegionID) {
	r.removeRegionFromOrder(rid)
	if len(r.gs.RegionOrder) == 0 {
		r.gs.RegionOrder = append(r.gs.RegionOrder, rid)
		return
	}
	for i, existing := range r.gs.RegionOrder {
		if existing == after {
			r.gs.RegionOrder = append(r.gs.RegionOrder, "")
			copy(r.gs.RegionOrder[i+2:], r.gs.RegionOrder[i+1:])
			r.gs.RegionOrder[i+1] = rid
			return
		}
	}
	r.gs.RegionOrder = append(r.gs.RegionOrder, rid)
}

func (r *Renderer) removeRegionFromOrder(rid world.RegionID) {
	out := r.gs.RegionOrder[:0]
	for _, existing := range r.gs.RegionOrder {
		if existing != rid {
			out = append(out, existing)
		}
	}
	r.gs.RegionOrder = out
}

type editRegionNeighborsSnapshot struct {
	Region    world.RegionID
	Neighbors []world.RegionID
}

func (r *Renderer) neighborSnapshot(rid world.RegionID, affected []world.RegionID) []editRegionNeighborsSnapshot {
	snaps := make([]editRegionNeighborsSnapshot, 0, len(affected)+1)
	snaps = append(snaps, r.singleNeighborSnapshot(rid))
	for _, nrid := range affected {
		if nrid != rid {
			snaps = append(snaps, r.singleNeighborSnapshot(nrid))
		}
	}
	return uniqueNeighborSnapshots(snaps)
}

func (r *Renderer) singleNeighborSnapshot(rid world.RegionID) editRegionNeighborsSnapshot {
	region := r.gs.Regions[rid]
	if region == nil {
		return editRegionNeighborsSnapshot{Region: rid}
	}
	neighbors := make([]world.RegionID, len(region.Neighbors))
	copy(neighbors, region.Neighbors)
	return editRegionNeighborsSnapshot{Region: rid, Neighbors: neighbors}
}

func (r *Renderer) applyVisualNeighbors(rid world.RegionID, visual []world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}
	oldNeighbors := region.Neighbors
	region.Neighbors = sortedRegionIDs(visual)
	for _, oldID := range oldNeighbors {
		if !regionIDContains(visual, oldID) {
			removeNeighborID(r.gs.Regions[oldID], rid)
		}
	}
	for _, nrid := range visual {
		addNeighborID(r.gs.Regions[nrid], rid)
	}
}

func (r *Renderer) restoreNeighborSnapshots(snaps []editRegionNeighborsSnapshot) {
	for _, snap := range snaps {
		region := r.gs.Regions[snap.Region]
		if region == nil {
			continue
		}
		region.Neighbors = make([]world.RegionID, len(snap.Neighbors))
		copy(region.Neighbors, snap.Neighbors)
	}
	r.editDraggingSettlement = false
	r.editDraggingRegion = false
	r.editRenaming = false
}

func uniqueNeighborSnapshots(snaps []editRegionNeighborsSnapshot) []editRegionNeighborsSnapshot {
	out := snaps[:0]
	for _, snap := range snaps {
		seen := false
		for _, existing := range out {
			if existing.Region == snap.Region {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, snap)
		}
	}
	return out
}

func cloneNeighborSnapshots(snaps []editRegionNeighborsSnapshot) []editRegionNeighborsSnapshot {
	out := make([]editRegionNeighborsSnapshot, len(snaps))
	for i, snap := range snaps {
		out[i].Region = snap.Region
		out[i].Neighbors = make([]world.RegionID, len(snap.Neighbors))
		copy(out[i].Neighbors, snap.Neighbors)
	}
	return out
}

func neighborSnapshotsEqual(a, b []editRegionNeighborsSnapshot) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Region != b[i].Region || !regionIDSlicesEqual(a[i].Neighbors, b[i].Neighbors) {
			return false
		}
	}
	return true
}

func sortedRegionIDs(ids []world.RegionID) []world.RegionID {
	out := make([]world.RegionID, 0, len(ids))
	for _, rid := range ids {
		if rid != "" && !regionIDContains(out, rid) {
			out = append(out, rid)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func addNeighborID(region *world.Region, rid world.RegionID) {
	if region == nil || rid == "" || regionHasNeighbor(region, rid) {
		return
	}
	region.Neighbors = append(region.Neighbors, rid)
	sort.Slice(region.Neighbors, func(i, j int) bool { return region.Neighbors[i] < region.Neighbors[j] })
}

func removeNeighborID(region *world.Region, rid world.RegionID) {
	if region == nil {
		return
	}
	out := region.Neighbors[:0]
	for _, nrid := range region.Neighbors {
		if nrid != rid {
			out = append(out, nrid)
		}
	}
	region.Neighbors = out
}

func regionIDContains(ids []world.RegionID, rid world.RegionID) bool {
	for _, id := range ids {
		if id == rid {
			return true
		}
	}
	return false
}

func regionIDSlicesEqual(a, b []world.RegionID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (r *Renderer) selectedEditFaction() *faction.Faction {
	if r.editSelectedFaction != "" {
		if f := r.gs.Factions[r.editSelectedFaction]; f != nil {
			return f
		}
	}
	if region := r.gs.Regions[r.editSelectedRegion]; region != nil && region.OwnerID != "" {
		return r.gs.Factions[faction.FactionID(region.OwnerID)]
	}
	if r.SelectedArmy != "" {
		if a := r.gs.Armies[r.SelectedArmy]; a != nil && a.OwnerID != "" {
			return r.gs.Factions[faction.FactionID(a.OwnerID)]
		}
	}
	return nil
}

func (r *Renderer) setEditFactionFromRegion(rid world.RegionID) {
	region := r.gs.Regions[rid]
	if region == nil {
		r.editSelectedFaction = ""
		return
	}
	r.editSelectedFaction = faction.FactionID(region.OwnerID)
}

func (r *Renderer) setEditFactionFromArmy(a *army.Army) {
	if a == nil {
		r.editSelectedFaction = ""
		return
	}
	r.editSelectedFaction = faction.FactionID(a.OwnerID)
}

func (r *Renderer) deleteSelectedFaction() {
	f := r.selectedEditFaction()
	if f == nil {
		return
	}
	before := r.worldSnapshot()
	fid := f.ID
	delete(r.gs.Factions, fid)
	for _, region := range r.gs.Regions {
		if region != nil && region.OwnerID == string(fid) {
			region.OwnerID = ""
		}
	}
	for aid, a := range r.gs.Armies {
		if a != nil && a.OwnerID == string(fid) {
			r.gs.RemoveArmy(aid)
		}
	}
	if r.gs.PlayerFactionID == fid {
		r.gs.PlayerFactionID = ""
	}
	for key, rel := range r.gs.Relations {
		if rel != nil && (rel.FactionA == fid || rel.FactionB == fid) {
			delete(r.gs.Relations, key)
		}
	}
	r.editSelectedFaction = ""
	r.SelectedArmy = ""
	r.worldMap.MarkDirty()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) openFactionCreateForm() {
	fid := nextFactionID(r.gs)
	form := editFactionFormState{
		show:     true,
		create:   true,
		active:   editFactionFieldNameTR,
		id:       string(fid),
		name:     "New Faction",
		nameTR:   "",
		religion: religion.Catholic,
		color:    editFactionColor(len(r.gs.Factions) + 1),
		playable: true,
		gold:     "500",
		grain:    "100",
		iron:     "50",
		timber:   "50",
		spice:    "0",
		cloth:    "0",
		ai:       "50",
	}
	if f := r.selectedEditFaction(); f != nil {
		form.religion = f.Religion
	}
	r.editFactionForm = form
	r.setFactionFormRelationTarget(firstRelationTarget(r.gs, fid))
}

func (r *Renderer) openFactionEditForm() {
	f := r.selectedEditFaction()
	if f == nil {
		return
	}
	r.editFactionForm = editFactionFormState{
		show:       true,
		create:     false,
		active:     editFactionFieldNameTR,
		originalID: f.ID,
		id:         string(f.ID),
		name:       f.Name,
		nameTR:     f.NameTR,
		religion:   f.Religion,
		color:      f.Color,
		playable:   f.IsPlayable,
		gold:       itoa(f.Gold),
		grain:      itoa(f.Grain),
		iron:       itoa(f.Iron),
		timber:     itoa(f.Timber),
		spice:      itoa(f.Spice),
		cloth:      itoa(f.Cloth),
		ai:         itoa(f.AIAggressiveness),
	}
	r.setFactionFormRelationTarget(firstRelationTarget(r.gs, f.ID))
}

func (r *Renderer) saveFactionForm() bool {
	form := &r.editFactionForm
	fid := faction.FactionID(strings.TrimSpace(form.id))
	if fid == "" {
		form.errorText = "ID bos olamaz."
		return false
	}
	if existing := r.gs.Factions[fid]; existing != nil && (form.create || fid != form.originalID) {
		form.errorText = "Bu faction ID zaten var."
		return false
	}
	name := strings.TrimSpace(form.name)
	nameTR := strings.TrimSpace(form.nameTR)
	if name == "" && nameTR == "" {
		form.errorText = "En az bir isim gir."
		return false
	}
	gold, ok := parseEditInt(form.gold, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceGold)
		return false
	}
	grain, ok := parseEditInt(form.grain, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceGrain)
		return false
	}
	iron, ok := parseEditInt(form.iron, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceIron)
		return false
	}
	timber, ok := parseEditInt(form.timber, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceTimber)
		return false
	}
	spice, ok := parseEditInt(form.spice, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceSpice)
		return false
	}
	cloth, ok := parseEditInt(form.cloth, 0, 999999)
	if !ok {
		form.errorText = economy.ResourceInvalidCountMessageTR(economy.ResourceCloth)
		return false
	}
	aiValue, ok := parseEditInt(form.ai, 0, 100)
	if !ok {
		form.errorText = "AI 0-100 araliginda olmali."
		return false
	}
	relationScore, ok := parseEditInt(form.relationScore, -100, 100)
	if !ok && form.relationTarget != "" {
		form.errorText = "Iliski skoru -100 ile 100 arasinda olmali."
		return false
	}

	before := r.worldSnapshot()
	var existingFaction *faction.Faction
	if !form.create && form.originalID != "" {
		existingFaction = r.gs.Factions[form.originalID]
	}
	if !form.create && form.originalID != "" && form.originalID != fid {
		delete(r.gs.Factions, form.originalID)
		r.renameFactionRelations(form.originalID, fid)
		for _, region := range r.gs.Regions {
			if region != nil && region.OwnerID == string(form.originalID) {
				region.OwnerID = string(fid)
			}
		}
		for _, a := range r.gs.Armies {
			if a != nil && a.OwnerID == string(form.originalID) {
				a.OwnerID = string(fid)
			}
		}
		if r.gs.PlayerFactionID == form.originalID {
			r.gs.PlayerFactionID = fid
		}
	}
	next := &faction.Faction{
		ID:               fid,
		Name:             name,
		NameTR:           nameTR,
		Religion:         form.religion,
		Color:            form.color,
		IsPlayable:       form.playable,
		Gold:             gold,
		Grain:            grain,
		Iron:             iron,
		Timber:           timber,
		Spice:            spice,
		Cloth:            cloth,
		AIAggressiveness: aiValue,
	}
	if existingFaction != nil {
		next.IsEliminated = existingFaction.IsEliminated
		next.Research = existingFaction.Research
	}
	r.gs.Factions[fid] = next
	r.ensureRelationsForFaction(fid)
	if form.relationTarget != "" && r.gs.Factions[form.relationTarget] != nil && form.relationTarget != fid {
		r.setRelationValue(fid, form.relationTarget, relationScore, form.relationStance)
	}
	r.editSelectedFaction = fid
	r.worldMap.MarkDirty()
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editFactionForm = editFactionFormState{}
	r.editDirty = true
	return true
}

func (r *Renderer) handleEditFactionFormInput() InputAction {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editFactionForm = editFactionFormState{}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		if r.saveFactionForm() {
			return InputAction{Kind: ActionSaveScenario}
		}
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) {
		r.editFactionFormBackspace()
	}
	if r.mouseJustPressed(ebiten.MouseButtonLeft) {
		if r.handleFactionFormClick(fx, fy) {
			return InputAction{Kind: ActionSaveScenario}
		}
	}
	if r.editFactionForm.active != editFactionFieldNone {
		for _, ch := range ebiten.AppendInputChars(nil) {
			r.appendFactionFormRune(ch)
		}
	}
	return InputAction{}
}

func (r *Renderer) handleFactionFormClick(fx, fy float64) bool {
	if !editFactionFormHit(fx, fy) {
		return false
	}
	for field := editFactionFieldID; field <= editFactionFieldAI; field++ {
		if buildEditFactionFieldButton(field, "").HitTest(fx, fy) {
			r.editFactionForm.active = field
			r.editFactionForm.errorText = ""
			return false
		}
	}
	switch {
	case buildEditFactionFormButton(editFactionFormSave, "Kaydet").HitTest(fx, fy):
		return r.saveFactionForm()
	case buildEditFactionFormButton(editFactionFormCancel, "Iptal").HitTest(fx, fy):
		r.editFactionForm = editFactionFormState{}
	case buildEditFactionFormButton(editFactionFormReligion, "").HitTest(fx, fy):
		r.editFactionForm.religion = nextEditReligion(r.editFactionForm.religion)
	case buildEditFactionFormButton(editFactionFormPlayable, "").HitTest(fx, fy):
		r.editFactionForm.playable = !r.editFactionForm.playable
	case buildEditFactionFormButton(editFactionFormRelationTarget, "").HitTest(fx, fy):
		r.cycleFactionFormRelationTarget()
	case buildEditFactionFormButton(editFactionFormRelationStance, "").HitTest(fx, fy):
		r.editFactionForm.relationStance = nextEditStance(r.editFactionForm.relationStance)
	case buildEditFactionFormButton(editFactionFormRelationScoreMinus, "").HitTest(fx, fy):
		r.adjustFactionFormRelationScore(-10)
	case buildEditFactionFormButton(editFactionFormRelationScorePlus, "").HitTest(fx, fy):
		r.adjustFactionFormRelationScore(10)
	case buildEditFactionFormButton(editFactionFormRedMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(0, -10)
	case buildEditFactionFormButton(editFactionFormRedPlus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(0, 10)
	case buildEditFactionFormButton(editFactionFormGreenMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(1, -10)
	case buildEditFactionFormButton(editFactionFormGreenPlus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(1, 10)
	case buildEditFactionFormButton(editFactionFormBlueMinus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(2, -10)
	case buildEditFactionFormButton(editFactionFormBluePlus, "").HitTest(fx, fy):
		r.adjustFactionFormColor(2, 10)
	}
	return false
}

func (r *Renderer) editFactionFormBackspace() {
	switch r.editFactionForm.active {
	case editFactionFieldID:
		r.editFactionForm.id = trimLastRune(r.editFactionForm.id)
	case editFactionFieldName:
		r.editFactionForm.name = trimLastRune(r.editFactionForm.name)
	case editFactionFieldNameTR:
		r.editFactionForm.nameTR = trimLastRune(r.editFactionForm.nameTR)
	case editFactionFieldGold:
		r.editFactionForm.gold = trimLastRune(r.editFactionForm.gold)
	case editFactionFieldGrain:
		r.editFactionForm.grain = trimLastRune(r.editFactionForm.grain)
	case editFactionFieldIron:
		r.editFactionForm.iron = trimLastRune(r.editFactionForm.iron)
	case editFactionFieldTimber:
		r.editFactionForm.timber = trimLastRune(r.editFactionForm.timber)
	case editFactionFieldSpice:
		r.editFactionForm.spice = trimLastRune(r.editFactionForm.spice)
	case editFactionFieldCloth:
		r.editFactionForm.cloth = trimLastRune(r.editFactionForm.cloth)
	case editFactionFieldAI:
		r.editFactionForm.ai = trimLastRune(r.editFactionForm.ai)
	}
}

func (r *Renderer) appendFactionFormRune(ch rune) {
	if r.editFactionForm.active >= editFactionFieldGold && r.editFactionForm.active <= editFactionFieldAI {
		if ch < '0' || ch > '9' {
			return
		}
	}
	switch r.editFactionForm.active {
	case editFactionFieldID:
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			r.editFactionForm.id = limitStringRunes(r.editFactionForm.id+string(ch), 40)
		}
	case editFactionFieldName:
		r.editFactionForm.name = limitStringRunes(r.editFactionForm.name+string(ch), 64)
	case editFactionFieldNameTR:
		r.editFactionForm.nameTR = limitStringRunes(r.editFactionForm.nameTR+string(ch), 64)
	case editFactionFieldGold:
		r.editFactionForm.gold = limitStringRunes(r.editFactionForm.gold+string(ch), 8)
	case editFactionFieldGrain:
		r.editFactionForm.grain = limitStringRunes(r.editFactionForm.grain+string(ch), 8)
	case editFactionFieldIron:
		r.editFactionForm.iron = limitStringRunes(r.editFactionForm.iron+string(ch), 8)
	case editFactionFieldTimber:
		r.editFactionForm.timber = limitStringRunes(r.editFactionForm.timber+string(ch), 8)
	case editFactionFieldSpice:
		r.editFactionForm.spice = limitStringRunes(r.editFactionForm.spice+string(ch), 8)
	case editFactionFieldCloth:
		r.editFactionForm.cloth = limitStringRunes(r.editFactionForm.cloth+string(ch), 8)
	case editFactionFieldAI:
		r.editFactionForm.ai = limitStringRunes(r.editFactionForm.ai+string(ch), 3)
	}
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	return string(runes[:len(runes)-1])
}

func limitStringRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func (r *Renderer) adjustFactionFormColor(index int, delta int) {
	value := int(r.editFactionForm.color[index]) + delta
	if value < 0 {
		value = 0
	}
	if value > 255 {
		value = 255
	}
	r.editFactionForm.color[index] = uint8(value)
}

func nextEditReligion(current religion.Type) religion.Type {
	return religion.Next(current)
}

func nextEditStance(current faction.DiplomaticStance) faction.DiplomaticStance {
	return faction.NextDiplomaticStance(current)
}

func nextFactionID(gs *state.GameState) faction.FactionID {
	for n := len(gs.Factions) + 1; ; n++ {
		fid := faction.FactionID("new_faction_" + itoa(n))
		if _, used := gs.Factions[fid]; !used {
			return fid
		}
	}
}

func editFactionColor(seed int) [3]uint8 {
	return [3]uint8{
		uint8(70 + (seed*53)%160),
		uint8(70 + (seed*97)%150),
		uint8(70 + (seed*139)%150),
	}
}

func parseEditInt(value string, minValue, maxValue int) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < minValue || n > maxValue {
		return 0, false
	}
	return n, true
}

func firstRelationTarget(gs *state.GameState, self faction.FactionID) faction.FactionID {
	for _, fid := range sortedFactionIDs(gs.Factions) {
		if fid != self {
			return fid
		}
	}
	return ""
}

func sortedFactionIDs(factions map[faction.FactionID]*faction.Faction) []faction.FactionID {
	ids := make([]faction.FactionID, 0, len(factions))
	for fid := range factions {
		ids = append(ids, fid)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (r *Renderer) setFactionFormRelationTarget(target faction.FactionID) {
	r.editFactionForm.relationTarget = target
	if target == "" {
		r.editFactionForm.relationScore = "0"
		r.editFactionForm.relationStance = faction.StancePeace
		return
	}
	self := faction.FactionID(strings.TrimSpace(r.editFactionForm.id))
	if self == "" {
		self = r.editFactionForm.originalID
	}
	if rel := r.relationForForm(self, target); rel != nil {
		r.editFactionForm.relationScore = itoa(rel.Score)
		r.editFactionForm.relationStance = rel.Stance
		return
	}
	targetFaction := r.gs.Factions[target]
	score := 0
	stance := faction.StancePeace
	if targetFaction != nil {
		score = religion.Relation(r.editFactionForm.religion, targetFaction.Religion)
		if (r.editFactionForm.religion == religion.Sunni && targetFaction.Religion == religion.Shia) ||
			(r.editFactionForm.religion == religion.Shia && targetFaction.Religion == religion.Sunni) {
			stance = faction.StanceWar
		}
	}
	r.editFactionForm.relationScore = itoa(score)
	r.editFactionForm.relationStance = stance
}

func (r *Renderer) relationForForm(self, target faction.FactionID) *faction.Relation {
	if self == "" || target == "" || self == target {
		return nil
	}
	return r.gs.Relations[faction.RelationKey(self, target)]
}

func (r *Renderer) cycleFactionFormRelationTarget() {
	self := faction.FactionID(strings.TrimSpace(r.editFactionForm.id))
	ids := sortedFactionIDs(r.gs.Factions)
	if len(ids) == 0 {
		r.setFactionFormRelationTarget("")
		return
	}
	current := r.editFactionForm.relationTarget
	start := 0
	for i, fid := range ids {
		if fid == current {
			start = i + 1
			break
		}
	}
	for offset := 0; offset < len(ids); offset++ {
		fid := ids[(start+offset)%len(ids)]
		if fid != self {
			r.setFactionFormRelationTarget(fid)
			return
		}
	}
	r.setFactionFormRelationTarget("")
}

func (r *Renderer) adjustFactionFormRelationScore(delta int) {
	score, ok := parseEditInt(r.editFactionForm.relationScore, -100, 100)
	if !ok {
		score = 0
	}
	score += delta
	if score < -100 {
		score = -100
	}
	if score > 100 {
		score = 100
	}
	r.editFactionForm.relationScore = itoa(score)
}

func (r *Renderer) ensureRelationsForFaction(fid faction.FactionID) {
	if r.gs.Relations == nil {
		r.gs.Relations = make(map[string]*faction.Relation)
	}
	self := r.gs.Factions[fid]
	for otherID, other := range r.gs.Factions {
		if otherID == fid || other == nil {
			continue
		}
		key := faction.RelationKey(fid, otherID)
		if r.gs.Relations[key] != nil {
			continue
		}
		score := 0
		stance := faction.StancePeace
		if self != nil {
			score = religion.Relation(self.Religion, other.Religion)
			if (self.Religion == religion.Sunni && other.Religion == religion.Shia) ||
				(self.Religion == religion.Shia && other.Religion == religion.Sunni) {
				stance = faction.StanceWar
			}
		}
		r.gs.Relations[key] = &faction.Relation{FactionA: fid, FactionB: otherID, Score: score, Stance: stance}
	}
}

func (r *Renderer) setRelationValue(a, b faction.FactionID, score int, stance faction.DiplomaticStance) {
	key := faction.RelationKey(a, b)
	r.gs.Relations[key] = &faction.Relation{FactionA: a, FactionB: b, Score: score, Stance: stance}
}

func (r *Renderer) renameFactionRelations(oldID, newID faction.FactionID) {
	next := make(map[string]*faction.Relation, len(r.gs.Relations))
	for _, rel := range r.gs.Relations {
		if rel == nil {
			continue
		}
		copyRel := *rel
		if copyRel.FactionA == oldID {
			copyRel.FactionA = newID
		}
		if copyRel.FactionB == oldID {
			copyRel.FactionB = newID
		}
		if copyRel.FactionA == copyRel.FactionB {
			continue
		}
		next[faction.RelationKey(copyRel.FactionA, copyRel.FactionB)] = &copyRel
	}
	r.gs.Relations = next
}

func (r *Renderer) moveSelectedArmyToEditRegion() {
	a := r.gs.Armies[r.SelectedArmy]
	region := r.gs.Regions[r.editSelectedRegion]
	if a == nil || region == nil || a.RegionID == region.ID {
		return
	}
	if (a.IsNaval && !region.IsSea) || (!a.IsNaval && region.IsSea) {
		return
	}
	aid := a.ID
	old := a.RegionID
	oldDockedRegion := a.DockedRegionID
	oldDockedSettlement := a.DockedSettlementID
	next := region.ID
	a.RegionID = next
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setArmyLocation(aid, old, oldDockedRegion, oldDockedSettlement) },
		redo: func(rr *Renderer) { rr.setArmyLocation(aid, next, "", "") },
	})
	r.editDirty = true
}

func (r *Renderer) addEditLandArmy() {
	region := r.gs.Regions[r.editSelectedRegion]
	if !r.canAddEditLandArmy(region) {
		return
	}
	ownerID := r.editOwnerForRegion(region)
	unitTypeID := r.defaultEditUnitType(false)
	if ownerID == "" || unitTypeID == "" {
		return
	}
	before := r.worldSnapshot()
	aid := nextEditArmyID(r.gs)
	r.gs.Armies[aid] = &army.Army{
		ID:            aid,
		OwnerID:       ownerID,
		RegionID:      region.ID,
		Units:         army.MakeUnits(unitTypeID, 1),
		MovePoints:    2,
		MaxMovePoints: 2,
		IsNaval:       false,
	}
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) addEditFleet() {
	region := r.gs.Regions[r.editSelectedRegion]
	if !r.canAddEditFleet(region) {
		return
	}
	ownerID := r.editOwnerForRegion(region)
	seaID := r.editFleetSeaRegion(region)
	unitTypeID := r.defaultEditUnitType(true)
	if ownerID == "" || seaID == "" || unitTypeID == "" {
		return
	}
	before := r.worldSnapshot()
	aid := nextEditArmyID(r.gs)
	r.gs.Armies[aid] = &army.Army{
		ID:                 aid,
		OwnerID:            ownerID,
		RegionID:           seaID,
		DockedRegionID:     region.ID,
		DockedSettlementID: r.editPreferredDockSettlementID(region),
		Units:              army.MakeUnits(unitTypeID, 1),
		MovePoints:         2,
		MaxMovePoints:      2,
		IsNaval:            true,
	}
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) deleteSelectedArmy() {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		return
	}
	before := r.worldSnapshot()
	r.gs.RemoveArmy(a.ID)
	r.SelectedArmy = ""
	r.editSelectedUnitType = ""
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) addSelectedArmyUnit() {
	a := r.gs.Armies[r.SelectedArmy]
	if !r.canAddSelectedArmyUnit() || a == nil {
		return
	}
	before := r.worldSnapshot()
	a.Units = append(a.Units, army.Unit{TypeID: r.editSelectedUnitType, CurrentHP: 100})
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

// setSelectedEditArmyUnitType, Edit Mode'da dropdown ile seçilen tipi seçili
// ordu/filonun mevcut birimlerine uygular. Böylece seçim yalnızca sonraki
// "Birim +" işlemini değil, kaydedilecek gerçek Unit.TypeID değerlerini de
// değiştirir.
func (r *Renderer) setSelectedEditArmyUnitType(typeID string) {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil || !r.unitTypeMatchesArmy(a, typeID) {
		return
	}
	if r.editSelectedUnitType == typeID && allArmyUnitsHaveType(a, typeID) {
		return
	}
	before := r.worldSnapshot()
	for i := range a.Units {
		a.Units[i].TypeID = typeID
	}
	r.editSelectedUnitType = typeID
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func allArmyUnitsHaveType(a *army.Army, typeID string) bool {
	if a == nil {
		return false
	}
	for _, unit := range a.Units {
		if unit.TypeID != typeID {
			return false
		}
	}
	return true
}

func (r *Renderer) removeSelectedArmyUnit() {
	a := r.gs.Armies[r.SelectedArmy]
	if !r.canRemoveSelectedArmyUnit() || a == nil {
		return
	}
	before := r.worldSnapshot()
	for i := len(a.Units) - 1; i >= 0; i-- {
		if a.Units[i].TypeID == r.editSelectedUnitType {
			a.Units = append(a.Units[:i], a.Units[i+1:]...)
			break
		}
	}
	after := r.worldSnapshot()
	r.pushWorldSnapshotCommand(before, after)
	r.editDirty = true
}

func (r *Renderer) toggleEditUnitTypeDropdown() {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		r.editUnitTypeDropdown.Close()
		return
	}
	r.ensureEditSelectedUnitType(a)
	dx, dy, _, _ := editOwnerDropdownRect()
	r.editUnitTypeDropdown.SetPosition(float64(dx), float64(dy))
	r.editUnitTypeDropdown.SetOptions(r.editUnitTypeOptions(a.IsNaval), r.editSelectedUnitType)
	r.editUnitTypeDropdown.Toggle()
}

func (r *Renderer) canAddEditLandArmy(region *world.Region) bool {
	return region != nil && !region.IsSea && !region.IsLocked && r.editOwnerForRegion(region) != "" && r.defaultEditUnitType(false) != ""
}

func (r *Renderer) canAddEditFleet(region *world.Region) bool {
	return region != nil && !region.IsSea && r.editOwnerForRegion(region) != "" &&
		region.HasPort() &&
		r.editFleetSeaRegion(region) != "" && r.defaultEditUnitType(true) != ""
}

func (r *Renderer) canAddSelectedArmyUnit() bool {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil || len(a.Units) >= army.MaxArmySize {
		return false
	}
	r.ensureEditSelectedUnitType(a)
	return r.editSelectedUnitType != "" && r.unitTypeMatchesArmy(a, r.editSelectedUnitType)
}

func (r *Renderer) canRemoveSelectedArmyUnit() bool {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil || len(a.Units) == 0 {
		return false
	}
	r.ensureEditSelectedUnitType(a)
	for _, u := range a.Units {
		if u.TypeID == r.editSelectedUnitType {
			return true
		}
	}
	return false
}

func (r *Renderer) ensureEditSelectedUnitType(a *army.Army) {
	if a == nil {
		r.editSelectedUnitType = ""
		return
	}
	if r.editSelectedUnitType != "" && r.unitTypeMatchesArmy(a, r.editSelectedUnitType) {
		return
	}
	if len(a.Units) > 0 && r.unitTypeMatchesArmy(a, a.Units[0].TypeID) {
		r.editSelectedUnitType = a.Units[0].TypeID
		return
	}
	r.editSelectedUnitType = r.defaultEditUnitType(a.IsNaval)
}

func (r *Renderer) editUnitTypeOptions(isNaval bool) []string {
	options := make([]string, 0, len(r.gs.UnitTypes))
	for typeID := range r.gs.UnitTypes {
		if r.unitTypeIsNaval(typeID) == isNaval {
			options = append(options, typeID)
		}
	}
	sort.Strings(options)
	return options
}

func (r *Renderer) unitTypeMatchesArmy(a *army.Army, typeID string) bool {
	if a == nil || r.gs.UnitTypes[typeID] == nil {
		return false
	}
	return r.unitTypeIsNaval(typeID) == a.IsNaval
}

func (r *Renderer) unitTypeIsNaval(typeID string) bool {
	utype := r.gs.UnitTypes[typeID]
	return utype != nil && utype.RequiredBldg == "port"
}

func (r *Renderer) defaultEditUnitType(isNaval bool) string {
	preferred := "militia"
	if isNaval {
		preferred = "transport"
	}
	if r.gs.UnitTypes[preferred] != nil && r.unitTypeIsNaval(preferred) == isNaval {
		return preferred
	}
	options := r.editUnitTypeOptions(isNaval)
	if len(options) == 0 {
		return ""
	}
	return options[0]
}

func (r *Renderer) editFleetSeaRegion(region *world.Region) world.RegionID {
	if region == nil {
		return ""
	}
	for _, nid := range region.Neighbors {
		if n := r.gs.Regions[nid]; n != nil && n.IsSea {
			return n.ID
		}
	}
	visual := r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
	for _, nid := range visual {
		if n := r.gs.Regions[nid]; n != nil && n.IsSea {
			return n.ID
		}
	}
	return ""
}

func (r *Renderer) editOwnerForRegion(region *world.Region) string {
	if region != nil && region.OwnerID != "" {
		return region.OwnerID
	}
	if r.editSelectedFaction != "" {
		return string(r.editSelectedFaction)
	}
	return ""
}

func nextEditArmyID(gs *state.GameState) army.ArmyID {
	for i := len(gs.Armies) + 1; ; i++ {
		id := army.ArmyID("army_edit_" + itoa(i))
		if gs.Armies[id] == nil {
			return id
		}
	}
}

func (r *Renderer) setSelectedArmyOwnerFromRegion() {
	a := r.gs.Armies[r.SelectedArmy]
	region := r.selectedArmyOwnerRegion(a)
	if a == nil || region == nil || region.OwnerID == "" || a.OwnerID == region.OwnerID {
		return
	}
	aid := a.ID
	old := a.OwnerID
	next := region.OwnerID
	a.OwnerID = next
	r.pushEditCommand(editCommand{
		undo: func(rr *Renderer) { rr.setArmyOwner(aid, old) },
		redo: func(rr *Renderer) { rr.setArmyOwner(aid, next) },
	})
	r.editDirty = true
}

func (r *Renderer) canAssignSelectedArmyToRegionOwner() bool {
	a := r.gs.Armies[r.SelectedArmy]
	region := r.selectedArmyOwnerRegion(a)
	return a != nil && region != nil && region.OwnerID != "" && a.OwnerID != region.OwnerID
}

func (r *Renderer) selectedArmyOwnerRegion(a *army.Army) *world.Region {
	if a == nil || r.gs == nil {
		return nil
	}
	rid := r.editSelectedRegion
	if a.IsNaval && a.DockedRegionID != "" {
		rid = a.DockedRegionID
	} else if rid == "" {
		rid = a.RegionID
	}
	return r.gs.Regions[rid]
}

func (r *Renderer) setArmyLocation(aid army.ArmyID, rid, dockedRegionID world.RegionID, dockedSettlementID string) {
	if a := r.gs.Armies[aid]; a != nil {
		a.RegionID = rid
		a.DockedRegionID = dockedRegionID
		a.DockedSettlementID = dockedSettlementID
		r.SelectedArmy = aid
		r.editSelectedRegion = rid
		r.editSelectedSettlement = -1
	}
}

func (r *Renderer) editPreferredDockSettlementID(region *world.Region) string {
	if region == nil {
		return ""
	}
	if r.editSelectedSettlement >= 0 && r.editSelectedSettlement < len(region.Settlements) {
		settlement := region.Settlements[r.editSelectedSettlement]
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	for _, settlement := range region.Settlements {
		if settlement.Type == world.SettlementPort {
			return settlement.ID
		}
	}
	if len(region.Settlements) > 0 {
		return region.Settlements[0].ID
	}
	return ""
}

func (r *Renderer) setArmyOwner(aid army.ArmyID, ownerID string) {
	if a := r.gs.Armies[aid]; a != nil {
		a.OwnerID = ownerID
		r.SelectedArmy = aid
		r.editSelectedFaction = faction.FactionID(ownerID)
	}
}

func editBoolLabel(value bool) string {
	if value {
		return "evet"
	}
	return "hayir"
}

func (r *Renderer) rebuildEditWorldMap() {
	r.invalidateShapeEditSession()
	world.SyncTerrainAreaRegions(r.gs.Regions, r.gs.TerrainAreas)
	r.worldMap = NewWorldMap(r.gs)
	r.buildRegionPaintBaseline()
	if !regionPaintOverridesEqual(r.editRegionPaintOverrides, r.gs.RegionPaintOverrides) {
		r.applyRegionPaintOverrides()
	}
	// Editör oturumundaki bölge boya override'ları arazi alanı hücrelerini
	// ezmesin diye alanlar son katman olarak yeniden boyanır.
	r.worldMap.applyTerrainAreaRegions(r.gs)
}

func (r *Renderer) buildRegionPaintBaseline() {
	if r.worldMap == nil {
		r.editRegionPaintBaseline = nil
		return
	}
	if len(r.worldMap.baseRegionAt) == len(r.worldMap.regionAt) {
		r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.baseRegionAt))
		copy(r.editRegionPaintBaseline, r.worldMap.baseRegionAt)
		return
	}
	if len(r.editRegionPaintOverrides) == 0 && len(r.gs.RegionPaintOverrides) == 0 {
		r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.regionAt))
		copy(r.editRegionPaintBaseline, r.worldMap.regionAt)
		return
	}
	r.editRegionPaintBaseline = make([]uint16, len(r.worldMap.regionAt))
	copy(r.editRegionPaintBaseline, r.worldMap.regionAt)
}

func (r *Renderer) applyRegionPaintOverrides() {
	if r.worldMap == nil || len(r.editRegionPaintOverrides) == 0 {
		return
	}
	for pIdx, rid := range r.editRegionPaintOverrides {
		r.applyRegionOverride(pIdx, rid)
	}
}

func cloneRegionPaintOverrides(src map[int]world.RegionID) map[int]world.RegionID {
	if src == nil {
		return nil
	}
	dst := make(map[int]world.RegionID, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func regionPaintOverridesEqual(a, b map[int]world.RegionID) bool {
	if len(a) != len(b) {
		return false
	}
	for pIdx, rid := range a {
		if b[pIdx] != rid {
			return false
		}
	}
	return true
}

func (r *Renderer) applyRegionOverride(pIdx int, rid world.RegionID) {
	if r.worldMap == nil || pIdx < 0 || pIdx >= len(r.worldMap.regionAt) {
		return
	}
	if rid == "" {
		return
	}
	newIdx, ok := r.worldMap.regionIdx[rid]
	if !ok {
		newIdx = uint16(len(r.worldMap.regionIDs))
		r.worldMap.regionIDs = append(r.worldMap.regionIDs, rid)
		r.worldMap.regionIdx[rid] = newIdx
	}
	oldIdx := r.worldMap.regionAt[pIdx]
	if oldIdx == newIdx {
		return
	}
	if oldIdx != 0 {
		oldID := r.worldMap.regionIDs[oldIdx]
		r.worldMap.regionPx[oldID] = removePixelIndex(r.worldMap.regionPx[oldID], pIdx)
	}
	r.worldMap.regionAt[pIdx] = newIdx
	r.worldMap.regionPx[rid] = append(r.worldMap.regionPx[rid], pIdx)
}

func removePixelIndex(slice []int, value int) []int {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func scenarioCoordsFromWorld(wx, wy float64) (int, int) {
	return int((wx-shapeOffX)/shapeScaleX + 0.5), int((wy-shapeOffY)/shapeScaleY + 0.5)
}

func editModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyShift) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyShiftRight)
}

func editAddModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyAlt) ||
		ebiten.IsKeyPressed(ebiten.KeyAltLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyAltRight)
}

func editCreateRegionModifierPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
}

func editOwnerOptions(factions map[faction.FactionID]*faction.Faction) []string {
	ids := make([]string, 0, len(factions)+1)
	ids = append(ids, "")
	for fid := range factions {
		ids = append(ids, string(fid))
	}
	sort.Strings(ids[1:])
	return ids
}

func editTerrainOptions() []world.TerrainType {
	return editRegionTerrainOptions()
}

func editRegionTerrainOptions() []world.TerrainType {
	return []world.TerrainType{
		world.TerrainPlain,
		world.TerrainForest,
		world.TerrainMountain,
		world.TerrainPass,
		world.TerrainCoast,
	}
}

func editTerrainAreaOptions() []world.TerrainType {
	return []world.TerrainType{
		world.TerrainMountain,
		world.TerrainDesert,
		world.TerrainLake,
		world.TerrainRiver,
		world.TerrainDenseForest,
		world.TerrainSwamp,
	}
}

func (r *Renderer) editOwnerLabel(ownerID string) string {
	if ownerID == "" {
		return "(sahipsiz)"
	}
	if f, ok := r.gs.Factions[faction.FactionID(ownerID)]; ok && f != nil {
		name := f.NameTR
		if name == "" {
			name = f.Name
		}
		if name != "" {
			return name + "  [" + ownerID + "]"
		}
	}
	return ownerID
}

func nextSettlementID(region *world.Region) string {
	base := string(region.ID) + "_settlement_"
	for n := len(region.Settlements) + 1; ; n++ {
		id := base + itoa(n)
		used := false
		for _, settlement := range region.Settlements {
			if settlement.ID == id {
				used = true
				break
			}
		}
		if !used {
			return id
		}
	}
}

func nextRegionID(gs *state.GameState) world.RegionID {
	for n := len(gs.Regions) + 1; ; n++ {
		rid := world.RegionID("new_region_" + itoa(n))
		if _, used := gs.Regions[rid]; !used {
			return rid
		}
	}
}

func (r *Renderer) transferSelectedSettlement(targetID world.RegionID, x, y int) {
	source := r.gs.Regions[r.editSelectedRegion]
	target := r.gs.Regions[targetID]
	if source == nil || !canAddSettlementToRegion(target) || r.editSelectedSettlement < 0 ||
		r.editSelectedSettlement >= len(source.Settlements) {
		return
	}
	r.ensureSettlementDragSnapshot(targetID)

	settlement := source.Settlements[r.editSelectedSettlement]
	settlement.X = x
	settlement.Y = y
	source.Settlements = append(source.Settlements[:r.editSelectedSettlement], source.Settlements[r.editSelectedSettlement+1:]...)
	source.RecalculatePopulation()

	if settlement.IsCenter {
		settlement.IsCenter = false
		ensurePrimarySettlement(source)
		syncRegionSuccessorToOwner(source)
	}
	if !hasCapitalSettlement(target) {
		settlement.IsCenter = true
	}

	target.Settlements = append(target.Settlements, settlement)
	target.RecalculatePopulation()
	syncRegionSuccessorToOwner(target)
	world.EnsureRequiredSettlementBuildings(source, r.gs.IsCapitalRegion(source))
	world.EnsureRequiredSettlementBuildings(target, r.gs.IsCapitalRegion(target))
	r.editSelectedRegion = targetID
	r.editSelectedSettlement = len(target.Settlements) - 1
	r.worldMap.RebuildSettlementAnchors(r.gs)
	r.editDirty = true
}

func canAddSettlementToRegion(region *world.Region) bool {
	return region != nil && !region.IsSea
}

func hasCapitalSettlement(region *world.Region) bool {
	for _, settlement := range region.Settlements {
		if settlement.IsCenter {
			return true
		}
	}
	return false
}

func syncRegionSuccessorToOwner(region *world.Region) bool {
	if region == nil || !hasCapitalSettlement(region) {
		return false
	}
	return setRegionSuccessorToOwner(region)
}

func setRegionSuccessorToOwner(region *world.Region) bool {
	if region == nil || region.OwnerID == "" || region.SuccessorFactionID == region.OwnerID {
		return false
	}
	region.SuccessorFactionID = region.OwnerID
	return true
}

func ensurePrimarySettlement(region *world.Region) {
	if region == nil || len(region.Settlements) == 0 || hasCapitalSettlement(region) {
		return
	}
	region.Settlements[0].IsCenter = true
}
