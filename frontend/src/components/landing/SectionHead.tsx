type Props = { eyebrow: string; title: string; lead: string }

export default function SectionHead({ eyebrow, title, lead }: Props) {
  return (
    <div style={{ maxWidth: 720 }}>
      <div
        className="mono"
        style={{
          fontSize: 11,
          color: 'var(--brand)',
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
          marginBottom: 14,
        }}
      >
        {eyebrow}
      </div>
      <h2
        style={{
          fontSize: 'clamp(28px, 4.4vw, 38px)',
          fontWeight: 600,
          letterSpacing: '-0.03em',
          lineHeight: 1.1,
          margin: '0 0 14px',
        }}
      >
        {title}
      </h2>
      <p
        style={{ fontSize: 15.5, color: 'var(--ink-2)', lineHeight: 1.6, margin: 0, maxWidth: 580 }}
      >
        {lead}
      </p>
    </div>
  )
}
