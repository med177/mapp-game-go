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
	audioContext *audio.Context
	soundCache   map[string][]byte
	soundEnabled = true
	soundVolume  = 0.35

	musicEnabled      = true
	musicVolume       = 0.45
	musicBaseDir      string
	musicPlaylist     []MusicTrack
	musicPlayer       *audio.Player
	musicCurrentIndex = -1
	musicUnavailable  bool
	musicRandom       = rand.New(rand.NewSource(time.Now().UnixNano()))
)

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
	audioContext = audio.NewContext(sampleRate)
}

// LoadGlobalSounds clears old sounds and loads all shared .wav effects.
func LoadGlobalSounds(soundsDir string) {
	// Clear old cache
	soundCache = make(map[string][]byte)

	entries, err := os.ReadDir(soundsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".wav" {
			continue
		}

		path := filepath.Join(soundsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Ses dosyası okunamadı %s: %v", path, err)
			continue
		}

		// Decode WAV to PCM
		s, err := wav.DecodeWithoutResampling(bytes.NewReader(data))
		if err != nil {
			log.Printf("WAV decode hatası %s: %v", path, err)
			continue
		}

		pcmData, err := io.ReadAll(s)
		if err != nil {
			log.Printf("WAV okuma hatası %s: %v", path, err)
			continue
		}

		name := entry.Name()[:len(entry.Name())-4] // Remove .wav extension
		soundCache[name] = pcmData
	}
}

func SetSoundEnabled(enabled bool) {
	soundEnabled = enabled
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
	player.SetVolume(soundVolume)
	player.Play()
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
	musicPlayer.SetVolume(musicVolume)
	musicPlayer.Play()
}

func ToggleMusic() bool {
	SetMusicEnabled(!musicEnabled)
	return musicEnabled
}

func SetMusicVolume(percent int) {
	musicVolume = percentToVolume(percent)
	if musicPlayer != nil {
		musicPlayer.SetVolume(musicVolume)
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
		musicPlayer.SetVolume(musicVolume)
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
	return audioContext.NewPlayer(stream)
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
