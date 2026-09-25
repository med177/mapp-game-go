package render

import (
	"image/color"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"mapp-game-go/internal/army"
	"mapp-game-go/internal/diplomacy"
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
	help := "Sol: sec | Sag surukle: yerlesim tasi | Alt+sol: ekle | Ctrl+Alt+sol: bolge | Shift+sol: merkez"
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
	if r.editRenaming {
		DrawText(screen, r.editTextLabel()+": "+string(r.editTextRunes), float64(x)+14, float64(y)+80, FaceSmall, ColorGold)
		if r.editTextError != "" {
			DrawText(screen, r.editTextError, float64(x)+14, float64(y)+100, FaceSmall, ColorRed)
		}
	} else {
		DrawText(screen, debugState+"   V: debug   Esc: ana menu", float64(x)+14, float64(y)+80, FaceSmall, ColorGray)
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
	status := ""
	if r.editMapBuildPending {
		status = "Harita hazırlanıyor..."
	} else if r.editMapLastBuildDuration > 0 {
		status = "Son harita: " + itoa(int(r.editMapLastBuildDuration.Milliseconds())) +
			" ms  raster: " + itoa(int(r.editMapLastRasterDuration.Milliseconds())) +
			" ms  post: " + itoa(int(r.editMapLastPostProcessDuration.Milliseconds())) + " ms"
	}
	if status != "" {
		DrawText(screen, status, float64(x)+14, float64(y)+118, FaceSmall, ColorGold)
	}
}

func (r *Renderer) drawEditInspector(screen *ebiten.Image) {
	x, y, w, h := editInspectorRect()
	drawUIPanelRect(screen, gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}, color.RGBA{16, 20, 24, 226}, panelBorder, 1)

	drawEditInspectorLabel(screen, float64(x)+14, float64(y)+10, "EDITOR", ColorGold, gameui.TextMedium)
	r.drawEditInspectorTab(screen, editInspectorSettlement, "Yerleşim")
	r.drawEditInspectorTab(screen, editInspectorRegion, "Bölge")
	r.drawEditInspectorTab(screen, editInspectorFaction, "Devlet")
	r.drawEditInspectorTab(screen, editInspectorMap, "Harita")
	r.drawEditInspectorTab(screen, editInspectorTerrainArea, "Arazi")
	r.drawEditInspectorTab(screen, editInspectorData, "Veri")
	ly := float64(y) + 82

	if r.editInspectorTab == editInspectorTerrainArea {
		r.drawEditTerrainAreaInspector(screen, ly)
		drawUIDropdown(screen, r.editTerrainDropdown)
		drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
		return
	}

	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape {
		r.drawEditShapeInspector(screen, ly)
		drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
		return
	}

	if r.editInspectorTab == editInspectorData {
		r.drawEditScenarioDataInspector(screen, ly)
		drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
		return
	}

	if r.editInspectorTab == editInspectorFaction {
		r.drawEditDataInspector(screen, ly)
		drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
		return
	}

	region := r.gs.Regions[r.editSelectedRegion]
	if r.SelectedArmy != "" {
		if a, ok := r.gs.Armies[r.SelectedArmy]; ok && a != nil {
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Ordu: "+string(a.ID), ColorWhite, gameui.TextSmall)
			ly += 18
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Bolge: "+string(a.RegionID), ColorGray, gameui.TextSmall)
			ly += 18
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Birim: "+itoa(len(a.Units))+" / 20", ColorGray, gameui.TextSmall)
			if r.editInspectorTab == editInspectorSettlement {
				r.drawEditSettlementButtons(screen, region)
			} else {
				r.drawEditRegionButtons(screen, region)
			}
			// Seçili ordu için yukarıdaki erken dönüşte de dropdown'ı çizmek
			// gerekir. Aksi halde Birim Tipi düğmesi listeyi açar, ancak liste
			// görünmediği için kullanıcı bir seçim yapamaz.
			drawUIDropdown(screen, r.editUnitTypeDropdown)
			drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
			return
		}
	}

	if region == nil {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Haritadan bir bolge veya yerlesim sec.", ColorGray, gameui.TextSmall)
		r.drawEditRegionButtons(screen, nil)
		drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
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
	drawEditInspectorLabel(screen, float64(x)+14, ly, name, ColorWhite, gameui.TextSmall)
	ly += 18
	drawEditInspectorLabel(screen, float64(x)+14, ly, "ID: "+string(region.ID), ColorGray, gameui.TextSmall)
	ly += 18
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Tur: "+regionKind+"   Sahip: "+ownerLabel+"   Arazi: "+string(region.Terrain), ColorGray, gameui.TextSmall)
	ly += 18
	successorLabel := region.SuccessorFactionID
	if successorLabel == "" {
		successorLabel = "-"
	}
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Ardil Devlet: "+successorLabel, ColorGray, gameui.TextSmall)
	ly += 18
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Merkez: "+itoa(region.WorldX)+","+itoa(region.WorldY)+"   Yerlesim: "+settlementLabel, ColorGray, gameui.TextSmall)
	ly += 22
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Kilit: "+editBoolLabel(region.IsLocked)+"   Acilis: "+itoa(region.UnlockTurn)+"   Komsu: "+itoa(len(region.Neighbors)), ColorGray, gameui.TextSmall)
	ly += 20

	if r.hasEditSelection() {
		settlement := region.Settlements[r.editSelectedSettlement]
		sName := settlement.NameTR
		if sName == "" {
			sName = settlement.Name
		}
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Secili yerlesim: "+sName, ColorGold, gameui.TextSmall)
		ly += 18
		drawEditInspectorLabel(screen, float64(x)+14, ly, settlement.ID+"  "+string(settlement.Type)+"  nüfus "+itoa(settlement.Population)+"  "+itoa(settlement.X)+","+itoa(settlement.Y), ColorGray, gameui.TextSmall)
		if settlement.IsCenter {
			ly += 18
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Ana yerlesim", ColorGray, gameui.TextSmall)
		}
	} else if region.IsSea {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Deniz bolgesinde yerlesim yok.", ColorGray, gameui.TextSmall)
	} else {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Yerlesim secili degil.", ColorGray, gameui.TextSmall)
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
	drawEditInspectorSaveButton(screen, !r.terrainAreaEditPending())
}

func drawEditInspectorLabel(screen *ebiten.Image, x, y float64, text string, col color.Color, variant gameui.TextVariant) {
	drawUILabel(screen, gameui.Rect{X: x, Y: y, W: 404}, text, col, variant, gameui.TextAlignStart)
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
	settlementIDLabel := "Yerleşim ID"
	if region != nil && region.IsSea {
		settlementIDLabel = "ID Yok"
	} else if !canSettlement {
		settlementIDLabel = "ID Seç"
	}
	drawEditInspectorButton(screen, editButtonAddSettlement, addSettlementLabel, canAdd)
	drawEditInspectorButton(screen, editButtonSettlementType, settlementTypeLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonSetCenterSettlement, "Merkez Yap", canSettlement)
	drawEditInspectorButton(screen, editButtonRenameSettlement, renameSettlementLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonDeleteSettlement, deleteSettlementLabel, canSettlement)
	drawEditInspectorButton(screen, editButtonSettlementID, settlementIDLabel, canSettlement)
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
		terrainLabel = "Bölge Tipi"
	}
	drawEditInspectorButton(screen, editButtonRegionTerrain, terrainLabel, canRegion && !region.IsTerrainArea)
	nameTRLabel := "Ad TR"
	nameLabel := "Ad EN"
	nameEnabled := canRegion
	if region != nil && region.IsTerrainArea {
		nameTRLabel = "Ad TR"
		nameLabel = "Arazi Adı Yok"
		nameEnabled = false
	}
	drawEditInspectorButton(screen, editButtonRegionNameTR, nameTRLabel, canRegion && !region.IsTerrainArea)
	drawEditInspectorButton(screen, editButtonRegionName, nameLabel, nameEnabled)
	drawEditInspectorButton(screen, editButtonRegionID, "ID", canRegion)
	drawEditInspectorButton(screen, editButtonRegionLock, "Kilit", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockMinus, "-10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonUnlockPlus, "+10 Tur", canRegion)
	drawEditInspectorButton(screen, editButtonSyncNeighbors, "Komşu Sync", canRegion)
	neighborLabel := "Komşu Ekle"
	if r.editNeighborAddMode && r.editNeighborAddFrom == r.editSelectedRegion {
		neighborLabel = "Uygula"
	}
	drawEditInspectorButton(screen, editButtonAddNeighbor, neighborLabel, canRegion)
	drawEditInspectorButton(screen, editButtonEditRegionData, "Bölge Verileri", canRegion)
}

func drawEditInspectorSaveButton(screen *ebiten.Image, enabled bool) {
	rect := editInspectorButtonRect(editButtonSaveScenario)
	drawTinyPanelButton(screen, float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3]), "Değişiklikleri Kaydet", enabled)
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

	drawEditInspectorLabel(screen, float64(x)+14, ly, "DEVLET VE ORDU", ColorGold, gameui.TextSmall)
	ly += 22
	if f == nil {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Sahipli bolge veya ordu sec.", ColorGray, gameui.TextSmall)
		ly += 20
	} else {
		name := f.NameTR
		if name == "" {
			name = f.Name
		}
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Devlet: "+name+" ["+string(f.ID)+"]", ColorWhite, gameui.TextSmall)
		ly += 18
		drawEditInspectorLabel(screen, float64(x)+14, ly, economy.FormatResourceAmountTR(economy.ResourceGold, f.Gold)+"  "+economy.FormatResourceAmountTR(economy.ResourceGrain, f.Grain)+"  "+economy.FormatResourceAmountTR(economy.ResourceIron, f.Iron), ColorGray, gameui.TextSmall)
		ly += 18
		drawEditInspectorLabel(screen, float64(x)+14, ly, economy.FormatResourceAmountTR(economy.ResourceTimber, f.Timber)+"  "+economy.FormatResourceAmountTR(economy.ResourceSpice, f.Spice)+"  "+economy.FormatResourceAmountTR(economy.ResourceCloth, f.Cloth), ColorGray, gameui.TextSmall)
		ly += 18
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Playable: "+editBoolLabel(f.IsPlayable)+"  AI: "+itoa(f.AIAggressiveness), ColorGray, gameui.TextSmall)
	}
	ly += 24

	if r.SelectedArmy != "" {
		if a := r.gs.Armies[r.SelectedArmy]; a != nil {
			r.ensureEditSelectedUnitType(a)
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Ordu: "+string(a.ID), ColorGold, gameui.TextSmall)
			ly += 18
			kind := "Kara"
			if a.IsNaval {
				kind = "Donanma"
			}
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Tip: "+kind+"  Sahip: "+a.OwnerID+"  Bolge: "+string(a.RegionID), ColorGray, gameui.TextSmall)
			ly += 18
			drawEditInspectorLabel(screen, float64(x)+14, ly, "Birim: "+itoa(len(a.Units))+" / "+itoa(army.MaxArmySize)+"  Secili: "+r.editSelectedUnitType, ColorGray, gameui.TextSmall)
			ly += 18
			r.drawEditArmyUnitCounts(screen, a, float64(x)+14, ly)
		}
	} else {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Ordu secili degil.", ColorGray, gameui.TextSmall)
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
	drawEditInspectorLabel(screen, float64(x)+14, ly, "SENARYO VERİLERİ", ColorGold, gameui.TextSmall)
	ly += 24
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Bu senaryodaki düzenlemeler geçici olarak tutulur.", ColorGray, gameui.TextSmall)
	ly += 18
	drawEditInspectorLabel(screen, float64(x)+14, ly, "Kaydet düğmesi tüm sekmelerde panelin altındadır.", ColorGray, gameui.TextSmall)
	ly += 24
	if r.editDirty {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Durum: Kaydedilmemiş değişiklikler var.", ColorGold, gameui.TextSmall)
	} else {
		drawEditInspectorLabel(screen, float64(x)+14, ly, "Durum: Tüm değişiklikler kayıtlı.", ColorGray, gameui.TextSmall)
	}
}

func (r *Renderer) drawEditArmyUnitCounts(screen *ebiten.Image, a *army.Army, x, y float64) {
	if len(a.Units) == 0 {
		drawEditInspectorLabel(screen, x, y, "Birim yok.", ColorGray, gameui.TextSmall)
		return
	}
	// Devlet sekmesinin düğmeleriyle çakışmaması için ayrıntılı dağılımı
	// satır satır büyütmek yerine tek satırlık özet tutuyoruz.
	drawEditInspectorLabel(screen, x, y, "Birim toplamı: "+itoa(len(a.Units)), ColorGray, gameui.TextSmall)
}

func (r *Renderer) drawEditFactionForm(screen *ebiten.Image) {
	if !r.editFactionForm.show {
		return
	}
	x, y, w, h := editFactionFormRect()
	formRect := gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}
	drawUIPanelRect(screen, formRect, color.RGBA{14, 18, 22, 248}, panelBorder, 1)
	title := "DEVLET EKLE"
	if !r.editFactionForm.create {
		title = "DEVLET DÜZENLE"
	}
	drawUILabel(screen, gameui.Rect{X: float64(x) + 24, Y: float64(y) + 16, W: float64(w) - 48}, title, ColorGold, gameui.TextLarge, gameui.TextAlignStart)
	drawUISectionLabel(screen, float64(x)+24, float64(y)+52, "KİMLİK VE KAYNAKLAR")

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

	drawUISectionLabel(screen, float64(x)+24, float64(y)+326, "DİPLOMASİ")
	drawUISectionLabel(screen, float64(x)+396, float64(y)+326, "AYARLAR VE RENK")
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
	drawUILabel(screen, gameui.Rect{X: float64(x) + 24, Y: float64(y) + 436, W: 340}, "İlişki skoru: "+r.editFactionForm.relationScore, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	overlordLabel := "Vassal üst devleti: yok"
	if r.editFactionForm.overlordID != "" {
		overlordLabel = "Vassal üst devleti: " + string(r.editFactionForm.overlordID)
	}
	drawEditFactionFormButton(screen, editFactionFormOverlordTarget, overlordLabel)

	col := r.editFactionForm.color
	preview := editFactionFormColorPreviewRect()
	vector.FillRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), color.RGBA{col[0], col[1], col[2], 255}, false)
	vector.StrokeRect(screen, float32(preview[0]), float32(preview[1]), float32(preview[2]), float32(preview[3]), 1, ColorGold, false)
	drawUILabel(screen, gameui.Rect{X: float64(x) + 396, Y: float64(y) + 446, W: 348}, "Renk: "+itoa(int(col[0]))+", "+itoa(int(col[1]))+", "+itoa(int(col[2])), ColorGray, gameui.TextSmall, gameui.TextAlignStart)
	drawEditFactionFormButton(screen, editFactionFormRedMinus, "R-")
	drawEditFactionFormButton(screen, editFactionFormRedPlus, "R+")
	drawEditFactionFormButton(screen, editFactionFormGreenMinus, "G-")
	drawEditFactionFormButton(screen, editFactionFormGreenPlus, "G+")
	drawEditFactionFormButton(screen, editFactionFormBlueMinus, "B-")
	drawEditFactionFormButton(screen, editFactionFormBluePlus, "B+")

	if r.editFactionForm.errorText != "" {
		drawUILabel(screen, gameui.Rect{X: float64(x) + 24, Y: float64(y) + float64(h) - 82, W: float64(w) - 48}, r.editFactionForm.errorText, ColorRed, gameui.TextSmall, gameui.TextAlignStart)
	}
	drawEditFactionFormButton(screen, editFactionFormSave, "Kaydet")
	drawEditFactionFormButton(screen, editFactionFormCancel, "Iptal")
}

