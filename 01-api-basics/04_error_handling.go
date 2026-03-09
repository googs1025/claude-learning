// 第一章示例 4：错误处理模式
// 运行方式: go run 04_error_handling.go
//
// 本示例演示如何：
// 1. 正确处理 API 调用中可能出现的各种错误
// 2. 使用 errors.As 提取 anthropic.Error 类型的详细信息
// 3. 根据错误状态码采取不同的处理策略
//
// 常见的 API 错误：
// - 401: API 密钥无效或缺失
// - 400: 请求参数错误（如 MaxTokens 为负数）
// - 429: 请求频率超限（Rate Limit）
// - 500: Anthropic 服务端内部错误
// - 529: API 过载（Overloaded）

package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	ctx := context.Background()

	// ==================== 场景 1：正常请求（成功案例） ====================
	fmt.Println("========== 场景 1：正常请求 ==========")
	normalRequest(ctx)

	// ==================== 场景 2：使用无效的 API 密钥 ====================
	fmt.Println("\n========== 场景 2：无效的 API 密钥 ==========")
	invalidKeyRequest(ctx)

	// ==================== 场景 3：无效的请求参数 ====================
	fmt.Println("\n========== 场景 3：无效的请求参数 ==========")
	invalidParamsRequest(ctx)
}

// normalRequest 演示正常的 API 请求及其错误处理
func normalRequest(ctx context.Context) {
	client := anthropic.NewClient()

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 128,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("说'你好'")),
		},
	})
	if err != nil {
		// 尝试提取 API 错误的详细信息
		handleError(err)
		return
	}

	// 请求成功，打印回复
	fmt.Println("请求成功！")
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Printf("Claude: %s\n", v.Text)
		}
	}
}

// invalidKeyRequest 演示使用无效 API 密钥时的错误处理
func invalidKeyRequest(ctx context.Context) {
	// 使用 option.WithAPIKey 手动指定一个无效的 API 密钥
	// 这会覆盖环境变量中的设置
	client := anthropic.NewClient(
		option.WithAPIKey("sk-ant-invalid-key-for-testing"),
	)

	_, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 128,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("你好")),
		},
	})
	if err != nil {
		// 预期会收到 401 认证错误
		handleError(err)
		return
	}

	fmt.Println("请求意外成功（不应到达此处）")
}

// invalidParamsRequest 演示请求参数错误时的处理
func invalidParamsRequest(ctx context.Context) {
	client := anthropic.NewClient()

	// 故意使用无效的模型名称来触发错误
	_, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "invalid-model-name", // 不存在的模型名称
		MaxTokens: 128,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("你好")),
		},
	})
	if err != nil {
		// 预期会收到 400 或 404 错误
		handleError(err)
		return
	}

	fmt.Println("请求意外成功（不应到达此处）")
}

// handleError 统一的错误处理函数
// 展示如何区分 API 错误和其他类型的错误
func handleError(err error) {
	// 使用 errors.As 尝试将错误转换为 anthropic.Error 类型
	// 这是 Go 标准库推荐的错误类型断言方式
	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) {
		// 成功提取为 API 错误，可以获取详细信息
		fmt.Printf("API 错误!\n")
		fmt.Printf("  状态码: %d\n", apiErr.StatusCode) // HTTP 状态码
		fmt.Printf("  错误信息: %s\n", apiErr.Error())    // 使用 Error() 方法获取错误描述

		// 根据状态码给出不同的处理建议
		switch apiErr.StatusCode {
		case 401:
			fmt.Println("  建议: 检查 API 密钥是否正确设置")
		case 400:
			fmt.Println("  建议: 检查请求参数是否符合 API 规范")
		case 404:
			fmt.Println("  建议: 检查模型名称或 API 路径是否正确")
		case 429:
			fmt.Println("  建议: 请求频率过高，请稍后重试或降低请求频率")
		case 500:
			fmt.Println("  建议: Anthropic 服务端错误，请稍后重试")
		case 529:
			fmt.Println("  建议: API 过载，请稍后重试")
		default:
			fmt.Printf("  建议: 未知错误码 %d，请查阅 API 文档\n", apiErr.StatusCode)
		}
	} else {
		// 非 API 错误（可能是网络连接错误、超时等）
		log.Printf("非 API 错误: %v\n", err)
		fmt.Println("  建议: 检查网络连接和 API 端点配置")
	}
}
