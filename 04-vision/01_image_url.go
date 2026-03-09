// 第四章示例 1：通过 URL 分析图片
// 运行方式: go run 01_image_url.go
//
// 本示例演示如何：
// 1. 使用公开 URL 将图片发送给 Claude
// 2. 让 Claude 描述图片中的内容
// 3. 在一条消息中混合图片和文本内容块

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

	// 使用一张公开的猫咪图片 URL
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Cat03.jpg/1200px-Cat03.jpg"

	fmt.Println("=== 通过 URL 分析图片 ===")
	fmt.Printf("图片 URL: %s\n\n", imageURL)

	// 发送图片分析请求
	// NewImageBlock 创建一个图片内容块
	// URLImageSourceParam 指定通过 URL 方式传入图片
	// 一条用户消息中可以包含多个内容块：这里同时包含图片块和文本块
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				// 图片内容块：通过 URL 加载图片
				anthropic.NewImageBlock(anthropic.URLImageSourceParam{
					URL: imageURL,
				}),
				// 文本内容块：告诉 Claude 要做什么
				anthropic.NewTextBlock("请详细描述这张图片中的内容，包括主体、颜色、背景等细节。用中文回答。"),
			),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 输出 Claude 的图片描述
	fmt.Println("=== Claude 的图片描述 ===")
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}

	// 输出使用统计
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
}
