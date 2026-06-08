#!/usr/bin/env bash
# Haiwo Server 部署 / 管理脚本
#
# 功能：
#   1. 选择语言（中 / 日 / 英）
#   2. 检测是否已安装 Server；已安装则提供 [重启 / 停止 / 更新]
#   3. 安装或更新：删除旧二进制 -> 下载新二进制；首次安装需输入绑定端口与登录密码
#
# 默认安装在用户目录，无需 root。可用环境变量覆盖：
#   HAIWO_SERVER_HOME         安装目录（默认 ${HOME:-/opt/haiwo}/.haiwo/server）
#   HAIWO_SERVER_BINARY_URL   下载地址（默认见下）
set -euo pipefail

: "${HAIWO_SERVER_HOME:=${HOME:-/opt/haiwo}/.haiwo/server}"
: "${HAIWO_SERVER_BINARY_URL:=https://github.com/dollarkillerx/Haiwo/releases/download/v0.0.3/haiwo-server-linux-amd64}"

BIN="${HAIWO_SERVER_HOME}/haiwo-server"
CONFIG="${HAIWO_SERVER_HOME}/config.toml"
PID_FILE="${HAIWO_SERVER_HOME}/haiwo-server.pid"
LOG_FILE="${HAIWO_SERVER_HOME}/haiwo-server.out"
DATA_DIR="${HAIWO_SERVER_HOME}/data"

LANG_SEL="en"

# ——— 国际化 ———
msg() {
  case "$LANG_SEL" in
  zh)
    case "$1" in
    installed_title) echo "检测到已安装 Haiwo Server（${HAIWO_SERVER_HOME}）" ;;
    running_yes) echo "当前状态：运行中 (pid=$2)" ;;
    running_no) echo "当前状态：未运行" ;;
    choose_action) echo "请选择操作：" ;;
    act_restart) echo "重启" ;;
    act_stop) echo "停止" ;;
    act_update) echo "更新" ;;
    act_passwd) echo "修改密码" ;;
    passwd_changed) echo "密码已修改" ;;
    prompt_choice) printf "输入序号: " ;;
    invalid) echo "无效选择" ;;
    fresh_title) echo "首次安装 Haiwo Server" ;;
    prompt_port) printf "请输入绑定端口 [默认 8080]: " ;;
    prompt_password) printf "请输入控制台登录密码: " ;;
    empty_password) echo "密码不能为空" ;;
    removing_old) echo "删除旧二进制..." ;;
    downloading) echo "正在下载: ${HAIWO_SERVER_BINARY_URL}" ;;
    need_tool) echo "需要 curl 或 wget 才能下载" ;;
    download_fail) echo "下载失败" ;;
    stopping) echo "正在停止..." ;;
    stopped) echo "已停止" ;;
    starting) echo "正在启动..." ;;
    started) echo "已启动 (pid=$2)" ;;
    restarting) echo "正在重启..." ;;
    updating) echo "正在更新..." ;;
    not_running) echo "服务未在运行" ;;
    done_install) echo "安装完成" ;;
    done_update) echo "更新完成" ;;
    url_hint) echo "控制台地址: $2" ;;
    pwd_hint) echo "登录密码: $2" ;;
    log_hint) echo "日志: $2" ;;
    esac ;;
  ja)
    case "$1" in
    installed_title) echo "Haiwo Server のインストールを検出しました（${HAIWO_SERVER_HOME}）" ;;
    running_yes) echo "状態：実行中 (pid=$2)" ;;
    running_no) echo "状態：停止中" ;;
    choose_action) echo "操作を選択してください：" ;;
    act_restart) echo "再起動" ;;
    act_stop) echo "停止" ;;
    act_update) echo "アップデート" ;;
    act_passwd) echo "パスワード変更" ;;
    passwd_changed) echo "パスワードを変更しました" ;;
    prompt_choice) printf "番号を入力: " ;;
    invalid) echo "無効な選択です" ;;
    fresh_title) echo "Haiwo Server を初回インストールします" ;;
    prompt_port) printf "バインドするポートを入力 [既定 8080]: " ;;
    prompt_password) printf "コンソールのログインパスワードを入力: " ;;
    empty_password) echo "パスワードは空にできません" ;;
    removing_old) echo "古いバイナリを削除中..." ;;
    downloading) echo "ダウンロード中: ${HAIWO_SERVER_BINARY_URL}" ;;
    need_tool) echo "ダウンロードには curl または wget が必要です" ;;
    download_fail) echo "ダウンロードに失敗しました" ;;
    stopping) echo "停止中..." ;;
    stopped) echo "停止しました" ;;
    starting) echo "起動中..." ;;
    started) echo "起動しました (pid=$2)" ;;
    restarting) echo "再起動中..." ;;
    updating) echo "アップデート中..." ;;
    not_running) echo "サービスは実行されていません" ;;
    done_install) echo "インストール完了" ;;
    done_update) echo "アップデート完了" ;;
    url_hint) echo "コンソール URL: $2" ;;
    pwd_hint) echo "ログインパスワード: $2" ;;
    log_hint) echo "ログ: $2" ;;
    esac ;;
  *)
    case "$1" in
    installed_title) echo "Existing Haiwo Server installation detected (${HAIWO_SERVER_HOME})" ;;
    running_yes) echo "Status: running (pid=$2)" ;;
    running_no) echo "Status: stopped" ;;
    choose_action) echo "Choose an action:" ;;
    act_restart) echo "Restart" ;;
    act_stop) echo "Stop" ;;
    act_update) echo "Update" ;;
    act_passwd) echo "Change password" ;;
    passwd_changed) echo "Password changed" ;;
    prompt_choice) printf "Enter number: " ;;
    invalid) echo "Invalid choice" ;;
    fresh_title) echo "First-time install of Haiwo Server" ;;
    prompt_port) printf "Bind port [default 8080]: " ;;
    prompt_password) printf "Console login password: " ;;
    empty_password) echo "Password cannot be empty" ;;
    removing_old) echo "Removing old binary..." ;;
    downloading) echo "Downloading: ${HAIWO_SERVER_BINARY_URL}" ;;
    need_tool) echo "curl or wget is required to download" ;;
    download_fail) echo "Download failed" ;;
    stopping) echo "Stopping..." ;;
    stopped) echo "Stopped" ;;
    starting) echo "Starting..." ;;
    started) echo "Started (pid=$2)" ;;
    restarting) echo "Restarting..." ;;
    updating) echo "Updating..." ;;
    not_running) echo "Service is not running" ;;
    done_install) echo "Install complete" ;;
    done_update) echo "Update complete" ;;
    url_hint) echo "Console URL: $2" ;;
    pwd_hint) echo "Login password: $2" ;;
    log_hint) echo "Log: $2" ;;
    esac ;;
  esac
}

