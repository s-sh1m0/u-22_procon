import AppMark from './AppMark'

type Props = { size?: number }

export default function Brand({ size = 16 }: Props) {
  return (
    <div className="flex items-center gap-2.5">
      <AppMark size={size + 12} />
      <span className="font-semibold tracking-[-0.025em] text-stone-900" style={{ fontSize: size }}>
        DiffGraph<span className="text-teal-700">.</span>
      </span>
    </div>
  )
}
