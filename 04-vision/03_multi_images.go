// 第四章示例 3：多图对比分析
// 运行方式: go run 03_multi_images.go
//
// 本示例演示如何：
// 1. 在一条消息中发送多张图片给 Claude
// 2. 让 Claude 对比分析多张图片的异同
// 3. 理解多图消息的构造方式

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

	// 准备两张不同的公开图片 URL
	// 图片 1：一只猫
	imageURL1 := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Cat03.jpg/1200px-Cat03.jpg"
	// 图片 2：一只狗
	imageURL2 := "https://upload.wikimedia.org/wikipedia/commons/thumb/2/26/YellowLabradorLooking_new.jpg/1200px-YellowLabradorLooking_new.jpg"

	fmt.Println("=== 多图对比分析 ===")
	fmt.Printf("图片 1: %s\n", imageURL1)
	fmt.Printf("图片 2: %s\n\n", imageURL2)

	// 发送多图分析请求
	// 在 NewUserMessage 中可以传入任意数量的内容块
	// 这里包含：图片1 + 图片2 + 文本提示
	// Claude 会同时理解所有图片并进行对比
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				// 第一张图片
				anthropic.NewImageBlock(anthropic.URLImageSourceParam{
					URL: imageURL1,
				}),
				// 第二张图片
				anthropic.NewImageBlock(anthropic.URLImageSourceParam{
					URL: imageURL2,
				}),
				// 文本提示：要求对比分析
				anthropic.NewTextBlock("请对比这两张图片，分析它们的异同点。包括：1) 各自的主体是什么 2) 颜色和构图的差异 3) 它们之间有什么共同点。用中文回答。"),
			),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 输出对比分析结果
	fmt.Println("=== Claude 的对比分析 ===")
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}

	// 输出使用统计
	// 注意：多图输入会消耗更多的输入 token
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d（多图输入会消耗更多 token）\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
}
