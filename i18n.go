package main

import (
	"syscall"
)

var (
	kernel32dll                  = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultUILanguage = kernel32dll.NewProc("GetUserDefaultUILanguage")
)

// Language 语言代码
type Language int

const (
	LangEnglish Language = iota
	LangChinese
)

var currentLang Language

// i18n 字符串映射
var i18nStrings = map[Language]map[string]string{
	LangEnglish: {
		"Title":          "Mouse Flow Configuration",
		"Appearance":     "Appearance",
		"Length":         "Length:",
		"Width":          "Width:",
		"ColorEffects":   "Color & Effects",
		"RainbowMode":    "Rainbow Mode",
		"ClickRipple":    "Click Ripple Effect",
		"Red":            "Red:",
		"Green":          "Green:",
		"Blue":           "Blue:",
		"SaveClose":      "Save & Close",
		"TrayTip":        "Mouse Flow - Mouse Trace Tool",
		"MenuConfig":     "Configuration",
		"MenuExit":       "Exit",
		"Language":       "Language:",
		"LangAuto":       "Auto",
		"LangEn":         "English",
		"LangZh":         "Chinese",
		"RippleSettings": "Ripple Settings",
		"RippleGrowth":   "Growth Speed:",
		"RippleDecay":    "Decay Speed:",
		"RippleWidth":    "Ripple Width:",
	},
	LangChinese: {
		"Title":          "鼠标轨迹配置",
		"Appearance":     "外观设置",
		"Length":         "轨迹长度:",
		"Width":          "轨迹宽度:",
		"ColorEffects":   "颜色与特效",
		"RainbowMode":    "彩虹模式",
		"ClickRipple":    "点击波纹特效",
		"Red":            "红色 (R):",
		"Green":          "绿色 (G):",
		"Blue":           "蓝色 (B):",
		"SaveClose":      "保存并关闭",
		"TrayTip":        "Mouse Flow - 鼠标痕迹工具",
		"MenuConfig":     "配置",
		"MenuExit":       "退出",
		"Language":       "语言设置:",
		"LangAuto":       "自动 (跟随系统)",
		"LangEn":         "English",
		"LangZh":         "简体中文",
		"RippleSettings": "波纹设置",
		"RippleGrowth":   "扩散速度:",
		"RippleDecay":    "消失速度:",
		"RippleWidth":    "波纹宽度:",
	},
}

func init() {
	SetLanguage("auto")
}

// SetLanguage 设置语言
func SetLanguage(lang string) {
	switch lang {
	case "zh":
		currentLang = LangChinese
	case "en":
		currentLang = LangEnglish
	default: // "auto" or others
		// 检测系统语言
		langID, _, _ := procGetUserDefaultUILanguage.Call()
		// 0x0804 是简体中文 (2052)
		if langID == 0x0804 {
			currentLang = LangChinese
		} else {
			currentLang = LangEnglish
		}
	}
}

// T 获取翻译后的字符串
func T(key string) string {
	if strMap, ok := i18nStrings[currentLang]; ok {
		if val, ok := strMap[key]; ok {
			return val
		}
	}
	// Fallback to English if not found
	if strMap, ok := i18nStrings[LangEnglish]; ok {
		if val, ok := strMap[key]; ok {
			return val
		}
	}
	return key
}
