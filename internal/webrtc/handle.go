package webrtc

const (
	// Event
	OPMove    = 0x01
	OPScroll  = 0x02
	OPKeyDown = 0x03
	OPKeyUp   = 0x04
)

type PayloadHeader struct {
	Event  uint8
	Length uint16
}

type PayloadMove struct {
	PayloadHeader
	X uint16
	Y uint16
}

type PayloadScroll struct {
	PayloadHeader
	X int16
	Y int16
}

type PayloadKey struct {
	PayloadHeader
	Key uint32
}
