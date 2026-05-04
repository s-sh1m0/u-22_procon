// variant-b.jsx — 案B: スイムレーン式 — 4クラスタを縦レーンに分けて視覚化
// 横軸: 4クラスタ (added / modified / refactor / removed)
// 縦軸: 4レイヤ (UI / Domain / Data / Infra)
// 同じセルに属するモジュールを並べる + クロスレーンエッジで依存関係を見せる

function VariantB() {
  const [selected, setSelected] = React.useState("m11");
  const [hover, setHover] = React.useState(null);

  const layout = React.useMemo(
    () => computeSwimlaneLayout(MODULES, EDGES, { nodeWidth: 142, nodeHeight: 44, layerGapY: 100, nodeGapY: 8, laneGapX: 22, headerH: 56 }),
    []
  );

  const highlightId = hover || selected;
  const connectedIds = React.useMemo(() => {
    if (!highlightId) return new Set();
    const s = new Set([highlightId]);
    for (const [a, b] of EDGES) {
      if (a === highlightId) s.add(b);
      if (b === highlightId) s.add(a);
    }
    return s;
  }, [highlightId]);

  const sel = MODULES.find((m) => m.id === selected);

  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, display: "flex", flexDirection: "column", overflow: "hidden" }}>
      {/* top bar */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <Brand size={15} />
          <div style={{ width: 1, height: 20, background: T.border }} />
          <div>
            <div style={{ fontSize: 13, fontWeight: 600, color: T.ink, letterSpacing: -0.2 }}>{PR_META.title}</div>
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: T.mono, marginTop: 1 }}>{PR_META.repo} #{PR_META.number} · スイムレーン表示</div>
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12, fontSize: 12 }}>
          <span style={{ color: T.added, fontFamily: T.mono, fontWeight: 500 }}>+{PR_META.additions}</span>
          <span style={{ color: T.removed, fontFamily: T.mono, fontWeight: 500 }}>−{PR_META.deletions}</span>
          <div style={{ width: 24, height: 24, borderRadius: 99, background: T.brand, color: "#fff", fontSize: 10, fontWeight: 600, display: "flex", alignItems: "center", justifyContent: "center" }}>{PR_META.authorAvatar}</div>
        </div>
      </div>

      {/* cluster summary header */}
      <div style={{ padding: "12px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff", display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 12 }}>
        {Object.entries(CLUSTERS).map(([k, c]) => {
          const stat = CLUSTER_STATS[k];
          return (
            <div key={k} style={{ background: c.bg, border: `1px solid ${c.border}`, borderRadius: 10, padding: "10px 14px", display: "flex", alignItems: "center", gap: 12 }}>
              <div style={{ width: 32, height: 32, borderRadius: 8, background: "#fff", color: c.color, fontSize: 16, fontWeight: 700, display: "flex", alignItems: "center", justifyContent: "center" }}>{c.icon}</div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 12, fontWeight: 600, color: c.color, letterSpacing: 0.3 }}>{c.label}</div>
                <div style={{ fontSize: 11, color: T.ink2, fontFamily: T.mono }}>{stat.count} mods · +{stat.additions} −{stat.deletions}</div>
              </div>
            </div>
          );
        })}
      </div>

      {/* main */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 360px", overflow: "hidden" }}>
        <div style={{ position: "relative", overflow: "auto", background: T.bg }}>
          <SwimlaneCanvas
            modules={MODULES}
            edges={EDGES}
            layout={layout}
            selected={selected}
            hover={hover}
            connectedIds={connectedIds}
            onSelect={setSelected}
            onHover={setHover}
          />
        </div>
        <SidePanel module={sel} />
      </div>
    </div>
  );
}

