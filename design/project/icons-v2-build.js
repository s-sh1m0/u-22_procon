// DiffGraph. — icon deep-dive (A1 / A4 / B1 / B3)
// Each function returns SVG inner content (no <svg> wrapper, no bg rect).
// The card wrapper supplies the background.

const INK = '#1c1917';
const INK3 = '#78716c';
const CREAM = '#fafaf9';
const TEAL = '#0f766e';
const TEAL_TINT = '#ccfbf1';
const TEAL_LIGHT = '#5eead4';

// VIEWBOX is 192x192 throughout.

// ──────────────────────────────────────────────────────────────────────
// A1 · Triangle node — 6 variations
// ──────────────────────────────────────────────────────────────────────

const A1 = {};

// a1.1 — baseline (current)
A1.v1 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="10" stroke-linecap="round" fill="none">
    <line x1="56" y1="64" x2="136" y2="64"/>
    <line x1="56" y1="64" x2="96" y2="136"/>
    <line x1="136" y1="64" x2="96" y2="136"/>
  </g>
  <circle cx="56" cy="64" r="14" fill="${fg}"/>
  <circle cx="136" cy="64" r="14" fill="${fg}"/>
  <circle cx="96" cy="136" r="18" fill="${ac}"/>`;

// a1.2 — chunky nodes, light edges
A1.v2 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="6" stroke-linecap="round" opacity="0.7" fill="none">
    <line x1="56" y1="64" x2="136" y2="64"/>
    <line x1="56" y1="64" x2="96" y2="136"/>
    <line x1="136" y1="64" x2="96" y2="136"/>
  </g>
  <circle cx="56" cy="64" r="20" fill="${fg}"/>
  <circle cx="136" cy="64" r="20" fill="${fg}"/>
  <circle cx="96" cy="136" r="22" fill="${ac}"/>`;

// a1.3 — open nodes (rings), accent is the only filled
A1.v3 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none">
    <line x1="58" y1="62" x2="134" y2="62"/>
    <line x1="58" y1="62" x2="96" y2="138"/>
    <line x1="134" y1="62" x2="96" y2="138"/>
  </g>
  <circle cx="58" cy="62" r="14" fill="white" stroke="${fg}" stroke-width="8"/>
  <circle cx="134" cy="62" r="14" fill="white" stroke="${fg}" stroke-width="8"/>
  <circle cx="96" cy="138" r="18" fill="${ac}"/>`;

// a1.4 — all monochrome accent (single color triangle)
A1.v4 = (fg, ac) => `
  <g stroke="${ac}" stroke-width="12" stroke-linecap="round" stroke-linejoin="round" fill="none">
    <path d="M58 60 L134 60 L96 138 Z"/>
  </g>
  <circle cx="58" cy="60" r="14" fill="${ac}"/>
  <circle cx="134" cy="60" r="14" fill="${ac}"/>
  <circle cx="96" cy="138" r="14" fill="${ac}"/>`;

// a1.5 — inverted triangle (apex up), accent at top
A1.v5 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="10" stroke-linecap="round" fill="none">
    <line x1="56" y1="132" x2="136" y2="132"/>
    <line x1="56" y1="132" x2="96" y2="56"/>
    <line x1="136" y1="132" x2="96" y2="56"/>
  </g>
  <circle cx="56" cy="132" r="14" fill="${fg}"/>
  <circle cx="136" cy="132" r="14" fill="${fg}"/>
  <circle cx="96" cy="56" r="18" fill="${ac}"/>`;

