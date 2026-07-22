---
type: architecture
tags: [render, ebitengine, camera, input, ui]
last_updated: 2026-07-22
related: [game-loop, state-management, shape-editor, systems/combat, architecture/ui-framework]
---

# Render Pipeline

Üst-sol kaynak satırları tur başı değişimi ve mevcut stoku `+üretim/mevcut` biçiminde gösterir; tahıl değeri sivil ve ordu tüketimi sonrası net değişimdir, negatif değerler `-` işaretiyle ve kırmızı renkle çizilir. Devlet üretim toplamı kuşatma altındaki bölgeleri dışarıda bırakan `GameState.FactionProductionSummary()` ile hesaplanır.

Üst-sol durum HUD'u oyuncu devletinin bayrağı ve adıyla birlikte mevcut askeri gücünü (`diplomacy.MilitaryPower`) ve aktif, elenmemiş devletler arasındaki güç sırasını gösterir. Aynı standing bilgisi seçilen devlet bilgi panelindeki `Durum` bölümünde de gösterilir; sıra hesabı ortak `factionMilitaryPowerStanding` helper'ından gelir. Sıralama eşit güçte faction ID'siyle deterministik olarak çözülür (`internal/render/panel.go`). Kaynak HUD'unda tahıl miktarının altında ayrıca `Ambar` satırı bulunur; kapasite ekonomi tick'i status'undan, status henüz oluşmamışsa `GameState.GrainStorageCapacityForFaction()` hesaplamasından alınır.

Devlet bilgi paneli açıkken yeni bir bölge seçilirse panel açık tutulur ve `SelectedRegion` bölgesinin `OwnerID` değerindeki devlete senkronlanır. Aynı devletin bölgeleri arasında geçiş panel scroll'unu korur; farklı devlete geçiş panel içeriğini baştan başlatır (`internal/render/renderer_input.go`).

Bölge seçimi askeri birim üretim panelini açmaz; seçim sırasında açık recruit paneli kapanır ve yalnızca alt HUD'daki `Ordu` butonuyla açılır. Deniz ve kara bölgesi seçimleri bu ortak input kuralını kullanır (`internal/render/renderer_input.go`).

Bina kartlarının inşa edilebilirlik ve hover davranışı yalnızca oyuncunun sahibi olduğu, kilitli olmayan kara bölgelerinde aktiftir. Oyuncuya ait olmayan bölgelerde kartlar görünmeye devam eder ancak kaynak yeterli olsa bile bina etiketi altın sarısı kalır ve bina detay popup'ı açılmaz; bina tıklama hit-test'i de aynı sahiplik kontrolünü kullanır (`internal/render/building_card_component.go`, `panel.go`, `hover_tooltip.go`).

Kuşatılan bölgedeki bölge sahibi veya müttefik kara ordusu komşu dost/sahipsiz hedefe sağ tıkladığında savaş planı kuşatan orduya karşı açılır; preview kaynak bölgenin arazisini, başarılı hareket ise seçilen hedefi kullanır. Ordu paneli ve birim kartları bu durumda `Takviye aktif` göstergesi çizmez.

Bölgede aktif kuşatma yürüten oyuncu ordusu seçildiğinde kuşatmaya ait hedef, durum, tahkimat, ilerleme, gedik, teslim süresi ve hücum uygunluğu bilgileri yalnızca seçili `Kuşatma Emri` panelinde gösterilir. Saldıran görünümde `Genel Hücum`, `Kuşatmayı Kaldır` ve `Teslimiyet Teklifi` düğmeleri üçlü, çakışmayan geometriyle çizilir; son düğme oyuncunun AI savunmacıya kuşatma teslimiyeti çağrısı göndermesini sağlar. Aynı bölge için teklif reddedilmişse `Teslimiyet Teklifi` o tur pasif çizilir ve hit-test'e girmez; diğer bölgelerin düğmeleri etkilenmez. Kuşatılan oyuncu ordusu veya yerleşimi seçildiğinde aynı panel savunma görünümüne geçer; kuşatan gücü, savunma gücü ve kuşatma durumu gösterilir. Bu görünümden `Huruç başlat` çalışır; `Teslim ol` yalnız aktif kuşatmaya ait AI teslimiyet teklifi varsa etkinleşir, teklif yokken düğme ve hit-test pasiftir. Oyuncuya gelen teklif modalı genişletilmiş dikey içerik alanında mesajı kırpmadan gösterir, kabul aksiyonu her teklif türünde `Kabul Et` olarak kalır ve bölge bağlı kuşatma teklifinde kamera `RegionID` ile arka plandaki kuşatma alanına odaklanır. Panel alt ordu detay paneliyle üst üste binmez (`internal/render/{renderer_dialogs.go,renderer_input.go}`, `army_panel.go`). Ticaret panelindeki `Pazar` sekmesi normal partnerli al/sat akışının yanında, tahıl seçiliyken veya partner bulunmadığında depo kapasitesi üstü tahılı doğrudan satmak için `ACİL TAHIL SAT` butonunu aynı action-card geometri/hit-test sözleşmesinde gösterir (`internal/render/trade.go`, `renderer_input.go`).

Birim görselleri artık tek bir sheet'ten değil, birim türü başına ayrı PNG'den yüklenir. `militia`/`infantry`/`elite_infantry`, süvari, kuşatma ve gemi türleri `sprites/eastern_army/` ile `sprites/western_army/` altındaki kanonik dosya eşlemesine bağlanır. Sünni/Şii fraksiyonlar doğu setini, Katolik/Ortodoks ve diğer tanımlı dinler batı setini kullanır; recruit paneli ve tooltip oyuncu fraksiyonunu, ordu detay paneli ise ilgili ordunun sahibini baz alır. Faksiyon bulunamıyorsa eski senaryolar için `sprites/army.png` fallback'i korunur. Ordu ve recruit kartlarında 10×2 slot düzeni korunurken kart yüksekliği 210×360 görsel oranına göre hesaplanır. Sprite'ın tamamı üstten başlayarak kırpılmadan çizilir; ordu/kuyruk kartlarının footer'ı opak beyaz kalırken recruit üretim kartlarının alt etiket alanı üretilebilirlik durumuna göre yeşil, amber, mavi veya kırmızı renklendirilir. Bu footer'lardaki isim ve tur metinleri siyah çizilir; isim, HP/progress etiketi ile kuyruk iptal butonu sprite'ın üzerine bindirilir (`internal/render/recruit_panel.go`, `army_panel.go`, `hover_tooltip.go`).

Birim detay hover popup'larındaki görsel, aynı 210×360 oranı korunarak mevcut yüksekliğine 50 px eklenmiş ölçüyle çizilir. Recruit ve ordu popup'larının metin alanı daralmasın diye popup genişliği bu yeni görsel kolonuna göre 358 px'e çıkarılmıştır (`internal/render/hover_tooltip.go`).

Ordu detay panelindeki birim kartları oyun state'indeki fiziksel birim sırasını değiştirmeden kategoriye göre gruplanır; aynı kategori içindeki sıra korunur ve görünüm önceliği piyade, süvari, kuşatma, ardından diğer kategorilerdir (`internal/render/army_panel.go`).

Ordu detay panelindeki oyuncu birim kartları, recruit panelinden ayrı bir ordu birim popup'ı kullanır. Hover hit-test'i çizimdeki kategori sırasını aynen izler; popup gerçek birim örneğinin ordudaki aynı tip adetini, tur başı tahıl bakımını, savaş değerlerini ve anlık canını gösterir. Üretim maliyeti/teknoloji gereksinimi burada çizilmez; düşman ordularındaki gizli kartlar tooltip'e bağlanmaz (`internal/render/army_panel.go`, `hover_tooltip.go`).

Ordu detay panelinin birim ızgarası altındaki sağ bilgi bandı seçili oyuncu ordusu için mevcut HP'ye göre `Güç: saldırı / savunma` değerini gösterir. Düşman ordularında aynı alan yalnız istihbaratla açılan birim tiplerinden hesaplanan `Görünen güç` değerini taşır; gizli birimlerin gücü toplama dahil edilmez (`internal/render/army_panel.go`).

Üst-sol durum HUD'unda oyuncu devletinin amblem alanı, aktif senaryonun `sprites/flags/<faction-id>.png` dosyasını bulduğunda 58×58 px kare bayrağı gösterir. Aynı kare rozet devlet bilgi paneli başlığında da devlet adının solunda 44×44 px ölçüyle kullanılır. Bölge bilgi paneli sahibinin bayrağı panelin üst-sol çerçevesine bitişik, çerçevenin hemen üstünde 48×48 px kimlik rozeti olarak çizilir; bölge ve sahip devlet adlarının mevcut sol başlık konumu korunur. Tur bitimindeki AI hamle paneli, 620×180 px genişletilmiş düzen içinde aktif hamleyi yapan faction ID'sini taşıyarak ülke adının solunda sarı iç çerçevesiz 128×128 px kare bayrak rozeti gösterir. AI adımının mesajı bu panelde gösterilir; aynı mesaj genel `Bilgi` popup'ında ikinci kez çizilmez. Tur başında önceki oyuncu bildirim popup'ı da temizlenir. Bayrak asset'i yoksa ilgili kare zeminde devlet baş harfi fallback'i korunur. Bayraklar ilk kullanımda yüklenir, yol bazlı cache'lenir ve senaryo değişiminde cache temizlenir (`internal/render/panel.go`, `renderer.go`).

