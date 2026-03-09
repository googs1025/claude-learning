//go:build ignore

// 第零章示例 1：最简单的 Claude CLI 调用
// 运行方式: go run 01_hello_claude.go
//
// 本示例演示如何：
// 1. 使用 os/exec 调用 claude CLI（无需 API Key）
// 2. 通过 -p 参数传入提示词
// 3. 获取并打印 Claude 的回复
//
// 前置条件：安装 claude CLI（npm install -g @anthropic-ai/claude-code）

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// 定义提示词
	prompt := "你好，Claude！请用一句话介绍你自己。并且你是什么模型"

	// 使用 exec.Command 调用 claude CLI
	// -p 参数表示非交互模式（pipe mode），直接传入提示词并返回结果
	// --model sonnet 选择 Sonnet 模型
	cmd := exec.Command("claude", "-p", prompt, "--model", "sonnet")

	// 运行命令并捕获标准输出
	// CombinedOutput 会等待命令执行完毕，返回 stdout + stderr 的合并输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("调用 claude CLI 失败: %v\n输出: %s", err, string(output))
	}

	// 打印 Claude 的回复
	fmt.Println("=== Claude 的回复 ===")
	fmt.Println(string(output))
}
