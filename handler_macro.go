package main

import (
	"context"
	"sync"
	"time"

	"github.com/bitly/go-simplejson"
)

// Version: V3.4.2
// MacroHandler 宏管理器
type MacroHandler struct {
	parent *TouchHandler
	macros map[string]*MacroItem // key: TRIGGER_KEY

	// 运行时状态
	runningMacros sync.Map // key: trigger_key, value: context.CancelFunc (用于终止循环/长按宏)
	toggleStates  sync.Map // key: trigger_key, value: bool (Toggle模式的开关状态)
	
	// [V3.4.2] 按抬模式激活状态表
	// 用于记录某个按抬宏是否通过了 PRESS 阶段的状态检查
	// key: trigger_key, value: bool (true=已激活Press)
	activePRMacros sync.Map 

	lock sync.RWMutex
}

// MacroItem 宏配置结构
type MacroItem struct {
	ID          string
	Label       string
	TriggerKey  string
	StopKey     string // 额外的终止键 (用于开关循环模板)
	TriggerMode string // ALWAYS, MAP_ON, MAP_OFF
	Type        string // PRESS_RELEASE, HOLD_LOOP, TOGGLE_LOOP

	LoopInterval int // 循环间隔 (ms)

	PressEvents   []*MacroEvent
	ReleaseEvents []*MacroEvent

	// 执行模式: SEQUENCE(依次), TIMEOUT(超时), SIMULTANEOUS(同时)
	ExecModePress   string
	ExecModeRelease string

	TimeoutPress   int // 超时模式下的限时 (ms)
	TimeoutRelease int
}

// MacroEvent 宏事件结构
type MacroEvent struct {
	Type     string // CLICK, SWIPE, DELAY, MAP_ON, MAP_OFF
	Pos      []float64
	PosList  [][]float64 // 滑动路径
	Duration int         // 点击时长 / 滑动间隔 / 延迟时长
}

func InitMacroHandler(parent *TouchHandler) *MacroHandler {
	return &MacroHandler{
		parent:       parent,
		macros:       make(map[string]*MacroItem),
		runningMacros: sync.Map{},
		toggleStates: sync.Map{},
		activePRMacros: sync.Map{},
	}
}

// LoadMacros 从配置加载宏
func (mh *MacroHandler) LoadMacros() {
	mh.lock.Lock()
	defer mh.lock.Unlock()

	mh.macros = make(map[string]*MacroItem)
	// 清理旧的运行状态
	mh.runningMacros.Range(func(key, value interface{}) bool {
		cancel := value.(context.CancelFunc)
		cancel()
		return true
	})
	mh.runningMacros = sync.Map{}
	mh.toggleStates = sync.Map{}
	mh.activePRMacros = sync.Map{}

	macroArr := mh.parent.config.Get("MACROS").MustArray()
	for i := range macroArr {
		itemJson := mh.parent.config.Get("MACROS").GetIndex(i)
		item := &MacroItem{
			ID:              itemJson.Get("ID").MustString(),
			Label:           itemJson.Get("LABEL").MustString(),
			TriggerKey:      itemJson.Get("TRIGGER_KEY").MustString(),
			StopKey:         itemJson.Get("STOP_KEY").MustString(),
			TriggerMode:     itemJson.Get("TRIGGER_MODE").MustString("ALWAYS"),
			Type:            itemJson.Get("TYPE").MustString("PRESS_RELEASE"),
			LoopInterval:    itemJson.Get("LOOP_INTERVAL").MustInt(0),
			ExecModePress:   itemJson.Get("EXEC_MODE_PRESS").MustString("SEQUENCE"),
			ExecModeRelease: itemJson.Get("EXEC_MODE_RELEASE").MustString("SEQUENCE"),
			TimeoutPress:    itemJson.Get("TIMEOUT_PRESS").MustInt(50),
			TimeoutRelease:  itemJson.Get("TIMEOUT_RELEASE").MustInt(50),
		}

		// 解析按下事件
		pressArr := itemJson.Get("PRESS_EVENTS").MustArray()
		for j := range pressArr {
			evtJson := itemJson.Get("PRESS_EVENTS").GetIndex(j)
			evt := parseMacroEvent(evtJson)
			item.PressEvents = append(item.PressEvents, evt)
		}

		// 解析抬起事件
		relArr := itemJson.Get("RELEASE_EVENTS").MustArray()
		for j := range relArr {
			evtJson := itemJson.Get("RELEASE_EVENTS").GetIndex(j)
			evt := parseMacroEvent(evtJson)
			item.ReleaseEvents = append(item.ReleaseEvents, evt)
		}

		if item.TriggerKey != "" {
			mh.macros[item.TriggerKey] = item
		}
	}
	logger.Infof("[Macro] 已加载 %d 个宏", len(mh.macros))
}

