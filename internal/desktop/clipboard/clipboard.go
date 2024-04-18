package clipboard

import (
	"errors"
	"runtime"
	"time"
	"unsafe"

	"github.com/m4n5ter/lindows/winapi"
)

// WriteTextToClipboard 把文本写入到剪切板
func WriteTextToClipboard(content string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var hMemory winapi.HGLOBAL
	var lpMemory winapi.LPVOID // 或者 *uint32，根据系统位数选择

	var err error
	contentSize := int32(len(content)) + 1
	err = waitOpenClipboard()
	if err != nil {
		return errors.New("OpenClipboard :剪切板打开失败")
	}
	defer winapi.CloseClipboard()

	// 清空剪切板
	if !winapi.EmptyClipboard() {
		return errors.New("EmptyClipboard :剪切板清空失败")
	}

	// 申请内存
	hMemory, err = winapi.GlobalAlloc(winapi.GmemMoveable, contentSize)
	if err != nil {
		return errors.New("GlobalAlloc :获取内存失败")
	}
	defer winapi.GlobalFree(hMemory)
	// 锁内存
	lpMemory, err = winapi.GlobalLock(hMemory)
	if err != nil {
		return errors.New("GlobalLock :内存锁定失败")
	}
	defer winapi.GlobalUnlock(hMemory)
	// 这里获取的内存地址
	cMemoryAddr := uintptr(lpMemory)

	// 这里是要传输的数据
	data := []byte(content)

	// 将数据复制到 C 语言申请的内存中
	// 遍历一个地址一个值
	for i := 0; i < len(data); i++ {
		b := data[i]
		*(*byte)(unsafe.Pointer(cMemoryAddr + uintptr(i))) = b
	}

	_, err = winapi.SetClipboardData(winapi.CFText, winapi.HANDLE(hMemory))
	if err != nil {
		return errors.New("SetClipboardData :剪切板设置失败")
	}
	return nil
}

// ReaderTextToClipboard 从剪切板读取文本数据
func ReaderTextToClipboard() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var lpMemory winapi.LPVOID // 或者 *uint32，根据系统位数选择
	var err error
	// 打开剪切板
	err = waitOpenClipboard()
	if err != nil {
		return "", errors.New("OpenClipboard :剪切板打开失败")
	}
	defer winapi.CloseClipboard()
	// 获取剪切板的数据
	r1, err := winapi.GetClipboardData(winapi.CFText)
	if err != nil {
		return "", errors.New("GetClipboardData:剪切板数据获取失败")
	}
	// 锁住 剪切板的内存
	lpMemory, err = winapi.GlobalLock(winapi.HGLOBAL(r1))
	if err != nil {
		return "", errors.New("GlobalLock:内存锁定失败")
	}
	defer winapi.GlobalUnlock(winapi.HGLOBAL(r1))

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
	return str, nil
}

// waitOpenClipboard 等待打开剪切板
func waitOpenClipboard() error {
	started := time.Now()
	limit := started.Add(time.Second)
	var r bool
	var err error
	for time.Now().Before(limit) {
		r = winapi.OpenClipboard(winapi.HWND(uintptr(0)))
		if r {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
	return err
}
