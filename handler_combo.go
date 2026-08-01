package main

import (
	"strings"
	"sync"
)

// Version: V3.5.0
// ComboHandler 组合键管理器
// 负责维护物理按键状态，并生成虚拟组合键事件
type ComboHandler struct {
	parent *TouchHandler // 指向主处理器，用于回传虚拟事件

	// 物理按键状态池: 记录当前哪些键被按下了
	// key: 物理键名 (如 "KEY_Z"), value: true(按下)
	pressed_keys map[string]bool

	// 虚拟组合键状态: 记录当前哪些组合键处于激活状态
	// key: 组合键名 (如 "KEY_Z+KEY_X"), value: true(激活)
	active_combos map[string]bool

	// 已注册的组合键列表 (从配置文件加载)
	// 存储格式为: "KEY_Z+KEY_X" -> ["KEY_Z", "KEY_X"]
	registered_combos map[string][]string

	lock sync.RWMutex
}

// InitComboHandler 初始化组合键管理器
func InitComboHandler(parent *TouchHandler) *ComboHandler {
	return &ComboHandler{
		parent:            parent,
		pressed_keys:      make(map[string]bool),
		active_combos:     make(map[string]bool),
		registered_combos: make(map[string][]string),
		lock:              sync.RWMutex{},
	}
}

// LoadCombos 从配置中加载所有包含 "+" 的组合键
// [V3.4.5] 新增：扫描 V_MOUSE_SETTINGS 里的等价键配置，解决手柄组合键点击悖论
func (self *ComboHandler) LoadCombos() {
	self.lock.Lock()
	defer self.lock.Unlock()

	// 清空旧配置
	self.registered_combos = make(map[string][]string)
	self.active_combos = make(map[string]bool)
	self.pressed_keys = make(map[string]bool)

	// 辅助注册函数
	register := func(key_name string) {
		if strings.Contains(key_name, "+") {
			parts := strings.Split(key_name, "+")
			valid_parts := make([]string, 0)
			for _, p := range parts {
				trim_p := strings.TrimSpace(p)
				if trim_p != "" {
					valid_parts = append(valid_parts, trim_p)
				}
			}
			if len(valid_parts) > 1 {
				self.registered_combos[key_name] = valid_parts
				logger.Infof("[Combo] 注册组合键: %s (成分: %v)", key_name, valid_parts)
			}
		}
	}

	// 1. 遍历 KEY_MAPS (按键映射)
	if key_maps, ok := self.parent.config.CheckGet("KEY_MAPS"); ok {
		keys_map, _ := key_maps.Map()
		for key_name := range keys_map {
			register(key_name)
		}
	}

	// 2. 遍历 MACROS (宏)
	if macros, err := self.parent.config.Get("MACROS").Array(); err == nil {
		for i := range macros {
			triggerKey := self.parent.config.Get("MACROS").GetIndex(i).Get("TRIGGER_KEY").MustString()
			if triggerKey != "" {
				register(triggerKey)
			}
			stopKey := self.parent.config.Get("MACROS").GetIndex(i).Get("STOP_KEY").MustString()
			if stopKey != "" {
				register(stopKey)
			}
		}
	}

	// 3. 遍历 MOUSE.SWITCH_KEYS (映射切换)
	if switchKeys, err := self.parent.config.Get("MOUSE").Get("SWITCH_KEYS").StringArray(); err == nil {
		for _, key := range switchKeys {
			register(key)
		}
	}

	// 4. 遍历 MOUSE.POINTER_SWITCH_KEYS (指针切换)
	if pointerKeys, err := self.parent.config.Get("MOUSE").Get("POINTER_SWITCH_KEYS").StringArray(); err == nil {
		for _, key := range pointerKeys {
			register(key)
		}
	}

	// 5. 遍历 GLOBAL_RECOIL (压枪相关)
	recoil := self.parent.config.Get("GLOBAL_RECOIL")
	if keys, err := recoil.Get("TRIGGER_KEYS").StringArray(); err == nil {
		for _, k := range keys {
			register(k)
		}
	}
	if keys, err := recoil.Get("SCOPE_KEYS").StringArray(); err == nil {
		for _, k := range keys {
			register(k)
		}
	}
	if keys, err := recoil.Get("RESET_SPEED_KEYS").StringArray(); err == nil {
		for _, k := range keys {
			register(k)
		}
	}

	// 6. [V3.4.5] 遍历 V_MOUSE_SETTINGS 等价点击键
	vmouse := self.parent.config.Get("V_MOUSE_SETTINGS")
	if keys, err := vmouse.Get("LEFT_CLICK_KEYS").StringArray(); err == nil {
		for _, k := range keys {
			register(k)
		}
	}
	if keys, err := vmouse.Get("RIGHT_CLICK_KEYS").StringArray(); err == nil {
		for _, k := range keys {
			register(k)
		}
	}
	// 7. [V3.5.0] 遍历 END_EXIT 退出快捷键
	if exitCfg, ok := self.parent.config.CheckGet("END_EXIT"); ok {
		if keys, err := exitCfg.Get("EXIT_KEYS").StringArray(); err == nil {
			for _, k := range keys {
				register(k)
			}
		}
	}
}

// HandlePhysicalInput 处理物理按键输入，并尝试触发组合键
func (self *ComboHandler) HandlePhysicalInput(key_name string, up_down int32) {
	if key_name == "" {
		return
	}

	self.lock.Lock()
	defer self.lock.Unlock()

	// 1. 更新物理按键池状态
	if up_down == DOWN {
		self.pressed_keys[key_name] = true
	} else {
		delete(self.pressed_keys, key_name)
	}

	// 2. 遍历所有注册的组合键，检查是否满足触发条件
	for combo_name, parts := range self.registered_combos {
		// 检查该组合键的所有成分是否都已按下
		all_pressed := true
		for _, part := range parts {
			if !self.pressed_keys[part] {
				all_pressed = false
				break
			}
		}

		// 3. 状态变迁逻辑
		is_active := self.active_combos[combo_name]

		if all_pressed && !is_active {
			// [激活] 所有键都按下了，且之前未激活 -> 发送 DOWN
			self.active_combos[combo_name] = true
			logger.Debugf("[Combo] 组合键激活: %s", combo_name)

			// 异步调用以防止阻塞物理按键处理
			// 使用 "VirtualCombo" 作为设备名以示区别
			go self.parent.handel_key_up_down(combo_name, DOWN, "VirtualCombo")

		} else if !all_pressed && is_active {
			// [释放] 任意键抬起了，且之前是激活的 -> 发送 UP
			delete(self.active_combos, combo_name)
			logger.Debugf("[Combo] 组合键释放: %s", combo_name)

			go self.parent.handel_key_up_down(combo_name, UP, "VirtualCombo")
		}
	}
}

// ResetState 强制重置状态 (用于切换映射模式时清理)
func (self *ComboHandler) ResetState() {
	self.lock.Lock()
	defer self.lock.Unlock()

	// 如果有正在激活的组合键，先发送抬起信号
	for combo_name := range self.active_combos {
		go self.parent.handel_key_up_down(combo_name, UP, "VirtualCombo")
	}

	self.active_combos = make(map[string]bool)
	self.pressed_keys = make(map[string]bool)
}