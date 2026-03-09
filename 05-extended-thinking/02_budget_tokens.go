// 第五章示例 2：思考预算控制
// 运行方式: go run 02_budget_tokens.go
//
// 本示例演示如何：
// 1. 使用不同的 BudgetTokens 值来控制推理深度
// 2. 对比低预算和高预算对思考过程的影响
// 3. 理解 BudgetTokens 的约束条件（最小 1024，必须小于 MaxTokens）

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// analyzeWithBudget 使用指定的思考预算发送请求并返回结果
func analyzeWithBudget(ctx context.Context, client anthropic.Client, budget int64, problem string) {
	fmt.Printf("\n{'='*50}\n")
	fmt.Printf(">>> 思考预算: %d tokens\n", budget)
	fmt.Println(strings.Repeat("=", 50))

	// MaxTokens 必须大于 BudgetTokens
	// 这里设置为 budget + 4000，确保有足够空间给最终回答
	maxTokens := budget + 4000

	message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: maxTokens,
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: budget,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(problem)),
		},
	})
	if err != nil {
		log.Printf("API 调用失败 (budget=%d): %v", budget, err)
		return
	}

	// 统计思考内容的长度
	thinkingLength := 0
	answerLength := 0

	for _, block := range message.Content {
		switch v := block.AsAny().(type) {
		case anthropic.ThinkingBlock:
			thinkingLength = len(v.Thinking)
			// 只显示思考过程的前 200 个字符，避免输出过长
			preview := v.Thinking
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("\n[思考过程预览] %s\n", preview)

		case anthropic.TextBlock:
			answerLength = len(v.Text)
			fmt.Printf("\n[最终回答]\n%s\n", v.Text)
		}
	}

	// 输出对比数据
	fmt.Printf("\n--- 统计 ---\n")
	fmt.Printf("思考内容长度: %d 字符\n", thinkingLength)
	fmt.Printf("回答内容长度: %d 字符\n", answerLength)
	fmt.Printf("输入 token: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token: %d\n", message.Usage.OutputTokens)
}

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	fmt.Println("=== 思考预算对比实验 ===")
	fmt.Println()
	fmt.Println("相同的问题，不同的思考预算，观察推理深度的变化。")
	fmt.Println("注意：BudgetTokens 最小值为 1024，且必须小于 MaxTokens。")

	// 一个需要多步推理的问题
	problem := `一列火车从 A 站出发，以每小时 60 公里的速度行驶。
2 小时后，另一列火车从 B 站出发，以每小时 90 公里的速度沿相同方向追赶。
A 站和 B 站相距 30 公里（B 站在 A 站和火车行驶方向之间）。
请问第二列火车出发后多久能追上第一列火车？`

	fmt.Printf("\n问题: %s\n", problem)

	// 对比两种不同的思考预算
	// 低预算：1024 tokens（最小值）
	analyzeWithBudget(ctx, client, 1024, problem)

	// 高预算：8000 tokens（更充裕的推理空间）
	analyzeWithBudget(ctx, client, 8000, problem)

	fmt.Println("\n=== 实验总结 ===")
	fmt.Println("更高的思考预算通常会产生更详细的推理过程。")
	fmt.Println("但并非总是需要最大预算——简单问题用较少的预算即可。")
	fmt.Println("根据问题复杂度合理设置预算，可以节省 token 开销。")
}
