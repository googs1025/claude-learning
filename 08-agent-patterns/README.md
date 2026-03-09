# 第8章：Agent 模式

## 概述

Agent（智能体）是基于大语言模型构建的自主系统，能够通过工具调用、推理和记忆来完成复杂任务。

## 核心设计模式

### 1. 简单 Agent
最基础的 Agent 模式：接收任务 → 思考 → 调用工具 → 返回结果。

### 2. ReAct 模式
**Re**asoning + **Act**ing 的结合：
```
思考(Thought) → 行动(Action) → 观察(Observation) → 思考 → ...
```
Agent 在每一步都明确表达推理过程，然后决定下一步行动。

### 3. 多 Agent 协作
将复杂任务拆分给多个专门化的 Agent：
- **规划 Agent**: 分解任务
- **执行 Agent**: 完成具体子任务
- **审查 Agent**: 验证结果

### 4. 带记忆的 Agent
通过维护对话历史和外部存储，让 Agent 具备长期记忆能力。

## 示例文件

| 文件 | 说明 |
|------|------|
| `01_simple_agent.go` | 简单 Agent：任务分解与执行 |
| `02_react_agent.go` | ReAct 模式：显式推理链 |
| `03_multi_agent.go` | 多 Agent 协作系统 |
| `04_agent_with_memory.go` | 带记忆的持久化 Agent |

## 关键概念

- **Agent 循环**: 持续调用 API 直到任务完成
- **工具编排**: Agent 自主决定调用哪些工具、以什么顺序
- **上下文管理**: 有效管理对话历史，避免超出 token 限制
- **错误恢复**: Agent 遇到错误时能自主调整策略
