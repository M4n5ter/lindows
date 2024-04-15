package winapi

import (
	"errors"
	"unsafe"

	"github.com/m4n5ter/lindows/pkg/yalog"
)

var (
	procOpenClipboard    = user32.MustFindProc("OpenClipboard")
	procCloseClipboard   = user32.MustFindProc("CloseClipboard")
	procEmptyClipboard   = user32.MustFindProc("EmptyClipboard")
	procSetClipboardData = user32.MustFindProc("SetClipboardData")
	procGetClipboardData = user32.MustFindProc("GetClipboardData")
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

// WriteTextToClipboard 把文本写入到剪切板
func WriteTextToClipboard(content string) {
	var hMemory HGLOBAL
	var lpMemory LPVOID // 或者 *uint32，根据系统位数选择
	var err error
	contentSize := int32(len(content)) + 1
	// 打开剪切板
	if !OpenClipboard(HWND(uintptr(0))) {
		yalog.Info("OpenClipboard", "剪切板打开失败")
		return
	}
	// 清空剪切板
	if !EmptyClipboard() {
		yalog.Info("EmptyClipboard", "剪切板清空失败")
		CloseClipboard()
		return
	}

	// 申请内存
	hMemory, err = GlobalAlloc(GmemMoveable, contentSize)
	if err != nil {
		yalog.Info("GlobalAlloc", "获取内存失败")
		CloseClipboard()
		return
	}
	defer GlobalFree(hMemory)
	// 锁内存
	lpMemory, err = GlobalLock(hMemory)
	if err != nil {
		yalog.Info("GlobalLock", "内存锁定失败")
		CloseClipboard()
		return
	}
	defer GlobalUnlock(hMemory)
	// 这里获取的内存地址
	cMemoryAddr := uintptr(lpMemory)

	// 这里是要传输的数据
	data := []byte(content)

	// 将数据复制到 C 语言申请的内存中
	// copy((*[n]byte)(unsafe.Pointer(cMemoryAddr))[:], data)
	// 遍历一个地址一个值
	for i := 0; i < len(data); i++ {
		b := data[i]
		*(*byte)(unsafe.Pointer(cMemoryAddr + uintptr(i))) = b
	}

	_, err = SetClipboardData(CFText, HANDLE(lpMemory))
	if err != nil {
		yalog.Info("SetClipboardData", "", "剪切板设置失败")
		return
	}
}

// 从剪切板读取文本数据
func ReaderTextToClipboard() (string, error) {
	var lpMemory LPVOID // 或者 *uint32，根据系统位数选择

	var err error
	// 打开剪切板
	if !OpenClipboard(HWND(uintptr(0))) {
		//	yalog.Info("OpenClipboard", "剪切板打开失败")
		return "", errors.New("OpenClipboard:剪切板打开失败")
	}
	// 获取剪切板的数据
	r1, err := GetClipboardData(CFText)
	if err != nil {
		// yalog.Info("GetClipboardData", "", "剪切板数据获取失败")
		return "", errors.New("GetClipboardData:剪切板数据获取失败")
	}
	// 锁住 剪切板的内存
	lpMemory, err = GlobalLock(HGLOBAL(r1))
	if err != nil {
		//	yalog.Info("GlobalLock", "内存锁定失败")
		CloseClipboard()
		return "", errors.New("GlobalLock:内存锁定失败")
	}
	defer GlobalUnlock(HGLOBAL(r1))

	// 将内存块的内容复制到字节切片中
	data := make([]byte, 0)
	for {
		b := *(*byte)(unsafe.Pointer(uintptr(lpMemory) + uintptr(len(data))))
		if b == 0 {
			break
		}
		data = append(data, b)
	}

	// 根据数据的编码解析为字符串
	str := string(data)
	//	yalog.Info("剪切板的内容", "value=", str)
	return str, nil
}
