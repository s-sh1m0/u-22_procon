const PALETTE = [
  { hex: '#0d9488', soft: '#ccfbf1', ink: '#134e4a' }, // teal
  { hex: '#4f46e5', soft: '#e0e7ff', ink: '#1e1b4b' }, // indigo
  { hex: '#d97706', soft: '#fef3c7', ink: '#451a03' }, // amber
  { hex: '#e11d48', soft: '#ffe4e6', ink: '#4c0519' }, // rose
  { hex: '#059669', soft: '#d1fae5', ink: '#064e3b' }, // emerald
  { hex: '#0284c7', soft: '#e0f2fe', ink: '#0c4a6e' }, // sky
  { hex: '#a21caf', soft: '#fae8ff', ink: '#581c87' }, // fuchsia
  { hex: '#65a30d', soft: '#ecfccb', ink: '#1a2e05' }, // lime
] as const

export type ClusterColor = { hex: string; soft: string; ink: string }

export function getClusterColor(clusterId: number): ClusterColor {
  return PALETTE[((clusterId % PALETTE.length) + PALETTE.length) % PALETTE.length]
}
