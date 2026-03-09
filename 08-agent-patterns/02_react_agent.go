// 运行方式: go run 02_react_agent.go
// ReAct 模式 Agent：通过显式的 思考→行动→观察 循环来解决问题

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// =====================
// 知识库模拟（Agent 的工具数据源）
// =====================

// 模拟一个简单的知识库
var knowledgeBase = map[string]string{
	"golang":  "Go（又称 Golang）是 Google 开发的开源编程语言，由 Robert Griesemer、Rob Pike 和 Ken Thompson 于 2007 年设计，2009 年发布。特点是编译速度快、内置并发支持、垃圾回收。",
	"rust":    "Rust 是由 Mozilla 研究院开发的系统编程语言，注重安全性、并发性和性能。通过所有权系统在编译时保证内存安全，无需垃圾回收。",
	"python":  "Python 是一种高级通用编程语言，由 Guido van Rossum 于 1991 年首次发布。以其简洁的语法和丰富的库生态系统著称。",
	"claude":  "Claude 是 Anthropic 开发的 AI 助手，基于大语言模型技术。支持文本对话、代码生成、图像理解等能力。最新版本为 Claude 4.6 系列。",
	"mcp":     "MCP（Model Context Protocol）是 Anthropic 提出的开放协议，用于标准化 AI 模型与外部工具和数据源之间的连接方式。",
}

// 定义 ReAct Agent 的工具
var reactTools = []anthropic.ToolUnionParam{
	{
		OfTool: &anthropic.ToolParam{
			Name:        "search_knowledge",
			Description: anthropic.String("在知识库中搜索关键词，返回相关信息"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"keyword": map[string]any{
						"type":        "string",
						"description": "搜索关键词（小写）",
					},
				},
				Required: []string{"keyword"},
			},
		},
	},
	{
		OfTool: &anthropic.ToolParam{
			Name:        "compare",
			Description: anthropic.String("比较两个主题的信息"),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"topic1": map[string]any{
						"type":        "string",
						"description": "第一个主题",
					},
					"topic2": map[string]any{
						"type":        "string",
						"description": "第二个主题",
					},
				},
				Required: []string{"topic1", "topic2"},
			},
		},
	},
}

// 执行工具调用
func executeReActTool(name string, inputJSON string) string {
	switch name {
	case "search_knowledge":
		var input struct {
			Keyword string `json:"keyword"`
		}
		json.Unmarshal([]byte(inputJSON), &input)

		keyword := strings.ToLower(input.Keyword)
		if info, ok := knowledgeBase[keyword]; ok {
			return info
		}
		// 模糊搜索
		for key, value := range knowledgeBase {
			if strings.Contains(key, keyword) || strings.Contains(keyword, key) {
				return value
			}
		}
		return fmt.Sprintf("未找到关于 '%s' 的信息", input.Keyword)

	case "compare":
		var input struct {
			Topic1 string `json:"topic1"`
			Topic2 string `json:"topic2"`
		}
		json.Unmarshal([]byte(inputJSON), &input)

		info1, ok1 := knowledgeBase[strings.ToLower(input.Topic1)]
		info2, ok2 := knowledgeBase[strings.ToLower(input.Topic2)]

		result := fmt.Sprintf("【%s】", input.Topic1)
		if ok1 {
			result += info1
		} else {
			result += "无相关信息"
		}
		result += fmt.Sprintf("\n【%s】", input.Topic2)
		if ok2 {
			result += info2
		} else {
			result += "无相关信息"
		}
		return result

	default:
		return "未知工具"
	}
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ReAct 模式的系统提示
	// 关键：要求 Agent 在每次行动前明确表达推理过程
	systemPrompt := `你是一个使用 ReAct（Reasoning + Acting）模式的智能助手。

在回答问题时，请遵循以下模式：

1. **思考 (Thought)**: 先分析问题，思考需要什么信息
2. **行动 (Action)**: 使用工具获取信息
3. **观察 (Observation)**: 分析工具返回的结果
4. **重复**: 如果信息不够，继续思考和行动
5. **回答**: 当信息充分时，给出最终答案

请在回复中明确标注每个步骤（思考、观察等），这样用户可以看到你的推理过程。`

	// 用户问题：需要多步推理才能回答
	question := "请比较 Go 和 Rust 这两种编程语言，并说明它们各自的特点。另外，Claude 和 MCP 有什么关系？"

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
	}

	fmt.Printf("❓ 问题: %s\n", question)
	fmt.Println(strings.Repeat("=", 60))

	// ReAct Agent 循环
	maxTurns := 10
	for turn := 1; turn <= maxTurns; turn++ {
		fmt.Printf("\n--- 第 %d 轮 ---\n", turn)

		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 2048,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Messages:  messages,
			Tools:     reactTools,
		})
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			return
		}

		messages = append(messages, message.ToParam())

		// 处理响应
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			switch v := block.AsAny().(type) {
			case anthropic.TextBlock:
				fmt.Printf("\n💭 %s\n", v.Text)
			case anthropic.ToolUseBlock:
				fmt.Printf("\n🔧 行动 → 调用 %s(%s)\n", v.Name, v.JSON.Input.Raw())

				result := executeReActTool(v.Name, v.JSON.Input.Raw())
				fmt.Printf("👁  观察 → %s\n", result)

				toolResults = append(toolResults,
					anthropic.NewToolResultBlock(v.ID, result, false),
				)
			}
		}

		// 任务完成
		if message.StopReason == "end_turn" {
			fmt.Println("\n✅ ReAct Agent 完成推理!")
			break
		}

		// 返回工具结果
		if len(toolResults) > 0 {
			messages = append(messages, anthropic.NewUserMessage(toolResults...))
		}
	}
}
