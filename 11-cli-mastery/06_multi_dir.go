//go:build ignore

// 第十一章示例 6：多目录上下文
// 运行方式: go run 06_multi_dir.go
//
// 本示例演示如何：
// 1. 使用 --add-dir 让 Claude 同时理解多个目录的代码
// 2. 跨目录代码分析和对比
// 3. Monorepo 和多项目场景的实践
//
// --add-dir 让 Claude 在回答问题时可以访问额外目录中的文件

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
)

// MultiDirResponse 解析 JSON 输出
type MultiDirResponse struct {
	Type   string `json:"type"`
	Result string `json:"result"`
}

func main() {
	// 获取项目根目录
	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(currentFile))

	// ========== 场景 1：跨章节代码分析 ==========
	fmt.Println("=== 场景 1：跨章节代码分析 ===")
	fmt.Println("使用 --add-dir 让 Claude 同时看到多个章节的代码")
	fmt.Println()

	// 同时分析第 0 章和第 1 章的代码
	chapter00 := filepath.Join(projectRoot, "00-cli-quick-start")
	chapter01 := filepath.Join(projectRoot, "01-api-basics")

	cmd1 := exec.Command("claude",
		"-p", "对比这两个目录的代码风格，列出 3 个主要区别",
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "3",
		"--add-dir", chapter00, // 添加第 0 章目录
		"--add-dir", chapter01, // 添加第 1 章目录
	)

	resp1 := runMultiDir(cmd1)
	fmt.Printf("分析结果:\n%s\n", truncateResult(resp1.Result, 300))

	// ========== 场景 2：Monorepo 场景 ==========
	fmt.Println("\n=== 场景 2：Monorepo 场景 ===")
	fmt.Println("在 monorepo 中，需要同时理解多个子项目的代码")
	fmt.Println()

	fmt.Println("示例命令:")
	fmt.Println("  claude -p '分析前后端的 API 接口是否一致' \\")
	fmt.Println("    --add-dir ./frontend/src/api \\")
	fmt.Println("    --add-dir ./backend/handlers")
	fmt.Println()
	fmt.Println("  claude -p '检查共享类型定义的一致性' \\")
	fmt.Println("    --add-dir ./packages/shared-types \\")
	fmt.Println("    --add-dir ./apps/web/src/types \\")
	fmt.Println("    --add-dir ./apps/mobile/src/types")

	// ========== 场景 3：跨项目重构 ==========
	fmt.Println("\n=== 场景 3：跨项目重构 ===")
	fmt.Println("在重构时，需要了解调用方和被调用方的代码")
	fmt.Println()

	fmt.Println("示例命令:")
	fmt.Println("  claude -p '将这个工具函数从 utils 迁移到 common，并更新所有引用' \\")
	fmt.Println("    --add-dir ./libs/utils \\")
	fmt.Println("    --add-dir ./libs/common \\")
	fmt.Println("    --add-dir ./apps/server/src")

	// ========== 注意事项 ==========
	fmt.Println("\n=== --add-dir 注意事项 ===")
	fmt.Println("1. --add-dir 的路径可以是相对路径或绝对路径")
	fmt.Println("2. 添加的目录越多，Claude 需要处理的上下文越大")
	fmt.Println("3. 建议只添加相关的子目录，而非整个项目根目录")
	fmt.Println("4. 配合 --allowedTools Read Grep Glob 使用更安全")
	fmt.Println("5. Claude 对添加的目录拥有和工作目录相同的访问权限")
}

// runMultiDir 执行多目录命令
func runMultiDir(cmd *exec.Cmd) MultiDirResponse {
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("多目录调用失败: %v", err)
	}

	var resp MultiDirResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析 JSON 失败: %v\n原始输出: %s", err, string(output))
	}

	return resp
}

// truncateResult 截断结果字符串
func truncateResult(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
