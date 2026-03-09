// 第一章示例 3：流式响应
// 运行方式: go run 03_streaming.go
//
// 本示例演示如何：
// 1. 使用 NewStreaming 发起流式请求
// 2. 实时接收并处理流式事件
// 3. 逐字打印 Claude 的回复，模拟打字机效果
//
// 流式响应的优势：
// - 用户可以立即看到回复的开头，无需等待完整响应生成
// - 提升用户体验，特别是在回复较长时
// - 适合聊天界面、CLI 工具等需要实时展示的场景

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// 创建客户端和上下文
	client := anthropic.NewClient()
	ctx := context.Background()

	// 构造请求参数（与非流式请求完全相同）
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"请用中文简要介绍 Go 语言的三个核心特性，每个特性用一段话说明。",
			)),
		},
	}

	// 发起流式请求
	// NewStreaming 返回一个 stream 对象，而不是直接返回完整消息
	// 我们通过循环读取 stream 来逐步获取内容
	fmt.Println("=== Claude 的流式回复 ===")
	fmt.Println("（以下内容将逐步显示，模拟实时生成效果）")
	fmt.Println()

	stream := client.Messages.NewStreaming(ctx, params)

	// 创建一个空消息，用于通过 Accumulate 逐步构建完整消息
	message := anthropic.Message{}

	// 使用 Next() 迭代流中的事件
	// Next() 返回 true 表示还有更多事件，返回 false 表示流结束
	for stream.Next() {
		// Current() 获取当前事件
		event := stream.Current()

		// 使用 Accumulate 将每个事件累积到 message 中
		// 流结束后 message 将包含完整的响应信息（内容、usage 统计等）
		message.Accumulate(event)

		// 使用 type switch 处理不同类型的事件
		// 流式 API 会发送多种事件类型，我们主要关注文本增量事件
		switch v := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			// ContentBlockDeltaEvent 表示内容块的增量更新
			// 进一步检查 delta 的具体类型
			switch delta := v.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				// TextDelta 包含一小段新生成的文本
				// 使用 Print（而非 Println）避免额外换行，实现逐字输出效果
				fmt.Print(delta.Text)
			}
		}
	}

	// 流结束后换行
	fmt.Println()

	// 检查流式传输过程中是否发生错误
	// 即使中间有内容输出，也需要检查最终是否有错误
	if err := stream.Err(); err != nil {
		log.Fatalf("流式传输出错: %v", err)
	}

	// 获取最终的完整消息
	// 通过 Accumulate 逐步构建的 message 包含完整的响应信息，包括 usage 统计
	// 注意：SDK 没有 stream.FinalMessage() 方法，需要在循环中用 message.Accumulate(event) 构建
	fmt.Println("\n=== 流式传输完成 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
	fmt.Printf("停止原因: %s\n", message.StopReason)
}
