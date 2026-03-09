// 第三章示例 3：工具选择模式 —— auto / any / tool
// 运行方式: go run 03_tool_modes.go
//
// 本示例演示如何：
// 1. 使用 auto 模式让 Claude 自行决定是否调用工具
// 2. 使用 any 模式强制 Claude 必须调用某个工具
// 3. 使用 tool 模式强制 Claude 调用指定的工具
// 4. 对比三种模式在相同问题下的不同行为

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 定义两个工具
	// ========================================
	tools := []anthropic.ToolUnionParam{
		// 工具 1：获取当前时间
		{
			OfTool: &anthropic.ToolParam{
				Name:        "get_current_time",
				Description: anthropic.String("获取当前的日期和时间信息。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"timezone": map[string]any{
							"type":        "string",
							"description": "时区名称，如 Asia/Shanghai, America/New_York",
						},
					},
					Required: []string{"timezone"},
				},
			},
		},
		// 工具 2：翻译文本
		{
			OfTool: &anthropic.ToolParam{
				Name:        "translate_text",
				Description: anthropic.String("将文本翻译成指定语言。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"text": map[string]any{
							"type":        "string",
							"description": "要翻译的文本",
						},
						"target_language": map[string]any{
							"type":        "string",
							"description": "目标语言，如 english, chinese, japanese",
						},
					},
					Required: []string{"text", "target_language"},
				},
			},
		},
	}

	// ========================================
	// 模式 1：auto —— Claude 自行决定
	// ========================================
	// auto 是默认模式。Claude 会根据问题判断：
	// - 如果问题需要工具才能回答，就调用工具
	// - 如果 Claude 自己就能回答，就直接回复
	fmt.Println("=" + repeat("=", 50))
	fmt.Println("模式 1: auto（自动决定）")
	fmt.Println(repeat("=", 51))

	// 问题 A：不需要工具就能回答的问题
	fmt.Println("\n--- 问题 A（不需要工具）: 什么是Go语言？---")
	msgAuto1, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Tools:     tools,
		// auto 模式：Claude 自由选择是否使用工具
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfAuto: &anthropic.ToolChoiceAutoParam{},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("用一句话说明什么是Go语言？")),
		},
	})
	if err != nil {
		log.Fatalf("auto 模式请求失败: %v", err)
	}
	fmt.Printf("停止原因: %s\n", msgAuto1.StopReason)
	printResponse(msgAuto1)
	// 预期结果：stop_reason 为 "end_turn"，Claude 直接回答，不调用工具

	// 问题 B：需要工具才能回答的问题
	fmt.Println("\n--- 问题 B（需要工具）: 现在几点？---")
	msgAuto2, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Tools:     tools,
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfAuto: &anthropic.ToolChoiceAutoParam{},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("北京现在几点了？")),
		},
	})
	if err != nil {
		log.Fatalf("auto 模式请求失败: %v", err)
	}
	fmt.Printf("停止原因: %s\n", msgAuto2.StopReason)
	printResponse(msgAuto2)
	// 预期结果：stop_reason 为 "tool_use"，Claude 调用 get_current_time

	// ========================================
	// 模式 2：any —— 强制使用工具
	// ========================================
	// any 模式：Claude 必须调用至少一个工具，即使问题不需要工具
	// 使用场景：你确定需要 Claude 使用工具，不希望它跳过
	fmt.Println("\n" + repeat("=", 51))
	fmt.Println("模式 2: any（强制使用任意工具）")
	fmt.Println(repeat("=", 51))

	fmt.Println("\n--- 问一个通常不需要工具的问题 ---")
	msgAny, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Tools:     tools,
		// any 模式：必须调用至少一个工具
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfAny: &anthropic.ToolChoiceAnyParam{},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("帮我把'你好世界'翻译成英文")),
		},
	})
	if err != nil {
		log.Fatalf("any 模式请求失败: %v", err)
	}
	fmt.Printf("停止原因: %s\n", msgAny.StopReason)
	printResponse(msgAny)
	// 预期结果：stop_reason 为 "tool_use"，Claude 被迫调用一个工具
	// 它可能会选择 translate_text 工具

	// ========================================
	// 模式 3：tool —— 强制调用指定工具
	// ========================================
	// tool 模式：Claude 必须调用你指定的那个工具
	// 使用场景：你明确知道应该使用哪个工具，不想让 Claude 做选择
	fmt.Println("\n" + repeat("=", 51))
	fmt.Println("模式 3: tool（强制使用指定工具）")
	fmt.Println(repeat("=", 51))

	fmt.Println("\n--- 强制使用 get_current_time 工具 ---")
	msgTool, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Tools:     tools,
		// tool 模式：必须调用指定名称的工具
		ToolChoice: anthropic.ToolChoiceUnionParam{
			OfTool: &anthropic.ToolChoiceToolParam{
				Name: "get_current_time", // 指定必须调用的工具名称
			},
		},
		Messages: []anthropic.MessageParam{
			// 注意：即使问题与时间无关，Claude 也会被迫调用 get_current_time
			anthropic.NewUserMessage(anthropic.NewTextBlock("随便聊聊天气")),
		},
	})
	if err != nil {
		log.Fatalf("tool 模式请求失败: %v", err)
	}
	fmt.Printf("停止原因: %s\n", msgTool.StopReason)
	printResponse(msgTool)
	// 预期结果：stop_reason 为 "tool_use"，Claude 被迫调用 get_current_time
	// 它会尝试找一个合理的理由来调用这个工具

	// ========================================
	// 补充演示：处理 tool 模式的完整流程
	// ========================================
	fmt.Println("\n" + repeat("=", 51))
	fmt.Println("补充：处理 tool 模式的完整流程")
	fmt.Println(repeat("=", 51))

	// 当使用 tool 模式时，Claude 一定会返回 tool_use
	// 我们需要执行工具并返回结果
	if msgTool.StopReason == "tool_use" {
		messages := []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("随便聊聊天气")),
			msgTool.ToParam(),
		}

		// 查找工具调用并执行
		for _, block := range msgTool.Content {
			if tu, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				// 模拟获取当前时间
				currentTime := time.Now().Format("2006-01-02 15:04:05 MST")
				result := fmt.Sprintf(`{"current_time": "%s"}`, currentTime)
				fmt.Printf("执行工具 %s，返回: %s\n", tu.Name, result)

				messages = append(messages,
					anthropic.NewUserMessage(anthropic.NewToolResultBlock(tu.ID, result, false)),
				)
			}
		}

		// 将结果发回给 Claude，注意这次不再强制使用工具
		// 让 Claude 自由回复（auto 模式或不设置 ToolChoice）
		finalMsg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 256,
			Tools:     tools,
			Messages:  messages,
			// 返回结果时通常切回 auto 模式，让 Claude 自由回复
			ToolChoice: anthropic.ToolChoiceUnionParam{
				OfAuto: &anthropic.ToolChoiceAutoParam{},
			},
		})
		if err != nil {
			log.Fatalf("最终回复失败: %v", err)
		}

		fmt.Println("\nClaude 的最终回复:")
		for _, block := range finalMsg.Content {
			if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Println(tb.Text)
			}
		}
	}

	// ========================================
	// 总结
	// ========================================
	fmt.Println("\n" + repeat("=", 51))
	fmt.Println("三种模式总结")
	fmt.Println(repeat("=", 51))
	fmt.Println("auto:  Claude 自行判断是否需要工具（默认，最常用）")
	fmt.Println("any:   Claude 必须调用某个工具（强制使用工具）")
	fmt.Println("tool:  Claude 必须调用指定的工具（精确控制）")
}

// printResponse 打印 Claude 响应中的所有内容块
func printResponse(msg *anthropic.Message) {
	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Printf("  [文本] %s\n", v.Text)
		case anthropic.ToolUseBlock:
			rawInput := json.RawMessage(v.Input)
			fmt.Printf("  [工具调用] %s, 参数: %s\n", v.Name, string(rawInput))
		}
	}
}

// repeat 重复字符串 n 次（辅助函数）
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