// a1.6 — extended graph: triangle + outgoing edge ("change propagates")
A1.v6 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="9" stroke-linecap="round" fill="none">
    <line x1="48" y1="76" x2="116" y2="76"/>
    <line x1="48" y1="76" x2="82" y2="140"/>
    <line x1="116" y1="76" x2="82" y2="140"/>
    <line x1="116" y1="76" x2="152" y2="56" stroke="${ac}"/>
  </g>
  <circle cx="48" cy="76" r="12" fill="${fg}"/>
  <circle cx="82" cy="140" r="12" fill="${fg}"/>
  <circle cx="116" cy="76" r="14" fill="${ac}"/>
  <circle cx="152" cy="56" r="10" fill="${ac}"/>`;

// ──────────────────────────────────────────────────────────────────────
// A4 · Bridged clusters — 6 variations
// ──────────────────────────────────────────────────────────────────────

const A4 = {};

// a4.1 — baseline: 3-node cluster ↔ 3-node cluster
A4.v1 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none">
    <line x1="52" y1="60" x2="52" y2="132"/>
    <line x1="52" y1="60" x2="84" y2="96"/>
    <line x1="52" y1="132" x2="84" y2="96"/>
    <line x1="84" y1="96" x2="108" y2="96"/>
    <line x1="108" y1="96" x2="140" y2="60" stroke="${ac}"/>
    <line x1="108" y1="96" x2="140" y2="132" stroke="${ac}"/>
    <line x1="140" y1="60" x2="140" y2="132" stroke="${ac}"/>
  </g>
  <circle cx="52" cy="60" r="11" fill="${fg}"/>
  <circle cx="52" cy="132" r="11" fill="${fg}"/>
  <circle cx="84" cy="96" r="11" fill="${fg}"/>
  <circle cx="108" cy="96" r="11" fill="${ac}"/>
  <circle cx="140" cy="60" r="11" fill="${ac}"/>
  <circle cx="140" cy="132" r="11" fill="${ac}"/>`;

// a4.2 — 2+2 minimal, single bridge
A4.v2 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="9" stroke-linecap="round" fill="none">
    <line x1="56" y1="60" x2="56" y2="132"/>
    <line x1="56" y1="60" x2="86" y2="96"/>
    <line x1="56" y1="132" x2="86" y2="96"/>
    <line x1="86" y1="96" x2="106" y2="96"/>
    <line x1="106" y1="96" x2="136" y2="60" stroke="${ac}"/>
    <line x1="106" y1="96" x2="136" y2="132" stroke="${ac}"/>
  </g>
  <circle cx="56" cy="60" r="12" fill="${fg}"/>
  <circle cx="56" cy="132" r="12" fill="${fg}"/>
  <circle cx="86" cy="96" r="12" fill="${fg}"/>
  <circle cx="106" cy="96" r="12" fill="${ac}"/>
  <circle cx="136" cy="60" r="12" fill="${ac}"/>
  <circle cx="136" cy="132" r="12" fill="${ac}"/>`;

// a4.3 — central pivot node (bridge as single point)
A4.v3 = (fg, ac) => `
  <g stroke-width="9" stroke-linecap="round" fill="none">
    <line x1="54" y1="58" x2="96" y2="96" stroke="${fg}"/>
    <line x1="54" y1="134" x2="96" y2="96" stroke="${fg}"/>
    <line x1="138" y1="58" x2="96" y2="96" stroke="${ac}"/>
    <line x1="138" y1="134" x2="96" y2="96" stroke="${ac}"/>
  </g>
  <circle cx="54" cy="58" r="12" fill="${fg}"/>
  <circle cx="54" cy="134" r="12" fill="${fg}"/>
  <circle cx="138" cy="58" r="12" fill="${ac}"/>
  <circle cx="138" cy="134" r="12" fill="${ac}"/>
  <circle cx="96" cy="96" r="16" fill="${fg}"/>`;

// a4.4 — asymmetric: big cluster, single change node
A4.v4 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none">
    <line x1="50" y1="50" x2="90" y2="80"/>
    <line x1="50" y1="50" x2="50" y2="142"/>
    <line x1="50" y1="142" x2="90" y2="112"/>
    <line x1="90" y1="80" x2="90" y2="112"/>
    <line x1="50" y1="96" x2="90" y2="96"/>
    <line x1="90" y1="96" x2="142" y2="96" stroke="${ac}"/>
  </g>
  <circle cx="50" cy="50" r="11" fill="${fg}"/>
  <circle cx="50" cy="96" r="11" fill="${fg}"/>
  <circle cx="50" cy="142" r="11" fill="${fg}"/>
  <circle cx="90" cy="80" r="11" fill="${fg}"/>
  <circle cx="90" cy="112" r="11" fill="${fg}"/>
  <circle cx="142" cy="96" r="16" fill="${ac}"/>`;

