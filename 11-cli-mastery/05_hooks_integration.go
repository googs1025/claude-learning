//go:build ignore

// 第十一章示例 5：Hooks 集成
// 运行方式: go run 05_hooks_integration.go
//
// 本示例演示如何：
// 1. 用 Go 结构体生成 hooks JSON 配置
// 2. 通过 --settings 临时加载 hooks（不修改全局配置）
// 3. 实现 PreToolUse 安全验证和 PostToolUse 自动格式化
//
// 适用场景：CI/CD 中动态注入安全策略，无需修改项目配置文件

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ========== Hooks 配置结构体 ==========

// HooksSettings 表示 claude settings.json 中的 hooks 配置
type HooksSettings struct {
	Hooks HooksConfig `json:"hooks"`
}

// HooksConfig 包含各类 hook 事件
type HooksConfig struct {
	PreToolUse  []HookRule `json:"PreToolUse,omitempty"`
	PostToolUse []HookRule `json:"PostToolUse,omitempty"`
}

// HookRule 表示单条 hook 规则
type HookRule struct {
	Matcher string `json:"matcher"` // 工具名匹配模式（正则或具体名称）
	Hooks   []Hook `json:"hooks"`   // 要执行的 hook 列表
}

// Hook 表示具体要执行的 hook 命令
type Hook struct {
	Type    string `json:"type"`    // "command"
	Command string `json:"command"` // shell 命令
}

func main() {
	// ========== 步骤 1：构建 hooks 配置 ==========
	fmt.Println("=== 步骤 1：构建 hooks 配置 ===")

	settings := HooksSettings{
		Hooks: HooksConfig{
			PreToolUse: []HookRule{
				{
					// 在任何 Bash 命令执行前，检查是否包含危险操作
					Matcher: "Bash",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: `echo "安全检查: 即将执行 Bash 命令" >&2`,
						},
					},
				},
				{
					// 在 Write 工具使用前记录日志
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: `echo "审计日志: 即将写入文件" >&2`,
						},
					},
				},
			},
			PostToolUse: []HookRule{
				{
					// 在 Write 工具使用后，自动检查文件格式
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: `echo "后置检查: 文件已写入，可在此处运行 linter" >&2`,
						},
					},
				},
			},
		},
	}

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Fatalf("序列化 JSON 失败: %v", err)
	}

	fmt.Println("生成的 hooks 配置:")
	fmt.Println(string(jsonData))

	// ========== 步骤 2：写入临时文件 ==========
	fmt.Println("\n=== 步骤 2：写入临时文件 ===")

	tmpDir, err := os.MkdirTemp("", "claude-hooks-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	settingsFile := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsFile, jsonData, 0644); err != nil {
		log.Fatalf("写入配置文件失败: %v", err)
	}

	fmt.Printf("配置文件: %s\n", settingsFile)

	// ========== 步骤 3：使用 --settings 加载 hooks ==========
	fmt.Println("\n=== 步骤 3：通过 --settings 加载 hooks ===")
	fmt.Println("claude 会在执行工具前后触发对应的 hook 命令")
	fmt.Println()

	cmd := exec.Command("claude",
		"-p", "请列出当前目录下的 .go 文件",
		"--output-format", "json",
		"--model", "sonnet",
		"--settings", settingsFile, // 加载自定义 hooks 配置
	)

	output, err := cmd.Output()
	if err != nil {
		// 即使有 stderr 输出（hooks 的 echo），Output() 仍然能正常工作
		log.Fatalf("调用失败: %v", err)
	}

	var resp struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析失败: %v", err)
	}

	result := []rune(resp.Result)
	if len(result) > 200 {
		fmt.Printf("回复: %s...\n", string(result[:200]))
	} else {
		fmt.Printf("回复: %s\n", resp.Result)
	}

	// ========== 实用场景总结 ==========
	fmt.Println("\n=== Hooks 实用场景 ===")
	fmt.Println("1. PreToolUse + Bash: 拦截危险命令（rm -rf, git push --force）")
	fmt.Println("2. PostToolUse + Write: 写文件后自动运行 gofmt / eslint")
	fmt.Println("3. PostToolUse + Bash: 命令执行后记录审计日志")
	fmt.Println("4. PreToolUse + Edit: 修改前自动备份文件")
	fmt.Println()
	fmt.Println("关键点：")
	fmt.Println("  --settings 只在本次调用生效，不影响全局配置")
	fmt.Println("  hook 的 stderr 输出不会混入 --output-format json 的结果")
	fmt.Println("  hook 返回非零退出码会阻止工具执行（PreToolUse）")
}
