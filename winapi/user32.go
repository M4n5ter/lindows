package winapi

import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/
var (
	user32            = syscall.MustLoadDLL("user32.dll")
	procFindWindow    = user32.MustFindProc("FindWindowW")
	procEnumWindows   = user32.MustFindProc("EnumWindows")
	procPrintWindow   = user32.MustFindProc("PrintWindow")
	procGetClassName  = user32.MustFindProc("GetClassNameW")
	procGetWindowText = user32.MustFindProc("GetWindowTextW")
	procGetDC         = user32.MustFindProc("GetDC")
	procReleaseDC     = user32.MustFindProc("ReleaseDC")
)

// FindWindow 查找窗口
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-findwindoww
//
//	HWND FindWindowW(
//		[in, optional] LPCWSTR lpClassName,
//		[in, optional] LPCWSTR lpWindowName
//	);
func FindWindow(className, windowName string) (hwnd HWND, err error) {
	var lpClassName, lpWindowName *uint16
	if className != "" {
		lpClassName, err = syscall.UTF16PtrFromString(className)
		if err != nil {
			return 0, err
		}
	}
	if windowName != "" {
		lpWindowName, err = syscall.UTF16PtrFromString(windowName)
		if err != nil {
			return 0, err
		}
	}

	r1, _, err := procFindWindow.Call(
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)))
	hwnd = HWND(r1)
	if hwnd == 0 {
		// hwnd == 0 时，表示系统调用失败，以下是为了处理一个特殊情况：
		// 当 procFindWindow.Call 失败，即 hwnd == 0，但是没有返回错误代码（即 err ==0）时，返回 syscall.EINVAL 表示无效参数
		// 之后的代码中类似的处理都是为了处理这种特殊情况
		if err.(syscall.Errno) == 0 {
			err = syscall.EINVAL
		}

		return 0, err
	}
	return hwnd, nil
}

// EnumWindows 枚举所有顶级窗口
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-enumwindows
//
//	BOOL EnumWindows(
//		[in] WNDENUMPROC lpEnumFunc,
//		[in] LPARAM      lParam
//	);
func EnumWindows(enumFunc enumWindowsProc, lParam uintptr) bool {
	// 从enumFunc创建回调函数
	callback := syscall.NewCallback(func(hwnd HWND, lParam uintptr) uintptr {
		// 调用enumFunc，如果返回false，则停止枚举
		if enumFunc(hwnd, lParam) {
			return 1 // 继续枚举
		}
		return 0 // 停止枚举
	})

	r1, _, _ := procEnumWindows.Call(callback, lParam)
	return r1 != 0
}

// enumWindowsProc 枚举窗口回调函数，返回false则停止枚举
type enumWindowsProc func(hwnd HWND, lParam uintptr) bool

// PrintWindow 将窗口内容绘制到指定的设备上下文(DC)
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-printwindow
//
//	BOOL PrintWindow(
//		[in] HWND hwnd,
//		[in] HDC  hdcBlt,
//		[in] UINT nFlags
func PrintWindow(hwnd HWND, hdc HDC, flags uint32) error {
	r1, _, err := procPrintWindow.Call(
		uintptr(hwnd),
		uintptr(hdc),
		uintptr(flags))
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return err
	}
	return nil
}

// GetClassName 获取窗口类名
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-getclassnamew
//
//	int GetClassNameW(
//		[in]  HWND   hWnd,
//		[out] LPWSTR lpClassName,
//		[in]  int    nMaxCount
//	);
func GetClassName(hwnd HWND) (string, error) {
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
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-getwindowtextw
//
//	int GetWindowTextW(
//		[in]  HWND   hWnd,
//		[out] LPWSTR lpString,
//		[in]  int    nMaxCount
//	);
func GetWindowText(hwnd HWND) (string, error) {
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
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-getdc
//
//	HDC GetDC(
//		[in] HWND hWnd
//	);
func GetDC(hwnd HWND) (HDC, error) {
	r1, _, err := procGetDC.Call(uintptr(hwnd))
	hdc := HDC(r1)
	if hdc == 0 {
		if err.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, err
	}
	return hdc, nil
}
