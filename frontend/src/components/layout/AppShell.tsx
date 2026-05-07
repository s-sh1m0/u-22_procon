type Props = {
  topBar: React.ReactNode
  children: React.ReactNode
  rightPanel?: React.ReactNode
}

export default function AppShell({ topBar, children, rightPanel }: Props) {
  return (
    <div className="grid h-svh grid-rows-[auto_1fr] overflow-hidden">
      {topBar}
      <div className="flex min-h-0">
        <div className="relative min-h-0 flex-1">{children}</div>
        {rightPanel}
      </div>
    </div>
  )
}
