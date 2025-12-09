// internal/utils/identity.go
package utils

import (
	"bytes"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modVersion                  = windows.NewLazySystemDLL("version.dll")
	procGetFileVersionInfoSizeW = modVersion.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfoW     = modVersion.NewProc("GetFileVersionInfoW")
	procVerQueryValueW          = modVersion.NewProc("VerQueryValueW")
)

// IsLinkRouter checks if the executable's ProductName is "LinkRouter"
func IsLinkRouter(path string) bool {
	// Get size of version info
	size, _, _ := procGetFileVersionInfoSizeW.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
		0,
	)
	if size == 0 {
		return false
	}

	// Read version info
	buf := make([]byte, size)
	ret, _, _ := procGetFileVersionInfoW.Call(
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(path))),
		0,
		size,
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if ret == 0 {
		return false
	}

	// Query ProductName (en-US, Unicode)
	productName, ok := queryStringValue(buf, `\StringFileInfo\040904B0\ProductName`)
	if !ok {
		return false
	}

	return bytes.Equal(productName, []byte("LinkRouter"))
}

// queryStringValue extracts a UTF-16 null-terminated string from version info
func queryStringValue(data []byte, subBlock string) ([]byte, bool) {
	var p uintptr
	var len uint32

	ret, _, _ := procVerQueryValueW.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(subBlock))),
		uintptr(unsafe.Pointer(&p)),
		uintptr(unsafe.Pointer(&len)),
	)
	if ret == 0 || p == 0 || len == 0 {
		return nil, false
	}

	// Convert UTF-16 to Go string (stop at null)
	utf16 := (*[1 << 20]uint16)(unsafe.Pointer(p))[:len:len]
	for i, u := range utf16 {
		if u == 0 {
			utf16 = utf16[:i]
			break
		}
	}
	return []byte(windows.UTF16ToString(utf16)), true
}
