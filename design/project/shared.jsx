// shared.jsx — ランディング・ログイン・PR入力・解析中などの共通フロー画面
// すべて 1080x720 (or 1080x780) の枠内に収まるようにデザイン

const T = {
  bg: "#fafaf9",
  bgAlt: "#f5f5f4",
  ink: "#1c1917",
  ink2: "#44403c",
  ink3: "#78716c",
  border: "#e7e5e4",
  borderStrong: "#d6d3d1",
  brand: "#0f766e",     // teal-700, モダンSaaS
  brandSoft: "#ccfbf1",
  brandInk: "#134e4a",
  added: "#16a34a", addedBg: "#dcfce7", addedBd: "#86efac",
  removed: "#dc2626", removedBg: "#fee2e2", removedBd: "#fca5a5",
  modified: "#d97706", modifiedBg: "#fef3c7", modifiedBd: "#fcd34d",
  refactor: "#7c3aed", refactorBg: "#ede9fe", refactorBd: "#c4b5fd",
  font: '"Inter Tight", "Inter", -apple-system, system-ui, sans-serif',
  mono: '"JetBrains Mono", ui-monospace, "SF Mono", Menlo, monospace',
};

// ロゴ（オリジナル — グラフ風アイコン）
function Logo({ size = 28, color = T.brand }) {
  return (
    <svg width={size} height={size} viewBox="0 0 32 32" fill="none">
      <circle cx="6" cy="8" r="3" fill={color} />
      <circle cx="26" cy="8" r="3" fill={color} opacity="0.5" />
      <circle cx="16" cy="22" r="3.5" fill={color} />
      <path d="M8 10 L14 20" stroke={color} strokeWidth="1.6" />
      <path d="M24 10 L18 20" stroke={color} strokeWidth="1.6" opacity="0.6" />
      <path d="M9 8 L23 8" stroke={color} strokeWidth="1.6" strokeDasharray="2 2" opacity="0.4" />
    </svg>
  );
}

function Brand({ size = 18 }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
      <Logo size={size + 8} />
      <span style={{ fontSize: size, fontWeight: 600, letterSpacing: -0.4, color: T.ink, fontFamily: T.font }}>
        diffmap<span style={{ color: T.brand }}>.</span>
      </span>
    </div>
  );
}

