package main

import (
	"encoding/json"
	"image/color"
	"os"
)

// Config 存储应用程序配置
type Config struct {
	TailColor         [4]uint8 `json:"tail_color"`          // 轨迹颜色 RGBA
	TailLength        int      `json:"tail_length"`         // 轨迹最大点数
	TailWidth         float64  `json:"tail_width"`          // 轨迹头部宽度
	DecaySpeed        float64  `json:"decay_speed"`         // 衰减速度 (每帧减少的透明度/宽度比例)
	IsRainbow         bool     `json:"is_rainbow"`          // 是否开启彩虹模式
	IsRipple          bool     `json:"is_ripple"`           // 是否开启点击波纹
	RippleGrowthSpeed float64  `json:"ripple_growth_speed"` // 波纹扩散速度
	RippleDecaySpeed  float64  `json:"ripple_decay_speed"`  // 波纹消失速度
	RippleWidth       float64  `json:"ripple_width"`        // 波纹圆环宽度
	Language          string   `json:"language"`            // 语言: "auto", "en", "zh"
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TailColor:         [4]uint8{255, 0, 0, 255}, // 红色
		TailLength:        20,
		TailWidth:         8.0,
		DecaySpeed:        0.95,
		IsRainbow:         false,
		IsRipple:          true,
		RippleGrowthSpeed: 3.0,
		RippleDecaySpeed:  0.04,
		RippleWidth:       5.0,
		Language:          "auto",
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return DefaultConfig(), nil // 如果文件不存在，返回默认配置
	}
	defer f.Close()

	cfg := DefaultConfig()
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return DefaultConfig(), err
	}

	// 检查并设置默认值 (针对旧配置文件缺少新字段的情况)
	if cfg.RippleGrowthSpeed == 0 {
		cfg.RippleGrowthSpeed = 3.0
	}
	if cfg.RippleDecaySpeed == 0 {
		cfg.RippleDecaySpeed = 0.04
	}
	if cfg.RippleWidth == 0 {
		cfg.RippleWidth = 5.0
	}
	if cfg.Language == "" {
		cfg.Language = "auto"
	}

	// 强制确保水波纹开启（响应用户恢复默认配置的请求）
	if !cfg.IsRipple {
		cfg.IsRipple = true
	}

	return cfg, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(filename string, cfg *Config) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// GetColor 返回 color.RGBA 对象
func (c *Config) GetColor() color.RGBA {
	return color.RGBA{c.TailColor[0], c.TailColor[1], c.TailColor[2], c.TailColor[3]}
}
