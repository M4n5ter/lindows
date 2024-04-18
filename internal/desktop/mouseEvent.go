package desktop

import (
	"github.com/m4n5ter/lindows/internal/desktop/mouse"
)

func (manager *Manager) MouseMoveEvent(x, y int32) {
	mouse.MoveEvent(x, y)
}

func (manager *Manager) SetMousePos(x, y int32) bool {
	return mouse.SetMousePos(x, y)
}

func (manager *Manager) MouseLeftEvent(me int) bool {
	mouse.LeftEvent(me)
	return true
}

func (manager *Manager) MouseRightEvent(me int) bool {
	mouse.RightEvent(me)
	return true
}

func (manager *Manager) MouseMiddleEvent(me int) bool {
	mouse.MiddleEvent(me)
	return true
}
