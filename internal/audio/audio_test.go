package audio

import (
	"path/filepath"
	"testing"
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
