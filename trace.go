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

// Ripple 点击波纹
type Ripple struct {
	X, Y   float64
	Radius float64
	Life   float64 // 1.0 -> 0.0
}

// TraceManager 管理轨迹生成和渲染
type TraceManager struct {
	points     []TracePoint // 优化：值类型切片
	ripples    []Ripple
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
		ripples:    make([]Ripple, 0, 20),
		config:     cfg,
		whiteImage: img,
		vertices:   make([]ebiten.Vertex, 0, 1000),
		indices:    make([]uint16, 0, 1000),
	}
}

// AddRipple 添加一个点击波纹
func (tm *TraceManager) AddRipple(x, y int) {
	if !tm.config.IsRipple {
		return
	}
	tm.ripples = append(tm.ripples, Ripple{
		X:      float64(x),
		Y:      float64(y),
		Radius: 2.0,
		Life:   1.0,
	})
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

	// 更新波纹
	activeRipples := 0
	for i := range tm.ripples {
		tm.ripples[i].Radius += tm.config.RippleGrowthSpeed // 扩散速度
		tm.ripples[i].Life -= tm.config.RippleDecaySpeed    // 消失速度
		if tm.ripples[i].Life > 0 {
			if activeRipples != i {
				tm.ripples[activeRipples] = tm.ripples[i]
			}
			activeRipples++
		}
	}
	tm.ripples = tm.ripples[:activeRipples]

	return len(tm.points) > 0 || len(tm.ripples) > 0
}

// Draw 绘制轨迹
func (tm *TraceManager) Draw(screen *ebiten.Image) {
	// 透明清屏，避免整屏黑底
	screen.Fill(color.RGBA{0, 0, 0, 0})

	if len(tm.points) < 2 && len(tm.ripples) == 0 {
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

	// 1. 绘制轨迹
	if len(tm.points) >= 2 {
		width := tm.config.TailWidth

		// 辅助函数：绘制实心圆 (用于平滑连接处)
		addCircle := func(x, y, radius float64, red, green, blue, alpha float32) {
			if radius < 0.5 {
				return
			}
			const circleSegments = 12
			centerIdx := uint16(len(tm.vertices))

			// Center vertex
			tm.vertices = append(tm.vertices, ebiten.Vertex{
				DstX:   float32(x),
				DstY:   float32(y),
				ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha,
			})

			for i := 0; i <= circleSegments; i++ {
				angle := float64(i) * 2 * math.Pi / circleSegments
				sin, cos := math.Sincos(angle)
				tm.vertices = append(tm.vertices, ebiten.Vertex{
					DstX:   float32(x + radius*cos),
					DstY:   float32(y + radius*sin),
					ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha,
				})
			}

			for i := 0; i < circleSegments; i++ {
				// center, current, next
				tm.indices = append(tm.indices, centerIdx, centerIdx+1+uint16(i), centerIdx+1+uint16(i+1))
			}
		}

		for i := 0; i < len(tm.points)-1; i++ {
			// 使用指针访问以避免复制大结构体（虽然这里结构体很小）
			p1 := &tm.points[i]
			p2 := &tm.points[i+1]

			// 计算方向向量
			dx := p2.X - p1.X
			dy := p2.Y - p1.Y
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

			// 在 p1 处绘制圆角连接
			addCircle(p1.X, p1.Y, w1, r*c1A, g*c1A, b*c1A, c1A)
		}

		// 在最后一个点绘制圆角端点
		lastP := &tm.points[len(tm.points)-1]
		lastW := width * lastP.Life
		lastAlpha := a * float32(lastP.Life)
		addCircle(lastP.X, lastP.Y, lastW, r*lastAlpha, g*lastAlpha, b*lastAlpha, lastAlpha)
	}

	// 2. 绘制波纹 (圆环)
	const segments = 20 // 降低分段数以优化性能
	thickness := tm.config.RippleWidth
	if thickness <= 0 {
		thickness = 2.0
	}

	for _, ripple := range tm.ripples {
		baseAlpha := float32(ripple.Life) * a
		if baseAlpha <= 0 {
			continue
		}

		rIn := ripple.Radius
		rOut := ripple.Radius + thickness

		centerIndex := uint16(len(tm.vertices))

		for i := 0; i <= segments; i++ {
			angle := float64(i) * 2 * math.Pi / segments
			sin, cos := math.Sincos(angle)

			// Inner vertex
			tm.vertices = append(tm.vertices, ebiten.Vertex{
				DstX:   float32(ripple.X + rIn*cos),
				DstY:   float32(ripple.Y + rIn*sin),
				ColorR: r * baseAlpha, ColorG: g * baseAlpha, ColorB: b * baseAlpha, ColorA: baseAlpha,
			})

			// Outer vertex
			tm.vertices = append(tm.vertices, ebiten.Vertex{
				DstX:   float32(ripple.X + rOut*cos),
				DstY:   float32(ripple.Y + rOut*sin),
				ColorR: r * baseAlpha, ColorG: g * baseAlpha, ColorB: b * baseAlpha, ColorA: baseAlpha,
			})
		}

		for i := 0; i < segments; i++ {
			idx := centerIndex + uint16(i*2)
			tm.indices = append(tm.indices, idx, idx+1, idx+2, idx+1, idx+3, idx+2)
		}
	}

	if len(tm.vertices) > 0 {
		// 使用 Max 混合模式解决重叠部分颜色变深的问题
		// 当半透明的圆角和线段重叠时，Max 模式会取最大透明度而不是叠加，从而保持颜色均匀
		blend := ebiten.Blend{
			BlendFactorSourceRGB:        ebiten.BlendFactorOne,
			BlendFactorDestinationRGB:   ebiten.BlendFactorOne,
			BlendOperationRGB:           ebiten.BlendOperationMax,
			BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
			BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
			BlendOperationAlpha:         ebiten.BlendOperationMax,
		}

		screen.DrawTriangles(tm.vertices, tm.indices, tm.whiteImage, &ebiten.DrawTrianglesOptions{
			Blend:     blend,
			AntiAlias: false, // 关闭抗锯齿以提高性能
		})
	}
}
