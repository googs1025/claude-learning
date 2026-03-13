//go:build ignore

// 第十三章示例 4：上下文增强 Hooks
// 运行方式: go run 04_context_hooks.go
//
// 本示例演示：
// 1. SessionStart hook：会话启动时注入上下文（stdout → Claude 的初始上下文）
// 2. Stop hook：会话结束后发送通知、记录摘要
// 3. 这两个 hook 的特殊机制与其他 hook 的区别

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
	PreToolUse   []HookRule `json:"PreToolUse,omitempty"`
	PostToolUse  []HookRule `json:"PostToolUse,omitempty"`
	SessionStart []HookRule `json:"SessionStart,omitempty"`
	Stop         []HookRule `json:"Stop,omitempty"`
}

type Settings struct {
	Hooks HooksConfig `json:"hooks"`
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║     第 13 章：Hooks 自动化守门人 - 上下文增强         ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	hooksDir, err := filepath.Abs("hooks")
	if err != nil {
		log.Fatalf("获取 hooks 目录失败: %v", err)
	}

	// ========== SessionStart 的特殊机制 ==========
	fmt.Println("━━━ SessionStart Hook 的特殊机制 ━━━")
	fmt.Println()
	fmt.Println("SessionStart 是唯一一个 stdout 有特殊用途的 hook：")
	fmt.Println()
	fmt.Println("  hook 脚本的 stdout")
	fmt.Println("        ↓")
	fmt.Println("  注入为 Claude 的初始上下文（系统提示的一部分）")
	fmt.Println()
	fmt.Println("这意味着：")
	fmt.Println("  • 可以向 Claude 注入动态环境信息（git 状态、依赖版本等）")
	fmt.Println("  • Claude 能感知当前项目的实际状态")
	fmt.Println("  • 避免 Claude 因不了解环境而犯错")
	fmt.Println()

	// ========== Stop Hook 的特殊机制 ==========
	fmt.Println("━━━ Stop Hook 的特殊机制 ━━━")
	fmt.Println()
	fmt.Println("Stop hook 在以下情况触发：")
	fmt.Println("  • Claude 自然完成任务（发送最后一条消息）")
	fmt.Println("  • 用户输入 /exit 或 Ctrl+C")
	fmt.Println("  • 交互式会话结束")
	fmt.Println()
	fmt.Println("典型用途：")
	fmt.Println("  • macOS/Linux 桌面通知（长任务完成提醒）")
	fmt.Println("  • 写入会话摘要到日志")
	fmt.Println("  • 清理临时文件")
	fmt.Println("  • 统计本次会话费用")
	fmt.Println()

	// ========== 构建上下文增强 Hooks 配置 ==========
	fmt.Println("━━━ 生成上下文增强 Hooks 配置 ━━━")
	fmt.Println()

	settings := Settings{
		Hooks: HooksConfig{
			SessionStart: []HookRule{
				{
					// SessionStart 没有 Matcher 字段（不匹配工具），直接列 hooks
					// 注意：SessionStart 用空字符串 matcher 或省略
					Matcher: "",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "07_session_check.sh"),
						},
					},
				},
			},
			Stop: []HookRule{
				{
					Matcher: "",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "08_stop_notify.sh"),
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

	fmt.Println("生成的上下文 hooks 配置:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// ========== 说明 07_session_check.sh 的输出 ==========
	fmt.Println("━━━ 07_session_check.sh 会向 Claude 注入的信息 ━━━")
	fmt.Println()

	// 直接运行脚本展示其输出（模拟 SessionStart 时注入的内容）
	sessionScript := filepath.Join(hooksDir, "07_session_check.sh")
	if _, err := os.Stat(sessionScript); err == nil {
		fmt.Println("脚本输出（这些内容会成为 Claude 的初始上下文）:")
		fmt.Println("┌" + "─────────────────────────────────────────────────" + "┐")

		checkCmd := exec.Command("bash", sessionScript)
		checkCmd.Stdin = os.Stdin // 传入空 stdin（模拟 JSON）
		checkOutput, _ := checkCmd.Output()

		// 格式化输出
		lines := string(checkOutput)
		for _, line := range splitLines(lines) {
			fmt.Printf("│ %-49s│\n", line)
		}
		fmt.Println("└" + "─────────────────────────────────────────────────" + "┘")
		fmt.Println()
	}

	// ========== 演示 ==========
	fmt.Println("━━━ 演示：会话启动时注入上下文 ━━━")
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", "claude-context-hooks-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	settingsFile := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsFile, jsonData, 0644); err != nil {
		log.Fatalf("写入配置失败: %v", err)
	}

	// 让 Claude 描述它感知到的环境信息
	prompt := "简短描述你在本次会话开始时收到的环境信息（git状态、Go版本等），用中文，不超过3句话。"
	fmt.Printf("提示词: %s\n\n", prompt)

	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--model", "sonnet",
		"--settings", settingsFile,
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
			fmt.Printf("Claude 的感知: %s\n", resp.Result)
		}
	}

	// ========== 实际使用建议 ==========
	fmt.Println()
	fmt.Println("━━━ 实际使用建议 ━━━")
	fmt.Println()
	fmt.Println("在 ~/.claude/settings.json 中全局配置 SessionStart hook:")
	fmt.Println()

	globalExample := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]interface{}{
				{
					"hooks": []map[string]interface{}{
						{
							"type":    "command",
							"command": "~/.claude/hooks/session_check.sh",
						},
					},
				},
			},
			"Stop": []map[string]interface{}{
				{
					"hooks": []map[string]interface{}{
						{
							"type":    "command",
							"command": "~/.claude/hooks/stop_notify.sh",
						},
					},
				},
			},
		},
	}

	globalJSON, _ := json.MarshalIndent(globalExample, "", "  ")
	fmt.Println(string(globalJSON))
	fmt.Println()
	fmt.Println("保存到 ~/.claude/settings.json 后，所有 Claude 会话自动生效。")
}

// splitLines 将字符串按换行符分割
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 49 {
				line = line[:49]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if len(line) > 49 {
			line = line[:49]
		}
		lines = append(lines, line)
	}
	return lines
}
