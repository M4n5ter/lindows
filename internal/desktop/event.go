package desktop

import "github.com/m4n5ter/lindows/winapi"

//nolint:unused
func mapping(i uint8) int {
	return key[i]
}

//nolint:unused
var key = [198]int{
	winapi.VK_SHIFT,
	winapi.VK_CTRL,
	winapi.VK_ALT,
	winapi.VK_LSHIFT,
	winapi.VK_RSHIFT,
	winapi.VK_LCONTROL,
	winapi.VK_RCONTROL,
	winapi.VK_LWIN,
	winapi.VK_RWIN,
	winapi.KEYEVENTF_KEYUP,
	winapi.KEYEVENTF_SCANCODE,
	winapi.VK_SP1,
	winapi.VK_SP2,
	winapi.VK_SP3,
	winapi.VK_SP4,
	winapi.VK_SP5,
	winapi.VK_SP6,
	winapi.VK_SP7,
	winapi.VK_SP8,
	winapi.VK_SP9,
	winapi.VK_SP10,
	winapi.VK_SP11,
	winapi.VK_SP12,
	winapi.VK_ESC,
	winapi.VK_1,
	winapi.VK_2,
	winapi.VK_3,
	winapi.VK_4,
	winapi.VK_5,
	winapi.VK_6,
	winapi.VK_7,
	winapi.VK_8,
	winapi.VK_9,
	winapi.VK_0,
	winapi.VK_Q,
	winapi.VK_W,
	winapi.VK_E,
	winapi.VK_R,
	winapi.VK_T,
	winapi.VK_Y,
	winapi.VK_U,
	winapi.VK_I,
	winapi.VK_O,
	winapi.VK_P,
	winapi.VK_A,
	winapi.VK_S,
	winapi.VK_D,
	winapi.VK_F,
	winapi.VK_G,
	winapi.VK_H,
	winapi.VK_J,
	winapi.VK_K,
	winapi.VK_L,
	winapi.VK_Z,
	winapi.VK_X,
	winapi.VK_C,
	winapi.VK_V,
	winapi.VK_B,
	winapi.VK_N,
	winapi.VK_M,
	winapi.VK_F1,
	winapi.VK_F2,
	winapi.VK_F3,
	winapi.VK_F4,
	winapi.VK_F5,
	winapi.VK_F6,
	winapi.VK_F7,
	winapi.VK_F8,
	winapi.VK_F9,
	winapi.VK_F10,
	winapi.VK_F11,
	winapi.VK_F12,
	winapi.VK_F13,
	winapi.VK_F14,
	winapi.VK_F15,
	winapi.VK_F16,
	winapi.VK_F17,
	winapi.VK_F18,
	winapi.VK_F19,
	winapi.VK_F20,
	winapi.VK_F21,
	winapi.VK_F22,
	winapi.VK_F23,
	winapi.VK_F24,
	winapi.VK_NUMLOCK,
	winapi.VK_SCROLLLOCK,
	winapi.VK_RESERVED,
	winapi.VK_MINUS,
	winapi.VK_EQUAL,
	winapi.VK_BACKSPACE,
	winapi.VK_TAB,
	winapi.VK_LEFTBRACE,
	winapi.VK_RIGHTBRACE,
	winapi.VK_ENTER,
	winapi.VK_SEMICOLON,
	winapi.VK_APOSTROPHE,
	winapi.VK_GRAVE,
	winapi.VK_BACKSLASH,
	winapi.VK_COMMA,
	winapi.VK_DOT,
	winapi.VK_SLASH,
	winapi.VK_KPASTERISK,
	winapi.VK_SPACE,
	winapi.VK_CAPSLOCK,
	winapi.VK_KP0,
	winapi.VK_KP1,
	winapi.VK_KP2,
	winapi.VK_KP3,
	winapi.VK_KP4,
	winapi.VK_KP5,
	winapi.VK_KP6,
	winapi.VK_KP7,
	winapi.VK_KP8,
	winapi.VK_KP9,
	winapi.VK_KPMINUS,
	winapi.VK_KPPLUS,
	winapi.VK_KPDOT,
	winapi.VK_LBUTTON,
	winapi.VK_RBUTTON,
	winapi.VK_CANCEL,
	winapi.VK_MBUTTON,
	winapi.VK_XBUTTON1,
	winapi.VK_XBUTTON2,
	winapi.VK_BACK,
	winapi.VK_CLEAR,
	winapi.VK_PAUSE,
	winapi.VK_CAPITAL,
	winapi.VK_KANA,
	winapi.VK_HANGUEL,
	winapi.VK_HANGUL,
	winapi.VK_JUNJA,
	winapi.VK_FINAL,
	winapi.VK_HANJA,
	winapi.VK_KANJI,
	winapi.VK_CONVERT,
	winapi.VK_NONCONVERT,
	winapi.VK_ACCEPT,
	winapi.VK_MODECHANGE,
	winapi.VK_PAGEUP,
	winapi.VK_PAGEDOWN,
	winapi.VK_END,
	winapi.VK_HOME,
	winapi.VK_LEFT,
	winapi.VK_UP,
	winapi.VK_RIGHT,
	winapi.VK_DOWN,
	winapi.VK_SELECT,
	winapi.VK_PRINT,
	winapi.VK_EXECUTE,
	winapi.VK_SNAPSHOT,
	winapi.VK_INSERT,
	winapi.VK_DELETE,
	winapi.VK_HELP,
	winapi.VK_SCROLL,
	winapi.VK_LMENU,
	winapi.VK_RMENU,
	winapi.VK_BROWSER_BACK,
	winapi.VK_BROWSER_FORWARD,
	winapi.VK_BROWSER_REFRESH,
	winapi.VK_BROWSER_STOP,
	winapi.VK_BROWSER_SEARCH,
	winapi.VK_BROWSER_FAVORITES,
	winapi.VK_BROWSER_HOME,
	winapi.VK_VOLUME_MUTE,
	winapi.VK_VOLUME_DOWN,
	winapi.VK_VOLUME_UP,
	winapi.VK_MEDIA_NEXT_TRACK,
	winapi.VK_MEDIA_PREV_TRACK,
	winapi.VK_MEDIA_STOP,
	winapi.VK_MEDIA_PLAY_PAUSE,
	winapi.VK_LAUNCH_MAIL,
	winapi.VK_LAUNCH_MEDIA_SELECT,
	winapi.VK_LAUNCH_APP1,
	winapi.VK_LAUNCH_APP2,
	winapi.VK_OEM_1,
	winapi.VK_OEM_PLUS,
	winapi.VK_OEM_COMMA,
	winapi.VK_OEM_MINUS,
	winapi.VK_OEM_PERIOD,
	winapi.VK_OEM_2,
	winapi.VK_OEM_3,
	winapi.VK_OEM_4,
	winapi.VK_OEM_5,
	winapi.VK_OEM_6,
	winapi.VK_OEM_7,
	winapi.VK_OEM_8,
	winapi.VK_OEM_102,
	winapi.VK_PROCESSKEY,
	winapi.VK_PACKET,
	winapi.VK_ATTN,
	winapi.VK_CRSEL,
	winapi.VK_EXSEL,
	winapi.VK_EREOF,
	winapi.VK_PLAY,
	winapi.VK_ZOOM,
	winapi.VK_NONAME,
	winapi.VK_PA1,
	winapi.VK_OEM_CLEAR,
}

//nolint:revive
const (
	VK_SHIFT uint8 = iota
	VK_CTRL
	VK_ALT
	VK_LSHIFT
	VK_RSHIFT
	VK_LCONTROL
	VK_RCONTROL
	VK_LWIN
	VK_RWIN
	KEYEVENTF_KEYUP
	KEYEVENTF_SCANCODE
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
