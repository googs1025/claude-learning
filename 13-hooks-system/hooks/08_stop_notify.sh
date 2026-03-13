#!/bin/bash
# Hook: 08_stop_notify.sh
# 事件: Stop
# 作用: Claude 会话结束后发送 macOS 桌面通知，并显示耗时摘要
#       适合长时间运行的任务：启动后去做其他事，完成时自动通知

# 从 stdin 读取 JSON（Stop 事件包含会话摘要）
INPUT=$(cat)

# 提取会话信息（Stop 事件的 JSON 结构）
SESSION_COST=$(echo "$INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
cost = d.get('usage', {}).get('total_cost_usd', 0)
print(f'\${cost:.4f}' if cost else '未知')
" 2>/dev/null || echo "未知")

TOOL_CALLS=$(echo "$INPUT" | python3 -c "
import sys, json
d = json.load(sys.stdin)
# 尝试获取工具调用次数
stats = d.get('stats', {})
print(stats.get('tool_calls', '未知'))
" 2>/dev/null || echo "未知")

TIMESTAMP=$(date '+%H:%M:%S')
AUDIT_LOG="/tmp/claude_hooks_audit.log"
echo "[$(date '+%Y-%m-%d %H:%M:%S')] [stop_notify] 会话结束" >> "$AUDIT_LOG"

# macOS 桌面通知（需要 macOS）
if command -v osascript &>/dev/null; then
    NOTIFY_TITLE="Claude Code 已完成"
    NOTIFY_MSG="任务执行完毕 ✅\n时间: $TIMESTAMP\n费用: $SESSION_COST"

    osascript -e "display notification \"任务执行完毕 ✅  时间: $TIMESTAMP  费用: $SESSION_COST\" with title \"Claude Code 已完成\" sound name \"Glass\"" 2>/dev/null
fi

# Linux 通知（如果有 notify-send）
if command -v notify-send &>/dev/null; then
    notify-send "Claude Code 已完成" "任务执行完毕 ✅\n时间: $TIMESTAMP" 2>/dev/null
fi

# 终端输出摘要
echo "[stop_notify] 🎉 Claude 会话已结束" >&2
echo "[stop_notify] 完成时间: $TIMESTAMP" >&2
echo "[stop_notify] 本次费用: $SESSION_COST" >&2

exit 0
