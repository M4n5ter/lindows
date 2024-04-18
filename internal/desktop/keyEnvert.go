package desktop

import "github.com/m4n5ter/lindows/winapi"

// import (
// 	"github.com/m4n5ter/lindows/internal/desktop/mouse"
// )

// func (manager *Manager) MouseMoveEvent(x, y int32) {
// 	mouse.MoveEvent(x, y)
// }

// func (manager *Manager) SetMousePos(x, y int32) bool {
// 	return mouse.SetMousePos(x, y)
// }

// func (manager *Manager) MouseLeftEvent(me int) bool {
// 	mouse.LeftEvent(me)
// 	return true
// }

// func (manager *Manager) MouseRightEvent(me int) bool {
// 	mouse.RightEvent(me)
// 	return true
// }

// func (manager *Manager) MouseMiddleEvent(me int) bool {
// 	mouse.MiddleEvent(me)
// 	return true
// }

/*
	_VK_SHIFT           = 0x10 + 0xFFF
	_VK_CTRL            = 0x11 + 0xFFF
	_VK_ALT             = 0x12 + 0xFFF
	_VK_LSHIFT          = 0xA0 + 0xFFF
	_VK_RSHIFT          = 0xA1 + 0xFFF
	_VK_LCONTROL        = 0xA2 + 0xFFF
	_VK_RCONTROL        = 0xA3 + 0xFFF
	_VK_LWIN            = 0x5B + 0xFFF
	_VK_RWIN            = 0x5C + 0xFFF
	_KEYEVENTF_KEYUP    = 0x0002
	_KEYEVENTF_SCANCODE = 0x0008

	VK_SP1  = 41
	VK_SP2  = 12
	VK_SP3  = 13
	VK_SP4  = 26
	VK_SP5  = 27
	VK_SP6  = 39
	VK_SP7  = 40
	VK_SP8  = 43
	VK_SP9  = 51
	VK_SP10 = 52
	VK_SP11 = 53
	VK_SP12 = 86

	VK_ESC = 1
	VK_1   = 2
	VK_2   = 3
	VK_3   = 4
	VK_4   = 5
	VK_5   = 6
	VK_6   = 7
	VK_7   = 8
	VK_8   = 9
	VK_9   = 10
	VK_0   = 11
	VK_Q   = 16
	VK_W   = 17
	VK_E   = 18
	VK_R   = 19
	VK_T   = 20
	VK_Y   = 21
	VK_U   = 22
	VK_I   = 23
	VK_O   = 24
	VK_P   = 25
	VK_A   = 30
	VK_S   = 31
	VK_D   = 32
	VK_F   = 33
	VK_G   = 34
	VK_H   = 35
	VK_J   = 36
	VK_K   = 37
	VK_L   = 38
	VK_Z   = 44
	VK_X   = 45
	VK_C   = 46
	VK_V   = 47
	VK_B   = 48
	VK_N   = 49
	VK_M   = 50
	VK_F1  = 59
	VK_F2  = 60
	VK_F3  = 61
	VK_F4  = 62
	VK_F5  = 63
	VK_F6  = 64
	VK_F7  = 65
	VK_F8  = 66
	VK_F9  = 67
	VK_F10 = 68
	VK_F11 = 87
	VK_F12 = 88

	VK_F13 = 0x7C + 0xFFF
	VK_F14 = 0x7D + 0xFFF
	VK_F15 = 0x7E + 0xFFF
	VK_F16 = 0x7F + 0xFFF
	VK_F17 = 0x80 + 0xFFF
	VK_F18 = 0x81 + 0xFFF
	VK_F19 = 0x82 + 0xFFF
	VK_F20 = 0x83 + 0xFFF
	VK_F21 = 0x84 + 0xFFF
	VK_F22 = 0x85 + 0xFFF
	VK_F23 = 0x86 + 0xFFF
	VK_F24 = 0x87 + 0xFFF

	VK_NUMLOCK    = 69
	VK_SCROLLLOCK = 70
	VK_RESERVED   = 0
	VK_MINUS      = 12
	VK_EQUAL      = 13
	VK_BACKSPACE  = 14
	VK_TAB        = 15
	VK_LEFTBRACE  = 26
	VK_RIGHTBRACE = 27
	VK_ENTER      = 28
	VK_SEMICOLON  = 39
	VK_APOSTROPHE = 40
	VK_GRAVE      = 41
	VK_BACKSLASH  = 43
	VK_COMMA      = 51
	VK_DOT        = 52
	VK_SLASH      = 53
	VK_KPASTERISK = 55
	VK_SPACE      = 57
	VK_CAPSLOCK   = 58

	VK_KP0     = 82
	VK_KP1     = 79
	VK_KP2     = 80
	VK_KP3     = 81
	VK_KP4     = 75
	VK_KP5     = 76
	VK_KP6     = 77
	VK_KP7     = 71
	VK_KP8     = 72
	VK_KP9     = 73
	VK_KPMINUS = 74
	VK_KPPLUS  = 78
	VK_KPDOT   = 83

	VK_LBUTTON    = 0x01 + 0xFFF
	VK_RBUTTON    = 0x02 + 0xFFF
	VK_CANCEL     = 0x03 + 0xFFF
	VK_MBUTTON    = 0x04 + 0xFFF
	VK_XBUTTON1   = 0x05 + 0xFFF
	VK_XBUTTON2   = 0x06 + 0xFFF
	VK_BACK       = 0x08 + 0xFFF
	VK_CLEAR      = 0x0C + 0xFFF
	VK_PAUSE      = 0x13 + 0xFFF
	VK_CAPITAL    = 0x14 + 0xFFF
	VK_KANA       = 0x15 + 0xFFF
	VK_HANGUEL    = 0x15 + 0xFFF
	VK_HANGUL     = 0x15 + 0xFFF
	VK_JUNJA      = 0x17 + 0xFFF
	VK_FINAL      = 0x18 + 0xFFF
	VK_HANJA      = 0x19 + 0xFFF
	VK_KANJI      = 0x19 + 0xFFF
	VK_CONVERT    = 0x1C + 0xFFF
	VK_NONCONVERT = 0x1D + 0xFFF
	VK_ACCEPT     = 0x1E + 0xFFF
	VK_MODECHANGE = 0x1F + 0xFFF
	VK_PAGEUP     = 0x21 + 0xFFF
	VK_PAGEDOWN   = 0x22 + 0xFFF
	VK_END        = 0x23 + 0xFFF
	VK_HOME       = 0x24 + 0xFFF
	VK_LEFT       = 0x25 + 0xFFF
	VK_UP         = 0x26 + 0xFFF
	VK_RIGHT      = 0x27 + 0xFFF
	VK_DOWN       = 0x28 + 0xFFF
	VK_SELECT     = 0x29 + 0xFFF
	VK_PRINT      = 0x2A + 0xFFF
	VK_EXECUTE    = 0x2B + 0xFFF
	VK_SNAPSHOT   = 0x2C + 0xFFF
	VK_INSERT     = 0x2D + 0xFFF
	VK_DELETE     = 0x2E + 0xFFF
	VK_HELP       = 0x2F + 0xFFF

	VK_SCROLL              = 0x91 + 0xFFF
	VK_LMENU               = 0xA4 + 0xFFF
	VK_RMENU               = 0xA5 + 0xFFF
	VK_BROWSER_BACK        = 0xA6 + 0xFFF
	VK_BROWSER_FORWARD     = 0xA7 + 0xFFF
	VK_BROWSER_REFRESH     = 0xA8 + 0xFFF
	VK_BROWSER_STOP        = 0xA9 + 0xFFF
	VK_BROWSER_SEARCH      = 0xAA + 0xFFF
	VK_BROWSER_FAVORITES   = 0xAB + 0xFFF
	VK_BROWSER_HOME        = 0xAC + 0xFFF
	VK_VOLUME_MUTE         = 0xAD + 0xFFF
	VK_VOLUME_DOWN         = 0xAE + 0xFFF
	VK_VOLUME_UP           = 0xAF + 0xFFF
	VK_MEDIA_NEXT_TRACK    = 0xB0 + 0xFFF
	VK_MEDIA_PREV_TRACK    = 0xB1 + 0xFFF
	VK_MEDIA_STOP          = 0xB2 + 0xFFF
	VK_MEDIA_PLAY_PAUSE    = 0xB3 + 0xFFF
	VK_LAUNCH_MAIL         = 0xB4 + 0xFFF
	VK_LAUNCH_MEDIA_SELECT = 0xB5 + 0xFFF
	VK_LAUNCH_APP1         = 0xB6 + 0xFFF
	VK_LAUNCH_APP2         = 0xB7 + 0xFFF
	VK_OEM_1               = 0xBA + 0xFFF
	VK_OEM_PLUS            = 0xBB + 0xFFF
	VK_OEM_COMMA           = 0xBC + 0xFFF
	VK_OEM_MINUS           = 0xBD + 0xFFF
	VK_OEM_PERIOD          = 0xBE + 0xFFF
	VK_OEM_2               = 0xBF + 0xFFF
	VK_OEM_3               = 0xC0 + 0xFFF
	VK_OEM_4               = 0xDB + 0xFFF
	VK_OEM_5               = 0xDC + 0xFFF
	VK_OEM_6               = 0xDD + 0xFFF
	VK_OEM_7               = 0xDE + 0xFFF
	VK_OEM_8               = 0xDF + 0xFFF
	VK_OEM_102             = 0xE2 + 0xFFF
	VK_PROCESSKEY          = 0xE5 + 0xFFF
	VK_PACKET              = 0xE7 + 0xFFF
	VK_ATTN                = 0xF6 + 0xFFF
	VK_CRSEL               = 0xF7 + 0xFFF
	VK_EXSEL               = 0xF8 + 0xFFF
	VK_EREOF               = 0xF9 + 0xFFF
	VK_PLAY                = 0xFA + 0xFFF
	VK_ZOOM                = 0xFB + 0xFFF
	VK_NONAME              = 0xFC + 0xFFF
	VK_PA1                 = 0xFD + 0xFFF
	VK_OEM_CLEAR           = 0xFE + 0xFFF
*/

