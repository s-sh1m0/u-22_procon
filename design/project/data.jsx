// サンプルPRのデータ — 大きめ（30+ファイル）の架空ネットアプリPR
// 4つのクラスタ: 追加(added) / 削除(removed) / 修正(modified) / リファクタ(refactor)

const PR_META = {
  repo: "kobold/checkout-service",
  number: 1284,
  title: "feat(payments): Stripe連携のリファクタと領収書PDF生成を追加",
  author: "okuda-y",
  authorAvatar: "OY",
  branch: "feat/payments-refactor",
  base: "main",
  commits: 18,
  filesChanged: 34,
  additions: 1247,
  deletions: 583,
  status: "open",
  createdAt: "2日前",
  description: "決済フローのStripe SDK v14対応、領収書PDF生成のサービス分離、レガシーWebhookハンドラの削除を含む。",
};

// モジュール（≒ファイル/ファイル群）
// id, name, path, cluster: added|removed|modified|refactor, layer: ui|domain|data|infra
// loc, additions, deletions, functions: [{name, change}]
const MODULES = [
  // === UI Layer ===
  { id: "m1", name: "CheckoutForm", path: "src/ui/checkout/CheckoutForm.tsx", cluster: "modified", layer: "ui", additions: 87, deletions: 42, loc: 312,
    functions: [{name: "handleSubmit", change: "modified"}, {name: "validateCard", change: "modified"}, {name: "renderError", change: "added"}] },
  { id: "m2", name: "PaymentMethodPicker", path: "src/ui/checkout/PaymentMethodPicker.tsx", cluster: "added", layer: "ui", additions: 156, deletions: 0, loc: 156,
    functions: [{name: "PaymentMethodPicker", change: "added"}, {name: "useMethodList", change: "added"}] },
  { id: "m3", name: "ReceiptModal", path: "src/ui/checkout/ReceiptModal.tsx", cluster: "added", layer: "ui", additions: 98, deletions: 0, loc: 98,
    functions: [{name: "ReceiptModal", change: "added"}, {name: "downloadPdf", change: "added"}] },
  { id: "m4", name: "OrderSummary", path: "src/ui/checkout/OrderSummary.tsx", cluster: "refactor", layer: "ui", additions: 41, deletions: 78, loc: 145,
    functions: [{name: "OrderSummary", change: "modified"}, {name: "formatLine", change: "refactor"}] },
  { id: "m5", name: "LegacyCardForm", path: "src/ui/checkout/LegacyCardForm.tsx", cluster: "removed", layer: "ui", additions: 0, deletions: 184, loc: 0,
    functions: [{name: "LegacyCardForm", change: "removed"}, {name: "validateLegacy", change: "removed"}] },
  { id: "m6", name: "Button", path: "src/ui/common/Button.tsx", cluster: "modified", layer: "ui", additions: 12, deletions: 8, loc: 88,
    functions: [{name: "Button", change: "modified"}] },
  { id: "m7", name: "useCheckoutFlow", path: "src/ui/hooks/useCheckoutFlow.ts", cluster: "refactor", layer: "ui", additions: 64, deletions: 91, loc: 142,
    functions: [{name: "useCheckoutFlow", change: "refactor"}, {name: "useStep", change: "refactor"}] },

  // === Domain Layer ===
  { id: "m8", name: "PaymentService", path: "src/domain/payment/PaymentService.ts", cluster: "refactor", layer: "domain", additions: 142, deletions: 168, loc: 287,
    functions: [{name: "charge", change: "refactor"}, {name: "refund", change: "refactor"}, {name: "capture", change: "modified"}] },
  { id: "m9", name: "StripeAdapter", path: "src/domain/payment/StripeAdapter.ts", cluster: "added", layer: "domain", additions: 218, deletions: 0, loc: 218,
    functions: [{name: "StripeAdapter", change: "added"}, {name: "createIntent", change: "added"}, {name: "confirmIntent", change: "added"}, {name: "mapError", change: "added"}] },
  { id: "m10", name: "LegacyStripeClient", path: "src/domain/payment/LegacyStripeClient.ts", cluster: "removed", layer: "domain", additions: 0, deletions: 245, loc: 0,
    functions: [{name: "LegacyStripeClient", change: "removed"}, {name: "legacyCharge", change: "removed"}] },
  { id: "m11", name: "ReceiptGenerator", path: "src/domain/receipt/ReceiptGenerator.ts", cluster: "added", layer: "domain", additions: 187, deletions: 0, loc: 187,
    functions: [{name: "ReceiptGenerator", change: "added"}, {name: "renderPdf", change: "added"}, {name: "buildLineItems", change: "added"}] },
  { id: "m12", name: "OrderRepository", path: "src/domain/order/OrderRepository.ts", cluster: "modified", layer: "domain", additions: 28, deletions: 14, loc: 196,
    functions: [{name: "save", change: "modified"}, {name: "findById", change: "modified"}, {name: "markPaid", change: "added"}] },
  { id: "m13", name: "TaxCalculator", path: "src/domain/tax/TaxCalculator.ts", cluster: "modified", layer: "domain", additions: 23, deletions: 11, loc: 142,
    functions: [{name: "calculate", change: "modified"}, {name: "applyRegion", change: "modified"}] },
  { id: "m14", name: "PriceFormatter", path: "src/domain/format/PriceFormatter.ts", cluster: "refactor", layer: "domain", additions: 18, deletions: 32, loc: 64,
    functions: [{name: "format", change: "refactor"}, {name: "parseCurrency", change: "refactor"}] },

  // === Data Layer ===
  { id: "m15", name: "PaymentRecord", path: "src/data/models/PaymentRecord.ts", cluster: "modified", layer: "data", additions: 22, deletions: 8, loc: 88,
    functions: [{name: "PaymentRecord", change: "modified"}] },
  { id: "m16", name: "ReceiptRecord", path: "src/data/models/ReceiptRecord.ts", cluster: "added", layer: "data", additions: 64, deletions: 0, loc: 64,
    functions: [{name: "ReceiptRecord", change: "added"}, {name: "toJson", change: "added"}] },
  { id: "m17", name: "OrderDao", path: "src/data/dao/OrderDao.ts", cluster: "modified", layer: "data", additions: 31, deletions: 18, loc: 154,
    functions: [{name: "insert", change: "modified"}, {name: "update", change: "modified"}] },
  { id: "m18", name: "MigrationsV14", path: "src/data/migrations/2026_04_v14.sql", cluster: "added", layer: "data", additions: 42, deletions: 0, loc: 42,
    functions: [{name: "up", change: "added"}, {name: "down", change: "added"}] },
  { id: "m19", name: "LegacyPaymentTable", path: "src/data/migrations/2024_legacy.sql", cluster: "removed", layer: "data", additions: 0, deletions: 38, loc: 0,
    functions: [{name: "drop_legacy", change: "removed"}] },

  // === Infra Layer ===
  { id: "m20", name: "WebhookHandler", path: "src/infra/webhook/WebhookHandler.ts", cluster: "refactor", layer: "infra", additions: 96, deletions: 124, loc: 188,
    functions: [{name: "handle", change: "refactor"}, {name: "verifySignature", change: "refactor"}] },
  { id: "m21", name: "LegacyWebhook", path: "src/infra/webhook/LegacyWebhook.ts", cluster: "removed", layer: "infra", additions: 0, deletions: 116, loc: 0,
    functions: [{name: "LegacyWebhook", change: "removed"}] },
  { id: "m22", name: "PdfRenderer", path: "src/infra/pdf/PdfRenderer.ts", cluster: "added", layer: "infra", additions: 142, deletions: 0, loc: 142,
    functions: [{name: "PdfRenderer", change: "added"}, {name: "renderTemplate", change: "added"}] },
  { id: "m23", name: "S3Uploader", path: "src/infra/storage/S3Uploader.ts", cluster: "modified", layer: "infra", additions: 18, deletions: 6, loc: 92,
    functions: [{name: "upload", change: "modified"}, {name: "presign", change: "added"}] },
  { id: "m24", name: "Logger", path: "src/infra/log/Logger.ts", cluster: "modified", layer: "infra", additions: 8, deletions: 4, loc: 76,
    functions: [{name: "info", change: "modified"}, {name: "error", change: "modified"}] },
  { id: "m25", name: "EnvConfig", path: "src/infra/config/env.ts", cluster: "modified", layer: "infra", additions: 14, deletions: 6, loc: 58,
    functions: [{name: "loadEnv", change: "modified"}] },

  // === Tests ===
  { id: "m26", name: "PaymentService.test", path: "src/domain/payment/__tests__/PaymentService.test.ts", cluster: "modified", layer: "data", additions: 88, deletions: 64, loc: 234,
    functions: [{name: "describe charge", change: "refactor"}, {name: "describe refund", change: "modified"}] },
  { id: "m27", name: "StripeAdapter.test", path: "src/domain/payment/__tests__/StripeAdapter.test.ts", cluster: "added", layer: "data", additions: 184, deletions: 0, loc: 184,
    functions: [{name: "describe StripeAdapter", change: "added"}] },
  { id: "m28", name: "ReceiptGenerator.test", path: "src/domain/receipt/__tests__/ReceiptGenerator.test.ts", cluster: "added", layer: "data", additions: 122, deletions: 0, loc: 122,
    functions: [{name: "describe ReceiptGenerator", change: "added"}] },
  { id: "m29", name: "checkout.e2e", path: "e2e/checkout.spec.ts", cluster: "modified", layer: "data", additions: 56, deletions: 32, loc: 198,
    functions: [{name: "checkout flow", change: "modified"}, {name: "receipt download", change: "added"}] },

  // === Misc ===
  { id: "m30", name: "package.json", path: "package.json", cluster: "modified", layer: "infra", additions: 6, deletions: 2, loc: 88,
    functions: [{name: "deps", change: "modified"}] },
  { id: "m31", name: "README", path: "README.md", cluster: "modified", layer: "infra", additions: 24, deletions: 8, loc: 142,
    functions: [{name: "section: payments", change: "modified"}] },
  { id: "m32", name: "CHANGELOG", path: "CHANGELOG.md", cluster: "modified", layer: "infra", additions: 18, deletions: 0, loc: 88,
    functions: [{name: "v14.0.0", change: "added"}] },
  { id: "m33", name: "stripe.types", path: "src/domain/payment/stripe.types.ts", cluster: "added", layer: "domain", additions: 78, deletions: 0, loc: 78,
    functions: [{name: "type StripeIntent", change: "added"}, {name: "type StripeError", change: "added"}] },
  { id: "m34", name: "PaymentError", path: "src/domain/payment/PaymentError.ts", cluster: "refactor", layer: "domain", additions: 32, deletions: 28, loc: 64,
    functions: [{name: "PaymentError", change: "refactor"}, {name: "fromStripe", change: "added"}] },
];

