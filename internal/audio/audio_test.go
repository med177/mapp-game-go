package audio

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

func TestLoadGlobalSoundsLoadsMP3Effects(t *testing.T) {
	LoadGlobalSounds(filepath.Join("..", "..", "assets", "sounds"))

	if !HasSound("select") {
		t.Fatal("global MP3 seçim sesi yüklenmedi")
	}
	if !HasSound("entrance_intro") {
		t.Fatal("global MP3 açılış sesi yüklenmedi")
	}
	if !HasSound("info_message") {
		t.Fatal("global MP3 bilgi mesajı sesi yüklenmedi")
	}
	if !HasSound("end_turn") {
		t.Fatal("global MP3 tur bitirme sesi yüklenmedi")
	}
	if !HasSound("zoom_in") {
		t.Fatal("global MP3 yakınlaştırma sesi yüklenmedi")
	}
}

func TestLoadScenarioSoundLoadsConqueredEffect(t *testing.T) {
	audioDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "audio")
	if _, ok := loadScenarioSound(audioDir, "conquered"); !ok {
		t.Fatal("senaryo fetih sesi yüklenmedi")
	}
}

func TestZoomInLoopRespectsMusicSetting(t *testing.T) {
	oldMusicEnabled := musicEnabled
	oldSoundEnabled := soundEnabled
	oldSoundVolume := soundVolume
	oldMusicDucked := zoomInMusicDucked
	defer func() {
		musicEnabled = oldMusicEnabled
		soundEnabled = oldSoundEnabled
		soundVolume = oldSoundVolume
		zoomInMusicDucked = oldMusicDucked
		StopGlobalSound("zoom_in")
	}()

	musicEnabled = false
	soundEnabled = true
	soundVolume = 1
	StartZoomInLoop()
	if _, ok := globalSoundPlayers["zoom_in"]; ok {
		t.Fatal("müzik kapalıyken zoom sesi başlatıldı")
	}

	zoomInMusicDucked = true
	SetMusicEnabled(false)
	if zoomInMusicDucked {
		t.Fatal("müzik kapatılırken zoom sesi ducking durumu temizlenmedi")
	}
}

func TestPlayScenarioSoundLoopTracksScenarioIntro(t *testing.T) {
	audioDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "audio")
	path := filepath.Join(audioDir, "scenario_intro.mp3")
	StopScenarioSound(audioDir, "scenario_intro")

	PlayScenarioSoundLoop(audioDir, "scenario_intro")
	player := scenarioSoundPlayers[path]
	if player == nil {
		t.Fatal("senaryo intro döngü oynatıcısı oluşturulmadı")
	}
	if !player.IsPlaying() {
		t.Fatal("senaryo intro döngü oynatıcısı başlatılmadı")
	}

	StopScenarioSound(audioDir, "scenario_intro")
}

func TestNewMusicPlayerCreatesStreamingScenarioPlaylistTracks(t *testing.T) {
	musicDir := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "musics")
	entries, err := os.ReadDir(musicDir)
	if err != nil {
		t.Fatalf("senaryo müzik dizini okunamadı: %v", err)
	}
	var tracks []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".mp3") {
			tracks = append(tracks, entry.Name())
		}
	}
	if len(tracks) == 0 {
		t.Fatal("senaryo müzik dizininde test edilecek parça bulunamadı")
	}

	for _, track := range tracks {
		player, err := newMusicPlayer(filepath.Join(musicDir, track))
		if err != nil {
			t.Fatalf("playlist parçası için stream oynatıcısı oluşturulamadı (%s): %v", track, err)
		}
		if player == nil {
			t.Fatalf("playlist parçası için oynatıcı oluşturulmadı (%s)", track)
		}
		_ = player.Close()
	}
}

func TestNewMusicPlayerDefersTrackDecode(t *testing.T) {
	path := filepath.Join("..", "..", "assets", "scenarios", "1300_ottoman_rise", "musics", "the_mountain-war-opening-145222.mp3")
	player, err := newMusicPlayer(path)
	if err != nil {
		t.Fatalf("stream oynatıcısı decode tamamlanmadan oluşturulamadı: %v", err)
	}
	if player == nil {
		t.Fatal("stream oynatıcısı oluşturulmadı")
	}
	_ = player.Close()
}

func TestStartMusicPlaylistDoesNotBlockOnTrackPreparation(t *testing.T) {
	oldLoader := musicPlayerLoader
	oldMusicEnabled := musicEnabled
	oldMusicVolume := musicVolume
	var releaseOnce sync.Once
	release := make(chan struct{})
	loaderStarted := make(chan struct{})
	releaseLoader := func() {
		releaseOnce.Do(func() { close(release) })
	}
	defer func() {
		releaseLoader()
		StopMusic()
		musicPlayerLoader = oldLoader
		musicEnabled = oldMusicEnabled
		musicVolume = oldMusicVolume
	}()

	StopMusic()
	musicEnabled = true
	musicVolume = 1
	musicPlayerLoader = func(string) (*ebitenaudio.Player, error) {
		close(loaderStarted)
		<-release
		return nil, errors.New("test müzik yükleme hatası")
	}

	startReturned := make(chan struct{})
	go func() {
		StartMusicPlaylist("test-music", []MusicTrack{{File: "track.mp3"}})
		close(startReturned)
	}()

	select {
	case <-loaderStarted:
	case <-time.After(time.Second):
		t.Fatal("müzik yükleme worker'ı başlatılmadı")
	}
	select {
	case <-startReturned:
	case <-time.After(time.Second):
		t.Fatal("playlist başlatma müzik hazırlamasını bekletti")
	}

	releaseLoader()
	deadline := time.After(time.Second)
	for musicLoadPending {
		UpdateMusic()
		select {
		case <-deadline:
			t.Fatal("müzik yükleme sonucu oyun akışına dönmedi")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}
