// 第二章示例 2：思维链提示（Chain of Thought Prompting）
// 运行方式: go run 02_chain_of_thought.go
//
// 本示例演示如何：
// 1. 对比直接回答和思维链回答在推理任务上的差异
// 2. 使用"请一步一步思考"来引导 Claude 进行逐步推理
// 3. 理解思维链提示对数学/逻辑问题的效果提升
//
// 思维链（CoT）的核心思想：
// - 人类在解决复杂问题时也需要列出中间步骤
// - 通过要求 Claude 展示思考过程，可以减少推理错误
// - 中间步骤的可见性也便于我们验证和调试

package main

import (
	"context"
	"fmt"
	"log"

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

// askClaude 是一个封装的 API 调用辅助函数
// 接受系统提示词和用户提问，返回 Claude 的回复文本
// 通过封装减少重复代码，让主函数更专注于演示逻辑
func askClaude(client *anthropic.Client, ctx context.Context, systemPrompt string, userPrompt string) string {
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 2048, // 思维链回答通常较长，需要更多 token
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	}

	// 如果提供了系统提示词，则添加到参数中
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
	// 示例 1：数学应用题 - 对比直接回答 vs 思维链
	// ========================================
	// 这类问题需要多步运算，直接回答容易出错
	mathProblem := `小明有 15 个苹果。他给了小红 3 个，又从小华那里拿到了原来数量的两倍。
然后他把手里苹果的三分之一给了妈妈。请问小明最后还有多少个苹果？`

	// --- 方式 1：直接回答，不要求推理过程 ---
	fmt.Println("=== 数学题：直接回答（无思维链）===")
	directAnswer := askClaude(&client, ctx,
		"请直接给出答案，不需要解释过程。",
		mathProblem,
	)
	fmt.Println(directAnswer)

	// --- 方式 2：要求逐步思考 ---
	// 通过在提示词中加入"请一步一步思考"来激活思维链
	fmt.Println("\n=== 数学题：思维链回答（一步一步思考）===")
	cotAnswer := askClaude(&client, ctx,
		"",
		mathProblem+"\n\n请一步一步思考，展示你的计算过程，最后给出答案。",
	)
	fmt.Println(cotAnswer)

	// ========================================
	// 示例 2：逻辑推理问题
	// ========================================
	// 逻辑推理问题更需要思维链，因为需要逐步排除和推断
	logicProblem := `有 A、B、C、D 四个人，他们的职业分别是医生、教师、工程师和律师（不一定按此顺序）。
已知以下条件：
1. A 和医生不是邻居
2. B 的收入比教师高
3. C 经常和工程师一起打球
4. D 认识律师但不认识医生
5. 医生的收入最低

请推断每个人的职业。`

	fmt.Println("\n=== 逻辑推理题：思维链回答 ===")
	logicAnswer := askClaude(&client, ctx,
		// 系统提示词中设定推理框架
		"你是一个逻辑推理专家。请按照以下步骤来解题："+
			"1. 列出所有已知条件；"+
			"2. 从最确定的线索开始推断；"+
			"3. 逐步排除不可能的选项；"+
			"4. 验证最终答案是否满足所有条件。",
		logicProblem,
	)
	fmt.Println(logicAnswer)

	// ========================================
	// 示例 3：代码分析问题
	// ========================================
	// 代码 bug 分析也是思维链的典型应用场景
	codeProblem := `请分析以下 Go 代码是否有 bug，如果有请指出：

func calculateAverage(numbers []int) float64 {
    sum := 0
    for _, n := range numbers {
        sum += n
    }
    return float64(sum) / float64(len(numbers))
}

func main() {
    scores := []int{}
    avg := calculateAverage(scores)
    fmt.Println("平均分:", avg)
}`

	fmt.Println("\n=== 代码分析：思维链回答 ===")
	codeAnswer := askClaude(&client, ctx,
		"你是一位资深 Go 开发者。分析代码时请按以下步骤："+
			"1. 先理解代码的意图；"+
			"2. 逐行检查可能的问题；"+
			"3. 考虑边界情况和异常输入；"+
			"4. 给出修复建议。",
		codeProblem+"\n\n请一步一步分析。",
	)
	fmt.Println(codeAnswer)

	// ========================================
	// 思维链提示技巧总结
	// ========================================
	fmt.Println("\n=== 思维链提示技巧总结 ===")
	fmt.Println("1. 适用场景：数学计算、逻辑推理、代码分析、复杂决策")
	fmt.Println("2. 触发方式：在提示词中加入'请一步一步思考'或'请展示推理过程'")
	fmt.Println("3. 系统提示词中设定推理框架（步骤 1、2、3...）效果更好")
	fmt.Println("4. 注意：思维链会增加输出 token 数量，需要适当增大 MaxTokens")
	fmt.Println("5. 对于简单的事实性问题，思维链反而可能过度复杂化")
}
