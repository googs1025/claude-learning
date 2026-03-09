// 第六章示例 1：Prompt 缓存
// 运行方式: go run 01_prompt_caching.go
//
// 本示例演示如何：
// 1. 使用 CacheControl 将大型系统提示词标记为可缓存
// 2. 多次请求复用缓存的系统提示词，减少延迟和费用
// 3. 通过 Usage 字段观察缓存命中情况
//
// 注意：Prompt 缓存要求内容至少 1024 个 token 才能被缓存。
// 缓存有效期为 5 分钟，在此期间的后续请求会命中缓存。

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// 构造一个足够长的系统提示词（需要至少 1024 个 token 才能被缓存）
	// 在实际应用中，这通常是产品规则文档、API 规范、代码库上下文等
	longSystemPrompt := buildLongSystemPrompt()
	fmt.Printf("系统提示词长度: %d 字符\n\n", len(longSystemPrompt))

	// 构造带缓存控制的系统提示词
	// CacheControl 设置为 "ephemeral"，表示该内容块可以被缓存
	// 缓存在创建后 5 分钟内有效
	systemPrompt := []anthropic.TextBlockParam{
		{
			Text: longSystemPrompt,
			CacheControl: anthropic.CacheControlEphemeralParam{
				Type: "ephemeral",
			},
		},
	}

	// ==================== 第一次请求（创建缓存） ====================
	fmt.Println("========== 第一次请求（创建缓存） ==========")

	message1, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("请简要介绍你作为技术文档助手的主要职责。")),
		},
	})
	if err != nil {
		log.Fatalf("第一次请求失败: %v", err)
	}

	printResponse("第一次回复", message1)
	printUsage("第一次", message1)

	// ==================== 第二次请求（命中缓存） ====================
	fmt.Println("\n========== 第二次请求（应命中缓存） ==========")

	message2, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System:    systemPrompt, // 使用相同的系统提示词，应命中缓存
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("用户问了一个关于 Go 语言错误处理的问题，你会怎么回答？")),
		},
	})
	if err != nil {
		log.Fatalf("第二次请求失败: %v", err)
	}

	printResponse("第二次回复", message2)
	printUsage("第二次", message2)

	// ==================== 第三次请求（继续命中缓存） ====================
	fmt.Println("\n========== 第三次请求（继续命中缓存） ==========")

	message3, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System:    systemPrompt, // 同样的系统提示词
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("请列出文档编写的三个最佳实践。")),
		},
	})
	if err != nil {
		log.Fatalf("第三次请求失败: %v", err)
	}

	printResponse("第三次回复", message3)
	printUsage("第三次", message3)

	// ==================== 总结 ====================
	fmt.Println("\n========== 缓存效果总结 ==========")
	fmt.Println("第一次请求：创建缓存，输入 token 按正常价格计费")
	fmt.Println("后续请求：命中缓存的部分按缓存价格计费（约为正常价格的 10%）")
	fmt.Println("提示：缓存有效期为 5 分钟，适合短时间内多次使用相同上下文的场景")
}

