// shadcn/ui 必須ユーティリティ
// TODO: npm install clsx tailwind-merge 後に実装
export function cn(...inputs: unknown[]): string {
  return inputs.filter(Boolean).join(" ");
}
