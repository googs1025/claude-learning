// 第四章示例 4：视觉 + 工具调用结合
// 运行方式: go run 04_image_with_tools.go
//
// 本示例演示如何：
// 1. 定义一个图片分类工具
// 2. 同时发送图片和文本提示给 Claude
// 3. 让 Claude 通过工具调用返回结构化的分类结果
// 4. 展示视觉能力与工具调用的协同使用

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	// 定义图片分类工具
	// 这个工具让 Claude 以结构化的方式返回图片分类结果
	classifyTool := anthropic.ToolParam{
		Name:        "classify_image",
		Description: anthropic.String("对图片进行分类，返回类别、置信度和描述信息"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"description": "图片的主要类别",
					"enum":        []string{"动物", "风景", "人物", "食物", "建筑", "交通工具", "其他"},
				},
				"sub_category": map[string]interface{}{
					"type":        "string",
					"description": "更具体的子类别，例如：猫、狗、山、河等",
				},
				"confidence": map[string]interface{}{
					"type":        "number",
					"description": "分类的置信度，范围 0.0 到 1.0",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "对图片内容的简要描述",
				},
				"colors": map[string]interface{}{
					"type":        "array",
					"description": "图片中的主要颜色列表",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			Required: []string{"category", "sub_category", "confidence", "description", "colors"},
		},
	}

	// 使用一张猫咪图片
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Cat03.jpg/1200px-Cat03.jpg"

	fmt.Println("=== 视觉 + 工具调用 ===")
	fmt.Printf("图片 URL: %s\n\n", imageURL)

	// 发送请求：图片 + 文本 + 工具定义
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Tools:    []anthropic.ToolUnionParam{{OfTool: &classifyTool}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				// 图片内容块
				anthropic.NewImageBlock(anthropic.URLImageSourceParam{
					URL: imageURL,
				}),
				// 文本提示：要求使用分类工具
				anthropic.NewTextBlock("请使用 classify_image 工具对这张图片进行分类分析。"),
			),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 解析响应，提取工具调用结果
	fmt.Println("=== 分类结果 ===")
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			// Claude 可能在调用工具前输出一些文本说明
			fmt.Printf("[文本] %s\n", v.Text)

		case anthropic.ToolUseBlock:
			// 工具调用块包含工具名称和输入参数
			fmt.Printf("[工具调用] 工具名: %s\n", v.Name)
			fmt.Printf("[工具调用] 工具 ID: %s\n\n", v.ID)

			// 将工具输入参数解析为结构化数据
			var result map[string]interface{}
			if err := json.Unmarshal(v.Input, &result); err != nil {
				log.Printf("解析工具输入失败: %v", err)
				continue
			}

			// 格式化输出分类结果
			fmt.Println("--- 分类详情 ---")
			if category, ok := result["category"]; ok {
				fmt.Printf("  类别:     %v\n", category)
			}
			if subCategory, ok := result["sub_category"]; ok {
				fmt.Printf("  子类别:   %v\n", subCategory)
			}
			if confidence, ok := result["confidence"]; ok {
				fmt.Printf("  置信度:   %.1f%%\n", confidence.(float64)*100)
			}
			if description, ok := result["description"]; ok {
				fmt.Printf("  描述:     %v\n", description)
			}
			if colors, ok := result["colors"]; ok {
				fmt.Printf("  主要颜色: %v\n", colors)
			}
		}
	}

	// 输出使用统计
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
	fmt.Printf("停止原因: %s\n", message.StopReason)
}
