package hid

import (
	"testing"
	"time"
)

func TestHIDBasic(t *testing.T) {
	dev, err := Init()
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	defer dev.Close()

	// 发送：按下 A，再释放
	err = dev.SendKeyboard(0, []byte{0x04}) // KEY_A = 0x04
	if err != nil {
		t.Fatalf("send keyboard down failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	err = dev.SendKeyboard(0, []byte{})
	if err != nil {
		t.Fatalf("send keyboard up failed: %v", err)
	}

	// 移动鼠标
	for i := 0; i < 10; i++ {
		err = dev.SendMouse(0, 0, -5, 0)
		if err != nil {
			t.Fatalf("send mouse failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
