package winapi

import (
	"syscall"
	"unsafe"
)

type (
	HWND    syscall.Handle
	HDC     syscall.Handle
	HGDIOBJ syscall.Handle
	HGLOBAL syscall.Handle
	LPVOID  unsafe.Pointer
	HANDLE  syscall.Handle
)

/*const (
GHND           = 0x0042
SRCCOPY        = 0x00CC0020
DIB_RGB_COLORS = 0
BI_RGB         = 0
)
*/

type (
	HBITMAP = HGDIOBJ
	HBRUSH  = HGDIOBJ
	HPEN    = HGDIOBJ
	HFONT   = HGDIOBJ
	HRGN    = HGDIOBJ
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type (
	LPINPUT = uintptr
)

// 事件
const (
	InputMouse uint32 = 0 // 事件是鼠标事件。 使用联合的 mi 结构。
)

type MInput struct {
	Type uint32
	MI   MouseInput
}

type MouseEvent = uint32

// 鼠标
const (
	MouseEventFMove          MouseEvent = 0x0001 // 发生了移动。
	MouseEventFLeftDown      MouseEvent = 0x0002 // 按下了左侧按钮。
	MouseEventFLeftUp        MouseEvent = 0x0004 // 左按钮已释放。
	MouseEventFRightDown     MouseEvent = 0x0008 // 按下了向右按钮。
	MouseEventFRightUp       MouseEvent = 0x0010 // 右侧按钮已松开。
	MouseEventFMiddleDown    MouseEvent = 0x0020 // 按下中间按钮。
	MouseEventFMiddleUp      MouseEvent = 0x0040 // 中间按钮已释放。
	MouseEventFXDown         MouseEvent = 0x0080 // 按下了 X 按钮。
	MouseEventFXUp           MouseEvent = 0x0100 // 已释放 X 按钮。
	MouseEventFWheel         MouseEvent = 0x0800 // 如果鼠标有滚轮，则滚轮已移动。移动量在 mouseData 中指定。
	MouseEventFHWheel        MouseEvent = 0x1000 // 如果鼠标有滚轮，则方向盘是水平移动的。移动量在 mouseData 中指定。
	MouseEventFMoveNcoALESce MouseEvent = 0x2000 // 不会合并WM_MOUSEMOVE消息。默认行为是合并 WM_MOUSEMOVE 消息。
	MouseEventFVirtualDesk   MouseEvent = 0x4000 // 将坐标映射到整个桌面。必须与 MouseEventF_ABSOLUTE 一起使用。
	MouseEventFABSolute      MouseEvent = 0x8000 // dx 和 dy 成员包含规范化的绝对坐标。如果未设置标志，dx和dy包含相对数据(自上次报告的位置)更改。无论哪种类型的鼠标或其他指针设备(如果有)连接到系统，都可以设置或不设置此标志。

)

type MouseInput struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// 内存分配属性

const (
	Ghnd         = 0x0042
	GmemMoveable = 0x0002
	GmemZeroInit = 0x0040
	GmemFixed    = 0x0000
	Gptr         = GmemFixed | GmemZeroInit
)

// 截切版格式
// 标准格式 :https://learn.microsoft.com/zh-cn/windows/desktop/dataxchg/standard-clipboard-formats
const (
	CFBitmap          = 2
	CFDIB             = 8
	CFDIBV5           = 17
	CFDIF             = 5
	CFDSPBitmap       = 0x0082
	CFDSPEnhMetaFile  = 0x008E
	CFDSPMetaFilePict = 0x0083
	CFDSPText         = 0x0081
	CFEnhMetaFile     = 14
	CFGdiObjFirst     = 0x0300
	CFGdiObjLast      = 0x03FF
	CFHDrop           = 15
	CFLocale          = 16
	CFMetaFilePict    = 3
	CFOEMText         = 7
	CFOwnerDisplay    = 0x0080
	CFPalette         = 9
	CFPenData         = 10
	CFPrivateFirst    = 0x0200
	CFPrivateLast     = 0x02FF
	CFRIFF            = 11
	CFSYLK            = 4
	CFText            = 1
	CFTIFF            = 6
	CFUnicodeText     = 13
	CFWave            = 12
)
