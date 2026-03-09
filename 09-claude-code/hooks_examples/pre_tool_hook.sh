#!/bin/bash
# Claude Code PreToolUse Hook 示例
# 此脚本在 Claude Code 调用工具前执行
#
# 用法：在 .claude/settings.json 中配置：
# {
#   "hooks": {
#     "PreToolUse": [{
#       "matcher": "Bash",
#       "hooks": [{"type": "command", "command": "./hooks/pre_tool_hook.sh"}]
#     }]
#   }
# }
#
# Hook 通过 stdin 接收 JSON 格式的工具调用信息
# 输出到 stdout 的内容会作为反馈返回给 Claude
# 退出码 0 = 允许执行, 非零 = 阻止执行

# 读取工具调用信息（JSON 格式通过 stdin 传入）
INPUT=$(cat)

# 提取工具名称和参数
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)
TOOL_INPUT=$(echo "$INPUT" | python3 -c "import sys,json; print(json.dumps(json.load(sys.stdin).get('tool_input',{})))" 2>/dev/null)

# 记录日志
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 工具调用: $TOOL_NAME" >> /tmp/claude_hook.log
echo "  参数: $TOOL_INPUT" >> /tmp/claude_hook.log

# 安全检查示例：阻止危险命令
if [ "$TOOL_NAME" = "Bash" ]; then
    COMMAND=$(echo "$TOOL_INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('command',''))" 2>/dev/null)

    # 检查是否包含危险命令
    if echo "$COMMAND" | grep -qE "(rm -rf /|sudo rm|drop table|format |mkfs)"; then
        echo "⚠️  阻止: 检测到危险命令 - $COMMAND"
        exit 1  # 非零退出码 = 阻止执行
    fi

    # 检查是否尝试修改系统文件
    if echo "$COMMAND" | grep -qE "(/etc/|/usr/|/System/)"; then
        echo "⚠️  警告: 该命令可能修改系统文件"
        # 退出码 0 仍然允许执行，但消息会显示给用户
    fi
fi

# 允许执行
exit 0
