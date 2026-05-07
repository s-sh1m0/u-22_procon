import type { LayerKind } from '@/types/graph'

export function inferLayer(packagePath: string): LayerKind {
  if (
    packagePath.includes('internal/api') ||
    packagePath.includes('cmd/') ||
    packagePath.includes('frontend/')
  )
    return 'ui'
  if (packagePath.includes('internal/usecase') || packagePath.includes('internal/domain'))
    return 'domain'
  if (packagePath.includes('internal/infra')) return 'infra'
  if (packagePath.includes('internal/pkg') || packagePath.includes('pkg/')) return 'data'
  return 'other'
}
