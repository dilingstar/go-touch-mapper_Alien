# 测试用脚本
# 打开串口然后进行读写
import serial
import time
import threading
import socket

CMD_DS5 = 0
CMD_KEYBOARD = 1
CMD_MOUSE = 2
CMD_TOUCH_REPORT = 3

#============MOUSE_KEY_DEFINES==============================================

MouseBtnLeft    = 1 << 0
MouseBtnRight   = 1 << 1
MouseBtnMiddle  = 1 << 2
MouseBtnBack    = 1 << 3
MouseBtnForward = 1 << 4

KeyLeftCtrl   = 0xe0
KeyLeftShift  = 0xe1
KeyLeftAlt    = 0xe2
KeyLeftGui    = 0xe3
KeyRightCtrl  = 0xe4
KeyRightShift = 0xe5
KeyRightAlt   = 0xe6
KeyRightGui   = 0xe8

KeyA = 0x04; KeyB = 0x05; KeyC = 0x06; KeyD = 0x07
KeyE = 0x08; KeyF = 0x09; KeyG = 0x0A; KeyH = 0x0B
KeyI = 0x0C; KeyJ = 0x0D; KeyK = 0x0E; KeyL = 0x0F
KeyM = 0x10; KeyN = 0x11; KeyO = 0x12; KeyP = 0x13
KeyQ = 0x14; KeyR = 0x15; KeyS = 0x16; KeyT = 0x17
KeyU = 0x18; KeyV = 0x19; KeyW = 0x1A; KeyX = 0x1B
KeyY = 0x1C; KeyZ = 0x1D

Key1 = 0x1E; Key2 = 0x1F; Key3 = 0x20; Key4 = 0x21
Key5 = 0x22; Key6 = 0x23; Key7 = 0x24; Key8 = 0x25
Key9 = 0x26; Key0 = 0x27

KeyReturn    = 0x28; KeyEnter     = 0x28
KeyEsc       = 0x29; KeyEscape    = 0x29
KeyBckspc    = 0x2A; KeyBackspace = 0x2A
KeyTab       = 0x2B
KeySpace     = 0x2C
KeyMinus     = 0x2D; KeyDash      = 0x2D
KeyEquals    = 0x2E; KeyEqual     = 0x2E
KeyLbracket  = 0x2F
KeyRbracket  = 0x30
KeyBackslash = 0x31
KeyHash      = 0x32; KeyNumber    = 0x32
KeySemicolon = 0x33
KeyQuote     = 0x34
KeyBackquote = 0x35; KeyTilde     = 0x35
KeyComma     = 0x36
KeyPeriod    = 0x37; KeyStop      = 0x37
KeySlash     = 0x38
KeyCapsLock  = 0x39

KeyF1  = 0x3A; KeyF2  = 0x3B; KeyF3  = 0x3C; KeyF4  = 0x3D
KeyF5  = 0x3E; KeyF6  = 0x3F; KeyF7  = 0x40; KeyF8  = 0x41
KeyF9  = 0x42; KeyF10 = 0x43; KeyF11 = 0x44; KeyF12 = 0x45

KeyPrint       = 0x46
KeyScrollLock  = 0x47
KeyPause       = 0x48
KeyInsert      = 0x49
KeyHome        = 0x4A
KeyPageup      = 0x4B; KeyPgup      = 0x4B
KeyDel         = 0x4C; KeyDelete    = 0x4C
KeyEnd         = 0x4D
KeyPagedown    = 0x4E; KeyPgdown    = 0x4E
KeyRight       = 0x4F
KeyLeft        = 0x50
KeyDown        = 0x51
KeyUp          = 0x52
KeyNumLock     = 0x53

KeyKpDivide    = 0x54; KeyKpMultiply = 0x55
KeyKpMinus     = 0x56; KeyKpPlus     = 0x57
KeyKpEnter     = 0x58; KeyKpReturn   = 0x58
KeyKp1 = 0x59; KeyKp2 = 0x5A; KeyKp3 = 0x5B
KeyKp4 = 0x5C; KeyKp5 = 0x5D; KeyKp6 = 0x5E
KeyKp7 = 0x5F; KeyKp8 = 0x60; KeyKp9 = 0x61
KeyKp0         = 0x62
KeyKpPeriod    = 0x63; KeyKpStop    = 0x63

KeyApplication = 0x65
KeyPower       = 0x66
KeyKpEquals    = 0x67; KeyKpEqual   = 0x67

KeyF13 = 0x68; KeyF14 = 0x69; KeyF15 = 0x6A; KeyF16 = 0x6B
KeyF17 = 0x6C; KeyF18 = 0x6D; KeyF19 = 0x6E; KeyF20 = 0x6F
KeyF21 = 0x70; KeyF22 = 0x71; KeyF23 = 0x72; KeyF24 = 0x73

