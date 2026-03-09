// 运行方式: go run 03_multi_agent.go
// 多 Agent 协作：规划 Agent 分解任务，执行 Agent 完成子任务，汇总 Agent 整合结果

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// Agent 代表一个专门化的智能体
type Agent struct {
	Name         string // Agent 名称
	SystemPrompt string // 系统提示，定义 Agent 的角色和行为
}

// callClaude 调用 Claude API 获取响应
func callClaude(client *anthropic.Client, system string, userMsg string) (string, error) {
	message, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System:    []anthropic.TextBlockParam{{Text: system}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})
	if err != nil {
		return "", err
	}

	// 提取文本响应
	var result strings.Builder
	for _, block := range message.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			result.WriteString(tb.Text)
		}
	}
	return result.String(), nil
}

func main() {
	client := anthropic.NewClient()

	// =====================
	// 定义三个专门化的 Agent
	// =====================

	// 1. 规划 Agent：负责分析任务并拆解为子任务
	planner := Agent{
		Name: "规划者",
		SystemPrompt: `你是一个任务规划专家。你的职责是：
1. 分析用户的需求
2. 将复杂任务拆解为 2-3 个清晰的子任务
3. 每个子任务用一行描述，格式为 "子任务N: 具体描述"
4. 只输出子任务列表，不要其他内容`,
	}

	// 2. 执行 Agent：负责完成具体的子任务
	executor := Agent{
		Name: "执行者",
		SystemPrompt: `你是一个任务执行专家。你会收到一个具体的子任务，请：
1. 直接完成该任务
2. 输出简洁、高质量的结果
3. 不要解释你的过程，直接给出结果`,
	}

	// 3. 汇总 Agent：负责整合所有子任务的结果
	summarizer := Agent{
		Name: "汇总者",
		SystemPrompt: `你是一个内容整合专家。你会收到多个子任务的结果，请：
1. 将所有结果整合为一个连贯、完整的最终输出
2. 确保内容流畅、结构清晰
3. 添加适当的标题和过渡语`,
	}

	// =====================
	// 多 Agent 协作流程
	// =====================

	// 用户任务
	task := "帮我写一篇关于 Go 语言在 AI 开发中应用的短文，包括优势分析和实际场景举例"

	fmt.Printf("📋 原始任务: %s\n", task)
	fmt.Println(strings.Repeat("=", 60))

	// 步骤 1: 规划 Agent 分解任务
	fmt.Printf("\n🧠 [%s] 正在分析任务...\n", planner.Name)
	plan, err := callClaude(&client, planner.SystemPrompt, task)
	if err != nil {
		fmt.Printf("❌ 规划失败: %v\n", err)
		return
	}
	fmt.Printf("📝 任务拆解:\n%s\n", plan)

	// 步骤 2: 执行 Agent 逐个完成子任务
	fmt.Println(strings.Repeat("-", 40))

	// 将规划结果按行分割，提取子任务
	subtasks := strings.Split(plan, "\n")
	var results []string

	for i, subtask := range subtasks {
		subtask = strings.TrimSpace(subtask)
		if subtask == "" {
			continue
		}

		fmt.Printf("\n⚡ [%s] 执行子任务 %d: %s\n", executor.Name, i+1, subtask)
		result, err := callClaude(&client, executor.SystemPrompt, subtask)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue
		}
		results = append(results, fmt.Sprintf("=== 子任务 %d 结果 ===\n%s", i+1, result))
		fmt.Printf("✅ 子任务 %d 完成\n", i+1)
	}

	// 步骤 3: 汇总 Agent 整合结果
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("\n📊 [%s] 正在整合结果...\n", summarizer.Name)

	allResults := strings.Join(results, "\n\n")
	summaryPrompt := fmt.Sprintf("原始任务: %s\n\n各子任务的完成结果如下:\n\n%s\n\n请整合以上内容为一篇完整的文章。", task, allResults)

	finalResult, err := callClaude(&client, summarizer.SystemPrompt, summaryPrompt)
	if err != nil {
		fmt.Printf("❌ 汇总失败: %v\n", err)
		return
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📄 最终输出:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(finalResult)
}
