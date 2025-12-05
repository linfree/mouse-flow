package main

import (
	"log"
	"runtime"
	"syscall"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lxn/win"
)

const (
	GWL_EXSTYLE      = -20
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_APPWINDOW  = 0x00040000
	SWP_NOSIZE       = 0x0001
	SWP_NOMOVE       = 0x0002
	SWP_NOZORDER     = 0x0004
	SWP_FRAMECHANGED = 0x0020
)

type Game struct {
	traceManager *TraceManager
	config       *Config
	quitChan     chan struct{}

	screenWidth  int
	screenHeight int

	// 窗口句柄
	hwnd win.HWND

	// 性能优化：空闲检测
	idleCounter int
}

func (g *Game) Update() error {
	// 检查退出信号
	select {
	case <-g.quitChan:
		return ebiten.Termination
	default:
	}

	// 获取鼠标位置
	var pt win.POINT
	win.GetCursorPos(&pt)

	// 计算相对于窗口的坐标
	// 如果找到了窗口句柄，就用真实的窗口位置计算
	// 否则回退到虚拟屏幕计算（虽然这可能是错的）
	mx, my := 0, 0

	if g.hwnd != 0 {
		var rect win.RECT
		win.GetWindowRect(g.hwnd, &rect)

		// 使用比例映射来解决 DPI 缩放导致的不一致问题
		// 窗口的物理像素大小
		windowWidth := int(rect.Right - rect.Left)
		windowHeight := int(rect.Bottom - rect.Top)

		// 避免除以 0
		if windowWidth > 0 && windowHeight > 0 {
			// 计算鼠标相对于窗口左上角的偏移量
			offsetX := int(pt.X) - int(rect.Left)
			offsetY := int(pt.Y) - int(rect.Top)

			// 计算归一化比例 (0.0 - 1.0)
			rx := float64(offsetX) / float64(windowWidth)
			ry := float64(offsetY) / float64(windowHeight)

			// 映射到 Ebiten 的 Layout 坐标系
			mx = int(rx * float64(g.screenWidth))
			my = int(ry * float64(g.screenHeight))
		} else {
			// 如果窗口大小异常，回退到简单差值
			mx = int(pt.X) - int(rect.Left)
			my = int(pt.Y) - int(rect.Top)
		}
	} else {
		// 尝试查找窗口句柄 (如果 main 中的协程还没找到)
		// 注意：频繁 FindWindow 可能有开销，但这里只有在找不到时才调用
		titlePtr := syscall.StringToUTF16Ptr("MouseFlowOverlay")
		hwnd := win.FindWindow(nil, titlePtr)
		if hwnd != 0 {
			g.hwnd = hwnd
			// 递归调用一次或直接使用上面的逻辑，为了简单，这里重复逻辑但简化
			var rect win.RECT
			win.GetWindowRect(g.hwnd, &rect)
			windowWidth := int(rect.Right - rect.Left)
			windowHeight := int(rect.Bottom - rect.Top)
			if windowWidth > 0 && windowHeight > 0 {
				offsetX := int(pt.X) - int(rect.Left)
				offsetY := int(pt.Y) - int(rect.Top)
				rx := float64(offsetX) / float64(windowWidth)
				ry := float64(offsetY) / float64(windowHeight)
				mx = int(rx * float64(g.screenWidth))
				my = int(ry * float64(g.screenHeight))
			} else {
				mx = int(pt.X) - int(rect.Left)
				my = int(pt.Y) - int(rect.Top)
			}
		} else {
			// 实在找不到，暂时使用虚拟屏幕原点
			x := win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
			y := win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)
			mx = int(pt.X) - int(x)
			my = int(pt.Y) - int(y)
		}
	}

	isActive := g.traceManager.Update(mx, my)

	// 智能休眠逻辑
	if isActive {
		g.idleCounter = 0
		ebiten.SetTPS(60) // 恢复高刷新率以保证流畅动画
	} else {
		g.idleCounter++
		if g.idleCounter > 60 { // 约 1 秒无活动
			ebiten.SetTPS(15) // 降低刷新率以节省 CPU/GPU
		}
	}

	// 如果彩虹模式
	if g.config.IsRainbow {
		g.updateRainbow()
	}

	return nil
}

