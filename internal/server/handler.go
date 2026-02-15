package server

import (
	"encoding/binary"
	"sync"
	"web-touchpad/internal/hid"
)

const (
	msgMouseMove  = 0x01
	msgMouseClick = 0x02
	msgScroll     = 0x03
	msgKeyboard   = 0x04
)

const (
	mouseBtnLeft  = 0x01
	mouseBtnRight = 0x02
	mouseBtnMid   = 0x04
)

const (
	mouseDown = 0x01
	mouseUp   = 0x00
)

type hidCtx struct {
	dev      *hid.Device
	btnState uint8
	mu       sync.Mutex
}

var ctx = &hidCtx{}

func dispatchMessage(dev *hid.Device, data []byte) {
	ctx.dev = dev

	if len(data) < 1 {
		return
	}

	switch data[0] {
	case msgMouseMove:
		handleMouseMove(data)
	case msgMouseClick:
		handleMouseClick(data)
	case msgScroll:
		handleScroll(data)
	case msgKeyboard:
		handleKeyboard(data)
	}
}

func handleMouseMove(data []byte) {
	if len(data) < 5 {
		return
	}
	x := int16(binary.LittleEndian.Uint16(data[1:3]))
	y := int16(binary.LittleEndian.Uint16(data[3:5]))

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	_ = ctx.dev.SendMouse(ctx.btnState, int8(x), int8(y), 0)
}

func handleMouseClick(data []byte) {
	if len(data) < 3 {
		return
	}
	btn := data[1]
	state := data[2]

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	switch btn {
	case mouseBtnLeft:
		if state == mouseDown {
			ctx.btnState |= mouseBtnLeft
		} else {
			ctx.btnState &^= mouseBtnLeft
		}
	case mouseBtnRight:
		if state == mouseDown {
			ctx.btnState |= mouseBtnRight
		} else {
			ctx.btnState &^= mouseBtnRight
		}
	case mouseBtnMid:
		if state == mouseDown {
			ctx.btnState |= mouseBtnMid
		} else {
			ctx.btnState &^= mouseBtnMid
		}
	}

	_ = ctx.dev.SendMouse(ctx.btnState, 0, 0, 0)
}

func handleScroll(data []byte) {
	if len(data) < 5 {
		return
	}
	y := int16(binary.LittleEndian.Uint16(data[3:5]))

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	_ = ctx.dev.SendMouse(ctx.btnState, 0, 0, int8(y))
}

func handleKeyboard(data []byte) {
	if len(data) < 5 {
		return
	}
	cp := rune(binary.LittleEndian.Uint32(data[1:5]))

	mod, key := mapRuneToHid(cp)
	if key == 0 {
		return
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	_ = ctx.dev.SendKeyboard(mod, []byte{key})
	_ = ctx.dev.SendKeyboard(0, nil)
}

func mapRuneToHid(r rune) (mod uint8, key uint8) {
	switch {
	case r >= 'a' && r <= 'z':
		return 0, uint8(r-'a') + 0x04
	case r >= 'A' && r <= 'Z':
		return 0x02, uint8(r-'A') + 0x04
	case r == ' ':
		return 0, 0x2C
	case r == '\n':
		return 0, 0x28
	case r == '\b':
		return 0, 0x2A
	}
	return 0, 0
}
