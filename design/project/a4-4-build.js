// DiffGraph. — A4·4 production set

const INK = '#1c1917';
const CREAM = '#fafaf9';
const TEAL = '#0f766e';
const TEAL_DEEP = '#0d5e57';
const TEAL_TINT = '#ccfbf1';
const TEAL_LIGHT = '#5eead4';
const TEAL_BG = '#f0fdfa';

// ──────────────────────────────────────────────────────────────────────
// Core refined geometry
// ──────────────────────────────────────────────────────────────────────
//
//   ●─────●─.
//   │ ╲ ╱ │  ╲
//   ●     │   ●●  ← accent
//   │ ╱ ╲ │  ╱
//   ●─────●─'
//
//   cluster: 5 ink nodes (left col 3 + right col 2)
//   accent : 1 large node, bridged from right-col via V-edges

function mark({ fg = INK, ac = TEAL, acFill = null, showSafe = false, showGrid = false, bridgeBoost = false } = {}) {
  acFill = acFill || ac;
  return `
    ${showGrid ? `
      <g stroke="${ac}" stroke-width="0.4" opacity="0.18">
        <line x1="0" y1="96" x2="192" y2="96"/>
        <line x1="96" y1="0" x2="96" y2="192"/>
      </g>` : ''}
    ${showSafe ? `
      <rect x="12" y="12" width="168" height="168" rx="38" fill="none" stroke="${ac}" stroke-width="0.6" stroke-dasharray="3 4" opacity="0.45"/>
      <circle cx="52" cy="52" r="11" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
      <circle cx="52" cy="96" r="11" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
      <circle cx="52" cy="140" r="11" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
      <circle cx="90" cy="76" r="11" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
      <circle cx="90" cy="116" r="11" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
      <circle cx="146" cy="96" r="16" fill="none" stroke="${ac}" stroke-width="0.4" opacity="0.3"/>
    ` : ''}
    <g stroke="${fg}" stroke-width="8" stroke-linecap="round" fill="none">
      <line x1="52" y1="52" x2="52" y2="140"/>
      <line x1="52" y1="52" x2="90" y2="76"/>
      <line x1="52" y1="140" x2="90" y2="116"/>
      <line x1="90" y1="76" x2="90" y2="116"/>
      <line x1="52" y1="96" x2="90" y2="76"/>
      <line x1="52" y1="96" x2="90" y2="116"/>
    </g>
    <g stroke="${ac}" stroke-width="${bridgeBoost ? 12 : 10}" stroke-linecap="round" fill="none">
      <line x1="90" y1="76" x2="146" y2="96"/>
      <line x1="90" y1="116" x2="146" y2="96"/>
    </g>
    <circle cx="52" cy="52" r="11" fill="${fg}"/>
    <circle cx="52" cy="96" r="11" fill="${fg}"/>
    <circle cx="52" cy="140" r="11" fill="${fg}"/>
    <circle cx="90" cy="76" r="11" fill="${fg}"/>
    <circle cx="90" cy="116" r="11" fill="${fg}"/>
    <circle cx="146" cy="96" r="16" fill="${acFill}"/>`;
}

const svg = (inner, attrs = '') =>
  `<svg viewBox="0 0 192 192" xmlns="http://www.w3.org/2000/svg" ${attrs}>${inner}</svg>`;

// ──────────────────────────────────────────────────────────────────────
// Hero
// ──────────────────────────────────────────────────────────────────────

document.getElementById('hero-svg').innerHTML = `
  <defs>
    <radialGradient id="heroBg" cx="0.3" cy="0.3" r="1">
      <stop offset="0%" stop-color="${TEAL_LIGHT}"/>
      <stop offset="100%" stop-color="${TEAL_DEEP}"/>
    </radialGradient>
  </defs>
  <rect width="192" height="192" fill="url(#heroBg)"/>
  ${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}
`;

document.getElementById('lockup-svg').innerHTML = `
  <rect width="192" height="192" fill="${TEAL}" rx="42"/>
  ${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}
`;

