// 第六章示例 4：Token 计数与成本估算
// 运行方式: go run 04_token_counting.go
//
// 本示例演示如何：
// 1. 跟踪每次 API 调用的输入/输出 token 使用量
// 2. 累计多次调用的总 token 消耗
// 3. 基于模型定价估算 API 调用费用
// 4. 比较不同提示词长度对成本的影响
//
// 了解 token 消耗是控制 API 成本的第一步。

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

// ModelPricing 定义模型的定价（美元/百万 token）
// 价格来自 Anthropic 官方定价页面，可能会更新
type ModelPricing struct {
	Name               string
	InputPerMillion    float64 // 输入 token 价格（美元/百万 token）
	OutputPerMillion   float64 // 输出 token 价格（美元/百万 token）
	CacheWritePerMillion  float64 // 缓存写入价格（美元/百万 token）
	CacheReadPerMillion   float64 // 缓存读取价格（美元/百万 token）
}

// UsageTracker 跟踪 token 使用量和成本
type UsageTracker struct {
	Pricing      ModelPricing
	TotalInput   int64
	TotalOutput  int64
	CallCount    int
	CallDetails  []CallDetail
}

// CallDetail 记录单次调用的详情
type CallDetail struct {
	Label       string
	InputTokens  int64
	OutputTokens int64
	Cost         float64
}

// NewUsageTracker 创建一个新的使用量跟踪器
func NewUsageTracker(pricing ModelPricing) *UsageTracker {
	return &UsageTracker{
		Pricing: pricing,
	}
}

// RecordUsage 记录一次 API 调用的 token 使用量
func (t *UsageTracker) RecordUsage(label string, msg *anthropic.Message) {
	inputTokens := msg.Usage.InputTokens
	outputTokens := msg.Usage.OutputTokens

	// 计算本次调用的费用
	cost := t.calculateCost(inputTokens, outputTokens)

	// 累计总量
	t.TotalInput += inputTokens
	t.TotalOutput += outputTokens
	t.CallCount++

	// 记录详情
	t.CallDetails = append(t.CallDetails, CallDetail{
		Label:        label,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		Cost:         cost,
	})

	// 实时打印本次调用的 token 信息
	fmt.Printf("  输入 token: %d (≈ $%.6f)\n", inputTokens,
		float64(inputTokens)/1_000_000*t.Pricing.InputPerMillion)
	fmt.Printf("  输出 token: %d (≈ $%.6f)\n", outputTokens,
		float64(outputTokens)/1_000_000*t.Pricing.OutputPerMillion)
	fmt.Printf("  本次费用: ≈ $%.6f\n", cost)
}

// calculateCost 计算一次调用的费用
func (t *UsageTracker) calculateCost(inputTokens, outputTokens int64) float64 {
	inputCost := float64(inputTokens) / 1_000_000 * t.Pricing.InputPerMillion
	outputCost := float64(outputTokens) / 1_000_000 * t.Pricing.OutputPerMillion
	return inputCost + outputCost
}

