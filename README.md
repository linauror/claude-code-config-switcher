# Claude Code 配置切换器

一个基于 Fyne 框架开发的桌面小工具，用于管理和切换 Claude Code 的配置（BASE_URL 和 API Key）。

## 题外话

这个工具是本人为了在体验使用不同编程 API 进行 vibe coding 时而开发的，整体代码全部由 AI 进行完成。

## 界面截图

![界面截图](https://github.com/linauror/claude-code-config-switcher/blob/main/ScreenShot.png)

## 系统要求

- Windows 10/11, macOS, 或 Linux
- 对于 Windows 用户：需要重启相关应用程序以使环境变量生效

## 使用方法

### 编译及运行应用

```
# 克隆代码
git clone https://github.com/linauror/claude-code-config-switcher.git

# 进入目录
cd claude-code-config-switcher

# 编译程序
go build
```

双击运行 `claude-code-config-switcher.exe` 即可启动应用程序。

如需编译为带图标的版本，可以使用以下命令：

```
# 安装 fyne 工具
go install fyne.io/tools/cmd/fyne@latest

# 为不同平台打包
fyne package -os windows -icon icon.png
fyne package -os linux -icon icon.png
fyne package -os darwin -icon icon.png
```

### 添加新配置

1. 在界面底部找到"添加新配置"区域
2. 填写以下信息：
   - **配置名称**: 为配置起一个易于识别的名称（例如：生产环境、测试环境、智谱 AI）
   - **BASE_URL**: Claude Code 的 API 端点 URL
     - 官方 API: `https://api.claude.ai`
     - 智谱 AI: `https://open.bigmodel.cn/api/anthropic`
   - **API Key**: 你的 Claude Code API Key
3. 点击"添加新配置"按钮
4. 新配置将添加到列表中，但不会自动激活
5. 需要手动点击"切换"按钮来激活配置

### 编辑配置

1. 在配置列表中找到要编辑的配置
2. 点击配置旁边的"编辑"按钮
3. 在弹出的对话框中修改配置信息
4. 点击"保存"按钮
5. 如果编辑的是当前激活的配置，修改会自动应用到系统

### 切换配置

1. 在配置列表中找到要切换的配置
2. 点击配置旁边的"切换"按钮
3. 配置将自动应用到系统：
   - **Windows**: 环境变量已更新（需要重启应用）
   - **Mac/Linux**: ~/.claude/settings.json 已更新
4. 当前激活的配置会显示 ✓ 标记
5. 顶部会显示当前激活的配置名称

### 删除配置

1. 在配置列表中找到要删除的配置
2. 点击配置旁边的"删除"按钮
3. 在确认对话框中点击"是"确认删除

## 配置文件位置

### 应用配置

- **应用自身配置**: `%USERPROFILE%\.claude-code-config-switcher\configs.json` (Windows) 或 `~/.claude-code-config-switcher/configs.json` (Mac/Linux)

### Claude Code 配置

- **Windows**: 环境变量（通过 `setx` 命令设置）
  - `ANTHROPIC_AUTH_TOKEN`
  - `ANTHROPIC_BASE_URL`
- **Mac/Linux**: `~/.claude/settings.json`

## 配置示例

对于智谱 AI 的配置，可以参考以下示例：

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "your_zhipu_api_key",
    "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic"
  }
}
```

## Windows 用户注意事项

在 Windows 上，环境变量的更新需要重启应用程序才能生效：

1. 切换配置后
2. 关闭所有正在运行的 Claude Code 实例
3. 重新启动 Claude Code

## 编译说明

如果需要重新编译：

### Windows（隐藏控制台窗口）

```bash
go build -ldflags -H=windowsgui -o claude-code-config-switcher.exe .
```

### Mac/Linux（标准编译）

```bash
go build -o claude-code-config-switcher .
```

**注意**: `-H=windowsgui` 参数会隐藏 Windows 控制台窗口，提供纯净的 GUI 体验。

## 技术栈

- **语言**: Go
- **GUI 框架**: Fyne v2.7.1
- **配置存储**: JSON 格式
- **跨平台支持**: Windows, macOS, Linux
