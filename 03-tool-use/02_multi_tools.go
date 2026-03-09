// 第三章示例 2：多工具协作 —— 天气查询 + 温度转换
// 运行方式: go run 02_multi_tools.go
//
// 本示例演示如何：
// 1. 同时定义多个工具供 Claude 使用
// 2. Claude 如何根据问题选择合适的工具
// 3. 处理 Claude 在一次响应中调用多个工具的情况
// 4. 实现完整的多工具调用循环

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ========================================
// 模拟的天气数据（实际应用中会调用真实 API）
// ========================================
var weatherData = map[string]map[string]any{
	"北京": {
		"temperature_celsius": 22.0,
		"humidity":            45,
		"condition":           "晴天",
		"wind":                "北风3级",
	},
	"上海": {
		"temperature_celsius": 28.0,
		"humidity":            72,
		"condition":           "多云",
		"wind":                "东南风2级",
	},
	"广州": {
		"temperature_celsius": 33.0,
		"humidity":            85,
		"condition":           "雷阵雨",
		"wind":                "南风4级",
	},
	"哈尔滨": {
		"temperature_celsius": -5.0,
		"humidity":            30,
		"condition":           "小雪",
		"wind":                "西北风5级",
	},
}

// executeWeatherTool 执行天气查询工具
// 根据城市名返回模拟的天气数据
func executeWeatherTool(input json.RawMessage) (string, bool) {
	var params struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err), true
	}

	// 在模拟数据中查找城市
	data, exists := weatherData[params.City]
	if !exists {
		// 尝试模糊匹配（去掉"市"后缀）
		cityName := strings.TrimSuffix(params.City, "市")
		data, exists = weatherData[cityName]
		if !exists {
			return fmt.Sprintf("未找到城市 %s 的天气数据。支持的城市：北京、上海、广州、哈尔滨", params.City), true
		}
	}

	// 将天气数据序列化为 JSON 字符串返回
	result, _ := json.Marshal(data)
	return string(result), false
}

// executeTemperatureConvertTool 执行温度转换工具
// 在摄氏度和华氏度之间转换
func executeTemperatureConvertTool(input json.RawMessage) (string, bool) {
	var params struct {
		Temperature float64 `json:"temperature"`
		FromUnit    string  `json:"from_unit"`
		ToUnit      string  `json:"to_unit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("参数解析失败: %v", err), true
	}

	var result float64
	switch {
	case params.FromUnit == "celsius" && params.ToUnit == "fahrenheit":
		// 摄氏度转华氏度：F = C * 9/5 + 32
		result = params.Temperature*9.0/5.0 + 32
	case params.FromUnit == "fahrenheit" && params.ToUnit == "celsius":
		// 华氏度转摄氏度：C = (F - 32) * 5/9
		result = (params.Temperature - 32) * 5.0 / 9.0
	case params.FromUnit == params.ToUnit:
		result = params.Temperature
	default:
		return fmt.Sprintf("不支持的转换：%s -> %s", params.FromUnit, params.ToUnit), true
	}

	return fmt.Sprintf("%.2f %s = %.2f %s", params.Temperature, params.FromUnit, result, params.ToUnit), false
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 定义多个工具
	// ========================================
	// 同时定义天气查询工具和温度转换工具
	// Claude 会根据用户的问题智能选择使用哪些工具
	tools := []anthropic.ToolUnionParam{
		// 工具 1：天气查询
		{
			OfTool: &anthropic.ToolParam{
				Name:        "get_weather",
				Description: anthropic.String("查询指定城市的当前天气信息，包括温度（摄氏度）、湿度、天气状况和风力。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"city": map[string]any{
							"type":        "string",
							"description": "要查询天气的城市名称，如：北京、上海、广州",
						},
					},
					Required: []string{"city"},
				},
			},
		},
		// 工具 2：温度转换
		{
			OfTool: &anthropic.ToolParam{
				Name:        "convert_temperature",
				Description: anthropic.String("在摄氏度和华氏度之间进行温度转换。"),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: map[string]any{
						"temperature": map[string]any{
							"type":        "number",
							"description": "要转换的温度值",
						},
						"from_unit": map[string]any{
							"type":        "string",
							"description": "原始温度单位",
							"enum":        []string{"celsius", "fahrenheit"},
						},
						"to_unit": map[string]any{
							"type":        "string",
							"description": "目标温度单位",
							"enum":        []string{"celsius", "fahrenheit"},
						},
					},
					Required: []string{"temperature", "from_unit", "to_unit"},
				},
			},
		},
	}

	// ========================================
	// 发送一个需要多个工具配合的问题
	// ========================================
	// 这个问题需要 Claude：
	// 1. 先调用天气工具获取温度（摄氏度）
	// 2. 再调用转换工具转换为华氏度
	userQuestion := "北京和广州现在的天气怎么样？请把温度都转换成华氏度告诉我。"
	fmt.Printf("用户: %s\n\n", userQuestion)

	// 初始消息列表
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userQuestion)),
	}

	// ========================================
	// Agentic 循环：持续处理直到 Claude 完成
	// ========================================
	// 因为 Claude 可能需要多轮工具调用，我们用循环处理
	round := 0
	for {
		round++
		fmt.Printf("--- 第 %d 轮对话 ---\n", round)

		// 发送请求
		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 1024,
			Tools:     tools,
			Messages:  messages,
		})
		if err != nil {
			log.Fatalf("第 %d 轮 API 调用失败: %v", round, err)
		}

		fmt.Printf("停止原因: %s\n", message.StopReason)

		// 如果 Claude 不需要调用工具，输出最终回复并退出循环
		if message.StopReason == "end_turn" {
			fmt.Println("\n=== Claude 的最终回复 ===")
			for _, block := range message.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					fmt.Println(tb.Text)
				}
			}
			break
		}

		// Claude 请求调用工具
		// 将 Claude 的响应加入消息历史
		messages = append(messages, message.ToParam())

		// 收集所有工具调用的结果
		// 注意：Claude 可能在一次响应中调用多个工具！
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				fmt.Printf("Claude: %s\n", v.Text)
			case anthropic.ToolUseBlock:
				fmt.Printf("\n调用工具: %s\n", v.Name)
				rawInput := json.RawMessage(v.Input)
				fmt.Printf("参数: %s\n", string(rawInput))

				// 根据工具名称分发到对应的处理函数
				var resultStr string
				var isError bool
				switch v.Name {
				case "get_weather":
					resultStr, isError = executeWeatherTool(rawInput)
				case "convert_temperature":
					resultStr, isError = executeTemperatureConvertTool(rawInput)
				default:
					resultStr = fmt.Sprintf("未知工具: %s", v.Name)
					isError = true
				}

				fmt.Printf("结果: %s\n", resultStr)

				// 为每个工具调用创建对应的结果
				// 每个 tool_result 必须通过 toolUseID 与对应的 tool_use 关联
				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(v.ID, resultStr, isError),
				)
			}
		}

		// 将所有工具结果作为一条 user 消息发回
		// 多个 tool_result 可以在同一条消息中
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	fmt.Printf("\n总共进行了 %d 轮对话\n", round)
}
