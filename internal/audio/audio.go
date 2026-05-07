package audio

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

var (
	audioContext *audio.Context
	soundCache   map[string][]byte
)

func init() {
	soundCache = make(map[string][]byte)
	audioContext = audio.NewContext(sampleRate)
}

// LoadScenarioSounds clears old sounds and loads all .wav files from the scenario's sounds folder
func LoadScenarioSounds(scenarioPath string) {
	// Clear old cache
	soundCache = make(map[string][]byte)

	soundsDir := filepath.Join(scenarioPath, "sounds")
	entries, err := os.ReadDir(soundsDir)
	if err != nil {
		// Normal if sounds folder doesn't exist yet
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

// PlaySound çalınacak sesin adını (ör. "click") alır ve varsa çalar.
func PlaySound(name string) {
	pcmData, ok := soundCache[name]
	if !ok {
		return
	}

	player := audioContext.NewPlayerFromBytes(pcmData)
	player.Play()
}
