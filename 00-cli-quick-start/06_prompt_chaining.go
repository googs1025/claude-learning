//go:build ignore

// 第零章示例 6：链式调用（Prompt Chaining）
// 运行方式: go run 06_prompt_chaining.go
//
// 本示例演示如何：
// 1. 将一个复杂任务拆分为多个步骤
// 2. 每个步骤调用一次 Claude
// 3. 前一步的输出作为后一步的输入
//
// 这是一种常见的 Agent 模式：将复杂问题分解为可管理的子任务

package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// askClaude 发送一个提示词给 Claude 并返回回复
func askClaude(prompt string) string {
	cmd := exec.Command("claude", "-p", prompt, "--model", "sonnet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v\n输出: %s", err, string(output))
	}
	return strings.TrimSpace(string(output))
}

func main() {
	fmt.Println("=== 链式调用演示：代码生成 → 代码审查 → 生成文档 ===")
	fmt.Println()

	// 步骤 1：生成代码
	fmt.Println("--- 步骤 1：生成代码 ---")
	code := askClaude("用 Go 语言写一个简单的 Stack（栈）数据结构，包含 Push、Pop、Peek、IsEmpty 方法。只输出代码，不要解释。")
	fmt.Println(code)
	fmt.Println()

	// 步骤 2：审查步骤 1 生成的代码
	fmt.Println("--- 步骤 2：代码审查 ---")
	reviewPrompt := fmt.Sprintf("请审查以下 Go 代码，指出潜在问题并给出改进建议（简要列出要点即可）：\n\n```go\n%s\n```", code)
	review := askClaude(reviewPrompt)
	fmt.Println(review)
	fmt.Println()

	// 步骤 3：基于代码和审查结果生成文档
	fmt.Println("--- 步骤 3：生成使用文档 ---")
	docPrompt := fmt.Sprintf("根据以下 Go 代码和审查意见，生成一份简短的使用文档（包含功能说明和使用示例）。\n\n代码：\n```go\n%s\n```\n\n审查意见：\n%s", code, review)
	doc := askClaude(docPrompt)
	fmt.Println(doc)
}
