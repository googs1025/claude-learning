// 第五章示例 1：基础扩展思考
// 运行方式: go run 01_basic_thinking.go
//
// 本示例演示如何：
// 1. 启用 Claude 的扩展思考功能
// 2. 让 Claude 解决复杂的数学推理问题
// 3. 分别展示思考过程和最终答案

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	fmt.Println("=== 扩展思考：解决复杂数学问题 ===")
	fmt.Println()

	// 一个需要深度推理的数学问题
	mathProblem := `请解决以下问题：

一个农场里有鸡和兔子，一共有 35 个头和 94 条腿。
请问鸡和兔子各有多少只？

请给出完整的推理过程和最终答案。`

	fmt.Printf("问题: %s\n\n", mathProblem)

	// 发送请求，启用扩展思考
	// 关键配置：
	// - Thinking: 启用扩展思考功能
	// - BudgetTokens: 设置思考预算（最少 1024，必须小于 MaxTokens）
	// - MaxTokens: 需要足够大以容纳思考 + 回答
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 16000, // 需要足够大以容纳思考过程和最终回答
		// 启用扩展思考
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 10000, // 思考预算：最多使用 10000 个 token 进行推理
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(mathProblem)),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 解析响应
	// 启用扩展思考后，响应中会包含两种类型的内容块：
	// 1. ThinkingBlock：Claude 的内部推理过程
	// 2. TextBlock：最终给用户的回答
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.ThinkingBlock:
			// 思考块：展示 Claude 的推理过程
			fmt.Println("=== 思考过程 ===")
			fmt.Println(v.Thinking)
			fmt.Println()
		case anthropic.TextBlock:
			// 文本块：最终回答
			fmt.Println("=== 最终回答 ===")
			fmt.Println(v.Text)
		}
	}

	// 输出使用统计
	// 注意：思考 token 会单独计算
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
}