// ─────────────── Landing ───────────────
function Landing() {
  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, overflow: "hidden", position: "relative" }}>
      {/* nav */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "20px 36px", borderBottom: `1px solid ${T.border}` }}>
        <Brand />
        <div style={{ display: "flex", alignItems: "center", gap: 24, fontSize: 13, color: T.ink2 }}>
          <span>機能</span><span>ドキュメント</span><span>料金</span>
          <button style={{ background: T.ink, color: "#fff", border: "none", padding: "8px 14px", borderRadius: 8, fontSize: 13, fontWeight: 500, fontFamily: "inherit", cursor: "pointer" }}>ログイン</button>
        </div>
      </div>

      {/* hero */}
      <div style={{ padding: "56px 36px 0", display: "grid", gridTemplateColumns: "1fr 1fr", gap: 48 }}>
        <div>
          <div style={{ display: "inline-flex", alignItems: "center", gap: 6, padding: "5px 10px", background: T.brandSoft, color: T.brandInk, borderRadius: 99, fontSize: 12, fontWeight: 500, marginBottom: 20 }}>
            <span style={{ width: 6, height: 6, borderRadius: 99, background: T.brand }} />
            <span>v1.0 リリース</span>
          </div>
          <h1 style={{ fontSize: 44, lineHeight: 1.05, fontWeight: 600, letterSpacing: -1.2, margin: "0 0 16px", textWrap: "pretty" }}>
            プルリクの依存を、<br/><span style={{ color: T.brand }}>一目で。</span>
          </h1>
          <p style={{ fontSize: 15, lineHeight: 1.55, color: T.ink2, margin: "0 0 24px", maxWidth: 460 }}>
            変更されたモジュールと関数の依存関係を可視化し、追加・削除・修正・リファクタの4クラスタに自動分類。レビューの&quot;認知負荷&quot;を下げます。
          </p>
          <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
            <button style={{ background: T.ink, color: "#fff", border: "none", padding: "12px 18px", borderRadius: 10, fontSize: 14, fontWeight: 500, fontFamily: "inherit", cursor: "pointer", display: "flex", alignItems: "center", gap: 8 }}>
              <svg width="14" height="14" viewBox="0 0 14 14" fill="currentColor"><path d="M7 0a7 7 0 0 0-2.21 13.64c.35.06.48-.15.48-.34v-1.2c-1.95.42-2.36-.94-2.36-.94-.32-.81-.78-1.03-.78-1.03-.64-.43.05-.42.05-.42.7.05 1.07.72 1.07.72.62 1.07 1.64.76 2.04.58.06-.45.24-.76.44-.94-1.55-.18-3.19-.78-3.19-3.46 0-.76.27-1.39.72-1.88-.07-.18-.31-.9.07-1.87 0 0 .59-.19 1.93.72A6.7 6.7 0 0 1 7 3.39a6.7 6.7 0 0 1 1.76.24c1.34-.91 1.93-.72 1.93-.72.38.97.14 1.69.07 1.87.45.49.72 1.12.72 1.88 0 2.69-1.64 3.28-3.2 3.46.25.22.48.65.48 1.31v1.94c0 .19.13.4.49.34A7 7 0 0 0 7 0z"/></svg>
              GitHubで始める
            </button>
            <button style={{ background: "transparent", color: T.ink, border: `1px solid ${T.borderStrong}`, padding: "12px 18px", borderRadius: 10, fontSize: 14, fontWeight: 500, fontFamily: "inherit", cursor: "pointer" }}>
              デモを見る →
            </button>
          </div>
          <div style={{ marginTop: 36, display: "flex", gap: 24, fontSize: 12, color: T.ink3 }}>
            <div><div style={{ fontSize: 22, fontWeight: 600, color: T.ink, fontVariantNumeric: "tabular-nums" }}>2,400+</div><div>解析されたPR</div></div>
            <div style={{ width: 1, background: T.border }} />
            <div><div style={{ fontSize: 22, fontWeight: 600, color: T.ink, fontVariantNumeric: "tabular-nums" }}>38%</div><div>レビュー時間削減</div></div>
            <div style={{ width: 1, background: T.border }} />
            <div><div style={{ fontSize: 22, fontWeight: 600, color: T.ink, fontVariantNumeric: "tabular-nums" }}>14</div><div>対応言語</div></div>
          </div>
        </div>

        {/* hero illust — ミニ依存グラフ */}
        <div style={{ position: "relative", height: 360, background: "#fff", borderRadius: 14, border: `1px solid ${T.border}`, overflow: "hidden", boxShadow: "0 12px 40px rgba(15,23,42,0.06)" }}>
          <div style={{ padding: "10px 14px", borderBottom: `1px solid ${T.border}`, display: "flex", alignItems: "center", gap: 8 }}>
            <div style={{ display: "flex", gap: 5 }}>
              <div style={{ width: 9, height: 9, borderRadius: 99, background: "#e5e5e5" }} />
              <div style={{ width: 9, height: 9, borderRadius: 99, background: "#e5e5e5" }} />
              <div style={{ width: 9, height: 9, borderRadius: 99, background: "#e5e5e5" }} />
            </div>
            <div style={{ fontFamily: T.mono, fontSize: 11, color: T.ink3 }}>kobold/checkout-service · #1284</div>
          </div>
          <svg viewBox="0 0 480 320" style={{ width: "100%", height: "calc(100% - 36px)" }}>
            {/* edges */}
            {[
              ["M120 80 C120 130, 200 130, 200 180", T.modified],
              ["M240 80 C240 130, 200 130, 200 180", T.added],
              ["M360 80 C360 130, 320 130, 320 180", T.added],
              ["M200 220 C200 260, 280 260, 280 280", T.removed],
              ["M320 220 C320 260, 280 260, 280 280", T.refactor],
              ["M120 220 C120 240, 160 250, 200 220", T.refactor],
            ].map((e, i) => (
              <path key={i} d={e[0]} stroke={e[1]} strokeWidth="1.5" fill="none" opacity="0.5" />
            ))}
            {/* nodes */}
            {[
              {x:80, y:60, label:"CheckoutForm", c:T.modified, bg:T.modifiedBg, bd:T.modifiedBd},
              {x:200, y:60, label:"PaymentMethodPicker", c:T.added, bg:T.addedBg, bd:T.addedBd},
              {x:320, y:60, label:"ReceiptModal", c:T.added, bg:T.addedBg, bd:T.addedBd},
              {x:160, y:170, label:"PaymentService", c:T.refactor, bg:T.refactorBg, bd:T.refactorBd},
              {x:280, y:170, label:"StripeAdapter", c:T.added, bg:T.addedBg, bd:T.addedBd},
              {x:80, y:170, label:"useCheckoutFlow", c:T.refactor, bg:T.refactorBg, bd:T.refactorBd},
              {x:240, y:270, label:"LegacyStripeClient", c:T.removed, bg:T.removedBg, bd:T.removedBd},
            ].map((n, i) => (
              <g key={i}>
                <rect x={n.x} y={n.y} width="80" height="32" rx="6" fill={n.bg} stroke={n.bd} />
                <text x={n.x + 40} y={n.y + 20} textAnchor="middle" fontSize="9" fill={n.c} fontFamily={T.font} fontWeight="600">{n.label}</text>
              </g>
            ))}
          </svg>
        </div>
      </div>

      {/* 4クラスタ説明 */}
      <div style={{ padding: "44px 36px 0", display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 14 }}>
        {[
          {label: "追加", desc: "新規モジュール", c: T.added, bg: T.addedBg, bd: T.addedBd, n: 11},
          {label: "削除", desc: "削除されたモジュール", c: T.removed, bg: T.removedBg, bd: T.removedBd, n: 4},
          {label: "修正", desc: "ロジック変更", c: T.modified, bg: T.modifiedBg, bd: T.modifiedBd, n: 13},
          {label: "リファクタ", desc: "振る舞い保存", c: T.refactor, bg: T.refactorBg, bd: T.refactorBd, n: 6},
        ].map((k) => (
          <div key={k.label} style={{ background: "#fff", border: `1px solid ${T.border}`, borderRadius: 12, padding: "14px 16px" }}>
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 6 }}>
              <div style={{ width: 28, height: 28, borderRadius: 8, background: k.bg, border: `1px solid ${k.bd}`, display: "flex", alignItems: "center", justifyContent: "center", color: k.c, fontWeight: 600, fontSize: 14 }}>●</div>
              <div style={{ fontFamily: T.mono, fontSize: 12, color: T.ink3 }}>n={k.n}</div>
            </div>
            <div style={{ fontSize: 14, fontWeight: 600, color: T.ink, marginTop: 4 }}>{k.label}</div>
            <div style={{ fontSize: 12, color: T.ink3, marginTop: 2 }}>{k.desc}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─────────────── PR Input (login済み) ───────────────
function PrInput() {
  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, overflow: "hidden" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "16px 28px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <Brand />
        <div style={{ display: "flex", alignItems: "center", gap: 12, fontSize: 13, color: T.ink2 }}>
          <span>履歴</span><span>設定</span>
          <div style={{ width: 28, height: 28, borderRadius: 99, background: T.brand, color: "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 600 }}>OY</div>
        </div>
      </div>

      <div style={{ padding: "60px 36px", maxWidth: 720, margin: "0 auto" }}>
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: T.mono, marginBottom: 8 }}>STEP 1 / 2</div>
        <h2 style={{ fontSize: 28, fontWeight: 600, letterSpacing: -0.6, margin: "0 0 8px" }}>解析するプルリクエストを指定</h2>
        <p style={{ fontSize: 14, color: T.ink2, margin: "0 0 28px" }}>GitHubのPR URLを貼り付けると、リポジトリをクローンし、依存解析を始めます。</p>

        <div style={{ background: "#fff", borderRadius: 12, border: `1px solid ${T.border}`, padding: 18, boxShadow: "0 1px 2px rgba(0,0,0,0.02)" }}>
          <label style={{ fontSize: 12, color: T.ink2, fontWeight: 500, display: "block", marginBottom: 6 }}>PR URL</label>
          <div style={{ display: "flex", gap: 8 }}>
            <div style={{ flex: 1, display: "flex", alignItems: "center", gap: 8, border: `1.5px solid ${T.brand}`, borderRadius: 8, padding: "10px 12px", background: "#fff" }}>
              <svg width="14" height="14" viewBox="0 0 14 14" fill={T.ink3}><path d="M7 0a7 7 0 0 0-2.21 13.64c.35.06.48-.15.48-.34v-1.2c-1.95.42-2.36-.94-2.36-.94-.32-.81-.78-1.03-.78-1.03-.64-.43.05-.42.05-.42.7.05 1.07.72 1.07.72.62 1.07 1.64.76 2.04.58.06-.45.24-.76.44-.94-1.55-.18-3.19-.78-3.19-3.46 0-.76.27-1.39.72-1.88-.07-.18-.31-.9.07-1.87 0 0 .59-.19 1.93.72A6.7 6.7 0 0 1 7 3.39a6.7 6.7 0 0 1 1.76.24c1.34-.91 1.93-.72 1.93-.72.38.97.14 1.69.07 1.87.45.49.72 1.12.72 1.88 0 2.69-1.64 3.28-3.2 3.46.25.22.48.65.48 1.31v1.94c0 .19.13.4.49.34A7 7 0 0 0 7 0z"/></svg>
              <input value="https://github.com/kobold/checkout-service/pull/1284" readOnly style={{ flex: 1, border: "none", outline: "none", fontFamily: T.mono, fontSize: 13, color: T.ink, background: "transparent" }} />
              <span style={{ color: T.added, fontSize: 12, fontWeight: 500 }}>✓ 有効</span>
            </div>
            <button style={{ background: T.ink, color: "#fff", border: "none", padding: "0 18px", borderRadius: 8, fontSize: 13, fontWeight: 500, fontFamily: "inherit", cursor: "pointer", whiteSpace: "nowrap" }}>解析を開始</button>
          </div>

          <div style={{ marginTop: 14, padding: 12, background: T.bgAlt, borderRadius: 8, fontSize: 12, color: T.ink2 }}>
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 4 }}>
              <span style={{ color: T.ink3 }}>リポジトリ</span><span style={{ fontFamily: T.mono }}>kobold/checkout-service</span>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 4 }}>
              <span style={{ color: T.ink3 }}>PR #</span><span style={{ fontFamily: T.mono }}>1284 · feat(payments): Stripe連携のリファクタ...</span>
            </div>
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <span style={{ color: T.ink3 }}>ブランチ</span><span style={{ fontFamily: T.mono }}>feat/payments-refactor → main</span>
            </div>
          </div>
        </div>

        <div style={{ marginTop: 22, fontSize: 12, color: T.ink3 }}>最近の解析</div>
        <div style={{ marginTop: 8, display: "flex", flexDirection: "column", gap: 6 }}>
          {[
            {repo: "kobold/checkout-service", n: 1278, t: "fix: Webhook署名検証", time: "1日前"},
            {repo: "kobold/api-gateway", n: 442, t: "refactor: ミドルウェアチェーン", time: "3日前"},
            {repo: "kobold/admin-ui", n: 88, t: "feat: ユーザー検索", time: "1週間前"},
          ].map((p, i) => (
            <div key={i} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 14px", background: "#fff", border: `1px solid ${T.border}`, borderRadius: 8, fontSize: 13 }}>
              <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
                <span style={{ fontFamily: T.mono, color: T.ink3, fontSize: 12 }}>{p.repo}#{p.n}</span>
                <span style={{ color: T.ink }}>{p.t}</span>
              </div>
              <span style={{ color: T.ink3, fontSize: 12 }}>{p.time}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─────────────── Analyzing (進捗) ───────────────
function Analyzing() {
  const steps = [
    { label: "リポジトリのクローン", status: "done", time: "0.8s" },
    { label: "依存グラフの構築 (AST解析)", status: "done", time: "2.4s" },
    { label: "差分との突き合わせ", status: "done", time: "1.1s" },
    { label: "クラスタ分類（4-way）", status: "active", time: "..." },
    { label: "可視化の生成", status: "pending", time: "" },
  ];
  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, overflow: "hidden", display: "flex", flexDirection: "column" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "16px 28px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <Brand />
        <div style={{ fontSize: 12, color: T.ink3, fontFamily: T.mono }}>kobold/checkout-service#1284</div>
      </div>
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", padding: 36 }}>
        <div style={{ width: 480 }}>
          <div style={{ display: "flex", gap: 14, alignItems: "center", marginBottom: 24 }}>
            <div style={{ width: 44, height: 44, borderRadius: 12, background: T.brandSoft, color: T.brand, display: "flex", alignItems: "center", justifyContent: "center" }}>
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="10" cy="10" r="7" strokeDasharray="32" strokeDashoffset="8"><animateTransform attributeName="transform" type="rotate" from="0 10 10" to="360 10 10" dur="1.4s" repeatCount="indefinite"/></circle></svg>
            </div>
            <div>
              <div style={{ fontSize: 20, fontWeight: 600, letterSpacing: -0.4 }}>解析中...</div>
              <div style={{ fontSize: 13, color: T.ink3 }}>34ファイルの依存関係を計算しています</div>
            </div>
          </div>

          <div style={{ background: "#fff", border: `1px solid ${T.border}`, borderRadius: 12, padding: 6 }}>
            {steps.map((s, i) => (
              <div key={i} style={{ display: "flex", alignItems: "center", gap: 12, padding: "10px 14px", borderBottom: i < steps.length - 1 ? `1px solid ${T.border}` : "none" }}>
                <div style={{ width: 18, height: 18, borderRadius: 99, display: "flex", alignItems: "center", justifyContent: "center",
                  background: s.status === "done" ? T.brand : s.status === "active" ? T.brandSoft : T.bgAlt,
                  color: s.status === "done" ? "#fff" : T.brand,
                  border: s.status === "active" ? `1.5px solid ${T.brand}` : "none" }}>
                  {s.status === "done" && <svg width="10" height="10" viewBox="0 0 10 10" fill="none" stroke="currentColor" strokeWidth="2"><path d="M2 5l2 2 4-4"/></svg>}
                  {s.status === "active" && <div style={{ width: 6, height: 6, borderRadius: 99, background: T.brand }} />}
                </div>
                <div style={{ flex: 1, fontSize: 13, color: s.status === "pending" ? T.ink3 : T.ink, fontWeight: s.status === "active" ? 500 : 400 }}>{s.label}</div>
                <div style={{ fontFamily: T.mono, fontSize: 11, color: T.ink3 }}>{s.time}</div>
              </div>
            ))}
          </div>

          <div style={{ marginTop: 14, fontSize: 12, color: T.ink3, fontFamily: T.mono }}>
            <span style={{ color: T.brand }}>►</span> module: <span style={{ color: T.ink2 }}>src/domain/payment/StripeAdapter.ts</span> ... 依存元 4 件を発見
          </div>
        </div>
      </div>
    </div>
  );
}

window.Landing = Landing;
window.PrInput = PrInput;
window.Analyzing = Analyzing;
window.Brand = Brand;
window.Logo = Logo;
window.T = T;
