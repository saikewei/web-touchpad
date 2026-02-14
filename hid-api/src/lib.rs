mod usb_hid;

use std::ffi::c_void;
use std::sync::Mutex;

use anyhow::Result;
use tokio::runtime::Runtime;

use usb_hid::{UsbKeyboardHidDevice, UsbMouseHidDevice, build_usb_hid_device};

pub struct HidContext {
    rt: Runtime,
    kb: Mutex<UsbKeyboardHidDevice>,
    mouse: Mutex<UsbMouseHidDevice>,
}

#[unsafe(no_mangle)]
pub extern "C" fn usb_hid_init() -> *mut c_void {
    let rt = match Runtime::new() {
        Ok(rt) => rt,
        Err(_) => return std::ptr::null_mut(),
    };

    let res: Result<(
        UsbKeyboardHidDevice,
        UsbKeyboardHidDevice,
        UsbMouseHidDevice,
    )> = rt.block_on(build_usb_hid_device());

    let (kb, _kb_clone, mouse) = match res {
        Ok(v) => v,
        Err(_) => return std::ptr::null_mut(),
    };

    let ctx = HidContext {
        rt,
        kb: Mutex::new(kb),
        mouse: Mutex::new(mouse),
    };

    Box::into_raw(Box::new(ctx)) as *mut c_void
}

#[unsafe(no_mangle)]
pub extern "C" fn usb_hid_send_keyboard(
    ctx: *mut c_void,
    modifiers: u8,
    keys_ptr: *const u8,
    keys_len: usize,
) -> i32 {
    if ctx.is_null() {
        return -1;
    }
    let ctx = unsafe { &mut *(ctx as *mut HidContext) };
    let keys = unsafe { std::slice::from_raw_parts(keys_ptr, keys_len) };

    let mut kb = match ctx.kb.lock() {
        Ok(guard) => guard,
        Err(_) => return -2,
    };

    let r = ctx.rt.block_on(kb.send_keyboard_report(modifiers, keys));
    if r.is_ok() { 0 } else { -3 }
}

#[unsafe(no_mangle)]
pub extern "C" fn usb_hid_send_mouse(
    ctx: *mut c_void,
    buttons: u8,
    x: i8,
    y: i8,
    wheel: i8,
) -> i32 {
    if ctx.is_null() {
        return -1;
    }
    let ctx = unsafe { &mut *(ctx as *mut HidContext) };

    let mut mouse = match ctx.mouse.lock() {
        Ok(guard) => guard,
        Err(_) => return -2,
    };

    let r = ctx
        .rt
        .block_on(mouse.send_mouse_report(buttons, x, y, wheel));
    if r.is_ok() { 0 } else { -3 }
}

#[unsafe(no_mangle)]
pub extern "C" fn usb_hid_free(ctx: *mut c_void) {
    if ctx.is_null() {
        return;
    }
    unsafe {
        drop(Box::from_raw(ctx as *mut HidContext));
    }
}
