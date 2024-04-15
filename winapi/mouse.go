package winapi

import "unsafe"

// 设置鼠标（x,y）位置
func SetMousePos(x, y int32) bool {
	return SetCursorPos(x, y)
}

// 左键
func MouseLeftEvent(dwFlags int) {
	SendInput(
		1,
		unsafe.Pointer(&MInput{
			Type: InputMouse, // INPUT_MOUSE
			MI: MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(MInput{})),
	)
}

// 中间按钮
func MouseMiddleEvent(dwFlags int) {
	SendInput(
		1,
		unsafe.Pointer(&MInput{
			Type: InputMouse, // INPUT_MOUSE
			MI: MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(MInput{})),
	)
}

// 右键
func MouseRightEvent(dwFlags int) {
	SendInput(
		1,
		unsafe.Pointer(&MInput{
			Type: InputMouse, // INPUT_MOUSE
			MI: MouseInput{
				DwFlags: uint32(dwFlags),
			},
		}),
		uint32(unsafe.Sizeof(MInput{})),
	)
}

// 移动
func MouseMoveEvent(x, y int32) {
	SetCursorPos(x, y)
	SendInput(
		1,
		unsafe.Pointer(&MInput{
			Type: InputMouse, // INPUT_MOUSE
			MI: MouseInput{
				DwFlags: MouseEventFMove,
			},
		}),
		uint32(unsafe.Sizeof(MInput{})),
	)
}

// 滚轮
func MouseWheelEvent(mouseData int) {
	SendInput(
		1,
		unsafe.Pointer(&MInput{
			Type: InputMouse, // INPUT_MOUSE
			MI: MouseInput{
				MouseData: uint32(mouseData),
				DwFlags:   MouseEventFWheel,
			},
		}),
		uint32(unsafe.Sizeof(MInput{})),
	)
}
