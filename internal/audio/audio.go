package audio

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

var (
	audioContext         *audio.Context
	soundCache           map[string][]byte
	globalSoundPlayers   map[string]*audio.Player
	scenarioSoundCache   map[string][]byte
	scenarioSoundPlayers map[string]*audio.Player
	soundEnabled         = true
	soundVolume          = 0.35
	soundGains           = map[string]float64{
		"combat":            0.1,
		"battle_land":       0.1,
		"battle_naval":      0.1,
		"battle_amphibious": 0.1,
		"battle_siege":      0.1,
	}

	musicEnabled      = true
	musicVolume       = 0.45
	musicBaseDir      string
	musicPlaylist     []MusicTrack
	musicPlayer       *audio.Player
	musicCurrentIndex = -1
	musicUnavailable  bool
	musicRandom       = rand.New(rand.NewSource(time.Now().UnixNano()))
	zoomInMusicDucked = false
)

const zoomInMusicVolumeFactor = 0.35

// MusicTrack points to a file under a scenario's musics/ folder.
type MusicTrack struct {
	File   string
	Weight int
}

type MusicStatus struct {
	HasPlaylist bool
	Playing     bool
	Track       string
	Volume      int
	Enabled     bool
}

func init() {
	soundCache = make(map[string][]byte)
	globalSoundPlayers = make(map[string]*audio.Player)
	scenarioSoundCache = make(map[string][]byte)
	scenarioSoundPlayers = make(map[string]*audio.Player)
	audioContext = audio.NewContext(sampleRate)
}

// LoadGlobalSounds clears old sounds and loads all shared WAV/MP3 effects.
func LoadGlobalSounds(soundsDir string) {
	// Clear old cache
	soundCache = make(map[string][]byte)

	entries, err := os.ReadDir(soundsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".wav" && ext != ".mp3" {
			continue
		}

		path := filepath.Join(soundsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Ses dosyası okunamadı %s: %v", path, err)
			continue
		}

		var s io.Reader
		var decodeErr error
		switch ext {
		case ".wav":
			// Decode WAV to PCM.
			s, decodeErr = wav.DecodeWithoutResampling(bytes.NewReader(data))
		case ".mp3":
			// Global menu efektleri de senaryo efektleri gibi MP3 olabilir.
			s, decodeErr = mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		}
		if decodeErr != nil {
			log.Printf("Ses decode hatası %s: %v", path, decodeErr)
			continue
		}

		pcmData, err := io.ReadAll(s)
		if err != nil {
			log.Printf("PCM okuma hatası %s: %v", path, err)
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		soundCache[name] = pcmData
	}
}

func SetSoundEnabled(enabled bool) {
	soundEnabled = enabled
	if !enabled {
		StopZoomInLoop()
	}
}

func SetSoundVolume(percent int) {
	soundVolume = percentToVolume(percent)
}

// PlaySound çalınacak sesin adını (ör. "click") alır ve varsa çalar.
func PlaySound(name string) {
	if !soundEnabled || soundVolume <= 0 {
		return
	}
	pcmData, ok := soundCache[name]
	if !ok {
		return
	}

	player := audioContext.NewPlayerFromBytes(pcmData)
	player.SetVolume(soundVolume * soundGain(name))
	player.Play()
}

// PlayGlobalSound, bitene kadar takip edilmesi gereken global bir MP3/WAV
// efektini çalar. Menü açılış sesi gibi efektler için PlaySound'dan ayrıdır.
func PlayGlobalSound(name string) {
	playGlobalSound(name, false)
}

// PlayGlobalSoundLoop, takip edilen global MP3/WAV efektini kesintisiz döngüde
// çalar. Oyun giriş müziği gibi menü boyunca sürmesi gereken sesler içindir.
func PlayGlobalSoundLoop(name string) {
	playGlobalSound(name, true)
}

func playGlobalSound(name string, loop bool) {
	if !soundEnabled || soundVolume <= 0 {
		return
	}
	pcmData, ok := soundCache[name]
	if !ok {
		return
	}
	if previous := globalSoundPlayers[name]; previous != nil {
		_ = previous.Close()
	}
	var player *audio.Player
	if loop {
		var err error
		player, err = audioContext.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(pcmData), int64(len(pcmData))))
		if err != nil {
			return
		}
	} else {
		player = audioContext.NewPlayerFromBytes(pcmData)
	}
	player.SetVolume(soundVolume * soundGain(name))
	globalSoundPlayers[name] = player
	player.Play()
}

