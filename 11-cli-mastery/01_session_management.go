//go:build ignore

// 第十一章示例 1：会话管理
// 运行方式: go run 01_session_management.go
//
// 本示例演示如何：
// 1. 使用 --output-format json 获取 session_id
// 2. 使用 --session-id 恢复指定会话
// 3. 使用 --continue 继续最近一次会话
//
// 会话管理让你能够在多次 CLI 调用之间保持上下文连贯

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// CLIResponse 表示 claude CLI 的 JSON 响应
// 使用 --output-format json 时返回此格式
type CLIResponse struct {
	Type       string `json:"type"`        // "result"
	Result     string `json:"result"`      // Claude 的文本回复
	SessionID  string `json:"session_id"`  // 会话 ID（UUID 格式）
	TotalCost  float64 `json:"total_cost_usd"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func main() {
	// ========== 步骤 1：发起新会话，获取 session_id ==========
	fmt.Println("=== 步骤 1：发起新会话 ===")

	firstPrompt := "我正在学习 Go 语言的并发编程。请简要介绍 goroutine 的概念。"

	cmd := exec.Command("claude",
		"-p", firstPrompt,
		"--output-format", "json",
		"--model", "sonnet",
	)

	// 使用 Output() 而非 CombinedOutput()，确保 stderr 不混入 JSON
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("第一次调用失败: %v", err)
	}

	// 解析 JSON 响应，提取 session_id
	var firstResponse CLIResponse
	if err := json.Unmarshal(output, &firstResponse); err != nil {
		log.Fatalf("解析 JSON 失败: %v\n原始输出: %s", err, string(output))
	}

	sessionID := firstResponse.SessionID
	fmt.Printf("回复: %s\n", truncate(firstResponse.Result, 100))
	fmt.Printf("Session ID: %s\n\n", sessionID)

	// ========== 步骤 2：使用 --session-id 恢复指定会话 ==========
	fmt.Println("=== 步骤 2：使用 --session-id 恢复会话 ===")

	// Claude 会记住上一轮的上下文（goroutine 的讨论）
	followUpPrompt := "接着上面的内容，channel 和 goroutine 如何配合使用？请举一个简单例子。"

	cmd2 := exec.Command("claude",
		"-p", followUpPrompt,
		"--output-format", "json",
		"--session-id", sessionID, // 关键：通过 session_id 恢复上下文
		"--model", "sonnet",
	)

	output2, err := cmd2.Output()
	if err != nil {
		log.Fatalf("第二次调用失败: %v", err)
	}

	var secondResponse CLIResponse
	if err := json.Unmarshal(output2, &secondResponse); err != nil {
		log.Fatalf("解析第二次响应失败: %v", err)
	}

	fmt.Printf("回复: %s\n", truncate(secondResponse.Result, 200))
	fmt.Printf("同一会话: %v\n\n", secondResponse.SessionID == sessionID)

	// ========== 步骤 3：使用 --continue 继续最近会话 ==========
	fmt.Println("=== 步骤 3：使用 --continue 继续最近会话 ===")
	fmt.Println("提示: --continue 会自动恢复最近一次会话，无需手动指定 session_id")
	fmt.Println("适合在交互式终端中快速续接上下文")
	fmt.Println()

	// --continue 等价于自动使用最近的 session_id
	// 适合交互式场景，脚本中建议使用明确的 --session-id
	cmd3 := exec.Command("claude",
		"-p", "总结一下我们刚才讨论的 Go 并发要点。",
		"--output-format", "json",
		"--continue", // 自动续接最近会话
		"--model", "sonnet",
	)

	output3, err := cmd3.Output()
	if err != nil {
		log.Fatalf("第三次调用失败: %v", err)
	}

	var thirdResponse CLIResponse
	if err := json.Unmarshal(output3, &thirdResponse); err != nil {
		log.Fatalf("解析第三次响应失败: %v", err)
	}

	fmt.Printf("回复: %s\n\n", truncate(thirdResponse.Result, 200))

	// ========== 总结 ==========
	fmt.Println("=== 会话管理总结 ===")
	fmt.Println("--session-id <uuid>  : 恢复指定会话（脚本推荐）")
	fmt.Println("--continue           : 继续最近会话（交互推荐）")
	fmt.Println("--resume             : 同 --continue 的别名")
}

// truncate 截断字符串到指定长度
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
