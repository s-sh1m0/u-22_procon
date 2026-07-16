// impact.jsx — diff影響範囲の計算 + 可視化案D
// 「変更されたモジュール」を起点に、依存元（呼び出し側）へ何hop伝播するかを計算

// 影響範囲を計算: 変更ノード集合から、エッジを「依存される→依存する」逆方向に辿る
// 結果: { moduleId: { hop: 0|1|2|3..., source: changedModuleId } }
// hop=0: 自分自身が変更されたモジュール
// hop>=1: 変更モジュールに依存している（影響を受ける可能性あり）
function computeImpact(modules, edges, changedClusters = ["added", "removed", "modified", "refactor"]) {
  const isChanged = new Set(modules.filter(m => changedClusters.includes(m.cluster)).map(m => m.id));
  // 逆向きエッジ: m -> [依存しているモジュール群]
  const reverseAdj = {};
  for (const [a, b] of edges) {
    // a -> b の意味は「a が b を使う / 依存する」
    // つまり b が変わると a が影響を受ける
    if (!reverseAdj[b]) reverseAdj[b] = [];
    reverseAdj[b].push(a);
  }
  const impact = {};
  for (const m of modules) {
    if (isChanged.has(m.id)) {
      impact[m.id] = { hop: 0, fromChange: true };
    }
  }
  // BFS
  let frontier = Array.from(isChanged);
  let hop = 0;
  while (frontier.length > 0 && hop < 5) {
    hop++;
    const next = [];
    for (const id of frontier) {
      const callers = reverseAdj[id] || [];
      for (const c of callers) {
        if (impact[c] === undefined || impact[c].hop > hop) {
          impact[c] = { hop, fromChange: isChanged.has(c) };
          next.push(c);
        }
      }
    }
    frontier = next;
  }
  return impact;
}

const IMPACT_LEVELS = {
  0: { label: "変更", color: "#0f766e", bg: "#ccfbf1", border: "#5eead4" },
  1: { label: "直接影響", color: "#ea580c", bg: "#ffedd5", border: "#fdba74" },
  2: { label: "2hop", color: "#ca8a04", bg: "#fef9c3", border: "#fde047" },
  3: { label: "3hop+", color: "#65a30d", bg: "#ecfccb", border: "#bef264" },
  none: { label: "影響なし", color: "#a8a29e", bg: "#f5f5f4", border: "#e7e5e4" },
};

function impactLevel(impactEntry) {
  if (!impactEntry) return "none";
  if (impactEntry.hop === 0) return 0;
  if (impactEntry.hop === 1) return 1;
  if (impactEntry.hop === 2) return 2;
  return 3;
}

