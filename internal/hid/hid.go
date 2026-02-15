package hid

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/hid
#cgo LDFLAGS: -L${SRCDIR}/../../third_party/hid -lhid_api
#include "hid.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type Device struct {
	ctx unsafe.Pointer
}

func Init() (*Device, error) {
	ctx := C.usb_hid_init()
	if ctx == nil {
		return nil, errors.New("usb_hid_init failed")
	}
	return &Device{ctx: ctx}, nil
}

func (d *Device) Close() {
	if d == nil || d.ctx == nil {
		return
	}
	C.usb_hid_free(d.ctx)
	d.ctx = nil
}

// modifiers: 修饰键位图
// keys: 最多6个键码（多余的会被Rust端截断）
func (d *Device) SendKeyboard(modifiers byte, keys []byte) error {
	if d == nil || d.ctx == nil {
		return errors.New("device not initialized")
	}

	var ptr *C.uchar
	var n C.size_t

	if len(keys) > 0 {
		ptr = (*C.uchar)(unsafe.Pointer(&keys[0]))
		n = C.size_t(len(keys))
	} else {
		ptr = nil
		n = 0
	}

	ret := C.usb_hid_send_keyboard(
		d.ctx,
		C.uchar(modifiers),
		ptr,
		n,
	)
	if ret != 0 {
		return errors.New("usb_hid_send_keyboard failed")
	}
	return nil
}

func (d *Device) SendMouse(buttons byte, x, y, wheel int8) error {
	if d == nil || d.ctx == nil {
		return errors.New("device not initialized")
	}

	ret := C.usb_hid_send_mouse(
		d.ctx,
		C.uchar(buttons),
		C.schar(x),
		C.schar(y),
		C.schar(wheel),
	)
	if ret != 0 {
		return errors.New("usb_hid_send_mouse failed")
	}
	return nil
}