// buildLongSystemPrompt 构造一个足够长的系统提示词
// 在实际应用中，这可能是从文件加载的产品文档、API 规范等
func buildLongSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString("你是一个专业的技术文档助手，专门帮助开发者编写和理解技术文档。\n\n")

	sb.WriteString("## 你的核心职责\n\n")
	sb.WriteString("1. 帮助用户理解复杂的技术概念\n")
	sb.WriteString("2. 提供清晰、准确的代码示例\n")
	sb.WriteString("3. 遵循最佳实践编写文档\n")
	sb.WriteString("4. 回答关于 Go、Python、JavaScript 等编程语言的问题\n\n")

	sb.WriteString("## 文档编写规范\n\n")
	sb.WriteString("### 代码示例规范\n")
	sb.WriteString("- 所有代码示例必须可运行\n")
	sb.WriteString("- 包含必要的导入语句\n")
	sb.WriteString("- 添加清晰的注释说明\n")
	sb.WriteString("- 使用有意义的变量名\n")
	sb.WriteString("- 处理所有可能的错误\n\n")

	sb.WriteString("### 文档结构规范\n")
	sb.WriteString("- 使用清晰的标题层级\n")
	sb.WriteString("- 每个章节有简要概述\n")
	sb.WriteString("- 提供前置条件说明\n")
	sb.WriteString("- 包含运行示例的步骤\n\n")

	// 添加详细的技术参考内容以达到缓存所需的最小长度
	topics := []struct {
		title   string
		content string
	}{
		{
			"Go 语言错误处理最佳实践",
			`Go 语言使用显式的错误返回值而非异常机制。每个可能失败的函数都应返回 error 类型作为最后一个返回值。
调用者必须检查错误并适当处理。使用 fmt.Errorf 和 %w 动词包装错误以保留错误链。
自定义错误类型应实现 error 接口。使用 errors.Is 和 errors.As 进行错误类型判断。
避免使用 panic 处理普通错误，panic 仅用于不可恢复的程序错误。`,
		},
		{
			"并发编程模式",
			`Go 的并发模型基于 CSP（Communicating Sequential Processes）理论。goroutine 是轻量级线程，channel 用于 goroutine 间通信。
使用 sync.WaitGroup 等待一组 goroutine 完成。使用 sync.Mutex 保护共享资源。
context.Context 用于控制 goroutine 的生命周期，支持超时和取消。
select 语句用于监听多个 channel 操作。避免 goroutine 泄漏，确保每个启动的 goroutine 都能正确退出。`,
		},
		{
			"接口设计原则",
			`Go 语言的接口是隐式实现的，一个类型只要实现了接口的所有方法就自动满足该接口。
接口应该小而精，一个接口通常只包含一到两个方法。io.Reader 和 io.Writer 是优秀接口设计的典范。
使用接口实现依赖注入，使代码更易测试。空接口 interface{} 应谨慎使用，优先使用具体的接口类型。
接口组合优于接口继承，Go 支持将多个小接口组合成更大的接口。`,
		},
		{
			"测试策略",
			`Go 内置了强大的测试框架。测试文件以 _test.go 结尾，测试函数以 Test 开头。
使用表驱动测试处理多个测试用例。使用 t.Run 创建子测试提高可读性。
使用 testify 等第三方库简化断言。编写基准测试评估性能。
使用 httptest 包测试 HTTP 处理器。使用 mock 和 stub 隔离外部依赖。
集成测试和单元测试应分开组织，使用 build tags 控制测试范围。`,
		},
		{
			"项目结构推荐",
			`Go 项目应遵循清晰的目录结构。cmd 目录存放可执行程序入口。internal 目录存放不对外暴露的包。
pkg 目录存放可供外部使用的库。api 目录存放 API 定义（如 protobuf 文件）。
web 目录存放前端资源。configs 目录存放配置文件。deployments 目录存放部署配置。
使用 go.mod 管理依赖，避免手动管理 vendor 目录。使用 Makefile 或 Taskfile 管理构建命令。`,
		},
		{
			"性能优化技巧",
			`使用 pprof 工具分析 CPU 和内存使用情况。避免不必要的内存分配，使用 sync.Pool 复用对象。
使用 strings.Builder 进行字符串拼接而非 + 操作符。预分配 slice 容量减少扩容开销。
使用 buffered channel 减少 goroutine 切换。避免在热路径上使用反射。
使用 atomic 包进行简单的原子操作，避免锁竞争。合理设置 GOMAXPROCS 控制并行度。
对于计算密集型任务，考虑使用 worker pool 模式控制并发数量。`,
		},
		{
			"安全编码实践",
			`永远不要将密钥硬编码在源代码中，使用环境变量或密钥管理服务。
验证所有用户输入，防止注入攻击。使用参数化查询防止 SQL 注入。
正确处理文件路径，防止路径遍历攻击。使用 crypto/rand 生成随机数而非 math/rand。
启用 HTTPS，正确配置 TLS。定期更新依赖以修复已知安全漏洞。
使用 gosec 等工具进行静态安全分析。遵循最小权限原则设计 API 接口。`,
		},
	}

	sb.WriteString("## 技术参考手册\n\n")
	for _, topic := range topics {
		sb.WriteString(fmt.Sprintf("### %s\n\n", topic.title))
		sb.WriteString(topic.content)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## 回答风格要求\n\n")
	sb.WriteString("1. 语言简洁明了，避免冗余\n")
	sb.WriteString("2. 提供可运行的代码示例\n")
	sb.WriteString("3. 解释为什么这样做，而不仅仅是怎么做\n")
	sb.WriteString("4. 对于复杂概念，使用类比帮助理解\n")
	sb.WriteString("5. 始终考虑边界情况和错误处理\n")
	sb.WriteString("6. 在适当时候提供性能相关的建议\n")
	sb.WriteString("7. 引用官方文档作为权威来源\n")

	return sb.String()
}

// printResponse 打印 Claude 的回复内容
func printResponse(label string, msg *anthropic.Message) {
	fmt.Printf("[%s]:\n", label)
	for _, block := range msg.Content {
		switch v := block.AsAny().(type) {
		case anthropic.TextBlock:
			fmt.Println(v.Text)
		}
	}
}

// printUsage 打印 token 使用情况，特别关注缓存相关字段
func printUsage(label string, msg *anthropic.Message) {
	fmt.Printf("\n--- %s Token 使用 ---\n", label)
	fmt.Printf("输入 token:       %d\n", msg.Usage.InputTokens)
	fmt.Printf("输出 token:       %d\n", msg.Usage.OutputTokens)

	// 缓存相关字段
	// CacheCreationInputTokens: 本次请求中被写入缓存的 token 数
	// CacheReadInputTokens: 本次请求中从缓存读取的 token 数
	fmt.Printf("缓存创建 token:   %d\n", msg.Usage.CacheCreationInputTokens)
	fmt.Printf("缓存读取 token:   %d\n", msg.Usage.CacheReadInputTokens)
}
