//go:build ignore

// 第零章示例 4：使用系统提示词
// 运行方式: go run 04_system_prompt.go
//
// 本示例演示如何：
// 1. 使用 --append-system-prompt 参数设置系统提示词
// 2. 通过系统提示词控制 Claude 的角色和行为
// 3. 对比有无系统提示词的输出差异
//
// 系统提示词用于定义 Claude 的角色、语气、输出格式等

package main

import (
	"fmt"
	"log"
	"os/exec"
)

// callClaudeWithSystem 使用系统提示词调用 Claude
func callClaudeWithSystem(prompt, systemPrompt string) string {
	cmd := exec.Command("claude",
		"-p", prompt,
		"--append-system-prompt", systemPrompt,
		"--model", "sonnet",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v\n输出: %s", err, string(output))
	}
	return string(output)
}

func main() {
	question := "什么是 goroutine？"

	// 场景 1：作为技术教师
	fmt.Println("=== 场景 1：技术教师角色 ===")
	reply1 := callClaudeWithSystem(question,
		"你是一位耐心的编程教师，面向初学者解释概念。使用简单的类比和例子，避免使用过多技术术语。回答控制在 3-4 句话以内。",
	)
	fmt.Println(reply1)

	// 场景 2：作为技术文档作者
	fmt.Println("=== 场景 2：技术文档作者角色 ===")
	reply2 := callClaudeWithSystem(question,
		"你是一位严谨的技术文档作者。回答要精确、专业，使用正式的技术术语。输出格式为 Markdown，包含定义和要点列表。",
	)
	fmt.Println(reply2)

	// 场景 3：作为诗人
	fmt.Println("=== 场景 3：诗人角色 ===")
	reply3 := callClaudeWithSystem(question,
		"你是一位诗人。用一首短诗（4-6行）来解释用户的问题，要求押韵、有意境。",
	)
	fmt.Println(reply3)
}