func parseMacroEvent(json *simplejson.Json) *MacroEvent {
	evt := &MacroEvent{
		Type:     json.Get("TYPE").MustString(),
		Duration: json.Get("DURATION").MustInt(0),
	}
	if posArr, err := json.Get("POS").Array(); err == nil && len(posArr) >= 2 {
		evt.Pos = []float64{json.Get("POS").GetIndex(0).MustFloat64(), json.Get("POS").GetIndex(1).MustFloat64()}
	}
	if listArr, err := json.Get("POS_LIST").Array(); err == nil {
		for k := range listArr {
			ptJson := json.Get("POS_LIST").GetIndex(k)
			if ptArr, err := ptJson.Array(); err == nil && len(ptArr) >= 2 {
				evt.PosList = append(evt.PosList, []float64{
					ptJson.GetIndex(0).MustFloat64(),
					ptJson.GetIndex(1).MustFloat64(),
				})
			}
		}
	}
	return evt
}

// HandleInput 处理输入
func (mh *MacroHandler) HandleInput(keyName string, upDown int32) {
	mh.lock.RLock()
	macro, exists := mh.macros[keyName]
	
	// 检查是否有额外的终止键触发 (针对开关循环模式)
	if !exists && upDown == DOWN {
		mh.toggleStates.Range(func(k, v interface{}) bool {
			tKey := k.(string)
			m := mh.macros[tKey]
			if m != nil && m.Type == "TOGGLE_LOOP" && m.StopKey == keyName {
				// 触发终止键
				mh.stopMacro(tKey)
				mh.toggleStates.Store(tKey, false)
			}
			return true
		})
	}
	mh.lock.RUnlock()

	if !exists {
		return
	}

	// 状态门槛检查 (仅在 Start 时检查)
	// [V3.4.2] 逻辑修正：如果按下时未通过检查，则不能执行按下事件，且标记为未激活，
	// 这样抬起时也不会执行抬起事件。
	if upDown == DOWN {
		allowed := false
		if macro.TriggerMode == "ALWAYS" {
			allowed = true
		} else if macro.TriggerMode == "MAP_ON" && mh.parent.map_on {
			allowed = true
		} else if macro.TriggerMode == "MAP_OFF" && !mh.parent.map_on {
			allowed = true
		}

		// 如果不满足条件，直接忽略，不执行任何操作
		if !allowed {
			// 确保清理激活状态
			mh.activePRMacros.Delete(keyName)
			return
		}
		
		// 满足条件，标记激活
		mh.activePRMacros.Store(keyName, true)
	}

	switch macro.Type {
	case "PRESS_RELEASE":
		if upDown == DOWN {
			go mh.executeEventList(macro.PressEvents, macro.ExecModePress, macro.TimeoutPress, nil)
		} else {
			// [V3.4.2] 抬起时检查：必须是之前成功激活过 Press 的宏才执行 Release
			if isActive, ok := mh.activePRMacros.Load(keyName); ok && isActive.(bool) {
				go mh.executeEventList(macro.ReleaseEvents, macro.ExecModeRelease, macro.TimeoutRelease, nil)
				// 执行完后清理状态
				mh.activePRMacros.Delete(keyName)
			}
		}

	case "HOLD_LOOP":
		if upDown == DOWN {
			// 停止旧的 (如果有)
			mh.stopMacro(keyName)
			
			ctx, cancel := context.WithCancel(context.Background())
			mh.runningMacros.Store(keyName, cancel)
			
			go mh.runLoop(ctx, macro, macro.PressEvents, macro.ExecModePress, macro.TimeoutPress)
		} else {
			// 松手即停
			mh.stopMacro(keyName)
		}

	case "TOGGLE_LOOP":
		if upDown == DOWN {
			isRunning, _ := mh.toggleStates.Load(keyName)
			if isRunning != nil && isRunning.(bool) {
				// 停止
				mh.stopMacro(keyName)
				mh.toggleStates.Store(keyName, false)
			} else {
				// 启动
				ctx, cancel := context.WithCancel(context.Background())
				mh.runningMacros.Store(keyName, cancel)
				mh.toggleStates.Store(keyName, true)
				
				go mh.runLoop(ctx, macro, macro.PressEvents, macro.ExecModePress, macro.TimeoutPress)
			}
		}
	}
}

