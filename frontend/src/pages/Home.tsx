import { Button } from "@/components/ui/button"
import Header from "@/components/layout/Header"

// TODO: PR URL 入力フォーム・解析リクエスト送信
export default function Home() {
  return (
    <div>
      <Header />
      <main className="p-8 space-y-4">
        <h1 className="text-2xl font-bold">ReviewArena</h1>
        <Button>動作確認</Button>
      </main>
    </div>
  );
}
