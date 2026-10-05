//go:build windows

package render

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const clipboardUnicodeText = 13

var (
	clipboardUser32   = syscall.NewLazyDLL("user32.dll")
	clipboardKernel32 = syscall.NewLazyDLL("kernel32.dll")
	openClipboard     = clipboardUser32.NewProc("OpenClipboard")
	closeClipboard    = clipboardUser32.NewProc("CloseClipboard")
	getClipboardData  = clipboardUser32.NewProc("GetClipboardData")
	globalLock        = clipboardKernel32.NewProc("GlobalLock")
	globalUnlock      = clipboardKernel32.NewProc("GlobalUnlock")
)

func readSystemClipboard() (string, bool) {
	if result, _, _ := openClipboard.Call(0); result == 0 {
		return "", false
	}
	defer closeClipboard.Call()

	handle, _, _ := getClipboardData.Call(clipboardUnicodeText)
	if handle == 0 {
		return "", false
	}
	data, _, _ := globalLock.Call(handle)
	if data == 0 {
		return "", false
	}
	defer globalUnlock.Call(handle)

	text := make([]uint16, 0, 64)
	for i := uintptr(0); i < 1<<20; i++ {
		value := *(*uint16)(unsafe.Pointer(data + i*unsafe.Sizeof(uint16(0))))
		if value == 0 {
			break
		}
		text = append(text, value)
	}
	return string(utf16.Decode(text)), true
}
