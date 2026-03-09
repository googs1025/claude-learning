// 第五章示例 4：扩展思考 + 工具调用
// 运行方式: go run 04_thinking_with_tools.go
//
// 本示例演示如何：
// 1. 同时启用扩展思考和工具调用
// 2. 观察 Claude 在决定使用工具前的推理过程
// 3. 处理包含思考块和工具调用块的混合响应
// 4. 完成工具调用循环（发送工具结果，获取最终回答）

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/anthropics/anthropic-sdk-go"
)

// 模拟的工具函数：计算复利
func calculateCompoundInterest(principal, rate float64, years int, compoundsPerYear int) map[string]interface{} {
	// 复利公式: A = P * (1 + r/n)^(n*t)
	r := rate / 100.0
	n := float64(compoundsPerYear)
	t := float64(years)

	amount := principal * math.Pow(1+r/n, n*t)
	interest := amount - principal

	return map[string]interface{}{
		"principal":         principal,
		"rate":              rate,
		"years":             years,
		"compounds_per_year": compoundsPerYear,
		"final_amount":      math.Round(amount*100) / 100,
		"total_interest":    math.Round(interest*100) / 100,
	}
}

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	fmt.Println("=== 扩展思考 + 工具调用 ===")
	fmt.Println()

	// 定义复利计算工具
	compoundInterestTool := anthropic.ToolParam{
		Name:        "calculate_compound_interest",
		Description: anthropic.String("计算复利。根据本金、年利率、年数和每年复利次数，计算最终金额和总利息。"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"principal": map[string]interface{}{
					"type":        "number",
					"description": "本金（元）",
				},
				"annual_rate": map[string]interface{}{
					"type":        "number",
					"description": "年利率（百分比，例如 5 表示 5%）",
				},
				"years": map[string]interface{}{
					"type":        "integer",
					"description": "投资年数",
				},
				"compounds_per_year": map[string]interface{}{
					"type":        "integer",
					"description": "每年复利次数（1=年复利, 4=季度复利, 12=月复利, 365=日复利）",
				},
			},
			Required: []string{"principal", "annual_rate", "years", "compounds_per_year"},
		},
	}

	// 一个需要思考和计算的金融问题
	question := `我有 10 万元想做定期存款。银行提供两种方案：
方案 A：年利率 4.5%，按年复利，存 5 年
方案 B：年利率 4.3%，按月复利，存 5 年

请帮我计算两种方案各能获得多少利息，哪种方案更划算？`

	fmt.Printf("问题: %s\n\n", question)

	// 第一轮请求：发送问题，启用思考和工具
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
	}

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 16000,
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 10000,
			},
		},
		Tools: []anthropic.ToolUnionParam{{OfTool: &compoundInterestTool}},
		Messages: messages,
	})
	if err != nil {
		log.Fatalf("第一轮 API 调用失败: %v", err)
	}

	// 处理第一轮响应
	// 由于启用了思考，响应中会包含思考块
	fmt.Println("=== 第一轮响应 ===")

	// 收集工具调用结果，用于构建后续消息
	var toolResults []anthropic.ContentBlockParamUnion

	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.ThinkingBlock:
			fmt.Println("[思考过程]")
			// 只显示前 300 字符的思考内容
			preview := v.Thinking
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			fmt.Println(preview)
			fmt.Println()

		case anthropic.TextBlock:
			fmt.Printf("[文本] %s\n\n", v.Text)

		case anthropic.ToolUseBlock:
			fmt.Printf("[工具调用] %s (ID: %s)\n", v.Name, v.ID)

			// 解析工具输入参数
			var params map[string]interface{}
			if err := json.Unmarshal(v.Input, &params); err != nil {
				log.Printf("解析工具参数失败: %v", err)
				continue
			}
			fmt.Printf("  参数: %v\n", params)

			// 执行工具函数
			principal := params["principal"].(float64)
			annualRate := params["annual_rate"].(float64)
			years := int(params["years"].(float64))
			compoundsPerYear := int(params["compounds_per_year"].(float64))

			result := calculateCompoundInterest(principal, annualRate, years, compoundsPerYear)
			fmt.Printf("  结果: %v\n\n", result)

			// 将工具结果序列化为 JSON
			resultJSON, _ := json.Marshal(result)

			// 构建工具结果内容块
			toolResults = append(toolResults, anthropic.NewToolResultBlock(
				v.ID,
				string(resultJSON),
				false,
			))
		}
	}

	// 如果 Claude 调用了工具，需要发送工具结果并获取最终回答
	if message.StopReason == "tool_use" && len(toolResults) > 0 {
		fmt.Println("=== 发送工具结果，获取最终回答 ===\n")

		// 构建完整的对话历史
		// 注意：必须包含之前的所有消息，形成完整的对话链
		messages = append(messages, message.ToParam())
		messages = append(messages, anthropic.NewUserMessage(toolResults...))

		// 第二轮请求：发送工具结果
		finalMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:    anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 16000,
			Thinking: anthropic.ThinkingConfigParamUnion{
				OfEnabled: &anthropic.ThinkingConfigEnabledParam{
					BudgetTokens: 10000,
				},
			},
			Tools:    []anthropic.ToolUnionParam{{OfTool: &compoundInterestTool}},
			Messages: messages,
		})
		if err != nil {
			log.Fatalf("第二轮 API 调用失败: %v", err)
		}

		// 输出最终回答
		for _, block := range finalMessage.Content {
			switch v := block.AsAny().(type) {
			case anthropic.ThinkingBlock:
				fmt.Println("[第二轮思考]")
				preview := v.Thinking
				if len(preview) > 300 {
					preview = preview[:300] + "..."
				}
				fmt.Println(preview)
				fmt.Println()
			case anthropic.TextBlock:
				fmt.Println("[最终回答]")
				fmt.Println(v.Text)
			}
		}

		fmt.Println("\n=== 使用统计（第二轮）===")
		fmt.Printf("输入 token 数: %d\n", finalMessage.Usage.InputTokens)
		fmt.Printf("输出 token 数: %d\n", finalMessage.Usage.OutputTokens)
	}
}
