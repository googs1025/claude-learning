// 第六章示例 3：重试策略（指数退避 + 抖动）
// 运行方式: go run 03_retry_strategy.go
//
// 本示例演示如何：
// 1. 实现带指数退避和随机抖动的重试逻辑
// 2. 识别可重试的错误（429 速率限制、529 服务过载）
// 3. 区分可重试错误和不可重试错误（如 401 认证失败）
// 4. 在实际 API 调用中应用重试策略
//
// 生产环境中，重试策略是保证服务稳定性的关键组件。

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// RetryConfig 定义重试策略的配置参数
type RetryConfig struct {
	MaxRetries  int           // 最大重试次数
	BaseDelay   time.Duration // 初始延迟时间
	MaxDelay    time.Duration // 最大延迟时间（防止指数增长过大）
	JitterRatio float64       // 抖动比例（0.0 ~ 1.0），用于随机化延迟
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,                    // 最多重试 3 次
		BaseDelay:   1 * time.Second,      // 初始延迟 1 秒
		MaxDelay:    30 * time.Second,     // 最大延迟 30 秒
		JitterRatio: 0.5,                  // 50% 的随机抖动
	}
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	config := DefaultRetryConfig()
	fmt.Println("========== 重试策略演示 ==========")
	fmt.Printf("最大重试次数: %d\n", config.MaxRetries)
	fmt.Printf("初始延迟: %v\n", config.BaseDelay)
	fmt.Printf("最大延迟: %v\n", config.MaxDelay)
	fmt.Printf("抖动比例: %.0f%%\n\n", config.JitterRatio*100)

	// 演示 1：正常请求（通常不需要重试）
	fmt.Println("--- 演示 1：正常 API 调用（带重试保护） ---")
	message, err := callWithRetry(ctx, &client, config, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("用一句话解释什么是指数退避（exponential backoff）。")),
		},
	})
	if err != nil {
		log.Fatalf("请求最终失败: %v", err)
	}

	fmt.Print("Claude 回复: ")
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}

	// 演示 2：展示延迟计算过程
	fmt.Println("\n--- 演示 2：延迟时间计算示例 ---")
	fmt.Println("以下展示每次重试的延迟时间（包含随机抖动）：")
	for i := 0; i <= config.MaxRetries; i++ {
		delay := calculateDelay(i, config)
		fmt.Printf("  第 %d 次重试: 延迟 %v\n", i+1, delay)
	}

	// 演示 3：展示错误分类逻辑
	fmt.Println("\n--- 演示 3：错误分类说明 ---")
	fmt.Println("可重试的错误:")
	fmt.Println("  - 429: 速率限制（请求过于频繁）")
	fmt.Println("  - 529: 服务过载（Anthropic 服务端繁忙）")
	fmt.Println("  - 500: 服务器内部错误")
	fmt.Println("  - 网络超时、连接中断等瞬时错误")
	fmt.Println()
	fmt.Println("不可重试的错误:")
	fmt.Println("  - 400: 请求参数错误（需要修改请求）")
	fmt.Println("  - 401: 认证失败（API Key 无效）")
	fmt.Println("  - 403: 权限不足")
	fmt.Println("  - 404: 资源不存在")

	fmt.Println("\n========== 演示完成 ==========")
	fmt.Println("提示：在生产环境中，建议将重试逻辑封装为中间件或包装函数，")
	fmt.Println("      避免在每个 API 调用处重复编写重试代码。")
}

// callWithRetry 使用重试策略调用 Claude API
// 如果遇到可重试的错误（速率限制、服务过载等），会自动重试
func callWithRetry(
	ctx context.Context,
	client *anthropic.Client,
	config RetryConfig,
	params anthropic.MessageNewParams,
) (*anthropic.Message, error) {

	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// 如果不是第一次尝试，先等待
		if attempt > 0 {
			delay := calculateDelay(attempt, config)
			fmt.Printf("  [重试 %d/%d] 等待 %v 后重试...\n", attempt, config.MaxRetries, delay)

			// 使用 select 支持上下文取消
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("重试被取消: %w", ctx.Err())
			case <-time.After(delay):
				// 等待完成，继续重试
			}
		}

		// 发送请求
		message, err := client.Messages.New(ctx, params)
		if err == nil {
			// 请求成功
			if attempt > 0 {
				fmt.Printf("  [成功] 第 %d 次重试成功\n", attempt)
			}
			return message, nil
		}

		lastErr = err

		// 判断是否可以重试
		if !isRetryableError(err) {
			fmt.Printf("  [不可重试] 错误类型不支持重试: %v\n", err)
			return nil, fmt.Errorf("不可重试的错误: %w", err)
		}

		fmt.Printf("  [可重试错误] %v\n", err)
	}

	return nil, fmt.Errorf("达到最大重试次数 (%d)，最后一次错误: %w", config.MaxRetries, lastErr)
}

// isRetryableError 判断错误是否可以重试
// 只有特定的 HTTP 状态码和网络错误才应该重试
func isRetryableError(err error) bool {
	// 检查是否是 Anthropic SDK 的错误类型
	var apiErr *anthropic.Error
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case 429: // 速率限制 - 请求过于频繁
			return true
		case 529: // 服务过载 - Anthropic 服务端繁忙
			return true
		case 500: // 服务器内部错误 - 可能是瞬时故障
			return true
		case 502, 503, 504: // 网关错误 - 通常是瞬时的
			return true
		default:
			// 其他状态码（400、401、403、404 等）不应重试
			return false
		}
	}

	// 对于非 API 错误（如网络超时），通常也值得重试
	// 这里简单地认为非 API 错误都可以重试
	// 在生产环境中，可能需要更精细的判断
	return true
}

// calculateDelay 计算第 n 次重试的延迟时间
// 使用指数退避 + 随机抖动策略
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// 指数退避：delay = baseDelay * 2^attempt
	// 例如：1s, 2s, 4s, 8s, 16s, ...
	delay := float64(config.BaseDelay) * math.Pow(2, float64(attempt))

	// 限制最大延迟，防止指数增长过大
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// 添加随机抖动，避免多个客户端同时重试导致"惊群效应"
	// 抖动范围：delay * (1 - jitterRatio) ~ delay * (1 + jitterRatio)
	if config.JitterRatio > 0 {
		jitter := delay * config.JitterRatio
		delay = delay - jitter + (rand.Float64() * 2 * jitter)
	}

	return time.Duration(delay)
}