func (mh *MacroHandler) stopMacro(keyName string) {
	if cancel, ok := mh.runningMacros.Load(keyName); ok {
		cancel.(context.CancelFunc)()
		mh.runningMacros.Delete(keyName)
	}
}

func (mh *MacroHandler) runLoop(ctx context.Context, macro *MacroItem, events []*MacroEvent, mode string, timeout int) {
	defer func() {
		logger.Debugf("[Macro] Loop stopped: %s", macro.Label)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 执行一轮
			mh.executeEventList(events, mode, timeout, ctx)
			
			// 循环间隔
			if macro.LoopInterval > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(macro.LoopInterval) * time.Millisecond):
				}
			} else {
				// 避免无间隔死循环导致CPU占用过高
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

// executeEventList 执行事件列表
// ctx 可选，用于在循环模式下提前终止
func (mh *MacroHandler) executeEventList(events []*MacroEvent, mode string, timeout int, ctx context.Context) {
	if len(events) == 0 {
		return
	}

	// 辅助函数: 检查 Context 是否取消
	isCancelled := func() bool {
		if ctx == nil { return false }
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	if mode == "SIMULTANEOUS" {
		// 并发执行
		var wg sync.WaitGroup
		for _, evt := range events {
			wg.Add(1)
			go func(e *MacroEvent) {
				defer wg.Done()
				mh.doAction(e)
			}(evt)
		}
		wg.Wait()
		return
	}

	for _, evt := range events {
		if isCancelled() { return }

		if mode == "TIMEOUT" {
			// 超时模式
			done := make(chan bool)
			go func() {
				mh.doAction(evt)
				close(done)
			}()

			waitDuration := time.Duration(timeout) * time.Millisecond
			
			select {
			case <-done:
				// 动作在超时前完成了
			case <-time.After(waitDuration):
				// 超时了，继续下一个
			case <-func() <-chan struct{} { if ctx != nil { return ctx.Done() } else { return nil } }():
				return
			}

		} else {
			// SEQUENCE: 依次执行
			mh.doAction(evt)
		}
	}
}

func (mh *MacroHandler) doAction(evt *MacroEvent) {
	switch evt.Type {
	case "CLICK":
		if len(evt.Pos) < 2 { return }
		// 应用随机落点 (调用 parent 的方法获取屏幕坐标)
		rx, ry := mh.parent.apply_key_jitter(
			int32(evt.Pos[0] * float64(mh.parent.rel_screen_x)),
			int32(evt.Pos[1] * float64(mh.parent.rel_screen_y)),
		)
		tid := mh.parent.touch_require(rx, ry, true)
		time.Sleep(time.Duration(evt.Duration) * time.Millisecond)
		mh.parent.touch_release(tid)

	case "SWIPE":
		if len(evt.PosList) < 2 { return }
		// 起点
		sx, sy := mh.parent.apply_key_jitter(
			int32(evt.PosList[0][0] * float64(mh.parent.rel_screen_x)),
			int32(evt.PosList[0][1] * float64(mh.parent.rel_screen_y)),
		)
		tid := mh.parent.touch_require(sx, sy, true)
		
		interval := time.Duration(evt.Duration) * time.Millisecond
		if interval < 5*time.Millisecond { interval = 5 * time.Millisecond }

		time.Sleep(interval)
		
		// 移动路径
		for i := 1; i < len(evt.PosList); i++ {
			mx := int32(evt.PosList[i][0] * float64(mh.parent.rel_screen_x))
			my := int32(evt.PosList[i][1] * float64(mh.parent.rel_screen_y))
			mh.parent.touch_move(tid, mx, my, true)
			time.Sleep(interval)
		}
		mh.parent.touch_release(tid)

	case "DELAY":
		time.Sleep(time.Duration(evt.Duration) * time.Millisecond)

	case "MAP_ON":
		if !mh.parent.map_on {
			mh.parent.switch_map_mode(true)
		}

	case "MAP_OFF":
		if mh.parent.map_on {
			mh.parent.switch_map_mode(false)
		}
	}
}