import AppMark from './AppMark'

type Props = { size?: number; color?: string }

export default function Brand({ size = 16, color = 'var(--ink, #1c1917)' }: Props) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <AppMark size={size + 12} />
      <span
        style={{
          fontFamily: "'Inter Tight', sans-serif",
          fontSize: size,
          fontWeight: 600,
          letterSpacing: '-0.025em',
          color,
        }}
      >
        diffmap<span style={{ color: 'var(--brand, #0f766e)' }}>.</span>
      </span>
    </div>
  )
}
