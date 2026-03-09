// 运行方式: go run 01_simple_agent.go
// 简单 Agent 示例：一个能够使用工具完成任务的基础 Agent

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// =====================
// 工具定义
// =====================

// 定义 Agent 可以使用的工具集
var agentTools = []anthropic.ToolUnionParam{
	{
		OfTool: &anthropic.ToolParam{
			Name:        "calculate",
			Description: anthropic.String("执行数学计算，支持基本运算和常用数学函数"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"expression": map[string]any{
						"type":        "string",
						"description": "数学表达式，如 '2+3' 或 'sqrt(16)'",
					},
				},
				Required: []string{"expression"},
			},
		},
	},
	{
		OfTool: &anthropic.ToolParam{
			Name:        "string_tool",
			Description: anthropic.String("字符串处理工具，支持大小写转换、长度计算、反转等操作"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"text": map[string]any{
						"type":        "string",
						"description": "要处理的文本",
					},
					"operation": map[string]any{
						"type":        "string",
						"description": "操作类型: uppercase, lowercase, length, reverse, word_count",
						"enum":        []string{"uppercase", "lowercase", "length", "reverse", "word_count"},
					},
				},
				Required: []string{"text", "operation"},
			},
		},
	},
}

// =====================
// 工具执行函数
// =====================

// executeTool 根据工具名称和参数执行对应的工具
func executeTool(name string, inputJSON string) (string, error) {
	switch name {
	case "calculate":
		return executeCalculate(inputJSON)
	case "string_tool":
		return executeStringTool(inputJSON)
	default:
		return "", fmt.Errorf("未知工具: %s", name)
	}
}

func executeCalculate(inputJSON string) (string, error) {
	var input struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return "", err
	}

	// 简单的数学计算演示
	result := "计算结果: "
	expr := input.Expression

	switch {
	case strings.HasPrefix(expr, "sqrt("):
		// 简单解析 sqrt
		var num float64
		fmt.Sscanf(expr, "sqrt(%f)", &num)
		result += fmt.Sprintf("%.2f", math.Sqrt(num))
	case strings.Contains(expr, "+"):
		var a, b float64
		fmt.Sscanf(expr, "%f+%f", &a, &b)
		result += fmt.Sprintf("%.2f", a+b)
	case strings.Contains(expr, "*"):
		var a, b float64
		fmt.Sscanf(expr, "%f*%f", &a, &b)
		result += fmt.Sprintf("%.2f", a*b)
	case strings.Contains(expr, "-"):
		var a, b float64
		fmt.Sscanf(expr, "%f-%f", &a, &b)
		result += fmt.Sprintf("%.2f", a-b)
	case strings.Contains(expr, "/"):
		var a, b float64
		fmt.Sscanf(expr, "%f/%f", &a, &b)
		if b == 0 {
			return "错误: 除以零", nil
		}
		result += fmt.Sprintf("%.2f", a/b)
	default:
		result += "无法解析表达式: " + expr
	}

	return result, nil
}

func executeStringTool(inputJSON string) (string, error) {
	var input struct {
		Text      string `json:"text"`
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return "", err
	}

	switch input.Operation {
	case "uppercase":
		return strings.ToUpper(input.Text), nil
	case "lowercase":
		return strings.ToLower(input.Text), nil
	case "length":
		return fmt.Sprintf("长度: %d 个字符", len([]rune(input.Text))), nil
	case "reverse":
		runes := []rune(input.Text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	case "word_count":
		words := strings.Fields(input.Text)
		return fmt.Sprintf("单词数: %d", len(words)), nil
	default:
		return "", fmt.Errorf("未知操作: %s", input.Operation)
	}
}

// =====================
// Agent 核心循环
// =====================

// runAgent 执行 Agent 的核心循环
// 1. 发送消息给 Claude
// 2. 如果 Claude 需要调用工具，执行工具并返回结果
// 3. 重复直到 Claude 给出最终回答
func runAgent(task string) {
	client := anthropic.NewClient()
	ctx := context.Background()

	// 初始化消息历史
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(task)),
	}

	// 系统提示：定义 Agent 的行为
	system := []anthropic.TextBlockParam{
		{Text: "你是一个helpful的助手。你可以使用工具来帮助用户完成任务。请先分析任务需要哪些步骤，然后逐步使用工具完成。"},
	}

	fmt.Printf("📋 任务: %s\n", task)
	fmt.Println(strings.Repeat("=", 50))

	// Agent 循环（设置最大轮次防止无限循环）
	maxTurns := 10
	for turn := 1; turn <= maxTurns; turn++ {
		fmt.Printf("\n🔄 第 %d 轮\n", turn)

		// 调用 Claude API
		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 1024,
			System:    system,
			Messages:  messages,
			Tools:     agentTools,
		})
		if err != nil {
			fmt.Printf("❌ API 错误: %v\n", err)
			return
		}

		// 将 Claude 的响应加入消息历史
		messages = append(messages, message.ToParam())

		// 处理响应中的每个内容块
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				fmt.Printf("💬 Claude: %s\n", v.Text)
			case anthropic.ToolUseBlock:
				fmt.Printf("🔧 调用工具: %s\n", v.Name)
				fmt.Printf("   参数: %s\n", v.JSON.Input.Raw())

				// 执行工具
				result, err := executeTool(v.Name, v.JSON.Input.Raw())
				if err != nil {
					result = fmt.Sprintf("工具执行错误: %v", err)
					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(v.ID, result, true),
					)
				} else {
					fmt.Printf("   结果: %s\n", result)
					toolResults = append(toolResults,
						anthropic.NewToolResultBlock(v.ID, result, false),
					)
				}
			}
		}

		// 如果没有工具调用，说明 Agent 已完成任务
		if message.StopReason == "end_turn" {
			fmt.Println("\n✅ Agent 任务完成!")
			break
		}

		// 将工具结果发送回 Claude
		if len(toolResults) > 0 {
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}
	}
}

func main() {
	// 给 Agent 一个需要多步骤完成的任务
	runAgent("请帮我完成以下任务：1) 计算 25 * 4 + 50 的结果 2) 将 'Hello World' 转成大写 3) 计算 'Go 语言很有趣' 这句话的字符长度")
}
