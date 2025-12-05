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

	// 获取虚拟屏幕的左上角坐标
	x := win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
	y := win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)

	// 计算相对于虚拟屏幕的坐标
	mx := int(pt.X) - int(x)
	my := int(pt.Y) - int(y)

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

	// 隐藏任务栏图标
	go func() {
		// 尝试多次，以防窗口创建延迟
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		titlePtr := syscall.StringToUTF16Ptr(title)

		for range ticker.C {
			// FindWindow 可能找不到，如果标题还没设置好。
			// Ebiten 的默认类名不确定，所以用 nil
			hwnd := win.FindWindow(nil, titlePtr)
			if hwnd != 0 {
				log.Printf("Found window HWND: %X", hwnd)

				// 获取当前扩展样式
				exStyle := win.GetWindowLong(hwnd, GWL_EXSTYLE)

				// 移除 APPWINDOW，添加 TOOLWINDOW
				newExStyle := (exStyle & ^WS_EX_APPWINDOW) | WS_EX_TOOLWINDOW

				if newExStyle != exStyle {
					win.SetWindowLong(hwnd, GWL_EXSTYLE, newExStyle)

					// 强制刷新窗口样式
					win.SetWindowPos(hwnd, 0, 0, 0, 0, 0,
						SWP_NOMOVE|SWP_NOSIZE|SWP_NOZORDER|SWP_FRAMECHANGED)

					log.Println("Window style updated to hide from taskbar")
				}

				// 找到并设置后可以退出循环，或者继续监控以防被重置？
				// 通常只需要设置一次。但为了保险起见，我们可以多检测几次。
				// 这里我们假设设置成功后就退出了。

				// 再次检查确认
				currentStyle := win.GetWindowLong(hwnd, GWL_EXSTYLE)
				if currentStyle&WS_EX_TOOLWINDOW != 0 {
					return
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
