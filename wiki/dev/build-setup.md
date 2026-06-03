---
type: dev
tags: [build, wsl, windows, ebiten]
last_updated: 2026-06-03
related: [architecture/render-pipeline, dev/progress, HOME]
---

# Build Kurulumu

**Kaynak:** `README.md`, `cmd/game/main.go`

## WSL Geliştirme Ortamı

Ebitengine Linux tarafında pencere, GL ve ses bağımlılıkları ister. Ubuntu/WSL içinde en pratik kurulum:

```bash
sudo apt update
sudo apt install -y \
  build-essential pkg-config libasound2-dev \
  libx11-dev libxrandr-dev libxcursor-dev libxinerama-dev \
  libxi-dev libxxf86vm-dev libgl1-mesa-dev
```

Notlar:
- WSLg varsa pencereyi doğrudan açabilirsin.
- Saf headless doğrulama için `go test ./...` yeterlidir; render testleri fiziksel pencere istemez.
- Ses cihazı hatalarında önce `libasound2-dev` eksik mi kontrol et.

## Çalıştırma

```bash
go run ./cmd/game
```

## Windows Çıktısı

Kalıcı binary hedefi:

```bash
mkdir -p bin
GOOS=windows GOARCH=amd64 go build -o bin/game.exe ./cmd/game
```

İstersen WSL içinden de Windows binary üretilebilir; çalıştırma Windows tarafında yapılır.

## Doğrulama Akışı

Önerilen sırayla:

```bash
go test ./...
GOOS=windows GOARCH=amd64 go build -o bin/game.exe ./cmd/game
```

Beklenen sonuç:
- testler temiz geçer
- dağıtılacak binary `bin/game.exe` olur
- kök dizindeki geçici `game.exe` zorunlu çıktı sayılmaz
