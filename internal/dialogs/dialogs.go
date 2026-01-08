package dialogs

import (
	"linkrouter/internal/globals"
	"syscall"
	"unsafe"
)

func ShowError(msg string) {
	ShowMessageBox("LinkRouter Error", msg, 0x00000010) // MB_ICONERROR
}

func ShowMessageBox(title, text string, icon uint) int {
	if globals.QuietMode {
		return 0
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	textPtr, _ := syscall.UTF16PtrFromString(text)

	ret, _, _ := msgBox.Call(
		0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(icon|0x00001000), // + MB_TOPMOST
	)
	return int(ret)
}
