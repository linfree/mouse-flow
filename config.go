package main

import (
	"encoding/json"
	"image/color"
	"os"
)

// Config 存储应用程序配置
type Config struct {
	TailColor  [4]uint8 `json:"tail_color"`  // 轨迹颜色 RGBA
	TailLength int      `json:"tail_length"` // 轨迹最大点数
	TailWidth  float64  `json:"tail_width"`  // 轨迹头部宽度
	DecaySpeed float64  `json:"decay_speed"` // 衰减速度 (每帧减少的透明度/宽度比例)
	IsRainbow  bool     `json:"is_rainbow"`  // 是否开启彩虹模式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TailColor:  [4]uint8{255, 0, 0, 255}, // 红色
		TailLength: 20,
		TailWidth:  8.0,
		DecaySpeed: 0.95,
		IsRainbow:  false,
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return DefaultConfig(), nil // 如果文件不存在，返回默认配置
	}
	defer f.Close()

	cfg := &Config{}
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return DefaultConfig(), err
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