// a4.5 — diagonal flow (top-left cluster → bottom-right cluster)
A4.v5 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none">
    <line x1="48" y1="52" x2="80" y2="52"/>
    <line x1="48" y1="52" x2="48" y2="84"/>
    <line x1="80" y1="52" x2="48" y2="84"/>
    <line x1="80" y1="52" x2="112" y2="84"/>
    <line x1="48" y1="84" x2="112" y2="140" stroke="${ac}" stroke-dasharray="2 10"/>
    <line x1="112" y1="84" x2="144" y2="140" stroke="${ac}"/>
    <line x1="112" y1="84" x2="144" y2="108" stroke="${ac}"/>
    <line x1="144" y1="108" x2="144" y2="140" stroke="${ac}"/>
    <line x1="112" y1="140" x2="144" y2="140" stroke="${ac}"/>
  </g>
  <circle cx="48" cy="52" r="11" fill="${fg}"/>
  <circle cx="80" cy="52" r="11" fill="${fg}"/>
  <circle cx="48" cy="84" r="11" fill="${fg}"/>
  <circle cx="112" cy="84" r="11" fill="${ac}"/>
  <circle cx="144" cy="108" r="11" fill="${ac}"/>
  <circle cx="112" cy="140" r="11" fill="${ac}"/>
  <circle cx="144" cy="140" r="11" fill="${ac}"/>`;

// a4.6 — bridge highlighted: muted clusters, bright bridge
A4.v6 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none" opacity="0.55">
    <line x1="52" y1="60" x2="52" y2="132"/>
    <line x1="52" y1="60" x2="80" y2="96"/>
    <line x1="52" y1="132" x2="80" y2="96"/>
    <line x1="112" y1="96" x2="140" y2="60"/>
    <line x1="112" y1="96" x2="140" y2="132"/>
    <line x1="140" y1="60" x2="140" y2="132"/>
  </g>
  <g opacity="0.55">
    <circle cx="52" cy="60" r="11" fill="${fg}"/>
    <circle cx="52" cy="132" r="11" fill="${fg}"/>
    <circle cx="80" cy="96" r="11" fill="${fg}"/>
    <circle cx="112" cy="96" r="11" fill="${fg}"/>
    <circle cx="140" cy="60" r="11" fill="${fg}"/>
    <circle cx="140" cy="132" r="11" fill="${fg}"/>
  </g>
  <line x1="80" y1="96" x2="112" y2="96" stroke="${ac}" stroke-width="12" stroke-linecap="round"/>`;

// ──────────────────────────────────────────────────────────────────────
// B1 · Plus / Minus — 6 variations
// ──────────────────────────────────────────────────────────────────────

const B1 = {};

// b1.1 — baseline horizontal, single row
B1.v1 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="14" stroke-linecap="round">
    <line x1="60" y1="68" x2="60" y2="124"/>
    <line x1="32" y1="96" x2="88" y2="96"/>
  </g>
  <line x1="104" y1="96" x2="160" y2="96" stroke="${ac}" stroke-width="14" stroke-linecap="round"/>`;

// b1.2 — stacked vertical (+ on top, − below)
B1.v2 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="14" stroke-linecap="round">
    <line x1="96" y1="38" x2="96" y2="86"/>
    <line x1="72" y1="62" x2="120" y2="62"/>
  </g>
  <line x1="72" y1="130" x2="120" y2="130" stroke="${ac}" stroke-width="14" stroke-linecap="round"/>`;

// b1.3 — pill badges (filled circles)
B1.v3 = (fg, ac) => `
  <circle cx="68" cy="96" r="32" fill="${fg}"/>
  <g stroke="white" stroke-width="8" stroke-linecap="round">
    <line x1="54" y1="96" x2="82" y2="96"/>
    <line x1="68" y1="82" x2="68" y2="110"/>
  </g>
  <circle cx="132" cy="96" r="32" fill="${ac}"/>
  <line x1="118" y1="96" x2="146" y2="96" stroke="white" stroke-width="8" stroke-linecap="round"/>`;

// b1.4 — bold split blocks (two squares meet)
B1.v4 = (fg, ac) => `
  <rect x="40" y="56" width="58" height="80" rx="10" fill="${fg}"/>
  <rect x="94" y="56" width="58" height="80" rx="10" fill="${ac}"/>
  <g stroke="white" stroke-width="8" stroke-linecap="round">
    <line x1="56" y1="96" x2="82" y2="96"/>
    <line x1="69" y1="83" x2="69" y2="109"/>
    <line x1="110" y1="96" x2="136" y2="96"/>
  </g>`;

