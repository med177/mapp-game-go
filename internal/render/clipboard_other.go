//go:build !windows

package render

import "os/exec"

func readSystemClipboard() (string, bool) {
	commands := [][]string{
		{"xclip", "-selection", "clipboard", "-o"},
		{"xsel", "--clipboard", "--output"},
	}
	for _, command := range commands {
		output, err := exec.Command(command[0], command[1:]...).Output()
		if err == nil {
			return string(output), true
		}
	}
	return "", false
}