KeyExecute     = 0x74; KeyHelp      = 0x75
KeyMenu        = 0x76; KeySelect    = 0x77
KeyCancel      = 0x78; KeyRedo      = 0x79
KeyUndo        = 0x7A; KeyCut       = 0x7B
KeyCopy        = 0x7C; KeyPaste     = 0x7D
KeyFind        = 0x7E; KeyMute      = 0x7F
KeyVolumeUp    = 0x80; KeyVolumeDown = 0x81

#==========================================================

import ctypes

# Hat switch 常量
DS_HAT_UP        = 0
DS_HAT_UP_RIGHT  = 1
DS_HAT_RIGHT     = 2
DS_HAT_DOWN_RIGHT = 3
DS_HAT_DOWN      = 4
DS_HAT_DOWN_LEFT = 5
DS_HAT_LEFT      = 6
DS_HAT_UP_LEFT   = 7
DS_HAT_NULL      = 8

class TouchPoint(ctypes.LittleEndianStructure):
    _pack_ = 1
    _fields_ = [
        ("contact", ctypes.c_uint8),
        ("raw",     ctypes.c_uint8 * 3),
    ]

    @property
    def x(self):
        return (self.raw[0] | (self.raw[1] & 0x0F) << 8) & 0xFFF

    @x.setter
    def x(self, val):
        self.raw[0] = val & 0xFF
        self.raw[1] = (self.raw[1] & 0xF0) | ((val >> 8) & 0x0F)

    @property
    def y(self):
        return ((self.raw[1] >> 4) | (self.raw[2] << 4)) & 0xFFF

    @y.setter
    def y(self, val):
        self.raw[1] = (self.raw[1] & 0x0F) | ((val & 0x0F) << 4)
        self.raw[2] = (val >> 4) & 0xFF

class DsButtons(ctypes.LittleEndianStructure):
    _pack_ = 1
    _fields_ = [
        ("dpad",     ctypes.c_uint32, 4),
        ("x",        ctypes.c_uint32, 1),
        ("a",        ctypes.c_uint32, 1),
        ("b",        ctypes.c_uint32, 1),
        ("y",        ctypes.c_uint32, 1),
        ("lb",       ctypes.c_uint32, 1),
        ("rb",       ctypes.c_uint32, 1),
        ("lt",       ctypes.c_uint32, 1),
        ("rt",       ctypes.c_uint32, 1),
        ("back",     ctypes.c_uint32, 1),
        ("start",    ctypes.c_uint32, 1),
        ("ls",       ctypes.c_uint32, 1),
        ("rs",       ctypes.c_uint32, 1),
        ("ps",       ctypes.c_uint32, 1),
        ("touchpad", ctypes.c_uint32, 1),
        ("mute",     ctypes.c_uint32, 1),
        ("reserved", ctypes.c_uint32, 13),
    ]

class DualSenseInputReport(ctypes.LittleEndianStructure):
    _pack_ = 1
    _fields_ = [
        ("ls_x",            ctypes.c_uint8),
        ("ls_y",            ctypes.c_uint8),
        ("rs_x",            ctypes.c_uint8),
        ("rs_y",            ctypes.c_uint8),
        ("lt",              ctypes.c_uint8),
        ("rt",              ctypes.c_uint8),
        ("seq_number",      ctypes.c_uint8),
        ("buttons",         DsButtons),
        ("reserved",        ctypes.c_uint32),
        ("gyro_x",          ctypes.c_uint16),
        ("gyro_y",          ctypes.c_uint16),
        ("gyro_z",          ctypes.c_uint16),
        ("accel_x",         ctypes.c_uint16),
        ("accel_y",         ctypes.c_uint16),
        ("accel_z",         ctypes.c_uint16),
        ("sensor_timestamp",ctypes.c_uint32),
        ("reserved2",       ctypes.c_uint8),
        ("points_1",        TouchPoint),
        ("points_2",        TouchPoint),
        ("reserved3",       ctypes.c_uint8 * 12),
        ("status",          ctypes.c_uint8),
        ("reserved4",       ctypes.c_uint8 * 10),
    ]

    def __init__(self):
        super().__init__()
        self.ls_x = 0x7f
        self.ls_y = 0x7d
        self.rs_x = 0x7f
        self.rs_y = 0x7e
        self.buttons.dpad = DS_HAT_NULL
        self.points_1.contact = 0xFF
        self.points_2.contact = 0xFF


