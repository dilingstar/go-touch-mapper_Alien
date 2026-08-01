package main

import (
	"embed"
	"encoding/binary" // 必须保留
	"encoding/json"
	"image"
	"image/jpeg"      // 必须保留
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Version: V3.5.0
//go:embed go-touch-mapper-gh-pages/build
var staticFS embed.FS


func screenHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 使用不带 -p 的命令，获取原始数据（速度极快）
	// 注意：由于你是通过 ADB 运行此程序，screencap 会直接拥有截屏权限
	cmd := exec.Command("screencap")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "Pipe error", http.StatusInternalServerError)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "Start error", http.StatusInternalServerError)
		return
	}
	// 确保函数结束时杀掉进程，防止僵尸进程占满手机内存
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// 2. 解析 12 字节的头部信息
	// 格式: [宽 4字节][高 4字节][格式 4字节]
	var header [12]byte
	if _, err := io.ReadFull(stdout, header[:]); err != nil {
		http.Error(w, "Read header failed", http.StatusInternalServerError)
		return
	}

	width := int(binary.LittleEndian.Uint32(header[0:4]))
	height := int(binary.LittleEndian.Uint32(header[4:8]))
	format := binary.LittleEndian.Uint32(header[8:12]) // 1=RGBA, 5=BGRA

	// 3. 计算数据大小并读取像素 (宽 * 高 * 4通道)
	dataSize := width * height * 4
	// 简单保护，防止读到奇怪的巨大数值导致 Termux 崩溃
	if dataSize <= 0 || dataSize > 50*1024*1024 {
		http.Error(w, "Invalid screen size", http.StatusInternalServerError)
		return
	}

	pixelData := make([]byte, dataSize)
	if _, err := io.ReadFull(stdout, pixelData); err != nil {
		http.Error(w, "Read pixels failed", http.StatusInternalServerError)
		return
	}

	// 4. 颜色修正 (部分手机是 BGRA 格式，如果你的画面颜色红蓝反了，把下面这段注释取消掉)
	/*
	if format == 5 {
		for i := 0; i < len(pixelData); i += 4 {
			// 交换 B 和 R
			pixelData[i], pixelData[i+2] = pixelData[i+2], pixelData[i]
		}
	}
	*/
	
	// 消除未使用变量报错
	_ = format

	// 5. 构建图像对象
	img := &image.RGBA{
		Pix:    pixelData,
		Stride: width * 4,
		Rect:   image.Rect(0, 0, width, height),
	}

	// 6. 发送响应
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "no-cache") // 禁止浏览器缓存

	// 7. 编码为 JPEG
	// Quality: 50 是为了在 WiFi 环境下保证极低延迟，你可以尝试改回 70 或 80
	if err := jpeg.Encode(w, img, &jpeg.Options{Quality: 70}); err != nil {
		logger.Errorf("Encode error: %v", err)
	}
}

