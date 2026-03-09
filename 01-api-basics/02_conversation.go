// 第一章示例 2：多轮对话
// 运行方式: go run 02_conversation.go
//
// 本示例演示如何：
// 1. 发送第一条消息并获取回复
// 2. 使用 ToParam() 将 Claude 的回复转换为可发送的消息格式
// 3. 追加新的用户消息，实现多轮对话
// 4. Claude 能够理解并引用之前的对话内容

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

	// ==================== 第一轮对话 ====================
	fmt.Println("========== 第一轮对话 ==========")
	fmt.Println("[用户]: 我叫小明，我是一名 Go 语言开发者。请记住我的信息。")

	// 初始化消息列表，包含第一条用户消息
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("我叫小明，我是一名 Go 语言开发者。请记住我的信息。")),
	}

	// 发送第一轮请求
	response1, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages:  messages,
	})
	if err != nil {
		log.Fatalf("第一轮对话失败: %v", err)
	}

	// 打印第一轮回复
	fmt.Print("[Claude]: ")
	printResponse(response1)

	// ==================== 第二轮对话 ====================
	fmt.Println("\n========== 第二轮对话 ==========")
	fmt.Println("[用户]: 我叫什么名字？我的职业是什么？")

	// 关键步骤：将 Claude 的回复转换为消息参数并追加到消息列表
	// ToParam() 方法将 API 响应转换为 MessageParam 格式
	// 这样 Claude 就能"看到"之前的对话历史
	messages = append(messages, response1.ToParam())

	// 追加第二轮用户消息
	messages = append(messages, anthropic.NewUserMessage(
		anthropic.NewTextBlock("我叫什么名字？我的职业是什么？"),
	))

	// 发送第二轮请求（包含完整对话历史）
	response2, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages:  messages, // 包含了之前所有的消息
	})
	if err != nil {
		log.Fatalf("第二轮对话失败: %v", err)
	}

	// 打印第二轮回复（Claude 应该能记住小明的信息）
	fmt.Print("[Claude]: ")
	printResponse(response2)

	// ==================== 第三轮对话 ====================
	fmt.Println("\n========== 第三轮对话 ==========")
	fmt.Println("[用户]: 根据我的背景，推荐一本适合我的技术书籍。")

	// 继续追加对话历史
	messages = append(messages, response2.ToParam())
	messages = append(messages, anthropic.NewUserMessage(
		anthropic.NewTextBlock("根据我的背景，推荐一本适合我的技术书籍。"),
	))

	// 发送第三轮请求
	response3, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages:  messages,
	})
	if err != nil {
		log.Fatalf("第三轮对话失败: %v", err)
	}

	// Claude 应该能根据"Go 语言开发者"的背景给出针对性推荐
	fmt.Print("[Claude]: ")
	printResponse(response3)

	// 打印对话统计信息
	fmt.Println("\n========== 对话统计 ==========")
	fmt.Printf("总共进行了 3 轮对话\n")
	fmt.Printf("最后一轮 - 输入 token: %d, 输出 token: %d\n",
		response3.Usage.InputTokens, response3.Usage.OutputTokens)
	fmt.Println("提示：随着对话轮数增加，输入 token 数会不断增长，因为每次都发送了完整历史")
}

// printResponse 从响应中提取并打印文本内容
// 这是一个辅助函数，用于遍历响应的内容块并输出文本
func printResponse(msg *anthropic.Message) {
	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}
}
