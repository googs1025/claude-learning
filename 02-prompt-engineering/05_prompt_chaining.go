// 第二章示例 5：提示词链（Prompt Chaining）
// 运行方式: go run 05_prompt_chaining.go
//
// 本示例演示如何：
// 1. 将复杂任务分解为多个简单步骤，逐步调用 API
// 2. 将上一步的输出作为下一步的输入（链式调用）
// 3. 在每一步使用不同的系统提示词来优化结果
//
// 提示词链的核心思想：
// - 一个大而复杂的提示词往往不如多个小而精确的提示词
// - 分步处理可以在每一步进行质量检查和错误纠正
// - 每一步可以使用不同的角色/约束来获得最佳结果
//
// 本示例的流水线：
//   原始文本 → 第 1 步：摘要 → 第 2 步：翻译 → 第 3 步：关键词提取

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// extractText 从 Claude 响应中提取文本内容
func extractText(message *anthropic.Message) string {
	var result string
	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			result += v.Text
		}
	}
	return result
}

// callClaude 封装 API 调用逻辑
// systemPrompt: 系统提示词（定义当前步骤的角色和任务）
// userInput: 用户输入（通常是上一步的输出）
// maxTokens: 最大输出 token 数
// 返回 Claude 的文本回复
func callClaude(client *anthropic.Client, ctx context.Context, systemPrompt string, userInput string, maxTokens int) string {
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
		},
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	message, err := client.Messages.New(ctx, params)
	if err != nil {
		log.Fatalf("API 调用失败: %v", err)
	}
	return extractText(message)
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 准备原始输入文本
	// ========================================
	// 这是一篇中文技术文章的段落，我们将对它进行多步处理
	originalText := `
Go 语言（又称 Golang）是由 Google 的 Robert Griesemer、Rob Pike 和 Ken Thompson
于 2007 年开始设计，2009 年正式发布的开源编程语言。Go 语言的设计目标是提高程序员的
生产力，特别是在大规模软件系统的开发中。

Go 语言最突出的特点是其内置的并发支持。通过 goroutine（轻量级线程）和 channel（通道），
开发者可以轻松地编写高并发程序。与传统的线程模型不同，goroutine 的创建和切换成本极低，
一个程序可以轻松创建数十万个 goroutine。

Go 语言的另一个重要特性是其简洁的语法设计。Go 有意省略了许多其他语言中常见的特性，
如继承、异常处理、泛型（直到 1.18 版本才加入）等。这种"少即是多"的设计哲学使得
Go 代码易于阅读和维护。

在实际应用中，Go 已经成为云原生基础设施的首选语言。Docker、Kubernetes、Prometheus、
Terraform 等众多知名项目都是用 Go 编写的。Go 在微服务、网络编程、DevOps 工具等
领域有着广泛的应用。

近年来，Go 语言在国内的普及度也越来越高，字节跳动、阿里巴巴、腾讯等大公司都在
大规模使用 Go 进行后端服务开发。Go 的简单性和高性能使其成为后端开发者的热门选择。`

	fmt.Println("========================================")
	fmt.Println("提示词链演示：文本处理流水线")
	fmt.Println("========================================")
	fmt.Println("\n原始文本（中文技术文章）:")
	fmt.Println(originalText)
	fmt.Println(strings.Repeat("=", 50))

	// ========================================
	// 第 1 步：摘要（Summarization）
	// ========================================
	// 角色：文本摘要专家
	// 任务：将长文本浓缩为简短的摘要
	fmt.Println("\n>>> 第 1 步：文本摘要")
	fmt.Println(strings.Repeat("-", 40))

	summary := callClaude(&client, ctx,
		// 系统提示词定义摘要专家的角色和约束
		`你是一个专业的文本摘要助手。请将用户提供的文本压缩为一段简洁的摘要。

要求：
1. 摘要长度不超过 100 字
2. 保留核心信息和关键数据
3. 使用中文
4. 不要添加原文没有的信息
5. 直接输出摘要内容，不要加"摘要："等前缀`,
		originalText,
		512,
	)

	fmt.Println("摘要结果:")
	fmt.Println(summary)

	// ========================================
	// 第 2 步：翻译（Translation）
	// ========================================
	// 角色：专业翻译
	// 输入：第 1 步的摘要（中文）
	// 输出：英文翻译
	// 注意：这里使用的是上一步的输出（summary），而不是原始文本
	fmt.Println("\n>>> 第 2 步：中译英")
	fmt.Println(strings.Repeat("-", 40))

	translation := callClaude(&client, ctx,
		`你是一位专业的技术文档翻译，精通中英文互译。

要求：
1. 将用户提供的中文文本翻译为英文
2. 保持技术术语的准确性（如 goroutine、channel 不翻译）
3. 使用正式的技术文档风格
4. 直接输出翻译结果，不要加任何前缀或说明`,
		// 关键：这里传入的是第 1 步的输出，不是原始文本
		summary,
		512,
	)

	fmt.Println("翻译结果:")
	fmt.Println(translation)

	// ========================================
	// 第 3 步：关键词提取（Keyword Extraction）
	// ========================================
	// 角色：信息提取专家
	// 输入：第 2 步的英文翻译
	// 输出：提取的关键词列表
	fmt.Println("\n>>> 第 3 步：关键词提取")
	fmt.Println(strings.Repeat("-", 40))

	keywords := callClaude(&client, ctx,
		`你是一个关键词提取专家。从用户提供的英文文本中提取最重要的技术关键词。

要求：
1. 提取 5-8 个关键词
2. 每行一个关键词
3. 关键词保持英文
4. 按重要性从高到低排序
5. 格式：每行一个，前面加序号（如 "1. keyword"）
6. 不要添加任何解释或说明`,
		// 传入第 2 步的翻译结果
		translation,
		256,
	)

	fmt.Println("关键词:")
	fmt.Println(keywords)

	// ========================================
	// 汇总流水线结果
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("流水线执行完毕！")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("\n原始文本长度: %d 字符\n", len([]rune(originalText)))
	fmt.Printf("摘要长度:     %d 字符\n", len([]rune(summary)))
	fmt.Printf("翻译长度:     %d 字符\n", len(translation))

	// ========================================
	// 高级用法：带验证的提示词链
	// ========================================
	// 在实际生产中，每一步之后可以加入验证逻辑
	// 如果验证不通过，可以重试或调整提示词
	fmt.Println("\n========================================")
	fmt.Println("高级用法：带验证的提示词链")
	fmt.Println("========================================")

	// 第 1 步：生成代码
	fmt.Println("\n>>> 步骤 A：根据需求生成 Go 代码")
	codeRequirement := "写一个 Go 函数，接受一个整数切片，返回其中所有偶数的平方和。"

	generatedCode := callClaude(&client, ctx,
		`你是一个 Go 语言代码生成器。根据用户的需求生成 Go 代码。

要求：
1. 只输出代码，不需要解释
2. 包含函数签名和实现
3. 代码简洁高效
4. 不需要 package 和 import 声明，只需要函数本身`,
		codeRequirement,
		512,
	)
	fmt.Println(generatedCode)

	// 第 2 步：审查代码（验证步骤）
	// 将上一步生成的代码交给"审查员"检查
	fmt.Println("\n>>> 步骤 B：审查生成的代码")
	reviewResult := callClaude(&client, ctx,
		`你是一个严格的 Go 代码审查员。审查用户提供的代码。

请检查：
1. 逻辑是否正确
2. 是否处理了边界情况（空切片等）
3. 命名是否符合 Go 规范
4. 是否有性能问题

输出格式：
- 如果代码没问题：输出 "PASS" 加一句简短说明
- 如果有问题：输出 "FAIL" 加问题描述和修复建议

用中文回答。`,
		fmt.Sprintf("需求：%s\n\n代码：\n%s", codeRequirement, generatedCode),
		512,
	)
	fmt.Println(reviewResult)

	// 第 3 步：生成测试用例
	fmt.Println("\n>>> 步骤 C：为代码生成测试用例")
	testCases := callClaude(&client, ctx,
		`你是一个 Go 测试专家。为用户提供的函数生成测试用例。

要求：
1. 使用 Go 的 testing 包
2. 包含正常情况和边界情况（空切片、全奇数、全偶数、负数等）
3. 使用表格驱动测试的风格
4. 只输出测试代码
5. 不需要 package 声明`,
		fmt.Sprintf("函数需求：%s\n\n函数实现：\n%s", codeRequirement, generatedCode),
		1024,
	)
	fmt.Println(testCases)

	// 总结
	fmt.Println("\n=== 提示词链技巧总结 ===")
	fmt.Println("1. 分解复杂任务：将大任务拆分为 2-5 个简单步骤")
	fmt.Println("2. 每步一个角色：为每个步骤设定专门的系统提示词")
	fmt.Println("3. 中间结果传递：上一步的输出直接作为下一步的输入")
	fmt.Println("4. 加入验证步骤：在关键步骤后增加审查/验证环节")
	fmt.Println("5. 控制 token 用量：每一步设定合理的 MaxTokens")
	fmt.Println("6. 注意延迟：每次 API 调用都有网络延迟，步骤过多会影响总体响应时间")
}