func correspond(k keys) int {
	// 将key转换为对应的键值
	switch k {
	case 1:
		return 0x10 + 0xFFF
	case 2:
		return 0x11 + 0xFFF
	case 3:
		return 0x12 + 0xFFF
	case 4:
		return 0xA0 + 0xFFF
	case 5:
		return 0xA1 + 0xFFF
	case 6:
		return 0xA2 + 0xFFF
	case 7:
		return 0xA3 + 0xFFF
	case 8:
		return 0x5B + 0xFFF
	case 9:
		return 0x5C + 0xFFF
	case 10:
		return 0x0002
	case 11:
		return 0x0008
	case 12:
		return 41
	case 13:
		return 12
	case 14:
		return 13
	case 15:
		return 26
	case 16:
		return 27

	case 17:
		return 39
	case 18:
		return 40
	case 19:
		return 43
	case 20:
		return 51
	case 21:
		return 52
	case 22:
		return 53
	case 23:
		return 86
	case 24:
		return 1
	case 25:
		return 2
	case 26:
		return 3
	case 27:
		return 4
	case 28:
		return 5
	case 29:
		return 6
	case 30:
		return 7
	case 31:
		return 8
	case 32:
		return 9
	case 33:
		return 10
	case 34:
		return 11
	case 35:
		return 16
	case 36:
		return 17
	case 37:
		return 18
	case 38:
		return 19
	case 39:
		return 20
	case 40:
		return 21
	case 41:
		return 22
	case 42:
		return 23
	case 43:
		return 24
	case 44:
		return 25
	case 45:
		return 30
	case 46:
		return 31
	case 47:
		return 32
	case 48:
		return 33
	case 49:
		return 34
	case 50:
		return 35
	case 51:
		return 36
	case 52:
		return 37
	case 53:
		return 38
	case 54:
		return 44
	case 55:
		return 45
	case 56:
		return 46
	case 57:
		return 47
	case 58:
		return 48
	case 59:
		return 49
	case 60:
		return 50
	case 61:
		return 59
	case 62:
		return 60
	case 63:
		return 61
	case 64:
		return 62
	case 65:
		return 63
	case 66:
		return 64
	case 67:
		return 65
	case 68:
		return 66
	case 69:
		return 67
	case 70:
		return 68
	case 71:
		return 87
	case 72:
		return 88
	case 73:
		return 0x7C + 0xFFF
	case 74:
		return 0x7D + 0xFFF
	case 75:
		return 0x7E + 0xFFF
	case 76:
		return 0x7F + 0xFFF
	case 77:
		return 0x80 + 0xFFF
	case 78:
		return 0x81 + 0xFFF
	case 79:
		return 0x82 + 0xFFF
	case 80:
		return 0x83 + 0xFFF
	case 81:
		return 0x84 + 0xFFF
	case 82:
		return 0x85 + 0xFFF
	case 83:
		return 0x86 + 0xFFF
	case 84:
		return 0x87 + 0xFFF
	case 85:
		return 69
	case 86:
		return 70
	case 87:
		return 0
	case 88:
		return 12
	case 89:
		return 13
	case 90:
		return 14
	case 91:
		return 15
	case 92:
		return 26
	case 93:
		return 27
	case 94:
		return 28
	case 95:
		return 39
	case 96:
		return 40
	case 97:
		return 41
	case 98:
		return 43
	case 99:
		return 51
	case 100:
		return 52
	case 101:
		return 53
	case 102:
		return 55
	case 103:
		return 57
	case 104:
		return 58
	case 105:
		return 82
	case 106:
		return 79
	case 107:
		return 80
	case 108:
		return 81
	case 109:
		return 75
	case 110:
		return 76
	case 111:
		return 77
	case 112:
		return 71
	case 113:
		return 72
	case 114:
		return 73
	case 115:
		return 74
	case 116:
		return 78
	case 117:
		return 83
	case 118:
		return 0x01 + 0xFFF
	case 119:
		return 0x02 + 0xFFF
	case 120:
		return 0x03 + 0xFFF
	case 121:
		return 0x04 + 0xFFF
	case 122:
		return 0x05 + 0xFFF
	case 123:
		return 0x06 + 0xFFF
	case 124:
		return 0x08 + 0xFFF
	case 125:
		return 0x0C + 0xFFF
	case 126:
		return 0x13 + 0xFFF
	case 127:
		return 0x14 + 0xFFF
	case 128:
		return 0x15 + 0xFFF
	case 129:
		return 0x15 + 0xFFF
	case 130:
		return 0x15 + 0xFFF
	case 131:
		return 0x17 + 0xFFF
	case 132:
		return 0x18 + 0xFFF
	case 133:
		return 0x19 + 0xFFF
	case 134:
		return 0x19 + 0xFFF
	case 135:
		return 0x1C + 0xFFF
	case 136:
		return 0x1D + 0xFFF
	case 137:
		return 0x1E + 0xFFF
	case 138:
		return 0x1F + 0xFFF
	case 139:
		return 0x21 + 0xFFF
	case 140:
		return 0x22 + 0xFFF
	case 141:
		return 0x23 + 0xFFF
	case 142:
		return 0x24 + 0xFFF
	case 143:
		return 0x25 + 0xFFF
	case 144:
		return 0x26 + 0xFFF
	case 145:
		return 0x27 + 0xFFF
	case 146:
		return 0x28 + 0xFFF
	case 147:
		return 0x29 + 0xFFF
	case 148:
		return 0x2A + 0xFFF
	case 149:
		return 0x2B + 0xFFF
	case 150:
		return 0x2C + 0xFFF
	case 151:
		return 0x2D + 0xFFF
	case 152:
		return 0x2E + 0xFFF
	case 153:
		return 0x2F + 0xFFF
	case 154:
		return 0x91 + 0xFFF
	case 155:
		return 0xA4 + 0xFFF
	case 156:
		return 0xA5 + 0xFFF
	case 157:
		return 0xA6 + 0xFFF
	case 158:
		return 0xA7 + 0xFFF
	case 159:
		return 0xA8 + 0xFFF
	case 160:
		return 0xA9 + 0xFFF
	case 161:
		return 0xAA + 0xFFF
	case 162:
		return 0xAB + 0xFFF
	case 163:
		return 0xAC + 0xFFF
	case 164:
		return 0xAD + 0xFFF
	case 165:
		return 0xAE + 0xFFF
	case 166:
		return 0xAF + 0xFFF
	case 167:
		return 0xB0 + 0xFFF
	case 168:
		return 0xB1 + 0xFFF
	case 169:
		return 0xB2 + 0xFFF
	case 170:
		return 0xB3 + 0xFFF
	case 171:
		return 0xB4 + 0xFFF
	case 172:
		return 0xB5 + 0xFFF
	case 173:
		return 0xB6 + 0xFFF
	case 174:
		return 0xB7 + 0xFFF
	case 175:
		return 0xBA + 0xFFF
	case 176:
		return 0xBB + 0xFFF
	case 177:
		return 0xBC + 0xFFF
	case 178:
		return 0xBD + 0xFFF
	case 179:
		return 0xBE + 0xFFF
	case 180:
		return 0xBF + 0xFFF
	case 181:
		return 0xC0 + 0xFFF
	case 182:
		return 0xDB + 0xFFF
	case 183:
		return 0xDC + 0xFFF
	case 184:
		return 0xDD + 0xFFF
	case 185:
		return 0xDE + 0xFFF
	case 186:
		return 0xDF + 0xFFF
	case 187:
		return 0xE2 + 0xFFF
	case 188:
		return 0xE5 + 0xFFF
	case 189:
		return 0xE7 + 0xFFF
	case 190:
		return 0xF6 + 0xFFF
	case 191:
		return 0xF7 + 0xFFF
	case 192:
		return 0xF8 + 0xFFF
	case 193:
		return 0xF9 + 0xFFF
	case 194:
		return 0xFA + 0xFFF
	case 195:
		return 0xFB + 0xFFF
	case 196:
		return 0xFC + 0xFFF
	case 197:
		return 0xFD + 0xFFF
	case 198:
		return 0xFE + 0xFFF
	default:
		return 0
	}
}

