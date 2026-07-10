// DiffGraph. icon explorations
// Each icon is a 192x192 SVG. The visible "glyph" lives in ~120x120 centered.

const TEAL = '#0f766e';
const TEAL_BG = '#ccfbf1';
const INK = '#1c1917';
const INK3 = '#78716c';
const ADD = '#16a34a';
const DEL = '#dc2626';
const MOD = '#d97706';
const REF = '#7c3aed';

// ──────────────────────────────────────────────────────────────────────
// A · Node + Edge
// ──────────────────────────────────────────────────────────────────────

// A1 — Three-node triangle, one anchor highlighted (diff origin)
function A1(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke="${fg}" stroke-width="10" stroke-linecap="round">
        <line x1="56" y1="64" x2="136" y2="64"/>
        <line x1="56" y1="64" x2="96" y2="136"/>
        <line x1="136" y1="64" x2="96" y2="136"/>
      </g>
      <circle cx="56" cy="64" r="14" fill="${fg}"/>
      <circle cx="136" cy="64" r="14" fill="${fg}"/>
      <circle cx="96" cy="136" r="18" fill="${accent}"/>
    </svg>`;
}

// A2 — 2x2 module grid with edges (mini DAG)
function A2(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke="${fg}" stroke-width="9" stroke-linecap="round">
        <line x1="60" y1="60" x2="132" y2="60"/>
        <line x1="60" y1="60" x2="60" y2="132"/>
        <line x1="132" y1="60" x2="132" y2="132"/>
        <line x1="60" y1="132" x2="132" y2="132"/>
        <line x1="60" y1="60" x2="132" y2="132"/>
      </g>
      <rect x="46" y="46" width="28" height="28" rx="6" fill="${fg}"/>
      <rect x="118" y="46" width="28" height="28" rx="6" fill="${fg}"/>
      <rect x="46" y="118" width="28" height="28" rx="6" fill="${fg}"/>
      <rect x="118" y="118" width="28" height="28" rx="6" fill="${accent}"/>
    </svg>`;
}

// A3 — Hub-and-spoke: central node radiating, period dot
function A3(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke="${fg}" stroke-width="9" stroke-linecap="round">
        <line x1="96" y1="96" x2="50" y2="58"/>
        <line x1="96" y1="96" x2="142" y2="58"/>
        <line x1="96" y1="96" x2="50" y2="134"/>
      </g>
      <circle cx="50" cy="58" r="12" fill="${fg}"/>
      <circle cx="142" cy="58" r="12" fill="${fg}"/>
      <circle cx="50" cy="134" r="12" fill="${fg}"/>
      <circle cx="96" cy="96" r="22" fill="${accent}"/>
      <circle cx="142" cy="134" r="10" fill="${accent}"/>
    </svg>`;
}

// A4 — Two clusters bridged (the "diff between two PRs" metaphor)
function A4(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke="${fg}" stroke-width="8" stroke-linecap="round">
        <line x1="52" y1="60" x2="52" y2="132"/>
        <line x1="52" y1="60" x2="84" y2="96"/>
        <line x1="52" y1="132" x2="84" y2="96"/>
        <line x1="84" y1="96" x2="108" y2="96"/>
        <line x1="108" y1="96" x2="140" y2="60"/>
        <line x1="108" y1="96" x2="140" y2="132"/>
        <line x1="140" y1="60" x2="140" y2="132"/>
      </g>
      <circle cx="52" cy="60" r="11" fill="${fg}"/>
      <circle cx="52" cy="132" r="11" fill="${fg}"/>
      <circle cx="84" cy="96" r="11" fill="${fg}"/>
      <circle cx="108" cy="96" r="11" fill="${accent}"/>
      <circle cx="140" cy="60" r="11" fill="${accent}"/>
      <circle cx="140" cy="132" r="11" fill="${accent}"/>
    </svg>`;
}

// ──────────────────────────────────────────────────────────────────────
// B · Diff Mark
// ──────────────────────────────────────────────────────────────────────