// ─────────── 案D: 影響範囲ヒートマップビュー ───────────
function VariantD() {
  const [selected, setSelected] = React.useState("m9");
  const [hover, setHover] = React.useState(null);
  const [mode, setMode] = React.useState("change"); // change|select

  const impact = React.useMemo(() => computeImpact(MODULES, EDGES), []);

  // 選択モジュールの波紋（選択モード時用）
  const ripple = React.useMemo(() => {
    if (mode !== "select") return null;
    return computeImpact(MODULES, EDGES, []).constructor === Object
      ? computeImpactFrom(selected, MODULES, EDGES)
      : null;
  }, [mode, selected]);

  const layout = React.useMemo(
    () => computeDagLayout(MODULES, EDGES, { nodeWidth: 130, nodeHeight: 48, layerGapY: 110, nodeGapX: 10 }),
    []
  );

  const sel = MODULES.find(m => m.id === selected);
  const activeImpact = mode === "select" ? ripple : impact;

  // 影響統計
  const impactStats = React.useMemo(() => {
    const counts = { 0: 0, 1: 0, 2: 0, 3: 0, none: 0 };
    for (const m of MODULES) {
      const lv = impactLevel(activeImpact && activeImpact[m.id]);
      counts[lv]++;
    }
    return counts;
  }, [activeImpact]);

  // 各モジュールの影響を受ける関数の数の概算
  const totalAffected = (impactStats[1] || 0) + (impactStats[2] || 0) + (impactStats[3] || 0);

  return (
    <div style={{ width: "100%", height: "100%", background: T.bg, fontFamily: T.font, color: T.ink, display: "flex", flexDirection: "column", overflow: "hidden" }}>
      {/* top bar */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <Brand size={15} />
          <div style={{ width: 1, height: 20, background: T.border }} />
          <div>
            <div style={{ fontSize: 13, fontWeight: 600, color: T.ink, letterSpacing: -0.2 }}>影響範囲ビュー — {PR_META.title}</div>
            <div style={{ fontSize: 11, color: T.ink3, fontFamily: T.mono, marginTop: 1 }}>{PR_META.repo} #{PR_META.number} · diff伝播解析</div>
          </div>
        </div>
        <div style={{ display: "flex", gap: 4, background: T.bgAlt, borderRadius: 8, padding: 3 }}>
          {[
            { k: "change", label: "変更全体の影響" },
            { k: "select", label: "選択モジュールの波紋" },
          ].map(t => (
            <button key={t.k} onClick={() => setMode(t.k)} style={{
              padding: "6px 12px", borderRadius: 6, border: "none", cursor: "pointer", fontFamily: "inherit",
              fontSize: 12, fontWeight: mode === t.k ? 600 : 500,
              background: mode === t.k ? "#fff" : "transparent",
              color: mode === t.k ? T.ink : T.ink3,
              boxShadow: mode === t.k ? "0 1px 2px rgba(0,0,0,0.06)" : "none",
            }}>{t.label}</button>
          ))}
        </div>
      </div>

      {/* impact stats */}
      <div style={{ padding: "10px 24px", borderBottom: `1px solid ${T.border}`, background: "#fff", display: "flex", gap: 10, alignItems: "center" }}>
        <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5 }}>影響度分布</div>
        <div style={{ flex: 1, height: 24, background: T.bgAlt, borderRadius: 6, overflow: "hidden", display: "flex" }}>
          {[0, 1, 2, 3, "none"].map(lv => {
            const il = IMPACT_LEVELS[lv];
            const count = impactStats[lv];
            const pct = (count / MODULES.length) * 100;
            if (pct === 0) return null;
            return (
              <div key={lv} style={{ width: `${pct}%`, background: il.color, opacity: lv === "none" ? 0.25 : 0.85,
                display: "flex", alignItems: "center", justifyContent: "center", color: lv === "none" ? T.ink3 : "#fff",
                fontSize: 10, fontWeight: 600, fontFamily: T.mono }}>
                {pct > 6 ? `${il.label} ${count}` : count}
              </div>
            );
          })}
        </div>
        <div style={{ fontSize: 11, color: T.ink2, fontFamily: T.mono }}>
          <span style={{ color: T.ink, fontWeight: 600 }}>{totalAffected}</span> / {MODULES.length} 件が影響を受ける可能性
        </div>
      </div>

      {/* main */}
      <div style={{ flex: 1, display: "grid", gridTemplateColumns: "1fr 320px", overflow: "hidden" }}>
        {/* graph with ripple coloring */}
        <div style={{ position: "relative", overflow: "auto", background: T.bg }}>
          <ImpactGraph
            modules={MODULES}
            edges={EDGES}
            layout={layout}
            impact={activeImpact}
            selected={selected}
            hover={hover}
            mode={mode}
            onSelect={setSelected}
            onHover={setHover}
          />
          {/* impact legend */}
          <div style={{ position: "absolute", right: 14, bottom: 14, background: "#fff", border: `1px solid ${T.border}`, borderRadius: 8, padding: "10px 14px", boxShadow: "0 2px 8px rgba(0,0,0,0.04)", fontSize: 11 }}>
            <div style={{ fontWeight: 600, marginBottom: 8, color: T.ink2 }}>影響度</div>
            <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
              {[0, 1, 2, 3, "none"].map(lv => {
                const il = IMPACT_LEVELS[lv];
                return (
                  <div key={lv} style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <div style={{ width: 12, height: 12, borderRadius: 3, background: il.bg, border: `1.5px solid ${il.border}` }} />
                    <span style={{ color: T.ink2 }}>{il.label}</span>
                    <span style={{ fontFamily: T.mono, color: T.ink3, marginLeft: "auto" }}>{impactStats[lv]}</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* impact panel */}
        <ImpactPanel module={sel} impact={activeImpact} mode={mode} />
      </div>
    </div>
  );
}

// 選択モジュールから下流（依存元）へ伝播
function computeImpactFrom(rootId, modules, edges) {
  const reverseAdj = {};
  for (const [a, b] of edges) {
    if (!reverseAdj[b]) reverseAdj[b] = [];
    reverseAdj[b].push(a);
  }
  const impact = { [rootId]: { hop: 0, fromChange: true } };
  let frontier = [rootId];
  let hop = 0;
  while (frontier.length > 0 && hop < 5) {
    hop++;
    const next = [];
    for (const id of frontier) {
      const callers = reverseAdj[id] || [];
      for (const c of callers) {
        if (impact[c] === undefined) {
          impact[c] = { hop, fromChange: false };
          next.push(c);
        }
      }
    }
    frontier = next;
  }
  return impact;
}

function ImpactGraph({ modules, edges, layout, impact, selected, hover, mode, onSelect, onHover }) {
  const { positions, edgePaths, totalWidth, totalHeight } = layout;
  return (
    <svg width={totalWidth} height={totalHeight} style={{ display: "block", minWidth: "100%" }}>
      <defs>
        {Object.entries(IMPACT_LEVELS).map(([k, il]) => (
          <marker key={k} id={`imp-arr-${k}`} viewBox="0 0 8 8" refX="7" refY="4" markerWidth="6" markerHeight="6" orient="auto">
            <path d="M0,0 L8,4 L0,8 z" fill={il.color} opacity="0.7" />
          </marker>
        ))}
      </defs>

      {/* edges — colored by source impact level */}
      {edgePaths.map((e, i) => {
        const aLv = impactLevel(impact && impact[e.from]);
        const bLv = impactLevel(impact && impact[e.to]);
        // 影響伝播エッジ: 両端ともにimpactあり
        const isRipple = aLv !== "none" && bLv !== "none";
        const il = IMPACT_LEVELS[isRipple ? Math.max(typeof aLv === "number" ? aLv : 0, typeof bLv === "number" ? bLv : 0) : "none"];
        return (
          <path key={i} d={e.d} fill="none"
            stroke={isRipple ? il.color : T.borderStrong}
            strokeWidth={isRipple ? 1.6 : 0.8}
            opacity={isRipple ? 0.8 : 0.3}
            markerEnd={`url(#imp-arr-${isRipple ? (typeof aLv === "number" ? aLv : 0) : "none"})`}
          />
        );
      })}

      {/* nodes */}
      {modules.map(m => {
        const p = positions[m.id];
        if (!p) return null;
        const lv = impactLevel(impact && impact[m.id]);
        const il = IMPACT_LEVELS[lv];
        const isSel = selected === m.id;
        const cluster = CLUSTERS[m.cluster];
        return (
          <g key={m.id} transform={`translate(${p.x},${p.y})`}
            onClick={() => onSelect(m.id)}
            onMouseEnter={() => onHover(m.id)} onMouseLeave={() => onHover(null)}
            style={{ cursor: "pointer" }}>
            {/* 波紋（hop>=1のみ） */}
            {typeof lv === "number" && lv >= 1 && (
              <rect x="-4" y="-4" width={p.w + 8} height={p.h + 8} rx="9" fill="none" stroke={il.color} strokeWidth="1.5" opacity={0.4 - lv * 0.08} strokeDasharray="3 3" />
            )}
            <rect width={p.w} height={p.h} rx="6"
              fill={il.bg}
              stroke={isSel ? T.ink : il.border}
              strokeWidth={isSel ? 2 : 1.2}
              opacity={lv === "none" ? 0.55 : 1} />
            {/* hop badge */}
            {typeof lv === "number" && (
              <g>
                <circle cx={p.w - 10} cy="10" r="8" fill={il.color} />
                <text x={p.w - 10} y="13" textAnchor="middle" fontSize="9" fontWeight="700" fill="#fff" fontFamily={T.mono}>{lv}</text>
              </g>
            )}
            {/* cluster left bar */}
            <rect x="0" y="0" width="3" height={p.h} rx="2" fill={cluster.color} opacity={lv === "none" ? 0.3 : 0.8} />
            <text x="9" y="18" fontSize="11" fontWeight="600" fill={T.ink} fontFamily={T.font} opacity={lv === "none" ? 0.5 : 1}>
              {m.name.length > 17 ? m.name.slice(0, 16) + "…" : m.name}
            </text>
            <text x="9" y="32" fontSize="9" fill={il.color} fontFamily={T.mono} fontWeight="500" opacity={lv === "none" ? 0.5 : 1}>
              {lv === 0 ? "● 変更元" : lv === "none" ? "影響なし" : `${lv}-hop 影響`}
            </text>
            <text x="9" y="42" fontSize="8" fill={T.ink3} fontFamily={T.mono} opacity={0.7}>
              {LAYERS[m.layer].label} · {m.functions.length}fn
            </text>
          </g>
        );
      })}
    </svg>
  );
}

function ImpactPanel({ module: m, impact, mode }) {
  if (!m) return <div style={{ borderLeft: `1px solid ${T.border}`, background: "#fff" }} />;
  const lv = impactLevel(impact && impact[m.id]);
  const il = IMPACT_LEVELS[lv];
  const cluster = CLUSTERS[m.cluster];

  // この選択モジュールが影響を及ぼすモジュール一覧
  const downstream = [];
  const upstream = [];
  for (const [a, b] of EDGES) {
    if (a === m.id) downstream.push(b);
    if (b === m.id) upstream.push(a);
  }
  const downstreamMods = downstream.map(id => MODULES.find(x => x.id === id)).filter(Boolean);
  const upstreamMods = upstream.map(id => MODULES.find(x => x.id === id)).filter(Boolean);

  // hop別の影響先
  const buckets = { 1: [], 2: [], 3: [] };
  if (mode === "select") {
    for (const mod of MODULES) {
      const e = impact && impact[mod.id];
      if (!e || e.hop === 0) continue;
      const b = e.hop >= 3 ? 3 : e.hop;
      buckets[b].push(mod);
    }
  }

  return (
    <div style={{ borderLeft: `1px solid ${T.border}`, background: "#fff", display: "flex", flexDirection: "column", overflow: "hidden" }}>
      <div style={{ padding: "16px 18px", borderBottom: `1px solid ${T.border}`, background: il.bg }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 6 }}>
          <span style={{ fontSize: 10, fontWeight: 700, color: il.color, background: "#fff", padding: "2px 8px", borderRadius: 99, letterSpacing: 0.4, textTransform: "uppercase" }}>{il.label}</span>
          <span style={{ fontSize: 10, color: cluster.color, background: "#fff", border: `1px solid ${cluster.border}`, padding: "2px 7px", borderRadius: 99, fontWeight: 600 }}>{cluster.short}</span>
          <span style={{ fontSize: 10, color: T.ink3, fontFamily: T.mono, padding: "2px 6px", background: "rgba(255,255,255,0.6)", borderRadius: 4 }}>{LAYERS[m.layer].label}</span>
        </div>
        <div style={{ fontSize: 16, fontWeight: 600, letterSpacing: -0.3, color: T.ink, marginBottom: 4 }}>{m.name}</div>
        <div style={{ fontSize: 11, fontFamily: T.mono, color: T.ink3, wordBreak: "break-all" }}>{m.path}</div>
      </div>

      <div style={{ flex: 1, overflow: "auto", padding: "12px 18px" }}>
        {/* 影響範囲サマリー */}
        <div style={{ marginBottom: 14 }}>
          <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5, marginBottom: 8 }}>
            {mode === "select" ? "このモジュールの波紋" : "依存関係サマリー"}
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
            <div style={{ padding: "10px 12px", background: T.bgAlt, borderRadius: 7 }}>
              <div style={{ fontSize: 10, color: T.ink3, fontWeight: 600 }}>↑ 依存元（影響を受ける）</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: T.ink, fontVariantNumeric: "tabular-nums", letterSpacing: -0.4 }}>{upstreamMods.length}</div>
            </div>
            <div style={{ padding: "10px 12px", background: T.bgAlt, borderRadius: 7 }}>
              <div style={{ fontSize: 10, color: T.ink3, fontWeight: 600 }}>↓ 依存先（影響を与える）</div>
              <div style={{ fontSize: 22, fontWeight: 600, color: T.ink, fontVariantNumeric: "tabular-nums", letterSpacing: -0.4 }}>{downstreamMods.length}</div>
            </div>
          </div>
        </div>

        {mode === "select" && (
          <div style={{ marginBottom: 14 }}>
            <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5, marginBottom: 8 }}>影響範囲（hop別）</div>
            {[1, 2, 3].map(hop => {
              const list = buckets[hop];
              const il2 = IMPACT_LEVELS[hop];
              if (list.length === 0) return null;
              return (
                <div key={hop} style={{ marginBottom: 8 }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 6, marginBottom: 4 }}>
                    <span style={{ width: 18, height: 18, borderRadius: 4, background: il2.color, color: "#fff", fontSize: 10, fontWeight: 700, fontFamily: T.mono, display: "flex", alignItems: "center", justifyContent: "center" }}>{hop}</span>
                    <span style={{ fontSize: 11, color: T.ink2, fontWeight: 600 }}>{il2.label}</span>
                    <span style={{ fontSize: 11, color: T.ink3, fontFamily: T.mono }}>{list.length}件</span>
                  </div>
                  <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                    {list.map(mod => (
                      <div key={mod.id} style={{ padding: "3px 7px", background: il2.bg, border: `1px solid ${il2.border}`, color: il2.color, borderRadius: 4, fontSize: 10, fontFamily: T.mono, fontWeight: 500 }}>
                        {mod.name}
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <div style={{ marginBottom: 14 }}>
          <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5, marginBottom: 6 }}>↑ 依存元 ({upstreamMods.length})</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 3 }}>
            {upstreamMods.slice(0, 6).map(mod => {
              const ml = impactLevel(impact && impact[mod.id]);
              const mil = IMPACT_LEVELS[ml];
              return (
                <div key={mod.id} style={{ display: "flex", alignItems: "center", gap: 6, padding: "5px 8px", background: T.bgAlt, borderRadius: 5, fontSize: 11 }}>
                  <div style={{ width: 6, height: 6, borderRadius: 99, background: mil.color }} />
                  <span style={{ fontFamily: T.mono, color: T.ink, flex: 1 }}>{mod.name}</span>
                  <span style={{ fontSize: 9, color: mil.color, fontWeight: 600 }}>{typeof ml === "number" ? `${ml}hop` : ""}</span>
                </div>
              );
            })}
          </div>
        </div>

        <div>
          <div style={{ fontSize: 11, color: T.ink3, fontWeight: 600, letterSpacing: 0.5, marginBottom: 6 }}>変更された関数</div>
          {m.functions.map((fn, i) => {
            const fc = CLUSTERS[fn.change] || CLUSTERS.modified;
            return (
              <div key={i} style={{ display: "flex", alignItems: "center", gap: 8, padding: "5px 8px", marginBottom: 3, background: T.bgAlt, borderRadius: 5 }}>
                <span style={{ width: 14, height: 14, borderRadius: 3, background: fc.bg, border: `1px solid ${fc.border}`, color: fc.color, fontSize: 10, display: "flex", alignItems: "center", justifyContent: "center", fontWeight: 700 }}>{fc.icon}</span>
                <span style={{ fontFamily: T.mono, fontSize: 11, color: T.ink, flex: 1 }}>{fn.name}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

window.VariantD = VariantD;
window.computeImpact = computeImpact;
window.computeImpactFrom = computeImpactFrom;
window.IMPACT_LEVELS = IMPACT_LEVELS;
