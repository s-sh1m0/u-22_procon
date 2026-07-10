// DiffGraph. — landing page

const { useState } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "authState": "guest",
  "ctaTone": "ink"
}/*EDITMODE-END*/;

// ──────────────────────────────────────────────────────────────────────
// Icons
// ──────────────────────────────────────────────────────────────────────

// The A4·4 mark (refined) — production app icon.
function AppMark({ size = 32, rx = 0.22, bg = "var(--brand)", fg = "white", accent = "white", shadow = false }) {
  const radius = rx * 192;
  return (
    <svg width={size} height={size} viewBox="0 0 192 192" xmlns="http://www.w3.org/2000/svg" style={shadow ? { filter: "drop-shadow(0 8px 24px rgba(15,118,110,0.28)) drop-shadow(0 2px 4px rgba(28,25,23,0.08))" } : undefined}>
      <rect width="192" height="192" rx={radius} fill={bg} />
      <g stroke={fg} strokeWidth="8" strokeLinecap="round" fill="none" opacity="0.85">
        <line x1="52" y1="52" x2="52" y2="140" />
        <line x1="52" y1="52" x2="90" y2="76" />
        <line x1="52" y1="140" x2="90" y2="116" />
        <line x1="90" y1="76" x2="90" y2="116" />
        <line x1="52" y1="96" x2="90" y2="76" />
        <line x1="52" y1="96" x2="90" y2="116" />
      </g>
      <g stroke={fg} strokeWidth="10" strokeLinecap="round" fill="none">
        <line x1="90" y1="76" x2="146" y2="96" />
        <line x1="90" y1="116" x2="146" y2="96" />
      </g>
      <circle cx="52" cy="52" r="11" fill={fg} />
      <circle cx="52" cy="96" r="11" fill={fg} />
      <circle cx="52" cy="140" r="11" fill={fg} />
      <circle cx="90" cy="76" r="11" fill={fg} />
      <circle cx="90" cy="116" r="11" fill={fg} />
      <circle cx="146" cy="96" r="16" fill={accent} />
    </svg>
  );
}

function Brand({ size = 18, color = "var(--ink)" }) {
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
      <AppMark size={size + 12} />
      <span style={{ fontSize: size, fontWeight: 600, letterSpacing: "-0.025em", color }}>
        DiffGraph<span style={{ color: "var(--brand)" }}>.</span>
      </span>
    </div>
  );
}

