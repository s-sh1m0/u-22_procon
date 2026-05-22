type Props = { eyebrow: string; title: string; lead: string }

export default function SectionHead({ eyebrow, title, lead }: Props) {
  return (
    <div className="max-w-3xl">
      <div className="mb-3.5 font-mono text-[11px] uppercase tracking-[0.08em] text-teal-700">
        {eyebrow}
      </div>
      <h2 className="m-0 mb-3.5 text-[clamp(28px,4.4vw,38px)] font-semibold leading-[1.1] tracking-[-0.03em] text-balance">
        {title}
      </h2>
      <p className="m-0 max-w-xl text-[15.5px] leading-relaxed text-stone-700">{lead}</p>
    </div>
  )
}
