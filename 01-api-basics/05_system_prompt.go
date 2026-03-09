// 第一章示例 5：使用系统提示词
// 运行方式: go run 05_system_prompt.go
//
// 本示例演示如何：
// 1. 使用系统提示词（System Prompt）控制 Claude 的行为
// 2. 通过不同的系统提示词实现不同的角色/人设
// 3. 对比有无系统提示词时 Claude 的回复差异
//
// 系统提示词的用途：
// - 设定 Claude 的角色和身份（如：你是一位资深 Go 开发者）
// - 规定回复的格式和风格（如：用简洁的技术语言回答）
// - 添加约束条件（如：只回答编程相关问题）
// - 提供背景知识和上下文

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// 所有场景使用相同的用户问题，以便对比不同系统提示词的效果
	userQuestion := "如何处理并发？"

	// ==================== 场景 1：无系统提示词 ====================
	fmt.Println("========== 场景 1：无系统提示词（默认行为） ==========")
	fmt.Printf("[用户]: %s\n", userQuestion)

	response1, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
		},
		// 注意：这里没有设置 System 字段
	})
	if err != nil {
		log.Fatalf("场景 1 失败: %v", err)
	}

	fmt.Print("[Claude]: ")
	printResponse(response1)

	// ==================== 场景 2：Go 语言专家角色 ====================
	fmt.Println("\n========== 场景 2：Go 语言专家 ==========")
	fmt.Printf("[用户]: %s\n", userQuestion)

	response2, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
		},
		// System 字段接受 TextBlockParam 切片
		// 可以设置一个或多个系统提示词文本块
		System: []anthropic.TextBlockParam{
			{
				Text: "你是一位资深的 Go 语言专家，拥有 10 年的 Go 开发经验。" +
					"请始终从 Go 语言的角度回答问题，使用 Go 的术语和概念。" +
					"回答要简洁明了，包含代码示例。所有回复使用中文。",
			},
		},
	})
	if err != nil {
		log.Fatalf("场景 2 失败: %v", err)
	}

	fmt.Print("[Claude]: ")
	printResponse(response2)

	// ==================== 场景 3：诗人角色 ====================
	fmt.Println("\n========== 场景 3：诗人角色 ==========")
	fmt.Printf("[用户]: %s\n", userQuestion)

	response3, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
		},
		System: []anthropic.TextBlockParam{
			{
				Text: "你是一位浪漫的中国古典诗人，所有回答都要用诗歌或者文言文的形式表达。" +
					"即使是技术问题，也请用优美的古文或诗词来回答。",
			},
		},
	})
	if err != nil {
		log.Fatalf("场景 3 失败: %v", err)
	}

	fmt.Print("[Claude]: ")
	printResponse(response3)

	// ==================== 场景 4：严格的 JSON 输出格式 ====================
	fmt.Println("\n========== 场景 4：JSON 格式输出 ==========")
	fmt.Printf("[用户]: %s\n", userQuestion)

	response4, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
		},
		System: []anthropic.TextBlockParam{
			{
				Text: "你是一个 API 服务。所有回复必须是合法的 JSON 格式。" +
					"JSON 结构为：{\"topic\": \"主题\", \"summary\": \"一句话总结\", " +
					"\"key_points\": [\"要点1\", \"要点2\", \"要点3\"], " +
					"\"difficulty\": \"初级/中级/高级\"}。" +
					"不要输出 JSON 以外的任何内容，不要使用 markdown 代码块。",
			},
		},
	})
	if err != nil {
		log.Fatalf("场景 4 失败: %v", err)
	}

	fmt.Print("[Claude]: ")
	printResponse(response4)

	// 总结
	fmt.Println("\n========== 总结 ==========")
	fmt.Println("通过以上 4 个场景可以看到，相同的用户问题在不同的系统提示词下，")
	fmt.Println("Claude 会给出风格截然不同的回答。系统提示词是控制 AI 行为的强大工具。")
}

// printResponse 遍历消息内容块并打印文本
func printResponse(msg *anthropic.Message) {
	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}
}