function GitHubGlyph({ size = 14, color = "currentColor" }) {
  return (
    <svg width={size} height={size} viewBox="0 0 14 14" fill={color}>
      <path d="M7 0a7 7 0 0 0-2.21 13.64c.35.06.48-.15.48-.34v-1.2c-1.95.42-2.36-.94-2.36-.94-.32-.81-.78-1.03-.78-1.03-.64-.43.05-.42.05-.42.7.05 1.07.72 1.07.72.62 1.07 1.64.76 2.04.58.06-.45.24-.76.44-.94-1.55-.18-3.19-.78-3.19-3.46 0-.76.27-1.39.72-1.88-.07-.18-.31-.9.07-1.87 0 0 .59-.19 1.93.72A6.7 6.7 0 0 1 7 3.39a6.7 6.7 0 0 1 1.76.24c1.34-.91 1.93-.72 1.93-.72.38.97.14 1.69.07 1.87.45.49.72 1.12.72 1.88 0 2.69-1.64 3.28-3.2 3.46.25.22.48.65.48 1.31v1.94c0 .19.13.4.49.34A7 7 0 0 0 7 0z"/>
    </svg>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Nav
// ──────────────────────────────────────────────────────────────────────

function Nav({ authState }) {
  return (
    <nav style={{ borderBottom: "1px solid var(--line)", background: "rgba(250,250,249,0.85)", backdropFilter: "blur(8px)", position: "sticky", top: 0, zIndex: 50 }}>
      <div className="wrap" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", height: 64 }}>
        <Brand size={17} />
        <div style={{ display: "flex", alignItems: "center", gap: 28, fontSize: 13, color: "var(--ink-2)" }}>
          <a style={navLinkStyle}>機能</a>
          <a style={navLinkStyle}>仕組み</a>
          <a style={navLinkStyle}>クラスタリング</a>
          <a style={navLinkStyle} href="#"><GitHubGlyph size={13} /> GitHub</a>
          <div style={{ width: 1, height: 18, background: "var(--line)" }} />
          {authState === "guest" ? (
            <button style={{ ...btn, ...btnInk, height: 34, padding: "0 14px", fontSize: 13 }}>ログイン</button>
          ) : (
            <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
              <span className="mono" style={{ fontSize: 12, color: "var(--ink-3)" }}>shimo-dev</span>
              <div style={{ width: 30, height: 30, borderRadius: 99, background: "linear-gradient(135deg, var(--brand) 0%, var(--brand-deep) 100%)", color: "white", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 11, fontWeight: 600 }}>S</div>
            </div>
          )}
        </div>
      </div>
    </nav>
  );
}

const navLinkStyle = { display: "inline-flex", alignItems: "center", gap: 5, color: "var(--ink-2)", textDecoration: "none", cursor: "pointer" };

// ──────────────────────────────────────────────────────────────────────
// Buttons
// ──────────────────────────────────────────────────────────────────────

const btn = {
  display: "inline-flex", alignItems: "center", justifyContent: "center", gap: 8,
  height: 44, padding: "0 18px", borderRadius: 10,
  fontSize: 14, fontWeight: 500, fontFamily: "inherit", letterSpacing: "-0.005em",
  cursor: "pointer", border: "1px solid transparent", whiteSpace: "nowrap",
  transition: "transform 80ms ease, box-shadow 80ms ease",
};
const btnInk = { background: "var(--ink)", color: "white" };
const btnBrand = { background: "var(--brand)", color: "white" };
const btnGhost = { background: "transparent", color: "var(--ink)", border: "1px solid var(--line-strong)" };

// ──────────────────────────────────────────────────────────────────────
// Hero
// ──────────────────────────────────────────────────────────────────────

function Hero({ authState, ctaTone }) {
  const ctaBg = ctaTone === "brand" ? btnBrand : btnInk;
  return (
    <section style={{ position: "relative", overflow: "hidden", borderBottom: "1px solid var(--line)" }}>
      {/* faint dotted grid */}
      <div style={{ position: "absolute", inset: 0, backgroundImage: "radial-gradient(var(--line-strong) 1px, transparent 1px)", backgroundSize: "24px 24px", opacity: 0.35, maskImage: "radial-gradient(80% 60% at 50% 30%, black 0%, transparent 75%)" }} />
      <div className="wrap" style={{ position: "relative", padding: "88px 36px 96px", display: "grid", gridTemplateColumns: "1.05fr 1fr", gap: 64, alignItems: "center" }}>
        <div>
          <div style={pillStyle}>
            <span style={{ width: 6, height: 6, borderRadius: 99, background: "var(--brand)" }} />
            <span>Go プロジェクト対応 · ベータ公開中</span>
          </div>
          <h1 style={{ fontSize: 60, lineHeight: 1.02, fontWeight: 600, letterSpacing: "-0.04em", margin: "20px 0 18px", textWrap: "pretty" }}>
            プルリクを、<br/><span style={{ color: "var(--brand)" }}>構造</span>として読む。
          </h1>
          <p style={{ fontSize: 17, lineHeight: 1.6, color: "var(--ink-2)", margin: "0 0 32px", maxWidth: 500 }}>
            変更ファイルの羅列ではなく、モジュールの依存関係として PR を可視化。大規模レビューも、意味のあるクラスタに自動分割します。
          </p>

          {/* === Auth-aware CTA === */}
          {authState === "guest" ? (
            <GuestCTA />
          ) : (
            <PrInputCTA />
          )}
        </div>

        <HeroIllustration />
      </div>
    </section>
  );
}

const pillStyle = { display: "inline-flex", alignItems: "center", gap: 8, padding: "6px 12px", background: "var(--brand-soft)", color: "var(--brand-ink)", borderRadius: 99, fontSize: 12, fontWeight: 500, border: "1px solid #99f6e4" };

// ──────────────────────────────────────────────────────────────────────
// CTA — Guest (GitHub OAuth)
// ──────────────────────────────────────────────────────────────────────

function GuestCTA() {
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 14, alignItems: "flex-start" }}>
      <div style={{ display: "flex", gap: 10, alignItems: "center" }}>
        <a href="/auth/github" style={{ ...btn, ...btnInk, height: 52, padding: "0 22px", fontSize: 15, textDecoration: "none", boxShadow: "0 4px 14px rgba(28,25,23,0.18), inset 0 1px 0 rgba(255,255,255,0.08)" }}>
          <GitHubGlyph size={16} />
          GitHub で始める
          <span style={{ opacity: 0.5, marginLeft: 4 }}>→</span>
        </a>
        <a href="#demo" style={{ ...btn, ...btnGhost, height: 52, padding: "0 22px", fontSize: 15, textDecoration: "none" }}>
          デモを見る
        </a>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: 14, fontSize: 12, color: "var(--ink-3)" }}>
        <span style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
          <CheckGlyph color="var(--added)" />
          無料で使い始められる
        </span>
        <span style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
          <CheckGlyph color="var(--added)" />
          パブリックリポジトリのみ要求
        </span>
      </div>
    </div>
  );
}

function CheckGlyph({ color = "var(--added)" }) {
  return (
    <svg width="11" height="11" viewBox="0 0 12 12" fill="none" stroke={color} strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M2 6.5 L4.8 9 L10 3.5"/>
    </svg>
  );
}

// ──────────────────────────────────────────────────────────────────────
// CTA — Logged in (PR URL input)
// ──────────────────────────────────────────────────────────────────────

