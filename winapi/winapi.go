package winapi

import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/
var (
	user32            = syscall.MustLoadDLL("user32.dll")
	procFindWindow    = user32.MustFindProc("FindWindowW")
	procEnumWindows   = user32.MustFindProc("EnumWindows")
	procPrintWindow   = user32.MustFindProc("PrintWindow")
	procGetClassName  = user32.MustFindProc("GetClassNameW")
	procGetWindowText = user32.MustFindProc("GetWindowTextW")
	procGetDC         = user32.MustFindProc("GetDC")
	procReleaseDC     = user32.MustFindProc("ReleaseDC")

	gdi32                      = syscall.MustLoadDLL("gdi32.dll")
	procCreateCompatibleDC     = gdi32.MustFindProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.MustFindProc("CreateCompatibleBitmap")
	procDeleteDC               = gdi32.MustFindProc("DeleteDC")
	procBitBlt                 = gdi32.MustFindProc("BitBlt")
)

// FindWindow 查找窗口,
func FindWindow(className, windowName string) (hwnd syscall.Handle, err error) {
	var lpClassName, lpWindowName *uint16
	if className != "" {
		lpClassName, err = syscall.UTF16PtrFromString(className)
		if err != nil {
			return
		}
	}
	if windowName != "" {
		lpWindowName, err = syscall.UTF16PtrFromString(windowName)
		if err != nil {
			return
		}
	}

	r0, _, e1 := syscall.SyscallN(procFindWindow.Addr(),
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)),
		0)
	hwnd = syscall.Handle(r0)
	if hwnd == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// EnumWindows 枚举所有顶级窗口
func EnumWindows(enumFunc enumWindowsProc, lParam uintptr) bool {
	// 从enumFunc创建回调函数
	callback := syscall.NewCallback(func(hwnd syscall.Handle, lParam uintptr) uintptr {
		// 调用enumFunc，如果返回false，则停止枚举
		if enumFunc(hwnd, lParam) {
			return 1 // 继续枚举
		}
		return 0 // 停止枚举
	})

	ret, _, _ := procEnumWindows.Call(callback, lParam)
	return ret != 0
}

// enumWindowsProc 枚举窗口回调函数，返回false则停止枚举
type enumWindowsProc func(hwnd syscall.Handle, lParam uintptr) bool

// PrintWindow 打印窗口
func PrintWindow(hwnd syscall.Handle, hdc syscall.Handle, nFlags uint32) bool {
	r1, _, err := syscall.SyscallN(procPrintWindow.Addr(), uintptr(hwnd), uintptr(hdc), uintptr(nFlags))
	if error(err) == nil {
		return false
	}
	if uint32(r1) != 0 {
		return true
	}
	return false
}

// GetClassName 获取窗口类名
func GetClassName(hwnd syscall.Handle) (string, error) {
	var className [256]uint16
	_, _, err := procGetClassName.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&className[0])),
		uintptr(len(className)),
	)
	if err != syscall.Errno(0) {
		return "", err
	}
	return syscall.UTF16ToString(className[:]), nil
}

// GetWindowText 获取窗口标题
func GetWindowText(hwnd syscall.Handle) (string, error) {
	var text [256]uint16
	_, _, err := procGetWindowText.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&text[0])),
		uintptr(len(text)),
	)
	if err != syscall.Errno(0) {
		return "", err
	}
	return syscall.UTF16ToString(text[:]), nil
}

// GetDC 获取设备上下文(HDC)
//
// 应用程序必须调用 ReleaseDC 函数来释放设备上下文。
func GetDC(hwnd syscall.Handle) (syscall.Handle, error) {
	r0, _, e1 := syscall.SyscallN(procGetDC.Addr(), uintptr(hwnd), 0, 0)
	hdc := syscall.Handle(r0)
	if hdc == 0 {
		if e1 != 0 {
			return 0, error(e1)
		} else {
			return 0, syscall.EINVAL
		}
	}
	return hdc, nil
}

// CreateCompatibleDC 创建与指定设备兼容的内存设备上下文
//
// 应用程序必须调用 DeleteDC 函数来删除设备上下文。
func CreateCompatibleDC(hdc syscall.Handle) (syscall.Handle, error) {
	r0, _, e1 := syscall.SyscallN(procCreateCompatibleDC.Addr(), uintptr(hdc), 0, 0)
	memDC := syscall.Handle(r0)
	if hdc == 0 {
		if e1 != 0 {
			return 0, error(e1)
		} else {
			return 0, syscall.EINVAL
		}
	}
	return memDC, nil
}

// CreateCompatibleBitmap 创建与指定设备兼容的位图, width和height为位图的宽和高,以像素为单位
func CreateCompatibleBitmap(hdc syscall.Handle, width, height int32) (syscall.Handle, error) {
	r0, _, e1 := syscall.SyscallN(procCreateCompatibleBitmap.Addr(), uintptr(hdc), uintptr(width), uintptr(height))
	bitmap := syscall.Handle(r0)
	if bitmap == 0 {
		if e1 != 0 {
			return 0, error(e1)
		} else {
			return 0, syscall.EINVAL
		}
	}
	return bitmap, nil
}

// ReleaseDC 释放设备上下文(HDC)
//
// 应用程序不能使用 ReleaseDC 函数释放通过调用 CreateDC 函数创建的 DC;相反，它必须使用 DeleteDC 函数。 ReleaseDC 必须从调用 GetDC 的同一线程调用。
func ReleaseDC(hwnd, hdc syscall.Handle) error {
	r0, _, e1 := syscall.SyscallN(procReleaseDC.Addr(), uintptr(hwnd), uintptr(hdc), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		} else {
			return syscall.EINVAL
		}
	}
	return nil
}

// DeleteDC 删除设备上下文(HDC)
//
// 不得删除通过调用 GetDC 函数获取其句柄的 DC。 相反，它必须调用 ReleaseDC 函数来释放 DC。
func DeleteDC(hdc syscall.Handle) error {
	r0, _, e1 := syscall.SyscallN(procDeleteDC.Addr(), uintptr(hdc), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		} else {
			return syscall.EINVAL
		}
	}
	return nil
}

// BitBlt 复制位图
func BitBlt(hdcDest, nXDest, nYDest, nWidth, nHeight int32, hdcSrc syscall.Handle, nXSrc, nYSrc int32, dwRop uint32) error {
	r0, _, e1 := syscall.SyscallN(procBitBlt.Addr(), uintptr(hdcDest), uintptr(nXDest), uintptr(nYDest), uintptr(nWidth), uintptr(nHeight), uintptr(hdcSrc), uintptr(nXSrc), uintptr(nYSrc), uintptr(dwRop))
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		} else {
			return syscall.EINVAL
		}
	}
	return nil
}
