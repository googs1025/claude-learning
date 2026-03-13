#!/bin/bash
# Hook: 07_session_check.sh
# 事件: SessionStart
# 作用: 每次 Claude 会话启动时，自动做环境健康体检
#
# 输出渠道：
#   1. stdout      → 注入 Claude 上下文（Claude 感知当前状态）
#   2. 日志文件    → /tmp/claude_session_health.txt（随时可查）
#   3. macOS 通知  → 桌面弹窗显示关键状态（用户可见）

INPUT=$(cat)

LOG_FILE="/tmp/claude_session_health.txt"
REPORT=""
NOTIFY_PARTS=()  # 通知摘要

# log: 追加到报告字符串
log() { REPORT="${REPORT}${*}\n"; }

log "=== 🔍 会话启动健康体检 ==="
log "时间: $(date '+%Y-%m-%d %H:%M:%S')"
log ""

# 1. Git 状态检查
if command -v git &>/dev/null && git rev-parse --is-inside-work-tree &>/dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null || echo "detached HEAD")
    UNCOMMITTED=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')
    UNPUSHED=$(git log @{u}.. --oneline 2>/dev/null | wc -l | tr -d ' ' || echo "0")

    log "📁 Git 状态:"
    log "  分支: $BRANCH"
    log "  未提交文件: $UNCOMMITTED 个"
    if [ "$UNPUSHED" -gt 0 ] 2>/dev/null; then
        log "  未推送提交: $UNPUSHED 个 ⚠️"
        NOTIFY_PARTS+=("Git:$BRANCH ⚠️未推送$UNPUSHED")
    else
        NOTIFY_PARTS+=("Git:$BRANCH 未提交$UNCOMMITTED")
    fi
    log ""
fi

# 2. Go 环境检查
if command -v go &>/dev/null; then
    GO_VERSION=$(go version 2>/dev/null | awk '{print $3}')
    log "🐹 Go 环境: $GO_VERSION"
    if [ -f "go.mod" ]; then
        MODULE=$(head -1 go.mod | awk '{print $2}')
        log "  模块: $MODULE"
    fi
    NOTIFY_PARTS+=("Go:$GO_VERSION")
    log ""
fi

# 3. 环境变量检查
log "🔑 环境变量:"
if [ -n "$ANTHROPIC_API_KEY" ]; then
    KEY_PREFIX="${ANTHROPIC_API_KEY:0:8}"
    log "  ANTHROPIC_API_KEY: ${KEY_PREFIX}... ✅"
    NOTIFY_PARTS+=("API:✅")
else
    log "  ANTHROPIC_API_KEY: 未设置 ⚠️（API 示例需要）"
    NOTIFY_PARTS+=("API:⚠️未设置")
fi
log ""

# 4. 磁盘空间检查
DISK_USAGE=$(df -h . 2>/dev/null | awk 'NR==2 {print $5}' | tr -d '%')
if [ -n "$DISK_USAGE" ] && [ "$DISK_USAGE" -gt 85 ] 2>/dev/null; then
    log "💾 磁盘空间: ${DISK_USAGE}% 已用 ⚠️ 磁盘接近满，注意清理"
    NOTIFY_PARTS+=("磁盘:${DISK_USAGE}%⚠️")
    log ""
fi

# 5. 上次备份时间
if [ -d "/tmp/claude_backup" ]; then
    LAST_BACKUP=$(ls -t /tmp/claude_backup/ 2>/dev/null | head -1)
    if [ -n "$LAST_BACKUP" ]; then
        log "💾 最近备份目录: /tmp/claude_backup/$LAST_BACKUP"
        log ""
    fi
fi

log "=== 体检完成，开始工作 ==="
log ""
log "📄 完整报告: $LOG_FILE"

# ── 输出到各渠道 ──────────────────────────────────────────

# 1. stdout → Claude 上下文
printf "%b" "$REPORT"

# 2. 写入日志文件
printf "%b" "$REPORT" > "$LOG_FILE"

# 3. macOS 桌面通知（拼接摘要）
NOTIFY_BODY=$(IFS=" | "; echo "${NOTIFY_PARTS[*]}")
osascript << APPLESCRIPT 2>/dev/null || true
display notification "$NOTIFY_BODY" with title "🔍 Claude Code 健康体检" sound name "Ping"
APPLESCRIPT