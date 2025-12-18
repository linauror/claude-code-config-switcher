package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Config struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"`
	IsActive bool   `json:"is_active"`
}

type ConfigManager struct {
	Configs    []Config `json:"configs"`
	configPath string
}

func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".claude-code-config-switcher")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "configs.json")
	manager := &ConfigManager{
		Configs:    []Config{},
		configPath: configPath,
	}

	if err := manager.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return manager, nil
}

func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &cm.Configs)
}

func (cm *ConfigManager) Save() error {
	data, err := json.MarshalIndent(cm.Configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

func (cm *ConfigManager) AddConfig(name, baseURL, token string) error {
	config := Config{
		Name:     name,
		BaseURL:  baseURL,
		Token:    token,
		IsActive: false,
	}

	cm.Configs = append(cm.Configs, config)
	return cm.Save()
}

func (cm *ConfigManager) EditConfig(index int, name, baseURL, token string) error {
	if index < 0 || index >= len(cm.Configs) {
		return fmt.Errorf("invalid config index")
	}

	cm.Configs[index].Name = name
	cm.Configs[index].BaseURL = baseURL
	cm.Configs[index].Token = token

	if err := cm.Save(); err != nil {
		return err
	}

	// If this is the active config, reapply it
	if cm.Configs[index].IsActive {
		return ApplyConfig(&cm.Configs[index])
	}

	return nil
}

func (cm *ConfigManager) SwitchConfig(index int) error {
	if index < 0 || index >= len(cm.Configs) {
		return fmt.Errorf("invalid config index")
	}

	for i := range cm.Configs {
		cm.Configs[i].IsActive = false
	}

	cm.Configs[index].IsActive = true
	if err := cm.Save(); err != nil {
		return err
	}

	// Apply the configuration to the system
	return ApplyConfig(&cm.Configs[index])
}

func (cm *ConfigManager) DeleteConfig(index int) error {
	if index < 0 || index >= len(cm.Configs) {
		return fmt.Errorf("invalid config index")
	}

	cm.Configs = append(cm.Configs[:index], cm.Configs[index+1:]...)
	return cm.Save()
}

func (cm *ConfigManager) GetActiveConfig() *Config {
	for i := range cm.Configs {
		if cm.Configs[i].IsActive {
			return &cm.Configs[i]
		}
	}
	return nil
}

// ClaudeSettings represents the structure of ~/.claude/settings.json
type ClaudeSettings struct {
	Env map[string]interface{} `json:"env"`
}

// ApplyConfig applies the configuration to the system
// On Windows: sets environment variables using setx
// On Mac/Linux: writes to ~/.claude/settings.json
func ApplyConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if runtime.GOOS == "windows" {
		return applyConfigWindows(config)
	}
	return applyConfigUnix(config)
}

// applyConfigWindows sets environment variables on Windows using setx
func applyConfigWindows(config *Config) error {
	// Set ANTHROPIC_AUTH_TOKEN
	cmd := exec.Command("setx", "ANTHROPIC_AUTH_TOKEN", config.Token)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set ANTHROPIC_AUTH_TOKEN: %w", err)
	}

	// Set ANTHROPIC_BASE_URL
	cmd = exec.Command("setx", "ANTHROPIC_BASE_URL", config.BaseURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set ANTHROPIC_BASE_URL: %w", err)
	}

	return nil
}

