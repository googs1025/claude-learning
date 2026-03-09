// 第五章示例 3：流式思考输出
// 运行方式: go run 03_streaming_thinking.go
//
// 本示例演示如何：
// 1. 在流式模式下使用扩展思考
// 2. 实时输出思考过程（ThinkingDelta 事件）
// 3. 实时输出文本回答（TextDelta 事件）
// 4. 区分思考阶段和回答阶段的流式内容

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// 创建客户端
	client := anthropic.NewClient()
	ctx := context.Background()

	fmt.Println("=== 流式扩展思考 ===")
	fmt.Println("实时观察 Claude 的思考过程和回答生成。")
	fmt.Println()

	// 一个需要逻辑推理的问题
	problem := `有三个盒子，分别标记为 A、B、C。
已知：
1. 金币在其中一个盒子里
2. A 盒上写着"金币在这里"
3. B 盒上写着"金币不在这里"
4. C 盒上写着"金币不在 A 盒"
5. 只有一个盒子上的标签是真话，其他两个是假话

请问金币在哪个盒子里？`

	fmt.Printf("问题: %s\n\n", problem)

	// 使用流式模式发送请求
	// NewStreaming 返回一个流式对象，可以逐步读取响应
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:    anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 16000,
		Thinking: anthropic.ThinkingConfigParamUnion{
			OfEnabled: &anthropic.ThinkingConfigEnabledParam{
				BudgetTokens: 10000,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(problem)),
		},
	})

	// 标记当前处于哪个阶段（思考 or 回答）
	inThinking := false
	inText := false

	// 使用 Accumulate 模式收集最终消息
	message := anthropic.Message{}

	// 逐个处理流式事件
	for stream.Next() {
		event := stream.Current()

		// 累积事件到最终消息
		message.Accumulate(event)

		// 流式事件需要先通过 event.AsAny() 获取具体事件类型
		// 然后对 ContentBlockDeltaEvent 中的 Delta 再做类型判断
		switch evt := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := evt.Delta.AsAny().(type) {
			case anthropic.ThinkingDelta:
				// 思考增量：Claude 的推理过程，逐步到达
				if !inThinking {
					fmt.Println("=== 思考过程（实时）===")
					inThinking = true
				}
				// 实时打印思考内容片段（不换行，让内容连续显示）
				fmt.Print(delta.Thinking)

			case anthropic.TextDelta:
				// 文本增量：最终回答的内容，逐步到达
				if !inText {
					if inThinking {
						fmt.Println() // 思考结束后换行
						fmt.Println()
					}
					fmt.Println("=== 最终回答（实时）===")
					inText = true
				}
				// 实时打印回答内容片段
				fmt.Print(delta.Text)
			}
		}
	}

	// 检查流式传输过程中是否有错误
	if err := stream.Err(); err != nil {
		log.Fatalf("流式传输错误: %v", err)
	}

	fmt.Println() // 最后的换行
	fmt.Println("\n=== 使用统计 ===")
	fmt.Printf("输入 token 数: %d\n", message.Usage.InputTokens)
	fmt.Printf("输出 token 数: %d\n", message.Usage.OutputTokens)
}
