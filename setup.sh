#!/usr/bin/env bash
set -euo pipefail

# どこから実行してもリポジトリ直下に .env を作る
cd "$(dirname "$0")"

ENV_FILE=".env"
APP_URL="http://localhost:20080"
CALLBACK_URL="${APP_URL}/auth/github/callback"

# 入力が途中で終わった（Ctrl+D など）ときに中止する
cancel() {
  echo ""
  echo "中止しました。"
  exit 1
}

# 空でない値が入力されるまで同じ質問を繰り返し、結果を変数 answer に入れる
ask_required() {
  local label="$1"
  answer=""
  while [ -z "$answer" ]; do
    printf "%s: " "$label"
    read -r answer || cancel
    if [ -z "$answer" ]; then
      echo "  空のままでは進めません。値を貼り付けて Enter を押してください。"
    fi
  done
}

cat <<EOF
=== DiffGraph セットアップ ===

DiffGraph の起動に必要な設定ファイル .env を作ります。
次の 3 ステップで進めます。

  1. GitHub で OAuth App を作る（ブラウザで操作します）
  2. 発行された Client ID と Client Secret をここに貼り付ける
  3. .env を書き出す

EOF

if [ -f "$ENV_FILE" ]; then
  echo ".env がすでにあります。上書きすると今の設定は消えます。"
  printf "上書きしますか？ [y/N]: "
  read -r overwrite || cancel
  if [[ ! "$overwrite" =~ ^[Yy]$ ]]; then
    echo "中止しました。既存の .env はそのまま残っています。"
    exit 0
  fi
  echo ""
fi

cat <<EOF
--- ステップ 1/3: GitHub で OAuth App を作る ---

DiffGraph には GitHub アカウントでログインします。
そのための OAuth App を、あなたの GitHub アカウントに登録します。

  1) ブラウザで次の URL を開く
       https://github.com/settings/developers
  2) 左のメニューで「OAuth Apps」を選び、「New OAuth App」をクリックする
     （OAuth App が 1 つもない場合は「Register a new application」）
  3) 次の値を入力して「Register application」をクリックする
       Application name           : DiffGraph（好きな名前でよい）
       Homepage URL               : ${APP_URL}
       Authorization callback URL : ${CALLBACK_URL}
     ※ Authorization callback URL は 1 文字でも違うとログインできません。
       上の値をそのままコピーしてください。
  4) 作成後の画面に出る「Client ID」をコピーする
  5) 「Generate a new client secret」をクリックし、表示された Client Secret をコピーする
     ※ Client Secret はこのとき一度しか表示されません。画面を閉じる前にコピーしてください。

EOF
printf "OAuth App を作ったら Enter を押してください（作成済みならそのまま Enter）: "
read -r _ || cancel

cat <<EOF

--- ステップ 2/3: Client ID と Client Secret を入力 ---

ステップ 1 でコピーした値を貼り付けて Enter を押してください。

EOF
ask_required "Client ID"
github_client_id="$answer"
ask_required "Client Secret"
github_client_secret="$answer"

cat <<EOF

--- ステップ 3/3: .env を書き出す ---

EOF

# openssl が無い環境でも動くよう、POSIX の od / tr で /dev/urandom の 32 バイトを 16 進文字列にする
session_secret=$(od -An -tx1 -N32 /dev/urandom | tr -d ' \n')

cat > "$ENV_FILE" <<EOF
GITHUB_CLIENT_ID=${github_client_id}
GITHUB_CLIENT_SECRET=${github_client_secret}
GITHUB_CALLBACK_URL=${CALLBACK_URL}
SESSION_SECRET=${session_secret}
EOF

cat <<EOF
.env を作りました: $(pwd)/${ENV_FILE}

  GITHUB_CLIENT_ID      入力した Client ID
  GITHUB_CLIENT_SECRET  入力した Client Secret
  GITHUB_CALLBACK_URL   ${CALLBACK_URL}
  SESSION_SECRET        自動で作ったランダムな文字列（ログイン状態の署名に使う）

=== セットアップ完了 ===

次の手順で DiffGraph を使い始められます。

  1) リポジトリのディレクトリで起動する（初回はビルドに数分かかります）
       cd $(pwd)
       docker compose -f compose.prod.yml up -d
  2) ブラウザで ${APP_URL} を開く
  3) 画面右上の「ログイン」から GitHub アカウントでログインする

停止するとき:
  docker compose -f compose.prod.yml down

ログインで OAuth のエラーが出たら、OAuth App の Authorization callback URL が
${CALLBACK_URL} と完全に一致しているか確認してください。
EOF