func (g *Game) updateRainbow() {
	// 简单的颜色循环
	g.config.TailColor[0] = uint8((int(g.config.TailColor[0]) + 1) % 255)
	g.config.TailColor[1] = uint8((int(g.config.TailColor[1]) + 2) % 255)
	g.config.TailColor[2] = uint8((int(g.config.TailColor[2]) + 3) % 255)
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 绘制轨迹
	g.traceManager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func main() {
	// 加载配置
	cfg, err := LoadConfig("config.json")
	if err != nil {
		log.Println("Config not found, using default")
		cfg = DefaultConfig()
	}

	// 获取虚拟屏幕位置和尺寸
	vx := int(win.GetSystemMetrics(win.SM_XVIRTUALSCREEN))
	vy := int(win.GetSystemMetrics(win.SM_YVIRTUALSCREEN))
	vw := int(win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN))
	vh := int(win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN))

	// 通信通道
	quitChan := make(chan struct{})
	openConfigChan := make(chan struct{})

	// 启动托盘
	go RunTray(quitChan, openConfigChan)

	// 监听配置请求
	go func() {
		// GUI 线程需要锁定
		runtime.LockOSThread()
		for range openConfigChan {
			log.Println("Opening config window...")
			ShowConfigWindow(cfg, func() {
				// 配置更新时的回调
			})
		}
	}()

	// 初始化游戏
	game := &Game{
		traceManager: NewTraceManager(cfg),
		config:       cfg,
		quitChan:     quitChan,
		screenWidth:  vw,
		screenHeight: vh,
	}

	// Ebiten 设置
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowSize(vw, vh)
	ebiten.SetWindowPosition(vx, vy)

	// 设置背景透明
	ebiten.SetScreenTransparent(true)

	// 初始设置为鼠标穿透
	ebiten.SetWindowMousePassthrough(true)

	title := "MouseFlowOverlay"
	ebiten.SetWindowTitle(title)

	// 隐藏任务栏图标并强制全屏覆盖
	go func() {
		// 尝试多次，以防窗口创建延迟
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		titlePtr := syscall.StringToUTF16Ptr(title)
		applyCount := 0

		for range ticker.C {
			// FindWindow 可能找不到，如果标题还没设置好。
			// Ebiten 的默认类名不确定，所以用 nil
			hwnd := win.FindWindow(nil, titlePtr)
			if hwnd != 0 {
				// 获取当前扩展样式
				exStyle := win.GetWindowLong(hwnd, GWL_EXSTYLE)

				// 移除 APPWINDOW，添加 TOOLWINDOW
				newExStyle := (exStyle & ^WS_EX_APPWINDOW) | WS_EX_TOOLWINDOW

				if newExStyle != exStyle {
					win.SetWindowLong(hwnd, GWL_EXSTYLE, newExStyle)
					log.Println("Window style updated to hide from taskbar")
				}

				// 强制设置窗口位置和大小，覆盖整个虚拟屏幕
				// 即使 Ebiten/GLFW 试图限制它，我们也强制覆盖
				// SWP_NOACTIVATE = 0x0010
				win.SetWindowPos(hwnd, 0, int32(vx), int32(vy), int32(vw), int32(vh),
					win.SWP_NOZORDER|0x0010|SWP_FRAMECHANGED)

				applyCount++
				if applyCount > 5 {
					// 设置 HWND 给 Game
					game.hwnd = hwnd
					// 再次检查确认
					currentStyle := win.GetWindowLong(hwnd, GWL_EXSTYLE)
					if currentStyle&WS_EX_TOOLWINDOW != 0 {
						return
					}
				}
			}
		}
	}()

	// 解决 walk 库可能的初始化问题
	// 需要确保 InitCommonControls 被调用，不过 walk 包通常会在 init 中做。
	// 关键是 manifest 文件。

	if err := ebiten.RunGame(game); err != nil {
		if err != ebiten.Termination {
			log.Fatal(err)
		}
	}
}
