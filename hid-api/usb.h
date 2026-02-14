#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <ostream>
#include <new>

extern "C" {

void *usb_hid_init();

int32_t usb_hid_send_keyboard(void *ctx,
                              uint8_t modifiers,
                              const uint8_t *keys_ptr,
                              uintptr_t keys_len);

int32_t usb_hid_send_mouse(void *ctx, uint8_t buttons, int8_t x, int8_t y, int8_t wheel);

void usb_hid_free(void *ctx);

}  // extern "C"
