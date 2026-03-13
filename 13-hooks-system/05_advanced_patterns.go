//go:build ignore

// 第十三章示例 5：高级 Hooks 组合模式
// 运行方式: go run 05_advanced_patterns.go
//
// 本示例演示：
// 1. 生成完整的综合安全审计系统（所有 9 个 hooks）
// 2. matcher 正则匹配高级用法（Write|Edit|MultiEdit）
// 3. JSON 输出决策的完整格式
// 4. 多 hooks 并发执行机制
// 5. 全局 vs 项目级 hooks 优先级

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	fmt.Println("║     第 13 章：Hooks 自动化守门人 - 高级组合模式       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	hooksDir, err := filepath.Abs("hooks")
	if err != nil {
		log.Fatalf("获取 hooks 目录失败: %v", err)
	}

	// ========== 第一节：Matcher 正则匹配高级用法 ==========
	fmt.Println("━━━ 1. Matcher 正则匹配 ━━━")
	fmt.Println()
	fmt.Println("matcher 字段支持正则表达式，匹配工具名称：")
	fmt.Println()
	fmt.Println(`  "Bash"          → 精确匹配 Bash 工具`)
	fmt.Println(`  "Write|Edit"    → 匹配 Write 或 Edit（OR）`)
	fmt.Println(`  "Write|Edit|MultiEdit" → 匹配三种写入工具`)
	fmt.Println(`  ".*"            → 匹配所有工具（通配符）`)
	fmt.Println(`  "^(?!Bash).*"   → 排除 Bash，匹配其他所有工具`)
	fmt.Println()
	fmt.Println("注意：matcher 是正则表达式，而非 glob 模式")
	fmt.Println()

	// ========== 第二节：JSON 输出决策完整格式 ==========
	fmt.Println("━━━ 2. JSON 输出决策完整格式 ━━━")
	fmt.Println()

	// PreToolUse deny 决策示例
	denyDecision := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":              "PreToolUse",
			"permissionDecision":         "deny",
			"permissionDecisionReason":   "检测到危险操作，已由安全守门人阻断",
		},
	}
	denyJSON, _ := json.MarshalIndent(denyDecision, "  ", "  ")
	fmt.Println("阻断决策（deny）:")
	fmt.Printf("  %s\n\n", denyJSON)

	// PreToolUse allow 决策示例（通常直接 exit 0 即可）
	allowDecision := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
		},
	}
	allowJSON, _ := json.MarshalIndent(allowDecision, "  ", "  ")
	fmt.Println("明确允许决策（allow）:")
	fmt.Printf("  %s\n\n", allowJSON)

	fmt.Println("退出码含义：")
	fmt.Println("  exit 0  → 正常（允许，JSON deny 也用 exit 0）")
	fmt.Println("  exit 1  → 错误（非阻断，Claude 收到错误信息）")
	fmt.Println("  exit 2  → 阻断（PreToolUse 专用，拒绝工具执行）")
	fmt.Println()

	// ========== 第三节：多 Hooks 执行顺序 ==========
	fmt.Println("━━━ 3. 多 Hooks 执行机制 ━━━")
	fmt.Println()
	fmt.Println("同一事件有多个 hooks 时的执行规则：")
	fmt.Println()
	fmt.Println("  [同一 matcher 的多个 hooks]")
	fmt.Println("    → 按配置顺序依次执行（串行）")
	fmt.Println("    → 任意一个 exit 2 或 deny，后续 hooks 不再执行")
	fmt.Println()
	fmt.Println("  [不同 matcher 但都匹配的 hooks]")
	fmt.Println("    → 并发执行（Claude Code 会并行运行）")
	fmt.Println("    → 所有 hooks 完成后才继续")
	fmt.Println()
	fmt.Println("  [PreToolUse 的短路逻辑]")
	fmt.Println("    → 一旦收到 deny 决策，立即停止，不执行工具")
	fmt.Println()

	// ========== 第四节：生成完整综合配置 ==========
	fmt.Println("━━━ 4. 完整综合审计系统配置 ━━━")
	fmt.Println()

	settings := Settings{
		Hooks: HooksConfig{
			SessionStart: []HookRule{
				{
					Matcher: "",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "07_session_check.sh")},
					},
				},
			},
			PreToolUse: []HookRule{
				{
					// 安全守门人：拦截危险 Bash 命令
					Matcher: "Bash",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "01_security_guard.sh")},
					},
				},
				{
					// 敏感文件保护：阻止覆写 .env / credentials
					Matcher: "Write",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "02_file_protector.sh")},
					},
				},
				{
					// 自动备份：Write/Edit/MultiEdit 前先备份
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "03_backup_before_edit.sh")},
					},
				},
			},
			PostToolUse: []HookRule{
				{
					// 敏感数据检测：写入后扫描 API Key / 密码
					Matcher: "Write",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "04_sensitive_data.sh")},
					},
				},
				{
					// 自动格式化：按扩展名运行格式化工具
					Matcher: "Write",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "05_auto_format.sh")},
					},
				},
				{
					// 自动测试：源码变更后运行测试套件
					Matcher: "Write",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "06_run_tests.sh")},
					},
				},
				{
					// 命令审计：记录所有 Bash 命令到审计日志
					Matcher: "Bash",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "09_audit_log.sh")},
					},
				},
			},
			Stop: []HookRule{
				{
					// 任务完成通知：macOS 桌面通知 + 耗时摘要
					Matcher: "",
					Hooks: []Hook{
						{Type: "command", Command: filepath.Join(hooksDir, "08_stop_notify.sh")},
					},
				},
			},
		},
	}

	jsonData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	fmt.Println("完整综合 hooks 配置（可直接复制到 ~/.claude/settings.json）:")
	fmt.Println(string(jsonData))
	fmt.Println()

	// 保存到文件
	outputFile := "complete_hooks_config.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("写入失败: %v", err)
	}
	fmt.Printf("✅ 配置已保存到: %s\n\n", outputFile)

	// ========== 第五节：全局 vs 项目级配置 ==========
	fmt.Println("━━━ 5. 全局 vs 项目级 Hooks 优先级 ━━━")
	fmt.Println()
	fmt.Println("配置文件位置（优先级从高到低）：")
	fmt.Println()
	fmt.Println("  1. 命令行 --settings <file>    最高优先级（临时覆盖）")
	fmt.Println("  2. .claude/settings.local.json  项目级个人配置（不提交 git）")
	fmt.Println("  3. .claude/settings.json        项目级配置（提交 git，团队共享）")
	fmt.Println("  4. ~/.claude/settings.json      全局用户配置（所有项目生效）")
	fmt.Println()
	fmt.Println("Hooks 的合并规则：")
	fmt.Println("  • 同一事件（如 PreToolUse）的 hooks 会合并，不会覆盖")
	fmt.Println("  • 全局 hooks + 项目 hooks 都会触发")
	fmt.Println("  • 执行顺序：全局 hooks 先执行，项目 hooks 后执行")
	fmt.Println()

	// ========== 第六节：查看审计日志 ==========
	fmt.Println("━━━ 6. 查看审计日志 ━━━")
	fmt.Println()
	fmt.Println("运行过程中，hooks 会写入以下日志文件：")
	fmt.Println()
	fmt.Println("  /tmp/claude_audit.log       — Bash 命令审计（09_audit_log.sh）")
	fmt.Println("  /tmp/claude_hooks_audit.log — 安全事件记录（security/file/sensitive hooks）")
	fmt.Println("  /tmp/claude_backup/         — 文件备份目录（03_backup_before_edit.sh）")
	fmt.Println()
	fmt.Println("查看审计日志命令：")
	fmt.Println()
	fmt.Println("  # 实时查看 Bash 审计日志")
	fmt.Println("  tail -f /tmp/claude_audit.log")
	fmt.Println()
	fmt.Println("  # 查看安全事件")
	fmt.Println("  cat /tmp/claude_hooks_audit.log")
	fmt.Println()
	fmt.Println("  # 查看备份文件")
	fmt.Println("  ls -la /tmp/claude_backup/$(date +%Y%m%d)/")
	fmt.Println()

	// ========== 总结 ==========
	fmt.Println("━━━ 综合总结：Hooks 设计原则 ━━━")
	fmt.Println()
	fmt.Println("✅ 正确做法:")
	fmt.Println("   • 用 INPUT=$(cat) 读取 stdin JSON，不要用环境变量")
	fmt.Println("   • PreToolUse 阻断用 exit 2 或 JSON deny + exit 0")
	fmt.Println("   • PostToolUse 只能建议（stderr 警告），不能阻断")
	fmt.Println("   • SessionStart stdout 会成为 Claude 的初始上下文")
	fmt.Println("   • 保持 hook 脚本快速（< 5s），避免拖慢 Claude 响应")
	fmt.Println()
	fmt.Println("❌ 常见错误:")
	fmt.Println("   • 用环境变量（如 $CLAUDE_TOOL_NAME）获取工具信息 → 错误！")
	fmt.Println("   • PreToolUse 用 exit 1 试图阻断 → 无效（只是报错）")
	fmt.Println("   • Hook 脚本执行耗时操作（网络请求、大量计算）")
	fmt.Println("   • 忘记对 hook 脚本添加执行权限（chmod +x）")
}
