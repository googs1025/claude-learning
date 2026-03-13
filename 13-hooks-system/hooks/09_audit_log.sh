#!/bin/bash
# Hook: 09_audit_log.sh
# 事件: PostToolUse (Bash)
# 作用: 记录 Claude 执行的所有 Bash 命令到审计日志
#       追踪：谁、何时、在哪个目录、执行了什么命令、退出码是多少

# 从 stdin 读取 Claude Code 传入的 JSON
INPUT=$(cat)

# 提取信息
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)

# 仅处理 Bash 工具
if [ "$TOOL_NAME" != "Bash" ]; then
    exit 0
fi

# 提取命令和退出码
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('command',''))" 2>/dev/null)
EXIT_CODE=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_response',{}).get('exit_code','?'))" 2>/dev/null)

# 审计日志文件
AUDIT_LOG="/tmp/claude_audit.log"

# 确保日志文件存在并设置权限
touch "$AUDIT_LOG" 2>/dev/null

# 写入结构化审计日志
cat >> "$AUDIT_LOG" <<EOF
---
时间: $(date '+%Y-%m-%d %H:%M:%S')
目录: $(pwd)
用户: $(whoami)
命令: $COMMAND
退出码: $EXIT_CODE
EOF

# 日志轮转：超过 10MB 时清理旧日志
LOG_SIZE=$(du -k "$AUDIT_LOG" 2>/dev/null | awk '{print $1}')
if [ "${LOG_SIZE:-0}" -gt 10240 ] 2>/dev/null; then
    # 保留最新的 1000 行
    tail -1000 "$AUDIT_LOG" > "${AUDIT_LOG}.tmp" && mv "${AUDIT_LOG}.tmp" "$AUDIT_LOG"
    echo "--- 日志轮转: $(date '+%Y-%m-%d %H:%M:%S') ---" >> "$AUDIT_LOG"
fi

exit 0
