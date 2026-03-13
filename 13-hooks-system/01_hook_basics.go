//go:build ignore

// 第十三章示例 1：Hooks 机制原理
// 运行方式: go run 01_hook_basics.go
//
// 本示例演示：
// 1. Claude Code Hooks 的工作原理（stdin JSON → 解析 → 决策 → 退出码）
// 2. 如何用 Go 生成一个"纯日志"hooks 配置并运行演示
// 3. Hook 脚本的核心代码模式

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║     第 13 章：Hooks 自动化守门人 - 机制原理           ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// ========== 第一节：Hooks 工作原理 ==========
	fmt.Println("━━━ 1. Hooks 工作原理 ━━━")
	fmt.Println()
	fmt.Println("Claude Code 执行工具时的 Hook 流程：")
	fmt.Println()
	fmt.Println("  [Claude 决定执行工具]")
	fmt.Println("         ↓")
	fmt.Println("  [PreToolUse Hook 触发]")
	fmt.Println("    → Claude Code 将 JSON 写入 hook 脚本的 stdin")
	fmt.Println("    → hook 脚本读取、分析、决策")
	fmt.Println("    → exit 0 : 允许执行")
	fmt.Println("    → exit 2 : 阻断执行（快速拒绝）")
	fmt.Println("    → JSON输出 + exit 0 : 携带原因的阻断/允许")
	fmt.Println("         ↓")
	fmt.Println("  [工具执行]")
	fmt.Println("         ↓")
	fmt.Println("  [PostToolUse Hook 触发]")
	fmt.Println("    → 同上，但不能阻断（工具已执行）")
	fmt.Println("    → 用于：格式化、测试、日志记录")
	fmt.Println()

	// ========== 第二节：stdin JSON 格式 ==========
	fmt.Println("━━━ 2. stdin JSON 格式（Hook 接收到的数据）━━━")
	fmt.Println()

	// 模拟 Claude Code 传给 hook 的 JSON
	type ToolInput struct {
		Command  string `json:"command,omitempty"`
		FilePath string `json:"file_path,omitempty"`
	}
	type HookInput struct {
		ToolName  string    `json:"tool_name"`
		ToolInput ToolInput `json:"tool_input"`
		SessionID string    `json:"session_id,omitempty"`
	}

	// PreToolUse 示例（Bash 工具）
	bashExample := HookInput{
		ToolName:  "Bash",
		ToolInput: ToolInput{Command: "rm -rf /tmp/test"},
		SessionID: "sess_abc123",
	}
	bashJSON, _ := json.MarshalIndent(bashExample, "  ", "  ")
	fmt.Println("PreToolUse (Bash) - hook 的 stdin:")
	fmt.Printf("  %s\n\n", bashJSON)

	// PreToolUse 示例（Write 工具）
	writeExample := HookInput{
		ToolName:  "Write",
		ToolInput: ToolInput{FilePath: "/project/.env"},
		SessionID: "sess_abc123",
	}
	writeJSON, _ := json.MarshalIndent(writeExample, "  ", "  ")
	fmt.Println("PreToolUse (Write) - hook 的 stdin:")
	fmt.Printf("  %s\n\n", writeJSON)

	// ========== 第三节：Hook 脚本核心模式 ==========
	fmt.Println("━━━ 3. Shell 脚本核心代码模式 ━━━")
	fmt.Println()

	corePattern := `#!/bin/bash
# 步骤 1: 从 stdin 读取 JSON（关键！不是环境变量）
INPUT=$(cat)

# 步骤 2: 用 python3 解析 JSON（bash 原生不支持 JSON）
TOOL_NAME=$(echo "$INPUT" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('tool_name',''))")
COMMAND=$(echo "$INPUT" | python3 -c \
    "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('command',''))")

# 步骤 3: 决策逻辑
if echo "$COMMAND" | grep -q "危险模式"; then
    # 方式A: JSON 输出（推荐）- 可携带原因
    echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse",
          "permissionDecision":"deny","permissionDecisionReason":"原因"}}'
    exit 0

    # 方式B: exit 2（快速阻断）- stderr 反馈给 Claude
    # echo "⚠️ 已阻断" >&2
    # exit 2
fi

# 步骤 4: 允许执行（静默通过）
exit 0`

	fmt.Println(corePattern)
	fmt.Println()

	// ========== 第四节：生成日志 Hook 并演示 ==========
	fmt.Println("━━━ 4. 生成纯日志 Hook 配置演示 ━━━")
	fmt.Println()

	// 构建一个只做日志记录的 hooks 配置
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

	logHook := `INPUT=$(cat); TOOL=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name','?'))" 2>/dev/null); echo "[$(date '+%H:%M:%S')] Hook触发: $TOOL" >&2`

	settings := Settings{
		Hooks: HooksConfig{
			PreToolUse: []HookRule{
				{Matcher: ".*", Hooks: []Hook{{Type: "command", Command: logHook}}},
			},
		},
	}

	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	fmt.Println("生成的 hooks 配置:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// 写入临时文件并运行演示
	tmpDir, err := os.MkdirTemp("", "claude-hooks-basics-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	settingsFile := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsFile, jsonData, 0644); err != nil {
		log.Fatalf("写入配置失败: %v", err)
	}

	fmt.Printf("临时配置文件: %s\n\n", settingsFile)
	fmt.Println("运行 claude（观察 stderr 中的 Hook 日志）:")

	cmd := exec.Command("claude",
		"-p", "请输出数字 42，不要做其他事",
		"--output-format", "json",
		"--model", "sonnet",
		"--settings", settingsFile,
	)
	cmd.Stderr = os.Stderr // 让 hook 的 stderr 输出可见

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("提示：如未安装 claude CLI，可跳过此步骤\n错误: %v\n", err)
	} else {
		var resp struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(output, &resp); err == nil {
			fmt.Printf("\nClaude 回复: %s\n", resp.Result)
		}
	}

	// ========== 总结 ==========
	fmt.Println()
	fmt.Println("━━━ 关键要点 ━━━")
	fmt.Println()
	fmt.Println("✅ Hook 通过 stdin 接收 JSON，不是环境变量")
	fmt.Println("✅ PreToolUse: 可阻断（exit 2 或 JSON deny）")
	fmt.Println("✅ PostToolUse: 不能阻断，用于后置处理")
	fmt.Println("✅ SessionStart: stdout 会注入 Claude 上下文")
	fmt.Println("✅ Stop: 会话结束后执行（通知、清理）")
	fmt.Println("✅ --settings 临时加载，不影响全局配置")
}
