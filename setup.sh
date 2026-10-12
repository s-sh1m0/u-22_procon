#!/usr/bin/env bash
set -euo pipefail

# どこから実行してもリポジトリ直下に .env を作る
cd "$(dirname "$0")"

ENV_FILE=".env"
APP_URL="http://localhost:20080"
CALLBACK_URL="${APP_URL}/auth/github/callback"

# 空でない値が入力されるまで同じ質問を繰り返し、結果を変数 $1 に入れる
ask_required() {
  local var_name="$1" label="$2" input
  while true; do
    printf "%s: " "$label"
    if ! read -r input; then
      echo ""
      echo "中止しました。"
      exit 1
    fi
    if [ -n "$input" ]; then
      printf -v "$var_name" '%s' "$input"
      return
    fi
    echo "  空のままでは進めません。値を貼り付けて Enter を押してください。"
  done
}

echo "=== DiffGraph セットアップ ==="
echo ""

if [ -f "$ENV_FILE" ]; then
  printf ".env が既に存在します。上書きしますか？ [y/N]: "
  read -r overwrite
  if [[ ! "$overwrite" =~ ^[Yy]$ ]]; then
    echo "中止しました。"
    exit 0
  fi
fi

ask_required github_client_id "GITHUB_CLIENT_ID"
ask_required github_client_secret "GITHUB_CLIENT_SECRET"

session_secret=$(openssl rand -base64 32)

cat > "$ENV_FILE" <<EOF
GITHUB_CLIENT_ID=${github_client_id}
GITHUB_CLIENT_SECRET=${github_client_secret}
GITHUB_CALLBACK_URL=${CALLBACK_URL}
SESSION_SECRET=${session_secret}
EOF

echo ""
echo ".env を生成しました。"
echo "起動: docker compose -f compose.prod.yml up -d"
