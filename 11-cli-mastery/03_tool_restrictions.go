//go:build ignore

// 第十一章示例 3：工具限制
// 运行方式: go run 03_tool_restrictions.go
//
// 本示例演示如何：
// 1. 使用 --allowedTools 限制 Claude 只能使用指定工具（白名单）
// 2. 使用 --disallowedTools 禁止使用特定工具（黑名单）
// 3. 使用 Bash 通配符限制特定命令
//
// 工具限制是构建安全沙箱的关键能力：
// - 代码审查场景：只允许 Read、Grep、Glob（只读）
// - 受限开发：禁止 Write、Edit（不能修改文件）
// - 安全执行：Bash(git log *) 只允许特定命令

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// ToolResponse 用于解析 JSON 输出
type ToolResponse struct {
	Type   string `json:"type"`
	Result string `json:"result"`
}

func main() {
	// ========== 场景 1：只读模式（白名单） ==========
	fmt.Println("=== 场景 1：只读模式 - allowedTools ===")
	fmt.Println("只允许 Read、Grep、Glob 三个工具，Claude 无法修改任何文件")
	fmt.Println()

	readOnlyCmd := exec.Command("claude",
		"-p", "请分析当前目录下的 Go 文件结构，列出所有 .go 文件",
		"--output-format", "json",
		"--model", "sonnet",
		"--allowedTools", "Read", "Grep", "Glob", // 只允许读取类工具
	)

	runAndPrint(readOnlyCmd, "只读模式")

	// ========== 场景 2：禁止修改（黑名单） ==========
	fmt.Println("\n=== 场景 2：禁止修改 - disallowedTools ===")
	fmt.Println("禁止 Edit 和 Write 工具，其他工具均可使用")
	fmt.Println()

	noWriteCmd := exec.Command("claude",
		"-p", "查看当前目录结构并说明各文件的作用",
		"--output-format", "json",
		"--model", "sonnet",
		"--disallowedTools", "Edit", "Write", // 禁止修改类工具
	)

	runAndPrint(noWriteCmd, "禁止修改模式")

	// ========== 场景 3：限制 Bash 命令 ==========
	fmt.Println("\n=== 场景 3：限制 Bash 命令 ===")
	fmt.Println("使用 Bash(pattern) 语法只允许特定的 shell 命令")
	fmt.Println()

	// Bash(git log *) 表示只允许以 "git log" 开头的 Bash 命令
	// Bash(go test *) 表示只允许以 "go test" 开头的 Bash 命令
	limitedBashCmd := exec.Command("claude",
		"-p", "查看最近 3 条 git 提交记录",
		"--output-format", "json",
		"--model", "sonnet",
		"--allowedTools", "Bash(git log *)", "Bash(git diff *)", "Read", "Grep", "Glob",
	)

	runAndPrint(limitedBashCmd, "受限 Bash 模式")

	// ========== 工具列表参考 ==========
	fmt.Println("\n=== 常用工具名称参考 ===")
	fmt.Println("读取类: Read, Grep, Glob")
	fmt.Println("修改类: Edit, Write, NotebookEdit")
	fmt.Println("执行类: Bash")
	fmt.Println("其他类: Agent, WebFetch, WebSearch")
	fmt.Println()
	fmt.Println("Bash 通配符示例:")
	fmt.Println("  Bash(git *)       - 所有 git 命令")
	fmt.Println("  Bash(go test *)   - go test 命令")
	fmt.Println("  Bash(npm run *)   - npm 脚本")
}

// runAndPrint 执行命令并打印结果
func runAndPrint(cmd *exec.Cmd, label string) {
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("%s 调用失败: %v", label, err)
	}

	var resp ToolResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("%s 解析失败: %v\n原始输出: %s", label, err, string(output))
	}

	// 截取前 200 字符展示
	result := []rune(resp.Result)
	if len(result) > 200 {
		fmt.Printf("[%s] 回复: %s...\n", label, string(result[:200]))
	} else {
		fmt.Printf("[%s] 回复: %s\n", label, resp.Result)
	}
}
