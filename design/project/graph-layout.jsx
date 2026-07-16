// graph-layout.jsx — DAG レイアウトヘルパ
// 入力: モジュール配列 + エッジ配列
// 出力: ノード位置 + エッジパス
// 階層型: レイヤごとに段を分ける + cluster でも同じレイヤ内でグルーピング

function computeDagLayout(modules, edges, opts = {}) {
  const {
    nodeWidth = 140,
    nodeHeight = 56,
    layerGapY = 140,
    nodeGapX = 24,
    layerOrder = ["ui", "domain", "data", "infra"],
    paddingX = 60,
    paddingY = 60,
    groupByCluster = false,
  } = opts;

  // レイヤごとにグループ化
  const byLayer = {};
  for (const ly of layerOrder) byLayer[ly] = [];
  for (const m of modules) {
    if (byLayer[m.layer]) byLayer[m.layer].push(m);
  }

  // 各レイヤ内でクラスタ順にソート
  const clusterOrder = ["added", "modified", "refactor", "removed"];
  for (const ly of layerOrder) {
    byLayer[ly].sort((a, b) => {
      if (groupByCluster) {
        const ca = clusterOrder.indexOf(a.cluster);
        const cb = clusterOrder.indexOf(b.cluster);
        if (ca !== cb) return ca - cb;
      }
      return a.name.localeCompare(b.name);
    });
  }

  // 配置
  const positions = {};
  let maxRowWidth = 0;
  for (let li = 0; li < layerOrder.length; li++) {
    const ly = layerOrder[li];
    const items = byLayer[ly];
    const rowWidth = items.length * nodeWidth + (items.length - 1) * nodeGapX;
    maxRowWidth = Math.max(maxRowWidth, rowWidth);
  }

  for (let li = 0; li < layerOrder.length; li++) {
    const ly = layerOrder[li];
    const items = byLayer[ly];
    const rowWidth = items.length * nodeWidth + (items.length - 1) * nodeGapX;
    const startX = paddingX + (maxRowWidth - rowWidth) / 2;
    const y = paddingY + li * layerGapY;
    items.forEach((m, idx) => {
      positions[m.id] = {
        x: startX + idx * (nodeWidth + nodeGapX),
        y,
        w: nodeWidth,
        h: nodeHeight,
        layer: ly,
        layerIndex: li,
      };
    });
  }

  // エッジパス（ベジェ曲線）
  const edgePaths = edges
    .map(([from, to]) => {
      const a = positions[from];
      const b = positions[to];
      if (!a || !b) return null;
      const x1 = a.x + a.w / 2;
      const y1 = a.y + a.h;
      const x2 = b.x + b.w / 2;
      const y2 = b.y;
      const dy = Math.abs(y2 - y1);
      const cy = dy * 0.45;
      // 同レイヤや上→上のエッジは横にカーブ
      if (b.layerIndex <= a.layerIndex) {
        const offset = 32;
        return {
          from, to,
          d: `M ${x1} ${y1 - a.h/2 + 2} C ${x1 + offset} ${y1 - a.h/2}, ${x2 + offset} ${y2 + b.h/2}, ${x2} ${y2 + b.h/2 - 2}`,
          x1, y1: y1 - a.h/2, x2, y2: y2 + b.h/2,
          backward: true,
        };
      }
      return {
        from, to,
        d: `M ${x1} ${y1} C ${x1} ${y1 + cy}, ${x2} ${y2 - cy}, ${x2} ${y2}`,
        x1, y1, x2, y2,
        backward: false,
      };
    })
    .filter(Boolean);

  const totalWidth = paddingX * 2 + maxRowWidth;
  const totalHeight = paddingY * 2 + (layerOrder.length - 1) * layerGapY + nodeHeight;

  return { positions, edgePaths, totalWidth, totalHeight, byLayer };
}