class tusb_ctrl:
    def __init__(self,screen_width=3440,screen_height=1440,sender=None):
        self.mouse_btn_bits = 0x00
        self.keyboard_modifier = 0x00
        self.keyboard_keycodes = [0] * 6  # 最多同时按下6个键
        self.screen_width = screen_width
        self.screen_height = screen_height
        self.dsreport = DualSenseInputReport()
        self.dsreport.ls_x = 0x7f
        self.dsreport.ls_y = 0x7d
        self.dsreport.rs_x = 0x7f
        self.dsreport.rs_y = 0x7e
        self.dsreport.buttons.dpad = DS_HAT_NULL
        self.dsreport.points_1.contact = 0xFF
        self.dsreport.points_2.contact = 0xFF
        self.sender = sender
        assert self.sender is not None, "sender must be provided"
    # --- 鼠标 ---

    def mouse_down(self, btn):
        """按下鼠标按钮: MouseBtnLeft / MouseBtnRight / ..."""
        self.mouse_btn_bits |= btn
        self._send_mouse()

    def mouse_up(self, btn):
        """释放鼠标按钮"""
        self.mouse_btn_bits &= ~btn
        self._send_mouse()

    def mouse_move(self, x=0, y=0, wheel=0):
        """鼠标移动, x/y/wheel 为 int8 (-128~127)"""
        payload = bytearray([CMD_MOUSE,
                             self.mouse_btn_bits,
                             x & 0xFF, y & 0xFF,
                             wheel & 0xFF, 0x00])
        self.sender(payload)

    def _send_mouse(self):
        payload = bytearray([CMD_MOUSE,
                             self.mouse_btn_bits,
                             0x00, 0x00, 0x00, 0x00])
        self.sender(payload)

    # --- 键盘 ---

    def _modifier_key_to_bit(self, key):
        """将修饰键 keycode 映射到 modifier bitmask"""
        m = {
            KeyLeftCtrl:   1 << 0,
            KeyLeftShift:  1 << 1,
            KeyLeftAlt:    1 << 2,
            KeyLeftGui:    1 << 3,
            KeyRightCtrl:  1 << 4,
            KeyRightShift: 1 << 5,
            KeyRightAlt:   1 << 6,
            KeyRightGui:   1 << 7,
        }
        return m.get(key, None)

    def keyboard_down(self, key):
        """按下键盘键"""
        bit = self._modifier_key_to_bit(key)
        if bit is not None:
            self.keyboard_modifier |= bit
        else:
            if key not in self.keyboard_keycodes and 0x00 in self.keyboard_keycodes:
                idx = self.keyboard_keycodes.index(0x00)
                self.keyboard_keycodes[idx] = key
        self._send_keyboard()

    def keyboard_up(self, key):
        """释放键盘键"""
        bit = self._modifier_key_to_bit(key)
        if bit is not None:
            self.keyboard_modifier &= ~bit
        else:
            if key in self.keyboard_keycodes:
                idx = self.keyboard_keycodes.index(key)
                self.keyboard_keycodes[idx] = 0x00
        self._send_keyboard()

    def keyboard_release_all(self):
        """释放所有按键"""
        self.keyboard_modifier = 0x00
        self.keyboard_keycodes = [0] * 6
        self._send_keyboard()

    def _send_keyboard(self):
        payload = bytearray([CMD_KEYBOARD, self.keyboard_modifier, 0x00])
        payload.extend(self.keyboard_keycodes)
        self.sender(payload)
        
    def touch_down(self,tid,x,y):
        """
        发送触摸按下包
        :param tid: 触摸ID (0-3)
        :param x: 触摸X坐标 (0-2^32-1)
        :param y: 触摸Y坐标 (0-2^32-1)
        :return: None
        """
        bytes = bytearray([CMD_TOUCH_REPORT, 0x01, tid])
        bytes.extend(x.to_bytes(4, 'little'))
        bytes.extend(y.to_bytes(4, 'little'))
        bytes.append(0x00)
        self.sender(bytes)

    def touch_up(self,tid):
        """
        发送触摸抬起包
        :param tid: 触摸ID (0-3)
        :return: None
        """
        bytes = bytearray([CMD_TOUCH_REPORT, 0x00, tid])
        bytes.extend((0).to_bytes(4, 'little'))
        bytes.extend((0).to_bytes(4, 'little'))
        bytes.append(0x00)
        self.sender(bytes)


    def touch_click_screen(self,x,y):
        """
        发送触摸点击包
        :param x: 触摸X坐标 (0~1 float)
        :param y: 触摸Y坐标 (0~1 float)
        :return: None
        """
        x = int(x/self.screen_width * 0x7ffffffe)
        y = int(y/self.screen_height * 0x7ffffffe)
        self.touch_down(0,x,y)
        time.sleep(0.01)
        self.touch_up(0)
        time.sleep(0.01)
    # --- DS5 手柄 ---

    def send_ds5(self):
        """发送当前 dsreport 的完整报告"""
        data = bytes(self.dsreport)
        payload = bytearray([CMD_DS5])
        payload.extend(data)
        self.sender(payload)

    def ds5_reset(self):
        """重置手柄到中立状态"""
        self.dsreport.ls_x = 0x7f
        self.dsreport.ls_y = 0x7d
        self.dsreport.rs_x = 0x7f
        self.dsreport.rs_y = 0x7e
        self.dsreport.lt = 0
        self.dsreport.rt = 0
        self.dsreport.buttons.dpad = DS_HAT_NULL
        self.dsreport.buttons.x = 0
        self.dsreport.buttons.a = 0
        self.dsreport.buttons.b = 0
        self.dsreport.buttons.y = 0
        self.dsreport.buttons.lb = 0
        self.dsreport.buttons.rb = 0
        self.dsreport.buttons.lt = 0
        self.dsreport.buttons.rt = 0
        self.dsreport.buttons.back = 0
        self.dsreport.buttons.start = 0
        self.dsreport.buttons.ls = 0
        self.dsreport.buttons.rs = 0
        self.dsreport.buttons.ps = 0
        self.dsreport.buttons.touchpad = 0
        self.dsreport.buttons.mute = 0
        self.send_ds5()

    def hid_interface_reset(self):
        """重置手柄接口"""
        payload = bytearray([0xFE,0x02,0xFF])
        self.sender(payload)

