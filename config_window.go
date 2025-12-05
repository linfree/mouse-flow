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
		TailLength float64
		TailWidth  float64
		IsRainbow  bool
		Red        int
		Green      int
		Blue       int
	}

	vm := &ConfigViewModel{
		TailLength: float64(cfg.TailLength),
		TailWidth:  cfg.TailWidth,
		IsRainbow:  cfg.IsRainbow,
		Red:        int(cfg.TailColor[0]),
		Green:      int(cfg.TailColor[1]),
		Blue:       int(cfg.TailColor[2]),
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
		cfg.TailColor[0] = uint8(vm.Red)
		cfg.TailColor[1] = uint8(vm.Green)
		cfg.TailColor[2] = uint8(vm.Blue)

		if onUpdate != nil {
			onUpdate()
		}
	}

	if _, err := (MainWindow{
		AssignTo: &mainWindow,
		Title:    "Mouse Flow Configuration",
		Size:     Size{Width: 300, Height: 300},
		Layout:   VBox{},
		DataBinder: DataBinder{
			AssignTo:       &db,
			Name:           "vm",
			DataSource:     vm,
			ErrorPresenter: ToolTipErrorPresenter{},
		},
		Children: []Widget{
			GroupBox{
				Title:  "Appearance",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Length:"},
					NumberEdit{
						Value:          Bind("TailLength"),
						OnValueChanged: update,
						Decimals:       0,
					},

					Label{Text: "Width:"},
					NumberEdit{
						Value:          Bind("TailWidth"),
						OnValueChanged: update,
						Decimals:       1,
					},
				},
			},

			GroupBox{
				Title:  "Color",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					CheckBox{
						Text:             "Rainbow Mode",
						Checked:          Bind("IsRainbow"),
						OnCheckedChanged: update,
						ColumnSpan:       2,
					},

					Label{Text: "Red:"},
					Slider{
						Value:          Bind("Red"),
						MinValue:       0,
						MaxValue:       255,
						OnValueChanged: update,
						Enabled:        Bind("!vm.IsRainbow"),
					},

					Label{Text: "Green:"},
					Slider{
						Value:          Bind("Green"),
						MinValue:       0,
						MaxValue:       255,
						OnValueChanged: update,
						Enabled:        Bind("!vm.IsRainbow"),
					},

					Label{Text: "Blue:"},
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
						Text: "Save & Close",
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