// StopGlobalSound, takip edilen global efektin oynatımını durdurur.
func StopGlobalSound(name string) {
	if player := globalSoundPlayers[name]; player != nil {
		_ = player.Close()
		delete(globalSoundPlayers, name)
	}
}

// StartZoomInLoop, haritanın son zoom seviyelerinde zoom sesini döngüde çalar
// ve arka plan müziğini geçici olarak kısar.
func StartZoomInLoop() {
	if !soundEnabled || !musicEnabled || soundVolume <= 0 {
		return
	}
	if player := globalSoundPlayers["zoom_in"]; player == nil || !player.IsPlaying() {
		PlayGlobalSoundLoop("zoom_in")
	}
	if globalSoundPlayers["zoom_in"] == nil {
		return
	}
	zoomInMusicDucked = true
	applyMusicVolume()
}

// StopZoomInLoop, son zoom seviyelerinden çıkıldığında efekt döngüsünü kapatır
// ve müziğin kullanıcı tarafından seçilen seviyesini geri yükler.
func StopZoomInLoop() {
	StopGlobalSound("zoom_in")
	if !zoomInMusicDucked {
		return
	}
	zoomInMusicDucked = false
	applyMusicVolume()
}

func HasSound(name string) bool {
	_, ok := soundCache[name]
	return ok
}

// PlayScenarioSound, aktif senaryonun audio klasöründeki kısa MP3 efektlerini
// cache'leyip çalar. Senaryo sesleri global WAV efektlerinden ayrı tutulur.
func PlayScenarioSound(audioDir, name string) {
	playScenarioSound(audioDir, name, false)
}

// PlayScenarioSoundLoop, aktif senaryonun sesini kesintisiz döngüde çalar.
// Senaryo seçimi ile oyun başlangıcı arasındaki intro gibi uzun akışlar için
// kullanılır.
func PlayScenarioSoundLoop(audioDir, name string) {
	playScenarioSound(audioDir, name, true)
}

func playScenarioSound(audioDir, name string, loop bool) {
	if !soundEnabled || soundVolume <= 0 || audioDir == "" || name == "" {
		return
	}
	pcmData, ok := loadScenarioSound(audioDir, name)
	if !ok {
		return
	}
	path := filepath.Join(audioDir, name+".mp3")
	if previous := scenarioSoundPlayers[path]; previous != nil {
		_ = previous.Close()
	}
	var player *audio.Player
	if loop {
		var err error
		player, err = audioContext.NewPlayer(audio.NewInfiniteLoop(bytes.NewReader(pcmData), int64(len(pcmData))))
		if err != nil {
			return
		}
	} else {
		player = audioContext.NewPlayerFromBytes(pcmData)
	}
	player.SetVolume(soundVolume * soundGain(name))
	scenarioSoundPlayers[path] = player
	player.Play()
}

// EnsureScenarioSoundPlaying, aynı ses zaten oynuyorsa onu yeniden başlatmaz.
// Çok segmentli hareket animasyonlarında sesin ara noktalarda kesilmemesini
// sağlar; ses doğal olarak bittiyse yeni oynatıcı oluşturur.
func EnsureScenarioSoundPlaying(audioDir, name string) {
	if !soundEnabled || soundVolume <= 0 || audioDir == "" || name == "" {
		return
	}
	path := filepath.Join(audioDir, name+".mp3")
	if player := scenarioSoundPlayers[path]; player != nil && player.IsPlaying() {
		return
	}
	PlayScenarioSound(audioDir, name)
}

