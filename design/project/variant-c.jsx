// variant-c.jsx — 案C: ダッシュボード型サマリー
// 4クラスタを4象限に分けて、各セクションにモジュールカードを並べる
// グラフは中央に小さめのオーバービュー、左右に統計

function VariantC() {
  const [selected, setSelected] = React.useState("m11");
  const [activeCluster, setActiveCluster] = React.useState(null);

  const sel = MODULES.find((m) => m.id === selected);

  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, display: "flex", flexDirection: "column", overflow: "hidden" }}>
      {/* top bar */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
          <Brand size={14} />
          <div style={{ width: 1, height: 18, background: T.border }} />
          <div style={{ fontSize: 12, fontWeight: 600, color: T.ink, letterSpacing: -0.2 }}>{PR_META.title}</div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 10, fontSize: 11, color: T.ink3, fontFamily: T.mono }}>
          <span>{PR_META.repo}#{PR_META.number}</span>
          <span style={{ color: T.added }}>+{PR_META.additions}</span>
          <span style={{ color: T.removed }}>−{PR_META.deletions}</span>
        </div>
      </div>

      {/* hero stats row */}
      <div style={{ padding: "14px 20px", background: "#fff", borderBottom: `1px solid ${T.border}`, display: "grid", gridTemplateColumns: "1.4fr repeat(4, 1fr)", gap: 12, alignItems: "stretch" }}>
        {/* total card */}
        <div style={{ background: T.ink, color: "#fff", borderRadius: 10, padding: "12px 16px", display: "flex", flexDirection: "column", justifyContent: "space-between" }}>
          <div style={{ fontSize: 11, opacity: 0.6, fontWeight: 500, letterSpacing: 0.4 }}>変更モジュール総数</div>
          <div style={{ display: "flex", alignItems: "baseline", gap: 8, marginTop: 6 }}>
            <span style={{ fontSize: 32, fontWeight: 600, fontVariantNumeric: "tabular-nums", letterSpacing: -1 }}>{MODULES.length}</span>
            <span style={{ fontSize: 12, opacity: 0.5 }}>modules</span>
          </div>
          <div style={{ fontSize: 10, opacity: 0.55, fontFamily: T.mono, marginTop: 4 }}>{EDGES.length} 依存エッジ · {PR_META.commits} commits</div>
        </div>
        {Object.entries(CLUSTERS).map(([k, c]) => {
          const stat = CLUSTER_STATS[k];
          const active = activeCluster === k;
          return (
            <button key={k} onClick={() => setActiveCluster(active ? null : k)} style={{
              background: active ? c.color : c.bg, color: active ? "#fff" : c.color,
              border: `1px solid ${active ? c.color : c.border}`, borderRadius: 10, padding: "12px 14px",
              display: "flex", flexDirection: "column", justifyContent: "space-between",
              fontFamily: "inherit", cursor: "pointer", textAlign: "left", transition: "all 0.15s",
            }}>
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
                <span style={{ fontSize: 11, fontWeight: 600, letterSpacing: 0.5 }}>{c.label.toUpperCase()}</span>
                <span style={{ width: 20, height: 20, borderRadius: 5, background: active ? "rgba(255,255,255,0.2)" : "#fff", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 12, fontWeight: 700 }}>{c.icon}</span>
              </div>
              <div style={{ display: "flex", alignItems: "baseline", gap: 6, marginTop: 8 }}>
                <span style={{ fontSize: 28, fontWeight: 600, fontVariantNumeric: "tabular-nums", letterSpacing: -0.8 }}>{stat.count}</span>
                <span style={{ fontSize: 11, opacity: 0.7 }}>mods</span>
              </div>
              <div style={{ fontSize: 10, fontFamily: T.mono, opacity: active ? 0.85 : 0.7, marginTop: 2 }}>+{stat.additions} −{stat.deletions}</div>
            </button>
          );
        })}
      </div>

      {/* 4-quadrant grid + center mini graph */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 1fr", gridTemplateRows: "1fr 1fr", gap: 12, padding: 16, overflow: "hidden" }}>
        {Object.entries(CLUSTERS).map(([k, c]) => (
          <ClusterQuadrant
            key={k}
            cluster={k}
            meta={c}
            modules={MODULES.filter((m) => m.cluster === k)}
            selected={selected}
            dim={activeCluster && activeCluster !== k}
            onSelect={setSelected}
          />
        ))}
      </div>
    </div>
  );
}

