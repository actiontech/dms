#!/usr/bin/env bash
# 提交前校验：社区版 / 试用版 / 企业版 / DMS 企业版 四种构建标签组合均能成功编译（与 Makefile install 一致）。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

readonly RED=$'\033[0;31m'
readonly GREEN=$'\033[0;32m'
readonly YELLOW=$'\033[1;33m'
readonly BLUE=$'\033[0;34m'
readonly CYAN=$'\033[0;36m'
readonly BOLD=$'\033[1m'
readonly DIM=$'\033[2m'
readonly NC=$'\033[0m'

bar() {
  printf '%b\n' "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

sub() {
  printf '%b %s\n' "${DIM}" "$1${NC}"
}

err() {
  printf '%b %s\n' "${RED}" "$1${NC}"
}

step_begin() {
  local idx="$1"
  local total="$2"
  local name="$3"
  local tags="$4"
  printf '\n'
  bar
  printf '%b Step %s/%s  %s\n' "${BOLD}" "$idx" "$total" "$name${NC}"
  printf '%b 构建标签 (GO_BUILD_TAGS): %s%s\n' "${CYAN}" "$tags" "${NC}"
  bar
  printf '%b ⏳ 正在执行 make install …%s\n' "${YELLOW}" "${NC}"
}

step_ok() {
  local secs="$1"
  printf '%b ✅ 本步编译成功%s' "${GREEN}" "${NC}"
  printf '%b （耗时 %ss）%s\n' "${DIM}" "$secs" "${NC}"
}

TOTAL_STEPS=4
FAILED=0

run_one() {
  local idx="$1" name="$2" tags="$3"
  shift 3
  local start=$SECONDS
  step_begin "$idx" "$TOTAL_STEPS" "$name" "$tags"
  sub "命令: make install $*"
  if make install "$@"; then
    step_ok "$((SECONDS - start))"
  else
    err "❌ 本步失败: ${name}"
    FAILED=1
    return 1
  fi
}

printf '\n%b\n' "${BOLD}${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
printf '%b\n' "${BOLD}${BLUE}║  🔍 DMS 多版本构建校验                                      ║${NC}"
printf '%b\n' "${BOLD}${BLUE}║  依次验证 Makefile install 在四种标签组合下均可通过         ║${NC}"
printf '%b\n' "${BOLD}${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
sub "工作目录: ${ROOT}"

# 1 社区版: dummyhead（EDITION 默认 ce）
run_one 1 "社区版 (CE)" "dummyhead" || true

# 2 试用版: dummyhead + trial
if [[ "$FAILED" -eq 0 ]]; then
  run_one 2 "试用版 (Trial)" "dummyhead,trial" EDITION=trial || true
fi

# 3 企业版: dummyhead + enterprise
if [[ "$FAILED" -eq 0 ]]; then
  run_one 3 "企业版 (EE)" "dummyhead,enterprise" EDITION=ee || true
fi

# 4 DMS 企业版: dummyhead + enterprise + dms
if [[ "$FAILED" -eq 0 ]]; then
  run_one 4 "DMS 企业版 (EE + DMS)" "dummyhead,enterprise,dms" EDITION=ee PRODUCT_CATEGORY=dms || true
fi

printf '\n'
bar
if [[ "$FAILED" -eq 0 ]]; then
  printf '%b🎉 全部 %s 个版本构建校验通过。%s\n' "${GREEN}${BOLD}" "$TOTAL_STEPS" "${NC}"
  bar
  printf '\n'
  exit 0
else
  printf '%b💥 构建校验未全部通过，请根据上方日志修复后重试。%s\n' "${RED}${BOLD}" "${NC}"
  bar
  printf '\n'
  exit 1
fi