// B1 — Plus / minus pair as primary glyph
function B1(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke="${fg}" stroke-width="14" stroke-linecap="round">
        <line x1="60" y1="68" x2="60" y2="124"/>
        <line x1="32" y1="96" x2="88" y2="96"/>
      </g>
      <line x1="104" y1="96" x2="160" y2="96" stroke="${accent}" stroke-width="14" stroke-linecap="round"/>
    </svg>`;
}

// B2 — Code diff lines stacked (the "hunk")
function B2(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <g stroke-linecap="round" stroke-width="12">
        <line x1="56" y1="58" x2="124" y2="58" stroke="${fg}" opacity="0.45"/>
        <line x1="56" y1="86" x2="148" y2="86" stroke="${accent}"/>
        <line x1="56" y1="114" x2="108" y2="114" stroke="${fg}" opacity="0.45"/>
        <line x1="56" y1="142" x2="136" y2="142" stroke="${fg}"/>
      </g>
    </svg>`;
}

// B3 — Branch split → merge (refactor metaphor)
function B3(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <path d="M60 50 L60 80 Q60 96 76 96 L116 96 Q132 96 132 112 L132 142"
        stroke="${fg}" stroke-width="10" fill="none" stroke-linecap="round"/>
      <path d="M132 50 L132 80 Q132 96 116 96"
        stroke="${accent}" stroke-width="10" fill="none" stroke-linecap="round"/>
      <circle cx="60" cy="46" r="11" fill="${fg}"/>
      <circle cx="132" cy="46" r="11" fill="${accent}"/>
      <circle cx="132" cy="146" r="11" fill="${fg}"/>
    </svg>`;
}

// B4 — Four-quadrant cluster squares (4-way classification)
function B4(fg, accent, bg) {
  // ignores fg/accent for cluster colors except in dark mode
  const useBrand = bg !== INK && bg !== TEAL;
  const add = useBrand ? ADD : '#86efac';
  const del = useBrand ? DEL : '#fca5a5';
  const mod = useBrand ? MOD : '#fcd34d';
  const ref = useBrand ? REF : '#c4b5fd';
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <rect x="48" y="48" width="44" height="44" rx="8" fill="${add}"/>
      <rect x="100" y="48" width="44" height="44" rx="8" fill="${del}"/>
      <rect x="48" y="100" width="44" height="44" rx="8" fill="${mod}"/>
      <rect x="100" y="100" width="44" height="44" rx="8" fill="${ref}"/>
    </svg>`;
}

// ──────────────────────────────────────────────────────────────────────
// C · Map / Cluster
// ──────────────────────────────────────────────────────────────────────

// C1 — Concentric impact rings (hop visualization)
function C1(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <circle cx="96" cy="96" r="56" stroke="${fg}" stroke-width="6" opacity="0.25" fill="none" stroke-dasharray="2 8" stroke-linecap="round"/>
      <circle cx="96" cy="96" r="38" stroke="${fg}" stroke-width="6" opacity="0.5" fill="none" stroke-dasharray="2 8" stroke-linecap="round"/>
      <circle cx="96" cy="96" r="20" fill="${accent}"/>
    </svg>`;
}

// C2 — "dm." monogram, geometric
function C2(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <!-- d: vertical bar + open circle stub -->
      <rect x="44" y="48" width="14" height="96" rx="3" fill="${fg}"/>
      <path d="M58 96 Q58 76 78 76 Q98 76 98 96 Q98 116 78 116 L58 116"
        stroke="${fg}" stroke-width="14" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
      <!-- m: three bars -->
      <rect x="112" y="76" width="12" height="68" rx="3" fill="${fg}"/>
      <rect x="130" y="76" width="12" height="68" rx="3" fill="${fg}"/>
      <rect x="148" y="76" width="12" height="68" rx="3" fill="${fg}"/>
      <rect x="112" y="76" width="48" height="12" rx="3" fill="${fg}"/>
      <!-- period accent -->
      <circle cx="162" cy="140" r="8" fill="${accent}"/>
    </svg>`;
}

