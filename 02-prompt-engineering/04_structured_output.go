// 第二章示例 4：结构化输出（Structured Output）
// 运行方式: go run 04_structured_output.go
//
// 本示例演示如何：
// 1. 通过系统提示词要求 Claude 输出严格的 JSON 格式
// 2. 在 Go 中解析 Claude 返回的 JSON 数据
// 3. 使用提示词技巧确保 JSON 的可靠性
//
// 结构化输出的核心思想：
// - LLM 的输出默认是自由文本，不方便程序处理
// - 通过精心设计的提示词，可以让 Claude 输出可解析的结构化数据
// - JSON 是最常用的结构化输出格式，Go 有完善的 JSON 解析支持

package main

import (
	"context"
	"encoding/json"
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

// ========================================
// 定义用于 JSON 解析的 Go 结构体
// ========================================

// BookReview 表示一本书的评论分析结果
// JSON tag 确保字段名与 Claude 输出的 JSON 键匹配
type BookReview struct {
	Title     string   `json:"title"`     // 书名
	Author    string   `json:"author"`    // 作者
	Rating    float64  `json:"rating"`    // 评分 (1-5)
	Summary   string   `json:"summary"`   // 摘要
	Pros      []string `json:"pros"`      // 优点列表
	Cons      []string `json:"cons"`      // 缺点列表
	Recommend bool     `json:"recommend"` // 是否推荐
}

// TaskList 表示从自然语言中提取的任务列表
type TaskList struct {
	Tasks []Task `json:"tasks"` // 任务列表
}

// Task 表示单个任务
type Task struct {
	Title    string `json:"title"`    // 任务标题
	Priority string `json:"priority"` // 优先级：high/medium/low
	Deadline string `json:"deadline"` // 截止日期
	Category string `json:"category"` // 分类
}

func main() {
	client := anthropic.NewClient()
	ctx := context.Background()

	// ========================================
	// 示例 1：书评分析 → JSON
	// ========================================
	// 让 Claude 分析一段书评文字，输出结构化的 JSON 数据
	fmt.Println("=== 示例 1：书评分析 → JSON ===")

	bookReviewText := `
我最近读完了《Go 语言程序设计》这本书，作者是 Alan Donovan 和 Brian Kernighan。
整体来说非常棒，给 4.5 分。这本书对 Go 的核心特性讲解得很透彻，
尤其是 goroutine 和 channel 那几章写得非常清晰。代码示例也很实用。
不过对于完全零基础的读者来说，前几章可能有点难度。
另外，书中关于泛型的内容比较少（因为出版时 Go 还没支持泛型）。
总的来说，我强烈推荐给有一定编程基础的人阅读。`

	reviewMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			// 系统提示词中明确要求 JSON 格式，并提供 schema
			// 关键技巧：
			// 1. 明确说"只输出 JSON"，不要有多余文字
			// 2. 提供 JSON 的结构说明
			// 3. 指定字段的数据类型
			{Text: `你是一个文本分析助手。请将用户提供的书评分析为 JSON 格式。

要求：
1. 只输出纯 JSON，不要包含 markdown 代码块标记或任何其他文字
2. 严格按照以下 schema：
{
  "title": "书名 (string)",
  "author": "作者 (string)",
  "rating": "评分 1-5 (number)",
  "summary": "一句话摘要 (string)",
  "pros": ["优点列表 (string[])"],
  "cons": ["缺点列表 (string[])"],
  "recommend": "是否推荐 (boolean)"
}
3. 所有字段都必须填写，不能为 null`},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(bookReviewText)),
		},
	})
	if err != nil {
		log.Fatalf("书评分析 API 调用失败: %v", err)
	}

	// 获取 Claude 返回的 JSON 字符串
	jsonStr := extractText(reviewMessage)
	fmt.Println("原始 JSON 输出:")
	fmt.Println(jsonStr)

	// 使用 Go 的 json.Unmarshal 解析 JSON
	// 如果 Claude 的输出不是合法 JSON，这里会报错
	var review BookReview
	if err := json.Unmarshal([]byte(jsonStr), &review); err != nil {
		log.Fatalf("JSON 解析失败: %v\n原始内容: %s", err, jsonStr)
	}

	// 使用解析后的结构化数据
	fmt.Println("\n解析后的结构化数据:")
	fmt.Printf("  书名: %s\n", review.Title)
	fmt.Printf("  作者: %s\n", review.Author)
	fmt.Printf("  评分: %.1f / 5.0\n", review.Rating)
	fmt.Printf("  摘要: %s\n", review.Summary)
	fmt.Printf("  优点: %v\n", review.Pros)
	fmt.Printf("  缺点: %v\n", review.Cons)
	fmt.Printf("  推荐: %v\n", review.Recommend)

	// ========================================
	// 示例 2：自然语言 → 任务列表
	// ========================================
	// 从一段自然语言描述中提取结构化的任务列表
	fmt.Println("\n=== 示例 2：自然语言 → 任务列表 ===")

	naturalText := `
明天上午要完成项目报告的初稿，这个比较紧急。
下周三之前需要把单元测试补全，优先级中等。
这周五要给新员工做一次 Go 语言培训，需要准备 PPT。
还有就是 bug #1234 要尽快修复，客户已经投诉了。
月底之前要完成 Q1 的技术规划文档。`

	taskMessage, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: `你是一个任务提取助手。从用户的自然语言描述中提取任务列表，输出为 JSON。

要求：
1. 只输出纯 JSON，不要包含 markdown 代码块标记或任何其他文字
2. 格式如下：
{
  "tasks": [
    {
      "title": "任务标题",
      "priority": "high/medium/low",
      "deadline": "截止日期（如果提到的话，格式 YYYY-MM-DD；如果没明确说，写 '待定'）",
      "category": "分类（如：开发、文档、培训、修复等）"
    }
  ]
}
3. 根据上下文判断优先级：紧急/尽快 → high，中等 → medium，其他 → low
4. 按优先级从高到低排序`},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(naturalText)),
		},
	})
	if err != nil {
		log.Fatalf("任务提取 API 调用失败: %v", err)
	}

	taskJSON := extractText(taskMessage)
	fmt.Println("原始 JSON 输出:")
	fmt.Println(taskJSON)

	// 解析任务列表
	var taskList TaskList
	if err := json.Unmarshal([]byte(taskJSON), &taskList); err != nil {
		log.Fatalf("任务 JSON 解析失败: %v\n原始内容: %s", err, taskJSON)
	}

	// 以表格形式展示解析后的任务
	fmt.Println("\n解析后的任务列表:")
	fmt.Printf("  %-5s %-30s %-10s %-12s %s\n", "序号", "任务", "优先级", "截止日期", "分类")
	fmt.Println("  " + fmt.Sprintf("%s", "-------------------------------------------------------------"))
	for i, task := range taskList.Tasks {
		// 根据优先级设置显示标记
		priorityMark := " "
		switch task.Priority {
		case "high":
			priorityMark = "[!]"
		case "medium":
			priorityMark = "[-]"
		case "low":
			priorityMark = "[ ]"
		}
		fmt.Printf("  %-5d %-30s %s %-7s %-12s %s\n",
			i+1, task.Title, priorityMark, task.Priority, task.Deadline, task.Category)
	}

	// ========================================
	// 示例 3：健壮的 JSON 解析（处理 Claude 输出的不确定性）
	// ========================================
	// 有时 Claude 可能会在 JSON 前后加上说明文字或 markdown 标记
	// 这里演示如何做更健壮的解析
	fmt.Println("\n=== 示例 3：健壮的 JSON 提取 ===")

	// 即使提示词要求"只输出 JSON"，Claude 偶尔仍可能添加额外文字
	// 下面这个辅助函数可以从混杂文字中提取 JSON
	rawText := `这是分析结果：{"name": "测试", "score": 95}希望对你有帮助！`
	fmt.Printf("原始文本: %s\n", rawText)
	extracted := extractJSON(rawText)
	fmt.Printf("提取的 JSON: %s\n", extracted)

	// 验证提取的 JSON 是否合法
	var testData map[string]interface{}
	if err := json.Unmarshal([]byte(extracted), &testData); err != nil {
		fmt.Printf("JSON 解析失败: %v\n", err)
	} else {
		fmt.Printf("解析成功: %v\n", testData)
	}

	// 总结
	fmt.Println("\n=== 结构化输出技巧总结 ===")
	fmt.Println("1. 在系统提示词中明确要求'只输出纯 JSON，不要包含其他文字'")
	fmt.Println("2. 提供 JSON schema 作为模板，让 Claude 严格遵循")
	fmt.Println("3. 指定字段的数据类型（string、number、boolean、array）")
	fmt.Println("4. 在 Go 端提前定义好对应的结构体（带 json tag）")
	fmt.Println("5. 做好错误处理：Claude 的输出不保证 100% 是合法 JSON")
	fmt.Println("6. 考虑添加 JSON 提取逻辑，处理 Claude 偶尔添加的额外文字")
}

// extractJSON 从可能包含非 JSON 内容的字符串中提取 JSON 部分
// 策略：查找第一个 '{' 和最后一个 '}' 之间的内容
// 这是一个简化的实现，适用于单个 JSON 对象的情况
func extractJSON(text string) string {
	// 查找第一个 '{' 的位置
	start := -1
	for i, c := range text {
		if c == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return text // 没找到 '{'，返回原文
	}

	// 从后向前查找最后一个 '}' 的位置
	end := -1
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '}' {
			end = i
			break
		}
	}
	if end == -1 || end <= start {
		return text // 没找到匹配的 '}'，返回原文
	}

	return text[start : end+1]
}