func drawEditFactionFormButton(screen *ebiten.Image, kind editFactionFormButton, label string) {
	button := buildEditFactionFormButton(kind, label)
	drawUIButtonWidget(screen, button, tinyButtonStyle)
}

func editRectButton(r uiRect, label string) gameui.Button {
	return gameui.NewButton(r[0], r[1], r[2], r[3], label)
}

func (r *Renderer) drawFactionFormField(screen *ebiten.Image, field editFactionFormField, label, value string) {
	rect := editFactionFieldRect(field)
	box := gameui.NewTextBox(rect[0], rect[1], rect[2], rect[3], "")
	box.Value = value
	box.Focused = r.editFactionForm.active == field
	gameui.DrawTextBox(screen, box, editFactionFormTextBoxStyle(), renderText)
	drawUILabel(screen, gameui.Rect{X: rect[0], Y: rect[1] - 16, W: rect[2]}, label, ColorGray, gameui.TextSmall, gameui.TextAlignStart)
}

func editFactionFormTextBoxStyle() gameui.TextBoxStyle {
	return gameui.TextBoxStyle{
		BG:          color.RGBA{28, 32, 38, 235},
		Border:      color.RGBA{120, 105, 60, 210},
		Focused:     ColorGold,
		Text:        ColorWhite,
		Placeholder: ColorGray,
		BorderWidth: 1,
		TextOffsetX: 8,
		TextOffsetY: 8,
		TextVariant: gameui.TextSmall,
	}
}

func rectXYWH(rect uiRect) (float32, float32, float32, float32) {
	return float32(rect[0]), float32(rect[1]), float32(rect[2]), float32(rect[3])
}

func editFactionFormRect() (float32, float32, float32, float32) {
	const w, h = float32(760), float32(620)
	return float32(ScreenWidth)/2 - w/2, float32(ScreenHeight)/2 - h/2, w, h
}

func editFactionFormHit(mx, my float64) bool {
	x, y, w, h := editFactionFormRect()
	return (gameui.Rect{X: float64(x), Y: float64(y), W: float64(w), H: float64(h)}).Hit(mx, my)
}

