//go:build ignore

// 第十三章示例 2：安全防御 Hooks
// 运行方式: go run 02_security_hooks.go
//
// 本示例演示：
// 1. 生成包含安全类 hooks 的 settings.json
// 2. 危险命令拦截（rm -rf /、sudo rm 等）
// 3. 敏感文件保护（.env、credentials 等）
// 4. 覆写前自动备份
// 5. 用 --settings 临时加载配置运行演示

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// ========== Settings 结构体 ==========

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
	SessionStart []HookRule `json:"SessionStart,omitempty"`
	Stop        []HookRule `json:"Stop,omitempty"`
}

type Settings struct {
	Hooks HooksConfig `json:"hooks"`
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║     第 13 章：Hooks 自动化守门人 - 安全防御           ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// ========== 获取 hooks 目录路径 ==========
	// 假设从 13-hooks-system/ 目录运行
	hooksDir, err := filepath.Abs("hooks")
	if err != nil {
		log.Fatalf("获取 hooks 目录失败: %v", err)
	}

	fmt.Printf("Hooks 目录: %s\n\n", hooksDir)

	// ========== 构建安全 Hooks 配置 ==========
	fmt.Println("━━━ 安全 Hooks 配置 ━━━")
	fmt.Println()

	settings := Settings{
		Hooks: HooksConfig{
			PreToolUse: []HookRule{
				{
					// Hook 1: 危险命令拦截
					// 匹配所有 Bash 工具调用
					Matcher: "Bash",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "01_security_guard.sh"),
						},
					},
				},
				{
					// Hook 2: 敏感文件保护
					// 匹配 Write 工具（创建/覆写文件）
					Matcher: "Write",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "02_file_protector.sh"),
						},
					},
				},
				{
					// Hook 3: 覆写前自动备份
					// 匹配 Write 和 Edit 工具（用正则 OR）
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: filepath.Join(hooksDir, "03_backup_before_edit.sh"),
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

	fmt.Println("生成的安全 hooks 配置:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// ========== 说明各 Hook 的工作原理 ==========
	fmt.Println("━━━ 各 Hook 功能说明 ━━━")
	fmt.Println()
	fmt.Println("1. 01_security_guard.sh (PreToolUse + Bash)")
	fmt.Println("   • 检测模式: rm -rf /, sudo rm, dd if=/dev/zero, fork炸弹等")
	fmt.Println("   • 阻断方式: 输出 JSON {permissionDecision: deny} + exit 0")
	fmt.Println("   • 发现危险命令时写入 /tmp/claude_hooks_audit.log")
	fmt.Println()
	fmt.Println("2. 02_file_protector.sh (PreToolUse + Write)")
	fmt.Println("   • 保护文件: .env, credentials, *.pem, *.key, id_rsa 等")
	fmt.Println("   • 阻断方式: 输出 JSON {permissionDecision: deny} + exit 0")
	fmt.Println("   • 防止 AI 直接覆写生产环境密钥文件")
	fmt.Println()
	fmt.Println("3. 03_backup_before_edit.sh (PreToolUse + Write|Edit|MultiEdit)")
	fmt.Println("   • 在任何文件被修改前，先复制到 /tmp/claude_backup/YYYYMMDD/")
	fmt.Println("   • 文件名格式: HHMMSS_原路径（/ 替换为 _）")
	fmt.Println("   • 不阻断操作，只做备份（辅助安全网）")
	fmt.Println()

	// ========== 写入临时文件并演示 ==========
	fmt.Println("━━━ 演示：尝试执行危险命令 ━━━")
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", "claude-security-hooks-*")
	if err != nil {
		log.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	settingsFile := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsFile, jsonData, 0644); err != nil {
		log.Fatalf("写入配置失败: %v", err)
	}

	fmt.Printf("临时配置: %s\n\n", settingsFile)

	// 让 Claude 尝试执行一个会被 security_guard 拦截的命令
	prompt := `请执行以下命令来测试：echo "安全测试" && echo "这是一个测试"。
只输出命令执行结果，不要解释。`

	fmt.Printf("提示词: %s\n\n", prompt)
	fmt.Println("（注意：如果 Claude 尝试执行危险命令，hook 会拦截；echo 命令是安全的）")
	fmt.Println()

	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--model", "sonnet",
		"--settings", settingsFile,
		"--allowedTools", "Bash",
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
			result := []rune(resp.Result)
			if len(result) > 300 {
				fmt.Printf("Claude 回复: %s...\n", string(result[:300]))
			} else {
				fmt.Printf("Claude 回复: %s\n", resp.Result)
			}
		}
	}

	// ========== 手动测试说明 ==========
	fmt.Println()
	fmt.Println("━━━ 手动测试 Hook 脚本 ━━━")
	fmt.Println()
	fmt.Println("# 测试危险命令拦截（应返回 JSON deny）:")
	fmt.Printf("echo '{\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"rm -rf /\"}}' | bash %s\n\n",
		filepath.Join(hooksDir, "01_security_guard.sh"))
	fmt.Println("# 测试安全命令（应静默通过，exit 0）:")
	fmt.Printf("echo '{\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"ls -la\"}}' | bash %s; echo \"退出码: $?\"\n\n",
		filepath.Join(hooksDir, "01_security_guard.sh"))
	fmt.Println("# 测试敏感文件保护（应返回 JSON deny）:")
	fmt.Printf("echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/project/.env\"}}' | bash %s\n",
		filepath.Join(hooksDir, "02_file_protector.sh"))
}
