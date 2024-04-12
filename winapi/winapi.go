package winapi

import "syscall"

type (
	HWND    syscall.Handle
	HDC     syscall.Handle
	HGDIOBJ syscall.Handle
)

type (
	HBITMAP = HGDIOBJ
	HBRUSH  = HGDIOBJ
	HPEN    = HGDIOBJ
	HFONT   = HGDIOBJ
	HRGN    = HGDIOBJ
)
