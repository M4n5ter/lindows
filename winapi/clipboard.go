package winapi

var (
	procOpenClipboard          = user32.MustFindProc("OpenClipboard")
	procCloseClipboard         = user32.MustFindProc("CloseClipboard")
	procEmptyClipboard         = user32.MustFindProc("EmptyClipboard")
	procSetClipboardData       = user32.MustFindProc("SetClipboardData")
	procGetClipboardData       = user32.MustFindProc("GetClipboardData")
	procGetOpenClipboardWindow = user32.MustFindProc("GetOpenClipboardWindow")
)

// OpenClipboard 打开剪切板
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-openclipboard
//
// BOOL OpenClipboard(
//
//	 [in, optional] HWND hWndNewOwner
//		);
func OpenClipboard(hWndNewOwner HWND) bool {
	r1, _, err := procOpenClipboard.Call(uintptr(hWndNewOwner))
	if r1 == 0 {
		return false
	}
	if isErr(err) {
		return false
	}
	return true
}

// CloseClipboard 关闭剪切板
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-closeclipboard
// BOOL CloseClipboard();
func CloseClipboard() bool {
	r1, _, err := procCloseClipboard.Call()
	if r1 == 0 {
		return false
	}
	if isErr(err) {
		return false
	}
	return true
}

// EmptyClipboard 清空剪切板
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-emptyclipboard
// BOOL EmptyClipboard();
func EmptyClipboard() bool {
	r1, _, err := procEmptyClipboard.Call()
	if r1 == 0 {
		return false
	}
	if isErr(err) {
		return false
	}
	return true
}

// SetClipboardData 设置剪切板数据
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-setclipboarddata
//
// HANDLE SetClipboardData(
//
//	[in]           UINT   uFormat,
//	[in, optional] HANDLE hMem
//
// );
// uFormat 剪贴板格式。
//
//		标准剪贴板格式 :https://learn.microsoft.com/zh-cn/windows/desktop/dataxchg/standard-clipboard-formats
//	 已注册的剪贴板格式 :https://learn.microsoft.com/zh-cn/windows/desktop/dataxchg/clipboard-formats
func SetClipboardData(uFormat uint32, hMem HANDLE) (HANDLE, error) {
	r1, _, err := procSetClipboardData.Call(uintptr(uFormat), uintptr(hMem))
	if r1 == 0 {
		return HANDLE(r1), err
	}
	if isErr(err) {
		return HANDLE(r1), err
	}
	return HANDLE(r1), nil
}

// GetClipboardData 获取剪切板数据
//
// https://learn.microsoft.com/zh-cn/windows/win32/api/winuser/nf-winuser-getclipboarddata
// HANDLE GetClipboardData(
//
//	[in] UINT uFormat
//
// );
func GetClipboardData(uFormat uint32) (HANDLE, error) {
	r1, _, err := procGetClipboardData.Call(uintptr(uFormat))
	if r1 == 0 {
		return HANDLE(r1), err
	}
	if isErr(err) {
		return HANDLE(r1), err
	}
	return HANDLE(r1), nil
}

// GetOpenClipboardWindow
func GetOpenClipboardWindow() HWND {
	r1, _, _ := procGetOpenClipboardWindow.Call()

	return HWND(r1)
}

// GetClipboardOwner
func GetClipboardOwner() HWND {
	r1, _, _ := procGetClipboardData.Call()

	return HWND(r1)
}
