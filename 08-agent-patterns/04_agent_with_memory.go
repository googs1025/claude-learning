// 运行方式: go run 04_agent_with_memory.go
// 带记忆的 Agent：维护对话历史和关键信息摘要

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// Memory 表示 Agent 的记忆系统
type Memory struct {
	// 短期记忆：完整的对话历史
	ConversationHistory []anthropic.MessageParam

	// 长期记忆：关键信息摘要（模拟持久化存储）
	KeyFacts []string

	// 最大对话轮次（超过后进行摘要压缩）
	MaxHistorySize int
}

// NewMemory 创建新的记忆系统
func NewMemory(maxSize int) *Memory {
	return &Memory{
		ConversationHistory: make([]anthropic.MessageParam, 0),
		KeyFacts:            make([]string, 0),
		MaxHistorySize:      maxSize,
	}
}

// AddUserMessage 添加用户消息到记忆
func (m *Memory) AddUserMessage(text string) {
	m.ConversationHistory = append(m.ConversationHistory,
		anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
	)
}

// AddAssistantResponse 添加助手响应到记忆
func (m *Memory) AddAssistantResponse(msg *anthropic.Message) {
	m.ConversationHistory = append(m.ConversationHistory, msg.ToParam())
}

// AddKeyFact 添加关键事实到长期记忆
func (m *Memory) AddKeyFact(fact string) {
	m.KeyFacts = append(m.KeyFacts, fact)
}

// GetSystemPromptWithMemory 构建包含记忆信息的系统提示
func (m *Memory) GetSystemPromptWithMemory() string {
	prompt := `你是一个有记忆能力的智能助手。你能够记住之前对话中的关键信息。

## 你的能力：
1. 记住用户告诉你的重要信息（姓名、偏好、项目细节等）
2. 在后续对话中主动使用这些信息
3. 当对话涉及之前提到的内容时，展示你的记忆

## 重要规则：
- 如果用户告诉你关于他们的新信息，请在回复末尾用 [记住: xxx] 的格式标注需要记忆的内容
- 主动关联之前的对话内容`

	// 如果有长期记忆，加入系统提示
	if len(m.KeyFacts) > 0 {
		prompt += "\n\n## 已记住的关键信息：\n"
		for i, fact := range m.KeyFacts {
			prompt += fmt.Sprintf("%d. %s\n", i+1, fact)
		}
	}

	return prompt
}

// CompressIfNeeded 如果对话历史过长，进行压缩
func (m *Memory) CompressIfNeeded(client *anthropic.Client) {
	if len(m.ConversationHistory) <= m.MaxHistorySize {
		return
	}

	fmt.Println("\n🗜  记忆压缩: 对话历史过长，正在提取关键信息...")

	// 使用 Claude 来摘要前面的对话
	summaryMessages := make([]anthropic.MessageParam, len(m.ConversationHistory))
	copy(summaryMessages, m.ConversationHistory)
	summaryMessages = append(summaryMessages,
		anthropic.NewUserMessage(anthropic.NewTextBlock(
			"请用 2-3 句话总结我们之前对话的核心内容和关键信息，每条信息一行。只输出摘要，不要其他内容。")),
	)

	msg, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 256,
		Messages:  summaryMessages,
	})
	if err != nil {
		fmt.Printf("⚠️  压缩失败: %v\n", err)
		return
	}

	// 提取摘要文本
	for _, block := range msg.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			// 将摘要加入长期记忆
			for _, line := range strings.Split(tb.Text, "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					m.AddKeyFact(line)
				}
			}
		}
	}

	// 只保留最近的几轮对话
	keep := m.MaxHistorySize / 2
	if keep < 2 {
		keep = 2
	}
	m.ConversationHistory = m.ConversationHistory[len(m.ConversationHistory)-keep:]
	fmt.Println("✅ 记忆压缩完成")
}

// extractMemoryTags 从回复中提取 [记住: xxx] 标签
func extractMemoryTags(text string) []string {
	var facts []string
	for _, line := range strings.Split(text, "\n") {
		if idx := strings.Index(line, "[记住:"); idx >= 0 {
			end := strings.Index(line[idx:], "]")
			if end > 0 {
				fact := strings.TrimSpace(line[idx+len("[记住:") : idx+end])
				facts = append(facts, fact)
			}
		}
		// 也支持英文格式
		if idx := strings.Index(line, "[Remember:"); idx >= 0 {
			end := strings.Index(line[idx:], "]")
			if end > 0 {
				fact := strings.TrimSpace(line[idx+len("[Remember:") : idx+end])
				facts = append(facts, fact)
			}
		}
	}
	return facts
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// 创建记忆系统（最多保留 10 轮对话）
	memory := NewMemory(10)

	fmt.Println("🧠 带记忆的 Agent（输入 'quit' 退出，输入 'memory' 查看记忆）")
	fmt.Println(strings.Repeat("=", 50))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n你: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}
		if input == "quit" {
			fmt.Println("👋 再见！")
			break
		}
		if input == "memory" {
			fmt.Println("\n📋 当前记忆状态:")
			fmt.Printf("   对话轮次: %d\n", len(memory.ConversationHistory))
			fmt.Printf("   长期记忆: %d 条\n", len(memory.KeyFacts))
			for i, fact := range memory.KeyFacts {
				fmt.Printf("   %d. %s\n", i+1, fact)
			}
			continue
		}

		// 检查是否需要压缩记忆
		memory.CompressIfNeeded(&client)

		// 添加用户消息
		memory.AddUserMessage(input)

		// 调用 Claude，使用带记忆的系统提示
		message, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_5_20250929,
			MaxTokens: 1024,
			System:    []anthropic.TextBlockParam{{Text: memory.GetSystemPromptWithMemory()}},
			Messages:  memory.ConversationHistory,
		})
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			continue
		}

		// 添加助手响应到记忆
		memory.AddAssistantResponse(message)

		// 显示回复并提取记忆标签
		for _, block := range message.Content {
			if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
				fmt.Printf("\n🤖 Agent: %s\n", tb.Text)

				// 自动提取需要记忆的信息
				facts := extractMemoryTags(tb.Text)
				for _, fact := range facts {
					memory.AddKeyFact(fact)
					fmt.Printf("   💾 已记住: %s\n", fact)
				}
			}
		}
	}
}
