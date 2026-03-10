//go:build ignore

// 第零章示例 3：流式输出
// 运行方式: go run 03_streaming.go
//
// 本示例演示如何：
// 1. 使用 --output-format stream-json 获取流式 JSON 输出
// 2. 通过 StdoutPipe 逐行读取输出
// 3. 解析 stream-json 格式，实时显示生成的文本
//
// stream-json 格式：每行是一个 JSON 对象，结构为：
// {"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"..."}}}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// StreamMessage 表示流式输出的顶层消息
type StreamMessage struct {
	Type  string      `json:"type"`  // 顶层类型：stream_event, result
	Event StreamEvent `json:"event"` // 事件详情
}

// StreamEvent 表示流中的一个事件
type StreamEvent struct {
	Type  string     `json:"type"` // 事件类型：content_block_delta, message_start 等
	Delta EventDelta `json:"delta"`
}

// EventDelta 表示增量内容
type EventDelta struct {
	Type string `json:"type"` // delta 类型：text_delta
	Text string `json:"text"` // 实际文本内容
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
	// 增大缓冲区以处理可能的长行 JSON
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// 解析每行 JSON
		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		// 提取文本增量：stream_event → content_block_delta → text_delta
		if msg.Type == "stream_event" &&
			msg.Event.Type == "content_block_delta" &&
			msg.Event.Delta.Type == "text_delta" {
			fmt.Print(msg.Event.Delta.Text)
		}
	}

	fmt.Println()

	// 等待命令执行完毕
	if err := cmd.Wait(); err != nil {
		log.Fatalf("命令执行失败: %v", err)
	}

	fmt.Println("\n=== 流式传输完成 ===")
}
