//go:build ignore

// 第十一章示例 4：预算控制
// 运行方式: go run 04_budget_control.go
//
// 本示例演示如何：
// 1. 使用 --max-turns 限制 Agent 循环次数
// 2. 使用 --max-budget-usd 设置费用上限
// 3. 结合 JSON 输出监控实际花费
//
// 预算控制在自动化和 CI/CD 场景中至关重要，防止 Agent 无限循环或费用失控

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// BudgetResponse 用于监控费用
type BudgetResponse struct {
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	Result       string  `json:"result"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func main() {
	// ========== 场景 1：限制 Agent 循环次数 ==========
	fmt.Println("=== 场景 1：--max-turns 限制 Agent 循环 ===")
	fmt.Println("max-turns 限制 Claude 的 agentic 轮次")
	fmt.Println("适合：防止复杂任务无限递归")
	fmt.Println()

	cmd1 := exec.Command("claude",
		"-p", "列出当前目录下的文件并简要说明每个文件的作用",
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "3", // 最多 3 轮 Agent 循环
	)

	resp1 := runWithBudget(cmd1)
	fmt.Printf("回复: %s\n", truncateStr(resp1.Result, 150))
	fmt.Printf("费用: $%.6f\n\n", resp1.TotalCostUSD)

	// ========== 场景 2：设置费用上限 ==========
	fmt.Println("=== 场景 2：--max-budget-usd 费用上限 ===")
	fmt.Println("max-budget-usd 设置单次调用的最大费用（美元）")
	fmt.Println("达到上限后 Claude 会停止执行并返回已有结果")
	fmt.Println()

	cmd2 := exec.Command("claude",
		"-p", "请简要解释 Go 的 interface 概念",
		"--output-format", "json",
		"--model", "sonnet",
		"--max-budget-usd", "0.05", // 最多花费 $0.05
	)

	resp2 := runWithBudget(cmd2)
	fmt.Printf("回复: %s\n", truncateStr(resp2.Result, 150))
	fmt.Printf("费用: $%.6f\n\n", resp2.TotalCostUSD)

	// ========== 场景 3：组合使用 ==========
	fmt.Println("=== 场景 3：组合使用多种限制 ===")
	fmt.Println("同时设置 max-turns 和 max-budget-usd，任一条件触发即停止")
	fmt.Println()

	cmd3 := exec.Command("claude",
		"-p", "分析当前目录的代码结构",
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "2",         // 最多 2 轮
		"--max-budget-usd", "0.03", // 最多 $0.03
	)

	resp3 := runWithBudget(cmd3)
	fmt.Printf("回复: %s\n", truncateStr(resp3.Result, 150))
	fmt.Printf("费用: $%.6f\n", resp3.TotalCostUSD)
	fmt.Printf("Token: 输入=%d, 输出=%d\n\n", resp3.Usage.InputTokens, resp3.Usage.OutputTokens)

	// ========== 总结 ==========
	fmt.Println("=== 预算控制参数总结 ===")
	fmt.Println("--max-turns N        : 限制 Agent 最大轮次")
	fmt.Println("--max-budget-usd X   : 限制单次最大费用（美元）")
	fmt.Println()
	fmt.Println("建议：")
	fmt.Println("  CI/CD 场景: --max-turns 5 --max-budget-usd 0.10")
	fmt.Println("  代码审查:   --max-turns 3 --max-budget-usd 0.05")
	fmt.Println("  快速问答:   --max-turns 1 --max-budget-usd 0.02")
}

// runWithBudget 执行命令并返回解析后的响应
func runWithBudget(cmd *exec.Cmd) BudgetResponse {
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	var resp BudgetResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析 JSON 失败: %v\n原始输出: %s", err, string(output))
	}

	return resp
}

// truncateStr 截断字符串
func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
