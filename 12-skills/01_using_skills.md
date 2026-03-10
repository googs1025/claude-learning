# 使用 Skills

## 内置 Skills

Claude Code 自带一些常用的 Skills，可以直接通过斜杠命令调用：

| 命令 | 说明 |
|------|------|
| `/commit` | 生成规范的 git commit message 并提交 |
| `/review-pr` | 审查当前 PR 的代码变更 |
| `/simplify` | 简化选中的代码，减少复杂度 |

### 使用示例

```bash
# 在交互式 claude 中使用
claude

# 然后输入斜杠命令
> /commit
> /review-pr 123
> /simplify src/utils.go
```

## 查看可用 Skills

在 Claude Code 交互模式中，输入 `/` 后会显示所有可用的 skills 列表，包括内置和自定义的。

## 社区 Skills

社区贡献的 Skills 可以从以下途径获取：

1. **GitHub**: 搜索 `claude-code-skills` 相关仓库
2. **官方文档**: https://docs.anthropic.com/en/docs/claude-code/skills

### 安装社区 Skill

```bash
# 克隆或下载 skill 目录
git clone https://github.com/example/awesome-claude-skills.git

# 复制到项目级或用户级目录
cp -r awesome-claude-skills/some-skill ~/.claude/skills/
```

## 调用自定义 Skills

### 交互模式

```bash
# 启动 claude
claude

# 调用自定义 skill（已安装到 .claude/skills/）
> /code-review src/handler.go
> /go-test-generator pkg/utils/string.go
> /safe-explorer
```

### 非交互模式（CLI）

```bash
# 目前 -p 模式不直接支持 /skill-name 语法
# 可以通过 --skill-prompt 或在 prompt 中引用 skill 的内容

# 方式 1：直接在 prompt 中描述 skill 的行为
claude -p "按照代码审查清单审查 src/handler.go" --model sonnet

# 方式 2：将 SKILL.md 内容作为 system prompt 的一部分
# 参见 02_invoke_skill.go 的实现
```

## Skills 存储结构

```
~/.claude/skills/              # 用户级（全局）
├── code-review/
│   └── SKILL.md
└── my-custom-skill/
    └── SKILL.md

project/.claude/skills/        # 项目级
├── team-review/
│   └── SKILL.md
└── deploy-check/
    ├── SKILL.md
    └── checklist.md           # 辅助文件
```

## 优先级

当同名 skill 同时存在于用户级和项目级时，**项目级优先**。这允许团队为特定项目定制 skill 行为。
