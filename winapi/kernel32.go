package winapi

/*
import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/en-us/windows/win32/api/libloaderapi/nf-libloaderapi-getmodulehandlew
var (
	kernel32        = syscall.MustLoadDLL("Kernel32.dll")
	procGlobalAlloc = kernel32.MustFindProc("GlobalAlloc")
	procGGlobalLock = kernel32.MustFindProc("GlobalLock")
)

// GlobalAlloc 从堆中分配指定的字节数
func GlobalAlloc(uFlags, dwBytes int32) (HWND, error) {
	r1, _, err := procGlobalAlloc.Call(uintptr(uFlags), uintptr(dwBytes))

	if r1 == 0 {
		return HWND(r1), err
	}
	return HWND(r1), nil
}

// GlobalLock
func GlobalLock(hwnd HWND) uintptr {
	r1, _, err := procGGlobalLock.Call(uintptr(hwnd))

	if r1 == 0 {
		var pt *int = nil
		return uintptr(unsafe.Pointer(pt))
	}
	return r1
}
*/
