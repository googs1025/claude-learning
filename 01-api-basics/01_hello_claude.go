// 第一章示例 1：最简单的 Claude API 调用
// 运行方式: go run 01_hello_claude.go
//
// 本示例演示如何：
// 1. 创建 Anthropic 客户端（自动读取 ANTHROPIC_API_KEY 环境变量）
// 2. 构造用户消息并发送请求
// 3. 解析并打印 Claude 的响应内容

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// 创建客户端
	// NewClient() 会自动从环境变量 ANTHROPIC_API_KEY 中读取 API 密钥
	// 如果环境变量未设置，运行时会报错
	client := anthropic.NewClient()

	// 创建上下文，用于控制请求的生命周期
	ctx := context.Background()

	// 发送消息请求
	// MessageNewParams 是请求参数的结构体，包含以下核心字段：
	// - Model: 指定使用的模型（这里使用 Claude Sonnet 4.5）
	// - MaxTokens: 限制回复的最大 token 数量
	// - Messages: 消息列表，至少包含一条用户消息
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929, // 使用 Claude Sonnet 4.5 模型
		MaxTokens: 1024,                                    // 最多返回 1024 个 token
		Messages: []anthropic.MessageParam{
			// NewUserMessage 创建一条用户消息
			// NewTextBlock 将纯文本包装为内容块
			anthropic.NewUserMessage(anthropic.NewTextBlock("你好，Claude！请用一句话介绍你自己。")),
		},
	})
	if err != nil {
		// 如果请求失败（网络错误、认证失败等），打印错误并退出
		log.Fatalf("API 调用失败: %v", err)
	}

	// 解析响应内容
	// Claude 的响应是一个 Content 数组，可能包含多个内容块（文本、工具调用等）
	// 我们遍历每个内容块，根据类型进行处理
	fmt.Println("=== Claude 的回复 ===")
	for _, block := range message.Content {
		// AsAny() 将内容块转换为具体类型
		// 使用 type switch 判断内容块的实际类型
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			// 文本类型的内容块，直接打印其文本内容
			fmt.Println(v.Text)
		default:
			// 其他类型的内容块（如工具调用），在本示例中不做处理
			fmt.Printf("收到非文本内容块，类型: %T\n", v)
		}
	}

	// 打印一些响应的元数据信息，帮助理解 API 的计费方式
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)  // 请求消耗的 token
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens) // 响应消耗的 token
	fmt.Printf("模型: %s\n", message.Model)                      // 实际使用的模型名称
	fmt.Printf("停止原因: %s\n", message.StopReason)               // 停止生成的原因（end_turn 表示自然结束）
}
