// 第二章示例 3：角色提示（Role Prompting）
// 运行方式: go run 03_role_prompting.go
//
// 本示例演示如何：
// 1. 通过系统提示词（System Prompt）为 Claude 指定不同角色
// 2. 对比同一问题在不同角色设定下的回答差异
// 3. 组合角色 + 约束条件来精确控制输出
//
// 角色提示的核心思想：
// - 系统提示词是设定 Claude 行为的最佳位置
// - 角色设定会影响回答的视角、深度、用词风格
// - 结合具体的约束条件可以获得更精准的输出

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

// askWithRole 使用指定的角色（系统提示词）向 Claude 提问
// roleName: 角色名称（仅用于日志打印）
// systemPrompt: 系统提示词，定义 Claude 的角色和行为
// question: 用户问题
func askWithRole(client *anthropic.Client, ctx context.Context, roleName string, systemPrompt string, question string) {
	fmt.Printf("\n--- 角色: %s ---\n", roleName)

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		// System 字段接受 TextBlockParam 切片
		// 这是设定角色的关键位置
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(question)),
		},
	})
	if err != nil {
		log.Printf("角色 [%s] 的 API 调用失败: %v", roleName, err)
		return
	}

	fmt.Println(extractText(message))
	fmt.Println(strings.Repeat("-", 50))
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 实验 1：同一个技术问题，不同专家角色的回答
	// ========================================
	// 问题：如何处理 Go 中的并发安全？
	// 不同角色会从不同角度回答这个问题
	techQuestion := "在 Go 语言中，如何安全地在多个 goroutine 之间共享数据？请给出建议和示例。"

	fmt.Println("========================================")
	fmt.Println("实验 1：同一问题，不同专家角色")
	fmt.Println("问题:", techQuestion)
	fmt.Println("========================================")

	// 角色 1：Go 语言专家
	// 这个角色会侧重于 Go 特有的并发原语和最佳实践
	askWithRole(&client, ctx,
		"Go 语言专家",
		`你是一位资深的 Go 语言专家，拥有 10 年的 Go 开发经验。
你的回答特点：
- 使用 Go 的惯用写法（idiomatic Go）
- 引用官方文档和标准库
- 提供可运行的代码示例
- 关注性能和最佳实践
请用中文回答，代码注释也用中文。回答控制在 300 字以内。`,
		techQuestion,
	)

	// 角色 2：计算机科学教授
	// 这个角色会从理论层面解释并发问题
	askWithRole(&client, ctx,
		"计算机科学教授",
		`你是一位计算机科学教授，专注于操作系统和并发编程领域。
你的回答特点：
- 从理论角度解释问题（互斥、信号量、内存模型等）
- 使用学术术语但会给出通俗解释
- 会提到经典的并发问题（如生产者-消费者、读者-写者）
- 注重概念理解而非具体实现
请用中文回答，回答控制在 300 字以内。`,
		techQuestion,
	)

	// 角色 3：技术写作者
	// 这个角色会用初学者友好的方式解释
	askWithRole(&client, ctx,
		"技术写作者",
		`你是一位经验丰富的技术博客作者，擅长将复杂概念用简单的语言解释。
你的回答特点：
- 使用生活中的类比来解释技术概念
- 循序渐进，从简单到复杂
- 避免过多的专业术语
- 提供直观易懂的示例
- 语言轻松友好
请用中文回答，回答控制在 300 字以内。`,
		techQuestion,
	)

	// ========================================
	// 实验 2：角色 + 详细约束条件
	// ========================================
	// 演示如何通过组合角色和约束来精确控制输出
	fmt.Println("\n========================================")
	fmt.Println("实验 2：角色 + 约束条件组合")
	fmt.Println("========================================")

	// 角色：代码审查员，带有严格的输出格式约束
	codeReviewQuestion := `请审查以下 Go 代码：

func getUser(id string) (*User, error) {
    resp, err := http.Get("http://api.example.com/users/" + id)
    if err != nil {
        return nil, err
    }
    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}`

	askWithRole(&client, ctx,
		"代码审查员（带约束）",
		`你是一位严格的 Go 代码审查员。审查代码时必须按以下固定格式输出：

## 问题列表
- [严重程度: 高/中/低] 问题描述

## 修复建议
针对每个问题给出修复代码

## 总体评价
一句话总结代码质量

注意：
- 关注安全性、错误处理、资源泄露、性能问题
- 每个问题必须标注严重程度
- 修复建议必须包含可运行的代码
- 用中文回答`,
		codeReviewQuestion,
	)

	// ========================================
	// 实验 3：多角色协作（模拟辩论）
	// ========================================
	// 演示如何让 Claude 从多个角度分析同一个决策
	fmt.Println("\n========================================")
	fmt.Println("实验 3：多角色协作 - 技术选型分析")
	fmt.Println("========================================")

	decisionQuestion := "我们团队在考虑将微服务架构改为单体应用，你怎么看？"

	// 支持方观点
	askWithRole(&client, ctx,
		"单体架构支持者",
		`你是一位单体架构的坚定支持者。你认为微服务被过度炒作了。
你的论点应包括：
- 运维复杂度的对比
- 开发效率的考量
- 适用团队规模的分析
请坚持你的立场，用中文回答，控制在 200 字以内。`,
		decisionQuestion,
	)

	// 反对方观点
	askWithRole(&client, ctx,
		"微服务架构支持者",
		`你是一位微服务架构的坚定支持者。你认为微服务是现代架构的最佳选择。
你的论点应包括：
- 可扩展性优势
- 技术栈灵活性
- 故障隔离能力
请坚持你的立场，用中文回答，控制在 200 字以内。`,
		decisionQuestion,
	)

	// 中立仲裁者
	askWithRole(&client, ctx,
		"架构顾问（中立）",
		`你是一位中立的资深架构顾问。你需要客观地分析微服务和单体架构各自的优劣。
你的分析应包括：
- 适用场景的对比
- 关键决策因素
- 给出具体的选型建议
请保持中立客观，用中文回答，控制在 200 字以内。`,
		decisionQuestion,
	)

	// 总结
	fmt.Println("\n=== 角色提示技巧总结 ===")
	fmt.Println("1. 系统提示词是设定角色的最佳位置，比在用户消息中说'请扮演...'更有效")
	fmt.Println("2. 详细的角色描述（经验、专长、风格）比简单的角色名称效果更好")
	fmt.Println("3. 角色 + 输出格式约束的组合可以获得高度可控的输出")
	fmt.Println("4. 多角色分析同一问题可以帮助做出更全面的决策")
	fmt.Println("5. 注意：角色设定不能让 Claude 做出危险或不道德的行为")
}
