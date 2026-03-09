// 运行方式: go run main.go
// 综合项目1：CLI 聊天机器人
// 功能：多轮对话、流式输出、对话历史管理

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ChatBot 聊天机器人
type ChatBot struct {
	client       anthropic.Client
	history      []anthropic.MessageParam
	systemPrompt string
	model        anthropic.Model
	maxTokens    int64
}

// NewChatBot 创建聊天机器人
func NewChatBot() *ChatBot {
	return &ChatBot{
		client: anthropic.NewClient(),
		history: make([]anthropic.MessageParam, 0),
		systemPrompt: `你是一个友好、知识渊博的 AI 助手。
请用中文回答问题，回答要简洁但信息丰富。
如果用户的问题涉及代码，请提供可运行的示例。`,
		model:     anthropic.ModelClaudeSonnet4_5_20250929,
		maxTokens: 2048,
	}
}

// Chat 发送消息并获取流式响应
func (bot *ChatBot) Chat(ctx context.Context, userInput string) error {
	// 添加用户消息到历史
	bot.history = append(bot.history,
		anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
	)

	// 使用流式 API
	stream := bot.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     bot.model,
		MaxTokens: bot.maxTokens,
		System:    []anthropic.TextBlockParam{{Text: bot.systemPrompt}},
		Messages:  bot.history,
	})

	fmt.Print("\n🤖 ")

	// 使用 Accumulate 收集完整消息
	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		if err := message.Accumulate(event); err != nil {
			return err
		}

		// 实时打印文本增量
		switch v := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch d := v.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				fmt.Print(d.Text)
			}
		}
	}

	if stream.Err() != nil {
		return stream.Err()
	}

	fmt.Println()

	// 将助手回复加入历史
	bot.history = append(bot.history, message.ToParam())

	// 显示 token 使用情况
	fmt.Printf("   [tokens: 输入=%d 输出=%d]\n",
		message.Usage.InputTokens, message.Usage.OutputTokens)

	return nil
}

// ClearHistory 清空对话历史
func (bot *ChatBot) ClearHistory() {
	bot.history = make([]anthropic.MessageParam, 0)
	fmt.Println("🗑  对话历史已清空")
}

// SetSystemPrompt 设置系统提示
func (bot *ChatBot) SetSystemPrompt(prompt string) {
	bot.systemPrompt = prompt
	fmt.Printf("✅ 系统提示已更新为: %s\n", prompt[:min(50, len(prompt))]+"...")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func printHelp() {
	fmt.Println(`
📋 可用命令:
  /clear    - 清空对话历史
  /system   - 设置新的系统提示（后面跟提示内容）
  /history  - 显示对话轮次
  /help     - 显示此帮助
  /quit     - 退出程序
`)
}

func main() {
	bot := NewChatBot()
	scanner := bufio.NewScanner(os.Stdin)
	// 增大缓冲区以支持长输入
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	ctx := context.Background()

	fmt.Println("💬 Claude 聊天机器人 (输入 /help 查看命令)")
	fmt.Println(strings.Repeat("=", 45))

	for {
		fmt.Print("\n你: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// 处理命令
		switch {
		case input == "/quit":
			fmt.Println("👋 再见！")
			return
		case input == "/clear":
			bot.ClearHistory()
			continue
		case input == "/help":
			printHelp()
			continue
		case input == "/history":
			fmt.Printf("📜 当前对话轮次: %d\n", len(bot.history)/2)
			continue
		case strings.HasPrefix(input, "/system "):
			bot.SetSystemPrompt(strings.TrimPrefix(input, "/system "))
			continue
		}

		// 正常对话
		if err := bot.Chat(ctx, input); err != nil {
			fmt.Printf("\n❌ 错误: %v\n", err)
		}
	}
}