// [V3.3.0] 随机端口 + 自动唤起浏览器
func serve(mapperFilePath string, reloadConfigureFunc func(mapperFilePath string), pm *PluginManager) {
	var configMutex sync.RWMutex
	webFS, err := fs.Sub(staticFS, "go-touch-mapper-gh-pages/build")
	if err != nil {
		logger.Errorf("无法加载静态文件: %v", err)
		return
	}
	http.Handle("/", http.FileServer(http.FS(webFS)))
	http.HandleFunc("/screen.png", screenHandler)

	http.HandleFunc("/configure/get", func(w http.ResponseWriter, r *http.Request) {
		configMutex.RLock()
		defer configMutex.RUnlock()

		content, err := os.ReadFile(mapperFilePath)
		if err != nil {
			http.Error(w, "无法读取配置文件", http.StatusInternalServerError)
			logger.Errorf("读取配置文件失败: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})

	http.HandleFunc("/configure/set", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "读取请求体失败", http.StatusBadRequest)
			return
		}

		// 验证是否为有效JSON
		if !json.Valid(body) {
			http.Error(w, "无效的JSON格式", http.StatusBadRequest)
			return
		}

		configMutex.Lock()
		defer configMutex.Unlock()

		// [V3.0.1] 解析 JSON 以分离 config 和 plugin 数据
		var json_data map[string]interface{}
		if err := json.Unmarshal(body, &json_data); err != nil {
			// 如果解析失败，可能是旧格式或者纯 config
			logger.Warnf("JSON解析警告: %v", err)
		}

		// 尝试提取 config 部分，如果不存在则假设整个 body 都是 config (兼容旧版)
		var config_bytes []byte
		if val, ok := json_data["config"]; ok {
			config_bytes, _ = json.Marshal(val)
		} else {
			config_bytes = body
		}

		// 备份原配置文件
		backupPath := mapperFilePath + ".bak"
		if err := os.Rename(mapperFilePath, backupPath); err != nil {
			// 备份失败仅记录，不阻断
			logger.Warnf("配置文件备份失败: %v", err)
		}

		// 写入新配置
		if err := os.WriteFile(mapperFilePath, config_bytes, 0644); err != nil {
			// 恢复备份
			os.Rename(backupPath, mapperFilePath)
			http.Error(w, "写入配置文件失败", http.StatusInternalServerError)
			logger.Errorf("写入配置文件失败: %v", err)
			return
		}

		// 删除备份
		os.Remove(backupPath)

		// 重新加载配置
		reloadConfigureFunc(mapperFilePath)

		// [V3.0.1 修复] 安全地更新插件配置
		if pm != nil {
			if plugin_val, ok := json_data["plugin"]; ok {
				if plugin_map, ok := plugin_val.(map[string]interface{}); ok {
					pm.update_user_config(plugin_map)
				}
			}
		}

		w.Write([]byte("配置更新成功"))
		logger.Info("配置文件已更新并重新加载")
	})

	// [V3.0.1] 移植原版插件接口 (带空指针检查)
	http.HandleFunc("/plugin/configure/getTemplate", func(w http.ResponseWriter, r *http.Request) {
		configMutex.RLock()
		defer configMutex.RUnlock()
		var content []byte
		if pm != nil {
			content = []byte(pm.config_template)
		} else {
			content = []byte("{}")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})

	http.HandleFunc("/plugin/configure/getConfig", func(w http.ResponseWriter, r *http.Request) {
		configMutex.RLock()
		defer configMutex.RUnlock()
		var content []byte
		if pm != nil {
			content, _ = json.Marshal(pm.user_config)
		} else {
			content = []byte("{}")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})

	// [V3.5.0] 新增：结束程序接口，用于接收前端发来的指令
	http.HandleFunc("/api/exit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		configMutex.RLock()
		content, err := os.ReadFile(mapperFilePath)
		configMutex.RUnlock()

		isEnable := false
		if err == nil {
			var jsonData map[string]interface{}
			if json.Unmarshal(content, &jsonData) == nil {
				if endExit, ok := jsonData["END_EXIT"].(map[string]interface{}); ok {
					if enable, ok := endExit["EXIT_ENABLE"].(bool); ok {
						isEnable = enable
					}
				}
			}
		}

		if isEnable {
			// 先返回成功状态，再延时自杀，确保前端能正常收到网络回应不报错
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("已结束运行"))
			go func() {
				time.Sleep(500 * time.Millisecond)
				TriggerSafeExit()
			}()
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte("未启用"))
		}
	})

	// [V3.3.0] 随机端口监听
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Fatal(err)
	}
	
	// 获取实际分配的端口
	realPort := listener.Addr().(*net.TCPAddr).Port
	
	// [V3.3.0] 启动浏览器唤起 (异步，不阻塞服务启动)
	// 但为了用户体验，我们稍微等待一下唤起动作，再打印日志
	go func() {
		// 尝试唤起浏览器 (静默)
		LaunchBrowser(realPort)
		
		// 唤起尝试结束后，打印访问地址
		time.Sleep(500 * time.Millisecond) // 稍微给浏览器一点启动时间
		
		urls := GetAvailableURLs(realPort)
		logger.Info("可从以下网址访问控制后台:")
		for _, url := range urls {
			logger.Info(url)
		}
	}()

	// 启动 HTTP 服务
	if err := http.Serve(listener, nil); err != nil {
		logger.Fatal(err)
	}
}