choose_language() {
  echo "请选择语言 / 言語を選択 / Select language:"
  echo "  1) 中文"
  echo "  2) 日本語"
  echo "  3) English"
  printf "> "
  read -r choice || true
  case "$choice" in
  1) LANG_SEL="zh" ;;
  2) LANG_SEL="ja" ;;
  *) LANG_SEL="en" ;;
  esac
}

is_installed() {
  [ -f "$CONFIG" ] && [ -x "$BIN" ]
}

running_pid() {
  [ -f "$PID_FILE" ] || return 1
  local pid
  pid="$(cat "$PID_FILE" 2>/dev/null || true)"
  [ -n "$pid" ] && kill -0 "$pid" >/dev/null 2>&1 || return 1
  echo "$pid"
}

stop_server() {
  msg stopping
  if [ -f "$PID_FILE" ]; then
    local pid
    pid="$(cat "$PID_FILE" 2>/dev/null || true)"
    if [ -n "$pid" ] && kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      for _ in $(seq 1 20); do
        kill -0 "$pid" >/dev/null 2>&1 || break
        sleep 0.2
      done
      kill -0 "$pid" >/dev/null 2>&1 && kill -9 "$pid" >/dev/null 2>&1 || true
    fi
    rm -f "$PID_FILE"
  fi
  if command -v pgrep >/dev/null 2>&1; then
    while IFS= read -r p; do
      [ -n "$p" ] || continue
      kill "$p" >/dev/null 2>&1 || true
    done <<EOF
$(pgrep -f "$BIN" || true)
EOF
  fi
  msg stopped
}

start_server() {
  msg starting
  mkdir -p "$HAIWO_SERVER_HOME" "$DATA_DIR"
  nohup "$BIN" -c config -cPath "$HAIWO_SERVER_HOME/" >>"$LOG_FILE" 2>&1 &
  echo $! >"$PID_FILE"
  sleep 1
  msg started "$(cat "$PID_FILE")"
  msg log_hint "$LOG_FILE"
}