// 依存関係（id -> id）
const EDGES = [
  // UI -> Domain
  ["m1", "m8"], ["m1", "m6"], ["m1", "m2"], ["m1", "m7"],
  ["m2", "m8"], ["m2", "m6"],
  ["m3", "m11"], ["m3", "m6"], ["m3", "m7"],
  ["m4", "m13"], ["m4", "m14"],
  ["m7", "m8"], ["m7", "m12"],
  // Domain -> Domain
  ["m8", "m9"], ["m8", "m34"], ["m8", "m12"], ["m8", "m13"],
  ["m9", "m33"], ["m9", "m34"],
  ["m11", "m22"], ["m11", "m14"], ["m11", "m12"],
  ["m12", "m17"], ["m12", "m15"],
  ["m13", "m14"],
  // Domain -> Data
  ["m8", "m15"], ["m11", "m16"], ["m17", "m15"], ["m17", "m16"],
  // Domain -> Infra
  ["m9", "m24"], ["m9", "m25"], ["m11", "m23"], ["m22", "m23"],
  ["m20", "m8"], ["m20", "m9"], ["m20", "m24"],
  // Tests -> targets
  ["m26", "m8"], ["m27", "m9"], ["m28", "m11"], ["m29", "m1"], ["m29", "m3"],
  // Migrations
  ["m18", "m15"], ["m18", "m16"],
];

