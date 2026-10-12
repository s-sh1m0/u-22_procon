# DiffGraph for VS Code

Go の GitHub PR を、関数単位の依存グラフとクラスタに分けて可視化する [DiffGraph](https://github.com/s-sh1m0/u-22_procon) の VS Code 拡張です。Web 版と同じ解析画面を VS Code の中で開けます。グラフで関数を選ぶと、その関数の差分が VS Code の diff エディタで開きます。

## 前提

DiffGraph サーバーがローカルで起動している必要があります。手順はリポジトリの README を参照してください。

```bash
docker compose -f compose.prod.yml up -d   # http://localhost:20080
```

## 使い方

1. コマンドパレット（Ctrl+Shift+P / Cmd+Shift+P）から **DiffGraph: PR を解析** を実行する
2. Go の GitHub PR の URL を入力する。クリップボードに PR の URL があれば、最初から入力されている
3. 初回だけ GitHub へのサインインを求められるので許可する（VS Code 標準の GitHub 認証、スコープは `repo` と `read:user`）
4. 解析が終わるとグラフが表示される。関数を選ぶと、隣のタブにその関数の diff が開く
5. 同じ解析結果を Web 版で見たいときは、タブ右上のアイコンか **DiffGraph: ブラウザで開く** を使う

## 設定

| 設定                  | 既定値                   | 説明                     |
| --------------------- | ------------------------ | ------------------------ |
| `diffgraph.serverUrl` | `http://localhost:20080` | DiffGraph サーバーの URL |

## メモ

- GitHub のトークンは拡張の中（拡張ホスト）だけで扱い、ローカルの DiffGraph サーバーにだけ送ります
- グラフ画面はライトテーマ固定です
