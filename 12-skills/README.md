# 第12章：Skills 技能扩展

## 概述

Skills 是 Claude Code 的可复用提示模板系统，让你把常用的工作流程打包为可调用的"技能"。通过 Skills，你可以标准化团队的代码审查流程、自动生成文档、限制 Claude 的工具使用范围等。

> **前置知识**: 本章假设你已了解第 9 章的 Claude Code 基础和第 11 章的 CLI 高级功能。

## 核心概念

### Skills 是什么？

Skills 本质上是带有 YAML frontmatter 的 Markdown 文件（`SKILL.md`），定义了：
- 一段提示词（Claude 的行为指令）
- 元数据（名称、描述、工具限制等）
- 可选的辅助文件引用

### Skills vs CLAUDE.md

| 特性 | Skills | CLAUDE.md |
|------|--------|-----------|
| 触发方式 | 用户主动调用（`/skill-name`） | 自动加载 |
| 作用范围 | 特定任务 | 整个会话 |
| 工具限制 | 可指定 `allowed-tools` | 不能 |
| 参数传递 | 支持 `$ARGUMENTS` | 不支持 |
| 存储位置 | `skills/` 目录 | 项目根或 `~/.claude/` |

### YAML Frontmatter 字段

```yaml
---
name: skill-name               # 技能名称（用于 /skill-name 调用）
description: 简要描述           # 显示在技能列表中
argument-hint: "[filename]"    # 参数提示（显示在 UI 中）
user-invocable: true           # 是否可被用户手动调用（默认 true）
disable-model-invocation: true # 禁止 Claude 自动调用此 skill
allowed-tools:                 # 允许使用的工具列表
  - Read
  - Grep
  - Glob
model: sonnet                  # 指定使用的模型
---
```

### 存储位置

| 位置 | 作用范围 | 路径 |
|------|----------|------|
| 项目级 | 当前项目 | `.claude/skills/skill-name/SKILL.md` |
| 用户级 | 所有项目 | `~/.claude/skills/skill-name/SKILL.md` |

### 字符串替换

| 变量 | 含义 | 示例 |
|------|------|------|
| `$ARGUMENTS` | 用户传入的参数 | `/review src/main.go` → `$ARGUMENTS` = `src/main.go` |
| `$0` | 同 `$ARGUMENTS` | 与 `$ARGUMENTS` 等价 |
| `${CLAUDE_SKILL_DIR}` | SKILL.md 所在目录 | 引用同目录下的辅助文件 |

## 文件说明

| 文件 | 主题 | 关键知识点 |
|------|------|-----------|
| `01_using_skills.md` | 使用现有 Skills | 内置 Skills、社区 Skills、调用方式 |
| `02_invoke_skill.go` | 通过 CLI 调用 Skill | Go 程序自动化调用 skill |
| `03_test_skills.go` | 测试自定义 Skill | 验证 SKILL.md 格式和功能 |
| `skills/code-review/` | 代码审查 Skill | 结构化审查清单 + 评分 |
| `skills/go-test-generator/` | 测试生成 Skill | `$ARGUMENTS` 参数传递 |
| `skills/safe-explorer/` | 安全浏览 Skill | `allowed-tools` 只读限制 |
| `skills/doc-generator/` | 文档生成 Skill | 辅助文件引用 |

## 运行方式

```bash
cd 12-skills

# 查看自定义 skill 示例
cat skills/code-review/SKILL.md

# 通过 Go 程序调用 skill
go run 02_invoke_skill.go

# 测试自定义 skill
go run 03_test_skills.go
```

## 安装自定义 Skills

```bash
# 方式 1：复制到项目级目录
mkdir -p .claude/skills
cp -r skills/code-review .claude/skills/

# 方式 2：复制到用户级目录（全局可用）
cp -r skills/code-review ~/.claude/skills/

# 方式 3：在交互式 Claude 中使用
# 进入 claude 交互模式后输入 /code-review src/main.go
```

## 学习要点

- Skills 是带 YAML frontmatter 的 Markdown 文件，存放在 `skills/` 目录
- `user-invocable: true` 的 skill 可通过 `/skill-name` 手动调用
- `disable-model-invocation: true` 防止 Claude 自动触发，需用户明确调用
- `allowed-tools` 限制 skill 执行期间可使用的工具，构建安全沙箱
- `$ARGUMENTS` 和 `${CLAUDE_SKILL_DIR}` 让 skill 模板更灵活
- 将团队约定封装为 Skills，比口头约定更可靠
