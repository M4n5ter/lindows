package mouse

import (
	"unsafe"

	"github.com/m4n5ter/lindows/winapi"
)

// 设置鼠标（x,y）位置
func SetMousePos(x, y int32) bool {
	return winapi.SetCursorPos(x, y)
}

// 左键
func LeftEvent(dwFlags int) {
	winapi.SendInput(
		1,
		unsafe.Pointer(&winapi.MInput{
			Type: winapi.InputMouse, // INPUT_MOUSE
			MI: winapi.MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(winapi.MInput{})),
	)
}

// 中间按钮
func MiddleEvent(dwFlags int) {
	winapi.SendInput(
		1,
		unsafe.Pointer(&winapi.MInput{
			Type: winapi.InputMouse, // INPUT_MOUSE
			MI: winapi.MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(winapi.MInput{})),
	)
}

// 右键
func RightEvent(dwFlags int) {
	winapi.SendInput(
		1,
		unsafe.Pointer(&winapi.MInput{
			Type: winapi.InputMouse, // INPUT_MOUSE
			MI: winapi.MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(winapi.MInput{})),
	)
}

// 移动
func MoveEvent(x, y int32) {
	winapi.SetCursorPos(x, y)
	winapi.SendInput(
		1,
		unsafe.Pointer(&winapi.MInput{
			Type: winapi.InputMouse, // INPUT_MOUSE
			MI: winapi.MouseInput{
				DwFlags: winapi.MouseEventFMove,
			},
		}),
		uint32(unsafe.Sizeof(winapi.MInput{})),
	)
}

// 滚轮
func WheelEvent(mouseData int) {
	winapi.SendInput(
		1,
		unsafe.Pointer(&winapi.MInput{
			Type: winapi.InputMouse, // INPUT_MOUSE
			MI: winapi.MouseInput{
				MouseData: uint32(mouseData),
				DwFlags:   winapi.MouseEventFWheel,
			},
		}),
		uint32(unsafe.Sizeof(winapi.MInput{})),
	)
}
