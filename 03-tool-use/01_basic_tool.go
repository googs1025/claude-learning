// 第三章示例 1：基础工具定义与使用 —— 计算器工具
// 运行方式: go run 01_basic_tool.go
//
// 本示例演示如何：
// 1. 定义一个计算器工具，支持加减乘除
// 2. 发送消息让 Claude 决定是否调用工具
// 3. 解析 Claude 返回的工具调用请求
// 4. 在本地执行计算并将结果返回给 Claude
// 5. 获取 Claude 的最终自然语言回复

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// ========================================
	// 第一步：创建客户端
	// ========================================
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 第二步：定义工具
	// ========================================
	// 工具定义包含三个核心要素：
	// - Name: 工具名称，Claude 会用这个名称来调用工具
	// - Description: 工具描述，帮助 Claude 理解何时应该使用这个工具
	// - InputSchema: 参数的 JSON Schema，定义工具接受哪些参数
	tools := []anthropic.ToolUnionParam{
		{
			OfTool: &anthropic.ToolParam{
				Name:        "calculator",
				Description: anthropic.String("一个简单的计算器，支持两个数字的加减乘除运算。当用户需要进行数学计算时使用此工具。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					// Properties 定义每个参数的类型和描述
					Properties: map[string]any{
						"operation": map[string]any{
							"type":        "string",
							"description": "运算类型：add（加法）、subtract（减法）、multiply（乘法）、divide（除法）",
							"enum":        []string{"add", "subtract", "multiply", "divide"},
						},
						"a": map[string]any{
							"type":        "number",
							"description": "第一个操作数",
						},
						"b": map[string]any{
							"type":        "number",
							"description": "第二个操作数",
						},
					},
					// Required 指定哪些参数是必填的
					Required: []string{"operation", "a", "b"},
				},
			},
		},
	}

	// ========================================
	// 第三步：发送请求，附带工具定义
	// ========================================
	fmt.Println("=== 向 Claude 提问一个数学问题 ===")
	userQuestion := "请帮我计算 1234 乘以 5678 等于多少？"
	fmt.Printf("用户: %s\n\n", userQuestion)

	// 将工具定义传入请求参数的 Tools 字段
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Tools:     tools, // 告诉 Claude 可以使用哪些工具
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// ========================================
	// 第四步：检查 Claude 是否要调用工具
	// ========================================
	// StopReason 有两种关键值：
	// - "end_turn": Claude 直接给出了回复，不需要工具
	// - "tool_use": Claude 需要调用工具来回答问题
	fmt.Printf("停止原因: %s\n", message.StopReason)

	if message.StopReason != "tool_use" {
		// 如果 Claude 没有调用工具，直接输出回复
		fmt.Println("Claude 没有使用工具，直接回复了：")
		for _, block := range message.Content {
			if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Println(tb.Text)
			}
		}
		return
	}

	// ========================================
	// 第五步：解析工具调用请求
	// ========================================
	fmt.Println("\n=== Claude 请求调用工具 ===")

	var toolUseID string    // 工具调用的唯一标识，返回结果时需要用到
	var toolName string     // Claude 调用的工具名称
	var toolInput json.RawMessage // 工具的输入参数（原始 JSON）

	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			// Claude 可能在调用工具前先输出一些文本（思考过程）
			fmt.Printf("Claude 的思考: %s\n", v.Text)
		case anthropic.ToolUseBlock:
			// 这是工具调用请求
			toolUseID = v.ID
			toolName = v.Name
			// 使用 v.Input 获取原始 JSON（类型为 json.RawMessage）
			toolInput = v.Input
			fmt.Printf("工具名称: %s\n", toolName)
			fmt.Printf("工具调用ID: %s\n", toolUseID)
			fmt.Printf("输入参数: %s\n", string(toolInput))
		}
	}

	// ========================================
	// 第六步：在本地执行工具
	// ========================================
	fmt.Println("\n=== 执行工具 ===")

	// 解析 JSON 参数
	var params struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}
	if err := json.Unmarshal(toolInput, &params); err != nil {
		log.Fatalf("解析工具参数失败: %v", err)
	}

	// 执行计算
	var result float64
	var calcErr error
	switch params.Operation {
	case "add":
		result = params.A + params.B
		fmt.Printf("计算: %.0f + %.0f = %.0f\n", params.A, params.B, result)
	case "subtract":
		result = params.A - params.B
		fmt.Printf("计算: %.0f - %.0f = %.0f\n", params.A, params.B, result)
	case "multiply":
		result = params.A * params.B
		fmt.Printf("计算: %.0f * %.0f = %.0f\n", params.A, params.B, result)
	case "divide":
		if params.B == 0 {
			calcErr = fmt.Errorf("除数不能为零")
		} else {
			result = params.A / params.B
			fmt.Printf("计算: %.0f / %.0f = %.4f\n", params.A, params.B, result)
		}
	default:
		calcErr = fmt.Errorf("未知运算: %s", params.Operation)
	}

	// 构造结果字符串
	var resultStr string
	var isError bool
	if calcErr != nil {
		resultStr = fmt.Sprintf("错误: %s", calcErr.Error())
		isError = true
		fmt.Printf("执行出错: %s\n", resultStr)
	} else {
		resultStr = fmt.Sprintf("%.6f", result)
		isError = false
		fmt.Printf("计算结果: %s\n", resultStr)
	}

	// ========================================
	// 第七步：将工具结果发回给 Claude
	// ========================================
	fmt.Println("\n=== 将结果返回给 Claude ===")

	// 构造完整的对话历史，包含：
	// 1. 用户的原始问题
	// 2. Claude 的工具调用请求（assistant 消息）
	// 3. 工具执行结果（user 消息中的 tool_result）
	finalMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Tools:     tools, // 仍然需要传入工具定义
		Messages: []anthropic.MessageParam{
			// 第一轮：用户的原始问题
			anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
			// 第二轮：Claude 的工具调用（原样传回 Claude 的响应）
			// 使用 message.ToParam() 将响应转换为消息参数
			message.ToParam(),
			// 第三轮：工具执行结果
			// NewToolResultBlock 创建工具结果内容块
			// 参数：工具调用ID、结果字符串、是否为错误
			anthropic.NewUserMessage(
				anthropic.NewToolResultBlock(toolUseID, resultStr, isError),
			),
		},
	})
	if err != nil {
		log.Fatalf("返回工具结果失败: %v", err)
	}

	// ========================================
	// 第八步：输出 Claude 的最终回复
	// ========================================
	fmt.Println("\n=== Claude 的最终回复 ===")
	for _, block := range finalMessage.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			fmt.Println(tb.Text)
		}
	}

	// 打印使用统计
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("最终回复 - 输入 token: %d, 输出 token: %d\n",
		finalMessage.Usage.InputTokens, finalMessage.Usage.OutputTokens)
}