def pack_hid_frame(payload):
    """
    将原始载荷打包为 0x55 0xAA 协议格式
    :param payload: bytes 或 list 类型
    :return: bytes 对象
    """
    p_len = len(payload)
    if p_len > 0xff:
        raise ValueError("Payload length exceeds 255 bytes")
    frame = bytearray([0x55, 0xAA, p_len])
    frame.extend(payload)
    checksum = p_len
    for b in payload:
        checksum ^= b
    frame.append(checksum)
    print(bytes(frame).hex())
    return bytes(frame)

def make_udp_sender(ip,port):
    udp_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    def sender(payload):
        udp_socket.sendto(pack_hid_frame(payload), (ip, port))
    return sender


def make_serial_sender(com_port,baudrate):
    ser = serial.Serial(com_port, baudrate)  
    def reader():
        while True:
            data = ser.readline()
            try:
                print(data.decode().strip())
            except:
                print(data)
    thread = threading.Thread(target=reader)
    thread.start()
    def sender(payload):
        ser.write(pack_hid_frame(payload))
        ser.flush()
    return sender


if __name__ == "__main__":
    RATE = 1500#设备hid报告率
    COM_PORT = 'COM4'#esp32的额外控制UART
    BAUDRATE = 115200#esp32的额外控制UART波特率
    UDP_TARGET_IP = '192.168.3.255' #使用广播地址
    UDP_TARGET_PORT = 61068#esp32的UDP端口号
    mk = tusb_ctrl(sender=make_udp_sender(UDP_TARGET_IP,UDP_TARGET_PORT))#有状态的控制信号，比如按键的按下抬起，在UDP模式下建议使用较大延迟，产生丢包则会导致按键重复
    # mk = tusb_ctrl(sender=make_serial_sender(COM_PORT,BAUDRATE))
    # mk = tusb_ctrl(sender=make_serial_sender('COM8',115200))
    # # #======测试键盘======
    # hello_keys = [KeyH, KeyE, KeyL, KeyL, KeyO]
    # for key in hello_keys:
    #     mk.keyboard_down(key)
    #     time.sleep(1/RATE)
    #     mk.keyboard_up(key)
    #     time.sleep(1/RATE)
    # time.sleep(0.5)
    # # #======测试鼠标移动======
    # mk.mouse_down(MouseBtnLeft)
    # time.sleep(1/RATE)
    # for i in range(3000):
    # # while True:
    #     mk.mouse_move(0,3)
    #     time.sleep(1/RATE)
    #     mk.mouse_move(3,0)
    #     time.sleep(1/RATE)
    #     mk.mouse_move(0,-3)
    #     time.sleep(1/RATE)
    #     mk.mouse_move(-3,0)
    #     time.sleep(1/RATE)
    # # mk.mouse_up(MouseBtnLeft)
    # time.sleep(1/RATE)
    # time.sleep(0.5)
    # # #======测试触摸======
    i = 0
    while i < 0x7ffffffe:
        mk.touch_down(0,i,i)
        i += 20000000
        time.sleep(1/RATE)
    mk.touch_up(0)
    ##======测试DS5手柄======
    # mk.ds5_reset()
    # time.sleep(0.5)
    # for _ in range(10000):
    #     mk.dsreport.buttons.x = 1
    #     mk.send_ds5()
    #     time.sleep(1/RATE)
    #     mk.dsreport.buttons.x = 0
    #     mk.send_ds5()
    #     time.sleep(1/RATE)
    # mk.ds5_reset()

    # mk.hid_interface_reset()