const URL_RE = /^https:\/\/github\.com\/([^/]+)\/([^/]+)\/pull\/(\d+)/;

function parsePr(url) {
  const m = url.trim().match(URL_RE);
  if (!m) return null;
  return { owner: m[1], repo: m[2], number: m[3] };
}

function PrInputCTA() {
  const [url, setUrl] = useState("https://github.com/kobold/checkout-service/pull/1284");
  const [submitted, setSubmitted] = useState(false);
  const parsed = parsePr(url);
  const valid = !!parsed;
  const empty = url.trim() === "";

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 14, maxWidth: 560 }}>
      <div style={{ background: "white", border: "1px solid var(--line)", borderRadius: 14, padding: 18, boxShadow: "0 1px 2px rgba(0,0,0,0.02), 0 8px 28px rgba(15,118,110,0.06)" }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 8 }}>
          <label className="mono" style={{ fontSize: 11, color: "var(--ink-3)", letterSpacing: "0.04em", textTransform: "uppercase" }}>Step 1 / 2 · PR URL</label>
          {valid && (
            <span style={{ display: "inline-flex", alignItems: "center", gap: 4, fontSize: 11, color: "var(--added)", fontWeight: 500 }}>
              <CheckGlyph color="var(--added)" /> 有効な URL
            </span>
          )}
        </div>
        <form onSubmit={(e) => { e.preventDefault(); if (valid) setSubmitted(true); }}>
          <div style={{ display: "flex", gap: 8 }}>
            <div style={{
              flex: 1, display: "flex", alignItems: "center", gap: 10,
              border: `1.5px solid ${valid ? "var(--brand)" : "var(--line-strong)"}`,
              borderRadius: 10, padding: "12px 14px", background: "var(--bg)",
              transition: "border-color 100ms ease, box-shadow 100ms ease",
              boxShadow: valid ? "0 0 0 4px rgba(15,118,110,0.08)" : "none",
            }}>
              <GitHubGlyph size={15} color="var(--ink-3)" />
              <input
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://github.com/owner/repo/pull/123"
                className="mono"
                style={{ flex: 1, border: "none", outline: "none", fontSize: 13, color: "var(--ink)", background: "transparent" }}
              />
            </div>
            <button type="submit" disabled={!valid} style={{ ...btn, ...btnInk, height: 48, padding: "0 18px", fontSize: 14, opacity: valid ? 1 : 0.45, cursor: valid ? "pointer" : "not-allowed" }}>
              解析を開始 →
            </button>
          </div>
          {!empty && !valid && (
            <div style={{ marginTop: 10, fontSize: 12, color: "var(--removed)" }}>
              有効な GitHub PR URL を入力してください
            </div>
          )}
          {valid && (
            <div style={{ marginTop: 12, padding: "10px 14px", background: "var(--bg-alt)", borderRadius: 8, fontSize: 12, color: "var(--ink-2)", display: "grid", gridTemplateColumns: "auto 1fr", gap: "4px 16px" }}>
              <span style={{ color: "var(--ink-3)" }}>リポジトリ</span>
              <span className="mono" style={{ textAlign: "right" }}>{parsed.owner}/{parsed.repo}</span>
              <span style={{ color: "var(--ink-3)" }}>PR</span>
              <span className="mono" style={{ textAlign: "right" }}>#{parsed.number}</span>
            </div>
          )}
        </form>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Hero illustration — mini dependency graph
// ──────────────────────────────────────────────────────────────────────

