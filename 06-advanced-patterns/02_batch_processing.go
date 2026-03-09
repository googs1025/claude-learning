// 第六章示例 2：批量并发处理
// 运行方式: go run 02_batch_processing.go
//
// 本示例演示如何：
// 1. 使用 goroutine 并发调用 Claude API
// 2. 使用 sync.WaitGroup 等待所有请求完成
// 3. 使用 channel 安全收集结果
// 4. 并发分析多段文本的情感倾向
//
// Go 语言的并发原语非常适合批量 API 调用，可以显著减少总耗时。

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnalysisResult 存储单个分析任务的结果
type AnalysisResult struct {
	Index    int    // 输入文本的索引
	Text     string // 原始输入文本（截取前 30 个字符用于展示）
	Sentiment string // Claude 分析出的情感倾向
	Err      error  // 错误信息（如果有）
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// 准备多段待分析的文本
	texts := []string{
		"今天天气真好，阳光明媚，心情非常愉快！出去散步感觉太棒了。",
		"这个产品太让人失望了，质量差，客服态度也很恶劣，再也不会购买。",
		"会议讨论了下季度的项目计划，预计在三月份完成第一阶段的开发工作。",
		"虽然考试没考好，但我学到了很多知识，下次一定会更好的。",
		"交通拥堵严重，上班迟到了半小时，老板脸色很不好看，今天真倒霉。",
	}

	fmt.Printf("准备并发分析 %d 段文本的情感倾向...\n\n", len(texts))

	// 记录开始时间，用于对比并发 vs 串行的耗时
	startTime := time.Now()

	// 创建 channel 用于收集结果
	// 缓冲大小设为文本数量，避免 goroutine 阻塞
	resultCh := make(chan AnalysisResult, len(texts))

	// 使用 WaitGroup 等待所有 goroutine 完成
	var wg sync.WaitGroup

	// 启动并发请求
	for i, text := range texts {
		wg.Add(1)

		// 每个文本启动一个 goroutine 进行分析
		// 注意：将 i 和 text 作为参数传入，避免闭包捕获问题
		go func(index int, inputText string) {
			defer wg.Done()

			// 调用 Claude API 进行情感分析
			result := analyzeSentiment(ctx, &client, index, inputText)
			resultCh <- result
		}(i, text)
	}

	// 启动一个 goroutine 等待所有任务完成后关闭 channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 从 channel 中收集所有结果
	results := make([]AnalysisResult, 0, len(texts))
	for result := range resultCh {
		results = append(results, result)
	}

	// 计算总耗时
	elapsed := time.Since(startTime)

	// 按原始索引排序输出结果（因为并发完成顺序不确定）
	// 使用简单的排序：创建有序数组
	orderedResults := make([]AnalysisResult, len(texts))
	for _, r := range results {
		orderedResults[r.Index] = r
	}

	// 打印分析结果
	fmt.Println("========== 情感分析结果 ==========")
	for _, r := range orderedResults {
		if r.Err != nil {
			fmt.Printf("\n[文本 %d] 错误: %v\n", r.Index+1, r.Err)
			continue
		}
		fmt.Printf("\n[文本 %d] %s\n", r.Index+1, r.Text)
		fmt.Printf("  情感分析: %s\n", r.Sentiment)
	}

	// 打印性能统计
	fmt.Println("\n========== 性能统计 ==========")
	fmt.Printf("总共处理: %d 个文本\n", len(texts))
	fmt.Printf("并发耗时: %v\n", elapsed)
	fmt.Printf("平均耗时: %v/个\n", elapsed/time.Duration(len(texts)))
	fmt.Println("提示：串行处理的话总耗时约为单个请求耗时 x 文本数量")
	fmt.Println("      并发处理的总耗时接近单个请求的耗时（取决于最慢的请求）")

	// 统计成功/失败数量
	successCount := 0
	for _, r := range orderedResults {
		if r.Err == nil {
			successCount++
		}
	}
	fmt.Printf("成功: %d, 失败: %d\n", successCount, len(texts)-successCount)
}

// analyzeSentiment 调用 Claude API 分析单段文本的情感倾向
func analyzeSentiment(ctx context.Context, client *anthropic.Client, index int, text string) AnalysisResult {
	// 截取文本用于展示
	displayText := text
	if len([]rune(displayText)) > 30 {
		displayText = string([]rune(displayText)[:30]) + "..."
	}

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				fmt.Sprintf(`请分析以下文本的情感倾向，只回复一个词：正面、负面 或 中性。然后用一句话简要说明原因。

格式：
情感：[正面/负面/中性]
原因：[一句话说明]

文本："%s"`, text),
			)),
		},
	})
	if err != nil {
		return AnalysisResult{
			Index: index,
			Text:  displayText,
			Err:   err,
		}
	}

	// 提取回复文本
	sentiment := ""
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			sentiment = v.Text
		}
	}

	return AnalysisResult{
		Index:     index,
		Text:      displayText,
		Sentiment: sentiment,
	}
}