// ──────────────────────────────────────────────────────────────────────
// Construction view
// ──────────────────────────────────────────────────────────────────────

document.getElementById('construct-svg').innerHTML = `
  <rect width="192" height="192" fill="${TEAL_BG}" rx="42"/>
  ${mark({ fg: INK, ac: TEAL, showSafe: true, showGrid: true })}
`;

// ──────────────────────────────────────────────────────────────────────
// 02 · Color treatments
// ──────────────────────────────────────────────────────────────────────

const treatments = [
  {
    label: 'A · primary',
    name: 'Ink on cream',
    desc: 'デフォルト。ドキュメント / Web の本文中。',
    bg: 'white',
    cls: '',
    fg: INK, ac: TEAL, acFill: TEAL,
  },
  {
    label: 'B · dark',
    name: 'Cream on ink',
    desc: 'ダークモード UI / ナイト用 OS アイコン。',
    bg: INK,
    cls: 'dark',
    fg: CREAM, ac: TEAL_LIGHT, acFill: TEAL_LIGHT,
  },
  {
    label: 'C · brand',
    name: 'White on teal',
    desc: 'アプリアイコン本命。ブランドの「顔」。',
    bg: TEAL,
    cls: 'teal',
    fg: 'white', ac: TEAL_TINT, acFill: 'white',
  },
  {
    label: 'D · tint',
    name: 'Ink on tint',
    desc: 'マーケ素材・スライド・名刺。柔らかい印象。',
    bg: TEAL_BG,
    cls: 'tint',
    fg: INK, ac: TEAL, acFill: TEAL,
  },
];

document.getElementById('treatments').innerHTML = treatments.map(t => `
  <div>
    <div class="stage ${t.cls}">
      <span class="badge">${t.label}</span>
      ${svg(mark({ fg: t.fg, ac: t.ac, acFill: t.acFill }))}
    </div>
    <div class="meta">
      <div class="name">${t.name}</div>
      <div class="desc">${t.desc}</div>
    </div>
  </div>
`).join('');

// ──────────────────────────────────────────────────────────────────────
// 03 · Gradient study
// ──────────────────────────────────────────────────────────────────────

function gradientIcon(kind) {
  if (kind === 'flat') {
    return `<rect width="192" height="192" fill="${TEAL}" rx="42"/>${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}`;
  }
  if (kind === 'subtle') {
    return `
      <defs>
        <linearGradient id="bgSubtle" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#14897f"/>
          <stop offset="100%" stop-color="${TEAL_DEEP}"/>
        </linearGradient>
      </defs>
      <rect width="192" height="192" fill="url(#bgSubtle)" rx="42"/>
      ${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}
    `;
  }
  if (kind === 'duotone') {
    return `
      <defs>
        <linearGradient id="bgDuo" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="${INK}"/>
          <stop offset="100%" stop-color="${TEAL_DEEP}"/>
        </linearGradient>
        <radialGradient id="accentDuo" cx="0.35" cy="0.35" r="0.8">
          <stop offset="0%" stop-color="${TEAL_LIGHT}"/>
          <stop offset="100%" stop-color="${TEAL}"/>
        </radialGradient>
      </defs>
      <rect width="192" height="192" fill="url(#bgDuo)" rx="42"/>
      ${mark({ fg: 'white', ac: 'url(#accentDuo)', acFill: 'url(#accentDuo)' })}
    `;
  }
}

document.getElementById('gradients').innerHTML = [
  ['flat', 'Flat', '単色 teal。最も静か。Web 用・ファビコン推奨。'],
  ['subtle', 'Subtle', '上から下へ微妙な濃淡。深さが出る。アプリアイコン本命。'],
  ['duotone', 'Duotone', 'ink → teal の対角。広告 / OG / プロモ用。']
].map(([k, name, desc]) => `
  <div>
    <div class="stage">
      <span class="badge mono">${k}</span>
      ${svg(gradientIcon(k))}
    </div>
    <div class="meta">
      <div class="name">${name}</div>
      <div class="desc">${desc}</div>
    </div>
  </div>
`).join('');