// スイムレーン版 — 横方向にクラスタ別レーン、縦方向にレイヤ
function computeSwimlaneLayout(modules, edges, opts = {}) {
  const {
    nodeWidth = 130,
    nodeHeight = 50,
    layerGapY = 110,
    nodeGapY = 12,
    laneGapX = 32,
    paddingX = 40,
    paddingY = 60,
    headerH = 48,
    laneOrder = ["added", "modified", "refactor", "removed"],
    layerOrder = ["ui", "domain", "data", "infra"],
  } = opts;

  // 各レーン×レイヤのセルにモジュールを格納
  const cells = {};
  for (const ly of layerOrder) {
    for (const cl of laneOrder) cells[`${ly}/${cl}`] = [];
  }
  for (const m of modules) {
    const k = `${m.layer}/${m.cluster}`;
    if (cells[k]) cells[k].push(m);
  }
  for (const k of Object.keys(cells)) {
    cells[k].sort((a, b) => a.name.localeCompare(b.name));
  }

  // 各レーンの幅 = max(各セルのノード数) * (nodeWidth)
  // ただしここでは各レーン1列に並べる: 幅 = nodeWidth
  const laneWidth = nodeWidth + 28;
  const positions = {};

  laneOrder.forEach((cl, ci) => {
    const laneX = paddingX + ci * (laneWidth + laneGapX) + 14;
    layerOrder.forEach((ly, li) => {
      const list = cells[`${ly}/${cl}`];
      const layerY = paddingY + headerH + li * layerGapY;
      list.forEach((m, ni) => {
        positions[m.id] = {
          x: laneX,
          y: layerY + ni * (nodeHeight + nodeGapY),
          w: nodeWidth,
          h: nodeHeight,
          lane: cl, layer: ly,
        };
      });
    });
  });

  // 各レーンの実際の高さを計算
  const laneCols = laneOrder.map((cl, ci) => paddingX + ci * (laneWidth + laneGapX));
  const totalWidth = paddingX * 2 + laneOrder.length * laneWidth + (laneOrder.length - 1) * laneGapX;

  // レイヤ行の高さを計算
  const layerHeights = layerOrder.map((ly) => {
    let max = 1;
    for (const cl of laneOrder) {
      max = Math.max(max, cells[`${ly}/${cl}`].length);
    }
    return max * (nodeHeight + nodeGapY) + 32;
  });
  // y位置を再計算
  const layerYStarts = [];
  let yc = paddingY + headerH;
  for (let i = 0; i < layerOrder.length; i++) {
    layerYStarts.push(yc);
    yc += layerHeights[i];
  }
  laneOrder.forEach((cl, ci) => {
    const laneX = paddingX + ci * (laneWidth + laneGapX) + 14;
    layerOrder.forEach((ly, li) => {
      const list = cells[`${ly}/${cl}`];
      const layerY = layerYStarts[li] + 12;
      list.forEach((m, ni) => {
        positions[m.id] = {
          x: laneX,
          y: layerY + ni * (nodeHeight + nodeGapY),
          w: nodeWidth,
          h: nodeHeight,
          lane: cl, layer: ly,
        };
      });
    });
  });
  const totalHeight = yc + paddingY;

  // エッジ
  const edgePaths = edges
    .map(([from, to]) => {
      const a = positions[from];
      const b = positions[to];
      if (!a || !b) return null;
      const sameLane = a.lane === b.lane;
      // 同レーン内: 縦の真ん中
      if (sameLane) {
        const x1 = a.x + a.w / 2;
        const y1 = a.y + a.h;
        const x2 = b.x + b.w / 2;
        const y2 = b.y;
        return { from, to, d: `M ${x1} ${y1} L ${x1} ${(y1+y2)/2} L ${x2} ${(y1+y2)/2} L ${x2} ${y2}`, sameLane: true };
      }
      // 他レーン: ノードのサイドを使う
      const x1 = a.x + a.w;
      const y1 = a.y + a.h / 2;
      const x2 = b.x;
      const y2 = b.y + b.h / 2;
      const mx = (x1 + x2) / 2;
      return { from, to, d: `M ${x1} ${y1} C ${mx} ${y1}, ${mx} ${y2}, ${x2} ${y2}`, sameLane: false };
    })
    .filter(Boolean);

  return { positions, edgePaths, totalWidth, totalHeight, layerYStarts, layerHeights, laneCols, laneWidth, laneGapX, headerH, layerOrder, laneOrder, cells };
}

window.computeDagLayout = computeDagLayout;
window.computeSwimlaneLayout = computeSwimlaneLayout;
