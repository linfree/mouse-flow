# Mouse Flow

[中文文档](README.md) | **English**

Mouse Flow is a lightweight mouse trace tool for Windows that adds a cool "Fruit Ninja" style trail effect to your cursor.

![Screenshot Placeholder](screenshots/preview.gif)

## ✨ Features

- **Cool Trail**: Smooth "Fruit Ninja" style mouse trail with customizable color and width.
- **Click Ripple**: Circular ripple effect on mouse click with adjustable growth and decay speeds.
- **Global Rendering**: Powered by the Ebiten game engine, it implements a borderless transparent overlay across the full screen (including multi-monitors) without affecting mouse click penetration.
- **Multi-language Support**: Configuration interface supports Chinese and English, with automatic system language detection.
- **Low Resource Usage**: Deeply optimized with smart sleep mode (reduces refresh rate when idle), significantly minimizing CPU/GPU usage.
- **Silent Operation**: Minimizes to the system tray upon startup, no taskbar icon distraction.
- **Visual Configuration**: Click the tray icon to open the configuration window and adjust parameters in real-time.
  - 🎨 **Color Settings**: Customize RGB color or enable "Rainbow Mode".
  - 🌊 **Ripple Settings**: Toggle ripple effect, adjust growth and decay speeds.
  - 📏 **Appearance Adjustment**: Freely adjust trace length and width.

## 🚀 Download & Install

### Method 1: Run Directly (Recommended)
1. Download the latest archive from the [Releases](https://github.com/linfree/mouse-flow/releases) page.
2. Extract all files to the same directory.
3. Run `mouse_flow.exe`.

### Method 2: Build from Source
If you are a developer, you can manually build the source code.

**Prerequisites**:
- Go 1.20+
- GCC Compiler (for CGO, TDM-GCC or MinGW-w64 is recommended on Windows)

**Build Steps**:
```bash
# 1. Clone repository
git clone https://github.com/linfree/mouse-flow.git
cd mouse-flow

# 2. Download dependencies
go mod tidy

# 3. Build (Use -H windowsgui to hide console window)
go build -ldflags="-H windowsgui" -o mouse_flow.exe
```

**Note**:
- The `mouse_flow.exe.manifest` file must be present in the running directory, otherwise the configuration window may not display correctly.

## 📖 Usage

1. **Start**: Double-click `mouse_flow.exe`. A mouse trail will appear on the screen, and a small icon will appear in the system tray.
2. **Config**:
   - Right-click the tray icon -> Select **Config**.
   - Or simply left-click the tray icon (if supported).
   - Adjust parameters in the popup window and click **Save & Close** to apply.
3. **Exit**: Right-click the tray icon -> Select **Exit**.

## ⚙️ Configuration

The program generates a `config.json` in the running directory. You can also modify it manually:

```json
{
  "tail_length": 20,      // Trace length
  "tail_width": 8.0,      // Trace width
  "tail_color": [255, 0, 0, 255], // RGBA color (0-255)
  "is_rainbow": false,    // Enable rainbow mode
  "decay_speed": 0.95,    // Decay factor
  "is_ripple": true,      // Enable click ripple
  "ripple_growth_speed": 3.0, // Ripple growth speed
  "ripple_decay_speed": 0.04, // Ripple decay speed
  "ripple_width": 5.0,    // Ripple line width
  "language": "auto"      // Language ("auto", "zh", "en")
}
```

## 🛠️ Tech Stack

- [Ebiten](https://ebiten.org/) - A dead simple 2D game library for Go.
- [Walk](https://github.com/lxn/walk) - A Windows GUI library for Go.
- [Win API](https://github.com/lxn/win) - Handling low-level window messages, tray icons, and multi-monitor support.

## 🤝 Contribution

Issues and Pull Requests are welcome!

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
