//go:build ignore

// 第十三章示例 3：代码质量 Hooks
// 运行方式: go run 03_quality_hooks.go
//
// 本示例演示：
// 1. PostToolUse hooks 用于代码质量自动化
// 2. 敏感数据检测（API Key、密码泄露告警）
// 3. 按扩展名自动运行格式化工具（gofmt、prettier、black）
// 4. 源码变更后自动触发测试套件

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type HookRule struct {
	Matcher string `json:"matcher"`
	Hooks   []Hook `json:"hooks"`
}

type HooksConfig struct {
	PreToolUse  []HookRule `json:"PreToolUse,omitempty"`
	PostToolUse []HookRule `json:"PostToolUse,omitempty"`
}

type Settings struct {
	Hooks HooksConfig `json:"hooks"`
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║     第 13 章：Hooks 自动化守门人 - 代码质量           ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	hooksDir, err := filepath.Abs("hooks")
	if err != nil {
		log.Fatalf("获取 hooks 目录失败: %v", err)
	}

	// ========== 构建质量 Hooks 配置 ==========
	fmt.Println("━━━ 代码质量 Hooks 配置 ━━━")
	fmt.Println()

	settings := Settings{
		Hooks: HooksConfig{
			PostToolUse: []HookRule{
				{
					// Hook 1: 敏感数据检测
					// Write 之后扫描文件中是否有 API Key / 密码
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "04_sensitive_data.sh"),
						},
					},
				},
				{
					// Hook 2: 自动代码格式化
					// Write 之后按扩展名运行 gofmt / prettier / black
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "05_auto_format.sh"),
						},
					},
				},
				{
					// Hook 3: 自动运行测试
					// Write 之后检测是否有对应测试，有则自动运行
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "06_run_tests.sh"),
						},
					},
				},
			},
		},
	}

	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	fmt.Println("生成的质量 hooks 配置:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// ========== 说明各 Hook 的工作原理 ==========
	fmt.Println("━━━ 各 Hook 功能说明 ━━━")
	fmt.Println()
	fmt.Println("1. 04_sensitive_data.sh (PostToolUse + Write)")
	fmt.Println("   • 检测模式: sk-xxx（Anthropic/OpenAI Key）、AKIA（AWS）、ghp_（GitHub Token）")
	fmt.Println("   • 检测模式: password=、secret=、private_key=、数据库连接串等")
	fmt.Println("   • 发现后通过 stderr 向 Claude 发警告（不阻断，因为已写入）")
	fmt.Println("   • 提示用户检查 .gitignore 和撤销真实密钥")
	fmt.Println()
	fmt.Println("2. 05_auto_format.sh (PostToolUse + Write)")
	fmt.Println("   • .go  → gofmt -w")
	fmt.Println("   • .js/.ts/.jsx/.tsx/.json/.css → prettier --write")
	fmt.Println("   • .py  → black（或 autopep8）")
	fmt.Println("   • .sh  → shfmt -w")
	fmt.Println("   • 其他扩展名：静默跳过")
	fmt.Println()
	fmt.Println("3. 06_run_tests.sh (PostToolUse + Write)")
	fmt.Println("   • .go  → 同目录有 _test.go 时运行 go test ./...")
	fmt.Println("   • .py  → 同目录有 test_*.py 时运行 pytest -q")
	fmt.Println("   • .js/.ts → 找到 package.json 时运行 npm test")
	fmt.Println("   • 跳过测试文件本身，避免无限循环")
	fmt.Println()

	// ========== PostToolUse 与 PreToolUse 的区别 ==========
	fmt.Println("━━━ PostToolUse vs PreToolUse ━━━")
	fmt.Println()
	fmt.Println("┌──────────────────┬────────────────────┬────────────────────┐")
	fmt.Println("│ 特性             │ PreToolUse         │ PostToolUse        │")
	fmt.Println("├──────────────────┼────────────────────┼────────────────────┤")
	fmt.Println("│ 触发时机         │ 工具执行前          │ 工具执行后          │")
	fmt.Println("│ 可以阻断         │ ✅ 可以             │ ❌ 不能             │")
	fmt.Println("│ 看到工具输出     │ ❌ 没有             │ ✅ tool_response   │")
	fmt.Println("│ 典型用途         │ 安全检查、权限控制  │ 格式化、测试、日志  │")
	fmt.Println("└──────────────────┴────────────────────┴────────────────────┘")
	fmt.Println()

	// ========== 演示：让 Claude 写一个 Go 文件，触发自动格式化 ==========
	fmt.Println("━━━ 演示：写入 Go 文件后自动格式化 ━━━")
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", "claude-quality-hooks-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	settingsFile := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsFile, jsonData, 0644); err != nil {
		log.Fatalf("写入配置失败: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "example.go")
	prompt := fmt.Sprintf(`请将以下内容写入文件 %s：

package main

import "fmt"

func main() {
fmt.Println("hello")
}

直接写入文件，不要解释。`, targetFile)

	fmt.Printf("提示词: 写入一个简单的 Go 文件到 %s\n\n", targetFile)
	fmt.Println("（gofmt hook 会在写入后自动格式化缩进）")
	fmt.Println()

	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--model", "sonnet",
		"--settings", settingsFile,
		"--allowedTools", "Write",
	)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("提示：如未安装 claude CLI，可跳过演示\n错误: %v\n", err)
	} else {
		var resp struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(output, &resp); err == nil {
			fmt.Printf("Claude 操作: %s\n", resp.Result)
		}

		// 检查文件是否已格式化
		if content, err := os.ReadFile(targetFile); err == nil {
			fmt.Printf("\n写入后的文件内容:\n%s\n", string(content))
		}
	}

	// ========== 手动测试说明 ==========
	fmt.Println()
	fmt.Println("━━━ 手动测试 Hook 脚本 ━━━")
	fmt.Println()
	fmt.Println("# 测试敏感数据检测（创建含 API Key 的文件后执行）:")
	fmt.Printf("echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/test.txt\"}}' | bash %s\n\n",
		filepath.Join(hooksDir, "04_sensitive_data.sh"))
	fmt.Println("# 测试自动格式化（需要 gofmt 已安装）:")
	fmt.Printf("echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/test.go\"}}' | bash %s\n",
		filepath.Join(hooksDir, "05_auto_format.sh"))
}