function ClusterQuadrant({ cluster, meta, modules, selected, dim, onSelect }) {
  return (
    <div style={{
      background: "#fff", borderRadius: 12, border: `1px solid ${T.border}`,
      borderLeft: `3px solid ${meta.color}`,
      display: "flex", flexDirection: "column", overflow: "hidden",
      opacity: dim ? 0.4 : 1, transition: "opacity 0.15s",
    }}>
      <div style={{ padding: "10px 14px", display: "flex", alignItems: "center", justifyContent: "space-between", borderBottom: `1px solid ${T.border}` }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span style={{ width: 22, height: 22, borderRadius: 5, background: meta.bg, color: meta.color, display: "flex", alignItems: "center", justifyContent: "center", fontWeight: 700, fontSize: 12 }}>{meta.icon}</span>
          <span style={{ fontSize: 13, fontWeight: 600, color: T.ink }}>{meta.label}</span>
          <span style={{ fontSize: 11, color: T.ink3, fontFamily: T.mono }}>· {modules.length} modules</span>
        </div>
        <div style={{ fontSize: 11, fontFamily: T.mono, color: T.ink3 }}>
          <span style={{ color: T.added }}>+{modules.reduce((s, m) => s + m.additions, 0)}</span>
          {" "}
          <span style={{ color: T.removed }}>−{modules.reduce((s, m) => s + m.deletions, 0)}</span>
        </div>
      </div>

      <div style={{ flex: 1, overflowY: "auto", padding: 8 }}>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(140px, 1fr))", gap: 6 }}>
          {modules.map((m) => {
            const isSel = selected === m.id;
            const max = Math.max(...modules.map((x) => x.additions + x.deletions), 1);
            const intensity = (m.additions + m.deletions) / max;
            const total = m.additions + m.deletions;
            return (
              <button key={m.id} onClick={() => onSelect(m.id)} style={{
                background: isSel ? meta.bg : "#fff",
                border: `1px solid ${isSel ? meta.color : T.border}`,
                borderRadius: 7, padding: "8px 10px", cursor: "pointer", textAlign: "left",
                fontFamily: "inherit", display: "flex", flexDirection: "column", gap: 4,
                position: "relative", overflow: "hidden",
              }}>
                {/* intensity bar at bottom */}
                <div style={{ position: "absolute", left: 0, bottom: 0, height: 2, width: `${intensity * 100}%`, background: meta.color, opacity: 0.5 }} />
                <div style={{ fontSize: 12, fontWeight: 600, color: T.ink, lineHeight: 1.2, textWrap: "pretty" }}>{m.name}</div>
                <div style={{ fontSize: 9, fontFamily: T.mono, color: T.ink3, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                  <span>{LAYERS[m.layer].label}</span>
                  <span>
                    {m.additions > 0 && <span style={{ color: T.added }}>+{m.additions}</span>}
                    {m.additions > 0 && m.deletions > 0 && " "}
                    {m.deletions > 0 && <span style={{ color: T.removed }}>−{m.deletions}</span>}
                  </span>
                </div>
                {/* sparkbar of functions */}
                <div style={{ display: "flex", gap: 2, marginTop: 2 }}>
                  {m.functions.slice(0, 8).map((fn, i) => {
                    const fc = CLUSTERS[fn.change] || CLUSTERS.modified;
                    return <div key={i} style={{ flex: 1, height: 3, borderRadius: 1, background: fc.color, opacity: 0.7 }} title={fn.name} />;
                  })}
                </div>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
}

window.VariantC = VariantC;
