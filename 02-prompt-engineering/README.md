# 第二章：提示词工程 (Prompt Engineering)

## 概述

提示词工程是与大语言模型（LLM）交互的核心技能。通过精心设计提示词，
我们可以显著提升 Claude 的输出质量、准确性和可用性。

本章通过 Go 代码示例，系统地介绍提示词工程的关键技术和最佳实践。

## 核心原则

### 1. 清晰明确 (Clarity)

- **具体而非模糊**：用精确的指令替代笼统的描述
- **提供约束条件**：明确输出格式、长度、语言等要求
- **避免歧义**：确保指令只有一种合理的解读方式

```
❌ "帮我写点东西"
✅ "请用中文写一段 200 字以内的产品描述，产品是一款智能手表，目标用户是年轻白领"
```

### 2. 示例驱动 (Few-shot Prompting)

- 通过提供 2-3 个输入/输出示例，让 Claude 理解预期的行为模式
- 示例比抽象的规则描述更有效
- 在 API 调用中，使用 user/assistant 消息对来构造示例

```
用户: "这个产品太棒了！" → 助手: "正面"
用户: "质量很差，不推荐" → 助手: "负面"
用户: "还行吧，一般般" → ?
```

### 3. 结构化输出 (Structured Output)

- 通过系统提示词指定输出格式（如 JSON、Markdown 表格等）
- 提供输出的 schema 或模板，让 Claude 严格遵循
- 便于程序解析和后续处理

### 4. 思维链 (Chain of Thought)

- 要求 Claude "一步一步思考"，可以显著提升推理任务的准确率
- 适用于数学计算、逻辑推理、代码分析等需要多步骤推理的场景
- 通过暴露中间推理步骤，可以验证和调试 Claude 的思考过程

### 5. 角色设定 (Role Prompting)

- 通过系统提示词为 Claude 指定角色和专业领域
- 不同的角色会产生不同风格和深度的回答
- 可以组合角色 + 约束条件来精确控制输出

### 6. 提示词链 (Prompt Chaining)

- 将复杂任务分解为多个简单步骤
- 每一步的输出作为下一步的输入
- 提高复杂任务的可靠性和可控性

## 示例文件

| 文件 | 主题 | 说明 |
|------|------|------|
| `01_few_shot.go` | Few-shot 提示 | 通过示例教会 Claude 执行情感分析 |
| `02_chain_of_thought.go` | 思维链提示 | 引导 Claude 逐步推理数学/逻辑问题 |
| `03_role_prompting.go` | 角色提示 | 通过系统提示词设定不同专家角色 |
| `04_structured_output.go` | 结构化输出 | 要求 Claude 输出可解析的 JSON 格式 |
| `05_prompt_chaining.go` | 提示词链 | 多步骤 API 调用实现复杂任务流水线 |

## 运行方式

```bash
# 确保已设置 API 密钥
export ANTHROPIC_API_KEY="your-api-key"

# 安装依赖
cd /path/to/claude-learning
go mod tidy

# 运行各示例
cd 02-prompt-engineering
go run 01_few_shot.go
go run 02_chain_of_thought.go
go run 03_role_prompting.go
go run 04_structured_output.go
go run 05_prompt_chaining.go
```

## 提示词工程最佳实践

1. **迭代优化**：第一版提示词很少是最优的，持续测试和改进
2. **从简单开始**：先写最直接的提示词，再根据需要添加约束和示例
3. **测试边界情况**：用各种输入测试提示词的鲁棒性
4. **记录有效模式**：建立自己的提示词模板库
5. **量化评估**：对重要的提示词进行系统化评估，而不只是凭感觉判断