Kuşatma emri paneli seçili ordu detay paneli görünür kalacak şekilde onun üstündeki boş alana yerleştirilir; böylece iki panelin alt-üst örtüşmesi engellenir. Kuşatma paneli butonları kendi alanında önceliklidir, panel dışındaki input ve imleç normal ordu paneline geçebilir (`internal/render/renderer_dialogs.go`, `renderer_input.go`, `cursor.go`). Recruit eğitim kuyruğu kartlarında birim adı ve üretim süresi iki satırda gösterilir; beyaz footer kartın tam genişliğini kapatır ve sprite kenarlardan görünmez (`internal/render/recruit_panel.go`).

Ordu komutan kartında sağdaki rol, seviye, savaş ve bonus bilgileri portre üst hizasına göre çizilir; bilgi bloğu portreyle aynı üst çizgiden başlar (`internal/render/commander_component.go`).

Aynı bölgede kuşatma ordusu bölünürse ikon yerleşiminde kuşatan parça solda, ayrılan parça sağda tutulur. Hit-test sağdan tarandığı için ayrılan parça deterministik olarak seçilebilir; bölme veya birleşme sonrası bölgede kalan aynı fraksiyon ordusuna kuşatma kaydı devredilebilir (`internal/render/renderer.go`, `internal/game/{game.go,siege.go}`).

Bölge bilgi panelinde bina grid'i artık önce çizilir; hemen altında genişletilebilir aksiyon barı içinde `Diplomasi` bulunur. Aktif olaylar ve geliştirme modundaki komşu listesi bu barın altındaki panel-local viewport'ta gösterilir; uzun içerik mouse wheel ile yalnız viewport içinde kaydırılır ve scrollbar ile sınırlandırılır. Ebitengine `SubImage` hedefinin ekran koordinat sözleşmesi nedeniyle viewport içeriği çizilirken metin origin'i viewport'un ekran konumuna taşınır; aksi halde çerçeve görünürken olay/komşu metinleri clipping dışında kalır (`internal/render/panel.go`, `renderer_input.go`).

Teknoloji paneli açıkken input modal olarak teknoloji ağacına yönlendirilir ve bu yönlendirme `handleCamera()` çağrısından önce yapılır. Böylece ağaç tekerleği yalnız teknoloji pan'ini değiştirir; panel içi sürükleme veya panel dışı tıklamalar kamera zoom/pan ve harita seçim akışına ulaşmaz (`internal/render/renderer_input.go`, `tech_panel.go`).

Teknoloji ağacındaki gerçek flow içeriği viewport'tan daha dar olduğunda ortalanır. Flow'un tree viewport içindeki sağ ve sol boşluklarına yapılan sol tıklama, üstteki teknoloji kapatma düğmesiyle aynı şekilde paneli kapatır; kategori sekmeleri, teknoloji kartları ve ağaç sürükleme alanı korunur (`internal/render/tech_panel.go`).

Üst HUD'da aktif araştırma adının bulunduğu teknoloji satırı tıklanabilir; tıklama alt HUD'daki `Teknoloji` düğmesiyle aynı şekilde teknoloji panelini açar ve recruit/diplomasi panellerini kapatır (`internal/render/panel.go`, `renderer_input.go`, `cursor.go`).

**Kaynak:** `internal/render/renderer.go`, `renderer_input.go`, `renderer_dialogs.go`, `map_editor.go`, `trade_overlay.go`

`Renderer` tek tip olarak korunur; dosyalar davranış sorumluluğuna göre ayrılır. Ana dosya yaşam döngüsü, kamera ve draw orkestrasyonunu; diğer dosyalar input, modal, editör ve ticaret haritası ayrıntılarını taşır.

## Renderer Yapısı

```go
type Renderer struct {
    gs       *state.GameState
    worldMap *WorldMap          // üretilmiş harita görüntüsü

    camX, camY float64          // dünya uzayında kamera merkezi
    camScale   float64          // zoom (min fit – 5.5)

    SelectedRegion world.RegionID
    SelectedArmy   army.ArmyID

    showDiplomacy, showTech bool
    eventLog []string           // son 8 olay
    combatLog string            // 3 saniyelik bildirim

    // Duraklama menüsü
    pauseCursor int

    // Kayıt/yükleme slot seçim ekranı
    slotCursor        int
    saveSelectMode    bool
    pendingDeleteSlot string  // onay bekleyen slot adı
}
```

---

## Draw Katman Sırası

`Draw(screen)` — `internal/render/renderer.go:938`