// C3 — Map territory: irregular polygon clusters
function C3(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <path d="M52 56 L92 48 L104 78 L82 102 L48 92 Z" fill="${fg}" opacity="0.85"/>
      <path d="M108 52 L142 64 L138 94 L112 86 Z" fill="${accent}"/>
      <path d="M56 110 L96 108 L100 142 L62 146 Z" fill="${fg}" opacity="0.5"/>
      <path d="M116 102 L150 110 L146 144 L120 144 Z" fill="${fg}" opacity="0.7"/>
    </svg>`;
}

// C4 — Lowercase "d" + period — pure mark
function C4(fg, accent, bg) {
  return `
    <svg class="icon" viewBox="0 0 192 192" fill="none" xmlns="http://www.w3.org/2000/svg">
      ${bg ? `<rect width="192" height="192" rx="40" fill="${bg}"/>` : ''}
      <!-- circle bowl of d -->
      <circle cx="84" cy="112" r="32" stroke="${fg}" stroke-width="14" fill="none"/>
      <!-- stem of d -->
      <rect x="109" y="44" width="14" height="100" rx="3" fill="${fg}"/>
      <!-- period -->
      <circle cx="146" cy="138" r="9" fill="${accent}"/>
    </svg>`;
}

// ──────────────────────────────────────────────────────────────────────
// Render
// ──────────────────────────────────────────────────────────────────────

function card(name, desc, builder) {
  return `
    <div class="card">
      <div class="stage">${builder(INK, TEAL, null)}</div>
      <div class="mini">
        <div class="swatch light">${builder(INK, TEAL, null).replace('class="icon"','width="28" height="28" viewBox="0 0 192 192"').replace(/<rect width="192" height="192".*?\/>/,'')}</div>
        <div class="swatch dark"><svg width="28" height="28" viewBox="0 0 192 192">${innerOf(builder('#fafaf9', '#5eead4', null))}</svg></div>
        <div class="swatch teal"><svg width="28" height="28" viewBox="0 0 192 192">${innerOf(builder('white', '#ccfbf1', null))}</svg></div>
      </div>
      <div class="meta">
        <div class="name">${name}</div>
        <div class="desc">${desc}</div>
      </div>
    </div>`;
}

// extract inner of <svg>...</svg>
function innerOf(svgString) {
  const m = svgString.match(/<svg[^>]*>([\s\S]*)<\/svg>/);
  return m ? m[1] : '';
}

const A = [
  ['A1 · Triangle node', '3 モジュール / 1 起点強調。最小単位のグラフ。', A1],
  ['A2 · 2×2 DAG', 'レイヤ構造を象徴。最も「依存グラフ」らしい。', A2],
  ['A3 · Hub & spoke', '影響範囲（hop）の中心ノードを表現。', A3],
  ['A4 · Bridged clusters', 'PR 間の橋渡し。diff の意味に最も近い。', A4],
];
const B = [
  ['B1 · Plus / minus', '差分そのもの。最もミニマル。teal は accent 側。', B1],
  ['B2 · Hunk lines', 'diff の hunk。コード差分の即時想起。', B2],
  ['B3 · Branch & merge', 'リファクタ／合流の動き。', B3],
  ['B4 · 4-way cluster', '4分類（add/del/mod/refactor）の色そのものを記号化。', B4],
];
const C = [
  ['C1 · Impact rings', 'hop 数の波紋。影響範囲ビューと直結。', C1],
  ['C2 · dm. monogram', '頭文字 + ピリオド。ロゴから派生。', C2],
  ['C3 · Territory map', '「map」の比喩。クラスタを大陸として描く。', C3],
  ['C4 · d. mark', 'ブランド名の核「d」+「.」。最もタイポ寄り。', C4],
];

function render(rowId, items) {
  document.getElementById(rowId).innerHTML = items.map(([n, d, b]) => card(n, d, b)).join('');
}

render('row-a', A);
render('row-b', B);
render('row-c', C);
