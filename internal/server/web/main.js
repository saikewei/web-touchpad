// --- 配置与状态 ---
const SENSITIVITY = 2.5; // 鼠标灵敏度
const SCROLL_SENSITIVITY = 0.1; // 滚轮灵敏度
const WS_URL = `ws://${window.location.host}/ws`;

const TAP_MAX_TIME = 200; // ms
const TAP_MAX_MOVE = 10;  // px

let tapStartX = 0;
let tapStartY = 0;
let tapStartTime = 0;
let tapCandidate = false;

// 消息类型定义
const MSG_TYPE = {
  MOUSE_MOVE: 0x01, // 鼠标移动
  MOUSE_CLICK: 0x02, // 鼠标点击
  SCROLL: 0x03, // 滚轮
  KEYBOARD: 0x04, // 键盘
};

const MOUSE_BUTTON = {
  LEFT: 0x01,
  RIGHT: 0x02,
  MIDDLE: 0x03,
};

const MOUSE_STATE = {
  DOWN: 0x01,
  UP: 0x00,
};

let ws = null;
let lastX = 0;
let lastY = 0;
let lastDistance = 0; // 双指距离
let isScrollMode = false; // 是否为滚动模式
let isKeyboardActive = false;
let retryCount = 0;

// 获取 DOM 元素
const statusEl = document.getElementById("status-bar");
const touchZone = document.getElementById("touch-zone");
const btnLeft = document.getElementById("btn-left");
const btnRight = document.getElementById("btn-right");
const btnKeyboard = document.getElementById("btn-keyboard");
const hiddenInput = document.getElementById("hidden-input");

// --- WebSocket 连接逻辑 ---
function connect() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return;
  }

  ws = new WebSocket(WS_URL);
  ws.binaryType = "arraybuffer";

  ws.onopen = () => {
    statusEl.textContent = "🟢 已连接";
    statusEl.className = "connected";
    retryCount = 0;
  };

  ws.onclose = (e) => {
    statusEl.textContent = "🔴 已断开，尝试重连...";
    statusEl.className = "disconnected";

    // 如果是“被新连接踢掉”，不要重连
    if (e.code === 4001) {
      console.log("WS 被新连接替换，停止重连");
      return;
    }

    if (!shouldReconnect) {
      return;
    }

    const delay = Math.min(Math.pow(2, retryCount) * 1000, 10000);
    setTimeout(() => {
      retryCount++;
      connect();
    }, delay);
  };

  ws.onerror = (err) => {
    console.error("WS 错误:", err);
  };
}

// --- 二进制消息构造函数 ---

// 鼠标移动: [type(1), x(2), y(2)] = 5 bytes
function createMouseMoveMsg(x, y) {
  const buffer = new ArrayBuffer(5);
  const view = new DataView(buffer);
  view.setUint8(0, MSG_TYPE.MOUSE_MOVE);
  view.setInt16(1, x, true); // little-endian
  view.setInt16(3, y, true);
  return buffer;
}

// 鼠标点击: [type(1), button(1), state(1)] = 3 bytes
function createMouseClickMsg(button, state) {
  const buffer = new ArrayBuffer(3);
  const view = new DataView(buffer);
  view.setUint8(0, MSG_TYPE.MOUSE_CLICK);
  view.setUint8(1, button);
  view.setUint8(2, state);
  return buffer;
}

// 滚轮: [type(1), x(2), y(2)] = 5 bytes
function createScrollMsg(x, y) {
  const buffer = new ArrayBuffer(5);
  const view = new DataView(buffer);
  view.setUint8(0, MSG_TYPE.SCROLL);
  view.setInt16(1, x, true);
  view.setInt16(3, y, true);
  return buffer;
}

// 键盘: [type(1), keyCode(4)] = 5 bytes (使用 UTF-32)
function createKeyboardMsg(char) {
  const buffer = new ArrayBuffer(5);
  const view = new DataView(buffer);
  view.setUint8(0, MSG_TYPE.KEYBOARD);

  if (typeof char === "string" && char.length > 0) {
    view.setUint32(1, char.codePointAt(0), true);
  } else {
    view.setUint32(1, 0, true);
  }
  return buffer;
}

// 发送二进制数据
function send(buffer) {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(buffer);
  }
  // 调试用：显示发送的数据
  const view = new DataView(buffer);
  const type = view.getUint8(0);
  // console.log(
  //   `发送消息类型: 0x${type.toString(16).padStart(2, "0")}, 长度: ${buffer.byteLength}`,
  // );
}