// PrintSummary 打印总结报告
func (t *UsageTracker) PrintSummary() {
	totalCost := t.calculateCost(t.TotalInput, t.TotalOutput)

	fmt.Println("\n╔══════════════════════════════════════════════╗")
	fmt.Println("║            Token 使用量与成本报告              ║")
	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Printf("║ 模型: %-39s ║\n", t.Pricing.Name)
	fmt.Printf("║ 总调用次数: %-33d ║\n", t.CallCount)
	fmt.Println("╠══════════════════════════════════════════════╣")

	// 打印每次调用的详情
	for i, detail := range t.CallDetails {
		fmt.Printf("║ 调用 %d: %-36s ║\n", i+1, detail.Label)
		fmt.Printf("║   输入: %-6d  输出: %-6d  费用: $%.6f ║\n",
			detail.InputTokens, detail.OutputTokens, detail.Cost)
	}

	fmt.Println("╠══════════════════════════════════════════════╣")
	fmt.Printf("║ 总输入 token:  %-30d ║\n", t.TotalInput)
	fmt.Printf("║ 总输出 token:  %-30d ║\n", t.TotalOutput)
	fmt.Printf("║ 总 token:      %-30d ║\n", t.TotalInput+t.TotalOutput)
	fmt.Println("╠══════════════════════════════════════════════╣")

	// 费用明细
	inputCost := float64(t.TotalInput) / 1_000_000 * t.Pricing.InputPerMillion
	outputCost := float64(t.TotalOutput) / 1_000_000 * t.Pricing.OutputPerMillion
	fmt.Printf("║ 输入费用:  $%-34.6f ║\n", inputCost)
	fmt.Printf("║ 输出费用:  $%-34.6f ║\n", outputCost)
	fmt.Printf("║ 总费用:    $%-34.6f ║\n", totalCost)
	fmt.Println("╠══════════════════════════════════════════════╣")

	// 费用换算（帮助理解规模）
	fmt.Printf("║ 按此速率，$1 可调用约 %d 次                   ║\n",
		int(1.0/totalCost*float64(t.CallCount)))
	fmt.Println("╚══════════════════════════════════════════════╝")
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// Claude Sonnet 4.5 的定价（截至 2025 年）
	// 请查阅 https://www.anthropic.com/pricing 获取最新价格
	sonnetPricing := ModelPricing{
		Name:               "Claude Sonnet 4.5",
		InputPerMillion:    3.00,  // $3.00 / 百万输入 token
		OutputPerMillion:   15.00, // $15.00 / 百万输出 token
		CacheWritePerMillion: 3.75, // $3.75 / 百万缓存写入 token
		CacheReadPerMillion:  0.30, // $0.30 / 百万缓存读取 token
	}

	tracker := NewUsageTracker(sonnetPricing)

	// ==================== 调用 1：简短问题 ====================
	fmt.Println("========== 调用 1：简短问题 ==========")
	msg1, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Go 语言的创建者是谁？请用一句话回答。")),
		},
	})
	if err != nil {
		log.Fatalf("调用 1 失败: %v", err)
	}
	printReply(msg1)
	tracker.RecordUsage("简短问题", msg1)

	// ==================== 调用 2：中等问题 ====================
	fmt.Println("\n========== 调用 2：中等长度问题 ==========")
	msg2, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"请解释 Go 语言中 goroutine 和 channel 的关系，以及它们在并发编程中的作用。请给出一个简单的示例代码。",
			)),
		},
	})
	if err != nil {
		log.Fatalf("调用 2 失败: %v", err)
	}
	printReply(msg2)
	tracker.RecordUsage("中等问题（含代码）", msg2)

	// ==================== 调用 3：带系统提示的问题 ====================
	fmt.Println("\n========== 调用 3：带系统提示的问题 ==========")
	msg3, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		System: []anthropic.TextBlockParam{
			{Text: "你是一个经验丰富的 Go 语言专家。回答问题时要简洁明了，突出重点。每个要点不超过一行。"},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"列出 Go 语言相比其他语言的 5 个主要优势。",
			)),
		},
	})
	if err != nil {
		log.Fatalf("调用 3 失败: %v", err)
	}
	printReply(msg3)
	tracker.RecordUsage("带系统提示", msg3)

	// ==================== 调用 4：多轮对话（token 较多） ====================
	fmt.Println("\n========== 调用 4：多轮对话 ==========")
	msg4, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("什么是 context.Context？")),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock(
				"context.Context 是 Go 标准库中的接口，用于在 goroutine 之间传递截止时间、取消信号和请求范围的值。它是 Go 并发编程中控制 goroutine 生命周期的核心工具。",
			)),
			anthropic.NewUserMessage(anthropic.NewTextBlock("请给出一个超时控制的实际代码示例。")),
		},
	})
	if err != nil {
		log.Fatalf("调用 4 失败: %v", err)
	}
	printReply(msg4)
	tracker.RecordUsage("多轮对话", msg4)

	// 打印总结报告
	tracker.PrintSummary()

	// 额外的成本优化建议
	fmt.Println("\n========== 成本优化建议 ==========")
	fmt.Println("1. 精简系统提示词：过长的系统提示会增加每次调用的输入 token")
	fmt.Println("2. 使用 Prompt 缓存：对频繁使用的系统提示词启用缓存")
	fmt.Println("3. 控制 MaxTokens：设置合理的上限，避免不必要的长回复")
	fmt.Println("4. 选择合适的模型：简单任务可使用 Haiku（更便宜）")
	fmt.Println("5. 批量处理：合并多个小请求为一个大请求")
	fmt.Println("6. 定期回顾对话历史：截断过长的对话避免 token 膨胀")
}

// printReply 打印 Claude 的回复文本
func printReply(msg *anthropic.Message) {
	fmt.Print("[Claude]: ")
	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}
	fmt.Println()
}
