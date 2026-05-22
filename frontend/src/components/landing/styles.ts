import type { CSSProperties } from 'react'

// Shared inline-style atoms for the landing page. Ported from the Claude Design
// handoff (`design/project/landing.jsx`); colors resolve against the scoped
// `.landing-root` palette in `landing.css`.

export const btn: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  justifyContent: 'center',
  gap: 8,
  height: 44,
  padding: '0 18px',
  borderRadius: 10,
  fontSize: 14,
  fontWeight: 500,
  fontFamily: 'inherit',
  letterSpacing: '-0.005em',
  cursor: 'pointer',
  border: '1px solid transparent',
  whiteSpace: 'nowrap',
  transition: 'transform 80ms ease, box-shadow 80ms ease',
  textDecoration: 'none',
}

export const btnInk: CSSProperties = { background: 'var(--ink)', color: 'white' }
export const btnBrand: CSSProperties = { background: 'var(--brand)', color: 'white' }
export const btnGhost: CSSProperties = {
  background: 'transparent',
  color: 'var(--ink)',
  border: '1px solid var(--line-strong)',
}

export const navLinkStyle: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: 5,
  color: 'var(--ink-2)',
  textDecoration: 'none',
  cursor: 'pointer',
}

export const pillStyle: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: 8,
  padding: '6px 12px',
  background: 'var(--brand-soft)',
  color: 'var(--brand-ink)',
  borderRadius: 99,
  fontSize: 12,
  fontWeight: 500,
  border: '1px solid #99f6e4',
}
