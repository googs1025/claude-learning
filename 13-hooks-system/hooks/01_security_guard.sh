#!/bin/bash
# Hook: 01_security_guard.sh
# 事件: PreToolUse (Bash)
# 作用: 拦截危险的 Bash 命令（rm -rf /、sudo rm、dd 等）
#
# 工作原理：
#   Claude Code 通过 stdin 传入 JSON，包含工具名和参数
#   hook 读取 stdin → 解析 JSON → 检测危险模式 → 决策

# 从 stdin 读取 Claude Code 传入的 JSON
INPUT=$(cat)

# 提取工具名和命令
TOOL_NAME=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_name',''))" 2>/dev/null)
COMMAND=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('command',''))" 2>/dev/null)

# 仅处理 Bash 工具
if [ "$TOOL_NAME" != "Bash" ]; then
    exit 0
fi

# 记录到审计日志
AUDIT_LOG="/tmp/claude_hooks_audit.log"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] [security_guard] 检查命令: $COMMAND" >> "$AUDIT_LOG"

# 危险命令模式列表
DANGEROUS_PATTERNS=(
    "rm -rf /"
    "rm -rf \*"
    "sudo rm"
    "dd if=/dev/zero"
    "dd if=/dev/random"
    "> /dev/sda"
    "mkfs\."
    "format "
    ":(){:|:&};:"           # Fork 炸弹
    "chmod -R 777 /"
    "chown -R .* /"
    "find / -delete"
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qE "$pattern" 2>/dev/null; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] [security_guard] 🚨 已阻断危险命令: $COMMAND" >> "$AUDIT_LOG"

        # 方式一：输出 JSON 决策（推荐，可携带详细原因）
        cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "安全守门人已阻断：检测到危险命令模式 '$pattern'。请确认操作必要性后重试。"
  }
}
EOF
        exit 0
    fi
done

# 命令安全，允许执行（不输出任何内容，静默通过）
exit 0
