package winapi

import "syscall"

// https://learn.microsoft.com/ja-JP/windows/win32/api/winbase
var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procGlobalAlloc  = kernel32.MustFindProc("GlobalAlloc")
	procGlobalLock   = kernel32.MustFindProc("GlobalLock")
	procGlobalUnlock = kernel32.MustFindProc("GlobalUnlock")
)

// GlobalAlloc 获取内存
//
// https://learn.microsoft.com/ja-jp/windows/win32/api/winbase/nf-winbase-globalalloc
// DECLSPEC_ALLOCATOR HGLOBAL GlobalAlloc(
//
//	[in] UINT   uFlags,
//	[in] SIZE_T dwBytes
//
// );
func GlobalAlloc(uFlag uint32, dwBytes int32) (HGLOBAL, error) {
	r1, _, err := procGlobalAlloc.Call(uintptr(uFlag), uintptr(dwBytes))
	if r1 == 0 {
		return HGLOBAL(r1), err
	}
	if isErr(err) {
		return HGLOBAL(r1), err
	}
	return HGLOBAL(r1), nil
}

// GlobalLock 锁定全局内存对象，并返回指向该对象内存块的第一个字节的指针。
//
// https://learn.microsoft.com/ja-JP/windows/win32/api/winbase/nf-winbase-globallock
//
// LPVOID GlobalLock(
//
//	[in] HGLOBAL hMem
//
// );
// 如果函数成功，则返回值是指向内存块的第一个字节的指针。 如果函数失败，则返回的值为 NULL。
func GlobalLock(hMem HGLOBAL) (LPVOID, error) {
	r1, _, err := procGlobalLock.Call(uintptr(hMem))
	if r1 == 0 {
		return nil, err
	}
	if isErr(err) {
		return nil, err
	}
	return LPVOID(r1), nil
}

// GlobalUnlock 解锁全局内存对象
//
// https://learn.microsoft.com/ja-jp/windows/win32/api/winbase/nf-winbase-globalunlock
//
// BOOL GlobalUnlock(
//
//	[in] HGLOBAL hMem
//
// );
// 如果在减少锁数后解锁内存对象，则此函数返回 0，GetLastError 返回 NO_ERROR。
// 如果函数失败，则返回值为零，并且 GetLastError 返回 NO_ERROR 以外的值。
func GlobalUnlock(hMem HGLOBAL) bool {
	r1, _, _ := procGlobalUnlock.Call(uintptr(hMem))
	return r1 != 0
}
