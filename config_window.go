package main

import (
	"fmt"
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// ShowConfigWindow 显示配置对话框
func ShowConfigWindow(cfg *Config, onUpdate func()) {
	var mainWindow *walk.MainWindow
	var db *walk.DataBinder

	// 临时结构体用于数据绑定
	type ConfigViewModel struct {
		TailLength        float64
		TailWidth         float64
		IsRainbow         bool
		IsRipple          bool
		RippleGrowthSpeed float64
		RippleDecaySpeed  float64
		RippleWidth       float64
		Red               int
		Green             int
		Blue              int
		Language          string
	}

	vm := &ConfigViewModel{
		TailLength:        float64(cfg.TailLength),
		TailWidth:         cfg.TailWidth,
		IsRainbow:         cfg.IsRainbow,
		IsRipple:          cfg.IsRipple,
		RippleGrowthSpeed: cfg.RippleGrowthSpeed,
		RippleDecaySpeed:  cfg.RippleDecaySpeed,
		RippleWidth:       cfg.RippleWidth,
		Red:               int(cfg.TailColor[0]),
		Green:             int(cfg.TailColor[1]),
		Blue:              int(cfg.TailColor[2]),
		Language:          cfg.Language,
	}

	// 语言选项
	type LangOption struct {
		Name  string
		Value string
	}
	langOptions := []*LangOption{
		{Name: T("LangAuto"), Value: "auto"},
		{Name: T("LangZh"), Value: "zh"},
		{Name: T("LangEn"), Value: "en"},
	}

	// 更新配置的回调
	update := func() {
		if err := db.Submit(); err != nil {
			log.Println(err)
			return
		}

		cfg.TailLength = int(vm.TailLength)
		cfg.TailWidth = vm.TailWidth
		cfg.IsRainbow = vm.IsRainbow
		cfg.IsRipple = vm.IsRipple
		cfg.RippleGrowthSpeed = vm.RippleGrowthSpeed
		cfg.RippleDecaySpeed = vm.RippleDecaySpeed
		cfg.RippleWidth = vm.RippleWidth
		cfg.TailColor[0] = uint8(vm.Red)
		cfg.TailColor[1] = uint8(vm.Green)
		cfg.TailColor[2] = uint8(vm.Blue)
		cfg.Language = vm.Language

		// 应用语言设置 (注意：当前窗口的文本不会立即刷新，但下次打开或托盘菜单会生效)
		SetLanguage(cfg.Language)

		if onUpdate != nil {
			onUpdate()
		}
	}

	if _, err := (MainWindow{
		AssignTo: &mainWindow,
		Title:    T("Title"),
		Size:     Size{Width: 320, Height: 450}, // 稍微增加高度
		Layout:   VBox{},
		DataBinder: DataBinder{
			AssignTo:       &db,
			Name:           "vm",
			DataSource:     vm,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		Children: []Widget{
			GroupBox{
				Title:  T("Language"),
				Layout: HBox{},
				Children: []Widget{
					Label{Text: T("Language")},
					ComboBox{
						Value:                 Bind("Language"),
						Model:                 langOptions,
						BindingMember:         "Value",
						DisplayMember:         "Name",
						OnCurrentIndexChanged: update, // 选择即生效
					},
				},
			},

			GroupBox{
				Title:  T("Appearance"),
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: T("Length")},
					NumberEdit{
						Value:          Bind("TailLength"),
						OnValueChanged: update,
						Decimals:       0,
					},

					Label{Text: T("Width")},
					NumberEdit{
						Value:          Bind("TailWidth"),
						OnValueChanged: update,
						Decimals:       1,
					},
				},
			},

			GroupBox{
				Title:  T("RippleSettings"),
				Layout: Grid{Columns: 2},
				Children: []Widget{
					CheckBox{
						Text:             T("ClickRipple"),
						Checked:          Bind("IsRipple"),
						OnCheckedChanged: update,
						ColumnSpan:       2,
					},
					Label{Text: T("RippleGrowth")},
					NumberEdit{
						Value:          Bind("RippleGrowthSpeed"),
						OnValueChanged: update,
						Decimals:       1,
						Enabled:        Bind("vm.IsRipple"),
					},
					Label{Text: T("RippleDecay")},
					NumberEdit{
						Value:          Bind("RippleDecaySpeed"),
						OnValueChanged: update,
						Decimals:       3,
						Enabled:        Bind("vm.IsRipple"),
					},
					Label{Text: T("RippleWidth")},
					NumberEdit{
						Value:          Bind("RippleWidth"),
						OnValueChanged: update,
						Decimals:       1,
						Enabled:        Bind("vm.IsRipple"),
					},
				},
			},

			GroupBox{
				Title:  T("ColorEffects"),
				Layout: Grid{Columns: 2},
				Children: []Widget{
					CheckBox{
						Text:             T("RainbowMode"),
						Checked:          Bind("IsRainbow"),
						OnCheckedChanged: update,
						ColumnSpan:       2,
					},

					Label{Text: T("Red")},
					Slider{
						Value:          Bind("Red"),
						MinValue:       0,
						MaxValue:       255,
						OnValueChanged: update,
						Enabled:        Bind("!vm.IsRainbow"),
					},

					Label{Text: T("Green")},
					Slider{
						Value:          Bind("Green"),
						MinValue:       0,
						MaxValue:       255,
						OnValueChanged: update,
						Enabled:        Bind("!vm.IsRainbow"),
					},

					Label{Text: T("Blue")},
					Slider{
						Value:          Bind("Blue"),
						MinValue:       0,
						MaxValue:       255,
						OnValueChanged: update,
						Enabled:        Bind("!vm.IsRainbow"),
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: T("SaveClose"),
						OnClicked: func() {
							update()
							SaveConfig("config.json", cfg)
							mainWindow.Close()
						},
					},
				},
			},
		},
	}).Run(); err != nil {
		fmt.Println(err)
	}
}