var key = [198]int{
	winapi._,
}

type keys uint8

//nolint:revive
const (
	_VK_SHIFT keys = iota + 1
	_VK_CTRL
	_VK_ALT
	_VK_LSHIFT
	_VK_RSHIFT
	_VK_LCONTROL
	_VK_RCONTROL
	_VK_LWIN
	_VK_RWIN
	_KEYEVENTF_KEYUP
	_KEYEVENTF_SCANCODE
	VK_SP1
	VK_SP2
	VK_SP3
	VK_SP4
	VK_SP5
	VK_SP6
	VK_SP7
	VK_SP8
	VK_SP9
	VK_SP10
	VK_SP11
	VK_SP12
	VK_ESC
	VK_1
	VK_2
	VK_3
	VK_4
	VK_5
	VK_6
	VK_7
	VK_8
	VK_9
	VK_0
	VK_Q
	VK_W
	VK_E
	VK_R
	VK_T
	VK_Y
	VK_U
	VK_I
	VK_O
	VK_P
	VK_A
	VK_S
	VK_D
	VK_F
	VK_G
	VK_H
	VK_J
	VK_K
	VK_L
	VK_Z
	VK_X
	VK_C
	VK_V
	VK_B
	VK_N
	VK_M
	VK_F1
	VK_F2
	VK_F3
	VK_F4
	VK_F5
	VK_F6
	VK_F7
	VK_F8
	VK_F9
	VK_F10
	VK_F11
	VK_F12
	VK_F13
	VK_F14
	VK_F15
	VK_F16
	VK_F17
	VK_F18
	VK_F19
	VK_F20
	VK_F21
	VK_F22
	VK_F23
	VK_F24
	VK_NUMLOCK
	VK_SCROLLLOCK
	VK_RESERVED
	VK_MINUS
	VK_EQUAL
	VK_BACKSPACE
	VK_TAB
	VK_LEFTBRACE
	VK_RIGHTBRACE
	VK_ENTER
	VK_SEMICOLON
	VK_APOSTROPHE
	VK_GRAVE
	VK_BACKSLASH
	VK_COMMA
	VK_DOT
	VK_SLASH
	VK_KPASTERISK
	VK_SPACE
	VK_CAPSLOCK
	VK_KP0
	VK_KP1
	VK_KP2
	VK_KP3
	VK_KP4
	VK_KP5
	VK_KP6
	VK_KP7
	VK_KP8
	VK_KP9
	VK_KPMINUS
	VK_KPPLUS
	VK_KPDOT
	VK_LBUTTON
	VK_RBUTTON
	VK_CANCEL
	VK_MBUTTON
	VK_XBUTTON1
	VK_XBUTTON2
	VK_BACK
	VK_CLEAR
	VK_PAUSE
	VK_CAPITAL
	VK_KANA
	VK_HANGUEL
	VK_HANGUL
	VK_JUNJA
	VK_FINAL
	VK_HANJA
	VK_KANJI
	VK_CONVERT
	VK_NONCONVERT
	VK_ACCEPT
	VK_MODECHANGE
	VK_PAGEUP
	VK_PAGEDOWN
	VK_END
	VK_HOME
	VK_LEFT
	VK_UP
	VK_RIGHT
	VK_DOWN
	VK_SELECT
	VK_PRINT
	VK_EXECUTE
	VK_SNAPSHOT
	VK_INSERT
	VK_DELETE
	VK_HELP
	VK_SCROLL
	VK_LMENU
	VK_RMENU
	VK_BROWSER_BACK
	VK_BROWSER_FORWARD
	VK_BROWSER_REFRESH
	VK_BROWSER_STOP
	VK_BROWSER_SEARCH
	VK_BROWSER_FAVORITES
	VK_BROWSER_HOME
	VK_VOLUME_MUTE
	VK_VOLUME_DOWN
	VK_VOLUME_UP
	VK_MEDIA_NEXT_TRACK
	VK_MEDIA_PREV_TRACK
	VK_MEDIA_STOP
	VK_MEDIA_PLAY_PAUSE
	VK_LAUNCH_MAIL
	VK_LAUNCH_MEDIA_SELECT
	VK_LAUNCH_APP1
	VK_LAUNCH_APP2
	VK_OEM_1
	VK_OEM_PLUS
	VK_OEM_COMMA
	VK_OEM_MINUS
	VK_OEM_PERIOD
	VK_OEM_2
	VK_OEM_3
	VK_OEM_4
	VK_OEM_5
	VK_OEM_6
	VK_OEM_7
	VK_OEM_8
	VK_OEM_102
	VK_PROCESSKEY
	VK_PACKET
	VK_ATTN
	VK_CRSEL
	VK_EXSEL
	VK_EREOF
	VK_PLAY
	VK_ZOOM
	VK_NONAME
	VK_PA1
	VK_OEM_CLEAR
)
