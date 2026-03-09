//go:build ignore

// 第零章示例 3：流式输出
// 运行方式: go run 03_streaming.go
//
// 本示例演示如何：
// 1. 使用 --output-format stream-json 获取流式 JSON 输出
// 2. 通过 StdoutPipe 逐行读取输出
// 3. 解析 stream-json 格式，实时显示生成的文本
//
// stream-json 格式：每行是一个 JSON 对象，包含 type 字段
// 我们关注 type="assistant" 的事件，其中包含生成的文本内容

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// StreamEvent 表示流式输出的一个事件
type StreamEvent struct {
	Type    string `json:"type"`              // 事件类型：assistant, result, 等
	Content string `json:"content,omitempty"` // 文本内容（type=assistant 时）
}

func main() {
	prompt := "请用中文简要介绍 Go 语言的三个核心特性，每个特性用一段话说明。"

	// 构建命令：使用 stream-json 输出格式
	cmd := exec.Command("claude", "-p", prompt, "--output-format", "stream-json", "--model", "sonnet")

	// 获取标准输出管道，用于逐行读取
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("创建输出管道失败: %v", err)
	}

	// 启动命令（异步执行，不等待完成）
	if err := cmd.Start(); err != nil {
		log.Fatalf("启动命令失败: %v", err)
	}

	fmt.Println("=== Claude 的流式回复 ===")
	fmt.Println("（以下内容将逐步显示，模拟实时生成效果）")
	fmt.Println()

	// 使用 Scanner 逐行读取流式输出
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		// 解析每行 JSON
		var event StreamEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// 跳过无法解析的行
			continue
		}

		// 只处理 assistant 类型的事件（包含实际生成的文本）
		if event.Type == "assistant" && event.Content != "" {
			fmt.Print(event.Content)
		}
	}

	fmt.Println()

	// 等待命令执行完毕
	if err := cmd.Wait(); err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}

	fmt.Println("\n=== 流式传输完成 ===")
}