download_binary() {
  msg downloading
  local tmp="${BIN}.tmp"
  rm -f "$tmp"
  if command -v curl >/dev/null 2>&1; then
    curl -fSL "$HAIWO_SERVER_BINARY_URL" -o "$tmp" || { msg download_fail; exit 1; }
  elif command -v wget >/dev/null 2>&1; then
    wget -qO "$tmp" "$HAIWO_SERVER_BINARY_URL" || { msg download_fail; exit 1; }
  else
    msg need_tool
    exit 1
  fi
  chmod +x "$tmp"
  mv "$tmp" "$BIN"
}

# 安装/更新二进制：删除旧的 -> 下载新的
replace_binary() {
  if [ -e "$BIN" ]; then
    msg removing_old
    rm -f "$BIN"
  fi
  download_binary
}

write_config() {
  local port="$1" password="$2" token
  token="$(generate_token)"
  cat >"$CONFIG" <<EOF
[ServiceConfiguration]
Addr = ":${port}"
Debug = false

[WebConfiguration]
Password = "${password}"

[AgentConfiguration]
Token = "${token}"

[StorageConfiguration]
Backend = "file"
DataFile = "${DATA_DIR}/haiwo.json"
LogFile = "${DATA_DIR}/haiwo.log"
LogMaxMB = 50
EOF
}

generate_token() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 16
  elif [ -r /dev/urandom ]; then
    head -c 16 /dev/urandom | od -An -tx1 | tr -d ' \n'
  else
    echo "haiwo-agent-token"
  fi
}

config_port() {
  awk -F'"' '/^Addr[[:space:]]*=/ {gsub(/[^0-9]/,"",$2); print $2}' "$CONFIG" 2>/dev/null
}

config_password() {
  awk -F'"' '/^Password[[:space:]]*=/ {print $2; exit}' "$CONFIG" 2>/dev/null
}

# 重写 config.toml 中 [WebConfiguration] 的 Password（仅替换第一处 Password 行）。
# 通过 ENVIRON 传值，避免密码中的特殊字符被 awk 转义处理。
set_config_password() {
  local tmp="${CONFIG}.tmp"
  NEWPASS="$1" awk '
    /^Password[[:space:]]*=/ && !done { print "Password = \"" ENVIRON["NEWPASS"] "\""; done=1; next }
    { print }
  ' "$CONFIG" >"$tmp" && mv "$tmp" "$CONFIG"
}

# 提示输入密码（隐藏输入），结果写入全局变量 NEW_PASSWORD。
NEW_PASSWORD=""
prompt_new_password() {
  NEW_PASSWORD=""
  while [ -z "$NEW_PASSWORD" ]; do
    msg prompt_password
    read -rs NEW_PASSWORD || true
    echo
    [ -n "$NEW_PASSWORD" ] || msg empty_password
  done
}

print_access() {
  local port pwd
  port="$(config_port)"
  pwd="$(config_password)"
  [ -n "$port" ] || port="8080"
  msg url_hint "http://<server-ip>:${port}/"
  msg pwd_hint "$pwd"
}

fresh_install() {
  msg fresh_title
  mkdir -p "$HAIWO_SERVER_HOME" "$DATA_DIR"

  msg prompt_port
  read -r port || true
  [ -n "$port" ] || port="8080"

  prompt_new_password

  write_config "$port" "$NEW_PASSWORD"
  replace_binary
  start_server
  msg done_install
  print_access
}

manage_existing() {
  msg installed_title
  local pid
  if pid="$(running_pid)"; then
    msg running_yes "$pid"
  else
    msg running_no
  fi
  msg choose_action
  echo "  1) $(msg act_restart)"
  echo "  2) $(msg act_stop)"
  echo "  3) $(msg act_update)"
  echo "  4) $(msg act_passwd)"
  msg prompt_choice
  read -r action || true
  case "$action" in
  1)
    msg restarting
    stop_server
    start_server
    print_access
    ;;
  2)
    if running_pid >/dev/null 2>&1; then
      stop_server
    else
      msg not_running
    fi
    ;;
  3)
    msg updating
    stop_server
    replace_binary
    start_server
    msg done_update
    print_access
    ;;
  4)
    prompt_new_password
    set_config_password "$NEW_PASSWORD"
    msg passwd_changed
    # 密码在启动时读取，需重启才生效
    if running_pid >/dev/null 2>&1; then
      stop_server
      start_server
    fi
    print_access
    ;;
  *)
    msg invalid
    exit 1
    ;;
  esac
}

main() {
  choose_language
  if is_installed; then
    manage_existing
  else
    fresh_install
  fi
}

main "$@"
