type Props = {
  /** Rendered width/height in px. */
  size?: number
  /** Corner radius as a fraction of the 192 viewBox (0 = square, 0.22 = iOS). */
  rx?: number
  /** Background fill of the rounded square. */
  bg?: string
  /** Stroke / node color of the dependency-graph glyph. */
  fg?: string
  /** Fill of the large accent (convergence) node. */
  accent?: string
  /** Drop a soft teal shadow (used on the dark CTA strip). */
  shadow?: boolean
}

/**
 * AppMark renders the diffmap. product mark (design "A4·4"): a small
 * dependency cluster of nodes converging via V-edges into one accent node.
 * Geometry mirrors the Claude Design handoff (`design/project/landing.jsx`).
 */
export default function AppMark({
  size = 32,
  rx = 0.22,
  bg = 'var(--brand, #0f766e)',
  fg = 'white',
  accent = 'white',
  shadow = false,
}: Props) {
  const radius = rx * 192
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 192 192"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
      style={
        shadow
          ? {
              filter:
                'drop-shadow(0 8px 24px rgba(15,118,110,0.28)) drop-shadow(0 2px 4px rgba(28,25,23,0.08))',
            }
          : undefined
      }
    >
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
  )
}