function HeroIllustration() {
  return (
    <div style={{ position: "relative" }}>
      <div style={{
        position: "relative",
        background: "white", borderRadius: 16, border: "1px solid var(--line)",
        boxShadow: "0 24px 60px rgba(15,23,42,0.08), 0 2px 8px rgba(15,23,42,0.04)",
        overflow: "hidden",
      }}>
        {/* window chrome */}
        <div style={{ padding: "12px 16px", borderBottom: "1px solid var(--line)", display: "flex", alignItems: "center", gap: 10 }}>
          <div style={{ display: "flex", gap: 6 }}>
            {["#ef4444", "#f59e0b", "#10b981"].map((c) => (
              <div key={c} style={{ width: 10, height: 10, borderRadius: 99, background: c, opacity: 0.55 }} />
            ))}
          </div>
          <div className="mono" style={{ fontSize: 11, color: "var(--ink-3)" }}>kobold/checkout-service · #1284 · 34 files</div>
          <div style={{ marginLeft: "auto", fontSize: 11, color: "var(--ink-3)" }}>解析完了</div>
        </div>

        {/* graph */}
        <svg viewBox="0 0 520 360" style={{ width: "100%", display: "block" }}>
          {/* layer labels */}
          <g fontFamily='"JetBrains Mono", monospace' fontSize="9" fill="#a8a29e">
            <text x="16" y="44">UI</text>
            <text x="16" y="160">DOMAIN</text>
            <text x="16" y="276">DATA</text>
          </g>
          {/* layer dividers */}
          <g stroke="#e7e5e4" strokeWidth="1" strokeDasharray="2 4">
            <line x1="0" y1="100" x2="520" y2="100"/>
            <line x1="0" y1="220" x2="520" y2="220"/>
          </g>

          {/* edges */}
          <g fill="none" strokeWidth="1.6" opacity="0.65">
            <path d="M120 60 C120 110, 180 130, 200 160" stroke="#d97706"/>
            <path d="M240 60 C240 110, 220 130, 220 160" stroke="#16a34a"/>
            <path d="M360 60 C360 110, 340 130, 320 160" stroke="#16a34a"/>
            <path d="M200 200 C200 240, 260 250, 280 280" stroke="#dc2626"/>
            <path d="M320 200 C320 240, 280 250, 280 280" stroke="#7c3aed"/>
            <path d="M120 200 C120 230, 160 250, 200 220" stroke="#7c3aed"/>
            <path d="M380 200 C380 240, 340 260, 320 280" stroke="#d97706"/>
          </g>

          {/* nodes */}
          {[
            {x:80, y:36, w:90, label:"CheckoutForm",       c:"#d97706", bg:"#fef3c7", bd:"#fcd34d"},
            {x:200, y:36, w:100, label:"PaymentPicker",    c:"#16a34a", bg:"#dcfce7", bd:"#86efac"},
            {x:320, y:36, w:84, label:"ReceiptModal",      c:"#16a34a", bg:"#dcfce7", bd:"#86efac"},
            {x:160, y:152, w:96, label:"PaymentService",   c:"#7c3aed", bg:"#ede9fe", bd:"#c4b5fd"},
            {x:280, y:152, w:90, label:"StripeAdapter",    c:"#16a34a", bg:"#dcfce7", bd:"#86efac"},
            {x:80, y:152, w:88, label:"useCheckoutFlow",   c:"#7c3aed", bg:"#ede9fe", bd:"#c4b5fd"},
            {x:340, y:152, w:88, label:"FeeCalculator",    c:"#d97706", bg:"#fef3c7", bd:"#fcd34d"},
            {x:240, y:272, w:104, label:"LegacyStripeClient", c:"#dc2626", bg:"#fee2e2", bd:"#fca5a5"},
            {x:80, y:272, w:88, label:"PaymentRepo",      c:"#d97706", bg:"#fef3c7", bd:"#fcd34d"},
          ].map((n, i) => (
            <g key={i}>
              <rect x={n.x} y={n.y} width={n.w} height={30} rx="6" fill={n.bg} stroke={n.bd}/>
              {/* left bar */}
              <rect x={n.x} y={n.y} width={3} height={30} rx="1" fill={n.c}/>
              <text x={n.x + 10} y={n.y + 19} fontSize="9.5" fill={n.c} fontFamily='"Inter Tight", sans-serif' fontWeight="600">{n.label}</text>
            </g>
          ))}
        </svg>

        {/* cluster legend */}
        <div style={{ borderTop: "1px solid var(--line)", padding: "10px 14px", display: "flex", justifyContent: "space-between", alignItems: "center", background: "var(--bg-alt)" }}>
          <div style={{ display: "flex", gap: 14, fontSize: 11 }}>
            {[
              ["#16a34a", "Cluster A"],
              ["#dc2626", "Cluster B"],
              ["#d97706", "Cluster C"],
              ["#7c3aed", "Cluster D"],
            ].map(([c, l]) => (
              <span key={l} style={{ display: "inline-flex", alignItems: "center", gap: 5, color: "var(--ink-2)" }}>
                <span style={{ width: 8, height: 8, borderRadius: 2, background: c }} />
                {l}
              </span>
            ))}
          </div>
          <span className="mono" style={{ fontSize: 10, color: "var(--ink-3)" }}>2.4s</span>
        </div>
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Section · Smart Clustering
// ──────────────────────────────────────────────────────────────────────

function ClusterSection() {
  const modes = [
    { label: "Louvain",       desc: "依存関係の密度からコミュニティを自動検出",   icon: "{}", badge: "既定" },
    { label: "パッケージ",     desc: "Go パッケージ単位でグルーピング",         icon: "/",  badge: "" },
    { label: "ファイル",       desc: "ソースファイル単位でグルーピング",         icon: "[]", badge: "" },
  ];
  return (
    <section id="showcase" style={{ padding: "112px 0 88px", borderBottom: "1px solid var(--line)" }}>
      <div className="wrap">
        <SectionHead eyebrow="Smart clustering" title="依存関係から、自動でクラスタを生成。" lead="Louvain アルゴリズムが呼び出しグラフのコミュニティ構造を検出し、関連する関数群を意味のあるまとまりに自動分割。パッケージ・ファイル単位への切替もワンクリック。" />
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16, marginTop: 48 }}>
          {modes.map((m) => (
            <div key={m.label} style={{ background: "white", border: "1px solid var(--line)", borderRadius: 14, padding: 22, position: "relative", overflow: "hidden" }}>
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
                <div className="mono" style={{ width: 38, height: 38, borderRadius: 10, background: "var(--brand-soft)", border: "1px solid #99f6e4", display: "flex", alignItems: "center", justifyContent: "center", color: "var(--brand)", fontWeight: 600, fontSize: 14 }}>{m.icon}</div>
                {m.badge && (
                  <span style={{ fontSize: 10, fontWeight: 500, color: "var(--brand)", background: "var(--brand-soft)", padding: "2px 8px", borderRadius: 99 }}>{m.badge}</span>
                )}
              </div>
              <div style={{ fontSize: 16, fontWeight: 600, color: "var(--ink)", marginTop: 16, letterSpacing: "-0.01em" }}>{m.label}</div>
              <div style={{ fontSize: 13, color: "var(--ink-3)", marginTop: 4, lineHeight: 1.5 }}>{m.desc}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Section · How it works
// ──────────────────────────────────────────────────────────────────────

function HowItWorks() {
  const steps = [
    {
      n: "01",
      title: "PR を貼り付ける",
      desc: "GitHub の PR URL を貼るだけ。OAuth 連携でプライベートリポジトリも対応。",
      mock: <StepMock1 />,
    },
    {
      n: "02",
      title: "依存グラフを構築",
      desc: "go/ast で関数・型を抽出し、呼び出しグラフを双方向に辿ります。変更行を AST ノードにマッピング。",
      mock: <StepMock2 />,
    },
    {
      n: "03",
      title: "クラスタで構造を把握",
      desc: "Louvain アルゴリズムが関連する関数群をクラスタに自動分割。変更ノードを優先度順にナビゲートして効率的にレビュー。",
      mock: <StepMock3 />,
    },
  ];

  return (
    <section style={{ padding: "112px 0", borderBottom: "1px solid var(--line)", background: "linear-gradient(180deg, var(--bg) 0%, white 100%)" }}>
      <div className="wrap">
        <SectionHead eyebrow="How it works" title="3 ステップで、PR を構造化。" lead="バックエンドが Go AST を解析し、フロントが xyflow でレンダリング。" />
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 20, marginTop: 56 }}>
          {steps.map((s) => (
            <div key={s.n} style={{ background: "white", border: "1px solid var(--line)", borderRadius: 16, padding: 24, display: "flex", flexDirection: "column", gap: 16, minHeight: 380 }}>
              <div style={{ display: "flex", alignItems: "baseline", gap: 12, flexWrap: "nowrap" }}>
                <span className="mono" style={{ fontSize: 11, color: "var(--brand)", letterSpacing: "0.06em", flexShrink: 0 }}>{s.n}</span>
                <h3 style={{ fontSize: 19, fontWeight: 600, letterSpacing: "-0.015em", margin: 0, whiteSpace: "nowrap" }}>{s.title}</h3>
              </div>
              <p style={{ fontSize: 13.5, color: "var(--ink-2)", lineHeight: 1.6, margin: 0 }}>{s.desc}</p>
              <div style={{ marginTop: "auto", flex: 1, display: "flex", alignItems: "center", justifyContent: "center", background: "var(--bg-alt)", border: "1px solid var(--line)", borderRadius: 12, padding: 16 }}>
                {s.mock}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function StepMock1() {
  return (
    <div style={{ width: "100%" }}>
      <div className="mono" style={{ fontSize: 10, color: "var(--ink-3)", marginBottom: 6 }}>PR URL</div>
      <div style={{ background: "white", border: "1.5px solid var(--brand)", borderRadius: 8, padding: "8px 10px", display: "flex", alignItems: "center", gap: 8 }}>
        <GitHubGlyph size={12} color="var(--ink-3)" />
        <span className="mono" style={{ fontSize: 11, color: "var(--ink-2)", flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>github.com/kobold/checkout-service/pull/1284</span>
        <span style={{ fontSize: 10, color: "var(--added)", fontWeight: 500 }}>✓</span>
      </div>
      <div style={{ marginTop: 8, display: "flex", justifyContent: "flex-end" }}>
        <div style={{ background: "var(--ink)", color: "white", padding: "6px 12px", borderRadius: 6, fontSize: 11 }}>解析を開始 →</div>
      </div>
    </div>
  );
}

function StepMock2() {
  return (
    <svg viewBox="0 0 240 160" style={{ width: "100%" }}>
      <g stroke="var(--line-strong)" strokeWidth="1" fill="none">
        <line x1="60" y1="40" x2="120" y2="80"/>
        <line x1="180" y1="40" x2="120" y2="80"/>
        <line x1="60" y1="120" x2="120" y2="80"/>
        <line x1="180" y1="120" x2="120" y2="80"/>
      </g>
      <g>
        <rect x="36" y="28" width="48" height="24" rx="4" fill="white" stroke="var(--line-strong)"/>
        <text x="60" y="44" textAnchor="middle" fontSize="9" fill="var(--ink-2)" fontFamily="Inter Tight">handler.go</text>
        <rect x="156" y="28" width="48" height="24" rx="4" fill="white" stroke="var(--line-strong)"/>
        <text x="180" y="44" textAnchor="middle" fontSize="9" fill="var(--ink-2)" fontFamily="Inter Tight">model.go</text>
        <rect x="36" y="108" width="48" height="24" rx="4" fill="white" stroke="var(--line-strong)"/>
        <text x="60" y="124" textAnchor="middle" fontSize="9" fill="var(--ink-2)" fontFamily="Inter Tight">repo.go</text>
        <rect x="156" y="108" width="48" height="24" rx="4" fill="white" stroke="var(--line-strong)"/>
        <text x="180" y="124" textAnchor="middle" fontSize="9" fill="var(--ink-2)" fontFamily="Inter Tight">util.go</text>
        <circle cx="120" cy="80" r="14" fill="var(--brand)"/>
        <text x="120" y="84" textAnchor="middle" fontSize="9" fill="white" fontFamily="JetBrains Mono" fontWeight="600">AST</text>
      </g>
    </svg>
  );
}

function StepMock3() {
  return (
    <svg viewBox="0 0 220 140" style={{ width: "100%" }}>
      <rect x="14" y="10" width="92" height="52" rx="8" fill="#f0fdfa" stroke="#99f6e4" strokeDasharray="3 2"/>
      <rect x="124" y="10" width="82" height="52" rx="8" fill="#f5f3ff" stroke="#c4b5fd" strokeDasharray="3 2"/>
      <rect x="69" y="74" width="82" height="52" rx="8" fill="#fffbeb" stroke="#fcd34d" strokeDasharray="3 2"/>
      {[
        {x:20, y:16, label:"Cluster A", color:"#0f766e"},
        {x:130, y:16, label:"Cluster B", color:"#7c3aed"},
        {x:75, y:80, label:"Cluster C", color:"#d97706"},
      ].map((n) => (
        <g key={n.label}>
          <circle cx={n.x+20} cy={n.y+22} r="6" fill={n.color}/>
          <circle cx={n.x+38} cy={n.y+18} r="4" fill={n.color} opacity="0.6"/>
          <circle cx={n.x+52} cy={n.y+26} r="5" fill={n.color} opacity="0.8"/>
          <text x={n.x+36} y={n.y+44} textAnchor="middle" fontSize="8" fill="#78716c" fontFamily="Inter Tight">{n.label}</text>
        </g>
      ))}
      <g stroke="#d6d3d1" strokeWidth="1" fill="none">
        <path d="M60 50 Q80 68 90 80"/>
        <path d="M150 58 Q140 68 130 80"/>
      </g>
    </svg>
  );
}

function SectionHead({ eyebrow, title, lead }) {
  return (
    <div style={{ maxWidth: 720 }}>
      <div className="mono" style={{ fontSize: 11, color: "var(--brand)", letterSpacing: "0.08em", textTransform: "uppercase", marginBottom: 14 }}>{eyebrow}</div>
      <h2 style={{ fontSize: 38, fontWeight: 600, letterSpacing: "-0.03em", lineHeight: 1.1, margin: "0 0 14px", textWrap: "pretty" }}>{title}</h2>
      <p style={{ fontSize: 15.5, color: "var(--ink-2)", lineHeight: 1.6, margin: 0, maxWidth: 580 }}>{lead}</p>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Section · Features (3-up)
// ──────────────────────────────────────────────────────────────────────

function Features() {
  const items = [
    {
      title: "影響範囲を hop で測る",
      desc: "変更ノードから caller/callee を双方向に辿り、何が壊れる可能性があるかを段階的に可視化。",
      art: <FeatureArt1 />,
    },
    {
      title: "AST 単位の diff",
      desc: "行ベースではなく関数・型単位で差分を表示。Monaco エディタによる VSCode 同等の体験。",
      art: <FeatureArt2 />,
    },
    {
      title: "新規循環参照を自動検出",
      desc: "PR で新たに生じた循環依存を赤でハイライト。構造リスクをレビュー前に把握できます。",
      art: <FeatureArt3 />,
    },
  ];
  return (
    <section style={{ padding: "112px 0", borderBottom: "1px solid var(--line)" }}>
      <div className="wrap">
        <SectionHead eyebrow="Features" title="グラフだけじゃない。" lead="解析の仕方そのものを設計し直したから、レビューの一歩一歩が短くなります。" />
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 20, marginTop: 48 }}>
          {items.map((f) => (
            <div key={f.title} style={{ background: "white", border: "1px solid var(--line)", borderRadius: 16, padding: 24, display: "flex", flexDirection: "column", gap: 16 }}>
              <div style={{ height: 140, background: "var(--brand-tint)", border: "1px solid #ccfbf1", borderRadius: 10, display: "flex", alignItems: "center", justifyContent: "center", padding: 16 }}>{f.art}</div>
              <h3 style={{ fontSize: 18, fontWeight: 600, letterSpacing: "-0.015em", margin: 0 }}>{f.title}</h3>
              <p style={{ fontSize: 13.5, color: "var(--ink-2)", lineHeight: 1.6, margin: 0 }}>{f.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function FeatureArt1() {
  return (
    <svg viewBox="0 0 200 120" style={{ width: "100%", height: "100%" }}>
      <circle cx="100" cy="60" r="50" fill="none" stroke="var(--brand-light)" strokeWidth="1" strokeDasharray="2 3" opacity="0.6"/>
      <circle cx="100" cy="60" r="32" fill="none" stroke="var(--brand-light)" strokeWidth="1" strokeDasharray="2 3" opacity="0.8"/>
      <circle cx="100" cy="60" r="14" fill="var(--brand)"/>
      <text x="100" y="64" textAnchor="middle" fontSize="9" fill="white" fontFamily="JetBrains Mono" fontWeight="600">0</text>
      <text x="100" y="22" textAnchor="middle" fontSize="9" fill="var(--brand)" fontFamily="JetBrains Mono">1 hop</text>
      <text x="100" y="6" textAnchor="middle" fontSize="9" fill="var(--brand)" fontFamily="JetBrains Mono" opacity="0.7">2 hop</text>
    </svg>
  );
}

function FeatureArt2() {
  return (
    <div className="mono" style={{ fontSize: 10, color: "var(--ink-2)", width: "100%" }}>
      <div style={{ background: "white", borderRadius: 6, border: "1px solid var(--line)", overflow: "hidden" }}>
        <div style={{ padding: "4px 8px", borderBottom: "1px solid var(--line)", color: "var(--ink-3)" }}>PaymentService.go</div>
        <div style={{ padding: "4px 8px", background: "var(--added-bg)", color: "var(--added)" }}>+ func (s *PaymentService) Charge(...</div>
        <div style={{ padding: "4px 8px", color: "var(--ink-2)" }}>&nbsp;&nbsp;return s.adapter.Process(...)</div>
        <div style={{ padding: "4px 8px", background: "var(--removed-bg)", color: "var(--removed)" }}>− func (s *PaymentService) Pay(...</div>
        <div style={{ padding: "4px 8px", color: "var(--ink-2)" }}>}</div>
      </div>
    </div>
  );
}

function FeatureArt3() {
  return (
    <svg viewBox="0 0 200 120" style={{ width: "100%", height: "100%" }}>
      <g stroke="#d6d3d1" strokeWidth="1" fill="none">
        <path d="M60 30 L100 60"/>
        <path d="M140 30 L100 60"/>
        <path d="M100 60 L60 90"/>
        <path d="M100 60 L140 90"/>
      </g>
      <rect x="40" y="18" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1"/>
      <text x="60" y="34" textAnchor="middle" fontSize="8" fill="#44403c" fontFamily="JetBrains Mono">funcA</text>
      <rect x="120" y="18" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1"/>
      <text x="140" y="34" textAnchor="middle" fontSize="8" fill="#44403c" fontFamily="JetBrains Mono">funcB</text>
      <circle cx="100" cy="60" r="14" fill="#ef4444"/>
      <text x="100" y="64" textAnchor="middle" fontSize="8" fill="white" fontFamily="JetBrains Mono" fontWeight="600">↻</text>
      <rect x="40" y="78" width="40" height="24" rx="4" fill="white" stroke="#d6d3d1"/>
      <text x="60" y="94" textAnchor="middle" fontSize="8" fill="#44403c" fontFamily="JetBrains Mono">funcC</text>
      <rect x="120" y="78" width="40" height="24" rx="4" fill="#fee2e2" stroke="#fca5a5"/>
      <text x="140" y="94" textAnchor="middle" fontSize="8" fill="#dc2626" fontFamily="JetBrains Mono">funcD</text>
      <path d="M140 78 L100 74" stroke="#ef4444" strokeWidth="2" fill="none"/>
    </svg>
  );
}

// ──────────────────────────────────────────────────────────────────────
// CTA strip
// ──────────────────────────────────────────────────────────────────────

function CTAStrip({ authState }) {
  return (
    <section style={{ padding: "104px 0", background: "var(--ink)", color: "white", position: "relative", overflow: "hidden" }}>
      <div style={{ position: "absolute", inset: 0, backgroundImage: "radial-gradient(120% 80% at 80% 50%, rgba(15,118,110,0.32) 0%, transparent 60%), radial-gradient(80% 60% at 0% 100%, rgba(94,234,212,0.08) 0%, transparent 60%)" }} />
      <div className="wrap" style={{ position: "relative", display: "grid", gridTemplateColumns: "1.4fr 1fr", alignItems: "center", gap: 56 }}>
        <div>
          <h2 style={{ fontSize: 48, fontWeight: 600, letterSpacing: "-0.035em", lineHeight: 1.05, margin: "0 0 16px" }}>
            次の大規模 PR から、<br/>
            <span style={{ color: "var(--brand-light)" }}>構造を読もう。</span>
          </h2>
          <p style={{ fontSize: 16, color: "rgba(255,255,255,0.65)", margin: "0 0 32px", maxWidth: 480, lineHeight: 1.6 }}>
            GitHub と連携するだけですぐに使えます。パブリックリポジトリは無料で解析可能。
          </p>
          {authState === "guest" ? (
            <a href="/auth/github" style={{ ...btn, background: "white", color: "var(--ink)", height: 52, padding: "0 22px", fontSize: 15, textDecoration: "none", boxShadow: "0 4px 14px rgba(0,0,0,0.25)" }}>
              <GitHubGlyph size={16} />
              GitHub で始める
            </a>
          ) : (
            <a href="#top" style={{ ...btn, background: "var(--brand)", color: "white", height: 52, padding: "0 22px", fontSize: 15, textDecoration: "none", boxShadow: "0 4px 14px rgba(15,118,110,0.45)" }}>
              新しい PR を解析する →
            </a>
          )}
        </div>
        <div style={{ display: "flex", justifyContent: "center" }}>
          <AppMark size={200} bg="var(--brand)" fg="white" accent="white" shadow />
        </div>
      </div>
    </section>
  );
}

// ──────────────────────────────────────────────────────────────────────
// Footer
// ──────────────────────────────────────────────────────────────────────

function Footer() {
  const cols = [
    { title: "プロダクト", links: [{ label: "機能", href: "#features" }, { label: "仕組み", href: "#how" }] },
    { title: "リソース", links: [{ label: "GitHub" }, { label: "README" }] },
  ];
  return (
    <footer style={{ padding: "72px 0 40px", background: "white", borderTop: "1px solid var(--line)" }}>
      <div className="wrap" style={{ display: "grid", gridTemplateColumns: "1.6fr repeat(2, 1fr)", gap: 48 }}>
        <div>
          <Brand size={16} />
          <p style={{ fontSize: 13, color: "var(--ink-3)", margin: "14px 0 20px", lineHeight: 1.6, maxWidth: 280 }}>
            PR を変更行ではなく依存グラフとして読むためのレビュー支援ツール。
          </p>
          <div style={{ display: "flex", gap: 10 }}>
            <a style={{ ...btnGhost, ...btn, height: 32, padding: "0 12px", fontSize: 12 }}>
              <GitHubGlyph size={12} />
              GitHub
            </a>
          </div>
        </div>
        {cols.map((c) => (
          <div key={c.title}>
            <div style={{ fontSize: 12, fontWeight: 600, color: "var(--ink-2)", marginBottom: 12, letterSpacing: "-0.005em" }}>{c.title}</div>
            <ul style={{ listStyle: "none", padding: 0, margin: 0, display: "flex", flexDirection: "column", gap: 8 }}>
              {c.links.map((l) => (
                <li key={l.label}><a href={l.href} style={{ fontSize: 13, color: "var(--ink-3)", textDecoration: "none", cursor: "pointer" }}>{l.label}</a></li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="wrap" style={{ marginTop: 56, paddingTop: 24, borderTop: "1px solid var(--line)", display: "flex", justifyContent: "space-between", fontSize: 12, color: "var(--ink-3)" }}>
        <span>© 2026 DiffGraph. — u-22 プログラミングコンテスト</span>
        <span className="mono">v0.4.2-beta</span>
      </div>
    </footer>
  );
}

// ──────────────────────────────────────────────────────────────────────
// App
// ──────────────────────────────────────────────────────────────────────

function App() {
  const [tweaks, setTweak] = useTweaks(TWEAK_DEFAULTS);
  return (
    <>
      <div id="top">
        <Nav authState={tweaks.authState} />
        <Hero authState={tweaks.authState} ctaTone={tweaks.ctaTone} />
        <ClusterSection />
        <HowItWorks />
        <Features />
        <CTAStrip authState={tweaks.authState} />
        <Footer />
      </div>

      <TweaksPanel title="Tweaks">
        <TweakSection title="ログイン状態">
          <TweakRadio
            label="状態"
            value={tweaks.authState}
            options={[
              { value: "guest", label: "未ログイン" },
              { value: "user",  label: "ログイン済み" },
            ]}
            onChange={(v) => setTweak("authState", v)}
          />
        </TweakSection>
        <TweakSection title="CTA">
          <TweakRadio
            label="プライマリボタンの色"
            value={tweaks.ctaTone}
            options={[
              { value: "ink",   label: "ink (#1c1917)" },
              { value: "brand", label: "brand teal" },
            ]}
            onChange={(v) => setTweak("ctaTone", v)}
          />
        </TweakSection>
      </TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);