// クラスタ統計
const CLUSTER_STATS = {
  added: { count: MODULES.filter(m => m.cluster === "added").length,
           additions: MODULES.filter(m => m.cluster === "added").reduce((s, m) => s + m.additions, 0),
           deletions: 0 },
  removed: { count: MODULES.filter(m => m.cluster === "removed").length,
             additions: 0,
             deletions: MODULES.filter(m => m.cluster === "removed").reduce((s, m) => s + m.deletions, 0) },
  modified: { count: MODULES.filter(m => m.cluster === "modified").length,
              additions: MODULES.filter(m => m.cluster === "modified").reduce((s, m) => s + m.additions, 0),
              deletions: MODULES.filter(m => m.cluster === "modified").reduce((s, m) => s + m.deletions, 0) },
  refactor: { count: MODULES.filter(m => m.cluster === "refactor").length,
              additions: MODULES.filter(m => m.cluster === "refactor").reduce((s, m) => s + m.additions, 0),
              deletions: MODULES.filter(m => m.cluster === "refactor").reduce((s, m) => s + m.deletions, 0) },
};

// クラスタ定義（色とラベル）
const CLUSTERS = {
  added:    { label: "追加",     short: "Added",    color: "#16a34a", bg: "#dcfce7", border: "#86efac", icon: "+" },
  removed:  { label: "削除",     short: "Removed",  color: "#dc2626", bg: "#fee2e2", border: "#fca5a5", icon: "−" },
  modified: { label: "修正",     short: "Modified", color: "#d97706", bg: "#fef3c7", border: "#fcd34d", icon: "~" },
  refactor: { label: "リファクタ", short: "Refactor", color: "#7c3aed", bg: "#ede9fe", border: "#c4b5fd", icon: "↻" },
};