// 计算两个触摸点之间的中心点
function getTouchCenter(touch1, touch2) {
  return {
    x: (touch1.clientX + touch2.clientX) / 2,
    y: (touch1.clientY + touch2.clientY) / 2,
  };
}

// 计算两个触摸点之间的距离
function getTouchDistance(touch1, touch2) {
  const dx = touch1.clientX - touch2.clientX;
  const dy = touch1.clientY - touch2.clientY;
  return Math.sqrt(dx * dx + dy * dy);
}

// --- 触控板逻辑 ---
touchZone.addEventListener(
  "touchstart",
  (e) => {
    if (e.touches.length === 1) {
      isScrollMode = false;
      lastX = e.touches[0].clientX;
      lastY = e.touches[0].clientY;

      // 轻点候选开始
      tapStartX = lastX;
      tapStartY = lastY;
      tapStartTime = Date.now();
      tapCandidate = true;
    } else if (e.touches.length === 2) {
      // 双指模式：滚轮
      isScrollMode = true;
      const center = getTouchCenter(e.touches[0], e.touches[1]);
      lastX = center.x;
      lastY = center.y;
      lastDistance = getTouchDistance(e.touches[0], e.touches[1]);
    }
  },
  { passive: false },
);

touchZone.addEventListener(
  "touchmove",
  (e) => {
    e.preventDefault();

    // 轻点候选判定：移动过大则取消
    if (tapCandidate && e.touches.length === 1) {
      const dx = e.touches[0].clientX - tapStartX;
      const dy = e.touches[0].clientY - tapStartY;
      if (Math.hypot(dx, dy) > TAP_MAX_MOVE) {
        tapCandidate = false;
      }
    }

    if (e.touches.length === 1 && !isScrollMode) {
      // 单指移动：鼠标指针
      const currentX = e.touches[0].clientX;
      const currentY = e.touches[0].clientY;

      const deltaX = (currentX - lastX) * SENSITIVITY;
      const deltaY = (currentY - lastY) * SENSITIVITY;

      if (Math.abs(deltaX) > 0.5 || Math.abs(deltaY) > 0.5) {
        send(createMouseMoveMsg(Math.round(deltaX), Math.round(deltaY)));
      }

      lastX = currentX;
      lastY = currentY;
    } else if (e.touches.length === 2) {
      // 双指滑动：滚轮
      const center = getTouchCenter(e.touches[0], e.touches[1]);
      const currentDistance = getTouchDistance(e.touches[0], e.touches[1]);

      const deltaY = (center.y - lastY) * SCROLL_SENSITIVITY;
      const deltaX = (center.x - lastX) * SCROLL_SENSITIVITY;

      if (Math.abs(deltaY) > 0.5 || Math.abs(deltaX) > 0.5) {
        send(
          createScrollMsg(
            Math.round(deltaX),
            Math.round(-deltaY),
          ),
        );
      }

      lastX = center.x;
      lastY = center.y;
      lastDistance = currentDistance;
    }
  },
  { passive: false },
);

touchZone.addEventListener(
  "touchend",
  (e) => {
    // 轻点判定：单指结束且未移动过大
    if (tapCandidate && e.touches.length === 0) {
      const dt = Date.now() - tapStartTime;
      if (dt <= TAP_MAX_TIME) {
        send(createMouseClickMsg(MOUSE_BUTTON.LEFT, MOUSE_STATE.DOWN));
        send(createMouseClickMsg(MOUSE_BUTTON.LEFT, MOUSE_STATE.UP));
      }
    }
    tapCandidate = false;

    // 重置状态
    if (e.touches.length < 2) {
      isScrollMode = false;
    }
    if (e.touches.length === 1) {
      lastX = e.touches[0].clientX;
      lastY = e.touches[0].clientY;
    }
  },
  { passive: false },
);

hiddenInput.addEventListener("input", (e) => {
  const char = e.data;
  if (char) {
    send(createKeyboardMsg(char));
  }
  hiddenInput.value = "";
});

hiddenInput.addEventListener("keydown", (e) => {
  if (e.key === "Backspace") {
    send(createKeyboardMsg("\b")); // Backspace 用 \b 表示
  } else if (e.key === "Enter") {
    send(createKeyboardMsg("\n")); // Enter 用 \n 表示
  }
});

// 页面离开时不重连，并主动关闭
window.addEventListener("pagehide", () => {
  shouldReconnect = false;
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.close(1000, "pagehide");
  }
});


// 初始化
connect();