// PreloadScenarioSounds, senaryo yüklenirken kullanılan kısa ses efektlerini
// okuyup MP3'ten PCM'e decode ederek cache'ler. Bulunmayan dosyalar sessizce
// atlanır; böylece eski senaryolar etkilenmez.
func PreloadScenarioSounds(audioDir string, names []string) {
	if audioDir == "" {
		return
	}
	for _, name := range names {
		_, _ = loadScenarioSound(audioDir, name)
	}
}

func loadScenarioSound(audioDir, name string) ([]byte, bool) {
	if audioDir == "" || name == "" {
		return nil, false
	}
	path := filepath.Join(audioDir, name+".mp3")
	if pcmData, ok := scenarioSoundCache[path]; ok {
		return pcmData, true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	stream, err := mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	pcmData, err := io.ReadAll(stream)
	if err != nil {
		return nil, false
	}
	scenarioSoundCache[path] = pcmData
	return pcmData, true
}

// StopScenarioSound, senaryo ses efektinin aktif oynatıcısını durdurur.
func StopScenarioSound(audioDir, name string) {
	if audioDir == "" || name == "" {
		return
	}
	path := filepath.Join(audioDir, name+".mp3")
	if player := scenarioSoundPlayers[path]; player != nil {
		_ = player.Close()
		delete(scenarioSoundPlayers, path)
	}
}

func soundGain(name string) float64 {
	if gain, ok := soundGains[name]; ok && gain >= 0 {
		return gain
	}
	return 1
}

// StartMusicPlaylist starts the given scenario playlist. Missing or empty playlists are silent.
func StartMusicPlaylist(baseDir string, tracks []MusicTrack) {
	StopMusic()
	musicBaseDir = baseDir
	musicPlaylist = tracks
	musicCurrentIndex = -1
	musicUnavailable = false
	if musicEnabled && musicVolume > 0 {
		playNextMusic()
	}
}

func StopMusic() {
	StopZoomInLoop()
	if musicPlayer != nil {
		_ = musicPlayer.Close()
		musicPlayer = nil
	}
	musicPlaylist = nil
	musicBaseDir = ""
	musicCurrentIndex = -1
	musicUnavailable = false
}

func SetMusicEnabled(enabled bool) {
	musicEnabled = enabled
	if !enabled {
		StopZoomInLoop()
	}
	if musicPlayer == nil {
		if enabled && musicVolume > 0 && len(musicPlaylist) > 0 {
			playNextMusic()
		}
		return
	}
	if !enabled {
		musicPlayer.Pause()
		return
	}
	musicPlayer.SetVolume(musicPlaybackVolume())
	musicPlayer.Play()
}

func ToggleMusic() bool {
	// HUD düğmesi gerçek oynatma durumunu gösterir. Oynatıcı, dosya yükleme
	// hatası veya başka bir akış nedeniyle nil/bitmiş olsa bile musicEnabled
	// true kalabilir; bu durumda düğmeye basmak müziği kapatmak yerine yeniden
	// başlatmayı denemelidir.
	SetMusicEnabled(!MusicStatusNow().Playing)
	return musicEnabled
}

func SetMusicVolume(percent int) {
	musicVolume = percentToVolume(percent)
	if musicPlayer != nil {
		musicPlayer.SetVolume(musicPlaybackVolume())
		if musicEnabled && musicVolume > 0 {
			musicPlayer.Play()
		}
	}
}

func AdjustMusicVolume(delta int) int {
	percent := int(musicVolume*100 + 0.5)
	SetMusicVolume(percent + delta)
	return int(musicVolume*100 + 0.5)
}

func MusicStatusNow() MusicStatus {
	status := MusicStatus{
		HasPlaylist: len(musicPlaylist) > 0,
		Enabled:     musicEnabled,
		Volume:      int(musicVolume*100 + 0.5),
	}
	if musicCurrentIndex >= 0 && musicCurrentIndex < len(musicPlaylist) {
		status.Track = musicPlaylist[musicCurrentIndex].File
	}
	if musicPlayer != nil {
		status.Playing = musicEnabled && musicPlayer.IsPlaying()
	}
	return status
}

func NextMusic() {
	musicUnavailable = false
	// Sonraki düğmesi de müziği başlatan bir kullanıcı eylemidir. Müzik daha
	// önce durdurulmuşsa playNextMusic tek başına oynatıcıyı başlatır ancak
	// musicEnabled değerini güncellemez; bu da HUD ikonunu "Çal" durumunda
	// bırakır.
	musicEnabled = true
	playNextMusic()
}

// UpdateMusic advances the scenario playlist when the current track ends.
func UpdateMusic() {
	if !musicEnabled || musicVolume <= 0 || len(musicPlaylist) == 0 || musicUnavailable {
		return
	}
	if musicPlayer == nil || !musicPlayer.IsPlaying() {
		playNextMusic()
	}
}

func musicPlaybackVolume() float64 {
	if zoomInMusicDucked {
		return musicVolume * zoomInMusicVolumeFactor
	}
	return musicVolume
}

func applyMusicVolume() {
	if musicPlayer != nil {
		musicPlayer.SetVolume(musicPlaybackVolume())
	}
}

func playNextMusic() {
	if len(musicPlaylist) == 0 || musicBaseDir == "" {
		return
	}
	if musicPlayer != nil {
		_ = musicPlayer.Close()
		musicPlayer = nil
	}
	for attempts := 0; attempts < len(musicPlaylist); attempts++ {
		next := chooseMusicIndex()
		if next < 0 {
			musicUnavailable = true
			return
		}
		track := musicPlaylist[next]
		path := filepath.Join(musicBaseDir, track.File)
		player, err := newMusicPlayer(path)
		if err != nil {
			log.Printf("Müzik yüklenemedi %s: %v", path, err)
			musicCurrentIndex = next
			continue
		}
		musicCurrentIndex = next
		musicPlayer = player
		musicPlayer.SetVolume(musicPlaybackVolume())
		musicPlayer.Play()
		return
	}
	musicUnavailable = true
}

func chooseMusicIndex() int {
	total := 0
	for _, track := range musicPlaylist {
		if track.File == "" {
			continue
		}
		weight := track.Weight
		if weight <= 0 {
			weight = 1
		}
		total += weight
	}
	if total <= 0 {
		return -1
	}
	pick := musicRandom.Intn(total)
	for i, track := range musicPlaylist {
		if track.File == "" {
			continue
		}
		weight := track.Weight
		if weight <= 0 {
			weight = 1
		}
		if pick < weight {
			if len(musicPlaylist) > 1 && i == musicCurrentIndex {
				return (i + 1) % len(musicPlaylist)
			}
			return i
		}
		pick -= weight
	}
	return -1
}

func newMusicPlayer(path string) (*audio.Player, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	var stream io.Reader
	switch ext {
	case ".ogg":
		stream, err = vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	case ".mp3":
		stream, err = mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	case ".wav":
		stream, err = wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	default:
		stream, err = vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	}
	if err != nil {
		return nil, err
	}

	// Müzik stream'ini doğrudan oto oynatıcısına vermek decoder hatalarını
	// oynatma sırasında ve dosya yolu olmadan raporlatabiliyordu. Özellikle
	// go-mp3 MPEG 2.5 frame'lerini desteklemediği için bozuk/uyumsuz bir parça
	// seçildiğinde hata ancak parça okunurken ortaya çıkıyordu. Tamamını burada
	// PCM'e çevirerek hatalı parçayı playlist'e girmeden yakala.
	pcmData, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}
	if len(pcmData) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return audioContext.NewPlayerFromBytes(pcmData), nil
}

func percentToVolume(percent int) float64 {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return float64(percent) / 100
}
