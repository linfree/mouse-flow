package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// TracePoint 轨迹上的单个点
// 优化：使用值类型而非指针，减少 GC
type TracePoint struct {
	X, Y float64
	Life float64 // 生命值 1.0 -> 0.0
}

// TraceManager 管理轨迹生成和渲染
type TraceManager struct {
	points     []TracePoint // 优化：值类型切片
	config     *Config
	whiteImage *ebiten.Image

	// 缓存切片，避免每帧分配
	vertices []ebiten.Vertex
	indices  []uint16

	// 状态追踪
	lastX, lastY float64
}

// NewTraceManager 创建新的轨迹管理器
func NewTraceManager(cfg *Config) *TraceManager {
	img := ebiten.NewImage(1, 1)
	img.Fill(color.White)

	// 预分配容量，减少扩容
	return &TraceManager{
		points:     make([]TracePoint, 0, 200),
		config:     cfg,
		whiteImage: img,
		vertices:   make([]ebiten.Vertex, 0, 1000),
		indices:    make([]uint16, 0, 1000),
	}
}

// Update 更新轨迹点
// 返回 true 表示有活动轨迹，false 表示空闲
func (tm *TraceManager) Update(mx, my int) bool {
	x, y := float64(mx), float64(my)

	moved := false
	// 检查是否移动
	if math.Abs(x-tm.lastX) > 0.1 || math.Abs(y-tm.lastY) > 0.1 {
		tm.lastX = x
		tm.lastY = y
		moved = true
	}

	// 添加新点逻辑
	if len(tm.points) == 0 {
		if moved {
			tm.points = append(tm.points, TracePoint{X: x, Y: y, Life: 1.0})
		}
	} else {
		last := tm.points[len(tm.points)-1]
		dist := math.Hypot(x-last.X, y-last.Y)

		// 只有移动了一定距离才添加新点
		if dist > 2.0 {
			tm.points = append(tm.points, TracePoint{X: x, Y: y, Life: 1.0})
		}
	}

	// 原地更新并过滤死亡点
	// 双指针法
	writeIdx := 0
	decay := 1.0 - tm.config.DecaySpeed

	// 预计算限制
	maxPoints := tm.config.TailLength
	startIdx := 0
	if len(tm.points) > maxPoints {
		startIdx = len(tm.points) - maxPoints
	}

	for i := startIdx; i < len(tm.points); i++ {
		// 直接修改切片中的元素
		tm.points[i].Life -= decay
		if tm.points[i].Life > 0 {
			// 如果需要移动元素 (即前面有被删除的元素)
			if writeIdx != i {
				tm.points[writeIdx] = tm.points[i]
			}
			writeIdx++
		}
	}

	// 裁剪切片
	tm.points = tm.points[:writeIdx]

	return len(tm.points) > 0
}

// Draw 绘制轨迹
func (tm *TraceManager) Draw(screen *ebiten.Image) {
	// 使用颜色键透明策略：清屏为纯黑，配合 LWA_COLORKEY 将黑色作为透明
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if len(tm.points) < 2 {
		return
	}

	// 复用切片
	tm.vertices = tm.vertices[:0]
	tm.indices = tm.indices[:0]

	// 预计算颜色分量，避免循环中重复计算
	r := float32(tm.config.TailColor[0]) / 255
	g := float32(tm.config.TailColor[1]) / 255
	b := float32(tm.config.TailColor[2]) / 255
	a := float32(tm.config.TailColor[3]) / 255

	width := tm.config.TailWidth

	for i := 0; i < len(tm.points)-1; i++ {
		// 使用指针访问以避免复制大结构体（虽然这里结构体很小）
		p1 := &tm.points[i]
		p2 := &tm.points[i+1]

		// 计算方向向量
		dx := p2.X - p1.X
		dy := p2.Y - p1.Y
		// 使用快速近似平方根？不需要，Hypot 够快且准确
		l := math.Hypot(dx, dy)
		if l == 0 {
			continue
		}

		// 归一化并旋转90度得到法向量
		nx := -dy / l
		ny := dx / l

		// 计算宽度
		w1 := width * p1.Life
		w2 := width * p2.Life

		// 计算颜色 alpha
		// 优化：只乘一次 alpha
		c1A := a * float32(p1.Life)
		c2A := a * float32(p2.Life)

		// P1 Left
		v1 := ebiten.Vertex{
			DstX:   float32(p1.X + nx*w1),
			DstY:   float32(p1.Y + ny*w1),
			ColorR: r * c1A, ColorG: g * c1A, ColorB: b * c1A, ColorA: c1A,
		}
		// P1 Right
		v2 := ebiten.Vertex{
			DstX:   float32(p1.X - nx*w1),
			DstY:   float32(p1.Y - ny*w1),
			ColorR: r * c1A, ColorG: g * c1A, ColorB: b * c1A, ColorA: c1A,
		}
		// P2 Left
		v3 := ebiten.Vertex{
			DstX:   float32(p2.X + nx*w2),
			DstY:   float32(p2.Y + ny*w2),
			ColorR: r * c2A, ColorG: g * c2A, ColorB: b * c2A, ColorA: c2A,
		}
		// P2 Right
		v4 := ebiten.Vertex{
			DstX:   float32(p2.X - nx*w2),
			DstY:   float32(p2.Y - ny*w2),
			ColorR: r * c2A, ColorG: g * c2A, ColorB: b * c2A, ColorA: c2A,
		}

		baseIndex := uint16(len(tm.vertices))
		tm.vertices = append(tm.vertices, v1, v2, v3, v4)
		tm.indices = append(tm.indices, baseIndex, baseIndex+1, baseIndex+2, baseIndex+1, baseIndex+3, baseIndex+2)
	}

	if len(tm.vertices) > 0 {
		screen.DrawTriangles(tm.vertices, tm.indices, tm.whiteImage, &ebiten.DrawTrianglesOptions{
			Blend:     ebiten.BlendSourceOver,
			AntiAlias: true, // 抗锯齿会增加一些开销，但效果好
		})
	}
}