func editFactionFieldRect(field editFactionFormField) uiRect {
	x, y, _, _ := editFactionFormRect()
	left := float64(x) + 24
	right := float64(x) + 396
	top := float64(y) + 78
	const fw, fh, gap = float64(348), float64(32), float64(16)
	row := func(n int) float64 { return top + float64(n)*(fh+gap) }
	switch field {
	case editFactionFieldID:
		return uiRect{left, row(0), fw, fh}
	case editFactionFieldNameTR:
		return uiRect{right, row(0), fw, fh}
	case editFactionFieldName:
		return uiRect{right, row(1), fw, fh}
	case editFactionFieldGold:
		return uiRect{right, row(2), fw/2 - 6, fh}
	case editFactionFieldGrain:
		return uiRect{right + fw/2 + 6, row(2), fw/2 - 6, fh}
	case editFactionFieldIron:
		return uiRect{left, row(3), fw/2 - 6, fh}
	case editFactionFieldTimber:
		return uiRect{left + fw/2 + 6, row(3), fw/2 - 6, fh}
	case editFactionFieldSpice:
		return uiRect{right, row(4), fw/2 - 6, fh}
	case editFactionFieldCloth:
		return uiRect{right + fw/2 + 6, row(4), fw/2 - 6, fh}
	case editFactionFieldAI:
		return uiRect{left, row(1), fw/2 - 6, fh}
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
	editFactionFormOverlordTarget
	editFactionFormRedMinus
	editFactionFormRedPlus
	editFactionFormGreenMinus
	editFactionFormGreenPlus
	editFactionFormBlueMinus
	editFactionFormBluePlus
)

func editFactionFormButtonRect(kind editFactionFormButton) uiRect {
	x, y, w, h := editFactionFormRect()
	left := float64(x) + 24
	right := float64(x) + 396
	switch kind {
	case editFactionFormReligion:
		return uiRect{right, float64(y) + 354, 166, 32}
	case editFactionFormPlayable:
		return uiRect{right + 182, float64(y) + 354, 166, 32}
	case editFactionFormRelationTarget:
		return uiRect{left, float64(y) + 354, 348, 32}
	case editFactionFormRelationStance:
		return uiRect{left, float64(y) + 394, 160, 32}
	case editFactionFormRelationScoreMinus:
		return uiRect{left + 172, float64(y) + 394, 78, 32}
	case editFactionFormRelationScorePlus:
		return uiRect{left + 258, float64(y) + 394, 78, 32}
	case editFactionFormOverlordTarget:
		return uiRect{left, float64(y) + 474, 348, 32}
	case editFactionFormRedMinus:
		return uiRect{right, float64(y) + 474, 52, 28}
	case editFactionFormRedPlus:
		return uiRect{right + 58, float64(y) + 474, 52, 28}
	case editFactionFormGreenMinus:
		return uiRect{right + 112, float64(y) + 474, 52, 28}
	case editFactionFormGreenPlus:
		return uiRect{right + 170, float64(y) + 474, 52, 28}
	case editFactionFormBlueMinus:
		return uiRect{right + 224, float64(y) + 474, 52, 28}
	case editFactionFormBluePlus:
		return uiRect{right + 282, float64(y) + 474, 52, 28}
	case editFactionFormSave:
		return uiRect{float64(x) + 24, float64(y) + float64(h) - 48, 140, 32}
	case editFactionFormCancel:
		return uiRect{float64(x) + float64(w) - 164, float64(y) + float64(h) - 48, 140, 32}
	default:
		return uiRect{}
	}
}

func buildEditFactionFormButton(kind editFactionFormButton, label string) gameui.Button {
	return editRectButton(editFactionFormButtonRect(kind), label)
}

func editFactionFormColorPreviewRect() uiRect {
	x, y, _, _ := editFactionFormRect()
	return uiRect{float64(x) + 396, float64(y) + 410, 348, 28}
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
	editButtonSettlementID
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
	editButtonEditRegionData
	editButtonDeleteRegion
	editButtonDeleteSettlement
	editButtonSaveScenario
	editButtonShapePaint
	editButtonShapeErase
	editButtonShapeRegionPaint
	editButtonShapeRegionErase
	editButtonShapeBrushMinus
	editButtonShapeBrushPlus
	editButtonShapeNew
	editButtonLandPassageAdd
	editButtonLandPassageAdjust
	editButtonLandPassageDelete
	editButtonAddNeighbor
	editButtonTerrainArea
	editButtonTerrainAreaAppend
	editButtonTerrainAreaType
	editButtonTerrainAreaCost
	editButtonTerrainAreaAttrition
	editButtonTerrainAreaDelete
	editButtonTerrainAreaCancel
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
	case editButtonSettlementID:
		return rightRect(2)
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
	case editButtonEditRegionData:
		return full(6)
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
	case editButtonShapeNew:
		return uiRect{left + bw - 140, float64(y) + 128, 140, bh}
	case editButtonLandPassageAdd:
		return leftRect(3)
	case editButtonLandPassageAdjust:
		return rightRect(3)
	case editButtonLandPassageDelete:
		return leftRect(4)
	case editButtonTerrainArea:
		return leftRect(6)
	case editButtonTerrainAreaAppend:
		return rightRect(5)
	case editButtonTerrainAreaType:
		return leftRect(5)
	case editButtonTerrainAreaCost:
		return rightRect(6)
	case editButtonTerrainAreaCancel:
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
	widths := [...]float64{82, 44, 44, 48, 48, 44}
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
	case editInspectorTerrainArea:
		index = 4
	case editInspectorData:
		index = 5
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
		editButtonSettlementID,
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
		editButtonEditRegionData,
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

func editShapeInspectorButtonKinds() []editInspectorButton {
	// Bu liste yalnız drawEditShapeInspector/drawEditShapeLandPassageButtons
	// tarafından gerçekten çizilen shape sekmesi düğmelerini içerir. Bölge
	// sekmesine ait Komşu Ekle gibi düğmeler burada bulunmaz; aksi halde aynı
	// rect'i paylaşan görünmez bir hit-test alanı oluşur.
	return []editInspectorButton{
		editButtonShapePaint,
		editButtonShapeErase,
		editButtonShapeRegionPaint,
		editButtonShapeRegionErase,
		editButtonShapeBrushMinus,
		editButtonShapeBrushPlus,
		editButtonShapeNew,
		editButtonLandPassageAdd,
		editButtonLandPassageAdjust,
		editButtonLandPassageDelete,
	}
}

func editTerrainAreaInspectorButtonKinds() []editInspectorButton {
	return []editInspectorButton{
		editButtonTerrainArea,
		editButtonTerrainAreaAppend,
		editButtonTerrainAreaType,
		editButtonTerrainAreaCost,
		editButtonTerrainAreaAttrition,
		editButtonTerrainAreaDelete,
		editButtonTerrainAreaCancel,
	}
}

func (r *Renderer) editShapeInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range editShapeInspectorButtonKinds() {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	return editButtonNone
}

func (r *Renderer) editTerrainAreaInspectorButtonAt(mx, my float64) editInspectorButton {
	for _, kind := range []editInspectorButton{editButtonRegionTerrain, editButtonRegionNameTR} {
		if buildEditInspectorActionButton(kind, "").HitTest(mx, my) {
			return kind
		}
	}
	for _, kind := range editTerrainAreaInspectorButtonKinds() {
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
		buildEditInspectorTabButton(editInspectorTerrainArea, "").HitTest(mx, my) ||
		buildEditInspectorTabButton(editInspectorData, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(mx, my) {
		return editButtonSaveScenario
	}
	if r.editInspectorTab == editInspectorTerrainArea {
		return r.editTerrainAreaInspectorButtonAt(mx, my)
	}
	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape {
		kind := r.editShapeInspectorButtonAt(mx, my)
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
	return r.editRegionInspectorButtonAt(mx, my)
}

func (r *Renderer) editRegionInspectorButtonAt(mx, my float64) editInspectorButton {
	kind := editRegionInspectorButtonAt(mx, my)
	if r == nil || r.gs == nil {
		return kind
	}
	region := r.gs.Regions[r.editSelectedRegion]
	if region != nil && region.IsTerrainArea && (kind == editButtonRegionTerrain || kind == editButtonRegionNameTR) {
		return editButtonNone
	}
	return kind
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

const editRegionCenterHitRadius = 12.0

type editRegionCenterMarker struct {
	ID    world.RegionID
	SX    float64
	SY    float64
	IsSea bool
}

func (r *Renderer) invalidateEditRegionCenterMarkers() {
	if r == nil {
		return
	}
	r.editCenterMarkersVersion++
}

// editRegionCenterMarkers ortak ekran geometrisini draw, hit-test ve cursor
// akışlarına sağlar. Kamera veya merkez verisi değişmedikçe region map'i tekrar
// taranmaz.
func (r *Renderer) editRegionCenterMarkers() []editRegionCenterMarker {
	if r == nil || r.gs == nil {
		return nil
	}
	if r.editCenterMarkersGameState == r.gs &&
		r.editCenterMarkersCacheVersion == r.editCenterMarkersVersion &&
		r.editCenterMarkersCamX == r.camX &&
		r.editCenterMarkersCamY == r.camY &&
		r.editCenterMarkersCamScale == r.camScale {
		return r.editCenterMarkers
	}

	markers := make([]editRegionCenterMarker, 0, len(r.gs.Regions))
	for rid, region := range r.gs.Regions {
		if region == nil || region.IsLocked || region.IsTerrainArea {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		markers = append(markers, editRegionCenterMarker{ID: rid, SX: sx, SY: sy, IsSea: region.IsSea})
	}
	// Map iteration sırası hit-test eşitliklerinde kararsızlık üretmesin.
	sort.Slice(markers, func(i, j int) bool { return markers[i].ID < markers[j].ID })
	r.editCenterMarkers = markers
	r.editCenterMarkersCamX = r.camX
	r.editCenterMarkersCamY = r.camY
	r.editCenterMarkersCamScale = r.camScale
	r.editCenterMarkersCacheVersion = r.editCenterMarkersVersion
	r.editCenterMarkersGameState = r.gs
	return markers
}

// editRegionCenterAt, rasterdaki bölge yerine Edit Mode'da çizilen merkez
// işaretini hedefler. Merkez işareti başka bir bölgenin raster alanının üstünde
// kalabileceği için merkez seçimi her zaman raster RegionAt'tan önce yapılır.
func (r *Renderer) editRegionCenterAt(fx, fy float64) (world.RegionID, bool) {
	if r == nil || r.gs == nil {
		return "", false
	}
	bestRID := world.RegionID("")
	bestDist := editRegionCenterHitRadius * editRegionCenterHitRadius
	for _, marker := range r.editRegionCenterMarkers() {
		dx, dy := fx-marker.SX, fy-marker.SY
		dist := dx*dx + dy*dy
		if dist <= bestDist {
			bestDist = dist
			bestRID = marker.ID
		}
	}
	return bestRID, bestRID != ""
}

func (r *Renderer) drawEditRegionCenters(screen *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	hoveredID, _ := r.editRegionCenterAt(float64(mx), float64(my))
	for _, marker := range r.editRegionCenterMarkers() {
		if marker.SX < -12 || marker.SY < -12 || marker.SX > float64(ScreenWidth+12) || marker.SY > float64(ScreenHeight+12) {
			continue
		}
		col := color.RGBA{80, 220, 255, 190}
		if marker.IsSea {
			col = color.RGBA{120, 210, 255, 210}
		}
		if marker.ID == r.editSelectedRegion && r.editSelectedSettlement < 0 {
			if marker.IsSea {
				col = color.RGBA{70, 235, 255, 245}
			} else {
				col = color.RGBA{255, 190, 45, 240}
			}
		}
		x, y := float32(marker.SX), float32(marker.SY)
		vector.StrokeCircle(screen, x, y, 6, 1.5, col, true)
		vector.StrokeLine(screen, x-8, y, x+8, y, 1.5, col, true)
		vector.StrokeLine(screen, x, y-8, x, y+8, 1.5, col, true)
		if marker.ID == hoveredID {
			vector.StrokeCircle(screen, x, y, 11, 2.5, color.RGBA{255, 220, 70, 245}, true)
		}
	}
}

// drawEditTerrainAreaCenters gösterge amaçlıdır: arazi alanının otomatik
// poligon merkezini görünür tutar, ancak merkez editör seçim/taşıma hedefi
// değildir.
func (r *Renderer) drawEditTerrainAreaCenters(screen *ebiten.Image) {
	if r == nil || r.gs == nil {
		return
	}
	for _, region := range r.gs.Regions {
		if region == nil || !region.IsTerrainArea {
			continue
		}
		sx, sy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
		col := color.RGBA{255, 220, 70, 220}
		if region.ID == r.editSelectedRegion {
			col = color.RGBA{255, 180, 35, 250}
		}
		x, y := float32(sx), float32(sy)
		vector.StrokeCircle(screen, x, y, 6, 1.5, col, true)
		vector.StrokeLine(screen, x-8, y, x+8, y, 1.5, col, true)
		vector.StrokeLine(screen, x, y-8, x, y+8, 1.5, col, true)
	}
}

func (r *Renderer) drawEditVoronoiDebug(screen *ebiten.Image) {
	if !r.editVoronoiDebug || r.editTerrainAreaMode {
		return
	}
	rid := r.editSelectedRegion
	if rid == "" {
		mx, my := ebiten.CursorPosition()
		rid = r.editRegionAt(float64(mx), float64(my))
	}
	region := r.gs.Regions[rid]
	if region == nil {
		return
	}

	if r.editVoronoiDebugWorldMap != r.worldMap || r.editVoronoiDebugRegion != rid {
		r.editVoronoiDebugVisualNeighborBuf = r.worldMap.VisualNeighbors(rid, r.editVoronoiDebugVisualNeighborBuf[:0])
		r.editVoronoiDebugBoundaryPixelBuf = r.worldMap.BoundaryPixels(rid, r.editVoronoiDebugBoundaryPixelBuf[:0])
		r.editVoronoiDebugWorldMap = r.worldMap
		r.editVoronoiDebugRegion = rid
	}
	r.drawEditVoronoiBoundary(screen, r.editVoronoiDebugBoundaryPixelBuf)

	cx, cy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	for _, nrid := range r.editVoronoiDebugVisualNeighborBuf {
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
		if visualNeighborContains(r.editVoronoiDebugVisualNeighborBuf, nrid) {
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

// drawEditNeighborLinks, editörün sürekli göstereceği ucuz komşuluk
// görünümüdür. Voronoi sınırı/BoundaryPixels hesabı yalnız ayrıca açılan
// debug görünümünde kalır.
func (r *Renderer) drawEditNeighborLinks(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.editSelectedRegion == "" {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	visual := r.cachedEditVisualNeighbors(region)
	cx, cy := r.worldToScreen(wcX(region.WorldX), wcY(region.WorldY))
	for _, nrid := range region.Neighbors {
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{180, 180, 180, 150}
		if visualNeighborContains(visual, nrid) {
			col = color.RGBA{90, 220, 125, 190}
		}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1.5, col, true)
		mx, my := (cx+nx)/2, (cy+ny)/2
		vector.FillRect(screen, float32(mx)-3, float32(my)-3, 6, 6, col, true)
	}
	for _, nrid := range visual {
		if regionHasNeighbor(region, nrid) {
			continue
		}
		neighbor := r.gs.Regions[nrid]
		if neighbor == nil {
			continue
		}
		nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
		col := color.RGBA{235, 80, 80, 220}
		vector.StrokeLine(screen, float32(cx), float32(cy), float32(nx), float32(ny), 1.5, col, true)
		mx, my := (cx+nx)/2, (cy+ny)/2
		vector.FillRect(screen, float32(mx)-3, float32(my)-3, 6, 6, col, true)
	}
	if r.editNeighborAddMode && r.editNeighborAddFrom == region.ID {
		for _, nrid := range r.editNeighborAddTargets {
			neighbor := r.gs.Regions[nrid]
			if neighbor == nil {
				continue
			}
			nx, ny := r.worldToScreen(wcX(neighbor.WorldX), wcY(neighbor.WorldY))
			drawEditNeighborArrow(screen, cx, cy, nx, ny, color.RGBA{255, 220, 70, 245})
		}
	}
}

func (r *Renderer) cachedEditVisualNeighbors(region *world.Region) []world.RegionID {
	if r == nil || r.worldMap == nil || region == nil {
		return nil
	}
	if r.editVisualNeighborWorldMap != r.worldMap || r.editVisualNeighborRegion != region.ID {
		if region.IsTerrainArea {
			r.editVisualNeighborBuf = r.terrainAreaVisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
		} else {
			r.editVisualNeighborBuf = r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
		}
		r.editVisualNeighborWorldMap = r.worldMap
		r.editVisualNeighborRegion = region.ID
	}
	return r.editVisualNeighborBuf
}

func (r *Renderer) invalidateEditVisualNeighborCache() {
	if r == nil {
		return
	}
	r.editVisualNeighborWorldMap = nil
	r.editVisualNeighborRegion = ""
}

func drawEditNeighborArrow(screen *ebiten.Image, x1, y1, x2, y2 float64, col color.RGBA) {
	vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 2.5, col, true)
	angle := math.Atan2(y2-y1, x2-x1)
	const head = 9.0
	leftX := x2 - math.Cos(angle-0.5)*head
	leftY := y2 - math.Sin(angle-0.5)*head
	rightX := x2 - math.Cos(angle+0.5)*head
	rightY := y2 - math.Sin(angle+0.5)*head
	vector.StrokeLine(screen, float32(x2), float32(y2), float32(leftX), float32(leftY), 2.5, col, true)
	vector.StrokeLine(screen, float32(x2), float32(y2), float32(rightX), float32(rightY), 2.5, col, true)
}

// drawSelectedEditRegionBoundary, Bölge sekmesinde seçili bölgenin yalnızca
// gerçek raster sınırını noktalı olarak gösterir. Voronoi veya görsel komşu
// hesabı kullanmaz; sınır yalnızca harita yeniden oluşturulduğunda hesaplanır.
func (r *Renderer) drawSelectedEditRegionBoundary(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.worldMap == nil || r.editInspectorTab != editInspectorRegion || r.editSelectedRegion == "" {
		return
	}
	if r.gs.Regions[r.editSelectedRegion] == nil {
		return
	}
	if r.editBoundaryWorldMap != r.worldMap || r.editBoundaryRegion != r.editSelectedRegion {
		r.editBoundaryPixelBuf = r.worldMap.BoundaryPixels(r.editSelectedRegion, r.editBoundaryPixelBuf[:0])
		r.editBoundaryWorldMap = r.worldMap
		r.editBoundaryRegion = r.editSelectedRegion
	}
	r.drawEditVoronoiBoundary(screen, r.editBoundaryPixelBuf)
}

// drawEditCountryHover, Devlet sekmesinde tıklanarak seçilmiş ülkenin shape'ini
// vurgular. Fare hareketiyle seçim veya raster sorgusu yapmaz.
func (r *Renderer) drawEditCountryHover(screen *ebiten.Image) {
	if r == nil || r.gs == nil || r.editInspectorTab != editInspectorFaction {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil || region.IsTerrainArea || region.ShapeID == "" {
		return
	}
	rings := r.gs.ShapeData.Shapes[region.ShapeID]
	if len(rings) == 0 {
		return
	}
	col := color.RGBA{255, 220, 70, 235}
	for _, ring := range rings {
		if len(ring) < 2 {
			continue
		}
		for i, point := range ring {
			next := ring[(i+1)%len(ring)]
			x1, y1 := r.worldToScreen(shapeRasterWorldPoint(point))
			x2, y2 := r.worldToScreen(shapeRasterWorldPoint(next))
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3, col, true)
		}
	}
}

func (r *Renderer) drawEditVoronoiLegendOverlay(screen *ebiten.Image) {
	if r.editTerrainAreaMode || r.editSelectedRegion == "" {
		return
	}
	rid := r.editSelectedRegion
	visual := []world.RegionID(nil)
	if r.editVoronoiDebug {
		rid = r.editVoronoiDebugRegion
		visual = r.editVoronoiDebugVisualNeighborBuf
	}
	r.drawEditVoronoiLegend(screen, rid, visual)
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
	title := "BÖLGE BİLGİSİ"
	line := "V ile Voronoi debug açılır"
	legend := "Komşuluk: JSON kaydı"
	if r.editVoronoiDebug {
		title = "VORONOI DEBUG"
		line = "camgobegi: raster sinir"
		legend = "yesil: gorunen+JSON   kirmizi: sadece gorunen"
	}
	DrawText(screen, title, float64(x)+12, float64(y)+10, FaceSmall, ColorGold)
	DrawText(screen, line, float64(x)+12, float64(y)+31, FaceSmall, ColorGray)
	DrawText(screen, legend, float64(x)+12, float64(y)+48, FaceSmall, ColorGray)

	hoverLabel := "kapalı"
	if r.editVoronoiDebug {
		mx, my := ebiten.CursorPosition()
		wx, wy := r.screenToWorld(float64(mx), float64(my))
		hover := r.worldMap.RegionAt(int(wx), int(wy))
		sx, sy := scenarioCoordsFromWorld(wx, wy)
		hoverLabel = string(hover) + "  " + itoa(sx) + "," + itoa(sy)
	}
	DrawText(screen, "Hover: "+hoverLabel, float64(x)+12, float64(y)+68, FaceSmall, ColorWhite)
	if rid != "" {
		region := r.gs.Regions[rid]
		jsonCount := 0
		if region != nil {
			jsonCount = len(region.Neighbors)
		}
		visualLabel := "-"
		if r.editVoronoiDebug {
			visualLabel = itoa(len(visual))
		}
		DrawText(screen, "Secili: "+string(rid)+"  visual/json: "+visualLabel+"/"+itoa(jsonCount),
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

func cloneSettlements(settlements []world.Settlement) []world.Settlement {
	if settlements == nil {
		return nil
	}
	clone := make([]world.Settlement, len(settlements))
	copy(clone, settlements)
	return clone
}

func editCtrlPressed() bool {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) ||
		ebiten.IsKeyPressed(ebiten.KeyControlLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyControlRight)
	return ctrl
}

func (r *Renderer) handleEditModeInput() InputAction {
	if r.editRenaming {
		return r.handleEditRenameInput()
	}
	if r.editNewShapeModal.show {
		return r.handleEditNewShapeModalInput()
	}
	if r.editFactionForm.show {
		return r.handleEditFactionFormInput()
	}
	if r.editRegionForm.show {
		return r.handleEditRegionFormInput()
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
	if r.keyJustPressed(ebiten.KeyDelete) && !r.terrainAreaEditPending() &&
		r.editTerrainAreaSelected >= 0 && r.editTerrainAreaSelected < len(r.gs.TerrainAreas) {
		r.deleteSelectedTerrainArea()
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
		r.editDraggingSettlement = false
		return InputAction{}
	}
	inspectorOverlayOpen := r.editOwnerDropdown.IsOpen() ||
		r.editSuccessorDropdown.IsOpen() ||
		r.editTerrainDropdown.IsOpen() ||
		r.editSettlementTypeDropdown.IsOpen() ||
		r.editUnitTypeDropdown.IsOpen()
	if r.editTerrainAreaMode && rightJustPressed && !inspectorOverlayOpen &&
		!editInspectorHit(fx, fy) && !r.editShapeHelpPanelHit(fx, fy) {
		r.resetTerrainAreaDrawing()
		return InputAction{}
	}
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
			r.editInspectorTab = editInspectorSettlement
			r.editDraggingRegion = false
			r.editRenaming = false
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
	if r.editTerrainAreaMode && !inspectorOverlayOpen && !r.editShapeHelpPanelHit(fx, fy) {
		if r.editShapePaintPending {
			return InputAction{}
		}
		if rightJustPressed {
			return InputAction{}
		}
		if leftJustPressed {
			if len(r.editTerrainAreaPolygon) == 0 && r.selectTerrainAreaAt(fx, fy) {
				return InputAction{}
			}
			if r.terrainAreaPolygonStartHovered(fx, fy) {
				r.finishTerrainAreaPolygon()
				return InputAction{}
			}
			r.addTerrainAreaPolygonPoint(fx, fy)
			return InputAction{}
		}
	}

	if (r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape || r.editInspectorTab == editInspectorRegion || r.editInspectorTab == editInspectorTerrainArea) && leftJustPressed && r.editShapeHelpPanelHit(fx, fy) {
		return InputAction{}
	}

	if r.editInspectorTab == editInspectorMap || r.editInspectorTab == editInspectorShape || r.editInspectorTab == editInspectorRegion || r.editInspectorTab == editInspectorTerrainArea {
		if leftJustPressed && r.beginShapePaintStroke(fx, fy) {
			return InputAction{}
		}
		if r.editShapePainting {
			r.continueShapePaintStroke(fx, fy)
			return InputAction{}
		}
	}

	if r.editDraggingRegion && !leftPressed {
		changed := r.finishRegionCenterDrag()
		r.editDraggingRegion = false
		if changed {
			region := r.gs.Regions[r.editSelectedRegion]
			if region != nil && r.regionCenterAffectsRaster(region) {
				if region.IsSea {
					r.requestEditWorldMapRebuild()
				} else {
					r.requestEditWorldMapRebuildForShape(region.ShapeID)
				}
			} else {
				r.invalidateEditRegionCenterMarkers()
			}
		}
	}

	if leftJustPressed {
		r.editSuccessorDropdown.Close()
		if editModifierPressed() {
			rid, centerHit := r.editRegionCenterAt(fx, fy)
			if !centerHit {
				rid = r.editRegionAt(fx, fy)
			}
			if rid != "" {
				r.editOwnerDropdown.Close()
				r.editTerrainDropdown.Close()
				r.editSettlementTypeDropdown.Close()
				r.editUnitTypeDropdown.Close()
				r.editSelectedRegion = rid
				r.setEditFactionFromRegion(rid)
				r.editSelectedSettlement = -1
				r.editInspectorTab = editInspectorRegion
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
			r.editInspectorTab = editInspectorSettlement
			r.editSelectedSettlement = -1
			r.editDraggingSettlement = false
			r.editDraggingRegion = false
			r.editRenaming = false
			return InputAction{}
		}

		if r.editShapeTool == editShapeToolNone && !r.editTerrainAreaMode && r.selectTerrainAreaAt(fx, fy) {
			return InputAction{}
		}

		if rid, ok := r.editRegionCenterAt(fx, fy); ok {
			r.editOwnerDropdown.Close()
			r.editTerrainDropdown.Close()
			r.editSettlementTypeDropdown.Close()
			r.editUnitTypeDropdown.Close()
			r.SelectedArmy = ""
			r.editSelectedRegion = rid
			r.rememberEditSelectedWorldPoint(rid, fx, fy)
			r.syncSelectedTerrainArea(rid)
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = -1
			r.editInspectorTab = editInspectorRegion
			r.editRenaming = false
			r.editDraggingRegion = false
			r.editDraggingSettlement = false
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
			r.editInspectorTab = editInspectorSettlement
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
			r.rememberEditSelectedWorldPoint(rid, fx, fy)
			r.syncSelectedTerrainArea(rid)
			r.setEditFactionFromRegion(rid)
			r.editSelectedSettlement = -1
			r.editInspectorTab = editInspectorRegion
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
	r.editTerrainAreaMoveCost = 0
	r.editTerrainAreaAttritionCost = 0
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

func (r *Renderer) selectTerrainAreaAt(fx, fy float64) bool {
	if r == nil || r.worldMap == nil || r.gs == nil {
		return false
	}
	rid, ok := r.terrainAreaRegionAt(fx, fy)
	if !ok {
		return false
	}
	r.editOwnerDropdown.Close()
	r.editTerrainDropdown.Close()
	r.editSettlementTypeDropdown.Close()
	r.editUnitTypeDropdown.Close()
	r.SelectedArmy = ""
	r.editSelectedRegion = rid
	r.rememberEditSelectedWorldPoint(rid, fx, fy)
	r.syncSelectedTerrainArea(rid)
	r.setEditFactionFromRegion(rid)
	r.editSelectedSettlement = -1
	r.editInspectorTab = editInspectorTerrainArea
	r.editRenaming = false
	r.editDraggingRegion = false
	r.editDraggingSettlement = false
	return true
}

func (r *Renderer) terrainAreaRegionAt(fx, fy float64) (world.RegionID, bool) {
	if r == nil || r.worldMap == nil || r.gs == nil {
		return "", false
	}
	wx, wy := r.screenToWorld(fx, fy)
	cellX, cellY := int(math.Floor(wx)), int(math.Floor(wy))
	rid := r.worldMap.RegionAt(cellX, cellY)
	if region := r.gs.Regions[rid]; region != nil && region.IsTerrainArea {
		return rid, true
	}
	for i := len(r.gs.TerrainAreas) - 1; i >= 0; i-- {
		area := r.gs.TerrainAreas[i]
		if !area.Contains(cellX, cellY) {
			continue
		}
		for candidate, candidateRegion := range r.gs.Regions {
			if candidateRegion != nil && candidateRegion.IsTerrainArea && candidateRegion.TerrainAreaID == area.ID {
				return candidate, true
			}
		}
	}
	return "", false
}

func (r *Renderer) terrainAreaRuntimeRegionID(areaID string) world.RegionID {
	if r == nil || r.gs == nil {
		return ""
	}
	for rid, region := range r.gs.Regions {
		if region != nil && region.IsTerrainArea && region.TerrainAreaID == areaID {
			return rid
		}
	}
	return ""
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
	if r.terrainAreaEditPending() {
		if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(fx, fy) {
			r.cancelTerrainAreaEdit()
			return InputAction{}, true
		}
		if buildEditInspectorTabButton(editInspectorSettlement, "Yerleşim Birimi").HitTest(fx, fy) ||
			buildEditInspectorTabButton(editInspectorRegion, "Bölge").HitTest(fx, fy) ||
			buildEditInspectorTabButton(editInspectorFaction, "Devlet").HitTest(fx, fy) ||
			buildEditInspectorTabButton(editInspectorMap, "Harita").HitTest(fx, fy) ||
			buildEditInspectorTabButton(editInspectorTerrainArea, "Arazi").HitTest(fx, fy) ||
			buildEditInspectorTabButton(editInspectorData, "Veri").HitTest(fx, fy) {
			return InputAction{}, true
		}
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
	if buildEditInspectorTabButton(editInspectorTerrainArea, "Arazi").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorTerrainArea
		return InputAction{}, true
	}
	if buildEditInspectorTabButton(editInspectorData, "Veri").HitTest(fx, fy) {
		r.editInspectorTab = editInspectorData
		return InputAction{}, true
	}
	if buildEditInspectorActionButton(editButtonSaveScenario, "").HitTest(fx, fy) {
		return InputAction{Kind: ActionSaveScenario}, true
	}
	if r.editInspectorTab == editInspectorTerrainArea {
		return r.handleEditShapeInspectorClick(fx, fy)
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
		case editButtonSettlementID:
			if r.hasEditSelection() {
				r.beginEditRename(editTextSettlementID)
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
	switch r.editRegionInspectorButtonAt(fx, fy) {
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
	case editButtonEditRegionData:
		r.openEditRegionForm()
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
	r.editOwnerDropdown.Close()
	r.editSuccessorDropdown.Close()
	r.editSettlementTypeDropdown.Close()
	r.editUnitTypeDropdown.Close()

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
	case editTextSettlementID:
		if !r.hasEditSelection() {
			return
		}
	case editTextRegionNameTR:
	case editTextRegionName:
		if region.IsTerrainArea {
			return
		}
	case editTextRegionID:
	default:
		return
	}
	r.editTextTarget = target
	r.editTextError = ""
	r.editTextRunes = r.editTextRunes[:0]
	if target == editTextRegionNameTR && region.IsTerrainArea {
		for _, area := range r.gs.TerrainAreas {
			if area.ID == region.TerrainAreaID {
				r.editTextRunes = append(r.editTextRunes, []rune(area.Name)...)
				break
			}
		}
	}
	if target == editTextRegionID {
		r.editTextRunes = append(r.editTextRunes, []rune(string(region.ID))...)
	}
	if target == editTextSettlementID && r.hasEditSelection() {
		r.editTextRunes = append(r.editTextRunes, []rune(region.Settlements[r.editSelectedSettlement].ID)...)
	}
	r.editRenaming = true
	r.editDraggingSettlement = false
}

func (r *Renderer) beginNewShapeCreation() {
	region := r.selectedRegionForShapeTools()
	if region == nil || !region.IsSea || r.editShapePaintPending {
		return
	}
	r.editRenaming = false
	r.editNewShapeModal.show = true
	r.editNewShapeID = ""
	r.editNewShapeRegion = region.ID
	r.editTextTarget = editTextShapeID
	r.editTextError = ""
	r.editTextRunes = r.editTextRunes[:0]
	r.editDraggingSettlement = false
}

func (r *Renderer) handleEditRenameInput() InputAction {
	if r.keyJustPressed(ebiten.KeyEscape) {
		r.editRenaming = false
		r.editTextTarget = editTextNone
		r.editTextError = ""
		r.editNewShapeID = ""
		r.editNewShapeRegion = ""
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyEnter) {
		r.commitEditRename()
		return InputAction{}
	}
	if r.keyJustPressed(ebiten.KeyBackspace) && len(r.editTextRunes) > 0 {
		r.editTextRunes = r.editTextRunes[:len(r.editTextRunes)-1]
	}
	if (r.editTextTarget == editTextRegionID || r.editTextTarget == editTextSettlementID || r.editTextTarget == editTextShapeID) && r.keyJustPressed(ebiten.KeyA) && editCtrlPressed() {
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
	if r.editTextTarget == editTextShapeID || r.editTextTarget == editTextShapeName {
		r.commitNewShapeInput()
		return
	}
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
	case editTextSettlementID:
		if !r.hasEditSelection() {
			break
		}
		idx := r.editSelectedSettlement
		oldID := region.Settlements[idx].ID
		newID := normalizeEditID(newName)
		if newID == "" {
			r.editTextError = "Yerleşim ID boş olamaz."
			return
		}
		if strings.IndexFunc(newID, unicode.IsSpace) >= 0 {
			r.editTextError = "Yerleşim ID boşluk içermemeli."
			return
		}
		if newID != oldID && r.settlementIDInUse(newID, rid, idx) {
			r.editTextError = "Bu yerleşim ID zaten var."
			return
		}
		if newID != oldID {
			r.renameSettlementID(rid, idx, oldID, newID)
			r.editDirty = true
		}
		r.editRenaming = false
		r.editTextTarget = editTextNone
		r.editTextError = ""
		return
	case editTextRegionID:
		newID := world.RegionID(normalizeEditID(newName))
		if newID == "" {
			r.editTextError = "ID bos olamaz."
			return
		}
		if strings.IndexFunc(string(newID), unicode.IsSpace) >= 0 {
			r.editTextError = "ID bosluk icermemeli."
			return
		}
		if newID != rid && r.gs.Regions[newID] != nil {
			r.editTextError = "Bu region ID zaten var."
			return
		}
		if newID != rid {
			r.renameRegionID(rid, newID)
			r.editDirty = true
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
			r.editDirty = true
		}
	case editTextRegionNameTR:
		if region.IsTerrainArea {
			oldName := ""
			for _, area := range r.gs.TerrainAreas {
				if area.ID == region.TerrainAreaID {
					oldName = area.Name
					break
				}
			}
			if newName != "" && oldName != newName {
				r.setTerrainAreaName(region.TerrainAreaID, newName)
				r.editDirty = true
			}
			break
		}
		oldName := region.NameTR
		if newName != "" && oldName != newName {
			region.NameTR = newName
			r.editDirty = true
		}
	case editTextRegionName:
		oldName := region.Name
		if newName != "" && oldName != newName {
			region.Name = newName
			r.editDirty = true
		}
	}
	r.editRenaming = false
	r.editTextTarget = editTextNone
	r.editTextError = ""
}

func (r *Renderer) commitNewShapeInput() {
	source := r.gs.Regions[r.editNewShapeRegion]
	if source == nil || !source.IsSea {
		r.editTextError = "Yeni Kara Sınırı için deniz bölgesi seçilmeli."
		return
	}
	value := strings.TrimSpace(string(r.editTextRunes))
	if r.editTextTarget == editTextShapeID {
		value = normalizeEditID(value)
		if value == "" {
			r.editTextError = "Shape ID boş olamaz."
			return
		}
		if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
			r.editTextError = "Shape ID boşluk içermemeli."
			return
		}
		if _, exists := r.gs.ShapeData.Shapes[value]; exists {
			r.editTextError = "Bu Shape ID zaten var."
			return
		}
		r.editNewShapeID = value
		r.editTextTarget = editTextShapeName
		r.editTextError = ""
		r.editTextRunes = r.editTextRunes[:0]
		return
	}
	if value == "" {
		r.editTextError = "Shape adı boş olamaz."
		return
	}
	shapeID := r.editNewShapeID
	if shapeID == "" {
		r.editTextError = "Önce Shape ID girilmeli."
		return
	}

	rings := r.initialRingsForNewShape(source, shapeID)
	if r.gs.ShapeData.Shapes == nil {
		r.gs.ShapeData.Shapes = make(map[string][][][2]float32)
	}
	if r.gs.ShapeData.Names == nil {
		r.gs.ShapeData.Names = make(map[string]string)
	}
	r.gs.ShapeData.Shapes[shapeID] = cloneFloatRings(rings)
	r.gs.ShapeData.Names[shapeID] = value
	newRegionID := nextRegionID(r.gs)
	newRegionX, newRegionY := source.WorldX, source.WorldY
	if r.editSelectedWorldPointSet {
		newRegionX, newRegionY = scenarioCoordsFromWorld(
			float64(r.editSelectedWorldX)+0.5,
			float64(r.editSelectedWorldY)+0.5,
		)
	}
	newRegion := &world.Region{
		ID:           newRegionID,
		Name:         value,
		NameTR:       value,
		Terrain:      world.TerrainPlain,
		WorldX:       newRegionX,
		WorldY:       newRegionY,
		ShapeID:      shapeID,
		Shape:        cloneFloatRings(rings),
		IsSea:        false,
		Satisfaction: 70,
		TaxRate:      45,
	}
	r.gs.Regions[newRegionID] = newRegion
	r.insertRegionOrderAfter(source.ID, newRegionID)
	recalculateCountryShapeBounds(&r.gs.ShapeData)
	r.editSelectedRegion = newRegionID
	r.editSelectedSettlement = -1
	r.editShapeSession = nil
	complete := func() {
		visual := r.worldMap.VisualNeighbors(newRegionID, r.editVisualNeighborBuf[:0])
		r.applyVisualNeighbors(newRegionID, visual)
		r.editDirty = true
		r.closeEditNewShapeModal()
	}
	if !r.requestEditWorldMapRebuildWithCompletion(complete) {
		r.rebuildEditWorldMap()
		complete()
	}
}

func (r *Renderer) initialRingsForNewShape(region *world.Region, shapeID string) [][][2]float32 {
	if region == nil {
		return nil
	}
	if region.IsSea {
		return r.initialSeaShapeRings(shapeID)
	}
	if len(region.Shape) > 0 {
		return cloneFloatRings(region.Shape)
	}
	if region.ShapeID != "" {
		if rings := r.gs.ShapeData.Shapes[region.ShapeID]; len(rings) > 0 {
			return cloneFloatRings(rings)
		}
	}
	if r.worldMap == nil {
		return nil
	}
	session := newBlankShapeEditSession(r.gs, shapeID)
	if session == nil {
		return nil
	}
	for _, pIdx := range r.worldMap.regionPx[region.ID] {
		x, y := pIdx%WorldW, pIdx/WorldW
		if session.inBounds(x, y) {
			session.Mask[session.index(x, y)] = 1
		}
	}
	return shapeMaskToFloatRings(session)
}

func (r *Renderer) initialSeaShapeRings(shapeID string) [][][2]float32 {
	if r == nil || r.gs == nil || r.worldMap == nil {
		return nil
	}
	session := newBlankShapeEditSession(r.gs, shapeID)
	if session == nil {
		return nil
	}
	cx, cy := r.editSelectedWorldX, r.editSelectedWorldY
	if !r.editSelectedWorldPointSet {
		region := r.gs.Regions[r.editSelectedRegion]
		if region == nil {
			return nil
		}
		cx = shapeRasterWorldPixelX(region.WorldX)
		cy = shapeRasterWorldPixelY(region.WorldY)
	}
	const radius = 3
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy > radius*radius || !session.inBounds(x, y) {
				continue
			}
			session.Mask[session.index(x, y)] = 1
		}
	}
	return shapeMaskToFloatRings(session)
}

func (r *Renderer) editTextLabel() string {
	switch r.editTextTarget {
	case editTextSettlementNameTR:
		return "Yerleşim Adı"
	case editTextSettlementID:
		return "Yerleşim ID"
	case editTextRegionNameTR:
		if region := r.gs.Regions[r.editSelectedRegion]; region != nil && region.IsTerrainArea {
			return "Arazi Adı"
		}
		return "Bolge Ad TR"
	case editTextRegionName:
		return "Bolge Ad EN"
	case editTextRegionID:
		return "Bolge ID"
	case editTextShapeID:
		return "Yeni Kara Sınırı ID"
	case editTextShapeName:
		return "Yeni Kara Sınırı Adı"
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

func (r *Renderer) rememberEditSelectedWorldPoint(rid world.RegionID, fx, fy float64) {
	r.editSelectedWorldPointSet = false
	region := r.gs.Regions[rid]
	if region == nil || !region.IsSea {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	r.editSelectedWorldX, r.editSelectedWorldY = shapePaintCellFromWorld(wx, wy)
	r.editSelectedWorldPointSet = true
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

func (r *Renderer) finishRegionCenterDrag() bool {
	start := r.editRegionDragStart
	r.editRegionDragStart = nil
	if start == nil {
		return false
	}
	region := r.gs.Regions[start.Region]
	if region == nil || region.IsTerrainArea || (region.WorldX == start.X && region.WorldY == start.Y) {
		return false
	}
	return true
}

func (r *Renderer) regionCenterAffectsRaster(region *world.Region) bool {
	if region == nil {
		return false
	}
	if region.IsSea || region.ShapeID == "" {
		return region.IsSea
	}
	count := 0
	for _, candidate := range r.gs.Regions {
		if candidate != nil && !candidate.IsSea && !candidate.IsTerrainArea && candidate.ShapeID == region.ShapeID {
			count++
			if count > 1 {
				return true
			}
		}
	}
	return false
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
	if r.regionCenterAffectsRaster(region) {
		if region.IsSea {
			r.requestEditWorldMapRebuild()
		} else {
			r.requestEditWorldMapRebuildForShape(region.ShapeID)
		}
	} else {
		r.invalidateEditRegionCenterMarkers()
	}
}

func (r *Renderer) moveSelectedSettlementTo(fx, fy float64) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || r.editSelectedSettlement < 0 ||
		r.editSelectedSettlement >= len(region.Settlements) {
		return
	}
	wx, wy := r.screenToWorld(fx, fy)
	if r.settlementPointBlockedByTerrain(wx, wy) {
		return
	}
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
	r.invalidateEditRegionCenterMarkers()
	r.editDirty = true
}

func (r *Renderer) addSettlementAt(fx, fy float64) {
	wx, wy := r.screenToWorld(fx, fy)
	if r.settlementPointBlockedByTerrain(wx, wy) {
		return
	}
	rid := r.worldMap.RegionAt(int(wx), int(wy))
	x, y := scenarioCoordsFromWorld(wx, wy)
	r.addSettlement(rid, x, y)
}

func (r *Renderer) addSettlementToSelectedRegion() {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || !canAddSettlementToRegion(region) {
		return
	}
	if r.settlementPointBlockedByTerrain(wcX(region.WorldX), wcY(region.WorldY)) {
		return
	}
	r.addSettlement(region.ID, region.WorldX, region.WorldY)
}

func (r *Renderer) addSettlement(rid world.RegionID, x, y int) {
	region, ok := r.gs.Regions[rid]
	if !ok || !canAddSettlementToRegion(region) {
		return
	}
	if r.settlementPointBlockedByTerrain(wcX(x), wcY(y)) {
		return
	}
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
}

func (r *Renderer) deleteSelectedSettlement() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
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
	complete := func() {
		visual := r.worldMap.VisualNeighbors(rid, r.editVisualNeighborBuf[:0])
		r.applyVisualNeighbors(rid, visual)
		r.editDirty = true
	}
	if !r.requestEditWorldMapRebuildWithCompletion(complete) {
		r.rebuildEditWorldMap()
		complete()
	}
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
	r.requestEditWorldMapRebuild()
	r.editDirty = true
}

func (r *Renderer) setSelectedSettlementCapital() {
	if !r.hasEditSelection() {
		return
	}
	region := r.gs.Regions[r.editSelectedRegion]
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
	if f.CapitalSettlementID == settlement.ID && f.PendingCapitalSettlementID == "" {
		successorChanged := setRegionSuccessorToOwner(region)
		infrastructureChanged := world.EnsureRequiredSettlementBuildings(region, true)
		if !successorChanged && !infrastructureChanged {
			return
		}
		r.editDirty = true
		return
	}

	if !r.gs.SetFactionCapital(fid, settlement.ID) {
		return
	}
	setRegionSuccessorToOwner(region)
	world.EnsureRequiredSettlementBuildings(region, true)
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionTerrain(terrain world.TerrainType) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil {
		return
	}
	if region.Terrain == terrain && (!region.IsTerrainArea || !r.terrainAreaHasTouched()) {
		return
	}
	if region.IsTerrainArea {
		if r.terrainAreaEditPending() {
			for i := range r.gs.TerrainAreas {
				if r.gs.TerrainAreas[i].ID == region.TerrainAreaID || r.terrainAreaWasTouched(i) {
					r.gs.TerrainAreas[i].Terrain = terrain
				}
			}
			r.syncTerrainAreaChildValues()
			return
		}
		for i := range r.gs.TerrainAreas {
			if r.gs.TerrainAreas[i].ID != region.TerrainAreaID && !r.terrainAreaWasTouched(i) {
				continue
			}
			for j := range r.gs.TerrainAreas {
				if r.gs.TerrainAreas[j].ID == region.TerrainAreaID || r.terrainAreaWasTouched(j) {
					r.gs.TerrainAreas[j].Terrain = terrain
				}
			}
			r.requestEditWorldMapRebuild()
			r.editDirty = true
			return
		}
	}
	region.Terrain = terrain
	r.editDirty = true
}

func (r *Renderer) terrainAreaWasTouched(index int) bool {
	if r == nil || index < 0 || index >= len(r.gs.TerrainAreas) || r.editTerrainAreaTouchedIDs == nil {
		return false
	}
	_, ok := r.editTerrainAreaTouchedIDs[r.gs.TerrainAreas[index].ID]
	return ok
}

func (r *Renderer) syncTerrainAreaChildValues() {
	if r == nil || r.gs == nil {
		return
	}
	byID := make(map[string]*world.TerrainArea, len(r.gs.TerrainAreas))
	for i := range r.gs.TerrainAreas {
		byID[r.gs.TerrainAreas[i].ID] = &r.gs.TerrainAreas[i]
	}
	for _, region := range r.gs.Regions {
		if region == nil || !region.IsTerrainArea {
			continue
		}
		area := byID[region.TerrainAreaID]
		if area == nil {
			continue
		}
		region.Terrain = area.Terrain
		region.IsLocked = !world.TerrainData[area.Terrain].Passable
	}
}

func (r *Renderer) terrainAreaHasTouched() bool {
	if r == nil || r.editTerrainAreaTouchedIDs == nil {
		return false
	}
	return len(r.editTerrainAreaTouchedIDs) > 0
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
	settlement.Type = st
	world.EnsureRequiredSettlementBuildings(region, r.gs.IsCapitalRegion(region))
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
	region.OwnerID = ownerID
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.worldMap.MarkDirty()
	r.editDirty = true
}

func (r *Renderer) setSelectedRegionSuccessor(successorID string) {
	region, ok := r.gs.Regions[r.editSelectedRegion]
	if !ok || region == nil || region.IsSea || region.SuccessorFactionID == successorID {
		return
	}
	region.SuccessorFactionID = successorID
	r.editDirty = true
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

func (r *Renderer) settlementIDInUse(id string, exceptRegion world.RegionID, exceptIndex int) bool {
	if r == nil || r.gs == nil || id == "" {
		return false
	}
	for rid, region := range r.gs.Regions {
		if region == nil {
			continue
		}
		for index, settlement := range region.Settlements {
			if rid == exceptRegion && index == exceptIndex {
				continue
			}
			if settlement.ID == id {
				return true
			}
		}
	}
	return false
}

func (r *Renderer) renameSettlementID(rid world.RegionID, index int, oldID, newID string) {
	if r == nil || r.gs == nil || oldID == "" || newID == "" || oldID == newID {
		return
	}
	region := r.gs.Regions[rid]
	if region == nil || index < 0 || index >= len(region.Settlements) || region.Settlements[index].ID != oldID {
		return
	}
	region.Settlements[index].ID = newID

	for _, f := range r.gs.Factions {
		if f == nil {
			continue
		}
		if f.CapitalSettlementID == oldID {
			f.CapitalSettlementID = newID
		}
		if f.PendingCapitalSettlementID == oldID {
			f.PendingCapitalSettlementID = newID
		}
	}
	for _, a := range r.gs.Armies {
		if a != nil && a.DockedSettlementID == oldID {
			a.DockedSettlementID = newID
		}
	}
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

func (r *Renderer) setTerrainAreaName(areaID, name string) {
	if r == nil || r.gs == nil {
		return
	}
	for i := range r.gs.TerrainAreas {
		if r.gs.TerrainAreas[i].ID != areaID {
			continue
		}
		r.gs.TerrainAreas[i].Name = name
		for _, region := range r.gs.Regions {
			if region != nil && region.IsTerrainArea && region.TerrainAreaID == areaID {
				region.Name = name
				region.NameTR = name
			}
		}
		r.editDirty = true
		r.editSelectedSettlement = -1
		return
	}
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
			if center.Links[j].RegionID == oldID {
				center.Links[j].RegionID = newID
			}
		}
	}
	r.gs.InvalidateHistoricalTradeFlowCache()
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
	if r.editVoronoiDebugRegion == oldID {
		r.editVoronoiDebugRegion = newID
	}
	if r.editBoundaryRegion == oldID {
		r.editBoundaryRegion = newID
	}
	r.invalidateEditVisualNeighborCache()
	r.editVoronoiDebugWorldMap = nil
	r.editBoundaryWorldMap = nil
	r.invalidateEditRegionCenterMarkers()
	if r.worldMap != nil {
		r.worldMap.renameRegionID(oldID, newID)
	}
}

func (r *Renderer) toggleSelectedRegionLock() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	region.IsLocked = !region.IsLocked
	r.editDirty = true
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
	region.UnlockTurn = next
	r.editDirty = true
}

func (r *Renderer) syncSelectedRegionNeighborsFromVisual() {
	region := r.gs.Regions[r.editSelectedRegion]
	if region == nil {
		return
	}
	if region.IsTerrainArea {
		visual := r.terrainAreaVisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
		changed := false
		for _, neighborID := range visual {
			neighbor := r.gs.Regions[neighborID]
			if neighbor == nil || neighbor.IsSea || neighbor.ID == region.ID {
				continue
			}
			if r.appendTerrainAreaExtraNeighbor(region.TerrainAreaID, neighbor.ID) {
				changed = true
			}
		}
		if !changed {
			return
		}
		r.requestEditWorldMapRebuild()
		r.editDirty = true
		return
	}
	visual := r.worldMap.VisualNeighbors(region.ID, r.editVisualNeighborBuf[:0])
	before := append([]world.RegionID(nil), region.Neighbors...)
	r.applyVisualNeighbors(region.ID, visual)
	if regionIDSlicesEqual(before, region.Neighbors) {
		return
	}
	r.editDirty = true
}

// terrainAreaVisualNeighbors, arazi proxy'sinin tek bir raster düğümüne bağlı
// kalmadan poligonun kapladığı gerçek bölgeleri bulur. Bu hesap yalnız Komşu
// Sync tıklamasında çalışır; normal çizim döngüsüne girmez.
func (r *Renderer) terrainAreaVisualNeighbors(regionID world.RegionID, dst []world.RegionID) []world.RegionID {
	dst = dst[:0]
	if r == nil || r.worldMap == nil || r.gs == nil {
		return dst
	}
	region := r.gs.Regions[regionID]
	if region == nil || !region.IsTerrainArea {
		return dst
	}
	return r.worldMap.VisualNeighbors(region.ID, dst)
}

func (r *Renderer) worldSnapshot() editWorldSnapshot {
	return editWorldSnapshot{
		Regions:              cloneRegionMap(r.gs.Regions),
		RegionOrder:          cloneRegionIDSlice(r.gs.RegionOrder),
		LandPassages:         cloneLandPassages(r.gs.LandPassages),
		Factions:             cloneFactionMap(r.gs.Factions),
		AIStrategies:         cloneAIStrategyMap(r.gs.AIStrategies),
		AIStrategyOrder:      append([]string(nil), r.gs.AIStrategyOrder...),
		TradeCenters:         cloneTradeCenterConfig(r.gs.TradeCenters),
		Armies:               cloneArmyMap(r.gs.Armies),
		ArmyOrder:            append([]army.ArmyID(nil), r.gs.ArmyOrder...),
		Relations:            cloneRelationMap(r.gs.Relations),
		RelationOrder:        append([]string(nil), r.gs.RelationOrder...),
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

func (r *Renderer) restoreWorldSnapshotSync(snapshot editWorldSnapshot) {
	r.restoreWorldSnapshotMode(snapshot, false, true)
}

func (r *Renderer) restoreWorldSnapshotDataOnly(snapshot editWorldSnapshot) {
	r.restoreWorldSnapshotMode(snapshot, false, false)
}

func (r *Renderer) restoreWorldSnapshotMode(snapshot editWorldSnapshot, asyncBuild, rebuildMap bool) {
	r.gs.Regions = cloneRegionMap(snapshot.Regions)
	r.gs.RegionOrder = cloneRegionIDSlice(snapshot.RegionOrder)
	r.gs.LandPassages = cloneLandPassages(snapshot.LandPassages)
	r.gs.TerrainAreas = cloneTerrainAreas(snapshot.TerrainAreas)
	r.gs.Factions = cloneFactionMap(snapshot.Factions)
	r.gs.AIStrategies = cloneAIStrategyMap(snapshot.AIStrategies)
	r.gs.AIStrategyOrder = append([]string(nil), snapshot.AIStrategyOrder...)
	r.gs.TradeCenters = cloneTradeCenterConfig(snapshot.TradeCenters)
	r.gs.InvalidateHistoricalTradeFlowCache()
	r.gs.Armies = cloneArmyMap(snapshot.Armies)
	r.gs.ArmyOrder = append([]army.ArmyID(nil), snapshot.ArmyOrder...)
	r.gs.Relations = cloneRelationMap(snapshot.Relations)
	r.gs.RelationOrder = append([]string(nil), snapshot.RelationOrder...)
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
	r.editShapeStrokeLandShapeIDs = nil
	r.editShapePendingLandShapeIDs = nil
	r.editRenaming = false
	if !rebuildMap {
		return
	}
	if asyncBuild {
		r.requestEditWorldMapRebuild()
	} else {
		r.rebuildEditWorldMap()
	}
}

func cloneRegionMap(src map[world.RegionID]*world.Region) map[world.RegionID]*world.Region {
	dst := make(map[world.RegionID]*world.Region, len(src))
	for rid, region := range src {
		if region == nil {
			continue
		}
		copyRegion := *region
		copyRegion.Neighbors = cloneRegionIDSlice(region.Neighbors)
		copyRegion.AreaNeighborOrder = cloneRegionIDSlice(region.AreaNeighborOrder)
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
		dst[i].Polygons = make([][][2]int, len(area.Polygons))
		for j := range area.Polygons {
			dst[i].Polygons[j] = append([][2]int(nil), area.Polygons[j]...)
		}
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
	if src.Centers == nil && src.Sources == nil {
		return world.TradeCenterConfig{}
	}
	dst := world.TradeCenterConfig{
		PrimaryTradeCapacityBonus:      src.PrimaryTradeCapacityBonus,
		SecondaryTradeCapacityBonus:    src.SecondaryTradeCapacityBonus,
		PrimaryTradeIncomeBonus:        src.PrimaryTradeIncomeBonus,
		SecondaryTradeIncomeBonus:      src.SecondaryTradeIncomeBonus,
		PrimaryMerchantCapacityBonus:   src.PrimaryMerchantCapacityBonus,
		SecondaryMerchantCapacityBonus: src.SecondaryMerchantCapacityBonus,
		PrimaryMerchantIncomeBonus:     src.PrimaryMerchantIncomeBonus,
		SecondaryMerchantIncomeBonus:   src.SecondaryMerchantIncomeBonus,
		HistoricalFlows:                append([]world.HistoricalTradeFlow(nil), src.HistoricalFlows...),
		Sources:                        make([]world.TradeCenterDef, len(src.Sources)),
		Centers:                        make([]world.TradeCenterDef, len(src.Centers)),
	}
	for i, source := range src.Sources {
		dst.Sources[i] = source
		dst.Sources[i].Links = append([]world.TradeCenterLink(nil), source.Links...)
		dst.Sources[i].CompetitionImpacts = append([]world.TradeCompetitionImpact(nil), source.CompetitionImpacts...)
		dst.Sources[i].SourceGoods = append([]world.HistoricalTradeGood(nil), source.SourceGoods...)
	}
	for i, center := range src.Centers {
		dst.Centers[i] = center
		dst.Centers[i].Links = append([]world.TradeCenterLink(nil), center.Links...)
		dst.Centers[i].CompetitionImpacts = append([]world.TradeCompetitionImpact(nil), center.CompetitionImpacts...)
		dst.Centers[i].SourceGoods = append([]world.HistoricalTradeGood(nil), center.SourceGoods...)
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
	r.invalidateEditVisualNeighborCache()
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
		id:         normalizeEditID(string(f.ID)),
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
		overlordID: f.OverlordID,
	}
	r.setFactionFormRelationTarget(firstRelationTarget(r.gs, f.ID))
}

func (r *Renderer) saveFactionForm() bool {
	form := &r.editFactionForm
	fid := faction.FactionID(normalizeEditID(form.id))
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
		for _, other := range r.gs.Factions {
			if other != nil && other.OverlordID == form.originalID {
				other.OverlordID = fid
			}
		}
		if r.gs.PlayerFactionID == form.originalID {
			r.gs.PlayerFactionID = fid
		}
	}
	next := &faction.Faction{}
	if existingFaction != nil {
		// Düzenlemede formun göstermediği alanlar (ör. tarihsel değişimler,
		// gelir dönemleri, bayrak ve stratejik hedefler) aynen korunmalıdır.
		preserved := *existingFaction
		next = &preserved
	}
	next.ID = fid
	next.Name = name
	next.NameTR = nameTR
	next.Religion = form.religion
	next.Color = form.color
	next.IsPlayable = form.playable
	next.Gold = gold
	next.Grain = grain
	next.Iron = iron
	next.Timber = timber
	next.Spice = spice
	next.Cloth = cloth
	next.AIAggressiveness = aiValue
	if !r.validEditFactionOverlord(fid, form.overlordID) {
		form.errorText = "Geçersiz vassal üst devleti seçimi."
		return false
	}
	next.OverlordID = form.overlordID
	if next.OverlordID != "" {
		next.TributeRate = diplomacy.VassalTributeRatePercent()
		next.TributeRateConfigured = true
		next.VassalizedTurn = r.gs.Turn
		if existingFaction != nil && existingFaction.OverlordID == next.OverlordID && existingFaction.VassalizedTurn > 0 {
			next.TributeRate = existingFaction.TributeRate
			next.TributeRateConfigured = existingFaction.TributeRateConfigured
			next.VassalizedTurn = existingFaction.VassalizedTurn
		}
		if !next.TributeRateConfigured {
			next.TributeRate = diplomacy.VassalTributeRatePercent()
			next.TributeRateConfigured = true
		}
	} else {
		next.TributeRate = 0
		next.TributeRateConfigured = false
		next.VassalizedTurn = 0
	}
	r.gs.Factions[fid] = next
	r.ensureRelationsForFaction(fid)
	if form.relationTarget != "" && r.gs.Factions[form.relationTarget] != nil && form.relationTarget != fid {
		r.setRelationValue(fid, form.relationTarget, relationScore, form.relationStance)
	}
	r.editSelectedFaction = fid
	r.worldMap.MarkDirty()
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
	case buildEditFactionFormButton(editFactionFormOverlordTarget, "").HitTest(fx, fy):
		r.cycleFactionFormOverlordTarget()
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

func (r *Renderer) cycleFactionFormOverlordTarget() {
	self := faction.FactionID(normalizeEditID(r.editFactionForm.id))
	ids := sortedFactionIDs(r.gs.Factions)
	options := []faction.FactionID{""}
	for _, fid := range ids {
		if fid != self && r.validEditFactionOverlord(self, fid) {
			options = append(options, fid)
		}
	}
	for i, fid := range options {
		if fid == r.editFactionForm.overlordID {
			r.editFactionForm.overlordID = options[(i+1)%len(options)]
			return
		}
	}
	r.editFactionForm.overlordID = ""
}

func (r *Renderer) validEditFactionOverlord(self, overlord faction.FactionID) bool {
	if overlord == "" {
		return true
	}
	overlordFaction := r.gs.Factions[overlord]
	return self != overlord && overlordFaction != nil && !overlordFaction.IsEliminated && overlordFaction.OverlordID == ""
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
		ch = unicode.ToLower(ch)
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
	self := faction.FactionID(normalizeEditID(r.editFactionForm.id))
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
	self := faction.FactionID(normalizeEditID(r.editFactionForm.id))
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
	for _, existing := range r.gs.RelationOrder {
		if existing == key {
			return
		}
	}
	r.gs.RelationOrder = append(r.gs.RelationOrder, key)
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
	if !a.IsNaval && !canPlaceEditLandArmy(r.gs, region) {
		return
	}
	a.RegionID = region.ID
	a.DockedRegionID = ""
	a.DockedSettlementID = ""
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
	r.gs.ArmyOrder = append(r.gs.ArmyOrder, aid)
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
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
	r.gs.ArmyOrder = append(r.gs.ArmyOrder, aid)
	r.SelectedArmy = aid
	r.editSelectedFaction = faction.FactionID(ownerID)
	r.editSelectedUnitType = unitTypeID
	r.editDirty = true
}

func (r *Renderer) deleteSelectedArmy() {
	a := r.gs.Armies[r.SelectedArmy]
	if a == nil {
		return
	}
	r.gs.RemoveArmy(a.ID)
	r.SelectedArmy = ""
	r.editSelectedUnitType = ""
	r.editDirty = true
}

func (r *Renderer) addSelectedArmyUnit() {
	a := r.gs.Armies[r.SelectedArmy]
	if !r.canAddSelectedArmyUnit() || a == nil {
		return
	}
	a.Units = append(a.Units, army.Unit{TypeID: r.editSelectedUnitType, CurrentHP: 100})
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
	for i := range a.Units {
		a.Units[i].TypeID = typeID
	}
	r.editSelectedUnitType = typeID
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
	for i := len(a.Units) - 1; i >= 0; i-- {
		if a.Units[i].TypeID == r.editSelectedUnitType {
			a.Units = append(a.Units[:i], a.Units[i+1:]...)
			break
		}
	}
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
	return canPlaceEditLandArmy(r.gs, region) && r.editOwnerForRegion(region) != "" && r.defaultEditUnitType(false) != ""
}

func canPlaceEditLandArmy(gs *state.GameState, region *world.Region) bool {
	if gs == nil || region == nil || region.IsSea || region.IsLocked {
		return false
	}
	_, blocked := gs.LandRegionMoveCost(region)
	return !blocked
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
	return utype != nil && utype.PrimaryBuildingID() == "port"
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
	a.OwnerID = region.OwnerID
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

func editBoolLabel(value bool) string {
	if value {
		return "evet"
	}
	return "hayir"
}

func (r *Renderer) rebuildEditWorldMap() {
	r.cancelEditMapBuild()
	r.invalidateShapeEditSession()
	r.invalidateEditRegionCenterMarkers()
	r.worldMap = NewWorldMap(r.gs)
	r.buildRegionPaintBaseline()
	overridesChanged := !regionPaintOverridesEqual(r.editRegionPaintOverrides, r.gs.RegionPaintOverrides)
	if overridesChanged {
		r.applyRegionPaintOverrides()
		// NewWorldMap önce kaynak override'larını, sonra terrain alanlarını işler.
		// Sadece edit oturumu override'ları değiştiyse terrain katmanı bir kez daha
		// uygulanmalı; aksi halde aynı terrain rasterı gereksiz yere iki kez üretilir.
		r.worldMap.applyTerrainAreaRegions(r.gs)
	}
	// Geçişlerin uçları sabit harita koordinatlarına bağlıdır; harita üzerinde
	// bölge ataması değiştiğinde From/To ilişkisini aynı rasterdan yenile.
	r.syncLandPassageRegionsFromMap()
}

// refreshRegionPaintInEditMap, ülke shape rasterını ve deniz BFS'ini yeniden
// üretmeden yalnız stroke sırasında değişen region paint pixel'lerini mevcut
// haritaya işler. Terrain alanları override'ların üst katmanı olduğu için
// sonunda tek seferde yeniden uygulanır.
func (r *Renderer) refreshRegionPaintInEditMap(dirtyPixels map[int]struct{}) {
	if r == nil || r.gs == nil || r.worldMap == nil {
		return
	}
	baseline := r.editRegionPaintBaseline
	if len(baseline) != len(r.worldMap.regionAt) {
		baseline = r.worldMap.baseRegionAt
	}

	for pIdx := range dirtyPixels {
		if pIdx < 0 || pIdx >= len(r.worldMap.regionAt) {
			continue
		}
		if target, ok := r.editRegionPaintOverrides[pIdx]; ok && target != "" {
			r.worldMap.regionAt[pIdx] = r.worldMap.ensureRegionIndex(target)
			continue
		}
		if pIdx < len(baseline) {
			r.worldMap.regionAt[pIdx] = baseline[pIdx]
		}
	}

	r.worldMap.rebuildRegionPixelsFromAssignments()
	world.SyncTerrainAreaRegions(r.gs.Regions, r.gs.TerrainAreas)
	r.worldMap.applyTerrainAreaRegions(r.gs)
	r.worldMap.rebuildBorderSegments(r.gs)
	clear(r.worldMap.regionAnchor)
	r.worldMap.computeRegionAnchors()
	clear(r.worldMap.settlementAnchor)
	clear(r.worldMap.primarySettlement)
	r.worldMap.computeSettlementAnchors(r.gs)
	r.worldMap.applyOwnership(r.gs, "", MapModeNormal)
	r.syncLandPassageRegionsFromMap()
}

// refreshTerrainAreasInEditMap güncellenen terrain polygonlarını mevcut temel
// rasterı yeniden üretmeden harita atamalarına işler. Terrain editöründe shape,
// sahiplik ve bölge override'ı değişmediği için NewWorldMap çağırmak gereksiz
// ve özellikle büyük haritalarda editörü saniyelerce kilitliyor.
func (r *Renderer) refreshTerrainAreasInEditMap() {
	if r == nil || r.gs == nil || r.worldMap == nil {
		return
	}
	// Önce eski terrain child atamalarını parent rasterına geri bırak; aksi
	// halde aynı WorldMap üzerinde yeni alanları uygularken eski child indeksleri
	// parent kontrolünü engeller.
	for rid, pixels := range r.worldMap.regionPx {
		region := r.gs.Regions[rid]
		if region == nil || !region.IsTerrainArea {
			continue
		}
		for _, pixel := range pixels {
			if pixel >= 0 && pixel < len(r.worldMap.regionAt) {
				if len(r.worldMap.baseRegionAt) == len(r.worldMap.regionAt) {
					r.worldMap.regionAt[pixel] = r.worldMap.baseRegionAt[pixel]
					continue
				}
				parentIdx := r.worldMap.regionIdx[region.ParentRegionID]
				r.worldMap.regionAt[pixel] = parentIdx
			}
		}
	}
	world.SyncTerrainAreaRegions(r.gs.Regions, r.gs.TerrainAreas)
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
	changed := false
	for pIdx, rid := range r.editRegionPaintOverrides {
		if pIdx < 0 || pIdx >= len(r.worldMap.regionAt) || rid == "" {
			continue
		}
		newIdx := r.worldMap.ensureRegionIndex(rid)
		if r.worldMap.regionAt[pIdx] == newIdx {
			continue
		}
		r.worldMap.regionAt[pIdx] = newIdx
		changed = true
	}
	if changed {
		// Her pikselde eski region pixel listesinden lineer silme yerine tek
		// geçişte ters indeksleri yeniden oluştur.
		r.worldMap.rebuildRegionPixelsFromAssignments()
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
	if r.settlementPointBlockedByTerrain(wcX(x), wcY(y)) {
		return
	}
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

func (r *Renderer) settlementPointBlockedByTerrain(wx, wy float64) bool {
	if r == nil || r.gs == nil {
		return false
	}
	return world.TerrainAreasContainPoint(r.gs.TerrainAreas, int(math.Floor(wx)), int(math.Floor(wy)))
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
