package ui

// Layer, ekran üzerindeki bir UI yüzeyinin input z-order bilgisini taşır.
// Katmanlar çizim sırasıyla eklenir; son eklenen görünür katman en üsttedir.
type Layer struct {
	ID           string
	Rect         Rect
	Visible      bool
	CaptureInput bool
}

// LayerStack, panel/modal/popup/HUD gibi farklı UI ailelerinin ortak input
// sınırıdır. Bir katman koordinatı kaplıyorsa, kendi kontrolü tıklanabilir
// olmasa bile altındaki katmanlara input aktarımı durdurulabilir.
type LayerStack struct {
	layers []Layer
}

// NewLayerStack önceden kapasite ayırarak frame başına katman yenilemesinde
// gereksiz allocation oluşmasını önler.
func NewLayerStack(capacity int) *LayerStack {
	if capacity < 0 {
		capacity = 0
	}
	return &LayerStack{layers: make([]Layer, 0, capacity)}
}

// Reset mevcut katmanları koruyarak yalnızca görünür frame listesini temizler.
func (s *LayerStack) Reset() {
	if s == nil {
		return
	}
	s.layers = s.layers[:0]
}

// Add katmanı mevcut çizim sırasının sonuna ekler.
func (s *LayerStack) Add(layer Layer) {
	if s == nil || !layer.Visible || !layer.CaptureInput {
		return
	}
	s.layers = append(s.layers, layer)
}

// AddRect, çoğu panel için kullanılan kısa katman ekleme yoludur.
func (s *LayerStack) AddRect(id string, rect Rect) {
	s.Add(Layer{ID: id, Rect: rect, Visible: true, CaptureInput: true})
}

// AddScreen, modal/menü/seçim ekranı gibi tam ekran input yüzeyi ekler.
func (s *LayerStack) AddScreen(id string, screenW, screenH float64) {
	s.AddRect(id, Rect{W: screenW, H: screenH})
}

// TopAt koordinata denk gelen en üst görünür katmanı döndürür.
func (s *LayerStack) TopAt(mx, my float64) (Layer, bool) {
	if s == nil {
		return Layer{}, false
	}
	for i := len(s.layers) - 1; i >= 0; i-- {
		layer := s.layers[i]
		if layer.Rect.Hit(mx, my) {
			return layer, true
		}
	}
	return Layer{}, false
}

// BlocksAt, koordinatın herhangi bir üst UI yüzeyi tarafından tüketildiğini
// bildirir. Renderer harita tıklaması ve harita hover hesaplarında bunu kullanır.
func (s *LayerStack) BlocksAt(mx, my float64) bool {
	_, ok := s.TopAt(mx, my)
	return ok
}

// Len test ve tanı amaçlı görünür katman sayısını verir.
func (s *LayerStack) Len() int {
	if s == nil {
		return 0
	}
	return len(s.layers)
}
