package main

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Version: V3.5.1
// Browser 定义浏览器信息
type Browser struct {
	Name      string
	Package   string
	Component string
}

// 预定义浏览器列表 (优先级排序)
var browserList = []Browser{
	// ==========================================
	//  国际主流 (最稳健)
	// ==========================================
	{Name: "Google Chrome", Package: "com.android.chrome", Component: "com.android.chrome/com.google.android.apps.chrome.Main"},
	{Name: "Microsoft Edge", Package: "com.microsoft.emmx", Component: "com.microsoft.emmx/com.microsoft.ruby.Main"},
	{Name: "Firefox", Package: "org.mozilla.firefox", Component: "org.mozilla.firefox/org.mozilla.firefox.App"},
	{Name: "Via Browser", Package: "mark.via", Component: "mark.via/mark.via.Shell"},
	
	// ==========================================
	//  游戏手机 & 冷门/小众品牌 (你的新需求)
	// ==========================================
	// 红魔 / 努比亚 (Nubia/Red Magic) - 通常叫 cn.nubia.browser
	{Name: "Nubia/RedMagic Browser", Package: "cn.nubia.browser", Component: ""},
	
	// 小米 / 红米 (MIUI/HyperOS)
	{Name: "Mi Browser (Global)", Package: "com.mi.global.browser", Component: "com.mi.global.browser/com.mi.global.browser.MainActivity"},
	
	// 联想 / ZUI (Lenovo/Legion/YOGA) - ZUI系统自带
	{Name: "Lenovo/ZUI Browser", Package: "com.zui.browser", Component: ""},
	{Name: "Lenovo Browser (Global)", Package: "com.lenovo.browser", Component: ""},

	// 摩托罗拉 (Motorola) - Moto通常接近原生，主要用Chrome，但部分国行版(MyUI)可能有定制
	{Name: "Motorola Browser", Package: "com.motorola.browser", Component: ""}, 

	// 华硕 ROG / ZenFone
	{Name: "Asus Browser", Package: "com.asus.browser", Component: ""},

	// 黑鲨 (Black Shark) - 基于MIUI，通常是通用包名或小米包名，但也可能独立
	{Name: "BlackShark Browser", Package: "com.blackshark.browser", Component: ""},

	// 索尼 (Sony Xperia) - 旧款有独立浏览器，新款主要用Chrome
	{Name: "Sony Browser", Package: "com.sony.nfx.app.sbrowser", Component: ""},

	// 锤子/坚果 (Smartisan) - 极少数存量用户
	{Name: "Smartisan Browser", Package: "com.smartisanos.browser", Component: ""},

	// HTC - 极少数存量用户
	{Name: "HTC Browser", Package: "com.htc.sense.browser", Component: ""},

	// ==========================================
	//  国内主流厂商 (Huawei, Xiaomi, OV)
	// ==========================================
	// 华为 / 荣耀
	{Name: "Huawei Browser", Package: "com.huawei.browser", Component: "com.huawei.browser/com.huawei.browser.Main"},
	{Name: "Honor Browser", Package: "com.hihonor.browser", Component: ""}, // 荣耀独立后
	
	{Name: "Mi Browser (CN)", Package: "com.android.browser", Component: "com.android.browser/com.android.browser.BrowserActivity"},
	
	// Vivo / iQOO
	{Name: "Vivo Browser", Package: "com.vivo.browser", Component: "com.vivo.browser/com.vivo.browser.BrowserActivity"},
	
	// Oppo / OnePlus / Realme
	{Name: "HeyTap Browser (Oppo/1+)", Package: "com.heytap.browser", Component: "com.heytap.browser/com.heytap.browser.BrowserActivity"},
	{Name: "ColorOS Browser (Old)", Package: "com.coloros.browser", Component: ""},
	
	// 魅族 (Meizu Flyme)
	{Name: "Meizu Browser", Package: "com.meizu.mbrowser", Component: "com.meizu.mbrowser/com.meizu.mbrowser.BrowserActivity"},

	// ==========================================
	//  国内互联网巨头APP (Quark, UC, QQ)
	// ==========================================
	{Name: "Quark (夸克)", Package: "com.quark.browser", Component: "com.quark.browser/com.uc.browser.ActivityUpdate"},
	{Name: "UC Browser", Package: "com.UCMobile", Component: "com.UCMobile/com.uc.browser.InnerUCMobile"},
	{Name: "QQ Browser", Package: "com.tencent.mtt", Component: "com.tencent.mtt/com.tencent.mtt.MainActivity"},
	{Name: "360 Browser", Package: "com.qihoo.browser", Component: "com.qihoo.browser/com.qihoo.browser.BrowserActivity"},
	{Name: "Sogou Browser", Package: "sogou.mobile.explorer", Component: "sogou.mobile.explorer/sogou.mobile.explorer.BrowserActivity"},
	{Name: "Baidu App", Package: "com.baidu.searchbox", Component: "com.baidu.searchbox/com.baidu.searchbox.MainActivity"},

	// ==========================================
	//  极客/隐私/轻量级
	// ==========================================
	{Name: "X Browser", Package: "com.mmbox.xbrowser", Component: "com.mmbox.xbrowser/com.mmbox.xbrowser.MainActivity"},
	{Name: "Brave", Package: "com.brave.browser", Component: "com.brave.browser/com.google.android.apps.chrome.Main"},
	{Name: "Kiwi", Package: "com.kiwibrowser.browser", Component: "com.kiwibrowser.browser/com.google.android.apps.chrome.Main"},
	{Name: "Yandex", Package: "com.yandex.browser", Component: "com.yandex.browser/com.yandex.browser.YandexBrowserActivity"},
	{Name: "Opera", Package: "com.opera.browser", Component: "com.opera.browser/com.opera.Opera"},
	{Name: "Soul Browser", Package: "com.mycompany.app.soulbrowser", Component: "com.mycompany.app.soulbrowser/com.mycompany.app.soulbrowser.MainActivity"},
	
	// ==========================================
	//  强力下载/极客工具 (1DM+, ADM)
	// ==========================================
	{Name: "1DM+ (Paid)", Package: "idm.internet.download.manager.plus", Component: "idm.internet.download.manager.plus/idm.internet.download.manager.MainActivity"},
	{Name: "1DM (Free)", Package: "idm.internet.download.manager", Component: "idm.internet.download.manager/idm.internet.download.manager.MainActivity"},
	{Name: "ADM (Advanced Download Manager)", Package: "com.dv.adm", Component: "com.dv.adm/com.dv.adm.AEditor"},
	
	// ==========================================
	// 终极保底
	// ==========================================
	// 很多改版ROM（包括旧版黑鲨、旧版努比亚）底层其实还是这个包名
	{Name: "Android Default", Package: "com.android.browser", Component: ""}, 
}

