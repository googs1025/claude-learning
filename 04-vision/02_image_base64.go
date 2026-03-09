// 第四章示例 2：通过 Base64 编码分析图片
// 运行方式: go run 02_image_base64.go
//
// 本示例演示如何：
// 1. 使用 Go 的 image/png 包在内存中生成一张简单的 PNG 图片
// 2. 将图片编码为 Base64 字符串
// 3. 通过 Base64 方式将图片发送给 Claude 进行分析

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

// createSampleImage 在内存中创建一张简单的测试 PNG 图片
// 图片包含红、绿、蓝、黄四个色块，方便验证 Claude 的视觉识别能力
func createSampleImage() ([]byte, error) {
	// 创建一张 200x200 像素的 RGBA 图片
	width, height := 200, 200
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充四个不同颜色的色块（2x2 网格布局）
	// 左上角：红色
	// 右上角：绿色
	// 左下角：蓝色
	// 右下角：黄色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var c color.RGBA
			switch {
			case x < width/2 && y < height/2:
				c = color.RGBA{R: 255, G: 0, B: 0, A: 255} // 红色
			case x >= width/2 && y < height/2:
				c = color.RGBA{R: 0, G: 255, B: 0, A: 255} // 绿色
			case x < width/2 && y >= height/2:
				c = color.RGBA{R: 0, G: 0, B: 255, A: 255} // 蓝色
			default:
				c = color.RGBA{R: 255, G: 255, B: 0, A: 255} // 黄色
			}
			img.Set(x, y, c)
		}
	}

	// 将图片编码为 PNG 格式的字节数据
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("PNG 编码失败: %w", err)
	}

	return buf.Bytes(), nil
}

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	fmt.Println("=== 通过 Base64 分析图片 ===")

	// 第一步：生成测试图片
	fmt.Println("正在生成测试图片（200x200 像素，四色块）...")
	imageData, err := createSampleImage()
	if err != nil {
		log.Fatalf("创建图片失败: %v", err)
	}
	fmt.Printf("图片大小: %d 字节\n\n", len(imageData))

	// 第二步：将图片数据编码为 Base64 字符串
	// Claude API 需要接收 Base64 编码的图片数据
	base64String := base64.StdEncoding.EncodeToString(imageData)

	// 第三步：发送 Base64 图片给 Claude
	// Base64ImageSourceParam 需要指定：
	// - MediaType: 图片的 MIME 类型（image/png、image/jpeg 等）
	// - Data: Base64 编码后的图片字符串
	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				// 图片内容块：通过 Base64 传入图片
				anthropic.NewImageBlock(anthropic.Base64ImageSourceParam{
					MediaType: "image/png",    // 指定图片格式
					Data:      base64String,   // Base64 编码后的图片数据
				}),
				// 文本内容块：要求 Claude 描述图片
				anthropic.NewTextBlock("请描述这张图片中的内容。你看到了什么颜色和形状？用中文回答。"),
			),
		},
	})
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}

	// 输出 Claude 的分析结果
	fmt.Println("=== Claude 的图片分析 ===")
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