function SwimlaneCanvas({ modules, edges, layout, selected, hover, connectedIds, onSelect, onHover }) {
  const { positions, edgePaths, totalWidth, totalHeight, layerYStarts, layerHeights, laneCols, laneWidth, laneGapX, layerOrder, laneOrder } = layout;
  return (
    <svg width={totalWidth} height={totalHeight} style={{ display: "block", minWidth: "100%" }}>
      <defs>
        {Object.entries(CLUSTERS).map(([k, c]) => (
          <marker key={k} id={`arrB-${k}`} viewBox="0 0 8 8" refX="7" refY="4" markerWidth="6" markerHeight="6" orient="auto">
            <path d="M0,0 L8,4 L0,8 z" fill={c.color} opacity="0.7" />
          </marker>
        ))}
        <marker id="arrB-dim" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M0,0 L8,4 L0,8 z" fill={T.borderStrong} />
        </marker>
      </defs>

      {/* lane backgrounds */}
      {laneOrder.map((cl, ci) => {
        const c = CLUSTERS[cl];
        const x = laneCols[ci];
        return (
          <g key={cl}>
            <rect x={x} y={40} width={laneWidth + 28} height={totalHeight - 60} rx="8" fill={c.bg} opacity="0.35" />
            <rect x={x} y={40} width={laneWidth + 28} height={32} rx="8" fill={c.bg} opacity="0.85" />
            <rect x={x} y={56} width={laneWidth + 28} height={16} fill={c.bg} opacity="0.85" />
            <text x={x + (laneWidth + 28) / 2} y={60} textAnchor="middle" fontSize="12" fontWeight="700" fill={c.color} fontFamily={T.font} letterSpacing="0.5">{c.label.toUpperCase()}</text>
          </g>
        );
      })}

      {/* layer separators + labels */}
      {layerOrder.map((ly, li) => {
        const y = layerYStarts[li];
        return (
          <g key={ly}>
            {li > 0 && <line x1={20} y1={y - 6} x2={totalWidth - 20} y2={y - 6} stroke={T.border} strokeDasharray="2 4" />}
            <text x={26} y={y + 6} fontSize="10" fontWeight="600" fill={T.ink3} fontFamily={T.mono} letterSpacing="0.5">{LAYERS[ly].label.toUpperCase()}</text>
          </g>
        );
      })}

      {/* edges */}
      {edgePaths.map((e, i) => {
        const fromMod = modules.find((m) => m.id === e.from);
        const cluster = fromMod ? fromMod.cluster : "modified";
        const dim = (selected || hover) && !(connectedIds.has(e.from) && connectedIds.has(e.to));
        return (
          <path key={i} d={e.d} fill="none"
            stroke={dim ? T.borderStrong : CLUSTERS[cluster].color}
            strokeWidth={dim ? 1 : 1.3}
            opacity={dim ? 0.3 : (e.sameLane ? 0.55 : 0.65)}
            strokeDasharray={e.sameLane ? "none" : "3 3"}
            markerEnd={dim ? "url(#arrB-dim)" : `url(#arrB-${cluster})`}
          />
        );
      })}

      {/* nodes */}
      {modules.map((m) => {
        const p = positions[m.id];
        if (!p) return null;
        const c = CLUSTERS[m.cluster];
        const isSel = selected === m.id;
        const dim = (selected || hover) && !connectedIds.has(m.id);
        return (
          <g key={m.id} transform={`translate(${p.x},${p.y})`}
            onClick={() => onSelect(m.id)}
            onMouseEnter={() => onHover(m.id)} onMouseLeave={() => onHover(null)}
            style={{ cursor: "pointer", opacity: dim ? 0.3 : 1 }}>
            <rect width={p.w} height={p.h} rx="5" fill="#fff"
              stroke={isSel ? T.ink : c.border}
              strokeWidth={isSel ? 2 : 1} />
            <rect x="0" y="0" width={p.w} height="3" rx="2" fill={c.color} />
            <text x="10" y="22" fontSize="11" fontWeight="600" fill={T.ink} fontFamily={T.font}>
              {m.name.length > 18 ? m.name.slice(0, 17) + "…" : m.name}
            </text>
            <text x="10" y="36" fontSize="9" fill={T.ink3} fontFamily={T.mono}>
              <tspan fill={T.added}>+{m.additions}</tspan>
              <tspan> </tspan>
              <tspan fill={T.removed}>−{m.deletions}</tspan>
              <tspan fill={T.ink3}> · {m.functions.length}fn</tspan>
            </text>
          </g>
        );
      })}
    </svg>
  );
}

window.VariantB = VariantB;
