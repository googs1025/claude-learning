// 第二章示例 1：Few-shot 提示（少样本提示）
// 运行方式: go run 01_few_shot.go
//
// 本示例演示如何：
// 1. 使用 user/assistant 消息对构造示例（few-shot examples）
// 2. 通过少量示例教会 Claude 执行情感分析任务
// 3. 对比有示例和无示例两种方式的输出差异
//
// Few-shot 提示的核心思想：
// - 不需要用大量文字描述规则，只需提供几个输入/输出示例
// - Claude 会从示例中学习模式，并将其应用到新的输入上
// - 在 API 中，示例通过交替的 user 和 assistant 消息来构造

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

// extractText 从 Claude 的响应中提取文本内容
// 这是一个通用的辅助函数，遍历响应的 Content 数组，
// 将所有文本块的内容拼接后返回
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

func main() {
	// 创建客户端（自动读取 ANTHROPIC_API_KEY 环境变量）
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 第一部分：不使用 few-shot 的情感分析（zero-shot）
	// ========================================
	// 直接要求 Claude 做情感分析，不提供任何示例
	// Claude 可能会返回比较冗长或格式不统一的回答
	fmt.Println("=== Zero-shot 情感分析（无示例）===")

	zeroShotMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"请判断以下评论的情感倾向：'这家餐厅的菜品味道一般，但服务员态度很好'",
			)),
		},
	})
	if err != nil {
		log.Fatalf("Zero-shot API 调用失败: %v", err)
	}
	fmt.Println(extractText(zeroShotMessage))

	// ========================================
	// 第二部分：使用 few-shot 的情感分析
	// ========================================
	// 通过 3 个 user/assistant 消息对作为示例
	// 教会 Claude 我们期望的输出格式：只返回"正面"、"负面"或"中性"
	fmt.Println("\n=== Few-shot 情感分析（有示例）===")

	fewShotMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		// 系统提示词说明任务要求
		System: []anthropic.TextBlockParam{
			{Text: "你是一个情感分析助手。对于每条用户评论，只返回一个词：正面、负面或中性。不要包含任何其他内容。"},
		},
		Messages: []anthropic.MessageParam{
			// === 示例 1：正面情感 ===
			// 用户提供一条正面评论
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"这个产品质量非常好，物超所值，强烈推荐！",
			)),
			// 助手给出期望的回答格式
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("正面")),

			// === 示例 2：负面情感 ===
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"太失望了，用了三天就坏了，客服也不管。",
			)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("负面")),

			// === 示例 3：中性情感 ===
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"包装还行，功能中规中矩，没什么特别的。",
			)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock("中性")),

			// === 实际要分析的评论 ===
			// 这是我们真正想要分析的新输入
			// Claude 会根据前面的示例，以相同的格式返回结果
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"这家餐厅的菜品味道一般，但服务员态度很好",
			)),
		},
	})
	if err != nil {
		log.Fatalf("Few-shot API 调用失败: %v", err)
	}
	fmt.Printf("情感判断结果: %s\n", extractText(fewShotMessage))

	// ========================================
	// 第三部分：批量分析多条评论
	// ========================================
	// 复用相同的 few-shot 模板，对多条评论进行分析
	// 演示如何在实际应用中循环调用 API
	fmt.Println("\n=== 批量情感分析 ===")

	// 待分析的评论列表
	reviews := []string{
		"太棒了！下次还会来！",
		"性价比很低，不值这个价格。",
		"还可以吧，马马虎虎。",
		"客服解决了我的问题，很满意。",
		"物流太慢了，等了两周才到。",
	}

	// few-shot 示例消息（作为"模板"复用）
	// 这些示例在每次请求中都相同，只是最后的实际输入不同
	exampleMessages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock("这个产品质量非常好，物超所值！")),
		anthropic.NewAssistantMessage(anthropic.NewTextBlock("正面")),
		anthropic.NewUserMessage(anthropic.NewTextBlock("太失望了，用了三天就坏了。")),
		anthropic.NewAssistantMessage(anthropic.NewTextBlock("负面")),
		anthropic.NewUserMessage(anthropic.NewTextBlock("一般般吧，没什么特别的。")),
		anthropic.NewAssistantMessage(anthropic.NewTextBlock("中性")),
	}

	for _, review := range reviews {
		// 将模板示例和实际输入组合成完整的消息列表
		// 注意：每次调用都需要重新构建消息列表（append 不会修改原始切片）
		messages := make([]anthropic.MessageParam, len(exampleMessages))
		copy(messages, exampleMessages)
		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(review)))

		response, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 64, // 只需要一个词，设置较小的 MaxTokens 节省资源
			System: []anthropic.TextBlockParam{
				{Text: "你是一个情感分析助手。对于每条用户评论，只返回一个词：正面、负面或中性。"},
			},
			Messages: messages,
		})
		if err != nil {
			log.Printf("分析评论失败: %v", err)
			continue
		}
		fmt.Printf("  评论: %-20s → 情感: %s\n", review, extractText(response))
	}

	fmt.Println("\n=== Few-shot 提示技巧总结 ===")
	fmt.Println("1. 示例数量：通常 2-5 个示例即可，太多反而增加成本")
	fmt.Println("2. 示例质量：选择有代表性的、覆盖不同情况的示例")
	fmt.Println("3. 格式一致：所有示例的输出格式必须统一")
	fmt.Println("4. 系统提示词 + few-shot 配合使用效果更佳")
}