// GetAvailableURLs 获取所有可用的访问地址 (用于 server_ctrl 打印)
func GetAvailableURLs(port int) []string {
	var urls []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return []string{fmt.Sprintf("http://127.0.0.1:%d", port)}
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				urls = append(urls, fmt.Sprintf("http://%s:%d", ipv4.String(), port))
			}
		}
	}
	// 127.0.0.1 总是可用的
	urls = append(urls, fmt.Sprintf("http://127.0.0.1:%d", port))
	return urls
}

// checkServer 内部检查目标地址是否可达 (TCP握手)
func checkServer(url string) bool {
	address := strings.TrimPrefix(url, "http://")
	conn, err := net.DialTimeout("tcp", address, 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// tryLaunchAndroid 尝试执行 Android am start 命令
func tryLaunchAndroid(cmd *exec.Cmd) bool {
	outputBytes, err := cmd.CombinedOutput()
	// 如果由于系统没有 am 命令(如在常规 Linux 上)导致执行错误，直接判定失败
	if err != nil {
		return false
	}
	output := string(outputBytes)
	// 检查 am start 的输出，如果包含错误则失败
	if strings.Contains(output, "Error") || strings.Contains(output, "Exception") || strings.Contains(output, "unable to resolve") {
		return false
	}
	return true
}

// launchAndroidBrowser 执行原有的 Android (Termux) 浏览器唤起策略
func launchAndroidBrowser(targetURL string) bool {
	// 1. 第一轮：精确打击 (使用 Component)
	for _, b := range browserList {
		if b.Component == "" {
			continue
		}

		cmd := exec.Command("am", "start",
			"-n", b.Component,
			"-a", "android.intent.action.VIEW",
			"-d", targetURL,
			"--activity-clear-top",
			"--user", "0",
		)

		if tryLaunchAndroid(cmd) {
			return true // 成功唤起
		}
	}

	// 2. 第二轮：范围打击 (使用 Package)
	for _, b := range browserList {
		if b.Package == "" {
			continue
		}

		cmd := exec.Command("am", "start",
			"-p", b.Package,
			"-a", "android.intent.action.VIEW",
			"-d", targetURL,
			"--activity-clear-top",
			"--user", "0",
		)

		if tryLaunchAndroid(cmd) {
			return true // 成功唤起
		}
	}
	
	return false
}

// launchDesktopBrowser 尝试调用 Windows / macOS / 桌面 Linux 的系统默认浏览器
func launchDesktopBrowser(targetURL string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows: cmd /c start http://...
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	case "darwin":
		// macOS: open http://...
		cmd = exec.Command("open", targetURL)
	case "linux":
		// 桌面 Linux: xdg-open http://... (Termux 环境这里其实不会触发，因为前面会先走安卓逻辑拦截)
		cmd = exec.Command("xdg-open", targetURL)
	default:
		// 未知系统尝试桌面 Linux 方案兜底
		cmd = exec.Command("xdg-open", targetURL)
	}

	// 使用 Start() 而不是 Run()，启动后立即返回控制权，绝不阻塞当前程序
	_ = cmd.Start()
}

// LaunchBrowser 智能唤起浏览器 (全平台兼容机制)
// port: 实际监听的端口
func LaunchBrowser(port int) {
	// 1. 确定目标 URL (优选策略)
	targetURL := ""

	candidates := GetAvailableURLs(port)
	sortedURLs := make([]string, 0, len(candidates))
	for _, u := range candidates {
		if strings.Contains(u, "192.168.") {
			sortedURLs = append([]string{u}, sortedURLs...)
		} else {
			sortedURLs = append(sortedURLs, u)
		}
	}

	for _, url := range sortedURLs {
		if checkServer(url) {
			targetURL = url
			break
		}
	}

	if targetURL == "" {
		targetURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	}

	// 2. 跨平台执行指令分发
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		// 原生 PC 系统，直接调用桌面级 API
		launchDesktopBrowser(targetURL)
	} else {
		// "linux" (包含 Termux 安卓环境 和 真正的 Linux 桌面环境)
		// 优先使用 Android 引擎 (am start)
		if !launchAndroidBrowser(targetURL) {
			// 如果 am start 全部失败或者找不到命令，判断为纯正 Linux 桌面，降级使用 xdg-open
			launchDesktopBrowser(targetURL)
		}
	}
}