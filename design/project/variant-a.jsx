// variant-a.jsx — 案A: クラシックDAG (右パネルでdiff詳細)
// 階層レイアウトで上から下へ流れる依存グラフ + クリックで右にdiff

function VariantA() {
  const [selected, setSelected] = React.useState("m9");
  const [hover, setHover] = React.useState(null);
  const [filter, setFilter] = React.useState("all");

  const filteredModules = filter === "all" ? MODULES : MODULES.filter((m) => m.cluster === filter);
  const filteredIds = new Set(filteredModules.map((m) => m.id));
  const filteredEdges = EDGES.filter(([a, b]) => filteredIds.has(a) && filteredIds.has(b));

  const layout = React.useMemo(
    () => computeDagLayout(filteredModules, filteredEdges, { nodeWidth: 134, nodeHeight: 50, layerGapY: 116, nodeGapX: 14 }),
    [filter]
  );

  // hover/selected に関わるエッジをハイライト
  const highlightId = hover || selected;
  const connectedIds = React.useMemo(() => {
    if (!highlightId) return new Set();
    const s = new Set([highlightId]);
    for (const [a, b] of filteredEdges) {
      if (a === highlightId) s.add(b);
      if (b === highlightId) s.add(a);
    }
    return s;
  }, [highlightId, filteredEdges]);

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
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: T.mono, marginTop: 1 }}>{PR_META.repo} #{PR_META.number} · {PR_META.branch} → {PR_META.base}</div>
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12, fontSize: 12 }}>
          <span style={{ color: T.added, fontFamily: T.mono, fontWeight: 500 }}>+{PR_META.additions}</span>
          <span style={{ color: T.removed, fontFamily: T.mono, fontWeight: 500 }}>−{PR_META.deletions}</span>
          <span style={{ color: T.ink3, fontFamily: T.mono }}>{PR_META.filesChanged} files</span>
          <div style={{ width: 24, height: 24, borderRadius: 99, background: T.brand, color: "#fff", fontSize: 10, fontWeight: 600, display: "flex", alignItems: "center", justifyContent: "center" }}>{PR_META.authorAvatar}</div>
        </div>
      </div>

      {/* tabs / filters */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "8px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <div style={{ display: "flex", gap: 4 }}>
          {[
            { k: "all", label: "全て", n: MODULES.length, c: T.ink },
            { k: "added", label: "追加", n: CLUSTER_STATS.added.count, c: T.added, bg: T.addedBg },
            { k: "removed", label: "削除", n: CLUSTER_STATS.removed.count, c: T.removed, bg: T.removedBg },
            { k: "modified", label: "修正", n: CLUSTER_STATS.modified.count, c: T.modified, bg: T.modifiedBg },
            { k: "refactor", label: "リファクタ", n: CLUSTER_STATS.refactor.count, c: T.refactor, bg: T.refactorBg },
          ].map((f) => {
            const active = filter === f.k;
            return (
              <button key={f.k} onClick={() => setFilter(f.k)} style={{
                padding: "6px 12px", borderRadius: 7, border: "none", cursor: "pointer", fontFamily: "inherit",
                fontSize: 12, fontWeight: active ? 600 : 500,
                background: active ? (f.bg || T.bgAlt) : "transparent",
                color: active ? f.c : T.ink2,
                display: "flex", alignItems: "center", gap: 6,
              }}>
                {f.k !== "all" && <span style={{ width: 6, height: 6, borderRadius: 99, background: f.c }} />}
                <span>{f.label}</span>
                <span style={{ fontFamily: T.mono, fontSize: 11, opacity: 0.7 }}>{f.n}</span>
              </button>
            );
          })}
        </div>
        <div style={{ display: "flex", gap: 6, fontSize: 11, color: T.ink3, fontFamily: T.mono }}>
          <span>レイアウト: 階層DAG</span>
          <span>·</span>
          <span>34 nodes / {EDGES.length} edges</span>
        </div>
      </div>

      {/* main: graph + side panel */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 380px", overflow: "hidden" }}>
        {/* graph */}
        <div style={{ position: "relative", overflow: "auto", background: T.bg }}>
          <GraphCanvas
            modules={filteredModules}
            edges={filteredEdges}
            layout={layout}
            selected={selected}
            hover={hover}
            connectedIds={connectedIds}
            onSelect={setSelected}
            onHover={setHover}
          />
          {/* layer labels overlay */}
          <div style={{ position: "absolute", left: 12, top: 60, display: "flex", flexDirection: "column", gap: 90, fontSize: 10, fontFamily: T.mono, color: T.ink3, pointerEvents: "none" }}>
            {["UI", "DOMAIN", "DATA", "INFRA"].map((l) => (
              <div key={l} style={{ background: "#fff", padding: "3px 6px", borderRadius: 4, border: `1px solid ${T.border}`, height: 18, display: "flex", alignItems: "center" }}>{l}</div>
            ))}
          </div>
          {/* legend */}
          <div style={{ position: "absolute", right: 14, bottom: 14, background: "#fff", border: `1px solid ${T.border}`, borderRadius: 8, padding: "8px 12px", boxShadow: "0 2px 8px rgba(0,0,0,0.04)", fontSize: 11 }}>
            <div style={{ fontWeight: 600, marginBottom: 6, color: T.ink2 }}>凡例</div>
            <div style={{ display: "grid", gridTemplateColumns: "auto auto", gap: "4px 12px" }}>
              {Object.entries(CLUSTERS).map(([k, c]) => (
                <div key={k} style={{ display: "flex", alignItems: "center", gap: 5, color: T.ink2 }}>
                  <div style={{ width: 9, height: 9, borderRadius: 2, background: c.bg, border: `1px solid ${c.border}` }} />
                  <span>{c.label}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* side panel */}
        <SidePanel module={sel} />
      </div>
    </div>
  );
}

// ─────────── Graph SVG ───────────
function GraphCanvas({ modules, edges, layout, selected, hover, connectedIds, onSelect, onHover }) {
  const { positions, edgePaths, totalWidth, totalHeight } = layout;
  return (
    <svg width={totalWidth} height={totalHeight} style={{ display: "block", minWidth: "100%" }}>
      <defs>
        {Object.entries(CLUSTERS).map(([k, c]) => (
          <marker key={k} id={`arr-${k}`} viewBox="0 0 8 8" refX="7" refY="4" markerWidth="6" markerHeight="6" orient="auto">
            <path d="M0,0 L8,4 L0,8 z" fill={c.color} opacity="0.7" />
          </marker>
        ))}
        <marker id="arr-dim" viewBox="0 0 8 8" refX="7" refY="4" markerWidth="5" markerHeight="5" orient="auto">
          <path d="M0,0 L8,4 L0,8 z" fill={T.borderStrong} />
        </marker>
      </defs>

      {/* edges */}
      {edgePaths.map((e, i) => {
        const fromMod = modules.find((m) => m.id === e.from);
        const cluster = fromMod ? fromMod.cluster : "modified";
        const dim = (selected || hover) && !(connectedIds.has(e.from) && connectedIds.has(e.to));
        const color = dim ? T.borderStrong : CLUSTERS[cluster].color;
        return (
          <path key={i} d={e.d} fill="none" stroke={color}
            strokeWidth={dim ? 1 : 1.4}
            opacity={dim ? 0.4 : 0.75}
            markerEnd={dim ? "url(#arr-dim)" : `url(#arr-${cluster})`}
          />
        );
      })}

      {/* nodes */}
      {modules.map((m) => {
        const p = positions[m.id];
        if (!p) return null;
        const c = CLUSTERS[m.cluster];
        const isSel = selected === m.id;
        const isHover = hover === m.id;
        const dim = (selected || hover) && !connectedIds.has(m.id);
        return (
          <g key={m.id} transform={`translate(${p.x},${p.y})`}
            onClick={() => onSelect(m.id)}
            onMouseEnter={() => onHover(m.id)} onMouseLeave={() => onHover(null)}
            style={{ cursor: "pointer", opacity: dim ? 0.35 : 1, transition: "opacity 0.15s" }}>
            <rect width={p.w} height={p.h} rx="6" ry="6"
              fill={c.bg}
              stroke={isSel ? T.ink : isHover ? c.color : c.border}
              strokeWidth={isSel ? 2 : isHover ? 1.5 : 1} />
            <rect x="0" y="0" width="3" height={p.h} rx="2" fill={c.color} />
            <text x={p.w / 2} y="20" textAnchor="middle" fontSize="11" fontWeight="600" fill={T.ink} fontFamily={T.font}>
              {m.name.length > 16 ? m.name.slice(0, 15) + "…" : m.name}
            </text>
            <text x={p.w / 2} y="35" textAnchor="middle" fontSize="9" fill={c.color} fontFamily={T.mono} fontWeight="500">
              +{m.additions} −{m.deletions}
            </text>
            <text x={p.w / 2} y="46" textAnchor="middle" fontSize="8" fill={T.ink3} fontFamily={T.mono}>
              {m.functions.length} fn
            </text>
          </g>
        );
      })}
    </svg>
  );
}

// ─────────── Side Panel ───────────
function SidePanel({ module: m }) {
  if (!m) return <div style={{ borderLeft: `1px solid ${T.border}`, background: "#fff" }} />;
  const c = CLUSTERS[m.cluster];
  const diff = SAMPLE_DIFF[m.id];
  return (
    <div style={{ borderLeft: `1px solid ${T.border}`, background: "#fff", display: "flex", flexDirection: "column", overflow: "hidden" }}>
      {/* header */}
      <div style={{ padding: "16px 18px", borderBottom: `1px solid ${T.border}` }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
          <span style={{ fontSize: 10, fontWeight: 600, color: c.color, background: c.bg, border: `1px solid ${c.border}`, padding: "2px 7px", borderRadius: 99, letterSpacing: 0.5, textTransform: "uppercase" }}>{c.short}</span>
          <span style={{ fontSize: 10, color: T.ink3, fontFamily: T.mono, padding: "2px 6px", background: T.bgAlt, borderRadius: 4 }}>{LAYERS[m.layer].label}</span>
        </div>
        <div style={{ fontSize: 16, fontWeight: 600, letterSpacing: -0.3, color: T.ink, marginBottom: 4, textWrap: "pretty" }}>{m.name}</div>
        <div style={{ fontSize: 11, fontFamily: T.mono, color: T.ink3, wordBreak: "break-all" }}>{m.path}</div>
        <div style={{ display: "flex", gap: 14, marginTop: 12, fontSize: 12 }}>
          <div><span style={{ color: T.added, fontWeight: 600, fontFamily: T.mono }}>+{m.additions}</span> <span style={{ color: T.ink3 }}>追加</span></div>
          <div><span style={{ color: T.removed, fontWeight: 600, fontFamily: T.mono }}>−{m.deletions}</span> <span style={{ color: T.ink3 }}>削除</span></div>
          <div><span style={{ color: T.ink, fontWeight: 600, fontFamily: T.mono }}>{m.loc}</span> <span style={{ color: T.ink3 }}>LOC</span></div>
        </div>
      </div>

      {/* functions */}
      <div style={{ padding: "12px 18px", borderBottom: `1px solid ${T.border}` }}>
        <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, marginBottom: 8, letterSpacing: 0.5 }}>変更された関数 ({m.functions.length})</div>
        <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
          {m.functions.map((fn, i) => {
            const fc = CLUSTERS[fn.change] || CLUSTERS.modified;
            return (
              <div key={i} style={{ display: "flex", alignItems: "center", gap: 8, padding: "6px 8px", background: T.bgAlt, borderRadius: 5 }}>
                <span style={{ width: 14, height: 14, borderRadius: 3, background: fc.bg, border: `1px solid ${fc.border}`, color: fc.color, fontSize: 10, display: "flex", alignItems: "center", justifyContent: "center", fontWeight: 700 }}>{fc.icon}</span>
                <span style={{ fontFamily: T.mono, fontSize: 12, color: T.ink, flex: 1 }}>{fn.name}</span>
                <span style={{ fontSize: 10, color: fc.color, textTransform: "uppercase", fontWeight: 600, letterSpacing: 0.4 }}>{fc.short}</span>
              </div>
            );
          })}
        </div>
      </div>

      {/* diff */}
      <div style={{ flex: 1, overflow: "auto", background: T.bgAlt }}>
        <div style={{ padding: "10px 18px", fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5, position: "sticky", top: 0, background: T.bgAlt, borderBottom: `1px solid ${T.border}` }}>
          DIFF プレビュー
        </div>
        {diff ? (
          <div style={{ fontFamily: T.mono, fontSize: 11, lineHeight: 1.55 }}>
            {diff.hunks.map((h, hi) => (
              <div key={hi}>
                <div style={{ padding: "4px 18px", background: "#e0e7ff", color: "#3730a3", fontSize: 10 }}>{h.header}</div>
                {h.lines.map((ln, li) => {
                  const bg = ln.kind === "add" ? "#ecfdf5" : ln.kind === "del" ? "#fef2f2" : "transparent";
                  const sign = ln.kind === "add" ? "+" : ln.kind === "del" ? "−" : " ";
                  const col = ln.kind === "add" ? T.added : ln.kind === "del" ? T.removed : T.ink2;
                  return (
                    <div key={li} style={{ display: "flex", background: bg }}>
                      <div style={{ width: 32, padding: "0 8px", textAlign: "right", color: T.ink3, fontSize: 10, borderRight: `1px solid ${T.border}` }}>{ln.n}</div>
                      <div style={{ width: 14, textAlign: "center", color: col, fontWeight: 600 }}>{sign}</div>
                      <div style={{ flex: 1, padding: "0 8px", color: T.ink, whiteSpace: "pre" }}>{ln.text}</div>
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        ) : (
          <div style={{ padding: "20px 18px", fontSize: 12, color: T.ink3, fontStyle: "italic" }}>
            このモジュールのdiffプレビューはサンプルに含まれていません。<br/>実際のツールでは GitHub API から取得します。
          </div>
        )}
      </div>
    </div>
  );
}

window.VariantA = VariantA;