// 4階層
const LAYERS = {
  ui:     { label: "UI",     desc: "プレゼンテーション層" },
  domain: { label: "Domain", desc: "ドメイン・ビジネスロジック" },
  data:   { label: "Data",   desc: "データ・モデル・テスト" },
  infra:  { label: "Infra",  desc: "インフラ・外部I/O" },
};

// サンプルdiff（モジュールクリック時の右パネル用）
const SAMPLE_DIFF = {
  m9: {
    file: "src/domain/payment/StripeAdapter.ts",
    hunks: [
      { header: "@@ -0,0 +1,40 @@", lines: [
        { kind: "add", n: 1,  text: "import Stripe from 'stripe';" },
        { kind: "add", n: 2,  text: "import type { StripeIntent, StripeError } from './stripe.types';" },
        { kind: "add", n: 3,  text: "import { PaymentError } from './PaymentError';" },
        { kind: "add", n: 4,  text: "" },
        { kind: "add", n: 5,  text: "export class StripeAdapter {" },
        { kind: "add", n: 6,  text: "  private client: Stripe;" },
        { kind: "add", n: 7,  text: "" },
        { kind: "add", n: 8,  text: "  constructor(apiKey: string) {" },
        { kind: "add", n: 9,  text: "    this.client = new Stripe(apiKey, { apiVersion: '2026-04-15' });" },
        { kind: "add", n: 10, text: "  }" },
        { kind: "add", n: 11, text: "" },
        { kind: "add", n: 12, text: "  async createIntent(amount: number, currency: string): Promise<StripeIntent> {" },
        { kind: "add", n: 13, text: "    try {" },
        { kind: "add", n: 14, text: "      const intent = await this.client.paymentIntents.create({" },
        { kind: "add", n: 15, text: "        amount, currency, automatic_payment_methods: { enabled: true }," },
        { kind: "add", n: 16, text: "      });" },
        { kind: "add", n: 17, text: "      return { id: intent.id, clientSecret: intent.client_secret! };" },
        { kind: "add", n: 18, text: "    } catch (e) {" },
        { kind: "add", n: 19, text: "      throw this.mapError(e);" },
        { kind: "add", n: 20, text: "    }" },
        { kind: "add", n: 21, text: "  }" },
      ]},
    ],
  },
  m8: {
    file: "src/domain/payment/PaymentService.ts",
    hunks: [
      { header: "@@ -12,28 +12,18 @@ export class PaymentService {", lines: [
        { kind: "ctx", n: 12, text: "  constructor(" },
        { kind: "del", n: 13, text: "    private readonly legacy: LegacyStripeClient," },
        { kind: "del", n: 14, text: "    private readonly logger: Logger," },
        { kind: "add", n: 13, text: "    private readonly stripe: StripeAdapter," },
        { kind: "add", n: 14, text: "    private readonly orders: OrderRepository," },
        { kind: "ctx", n: 15, text: "  ) {}" },
        { kind: "ctx", n: 16, text: "" },
        { kind: "del", n: 17, text: "  async charge(amount: number, token: string): Promise<void> {" },
        { kind: "del", n: 18, text: "    const result = await this.legacy.legacyCharge(amount, token);" },
        { kind: "del", n: 19, text: "    if (!result.ok) throw new Error(result.code);" },
        { kind: "add", n: 17, text: "  async charge(orderId: string, amount: number): Promise<PaymentRecord> {" },
        { kind: "add", n: 18, text: "    const intent = await this.stripe.createIntent(amount, 'jpy');" },
        { kind: "add", n: 19, text: "    const record = await this.orders.markPaid(orderId, intent.id);" },
        { kind: "add", n: 20, text: "    return record;" },
        { kind: "ctx", n: 21, text: "  }" },
      ]},
    ],
  },
};

window.PR_META = PR_META;
window.MODULES = MODULES;
window.EDGES = EDGES;
window.CLUSTER_STATS = CLUSTER_STATS;
window.CLUSTERS = CLUSTERS;
window.LAYERS = LAYERS;
window.SAMPLE_DIFF = SAMPLE_DIFF;
