# Çoklu platform derlemesi

`Build.ps1`, projeyi Windows 64-bit, Linux 64-bit, Intel Mac ve Apple Silicon
Mac için derler. Script, proje kökünden çalıştırılmayı gerektirmez ve PowerShell
7 ile hem Windows hem Linux üzerinde kullanılabilir.

## Gereksinimler

- Go 1.25 veya üzeri
- PowerShell 7 veya üzeri (`pwsh`)
- İlk derlemede Go modüllerini indirmek için internet erişimi

Ebitengine masaüstü derlemelerinde cgo gerektirmeyen sürümüyle kullanılır.
Linux'ta çalıştırma için hedef makinede uygun grafik/ses sistemi ve sistem
kütüphaneleri bulunmalıdır.

## Derleme

VS Code'da **Ctrl+Shift+B** veya **Terminal > Run Build Task...** içindeki
varsayılan **Build Game (All 4 Targets)** görevi Windows, Linux, Intel Mac ve
Apple Silicon Mac sürümlerinin tamamını üretir. Sadece tek platform için aynı
menüden ilgili hedef görevini seçebilirsiniz. MacBook Neo çıktısı için hedef
klasör `build/dist/macos-arm64/` olur.

Windows PowerShell 7'de:

```powershell
pwsh -File .\build\Build.ps1
```

Linux'ta:

```bash
pwsh -File ./build/Build.ps1
```

Parametresiz çalıştırma dört hedefin tamamını üretir. Tek hedef seçmek için:

```powershell
pwsh -File .\build\Build.ps1 -Target windows
pwsh -File .\build\Build.ps1 -Target linux
pwsh -File .\build\Build.ps1 -Target macos-amd64
pwsh -File .\build\Build.ps1 -Target macos-arm64
```

Çıktılar varsayılan olarak `build/dist/` altındaki platform klasörlerine
yazılır:

```text
build/dist/
├── assets/
├── windows-amd64/
│   └── mapp-game.exe
├── linux-amd64/
│   └── mapp-game
├── macos-amd64/
│   └── mapp-game
└── macos-arm64/
    └── mapp-game
```

Platform klasörlerinde çalıştırılabilir dosyalar, ortak `assets/` klasöründe
oyunun ihtiyaç duyduğu veriler bulunur. Birden çok platform derlemek asset
dosyalarının kopyalarını çoğaltmaz. Başka bir konum seçmek için:

```powershell
pwsh -File .\build\Build.ps1 -Target macos-arm64 -OutputDirectory C:\Mapp\dist
```

## MacBook Neo'da çalıştırma

MacBook Neo, `macos-arm64` hedefini kullanır. `build/dist/macos-arm64/`
klasörüyle `build/dist/assets/` klasörünü aynı konuma, örneğin Mac'teki bir
`Mapp` klasörüne kopyalayın; yalnızca çalıştırılabilir dosyayı kopyalamak
yeterli değildir. Mac'te Terminal'i `Mapp` klasöründe açıp çalıştırın:

```bash
./macos-arm64/mapp-game
```

Intel Mac için komutu `./macos-amd64/mapp-game` olarak değiştirin. Alternatif
olarak `build/dist/` klasörünün tamamını Mac'e kopyalayıp komutu o klasörden
çalıştırabilirsiniz. Bu çıktılar imzalı veya notarize edilmiş `.app` paketleri
değildir; Mac'teki ilk çalıştırmayı hedef makinede doğrulayın.

**Debug Game**, **Run Game** ve **Run Built Game (Windows)** başlatma profilleri
çalıştırılmadan önce dört platformun da build görevini çalıştırır. **Run Built
Game (Windows)** ardından Windows çıktısını başlatır. macOS/Linux hedefleri
Windows'ta derlenebilir; ancak bu işletim sistemlerinin çalıştırılabilir
dosyalarını Windows'ta VS Code debugger ile başlatamazsınız.

## Notlar

- Script mevcut `internal/buildinfo/generated.go` dosyasını değiştirmez.
- Çıktı paketleri yerel kayıtları (`saves/`) ve `.env` dosyasını içermez.
- Linux paketinde grafik, pencere sistemi ve ses için hedef dağıtımın runtime
  kütüphaneleri gerekir.