// b1.5 — large + with minus as accent stem (combined glyph)
B1.v5 = (fg, ac) => `
  <line x1="96" y1="44" x2="96" y2="148" stroke="${fg}" stroke-width="16" stroke-linecap="round"/>
  <line x1="44" y1="96" x2="96" y2="96" stroke="${fg}" stroke-width="16" stroke-linecap="round"/>
  <line x1="100" y1="96" x2="148" y2="96" stroke="${ac}" stroke-width="16" stroke-linecap="round"/>`;

// b1.6 — diff with period: + − .
B1.v6 = (fg, ac) => `
  <g stroke="${fg}" stroke-width="12" stroke-linecap="round">
    <line x1="44" y1="64" x2="44" y2="106"/>
    <line x1="23" y1="85" x2="65" y2="85"/>
  </g>
  <line x1="83" y1="85" x2="125" y2="85" stroke="${fg}" stroke-width="12" stroke-linecap="round"/>
  <circle cx="158" cy="138" r="14" fill="${ac}"/>`;

// ──────────────────────────────────────────────────────────────────────
// B3 · Branch & merge — 6 variations
// ──────────────────────────────────────────────────────────────────────

const B3 = {};

// b3.1 — baseline: two lines merge then continue
B3.v1 = (fg, ac) => `
  <path d="M60 50 L60 80 Q60 96 76 96 L116 96 Q132 96 132 112 L132 142"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <path d="M132 50 L132 80 Q132 96 116 96" stroke="${ac}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <circle cx="60" cy="46" r="11" fill="${fg}"/>
  <circle cx="132" cy="46" r="11" fill="${ac}"/>
  <circle cx="132" cy="146" r="11" fill="${fg}"/>`;

// b3.2 — two branches merging (Y shape)
B3.v2 = (fg, ac) => `
  <path d="M56 44 L56 80 Q56 100 80 100 L112 100 Q136 100 136 120 L136 148"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <path d="M136 44 L136 80 Q136 100 112 100"
    stroke="${ac}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <circle cx="56" cy="40" r="12" fill="${fg}"/>
  <circle cx="136" cy="40" r="12" fill="${ac}"/>
  <circle cx="136" cy="152" r="12" fill="${fg}"/>`;

// b3.3 — split then re-merge (refactor loop)
B3.v3 = (fg, ac) => `
  <path d="M96 36 L96 56 Q96 76 76 76 L60 76 Q44 76 44 96 L44 112 Q44 132 60 132 L132 132 Q148 132 148 112 L148 96 Q148 76 132 76 L116 76 Q96 76 96 56"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <circle cx="96" cy="32" r="11" fill="${fg}"/>
  <circle cx="44" cy="104" r="11" fill="${fg}"/>
  <circle cx="148" cy="104" r="11" fill="${ac}"/>
  <circle cx="96" cy="132" r="14" fill="${ac}"/>`;

// b3.4 — three branches converging into one (fan-in)
B3.v4 = (fg, ac) => `
  <path d="M44 44 L44 72 Q44 96 68 96 L124 96 Q148 96 148 72 L148 44"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <path d="M96 44 L96 96" stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <line x1="96" y1="96" x2="96" y2="148" stroke="${ac}" stroke-width="10" stroke-linecap="round"/>
  <circle cx="44" cy="40" r="11" fill="${fg}"/>
  <circle cx="96" cy="40" r="11" fill="${fg}"/>
  <circle cx="148" cy="40" r="11" fill="${fg}"/>
  <circle cx="96" cy="152" r="13" fill="${ac}"/>`;

// b3.5 — diagonal merge (asymmetric, motion feel)
B3.v5 = (fg, ac) => `
  <path d="M48 44 L48 96 Q48 120 72 120 L144 120"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <path d="M144 44 L144 96 Q144 120 120 120"
    stroke="${ac}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <circle cx="48" cy="40" r="11" fill="${fg}"/>
  <circle cx="144" cy="40" r="11" fill="${ac}"/>
  <circle cx="144" cy="124" r="11" fill="${fg}"/>`;

// b3.6 — split into two outputs (fan-out, mirror of b3.2)
B3.v6 = (fg, ac) => `
  <path d="M96 44 L96 60 Q96 80 76 80 L60 80 Q44 80 44 100 L44 148"
    stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <path d="M96 60 Q96 80 116 80 L132 80 Q148 80 148 100 L148 148"
    stroke="${ac}" stroke-width="10" fill="none" stroke-linecap="round"/>
  <circle cx="96" cy="40" r="12" fill="${fg}"/>
  <circle cx="44" cy="152" r="12" fill="${fg}"/>
  <circle cx="148" cy="152" r="12" fill="${ac}"/>`;

