//go:build ignore

// 第十二章示例 2：通过 CLI 调用 Skill
// 运行方式: go run 02_invoke_skill.go
//
// 本示例演示如何：
// 1. 读取 SKILL.md 文件内容作为提示词
// 2. 通过 claude CLI 执行 skill 定义的任务
// 3. 将 skill 的 allowed-tools 转换为 CLI 参数
//
// 这种方式适合在 CI/CD 或自动化脚本中使用自定义 Skills

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// SkillResponse 用于解析 CLI JSON 输出
type SkillResponse struct {
	Type         string  `json:"type"`
	Result       string  `json:"result"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

func main() {
	// 获取当前文件所在目录
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)

	// ========== 示例 1：调用 code-review skill ==========
	fmt.Println("=== 示例 1：调用 code-review skill ===")
	fmt.Println("读取 SKILL.md 内容，作为 prompt 的一部分发送给 Claude")
	fmt.Println()

	// 读取 skill 定义
	skillPath := filepath.Join(baseDir, "skills", "code-review", "SKILL.md")
	skillContent, err := os.ReadFile(skillPath)
	if err != nil {
		log.Fatalf("读取 SKILL.md 失败: %v", err)
	}

	// 提取 frontmatter 之后的内容作为 system prompt
	body := extractBody(string(skillContent))

	// 构建完整 prompt：skill 指令 + 用户参数
	targetFile := filepath.Join(baseDir, "02_invoke_skill.go") // 审查自身作为示例
	prompt := fmt.Sprintf("%s\n\n请审查以下文件: %s", body, targetFile)

	cmd := exec.Command("claude",
		"-p", prompt,
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "3",
		"--allowedTools", "Read", "Grep", "Glob", // 模拟 skill 的 allowed-tools
	)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("调用 code-review skill 失败: %v", err)
	}

	var resp SkillResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	printResult("code-review", resp)

	// ========== 示例 2：调用带参数的 skill ==========
	fmt.Println("\n=== 示例 2：调用 go-test-generator skill ===")
	fmt.Println("使用 $ARGUMENTS 替换机制传入目标文件")
	fmt.Println()

	testGenSkillPath := filepath.Join(baseDir, "skills", "go-test-generator", "SKILL.md")
	testGenContent, err := os.ReadFile(testGenSkillPath)
	if err != nil {
		log.Fatalf("读取 go-test-generator SKILL.md 失败: %v", err)
	}

	// 模拟 $ARGUMENTS 替换
	testGenBody := extractBody(string(testGenContent))
	testGenBody = strings.ReplaceAll(testGenBody, "$ARGUMENTS", targetFile)

	cmd2 := exec.Command("claude",
		"-p", testGenBody,
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "3",
		"--allowedTools", "Read", "Grep", "Glob", "Write",
	)

	output2, err := cmd2.Output()
	if err != nil {
		log.Fatalf("调用 go-test-generator skill 失败: %v", err)
	}

	var resp2 SkillResponse
	if err := json.Unmarshal(output2, &resp2); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	printResult("go-test-generator", resp2)
}

// extractBody 提取 SKILL.md 中 frontmatter 之后的正文内容
// frontmatter 由 --- 分隔
func extractBody(content string) string {
	// 查找第二个 ---
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	// 如果没有 frontmatter，返回全部内容
	return strings.TrimSpace(content)
}

// printResult 打印 skill 执行结果
func printResult(skillName string, resp SkillResponse) {
	result := []rune(resp.Result)
	if len(result) > 300 {
		fmt.Printf("[%s] 结果:\n%s...\n", skillName, string(result[:300]))
	} else {
		fmt.Printf("[%s] 结果:\n%s\n", skillName, resp.Result)
	}
	fmt.Printf("[%s] 费用: $%.6f\n", skillName, resp.TotalCostUSD)
}