// applyConfigUnix writes configuration to ~/.claude/settings.json on Mac/Linux
func applyConfigUnix(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Read existing settings or create new one
	settings := ClaudeSettings{
		Env: make(map[string]interface{}),
	}

	if data, err := os.ReadFile(settingsPath); err == nil {
		// File exists, try to parse it
		if err := json.Unmarshal(data, &settings); err != nil {
			// If parsing fails, start fresh
			settings.Env = make(map[string]interface{})
		}
	}

	// Ensure Env map exists
	if settings.Env == nil {
		settings.Env = make(map[string]interface{})
	}

	// Update the configuration
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = config.Token
	settings.Env["ANTHROPIC_BASE_URL"] = config.BaseURL

	// Write back to file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Claude Code 配置切换器")
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.SetFixedSize(true)

	manager, err := NewConfigManager()
	if err != nil {
		dialog.ShowError(err, myWindow)
		return
	}

	var configList *widget.List
	var activeLabel *widget.Label
	var processingDialog dialog.Dialog

	updateUI := func() {
		configList.Refresh()
		activeConfig := manager.GetActiveConfig()
		if activeConfig != nil {
			activeLabel.SetText(fmt.Sprintf("当前激活: %s", activeConfig.Name))
		} else {
			activeLabel.SetText("当前激活: 无")
		}
	}

	activeLabel = widget.NewLabel("当前激活: 无")
	activeConfig := manager.GetActiveConfig()
	if activeConfig != nil {
		activeLabel.SetText(fmt.Sprintf("当前激活: %s", activeConfig.Name))
	}

	configList = widget.NewList(
		func() int {
			return len(manager.Configs)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				container.NewHBox(
					widget.NewButton("切换", nil),
					widget.NewButton("编辑", nil),
					widget.NewButton("删除", nil),
				),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			config := manager.Configs[id]
			borderContainer := obj.(*fyne.Container)

			label := borderContainer.Objects[0].(*widget.Label)
			activeIndicator := ""
			if config.IsActive {
				activeIndicator = "✓ "
			}
			label.SetText(fmt.Sprintf("%s%s", activeIndicator, config.Name))

			buttonsBox := borderContainer.Objects[1].(*fyne.Container)
			switchBtn := buttonsBox.Objects[0].(*widget.Button)
			editBtn := buttonsBox.Objects[1].(*widget.Button)
			deleteBtn := buttonsBox.Objects[2].(*widget.Button)

			switchBtn.OnTapped = func() {
				// 显示处理中对话框
				progress := widget.NewProgressBarInfinite()
				processingContent := container.NewVBox(
					widget.NewLabel("正在切换配置，请稍候..."),
					progress,
				)
				processingDialog = dialog.NewCustomWithoutButtons("处理中", processingContent, myWindow)
				processingDialog.Show()

				// 异步处理切换操作
				go func() {
					err := manager.SwitchConfig(id)

					// 在主线程更新UI
					myWindow.Canvas().SetContent(myWindow.Canvas().Content())
					processingDialog.Hide()

					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}

					updateUI()

					var msg string
					if runtime.GOOS == "windows" {
						msg = fmt.Sprintf("已切换到配置: %s\n\n环境变量已更新:\n- ANTHROPIC_AUTH_TOKEN\n- ANTHROPIC_BASE_URL\n\n注意：需要重启应用程序才能生效", config.Name)
					} else {
						msg = fmt.Sprintf("已切换到配置: %s\n\n配置已写入 ~/.claude/settings.json", config.Name)
					}
					dialog.ShowInformation("成功", msg, myWindow)
				}()
			}

			editBtn.OnTapped = func() {
				// Create edit dialog
				editNameEntry := widget.NewEntry()
				editNameEntry.SetText(config.Name)

				editBaseURLEntry := widget.NewEntry()
				editBaseURLEntry.SetText(config.BaseURL)

				editTokenEntry := widget.NewPasswordEntry()
				editTokenEntry.SetText(config.Token)

				editForm := container.NewVBox(
					widget.NewLabel("配置名称"),
					editNameEntry,
					widget.NewLabel("BASE_URL"),
					editBaseURLEntry,
					widget.NewLabel("API Key"),
					editTokenEntry,
				)

				editDialog := dialog.NewCustomConfirm("编辑配置", "保存", "取消", editForm,
					func(save bool) {
						if save {
							name := editNameEntry.Text
							baseURL := editBaseURLEntry.Text
							token := editTokenEntry.Text

							if name == "" || baseURL == "" || token == "" {
								dialog.ShowError(fmt.Errorf("请填写所有字段"), myWindow)
								return
							}

							// 显示处理中对话框
							progress := widget.NewProgressBarInfinite()
							processingContent := container.NewVBox(
								widget.NewLabel("正在保存配置，请稍候..."),
								progress,
							)
							processingDialog = dialog.NewCustomWithoutButtons("处理中", processingContent, myWindow)
							processingDialog.Show()

							// 异步处理编辑操作
							go func() {
								err := manager.EditConfig(id, name, baseURL, token)

								// 在主线程更新UI
								myWindow.Canvas().SetContent(myWindow.Canvas().Content())
								processingDialog.Hide()

								if err != nil {
									dialog.ShowError(err, myWindow)
									return
								}

								updateUI()

								var msg string
								if config.IsActive {
									if runtime.GOOS == "windows" {
										msg = fmt.Sprintf("配置 '%s' 已更新并重新应用\n\n环境变量已更新:\n- ANTHROPIC_AUTH_TOKEN\n- ANTHROPIC_BASE_URL\n\n注意：需要重启应用程序才能生效", name)
									} else {
										msg = fmt.Sprintf("配置 '%s' 已更新并重新应用\n\n配置已写入 ~/.claude/settings.json", name)
									}
								} else {
									msg = fmt.Sprintf("配置 '%s' 已更新", name)
								}
								dialog.ShowInformation("成功", msg, myWindow)
							}()
						}
					}, myWindow)

				editDialog.Resize(fyne.NewSize(500, 300))
				editDialog.Show()
			}

			deleteBtn.OnTapped = func() {
				// 检查是否为激活状态
				if config.IsActive {
					dialog.ShowError(fmt.Errorf("无法删除正在激活的配置\n\n请先切换到其他配置后再删除"), myWindow)
					return
				}

				dialog.ShowConfirm("确认删除",
					fmt.Sprintf("确定要删除配置 '%s' 吗？", config.Name),
					func(confirmed bool) {
						if confirmed {
							if err := manager.DeleteConfig(id); err != nil {
								dialog.ShowError(err, myWindow)
								return
							}
							updateUI()
							dialog.ShowInformation("成功", "配置已删除", myWindow)
						}
					}, myWindow)
			}
		},
	)

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("配置名称（例如：生产环境）")

	baseURLEntry := widget.NewEntry()
	baseURLEntry.SetPlaceHolder("BASE_URL（例如：https://api.claude.ai）")

	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetPlaceHolder("API Key")

	addButton := widget.NewButton("添加新配置", func() {
		name := nameEntry.Text
		baseURL := baseURLEntry.Text
		token := tokenEntry.Text

		if name == "" || baseURL == "" || token == "" {
			dialog.ShowError(fmt.Errorf("请填写所有字段"), myWindow)
			return
		}

		if err := manager.AddConfig(name, baseURL, token); err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		nameEntry.SetText("")
		baseURLEntry.SetText("")
		tokenEntry.SetText("")

		updateUI()
		dialog.ShowInformation("成功", fmt.Sprintf("配置 '%s' 已添加\n\n请点击'切换'按钮激活此配置", name), myWindow)
	})

	addForm := container.NewVBox(
		widget.NewLabel("添加新配置"),
		nameEntry,
		baseURLEntry,
		tokenEntry,
		addButton,
	)

	content := container.NewBorder(
		container.NewVBox(activeLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), addForm),
		nil,
		nil,
		container.NewScroll(configList),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