| Sıra | Katman | Dosya |
|---|---|---|
| 0 | Özel tam ekranlar: ana menü, ayarlar, fraksiyon seçim, zafer seçim, game over | `main_menu.go`, `settings.go`, `faction_select.go`, `victory_select.go` |
| 0 | Yükleme ekranı: spinner + durum metni | `loading.go` |
| 0 | Kayıt slot seçim ekranları (PhaseLoadSelect / PhaseSaveSelect) | `load_select.go` |
| 0 | Duraklama menüsü (PhasePauseMenu) — harita altta, overlay üstte | `pause_menu.go` |
| 1 | Dünya haritası (WorldMap cache) | `mapgen.go`, `tile.go` |
| 2 | Seçim halkası (bölge) | `renderer.go` |
| 3 | Ticaret koridorları (çift yön rotalar tek hatta birleştirilir; uzak zoom'da yalnızca oyuncuya bağlı koridorlar çizilir; `trade_centers.json` içindeki `off_map` düğümler sadece etiket + bağlantı olarak çizilip bölge boyamasına katılmaz) | `trade_overlay.go` |
| 3 | Hareket hedefleri (ordu komşuları) | `renderer.go` |
| 4 | Bölge etiketleri + şehir noktası; edit mode'da bölge merkezi işaretleri, Voronoi debug overlay ve `Shape` sekmesi aktifken country shape outline/brush overlay'i; etiketler stabil sıralanır ve çakışan metinler atlanır. Vassal bölgenin ana yerleşim marker'ı üzerinde ayrıca küçük bir bağlılık rozeti çizilir; iç dolgu overlord rengini, dış halka altın vurguyu taşır | `renderer.go` |
| 5 | Ordu ikonları; çizim sırası ekran konumu + ID ile deterministiktir; edit mode'da tüm ordu/donanma birim sayıları görünür; ikon üstü sayı metni fraksiyon rengine göre kontrast uyarlamalıdır; bu tur lojistik zayiat alan ordular küçük kırmızı `!` rozeti taşır. Aktif kuşatma yürüten kara orduları ayrıca kare çerçevesiz, beyaz daire arka planlı küçük kılıç rozeti ile işaretlenir | `renderer.go` |
| 6 | UI panelleri (üst-sol durum paneli, sağ-üst tarih/menü HUD, alt-orta aksiyon HUD, bölge/ordu/minimap/event log); seçili bölge paneli memnuniyet/vergi altında `İkmal` metre satırı ile kapasite/yük ve aşım zayiatını gösterir; sahip devlet satırı tıklanabilir ve sağ tarafta devlet özet paneli açar; sahip adı fraksiyon renginde çizilir ama koyu/açık tona göre adaptif outline eklenir, böylece koyu bayrak renklerinde isim arka plan üstünde kaybolmaz. Bölge sahibi bir vassalsa bu satırın altında ayrıca `Bağlı: <overlord>` bilgisi gösterilir; eğer bu vassal doğrudan oyuncuya bağlıysa aynı blokta `Haraç: +X altın/tur` satırı da görünür ve değer tur çözümündeki gerçek vassal haraç hesabıyla aynı formülden türetilir. Devlet ismine tıklayınca açılan devlet bilgi paneli artık scroll'lu diplomasi özeti içerir; üst devlet, vassal, ittifak, ticaret ve düşman listeleri aynı panelde uzun liste halinde kaydırılabilir. Bölge panelinin en altındaki eski `Savaş/Barış/İttifak/Ticaret` hızlı aksiyonları kaldırılmıştır; onların yerine tek `Diplomasi` butonu vardır ve bu buton ilgili bölge sahibinin teklif sayfasını doğrudan seçili açar; aktif event'ler ana harita ve minimap marker'ı olmadan yalnız seçili bölgenin bilgi panelindeki `AKTİF OLAYLAR` bölümünde gösterilir. Her satır olay adını, tipini ve kalan tur sayısını taşır; aynı bölgedeki çoklu event'ler ayrı satırlarda listelenir | `panel.go`, `diplom.go`, `renderer.go` |
| 6 | Ordu detay paneli — panel artık enine genişletilmiş iki kolon düzeni kullanır: solda komutan kartı, sağda 20 slot birlik ızgarası. Komutan kartı portre, isim, seviye/XP, savaş-zafer istatistiği, saldırı-savunma bonusu ile yeni `moral / hareket / kuşatma` katkılarını gösterir; bu katkılar tek uzun satır yerine dar alana sığacak şekilde ayrı satırlara bölünür. Trait'ler düz metin yerine renk kodlu kompakt rozetlerle çizilir; rozet metinleri CamelCase görünür ve badge içinde dikey ortalanır. Bu çizim `commander_component.go` içindeki ortak helper’lar üzerinden paylaşıldığı için oyuncu kartı, düşman kartı ve komutan detay profili aynı chrome’u kullanır; dar alanlı komutan listesi ise daha yüksek satır ve daha büyük portre ile aynı sistemin tek satırlık `+N` taşma rozetli varyantını kullanır. Ordu paneli varyantında alt bölümde tekrar eden etki özeti çizilmez ve komutan aksiyonu ayrı buton yerine doğrudan portre tıklamasına bağlanır; oyuncu komutan portresine tıklayarak `Komutan Ata / Değiştir` panelini açar. Komutan yoksa aynı alanda atanmadı uyarısı ve portre yer tutucusu görünür. Dost toprakta toparlanan hasarlı birimler kart köşesinde küçük `+` rozeti taşır; aynı tur lojistik attrition alan ordular başlık satırında `Lojistik zayiat` metni gösterir; nakliye filosu taşınan kara komutanını da komutan kartında özetler ve komutan panelinde filo komutanından ayrı profiller/ayırma aksiyonları sunar. Panelin boş alanına tıklamak paneli kapatmaz; yalnız kapatma düğmesi veya başka seçim akışı paneli kapatır. Komutan kartı profil metinlerinin yatay taşmaması için 260 px genişlikte ve uzmanlık taşma rozetinin sığması için 12 px ek dikey payla çizilir; istatistik satırları kartın kullanılabilir metin genişliğine göre kırpılır. Başlık bandındaki `BÖL` ve `BİRLEŞTİR` butonları sağdaki `Hareket`/`Takviye aktif` durum metinleri için 92 px ek sağ pay bırakır; `BÖL` her zaman en sağdaki aksiyon düğmesidir, `BİRLEŞTİR` görünürse onun soluna yerleşir ve çizim/hit-test aynı buton rect helper’larını kullanır. `portrait_asset` dosya adı verilirse senaryo `sprites/commanders/` altında, relative alt klasör verilirse doğrudan `sprites/` altında cache'lenir; dosya yoksa yer tutucu çizilir. Bu sayede kısa JSON değerleri ile eski `commanders/...` kayıtları birlikte desteklenir; oyuncu `reveal_enemy_strength` etkisi açtıysa düşman ordu paneli menzil şartı olmadan tam istihbarat moduna geçer. Düşman ordu paneli de artık aynı iki kolon yerleşimi paylaşır: solda komutan kartı, sağda birlik alanı. Tam istihbaratta gerçek/kısmi birimler, istihbarat yoksa gizli kartlar çizilir; komutanın atanıp atanmadığı, adı/portresi, rozetleri ile muharebe ve operasyon katkıları seçili düşman orduda görülebilir | `army_panel.go`, `commander_panel.go`, `commander_component.go` |
| 6 | Bölge üretim UI — bina kartlarında seviye (`Lv`) + kuyruk adet/ilk tamamlanma turu etiketi ve tekrar tıklayınca iptal; kuyruktaki `N Tur` bilgisi artık sprite üstüne çıplak metin olarak değil, kart altına yakın kontrastlı koyu-altın pill rozeti içinde çizilir; bina gereksinim satırı kart üstünde çizilmez, hover hint içinde gösterilir; `port` kartı ve liman inşası literal `terrain == coast` yerine `Region.IsCoastal` predicate'ine bakar, yani denize komşu kara bölgelerinde görünür ve iç bölgelerde gizlenir; bina ve birim kartlarının uygunluk/soluk görünümü altınla sınırlı değil, `ResourceCost.CanAfford` üzerinden tüm mallara göre hesaplanır; kaynakları yeterli, maksimum seviyeye ulaşmamış ve inşaat kuyruğunda olmayan bina kartlarının isim etiketi yeşil çizilir; birim kartları sadece isim + tur süresi gösterir; recruit başlığı aynı tur kapasitesini kara için `Kışla`, kıyı bölgelerinde ayrıca deniz için `Liman` ayrı hatlarıyla gösterir; recruit kuyruğunda aynı tur kapasite içine giren emirler daha parlak, bekleyen emirler ise daha soluk kart stiliyle ayrılır; hover tooltip’ler durum/maliyet/gereksinim bölümlerine ayrılır, eksik kaynak satırları ve karşılanmayan gereksinimler kırmızı vurgulanır; bina maliyetleri `mevcut/ihtiyaç` formatını korurken birim maliyetleri yalnız gerekli miktarı gösterir ve karşılanmayan satıra `eksik` uyarısı ekler; bölge bilgi paneli artık `Altın/Tahıl` ile sınırlı değildir, efektif `Altın/Tahıl/Demir/Kereste/Taş/Baharat/Kumaş` üretimini iki kolon grid içinde gösterir, sahip satırını başlığın hemen altında etiketsiz çizer ve memnuniyet/vergi satırlarında değer metni, bar ve vergi düğmelerini çakışmadan ayıran özel metre yerleşimi kullanır; maksimum seviyedeki bina kartları alt satırda ayrı `Maks` yazısı basmaz, bunun yerine sol üstteki `Lv` rozetinin arka planı kırmızı uyarı tonuna döner; kaynak etiketleri `economy.ResourceKind`, din adları `religion.DisplayNameTR`, diplomasi durumları `faction` stance metadata'sı ve arazi/yerleşim etiketleri `world` metadata'sı üzerinden beslendiği için tooltip/HUD/panel metinleri tek modelden gelir; `- xN +` çoklu eğitim kontrolü korunur | `panel.go`, `recruit_panel.go`, `hover_tooltip.go`, `building_card_component.go` |
| 6 | Olay logu akordiyonu — daralt/genişlet, wrap edilmiş kartlar, X ile kapatma, tıklayınca detay popup; üstteki `Kodex` düğmesi `Tümü/Hazır/Takvim/Kilitli` filtreli, solda kısa özetli liste sağda detay gösteren daha geniş historical chain popup'ını açar; savaş olaylarının detayında komutan XP/seviye/trait ilerlemesi de listelenir; Kodex listesi focus-scroll ve mouse wheel ile görünür pencere içinde kalır | `panel.go`, `renderer.go`, `game.go`, `ui_modals.go` |
| 6 | Edit mode alt-sol bilgi HUD'u — seçili bölge/settlement/ordu özeti ve edit butonları | `map_editor.go` |
| 7 | Diplomasi paneli (Tab) — tam ekran overlay; sağ kolon varsayılan olarak seçili devletin savaş, dış ittifak ve aktif ticaret ortaklarını gösterir, vassal/overlord hiyerarşisini ayrıca belirtir. `Geçmiş` düğmesi aynı kolonu teklif geçmişine çevirir; burada `Gelen / Giden / Tümü` ile `Barış / Ticaret / İttifak / Savaş` filtreleri, kart üstü küçük aksiyon ikonları ve sonuç rozetleri vardır. Standart teklifler gerçek blok nedenine göre aktif veya `PASİF` çizilir ve pasif olanlar seçilemez. Kurulu dış ilişkide aynı satırlar `İttifakı Bitir / Ticareti Bitir` işlemine, footer da `Anlaşmayı Bitir` aksiyonuna dönüşür; ittifak ve ticaret birbirinden bağımsız kapatılır. Seçili hedef oyuncunun doğrudan vassalıysa sağ-alt kartta onaylı `Vasallığı Bitir / İlhak Et` eylemleri görünür. Görünüm düğmesi, history filtreleri ve kartları yalnız gerçek tıklamada state değiştirir; hover browse veya filtre seçimi yapmaz | `diplom.go`, `renderer.go`, `cursor.go` |
| 8 | Teknoloji paneli (T) — tam ekran ağaç görünümü; prerequisite bağlantıları node arkasında köşeli/ortogonal hatlarla çizilir, çoklu prerequisite dalları küçük lane offset'leri ile ayrılır; teknoloji kartının görünüm ve hit-test seam'i ortak `techCardComponent` üstünden yürütülür, böylece çizilen rect ile tıklama rect'i aynı projection çıktısını paylaşır. Kart başlığı badge alanını hesaba katar, gerekirse iki satıra wrap olur; effect özeti de kontrollü çok satır kullanır, bu yüzden uzun Osmanlı/Türkçe teknoloji adları kart dışına taşmaz. Kart kategori ikonları 20 px, üst filtre sekmesi ikonları 22 px çizilir ve başlık/etiket boşlukları bu ölçülere göre ayrılır | `tech_panel.go`, `tech_card_component.go` |
| 8 | AI tur overlay'i — aktif AI devletinin adı ve yakın/uzak hamle durumu üst-orta bantta gösterilir | `renderer.go` |
| 9 | Info popup bildirimi (combatLog, olay loguna yazmaz); normal akışta üst-orta konumu biraz aşağı alınır, AI tur overlay'i görünürken HAMLELER panelinin altına güvenli boşlukla yerleşir | `renderer.go`, `panel.go` |
| 10 | Savaş ilan, savaş planı, kuşatma kararı, savaş raporu, genel onay, zafer detay ve event detail diyalogları; savaş ilan modalı artık tek satırlık `evet/hayır` onayı değil, iki cepheli koalisyon önizlemesi çizer. Sol tarafta oyuncu cephesi, sağ tarafta savunan cephe görünür; kesin katılan vassallar ve mevcut savaş içindeki müttefikler ayrı kartlarda listelenir, oyuncu kendi müttefikleri için checkbox ile çağrı seçer. Bu seçim `InputAction.WarAllies` alanıyla oyun katmanına taşınır. İlan resolve olduktan hemen sonra ayrıca ayrı `Savaş Özeti` modalı açılarak gerçek katılımcıları gösterir. Savaş planı modalı üç duruş kartı (`Agresif/Dengeli/Savunmacı`) ve combat matematiğinden türetilmiş önizlemeleri taşır; üst satırda saldıran ve savunan komutanın gerçek muharebe bonusları (`saldırı/savunma/moral`) ayrıca yazılır. Tahkimli kara hedefleri ayrıca kuşatma kararı akışına girer; kuşatma kararı ve aktif `Kuşatma Emri` paneli de seçili komutanın `moral / hareket / kuşatma` katkılarını ayrıca görünür tutar. Savaş raporu taraf kartları da artık ayrı `Komutan` bloğu içerir; her taraf için portre, isim, muharebe etkileri ve operasyon etkileri aynı modalda görünür, alttaki `Komutan gelişimi` bölümü ise savaş sonrası XP/trait ilerlemesini taşımaya devam eder. Bu komutan bloğu da `commander_component.go` içindeki ortak kompakt strip helper’ını kullanır; böylece ordu paneli, komutan paneli ve savaş raporu aynı commander chrome dilini paylaşır | `renderer_dialogs.go`, `ui_modals.go`, `battle_report.go`, `war_summary.go`, `commander_component.go` |
| 11 | Tarihsel olay popup; choice varsa aynı modal üzerinde A/B karar butonları, effect özeti, follow-up event etiketi ve trigger koşulu önizlemesi çizer. Bu popup draw ve input tarafında gerçek üst modal önceliğine sahiptir; altta bekleyen onay/teklif diyalogları choice butonlarının tıklamasını yutamaz | `panel.go`, `ui_modals.go`, `game.go`, `renderer.go`, `cursor.go` |

Not: Diplomasi panelindeki liste üretimi `sortedDiplomacyFactions()` üzerinden yapılır ve elenmiş (`IsEliminated=true`) fraksiyonlar listelenmez. Liste üstündeki `Alfabetik`, `İlişki` ve `Güç Sıralaması` butonları renderer state'indeki sıralama modunu değiştirir; `İlişki` modu oyuncuyla olan `Relation.Score` değerini azalan sıralar, eşitlikte oyuncuyla kara sınırı paylaşan fraksiyonu öne alır ve son eşitliği faction ID'siyle çözer. `Güç Sıralaması` aktif devletler arasında `factionMilitaryPowerStanding` kaynağını kullanır; seçim sonrası focus ve scroll başa alınır.
Not: Diplomasi hedef listesindeki her satırın en solunda 40×40 px kare faction bayrağı çizilir; asset bulunamazsa devlet adının baş harfi fallback'i kullanılır. Devlet adı kolonu bayrak ve 10 px boşluk sonrasında toplam 340 px içinde çizilir; ilişki/durum kolonu aynı satır geometrisinden daha solda ve kalan genişlikte çizilir. Uzun durum etiketleri ilişki kolonunun genişliğine göre kırpılır (`internal/render/diplom.go`).
Not: Edit mode region paint override'ları için `WorldMap` override öncesi `baseRegionAt` snapshot'ını saklar; böylece baseline hesabı ikinci bir tam `prepareWorldMapData` çağrısı yapmadan aynı build içinden alınır.
Not: Senaryo ticaret ağı artık hem region tabanlı merkezleri hem de `trade_centers.json` içindeki `off_map=true` dış hat düğümlerini destekler; bu düğümler `name_tr`, `world_x`, `world_y` ve `links` ile tanımlanır, yalnız render etiketinde görünür ve nearest-center / trade-map tint hesabında dışarıda bırakılır. `unlock_year` verildiğinde düğüm ve ona bağlı koridorlar belirtilen yıl gelene kadar tamamen pasif kalır.
Not: Oyuncuya gelen diplomasi teklifleri (ilk sürüm: barış) ortak modal/panel/button geometrisi kullanan anlaşma paneli ile `Kabul Et` / `Reddet` olarak yanıtlanır.
Not: Diplomasi ekranı iki sayfadır: ilk sayfa devlet listesi, seçilen devlet için ikinci sayfa teklif paneli açılır; `Geri` ile listeye dönülür. Her iki sayfadaki sağ kolon seçili devletin aktif ilişkilerini esas alır; savaş ve ittifak `Relation.Stance`, ticaret ise ittifaktan bağımsız aktif `TradeRoutes` kayıtları üzerinden listelenir. Vassal realm içindeki zorunlu allied kayıtları dış ittifak listesine karıştırılmaz. Geçmiş sürekli açık tutulmaz; `Geçmiş / İlişkiler` düğmesi aynı kolonda görünüm değiştirir. Liste sayfası mouse wheel'i panel gövdesinde tüketir, görünür scrollbar çizer ve kart chrome'u ortak compose helper'larıyla render edilir. Liste seçimi press yerine release ile tamamlanır; drag eşiği aşılırsa satır seçimi iptal edilip liste row-height bazlı sürükleme scroll'una geçer. Vassal satırına ikinci kez tıklanınca dış overlord teklif paneli açılır; oyuncunun kendi vassalında mevcut vassal yönetim kartı korunur. Tam teklif paneli ile bölge panelindeki hızlı diplomasi butonları aynı validasyon helper'ını kullanır; geçersiz aksiyonlar `PASİF` görünür, fare/klavye odağına alınmaz ve `Teklif Gönder` backend çağrısından önce de bloklanır. Doğrudan oyuncu vassalında standart tekliflerden ayrı yönetim kartı açılır; iki yönetim eylemi de genel confirm modalı üzerinden oyun katmanına iletilir. Ticaret teklifindeki yüzde gerçek kabul motoruyla aynı helper'dan gelir; yani panelde görülen olasılık, submit anındaki trade acceptance kuralından ayrı akmaz. Teklif paneli ayrıca `Heyet`, `Hediye` ve `Vassallık` aksiyonlarını taşır; vassal / overlord hedeflerinde durum satırı özel hiyerarşi etiketi gösterir.
Not: Sağ tık savaş onayı deniz-donanma hareketinde düşman deniz bölgesine giriş için açılmaz; bu hareket savaştan bağımsız serbesttir. Ancak hedef deniz bölgesinde savaş halindeki düşman donanma varsa artık doğrudan move üretilmez; önce `Deniz Muharebesi Planı` modalı açılır ve seçilen duruş `ActionMoveArmy.BattleStance` alanıyla oyun katmanına taşınır. Kara ordusu savaş halindeki veya savaş ilanı sonrası çatışmaya gireceği düşman kara bölgesine sağ tıkladığında da aynı modal `Kara Muharebesi` başlığıyla açılır. Hedef kara bölgesi tahkimliyse savaş ilanı onayından sonra doğrudan move yerine `Kuşatma Kararı` modalına geçilir; kuşatma birimi olmayan ordu da kuşatma başlatabilir ve `Genel Hücum` düğmesini kullanabilir, ancak gedik yoksa tahkimat doğrudan düşmez. Aktif kuşatma yürüten bir ordu seçildiğinde ayrı `Kuşatma Emri` paneli açılır; panel ve kuşatma karar mesajı komutanın operasyon bonuslarını (`moral / hareket / kuşatma`) da açıkça yazar. Oyuncu bu panel açıkken başka komşu bölgeye sağ tıklarsa kuşatma ayrıca onay istemeden kaldırılmış sayılır ve hareket emri çözülür. Aynı aktif kuşatmaya aynı fraksiyon, müttefik fraksiyon veya aynı vassal zincirindeki realm üyeleri normal hareketle destek verebilir; ilgisiz üçüncü devletlerin ikinci kuşatma hamlesi render'da da bloklanır. Tahkimli ve zaten kuşatılmış kara bölgesinde enemy besieger ile savaş mümkünse sağ tık hareketi kuşatma kararına gitmez; battle plan / declare-war flow açılır ve kazanılan savaş kuşatmayı kaldırır ama yeni kuşatma oluşturmaz. Seçili nakliye filosu düşman kıyıya çıkarma yaparken savunan ordu varsa bu kez `Çıkarma Muharebesi Planı` açılır ve seçim `ActionDisembarkArmy.BattleStance` olarak iletilir; preview üst bandı bu durumda da gerçek çıkarma komutanını gösterir. Donanma ayrıca sahibi olunan, müttefik olunan veya aynı realm içinde olunan, port settlement içeren komşu kara bölgesine docking emri de alabilir; bu durumda savaş onayı açılmaz ve filo deniz bölgesi konumunu koruyup liman anchor'ına bağlanır. Seçili kara ordusu, komşu deniz bölgesindeki dost nakliye filosuna doğrudan sağ tıklarsa ayrı `Gemiye Bin` onay diyaloğu açılır; UI yalnız filonun kalan kapasitesi orduyu taşımaya yetiyorsa bu emri sunar ve uygun filo ikonu ayrıca `BIN` hedef rozeti ile vurgulanır. Seçili nakliye filosu gemide birlik taşırken dost, aynı realm veya boş kara bölgesine indirme yapabiliyorsa hedef bölge üstünde `IN` rozeti görünür ve sağ tık önce `Karaya In` onay diyaloğu açar; bu onay normal move değil zorunlu indirme aksiyonunu tetikler, yani liman uygunsa bile asker önce karaya iner ve filo denizde kalır.
Not: Tam ekran seçim/menü ailesindeki metinler (`main_menu`, `scenario_select`, `faction_select`, `victory_select`, `load_select`) artık doğrudan `DrawText*` çağrılarıyla değil, `internal/ui.Label` üstünden ortak `TextRenderer` ile çizilir; font varyantı ve hizalama UI primitive'inde tanımlanır. Modal açıklamaları, info popup, event detail/codex detail ve historical event açıklama blokları `WrappedLabel`, ikon/sayaç gölgeleri ise `OutlinedLabel` primitive'i üzerinden ortaklaştırılmıştır. Zafer seçim ekranındaki kartlar da aynı wrap primitive'lerini kullanır; açıklama ve hedef özeti badge alanına çarpmadan iki satıra akabilir ve uzun senaryo hedeflerinde kart yüksekliği buna göre artırılmıştır.

---

## UI Components

### Shared UI Layer

**Kaynak:** `internal/ui/*.go`, `internal/render/ui_buttons.go`, `internal/render/ui_bridge.go`, `internal/render/ui_modals.go`

Tam widget ağacı henüz tüm render katmanına uygulanmış değil; ancak ana ekranlar ve yüksek riskli etkileşim yüzeyleri artık ortak UI geometrisi üstünden çalışır. Pratikte bu şu anlama gelir:

- input frame başına bir kez `gameui.InputState` olarak toplanır,
- tıklanabilir yüzeyler `gameui.Button` ve `gameui.Rect` builder'larıyla tek yerde tanımlanır,
- draw, cursor ve click hit-test aynı yüzeyi paylaşır,
- panel açıkken arka harita etkileşimi ilgili overlay hit-test'i ile tüketilir.
- oyuncu turu dışına (`PhaseAITurn`, `PhaseTurnResolution`) geçerken `Renderer.PrepareForTurnAdvance()` seçili bölge/ordu, recruit-diplomasi-teknoloji-ticaret panelleri, event/victory detail popup'ları ve trade map modunu temizler; böylece AI overlay veya modal event/teklif akışında eski oyuncu panelleri ekranda kalmaz. Turn resolution tamamlanıp `PhasePlayerTurn` yeniden açıldığında `SelectPlayerCapitalRegion()` aktif oyuncu başkentinin region'ını aynı seçim akışıyla seçer; geçerli başkent bulunamazsa seçim boş bırakılır.
- ortak button/dropdown/modal/shape yardım paneli stilleri `internal/render/ui_theme.go` altında tutulur.
- `internal/ui.Button` primitive'i artık cache'li ikon (`IconID`) taşıyabilir; ikon bitmap'leri `assets/ui/icons/*.png` altından yüklenir ve aynı draw path içinde label ile birlikte hizalanır. Böylece close/back/menu/codex ile birlikte modal onayları, trade aksiyonları ve save/delete mini aksiyonlarında da ayrı manuel ikon çizim yolu açılmaz.
- Aynı primitive'de `ButtonStyle.TextOffsetY` fiilen draw hattında uygulanır; bu sayede farklı yükseklikte footer/modal butonlarında icon ve label baseline'ı ekran bazlı sabitlerle yeniden ayrı ayrı ayarlanmaz.
- `internal/ui.Manager`, focus edilebilir widget'lar için ileri/geri tab-order davranışını test edilebilir şekilde merkezileştirir.
- ana menü, senaryo, fraksiyon, zafer, pause ve kayıt/yükleme slot ekranları `Tab` focus geçişini `internal/ui.Manager` üzerinden kullanır.
- temel HUD/modal geometri yüzeyleri 1280x720, 1600x900 ve 1920x1080 headless smoke testinden geçer; ana menü aynı çözünürlüklerde headless draw-call smoke testinden de geçer.

Bu hat şu ekranlarda aktif kullanımdadır:

- trade
- diplomasi
- teknoloji
- pause ve save/load
- ana menü, senaryo, fraksiyon ve zafer seçim ekranları
- HUD üzerindeki küçük aksiyon düğmeleri
- recruit panel close / kart / kuyruk iptal yüzeyleri
- ordu split/merge overlay yüzeyleri
- edit mode inspector/form tab ve aksiyon yüzeyleri
- confirm / war confirm / battle plan / battle report / event detail / historical event modal yüzeyleri
- historical modal açıkken arka oyun inputu tamamen bloke edilir; choice varsa `1/2`, `Enter`, ok tuşları veya mouse ile seçim yapılır ve imleç önceliği bu modalın butonlarına verilir
- oyuncuya gelen diplomasi teklif modal yüzeyi; son çözümlenmiş teklif geçmişiyle birlikte çizilir ve history kartları ilgili fraksiyonun teklif ekranına hızlı geçiş sağlar
- `HandleInput()` artık `PhaseAITurn` ve `PhaseTurnResolution` sırasında genel harita/HUD inputunu tamamen kilitler; bu fazlarda yalnız üst modal teklifler, historical event seçimleri ve tam ekran menü akışları etkileşim alır.
- edit mode shape yardım paneli ve stroke preview overlay primitive'i

Ortak modal builder'ları sıcak çizim hattında gereksiz `Children` slice allocation'ı yapmaz; çocuk widget gerekiyorsa modal örneğine açıkça eklenir. Çekirdek modal/button builder'ları allocation testleriyle korunur.

### Dropdown Component

**Kaynak:** `internal/ui/dropdown.go`

Edit mode'da kullanılan yeniden kullanılabilir dropdown component. Sahip, arazi, yerleşim tipi ve veri editöründeki birim tipi seçimlerinde kullanılır.

```go
type Dropdown struct {
    x, y, w, h int
    options    []string
    selected   string
    scroll     int
    open       bool
}
```

**Metodlar:**
- `SetPosition(x, y float32)` — dropdown konumunu ayarlar
- `SetOptions(options []string, selected string)` — seçenekleri ve seçili değeri ayarlar
- `Toggle()` — aç/kapat
- `Close()` — kapat
- `IsOpen() bool` — açık mı kontrolü
- `HitTest(mx, my float64) bool` — fare pozisyonu dropdown içinde mi
- `Scroll(dy float64)` — tekerlek ile kaydırma
- `GetSelectedOption(mx, my float64) (int, bool)` — tıklanan seçeneği döndürür
- `Draw(screen *ebiten.Image)` — render

---

## Kamera Sistemi

**Koordinat sistemi:** Dünya uzayı `(WorldW × WorldH)` px, ekran uzayına dönüşüm:

`WorldW`, `WorldH`, `shape_offset_*` ve `shape_scale_*` aktif senaryonun `scenario.json` içindeki `map` alanından okunur. Alan eksikse renderer eski varsayılanları kullanır (`2892×1440`, offset `-530/-180`, scale `2.025/2.025`).

```
screenX = (worldX - camX + worldY * mapShearX) * camScale + ScreenWidth/2
screenY = (worldY - camY) * camScale * mapPitchY + ScreenHeight/2
```

`mapPitchY = 1.0`, `mapShearX = 0.0` → şu an düz 2D (izometrik bükme kapalı)

**Zoom:** Fare tekerleği ile fare pozisyonuna odaklanarak büyütür. Uzaklaşma limiti `internal/render/renderer.go:minCameraScale` üzerinden aktif senaryonun `world_width` / `world_height` değerlerinden gelen `WorldW` / `WorldH` boyutuna göre hesaplanır; oyuncu haritayı ekrana tamamen sığdıran ölçeğin altına inemez. `resetCamera()` ise ilk açılışta ve kamera resetlerinde bu minimumun `1.55x` üstünden başlar; oyuncu fraksiyonu ve geçerli `capital_settlement_id` varsa kamera dünya merkezine değil doğrudan o başkent settlement koordinatına odaklanır. Başkent ekran kenarına çok yakınsa ilk frame viewport yarı-genişliği kadar clamp uygulanır; böylece save load veya yeni oyun sonrası harita boşluğa taşmadan başkentte açılır. Yakınlaşma üst sınırı `5.5`.

**Sürükleme:** Orta fare tuşu basılıyken dünya uzayı delta hesaplanır.

`Renderer` artık AI tur sinematiği için üç ek seam taşır:

- `CameraSnapshot()` mevcut `camX/camY/camScale` değerlerini saklar
- `CenterCameraOnRegion(rid)` kamerayı ilgili bölge merkezine taşır
- `RestoreCamera(state)` AI turu bitince oyuncunun eski bakışını geri yükler

Bu seam, `internal/game/game.go` içindeki AI stepper akışının yalnız yakın hamlelerde kamerayı oynatmasını sağlar; uzak hamlelerde overlay güncellenir ama kamera zıplamaz.

---

## WorldMap Cache

`WorldMap` — `internal/render/mapgen.go`

Harita, her fraksiyon sahipliği değişiminde `MarkDirty()` ile işaretlenir ve bir sonraki `Refresh()` çağrısında yeniden üretilir. `Refresh()` ayrıca oyuncu, relation stance ve `OverlordID` alanlarından allocation üretmeyen bir diplomasi imzası çıkarır; bu imza değiştiğinde AI veya oyuncu kaynaklı ittifak/vassallık güncellemesi harita dokusunu bir kez yeniler. Bölge poligonları normalde `country_shapes.json`'dan gelir; edit mode sırasında ise `GameState.ShapeData` içindeki anlık shape verisi önceliklidir, böylece paint edit sonrası `rebuildEditWorldMap()` doğrudan yeni sınırı gösterir.

Normal harita modunda sınırlar realm bazında sınıflandırılır. Oyuncu ile vassalları tek realm sayılır ve dış konturları tek dünya pikseli kalınlığında keskin altın renkle çizilir. Oyuncunun müttefikleri ile onların vassalları doygun turkuaz-yeşil, savaş halindeki düşman realm'leri ve vassalları doygun kırmızı dış kontur alır. Ortak kara sınırı iki taraflı boyanmaz; tek taraflı tek piksel üretildiği için zoom altında katlanarak kalınlaşmaz. Aynı devlet veya aynı vassal realm içindeki bölge sınırları dış konturla aynı tek piksel kalınlığındadır ancak arazi rengine düşük oranlı koyu blend uygulanarak daha soluk idari çizgi halinde kalır. Tarafsız devletlerin dış sınırları da net görünür; ticaret modu kendi ticaret merkezi sınır paletini korur.

Deniz bölgeleri `internal/render/mapgen.go:buildSeaRegions` içinde kara pikselleri bariyer kabul eden multi-source BFS ile üretilir. Seed araması önce mevcut shape dönüşümlü koordinatı, sonuç çıkmazsa ham `world_x/world_y` koordinatını dener; bu, senaryo verisindeki deniz merkezlerinin dünya pikseli olarak tutulduğu durumlarda `_sea_*` seed uyarılarını engeller.

Deniz ve kara region raster alanlarından `WorldMap.RegionAnchor` hesaplanır. Deniz orduları ve deniz hareket hedefleri JSON merkez koordinatı yerine bu gerçek piksel anchor'ını kullanır; anchor, bölgenin kendi piksel alanı içinden seçildiği için kıyıda kara poligonunun kapattığı deniz bölgelerinde filo ikonları karanın üstüne düşmez.

Kara bölgelerde görünen yerleşim işaretleri `regions.json` içindeki `settlements[]` alanından gelir. `WorldMap` her yerleşim için `SettlementAnchor` hesaplar; koordinat yanlışlıkla bölge dışına verilirse log uyarısı basılır ve aynı region içindeki en yakın piksele fallback yapılır. `port` settlement'lar liman simgesi, `fortress` settlement'lar kale simgesi, diğerleri nokta olarak çizilir; ulusal başkent olan settlement'lara bunların yanında ek bir yıldız rozeti çizilir. Haritada seçili settlement etiketi altın tonla vurgulanır. Kara ordu ikonları mümkünse `port` olmayan yerleşim anchor'ına kaydırılır; dock edilmiş filolar ise `DockedSettlementID` ile doğrudan liman anchor'ında görünür. Nakliye filosu cargo taşıyorsa yuvarlak filo ikonunun üstüne küçük kare bir badge ve içindeki taşınan birlik sayısı çizilir.

Yerleşim marker sprite'ları beyaz daire arka planıyla aynı `(sx, sy)` merkezinde çizilir; sprite'ın dikey eksende ayrıca kaydırılması kullanılmaz. Böylece ikonun beyaz daire içinde üstte fazla, altta eksik boşluk bırakması engellenir (`internal/render/renderer.go:2371`).

Edit mode'da `world_x/world_y` merkezleri ayrı işaretlerle çizilir. Kara ve deniz bölgesi odak noktaları farklı renktedir; deniz seçiliyken odak işareti kara seçiminden farklı mavi/camgöbeği tona döner. Shift + sol sürükleme bu koordinatları değiştirir; Voronoi sınırları `WorldMap` raster cache'ine bağlı olduğu için sürükleme sırasında sadece merkez işareti güncellenir, fare bırakıldığında cache bir kez yeniden oluşturulur.

---

## Input Yönetimi

`HandleInput()` döner: `InputAction{Kind, ArmyID, TargetArmyID, TargetRegion, TargetFaction, BuildingID, Delta}`

**Just-pressed takibi:** `prevKeys`, `prevMouse` map'leri tutulur; `keyJustPressed()` / `mouseJustPressed()` bir frame'lik tetikleme sağlar.

**Tık öncelik sırası:**
1. Açık detay paneli kapatma düğmeleri (bölge/ordu)
2. Alt-orta aksiyon HUD butonları (diplomasi, teknoloji, tur bitir)
3. Olay logu akordiyonu: başlık butonu paneli daraltır/genişletir, kart X'i olayı kapatır, kart gövdesi detay popup açar
4. UI bölgesi (üst-sol durum paneli / sağ-üst tarih-menü HUD / alt-orta aksiyon HUD / sağ panel) → geçersiz say
5. Bölge paneli aksiyonları: vergi +/- düğmeleri, bina kartına (kurulu kartlar dahil) tıklayarak inşa/yükseltme veya kuyruk iptali; `is_locked=true` bölgelerde vergi/inşa/birim alımı hit-test'te kapatılır
6. Birim oluştur paneli (`recruit_panel.go:RecruitPanelHitTest`); kıyı olmayan bölgelerde deniz birimleri gösterilmez
7. Bölge/birim oluştur paneli boş alan tıklamaları → tüketilir, arkadaki haritaya düşmez
8. BÖL/BİRLEŞTİR butonları (seçili ordu varsa, `army_panel.go` hit-test)
9. Ordu/donanma etiketi tıklaması — `armyHitAt()` üzerinden offset'li 14px yarıçap; çakışmada settlement/bölge seçimine üstün gelir
10. Aktif bölge event marker'ı — `activeRegionEventHitAt()`; aynı settlement'taki orduyla çakışsa army seçimi önce gelir
11. Yerleşim seçimi (`settlementHitAt`)
12. Bölge seçimi (WorldMap pixel lookup)

Bölge bilgi panelindeki sahip devlet satırı ayrı hit-test yüzeyidir. Bu satıra tıklanınca yerleşim paneli slotunda devlet detay paneli açılır; aktif araştırma, tamamlanan teknolojiler, toplam mallar ve ticaret özeti gösterilir. Settlement paneli ayrıca başkent statüsü, aktif başkent taşıma kuyruğu ve oyuncu için `Başkent Yap` aksiyonunu gösterebilir. Panel açıkken tıklamalar arkadaki haritaya düşmez.

Edit mode'da oyun HUD/panelleri çizilmez; harita, üst edit HUD ve alt-sol sekmeli inspector görünür. Sol tık settlement, bölge veya ordu seçer; settlement sürükleme koordinatı canlı taşır ve başka kara region'a bırakılan settlement o region'ın `settlements[]` listesine aktarılır. Alt + sol tık tıklanan kara bölgeye yeni settlement ekler; Ctrl + Alt + sol tık tıklanan bölgenin `shape_id` alanını paylaşan yeni Voronoi seed region oluşturur. Delete seçili settlement'ı, settlement seçili değilse seçili region'ı siler. Shift + sol sürükleme seçili bölgenin `world_x/world_y` merkezini taşır ve fare bırakıldığında harita cache'ini yeniler. Inspector `Harita` sekmesindeki `Yerlesim Ekle`, `Tip`, `Ana Yap`, `Isim`, `Arazi`, `Sahip`, `Ad TR`, `Ad EN`, `Kilit`, `-10 Tur`, `+10 Tur`, `Komsu Sync`, `Bolge Ekle`, `Bolge Sil`, `Yerlesim Sil` ve `Kaydet` butonları region/settlement metadata işlemlerini doğrudan çalıştırır. `+10/-10 Tur`, `unlock_turn` alanını değiştirir; `is_locked=true` ve `unlock_turn>0` ise bölge aktif tur o değere ulaştığında otomatik açılır. Deniz region'larında settlement işlemleri kapalı kalır ama bölge odaklı seçim, merkez taşıma, komşu sync, ekleme/silme ve owner/terrain düzenleme aynıdır; inspector bu seçimlerde açıkça `Deniz Bolgesi`, `Deniz bolgesinde yerlesim yok.` ve pasif `Denizde Yok` etiketi gösterir. Settlement odaklı pasif butonlar da bağlama göre `Tip Yok`, `Isim Yok`, `Silinmez` ya da settlement seçimi bekleniyorsa `Tip Sec`, `Isim Sec`, `Sil Sec` etiketine döner. `Shape` sekmesinde `Shape Boya/Sil` yalnız `shape_id` taşıyan kara region'ları düzenler; `Bolge Boya/Sil` ise kara veya deniz seçiliyken doğrudan `region_shapes.json` override katmanına yazar, böylece ülke dış sınırının dışına taşan kara genişletmeleri ve deniz alanı dağılımı restart sonrası da korunur. Stroke baseline'ı override öncesi world map'ten alındığı için aynı dış sınır üstünden tekrar geçmek eski override kaydını düşürmez. Sağ üst yardım paneli ortak `Panel + Label` primitive'leriyle çizilir; canlı brush preview katmanı ise halen render-spesifik overlay olarak kalır. `Tip`, `Arazi` ve `Sahip` inspector yanında kaydırılabilir dropdown açar; seçilen satır ilgili `type`, `terrain` veya `owner_id` değerini doğrudan yazar. `Veri` sekmesi faction ekleme/düzenleme formu, faction silme, başlangıç kaynakları/playable/AI değeri, başlangıç kara ordusu/donanma ekleme-silme ve seçili ordu/donanma birim tip-sayılarını düzenler. Donanma ekleme liman tipli yerleşimin kara region'ından komşu deniz region'ına `is_naval: true` ordu yerleştirir. Faction formu ID, `name`, `name_tr`, din, renk, playable, kaynaklar, AI, hedef faction, diplomasi `stance` ve `score` alanlarını tek yerde toplar; formdaki `Kaydet` değişikliği uygular ve senaryo JSON dosyalarını yazar. `Kaydet` / `Ctrl+S` artık `regions.json`, `country_shapes.json`, `factions.json`, `relations.json`, `armies.json` ve gerektiğinde `region_shapes.json` dosyalarını birlikte yazar. F2/Enter seçili settlement adını düzenler, Ctrl+S `ActionSaveScenario` üretir.

Voronoi debug overlay `V` ile açılıp kapanır. Overlay `WorldMap.BoundaryPixels` ile seçili veya hover bölgenin gerçek raster sınırını camgöbeği piksellerle çizer. `WorldMap.VisualNeighbors` üzerinden raster sınır komşularını çıkarır ve JSON `neighbors` listesiyle karşılaştırır: yeşil çizgi görsel+JSON komşu, kırmızı çizgi sadece görsel komşu, gri çizgi sadece JSON komşudur. Sağ üst panel hover pixel'in `RegionAt` sonucunu, senaryo koordinatını ve seçili bölgenin visual/json komşu sayısını gösterir. `Komsu Sync`, seçili region'ın görsel komşularını JSON `neighbors` listesine yazar; eklenen/çıkarılan her komşuda karşı region listesi de iki yönlü güncellenir.

Edit mode'da `editDirty` true iken ESC doğrudan çıkmaz; genel onay modalı üç seçenekle açılır: `Kaydet` önce `ActionSaveScenarioAndGoMainMenu` üretir, kayıt başarılıysa ana menüye döner; `Kaydetmeden Cik` doğrudan `ActionGoMainMenu` üretir; `Iptal` modalı kapatır.

Undo/redo edit mode içinde `editUndoStack` / `editRedoStack` ile tutulur. Settlement işlemleri yalnızca etkilenen region'ların `settlements[]` snapshot'ını alır; region center değişiklikleri sadece eski/yeni `world_x/world_y`, owner/terrain/type/name/lock/unlock değişiklikleri ilgili alan snapshot'ını tutar. Neighbor sync etkilenen tüm region `neighbors[]` listelerini snapshot'lar; region ekleme/silme, shape paint commit'i ve ordu/donanma ekleme-silme/birim sayısı değişiklikleri region map, order, başlangıç orduları ve `ShapeData` için dünya snapshot'ı kullanır; geniş veri editörü faction/army alanları için küçük alan command'leri üretir. `Ctrl+Z` undo, `Ctrl+Y` veya `Ctrl+Shift+Z` redo üretir; drag işlemleri command'i frame frame değil mouse bırakıldığında tek kez push eder.

Menü ve üst paneller fareyle tamamlanabilir: senaryo/fraksiyon/zafer ve kayıt ekranlarında `Geri` düğmesi vardır; diplomasi ve teknoloji panelleri X düğmesiyle kapanır; kayıt silme onayı kart içi `Sil`/`İptal` düğmeleriyle yapılır. Save/load kartında silme onayı açıkken slot adı üst bantta daha küçük çizilir ve onay sorusu ayrı satıra alınır; böylece başlık ile `Silinecek! Emin misiniz?` metni üst üste binmez. Ayarlar ekranında müzik/ses efektleri aç-kapat ve her ikisi için `0-100` arası ayrı seviye bulunur. Paylaşılan efektler `assets/sounds/` altından yüklenir; senaryo müziği `scenario.json` içindeki `music.default_playlist` ile başlar ve dosyaları senaryo `musics/` klasöründen okur. Oyun içi müzik HUD'u aktif parçayı gösterir ve `Dur/Cal` ile `Sonr` kontrollerini sunar; ESC menüsünde müzik aç/kapat ve müzik seviyesi hızlıca değiştirilebilir.

Harita modu anahtarı alt-orta aksiyon HUD'unun üstündeki `Normal | Ticaret` segmentinde yer alır. `M` kısayolu veya bu segment ile mod değişir. Ticaret koridor çizimi yalnızca `Ticaret` modunda render edilir; `Normal` modda ticaret çizgileri tamamen gizlidir.

`Ticaret` modunda harita üstüne hafif desatüre/sisli bir overlay eklenir ve çizim tüm fraksiyon çiftleri arasında birebir mesh yerine `ticaret merkezi` odaklı yapılır: merkez düğümleri senaryo bazlı `data/trade_centers.json` dosyasından okunur, fraksiyonlar en yakın merkeze ince spoke ile bağlanır, ana ağ ise merkezler arası kavisli bezier glow/core koridorlar olarak çizilir. Merkezler arası akış doğrudan her çift arasında çizilmez; senaryoda tanımlı `links` graph'ı üzerindeki kısa yol boyunca dağıtılır (ör. Halep -> Konstantinopolis -> Venedik). Çizim sırası deterministik tutulduğu için frame-frame titreme/yanıp sönme engellenir.

Trade overlay görünürlüğü `Renderer.tradeOverlayVisible()` ile merkezileştirilmiştir. Yani `MapModeTrade` seçili kalsa bile teknoloji/diplomasi/ticaret overlay'i, confirm modalı, event detail/codex, historical event veya gelen diplomasi teklifi gibi üst ekranlar açıkken trade backdrop, koridorlar, merkez tabelaları ve hover tooltip'i hiç çizilmez; overlay kapanınca trade görünümü aynı mod state'iyle geri gelir.

Trade overlay ayrıca HUD rect'lerini de saygılar. Alt aksiyon HUD'u, map-mode/Pazar butonları, üst durum-tarih-müzik kartları, turn-tech kartı, event log, minimap ve açık bilgi panellerinin üstünden geçen trade segmentleri çizim ve hover hit-test'ten düşürülür; böylece yarı saydam HUD altında koridor parlama/tooltip sızması olmaz.

Ticaret koridorları etkileşimlidir: koridor üzerine hover yapıldığında koridor focus moduna geçilir (arka ağ karartılır, seçili hat parlatılır) ve tooltip `merkez A ↔ merkez B`, `hacim/tur`, `bağlı fraksiyon` ve baskın emtia özeti gösterir; sol tık aynı bilgiyi kısa bildirim olarak yazar. Ticaret panel overlay'i artık viewport'a göre ölçeklenir; 4K hedefinde sekmeler, filtre/sıralama satırı, iki kolonlu liste alanı ve al/sat kartı sabit küçük pencere gibi kalmaz. `Yeni Rota` sekmesi gerçek anlaşma adaylarını ve engel nedenlerini gösterir; manuel al/sat akışı ayrı `Pazar` sekmesine taşınmıştır ve yalnız aktif ticaret ağına bağlı partnerleri listeler. Pazar listeleri tıklanan satırı press anında resolve eder ve trade panel input'u tam mouse state (`LeftPressed` / `LeftJustPressed` / `LeftJustReleased`) ile beslendiği için fraksiyon ve mal seçimleri release-state kaybetmeden çalışır. Geometri yalnız primitive button/list view üzerinde değil, `internal/ui/box.go` içindeki ortak cut/split layout akışı üzerinden draw ve hit-test tarafında aynı slotlardan türetilir. Aynı slot temelli yaklaşım diplomasi ve teknoloji overlay panellerine de uygulanmıştır; liste/teklif footer'ları, aktif araştırma satırı ve close alanları ortak rect akışından beslenir. Tam ekran menü/seçim yüzeylerinde de `internal/render/screen_layouts.go` centered stack/grid helper'ları kullanılır; pause, save/load, scenario, faction, victory ve settings ekranları aynı merkezleme ve aralık kurallarıyla yerleşir. Aynı ekran ailesinin chrome ve kart yüzeyi de `internal/render/ui_compose.go` altındaki ortak helper'larla çizilir; üst/alt bant, başlık, kart bordürü ve accent şeridi ekran bazında tekrar yazılmaz.
Pazar sekmesinin kontrol satırındaki `Oto. İhracat: AÇIK/KAPALI` butonu aynı ticaret paneli rect/hit-test akışıyla `ActionToggleAutoGrainExport` üretir.
Pazar mal listesinde tahıl seçiliyken hedef fraksiyonun pozitif `StrategicGrainDemand` değeri `İthalat ihtiyacı` etiketiyle gösterilir; değer state snapshot'ından okunur ve ayrı bir hit-test yüzeyi oluşturmaz.

Bu compose katmanı sadece tam ekran seçim yüzeylerinde kalmaz; oyun içi HUD/panel ailesi de aynı helper'lara bağlanır. `DrawBottomPanel`, event log, date/music/turn-tech mini kartları, recruit paneli ve region/army/sea/settlement bilgi panelleri ortak `drawUIPanelFrame`, `drawUICardRect` ve `drawUISeparator` çizimlerini kullanır. Böylece panel sınırı, üst accent çizgisi ve section separator'ları birden fazla dosyada elle tekrar edilmez.

Merkez odak modu: Ticaret merkezi düğümlerinden birinin üzerine hover yapıldığında yalnız o merkeze bağlı koridorlar belirgin tutulur, diğer ağ düşük alpha ile geri plana atılır. Merkeze tıklama, bağlı koridor sayısı ve toplam hacim özetini verir.

Üst-sol durum paneli `internal/render/panel.go:185` içinde çizilir. Sağ taraftaki zafer hedefi ve askeri kapasite alanları, sabit ölçülü iç kartlar ve kendi ilerleme barı çizimiyle sınırlandırılır; böylece zafer barı askeri kapasite ayırıcısına veya panel sağ sınırına taşmaz.

Uzun sürebilen senaryo/kayıt yükleme işleri `PhaseLoading` ekranına geçer. `internal/game/game.go` yükleme işini arka planda başlatır; renderer bu sırada `loading.go` içindeki gerçek zaman tabanlı spinner'ı ve yüzde/progress bar'ı çizer. Senaryo yüklemesi `faction_select` veya `victory_select` gibi harita gerektirmeyen ekranlara düşüyorsa `Renderer.ReloadGameState` artık `WorldMap`'i hemen kurmaz; harita cache'i yalnızca `player_turn`, `ai_turn`, `resolution`, `pause_menu` veya `edit_mode` fazında ilk gerçekten ihtiyaç olduğunda lazy-init edilir. Ayrıca zafer koşulu seçimi sonrası ve kayıt yükleme akışında ağır `WorldMap` raster/veri hazırlığı arka planda `render.PrepareWorldMap` ile tamamlanır; ana thread yalnızca son `ebiten.Image` oluşturma + `WritePixels` upload'unu yapar. Böylece loader'dan oyuncu turuna geçişteki blokaj belirgin biçimde azalır.

Rakip orduları seçilebilir ama emir verilemez. Renderer rakip ordusu için hareket hedefi çizmez ve sağ/sol tık hareket aksiyonu üretmez. Oyuncu ordularından birinin mevcut hareket menzilindeki rakip ordularda ikon birim sayısını gösterir; detay panelinde birimlerin yaklaşık yarısı görünür, kalanları `Gizli` kartlarıyla saklanır. Menzil dışındaki rakip ordularda birim sayısı ve hareket/birim detayları gizli kalır.

Bina ve birim kartlarında hover tooltip vardır. Tooltip maliyet, gereksinim, temel etki/istatistik ve kart görselini gösterir. Bina maliyetlerinde mevcut/gerekli kaynak formatı korunur; recruit birim popup'ında yalnız gerekli miktar gösterilir ve kaynak yetersizse satırın sonuna `eksik` uyarısı eklenir. Bölgeye uygun olmayan bina kartları render edilmez; liman son sıradadır ve kıyı olmayan bölgelerde görünmez. Ordu detay paneli dost toprakta toparlanan hasarlı birimler için başlıkta kısa durum metni ve kart sağ üstünde küçük `+` rozeti gösterir.

Bölge bilgi panelinde parmak imleci panelin tamamında değil, yalnızca kapatma düğmesi, vergi `-/+` düğmeleri, bina/olay sekmeleri, inşa edilebilir bina kartları ve olay satırları üzerinde gösterilir. Oyun içi HUD/panel cursor davranışı gerçek etkileşim alanlarına bağlıdır: sağ üstte yalnızca `Menü`, alt HUD'da yalnızca üç aksiyon butonu, olay logunda toggle/kart/X, birim panelinde yalnızca birim kartları pointer üretir. Boş panel alanları tıklamayı tüketmeye devam eder ama clickable cursor üretmez. Aynı panelde kaynak satırı ile memnuniyet/vergi çubukları artık ortak `regionPanelStatRowGap` ve `regionPanelBarYOffset` sabitleriyle yerleşir; böylece `Tahıl` satırı ile memnuniyet göstergesi üst üste binmez ve vergi düğmeleri de aynı geometriye bağlı kalır.

Kara bölge panelinde `BİNALAR` ve `OLAYLAR` sekmeleri aynı içerik alanını paylaşır. `BİNALAR` bina kartlarını ve bina hover tooltip'lerini, `OLAYLAR` ise mevcut aktif olaylar ve `Komşu Bölgeler` listesini panel-local scroll viewport'unda gösterir; olay satırına tıklama mevcut detay popup'ını açar, komşu başlığındaki `[Daralt] / [Tümünü Göster]` kontrolü liste görünümünü değiştirir ve wheel yalnız olaylar sekmesinin viewport'unda tüketilir. OLAYLAR sekmesinde bina kartları çizilmediği için bina hover tooltip'i de kapalıdır. Deniz komşuları mavi tonla ayrıştırılır. Aksiyon bandının çizim ve hit-test başlangıcı aktif sekmenin içerik yüksekliğinden türetilir; böylece `Diplomasi` veya `Tahıl Yardımı` düğmesi iki görünümde de paneldeki gerçek konumuyla tıklanabilir (`internal/render/panel.go`, `internal/render/renderer_input.go`, `internal/render/hover_tooltip.go`).

Bina kartı görselleri slot içine daha sıkı fit edilir (`innerW-2`, `spriteH-2`) ve inşa edilmiş kartların `Lv` etiketi koyu zemin + açık metinle çizilir; parlak sprite arka planlarında okunabilirlik korunur (`internal/render/panel.go:1689` civarı). Bina/birim kart gövdeleri beyaz tabanlı palette taşınmıştır; hover tooltip kutusu eski koyu panel temasını korur. Tooltip içinde yalnız ikon alanı beyaz kart stili çerçeve ile çizilir (`internal/render/hover_tooltip.go`). Bina hover/hit-test ölçüleri kart çizim ölçüleriyle eşitlenmiştir (`spriteH=76`, `nameH=18`, `rowH=+7`) ve dev mode komşu listesi satır yüksekliği `buildingGridStartY` hesabına dahil edilmiştir (`internal/render/panel.go`).

---

## Bildirim Sistemi

```
ShowCombatResult(msg)          → combatLogTimer = 180 frame (~3 sn), ayrı info popup; eventLog'a eklemez
ShowHistoricalEvent(title,desc,prompt,choices) → tam ekran popup; `choices` boşsa bilgi modalı, doluysa karar modalı
AddEvent(msg)                  → sağ olay logundaki kalıcı kart listesine ekler
```

`eventLog` maksimum 8 girdi tutar; yeniler öne eklenir, sondan taşanlar düşer.

---

## Ordu İkon Sistemi

Aynı bölgede birden fazla ordu bulunabilir. `armyIconPositions()` (`renderer.go`) kara ordularını `RegionID`/yerleşim anchor'ına göre, donanmaları ise sadece `docked_region_id` / `docked_settlement_id` doluyken bağlı liman anchor'ında; aksi halde deniz bölgesi anchor'ında gruplar. Dock state varsa renderer bunu doğrudan kullanır. Her grubun ikonlarını 26px aralıklarla yatayda ortalar. Hem `drawArmies` hem `handleLeftClick` hem de `cursor.go:inGameHovering` bu tek fonksiyonu kullanır — tutarsızlık riski yoktur.

```
Tek ordu  →  bölge merkezinde
İki ordu  →  merkez ±13px
Üç ordu   →  merkez -26px, 0px, +26px
```

---

## Minimap — Ordu Konumları

`panel.go:drawMinimapArmies` bölge sahiplik noktaları yerine orduların konumlarını çizer. Her ordu fraksiyon rengiyle dolu bir daire + ortada beyaz nokta olarak gösterilir; oyuncunun orduları altın kenarlıkla ayrışır.

---

## Ordu Bölme (Split)

`army_panel.go:DrawArmyDetailPanel` seçili ordunun panel başlığında "✂ BÖL" butonu gösterir (≥2 birim şartı). `SplitButtonHitTest()` hit-test fonksiyonu `renderer.go` ve `cursor.go` tarafından kullanılır. Buton tıklandığında `ActionSplitArmy` üretilir; `game.go:splitArmy()` birimleri ikiye böler ve yeni ordu oluşturur.

---

## İmleç Yönetimi (`cursor.go`)

`updateCursorShape()` her frame çalışır. Aşağıdaki fazlarda parmak imleci gösterilir:

| Faz | Koşul |
|---|---|
| PhaseMainMenu | `mainMenuHoverIndex >= 0` |
| PhaseFactionSelect | `factionCardHoverIndex >= 0` |
| PhaseVictorySelect | `victoryCardHoverIndex >= 0` |
| PhasePlayerTurn | `inGameHovering` (butonlar, ordu ikonları, BÖL butonu) |
| PhasePauseMenu | `pauseMenuHoverIndex >= 0` |
| PhaseLoadSelect / PhaseSaveSelect | `slotHoverIndex >= 0` |
| PhaseSettings | `settingsHovering` |

---

## Dosya Sorumlulukları

| Dosya | İçerik |
|---|---|
| `renderer.go` | Renderer state'i, yaşam döngüsü, kamera, draw orkestrasyonu ve dünya ikonları |
| `renderer_input.go` | `HandleInput`, ana input yönlendirme, dünya seçimleri ve kamera kontrolü |
| `renderer_dialogs.go` | Kuşatma/savaş planı, savaş ilanı, diplomasi teklifi, tarihsel olay ve confirm modal akışları |
| `map_editor.go` | Edit mode HUD/input, undo-redo, region/settlement/faction/army düzenleme ve snapshot yardımcıları |
| `trade_overlay.go` | Ticaret merkezi/koridor modeli, rota çizimi, hover ve hit-test akışı |
| `mapgen.go` | WorldMap cache, poligon doldurma |
| `tile.go` | Arazi renk/doku katmanı |
| `panel.go` | Alt bar, bölge/ordu/minimap/event log panelleri; event log kaydırma geometrisi; minimap'te ordu konumları |
| `army_panel.go` | Ordu detay paneli — 20 slot ızgara, HP çubuğu, BÖL butonu |
| `diplom.go` | Diplomasi paneli UI + input |
| `tech_panel.go` | Teknoloji ağacı paneli + input |
| `pause_menu.go` | Oyun içi duraklama menüsü (ESC) |
| `load_select.go` | Kayıt slot seçim ekranı (yükleme + kaydetme + silme) |
| `recruit_panel.go` | Birlik alım paneli |
| `action.go` | `InputAction` ve `ActionKind` tanımları |
| `font.go` | Font yükleme, `DrawText`, `MeasureText` |
| `assets.go` | Görsel varlık yükleme |
| `cursor.go` | İmleç şekli yönetimi (tüm fazlar) |
| `faction_select.go` | Fraksiyon seçim ekranı |
| `victory_select.go` | Zafer koşulu seçim ekranı |
| `main_menu.go` | Ana menü ("Devam et" → en yeni `autosave`/`quicksave`, "Kayıttan Yükle" → slot seçim ekranı) |
| `settings.go` | Ayarlar ekranı |
