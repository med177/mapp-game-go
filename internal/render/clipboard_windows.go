//go:build windows

package render

import (
	"runtime"
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
	globalSize        = clipboardKernel32.NewProc("GlobalSize")
	copyMemory        = clipboardKernel32.NewProc("RtlMoveMemory")
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

	size, _, _ := globalSize.Call(handle)
	if size < 2 {
		return "", false
	}
	buffer := make([]uint16, int(size/2))
	copyMemory.Call(uintptr(unsafe.Pointer(&buffer[0])), data, uintptr(len(buffer))*2)
	runtime.KeepAlive(buffer)
	for i, value := range buffer {
		if value == 0 {
			buffer = buffer[:i]
			break
		}
	}
	return string(utf16.Decode(buffer)), true
}
