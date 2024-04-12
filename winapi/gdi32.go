package winapi

import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/
var (
	gdi32                      = syscall.MustLoadDLL("gdi32.dll")
	procCreateCompatibleDC     = gdi32.MustFindProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.MustFindProc("CreateCompatibleBitmap")
	procDeleteDC               = gdi32.MustFindProc("DeleteDC")
	procBitBlt                 = gdi32.MustFindProc("BitBlt")
	procDeleteObject           = gdi32.MustFindProc("DeleteObject")
	procSelectObject           = gdi32.MustFindProc("SelectObject")
	procGetDIBits              = gdi32.MustFindProc("GetDIBits")
)

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD // This is a placeholder, actual color table size varies
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

// CreateCompatibleDC 创建与指定设备兼容的内存设备上下文
//
// 应用程序必须调用 DeleteDC 函数来删除设备上下文。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-createcompatibledc
//
//	HDC CreateCompatibleDC(
//		[in] HDC hdc
//	);
func CreateCompatibleDC(hdc HDC) (HDC, error) {
	r1, _, err := procCreateCompatibleDC.Call(uintptr(hdc))
	memDC := HDC(r1)
	if hdc == 0 {
		if err.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, err
	}
	return memDC, nil
}

// CreateCompatibleBitmap 创建与指定设备兼容的位图, width和height为位图的宽和高,以像素为单位
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-createcompatiblebitmap
//
//	HBITMAP CreateCompatibleBitmap(
//		[in] HDC hdc,
//		[in] int cx,
//		[in] int cy
//	);
func CreateCompatibleBitmap(hdc HDC, cx, cy int32) (HBITMAP, error) {
	r1, _, err := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptr(cx), uintptr(cy))
	hbm := HBITMAP(r1)
	if hbm == 0 {
		if err.(syscall.Errno) == 0 {
			return 0, syscall.EINVAL
		}

		return 0, err
	}
	return hbm, nil
}

// ReleaseDC 释放设备上下文(HDC)
//
// 应用程序不能使用 ReleaseDC 函数释放通过调用 CreateDC 函数创建的 DC;相反，它必须使用 DeleteDC 函数。 ReleaseDC 必须从调用 GetDC 的同一线程调用。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-releasedc
//
//	int ReleaseDC(
//		[in] HWND hWnd,
//		[in] HDC  hDC
//	);
func ReleaseDC(hwnd HWND, hdc HDC) error {
	r1, _, err := procReleaseDC.Call(uintptr(hwnd), uintptr(hdc))
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return err
	}
	return nil
}

// DeleteDC 删除设备上下文(HDC)
//
// 不得删除通过调用 GetDC 函数获取其句柄的 DC。 相反，它必须调用 ReleaseDC 函数来释放 DC。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-deletedc
//
//	BOOL DeleteDC(
//		[in] HDC hdc
//	);
func DeleteDC(hdc HDC) error {
	r1, _, err := procDeleteDC.Call(uintptr(hdc))
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return err
	}
	return nil
}

// BitBlt 复制位图
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-bitblt
//
//	BOOL BitBlt(
//		[in] HDC   hdc,
//		[in] int   x,
//		[in] int   y,
//		[in] int   cx,
//		[in] int   cy,
//		[in] HDC   hdcSrc,
//		[in] int   x1,
//		[in] int   y1,
//		[in] DWORD rop
//	);
func BitBlt(hdc HDC, x, y, cx, cy int32, hdcSrc HDC, x1, y1 int32, rop uint32) error {
	r1, _, err := procBitBlt.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(hdcSrc), uintptr(x1), uintptr(y1), uintptr(rop))
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return err
	}
	return nil
}

// DeleteObject  放与对象关联的所有系统资源
//
// 当绘图对象仍被选入 DC 时，请勿删除 (笔或画笔) 的绘图对象。删除图案画笔时，不会删除与画笔关联的位图。必须单独删除位图。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-deleteobject
//
//	BOOL DeleteObject(
//		[in] HGDIOBJ ho
//	);
func DeleteObject(ho HGDIOBJ) error {
	r1, _, err := procDeleteObject.Call(uintptr(ho))
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINTR
		}

		return err
	}
	return nil
}

// SelectObject 选择一个对象到指定的设备上下文
//
// 如果所选对象不是区域且函数成功，则返回值是所替换对象的句柄。 如果所选对象是区域且函数成功，则返回值是以下值之一。
// NULLREGION：区域为空。
// SIMPLEREGION：区域为矩形。
// COMPLEXREGION：区域为复杂形状。
// 如果发生错误，并且所选对象不是区域，则返回值为 NULL。 否则，它将HGDI_ERROR。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-selectobject
//
//	HGDIOBJ SelectObject(
//		[in] HDC     hdc,
//		[in] HGDIOBJ h
//	);
func SelectObject(hdc HDC, h HGDIOBJ) (HGDIOBJ, error) {
	r1, _, err := procSelectObject.Call(uintptr(hdc), uintptr(h), 0)
	if err.(syscall.Errno) != 0 {
		return HGDIOBJ(r1), err
	}

	return HGDIOBJ(r1), nil
}

// GetDIBits 从指定的设备上下文中检索位图的位
//
// 检索指定兼容位图的位，并使用指定格式将其作为 DIB 复制到缓冲区中。
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/wingdi/nf-wingdi-getdibits
//
//	int GetDIBits(
//		[in]      HDC          hdc,
//		[in]      HBITMAP      hbm,
//		[in]      UINT         start,
//		[in]      UINT         cLines,
//		[out]     LPVOID       lpvBits,
//		[in, out] LPBITMAPINFO lpbmi,
//		[in]      UINT         usage
//	);
func GetDIBits(hdc HDC, hbm HBITMAP, start, cLines uint32, lpvBits unsafe.Pointer, lpbmi *BITMAPINFO, usage uint32) error {
	r1, _, err := procGetDIBits.Call(
		uintptr(hdc),
		uintptr(hbm),
		uintptr(start),
		uintptr(cLines),
		uintptr(lpvBits),
		uintptr(unsafe.Pointer(lpbmi)),
		uintptr(usage),
	)
	if r1 == 0 {
		if err.(syscall.Errno) == 0 {
			return syscall.EINVAL
		}

		return err
	}
	return nil
}
