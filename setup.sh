#!/usr/bin/env bash
set -euo pipefail

ENV_FILE=".env"

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

printf "GITHUB_CLIENT_ID: "
read -r github_client_id
if [ -z "$github_client_id" ]; then
  echo "エラー: GITHUB_CLIENT_ID は必須です。"
  exit 1
fi

printf "GITHUB_CLIENT_SECRET: "
read -r github_client_secret
if [ -z "$github_client_secret" ]; then
  echo "エラー: GITHUB_CLIENT_SECRET は必須です。"
  exit 1
fi

default_callback="http://localhost:20080/auth/github/callback"
printf "GITHUB_CALLBACK_URL [%s]: " "$default_callback"
read -r github_callback_url
github_callback_url="${github_callback_url:-$default_callback}"

session_secret=$(openssl rand -base64 32)

cat > "$ENV_FILE" <<EOF
GITHUB_CLIENT_ID=${github_client_id}
GITHUB_CLIENT_SECRET=${github_client_secret}
GITHUB_CALLBACK_URL=${github_callback_url}
SESSION_SECRET=${session_secret}
EOF

echo ""
echo ".env を生成しました。"
echo "起動: docker compose -f compose.prod.yml up -d"
