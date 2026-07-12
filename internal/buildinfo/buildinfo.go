package buildinfo

//go:generate go run ../../tools/buildinfo

// Build metadata is populated by go generate into generated.go.
// Defaults keep the project buildable even if generation is skipped.
var Version = "dev"
var Commit = "unknown"
var Date = "unknown"

// SaveVersion save dosyasına yazılacak sürüm etiketini üretir.
func SaveVersion() string {
	switch {
	case Version != "" && Commit != "" && Commit != "unknown":
		if Version == Commit {
			return Version
		}
		return Version + "+" + Commit
	case Version != "":
		return Version
	case Commit != "" && Commit != "unknown":
		return Commit
	default:
		return "dev"
	}
}
