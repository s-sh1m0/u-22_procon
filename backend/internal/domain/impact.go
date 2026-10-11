package domain

// ImpactSubgraph は差分グラフ g から、PR で変わった関数（変更・追加・削除）と、
// その直接の呼び出し元・呼び出し先だけを残した部分グラフを返す。
//
// 呼び出しグラフは循環検出のために数ホップ先まで収集するが、レビューで見るのは
// 変わった関数とその直接の依存関係なので、保存・表示するグラフはここで絞る。
//
//   - added / removed のノードのうち、File が変更ファイル（changed の Filename /
//     PreviousFilename）に無いものは existing に戻す。関数が増減するのはそのファイルが
//     変わったときだけで、それ以外は base / head で探索の到達範囲が違っただけの誤判定のため。
//     エッジも同様に、両端とも変更ファイルに無ければ existing に戻す。
//   - 起点は Changed または DiffStatus が existing 以外のノード。
//   - 起点と、起点に直接つながるノードを残す。
//   - 起点を含む新規循環（IsNew）は、構成ノードも残す。
//   - エッジは両端が残ったものだけ、循環は構成ノードがすべて残ったものだけを返す。
//
// g と cycles は書き換えない。
func ImpactSubgraph(g *Graph, cycles []Cycle, changed []ChangedFile) (*Graph, []Cycle) {
	if g == nil {
		return &Graph{}, nil
	}

	changedFiles := make(map[string]struct{}, len(changed))
	for _, f := range changed {
		changedFiles[f.Filename] = struct{}{}
		if f.PreviousFilename != "" {
			changedFiles[f.PreviousFilename] = struct{}{}
		}
	}
	inChangedFile := make(map[NodeID]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		_, ok := changedFiles[n.File]
		inChangedFile[n.ID] = ok
	}

	seeds := make(map[NodeID]struct{})
	for _, n := range g.Nodes {
		if n.Changed || (n.DiffStatus != DiffStatusExisting && inChangedFile[n.ID]) {
			seeds[n.ID] = struct{}{}
		}
	}

	keep := make(map[NodeID]struct{}, len(seeds))
	for id := range seeds {
		keep[id] = struct{}{}
	}
	for _, e := range g.Edges {
		if _, ok := seeds[e.From]; ok {
			keep[e.To] = struct{}{}
		}
		if _, ok := seeds[e.To]; ok {
			keep[e.From] = struct{}{}
		}
	}
	for _, c := range cycles {
		if c.IsNew && containsAnyNode(c.Nodes, seeds) {
			for _, id := range c.Nodes {
				keep[id] = struct{}{}
			}
		}
	}

	out := &Graph{}
	for _, n := range g.Nodes {
		if _, ok := keep[n.ID]; !ok {
			continue
		}
		if !inChangedFile[n.ID] {
			n.DiffStatus = DiffStatusExisting
		}
		out.Nodes = append(out.Nodes, n)
	}
	for _, e := range g.Edges {
		_, fromOK := keep[e.From]
		_, toOK := keep[e.To]
		if !fromOK || !toOK {
			continue
		}
		if !inChangedFile[e.From] && !inChangedFile[e.To] {
			e.Status = DiffStatusExisting
		}
		out.Edges = append(out.Edges, e)
	}

	outCycles := make([]Cycle, 0, len(cycles))
	for _, c := range cycles {
		if containsAllNodes(c.Nodes, keep) {
			outCycles = append(outCycles, c)
		}
	}
	return out, outCycles
}

func containsAnyNode(ids []NodeID, set map[NodeID]struct{}) bool {
	for _, id := range ids {
		if _, ok := set[id]; ok {
			return true
		}
	}
	return false
}

func containsAllNodes(ids []NodeID, set map[NodeID]struct{}) bool {
	for _, id := range ids {
		if _, ok := set[id]; !ok {
			return false
		}
	}
	return true
}
