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
			return hwnd, err
		}
	}
	if windowName != "" {
		lpWindowName, err = syscall.UTF16PtrFromString(windowName)
		if err != nil {
			return hwnd, err
		}
	}

	r0, _, e1 := procFindWindow.Call(
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)))
	hwnd = syscall.Handle(r0)
	if hwnd == 0 {
		// hwnd == 0 时，表示系统调用失败，以下是为了处理一个特殊情况：
		// 当 procFindWindow.Call 失败，即 hwnd == 0，但是没有返回错误代码（即 e1 ==0）时，返回 syscall.EINVAL 表示无效参数
		// 之后的代码中类似的处理都是为了处理这种特殊情况
		if e1.(syscall.Errno) == 0 {
			err = syscall.EINVAL
		} else {
			err = e1
		}
	}
	return hwnd, err
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

// PrintWindow 将窗口内容绘制到指定的设备上下文(DC)
func PrintWindow(hwnd, hdc syscall.Handle, flags uint32) error {
	r0, _, e1 := procPrintWindow.Call(
		uintptr(hwnd),
		uintptr(hdc),
		uintptr(flags))
	if r0 == 0 {
		if e1.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return e1
	}
	return nil
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
	r0, _, e1 := procGetDC.Call(uintptr(hwnd))
	hdc := syscall.Handle(r0)
	if hdc == 0 {
		if e1.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, e1
	}
	return hdc, nil
}

// CreateCompatibleDC 创建与指定设备兼容的内存设备上下文
//
// 应用程序必须调用 DeleteDC 函数来删除设备上下文。
func CreateCompatibleDC(hdc syscall.Handle) (syscall.Handle, error) {
	r0, _, e1 := procCreateCompatibleDC.Call(uintptr(hdc))
	memDC := syscall.Handle(r0)
	if hdc == 0 {
		if e1.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, e1
	}
	return memDC, nil
}

// CreateCompatibleBitmap 创建与指定设备兼容的位图, width和height为位图的宽和高,以像素为单位
func CreateCompatibleBitmap(hdc syscall.Handle, width, height int32) (syscall.Handle, error) {
	r0, _, e1 := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptr(width), uintptr(height))
	bitmap := syscall.Handle(r0)
	if bitmap == 0 {
		if e1.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, e1
	}
	return bitmap, nil
}

// ReleaseDC 释放设备上下文(HDC)
//
// 应用程序不能使用 ReleaseDC 函数释放通过调用 CreateDC 函数创建的 DC;相反，它必须使用 DeleteDC 函数。 ReleaseDC 必须从调用 GetDC 的同一线程调用。
func ReleaseDC(hwnd, hdc syscall.Handle) error {
	r0, _, e1 := procReleaseDC.Call(uintptr(hwnd), uintptr(hdc))
	if r0 == 0 {
		if e1.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return e1
	}
	return nil
}

// DeleteDC 删除设备上下文(HDC)
//
// 不得删除通过调用 GetDC 函数获取其句柄的 DC。 相反，它必须调用 ReleaseDC 函数来释放 DC。
func DeleteDC(hdc syscall.Handle) error {
	r0, _, e1 := procDeleteDC.Call(uintptr(hdc))
	if r0 == 0 {
		if e1.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return e1
	}
	return nil
}

// BitBlt 复制位图
func BitBlt(hdcDest, nXDest, nYDest, nWidth, nHeight int32, hdcSrc syscall.Handle, nXSrc, nYSrc int32, dwRop uint32) error {
	r0, _, e1 := procBitBlt.Call(uintptr(hdcDest), uintptr(nXDest), uintptr(nYDest), uintptr(nWidth), uintptr(nHeight), uintptr(hdcSrc), uintptr(nXSrc), uintptr(nYSrc), uintptr(dwRop))
	if r0 == 0 {
		if e1.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return e1
	}
	return nil
}
