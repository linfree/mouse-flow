# Release v1.1.1 - Smooth Trace Rendering | 丝滑轨迹渲染优化

## 🇬🇧 English

### 🚀 Improvements
*   **Smooth Trace Rendering**: Implemented round joints and caps for the mouse trace. This eliminates jagged edges when the mouse moves quickly or turns sharply, resulting in a much smoother look.
*   **Better Blending**: Switched to a "Max" blending mode for the trace rendering. This prevents the "dark spots" where trace segments overlap, ensuring a uniform color and transparency throughout the trail.

---

## 🇨🇳 中文

### 🚀 改进
*   **丝滑轨迹渲染**: 为鼠标轨迹实现了圆角连接和圆角端点。这消除了鼠标快速移动或急转弯时的锯齿边缘，使轨迹看起来更加平滑圆润。
*   **混合模式优化**: 将轨迹渲染的混合模式切换为 "Max" 模式。这解决了轨迹段重叠处颜色变深的问题，确保了整个拖尾的颜色和透明度均匀一致。

---

# Release v1.1.0 - Ripple Effect & Multi-language Support | 水波纹特效与多语言支持

## 🇬🇧 English

### ✨ New Features
*   **Mouse Click Ripple Effect**: Added a cool ripple animation when clicking the mouse. You can customize the growth speed, decay speed, and width of the ripple in the settings.
*   **Multi-language Support**: The configuration window now supports both English and Chinese. It automatically detects your system language, or you can manually select your preferred language.
*   **Enhanced Configuration**: Added new options to fine-tune the ripple effect.

### 🚀 Improvements
*   **Performance Optimization**: Implemented smart idle detection. The application significantly reduces resource usage (CPU/GPU) when the mouse is inactive.
*   **Better Window Handling**: Improved window Z-order maintenance to prevent the overlay from being covered by other "always-on-top" windows.
*   **High DPI Support**: Increased default ripple width for better visibility on high-resolution screens.

### 🐛 Bug Fixes
*   Fixed an issue where the screen might turn black on some systems or specific resolutions.
*   Fixed a bug where configuration values could be loaded incorrectly.
*   Fixed window occlusion issues where the trace was hidden behind other windows.

---

## 🇨🇳 中文

### ✨ 新特性
*   **鼠标点击水波纹**: 新增鼠标点击时的波纹扩散效果。您可以在设置中自定义波纹的扩散速度、消失速度和线条粗细。
*   **多语言支持**: 配置窗口现在支持中文和英文。程序会自动检测您的系统语言，您也可以手动选择偏好语言。
*   **配置增强**: 添加了用于微调水波纹效果的新选项。

### 🚀 改进
*   **性能优化**: 实现了智能休眠检测。当鼠标静止时，程序会大幅降低资源占用（CPU/GPU）。
*   **窗口层级优化**: 改进了窗口置顶逻辑，防止轨迹层被其他“总在最前”的窗口遮挡。
*   **高分屏支持**: 增加了默认波纹宽度，确保在高分辨率屏幕上清晰可见。

### 🐛 问题修复
*   修复了在某些系统或特定分辨率下可能导致屏幕变黑的问题。
*   修复了配置文件数值可能加载错误的问题。
*   修复了轨迹层可能被其他窗口遮挡的问题。
