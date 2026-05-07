type Props = { size?: number }

export default function Brand({ size = 16 }: Props) {
  return (
    <div className="flex items-center gap-1.5">
      <svg width={size + 2} height={size + 2} viewBox="0 0 18 18" fill="none" aria-hidden="true">
        <rect x="1" y="1" width="16" height="16" rx="4" fill="#0f766e" />
        <path d="M5 9h8M9 5v8" stroke="white" strokeWidth="2" strokeLinecap="round" />
      </svg>
      <span
        style={{
          fontFamily: "'Inter Tight', sans-serif",
          fontSize: size,
          color: '#0f766e',
          fontWeight: 600,
          letterSpacing: '-0.01em',
        }}
      >
        diffmap.
      </span>
    </div>
  )
}