// ──────────────────────────────────────────────────────────────────────
// 04 · Corner radius
// ──────────────────────────────────────────────────────────────────────

const radii = [
  { rx: 0,  label: 'square (0%)', use: 'Print / favicon ICO' },
  { rx: 24, label: 'soft (12%)',  use: 'Android adaptive (square mask)' },
  { rx: 42, label: 'iOS (22%)',   use: 'macOS / iOS / 標準アプリ' },
  { rx: 96, label: 'circle (50%)', use: 'Avatar / Slack icon' },
];

document.getElementById('radii').innerHTML = radii.map(r => `
  <div class="item">
    <div class="icon-wrap">
      <svg class="app" viewBox="0 0 192 192" style="border-radius: ${(r.rx / 192) * 100}%;">
        <defs>
          <linearGradient id="g-${r.rx}" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#14897f"/>
            <stop offset="100%" stop-color="${TEAL_DEEP}"/>
          </linearGradient>
        </defs>
        <rect width="192" height="192" fill="url(#g-${r.rx})" rx="${r.rx}"/>
        ${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}
      </svg>
    </div>
    <div class="meta">
      <div class="label">${r.rx === 0 ? '0' : Math.round(r.rx / 192 * 100) + '%'}</div>
      <div class="name">${r.label}</div>
      <div class="desc">${r.use}</div>
    </div>
  </div>
`).join('');

// ──────────────────────────────────────────────────────────────────────
// 05 · Favicons
// ──────────────────────────────────────────────────────────────────────

const sizes = [16, 24, 32, 48, 64, 128];
document.getElementById('favicons').innerHTML = sizes.map(s => {
  // Below 32px, simplify: drop the cross edges to keep readable
  const inner = s < 32
    ? `
      <rect width="192" height="192" fill="${TEAL}" rx="42"/>
      <g stroke="white" stroke-width="14" stroke-linecap="round" fill="none">
        <line x1="52" y1="52" x2="52" y2="140"/>
        <line x1="52" y1="52" x2="90" y2="76"/>
        <line x1="52" y1="140" x2="90" y2="116"/>
        <line x1="90" y1="76" x2="90" y2="116"/>
      </g>
      <g stroke="${TEAL_TINT}" stroke-width="16" stroke-linecap="round" fill="none">
        <line x1="90" y1="76" x2="146" y2="96"/>
        <line x1="90" y1="116" x2="146" y2="96"/>
      </g>
      <circle cx="52" cy="52" r="16" fill="white"/>
      <circle cx="52" cy="96" r="16" fill="white"/>
      <circle cx="52" cy="140" r="16" fill="white"/>
      <circle cx="90" cy="76" r="16" fill="white"/>
      <circle cx="90" cy="116" r="16" fill="white"/>
      <circle cx="146" cy="96" r="22" fill="white"/>
    `
    : `<rect width="192" height="192" fill="${TEAL}" rx="42"/>${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}`;
  return `
    <div class="fav-item">
      <svg width="${s}" height="${s}" viewBox="0 0 192 192" style="border-radius: ${(42 / 192) * s}px;">${inner}</svg>
      <span class="px">${s}×${s}</span>
    </div>`;
}).join('');

// ──────────────────────────────────────────────────────────────────────
// 06 · OG mark
// ──────────────────────────────────────────────────────────────────────

document.getElementById('og-mark').innerHTML = `
  <defs>
    <linearGradient id="ogBg" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color="#14897f"/>
      <stop offset="100%" stop-color="${TEAL_DEEP}"/>
    </linearGradient>
  </defs>
  <rect width="192" height="192" fill="url(#ogBg)"/>
  ${mark({ fg: 'white', ac: TEAL_TINT, acFill: 'white' })}
`;