// ──────────────────────────────────────────────────────────────────────
// Render
// ──────────────────────────────────────────────────────────────────────

const PALETTES = {
  light: { fg: INK,   ac: TEAL,       bgClass: '' },
  dark:  { fg: CREAM, ac: TEAL_LIGHT, bgClass: 'dark' },
  teal:  { fg: 'white', ac: TEAL_TINT, bgClass: 'teal' },
};

function svg(content, w = '100%') {
  return `<svg viewBox="0 0 192 192" width="${w}" height="${w}" xmlns="http://www.w3.org/2000/svg">${content}</svg>`;
}

function card(id, label, desc, builder) {
  const lightSvg = svg(builder(PALETTES.light.fg, PALETTES.light.ac));
  const darkSvg  = svg(builder(PALETTES.dark.fg,  PALETTES.dark.ac));
  const tealSvg  = svg(builder(PALETTES.teal.fg,  PALETTES.teal.ac));
  return `
    <div class="card">
      <div class="stage">${lightSvg}</div>
      <div class="ctx">
        <div class="mini light">${svg(builder(PALETTES.light.fg, PALETTES.light.ac))}</div>
        <div class="mini dark">${darkSvg}</div>
        <div class="mini teal">${tealSvg}</div>
      </div>
      <div class="meta">
        <div class="label">${id}</div>
        <div class="name">${label}</div>
        <div class="desc">${desc}</div>
      </div>
    </div>`;
}

const ROWS = {
  'row-a1': [
    ['A1·1', 'baseline', '元案。下向き三角・底点 accent。', A1.v1],
    ['A1·2', 'chunky', '太いノード、細いエッジ。重心が下がる。', A1.v2],
    ['A1·3', 'rings', '開いたリング。線画寄り、軽い印象。', A1.v3],
    ['A1·4', 'mono triangle', '単色三角形。エッジが図形化。', A1.v4],
    ['A1·5', 'apex up', '上向き三角。頂点 accent で記号性 ↑。', A1.v5],
    ['A1·6', 'with edge out', '外への伝播エッジ。「影響範囲」感。', A1.v6],
  ],
  'row-a4': [
    ['A4·1', 'baseline 3+3', '元案。', A4.v1],
    ['A4·2', 'minimal 3+3', '同形でブリッジ 1 本。シンプル。', A4.v2],
    ['A4·3', 'pivot bridge', '中央 1 点で接続。シンメトリ。', A4.v3],
    ['A4·4', 'big + 1 change', '元側が大きく、変更先 1 ノード。', A4.v4],
    ['A4·5', 'diagonal flow', '左上→右下、流れがある。', A4.v5],
    ['A4·6', 'bridge focus', 'クラスタは淡く、橋だけ強調。', A4.v6],
  ],
  'row-b1': [
    ['B1·1', 'baseline', '元案。横並び。', B1.v1],
    ['B1·2', 'stacked', '+ の下に − を置く縦並び。', B1.v2],
    ['B1·3', 'badges', '丸バッジ内に + / −。GitHub 風。', B1.v3],
    ['B1·4', 'split block', '矩形 2 枚で diff を象徴。', B1.v4],
    ['B1·5', 'fused', '+ と − が連結。ひと筆のサイン感。', B1.v5],
    ['B1·6', '+ − .', 'ピリオド付き。ブランド名「DiffGraph.」に呼応。', B1.v6],
  ],
  'row-b3': [
    ['B3·1', 'baseline', '元案。', B3.v1],
    ['B3·2', 'Y merge', 'Y 字。2 ブランチ → 1。', B3.v2],
    ['B3·3', 'split & merge', 'ループ。リファクタの「振る舞い保存」感。', B3.v3],
    ['B3·4', '3-fan-in', '3 ブランチ → 1。大規模 PR 向き。', B3.v4],
    ['B3·5', 'diagonal', '対角に合流。動きがある。', B3.v5],
    ['B3·6', 'fan-out', '1 → 2 に分岐。「影響が広がる」感。', B3.v6],
  ],
};

for (const [rowId, items] of Object.entries(ROWS)) {
  document.getElementById(rowId).innerHTML = items.map(it => card(...it)).join('');
}
