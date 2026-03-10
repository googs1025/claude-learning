//go:build ignore

// 第十二章示例 3：测试自定义 Skills
// 运行方式: go run 03_test_skills.go
//
// 本示例演示如何：
// 1. 验证 SKILL.md 文件的 YAML frontmatter 格式
// 2. 检查必要字段是否存在
// 3. 通过实际调用测试 skill 的行为
//
// 在发布 skill 之前，务必验证格式和功能的正确性

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

// SkillMeta 表示 SKILL.md 的 frontmatter 元数据
type SkillMeta struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	ArgumentHint           string   `json:"argument-hint"`
	UserInvocable          bool     `json:"user-invocable"`
	DisableModelInvocation bool     `json:"disable-model-invocation"`
	AllowedTools           []string `json:"allowed-tools"`
}

// TestResult 记录单个 skill 的测试结果
type TestResult struct {
	SkillName string
	Passed    bool
	Messages  []string
}

func main() {
	// 获取 skills 目录
	_, currentFile, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(currentFile)
	skillsDir := filepath.Join(baseDir, "skills")

	fmt.Println("=== Skills 测试工具 ===")
	fmt.Println()

	// 发现所有 skill
	skills := discoverSkills(skillsDir)
	fmt.Printf("发现 %d 个 skills\n\n", len(skills))

	// 测试每个 skill
	var results []TestResult
	for _, skillPath := range skills {
		result := testSkill(skillPath)
		results = append(results, result)
	}

	// 输出测试报告
	fmt.Println("\n=== 测试报告 ===")
	passed := 0
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		} else {
			passed++
		}
		fmt.Printf("[%s] %s\n", status, r.SkillName)
		for _, msg := range r.Messages {
			fmt.Printf("      %s\n", msg)
		}
	}
	fmt.Printf("\n总计: %d/%d 通过\n", passed, len(results))

	// ========== 功能测试：实际调用一个 skill ==========
	fmt.Println("\n=== 功能测试：实际调用 safe-explorer skill ===")

	safeExplorerPath := filepath.Join(skillsDir, "safe-explorer", "SKILL.md")
	content, err := os.ReadFile(safeExplorerPath)
	if err != nil {
		log.Fatalf("读取 safe-explorer SKILL.md 失败: %v", err)
	}

	body := extractSkillBody(string(content))
	cmd := exec.Command("claude",
		"-p", body+"\n\n请描述当前目录的文件结构",
		"--output-format", "json",
		"--model", "sonnet",
		"--max-turns", "2",
		"--allowedTools", "Read", "Grep", "Glob", // safe-explorer 的工具限制
	)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("功能测试失败: %v", err)
	}

	var resp struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &resp); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	result := []rune(resp.Result)
	if len(result) > 200 {
		fmt.Printf("safe-explorer 回复: %s...\n", string(result[:200]))
	} else {
		fmt.Printf("safe-explorer 回复: %s\n", resp.Result)
	}
	fmt.Println("\n功能测试通过！skill 在只读模式下正常工作")
}

// discoverSkills 扫描目录发现所有 SKILL.md 文件
func discoverSkills(dir string) []string {
	var skills []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("读取 skills 目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skills = append(skills, skillFile)
			}
		}
	}

	return skills
}

// testSkill 验证单个 SKILL.md 文件
func testSkill(path string) TestResult {
	skillName := filepath.Base(filepath.Dir(path))
	result := TestResult{SkillName: skillName, Passed: true}

	// 读取文件
	content, err := os.ReadFile(path)
	if err != nil {
		result.Passed = false
		result.Messages = append(result.Messages, fmt.Sprintf("无法读取文件: %v", err))
		return result
	}

	text := string(content)

	// 检查 1：是否有 YAML frontmatter
	if !strings.HasPrefix(text, "---") {
		result.Passed = false
		result.Messages = append(result.Messages, "缺少 YAML frontmatter（应以 --- 开头）")
		return result
	}

	// 检查 2：frontmatter 是否正确闭合
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		result.Passed = false
		result.Messages = append(result.Messages, "YAML frontmatter 未正确闭合（需要两个 ---）")
		return result
	}

	frontmatter := parts[1]
	body := strings.TrimSpace(parts[2])

	// 检查 3：必须有 name 字段
	if !strings.Contains(frontmatter, "name:") {
		result.Passed = false
		result.Messages = append(result.Messages, "缺少 name 字段")
	}

	// 检查 4：必须有 description 字段
	if !strings.Contains(frontmatter, "description:") {
		result.Passed = false
		result.Messages = append(result.Messages, "缺少 description 字段")
	}

	// 检查 5：正文不能为空
	if body == "" {
		result.Passed = false
		result.Messages = append(result.Messages, "正文内容为空")
	}

	// 检查 6：文件不超过 500 行
	lines := strings.Count(text, "\n") + 1
	if lines > 500 {
		result.Passed = false
		result.Messages = append(result.Messages, fmt.Sprintf("超过 500 行限制（当前 %d 行）", lines))
	}

	if result.Passed {
		result.Messages = append(result.Messages, fmt.Sprintf("格式正确（%d 行）", lines))
	}

	return result
}

// extractSkillBody 提取 frontmatter 之后的正文
func extractSkillBody(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(content)
}
