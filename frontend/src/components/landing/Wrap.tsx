import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

/** Centered max-width content container shared by every landing section. */
export default function Wrap({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('mx-auto max-w-[1200px] px-5 sm:px-9', className)}>{children}</div>
}
