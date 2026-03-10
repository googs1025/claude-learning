# 第11章：CLI 高级技巧

## 概述

本章深入探索 Claude CLI 的高级功能，涵盖会话管理、输出格式解析、工具限制、预算控制、Hooks 集成和多目录操作。这些技巧让你能够在自动化脚本和 CI/CD 流水线中更好地使用 Claude。

> **前置知识**: 本章假设你已了解第 0 章和第 9 章的 CLI 基础用法。

## 核心内容

### 功能一览

| 功能 | CLI 参数 | 典型场景 |
|------|----------|----------|
| 会话管理 | `--resume`, `--continue`, `--session-id` | 多轮对话、上下文续接 |
| 输出格式 | `--output-format json/stream-json/text` | 自动化解析、费用监控 |
| 工具限制 | `--allowedTools`, `--disallowedTools` | 只读审查、安全沙箱 |
| 预算控制 | `--max-turns`, `--max-budget-usd` | 成本管控、Agent 限制 |
| Hooks 集成 | `--settings` + hooks JSON | 安全验证、自动格式化 |
| 多目录 | `--add-dir` | Monorepo、跨项目分析 |

### 自动化流程

```
Go 程序 → 构建 CLI 参数 → exec.Command("claude", ...) → 解析 JSON 输出
                                    ↓
                            --output-format json  →  结构化结果
                            --session-id <uuid>   →  会话续接
                            --max-budget-usd 0.1  →  成本控制
```

## 文件说明

| 文件 | 主题 | 关键知识点 |
|------|------|-----------|
| `01_session_management.go` | 会话管理 | `--resume`, `--continue`, `--session-id` |
| `02_output_formats.go` | 输出格式解析 | text / json / stream-json 对比 |
| `03_tool_restrictions.go` | 工具限制 | `--allowedTools`, `--disallowedTools` |
| `04_budget_control.go` | 预算控制 | `--max-turns`, `--max-budget-usd` |
| `05_hooks_integration.go` | Hooks 集成 | Go 生成 hooks JSON + `--settings` |
| `06_multi_dir.go` | 多目录上下文 | `--add-dir` 跨目录分析 |

## 运行方式

```bash
cd 11-cli-mastery

# 运行单个示例
go run 01_session_management.go
go run 02_output_formats.go

# 注意：所有示例需要安装 claude CLI
# npm install -g @anthropic-ai/claude-code
```

## 学习要点

- `--output-format json` 配合 `cmd.Output()` 获取干净的 JSON（不混入 stderr）
- `--session-id` 可实现跨进程的会话续接
- `--allowedTools` 和 `--disallowedTools` 构建安全沙箱
- `--max-budget-usd` 在自动化场景中防止费用失控
- Hooks 可通过 `--settings` 临时加载，无需修改全局配置
- `--add-dir` 让 Claude 同时理解多个目录的上下文
