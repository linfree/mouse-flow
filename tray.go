package main

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procAppendMenuW = user32.NewProc("AppendMenuW")
)

func MAKEINTRESOURCE(id uintptr) *uint16 {
	return (*uint16)(unsafe.Pointer(id))
}

func AppendMenu(hMenu win.HMENU, uFlags uint32, uIDNewItem uintptr, lpNewItem *uint16) bool {
	ret, _, _ := procAppendMenuW.Call(
		uintptr(hMenu),
		uintptr(uFlags),
		uIDNewItem,
		uintptr(unsafe.Pointer(lpNewItem)),
	)
	return ret != 0
}

const (
	WM_TRAY = win.WM_USER + 1
	ID_TRAY = 1

	// 菜单 ID
	IDM_CONFIG = 1001
	IDM_EXIT   = 1002
)

// 全局变量用于通信
var (
	trayQuitChan       chan struct{}
	trayOpenConfigChan chan struct{}
)

func RunTray(quitChan chan struct{}, openConfigChan chan struct{}) {
	// 必须锁定 OS 线程，因为 Windows 消息循环和窗口是线程绑定的
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// 确保函数退出时通知主程序退出
	defer close(quitChan)

	trayQuitChan = quitChan
	trayOpenConfigChan = openConfigChan

	hInstance := win.GetModuleHandle(nil)
	className := syscall.StringToUTF16Ptr("MouseFlowTrayClass")

	// 注册窗口类
	wndClass := win.WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(win.WNDCLASSEX{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}

	if win.RegisterClassEx(&wndClass) == 0 {
		// 注册失败
		return
	}

	// 创建隐藏窗口 (Message-Only Window)
	// HWND_MESSAGE = -3 (0xFFFFFFFD) 但 win 包可能没定义，用 0 parent 也是隐藏的如果大小是0
	hwnd := win.CreateWindowEx(
		0,
		className,
		syscall.StringToUTF16Ptr("MouseFlowTray"),
		0,
		0, 0, 0, 0,
		0, // Parent
		0,
		hInstance,
		nil,
	)

	if hwnd == 0 {
		return
	}

	// 添加托盘图标
	var nid win.NOTIFYICONDATA
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = ID_TRAY
	nid.UFlags = win.NIF_ICON | win.NIF_MESSAGE | win.NIF_TIP
	nid.UCallbackMessage = WM_TRAY

	// 加载系统图标 (IDI_APPLICATION)
	nid.HIcon = win.LoadIcon(0, MAKEINTRESOURCE(win.IDI_APPLICATION))

	// 设置提示文本
	tip := syscall.StringToUTF16("Mouse Flow - 鼠标痕迹工具")
	copy(nid.SzTip[:], tip)

	win.Shell_NotifyIcon(win.NIM_ADD, &nid)

	// 消息循环
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}

	// 清理
	win.Shell_NotifyIcon(win.NIM_DELETE, &nid)
}

// 窗口过程
func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAY:
		// 处理托盘图标事件
		if lParam == win.WM_RBUTTONUP {
			// 获取鼠标位置
			var p win.POINT
			win.GetCursorPos(&p)

			// 创建弹出菜单
			hMenu := win.CreatePopupMenu()
			AppendMenu(hMenu, win.MF_STRING, IDM_CONFIG, syscall.StringToUTF16Ptr("配置 (Config)"))
			AppendMenu(hMenu, win.MF_STRING, IDM_EXIT, syscall.StringToUTF16Ptr("退出 (Exit)"))

			// 必须设置前台窗口，否则菜单点击后不会消失
			win.SetForegroundWindow(hwnd)

			// 显示菜单
			win.TrackPopupMenu(hMenu, win.TPM_BOTTOMALIGN|win.TPM_LEFTALIGN, p.X, p.Y, 0, hwnd, nil)
			win.DestroyMenu(hMenu)
		}
		return 0

	case win.WM_COMMAND:
		// 处理菜单点击
		id := win.LOWORD(uint32(wParam))
		switch id {
		case IDM_CONFIG:
			// 通知主线程打开配置
			select {
			case trayOpenConfigChan <- struct{}{}:
			default:
			}
		case IDM_EXIT:
			// 通知退出
			win.PostQuitMessage(0)
			// close(trayQuitChan) // 由 defer 处理
		}
		return 0

	case win.WM_DESTROY:
		win.PostQuitMessage(0)
		return 0
	}

	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}
